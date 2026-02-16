package examplescenario

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo ExampleScenarioRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo ExampleScenarioRepository) *Service {
	return &Service{repo: repo}
}

var validStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

func (s *Service) CreateExampleScenario(ctx context.Context, e *ExampleScenario) error {
	if e.Status == "" {
		e.Status = "draft"
	}
	if !validStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return err
	}
	e.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ExampleScenario", e.FHIRID, e.ToFHIR())
	}
	return nil
}

func (s *Service) GetExampleScenario(ctx context.Context, id uuid.UUID) (*ExampleScenario, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetExampleScenarioByFHIRID(ctx context.Context, fhirID string) (*ExampleScenario, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateExampleScenario(ctx context.Context, e *ExampleScenario) error {
	if e.Status != "" && !validStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ExampleScenario", e.FHIRID, e.VersionID, e.ToFHIR())
		if err == nil {
			e.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, e)
}

func (s *Service) DeleteExampleScenario(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		e, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ExampleScenario", e.FHIRID, e.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchExampleScenarios(ctx context.Context, params map[string]string, limit, offset int) ([]*ExampleScenario, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
