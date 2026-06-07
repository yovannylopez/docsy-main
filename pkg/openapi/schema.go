package openapi

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// SchemaGenerator generates OpenAPI schemas from Go structs
type SchemaGenerator struct {
	generator *Generator
}

// NewSchemaGenerator creates a new schema generator
func NewSchemaGenerator(generator *Generator) *SchemaGenerator {
	return &SchemaGenerator{
		generator: generator,
	}
}

// GenerateSchemaFromStruct generates an OpenAPI schema from a Go struct
func (sg *SchemaGenerator) GenerateSchemaFromStruct(name string, structType any) *Schema {
	t := reflect.TypeOf(structType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Verify that it is a struct
	if t.Kind() != reflect.Struct {
		// Simple schema (e.g. map used as ErrorResponse in GenerateResponses).
		// Must be registered in components.schemas so that $ref "#/components/schemas/{name}" resolves.
		schema := sg.generateSimpleSchema(t)
		sg.generator.AddSchema(name, schema)

		return schema
	}

	return sg.generateSchemaFromType(name, t)
}

// generateSimpleSchema generates a simple schema for non-struct types
func (sg *SchemaGenerator) generateSimpleSchema(t reflect.Type) *Schema {
	switch t.Kind() {
	case reflect.Map:
		// For map[string]any, create an object schema
		return &Schema{
			Type: "object",
			AdditionalProperties: &Schema{
				Type: "string",
			},
		}
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Schema{Type: "integer"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	default:
		return &Schema{Type: "string"}
	}
}

// objectSchemaFromExportedFields builds type: object with properties from a struct (including nested).
func (sg *SchemaGenerator) objectSchemaFromExportedFields(t reflect.Type) *Schema {
	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
		Required:   []string{},
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldSchema := sg.generateFieldSchema(field)
		if fieldSchema == nil {
			continue
		}

		jsonTag := field.Tag.Get("json")
		fieldName := sg.extractFieldName(field.Name, jsonTag)
		if fieldName == "-" {
			continue
		}

		schema.Properties[fieldName] = fieldSchema
		if sg.isFieldRequired(field) {
			schema.Required = append(schema.Required, fieldName)
		}
	}

	return schema
}

// generateSchemaFromType generates a schema from a reflect type
func (sg *SchemaGenerator) generateSchemaFromType(name string, t reflect.Type) *Schema {
	if t.Kind() != reflect.Struct {
		return sg.generateSimpleSchema(t)
	}

	schema := sg.objectSchemaFromExportedFields(t)
	sg.generator.AddSchema(name, schema)

	return schema
}

// generateFieldSchema generates a schema for a specific field
func (sg *SchemaGenerator) generateFieldSchema(field reflect.StructField) *Schema {
	fieldType := field.Type

	// Handle pointers
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	// Get validations from the validate tag
	validateTag := field.Tag.Get("validate")
	validations := sg.parseValidations(validateTag)

	// Generate base schema according to the type
	schema := sg.generateBaseSchema(fieldType, validations)

	// Apply validations
	sg.applyValidations(schema, validations)

	// Add description if present
	if description := field.Tag.Get("description"); description != "" {
		schema.Description = description
	}

	// Add example if present
	if example := field.Tag.Get("example"); example != "" {
		schema.Example = example
	}

	return schema
}

// generateBaseSchema generates a base schema according to the Go type
func (sg *SchemaGenerator) generateBaseSchema(t reflect.Type, validations map[string]string) *Schema {
	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Schema{Type: "integer"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	case reflect.Slice, reflect.Array:
		itemSchema := sg.generateBaseSchema(t.Elem(), validations)
		return &Schema{
			Type:  "array",
			Items: itemSchema,
		}
	case reflect.Struct:
		// Handle special types
		switch t {
		case reflect.TypeOf(time.Time{}):
			return &Schema{
				Type:   "string",
				Format: "date-time",
			}
		default:
			return sg.objectSchemaFromExportedFields(t)
		}
	case reflect.Map:
		// For map[string]any, create an object schema
		return &Schema{
			Type: "object",
			AdditionalProperties: &Schema{
				Type: "string",
			},
		}
	default:
		return &Schema{Type: "string"}
	}
}

// parseValidations parses the validations from the validate tag
func (sg *SchemaGenerator) parseValidations(validateTag string) map[string]string {
	validations := make(map[string]string)

	if validateTag == "" {
		return validations
	}

	parts := strings.Split(validateTag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "=") {
			keyValue := strings.SplitN(part, "=", 2)
			validations[keyValue[0]] = keyValue[1]
		} else {
			validations[part] = ""
		}
	}

	return validations
}

