package medproductingredient

import (
	"context"
	"github.com/google/uuid"
)

type MedicinalProductIngredientRepository interface {
	Create(ctx context.Context, m *MedicinalProductIngredient) error
	GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductIngredient, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductIngredient, error)
	Update(ctx context.Context, m *MedicinalProductIngredient) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*MedicinalProductIngredient, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductIngredient, int, error)
}
