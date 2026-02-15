package fhirlist

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	lists FHIRListRepository
	vt    *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(lists FHIRListRepository) *Service {
	return &Service{lists: lists}
}

var validListStatuses = map[string]bool{
	"current": true, "retired": true, "entered-in-error": true,
}

var validListModes = map[string]bool{
	"working": true, "snapshot": true, "changes": true,
}

func (s *Service) CreateFHIRList(ctx context.Context, l *FHIRList) error {
	if l.Status == "" {
		l.Status = "current"
	}
	if !validListStatuses[l.Status] {
		return fmt.Errorf("invalid status: %s", l.Status)
	}
	if l.Mode == "" {
		l.Mode = "working"
	}
	if !validListModes[l.Mode] {
		return fmt.Errorf("invalid mode: %s", l.Mode)
	}
	if err := s.lists.Create(ctx, l); err != nil {
		return err
	}
	l.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "List", l.FHIRID, l.ToFHIR())
	}
	return nil
}

func (s *Service) GetFHIRList(ctx context.Context, id uuid.UUID) (*FHIRList, error) {
	return s.lists.GetByID(ctx, id)
}

func (s *Service) GetFHIRListByFHIRID(ctx context.Context, fhirID string) (*FHIRList, error) {
	return s.lists.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateFHIRList(ctx context.Context, l *FHIRList) error {
	if l.Status != "" && !validListStatuses[l.Status] {
		return fmt.Errorf("invalid status: %s", l.Status)
	}
	if l.Mode != "" && !validListModes[l.Mode] {
		return fmt.Errorf("invalid mode: %s", l.Mode)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "List", l.FHIRID, l.VersionID, l.ToFHIR())
		if err == nil {
			l.VersionID = newVer
		}
	}
	return s.lists.Update(ctx, l)
}

func (s *Service) DeleteFHIRList(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		l, err := s.lists.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "List", l.FHIRID, l.VersionID)
		}
	}
	return s.lists.Delete(ctx, id)
}

func (s *Service) SearchFHIRLists(ctx context.Context, params map[string]string, limit, offset int) ([]*FHIRList, int, error) {
	return s.lists.Search(ctx, params, limit, offset)
}

func (s *Service) AddEntry(ctx context.Context, entry *FHIRListEntry) error {
	if entry.ListID == uuid.Nil {
		return fmt.Errorf("list_id is required")
	}
	if entry.ItemReference == "" {
		return fmt.Errorf("item_reference is required")
	}
	return s.lists.AddEntry(ctx, entry)
}

func (s *Service) GetEntries(ctx context.Context, listID uuid.UUID) ([]*FHIRListEntry, error) {
	return s.lists.GetEntries(ctx, listID)
}
