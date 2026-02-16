package valueset

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo ValueSetRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo ValueSetRepository) *Service {
	return &Service{repo: repo}
}

var validValueSetStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

func (s *Service) CreateValueSet(ctx context.Context, vs *ValueSet) error {
	if vs.Status == "" {
		vs.Status = "draft"
	}
	if !validValueSetStatuses[vs.Status] {
		return fmt.Errorf("invalid status: %s", vs.Status)
	}
	if err := s.repo.Create(ctx, vs); err != nil {
		return err
	}
	vs.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ValueSet", vs.FHIRID, vs.ToFHIR())
	}
	return nil
}

func (s *Service) GetValueSet(ctx context.Context, id uuid.UUID) (*ValueSet, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetValueSetByFHIRID(ctx context.Context, fhirID string) (*ValueSet, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateValueSet(ctx context.Context, vs *ValueSet) error {
	if vs.Status != "" && !validValueSetStatuses[vs.Status] {
		return fmt.Errorf("invalid status: %s", vs.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ValueSet", vs.FHIRID, vs.VersionID, vs.ToFHIR())
		if err == nil {
			vs.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, vs)
}

func (s *Service) DeleteValueSet(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		vs, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ValueSet", vs.FHIRID, vs.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchValueSets(ctx context.Context, params map[string]string, limit, offset int) ([]*ValueSet, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
