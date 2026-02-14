package surgery

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	orRooms    ORRoomRepository
	cases      SurgicalCaseRepository
	prefCards  PreferenceCardRepository
	implants   ImplantLogRepository
}

func NewService(orRooms ORRoomRepository, cases SurgicalCaseRepository, prefCards PreferenceCardRepository, implants ImplantLogRepository) *Service {
	return &Service{orRooms: orRooms, cases: cases, prefCards: prefCards, implants: implants}
}

// -- OR Room --

var validORRoomStatuses = map[string]bool{
	"available": true, "in-use": true, "turnover": true,
	"blocked": true, "maintenance": true,
}

func (s *Service) CreateORRoom(ctx context.Context, r *ORRoom) error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if r.Status == "" {
		r.Status = "available"
	}
	if !validORRoomStatuses[r.Status] {
		return fmt.Errorf("invalid status: %s", r.Status)
	}
	r.IsActive = true
	return s.orRooms.Create(ctx, r)
}

func (s *Service) GetORRoom(ctx context.Context, id uuid.UUID) (*ORRoom, error) {
	return s.orRooms.GetByID(ctx, id)
}

func (s *Service) UpdateORRoom(ctx context.Context, r *ORRoom) error {
	if r.Status != "" && !validORRoomStatuses[r.Status] {
		return fmt.Errorf("invalid status: %s", r.Status)
	}
	return s.orRooms.Update(ctx, r)
}

func (s *Service) DeleteORRoom(ctx context.Context, id uuid.UUID) error {
	return s.orRooms.Delete(ctx, id)
}

func (s *Service) ListORRooms(ctx context.Context, limit, offset int) ([]*ORRoom, int, error) {
	return s.orRooms.List(ctx, limit, offset)
}

func (s *Service) SearchORRooms(ctx context.Context, params map[string]string, limit, offset int) ([]*ORRoom, int, error) {
	return s.orRooms.Search(ctx, params, limit, offset)
}

// -- Surgical Case --

var validSurgicalCaseStatuses = map[string]bool{
	"scheduled": true, "pre-op": true, "in-or": true,
	"in-pacu": true, "completed": true, "cancelled": true, "postponed": true,
}

func (s *Service) CreateSurgicalCase(ctx context.Context, sc *SurgicalCase) error {
	if sc.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if sc.PrimarySurgeonID == uuid.Nil {
		return fmt.Errorf("primary_surgeon_id is required")
	}
	if sc.ScheduledDate.IsZero() {
		return fmt.Errorf("scheduled_date is required")
	}
	if sc.Status == "" {
		sc.Status = "scheduled"
	}
	if !validSurgicalCaseStatuses[sc.Status] {
		return fmt.Errorf("invalid status: %s", sc.Status)
	}
	return s.cases.Create(ctx, sc)
}

func (s *Service) GetSurgicalCase(ctx context.Context, id uuid.UUID) (*SurgicalCase, error) {
	return s.cases.GetByID(ctx, id)
}

func (s *Service) UpdateSurgicalCase(ctx context.Context, sc *SurgicalCase) error {
	if sc.Status != "" && !validSurgicalCaseStatuses[sc.Status] {
		return fmt.Errorf("invalid status: %s", sc.Status)
	}
	return s.cases.Update(ctx, sc)
}

func (s *Service) DeleteSurgicalCase(ctx context.Context, id uuid.UUID) error {
	return s.cases.Delete(ctx, id)
}

func (s *Service) ListSurgicalCases(ctx context.Context, limit, offset int) ([]*SurgicalCase, int, error) {
	return s.cases.List(ctx, limit, offset)
}

