package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	apperrors "github.com/yovannylopez/docsy-main/pkg/errors"
	"github.com/yovannylopez/docsy-main/pkg/logging"
)

// UserRepository is the sqlx-based implementation for user operations
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin_transaction", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			_ = rollbackErr
		}
	}()

	query := `
        INSERT INTO users (
            id, email, username, password_hash, first_name, last_name,
            identification_number, identification_type, phone,
            is_active, is_verified, password_changed_at, must_change_password,
            created_at, updated_at, created_by, updated_by
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
    `

	if _, err = tx.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.IdentificationNumber,
		user.IdentificationType,
		user.Phone,
		user.IsActive,
		user.IsVerified,
		user.PasswordChangedAt,
		user.MustChangePassword,
		user.CreatedAt,
		user.UpdatedAt,
		user.CreatedBy,
		user.UpdatedBy,
	); err != nil {
		return apperrors.DatabaseError("insert_user", err)
	}

	// Batch insert roles to avoid N+1 queries
	if len(user.Roles) > 0 {
		valueStrings := make([]string, 0, len(user.Roles))
		valueArgs := make([]any, 0, len(user.Roles)*6) //nolint:mnd

		for i, role := range user.Roles {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)",
				i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6)) //nolint:mnd
			userRoleID := uuid.New().String()
			valueArgs = append(valueArgs, userRoleID, user.ID, role.ID, user.CreatedAt, user.CreatedAt, true)
		}

		roleQuery := fmt.Sprintf(`
            INSERT INTO user_roles (id, user_id, role_id, assigned_at, created_at, is_active)
            VALUES %s`, strings.Join(valueStrings, ","))

		if _, err = tx.ExecContext(ctx, roleQuery, valueArgs...); err != nil {
			return apperrors.DatabaseError("batch_insert_user_roles", err)
		}

		logging.Info("Batch inserted user roles",
			zap.String("user_id", user.ID),
			zap.Int("roles_count", len(user.Roles)))
	}

	if err = tx.Commit(); err != nil {
		return apperrors.DatabaseError("commit_transaction", err)
	}

	return nil
}

func (r *UserRepository) loadUserPermissionNames(ctx context.Context, userID string) ([]string, error) {
	q := `
		SELECT DISTINCT p.name
		FROM permissions p
		INNER JOIN role_permissions rp ON rp.permission_id = p.id
		INNER JOIN user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = $1 AND ur.is_active = true
	`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, apperrors.DatabaseError("query_user_permissions", err)
	}
	defer func() { _ = rows.Close() }()

	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, apperrors.DatabaseError("scan_user_permission", err)
		}
		names = append(names, n)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.DatabaseError("check_user_permissions_rows", err)
	}
	return names, nil
}

func (r *UserRepository) loadUserRoles(ctx context.Context, userID string) ([]entities.Role, error) {
	rolesQuery := `
        SELECT r.id, r.name, r.description, r.is_system_role, r.is_active, r.created_at, r.updated_at
        FROM roles r
        JOIN user_roles ur ON r.id = ur.role_id
        WHERE ur.user_id = $1 AND ur.is_active = true
    `
	rows, err := r.db.QueryContext(ctx, rolesQuery, userID)
	if err != nil {
		return nil, apperrors.DatabaseError("query_user_roles", err)
	}
	defer func() { _ = rows.Close() }()

	var roles []entities.Role
	for rows.Next() {
		var role entities.Role
		if err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.IsSystemRole,
			&role.IsActive,
			&role.CreatedAt,
			&role.UpdatedAt,
		); err != nil {
			return nil, apperrors.DatabaseError("scan_user_role", err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.DatabaseError("check_user_roles_rows_error", err)
	}
	return roles, nil
}

func (r *UserRepository) scanUserFromRow(row *sql.Row) (*entities.User, error) {
	var user entities.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.IdentificationNumber,
		&user.IdentificationType,
		&user.Phone,
		&user.IsActive,
		&user.IsVerified,
		&user.LastLoginAt,
		&user.FailedLoginAttempts,
		&user.LastFailedLoginAt,
		&user.LockedUntil,
		&user.MFAEnabled,
		&user.MFASecret,
		&user.PasswordChangedAt,
		&user.MustChangePassword,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.CreatedBy,
		&user.UpdatedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, apperrors.DatabaseError("scan_user_from_row", err)
	}
	return &user, nil
}

