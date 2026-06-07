package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	authdomain "github.com/yovannylopez/docsy-main/internal/auth/domain"
	authentities "github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	authports "github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	"github.com/yovannylopez/docsy-main/internal/users/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/users/domain/entities"
	domainerrors "github.com/yovannylopez/docsy-main/internal/users/domain/errors"
	"github.com/yovannylopez/docsy-main/internal/users/domain/ports"
	"github.com/yovannylopez/docsy-main/pkg/constants"
)

// CreateUsersUseCase holds the repositories needed for the user creation use case
type CreateUsersUseCase struct {
	userProfileRepo ports.UserProfileRepository
	passwordHasher  ports.PasswordHasher
	auditRepo       authports.AuditRepository
}

// NewCreateUsersUseCase creates a new instance of CreateUsersUseCase
func NewCreateUsersUseCase(
	userProfileRepo ports.UserProfileRepository,
	passwordHasher ports.PasswordHasher,
	auditRepo authports.AuditRepository,
) *CreateUsersUseCase {
	return &CreateUsersUseCase{
		userProfileRepo: userProfileRepo,
		passwordHasher:  passwordHasher,
		auditRepo:       auditRepo,
	}
}

// Execute runs the user creation use case
func (uc *CreateUsersUseCase) Execute(
	ctx context.Context,
	request *dtos.CreateUsersRequest,
	createdByUserID string,
) (*dtos.CreateUsersResponse, error) {
	if len(request.Users) == 0 {
		return nil, domainerrors.ErrAtLeastOneUserRequired
	}

	if len(request.Users) > constants.MaxUsersBatchSize {
		return nil, fmt.Errorf("%w: maximum %d users per batch", domainerrors.ErrBatchSizeExceeded, constants.MaxUsersBatchSize)
	}

	var createdUsers []dtos.CreateUserResponse
	var creationErrors []dtos.UserCreationError

	for i, userReq := range request.Users {
		if err := uc.validateUserRequest(&userReq); err != nil {
			creationErrors = append(creationErrors, dtos.UserCreationError{
				Index: i,
				Email: userReq.Email,
				Error: err.Error(),
			})
			continue
		}

		userResponse, err := uc.createSingleUser(ctx, &userReq, createdByUserID)
		if err != nil {
			creationErrors = append(creationErrors, dtos.UserCreationError{
				Index: i,
				Email: userReq.Email,
				Error: err.Error(),
			})
			continue
		}

		createdUsers = append(createdUsers, *userResponse)
	}

	return &dtos.CreateUsersResponse{
		CreatedUsers: createdUsers,
		TotalCreated: len(createdUsers),
		Errors:       creationErrors,
	}, nil
}

func (uc *CreateUsersUseCase) validateUserRequest(req *dtos.CreateUserRequest) error {
	if req.Email == "" {
		return domainerrors.ErrEmailRequired
	}

	if req.Password == "" {
		return domainerrors.ErrPasswordRequired
	}

	if len(req.Password) < constants.MinPasswordLength {
		return domainerrors.ErrPasswordTooShort
	}

	if req.FirstName == "" {
		return domainerrors.ErrFirstNameRequired
	}

	if req.LastName == "" {
		return domainerrors.ErrLastNameRequired
	}

	if req.RoleName == "" {
		return domainerrors.ErrRoleNameRequired
	}

	if err := entities.ValidateIdentificationType(req.IdentificationType); err != nil {
		return err
	}

	return nil
}

func (uc *CreateUsersUseCase) createSingleUser(
	ctx context.Context,
	req *dtos.CreateUserRequest,
	createdByUserID string,
) (*dtos.CreateUserResponse, error) {
	existingUser, err := uc.userProfileRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("error checking existing user: %w", err)
	}

	if existingUser != nil {
		return nil, fmt.Errorf("%w: %s", domainerrors.ErrEmailAlreadyExists, req.Email)
	}

	if req.Username != nil && *req.Username != "" {
		existingUserByUsername, uerr := uc.userProfileRepo.FindByUsername(ctx, *req.Username)
		if uerr != nil {
			return nil, fmt.Errorf("error checking username: %w", uerr)
		}

		if existingUserByUsername != nil {
			return nil, fmt.Errorf("%w: %s", domainerrors.ErrUsernameAlreadyExists, *req.Username)
		}
	}

	role, err := uc.userProfileRepo.GetRoleByName(ctx, req.RoleName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving role '%s': %w", req.RoleName, err)
	}

	hashedPassword, err := uc.passwordHasher.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("error generating password hash: %w", err)
	}

	now := time.Now()
	userID := uuid.New().String()

	isActive := false
	isVerified := false
	mfaEnabled := false

	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	if req.IsVerified != nil {
		isVerified = *req.IsVerified
	}
	if req.MFAEnabled != nil {
		mfaEnabled = *req.MFAEnabled
	}

	var normalizedIdentificationType *string
	if req.IdentificationType != nil && *req.IdentificationType != "" {
		normalized := strings.ToLower(*req.IdentificationType)
		normalizedIdentificationType = &normalized
	}

	var createdBy *string
	if createdByUserID != "" {
		createdBy = &createdByUserID
	}

	user := &entities.User{
		ID:                   userID,
		Email:                req.Email,
		Username:             req.Username,
		PasswordHash:         hashedPassword,
		FirstName:            req.FirstName,
		LastName:             req.LastName,
		IdentificationNumber: req.IdentificationNumber,
		IdentificationType:   normalizedIdentificationType,
		Phone:                req.Phone,
		IsActive:             isActive,
		IsVerified:           isVerified,
		FailedLoginAttempts:  0,
		MFAEnabled:           mfaEnabled,
		PasswordChangedAt:    now,
		MustChangePassword:   false,
		CreatedAt:            now,
		UpdatedAt:            now,
		CreatedBy:            createdBy,
		UpdatedBy:            createdBy,
		Roles:                []entities.Role{*role},
	}

	if err := uc.userProfileRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	uc.logUserCreated(ctx, user.ID)

	return &dtos.CreateUserResponse{
		ID:                   user.ID,
		Username:             user.Username,
		Email:                user.Email,
		FirstName:            user.FirstName,
		LastName:             user.LastName,
		IdentificationNumber: user.IdentificationNumber,
		IdentificationType:   user.IdentificationType,
		Phone:                user.Phone,
		IsActive:             user.IsActive,
		IsVerified:           user.IsVerified,
		MFAEnabled:           user.MFAEnabled,
		CreatedAt:            user.CreatedAt,
		CreatedBy:            user.CreatedBy,
		UpdatedBy:            user.UpdatedBy,
	}, nil
}

func (uc *CreateUsersUseCase) logUserCreated(ctx context.Context, userID string) {
	if uc.auditRepo == nil {
		return
	}

	resource := "user"
	message := "User created successfully"
	auditLog := &authentities.AuditLog{
		ID:         uuid.New().String(),
		UserID:     &userID,
		Action:     authdomain.AuditActionUserCreated,
		Resource:   &resource,
		ResourceID: &userID,
		Result:     authdomain.AuditResultSuccess,
		Message:    &message,
		CreatedAt:  time.Now(),
	}
	_ = uc.auditRepo.LogAction(ctx, auditLog)
}
