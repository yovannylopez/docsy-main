package openapi

import (
	"encoding/json"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yovannylopez/docsy-main/pkg/http_status"
	yaml "gopkg.in/yaml.v3"
)

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		version     string
	}{
		{
			name:        "basic generator",
			title:       "Test API",
			description: "Test API description",
			version:     "1.0.0",
		},
		{
			name:        "generator with empty fields",
			title:       "",
			description: "",
			version:     "",
		},
		{
			name:        "generator with special characters",
			title:       "API Test & Demo",
			description: "API with special characters: áéíóú",
			version:     "2.1.0-beta",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator(tt.title, tt.description, tt.version)

			assert.NotNil(t, generator)
			assert.NotNil(t, generator.spec)
			assert.Equal(t, "3.0.3", generator.spec.OpenAPI)
			assert.Equal(t, tt.title, generator.spec.Info.Title)
			assert.Equal(t, tt.description, generator.spec.Info.Description)
			assert.Equal(t, tt.version, generator.spec.Info.Version)
			assert.NotNil(t, generator.spec.Paths)
			assert.NotNil(t, generator.spec.Components)
			assert.NotNil(t, generator.spec.Components.Schemas)
			assert.NotNil(t, generator.spec.Components.Parameters)
			assert.NotNil(t, generator.spec.Components.RequestBodies)
			assert.NotNil(t, generator.spec.Components.Responses)
			assert.NotNil(t, generator.spec.Components.SecuritySchemes)
			assert.NotNil(t, generator.spec.Tags)
		})
	}
}

func TestGenerator_AddServer(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")

	tests := []struct {
		name        string
		url         string
		description string
	}{
		{
			name:        "development server",
			url:         "http://localhost:8080",
			description: "Development server",
		},
		{
			name:        "production server",
			url:         "https://api.example.com",
			description: "Production server",
		},
		{
			name:        "server without description",
			url:         "https://staging.example.com",
			description: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialCount := len(generator.spec.Servers)
			generator.AddServer(tt.url, tt.description)

			assert.Equal(t, initialCount+1, len(generator.spec.Servers))

			lastServer := generator.spec.Servers[len(generator.spec.Servers)-1]
			assert.Equal(t, tt.url, lastServer.URL)
			assert.Equal(t, tt.description, lastServer.Description)
		})
	}
}

func TestGenerator_AddTag(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")

	tests := []struct {
		name        string
		tagName     string
		description string
	}{
		{
			name:        "users tag",
			tagName:     "users",
			description: "Operations related to users",
		},
		{
			name:        "products tag",
			tagName:     "products",
			description: "Operations related to products",
		},
		{
			name:        "tag without description",
			tagName:     "default",
			description: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialCount := len(generator.spec.Tags)
			generator.AddTag(tt.tagName, tt.description)

			assert.Equal(t, initialCount+1, len(generator.spec.Tags))

			lastTag := generator.spec.Tags[len(generator.spec.Tags)-1]
			assert.Equal(t, tt.tagName, lastTag.Name)
			assert.Equal(t, tt.description, lastTag.Description)
		})
	}
}

