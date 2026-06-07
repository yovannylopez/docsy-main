package openapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterStandardErrorResponseSchema(t *testing.T) {
	g := NewGenerator("Test", "d", "1.0")
	RegisterStandardErrorResponseSchema(g)

	spec := g.GetSpec()
	require.Contains(t, spec.Components.Schemas, "ErrorResponse")
	s := spec.Components.Schemas["ErrorResponse"]
	assert.Equal(t, "object", s.Type)
	assert.Contains(t, s.Properties, "status")
	assert.Contains(t, s.Properties, "error")
}
