package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func TestCORSPolicyAllowsConfiguredOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(cors.New(corsConfig([]string{"http://localhost:5173"})))
	router.GET("/api/status/summary", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodOptions, "/api/status/summary", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want %d", resp.Code, http.StatusNoContent)
	}
	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want allowed origin", got)
	}
}

func TestCORSPolicyRejectsUnconfiguredOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(cors.New(corsConfig([]string{"http://localhost:5173"})))
	router.GET("/api/status/summary", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodOptions, "/api/status/summary", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty for rejected origin", got)
	}
}
