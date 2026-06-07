package ports

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

// SystemConfigRepository defines the operations for the system configuration repository
type SystemConfigRepository interface {
	GetConfig(ctx context.Context, key string) (*entities.SystemConfig, error)
	SetConfig(ctx context.Context, config *entities.SystemConfig) error
	GetAllConfig(ctx context.Context) ([]entities.SystemConfig, error)
	DeleteConfig(ctx context.Context, key string) error
}
