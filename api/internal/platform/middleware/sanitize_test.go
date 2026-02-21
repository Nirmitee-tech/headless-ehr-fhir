package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// sanitizeOKHandler is a simple handler that returns 200 OK for pass-through tests.
func sanitizeOKHandler(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

func newSanitizeEcho() *echo.Echo {
	e := echo.New()
	logger := zerolog.New(os.Stderr).With().Logger()
	e.Use(SanitizeWithLogger(logger))
	e.GET("/*", sanitizeOKHandler)
	e.POST("/*", sanitizeOKHandler)
	return e
}

// ---------------------------------------------------------------------------
// Path traversal tests
// ---------------------------------------------------------------------------

func TestSanitize_PathTraversal_DotDot(t *testing.T) {
	e := newSanitizeEcho()

	req := httptest.NewRequest(http.MethodGet, "/../../etc/passwd", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	assertOperationOutcome(t, rec)
}

func TestSanitize_PathTraversal_EncodedDotDot(t *testing.T) {
	e := newSanitizeEcho()

	// Use a raw URL with encoded path traversal
	req := httptest.NewRequest(http.MethodGet, "/%2e%2e/%2e%2e/etc/passwd", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	assertOperationOutcome(t, rec)
}

func TestSanitize_PathTraversal_DoubleEncoded(t *testing.T) {
	e := newSanitizeEcho()

	req := httptest.NewRequest(http.MethodGet, "/%252e%252e/etc/passwd", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	assertOperationOutcome(t, rec)
}

// ---------------------------------------------------------------------------
// Null byte injection tests
// ---------------------------------------------------------------------------

func TestSanitize_NullByte_InPath(t *testing.T) {
	e := newSanitizeEcho()

	req := httptest.NewRequest(http.MethodGet, "/file%00.txt", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	assertOperationOutcome(t, rec)
}

func TestSanitize_NullByte_InQueryParam(t *testing.T) {
	e := newSanitizeEcho()

	req := httptest.NewRequest(http.MethodGet, "/test?name=foo%00bar", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	assertOperationOutcome(t, rec)
}

// ---------------------------------------------------------------------------
// Header injection tests
// ---------------------------------------------------------------------------

func TestSanitize_HeaderInjection_Newline(t *testing.T) {
	e := newSanitizeEcho()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Custom", "value\r\nInjected: header")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	assertOperationOutcome(t, rec)
}

func TestSanitize_HeaderInjection_CR(t *testing.T) {
	e := newSanitizeEcho()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Custom", "value\rinjected")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestSanitize_HeaderInjection_LF(t *testing.T) {
	e := newSanitizeEcho()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Custom", "value\ninjected")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestSanitize_OversizedHeader(t *testing.T) {
	e := newSanitizeEcho()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// Create a header value larger than 8KB
	bigValue := make([]byte, maxHeaderValueSize+1)
	for i := range bigValue {
		bigValue[i] = 'A'
	}
	req.Header.Set("X-Big", string(bigValue))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	assertOperationOutcome(t, rec)
}

// ---------------------------------------------------------------------------
// Normal requests pass through
// ---------------------------------------------------------------------------

func TestSanitize_NormalRequest_PassesThrough(t *testing.T) {
	e := newSanitizeEcho()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients?name=John", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

func TestSanitize_FHIRPath_Normal(t *testing.T) {
	e := newSanitizeEcho()

	paths := []string{
		"/fhir/Patient/123",
		"/fhir/Patient?name=John&birthdate=1990-01-01",
		"/fhir/Observation?code=http://loinc.org|1234-5",
		"/fhir/metadata",
		"/fhir/Patient/123/_history/2",
		"/fhir/Encounter?_include=Encounter:patient",
	}

	for _, p := range paths {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("path %s: expected 200, got %d", p, rec.Code)
		}
	}
}

// ---------------------------------------------------------------------------
// SQL injection warning (passes through, only logs)
// ---------------------------------------------------------------------------

func TestSanitize_SQLInjection_Warning_PassesThrough(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	e := echo.New()
	e.Use(SanitizeWithLogger(logger))
	e.GET("/*", sanitizeOKHandler)

	tests := []struct {
		name  string
		path  string
		param string
		value string
	}{
		{"drop", "/test", "name", "'; DROP TABLE patients;--"},
		{"union_select", "/test", "name", "1 UNION SELECT * FROM users"},
		{"or_1_1", "/test", "name", "' OR 1=1--"},
		{"1_eq_1", "/test", "id", "1=1"},
	}

	for _, tt := range tests {
		buf.Reset()
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		q := req.URL.Query()
		q.Set(tt.param, tt.value)
		req.URL.RawQuery = q.Encode()
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		// Should pass through (not blocked)
		if rec.Code != http.StatusOK {
			t.Errorf("%s: expected 200 (pass-through), got %d", tt.name, rec.Code)
		}

		// Should have logged a warning
		if !bytes.Contains(buf.Bytes(), []byte("potential SQL injection")) {
			t.Errorf("%s: expected SQL injection warning in logs", tt.name)
		}
	}
}

// ---------------------------------------------------------------------------
// Script injection tests (blocked)
// ---------------------------------------------------------------------------

func TestSanitize_ScriptInjection_Blocked(t *testing.T) {
	e := newSanitizeEcho()

	tests := []struct {
		name  string
		param string
		value string
	}{
		{"script_tag", "name", "<script>alert(1)</script>"},
		{"javascript_uri", "url", "javascript:alert(1)"},
		{"event_handler", "val", "onload=alert(1)"},
		{"onclick", "val", "onclick=alert(1)"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		q := req.URL.Query()
		q.Set(tt.param, tt.value)
		req.URL.RawQuery = q.Encode()
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", tt.name, rec.Code)
		}
		assertOperationOutcome(t, rec)
	}
}

// ---------------------------------------------------------------------------
// SanitizeString tests
// ---------------------------------------------------------------------------

func TestSanitizeString_RemovesNullBytes(t *testing.T) {
	input := "hello\x00world"
	result := SanitizeString(input)
	if result != "helloworld" {
		t.Errorf("expected 'helloworld', got %q", result)
	}
}

func TestSanitizeString_RemovesControlChars(t *testing.T) {
	// \x01 (SOH), \x07 (BEL), \x1B (ESC) should be stripped
	input := "hello\x01world\x07test\x1Bend"
	result := SanitizeString(input)
	if result != "helloworldtestend" {
		t.Errorf("expected 'helloworldtestend', got %q", result)
	}
}

func TestSanitizeString_PreservesNewlineTabCR(t *testing.T) {
	input := "line1\nline2\ttab\rreturn"
	result := SanitizeString(input)
	if result != "line1\nline2\ttab\rreturn" {
		t.Errorf("expected preserved whitespace chars, got %q", result)
	}
}

func TestSanitizeString_PreservesNormalText(t *testing.T) {
	input := "John Doe, M.D. (Cardiology) - Patient #12345"
	result := SanitizeString(input)
	if result != input {
		t.Errorf("expected unchanged text, got %q", result)
	}
}

func TestSanitizeString_TrimsWhitespace(t *testing.T) {
	input := "   hello world   "
	result := SanitizeString(input)
	if result != "hello world" {
		t.Errorf("expected 'hello world', got %q", result)
	}
}

func TestSanitizeString_EmptyString(t *testing.T) {
	result := SanitizeString("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestSanitizeString_OnlyNullBytes(t *testing.T) {
	result := SanitizeString("\x00\x00\x00")
	if result != "" {
		t.Errorf("expected empty string after stripping nulls, got %q", result)
	}
}

func TestSanitizeString_UnicodePreserved(t *testing.T) {
	input := "Jornada medica: examen de sangre"
	result := SanitizeString(input)
	if result != input {
		t.Errorf("expected unicode preserved, got %q", result)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func assertOperationOutcome(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if body["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", body["resourceType"])
	}
	issues, ok := body["issue"].([]interface{})
	if !ok || len(issues) == 0 {
		t.Error("expected at least one issue in OperationOutcome")
	}
}
