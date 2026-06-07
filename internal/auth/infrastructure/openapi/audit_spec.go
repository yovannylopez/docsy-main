// Package openapi (audit): list of logs; errors using JSONErrorRefContent / ErrorEnvelopeExample.
package openapi

import (
	authdomain "github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/pkg/openapi"
)

// SetupAuditSpec registers the tag, AuditLog schema, and the GET /api/v1/auditoria operation.
func SetupAuditSpec(generator *openapi.Generator) {
	generator.AddTag("auditoria", "Audit and traceability operations")

	setupListAuditLogsOperation(generator)
	setupAuditLogSchema(generator)
}

// setupListAuditLogsOperation configures the list audit logs operation
func setupListAuditLogsOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()

	// Create or get the path item
	pathItem := spec.Paths["/api/v1/auditoria"]
	if pathItem.Get == nil {
		pathItem.Get = &openapi.Operation{}
	}

	operation := pathItem.Get
	operation.Summary = "List audit logs"
	operation.Description = `Retrieves a paginated list of audit logs with advanced filters.

Audit logs automatically record all CRUD operations performed in the system, providing complete traceability according to regulations.

**Important features:**
- Audit records are **immutable** (cannot be modified or deleted)
- Automatically captured: user, date/time, action, previous and new data
- Supports advanced filters by user, resource, action, date, etc.
- Includes context information: IP, user agent, request ID

**Usage examples:**
- Get all actions by a user: ?user_id=xxx
- View changes on a specific resource: ?resource=comunicaciones&resource_id=xxx
- Filter by date range: ?start_date=2024-01-01&end_date=2024-12-31
- Search for update actions: ?action=update`
	operation.OperationID = "listAuditLogs"
	operation.Tags = []string{"auditoria"}

	operation.Security = []map[string][]string{
		{openapi.SecuritySchemeBearer: {}},
	}

	// Query parameters
	operation.Parameters = []openapi.Parameter{
		{
			Name:        "user_id",
			In:          "query",
			Description: "Filter by user ID",
			Required:    false,
			Schema: &openapi.Schema{
				Type:   "string",
				Format: "uuid",
			},
		},
		{
			Name:        "session_id",
			In:          "query",
			Description: "Filter by session ID",
			Required:    false,
			Schema: &openapi.Schema{
				Type:   "string",
				Format: "uuid",
			},
		},
		{
			Name:        "action",
			In:          "query",
			Description: "Filter by action type. Possible values: `create`, `update`, `delete`, `read`",
			Required:    false,
			Schema: &openapi.Schema{
				Type: "string",
				Enum: []any{
					authdomain.AuditActionCreate,
					authdomain.AuditActionUpdate,
					authdomain.AuditActionDelete,
					authdomain.AuditActionRead,
				},
			},
			Example: "update",
		},
		{
			Name:        "resource",
			In:          "query",
			Description: "Filter by affected resource name. Examples: `comunicaciones`, `empresas`, `dependencias`, `usuarios`, `archivos`, `clasificacion_documental`",
			Required:    false,
			Schema: &openapi.Schema{
				Type: "string",
			},
			Example: "comunicaciones",
		},
		{
			Name:        "resource_id",
			In:          "query",
			Description: "Filter by specific resource ID",
			Required:    false,
			Schema: &openapi.Schema{
				Type: "string",
			},
		},
		{
			Name:        "result",
			In:          "query",
			Description: "Filter by action result. Possible values: `success`, `failure`, `error`",
			Required:    false,
			Schema: &openapi.Schema{
				Type: "string",
				Enum: []any{
					authdomain.AuditResultSuccess,
					authdomain.AuditResultFailure,
					authdomain.AuditResultError,
				},
			},
			Example: authdomain.AuditResultSuccess,
		},
		{
			Name:        "message",
			In:          "query",
			Description: "Search by text in the log descriptive message (partial search, case-insensitive)",
			Required:    false,
			Schema: &openapi.Schema{
				Type: "string",
			},
			Example: "created",
		},
		{
			Name:        "start_date",
			In:          "query",
			Description: "Start date of the range (format: YYYY-MM-DD)",
			Required:    false,
			Schema: &openapi.Schema{
				Type:   "string",
				Format: "date",
			},
		},
		{
			Name:        "end_date",
			In:          "query",
			Description: "End date of the range (format: YYYY-MM-DD)",
			Required:    false,
			Schema: &openapi.Schema{
				Type:   "string",
				Format: "date",
			},
		},
		{
			Name:        "limit",
			In:          "query",
			Description: "Maximum number of results per page. Default value: 20. Maximum allowed: 100",
			Required:    false,
			Schema: func() *openapi.Schema {
				min := float64(1)
				max := float64(100)
				return &openapi.Schema{
					Type:    "integer",
					Minimum: &min,
					Maximum: &max,
				}
			}(),
			Example: 20,
		},
		{
			Name:        "offset",
			In:          "query",
			Description: "Number of results to skip for pagination. Default value: 0. Must be greater than or equal to 0",
			Required:    false,
			Schema: func() *openapi.Schema {
				min := float64(0)
				return &openapi.Schema{
					Type:    "integer",
					Minimum: &min,
				}
			}(),
			Example: 0,
		},
	}

	operation.Responses = map[string]openapi.Response{
		"200": {
			Description: "Audit log list retrieved successfully",
			Content: map[string]openapi.MediaType{
				"application/json": {
					Schema: &openapi.Schema{
						Type: "object",
						Properties: map[string]*openapi.Schema{
							"status": {
								Type: "object",
								Properties: map[string]*openapi.Schema{
									"code": {
										Type:    "integer",
										Example: 200,
									},
									"description": {
										Type:    "string",
										Example: "Successful operation",
									},
								},
							},
							"message": {
								Type:    "string",
								Example: "audit logs retrieved successfully",
							},
							"data": {
								Type: "array",
								Items: &openapi.Schema{
									Ref: "#/components/schemas/AuditLog",
								},
							},
							"pagination": {
								Ref: "#/components/schemas/PaginationMetadata",
							},
						},
					},
					Example: map[string]any{
						"status": map[string]any{
							"code":        200,
							"description": "Successful operation",
						},
						"message": "audit logs retrieved successfully",
						"data": []any{
							map[string]any{
								"id":          "550e8400-e29b-41d4-a716-446655440000",
								"user_id":     "123e4567-e89b-12d3-a456-426614174000",
								"action":      "create",
								"resource":    "comunicaciones",
								"resource_id": "789e0123-e45b-67c8-d901-234567890abc",
								"result":      authdomain.AuditResultSuccess,
								"message":     "Created comunicaciones with ID 789e0123-e45b-67c8-d901-234567890abc",
								"ip_address":  "192.168.1.100",
								"user_agent":  "Mozilla/5.0...",
								"new_data": map[string]any{
									"id":        "789e0123-e45b-67c8-d901-234567890abc",
									"asunto":    "Solicitud de información",
									"estado_id": "estado-123",
								},
								"changed_fields": []any{},
								"created_at":     "2024-01-15T10:30:00Z",
							},
						},
						"pagination": map[string]any{
							"total":        150,
							"limit":        20,
							"offset":       0,
							"total_pages":  8,
							"current_page": 1,
							"has_next":     true,
							"has_previous": false,
						},
					},
				},
			},
		},
		"400": {
			Description: "Invalid request parameters (e.g.: incorrect date format, limit exceeds maximum allowed)",
			Content:     openapi.JSONErrorRefContent(openapi.ErrorEnvelopeExample(400, "Invalid request", openapi.PaginationQueryErrorLimitAboveMax)),
		},
		"401": {
			Description: "Invalid or missing authentication token",
			Content:     openapi.JSONErrorRefContent(nil),
		},
		"500": {
			Description: "Internal server error processing the request",
			Content:     openapi.JSONErrorRefContent(nil),
		},
	}

	spec.Paths["/api/v1/auditoria"] = pathItem
}

