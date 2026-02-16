package auditevent

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo AuditEventRepository
}

func NewService(repo AuditEventRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAuditEvent(ctx context.Context, id uuid.UUID) (*AuditEvent, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetAuditEventByFHIRID(ctx context.Context, fhirID string) (*AuditEvent, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) SearchAuditEvents(ctx context.Context, params map[string]string, limit, offset int) ([]*AuditEvent, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
