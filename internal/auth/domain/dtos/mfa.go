package dtos

// MFASetupResponse is returned by POST /auth/mfa/setup.
type MFASetupResponse struct {
	Secret     string `json:"secret"`
	QRCodeURL  string `json:"qr_code_url"`
	SetupToken string `json:"setup_token"`
}

// MFAConfirmRequest is the body for POST /auth/mfa/confirm.
type MFAConfirmRequest struct {
	SetupToken string `json:"setup_token"`
	TOTPCode   string `json:"totp_code"`
}

// MFAVerifyRequest is the body for POST /auth/mfa/verify (public route during login).
type MFAVerifyRequest struct {
	ChallengeToken string `json:"challenge_token"`
	TOTPCode       string `json:"totp_code"`
}

// MFADisableRequest is the body for POST /auth/mfa/disable.
type MFADisableRequest struct {
	TOTPCode string `json:"totp_code"`
}
