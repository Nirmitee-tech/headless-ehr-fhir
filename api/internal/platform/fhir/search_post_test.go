package fhir

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestMergeSearchParams(t *testing.T) {
	tests := []struct {
		name     string
		queryStr string
		formBody url.Values
		wantKeys map[string][]string
	}{
		{
			name:     "empty query and empty form body",
			queryStr: "",
			formBody: url.Values{},
			wantKeys: map[string][]string{},
		},
		{
			name:     "empty query with form body params",
			queryStr: "",
			formBody: url.Values{
				"status": {"active"},
				"code":   {"1234"},
			},
			wantKeys: map[string][]string{
				"status": {"active"},
				"code":   {"1234"},
			},
		},
		{
			name:     "query params with empty form body",
			queryStr: "status=active&code=1234",
			formBody: url.Values{},
			wantKeys: map[string][]string{
				"status": {"active"},
				"code":   {"1234"},
			},
		},
		{
			name:     "overlapping keys - query takes precedence",
			queryStr: "status=active",
			formBody: url.Values{
				"status": {"inactive"},
			},
			wantKeys: map[string][]string{
				"status": {"active"},
			},
		},
		{
			name:     "non-overlapping merge",
			queryStr: "status=active",
			formBody: url.Values{
				"code": {"1234"},
			},
			wantKeys: map[string][]string{
				"status": {"active"},
				"code":   {"1234"},
			},
		},
		{
			name:     "query has multiple values for same key",
			queryStr: "status=active&status=inactive",
			formBody: url.Values{
				"status": {"draft"},
			},
			wantKeys: map[string][]string{
				"status": {"active", "inactive"},
			},
		},
		{
			name:     "form body has multiple values for same key not in query",
			queryStr: "",
			formBody: url.Values{
				"status": {"active", "inactive"},
			},
			wantKeys: map[string][]string{
				"status": {"active", "inactive"},
			},
		},
		{
			name:     "nil form body",
			queryStr: "name=John",
			formBody: nil,
			wantKeys: map[string][]string{
				"name": {"John"},
			},
		},
		{
			name:     "mixed overlapping and non-overlapping",
			queryStr: "patient=123&_count=10",
			formBody: url.Values{
				"patient": {"456"},
				"code":    {"abc"},
			},
			wantKeys: map[string][]string{
				"patient": {"123"},
				"_count":  {"10"},
				"code":    {"abc"},
			},
		},
		{
			name:     "invalid query string is ignored gracefully",
			queryStr: "%;invalid",
			formBody: url.Values{
				"code": {"123"},
			},
			wantKeys: map[string][]string{
				"code": {"123"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeSearchParams(tt.queryStr, tt.formBody)

			// Check that all expected keys exist with correct values
			for k, wantVals := range tt.wantKeys {
				gotVals, ok := got[k]
				if !ok {
					t.Errorf("missing key %q in result", k)
					continue
				}
				if len(gotVals) != len(wantVals) {
					t.Errorf("key %q: got %d values, want %d (got=%v, want=%v)", k, len(gotVals), len(wantVals), gotVals, wantVals)
					continue
				}
				for i, v := range wantVals {
					if gotVals[i] != v {
						t.Errorf("key %q value[%d] = %q, want %q", k, i, gotVals[i], v)
					}
				}
			}

			// Check no extra keys
			for k := range got {
				if _, ok := tt.wantKeys[k]; !ok {
					t.Errorf("unexpected key %q in result with values %v", k, got[k])
				}
			}
		})
	}
}

func TestSearchPostMiddleware_FormPost(t *testing.T) {
	e := echo.New()
	mw := SearchPostMiddleware()

	var capturedQuery string
	handler := func(c echo.Context) error {
		capturedQuery = c.Request().URL.RawQuery
		return c.String(http.StatusOK, "ok")
	}

	e.POST("/Patient/_search", handler, mw)

	body := strings.NewReader("status=active&code=1234")
	req := httptest.NewRequest(http.MethodPost, "/Patient/_search", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	// The query params should have been merged from form body
	if capturedQuery == "" {
		t.Error("expected query params to be populated from form body")
	}
	if !strings.Contains(capturedQuery, "status=active") {
		t.Errorf("expected query to contain status=active, got %q", capturedQuery)
	}
	if !strings.Contains(capturedQuery, "code=1234") {
		t.Errorf("expected query to contain code=1234, got %q", capturedQuery)
	}
}

func TestSearchPostMiddleware_NonFormPost(t *testing.T) {
	e := echo.New()
	mw := SearchPostMiddleware()

	var capturedQuery string
	handler := func(c echo.Context) error {
		capturedQuery = c.Request().URL.RawQuery
		return c.String(http.StatusOK, "ok")
	}

	e.POST("/Patient", handler, mw)

	// JSON body, not form-encoded
	body := strings.NewReader(`{"resourceType":"Patient"}`)
	req := httptest.NewRequest(http.MethodPost, "/Patient", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	// Query should remain empty since content type is not form-encoded
	if capturedQuery != "" {
		t.Errorf("expected empty query for non-form POST, got %q", capturedQuery)
	}
}

func TestSearchPostMiddleware_GetPassthrough(t *testing.T) {
	e := echo.New()
	mw := SearchPostMiddleware()

	var capturedQuery string
	handler := func(c echo.Context) error {
		capturedQuery = c.Request().URL.RawQuery
		return c.String(http.StatusOK, "ok")
	}

	e.GET("/Patient", handler, mw)

	req := httptest.NewRequest(http.MethodGet, "/Patient?name=Smith", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if capturedQuery != "name=Smith" {
		t.Errorf("expected query 'name=Smith', got %q", capturedQuery)
	}
}

func TestSearchPostMiddleware_MergeWithExistingQuery(t *testing.T) {
	e := echo.New()
	mw := SearchPostMiddleware()

	var capturedQuery string
	handler := func(c echo.Context) error {
		capturedQuery = c.Request().URL.RawQuery
		return c.String(http.StatusOK, "ok")
	}

	e.POST("/Patient/_search", handler, mw)

	body := strings.NewReader("code=1234")
	req := httptest.NewRequest(http.MethodPost, "/Patient/_search?name=Smith", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	// Both query param and form param should be present
	if !strings.Contains(capturedQuery, "name=Smith") {
		t.Errorf("expected query to contain name=Smith, got %q", capturedQuery)
	}
	if !strings.Contains(capturedQuery, "code=1234") {
		t.Errorf("expected query to contain code=1234, got %q", capturedQuery)
	}
}
