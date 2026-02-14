package billing

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

// -- Coverage Handler Tests --

func TestHandler_CreateCoverage(t *testing.T) {
	h, e := newTestHandler()
	payorOrgID := uuid.New()
	body := `{"patient_id":"` + uuid.New().String() + `","payor_org_id":"` + payorOrgID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateCoverage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateCoverage_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateCoverage(c)
	if err == nil {
		t.Error("expected error for missing payor")
	}
}

func TestHandler_GetCoverage(t *testing.T) {
	h, e := newTestHandler()
	payorName := "Aetna"
	cov := &Coverage{PatientID: uuid.New(), PayorName: &payorName}
	h.svc.CreateCoverage(nil, cov)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cov.ID.String())

	err := h.GetCoverage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetCoverage_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetCoverage(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_DeleteCoverage(t *testing.T) {
	h, e := newTestHandler()
	payorName := "Aetna"
	cov := &Coverage{PatientID: uuid.New(), PayorName: &payorName}
	h.svc.CreateCoverage(nil, cov)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cov.ID.String())

	err := h.DeleteCoverage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// -- Claim Handler Tests --

func TestHandler_CreateClaim(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateClaim(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateClaim_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateClaim(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetClaim(t *testing.T) {
	h, e := newTestHandler()
	cl := &Claim{PatientID: uuid.New()}
	h.svc.CreateClaim(nil, cl)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cl.ID.String())

	err := h.GetClaim(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteClaim(t *testing.T) {
	h, e := newTestHandler()
	cl := &Claim{PatientID: uuid.New()}
	h.svc.CreateClaim(nil, cl)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cl.ID.String())

	err := h.DeleteClaim(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddClaimDiagnosis(t *testing.T) {
	h, e := newTestHandler()
	cl := &Claim{PatientID: uuid.New()}
	h.svc.CreateClaim(nil, cl)

	body := `{"diagnosis_code":"J06.9","diagnosis_display":"URI"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cl.ID.String())

	err := h.AddClaimDiagnosis(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var result ClaimDiagnosis
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result.ClaimID != cl.ID {
		t.Error("expected claim_id to match")
	}
}

func TestHandler_AddClaimProcedure(t *testing.T) {
	h, e := newTestHandler()
	cl := &Claim{PatientID: uuid.New()}
	h.svc.CreateClaim(nil, cl)

	body := `{"procedure_code":"99213"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cl.ID.String())

	err := h.AddClaimProcedure(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_AddClaimItem(t *testing.T) {
	h, e := newTestHandler()
	cl := &Claim{PatientID: uuid.New()}
	h.svc.CreateClaim(nil, cl)

	body := `{"product_or_service_code":"99213"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cl.ID.String())

	err := h.AddClaimItem(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

// -- ClaimResponse Handler Tests --

func TestHandler_CreateClaimResponse(t *testing.T) {
	h, e := newTestHandler()
	body := `{"claim_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateClaimResponse(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetClaimResponse(t *testing.T) {
	h, e := newTestHandler()
	cr := &ClaimResponse{ClaimID: uuid.New()}
	h.svc.CreateClaimResponse(nil, cr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cr.ID.String())

	err := h.GetClaimResponse(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- Invoice Handler Tests --

func TestHandler_CreateInvoice(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateInvoice(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateInvoice_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateInvoice(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetInvoice(t *testing.T) {
	h, e := newTestHandler()
	inv := &Invoice{PatientID: uuid.New()}
	h.svc.CreateInvoice(nil, inv)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(inv.ID.String())

	err := h.GetInvoice(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteInvoice(t *testing.T) {
	h, e := newTestHandler()
	inv := &Invoice{PatientID: uuid.New()}
	h.svc.CreateInvoice(nil, inv)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(inv.ID.String())

	err := h.DeleteInvoice(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddInvoiceLineItem(t *testing.T) {
	h, e := newTestHandler()
	inv := &Invoice{PatientID: uuid.New()}
	h.svc.CreateInvoice(nil, inv)

	body := `{"sequence":1,"description":"Office Visit"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(inv.ID.String())

	err := h.AddInvoiceLineItem(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

// -- List/Update Tests --

func TestHandler_ListCoverages(t *testing.T) {
	h, e := newTestHandler()
	payorName := "Aetna"
	cov := &Coverage{PatientID: uuid.New(), PayorName: &payorName}
	h.svc.CreateCoverage(nil, cov)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListCoverages(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateCoverage(t *testing.T) {
	h, e := newTestHandler()
	payorName := "Aetna"
	cov := &Coverage{PatientID: uuid.New(), PayorName: &payorName}
	h.svc.CreateCoverage(nil, cov)
	body := `{"patient_id":"` + cov.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cov.ID.String())
	err := h.UpdateCoverage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListClaims(t *testing.T) {
	h, e := newTestHandler()
	cl := &Claim{PatientID: uuid.New()}
	h.svc.CreateClaim(nil, cl)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListClaims(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateClaim(t *testing.T) {
	h, e := newTestHandler()
	cl := &Claim{PatientID: uuid.New()}
	h.svc.CreateClaim(nil, cl)
	body := `{"patient_id":"` + cl.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cl.ID.String())
	err := h.UpdateClaim(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListClaimResponses(t *testing.T) {
	h, e := newTestHandler()
	cr := &ClaimResponse{ClaimID: uuid.New()}
	h.svc.CreateClaimResponse(nil, cr)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListClaimResponses(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListInvoices(t *testing.T) {
	h, e := newTestHandler()
	inv := &Invoice{PatientID: uuid.New()}
	h.svc.CreateInvoice(nil, inv)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListInvoices(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateInvoice(t *testing.T) {
	h, e := newTestHandler()
	inv := &Invoice{PatientID: uuid.New()}
	h.svc.CreateInvoice(nil, inv)
	body := `{"patient_id":"` + inv.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(inv.ID.String())
	err := h.UpdateInvoice(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetClaimDiagnoses(t *testing.T) {
	h, e := newTestHandler()
	cl := &Claim{PatientID: uuid.New()}
	h.svc.CreateClaim(nil, cl)
	diag := &ClaimDiagnosis{ClaimID: cl.ID, DiagnosisCode: "J06.9"}
	h.svc.AddClaimDiagnosis(nil, diag)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cl.ID.String())
	err := h.GetClaimDiagnoses(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetClaimProcedures(t *testing.T) {
	h, e := newTestHandler()
	cl := &Claim{PatientID: uuid.New()}
	h.svc.CreateClaim(nil, cl)
	proc := &ClaimProcedure{ClaimID: cl.ID, ProcedureCode: "99213"}
	h.svc.AddClaimProcedure(nil, proc)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cl.ID.String())
	err := h.GetClaimProcedures(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetClaimItems(t *testing.T) {
	h, e := newTestHandler()
	cl := &Claim{PatientID: uuid.New()}
	h.svc.CreateClaim(nil, cl)
	item := &ClaimItem{ClaimID: cl.ID, ProductOrServiceCode: "99213"}
	h.svc.AddClaimItem(nil, item)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cl.ID.String())
	err := h.GetClaimItems(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetInvoiceLineItems(t *testing.T) {
	h, e := newTestHandler()
	inv := &Invoice{PatientID: uuid.New()}
	h.svc.CreateInvoice(nil, inv)
	li := &InvoiceLineItem{InvoiceID: inv.ID, Sequence: 1}
	h.svc.AddInvoiceLineItem(nil, li)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(inv.ID.String())
	err := h.GetInvoiceLineItems(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- FHIR Endpoint Tests --

func TestHandler_SearchCoveragesFHIR(t *testing.T) {
	h, e := newTestHandler()
	payorName := "Aetna"
	cov := &Coverage{PatientID: uuid.New(), PayorName: &payorName}
	h.svc.CreateCoverage(nil, cov)
	req := httptest.NewRequest(http.MethodGet, "/fhir/Coverage", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchCoveragesFHIR(c)
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

func TestHandler_GetCoverageFHIR(t *testing.T) {
	h, e := newTestHandler()
	payorName := "Aetna"
	cov := &Coverage{PatientID: uuid.New(), PayorName: &payorName}
	h.svc.CreateCoverage(nil, cov)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cov.FHIRID)
	err := h.GetCoverageFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetCoverageFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	_ = h.GetCoverageFHIR(c)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateCoverageFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","payor_name":"Aetna"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateCoverageFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc == "" || !strings.Contains(loc, "/fhir/Coverage/") {
		t.Errorf("expected Location header, got %q", loc)
	}
}

func TestHandler_UpdateCoverageFHIR(t *testing.T) {
	h, e := newTestHandler()
	payorName := "Aetna"
	cov := &Coverage{PatientID: uuid.New(), PayorName: &payorName}
	h.svc.CreateCoverage(nil, cov)
	body := `{"patient_id":"` + cov.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cov.FHIRID)
	err := h.UpdateCoverageFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteCoverageFHIR(t *testing.T) {
	h, e := newTestHandler()
	payorName := "Aetna"
	cov := &Coverage{PatientID: uuid.New(), PayorName: &payorName}
	h.svc.CreateCoverage(nil, cov)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cov.FHIRID)
	err := h.DeleteCoverageFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchCoverageFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	payorName := "Aetna"
	cov := &Coverage{PatientID: uuid.New(), PayorName: &payorName}
	h.svc.CreateCoverage(nil, cov)
	body := `{"status":"cancelled"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cov.FHIRID)
	err := h.PatchCoverageFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadCoverageFHIR(t *testing.T) {
	h, e := newTestHandler()
	payorName := "Aetna"
	cov := &Coverage{PatientID: uuid.New(), PayorName: &payorName}
	h.svc.CreateCoverage(nil, cov)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(cov.FHIRID, "1")
	err := h.VreadCoverageFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryCoverageFHIR(t *testing.T) {
	h, e := newTestHandler()
	payorName := "Aetna"
	cov := &Coverage{PatientID: uuid.New(), PayorName: &payorName}
	h.svc.CreateCoverage(nil, cov)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cov.FHIRID)
	err := h.HistoryCoverageFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_SearchClaimsFHIR(t *testing.T) {
	h, e := newTestHandler()
	cl := &Claim{PatientID: uuid.New()}
	h.svc.CreateClaim(nil, cl)
	req := httptest.NewRequest(http.MethodGet, "/fhir/Claim", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchClaimsFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetClaimFHIR(t *testing.T) {
	h, e := newTestHandler()
	cl := &Claim{PatientID: uuid.New()}
	h.svc.CreateClaim(nil, cl)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cl.FHIRID)
	err := h.GetClaimFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_CreateClaimFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateClaimFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc == "" || !strings.Contains(loc, "/fhir/Claim/") {
		t.Errorf("expected Location header, got %q", loc)
	}
}

func TestHandler_UpdateClaimFHIR(t *testing.T) {
	h, e := newTestHandler()
	cl := &Claim{PatientID: uuid.New()}
	h.svc.CreateClaim(nil, cl)
	body := `{"patient_id":"` + cl.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cl.FHIRID)
	err := h.UpdateClaimFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteClaimFHIR(t *testing.T) {
	h, e := newTestHandler()
	cl := &Claim{PatientID: uuid.New()}
	h.svc.CreateClaim(nil, cl)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cl.FHIRID)
	err := h.DeleteClaimFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchClaimFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	cl := &Claim{PatientID: uuid.New()}
	h.svc.CreateClaim(nil, cl)
	body := `{"status":"cancelled"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cl.FHIRID)
	err := h.PatchClaimFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadClaimFHIR(t *testing.T) {
	h, e := newTestHandler()
	cl := &Claim{PatientID: uuid.New()}
	h.svc.CreateClaim(nil, cl)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(cl.FHIRID, "1")
	err := h.VreadClaimFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryClaimFHIR(t *testing.T) {
	h, e := newTestHandler()
	cl := &Claim{PatientID: uuid.New()}
	h.svc.CreateClaim(nil, cl)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cl.FHIRID)
	err := h.HistoryClaimFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_SearchClaimResponsesFHIR(t *testing.T) {
	h, e := newTestHandler()
	cr := &ClaimResponse{ClaimID: uuid.New()}
	h.svc.CreateClaimResponse(nil, cr)
	req := httptest.NewRequest(http.MethodGet, "/fhir/ClaimResponse", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchClaimResponsesFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetClaimResponseFHIR(t *testing.T) {
	h, e := newTestHandler()
	cr := &ClaimResponse{ClaimID: uuid.New()}
	h.svc.CreateClaimResponse(nil, cr)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cr.FHIRID)
	err := h.GetClaimResponseFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_CreateClaimResponseFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"claim_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateClaimResponseFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_SearchEOBsFHIR(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/fhir/ExplanationOfBenefit", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchEOBsFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetEOBFHIR(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	_ = h.GetEOBFHIR(c)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
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
		"POST:/api/v1/coverages",
		"GET:/api/v1/coverages",
		"GET:/api/v1/coverages/:id",
		"PUT:/api/v1/coverages/:id",
		"DELETE:/api/v1/coverages/:id",
		"POST:/api/v1/claims",
		"GET:/api/v1/claims",
		"GET:/api/v1/claims/:id",
		"PUT:/api/v1/claims/:id",
		"DELETE:/api/v1/claims/:id",
		"POST:/api/v1/claim-responses",
		"GET:/api/v1/claim-responses",
		"POST:/api/v1/invoices",
		"GET:/api/v1/invoices",
		"GET:/api/v1/invoices/:id",
		"GET:/fhir/Coverage",
		"GET:/fhir/Coverage/:id",
		"POST:/fhir/Coverage",
		"PUT:/fhir/Coverage/:id",
		"DELETE:/fhir/Coverage/:id",
		"PATCH:/fhir/Coverage/:id",
		"GET:/fhir/Claim",
		"GET:/fhir/Claim/:id",
		"POST:/fhir/Claim",
		"GET:/fhir/ClaimResponse",
		"GET:/fhir/ClaimResponse/:id",
		"GET:/fhir/ExplanationOfBenefit",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
