package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	infraTypes "github.com/yovannylopez/docsy-main/internal/auth/infrastructure/types"
	apperrors "github.com/yovannylopez/docsy-main/pkg/errors"
)

// AuditLogRow represents an audit_logs row for scanning from the database
type AuditLogRow struct {
	ID            string            `db:"id"`
	UserID        *string           `db:"user_id"`
	SessionID     *string           `db:"session_id"`
	Action        string            `db:"action"`
	Resource      *string           `db:"resource"`
	ResourceID    *string           `db:"resource_id"`
	Result        string            `db:"result"`
	Message       *string           `db:"message"`
	IPAddress     *string           `db:"ip_address"`
	UserAgent     *string           `db:"user_agent"`
	RequestID     *string           `db:"request_id"`
	PreviousData  *infraTypes.JSONB `db:"previous_data"`
	NewData       *infraTypes.JSONB `db:"new_data"`
	ChangedFields pq.StringArray    `db:"changed_fields"`
	CreatedAt     time.Time         `db:"created_at"`
}

// AuditRepositoryAdapter implements AuditRepository using sqlx
type AuditRepositoryAdapter struct {
	db *sqlx.DB
}

// NewAuditRepositoryAdapter creates a new instance of AuditRepositoryAdapter
func NewAuditRepositoryAdapter(db *sqlx.DB) ports.AuditRepository {
	return &AuditRepositoryAdapter{db: db}
}

// LogAction inserts a new audit log
func (r *AuditRepositoryAdapter) LogAction(ctx context.Context, log *entities.AuditLog) error {
	if log == nil {
		return apperrors.ValidationError("NIL_AUDIT_LOG", "audit log cannot be nil")
	}

	// Generate ID if it does not exist
	if log.ID == "" {
		log.ID = uuid.NewString()
	}

	// Convert map to JSONB for serialization in PostgreSQL
	var previousDataValue, newDataValue any
	var err error

	if log.PreviousData != nil {
		j := infraTypes.JSONB(*log.PreviousData)
		previousDataValue, err = j.Value()
		if err != nil {
			return apperrors.InternalError("failed to convert previous_data to driver.Value", err)
		}
	}

	if log.NewData != nil {
		j := infraTypes.JSONB(*log.NewData)
		newDataValue, err = j.Value()
		if err != nil {
			return apperrors.InternalError("failed to convert new_data to driver.Value", err)
		}
	}

	query := `
		INSERT INTO audit_logs (
			id, user_id, session_id, action, resource, resource_id,
			result, message, ip_address, user_agent, request_id,
			previous_data, new_data, changed_fields, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err = r.db.ExecContext(ctx, query,
		log.ID,
		log.UserID,
		log.SessionID,
		log.Action,
		log.Resource,
		log.ResourceID,
		log.Result,
		log.Message,
		log.IPAddress,
		log.UserAgent,
		log.RequestID,
		previousDataValue,
		newDataValue,
		pq.Array(log.ChangedFields),
		log.CreatedAt,
	)
	if err != nil {
		return apperrors.DatabaseError("insert_audit_log", err)
	}

	return nil
}

// GetUserAuditLogs retrieves audit logs for a specific user
func (r *AuditRepositoryAdapter) GetUserAuditLogs(ctx context.Context, userID string, limit, offset int) ([]entities.AuditLog, error) {
	query := `
		SELECT 
			id, user_id, session_id, action, resource, resource_id,
			result, message, ip_address, user_agent, request_id,
			previous_data, new_data, changed_fields, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var rows []AuditLogRow
	err := r.db.SelectContext(ctx, &rows, query, userID, limit, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			return []entities.AuditLog{}, nil
		}
		return nil, apperrors.DatabaseError("get_user_audit_logs", err)
	}

	return r.convertRowsToAuditLogs(rows), nil
}

