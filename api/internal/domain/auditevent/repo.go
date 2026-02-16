package auditevent

import (
	"context"

	"github.com/google/uuid"
)

type AuditEventRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*AuditEvent, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*AuditEvent, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*AuditEvent, int, error)
}
