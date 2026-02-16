package devicerequest

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo DeviceRequestRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo DeviceRequestRepository) *Service {
	return &Service{repo: repo}
}

var validDeviceRequestStatuses = map[string]bool{
	"draft": true, "active": true, "on-hold": true, "revoked": true,
	"completed": true, "entered-in-error": true, "unknown": true,
}

var validDeviceRequestIntents = map[string]bool{
	"proposal": true, "plan": true, "directive": true, "order": true,
	"original-order": true, "reflex-order": true, "filler-order": true,
	"instance-order": true, "option": true,
}

func (s *Service) CreateDeviceRequest(ctx context.Context, d *DeviceRequest) error {
	if d.SubjectPatientID == uuid.Nil {
		return fmt.Errorf("subject_patient_id is required")
	}
	if d.Status == "" {
		d.Status = "active"
	}
	if !validDeviceRequestStatuses[d.Status] {
		return fmt.Errorf("invalid status: %s", d.Status)
	}
	if d.Intent == "" {
		d.Intent = "order"
	}
	if !validDeviceRequestIntents[d.Intent] {
		return fmt.Errorf("invalid intent: %s", d.Intent)
	}
	if err := s.repo.Create(ctx, d); err != nil {
		return err
	}
	d.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "DeviceRequest", d.FHIRID, d.ToFHIR())
	}
	return nil
}

func (s *Service) GetDeviceRequest(ctx context.Context, id uuid.UUID) (*DeviceRequest, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetDeviceRequestByFHIRID(ctx context.Context, fhirID string) (*DeviceRequest, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateDeviceRequest(ctx context.Context, d *DeviceRequest) error {
	if d.Status != "" && !validDeviceRequestStatuses[d.Status] {
		return fmt.Errorf("invalid status: %s", d.Status)
	}
	if d.Intent != "" && !validDeviceRequestIntents[d.Intent] {
		return fmt.Errorf("invalid intent: %s", d.Intent)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "DeviceRequest", d.FHIRID, d.VersionID, d.ToFHIR())
		if err == nil {
			d.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, d)
}

func (s *Service) DeleteDeviceRequest(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		d, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "DeviceRequest", d.FHIRID, d.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchDeviceRequests(ctx context.Context, params map[string]string, limit, offset int) ([]*DeviceRequest, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
