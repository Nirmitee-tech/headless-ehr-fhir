package familyhistory

import (
	"context"

	"github.com/google/uuid"
)

// FamilyMemberHistoryRepository defines CRUD operations for FamilyMemberHistory resources.
type FamilyMemberHistoryRepository interface {
	Create(ctx context.Context, f *FamilyMemberHistory) error
	GetByID(ctx context.Context, id uuid.UUID) (*FamilyMemberHistory, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*FamilyMemberHistory, error)
	Update(ctx context.Context, f *FamilyMemberHistory) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*FamilyMemberHistory, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*FamilyMemberHistory, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*FamilyMemberHistory, int, error)
	// Conditions
	AddCondition(ctx context.Context, c *FamilyMemberCondition) error
	GetConditions(ctx context.Context, familyMemberID uuid.UUID) ([]*FamilyMemberCondition, error)
}
