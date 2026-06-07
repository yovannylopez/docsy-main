package security

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAESGCMEncryptor_EncryptDecrypt(t *testing.T) {
	// 32-byte hex key (64 chars)
	const hexKey = "0000000000000000000000000000000000000000000000000000000000000001"

	enc, err := NewAESGCMEncryptor(hexKey)
	require.NoError(t, err)

	tests := []struct {
		name      string
		plaintext string
	}{
		{"short secret", "JBSWY3DPEHPK3PXP"},
		{"longer secret", strings.Repeat("ABCDEF", 10)},
		{"empty", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			ciphertext, err := enc.Encrypt(ctx, tc.plaintext)
			require.NoError(t, err)
			assert.NotEmpty(t, ciphertext)
			assert.NotEqual(t, tc.plaintext, ciphertext)

			decrypted, err := enc.Decrypt(ctx, ciphertext)
			require.NoError(t, err)
			assert.Equal(t, tc.plaintext, decrypted)
		})
	}
}

func TestAESGCMEncryptor_DifferentCiphertexts(t *testing.T) {
	const hexKey = "0000000000000000000000000000000000000000000000000000000000000001"
	enc, _ := NewAESGCMEncryptor(hexKey)

	c1, _ := enc.Encrypt(context.Background(), "same")
	c2, _ := enc.Encrypt(context.Background(), "same")
	// Each encryption uses a random nonce; ciphertexts must differ
	assert.NotEqual(t, c1, c2, "ciphertexts must be nonce-unique")
}

func TestNewAESGCMEncryptor_InvalidKey(t *testing.T) {
	tests := []struct {
		name   string
		hexKey string
	}{
		{"too short", "deadbeef"},
		{"non-hex", strings.Repeat("ZZ", 32)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewAESGCMEncryptor(tc.hexKey)
			assert.Error(t, err)
		})
	}
}
