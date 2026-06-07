package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/policies"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	"github.com/yovannylopez/docsy-main/internal/auth/infrastructure/security"
	"github.com/yovannylopez/docsy-main/pkg/constants"
)

// LoginUseCase is the struct that holds the repositories and services needed for the login use case
type LoginUseCase struct {
	userRepo       ports.UserRepository
	sessionRepo    ports.SessionRepository
	passwordHasher ports.PasswordHasher
	tokenGenerator ports.TokenGenerator
	auditRepo      ports.AuditRepository
	sessionPolicy  policies.SessionPolicy
	lockout        policies.FailedLoginLockoutPolicy
	extAuth        ports.ExternalIdentityProvider
	tokenRepo      ports.VerificationTokenRepository
	encryptor      ports.MFASecretEncryptor
}

// NewLoginUseCase creates a new LoginUseCase instance
func NewLoginUseCase(
	userRepo ports.UserRepository,
	sessionRepo ports.SessionRepository,
	passwordHasher ports.PasswordHasher,
	tokenGenerator ports.TokenGenerator,
	auditRepo ports.AuditRepository,
	sessionPolicy policies.SessionPolicy,
	extAuth ports.ExternalIdentityProvider,
	lockout policies.FailedLoginLockoutPolicy,
) *LoginUseCase {
	if extAuth == nil {
		extAuth = noopExtAuth{}
	}
	return &LoginUseCase{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		passwordHasher: passwordHasher,
		tokenGenerator: tokenGenerator,
		auditRepo:      auditRepo,
		sessionPolicy:  sessionPolicy,
		lockout:        lockout,
		extAuth:        extAuth,
	}
}

// WithMFA attaches MFA-related dependencies to a LoginUseCase.
// Call after NewLoginUseCase when MFA TOTP is enabled.
func (uc *LoginUseCase) WithMFA(
	tokenRepo ports.VerificationTokenRepository, encryptor ports.MFASecretEncryptor,
) *LoginUseCase {
	uc.tokenRepo = tokenRepo
	uc.encryptor = encryptor
	return uc
}

type noopExtAuth struct{}

func (noopExtAuth) Enabled() bool { return false }

func (noopExtAuth) Authenticate(context.Context, string, string) (*entities.User, error) {
	return nil, nil
}

// Execute is the method responsible for running the login use case
func (uc *LoginUseCase) Execute(
	ctx context.Context,
	request *dtos.LoginRequest,
	userAgent,
	ipAddress string,
) (*dtos.LoginResponse, error) {
	var user *entities.User
	var err error

	if uc.extAuth.Enabled() {
		user, err = uc.extAuth.Authenticate(ctx, request.Email, request.Password)
		if err != nil {
			uc.logLoginAttempt(ctx, request.Email, ipAddress, userAgent, false, nil)
			return nil, err
		}
		if user == nil {
			uc.logLoginAttempt(ctx, request.Email, ipAddress, userAgent, false, nil)
			return nil, errors.New("invalid credentials")
		}
	} else {
		user, err = uc.userRepo.FindByEmail(ctx, request.Email)
		if err != nil || user == nil {
			uc.logLoginAttempt(ctx, request.Email, ipAddress, userAgent, false, nil)
			return nil, errors.New("invalid credentials")
		}

		if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
			uc.logLoginAttempt(ctx, request.Email, ipAddress, userAgent, false, nil)
			return nil, errors.New("account is locked")
		}
		if !user.IsActive {
			uc.logLoginAttempt(ctx, request.Email, ipAddress, userAgent, false, nil)
			return nil, errors.New("account is not active")
		}

		valid, vErr := uc.passwordHasher.VerifyPassword(request.Password, user.PasswordHash)
		if vErr != nil || !valid {
			// Local password only: count failures and optionally lock (LDAP path does not use this counter).
			rec, incErr := uc.userRepo.RecordFailedPasswordAttempt(
				ctx, user.ID, uc.lockout.MaxAttempts, uc.lockout.LockDuration,
			)
			if incErr != nil {
				_ = incErr
			}
			uc.logLoginAttempt(ctx, request.Email, ipAddress, userAgent, false, nil)
			if uc.lockout.MaxAttempts > 0 && rec.FailedAttempts >= uc.lockout.MaxAttempts {
				return nil, errors.New("account is locked")
			}
			return nil, errors.New("invalid credentials")
		}
	}

	if uc.extAuth.Enabled() {
		if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
			uc.logLoginAttempt(ctx, request.Email, ipAddress, userAgent, false, nil)
			return nil, errors.New("account is locked")
		}
		if !user.IsActive {
			uc.logLoginAttempt(ctx, request.Email, ipAddress, userAgent, false, nil)
			return nil, errors.New("account is not active")
		}
	}

	if err := uc.userRepo.ResetFailedLoginAttempts(ctx, user.ID); err != nil {
		_ = err
	}
	if err := uc.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		_ = err
	}

	// MFA challenge fork: when MFA is active, issue a challenge token instead of a full session.
	if user.MFAEnabled && uc.tokenRepo != nil {
		return uc.createMFAChallengeResponse(ctx, user)
	}

	return uc.createSessionAndLoginResponse(ctx, user, request.Email, userAgent, ipAddress)
}

