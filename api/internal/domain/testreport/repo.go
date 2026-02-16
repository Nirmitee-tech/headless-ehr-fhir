package testreport

import (
	"context"

	"github.com/google/uuid"
)

type TestReportRepository interface {
	Create(ctx context.Context, e *TestReport) error
	GetByID(ctx context.Context, id uuid.UUID) (*TestReport, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*TestReport, error)
	Update(ctx context.Context, e *TestReport) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*TestReport, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*TestReport, int, error)
}
