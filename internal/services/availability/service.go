package availability

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/fresp/Statora/internal/domain/shared"
	"github.com/fresp/Statora/internal/models"
	"github.com/fresp/Statora/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PeriodInfo describes the resolved reporting period.
type PeriodInfo struct {
	Label string    `json:"label"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// OverallAvailability is the platform-wide availability for the period.
type OverallAvailability struct {
	Availability    float64 `json:"availability"`
	TotalMinutes    float64 `json:"totalMinutes"`
	DowntimeMinutes float64 `json:"downtimeMinutes"`
	IncidentCount   int     `json:"incidentCount"`
}

// IncidentBreakdown is one incident's contribution to the period downtime.
type IncidentBreakdown struct {
	ID                       string                         `json:"id"`
	Title                    string                         `json:"title"`
	Impact                   models.IncidentImpact          `json:"impact"`
	Status                   models.IncidentStatus          `json:"status"`
	StartedAt                time.Time                      `json:"startedAt"`
	ResolvedAt               *time.Time                     `json:"resolvedAt"`
	EffectiveDowntimeMinutes float64                        `json:"effectiveDowntimeMinutes"`
	AffectedComponents       []IncidentAffectedComponentRef `json:"affectedComponents"`
}

// IncidentAffectedComponentRef names a component targeted by an incident.
type IncidentAffectedComponentRef struct {
	ID   primitive.ObjectID `json:"id"`
	Name string             `json:"name"`
}

// ServiceAvailability is one component's availability for the period.
type ServiceAvailability struct {
	ComponentID     primitive.ObjectID `json:"componentId"`
	Name            string             `json:"name"`
	Availability    float64            `json:"availability"`
	DowntimeMinutes float64            `json:"downtimeMinutes"`
	IncidentCount   int                `json:"incidentCount"`
}

// AvailabilityResult is the full API response payload.
type AvailabilityResult struct {
	Period    PeriodInfo            `json:"period"`
	Overall   OverallAvailability   `json:"overall"`
	Incidents []IncidentBreakdown   `json:"incidents"`
	Services  []ServiceAvailability `json:"services"`
}

// Service computes availability metrics from incident downtime.
type Service struct {
	repo repository.StatusRepository
	now  func() time.Time
}

// NewService creates an availability service. Pass nil for now to use time.Now.
func NewService(repo repository.StatusRepository, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{repo: repo, now: now}
}

// ComputeAvailability calculates overall and per-service availability for the
// period [periodStart, periodEnd).
func (s *Service) ComputeAvailability(ctx context.Context, periodStart, periodEnd time.Time, label string) (*AvailabilityResult, error) {
	if !periodStart.Before(periodEnd) {
		return nil, fmt.Errorf("%w: period start must be before end", shared.ErrInvalidInput)
	}

	incidents, err := s.repo.ListIncidentsOverlappingPeriod(ctx, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	components, err := s.repo.ListComponents(ctx)
	if err != nil {
		return nil, err
	}

	now := s.now()

	// Collect clipped intervals + valid incidents.
	type entry struct {
		incident  models.Incident
		interval  Interval
		targetIDs []primitive.ObjectID
	}
	entries := make([]entry, 0, len(incidents))
	for _, incident := range incidents {
		if !incident.CreatedAt.Before(periodEnd) {
			continue
		}
		end := now
		if incident.ResolvedAt != nil {
			end = *incident.ResolvedAt
		}
		if !incident.CreatedAt.Before(end) {
			// Defensive: invalid or zero-length window.
			continue
		}
		clipped := ClipInterval(Interval{Start: incident.CreatedAt, End: end}, periodStart, periodEnd)
		if clipped == nil {
			continue
		}
		entries = append(entries, entry{
			incident:  incident,
			interval:  *clipped,
			targetIDs: incidentTargetComponentIDs(incident),
		})
	}

	// Overall downtime: merge all intervals.
	allIntervals := make([]Interval, 0, len(entries))
	for _, e := range entries {
		allIntervals = append(allIntervals, e.interval)
	}
	merged := MergeIntervals(allIntervals)
	totalMinutes := periodEnd.Sub(periodStart).Minutes()
	downtime := TotalDuration(merged)

	result := &AvailabilityResult{
		Period: PeriodInfo{
			Label: label,
			Start: periodStart,
			End:   periodEnd,
		},
		Overall: OverallAvailability{
			Availability:    AvailabilityPercent(totalMinutes, downtime.Minutes()),
			TotalMinutes:    totalMinutes,
			DowntimeMinutes: downtime.Minutes(),
			IncidentCount:   len(entries),
		},
		Incidents: []IncidentBreakdown{},
		Services:  []ServiceAvailability{},
	}

	// Incidents breakdown, newest first.
	sortedEntries := make([]entry, len(entries))
	copy(sortedEntries, entries)
	sort.Slice(sortedEntries, func(i, j int) bool {
		return sortedEntries[i].incident.CreatedAt.After(sortedEntries[j].incident.CreatedAt)
	})
	for _, e := range sortedEntries {
		result.Incidents = append(result.Incidents, IncidentBreakdown{
			ID:                       e.incident.ID.Hex(),
			Title:                    e.incident.Title,
			Impact:                   e.incident.Impact,
			Status:                   e.incident.Status,
			StartedAt:                e.incident.CreatedAt,
			ResolvedAt:               e.incident.ResolvedAt,
			EffectiveDowntimeMinutes: e.interval.End.Sub(e.interval.Start).Minutes(),
			AffectedComponents:       s.componentRefs(e.targetIDs, components),
		})
	}

	// Per-service breakdown.
	downtimeByComponent := make(map[primitive.ObjectID]time.Duration, len(components))
	countByComponent := make(map[primitive.ObjectID]int, len(components))
	for _, component := range components {
		var componentIntervals []Interval
		for _, e := range entries {
			for _, id := range e.targetIDs {
				if id == component.ID {
					componentIntervals = append(componentIntervals, e.interval)
					countByComponent[component.ID]++
					break
				}
			}
		}
		if len(componentIntervals) == 0 {
			continue
		}
		downtimeByComponent[component.ID] = TotalDuration(MergeIntervals(componentIntervals))
	}

	for _, component := range components {
		downtimeForComponent, ok := downtimeByComponent[component.ID]
		if !ok {
			continue
		}
		result.Services = append(result.Services, ServiceAvailability{
			ComponentID:     component.ID,
			Name:            component.Name,
			Availability:    AvailabilityPercent(totalMinutes, downtimeForComponent.Minutes()),
			DowntimeMinutes: downtimeForComponent.Minutes(),
			IncidentCount:   countByComponent[component.ID],
		})
	}

	return result, nil
}

// componentRefs resolves target IDs to id/name refs, skipping unknown IDs.
func (s *Service) componentRefs(targetIDs []primitive.ObjectID, components []models.Component) []IncidentAffectedComponentRef {
	if len(targetIDs) == 0 {
		return []IncidentAffectedComponentRef{}
	}
	byID := make(map[primitive.ObjectID]string, len(components))
	for _, component := range components {
		byID[component.ID] = component.Name
	}
	refs := make([]IncidentAffectedComponentRef, 0, len(targetIDs))
	for _, id := range targetIDs {
		if name, ok := byID[id]; ok {
			refs = append(refs, IncidentAffectedComponentRef{ID: id, Name: name})
		}
	}
	if len(refs) == 0 {
		return []IncidentAffectedComponentRef{}
	}
	return refs
}

// incidentTargetComponentIDs returns the distinct component IDs an incident
// targets, merging the legacy flat list with the structured targets
// (sub-component targets count against the parent component).
func incidentTargetComponentIDs(incident models.Incident) []primitive.ObjectID {
	seen := make(map[primitive.ObjectID]struct{})
	ids := make([]primitive.ObjectID, 0, len(incident.AffectedComponents)+len(incident.AffectedComponentTargets))

	add := func(id primitive.ObjectID) {
		if id == primitive.NilObjectID {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	for _, id := range incident.AffectedComponents {
		add(id)
	}
	for _, target := range incident.AffectedComponentTargets {
		add(target.ComponentID)
	}

	return ids
}
