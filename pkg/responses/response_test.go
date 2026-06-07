package responses

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yovannylopez/docsy-main/pkg/http_status"
	"github.com/yovannylopez/docsy-main/pkg/pagination"
)

func TestSuccessResponse(t *testing.T) {
	tests := []struct {
		name     string
		status   *http_status.Status
		data     any
		message  string
		expected *Response
	}{
		{
			name:    "basic successful response",
			status:  &http_status.OK,
			data:    map[string]string{"message": "success"},
			message: "Successful operation",
			expected: &Response{
				Status:  &http_status.OK,
				Message: "Successful operation",
				Data:    map[string]string{"message": "success"},
			},
		},
		{
			name:    "response with complex data",
			status:  &http_status.Created,
			data:    map[string]any{"id": 123, "name": "test"},
			message: "Resource created",
			expected: &Response{
				Status:  &http_status.Created,
				Message: "Resource created",
				Data:    map[string]any{"id": 123, "name": "test"},
			},
		},
		{
			name:    "response without message",
			status:  &http_status.OK,
			data:    "data",
			message: "",
			expected: &Response{
				Status:  &http_status.OK,
				Message: "",
				Data:    "data",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := SuccessResponse(tt.status, tt.data, tt.message)

			assert.NotNil(t, response)
			assert.Equal(t, tt.expected.Status, response.Status)
			assert.Equal(t, tt.expected.Message, response.Message)
			assert.Equal(t, tt.expected.Data, response.Data)
			assert.Empty(t, response.Error)
			assert.Nil(t, response.Meta)
		})
	}
}

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		name     string
		status   *http_status.Status
		errorMsg string
		expected *Response
	}{
		{
			name:     "validation error",
			status:   &http_status.BadRequest,
			errorMsg: "Invalid fields",
			expected: &Response{
				Status: &http_status.BadRequest,
				Error:  "Invalid fields",
			},
		},
		{
			name:     "internal error",
			status:   &http_status.InternalError,
			errorMsg: "Internal server error",
			expected: &Response{
				Status: &http_status.InternalError,
				Error:  "Internal server error",
			},
		},
		{
			name:     "authentication error",
			status:   &http_status.Unauthorized,
			errorMsg: "Invalid token",
			expected: &Response{
				Status: &http_status.Unauthorized,
				Error:  "Invalid token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := ErrorResponse(tt.status, tt.errorMsg)

			assert.NotNil(t, response)
			assert.Equal(t, tt.expected.Status, response.Status)
			assert.Equal(t, tt.expected.Error, response.Error)
			assert.Empty(t, response.Message)
			assert.Nil(t, response.Data)
			assert.Nil(t, response.Meta)
		})
	}
}

func TestJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		data           any
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful response",
			status:         http_status.OK.Code,
			data:           map[string]string{"message": "success"},
			expectedStatus: http_status.OK.Code,
			expectedBody:   `{"message":"success"}`,
		},
		{
			name:           "error response",
			status:         http_status.BadRequest.Code,
			data:           map[string]string{"error": "bad request"},
			expectedStatus: http_status.BadRequest.Code,
			expectedBody:   `{"error":"bad request"}`,
		},
		{
			name:           "response with complex data",
			status:         http_status.Created.Code,
			data:           map[string]any{"id": 123, "name": "test", "active": true},
			expectedStatus: http_status.Created.Code,
			expectedBody:   `{"id":123,"name":"test","active":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			require.NoError(t, JSON(rec, tt.status, tt.data))

			// Verify status code
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Verify content type
			contentType := rec.Header().Get("Content-Type")
			assert.Equal(t, "application/json", contentType)

			// Verify body
			body := rec.Body.String()
			assert.JSONEq(t, tt.expectedBody, body)
		})
	}
}

func TestJSON_EncodeError(t *testing.T) {
	rec := httptest.NewRecorder()
	err := JSON(rec, http_status.OK.Code, make(chan int))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "responses.JSON")
}

func TestEchoJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         *http_status.Status
		data           any
		message        string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful response",
			status:         &http_status.OK,
			data:           map[string]string{"message": "success"},
			message:        "Successful operation",
			expectedStatus: http_status.OK.Code,
			expectedBody:   `{"status":{"code":200,"description":"Successful operation"},"message":"Successful operation","data":{"message":"success"}}`,
		},
		{
			name:           "created response",
			status:         &http_status.Created,
			data:           map[string]any{"id": 123},
			message:        "Resource created",
			expectedStatus: http_status.Created.Code,
			expectedBody:   `{"status":{"code":201,"description":"Resource created successfully"},"message":"Resource created","data":{"id":123}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := EchoJSON(c, tt.status, tt.data, tt.message)

			// Verify no error
			assert.NoError(t, err)

			// Verify status code
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Verify content type
			contentType := rec.Header().Get("Content-Type")
			assert.Equal(t, "application/json", contentType)

			// Verify body
			body := rec.Body.String()
			assert.JSONEq(t, tt.expectedBody, body)
		})
	}
}

