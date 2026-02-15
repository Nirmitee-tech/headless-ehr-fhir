package familyhistory

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Service provides business logic for the FamilyHistory domain.
type Service struct {
	histories FamilyMemberHistoryRepository
	vt        *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

// NewService creates a new FamilyHistory domain service.
func NewService(h FamilyMemberHistoryRepository) *Service {
	return &Service{histories: h}
}

var validFMHStatuses = map[string]bool{
	"partial": true, "completed": true, "entered-in-error": true, "health-unknown": true,
}

func (s *Service) CreateFamilyMemberHistory(ctx context.Context, f *FamilyMemberHistory) error {
	if f.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if f.RelationshipCode == "" {
		return fmt.Errorf("relationship_code is required")
	}
	if f.RelationshipDisplay == "" {
		return fmt.Errorf("relationship_display is required")
	}
	if f.Status == "" {
		f.Status = "completed"
	}
	if !validFMHStatuses[f.Status] {
		return fmt.Errorf("invalid status: %s", f.Status)
	}
	if err := s.histories.Create(ctx, f); err != nil {
		return err
	}
	f.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "FamilyMemberHistory", f.FHIRID, f.ToFHIR())
	}
	return nil
}

func (s *Service) GetFamilyMemberHistory(ctx context.Context, id uuid.UUID) (*FamilyMemberHistory, error) {
	return s.histories.GetByID(ctx, id)
}

func (s *Service) GetFamilyMemberHistoryByFHIRID(ctx context.Context, fhirID string) (*FamilyMemberHistory, error) {
	return s.histories.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateFamilyMemberHistory(ctx context.Context, f *FamilyMemberHistory) error {
	if f.Status != "" && !validFMHStatuses[f.Status] {
		return fmt.Errorf("invalid status: %s", f.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "FamilyMemberHistory", f.FHIRID, f.VersionID, f.ToFHIR())
		if err == nil {
			f.VersionID = newVer
		}
	}
	return s.histories.Update(ctx, f)
}

func (s *Service) DeleteFamilyMemberHistory(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		f, err := s.histories.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "FamilyMemberHistory", f.FHIRID, f.VersionID)
		}
	}
	return s.histories.Delete(ctx, id)
}

func (s *Service) ListFamilyMemberHistoriesByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*FamilyMemberHistory, int, error) {
	return s.histories.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchFamilyMemberHistories(ctx context.Context, params map[string]string, limit, offset int) ([]*FamilyMemberHistory, int, error) {
	return s.histories.Search(ctx, params, limit, offset)
}

func (s *Service) AddCondition(ctx context.Context, c *FamilyMemberCondition) error {
	if c.FamilyMemberID == uuid.Nil {
		return fmt.Errorf("family_member_id is required")
	}
	if c.Code == "" {
		return fmt.Errorf("code is required")
	}
	if c.Display == "" {
		return fmt.Errorf("display is required")
	}
	return s.histories.AddCondition(ctx, c)
}

func (s *Service) GetConditions(ctx context.Context, familyMemberID uuid.UUID) ([]*FamilyMemberCondition, error) {
	return s.histories.GetConditions(ctx, familyMemberID)
}
