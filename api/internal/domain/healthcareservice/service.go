package healthcareservice

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	services HealthcareServiceRepository
	vt       *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(services HealthcareServiceRepository) *Service {
	return &Service{services: services}
}

func (s *Service) CreateHealthcareService(ctx context.Context, hs *HealthcareService) error {
	if hs.Name == "" {
		return fmt.Errorf("name is required")
	}
	if err := s.services.Create(ctx, hs); err != nil {
		return err
	}
	hs.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "HealthcareService", hs.FHIRID, hs.ToFHIR())
	}
	return nil
}

func (s *Service) GetHealthcareService(ctx context.Context, id uuid.UUID) (*HealthcareService, error) {
	return s.services.GetByID(ctx, id)
}

func (s *Service) GetHealthcareServiceByFHIRID(ctx context.Context, fhirID string) (*HealthcareService, error) {
	return s.services.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateHealthcareService(ctx context.Context, hs *HealthcareService) error {
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "HealthcareService", hs.FHIRID, hs.VersionID, hs.ToFHIR())
		if err == nil {
			hs.VersionID = newVer
		}
	}
	return s.services.Update(ctx, hs)
}

func (s *Service) DeleteHealthcareService(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		hs, err := s.services.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "HealthcareService", hs.FHIRID, hs.VersionID)
		}
	}
	return s.services.Delete(ctx, id)
}

func (s *Service) SearchHealthcareServices(ctx context.Context, params map[string]string, limit, offset int) ([]*HealthcareService, int, error) {
	return s.services.Search(ctx, params, limit, offset)
}
