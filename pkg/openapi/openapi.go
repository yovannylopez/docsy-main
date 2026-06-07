// Package openapi builds OpenAPI 3.0 models, generates spec from Echo and helpers aligned with pkg/responses.
package openapi

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
	yaml "gopkg.in/yaml.v3"
)

// Spec represents the OpenAPI 3.0 specification
type Spec struct {
	OpenAPI    string              `json:"openapi"`
	Info       Info                `json:"info"`
	Servers    []Server            `json:"servers"`
	Paths      map[string]PathItem `json:"paths"`
	Components Components          `json:"components"`
	Tags       []Tag               `json:"tags"`
}

// Info contains information about the API
type Info struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Version     string  `json:"version"`
	Contact     Contact `json:"contact,omitempty"`
}

// Contact contains contact information
type Contact struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

// Server represents a server
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// PathItem represents a path in the API
type PathItem struct {
	Get     *Operation `json:"get,omitempty"`
	Post    *Operation `json:"post,omitempty"`
	Put     *Operation `json:"put,omitempty"`
	Delete  *Operation `json:"delete,omitempty"`
	Patch   *Operation `json:"patch,omitempty"`
	Options *Operation `json:"options,omitempty"`
	Head    *Operation `json:"head,omitempty"`
	Trace   *Operation `json:"trace,omitempty"`
}

// Operation represents an HTTP operation
type Operation struct {
	Tags        []string            `json:"tags,omitempty"`
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	OperationID string              `json:"operationId,omitempty"`
	Parameters  []Parameter         `json:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
	// Security: without omitempty to be able to serialize [] (without JWT) against ambiguous inheritance.
	// Typical values: []map[string][]string{{SecuritySchemeBearer: {}}} (JWT) or {} (public).
	Security []map[string][]string `json:"security"`
}

// Parameter represents a parameter
type Parameter struct {
	Name        string  `json:"name"`
	In          string  `json:"in"`
	Description string  `json:"description,omitempty"`
	Required    bool    `json:"required,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
	Example     any     `json:"example,omitempty"`
}

// RequestBody represents the body of a request
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"`
	Content     map[string]MediaType `json:"content"`
}

// Response represents a response
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
	Headers     map[string]Header    `json:"headers,omitempty"`
}

// MediaType represents a media type
type MediaType struct {
	Schema   *Schema            `json:"schema,omitempty"`
	Example  any                `json:"example,omitempty"`
	Examples map[string]Example `json:"examples,omitempty"`
}

// Header represents an HTTP header
type Header struct {
	Description string  `json:"description,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
}

// Example represents an example
type Example struct {
	Summary       string `json:"summary,omitempty"`
	Description   string `json:"description,omitempty"`
	Value         any    `json:"value,omitempty"`
	ExternalValue string `json:"externalValue,omitempty"`
}

// Components represents the reusable components
type Components struct {
	Schemas         map[string]*Schema         `json:"schemas,omitempty"`
	Parameters      map[string]*Parameter      `json:"parameters,omitempty"`
	RequestBodies   map[string]*RequestBody    `json:"requestBodies,omitempty"`
	Responses       map[string]*Response       `json:"responses,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty"`
}

// Schema represents a JSON schema
type Schema struct {
	Type                 string                 `json:"type,omitempty"`
	Format               string                 `json:"format,omitempty"`
	Description          string                 `json:"description,omitempty"`
	Properties           map[string]*Schema     `json:"properties,omitempty"`
	Required             []string               `json:"required,omitempty"`
	Items                *Schema                `json:"items,omitempty"`
	Ref                  string                 `json:"$ref,omitempty"`
	Example              any                    `json:"example,omitempty"`
	Enum                 []any                  `json:"enum,omitempty"`
	MinLength            *int                   `json:"minLength,omitempty"`
	MaxLength            *int                   `json:"maxLength,omitempty"`
	Pattern              string                 `json:"pattern,omitempty"`
	Minimum              *float64               `json:"minimum,omitempty"`
	Maximum              *float64               `json:"maximum,omitempty"`
	ExclusiveMinimum     bool                   `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum     bool                   `json:"exclusiveMaximum,omitempty"`
	MultipleOf           *float64               `json:"multipleOf,omitempty"`
	MinItems             *int                   `json:"minItems,omitempty"`
	MaxItems             *int                   `json:"maxItems,omitempty"`
	UniqueItems          bool                   `json:"uniqueItems,omitempty"`
	MinProperties        *int                   `json:"minProperties,omitempty"`
	MaxProperties        *int                   `json:"maxProperties,omitempty"`
	AdditionalProperties *Schema                `json:"additionalProperties,omitempty"`
	AllOf                []*Schema              `json:"allOf,omitempty"`
	AnyOf                []*Schema              `json:"anyOf,omitempty"`
	OneOf                []*Schema              `json:"oneOf,omitempty"`
	Not                  *Schema                `json:"not,omitempty"`
	Nullable             bool                   `json:"nullable,omitempty"`
	Discriminator        *Discriminator         `json:"discriminator,omitempty"`
	ReadOnly             bool                   `json:"readOnly,omitempty"`
	WriteOnly            bool                   `json:"writeOnly,omitempty"`
	XML                  *XML                   `json:"xml,omitempty"`
	ExternalDocs         *ExternalDocumentation `json:"externalDocs,omitempty"`
	Deprecated           bool                   `json:"deprecated,omitempty"`
}

