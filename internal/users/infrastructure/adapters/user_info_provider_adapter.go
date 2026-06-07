package adapters

import (
	"context"
	"fmt"
	"strings"

	"github.com/yovannylopez/docsy-main/internal/users/domain/ports"
)

// UserInfoProviderAdapter adapts UserProfileRepository to provide user names.
// Implements the interface required by the dependencies module without direct coupling.
type UserInfoProviderAdapter struct {
	userRepo ports.UserProfileRepository
}

// NewUserInfoProviderAdapter creates a new adapter
func NewUserInfoProviderAdapter(userRepo ports.UserProfileRepository) *UserInfoProviderAdapter {
	return &UserInfoProviderAdapter{userRepo: userRepo}
}

// GetUserName retrieves the full name of the user (FirstName + LastName)
func (a *UserInfoProviderAdapter) GetUserName(ctx context.Context, userID string) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("userID cannot be empty")
	}

	user, err := a.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return "", nil
	}

	parts := []string{}
	if user.FirstName != "" {
		parts = append(parts, user.FirstName)
	}
	if user.LastName != "" {
		parts = append(parts, user.LastName)
	}

	return strings.TrimSpace(strings.Join(parts, " ")), nil
}
