package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/pkg/constants"
	domerrs "github.com/yovannylopez/docsy-main/pkg/errors"
	"github.com/yovannylopez/docsy-main/pkg/http_status"
	"github.com/yovannylopez/docsy-main/pkg/logging"
	"github.com/yovannylopez/docsy-main/pkg/responses"
)

func TestMain(m *testing.M) {
	_ = logging.Init(false)
	os.Exit(m.Run())
}

func newEchoContext(path string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func parseBody(t *testing.T, body string) map[string]any {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(body))
	dec.UseNumber()
	var m map[string]any
	require.NoError(t, dec.Decode(&m))
	return m
}

func TestErrorHandler_AppError(t *testing.T) {
	mw := ErrorHandler()
	next := func(c echo.Context) error {
		return responses.NewBadRequestError("invalid")
	}

	c, rec := newEchoContext("/test")
	c.Response().Header().Set(constants.RequestIDHeader, "req-123")

	err := mw(next)(c)
	assert.NoError(t, err)
	assert.Equal(t, http_status.BadRequest.Code, rec.Code)

	payload := parseBody(t, rec.Body.String())
	// error is marshaled AppError
	errObj := payload["error"].(map[string]any)
	assert.Equal(t, "BAD_REQUEST_ERROR", errObj["type"])
	assert.Equal(t, json.Number("400"), errObj["code"])
	assert.Equal(t, "invalid", errObj["message"])
	assert.Equal(t, "req-123", payload["request_id"])
}

func TestErrorHandler_HTTPError(t *testing.T) {
	mw := ErrorHandler()
	next := func(c echo.Context) error {
		return echo.NewHTTPError(http_status.NotFound.Code, "not found")
	}

	c, rec := newEchoContext("/missing")
	c.Response().Header().Set(constants.RequestIDHeader, "req-xyz")

	err := mw(next)(c)
	assert.NoError(t, err)
	assert.Equal(t, http_status.NotFound.Code, rec.Code)

	payload := parseBody(t, rec.Body.String())
	errObj := payload["error"].(map[string]any)
	assert.Equal(t, "HTTP_ERROR", errObj["type"])
	assert.Equal(t, json.Number("404"), errObj["code"])
	assert.Equal(t, "not found", errObj["message"])
	assert.Equal(t, "req-xyz", payload["request_id"])
}

func TestCentralHTTPErrorHandler_DomainNotFound(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = CentralHTTPErrorHandler

	req := httptest.NewRequest(http.MethodGet, "/m", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set(constants.RequestIDHeader, "rid-dom")

	CentralHTTPErrorHandler(domerrs.NotFoundError("municipality", "m1"), c)

	assert.Equal(t, http_status.NotFound.Code, rec.Code)
	payload := parseBody(t, rec.Body.String())
	errObj := payload["error"].(map[string]any)
	assert.Equal(t, "NOT_FOUND_ERROR", errObj["type"])
	assert.Equal(t, json.Number("404"), errObj["code"])
	assert.Equal(t, "rid-dom", payload["request_id"])
}

func TestErrorHandler_GenericError(t *testing.T) {
	mw := ErrorHandler()
	next := func(c echo.Context) error { return assert.AnError }

	c, rec := newEchoContext("/oops")

	err := mw(next)(c)
	assert.NoError(t, err)
	assert.Equal(t, http_status.InternalError.Code, rec.Code)

	payload := parseBody(t, rec.Body.String())
	errObj := payload["error"].(map[string]any)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", errObj["type"])
	assert.Equal(t, json.Number("500"), errObj["code"])
	assert.Equal(t, http_status.EnvelopeInternalServerErrorMessageEN, errObj["message"])
	assert.Equal(t, "unknown", payload["request_id"])
}

func TestErrorHandler_NoErrorPassesThrough(t *testing.T) {
	mw := ErrorHandler()
	next := func(c echo.Context) error {
		return c.String(http_status.OK.Code, "ok")
	}

	c, rec := newEchoContext("/ok")

	err := mw(next)(c)
	assert.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

func TestErrorHandler_UsesUnknownRequestIDWhenHeaderMissing(t *testing.T) {
	mw := ErrorHandler()
	next := func(c echo.Context) error {
		return responses.NewBadRequestError("missing header case")
	}

	c, rec := newEchoContext("/nohdr")
	// no X-Request-ID in the response yet; the error middleware reads the response header (empty → "unknown")

	err := mw(next)(c)
	assert.NoError(t, err)
	payload := parseBody(t, rec.Body.String())
	assert.Equal(t, "unknown", payload["request_id"])
}

func TestRequestIDMiddleware_SetsHeaderAndContext(t *testing.T) {
	mw := RequestIDMiddleware()
	next := func(c echo.Context) error { return nil }

	c, _ := newEchoContext("/rid")

	err := mw(next)(c)
	assert.NoError(t, err)

	rid := c.Response().Header().Get(constants.RequestIDHeader)
	assert.NotEmpty(t, rid)
	assert.True(t, strings.HasPrefix(rid, "req-"))
	// "req-" + 32 hex chars from crypto/rand
	assert.Len(t, rid, len("req-")+32)
	ctxVal := c.Get("request_id")
	assert.Equal(t, rid, ctxVal)
}
