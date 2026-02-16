package library

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo LibraryRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo LibraryRepository) *Service {
	return &Service{repo: repo}
}

var validLibraryStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

func (s *Service) CreateLibrary(ctx context.Context, l *Library) error {
	if l.TypeCode == "" {
		return fmt.Errorf("type code is required")
	}
	if l.Status == "" {
		l.Status = "draft"
	}
	if !validLibraryStatuses[l.Status] {
		return fmt.Errorf("invalid status: %s", l.Status)
	}
	if err := s.repo.Create(ctx, l); err != nil {
		return err
	}
	l.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Library", l.FHIRID, l.ToFHIR())
	}
	return nil
}

func (s *Service) GetLibrary(ctx context.Context, id uuid.UUID) (*Library, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetLibraryByFHIRID(ctx context.Context, fhirID string) (*Library, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateLibrary(ctx context.Context, l *Library) error {
	if l.Status != "" && !validLibraryStatuses[l.Status] {
		return fmt.Errorf("invalid status: %s", l.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Library", l.FHIRID, l.VersionID, l.ToFHIR())
		if err == nil {
			l.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, l)
}

func (s *Service) DeleteLibrary(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		l, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Library", l.FHIRID, l.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchLibraries(ctx context.Context, params map[string]string, limit, offset int) ([]*Library, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
