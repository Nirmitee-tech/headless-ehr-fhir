package middleware

import (
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// maxHeaderValueSize is the maximum allowed size for any single header value.
const maxHeaderValueSize = 8192 // 8KB

// Compiled patterns for injection detection.
var (
	// SQL injection patterns (defense-in-depth warning only).
	sqlPatterns = regexp.MustCompile(`(?i)('+\s*;\s*DROP\b|UNION\s+SELECT\b|'\s+OR\s+1\s*=\s*1|1\s*=\s*1)`)

	// Script injection patterns (block).
	scriptPatterns = regexp.MustCompile(`(?i)(<script|javascript\s*:|on\w+\s*=)`)
)

// Sanitize returns middleware that validates and sanitizes incoming requests.
// It checks for common attack patterns in headers, query parameters, and path
// parameters. Blocked requests receive a 400 Bad Request with a FHIR
// OperationOutcome body.
func Sanitize() echo.MiddlewareFunc {
	return SanitizeWithLogger(zerolog.Nop())
}

// SanitizeWithLogger returns the sanitize middleware configured with a logger
// for defense-in-depth SQL injection warnings.
func SanitizeWithLogger(logger zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			path := req.URL.Path
			rawPath := req.URL.RawPath
			if rawPath == "" {
				rawPath = path
			}

			// 1. Path traversal prevention
			if containsPathTraversal(path) || containsPathTraversal(rawPath) {
				return operationOutcome(c, "Path traversal detected")
			}

			// 2. Null byte injection in path
			if containsNullByte(path) || containsNullByte(rawPath) {
				return operationOutcome(c, "Null byte injection detected")
			}

			// 3. Header injection and oversized headers
			for name, values := range req.Header {
				for _, v := range values {
					// Oversized header check
					if len(v) > maxHeaderValueSize {
						return operationOutcome(c, "Header value exceeds maximum size: "+name)
					}
					// Newline injection check
					if strings.ContainsAny(v, "\r\n") {
						return operationOutcome(c, "Header injection detected: "+name)
					}
				}
			}

			// 4. Query parameter checks
			for key, values := range req.URL.Query() {
				for _, v := range values {
					// Null byte in query param
					if containsNullByte(v) || containsNullByte(key) {
						return operationOutcome(c, "Null byte injection detected in query parameter")
					}

					// SQL injection warning (defense-in-depth logging, not blocking)
					if sqlPatterns.MatchString(v) {
						logger.Warn().
							Str("param", key).
							Str("path", path).
							Str("remote_ip", c.RealIP()).
							Msg("potential SQL injection pattern detected in query parameter")
					}

					// Script injection (block)
					if scriptPatterns.MatchString(v) || scriptPatterns.MatchString(key) {
						return operationOutcome(c, "Script injection detected in query parameter")
					}
				}
			}

			return next(c)
		}
	}
}

// containsPathTraversal checks for path traversal sequences in raw and
// percent-encoded forms.
func containsPathTraversal(s string) bool {
	if strings.Contains(s, "..") {
		return true
	}
	lower := strings.ToLower(s)
	if strings.Contains(lower, "%2e%2e") {
		return true
	}
	if strings.Contains(lower, "%252e") {
		return true
	}
	return false
}

// containsNullByte checks for null bytes in raw and percent-encoded forms.
func containsNullByte(s string) bool {
	if strings.ContainsRune(s, '\x00') {
		return true
	}
	lower := strings.ToLower(s)
	if strings.Contains(lower, "%00") {
		return true
	}
	return false
}

// operationOutcome returns a 400 Bad Request with a FHIR OperationOutcome.
func operationOutcome(c echo.Context, diagnostics string) error {
	return c.JSON(http.StatusBadRequest, map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue": []map[string]interface{}{
			{
				"severity":    "error",
				"code":        "invalid",
				"diagnostics": diagnostics,
			},
		},
	})
}

// SanitizeString removes or escapes potentially dangerous characters from a
// string value. It strips null bytes and control characters (except \n, \r, \t)
// and trims excessive whitespace. Handlers can use this for additional
// field-level sanitization.
func SanitizeString(input string) string {
	// Strip null bytes and control characters except \n, \r, \t
	var b strings.Builder
	b.Grow(len(input))
	for _, r := range input {
		if r == '\x00' {
			continue
		}
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			continue
		}
		b.WriteRune(r)
	}

	// Trim excessive leading/trailing whitespace
	return strings.TrimSpace(b.String())
}
