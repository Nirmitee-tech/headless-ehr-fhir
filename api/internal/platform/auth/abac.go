package auth

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
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

// ABACDecision represents the result of an ABAC policy evaluation.
type ABACDecision struct {
	Allowed        bool   `json:"allowed"`
	Reason         string `json:"reason"`
	RequireConsent bool   `json:"require_consent,omitempty"`
	RequireCareTeam bool  `json:"require_care_team,omitempty"`
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
			return &ABACDecision{
				Allowed:         true,
				Reason:          "policy match",
				RequireConsent:  policy.RequireConsent,
				RequireCareTeam: policy.RequireCareTeam,
			}
		}
	}

	// No policy found - default deny
	return &ABACDecision{Allowed: false, Reason: "no policy for " + resourceType}
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

			// Propagate consent/care-team requirements via echo context
			// so downstream middleware (ConsentEnforcementMiddleware) can check them.
			if decision.RequireConsent {
				c.Set("require_consent", true)
			}
			if decision.RequireCareTeam {
				c.Set("require_care_team", true)
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

// ---------------------------------------------------------------------------
// Consent enforcement
// ---------------------------------------------------------------------------

// ConsentInfo holds the fields relevant for consent enforcement decisions.
// This is a package-local type that avoids importing domain packages.
type ConsentInfo struct {
	Status          string
	Scope           string
	ProvisionType   string
	ProvisionAction string
	ProvisionStart  *time.Time
	ProvisionEnd    *time.Time
}

// ConsentChecker is the interface that the consent enforcement middleware
// uses to look up active consents for a patient.  The documents service (or
// any adapter around the ConsentRepository) should implement this interface.
type ConsentChecker interface {
	ListActiveConsentsForPatient(ctx context.Context, patientID uuid.UUID) ([]*ConsentInfo, error)
}

// consentOperationOutcome returns a FHIR OperationOutcome JSON body for a
// consent-related 403 response.
func consentOperationOutcome(diagnostics string) map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue": []map[string]interface{}{
			{
				"severity":    "error",
				"code":        "forbidden",
				"diagnostics": diagnostics,
			},
		},
	}
}

// httpMethodToFHIRAction maps an HTTP method to a FHIR consent provision action.
func httpMethodToFHIRAction(method string) string {
	switch method {
	case http.MethodGet, http.MethodHead:
		return "access"
	case http.MethodPost:
		// POST can be either create or search; treat as "access" for consent purposes
		// since FHIR search is POST-based as well.  Write-side POST (create) maps to
		// "correct" but the more permissive default is "access".
		return "access"
	case http.MethodPut, http.MethodPatch, http.MethodDelete:
		return "correct"
	default:
		return "access"
	}
}

// extractPatientID attempts to extract a patient UUID from the echo context.
// It checks, in order: path param "id" (for routes like /fhir/Patient/:id),
// path param "patient_id", query param "patient", query param "subject".
func extractPatientID(c echo.Context) (uuid.UUID, bool) {
	// For Patient resource routes the :id IS the patient ID.
	resourceType := extractABACResourceType(c.Path())

	if resourceType == "Patient" {
		if idStr := c.Param("id"); idStr != "" {
			if id, err := uuid.Parse(idStr); err == nil {
				return id, true
			}
		}
	}

	// Explicit patient_id path param (used by some sub-resource routes).
	if pidStr := c.Param("patient_id"); pidStr != "" {
		if id, err := uuid.Parse(pidStr); err == nil {
			return id, true
		}
	}

	// Query params: ?patient=<uuid> or ?subject=<uuid>
	for _, qp := range []string{"patient", "subject"} {
		val := c.QueryParam(qp)
		if val == "" {
			continue
		}
		// Strip a possible "Patient/" FHIR reference prefix.
		val = strings.TrimPrefix(val, "Patient/")
		if id, err := uuid.Parse(val); err == nil {
			return id, true
		}
	}

	return uuid.Nil, false
}

// ConsentEnforcementMiddleware checks active Consent resources for the patient
// being accessed.  If the ABAC policy for the resource type requires consent
// (signalled by ABACMiddleware setting "require_consent" on the echo context),
// the middleware verifies that an active, non-expired consent exists that
// permits the requested action.
//
// If checker is nil the middleware logs a warning once and passes all requests
// through (backward-compatible).
func ConsentEnforcementMiddleware(checker ConsentChecker) echo.MiddlewareFunc {
	if checker == nil {
		log.Println("WARN: ConsentEnforcementMiddleware initialized without a ConsentChecker; all requests will pass through")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// ---- fast-path: no checker configured ----
			if checker == nil {
				return next(c)
			}

			// ---- check if ABAC flagged this request as requiring consent ----
			requireConsent, _ := c.Get("require_consent").(bool)
			if !requireConsent {
				return next(c)
			}

			// ---- admin bypass ----
			roles := RolesFromContext(c.Request().Context())
			for _, r := range roles {
				if r == "admin" {
					return next(c)
				}
			}

			// ---- extract patient ID ----
			patientID, ok := extractPatientID(c)
			if !ok {
				// Cannot determine patient -- deny to be safe.
				return c.JSON(http.StatusForbidden, consentOperationOutcome(
					"consent required but patient could not be identified from the request"))
			}

			// ---- look up active consents ----
			consents, err := checker.ListActiveConsentsForPatient(c.Request().Context(), patientID)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, consentOperationOutcome(
					"error retrieving consent records"))
			}

			action := httpMethodToFHIRAction(c.Request().Method)
			now := time.Now()

			// Evaluate consents.  "deny" takes precedence over "permit".
			hasPermit := false
			for _, consent := range consents {
				// Only consider active consents.
				if consent.Status != "active" {
					continue
				}

				// Check provision period.
				if consent.ProvisionStart != nil && now.Before(*consent.ProvisionStart) {
					continue
				}
				if consent.ProvisionEnd != nil && now.After(*consent.ProvisionEnd) {
					continue
				}

				// Check provision action matches (empty action = applies to all).
				if consent.ProvisionAction != "" && consent.ProvisionAction != action {
					continue
				}

				// Evaluate provision type.
				switch consent.ProvisionType {
				case "deny":
					return c.JSON(http.StatusForbidden, consentOperationOutcome(
						"access denied by patient consent directive"))
				case "permit":
					hasPermit = true
				default:
					// If provision type is empty/unknown, treat as permit.
					hasPermit = true
				}
			}

			if !hasPermit {
				return c.JSON(http.StatusForbidden, consentOperationOutcome(
					"no active consent permits this action for the patient"))
			}

			return next(c)
		}
	}
}
