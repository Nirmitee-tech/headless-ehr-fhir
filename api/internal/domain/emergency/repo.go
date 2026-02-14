package emergency

import (
	"context"

	"github.com/google/uuid"
)

type TriageRepository interface {
	Create(ctx context.Context, t *TriageRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*TriageRecord, error)
	Update(ctx context.Context, t *TriageRecord) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*TriageRecord, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*TriageRecord, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*TriageRecord, int, error)
}

type EDTrackingRepository interface {
	Create(ctx context.Context, t *EDTracking) error
	GetByID(ctx context.Context, id uuid.UUID) (*EDTracking, error)
	Update(ctx context.Context, t *EDTracking) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*EDTracking, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*EDTracking, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*EDTracking, int, error)
	// Status History
	AddStatusHistory(ctx context.Context, h *EDStatusHistory) error
	GetStatusHistory(ctx context.Context, trackingID uuid.UUID) ([]*EDStatusHistory, error)
}

type TraumaRepository interface {
	Create(ctx context.Context, t *TraumaActivation) error
	GetByID(ctx context.Context, id uuid.UUID) (*TraumaActivation, error)
	Update(ctx context.Context, t *TraumaActivation) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*TraumaActivation, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*TraumaActivation, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*TraumaActivation, int, error)
}
