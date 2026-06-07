package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/yovannylopez/docsy-main/pkg/constants"
	"github.com/yovannylopez/docsy-main/pkg/http_status"
	"github.com/yovannylopez/docsy-main/pkg/logging"
	"github.com/yovannylopez/docsy-main/pkg/responses"
)

// CentralHTTPErrorHandler is the global Echo error handler for errors returned by handlers.
// Translates *responses.AppError, errors from pkg/errors (via responses.ToHTTPAppError), *echo.HTTPError
// and everything else to a homogeneous JSON response with request_id.
func CentralHTTPErrorHandler(err error, c echo.Context) {
	if err == nil {
		return
	}
	if c.Response().Committed {
		return
	}

	requestID := requestIDFromContext(c)

	logging.Error("HTTP error",
		logging.WithError(err),
		logging.WithRequestID(requestID),
		zap.String("method", c.Request().Method),
		zap.String("path", c.Request().URL.Path),
	)

	_ = respondAPIError(err, c, requestID)
}

// ErrorHandler is a middleware that handles errors in a centralized way (same logic as CentralHTTPErrorHandler).
func ErrorHandler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			requestID := requestIDFromContext(c)

			logging.Error("Request error",
				logging.WithError(err),
				logging.WithRequestID(requestID),
				zap.String("method", c.Request().Method),
				zap.String("path", c.Request().URL.Path),
			)

			return respondAPIError(err, c, requestID)
		}
	}
}

func requestIDFromContext(c echo.Context) string {
	if rid, ok := c.Get("request_id").(string); ok && rid != "" {
		return rid
	}
	if rid := c.Response().Header().Get(constants.RequestIDHeader); rid != "" {
		return rid
	}
	return "unknown"
}

func respondAPIError(err error, c echo.Context, requestID string) error {
	var respApp *responses.AppError
	if errors.As(err, &respApp) {
		return c.JSON(respApp.Code, map[string]any{
			constants.JSONFieldError:     respApp,
			constants.JSONFieldRequestID: requestID,
		})
	}

	if mapped, ok := responses.ToHTTPAppError(err); ok {
		return c.JSON(mapped.Code, map[string]any{
			constants.JSONFieldError:     mapped,
			constants.JSONFieldRequestID: requestID,
		})
	}

	var he *echo.HTTPError
	if errors.As(err, &he) {
		return c.JSON(he.Code, map[string]any{
			constants.JSONFieldError: map[string]any{
				"type":    "HTTP_ERROR",
				"code":    he.Code,
				"message": httpErrorMessage(he),
			},
			constants.JSONFieldRequestID: requestID,
		})
	}

	return c.JSON(http_status.InternalError.Code, map[string]any{
		constants.JSONFieldError: map[string]any{
			"type":    "INTERNAL_SERVER_ERROR",
			"code":    http_status.InternalError.Code,
			"message": http_status.EnvelopeInternalServerErrorMessageEN,
		},
		constants.JSONFieldRequestID: requestID,
	})
}

func httpErrorMessage(he *echo.HTTPError) string {
	switch m := he.Message.(type) {
	case string:
		return m
	case error:
		return m.Error()
	case nil:
		return http.StatusText(he.Code)
	default:
		return fmt.Sprint(m)
	}
}

// RequestIDMiddleware injects a unique request ID into the response header and Echo context.
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := generateRequestID()
			c.Response().Header().Set(constants.RequestIDHeader, requestID)
			c.Set("request_id", requestID)

			return next(c)
		}
	}
}

// generateRequestID generates a cryptographically secure unique ID for the request.
// Uses crypto/rand to guarantee uniqueness under high concurrency.
func generateRequestID() string {
	b := make([]byte, 16) //nolint:mnd
	if _, err := rand.Read(b); err != nil {
		return "req-fallback"
	}

	return "req-" + hex.EncodeToString(b)
}