func TestEchoError(t *testing.T) {
	tests := []struct {
		name           string
		status         *http_status.Status
		errorMsg       string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "validation error",
			status:         &http_status.BadRequest,
			errorMsg:       "Invalid fields",
			expectedStatus: http_status.BadRequest.Code,
			expectedBody:   `{"status":{"code":400,"description":"Invalid request"},"error":"Invalid fields"}`,
		},
		{
			name:           "internal error",
			status:         &http_status.InternalError,
			errorMsg:       "Internal server error",
			expectedStatus: http_status.InternalError.Code,
			expectedBody:   `{"status":{"code":500,"description":"Internal server error"},"error":"Internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := EchoError(c, tt.status, tt.errorMsg)

			// Verify no error
			assert.NoError(t, err)

			// Verify status code
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Verify content type
			contentType := rec.Header().Get("Content-Type")
			assert.Equal(t, "application/json", contentType)

			// Verify body
			body := rec.Body.String()
			assert.JSONEq(t, tt.expectedBody, body)
		})
	}
}

func TestBindJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonBody    string
		target      any
		expectError bool
	}{
		{
			name:        "valid JSON",
			jsonBody:    `{"name":"test","age":25}`,
			target:      &map[string]any{},
			expectError: false,
		},
		{
			name:        "invalid JSON",
			jsonBody:    `{"name":"test","age":25,}`,
			target:      &map[string]any{},
			expectError: true,
		},
		{
			name:        "empty JSON",
			jsonBody:    `{}`,
			target:      &map[string]any{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.jsonBody))
			req.Header.Set("Content-Type", "application/json")

			err := BindJSON(req, tt.target)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBindJSONWithLimit_BodyTooLarge(t *testing.T) {
	payload := bytes.Repeat([]byte("a"), 150)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	var v map[string]any
	err := BindJSONWithLimit(req, &v, 100)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrJSONBodyTooLarge), "expected ErrJSONBodyTooLarge wrapped")
}

func TestBindJSONWithLimit_EmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(nil))
	var v map[string]any
	err := BindJSONWithLimit(req, &v, DefaultMaxJSONBodyBytes)
	require.Error(t, err)
	assert.True(t, errors.Is(err, io.EOF))
}

func TestBindJSONWithLimit_NilRequest(t *testing.T) {
	var v map[string]any
	err := BindJSONWithLimit(nil, &v, 100)
	require.Error(t, err)
}

func TestBindJSONWithLimit_InvalidMaxBytes(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
	var v map[string]any
	err := BindJSONWithLimit(req, &v, 0)
	require.Error(t, err)
}

func TestEchoAppError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	appErr := NewValidationError("required field").WithDetails(map[string]any{"field": "email"})
	require.NoError(t, EchoAppError(c, &http_status.BadRequest, appErr))
	assert.Equal(t, http_status.BadRequest.Code, rec.Code)

	var out Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.NotNil(t, out.Status)
	assert.Equal(t, http_status.BadRequest.Code, out.Status.Code)
	assert.Equal(t, "required field", out.Error)

	data, ok := out.Data.(map[string]any)
	require.True(t, ok, "data must be JSON object of AppError")
	assert.Equal(t, string(ValidationError), data["type"])
}

func TestEchoAppError_NilAppError(t *testing.T) {
	e := echo.New()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	err := EchoAppError(c, &http_status.BadRequest, nil)
	require.Error(t, err)
}

func TestBadRequestAppError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, BadRequestAppError(c, NewBadRequestError("invalid request")))
	assert.Equal(t, http_status.BadRequest.Code, rec.Code)
}

func TestValidate(t *testing.T) {
	// The Validate function currently always returns nil
	// This test verifies the current behavior
	tests := []struct {
		name        string
		data        any
		expectError bool
	}{
		{
			name:        "valid data",
			data:        map[string]string{"name": "test"},
			expectError: false,
		},
		{
			name:        "empty data",
			data:        map[string]string{},
			expectError: false,
		},
		{
			name:        "nil data",
			data:        nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.data)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOK(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	data := map[string]string{"message": "success"}
	message := "Successful operation"

	err := OK(c, data, message)

	assert.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, rec.Code)

	// Verify the response is valid JSON
	body := rec.Body.String()
	var response Response
	err = json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err)
	assert.Equal(t, http_status.OK.Code, response.Status.Code)
	assert.Equal(t, message, response.Message)
	// Verify data is correct (types may change during JSON serialization)
	assert.Equal(t, data["message"], response.Data.(map[string]any)["message"])
}

func TestCreated(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	data := map[string]any{"id": 123, "name": "test"}
	message := "Resource created"

	err := Created(c, data, message)

	assert.NoError(t, err)
	assert.Equal(t, http_status.Created.Code, rec.Code)

	// Verify the response is valid JSON
	body := rec.Body.String()
	var response Response
	err = json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err)
	assert.Equal(t, http_status.Created.Code, response.Status.Code)
	assert.Equal(t, message, response.Message)
	// Verify data is correct (types may change during JSON serialization)
	assert.Equal(t, float64(data["id"].(int)), response.Data.(map[string]any)["id"])
	assert.Equal(t, data["name"], response.Data.(map[string]any)["name"])
}

func TestBadRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	errorMsg := "Invalid fields"

	err := BadRequest(c, errorMsg)

	assert.NoError(t, err)
	assert.Equal(t, http_status.BadRequest.Code, rec.Code)

	// Verify the response is valid JSON
	body := rec.Body.String()
	var response Response
	err = json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err)
	assert.Equal(t, http_status.BadRequest.Code, response.Status.Code)
	assert.Equal(t, errorMsg, response.Error)
}

func TestUnauthorized(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	errorMsg := "Invalid token"

	err := Unauthorized(c, errorMsg)

	assert.NoError(t, err)
	assert.Equal(t, http_status.Unauthorized.Code, rec.Code)

	// Verify the response is valid JSON
	body := rec.Body.String()
	var response Response
	err = json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err)
	assert.Equal(t, http_status.Unauthorized.Code, response.Status.Code)
	assert.Equal(t, errorMsg, response.Error)
}

func TestForbidden(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	errorMsg := "Access denied"

	err := Forbidden(c, errorMsg)

	assert.NoError(t, err)
	assert.Equal(t, http_status.Forbidden.Code, rec.Code)

	// Verify the response is valid JSON
	body := rec.Body.String()
	var response Response
	err = json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err)
	assert.Equal(t, http_status.Forbidden.Code, response.Status.Code)
	assert.Equal(t, errorMsg, response.Error)
}

func TestNotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	errorMsg := "Resource not found"

	err := NotFound(c, errorMsg)

	assert.NoError(t, err)
	assert.Equal(t, http_status.NotFound.Code, rec.Code)

	// Verify the response is valid JSON
	body := rec.Body.String()
	var response Response
	err = json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err)
	assert.Equal(t, http_status.NotFound.Code, response.Status.Code)
	assert.Equal(t, errorMsg, response.Error)
}

func TestConflict(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	errorMsg := "Data conflict"

	err := Conflict(c, errorMsg)

	assert.NoError(t, err)
	assert.Equal(t, http_status.Conflict.Code, rec.Code)

	// Verify the response is valid JSON
	body := rec.Body.String()
	var response Response
	err = json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err)
	assert.Equal(t, http_status.Conflict.Code, response.Status.Code)
	assert.Equal(t, errorMsg, response.Error)
}

func TestUnprocessableEntity(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	errorMsg := "Unprocessable entity"

	err := UnprocessableEntity(c, errorMsg)

	assert.NoError(t, err)
	assert.Equal(t, http_status.UnprocessableEntity.Code, rec.Code)

	// Verify the response is valid JSON
	body := rec.Body.String()
	var response Response
	err = json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err)
	assert.Equal(t, http_status.UnprocessableEntity.Code, response.Status.Code)
	assert.Equal(t, errorMsg, response.Error)
}

func TestInternalError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	errorMsg := "Internal server error"

	err := InternalError(c, errorMsg)

	assert.NoError(t, err)
	assert.Equal(t, http_status.InternalError.Code, rec.Code)

	// Verify the response is valid JSON
	body := rec.Body.String()
	var response Response
	err = json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err)
	assert.Equal(t, http_status.InternalError.Code, response.Status.Code)
	assert.Equal(t, errorMsg, response.Error)
}

func TestNotImplemented(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	errorMsg := "Function not implemented"

	err := NotImplemented(c, errorMsg)

	assert.NoError(t, err)
	assert.Equal(t, http_status.NotImplemented.Code, rec.Code)

	// Verify the response is valid JSON
	body := rec.Body.String()
	var response Response
	err = json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err)
	assert.Equal(t, http_status.NotImplemented.Code, response.Status.Code)
	assert.Equal(t, errorMsg, response.Error)
}

func TestResponse_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		response *Response
		expected string
	}{
		{
			name: "complete successful response",
			response: &Response{
				Status:  &http_status.OK,
				Message: "Successful operation",
				Data:    map[string]string{"message": "success"},
				Meta:    map[string]any{"count": 1},
			},
			expected: `{"status":{"code":200,"description":"Successful operation"},"message":"Successful operation","data":{"message":"success"},"meta":{"count":1}}`,
		},
		{
			name: "error response",
			response: &Response{
				Status: &http_status.BadRequest,
				Error:  "Invalid fields",
			},
			expected: `{"status":{"code":400,"description":"Invalid request"},"error":"Invalid fields"}`,
		},
		{
			name: "minimal response",
			response: &Response{
				Status: &http_status.OK,
			},
			expected: `{"status":{"code":200,"description":"Successful operation"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.response)
			assert.NoError(t, err)

			// Verify the generated JSON is correct
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Verify it can be deserialized correctly
			var result Response
			err = json.Unmarshal(jsonData, &result)
			assert.NoError(t, err)

			// Verify main fields are correct
			assert.Equal(t, tt.response.Status.Code, result.Status.Code)
			assert.Equal(t, tt.response.Message, result.Message)
			assert.Equal(t, tt.response.Error, result.Error)
		})
	}
}

