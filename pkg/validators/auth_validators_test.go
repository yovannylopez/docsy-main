package validators

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		validator PasswordValidator
		value     any
		expectErr bool
		errMsg    string
	}{
		{
			name: "complete valid password",
			validator: PasswordValidator{
				Required:       true,
				MinLength:      8,
				RequireUpper:   true,
				RequireLower:   true,
				RequireNumbers: true,
				RequireSymbols: true,
			},
			value:     "ValidPass123!",
			expectErr: false,
		},
		{
			name: "password without uppercase",
			validator: PasswordValidator{
				Required:       true,
				MinLength:      8,
				RequireUpper:   true,
				RequireLower:   true,
				RequireNumbers: true,
				RequireSymbols: false,
			},
			value:     "validpass123",
			expectErr: true,
			errMsg:    "must contain at least one uppercase letter",
		},
		{
			name: "password without lowercase",
			validator: PasswordValidator{
				Required:       true,
				MinLength:      8,
				RequireUpper:   true,
				RequireLower:   true,
				RequireNumbers: true,
				RequireSymbols: false,
			},
			value:     "VALIDPASS123",
			expectErr: true,
			errMsg:    "must contain at least one lowercase letter",
		},
		{
			name: "password without numbers",
			validator: PasswordValidator{
				Required:       true,
				MinLength:      8,
				RequireUpper:   true,
				RequireLower:   true,
				RequireNumbers: true,
				RequireSymbols: false,
			},
			value:     "ValidPass",
			expectErr: true,
			errMsg:    "must contain at least one number",
		},
		{
			name: "password without symbols",
			validator: PasswordValidator{
				Required:       true,
				MinLength:      8,
				RequireUpper:   true,
				RequireLower:   true,
				RequireNumbers: true,
				RequireSymbols: true,
			},
			value:     "ValidPass123",
			expectErr: true,
			errMsg:    "must contain at least one special symbol",
		},
		{
			name: "password too short",
			validator: PasswordValidator{
				Required:       true,
				MinLength:      10,
				RequireUpper:   false,
				RequireLower:   false,
				RequireNumbers: false,
				RequireSymbols: false,
			},
			value:     "short",
			expectErr: true,
			errMsg:    "minimum length: 10 characters",
		},
		{
			name: "required password with nil value",
			validator: PasswordValidator{
				Required: true,
			},
			value:     nil,
			expectErr: true,
			errMsg:    "password required",
		},
		{
			name: "optional password with nil value",
			validator: PasswordValidator{
				Required: false,
			},
			value:     nil,
			expectErr: false,
		},
		{
			name: "required password with empty string",
			validator: PasswordValidator{
				Required: true,
			},
			value:     "",
			expectErr: true,
			errMsg:    "password required",
		},
		{
			name: "optional password with empty string",
			validator: PasswordValidator{
				Required: false,
			},
			value:     "",
			expectErr: false,
		},
		{
			name: "incorrect type",
			validator: PasswordValidator{
				Required: true,
			},
			value:     123,
			expectErr: true,
			errMsg:    "must be a string",
		},
		{
			name: "valid simple password",
			validator: PasswordValidator{
				Required:       true,
				MinLength:      6,
				RequireUpper:   false,
				RequireLower:   true,
				RequireNumbers: false,
				RequireSymbols: false,
			},
			value:     "simple",
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

func TestPhoneValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		validator PhoneValidator
		value     any
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid international phone",
			validator: PhoneValidator{
				Required: true,
				Pattern:  `^\+?[1-9]\d{1,14}$`,
			},
			value:     "+1234567890",
			expectErr: false,
		},
		{
			name: "valid phone without +",
			validator: PhoneValidator{
				Required: true,
				Pattern:  `^\+?[1-9]\d{1,14}$`,
			},
			value:     "1234567890",
			expectErr: false,
		},
		{
			name: "phone with custom pattern",
			validator: PhoneValidator{
				Required: true,
				Pattern:  `^\d{3}-\d{3}-\d{4}$`,
			},
			value:     "123-456-7890",
			expectErr: false,
		},
		{
			name: "phone not matching custom pattern",
			validator: PhoneValidator{
				Required: true,
				Pattern:  `^\d{3}-\d{3}-\d{4}$`,
			},
			value:     "1234567890",
			expectErr: true,
			errMsg:    "invalid phone format",
		},
		{
			name: "phone starting with 0",
			validator: PhoneValidator{
				Required: true,
			},
			value:     "0123456789",
			expectErr: true,
			errMsg:    "invalid phone format",
		},
		{
			name: "phone with letters",
			validator: PhoneValidator{
				Required: true,
			},
			value:     "123abc456",
			expectErr: true,
			errMsg:    "invalid phone format",
		},
		{
			name: "required phone with nil value",
			validator: PhoneValidator{
				Required: true,
			},
			value:     nil,
			expectErr: true,
			errMsg:    "phone required",
		},
		{
			name: "optional phone with nil value",
			validator: PhoneValidator{
				Required: false,
			},
			value:     nil,
			expectErr: false,
		},
		{
			name: "required phone with empty string",
			validator: PhoneValidator{
				Required: true,
			},
			value:     "",
			expectErr: true,
			errMsg:    "phone required",
		},
		{
			name: "optional phone with empty string",
			validator: PhoneValidator{
				Required: false,
			},
			value:     "",
			expectErr: false,
		},
		{
			name: "incorrect type",
			validator: PhoneValidator{
				Required: true,
			},
			value:     123,
			expectErr: true,
			errMsg:    "must be a string",
		},
		{
			name: "invalid regex pattern",
			validator: PhoneValidator{
				Required: true,
				Pattern:  `[invalid`,
			},
			value:     "1234567890",
			expectErr: true,
			errMsg:    "invalid phone pattern",
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

func TestRoleValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		validator RoleValidator
		value     any
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid role",
			validator: RoleValidator{
				Required:     true,
				AllowedRoles: []string{"user", "admin", "viewer"},
			},
			value:     "user",
			expectErr: false,
		},
		{
			name: "invalid role",
			validator: RoleValidator{
				Required:     true,
				AllowedRoles: []string{"user", "admin", "viewer"},
			},
			value:     "invalid_role",
			expectErr: true,
			errMsg:    "role must be one of: user, admin, viewer",
		},
		{
			name: "required role with nil value",
			validator: RoleValidator{
				Required:     true,
				AllowedRoles: []string{"user", "admin"},
			},
			value:     nil,
			expectErr: true,
			errMsg:    "role required",
		},
		{
			name: "optional role with nil value",
			validator: RoleValidator{
				Required:     false,
				AllowedRoles: []string{"user", "admin"},
			},
			value:     nil,
			expectErr: false,
		},
		{
			name: "required role with empty string",
			validator: RoleValidator{
				Required:     true,
				AllowedRoles: []string{"user", "admin"},
			},
			value:     "",
			expectErr: true,
			errMsg:    "role required",
		},
		{
			name: "optional role with empty string",
			validator: RoleValidator{
				Required:     false,
				AllowedRoles: []string{"user", "admin"},
			},
			value:     "",
			expectErr: false,
		},
		{
			name: "incorrect type",
			validator: RoleValidator{
				Required:     true,
				AllowedRoles: []string{"user", "admin"},
			},
			value:     123,
			expectErr: true,
			errMsg:    "must be a string",
		},
		{
			name: "without allowed roles",
			validator: RoleValidator{
				Required:     true,
				AllowedRoles: []string{},
			},
			value:     "any_role",
			expectErr: true,
			errMsg:    "allowed roles list is empty",
		},
		{
			name: "role with spaces",
			validator: RoleValidator{
				Required:     true,
				AllowedRoles: []string{"user", "admin"},
			},
			value:     " user ",
			expectErr: true,
			errMsg:    "role must be one of: user, admin",
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

func TestNameValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		validator NameValidator
		value     any
		expectErr bool
		errMsg    string
	}{
		{
			name: "simple valid name",
			validator: NameValidator{
				Required:  true,
				MinLength: 2,
				MaxLength: 50,
			},
			value:     "John",
			expectErr: false,
		},
		{
			name: "name with last name",
			validator: NameValidator{
				Required:  true,
				MinLength: 2,
				MaxLength: 100,
			},
			value:     "John Doe",
			expectErr: false,
		},
		{
			name: "name with special characters",
			validator: NameValidator{
				Required:  true,
				MinLength: 2,
				MaxLength: 100,
			},
			value:     "José María",
			expectErr: false,
		},
		{
			name: "name with hyphens",
			validator: NameValidator{
				Required:  true,
				MinLength: 2,
				MaxLength: 100,
			},
			value:     "Jean-Pierre",
			expectErr: false,
		},
		{
			name: "name with apostrophe",
			validator: NameValidator{
				Required:  true,
				MinLength: 2,
				MaxLength: 100,
			},
			value:     "O'Connor",
			expectErr: false,
		},
		{
			name: "name with period",
			validator: NameValidator{
				Required:  true,
				MinLength: 2,
				MaxLength: 100,
			},
			value:     "St. John",
			expectErr: false,
		},
		{
			name: "name too short",
			validator: NameValidator{
				Required:  true,
				MinLength: 3,
				MaxLength: 50,
			},
			value:     "Jo",
			expectErr: true,
			errMsg:    "minimum length: 3 characters",
		},
		{
			name: "name too long",
			validator: NameValidator{
				Required:  true,
				MinLength: 2,
				MaxLength: 10,
			},
			value:     "VeryLongName",
			expectErr: true,
			errMsg:    "maximum length: 10 characters",
		},
		{
			name: "name with numbers",
			validator: NameValidator{
				Required:  true,
				MinLength: 2,
				MaxLength: 50,
			},
			value:     "John123",
			expectErr: true,
			errMsg:    "only letters, spaces, and some special characters are allowed",
		},
		{
			name: "name with invalid symbols",
			validator: NameValidator{
				Required:  true,
				MinLength: 2,
				MaxLength: 50,
			},
			value:     "John@Doe",
			expectErr: true,
			errMsg:    "only letters, spaces, and some special characters are allowed",
		},
		{
			name: "required name with nil value",
			validator: NameValidator{
				Required:  true,
				MinLength: 2,
				MaxLength: 50,
			},
			value:     nil,
			expectErr: true,
			errMsg:    "name required",
		},
		{
			name: "optional name with nil value",
			validator: NameValidator{
				Required:  false,
				MinLength: 2,
				MaxLength: 50,
			},
			value:     nil,
			expectErr: false,
		},
		{
			name: "required name with empty string",
			validator: NameValidator{
				Required:  true,
				MinLength: 2,
				MaxLength: 50,
			},
			value:     "",
			expectErr: true,
			errMsg:    "name required",
		},
		{
			name: "optional name with empty string",
			validator: NameValidator{
				Required:  false,
				MinLength: 2,
				MaxLength: 50,
			},
			value:     "",
			expectErr: false,
		},
		{
			name: "incorrect type",
			validator: NameValidator{
				Required:  true,
				MinLength: 2,
				MaxLength: 50,
			},
			value:     123,
			expectErr: true,
			errMsg:    "must be a string",
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

func TestStandardPassword(t *testing.T) {
	validator := StandardPassword()

	// Verify configuration
	assert.True(t, validator.Required)
	assert.Equal(t, 8, validator.MinLength)
	assert.True(t, validator.RequireUpper)
	assert.True(t, validator.RequireLower)
	assert.True(t, validator.RequireNumbers)
	assert.True(t, validator.RequireSymbols)

	// Test validation
	tests := []struct {
		name      string
		value     any
		expectErr bool
	}{
		{
			name:      "valid password",
			value:     "ValidPass123!",
			expectErr: false,
		},
		{
			name:      "password without uppercase",
			value:     "validpass123!",
			expectErr: true,
		},
		{
			name:      "password without lowercase",
			value:     "VALIDPASS123!",
			expectErr: true,
		},
		{
			name:      "password without numbers",
			value:     "ValidPass!",
			expectErr: true,
		},
		{
			name:      "password without symbols",
			value:     "ValidPass123",
			expectErr: true,
		},
		{
			name:      "password too short",
			value:     "Short1!",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSimplePassword(t *testing.T) {
	validator := SimplePassword()

	// Verify configuration
	assert.True(t, validator.Required)
	assert.Equal(t, 6, validator.MinLength)
	assert.False(t, validator.RequireUpper)
	assert.True(t, validator.RequireLower)
	assert.False(t, validator.RequireNumbers)
	assert.False(t, validator.RequireSymbols)

	// Test validation
	tests := []struct {
		name      string
		value     any
		expectErr bool
	}{
		{
			name:      "valid password",
			value:     "simple",
			expectErr: false,
		},
		{
			name:      "password too short",
			value:     "hi",
			expectErr: true,
		},
		{
			name:      "password without lowercase",
			value:     "SIMPLE",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMinLengthPassword(t *testing.T) {
	v := MinLengthPassword(8)
	assert.NoError(t, v.Validate("12345678"))
	assert.Error(t, v.Validate("short"))
	assert.Error(t, v.Validate(""))
}

func TestInternationalPhone(t *testing.T) {
	validator := InternationalPhone()

	// Verify configuration
	assert.True(t, validator.Required)
	assert.Equal(t, `^\+?[1-9]\d{1,14}$`, validator.Pattern)

	// Test validation
	tests := []struct {
		name      string
		value     any
		expectErr bool
	}{
		{
			name:      "valid phone with +",
			value:     "+1234567890",
			expectErr: false,
		},
		{
			name:      "valid phone without +",
			value:     "1234567890",
			expectErr: false,
		},
		{
			name:      "phone starting with 0",
			value:     "0123456789",
			expectErr: true,
		},
		{
			name:      "phone with letters",
			value:     "123abc456",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStandardPhone(t *testing.T) {
	validator := StandardPhone()

	// Verify configuration
	assert.False(t, validator.Required)
	assert.Equal(t, `^\+?[1-9]\d{1,14}$`, validator.Pattern)

	// Test validation
	tests := []struct {
		name      string
		value     any
		expectErr bool
	}{
		{
			name:      "valid phone",
			value:     "+1234567890",
			expectErr: false,
		},
		{
			name:      "nil value (optional)",
			value:     nil,
			expectErr: false,
		},
		{
			name:      "empty string (optional)",
			value:     "",
			expectErr: false,
		},
		{
			name:      "invalid phone",
			value:     "invalid",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSystemRoles(t *testing.T) {
	validator := SystemRoles()

	// Verify configuration (aligned with migrations / SystemRoles in auth_validators.go)
	assert.True(t, validator.Required)
	assert.Equal(t, []string{
		"super_admin",
		"correspondence_admin",
		"correspondence_operator",
		"dependency_manager",
		"funcionario",
		"user",
		"visualizador",
	}, validator.AllowedRoles)

	// Test validation
	tests := []struct {
		name      string
		value     any
		expectErr bool
	}{
		{
			name:      "valid role",
			value:     "user",
			expectErr: false,
		},
		{
			name:      "correspondence_admin role",
			value:     "correspondence_admin",
			expectErr: false,
		},
		{
			name:      "super_admin role",
			value:     "super_admin",
			expectErr: false,
		},
		{
			name:      "invalid role",
			value:     "invalid_role",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserRoles(t *testing.T) {
	validator := UserRoles()

	// Verify configuration (without super_admin; aligned with UserRoles in auth_validators.go)
	assert.True(t, validator.Required)
	assert.Equal(t, []string{
		"correspondence_admin",
		"correspondence_operator",
		"dependency_manager",
		"funcionario",
		"user",
		"visualizador",
	}, validator.AllowedRoles)

	// Test validation
	tests := []struct {
		name      string
		value     any
		expectErr bool
	}{
		{
			name:      "valid role",
			value:     "user",
			expectErr: false,
		},
		{
			name:      "correspondence_operator role",
			value:     "correspondence_operator",
			expectErr: false,
		},
		{
			name:      "visualizador role",
			value:     "visualizador",
			expectErr: false,
		},
		{
			name:      "super_admin role (not allowed)",
			value:     "super_admin",
			expectErr: true,
		},
		{
			name:      "invalid role",
			value:     "invalid_role",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPersonName(t *testing.T) {
	validator := PersonName()

	// Verify configuration
	assert.True(t, validator.Required)
	assert.Equal(t, 2, validator.MinLength)
	assert.Equal(t, 100, validator.MaxLength)

	// Test validation
	tests := []struct {
		name      string
		value     any
		expectErr bool
	}{
		{
			name:      "valid name",
			value:     "John",
			expectErr: false,
		},
		{
			name:      "name with last name",
			value:     "John Doe",
			expectErr: false,
		},
		{
			name:      "name too short",
			value:     "J",
			expectErr: true,
		},
		{
			name:      "name with numbers",
			value:     "John123",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOptionalPersonName(t *testing.T) {
	validator := OptionalPersonName()

	// Verify configuration
	assert.False(t, validator.Required)
	assert.Equal(t, 2, validator.MinLength)
	assert.Equal(t, 100, validator.MaxLength)

	// Test validation
	tests := []struct {
		name      string
		value     any
		expectErr bool
	}{
		{
			name:      "valid name",
			value:     "John",
			expectErr: false,
		},
		{
			name:      "nil value (optional)",
			value:     nil,
			expectErr: false,
		},
		{
			name:      "empty string (optional)",
			value:     "",
			expectErr: false,
		},
		{
			name:      "name too short",
			value:     "J",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthValidators_EdgeCases(t *testing.T) {
	t.Run("PasswordValidator with extreme configuration", func(t *testing.T) {
		validator := PasswordValidator{
			Required:       true,
			MinLength:      20,
			RequireUpper:   true,
			RequireLower:   true,
			RequireNumbers: true,
			RequireSymbols: true,
		}

		// Password that meets all requirements
		validPassword := "VeryLongPassword123!"
		err := validator.Validate(validPassword)
		assert.NoError(t, err)

		// Password that doesn't meet all requirements
		invalidPassword := "verylongpassword123"
		err = validator.Validate(invalidPassword)
		assert.Error(t, err)
		// The error can be about minimum length or uppercase, verify it's one of them
		errMsg := err.Error()
		assert.True(t,
			strings.Contains(errMsg, "must contain at least one uppercase letter") ||
				strings.Contains(errMsg, "minimum length: 20 characters"),
			"Error message should contain either length or uppercase requirement: %s", errMsg)
	})

	t.Run("PhoneValidator with complex patterns", func(t *testing.T) {
		// Pattern for Spanish phones
		spanishPattern := `^\+34[6-9]\d{8}$`
		validator := PhoneValidator{
			Required: true,
			Pattern:  spanishPattern,
		}

		err := validator.Validate("+34612345678")
		assert.NoError(t, err)

		err = validator.Validate("+34123456789") // Starts with 1, not valid
		assert.Error(t, err)
	})

	t.Run("RoleValidator with many roles", func(t *testing.T) {
		roles := []string{"user", "admin", "visualizador", "editor", "moderator", "super_admin", "guest"}
		validator := RoleValidator{
			Required:     true,
			AllowedRoles: roles,
		}

		for _, role := range roles {
			err := validator.Validate(role)
			assert.NoError(t, err)
		}

		err := validator.Validate("invalid_role")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "role must be one of:")
	})

	t.Run("NameValidator with international names", func(t *testing.T) {
		validator := NameValidator{
			Required:  true,
			MinLength: 2,
			MaxLength: 100,
		}

		// Names with special characters
		names := []string{
			"José María",
			"Jean-Pierre",
			"O'Connor",
			"St. John",
			"José-Luis",
		}

		for _, name := range names {
			err := validator.Validate(name)
			assert.NoError(t, err)
		}

		// Invalid names
		invalidNames := []string{
			"John123",
			"Mary@Doe",
			"José#Luis",
		}

		for _, name := range invalidNames {
			err := validator.Validate(name)
			assert.Error(t, err)
		}
	})
}

func TestAuthValidators_Performance(t *testing.T) {
	t.Run("PasswordValidator performance", func(t *testing.T) {
		validator := StandardPassword()
		validPassword := "ValidPass123!"

		for i := 0; i < 1000; i++ {
			err := validator.Validate(validPassword)
			assert.NoError(t, err)
		}
	})

	t.Run("PhoneValidator performance", func(t *testing.T) {
		validator := InternationalPhone()
		validPhone := "+1234567890"

		for i := 0; i < 1000; i++ {
			err := validator.Validate(validPhone)
			assert.NoError(t, err)
		}
	})

	t.Run("RoleValidator performance", func(t *testing.T) {
		validator := SystemRoles()
		validRole := "user"

		for i := 0; i < 1000; i++ {
			err := validator.Validate(validRole)
			assert.NoError(t, err)
		}
	})

	t.Run("NameValidator performance", func(t *testing.T) {
		validator := PersonName()
		validName := "John Doe"

		for i := 0; i < 1000; i++ {
			err := validator.Validate(validName)
			assert.NoError(t, err)
		}
	})
}

// Benchmarks

func BenchmarkPasswordValidator_Validate(b *testing.B) {
	validator := StandardPassword()
	validPassword := "ValidPass123!"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validPassword)
	}
}

func BenchmarkPhoneValidator_Validate(b *testing.B) {
	validator := InternationalPhone()
	validPhone := "+1234567890"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validPhone)
	}
}

func BenchmarkRoleValidator_Validate(b *testing.B) {
	validator := SystemRoles()
	validRole := "user"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validRole)
	}
}

func BenchmarkNameValidator_Validate(b *testing.B) {
	validator := PersonName()
	validName := "John Doe"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(validName)
	}
}

func BenchmarkStandardPassword(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = StandardPassword()
	}
}

func BenchmarkSimplePassword(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = SimplePassword()
	}
}

func BenchmarkInternationalPhone(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = InternationalPhone()
	}
}

func BenchmarkStandardPhone(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = StandardPhone()
	}
}

func BenchmarkSystemRoles(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = SystemRoles()
	}
}

func BenchmarkUserRoles(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = UserRoles()
	}
}

func BenchmarkPersonName(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = PersonName()
	}
}

func BenchmarkOptionalPersonName(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = OptionalPersonName()
	}
}
