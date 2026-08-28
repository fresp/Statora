package availability

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fresp/Statora/internal/domain/shared"
	"github.com/fresp/Statora/internal/models"
	"github.com/fresp/Statora/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type stubRepo struct {
	components []models.Component
	incidents  []models.Incident

	componentsErr error
	incidentsErr  error

	gotStart time.Time
	gotEnd   time.Time
}

func ctxNothing() context.Context { return context.Background() }

func (s *stubRepo) CountActiveIncidents(context.Context) (int64, error) { return 0, nil }

func (s *stubRepo) CountActiveMaintenanceAt(context.Context, time.Time) (int64, error) {
	return 0, nil
}

func (s *stubRepo) ListActiveIncidents(context.Context) ([]models.Incident, error) {
	return nil, nil
}

func (s *stubRepo) ListActiveMaintenanceAt(context.Context, time.Time) ([]models.Maintenance, error) {
	return nil, nil
}

func (s *stubRepo) ListComponentsByIDs(context.Context, []primitive.ObjectID) ([]models.Component, error) {
	return nil, nil
}

func (s *stubRepo) ListDailyUptimeSinceByMonitorIDs(context.Context, []primitive.ObjectID, time.Time) ([]models.DailyUptime, error) {
	return nil, nil
}

func (s *stubRepo) ListIncidentUpdatesByIncidentIDs(context.Context, []primitive.ObjectID) (map[primitive.ObjectID][]models.IncidentUpdate, error) {
	return nil, nil
}

func (s *stubRepo) ListIncidentsByAffectedComponents(context.Context, []primitive.ObjectID, int64) ([]models.Incident, error) {
	return nil, nil
}

func (s *stubRepo) ListIncidentsByCreatedAtRange(context.Context, time.Time, time.Time) ([]models.Incident, error) {
	return nil, nil
}

func (s *stubRepo) ListMonitorsByServiceID(context.Context, primitive.ObjectID) ([]models.Monitor, error) {
	return nil, nil
}

func (s *stubRepo) ListMonitorsByTargets(context.Context, []primitive.ObjectID, []primitive.ObjectID) ([]models.Monitor, error) {
	return nil, nil
}

func (s *stubRepo) ListMonitorLogsByMonitorIDsSince(context.Context, []primitive.ObjectID, time.Time) ([]models.MonitorLog, error) {
	return nil, nil
}

func (s *stubRepo) ListResolvedIncidentsSince(context.Context, time.Time) ([]models.Incident, error) {
	return nil, nil
}

func (s *stubRepo) ListSubComponentsByComponentIDs(context.Context, []primitive.ObjectID) ([]models.SubComponent, error) {
	return nil, nil
}

func (s *stubRepo) ListSubComponentsByIDs(context.Context, []primitive.ObjectID) ([]models.SubComponent, error) {
	return nil, nil
}

func (s *stubRepo) ListAllSubComponents(context.Context) ([]models.SubComponent, error) {
	return nil, nil
}

func (s *stubRepo) FindMonitorByID(context.Context, primitive.ObjectID) (*models.Monitor, error) {
	return nil, nil
}

func (s *stubRepo) FindMonitorBySubComponentID(context.Context, primitive.ObjectID) (*models.Monitor, error) {
	return nil, nil
}

func (s *stubRepo) FindLatestIncidentByComponent(context.Context, primitive.ObjectID) (*models.Incident, error) {
	return nil, nil
}

func (s *stubRepo) ListIncidentsOverlappingPeriod(_ context.Context, start, end time.Time) ([]models.Incident, error) {
	s.gotStart = start
	s.gotEnd = end
	return s.incidents, s.incidentsErr
}

func (s *stubRepo) ListComponents(_ context.Context) ([]models.Component, error) {
	return s.components, s.componentsErr
}

func newTestService(repo *stubRepo, now time.Time) *Service {
	return NewService(repo, func() time.Time { return now })
}

var (
	serviceNow = time.Date(2026, 8, 28, 14, 30, 0, 0, time.UTC)

	apiGatewayID = primitive.NewObjectID()
	whatsappID   = primitive.NewObjectID()
)

func component(id primitive.ObjectID, name string) models.Component {
	return models.Component{ID: id, Name: name}
}

func incidentAt(start, end time.Time, targets ...primitive.ObjectID) models.Incident {
	incident := models.Incident{
		ID:        primitive.NewObjectID(),
		Title:     "Test Incident",
		Status:    models.IncidentResolved,
		Impact:    models.ImpactMinor,
		CreatedAt: start,
	}
	if !end.IsZero() {
		resolvedAt := end
		incident.ResolvedAt = &resolvedAt
	}
	for _, id := range targets {
		incident.AffectedComponents = append(incident.AffectedComponents, id)
	}
	return incident
}

