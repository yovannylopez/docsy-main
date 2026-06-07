package domain

// Razones persistidas en sessions.revoked_reason.
const (
	SessionRevokeReasonNewLogin       = "new_login"
	SessionRevokeReasonPasswordChange = "password_change"
	SessionRevokeReasonLogout         = "logout"
	SessionRevokeReasonLogoutAll      = "logout_all"
)
