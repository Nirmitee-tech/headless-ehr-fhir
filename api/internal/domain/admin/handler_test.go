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

// -- REST: Missing Handler Tests --

func TestHandler_UpdateOrganization(t *testing.T) {
	h, e := newTestHandler()
	org := &Organization{Name: "Test"}
	h.svc.CreateOrganization(nil, org)
	body := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(org.ID.String())
	if err := h.UpdateOrganization(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetDepartment(t *testing.T) {
	h, e := newTestHandler()
	dept := &Department{Name: "Cardiology", OrganizationID: uuid.New()}
	h.svc.CreateDepartment(nil, dept)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(dept.ID.String())
	if err := h.GetDepartment(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListDepartments(t *testing.T) {
	h, e := newTestHandler()
	orgID := uuid.New()
	h.svc.CreateDepartment(nil, &Department{Name: "D1", OrganizationID: orgID})
	req := httptest.NewRequest(http.MethodGet, "/?organization_id="+orgID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListDepartments(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateDepartment(t *testing.T) {
	h, e := newTestHandler()
	dept := &Department{Name: "D1", OrganizationID: uuid.New()}
	h.svc.CreateDepartment(nil, dept)
	body := `{"name":"D2"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(dept.ID.String())
	if err := h.UpdateDepartment(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteDepartment(t *testing.T) {
	h, e := newTestHandler()
	dept := &Department{Name: "D1", OrganizationID: uuid.New()}
	h.svc.CreateDepartment(nil, dept)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(dept.ID.String())
	if err := h.DeleteDepartment(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_GetLocation(t *testing.T) {
	h, e := newTestHandler()
	loc := &Location{Name: "Main"}
	h.svc.CreateLocation(nil, loc)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(loc.ID.String())
	if err := h.GetLocation(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListLocations(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateLocation(nil, &Location{Name: "L1"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListLocations(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateLocation(t *testing.T) {
	h, e := newTestHandler()
	loc := &Location{Name: "L1"}
	h.svc.CreateLocation(nil, loc)
	body := `{"name":"L2"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(loc.ID.String())
	if err := h.UpdateLocation(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteLocation(t *testing.T) {
	h, e := newTestHandler()
	loc := &Location{Name: "L1"}
	h.svc.CreateLocation(nil, loc)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(loc.ID.String())
	if err := h.DeleteLocation(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_GetSystemUser(t *testing.T) {
	h, e := newTestHandler()
	u := &SystemUser{Username: "jdoe", UserType: "provider"}
	h.svc.CreateSystemUser(nil, u)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(u.ID.String())
	if err := h.GetSystemUser(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListSystemUsers(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateSystemUser(nil, &SystemUser{Username: "u1", UserType: "provider"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListSystemUsers(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateSystemUser(t *testing.T) {
	h, e := newTestHandler()
	u := &SystemUser{Username: "jdoe", UserType: "provider"}
	h.svc.CreateSystemUser(nil, u)
	body := `{"username":"jdoe","user_type":"admin"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(u.ID.String())
	if err := h.UpdateSystemUser(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteSystemUser(t *testing.T) {
	h, e := newTestHandler()
	u := &SystemUser{Username: "jdoe", UserType: "provider"}
	h.svc.CreateSystemUser(nil, u)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(u.ID.String())
	if err := h.DeleteSystemUser(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AssignRole(t *testing.T) {
	h, e := newTestHandler()
	u := &SystemUser{Username: "jdoe", UserType: "provider"}
	h.svc.CreateSystemUser(nil, u)
	body := `{"role_name":"admin"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(u.ID.String())
	if err := h.AssignRole(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetUserRoles(t *testing.T) {
	h, e := newTestHandler()
	u := &SystemUser{Username: "jdoe", UserType: "provider"}
	h.svc.CreateSystemUser(nil, u)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(u.ID.String())
	if err := h.GetUserRoles(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_RemoveRole(t *testing.T) {
	h, e := newTestHandler()
	u := &SystemUser{Username: "jdoe", UserType: "provider"}
	h.svc.CreateSystemUser(nil, u)
	role := &UserRoleAssignment{UserID: u.ID, RoleName: "admin"}
	h.svc.AssignRole(nil, role)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "role_id")
	c.SetParamValues(u.ID.String(), role.ID.String())
	if err := h.RemoveRole(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// -- FHIR Organization Handlers --

func TestHandler_SearchOrganizationsFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateOrganization(nil, &Organization{Name: "Test"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.SearchOrganizationsFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetOrganizationFHIR(t *testing.T) {
	h, e := newTestHandler()
	org := &Organization{Name: "Test"}
	h.svc.CreateOrganization(nil, org)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(org.FHIRID)
	if err := h.GetOrganizationFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetOrganizationFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	if err := h.GetOrganizationFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateOrganizationFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"name":"FHIR Org"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.CreateOrganizationFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	if rec.Header().Get("Location") == "" {
		t.Error("expected Location header")
	}
}

func TestHandler_UpdateOrganizationFHIR(t *testing.T) {
	h, e := newTestHandler()
	org := &Organization{Name: "Test"}
	h.svc.CreateOrganization(nil, org)
	body := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(org.FHIRID)
	if err := h.UpdateOrganizationFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteOrganizationFHIR(t *testing.T) {
	h, e := newTestHandler()
	org := &Organization{Name: "Test"}
	h.svc.CreateOrganization(nil, org)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(org.FHIRID)
	if err := h.DeleteOrganizationFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchOrganizationFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	org := &Organization{Name: "Test"}
	h.svc.CreateOrganization(nil, org)
	body := `{"name":"Patched"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(org.FHIRID)
	if err := h.PatchOrganizationFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadOrganizationFHIR(t *testing.T) {
	h, e := newTestHandler()
	org := &Organization{Name: "Test"}
	h.svc.CreateOrganization(nil, org)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(org.FHIRID, "1")
	if err := h.VreadOrganizationFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryOrganizationFHIR(t *testing.T) {
	h, e := newTestHandler()
	org := &Organization{Name: "Test"}
	h.svc.CreateOrganization(nil, org)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(org.FHIRID)
	if err := h.HistoryOrganizationFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- FHIR Location Handlers --

func TestHandler_SearchLocationsFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateLocation(nil, &Location{Name: "L1"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.SearchLocationsFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetLocationFHIR(t *testing.T) {
	h, e := newTestHandler()
	loc := &Location{Name: "L1"}
	h.svc.CreateLocation(nil, loc)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(loc.FHIRID)
	if err := h.GetLocationFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetLocationFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	if err := h.GetLocationFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateLocationFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"name":"FHIR Loc"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.CreateLocationFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_UpdateLocationFHIR(t *testing.T) {
	h, e := newTestHandler()
	loc := &Location{Name: "L1"}
	h.svc.CreateLocation(nil, loc)
	body := `{"name":"L2"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(loc.FHIRID)
	if err := h.UpdateLocationFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteLocationFHIR(t *testing.T) {
	h, e := newTestHandler()
	loc := &Location{Name: "L1"}
	h.svc.CreateLocation(nil, loc)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(loc.FHIRID)
	if err := h.DeleteLocationFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_PatchLocationFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	loc := &Location{Name: "L1"}
	h.svc.CreateLocation(nil, loc)
	body := `{"name":"Patched"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(loc.FHIRID)
	if err := h.PatchLocationFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_VreadLocationFHIR(t *testing.T) {
	h, e := newTestHandler()
	loc := &Location{Name: "L1"}
	h.svc.CreateLocation(nil, loc)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(loc.FHIRID, "1")
	if err := h.VreadLocationFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryLocationFHIR(t *testing.T) {
	h, e := newTestHandler()
	loc := &Location{Name: "L1"}
	h.svc.CreateLocation(nil, loc)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(loc.FHIRID)
	if err := h.HistoryLocationFHIR(c); err != nil {
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
