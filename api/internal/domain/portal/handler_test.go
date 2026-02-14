package portal

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

// ── Portal Account Handlers ──

func TestHandler_CreatePortalAccount(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreatePortalAccount(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreatePortalAccount_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreatePortalAccount(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetPortalAccount(t *testing.T) {
	h, e := newTestHandler()
	a := &PortalAccount{PatientID: uuid.New()}
	h.svc.CreatePortalAccount(nil, a)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())
	err := h.GetPortalAccount(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetPortalAccount_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetPortalAccount(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_DeletePortalAccount(t *testing.T) {
	h, e := newTestHandler()
	a := &PortalAccount{PatientID: uuid.New()}
	h.svc.CreatePortalAccount(nil, a)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())
	err := h.DeletePortalAccount(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// ── Portal Message Handlers ──

func TestHandler_CreatePortalMessage(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","body":"Hello doc"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreatePortalMessage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreatePortalMessage_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreatePortalMessage(c)
	if err == nil {
		t.Error("expected error for missing body")
	}
}

func TestHandler_GetPortalMessage(t *testing.T) {
	h, e := newTestHandler()
	m := &PortalMessage{PatientID: uuid.New(), Body: "Hello"}
	h.svc.CreatePortalMessage(nil, m)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.ID.String())
	err := h.GetPortalMessage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeletePortalMessage(t *testing.T) {
	h, e := newTestHandler()
	m := &PortalMessage{PatientID: uuid.New(), Body: "Hello"}
	h.svc.CreatePortalMessage(nil, m)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.ID.String())
	err := h.DeletePortalMessage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// ── Questionnaire Handlers ──

func TestHandler_CreateQuestionnaire(t *testing.T) {
	h, e := newTestHandler()
	body := `{"name":"PHQ-9"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateQuestionnaire(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateQuestionnaire_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateQuestionnaire(c)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestHandler_GetQuestionnaire(t *testing.T) {
	h, e := newTestHandler()
	q := &Questionnaire{Name: "PHQ-9"}
	h.svc.CreateQuestionnaire(nil, q)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(q.ID.String())
	err := h.GetQuestionnaire(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetQuestionnaire_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetQuestionnaire(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_DeleteQuestionnaire(t *testing.T) {
	h, e := newTestHandler()
	q := &Questionnaire{Name: "PHQ-9"}
	h.svc.CreateQuestionnaire(nil, q)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(q.ID.String())
	err := h.DeleteQuestionnaire(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddQuestionnaireItem(t *testing.T) {
	h, e := newTestHandler()
	qID := uuid.New()
	body := `{"link_id":"q1","text":"How are you?"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(qID.String())
	err := h.AddQuestionnaireItem(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_AddQuestionnaireItem_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	qID := uuid.New()
	body := `{"link_id":"q1"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(qID.String())
	err := h.AddQuestionnaireItem(c)
	if err == nil {
		t.Error("expected error for missing text")
	}
}

// ── Questionnaire Response Handlers ──

func TestHandler_CreateQuestionnaireResponse(t *testing.T) {
	h, e := newTestHandler()
	body := `{"questionnaire_id":"` + uuid.New().String() + `","patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateQuestionnaireResponse(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateQuestionnaireResponse_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"questionnaire_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateQuestionnaireResponse(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetQuestionnaireResponse(t *testing.T) {
	h, e := newTestHandler()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateQuestionnaireResponse(nil, qr)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(qr.ID.String())
	err := h.GetQuestionnaireResponse(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteQuestionnaireResponse(t *testing.T) {
	h, e := newTestHandler()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateQuestionnaireResponse(nil, qr)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(qr.ID.String())
	err := h.DeleteQuestionnaireResponse(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddQuestionnaireResponseItem(t *testing.T) {
	h, e := newTestHandler()
	respID := uuid.New()
	body := `{"link_id":"q1"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(respID.String())
	err := h.AddQuestionnaireResponseItem(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_AddQuestionnaireResponseItem_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	respID := uuid.New()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(respID.String())
	err := h.AddQuestionnaireResponseItem(c)
	if err == nil {
		t.Error("expected error for missing link_id")
	}
}

// ── Patient Checkin Handlers ──

func TestHandler_CreatePatientCheckin(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreatePatientCheckin(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreatePatientCheckin_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreatePatientCheckin(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetPatientCheckin(t *testing.T) {
	h, e := newTestHandler()
	ch := &PatientCheckin{PatientID: uuid.New()}
	h.svc.CreatePatientCheckin(nil, ch)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ch.ID.String())
	err := h.GetPatientCheckin(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeletePatientCheckin(t *testing.T) {
	h, e := newTestHandler()
	ch := &PatientCheckin{PatientID: uuid.New()}
	h.svc.CreatePatientCheckin(nil, ch)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ch.ID.String())
	err := h.DeletePatientCheckin(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// ── List/Update Tests for PortalAccount ──

func TestHandler_ListPortalAccounts(t *testing.T) {
	h, e := newTestHandler()
	a := &PortalAccount{PatientID: uuid.New()}
	h.svc.CreatePortalAccount(nil, a)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListPortalAccounts(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdatePortalAccount(t *testing.T) {
	h, e := newTestHandler()
	a := &PortalAccount{PatientID: uuid.New()}
	h.svc.CreatePortalAccount(nil, a)
	body := `{"patient_id":"` + a.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(a.ID.String())
	err := h.UpdatePortalAccount(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// ── List/Update Tests for PortalMessage ──

func TestHandler_ListPortalMessages(t *testing.T) {
	h, e := newTestHandler()
	m := &PortalMessage{PatientID: uuid.New(), Body: "Test"}
	h.svc.CreatePortalMessage(nil, m)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListPortalMessages(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdatePortalMessage(t *testing.T) {
	h, e := newTestHandler()
	m := &PortalMessage{PatientID: uuid.New(), Body: "Hello"}
	h.svc.CreatePortalMessage(nil, m)
	body := `{"patient_id":"` + m.PatientID.String() + `","body":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.ID.String())
	err := h.UpdatePortalMessage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// ── List/Update Tests for Questionnaire ──

func TestHandler_ListQuestionnaires(t *testing.T) {
	h, e := newTestHandler()
	q := &Questionnaire{Name: "PHQ-9"}
	h.svc.CreateQuestionnaire(nil, q)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListQuestionnaires(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateQuestionnaire(t *testing.T) {
	h, e := newTestHandler()
	q := &Questionnaire{Name: "PHQ-9"}
	h.svc.CreateQuestionnaire(nil, q)
	body := `{"name":"GAD-7"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(q.ID.String())
	err := h.UpdateQuestionnaire(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetQuestionnaireItems(t *testing.T) {
	h, e := newTestHandler()
	qID := uuid.New()
	item := &QuestionnaireItem{QuestionnaireID: qID, LinkID: "q1", Text: "How are you?"}
	h.svc.AddQuestionnaireItem(nil, item)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(qID.String())
	err := h.GetQuestionnaireItems(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// ── List/Update Tests for QuestionnaireResponse ──

func TestHandler_ListQuestionnaireResponses(t *testing.T) {
	h, e := newTestHandler()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateQuestionnaireResponse(nil, qr)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListQuestionnaireResponses(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateQuestionnaireResponse(t *testing.T) {
	h, e := newTestHandler()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateQuestionnaireResponse(nil, qr)
	body := `{"questionnaire_id":"` + qr.QuestionnaireID.String() + `","patient_id":"` + qr.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(qr.ID.String())
	err := h.UpdateQuestionnaireResponse(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetQuestionnaireResponseItems(t *testing.T) {
	h, e := newTestHandler()
	respID := uuid.New()
	item := &QuestionnaireResponseItem{ResponseID: respID, LinkID: "q1"}
	h.svc.AddQuestionnaireResponseItem(nil, item)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(respID.String())
	err := h.GetQuestionnaireResponseItems(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// ── List/Update Tests for PatientCheckin ──

func TestHandler_ListPatientCheckins(t *testing.T) {
	h, e := newTestHandler()
	ch := &PatientCheckin{PatientID: uuid.New()}
	h.svc.CreatePatientCheckin(nil, ch)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListPatientCheckins(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdatePatientCheckin(t *testing.T) {
	h, e := newTestHandler()
	ch := &PatientCheckin{PatientID: uuid.New()}
	h.svc.CreatePatientCheckin(nil, ch)
	body := `{"patient_id":"` + ch.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ch.ID.String())
	err := h.UpdatePatientCheckin(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// ── FHIR Questionnaire Endpoints ──

func TestHandler_SearchQuestionnairesFHIR(t *testing.T) {
	h, e := newTestHandler()
	q := &Questionnaire{Name: "PHQ-9"}
	h.svc.CreateQuestionnaire(nil, q)
	req := httptest.NewRequest(http.MethodGet, "/fhir/Questionnaire", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchQuestionnairesFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Bundle") {
		t.Error("expected Bundle in response")
	}
}

func TestHandler_GetQuestionnaireFHIR(t *testing.T) {
	h, e := newTestHandler()
	q := &Questionnaire{Name: "PHQ-9"}
	h.svc.CreateQuestionnaire(nil, q)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(q.FHIRID)
	err := h.GetQuestionnaireFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetQuestionnaireFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	_ = h.GetQuestionnaireFHIR(c)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateQuestionnaireFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"name":"GAD-7"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateQuestionnaireFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc == "" || !strings.Contains(loc, "/fhir/Questionnaire/") {
		t.Errorf("expected Location header, got %q", loc)
	}
}

func TestHandler_UpdateQuestionnaireFHIR(t *testing.T) {
	h, e := newTestHandler()
	q := &Questionnaire{Name: "PHQ-9"}
	h.svc.CreateQuestionnaire(nil, q)
	body := `{"name":"GAD-7"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(q.FHIRID)
	err := h.UpdateQuestionnaireFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteQuestionnaireFHIR(t *testing.T) {
	h, e := newTestHandler()
	q := &Questionnaire{Name: "PHQ-9"}
	h.svc.CreateQuestionnaire(nil, q)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(q.FHIRID)
	err := h.DeleteQuestionnaireFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchQuestionnaireFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	q := &Questionnaire{Name: "PHQ-9"}
	h.svc.CreateQuestionnaire(nil, q)
	body := `{"status":"retired"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(q.FHIRID)
	err := h.PatchQuestionnaireFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadQuestionnaireFHIR(t *testing.T) {
	h, e := newTestHandler()
	q := &Questionnaire{Name: "PHQ-9"}
	h.svc.CreateQuestionnaire(nil, q)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(q.FHIRID, "1")
	err := h.VreadQuestionnaireFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryQuestionnaireFHIR(t *testing.T) {
	h, e := newTestHandler()
	q := &Questionnaire{Name: "PHQ-9"}
	h.svc.CreateQuestionnaire(nil, q)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(q.FHIRID)
	err := h.HistoryQuestionnaireFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Bundle") {
		t.Error("expected Bundle in history response")
	}
}

// ── FHIR QuestionnaireResponse Endpoints ──

func TestHandler_SearchQuestionnaireResponsesFHIR(t *testing.T) {
	h, e := newTestHandler()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateQuestionnaireResponse(nil, qr)
	req := httptest.NewRequest(http.MethodGet, "/fhir/QuestionnaireResponse", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchQuestionnaireResponsesFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetQuestionnaireResponseFHIR(t *testing.T) {
	h, e := newTestHandler()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateQuestionnaireResponse(nil, qr)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(qr.FHIRID)
	err := h.GetQuestionnaireResponseFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetQuestionnaireResponseFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	_ = h.GetQuestionnaireResponseFHIR(c)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateQuestionnaireResponseFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"questionnaire_id":"` + uuid.New().String() + `","patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateQuestionnaireResponseFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc == "" || !strings.Contains(loc, "/fhir/QuestionnaireResponse/") {
		t.Errorf("expected Location header, got %q", loc)
	}
}

func TestHandler_UpdateQuestionnaireResponseFHIR(t *testing.T) {
	h, e := newTestHandler()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateQuestionnaireResponse(nil, qr)
	body := `{"questionnaire_id":"` + qr.QuestionnaireID.String() + `","patient_id":"` + qr.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(qr.FHIRID)
	err := h.UpdateQuestionnaireResponseFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteQuestionnaireResponseFHIR(t *testing.T) {
	h, e := newTestHandler()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateQuestionnaireResponse(nil, qr)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(qr.FHIRID)
	err := h.DeleteQuestionnaireResponseFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchQuestionnaireResponseFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateQuestionnaireResponse(nil, qr)
	body := `{"status":"completed"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(qr.FHIRID)
	err := h.PatchQuestionnaireResponseFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadQuestionnaireResponseFHIR(t *testing.T) {
	h, e := newTestHandler()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateQuestionnaireResponse(nil, qr)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(qr.FHIRID, "1")
	err := h.VreadQuestionnaireResponseFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryQuestionnaireResponseFHIR(t *testing.T) {
	h, e := newTestHandler()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateQuestionnaireResponse(nil, qr)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(qr.FHIRID)
	err := h.HistoryQuestionnaireResponseFHIR(c)
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
	fhirGroup := e.Group("/fhir")
	h.RegisterRoutes(api, fhirGroup)

	routes := e.Routes()
	if len(routes) == 0 {
		t.Error("expected routes to be registered")
	}
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}
	expected := []string{
		"POST:/api/v1/portal-accounts",
		"GET:/api/v1/portal-accounts",
		"GET:/api/v1/portal-accounts/:id",
		"POST:/api/v1/portal-messages",
		"GET:/api/v1/portal-messages",
		"POST:/api/v1/questionnaires",
		"GET:/api/v1/questionnaires",
		"GET:/api/v1/questionnaires/:id",
		"GET:/api/v1/questionnaires/:id/items",
		"POST:/api/v1/questionnaire-responses",
		"GET:/api/v1/questionnaire-responses",
		"POST:/api/v1/patient-checkins",
		"GET:/api/v1/patient-checkins",
		"GET:/fhir/Questionnaire",
		"GET:/fhir/Questionnaire/:id",
		"POST:/fhir/Questionnaire",
		"PUT:/fhir/Questionnaire/:id",
		"DELETE:/fhir/Questionnaire/:id",
		"PATCH:/fhir/Questionnaire/:id",
		"GET:/fhir/QuestionnaireResponse",
		"GET:/fhir/QuestionnaireResponse/:id",
		"POST:/fhir/QuestionnaireResponse",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
