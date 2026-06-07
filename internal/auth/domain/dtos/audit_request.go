package dtos

import "time"

// AuditLogFilters represents the filters for listing audit logs
type AuditLogFilters struct {
	UserID     *string    `json:"user_id,omitempty"`
	SessionID  *string    `json:"session_id,omitempty"`
	Action     *string    `json:"action,omitempty"`
	Resource   *string    `json:"resource,omitempty"`
	ResourceID *string    `json:"resource_id,omitempty"`
	Result     *string    `json:"result,omitempty"`
	Message    *string    `json:"message,omitempty"` // Text search in message
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}

// AuditLogListResponse represents the audit log list response
type AuditLogListResponse struct {
	Logs   []any `json:"logs"`
	Total  int   `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}
