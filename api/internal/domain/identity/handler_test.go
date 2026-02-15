package identity

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

func TestHandler_CreatePatient(t *testing.T) {
	h, e := newTestHandler()

	body := `{"first_name":"John","last_name":"Doe","mrn":"MRN001"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/patients", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreatePatient(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var p Patient
	json.Unmarshal(rec.Body.Bytes(), &p)
	if p.FirstName != "John" {
		t.Errorf("expected John, got %s", p.FirstName)
	}
}

func TestHandler_CreatePatient_BadRequest(t *testing.T) {
	h, e := newTestHandler()

	body := `{"last_name":"Doe"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/patients", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreatePatient(c)
	if err == nil {
		t.Error("expected error for missing fields")
	}
}

func TestHandler_GetPatient(t *testing.T) {
	h, e := newTestHandler()

	p := &Patient{FirstName: "Jane", LastName: "Smith", MRN: "MRN002"}
	h.svc.CreatePatient(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.GetPatient(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetPatient_NotFound(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetPatient(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_DeletePatient(t *testing.T) {
	h, e := newTestHandler()

	p := &Patient{FirstName: "Delete", LastName: "Me", MRN: "MRN-DEL"}
	h.svc.CreatePatient(nil, p)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.DeletePatient(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_ListPatients(t *testing.T) {
	h, e := newTestHandler()

	h.svc.CreatePatient(nil, &Patient{FirstName: "P1", LastName: "L1", MRN: "M1"})
	h.svc.CreatePatient(nil, &Patient{FirstName: "P2", LastName: "L2", MRN: "M2"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListPatients(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_CreatePractitioner(t *testing.T) {
	h, e := newTestHandler()

	body := `{"first_name":"Dr. Sarah","last_name":"Johnson"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/practitioners", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreatePractitioner(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
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
		"POST:/api/v1/patients",
		"GET:/api/v1/patients",
		"GET:/api/v1/patients/:id",
		"POST:/api/v1/practitioners",
		"GET:/fhir/Patient",
		"GET:/fhir/Patient/:id",
		"GET:/fhir/Practitioner",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}

// -- Missing REST Handler Tests --

func TestHandler_UpdatePatient(t *testing.T) {
	h, e := newTestHandler()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN001"}
	h.svc.CreatePatient(nil, p)

	body := `{"first_name":"Jane","last_name":"Doe","mrn":"MRN001"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.UpdatePatient(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_AddPatientContact(t *testing.T) {
	h, e := newTestHandler()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN001"}
	h.svc.CreatePatient(nil, p)

	body := `{"relationship":"parent","name":"Emergency Contact"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.AddPatientContact(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetPatientContacts(t *testing.T) {
	h, e := newTestHandler()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN001"}
	h.svc.CreatePatient(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.GetPatientContacts(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_RemovePatientContact(t *testing.T) {
	h, e := newTestHandler()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN001"}
	h.svc.CreatePatient(nil, p)
	contact := &PatientContact{PatientID: p.ID}
	h.svc.AddPatientContact(nil, contact)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "contact_id")
	c.SetParamValues(p.ID.String(), contact.ID.String())

	err := h.RemovePatientContact(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddPatientIdentifier(t *testing.T) {
	h, e := newTestHandler()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN001"}
	h.svc.CreatePatient(nil, p)

	body := `{"system_uri":"http://hospital.org","value":"12345"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.AddPatientIdentifier(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetPatientIdentifiers(t *testing.T) {
	h, e := newTestHandler()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN001"}
	h.svc.CreatePatient(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.GetPatientIdentifiers(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetPractitioner(t *testing.T) {
	h, e := newTestHandler()
	pr := &Practitioner{FirstName: "Dr", LastName: "Smith"}
	h.svc.CreatePractitioner(nil, pr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(pr.ID.String())

	err := h.GetPractitioner(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListPractitioners(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreatePractitioner(nil, &Practitioner{FirstName: "Dr", LastName: "A"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/practitioners", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListPractitioners(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdatePractitioner(t *testing.T) {
	h, e := newTestHandler()
	pr := &Practitioner{FirstName: "Dr", LastName: "Smith"}
	h.svc.CreatePractitioner(nil, pr)

	body := `{"first_name":"Dr","last_name":"Jones"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(pr.ID.String())

	err := h.UpdatePractitioner(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeletePractitioner(t *testing.T) {
	h, e := newTestHandler()
	pr := &Practitioner{FirstName: "Dr", LastName: "Smith"}
	h.svc.CreatePractitioner(nil, pr)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(pr.ID.String())

	err := h.DeletePractitioner(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddPractitionerRole(t *testing.T) {
	h, e := newTestHandler()
	pr := &Practitioner{FirstName: "Dr", LastName: "Smith"}
	h.svc.CreatePractitioner(nil, pr)

	body := `{"role_code":"physician","specialty_code":"cardiology"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(pr.ID.String())

	err := h.AddPractitionerRole(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetPractitionerRoles(t *testing.T) {
	h, e := newTestHandler()
	pr := &Practitioner{FirstName: "Dr", LastName: "Smith"}
	h.svc.CreatePractitioner(nil, pr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(pr.ID.String())

	err := h.GetPractitionerRoles(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- FHIR Patient Handler Tests --

func TestHandler_SearchPatientsFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreatePatient(nil, &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN001"})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchPatientsFHIR(c)
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

func TestHandler_GetPatientFHIR(t *testing.T) {
	h, e := newTestHandler()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN001"}
	h.svc.CreatePatient(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.FHIRID)
	err := h.GetPatientFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetPatientFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	err := h.GetPatientFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreatePatientFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"first_name":"John","last_name":"Doe","mrn":"MRN-FHIR"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreatePatientFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	if rec.Header().Get("Location") == "" {
		t.Error("expected Location header")
	}
}

func TestHandler_UpdatePatientFHIR(t *testing.T) {
	h, e := newTestHandler()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN001"}
	h.svc.CreatePatient(nil, p)

	body := `{"first_name":"Jane","last_name":"Doe"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.FHIRID)
	err := h.UpdatePatientFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdatePatientFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	err := h.UpdatePatientFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_DeletePatientFHIR(t *testing.T) {
	h, e := newTestHandler()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN001"}
	h.svc.CreatePatient(nil, p)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.FHIRID)
	err := h.DeletePatientFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchPatientFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN001"}
	h.svc.CreatePatient(nil, p)

	body := `{"active":true}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.FHIRID)
	err := h.PatchPatientFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_PatchPatientFHIR_JSONPatch(t *testing.T) {
	h, e := newTestHandler()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN001"}
	h.svc.CreatePatient(nil, p)

	body := `[{"op":"replace","path":"/active","value":true}]`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.FHIRID)
	err := h.PatchPatientFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_PatchPatientFHIR_UnsupportedMediaType(t *testing.T) {
	h, e := newTestHandler()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN001"}
	h.svc.CreatePatient(nil, p)

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.FHIRID)
	err := h.PatchPatientFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415, got %d", rec.Code)
	}
}

func TestHandler_VreadPatientFHIR(t *testing.T) {
	h, e := newTestHandler()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN001"}
	h.svc.CreatePatient(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(p.FHIRID, "1")
	err := h.VreadPatientFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("ETag") == "" {
		t.Error("expected ETag header")
	}
}

func TestHandler_HistoryPatientFHIR(t *testing.T) {
	h, e := newTestHandler()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN001"}
	h.svc.CreatePatient(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.FHIRID)
	err := h.HistoryPatientFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle["type"] != "history" {
		t.Errorf("expected history, got %v", bundle["type"])
	}
}

// -- FHIR Practitioner Handler Tests --

func TestHandler_SearchPractitionersFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreatePractitioner(nil, &Practitioner{FirstName: "Dr", LastName: "Smith"})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Practitioner", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchPractitionersFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetPractitionerFHIR(t *testing.T) {
	h, e := newTestHandler()
	pr := &Practitioner{FirstName: "Dr", LastName: "Smith"}
	h.svc.CreatePractitioner(nil, pr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(pr.FHIRID)
	err := h.GetPractitionerFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetPractitionerFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	err := h.GetPractitionerFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreatePractitionerFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"first_name":"Dr","last_name":"New"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Practitioner", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreatePractitionerFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	if rec.Header().Get("Location") == "" {
		t.Error("expected Location header")
	}
}

func TestHandler_UpdatePractitionerFHIR(t *testing.T) {
	h, e := newTestHandler()
	pr := &Practitioner{FirstName: "Dr", LastName: "Smith"}
	h.svc.CreatePractitioner(nil, pr)

	body := `{"first_name":"Dr","last_name":"Jones"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(pr.FHIRID)
	err := h.UpdatePractitionerFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeletePractitionerFHIR(t *testing.T) {
	h, e := newTestHandler()
	pr := &Practitioner{FirstName: "Dr", LastName: "Smith"}
	h.svc.CreatePractitioner(nil, pr)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(pr.FHIRID)
	err := h.DeletePractitionerFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchPractitionerFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	pr := &Practitioner{FirstName: "Dr", LastName: "Smith"}
	h.svc.CreatePractitioner(nil, pr)

	body := `{"active":true}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(pr.FHIRID)
	err := h.PatchPractitionerFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadPractitionerFHIR(t *testing.T) {
	h, e := newTestHandler()
	pr := &Practitioner{FirstName: "Dr", LastName: "Smith"}
	h.svc.CreatePractitioner(nil, pr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(pr.FHIRID, "1")
	err := h.VreadPractitionerFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryPractitionerFHIR(t *testing.T) {
	h, e := newTestHandler()
	pr := &Practitioner{FirstName: "Dr", LastName: "Smith"}
	h.svc.CreatePractitioner(nil, pr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(pr.FHIRID)
	err := h.HistoryPractitionerFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- applyPatientPatch Tests --

func TestApplyPatientPatch_BasicFields(t *testing.T) {
	p := &Patient{FirstName: "John", LastName: "Doe", Active: false}

	patched := map[string]interface{}{
		"active":    true,
		"gender":    "male",
		"birthDate": "1990-05-15",
	}
	applyPatientPatch(p, patched)

	if p.Active != true {
		t.Errorf("expected Active=true, got %v", p.Active)
	}
	if p.Gender == nil || *p.Gender != "male" {
		t.Errorf("expected Gender=male, got %v", p.Gender)
	}
	if p.BirthDate == nil {
		t.Fatal("expected BirthDate to be set")
	}
	if p.BirthDate.Format("2006-01-02") != "1990-05-15" {
		t.Errorf("expected BirthDate=1990-05-15, got %s", p.BirthDate.Format("2006-01-02"))
	}
	// Unchanged fields
	if p.FirstName != "John" {
		t.Errorf("expected FirstName=John, got %s", p.FirstName)
	}
	if p.LastName != "Doe" {
		t.Errorf("expected LastName=Doe, got %s", p.LastName)
	}
}

func TestApplyPatientPatch_Name(t *testing.T) {
	p := &Patient{FirstName: "Old", LastName: "Name"}

	patched := map[string]interface{}{
		"name": []interface{}{
			map[string]interface{}{
				"family": "Smith",
				"given":  []interface{}{"Jane", "Marie"},
				"prefix": []interface{}{"Dr."},
				"suffix": []interface{}{"Jr."},
			},
		},
	}
	applyPatientPatch(p, patched)

	if p.LastName != "Smith" {
		t.Errorf("expected LastName=Smith, got %s", p.LastName)
	}
	if p.FirstName != "Jane" {
		t.Errorf("expected FirstName=Jane, got %s", p.FirstName)
	}
	if p.MiddleName == nil || *p.MiddleName != "Marie" {
		t.Errorf("expected MiddleName=Marie, got %v", p.MiddleName)
	}
	if p.Prefix == nil || *p.Prefix != "Dr." {
		t.Errorf("expected Prefix=Dr., got %v", p.Prefix)
	}
	if p.Suffix == nil || *p.Suffix != "Jr." {
		t.Errorf("expected Suffix=Jr., got %v", p.Suffix)
	}
}

func TestApplyPatientPatch_Telecom(t *testing.T) {
	p := &Patient{FirstName: "John", LastName: "Doe"}

	patched := map[string]interface{}{
		"telecom": []interface{}{
			map[string]interface{}{
				"system": "phone",
				"value":  "555-0100",
				"use":    "mobile",
			},
			map[string]interface{}{
				"system": "phone",
				"value":  "555-0101",
				"use":    "home",
			},
			map[string]interface{}{
				"system": "email",
				"value":  "john@example.com",
			},
		},
	}
	applyPatientPatch(p, patched)

	if p.PhoneMobile == nil || *p.PhoneMobile != "555-0100" {
		t.Errorf("expected PhoneMobile=555-0100, got %v", p.PhoneMobile)
	}
	if p.PhoneHome == nil || *p.PhoneHome != "555-0101" {
		t.Errorf("expected PhoneHome=555-0101, got %v", p.PhoneHome)
	}
	if p.Email == nil || *p.Email != "john@example.com" {
		t.Errorf("expected Email=john@example.com, got %v", p.Email)
	}
}

func TestApplyPatientPatch_Address(t *testing.T) {
	p := &Patient{FirstName: "John", LastName: "Doe"}

	patched := map[string]interface{}{
		"address": []interface{}{
			map[string]interface{}{
				"use":        "home",
				"line":       []interface{}{"123 Main St", "Apt 4B"},
				"city":       "Springfield",
				"district":   "Clark",
				"state":      "IL",
				"postalCode": "62701",
				"country":    "US",
			},
		},
	}
	applyPatientPatch(p, patched)

	if p.AddressUse == nil || *p.AddressUse != "home" {
		t.Errorf("expected AddressUse=home, got %v", p.AddressUse)
	}
	if p.AddressLine1 == nil || *p.AddressLine1 != "123 Main St" {
		t.Errorf("expected AddressLine1=123 Main St, got %v", p.AddressLine1)
	}
	if p.AddressLine2 == nil || *p.AddressLine2 != "Apt 4B" {
		t.Errorf("expected AddressLine2=Apt 4B, got %v", p.AddressLine2)
	}
	if p.City == nil || *p.City != "Springfield" {
		t.Errorf("expected City=Springfield, got %v", p.City)
	}
	if p.District == nil || *p.District != "Clark" {
		t.Errorf("expected District=Clark, got %v", p.District)
	}
	if p.State == nil || *p.State != "IL" {
		t.Errorf("expected State=IL, got %v", p.State)
	}
	if p.PostalCode == nil || *p.PostalCode != "62701" {
		t.Errorf("expected PostalCode=62701, got %v", p.PostalCode)
	}
	if p.Country == nil || *p.Country != "US" {
		t.Errorf("expected Country=US, got %v", p.Country)
	}
}

func TestApplyPatientPatch_EmptyMap(t *testing.T) {
	p := &Patient{FirstName: "John", LastName: "Doe", Active: true}

	patched := map[string]interface{}{}
	applyPatientPatch(p, patched)

	if p.FirstName != "John" {
		t.Errorf("expected FirstName=John, got %s", p.FirstName)
	}
	if p.LastName != "Doe" {
		t.Errorf("expected LastName=Doe, got %s", p.LastName)
	}
	if p.Active != true {
		t.Errorf("expected Active=true, got %v", p.Active)
	}
	if p.Gender != nil {
		t.Errorf("expected Gender=nil, got %v", p.Gender)
	}
}

// -- applyPractitionerPatch Tests --

func TestApplyPractitionerPatch_BasicFields(t *testing.T) {
	pr := &Practitioner{FirstName: "Dr", LastName: "Smith", Active: false}

	patched := map[string]interface{}{
		"active": true,
		"gender": "female",
		"name": []interface{}{
			map[string]interface{}{
				"family": "Johnson",
				"given":  []interface{}{"Sarah", "Ann"},
				"prefix": []interface{}{"Dr."},
				"suffix": []interface{}{"MD"},
			},
		},
		"telecom": []interface{}{
			map[string]interface{}{
				"system": "phone",
				"value":  "555-9999",
			},
			map[string]interface{}{
				"system": "email",
				"value":  "sarah@hospital.com",
			},
		},
	}
	applyPractitionerPatch(pr, patched)

	if pr.Active != true {
		t.Errorf("expected Active=true, got %v", pr.Active)
	}
	if pr.Gender == nil || *pr.Gender != "female" {
		t.Errorf("expected Gender=female, got %v", pr.Gender)
	}
	if pr.LastName != "Johnson" {
		t.Errorf("expected LastName=Johnson, got %s", pr.LastName)
	}
	if pr.FirstName != "Sarah" {
		t.Errorf("expected FirstName=Sarah, got %s", pr.FirstName)
	}
	if pr.MiddleName == nil || *pr.MiddleName != "Ann" {
		t.Errorf("expected MiddleName=Ann, got %v", pr.MiddleName)
	}
	if pr.Prefix == nil || *pr.Prefix != "Dr." {
		t.Errorf("expected Prefix=Dr., got %v", pr.Prefix)
	}
	if pr.Suffix == nil || *pr.Suffix != "MD" {
		t.Errorf("expected Suffix=MD, got %v", pr.Suffix)
	}
	if pr.Phone == nil || *pr.Phone != "555-9999" {
		t.Errorf("expected Phone=555-9999, got %v", pr.Phone)
	}
	if pr.Email == nil || *pr.Email != "sarah@hospital.com" {
		t.Errorf("expected Email=sarah@hospital.com, got %v", pr.Email)
	}
}
