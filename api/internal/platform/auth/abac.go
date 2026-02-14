package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// ABACPolicy defines an attribute-based access control policy.
type ABACPolicy struct {
	ResourceType    string   `json:"resource_type"`
	AllowedRoles    []string `json:"allowed_roles"`
	RequireCareTeam bool     `json:"require_care_team"`
	RequireConsent  bool     `json:"require_consent"`
	DepartmentScope []string `json:"department_scope,omitempty"`
}

// ABACEngine evaluates attribute-based access control policies.
type ABACEngine struct {
	policies []ABACPolicy
}

// NewABACEngine creates a new ABAC engine with the given policies.
func NewABACEngine(policies []ABACPolicy) *ABACEngine {
	return &ABACEngine{policies: policies}
}

// DefaultPolicies returns the default ABAC policies for the EHR.
func DefaultPolicies() []ABACPolicy {
	return []ABACPolicy{
		{ResourceType: "Patient", AllowedRoles: []string{"admin", "physician", "nurse", "receptionist"}, RequireCareTeam: false, RequireConsent: false},
		{ResourceType: "Condition", AllowedRoles: []string{"admin", "physician", "nurse"}, RequireCareTeam: false, RequireConsent: true},
		{ResourceType: "Observation", AllowedRoles: []string{"admin", "physician", "nurse"}, RequireCareTeam: false, RequireConsent: true},
		{ResourceType: "MedicationRequest", AllowedRoles: []string{"admin", "physician"}, RequireCareTeam: false, RequireConsent: true},
		{ResourceType: "Encounter", AllowedRoles: []string{"admin", "physician", "nurse"}, RequireCareTeam: false, RequireConsent: false},
	}
}

// Evaluate checks if the given context allows access to the resource.
func (e *ABACEngine) Evaluate(ctx context.Context, resourceType string) *ABACDecision {
	roles := RolesFromContext(ctx)

	// Admin bypass
	for _, r := range roles {
		if r == "admin" {
			return &ABACDecision{Allowed: true, Reason: "admin role"}
		}
	}

	// Find matching policy
	for _, policy := range e.policies {
		if policy.ResourceType == resourceType {
			// Check role
			roleMatch := false
			for _, allowedRole := range policy.AllowedRoles {
				for _, userRole := range roles {
					if userRole == allowedRole {
						roleMatch = true
						break
					}
				}
				if roleMatch {
					break
				}
			}
			if !roleMatch {
				return &ABACDecision{Allowed: false, Reason: "insufficient role for " + resourceType}
			}
			return &ABACDecision{Allowed: true, Reason: "policy match"}
		}
	}

	// No policy found - default deny
	return &ABACDecision{Allowed: false, Reason: "no policy for " + resourceType}
}

// ABACDecision represents the result of an ABAC policy evaluation.
type ABACDecision struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

// ABACMiddleware returns middleware that enforces ABAC policies.
func ABACMiddleware(engine *ABACEngine) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract resource type from path
			path := c.Path()
			resourceType := extractABACResourceType(path)
			if resourceType == "" {
				return next(c)
			}

			decision := engine.Evaluate(c.Request().Context(), resourceType)
			if !decision.Allowed {
				return echo.NewHTTPError(http.StatusForbidden, decision.Reason)
			}

			return next(c)
		}
	}
}

// extractABACResourceType extracts the FHIR resource type from a path like /fhir/Patient/123.
func extractABACResourceType(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "fhir" {
		return parts[1]
	}
	return ""
}

// ConsentEnforcementMiddleware checks patient consent before allowing access.
func ConsentEnforcementMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Consent enforcement is a post-filter - for now, pass through
			// In full implementation, this would check active Consent resources
			// for the patient and filter results accordingly
			return next(c)
		}
	}
}
