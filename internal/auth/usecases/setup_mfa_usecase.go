package usecases

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	"github.com/yovannylopez/docsy-main/pkg/constants"
)

// SetupMFAUseCase implements ports.MFASetupService (US1, step 1).
type SetupMFAUseCase struct {
	userRepo  ports.UserRepository
	tokenRepo ports.VerificationTokenRepository
	totp      ports.TOTPProvider
	encryptor ports.MFASecretEncryptor
}

// NewSetupMFAUseCase creates a new SetupMFAUseCase.
func NewSetupMFAUseCase(
	userRepo ports.UserRepository,
	tokenRepo ports.VerificationTokenRepository,
	totp ports.TOTPProvider,
	encryptor ports.MFASecretEncryptor,
) *SetupMFAUseCase {
	return &SetupMFAUseCase{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		totp:      totp,
		encryptor: encryptor,
	}
}

// Setup initiates MFA setup for the authenticated user.
// Returns the plain secret (for display/QR), the QR URL, and a single-use setup token.
func (uc *SetupMFAUseCase) Setup(ctx context.Context, userID string) (*dtos.MFASetupResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("setup mfa: user not found: %w", domain.ErrMFAInvalidToken)
	}

	if user.MFAEnabled {
		return nil, domain.ErrMFAAlreadyEnabled
	}

	// Generate TOTP secret
	secret, qrURL, err := uc.totp.GenerateSecret(ctx, "", user.Email)
	if err != nil {
		return nil, fmt.Errorf("setup mfa: generate secret: %w", err)
	}

	// Encrypt and store the pending secret
	encrypted, err := uc.encryptor.Encrypt(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("setup mfa: encrypt secret: %w", err)
	}
	if err := uc.userRepo.UpdateMFASecret(ctx, userID, encrypted); err != nil {
		return nil, fmt.Errorf("setup mfa: store pending secret: %w", err)
	}

	// Create a single-use setup token
	rawToken, tokenHash := generateToken()
	token := &entities.VerificationToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		TokenHash: tokenHash,
		TokenType: domain.VerificationTokenTypeMFASetup,
		ExpiresAt: time.Now().Add(time.Duration(constants.MFASetupTokenTTLMinutes) * time.Minute),
		CreatedAt: time.Now(),
	}
	if err := uc.tokenRepo.CreateToken(ctx, token); err != nil {
		return nil, fmt.Errorf("setup mfa: create setup token: %w", err)
	}

	return &dtos.MFASetupResponse{
		Secret:     secret,
		QRCodeURL:  qrURL,
		SetupToken: rawToken,
	}, nil
}

// tokenByteSize is the number of random bytes used to generate an opaque token.
const tokenByteSize = 32

// generateToken creates a cryptographically random opaque token and its SHA-256 hash.
// The raw token is returned to the client; only the hash is stored in the database.
func generateToken() (rawToken, tokenHash string) {
	b := make([]byte, tokenByteSize)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("generateToken: rand.Read: %v", err))
	}
	rawToken = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(rawToken))
	tokenHash = hex.EncodeToString(sum[:])
	return rawToken, tokenHash
}
