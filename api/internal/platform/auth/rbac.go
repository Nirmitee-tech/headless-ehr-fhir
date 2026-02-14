package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// RequireRole returns middleware that checks if the user has at least one of the specified roles.
func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRoles := RolesFromContext(c.Request().Context())
			for _, required := range roles {
				for _, has := range userRoles {
					if has == required || has == "admin" {
						return next(c)
					}
				}
			}
			return echo.NewHTTPError(http.StatusForbidden,
				fmt.Sprintf("required role: %s", strings.Join(roles, " or ")))
		}
	}
}

// RequireScope returns middleware that checks if the user has the required FHIR scope.
// Scopes follow SMART on FHIR format: "resource/operation" (e.g., "Patient.read", "user/*.read").
func RequireScope(resource, operation string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			scopes := ScopesFromContext(c.Request().Context())
			required := fmt.Sprintf("%s.%s", resource, operation)

			for _, scope := range scopes {
				if matchScope(scope, required) {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden,
				fmt.Sprintf("required scope: %s", required))
		}
	}
}

// matchScope checks if a granted scope covers the required scope.
// Supports wildcards: "user/*.*" matches everything, "patient/*.read" matches any read.
func matchScope(granted, required string) bool {
	if granted == required {
		return true
	}

	gParts := strings.SplitN(granted, ".", 2)
	rParts := strings.SplitN(required, ".", 2)

	if len(gParts) != 2 || len(rParts) != 2 {
		return false
	}

	gRes, gOp := gParts[0], gParts[1]
	rRes, rOp := rParts[0], rParts[1]

	// Check resource match
	resMatch := gRes == rRes || gRes == "user/*" || gRes == "patient/*"
	// Check operation match
	opMatch := gOp == rOp || gOp == "*"

	return resMatch && opMatch
}
