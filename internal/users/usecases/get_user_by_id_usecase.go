package usecases

import (
	"context"
	"fmt"

	"github.com/yovannylopez/docsy-main/internal/users/domain/entities"
	domainerrors "github.com/yovannylopez/docsy-main/internal/users/domain/errors"
	"github.com/yovannylopez/docsy-main/internal/users/domain/ports"
)

// ErrUserNotFound is returned when the user does not exist (alias for handler compatibility)
var ErrUserNotFound = domainerrors.ErrUserNotFound

// GetUserByIDUseCase handles retrieving a user by ID
type GetUserByIDUseCase struct {
	userProfileRepo ports.UserProfileRepository
}

// NewGetUserByIDUseCase creates a new instance of GetUserByIDUseCase
func NewGetUserByIDUseCase(userProfileRepo ports.UserProfileRepository) *GetUserByIDUseCase {
	return &GetUserByIDUseCase{
		userProfileRepo: userProfileRepo,
	}
}

// Execute runs the user retrieval by ID
func (uc *GetUserByIDUseCase) Execute(ctx context.Context, userID string) (*entities.User, error) {
	if userID == "" {
		return nil, domainerrors.ErrUserIDRequired
	}

	user, err := uc.userProfileRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}

	if user == nil {
		return nil, domainerrors.ErrUserNotFound
	}

	return user, nil
}
