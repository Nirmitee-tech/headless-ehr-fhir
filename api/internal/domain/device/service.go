package device

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	devices DeviceRepository
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

func NewService(devices DeviceRepository) *Service {
	return &Service{devices: devices}
}

var validDeviceStatuses = map[string]bool{
	"active": true, "inactive": true, "entered-in-error": true, "unknown": true,
}

func (s *Service) CreateDevice(ctx context.Context, d *Device) error {
	if d.DeviceName == "" {
		return fmt.Errorf("device_name is required")
	}
	if d.Status == "" {
		d.Status = "active"
	}
	if !validDeviceStatuses[d.Status] {
		return fmt.Errorf("invalid status: %s", d.Status)
	}
	if err := s.devices.Create(ctx, d); err != nil {
		return err
	}
	d.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Device", d.FHIRID, d.ToFHIR())
	}
	return nil
}

func (s *Service) GetDevice(ctx context.Context, id uuid.UUID) (*Device, error) {
	return s.devices.GetByID(ctx, id)
}

func (s *Service) GetDeviceByFHIRID(ctx context.Context, fhirID string) (*Device, error) {
	return s.devices.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateDevice(ctx context.Context, d *Device) error {
	if d.Status != "" && !validDeviceStatuses[d.Status] {
		return fmt.Errorf("invalid status: %s", d.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Device", d.FHIRID, d.VersionID, d.ToFHIR())
		if err == nil {
			d.VersionID = newVer
		}
	}
	return s.devices.Update(ctx, d)
}

func (s *Service) DeleteDevice(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		d, err := s.devices.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Device", d.FHIRID, d.VersionID)
		}
	}
	return s.devices.Delete(ctx, id)
}

func (s *Service) ListDevicesByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Device, int, error) {
	return s.devices.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchDevices(ctx context.Context, params map[string]string, limit, offset int) ([]*Device, int, error) {
	return s.devices.Search(ctx, params, limit, offset)
}
