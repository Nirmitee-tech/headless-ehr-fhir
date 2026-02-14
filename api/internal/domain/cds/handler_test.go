package cds

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

// ── CDS Rule Handlers ──

func TestHandler_CreateCDSRule(t *testing.T) {
	h, e := newTestHandler()
	body := `{"rule_name":"Drug Allergy","rule_type":"allergy-check"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateCDSRule(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateCDSRule_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"rule_name":"Drug Allergy"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateCDSRule(c)
	if err == nil {
		t.Error("expected error for missing rule_type")
	}
}

func TestHandler_GetCDSRule(t *testing.T) {
	h, e := newTestHandler()
	r := &CDSRule{RuleName: "Drug Allergy", RuleType: "allergy-check"}
	h.svc.CreateCDSRule(nil, r)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(r.ID.String())
	err := h.GetCDSRule(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetCDSRule_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetCDSRule(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_DeleteCDSRule(t *testing.T) {
	h, e := newTestHandler()
	r := &CDSRule{RuleName: "Drug Allergy", RuleType: "allergy-check"}
	h.svc.CreateCDSRule(nil, r)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(r.ID.String())
	err := h.DeleteCDSRule(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// ── CDS Alert Handlers ──

func TestHandler_CreateCDSAlert(t *testing.T) {
	h, e := newTestHandler()
	body := `{"rule_id":"` + uuid.New().String() + `","patient_id":"` + uuid.New().String() + `","summary":"Drug allergy detected"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateCDSAlert(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateCDSAlert_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"rule_id":"` + uuid.New().String() + `","patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateCDSAlert(c)
	if err == nil {
		t.Error("expected error for missing summary")
	}
}

func TestHandler_GetCDSAlert(t *testing.T) {
	h, e := newTestHandler()
	a := &CDSAlert{RuleID: uuid.New(), PatientID: uuid.New(), Summary: "Alert"}
	h.svc.CreateCDSAlert(nil, a)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())
	err := h.GetCDSAlert(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteCDSAlert(t *testing.T) {
	h, e := newTestHandler()
	a := &CDSAlert{RuleID: uuid.New(), PatientID: uuid.New(), Summary: "Alert"}
	h.svc.CreateCDSAlert(nil, a)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())
	err := h.DeleteCDSAlert(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddAlertResponse(t *testing.T) {
	h, e := newTestHandler()
	alertID := uuid.New()
	body := `{"practitioner_id":"` + uuid.New().String() + `","action":"accept"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(alertID.String())
	err := h.AddAlertResponse(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_AddAlertResponse_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	alertID := uuid.New()
	body := `{"practitioner_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(alertID.String())
	err := h.AddAlertResponse(c)
	if err == nil {
		t.Error("expected error for missing action")
	}
}

// ── Drug Interaction Handlers ──

func TestHandler_CreateDrugInteraction(t *testing.T) {
	h, e := newTestHandler()
	body := `{"medication_a_name":"Warfarin","medication_b_name":"Aspirin","severity":"high"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateDrugInteraction(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateDrugInteraction_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"medication_a_name":"Warfarin","medication_b_name":"Aspirin"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateDrugInteraction(c)
	if err == nil {
		t.Error("expected error for missing severity")
	}
}

func TestHandler_GetDrugInteraction(t *testing.T) {
	h, e := newTestHandler()
	d := &DrugInteraction{MedicationAName: "Warfarin", MedicationBName: "Aspirin", Severity: "high"}
	h.svc.CreateDrugInteraction(nil, d)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(d.ID.String())
	err := h.GetDrugInteraction(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteDrugInteraction(t *testing.T) {
	h, e := newTestHandler()
	d := &DrugInteraction{MedicationAName: "Warfarin", MedicationBName: "Aspirin", Severity: "high"}
	h.svc.CreateDrugInteraction(nil, d)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(d.ID.String())
	err := h.DeleteDrugInteraction(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// ── Order Set Handlers ──

func TestHandler_CreateOrderSet(t *testing.T) {
	h, e := newTestHandler()
	body := `{"name":"Sepsis Bundle"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateOrderSet(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateOrderSet_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateOrderSet(c)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestHandler_GetOrderSet(t *testing.T) {
	h, e := newTestHandler()
	o := &OrderSet{Name: "Sepsis Bundle"}
	h.svc.CreateOrderSet(nil, o)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(o.ID.String())
	err := h.GetOrderSet(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteOrderSet(t *testing.T) {
	h, e := newTestHandler()
	o := &OrderSet{Name: "Sepsis Bundle"}
	h.svc.CreateOrderSet(nil, o)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(o.ID.String())
	err := h.DeleteOrderSet(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddOrderSetSection(t *testing.T) {
	h, e := newTestHandler()
	osID := uuid.New()
	body := `{"name":"Antibiotics"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(osID.String())
	err := h.AddOrderSetSection(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_AddOrderSetItem(t *testing.T) {
	h, e := newTestHandler()
	secID := uuid.New()
	body := `{"item_name":"Ceftriaxone"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(secID.String())
	err := h.AddOrderSetItem(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

// ── Clinical Pathway Handlers ──

func TestHandler_CreateClinicalPathway(t *testing.T) {
	h, e := newTestHandler()
	body := `{"name":"Heart Failure"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateClinicalPathway(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateClinicalPathway_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateClinicalPathway(c)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestHandler_GetClinicalPathway(t *testing.T) {
	h, e := newTestHandler()
	p := &ClinicalPathway{Name: "Heart Failure"}
	h.svc.CreateClinicalPathway(nil, p)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())
	err := h.GetClinicalPathway(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteClinicalPathway(t *testing.T) {
	h, e := newTestHandler()
	p := &ClinicalPathway{Name: "Heart Failure"}
	h.svc.CreateClinicalPathway(nil, p)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())
	err := h.DeleteClinicalPathway(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddPathwayPhase(t *testing.T) {
	h, e := newTestHandler()
	pwID := uuid.New()
	body := `{"name":"Acute Phase"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(pwID.String())
	err := h.AddPathwayPhase(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

// ── Pathway Enrollment Handlers ──

func TestHandler_CreatePathwayEnrollment(t *testing.T) {
	h, e := newTestHandler()
	body := `{"pathway_id":"` + uuid.New().String() + `","patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreatePathwayEnrollment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreatePathwayEnrollment_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"pathway_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreatePathwayEnrollment(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetPathwayEnrollment(t *testing.T) {
	h, e := newTestHandler()
	en := &PatientPathwayEnrollment{PathwayID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreatePathwayEnrollment(nil, en)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(en.ID.String())
	err := h.GetPathwayEnrollment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeletePathwayEnrollment(t *testing.T) {
	h, e := newTestHandler()
	en := &PatientPathwayEnrollment{PathwayID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreatePathwayEnrollment(nil, en)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(en.ID.String())
	err := h.DeletePathwayEnrollment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// ── Formulary Handlers ──

func TestHandler_CreateFormulary(t *testing.T) {
	h, e := newTestHandler()
	body := `{"name":"2025 Formulary"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateFormulary(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateFormulary_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateFormulary(c)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestHandler_GetFormulary(t *testing.T) {
	h, e := newTestHandler()
	f := &Formulary{Name: "2025 Formulary"}
	h.svc.CreateFormulary(nil, f)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(f.ID.String())
	err := h.GetFormulary(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteFormulary(t *testing.T) {
	h, e := newTestHandler()
	f := &Formulary{Name: "2025 Formulary"}
	h.svc.CreateFormulary(nil, f)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(f.ID.String())
	err := h.DeleteFormulary(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddFormularyItem(t *testing.T) {
	h, e := newTestHandler()
	fID := uuid.New()
	body := `{"medication_name":"Metformin"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fID.String())
	err := h.AddFormularyItem(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

// ── Medication Reconciliation Handlers ──

func TestHandler_CreateMedReconciliation(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateMedReconciliation(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateMedReconciliation_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateMedReconciliation(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetMedReconciliation(t *testing.T) {
	h, e := newTestHandler()
	mr := &MedicationReconciliation{PatientID: uuid.New()}
	h.svc.CreateMedReconciliation(nil, mr)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(mr.ID.String())
	err := h.GetMedReconciliation(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteMedReconciliation(t *testing.T) {
	h, e := newTestHandler()
	mr := &MedicationReconciliation{PatientID: uuid.New()}
	h.svc.CreateMedReconciliation(nil, mr)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(mr.ID.String())
	err := h.DeleteMedReconciliation(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddMedReconciliationItem(t *testing.T) {
	h, e := newTestHandler()
	mrID := uuid.New()
	body := `{"medication_name":"Lisinopril"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(mrID.String())
	err := h.AddMedReconciliationItem(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}
