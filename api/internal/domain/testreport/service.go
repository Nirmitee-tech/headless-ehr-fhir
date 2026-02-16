package testreport

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo TestReportRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo TestReportRepository) *Service {
	return &Service{repo: repo}
}

var validStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

var validResults = map[string]bool{
	"pass": true, "fail": true, "pending": true,
}

func (s *Service) CreateTestReport(ctx context.Context, e *TestReport) error {
	if e.Result == "" {
		return fmt.Errorf("result is required")
	}
	if !validResults[e.Result] {
		return fmt.Errorf("invalid result: %s", e.Result)
	}
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
		_ = s.vt.RecordCreate(ctx, "TestReport", e.FHIRID, e.ToFHIR())
	}
	return nil
}

func (s *Service) GetTestReport(ctx context.Context, id uuid.UUID) (*TestReport, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetTestReportByFHIRID(ctx context.Context, fhirID string) (*TestReport, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateTestReport(ctx context.Context, e *TestReport) error {
	if e.Status != "" && !validStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "TestReport", e.FHIRID, e.VersionID, e.ToFHIR())
		if err == nil {
			e.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, e)
}

func (s *Service) DeleteTestReport(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		e, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "TestReport", e.FHIRID, e.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchTestReports(ctx context.Context, params map[string]string, limit, offset int) ([]*TestReport, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
