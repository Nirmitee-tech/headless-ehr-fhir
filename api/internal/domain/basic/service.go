package basic

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo BasicRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo BasicRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateBasic(ctx context.Context, b *Basic) error {
	if b.CodeCode == "" {
		return fmt.Errorf("code_code is required")
	}
	if err := s.repo.Create(ctx, b); err != nil {
		return err
	}
	b.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Basic", b.FHIRID, b.ToFHIR())
	}
	return nil
}

func (s *Service) GetBasic(ctx context.Context, id uuid.UUID) (*Basic, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetBasicByFHIRID(ctx context.Context, fhirID string) (*Basic, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateBasic(ctx context.Context, b *Basic) error {
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Basic", b.FHIRID, b.VersionID, b.ToFHIR())
		if err == nil {
			b.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, b)
}

func (s *Service) DeleteBasic(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		b, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Basic", b.FHIRID, b.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchBasics(ctx context.Context, params map[string]string, limit, offset int) ([]*Basic, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
