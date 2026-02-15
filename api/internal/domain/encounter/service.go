package encounter

import (
	"context"
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo Repository
	vt   *fhir.VersionTracker
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
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
	if err := s.repo.Create(ctx, enc); err != nil {
		return err
	}
	enc.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Encounter", enc.FHIRID, enc.ToFHIR())
	}
	return nil
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
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Encounter", enc.FHIRID, enc.VersionID, enc.ToFHIR())
		if err == nil {
			enc.VersionID = newVer
		}
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
	if s.vt != nil {
		enc, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Encounter", enc.FHIRID, enc.VersionID)
		}
	}
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
