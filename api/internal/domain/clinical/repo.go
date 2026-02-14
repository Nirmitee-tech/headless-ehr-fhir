package clinical

import (
	"context"

	"github.com/google/uuid"
)

type ConditionRepository interface {
	Create(ctx context.Context, c *Condition) error
	GetByID(ctx context.Context, id uuid.UUID) (*Condition, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Condition, error)
	Update(ctx context.Context, c *Condition) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Condition, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Condition, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Condition, int, error)
}

type ObservationRepository interface {
	Create(ctx context.Context, o *Observation) error
	GetByID(ctx context.Context, id uuid.UUID) (*Observation, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Observation, error)
	Update(ctx context.Context, o *Observation) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Observation, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Observation, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Observation, int, error)
	// Components
	AddComponent(ctx context.Context, c *ObservationComponent) error
	GetComponents(ctx context.Context, observationID uuid.UUID) ([]*ObservationComponent, error)
}

type AllergyRepository interface {
	Create(ctx context.Context, a *AllergyIntolerance) error
	GetByID(ctx context.Context, id uuid.UUID) (*AllergyIntolerance, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*AllergyIntolerance, error)
	Update(ctx context.Context, a *AllergyIntolerance) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*AllergyIntolerance, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*AllergyIntolerance, int, error)
	// Reactions
	AddReaction(ctx context.Context, r *AllergyReaction) error
	GetReactions(ctx context.Context, allergyID uuid.UUID) ([]*AllergyReaction, error)
	RemoveReaction(ctx context.Context, id uuid.UUID) error
}

type ProcedureRepository interface {
	Create(ctx context.Context, p *ProcedureRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*ProcedureRecord, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ProcedureRecord, error)
	Update(ctx context.Context, p *ProcedureRecord) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ProcedureRecord, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ProcedureRecord, int, error)
	// Performers
	AddPerformer(ctx context.Context, pf *ProcedurePerformer) error
	GetPerformers(ctx context.Context, procedureID uuid.UUID) ([]*ProcedurePerformer, error)
	RemovePerformer(ctx context.Context, id uuid.UUID) error
}
