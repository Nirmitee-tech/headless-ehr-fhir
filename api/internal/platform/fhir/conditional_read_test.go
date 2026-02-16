package fhir

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestConditionalReadMiddleware_IfNoneMatchMatching(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/1", nil)
	req.Header.Set("If-None-Match", `W/"5"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConditionalReadMiddleware()(func(c echo.Context) error {
		c.Response().Header().Set("ETag", `W/"5"`)
		c.Response().Header().Set("Last-Modified", "2024-01-15T10:30:00Z")
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusNotModified {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotModified)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body for 304, got: %s", rec.Body.String())
	}
	// ETag should still be present on 304 responses.
	if etag := rec.Header().Get("ETag"); etag != `W/"5"` {
		t.Errorf("ETag = %q, want W/\"5\"", etag)
	}
}

func TestConditionalReadMiddleware_IfNoneMatchNotMatching(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/1", nil)
	req.Header.Set("If-None-Match", `W/"3"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConditionalReadMiddleware()(func(c echo.Context) error {
		c.Response().Header().Set("ETag", `W/"5"`)
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != `{"resourceType":"Patient","id":"1"}` {
		t.Errorf("body = %q, want full resource", rec.Body.String())
	}
}

func TestConditionalReadMiddleware_NoConditionalHeaders(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConditionalReadMiddleware()(func(c echo.Context) error {
		c.Response().Header().Set("ETag", `W/"5"`)
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != `{"resourceType":"Patient","id":"1"}` {
		t.Errorf("body = %q, want full resource", rec.Body.String())
	}
}

func TestConditionalReadMiddleware_PostNotAffected(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("If-None-Match", `W/"5"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConditionalReadMiddleware()(func(c echo.Context) error {
		c.Response().Header().Set("ETag", `W/"5"`)
		return c.String(http.StatusCreated, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	// POST should pass through regardless of If-None-Match.
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if rec.Body.String() != `{"resourceType":"Patient","id":"1"}` {
		t.Errorf("body = %q, want full resource", rec.Body.String())
	}
}

func TestConditionalReadMiddleware_Non200Passthrough(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/1", nil)
	req.Header.Set("If-None-Match", `W/"5"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConditionalReadMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusNotFound, `{"resourceType":"OperationOutcome"}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	// 404 should pass through unchanged.
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if rec.Body.String() != `{"resourceType":"OperationOutcome"}` {
		t.Errorf("body = %q, want OperationOutcome", rec.Body.String())
	}
}

func TestConditionalReadMiddleware_IfModifiedSinceNotModified(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/1", nil)
	req.Header.Set("If-Modified-Since", "2024-06-01T00:00:00Z")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConditionalReadMiddleware()(func(c echo.Context) error {
		c.Response().Header().Set("ETag", `W/"5"`)
		c.Response().Header().Set("Last-Modified", "2024-01-15T10:30:00Z")
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	// Resource was last modified on 2024-01-15, client has data from 2024-06-01,
	// so the resource has NOT been modified since then -> 304.
	if rec.Code != http.StatusNotModified {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotModified)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body for 304, got: %s", rec.Body.String())
	}
}

func TestConditionalReadMiddleware_IfModifiedSinceModified(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/1", nil)
	req.Header.Set("If-Modified-Since", "2024-01-01T00:00:00Z")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConditionalReadMiddleware()(func(c echo.Context) error {
		c.Response().Header().Set("ETag", `W/"5"`)
		c.Response().Header().Set("Last-Modified", "2024-01-15T10:30:00Z")
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	// Resource was last modified on 2024-01-15, client has data from 2024-01-01,
	// so the resource HAS been modified since then -> 200 with full body.
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != `{"resourceType":"Patient","id":"1"}` {
		t.Errorf("body = %q, want full resource", rec.Body.String())
	}
}

func TestConditionalReadMiddleware_SearchBundlePassthrough(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("If-None-Match", `W/"5"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	bundleJSON := `{"resourceType":"Bundle","type":"searchset","total":1}`

	handler := ConditionalReadMiddleware()(func(c echo.Context) error {
		c.Response().Header().Set("ETag", `W/"5"`)
		return c.String(http.StatusOK, bundleJSON)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	// Search bundles should pass through, not be subject to conditional read.
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != bundleJSON {
		t.Errorf("body = %q, want bundle JSON", rec.Body.String())
	}
}

func TestConditionalReadMiddleware_HandlerError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/1", nil)
	req.Header.Set("If-None-Match", `W/"5"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConditionalReadMiddleware()(func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	})

	err := handler(c)
	if err == nil {
		t.Fatal("expected error from handler")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusInternalServerError {
		t.Errorf("error code = %d, want %d", httpErr.Code, http.StatusInternalServerError)
	}
}

func TestConditionalReadMiddleware_WildcardIfNoneMatch(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/1", nil)
	req.Header.Set("If-None-Match", "*")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConditionalReadMiddleware()(func(c echo.Context) error {
		c.Response().Header().Set("ETag", `W/"5"`)
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	// Wildcard * should match any ETag.
	if rec.Code != http.StatusNotModified {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotModified)
	}
}

func TestConditionalReadMiddleware_NoETagInResponse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/1", nil)
	req.Header.Set("If-None-Match", `W/"5"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConditionalReadMiddleware()(func(c echo.Context) error {
		// Handler does not set ETag header.
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	// Without an ETag in the response, conditional read cannot match.
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != `{"resourceType":"Patient","id":"1"}` {
		t.Errorf("body = %q, want full resource", rec.Body.String())
	}
}

func TestEtagsMatch(t *testing.T) {
	tests := []struct {
		name          string
		ifNoneMatch   string
		responseETag  string
		expectMatch   bool
	}{
		{"exact match", `W/"5"`, `W/"5"`, true},
		{"version mismatch", `W/"3"`, `W/"5"`, false},
		{"strong vs weak", `"5"`, `W/"5"`, true},
		{"wildcard", "*", `W/"5"`, true},
		{"invalid client etag", "invalid", `W/"5"`, false},
		{"invalid server etag", `W/"5"`, "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := etagsMatch(tt.ifNoneMatch, tt.responseETag)
			if got != tt.expectMatch {
				t.Errorf("etagsMatch(%q, %q) = %v, want %v", tt.ifNoneMatch, tt.responseETag, got, tt.expectMatch)
			}
		})
	}
}

func TestModifiedSince(t *testing.T) {
	tests := []struct {
		name            string
		lastModified    string
		ifModifiedSince string
		expectModified  bool
	}{
		{
			"modified after",
			"2024-06-15T10:30:00Z",
			"2024-01-01T00:00:00Z",
			true,
		},
		{
			"not modified",
			"2024-01-15T10:30:00Z",
			"2024-06-01T00:00:00Z",
			false,
		},
		{
			"same time",
			"2024-01-15T10:30:00Z",
			"2024-01-15T10:30:00Z",
			false,
		},
		{
			"invalid lastModified",
			"not-a-date",
			"2024-01-01T00:00:00Z",
			true, // Parse failure defaults to modified.
		},
		{
			"invalid ifModifiedSince",
			"2024-01-15T10:30:00Z",
			"not-a-date",
			true, // Parse failure defaults to modified.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := modifiedSince(tt.lastModified, tt.ifModifiedSince)
			if got != tt.expectModified {
				t.Errorf("modifiedSince(%q, %q) = %v, want %v", tt.lastModified, tt.ifModifiedSince, got, tt.expectModified)
			}
		})
	}
}

func TestIsSearchBundle(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		expect bool
	}{
		{"searchset bundle", `{"resourceType":"Bundle","type":"searchset","total":1}`, true},
		{"single resource", `{"resourceType":"Patient","id":"1"}`, false},
		{"empty body", "", false},
		{"transaction bundle", `{"resourceType":"Bundle","type":"transaction"}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSearchBundle([]byte(tt.body))
			if got != tt.expect {
				t.Errorf("isSearchBundle(%q) = %v, want %v", tt.body, got, tt.expect)
			}
		})
	}
}
