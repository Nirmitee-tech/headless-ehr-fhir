package devicedefinition

import (
	"context"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo DeviceDefinitionRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo DeviceDefinitionRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateDeviceDefinition(ctx context.Context, d *DeviceDefinition) error {
	if err := s.repo.Create(ctx, d); err != nil {
		return err
	}
	d.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "DeviceDefinition", d.FHIRID, d.ToFHIR())
	}
	return nil
}

func (s *Service) GetDeviceDefinition(ctx context.Context, id uuid.UUID) (*DeviceDefinition, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetDeviceDefinitionByFHIRID(ctx context.Context, fhirID string) (*DeviceDefinition, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateDeviceDefinition(ctx context.Context, d *DeviceDefinition) error {
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "DeviceDefinition", d.FHIRID, d.VersionID, d.ToFHIR())
		if err == nil {
			d.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, d)
}

func (s *Service) DeleteDeviceDefinition(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		d, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "DeviceDefinition", d.FHIRID, d.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchDeviceDefinitions(ctx context.Context, params map[string]string, limit, offset int) ([]*DeviceDefinition, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
