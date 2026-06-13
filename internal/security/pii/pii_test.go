package pii

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	require.Equal(t, "admin@example.com", Normalize("  Admin@Example.com  "))
}

func TestHashDeterministic(t *testing.T) {
	normalized := "admin@example.com"
	require.Equal(t, Hash(normalized), Hash(normalized))
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := []byte("12345678901234567890123456789012")

	encrypted, err := Encrypt("admin@example.com", key)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	decrypted, err := Decrypt(encrypted, key)
	require.NoError(t, err)
	require.Equal(t, "admin@example.com", decrypted)
}

func TestEncryptRejectsInvalidKeyLength(t *testing.T) {
	_, err := Encrypt("admin@example.com", []byte("short"))
	require.Error(t, err)
}

func TestDecryptRejectsMalformedCiphertext(t *testing.T) {
	_, err := Decrypt("not-base64", []byte("12345678901234567890123456789012"))
	require.Error(t, err)
}

func TestProcessReturnsNormalizedHashAndEncrypted(t *testing.T) {
	key := []byte("12345678901234567890123456789012")

	processed, err := Process("  Admin@Example.com  ", key)
	require.NoError(t, err)
	require.Equal(t, "admin@example.com", processed.Normalized)
	require.Equal(t, Hash("admin@example.com"), processed.Hash)
	require.NotEmpty(t, processed.Encrypted)

	decrypted, err := Decrypt(processed.Encrypted, key)
	require.NoError(t, err)
	require.Equal(t, "admin@example.com", decrypted)
}
