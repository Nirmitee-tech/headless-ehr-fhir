package linkage

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo LinkageRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo LinkageRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateLinkage(ctx context.Context, l *Linkage) error {
	if l.SourceType == "" {
		return fmt.Errorf("source_type is required")
	}
	if l.SourceReference == "" {
		return fmt.Errorf("source_reference is required")
	}
	if err := s.repo.Create(ctx, l); err != nil {
		return err
	}
	l.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Linkage", l.FHIRID, l.ToFHIR())
	}
	return nil
}

func (s *Service) GetLinkage(ctx context.Context, id uuid.UUID) (*Linkage, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetLinkageByFHIRID(ctx context.Context, fhirID string) (*Linkage, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateLinkage(ctx context.Context, l *Linkage) error {
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Linkage", l.FHIRID, l.VersionID, l.ToFHIR())
		if err == nil {
			l.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, l)
}

func (s *Service) DeleteLinkage(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		l, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Linkage", l.FHIRID, l.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchLinkages(ctx context.Context, params map[string]string, limit, offset int) ([]*Linkage, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
