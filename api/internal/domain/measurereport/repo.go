package measurereport

import (
	"context"

	"github.com/google/uuid"
)

type MeasureReportRepository interface {
	Create(ctx context.Context, mr *MeasureReport) error
	GetByID(ctx context.Context, id uuid.UUID) (*MeasureReport, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MeasureReport, error)
	Update(ctx context.Context, mr *MeasureReport) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*MeasureReport, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MeasureReport, int, error)
}
