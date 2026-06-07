package openapi

import (
	dto "github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/pkg/openapi"
)

// SetupMFASpec registers all MFA-related schemas and operations in the OpenAPI generator.
func SetupMFASpec(generator *openapi.Generator) {
	schemaGen := openapi.NewSchemaGenerator(generator)

	schemaGen.GenerateSchemaFromStruct("MFASetupResponse", dto.MFASetupResponse{})
	schemaGen.GenerateSchemaFromStruct("MFAConfirmRequest", dto.MFAConfirmRequest{})
	schemaGen.GenerateSchemaFromStruct("MFAVerifyRequest", dto.MFAVerifyRequest{})
	schemaGen.GenerateSchemaFromStruct("MFADisableRequest", dto.MFADisableRequest{})

	requestBodies := map[string]any{
		"MFAConfirmRequest": dto.MFAConfirmRequest{},
		"MFAVerifyRequest":  dto.MFAVerifyRequest{},
		"MFADisableRequest": dto.MFADisableRequest{},
	}
	schemaGen.GenerateRequestBodies(requestBodies)

	setupMFASetupOperation(generator)
	setupMFAConfirmOperation(generator)
	setupMFAVerifyOperation(generator)
	setupMFADisableOperation(generator)
}

func setupMFASetupOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	path := "/api/v1/auth/mfa/setup"
	pathItem, exists := spec.Paths[path]
	if !exists || pathItem.Post == nil {
		return
	}
	operation := pathItem.Post

	operation.Summary = "Initiate MFA TOTP setup"
	operation.Description = "Generates a TOTP secret and QR URL for the authenticator app. " +
		"Returns a single-use setup_token (TTL 10 min) to be used in /mfa/confirm. " +
		"Requires a valid JWT. User must not have MFA already active."
	operation.OperationID = "setupMFA"
	operation.Tags = []string{"mfa"}

	operation.Responses = map[string]openapi.Response{
		"200": {
			Description: "MFA setup initiated",
			Content: map[string]openapi.MediaType{
				"application/json": {
					Schema: &openapi.Schema{Ref: "#/components/schemas/MFASetupResponse"},
				},
			},
		},
		"401": {
			Description: "Unauthorized",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"409": {
			Description: "MFA already enabled",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"500": {
			Description: "Internal server error",
			Content:     openapi.JSONErrorRefContent(nil),
		},
	}
}

func setupMFAConfirmOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	path := "/api/v1/auth/mfa/confirm"
	pathItem, exists := spec.Paths[path]
	if !exists || pathItem.Post == nil {
		return
	}
	operation := pathItem.Post

	operation.Summary = "Confirm MFA setup with TOTP code"
	operation.Description = "Validates the setup_token and a TOTP code from the user's app. " +
		"On success, activates MFA (mfa_enabled=true) and stores the encrypted secret. " +
		"The setup_token is marked as used. Requires a valid JWT."
	operation.OperationID = "confirmMFA"
	operation.Tags = []string{"mfa"}

	operation.RequestBody = &openapi.RequestBody{
		Required: true,
		Content: map[string]openapi.MediaType{
			"application/json": {
				Schema: &openapi.Schema{Ref: "#/components/schemas/MFAConfirmRequest"},
			},
		},
	}

	operation.Responses = map[string]openapi.Response{
		"200": {
			Description: "MFA activated successfully",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"400": {
			Description: "Invalid or expired setup token / invalid TOTP code",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"401": {
			Description: "Unauthorized",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"500": {
			Description: "Internal server error",
			Content:     openapi.JSONErrorRefContent(nil),
		},
	}
}

func setupMFAVerifyOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	path := "/api/v1/auth/mfa/verify"
	pathItem, exists := spec.Paths[path]
	if !exists || pathItem.Post == nil {
		return
	}
	operation := pathItem.Post

	operation.Summary = "Complete MFA login challenge"
	operation.Description = "Validates the mfa_challenge_token (from login step 1) and the TOTP code. " +
		"On success, issues a full session with access + refresh tokens. " +
		"This endpoint is public (no JWT required)."
	operation.OperationID = "verifyMFA"
	operation.Tags = []string{"mfa"}

	operation.RequestBody = &openapi.RequestBody{
		Required: true,
		Content: map[string]openapi.MediaType{
			"application/json": {
				Schema: &openapi.Schema{Ref: "#/components/schemas/MFAVerifyRequest"},
			},
		},
	}

	operation.Responses = map[string]openapi.Response{
		"200": {
			Description: "Login completed — full session tokens returned",
			Content: map[string]openapi.MediaType{
				"application/json": {
					Schema: &openapi.Schema{Ref: "#/components/schemas/LoginResponse"},
				},
			},
		},
		"400": {
			Description: "Invalid, expired or already-used challenge token / invalid TOTP code",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"500": {
			Description: "Internal server error",
			Content:     openapi.JSONErrorRefContent(nil),
		},
	}
}

func setupMFADisableOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()
	path := "/api/v1/auth/mfa/disable"
	pathItem, exists := spec.Paths[path]
	if !exists || pathItem.Post == nil {
		return
	}
	operation := pathItem.Post

	operation.Summary = "Disable MFA for the authenticated user"
	operation.Description = "Validates the current password and TOTP code, then deactivates MFA " +
		"(mfa_enabled=false) and clears the stored secret. Requires a valid JWT."
	operation.OperationID = "disableMFA"
	operation.Tags = []string{"mfa"}

	operation.RequestBody = &openapi.RequestBody{
		Required: true,
		Content: map[string]openapi.MediaType{
			"application/json": {
				Schema: &openapi.Schema{Ref: "#/components/schemas/MFADisableRequest"},
			},
		},
	}

	operation.Responses = map[string]openapi.Response{
		"200": {
			Description: "MFA disabled successfully",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"400": {
			Description: "Invalid TOTP code or password",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"401": {
			Description: "Unauthorized",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"404": {
			Description: "MFA not enabled",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"500": {
			Description: "Internal server error",
			Content:     openapi.JSONErrorRefContent(nil),
		},
	}
}
