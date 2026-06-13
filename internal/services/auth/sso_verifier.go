package auth

import (
	"crypto/rsa"
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/fresp/Statora/internal/models"
)

type externalSSOClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type jwtSSOVerifier struct {
	settings  models.StatusPageSSOSettings
	hmacKey   []byte
	rsaPubKey *rsa.PublicKey
}

func newSSOVerifier(settings models.StatusPageSSOSettings) (ssoVerifier, error) {
	algorithm := strings.ToUpper(strings.TrimSpace(settings.Algorithm))
	verifier := &jwtSSOVerifier{settings: settings}

	switch algorithm {
	case "HS256":
		secret := strings.TrimSpace(settings.SharedSecret)
		if secret == "" {
			return nil, ErrSSONotConfigured
		}
		verifier.hmacKey = []byte(secret)
	case "RS256":
		publicKeyPEM := strings.TrimSpace(settings.PublicKeyPEM)
		if publicKeyPEM == "" {
			return nil, ErrSSONotConfigured
		}
		publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyPEM))
		if err != nil {
			return nil, ErrInvalidToken
		}
		verifier.rsaPubKey = publicKey
	default:
		return nil, ErrSSONotConfigured
	}

	verifier.settings.Algorithm = algorithm
	return verifier, nil
}

func (v *jwtSSOVerifier) Verify(rawToken string) (*externalSSOClaims, error) {
	claims := &externalSSOClaims{}
	parser := jwt.NewParser(
		jwt.WithIssuer(strings.TrimSpace(v.settings.Issuer)),
		jwt.WithAudience(strings.TrimSpace(v.settings.Audience)),
	)

	token, err := parser.ParseWithClaims(rawToken, claims, v.keyFunc)
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	if strings.TrimSpace(claims.Email) == "" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (v *jwtSSOVerifier) keyFunc(token *jwt.Token) (interface{}, error) {
	switch v.settings.Algorithm {
	case "HS256":
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return v.hmacKey, nil
	case "RS256":
		if token.Method != jwt.SigningMethodRS256 {
			return nil, errors.New("unexpected signing method")
		}
		return v.rsaPubKey, nil
	default:
		return nil, ErrSSONotConfigured
	}
}
