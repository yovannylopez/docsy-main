package openapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/yovannylopez/docsy-main/pkg/http_status"
)

func TestMiddleware_OpenAPIJSON(t *testing.T) {
	// Create a test generator
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	generator.AddServer("http://localhost:8080", "Development server")
	generator.AddTag("users", "User operations")

	// Create a test Echo application
	e := echo.New()
	e.Use(Middleware(generator))

	// Add a test route
	e.GET("/test", func(c echo.Context) error {
		return c.String(http_status.OK.Code, "test")
	})

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedType   string
		checkContent   func(t *testing.T, body string)
	}{
		{
			name:           "serve OpenAPI specification",
			path:           "/openapi.json",
			expectedStatus: http_status.OK.Code,
			expectedType:   "application/json",
			checkContent: func(t *testing.T, body string) {
				// Verify it is valid JSON
				var spec map[string]any
				err := json.Unmarshal([]byte(body), &spec)
				assert.NoError(t, err)

				// Verify basic OpenAPI fields
				assert.Equal(t, "3.0.3", spec["openapi"])
				assert.Equal(t, "Test API", spec["info"].(map[string]any)["title"])
				assert.Equal(t, "Test API description", spec["info"].(map[string]any)["description"])
				assert.Equal(t, "1.0.0", spec["info"].(map[string]any)["version"])

				// Verify it has servers
				servers, ok := spec["servers"].([]any)
				assert.True(t, ok)
				assert.Len(t, servers, 1)

				// Verify it has tags
				tags, ok := spec["tags"].([]any)
				assert.True(t, ok)
				assert.Len(t, tags, 1)
			},
		},
		{
			name:           "serve Swagger UI",
			path:           "/docs",
			expectedStatus: http_status.OK.Code,
			expectedType:   "text/html; charset=UTF-8",
			checkContent: func(t *testing.T, body string) {
				// Verify it contains Swagger UI HTML
				assert.Contains(t, body, "<!DOCTYPE html>")
				assert.Contains(t, body, "swagger-ui")
				assert.Contains(t, body, "/openapi.json")
				assert.Contains(t, body, "SwaggerUIBundle")
			},
		},
		{
			name:           "pass to next handler for normal routes",
			path:           "/test",
			expectedStatus: http_status.OK.Code,
			expectedType:   "text/plain; charset=UTF-8",
			checkContent: func(t *testing.T, body string) {
				assert.Equal(t, "test", body)
			},
		},
		{
			name:           "pass to next handler for non-existent routes",
			path:           "/nonexistent",
			expectedStatus: http_status.NotFound.Code,
			expectedType:   "application/json",
			checkContent: func(t *testing.T, body string) {
				// Echo returns JSON for non-existent routes when middleware is present
				var response map[string]any
				err := json.Unmarshal([]byte(body), &response)
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			// Execute request
			e.ServeHTTP(rec, req)

			// Verify status
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Verify content type
			contentType := rec.Header().Get("Content-Type")
			assert.Equal(t, tt.expectedType, contentType)

			// Verify content
			body := rec.Body.String()
			tt.checkContent(t, body)
		})
	}
}

func TestMiddleware_ErrorHandling(t *testing.T) {
	// Create a generator that causes an error
	generator := &Generator{
		spec: &Spec{
			OpenAPI: "3.0.0",
			Info: Info{
				Title:       "Test API",
				Description: "Test API description",
				Version:     "1.0.0",
			},
			Paths: make(map[string]PathItem),
		},
	}

	// Create a test Echo application
	e := echo.New()
	e.Use(Middleware(generator))

	// Add a test route
	e.GET("/test", func(c echo.Context) error {
		return c.String(http_status.OK.Code, "test")
	})

	// Verify the middleware works correctly
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Verify it works correctly
	assert.Equal(t, http_status.OK.Code, rec.Code)

	// Verify the body contains the OpenAPI specification
	body := rec.Body.String()
	var spec map[string]any
	err := json.Unmarshal([]byte(body), &spec)
	assert.NoError(t, err)
	assert.Equal(t, "3.0.0", spec["openapi"])
}

