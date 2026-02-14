package immunization

import (
	"context"

	"github.com/google/uuid"
)

type ImmunizationRepository interface {
	Create(ctx context.Context, im *Immunization) error
	GetByID(ctx context.Context, id uuid.UUID) (*Immunization, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Immunization, error)
	Update(ctx context.Context, im *Immunization) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Immunization, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Immunization, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Immunization, int, error)
}

type RecommendationRepository interface {
	Create(ctx context.Context, r *ImmunizationRecommendation) error
	GetByID(ctx context.Context, id uuid.UUID) (*ImmunizationRecommendation, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ImmunizationRecommendation, error)
	Update(ctx context.Context, r *ImmunizationRecommendation) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*ImmunizationRecommendation, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ImmunizationRecommendation, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ImmunizationRecommendation, int, error)
}
