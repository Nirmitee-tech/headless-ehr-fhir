package encounter

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Valid encounter statuses per FHIR R4.
var validStatuses = map[string]bool{
	"planned":          true,
	"arrived":          true,
	"triaged":          true,
	"in-progress":      true,
	"onleave":          true,
	"finished":         true,
	"cancelled":        true,
	"entered-in-error": true,
}

func (s *Service) CreateEncounter(ctx context.Context, enc *Encounter) error {
	if enc.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if enc.ClassCode == "" {
		return fmt.Errorf("class_code is required")
	}
	if enc.Status == "" {
		enc.Status = "planned"
	}
	if !validStatuses[enc.Status] {
		return fmt.Errorf("invalid status: %s", enc.Status)
	}
	if enc.PeriodStart.IsZero() {
		enc.PeriodStart = time.Now().UTC()
	}
	return s.repo.Create(ctx, enc)
}

func (s *Service) GetEncounter(ctx context.Context, id uuid.UUID) (*Encounter, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetEncounterByFHIRID(ctx context.Context, fhirID string) (*Encounter, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateEncounter(ctx context.Context, enc *Encounter) error {
	if enc.Status != "" && !validStatuses[enc.Status] {
		return fmt.Errorf("invalid status: %s", enc.Status)
	}
	return s.repo.Update(ctx, enc)
}

func (s *Service) UpdateEncounterStatus(ctx context.Context, id uuid.UUID, newStatus string) error {
	if !validStatuses[newStatus] {
		return fmt.Errorf("invalid status: %s", newStatus)
	}

	enc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("encounter not found: %w", err)
	}

	oldStatus := enc.Status
	now := time.Now().UTC()

	// Record status history
	history := &EncounterStatusHistory{
		EncounterID: id,
		Status:      oldStatus,
		PeriodStart: enc.PeriodStart,
		PeriodEnd:   &now,
	}
	if err := s.repo.AddStatusHistory(ctx, history); err != nil {
		return fmt.Errorf("add status history: %w", err)
	}

	enc.Status = newStatus
	if newStatus == "finished" && enc.PeriodEnd == nil {
		enc.PeriodEnd = &now
	}

	return s.repo.Update(ctx, enc)
}

func (s *Service) DeleteEncounter(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) ListEncounters(ctx context.Context, limit, offset int) ([]*Encounter, int, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *Service) ListEncountersByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Encounter, int, error) {
	return s.repo.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchEncounters(ctx context.Context, params map[string]string, limit, offset int) ([]*Encounter, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}

func (s *Service) AddParticipant(ctx context.Context, p *EncounterParticipant) error {
	if p.EncounterID == uuid.Nil {
		return fmt.Errorf("encounter_id is required")
	}
	if p.PractitionerID == uuid.Nil {
		return fmt.Errorf("practitioner_id is required")
	}
	if p.TypeCode == "" {
		p.TypeCode = "ATND"
	}
	return s.repo.AddParticipant(ctx, p)
}

func (s *Service) GetParticipants(ctx context.Context, encounterID uuid.UUID) ([]*EncounterParticipant, error) {
	return s.repo.GetParticipants(ctx, encounterID)
}

func (s *Service) RemoveParticipant(ctx context.Context, id uuid.UUID) error {
	return s.repo.RemoveParticipant(ctx, id)
}

func (s *Service) AddDiagnosis(ctx context.Context, d *EncounterDiagnosis) error {
	if d.EncounterID == uuid.Nil {
		return fmt.Errorf("encounter_id is required")
	}
	return s.repo.AddDiagnosis(ctx, d)
}

func (s *Service) GetDiagnoses(ctx context.Context, encounterID uuid.UUID) ([]*EncounterDiagnosis, error) {
	return s.repo.GetDiagnoses(ctx, encounterID)
}

func (s *Service) RemoveDiagnosis(ctx context.Context, id uuid.UUID) error {
	return s.repo.RemoveDiagnosis(ctx, id)
}

func (s *Service) GetStatusHistory(ctx context.Context, encounterID uuid.UUID) ([]*EncounterStatusHistory, error) {
	return s.repo.GetStatusHistory(ctx, encounterID)
}
