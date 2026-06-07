package domain

// Valores persistidos / expuestos en API para el campo result de audit_logs.
const (
	AuditResultSuccess = "success"
	AuditResultFailure = "failure"
	AuditResultError   = "error"
)

// Verbos CRUD genéricos usados por AuditService.
const (
	AuditActionCreate = "create"
	AuditActionUpdate = "update"
	AuditActionDelete = "delete"
	AuditActionRead   = "read"
)

// Acciones de negocio concretas (convención namespace.action).
const (
	AuditActionUserLoginAttempt     = "user.login_attempt"
	AuditActionUserLogout           = "user.logout"
	AuditActionUserCreated          = "user.created"
	AuditActionPasswordChanged      = "password_changed"
	AuditActionPasswordChangeFailed = "password_change_failed"
)

// ValidAuditResult indica si s es un valor permitido para el filtro `result` y los valores persistidos.
func ValidAuditResult(s string) bool {
	switch s {
	case AuditResultSuccess, AuditResultFailure, AuditResultError:
		return true
	default:
		return false
	}
}

// AuditResultFromBool mapea éxito de operación a resultado de auditoría.
func AuditResultFromBool(ok bool) string {
	if ok {
		return AuditResultSuccess
	}
	return AuditResultFailure
}
