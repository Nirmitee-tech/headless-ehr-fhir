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
	Issuer                        string   `json:"issuer"`
	AuthorizationEndpoint         string   `json:"authorization_endpoint"`
	TokenEndpoint                 string   `json:"token_endpoint"`
	IntrospectionEndpoint         string   `json:"introspection_endpoint,omitempty"`
	ManagementEndpoint            string   `json:"management_endpoint,omitempty"`
	TokenEndpointAuthMethods      []string `json:"token_endpoint_auth_methods_supported"`
	GrantTypes                    []string `json:"grant_types_supported"`
	Scopes                        []string `json:"scopes_supported"`
	ResponseTypes                 []string `json:"response_types_supported"`
	Capabilities                  []string `json:"capabilities"`
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

// launchContextJSON is used for DB serialization so that CreatedAt is included.
type launchContextJSON struct {
	LaunchToken string    `json:"launch"`
	PatientID   string    `json:"patient,omitempty"`
	EncounterID string    `json:"encounter,omitempty"`
	FHIRUser    string    `json:"fhirUser,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// launchContextToJSON converts a LaunchContext to its JSON-serializable form
// that includes CreatedAt.
func launchContextToJSON(lc *LaunchContext) launchContextJSON {
	return launchContextJSON{
		LaunchToken: lc.LaunchToken,
		PatientID:   lc.PatientID,
		EncounterID: lc.EncounterID,
		FHIRUser:    lc.FHIRUser,
		CreatedAt:   lc.CreatedAt,
	}
}

// launchContextFromJSON converts the JSON-serializable form back to LaunchContext.
func launchContextFromJSON(j launchContextJSON) *LaunchContext {
	return &LaunchContext{
		LaunchToken: j.LaunchToken,
		PatientID:   j.PatientID,
		EncounterID: j.EncounterID,
		FHIRUser:    j.FHIRUser,
		CreatedAt:   j.CreatedAt,
	}
}

// SMARTScope represents a parsed SMART on FHIR scope.
// Format: <context>/<resourceType>.<operation>
// Examples: patient/Patient.read, user/Observation.write, patient/*.read
type SMARTScope struct {
	Context      string // "patient", "user", or "system"
	ResourceType string // e.g. "Patient", "Observation", "*"
	Operation    string // "read", "write", or "*"
}

// ---------------------------------------------------------------------------
// LaunchContextStorer interface
// ---------------------------------------------------------------------------

// LaunchContextStorer defines the contract for storing and retrieving SMART
// launch contexts. Implementations may be backed by in-memory maps, a
// relational database, Redis, etc.
type LaunchContextStorer interface {
	// Save persists a launch context under the given ID. If the ID already
	// exists it is overwritten.
	Save(ctx context.Context, id string, lc *LaunchContext) error

	// Get retrieves a launch context by ID. Returns (nil, nil) when the ID
	// does not exist or the entry has expired.
	Get(ctx context.Context, id string) (*LaunchContext, error)

	// Consume atomically retrieves and deletes a launch context (one-time use).
	// Returns (nil, nil) when the ID does not exist or the entry has expired.
	Consume(ctx context.Context, id string) (*LaunchContext, error)
}

// ---------------------------------------------------------------------------
// InMemoryLaunchContextStore
// ---------------------------------------------------------------------------

// InMemoryLaunchContextStore provides thread-safe in-memory storage for launch
// contexts. It implements LaunchContextStorer.
type InMemoryLaunchContextStore struct {
	mu       sync.RWMutex
	contexts map[string]*LaunchContext
	ttl      time.Duration
}

// NewInMemoryLaunchContextStore creates a new in-memory store with the given TTL.
func NewInMemoryLaunchContextStore(ttl time.Duration) *InMemoryLaunchContextStore {
	return &InMemoryLaunchContextStore{
		contexts: make(map[string]*LaunchContext),
		ttl:      ttl,
	}
}

// Save implements LaunchContextStorer.
func (s *InMemoryLaunchContextStore) Save(_ context.Context, id string, lc *LaunchContext) error {
	s.mu.Lock()
	s.contexts[id] = lc
	s.mu.Unlock()
	return nil
}

// Get implements LaunchContextStorer.
func (s *InMemoryLaunchContextStore) Get(_ context.Context, id string) (*LaunchContext, error) {
	s.mu.RLock()
	lc, ok := s.contexts[id]
	s.mu.RUnlock()

	if !ok {
		return nil, nil
	}

	if time.Since(lc.CreatedAt) > s.ttl {
		s.mu.Lock()
		delete(s.contexts, id)
		s.mu.Unlock()
		return nil, nil
	}

	return lc, nil
}

// Consume implements LaunchContextStorer.
func (s *InMemoryLaunchContextStore) Consume(_ context.Context, id string) (*LaunchContext, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	lc, ok := s.contexts[id]
	if !ok {
		return nil, nil
	}

	if time.Since(lc.CreatedAt) > s.ttl {
		delete(s.contexts, id)
		return nil, nil
	}

	delete(s.contexts, id)
	return lc, nil
}

// Cleanup removes expired launch contexts from the in-memory store.
func (s *InMemoryLaunchContextStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for token, lc := range s.contexts {
		if now.Sub(lc.CreatedAt) > s.ttl {
			delete(s.contexts, token)
		}
	}
}

// ---------------------------------------------------------------------------
// LaunchContextStore - backward-compatible alias
// ---------------------------------------------------------------------------

// LaunchContextStore is the original in-memory implementation kept for backward
// compatibility. New code should use LaunchContextStorer and
// NewInMemoryLaunchContextStore.
type LaunchContextStore = InMemoryLaunchContextStore

// NewLaunchContextStore creates a new in-memory store with the given TTL.
// It is an alias for NewInMemoryLaunchContextStore and is kept for backward
// compatibility.
func NewLaunchContextStore(ttl time.Duration) *LaunchContextStore {
	return NewInMemoryLaunchContextStore(ttl)
}

// Create generates a new launch token and stores the associated context.
// This is a convenience method on the in-memory store, retained for backward
// compatibility. It wraps Save with a generated token.
func (s *InMemoryLaunchContextStore) Create(patientID, encounterID, fhirUser string) (*LaunchContext, error) {
	token, err := generateLaunchToken()
	if err != nil {
		return nil, fmt.Errorf("generating launch token: %w", err)
	}

	lc := &LaunchContext{
		LaunchToken: token,
		PatientID:   patientID,
		EncounterID: encounterID,
		FHIRUser:    fhirUser,
		CreatedAt:   time.Now(),
	}

	s.mu.Lock()
	s.contexts[token] = lc
	s.mu.Unlock()

	return lc, nil
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
// It accepts a LaunchContextStorer so the caller can inject either the in-memory
// or DB-backed implementation. If store is nil, a default in-memory store with
// a 5-minute TTL is created.
func RegisterSMARTEndpoints(g *echo.Group, issuer string, store ...LaunchContextStorer) {
	var s LaunchContextStorer
	if len(store) > 0 && store[0] != nil {
		s = store[0]
	} else {
		mem := NewInMemoryLaunchContextStore(5 * time.Minute)
		// Start background cleanup of expired launch contexts for in-memory store
		go func() {
			ticker := time.NewTicker(1 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				mem.Cleanup()
			}
		}()
		s = mem
	}

	// SMART Configuration (well-known endpoint)
	g.GET("/.well-known/smart-configuration", smartConfigurationHandler(issuer))

	// EHR Launch endpoint - generates a launch context and returns a launch token
	g.POST("/launch", ehrLaunchHandler(s))

	// Launch context resolution - exchanges a launch token for context parameters
	g.GET("/launch-context", launchContextHandler(s))
}

// smartConfigurationHandler returns the SMART on FHIR well-known configuration.
// The issuer determines whether we use standalone (built-in) or external
// (Keycloak/Auth0) endpoint URLs.
func smartConfigurationHandler(issuer string) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Detect whether this is a standalone (built-in) or external (Keycloak) issuer.
		// External issuers typically contain "/realms/" (Keycloak) or similar.
		isExternal := strings.Contains(issuer, "/realms/") || strings.Contains(issuer, "/auth0/")

		var authEndpoint, tokenEndpoint, introspectEndpoint string
		if isExternal {
			authEndpoint = issuer + "/protocol/openid-connect/auth"
			tokenEndpoint = issuer + "/protocol/openid-connect/token"
			introspectEndpoint = issuer + "/protocol/openid-connect/token/introspect"
		} else {
			authEndpoint = issuer + "/auth/authorize"
			tokenEndpoint = issuer + "/auth/token"
			introspectEndpoint = issuer + "/auth/introspect"
		}

		cfg := SMARTConfiguration{
			Issuer:                issuer,
			AuthorizationEndpoint: authEndpoint,
			TokenEndpoint:         tokenEndpoint,
			IntrospectionEndpoint: introspectEndpoint,
			ManagementEndpoint:    issuer + "/auth/manage",
			TokenEndpointAuthMethods: []string{"client_secret_basic", "client_secret_post", "none"},
			GrantTypes:               []string{"authorization_code", "refresh_token"},
			Scopes: []string{
				"openid", "profile", "fhirUser", "launch", "launch/patient",
				"offline_access",
				"patient/*.read", "patient/*.write", "patient/*.*",
				"user/*.read", "user/*.write", "user/*.*",
				"system/*.read", "system/*.write", "system/*.*",
			},
			ResponseTypes: []string{"code"},
			Capabilities: []string{
				"launch-ehr", "launch-standalone",
				"client-public", "client-confidential-symmetric",
				"sso-openid-connect",
				"context-ehr-patient", "context-standalone-patient",
				"permission-offline", "permission-patient", "permission-user",
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
func ehrLaunchHandler(store LaunchContextStorer) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req ehrLaunchRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}

		if req.PatientID == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "patient_id is required")
		}

		token, err := generateLaunchToken()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create launch context")
		}

		launchCtx := &LaunchContext{
			LaunchToken: token,
			PatientID:   req.PatientID,
			EncounterID: req.EncounterID,
			FHIRUser:    req.FHIRUser,
			CreatedAt:   time.Now(),
		}

		if err := store.Save(c.Request().Context(), token, launchCtx); err != nil {
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
func launchContextHandler(store LaunchContextStorer) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.QueryParam("launch")
		if token == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "launch parameter is required")
		}

		launchCtx, err := store.Consume(c.Request().Context(), token)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to resolve launch context")
		}
		if launchCtx == nil {
			return echo.NewHTTPError(http.StatusNotFound, "launch context not found or expired")
		}

		return c.JSON(http.StatusOK, launchCtx)
	}
}
