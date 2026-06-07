// Package openapi registers routes and OpenAPI components for the auth slice.
//
// Documentation convention (aligned with pkg/responses):
//   - Success via responses.OK / responses.Created: body with "status", "message", "data".
//   - Errors via responses.EchoError: "status" and "error".
//   - Domain example data: *_openapi_examples.go files; generic helpers in pkg/openapi.
package openapi

import (
	dto "github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/pkg/openapi"
)

// SetupAuthSpec configures OpenAPI schemas and operations for authentication.
func SetupAuthSpec(generator *openapi.Generator) {
	schemaGen := openapi.NewSchemaGenerator(generator)

	schemaGen.GenerateSchemaFromStruct("LoginRequest", dto.LoginRequest{})
	schemaGen.GenerateSchemaFromStruct("RefreshTokenRequest", dto.RefreshTokenRequest{})
	schemaGen.GenerateSchemaFromStruct("LogoutRequest", dto.LogoutRequest{})
	// LoginResponse now includes optional mfa_required and mfa_challenge_token fields
	// for the MFA two-step login flow (US2). When mfa_required=true, Token/Session/User
	// are nil and the client must call POST /auth/mfa/verify with mfa_challenge_token.
	schemaGen.GenerateSchemaFromStruct("LoginResponse", dto.LoginResponse{})
	schemaGen.GenerateSchemaFromStruct("UserResponse", dto.UserResponse{})
	schemaGen.GenerateSchemaFromStruct("TokenResponse", dto.TokenResponse{})
	schemaGen.GenerateSchemaFromStruct("RoleResponse", dto.RoleResponse{})
	schemaGen.GenerateSchemaFromStruct("ChangePasswordRequest", dto.ChangePasswordRequest{})
	schemaGen.GenerateSchemaFromStruct("ChangePasswordResponse", dto.ChangePasswordResponse{})

	requestBodies := map[string]any{
		"LoginRequest":          dto.LoginRequest{},
		"RefreshTokenRequest":   dto.RefreshTokenRequest{},
		"LogoutRequest":         dto.LogoutRequest{},
		"ChangePasswordRequest": dto.ChangePasswordRequest{},
	}
	schemaGen.GenerateRequestBodies(requestBodies)

	responses := map[string]any{
		"LoginResponse":          dto.LoginResponse{},
		"ChangePasswordResponse": dto.ChangePasswordResponse{},
	}
	schemaGen.GenerateResponses(responses)

	setupLoginOperation(generator)
	setupRefreshOperation(generator)
	setupLogoutOperation(generator)
	setupChangePasswordOperation(generator)

	// MFA TOTP operations (US1–US3)
	SetupMFASpec(generator)
}