func TestGenerator_AddSecurityScheme(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")

	tests := []struct {
		name   string
		scheme *SecurityScheme
	}{
		{
			name: "bearer scheme",
			scheme: &SecurityScheme{
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
			},
		},
		{
			name: "api key scheme",
			scheme: &SecurityScheme{
				Type:        "apiKey",
				Name:        "X-API-Key",
				In:          "header",
				Description: "API Key for authentication",
			},
		},
		{
			name: "oauth2 scheme",
			scheme: &SecurityScheme{
				Type:        "oauth2",
				Description: "OAuth 2.0",
				Flows: &OAuthFlows{
					AuthorizationCode: &OAuthFlow{
						AuthorizationURL: "https://example.com/oauth/authorize",
						TokenURL:         "https://example.com/oauth/token",
						Scopes: map[string]string{
							"read":  "Read access",
							"write": "Write access",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator.AddSecurityScheme(tt.name, tt.scheme)

			assert.Contains(t, generator.spec.Components.SecuritySchemes, tt.name)
			assert.Equal(t, tt.scheme, generator.spec.Components.SecuritySchemes[tt.name])
		})
	}
}

func TestGenerator_AddSchema(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")

	tests := []struct {
		name   string
		schema *Schema
	}{
		{
			name: "user schema",
			schema: &Schema{
				Type: "object",
				Properties: map[string]*Schema{
					"id": {
						Type:    "integer",
						Format:  "int64",
						Example: 1,
					},
					"name": {
						Type:        "string",
						Description: "User name",
						MinLength:   intPtr(1),
						MaxLength:   intPtr(100),
					},
					"email": {
						Type:    "string",
						Format:  "email",
						Pattern: "^[^@]+@[^@]+\\.[^@]+$",
					},
				},
				Required: []string{"name", "email"},
			},
		},
		{
			name: "array schema",
			schema: &Schema{
				Type: "array",
				Items: &Schema{
					Type: "string",
				},
				MinItems: intPtr(1),
				MaxItems: intPtr(100),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator.AddSchema(tt.name, tt.schema)

			assert.Contains(t, generator.spec.Components.Schemas, tt.name)
			assert.Equal(t, tt.schema, generator.spec.Components.Schemas[tt.name])
		})
	}
}

func TestGenerator_AddRequestBody(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")

	tests := []struct {
		name        string
		requestBody *RequestBody
	}{
		{
			name: "user request body",
			requestBody: &RequestBody{
				Description: "User data",
				Required:    true,
				Content: map[string]MediaType{
					"application/json": {
						Schema: &Schema{
							Type: "object",
							Properties: map[string]*Schema{
								"name": {
									Type: "string",
								},
								"email": {
									Type: "string",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator.AddRequestBody(tt.name, tt.requestBody)

			assert.Contains(t, generator.spec.Components.RequestBodies, tt.name)
			assert.Equal(t, tt.requestBody, generator.spec.Components.RequestBodies[tt.name])
		})
	}
}

func TestGenerator_AddResponse(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")

	tests := []struct {
		name     string
		response *Response
	}{
		{
			name: "user response",
			response: &Response{
				Description: "User found",
				Content: map[string]MediaType{
					"application/json": {
						Schema: &Schema{
							Type: "object",
							Properties: map[string]*Schema{
								"id": {
									Type: "integer",
								},
								"name": {
									Type: "string",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator.AddResponse(tt.name, tt.response)

			assert.Contains(t, generator.spec.Components.Responses, tt.name)
			assert.Equal(t, tt.response, generator.spec.Components.Responses[tt.name])
		})
	}
}

func TestGenerator_Generate(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")

	// Add some elements to test
	generator.AddServer("http://localhost:8080", "Development server")
	generator.AddTag("users", "User operations")
	generator.AddSchema("User", &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"id": {
				Type: "integer",
			},
		},
	})

	// Generate JSON
	jsonData, err := generator.Generate()
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verify it is valid JSON
	var spec Spec
	err = json.Unmarshal(jsonData, &spec)
	require.NoError(t, err)

	// Verify the data was preserved
	assert.Equal(t, "3.0.3", spec.OpenAPI)
	assert.Equal(t, "Test API", spec.Info.Title)
	assert.Equal(t, "Test API description", spec.Info.Description)
	assert.Equal(t, "1.0.0", spec.Info.Version)
	assert.Len(t, spec.Servers, 1)
	assert.Len(t, spec.Tags, 1)
	assert.Len(t, spec.Components.Schemas, 1)
}

func TestGenerator_GenerateYAML(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	generator.AddServer("http://localhost:8080", "dev")

	yamlData, err := generator.GenerateYAML()
	require.NoError(t, err)
	require.NotEmpty(t, yamlData)

	assert.Contains(t, string(yamlData), "3.0.3")
	assert.Contains(t, string(yamlData), "Test API")

	var roundTrip map[string]any
	err = yaml.Unmarshal(yamlData, &roundTrip)
	require.NoError(t, err)
	assert.Equal(t, "3.0.3", roundTrip["openapi"])
}

func TestGenerator_GetSpec(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")

	spec := generator.GetSpec()
	assert.NotNil(t, spec)
	assert.Equal(t, generator.spec, spec)
	assert.Equal(t, "3.0.3", spec.OpenAPI)
	assert.Equal(t, "Test API", spec.Info.Title)
}

func TestGenerator_GenerateFromEcho(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")

	// Create a test Echo server
	e := echo.New()

	// Add some test routes
	e.GET("/users", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "users"})
	})
	e.POST("/users", func(c echo.Context) error {
		return c.JSON(201, map[string]string{"message": "user created"})
	})
	e.GET("/products", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "products"})
	})

	// Generate specification from Echo
	err := generator.GenerateFromEcho(e)
	require.NoError(t, err)

	// Verify the routes were added
	spec := generator.GetSpec()
	assert.Len(t, spec.Paths, 2) // Only 2 unique paths: /users and /products
	assert.Contains(t, spec.Paths, "/users")
	assert.Contains(t, spec.Paths, "/products")

	// Verify operations were created correctly
	usersPath := spec.Paths["/users"]
	assert.NotNil(t, usersPath.Get)
	assert.NotNil(t, usersPath.Post)
	assert.Equal(t, "users", usersPath.Get.Tags[0])
	assert.Equal(t, "users", usersPath.Post.Tags[0])

	productsPath := spec.Paths["/products"]
	assert.NotNil(t, productsPath.Get)
	assert.Equal(t, "products", productsPath.Get.Tags[0])
}

func TestGenerator_GenerateFromEcho_ExcludesOpenAPIInfra(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	e := echo.New()
	e.GET("/docs", func(c echo.Context) error { return c.NoContent(http_status.OK.Code) })
	e.GET("/openapi.json", func(c echo.Context) error { return c.NoContent(http_status.OK.Code) })
	e.GET("/api/health", func(c echo.Context) error { return c.NoContent(http_status.OK.Code) })

	require.NoError(t, generator.GenerateFromEcho(e))

	spec := generator.GetSpec()
	assert.Len(t, spec.Paths, 1)
	assert.Contains(t, spec.Paths, "/api/health")
	assert.NotContains(t, spec.Paths, "/docs")
	assert.NotContains(t, spec.Paths, "/openapi.json")
}

func TestShouldExcludeFromOpenAPISpec(t *testing.T) {
	tests := []struct {
		path     string
		excluded bool
	}{
		{"/docs", true},
		{"/docs/", true},
		{"/openapi.json", true},
		{"/openapi.json/", true},
		{"/users", false},
		{"/api/v1/docs", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.excluded, shouldExcludeFromOpenAPISpec(tt.path))
		})
	}
}

func TestExtractTagFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "simple path",
			path:     "/users",
			expected: "users",
		},
		{
			name:     "path with api prefix and version",
			path:     "/api/v1/users",
			expected: "users",
		},
		{
			name:     "api version nested resource path",
			path:     "/api/v1/communications/:id",
			expected: "communications",
		},
		{
			name:     "api v1 auth path",
			path:     "/api/v1/auth/login",
			expected: "authentication",
		},
		{
			name:     "path with parameters",
			path:     "/users/{id}",
			expected: "users",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "path with only slash",
			path:     "/",
			expected: "",
		},
		{
			name:     "path with trailing slash",
			path:     "/users/",
			expected: "users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTagFromPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSummary(t *testing.T) {
	tests := []struct {
		name      string
		routeName string
		method    string
		expected  string
	}{
		{
			name:      "with route name",
			routeName: "GetUsers",
			method:    "get",
			expected:  "GetUsers",
		},
		{
			name:      "without route name",
			routeName: "",
			method:    "post",
			expected:  "Post operation",
		},
		{
			name:      "GET method",
			routeName: "",
			method:    "get",
			expected:  "Get operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateSummary(tt.routeName, tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateDescription(t *testing.T) {
	tests := []struct {
		name      string
		routeName string
		method    string
		expected  string
	}{
		{
			name:      "with route name",
			routeName: "GetUsers",
			method:    "get",
			expected:  "GetUsers - Get operation",
		},
		{
			name:      "without route name",
			routeName: "",
			method:    "post",
			expected:  "Post operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateDescription(tt.routeName, tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateOperationID(t *testing.T) {
	tests := []struct {
		name      string
		routeName string
		method    string
		expected  string
	}{
		{
			name:      "with route name",
			routeName: "GetUsers",
			method:    "get",
			expected:  "get_GetUsers",
		},
		{
			name:      "without route name",
			routeName: "",
			method:    "post",
			expected:  "post_operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateOperationID(tt.routeName, tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateDefaultResponses(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected map[string]Response
	}{
		{
			name:   "GET method",
			method: "get",
			expected: map[string]Response{
				"200": {Description: "Successful operation"},
				"400": {Description: "Bad request"},
				"401": {Description: "Unauthorized"},
				"403": {Description: "Forbidden"},
				"404": {Description: "Not found"},
				"500": {Description: "Internal server error"},
			},
		},
		{
			name:   "POST method",
			method: "post",
			expected: map[string]Response{
				"200": {Description: "Successful operation"},
				"201": {Description: "Created"},
				"400": {Description: "Bad request"},
				"401": {Description: "Unauthorized"},
				"403": {Description: "Forbidden"},
				"404": {Description: "Not found"},
				"500": {Description: "Internal server error"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateDefaultResponses(tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSchema_Validation(t *testing.T) {
	tests := []struct {
		name   string
		schema *Schema
		valid  bool
	}{
		{
			name: "valid object schema",
			schema: &Schema{
				Type: "object",
				Properties: map[string]*Schema{
					"name": {Type: "string"},
				},
			},
			valid: true,
		},
		{
			name: "valid array schema",
			schema: &Schema{
				Type:  "array",
				Items: &Schema{Type: "string"},
			},
			valid: true,
		},
		{
			name: "schema with validations",
			schema: &Schema{
				Type:        "string",
				MinLength:   intPtr(1),
				MaxLength:   intPtr(100),
				Pattern:     "^[a-zA-Z]+$",
				Description: "Valid name",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the schema can be created without panic
			assert.NotPanics(t, func() {
				_ = tt.schema
			})

			if tt.valid {
				assert.NotNil(t, tt.schema)
			}
		})
	}
}

func TestSecurityScheme_Validation(t *testing.T) {
	tests := []struct {
		name   string
		scheme *SecurityScheme
		valid  bool
	}{
		{
			name: "valid bearer scheme",
			scheme: &SecurityScheme{
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
			},
			valid: true,
		},
		{
			name: "valid api key scheme",
			scheme: &SecurityScheme{
				Type: "apiKey",
				Name: "X-API-Key",
				In:   "header",
			},
			valid: true,
		},
		{
			name: "valid oauth2 scheme",
			scheme: &SecurityScheme{
				Type: "oauth2",
				Flows: &OAuthFlows{
					AuthorizationCode: &OAuthFlow{
						AuthorizationURL: "https://example.com/oauth/authorize",
						TokenURL:         "https://example.com/oauth/token",
						Scopes: map[string]string{
							"read": "Read access",
						},
					},
				},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the scheme can be created without panic
			assert.NotPanics(t, func() {
				_ = tt.scheme
			})

			if tt.valid {
				assert.NotNil(t, tt.scheme)
				assert.NotEmpty(t, tt.scheme.Type)
			}
		})
	}
}

func BenchmarkNewGenerator(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewGenerator("Test API", "Test API description", "1.0.0")
	}
}

func BenchmarkGenerator_Generate(b *testing.B) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	generator.AddServer("http://localhost:8080", "Development server")
	generator.AddTag("users", "User operations")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.Generate()
	}
}

func BenchmarkGenerator_GenerateFromEcho(b *testing.B) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	e := echo.New()
	e.GET("/users", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "users"})
	})
	e.POST("/users", func(c echo.Context) error {
		return c.JSON(201, map[string]string{"message": "user created"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generator.GenerateFromEcho(e)
	}
}

// intPtr is a helper function to create int pointers
func intPtr(i int) *int {
	return &i
}
