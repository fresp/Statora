package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateIncidentRejectsMissingAdminContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/incidents", CreateIncident(nil, nil))

	body, _ := json.Marshal(map[string]any{
		"title":              "API outage",
		"description":        "Investigating elevated errors",
		"status":             "investigating",
		"impact":             "major",
		"affectedComponents": []string{},
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/incidents", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateMaintenanceRejectsMissingAdminContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/maintenance", CreateMaintenance(nil))

	body, _ := json.Marshal(map[string]any{
		"title":       "Database maintenance",
		"description": "Planned database index optimization",
		"components":  []string{},
		"startTime":   "2026-03-20T10:00:00Z",
		"endTime":     "2026-03-20T11:00:00Z",
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/maintenance", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
