package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const AuthCookieName = "statora_auth"

type Claims struct {
	UserID      string `json:"userId"`
	Username    string `json:"username"`
	Role        string `json:"role,omitempty"`
	MFAVerified bool   `json:"mfaVerified,omitempty"`
	jwt.RegisteredClaims
}

type TokenClaimsInput struct {
	UserID      string
	Username    string
	Role        string
	MFAVerified bool
	Secret      string
	Now         time.Time
}

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := extractAuthToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		claims, err := parseClaims(tokenStr, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("mfaVerified", claims.MFAVerified)
		c.Next()
	}
}

func extractAuthToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			return "", errors.New("invalid authorization header format")
		}
		return parts[1], nil
	}

	cookieToken, err := c.Cookie(AuthCookieName)
	if err == nil && strings.TrimSpace(cookieToken) != "" {
		return cookieToken, nil
	}

	return "", errors.New("authorization required")
}

func parseClaims(tokenStr, jwtSecret string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		normalized := strings.ToLower(strings.TrimSpace(role))
		if normalized == "" {
			continue
		}
		allowed[normalized] = struct{}{}
	}

	return func(c *gin.Context) {
		rawRole, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		role, ok := rawRole.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		if _, ok := allowed[strings.ToLower(strings.TrimSpace(role))]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		c.Next()
	}
}

func RequireMFAVerified() gin.HandlerFunc {
	return func(c *gin.Context) {
		rawMFAVerified, exists := c.Get("mfaVerified")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "mfa verification required"})
			return
		}

		mfaVerified, ok := rawMFAVerified.(bool)
		if !ok || !mfaVerified {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "mfa verification required"})
			return
		}

		c.Next()
	}
}

func GenerateToken(userID, username, secret string) (string, error) {
	return GenerateTokenWithClaims(TokenClaimsInput{
		UserID:      userID,
		Username:    username,
		Role:        "",
		MFAVerified: true,
		Secret:      secret,
		Now:         time.Now(),
	})
}

func GenerateTokenWithClaims(input TokenClaimsInput) (string, error) {
	claims := &Claims{
		UserID:      input.UserID,
		Username:    input.Username,
		Role:        input.Role,
		MFAVerified: input.MFAVerified,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(input.Secret))
}
