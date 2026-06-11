package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// StandaloneAuthMiddleware validates JWTs issued by the built-in SMART on FHIR
// server. Unlike JWTMiddleware (which expects external JWKS), this middleware
// uses the same HMAC signing key that the SMART server uses and bridges the
// SMART token claim format (single "scope" string) to the context values that
// downstream middleware (RBAC, ABAC, FHIRScopeMiddleware) expects.
//
// Auth modes:
//   - "development"  → DevAuthMiddleware (no auth, admin access)
//   - "standalone"   → this middleware (built-in SMART server)
//   - "external"     → JWTMiddleware (Keycloak, Auth0, etc.)
func StandaloneAuthMiddleware(smartServer *SMARTServer, defaultTenant string, skipper func(echo.Context) bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip for health/metrics/discovery endpoints.
			if skipper != nil && skipper(c) {
				return next(c)
			}

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization format")
			}

			tokenStr := parts[1]

			// Introspect the token using the built-in SMART server.
			claims, err := smartServer.IntrospectToken(tokenStr)
			if err != nil || !claims.Active {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			// Bridge SMART claims to the context format expected by downstream middleware.
			tenantID := defaultTenant
			userID := claims.Subject
			if userID == "" {
				userID = claims.ClientID
			}

			// Convert SMART scope string to []string for FHIRScopeMiddleware.
			var fhirScopes []string
			if claims.Scope != "" {
				fhirScopes = strings.Fields(claims.Scope)
			}

			// Infer roles from scopes for RBAC compatibility.
			roles := inferRolesFromScopes(fhirScopes)

			// Set tenant on echo context for TenantMiddleware.
			c.Set("jwt_tenant_id", tenantID)

			// Set values on request context for downstream middleware.
			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, UserIDKey, userID)
			ctx = context.WithValue(ctx, UserRolesKey, roles)
			ctx = context.WithValue(ctx, UserScopesKey, fhirScopes)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// inferRolesFromScopes determines RBAC roles from SMART scopes. This bridges
// scope-based auth (SMART) to role-based auth (RBAC) for handlers that use
// auth.RequireRole().
func inferRolesFromScopes(scopes []string) []string {
	roles := make(map[string]bool)

	for _, s := range scopes {
		// user/*.*  or system/*.* → admin
		if s == "user/*.*" || s == "system/*.*" {
			roles["admin"] = true
			continue
		}

		// user/<Resource>.read or user/<Resource>.write → physician
		if strings.HasPrefix(s, "user/") {
			roles["physician"] = true
		}

		// patient/<Resource>.read → patient
		if strings.HasPrefix(s, "patient/") {
			roles["patient"] = true
		}

		// Specific resource scopes → infer specialty roles
		if strings.Contains(s, "MedicationRequest") || strings.Contains(s, "MedicationDispense") {
			roles["pharmacist"] = true
		}
		if strings.Contains(s, "DiagnosticReport") || strings.Contains(s, "Observation") {
			roles["lab_tech"] = true
		}
	}

	// If user has broad user/*.read scope, grant common clinical roles.
	for _, s := range scopes {
		if s == "user/*.read" || s == "user/*.*" {
			roles["physician"] = true
			roles["nurse"] = true
			roles["pharmacist"] = true
			roles["lab_tech"] = true
			roles["registrar"] = true
		}
	}

	result := make([]string, 0, len(roles))
	for r := range roles {
		result = append(result, r)
	}
	return result
}

// RegisterDefaultSMARTClient registers a default Inferno-compatible SMART
// client for testing. This client supports both EHR launch and standalone
// launch flows with PKCE.
func RegisterDefaultSMARTClient(server *SMARTServer) {
	// Set default patient for standalone launch (matches seed_inferno.sql).
	server.SetDefaultPatientID("patient-john-smith")

	// Inferno test client — public client with PKCE
	infernoClient := &SMARTClient{
		ClientID:     "inferno-test",
		RedirectURIs: []string{
			"http://localhost:4567/inferno/callback",
			"http://localhost:4567/custom/smart/redirect",
			"http://localhost:4567/inferno/oauth2/static/redirect",
			"http://localhost/inferno/callback",
			"http://localhost/custom/smart/redirect",
			"http://localhost/inferno/oauth2/static/redirect",
		},
		Scope:        "launch launch/patient openid fhirUser offline_access patient/*.read patient/*.rs user/*.read user/*.write user/*.rs",
		Name:         "Inferno Test Client",
		IsPublic:     true,
	}
	if err := server.RegisterClient(infernoClient); err != nil {
		log.Printf("SMART: default inferno client already registered: %v", err)
	}

	// Generic test client — confidential client for development + Inferno tests
	testClient := &SMARTClient{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURIs: []string{
			"http://localhost:3000/callback",
			"http://localhost:8080/callback",
			"http://localhost:4567/custom/smart/redirect",
			"http://localhost:4567/inferno/oauth2/static/redirect",
			"http://localhost/custom/smart/redirect",
			"http://localhost/inferno/oauth2/static/redirect",
		},
		Scope:        "launch launch/patient openid fhirUser offline_access user/*.* patient/*.read patient/*.rs user/*.rs",
		Name:         "Test Client",
		IsPublic:     false,
	}
	if err := server.RegisterClient(testClient); err != nil {
		log.Printf("SMART: default test client already registered: %v", err)
	}

	// Asymmetric confidential client (SMART private_key_jwt) — registered with
	// the Inferno test framework's published public JWKS so the ONC g10 kit's
	// asymmetric scenarios can authenticate. Production deployments register
	// their own clients with their own key sets.
	asymmetricClient := &SMARTClient{
		ClientID: "test-client-asymmetric",
		RedirectURIs: []string{
			"http://localhost:4567/custom/smart/redirect",
			"http://localhost:4567/inferno/oauth2/static/redirect",
			"http://localhost/custom/smart/redirect",
			"http://localhost/inferno/oauth2/static/redirect",
		},
		Scope:    "launch launch/patient openid fhirUser offline_access patient/*.read patient/*.rs user/*.read user/*.rs",
		Name:     "Asymmetric Test Client (private_key_jwt)",
		IsPublic: false,
		JWKS:     infernoPublicJWKS,
	}
	if err := server.RegisterClient(asymmetricClient); err != nil {
		log.Printf("SMART: asymmetric test client already registered: %v", err)
	}
}

