package measurereport

import (
	"context"
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	reports MeasureReportRepository
	vt      *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(reports MeasureReportRepository) *Service {
	return &Service{reports: reports}
}

var validMeasureReportStatuses = map[string]bool{
	"complete": true, "pending": true, "error": true,
}

func (s *Service) CreateMeasureReport(ctx context.Context, mr *MeasureReport) error {
	if mr.PeriodStart.IsZero() {
		return fmt.Errorf("period_start is required")
	}
	if mr.PeriodEnd.IsZero() {
		return fmt.Errorf("period_end is required")
	}
	if mr.PeriodEnd.Before(mr.PeriodStart) {
		return fmt.Errorf("period_end must be after period_start")
	}
	if mr.Status == "" {
		mr.Status = "pending"
	}
	if !validMeasureReportStatuses[mr.Status] {
		return fmt.Errorf("invalid status: %s", mr.Status)
	}
	if mr.Type == "" {
		mr.Type = "summary"
	}
	if mr.Date == nil {
		now := time.Now()
		mr.Date = &now
	}
	if err := s.reports.Create(ctx, mr); err != nil {
		return err
	}
	mr.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "MeasureReport", mr.FHIRID, mr.ToFHIR())
	}
	return nil
}

func (s *Service) GetMeasureReport(ctx context.Context, id uuid.UUID) (*MeasureReport, error) {
	return s.reports.GetByID(ctx, id)
}

func (s *Service) GetMeasureReportByFHIRID(ctx context.Context, fhirID string) (*MeasureReport, error) {
	return s.reports.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateMeasureReport(ctx context.Context, mr *MeasureReport) error {
	if mr.Status != "" && !validMeasureReportStatuses[mr.Status] {
		return fmt.Errorf("invalid status: %s", mr.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "MeasureReport", mr.FHIRID, mr.VersionID, mr.ToFHIR())
		if err == nil {
			mr.VersionID = newVer
		}
	}
	return s.reports.Update(ctx, mr)
}

func (s *Service) DeleteMeasureReport(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		mr, err := s.reports.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "MeasureReport", mr.FHIRID, mr.VersionID)
		}
	}
	return s.reports.Delete(ctx, id)
}

func (s *Service) SearchMeasureReports(ctx context.Context, params map[string]string, limit, offset int) ([]*MeasureReport, int, error) {
	return s.reports.Search(ctx, params, limit, offset)
}
