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

// -- REST: Update & List Tests --

func TestHandler_UpdateConsent(t *testing.T) {
	h, e := newTestHandler()
	consent := &Consent{PatientID: uuid.New()}
	h.svc.CreateConsent(nil, consent)

	body := `{"status":"active"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(consent.ID.String())

	err := h.UpdateConsent(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListConsents(t *testing.T) {
	h, e := newTestHandler()
	pid := uuid.New()
	h.svc.CreateConsent(nil, &Consent{PatientID: pid})

	req := httptest.NewRequest(http.MethodGet, "/?patient_id="+pid.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListConsents(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateDocumentReference(t *testing.T) {
	h, e := newTestHandler()
	doc := &DocumentReference{PatientID: uuid.New()}
	h.svc.CreateDocumentReference(nil, doc)

	body := `{"status":"superseded"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(doc.ID.String())

	err := h.UpdateDocumentReference(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListDocumentReferences(t *testing.T) {
	h, e := newTestHandler()
	pid := uuid.New()
	h.svc.CreateDocumentReference(nil, &DocumentReference{PatientID: pid})

	req := httptest.NewRequest(http.MethodGet, "/?patient_id="+pid.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListDocumentReferences(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateClinicalNote(t *testing.T) {
	h, e := newTestHandler()
	note := &ClinicalNote{PatientID: uuid.New(), AuthorID: uuid.New(), NoteType: "progress"}
	h.svc.CreateClinicalNote(nil, note)

	body := `{"status":"final"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(note.ID.String())

	err := h.UpdateClinicalNote(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListClinicalNotes(t *testing.T) {
	h, e := newTestHandler()
	pid := uuid.New()
	h.svc.CreateClinicalNote(nil, &ClinicalNote{PatientID: pid, AuthorID: uuid.New(), NoteType: "progress"})

	req := httptest.NewRequest(http.MethodGet, "/?patient_id="+pid.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListClinicalNotes(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateComposition(t *testing.T) {
	h, e := newTestHandler()
	comp := &Composition{PatientID: uuid.New()}
	h.svc.CreateComposition(nil, comp)

	body := `{"status":"final"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(comp.ID.String())

	err := h.UpdateComposition(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListCompositions(t *testing.T) {
	h, e := newTestHandler()
	pid := uuid.New()
	h.svc.CreateComposition(nil, &Composition{PatientID: pid})

	req := httptest.NewRequest(http.MethodGet, "/?patient_id="+pid.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListCompositions(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetSections(t *testing.T) {
	h, e := newTestHandler()
	comp := &Composition{PatientID: uuid.New()}
	h.svc.CreateComposition(nil, comp)
	title := "HPI"
	h.svc.AddCompositionSection(nil, &CompositionSection{CompositionID: comp.ID, Title: &title})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(comp.ID.String())

	err := h.GetSections(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- FHIR Consent Handlers --

func TestHandler_SearchConsentsFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateConsent(nil, &Consent{PatientID: uuid.New()})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchConsentsFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetConsentFHIR(t *testing.T) {
	h, e := newTestHandler()
	consent := &Consent{PatientID: uuid.New()}
	h.svc.CreateConsent(nil, consent)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(consent.FHIRID)

	err := h.GetConsentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetConsentFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.GetConsentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateConsentFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateConsentFHIR(c)
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

func TestHandler_UpdateConsentFHIR(t *testing.T) {
	h, e := newTestHandler()
	consent := &Consent{PatientID: uuid.New()}
	h.svc.CreateConsent(nil, consent)

	body := `{"status":"active","patient_id":"` + consent.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(consent.FHIRID)

	err := h.UpdateConsentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateConsentFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.UpdateConsentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_DeleteConsentFHIR(t *testing.T) {
	h, e := newTestHandler()
	consent := &Consent{PatientID: uuid.New()}
	h.svc.CreateConsent(nil, consent)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(consent.FHIRID)

	err := h.DeleteConsentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_DeleteConsentFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.DeleteConsentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_PatchConsentFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	consent := &Consent{PatientID: uuid.New()}
	h.svc.CreateConsent(nil, consent)

	body := `{"status":"active"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(consent.FHIRID)

	err := h.PatchConsentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_PatchConsentFHIR_JSONPatch(t *testing.T) {
	h, e := newTestHandler()
	consent := &Consent{PatientID: uuid.New()}
	h.svc.CreateConsent(nil, consent)

	body := `[{"op":"replace","path":"/status","value":"active"}]`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(consent.FHIRID)

	err := h.PatchConsentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_PatchConsentFHIR_UnsupportedMediaType(t *testing.T) {
	h, e := newTestHandler()
	consent := &Consent{PatientID: uuid.New()}
	h.svc.CreateConsent(nil, consent)

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(consent.FHIRID)

	err := h.PatchConsentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415, got %d", rec.Code)
	}
}

func TestHandler_PatchConsentFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"status":"active"}`))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.PatchConsentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_VreadConsentFHIR(t *testing.T) {
	h, e := newTestHandler()
	consent := &Consent{PatientID: uuid.New()}
	h.svc.CreateConsent(nil, consent)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(consent.FHIRID, "1")

	err := h.VreadConsentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryConsentFHIR(t *testing.T) {
	h, e := newTestHandler()
	consent := &Consent{PatientID: uuid.New()}
	h.svc.CreateConsent(nil, consent)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(consent.FHIRID)

	err := h.HistoryConsentFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- FHIR DocumentReference Handlers --

func TestHandler_SearchDocumentReferencesFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateDocumentReference(nil, &DocumentReference{PatientID: uuid.New()})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchDocumentReferencesFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetDocumentReferenceFHIR(t *testing.T) {
	h, e := newTestHandler()
	doc := &DocumentReference{PatientID: uuid.New()}
	h.svc.CreateDocumentReference(nil, doc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(doc.FHIRID)

	err := h.GetDocumentReferenceFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetDocumentReferenceFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.GetDocumentReferenceFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateDocumentReferenceFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateDocumentReferenceFHIR(c)
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

func TestHandler_UpdateDocumentReferenceFHIR(t *testing.T) {
	h, e := newTestHandler()
	doc := &DocumentReference{PatientID: uuid.New()}
	h.svc.CreateDocumentReference(nil, doc)

	body := `{"status":"superseded","patient_id":"` + doc.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(doc.FHIRID)

	err := h.UpdateDocumentReferenceFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateDocumentReferenceFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.UpdateDocumentReferenceFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_DeleteDocumentReferenceFHIR(t *testing.T) {
	h, e := newTestHandler()
	doc := &DocumentReference{PatientID: uuid.New()}
	h.svc.CreateDocumentReference(nil, doc)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(doc.FHIRID)

	err := h.DeleteDocumentReferenceFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchDocumentReferenceFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	doc := &DocumentReference{PatientID: uuid.New()}
	h.svc.CreateDocumentReference(nil, doc)

	body := `{"status":"superseded"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(doc.FHIRID)

	err := h.PatchDocumentReferenceFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadDocumentReferenceFHIR(t *testing.T) {
	h, e := newTestHandler()
	doc := &DocumentReference{PatientID: uuid.New()}
	h.svc.CreateDocumentReference(nil, doc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(doc.FHIRID, "1")

	err := h.VreadDocumentReferenceFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryDocumentReferenceFHIR(t *testing.T) {
	h, e := newTestHandler()
	doc := &DocumentReference{PatientID: uuid.New()}
	h.svc.CreateDocumentReference(nil, doc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(doc.FHIRID)

	err := h.HistoryDocumentReferenceFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- FHIR Composition Handlers --

func TestHandler_SearchCompositionsFHIR(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchCompositionsFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetCompositionFHIR(t *testing.T) {
	h, e := newTestHandler()
	comp := &Composition{PatientID: uuid.New()}
	h.svc.CreateComposition(nil, comp)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(comp.FHIRID)

	err := h.GetCompositionFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetCompositionFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.GetCompositionFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateCompositionFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateCompositionFHIR(c)
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

func TestHandler_UpdateCompositionFHIR(t *testing.T) {
	h, e := newTestHandler()
	comp := &Composition{PatientID: uuid.New()}
	h.svc.CreateComposition(nil, comp)

	body := `{"status":"final","patient_id":"` + comp.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(comp.FHIRID)

	err := h.UpdateCompositionFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteCompositionFHIR(t *testing.T) {
	h, e := newTestHandler()
	comp := &Composition{PatientID: uuid.New()}
	h.svc.CreateComposition(nil, comp)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(comp.FHIRID)

	err := h.DeleteCompositionFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchCompositionFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	comp := &Composition{PatientID: uuid.New()}
	h.svc.CreateComposition(nil, comp)

	body := `{"status":"final"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(comp.FHIRID)

	err := h.PatchCompositionFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadCompositionFHIR(t *testing.T) {
	h, e := newTestHandler()
	comp := &Composition{PatientID: uuid.New()}
	h.svc.CreateComposition(nil, comp)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(comp.FHIRID, "1")

	err := h.VreadCompositionFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryCompositionFHIR(t *testing.T) {
	h, e := newTestHandler()
	comp := &Composition{PatientID: uuid.New()}
	h.svc.CreateComposition(nil, comp)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(comp.FHIRID)

	err := h.HistoryCompositionFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
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
