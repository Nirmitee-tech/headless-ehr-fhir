package fhir

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestContentNegotiation_DefaultContentType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ContentNegotiationMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != FHIRContentType {
		t.Errorf("expected Content-Type %q, got %q", FHIRContentType, ct)
	}
}

func TestContentNegotiation_FormatJSON(t *testing.T) {
	formats := []string{"json", "application/json", "application/fhir+json"}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_format="+format, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := ContentNegotiationMiddleware()(func(c echo.Context) error {
				return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
			})

			if err := handler(c); err != nil {
				t.Fatal(err)
			}
			if rec.Code != http.StatusOK {
				t.Errorf("expected 200, got %d", rec.Code)
			}
			ct := rec.Header().Get("Content-Type")
			if ct != FHIRContentType {
				t.Errorf("expected Content-Type %q, got %q", FHIRContentType, ct)
			}
		})
	}
}

func TestContentNegotiation_FormatXMLReturns406(t *testing.T) {
	formats := []string{"xml", "application/xml", "application/fhir+xml"}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_format="+format, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := ContentNegotiationMiddleware()(func(c echo.Context) error {
				return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
			})

			if err := handler(c); err != nil {
				t.Fatal(err)
			}
			if rec.Code != http.StatusNotAcceptable {
				t.Errorf("expected 406, got %d", rec.Code)
			}
			body := rec.Body.String()
			if !strings.Contains(body, "OperationOutcome") {
				t.Errorf("expected OperationOutcome in body, got: %s", body)
			}
			if !strings.Contains(body, "XML format is not supported") {
				t.Errorf("expected XML not supported message, got: %s", body)
			}
		})
	}
}

func TestContentNegotiation_AcceptFHIRJSON(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Accept", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ContentNegotiationMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != FHIRContentType {
		t.Errorf("expected Content-Type %q, got %q", FHIRContentType, ct)
	}
}

func TestContentNegotiation_AcceptFHIRXMLReturns406(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Accept", "application/fhir+xml")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ContentNegotiationMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotAcceptable {
		t.Errorf("expected 406, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "OperationOutcome") {
		t.Errorf("expected OperationOutcome in body, got: %s", body)
	}
}

func TestContentNegotiation_FormatTakesPrecedenceOverAccept(t *testing.T) {
	e := echo.New()
	// _format=json should win even though Accept says XML
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_format=json", nil)
	req.Header.Set("Accept", "application/fhir+xml")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ContentNegotiationMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != FHIRContentType {
		t.Errorf("expected Content-Type %q, got %q", FHIRContentType, ct)
	}
}

func TestContentNegotiation_FormatXMLTakesPrecedenceOverAcceptJSON(t *testing.T) {
	e := echo.New()
	// _format=xml should win even though Accept says JSON
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_format=xml", nil)
	req.Header.Set("Accept", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ContentNegotiationMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotAcceptable {
		t.Errorf("expected 406, got %d", rec.Code)
	}
}

func TestContentNegotiation_NoFormatNoAccept(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ContentNegotiationMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != FHIRContentType {
		t.Errorf("expected Content-Type %q, got %q", FHIRContentType, ct)
	}
}

func TestContentNegotiation_AcceptWildcard(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Accept", "*/*")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ContentNegotiationMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != FHIRContentType {
		t.Errorf("expected Content-Type %q, got %q", FHIRContentType, ct)
	}
}
