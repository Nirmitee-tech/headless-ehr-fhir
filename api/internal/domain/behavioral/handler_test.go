package behavioral

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func newTestHandler() (*Handler, *echo.Echo) {
	svc := newTestService()
	h := NewHandler(svc)
	e := echo.New()
	return h, e
}

func TestHandler_CreatePsychAssessment(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","encounter_id":"` + uuid.New().String() + `","assessor_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreatePsychAssessment(c)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusCreated { t.Errorf("expected 201, got %d", rec.Code) }
}

func TestHandler_CreatePsychAssessment_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreatePsychAssessment(c)
	if err == nil { t.Error("expected error for missing required fields") }
}

func TestHandler_GetPsychAssessment(t *testing.T) {
	h, e := newTestHandler()
	a := &PsychiatricAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), AssessorID: uuid.New()}
	h.svc.CreatePsychAssessment(nil, a)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())
	err := h.GetPsychAssessment(c)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_GetPsychAssessment_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetPsychAssessment(c)
	if err == nil { t.Error("expected error for not found") }
}

func TestHandler_CreateSafetyPlan(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","created_by_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateSafetyPlan(c)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusCreated { t.Errorf("expected 201, got %d", rec.Code) }
}

func TestHandler_CreateLegalHold(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","initiated_by_id":"` + uuid.New().String() + `","hold_type":"5150","reason":"danger to self"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateLegalHold(c)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusCreated { t.Errorf("expected 201, got %d", rec.Code) }
}

func TestHandler_CreateLegalHold_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateLegalHold(c)
	if err == nil { t.Error("expected error for missing required fields") }
}

func TestHandler_CreateSeclusionRestraint(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","ordered_by_id":"` + uuid.New().String() + `","event_type":"seclusion","reason":"agitated"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateSeclusionRestraint(c)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusCreated { t.Errorf("expected 201, got %d", rec.Code) }
}

func TestHandler_CreateGroupTherapySession(t *testing.T) {
	h, e := newTestHandler()
	body := `{"session_name":"CBT Group","facilitator_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateGroupTherapySession(c)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusCreated { t.Errorf("expected 201, got %d", rec.Code) }
}

func TestHandler_CreateGroupTherapySession_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"session_name":"CBT Group"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateGroupTherapySession(c)
	if err == nil { t.Error("expected error for missing facilitator_id") }
}

func TestHandler_AddAttendance(t *testing.T) {
	h, e := newTestHandler()
	sessionID := uuid.New()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sessionID.String())
	err := h.AddAttendance(c)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusCreated { t.Errorf("expected 201, got %d", rec.Code) }
}

func TestHandler_DeletePsychAssessment(t *testing.T) {
	h, e := newTestHandler()
	a := &PsychiatricAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), AssessorID: uuid.New()}
	h.svc.CreatePsychAssessment(nil, a)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())
	err := h.DeletePsychAssessment(c)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusNoContent { t.Errorf("expected 204, got %d", rec.Code) }
}
