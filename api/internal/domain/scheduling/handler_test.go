package scheduling

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func newTestHandler() (*Handler, *echo.Echo) {
	svc := newTestService()
	h := NewHandler(svc)
	e := echo.New()
	return h, e
}

func TestHandler_CreateSchedule(t *testing.T) {
	h, e := newTestHandler()
	body := `{"practitioner_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateSchedule(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateSchedule_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateSchedule(c)
	if err == nil {
		t.Error("expected error for missing practitioner_id")
	}
}

func TestHandler_GetSchedule(t *testing.T) {
	h, e := newTestHandler()
	s := &Schedule{PractitionerID: uuid.New()}
	h.svc.CreateSchedule(nil, s)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.ID.String())

	err := h.GetSchedule(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetSchedule_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetSchedule(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_GetSchedule_InvalidID(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.GetSchedule(c)
	if err == nil {
		t.Error("expected error for invalid id")
	}
}

func TestHandler_DeleteSchedule(t *testing.T) {
	h, e := newTestHandler()
	s := &Schedule{PractitionerID: uuid.New()}
	h.svc.CreateSchedule(nil, s)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.ID.String())

	err := h.DeleteSchedule(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_CreateSlot(t *testing.T) {
	h, e := newTestHandler()
	start := time.Now().Format(time.RFC3339)
	end := time.Now().Add(30 * time.Minute).Format(time.RFC3339)
	body := `{"schedule_id":"` + uuid.New().String() + `","start_time":"` + start + `","end_time":"` + end + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateSlot(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateSlot_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"start_time":"` + time.Now().Format(time.RFC3339) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateSlot(c)
	if err == nil {
		t.Error("expected error for missing schedule_id")
	}
}

func TestHandler_GetSlot(t *testing.T) {
	h, e := newTestHandler()
	sl := &Slot{ScheduleID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)}
	h.svc.CreateSlot(nil, sl)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sl.ID.String())

	err := h.GetSlot(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteSlot(t *testing.T) {
	h, e := newTestHandler()
	sl := &Slot{ScheduleID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)}
	h.svc.CreateSlot(nil, sl)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sl.ID.String())

	err := h.DeleteSlot(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_CreateAppointment(t *testing.T) {
	h, e := newTestHandler()
	start := time.Now().Format(time.RFC3339)
	body := `{"patient_id":"` + uuid.New().String() + `","start_time":"` + start + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateAppointment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateAppointment_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"start_time":"` + time.Now().Format(time.RFC3339) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateAppointment(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetAppointment(t *testing.T) {
	h, e := newTestHandler()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	h.svc.CreateAppointment(nil, a)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())

	err := h.GetAppointment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteAppointment(t *testing.T) {
	h, e := newTestHandler()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	h.svc.CreateAppointment(nil, a)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())

	err := h.DeleteAppointment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddParticipant(t *testing.T) {
	h, e := newTestHandler()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	h.svc.CreateAppointment(nil, a)

	body := `{"actor_type":"Practitioner","actor_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())

	err := h.AddParticipant(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var result AppointmentParticipant
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result.AppointmentID != a.ID {
		t.Error("expected appointment_id to match")
	}
}

func TestHandler_CreateWaitlistEntry(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateWaitlistEntry(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetWaitlistEntry(t *testing.T) {
	h, e := newTestHandler()
	w := &Waitlist{PatientID: uuid.New()}
	h.svc.CreateWaitlistEntry(nil, w)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(w.ID.String())

	err := h.GetWaitlistEntry(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteWaitlistEntry(t *testing.T) {
	h, e := newTestHandler()
	w := &Waitlist{PatientID: uuid.New()}
	h.svc.CreateWaitlistEntry(nil, w)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(w.ID.String())

	err := h.DeleteWaitlistEntry(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// -- List/Update Tests --

func TestHandler_ListSchedules(t *testing.T) {
	h, e := newTestHandler()
	s := &Schedule{PractitionerID: uuid.New()}
	h.svc.CreateSchedule(nil, s)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListSchedules(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateSchedule(t *testing.T) {
	h, e := newTestHandler()
	s := &Schedule{PractitionerID: uuid.New()}
	h.svc.CreateSchedule(nil, s)
	body := `{"practitioner_id":"` + s.PractitionerID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.ID.String())
	err := h.UpdateSchedule(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListSlots(t *testing.T) {
	h, e := newTestHandler()
	sl := &Slot{ScheduleID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)}
	h.svc.CreateSlot(nil, sl)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListSlots(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateSlot(t *testing.T) {
	h, e := newTestHandler()
	sl := &Slot{ScheduleID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)}
	h.svc.CreateSlot(nil, sl)
	start := time.Now().Format(time.RFC3339)
	end := time.Now().Add(60 * time.Minute).Format(time.RFC3339)
	body := `{"schedule_id":"` + sl.ScheduleID.String() + `","start_time":"` + start + `","end_time":"` + end + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sl.ID.String())
	err := h.UpdateSlot(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListAppointments(t *testing.T) {
	h, e := newTestHandler()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	h.svc.CreateAppointment(nil, a)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListAppointments(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateAppointment(t *testing.T) {
	h, e := newTestHandler()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	h.svc.CreateAppointment(nil, a)
	body := `{"patient_id":"` + a.PatientID.String() + `","start_time":"` + start.Format(time.RFC3339) + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())
	err := h.UpdateAppointment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetParticipants(t *testing.T) {
	h, e := newTestHandler()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	h.svc.CreateAppointment(nil, a)
	p := &AppointmentParticipant{AppointmentID: a.ID, ActorType: "Practitioner", ActorID: uuid.New()}
	h.svc.AddAppointmentParticipant(nil, p)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())
	err := h.GetParticipants(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateWaitlistEntry(t *testing.T) {
	h, e := newTestHandler()
	w := &Waitlist{PatientID: uuid.New()}
	h.svc.CreateWaitlistEntry(nil, w)
	body := `{"patient_id":"` + w.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(w.ID.String())
	err := h.UpdateWaitlistEntry(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- FHIR Schedule Endpoints --

func TestHandler_SearchSchedulesFHIR(t *testing.T) {
	h, e := newTestHandler()
	s := &Schedule{PractitionerID: uuid.New()}
	h.svc.CreateSchedule(nil, s)
	req := httptest.NewRequest(http.MethodGet, "/fhir/Schedule", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchSchedulesFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Bundle") {
		t.Error("expected Bundle in response")
	}
}

func TestHandler_GetScheduleFHIR(t *testing.T) {
	h, e := newTestHandler()
	s := &Schedule{PractitionerID: uuid.New()}
	h.svc.CreateSchedule(nil, s)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.FHIRID)
	err := h.GetScheduleFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetScheduleFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	_ = h.GetScheduleFHIR(c)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateScheduleFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"practitioner_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateScheduleFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc == "" || !strings.Contains(loc, "/fhir/Schedule/") {
		t.Errorf("expected Location header, got %q", loc)
	}
}

func TestHandler_UpdateScheduleFHIR(t *testing.T) {
	h, e := newTestHandler()
	s := &Schedule{PractitionerID: uuid.New()}
	h.svc.CreateSchedule(nil, s)
	body := `{"practitioner_id":"` + s.PractitionerID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.FHIRID)
	err := h.UpdateScheduleFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteScheduleFHIR(t *testing.T) {
	h, e := newTestHandler()
	s := &Schedule{PractitionerID: uuid.New()}
	h.svc.CreateSchedule(nil, s)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.FHIRID)
	err := h.DeleteScheduleFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchScheduleFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	s := &Schedule{PractitionerID: uuid.New()}
	h.svc.CreateSchedule(nil, s)
	body := `{"comment":"Updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.FHIRID)
	err := h.PatchScheduleFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadScheduleFHIR(t *testing.T) {
	h, e := newTestHandler()
	s := &Schedule{PractitionerID: uuid.New()}
	h.svc.CreateSchedule(nil, s)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(s.FHIRID, "1")
	err := h.VreadScheduleFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryScheduleFHIR(t *testing.T) {
	h, e := newTestHandler()
	s := &Schedule{PractitionerID: uuid.New()}
	h.svc.CreateSchedule(nil, s)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.FHIRID)
	err := h.HistoryScheduleFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- FHIR Slot Endpoints --

func TestHandler_SearchSlotsFHIR(t *testing.T) {
	h, e := newTestHandler()
	sl := &Slot{ScheduleID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)}
	h.svc.CreateSlot(nil, sl)
	req := httptest.NewRequest(http.MethodGet, "/fhir/Slot", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchSlotsFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetSlotFHIR(t *testing.T) {
	h, e := newTestHandler()
	sl := &Slot{ScheduleID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)}
	h.svc.CreateSlot(nil, sl)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sl.FHIRID)
	err := h.GetSlotFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_CreateSlotFHIR(t *testing.T) {
	h, e := newTestHandler()
	start := time.Now().Format(time.RFC3339)
	end := time.Now().Add(30 * time.Minute).Format(time.RFC3339)
	body := `{"schedule_id":"` + uuid.New().String() + `","start_time":"` + start + `","end_time":"` + end + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateSlotFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_UpdateSlotFHIR(t *testing.T) {
	h, e := newTestHandler()
	sl := &Slot{ScheduleID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)}
	h.svc.CreateSlot(nil, sl)
	start := time.Now().Format(time.RFC3339)
	end := time.Now().Add(60 * time.Minute).Format(time.RFC3339)
	body := `{"schedule_id":"` + sl.ScheduleID.String() + `","start_time":"` + start + `","end_time":"` + end + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sl.FHIRID)
	err := h.UpdateSlotFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteSlotFHIR(t *testing.T) {
	h, e := newTestHandler()
	sl := &Slot{ScheduleID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)}
	h.svc.CreateSlot(nil, sl)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sl.FHIRID)
	err := h.DeleteSlotFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchSlotFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	sl := &Slot{ScheduleID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)}
	h.svc.CreateSlot(nil, sl)
	body := `{"status":"busy"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sl.FHIRID)
	err := h.PatchSlotFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadSlotFHIR(t *testing.T) {
	h, e := newTestHandler()
	sl := &Slot{ScheduleID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)}
	h.svc.CreateSlot(nil, sl)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(sl.FHIRID, "1")
	err := h.VreadSlotFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistorySlotFHIR(t *testing.T) {
	h, e := newTestHandler()
	sl := &Slot{ScheduleID: uuid.New(), StartTime: time.Now(), EndTime: time.Now().Add(30 * time.Minute)}
	h.svc.CreateSlot(nil, sl)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sl.FHIRID)
	err := h.HistorySlotFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- FHIR Appointment Endpoints --

func TestHandler_SearchAppointmentsFHIR(t *testing.T) {
	h, e := newTestHandler()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	h.svc.CreateAppointment(nil, a)
	req := httptest.NewRequest(http.MethodGet, "/fhir/Appointment", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchAppointmentsFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetAppointmentFHIR(t *testing.T) {
	h, e := newTestHandler()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	h.svc.CreateAppointment(nil, a)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.FHIRID)
	err := h.GetAppointmentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetAppointmentFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	_ = h.GetAppointmentFHIR(c)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateAppointmentFHIR(t *testing.T) {
	h, e := newTestHandler()
	start := time.Now().Format(time.RFC3339)
	body := `{"patient_id":"` + uuid.New().String() + `","start_time":"` + start + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateAppointmentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc == "" || !strings.Contains(loc, "/fhir/Appointment/") {
		t.Errorf("expected Location header, got %q", loc)
	}
}

func TestHandler_UpdateAppointmentFHIR(t *testing.T) {
	h, e := newTestHandler()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	h.svc.CreateAppointment(nil, a)
	body := `{"patient_id":"` + a.PatientID.String() + `","start_time":"` + start.Format(time.RFC3339) + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.FHIRID)
	err := h.UpdateAppointmentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteAppointmentFHIR(t *testing.T) {
	h, e := newTestHandler()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	h.svc.CreateAppointment(nil, a)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.FHIRID)
	err := h.DeleteAppointmentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchAppointmentFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	h.svc.CreateAppointment(nil, a)
	body := `{"status":"cancelled"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.FHIRID)
	err := h.PatchAppointmentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadAppointmentFHIR(t *testing.T) {
	h, e := newTestHandler()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	h.svc.CreateAppointment(nil, a)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(a.FHIRID, "1")
	err := h.VreadAppointmentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryAppointmentFHIR(t *testing.T) {
	h, e := newTestHandler()
	start := time.Now()
	a := &Appointment{PatientID: uuid.New(), StartTime: &start}
	h.svc.CreateAppointment(nil, a)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.FHIRID)
	err := h.HistoryAppointmentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- Route Registration --

func TestHandler_RegisterRoutes(t *testing.T) {
	h, e := newTestHandler()
	api := e.Group("/api/v1")
	fhir := e.Group("/fhir")
	h.RegisterRoutes(api, fhir)

	routes := e.Routes()
	if len(routes) == 0 {
		t.Error("expected routes to be registered")
	}
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}
	expected := []string{
		"POST:/api/v1/schedules",
		"GET:/api/v1/schedules",
		"GET:/api/v1/schedules/:id",
		"PUT:/api/v1/schedules/:id",
		"DELETE:/api/v1/schedules/:id",
		"POST:/api/v1/slots",
		"GET:/api/v1/slots",
		"GET:/api/v1/slots/:id",
		"PUT:/api/v1/slots/:id",
		"POST:/api/v1/appointments",
		"GET:/api/v1/appointments",
		"GET:/api/v1/appointments/:id",
		"PUT:/api/v1/appointments/:id",
		"GET:/api/v1/appointments/:id/participants",
		"POST:/api/v1/waitlist",
		"GET:/api/v1/waitlist/:id",
		"PUT:/api/v1/waitlist/:id",
		"GET:/fhir/Schedule",
		"GET:/fhir/Schedule/:id",
		"POST:/fhir/Schedule",
		"PUT:/fhir/Schedule/:id",
		"DELETE:/fhir/Schedule/:id",
		"PATCH:/fhir/Schedule/:id",
		"GET:/fhir/Slot",
		"GET:/fhir/Slot/:id",
		"POST:/fhir/Slot",
		"GET:/fhir/Appointment",
		"GET:/fhir/Appointment/:id",
		"POST:/fhir/Appointment",
		"PUT:/fhir/Appointment/:id",
		"DELETE:/fhir/Appointment/:id",
		"PATCH:/fhir/Appointment/:id",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
