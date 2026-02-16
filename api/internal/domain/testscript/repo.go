package testscript

import (
	"context"

	"github.com/google/uuid"
)

type TestScriptRepository interface {
	Create(ctx context.Context, ts *TestScript) error
	GetByID(ctx context.Context, id uuid.UUID) (*TestScript, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*TestScript, error)
	Update(ctx context.Context, ts *TestScript) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*TestScript, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*TestScript, int, error)
}
