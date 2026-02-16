package fhir

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// FHIRContentType is the FHIR JSON content type with charset.
const FHIRContentType = "application/fhir+json; charset=utf-8"

// ContentNegotiationMiddleware handles FHIR content negotiation per the FHIR
// specification. It checks the _format query parameter first (highest priority),
// then falls back to the Accept header. All successful responses are served as
// application/fhir+json. XML formats are rejected with 406 Not Acceptable.
func ContentNegotiationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// _format query parameter takes highest priority per FHIR spec.
			format := c.QueryParam("_format")
			if format != "" {
				if isXMLFormat(format) {
					return c.JSON(http.StatusNotAcceptable, ErrorOutcome("XML format is not supported. Use application/fhir+json."))
				}
				if isJSONFormat(format) {
					c.Response().Header().Set(echo.HeaderContentType, FHIRContentType)
					return next(c)
				}
				// Unknown format value: reject.
				return c.JSON(http.StatusNotAcceptable, ErrorOutcome("Unsupported _format value: "+format))
			}

			// Fall back to Accept header.
			accept := c.Request().Header.Get("Accept")
			if accept != "" {
				if negotiateAccept(accept) {
					c.Response().Header().Set(echo.HeaderContentType, FHIRContentType)
					return next(c)
				}
				// Accept header present but no acceptable type found.
				return c.JSON(http.StatusNotAcceptable, ErrorOutcome("Accept header does not include a supported FHIR content type. Use application/fhir+json."))
			}

			// No _format and no Accept header: default to FHIR JSON.
			c.Response().Header().Set(echo.HeaderContentType, FHIRContentType)
			return next(c)
		}
	}
}

// normalizeFormat normalises a format string by lowercasing, trimming
// whitespace, and restoring the "+" that HTTP query-string decoding may have
// converted to a space (e.g. "application/fhir json" -> "application/fhir+json").
func normalizeFormat(raw string) string {
	f := strings.TrimSpace(strings.ToLower(raw))
	f = strings.ReplaceAll(f, "fhir json", "fhir+json")
	f = strings.ReplaceAll(f, "fhir xml", "fhir+xml")
	return f
}

// isJSONFormat returns true if the format string represents a JSON content type.
func isJSONFormat(format string) bool {
	switch normalizeFormat(format) {
	case "json", "application/json", "application/fhir+json":
		return true
	}
	return false
}

// isXMLFormat returns true if the format string represents an XML content type.
func isXMLFormat(format string) bool {
	switch normalizeFormat(format) {
	case "xml", "application/xml", "application/fhir+xml":
		return true
	}
	return false
}

// negotiateAccept parses the Accept header and returns true if any of the
// listed media types are JSON-compatible. Returns false if only XML or
// unsupported types are present.
func negotiateAccept(accept string) bool {
	for _, part := range strings.Split(accept, ",") {
		// Strip quality parameters (e.g., ";q=0.9").
		mediaType := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		mediaType = strings.ToLower(mediaType)
		switch mediaType {
		case "application/fhir+json", "application/json", "json", "*/*":
			return true
		}
	}
	return false
}