// FindByEmail finds a user by their email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	query := `
        SELECT u.id, u.email, u.username, u.password_hash, u.first_name, u.last_name, 
               u.identification_number, u.identification_type, u.phone,
               u.is_active, u.is_verified, u.last_login_at, u.failed_login_attempts,
               u.last_failed_login_at, u.locked_until, u.mfa_enabled, u.mfa_secret,
               u.password_changed_at, u.must_change_password, u.created_at, u.updated_at,
               u.created_by, u.updated_by
        FROM users u
        WHERE u.email = $1
    `
	user, err := r.scanUserFromRow(r.db.QueryRowContext(ctx, query, email))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	roles, err := r.loadUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles
	perms, err := r.loadUserPermissionNames(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.PermissionNames = perms
	return user, nil
}

// FindByID finds a user by their ID
func (r *UserRepository) FindByID(ctx context.Context, userID string) (*entities.User, error) {
	query := `
        SELECT u.id, u.email, u.username, u.password_hash, u.first_name, u.last_name, 
               u.identification_number, u.identification_type, u.phone,
               u.is_active, u.is_verified, u.last_login_at, u.failed_login_attempts,
               u.last_failed_login_at, u.locked_until, u.mfa_enabled, u.mfa_secret,
               u.password_changed_at, u.must_change_password, u.created_at, u.updated_at,
               u.created_by, u.updated_by
        FROM users u
        WHERE u.id = $1
    `
	user, err := r.scanUserFromRow(r.db.QueryRowContext(ctx, query, userID))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	roles, err := r.loadUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles
	perms, err := r.loadUserPermissionNames(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.PermissionNames = perms
	return user, nil
}

// FindByUsername finds a user by the username column (distinct from email).
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	query := `
        SELECT u.id, u.email, u.username, u.password_hash, u.first_name, u.last_name, 
               u.identification_number, u.identification_type, u.phone,
               u.is_active, u.is_verified, u.last_login_at, u.failed_login_attempts,
               u.last_failed_login_at, u.locked_until, u.mfa_enabled, u.mfa_secret,
               u.password_changed_at, u.must_change_password, u.created_at, u.updated_at,
               u.created_by, u.updated_by
        FROM users u
        WHERE u.username = $1
    `
	user, err := r.scanUserFromRow(r.db.QueryRowContext(ctx, query, username))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	roles, err := r.loadUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles
	perms, err := r.loadUserPermissionNames(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.PermissionNames = perms
	return user, nil
}

// Update updates a user (editable fields only)
func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	query := `
        UPDATE users
        SET email = $2, username = $3, first_name = $4, last_name = $5, 
            identification_number = $6, identification_type = $7, phone = $8,
            is_active = $9, is_verified = $10, 
            mfa_enabled = $11, updated_at = $12, updated_by = $13
        WHERE id = $1
    `
	result, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Username,
		user.FirstName,
		user.LastName,
		user.IdentificationNumber,
		user.IdentificationType,
		user.Phone,
		user.IsActive,
		user.IsVerified,
		user.MFAEnabled,
		user.UpdatedAt,
		user.UpdatedBy,
	)
	if err != nil {
		return apperrors.DatabaseError("update_user", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.DatabaseError("get_rows_affected", err)
	}
	if rowsAffected == 0 {
		return apperrors.NotFoundError("user", user.ID)
	}
	return nil
}

// GetRoleByName retrieves a role by its name
func (r *UserRepository) GetRoleByName(ctx context.Context, roleName string) (*entities.Role, error) {
	query := `
        SELECT id, name, description, is_system_role, is_active, created_at, updated_at
        FROM roles
        WHERE name = $1 AND is_active = true
    `
	var role entities.Role
	if err := r.db.QueryRowContext(ctx, query, roleName).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.IsSystemRole,
		&role.IsActive,
		&role.CreatedAt,
		&role.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFoundError("role", roleName)
		}
		return nil, apperrors.DatabaseError("get_role_by_name", err)
	}
	return &role, nil
}

// UpdateLastLogin updates the last login date
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	query := `
        UPDATE users
        SET last_login_at = NOW(), updated_at = NOW()
        WHERE id = $1
    `
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return apperrors.DatabaseError("update_last_login", err)
	}
	return nil
}

// IncrementFailedLoginAttempts increments failed login attempts without changing lockout state.
func (r *UserRepository) IncrementFailedLoginAttempts(ctx context.Context, userID string) error {
	_, err := r.RecordFailedPasswordAttempt(ctx, userID, 0, 0)
	return err
}

// RecordFailedPasswordAttempt increments failed_login_attempts and optionally sets locked_until
// when the new count reaches maxAttempts (maxAttempts must be > 0 for lockout branch).
func (r *UserRepository) RecordFailedPasswordAttempt(
	ctx context.Context, userID string, maxAttempts int, lockDuration time.Duration,
) (ports.FailedPasswordAttemptResult, error) {
	lockSeconds := int64(0)
	if maxAttempts > 0 && lockDuration > 0 {
		lockSeconds = int64(lockDuration.Round(time.Second) / time.Second)
		if lockSeconds < 1 {
			lockSeconds = 1
		}
	}

	query := `
        UPDATE users
        SET failed_login_attempts = failed_login_attempts + 1,
            last_failed_login_at = NOW(),
            locked_until = CASE
                WHEN $2::int > 0 AND (failed_login_attempts + 1) >= $2::int
                    THEN NOW() + ($3::bigint * interval '1 second')
                ELSE locked_until
            END,
            updated_at = NOW()
        WHERE id = $1
        RETURNING failed_login_attempts, locked_until
    `

	var (
		attempts int
		lu       sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, query, userID, maxAttempts, lockSeconds).Scan(&attempts, &lu)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ports.FailedPasswordAttemptResult{}, apperrors.NotFoundError("user", userID)
		}
		return ports.FailedPasswordAttemptResult{}, apperrors.DatabaseError("record_failed_password_attempt", err)
	}

	var lockedUntil *time.Time
	if lu.Valid {
		t := lu.Time
		lockedUntil = &t
	}
	return ports.FailedPasswordAttemptResult{FailedAttempts: attempts, LockedUntil: lockedUntil}, nil
}

// ResetFailedLoginAttempts resets failed login attempts
func (r *UserRepository) ResetFailedLoginAttempts(ctx context.Context, userID string) error {
	query := `
        UPDATE users
        SET failed_login_attempts = 0,
            last_failed_login_at = NULL,
            updated_at = NOW()
        WHERE id = $1
    `
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return apperrors.DatabaseError("reset_failed_login_attempts", err)
	}
	return nil
}

// LockUserAccount locks the user account
func (r *UserRepository) LockUserAccount(ctx context.Context, userID string, until *time.Time) error {
	query := `
        UPDATE users
        SET locked_until = $2, updated_at = NOW()
        WHERE id = $1
    `
	_, err := r.db.ExecContext(ctx, query, userID, until)
	if err != nil {
		return apperrors.DatabaseError("lock_user_account", err)
	}
	return nil
}

// UnlockUserAccount unlocks the user account
func (r *UserRepository) UnlockUserAccount(ctx context.Context, userID string) error {
	query := `
        UPDATE users
        SET locked_until = NULL, updated_at = NOW()
        WHERE id = $1
    `
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return apperrors.DatabaseError("unlock_user_account", err)
	}
	return nil
}

// UpdatePassword updates the user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	query := `
        UPDATE users
        SET password_hash = $2, password_changed_at = NOW(), updated_at = NOW()
        WHERE id = $1
    `
	_, err := r.db.ExecContext(ctx, query, userID, passwordHash)
	if err != nil {
		return apperrors.DatabaseError("update_password", err)
	}
	return nil
}

// SetMustChangePassword sets whether the user must change their password
func (r *UserRepository) SetMustChangePassword(ctx context.Context, userID string, mustChange bool) error {
	query := `
        UPDATE users
        SET must_change_password = $2, updated_at = NOW()
        WHERE id = $1
    `
	_, err := r.db.ExecContext(ctx, query, userID, mustChange)
	if err != nil {
		return apperrors.DatabaseError("set_must_change_password", err)
	}
	return nil
}

// UpdateMFASecret stores the encrypted secret without enabling MFA.
// Used during setup while the user has not yet confirmed their TOTP code.
func (r *UserRepository) UpdateMFASecret(ctx context.Context, userID, encryptedSecret string) error {
	query := `
        UPDATE users
        SET mfa_secret = $2, updated_at = NOW()
        WHERE id = $1
    `
	_, err := r.db.ExecContext(ctx, query, userID, encryptedSecret)
	if err != nil {
		return apperrors.DatabaseError("update_mfa_secret", err)
	}
	return nil
}

// EnableMFA enables MFA for the user
func (r *UserRepository) EnableMFA(ctx context.Context, userID, secret string) error {
	query := `
        UPDATE users
        SET mfa_enabled = true, mfa_secret = $2, updated_at = NOW()
        WHERE id = $1
    `
	_, err := r.db.ExecContext(ctx, query, userID, secret)
	if err != nil {
		return apperrors.DatabaseError("enable_mfa", err)
	}
	return nil
}

// DisableMFA disables MFA for the user
func (r *UserRepository) DisableMFA(ctx context.Context, userID string) error {
	query := `
        UPDATE users
        SET mfa_enabled = false, mfa_secret = NULL, updated_at = NOW()
        WHERE id = $1
    `
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return apperrors.DatabaseError("disable_mfa", err)
	}
	return nil
}

// VerifyUser marks the user as verified
func (r *UserRepository) VerifyUser(ctx context.Context, userID string) error {
	query := `
        UPDATE users
        SET is_verified = true, updated_at = NOW()
        WHERE id = $1
    `
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return apperrors.DatabaseError("verify_user", err)
	}
	return nil
}

// GetAllUsers retrieves all users with pagination
func (r *UserRepository) GetAllUsers(ctx context.Context, limit, offset int) ([]entities.User, error) {
	query := `
        SELECT u.id, u.email, u.username, u.password_hash, u.first_name, u.last_name, 
               u.identification_number, u.identification_type, u.phone,
               u.is_active, u.is_verified, u.last_login_at, u.failed_login_attempts,
               u.last_failed_login_at, u.locked_until, u.mfa_enabled, u.mfa_secret,
               u.password_changed_at, u.must_change_password, u.created_at, u.updated_at,
               u.created_by, u.updated_by
        FROM users u
        ORDER BY u.created_at DESC
        LIMIT $1 OFFSET $2
    `
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, apperrors.DatabaseError("get_all_users", err)
	}
	defer func() { _ = rows.Close() }()

	var users []entities.User
	for rows.Next() {
		var user entities.User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.PasswordHash,
			&user.FirstName,
			&user.LastName,
			&user.IdentificationNumber,
			&user.IdentificationType,
			&user.Phone,
			&user.IsActive,
			&user.IsVerified,
			&user.LastLoginAt,
			&user.FailedLoginAttempts,
			&user.LastFailedLoginAt,
			&user.LockedUntil,
			&user.MFAEnabled,
			&user.MFASecret,
			&user.PasswordChangedAt,
			&user.MustChangePassword,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.CreatedBy,
			&user.UpdatedBy,
		); err != nil {
			return nil, apperrors.DatabaseError("get_all_users_scan_row", err)
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, apperrors.DatabaseError("get_all_users_rows_iterate", err)
	}
	return users, nil
}

// GetTotalUsersCount obtiene el total de usuarios en la base de datos
func (r *UserRepository) GetTotalUsersCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`
	var count int
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, apperrors.DatabaseError("get_total_users_count", err)
	}
	return count, nil
}

