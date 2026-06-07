package ports

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/users/domain/entities"
)

// PasswordHistoryRepository defines the interface for the password history repository
type PasswordHistoryRepository interface {
	Create(ctx context.Context, passwordHistory *entities.PasswordHistory) error
	GetUserPasswordHistory(ctx context.Context, userID string, limit int) ([]entities.PasswordHistory, error)
	CheckPasswordInHistory(ctx context.Context, userID, passwordHash string) (bool, error)
	CleanOldPasswordHistory(ctx context.Context, userID string, keepCount int) error
}

// UserService defines the interface for user services
type UserService interface {
	CreateUser(ctx context.Context, user *entities.User) error
	GetUserByID(ctx context.Context, userID string) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	UpdateUser(ctx context.Context, user *entities.User) error
	DeleteUser(ctx context.Context, userID string) error
	GetUsers(ctx context.Context, limit, offset int) ([]entities.User, error)
	VerifyUserEmail(ctx context.Context, tokenHash string) error
	ResetPassword(ctx context.Context, email string) error
	ChangePassword(ctx context.Context, userID, newPassword string) error
}
