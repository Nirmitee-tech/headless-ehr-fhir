package structuredefinition

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo StructureDefinitionRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo StructureDefinitionRepository) *Service {
	return &Service{repo: repo}
}

var validStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

var validKinds = map[string]bool{
	"primitive-type": true, "complex-type": true, "resource": true, "logical": true,
}

func (s *Service) CreateStructureDefinition(ctx context.Context, sd *StructureDefinition) error {
	if sd.URL == "" {
		return fmt.Errorf("url is required")
	}
	if sd.Name == "" {
		return fmt.Errorf("name is required")
	}
	if sd.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if sd.Type == "" {
		return fmt.Errorf("type is required")
	}
	if sd.Status == "" {
		sd.Status = "draft"
	}
	if !validStatuses[sd.Status] {
		return fmt.Errorf("invalid status: %s", sd.Status)
	}
	if !validKinds[sd.Kind] {
		return fmt.Errorf("invalid kind: %s", sd.Kind)
	}
	if err := s.repo.Create(ctx, sd); err != nil {
		return err
	}
	sd.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "StructureDefinition", sd.FHIRID, sd.ToFHIR())
	}
	return nil
}

func (s *Service) GetStructureDefinition(ctx context.Context, id uuid.UUID) (*StructureDefinition, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetStructureDefinitionByFHIRID(ctx context.Context, fhirID string) (*StructureDefinition, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateStructureDefinition(ctx context.Context, sd *StructureDefinition) error {
	if sd.Status != "" && !validStatuses[sd.Status] {
		return fmt.Errorf("invalid status: %s", sd.Status)
	}
	if sd.Kind != "" && !validKinds[sd.Kind] {
		return fmt.Errorf("invalid kind: %s", sd.Kind)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "StructureDefinition", sd.FHIRID, sd.VersionID, sd.ToFHIR())
		if err == nil {
			sd.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, sd)
}

func (s *Service) DeleteStructureDefinition(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		sd, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "StructureDefinition", sd.FHIRID, sd.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchStructureDefinitions(ctx context.Context, params map[string]string, limit, offset int) ([]*StructureDefinition, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
