package substancespecification

import (
	"context"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo SubstanceSpecificationRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo SubstanceSpecificationRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateSubstanceSpecification(ctx context.Context, ss *SubstanceSpecification) error {
	if err := s.repo.Create(ctx, ss); err != nil {
		return err
	}
	ss.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "SubstanceSpecification", ss.FHIRID, ss.ToFHIR())
	}
	return nil
}

func (s *Service) GetSubstanceSpecification(ctx context.Context, id uuid.UUID) (*SubstanceSpecification, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetSubstanceSpecificationByFHIRID(ctx context.Context, fhirID string) (*SubstanceSpecification, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateSubstanceSpecification(ctx context.Context, ss *SubstanceSpecification) error {
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "SubstanceSpecification", ss.FHIRID, ss.VersionID, ss.ToFHIR())
		if err == nil {
			ss.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, ss)
}

func (s *Service) DeleteSubstanceSpecification(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		ss, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "SubstanceSpecification", ss.FHIRID, ss.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchSubstanceSpecifications(ctx context.Context, params map[string]string, limit, offset int) ([]*SubstanceSpecification, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
