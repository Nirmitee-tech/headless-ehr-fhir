package research

import (
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

// -- REST: Missing List/Update/GetArms Tests --

func TestHandler_ListStudies(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateStudy(nil, &ResearchStudy{ProtocolNumber: "P-1", Title: "T1"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListStudies(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateStudy(t *testing.T) {
	h, e := newTestHandler()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	h.svc.CreateStudy(nil, s)
	body := `{"title":"Updated","protocol_number":"P-1","status":"active-recruiting"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.ID.String())
	if err := h.UpdateStudy(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetArms(t *testing.T) {
	h, e := newTestHandler()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	h.svc.CreateStudy(nil, s)
	h.svc.AddStudyArm(nil, &ResearchArm{StudyID: s.ID, Name: "Arm A"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.ID.String())
	if err := h.GetArms(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListEnrollments(t *testing.T) {
	h, e := newTestHandler()
	sid := uuid.New()
	h.svc.CreateEnrollment(nil, &ResearchEnrollment{StudyID: sid, PatientID: uuid.New()})
	req := httptest.NewRequest(http.MethodGet, "/?study_id="+sid.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListEnrollments(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateEnrollment(t *testing.T) {
	h, e := newTestHandler()
	en := &ResearchEnrollment{StudyID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateEnrollment(nil, en)
	body := `{"status":"enrolled","study_id":"` + en.StudyID.String() + `","patient_id":"` + en.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(en.ID.String())
	if err := h.UpdateEnrollment(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListAdverseEvents(t *testing.T) {
	h, e := newTestHandler()
	eid := uuid.New()
	h.svc.CreateAdverseEvent(nil, &ResearchAdverseEvent{EnrollmentID: eid, Description: "AE"})
	req := httptest.NewRequest(http.MethodGet, "/?enrollment_id="+eid.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListAdverseEvents(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateAdverseEvent(t *testing.T) {
	h, e := newTestHandler()
	ae := &ResearchAdverseEvent{EnrollmentID: uuid.New(), Description: "AE"}
	h.svc.CreateAdverseEvent(nil, ae)
	body := `{"description":"Updated AE","enrollment_id":"` + ae.EnrollmentID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ae.ID.String())
	if err := h.UpdateAdverseEvent(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListDeviations(t *testing.T) {
	h, e := newTestHandler()
	eid := uuid.New()
	h.svc.CreateDeviation(nil, &ResearchProtocolDeviation{EnrollmentID: eid, Description: "D"})
	req := httptest.NewRequest(http.MethodGet, "/?enrollment_id="+eid.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListDeviations(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateDeviation(t *testing.T) {
	h, e := newTestHandler()
	d := &ResearchProtocolDeviation{EnrollmentID: uuid.New(), Description: "D"}
	h.svc.CreateDeviation(nil, d)
	body := `{"description":"Updated","enrollment_id":"` + d.EnrollmentID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(d.ID.String())
	if err := h.UpdateDeviation(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- FHIR ResearchStudy Handlers --

func TestHandler_SearchStudiesFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateStudy(nil, &ResearchStudy{ProtocolNumber: "P-1", Title: "T"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.SearchStudiesFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetStudyFHIR(t *testing.T) {
	h, e := newTestHandler()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	h.svc.CreateStudy(nil, s)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.FHIRID)
	if err := h.GetStudyFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetStudyFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	if err := h.GetStudyFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateStudyFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"protocol_number":"P-1","title":"T"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.CreateStudyFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	if rec.Header().Get("Location") == "" {
		t.Error("expected Location header")
	}
}

func TestHandler_UpdateStudyFHIR(t *testing.T) {
	h, e := newTestHandler()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	h.svc.CreateStudy(nil, s)
	body := `{"protocol_number":"P-1","title":"Updated","status":"active-recruiting"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.FHIRID)
	if err := h.UpdateStudyFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteStudyFHIR(t *testing.T) {
	h, e := newTestHandler()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	h.svc.CreateStudy(nil, s)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.FHIRID)
	if err := h.DeleteStudyFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchStudyFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	h.svc.CreateStudy(nil, s)
	body := `{"title":"Patched"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.FHIRID)
	if err := h.PatchStudyFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_PatchStudyFHIR_JSONPatch(t *testing.T) {
	h, e := newTestHandler()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	h.svc.CreateStudy(nil, s)
	body := `[{"op":"replace","path":"/title","value":"JP"}]`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.FHIRID)
	if err := h.PatchStudyFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_PatchStudyFHIR_UnsupportedMediaType(t *testing.T) {
	h, e := newTestHandler()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	h.svc.CreateStudy(nil, s)
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.FHIRID)
	if err := h.PatchStudyFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415, got %d", rec.Code)
	}
}

func TestHandler_VreadStudyFHIR(t *testing.T) {
	h, e := newTestHandler()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	h.svc.CreateStudy(nil, s)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(s.FHIRID, "1")
	if err := h.VreadStudyFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryStudyFHIR(t *testing.T) {
	h, e := newTestHandler()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	h.svc.CreateStudy(nil, s)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(s.FHIRID)
	if err := h.HistoryStudyFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
