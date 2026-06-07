package validators

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		error    ValidationError
		expected string
	}{
		{
			name: "basic error",
			error: ValidationError{
				Field:   "email",
				Message: "invalid format",
				Value:   "invalid-email",
			},
			expected: "validation failed for field 'email': invalid format (value: invalid-email)",
		},
		{
			name: "error without field",
			error: ValidationError{
				Message: "required field",
				Value:   nil,
			},
			expected: "validation failed for field '': required field (value: <nil>)",
		},
		{
			name: "error with complex value",
			error: ValidationError{
				Field:   "age",
				Message: "must be greater than 18",
				Value:   15,
			},
			expected: "validation failed for field 'age': must be greater than 18 (value: 15)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.error.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidationError_ClientMessage(t *testing.T) {
	e := ValidationError{Message: "required field", Value: "x"}
	assert.Equal(t, "required field", e.ClientMessage())
}

func TestErrorClientMessage(t *testing.T) {
	ve := ValidationError{Message: "minimum length: 2 characters"}
	assert.Equal(t, "minimum length: 2 characters", ErrorClientMessage(ve))
	assert.Equal(t, "minimum length: 2 characters", ErrorClientMessage(fmt.Errorf("wrap: %w", ve)))

	plain := errors.New("otro")
	assert.Equal(t, "otro", ErrorClientMessage(plain))
}

func TestStringValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		validator StringValidator
		value     any
		expectErr bool
		errMsg    string
	}{
		// Cases with nil value
		{
			name:      "required nil value",
			validator: StringValidator{Required: true},
			value:     nil,
			expectErr: true,
			errMsg:    "field required",
		},
		{
			name:      "non-required nil value",
			validator: StringValidator{Required: false},
			value:     nil,
			expectErr: false,
		},
		// Cases with incorrect type
		{
			name:      "incorrect type",
			validator: StringValidator{},
			value:     123,
			expectErr: true,
			errMsg:    "must be a string",
		},
		// Cases with empty string
		{
			name:      "required empty string",
			validator: StringValidator{Required: true},
			value:     "",
			expectErr: true,
			errMsg:    "field required",
		},
		{
			name:      "required string with spaces",
			validator: StringValidator{Required: true},
			value:     "   ",
			expectErr: true,
			errMsg:    "field required",
		},
		{
			name:      "non-required empty string",
			validator: StringValidator{Required: false},
			value:     "",
			expectErr: false,
		},
		// Cases with minimum length
		{
			name:      "length below minimum",
			validator: StringValidator{MinLength: 5},
			value:     "abc",
			expectErr: true,
			errMsg:    "minimum length: 5 characters",
		},
		{
			name:      "length equal to minimum",
			validator: StringValidator{MinLength: 3},
			value:     "abc",
			expectErr: false,
		},
		{
			name:      "length above minimum",
			validator: StringValidator{MinLength: 2},
			value:     "abc",
			expectErr: false,
		},
		// Cases with maximum length
		{
			name:      "length above maximum",
			validator: StringValidator{MaxLength: 5},
			value:     "abcdef",
			expectErr: true,
			errMsg:    "maximum length: 5 characters",
		},
		{
			name:      "length equal to maximum",
			validator: StringValidator{MaxLength: 3},
			value:     "abc",
			expectErr: false,
		},
		{
			name:      "length below maximum",
			validator: StringValidator{MaxLength: 5},
			value:     "abc",
			expectErr: false,
		},
		// Cases with pattern
		{
			name:      "valid pattern",
			validator: StringValidator{Pattern: `^[a-z]+$`},
			value:     "abc",
			expectErr: false,
		},
		{
			name:      "invalid pattern",
			validator: StringValidator{Pattern: `^[a-z]+$`},
			value:     "abc123",
			expectErr: true,
			errMsg:    "does not match the required pattern",
		},
		{
			name:      "invalid regex pattern",
			validator: StringValidator{Pattern: `[invalid`},
			value:     "test",
			expectErr: true,
			errMsg:    "invalid validation pattern",
		},
		// Valid cases
		{
			name:      "valid string",
			validator: StringValidator{},
			value:     "test",
			expectErr: false,
		},
		{
			name: "valid string with complete configuration",
			validator: StringValidator{
				MinLength: 2,
				MaxLength: 10,
				Pattern:   `^[a-zA-Z]+$`,
				Required:  true,
			},
			value:     "Hello",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator.Validate(tt.value)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNumberValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		validator NumberValidator
		value     any
		expectErr bool
		errMsg    string
	}{
		// Cases with nil value
		{
			name:      "required nil value",
			validator: NumberValidator{Required: true},
			value:     nil,
			expectErr: true,
			errMsg:    "field required",
		},
		{
			name:      "non-required nil value",
			validator: NumberValidator{Required: false},
			value:     nil,
			expectErr: false,
		},
		// Cases with incorrect type
		{
			name:      "incorrect type",
			validator: NumberValidator{},
			value:     "not a number",
			expectErr: true,
			errMsg:    "must be a number",
		},
		// Cases with positive numbers
		{
			name:      "negative number when it should be positive",
			validator: NumberValidator{Positive: true},
			value:     -5,
			expectErr: true,
			errMsg:    "must be a positive number",
		},
		{
			name:      "zero when it should be positive",
			validator: NumberValidator{Positive: true},
			value:     0,
			expectErr: true,
			errMsg:    "must be a positive number",
		},
		{
			name:      "valid positive number",
			validator: NumberValidator{Positive: true},
			value:     5,
			expectErr: false,
		},
		// Cases with minimum value
		{
			name:      "value below minimum",
			validator: NumberValidator{Min: float64p(10)},
			value:     5,
			expectErr: true,
			errMsg:    "minimum value: 10",
		},
		{
			name:      "value equal to minimum",
			validator: NumberValidator{Min: float64p(10)},
			value:     10,
			expectErr: false,
		},
		{
			name:      "value above minimum",
			validator: NumberValidator{Min: float64p(10)},
			value:     15,
			expectErr: false,
		},
		// Cases with maximum value
		{
			name:      "value above maximum",
			validator: NumberValidator{Max: float64p(100)},
			value:     150,
			expectErr: true,
			errMsg:    "maximum value: 100",
		},
		{
			name:      "value equal to maximum",
			validator: NumberValidator{Max: float64p(100)},
			value:     100,
			expectErr: false,
		},
		{
			name:      "value below maximum",
			validator: NumberValidator{Max: float64p(100)},
			value:     50,
			expectErr: false,
		},
		// Cases with different numeric types
		{
			name:      "valid int",
			validator: NumberValidator{},
			value:     42,
			expectErr: false,
		},
		{
			name:      "valid int32",
			validator: NumberValidator{},
			value:     int32(42),
			expectErr: false,
		},
		{
			name:      "valid int64",
			validator: NumberValidator{},
			value:     int64(42),
			expectErr: false,
		},
		{
			name:      "valid float32",
			validator: NumberValidator{},
			value:     float32(42.5),
			expectErr: false,
		},
		{
			name:      "valid float64",
			validator: NumberValidator{},
			value:     42.5,
			expectErr: false,
		},
		// Valid cases
		{
			name:      "valid number without restrictions",
			validator: NumberValidator{},
			value:     42,
			expectErr: false,
		},
		{
			name: "valid number with complete configuration",
			validator: NumberValidator{
				Min:      float64p(10),
				Max:      float64p(100),
				Positive: true,
				Required: true,
			},
			value:     50,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator.Validate(tt.value)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequiredNumberRange_IncludesZero(t *testing.T) {
	v := RequiredNumberRange(0, 100)
	assert.NoError(t, v.Validate(0))
	assert.NoError(t, v.Validate(100))
	assert.Error(t, v.Validate(-1))
	assert.Error(t, v.Validate(101))
}

func TestStringValidator_UnicodeRuneLength(t *testing.T) {
	v := StringValidator{Required: true, MaxLength: 2, MinLength: 1}
	assert.NoError(t, v.Validate("漢"))
	assert.NoError(t, v.Validate("áé"))
	assert.NoError(t, v.Validate("漢字"))
	assert.Error(t, v.Validate("漢字字"))
}

func TestEmailValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		validator EmailValidator
		value     any
		expectErr bool
		errMsg    string
	}{
		// Cases with nil value
		{
			name:      "required nil value",
			validator: EmailValidator{Required: true},
			value:     nil,
			expectErr: true,
			errMsg:    "field required",
		},
		{
			name:      "non-required nil value",
			validator: EmailValidator{Required: false},
			value:     nil,
			expectErr: false,
		},
		// Cases with incorrect type
		{
			name:      "incorrect type",
			validator: EmailValidator{},
			value:     123,
			expectErr: true,
			errMsg:    "must be a string",
		},
		// Cases with empty string
		{
			name:      "required empty email",
			validator: EmailValidator{Required: true},
			value:     "",
			expectErr: true,
			errMsg:    "field required",
		},
		{
			name:      "required email with spaces",
			validator: EmailValidator{Required: true},
			value:     "   ",
			expectErr: true,
			errMsg:    "field required",
		},
		{
			name:      "non-required empty email",
			validator: EmailValidator{Required: false},
			value:     "",
			expectErr: false,
		},
		// Cases with valid emails
		{
			name:      "simple valid email",
			validator: EmailValidator{},
			value:     "test@example.com",
			expectErr: false,
		},
		{
			name:      "email with subdomain",
			validator: EmailValidator{},
			value:     "user@sub.example.com",
			expectErr: false,
		},
		{
			name:      "email with special characters",
			validator: EmailValidator{},
			value:     "user+tag@example.com",
			expectErr: false,
		},
		{
			name:      "email with dots",
			validator: EmailValidator{},
			value:     "user.name@example.com",
			expectErr: false,
		},
		{
			name:      "email with hyphens",
			validator: EmailValidator{},
			value:     "user-name@example-domain.com",
			expectErr: false,
		},
		// Cases with invalid emails
		{
			name:      "email without @",
			validator: EmailValidator{},
			value:     "testexample.com",
			expectErr: true,
			errMsg:    "invalid email format",
		},
		{
			name:      "email without domain",
			validator: EmailValidator{},
			value:     "test@",
			expectErr: true,
			errMsg:    "invalid email format",
		},
		{
			name:      "email without user",
			validator: EmailValidator{},
			value:     "@example.com",
			expectErr: true,
			errMsg:    "invalid email format",
		},
		{
			name:      "email with very short domain",
			validator: EmailValidator{},
			value:     "test@example.c",
			expectErr: true,
			errMsg:    "invalid email format",
		},
		{
			name:      "email with spaces",
			validator: EmailValidator{},
			value:     "test @example.com",
			expectErr: true,
			errMsg:    "invalid email format",
		},
		{
			name:      "email with invalid characters",
			validator: EmailValidator{},
			value:     "test<@example.com",
			expectErr: true,
			errMsg:    "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator.Validate(tt.value)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCompositeValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		validator CompositeValidator
		value     any
		expectErr bool
		errMsg    string
	}{
		{
			name: "successful validation with multiple validators",
			validator: CompositeValidator{
				Validators: []Validator{
					StringValidator{MinLength: 3, MaxLength: 10},
					StringValidator{Pattern: `^[a-zA-Z]+$`},
				},
			},
			value:     "Hello",
			expectErr: false,
		},
		{
			name: "first validator fails",
			validator: CompositeValidator{
				Validators: []Validator{
					StringValidator{MinLength: 10}, // Will fail
					StringValidator{Pattern: `^[a-zA-Z]+$`},
				},
			},
			value:     "Hi",
			expectErr: true,
			errMsg:    "minimum length: 10 characters",
		},
		{
			name: "second validator fails",
			validator: CompositeValidator{
				Validators: []Validator{
					StringValidator{MinLength: 3},
					StringValidator{Pattern: `^[a-zA-Z]+$`}, // Will fail
				},
			},
			value:     "Hi123",
			expectErr: true,
			errMsg:    "does not match the required pattern",
		},
		{
			name: "no validators",
			validator: CompositeValidator{
				Validators: []Validator{},
			},
			value:     "any value",
			expectErr: false,
		},
		{
			name: "combination of different validator types",
			validator: CompositeValidator{
				Validators: []Validator{
					StringValidator{MinLength: 5},
					EmailValidator{Required: true},
				},
			},
			value:     "test@example.com",
			expectErr: false,
		},
		{
			name: "combination with email failure",
			validator: CompositeValidator{
				Validators: []Validator{
					StringValidator{MinLength: 5},
					EmailValidator{Required: true},
				},
			},
			value:     "invalid-email",
			expectErr: true,
			errMsg:    "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator.Validate(tt.value)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidators_EdgeCases(t *testing.T) {
	t.Run("StringValidator with extreme configuration", func(t *testing.T) {
		validator := StringValidator{
			MinLength: 1,
			MaxLength: 1000,
			Pattern:   `^[a-zA-Z0-9\s]+$`,
			Required:  true,
		}

		// Very long string
		longString := string(make([]byte, 1001))
		err := validator.Validate(longString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "maximum length: 1000 characters")

		// String with special characters
		err = validator.Validate("test@123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not match the required pattern")
	})

	t.Run("NumberValidator with extreme values", func(t *testing.T) {
		validator := NumberValidator{
			Min:      float64p(-1000),
			Max:      float64p(1000),
			Positive: false,
			Required: true,
		}

		// Values at the boundaries
		err := validator.Validate(-1000)
		assert.NoError(t, err)

		err = validator.Validate(1000)
		assert.NoError(t, err)

		// Values outside the boundaries
		err = validator.Validate(-1001)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "minimum value: -1000")

		err = validator.Validate(1001)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "maximum value: 1000")
	})

	t.Run("EmailValidator with edge cases", func(t *testing.T) {
		validator := EmailValidator{Required: true}

		// Email with many special characters
		err := validator.Validate("user+tag-name.test@sub-domain.example.co.uk")
		assert.NoError(t, err)

		// Email with very long domain
		longDomain := "test@" + string(make([]byte, 100)) + ".com"
		err = validator.Validate(longDomain)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
	})
}

func TestValidators_Performance(t *testing.T) {
	t.Run("StringValidator performance", func(t *testing.T) {
		validator := StringValidator{
			MinLength: 10,
			MaxLength: 100,
			Pattern:   `^[a-zA-Z0-9]+$`,
			Required:  true,
		}

		validString := "ValidString123"
		for i := 0; i < 1000; i++ {
			err := validator.Validate(validString)
			assert.NoError(t, err)
		}
	})

	t.Run("NumberValidator performance", func(t *testing.T) {
		validator := NumberValidator{
			Min:      float64p(1),
			Max:      float64p(1000),
			Positive: true,
			Required: true,
		}

		for i := 1; i <= 1000; i++ {
			err := validator.Validate(i)
			assert.NoError(t, err)
		}
	})

	t.Run("CompositeValidator performance", func(t *testing.T) {
		validator := CompositeValidator{
			Validators: []Validator{
				StringValidator{MinLength: 5, MaxLength: 50},
				EmailValidator{Required: true},
				NumberValidator{Min: float64p(1), Max: float64p(100)},
			},
		}

		// This test will fail because we are passing a string to a NumberValidator
		// but it is to demonstrate the performance of the composite
		err := validator.Validate("test@example.com")
		assert.Error(t, err) // Will fail on the NumberValidator
	})
}

// Benchmarks

func BenchmarkStringValidator_Validate(b *testing.B) {
	validator := StringValidator{
		MinLength: 5,
		MaxLength: 50,
		Pattern:   `^[a-zA-Z0-9]+$`,
		Required:  true,
	}

	validValue := "ValidString123"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validValue)
	}
}

func BenchmarkNumberValidator_Validate(b *testing.B) {
	validator := NumberValidator{
		Min:      float64p(1),
		Max:      float64p(1000),
		Positive: true,
		Required: true,
	}

	validValue := 500
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validValue)
	}
}

func BenchmarkEmailValidator_Validate(b *testing.B) {
	validator := EmailValidator{Required: true}

	validEmail := "test@example.com"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validEmail)
	}
}

func BenchmarkCompositeValidator_Validate(b *testing.B) {
	validator := CompositeValidator{
		Validators: []Validator{
			StringValidator{MinLength: 5, MaxLength: 50},
			EmailValidator{Required: true},
		},
	}

	validValue := "test@example.com"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validValue)
	}
}

func BenchmarkValidationError_Error(b *testing.B) {
	validationError := ValidationError{
		Field:   "email",
		Message: "invalid format",
		Value:   "invalid-email",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validationError.Error()
	}
}
