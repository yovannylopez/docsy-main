package ports

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

// SessionRepository define las operaciones del repositorio de sesiones
type SessionRepository interface {
	Create(ctx context.Context, session *entities.Session) error
	FindByID(ctx context.Context, sessionID string) (*entities.Session, error)
	FindByUserID(ctx context.Context, userID string) ([]entities.Session, error)
	FindByRefreshToken(ctx context.Context, refreshTokenHash string) (*entities.Session, error)
	Update(ctx context.Context, session *entities.Session) error
	RevokeSession(ctx context.Context, sessionID, reason string) error
	RevokeAllUserSessions(ctx context.Context, userID, reason string) error
	CleanupExpiredSessions(ctx context.Context) error
	UpdateLastUsed(ctx context.Context, sessionID string) error
}
