package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// OIDCProvider represents an OpenID Connect provider discovered via the
// .well-known/openid-configuration endpoint.
type OIDCProvider struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	UserinfoEndpoint      string   `json:"userinfo_endpoint"`
	JWKSURI               string   `json:"jwks_uri"`
	ScopesSupported       []string `json:"scopes_supported"`
	ResponseTypesSupported []string `json:"response_types_supported"`
	GrantTypesSupported   []string `json:"grant_types_supported"`
	SubjectTypesSupported []string `json:"subject_types_supported"`
	IDTokenSigningAlgValues []string `json:"id_token_signing_alg_values_supported"`
	TokenEndpointAuthMethods []string `json:"token_endpoint_auth_methods_supported"`
}

// NewOIDCProvider fetches and parses the OpenID Connect discovery document from
// the given issuer URL. It constructs the well-known URL by appending
// /.well-known/openid-configuration to the issuer.
//
// This works with any OIDC-compliant provider including Keycloak, Auth0, Okta,
// Azure AD, and Google.
func NewOIDCProvider(issuerURL string) (*OIDCProvider, error) {
	issuerURL = strings.TrimRight(issuerURL, "/")
	discoveryURL := issuerURL + "/.well-known/openid-configuration"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(discoveryURL)
	if err != nil {
		return nil, fmt.Errorf("fetching OIDC discovery document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OIDC discovery endpoint returned status %d", resp.StatusCode)
	}

	var provider OIDCProvider
	if err := json.NewDecoder(resp.Body).Decode(&provider); err != nil {
		return nil, fmt.Errorf("decoding OIDC discovery document: %w", err)
	}

	if provider.JWKSURI == "" {
		return nil, fmt.Errorf("OIDC discovery document missing jwks_uri")
	}

	return &provider, nil
}

// JWKSKeyFunc returns a jwt.Keyfunc backed by the JWKS URI discovered from this
// provider. Keys are cached in memory with a 5-minute TTL and automatically
// refreshed when a key ID is not found (to handle key rotation).
func (p *OIDCProvider) JWKSKeyFunc() jwt.Keyfunc {
	return jwksKeyFunc(p.JWKSURI)
}

// SupportsScope returns true if the provider advertises support for the given scope.
func (p *OIDCProvider) SupportsScope(scope string) bool {
	for _, s := range p.ScopesSupported {
		if s == scope {
			return true
		}
	}
	return false
}
