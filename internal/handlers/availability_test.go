package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/fresp/Statora/internal/domain/shared"
	"github.com/fresp/Statora/internal/models"
	"github.com/fresp/Statora/internal/services/availability"
)

type stubAvailabilityService struct {
	result *availability.AvailabilityResult
	err    error

	lastStart time.Time
	lastEnd   time.Time
	lastLabel string
	calls     int
}

func (s *stubAvailabilityService) ComputeAvailability(_ context.Context, start, end time.Time, label string) (*availability.AvailabilityResult, error) {
	s.lastStart = start
	s.calls++
	s.lastEnd = end
	s.lastLabel = label
	if s.err != nil {
		return nil, s.err
	}
	return s.result, nil
}

func availabilityResultFixture() *availability.AvailabilityResult {
	componentID := primitive.NewObjectID()
	resolvedAt := time.Date(2026, 3, 15, 10, 37, 0, 0, time.UTC)

	return &availability.AvailabilityResult{
		Period: availability.PeriodInfo{
			Label: "Last 30 Days",
			Start: time.Date(2026, 2, 26, 14, 30, 0, 0, time.UTC),
			End:   time.Date(2026, 3, 28, 14, 30, 0, 0, time.UTC),
		},
		Overall: availability.OverallAvailability{
			Availability:    99.0718,
			TotalMinutes:    43200,
			DowntimeMinutes: 401,
			IncidentCount:   1,
		},
		Incidents: []availability.IncidentBreakdown{
			{
				ID:                       primitive.NewObjectID().Hex(),
				Title:                    "Intermittent WhatsApp API",
				Impact:                   models.ImpactMinor,
				Status:                   models.IncidentResolved,
				StartedAt:                time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC),
				ResolvedAt:               &resolvedAt,
				EffectiveDowntimeMinutes: 37,
				AffectedComponents: []availability.IncidentAffectedComponentRef{
					{ID: componentID, Name: "WhatsApp API"},
				},
			},
		},
		Services: []availability.ServiceAvailability{
			{
				ComponentID:     componentID,
				Name:            "WhatsApp API",
				Availability:    99.9144,
				DowntimeMinutes: 37,
				IncidentCount:   1,
			},
		},
	}
}

func setupAvailabilityRouter(t *testing.T, svc *stubAvailabilityService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/status/availability", getAvailabilityWithService(svc, nil))
	return router
}

func TestGetAvailabilityReturnsPayload(t *testing.T) {
	svc := &stubAvailabilityService{result: availabilityResultFixture()}
	router := setupAvailabilityRouter(t, svc)

	req, _ := http.NewRequest(http.MethodGet, "/api/status/availability?period=30d", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body availability.AvailabilityResult
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "Last 30 Days", body.Period.Label)
	assert.InDelta(t, 99.0718, body.Overall.Availability, 0.0001)
	assert.Equal(t, 43200.0, body.Overall.TotalMinutes)
	assert.Equal(t, 401.0, body.Overall.DowntimeMinutes)
	assert.Equal(t, 1, body.Overall.IncidentCount)
	require.Len(t, body.Incidents, 1)
	assert.Equal(t, "Intermittent WhatsApp API", body.Incidents[0].Title)
	require.Len(t, body.Services, 1)
	assert.Equal(t, "WhatsApp API", body.Services[0].Name)
}

func TestGetAvailabilityDefaultPeriodIs30d(t *testing.T) {
	svc := &stubAvailabilityService{result: availabilityResultFixture()}
	router := setupAvailabilityRouter(t, svc)

	req, _ := http.NewRequest(http.MethodGet, "/api/status/availability", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Last 30 Days", svc.lastLabel)
}

func TestGetAvailabilityYearParam(t *testing.T) {
	svc := &stubAvailabilityService{result: availabilityResultFixture()}
	router := setupAvailabilityRouter(t, svc)

	req, _ := http.NewRequest(http.MethodGet, "/api/status/availability?year=2025", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "2025 YTD", svc.lastLabel)

	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, start.Equal(svc.lastStart))
}

func TestGetAvailabilityCustomRange(t *testing.T) {
	svc := &stubAvailabilityService{result: availabilityResultFixture()}
	router := setupAvailabilityRouter(t, svc)

	req, _ := http.NewRequest(http.MethodGet, "/api/status/availability?from=2026-03-01&to=2026-03-15", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	expectedStart := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC) // date-only "to" = full day
	assert.True(t, expectedStart.Equal(svc.lastStart))
	assert.True(t, expectedEnd.Equal(svc.lastEnd))
}

func TestGetAvailabilityValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantErr    string
	}{
		{
			name:       "invalid period",
			query:      "?period=7w",
			wantStatus: http.StatusBadRequest,
			wantErr:    "invalid period",
		},
		{
			name:       "invalid year",
			query:      "?year=abcd",
			wantStatus: http.StatusBadRequest,
			wantErr:    "invalid year",
		},
		{
			name:       "future year",
			query:      "?year=2999",
			wantStatus: http.StatusBadRequest,
			wantErr:    "year must be between",
		},
		{
			name:       "invalid from format",
			query:      "?from=not-a-date&to=2026-03-15",
			wantStatus: http.StatusBadRequest,
			wantErr:    "invalid from",
		},
		{
			name:       "invalid to format",
			query:      "?from=2026-03-01&to=also-bad",
			wantStatus: http.StatusBadRequest,
			wantErr:    "invalid to",
		},
		{
			name:       "from after to",
			query:      "?from=2026-03-15&to=2026-03-01",
			wantStatus: http.StatusBadRequest,
			wantErr:    "from must be before to",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &stubAvailabilityService{result: availabilityResultFixture()}
			router := setupAvailabilityRouter(t, svc)

			req, _ := http.NewRequest(http.MethodGet, "/api/status/availability"+tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			require.Equal(t, tt.wantStatus, w.Code)

			var body map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &body)
			require.NoError(t, err)
			assert.Contains(t, body["error"], tt.wantErr)
			assert.Zero(t, svc.calls, "service must not be called on validation failure")
		})
	}
}

func TestGetAvailabilityServiceError(t *testing.T) {
	svc := &stubAvailabilityService{err: shared.ErrInternal}
	router := setupAvailabilityRouter(t, svc)

	req, _ := http.NewRequest(http.MethodGet, "/api/status/availability", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var body map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.NotEmpty(t, body["error"])
}

func TestGetAvailabilityInvalidInputFromService(t *testing.T) {
	svc := &stubAvailabilityService{err: fmt.Errorf("%w: boom", shared.ErrInvalidInput)}
	router := setupAvailabilityRouter(t, svc)

	req, _ := http.NewRequest(http.MethodGet, "/api/status/availability", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
