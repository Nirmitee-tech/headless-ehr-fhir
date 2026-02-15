package scheduling

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------- Helper ----------

func newTestManager() *SelfScheduleManager {
	return NewSelfScheduleManager()
}

func seedSlots(m *SelfScheduleManager) {
	base := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	m.AddSlot(AvailableSlot{
		ID: "slot-1", ScheduleID: "sched-A", Start: base, End: base.Add(30 * time.Minute),
		Duration: 30, ServiceType: "general", ProviderName: "Dr. Smith", Status: "free",
	})
	m.AddSlot(AvailableSlot{
		ID: "slot-2", ScheduleID: "sched-A", Start: base.Add(1 * time.Hour), End: base.Add(90 * time.Minute),
		Duration: 30, ServiceType: "specialist", ProviderName: "Dr. Jones", Status: "free",
	})
	m.AddSlot(AvailableSlot{
		ID: "slot-3", ScheduleID: "sched-B", Start: base.Add(2 * time.Hour), End: base.Add(150 * time.Minute),
		Duration: 30, ServiceType: "general", ProviderName: "Dr. Lee", Status: "free",
	})
	// Slot on a different day.
	m.AddSlot(AvailableSlot{
		ID: "slot-4", ScheduleID: "sched-A", Start: base.AddDate(0, 0, 5), End: base.AddDate(0, 0, 5).Add(30 * time.Minute),
		Duration: 30, ServiceType: "general", Status: "free",
	})
	// Slot with non-free status.
	m.AddSlot(AvailableSlot{
		ID: "slot-5", ScheduleID: "sched-A", Start: base.Add(3 * time.Hour), End: base.Add(210 * time.Minute),
		Duration: 30, ServiceType: "general", Status: "busy",
	})
}

// ---------- Manager Tests ----------

func TestSelfSchedule_FindAvailableSlots_DateRange(t *testing.T) {
	m := newTestManager()
	seedSlots(m)

	params := SlotSearchParams{
		StartDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
	}

	results, err := m.FindAvailableSlots(nil, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// slot-1, slot-2, slot-3 are on Jan 15 and free; slot-4 is Jan 20; slot-5 is busy.
	if len(results) != 3 {
		t.Errorf("expected 3 slots within date range, got %d", len(results))
	}
	// Verify sorted by start time.
	for i := 1; i < len(results); i++ {
		if results[i].Start.Before(results[i-1].Start) {
			t.Error("results are not sorted by start time")
		}
	}
}

func TestSelfSchedule_FindAvailableSlots_ServiceType(t *testing.T) {
	m := newTestManager()
	seedSlots(m)

	params := SlotSearchParams{
		StartDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC),
		ServiceType: "specialist",
	}

	results, err := m.FindAvailableSlots(nil, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 specialist slot, got %d", len(results))
	}
	if len(results) > 0 && results[0].ServiceType != "specialist" {
		t.Errorf("expected service type specialist, got %s", results[0].ServiceType)
	}
}

func TestSelfSchedule_FindAvailableSlots_ScheduleID(t *testing.T) {
	m := newTestManager()
	seedSlots(m)

	params := SlotSearchParams{
		StartDate:  time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
		ScheduleID: "sched-B",
	}

	results, err := m.FindAvailableSlots(nil, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 slot for sched-B, got %d", len(results))
	}
	if len(results) > 0 && results[0].ScheduleID != "sched-B" {
		t.Errorf("expected schedule ID sched-B, got %s", results[0].ScheduleID)
	}
}

