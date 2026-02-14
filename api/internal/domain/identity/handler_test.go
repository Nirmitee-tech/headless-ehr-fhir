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
