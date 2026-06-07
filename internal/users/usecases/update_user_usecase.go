package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yovannylopez/docsy-main/internal/users/domain/dtos"
	domainerrors "github.com/yovannylopez/docsy-main/internal/users/domain/errors"
	"github.com/yovannylopez/docsy-main/internal/users/domain/ports"
)

// UpdateUserUseCase handles the user update logic
type UpdateUserUseCase struct {
	userRepo ports.UserProfileRepository
}

// NewUpdateUserUseCase creates a new instance of UpdateUserUseCase
func NewUpdateUserUseCase(userRepo ports.UserProfileRepository) *UpdateUserUseCase {
	return &UpdateUserUseCase{
		userRepo: userRepo,
	}
}

// Execute runs the user update
func (uc *UpdateUserUseCase) Execute(
	ctx context.Context,
	userID string,
	req *dtos.UpdateUserRequest,
	updatedByUserID string,
) (*dtos.UpdateUserResponse, error) {
	// Get existing user
	existingUser, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}
	if existingUser == nil {
		return nil, domainerrors.ErrUserNotFound
	}

	// Update only the fields that were sent
	if req.Email != nil && *req.Email != "" {
		// Verify the email is not already in use by another user
		userByEmail, err := uc.userRepo.FindByEmail(ctx, *req.Email)
		if err == nil && userByEmail != nil && userByEmail.ID != userID {
			return nil, domainerrors.ErrEmailAlreadyExists
		}
		existingUser.Email = *req.Email
	}

	if req.Username != nil {
		// Verify the username is not already in use by another user
		if *req.Username != "" {
			userByUsername, err := uc.userRepo.FindByUsername(ctx, *req.Username)
			if err == nil && userByUsername != nil && userByUsername.ID != userID {
				return nil, domainerrors.ErrUsernameAlreadyExists
			}
		}
		existingUser.Username = req.Username
	}

	if req.FirstName != nil && *req.FirstName != "" {
		existingUser.FirstName = *req.FirstName
	}

	if req.LastName != nil && *req.LastName != "" {
		existingUser.LastName = *req.LastName
	}

	if req.IdentificationNumber != nil {
		existingUser.IdentificationNumber = req.IdentificationNumber
	}

	if req.IdentificationType != nil && *req.IdentificationType != "" {
		// Normalize to lowercase
		normalized := strings.ToLower(*req.IdentificationType)
		existingUser.IdentificationType = &normalized
	}

	if req.Phone != nil {
		existingUser.Phone = req.Phone
	}

	if req.IsActive != nil {
		existingUser.IsActive = *req.IsActive
	}

	if req.IsVerified != nil {
		existingUser.IsVerified = *req.IsVerified
	}

	if req.MFAEnabled != nil {
		existingUser.MFAEnabled = *req.MFAEnabled
	}

	// Update metadata
	existingUser.UpdatedAt = time.Now()
	if updatedByUserID != "" {
		existingUser.UpdatedBy = &updatedByUserID
	}

	// Save changes
	if err := uc.userRepo.Update(ctx, existingUser); err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	// Return updated user
	return &dtos.UpdateUserResponse{
		ID:                   existingUser.ID,
		Email:                existingUser.Email,
		Username:             existingUser.Username,
		FirstName:            existingUser.FirstName,
		LastName:             existingUser.LastName,
		IdentificationNumber: existingUser.IdentificationNumber,
		IdentificationType:   existingUser.IdentificationType,
		Phone:                existingUser.Phone,
		IsActive:             existingUser.IsActive,
		IsVerified:           existingUser.IsVerified,
		MFAEnabled:           existingUser.MFAEnabled,
		UpdatedAt:            existingUser.UpdatedAt,
		UpdatedBy:            existingUser.UpdatedBy,
	}, nil
}
