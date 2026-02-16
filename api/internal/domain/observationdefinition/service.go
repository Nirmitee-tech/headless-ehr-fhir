package observationdefinition

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo ObservationDefinitionRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo ObservationDefinitionRepository) *Service {
	return &Service{repo: repo}
}

var validObservationDefinitionStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

func (s *Service) CreateObservationDefinition(ctx context.Context, od *ObservationDefinition) error {
	if od.CodeCode == "" {
		return fmt.Errorf("code is required")
	}
	if od.Status == "" {
		od.Status = "draft"
	}
	if !validObservationDefinitionStatuses[od.Status] {
		return fmt.Errorf("invalid status: %s", od.Status)
	}
	if err := s.repo.Create(ctx, od); err != nil {
		return err
	}
	od.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ObservationDefinition", od.FHIRID, od.ToFHIR())
	}
	return nil
}

func (s *Service) GetObservationDefinition(ctx context.Context, id uuid.UUID) (*ObservationDefinition, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetObservationDefinitionByFHIRID(ctx context.Context, fhirID string) (*ObservationDefinition, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateObservationDefinition(ctx context.Context, od *ObservationDefinition) error {
	if od.Status != "" && !validObservationDefinitionStatuses[od.Status] {
		return fmt.Errorf("invalid status: %s", od.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ObservationDefinition", od.FHIRID, od.VersionID, od.ToFHIR())
		if err == nil {
			od.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, od)
}

func (s *Service) DeleteObservationDefinition(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		od, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ObservationDefinition", od.FHIRID, od.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchObservationDefinitions(ctx context.Context, params map[string]string, limit, offset int) ([]*ObservationDefinition, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
