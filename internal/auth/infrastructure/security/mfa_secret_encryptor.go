package security

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// AESGCMEncryptor implements ports.MFASecretEncryptor using AES-256-GCM.
// The key must be exactly 32 bytes (256 bits), supplied as a 64-character hex string
// via the MFA_SECRET_KEY environment variable.
type AESGCMEncryptor struct {
	key []byte
}

// aesKeySize is the required AES-256 key length in bytes.
const aesKeySize = 32

// NewAESGCMEncryptor creates a new AESGCMEncryptor from a 64-character hex key string.
func NewAESGCMEncryptor(hexKey string) (*AESGCMEncryptor, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("mfa encryptor: invalid hex key: %w", err)
	}
	if len(key) != aesKeySize {
		return nil, fmt.Errorf("mfa encryptor: key must be %d bytes (64 hex chars), got %d bytes", aesKeySize, len(key))
	}
	return &AESGCMEncryptor{key: key}, nil
}

// Encrypt returns the nonce+ciphertext as a hex string.
// Format: hex(12-byte nonce || ciphertext || 16-byte tag)
func (e *AESGCMEncryptor) Encrypt(_ context.Context, plainSecret string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("mfa encrypt: aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("mfa encrypt: gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("mfa encrypt: nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plainSecret), nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt reverses Encrypt and returns the plain secret.
func (e *AESGCMEncryptor) Decrypt(_ context.Context, encrypted string) (string, error) {
	data, err := hex.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("mfa decrypt: hex decode: %w", err)
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("mfa decrypt: aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("mfa decrypt: gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("mfa decrypt: ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("mfa decrypt: open: %w", err)
	}

	return string(plaintext), nil
}
