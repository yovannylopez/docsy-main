package openapi

// APIErrorStatus describes the status object in errors returned by pkg/responses.EchoError.
type APIErrorStatus struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

// APIErrorEnvelope documents the standard JSON error body (status + error).
type APIErrorEnvelope struct {
	Status APIErrorStatus `json:"status"`
	Error  string         `json:"error"`
}

// RegisterStandardErrorResponseSchema registers or overwrites components.schemas.ErrorResponse
// to match pkg/responses.ErrorResponse. Must be called at the end of SetupAllSpecs
// to overwrite schemas generated from map[string]any in individual modules.
func RegisterStandardErrorResponseSchema(g *Generator) {
	sg := NewSchemaGenerator(g)
	_ = sg.GenerateSchemaFromStruct("ErrorResponse", APIErrorEnvelope{})
}
