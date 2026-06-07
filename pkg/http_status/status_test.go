package http_status

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus_IsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected bool
	}{
		{
			name:     "200 OK is successful",
			status:   OK,
			expected: true,
		},
		{
			name:     "201 Created is successful",
			status:   Created,
			expected: true,
		},
		{
			name:     "204 No Content is successful",
			status:   NoContent,
			expected: true,
		},
		{
			name:     "299 is successful",
			status:   Custom(299, "Custom Success"),
			expected: true,
		},
		{
			name:     "199 is not successful",
			status:   Custom(199, "Informational"),
			expected: false,
		},
		{
			name:     "300 is not successful",
			status:   Custom(300, "Redirection"),
			expected: false,
		},
		{
			name:     "400 Bad Request is not successful",
			status:   BadRequest,
			expected: false,
		},
		{
			name:     "500 Internal Error is not successful",
			status:   InternalError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsSuccess()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatus_IsClientError(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected bool
	}{
		{
			name:     "400 Bad Request is client error",
			status:   BadRequest,
			expected: true,
		},
		{
			name:     "401 Unauthorized is client error",
			status:   Unauthorized,
			expected: true,
		},
		{
			name:     "404 Not Found is client error",
			status:   NotFound,
			expected: true,
		},
		{
			name:     "499 is client error",
			status:   Custom(499, "Custom Client Error"),
			expected: true,
		},
		{
			name:     "399 is not client error",
			status:   Custom(399, "Redirection"),
			expected: false,
		},
		{
			name:     "500 is not client error",
			status:   Custom(500, "Server Error"),
			expected: false,
		},
		{
			name:     "200 OK is not client error",
			status:   OK,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsClientError()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatus_IsServerError(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected bool
	}{
		{
			name:     "500 Internal Error is server error",
			status:   InternalError,
			expected: true,
		},
		{
			name:     "502 Bad Gateway is server error",
			status:   BadGateway,
			expected: true,
		},
		{
			name:     "503 Service Unavailable is server error",
			status:   ServiceUnavailable,
			expected: true,
		},
		{
			name:     "599 is server error",
			status:   Custom(599, "Custom Server Error"),
			expected: true,
		},
		{
			name:     "499 is not server error",
			status:   Custom(499, "Client Error"),
			expected: false,
		},
		{
			name:     "600 is not server error",
			status:   Custom(600, "Unknown"),
			expected: false,
		},
		{
			name:     "200 OK is not server error",
			status:   OK,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsServerError()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatus_GetHTTPStatusText(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected string
	}{
		{
			name:     "200 OK",
			status:   OK,
			expected: "OK",
		},
		{
			name:     "201 Created",
			status:   Created,
			expected: "Created",
		},
		{
			name:     "400 Bad Request",
			status:   BadRequest,
			expected: "Bad Request",
		},
		{
			name:     "404 Not Found",
			status:   NotFound,
			expected: "Not Found",
		},
		{
			name:     "500 Internal Server Error",
			status:   InternalError,
			expected: "Internal Server Error",
		},
		{
			name:     "custom code",
			status:   Custom(299, "Custom Status"),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.GetHTTPStatusText()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCustom(t *testing.T) {
	tests := []struct {
		name         string
		code         int
		description  string
		expectedCode int
	}{
		{
			name:         "code 200",
			code:         200,
			description:  "Custom OK",
			expectedCode: 200,
		},
		{
			name:         "code 404",
			code:         404,
			description:  "Custom Not Found",
			expectedCode: 404,
		},
		{
			name:         "code 500",
			code:         500,
			description:  "Custom Server Error",
			expectedCode: 500,
		},
		{
			name:         "code 0",
			code:         0,
			description:  "Zero Code",
			expectedCode: 0,
		},
		{
			name:         "code negative",
			code:         -1,
			description:  "Negative Code",
			expectedCode: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Custom(tt.code, tt.description)
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.description, result.Description)
		})
	}
}

func TestLookupByCode(t *testing.T) {
	s, ok := LookupByCode(http.StatusNotFound)
	require.True(t, ok)
	assert.Equal(t, NotFound, s)

	_, ok = LookupByCode(99999)
	assert.False(t, ok)
}

func TestPtr(t *testing.T) {
	p := Ptr(OK)
	require.NotNil(t, p)
	assert.Equal(t, OK.Code, p.Code)
	assert.Equal(t, OK.Description, p.Description)
	// It should not be the same pointer as another Ptr
	q := Ptr(OK)
	assert.NotSame(t, p, q)
}

func TestCommonStatusCodes(t *testing.T) {
	codes := CommonStatusCodes()

	// Verify that the map is not empty
	assert.NotEmpty(t, codes)

	// Verify that it contains all the expected codes
	expectedCodes := []string{
		"OK", "Created", "Accepted", "NoContent",
		"BadRequest", "Unauthorized", "Forbidden", "NotFound",
		"MethodNotAllowed", "Conflict", "UnprocessableEntity", "TooManyRequests",
		"InternalError", "NotImplemented", "BadGateway", "ServiceUnavailable", "GatewayTimeout",
	}

	for _, code := range expectedCodes {
		assert.Contains(t, codes, code)
	}

	// Modifying the copy should not affect the internal catalog.
	delete(codes, "OK")
	codes2 := CommonStatusCodes()
	assert.Contains(t, codes2, "OK")

	// Verify that the codes have valid values
	for name, status := range codes2 {
		assert.NotEmpty(t, name)
		assert.GreaterOrEqual(t, status.Code, 0)
		assert.NotEmpty(t, status.Description)
	}
}

func TestStatus_JSONMarshaling(t *testing.T) {
	status := Status{Code: 200, Description: "Test Status"}

	encoded, err := json.Marshal(status)
	require.NoError(t, err)
	assert.JSONEq(t, `{"code":200,"description":"Test Status"}`, string(encoded))

	var decoded Status
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	assert.Equal(t, status, decoded)
}

func TestPredefinedStatusCodes(t *testing.T) {
	tests := []struct {
		name         string
		status       Status
		expectedCode int
		expectedDesc string
	}{
		{
			name:         "OK",
			status:       OK,
			expectedCode: http.StatusOK,
			expectedDesc: "Operation successful",
		},
		{
			name:         "Created",
			status:       Created,
			expectedCode: http.StatusCreated,
			expectedDesc: "Resource created successfully",
		},
		{
			name:         "Accepted",
			status:       Accepted,
			expectedCode: http.StatusAccepted,
			expectedDesc: "Request accepted",
		},
		{
			name:         "NoContent",
			status:       NoContent,
			expectedCode: http.StatusNoContent,
			expectedDesc: "No content",
		},
		{
			name:         "BadRequest",
			status:       BadRequest,
			expectedCode: http.StatusBadRequest,
			expectedDesc: "Invalid request",
		},
		{
			name:         "Unauthorized",
			status:       Unauthorized,
			expectedCode: http.StatusUnauthorized,
			expectedDesc: "Unauthorized",
		},
		{
			name:         "Forbidden",
			status:       Forbidden,
			expectedCode: http.StatusForbidden,
			expectedDesc: "Forbidden",
		},
		{
			name:         "NotFound",
			status:       NotFound,
			expectedCode: http.StatusNotFound,
			expectedDesc: "Resource not found",
		},
		{
			name:         "MethodNotAllowed",
			status:       MethodNotAllowed,
			expectedCode: http.StatusMethodNotAllowed,
			expectedDesc: "Method not allowed",
		},
		{
			name:         "Conflict",
			status:       Conflict,
			expectedCode: http.StatusConflict,
			expectedDesc: "Conflict",
		},
		{
			name:         "UnprocessableEntity",
			status:       UnprocessableEntity,
			expectedCode: http.StatusUnprocessableEntity,
			expectedDesc: "Unprocessable entity",
		},
		{
			name:         "TooManyRequests",
			status:       TooManyRequests,
			expectedCode: http.StatusTooManyRequests,
			expectedDesc: "Too many requests",
		},
		{
			name:         "InternalError",
			status:       InternalError,
			expectedCode: http.StatusInternalServerError,
			expectedDesc: "Internal server error",
		},
		{
			name:         "NotImplemented",
			status:       NotImplemented,
			expectedCode: http.StatusNotImplemented,
			expectedDesc: "Not implemented",
		},
		{
			name:         "BadGateway",
			status:       BadGateway,
			expectedCode: http.StatusBadGateway,
			expectedDesc: "Bad gateway",
		},
		{
			name:         "ServiceUnavailable",
			status:       ServiceUnavailable,
			expectedCode: http.StatusServiceUnavailable,
			expectedDesc: "Service unavailable",
		},
		{
			name:         "GatewayTimeout",
			status:       GatewayTimeout,
			expectedCode: http.StatusGatewayTimeout,
			expectedDesc: "Gateway timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedCode, tt.status.Code)
			assert.Equal(t, tt.expectedDesc, tt.status.Description)
		})
	}
}

func TestStatus_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected bool
	}{
		{
			name:     "code 0",
			status:   Custom(0, "Zero"),
			expected: false, // Not successful
		},
		{
			name:     "code 100",
			status:   Custom(100, "Continue"),
			expected: false, // Not successful
		},
		{
			name:     "code 299",
			status:   Custom(299, "Custom Success"),
			expected: true, // Successful
		},
		{
			name:     "code 399",
			status:   Custom(399, "Redirection"),
			expected: false, // Not client error
		},
		{
			name:     "code 499",
			status:   Custom(499, "Client Error"),
			expected: true, // Client error
		},
		{
			name:     "code 599",
			status:   Custom(599, "Server Error"),
			expected: true, // Server error
		},
		{
			name:     "code 600",
			status:   Custom(600, "Unknown"),
			expected: false, // Not server error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify that it does not cause a panic
			assert.NotPanics(t, func() {
				_ = tt.status.IsSuccess()
				_ = tt.status.IsClientError()
				_ = tt.status.IsServerError()
				_ = tt.status.GetHTTPStatusText()
			})
		})
	}
}

func BenchmarkStatus_IsSuccess(b *testing.B) {
	status := OK
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = status.IsSuccess()
	}
}

func BenchmarkStatus_IsClientError(b *testing.B) {
	status := BadRequest
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = status.IsClientError()
	}
}

func BenchmarkStatus_IsServerError(b *testing.B) {
	status := InternalError
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = status.IsServerError()
	}
}

func BenchmarkStatus_GetHTTPStatusText(b *testing.B) {
	status := OK
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = status.GetHTTPStatusText()
	}
}

func BenchmarkCustom(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Custom(200, "OK")
	}
}

func BenchmarkCommonStatusCodes(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CommonStatusCodes()
	}
}
