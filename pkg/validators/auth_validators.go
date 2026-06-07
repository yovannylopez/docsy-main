package validators

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// PasswordValidator validates passwords with configurable policies
type PasswordValidator struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumbers bool
	RequireSymbols bool
	Required       bool
}

// Validate validates a password
func (v PasswordValidator) Validate(value any) error {
	if value == nil {
		if v.Required {
			return ValidationError{Message: "password required"}
		}
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return ValidationError{Message: "must be a string", Value: value}
	}

	if isBlank(str) {
		if v.Required {
			return ValidationError{Message: "password required", Value: value}
		}
		return nil
	}

	if v.MinLength > 0 && utf8.RuneCountInString(str) < v.MinLength {
		return ValidationError{Message: fmt.Sprintf("minimum length: %d characters", v.MinLength), Value: value}
	}

	if v.RequireUpper && !passwordHasUpperRe.MatchString(str) {
		return ValidationError{Message: "must contain at least one uppercase letter", Value: value}
	}

	if v.RequireLower && !passwordHasLowerRe.MatchString(str) {
		return ValidationError{Message: "must contain at least one lowercase letter", Value: value}
	}

	if v.RequireNumbers && !passwordHasNumberRe.MatchString(str) {
		return ValidationError{Message: "must contain at least one number", Value: value}
	}

	if v.RequireSymbols && !passwordHasSymbolRe.MatchString(str) {
		return ValidationError{Message: "must contain at least one special symbol", Value: value}
	}

	return nil
}

// PhoneValidator validates phone numbers
type PhoneValidator struct {
	Required bool
	Pattern  string // Custom regex pattern
}

// Validate validates a phone number
func (v PhoneValidator) Validate(value any) error {
	if value == nil {
		if v.Required {
			return ValidationError{Message: "phone required"}
		}
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return ValidationError{Message: "must be a string", Value: value}
	}

	if strings.TrimSpace(str) == "" {
		if v.Required {
			return ValidationError{Message: "phone required", Value: value}
		}
		return nil
	}

	// Use custom pattern or default pattern
	pattern := v.Pattern
	if pattern == "" {
		// Basic international pattern
		pattern = `^\+?[1-9]\d{1,14}$`
	}

	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		return ValidationError{Message: "invalid phone pattern", Value: value}
	}
	if !matched {
		return ValidationError{Message: "invalid phone format", Value: value}
	}

	return nil
}

// RoleValidator validates user roles
type RoleValidator struct {
	AllowedRoles []string
	Required     bool
}

// Validate validates a user role
func (v RoleValidator) Validate(value any) error {
	if value == nil {
		if v.Required {
			return ValidationError{Message: "role required"}
		}
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return ValidationError{Message: "must be a string", Value: value}
	}

	if isBlank(str) {
		if v.Required {
			return ValidationError{Message: "role required", Value: value}
		}
		return nil
	}

	if len(v.AllowedRoles) == 0 {
		return ValidationError{Message: "allowed roles list is empty", Value: value}
	}

	// Check if the role is in the allowed roles list
	for _, allowedRole := range v.AllowedRoles {
		if str == allowedRole {
			return nil
		}
	}

	return ValidationError{Message: fmt.Sprintf("role must be one of: %s", strings.Join(v.AllowedRoles, ", ")), Value: value}
}

// NameValidator validates person names
type NameValidator struct {
	MinLength int
	MaxLength int
	Required  bool
}

// Validate validates a person name
func (v NameValidator) Validate(value any) error {
	if value == nil {
		if v.Required {
			return ValidationError{Message: "name required"}
		}
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return ValidationError{Message: "must be a string", Value: value}
	}

	if isBlank(str) {
		if v.Required {
			return ValidationError{Message: "name required", Value: value}
		}
		return nil
	}

	runes := utf8.RuneCountInString(str)
	if v.MinLength > 0 && runes < v.MinLength {
		return ValidationError{Message: fmt.Sprintf("minimum length: %d characters", v.MinLength), Value: value}
	}

	if v.MaxLength > 0 && runes > v.MaxLength {
		return ValidationError{Message: fmt.Sprintf("maximum length: %d characters", v.MaxLength), Value: value}
	}

	if !personNameCharsRe.MatchString(str) {
		return ValidationError{Message: "only letters, spaces and some special characters are allowed", Value: value}
	}

	return nil
}

// Helpers for creating authentication validators

// StandardPassword creates a standard password validator
func StandardPassword() PasswordValidator {
	return PasswordValidator{
		Required:       true,
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumbers: true,
		RequireSymbols: true,
	}
}

// SimplePassword creates a simple password validator
func SimplePassword() PasswordValidator {
	return PasswordValidator{
		Required:       true,
		MinLength:      6,
		RequireUpper:   false,
		RequireLower:   true,
		RequireNumbers: false,
		RequireSymbols: false,
	}
}

// MinLengthPassword valida solo longitud mínima (sin reglas de complejidad).
func MinLengthPassword(minLen int) PasswordValidator {
	return PasswordValidator{
		Required:  true,
		MinLength: minLen,
	}
}

// InternationalPhone creates an international phone validator
func InternationalPhone() PhoneValidator {
	return PhoneValidator{
		Required: true,
		Pattern:  `^\+?[1-9]\d{1,14}$`,
	}
}

// StandardPhone creates a standard phone validator
func StandardPhone() PhoneValidator {
	return PhoneValidator{
		Required: false,
		Pattern:  `^\+?[1-9]\d{1,14}$`,
	}
}

// SystemRoles creates a validator for system roles
func SystemRoles() RoleValidator {
	return RoleValidator{
		Required: true,
		// Must reflect the roles seeded in migrations (000006_insert_initial_data.up.sql)
		AllowedRoles: []string{
			"super_admin",
			"correspondence_admin",
			"correspondence_operator",
			"dependency_manager",
			"funcionario",
			"user",
			"visualizador",
		},
	}
}

// UserRoles creates a validator for user roles (without super_admin)
func UserRoles() RoleValidator {
	return RoleValidator{
		Required: true,
		// Variant without super_admin, but aligned to existing roles
		AllowedRoles: []string{
			"correspondence_admin",
			"correspondence_operator",
			"dependency_manager",
			"funcionario",
			"user",
			"visualizador",
		},
	}
}

// PersonName creates a validator for person names
func PersonName() NameValidator {
	return NameValidator{
		Required:  true,
		MinLength: 2,
		MaxLength: 100,
	}
}

// OptionalPersonName creates an optional validator for person names
func OptionalPersonName() NameValidator {
	return NameValidator{
		Required:  false,
		MinLength: 2,
		MaxLength: 100,
	}
}
