package surgery

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// -- Mock Repositories --

type mockORRoomRepo struct {
	rooms map[uuid.UUID]*ORRoom
}

func newMockORRoomRepo() *mockORRoomRepo {
	return &mockORRoomRepo{rooms: make(map[uuid.UUID]*ORRoom)}
}

func (m *mockORRoomRepo) Create(_ context.Context, r *ORRoom) error {
	r.ID = uuid.New()
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	m.rooms[r.ID] = r
	return nil
}

func (m *mockORRoomRepo) GetByID(_ context.Context, id uuid.UUID) (*ORRoom, error) {
	r, ok := m.rooms[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return r, nil
}

func (m *mockORRoomRepo) Update(_ context.Context, r *ORRoom) error {
	m.rooms[r.ID] = r
	return nil
}

func (m *mockORRoomRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.rooms, id)
	return nil
}

func (m *mockORRoomRepo) List(_ context.Context, limit, offset int) ([]*ORRoom, int, error) {
	var result []*ORRoom
	for _, r := range m.rooms {
		result = append(result, r)
	}
	return result, len(result), nil
}

func (m *mockORRoomRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*ORRoom, int, error) {
	return m.List(context.Background(), limit, offset)
}

type mockSurgicalCaseRepo struct {
	cases      map[uuid.UUID]*SurgicalCase
	procedures map[uuid.UUID]*SurgicalCaseProcedure
	team       map[uuid.UUID]*SurgicalCaseTeam
	events     map[uuid.UUID]*SurgicalTimeEvent
	counts     map[uuid.UUID]*SurgicalCount
	supplies   map[uuid.UUID]*SurgicalSupplyUsed
}

func newMockSurgicalCaseRepo() *mockSurgicalCaseRepo {
	return &mockSurgicalCaseRepo{
		cases:      make(map[uuid.UUID]*SurgicalCase),
		procedures: make(map[uuid.UUID]*SurgicalCaseProcedure),
		team:       make(map[uuid.UUID]*SurgicalCaseTeam),
		events:     make(map[uuid.UUID]*SurgicalTimeEvent),
		counts:     make(map[uuid.UUID]*SurgicalCount),
		supplies:   make(map[uuid.UUID]*SurgicalSupplyUsed),
	}
}

func (m *mockSurgicalCaseRepo) Create(_ context.Context, sc *SurgicalCase) error {
	sc.ID = uuid.New()
	m.cases[sc.ID] = sc
	return nil
}

func (m *mockSurgicalCaseRepo) GetByID(_ context.Context, id uuid.UUID) (*SurgicalCase, error) {
	sc, ok := m.cases[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return sc, nil
}

func (m *mockSurgicalCaseRepo) Update(_ context.Context, sc *SurgicalCase) error {
	m.cases[sc.ID] = sc
	return nil
}

func (m *mockSurgicalCaseRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.cases, id)
	return nil
}

func (m *mockSurgicalCaseRepo) List(_ context.Context, limit, offset int) ([]*SurgicalCase, int, error) {
	var result []*SurgicalCase
	for _, sc := range m.cases {
		result = append(result, sc)
	}
	return result, len(result), nil
}

func (m *mockSurgicalCaseRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*SurgicalCase, int, error) {
	var result []*SurgicalCase
	for _, sc := range m.cases {
		if sc.PatientID == patientID {
			result = append(result, sc)
		}
	}
	return result, len(result), nil
}

func (m *mockSurgicalCaseRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*SurgicalCase, int, error) {
	return m.List(context.Background(), limit, offset)
}

func (m *mockSurgicalCaseRepo) AddProcedure(_ context.Context, p *SurgicalCaseProcedure) error {
	p.ID = uuid.New()
	m.procedures[p.ID] = p
	return nil
}