// setupAuditLogSchema configures the AuditLog schema
func setupAuditLogSchema(generator *openapi.Generator) {
	spec := generator.GetSpec()

	spec.Components.Schemas["AuditLog"] = &openapi.Schema{
		Type: "object",
		Properties: map[string]*openapi.Schema{
			"id": {
				Type:        "string",
				Format:      "uuid",
				Description: "Unique identifier for the audit log",
			},
			"user_id": {
				Type:        "string",
				Format:      "uuid",
				Nullable:    true,
				Description: "ID of the user who performed the action",
			},
			"session_id": {
				Type:        "string",
				Format:      "uuid",
				Nullable:    true,
				Description: "ID of the session where the action was performed",
			},
			"action": {
				Type:        "string",
				Description: "Type of action performed. Possible values: `create` (creation), `update` (update), `delete` (deletion), `read` (read)",
				Example:     authdomain.AuditActionCreate,
			},
			"resource": {
				Type:        "string",
				Nullable:    true,
				Description: "Name of the system resource affected by the action. Examples: `comunicaciones`, `empresas`, `dependencias`, `usuarios`, `archivos`, `clasificacion_documental`",
				Example:     "comunicaciones",
			},
			"resource_id": {
				Type:        "string",
				Nullable:    true,
				Description: "Specific ID of the affected resource",
			},
			"result": {
				Type:        "string",
				Description: "Result of the action. Possible values: `success` (success), `failure` (failure), `error` (error)",
				Example:     authdomain.AuditResultSuccess,
			},
			"message": {
				Type:        "string",
				Nullable:    true,
				Description: "Descriptive message of the action",
			},
			"ip_address": {
				Type:        "string",
				Nullable:    true,
				Description: "IP address from which the action was performed",
			},
			"user_agent": {
				Type:        "string",
				Nullable:    true,
				Description: "User agent of the browser or client application",
			},
			"request_id": {
				Type:        "string",
				Nullable:    true,
				Description: "Unique request identifier for correlation",
			},
			"previous_data": {
				Type:        "object",
				Nullable:    true,
				Description: "Previous resource data before the modification, in JSON format. Only present in `update` and `delete` action types. Contains the complete resource state before the change.",
			},
			"new_data": {
				Type:        "object",
				Nullable:    true,
				Description: "New resource data after the modification, in JSON format. Present in `create` and `update` action types. Contains the complete resource state after the change.",
			},
			"changed_fields": {
				Type: "array",
				Items: &openapi.Schema{
					Type: "string",
				},
				Description: "Array of field names that were modified in the action. Only present in `update` action types. Useful for quickly identifying which specific fields changed.",
				Example:     []any{"name", "email", "updated_at"},
			},
			"created_at": {
				Type:        "string",
				Format:      "date-time",
				Description: "Creation date and time of the log",
			},
		},
	}
}
