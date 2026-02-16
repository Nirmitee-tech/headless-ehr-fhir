package codesystem

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo CodeSystemRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo CodeSystemRepository) *Service {
	return &Service{repo: repo}
}

var validCodeSystemStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

var validContentValues = map[string]bool{
	"not-present": true, "example": true, "fragment": true, "complete": true, "supplement": true,
}

func (s *Service) CreateCodeSystem(ctx context.Context, cs *CodeSystem) error {
	if cs.Content == "" {
		return fmt.Errorf("content is required")
	}
	if !validContentValues[cs.Content] {
		return fmt.Errorf("invalid content: %s", cs.Content)
	}
	if cs.Status == "" {
		cs.Status = "draft"
	}
	if !validCodeSystemStatuses[cs.Status] {
		return fmt.Errorf("invalid status: %s", cs.Status)
	}
	if err := s.repo.Create(ctx, cs); err != nil {
		return err
	}
	cs.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "CodeSystem", cs.FHIRID, cs.ToFHIR())
	}
	return nil
}

func (s *Service) GetCodeSystem(ctx context.Context, id uuid.UUID) (*CodeSystem, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetCodeSystemByFHIRID(ctx context.Context, fhirID string) (*CodeSystem, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateCodeSystem(ctx context.Context, cs *CodeSystem) error {
	if cs.Status != "" && !validCodeSystemStatuses[cs.Status] {
		return fmt.Errorf("invalid status: %s", cs.Status)
	}
	if cs.Content != "" && !validContentValues[cs.Content] {
		return fmt.Errorf("invalid content: %s", cs.Content)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "CodeSystem", cs.FHIRID, cs.VersionID, cs.ToFHIR())
		if err == nil {
			cs.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, cs)
}

func (s *Service) DeleteCodeSystem(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		cs, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "CodeSystem", cs.FHIRID, cs.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchCodeSystems(ctx context.Context, params map[string]string, limit, offset int) ([]*CodeSystem, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
