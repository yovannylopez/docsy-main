package validators

// float64p returns a pointer to f (optional limits in NumberValidator).
func float64p(f float64) *float64 {
	p := f
	return &p
}

// Helpers for creating common validators

// RequiredString creates a required string validator
func RequiredString(minLength, maxLength int) StringValidator {
	return StringValidator{
		Required:  true,
		MinLength: minLength,
		MaxLength: maxLength,
	}
}

// OptionalString creates an optional string validator
func OptionalString(minLength, maxLength int) StringValidator {
	return StringValidator{
		Required:  false,
		MinLength: minLength,
		MaxLength: maxLength,
	}
}

// RequiredEmail creates a required email validator
func RequiredEmail() EmailValidator {
	return EmailValidator{Required: true}
}

// OptionalEmail creates an optional email validator
func OptionalEmail() EmailValidator {
	return EmailValidator{Required: false}
}

// RequiredPositiveNumber creates a required positive number validator
func RequiredPositiveNumber() NumberValidator {
	return NumberValidator{
		Required: true,
		Positive: true,
	}
}

// OptionalPositiveNumber creates an optional positive number validator
func OptionalPositiveNumber() NumberValidator {
	return NumberValidator{
		Required: false,
		Positive: true,
	}
}

// RequiredNumberRange creates a number validator for a specific range
// (includes min and max; if min is 0, zero is validated correctly).
func RequiredNumberRange(min, max float64) NumberValidator {
	return NumberValidator{
		Required: true,
		Min:      float64p(min),
		Max:      float64p(max),
	}
}

// OptionalNumberRange creates an optional number validator for a specific range
func OptionalNumberRange(min, max float64) NumberValidator {
	return NumberValidator{
		Required: false,
		Min:      float64p(min),
		Max:      float64p(max),
	}
}

// StringWithPattern creates a string validator with a regex pattern
func StringWithPattern(pattern string, required bool, minLength, maxLength int) StringValidator {
	return StringValidator{
		Required:  required,
		MinLength: minLength,
		MaxLength: maxLength,
		Pattern:   pattern,
	}
}

// Combine combines multiple validators
func Combine(validators ...Validator) CompositeValidator {
	return CompositeValidator{Validators: validators}
}

// OptionalMaxLength validates an optional string: empty or blank is valid; if it has text, only maximum length (runes) is applied.
func OptionalMaxLength(max int) StringValidator {
	return StringValidator{
		Required:  false,
		MaxLength: max,
	}
}
