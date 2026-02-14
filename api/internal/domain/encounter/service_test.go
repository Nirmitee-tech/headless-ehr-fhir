package encounter

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// -- Mock Repository --

type mockRepo struct {
	encounters    map[uuid.UUID]*Encounter
	participants  map[uuid.UUID]*EncounterParticipant
	diagnoses     map[uuid.UUID]*EncounterDiagnosis
	statusHistory map[uuid.UUID]*EncounterStatusHistory
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		encounters:    make(map[uuid.UUID]*Encounter),
		participants:  make(map[uuid.UUID]*EncounterParticipant),
		diagnoses:     make(map[uuid.UUID]*EncounterDiagnosis),
		statusHistory: make(map[uuid.UUID]*EncounterStatusHistory),
	}
}

func (m *mockRepo) Create(_ context.Context, enc *Encounter) error {
	enc.ID = uuid.New()
	if enc.FHIRID == "" {
		enc.FHIRID = enc.ID.String()
	}
	enc.CreatedAt = time.Now()
	enc.UpdatedAt = time.Now()
	m.encounters[enc.ID] = enc
	return nil
}

func (m *mockRepo) GetByID(_ context.Context, id uuid.UUID) (*Encounter, error) {
	enc, ok := m.encounters[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return enc, nil
}

func (m *mockRepo) GetByFHIRID(_ context.Context, fhirID string) (*Encounter, error) {
	for _, enc := range m.encounters {
		if enc.FHIRID == fhirID {
			return enc, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockRepo) Update(_ context.Context, enc *Encounter) error {
	m.encounters[enc.ID] = enc
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.encounters, id)
	return nil
}

func (m *mockRepo) List(_ context.Context, limit, offset int) ([]*Encounter, int, error) {
	var result []*Encounter
	for _, enc := range m.encounters {
		result = append(result, enc)
	}
	return result, len(result), nil
}

func (m *mockRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*Encounter, int, error) {
	var result []*Encounter
	for _, enc := range m.encounters {
		if enc.PatientID == patientID {
			result = append(result, enc)
		}
	}
	return result, len(result), nil
}

func (m *mockRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*Encounter, int, error) {
	return m.List(context.Background(), limit, offset)
}

func (m *mockRepo) AddParticipant(_ context.Context, p *EncounterParticipant) error {
	p.ID = uuid.New()
	m.participants[p.ID] = p
	return nil
}

func (m *mockRepo) GetParticipants(_ context.Context, encounterID uuid.UUID) ([]*EncounterParticipant, error) {
	var result []*EncounterParticipant
	for _, p := range m.participants {
		if p.EncounterID == encounterID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockRepo) RemoveParticipant(_ context.Context, id uuid.UUID) error {
	delete(m.participants, id)
	return nil
}

func (m *mockRepo) AddDiagnosis(_ context.Context, d *EncounterDiagnosis) error {
	d.ID = uuid.New()
	m.diagnoses[d.ID] = d
	return nil
}

func (m *mockRepo) GetDiagnoses(_ context.Context, encounterID uuid.UUID) ([]*EncounterDiagnosis, error) {
	var result []*EncounterDiagnosis
	for _, d := range m.diagnoses {
		if d.EncounterID == encounterID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (m *mockRepo) RemoveDiagnosis(_ context.Context, id uuid.UUID) error {
	delete(m.diagnoses, id)
	return nil
}

func (m *mockRepo) AddStatusHistory(_ context.Context, sh *EncounterStatusHistory) error {
	sh.ID = uuid.New()
	m.statusHistory[sh.ID] = sh
	return nil
}

func (m *mockRepo) GetStatusHistory(_ context.Context, encounterID uuid.UUID) ([]*EncounterStatusHistory, error) {
	var result []*EncounterStatusHistory
	for _, sh := range m.statusHistory {
		if sh.EncounterID == encounterID {
			result = append(result, sh)
		}
	}
	return result, nil
}

// -- Tests --

func newTestService() *Service {
	return NewService(newMockRepo())
}

func TestCreateEncounter(t *testing.T) {
	svc := newTestService()

	patientID := uuid.New()
	enc := &Encounter{
		PatientID: patientID,
		ClassCode: "AMB",
	}
	err := svc.CreateEncounter(context.Background(), enc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enc.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if enc.Status != "planned" {
		t.Errorf("expected default status 'planned', got %s", enc.Status)
	}
	if enc.FHIRID == "" {
		t.Error("expected FHIR ID to be set")
	}
	if enc.PeriodStart.IsZero() {
		t.Error("expected period_start to be set")
	}
}

func TestCreateEncounter_PatientRequired(t *testing.T) {
	svc := newTestService()

	enc := &Encounter{ClassCode: "AMB"}
	err := svc.CreateEncounter(context.Background(), enc)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateEncounter_ClassRequired(t *testing.T) {
	svc := newTestService()

	enc := &Encounter{PatientID: uuid.New()}
	err := svc.CreateEncounter(context.Background(), enc)
	if err == nil {
		t.Error("expected error for missing class_code")
	}
}

func TestCreateEncounter_InvalidStatus(t *testing.T) {
	svc := newTestService()

	enc := &Encounter{
		PatientID: uuid.New(),
		ClassCode: "AMB",
		Status:    "bogus",
	}
	err := svc.CreateEncounter(context.Background(), enc)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestCreateEncounter_ExplicitStatus(t *testing.T) {
	svc := newTestService()

	enc := &Encounter{
		PatientID: uuid.New(),
		ClassCode: "IMP",
		Status:    "arrived",
	}
	err := svc.CreateEncounter(context.Background(), enc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enc.Status != "arrived" {
		t.Errorf("expected 'arrived', got %s", enc.Status)
	}
}

func TestGetEncounter(t *testing.T) {
	svc := newTestService()

	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	svc.CreateEncounter(context.Background(), enc)

	fetched, err := svc.GetEncounter(context.Background(), enc.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ClassCode != "AMB" {
		t.Errorf("expected AMB, got %s", fetched.ClassCode)
	}
}

func TestGetEncounterByFHIRID(t *testing.T) {
	svc := newTestService()

	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	svc.CreateEncounter(context.Background(), enc)

	fetched, err := svc.GetEncounterByFHIRID(context.Background(), enc.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != enc.ID {
		t.Errorf("expected same ID")
	}
}

func TestDeleteEncounter(t *testing.T) {
	svc := newTestService()

	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	svc.CreateEncounter(context.Background(), enc)

	err := svc.DeleteEncounter(context.Background(), enc.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.GetEncounter(context.Background(), enc.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestUpdateEncounterStatus(t *testing.T) {
	svc := newTestService()

	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	svc.CreateEncounter(context.Background(), enc)

	err := svc.UpdateEncounterStatus(context.Background(), enc.ID, "in-progress")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := svc.GetEncounter(context.Background(), enc.ID)
	if updated.Status != "in-progress" {
		t.Errorf("expected in-progress, got %s", updated.Status)
	}

	// Check status history was recorded
	history, err := svc.GetStatusHistory(context.Background(), enc.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(history))
	}
	if history[0].Status != "planned" {
		t.Errorf("history should record old status 'planned', got %s", history[0].Status)
	}
}

func TestUpdateEncounterStatus_Finished(t *testing.T) {
	svc := newTestService()

	enc := &Encounter{PatientID: uuid.New(), ClassCode: "IMP"}
	svc.CreateEncounter(context.Background(), enc)

	err := svc.UpdateEncounterStatus(context.Background(), enc.ID, "finished")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := svc.GetEncounter(context.Background(), enc.ID)
	if updated.PeriodEnd == nil {
		t.Error("expected period_end to be set when finished")
	}
}

func TestUpdateEncounterStatus_InvalidStatus(t *testing.T) {
	svc := newTestService()

	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	svc.CreateEncounter(context.Background(), enc)

	err := svc.UpdateEncounterStatus(context.Background(), enc.ID, "invalid-status")
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestUpdateEncounterStatus_NotFound(t *testing.T) {
	svc := newTestService()

	err := svc.UpdateEncounterStatus(context.Background(), uuid.New(), "in-progress")
	if err == nil {
		t.Error("expected error for non-existent encounter")
	}
}

func TestListEncountersByPatient(t *testing.T) {
	svc := newTestService()

	patientID := uuid.New()
	otherPatient := uuid.New()

	svc.CreateEncounter(context.Background(), &Encounter{PatientID: patientID, ClassCode: "AMB"})
	svc.CreateEncounter(context.Background(), &Encounter{PatientID: patientID, ClassCode: "IMP"})
	svc.CreateEncounter(context.Background(), &Encounter{PatientID: otherPatient, ClassCode: "AMB"})

	result, total, err := svc.ListEncountersByPatient(context.Background(), patientID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}
}

func TestAddParticipant(t *testing.T) {
	svc := newTestService()

	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	svc.CreateEncounter(context.Background(), enc)

	practID := uuid.New()
	p := &EncounterParticipant{
		EncounterID:    enc.ID,
		PractitionerID: practID,
	}
	err := svc.AddParticipant(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.TypeCode != "ATND" {
		t.Errorf("expected default type_code 'ATND', got %s", p.TypeCode)
	}

	parts, err := svc.GetParticipants(context.Background(), enc.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parts) != 1 {
		t.Fatalf("expected 1 participant, got %d", len(parts))
	}
	if parts[0].PractitionerID != practID {
		t.Error("expected matching practitioner ID")
	}
}

func TestAddParticipant_Validation(t *testing.T) {
	svc := newTestService()

	// Missing encounter_id
	p := &EncounterParticipant{PractitionerID: uuid.New()}
	err := svc.AddParticipant(context.Background(), p)
	if err == nil {
		t.Error("expected error for missing encounter_id")
	}

	// Missing practitioner_id
	p2 := &EncounterParticipant{EncounterID: uuid.New()}
	err = svc.AddParticipant(context.Background(), p2)
	if err == nil {
		t.Error("expected error for missing practitioner_id")
	}
}

func TestAddDiagnosis(t *testing.T) {
	svc := newTestService()

	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	svc.CreateEncounter(context.Background(), enc)

	useCode := "AD"
	rank := 1
	d := &EncounterDiagnosis{
		EncounterID: enc.ID,
		UseCode:     &useCode,
		Rank:        &rank,
	}
	err := svc.AddDiagnosis(context.Background(), d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	diags, err := svc.GetDiagnoses(context.Background(), enc.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnosis, got %d", len(diags))
	}
}

func TestAddDiagnosis_EncounterRequired(t *testing.T) {
	svc := newTestService()

	d := &EncounterDiagnosis{}
	err := svc.AddDiagnosis(context.Background(), d)
	if err == nil {
		t.Error("expected error for missing encounter_id")
	}
}

func TestRemoveParticipant(t *testing.T) {
	svc := newTestService()

	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	svc.CreateEncounter(context.Background(), enc)

	p := &EncounterParticipant{
		EncounterID:    enc.ID,
		PractitionerID: uuid.New(),
	}
	svc.AddParticipant(context.Background(), p)

	err := svc.RemoveParticipant(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parts, _ := svc.GetParticipants(context.Background(), enc.ID)
	if len(parts) != 0 {
		t.Errorf("expected 0 participants after removal, got %d", len(parts))
	}
}

func TestEncounterToFHIR(t *testing.T) {
	patientID := uuid.New()
	practID := uuid.New()
	orgID := uuid.New()
	locID := uuid.New()
	classDisp := "ambulatory"
	typeCode := "consult"
	typeDisp := "Consultation"
	reason := "Annual checkup"
	admitCode := "hosp-trans"
	admitDisp := "Hospital Transfer"
	dischCode := "home"
	dischDisp := "Home"

	enc := &Encounter{
		FHIRID:                   "enc-123",
		Status:                   "in-progress",
		ClassCode:                "AMB",
		ClassDisplay:             &classDisp,
		TypeCode:                 &typeCode,
		TypeDisplay:              &typeDisp,
		PatientID:                patientID,
		PrimaryPractitionerID:    &practID,
		ServiceProviderID:        &orgID,
		LocationID:               &locID,
		PeriodStart:              time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC),
		ReasonText:               &reason,
		AdmitSourceCode:          &admitCode,
		AdmitSourceDisplay:       &admitDisp,
		DischargeDispositionCode: &dischCode,
		DischargeDispositionDisp: &dischDisp,
		ReAdmission:              true,
		UpdatedAt:                time.Now(),
	}

	fhirEnc := enc.ToFHIR()

	if fhirEnc["resourceType"] != "Encounter" {
		t.Errorf("expected Encounter, got %v", fhirEnc["resourceType"])
	}
	if fhirEnc["id"] != "enc-123" {
		t.Errorf("expected enc-123, got %v", fhirEnc["id"])
	}
	if fhirEnc["status"] != "in-progress" {
		t.Errorf("expected in-progress, got %v", fhirEnc["status"])
	}
	if fhirEnc["subject"] == nil {
		t.Error("expected subject")
	}
	if fhirEnc["participant"] == nil {
		t.Error("expected participant")
	}
	if fhirEnc["serviceProvider"] == nil {
		t.Error("expected serviceProvider")
	}
	if fhirEnc["location"] == nil {
		t.Error("expected location")
	}
	if fhirEnc["type"] == nil {
		t.Error("expected type")
	}
	if fhirEnc["reasonCode"] == nil {
		t.Error("expected reasonCode")
	}
	if fhirEnc["hospitalization"] == nil {
		t.Error("expected hospitalization")
	}
}

func TestEncounterToFHIR_Minimal(t *testing.T) {
	enc := &Encounter{
		FHIRID:      "enc-min",
		Status:      "planned",
		ClassCode:   "AMB",
		PatientID:   uuid.New(),
		PeriodStart: time.Now(),
		UpdatedAt:   time.Now(),
	}

	fhirEnc := enc.ToFHIR()

	if fhirEnc["resourceType"] != "Encounter" {
		t.Errorf("expected Encounter, got %v", fhirEnc["resourceType"])
	}
	if fhirEnc["participant"] != nil {
		t.Error("expected no participant for minimal encounter")
	}
	if fhirEnc["hospitalization"] != nil {
		t.Error("expected no hospitalization for minimal encounter")
	}
	if fhirEnc["type"] != nil {
		t.Error("expected no type for minimal encounter")
	}
}

func TestValidStatuses(t *testing.T) {
	expected := []string{"planned", "arrived", "triaged", "in-progress", "onleave", "finished", "cancelled", "entered-in-error"}
	for _, s := range expected {
		if !validStatuses[s] {
			t.Errorf("expected %s to be a valid status", s)
		}
	}
}
