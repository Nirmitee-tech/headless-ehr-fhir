package clinical

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func newTestHandler() (*Handler, *echo.Echo) {
	svc := newTestService()
	h := NewHandler(svc, nil)
	e := echo.New()
	return h, e
}

// ── Condition Handlers ──

func TestHandler_CreateCondition(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","code_value":"J06.9","code_display":"URI"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateCondition(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateCondition_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"code_display":"URI"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateCondition(c)
	if err == nil {
		t.Error("expected error for missing required fields")
	}
}

func TestHandler_GetCondition(t *testing.T) {
	h, e := newTestHandler()
	cond := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"}
	h.svc.CreateCondition(nil, cond)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cond.ID.String())
	err := h.GetCondition(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetCondition_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetCondition(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_GetCondition_InvalidID(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")
	err := h.GetCondition(c)
	if err == nil {
		t.Error("expected error for invalid id")
	}
}

func TestHandler_ListConditions(t *testing.T) {
	h, e := newTestHandler()
	pid := uuid.New()
	h.svc.CreateCondition(nil, &Condition{PatientID: pid, CodeValue: "J06.9", CodeDisplay: "URI"})
	h.svc.CreateCondition(nil, &Condition{PatientID: pid, CodeValue: "E11.9", CodeDisplay: "T2DM"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListConditions(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateCondition(t *testing.T) {
	h, e := newTestHandler()
	cond := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"}
	h.svc.CreateCondition(nil, cond)

	body := `{"clinical_status":"resolved","code_value":"J06.9","code_display":"URI"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cond.ID.String())
	err := h.UpdateCondition(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateCondition_InvalidID(t *testing.T) {
	h, e := newTestHandler()
	body := `{"clinical_status":"resolved"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")
	err := h.UpdateCondition(c)
	if err == nil {
		t.Error("expected error for invalid id")
	}
}

func TestHandler_DeleteCondition(t *testing.T) {
	h, e := newTestHandler()
	cond := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"}
	h.svc.CreateCondition(nil, cond)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cond.ID.String())
	err := h.DeleteCondition(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// ── Observation Handlers ──

func TestHandler_CreateObservation(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","code_value":"8310-5","code_display":"Body temp"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateObservation(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateObservation_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"code_display":"Body temp"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateObservation(c)
	if err == nil {
		t.Error("expected error for missing required fields")
	}
}

func TestHandler_GetObservation(t *testing.T) {
	h, e := newTestHandler()
	obs := &Observation{PatientID: uuid.New(), CodeValue: "8310-5", CodeDisplay: "Body temp"}
	h.svc.CreateObservation(nil, obs)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(obs.ID.String())
	err := h.GetObservation(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetObservation_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetObservation(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_ListObservations(t *testing.T) {
	h, e := newTestHandler()
	pid := uuid.New()
	h.svc.CreateObservation(nil, &Observation{PatientID: pid, CodeValue: "8310-5", CodeDisplay: "Body temp"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListObservations(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateObservation(t *testing.T) {
	h, e := newTestHandler()
	obs := &Observation{PatientID: uuid.New(), CodeValue: "8310-5", CodeDisplay: "Body temp"}
	h.svc.CreateObservation(nil, obs)

	body := `{"status":"amended"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(obs.ID.String())
	err := h.UpdateObservation(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteObservation(t *testing.T) {
	h, e := newTestHandler()
	obs := &Observation{PatientID: uuid.New(), CodeValue: "8310-5", CodeDisplay: "Body temp"}
	h.svc.CreateObservation(nil, obs)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(obs.ID.String())
	err := h.DeleteObservation(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddObservationComponent(t *testing.T) {
	h, e := newTestHandler()
	obs := &Observation{PatientID: uuid.New(), CodeValue: "85354-9", CodeDisplay: "BP"}
	h.svc.CreateObservation(nil, obs)

	body := `{"code_value":"8480-6","code_display":"Systolic"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(obs.ID.String())
	err := h.AddObservationComponent(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetObservationComponents(t *testing.T) {
	h, e := newTestHandler()
	obs := &Observation{PatientID: uuid.New(), CodeValue: "85354-9", CodeDisplay: "BP"}
	h.svc.CreateObservation(nil, obs)
	h.svc.AddObservationComponent(nil, &ObservationComponent{ObservationID: obs.ID, CodeValue: "8480-6", CodeDisplay: "Systolic"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(obs.ID.String())
	err := h.GetObservationComponents(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// ── Allergy Handlers ──

func TestHandler_CreateAllergy(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateAllergy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateAllergy_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateAllergy(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetAllergy(t *testing.T) {
	h, e := newTestHandler()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	h.svc.CreateAllergy(nil, a)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())
	err := h.GetAllergy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetAllergy_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetAllergy(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_ListAllergies(t *testing.T) {
	h, e := newTestHandler()
	pid := uuid.New()
	h.svc.CreateAllergy(nil, &AllergyIntolerance{PatientID: pid})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListAllergies(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateAllergy(t *testing.T) {
	h, e := newTestHandler()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	h.svc.CreateAllergy(nil, a)

	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())
	err := h.UpdateAllergy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteAllergy(t *testing.T) {
	h, e := newTestHandler()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	h.svc.CreateAllergy(nil, a)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())
	err := h.DeleteAllergy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddReaction(t *testing.T) {
	h, e := newTestHandler()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	h.svc.CreateAllergy(nil, a)

	body := `{"manifestation_code":"39579001","manifestation_display":"Urticaria"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())
	err := h.AddReaction(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetReactions(t *testing.T) {
	h, e := newTestHandler()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	h.svc.CreateAllergy(nil, a)
	h.svc.AddAllergyReaction(nil, &AllergyReaction{AllergyID: a.ID, ManifestationCode: "39579001", ManifestationDisplay: "Urticaria"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())
	err := h.GetReactions(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// ── Procedure Handlers ──

func TestHandler_CreateProcedure(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","code_value":"80146002","code_display":"Appendectomy"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateProcedure(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateProcedure_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"code_display":"Appendectomy"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateProcedure(c)
	if err == nil {
		t.Error("expected error for missing required fields")
	}
}

func TestHandler_GetProcedure(t *testing.T) {
	h, e := newTestHandler()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002", CodeDisplay: "Appendectomy"}
	h.svc.CreateProcedure(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())
	err := h.GetProcedure(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetProcedure_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetProcedure(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_ListProcedures(t *testing.T) {
	h, e := newTestHandler()
	pid := uuid.New()
	h.svc.CreateProcedure(nil, &ProcedureRecord{PatientID: pid, CodeValue: "80146002", CodeDisplay: "Appendectomy"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListProcedures(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateProcedure(t *testing.T) {
	h, e := newTestHandler()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002", CodeDisplay: "Appendectomy"}
	h.svc.CreateProcedure(nil, p)

	body := `{"status":"completed"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())
	err := h.UpdateProcedure(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteProcedure(t *testing.T) {
	h, e := newTestHandler()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002", CodeDisplay: "Appendectomy"}
	h.svc.CreateProcedure(nil, p)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())
	err := h.DeleteProcedure(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddPerformer(t *testing.T) {
	h, e := newTestHandler()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002", CodeDisplay: "Appendectomy"}
	h.svc.CreateProcedure(nil, p)

	body := `{"practitioner_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())
	err := h.AddPerformer(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetPerformers(t *testing.T) {
	h, e := newTestHandler()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002", CodeDisplay: "Appendectomy"}
	h.svc.CreateProcedure(nil, p)
	h.svc.AddProcedurePerformer(nil, &ProcedurePerformer{ProcedureID: p.ID, PractitionerID: uuid.New()})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())
	err := h.GetPerformers(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// ── FHIR Search Endpoints ──

func TestHandler_SearchConditionsFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateCondition(nil, &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Condition", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchConditionsFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle, got %v", bundle["resourceType"])
	}
}

func TestHandler_GetConditionFHIR(t *testing.T) {
	h, e := newTestHandler()
	cond := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"}
	h.svc.CreateCondition(nil, cond)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cond.FHIRID)
	err := h.GetConditionFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetConditionFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	err := h.GetConditionFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateConditionFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","code_value":"J06.9","code_display":"URI"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Condition", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateConditionFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc == "" {
		t.Error("expected Location header")
	}
}

func TestHandler_UpdateConditionFHIR(t *testing.T) {
	h, e := newTestHandler()
	cond := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"}
	h.svc.CreateCondition(nil, cond)

	body := `{"clinical_status":"resolved","code_value":"J06.9","code_display":"URI"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cond.FHIRID)
	err := h.UpdateConditionFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteConditionFHIR(t *testing.T) {
	h, e := newTestHandler()
	cond := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"}
	h.svc.CreateCondition(nil, cond)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cond.FHIRID)
	err := h.DeleteConditionFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchConditionFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	cond := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"}
	h.svc.CreateCondition(nil, cond)

	body := `{"clinicalStatus":{"coding":[{"code":"resolved"}]}}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cond.FHIRID)
	err := h.PatchConditionFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_PatchConditionFHIR_JSONPatch(t *testing.T) {
	h, e := newTestHandler()
	cond := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"}
	h.svc.CreateCondition(nil, cond)

	body := `[{"op":"replace","path":"/clinicalStatus/coding/0/code","value":"resolved"}]`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cond.FHIRID)
	err := h.PatchConditionFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_PatchConditionFHIR_UnsupportedMediaType(t *testing.T) {
	h, e := newTestHandler()
	cond := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"}
	h.svc.CreateCondition(nil, cond)

	body := `{}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cond.FHIRID)
	err := h.PatchConditionFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415, got %d", rec.Code)
	}
}

func TestHandler_VreadConditionFHIR(t *testing.T) {
	h, e := newTestHandler()
	cond := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"}
	h.svc.CreateCondition(nil, cond)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(cond.FHIRID, "1")
	err := h.VreadConditionFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryConditionFHIR(t *testing.T) {
	h, e := newTestHandler()
	cond := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"}
	h.svc.CreateCondition(nil, cond)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cond.FHIRID)
	err := h.HistoryConditionFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle["type"] != "history" {
		t.Errorf("expected history bundle type, got %v", bundle["type"])
	}
}

// ── FHIR Observation Endpoints ──

func TestHandler_SearchObservationsFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateObservation(nil, &Observation{PatientID: uuid.New(), CodeValue: "8310-5", CodeDisplay: "Body temp"})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchObservationsFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetObservationFHIR(t *testing.T) {
	h, e := newTestHandler()
	obs := &Observation{PatientID: uuid.New(), CodeValue: "8310-5", CodeDisplay: "Body temp"}
	h.svc.CreateObservation(nil, obs)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(obs.FHIRID)
	err := h.GetObservationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_CreateObservationFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","code_value":"8310-5","code_display":"Body temp"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Observation", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateObservationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_UpdateObservationFHIR(t *testing.T) {
	h, e := newTestHandler()
	obs := &Observation{PatientID: uuid.New(), CodeValue: "8310-5", CodeDisplay: "Body temp"}
	h.svc.CreateObservation(nil, obs)

	body := `{"status":"amended"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(obs.FHIRID)
	err := h.UpdateObservationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteObservationFHIR(t *testing.T) {
	h, e := newTestHandler()
	obs := &Observation{PatientID: uuid.New(), CodeValue: "8310-5", CodeDisplay: "Body temp"}
	h.svc.CreateObservation(nil, obs)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(obs.FHIRID)
	err := h.DeleteObservationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchObservationFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	obs := &Observation{PatientID: uuid.New(), CodeValue: "8310-5", CodeDisplay: "Body temp"}
	h.svc.CreateObservation(nil, obs)

	body := `{"status":"amended"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(obs.FHIRID)
	err := h.PatchObservationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadObservationFHIR(t *testing.T) {
	h, e := newTestHandler()
	obs := &Observation{PatientID: uuid.New(), CodeValue: "8310-5", CodeDisplay: "Body temp"}
	h.svc.CreateObservation(nil, obs)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(obs.FHIRID, "1")
	err := h.VreadObservationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryObservationFHIR(t *testing.T) {
	h, e := newTestHandler()
	obs := &Observation{PatientID: uuid.New(), CodeValue: "8310-5", CodeDisplay: "Body temp"}
	h.svc.CreateObservation(nil, obs)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(obs.FHIRID)
	err := h.HistoryObservationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// ── FHIR Allergy Endpoints ──

func TestHandler_SearchAllergiesFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateAllergy(nil, &AllergyIntolerance{PatientID: uuid.New()})

	req := httptest.NewRequest(http.MethodGet, "/fhir/AllergyIntolerance", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchAllergiesFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetAllergyFHIR(t *testing.T) {
	h, e := newTestHandler()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	h.svc.CreateAllergy(nil, a)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.FHIRID)
	err := h.GetAllergyFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_CreateAllergyFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/AllergyIntolerance", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateAllergyFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_UpdateAllergyFHIR(t *testing.T) {
	h, e := newTestHandler()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	h.svc.CreateAllergy(nil, a)

	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.FHIRID)
	err := h.UpdateAllergyFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteAllergyFHIR(t *testing.T) {
	h, e := newTestHandler()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	h.svc.CreateAllergy(nil, a)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.FHIRID)
	err := h.DeleteAllergyFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchAllergyFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	h.svc.CreateAllergy(nil, a)

	body := `{"clinicalStatus":{"coding":[{"code":"resolved"}]}}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.FHIRID)
	err := h.PatchAllergyFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadAllergyFHIR(t *testing.T) {
	h, e := newTestHandler()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	h.svc.CreateAllergy(nil, a)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(a.FHIRID, "1")
	err := h.VreadAllergyFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryAllergyFHIR(t *testing.T) {
	h, e := newTestHandler()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	h.svc.CreateAllergy(nil, a)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.FHIRID)
	err := h.HistoryAllergyFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// ── FHIR Procedure Endpoints ──

func TestHandler_SearchProceduresFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateProcedure(nil, &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002", CodeDisplay: "Appendectomy"})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Procedure", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchProceduresFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetProcedureFHIR(t *testing.T) {
	h, e := newTestHandler()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002", CodeDisplay: "Appendectomy"}
	h.svc.CreateProcedure(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.FHIRID)
	err := h.GetProcedureFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_CreateProcedureFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","code_value":"80146002","code_display":"Appendectomy"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Procedure", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateProcedureFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_UpdateProcedureFHIR(t *testing.T) {
	h, e := newTestHandler()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002", CodeDisplay: "Appendectomy"}
	h.svc.CreateProcedure(nil, p)

	body := `{"status":"completed"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.FHIRID)
	err := h.UpdateProcedureFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteProcedureFHIR(t *testing.T) {
	h, e := newTestHandler()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002", CodeDisplay: "Appendectomy"}
	h.svc.CreateProcedure(nil, p)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.FHIRID)
	err := h.DeleteProcedureFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchProcedureFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002", CodeDisplay: "Appendectomy"}
	h.svc.CreateProcedure(nil, p)

	body := `{"status":"completed"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.FHIRID)
	err := h.PatchProcedureFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadProcedureFHIR(t *testing.T) {
	h, e := newTestHandler()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002", CodeDisplay: "Appendectomy"}
	h.svc.CreateProcedure(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(p.FHIRID, "1")
	err := h.VreadProcedureFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryProcedureFHIR(t *testing.T) {
	h, e := newTestHandler()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002", CodeDisplay: "Appendectomy"}
	h.svc.CreateProcedure(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.FHIRID)
	err := h.HistoryProcedureFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// ── RegisterRoutes ──

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
		"POST:/api/v1/conditions",
		"GET:/api/v1/conditions/:id",
		"GET:/api/v1/conditions",
		"PUT:/api/v1/conditions/:id",
		"DELETE:/api/v1/conditions/:id",
		"POST:/api/v1/observations",
		"GET:/api/v1/observations/:id",
		"POST:/api/v1/allergies",
		"GET:/api/v1/allergies/:id",
		"POST:/api/v1/procedures",
		"GET:/api/v1/procedures/:id",
		"GET:/fhir/Condition",
		"GET:/fhir/Observation",
		"GET:/fhir/AllergyIntolerance",
		"GET:/fhir/Procedure",
		"POST:/fhir/Condition",
		"PUT:/fhir/Condition/:id",
		"DELETE:/fhir/Condition/:id",
		"PATCH:/fhir/Condition/:id",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
