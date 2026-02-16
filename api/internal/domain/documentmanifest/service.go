package documentmanifest

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo DocumentManifestRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo DocumentManifestRepository) *Service {
	return &Service{repo: repo}
}

var validDocumentManifestStatuses = map[string]bool{
	"current": true, "superseded": true, "entered-in-error": true,
}

func (s *Service) CreateDocumentManifest(ctx context.Context, d *DocumentManifest) error {
	if d.Status == "" {
		d.Status = "current"
	}
	if !validDocumentManifestStatuses[d.Status] {
		return fmt.Errorf("invalid status: %s", d.Status)
	}
	if err := s.repo.Create(ctx, d); err != nil {
		return err
	}
	d.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "DocumentManifest", d.FHIRID, d.ToFHIR())
	}
	return nil
}

func (s *Service) GetDocumentManifest(ctx context.Context, id uuid.UUID) (*DocumentManifest, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetDocumentManifestByFHIRID(ctx context.Context, fhirID string) (*DocumentManifest, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateDocumentManifest(ctx context.Context, d *DocumentManifest) error {
	if d.Status != "" && !validDocumentManifestStatuses[d.Status] {
		return fmt.Errorf("invalid status: %s", d.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "DocumentManifest", d.FHIRID, d.VersionID, d.ToFHIR())
		if err == nil {
			d.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, d)
}

func (s *Service) DeleteDocumentManifest(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		d, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "DocumentManifest", d.FHIRID, d.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchDocumentManifests(ctx context.Context, params map[string]string, limit, offset int) ([]*DocumentManifest, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
