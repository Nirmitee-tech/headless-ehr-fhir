package compartmentdefinition

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo CompartmentDefinitionRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo CompartmentDefinitionRepository) *Service {
	return &Service{repo: repo}
}

var validStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

var validCodes = map[string]bool{
	"Patient": true, "Encounter": true, "RelatedPerson": true, "Practitioner": true, "Device": true,
}

func (s *Service) CreateCompartmentDefinition(ctx context.Context, cd *CompartmentDefinition) error {
	if cd.URL == "" {
		return fmt.Errorf("url is required")
	}
	if cd.Name == "" {
		return fmt.Errorf("name is required")
	}
	if cd.Code == "" {
		return fmt.Errorf("code is required")
	}
	if !validCodes[cd.Code] {
		return fmt.Errorf("invalid code: %s", cd.Code)
	}
	if cd.Status == "" {
		cd.Status = "draft"
	}
	if !validStatuses[cd.Status] {
		return fmt.Errorf("invalid status: %s", cd.Status)
	}
	if err := s.repo.Create(ctx, cd); err != nil {
		return err
	}
	cd.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "CompartmentDefinition", cd.FHIRID, cd.ToFHIR())
	}
	return nil
}

func (s *Service) GetCompartmentDefinition(ctx context.Context, id uuid.UUID) (*CompartmentDefinition, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetCompartmentDefinitionByFHIRID(ctx context.Context, fhirID string) (*CompartmentDefinition, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateCompartmentDefinition(ctx context.Context, cd *CompartmentDefinition) error {
	if cd.Status != "" && !validStatuses[cd.Status] {
		return fmt.Errorf("invalid status: %s", cd.Status)
	}
	if cd.Code != "" && !validCodes[cd.Code] {
		return fmt.Errorf("invalid code: %s", cd.Code)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "CompartmentDefinition", cd.FHIRID, cd.VersionID, cd.ToFHIR())
		if err == nil {
			cd.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, cd)
}

func (s *Service) DeleteCompartmentDefinition(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		cd, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "CompartmentDefinition", cd.FHIRID, cd.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchCompartmentDefinitions(ctx context.Context, params map[string]string, limit, offset int) ([]*CompartmentDefinition, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
