package surgery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func newTestHandler() (*Handler, *echo.Echo) {
	svc := newTestService()
	h := NewHandler(svc)
	e := echo.New()
	return h, e
}

func TestHandler_CreateORRoom(t *testing.T) {
	h, e := newTestHandler()
	body := `{"name":"OR-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/or-rooms", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateORRoom(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateORRoom_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/or-rooms", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateORRoom(c)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestHandler_GetORRoom(t *testing.T) {
	h, e := newTestHandler()
	r := &ORRoom{Name: "OR-1"}
	h.svc.CreateORRoom(nil, r)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(r.ID.String())

	err := h.GetORRoom(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetORRoom_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetORRoom(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_GetORRoom_InvalidID(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.GetORRoom(c)
	if err == nil {
		t.Error("expected error for invalid id")
	}
}

func TestHandler_DeleteORRoom(t *testing.T) {
	h, e := newTestHandler()
	r := &ORRoom{Name: "OR-1"}
	h.svc.CreateORRoom(nil, r)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(r.ID.String())

	err := h.DeleteORRoom(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_CreateSurgicalCase(t *testing.T) {
	h, e := newTestHandler()
	patientID := uuid.New()
	surgeonID := uuid.New()
	body := `{"patient_id":"` + patientID.String() + `","primary_surgeon_id":"` + surgeonID.String() + `","scheduled_date":"2025-06-01T10:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/surgical-cases", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateSurgicalCase(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var sc SurgicalCase
	json.Unmarshal(rec.Body.Bytes(), &sc)
	if sc.Status != "scheduled" {
		t.Errorf("expected status 'scheduled', got %s", sc.Status)
	}
}

func TestHandler_CreateSurgicalCase_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/surgical-cases", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateSurgicalCase(c)
	if err == nil {
		t.Error("expected error for missing required fields")
	}
}

func TestHandler_GetSurgicalCase(t *testing.T) {
	h, e := newTestHandler()
	sc := &SurgicalCase{PatientID: uuid.New(), PrimarySurgeonID: uuid.New(), ScheduledDate: time.Now()}
	h.svc.CreateSurgicalCase(nil, sc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sc.ID.String())

	err := h.GetSurgicalCase(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetSurgicalCase_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetSurgicalCase(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_CreatePreferenceCard(t *testing.T) {
	h, e := newTestHandler()
	surgeonID := uuid.New()
	body := `{"surgeon_id":"` + surgeonID.String() + `","procedure_code":"12345"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/preference-cards", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreatePreferenceCard(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateImplantLog(t *testing.T) {
	h, e := newTestHandler()
	patientID := uuid.New()
	body := `{"patient_id":"` + patientID.String() + `","implant_type":"knee"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/implant-logs", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateImplantLog(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_AddCaseProcedure(t *testing.T) {
	h, e := newTestHandler()
	caseID := uuid.New()
	body := `{"procedure_code":"12345"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(caseID.String())

	err := h.AddCaseProcedure(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}
