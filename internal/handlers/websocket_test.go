package handlers

import (
	"net/http"
	"testing"
)

func TestWebSocketOriginPolicyAllowsConfiguredOrigin(t *testing.T) {
	policy := NewOriginPolicy([]string{"http://localhost:5173"})
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://localhost:5173")

	if !policy.AllowRequest(req) {
		t.Fatal("AllowRequest() = false, want true for configured origin")
	}
}

func TestWebSocketOriginPolicyRejectsUnconfiguredOrigin(t *testing.T) {
	policy := NewOriginPolicy([]string{"http://localhost:5173"})
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "https://evil.example")

	if policy.AllowRequest(req) {
		t.Fatal("AllowRequest() = true, want false for unconfigured origin")
	}
}
