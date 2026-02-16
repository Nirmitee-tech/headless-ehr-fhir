package fhir

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// PreferMiddleware handles the FHIR Prefer header for write operations.
// Supports: return=minimal, return=representation, return=OperationOutcome.
// The default behavior (no Prefer header) returns the full representation.
func PreferMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			method := c.Request().Method
			if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch {
				return next(c)
			}

			prefer := c.Request().Header.Get("Prefer")
			if prefer == "" {
				return next(c)
			}

			returnPref := parsePreferReturn(prefer)
			if returnPref == "" || returnPref == "representation" {
				return next(c)
			}

			// Capture the response for post-processing
			origWriter := c.Response().Writer
			rec := &preferRecorder{
				ResponseWriter: origWriter,
				body:           &bytes.Buffer{},
				statusCode:     http.StatusOK,
			}
			c.Response().Writer = rec

			if err := next(c); err != nil {
				c.Response().Writer = origWriter
				return err
			}

			// Copy headers from recorded response
			for k, vals := range rec.Header() {
				for _, v := range vals {
					origWriter.Header().Set(k, v)
				}
			}

			switch returnPref {
			case "minimal":
				// Return empty body with original status code and headers
				origWriter.Header().Set("Content-Length", "0")
				origWriter.WriteHeader(rec.statusCode)
				return nil

			case "OperationOutcome":
				// Return a success OperationOutcome instead of the resource
				outcome := map[string]interface{}{
					"resourceType": "OperationOutcome",
					"issue": []interface{}{
						map[string]interface{}{
							"severity":    "information",
							"code":        "informational",
							"diagnostics": "Operation completed successfully",
						},
					},
				}
				data, err := json.Marshal(outcome)
				if err != nil {
					origWriter.WriteHeader(rec.statusCode)
					_, _ = origWriter.Write(rec.body.Bytes())
					return nil
				}
				origWriter.Header().Set(echo.HeaderContentType, "application/fhir+json; charset=utf-8")
				origWriter.WriteHeader(rec.statusCode)
				_, writeErr := origWriter.Write(data)
				return writeErr
			}

			// Fallback: write the original response
			origWriter.WriteHeader(rec.statusCode)
			_, err := origWriter.Write(rec.body.Bytes())
			return err
		}
	}
}

// parsePreferReturn extracts the return preference from a Prefer header value.
// Handles: "return=minimal", "return=minimal; handling=strict",
// "return=minimal, handling=strict", "handling=strict; return=minimal"
func parsePreferReturn(prefer string) string {
	// Split on both semicolons and commas to handle all formats
	for _, sep := range []string{",", ";"} {
		for _, part := range strings.Split(prefer, sep) {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "return=") {
				return strings.TrimSpace(part[7:])
			}
		}
	}
	return ""
}

// preferRecorder captures HTTP response for Prefer header post-processing.
type preferRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	wroteHead  bool
}

func (r *preferRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.wroteHead = true
}

func (r *preferRecorder) Write(b []byte) (int, error) {
	if !r.wroteHead {
		r.statusCode = http.StatusOK
		r.wroteHead = true
	}
	return r.body.Write(b)
}
