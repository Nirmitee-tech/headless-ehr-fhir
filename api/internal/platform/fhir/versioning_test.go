package fhir

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestParseETag(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{`W/"3"`, 3, false},
		{`"5"`, 5, false},
		{`W/"1"`, 1, false},
		{`"abc"`, 0, true},
		{`W/""`, 0, true},
		{`42`, 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseETag(tt.input)
			if tt.wantErr && err == nil {
				t.Errorf("ParseETag(%q) should have returned error", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ParseETag(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseETag(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatETag(t *testing.T) {
	tests := []struct {
		version int
		want    string
	}{
		{1, `W/"1"`},
		{42, `W/"42"`},
		{0, `W/"0"`},
	}

	for _, tt := range tests {
		got := FormatETag(tt.version)
		if got != tt.want {
			t.Errorf("FormatETag(%d) = %q, want %q", tt.version, got, tt.want)
		}
	}
}

func TestParseETagRoundTrip(t *testing.T) {
	for _, v := range []int{1, 5, 42, 100} {
		etag := FormatETag(v)
		parsed, err := ParseETag(etag)
		if err != nil {
			t.Errorf("round-trip failed for %d: %v", v, err)
		}
		if parsed != v {
			t.Errorf("round-trip for %d: got %d", v, parsed)
		}
	}
}

func TestSetVersionHeaders_WithLastModified(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	SetVersionHeaders(c, 5, "2024-01-15T10:30:00Z")

	etag := rec.Header().Get("ETag")
	if etag != `W/"5"` {
		t.Errorf("expected ETag W/\"5\", got %q", etag)
	}
	lm := rec.Header().Get("Last-Modified")
	if lm != "2024-01-15T10:30:00Z" {
		t.Errorf("expected Last-Modified '2024-01-15T10:30:00Z', got %q", lm)
	}
}

func TestSetVersionHeaders_WithoutLastModified(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	SetVersionHeaders(c, 3, "")

	etag := rec.Header().Get("ETag")
	if etag != `W/"3"` {
		t.Errorf("expected ETag W/\"3\", got %q", etag)
	}
	lm := rec.Header().Get("Last-Modified")
	if lm != "" {
		t.Errorf("expected empty Last-Modified, got %q", lm)
	}
}

func TestCheckIfMatch_NoHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	version, err := CheckIfMatch(c, 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if version != 0 {
		t.Errorf("expected version 0 (unconditional), got %d", version)
	}
}

func TestCheckIfMatch_MatchingVersion(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	req.Header.Set("If-Match", `W/"5"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	version, err := CheckIfMatch(c, 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if version != 5 {
		t.Errorf("expected version 5, got %d", version)
	}
}

func TestCheckIfMatch_VersionMismatch(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	req.Header.Set("If-Match", `W/"3"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, err := CheckIfMatch(c, 5)
	if err == nil {
		t.Fatal("expected error for version mismatch")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", he.Code)
	}
}

func TestCheckIfMatch_InvalidETag(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	req.Header.Set("If-Match", `W/"abc"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, err := CheckIfMatch(c, 5)
	if err == nil {
		t.Fatal("expected error for invalid ETag")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", he.Code)
	}
}

func TestCheckIfNoneMatch_NoHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	match := CheckIfNoneMatch(c, 5)
	if match {
		t.Error("expected false when no If-None-Match header")
	}
}

func TestCheckIfNoneMatch_MatchingVersion(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-None-Match", `W/"5"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	match := CheckIfNoneMatch(c, 5)
	if !match {
		t.Error("expected true when version matches")
	}
}

func TestCheckIfNoneMatch_NonMatchingVersion(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-None-Match", `W/"3"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	match := CheckIfNoneMatch(c, 5)
	if match {
		t.Error("expected false when version does not match")
	}
}

func TestCheckIfNoneMatch_InvalidETag(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-None-Match", `W/"notanumber"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	match := CheckIfNoneMatch(c, 5)
	if match {
		t.Error("expected false when ETag is invalid")
	}
}

func TestParseETag_Whitespace(t *testing.T) {
	// ParseETag should trim whitespace
	v, err := ParseETag("  W/\"10\"  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 10 {
		t.Errorf("expected 10, got %d", v)
	}
}

func TestFormatETag_LargeVersion(t *testing.T) {
	got := FormatETag(999999)
	want := `W/"999999"`
	if got != want {
		t.Errorf("FormatETag(999999) = %q, want %q", got, want)
	}
}

func TestSetVersionHeaders_ETagFormat(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	SetVersionHeaders(c, 42, "")

	etag := rec.Header().Get("ETag")
	if etag != `W/"42"` {
		t.Errorf("expected ETag W/\"42\", got %q", etag)
	}
}

func TestCheckIfMatch_ConflictMessage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	req.Header.Set("If-Match", `W/"1"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, err := CheckIfMatch(c, 5)
	if err == nil {
		t.Fatal("expected error")
	}
	he := err.(*echo.HTTPError)
	msg, ok := he.Message.(string)
	if !ok {
		t.Fatal("expected string message")
	}
	if !strings.Contains(msg, "version 1") || !strings.Contains(msg, "version 5") {
		t.Errorf("error message should mention both versions, got %q", msg)
	}
}

func TestCheckIfMatch_BadRequestMessage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	req.Header.Set("If-Match", `W/"notanum"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, err := CheckIfMatch(c, 5)
	if err == nil {
		t.Fatal("expected error")
	}
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", he.Code)
	}
}

func TestParseETag_JustNumber(t *testing.T) {
	v, err := ParseETag("7")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 7 {
		t.Errorf("expected 7, got %d", v)
	}
}

func TestParseETag_NegativeNumber(t *testing.T) {
	// Negative numbers are technically valid integers
	v, err := ParseETag("-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != -1 {
		t.Errorf("expected -1, got %d", v)
	}
}
