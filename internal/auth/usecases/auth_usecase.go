package usecases

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	authdomain "github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	"github.com/yovannylopez/docsy-main/internal/auth/infrastructure/security"
	"github.com/yovannylopez/docsy-main/pkg/constants"
)

// AuthUseCase holds the repositories and services needed for token validation and refresh.
type AuthUseCase struct {
	userRepo       ports.UserRepository
	tokenGenerator ports.TokenGenerator
	sessionRepo    ports.SessionRepository
	auditRepo      ports.AuditRepository
}

// NewAuthUseCase creates a new AuthUseCase instance
func NewAuthUseCase(
	userRepo ports.UserRepository,
	tokenGenerator ports.TokenGenerator,
	sessionRepo ports.SessionRepository,
	auditRepo ports.AuditRepository,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:       userRepo,
		tokenGenerator: tokenGenerator,
		sessionRepo:    sessionRepo,
		auditRepo:      auditRepo,
	}
}

// Authenticate authenticates a user by username and password
func (uc *AuthUseCase) Authenticate(ctx context.Context, username, password string) (*entities.AuthToken, error) {
	user, err := uc.userRepo.FindByEmail(ctx, username)
	if err != nil || user == nil {
		return nil, authdomain.ErrInvalidCredentials
	}

	token, err := uc.tokenGenerator.GenerateToken(user, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate token for user %s: %w", user.ID, err)
	}

	return token, nil
}

// ValidateToken validates an access JWT and returns the user with roles and permissions.
func (uc *AuthUseCase) ValidateToken(ctx context.Context, token string) (*entities.User, error) {
	claims, err := uc.tokenGenerator.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if t, ok := claims[constants.JWTClaimType].(string); ok && t == constants.JWTTokenTypeRefresh {
		return nil, errors.New("use access token, not refresh token")
	}

	userID, ok := claims[constants.JWTClaimUserID].(string)
	if !ok || userID == "" {
		return nil, errors.New("invalid token subject")
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if !user.IsActive {
		return nil, errors.New("account is not active")
	}

	// Validate password_changed_at claim when present.
	// Tokens issued before this claim existed (legacy) are accepted for continuity;
	// once the claim is present a mismatch means the password was changed after issuance.
	if pca, ok := claims[constants.JWTClaimPasswordChangedAt]; ok {
		claimPCA, ok := pca.(float64)
		if !ok {
			return nil, errors.New("invalid token claims")
		}
		if int64(claimPCA) != user.PasswordChangedAt.Unix() {
			return nil, errors.New("token invalidated by password change")
		}
	}

	sessionID, ok := claims[constants.JWTClaimSessionID].(string)
	if !ok || sessionID == "" {
		return nil, errors.New("session_id missing in token")
	}
	if err := uc.loadAndValidateSession(ctx, sessionID, userID); err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return user, nil
}

// RefreshToken rotates tokens using a refresh JWT + session in DB.
func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (*entities.AuthToken, error) {
	userID, sessionID, err := uc.tokenGenerator.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}
	if sessionID == "" {
		return nil, errors.New("session-bound refresh required")
	}

	session, err := uc.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session lookup: %w", err)
	}
	if session == nil {
		return nil, errors.New("invalid session")
	}
	if session.UserID != userID {
		return nil, errors.New("invalid refresh token")
	}
	if !session.IsActive {
		return nil, errors.New("session revoked")
	}
	if session.RevokedAt != nil {
		return nil, errors.New("session revoked")
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("session expired")
	}

	fp := security.RefreshTokenFingerprint(refreshToken)
	if subtle.ConstantTimeCompare([]byte(session.RefreshTokenHash), []byte(fp)) != 1 {
		return nil, errors.New("invalid refresh token")
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	if !user.IsActive {
		return nil, errors.New("account is not active")
	}

	newTok, err := uc.tokenGenerator.GenerateToken(user, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	session.RefreshTokenHash = security.RefreshTokenFingerprint(newTok.RefreshToken)
	session.LastUsedAt = time.Now()
	session.ExpiresAt = time.Now().Add(time.Hour * 24 * time.Duration(constants.SessionExpirationDays))
	if err := uc.sessionRepo.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return newTok, nil
}

// Login is not implemented in this use case (use LoginUseCase).
func (uc *AuthUseCase) Login(
	ctx context.Context,
	email, password, userAgent, ipAddress string,
) (*entities.AuthToken, *entities.Session, error) {
	return nil, nil, errors.New("not implemented")
}

// Logout revokes the session using an access JWT (session_id) or a refresh JWT (session_id in claims).
func (uc *AuthUseCase) Logout(ctx context.Context, token, userAgent, ipAddress string) error {
	sessionID, userID, err := uc.resolveLogoutToken(token)
	if err != nil {
		uc.logLogout(ctx, nil, nil, "", ipAddress, userAgent, false)
		return err
	}

	var userIDPtr, sessionIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}
	if sessionID != "" {
		sessionIDPtr = &sessionID
	}
	message := uc.logoutMessage(ctx, userID)

	if err := uc.sessionRepo.RevokeSession(ctx, sessionID, authdomain.SessionRevokeReasonLogout); err != nil {
		uc.logLogout(ctx, userIDPtr, sessionIDPtr, message, ipAddress, userAgent, false)
		return err
	}

	uc.logLogout(ctx, userIDPtr, sessionIDPtr, message, ipAddress, userAgent, true)
	return nil
}

// LogoutAll revokes all sessions for a user (requires user_id in claims).
func (uc *AuthUseCase) LogoutAll(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user_id required")
	}
	return uc.sessionRepo.RevokeAllUserSessions(ctx, userID, "logout_all")
}

// ValidateSession validates a session by ID.
func (uc *AuthUseCase) ValidateSession(ctx context.Context, sessionID string) (*entities.Session, error) {
	s, err := uc.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if err := validateActiveSession(s, ""); err != nil {
		return nil, err
	}
	return s, nil
}

// RevokeSession revokes a specific session.
func (uc *AuthUseCase) RevokeSession(ctx context.Context, sessionID string, reason string) error {
	return uc.sessionRepo.RevokeSession(ctx, sessionID, reason)
}
