package scheduling

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// -- Mock Repositories --

type mockScheduleRepo struct {
	scheds map[uuid.UUID]*Schedule
}

func newMockScheduleRepo() *mockScheduleRepo {
	return &mockScheduleRepo{scheds: make(map[uuid.UUID]*Schedule)}
}

func (m *mockScheduleRepo) Create(_ context.Context, s *Schedule) error {
	s.ID = uuid.New()
	if s.FHIRID == "" {
		s.FHIRID = s.ID.String()
	}
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	m.scheds[s.ID] = s
	return nil
}

func (m *mockScheduleRepo) GetByID(_ context.Context, id uuid.UUID) (*Schedule, error) {
	s, ok := m.scheds[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return s, nil
}

func (m *mockScheduleRepo) GetByFHIRID(_ context.Context, fhirID string) (*Schedule, error) {
	for _, s := range m.scheds {
		if s.FHIRID == fhirID {
			return s, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockScheduleRepo) Update(_ context.Context, s *Schedule) error {
	m.scheds[s.ID] = s
	return nil
}

func (m *mockScheduleRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.scheds, id)
	return nil
}

func (m *mockScheduleRepo) ListByPractitioner(_ context.Context, practitionerID uuid.UUID, limit, offset int) ([]*Schedule, int, error) {
	var result []*Schedule
	for _, s := range m.scheds {
		if s.PractitionerID == practitionerID {
			result = append(result, s)
		}
	}
	return result, len(result), nil
}

func (m *mockScheduleRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*Schedule, int, error) {
	var result []*Schedule
	for _, s := range m.scheds {
		result = append(result, s)
	}
	return result, len(result), nil
}

type mockSlotRepo struct {
	slots map[uuid.UUID]*Slot
}

func newMockSlotRepo() *mockSlotRepo {
	return &mockSlotRepo{slots: make(map[uuid.UUID]*Slot)}
}

func (m *mockSlotRepo) Create(_ context.Context, sl *Slot) error {
	sl.ID = uuid.New()
	if sl.FHIRID == "" {
		sl.FHIRID = sl.ID.String()
	}
	sl.CreatedAt = time.Now()
	sl.UpdatedAt = time.Now()
	m.slots[sl.ID] = sl
	return nil
}

func (m *mockSlotRepo) GetByID(_ context.Context, id uuid.UUID) (*Slot, error) {
	sl, ok := m.slots[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return sl, nil
}

func (m *mockSlotRepo) GetByFHIRID(_ context.Context, fhirID string) (*Slot, error) {
	for _, sl := range m.slots {
		if sl.FHIRID == fhirID {
			return sl, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockSlotRepo) Update(_ context.Context, sl *Slot) error {
	m.slots[sl.ID] = sl
	return nil
}

func (m *mockSlotRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.slots, id)
	return nil
}

func (m *mockSlotRepo) ListBySchedule(_ context.Context, scheduleID uuid.UUID, limit, offset int) ([]*Slot, int, error) {
	var result []*Slot
	for _, sl := range m.slots {
		if sl.ScheduleID == scheduleID {
			result = append(result, sl)
		}
	}
	return result, len(result), nil
}

func (m *mockSlotRepo) SearchAvailable(_ context.Context, _ map[string]string, limit, offset int) ([]*Slot, int, error) {
	var result []*Slot
	for _, sl := range m.slots {
		result = append(result, sl)
	}
	return result, len(result), nil
}

type mockApptRepo struct {
	appts        map[uuid.UUID]*Appointment
	participants map[uuid.UUID]*AppointmentParticipant
}

func newMockApptRepo() *mockApptRepo {
	return &mockApptRepo{
		appts:        make(map[uuid.UUID]*Appointment),
		participants: make(map[uuid.UUID]*AppointmentParticipant),
	}
}

func (m *mockApptRepo) Create(_ context.Context, a *Appointment) error {
	a.ID = uuid.New()
	if a.FHIRID == "" {
		a.FHIRID = a.ID.String()
	}
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()
	m.appts[a.ID] = a
	return nil
}

func (m *mockApptRepo) GetByID(_ context.Context, id uuid.UUID) (*Appointment, error) {
	a, ok := m.appts[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return a, nil
}

func (m *mockApptRepo) GetByFHIRID(_ context.Context, fhirID string) (*Appointment, error) {
	for _, a := range m.appts {
		if a.FHIRID == fhirID {
			return a, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockApptRepo) Update(_ context.Context, a *Appointment) error {
	m.appts[a.ID] = a
	return nil
}

func (m *mockApptRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.appts, id)
	return nil
}

func (m *mockApptRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*Appointment, int, error) {
	var result []*Appointment
	for _, a := range m.appts {
		if a.PatientID == patientID {
			result = append(result, a)
		}
	}
	return result, len(result), nil
}

func (m *mockApptRepo) ListByPractitioner(_ context.Context, practitionerID uuid.UUID, limit, offset int) ([]*Appointment, int, error) {
	var result []*Appointment
	for _, a := range m.appts {
		if a.PractitionerID != nil && *a.PractitionerID == practitionerID {
			result = append(result, a)
		}
	}
	return result, len(result), nil
}

func (m *mockApptRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*Appointment, int, error) {
	var result []*Appointment
	for _, a := range m.appts {
		result = append(result, a)
	}
	return result, len(result), nil
}

func (m *mockApptRepo) AddParticipant(_ context.Context, p *AppointmentParticipant) error {
	p.ID = uuid.New()
	m.participants[p.ID] = p
	return nil
}

func (m *mockApptRepo) GetParticipants(_ context.Context, appointmentID uuid.UUID) ([]*AppointmentParticipant, error) {
	var result []*AppointmentParticipant
	for _, p := range m.participants {
		if p.AppointmentID == appointmentID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockApptRepo) RemoveParticipant(_ context.Context, id uuid.UUID) error {
	delete(m.participants, id)
	return nil
}

type mockWaitlistRepo struct {
	entries map[uuid.UUID]*Waitlist
}

func newMockWaitlistRepo() *mockWaitlistRepo {
	return &mockWaitlistRepo{entries: make(map[uuid.UUID]*Waitlist)}
}

func (m *mockWaitlistRepo) Create(_ context.Context, w *Waitlist) error {
	w.ID = uuid.New()
	w.CreatedAt = time.Now()
	w.UpdatedAt = time.Now()
	m.entries[w.ID] = w
	return nil
}

func (m *mockWaitlistRepo) GetByID(_ context.Context, id uuid.UUID) (*Waitlist, error) {
	w, ok := m.entries[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return w, nil
}

func (m *mockWaitlistRepo) Update(_ context.Context, w *Waitlist) error {
	m.entries[w.ID] = w
	return nil
}

func (m *mockWaitlistRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.entries, id)
	return nil
}

func (m *mockWaitlistRepo) ListByDepartment(_ context.Context, department string, limit, offset int) ([]*Waitlist, int, error) {
	var result []*Waitlist
	for _, w := range m.entries {
		if w.Department != nil && *w.Department == department {
			result = append(result, w)
		}
	}
	return result, len(result), nil
}

func (m *mockWaitlistRepo) ListByPractitioner(_ context.Context, practitionerID uuid.UUID, limit, offset int) ([]*Waitlist, int, error) {
	var result []*Waitlist
	for _, w := range m.entries {
		if w.PractitionerID != nil && *w.PractitionerID == practitionerID {
			result = append(result, w)
		}
	}
	return result, len(result), nil
}

// -- Mock AppointmentResponse Repo --

type mockApptRespRepo struct {
	resps map[uuid.UUID]*AppointmentResponse
}

func newMockApptRespRepo() *mockApptRespRepo {
	return &mockApptRespRepo{resps: make(map[uuid.UUID]*AppointmentResponse)}
}

func (m *mockApptRespRepo) Create(_ context.Context, ar *AppointmentResponse) error {
	ar.ID = uuid.New()
	if ar.FHIRID == "" {
		ar.FHIRID = ar.ID.String()
	}
	ar.CreatedAt = time.Now()
	ar.UpdatedAt = time.Now()
	m.resps[ar.ID] = ar
	return nil
}

func (m *mockApptRespRepo) GetByID(_ context.Context, id uuid.UUID) (*AppointmentResponse, error) {
	ar, ok := m.resps[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return ar, nil
}

func (m *mockApptRespRepo) GetByFHIRID(_ context.Context, fhirID string) (*AppointmentResponse, error) {
	for _, ar := range m.resps {
		if ar.FHIRID == fhirID {
			return ar, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockApptRespRepo) Update(_ context.Context, ar *AppointmentResponse) error {
	m.resps[ar.ID] = ar
	return nil
}

func (m *mockApptRespRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.resps, id)
	return nil
}

func (m *mockApptRespRepo) List(_ context.Context, limit, offset int) ([]*AppointmentResponse, int, error) {
	var result []*AppointmentResponse
	for _, ar := range m.resps {
		result = append(result, ar)
	}
	return result, len(result), nil
}

func (m *mockApptRespRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*AppointmentResponse, int, error) {
	var result []*AppointmentResponse
	for _, ar := range m.resps {
		result = append(result, ar)
	}
	return result, len(result), nil
}

// -- Tests --

func newTestService() *Service {
	return NewService(newMockScheduleRepo(), newMockSlotRepo(), newMockApptRepo(), newMockApptRespRepo(), newMockWaitlistRepo())
}

func TestCreateSchedule(t *testing.T) {
	svc := newTestService()
	s := &Schedule{PractitionerID: uuid.New()}
	err := svc.CreateSchedule(context.Background(), s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Active == nil || !*s.Active {
		t.Error("expected active to default to true")
	}
}

func TestCreateSchedule_PractitionerIDRequired(t *testing.T) {
	svc := newTestService()
	s := &Schedule{}
	err := svc.CreateSchedule(context.Background(), s)
	if err == nil {
		t.Error("expected error for missing practitioner_id")
	}
}

func TestGetSchedule(t *testing.T) {
	svc := newTestService()
	s := &Schedule{PractitionerID: uuid.New()}
	svc.CreateSchedule(context.Background(), s)

	fetched, err := svc.GetSchedule(context.Background(), s.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != s.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestDeleteSchedule(t *testing.T) {
	svc := newTestService()
	s := &Schedule{PractitionerID: uuid.New()}
	svc.CreateSchedule(context.Background(), s)
	err := svc.DeleteSchedule(context.Background(), s.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetSchedule(context.Background(), s.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestCreateSlot(t *testing.T) {
	svc := newTestService()
	sl := &Slot{
		ScheduleID: uuid.New(),
		StartTime:  time.Now(),
		EndTime:    time.Now().Add(30 * time.Minute),
	}
	err := svc.CreateSlot(context.Background(), sl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sl.Status != "free" {
		t.Errorf("expected default status 'free', got %s", sl.Status)
	}
}

func TestCreateSlot_ScheduleIDRequired(t *testing.T) {
	svc := newTestService()
	sl := &Slot{StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)}
	err := svc.CreateSlot(context.Background(), sl)
	if err == nil {
		t.Error("expected error for missing schedule_id")
	}
}

func TestCreateSlot_StartTimeRequired(t *testing.T) {
	svc := newTestService()
	sl := &Slot{ScheduleID: uuid.New(), EndTime: time.Now()}
	err := svc.CreateSlot(context.Background(), sl)
	if err == nil {
		t.Error("expected error for missing start_time")
	}
}

func TestCreateAppointment(t *testing.T) {
	svc := newTestService()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	err := svc.CreateAppointment(context.Background(), a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Status != "proposed" {
		t.Errorf("expected default status 'proposed', got %s", a.Status)
	}
}

func TestCreateAppointment_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	start := time.Now()
	a := &Appointment{StartTime: &start}
	err := svc.CreateAppointment(context.Background(), a)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateAppointment_StartTimeRequired(t *testing.T) {
	svc := newTestService()
	a := &Appointment{PatientID: uuid.New()}
	err := svc.CreateAppointment(context.Background(), a)
	if err == nil {
		t.Error("expected error for missing start_time")
	}
}

func TestAddAppointmentParticipant(t *testing.T) {
	svc := newTestService()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	svc.CreateAppointment(context.Background(), a)

	p := &AppointmentParticipant{
		AppointmentID: a.ID,
		ActorType:     "Practitioner",
		ActorID:       uuid.New(),
	}
	err := svc.AddAppointmentParticipant(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Status != "needs-action" {
		t.Errorf("expected default status 'needs-action', got %s", p.Status)
	}
}

func TestAddAppointmentParticipant_ActorTypeRequired(t *testing.T) {
	svc := newTestService()
	p := &AppointmentParticipant{AppointmentID: uuid.New(), ActorID: uuid.New()}
	err := svc.AddAppointmentParticipant(context.Background(), p)
	if err == nil {
		t.Error("expected error for missing actor_type")
	}
}

func TestCreateWaitlistEntry(t *testing.T) {
	svc := newTestService()
	w := &Waitlist{PatientID: uuid.New()}
	err := svc.CreateWaitlistEntry(context.Background(), w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Status != "waiting" {
		t.Errorf("expected default status 'waiting', got %s", w.Status)
	}
}

func TestCreateWaitlistEntry_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	w := &Waitlist{}
	err := svc.CreateWaitlistEntry(context.Background(), w)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestScheduleToFHIR(t *testing.T) {
	s := &Schedule{
		FHIRID:         "sched-123",
		PractitionerID: uuid.New(),
		UpdatedAt:      time.Now(),
	}
	active := true
	s.Active = &active
	fhirRes := s.ToFHIR()
	if fhirRes["resourceType"] != "Schedule" {
		t.Errorf("expected Schedule, got %v", fhirRes["resourceType"])
	}
	if fhirRes["active"] != true {
		t.Error("expected active true")
	}
}

// -- Additional Schedule Tests --

func TestGetScheduleByFHIRID(t *testing.T) {
	svc := newTestService()
	s := &Schedule{PractitionerID: uuid.New()}
	svc.CreateSchedule(context.Background(), s)
	fetched, err := svc.GetScheduleByFHIRID(context.Background(), s.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != s.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestGetScheduleByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetScheduleByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateSchedule(t *testing.T) {
	svc := newTestService()
	s := &Schedule{PractitionerID: uuid.New()}
	svc.CreateSchedule(context.Background(), s)
	err := svc.UpdateSchedule(context.Background(), s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListSchedulesByPractitioner(t *testing.T) {
	svc := newTestService()
	practID := uuid.New()
	svc.CreateSchedule(context.Background(), &Schedule{PractitionerID: practID})
	svc.CreateSchedule(context.Background(), &Schedule{PractitionerID: practID})
	items, total, err := svc.ListSchedulesByPractitioner(context.Background(), practID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("expected 2, got %d", len(items))
	}
}

func TestSearchSchedules(t *testing.T) {
	svc := newTestService()
	svc.CreateSchedule(context.Background(), &Schedule{PractitionerID: uuid.New()})
	items, total, err := svc.SearchSchedules(context.Background(), map[string]string{}, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(items) < 1 {
		t.Error("expected items")
	}
}

// -- Additional Slot Tests --

func TestGetSlot(t *testing.T) {
	svc := newTestService()
	sl := &Slot{ScheduleID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)}
	svc.CreateSlot(context.Background(), sl)
	fetched, err := svc.GetSlot(context.Background(), sl.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != sl.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestGetSlot_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetSlot(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestGetSlotByFHIRID(t *testing.T) {
	svc := newTestService()
	sl := &Slot{ScheduleID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)}
	svc.CreateSlot(context.Background(), sl)
	fetched, err := svc.GetSlotByFHIRID(context.Background(), sl.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != sl.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestGetSlotByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetSlotByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateSlot(t *testing.T) {
	svc := newTestService()
	sl := &Slot{ScheduleID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)}
	svc.CreateSlot(context.Background(), sl)
	sl.Status = "busy"
	err := svc.UpdateSlot(context.Background(), sl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateSlot_InvalidStatus(t *testing.T) {
	svc := newTestService()
	sl := &Slot{ScheduleID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)}
	svc.CreateSlot(context.Background(), sl)
	sl.Status = "bogus"
	err := svc.UpdateSlot(context.Background(), sl)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestDeleteSlot(t *testing.T) {
	svc := newTestService()
	sl := &Slot{ScheduleID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)}
	svc.CreateSlot(context.Background(), sl)
	err := svc.DeleteSlot(context.Background(), sl.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetSlot(context.Background(), sl.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestListSlotsBySchedule(t *testing.T) {
	svc := newTestService()
	schedID := uuid.New()
	svc.CreateSlot(context.Background(), &Slot{ScheduleID: schedID, StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)})
	slots, total, err := svc.ListSlotsBySchedule(context.Background(), schedID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(slots) != 1 {
		t.Errorf("expected 1 slot, got %d", len(slots))
	}
}

func TestSearchAvailableSlots(t *testing.T) {
	svc := newTestService()
	svc.CreateSlot(context.Background(), &Slot{ScheduleID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)})
	slots, total, err := svc.SearchAvailableSlots(context.Background(), map[string]string{}, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(slots) < 1 {
		t.Error("expected slots")
	}
}

// -- Additional Appointment Tests --

func TestGetAppointment(t *testing.T) {
	svc := newTestService()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	svc.CreateAppointment(context.Background(), a)
	fetched, err := svc.GetAppointment(context.Background(), a.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != a.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestGetAppointment_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetAppointment(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestGetAppointmentByFHIRID(t *testing.T) {
	svc := newTestService()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	svc.CreateAppointment(context.Background(), a)
	fetched, err := svc.GetAppointmentByFHIRID(context.Background(), a.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != a.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestGetAppointmentByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetAppointmentByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateAppointment(t *testing.T) {
	svc := newTestService()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	svc.CreateAppointment(context.Background(), a)
	a.Status = "booked"
	err := svc.UpdateAppointment(context.Background(), a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateAppointment_InvalidStatus(t *testing.T) {
	svc := newTestService()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	svc.CreateAppointment(context.Background(), a)
	a.Status = "bogus"
	err := svc.UpdateAppointment(context.Background(), a)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestDeleteAppointment(t *testing.T) {
	svc := newTestService()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	svc.CreateAppointment(context.Background(), a)
	err := svc.DeleteAppointment(context.Background(), a.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetAppointment(context.Background(), a.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestListAppointmentsByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	start := time.Now()
	svc.CreateAppointment(context.Background(), &Appointment{PatientID: patientID, StartTime: &start})
	items, total, err := svc.ListAppointmentsByPatient(context.Background(), patientID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1, got %d", len(items))
	}
}

func TestListAppointmentsByPractitioner(t *testing.T) {
	svc := newTestService()
	practID := uuid.New()
	start := time.Now()
	svc.CreateAppointment(context.Background(), &Appointment{PatientID: uuid.New(), StartTime: &start, PractitionerID: &practID})
	items, total, err := svc.ListAppointmentsByPractitioner(context.Background(), practID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1, got %d", len(items))
	}
}

func TestSearchAppointments(t *testing.T) {
	svc := newTestService()
	start := time.Now()
	svc.CreateAppointment(context.Background(), &Appointment{PatientID: uuid.New(), StartTime: &start})
	items, total, err := svc.SearchAppointments(context.Background(), map[string]string{}, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(items) < 1 {
		t.Error("expected items")
	}
}

func TestGetAppointmentParticipants(t *testing.T) {
	svc := newTestService()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	svc.CreateAppointment(context.Background(), a)
	svc.AddAppointmentParticipant(context.Background(), &AppointmentParticipant{AppointmentID: a.ID, ActorType: "Practitioner", ActorID: uuid.New()})
	participants, err := svc.GetAppointmentParticipants(context.Background(), a.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(participants) != 1 {
		t.Errorf("expected 1 participant, got %d", len(participants))
	}
}

func TestRemoveAppointmentParticipant(t *testing.T) {
	svc := newTestService()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	svc.CreateAppointment(context.Background(), a)
	p := &AppointmentParticipant{AppointmentID: a.ID, ActorType: "Practitioner", ActorID: uuid.New()}
	svc.AddAppointmentParticipant(context.Background(), p)
	err := svc.RemoveAppointmentParticipant(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parts, _ := svc.GetAppointmentParticipants(context.Background(), a.ID)
	if len(parts) != 0 {
		t.Errorf("expected 0 after removal, got %d", len(parts))
	}
}

// -- Additional Waitlist Tests --

func TestGetWaitlistEntry(t *testing.T) {
	svc := newTestService()
	w := &Waitlist{PatientID: uuid.New()}
	svc.CreateWaitlistEntry(context.Background(), w)
	fetched, err := svc.GetWaitlistEntry(context.Background(), w.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != w.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestGetWaitlistEntry_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetWaitlistEntry(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateWaitlistEntry(t *testing.T) {
	svc := newTestService()
	w := &Waitlist{PatientID: uuid.New()}
	svc.CreateWaitlistEntry(context.Background(), w)
	w.Status = "called"
	err := svc.UpdateWaitlistEntry(context.Background(), w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateWaitlistEntry_InvalidStatus(t *testing.T) {
	svc := newTestService()
	w := &Waitlist{PatientID: uuid.New()}
	svc.CreateWaitlistEntry(context.Background(), w)
	w.Status = "bogus"
	err := svc.UpdateWaitlistEntry(context.Background(), w)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestDeleteWaitlistEntry(t *testing.T) {
	svc := newTestService()
	w := &Waitlist{PatientID: uuid.New()}
	svc.CreateWaitlistEntry(context.Background(), w)
	err := svc.DeleteWaitlistEntry(context.Background(), w.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetWaitlistEntry(context.Background(), w.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestListWaitlistByDepartment(t *testing.T) {
	svc := newTestService()
	dept := "cardiology"
	svc.CreateWaitlistEntry(context.Background(), &Waitlist{PatientID: uuid.New(), Department: &dept})
	items, total, err := svc.ListWaitlistByDepartment(context.Background(), "cardiology", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1, got %d", len(items))
	}
}

func TestListWaitlistByPractitioner(t *testing.T) {
	svc := newTestService()
	practID := uuid.New()
	svc.CreateWaitlistEntry(context.Background(), &Waitlist{PatientID: uuid.New(), PractitionerID: &practID})
	items, total, err := svc.ListWaitlistByPractitioner(context.Background(), practID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1, got %d", len(items))
	}
}

func TestAppointmentToFHIR(t *testing.T) {
	start := time.Now()
	a := &Appointment{
		FHIRID:    "appt-123",
		Status:    "booked",
		PatientID: uuid.New(),
		StartTime: &start,
		UpdatedAt: time.Now(),
	}
	fhirRes := a.ToFHIR()
	if fhirRes["resourceType"] != "Appointment" {
		t.Errorf("expected Appointment, got %v", fhirRes["resourceType"])
	}
	if fhirRes["status"] != "booked" {
		t.Errorf("expected booked, got %v", fhirRes["status"])
	}
}

// =========== Version Tracking Tests ===========

func TestCreateSchedule_WithVersionTracker_SetsVersion1(t *testing.T) {
	svc := newTestService()
	histRepo := fhir.NewHistoryRepository()
	vt := fhir.NewVersionTracker(histRepo)
	svc.SetVersionTracker(vt)

	s := &Schedule{PractitionerID: uuid.New()}
	if err := svc.CreateSchedule(context.Background(), s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.VersionID != 1 {
		t.Errorf("expected VersionID 1 after create, got %d", s.VersionID)
	}
}

func TestCreateAppointment_WithVersionTracker_SetsVersion1(t *testing.T) {
	svc := newTestService()
	histRepo := fhir.NewHistoryRepository()
	vt := fhir.NewVersionTracker(histRepo)
	svc.SetVersionTracker(vt)

	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	if err := svc.CreateAppointment(context.Background(), a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.VersionID != 1 {
		t.Errorf("expected VersionID 1 after create, got %d", a.VersionID)
	}
}

func TestCreateSchedule_NilVersionTracker_NoError(t *testing.T) {
	svc := newTestService()
	s := &Schedule{PractitionerID: uuid.New()}
	if err := svc.CreateSchedule(context.Background(), s); err != nil {
		t.Fatalf("unexpected error with nil version tracker: %v", err)
	}
	if s.VersionID != 1 {
		t.Errorf("expected VersionID 1 after create (even with nil tracker), got %d", s.VersionID)
	}
}

func TestUpdateAppointment_WithVersionTracker(t *testing.T) {
	svc := newTestService()
	histRepo := fhir.NewHistoryRepository()
	vt := fhir.NewVersionTracker(histRepo)
	svc.SetVersionTracker(vt)

	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	svc.CreateAppointment(context.Background(), a)

	a.Status = "booked"
	if err := svc.UpdateAppointment(context.Background(), a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteAppointment_WithVersionTracker(t *testing.T) {
	svc := newTestService()
	histRepo := fhir.NewHistoryRepository()
	vt := fhir.NewVersionTracker(histRepo)
	svc.SetVersionTracker(vt)

	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	svc.CreateAppointment(context.Background(), a)

	if err := svc.DeleteAppointment(context.Background(), a.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSchedulingVersionTrackerAccessor(t *testing.T) {
	svc := newTestService()
	if svc.VersionTracker() != nil {
		t.Error("expected nil VersionTracker initially")
	}

	histRepo := fhir.NewHistoryRepository()
	vt := fhir.NewVersionTracker(histRepo)
	svc.SetVersionTracker(vt)
	if svc.VersionTracker() != vt {
		t.Error("expected VersionTracker to match")
	}
}
