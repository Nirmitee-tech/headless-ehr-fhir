package devicemetric

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo DeviceMetricRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo DeviceMetricRepository) *Service {
	return &Service{repo: repo}
}

var validOperationalStatuses = map[string]bool{
	"on": true, "off": true, "standby": true, "entered-in-error": true,
}

var validCategories = map[string]bool{
	"measurement": true, "setting": true, "calculation": true, "unspecified": true,
}

func (s *Service) CreateDeviceMetric(ctx context.Context, m *DeviceMetric) error {
	if m.TypeCode == "" {
		return fmt.Errorf("type code is required")
	}
	if m.Category == "" {
		m.Category = "unspecified"
	}
	if !validCategories[m.Category] {
		return fmt.Errorf("invalid category: %s", m.Category)
	}
	if m.OperationalStatus != nil && !validOperationalStatuses[*m.OperationalStatus] {
		return fmt.Errorf("invalid operational status: %s", *m.OperationalStatus)
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return err
	}
	m.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "DeviceMetric", m.FHIRID, m.ToFHIR())
	}
	return nil
}

func (s *Service) GetDeviceMetric(ctx context.Context, id uuid.UUID) (*DeviceMetric, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetDeviceMetricByFHIRID(ctx context.Context, fhirID string) (*DeviceMetric, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateDeviceMetric(ctx context.Context, m *DeviceMetric) error {
	if m.OperationalStatus != nil && !validOperationalStatuses[*m.OperationalStatus] {
		return fmt.Errorf("invalid operational status: %s", *m.OperationalStatus)
	}
	if m.Category != "" && !validCategories[m.Category] {
		return fmt.Errorf("invalid category: %s", m.Category)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "DeviceMetric", m.FHIRID, m.VersionID, m.ToFHIR())
		if err == nil {
			m.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, m)
}

func (s *Service) DeleteDeviceMetric(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		m, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "DeviceMetric", m.FHIRID, m.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchDeviceMetrics(ctx context.Context, params map[string]string, limit, offset int) ([]*DeviceMetric, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
