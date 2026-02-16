package implementationguide

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo ImplementationGuideRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo ImplementationGuideRepository) *Service {
	return &Service{repo: repo}
}

var validStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

func (s *Service) CreateImplementationGuide(ctx context.Context, ig *ImplementationGuide) error {
	if ig.URL == "" {
		return fmt.Errorf("url is required")
	}
	if ig.Name == "" {
		return fmt.Errorf("name is required")
	}
	if ig.Status == "" {
		ig.Status = "draft"
	}
	if !validStatuses[ig.Status] {
		return fmt.Errorf("invalid status: %s", ig.Status)
	}
	if err := s.repo.Create(ctx, ig); err != nil {
		return err
	}
	ig.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ImplementationGuide", ig.FHIRID, ig.ToFHIR())
	}
	return nil
}

func (s *Service) GetImplementationGuide(ctx context.Context, id uuid.UUID) (*ImplementationGuide, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetImplementationGuideByFHIRID(ctx context.Context, fhirID string) (*ImplementationGuide, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateImplementationGuide(ctx context.Context, ig *ImplementationGuide) error {
	if ig.Status != "" && !validStatuses[ig.Status] {
		return fmt.Errorf("invalid status: %s", ig.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ImplementationGuide", ig.FHIRID, ig.VersionID, ig.ToFHIR())
		if err == nil {
			ig.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, ig)
}

func (s *Service) DeleteImplementationGuide(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		ig, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ImplementationGuide", ig.FHIRID, ig.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchImplementationGuides(ctx context.Context, params map[string]string, limit, offset int) ([]*ImplementationGuide, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
