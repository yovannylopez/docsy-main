package ports

import (
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

// TokenGenerator defines the operations for token generation
type TokenGenerator interface {
	// GenerateToken issues access and refresh tokens. If sessionID is not empty, it is included in the claims (login with session in DB).
	GenerateToken(user *entities.User, sessionID string) (*entities.AuthToken, error)
	ValidateToken(tokenString string) (map[string]any, error)
	// ParseRefreshToken validates the signature and claims of a refresh JWT (without touching the DB).
	ParseRefreshToken(refreshToken string) (userID, sessionID string, err error)
}
