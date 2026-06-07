package ports

import "context"

// TOTPProvider defines the interface for TOTP (RFC 6238) operations.
// Infrastructure adapters (e.g. pquerna/otp) implement this port.
type TOTPProvider interface {
	// GenerateSecret creates a new random TOTP secret for the given issuer and account.
	// Returns the base32-encoded secret and a QR-code URL (otpauth://).
	GenerateSecret(ctx context.Context, issuer, accountName string) (secret, qrURL string, err error)

	// ValidateCode checks whether the given code is valid for the secret at the current time.
	// Uses a ±1-window tolerance to allow for minor clock drift.
	ValidateCode(ctx context.Context, secret, code string) (bool, error)
}

// MFASecretEncryptor defines the interface for encrypting and decrypting MFA TOTP secrets at rest.
// Infrastructure adapters (e.g. AES-GCM-256) implement this port.
type MFASecretEncryptor interface {
	// Encrypt returns the ciphertext (hex or base64-encoded) of the plain secret.
	Encrypt(ctx context.Context, plainSecret string) (encrypted string, err error)

	// Decrypt returns the plain secret from an encrypted ciphertext.
	Decrypt(ctx context.Context, encrypted string) (plainSecret string, err error)
}