func (m *mockSurgicalCaseRepo) GetProcedures(_ context.Context, caseID uuid.UUID) ([]*SurgicalCaseProcedure, error) {
	var result []*SurgicalCaseProcedure
	for _, p := range m.procedures {
		if p.SurgicalCaseID == caseID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockSurgicalCaseRepo) RemoveProcedure(_ context.Context, id uuid.UUID) error {
	delete(m.procedures, id)
	return nil
}

func (m *mockSurgicalCaseRepo) AddTeamMember(_ context.Context, t *SurgicalCaseTeam) error {
	t.ID = uuid.New()
	m.team[t.ID] = t
	return nil
}

func (m *mockSurgicalCaseRepo) GetTeamMembers(_ context.Context, caseID uuid.UUID) ([]*SurgicalCaseTeam, error) {
	var result []*SurgicalCaseTeam
	for _, t := range m.team {
		if t.SurgicalCaseID == caseID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockSurgicalCaseRepo) RemoveTeamMember(_ context.Context, id uuid.UUID) error {
	delete(m.team, id)
	return nil
}

func (m *mockSurgicalCaseRepo) AddTimeEvent(_ context.Context, e *SurgicalTimeEvent) error {
	e.ID = uuid.New()
	m.events[e.ID] = e
	return nil
}

func (m *mockSurgicalCaseRepo) GetTimeEvents(_ context.Context, caseID uuid.UUID) ([]*SurgicalTimeEvent, error) {
	var result []*SurgicalTimeEvent
	for _, e := range m.events {
		if e.SurgicalCaseID == caseID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *mockSurgicalCaseRepo) AddCount(_ context.Context, c *SurgicalCount) error {
	c.ID = uuid.New()
	m.counts[c.ID] = c
	return nil
}

func (m *mockSurgicalCaseRepo) GetCounts(_ context.Context, caseID uuid.UUID) ([]*SurgicalCount, error) {
	var result []*SurgicalCount
	for _, c := range m.counts {
		if c.SurgicalCaseID == caseID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockSurgicalCaseRepo) AddSupply(_ context.Context, s *SurgicalSupplyUsed) error {
	s.ID = uuid.New()
	m.supplies[s.ID] = s
	return nil
}

func (m *mockSurgicalCaseRepo) GetSupplies(_ context.Context, caseID uuid.UUID) ([]*SurgicalSupplyUsed, error) {
	var result []*SurgicalSupplyUsed
	for _, s := range m.supplies {
		if s.SurgicalCaseID == caseID {
			result = append(result, s)
		}
	}
	return result, nil
}

type mockPrefCardRepo struct {
	cards map[uuid.UUID]*SurgicalPreferenceCard
}

func newMockPrefCardRepo() *mockPrefCardRepo {
	return &mockPrefCardRepo{cards: make(map[uuid.UUID]*SurgicalPreferenceCard)}
}

func (m *mockPrefCardRepo) Create(_ context.Context, pc *SurgicalPreferenceCard) error {
	pc.ID = uuid.New()
	m.cards[pc.ID] = pc
	return nil
}

func (m *mockPrefCardRepo) GetByID(_ context.Context, id uuid.UUID) (*SurgicalPreferenceCard, error) {
	pc, ok := m.cards[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return pc, nil
}

func (m *mockPrefCardRepo) Update(_ context.Context, pc *SurgicalPreferenceCard) error {
	m.cards[pc.ID] = pc
	return nil
}

func (m *mockPrefCardRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.cards, id)
	return nil
}

func (m *mockPrefCardRepo) ListBySurgeon(_ context.Context, surgeonID uuid.UUID, limit, offset int) ([]*SurgicalPreferenceCard, int, error) {
	var result []*SurgicalPreferenceCard
	for _, pc := range m.cards {
		if pc.SurgeonID == surgeonID {
			result = append(result, pc)
		}
	}
	return result, len(result), nil
}

func (m *mockPrefCardRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*SurgicalPreferenceCard, int, error) {
	var result []*SurgicalPreferenceCard
	for _, pc := range m.cards {
		result = append(result, pc)
	}
	return result, len(result), nil
}

type mockImplantLogRepo struct {
	logs map[uuid.UUID]*ImplantLog
}

func newMockImplantLogRepo() *mockImplantLogRepo {
	return &mockImplantLogRepo{logs: make(map[uuid.UUID]*ImplantLog)}
}

func (m *mockImplantLogRepo) Create(_ context.Context, il *ImplantLog) error {
	il.ID = uuid.New()
	m.logs[il.ID] = il
	return nil
}

func (m *mockImplantLogRepo) GetByID(_ context.Context, id uuid.UUID) (*ImplantLog, error) {
	il, ok := m.logs[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return il, nil
}

func (m *mockImplantLogRepo) Update(_ context.Context, il *ImplantLog) error {
	m.logs[il.ID] = il
	return nil
}

func (m *mockImplantLogRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.logs, id)
	return nil
}

func (m *mockImplantLogRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*ImplantLog, int, error) {
	var result []*ImplantLog
	for _, il := range m.logs {
		if il.PatientID == patientID {
			result = append(result, il)
		}
	}
	return result, len(result), nil
}

func (m *mockImplantLogRepo) ListByCase(_ context.Context, caseID uuid.UUID, limit, offset int) ([]*ImplantLog, int, error) {
	var result []*ImplantLog
	for _, il := range m.logs {
		if il.SurgicalCaseID != nil && *il.SurgicalCaseID == caseID {
			result = append(result, il)
		}
	}
	return result, len(result), nil
}

func (m *mockImplantLogRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*ImplantLog, int, error) {
	var result []*ImplantLog
	for _, il := range m.logs {
		result = append(result, il)
	}
	return result, len(result), nil
}

// -- Tests --

func newTestService() *Service {
	return NewService(newMockORRoomRepo(), newMockSurgicalCaseRepo(), newMockPrefCardRepo(), newMockImplantLogRepo())
}

func TestCreateORRoom(t *testing.T) {
	svc := newTestService()
	r := &ORRoom{Name: "OR-1"}
	err := svc.CreateORRoom(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if r.Status != "available" {
		t.Errorf("expected default status 'available', got %s", r.Status)
	}
	if !r.IsActive {
		t.Error("expected is_active to be true")
	}
}

func TestCreateORRoom_NameRequired(t *testing.T) {
	svc := newTestService()
	r := &ORRoom{}
	err := svc.CreateORRoom(context.Background(), r)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestCreateORRoom_InvalidStatus(t *testing.T) {
	svc := newTestService()
	r := &ORRoom{Name: "OR-1", Status: "invalid"}
	err := svc.CreateORRoom(context.Background(), r)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestGetORRoom(t *testing.T) {
	svc := newTestService()
	r := &ORRoom{Name: "OR-1"}
	svc.CreateORRoom(context.Background(), r)

	fetched, err := svc.GetORRoom(context.Background(), r.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.Name != "OR-1" {
		t.Errorf("expected name 'OR-1', got %s", fetched.Name)
	}
}

func TestDeleteORRoom(t *testing.T) {
	svc := newTestService()
	r := &ORRoom{Name: "OR-1"}
	svc.CreateORRoom(context.Background(), r)

	err := svc.DeleteORRoom(context.Background(), r.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetORRoom(context.Background(), r.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestCreateSurgicalCase(t *testing.T) {
	svc := newTestService()
	sc := &SurgicalCase{
		PatientID:        uuid.New(),
		PrimarySurgeonID: uuid.New(),
		ScheduledDate:    time.Now(),
	}
	err := svc.CreateSurgicalCase(context.Background(), sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sc.Status != "scheduled" {
		t.Errorf("expected default status 'scheduled', got %s", sc.Status)
	}
}

func TestCreateSurgicalCase_PatientRequired(t *testing.T) {
	svc := newTestService()
	sc := &SurgicalCase{PrimarySurgeonID: uuid.New(), ScheduledDate: time.Now()}
	err := svc.CreateSurgicalCase(context.Background(), sc)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateSurgicalCase_SurgeonRequired(t *testing.T) {
	svc := newTestService()
	sc := &SurgicalCase{PatientID: uuid.New(), ScheduledDate: time.Now()}
	err := svc.CreateSurgicalCase(context.Background(), sc)
	if err == nil {
		t.Error("expected error for missing primary_surgeon_id")
	}
}

func TestCreateSurgicalCase_DateRequired(t *testing.T) {
	svc := newTestService()
	sc := &SurgicalCase{PatientID: uuid.New(), PrimarySurgeonID: uuid.New()}
	err := svc.CreateSurgicalCase(context.Background(), sc)
	if err == nil {
		t.Error("expected error for missing scheduled_date")
	}
}

func TestAddCaseProcedure(t *testing.T) {
	svc := newTestService()
	caseID := uuid.New()
	p := &SurgicalCaseProcedure{SurgicalCaseID: caseID, ProcedureCode: "12345"}
	err := svc.AddCaseProcedure(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	procs, err := svc.GetCaseProcedures(context.Background(), caseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(procs) != 1 {
		t.Fatalf("expected 1 procedure, got %d", len(procs))
	}
}

func TestAddCaseProcedure_CodeRequired(t *testing.T) {
	svc := newTestService()
	p := &SurgicalCaseProcedure{SurgicalCaseID: uuid.New()}
	err := svc.AddCaseProcedure(context.Background(), p)
	if err == nil {
		t.Error("expected error for missing procedure_code")
	}
}

func TestAddCaseTeamMember(t *testing.T) {
	svc := newTestService()
	caseID := uuid.New()
	tm := &SurgicalCaseTeam{SurgicalCaseID: caseID, PractitionerID: uuid.New(), Role: "surgeon"}
	err := svc.AddCaseTeamMember(context.Background(), tm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	members, _ := svc.GetCaseTeamMembers(context.Background(), caseID)
	if len(members) != 1 {
		t.Fatalf("expected 1 team member, got %d", len(members))
	}
}

func TestAddCaseTeamMember_RoleRequired(t *testing.T) {
	svc := newTestService()
	tm := &SurgicalCaseTeam{SurgicalCaseID: uuid.New(), PractitionerID: uuid.New()}
	err := svc.AddCaseTeamMember(context.Background(), tm)
	if err == nil {
		t.Error("expected error for missing role")
	}
}

func TestAddCaseTimeEvent(t *testing.T) {
	svc := newTestService()
	caseID := uuid.New()
	e := &SurgicalTimeEvent{SurgicalCaseID: caseID, EventType: "incision"}
	err := svc.AddCaseTimeEvent(context.Background(), e)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.EventTime.IsZero() {
		t.Error("expected event_time to be defaulted")
	}
}

func TestAddCaseCount(t *testing.T) {
	svc := newTestService()
	caseID := uuid.New()
	c := &SurgicalCount{SurgicalCaseID: caseID, ItemName: "sponge", ExpectedCount: 10, ActualCount: 10}
	err := svc.AddCaseCount(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.IsCorrect {
		t.Error("expected is_correct to be true when counts match")
	}
}

func TestAddCaseSupply(t *testing.T) {
	svc := newTestService()
	caseID := uuid.New()
	su := &SurgicalSupplyUsed{SurgicalCaseID: caseID, SupplyName: "gauze"}
	err := svc.AddCaseSupply(context.Background(), su)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if su.Quantity != 1 {
		t.Errorf("expected default quantity 1, got %d", su.Quantity)
	}
}

func TestCreatePreferenceCard(t *testing.T) {
	svc := newTestService()
	pc := &SurgicalPreferenceCard{SurgeonID: uuid.New(), ProcedureCode: "12345"}
	err := svc.CreatePreferenceCard(context.Background(), pc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !pc.IsActive {
		t.Error("expected is_active to be true")
	}
}

func TestCreatePreferenceCard_SurgeonRequired(t *testing.T) {
	svc := newTestService()
	pc := &SurgicalPreferenceCard{ProcedureCode: "12345"}
	err := svc.CreatePreferenceCard(context.Background(), pc)
	if err == nil {
		t.Error("expected error for missing surgeon_id")
	}
}

func TestCreateImplantLog(t *testing.T) {
	svc := newTestService()
	il := &ImplantLog{PatientID: uuid.New(), ImplantType: "knee"}
	err := svc.CreateImplantLog(context.Background(), il)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if il.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestCreateImplantLog_PatientRequired(t *testing.T) {
	svc := newTestService()
	il := &ImplantLog{ImplantType: "knee"}
	err := svc.CreateImplantLog(context.Background(), il)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateImplantLog_TypeRequired(t *testing.T) {
	svc := newTestService()
	il := &ImplantLog{PatientID: uuid.New()}
	err := svc.CreateImplantLog(context.Background(), il)
	if err == nil {
		t.Error("expected error for missing implant_type")
	}
}

// -- Additional OR Room Tests --

func TestUpdateORRoom(t *testing.T) {
	svc := newTestService()
	r := &ORRoom{Name: "OR-1"}
	svc.CreateORRoom(context.Background(), r)
	r.Status = "in-use"
	err := svc.UpdateORRoom(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateORRoom_InvalidStatus(t *testing.T) {
	svc := newTestService()
	r := &ORRoom{Name: "OR-1"}
	svc.CreateORRoom(context.Background(), r)
	r.Status = "bogus"
	err := svc.UpdateORRoom(context.Background(), r)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestListORRooms(t *testing.T) {
	svc := newTestService()
	svc.CreateORRoom(context.Background(), &ORRoom{Name: "OR-1"})
	svc.CreateORRoom(context.Background(), &ORRoom{Name: "OR-2"})
	rooms, total, err := svc.ListORRooms(context.Background(), 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(rooms) != 2 {
		t.Errorf("expected 2 rooms, got %d", len(rooms))
	}
}

func TestSearchORRooms(t *testing.T) {
	svc := newTestService()
	svc.CreateORRoom(context.Background(), &ORRoom{Name: "OR-1"})
	rooms, total, err := svc.SearchORRooms(context.Background(), map[string]string{"status": "available"}, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(rooms) < 1 {
		t.Error("expected rooms")
	}
}

// -- Additional Surgical Case Tests --

func TestGetSurgicalCase(t *testing.T) {
	svc := newTestService()
	sc := &SurgicalCase{PatientID: uuid.New(), PrimarySurgeonID: uuid.New(), ScheduledDate: time.Now()}
	svc.CreateSurgicalCase(context.Background(), sc)
	fetched, err := svc.GetSurgicalCase(context.Background(), sc.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != sc.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestGetSurgicalCase_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetSurgicalCase(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateSurgicalCase(t *testing.T) {
	svc := newTestService()
	sc := &SurgicalCase{PatientID: uuid.New(), PrimarySurgeonID: uuid.New(), ScheduledDate: time.Now()}
	svc.CreateSurgicalCase(context.Background(), sc)
	sc.Status = "in-or"
	err := svc.UpdateSurgicalCase(context.Background(), sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateSurgicalCase_InvalidStatus(t *testing.T) {
	svc := newTestService()
	sc := &SurgicalCase{PatientID: uuid.New(), PrimarySurgeonID: uuid.New(), ScheduledDate: time.Now()}
	svc.CreateSurgicalCase(context.Background(), sc)
	sc.Status = "bogus"
	err := svc.UpdateSurgicalCase(context.Background(), sc)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestDeleteSurgicalCase(t *testing.T) {
	svc := newTestService()
	sc := &SurgicalCase{PatientID: uuid.New(), PrimarySurgeonID: uuid.New(), ScheduledDate: time.Now()}
	svc.CreateSurgicalCase(context.Background(), sc)
	err := svc.DeleteSurgicalCase(context.Background(), sc.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetSurgicalCase(context.Background(), sc.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestListSurgicalCases(t *testing.T) {
	svc := newTestService()
	svc.CreateSurgicalCase(context.Background(), &SurgicalCase{PatientID: uuid.New(), PrimarySurgeonID: uuid.New(), ScheduledDate: time.Now()})
	cases, total, err := svc.ListSurgicalCases(context.Background(), 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(cases) != 1 {
		t.Errorf("expected 1 case, got %d", len(cases))
	}
}

func TestListSurgicalCasesByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreateSurgicalCase(context.Background(), &SurgicalCase{PatientID: patientID, PrimarySurgeonID: uuid.New(), ScheduledDate: time.Now()})
	svc.CreateSurgicalCase(context.Background(), &SurgicalCase{PatientID: uuid.New(), PrimarySurgeonID: uuid.New(), ScheduledDate: time.Now()})
	cases, total, err := svc.ListSurgicalCasesByPatient(context.Background(), patientID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(cases) != 1 {
		t.Errorf("expected 1 case, got %d", len(cases))
	}
}

func TestSearchSurgicalCases(t *testing.T) {
	svc := newTestService()
	svc.CreateSurgicalCase(context.Background(), &SurgicalCase{PatientID: uuid.New(), PrimarySurgeonID: uuid.New(), ScheduledDate: time.Now()})
	cases, total, err := svc.SearchSurgicalCases(context.Background(), map[string]string{}, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(cases) < 1 {
		t.Error("expected cases")
	}
}

// -- Additional Sub-Resource Tests --

func TestRemoveCaseProcedure(t *testing.T) {
	svc := newTestService()
	caseID := uuid.New()
	p := &SurgicalCaseProcedure{SurgicalCaseID: caseID, ProcedureCode: "12345"}
	svc.AddCaseProcedure(context.Background(), p)
	err := svc.RemoveCaseProcedure(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	procs, _ := svc.GetCaseProcedures(context.Background(), caseID)
	if len(procs) != 0 {
		t.Errorf("expected 0 procedures after removal, got %d", len(procs))
	}
}

func TestRemoveCaseTeamMember(t *testing.T) {
	svc := newTestService()
	caseID := uuid.New()
	tm := &SurgicalCaseTeam{SurgicalCaseID: caseID, PractitionerID: uuid.New(), Role: "surgeon"}
	svc.AddCaseTeamMember(context.Background(), tm)
	err := svc.RemoveCaseTeamMember(context.Background(), tm.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	members, _ := svc.GetCaseTeamMembers(context.Background(), caseID)
	if len(members) != 0 {
		t.Errorf("expected 0 members after removal, got %d", len(members))
	}
}

func TestGetCaseTimeEvents(t *testing.T) {
	svc := newTestService()
	caseID := uuid.New()
	svc.AddCaseTimeEvent(context.Background(), &SurgicalTimeEvent{SurgicalCaseID: caseID, EventType: "incision"})
	events, err := svc.GetCaseTimeEvents(context.Background(), caseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
}

func TestGetCaseCounts(t *testing.T) {
	svc := newTestService()
	caseID := uuid.New()
	svc.AddCaseCount(context.Background(), &SurgicalCount{SurgicalCaseID: caseID, ItemName: "sponge", ExpectedCount: 10, ActualCount: 10})
	counts, err := svc.GetCaseCounts(context.Background(), caseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(counts) != 1 {
		t.Errorf("expected 1 count, got %d", len(counts))
	}
}

func TestGetCaseSupplies(t *testing.T) {
	svc := newTestService()
	caseID := uuid.New()
	svc.AddCaseSupply(context.Background(), &SurgicalSupplyUsed{SurgicalCaseID: caseID, SupplyName: "gauze"})
	supplies, err := svc.GetCaseSupplies(context.Background(), caseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(supplies) != 1 {
		t.Errorf("expected 1 supply, got %d", len(supplies))
	}
}

// -- Additional Preference Card Tests --

func TestGetPreferenceCard(t *testing.T) {
	svc := newTestService()
	pc := &SurgicalPreferenceCard{SurgeonID: uuid.New(), ProcedureCode: "12345"}
	svc.CreatePreferenceCard(context.Background(), pc)
	fetched, err := svc.GetPreferenceCard(context.Background(), pc.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != pc.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestGetPreferenceCard_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetPreferenceCard(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdatePreferenceCard(t *testing.T) {
	svc := newTestService()
	pc := &SurgicalPreferenceCard{SurgeonID: uuid.New(), ProcedureCode: "12345"}
	svc.CreatePreferenceCard(context.Background(), pc)
	pc.ProcedureCode = "67890"
	err := svc.UpdatePreferenceCard(context.Background(), pc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeletePreferenceCard(t *testing.T) {
	svc := newTestService()
	pc := &SurgicalPreferenceCard{SurgeonID: uuid.New(), ProcedureCode: "12345"}
	svc.CreatePreferenceCard(context.Background(), pc)
	err := svc.DeletePreferenceCard(context.Background(), pc.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetPreferenceCard(context.Background(), pc.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestListPreferenceCardsBySurgeon(t *testing.T) {
	svc := newTestService()
	surgeonID := uuid.New()
	svc.CreatePreferenceCard(context.Background(), &SurgicalPreferenceCard{SurgeonID: surgeonID, ProcedureCode: "111"})
	svc.CreatePreferenceCard(context.Background(), &SurgicalPreferenceCard{SurgeonID: surgeonID, ProcedureCode: "222"})
	cards, total, err := svc.ListPreferenceCardsBySurgeon(context.Background(), surgeonID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(cards) != 2 {
		t.Errorf("expected 2 cards, got %d", len(cards))
	}
}

func TestSearchPreferenceCards(t *testing.T) {
	svc := newTestService()
	svc.CreatePreferenceCard(context.Background(), &SurgicalPreferenceCard{SurgeonID: uuid.New(), ProcedureCode: "111"})
	cards, total, err := svc.SearchPreferenceCards(context.Background(), map[string]string{}, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(cards) < 1 {
		t.Error("expected cards")
	}
}

// -- Additional Implant Log Tests --

func TestGetImplantLog(t *testing.T) {
	svc := newTestService()
	il := &ImplantLog{PatientID: uuid.New(), ImplantType: "knee"}
	svc.CreateImplantLog(context.Background(), il)
	fetched, err := svc.GetImplantLog(context.Background(), il.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != il.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestGetImplantLog_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetImplantLog(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateImplantLog(t *testing.T) {
	svc := newTestService()
	il := &ImplantLog{PatientID: uuid.New(), ImplantType: "knee"}
	svc.CreateImplantLog(context.Background(), il)
	il.ImplantType = "hip"
	err := svc.UpdateImplantLog(context.Background(), il)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteImplantLog(t *testing.T) {
	svc := newTestService()
	il := &ImplantLog{PatientID: uuid.New(), ImplantType: "knee"}
	svc.CreateImplantLog(context.Background(), il)
	err := svc.DeleteImplantLog(context.Background(), il.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetImplantLog(context.Background(), il.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestListImplantLogsByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreateImplantLog(context.Background(), &ImplantLog{PatientID: patientID, ImplantType: "knee"})
	logs, total, err := svc.ListImplantLogsByPatient(context.Background(), patientID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}
}

func TestSearchImplantLogs(t *testing.T) {
	svc := newTestService()
	svc.CreateImplantLog(context.Background(), &ImplantLog{PatientID: uuid.New(), ImplantType: "knee"})
	logs, total, err := svc.SearchImplantLogs(context.Background(), map[string]string{}, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(logs) < 1 {
		t.Error("expected logs")
	}
}
