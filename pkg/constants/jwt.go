package constants

// Standard jwt.MapClaims keys for access and refresh tokens issued by auth.
const (
	JWTClaimUserID            = "user_id"
	JWTClaimEmail             = "email"
	JWTClaimRole              = "role"
	JWTClaimExp               = "exp"
	JWTClaimIat               = "iat"
	JWTClaimPasswordChangedAt = "password_changed_at"
	JWTClaimSessionID         = "session_id"
	JWTClaimType              = "type"
	JWTTokenTypeRefresh       = "refresh"
)
