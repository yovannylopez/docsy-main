package openapi

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs for the tests
type User struct {
	ID       int            `json:"id" validate:"required"`
	Name     string         `json:"name" validate:"required,min=1,max=100"`
	Email    string         `json:"email" validate:"required,email"`
	Age      int            `json:"age" validate:"min=0,max=150"`
	Active   bool           `json:"active"`
	Created  time.Time      `json:"created"`
	Metadata map[string]any `json:"metadata"`
}

type Product struct {
	ID          int      `json:"id" validate:"required"`
	Name        string   `json:"name" validate:"required,min=1,max=200"`
	Price       float64  `json:"price" validate:"min=0"`
	Category    string   `json:"category" validate:"oneof=electronics clothing books"`
	Tags        []string `json:"tags"`
	Description string   `json:"description" validate:"max=1000"`
	InStock     bool     `json:"in_stock"`
}

type Address struct {
	Street  string `json:"street" validate:"required"`
	City    string `json:"city" validate:"required"`
	Country string `json:"country" validate:"required"`
	ZipCode string `json:"zip_code" validate:"pattern=^[0-9]{5}$"`
}

type Order struct {
	ID       int       `json:"id" validate:"required"`
	UserID   int       `json:"user_id" validate:"required"`
	Products []Product `json:"products"`
	Total    float64   `json:"total" validate:"min=0"`
	Status   string    `json:"status" validate:"oneof=pending paid shipped delivered"`
	Address  Address   `json:"address"`
}

func TestNewSchemaGenerator(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	assert.NotNil(t, schemaGenerator)
	assert.NotNil(t, schemaGenerator.generator)
	assert.Equal(t, generator, schemaGenerator.generator)
}

