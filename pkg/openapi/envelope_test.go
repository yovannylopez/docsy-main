package openapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuccessEnvelopeExample(t *testing.T) {
	data := map[string]any{"id": "x"}
	ex := SuccessEnvelopeExample(201, "Resource created successfully", "ok", data)

	require.Contains(t, ex, "status")
	require.Contains(t, ex, "message")
	require.Contains(t, ex, "data")

	st, ok := ex["status"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, 201, st["code"])
	assert.Equal(t, "Resource created successfully", st["description"])
	assert.Equal(t, "ok", ex["message"])
	assert.Equal(t, data, ex["data"])
}

func TestErrorEnvelopeExample(t *testing.T) {
	ex := ErrorEnvelopeExample(400, "Invalid request", "detail")

	st, ok := ex["status"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, 400, st["code"])
	assert.Equal(t, "Invalid request", st["description"])
	assert.Equal(t, "detail", ex["error"])
}

func TestSuccessEnvelopeSchema(t *testing.T) {
	s := SuccessEnvelopeSchema(201, "desc", "msg", map[string]*Schema{
		"user": {Ref: "#/components/schemas/UserResponse"},
	})
	require.Equal(t, "object", s.Type)
	require.Contains(t, s.Properties, "status")
	require.Contains(t, s.Properties, "message")
	require.Contains(t, s.Properties, "data")
	assert.Equal(t, "#/components/schemas/UserResponse", s.Properties["data"].Properties["user"].Ref)
}

func TestApplicationJSONContent(t *testing.T) {
	schema := &Schema{Type: "string"}
	ex := "hello"
	c := ApplicationJSONContent(schema, ex)
	require.Contains(t, c, "application/json")
	assert.Equal(t, schema, c["application/json"].Schema)
	assert.Equal(t, ex, c["application/json"].Example)
}

func TestApplicationJSONContent_NilExample(t *testing.T) {
	schema := &Schema{Type: "boolean"}
	c := ApplicationJSONContent(schema, nil)
	assert.Nil(t, c["application/json"].Example)
}

func TestSuccessEnvelopeSchemaDataRef(t *testing.T) {
	s := SuccessEnvelopeSchemaDataRef(200, "ok", "msg", "#/components/schemas/UserResponse")
	require.Equal(t, "#/components/schemas/UserResponse", s.Properties["data"].Ref)
}

func TestJSONErrorRefContent(t *testing.T) {
	ex := ErrorEnvelopeExample(401, "Unauthorized", "token")
	c := JSONErrorRefContent(ex)
	require.Contains(t, c, "application/json")
	assert.Equal(t, "#/components/schemas/ErrorResponse", c["application/json"].Schema.Ref)
	assert.Equal(t, ex, c["application/json"].Example)
}

func TestSchemaRef(t *testing.T) {
	assert.Equal(t, "#/x", SchemaRef("#/x").Ref)
}

func TestPaginationQueryBadRequestContent(t *testing.T) {
	c := PaginationQueryBadRequestContent()
	require.Contains(t, c, "application/json")
	ex, ok := c["application/json"].Example.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, PaginationQueryErrorLimitAboveMax, ex["error"])
}
