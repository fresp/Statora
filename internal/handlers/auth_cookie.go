package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fresp/Statora/internal/middleware"
)

func setAuthCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(middleware.AuthCookieName, token, 0, "/", "", true, true)
}

func clearAuthCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(middleware.AuthCookieName, "", -1, "/", "", true, true)
}
