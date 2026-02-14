package surgery

import (
	"context"

	"github.com/google/uuid"
)

type ORRoomRepository interface {
	Create(ctx context.Context, r *ORRoom) error
	GetByID(ctx context.Context, id uuid.UUID) (*ORRoom, error)
	Update(ctx context.Context, r *ORRoom) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*ORRoom, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ORRoom, int, error)
}

type SurgicalCaseRepository interface {
	Create(ctx context.Context, sc *SurgicalCase) error
	GetByID(ctx context.Context, id uuid.UUID) (*SurgicalCase, error)
	Update(ctx context.Context, sc *SurgicalCase) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*SurgicalCase, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*SurgicalCase, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SurgicalCase, int, error)
	// Procedures
	AddProcedure(ctx context.Context, p *SurgicalCaseProcedure) error
	GetProcedures(ctx context.Context, caseID uuid.UUID) ([]*SurgicalCaseProcedure, error)
	RemoveProcedure(ctx context.Context, id uuid.UUID) error
	// Team
	AddTeamMember(ctx context.Context, t *SurgicalCaseTeam) error
	GetTeamMembers(ctx context.Context, caseID uuid.UUID) ([]*SurgicalCaseTeam, error)
	RemoveTeamMember(ctx context.Context, id uuid.UUID) error
	// Time Events
	AddTimeEvent(ctx context.Context, e *SurgicalTimeEvent) error
	GetTimeEvents(ctx context.Context, caseID uuid.UUID) ([]*SurgicalTimeEvent, error)
	// Counts
	AddCount(ctx context.Context, c *SurgicalCount) error
	GetCounts(ctx context.Context, caseID uuid.UUID) ([]*SurgicalCount, error)
	// Supplies
	AddSupply(ctx context.Context, s *SurgicalSupplyUsed) error
	GetSupplies(ctx context.Context, caseID uuid.UUID) ([]*SurgicalSupplyUsed, error)
}

type PreferenceCardRepository interface {
	Create(ctx context.Context, pc *SurgicalPreferenceCard) error
	GetByID(ctx context.Context, id uuid.UUID) (*SurgicalPreferenceCard, error)
	Update(ctx context.Context, pc *SurgicalPreferenceCard) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListBySurgeon(ctx context.Context, surgeonID uuid.UUID, limit, offset int) ([]*SurgicalPreferenceCard, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SurgicalPreferenceCard, int, error)
}

type ImplantLogRepository interface {
	Create(ctx context.Context, il *ImplantLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*ImplantLog, error)
	Update(ctx context.Context, il *ImplantLog) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ImplantLog, int, error)
	ListByCase(ctx context.Context, caseID uuid.UUID, limit, offset int) ([]*ImplantLog, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ImplantLog, int, error)
}
