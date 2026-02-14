package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func TestRequestID_GeneratesNew(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		rid := c.Get("request_id").(string)
		if rid == "" {
			t.Error("expected request_id to be generated")
		}
		return c.String(http.StatusOK, "ok")
	}

	mw := RequestID()
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Header().Get(RequestIDHeader) == "" {
		t.Error("expected X-Request-ID response header")
	}
}

func TestRequestID_PreservesExisting(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "my-custom-id")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		rid := c.Get("request_id").(string)
		if rid != "my-custom-id" {
			t.Errorf("expected my-custom-id, got %s", rid)
		}
		return c.String(http.StatusOK, "ok")
	}

	mw := RequestID()
	h := mw(handler)
	h(c)

	if rec.Header().Get(RequestIDHeader) != "my-custom-id" {
		t.Errorf("expected my-custom-id in response header, got %s", rec.Header().Get(RequestIDHeader))
	}
}

func TestLogger_LogsRequest(t *testing.T) {
	logger := zerolog.New(os.Stderr).With().Logger()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := Logger(logger)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecovery_CatchesPanic(t *testing.T) {
	logger := zerolog.New(os.Stderr).With().Logger()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		panic("test panic")
	}

	mw := Recovery(logger)
	h := mw(handler)
	err := h(c)

	if err == nil {
		t.Fatal("expected error from recovered panic")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", httpErr.Code)
	}
}

func TestRecovery_PassesThrough(t *testing.T) {
	logger := zerolog.New(os.Stderr).With().Logger()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := Recovery(logger)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAudit_LogsEvent(t *testing.T) {
	logger := zerolog.New(os.Stderr).With().Logger()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("tenant_id", "test-tenant")
	c.Set("request_id", "req-123")

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := Audit(logger)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
