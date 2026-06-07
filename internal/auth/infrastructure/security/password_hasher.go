package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
)

// PasswordHasher implements ports.PasswordHasher using bcrypt
type PasswordHasher struct{}

// NewPasswordHasher creates a new instance of PasswordHasher
func NewPasswordHasher() ports.PasswordHasher {
	return &PasswordHasher{}
}

// HashPassword generates the hash of a password
func (h *PasswordHasher) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(bytes), nil
}

// VerifyPassword verifies a password against its hash
func (h *PasswordHasher) VerifyPassword(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err == nil, nil
}
