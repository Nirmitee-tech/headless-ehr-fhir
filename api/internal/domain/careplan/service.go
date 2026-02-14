package careplan

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	carePlans  CarePlanRepository
	goals      GoalRepository
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
	return s.carePlans.Create(ctx, cp)
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
	return s.carePlans.Update(ctx, cp)
}

func (s *Service) DeleteCarePlan(ctx context.Context, id uuid.UUID) error {
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
	return s.goals.Create(ctx, g)
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
	return s.goals.Update(ctx, g)
}

func (s *Service) DeleteGoal(ctx context.Context, id uuid.UUID) error {
	return s.goals.Delete(ctx, id)
}

func (s *Service) ListGoalsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Goal, int, error) {
	return s.goals.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchGoals(ctx context.Context, params map[string]string, limit, offset int) ([]*Goal, int, error) {
	return s.goals.Search(ctx, params, limit, offset)
}
