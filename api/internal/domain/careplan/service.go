package careplan

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	carePlans  CarePlanRepository
	goals      GoalRepository
	vt         *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(cp CarePlanRepository, g GoalRepository) *Service {
	return &Service{carePlans: cp, goals: g}
}

// -- CarePlan --

var validCPStatuses = map[string]bool{
	"draft": true, "active": true, "on-hold": true,
	"completed": true, "revoked": true, "entered-in-error": true,
}

var validCPIntents = map[string]bool{
	"proposal": true, "plan": true, "order": true, "option": true,
}

func (s *Service) CreateCarePlan(ctx context.Context, cp *CarePlan) error {
	if cp.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if cp.Intent == "" {
		return fmt.Errorf("intent is required")
	}
	if !validCPIntents[cp.Intent] {
		return fmt.Errorf("invalid intent: %s", cp.Intent)
	}
	if cp.Status == "" {
		cp.Status = "draft"
	}
	if !validCPStatuses[cp.Status] {
		return fmt.Errorf("invalid status: %s", cp.Status)
	}
	if err := s.carePlans.Create(ctx, cp); err != nil {
		return err
	}
	cp.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "CarePlan", cp.FHIRID, cp.ToFHIR())
	}
	return nil
}

func (s *Service) GetCarePlan(ctx context.Context, id uuid.UUID) (*CarePlan, error) {
	return s.carePlans.GetByID(ctx, id)
}

func (s *Service) GetCarePlanByFHIRID(ctx context.Context, fhirID string) (*CarePlan, error) {
	return s.carePlans.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateCarePlan(ctx context.Context, cp *CarePlan) error {
	if cp.Status != "" && !validCPStatuses[cp.Status] {
		return fmt.Errorf("invalid status: %s", cp.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "CarePlan", cp.FHIRID, cp.VersionID, cp.ToFHIR())
		if err == nil {
			cp.VersionID = newVer
		}
	}
	return s.carePlans.Update(ctx, cp)
}

func (s *Service) DeleteCarePlan(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		cp, err := s.carePlans.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "CarePlan", cp.FHIRID, cp.VersionID)
		}
	}
	return s.carePlans.Delete(ctx, id)
}

func (s *Service) ListCarePlansByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*CarePlan, int, error) {
	return s.carePlans.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchCarePlans(ctx context.Context, params map[string]string, limit, offset int) ([]*CarePlan, int, error) {
	return s.carePlans.Search(ctx, params, limit, offset)
}

func (s *Service) AddActivity(ctx context.Context, a *CarePlanActivity) error {
	if a.CarePlanID == uuid.Nil {
		return fmt.Errorf("care_plan_id is required")
	}
	if a.Status == "" {
		return fmt.Errorf("status is required")
	}
	return s.carePlans.AddActivity(ctx, a)
}

func (s *Service) GetActivities(ctx context.Context, carePlanID uuid.UUID) ([]*CarePlanActivity, error) {
	return s.carePlans.GetActivities(ctx, carePlanID)
}

// -- Goal --

var validGoalStatuses = map[string]bool{
	"proposed": true, "planned": true, "accepted": true, "active": true,
	"on-hold": true, "completed": true, "cancelled": true,
	"entered-in-error": true, "rejected": true,
}

func (s *Service) CreateGoal(ctx context.Context, g *Goal) error {
	if g.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if g.Description == "" {
		return fmt.Errorf("description is required")
	}
	if g.LifecycleStatus == "" {
		g.LifecycleStatus = "proposed"
	}
	if !validGoalStatuses[g.LifecycleStatus] {
		return fmt.Errorf("invalid lifecycle_status: %s", g.LifecycleStatus)
	}
	if err := s.goals.Create(ctx, g); err != nil {
		return err
	}
	g.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Goal", g.FHIRID, g.ToFHIR())
	}
	return nil
}

func (s *Service) GetGoal(ctx context.Context, id uuid.UUID) (*Goal, error) {
	return s.goals.GetByID(ctx, id)
}

func (s *Service) GetGoalByFHIRID(ctx context.Context, fhirID string) (*Goal, error) {
	return s.goals.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateGoal(ctx context.Context, g *Goal) error {
	if g.LifecycleStatus != "" && !validGoalStatuses[g.LifecycleStatus] {
		return fmt.Errorf("invalid lifecycle_status: %s", g.LifecycleStatus)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Goal", g.FHIRID, g.VersionID, g.ToFHIR())
		if err == nil {
			g.VersionID = newVer
		}
	}
	return s.goals.Update(ctx, g)
}

func (s *Service) DeleteGoal(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		g, err := s.goals.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Goal", g.FHIRID, g.VersionID)
		}
	}
	return s.goals.Delete(ctx, id)
}

func (s *Service) ListGoalsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Goal, int, error) {
	return s.goals.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchGoals(ctx context.Context, params map[string]string, limit, offset int) ([]*Goal, int, error) {
	return s.goals.Search(ctx, params, limit, offset)
}
