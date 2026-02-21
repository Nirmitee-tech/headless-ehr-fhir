package diagnostics

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
	h := NewHandler(svc, nil)
	e := echo.New()
	return h, e
}

func TestHandler_CreateServiceRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","requester_id":"` + uuid.New().String() + `","code_value":"CBC","code_display":"Complete Blood Count"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateServiceRequest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateServiceRequest_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"code_value":"CBC"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateServiceRequest(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetServiceRequest(t *testing.T) {
	h, e := newTestHandler()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateServiceRequest(nil, sr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sr.ID.String())

	err := h.GetServiceRequest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetServiceRequest_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetServiceRequest(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_CreateSpecimen(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateSpecimen(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateDiagnosticReport(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","code_value":"CBC","code_display":"Complete Blood Count"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateDiagnosticReport(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateImagingStudy(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateImagingStudy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_AddResult(t *testing.T) {
	h, e := newTestHandler()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateDiagnosticReport(nil, dr)

	obsID := uuid.New()
	body := `{"observation_id":"` + obsID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(dr.ID.String())

	err := h.AddResult(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["diagnostic_report_id"] != dr.ID.String() {
		t.Errorf("unexpected diagnostic_report_id")
	}
}

func TestHandler_DeleteServiceRequest(t *testing.T) {
	h, e := newTestHandler()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateServiceRequest(nil, sr)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sr.ID.String())

	err := h.DeleteServiceRequest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_ListServiceRequests(t *testing.T) {
	h, e := newTestHandler()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateServiceRequest(nil, sr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListServiceRequests(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateServiceRequest(t *testing.T) {
	h, e := newTestHandler()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateServiceRequest(nil, sr)

	body := `{"patient_id":"` + sr.PatientID.String() + `","requester_id":"` + sr.RequesterID.String() + `","code_value":"CMP","code_display":"Comprehensive Metabolic Panel"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sr.ID.String())

	err := h.UpdateServiceRequest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetSpecimen(t *testing.T) {
	h, e := newTestHandler()
	sp := &Specimen{PatientID: uuid.New()}
	h.svc.CreateSpecimen(nil, sp)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sp.ID.String())

	err := h.GetSpecimen(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListSpecimens(t *testing.T) {
	h, e := newTestHandler()
	sp := &Specimen{PatientID: uuid.New()}
	h.svc.CreateSpecimen(nil, sp)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListSpecimens(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateSpecimen(t *testing.T) {
	h, e := newTestHandler()
	sp := &Specimen{PatientID: uuid.New()}
	h.svc.CreateSpecimen(nil, sp)

	body := `{"patient_id":"` + sp.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sp.ID.String())

	err := h.UpdateSpecimen(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteSpecimen(t *testing.T) {
	h, e := newTestHandler()
	sp := &Specimen{PatientID: uuid.New()}
	h.svc.CreateSpecimen(nil, sp)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sp.ID.String())

	err := h.DeleteSpecimen(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_GetDiagnosticReport(t *testing.T) {
	h, e := newTestHandler()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateDiagnosticReport(nil, dr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(dr.ID.String())

	err := h.GetDiagnosticReport(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListDiagnosticReports(t *testing.T) {
	h, e := newTestHandler()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateDiagnosticReport(nil, dr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListDiagnosticReports(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateDiagnosticReport(t *testing.T) {
	h, e := newTestHandler()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateDiagnosticReport(nil, dr)

	body := `{"patient_id":"` + dr.PatientID.String() + `","code_value":"CMP","code_display":"Comprehensive Metabolic Panel"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(dr.ID.String())

	err := h.UpdateDiagnosticReport(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteDiagnosticReport(t *testing.T) {
	h, e := newTestHandler()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateDiagnosticReport(nil, dr)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(dr.ID.String())

	err := h.DeleteDiagnosticReport(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_GetResults(t *testing.T) {
	h, e := newTestHandler()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateDiagnosticReport(nil, dr)
	obsID := uuid.New()
	h.svc.AddDiagnosticReportResult(nil, dr.ID, obsID)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(dr.ID.String())

	err := h.GetResults(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_RemoveResult(t *testing.T) {
	h, e := newTestHandler()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateDiagnosticReport(nil, dr)
	obsID := uuid.New()
	h.svc.AddDiagnosticReportResult(nil, dr.ID, obsID)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "observationId")
	c.SetParamValues(dr.ID.String(), obsID.String())

	err := h.RemoveResult(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_GetImagingStudy(t *testing.T) {
	h, e := newTestHandler()
	is := &ImagingStudy{PatientID: uuid.New()}
	h.svc.CreateImagingStudy(nil, is)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(is.ID.String())

	err := h.GetImagingStudy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListImagingStudies(t *testing.T) {
	h, e := newTestHandler()
	is := &ImagingStudy{PatientID: uuid.New()}
	h.svc.CreateImagingStudy(nil, is)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListImagingStudies(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateImagingStudy(t *testing.T) {
	h, e := newTestHandler()
	is := &ImagingStudy{PatientID: uuid.New()}
	h.svc.CreateImagingStudy(nil, is)

	body := `{"patient_id":"` + is.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(is.ID.String())

	err := h.UpdateImagingStudy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteImagingStudy(t *testing.T) {
	h, e := newTestHandler()
	is := &ImagingStudy{PatientID: uuid.New()}
	h.svc.CreateImagingStudy(nil, is)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(is.ID.String())

	err := h.DeleteImagingStudy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// -- FHIR ServiceRequest Tests --

func TestHandler_SearchServiceRequestsFHIR(t *testing.T) {
	h, e := newTestHandler()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateServiceRequest(nil, sr)

	req := httptest.NewRequest(http.MethodGet, "/fhir/ServiceRequest", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchServiceRequestsFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle resourceType")
	}
}

func TestHandler_GetServiceRequestFHIR(t *testing.T) {
	h, e := newTestHandler()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateServiceRequest(nil, sr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sr.FHIRID)

	err := h.GetServiceRequestFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetServiceRequestFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.GetServiceRequestFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateServiceRequestFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","requester_id":"` + uuid.New().String() + `","code_value":"CBC","code_display":"Complete Blood Count"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateServiceRequestFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if !strings.Contains(loc, "/fhir/ServiceRequest/") {
		t.Errorf("expected Location header, got %s", loc)
	}
}

func TestHandler_UpdateServiceRequestFHIR(t *testing.T) {
	h, e := newTestHandler()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateServiceRequest(nil, sr)

	body := `{"patient_id":"` + sr.PatientID.String() + `","requester_id":"` + sr.RequesterID.String() + `","code_value":"CMP"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sr.FHIRID)

	err := h.UpdateServiceRequestFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteServiceRequestFHIR(t *testing.T) {
	h, e := newTestHandler()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateServiceRequest(nil, sr)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sr.FHIRID)

	err := h.DeleteServiceRequestFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchServiceRequestFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateServiceRequest(nil, sr)

	body := `{"status":"completed"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sr.FHIRID)

	err := h.PatchServiceRequestFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadServiceRequestFHIR(t *testing.T) {
	h, e := newTestHandler()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateServiceRequest(nil, sr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(sr.FHIRID, "1")

	err := h.VreadServiceRequestFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryServiceRequestFHIR(t *testing.T) {
	h, e := newTestHandler()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateServiceRequest(nil, sr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sr.FHIRID)

	err := h.HistoryServiceRequestFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- FHIR DiagnosticReport Tests --

func TestHandler_SearchDiagnosticReportsFHIR(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/fhir/DiagnosticReport", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchDiagnosticReportsFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetDiagnosticReportFHIR(t *testing.T) {
	h, e := newTestHandler()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateDiagnosticReport(nil, dr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(dr.FHIRID)

	err := h.GetDiagnosticReportFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_CreateDiagnosticReportFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","code_value":"CBC","code_display":"Complete Blood Count"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateDiagnosticReportFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if !strings.Contains(loc, "/fhir/DiagnosticReport/") {
		t.Errorf("expected Location header, got %s", loc)
	}
}

func TestHandler_UpdateDiagnosticReportFHIR(t *testing.T) {
	h, e := newTestHandler()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateDiagnosticReport(nil, dr)

	body := `{"patient_id":"` + dr.PatientID.String() + `","code_value":"CMP"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(dr.FHIRID)

	err := h.UpdateDiagnosticReportFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteDiagnosticReportFHIR(t *testing.T) {
	h, e := newTestHandler()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateDiagnosticReport(nil, dr)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(dr.FHIRID)

	err := h.DeleteDiagnosticReportFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchDiagnosticReportFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateDiagnosticReport(nil, dr)

	body := `{"status":"final"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(dr.FHIRID)

	err := h.PatchDiagnosticReportFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadDiagnosticReportFHIR(t *testing.T) {
	h, e := newTestHandler()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateDiagnosticReport(nil, dr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(dr.FHIRID, "1")

	err := h.VreadDiagnosticReportFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryDiagnosticReportFHIR(t *testing.T) {
	h, e := newTestHandler()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	h.svc.CreateDiagnosticReport(nil, dr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(dr.FHIRID)

	err := h.HistoryDiagnosticReportFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- FHIR Specimen Tests --

func TestHandler_SearchSpecimensFHIR(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Specimen", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchSpecimensFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetSpecimenFHIR(t *testing.T) {
	h, e := newTestHandler()
	sp := &Specimen{PatientID: uuid.New()}
	h.svc.CreateSpecimen(nil, sp)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sp.FHIRID)

	err := h.GetSpecimenFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_CreateSpecimenFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateSpecimenFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if !strings.Contains(loc, "/fhir/Specimen/") {
		t.Errorf("expected Location header, got %s", loc)
	}
}

func TestHandler_UpdateSpecimenFHIR(t *testing.T) {
	h, e := newTestHandler()
	sp := &Specimen{PatientID: uuid.New()}
	h.svc.CreateSpecimen(nil, sp)

	body := `{"patient_id":"` + sp.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sp.FHIRID)

	err := h.UpdateSpecimenFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteSpecimenFHIR(t *testing.T) {
	h, e := newTestHandler()
	sp := &Specimen{PatientID: uuid.New()}
	h.svc.CreateSpecimen(nil, sp)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sp.FHIRID)

	err := h.DeleteSpecimenFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchSpecimenFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	sp := &Specimen{PatientID: uuid.New()}
	h.svc.CreateSpecimen(nil, sp)

	body := `{"status":"unavailable"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sp.FHIRID)

	err := h.PatchSpecimenFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadSpecimenFHIR(t *testing.T) {
	h, e := newTestHandler()
	sp := &Specimen{PatientID: uuid.New()}
	h.svc.CreateSpecimen(nil, sp)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(sp.FHIRID, "1")

	err := h.VreadSpecimenFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistorySpecimenFHIR(t *testing.T) {
	h, e := newTestHandler()
	sp := &Specimen{PatientID: uuid.New()}
	h.svc.CreateSpecimen(nil, sp)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sp.FHIRID)

	err := h.HistorySpecimenFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- FHIR ImagingStudy Tests --

func TestHandler_SearchImagingStudiesFHIR(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/fhir/ImagingStudy", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchImagingStudiesFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetImagingStudyFHIR(t *testing.T) {
	h, e := newTestHandler()
	is := &ImagingStudy{PatientID: uuid.New()}
	h.svc.CreateImagingStudy(nil, is)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(is.FHIRID)

	err := h.GetImagingStudyFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_CreateImagingStudyFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateImagingStudyFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if !strings.Contains(loc, "/fhir/ImagingStudy/") {
		t.Errorf("expected Location header, got %s", loc)
	}
}

func TestHandler_UpdateImagingStudyFHIR(t *testing.T) {
	h, e := newTestHandler()
	is := &ImagingStudy{PatientID: uuid.New()}
	h.svc.CreateImagingStudy(nil, is)

	body := `{"patient_id":"` + is.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(is.FHIRID)

	err := h.UpdateImagingStudyFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteImagingStudyFHIR(t *testing.T) {
	h, e := newTestHandler()
	is := &ImagingStudy{PatientID: uuid.New()}
	h.svc.CreateImagingStudy(nil, is)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(is.FHIRID)

	err := h.DeleteImagingStudyFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchImagingStudyFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	is := &ImagingStudy{PatientID: uuid.New()}
	h.svc.CreateImagingStudy(nil, is)

	body := `{"status":"available"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(is.FHIRID)

	err := h.PatchImagingStudyFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadImagingStudyFHIR(t *testing.T) {
	h, e := newTestHandler()
	is := &ImagingStudy{PatientID: uuid.New()}
	h.svc.CreateImagingStudy(nil, is)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(is.FHIRID, "1")

	err := h.VreadImagingStudyFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryImagingStudyFHIR(t *testing.T) {
	h, e := newTestHandler()
	is := &ImagingStudy{PatientID: uuid.New()}
	h.svc.CreateImagingStudy(nil, is)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(is.FHIRID)

	err := h.HistoryImagingStudyFHIR(c)
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
	if len(routes) == 0 {
		t.Error("expected routes to be registered")
	}
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}
	expected := []string{
		"POST:/api/v1/service-requests",
		"GET:/api/v1/specimens/:id",
		"POST:/api/v1/diagnostic-reports",
		"GET:/fhir/ServiceRequest",
		"GET:/fhir/DiagnosticReport",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}

// -- applyServiceRequestPatch unit tests --

func TestApplyServiceRequestPatch_Status(t *testing.T) {
	sr := &ServiceRequest{
		Status:      "draft",
		Intent:      "order",
		CodeValue:   "CBC",
		CodeDisplay: "Complete Blood Count",
		PatientID:   uuid.New(),
		RequesterID: uuid.New(),
	}

	patched := map[string]interface{}{
		"status":   "active",
		"intent":   "plan",
		"priority": "urgent",
	}
	applyServiceRequestPatch(sr, patched)

	if sr.Status != "active" {
		t.Errorf("expected status 'active', got %q", sr.Status)
	}
	if sr.Intent != "plan" {
		t.Errorf("expected intent 'plan', got %q", sr.Intent)
	}
	if sr.Priority == nil || *sr.Priority != "urgent" {
		t.Errorf("expected priority 'urgent', got %v", sr.Priority)
	}
	if sr.CodeValue != "CBC" {
		t.Errorf("expected CodeValue unchanged 'CBC', got %q", sr.CodeValue)
	}
}

func TestApplyServiceRequestPatch_Code(t *testing.T) {
	sr := &ServiceRequest{
		Status:      "draft",
		Intent:      "order",
		CodeValue:   "CBC",
		CodeDisplay: "Complete Blood Count",
		PatientID:   uuid.New(),
		RequesterID: uuid.New(),
	}

	patched := map[string]interface{}{
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"code":    "CMP",
					"display": "Comprehensive Metabolic Panel",
					"system":  "http://loinc.org",
				},
			},
		},
		"bodySite": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code":    "368209003",
						"display": "Right arm",
					},
				},
			},
		},
		"note": []interface{}{
			map[string]interface{}{
				"text": "Fasting required",
			},
		},
		"patientInstruction": "Do not eat for 12 hours before test",
	}
	applyServiceRequestPatch(sr, patched)

	if sr.CodeValue != "CMP" {
		t.Errorf("expected CodeValue 'CMP', got %q", sr.CodeValue)
	}
	if sr.CodeDisplay != "Comprehensive Metabolic Panel" {
		t.Errorf("expected CodeDisplay 'Comprehensive Metabolic Panel', got %q", sr.CodeDisplay)
	}
	if sr.CodeSystem == nil || *sr.CodeSystem != "http://loinc.org" {
		t.Errorf("expected CodeSystem 'http://loinc.org', got %v", sr.CodeSystem)
	}
	if sr.BodySiteCode == nil || *sr.BodySiteCode != "368209003" {
		t.Errorf("expected BodySiteCode '368209003', got %v", sr.BodySiteCode)
	}
	if sr.BodySiteDisplay == nil || *sr.BodySiteDisplay != "Right arm" {
		t.Errorf("expected BodySiteDisplay 'Right arm', got %v", sr.BodySiteDisplay)
	}
	if sr.Note == nil || *sr.Note != "Fasting required" {
		t.Errorf("expected Note 'Fasting required', got %v", sr.Note)
	}
	if sr.PatientInstruction == nil || *sr.PatientInstruction != "Do not eat for 12 hours before test" {
		t.Errorf("expected PatientInstruction set, got %v", sr.PatientInstruction)
	}
	if sr.Status != "draft" {
		t.Errorf("expected Status unchanged 'draft', got %q", sr.Status)
	}
}

// -- applyDiagnosticReportPatch unit tests --

func TestApplyDiagnosticReportPatch_StatusConclusion(t *testing.T) {
	dr := &DiagnosticReport{
		Status:      "preliminary",
		CodeValue:   "CBC",
		CodeDisplay: "Complete Blood Count",
		PatientID:   uuid.New(),
	}

	patched := map[string]interface{}{
		"status":     "final",
		"conclusion": "All values within normal range",
		"issued":     "2025-07-15T08:00:00Z",
		"conclusionCode": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code":    "N",
						"display": "Normal",
					},
				},
			},
		},
		"presentedForm": []interface{}{
			map[string]interface{}{
				"url":         "https://example.com/report.pdf",
				"contentType": "application/pdf",
			},
		},
	}
	applyDiagnosticReportPatch(dr, patched)

	if dr.Status != "final" {
		t.Errorf("expected status 'final', got %q", dr.Status)
	}
	if dr.Conclusion == nil || *dr.Conclusion != "All values within normal range" {
		t.Errorf("expected Conclusion set, got %v", dr.Conclusion)
	}
	if dr.Issued == nil {
		t.Fatal("expected Issued to be set")
	}
	if dr.ConclusionCode == nil || *dr.ConclusionCode != "N" {
		t.Errorf("expected ConclusionCode 'N', got %v", dr.ConclusionCode)
	}
	if dr.ConclusionDisplay == nil || *dr.ConclusionDisplay != "Normal" {
		t.Errorf("expected ConclusionDisplay 'Normal', got %v", dr.ConclusionDisplay)
	}
	if dr.PresentedFormURL == nil || *dr.PresentedFormURL != "https://example.com/report.pdf" {
		t.Errorf("expected PresentedFormURL set, got %v", dr.PresentedFormURL)
	}
	if dr.PresentedFormType == nil || *dr.PresentedFormType != "application/pdf" {
		t.Errorf("expected PresentedFormType 'application/pdf', got %v", dr.PresentedFormType)
	}
	if dr.CodeValue != "CBC" {
		t.Errorf("expected CodeValue unchanged 'CBC', got %q", dr.CodeValue)
	}
}

// -- applySpecimenPatch unit tests --

func TestApplySpecimenPatch_Status(t *testing.T) {
	sp := &Specimen{
		Status:    "available",
		PatientID: uuid.New(),
	}

	patched := map[string]interface{}{
		"status": "unavailable",
		"type": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"code":    "122555007",
					"display": "Venous blood specimen",
				},
			},
		},
		"accessionIdentifier": map[string]interface{}{
			"value": "ACC-12345",
		},
		"receivedTime": "2025-08-01T09:00:00Z",
		"collection": map[string]interface{}{
			"collectedDateTime": "2025-08-01T08:30:00Z",
			"quantity": map[string]interface{}{
				"value": float64(10),
				"unit":  "mL",
			},
			"method": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code": "venipuncture",
					},
				},
			},
			"bodySite": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code": "368209003",
					},
				},
			},
		},
		"condition": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code":    "ROOM",
						"display": "Room temperature",
					},
				},
			},
		},
		"note": []interface{}{
			map[string]interface{}{
				"text": "Handle with care",
			},
		},
	}
	applySpecimenPatch(sp, patched)

	if sp.Status != "unavailable" {
		t.Errorf("expected status 'unavailable', got %q", sp.Status)
	}
	if sp.TypeCode == nil || *sp.TypeCode != "122555007" {
		t.Errorf("expected TypeCode '122555007', got %v", sp.TypeCode)
	}
	if sp.TypeDisplay == nil || *sp.TypeDisplay != "Venous blood specimen" {
		t.Errorf("expected TypeDisplay 'Venous blood specimen', got %v", sp.TypeDisplay)
	}
	if sp.AccessionID == nil || *sp.AccessionID != "ACC-12345" {
		t.Errorf("expected AccessionID 'ACC-12345', got %v", sp.AccessionID)
	}
	if sp.ReceivedTime == nil {
		t.Fatal("expected ReceivedTime to be set")
	}
	if sp.CollectionDatetime == nil {
		t.Fatal("expected CollectionDatetime to be set")
	}
	if sp.CollectionQuantity == nil || *sp.CollectionQuantity != 10 {
		t.Errorf("expected CollectionQuantity 10, got %v", sp.CollectionQuantity)
	}
	if sp.CollectionUnit == nil || *sp.CollectionUnit != "mL" {
		t.Errorf("expected CollectionUnit 'mL', got %v", sp.CollectionUnit)
	}
	if sp.CollectionMethod == nil || *sp.CollectionMethod != "venipuncture" {
		t.Errorf("expected CollectionMethod 'venipuncture', got %v", sp.CollectionMethod)
	}
	if sp.CollectionBodySite == nil || *sp.CollectionBodySite != "368209003" {
		t.Errorf("expected CollectionBodySite '368209003', got %v", sp.CollectionBodySite)
	}
	if sp.ConditionCode == nil || *sp.ConditionCode != "ROOM" {
		t.Errorf("expected ConditionCode 'ROOM', got %v", sp.ConditionCode)
	}
	if sp.Note == nil || *sp.Note != "Handle with care" {
		t.Errorf("expected Note 'Handle with care', got %v", sp.Note)
	}
}

// -- applyImagingStudyPatch unit tests --

func TestApplyImagingStudyPatch_StatusModality(t *testing.T) {
	is := &ImagingStudy{
		Status:    "registered",
		PatientID: uuid.New(),
	}

	patched := map[string]interface{}{
		"status": "available",
		"modality": []interface{}{
			map[string]interface{}{
				"code":    "CT",
				"display": "Computed Tomography",
			},
		},
		"description":       "CT scan of abdomen",
		"started":           "2025-09-01T10:00:00Z",
		"numberOfSeries":    float64(3),
		"numberOfInstances": float64(120),
		"reasonCode": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code":    "R10.9",
						"display": "Abdominal pain",
					},
				},
			},
		},
		"endpoint": []interface{}{
			map[string]interface{}{
				"reference": "Endpoint/pacs-1",
			},
		},
		"identifier": []interface{}{
			map[string]interface{}{
				"value": "1.2.840.113619.2.55.3.604688119",
			},
		},
		"note": []interface{}{
			map[string]interface{}{
				"text": "Contrast administered",
			},
		},
	}
	applyImagingStudyPatch(is, patched)

	if is.Status != "available" {
		t.Errorf("expected status 'available', got %q", is.Status)
	}
	if is.ModalityCode == nil || *is.ModalityCode != "CT" {
		t.Errorf("expected ModalityCode 'CT', got %v", is.ModalityCode)
	}
	if is.ModalityDisplay == nil || *is.ModalityDisplay != "Computed Tomography" {
		t.Errorf("expected ModalityDisplay 'Computed Tomography', got %v", is.ModalityDisplay)
	}
	if is.Description == nil || *is.Description != "CT scan of abdomen" {
		t.Errorf("expected Description 'CT scan of abdomen', got %v", is.Description)
	}
	if is.Started == nil {
		t.Fatal("expected Started to be set")
	}
	if is.NumberOfSeries == nil || *is.NumberOfSeries != 3 {
		t.Errorf("expected NumberOfSeries 3, got %v", is.NumberOfSeries)
	}
	if is.NumberOfInstances == nil || *is.NumberOfInstances != 120 {
		t.Errorf("expected NumberOfInstances 120, got %v", is.NumberOfInstances)
	}
	if is.ReasonCode == nil || *is.ReasonCode != "R10.9" {
		t.Errorf("expected ReasonCode 'R10.9', got %v", is.ReasonCode)
	}
	if is.ReasonDisplay == nil || *is.ReasonDisplay != "Abdominal pain" {
		t.Errorf("expected ReasonDisplay 'Abdominal pain', got %v", is.ReasonDisplay)
	}
	if is.Endpoint == nil || *is.Endpoint != "Endpoint/pacs-1" {
		t.Errorf("expected Endpoint 'Endpoint/pacs-1', got %v", is.Endpoint)
	}
	if is.StudyUID == nil || *is.StudyUID != "1.2.840.113619.2.55.3.604688119" {
		t.Errorf("expected StudyUID set, got %v", is.StudyUID)
	}
	if is.Note == nil || *is.Note != "Contrast administered" {
		t.Errorf("expected Note 'Contrast administered', got %v", is.Note)
	}
}
