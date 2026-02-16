package searchparameter

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo SearchParameterRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo SearchParameterRepository) *Service {
	return &Service{repo: repo}
}

var validStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

var validTypes = map[string]bool{
	"number": true, "date": true, "string": true, "token": true,
	"reference": true, "composite": true, "quantity": true, "uri": true, "special": true,
}

func (s *Service) CreateSearchParameter(ctx context.Context, sp *SearchParameter) error {
	if sp.URL == "" {
		return fmt.Errorf("url is required")
	}
	if sp.Name == "" {
		return fmt.Errorf("name is required")
	}
	if sp.Description == "" {
		return fmt.Errorf("description is required")
	}
	if sp.Code == "" {
		return fmt.Errorf("code is required")
	}
	if sp.Base == "" {
		return fmt.Errorf("base is required")
	}
	if sp.Type == "" {
		return fmt.Errorf("type is required")
	}
	if sp.Status == "" {
		sp.Status = "draft"
	}
	if !validStatuses[sp.Status] {
		return fmt.Errorf("invalid status: %s", sp.Status)
	}
	if !validTypes[sp.Type] {
		return fmt.Errorf("invalid type: %s", sp.Type)
	}
	if err := s.repo.Create(ctx, sp); err != nil {
		return err
	}
	sp.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "SearchParameter", sp.FHIRID, sp.ToFHIR())
	}
	return nil
}

func (s *Service) GetSearchParameter(ctx context.Context, id uuid.UUID) (*SearchParameter, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetSearchParameterByFHIRID(ctx context.Context, fhirID string) (*SearchParameter, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateSearchParameter(ctx context.Context, sp *SearchParameter) error {
	if sp.Status != "" && !validStatuses[sp.Status] {
		return fmt.Errorf("invalid status: %s", sp.Status)
	}
	if sp.Type != "" && !validTypes[sp.Type] {
		return fmt.Errorf("invalid type: %s", sp.Type)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "SearchParameter", sp.FHIRID, sp.VersionID, sp.ToFHIR())
		if err == nil {
			sp.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, sp)
}

func (s *Service) DeleteSearchParameter(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		sp, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "SearchParameter", sp.FHIRID, sp.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchSearchParameters(ctx context.Context, params map[string]string, limit, offset int) ([]*SearchParameter, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
