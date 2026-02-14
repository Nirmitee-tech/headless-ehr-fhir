package nursing

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

func TestHandler_CreateTemplate(t *testing.T) {
	h, e := newTestHandler()
	body := `{"name":"Vitals"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateTemplate(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateTemplate_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateTemplate(c)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestHandler_GetTemplate(t *testing.T) {
	h, e := newTestHandler()
	tmpl := &FlowsheetTemplate{Name: "Vitals"}
	h.svc.CreateTemplate(nil, tmpl)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tmpl.ID.String())

	err := h.GetTemplate(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetTemplate_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetTemplate(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_DeleteTemplate(t *testing.T) {
	h, e := newTestHandler()
	tmpl := &FlowsheetTemplate{Name: "Vitals"}
	h.svc.CreateTemplate(nil, tmpl)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tmpl.ID.String())

	err := h.DeleteTemplate(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddTemplateRow(t *testing.T) {
	h, e := newTestHandler()
	tmplID := uuid.New()
	body := `{"label":"Heart Rate"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tmplID.String())

	err := h.AddTemplateRow(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateEntry(t *testing.T) {
	h, e := newTestHandler()
	body := `{"template_id":"` + uuid.New().String() + `","row_id":"` + uuid.New().String() + `","patient_id":"` + uuid.New().String() + `","encounter_id":"` + uuid.New().String() + `","recorded_by_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateEntry(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateEntry_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"template_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateEntry(c)
	if err == nil {
		t.Error("expected error for missing required fields")
	}
}

func TestHandler_CreateAssessment(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","encounter_id":"` + uuid.New().String() + `","nurse_id":"` + uuid.New().String() + `","assessment_type":"admission"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateAssessment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateFallRisk(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","assessed_by_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateFallRisk(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateSkinAssessment(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","assessed_by_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateSkinAssessment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreatePainAssessment(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","assessed_by_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreatePainAssessment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateLinesDrains(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","encounter_id":"` + uuid.New().String() + `","type":"IV"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateLinesDrains(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateRestraint(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","restraint_type":"wrist","applied_by_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateRestraint(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateIntakeOutput(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","encounter_id":"` + uuid.New().String() + `","category":"intake","recorded_by_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateIntakeOutput(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateIntakeOutput_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateIntakeOutput(c)
	if err == nil {
		t.Error("expected error for missing required fields")
	}
}

// -- List/Update/Get/Delete tests for all resource types --

func TestHandler_ListTemplates(t *testing.T) {
	h, e := newTestHandler()
	tmpl := &FlowsheetTemplate{Name: "Vitals"}
	h.svc.CreateTemplate(nil, tmpl)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListTemplates(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateTemplate(t *testing.T) {
	h, e := newTestHandler()
	tmpl := &FlowsheetTemplate{Name: "Vitals"}
	h.svc.CreateTemplate(nil, tmpl)

	body := `{"name":"Vitals Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tmpl.ID.String())

	err := h.UpdateTemplate(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetTemplateRows(t *testing.T) {
	h, e := newTestHandler()
	tmplID := uuid.New()
	row := &FlowsheetRow{TemplateID: tmplID, Label: "Heart Rate"}
	h.svc.AddTemplateRow(nil, row)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tmplID.String())

	err := h.GetTemplateRows(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetEntry(t *testing.T) {
	h, e := newTestHandler()
	entry := &FlowsheetEntry{
		TemplateID:   uuid.New(),
		RowID:        uuid.New(),
		PatientID:    uuid.New(),
		EncounterID:  uuid.New(),
		RecordedByID: uuid.New(),
	}
	h.svc.CreateEntry(nil, entry)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(entry.ID.String())

	err := h.GetEntry(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListEntries(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListEntries(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteEntry(t *testing.T) {
	h, e := newTestHandler()
	entry := &FlowsheetEntry{
		TemplateID:   uuid.New(),
		RowID:        uuid.New(),
		PatientID:    uuid.New(),
		EncounterID:  uuid.New(),
		RecordedByID: uuid.New(),
	}
	h.svc.CreateEntry(nil, entry)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(entry.ID.String())

	err := h.DeleteEntry(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_GetAssessment(t *testing.T) {
	h, e := newTestHandler()
	a := &NursingAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), NurseID: uuid.New(), AssessmentType: "admission"}
	h.svc.CreateAssessment(nil, a)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())

	err := h.GetAssessment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListAssessments(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListAssessments(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateAssessment(t *testing.T) {
	h, e := newTestHandler()
	a := &NursingAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), NurseID: uuid.New(), AssessmentType: "admission"}
	h.svc.CreateAssessment(nil, a)

	body := `{"patient_id":"` + a.PatientID.String() + `","encounter_id":"` + a.EncounterID.String() + `","nurse_id":"` + a.NurseID.String() + `","assessment_type":"discharge"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())

	err := h.UpdateAssessment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteAssessment(t *testing.T) {
	h, e := newTestHandler()
	a := &NursingAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), NurseID: uuid.New(), AssessmentType: "admission"}
	h.svc.CreateAssessment(nil, a)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())

	err := h.DeleteAssessment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_GetFallRisk(t *testing.T) {
	h, e := newTestHandler()
	a := &FallRiskAssessment{PatientID: uuid.New(), AssessedByID: uuid.New()}
	h.svc.CreateFallRisk(nil, a)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())

	err := h.GetFallRisk(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListFallRisk(t *testing.T) {
	h, e := newTestHandler()
	pid := uuid.New()
	a := &FallRiskAssessment{PatientID: pid, AssessedByID: uuid.New()}
	h.svc.CreateFallRisk(nil, a)

	req := httptest.NewRequest(http.MethodGet, "/?patient_id="+pid.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/?patient_id=" + pid.String())
	c.QueryParams().Set("patient_id", pid.String())

	err := h.ListFallRisk(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetSkinAssessment(t *testing.T) {
	h, e := newTestHandler()
	a := &SkinAssessment{PatientID: uuid.New(), AssessedByID: uuid.New()}
	h.svc.CreateSkinAssessment(nil, a)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())

	err := h.GetSkinAssessment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListSkinAssessments(t *testing.T) {
	h, e := newTestHandler()
	pid := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/?patient_id="+pid.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.QueryParams().Set("patient_id", pid.String())

	err := h.ListSkinAssessments(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetPainAssessment(t *testing.T) {
	h, e := newTestHandler()
	a := &PainAssessment{PatientID: uuid.New(), AssessedByID: uuid.New()}
	h.svc.CreatePainAssessment(nil, a)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())

	err := h.GetPainAssessment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListPainAssessments(t *testing.T) {
	h, e := newTestHandler()
	pid := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/?patient_id="+pid.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.QueryParams().Set("patient_id", pid.String())

	err := h.ListPainAssessments(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetLinesDrains(t *testing.T) {
	h, e := newTestHandler()
	l := &LinesDrainsAirways{PatientID: uuid.New(), EncounterID: uuid.New(), Type: "IV"}
	h.svc.CreateLinesDrains(nil, l)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(l.ID.String())

	err := h.GetLinesDrains(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListLinesDrains(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListLinesDrains(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateLinesDrains(t *testing.T) {
	h, e := newTestHandler()
	l := &LinesDrainsAirways{PatientID: uuid.New(), EncounterID: uuid.New(), Type: "IV"}
	h.svc.CreateLinesDrains(nil, l)

	body := `{"patient_id":"` + l.PatientID.String() + `","encounter_id":"` + l.EncounterID.String() + `","type":"Central Line"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(l.ID.String())

	err := h.UpdateLinesDrains(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteLinesDrains(t *testing.T) {
	h, e := newTestHandler()
	l := &LinesDrainsAirways{PatientID: uuid.New(), EncounterID: uuid.New(), Type: "IV"}
	h.svc.CreateLinesDrains(nil, l)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(l.ID.String())

	err := h.DeleteLinesDrains(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_GetRestraint(t *testing.T) {
	h, e := newTestHandler()
	r := &RestraintRecord{PatientID: uuid.New(), RestraintType: "wrist", AppliedByID: uuid.New()}
	h.svc.CreateRestraint(nil, r)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(r.ID.String())

	err := h.GetRestraint(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListRestraints(t *testing.T) {
	h, e := newTestHandler()
	pid := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/?patient_id="+pid.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.QueryParams().Set("patient_id", pid.String())

	err := h.ListRestraints(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateRestraint(t *testing.T) {
	h, e := newTestHandler()
	r := &RestraintRecord{PatientID: uuid.New(), RestraintType: "wrist", AppliedByID: uuid.New()}
	h.svc.CreateRestraint(nil, r)

	body := `{"patient_id":"` + r.PatientID.String() + `","restraint_type":"ankle","applied_by_id":"` + r.AppliedByID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(r.ID.String())

	err := h.UpdateRestraint(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetIntakeOutput(t *testing.T) {
	h, e := newTestHandler()
	r := &IntakeOutputRecord{PatientID: uuid.New(), EncounterID: uuid.New(), Category: "intake", RecordedByID: uuid.New()}
	h.svc.CreateIntakeOutput(nil, r)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(r.ID.String())

	err := h.GetIntakeOutput(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListIntakeOutput(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListIntakeOutput(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteIntakeOutput(t *testing.T) {
	h, e := newTestHandler()
	r := &IntakeOutputRecord{PatientID: uuid.New(), EncounterID: uuid.New(), Category: "intake", RecordedByID: uuid.New()}
	h.svc.CreateIntakeOutput(nil, r)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(r.ID.String())

	err := h.DeleteIntakeOutput(c)
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
	h.RegisterRoutes(api)

	routes := e.Routes()
	if len(routes) == 0 {
		t.Error("expected routes to be registered")
	}
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}
	expected := []string{
		"POST:/api/v1/flowsheet-templates",
		"GET:/api/v1/flowsheet-templates/:id",
		"POST:/api/v1/flowsheet-entries",
		"POST:/api/v1/nursing-assessments",
		"POST:/api/v1/fall-risk-assessments",
		"POST:/api/v1/skin-assessments",
		"POST:/api/v1/pain-assessments",
		"POST:/api/v1/lines-drains",
		"POST:/api/v1/restraints",
		"POST:/api/v1/intake-output",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
