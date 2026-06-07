package responses

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domerrs "github.com/yovannylopez/docsy-main/pkg/errors"
	"github.com/yovannylopez/docsy-main/pkg/http_status"
)

func TestToHTTPAppError_NotFound(t *testing.T) {
	inner := domerrs.NotFoundError("municipality", "x-1")
	out, ok := ToHTTPAppError(inner)
	require.True(t, ok)
	assert.Equal(t, NotFoundError, out.Type)
	assert.Equal(t, http_status.NotFound.Code, out.Code)
	assert.Contains(t, out.Message, "municipality")
}

func TestToHTTPAppError_DatabaseGenericMessage(t *testing.T) {
	inner := domerrs.DatabaseError("search_x", errors.New("connection refused"))
	out, ok := ToHTTPAppError(inner)
	require.True(t, ok)
	assert.Equal(t, DatabaseError, out.Type)
	assert.Equal(t, http_status.InternalError.Code, out.Code)
	assert.Equal(t, msgDatabaseGenericES, out.Message)
	assert.Nil(t, out.Details)
}

func TestToHTTPAppError_ValidationDetails(t *testing.T) {
	inner := domerrs.ValidationError("EMPTY_ID", "id cannot be empty")
	out, ok := ToHTTPAppError(inner)
	require.True(t, ok)
	assert.Equal(t, ValidationError, out.Type)
	assert.Equal(t, http_status.BadRequest.Code, out.Code)
	require.NotNil(t, out.Details)
	assert.Equal(t, "EMPTY_ID", out.Details["domain_code"])
}

func TestToHTTPAppError_UserMessageOverrides(t *testing.T) {
	inner := domerrs.DatabaseError("op", errors.New("secret")).
		WithUserMessage("Controlled message")
	out, ok := ToHTTPAppError(inner)
	require.True(t, ok)
	assert.Equal(t, "Controlled message", out.Message)
}

func TestToHTTPAppError_WrappedChain(t *testing.T) {
	root := domerrs.NotFoundError("user", "u1")
	wrapped := domerrs.Wrap(root, "signup flow")
	out, ok := ToHTTPAppError(wrapped)
	require.True(t, ok)
	assert.Equal(t, http_status.NotFound.Code, out.Code)
}

func TestToHTTPAppError_NotDomain(t *testing.T) {
	out, ok := ToHTTPAppError(errors.New("plain"))
	assert.False(t, ok)
	assert.Nil(t, out)
}