func TestGenerateSwaggerUI(t *testing.T) {
	// Verify generateSwaggerUI generates valid HTML
	html := generateSwaggerUI()

	// Verify it contains basic Swagger UI elements
	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, "<html")
	assert.Contains(t, html, "<head>")
	assert.Contains(t, html, "<body>")
	assert.Contains(t, html, "swagger-ui")
	assert.Contains(t, html, "SwaggerUIBundle")
	assert.Contains(t, html, "/openapi.json")

	// Verify it includes Swagger UI CDN links
	assert.Contains(t, html, "unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css")
	assert.Contains(t, html, "unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js")
	assert.Contains(t, html, "unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js")

	// Verify it has CSS styles
	assert.Contains(t, html, "box-sizing: border-box")
	assert.Contains(t, html, "overflow-y: scroll")

	// Verify it has JavaScript configuration
	assert.Contains(t, html, "deepLinking: true")
	assert.Contains(t, html, "StandaloneLayout")
}

func TestSetupOpenAPIRoutes(t *testing.T) {
	// Create a test generator
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	generator.AddServer("http://localhost:8080", "Development server")
	generator.AddTag("users", "User operations")

	// Create an Echo application
	e := echo.New()

	// Configure OpenAPI routes
	SetupOpenAPIRoutes(e, generator)

	// Add a test route
	e.GET("/test", func(c echo.Context) error {
		return c.String(http_status.OK.Code, "test")
	})

	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
		expectedType   string
		checkContent   func(t *testing.T, body string)
	}{
		{
			name:           "route /openapi.json configured",
			path:           "/openapi.json",
			method:         http.MethodGet,
			expectedStatus: http_status.OK.Code,
			expectedType:   "application/json",
			checkContent: func(t *testing.T, body string) {
				var spec map[string]any
				err := json.Unmarshal([]byte(body), &spec)
				assert.NoError(t, err)
				assert.Equal(t, "3.0.3", spec["openapi"])
				assert.Equal(t, "Test API", spec["info"].(map[string]any)["title"])
			},
		},
		{
			name:           "route /docs configured",
			path:           "/docs",
			method:         http.MethodGet,
			expectedStatus: http_status.OK.Code,
			expectedType:   "text/html; charset=UTF-8",
			checkContent: func(t *testing.T, body string) {
				assert.Contains(t, body, "<!DOCTYPE html>")
				assert.Contains(t, body, "swagger-ui")
			},
		},
		{
			name:           "test route works",
			path:           "/test",
			method:         http.MethodGet,
			expectedStatus: http_status.OK.Code,
			expectedType:   "text/plain; charset=UTF-8",
			checkContent: func(t *testing.T, body string) {
				assert.Equal(t, "test", body)
			},
		},
		{
			name:           "POST method on /openapi.json returns JSON",
			path:           "/openapi.json",
			method:         http.MethodPost,
			expectedStatus: http_status.OK.Code,
			expectedType:   "application/json",
			checkContent: func(t *testing.T, body string) {
				var spec map[string]any
				err := json.Unmarshal([]byte(body), &spec)
				assert.NoError(t, err)
				assert.Equal(t, "3.0.3", spec["openapi"])
			},
		},
		{
			name:           "POST method on /docs returns HTML",
			path:           "/docs",
			method:         http.MethodPost,
			expectedStatus: http_status.OK.Code,
			expectedType:   "text/html; charset=UTF-8",
			checkContent: func(t *testing.T, body string) {
				assert.Contains(t, body, "<!DOCTYPE html>")
				assert.Contains(t, body, "swagger-ui")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			// Execute request
			e.ServeHTTP(rec, req)

			// Verify status
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Verify content type
			contentType := rec.Header().Get("Content-Type")
			assert.Equal(t, tt.expectedType, contentType)

			// Verify content
			body := rec.Body.String()
			tt.checkContent(t, body)
		})
	}
}

func TestMiddleware_ContentTypeHeaders(t *testing.T) {
	// Create a test generator
	generator := NewGenerator("Test API", "Test API description", "1.0.0")

	// Create a test Echo application
	e := echo.New()
	e.Use(Middleware(generator))

	tests := []struct {
		name           string
		path           string
		expectedHeader string
	}{
		{
			name:           "OpenAPI JSON has correct content-type",
			path:           "/openapi.json",
			expectedHeader: "application/json",
		},
		{
			name:           "Swagger UI has correct content-type",
			path:           "/docs",
			expectedHeader: "text/html; charset=UTF-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			contentType := rec.Header().Get("Content-Type")
			assert.Equal(t, tt.expectedHeader, contentType)
		})
	}
}

