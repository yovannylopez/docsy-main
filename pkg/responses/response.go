// Package responses defines the standard JSON wrapper for the Core (status + message/data or error)
// and Echo helpers aligned with pkg/http_status. Listings with limit/offset typically use
// OKPaginated together with pkg/pagination; municipalities/search keeps pagination inside data.
package responses

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/pkg/constants"
	"github.com/yovannylopez/docsy-main/pkg/http_status"
	"github.com/yovannylopez/docsy-main/pkg/pagination"
)

// DefaultMaxJSONBodyBytes is the default limit applied by BindJSON when reading the body (1 MiB).
const DefaultMaxJSONBodyBytes int64 = 1 << 20

// ErrJSONBodyTooLarge is returned (wrapped) when the body exceeds the maximum configured in BindJSONWithLimit.
var ErrJSONBodyTooLarge = errors.New("responses: JSON body exceeds the maximum allowed size")

// statusObject serializes the "status" block { code, description } as in JSON responses.
func statusObject(st *http_status.Status) map[string]any {
	return map[string]any{
		"code":        st.Code,
		"description": st.Description,
	}
}

// Response representa una respuesta HTTP estándar
type Response struct {
	Status  *http_status.Status `json:"status"`
	Message string              `json:"message,omitempty"`
	Data    any                 `json:"data,omitempty"`
	Error   string              `json:"error,omitempty"`
	Meta    map[string]any      `json:"meta,omitempty"`
}

// SuccessResponse crea una respuesta exitosa
func SuccessResponse(status *http_status.Status, data any, message string) *Response {
	return &Response{
		Status:  status,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse crea una respuesta de error
func ErrorResponse(status *http_status.Status, error string) *Response {
	return &Response{
		Status: status,
		Error:  error,
	}
}

// JSON writes a JSON response with the given HTTP code. Returns an error if serialization fails
// (e.g. non-JSON-serializable types); the client may have already received partial headers and status code.
func JSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("responses.JSON: %w", err)
	}
	return nil
}

// EchoJSON envía una respuesta JSON usando Echo
func EchoJSON(c echo.Context, status *http_status.Status, data any, message string) error {
	response := SuccessResponse(status, data, message)
	return c.JSON(status.Code, response)
}

// EchoError envía una respuesta de error usando Echo
func EchoError(c echo.Context, status *http_status.Status, error string) error {
	response := ErrorResponse(status, error)
	return c.JSON(status.Code, response)
}

// EchoAppError envía el envoltorio estándar con error = AppError.Message y data = el AppError completo
// (type, code, description, details, request_id) para clientes que necesiten detalle estructurado.
func EchoAppError(c echo.Context, status *http_status.Status, appErr *AppError) error {
	if appErr == nil {
		return fmt.Errorf("responses.EchoAppError: appErr is nil")
	}
	return c.JSON(status.Code, &Response{
		Status: status,
		Error:  appErr.Message,
		Data:   appErr,
	})
}

// BadRequestAppError equivale a EchoAppError con 400 Bad Request.
func BadRequestAppError(c echo.Context, appErr *AppError) error {
	return EchoAppError(c, &http_status.BadRequest, appErr)
}

// BindJSON decodes the JSON body into v reading at most DefaultMaxJSONBodyBytes bytes.
func BindJSON(r *http.Request, v any) error {
	return BindJSONWithLimit(r, v, DefaultMaxJSONBodyBytes)
}

// BindJSONWithLimit reads the full body (up to maxBytes+1 to detect overflow) and calls json.Unmarshal.
// Empty body returns a wrapped io.EOF. If len(body) > maxBytes, the error wraps ErrJSONBodyTooLarge.
func BindJSONWithLimit(r *http.Request, v any, maxBytes int64) error {
	if r == nil {
		return fmt.Errorf("responses.BindJSONWithLimit: request is nil")
	}
	if maxBytes < 1 {
		return fmt.Errorf("responses.BindJSONWithLimit: maxBytes must be >= 1")
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxBytes+1))
	if err != nil {
		return fmt.Errorf("responses.BindJSONWithLimit: read: %w", err)
	}
	if int64(len(body)) > maxBytes {
		return fmt.Errorf("responses.BindJSONWithLimit: %w (maximum %d bytes)", ErrJSONBodyTooLarge, maxBytes)
	}
	if len(body) == 0 {
		return fmt.Errorf("responses.BindJSONWithLimit: %w", io.EOF)
	}
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("responses.BindJSONWithLimit: %w", err)
	}
	return nil
}

// Deprecated: Validate performs no checks. Validate in handlers or use cases
// (e.g. github.com/go-playground/validator/v10). Kept for backward compatibility.
func Validate(v any) error {
	return nil
}

// OK crea una respuesta exitosa
func OK(c echo.Context, data any, message string) error {
	return EchoJSON(c, &http_status.OK, data, message)
}

// OKPaginated envía 200 con el formato estándar del Core: status, message, data y pagination
// (metadata limit/offset/total_pages de pkg/pagination). Municipios/search también usa este envoltorio
// conservando el query `page` además de `limit` para el caso de uso.
func OKPaginated(c echo.Context, message string, body pagination.Response) error {
	return c.JSON(http_status.OK.Code, map[string]any{
		"status":     statusObject(&http_status.OK),
		"message":    message,
		"data":       body.Data,
		"pagination": body.Metadata,
	})
}

// Created crea una respuesta exitosa
func Created(c echo.Context, data any, message string) error {
	return EchoJSON(c, &http_status.Created, data, message)
}

// BadRequest crea una respuesta de error
func BadRequest(c echo.Context, error string) error {
	return EchoError(c, &http_status.BadRequest, error)
}

// Unauthorized crea una respuesta de error
func Unauthorized(c echo.Context, error string) error {
	return EchoError(c, &http_status.Unauthorized, error)
}

// Forbidden crea una respuesta de error
func Forbidden(c echo.Context, error string) error {
	return EchoError(c, &http_status.Forbidden, error)
}

// NotFound crea una respuesta de error
func NotFound(c echo.Context, error string) error {
	return EchoError(c, &http_status.NotFound, error)
}

// Conflict crea una respuesta de error
func Conflict(c echo.Context, error string) error {
	return EchoError(c, &http_status.Conflict, error)
}

// UnprocessableEntity crea una respuesta de error
func UnprocessableEntity(c echo.Context, error string) error {
	return EchoError(c, &http_status.UnprocessableEntity, error)
}

// InternalError crea una respuesta de error
func InternalError(c echo.Context, error string) error {
	return EchoError(c, &http_status.InternalError, error)
}

// NotImplemented crea una respuesta de error
func NotImplemented(c echo.Context, error string) error {
	return EchoError(c, &http_status.NotImplemented, error)
}
