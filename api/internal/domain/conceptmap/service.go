package conceptmap

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo ConceptMapRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo ConceptMapRepository) *Service {
	return &Service{repo: repo}
}

var validConceptMapStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

func (s *Service) CreateConceptMap(ctx context.Context, cm *ConceptMap) error {
	if cm.Status == "" {
		cm.Status = "draft"
	}
	if !validConceptMapStatuses[cm.Status] {
		return fmt.Errorf("invalid status: %s", cm.Status)
	}
	if err := s.repo.Create(ctx, cm); err != nil {
		return err
	}
	cm.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ConceptMap", cm.FHIRID, cm.ToFHIR())
	}
	return nil
}

func (s *Service) GetConceptMap(ctx context.Context, id uuid.UUID) (*ConceptMap, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetConceptMapByFHIRID(ctx context.Context, fhirID string) (*ConceptMap, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateConceptMap(ctx context.Context, cm *ConceptMap) error {
	if cm.Status != "" && !validConceptMapStatuses[cm.Status] {
		return fmt.Errorf("invalid status: %s", cm.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ConceptMap", cm.FHIRID, cm.VersionID, cm.ToFHIR())
		if err == nil {
			cm.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, cm)
}

func (s *Service) DeleteConceptMap(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		cm, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ConceptMap", cm.FHIRID, cm.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchConceptMaps(ctx context.Context, params map[string]string, limit, offset int) ([]*ConceptMap, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
