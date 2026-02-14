package familyhistory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service provides business logic for the FamilyHistory domain.
type Service struct {
	histories FamilyMemberHistoryRepository
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
	return s.histories.Create(ctx, f)
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
	return s.histories.Update(ctx, f)
}

func (s *Service) DeleteFamilyMemberHistory(ctx context.Context, id uuid.UUID) error {
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