func (uc *LoginUseCase) createSessionAndLoginResponse(
	ctx context.Context,
	user *entities.User,
	email, userAgent, ipAddress string,
) (*dtos.LoginResponse, error) {
	if err := uc.sessionRepo.RevokeAllUserSessions(ctx, user.ID, domain.SessionRevokeReasonNewLogin); err != nil {
		return nil, fmt.Errorf("failed to revoke previous sessions for user %s: %w", user.ID, err)
	}

	sessionID := uuid.New().String()

	token, err := uc.tokenGenerator.GenerateToken(user, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate authentication token for user %s: %w", user.ID, err)
	}

	session := &entities.Session{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: security.RefreshTokenFingerprint(token.RefreshToken),
		AccessTokenJTI:   nil,
		UserAgent:        &userAgent,
		IPAddress:        &ipAddress,
		CreatedAt:        time.Now(),
		LastUsedAt:       time.Now(),
		ExpiresAt:        time.Now().Add(time.Hour * 24 * time.Duration(uc.sessionPolicy.ExpirationDays)),
		IsActive:         true,
	}

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session for user %s: %w", user.ID, err)
	}

	uc.logLoginAttempt(ctx, email, ipAddress, userAgent, true, &session.ID)

	return uc.buildLoginResponse(token, session, user), nil
}

// createMFAChallengeResponse creates a short-lived challenge token instead of a full session.
// The client must call POST /auth/mfa/verify with this token + TOTP code to complete login.
func (uc *LoginUseCase) createMFAChallengeResponse(
	ctx context.Context, user *entities.User,
) (*dtos.LoginResponse, error) {
	rawToken, tokenHash := generateToken()
	token := &entities.VerificationToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		TokenType: domain.VerificationTokenTypeMFAChallenge,
		ExpiresAt: time.Now().Add(time.Duration(constants.MFAChallengeTokenTTLMinutes) * time.Minute),
		CreatedAt: time.Now(),
	}
	if err := uc.tokenRepo.CreateToken(ctx, token); err != nil {
		return nil, fmt.Errorf("login mfa challenge: create token: %w", err)
	}
	return &dtos.LoginResponse{
		MFARequired:       true,
		MFAChallengeToken: rawToken,
	}, nil
}

// logLoginAttempt records a login attempt in the audit log
func (uc *LoginUseCase) logLoginAttempt(
	ctx context.Context,
	email,
	ipAddress,
	userAgent string,
	success bool,
	sessionID *string,
) {
	// Look up user by email to get the ID if it exists
	user, _ := uc.userRepo.FindByEmail(ctx, email)

	var userID *string
	if user != nil {
		userID = &user.ID
	}

	// Generate request ID if not available
	requestID := "system-generated"

	auditLog := &entities.AuditLog{
		ID:         uuid.New().String(),
		UserID:     userID,
		SessionID:  sessionID,
		Action:     domain.AuditActionUserLoginAttempt,
		Resource:   nil,
		ResourceID: nil,
		Result:     domain.AuditResultFromBool(success),
		Message:    &email,
		IPAddress:  &ipAddress,
		UserAgent:  &userAgent,
		RequestID:  &requestID,
		CreatedAt:  time.Now(),
	}

	if uc.auditRepo != nil {
		if err := uc.auditRepo.LogAction(ctx, auditLog); err != nil {
			// In a production environment, an appropriate logger should be used here
			_ = err // Silence the error to avoid failing the main flow
		}
	}
}

// buildLoginResponse builds the login response
func (uc *LoginUseCase) buildLoginResponse(
	token *entities.AuthToken,
	session *entities.Session,
	user *entities.User,
) *dtos.LoginResponse {
	// Convert Session entity to DTO
	var sessionResponse *dtos.SessionResponse
	if session != nil {
		sessionResponse = &dtos.SessionResponse{
			ID:                session.ID,
			UserID:            session.UserID,
			AccessTokenJTI:    session.AccessTokenJTI,
			UserAgent:         session.UserAgent,
			IPAddress:         session.IPAddress,
			Location:          session.Location,
			DeviceFingerprint: session.DeviceFingerprint,
			CreatedAt:         session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			LastUsedAt:        session.LastUsedAt.Format("2006-01-02T15:04:05Z07:00"),
			ExpiresAt:         session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			IsActive:          session.IsActive,
			RevokedAt:         nil, // Sensitive information not included
			RevokedReason:     nil, // Sensitive information not included
		}
	}

	// Convert AuthToken entity to DTO
	tokenResponse := &dtos.TokenResponse{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		RefreshToken: token.RefreshToken,
	}

	// Convert User entity to DTO
	var userResponse *dtos.UserResponse
	if user != nil {
		userResponse = &dtos.UserResponse{
			ID:                  user.ID,
			Email:               user.Email,
			Username:            user.Username,
			FirstName:           user.FirstName,
			LastName:            user.LastName,
			Phone:               user.Phone,
			IsActive:            user.IsActive,
			IsVerified:          user.IsVerified,
			FailedLoginAttempts: user.FailedLoginAttempts,
			MFAEnabled:          user.MFAEnabled,
			PasswordChangedAt:   user.PasswordChangedAt.Format("2006-01-02T15:04:05Z07:00"),
			MustChangePassword:  user.MustChangePassword,
			CreatedAt:           user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:           user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Roles:               make([]dtos.RoleResponse, len(user.Roles)),
		}

		// Convert roles to DTO
		for i, role := range user.Roles {
			userResponse.Roles[i] = dtos.RoleResponse{
				ID:           role.ID,
				Name:         role.Name,
				Description:  role.Description,
				IsSystemRole: role.IsSystemRole,
				IsActive:     role.IsActive,
				CreatedAt:    role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:    role.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
		}
	}

	return &dtos.LoginResponse{
		Token:   tokenResponse,
		Session: sessionResponse,
		User:    userResponse,
	}
}

// Login implements the LoginService interface
func (uc *LoginUseCase) Login(
	ctx context.Context,
	request *dtos.LoginRequest,
	userAgent,
	ipAddress string,
) (*dtos.LoginResponse, error) {
	return uc.Execute(ctx, request, userAgent, ipAddress)
}
