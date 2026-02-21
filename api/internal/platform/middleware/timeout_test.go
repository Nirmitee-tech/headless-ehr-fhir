package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func TestRequestTimeout_CompletesWithinDeadline(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	mw := RequestTimeout(5 * time.Second)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestRequestTimeout_ReturnsTimeoutOnExpiry(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		// Simulate a long-running handler
		select {
		case <-time.After(5 * time.Second):
			return c.String(http.StatusOK, "ok")
		case <-c.Request().Context().Done():
			return c.Request().Context().Err()
		}
	}

	mw := RequestTimeout(50 * time.Millisecond)
	h := mw(handler)
	err := h(c)

	// The middleware should have returned a 504 JSON response directly.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusGatewayTimeout {
		t.Errorf("expected status 504, got %d", rec.Code)
	}

	// Verify FHIR OperationOutcome body
	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if outcome["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", outcome["resourceType"])
	}
}

func TestRequestTimeout_SkipsWebSocketPaths(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ws/notifications", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	handler := func(c echo.Context) error {
		called = true
		// Verify there is no short deadline on the context
		deadline, ok := c.Request().Context().Deadline()
		if ok && time.Until(deadline) < 1*time.Second {
			t.Error("expected no short deadline for WebSocket path")
		}
		return c.String(http.StatusOK, "ws ok")
	}

	mw := RequestTimeout(50 * time.Millisecond)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called for WebSocket path")
	}
}

func TestRequestTimeout_ContextHasDeadline(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		_, ok := c.Request().Context().Deadline()
		if !ok {
			t.Error("expected context to have a deadline")
		}
		return c.String(http.StatusOK, "ok")
	}

	mw := RequestTimeout(30 * time.Second)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequestTimeout_PropagatesHandlerError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}

	mw := RequestTimeout(5 * time.Second)
	h := mw(handler)
	err := h(c)

	if err == nil {
		t.Fatal("expected error from handler")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", httpErr.Code)
	}
}
