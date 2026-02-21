package auth

import (
	"github.com/labstack/echo/v4"
)

// publicPaths lists URL paths that should bypass authentication and tenant
// resolution. These are infrastructure endpoints (health checks, metrics)
// and FHIR discovery endpoints that must be accessible without credentials.
var publicPaths = map[string]bool{
	"/health":                                 true,
	"/health/db":                              true,
	"/metrics":                                true,
	"/.well-known/smart-configuration":        true,
	"/fhir/.well-known/smart-configuration":   true,
	"/fhir/metadata":                          true,
	"/auth/authorize":                         true,
	"/auth/token":                             true,
	"/auth/introspect":                        true,
	"/auth/register":                          true,
}

// AuthSkipper returns true for requests whose path should skip authentication.
// Pass this function as the Skipper on JWTConfig or DevAuthMiddleware so that
// health-check, metrics, and FHIR discovery endpoints remain accessible
// without a bearer token or tenant context.
func AuthSkipper(c echo.Context) bool {
	return publicPaths[c.Path()]
}

// IsPublicPath reports whether the given path is a public infrastructure
// endpoint that should bypass auth and tenant middleware.
func IsPublicPath(path string) bool {
	return publicPaths[path]
}
