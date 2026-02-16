package measure

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo MeasureRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo MeasureRepository) *Service {
	return &Service{repo: repo}
}

var validMeasureStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

func (s *Service) CreateMeasure(ctx context.Context, m *Measure) error {
	if m.Status == "" {
		m.Status = "draft"
	}
	if !validMeasureStatuses[m.Status] {
		return fmt.Errorf("invalid status: %s", m.Status)
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return err
	}
	m.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Measure", m.FHIRID, m.ToFHIR())
	}
	return nil
}

func (s *Service) GetMeasure(ctx context.Context, id uuid.UUID) (*Measure, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetMeasureByFHIRID(ctx context.Context, fhirID string) (*Measure, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateMeasure(ctx context.Context, m *Measure) error {
	if m.Status != "" && !validMeasureStatuses[m.Status] {
		return fmt.Errorf("invalid status: %s", m.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Measure", m.FHIRID, m.VersionID, m.ToFHIR())
		if err == nil {
			m.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, m)
}

func (s *Service) DeleteMeasure(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		m, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Measure", m.FHIRID, m.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchMeasures(ctx context.Context, params map[string]string, limit, offset int) ([]*Measure, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
