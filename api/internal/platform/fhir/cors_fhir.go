package fhir

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// FHIRCORSConfig holds CORS configuration tailored to the FHIR specification.
// FHIR servers must advertise support for specific headers used by FHIR clients
// (e.g., Prefer, If-Match, If-None-Exist) in addition to standard CORS headers.
type FHIRCORSConfig struct {
	AllowOrigins     []string
	AllowCredentials bool
	MaxAge           int // in seconds
}

// DefaultFHIRCORSConfig returns the recommended CORS defaults for a FHIR server.
// AllowOrigins is set to ["*"] (all origins), credentials are disabled (required
// when using the wildcard origin), and preflight results are cached for 1 hour.
func DefaultFHIRCORSConfig() FHIRCORSConfig {
	return FHIRCORSConfig{
		AllowOrigins:     []string{"*"},
		AllowCredentials: false,
		MaxAge:           3600,
	}
}

// fhirAllowMethods lists every HTTP method a FHIR server typically supports.
var fhirAllowMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodHead,
	http.MethodOptions,
}

// fhirAllowHeaders includes standard, FHIR-specific, and custom headers that
// clients may send when interacting with a FHIR API.
var fhirAllowHeaders = []string{
	// Standard headers
	"Content-Type",
	"Authorization",
	"Accept",
	"Cache-Control",
	// FHIR-specific headers
	"Prefer",
	"If-Match",
	"If-None-Match",
	"If-Modified-Since",
	"If-None-Exist",
	// Custom headers
	"X-Tenant-ID",
	"X-Request-ID",
	"X-Break-Glass",
	"X-Security-Labels",
	"X-Purpose-Of-Use",
}

// fhirExposeHeaders lists response headers that should be visible to
// browser-based FHIR clients.
var fhirExposeHeaders = []string{
	"ETag",
	"Last-Modified",
	"Location",
	"Content-Location",
	"X-Request-ID",
	"X-RateLimit-Limit",
	"X-RateLimit-Remaining",
	"X-RateLimit-Reset",
	"Retry-After",
}

// FHIRCORSMiddleware returns Echo middleware that sets CORS headers required by
// the FHIR specification. When called without arguments the DefaultFHIRCORSConfig
// is used. Preflight OPTIONS requests are handled by returning 204 No Content.
func FHIRCORSMiddleware(config ...FHIRCORSConfig) echo.MiddlewareFunc {
	cfg := DefaultFHIRCORSConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	allowMethods := strings.Join(fhirAllowMethods, ", ")
	allowHeaders := strings.Join(fhirAllowHeaders, ", ")
	exposeHeaders := strings.Join(fhirExposeHeaders, ", ")
	maxAge := strconv.Itoa(cfg.MaxAge)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			origin := c.Request().Header.Get("Origin")
			if origin == "" {
				// Not a CORS request; skip CORS headers entirely.
				return next(c)
			}

			h := c.Response().Header()

			// Determine the allowed origin value.
			allowOrigin := resolveAllowOrigin(cfg.AllowOrigins, origin)
			if allowOrigin == "" {
				// Origin not permitted; proceed without CORS headers.
				return next(c)
			}
			h.Set("Access-Control-Allow-Origin", allowOrigin)

			// When a specific origin is returned (not "*"), the Vary header
			// must include Origin so caches distinguish per-origin responses.
			if allowOrigin != "*" {
				h.Add("Vary", "Origin")
			}

			if cfg.AllowCredentials {
				h.Set("Access-Control-Allow-Credentials", "true")
			}

			// Always expose FHIR-relevant response headers.
			h.Set("Access-Control-Expose-Headers", exposeHeaders)

			// Handle preflight requests.
			if c.Request().Method == http.MethodOptions {
				h.Set("Access-Control-Allow-Methods", allowMethods)
				h.Set("Access-Control-Allow-Headers", allowHeaders)
				h.Set("Access-Control-Max-Age", maxAge)
				return c.NoContent(http.StatusNoContent)
			}

			return next(c)
		}
	}
}

// resolveAllowOrigin returns the value for the Access-Control-Allow-Origin
// header. If the configured origins contain "*", the wildcard is returned.
// Otherwise the request origin is returned only if it appears in the allowed
// list. An empty string signals that the origin is not permitted.
func resolveAllowOrigin(allowed []string, origin string) string {
	for _, o := range allowed {
		if o == "*" {
			return "*"
		}
		if o == origin {
			return origin
		}
	}
	return ""
}
