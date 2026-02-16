package graphdefinition

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo GraphDefinitionRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo GraphDefinitionRepository) *Service {
	return &Service{repo: repo}
}

var validGraphDefinitionStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

func (s *Service) CreateGraphDefinition(ctx context.Context, g *GraphDefinition) error {
	if g.Name == "" {
		return fmt.Errorf("name is required")
	}
	if g.StartType == "" {
		return fmt.Errorf("start type is required")
	}
	if g.Status == "" {
		g.Status = "draft"
	}
	if !validGraphDefinitionStatuses[g.Status] {
		return fmt.Errorf("invalid status: %s", g.Status)
	}
	if err := s.repo.Create(ctx, g); err != nil {
		return err
	}
	g.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "GraphDefinition", g.FHIRID, g.ToFHIR())
	}
	return nil
}

func (s *Service) GetGraphDefinition(ctx context.Context, id uuid.UUID) (*GraphDefinition, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetGraphDefinitionByFHIRID(ctx context.Context, fhirID string) (*GraphDefinition, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateGraphDefinition(ctx context.Context, g *GraphDefinition) error {
	if g.Status != "" && !validGraphDefinitionStatuses[g.Status] {
		return fmt.Errorf("invalid status: %s", g.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "GraphDefinition", g.FHIRID, g.VersionID, g.ToFHIR())
		if err == nil {
			g.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, g)
}

func (s *Service) DeleteGraphDefinition(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		g, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "GraphDefinition", g.FHIRID, g.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchGraphDefinitions(ctx context.Context, params map[string]string, limit, offset int) ([]*GraphDefinition, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
