package fhir

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestParseSearchString(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]string
	}{
		{"identifier=foo&name=bar", map[string]string{"identifier": "foo", "name": "bar"}},
		{"?status=active", map[string]string{"status": "active"}},
		{"", map[string]string{}},
	}
	for _, tt := range tests {
		result := parseSearchString(tt.input)
		for k, v := range tt.expected {
			if result[k] != v {
				t.Errorf("parseSearchString(%q)[%q] = %q, want %q", tt.input, k, result[k], v)
			}
		}
	}
}

func TestConditionalCreateMiddleware_NoHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	handler := ConditionalCreateMiddleware(nil)(func(c echo.Context) error {
		called = true
		return c.String(http.StatusCreated, "created")
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("expected next handler to be called when no If-None-Exist header")
	}
}

func TestConditionalCreateMiddleware_NoMatch(t *testing.T) {
	searcher := func(c echo.Context, params map[string]string) (*ConditionalResult, error) {
		return &ConditionalResult{Count: 0}, nil
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("If-None-Exist", "identifier=12345")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	handler := ConditionalCreateMiddleware(searcher)(func(c echo.Context) error {
		called = true
		return c.String(http.StatusCreated, "created")
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("expected next handler to be called when 0 matches")
	}
}

func TestConditionalCreateMiddleware_OneMatch(t *testing.T) {
	searcher := func(c echo.Context, params map[string]string) (*ConditionalResult, error) {
		return &ConditionalResult{Count: 1, FHIRID: "existing-id"}, nil
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("If-None-Exist", "identifier=12345")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConditionalCreateMiddleware(searcher)(func(c echo.Context) error {
		t.Error("next handler should not be called when 1 match exists")
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestConditionalCreateMiddleware_MultipleMatches(t *testing.T) {
	searcher := func(c echo.Context, params map[string]string) (*ConditionalResult, error) {
		return &ConditionalResult{Count: 3}, nil
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("If-None-Exist", "identifier=12345")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConditionalCreateMiddleware(searcher)(func(c echo.Context) error {
		t.Error("next handler should not be called when multiple matches")
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusPreconditionFailed {
		t.Errorf("expected 412, got %d", rec.Code)
	}
}