// TestComputeAvailabilityEdgeCases covers the edge-case matrix from the plan:
// no incidents, single incident, non-overlapping, overlapping, adjacency,
// clipping on both ends, active incidents, per-service attribution.
func TestComputeAvailabilityEdgeCases(t *testing.T) {
	periodStart := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC) // March: 43200 minutes

	t.Run("no incidents in period", func(t *testing.T) {
		repo := &stubRepo{components: []models.Component{}}
		svc := newTestService(repo, serviceNow)

		result, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.NoError(t, err)
		assert.InDelta(t, 100.0, result.Overall.Availability, 0)
		assert.InDelta(t, 0.0, result.Overall.DowntimeMinutes, 0)
		assert.Equal(t, 0, result.Overall.IncidentCount)
		assert.Empty(t, result.Incidents)
		assert.Empty(t, result.Services)
		assert.Equal(t, "March", result.Period.Label)
	})

	t.Run("single resolved incident fully inside", func(t *testing.T) {
		repo := &stubRepo{
			components: []models.Component{},
			incidents: []models.Incident{
				incidentAt(periodStart.Add(10*time.Hour), periodStart.Add(10*time.Hour+37*time.Minute)),
			},
		}
		svc := newTestService(repo, serviceNow)

		result, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.NoError(t, err)
		// March = 44640 minutes. (44640-37)/44640*100 = 99.9171
		assert.InDelta(t, 99.9171, result.Overall.Availability, 0.0001)
		assert.Equal(t, 1, result.Overall.IncidentCount)
		require.Len(t, result.Incidents, 1)
		assert.InDelta(t, 37.0, result.Incidents[0].EffectiveDowntimeMinutes, 0.0001)
	})

	t.Run("overlapping incidents merged not double counted", func(t *testing.T) {
		repo := &stubRepo{
			components: []models.Component{},
			incidents: []models.Incident{
				incidentAt(periodStart.Add(1*time.Hour), periodStart.Add(2*time.Hour)),
				incidentAt(periodStart.Add(90*time.Minute), periodStart.Add(3*time.Hour)),
			},
		}
		svc := newTestService(repo, serviceNow)

		result, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.NoError(t, err)
		// 2h merged window, not 2.5h sum.
		assert.InDelta(t, 120.0, result.Overall.DowntimeMinutes, 0.0001)
		assert.Equal(t, 2, result.Overall.IncidentCount)
	})

	t.Run("adjacent incidents merged", func(t *testing.T) {
		repo := &stubRepo{
			components: []models.Component{},
			incidents: []models.Incident{
				incidentAt(periodStart.Add(0), periodStart.Add(time.Hour)),
				incidentAt(periodStart.Add(time.Hour), periodStart.Add(2*time.Hour)),
			},
		}
		svc := newTestService(repo, serviceNow)

		result, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.NoError(t, err)
		assert.InDelta(t, 120.0, result.Overall.DowntimeMinutes, 0.0001)
	})

	t.Run("incident clipped at period start", func(t *testing.T) {
		repo := &stubRepo{
			components: []models.Component{},
			incidents: []models.Incident{
				incidentAt(periodStart.Add(-2*time.Hour), periodStart.Add(30*time.Minute)),
			},
		}
		svc := newTestService(repo, serviceNow)

		result, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.NoError(t, err)
		assert.InDelta(t, 30.0, result.Overall.DowntimeMinutes, 0.0001)
	})

	t.Run("incident clipped at period end", func(t *testing.T) {
		repo := &stubRepo{
			components: []models.Component{},
			incidents: []models.Incident{
				incidentAt(periodEnd.Add(-30*time.Minute), periodEnd.Add(2*time.Hour)),
			},
		}
		svc := newTestService(repo, serviceNow)

		result, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.NoError(t, err)
		assert.InDelta(t, 30.0, result.Overall.DowntimeMinutes, 0.0001)
	})

	t.Run("active incident counts until now", func(t *testing.T) {
		activeStart := periodStart.Add(2 * time.Hour)
		active := incidentAt(activeStart, time.Time{}) // no ResolvedAt
		active.Status = models.IncidentInvestigating
		repo := &stubRepo{
			components: []models.Component{},
			incidents:  []models.Incident{active},
		}
		nowInsidePeriod := periodStart.Add(3 * time.Hour)
		svc := newTestService(repo, nowInsidePeriod)

		result, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.NoError(t, err)
		// Active incident: now - start = 1h. periodEnd is in the future relative
		// to service now, so end clamps to now.
		assert.InDelta(t, 60.0, result.Overall.DowntimeMinutes, 0.0001)
		assert.Equal(t, models.IncidentInvestigating, result.Incidents[0].Status)
		require.Nil(t, result.Incidents[0].ResolvedAt)
	})

	t.Run("active incident with future period end extends to period end", func(t *testing.T) {
		active := incidentAt(periodStart.Add(2*time.Hour), time.Time{})
		repo := &stubRepo{
			components: []models.Component{},
			incidents:  []models.Incident{active},
		}
		// service now after periodEnd → interval end = periodEnd.
		svc := newTestService(repo, serviceNow)

		result, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.NoError(t, err)
		expected := periodEnd.Sub(periodStart.Add(2 * time.Hour)).Minutes()
		assert.InDelta(t, expected, result.Overall.DowntimeMinutes, 0.0001)
	})

	t.Run("incident entirely outside period excluded", func(t *testing.T) {
		repo := &stubRepo{
			components: []models.Component{},
			incidents: []models.Incident{
				incidentAt(periodEnd.Add(time.Hour), periodEnd.Add(2*time.Hour)),
				incidentAt(periodStart.Add(-3*time.Hour), periodStart.Add(-2*time.Hour)),
			},
		}
		svc := newTestService(repo, serviceNow)

		result, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.NoError(t, err)
		assert.InDelta(t, 0.0, result.Overall.DowntimeMinutes, 0)
		assert.Equal(t, 0, result.Overall.IncidentCount)
		assert.Empty(t, result.Incidents)
	})

	t.Run("invalid window skipped", func(t *testing.T) {
		invalid := incidentAt(periodStart.Add(2*time.Hour), periodStart.Add(time.Hour))
		repo := &stubRepo{
			components: []models.Component{},
			incidents:  []models.Incident{invalid},
		}
		svc := newTestService(repo, serviceNow)

		result, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.NoError(t, err)
		assert.Equal(t, 0, result.Overall.IncidentCount)
	})
}

