package test_utils

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/shared/transport/middleware"
	"github.com/yovannylopez/docsy-main/pkg/http_status"
)

// NewEchoWithCentralHTTPErrorHandler returns an Echo instance with the same HTTPErrorHandler as cmd/composition.
func NewEchoWithCentralHTTPErrorHandler() *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = middleware.CentralHTTPErrorHandler
	return e
}

// ServeEcho runs e.ServeHTTP and returns the recorder with the written response.
func ServeEcho(e *echo.Echo, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// DecodeAPIErrorEnvelope parses the JSON body from CentralHTTPErrorHandler: { "error": {...}, "request_id": "..." }.
func DecodeAPIErrorEnvelope(t *testing.T, body io.Reader) (requestID string, errObj map[string]any) {
	t.Helper()

	dec := json.NewDecoder(body)
	dec.UseNumber()

	var payload map[string]any
	require.NoError(t, dec.Decode(&payload))

	rid, _ := payload["request_id"].(string)
	errMap, ok := payload["error"].(map[string]any)
	require.True(t, ok, "error must be a JSON object")

	return rid, errMap
}

// AssertCentralJSONInternalServerError checks for 500 and the homogeneous generic internal message.
// If wantRequestID is not empty, also checks the request_id in the payload.
func AssertCentralJSONInternalServerError(t *testing.T, rec *httptest.ResponseRecorder, wantRequestID string) {
	t.Helper()

	require.Equal(t, http_status.InternalError.Code, rec.Code)

	rid, errObj := DecodeAPIErrorEnvelope(t, rec.Body)
	if wantRequestID != "" {
		require.Equal(t, wantRequestID, rid)
	}

	require.Equal(t, "INTERNAL_SERVER_ERROR", errObj["type"])
	require.Equal(t, json.Number(strconv.Itoa(http_status.InternalError.Code)), errObj["code"])
	require.Equal(t, http_status.EnvelopeInternalServerErrorMessageEN, errObj["message"])
}
