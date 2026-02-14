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
		"GET:/api/v1/coverages/:id",
		"POST:/api/v1/claims",
		"GET:/api/v1/claims/:id",
		"POST:/api/v1/claim-responses",
		"POST:/api/v1/invoices",
		"GET:/fhir/Coverage",
		"GET:/fhir/Claim",
		"GET:/fhir/ClaimResponse",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