// SecurityScheme represents a security scheme
type SecurityScheme struct {
	Type             string      `json:"type"`
	Description      string      `json:"description,omitempty"`
	Name             string      `json:"name,omitempty"`
	In               string      `json:"in,omitempty"`
	Scheme           string      `json:"scheme,omitempty"`
	BearerFormat     string      `json:"bearerFormat,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty"`
}

// OAuthFlows represents the OAuth flows
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

// OAuthFlow represents an OAuth flow
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// Tag represents a tag
type Tag struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
}

// XML represents XML information
type XML struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
	Attribute bool   `json:"attribute,omitempty"`
	Wrapped   bool   `json:"wrapped,omitempty"`
}

// Discriminator represents a discriminator
type Discriminator struct {
	PropertyName string            `json:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty"`
}

// ExternalDocumentation represents external documentation
type ExternalDocumentation struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}

// Generator generates OpenAPI specifications dynamically
type Generator struct {
	spec *Spec
}

// NewGenerator creates a new OpenAPI generator
func NewGenerator(title, description, version string) *Generator {
	return &Generator{
		spec: &Spec{
			OpenAPI: "3.0.3",
			Info: Info{
				Title:       title,
				Description: description,
				Version:     version,
			},
			Paths: make(map[string]PathItem),
			Components: Components{
				Schemas:         make(map[string]*Schema),
				Parameters:      make(map[string]*Parameter),
				RequestBodies:   make(map[string]*RequestBody),
				Responses:       make(map[string]*Response),
				SecuritySchemes: make(map[string]*SecurityScheme),
			},
			Tags: []Tag{},
		},
	}
}

// AddServer adds a server to the specification
func (g *Generator) AddServer(url, description string) {
	g.spec.Servers = append(g.spec.Servers, Server{
		URL:         url,
		Description: description,
	})
}

// AddTag adds a tag
func (g *Generator) AddTag(name, description string) {
	g.spec.Tags = append(g.spec.Tags, Tag{
		Name:        name,
		Description: description,
	})
}

// AddSecurityScheme adds a security scheme
func (g *Generator) AddSecurityScheme(name string, scheme *SecurityScheme) {
	g.spec.Components.SecuritySchemes[name] = scheme
}

// GenerateFromEcho generates the OpenAPI specification from an Echo server
func (g *Generator) GenerateFromEcho(e *echo.Echo) error {
	// Extract routes from Echo
	routes := e.Routes()

	for _, route := range routes {
		if shouldExcludeFromOpenAPISpec(route.Path) {
			continue
		}

		if err := g.addRoute(*route); err != nil {
			return fmt.Errorf("error adding route %s: %w", route.Path, err)
		}
	}

	return nil
}

// shouldExcludeFromOpenAPISpec avoids documenting internal Swagger/OpenAPI routes.
func shouldExcludeFromOpenAPISpec(path string) bool {
	p := strings.TrimSuffix(strings.TrimSpace(path), "/")
	switch p {
	case "/docs", "/openapi.json":
		return true
	default:
		return false
	}
}

