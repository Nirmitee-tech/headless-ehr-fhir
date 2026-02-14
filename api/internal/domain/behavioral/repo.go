package behavioral

import (
	"context"

	"github.com/google/uuid"
)

type PsychAssessmentRepository interface {
	Create(ctx context.Context, a *PsychiatricAssessment) error
	GetByID(ctx context.Context, id uuid.UUID) (*PsychiatricAssessment, error)
	Update(ctx context.Context, a *PsychiatricAssessment) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PsychiatricAssessment, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*PsychiatricAssessment, int, error)
}

type SafetyPlanRepository interface {
	Create(ctx context.Context, s *SafetyPlan) error
	GetByID(ctx context.Context, id uuid.UUID) (*SafetyPlan, error)
	Update(ctx context.Context, s *SafetyPlan) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*SafetyPlan, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SafetyPlan, int, error)
}

type LegalHoldRepository interface {
	Create(ctx context.Context, h *LegalHold) error
	GetByID(ctx context.Context, id uuid.UUID) (*LegalHold, error)
	Update(ctx context.Context, h *LegalHold) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*LegalHold, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*LegalHold, int, error)
}

type SeclusionRestraintRepository interface {
	Create(ctx context.Context, e *SeclusionRestraintEvent) error
	GetByID(ctx context.Context, id uuid.UUID) (*SeclusionRestraintEvent, error)
	Update(ctx context.Context, e *SeclusionRestraintEvent) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*SeclusionRestraintEvent, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SeclusionRestraintEvent, int, error)
}

type GroupTherapyRepository interface {
	Create(ctx context.Context, s *GroupTherapySession) error
	GetByID(ctx context.Context, id uuid.UUID) (*GroupTherapySession, error)
	Update(ctx context.Context, s *GroupTherapySession) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*GroupTherapySession, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*GroupTherapySession, int, error)
	// Attendance
	AddAttendance(ctx context.Context, a *GroupTherapyAttendance) error
	GetAttendance(ctx context.Context, sessionID uuid.UUID) ([]*GroupTherapyAttendance, error)
}
