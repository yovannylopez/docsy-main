package container

import (
	"log"

	"github.com/jmoiron/sqlx"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/policies"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	"github.com/yovannylopez/docsy-main/internal/auth/infrastructure/identity"
	"github.com/yovannylopez/docsy-main/internal/auth/infrastructure/repositories"
	"github.com/yovannylopez/docsy-main/internal/auth/infrastructure/security"
	"github.com/yovannylopez/docsy-main/internal/auth/infrastructure/services"
	"github.com/yovannylopez/docsy-main/internal/auth/transport/handlers"
	authmiddleware "github.com/yovannylopez/docsy-main/internal/auth/transport/middleware"
	"github.com/yovannylopez/docsy-main/internal/auth/usecases"
	sharedcfg "github.com/yovannylopez/docsy-main/internal/shared/infrastructure/config"
	"github.com/yovannylopez/docsy-main/pkg/constants"
)

// AuthContainer holds all dependencies for the auth module
type AuthContainer struct {
	// Repositories
	UserRepository    ports.UserRepository
	SessionRepository ports.SessionRepository

	// Security services
	PasswordHasher ports.PasswordHasher
	TokenGenerator ports.TokenGenerator

	// Use cases
	LoginUseCase          ports.LoginService
	AuthUseCase           ports.AuthenticationService
	ListAuditLogsUseCase  ports.ListAuditLogsUseCase
	ChangePasswordUseCase ports.ChangePasswordService

	// MFA use cases
	MFASetupUseCase   ports.MFASetupService
	MFAConfirmUseCase ports.MFAConfirmService
	MFAVerifyUseCase  ports.MFAVerifyService
	MFADisableUseCase ports.MFADisableService

	// MFA handler
	MFAHandler *handlers.MFAHandler

	// Services
	AuditRepository ports.AuditRepository
	AuditService    *services.AuditService

	// Handlers
	AuthHandler  *handlers.AuthHandler
	AuditHandler *handlers.AuditHandler

	AuthHTTPMiddleware *authmiddleware.AuthMiddleware
}

// NewAuthContainer creates a new instance of the auth container
func NewAuthContainer(db *sqlx.DB, jwtSecret string, ldapCfg sharedcfg.LDAPConfig) *AuthContainer {
	return newAuthContainerFull(db, jwtSecret, ldapCfg, sharedcfg.MFAConfig{}, policies.DefaultFailedLoginLockout())
}

// NewAuthContainerWithMFA creates a new instance of the auth container with MFA TOTP support.
func NewAuthContainerWithMFA(
	db *sqlx.DB, jwtSecret string, ldapCfg sharedcfg.LDAPConfig, mfaCfg sharedcfg.MFAConfig,
	lockout policies.FailedLoginLockoutPolicy,
) *AuthContainer {
	return newAuthContainerFull(db, jwtSecret, ldapCfg, mfaCfg, lockout)
}