func TestResponse_EmptyFields(t *testing.T) {
	// Test that optional fields are handled correctly when empty
	response := &Response{
		Status: &http_status.OK,
	}

	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)

	var jsonMap map[string]any
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)

	// Verify only required fields are present
	assert.Contains(t, jsonMap, "status")
	assert.NotContains(t, jsonMap, "message")
	assert.NotContains(t, jsonMap, "data")
	assert.NotContains(t, jsonMap, "error")
	assert.NotContains(t, jsonMap, "meta")
}

func BenchmarkSuccessResponse(b *testing.B) {
	status := &http_status.OK
	data := map[string]string{"message": "success"}
	message := "Successful operation"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SuccessResponse(status, data, message)
	}
}

func BenchmarkErrorResponse(b *testing.B) {
	status := &http_status.BadRequest
	errorMsg := "Invalid fields"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ErrorResponse(status, errorMsg)
	}
}

func BenchmarkJSON(b *testing.B) {
	data := map[string]string{"message": "success"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		_ = JSON(rec, http_status.OK.Code, data)
	}
}

func BenchmarkEchoJSON(b *testing.B) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	status := &http_status.OK
	data := map[string]string{"message": "success"}
	message := "Successful operation"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EchoJSON(c, status, data, message)
	}
}

func BenchmarkEchoError(b *testing.B) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	status := &http_status.BadRequest
	errorMsg := "Invalid fields"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EchoError(c, status, errorMsg)
	}
}

