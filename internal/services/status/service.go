package status

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fresp/StatusForge/internal/models"
	"github.com/fresp/StatusForge/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repo repository.StatusRepository
}

func NewService(repo repository.StatusRepository) *Service {
	return &Service{repo: repo}
}

type CategoryService struct {
	ID            primitive.ObjectID     `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Status        models.ComponentStatus `json:"status"`
	Uptime90d     float64                `json:"uptime90d"`
	UptimeHistory []DailyUptimeBar       `json:"uptimeHistory"`
}

type CategorySummary struct {
	Prefix          string                       `json:"prefix"`
	Name            string                       `json:"name"`
	Description     string                       `json:"description"`
	AggregateStatus string                       `json:"aggregateStatus"`
	Uptime90d       float64                      `json:"uptime90d"`
	Services        []CategoryService            `json:"services"`
	Incidents       []models.IncidentWithUpdates `json:"incidents"`
}

type DailyUptimeBar struct {
	Date          string  `json:"date"`
	UptimePercent float64 `json:"uptimePercent"`
	Status        string  `json:"status"`
}

func (s *Service) BuildCategorySummary(ctx context.Context, prefix string) (*CategorySummary, error) {
	components, err := s.repo.ListComponents(ctx)
	if err != nil {
		return nil, err
	}

	if len(components) == 0 {
		return nil, ErrCategoryNotFound
	}

	categoryComponent := findCategoryComponent(components, prefix)
	if categoryComponent == nil {
		return nil, ErrCategoryNotFound
	}

	subs, err := s.repo.ListSubComponentsByComponentIDs(ctx, []primitive.ObjectID{categoryComponent.ID})
	if err != nil {
		return nil, err
	}

	subIDs := make([]primitive.ObjectID, 0, len(subs))
	for _, sub := range subs {
		subIDs = append(subIDs, sub.ID)
	}

	monitors, err := s.repo.ListMonitorsByTargets(ctx, []primitive.ObjectID{categoryComponent.ID}, subIDs)
	if err != nil {
		return nil, err
	}

	monitorIDs := make([]primitive.ObjectID, 0, len(monitors))
	for _, monitor := range monitors {
		monitorIDs = append(monitorIDs, monitor.ID)
	}

	uptimeRecords, err := s.repo.ListDailyUptimeSinceByMonitorIDs(ctx, monitorIDs, time.Now().AddDate(0, 0, -90))
	if err != nil {
		return nil, err
	}

	monitorsBySubID := map[primitive.ObjectID][]primitive.ObjectID{}
	componentMonitorIDs := []primitive.ObjectID{}
	for _, monitor := range monitors {
		if !monitor.SubComponentID.IsZero() {
			monitorsBySubID[monitor.SubComponentID] = append(monitorsBySubID[monitor.SubComponentID], monitor.ID)
			continue
		}
		if !monitor.ComponentID.IsZero() && monitor.ComponentID == categoryComponent.ID {
			componentMonitorIDs = append(componentMonitorIDs, monitor.ID)
		}
	}

	uptimeByMonitorID := map[primitive.ObjectID][]models.DailyUptime{}
	for _, record := range uptimeRecords {
		uptimeByMonitorID[record.MonitorID] = append(uptimeByMonitorID[record.MonitorID], record)
	}

	services := make([]CategoryService, 0, len(subs))
	if len(subs) > 0 {
		for _, sub := range subs {
			history := build90DayBars(monitorsBySubID[sub.ID], uptimeByMonitorID)
			services = append(services, CategoryService{
				ID:            sub.ID,
				Name:          sub.Name,
				Description:   sub.Description,
				Status:        sub.Status,
				Uptime90d:     averageUptime(history),
				UptimeHistory: history,
			})
		}
	} else {
		history := build90DayBars(componentMonitorIDs, uptimeByMonitorID)
		services = append(services, CategoryService{
			ID:            categoryComponent.ID,
			Name:          categoryComponent.Name,
			Description:   categoryComponent.Description,
			Status:        categoryComponent.Status,
			Uptime90d:     averageUptime(history),
			UptimeHistory: history,
		})
	}

	aggregateStatus := aggregateStatusFromServices(services)
	categoryUptime := 0.0
	if len(services) > 0 {
		total := 0.0
		for _, service := range services {
			total += service.Uptime90d
		}
		categoryUptime = total / float64(len(services))
	}

	affectedTargets := []primitive.ObjectID{categoryComponent.ID}
	affectedTargets = append(affectedTargets, subIDs...)

	incidents, err := s.repo.ListIncidentsByAffectedComponents(ctx, affectedTargets, 20)
	if err != nil {
		return nil, err
	}

	incidentIDs := make([]primitive.ObjectID, 0, len(incidents))
	for _, incident := range incidents {
		incidentIDs = append(incidentIDs, incident.ID)
	}

	updatesByIncident, err := s.repo.ListIncidentUpdatesByIncidentIDs(ctx, incidentIDs)
	if err != nil {
		return nil, err
	}

	incidentsWithUpdates := make([]models.IncidentWithUpdates, 0, len(incidents))
	incidentComponentIDs := make([]primitive.ObjectID, 0)
	incidentSubComponentIDs := make([]primitive.ObjectID, 0)
	for _, incident := range incidents {
		if len(incident.AffectedComponentTargets) > 0 {
			for _, target := range incident.AffectedComponentTargets {
				incidentComponentIDs = append(incidentComponentIDs, target.ComponentID)
				incidentSubComponentIDs = append(incidentSubComponentIDs, target.SubComponentIDs...)
			}
			continue
		}
		incidentComponentIDs = append(incidentComponentIDs, incident.AffectedComponents...)
	}

	incidentComponentMap := map[primitive.ObjectID]models.Component{}
	incidentSubComponentMap := map[primitive.ObjectID]models.SubComponent{}

	for _, component := range components {
		incidentComponentMap[component.ID] = component
	}
	for _, subComponent := range subs {
		incidentSubComponentMap[subComponent.ID] = subComponent
	}

	for _, incident := range incidents {
		targets := incident.AffectedComponentTargets
		if len(targets) == 0 {
			targets = make([]models.IncidentAffectedComponent, 0, len(incident.AffectedComponents))
			for _, componentID := range incident.AffectedComponents {
				targets = append(targets, models.IncidentAffectedComponent{ComponentID: componentID})
			}
		}

		expandedTargets := make([]models.IncidentAffectedComponentExpanded, 0, len(targets))
		expandedComponents := make([]models.Component, 0, len(targets))
		seenComponents := map[primitive.ObjectID]struct{}{}
		for _, target := range targets {
			component, ok := incidentComponentMap[target.ComponentID]
			if !ok {
				continue
			}

			if _, exists := seenComponents[component.ID]; !exists {
				expandedComponents = append(expandedComponents, component)
				seenComponents[component.ID] = struct{}{}
			}

			expandedSubComponents := make([]models.SubComponent, 0, len(target.SubComponentIDs))
			for _, subComponentID := range target.SubComponentIDs {
				if subComponent, exists := incidentSubComponentMap[subComponentID]; exists {
					expandedSubComponents = append(expandedSubComponents, subComponent)
				}
			}

			expandedTargets = append(expandedTargets, models.IncidentAffectedComponentExpanded{
				Component:     component,
				SubComponents: expandedSubComponents,
			})
		}

		incidentsWithUpdates = append(incidentsWithUpdates, models.IncidentWithUpdates{
			Incident:                 incident,
			Updates:                  updatesByIncident[incident.ID],
			AffectedComponents:       expandedComponents,
			AffectedComponentTargets: expandedTargets,
		})
	}

	return &CategorySummary{
		Prefix:          componentPrefix(categoryComponent.Name),
		Name:            categoryComponent.Name,
		Description:     categoryComponent.Description,
		AggregateStatus: aggregateStatus,
		Uptime90d:       categoryUptime,
		Services:        services,
		Incidents:       incidentsWithUpdates,
	}, nil
}

var ErrCategoryNotFound = fmt.Errorf("category not found")

func findCategoryComponent(components []models.Component, prefix string) *models.Component {
	normalizedPrefix := normalizeCategoryPrefix(prefix)
	if normalizedPrefix == "" {
		return nil
	}

	for i := range components {
		if componentPrefix(components[i].Name) == normalizedPrefix {
			return &components[i]
		}
	}

	for i := range components {
		if strings.HasPrefix(componentPrefix(components[i].Name), normalizedPrefix) {
			return &components[i]
		}
	}

	return nil
}

func normalizeCategoryPrefix(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return ""
	}

	parts := strings.FieldsFunc(v, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	})
	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "-")
}

func componentPrefix(name string) string {
	return normalizeCategoryPrefix(name)
}

func build90DayBars(
	monitorIDs []primitive.ObjectID,
	uptimeByMonitorID map[primitive.ObjectID][]models.DailyUptime,
) []DailyUptimeBar {
	if len(monitorIDs) == 0 {
		return []DailyUptimeBar{}
	}

	bars := make([]DailyUptimeBar, 0, 90)
	now := time.Now()

	for i := 89; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		dayKey := day.Format("2006-01-02")

		totalChecks := 0
		successfulChecks := 0

		for _, monitorID := range monitorIDs {
			for _, record := range uptimeByMonitorID[monitorID] {
				if record.Date.Format("2006-01-02") == dayKey {
					totalChecks += record.TotalChecks
					successfulChecks += record.SuccessfulChecks
				}
			}
		}
		uptime := 100.0
		status := string(models.StatusOperational)

		if totalChecks > 0 {
			uptime = (float64(successfulChecks) / float64(totalChecks)) * 100

			switch {
			case uptime < 50:
				status = string(models.StatusMajorOutage)
			case uptime < 99.9:
				status = string(models.StatusDegradedPerf)
			default:
				status = string(models.StatusOperational)
			}
		}

		bars = append(bars, DailyUptimeBar{
			Date:          dayKey,
			UptimePercent: uptime,
			Status:        status,
		})
	}

	return bars
}

func averageUptime(bars []DailyUptimeBar) float64 {
	if len(bars) == 0 {
		return 0
	}

	total := 0.0
	for _, bar := range bars {
		total += bar.UptimePercent
	}

	return total / float64(len(bars))
}

func aggregateStatusFromServices(services []CategoryService) string {
	if len(services) == 0 {
		return string(models.StatusOperational)
	}

	statuses := make([]models.ComponentStatus, 0, len(services))
	for _, service := range services {
		statuses = append(statuses, service.Status)
	}

	sort.SliceStable(statuses, func(i, j int) bool {
		return severityScore(statuses[i]) > severityScore(statuses[j])
	})

	return string(statuses[0])
}

func severityScore(status models.ComponentStatus) int {
	switch status {
	case models.StatusMajorOutage:
		return 5
	case models.StatusPartialOutage:
		return 4
	case models.StatusDegradedPerf:
		return 3
	case models.StatusMaintenance:
		return 2
	case models.StatusOperational:
		return 1
	default:
		return 0
	}
}