func newAuthContainerFull(
	db *sqlx.DB, jwtSecret string, ldapCfg sharedcfg.LDAPConfig, mfaCfg sharedcfg.MFAConfig,
	lockout policies.FailedLoginLockoutPolicy,
) *AuthContainer {
	// Create repositories and security services
	userRepo := repositories.NewUserRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	passwordHasher := security.NewPasswordHasher()
	tokenGenerator := security.NewTokenGenerator(jwtSecret)
	tokenRepo := repositories.NewVerificationTokenRepository(db)

	// Create domain policies from application configuration
	sessionPolicy := policies.SessionPolicy{ExpirationDays: constants.SessionExpirationDays}
	passwordPolicy := policies.PasswordPolicy{
		MinLength:        constants.MinPasswordLength,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumber:    true,
		RequireSymbol:    true,
		AllowedSymbols:   policies.DefaultSymbols,
	}

	// Create use cases and audit services
	auditRepo := repositories.NewAuditRepositoryAdapter(db)
	auditService := services.NewAuditService(auditRepo)
	var extAuth ports.ExternalIdentityProvider = identity.NoopExternalIdentity{}
	if ldapCfg.Enabled {
		extAuth = identity.NewLDAPProvider(ldapCfg)
	}

	loginUC := usecases.NewLoginUseCase(
		userRepo, sessionRepo, passwordHasher, tokenGenerator, auditRepo, sessionPolicy, extAuth, lockout,
	)

	authUseCase := usecases.NewAuthUseCase(userRepo, tokenGenerator, sessionRepo, auditRepo)
	listAuditLogsUseCase := usecases.NewListAuditLogsUseCase(auditRepo)
	passwordHistoryRepo := repositories.NewPasswordHistoryRepository(db)
	systemConfigRepo := repositories.NewSystemConfigRepository(db)
	changePasswordUseCase := usecases.NewChangePasswordUseCase(
		userRepo, sessionRepo, passwordHasher, auditRepo, passwordPolicy,
		passwordHistoryRepo, systemConfigRepo,
	)

	authHTTPMiddleware := authmiddleware.NewAuthMiddleware(authUseCase)

	// Wire MFA components when a secret key is available
	var (
		mfaSetupUC   ports.MFASetupService
		mfaConfirmUC ports.MFAConfirmService
		mfaVerifyUC  ports.MFAVerifyService
		mfaDisableUC ports.MFADisableService
		mfaHandler   *handlers.MFAHandler
	)

	if mfaCfg.SecretKey != "" {
		encryptor, err := security.NewAESGCMEncryptor(mfaCfg.SecretKey)
		if err != nil {
			log.Printf("[WARN] MFA disabled: invalid MFA_SECRET_KEY: %v", err)
		} else {
			totpProvider := security.NewTOTPProviderAdapter(mfaCfg.Issuer)

			mfaSetupUC = usecases.NewSetupMFAUseCase(userRepo, tokenRepo, totpProvider, encryptor)
			mfaConfirmUC = usecases.NewConfirmMFAUseCase(userRepo, tokenRepo, totpProvider, encryptor)
			mfaVerifyUC = usecases.NewVerifyMFAUseCase(
				userRepo, sessionRepo, tokenRepo, tokenGenerator, encryptor, totpProvider, auditRepo, sessionPolicy,
			)
			mfaDisableUC = usecases.NewDisableMFAUseCase(userRepo, totpProvider, encryptor)
			mfaHandler = handlers.NewMFAHandler(mfaSetupUC, mfaConfirmUC, mfaVerifyUC, mfaDisableUC)

			// Attach MFA dependencies to the login use case for the challenge fork
			loginUC.WithMFA(tokenRepo, encryptor)
		}
	}

	// If MFA is not configured, wire a noop handler so routes don't panic
	if mfaHandler == nil {
		mfaHandler = handlers.NewMFAHandler(
			noopMFASetup{}, noopMFAConfirm{}, noopMFAVerify{}, noopMFADisable{},
		)
	}

	// Create handlers
	authHandler := handlers.NewAuthHandler(loginUC, authUseCase, changePasswordUseCase)
	auditHandler := handlers.NewAuditHandler(listAuditLogsUseCase)

	return &AuthContainer{
		UserRepository:        userRepo,
		SessionRepository:     sessionRepo,
		PasswordHasher:        passwordHasher,
		TokenGenerator:        tokenGenerator,
		LoginUseCase:          loginUC,
		AuthUseCase:           authUseCase,
		ListAuditLogsUseCase:  listAuditLogsUseCase,
		ChangePasswordUseCase: changePasswordUseCase,
		MFASetupUseCase:       mfaSetupUC,
		MFAConfirmUseCase:     mfaConfirmUC,
		MFAVerifyUseCase:      mfaVerifyUC,
		MFADisableUseCase:     mfaDisableUC,
		MFAHandler:            mfaHandler,
		AuditRepository:       auditRepo,
		AuditService:          auditService,
		AuthHandler:           authHandler,
		AuditHandler:          auditHandler,
		AuthHTTPMiddleware:    authHTTPMiddleware,
	}
}

// GetAuthHandler returns the authentication handler
func (c *AuthContainer) GetAuthHandler() *handlers.AuthHandler {
	return c.AuthHandler
}

// GetMFAHandler returns the MFA handler
func (c *AuthContainer) GetMFAHandler() *handlers.MFAHandler {
	return c.MFAHandler
}

// GetAuditHandler returns the audit handler
func (c *AuthContainer) GetAuditHandler() *handlers.AuditHandler {
	return c.AuditHandler
}

// GetAuditRepository returns the audit repository
func (c *AuthContainer) GetAuditRepository() ports.AuditRepository {
	return c.AuditRepository
}

// GetAuditService returns the audit service
func (c *AuthContainer) GetAuditService() *services.AuditService {
	return c.AuditService
}

// GetPasswordHasher returns the password hasher (compatible with users.PasswordHasher)
func (c *AuthContainer) GetPasswordHasher() ports.PasswordHasher {
	return c.PasswordHasher
}
