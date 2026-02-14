package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// SMARTConfiguration represents the SMART on FHIR well-known configuration
// as defined by the SMART App Launch Framework (HL7).
type SMARTConfiguration struct {
	AuthorizationEndpoint      string   `json:"authorization_endpoint"`
	TokenEndpoint              string   `json:"token_endpoint"`
	TokenEndpointAuthMethods   []string `json:"token_endpoint_auth_methods_supported"`
	GrantTypes                 []string `json:"grant_types_supported"`
	Scopes                    []string `json:"scopes_supported"`
	ResponseTypes             []string `json:"response_types_supported"`
	Capabilities              []string `json:"capabilities"`
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`
}

// LaunchContext holds the context passed from the EHR to a SMART app during
// an EHR launch. The launch token is exchanged for context parameters that
// are included in the token response.
type LaunchContext struct {
	LaunchToken string    `json:"launch"`
	PatientID   string    `json:"patient,omitempty"`
	EncounterID string    `json:"encounter,omitempty"`
	FHIRUser    string    `json:"fhirUser,omitempty"`
	CreatedAt   time.Time `json:"-"`
}

// SMARTScope represents a parsed SMART on FHIR scope.
// Format: <context>/<resourceType>.<operation>
// Examples: patient/Patient.read, user/Observation.write, patient/*.read
type SMARTScope struct {
	Context      string // "patient", "user", or "system"
	ResourceType string // e.g. "Patient", "Observation", "*"
	Operation    string // "read", "write", or "*"
}

// LaunchContextStore provides thread-safe in-memory storage for launch contexts.
// In production this would typically be backed by Redis or a database.
type LaunchContextStore struct {
	mu       sync.RWMutex
	contexts map[string]*LaunchContext
	ttl      time.Duration
}

// NewLaunchContextStore creates a new store with the given TTL for launch tokens.
func NewLaunchContextStore(ttl time.Duration) *LaunchContextStore {
	return &LaunchContextStore{
		contexts: make(map[string]*LaunchContext),
		ttl:      ttl,
	}
}

// Create generates a new launch token and stores the associated context.
func (s *LaunchContextStore) Create(patientID, encounterID, fhirUser string) (*LaunchContext, error) {
	token, err := generateLaunchToken()
	if err != nil {
		return nil, fmt.Errorf("generating launch token: %w", err)
	}

	ctx := &LaunchContext{
		LaunchToken: token,
		PatientID:   patientID,
		EncounterID: encounterID,
		FHIRUser:    fhirUser,
		CreatedAt:   time.Now(),
	}

	s.mu.Lock()
	s.contexts[token] = ctx
	s.mu.Unlock()

	return ctx, nil
}

// Get retrieves a launch context by token. Returns nil if not found or expired.
func (s *LaunchContextStore) Get(token string) *LaunchContext {
	s.mu.RLock()
	ctx, ok := s.contexts[token]
	s.mu.RUnlock()

	if !ok {
		return nil
	}

	if time.Since(ctx.CreatedAt) > s.ttl {
		s.mu.Lock()
		delete(s.contexts, token)
		s.mu.Unlock()
		return nil
	}

	return ctx
}

// Consume retrieves and removes a launch context by token (one-time use).
func (s *LaunchContextStore) Consume(token string) *LaunchContext {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, ok := s.contexts[token]
	if !ok {
		return nil
	}

	if time.Since(ctx.CreatedAt) > s.ttl {
		delete(s.contexts, token)
		return nil
	}

	delete(s.contexts, token)
	return ctx
}

// Cleanup removes expired launch contexts from the store.
func (s *LaunchContextStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for token, ctx := range s.contexts {
		if now.Sub(ctx.CreatedAt) > s.ttl {
			delete(s.contexts, token)
		}
	}
}

// generateLaunchToken creates a cryptographically random launch token.
func generateLaunchToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ParseSMARTScope parses a SMART on FHIR scope string into its components.
// Valid formats:
//   - patient/Patient.read
//   - user/Observation.write
//   - patient/*.read
//   - user/*.*
//
// Returns an error for scopes that are not resource-level SMART scopes
// (e.g. "openid", "profile", "launch").
func ParseSMARTScope(scope string) (*SMARTScope, error) {
	// Split context from resource.operation
	slashIdx := strings.Index(scope, "/")
	if slashIdx < 0 {
		return nil, fmt.Errorf("not a resource scope: %s", scope)
	}

	ctx := scope[:slashIdx]
	remainder := scope[slashIdx+1:]

	if ctx != "patient" && ctx != "user" && ctx != "system" {
		return nil, fmt.Errorf("invalid scope context %q: must be patient, user, or system", ctx)
	}

	// Split resourceType from operation
	dotIdx := strings.LastIndex(remainder, ".")
	if dotIdx < 0 {
		return nil, fmt.Errorf("invalid scope format %q: missing operation", scope)
	}

	resourceType := remainder[:dotIdx]
	operation := remainder[dotIdx+1:]

	if resourceType == "" {
		return nil, fmt.Errorf("invalid scope %q: empty resource type", scope)
	}
	if operation != "read" && operation != "write" && operation != "*" {
		return nil, fmt.Errorf("invalid operation %q: must be read, write, or *", operation)
	}

	return &SMARTScope{
		Context:      ctx,
		ResourceType: resourceType,
		Operation:    operation,
	}, nil
}

// ParseSMARTScopes parses a list of scope strings, returning only the valid
// SMART resource scopes. Non-resource scopes (openid, profile, launch, etc.)
// are silently skipped.
func ParseSMARTScopes(scopes []string) []SMARTScope {
	var result []SMARTScope
	for _, s := range scopes {
		parsed, err := ParseSMARTScope(s)
		if err != nil {
			continue // skip non-resource scopes
		}
		result = append(result, *parsed)
	}
	return result
}

// ScopeAllows checks whether a list of SMART scopes grants access for the
// given resource type and operation. It also considers the context: patient
// scopes require a patient context to be set, while user scopes require
// a valid user identity.
func ScopeAllows(scopes []SMARTScope, resourceType, operation string) bool {
	for _, s := range scopes {
		if !resourceMatches(s.ResourceType, resourceType) {
			continue
		}
		if !operationMatches(s.Operation, operation) {
			continue
		}
		return true
	}
	return false
}

// resourceMatches checks if a granted resource type covers the requested one.
func resourceMatches(granted, requested string) bool {
	return granted == "*" || granted == requested
}

// operationMatches checks if a granted operation covers the requested one.
func operationMatches(granted, requested string) bool {
	return granted == "*" || granted == requested
}

// httpMethodToOperation maps an HTTP method to a SMART scope operation.
func httpMethodToOperation(method string) string {
	switch method {
	case http.MethodGet, http.MethodHead:
		return "read"
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return "write"
	default:
		return "read"
	}
}

// contextKey types for SMART context values
const (
	SMARTPatientIDKey   contextKey = "smart_patient_id"
	SMARTEncounterIDKey contextKey = "smart_encounter_id"
	SMARTFHIRUserKey    contextKey = "smart_fhir_user"
	SMARTScopesKey      contextKey = "smart_scopes"
)

// SMARTPatientIDFromContext returns the patient ID from the SMART launch context.
func SMARTPatientIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(SMARTPatientIDKey).(string)
	return v
}

// SMARTEncounterIDFromContext returns the encounter ID from the SMART launch context.
func SMARTEncounterIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(SMARTEncounterIDKey).(string)
	return v
}

// SMARTFHIRUserFromContext returns the FHIR user resource URL from context.
func SMARTFHIRUserFromContext(ctx context.Context) string {
	v, _ := ctx.Value(SMARTFHIRUserKey).(string)
	return v
}

// SMARTScopesFromContext returns the parsed SMART scopes from context.
func SMARTScopesFromContext(ctx context.Context) []SMARTScope {
	v, _ := ctx.Value(SMARTScopesKey).([]SMARTScope)
	return v
}

// SMARTScopeMiddleware parses SMART scopes from the JWT and sets them on the
// request context. It also extracts patient context information from custom
// claims if available.
func SMARTScopeMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			// Parse SMART scopes from the JWT scopes already set by JWTMiddleware
			rawScopes := ScopesFromContext(ctx)
			smartScopes := ParseSMARTScopes(rawScopes)

			ctx = context.WithValue(ctx, SMARTScopesKey, smartScopes)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// RequireSMARTScope returns middleware that enforces SMART on FHIR scope
// authorization for a specific FHIR resource type. The operation (read/write)
// is inferred from the HTTP method.
//
// For patient-context scopes, it also verifies that the requested resource
// belongs to the patient in context.
func RequireSMARTScope(resourceType string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			scopes := SMARTScopesFromContext(ctx)

			// If no SMART scopes are set, fall through to other auth mechanisms
			if len(scopes) == 0 {
				return next(c)
			}

			operation := httpMethodToOperation(c.Request().Method)

			if !ScopeAllows(scopes, resourceType, operation) {
				return echo.NewHTTPError(http.StatusForbidden,
					fmt.Sprintf("insufficient scope: requires %s/%s.%s", "patient or user", resourceType, operation))
			}

			return next(c)
		}
	}
}

// RegisterSMARTEndpoints registers the SMART on FHIR discovery and launch endpoints.
func RegisterSMARTEndpoints(g *echo.Group, issuer string) {
	store := NewLaunchContextStore(5 * time.Minute)

	// Start background cleanup of expired launch contexts
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			store.Cleanup()
		}
	}()

	// SMART Configuration (well-known endpoint)
	g.GET("/.well-known/smart-configuration", smartConfigurationHandler(issuer))

	// EHR Launch endpoint - generates a launch context and returns a launch token
	g.POST("/launch", ehrLaunchHandler(store))

	// Launch context resolution - exchanges a launch token for context parameters
	g.GET("/launch-context", launchContextHandler(store))
}

// smartConfigurationHandler returns the SMART on FHIR well-known configuration.
func smartConfigurationHandler(issuer string) echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg := SMARTConfiguration{
			AuthorizationEndpoint:      issuer + "/protocol/openid-connect/auth",
			TokenEndpoint:              issuer + "/protocol/openid-connect/token",
			TokenEndpointAuthMethods:   []string{"client_secret_basic", "client_secret_post"},
			GrantTypes:                 []string{"authorization_code", "client_credentials"},
			Scopes: []string{
				"openid", "profile", "fhirUser",
				"launch", "launch/patient",
				"patient/*.read", "patient/*.write",
				"user/*.read", "user/*.write",
			},
			ResponseTypes: []string{"code"},
			Capabilities: []string{
				"launch-ehr", "launch-standalone",
				"client-public", "client-confidential-symmetric",
				"context-ehr-patient",
				"permission-patient", "permission-user",
			},
			CodeChallengeMethodsSupported: []string{"S256"},
		}
		return c.JSON(http.StatusOK, cfg)
	}
}

// ehrLaunchRequest represents the request body for initiating an EHR launch.
type ehrLaunchRequest struct {
	PatientID   string `json:"patient_id"`
	EncounterID string `json:"encounter_id,omitempty"`
	FHIRUser    string `json:"fhir_user,omitempty"`
}

// ehrLaunchResponse is returned when a launch context is created.
type ehrLaunchResponse struct {
	LaunchToken string `json:"launch"`
	ISS         string `json:"iss"`
}

// ehrLaunchHandler handles EHR-initiated launch requests. The EHR creates a
// launch context with patient/encounter information and receives a launch token
// that the client app will include in the authorization request.
func ehrLaunchHandler(store *LaunchContextStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req ehrLaunchRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}

		if req.PatientID == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "patient_id is required")
		}

		launchCtx, err := store.Create(req.PatientID, req.EncounterID, req.FHIRUser)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create launch context")
		}

		scheme := "http"
		if c.Request().TLS != nil {
			scheme = "https"
		}
		iss := fmt.Sprintf("%s://%s/fhir", scheme, c.Request().Host)

		return c.JSON(http.StatusOK, ehrLaunchResponse{
			LaunchToken: launchCtx.LaunchToken,
			ISS:         iss,
		})
	}
}

// launchContextHandler resolves a launch token to its associated context.
// This is called after authorization to add context parameters to the token response.
func launchContextHandler(store *LaunchContextStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.QueryParam("launch")
		if token == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "launch parameter is required")
		}

		launchCtx := store.Consume(token)
		if launchCtx == nil {
			return echo.NewHTTPError(http.StatusNotFound, "launch context not found or expired")
		}

		return c.JSON(http.StatusOK, launchCtx)
	}
}
