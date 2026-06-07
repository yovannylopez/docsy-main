package openapi

func envelopeStatusSchema(statusCodeExample int, statusDescExample string) *Schema {
	return &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"code": {
				Type:    "integer",
				Example: statusCodeExample,
			},
			"description": {
				Type:    "string",
				Example: statusDescExample,
			},
		},
	}
}

// SuccessEnvelopeSchema builds an OpenAPI schema aligned with pkg/responses.Response for successful responses:
// { "status": { "code", "description" }, "message", "data": { ...dataProperties } }.
func SuccessEnvelopeSchema(statusCodeExample int, statusDescExample, messageExample string, dataProperties map[string]*Schema) *Schema {
	return &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"status":  envelopeStatusSchema(statusCodeExample, statusDescExample),
			"message": {Type: "string", Example: messageExample},
			"data": {
				Type:       "object",
				Properties: dataProperties,
			},
		},
	}
}

// SuccessEnvelopeSchemaDataRef same as the success envelope, but "data" is a single $ref (e.g. UserResponse).
func SuccessEnvelopeSchemaDataRef(statusCodeExample int, statusDescExample, messageExample, dataRef string) *Schema {
	return &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"status":  envelopeStatusSchema(statusCodeExample, statusDescExample),
			"message": {Type: "string", Example: messageExample},
			"data":    {Ref: dataRef},
		},
	}
}

// SchemaRef returns a schema by reference to components.schemas.
func SchemaRef(ref string) *Schema {
	return &Schema{Ref: ref}
}

// SuccessEnvelopeExample builds an example value for documentation (Swagger example / curl).
func SuccessEnvelopeExample(statusCode int, statusDescription, message string, data any) map[string]any {
	return map[string]any{
		"status": map[string]any{
			"code":        statusCode,
			"description": statusDescription,
		},
		"message": message,
		"data":    data,
	}
}

// ErrorEnvelopeExample builds an error body example aligned with pkg/responses.ErrorResponse.
func ErrorEnvelopeExample(statusCode int, statusDescription, errorText string) map[string]any {
	return map[string]any{
		"status": map[string]any{
			"code":        statusCode,
			"description": statusDescription,
		},
		"error": errorText,
	}
}

// ApplicationJSONContent returns content["application/json"] with schema and optional example.
func ApplicationJSONContent(schema *Schema, example any) map[string]MediaType {
	mt := MediaType{Schema: schema}
	if example != nil {
		mt.Example = example
	}

	return map[string]MediaType{
		"application/json": mt,
	}
}

// JSONErrorRefContent documents standard error responses (status + error in pkg/responses).
func JSONErrorRefContent(example any) map[string]MediaType {
	return ApplicationJSONContent(SchemaRef("#/components/schemas/ErrorResponse"), example)
}