func TestSelfSchedule_FindAvailableSlots_ExcludesBooked(t *testing.T) {
	m := newTestManager()
	seedSlots(m)

	// Book slot-1.
	_, err := m.BookAppointment(nil, BookingRequest{SlotID: "slot-1", PatientID: "patient-1"})
	if err != nil {
		t.Fatalf("booking failed: %v", err)
	}

	params := SlotSearchParams{
		StartDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
	}

	results, err := m.FindAvailableSlots(nil, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// slot-1 is booked, slot-5 is busy, so only slot-2 and slot-3.
	if len(results) != 2 {
		t.Errorf("expected 2 available slots (slot-1 booked), got %d", len(results))
	}
	for _, s := range results {
		if s.ID == "slot-1" {
			t.Error("booked slot-1 should not appear in results")
		}
	}
}

func TestSelfSchedule_BookAppointment_Success(t *testing.T) {
	m := newTestManager()
	seedSlots(m)

	conf, err := m.BookAppointment(nil, BookingRequest{
		SlotID:    "slot-1",
		PatientID: "patient-1",
		Reason:    "Check-up",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conf.Status != "booked" {
		t.Errorf("expected status booked, got %s", conf.Status)
	}
	if conf.AppointmentID == "" {
		t.Error("expected non-empty appointment ID")
	}
	if conf.SlotID != "slot-1" {
		t.Errorf("expected slot_id slot-1, got %s", conf.SlotID)
	}
	if conf.PatientID != "patient-1" {
		t.Errorf("expected patient_id patient-1, got %s", conf.PatientID)
	}
	if conf.Reason != "Check-up" {
		t.Errorf("expected reason Check-up, got %s", conf.Reason)
	}
	if conf.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}
}

func TestSelfSchedule_BookAppointment_DoubleBooking(t *testing.T) {
	m := newTestManager()
	seedSlots(m)

	_, err := m.BookAppointment(nil, BookingRequest{SlotID: "slot-1", PatientID: "patient-1"})
	if err != nil {
		t.Fatalf("first booking failed: %v", err)
	}

	_, err = m.BookAppointment(nil, BookingRequest{SlotID: "slot-1", PatientID: "patient-2"})
	if err == nil {
		t.Error("expected error for double booking")
	}
	if err != ErrSlotAlreadyBooked {
		t.Errorf("expected ErrSlotAlreadyBooked, got %v", err)
	}
}

func TestSelfSchedule_BookAppointment_MissingSlotID(t *testing.T) {
	m := newTestManager()

	_, err := m.BookAppointment(nil, BookingRequest{PatientID: "patient-1"})
	if err == nil {
		t.Error("expected error for missing slot_id")
	}
	if err != ErrMissingSlotID {
		t.Errorf("expected ErrMissingSlotID, got %v", err)
	}
}

func TestSelfSchedule_BookAppointment_InvalidSlot(t *testing.T) {
	m := newTestManager()

	_, err := m.BookAppointment(nil, BookingRequest{SlotID: "nonexistent", PatientID: "patient-1"})
	if err == nil {
		t.Error("expected error for invalid slot")
	}
	if err != ErrSlotNotFound {
		t.Errorf("expected ErrSlotNotFound, got %v", err)
	}
}

func TestSelfSchedule_CancelAppointment_Success(t *testing.T) {
	m := newTestManager()
	seedSlots(m)

	conf, _ := m.BookAppointment(nil, BookingRequest{SlotID: "slot-1", PatientID: "patient-1"})

	err := m.CancelAppointment(nil, conf.AppointmentID, "patient-1", "No longer needed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify status changed.
	appt, _ := m.GetAppointment(nil, conf.AppointmentID, "patient-1")
	if appt.Status != "cancelled" {
		t.Errorf("expected status cancelled, got %s", appt.Status)
	}

	// Verify slot is freed — it should appear in available slots again.
	params := SlotSearchParams{
		StartDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
	}
	results, _ := m.FindAvailableSlots(nil, params)
	found := false
	for _, s := range results {
		if s.ID == "slot-1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected slot-1 to be available again after cancellation")
	}
}

func TestSelfSchedule_CancelAppointment_WrongPatient(t *testing.T) {
	m := newTestManager()
	seedSlots(m)

	conf, _ := m.BookAppointment(nil, BookingRequest{SlotID: "slot-1", PatientID: "patient-1"})

	err := m.CancelAppointment(nil, conf.AppointmentID, "patient-other", "")
	if err == nil {
		t.Error("expected error for wrong patient")
	}
	if err != ErrWrongPatient {
		t.Errorf("expected ErrWrongPatient, got %v", err)
	}
}

func TestSelfSchedule_ListPatientAppointments(t *testing.T) {
	m := newTestManager()
	seedSlots(m)

	// Book two appointments for patient-1 and one for patient-2.
	m.BookAppointment(nil, BookingRequest{SlotID: "slot-1", PatientID: "patient-1"})
	m.BookAppointment(nil, BookingRequest{SlotID: "slot-2", PatientID: "patient-1"})
	m.BookAppointment(nil, BookingRequest{SlotID: "slot-3", PatientID: "patient-2"})

	results, err := m.ListPatientAppointments(nil, "patient-1", "", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 appointments for patient-1, got %d", len(results))
	}
	for _, a := range results {
		if a.PatientID != "patient-1" {
			t.Errorf("expected patient-1, got %s", a.PatientID)
		}
	}

	// Test status filter.
	conf, _ := m.BookAppointment(nil, BookingRequest{SlotID: "slot-4", PatientID: "patient-1"})
	m.CancelAppointment(nil, conf.AppointmentID, "patient-1", "")
	bookedResults, _ := m.ListPatientAppointments(nil, "patient-1", "booked", 10)
	if len(bookedResults) != 2 {
		t.Errorf("expected 2 booked appointments for patient-1, got %d", len(bookedResults))
	}
}

func TestSelfSchedule_GetAppointment(t *testing.T) {
	m := newTestManager()
	seedSlots(m)

	conf, _ := m.BookAppointment(nil, BookingRequest{SlotID: "slot-1", PatientID: "patient-1"})

	appt, err := m.GetAppointment(nil, conf.AppointmentID, "patient-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if appt.AppointmentID != conf.AppointmentID {
		t.Errorf("expected appointment ID %s, got %s", conf.AppointmentID, appt.AppointmentID)
	}

	// Wrong patient.
	_, err = m.GetAppointment(nil, conf.AppointmentID, "patient-other")
	if err != ErrWrongPatient {
		t.Errorf("expected ErrWrongPatient, got %v", err)
	}

	// Not found.
	_, err = m.GetAppointment(nil, "nonexistent", "patient-1")
	if err != ErrAppointmentNotFound {
		t.Errorf("expected ErrAppointmentNotFound, got %v", err)
	}
}

// ---------- Handler Tests ----------

func newTestHandler() (*SelfScheduleHandler, *SelfScheduleManager, *echo.Echo) {
	mgr := NewSelfScheduleManager()
	h := NewSelfScheduleHandler(mgr, mgr)
	e := echo.New()
	return h, mgr, e
}

func TestSelfScheduleHandler_SearchSlots(t *testing.T) {
	h, mgr, e := newTestHandler()
	seedSlots(mgr)

	req := httptest.NewRequest(http.MethodGet, "/scheduling/slots?start=2024-01-15&end=2024-01-15", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchSlots(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var slots []AvailableSlot
	if err := json.Unmarshal(rec.Body.Bytes(), &slots); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(slots) != 3 {
		t.Errorf("expected 3 slots, got %d", len(slots))
	}
}

func TestSelfScheduleHandler_SearchSlots_MissingDates(t *testing.T) {
	h, _, e := newTestHandler()

	// Missing both start and end.
	req := httptest.NewRequest(http.MethodGet, "/scheduling/slots", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchSlots(c)
	if err == nil {
		t.Error("expected error for missing dates")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 error, got %v", err)
	}

	// Missing end only.
	req2 := httptest.NewRequest(http.MethodGet, "/scheduling/slots?start=2024-01-15", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err2 := h.SearchSlots(c2)
	if err2 == nil {
		t.Error("expected error for missing end date")
	}
}

func TestSelfScheduleHandler_BookAppointment(t *testing.T) {
	h, mgr, e := newTestHandler()
	seedSlots(mgr)

	body := `{"slot_id":"slot-1","patient_id":"patient-1","reason":"Check-up"}`
	req := httptest.NewRequest(http.MethodPost, "/scheduling/book", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.BookAppointment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var conf BookingConfirmation
	if err := json.Unmarshal(rec.Body.Bytes(), &conf); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if conf.Status != "booked" {
		t.Errorf("expected status booked, got %s", conf.Status)
	}
	if conf.AppointmentID == "" {
		t.Error("expected non-empty appointment ID")
	}
}

func TestSelfScheduleHandler_BookAppointment_Conflict(t *testing.T) {
	h, mgr, e := newTestHandler()
	seedSlots(mgr)

	// First booking succeeds.
	body := `{"slot_id":"slot-1","patient_id":"patient-1"}`
	req := httptest.NewRequest(http.MethodPost, "/scheduling/book", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.BookAppointment(c)
	if err != nil {
		t.Fatalf("first booking failed: %v", err)
	}

	// Second booking should return 409 Conflict.
	body2 := `{"slot_id":"slot-1","patient_id":"patient-2"}`
	req2 := httptest.NewRequest(http.MethodPost, "/scheduling/book", strings.NewReader(body2))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err2 := h.BookAppointment(c2)
	if err2 == nil {
		t.Error("expected error for double booking")
	}
	httpErr, ok := err2.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict, got %v", err2)
	}
}

func TestSelfScheduleHandler_CancelAppointment(t *testing.T) {
	h, mgr, e := newTestHandler()
	seedSlots(mgr)

	// Book first.
	conf, _ := mgr.BookAppointment(nil, BookingRequest{SlotID: "slot-1", PatientID: "patient-1"})

	body := `{"patient_id":"patient-1","reason":"schedule conflict"}`
	req := httptest.NewRequest(http.MethodPost, "/scheduling/cancel/"+conf.AppointmentID, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(conf.AppointmentID)

	err := h.CancelAppointment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result BookingConfirmation
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result.Status != "cancelled" {
		t.Errorf("expected status cancelled, got %s", result.Status)
	}
}

func TestSelfScheduleHandler_CancelAppointment_NotFound(t *testing.T) {
	h, _, e := newTestHandler()

	body := `{"patient_id":"patient-1","reason":"no reason"}`
	req := httptest.NewRequest(http.MethodPost, "/scheduling/cancel/nonexistent", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.CancelAppointment(c)
	if err == nil {
		t.Error("expected error for not found")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %v", err)
	}
}

func TestSelfScheduleHandler_ListAppointments(t *testing.T) {
	h, mgr, e := newTestHandler()
	seedSlots(mgr)

	mgr.BookAppointment(nil, BookingRequest{SlotID: "slot-1", PatientID: "patient-1"})
	mgr.BookAppointment(nil, BookingRequest{SlotID: "slot-2", PatientID: "patient-1"})
	mgr.BookAppointment(nil, BookingRequest{SlotID: "slot-3", PatientID: "patient-2"})

	req := httptest.NewRequest(http.MethodGet, "/scheduling/appointments?patient_id=patient-1&status=booked&limit=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListAppointments(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var appointments []BookingConfirmation
	if err := json.Unmarshal(rec.Body.Bytes(), &appointments); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(appointments) != 2 {
		t.Errorf("expected 2 appointments for patient-1, got %d", len(appointments))
	}
}

func TestSelfScheduleHandler_GetAppointment(t *testing.T) {
	h, mgr, e := newTestHandler()
	seedSlots(mgr)

	conf, _ := mgr.BookAppointment(nil, BookingRequest{SlotID: "slot-1", PatientID: "patient-1"})

	req := httptest.NewRequest(http.MethodGet, "/scheduling/appointments/"+conf.AppointmentID+"?patient_id=patient-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(conf.AppointmentID)

	err := h.GetAppointment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result BookingConfirmation
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result.AppointmentID != conf.AppointmentID {
		t.Errorf("expected appointment ID %s, got %s", conf.AppointmentID, result.AppointmentID)
	}
	if result.PatientID != "patient-1" {
		t.Errorf("expected patient_id patient-1, got %s", result.PatientID)
	}
}
