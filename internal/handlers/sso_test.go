package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	authservice "github.com/fresp/Statora/internal/services/auth"
)

type stubSSOService struct {
	result *authservice.LoginResult
	err    error
}

func (s *stubSSOService) AuthenticateSSO(_ context.Context, _ string) (*authservice.LoginResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.result, nil
}

func TestSSOCallbackRedirectsToAdminAndSetsCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	result := &authservice.LoginResult{Token: "internal-token"}
	router := gin.New()
	router.GET("/sso/callback", ssoCallbackWithService(&stubSSOService{result: result}))

	req, _ := http.NewRequest(http.MethodGet, "/sso/callback?token=external", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusFound, resp.Code)
	assert.Equal(t, "/admin", resp.Header().Get("Location"))
	assert.NotEmpty(t, resp.Result().Cookies())
}

func TestSSOCallbackMapsErrorsToRedirects(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{name: "not configured", err: authservice.ErrSSONotConfigured, expected: "/login?error=sso_not_configured"},
		{name: "disabled", err: authservice.ErrSSODisabled, expected: "/login?error=sso_disabled"},
		{name: "user missing", err: authservice.ErrUserNotFound, expected: "/login?error=user_not_found"},
		{name: "not allowed", err: authservice.ErrSSONotAllowed, expected: "/login?error=sso_not_allowed"},
		{name: "invalid", err: authservice.ErrInvalidToken, expected: "/login?error=invalid_token"},
		{name: "other", err: errors.New("boom"), expected: "/login?error=invalid_token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/sso/callback", ssoCallbackWithService(&stubSSOService{err: tt.err}))

			req, _ := http.NewRequest(http.MethodGet, "/sso/callback?token=external", nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusFound, resp.Code)
			assert.Equal(t, tt.expected, resp.Header().Get("Location"))
		})
	}
}
