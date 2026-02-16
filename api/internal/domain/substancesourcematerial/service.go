package substancesourcematerial

import (
	"context"
	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo SubstanceSourceMaterialRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }
func NewService(repo SubstanceSourceMaterialRepository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, m *SubstanceSourceMaterial) error {
	if err := s.repo.Create(ctx, m); err != nil { return err }
	m.VersionID = 1
	if s.vt != nil { _ = s.vt.RecordCreate(ctx, "SubstanceSourceMaterial", m.FHIRID, m.ToFHIR()) }
	return nil
}
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*SubstanceSourceMaterial, error) { return s.repo.GetByID(ctx, id) }
func (s *Service) GetByFHIRID(ctx context.Context, fhirID string) (*SubstanceSourceMaterial, error) { return s.repo.GetByFHIRID(ctx, fhirID) }
func (s *Service) Update(ctx context.Context, m *SubstanceSourceMaterial) error {
	if s.vt != nil { nv, err := s.vt.RecordUpdate(ctx, "SubstanceSourceMaterial", m.FHIRID, m.VersionID, m.ToFHIR()); if err == nil { m.VersionID = nv } }
	return s.repo.Update(ctx, m)
}
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil { m, err := s.repo.GetByID(ctx, id); if err == nil { _ = s.vt.RecordDelete(ctx, "SubstanceSourceMaterial", m.FHIRID, m.VersionID) } }
	return s.repo.Delete(ctx, id)
}
func (s *Service) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SubstanceSourceMaterial, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