// SearchUsers searches users by text across multiple fields with an optional active status filter
func (r *UserRepository) SearchUsers(ctx context.Context, query string, activo *bool, limit, offset int) ([]entities.User, error) {
	// Build dynamic query
	baseQuery := `
        SELECT u.id, u.email, u.username, u.password_hash, u.first_name, u.last_name, 
               u.identification_number, u.identification_type, u.phone,
               u.is_active, u.is_verified, u.last_login_at, u.failed_login_attempts,
               u.last_failed_login_at, u.locked_until, u.mfa_enabled, u.mfa_secret,
               u.password_changed_at, u.must_change_password, u.created_at, u.updated_at,
               u.created_by, u.updated_by
        FROM users u
        WHERE (
            to_tsvector('spanish', u.email) @@ plainto_tsquery('spanish', $1) OR 
            to_tsvector('spanish', u.first_name || ' ' || u.last_name) @@ plainto_tsquery('spanish', $1) OR 
            to_tsvector('spanish', COALESCE(u.username, '')) @@ plainto_tsquery('spanish', $1) OR 
            u.identification_number ILIKE $2
        )
    `

	args := []any{query, "%" + query + "%"}
	argCounter := 2

	// Add active status filter if present
	if activo != nil {
		baseQuery += fmt.Sprintf(" AND u.is_active = $%d", argCounter)
		args = append(args, *activo)
		argCounter++
	}

	// Add ordering and pagination
	baseQuery += " ORDER BY u.created_at DESC"
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("search_users", err)
	}
	defer func() { _ = rows.Close() }()

	var users []entities.User
	for rows.Next() {
		var user entities.User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.PasswordHash,
			&user.FirstName,
			&user.LastName,
			&user.IdentificationNumber,
			&user.IdentificationType,
			&user.Phone,
			&user.IsActive,
			&user.IsVerified,
			&user.LastLoginAt,
			&user.FailedLoginAttempts,
			&user.LastFailedLoginAt,
			&user.LockedUntil,
			&user.MFAEnabled,
			&user.MFASecret,
			&user.PasswordChangedAt,
			&user.MustChangePassword,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.CreatedBy,
			&user.UpdatedBy,
		); err != nil {
			return nil, apperrors.DatabaseError("search_users_scan_row", err)
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, apperrors.DatabaseError("search_users_rows_iterate", err)
	}
	return users, nil
}

// CountSearchUsers counts the total users matching the search and filter
func (r *UserRepository) CountSearchUsers(ctx context.Context, query string, activo *bool) (int, error) {
	baseQuery := `
        SELECT COUNT(*)
        FROM users u
        WHERE (
            to_tsvector('spanish', u.email) @@ plainto_tsquery('spanish', $1) OR 
            to_tsvector('spanish', u.first_name || ' ' || u.last_name) @@ plainto_tsquery('spanish', $1) OR 
            to_tsvector('spanish', COALESCE(u.username, '')) @@ plainto_tsquery('spanish', $1) OR 
            u.identification_number ILIKE $2
        )
    `

	args := []any{query, "%" + query + "%"}
	argCounter := 2

	// Add active status filter if present
	if activo != nil {
		baseQuery += fmt.Sprintf(" AND u.is_active = $%d", argCounter)
		args = append(args, *activo)
	}

	var count int
	if err := r.db.QueryRowContext(ctx, baseQuery, args...).Scan(&count); err != nil {
		return 0, apperrors.DatabaseError("count_search_users", err)
	}
	return count, nil
}
