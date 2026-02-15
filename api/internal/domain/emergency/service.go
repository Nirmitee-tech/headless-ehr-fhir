package emergency

import (
	"context"
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	triage   TriageRepository
	tracking EDTrackingRepository
	trauma   TraumaRepository
	vt       *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(triage TriageRepository, tracking EDTrackingRepository, trauma TraumaRepository) *Service {
	return &Service{triage: triage, tracking: tracking, trauma: trauma}
}

// -- Triage Record --

func (s *Service) CreateTriageRecord(ctx context.Context, t *TriageRecord) error {
	if t.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if t.EncounterID == uuid.Nil {
		return fmt.Errorf("encounter_id is required")
	}
	if t.TriageNurseID == uuid.Nil {
		return fmt.Errorf("triage_nurse_id is required")
	}
	if t.ChiefComplaint == "" {
		return fmt.Errorf("chief_complaint is required")
	}
	if t.TriageTime == nil {
		now := time.Now()
		t.TriageTime = &now
	}
	return s.triage.Create(ctx, t)
}

func (s *Service) GetTriageRecord(ctx context.Context, id uuid.UUID) (*TriageRecord, error) {
	return s.triage.GetByID(ctx, id)
}

func (s *Service) UpdateTriageRecord(ctx context.Context, t *TriageRecord) error {
	return s.triage.Update(ctx, t)
}

func (s *Service) DeleteTriageRecord(ctx context.Context, id uuid.UUID) error {
	return s.triage.Delete(ctx, id)
}

func (s *Service) ListTriageRecords(ctx context.Context, limit, offset int) ([]*TriageRecord, int, error) {
	return s.triage.List(ctx, limit, offset)
}

func (s *Service) ListTriageRecordsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*TriageRecord, int, error) {
	return s.triage.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchTriageRecords(ctx context.Context, params map[string]string, limit, offset int) ([]*TriageRecord, int, error) {
	return s.triage.Search(ctx, params, limit, offset)
}

// -- ED Tracking --

func (s *Service) CreateEDTracking(ctx context.Context, t *EDTracking) error {
	if t.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if t.EncounterID == uuid.Nil {
		return fmt.Errorf("encounter_id is required")
	}
	if t.CurrentStatus == "" {
		t.CurrentStatus = "waiting"
	}
	return s.tracking.Create(ctx, t)
}

func (s *Service) GetEDTracking(ctx context.Context, id uuid.UUID) (*EDTracking, error) {
	return s.tracking.GetByID(ctx, id)
}

func (s *Service) UpdateEDTracking(ctx context.Context, t *EDTracking) error {
	return s.tracking.Update(ctx, t)
}

func (s *Service) DeleteEDTracking(ctx context.Context, id uuid.UUID) error {
	return s.tracking.Delete(ctx, id)
}

func (s *Service) ListEDTrackings(ctx context.Context, limit, offset int) ([]*EDTracking, int, error) {
	return s.tracking.List(ctx, limit, offset)
}

func (s *Service) ListEDTrackingsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*EDTracking, int, error) {
	return s.tracking.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchEDTrackings(ctx context.Context, params map[string]string, limit, offset int) ([]*EDTracking, int, error) {
	return s.tracking.Search(ctx, params, limit, offset)
}

func (s *Service) AddEDStatusHistory(ctx context.Context, h *EDStatusHistory) error {
	if h.EDTrackingID == uuid.Nil {
		return fmt.Errorf("ed_tracking_id is required")
	}
	if h.Status == "" {
		return fmt.Errorf("status is required")
	}
	if h.ChangedAt.IsZero() {
		h.ChangedAt = time.Now()
	}
	return s.tracking.AddStatusHistory(ctx, h)
}

func (s *Service) GetEDStatusHistory(ctx context.Context, trackingID uuid.UUID) ([]*EDStatusHistory, error) {
	return s.tracking.GetStatusHistory(ctx, trackingID)
}

// -- Trauma Activation --

func (s *Service) CreateTraumaActivation(ctx context.Context, t *TraumaActivation) error {
	if t.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if t.ActivationLevel == "" {
		return fmt.Errorf("activation_level is required")
	}
	if t.ActivationTime.IsZero() {
		t.ActivationTime = time.Now()
	}
	return s.trauma.Create(ctx, t)
}

func (s *Service) GetTraumaActivation(ctx context.Context, id uuid.UUID) (*TraumaActivation, error) {
	return s.trauma.GetByID(ctx, id)
}

func (s *Service) UpdateTraumaActivation(ctx context.Context, t *TraumaActivation) error {
	return s.trauma.Update(ctx, t)
}

func (s *Service) DeleteTraumaActivation(ctx context.Context, id uuid.UUID) error {
	return s.trauma.Delete(ctx, id)
}

func (s *Service) ListTraumaActivations(ctx context.Context, limit, offset int) ([]*TraumaActivation, int, error) {
	return s.trauma.List(ctx, limit, offset)
}

func (s *Service) ListTraumaActivationsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*TraumaActivation, int, error) {
	return s.trauma.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchTraumaActivations(ctx context.Context, params map[string]string, limit, offset int) ([]*TraumaActivation, int, error) {
	return s.trauma.Search(ctx, params, limit, offset)
}
