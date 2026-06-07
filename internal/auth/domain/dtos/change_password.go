package dtos

// ChangePasswordRequest represents the request body for changing a user's password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePasswordResponse is the success response body for the change-password endpoint.
// It intentionally contains no token fields (FR-006).
type ChangePasswordResponse struct {
	Message string `json:"message"`
}
