package specimendefinition

import (
	"context"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo SpecimenDefinitionRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo SpecimenDefinitionRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateSpecimenDefinition(ctx context.Context, sd *SpecimenDefinition) error {
	if err := s.repo.Create(ctx, sd); err != nil {
		return err
	}
	sd.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "SpecimenDefinition", sd.FHIRID, sd.ToFHIR())
	}
	return nil
}

func (s *Service) GetSpecimenDefinition(ctx context.Context, id uuid.UUID) (*SpecimenDefinition, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetSpecimenDefinitionByFHIRID(ctx context.Context, fhirID string) (*SpecimenDefinition, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateSpecimenDefinition(ctx context.Context, sd *SpecimenDefinition) error {
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "SpecimenDefinition", sd.FHIRID, sd.VersionID, sd.ToFHIR())
		if err == nil {
			sd.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, sd)
}

func (s *Service) DeleteSpecimenDefinition(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		sd, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "SpecimenDefinition", sd.FHIRID, sd.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchSpecimenDefinitions(ctx context.Context, params map[string]string, limit, offset int) ([]*SpecimenDefinition, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
