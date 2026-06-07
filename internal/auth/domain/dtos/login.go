package dtos

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the login response.
// When MFARequired is true, Token, Session and User are nil; the client must
// complete the challenge via POST /auth/mfa/verify using MFAChallengeToken.
type LoginResponse struct {
	User              *UserResponse    `json:"user,omitempty"`
	Token             *TokenResponse   `json:"token,omitempty"`
	Session           *SessionResponse `json:"session,omitempty"`
	MFARequired       bool             `json:"mfa_required,omitempty"`
	MFAChallengeToken string           `json:"mfa_challenge_token,omitempty"`
}
