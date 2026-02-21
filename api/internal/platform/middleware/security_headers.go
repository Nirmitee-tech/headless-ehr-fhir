package middleware

import (
	"github.com/labstack/echo/v4"
)

// SecurityHeaders returns middleware that sets security response headers on
// every request. These headers protect against common web vulnerabilities
// and enforce strict transport security for an API that handles PHI.
func SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			h := c.Response().Header()

			// Prevent MIME type sniffing
			h.Set("X-Content-Type-Options", "nosniff")

			// Prevent clickjacking
			h.Set("X-Frame-Options", "DENY")

			// Disable browser XSS filter — modern best practice is to rely
			// on Content-Security-Policy instead of the legacy filter.
			h.Set("X-XSS-Protection", "0")

			// Strict CSP for a JSON API: deny all resource loading and
			// frame embedding.
			h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

			// HTTP Strict Transport Security — 1 year including subdomains.
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

			// Do not send Referer header to downstream services.
			h.Set("Referrer-Policy", "no-referrer")

			// Disable browser features that an API does not need.
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

			// Prevent caching of API responses that may contain PHI.
			h.Set("Cache-Control", "no-store")

			return next(c)
		}
	}
}