func TestComputeAvailabilityPerService(t *testing.T) {
	periodStart := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	t.Run("single incident on one service", func(t *testing.T) {
		repo := &stubRepo{
			components: []models.Component{component(whatsappID, "WhatsApp API")},
			incidents: []models.Incident{
				incidentAt(periodStart.Add(time.Hour), periodStart.Add(2*time.Hour), whatsappID),
			},
		}
		svc := newTestService(repo, serviceNow)

		result, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.NoError(t, err)
		require.Len(t, result.Services, 1)
		assert.Equal(t, "WhatsApp API", result.Services[0].Name)
		assert.InDelta(t, 60.0, result.Services[0].DowntimeMinutes, 0.0001)
		assert.Equal(t, 1, result.Services[0].IncidentCount)
	})

	t.Run("overlapping incidents on same service merged", func(t *testing.T) {
		repo := &stubRepo{
			components: []models.Component{component(whatsappID, "WhatsApp API")},
			incidents: []models.Incident{
				incidentAt(periodStart.Add(1*time.Hour), periodStart.Add(2*time.Hour), whatsappID),
				incidentAt(periodStart.Add(90*time.Minute), periodStart.Add(3*time.Hour), whatsappID),
			},
		}
		svc := newTestService(repo, serviceNow)

		result, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.NoError(t, err)
		require.Len(t, result.Services, 1)
		assert.InDelta(t, 120.0, result.Services[0].DowntimeMinutes, 0.0001)
		assert.Equal(t, 2, result.Services[0].IncidentCount)
	})

	t.Run("incident on multiple services counts against each", func(t *testing.T) {
		repo := &stubRepo{
			components: []models.Component{
				component(apiGatewayID, "API Gateway"),
				component(whatsappID, "WhatsApp API"),
			},
			incidents: []models.Incident{
				incidentAt(periodStart.Add(time.Hour), periodStart.Add(2*time.Hour), apiGatewayID, whatsappID),
			},
		}
		svc := newTestService(repo, serviceNow)

		result, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.NoError(t, err)
		require.Len(t, result.Services, 2)
		byID := map[primitive.ObjectID]availabilityServiceAssert{}
		for _, s := range result.Services {
			byID[s.ComponentID] = availabilityServiceAssert{Name: s.Name, Downtime: s.DowntimeMinutes}
		}
		assert.Equal(t, "API Gateway", byID[apiGatewayID].Name)
		assert.InDelta(t, 60.0, byID[apiGatewayID].Downtime, 0.0001)
		assert.Equal(t, "WhatsApp API", byID[whatsappID].Name)
		assert.InDelta(t, 60.0, byID[whatsappID].Downtime, 0.0001)
	})

	t.Run("sub-component target counts against parent", func(t *testing.T) {
		incident := models.Incident{
			ID:        primitive.NewObjectID(),
			Status:    models.IncidentResolved,
			CreatedAt: periodStart.Add(time.Hour),
		}
		resolvedAt := periodStart.Add(2 * time.Hour)
		incident.ResolvedAt = &resolvedAt
		incident.AffectedComponentTargets = []models.IncidentAffectedComponent{
			{ComponentID: whatsappID, SubComponentIDs: []primitive.ObjectID{primitive.NewObjectID()}},
		}

		repo := &stubRepo{
			components: []models.Component{component(whatsappID, "WhatsApp API")},
			incidents:  []models.Incident{incident},
		}
		svc := newTestService(repo, serviceNow)

		result, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.NoError(t, err)
		require.Len(t, result.Services, 1)
		assert.InDelta(t, 60.0, result.Services[0].DowntimeMinutes, 0.0001)
	})

	t.Run("legacy and structured targets deduplicated", func(t *testing.T) {
		incident := incidentAt(periodStart.Add(time.Hour), periodStart.Add(2*time.Hour), whatsappID)
		incident.AffectedComponentTargets = []models.IncidentAffectedComponent{
			{ComponentID: whatsappID},
		}

		repo := &stubRepo{
			components: []models.Component{component(whatsappID, "WhatsApp API")},
			incidents:  []models.Incident{incident},
		}
		svc := newTestService(repo, serviceNow)

		result, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.NoError(t, err)
		require.Len(t, result.Incidents, 1)
		require.Len(t, result.Incidents[0].AffectedComponents, 1, "same component listed twice must be deduplicated")
		require.Len(t, result.Services, 1)
		assert.Equal(t, 1, result.Services[0].IncidentCount)
	})

	t.Run("unknown component target omitted from refs", func(t *testing.T) {
		unknownID := primitive.NewObjectID()
		repo := &stubRepo{
			components: []models.Component{},
			incidents: []models.Incident{
				incidentAt(periodStart.Add(time.Hour), periodStart.Add(2*time.Hour), unknownID),
			},
		}
		svc := newTestService(repo, serviceNow)

		result, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.NoError(t, err)
		require.Len(t, result.Incidents, 1)
		assert.Empty(t, result.Incidents[0].AffectedComponents)
	})
}

