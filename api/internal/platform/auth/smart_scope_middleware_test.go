package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// newScopeTestContext creates an echo.Context for testing FHIRScopeMiddleware.
// It sets up the request context with the provided roles and scopes.
func newScopeTestContext(method, path string, roles []string, scopes []string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	ctx := req.Context()
	if roles != nil {
		ctx = context.WithValue(ctx, UserRolesKey, roles)
	}
	if scopes != nil {
		ctx = context.WithValue(ctx, UserScopesKey, scopes)
	}
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// scopeOkHandler is a simple handler that returns 200 OK to signal the middleware passed through.
var scopeOkHandler = func(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

// parseFHIROutcome parses the response body as a FHIR OperationOutcome and
// returns the diagnostics string from the first issue.
func parseFHIROutcome(rec *httptest.ResponseRecorder) (string, error) {
	var outcome struct {
		ResourceType string `json:"resourceType"`
		Issue        []struct {
			Severity    string `json:"severity"`
			Code        string `json:"code"`
			Diagnostics string `json:"diagnostics"`
		} `json:"issue"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		return "", err
	}
	if len(outcome.Issue) == 0 {
		return "", nil
	}
	return outcome.Issue[0].Diagnostics, nil
}

func TestFHIRScopeMiddleware_ReadAllowed(t *testing.T) {
	// GET /fhir/Patient/123 with scope "user/Patient.read" -> allowed
	c, rec := newScopeTestContext(http.MethodGet, "/fhir/Patient/123", []string{"physician"}, []string{"user/Patient.read"})

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_WrongResource(t *testing.T) {
	// GET /fhir/Patient/123 with scope "user/Observation.read" -> forbidden
	c, rec := newScopeTestContext(http.MethodGet, "/fhir/Patient/123", []string{"physician"}, []string{"user/Observation.read"})

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected echo error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}

	diag, parseErr := parseFHIROutcome(rec)
	if parseErr != nil {
		t.Fatalf("failed to parse OperationOutcome: %v", parseErr)
	}
	if diag != "insufficient scope: required Patient.read" {
		t.Errorf("unexpected diagnostics: %s", diag)
	}
}

func TestFHIRScopeMiddleware_CreateAllowed(t *testing.T) {
	// POST /fhir/Patient with scope "user/Patient.write" -> allowed (create)
	c, rec := newScopeTestContext(http.MethodPost, "/fhir/Patient", []string{"physician"}, []string{"user/Patient.write"})

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_SearchPOSTAllowed(t *testing.T) {
	// POST /fhir/Patient/_search with scope "user/Patient.read" -> allowed (search is read)
	c, rec := newScopeTestContext(http.MethodPost, "/fhir/Patient/_search", []string{"physician"}, []string{"user/Patient.read"})

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_WriteRequired(t *testing.T) {
	// PUT /fhir/Patient/123 with scope "user/Patient.read" -> forbidden (write needed)
	c, rec := newScopeTestContext(http.MethodPut, "/fhir/Patient/123", []string{"physician"}, []string{"user/Patient.read"})

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected echo error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}

	diag, parseErr := parseFHIROutcome(rec)
	if parseErr != nil {
		t.Fatalf("failed to parse OperationOutcome: %v", parseErr)
	}
	if diag != "insufficient scope: required Patient.write" {
		t.Errorf("unexpected diagnostics: %s", diag)
	}
}

func TestFHIRScopeMiddleware_DeleteAllowed(t *testing.T) {
	// DELETE /fhir/Patient/123 with scope "user/Patient.write" -> allowed
	c, rec := newScopeTestContext(http.MethodDelete, "/fhir/Patient/123", []string{"physician"}, []string{"user/Patient.write"})

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_WildcardResource(t *testing.T) {
	// GET /fhir/Patient/123 with scope "user/*.read" -> allowed (wildcard resource)
	c, rec := newScopeTestContext(http.MethodGet, "/fhir/Patient/123", []string{"physician"}, []string{"user/*.read"})

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_WildcardAll(t *testing.T) {
	// GET /fhir/Patient/123 with scope "user/*.*" -> allowed (wildcard all)
	c, rec := newScopeTestContext(http.MethodGet, "/fhir/Patient/123", []string{"physician"}, []string{"user/*.*"})

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_PatientContext(t *testing.T) {
	// GET /fhir/Patient/123 with scope "patient/Patient.read" -> allowed (patient context)
	c, rec := newScopeTestContext(http.MethodGet, "/fhir/Patient/123", []string{"physician"}, []string{"patient/Patient.read"})

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_MetadataSkip(t *testing.T) {
	// GET /fhir/metadata with any scope -> allowed (skip check)
	c, rec := newScopeTestContext(http.MethodGet, "/fhir/metadata", []string{"physician"}, []string{"user/Observation.read"})

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_AdminBypass(t *testing.T) {
	// Admin role with no matching scope -> allowed (admin bypass)
	c, rec := newScopeTestContext(http.MethodGet, "/fhir/Patient/123", []string{"admin"}, []string{"user/Observation.read"})

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_NoScopesDevMode(t *testing.T) {
	// No scopes in context -> allowed (dev mode backward compat)
	c, rec := newScopeTestContext(http.MethodGet, "/fhir/Patient/123", []string{"physician"}, nil)

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_EmptyScopesDevMode(t *testing.T) {
	// Empty scopes slice -> allowed (dev mode backward compat)
	c, rec := newScopeTestContext(http.MethodGet, "/fhir/Patient/123", []string{"physician"}, []string{})

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_SearchGETAllowed(t *testing.T) {
	// GET /fhir/Patient with scope "user/Patient.read" -> allowed (search)
	c, rec := newScopeTestContext(http.MethodGet, "/fhir/Patient", []string{"physician"}, []string{"user/Patient.read"})

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_WellKnownSkip(t *testing.T) {
	// GET /fhir/.well-known/smart-configuration -> allowed (skip check)
	c, rec := newScopeTestContext(http.MethodGet, "/fhir/.well-known/smart-configuration", nil, nil)

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_SystemScope(t *testing.T) {
	// GET /fhir/Patient/123 with scope "system/Patient.read" -> allowed
	c, rec := newScopeTestContext(http.MethodGet, "/fhir/Patient/123", []string{"physician"}, []string{"system/Patient.read"})

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_PATCHRequiresWrite(t *testing.T) {
	// PATCH /fhir/Patient/123 with scope "user/Patient.read" -> forbidden
	c, rec := newScopeTestContext(http.MethodPatch, "/fhir/Patient/123", []string{"physician"}, []string{"user/Patient.read"})

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected echo error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_MultipleScopes(t *testing.T) {
	// GET /fhir/Observation/456 with multiple scopes including a match -> allowed
	c, rec := newScopeTestContext(http.MethodGet, "/fhir/Observation/456",
		[]string{"physician"},
		[]string{"user/Patient.read", "user/Observation.read", "user/Encounter.write"},
	)

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_NonResourceScopesIgnored(t *testing.T) {
	// Non-resource scopes like "openid" and "profile" are silently skipped
	// during parsing; only the resource scope matters.
	c, rec := newScopeTestContext(http.MethodGet, "/fhir/Patient/123",
		[]string{"physician"},
		[]string{"openid", "profile", "user/Patient.read"},
	)

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_OnlyNonResourceScopes(t *testing.T) {
	// If only non-resource scopes like "openid" are present, after parsing
	// the smart scopes list is empty but rawScopes is non-empty. The middleware
	// should still enforce -> forbidden because rawScopes is non-empty.
	c, rec := newScopeTestContext(http.MethodGet, "/fhir/Patient/123",
		[]string{"physician"},
		[]string{"openid", "profile", "launch"},
	)

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected echo error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestFHIRScopeMiddleware_OperationOutcomeFormat(t *testing.T) {
	// Verify the forbidden response is a valid FHIR OperationOutcome.
	c, rec := newScopeTestContext(http.MethodGet, "/fhir/Patient/123", []string{"physician"}, []string{"user/Observation.read"})

	mw := FHIRScopeMiddleware()
	h := mw(scopeOkHandler)
	_ = h(c)

	var outcome struct {
		ResourceType string `json:"resourceType"`
		Issue        []struct {
			Severity    string `json:"severity"`
			Code        string `json:"code"`
			Diagnostics string `json:"diagnostics"`
		} `json:"issue"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response as JSON: %v", err)
	}
	if outcome.ResourceType != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %s", outcome.ResourceType)
	}
	if len(outcome.Issue) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(outcome.Issue))
	}
	if outcome.Issue[0].Severity != "error" {
		t.Errorf("expected severity error, got %s", outcome.Issue[0].Severity)
	}
	if outcome.Issue[0].Code != "forbidden" {
		t.Errorf("expected code forbidden, got %s", outcome.Issue[0].Code)
	}
}

