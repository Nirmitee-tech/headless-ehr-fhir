package structuremap

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo StructureMapRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo StructureMapRepository) *Service {
	return &Service{repo: repo}
}

var validStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

var validStructureModes = map[string]bool{
	"source": true, "queried": true, "target": true, "produced": true,
}

func (s *Service) CreateStructureMap(ctx context.Context, sm *StructureMap) error {
	if sm.URL == "" {
		return fmt.Errorf("url is required")
	}
	if sm.Name == "" {
		return fmt.Errorf("name is required")
	}
	if sm.Status == "" {
		sm.Status = "draft"
	}
	if !validStatuses[sm.Status] {
		return fmt.Errorf("invalid status: %s", sm.Status)
	}
	if sm.StructureMode != nil && !validStructureModes[*sm.StructureMode] {
		return fmt.Errorf("invalid structure mode: %s", *sm.StructureMode)
	}
	if err := s.repo.Create(ctx, sm); err != nil {
		return err
	}
	sm.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "StructureMap", sm.FHIRID, sm.ToFHIR())
	}
	return nil
}

func (s *Service) GetStructureMap(ctx context.Context, id uuid.UUID) (*StructureMap, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetStructureMapByFHIRID(ctx context.Context, fhirID string) (*StructureMap, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateStructureMap(ctx context.Context, sm *StructureMap) error {
	if sm.Status != "" && !validStatuses[sm.Status] {
		return fmt.Errorf("invalid status: %s", sm.Status)
	}
	if sm.StructureMode != nil && !validStructureModes[*sm.StructureMode] {
		return fmt.Errorf("invalid structure mode: %s", *sm.StructureMode)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "StructureMap", sm.FHIRID, sm.VersionID, sm.ToFHIR())
		if err == nil {
			sm.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, sm)
}

func (s *Service) DeleteStructureMap(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		sm, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "StructureMap", sm.FHIRID, sm.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchStructureMaps(ctx context.Context, params map[string]string, limit, offset int) ([]*StructureMap, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
