package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

func validateActiveSession(session *entities.Session, userID string) error {
	if session == nil {
		return errors.New("invalid session")
	}
	if userID != "" && session.UserID != userID {
		return errors.New("invalid session")
	}
	if !session.IsActive || session.RevokedAt != nil {
		return errors.New("session revoked")
	}
	if time.Now().After(session.ExpiresAt) {
		return errors.New("session expired")
	}
	return nil
}

func (uc *AuthUseCase) loadAndValidateSession(ctx context.Context, sessionID, userID string) error {
	s, err := uc.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return err
	}
	return validateActiveSession(s, userID)
}
