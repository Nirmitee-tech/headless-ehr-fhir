package deviceusestatement

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo DeviceUseStatementRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo DeviceUseStatementRepository) *Service {
	return &Service{repo: repo}
}

var validDeviceUseStatementStatuses = map[string]bool{
	"active": true, "completed": true, "entered-in-error": true,
	"intended": true, "stopped": true, "on-hold": true,
}

func (s *Service) CreateDeviceUseStatement(ctx context.Context, d *DeviceUseStatement) error {
	if d.SubjectPatientID == uuid.Nil {
		return fmt.Errorf("subject_patient_id is required")
	}
	if d.Status == "" {
		d.Status = "active"
	}
	if !validDeviceUseStatementStatuses[d.Status] {
		return fmt.Errorf("invalid status: %s", d.Status)
	}
	if err := s.repo.Create(ctx, d); err != nil {
		return err
	}
	d.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "DeviceUseStatement", d.FHIRID, d.ToFHIR())
	}
	return nil
}

func (s *Service) GetDeviceUseStatement(ctx context.Context, id uuid.UUID) (*DeviceUseStatement, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetDeviceUseStatementByFHIRID(ctx context.Context, fhirID string) (*DeviceUseStatement, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateDeviceUseStatement(ctx context.Context, d *DeviceUseStatement) error {
	if d.Status != "" && !validDeviceUseStatementStatuses[d.Status] {
		return fmt.Errorf("invalid status: %s", d.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "DeviceUseStatement", d.FHIRID, d.VersionID, d.ToFHIR())
		if err == nil {
			d.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, d)
}

func (s *Service) DeleteDeviceUseStatement(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		d, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "DeviceUseStatement", d.FHIRID, d.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchDeviceUseStatements(ctx context.Context, params map[string]string, limit, offset int) ([]*DeviceUseStatement, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
