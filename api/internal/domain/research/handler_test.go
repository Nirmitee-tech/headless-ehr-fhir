package research

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

// ── Study Handlers ──

func TestHandler_CreateStudy(t *testing.T) {
	h, e := newTestHandler()
	body := `{"protocol_number":"PROTO-001","title":"Test Study"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateStudy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateStudy_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"protocol_number":"PROTO-001"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateStudy(c)
	if err == nil {
		t.Error("expected error for missing title")
	}
}

func TestHandler_GetStudy(t *testing.T) {
	h, e := newTestHandler()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	h.svc.CreateStudy(nil, s)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.ID.String())
	err := h.GetStudy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetStudy_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetStudy(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_DeleteStudy(t *testing.T) {
	h, e := newTestHandler()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	h.svc.CreateStudy(nil, s)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.ID.String())
	err := h.DeleteStudy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddArm(t *testing.T) {
	h, e := newTestHandler()
	studyID := uuid.New()
	body := `{"name":"Treatment A"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(studyID.String())
	err := h.AddArm(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_AddArm_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	studyID := uuid.New()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(studyID.String())
	err := h.AddArm(c)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

// ── Enrollment Handlers ──

func TestHandler_CreateEnrollment(t *testing.T) {
	h, e := newTestHandler()
	body := `{"study_id":"` + uuid.New().String() + `","patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateEnrollment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateEnrollment_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"study_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateEnrollment(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetEnrollment(t *testing.T) {
	h, e := newTestHandler()
	en := &ResearchEnrollment{StudyID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateEnrollment(nil, en)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(en.ID.String())
	err := h.GetEnrollment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteEnrollment(t *testing.T) {
	h, e := newTestHandler()
	en := &ResearchEnrollment{StudyID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateEnrollment(nil, en)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(en.ID.String())
	err := h.DeleteEnrollment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// ── Adverse Event Handlers ──

func TestHandler_CreateAdverseEvent(t *testing.T) {
	h, e := newTestHandler()
	body := `{"enrollment_id":"` + uuid.New().String() + `","description":"Nausea"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateAdverseEvent(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateAdverseEvent_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"enrollment_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateAdverseEvent(c)
	if err == nil {
		t.Error("expected error for missing description")
	}
}

func TestHandler_GetAdverseEvent(t *testing.T) {
	h, e := newTestHandler()
	ae := &ResearchAdverseEvent{EnrollmentID: uuid.New(), Description: "Nausea"}
	h.svc.CreateAdverseEvent(nil, ae)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ae.ID.String())
	err := h.GetAdverseEvent(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteAdverseEvent(t *testing.T) {
	h, e := newTestHandler()
	ae := &ResearchAdverseEvent{EnrollmentID: uuid.New(), Description: "Nausea"}
	h.svc.CreateAdverseEvent(nil, ae)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ae.ID.String())
	err := h.DeleteAdverseEvent(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// ── Deviation Handlers ──

func TestHandler_CreateDeviation(t *testing.T) {
	h, e := newTestHandler()
	body := `{"enrollment_id":"` + uuid.New().String() + `","description":"Wrong dose"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateDeviation(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateDeviation_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"enrollment_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateDeviation(c)
	if err == nil {
		t.Error("expected error for missing description")
	}
}

func TestHandler_GetDeviation(t *testing.T) {
	h, e := newTestHandler()
	d := &ResearchProtocolDeviation{EnrollmentID: uuid.New(), Description: "Wrong dose"}
	h.svc.CreateDeviation(nil, d)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(d.ID.String())
	err := h.GetDeviation(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteDeviation(t *testing.T) {
	h, e := newTestHandler()
	d := &ResearchProtocolDeviation{EnrollmentID: uuid.New(), Description: "Wrong dose"}
	h.svc.CreateDeviation(nil, d)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(d.ID.String())
	err := h.DeleteDeviation(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}