func TestSchemaGenerator_GenerateSchemaFromStruct(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	tests := []struct {
		name       string
		structType any
		expected   *Schema
	}{
		{
			name:       "User struct",
			structType: User{},
			expected: &Schema{
				Type: "object",
				Properties: map[string]*Schema{
					"id": {
						Type: "integer",
					},
					"name": {
						Type:      "string",
						MinLength: intPtr(1),
						MaxLength: intPtr(100),
					},
					"email": {
						Type:   "string",
						Format: "email",
					},
					"age": {
						Type:    "integer",
						Minimum: floatPtr(0),
						Maximum: floatPtr(150),
					},
					"active": {
						Type: "boolean",
					},
					"created": {
						Type:   "string",
						Format: "date-time",
					},
					"metadata": {
						Type: "object",
						AdditionalProperties: &Schema{
							Type: "string",
						},
					},
				},
				Required: []string{"id", "name", "email"},
			},
		},
		{
			name:       "Product struct",
			structType: Product{},
			expected: &Schema{
				Type: "object",
				Properties: map[string]*Schema{
					"id": {
						Type: "integer",
					},
					"name": {
						Type:      "string",
						MinLength: intPtr(1),
						MaxLength: intPtr(200),
					},
					"price": {
						Type:    "number",
						Minimum: floatPtr(0),
					},
					"category": {
						Type: "string",
						Enum: []any{"electronics", "clothing", "books"},
					},
					"tags": {
						Type:  "array",
						Items: &Schema{Type: "string"},
					},
					"description": {
						Type:      "string",
						MaxLength: intPtr(1000),
					},
					"in_stock": {
						Type: "boolean",
					},
				},
				Required: []string{"id", "name"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := schemaGenerator.GenerateSchemaFromStruct(tt.name, tt.structType)

			assert.NotNil(t, schema)
			assert.Equal(t, "object", schema.Type)
			assert.NotEmpty(t, schema.Properties)
			assert.NotEmpty(t, schema.Required)

			// Verify the schema was added to components
			assert.Contains(t, generator.spec.Components.Schemas, tt.name)
		})
	}
}

func TestSchemaGenerator_GenerateSchemaFromStruct_SimpleTypes(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	tests := []struct {
		name       string
		structType any
		expected   *Schema
	}{
		{
			name:       "string",
			structType: "",
			expected:   &Schema{Type: "string"},
		},
		{
			name:       "int",
			structType: 0,
			expected:   &Schema{Type: "integer"},
		},
		{
			name:       "float64",
			structType: 0.0,
			expected:   &Schema{Type: "number"},
		},
		{
			name:       "bool",
			structType: false,
			expected:   &Schema{Type: "boolean"},
		},
		{
			name:       "map",
			structType: map[string]any{},
			expected: &Schema{
				Type: "object",
				AdditionalProperties: &Schema{
					Type: "string",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := schemaGenerator.GenerateSchemaFromStruct(tt.name, tt.structType)

			assert.NotNil(t, schema)
			assert.Equal(t, tt.expected.Type, schema.Type)
		})
	}
}

func TestSchemaGenerator_GenerateSchemaFromStruct_Pointer(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	user := &User{
		ID:     1,
		Name:   "John Doe",
		Email:  "john@example.com",
		Active: true,
	}

	schema := schemaGenerator.GenerateSchemaFromStruct("User", user)

	assert.NotNil(t, schema)
	assert.Equal(t, "object", schema.Type)
	assert.NotEmpty(t, schema.Properties)
	assert.Contains(t, schema.Properties, "id")
	assert.Contains(t, schema.Properties, "name")
	assert.Contains(t, schema.Properties, "email")
}

func TestSchemaGenerator_GenerateSchemaFromStruct_WithTags(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	type TestStruct struct {
		Field1 string `json:"field_1" validate:"required,min=1,max=50" description:"Test field"`
		Field2 int    `json:"field_2" validate:"min=0,max=100" example:"42"`
		Field3 bool   `json:"-"` // omitted field
	}

	schema := schemaGenerator.GenerateSchemaFromStruct("TestStruct", TestStruct{})

	assert.NotNil(t, schema)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "field_1")
	assert.Contains(t, schema.Properties, "field_2")
	assert.NotContains(t, schema.Properties, "field_3") // omitted field
	assert.Contains(t, schema.Required, "field_1")
}

func TestSchemaGenerator_GenerateRequestBodies(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	dtos := map[string]any{
		"CreateUser":    User{},
		"CreateProduct": Product{},
	}

	schemaGenerator.GenerateRequestBodies(dtos)

	// Verify request bodies were generated
	assert.Contains(t, generator.spec.Components.RequestBodies, "CreateUser")
	assert.Contains(t, generator.spec.Components.RequestBodies, "CreateProduct")

	// Verify request bodies have the correct content
	createUserBody := generator.spec.Components.RequestBodies["CreateUser"]
	assert.True(t, createUserBody.Required)
	assert.Contains(t, createUserBody.Content, "application/json")
	assert.NotNil(t, createUserBody.Content["application/json"].Schema)

	createProductBody := generator.spec.Components.RequestBodies["CreateProduct"]
	assert.True(t, createProductBody.Required)
	assert.Contains(t, createProductBody.Content, "application/json")
	assert.NotNil(t, createProductBody.Content["application/json"].Schema)
}

func TestSchemaGenerator_GenerateResponses(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	responses := map[string]any{
		"UserResponse":    User{},
		"ProductResponse": Product{},
		"OrderResponse":   Order{},
	}

	schemaGenerator.GenerateResponses(responses)

	// Verify responses were generated
	assert.Contains(t, generator.spec.Components.Responses, "UserResponse")
	assert.Contains(t, generator.spec.Components.Responses, "ProductResponse")
	assert.Contains(t, generator.spec.Components.Responses, "OrderResponse")

	// Verify responses have the correct content
	userResponse := generator.spec.Components.Responses["UserResponse"]
	assert.Equal(t, "UserResponse response", userResponse.Description)
	assert.Contains(t, userResponse.Content, "application/json")
	assert.NotNil(t, userResponse.Content["application/json"].Schema)

	productResponse := generator.spec.Components.Responses["ProductResponse"]
	assert.Equal(t, "ProductResponse response", productResponse.Description)
	assert.Contains(t, productResponse.Content, "application/json")
	assert.NotNil(t, productResponse.Content["application/json"].Schema)
}

// Responses declared as map (e.g. ErrorResponse in modules) must be registered in
// components.schemas so that $ref "#/components/schemas/ErrorResponse" resolves.
func TestSchemaGenerator_GenerateResponses_MapBodyRegisteredAsSchema(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	responses := map[string]any{
		"ErrorResponse": map[string]any{"error": "string", "message": "string"},
	}

	schemaGenerator.GenerateResponses(responses)

	assert.Contains(t, generator.spec.Components.Schemas, "ErrorResponse")
	assert.Equal(t, "object", generator.spec.Components.Schemas["ErrorResponse"].Type)
	assert.Contains(t, generator.spec.Components.Responses, "ErrorResponse")
}

func TestSchemaGenerator_ParseValidations(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	tests := []struct {
		name     string
		validate string
		expected map[string]string
	}{
		{
			name:     "empty validation",
			validate: "",
			expected: map[string]string{},
		},
		{
			name:     "simple validation",
			validate: "required",
			expected: map[string]string{
				"required": "",
			},
		},
		{
			name:     "validation with value",
			validate: "min=1,max=100",
			expected: map[string]string{
				"min": "1",
				"max": "100",
			},
		},
		{
			name:     "mixed validation",
			validate: "required,min=1,max=100,email",
			expected: map[string]string{
				"required": "",
				"min":      "1",
				"max":      "100",
				"email":    "",
			},
		},
		{
			name:     "validation with spaces",
			validate: " required , min=1 , max=100 ",
			expected: map[string]string{
				"required": "",
				"min":      "1",
				"max":      "100",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schemaGenerator.parseValidations(tt.validate)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSchemaGenerator_ApplyValidations(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	tests := []struct {
		name        string
		schema      *Schema
		validations map[string]string
		expected    *Schema
	}{
		{
			name:   "string validation",
			schema: &Schema{Type: "string"},
			validations: map[string]string{
				"min":     "1",
				"max":     "100",
				"email":   "",
				"pattern": "^[a-zA-Z]+$",
			},
			expected: &Schema{
				Type:      "string",
				MinLength: intPtr(1),
				MaxLength: intPtr(100),
				Format:    "email",
				Pattern:   "^[a-zA-Z]+$",
			},
		},
		{
			name:   "integer validation",
			schema: &Schema{Type: "integer"},
			validations: map[string]string{
				"min": "0",
				"max": "150",
			},
			expected: &Schema{
				Type:    "integer",
				Minimum: floatPtr(0),
				Maximum: floatPtr(150),
			},
		},
		{
			name:   "array validation",
			schema: &Schema{Type: "array"},
			validations: map[string]string{
				"min": "1",
				"max": "10",
			},
			expected: &Schema{
				Type:     "array",
				MinItems: intPtr(1),
				MaxItems: intPtr(10),
			},
		},
		{
			name:   "enum validation",
			schema: &Schema{Type: "string"},
			validations: map[string]string{
				"oneof": "pending paid shipped",
			},
			expected: &Schema{
				Type: "string",
				Enum: []any{"pending", "paid", "shipped"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaGenerator.applyValidations(tt.schema, tt.validations)

			if tt.expected.MinLength != nil {
				assert.Equal(t, tt.expected.MinLength, tt.schema.MinLength)
			}
			if tt.expected.MaxLength != nil {
				assert.Equal(t, tt.expected.MaxLength, tt.schema.MaxLength)
			}
			if tt.expected.Minimum != nil {
				assert.Equal(t, tt.expected.Minimum, tt.schema.Minimum)
			}
			if tt.expected.Maximum != nil {
				assert.Equal(t, tt.expected.Maximum, tt.schema.Maximum)
			}
			if tt.expected.MinItems != nil {
				assert.Equal(t, tt.expected.MinItems, tt.schema.MinItems)
			}
			if tt.expected.MaxItems != nil {
				assert.Equal(t, tt.expected.MaxItems, tt.schema.MaxItems)
			}
			if tt.expected.Format != "" {
				assert.Equal(t, tt.expected.Format, tt.schema.Format)
			}
			if tt.expected.Pattern != "" {
				assert.Equal(t, tt.expected.Pattern, tt.schema.Pattern)
			}
			if tt.expected.Enum != nil {
				assert.Equal(t, tt.expected.Enum, tt.schema.Enum)
			}
		})
	}
}

func TestSchemaGenerator_ExtractFieldName(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	tests := []struct {
		name      string
		fieldName string
		jsonTag   string
		expected  string
	}{
		{
			name:      "without JSON tag",
			fieldName: "UserName",
			jsonTag:   "",
			expected:  "UserName",
		},
		{
			name:      "with simple JSON tag",
			fieldName: "UserName",
			jsonTag:   "user_name",
			expected:  "user_name",
		},
		{
			name:      "with JSON tag with options",
			fieldName: "UserName",
			jsonTag:   "user_name,omitempty",
			expected:  "user_name",
		},
		{
			name:      "omitted field",
			fieldName: "InternalField",
			jsonTag:   "-",
			expected:  "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schemaGenerator.extractFieldName(tt.fieldName, tt.jsonTag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSchemaGenerator_IsFieldRequired(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	tests := []struct {
		name     string
		validate string
		expected bool
	}{
		{
			name:     "without validations",
			validate: "",
			expected: false,
		},
		{
			name:     "with required",
			validate: "required",
			expected: true,
		},
		{
			name:     "with required and other validations",
			validate: "required,min=1,max=100",
			expected: true,
		},
		{
			name:     "without required",
			validate: "min=1,max=100",
			expected: false,
		},
		{
			name:     "with spaces",
			validate: " required , min=1 ",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary struct field for the test
			field := reflect.StructField{
				Tag: reflect.StructTag("validate:\"" + tt.validate + "\""),
			}

			result := schemaGenerator.isFieldRequired(field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSchemaGenerator_ParseInt(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	tests := []struct {
		name     string
		input    string
		expected int
		valid    bool
	}{
		{
			name:     "valid number",
			input:    "42",
			expected: 42,
			valid:    true,
		},
		{
			name:     "zero",
			input:    "0",
			expected: 0,
			valid:    true,
		},
		{
			name:     "negative number",
			input:    "-10",
			expected: -10,
			valid:    true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
			valid:    false,
		},
		{
			name:     "non-numeric string",
			input:    "abc",
			expected: 0,
			valid:    false,
		},
		{
			name:     "float as string",
			input:    "3.14",
			expected: 3,
			valid:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, valid := schemaGenerator.parseInt(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestSchemaGenerator_ParseFloat(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	tests := []struct {
		name     string
		input    string
		expected float64
		valid    bool
	}{
		{
			name:     "valid float",
			input:    "3.14",
			expected: 3.14,
			valid:    true,
		},
		{
			name:     "integer as float",
			input:    "42",
			expected: 42.0,
			valid:    true,
		},
		{
			name:     "zero",
			input:    "0",
			expected: 0.0,
			valid:    true,
		},
		{
			name:     "negative number",
			input:    "-10.5",
			expected: -10.5,
			valid:    true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0.0,
			valid:    false,
		},
		{
			name:     "non-numeric string",
			input:    "abc",
			expected: 0.0,
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, valid := schemaGenerator.parseFloat(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestSchemaGenerator_GenerateBaseSchema(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	tests := []struct {
		name       string
		structType any
		expected   *Schema
	}{
		{
			name:       "string",
			structType: "",
			expected:   &Schema{Type: "string"},
		},
		{
			name:       "int",
			structType: 0,
			expected:   &Schema{Type: "integer"},
		},
		{
			name:       "float64",
			structType: 0.0,
			expected:   &Schema{Type: "number"},
		},
		{
			name:       "bool",
			structType: false,
			expected:   &Schema{Type: "boolean"},
		},
		{
			name:       "slice",
			structType: []string{},
			expected: &Schema{
				Type:  "array",
				Items: &Schema{Type: "string"},
			},
		},
		{
			name:       "time.Time",
			structType: time.Time{},
			expected: &Schema{
				Type:   "string",
				Format: "date-time",
			},
		},
		{
			name:       "map",
			structType: map[string]any{},
			expected: &Schema{
				Type: "object",
				AdditionalProperties: &Schema{
					Type: "string",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reflectType := reflect.TypeOf(tt.structType)
			schema := schemaGenerator.generateBaseSchema(reflectType, map[string]string{})

			assert.NotNil(t, schema)
			assert.Equal(t, tt.expected.Type, schema.Type)
			if tt.expected.Format != "" {
				assert.Equal(t, tt.expected.Format, schema.Format)
			}
			if tt.expected.Items != nil {
				assert.NotNil(t, schema.Items)
				assert.Equal(t, tt.expected.Items.Type, schema.Items.Type)
			}
			if tt.expected.AdditionalProperties != nil {
				assert.NotNil(t, schema.AdditionalProperties)
				assert.Equal(t, tt.expected.AdditionalProperties.Type, schema.AdditionalProperties.Type)
			}
		})
	}
}

func TestSchemaGenerator_ComplexStruct(t *testing.T) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	// Test with a complex struct that includes nesting
	schema := schemaGenerator.GenerateSchemaFromStruct("Order", Order{})

	assert.NotNil(t, schema)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "id")
	assert.Contains(t, schema.Properties, "user_id")
	assert.Contains(t, schema.Properties, "products")
	assert.Contains(t, schema.Properties, "total")
	assert.Contains(t, schema.Properties, "status")
	assert.Contains(t, schema.Properties, "address")
	addr := schema.Properties["address"]
	require.NotNil(t, addr)
	assert.Equal(t, "object", addr.Type)
	assert.Contains(t, addr.Properties, "street")
	assert.Contains(t, addr.Properties, "city")
	assert.Contains(t, addr.Properties, "zip_code")

	// Verify the schema was added to components
	assert.Contains(t, generator.spec.Components.Schemas, "Order")
}

func BenchmarkSchemaGenerator_GenerateSchemaFromStruct(b *testing.B) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = schemaGenerator.GenerateSchemaFromStruct("User", User{})
	}
}

func BenchmarkSchemaGenerator_GenerateRequestBodies(b *testing.B) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	dtos := map[string]any{
		"CreateUser":    User{},
		"CreateProduct": Product{},
		"CreateOrder":   Order{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schemaGenerator.GenerateRequestBodies(dtos)
	}
}

func BenchmarkSchemaGenerator_GenerateResponses(b *testing.B) {
	generator := NewGenerator("Test API", "Test API description", "1.0.0")
	schemaGenerator := NewSchemaGenerator(generator)

	responses := map[string]any{
		"UserResponse":    User{},
		"ProductResponse": Product{},
		"OrderResponse":   Order{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schemaGenerator.GenerateResponses(responses)
	}
}

// Helper functions to create pointers
func floatPtr(f float64) *float64 {
	return &f
}
