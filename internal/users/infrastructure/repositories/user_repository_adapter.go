package repositories

import (
	"context"

	authPorts "github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	"github.com/yovannylopez/docsy-main/internal/users/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/users/domain/ports"
	apperrors "github.com/yovannylopez/docsy-main/pkg/errors"
)

// UserProfileRepositoryAdapter adapts auth.UserRepository to the UserProfileRepository port of the users BC.
// Exposes only profile/listing operations; security remains in auth.
type UserProfileRepositoryAdapter struct {
	userRepo authPorts.UserRepository
}

// NewUserProfileRepositoryAdapter creates the adapter that implements ports.UserProfileRepository.
func NewUserProfileRepositoryAdapter(userRepo authPorts.UserRepository) ports.UserProfileRepository {
	return &UserProfileRepositoryAdapter{
		userRepo: userRepo,
	}
}

// Create creates a new user
func (a *UserProfileRepositoryAdapter) Create(ctx context.Context, user *entities.User) error {
	if err := a.userRepo.Create(ctx, user); err != nil {
		return apperrors.Wrap(err, "user_profile_adapter: create")
	}
	return nil
}

// FindByEmail finds a user by their email
func (a *UserProfileRepositoryAdapter) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	user, err := a.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, apperrors.Wrap(err, "user_profile_adapter: find by email")
	}
	return user, nil
}

// FindByID finds a user by their ID
func (a *UserProfileRepositoryAdapter) FindByID(ctx context.Context, userID string) (*entities.User, error) {
	user, err := a.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperrors.Wrap(err, "user_profile_adapter: find by id")
	}
	return user, nil
}

// FindByUsername finds a user by their username
func (a *UserProfileRepositoryAdapter) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	user, err := a.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, apperrors.Wrap(err, "user_profile_adapter: find by username")
	}
	return user, nil
}

// Update updates a user
func (a *UserProfileRepositoryAdapter) Update(ctx context.Context, user *entities.User) error {
	if err := a.userRepo.Update(ctx, user); err != nil {
		return apperrors.Wrap(err, "user_profile_adapter: update")
	}
	return nil
}

// GetAllUsers retrieves all users with pagination
func (a *UserProfileRepositoryAdapter) GetAllUsers(ctx context.Context, limit, offset int) ([]entities.User, error) {
	users, err := a.userRepo.GetAllUsers(ctx, limit, offset)
	if err != nil {
		return nil, apperrors.Wrap(err, "user_profile_adapter: get all users")
	}
	return users, nil
}

// GetTotalUsersCount retrieves the total number of users in the database
func (a *UserProfileRepositoryAdapter) GetTotalUsersCount(ctx context.Context) (int, error) {
	total, err := a.userRepo.GetTotalUsersCount(ctx)
	if err != nil {
		return 0, apperrors.Wrap(err, "user_profile_adapter: total users count")
	}
	return total, nil
}

// SearchUsers searches users by text across multiple fields with an optional active status filter
func (a *UserProfileRepositoryAdapter) SearchUsers(ctx context.Context, query string, activo *bool, limit, offset int) ([]entities.User, error) {
	users, err := a.userRepo.SearchUsers(ctx, query, activo, limit, offset)
	if err != nil {
		return nil, apperrors.Wrap(err, "user_profile_adapter: search users")
	}
	return users, nil
}

// CountSearchUsers counts the total users matching the search query and filter
func (a *UserProfileRepositoryAdapter) CountSearchUsers(ctx context.Context, query string, activo *bool) (int, error) {
	total, err := a.userRepo.CountSearchUsers(ctx, query, activo)
	if err != nil {
		return 0, apperrors.Wrap(err, "user_profile_adapter: count search users")
	}
	return total, nil
}

// GetRoleByName retrieves a role by name
func (a *UserProfileRepositoryAdapter) GetRoleByName(ctx context.Context, roleName string) (*entities.Role, error) {
	role, err := a.userRepo.GetRoleByName(ctx, roleName)
	if err != nil {
		return nil, apperrors.Wrap(err, "user_profile_adapter: get role by name")
	}
	return role, nil
}
