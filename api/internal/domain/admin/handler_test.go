package admin

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

func TestHandler_CreateOrganization(t *testing.T) {
	h, e := newTestHandler()

	body := `{"name":"Test Hospital","type_code":"prov"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/organizations", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateOrganization(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var org Organization
	json.Unmarshal(rec.Body.Bytes(), &org)
	if org.Name != "Test Hospital" {
		t.Errorf("expected 'Test Hospital', got %s", org.Name)
	}
}

func TestHandler_CreateOrganization_BadRequest(t *testing.T) {
	h, e := newTestHandler()

	body := `{"type_code":"prov"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/organizations", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateOrganization(c)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestHandler_GetOrganization(t *testing.T) {
	h, e := newTestHandler()

	// Create first
	org := &Organization{Name: "Test"}
	h.svc.CreateOrganization(nil, org)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(org.ID.String())

	err := h.GetOrganization(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetOrganization_NotFound(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetOrganization(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_GetOrganization_InvalidID(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.GetOrganization(c)
	if err == nil {
		t.Error("expected error for invalid id")
	}
}

func TestHandler_DeleteOrganization(t *testing.T) {
	h, e := newTestHandler()

	org := &Organization{Name: "ToDelete"}
	h.svc.CreateOrganization(nil, org)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(org.ID.String())

	err := h.DeleteOrganization(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_ListOrganizations(t *testing.T) {
	h, e := newTestHandler()

	h.svc.CreateOrganization(nil, &Organization{Name: "Org1"})
	h.svc.CreateOrganization(nil, &Organization{Name: "Org2"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListOrganizations(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_CreateSystemUser(t *testing.T) {
	h, e := newTestHandler()

	body := `{"username":"jdoe","user_type":"provider"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateSystemUser(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateLocation(t *testing.T) {
	h, e := newTestHandler()

	body := `{"name":"Main Building"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/locations", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateLocation(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateDepartment(t *testing.T) {
	h, e := newTestHandler()
	orgID := uuid.New()

	body := `{"name":"Cardiology","organization_id":"` + orgID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/departments", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateDepartment(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
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

	// Check some key routes exist
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}

	expected := []string{
		"POST:/api/v1/organizations",
		"GET:/api/v1/organizations",
		"GET:/api/v1/organizations/:id",
		"POST:/api/v1/users",
		"GET:/fhir/Organization",
		"GET:/fhir/Organization/:id",
		"GET:/fhir/Location",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
