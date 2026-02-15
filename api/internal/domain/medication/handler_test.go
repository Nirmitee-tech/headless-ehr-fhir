package medication

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
	h := NewHandler(svc)
	e := echo.New()
	return h, e
}

func TestHandler_CreateMedication(t *testing.T) {
	h, e := newTestHandler()

	body := `{"code_value":"12345","code_display":"Aspirin"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medications", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateMedication(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var m Medication
	json.Unmarshal(rec.Body.Bytes(), &m)
	if m.CodeDisplay != "Aspirin" {
		t.Errorf("expected 'Aspirin', got %s", m.CodeDisplay)
	}
}

func TestHandler_CreateMedication_BadRequest(t *testing.T) {
	h, e := newTestHandler()

	body := `{"code_display":"Aspirin"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medications", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateMedication(c)
	if err == nil {
		t.Error("expected error for missing code_value")
	}
}

func TestHandler_GetMedication(t *testing.T) {
	h, e := newTestHandler()

	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	h.svc.CreateMedication(nil, m)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.ID.String())

	err := h.GetMedication(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetMedication_NotFound(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetMedication(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_GetMedication_InvalidID(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.GetMedication(c)
	if err == nil {
		t.Error("expected error for invalid id")
	}
}

func TestHandler_DeleteMedication(t *testing.T) {
	h, e := newTestHandler()

	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	h.svc.CreateMedication(nil, m)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.ID.String())

	err := h.DeleteMedication(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_CreateMedicationRequest(t *testing.T) {
	h, e := newTestHandler()

	patientID := uuid.New()
	medID := uuid.New()
	requesterID := uuid.New()
	body := `{"patient_id":"` + patientID.String() + `","medication_id":"` + medID.String() + `","requester_id":"` + requesterID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medication-requests", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateMedicationRequest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateMedicationRequest_BadRequest(t *testing.T) {
	h, e := newTestHandler()

	body := `{"medication_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medication-requests", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateMedicationRequest(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_CreateMedicationAdministration(t *testing.T) {
	h, e := newTestHandler()

	body := `{"patient_id":"` + uuid.New().String() + `","medication_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medication-administrations", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateMedicationAdministration(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateMedicationDispense(t *testing.T) {
	h, e := newTestHandler()

	body := `{"patient_id":"` + uuid.New().String() + `","medication_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medication-dispenses", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateMedicationDispense(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateMedicationStatement(t *testing.T) {
	h, e := newTestHandler()

	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medication-statements", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateMedicationStatement(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

// ── Additional REST Tests ──

func TestHandler_ListMedications(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateMedication(nil, &Medication{CodeValue: "1", CodeDisplay: "A"})
	h.svc.CreateMedication(nil, &Medication{CodeValue: "2", CodeDisplay: "B"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListMedications(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateMedication(t *testing.T) {
	h, e := newTestHandler()
	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	h.svc.CreateMedication(nil, m)

	body := `{"code_value":"12345","code_display":"Aspirin 500mg"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.ID.String())
	err := h.UpdateMedication(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_AddIngredient(t *testing.T) {
	h, e := newTestHandler()
	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	h.svc.CreateMedication(nil, m)

	body := `{"item_display":"Acetylsalicylic acid"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.ID.String())
	err := h.AddIngredient(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetIngredients(t *testing.T) {
	h, e := newTestHandler()
	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	h.svc.CreateMedication(nil, m)
	h.svc.AddIngredient(nil, &MedicationIngredient{MedicationID: m.ID, ItemDisplay: "Acetylsalicylic acid"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.ID.String())
	err := h.GetIngredients(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_RemoveIngredient(t *testing.T) {
	h, e := newTestHandler()
	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	h.svc.CreateMedication(nil, m)
	ing := &MedicationIngredient{MedicationID: m.ID, ItemDisplay: "Acetylsalicylic acid"}
	h.svc.AddIngredient(nil, ing)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "ingredientId")
	c.SetParamValues(m.ID.String(), ing.ID.String())
	err := h.RemoveIngredient(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_GetMedicationRequest(t *testing.T) {
	h, e := newTestHandler()
	mr := &MedicationRequest{PatientID: uuid.New(), MedicationID: uuid.New(), RequesterID: uuid.New()}
	h.svc.CreateMedicationRequest(nil, mr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(mr.ID.String())
	err := h.GetMedicationRequest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListMedicationRequests(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateMedicationRequest(nil, &MedicationRequest{PatientID: uuid.New(), MedicationID: uuid.New(), RequesterID: uuid.New()})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListMedicationRequests(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateMedicationRequest(t *testing.T) {
	h, e := newTestHandler()
	mr := &MedicationRequest{PatientID: uuid.New(), MedicationID: uuid.New(), RequesterID: uuid.New()}
	h.svc.CreateMedicationRequest(nil, mr)

	body := `{"status":"active"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(mr.ID.String())
	err := h.UpdateMedicationRequest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteMedicationRequest(t *testing.T) {
	h, e := newTestHandler()
	mr := &MedicationRequest{PatientID: uuid.New(), MedicationID: uuid.New(), RequesterID: uuid.New()}
	h.svc.CreateMedicationRequest(nil, mr)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(mr.ID.String())
	err := h.DeleteMedicationRequest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_GetMedicationAdministration(t *testing.T) {
	h, e := newTestHandler()
	ma := &MedicationAdministration{PatientID: uuid.New(), MedicationID: uuid.New()}
	h.svc.CreateMedicationAdministration(nil, ma)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ma.ID.String())
	err := h.GetMedicationAdministration(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListMedicationAdministrations(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateMedicationAdministration(nil, &MedicationAdministration{PatientID: uuid.New(), MedicationID: uuid.New()})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListMedicationAdministrations(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateMedicationAdministration(t *testing.T) {
	h, e := newTestHandler()
	ma := &MedicationAdministration{PatientID: uuid.New(), MedicationID: uuid.New()}
	h.svc.CreateMedicationAdministration(nil, ma)

	body := `{"status":"completed"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ma.ID.String())
	err := h.UpdateMedicationAdministration(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteMedicationAdministration(t *testing.T) {
	h, e := newTestHandler()
	ma := &MedicationAdministration{PatientID: uuid.New(), MedicationID: uuid.New()}
	h.svc.CreateMedicationAdministration(nil, ma)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ma.ID.String())
	err := h.DeleteMedicationAdministration(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_GetMedicationDispense(t *testing.T) {
	h, e := newTestHandler()
	md := &MedicationDispense{PatientID: uuid.New(), MedicationID: uuid.New()}
	h.svc.CreateMedicationDispense(nil, md)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(md.ID.String())
	err := h.GetMedicationDispense(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListMedicationDispenses(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateMedicationDispense(nil, &MedicationDispense{PatientID: uuid.New(), MedicationID: uuid.New()})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListMedicationDispenses(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateMedicationDispense(t *testing.T) {
	h, e := newTestHandler()
	md := &MedicationDispense{PatientID: uuid.New(), MedicationID: uuid.New()}
	h.svc.CreateMedicationDispense(nil, md)

	body := `{"status":"completed"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(md.ID.String())
	err := h.UpdateMedicationDispense(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteMedicationDispense(t *testing.T) {
	h, e := newTestHandler()
	md := &MedicationDispense{PatientID: uuid.New(), MedicationID: uuid.New()}
	h.svc.CreateMedicationDispense(nil, md)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(md.ID.String())
	err := h.DeleteMedicationDispense(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_GetMedicationStatement(t *testing.T) {
	h, e := newTestHandler()
	ms := &MedicationStatement{PatientID: uuid.New()}
	h.svc.CreateMedicationStatement(nil, ms)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ms.ID.String())
	err := h.GetMedicationStatement(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListMedicationStatements(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateMedicationStatement(nil, &MedicationStatement{PatientID: uuid.New()})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListMedicationStatements(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateMedicationStatement(t *testing.T) {
	h, e := newTestHandler()
	ms := &MedicationStatement{PatientID: uuid.New()}
	h.svc.CreateMedicationStatement(nil, ms)

	body := `{"status":"completed"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ms.ID.String())
	err := h.UpdateMedicationStatement(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteMedicationStatement(t *testing.T) {
	h, e := newTestHandler()
	ms := &MedicationStatement{PatientID: uuid.New()}
	h.svc.CreateMedicationStatement(nil, ms)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ms.ID.String())
	err := h.DeleteMedicationStatement(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// ── FHIR Medication Endpoints ──

func TestHandler_SearchMedicationsFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateMedication(nil, &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Medication", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchMedicationsFHIR(c)
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

func TestHandler_GetMedicationFHIR(t *testing.T) {
	h, e := newTestHandler()
	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	h.svc.CreateMedication(nil, m)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.FHIRID)
	err := h.GetMedicationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetMedicationFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	err := h.GetMedicationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateMedicationFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"code_value":"12345","code_display":"Aspirin"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Medication", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateMedicationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc == "" {
		t.Error("expected Location header")
	}
}

func TestHandler_UpdateMedicationFHIR(t *testing.T) {
	h, e := newTestHandler()
	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	h.svc.CreateMedication(nil, m)

	body := `{"code_value":"12345","code_display":"Aspirin 500mg"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.FHIRID)
	err := h.UpdateMedicationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteMedicationFHIR(t *testing.T) {
	h, e := newTestHandler()
	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	h.svc.CreateMedication(nil, m)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.FHIRID)
	err := h.DeleteMedicationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchMedicationFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	h.svc.CreateMedication(nil, m)

	body := `{"status":"inactive"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.FHIRID)
	err := h.PatchMedicationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadMedicationFHIR(t *testing.T) {
	h, e := newTestHandler()
	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	h.svc.CreateMedication(nil, m)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(m.FHIRID, "1")
	err := h.VreadMedicationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryMedicationFHIR(t *testing.T) {
	h, e := newTestHandler()
	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	h.svc.CreateMedication(nil, m)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.FHIRID)
	err := h.HistoryMedicationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// ── FHIR MedicationRequest Endpoints ──

func TestHandler_SearchMedicationRequestsFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateMedicationRequest(nil, &MedicationRequest{PatientID: uuid.New(), MedicationID: uuid.New(), RequesterID: uuid.New()})

	req := httptest.NewRequest(http.MethodGet, "/fhir/MedicationRequest", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchMedicationRequestsFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetMedicationRequestFHIR(t *testing.T) {
	h, e := newTestHandler()
	mr := &MedicationRequest{PatientID: uuid.New(), MedicationID: uuid.New(), RequesterID: uuid.New()}
	h.svc.CreateMedicationRequest(nil, mr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(mr.FHIRID)
	err := h.GetMedicationRequestFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_CreateMedicationRequestFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","medication_id":"` + uuid.New().String() + `","requester_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/MedicationRequest", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateMedicationRequestFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_UpdateMedicationRequestFHIR(t *testing.T) {
	h, e := newTestHandler()
	mr := &MedicationRequest{PatientID: uuid.New(), MedicationID: uuid.New(), RequesterID: uuid.New()}
	h.svc.CreateMedicationRequest(nil, mr)

	body := `{"status":"active"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(mr.FHIRID)
	err := h.UpdateMedicationRequestFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteMedicationRequestFHIR(t *testing.T) {
	h, e := newTestHandler()
	mr := &MedicationRequest{PatientID: uuid.New(), MedicationID: uuid.New(), RequesterID: uuid.New()}
	h.svc.CreateMedicationRequest(nil, mr)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(mr.FHIRID)
	err := h.DeleteMedicationRequestFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_VreadMedicationRequestFHIR(t *testing.T) {
	h, e := newTestHandler()
	mr := &MedicationRequest{PatientID: uuid.New(), MedicationID: uuid.New(), RequesterID: uuid.New()}
	h.svc.CreateMedicationRequest(nil, mr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(mr.FHIRID, "1")
	err := h.VreadMedicationRequestFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryMedicationRequestFHIR(t *testing.T) {
	h, e := newTestHandler()
	mr := &MedicationRequest{PatientID: uuid.New(), MedicationID: uuid.New(), RequesterID: uuid.New()}
	h.svc.CreateMedicationRequest(nil, mr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(mr.FHIRID)
	err := h.HistoryMedicationRequestFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// ── FHIR MedicationAdministration Endpoints ──

func TestHandler_SearchMedicationAdministrationsFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateMedicationAdministration(nil, &MedicationAdministration{PatientID: uuid.New(), MedicationID: uuid.New()})

	req := httptest.NewRequest(http.MethodGet, "/fhir/MedicationAdministration", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchMedicationAdministrationsFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetMedicationAdministrationFHIR(t *testing.T) {
	h, e := newTestHandler()
	ma := &MedicationAdministration{PatientID: uuid.New(), MedicationID: uuid.New()}
	h.svc.CreateMedicationAdministration(nil, ma)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ma.FHIRID)
	err := h.GetMedicationAdministrationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_CreateMedicationAdministrationFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","medication_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/MedicationAdministration", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateMedicationAdministrationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_DeleteMedicationAdministrationFHIR(t *testing.T) {
	h, e := newTestHandler()
	ma := &MedicationAdministration{PatientID: uuid.New(), MedicationID: uuid.New()}
	h.svc.CreateMedicationAdministration(nil, ma)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ma.FHIRID)
	err := h.DeleteMedicationAdministrationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_VreadMedicationAdministrationFHIR(t *testing.T) {
	h, e := newTestHandler()
	ma := &MedicationAdministration{PatientID: uuid.New(), MedicationID: uuid.New()}
	h.svc.CreateMedicationAdministration(nil, ma)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(ma.FHIRID, "1")
	err := h.VreadMedicationAdministrationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryMedicationAdministrationFHIR(t *testing.T) {
	h, e := newTestHandler()
	ma := &MedicationAdministration{PatientID: uuid.New(), MedicationID: uuid.New()}
	h.svc.CreateMedicationAdministration(nil, ma)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ma.FHIRID)
	err := h.HistoryMedicationAdministrationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// ── FHIR MedicationDispense Endpoints ──

func TestHandler_SearchMedicationDispensesFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateMedicationDispense(nil, &MedicationDispense{PatientID: uuid.New(), MedicationID: uuid.New()})

	req := httptest.NewRequest(http.MethodGet, "/fhir/MedicationDispense", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchMedicationDispensesFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetMedicationDispenseFHIR(t *testing.T) {
	h, e := newTestHandler()
	md := &MedicationDispense{PatientID: uuid.New(), MedicationID: uuid.New()}
	h.svc.CreateMedicationDispense(nil, md)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(md.FHIRID)
	err := h.GetMedicationDispenseFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_CreateMedicationDispenseFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","medication_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/MedicationDispense", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateMedicationDispenseFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_DeleteMedicationDispenseFHIR(t *testing.T) {
	h, e := newTestHandler()
	md := &MedicationDispense{PatientID: uuid.New(), MedicationID: uuid.New()}
	h.svc.CreateMedicationDispense(nil, md)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(md.FHIRID)
	err := h.DeleteMedicationDispenseFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_VreadMedicationDispenseFHIR(t *testing.T) {
	h, e := newTestHandler()
	md := &MedicationDispense{PatientID: uuid.New(), MedicationID: uuid.New()}
	h.svc.CreateMedicationDispense(nil, md)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(md.FHIRID, "1")
	err := h.VreadMedicationDispenseFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryMedicationDispenseFHIR(t *testing.T) {
	h, e := newTestHandler()
	md := &MedicationDispense{PatientID: uuid.New(), MedicationID: uuid.New()}
	h.svc.CreateMedicationDispense(nil, md)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(md.FHIRID)
	err := h.HistoryMedicationDispenseFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
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
		"POST:/api/v1/medications",
		"GET:/api/v1/medications/:id",
		"POST:/api/v1/medication-requests",
		"GET:/fhir/Medication",
		"GET:/fhir/MedicationRequest",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}

// -- applyMedicationPatch unit tests --

func TestApplyMedicationPatch_Status(t *testing.T) {
	m := &Medication{
		Status:      "active",
		CodeValue:   "12345",
		CodeDisplay: "Aspirin",
	}

	patched := map[string]interface{}{
		"status": "inactive",
	}
	applyMedicationPatch(m, patched)

	if m.Status != "inactive" {
		t.Errorf("expected status 'inactive', got %q", m.Status)
	}
	if m.CodeValue != "12345" {
		t.Errorf("expected CodeValue unchanged '12345', got %q", m.CodeValue)
	}
	if m.CodeDisplay != "Aspirin" {
		t.Errorf("expected CodeDisplay unchanged 'Aspirin', got %q", m.CodeDisplay)
	}
}

func TestApplyMedicationPatch_Code(t *testing.T) {
	m := &Medication{
		Status:      "active",
		CodeValue:   "12345",
		CodeDisplay: "Aspirin",
	}

	patched := map[string]interface{}{
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"code":    "67890",
					"display": "Ibuprofen",
					"system":  "http://www.nlm.nih.gov/research/umls/rxnorm",
				},
			},
		},
	}
	applyMedicationPatch(m, patched)

	if m.CodeValue != "67890" {
		t.Errorf("expected CodeValue '67890', got %q", m.CodeValue)
	}
	if m.CodeDisplay != "Ibuprofen" {
		t.Errorf("expected CodeDisplay 'Ibuprofen', got %q", m.CodeDisplay)
	}
	if m.CodeSystem == nil || *m.CodeSystem != "http://www.nlm.nih.gov/research/umls/rxnorm" {
		t.Errorf("expected CodeSystem to be set, got %v", m.CodeSystem)
	}
	if m.Status != "active" {
		t.Errorf("expected Status unchanged 'active', got %q", m.Status)
	}
}

// -- applyMedicationRequestPatch unit tests --

func TestApplyMedicationRequestPatch_StatusIntent(t *testing.T) {
	mr := &MedicationRequest{
		Status:       "draft",
		Intent:       "proposal",
		PatientID:    uuid.New(),
		MedicationID: uuid.New(),
		RequesterID:  uuid.New(),
	}

	patched := map[string]interface{}{
		"status":   "active",
		"intent":   "order",
		"priority": "urgent",
	}
	applyMedicationRequestPatch(mr, patched)

	if mr.Status != "active" {
		t.Errorf("expected status 'active', got %q", mr.Status)
	}
	if mr.Intent != "order" {
		t.Errorf("expected intent 'order', got %q", mr.Intent)
	}
	if mr.Priority == nil || *mr.Priority != "urgent" {
		t.Errorf("expected priority 'urgent', got %v", mr.Priority)
	}
}

func TestApplyMedicationRequestPatch_DosageInstruction(t *testing.T) {
	mr := &MedicationRequest{
		Status:       "active",
		Intent:       "order",
		PatientID:    uuid.New(),
		MedicationID: uuid.New(),
		RequesterID:  uuid.New(),
	}

	patched := map[string]interface{}{
		"dosageInstruction": []interface{}{
			map[string]interface{}{
				"text":             "Take 1 tablet daily",
				"asNeededBoolean":  true,
				"route": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{
							"code":    "26643006",
							"display": "Oral",
						},
					},
				},
				"doseAndRate": []interface{}{
					map[string]interface{}{
						"doseQuantity": map[string]interface{}{
							"value": float64(500),
							"unit":  "mg",
						},
					},
				},
			},
		},
	}
	applyMedicationRequestPatch(mr, patched)

	if mr.DosageText == nil || *mr.DosageText != "Take 1 tablet daily" {
		t.Errorf("expected DosageText 'Take 1 tablet daily', got %v", mr.DosageText)
	}
	if mr.AsNeeded == nil || *mr.AsNeeded != true {
		t.Errorf("expected AsNeeded true, got %v", mr.AsNeeded)
	}
	if mr.DosageRouteCode == nil || *mr.DosageRouteCode != "26643006" {
		t.Errorf("expected DosageRouteCode '26643006', got %v", mr.DosageRouteCode)
	}
	if mr.DosageRouteDisplay == nil || *mr.DosageRouteDisplay != "Oral" {
		t.Errorf("expected DosageRouteDisplay 'Oral', got %v", mr.DosageRouteDisplay)
	}
	if mr.DoseQuantity == nil || *mr.DoseQuantity != 500 {
		t.Errorf("expected DoseQuantity 500, got %v", mr.DoseQuantity)
	}
	if mr.DoseUnit == nil || *mr.DoseUnit != "mg" {
		t.Errorf("expected DoseUnit 'mg', got %v", mr.DoseUnit)
	}
	if mr.Status != "active" {
		t.Errorf("expected Status unchanged 'active', got %q", mr.Status)
	}
}

// -- applyMedicationAdministrationPatch unit tests --

func TestApplyMedicationAdministrationPatch_Status(t *testing.T) {
	ma := &MedicationAdministration{
		Status:       "in-progress",
		PatientID:    uuid.New(),
		MedicationID: uuid.New(),
	}

	patched := map[string]interface{}{
		"status":            "completed",
		"effectiveDateTime": "2025-06-15T14:30:00Z",
		"dosage": map[string]interface{}{
			"text": "500mg IV infusion",
			"route": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code":    "47625008",
						"display": "Intravenous",
					},
				},
			},
			"dose": map[string]interface{}{
				"value": float64(500),
				"unit":  "mg",
			},
		},
	}
	applyMedicationAdministrationPatch(ma, patched)

	if ma.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", ma.Status)
	}
	if ma.EffectiveDatetime == nil {
		t.Fatal("expected EffectiveDatetime to be set")
	}
	if ma.EffectiveDatetime.Year() != 2025 || ma.EffectiveDatetime.Month() != 6 {
		t.Errorf("unexpected EffectiveDatetime: %v", *ma.EffectiveDatetime)
	}
	if ma.DosageText == nil || *ma.DosageText != "500mg IV infusion" {
		t.Errorf("expected DosageText '500mg IV infusion', got %v", ma.DosageText)
	}
	if ma.DosageRouteCode == nil || *ma.DosageRouteCode != "47625008" {
		t.Errorf("expected DosageRouteCode '47625008', got %v", ma.DosageRouteCode)
	}
	if ma.DoseQuantity == nil || *ma.DoseQuantity != 500 {
		t.Errorf("expected DoseQuantity 500, got %v", ma.DoseQuantity)
	}
	if ma.DoseUnit == nil || *ma.DoseUnit != "mg" {
		t.Errorf("expected DoseUnit 'mg', got %v", ma.DoseUnit)
	}
}

// -- applyMedicationDispensePatch unit tests --

func TestApplyMedicationDispensePatch_Status(t *testing.T) {
	md := &MedicationDispense{
		Status:       "preparation",
		PatientID:    uuid.New(),
		MedicationID: uuid.New(),
	}

	patched := map[string]interface{}{
		"status": "completed",
		"quantity": map[string]interface{}{
			"value": float64(30),
			"unit":  "tablets",
		},
		"daysSupply": map[string]interface{}{
			"value": float64(30),
		},
		"whenPrepared":  "2025-07-01T10:00:00Z",
		"whenHandedOver": "2025-07-01T11:00:00Z",
		"substitution": map[string]interface{}{
			"wasSubstituted": true,
			"type": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code": "G",
					},
				},
			},
			"reason": []interface{}{
				map[string]interface{}{
					"text": "generic substitution",
				},
			},
		},
		"note": []interface{}{
			map[string]interface{}{
				"text": "Dispensed at pharmacy",
			},
		},
	}
	applyMedicationDispensePatch(md, patched)

	if md.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", md.Status)
	}
	if md.QuantityValue == nil || *md.QuantityValue != 30 {
		t.Errorf("expected QuantityValue 30, got %v", md.QuantityValue)
	}
	if md.QuantityUnit == nil || *md.QuantityUnit != "tablets" {
		t.Errorf("expected QuantityUnit 'tablets', got %v", md.QuantityUnit)
	}
	if md.DaysSupply == nil || *md.DaysSupply != 30 {
		t.Errorf("expected DaysSupply 30, got %v", md.DaysSupply)
	}
	if md.WhenPrepared == nil {
		t.Fatal("expected WhenPrepared to be set")
	}
	if md.WhenHandedOver == nil {
		t.Fatal("expected WhenHandedOver to be set")
	}
	if md.WasSubstituted == nil || *md.WasSubstituted != true {
		t.Errorf("expected WasSubstituted true, got %v", md.WasSubstituted)
	}
	if md.SubstitutionTypeCode == nil || *md.SubstitutionTypeCode != "G" {
		t.Errorf("expected SubstitutionTypeCode 'G', got %v", md.SubstitutionTypeCode)
	}
	if md.SubstitutionReason == nil || *md.SubstitutionReason != "generic substitution" {
		t.Errorf("expected SubstitutionReason 'generic substitution', got %v", md.SubstitutionReason)
	}
	if md.Note == nil || *md.Note != "Dispensed at pharmacy" {
		t.Errorf("expected Note 'Dispensed at pharmacy', got %v", md.Note)
	}
}
