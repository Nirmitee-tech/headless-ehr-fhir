package behavioral

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	assessments PsychAssessmentRepository
	safetyPlans SafetyPlanRepository
	legalHolds  LegalHoldRepository
	seclusions  SeclusionRestraintRepository
	groups      GroupTherapyRepository
	vt          *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(
	assessments PsychAssessmentRepository,
	safetyPlans SafetyPlanRepository,
	legalHolds LegalHoldRepository,
	seclusions SeclusionRestraintRepository,
	groups GroupTherapyRepository,
) *Service {
	return &Service{
		assessments: assessments,
		safetyPlans: safetyPlans,
		legalHolds:  legalHolds,
		seclusions:  seclusions,
		groups:      groups,
	}
}

// -- Psychiatric Assessment --

func (s *Service) CreatePsychAssessment(ctx context.Context, a *PsychiatricAssessment) error {
	if a.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if a.EncounterID == uuid.Nil {
		return fmt.Errorf("encounter_id is required")
	}
	if a.AssessorID == uuid.Nil {
		return fmt.Errorf("assessor_id is required")
	}
	return s.assessments.Create(ctx, a)
}

func (s *Service) GetPsychAssessment(ctx context.Context, id uuid.UUID) (*PsychiatricAssessment, error) {
	return s.assessments.GetByID(ctx, id)
}

func (s *Service) UpdatePsychAssessment(ctx context.Context, a *PsychiatricAssessment) error {
	return s.assessments.Update(ctx, a)
}

func (s *Service) DeletePsychAssessment(ctx context.Context, id uuid.UUID) error {
	return s.assessments.Delete(ctx, id)
}

