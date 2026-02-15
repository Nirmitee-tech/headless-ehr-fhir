package careteam

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCareTeam_ToFHIR(t *testing.T) {
	now := time.Now().UTC()
	patientID := uuid.New()
	ct := &CareTeam{
		ID:        uuid.New(),
		FHIRID:    "ct-123",
		Status:    "active",
		Name:      strPtr("Primary Care Team"),
		PatientID: patientID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := ct.ToFHIR()

	if result["resourceType"] != "CareTeam" {
		t.Errorf("expected resourceType 'CareTeam', got %v", result["resourceType"])
	}
	if result["id"] != "ct-123" {
		t.Errorf("expected id 'ct-123', got %v", result["id"])
	}
	if result["status"] != "active" {
		t.Errorf("expected status 'active', got %v", result["status"])
	}
	if result["name"] != "Primary Care Team" {
		t.Errorf("expected name 'Primary Care Team', got %v", result["name"])
	}
	if result["subject"] == nil {
		t.Error("expected subject to be set")
	}
}

func TestCareTeam_ToFHIR_WithParticipants(t *testing.T) {
	now := time.Now().UTC()
	patientID := uuid.New()
	memberID := uuid.New()
	ct := &CareTeam{
		ID:        uuid.New(),
		FHIRID:    "ct-456",
		Status:    "active",
		PatientID: patientID,
		CreatedAt: now,
		UpdatedAt: now,
		Participants: []CareTeamParticipant{
			{
				ID:          uuid.New(),
				CareTeamID:  uuid.New(),
				MemberID:    memberID,
				MemberType:  "Practitioner",
				RoleCode:    strPtr("doctor"),
				RoleDisplay: strPtr("Doctor"),
			},
		},
	}

	result := ct.ToFHIR()

	participants, ok := result["participant"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected participant to be []map[string]interface{}")
	}
	if len(participants) != 1 {
		t.Fatalf("expected 1 participant, got %d", len(participants))
	}
	if participants[0]["member"] == nil {
		t.Error("expected member to be set")
	}
	if participants[0]["role"] == nil {
		t.Error("expected role to be set")
	}
}

func TestCareTeam_ToFHIR_WithOptionalFields(t *testing.T) {
	now := time.Now().UTC()
	patientID := uuid.New()
	encounterID := uuid.New()
	managingOrgID := uuid.New()
	periodStart := now.Add(-24 * time.Hour)
	periodEnd := now

	ct := &CareTeam{
		ID:                     uuid.New(),
		FHIRID:                 "ct-789",
		Status:                 "active",
		Name:                   strPtr("Oncology Team"),
		PatientID:              patientID,
		EncounterID:            &encounterID,
		CategoryCode:           strPtr("longitudinal"),
		CategoryDisplay:        strPtr("Longitudinal Care Coordination"),
		PeriodStart:            &periodStart,
		PeriodEnd:              &periodEnd,
		ManagingOrganizationID: &managingOrgID,
		ReasonCode:             strPtr("diabetes"),
		ReasonDisplay:          strPtr("Diabetes Management"),
		Note:                   strPtr("Weekly reviews"),
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	result := ct.ToFHIR()

	if result["encounter"] == nil {
		t.Error("expected encounter to be set")
	}
	if result["category"] == nil {
		t.Error("expected category to be set")
	}
	if result["period"] == nil {
		t.Error("expected period to be set")
	}
	if result["managingOrganization"] == nil {
		t.Error("expected managingOrganization to be set")
	}
	if result["reasonCode"] == nil {
		t.Error("expected reasonCode to be set")
	}
	if result["note"] == nil {
		t.Error("expected note to be set")
	}
}

func TestCareTeamParticipant_ToFHIR(t *testing.T) {
	memberID := uuid.New()
	onBehalfOfID := uuid.New()
	periodStart := time.Now().UTC()

	p := &CareTeamParticipant{
		ID:           uuid.New(),
		CareTeamID:   uuid.New(),
		MemberID:     memberID,
		MemberType:   "Practitioner",
		RoleCode:     strPtr("primary"),
		RoleDisplay:  strPtr("Primary Care Physician"),
		PeriodStart:  &periodStart,
		OnBehalfOfID: &onBehalfOfID,
	}

	result := p.ToFHIR()

	if result["member"] == nil {
		t.Error("expected member to be set")
	}
	if result["role"] == nil {
		t.Error("expected role to be set")
	}
	if result["period"] == nil {
		t.Error("expected period to be set")
	}
	if result["onBehalfOf"] == nil {
		t.Error("expected onBehalfOf to be set")
	}
}

func strPtr(s string) *string { return &s }
