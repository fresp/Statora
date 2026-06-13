package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"github.com/fresp/Statora/internal/models"
)

func TestSSOVerifierHS256(t *testing.T) {
	settings := models.StatusPageSSOSettings{
		Enabled:      true,
		Issuer:       "oca-dashboard",
		Audience:     "statora",
		Algorithm:    "HS256",
		SharedSecret: "shared-secret",
	}

	verifier, err := newSSOVerifier(settings)
	assert.NoError(t, err)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, externalSSOClaims{
		Email: "user@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "oca-dashboard",
			Audience:  jwt.ClaimStrings{"statora"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	raw, err := token.SignedString([]byte("shared-secret"))
	assert.NoError(t, err)

	claims, err := verifier.Verify(raw)
	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", claims.Email)
}

func TestSSOVerifierRS256(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyDER})

	settings := models.StatusPageSSOSettings{
		Enabled:      true,
		Issuer:       "oca-dashboard",
		Audience:     "statora",
		Algorithm:    "RS256",
		PublicKeyPEM: string(publicKeyPEM),
	}

	verifier, err := newSSOVerifier(settings)
	assert.NoError(t, err)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, externalSSOClaims{
		Email: "user@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "oca-dashboard",
			Audience:  jwt.ClaimStrings{"statora"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	raw, err := token.SignedString(privateKey)
	assert.NoError(t, err)

	claims, err := verifier.Verify(raw)
	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", claims.Email)
}
