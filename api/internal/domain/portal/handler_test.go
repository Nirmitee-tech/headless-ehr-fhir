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