// applyValidations applies the validations to the schema
func (sg *SchemaGenerator) applyValidations(schema *Schema, validations map[string]string) {
	for validation, value := range validations {
		switch validation {
		case "required":
			// Already handled at the struct level
		case "min":
			if schema.Type == "string" {
				if min, ok := sg.parseInt(value); ok {
					schema.MinLength = &min
				}
			} else if schema.Type == "integer" || schema.Type == "number" {
				if min, ok := sg.parseFloat(value); ok {
					schema.Minimum = &min
				}
			} else if schema.Type == "array" {
				if min, ok := sg.parseInt(value); ok {
					schema.MinItems = &min
				}
			}
		case "max":
			if schema.Type == "string" {
				if max, ok := sg.parseInt(value); ok {
					schema.MaxLength = &max
				}
			} else if schema.Type == "integer" || schema.Type == "number" {
				if max, ok := sg.parseFloat(value); ok {
					schema.Maximum = &max
				}
			} else if schema.Type == "array" {
				if max, ok := sg.parseInt(value); ok {
					schema.MaxItems = &max
				}
			}
		case "email":
			if schema.Type == "string" {
				schema.Format = "email"
			}
		case "oneof":
			if schema.Type == "string" {
				enumValues := strings.Split(value, " ")
				schema.Enum = make([]any, len(enumValues))
				for i, v := range enumValues {
					schema.Enum[i] = v
				}
			}
		case "pattern":
			if schema.Type == "string" {
				schema.Pattern = value
			}
		}
	}
}

// extractFieldName extracts the field name from the JSON tag
func (sg *SchemaGenerator) extractFieldName(fieldName, jsonTag string) string {
	if jsonTag == "" {
		return fieldName
	}

	parts := strings.Split(jsonTag, ",")
	return parts[0]
}

// isFieldRequired determines if a field is required
func (sg *SchemaGenerator) isFieldRequired(field reflect.StructField) bool {
	validateTag := field.Tag.Get("validate")
	if validateTag == "" {
		return false
	}

	validations := strings.Split(validateTag, ",")
	for _, validation := range validations {
		if strings.TrimSpace(validation) == "required" {
			return true
		}
	}

	return false
}

// parseInt parses a string to int
func (sg *SchemaGenerator) parseInt(s string) (int, bool) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err == nil
}

// parseFloat parses a string to float64
func (sg *SchemaGenerator) parseFloat(s string) (float64, bool) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err == nil
}

// GenerateRequestBodies generates request bodies from the DTOs
func (sg *SchemaGenerator) GenerateRequestBodies(dtos map[string]any) {
	for name, dto := range dtos {
		schema := sg.GenerateSchemaFromStruct(name, dto)

		requestBody := &RequestBody{
			Required: true,
			Content: map[string]MediaType{
				"application/json": {
					Schema: schema,
				},
			},
		}

		sg.generator.AddRequestBody(name, requestBody)
	}
}

// GenerateResponses generates responses from the DTOs
func (sg *SchemaGenerator) GenerateResponses(responses map[string]any) {
	for name, response := range responses {
		schema := sg.GenerateSchemaFromStruct(name, response)

		openapiResponse := &Response{
			Description: fmt.Sprintf("%s response", name),
			Content: map[string]MediaType{
				"application/json": {
					Schema: schema,
				},
			},
		}

		sg.generator.AddResponse(name, openapiResponse)
	}
}