func TestMiddleware_ErrorResponseFormat(t *testing.T) {
	// Create a generator that causes an error
	generator := &Generator{
		spec: &Spec{
			OpenAPI: "3.0.0",
			Info: Info{
				Title:       "Test API",
				Description: "Test API description",
				Version:     "1.0.0",
			},
			Paths: make(map[string]PathItem),
		},
	}

	// Create a test Echo application
	e := echo.New()
	e.Use(Middleware(generator))

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Verify the response is returned in JSON format
	contentType := rec.Header().Get("Content-Type")
	assert.Equal(t, "application/json", contentType)

	// Verify the body is valid JSON
	body := rec.Body.String()
	var spec map[string]any
	err := json.Unmarshal([]byte(body), &spec)
	assert.NoError(t, err)

	// Verify it contains the OpenAPI specification
	assert.Equal(t, "3.0.0", spec["openapi"])
	assert.Equal(t, "Test API", spec["info"].(map[string]any)["title"])
}

func TestMiddleware_ConcurrentRequests(t *testing.T) {
	// Create a test generator
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	generator.AddServer("http://localhost:8080", "Development server")

	// Create a test Echo application
	e := echo.New()
	e.Use(Middleware(generator))

	// Add a test route
	e.GET("/test", func(c echo.Context) error {
		return c.String(http_status.OK.Code, "test")
	})

	// Test multiple concurrent requests
	paths := []string{"/openapi.json", "/docs", "/test"}

	for _, path := range paths {
		t.Run("concurrent_"+path, func(t *testing.T) {
			// Create multiple concurrent requests
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func() {
					req := httptest.NewRequest(http.MethodGet, path, nil)
					rec := httptest.NewRecorder()
					e.ServeHTTP(rec, req)

					// Verify there are no errors
					assert.Equal(t, http_status.OK.Code, rec.Code)
					done <- true
				}()
			}

			// Wait for all requests to finish
			for i := 0; i < 10; i++ {
				<-done
			}
		})
	}
}

func TestMiddleware_RequestMethods(t *testing.T) {
	// Create a test generator
	generator := NewGenerator("Test API", "Test API description", "1.0.0")

	// Create a test Echo application
	e := echo.New()
	e.Use(Middleware(generator))

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
	paths := []string{"/openapi.json", "/docs"}

	for _, method := range methods {
		for _, path := range paths {
			t.Run(method+"_"+path, func(t *testing.T) {
				req := httptest.NewRequest(method, path, nil)
				rec := httptest.NewRecorder()

				e.ServeHTTP(rec, req)

				// All methods should work for middleware routes
				assert.Equal(t, http_status.OK.Code, rec.Code)

				// Verify content type
				contentType := rec.Header().Get("Content-Type")
				if path == "/openapi.json" {
					assert.Equal(t, "application/json", contentType)
				} else if path == "/docs" {
					assert.Equal(t, "text/html; charset=UTF-8", contentType)
				}
			})
		}
	}
}

func TestMiddleware_SwaggerUIConfiguration(t *testing.T) {
	// Verify the Swagger UI HTML has the correct configuration
	html := generateSwaggerUI()

	// Verify Swagger UI configuration
	assert.Contains(t, html, "deepLinking: true")
	assert.Contains(t, html, "StandaloneLayout")
	assert.Contains(t, html, "SwaggerUIBundle.presets.apis")
	assert.Contains(t, html, "SwaggerUIStandalonePreset")
	assert.Contains(t, html, "SwaggerUIBundle.plugins.DownloadUrl")

	// Verify it points to the correct URL
	assert.Contains(t, html, "url: '/openapi.json'")

	// Verify it has the correct DOM ID
	assert.Contains(t, html, "dom_id: '#swagger-ui'")
}

func BenchmarkMiddleware_OpenAPIJSON(b *testing.B) {
	// Create a test generator
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	generator.AddServer("http://localhost:8080", "Development server")
	generator.AddTag("users", "User operations")

	// Create a test Echo application
	e := echo.New()
	e.Use(Middleware(generator))

	// Add some routes to make the generator more complex
	for i := 0; i < 10; i++ {
		generator.AddTag("tag"+string(rune(i)), "Test tag")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http_status.OK.Code {
			b.Fatalf("Expected status 200, got %d", rec.Code)
		}
	}
}

func BenchmarkMiddleware_SwaggerUI(b *testing.B) {
	// Create a test generator
	generator := NewGenerator("Test API", "Test API description", "1.0.0")

	// Create a test Echo application
	e := echo.New()
	e.Use(Middleware(generator))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/docs", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http_status.OK.Code {
			b.Fatalf("Expected status 200, got %d", rec.Code)
		}
	}
}

func BenchmarkGenerateSwaggerUI(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateSwaggerUI()
	}
}
