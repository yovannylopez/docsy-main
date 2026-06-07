// Package openapi defines the OpenAPI documentation for the users API.
//
// Convention: same as in internal/auth/infrastructure/openapi (auth_schemas.go)
// and helpers in pkg/openapi (JSONErrorRefContent, SuccessEnvelope*, etc.).
package openapi

import (
	authDto "github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	userDto "github.com/yovannylopez/docsy-main/internal/users/domain/dtos"
	"github.com/yovannylopez/docsy-main/pkg/openapi"
)

// SetupUsersSpec registers OpenAPI schemas and operations for users.
func SetupUsersSpec(generator *openapi.Generator) {
	schemaGen := openapi.NewSchemaGenerator(generator)

	schemaGen.GenerateSchemaFromStruct("UserListResponse", userDto.UserListResponse{})
	schemaGen.GenerateSchemaFromStruct("UsersListResponse", userDto.UsersListResponse{})
	schemaGen.GenerateSchemaFromStruct("UserResponse", authDto.UserResponse{})
	schemaGen.GenerateSchemaFromStruct("CreateUserRequest", userDto.CreateUserRequest{})
	schemaGen.GenerateSchemaFromStruct("CreateUsersRequest", userDto.CreateUsersRequest{})
	schemaGen.GenerateSchemaFromStruct("CreateUsersResponse", userDto.CreateUsersResponse{})

	responses := map[string]any{
		"UserListResponse":    userDto.UserListResponse{},
		"UsersListResponse":   userDto.UsersListResponse{},
		"UserResponse":        authDto.UserResponse{},
		"CreateUserRequest":   userDto.CreateUserRequest{},
		"CreateUsersRequest":  userDto.CreateUsersRequest{},
		"CreateUsersResponse": userDto.CreateUsersResponse{},
	}
	schemaGen.GenerateResponses(responses)

	setupGetProfileOperation(generator)
	setupUpdateProfileOperation(generator)
	setupGetUsersOperation(generator)
	setupSearchUsersOperation(generator)
	setupGetUserByIDOperation(generator)
	setupCreateUserOperation(generator)
	setupPatchUserOperation(generator)
	setupResetPasswordOperation(generator)
}