// infernoPublicJWKS is the Inferno test framework's published public key set
// (RS384 + ES384), used by its SMART asymmetric-auth test clients.
const infernoPublicJWKS = `{"keys": [{"kty": "EC", "crv": "P-384", "x": "JQKTsV6PT5Szf4QtDA1qrs0EJ1pbimQmM2SKvzOlIAqlph3h1OHmZ2i7MXahIF2C", "y": "bRWWQRJBgDa6CTgwofYrHjVGcO-A7WNEnu4oJA5OUJPPPpczgx1g2NsfinK-D2Rw", "use": "sig", "key_ops": ["verify"], "ext": true, "kid": "4b49a739d1eb115b3225f4cf9beb6d1b", "alg": "ES384"}, {"kty": "RSA", "alg": "RS384", "n": "vjbIzTqiY8K8zApeNng5ekNNIxJfXAue9BjoMrZ9Qy9m7yIA-tf6muEupEXWhq70tC7vIGLqJJ4O8m7yiH8H2qklX2mCAMg3xG3nbykY2X7JXtW9P8VIdG0sAMt5aZQnUGCgSS3n0qaooGn2LUlTGIR88Qi-4Nrao9_3Ki3UCiICeCiAE224jGCg0OlQU6qj2gEB3o-DWJFlG_dz1y-Mxo5ivaeM0vWuodjDrp-aiabJcSF_dx26sdC9dZdBKXFDq0t19I9S9AyGpGDJwzGRtWHY6LsskNHLvo8Zb5AsJ9eRZKpnh30SYBZI9WHtzU85M9WQqdScR69Vyp-6Uhfbvw", "e": "AQAB", "use": "sig", "key_ops": ["verify"], "ext": true, "kid": "b41528b6f37a9500edb8a905a595bdd7"}]}`

// PrintStandaloneAuthInfo logs the standalone auth configuration for developer reference.
func PrintStandaloneAuthInfo(issuer, port string) {
	info := fmt.Sprintf(`
============================================================
  AUTH MODE: standalone (built-in SMART on FHIR server)
============================================================
  SMART Discovery:  http://localhost:%s/fhir/.well-known/smart-configuration
  Authorize:        http://localhost:%s/auth/authorize
  Token:            http://localhost:%s/auth/token
  Introspect:       http://localhost:%s/auth/introspect

  Pre-registered clients:
    - inferno-test  (public, PKCE)    → for Inferno g(10) testing
    - test-client   (secret: test-secret) → for general development

  Quick test:
    curl -X POST http://localhost:%s/auth/token \
      -d "grant_type=authorization_code&client_id=test-client&client_secret=test-secret&code=<code>"
============================================================
`, port, port, port, port, port)

	lines := strings.Split(info, "\n")
	for _, line := range lines {
		if line != "" {
			log.Println(line)
		}
	}
}

// SmartWellKnownJSON returns the .well-known/smart-configuration JSON for
// standalone mode, using the server's own URLs.
func SmartWellKnownJSON(issuer string) []byte {
	config := map[string]interface{}{
		"issuer":                                issuer,
		"authorization_endpoint":                issuer + "/auth/authorize",
		"token_endpoint":                        issuer + "/auth/token",
		"introspection_endpoint":                issuer + "/auth/introspect",
		"management_endpoint":                   issuer + "/auth/manage",
		"jwks_uri":                              issuer + "/auth/jwks",
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic", "private_key_jwt"},
		"token_endpoint_auth_signing_alg_values_supported": []string{"RS384", "ES384"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"scopes_supported": []string{
			"openid", "fhirUser", "launch", "launch/patient",
			"offline_access",
			"patient/*.read", "patient/*.write", "patient/*.*",
			"user/*.read", "user/*.write", "user/*.*",
			"system/*.read", "system/*.write", "system/*.*",
			"patient/*.rs", "patient/*.cruds", "user/*.rs", "user/*.cruds",
		},
		"response_types_supported":                   []string{"code"},
		"capabilities":                               []string{"launch-ehr", "launch-standalone", "client-public", "client-confidential-symmetric", "client-confidential-asymmetric", "sso-openid-connect", "authorize-post", "context-ehr-patient", "context-standalone-patient", "permission-offline", "permission-patient", "permission-user", "permission-v1", "permission-v2"},
		"code_challenge_methods_supported":            []string{"S256"},
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	return data
}
