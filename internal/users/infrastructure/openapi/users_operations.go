package openapi

import (
	"github.com/yovannylopez/docsy-main/pkg/openapi"
)

func setupGetProfileOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()

	if pathItem, exists := spec.Paths["/api/v1/users/profile"]; exists && pathItem.Get != nil {
		operation := pathItem.Get

		operation.Summary = "Get user profile"
		operation.Description = "Returns the profile information of the authenticated user"
		operation.OperationID = "getProfile"
		operation.Tags = []string{"users"}

		operation.Security = []map[string][]string{
			{"BearerAuth": {}},
		}

		operation.Responses = map[string]openapi.Response{
			"200": {
				Description: "Profile retrieved successfully",
				Content: map[string]openapi.MediaType{
					"application/json": {
						Schema: &openapi.Schema{
							Ref: "#/components/schemas/UserResponse",
						},
					},
				},
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

		spec.Paths["/api/v1/users/profile"] = pathItem
	}
}

func setupUpdateProfileOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()

	if pathItem, exists := spec.Paths["/api/v1/users/profile"]; exists && pathItem.Put != nil {
		operation := pathItem.Put

		operation.Summary = "Update user profile"
		operation.Description = "Updates the profile information of the authenticated user"
		operation.OperationID = "updateProfile"
		operation.Tags = []string{"users"}

		operation.Security = []map[string][]string{
			{"BearerAuth": {}},
		}

		operation.RequestBody = &openapi.RequestBody{
			Required: true,
			Content: map[string]openapi.MediaType{
				"application/json": {
					Schema: &openapi.Schema{
						Type: "object",
						Properties: map[string]*openapi.Schema{
							"first_name": {Type: "string"},
							"last_name":  {Type: "string"},
							"phone":      {Type: "string"},
						},
					},
				},
			},
		}

		operation.Responses = map[string]openapi.Response{
			"200": {
				Description: "Profile updated successfully",
				Content: map[string]openapi.MediaType{
					"application/json": {
						Schema: &openapi.Schema{
							Ref: "#/components/schemas/UserResponse",
						},
					},
				},
			},
			"400": {
				Description: "Invalid input data",
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

		spec.Paths["/api/v1/users/profile"] = pathItem
	}
}

func setupGetUsersOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()

	if pathItem, exists := spec.Paths["/api/v1/users"]; exists && pathItem.Get != nil {
		operation := pathItem.Get

		operation.Summary = "Get list of users"
		operation.Description = "Returns a paginated list of system users"
		operation.OperationID = "getUsers"
		operation.Tags = []string{"users"}

		operation.Security = []map[string][]string{
			{"BearerAuth": {}},
		}

		operation.Parameters = []openapi.Parameter{
			{Name: "limit", In: "query", Description: "Maximum number of users to return (default: 10, maximum: 100)", Required: false, Schema: &openapi.Schema{Type: "integer"}},
			{Name: "offset", In: "query", Description: "Number of users to skip (default: 0)", Required: false, Schema: &openapi.Schema{Type: "integer"}},
		}

		operation.Responses = map[string]openapi.Response{
			"200": {
				Description: "Users list retrieved successfully",
				Content: map[string]openapi.MediaType{
					"application/json": {
						Schema: &openapi.Schema{
							Ref: "#/components/schemas/UsersListResponse",
						},
					},
				},
			},
			"400": {
				Description: "Invalid pagination parameters (`limit`/`offset` non-numeric, out of range or negative offset; same semantics as pkg/pagination)",
				Content:     openapi.PaginationQueryBadRequestContent(),
			},
			"401": {
				Description: "Invalid token",
				Content:     openapi.JSONErrorRefContent(nil),
			},
			"403": {
				Description: "No permission to access the users list",
				Content:     openapi.JSONErrorRefContent(nil),
			},
			"500": {
				Description: "Internal server error",
				Content:     openapi.JSONErrorRefContent(nil),
			},
		}

		spec.Paths["/api/v1/users"] = pathItem
	}
}

func setupGetUserByIDOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()

	if pathItem, exists := spec.Paths["/api/v1/users/{id}"]; exists && pathItem.Get != nil {
		operation := pathItem.Get

		operation.Summary = "Get user by ID"
		operation.Description = "Retrieves the details of a specific user by their ID"
		operation.OperationID = "getUserByID"
		operation.Tags = []string{"users"}

		operation.Security = []map[string][]string{
			{"BearerAuth": {}},
		}

		operation.Parameters = []openapi.Parameter{
			{Name: "id", In: "path", Description: "ID of the user to retrieve", Required: true, Schema: &openapi.Schema{Type: "string"}},
		}

		operation.Responses = map[string]openapi.Response{
			"200": {
				Description: "User retrieved successfully",
				Content: map[string]openapi.MediaType{
					"application/json": {
						Schema: &openapi.Schema{
							Ref: "#/components/schemas/UserResponse",
						},
					},
				},
			},
			"400": {
				Description: "Invalid user ID",
				Content:     openapi.JSONErrorRefContent(nil),
			},
			"401": {
				Description: "Invalid token",
				Content:     openapi.JSONErrorRefContent(nil),
			},
			"404": {
				Description: "User not found",
				Content:     openapi.JSONErrorRefContent(nil),
			},
			"500": {
				Description: "Internal server error",
				Content:     openapi.JSONErrorRefContent(nil),
			},
		}

		spec.Paths["/api/v1/users/{id}"] = pathItem
	}
}

func setupCreateUserOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()

	if pathItem, exists := spec.Paths["/api/v1/users"]; exists && pathItem.Post != nil {
		operation := pathItem.Post

		operation.Summary = "Create user"
		operation.Description = "Creates a new user in the system"
		operation.OperationID = "createUser"
		operation.Tags = []string{"users"}

		operation.Security = []map[string][]string{
			{"BearerAuth": {}},
		}

		operation.RequestBody = &openapi.RequestBody{
			Required: true,
			Content: map[string]openapi.MediaType{
				"application/json": {
					Schema: &openapi.Schema{
						OneOf: []*openapi.Schema{
							{Ref: "#/components/schemas/CreateUserRequest"},
							{Ref: "#/components/schemas/CreateUsersRequest"},
						},
					},
				},
			},
		}

		operation.Responses = map[string]openapi.Response{
			"201": {
				Description: "User(s) created successfully",
				Content: map[string]openapi.MediaType{
					"application/json": {
						Schema: &openapi.Schema{
							Ref: "#/components/schemas/CreateUsersResponse",
						},
					},
				},
			},
			"400": {
				Description: "Invalid input data",
				Content:     openapi.JSONErrorRefContent(nil),
			},
			"401": {
				Description: "Invalid token",
				Content:     openapi.JSONErrorRefContent(nil),
			},
			"409": {
				Description: "Email is already registered",
				Content:     openapi.JSONErrorRefContent(nil),
			},
			"500": {
				Description: "Internal server error",
				Content:     openapi.JSONErrorRefContent(nil),
			},
		}

		spec.Paths["/api/v1/users"] = pathItem
	}
}

func setupSearchUsersOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()

	if pathItem, exists := spec.Paths["/api/v1/users/search"]; exists && pathItem.Get != nil {
		operation := pathItem.Get

		operation.Summary = "Search users"
		operation.Description = "Search users by query string with optional active filter and pagination"
		operation.OperationID = "searchUsers"
		operation.Tags = []string{"users"}

		operation.Security = []map[string][]string{
			{"BearerAuth": {}},
		}

		operation.Parameters = []openapi.Parameter{
			{Name: "q", In: "query", Description: "Search query", Required: false, Schema: &openapi.Schema{Type: "string"}},
			{Name: "activo", In: "query", Description: "Filter by active flag (true/false)", Required: false, Schema: &openapi.Schema{Type: "string"}},
			{Name: "limit", In: "query", Description: "Page size (default: 10, max: 100)", Required: false, Schema: &openapi.Schema{Type: "integer"}},
			{Name: "offset", In: "query", Description: "Offset for pagination", Required: false, Schema: &openapi.Schema{Type: "integer"}},
		}

		operation.Responses = map[string]openapi.Response{
			"200": {
				Description: "Search results",
				Content: map[string]openapi.MediaType{
					"application/json": {
						Schema: &openapi.Schema{
							Ref: "#/components/schemas/UsersListResponse",
						},
					},
				},
			},
			"400": {
				Description: "Invalid query parameters",
				Content:     openapi.JSONErrorRefContent(nil),
			},
			"401": {
				Description: "Invalid token",
				Content:     openapi.JSONErrorRefContent(nil),
			},
			"403": {
				Description: "No permission",
				Content:     openapi.JSONErrorRefContent(nil),
			},
			"500": {
				Description: "Internal server error",
				Content:     openapi.JSONErrorRefContent(nil),
			},
		}

		spec.Paths["/api/v1/users/search"] = pathItem
	}
}

func setupPatchUserOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()

	if pathItem, exists := spec.Paths["/api/v1/users/{id}"]; exists && pathItem.Patch != nil {
		operation := pathItem.Patch

		operation.Summary = "Partially update user"
		operation.Description = "Updates selected fields of an existing user (PATCH)"
		operation.OperationID = "patchUser"
		operation.Tags = []string{"users"}

		operation.Security = []map[string][]string{
			{"BearerAuth": {}},
		}

		operation.Parameters = []openapi.Parameter{
			{Name: "id", In: "path", Description: "ID of the user to update", Required: true, Schema: &openapi.Schema{Type: "string"}},
		}

		operation.RequestBody = &openapi.RequestBody{
			Required: true,
			Content: map[string]openapi.MediaType{
				"application/json": {
					Schema: &openapi.Schema{
						Type: "object",
						Properties: map[string]*openapi.Schema{
							"first_name": {Type: "string"},
							"last_name":  {Type: "string"},
							"phone":      {Type: "string"},
							"role_name":  {Type: "string"},
						},
					},
				},
			},
		}

		operation.Responses = map[string]openapi.Response{
			"200": {
				Description: "User updated successfully",
				Content: map[string]openapi.MediaType{
					"application/json": {
						Schema: &openapi.Schema{
							Ref: "#/components/schemas/UserResponse",
						},
					},
				},
			},
			"400": {
				Description: "Invalid input data",
				Content:     openapi.JSONErrorRefContent(nil),
			},
			"401": {
				Description: "Invalid token",
				Content:     openapi.JSONErrorRefContent(nil),
			},
			"404": {
				Description: "User not found",
				Content:     openapi.JSONErrorRefContent(nil),
			},
			"500": {
				Description: "Internal server error",
				Content:     openapi.JSONErrorRefContent(nil),
			},
		}

		spec.Paths["/api/v1/users/{id}"] = pathItem
	}
}

func setupResetPasswordOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()

	if pathItem, exists := spec.Paths["/api/v1/users/reset-password"]; exists && pathItem.Post != nil {
		operation := pathItem.Post

		operation.Summary = "Password reset"
		operation.Description = "Initiates the password reset process for a user"
		operation.OperationID = "resetPassword"
		operation.Tags = []string{"users"}

		operation.Security = []map[string][]string{
			{"BearerAuth": {}},
		}

		operation.RequestBody = &openapi.RequestBody{
			Required: true,
			Content: map[string]openapi.MediaType{
				"application/json": {
					Schema: &openapi.Schema{
						Type: "object",
						Properties: map[string]*openapi.Schema{
							"email": {Type: "string", Format: "email"},
						},
						Required: []string{"email"},
					},
				},
			},
		}

		operation.Responses = map[string]openapi.Response{
			"200": {
				Description: "Reset process initiated successfully",
				Content: map[string]openapi.MediaType{
					"application/json": {
						Schema: &openapi.Schema{
							Type: "object",
							Properties: map[string]*openapi.Schema{
								"message": {Type: "string"},
							},
						},
					},
				},
			},
			"400": {
				Description: "Invalid email",
				Content:     openapi.JSONErrorRefContent(nil),
			},
			"401": {
				Description: "Invalid token",
				Content:     openapi.JSONErrorRefContent(nil),
			},
			"404": {
				Description: "User not found",
				Content:     openapi.JSONErrorRefContent(nil),
			},
			"500": {
				Description: "Internal server error",
				Content:     openapi.JSONErrorRefContent(nil),
			},
		}

		spec.Paths["/api/v1/users/reset-password"] = pathItem
	}
}
