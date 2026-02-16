package organizationaffiliation

import (
	"context"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo OrganizationAffiliationRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo OrganizationAffiliationRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateOrganizationAffiliation(ctx context.Context, o *OrganizationAffiliation) error {
	if err := s.repo.Create(ctx, o); err != nil {
		return err
	}
	o.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "OrganizationAffiliation", o.FHIRID, o.ToFHIR())
	}
	return nil
}

func (s *Service) GetOrganizationAffiliation(ctx context.Context, id uuid.UUID) (*OrganizationAffiliation, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetOrganizationAffiliationByFHIRID(ctx context.Context, fhirID string) (*OrganizationAffiliation, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateOrganizationAffiliation(ctx context.Context, o *OrganizationAffiliation) error {
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "OrganizationAffiliation", o.FHIRID, o.VersionID, o.ToFHIR())
		if err == nil {
			o.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, o)
}

func (s *Service) DeleteOrganizationAffiliation(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		o, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "OrganizationAffiliation", o.FHIRID, o.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchOrganizationAffiliations(ctx context.Context, params map[string]string, limit, offset int) ([]*OrganizationAffiliation, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
