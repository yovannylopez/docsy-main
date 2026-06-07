package dtos

// RefreshTokenRequest body for POST /api/v1/auth/refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutRequest optional body if the Authorization header is not sent (logout with refresh).
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}
