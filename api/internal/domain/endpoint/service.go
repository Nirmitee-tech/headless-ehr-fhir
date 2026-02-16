package endpoint

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo EndpointRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo EndpointRepository) *Service {
	return &Service{repo: repo}
}

var validEndpointStatuses = map[string]bool{
	"active": true, "suspended": true, "error": true, "off": true, "entered-in-error": true, "test": true,
}

func (s *Service) CreateEndpoint(ctx context.Context, e *Endpoint) error {
	if e.Address == "" {
		return fmt.Errorf("address is required")
	}
	if e.Status == "" {
		e.Status = "active"
	}
	if !validEndpointStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return err
	}
	e.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Endpoint", e.FHIRID, e.ToFHIR())
	}
	return nil
}

func (s *Service) GetEndpoint(ctx context.Context, id uuid.UUID) (*Endpoint, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetEndpointByFHIRID(ctx context.Context, fhirID string) (*Endpoint, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateEndpoint(ctx context.Context, e *Endpoint) error {
	if e.Status != "" && !validEndpointStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Endpoint", e.FHIRID, e.VersionID, e.ToFHIR())
		if err == nil {
			e.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, e)
}

func (s *Service) DeleteEndpoint(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		e, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Endpoint", e.FHIRID, e.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchEndpoints(ctx context.Context, params map[string]string, limit, offset int) ([]*Endpoint, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