// ---------------------------------------------------------------------------
// Unit tests for helper functions
// ---------------------------------------------------------------------------

func TestExtractFHIRResourceType(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/fhir/Patient/123", "Patient"},
		{"/fhir/Patient", "Patient"},
		{"/fhir/Patient/_search", "Patient"},
		{"/fhir/Observation/456", "Observation"},
		{"/fhir/metadata", ""},
		{"/fhir/$export", ""},
		{"/fhir/.well-known/smart-configuration", ""},
		{"/fhir/", ""},
		{"/fhir", ""},
		{"/api/v1/patients", ""},
		{"/other/Patient/123", ""},
	}

	for _, tt := range tests {
		got := extractFHIRResourceType(tt.path)
		if got != tt.want {
			t.Errorf("extractFHIRResourceType(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestFhirMethodToOperation(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   string
	}{
		{http.MethodGet, "/fhir/Patient/123", "read"},
		{http.MethodHead, "/fhir/Patient/123", "read"},
		{http.MethodPost, "/fhir/Patient", "write"},
		{http.MethodPost, "/fhir/Patient/_search", "read"},
		{http.MethodPut, "/fhir/Patient/123", "write"},
		{http.MethodPatch, "/fhir/Patient/123", "write"},
		{http.MethodDelete, "/fhir/Patient/123", "write"},
	}

	for _, tt := range tests {
		got := fhirMethodToOperation(tt.method, tt.path)
		if got != tt.want {
			t.Errorf("fhirMethodToOperation(%q, %q) = %q, want %q", tt.method, tt.path, got, tt.want)
		}
	}
}

func TestIsScopeExemptPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/fhir/metadata", true},
		{"/fhir/metadata/", true},
		{"/fhir/.well-known/smart-configuration", true},
		{"/fhir/.well-known/openid-configuration", true},
		{"/fhir/Patient/123", false},
		{"/fhir/Patient", false},
		{"/fhir/$export", false},
	}

	for _, tt := range tests {
		got := isScopeExemptPath(tt.path)
		if got != tt.want {
			t.Errorf("isScopeExemptPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestIsSearchPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/fhir/Patient/_search", true},
		{"/fhir/Patient/_search/", true},
		{"/fhir/Patient", false},
		{"/fhir/Patient/123", false},
	}

	for _, tt := range tests {
		got := isSearchPath(tt.path)
		if got != tt.want {
			t.Errorf("isSearchPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
