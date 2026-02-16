package media

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo MediaRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo MediaRepository) *Service {
	return &Service{repo: repo}
}

var validMediaStatuses = map[string]bool{
	"preparation": true, "in-progress": true, "not-done": true,
	"on-hold": true, "stopped": true, "completed": true,
	"entered-in-error": true, "unknown": true,
}

func (s *Service) CreateMedia(ctx context.Context, m *Media) error {
	if m.Status == "" {
		m.Status = "completed"
	}
	if !validMediaStatuses[m.Status] {
		return fmt.Errorf("invalid status: %s", m.Status)
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return err
	}
	m.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Media", m.FHIRID, m.ToFHIR())
	}
	return nil
}

func (s *Service) GetMedia(ctx context.Context, id uuid.UUID) (*Media, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetMediaByFHIRID(ctx context.Context, fhirID string) (*Media, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateMedia(ctx context.Context, m *Media) error {
	if m.Status != "" && !validMediaStatuses[m.Status] {
		return fmt.Errorf("invalid status: %s", m.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Media", m.FHIRID, m.VersionID, m.ToFHIR())
		if err == nil {
			m.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, m)
}

func (s *Service) DeleteMedia(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		m, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Media", m.FHIRID, m.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchMedia(ctx context.Context, params map[string]string, limit, offset int) ([]*Media, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
