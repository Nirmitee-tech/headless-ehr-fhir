package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// FHIRScopeMiddleware is Echo middleware that enforces SMART on FHIR scopes
// on the FHIR route group. It extracts the resource type from the URL path
// and the operation from the HTTP method, then checks if the user's scopes
// (from JWT) grant access.
//
// Operations mapping:
//
//	GET/HEAD  -> "read"
//	POST      -> "write" (create) or "read" (search, when path ends with _search)
//	PUT/PATCH -> "write"
//	DELETE    -> "write"
//
// Scope format: "patient/<Resource>.<op>" or "user/<Resource>.<op>" or "system/<Resource>.<op>"
// Wildcards: "user/*.read" matches all resources for read, "user/*.*" matches everything.
//
// Bypass conditions:
//   - /fhir/metadata and /fhir/.well-known/* endpoints (always public)
//   - Users with the "admin" role
//   - Requests with no scopes in context (backward compatible with dev mode)
func FHIRScopeMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path

			// Skip scope enforcement for discovery/capability endpoints.
			if isScopeExemptPath(path) {
				return next(c)
			}

			ctx := c.Request().Context()

			// Admin bypass: users with the "admin" role skip scope checks.
			roles := RolesFromContext(ctx)
			for _, r := range roles {
				if r == "admin" {
					return next(c)
				}
			}

			// Get raw scopes from JWT context.
			rawScopes := ScopesFromContext(ctx)

			// Dev mode / backward compatibility: if no scopes are present
			// in context, pass through. DevAuthMiddleware sets "user/*.*"
			// so this only triggers when the context is completely empty
			// (e.g. during development without any auth middleware).
			if len(rawScopes) == 0 {
				return next(c)
			}

			// Extract FHIR resource type from the URL path.
			resourceType := extractFHIRResourceType(path)
			if resourceType == "" {
				// Cannot determine resource type (e.g. system-level operation);
				// let downstream handlers decide.
				return next(c)
			}

			// Determine the required operation from HTTP method and path.
			operation := fhirMethodToOperation(c.Request().Method, path)

			// Parse the raw scopes into structured SMART scopes.
			smartScopes := ParseSMARTScopes(rawScopes)

			// Check if any granted scope covers the required resource + operation.
			if ScopeAllows(smartScopes, resourceType, operation) {
				return next(c)
			}

			// Scope check failed -- return a FHIR OperationOutcome.
			return fhirScopeForbidden(c, resourceType, operation)
		}
	}
}

// isScopeExemptPath returns true for FHIR paths that should not require
// scope authorization (discovery and capability endpoints).
func isScopeExemptPath(path string) bool {
	// Normalize: remove trailing slash.
	p := strings.TrimRight(path, "/")

	// CapabilityStatement
	if p == "/fhir/metadata" {
		return true
	}

	// SMART well-known discovery endpoints.
	if strings.HasPrefix(p, "/fhir/.well-known/") || p == "/fhir/.well-known" {
		return true
	}

	return false
}

// extractFHIRResourceType extracts the FHIR resource type from a path like
// /fhir/Patient/123 -> "Patient" or /fhir/Patient/_search -> "Patient".
// Returns an empty string if the resource type cannot be determined.
func extractFHIRResourceType(path string) string {
	// Trim leading slash and split: ["fhir", "Patient", "123", ...]
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) < 2 || parts[0] != "fhir" {
		return ""
	}

	candidate := parts[1]

	// Skip system-level operations (e.g. /fhir/$export, /fhir/metadata).
	if candidate == "" || strings.HasPrefix(candidate, "$") || strings.HasPrefix(candidate, ".") {
		return ""
	}
	// Skip "metadata" -- handled by isScopeExemptPath, but belt-and-suspenders.
	if candidate == "metadata" {
		return ""
	}

	return candidate
}

// fhirMethodToOperation maps an HTTP method (and request path) to the SMART
// scope operation ("read" or "write").
//
// POST is context-sensitive:
//   - POST to /fhir/<Resource>/_search is a search -> "read"
//   - POST to /fhir/<Resource> (no _search) is a create -> "write"
func fhirMethodToOperation(method, path string) string {
	switch method {
	case http.MethodGet, http.MethodHead:
		return "read"
	case http.MethodPost:
		if isSearchPath(path) {
			return "read"
		}
		return "write"
	case http.MethodPut, http.MethodPatch, http.MethodDelete:
		return "write"
	default:
		return "read"
	}
}

// isSearchPath returns true if the path indicates a FHIR search operation,
// i.e. it ends with /_search or contains /_search/.
func isSearchPath(path string) bool {
	return strings.HasSuffix(path, "/_search") || strings.Contains(path, "/_search/")
}

// fhirScopeForbidden writes a 403 response with a FHIR OperationOutcome body.
func fhirScopeForbidden(c echo.Context, resourceType, operation string) error {
	outcome := map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue": []map[string]interface{}{
			{
				"severity":    "error",
				"code":        "forbidden",
				"diagnostics": fmt.Sprintf("insufficient scope: required %s.%s", resourceType, operation),
			},
		},
	}

	body, err := json.Marshal(outcome)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden,
			fmt.Sprintf("insufficient scope: required %s.%s", resourceType, operation))
	}

	return c.JSONBlob(http.StatusForbidden, body)
}
