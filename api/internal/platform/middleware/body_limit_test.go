package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestParseLimit(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"1M", 1 << 20},
		{"10M", 10 << 20},
		{"512K", 512 << 10},
		{"1G", 1 << 30},
		{"1024", 1024},
		{"", 1 << 20},       // default
		{"invalid", 1 << 20}, // default on error
	}

	for _, tt := range tests {
		got := parseLimit(tt.input)
		if got != tt.want {
			t.Errorf("parseLimit(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestBodyLimit_AllowsSmallBody(t *testing.T) {
	e := echo.New()
	body := strings.NewReader(`{"resourceType":"Patient"}`)
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", body)
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	handler := func(c echo.Context) error {
		// Read the body to verify it is accessible
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		if len(b) == 0 {
			t.Error("expected non-empty body")
		}
		called = true
		return c.String(http.StatusCreated, "created")
	}

	mw := BodyLimit("1M", "10M")
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestBodyLimit_RejectsOversizedBody_ContentLength(t *testing.T) {
	e := echo.New()
	// Create a body larger than 1K limit
	largeBody := bytes.Repeat([]byte("x"), 2048)
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", bytes.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		t.Error("handler should not be called when body exceeds limit")
		return c.String(http.StatusCreated, "created")
	}

	mw := BodyLimit("1K", "10M")
	h := mw(handler)
	err := h(c)

	// The middleware returns a JSON response directly for Content-Length rejection
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected status 413, got %d", rec.Code)
	}

	// Verify FHIR OperationOutcome response
	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if outcome["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", outcome["resourceType"])
	}
}

func TestBodyLimit_UsesLargerLimitForFHIRBundle(t *testing.T) {
	e := echo.New()
	// Create a body that is larger than 1K but less than 10M
	bundleBody := bytes.Repeat([]byte("x"), 2048)
	req := httptest.NewRequest(http.MethodPost, "/fhir", bytes.NewReader(bundleBody))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	// Default limit is 1K but bundle limit is 10M — the bundle endpoint
	// should use the larger limit.
	mw := BodyLimit("1K", "10M")
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called for bundle endpoint within limit")
	}
}

func TestBodyLimit_RejectsBundleOverLimit(t *testing.T) {
	e := echo.New()
	// Create a body larger than 1K bundle limit
	largeBundle := bytes.Repeat([]byte("x"), 2048)
	req := httptest.NewRequest(http.MethodPost, "/fhir", bytes.NewReader(largeBundle))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		t.Error("handler should not be called when bundle exceeds limit")
		return c.String(http.StatusOK, "ok")
	}

	// Both limits are small so the bundle will exceed them.
	mw := BodyLimit("512", "1K")
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected status 413, got %d", rec.Code)
	}
}

func TestBodyLimit_SkipsNilBody(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	mw := BodyLimit("1M", "10M")
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called for GET with no body")
	}
}

func TestBodyLimit_EnforcesLimitDuringRead(t *testing.T) {
	e := echo.New()
	// Create a body that exceeds 512 bytes but don't set Content-Length
	largeBody := bytes.Repeat([]byte("a"), 1024)
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", bytes.NewReader(largeBody))
	req.ContentLength = -1 // Unknown content length
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		// Attempt to read the full body — should fail partway through
		_, err := io.ReadAll(c.Request().Body)
		return err
	}

	mw := BodyLimit("512", "10M")
	h := mw(handler)
	err := h(c)

	if err == nil {
		t.Fatal("expected error when reading body exceeds limit")
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", httpErr.Code)
	}
}