// addRoute adds a route to the specification
func (g *Generator) addRoute(route echo.Route) error {
	method := strings.ToLower(route.Method)
	switch method {
	case "get", "post", "put", "delete", "patch", "options", "head", "trace":
	default:
		return nil
	}

	path := route.Path
	if _, exists := g.spec.Paths[path]; !exists {
		g.spec.Paths[path] = PathItem{}
	}

	pathItem := g.spec.Paths[path]

	operation := &Operation{
		Tags:        []string{extractTagFromPath(path)},
		Summary:     generateSummary(route.Name, method),
		Description: generateDescription(route.Name, method),
		OperationID: generateOperationID(route.Name, method),
		Responses:   generateDefaultResponses(method),
		Security:    []map[string][]string{{SecuritySchemeBearer: {}}},
	}

	switch method {
	case "get":
		pathItem.Get = operation
	case "post":
		pathItem.Post = operation
	case "put":
		pathItem.Put = operation
	case "delete":
		pathItem.Delete = operation
	case "patch":
		pathItem.Patch = operation
	case "options":
		pathItem.Options = operation
	case "head":
		pathItem.Head = operation
	case "trace":
		pathItem.Trace = operation
	}

	g.spec.Paths[path] = pathItem
	return nil
}

// AddSchema adds a schema to the components
func (g *Generator) AddSchema(name string, schema *Schema) {
	g.spec.Components.Schemas[name] = schema
}

// AddRequestBody adds a request body to the components
func (g *Generator) AddRequestBody(name string, requestBody *RequestBody) {
	g.spec.Components.RequestBodies[name] = requestBody
}

// AddResponse adds a response to the components
func (g *Generator) AddResponse(name string, response *Response) {
	g.spec.Components.Responses[name] = response
}

// Generate generates the OpenAPI specification as JSON
func (g *Generator) Generate() ([]byte, error) {
	return json.MarshalIndent(g.spec, "", "  ")
}

// GenerateYAML serializes the spec as YAML (via intermediate JSON to respect json tags in the model).
func (g *Generator) GenerateYAML() ([]byte, error) {
	j, err := json.Marshal(g.spec)
	if err != nil {
		return nil, fmt.Errorf("serializing spec to intermediate JSON: %w", err)
	}

	var doc any
	if err := json.Unmarshal(j, &doc); err != nil {
		return nil, fmt.Errorf("parsing intermediate JSON: %w", err)
	}

	out, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("serializing spec to YAML: %w", err)
	}

	return out, nil
}

// GetSpec returns the current OpenAPI specification
func (g *Generator) GetSpec() *Spec {
	return g.spec
}

// Helper functions
func extractTagFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	segments := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			segments = append(segments, p)
		}
	}
	if len(segments) == 0 {
		return ""
	}

	i := 0
	if segments[i] == "api" {
		i++
		if i >= len(segments) {
			return "default"
		}
	}
	if isAPIVersionSegment(segments[i]) {
		i++
		if i >= len(segments) {
			return "default"
		}
	}

	// Keep auth endpoints under a single OpenAPI tag (matches AddTag("authentication", ...) in bootstrap).
	if segments[i] == "auth" {
		return "authentication"
	}

	return segments[i]
}

// isAPIVersionSegment recognizes segments like v1, v2, v10 (not "versions" or "v").
func isAPIVersionSegment(s string) bool {
	if len(s) < 2 || s[0] != 'v' {
		return false
	}
	for _, r := range s[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

// titleizeHTTPMethod capitalizes the HTTP verb in ASCII (e.g. "get" -> "Get"); avoids deprecated strings.Title.
func titleizeHTTPMethod(method string) string {
	if method == "" {
		return method
	}
	return strings.ToUpper(method[:1]) + method[1:]
}

// generateSummary generates a summary for the operation
func generateSummary(name, method string) string {
	if name != "" {
		return name
	}
	return fmt.Sprintf("%s operation", titleizeHTTPMethod(method))
}

// generateDescription generates a description for the operation
func generateDescription(name, method string) string {
	if name != "" {
		return fmt.Sprintf("%s - %s operation", name, titleizeHTTPMethod(method))
	}
	return fmt.Sprintf("%s operation", titleizeHTTPMethod(method))
}

// generateOperationID generates an ID for the operation
func generateOperationID(name, method string) string {
	if name != "" {
		return fmt.Sprintf("%s_%s", method, name)
	}
	return fmt.Sprintf("%s_operation", method)
}

// generateDefaultResponses generates the default responses for the operation
func generateDefaultResponses(method string) map[string]Response {
	responses := map[string]Response{
		"200": {
			Description: "Successful operation",
		},
		"400": {
			Description: "Bad request",
		},
		"401": {
			Description: "Unauthorized",
		},
		"403": {
			Description: "Forbidden",
		},
		"404": {
			Description: "Not found",
		},
		"500": {
			Description: "Internal server error",
		},
	}

	// Add 201 for POST
	if method == "post" {
		responses["201"] = Response{
			Description: "Created",
		}
	}

	return responses
}
