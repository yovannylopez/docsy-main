package repositories

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	"github.com/yovannylopez/docsy-main/pkg/errors"
)

// SessionRepository implements ports.SessionRepository using sqlx and PostgreSQL
type SessionRepository struct {
	db *sqlx.DB
}

// NewSessionRepository creates a new instance of SessionRepository
func NewSessionRepository(db *sqlx.DB) ports.SessionRepository {
	return &SessionRepository{db: db}
}

// Create creates a new session
func (r *SessionRepository) Create(ctx context.Context, session *entities.Session) error {
	query := `
    INSERT INTO sessions (
      id, user_id, refresh_token_hash, access_token_jti, user_agent,
      ip_address, location, device_fingerprint, created_at, last_used_at,
      expires_at, is_active
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
  `

	_, err := r.db.ExecContext(
		ctx,
		query,
		session.ID,
		session.UserID,
		session.RefreshTokenHash,
		session.AccessTokenJTI,
		session.UserAgent,
		session.IPAddress,
		session.Location,
		session.DeviceFingerprint,
		session.CreatedAt,
		session.LastUsedAt,
		session.ExpiresAt,
		session.IsActive,
	)
	if err != nil {
		return errors.DatabaseError("create_session", err)
	}

	return nil
}

// FindByID finds a session by its ID
func (r *SessionRepository) FindByID(ctx context.Context, sessionID string) (*entities.Session, error) {
	query := `
    SELECT id, user_id, refresh_token_hash, access_token_jti, user_agent,
      ip_address, location, device_fingerprint, created_at, last_used_at,
      expires_at, is_active, revoked_at, revoked_reason
    FROM sessions
    WHERE id = $1
  `

	var session entities.Session
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.AccessTokenJTI,
		&session.UserAgent,
		&session.IPAddress,
		&session.Location,
		&session.DeviceFingerprint,
		&session.CreatedAt,
		&session.LastUsedAt,
		&session.ExpiresAt,
		&session.IsActive,
		&session.RevokedAt,
		&session.RevokedReason,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.DatabaseError("find_session_by_id", err)
	}

	return &session, nil
}

// FindByUserID retrieves all (active) sessions for a user
func (r *SessionRepository) FindByUserID(ctx context.Context, userID string) ([]entities.Session, error) {
	query := `
    SELECT id, user_id, refresh_token_hash, access_token_jti, user_agent,
      ip_address, location, device_fingerprint, created_at, last_used_at,
      expires_at, is_active, revoked_at, revoked_reason
    FROM sessions
    WHERE user_id = $1 AND is_active = true
    ORDER BY last_used_at DESC
  `

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.DatabaseError("find_sessions_by_user_id", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	var sessions []entities.Session
	for rows.Next() {
		var session entities.Session
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.RefreshTokenHash,
			&session.AccessTokenJTI,
			&session.UserAgent,
			&session.IPAddress,
			&session.Location,
			&session.DeviceFingerprint,
			&session.CreatedAt,
			&session.LastUsedAt,
			&session.ExpiresAt,
			&session.IsActive,
			&session.RevokedAt,
			&session.RevokedReason,
		)
		if err != nil {
			return nil, errors.DatabaseError("scan_session_row", err)
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.DatabaseError("check_rows_error", err)
	}

	return sessions, nil
}

// FindByRefreshToken finds a session by its refresh token hash
func (r *SessionRepository) FindByRefreshToken(
	ctx context.Context,
	refreshTokenHash string,
) (*entities.Session, error) {
	query := `
    SELECT id, user_id, refresh_token_hash, access_token_jti, user_agent,
      ip_address, location, device_fingerprint, created_at, last_used_at,
      expires_at, is_active, revoked_at, revoked_reason
    FROM sessions
    WHERE refresh_token_hash = $1 AND is_active = true
  `

	var session entities.Session
	err := r.db.QueryRowContext(ctx, query, refreshTokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.AccessTokenJTI,
		&session.UserAgent,
		&session.IPAddress,
		&session.Location,
		&session.DeviceFingerprint,
		&session.CreatedAt,
		&session.LastUsedAt,
		&session.ExpiresAt,
		&session.IsActive,
		&session.RevokedAt,
		&session.RevokedReason,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.DatabaseError("find_session_by_refresh_token", err)
	}

	return &session, nil
}

// Update updates a complete session (common fields)
func (r *SessionRepository) Update(ctx context.Context, session *entities.Session) error {
	query := `
    UPDATE sessions
    SET refresh_token_hash = $2,
        access_token_jti   = $3,
        user_agent         = $4,
        ip_address         = $5,
        location           = $6,
        device_fingerprint = $7,
        last_used_at       = $8,
        expires_at         = $9,
        is_active          = $10,
        revoked_at         = $11,
        revoked_reason     = $12
    WHERE id = $1
  `

	_, err := r.db.ExecContext(
		ctx,
		query,
		session.ID,
		session.RefreshTokenHash,
		session.AccessTokenJTI,
		session.UserAgent,
		session.IPAddress,
		session.Location,
		session.DeviceFingerprint,
		session.LastUsedAt,
		session.ExpiresAt,
		session.IsActive,
		session.RevokedAt,
		session.RevokedReason,
	)
	if err != nil {
		return errors.DatabaseError("update_session", err)
	}
	return nil
}

// UpdateLastUsed updates the last used date of the session
func (r *SessionRepository) UpdateLastUsed(ctx context.Context, sessionID string) error {
	query := `
    UPDATE sessions
    SET last_used_at = NOW()
    WHERE id = $1
  `

	_, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return errors.DatabaseError("update_session_last_used", err)
	}
	return nil
}

// RevokeSession revokes a session
func (r *SessionRepository) RevokeSession(ctx context.Context, sessionID string, reason string) error {
	query := `
    UPDATE sessions
    SET is_active = false, revoked_at = NOW(), revoked_reason = $2
    WHERE id = $1
  `

	_, err := r.db.ExecContext(ctx, query, sessionID, reason)
	if err != nil {
		return errors.DatabaseError("revoke_session", err)
	}
	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (r *SessionRepository) RevokeAllUserSessions(ctx context.Context, userID string, reason string) error {
	query := `
    UPDATE sessions
    SET is_active = false, revoked_at = NOW(), revoked_reason = $2
    WHERE user_id = $1 AND is_active = true
  `

	_, err := r.db.ExecContext(ctx, query, userID, reason)
	if err != nil {
		return errors.DatabaseError("revoke_all_user_sessions", err)
	}
	return nil
}

// CleanupExpiredSessions cleans up expired sessions
func (r *SessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	query := `
    UPDATE sessions
    SET is_active = false, revoked_at = NOW(), revoked_reason = 'Session expired'
    WHERE expires_at < NOW() AND is_active = true
  `

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return errors.DatabaseError("clean_expired_sessions", err)
	}
	return nil
}
