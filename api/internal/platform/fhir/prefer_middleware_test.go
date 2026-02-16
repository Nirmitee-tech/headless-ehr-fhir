package fhir

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestPreferMiddleware_NoHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := PreferMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusCreated, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	if rec.Body.String() != `{"resourceType":"Patient","id":"1"}` {
		t.Errorf("expected full body, got: %s", rec.Body.String())
	}
}

func TestPreferMiddleware_ReturnMinimal(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("Prefer", "return=minimal")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := PreferMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusCreated, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body for minimal, got: %s", rec.Body.String())
	}
}

func TestPreferMiddleware_ReturnOperationOutcome(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/fhir/Patient/1", nil)
	req.Header.Set("Prefer", "return=OperationOutcome")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := PreferMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if body == "" || !strings.Contains(body, "OperationOutcome") {
		t.Errorf("expected OperationOutcome body, got: %s", body)
	}
}

func TestPreferMiddleware_GetRequestIgnored(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/1", nil)
	req.Header.Set("Prefer", "return=minimal")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := PreferMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	// GET requests should not be affected by Prefer
	if rec.Body.String() != `{"resourceType":"Patient","id":"1"}` {
		t.Errorf("expected full body for GET, got: %s", rec.Body.String())
	}
}

func TestParsePreferReturn(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"return=minimal", "minimal"},
		{"return=representation", "representation"},
		{"return=OperationOutcome", "OperationOutcome"},
		{"handling=strict; return=minimal", "minimal"},
		{"return=minimal, handling=strict", "minimal"},
		{"", ""},
		{"handling=strict", ""},
	}
	for _, tt := range tests {
		got := parsePreferReturn(tt.input)
		if got != tt.expected {
			t.Errorf("parsePreferReturn(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

