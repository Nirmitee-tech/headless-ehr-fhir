package encounter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	_ "github.com/ehr/ehr/internal/platform/fhir"
)

func newTestHandler() (*Handler, *echo.Echo) {
	svc := newTestService()
	h := NewHandler(svc, nil)
	e := echo.New()
	return h, e
}

func TestHandler_CreateEncounter(t *testing.T) {
	h, e := newTestHandler()

	patientID := uuid.New()
	body := `{"patient_id":"` + patientID.String() + `","class_code":"AMB"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/encounters", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateEncounter(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var enc Encounter
	json.Unmarshal(rec.Body.Bytes(), &enc)
	if enc.ClassCode != "AMB" {
		t.Errorf("expected AMB, got %s", enc.ClassCode)
	}
}

func TestHandler_CreateEncounter_BadRequest(t *testing.T) {
	h, e := newTestHandler()

	body := `{"class_code":"AMB"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/encounters", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateEncounter(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetEncounter(t *testing.T) {
	h, e := newTestHandler()

	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.ID.String())

	err := h.GetEncounter(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetEncounter_NotFound(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetEncounter(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_DeleteEncounter(t *testing.T) {
	h, e := newTestHandler()

	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.ID.String())

	err := h.DeleteEncounter(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_UpdateStatus(t *testing.T) {
	h, e := newTestHandler()

	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	body := `{"status":"in-progress"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.ID.String())

	err := h.UpdateEncounterStatus(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListEncounters(t *testing.T) {
	h, e := newTestHandler()

	h.svc.CreateEncounter(nil, &Encounter{PatientID: uuid.New(), ClassCode: "AMB"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/encounters", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListEncounters(c)
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
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}

	expected := []string{
		"POST:/api/v1/encounters",
		"GET:/api/v1/encounters",
		"GET:/api/v1/encounters/:id",
		"GET:/fhir/Encounter",
		"GET:/fhir/Encounter/:id",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}

// -- Missing REST Handler Tests --

func TestHandler_UpdateEncounter(t *testing.T) {
	h, e := newTestHandler()
	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	body := `{"class_code":"IMP"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.ID.String())

	err := h.UpdateEncounter(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateEncounter_InvalidID(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")
	err := h.UpdateEncounter(c)
	if err == nil {
		t.Error("expected error for invalid id")
	}
}

func TestHandler_AddParticipant(t *testing.T) {
	h, e := newTestHandler()
	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	practID := uuid.New()
	body := `{"practitioner_id":"` + practID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.ID.String())

	err := h.AddParticipant(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_AddParticipant_InvalidID(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")
	err := h.AddParticipant(c)
	if err == nil {
		t.Error("expected error for invalid id")
	}
}

func TestHandler_GetParticipants(t *testing.T) {
	h, e := newTestHandler()
	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	h.svc.AddParticipant(nil, &EncounterParticipant{EncounterID: enc.ID, PractitionerID: uuid.New()})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.ID.String())

	err := h.GetParticipants(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetParticipants_InvalidID(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")
	err := h.GetParticipants(c)
	if err == nil {
		t.Error("expected error for invalid id")
	}
}

func TestHandler_AddDiagnosis(t *testing.T) {
	h, e := newTestHandler()
	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	body := `{"use_code":"AD"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.ID.String())

	err := h.AddDiagnosis(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_AddDiagnosis_InvalidID(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")
	err := h.AddDiagnosis(c)
	if err == nil {
		t.Error("expected error for invalid id")
	}
}

func TestHandler_GetDiagnoses(t *testing.T) {
	h, e := newTestHandler()
	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.ID.String())

	err := h.GetDiagnoses(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetStatusHistory(t *testing.T) {
	h, e := newTestHandler()
	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.ID.String())

	err := h.GetStatusHistory(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- FHIR Handler Tests --

func TestHandler_SearchEncountersFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateEncounter(nil, &Encounter{PatientID: uuid.New(), ClassCode: "AMB"})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Encounter", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchEncountersFHIR(c)
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

func TestHandler_GetEncounterFHIR(t *testing.T) {
	h, e := newTestHandler()
	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.FHIRID)

	err := h.GetEncounterFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["resourceType"] != "Encounter" {
		t.Errorf("expected Encounter, got %v", result["resourceType"])
	}
}

func TestHandler_GetEncounterFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.GetEncounterFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateEncounterFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","class_code":"AMB"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Encounter", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateEncounterFHIR(c)
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
	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["resourceType"] != "Encounter" {
		t.Errorf("expected Encounter, got %v", result["resourceType"])
	}
}

func TestHandler_CreateEncounterFHIR_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"class_code":"AMB"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Encounter", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateEncounterFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandler_UpdateEncounterFHIR(t *testing.T) {
	h, e := newTestHandler()
	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	body := `{"class_code":"IMP"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.FHIRID)

	err := h.UpdateEncounterFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateEncounterFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	body := `{"class_code":"IMP"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.UpdateEncounterFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_DeleteEncounterFHIR(t *testing.T) {
	h, e := newTestHandler()
	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.FHIRID)

	err := h.DeleteEncounterFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_DeleteEncounterFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.DeleteEncounterFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_PatchEncounterFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	body := `{"status":"in-progress"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.FHIRID)

	err := h.PatchEncounterFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_PatchEncounterFHIR_JSONPatch(t *testing.T) {
	h, e := newTestHandler()
	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	body := `[{"op":"replace","path":"/status","value":"in-progress"}]`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.FHIRID)

	err := h.PatchEncounterFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_PatchEncounterFHIR_UnsupportedMediaType(t *testing.T) {
	h, e := newTestHandler()
	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.FHIRID)

	err := h.PatchEncounterFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415, got %d", rec.Code)
	}
}

func TestHandler_PatchEncounterFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.PatchEncounterFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_VreadEncounterFHIR(t *testing.T) {
	h, e := newTestHandler()
	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(enc.FHIRID, "1")

	err := h.VreadEncounterFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Error("expected ETag header")
	}
}

func TestHandler_VreadEncounterFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues("nonexistent", "1")

	err := h.VreadEncounterFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_HistoryEncounterFHIR(t *testing.T) {
	h, e := newTestHandler()
	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.FHIRID)

	err := h.HistoryEncounterFHIR(c)
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
	if bundle["type"] != "history" {
		t.Errorf("expected history, got %v", bundle["type"])
	}
}

func TestHandler_HistoryEncounterFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.HistoryEncounterFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// -- applyEncounterPatch unit tests --

func TestApplyEncounterPatch_Status(t *testing.T) {
	enc := &Encounter{
		Status:    "planned",
		ClassCode: "AMB",
		PatientID: uuid.New(),
	}

	patched := map[string]interface{}{
		"status": "in-progress",
	}
	applyEncounterPatch(enc, patched)

	if enc.Status != "in-progress" {
		t.Errorf("expected status 'in-progress', got %q", enc.Status)
	}
	if enc.ClassCode != "AMB" {
		t.Errorf("expected ClassCode unchanged 'AMB', got %q", enc.ClassCode)
	}
}

func TestApplyEncounterPatch_Class(t *testing.T) {
	enc := &Encounter{
		Status:    "planned",
		ClassCode: "AMB",
	}

	patched := map[string]interface{}{
		"class": map[string]interface{}{
			"code":    "IMP",
			"display": "inpatient encounter",
		},
	}
	applyEncounterPatch(enc, patched)

	if enc.ClassCode != "IMP" {
		t.Errorf("expected ClassCode 'IMP', got %q", enc.ClassCode)
	}
	if enc.ClassDisplay == nil || *enc.ClassDisplay != "inpatient encounter" {
		t.Errorf("expected ClassDisplay 'inpatient encounter', got %v", enc.ClassDisplay)
	}
	if enc.Status != "planned" {
		t.Errorf("expected Status unchanged 'planned', got %q", enc.Status)
	}
}

func TestApplyEncounterPatch_Period(t *testing.T) {
	enc := &Encounter{
		Status:    "in-progress",
		ClassCode: "AMB",
	}

	patched := map[string]interface{}{
		"period": map[string]interface{}{
			"start": "2025-06-01T09:00:00Z",
			"end":   "2025-06-01T10:30:00Z",
		},
	}
	applyEncounterPatch(enc, patched)

	if enc.PeriodStart.IsZero() {
		t.Fatal("expected PeriodStart to be set")
	}
	if enc.PeriodStart.Year() != 2025 || enc.PeriodStart.Month() != 6 || enc.PeriodStart.Day() != 1 {
		t.Errorf("unexpected PeriodStart: %v", enc.PeriodStart)
	}
	if enc.PeriodEnd == nil {
		t.Fatal("expected PeriodEnd to be set")
	}
	if enc.PeriodEnd.Hour() != 10 || enc.PeriodEnd.Minute() != 30 {
		t.Errorf("unexpected PeriodEnd: %v", *enc.PeriodEnd)
	}
}

func TestApplyEncounterPatch_EmptyMap(t *testing.T) {
	orig := "AMB"
	enc := &Encounter{
		Status:    "planned",
		ClassCode: "AMB",
		ClassDisplay: &orig,
	}

	patched := map[string]interface{}{}
	applyEncounterPatch(enc, patched)

	if enc.Status != "planned" {
		t.Errorf("expected Status unchanged 'planned', got %q", enc.Status)
	}
	if enc.ClassCode != "AMB" {
		t.Errorf("expected ClassCode unchanged 'AMB', got %q", enc.ClassCode)
	}
	if enc.ClassDisplay == nil || *enc.ClassDisplay != "AMB" {
		t.Errorf("expected ClassDisplay unchanged")
	}
}
