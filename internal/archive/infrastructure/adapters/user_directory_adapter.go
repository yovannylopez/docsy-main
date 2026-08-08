package adapters

import (
	"context"
	"fmt"
	"strings"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/ports"
	authentities "github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

// AuthUserFinder is the subset of auth UserRepository needed for invitations.
type AuthUserFinder interface {
	FindByEmail(ctx context.Context, email string) (*authentities.User, error)
}

// UserDirectoryAdapter adapts auth user lookup to archive ports.UserDirectory.
type UserDirectoryAdapter struct {
	users AuthUserFinder
}

// NewUserDirectoryAdapter creates the adapter.
func NewUserDirectoryAdapter(users AuthUserFinder) *UserDirectoryAdapter {
	return &UserDirectoryAdapter{users: users}
}

// FindByEmail returns a UserRef or nil when not found.
func (a *UserDirectoryAdapter) FindByEmail(ctx context.Context, email string) (*ports.UserRef, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, nil
	}
	user, err := a.users.FindByEmail(ctx, email)
	if err != nil {
		// auth repo returns error on not found in some paths; treat empty as nil.
		if user == nil {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	if user == nil {
		return nil, nil
	}
	name := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if name == "" {
		name = user.Email
	}
	return &ports.UserRef{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: name,
	}, nil
}
