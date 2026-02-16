package fhir

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// FHIR Consent-based access control enforcement
// ---------------------------------------------------------------------------

// ConsentScope represents the scope of a FHIR Consent resource, categorizing
// the broad area of policy (privacy, research, advanced directives, treatment).
type ConsentScope string

const (
	ConsentScopePatientPrivacy ConsentScope = "patient-privacy"
	ConsentScopeResearch       ConsentScope = "research"
	ConsentScopeADR            ConsentScope = "adr"
	ConsentScopeTreatment      ConsentScope = "treatment"
)

// ConsentStatus represents the lifecycle status of a FHIR Consent resource.
type ConsentStatus string

const (
	ConsentStatusDraft          ConsentStatus = "draft"
	ConsentStatusProposed       ConsentStatus = "proposed"
	ConsentStatusActive         ConsentStatus = "active"
	ConsentStatusRejected       ConsentStatus = "rejected"
	ConsentStatusInactive       ConsentStatus = "inactive"
	ConsentStatusEnteredInError ConsentStatus = "entered-in-error"
)

// ConsentDecision represents the outcome of evaluating consent policies
// against an access request.
type ConsentDecision string

const (
	ConsentDecisionPermit    ConsentDecision = "permit"
	ConsentDecisionDeny      ConsentDecision = "deny"
	ConsentDecisionNoConsent ConsentDecision = "no-consent"
)

// Contains returns true if the given time falls within the period. A nil
// bound means the period is open-ended in that direction.
func (p *Period) Contains(t time.Time) bool {
	if p == nil {
		return true
	}
	if p.Start != nil && t.Before(*p.Start) {
		return false
	}
	if p.End != nil && t.After(*p.End) {
		return false
	}
	return true
}

// ConsentActor identifies an actor (e.g., practitioner, organization) and
// their role within a consent provision.
type ConsentActor struct {
	// Role describes the actor's participation (e.g., "primary", "delegated").
	Role string

	// Reference is a FHIR reference to the actor (e.g., "Practitioner/123").
	Reference string
}

// ConsentProvision describes the rules that apply when a consent is evaluated.
// It specifies who can do what with which data under what conditions.
type ConsentProvision struct {
	// Type is either "deny" or "permit".
	Type string

	// Period is the time window during which this provision is effective.
	Period *Period

	// Actor lists the actors to whom this provision applies.
	Actor []ConsentActor

	// Action lists the permitted or denied actions (e.g., "access", "correct", "disclose").
	Action []string

	// SecurityLabel lists security label codes that scope this provision.
	SecurityLabel []string

	// Purpose lists purpose-of-use codes (e.g., "TREAT", "HPAYMT", "HOPERAT").
	Purpose []string

	// ResourceClass lists FHIR resource types this provision applies to
	// (e.g., "Observation", "MedicationRequest").
	ResourceClass []string

	// DataPeriod restricts the provision to data created within this period.
	DataPeriod *Period
}

// ConsentPolicy represents a complete FHIR Consent resource with its
// provision rules, used for access control enforcement.
type ConsentPolicy struct {
	ID        string
	PatientID string
	Scope     ConsentScope
	Status    ConsentStatus
	Provision ConsentProvision
	CreatedAt time.Time
}

// ConsentAccessRequest describes the parameters of an access request that
// should be evaluated against consent policies.
type ConsentAccessRequest struct {
	PatientID      string
	ActorReference string
	ResourceType   string
	Purpose        string
	SecurityLabels []string
	AccessTime     time.Time
}

// ---------------------------------------------------------------------------
// ConsentStore – persistence interface and in-memory implementation
// ---------------------------------------------------------------------------

// ConsentStore defines the interface for consent policy persistence.
type ConsentStore interface {
	// GetActiveConsents returns all active consent policies for the given patient.
	GetActiveConsents(patientID string) ([]ConsentPolicy, error)

	// AddConsent stores a new consent policy.
	AddConsent(policy ConsentPolicy) error

	// RevokeConsent sets the identified consent policy status to inactive.
	RevokeConsent(id string) error
}

// InMemoryConsentStore is a thread-safe, in-memory implementation of ConsentStore
// suitable for testing and development.
type InMemoryConsentStore struct {
	mu       sync.RWMutex
	policies map[string]ConsentPolicy
}

// NewInMemoryConsentStore creates a new empty in-memory consent store.
func NewInMemoryConsentStore() *InMemoryConsentStore {
	return &InMemoryConsentStore{
		policies: make(map[string]ConsentPolicy),
	}
}

