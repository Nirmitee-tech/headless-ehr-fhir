package bodystructure

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo BodyStructureRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo BodyStructureRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateBodyStructure(ctx context.Context, b *BodyStructure) error {
	if b.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if err := s.repo.Create(ctx, b); err != nil {
		return err
	}
	b.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "BodyStructure", b.FHIRID, b.ToFHIR())
	}
	return nil
}

func (s *Service) GetBodyStructure(ctx context.Context, id uuid.UUID) (*BodyStructure, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetBodyStructureByFHIRID(ctx context.Context, fhirID string) (*BodyStructure, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateBodyStructure(ctx context.Context, b *BodyStructure) error {
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "BodyStructure", b.FHIRID, b.VersionID, b.ToFHIR())
		if err == nil {
			b.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, b)
}

func (s *Service) DeleteBodyStructure(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		b, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "BodyStructure", b.FHIRID, b.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchBodyStructures(ctx context.Context, params map[string]string, limit, offset int) ([]*BodyStructure, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
