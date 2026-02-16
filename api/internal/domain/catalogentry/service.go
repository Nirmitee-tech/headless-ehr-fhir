package catalogentry

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo CatalogEntryRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo CatalogEntryRepository) *Service {
	return &Service{repo: repo}
}

var validCatalogEntryStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

func (s *Service) CreateCatalogEntry(ctx context.Context, ce *CatalogEntry) error {
	if ce.ReferencedItemType == "" || ce.ReferencedItemReference == "" {
		return fmt.Errorf("referencedItem type and reference are required")
	}
	if ce.Status == "" {
		ce.Status = "draft"
	}
	if !validCatalogEntryStatuses[ce.Status] {
		return fmt.Errorf("invalid status: %s", ce.Status)
	}
	if err := s.repo.Create(ctx, ce); err != nil {
		return err
	}
	ce.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "CatalogEntry", ce.FHIRID, ce.ToFHIR())
	}
	return nil
}

func (s *Service) GetCatalogEntry(ctx context.Context, id uuid.UUID) (*CatalogEntry, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetCatalogEntryByFHIRID(ctx context.Context, fhirID string) (*CatalogEntry, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateCatalogEntry(ctx context.Context, ce *CatalogEntry) error {
	if ce.Status != "" && !validCatalogEntryStatuses[ce.Status] {
		return fmt.Errorf("invalid status: %s", ce.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "CatalogEntry", ce.FHIRID, ce.VersionID, ce.ToFHIR())
		if err == nil {
			ce.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, ce)
}

func (s *Service) DeleteCatalogEntry(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		ce, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "CatalogEntry", ce.FHIRID, ce.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchCatalogEntries(ctx context.Context, params map[string]string, limit, offset int) ([]*CatalogEntry, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