type availabilityServiceAssert struct {
	Name     string
	Downtime float64
}

func TestComputeAvailabilityErrors(t *testing.T) {
	periodStart := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	t.Run("invalid period rejected", func(t *testing.T) {
		repo := &stubRepo{}
		svc := newTestService(repo, serviceNow)

		_, err := svc.ComputeAvailability(ctxNothing(), periodEnd, periodStart, "reversed")
		require.Error(t, err)
		assert.True(t, errors.Is(err, shared.ErrInvalidInput))
	})

	t.Run("zero-length period rejected", func(t *testing.T) {
		repo := &stubRepo{}
		svc := newTestService(repo, serviceNow)

		_, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodStart, "zero")
		require.Error(t, err)
		assert.True(t, errors.Is(err, shared.ErrInvalidInput))
	})

	t.Run("repository error propagates", func(t *testing.T) {
		repo := &stubRepo{incidentsErr: errors.New("db down")}
		svc := newTestService(repo, serviceNow)

		_, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db down")
	})

	t.Run("components error propagates", func(t *testing.T) {
		repo := &stubRepo{componentsErr: errors.New("db down")}
		svc := newTestService(repo, serviceNow)

		_, err := svc.ComputeAvailability(ctxNothing(), periodStart, periodEnd, "March")
		require.Error(t, err)
	})
}

func TestComputeAvailabilityPeriodBoundsPassed(t *testing.T) {
	repo := &stubRepo{components: []models.Component{}}
	svc := newTestService(repo, serviceNow)

	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	_, err := svc.ComputeAvailability(ctxNothing(), start, end, "May")
	require.NoError(t, err)
	assert.True(t, start.Equal(repo.gotStart))
	assert.True(t, end.Equal(repo.gotEnd))
}

func TestCompileTimeInterfaces(t *testing.T) {
	var _ repository.StatusRepository = (*stubRepo)(nil)
}
