package eventdefinition

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo EventDefinitionRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo EventDefinitionRepository) *Service {
	return &Service{repo: repo}
}

var validEventDefinitionStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

func (s *Service) CreateEventDefinition(ctx context.Context, e *EventDefinition) error {
	if e.Status == "" {
		e.Status = "draft"
	}
	if !validEventDefinitionStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return err
	}
	e.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "EventDefinition", e.FHIRID, e.ToFHIR())
	}
	return nil
}

func (s *Service) GetEventDefinition(ctx context.Context, id uuid.UUID) (*EventDefinition, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetEventDefinitionByFHIRID(ctx context.Context, fhirID string) (*EventDefinition, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateEventDefinition(ctx context.Context, e *EventDefinition) error {
	if e.Status != "" && !validEventDefinitionStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "EventDefinition", e.FHIRID, e.VersionID, e.ToFHIR())
		if err == nil {
			e.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, e)
}

func (s *Service) DeleteEventDefinition(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		e, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "EventDefinition", e.FHIRID, e.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchEventDefinitions(ctx context.Context, params map[string]string, limit, offset int) ([]*EventDefinition, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
