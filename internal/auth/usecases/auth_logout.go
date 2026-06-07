package usecases

import (
	"context"
	"errors"
	"fmt"

	authdomain "github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/pkg/constants"
)

func (uc *AuthUseCase) resolveLogoutToken(token string) (sessionID, userID string, err error) {
	claims, verr := uc.tokenGenerator.ValidateToken(token)
	if verr == nil {
		if t, ok := claims[constants.JWTClaimType].(string); ok && t == constants.JWTTokenTypeRefresh {
			return uc.tokenGenerator.ParseRefreshToken(token)
		}
		userID, _ = claims[constants.JWTClaimUserID].(string)
		sessionID, _ = claims[constants.JWTClaimSessionID].(string)
		if sessionID != "" {
			return sessionID, userID, nil
		}
	}

	userID, sessionID, err = uc.tokenGenerator.ParseRefreshToken(token)
	if err != nil {
		return "", "", fmt.Errorf("invalid token: %w", err)
	}
	if sessionID == "" {
		return "", "", errors.New("session_id missing in refresh token")
	}
	return sessionID, userID, nil
}

func (uc *AuthUseCase) logoutMessage(ctx context.Context, userID string) string {
	if userID == "" {
		return ""
	}
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return userID
	}
	return user.Email
}

func (uc *AuthUseCase) logLogout(
	ctx context.Context,
	userID, sessionID *string,
	message, ipAddress, userAgent string,
	success bool,
) {
	logAuthEvent(
		ctx,
		uc.auditRepo,
		userID,
		sessionID,
		authdomain.AuditActionUserLogout,
		authdomain.AuditResultFromBool(success),
		message,
		ipAddress,
		userAgent,
	)
}
