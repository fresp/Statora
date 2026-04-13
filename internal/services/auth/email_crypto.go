package auth

import "github.com/fresp/Statora/internal/security/pii"

func NormalizeEmail(email string) string {
	return pii.Normalize(email)
}

func HashEmail(normalized string) string {
	return pii.Hash(normalized)
}

func EncryptEmail(plain string, key []byte) (string, error) {
	return pii.Encrypt(plain, key)
}

func DecryptEmail(ciphertext string, key []byte) (string, error) {
	return pii.Decrypt(ciphertext, key)
}