func BenchmarkBindJSON(b *testing.B) {
	jsonBody := `{"name":"test","age":25}`
	target := &map[string]any{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		_ = BindJSON(req, target)
	}
}

func BenchmarkOK(b *testing.B) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	data := map[string]string{"message": "success"}
	message := "Successful operation"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = OK(c, data, message)
	}
}

func BenchmarkCreated(b *testing.B) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	data := map[string]any{"id": 123, "name": "test"}
	message := "Resource created"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Created(c, data, message)
	}
}

func BenchmarkBadRequest(b *testing.B) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	errorMsg := "Invalid fields"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BadRequest(c, errorMsg)
	}
}

func TestOKPaginated(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	page := pagination.CreateResponse(
		[]string{"row"},
		&pagination.Params{Limit: 10, Offset: 0},
		25,
	)
	require.NoError(t, OKPaginated(c, "done", page))
	assert.Equal(t, http_status.OK.Code, rec.Code)

	var out map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	var statusObj map[string]any
	require.NoError(t, json.Unmarshal(out["status"], &statusObj))
	assert.EqualValues(t, http_status.OK.Code, statusObj["code"])
	assert.Equal(t, http_status.OK.Description, statusObj["description"])
	assert.Contains(t, string(out["pagination"]), "total")
	assert.Contains(t, string(out["pagination"]), "25")
}
