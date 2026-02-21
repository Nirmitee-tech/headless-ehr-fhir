package middleware

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// BodyLimit returns middleware that limits the maximum request body size.
// defaultLimit applies to most endpoints while bundleLimit applies to
// POST /fhir (FHIR transaction/batch bundles can be significantly larger).
//
// Limits are specified as human-readable strings: "1M" for 1 megabyte,
// "10M" for 10 megabytes, etc. Supported suffixes are K (kilobytes),
// M (megabytes), and G (gigabytes). A bare number is treated as bytes.
//
// When the limit is exceeded, the middleware returns HTTP 413 with a FHIR
// OperationOutcome body.
func BodyLimit(defaultLimit string, bundleLimit string) echo.MiddlewareFunc {
	defaultBytes := parseLimit(defaultLimit)
	bundleBytes := parseLimit(bundleLimit)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Body == nil || c.Request().Body == http.NoBody {
				return next(c)
			}

			// Determine which limit to apply based on the request path
			// and method. FHIR transaction/batch bundles are POSTed to
			// the root /fhir endpoint.
			limit := defaultBytes
			path := c.Request().URL.Path
			method := c.Request().Method
			if method == http.MethodPost && (path == "/fhir" || path == "/fhir/") {
				limit = bundleBytes
			}

			// Check Content-Length header first for early rejection
			if c.Request().ContentLength > limit {
				return payloadTooLargeError(c, limit)
			}

			// Wrap the body with a limiting reader to enforce the limit
			// even when Content-Length is missing or incorrect.
			c.Request().Body = &limitedReadCloser{
				ReadCloser: c.Request().Body,
				remaining:  limit,
				limit:      limit,
				c:          c,
			}

			return next(c)
		}
	}
}

// limitedReadCloser wraps an io.ReadCloser and returns an error once the
// read limit is exceeded.
type limitedReadCloser struct {
	io.ReadCloser
	remaining int64
	limit     int64
	exceeded  bool
	c         echo.Context
}

func (r *limitedReadCloser) Read(p []byte) (n int, err error) {
	if r.exceeded {
		return 0, echo.NewHTTPError(http.StatusRequestEntityTooLarge, "request body too large")
	}

	// Only read up to the remaining allowed bytes + 1 (to detect overflow)
	toRead := int64(len(p))
	if toRead > r.remaining+1 {
		toRead = r.remaining + 1
	}

	n, err = r.ReadCloser.Read(p[:toRead])
	r.remaining -= int64(n)

	if r.remaining < 0 {
		r.exceeded = true
		return 0, echo.NewHTTPError(http.StatusRequestEntityTooLarge, "request body too large")
	}

	return n, err
}

// payloadTooLargeError returns a 413 response with a FHIR OperationOutcome.
func payloadTooLargeError(c echo.Context, limit int64) error {
	outcome := map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue": []map[string]interface{}{
			{
				"severity":    "error",
				"code":        "too-costly",
				"diagnostics": fmt.Sprintf("Request body exceeds maximum allowed size of %d bytes", limit),
			},
		},
	}
	return c.JSON(http.StatusRequestEntityTooLarge, outcome)
}

// parseLimit parses a human-readable size string (e.g. "1M", "512K", "10G")
// into the number of bytes. If the string cannot be parsed, it defaults to
// 1 MB.
func parseLimit(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 1 << 20 // 1 MB default
	}

	s = strings.ToUpper(s)
	var multiplier int64 = 1

	if strings.HasSuffix(s, "G") || strings.HasSuffix(s, "GB") {
		multiplier = 1 << 30
		s = strings.TrimRight(s, "GB")
	} else if strings.HasSuffix(s, "M") || strings.HasSuffix(s, "MB") {
		multiplier = 1 << 20
		s = strings.TrimRight(s, "MB")
	} else if strings.HasSuffix(s, "K") || strings.HasSuffix(s, "KB") {
		multiplier = 1 << 10
		s = strings.TrimRight(s, "KB")
	}

	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 1 << 20 // 1 MB default on parse failure
	}

	return n * multiplier
}
