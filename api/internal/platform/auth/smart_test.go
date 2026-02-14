package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func TestSMARTConfigurationEndpoint(t *testing.T) {
	e := echo.New()
	g := e.Group("/fhir")
	RegisterSMARTEndpoints(g, "http://localhost:8080/realms/ehr")

	req := httptest.NewRequest(http.MethodGet, "/fhir/.well-known/smart-configuration", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var cfg SMARTConfiguration
	if err := json.Unmarshal(rec.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if cfg.AuthorizationEndpoint == "" {
		t.Error("expected authorization_endpoint to be set")
	}
	if cfg.TokenEndpoint == "" {
		t.Error("expected token_endpoint to be set")
	}
	if len(cfg.Scopes) == 0 {
		t.Error("expected scopes to be populated")
	}
	if len(cfg.Capabilities) == 0 {
		t.Error("expected capabilities to be populated")
	}

	// Verify required SMART scopes
	scopeMap := make(map[string]bool)
	for _, s := range cfg.Scopes {
		scopeMap[s] = true
	}
	requiredScopes := []string{"openid", "profile", "fhirUser", "launch", "launch/patient", "patient/*.read", "patient/*.write", "user/*.read", "user/*.write"}
	for _, s := range requiredScopes {
		if !scopeMap[s] {
			t.Errorf("missing required scope: %s", s)
		}
	}

	// Verify required capabilities
	capMap := make(map[string]bool)
	for _, c := range cfg.Capabilities {
		capMap[c] = true
	}
	requiredCaps := []string{"launch-ehr", "launch-standalone", "client-public", "client-confidential-symmetric", "context-ehr-patient", "permission-patient", "permission-user"}
	for _, c := range requiredCaps {
		if !capMap[c] {
			t.Errorf("missing required capability: %s", c)
		}
	}

	// Verify response types
	if len(cfg.ResponseTypes) == 0 || cfg.ResponseTypes[0] != "code" {
		t.Error("expected response_types_supported to include 'code'")
	}

	// Verify PKCE support
	if len(cfg.CodeChallengeMethodsSupported) == 0 || cfg.CodeChallengeMethodsSupported[0] != "S256" {
		t.Error("expected code_challenge_methods_supported to include 'S256'")
	}
}

func TestSMARTConfigurationEndpoints(t *testing.T) {
	e := echo.New()
	g := e.Group("/fhir")
	RegisterSMARTEndpoints(g, "http://keycloak:8080/realms/ehr")

	req := httptest.NewRequest(http.MethodGet, "/fhir/.well-known/smart-configuration", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var cfg SMARTConfiguration
	json.Unmarshal(rec.Body.Bytes(), &cfg)

	expectedAuth := "http://keycloak:8080/realms/ehr/protocol/openid-connect/auth"
	if cfg.AuthorizationEndpoint != expectedAuth {
		t.Errorf("expected authorization_endpoint %q, got %q", expectedAuth, cfg.AuthorizationEndpoint)
	}

	expectedToken := "http://keycloak:8080/realms/ehr/protocol/openid-connect/token"
	if cfg.TokenEndpoint != expectedToken {
		t.Errorf("expected token_endpoint %q, got %q", expectedToken, cfg.TokenEndpoint)
	}
}

func TestParseSMARTScope(t *testing.T) {
	tests := []struct {
		name      string
		scope     string
		wantCtx   string
		wantRes   string
		wantOp    string
		wantErr   bool
	}{
		{
			name:    "patient read",
			scope:   "patient/Patient.read",
			wantCtx: "patient",
			wantRes: "Patient",
			wantOp:  "read",
		},
		{
			name:    "user write",
			scope:   "user/Observation.write",
			wantCtx: "user",
			wantRes: "Observation",
			wantOp:  "write",
		},
		{
			name:    "patient wildcard resource read",
			scope:   "patient/*.read",
			wantCtx: "patient",
			wantRes: "*",
			wantOp:  "read",
		},
		{
			name:    "user wildcard all",
			scope:   "user/*.*",
			wantCtx: "user",
			wantRes: "*",
			wantOp:  "*",
		},
		{
			name:    "system scope",
			scope:   "system/Patient.read",
			wantCtx: "system",
			wantRes: "Patient",
			wantOp:  "read",
		},
		{
			name:    "patient wildcard write",
			scope:   "patient/*.write",
			wantCtx: "patient",
			wantRes: "*",
			wantOp:  "write",
		},
		{
			name:    "non-resource scope openid",
			scope:   "openid",
			wantErr: true,
		},
		{
			name:    "non-resource scope profile",
			scope:   "profile",
			wantErr: true,
		},
		{
			name:    "non-resource scope launch",
			scope:   "launch",
			wantErr: true,
		},
		{
			name:    "launch/patient is not a resource scope",
			scope:   "launch/patient",
			wantErr: true,
		},
		{
			name:    "invalid context",
			scope:   "admin/Patient.read",
			wantErr: true,
		},
		{
			name:    "missing operation",
			scope:   "patient/Patient",
			wantErr: true,
		},
		{
			name:    "invalid operation",
			scope:   "patient/Patient.delete",
			wantErr: true,
		},
		{
			name:    "empty resource type",
			scope:   "patient/.read",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := ParseSMARTScope(tt.scope)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for scope %q, got nil", tt.scope)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if s.Context != tt.wantCtx {
				t.Errorf("context = %q, want %q", s.Context, tt.wantCtx)
			}
			if s.ResourceType != tt.wantRes {
				t.Errorf("resourceType = %q, want %q", s.ResourceType, tt.wantRes)
			}
			if s.Operation != tt.wantOp {
				t.Errorf("operation = %q, want %q", s.Operation, tt.wantOp)
			}
		})
	}
}

func TestParseSMARTScopes(t *testing.T) {
	scopes := []string{
		"openid",
		"profile",
		"fhirUser",
		"launch",
		"patient/Patient.read",
		"user/Observation.write",
		"patient/*.read",
	}

	parsed := ParseSMARTScopes(scopes)
	if len(parsed) != 3 {
		t.Fatalf("expected 3 parsed scopes, got %d", len(parsed))
	}

	// Verify first parsed scope
	if parsed[0].Context != "patient" || parsed[0].ResourceType != "Patient" || parsed[0].Operation != "read" {
		t.Errorf("unexpected first scope: %+v", parsed[0])
	}
}

func TestScopeAllows(t *testing.T) {
	tests := []struct {
		name         string
		scopes       []SMARTScope
		resourceType string
		operation    string
		want         bool
	}{
		{
			name: "exact match allows",
			scopes: []SMARTScope{
				{Context: "patient", ResourceType: "Patient", Operation: "read"},
			},
			resourceType: "Patient",
			operation:    "read",
			want:         true,
		},
		{
			name: "wildcard resource allows",
			scopes: []SMARTScope{
				{Context: "patient", ResourceType: "*", Operation: "read"},
			},
			resourceType: "Observation",
			operation:    "read",
			want:         true,
		},
		{
			name: "wildcard operation allows",
			scopes: []SMARTScope{
				{Context: "user", ResourceType: "Patient", Operation: "*"},
			},
			resourceType: "Patient",
			operation:    "write",
			want:         true,
		},
		{
			name: "wildcard both allows",
			scopes: []SMARTScope{
				{Context: "user", ResourceType: "*", Operation: "*"},
			},
			resourceType: "Encounter",
			operation:    "write",
			want:         true,
		},
		{
			name: "wrong resource denies",
			scopes: []SMARTScope{
				{Context: "patient", ResourceType: "Patient", Operation: "read"},
			},
			resourceType: "Observation",
			operation:    "read",
			want:         false,
		},
		{
			name: "wrong operation denies",
			scopes: []SMARTScope{
				{Context: "patient", ResourceType: "Patient", Operation: "read"},
			},
			resourceType: "Patient",
			operation:    "write",
			want:         false,
		},
		{
			name:         "empty scopes denies",
			scopes:       nil,
			resourceType: "Patient",
			operation:    "read",
			want:         false,
		},
		{
			name: "multiple scopes one matches",
			scopes: []SMARTScope{
				{Context: "patient", ResourceType: "Patient", Operation: "read"},
				{Context: "user", ResourceType: "Observation", Operation: "write"},
			},
			resourceType: "Observation",
			operation:    "write",
			want:         true,
		},
		{
			name: "multiple scopes none match",
			scopes: []SMARTScope{
				{Context: "patient", ResourceType: "Patient", Operation: "read"},
				{Context: "user", ResourceType: "Observation", Operation: "read"},
			},
			resourceType: "Encounter",
			operation:    "write",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScopeAllows(tt.scopes, tt.resourceType, tt.operation)
			if got != tt.want {
				t.Errorf("ScopeAllows() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLaunchContextStore(t *testing.T) {
	store := NewLaunchContextStore(5 * time.Minute)

	t.Run("create and get", func(t *testing.T) {
		ctx, err := store.Create("patient-123", "encounter-456", "Practitioner/dr-smith")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ctx.LaunchToken == "" {
			t.Fatal("expected non-empty launch token")
		}
		if ctx.PatientID != "patient-123" {
			t.Errorf("expected patient-123, got %s", ctx.PatientID)
		}
		if ctx.EncounterID != "encounter-456" {
			t.Errorf("expected encounter-456, got %s", ctx.EncounterID)
		}
		if ctx.FHIRUser != "Practitioner/dr-smith" {
			t.Errorf("expected Practitioner/dr-smith, got %s", ctx.FHIRUser)
		}

		// Get should return the context
		got := store.Get(ctx.LaunchToken)
		if got == nil {
			t.Fatal("expected to find context")
		}
		if got.PatientID != "patient-123" {
			t.Errorf("expected patient-123, got %s", got.PatientID)
		}
	})

	t.Run("consume removes context", func(t *testing.T) {
		ctx, _ := store.Create("patient-789", "", "")
		token := ctx.LaunchToken

		consumed := store.Consume(token)
		if consumed == nil {
			t.Fatal("expected to consume context")
		}
		if consumed.PatientID != "patient-789" {
			t.Errorf("expected patient-789, got %s", consumed.PatientID)
		}

		// Second consume should return nil
		second := store.Consume(token)
		if second != nil {
			t.Error("expected nil on second consume")
		}

		// Get should also return nil
		got := store.Get(token)
		if got != nil {
			t.Error("expected nil after consume")
		}
	})

	t.Run("not found", func(t *testing.T) {
		got := store.Get("nonexistent-token")
		if got != nil {
			t.Error("expected nil for nonexistent token")
		}
	})
}

func TestLaunchContextStoreExpiry(t *testing.T) {
	store := NewLaunchContextStore(50 * time.Millisecond)

	ctx, _ := store.Create("patient-expire", "", "")
	token := ctx.LaunchToken

	// Should be available immediately
	got := store.Get(token)
	if got == nil {
		t.Fatal("expected context to be available immediately")
	}

	// Wait for expiry
	time.Sleep(100 * time.Millisecond)

	got = store.Get(token)
	if got != nil {
		t.Error("expected context to be expired")
	}
}

func TestLaunchContextStoreCleanup(t *testing.T) {
	store := NewLaunchContextStore(50 * time.Millisecond)

	store.Create("patient-1", "", "")
	store.Create("patient-2", "", "")

	time.Sleep(100 * time.Millisecond)

	store.Cleanup()

	store.mu.RLock()
	count := len(store.contexts)
	store.mu.RUnlock()

	if count != 0 {
		t.Errorf("expected 0 contexts after cleanup, got %d", count)
	}
}

func TestEHRLaunchEndpoint(t *testing.T) {
	e := echo.New()
	g := e.Group("/fhir")
	RegisterSMARTEndpoints(g, "http://localhost:8080/realms/ehr")

	t.Run("successful launch", func(t *testing.T) {
		body := `{"patient_id":"patient-123","encounter_id":"enc-456","fhir_user":"Practitioner/dr-smith"}`
		req := httptest.NewRequest(http.MethodPost, "/fhir/launch", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp ehrLaunchResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if resp.LaunchToken == "" {
			t.Error("expected non-empty launch token")
		}
		if resp.ISS == "" {
			t.Error("expected non-empty ISS")
		}
	})

	t.Run("missing patient_id", func(t *testing.T) {
		body := `{"encounter_id":"enc-456"}`
		req := httptest.NewRequest(http.MethodPost, "/fhir/launch", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})
}

func TestLaunchContextEndpoint(t *testing.T) {
	e := echo.New()
	g := e.Group("/fhir")
	RegisterSMARTEndpoints(g, "http://localhost:8080/realms/ehr")

	// First, create a launch context
	body := `{"patient_id":"patient-ctx","encounter_id":"enc-ctx","fhir_user":"Practitioner/doc"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/launch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var launchResp ehrLaunchResponse
	json.Unmarshal(rec.Body.Bytes(), &launchResp)
	token := launchResp.LaunchToken

	t.Run("resolve launch context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/fhir/launch-context?launch="+token, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var ctx LaunchContext
		json.Unmarshal(rec.Body.Bytes(), &ctx)

		if ctx.PatientID != "patient-ctx" {
			t.Errorf("expected patient-ctx, got %s", ctx.PatientID)
		}
		if ctx.EncounterID != "enc-ctx" {
			t.Errorf("expected enc-ctx, got %s", ctx.EncounterID)
		}
		if ctx.FHIRUser != "Practitioner/doc" {
			t.Errorf("expected Practitioner/doc, got %s", ctx.FHIRUser)
		}
	})

	t.Run("second consume returns not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/fhir/launch-context?launch="+token, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404 on second consume, got %d", rec.Code)
		}
	})

	t.Run("missing launch param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/fhir/launch-context", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})
}

func TestRequireSMARTScope(t *testing.T) {
	t.Run("allows matching scope", func(t *testing.T) {
		e := echo.New()
		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		}

		mw := RequireSMARTScope("Patient")
		e.GET("/fhir/Patient", handler, mw)

		req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
		ctx := context.WithValue(req.Context(), SMARTScopesKey, []SMARTScope{
			{Context: "patient", ResourceType: "Patient", Operation: "read"},
		})
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("denies wrong resource", func(t *testing.T) {
		e := echo.New()
		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		}

		mw := RequireSMARTScope("Observation")
		e.GET("/fhir/Observation", handler, mw)

		req := httptest.NewRequest(http.MethodGet, "/fhir/Observation", nil)
		ctx := context.WithValue(req.Context(), SMARTScopesKey, []SMARTScope{
			{Context: "patient", ResourceType: "Patient", Operation: "read"},
		})
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("denies wrong operation", func(t *testing.T) {
		e := echo.New()
		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		}

		mw := RequireSMARTScope("Patient")
		e.POST("/fhir/Patient", handler, mw)

		req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
		ctx := context.WithValue(req.Context(), SMARTScopesKey, []SMARTScope{
			{Context: "patient", ResourceType: "Patient", Operation: "read"},
		})
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("allows wildcard resource", func(t *testing.T) {
		e := echo.New()
		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		}

		mw := RequireSMARTScope("Observation")
		e.GET("/fhir/Observation", handler, mw)

		req := httptest.NewRequest(http.MethodGet, "/fhir/Observation", nil)
		ctx := context.WithValue(req.Context(), SMARTScopesKey, []SMARTScope{
			{Context: "patient", ResourceType: "*", Operation: "read"},
		})
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("passes through with no SMART scopes", func(t *testing.T) {
		e := echo.New()
		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		}

		mw := RequireSMARTScope("Patient")
		e.GET("/fhir/Patient", handler, mw)

		req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 (passthrough), got %d", rec.Code)
		}
	})
}

func TestHttpMethodToOperation(t *testing.T) {
	tests := []struct {
		method string
		want   string
	}{
		{http.MethodGet, "read"},
		{http.MethodHead, "read"},
		{http.MethodPost, "write"},
		{http.MethodPut, "write"},
		{http.MethodPatch, "write"},
		{http.MethodDelete, "write"},
		{"OPTIONS", "read"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			got := httpMethodToOperation(tt.method)
			if got != tt.want {
				t.Errorf("httpMethodToOperation(%s) = %s, want %s", tt.method, got, tt.want)
			}
		})
	}
}

func TestSMARTContextHelpers(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, SMARTPatientIDKey, "patient-123")
	ctx = context.WithValue(ctx, SMARTEncounterIDKey, "enc-456")
	ctx = context.WithValue(ctx, SMARTFHIRUserKey, "Practitioner/dr-smith")
	ctx = context.WithValue(ctx, SMARTScopesKey, []SMARTScope{
		{Context: "patient", ResourceType: "Patient", Operation: "read"},
	})

	if got := SMARTPatientIDFromContext(ctx); got != "patient-123" {
		t.Errorf("SMARTPatientIDFromContext = %s, want patient-123", got)
	}
	if got := SMARTEncounterIDFromContext(ctx); got != "enc-456" {
		t.Errorf("SMARTEncounterIDFromContext = %s, want enc-456", got)
	}
	if got := SMARTFHIRUserFromContext(ctx); got != "Practitioner/dr-smith" {
		t.Errorf("SMARTFHIRUserFromContext = %s, want Practitioner/dr-smith", got)
	}
	scopes := SMARTScopesFromContext(ctx)
	if len(scopes) != 1 {
		t.Fatalf("expected 1 scope, got %d", len(scopes))
	}

	// Test empty context returns zero values
	emptyCtx := context.Background()
	if got := SMARTPatientIDFromContext(emptyCtx); got != "" {
		t.Errorf("expected empty string, got %s", got)
	}
	if got := SMARTEncounterIDFromContext(emptyCtx); got != "" {
		t.Errorf("expected empty string, got %s", got)
	}
	if got := SMARTFHIRUserFromContext(emptyCtx); got != "" {
		t.Errorf("expected empty string, got %s", got)
	}
	if scopes := SMARTScopesFromContext(emptyCtx); scopes != nil {
		t.Errorf("expected nil scopes, got %v", scopes)
	}
}

func TestGenerateLaunchToken(t *testing.T) {
	token1, err := generateLaunchToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(token1) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("expected 64 char token, got %d chars", len(token1))
	}

	token2, _ := generateLaunchToken()
	if token1 == token2 {
		t.Error("expected unique tokens")
	}
}
