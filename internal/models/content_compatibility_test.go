package models

import "testing"

func TestMaintenanceStatusSupportsDraftAndActive(t *testing.T) {
	statuses := []MaintenanceStatus{
		MaintenanceDraft,
		MaintenanceScheduled,
		MaintenanceActive,
		MaintenanceCompleted,
	}

	if len(statuses) != 4 {
		t.Fatalf("expected 4 statuses, got %d", len(statuses))
	}
}

func TestIncidentSupportsOptionalRichContentFields(t *testing.T) {
	incident := Incident{Description: "legacy"}
	if incident.Description != "legacy" {
		t.Fatalf("expected legacy description fallback")
	}

	if incident.DescriptionJSON != nil {
		t.Fatalf("expected rich content field to remain optional")
	}
}
