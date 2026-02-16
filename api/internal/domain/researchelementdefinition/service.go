package researchelementdefinition

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo ResearchElementDefinitionRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo ResearchElementDefinitionRepository) *Service {
	return &Service{repo: repo}
}

var validStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

var validTypes = map[string]bool{
	"population": true, "exposure": true, "outcome": true,
}

func (s *Service) CreateResearchElementDefinition(ctx context.Context, e *ResearchElementDefinition) error {
	if e.Status == "" {
		e.Status = "draft"
	}
	if !validStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	if e.Type == "" {
		return fmt.Errorf("type is required")
	}
	if !validTypes[e.Type] {
		return fmt.Errorf("invalid type: %s", e.Type)
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return err
	}
	e.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ResearchElementDefinition", e.FHIRID, e.ToFHIR())
	}
	return nil
}

func (s *Service) GetResearchElementDefinition(ctx context.Context, id uuid.UUID) (*ResearchElementDefinition, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetResearchElementDefinitionByFHIRID(ctx context.Context, fhirID string) (*ResearchElementDefinition, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateResearchElementDefinition(ctx context.Context, e *ResearchElementDefinition) error {
	if e.Status != "" && !validStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ResearchElementDefinition", e.FHIRID, e.VersionID, e.ToFHIR())
		if err == nil {
			e.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, e)
}

func (s *Service) DeleteResearchElementDefinition(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		e, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ResearchElementDefinition", e.FHIRID, e.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchResearchElementDefinitions(ctx context.Context, params map[string]string, limit, offset int) ([]*ResearchElementDefinition, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
