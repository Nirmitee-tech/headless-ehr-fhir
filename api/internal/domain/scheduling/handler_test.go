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
		"GET:/api/v1/schedules/:id",
		"POST:/api/v1/slots",
		"POST:/api/v1/appointments",
		"GET:/api/v1/appointments/:id",
		"POST:/api/v1/waitlist",
		"GET:/fhir/Schedule",
		"GET:/fhir/Slot",
		"GET:/fhir/Appointment",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
