package documents

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

// -- Consent Handler Tests --

func TestHandler_CreateConsent(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateConsent(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateConsent_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateConsent(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetConsent(t *testing.T) {
	h, e := newTestHandler()
	consent := &Consent{PatientID: uuid.New()}
	h.svc.CreateConsent(nil, consent)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(consent.ID.String())

	err := h.GetConsent(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetConsent_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetConsent(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_DeleteConsent(t *testing.T) {
	h, e := newTestHandler()
	consent := &Consent{PatientID: uuid.New()}
	h.svc.CreateConsent(nil, consent)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(consent.ID.String())

	err := h.DeleteConsent(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// -- DocumentReference Handler Tests --

func TestHandler_CreateDocumentReference(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateDocumentReference(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetDocumentReference(t *testing.T) {
	h, e := newTestHandler()
	doc := &DocumentReference{PatientID: uuid.New()}
	h.svc.CreateDocumentReference(nil, doc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(doc.ID.String())

	err := h.GetDocumentReference(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteDocumentReference(t *testing.T) {
	h, e := newTestHandler()
	doc := &DocumentReference{PatientID: uuid.New()}
	h.svc.CreateDocumentReference(nil, doc)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(doc.ID.String())

	err := h.DeleteDocumentReference(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// -- ClinicalNote Handler Tests --

func TestHandler_CreateClinicalNote(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","author_id":"` + uuid.New().String() + `","note_type":"progress"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateClinicalNote(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateClinicalNote_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"note_type":"progress"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateClinicalNote(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetClinicalNote(t *testing.T) {
	h, e := newTestHandler()
	note := &ClinicalNote{PatientID: uuid.New(), AuthorID: uuid.New(), NoteType: "progress"}
	h.svc.CreateClinicalNote(nil, note)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(note.ID.String())

	err := h.GetClinicalNote(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteClinicalNote(t *testing.T) {
	h, e := newTestHandler()
	note := &ClinicalNote{PatientID: uuid.New(), AuthorID: uuid.New(), NoteType: "progress"}
	h.svc.CreateClinicalNote(nil, note)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(note.ID.String())

	err := h.DeleteClinicalNote(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// -- Composition Handler Tests --

func TestHandler_CreateComposition(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateComposition(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetComposition(t *testing.T) {
	h, e := newTestHandler()
	comp := &Composition{PatientID: uuid.New()}
	h.svc.CreateComposition(nil, comp)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(comp.ID.String())

	err := h.GetComposition(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteComposition(t *testing.T) {
	h, e := newTestHandler()
	comp := &Composition{PatientID: uuid.New()}
	h.svc.CreateComposition(nil, comp)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(comp.ID.String())

	err := h.DeleteComposition(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddSection(t *testing.T) {
	h, e := newTestHandler()
	comp := &Composition{PatientID: uuid.New()}
	h.svc.CreateComposition(nil, comp)

	body := `{"title":"HPI","code_value":"10164-2"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(comp.ID.String())

	err := h.AddSection(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var result CompositionSection
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result.CompositionID != comp.ID {
		t.Error("expected composition_id to match")
	}
}

// -- Route Registration --

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
		"POST:/api/v1/consents",
		"GET:/api/v1/consents/:id",
		"POST:/api/v1/document-references",
		"POST:/api/v1/clinical-notes",
		"POST:/api/v1/compositions",
		"POST:/api/v1/compositions/:id/sections",
		"GET:/fhir/Consent",
		"GET:/fhir/DocumentReference",
		"GET:/fhir/Composition",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