func (s *Service) ListSurgicalCasesByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*SurgicalCase, int, error) {
	return s.cases.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchSurgicalCases(ctx context.Context, params map[string]string, limit, offset int) ([]*SurgicalCase, int, error) {
	return s.cases.Search(ctx, params, limit, offset)
}

// -- Surgical Case Sub-Resources --

func (s *Service) AddCaseProcedure(ctx context.Context, p *SurgicalCaseProcedure) error {
	if p.SurgicalCaseID == uuid.Nil {
		return fmt.Errorf("surgical_case_id is required")
	}
	if p.ProcedureCode == "" {
		return fmt.Errorf("procedure_code is required")
	}
	return s.cases.AddProcedure(ctx, p)
}

func (s *Service) GetCaseProcedures(ctx context.Context, caseID uuid.UUID) ([]*SurgicalCaseProcedure, error) {
	return s.cases.GetProcedures(ctx, caseID)
}

func (s *Service) RemoveCaseProcedure(ctx context.Context, id uuid.UUID) error {
	return s.cases.RemoveProcedure(ctx, id)
}

func (s *Service) AddCaseTeamMember(ctx context.Context, t *SurgicalCaseTeam) error {
	if t.SurgicalCaseID == uuid.Nil {
		return fmt.Errorf("surgical_case_id is required")
	}
	if t.PractitionerID == uuid.Nil {
		return fmt.Errorf("practitioner_id is required")
	}
	if t.Role == "" {
		return fmt.Errorf("role is required")
	}
	return s.cases.AddTeamMember(ctx, t)
}

func (s *Service) GetCaseTeamMembers(ctx context.Context, caseID uuid.UUID) ([]*SurgicalCaseTeam, error) {
	return s.cases.GetTeamMembers(ctx, caseID)
}

func (s *Service) RemoveCaseTeamMember(ctx context.Context, id uuid.UUID) error {
	return s.cases.RemoveTeamMember(ctx, id)
}

func (s *Service) AddCaseTimeEvent(ctx context.Context, e *SurgicalTimeEvent) error {
	if e.SurgicalCaseID == uuid.Nil {
		return fmt.Errorf("surgical_case_id is required")
	}
	if e.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if e.EventTime.IsZero() {
		e.EventTime = time.Now()
	}
	return s.cases.AddTimeEvent(ctx, e)
}

func (s *Service) GetCaseTimeEvents(ctx context.Context, caseID uuid.UUID) ([]*SurgicalTimeEvent, error) {
	return s.cases.GetTimeEvents(ctx, caseID)
}

func (s *Service) AddCaseCount(ctx context.Context, c *SurgicalCount) error {
	if c.SurgicalCaseID == uuid.Nil {
		return fmt.Errorf("surgical_case_id is required")
	}
	if c.ItemName == "" {
		return fmt.Errorf("item_name is required")
	}
	if c.CountTime.IsZero() {
		c.CountTime = time.Now()
	}
	c.IsCorrect = c.ExpectedCount == c.ActualCount
	return s.cases.AddCount(ctx, c)
}

func (s *Service) GetCaseCounts(ctx context.Context, caseID uuid.UUID) ([]*SurgicalCount, error) {
	return s.cases.GetCounts(ctx, caseID)
}

func (s *Service) AddCaseSupply(ctx context.Context, su *SurgicalSupplyUsed) error {
	if su.SurgicalCaseID == uuid.Nil {
		return fmt.Errorf("surgical_case_id is required")
	}
	if su.SupplyName == "" {
		return fmt.Errorf("supply_name is required")
	}
	if su.Quantity <= 0 {
		su.Quantity = 1
	}
	return s.cases.AddSupply(ctx, su)
}

func (s *Service) GetCaseSupplies(ctx context.Context, caseID uuid.UUID) ([]*SurgicalSupplyUsed, error) {
	return s.cases.GetSupplies(ctx, caseID)
}

// -- Preference Card --

func (s *Service) CreatePreferenceCard(ctx context.Context, pc *SurgicalPreferenceCard) error {
	if pc.SurgeonID == uuid.Nil {
		return fmt.Errorf("surgeon_id is required")
	}
	if pc.ProcedureCode == "" {
		return fmt.Errorf("procedure_code is required")
	}
	pc.IsActive = true
	return s.prefCards.Create(ctx, pc)
}

func (s *Service) GetPreferenceCard(ctx context.Context, id uuid.UUID) (*SurgicalPreferenceCard, error) {
	return s.prefCards.GetByID(ctx, id)
}

func (s *Service) UpdatePreferenceCard(ctx context.Context, pc *SurgicalPreferenceCard) error {
	return s.prefCards.Update(ctx, pc)
}

func (s *Service) DeletePreferenceCard(ctx context.Context, id uuid.UUID) error {
	return s.prefCards.Delete(ctx, id)
}

func (s *Service) ListPreferenceCardsBySurgeon(ctx context.Context, surgeonID uuid.UUID, limit, offset int) ([]*SurgicalPreferenceCard, int, error) {
	return s.prefCards.ListBySurgeon(ctx, surgeonID, limit, offset)
}

func (s *Service) SearchPreferenceCards(ctx context.Context, params map[string]string, limit, offset int) ([]*SurgicalPreferenceCard, int, error) {
	return s.prefCards.Search(ctx, params, limit, offset)
}

// -- Implant Log --

func (s *Service) CreateImplantLog(ctx context.Context, il *ImplantLog) error {
	if il.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if il.ImplantType == "" {
		return fmt.Errorf("implant_type is required")
	}
	return s.implants.Create(ctx, il)
}

func (s *Service) GetImplantLog(ctx context.Context, id uuid.UUID) (*ImplantLog, error) {
	return s.implants.GetByID(ctx, id)
}

func (s *Service) UpdateImplantLog(ctx context.Context, il *ImplantLog) error {
	return s.implants.Update(ctx, il)
}

func (s *Service) DeleteImplantLog(ctx context.Context, id uuid.UUID) error {
	return s.implants.Delete(ctx, id)
}

func (s *Service) ListImplantLogsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ImplantLog, int, error) {
	return s.implants.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchImplantLogs(ctx context.Context, params map[string]string, limit, offset int) ([]*ImplantLog, int, error) {
	return s.implants.Search(ctx, params, limit, offset)
}
