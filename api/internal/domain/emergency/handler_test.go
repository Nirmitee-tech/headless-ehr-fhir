package emergency

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

func TestHandler_CreateTriageRecord(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","encounter_id":"` + uuid.New().String() + `","triage_nurse_id":"` + uuid.New().String() + `","chief_complaint":"chest pain"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateTriageRecord(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateTriageRecord_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateTriageRecord(c)
	if err == nil {
		t.Error("expected error for missing required fields")
	}
}

func TestHandler_GetTriageRecord(t *testing.T) {
	h, e := newTestHandler()
	tr := &TriageRecord{PatientID: uuid.New(), EncounterID: uuid.New(), TriageNurseID: uuid.New(), ChiefComplaint: "pain"}
	h.svc.CreateTriageRecord(nil, tr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tr.ID.String())

	err := h.GetTriageRecord(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetTriageRecord_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetTriageRecord(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_CreateEDTracking(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","encounter_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateEDTracking(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateTraumaActivation(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","activation_level":"level-1"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateTraumaActivation(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_DeleteTriageRecord(t *testing.T) {
	h, e := newTestHandler()
	tr := &TriageRecord{PatientID: uuid.New(), EncounterID: uuid.New(), TriageNurseID: uuid.New(), ChiefComplaint: "pain"}
	h.svc.CreateTriageRecord(nil, tr)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tr.ID.String())

	err := h.DeleteTriageRecord(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_ListTriageRecords(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListTriageRecords(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_ListTriageRecords_ByPatient(t *testing.T) {
	h, e := newTestHandler()
	patientID := uuid.New()
	tr := &TriageRecord{PatientID: patientID, EncounterID: uuid.New(), TriageNurseID: uuid.New(), ChiefComplaint: "pain"}
	h.svc.CreateTriageRecord(nil, tr)
	req := httptest.NewRequest(http.MethodGet, "/?patient_id="+patientID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListTriageRecords(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_UpdateTriageRecord(t *testing.T) {
	h, e := newTestHandler()
	tr := &TriageRecord{PatientID: uuid.New(), EncounterID: uuid.New(), TriageNurseID: uuid.New(), ChiefComplaint: "pain"}
	h.svc.CreateTriageRecord(nil, tr)
	body := `{"chief_complaint":"severe pain"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tr.ID.String())
	if err := h.UpdateTriageRecord(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_GetEDTracking(t *testing.T) {
	h, e := newTestHandler()
	ed := &EDTracking{PatientID: uuid.New(), EncounterID: uuid.New()}
	h.svc.CreateEDTracking(nil, ed)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ed.ID.String())
	if err := h.GetEDTracking(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_GetEDTracking_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetEDTracking(c)
	if err == nil { t.Error("expected error for not found") }
}

func TestHandler_ListEDTrackings(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListEDTrackings(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_ListEDTrackings_ByPatient(t *testing.T) {
	h, e := newTestHandler()
	patientID := uuid.New()
	ed := &EDTracking{PatientID: patientID, EncounterID: uuid.New()}
	h.svc.CreateEDTracking(nil, ed)
	req := httptest.NewRequest(http.MethodGet, "/?patient_id="+patientID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListEDTrackings(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_UpdateEDTracking(t *testing.T) {
	h, e := newTestHandler()
	ed := &EDTracking{PatientID: uuid.New(), EncounterID: uuid.New()}
	h.svc.CreateEDTracking(nil, ed)
	body := `{"current_status":"triaged"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ed.ID.String())
	if err := h.UpdateEDTracking(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_DeleteEDTracking(t *testing.T) {
	h, e := newTestHandler()
	ed := &EDTracking{PatientID: uuid.New(), EncounterID: uuid.New()}
	h.svc.CreateEDTracking(nil, ed)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ed.ID.String())
	if err := h.DeleteEDTracking(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusNoContent { t.Errorf("expected 204, got %d", rec.Code) }
}

func TestHandler_GetEDStatusHistory(t *testing.T) {
	h, e := newTestHandler()
	trackingID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(trackingID.String())
	if err := h.GetEDStatusHistory(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_GetTraumaActivation(t *testing.T) {
	h, e := newTestHandler()
	ta := &TraumaActivation{PatientID: uuid.New(), ActivationLevel: "level-1"}
	h.svc.CreateTraumaActivation(nil, ta)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ta.ID.String())
	if err := h.GetTraumaActivation(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_GetTraumaActivation_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetTraumaActivation(c)
	if err == nil { t.Error("expected error for not found") }
}

func TestHandler_ListTraumaActivations(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListTraumaActivations(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_ListTraumaActivations_ByPatient(t *testing.T) {
	h, e := newTestHandler()
	patientID := uuid.New()
	ta := &TraumaActivation{PatientID: patientID, ActivationLevel: "level-1"}
	h.svc.CreateTraumaActivation(nil, ta)
	req := httptest.NewRequest(http.MethodGet, "/?patient_id="+patientID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListTraumaActivations(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_UpdateTraumaActivation(t *testing.T) {
	h, e := newTestHandler()
	ta := &TraumaActivation{PatientID: uuid.New(), ActivationLevel: "level-1"}
	h.svc.CreateTraumaActivation(nil, ta)
	body := `{"activation_level":"level-2"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ta.ID.String())
	if err := h.UpdateTraumaActivation(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_DeleteTraumaActivation(t *testing.T) {
	h, e := newTestHandler()
	ta := &TraumaActivation{PatientID: uuid.New(), ActivationLevel: "level-1"}
	h.svc.CreateTraumaActivation(nil, ta)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ta.ID.String())
	if err := h.DeleteTraumaActivation(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusNoContent { t.Errorf("expected 204, got %d", rec.Code) }
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
		"GET:/api/v1/triage-records",
		"GET:/api/v1/triage-records/:id",
		"POST:/api/v1/triage-records",
		"PUT:/api/v1/triage-records/:id",
		"DELETE:/api/v1/triage-records/:id",
		"GET:/api/v1/ed-tracking",
		"GET:/api/v1/ed-tracking/:id",
		"POST:/api/v1/ed-tracking",
		"PUT:/api/v1/ed-tracking/:id",
		"DELETE:/api/v1/ed-tracking/:id",
		"POST:/api/v1/ed-tracking/:id/status-history",
		"GET:/api/v1/ed-tracking/:id/status-history",
		"GET:/api/v1/trauma-activations",
		"GET:/api/v1/trauma-activations/:id",
		"POST:/api/v1/trauma-activations",
		"PUT:/api/v1/trauma-activations/:id",
		"DELETE:/api/v1/trauma-activations/:id",
	}
	for _, path := range expected {
		if !routePaths[path] { t.Errorf("missing expected route: %s", path) }
	}
}

func TestHandler_AddEDStatusHistory(t *testing.T) {
	h, e := newTestHandler()
	trackingID := uuid.New()
	body := `{"status":"triaged"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(trackingID.String())

	err := h.AddEDStatusHistory(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}
