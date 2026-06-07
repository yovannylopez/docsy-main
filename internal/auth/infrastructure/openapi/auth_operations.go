package openapi

import (
	"github.com/yovannylopez/docsy-main/pkg/openapi"
)

func setupLoginOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	path := "/api/v1/auth/login"
	pathItem, exists := spec.Paths[path]
	if !exists || pathItem.Post == nil {
		return
	}
	operation := pathItem.Post

	operation.Summary = "Log in"
	operation.Description = "Authenticates a user with email and password, returning access tokens."
	operation.OperationID = "login"
	operation.Tags = []string{"authentication"}

	operation.RequestBody = &openapi.RequestBody{
		Required: true,
		Content: map[string]openapi.MediaType{
			"application/json": {
				Schema: &openapi.Schema{
					Ref: "#/components/schemas/LoginRequest",
				},
			},
		},
	}

	operation.Responses = map[string]openapi.Response{
		"200": {
			Description: "Login successful",
			Content: map[string]openapi.MediaType{
				"application/json": {
					Schema: &openapi.Schema{
						Ref: "#/components/schemas/LoginResponse",
					},
				},
			},
		},
		"400": {
			Description: "Invalid input data, invalid credentials, or account temporarily locked " +
				"after repeated failed password attempts (same HTTP status as other client errors; " +
				"lockout thresholds via AUTH_LOCKOUT_MAX_ATTEMPTS / AUTH_LOCKOUT_DURATION).",
			Content: openapi.JSONErrorRefContent(nil),
		},
		"401": {
			Description: "Invalid credentials",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"423": {
			Description: "Account locked",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"429": {
			Description: "Too many attempts (rate limit)",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"500": {
			Description: "Internal server error",
			Content:     openapi.JSONErrorRefContent(nil),
		},
	}
}

func setupRefreshOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	path := "/api/v1/auth/refresh"
	pathItem, exists := spec.Paths[path]
	if !exists || pathItem.Post == nil {
		return
	}
	operation := pathItem.Post

	operation.Summary = "Refresh access token"
	operation.Description = "Generates new access and refresh JWTs; requires refresh_token from login with server-side session"
	operation.OperationID = "refreshToken"
	operation.Tags = []string{"authentication"}

	operation.RequestBody = &openapi.RequestBody{
		Required: true,
		Content: map[string]openapi.MediaType{
			"application/json": {
				Schema: &openapi.Schema{
					Ref: "#/components/schemas/RefreshTokenRequest",
				},
			},
		},
	}

	operation.Responses = map[string]openapi.Response{
		"200": {
			Description: "Token refreshed successfully",
			Content: map[string]openapi.MediaType{
				"application/json": {
					Schema: &openapi.Schema{
						Ref: "#/components/schemas/TokenResponse",
					},
				},
			},
		},
		"401": {
			Description: "Invalid refresh token or revoked session",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"429": {
			Description: "Too many attempts (rate limit)",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"500": {
			Description: "Internal server error",
			Content:     openapi.JSONErrorRefContent(nil),
		},
	}
}

func setupLogoutOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	path := "/api/v1/auth/logout"
	pathItem, exists := spec.Paths[path]
	if !exists || pathItem.Post == nil {
		return
	}
	operation := pathItem.Post

	operation.Summary = "Log out"
	operation.Description = "Revokes the session using Bearer access JWT or refresh_token in the body. " +
		"Records a user.logout audit event in audit_logs on success."
	operation.OperationID = "logout"
	operation.Tags = []string{"authentication"}

	operation.RequestBody = &openapi.RequestBody{
		Required: false,
		Content: map[string]openapi.MediaType{
			"application/json": {
				Schema: &openapi.Schema{
					Ref: "#/components/schemas/LogoutRequest",
				},
			},
		},
	}

	operation.Responses = map[string]openapi.Response{
		"200": {
			Description: "Session closed successfully",
			Content: map[string]openapi.MediaType{
				"application/json": {
					Schema: &openapi.Schema{
						Type: "object",
						Properties: map[string]*openapi.Schema{
							"message": {
								Type: "string",
							},
						},
					},
				},
			},
		},
		"400": {
			Description: "Invalid request",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"401": {
			Description: "Invalid token",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"500": {
			Description: "Internal server error",
			Content:     openapi.JSONErrorRefContent(nil),
		},
	}
}

func setupChangePasswordOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	path := "/api/v1/auth/change-password"
	pathItem, exists := spec.Paths[path]
	if !exists || pathItem.Post == nil {
		return
	}
	operation := pathItem.Post

	operation.Summary = "Change password"
	operation.Description = "Changes the authenticated user's password. " +
		"All existing sessions are revoked on success; the client must log in again to obtain new tokens."
	operation.OperationID = "changePassword"
	operation.Tags = []string{"authentication"}

	operation.RequestBody = &openapi.RequestBody{
		Required: true,
		Content: map[string]openapi.MediaType{
			"application/json": {
				Schema: &openapi.Schema{
					Ref: "#/components/schemas/ChangePasswordRequest",
				},
			},
		},
	}

	operation.Responses = map[string]openapi.Response{
		"200": {
			Description: "Password changed successfully",
			Content: map[string]openapi.MediaType{
				"application/json": {
					Schema: &openapi.Schema{
						Ref: "#/components/schemas/ChangePasswordResponse",
					},
				},
			},
		},
		"400": {
			Description: "Validation error or incorrect current password",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"401": {
			Description: "Not authenticated or token invalidated",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"500": {
			Description: "Internal server error",
			Content:     openapi.JSONErrorRefContent(nil),
		},
	}
}