func (s *Service) ListPsychAssessmentsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PsychiatricAssessment, int, error) {
	return s.assessments.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchPsychAssessments(ctx context.Context, params map[string]string, limit, offset int) ([]*PsychiatricAssessment, int, error) {
	return s.assessments.Search(ctx, params, limit, offset)
}

// -- Safety Plan --

var validSafetyPlanStatuses = map[string]bool{
	"active": true, "superseded": true, "entered-in-error": true,
}

func (s *Service) CreateSafetyPlan(ctx context.Context, sp *SafetyPlan) error {
	if sp.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if sp.CreatedByID == uuid.Nil {
		return fmt.Errorf("created_by_id is required")
	}
	if sp.Status == "" {
		sp.Status = "active"
	}
	if !validSafetyPlanStatuses[sp.Status] {
		return fmt.Errorf("invalid status: %s", sp.Status)
	}
	return s.safetyPlans.Create(ctx, sp)
}

func (s *Service) GetSafetyPlan(ctx context.Context, id uuid.UUID) (*SafetyPlan, error) {
	return s.safetyPlans.GetByID(ctx, id)
}

func (s *Service) UpdateSafetyPlan(ctx context.Context, sp *SafetyPlan) error {
	if sp.Status != "" && !validSafetyPlanStatuses[sp.Status] {
		return fmt.Errorf("invalid status: %s", sp.Status)
	}
	return s.safetyPlans.Update(ctx, sp)
}

func (s *Service) DeleteSafetyPlan(ctx context.Context, id uuid.UUID) error {
	return s.safetyPlans.Delete(ctx, id)
}

func (s *Service) ListSafetyPlansByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*SafetyPlan, int, error) {
	return s.safetyPlans.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchSafetyPlans(ctx context.Context, params map[string]string, limit, offset int) ([]*SafetyPlan, int, error) {
	return s.safetyPlans.Search(ctx, params, limit, offset)
}

// -- Legal Hold --

var validLegalHoldStatuses = map[string]bool{
	"active": true, "expired": true, "converted": true, "released": true, "rescinded": true,
}

func (s *Service) CreateLegalHold(ctx context.Context, h *LegalHold) error {
	if h.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if h.InitiatedByID == uuid.Nil {
		return fmt.Errorf("initiated_by_id is required")
	}
	if h.HoldType == "" {
		return fmt.Errorf("hold_type is required")
	}
	if h.Reason == "" {
		return fmt.Errorf("reason is required")
	}
	if h.Status == "" {
		h.Status = "active"
	}
	if !validLegalHoldStatuses[h.Status] {
		return fmt.Errorf("invalid status: %s", h.Status)
	}
	return s.legalHolds.Create(ctx, h)
}

func (s *Service) GetLegalHold(ctx context.Context, id uuid.UUID) (*LegalHold, error) {
	return s.legalHolds.GetByID(ctx, id)
}

func (s *Service) UpdateLegalHold(ctx context.Context, h *LegalHold) error {
	if h.Status != "" && !validLegalHoldStatuses[h.Status] {
		return fmt.Errorf("invalid status: %s", h.Status)
	}
	return s.legalHolds.Update(ctx, h)
}

func (s *Service) DeleteLegalHold(ctx context.Context, id uuid.UUID) error {
	return s.legalHolds.Delete(ctx, id)
}

func (s *Service) ListLegalHoldsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*LegalHold, int, error) {
	return s.legalHolds.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchLegalHolds(ctx context.Context, params map[string]string, limit, offset int) ([]*LegalHold, int, error) {
	return s.legalHolds.Search(ctx, params, limit, offset)
}

// -- Seclusion / Restraint --

func (s *Service) CreateSeclusionRestraint(ctx context.Context, e *SeclusionRestraintEvent) error {
	if e.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if e.OrderedByID == uuid.Nil {
		return fmt.Errorf("ordered_by_id is required")
	}
	if e.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if e.Reason == "" {
		return fmt.Errorf("reason is required")
	}
	return s.seclusions.Create(ctx, e)
}

func (s *Service) GetSeclusionRestraint(ctx context.Context, id uuid.UUID) (*SeclusionRestraintEvent, error) {
	return s.seclusions.GetByID(ctx, id)
}

func (s *Service) UpdateSeclusionRestraint(ctx context.Context, e *SeclusionRestraintEvent) error {
	return s.seclusions.Update(ctx, e)
}

func (s *Service) DeleteSeclusionRestraint(ctx context.Context, id uuid.UUID) error {
	return s.seclusions.Delete(ctx, id)
}

func (s *Service) ListSeclusionRestraintsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*SeclusionRestraintEvent, int, error) {
	return s.seclusions.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchSeclusionRestraints(ctx context.Context, params map[string]string, limit, offset int) ([]*SeclusionRestraintEvent, int, error) {
	return s.seclusions.Search(ctx, params, limit, offset)
}

// -- Group Therapy --

var validGroupTherapyStatuses = map[string]bool{
	"scheduled": true, "completed": true, "cancelled": true,
}

func (s *Service) CreateGroupTherapySession(ctx context.Context, gs *GroupTherapySession) error {
	if gs.SessionName == "" {
		return fmt.Errorf("session_name is required")
	}
	if gs.FacilitatorID == uuid.Nil {
		return fmt.Errorf("facilitator_id is required")
	}
	if gs.Status == "" {
		gs.Status = "scheduled"
	}
	if !validGroupTherapyStatuses[gs.Status] {
		return fmt.Errorf("invalid status: %s", gs.Status)
	}
	return s.groups.Create(ctx, gs)
}

func (s *Service) GetGroupTherapySession(ctx context.Context, id uuid.UUID) (*GroupTherapySession, error) {
	return s.groups.GetByID(ctx, id)
}

func (s *Service) UpdateGroupTherapySession(ctx context.Context, gs *GroupTherapySession) error {
	if gs.Status != "" && !validGroupTherapyStatuses[gs.Status] {
		return fmt.Errorf("invalid status: %s", gs.Status)
	}
	return s.groups.Update(ctx, gs)
}

func (s *Service) DeleteGroupTherapySession(ctx context.Context, id uuid.UUID) error {
	return s.groups.Delete(ctx, id)
}

func (s *Service) ListGroupTherapySessions(ctx context.Context, limit, offset int) ([]*GroupTherapySession, int, error) {
	return s.groups.List(ctx, limit, offset)
}

func (s *Service) SearchGroupTherapySessions(ctx context.Context, params map[string]string, limit, offset int) ([]*GroupTherapySession, int, error) {
	return s.groups.Search(ctx, params, limit, offset)
}

func (s *Service) AddGroupTherapyAttendance(ctx context.Context, a *GroupTherapyAttendance) error {
	if a.SessionID == uuid.Nil {
		return fmt.Errorf("session_id is required")
	}
	if a.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	return s.groups.AddAttendance(ctx, a)
}

func (s *Service) GetGroupTherapyAttendance(ctx context.Context, sessionID uuid.UUID) ([]*GroupTherapyAttendance, error) {
	return s.groups.GetAttendance(ctx, sessionID)
}
