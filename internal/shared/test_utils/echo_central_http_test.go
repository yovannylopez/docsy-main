package test_utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewEchoWithCentralHTTPErrorHandler_GenericPropagatedError(t *testing.T) {
	e := NewEchoWithCentralHTTPErrorHandler()
	e.GET("/x", func(c echo.Context) error {
		return assert.AnError
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := ServeEcho(e, req)

	AssertCentralJSONInternalServerError(t, rec, "unknown")
}
