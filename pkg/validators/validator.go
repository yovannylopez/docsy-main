// Package validators offers reusable validators (string, number, email, composite)
// and domain validators for authentication (password, phone, role, name).
// Messages are oriented towards API responses in English.
package validators

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Value   any
}

// Error returns the error message
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// ClientMessage returns only the client-oriented message (without field or raw value).
func (e ValidationError) ClientMessage() string {
	return e.Message
}

// ErrorClientMessage returns ClientMessage if err is or wraps ValidationError; otherwise err.Error().
func ErrorClientMessage(err error) string {
	var ve ValidationError
	if errors.As(err, &ve) {
		return ve.ClientMessage()
	}
	return err.Error()
}

// Validator defines the interface for validators
type Validator interface {
	Validate(value any) error
}

// StringValidator validates strings
type StringValidator struct {
	MinLength int
	MaxLength int
	Pattern   string
	Required  bool
}

// Validate validates a string
func (v StringValidator) Validate(value any) error {
	if value == nil {
		if v.Required {
			return ValidationError{Message: "field required"}
		}
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return ValidationError{Message: "must be a string", Value: value}
	}

	if isBlank(str) {
		if v.Required {
			return ValidationError{Message: "field required", Value: value}
		}
		return nil // if not required, empty string is valid
	}

	runes := utf8.RuneCountInString(str)
	if v.MinLength > 0 && runes < v.MinLength {
		return ValidationError{Message: fmt.Sprintf("minimum length: %d characters", v.MinLength), Value: value}
	}

	if v.MaxLength > 0 && runes > v.MaxLength {
		return ValidationError{Message: fmt.Sprintf("maximum length: %d characters", v.MaxLength), Value: value}
	}

	if v.Pattern != "" {
		re, err := cachedRegexp(v.Pattern)
		if err != nil {
			return ValidationError{Message: "invalid validation pattern", Value: value}
		}
		if !re.MatchString(str) {
			return ValidationError{Message: "does not match the required pattern", Value: value}
		}
	}

	return nil
}

// NumberValidator validates numbers. Min and Max are optional: nil = no limit at that end
// (allows ranges that include 0, e.g. [0, 100] via RequiredNumberRange(0, 100)).
type NumberValidator struct {
	Min      *float64
	Max      *float64
	Positive bool
	Required bool
}

// Validate validates a number
func (v NumberValidator) Validate(value any) error {
	if value == nil {
		if v.Required {
			return ValidationError{Message: "field required"}
		}
		return nil
	}

	var num float64
	switch n := value.(type) {
	case int:
		num = float64(n)
	case int32:
		num = float64(n)
	case int64:
		num = float64(n)
	case float32:
		num = float64(n)
	case float64:
		num = n
	default:
		return ValidationError{Message: "must be a number", Value: value}
	}

	if v.Positive && num <= 0 {
		return ValidationError{Message: "must be a positive number", Value: value}
	}

	if v.Min != nil && num < *v.Min {
		return ValidationError{Message: fmt.Sprintf("minimum value: %v", *v.Min), Value: value}
	}

	if v.Max != nil && num > *v.Max {
		return ValidationError{Message: fmt.Sprintf("maximum value: %v", *v.Max), Value: value}
	}

	return nil
}

// EmailValidator validates emails
type EmailValidator struct {
	Required bool
}

// Validate validates an email
func (v EmailValidator) Validate(value any) error {
	if value == nil {
		if v.Required {
			return ValidationError{Message: "field required"}
		}
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return ValidationError{Message: "must be a string", Value: value}
	}

	if isBlank(str) {
		if v.Required {
			return ValidationError{Message: "field required", Value: value}
		}
		return nil // if not required, empty string is valid
	}

	if !emailFormatRegex.MatchString(str) {
		return ValidationError{Message: "invalid email format", Value: value}
	}

	return nil
}

// CompositeValidator allows combining multiple validators
type CompositeValidator struct {
	Validators []Validator
}

// Validate validates a value combining multiple validators
func (v CompositeValidator) Validate(value any) error {
	for _, validator := range v.Validators {
		if err := validator.Validate(value); err != nil {
			return err
		}
	}
	return nil
}