// GetActiveConsents returns all consent policies for the given patient that
// have an active status.
func (s *InMemoryConsentStore) GetActiveConsents(patientID string) ([]ConsentPolicy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []ConsentPolicy
	for _, p := range s.policies {
		if p.PatientID == patientID && p.Status == ConsentStatusActive {
			result = append(result, p)
		}
	}
	return result, nil
}

// AddConsent stores a consent policy. If a policy with the same ID already
// exists it is overwritten.
func (s *InMemoryConsentStore) AddConsent(policy ConsentPolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.policies[policy.ID] = policy
	return nil
}

// RevokeConsent sets the status of the identified policy to inactive. Returns
// an error if the policy is not found.
func (s *InMemoryConsentStore) RevokeConsent(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.policies[id]
	if !ok {
		return fmt.Errorf("consent policy %q not found", id)
	}
	p.Status = ConsentStatusInactive
	s.policies[id] = p
	return nil
}

// ---------------------------------------------------------------------------
// Consent evaluation engine
// ---------------------------------------------------------------------------

// EvaluateConsent evaluates all provided consent policies against the given
// access request and returns a decision.
//
// Rules:
//  1. Only policies with status "active" are considered.
//  2. A policy's provision must match the request on all specified dimensions
//     (actor, resource type, purpose, security labels, provision period, data period).
//  3. If any matching provision has type "deny", the overall decision is "deny"
//     (deny overrides permit).
//  4. If any matching provision has type "permit" and no deny matched, the
//     decision is "permit".
//  5. If no provisions match, the decision is "no-consent".
func EvaluateConsent(policies []ConsentPolicy, request ConsentAccessRequest) ConsentDecision {
	hasPermit := false
	hasDeny := false

	for _, policy := range policies {
		if policy.Status != ConsentStatusActive {
			continue
		}

		if !provisionMatches(&policy.Provision, &request) {
			continue
		}

		switch policy.Provision.Type {
		case "deny":
			hasDeny = true
		case "permit":
			hasPermit = true
		}
	}

	// Deny overrides permit.
	if hasDeny {
		return ConsentDecisionDeny
	}
	if hasPermit {
		return ConsentDecisionPermit
	}
	return ConsentDecisionNoConsent
}

// provisionMatches returns true if the provision's constraints are all
// satisfied by the given request. An empty/nil constraint is treated as
// matching everything (no restriction on that dimension).
func provisionMatches(prov *ConsentProvision, req *ConsentAccessRequest) bool {
	// Check provision period.
	if prov.Period != nil && !prov.Period.Contains(req.AccessTime) {
		return false
	}

	// Check actor restriction.
	if len(prov.Actor) > 0 {
		actorMatch := false
		for _, a := range prov.Actor {
			if a.Reference == req.ActorReference {
				actorMatch = true
				break
			}
		}
		if !actorMatch {
			return false
		}
	}

	// Check resource type restriction.
	if len(prov.ResourceClass) > 0 {
		resourceMatch := false
		for _, rt := range prov.ResourceClass {
			if rt == req.ResourceType {
				resourceMatch = true
				break
			}
		}
		if !resourceMatch {
			return false
		}
	}

	// Check purpose restriction.
	if len(prov.Purpose) > 0 {
		purposeMatch := false
		for _, p := range prov.Purpose {
			if p == req.Purpose {
				purposeMatch = true
				break
			}
		}
		if !purposeMatch {
			return false
		}
	}

	// Check security label restriction.
	if len(prov.SecurityLabel) > 0 {
		labelMatch := false
		reqLabelSet := make(map[string]bool, len(req.SecurityLabels))
		for _, l := range req.SecurityLabels {
			reqLabelSet[l] = true
		}
		for _, sl := range prov.SecurityLabel {
			if reqLabelSet[sl] {
				labelMatch = true
				break
			}
		}
		if !labelMatch {
			return false
		}
	}

	// Check data period – the access time must fall within the data period
	// to indicate the request concerns data from that window.
	if prov.DataPeriod != nil && !prov.DataPeriod.Contains(req.AccessTime) {
		return false
	}

	return true
}

// ---------------------------------------------------------------------------
// Consent enforcement middleware
// ---------------------------------------------------------------------------

