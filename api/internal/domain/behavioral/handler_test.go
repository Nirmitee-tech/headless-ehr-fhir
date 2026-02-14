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

func TestHandler_ListPsychAssessments(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListPsychAssessments(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_ListPsychAssessments_ByPatient(t *testing.T) {
	h, e := newTestHandler()
	patientID := uuid.New()
	a := &PsychiatricAssessment{PatientID: patientID, EncounterID: uuid.New(), AssessorID: uuid.New()}
	h.svc.CreatePsychAssessment(nil, a)
	req := httptest.NewRequest(http.MethodGet, "/?patient_id="+patientID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListPsychAssessments(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_UpdatePsychAssessment(t *testing.T) {
	h, e := newTestHandler()
	a := &PsychiatricAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), AssessorID: uuid.New()}
	h.svc.CreatePsychAssessment(nil, a)
	body := `{"chief_complaint":"updated"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())
	if err := h.UpdatePsychAssessment(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_GetSafetyPlan(t *testing.T) {
	h, e := newTestHandler()
	sp := &SafetyPlan{PatientID: uuid.New(), CreatedByID: uuid.New()}
	h.svc.CreateSafetyPlan(nil, sp)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sp.ID.String())
	if err := h.GetSafetyPlan(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_GetSafetyPlan_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetSafetyPlan(c)
	if err == nil { t.Error("expected error for not found") }
}

func TestHandler_ListSafetyPlans(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListSafetyPlans(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_ListSafetyPlans_ByPatient(t *testing.T) {
	h, e := newTestHandler()
	patientID := uuid.New()
	sp := &SafetyPlan{PatientID: patientID, CreatedByID: uuid.New()}
	h.svc.CreateSafetyPlan(nil, sp)
	req := httptest.NewRequest(http.MethodGet, "/?patient_id="+patientID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListSafetyPlans(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_UpdateSafetyPlan(t *testing.T) {
	h, e := newTestHandler()
	sp := &SafetyPlan{PatientID: uuid.New(), CreatedByID: uuid.New()}
	h.svc.CreateSafetyPlan(nil, sp)
	body := `{"warning_signs":"updated signs"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sp.ID.String())
	if err := h.UpdateSafetyPlan(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_DeleteSafetyPlan(t *testing.T) {
	h, e := newTestHandler()
	sp := &SafetyPlan{PatientID: uuid.New(), CreatedByID: uuid.New()}
	h.svc.CreateSafetyPlan(nil, sp)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sp.ID.String())
	if err := h.DeleteSafetyPlan(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusNoContent { t.Errorf("expected 204, got %d", rec.Code) }
}

func TestHandler_GetLegalHold(t *testing.T) {
	h, e := newTestHandler()
	lh := &LegalHold{PatientID: uuid.New(), InitiatedByID: uuid.New(), HoldType: "5150", Reason: "danger to self"}
	h.svc.CreateLegalHold(nil, lh)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(lh.ID.String())
	if err := h.GetLegalHold(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_GetLegalHold_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetLegalHold(c)
	if err == nil { t.Error("expected error for not found") }
}

func TestHandler_ListLegalHolds(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListLegalHolds(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_ListLegalHolds_ByPatient(t *testing.T) {
	h, e := newTestHandler()
	patientID := uuid.New()
	lh := &LegalHold{PatientID: patientID, InitiatedByID: uuid.New(), HoldType: "5150", Reason: "danger"}
	h.svc.CreateLegalHold(nil, lh)
	req := httptest.NewRequest(http.MethodGet, "/?patient_id="+patientID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListLegalHolds(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_UpdateLegalHold(t *testing.T) {
	h, e := newTestHandler()
	lh := &LegalHold{PatientID: uuid.New(), InitiatedByID: uuid.New(), HoldType: "5150", Reason: "danger to self"}
	h.svc.CreateLegalHold(nil, lh)
	body := `{"reason":"updated reason"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(lh.ID.String())
	if err := h.UpdateLegalHold(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_DeleteLegalHold(t *testing.T) {
	h, e := newTestHandler()
	lh := &LegalHold{PatientID: uuid.New(), InitiatedByID: uuid.New(), HoldType: "5150", Reason: "danger to self"}
	h.svc.CreateLegalHold(nil, lh)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(lh.ID.String())
	if err := h.DeleteLegalHold(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusNoContent { t.Errorf("expected 204, got %d", rec.Code) }
}

func TestHandler_GetSeclusionRestraint(t *testing.T) {
	h, e := newTestHandler()
	sr := &SeclusionRestraintEvent{PatientID: uuid.New(), OrderedByID: uuid.New(), EventType: "seclusion", Reason: "agitated"}
	h.svc.CreateSeclusionRestraint(nil, sr)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sr.ID.String())
	if err := h.GetSeclusionRestraint(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_GetSeclusionRestraint_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetSeclusionRestraint(c)
	if err == nil { t.Error("expected error for not found") }
}

func TestHandler_ListSeclusionRestraints(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListSeclusionRestraints(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_ListSeclusionRestraints_ByPatient(t *testing.T) {
	h, e := newTestHandler()
	patientID := uuid.New()
	sr := &SeclusionRestraintEvent{PatientID: patientID, OrderedByID: uuid.New(), EventType: "seclusion", Reason: "agitated"}
	h.svc.CreateSeclusionRestraint(nil, sr)
	req := httptest.NewRequest(http.MethodGet, "/?patient_id="+patientID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListSeclusionRestraints(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_UpdateSeclusionRestraint(t *testing.T) {
	h, e := newTestHandler()
	sr := &SeclusionRestraintEvent{PatientID: uuid.New(), OrderedByID: uuid.New(), EventType: "seclusion", Reason: "agitated"}
	h.svc.CreateSeclusionRestraint(nil, sr)
	body := `{"reason":"updated reason"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sr.ID.String())
	if err := h.UpdateSeclusionRestraint(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_DeleteSeclusionRestraint(t *testing.T) {
	h, e := newTestHandler()
	sr := &SeclusionRestraintEvent{PatientID: uuid.New(), OrderedByID: uuid.New(), EventType: "seclusion", Reason: "agitated"}
	h.svc.CreateSeclusionRestraint(nil, sr)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sr.ID.String())
	if err := h.DeleteSeclusionRestraint(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusNoContent { t.Errorf("expected 204, got %d", rec.Code) }
}

func TestHandler_GetGroupTherapySession(t *testing.T) {
	h, e := newTestHandler()
	gs := &GroupTherapySession{SessionName: "CBT Group", FacilitatorID: uuid.New()}
	h.svc.CreateGroupTherapySession(nil, gs)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gs.ID.String())
	if err := h.GetGroupTherapySession(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_GetGroupTherapySession_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetGroupTherapySession(c)
	if err == nil { t.Error("expected error for not found") }
}

func TestHandler_ListGroupTherapySessions(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListGroupTherapySessions(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_UpdateGroupTherapySession(t *testing.T) {
	h, e := newTestHandler()
	gs := &GroupTherapySession{SessionName: "CBT Group", FacilitatorID: uuid.New()}
	h.svc.CreateGroupTherapySession(nil, gs)
	body := `{"session_name":"Updated CBT Group"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gs.ID.String())
	if err := h.UpdateGroupTherapySession(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_DeleteGroupTherapySession(t *testing.T) {
	h, e := newTestHandler()
	gs := &GroupTherapySession{SessionName: "CBT Group", FacilitatorID: uuid.New()}
	h.svc.CreateGroupTherapySession(nil, gs)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gs.ID.String())
	if err := h.DeleteGroupTherapySession(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusNoContent { t.Errorf("expected 204, got %d", rec.Code) }
}

func TestHandler_GetAttendance(t *testing.T) {
	h, e := newTestHandler()
	sessionID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sessionID.String())
	if err := h.GetAttendance(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_RegisterRoutes(t *testing.T) {
	h, e := newTestHandler()
	api := e.Group("/api/v1")
	fhir := e.Group("/fhir")
	h.RegisterRoutes(api, fhir)
	routes := e.Routes()
	if len(routes) == 0 { t.Error("expected routes to be registered") }
	routePaths := make(map[string]bool)
	for _, r := range routes { routePaths[r.Method+":"+r.Path] = true }
	expected := []string{
		"GET:/api/v1/psychiatric-assessments",
		"GET:/api/v1/psychiatric-assessments/:id",
		"POST:/api/v1/psychiatric-assessments",
		"PUT:/api/v1/psychiatric-assessments/:id",
		"DELETE:/api/v1/psychiatric-assessments/:id",
		"GET:/api/v1/safety-plans",
		"GET:/api/v1/safety-plans/:id",
		"POST:/api/v1/safety-plans",
		"PUT:/api/v1/safety-plans/:id",
		"DELETE:/api/v1/safety-plans/:id",
		"GET:/api/v1/legal-holds",
		"GET:/api/v1/legal-holds/:id",
		"POST:/api/v1/legal-holds",
		"PUT:/api/v1/legal-holds/:id",
		"DELETE:/api/v1/legal-holds/:id",
		"GET:/api/v1/seclusion-restraints",
		"GET:/api/v1/seclusion-restraints/:id",
		"POST:/api/v1/seclusion-restraints",
		"PUT:/api/v1/seclusion-restraints/:id",
		"DELETE:/api/v1/seclusion-restraints/:id",
		"GET:/api/v1/group-therapy-sessions",
		"GET:/api/v1/group-therapy-sessions/:id",
		"POST:/api/v1/group-therapy-sessions",
		"PUT:/api/v1/group-therapy-sessions/:id",
		"DELETE:/api/v1/group-therapy-sessions/:id",
		"POST:/api/v1/group-therapy-sessions/:id/attendance",
		"GET:/api/v1/group-therapy-sessions/:id/attendance",
	}
	for _, path := range expected {
		if !routePaths[path] { t.Errorf("missing expected route: %s", path) }
	}
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
