package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/fresp/Statora/configs"
	authservice "github.com/fresp/Statora/internal/services/auth"
)

type ssoService interface {
	AuthenticateSSO(ctx context.Context, rawToken string) (*authservice.LoginResult, error)
}

func SSOCallback(db *mongo.Database, cfg *configs.Config) gin.HandlerFunc {
	authSvc := authservice.NewServiceFromDB(db, cfg.JWTSecret, cfg.EmailEncryptionKey)
	return ssoCallbackWithService(authSvc)
}

func ssoCallbackWithService(authSvc ssoService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken := c.Query("token")
		if rawToken == "" {
			redirectSSOError(c, authservice.ErrInvalidToken)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := authSvc.AuthenticateSSO(ctx, rawToken)
		if err != nil {
			redirectSSOError(c, err)
			return
		}

		setAuthCookie(c, result.Token)
		c.Redirect(http.StatusFound, "/admin")
	}
}

func redirectSSOError(c *gin.Context, err error) {
	code := "invalid_token"

	switch {
	case errors.Is(err, authservice.ErrSSONotConfigured):
		code = "sso_not_configured"
	case errors.Is(err, authservice.ErrSSODisabled):
		code = "sso_disabled"
	case errors.Is(err, authservice.ErrUserNotFound):
		code = "user_not_found"
	case errors.Is(err, authservice.ErrSSONotAllowed):
		code = "sso_not_allowed"
	case errors.Is(err, authservice.ErrInvalidToken):
		code = "invalid_token"
	}

	c.Redirect(http.StatusFound, "/login?error="+url.QueryEscape(code))
}