// ConsentEnforcementConfig controls the behaviour of the consent enforcement
// middleware.
type ConsentEnforcementConfig struct {
	// DefaultDecision is used when no consent policies match. Set to "permit"
	// for opt-out systems (access allowed unless denied) or "deny" for opt-in
	// systems (access denied unless explicitly permitted).
	DefaultDecision ConsentDecision

	// RequireConsent when true treats a "no-consent" evaluation result as a
	// deny, regardless of DefaultDecision. This enforces that explicit consent
	// must exist for every access.
	RequireConsent bool

	// ExemptResourceTypes lists resource types that bypass consent enforcement
	// entirely (e.g., "CapabilityStatement", "OperationDefinition").
	ExemptResourceTypes []string
}

// ConsentEnforcementMiddleware returns Echo middleware that enforces FHIR
// Consent policies using the default configuration (opt-out: no-consent is
// treated as permit).
func ConsentEnforcementMiddleware(store ConsentStore) echo.MiddlewareFunc {
	return NewConsentEnforcementMiddleware(store, ConsentEnforcementConfig{
		DefaultDecision: ConsentDecisionPermit,
	})
}

// NewConsentEnforcementMiddleware returns Echo middleware that enforces FHIR
// Consent policies with the provided configuration.
//
// The middleware:
//   - Extracts the patient ID from the X-Patient-ID request header or the
//     :patientId path parameter.
//   - Evaluates consent policies from the store against the request.
//   - Returns 403 with an OperationOutcome if consent denies access.
//   - Passes through if consent permits access or if no consent exists (as
//     governed by the configuration).
//   - Sets the X-Consent-Decision response header.
func NewConsentEnforcementMiddleware(store ConsentStore, config ConsentEnforcementConfig) echo.MiddlewareFunc {
	exemptSet := make(map[string]bool, len(config.ExemptResourceTypes))
	for _, rt := range config.ExemptResourceTypes {
		exemptSet[rt] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Determine the resource type from the request path.
			resourceType := extractResourceTypeFromPath(c.Request().URL.Path)

			// Skip exempt resource types.
			if exemptSet[resourceType] {
				c.Response().Header().Set("X-Consent-Decision", string(ConsentDecisionPermit))
				return next(c)
			}

			// Extract patient ID from header or path parameter.
			patientID := c.Request().Header.Get("X-Patient-ID")
			if patientID == "" {
				patientID = c.Param("patientId")
			}

			// If we cannot identify the patient, apply the default decision.
			if patientID == "" {
				decision := config.DefaultDecision
				if decision == "" {
					decision = ConsentDecisionPermit
				}
				c.Response().Header().Set("X-Consent-Decision", string(decision))
				if decision == ConsentDecisionDeny {
					return c.JSON(http.StatusForbidden, NewOperationOutcome(
						"error", "forbidden",
						"Access denied: no patient context and consent is required"))
				}
				return next(c)
			}

			// Retrieve active consent policies.
			policies, err := store.GetActiveConsents(patientID)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
					"error", "exception",
					fmt.Sprintf("Failed to retrieve consent policies: %v", err)))
			}

			// Build the access request from context.
			req := ConsentAccessRequest{
				PatientID:    patientID,
				ResourceType: resourceType,
				AccessTime:   time.Now(),
			}

			if actor := c.Request().Header.Get("X-Actor-Reference"); actor != "" {
				req.ActorReference = actor
			}
			if purpose := c.Request().Header.Get("X-Purpose-Of-Use"); purpose != "" {
				req.Purpose = purpose
			}
			if labels := c.Request().Header.Get("X-Security-Labels"); labels != "" {
				parts := strings.Split(labels, ",")
				for i := range parts {
					parts[i] = strings.TrimSpace(parts[i])
				}
				req.SecurityLabels = parts
			}

			// Evaluate consent.
			decision := EvaluateConsent(policies, req)

			// Apply default decision / require-consent logic.
			if decision == ConsentDecisionNoConsent {
				if config.RequireConsent {
					decision = ConsentDecisionDeny
				} else if config.DefaultDecision != "" {
					decision = config.DefaultDecision
				} else {
					decision = ConsentDecisionPermit
				}
			}

			c.Response().Header().Set("X-Consent-Decision", string(decision))

			if decision == ConsentDecisionDeny {
				return c.JSON(http.StatusForbidden, NewOperationOutcome(
					"error", "forbidden",
					"Access denied: consent policy does not permit this access"))
			}

			return next(c)
		}
	}
}

// extractResourceTypeFromPath extracts the FHIR resource type from a URL path.
// It looks for the first path segment that starts with an uppercase letter,
// which follows the FHIR convention for resource type names.
func extractResourceTypeFromPath(path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	for _, seg := range segments {
		if len(seg) > 0 && seg[0] >= 'A' && seg[0] <= 'Z' {
			return seg
		}
	}
	return ""
}
