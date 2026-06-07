package ports

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/users/domain/entities"
)

// UserProfileRepository defines the user profile operations used by the use cases.
// Contains only the methods required for CreateUsersUseCase, GetUsersUseCase,
// UpdateUserUseCase and SearchUsersUseCase, separating security responsibilities.
type UserProfileRepository interface {
	Create(ctx context.Context, user *entities.User) error
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByID(ctx context.Context, userID string) (*entities.User, error)
	FindByUsername(ctx context.Context, username string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	GetAllUsers(ctx context.Context, limit, offset int) ([]entities.User, error)
	GetRoleByName(ctx context.Context, roleName string) (*entities.Role, error)
	GetTotalUsersCount(ctx context.Context) (int, error)
	SearchUsers(ctx context.Context, query string, activo *bool, limit, offset int) ([]entities.User, error)
	CountSearchUsers(ctx context.Context, query string, activo *bool) (int, error)
}
