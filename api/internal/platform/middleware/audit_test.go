package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/ehr/ehr/internal/platform/auth"
)

// mockRecorder collects audit entries for assertions.
type mockRecorder struct {
	mu      sync.Mutex
	entries []AuditEntry
	err     error // if set, RecordAccess returns this error
}

func (m *mockRecorder) RecordAccess(entry AuditEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, entry)
	return m.err
}

func (m *mockRecorder) last() AuditEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.entries[len(m.entries)-1]
}

func (m *mockRecorder) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.entries)
}

// newTestContext creates an echo context with optional auth context values set.
func newTestContext(method, path string, opts ...func(*http.Request)) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	for _, opt := range opts {
		opt(req)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func withAuth(userID string, roles []string) func(*http.Request) {
	return func(req *http.Request) {
		ctx := req.Context()
		ctx = context.WithValue(ctx, auth.UserIDKey, userID)
		ctx = context.WithValue(ctx, auth.UserRolesKey, roles)
		*req = *req.WithContext(ctx)
	}
}

func withBreakGlass(reason string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set("X-Break-Glass", reason)
	}
}

func okHandler(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

// --- Tests ---

func TestAudit_FHIRPatientRead(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rec := &mockRecorder{}
	patientID := uuid.New().String()

	c, _ := newTestContext(http.MethodGet,
		fmt.Sprintf("/fhir/Patient/%s", patientID),
		withAuth("user-1", []string{"physician"}),
	)
	c.Set("request_id", "req-abc")

	mw := Audit(logger, rec)
	h := mw(okHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.count() != 1 {
		t.Fatalf("expected 1 audit entry, got %d", rec.count())
	}
	entry := rec.last()
	if entry.UserID != "user-1" {
		t.Errorf("expected user_id 'user-1', got %q", entry.UserID)
	}
	if entry.ResourceType != "Patient" {
		t.Errorf("expected resource_type 'Patient', got %q", entry.ResourceType)
	}
	if entry.PatientID != patientID {
		t.Errorf("expected patient_id %q, got %q", patientID, entry.PatientID)
	}
	if entry.Action != "read" {
		t.Errorf("expected action 'read', got %q", entry.Action)
	}
	if entry.RequestID != "req-abc" {
		t.Errorf("expected request_id 'req-abc', got %q", entry.RequestID)
	}
	if entry.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", entry.StatusCode)
	}
}

func TestAudit_FHIRObservationCreate(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rec := &mockRecorder{}

	c, _ := newTestContext(http.MethodPost,
		"/fhir/Observation?patient=Patient/p-123",
		withAuth("user-2", []string{"nurse"}),
	)

	mw := Audit(logger, rec)
	h := mw(okHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entry := rec.last()
	if entry.Action != "create" {
		t.Errorf("expected action 'create', got %q", entry.Action)
	}
	if entry.ResourceType != "Observation" {
		t.Errorf("expected resource_type 'Observation', got %q", entry.ResourceType)
	}
	if entry.PatientID != "p-123" {
		t.Errorf("expected patient_id 'p-123', got %q", entry.PatientID)
	}
}

func TestAudit_APIRoute(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rec := &mockRecorder{}

	patientID := uuid.New().String()
	c, _ := newTestContext(http.MethodPut,
		fmt.Sprintf("/api/v1/patients/%s", patientID),
		withAuth("user-3", []string{"admin"}),
	)

	mw := Audit(logger, rec)
	h := mw(okHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entry := rec.last()
	if entry.Action != "update" {
		t.Errorf("expected action 'update', got %q", entry.Action)
	}
	if entry.ResourceType != "patients" {
		t.Errorf("expected resource_type 'patients', got %q", entry.ResourceType)
	}
	if entry.PatientID != patientID {
		t.Errorf("expected patient_id %q, got %q", patientID, entry.PatientID)
	}
}

func TestAudit_BreakGlass(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rec := &mockRecorder{}

	patientID := uuid.New().String()
	c, _ := newTestContext(http.MethodGet,
		fmt.Sprintf("/fhir/Patient/%s", patientID),
		withAuth("user-4", []string{"physician"}),
		withBreakGlass("emergency cardiac arrest"),
	)

	mw := Audit(logger, rec)
	h := mw(okHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entry := rec.last()
	if !entry.IsBreakGlass {
		t.Error("expected is_break_glass to be true")
	}
	if entry.BreakGlassReason != "emergency cardiac arrest" {
		t.Errorf("expected break_glass_reason 'emergency cardiac arrest', got %q", entry.BreakGlassReason)
	}
}

func TestAudit_SkipsNonAuditablePaths(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rec := &mockRecorder{}

	paths := []string{"/health", "/metrics", "/", "/other/path"}
	for _, path := range paths {
		c, _ := newTestContext(http.MethodGet, path)
		mw := Audit(logger, rec)
		h := mw(okHandler)
		err := h(c)
		if err != nil {
			t.Fatalf("unexpected error for path %s: %v", path, err)
		}
	}

	if rec.count() != 0 {
		t.Errorf("expected 0 audit entries for non-auditable paths, got %d", rec.count())
	}
}

func TestAudit_DeleteAction(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rec := &mockRecorder{}

	c, _ := newTestContext(http.MethodDelete,
		"/fhir/Condition/cond-1",
		withAuth("user-5", []string{"admin"}),
	)

	mw := Audit(logger, rec)
	h := mw(okHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entry := rec.last()
	if entry.Action != "delete" {
		t.Errorf("expected action 'delete', got %q", entry.Action)
	}
	if entry.ResourceType != "Condition" {
		t.Errorf("expected resource_type 'Condition', got %q", entry.ResourceType)
	}
}

func TestAudit_RecorderError_DoesNotBreakRequest(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rec := &mockRecorder{err: errors.New("database connection failed")}

	c, _ := newTestContext(http.MethodGet,
		"/fhir/Patient",
		withAuth("user-6", []string{"physician"}),
	)

	mw := Audit(logger, rec)
	h := mw(okHandler)
	err := h(c)

	// The request should still succeed even if the recorder fails
	if err != nil {
		t.Fatalf("expected no error even when recorder fails, got: %v", err)
	}
}

func TestAudit_NoRecorder_LogOnly(t *testing.T) {
	logger := zerolog.New(os.Stderr)

	c, _ := newTestContext(http.MethodGet,
		"/fhir/Patient",
		withAuth("user-7", []string{"physician"}),
	)

	// Pass no recorder -- should only log, not panic
	mw := Audit(logger)
	h := mw(okHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAudit_PatientIDFromQuery(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rec := &mockRecorder{}

	c, _ := newTestContext(http.MethodGet,
		"/fhir/Observation?patient=patient-abc",
		withAuth("user-8", []string{"nurse"}),
	)

	mw := Audit(logger, rec)
	h := mw(okHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entry := rec.last()
	if entry.PatientID != "patient-abc" {
		t.Errorf("expected patient_id 'patient-abc', got %q", entry.PatientID)
	}
}

func TestAudit_CapturesIPAndUserAgent(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rec := &mockRecorder{}

	c, _ := newTestContext(http.MethodGet,
		"/fhir/Patient",
		withAuth("user-9", []string{"physician"}),
		func(req *http.Request) {
			req.Header.Set("User-Agent", "EHR-Client/1.0")
		},
	)

	mw := Audit(logger, rec)
	h := mw(okHandler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entry := rec.last()
	if entry.UserAgent != "EHR-Client/1.0" {
		t.Errorf("expected user_agent 'EHR-Client/1.0', got %q", entry.UserAgent)
	}
	// IP should be non-empty (httptest uses 192.0.2.1 by default)
	if entry.IPAddress == "" {
		t.Error("expected non-empty IP address")
	}
}

// --- Unit tests for helper functions ---

func TestIsAuditablePath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/fhir/Patient", true},
		{"/fhir/Observation/123", true},
		{"/api/v1/patients", true},
		{"/api/v1/encounters/abc", true},
		{"/health", false},
		{"/", false},
		{"/fhir", false}, // no trailing slash
		{"/api/v1", false},
	}
	for _, tt := range tests {
		if got := isAuditablePath(tt.path); got != tt.want {
			t.Errorf("isAuditablePath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestHttpMethodToAction(t *testing.T) {
	tests := []struct {
		method string
		want   string
	}{
		{http.MethodGet, "read"},
		{http.MethodHead, "read"},
		{http.MethodPost, "create"},
		{http.MethodPut, "update"},
		{http.MethodPatch, "update"},
		{http.MethodDelete, "delete"},
		{http.MethodOptions, "read"},
	}
	for _, tt := range tests {
		if got := httpMethodToAction(tt.method); got != tt.want {
			t.Errorf("httpMethodToAction(%q) = %q, want %q", tt.method, got, tt.want)
		}
	}
}

func TestExtractResourceType(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/fhir/Patient", "Patient"},
		{"/fhir/Patient/123", "Patient"},
		{"/fhir/Observation", "Observation"},
		{"/api/v1/patients", "patients"},
		{"/api/v1/patients/123", "patients"},
		{"/api/v1/encounters/abc/notes", "encounters"},
		{"/other/path", "unknown"},
	}
	for _, tt := range tests {
		if got := extractResourceType(tt.path); got != tt.want {
			t.Errorf("extractResourceType(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestExtractPatientID(t *testing.T) {
	patientUUID := uuid.New().String()

	tests := []struct {
		name string
		path string
		want string
	}{
		{"fhir patient path", fmt.Sprintf("/fhir/Patient/%s", patientUUID), patientUUID},
		{"api patient path", fmt.Sprintf("/api/v1/patients/%s", patientUUID), patientUUID},
		{"query param plain", "/fhir/Observation?patient=p-123", "p-123"},
		{"query param fhir ref", "/fhir/Observation?patient=Patient/p-456", "p-456"},
		{"no patient", "/fhir/Medication", ""},
		{"non-uuid fhir path", "/fhir/Patient/search", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newTestContext(http.MethodGet, tt.path)
			got := extractPatientID(c)
			if got != tt.want {
				t.Errorf("extractPatientID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsUUIDLike(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{uuid.New().String(), true},
		{"not-a-uuid", false},
		{"", false},
		{"12345678-1234-1234-1234-123456789012", true},
	}
	for _, tt := range tests {
		if got := isUUIDLike(tt.input); got != tt.want {
			t.Errorf("isUUIDLike(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestAuditRecorderFunc(t *testing.T) {
	var called bool
	fn := AuditRecorderFunc(func(entry AuditEntry) error {
		called = true
		return nil
	})

	err := fn.RecordAccess(AuditEntry{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected function to be called")
	}
}
