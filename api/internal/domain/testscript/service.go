package testscript

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo TestScriptRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo TestScriptRepository) *Service {
	return &Service{repo: repo}
}

var validStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

func (s *Service) CreateTestScript(ctx context.Context, ts *TestScript) error {
	if ts.Name == "" {
		return fmt.Errorf("name is required")
	}
	if ts.Status == "" {
		ts.Status = "draft"
	}
	if !validStatuses[ts.Status] {
		return fmt.Errorf("invalid status: %s", ts.Status)
	}
	if err := s.repo.Create(ctx, ts); err != nil {
		return err
	}
	ts.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "TestScript", ts.FHIRID, ts.ToFHIR())
	}
	return nil
}

func (s *Service) GetTestScript(ctx context.Context, id uuid.UUID) (*TestScript, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetTestScriptByFHIRID(ctx context.Context, fhirID string) (*TestScript, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateTestScript(ctx context.Context, ts *TestScript) error {
	if ts.Status != "" && !validStatuses[ts.Status] {
		return fmt.Errorf("invalid status: %s", ts.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "TestScript", ts.FHIRID, ts.VersionID, ts.ToFHIR())
		if err == nil {
			ts.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, ts)
}

func (s *Service) DeleteTestScript(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		ts, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "TestScript", ts.FHIRID, ts.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchTestScripts(ctx context.Context, params map[string]string, limit, offset int) ([]*TestScript, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