// GetSessionAuditLogs retrieves audit logs for a specific session
func (r *AuditRepositoryAdapter) GetSessionAuditLogs(ctx context.Context, sessionID string, limit, offset int) ([]entities.AuditLog, error) {
	query := `
		SELECT 
			id, user_id, session_id, action, resource, resource_id,
			result, message, ip_address, user_agent, request_id,
			previous_data, new_data, changed_fields, created_at
		FROM audit_logs
		WHERE session_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var rows []AuditLogRow
	err := r.db.SelectContext(ctx, &rows, query, sessionID, limit, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			return []entities.AuditLog{}, nil
		}
		return nil, apperrors.DatabaseError("get_session_audit_logs", err)
	}

	return r.convertRowsToAuditLogs(rows), nil
}

// List returns audit logs with advanced filters
func (r *AuditRepositoryAdapter) List(ctx context.Context, filters *dtos.AuditLogFilters) ([]entities.AuditLog, int, error) {
	// Build query with dynamic filters
	whereClauses := []string{}
	args := []any{}
	argIndex := 1

	// Apply filters
	if filters.UserID != nil && *filters.UserID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *filters.UserID)
		argIndex++
	}

	if filters.SessionID != nil && *filters.SessionID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("session_id = $%d", argIndex))
		args = append(args, *filters.SessionID)
		argIndex++
	}

	if filters.Action != nil && *filters.Action != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("action = $%d", argIndex))
		args = append(args, *filters.Action)
		argIndex++
	}

	if filters.Resource != nil && *filters.Resource != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("resource = $%d", argIndex))
		args = append(args, *filters.Resource)
		argIndex++
	}

	if filters.ResourceID != nil && *filters.ResourceID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("resource_id = $%d", argIndex))
		args = append(args, *filters.ResourceID)
		argIndex++
	}

	if filters.Result != nil && *filters.Result != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("result = $%d", argIndex))
		args = append(args, *filters.Result)
		argIndex++
	}

	if filters.Message != nil && *filters.Message != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("to_tsvector('spanish', COALESCE(message, '')) @@ plainto_tsquery('spanish', $%d)", argIndex))
		args = append(args, *filters.Message)
		argIndex++
	}

	if filters.StartDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filters.EndDate)
		argIndex++
	}

	// Build WHERE clause
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Query to count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, apperrors.DatabaseError("count_audit_logs", err)
	}

	// Query to retrieve logs with pagination
	query := fmt.Sprintf(`
		SELECT 
			id, user_id, session_id, action, resource, resource_id,
			result, message, ip_address, user_agent, request_id,
			previous_data, new_data, changed_fields, created_at
		FROM audit_logs
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, filters.Limit, filters.Offset)

	var rows []AuditLogRow
	err = r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return []entities.AuditLog{}, total, nil
		}
		return nil, 0, apperrors.DatabaseError("list_audit_logs", err)
	}

	return r.convertRowsToAuditLogs(rows), total, nil
}

// convertRowsToAuditLogs converts AuditLogRow to entities.AuditLog
func (r *AuditRepositoryAdapter) convertRowsToAuditLogs(rows []AuditLogRow) []entities.AuditLog {
	logs := make([]entities.AuditLog, len(rows))
	for i, row := range rows {
		logs[i] = entities.AuditLog{
			ID:            row.ID,
			UserID:        row.UserID,
			SessionID:     row.SessionID,
			Action:        row.Action,
			Resource:      row.Resource,
			ResourceID:    row.ResourceID,
			Result:        row.Result,
			Message:       row.Message,
			IPAddress:     row.IPAddress,
			UserAgent:     row.UserAgent,
			RequestID:     row.RequestID,
			ChangedFields: []string(row.ChangedFields),
			CreatedAt:     row.CreatedAt,
		}

		// Convert infrastructure JSONB to domain map
		if row.PreviousData != nil {
			m := map[string]any(*row.PreviousData)
			logs[i].PreviousData = &m
		}
		if row.NewData != nil {
			m := map[string]any(*row.NewData)
			logs[i].NewData = &m
		}
	}
	return logs
}
