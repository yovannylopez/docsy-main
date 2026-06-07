package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequiredString(t *testing.T) {
	tests := []struct {
		name      string
		minLength int
		maxLength int
		value     any
		expectErr bool
	}{
		{
			name:      "valid string",
			minLength: 3,
			maxLength: 10,
			value:     "hello",
			expectErr: false,
		},
		{
			name:      "string too short",
			minLength: 5,
			maxLength: 10,
			value:     "hi",
			expectErr: true,
		},
		{
			name:      "string too long",
			minLength: 3,
			maxLength: 5,
			value:     "hello world",
			expectErr: true,
		},
		{
			name:      "nil value",
			minLength: 3,
			maxLength: 10,
			value:     nil,
			expectErr: true,
		},
		{
			name:      "incorrect type",
			minLength: 3,
			maxLength: 10,
			value:     123,
			expectErr: true,
		},
		{
			name:      "empty string",
			minLength: 3,
			maxLength: 10,
			value:     "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := RequiredString(tt.minLength, tt.maxLength)
			err := validator.Validate(tt.value)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOptionalString(t *testing.T) {
	tests := []struct {
		name      string
		minLength int
		maxLength int
		value     any
		expectErr bool
	}{
		{
			name:      "valid string",
			minLength: 3,
			maxLength: 10,
			value:     "hello",
			expectErr: false,
		},
		{
			name:      "nil value",
			minLength: 3,
			maxLength: 10,
			value:     nil,
			expectErr: false, // Optional allows nil
		},
		{
			name:      "empty string",
			minLength: 3,
			maxLength: 10,
			value:     "",
			expectErr: false, // Optional allows empty string
		},
		{
			name:      "string too short",
			minLength: 5,
			maxLength: 10,
			value:     "hi",
			expectErr: true,
		},
		{
			name:      "string too long",
			minLength: 3,
			maxLength: 5,
			value:     "hello world",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := OptionalString(tt.minLength, tt.maxLength)
			err := validator.Validate(tt.value)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequiredEmail(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		expectErr bool
	}{
		{
			name:      "valid email",
			value:     "test@example.com",
			expectErr: false,
		},
		{
			name:      "invalid email",
			value:     "invalid-email",
			expectErr: true,
		},
		{
			name:      "nil value",
			value:     nil,
			expectErr: true,
		},
		{
			name:      "empty string",
			value:     "",
			expectErr: true,
		},
		{
			name:      "incorrect type",
			value:     123,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := RequiredEmail()
			err := validator.Validate(tt.value)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOptionalEmail(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		expectErr bool
	}{
		{
			name:      "valid email",
			value:     "test@example.com",
			expectErr: false,
		},
		{
			name:      "nil value",
			value:     nil,
			expectErr: false, // Optional allows nil
		},
		{
			name:      "empty string",
			value:     "",
			expectErr: false, // Optional allows empty string
		},
		{
			name:      "invalid email",
			value:     "invalid-email",
			expectErr: true,
		},
		{
			name:      "incorrect type",
			value:     123,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := OptionalEmail()
			err := validator.Validate(tt.value)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequiredPositiveNumber(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		expectErr bool
	}{
		{
			name:      "positive number",
			value:     42,
			expectErr: false,
		},
		{
			name:      "negative number",
			value:     -5,
			expectErr: true,
		},
		{
			name:      "zero",
			value:     0,
			expectErr: true,
		},
		{
			name:      "nil value",
			value:     nil,
			expectErr: true,
		},
		{
			name:      "incorrect type",
			value:     "not a number",
			expectErr: true,
		},
		{
			name:      "positive float",
			value:     3.14,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := RequiredPositiveNumber()
			err := validator.Validate(tt.value)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOptionalPositiveNumber(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		expectErr bool
	}{
		{
			name:      "positive number",
			value:     42,
			expectErr: false,
		},
		{
			name:      "nil value",
			value:     nil,
			expectErr: false, // Optional allows nil
		},
		{
			name:      "negative number",
			value:     -5,
			expectErr: true,
		},
		{
			name:      "zero",
			value:     0,
			expectErr: true,
		},
		{
			name:      "incorrect type",
			value:     "not a number",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := OptionalPositiveNumber()
			err := validator.Validate(tt.value)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequiredNumberRange(t *testing.T) {
	tests := []struct {
		name      string
		min       float64
		max       float64
		value     any
		expectErr bool
	}{
		{
			name:      "value in range",
			min:       10,
			max:       100,
			value:     50,
			expectErr: false,
		},
		{
			name:      "value at minimum boundary",
			min:       10,
			max:       100,
			value:     10,
			expectErr: false,
		},
		{
			name:      "value at maximum boundary",
			min:       10,
			max:       100,
			value:     100,
			expectErr: false,
		},
		{
			name:      "value below minimum",
			min:       10,
			max:       100,
			value:     5,
			expectErr: true,
		},
		{
			name:      "value above maximum",
			min:       10,
			max:       100,
			value:     150,
			expectErr: true,
		},
		{
			name:      "nil value",
			min:       10,
			max:       100,
			value:     nil,
			expectErr: true,
		},
		{
			name:      "incorrect type",
			min:       10,
			max:       100,
			value:     "not a number",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := RequiredNumberRange(tt.min, tt.max)
			err := validator.Validate(tt.value)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOptionalNumberRange(t *testing.T) {
	tests := []struct {
		name      string
		min       float64
		max       float64
		value     any
		expectErr bool
	}{
		{
			name:      "value in range",
			min:       10,
			max:       100,
			value:     50,
			expectErr: false,
		},
		{
			name:      "nil value",
			min:       10,
			max:       100,
			value:     nil,
			expectErr: false, // Optional allows nil
		},
		{
			name:      "value below minimum",
			min:       10,
			max:       100,
			value:     5,
			expectErr: true,
		},
		{
			name:      "value above maximum",
			min:       10,
			max:       100,
			value:     150,
			expectErr: true,
		},
		{
			name:      "incorrect type",
			min:       10,
			max:       100,
			value:     "not a number",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := OptionalNumberRange(tt.min, tt.max)
			err := validator.Validate(tt.value)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStringWithPattern(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		required  bool
		minLength int
		maxLength int
		value     any
		expectErr bool
	}{
		{
			name:      "string matching pattern",
			pattern:   `^[a-z]+$`,
			required:  true,
			minLength: 3,
			maxLength: 10,
			value:     "hello",
			expectErr: false,
		},
		{
			name:      "string not matching pattern",
			pattern:   `^[a-z]+$`,
			required:  true,
			minLength: 3,
			maxLength: 10,
			value:     "Hello123",
			expectErr: true,
		},
		{
			name:      "required string with nil value",
			pattern:   `^[a-z]+$`,
			required:  true,
			minLength: 3,
			maxLength: 10,
			value:     nil,
			expectErr: true,
		},
		{
			name:      "optional string with nil value",
			pattern:   `^[a-z]+$`,
			required:  false,
			minLength: 3,
			maxLength: 10,
			value:     nil,
			expectErr: false, // Optional allows nil
		},
		{
			name:      "string too short",
			pattern:   `^[a-z]+$`,
			required:  true,
			minLength: 5,
			maxLength: 10,
			value:     "hi",
			expectErr: true,
		},
		{
			name:      "string too long",
			pattern:   `^[a-z]+$`,
			required:  true,
			minLength: 3,
			maxLength: 5,
			value:     "helloworld",
			expectErr: true,
		},
		{
			name:      "invalid pattern",
			pattern:   `[invalid`,
			required:  true,
			minLength: 3,
			maxLength: 10,
			value:     "test",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := StringWithPattern(tt.pattern, tt.required, tt.minLength, tt.maxLength)
			err := validator.Validate(tt.value)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCombine(t *testing.T) {
	tests := []struct {
		name       string
		validators []Validator
		value      any
		expectErr  bool
	}{
		{
			name: "multiple successful validators",
			validators: []Validator{
				StringValidator{MinLength: 3, MaxLength: 10},
				StringValidator{Pattern: `^[a-zA-Z]+$`},
			},
			value:     "Hello",
			expectErr: false,
		},
		{
			name: "first validator fails",
			validators: []Validator{
				StringValidator{MinLength: 10}, // Will fail
				StringValidator{Pattern: `^[a-zA-Z]+$`},
			},
			value:     "Hi",
			expectErr: true,
		},
		{
			name: "second validator fails",
			validators: []Validator{
				StringValidator{MinLength: 3},
				StringValidator{Pattern: `^[a-zA-Z]+$`}, // Will fail
			},
			value:     "Hi123",
			expectErr: true,
		},
		{
			name:       "no validators",
			validators: []Validator{},
			value:      "any value",
			expectErr:  false,
		},
		{
			name: "different validator types",
			validators: []Validator{
				StringValidator{MinLength: 5},
				EmailValidator{Required: true},
			},
			value:     "test@example.com",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := Combine(tt.validators...)
			err := validator.Validate(tt.value)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHelpers_EdgeCases(t *testing.T) {
	t.Run("RequiredString with extreme values", func(t *testing.T) {
		validator := RequiredString(1, 1000)

		// Very long string
		longString := string(make([]byte, 1001))
		err := validator.Validate(longString)
		assert.Error(t, err)

		// Single character string
		err = validator.Validate("a")
		assert.NoError(t, err)
	})

	t.Run("RequiredNumberRange with negative values", func(t *testing.T) {
		validator := RequiredNumberRange(-100, 100)

		err := validator.Validate(-50)
		assert.NoError(t, err)

		err = validator.Validate(-150)
		assert.Error(t, err)

		err = validator.Validate(150)
		assert.Error(t, err)
	})

	t.Run("StringWithPattern with complex patterns", func(t *testing.T) {
		// Pattern for validating emails
		emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		validator := StringWithPattern(emailPattern, true, 5, 100)

		err := validator.Validate("test@example.com")
		assert.NoError(t, err)

		err = validator.Validate("invalid-email")
		assert.Error(t, err)
	})

	t.Run("Combine with many validators", func(t *testing.T) {
		validators := []Validator{
			StringValidator{MinLength: 5, MaxLength: 50},
			StringValidator{Pattern: `^[a-zA-Z]+$`},
			EmailValidator{Required: true},
		}

		validator := Combine(validators...)

		// This test will fail because we are passing an email to a StringValidator
		// but it is to demonstrate the behavior of the composite
		err := validator.Validate("test@example.com")
		assert.Error(t, err)
	})
}

func TestHelpers_Performance(t *testing.T) {
	t.Run("RequiredString performance", func(t *testing.T) {
		validator := RequiredString(5, 50)
		validValue := "ValidString"

		for i := 0; i < 1000; i++ {
			err := validator.Validate(validValue)
			assert.NoError(t, err)
		}
	})

	t.Run("RequiredEmail performance", func(t *testing.T) {
		validator := RequiredEmail()
		validEmail := "test@example.com"

		for i := 0; i < 1000; i++ {
			err := validator.Validate(validEmail)
			assert.NoError(t, err)
		}
	})

	t.Run("RequiredPositiveNumber performance", func(t *testing.T) {
		validator := RequiredPositiveNumber()

		for i := 1; i <= 1000; i++ {
			err := validator.Validate(i)
			assert.NoError(t, err)
		}
	})

	t.Run("Combine performance", func(t *testing.T) {
		validators := []Validator{
			StringValidator{MinLength: 5, MaxLength: 50},
			StringValidator{Pattern: `^[a-zA-Z]+$`},
		}

		validator := Combine(validators...)
		validValue := "ValidString"

		for i := 0; i < 1000; i++ {
			err := validator.Validate(validValue)
			assert.NoError(t, err)
		}
	})
}

// Benchmarks

func BenchmarkRequiredString(b *testing.B) {
	validator := RequiredString(5, 50)
	validValue := "ValidString"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validValue)
	}
}

func BenchmarkOptionalString(b *testing.B) {
	validator := OptionalString(5, 50)
	validValue := "ValidString"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validValue)
	}
}

func BenchmarkRequiredEmail(b *testing.B) {
	validator := RequiredEmail()
	validEmail := "test@example.com"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validEmail)
	}
}

func BenchmarkOptionalEmail(b *testing.B) {
	validator := OptionalEmail()
	validEmail := "test@example.com"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validEmail)
	}
}

func BenchmarkRequiredPositiveNumber(b *testing.B) {
	validator := RequiredPositiveNumber()
	validValue := 42

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validValue)
	}
}

func BenchmarkOptionalPositiveNumber(b *testing.B) {
	validator := OptionalPositiveNumber()
	validValue := 42

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validValue)
	}
}

func BenchmarkRequiredNumberRange(b *testing.B) {
	validator := RequiredNumberRange(1, 100)
	validValue := 50

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validValue)
	}
}

func BenchmarkOptionalNumberRange(b *testing.B) {
	validator := OptionalNumberRange(1, 100)
	validValue := 50

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validValue)
	}
}

func TestOptionalMaxLength(t *testing.T) {
	v := OptionalMaxLength(5)
	assert.NoError(t, v.Validate(""))
	assert.NoError(t, v.Validate("   "))
	assert.NoError(t, v.Validate("hi"))
	assert.Error(t, v.Validate("abcdef"))
}

func BenchmarkStringWithPattern(b *testing.B) {
	validator := StringWithPattern(`^[a-zA-Z]+$`, true, 5, 50)
	validValue := "ValidString"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validValue)
	}
}

func BenchmarkCombine(b *testing.B) {
	validators := []Validator{
		StringValidator{MinLength: 5, MaxLength: 50},
		StringValidator{Pattern: `^[a-zA-Z]+$`},
	}

	validator := Combine(validators...)
	validValue := "ValidString"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validValue)
	}
}
