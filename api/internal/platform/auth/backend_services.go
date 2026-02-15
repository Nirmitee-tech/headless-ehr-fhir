package auth

import (
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Data Structures
// ---------------------------------------------------------------------------

// BackendServiceClient represents a registered SMART Backend Services client
// that authenticates via client_credentials with a JWT assertion (per SMART
// App Launch v2.0, §5 — Backend Services).
type BackendServiceClient struct {
	ClientID      string              `json:"client_id"`
	ClientName    string              `json:"client_name"`
	JWKSURL       string              `json:"jwks_url"`
	PublicKeys    []BackendServiceKey `json:"-"`
	Scopes        []string            `json:"scopes"`
	TenantID      string              `json:"tenant_id"`
	Status        string              `json:"status"`
	CreatedAt     time.Time           `json:"created_at"`
	TokenLifetime time.Duration       `json:"token_lifetime"`
}

// BackendServiceKey holds a single public key for verifying client assertions.
type BackendServiceKey struct {
	KID       string      `json:"kid"`
	Algorithm string      `json:"alg"`
	PublicKey interface{} `json:"-"` // crypto.PublicKey — RSA or EC
}

// BackendServiceToken is the token response returned to backend service clients.
type BackendServiceToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// ---------------------------------------------------------------------------
// Store Interface
// ---------------------------------------------------------------------------

// BackendServiceStore provides persistence for backend service client
// registrations.
type BackendServiceStore interface {
	RegisterClient(ctx context.Context, client *BackendServiceClient) error
	GetClient(ctx context.Context, clientID string) (*BackendServiceClient, error)
	ListClients(ctx context.Context, tenantID string) ([]*BackendServiceClient, error)
	UpdateClient(ctx context.Context, client *BackendServiceClient) error
	DeleteClient(ctx context.Context, clientID string) error
}

// ---------------------------------------------------------------------------
// In-Memory Store
// ---------------------------------------------------------------------------

// InMemoryBackendServiceStore is a thread-safe in-memory implementation of
// BackendServiceStore, suitable for development and testing.
type InMemoryBackendServiceStore struct {
	mu      sync.RWMutex
	clients map[string]*BackendServiceClient
}

// NewInMemoryBackendServiceStore creates a new in-memory store.
func NewInMemoryBackendServiceStore() *InMemoryBackendServiceStore {
	return &InMemoryBackendServiceStore{
		clients: make(map[string]*BackendServiceClient),
	}
}

func (s *InMemoryBackendServiceStore) RegisterClient(_ context.Context, client *BackendServiceClient) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.clients[client.ClientID]; exists {
		return fmt.Errorf("client %q already registered", client.ClientID)
	}
	s.clients[client.ClientID] = client
	return nil
}

func (s *InMemoryBackendServiceStore) GetClient(_ context.Context, clientID string) (*BackendServiceClient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	client, ok := s.clients[clientID]
	if !ok {
		return nil, fmt.Errorf("backend service client %q not found", clientID)
	}
	return client, nil
}

func (s *InMemoryBackendServiceStore) ListClients(_ context.Context, tenantID string) ([]*BackendServiceClient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*BackendServiceClient
	for _, c := range s.clients {
		if c.TenantID == tenantID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (s *InMemoryBackendServiceStore) UpdateClient(_ context.Context, client *BackendServiceClient) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.clients[client.ClientID]; !exists {
		return fmt.Errorf("backend service client %q not found", client.ClientID)
	}
	s.clients[client.ClientID] = client
	return nil
}

func (s *InMemoryBackendServiceStore) DeleteClient(_ context.Context, clientID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.clients[clientID]; !exists {
		return fmt.Errorf("backend service client %q not found", clientID)
	}
	delete(s.clients, clientID)
	return nil
}

// ---------------------------------------------------------------------------
// Backend Service Manager
// ---------------------------------------------------------------------------

// BackendServiceManager orchestrates SMART Backend Services registration,
// authentication, and token issuance.
type BackendServiceManager struct {
	store      BackendServiceStore
	signingKey []byte
	issuer     string
	tokenURL   string

	// JTI replay protection
	jtiMu    sync.RWMutex
	jtiCache map[string]time.Time
}

// NewBackendServiceManager creates a new BackendServiceManager.
//
//   - store:      persistence layer for client registrations
//   - signingKey: HMAC key for signing issued access tokens
//   - issuer:     the "iss" claim in issued access tokens
//   - tokenURL:   the token endpoint URL (used as expected "aud" in client assertions)
func NewBackendServiceManager(store BackendServiceStore, signingKey []byte, issuer, tokenURL string) *BackendServiceManager {
	return &BackendServiceManager{
		store:      store,
		signingKey: signingKey,
		issuer:     issuer,
		tokenURL:   tokenURL,
		jtiCache:   make(map[string]time.Time),
	}
}

// RegisterClient creates a new backend service client registration.
func (m *BackendServiceManager) RegisterClient(ctx context.Context, name, jwksURL string, scopes []string, tenantID string) (*BackendServiceClient, error) {
	if name == "" {
		return nil, fmt.Errorf("client name is required")
	}
	if jwksURL == "" {
		return nil, fmt.Errorf("JWKS URL is required")
	}

	client := &BackendServiceClient{
		ClientID:      uuid.New().String(),
		ClientName:    name,
		JWKSURL:       jwksURL,
		Scopes:        scopes,
		TenantID:      tenantID,
		Status:        "active",
		CreatedAt:     time.Now(),
		TokenLifetime: 5 * time.Minute,
	}

	if err := m.store.RegisterClient(ctx, client); err != nil {
		return nil, fmt.Errorf("registering backend service client: %w", err)
	}
	return client, nil
}

// AuthenticateClient verifies a JWT client assertion per SMART Backend
// Services (RFC 7523 / SMART App Launch v2.0 §5.2).
//
// The assertion MUST:
//   - Be signed with RS384 using a key registered for the client
//   - Have iss == sub == client_id
//   - Have aud == this server's token endpoint URL
//   - Contain a unique jti (replay protection)
//   - Have exp present and not more than 5 minutes in the future
func (m *BackendServiceManager) AuthenticateClient(ctx context.Context, assertion string) (*BackendServiceClient, error) {
	if assertion == "" {
		return nil, fmt.Errorf("client assertion is required")
	}

	// ---------------------------------------------------------------
	// Step 1: Parse the JWT without verification to extract claims
	// ---------------------------------------------------------------
	parser := jwt.NewParser(
		jwt.WithoutClaimsValidation(),
	)
	unverified, _, err := parser.ParseUnverified(assertion, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("parsing client assertion: %w", err)
	}

	claims, ok := unverified.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid assertion claims")
	}

	// Extract iss (== client_id)
	issuer, _ := claims["iss"].(string)
	if issuer == "" {
		return nil, fmt.Errorf("assertion missing iss claim")
	}

	// Verify sub == iss
	subject, _ := claims["sub"].(string)
	if subject != issuer {
		return nil, fmt.Errorf("assertion sub (%q) must equal iss (%q)", subject, issuer)
	}

	// Verify aud
	audClaim := claims["aud"]
	if !m.verifyAudience(audClaim) {
		return nil, fmt.Errorf("assertion aud does not match token endpoint %q", m.tokenURL)
	}

	// Verify jti present
	jti, _ := claims["jti"].(string)
	if jti == "" {
		return nil, fmt.Errorf("assertion missing jti claim")
	}

	// Verify exp present and not too far in the future
	expClaim, expPresent := claims["exp"]
	if !expPresent {
		return nil, fmt.Errorf("assertion missing exp claim")
	}
	expTime, err := parseNumericDate(expClaim)
	if err != nil {
		return nil, fmt.Errorf("invalid exp claim: %w", err)
	}
	if time.Now().After(expTime) {
		return nil, fmt.Errorf("assertion has expired")
	}
	maxExp := time.Now().Add(5*time.Minute + 30*time.Second) // 30s leeway
	if expTime.After(maxExp) {
		return nil, fmt.Errorf("assertion exp is too far in the future (max 5 minutes)")
	}

	// ---------------------------------------------------------------
	// Step 2: Look up client
	// ---------------------------------------------------------------
	client, err := m.store.GetClient(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("unknown client %q: %w", issuer, err)
	}

	if client.Status != "active" {
		return nil, fmt.Errorf("client %q is %s", client.ClientID, client.Status)
	}

	// ---------------------------------------------------------------
	// Step 3: Check JTI replay
	// ---------------------------------------------------------------
	if err := m.checkAndRecordJTI(jti, expTime); err != nil {
		return nil, err
	}

	// ---------------------------------------------------------------
	// Step 4: Verify signature against client's registered public keys
	// ---------------------------------------------------------------
	kid, _ := unverified.Header["kid"].(string)
	publicKey, err := m.findPublicKey(client, kid)
	if err != nil {
		return nil, fmt.Errorf("finding public key: %w", err)
	}

	// Re-parse with full verification
	verifiedToken, err := jwt.Parse(assertion, func(t *jwt.Token) (interface{}, error) {
		// Ensure the signing method is RS384
		if t.Method.Alg() != "RS384" {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}
		return publicKey, nil
	}, jwt.WithExpirationRequired())
	if err != nil {
		return nil, fmt.Errorf("verifying assertion signature: %w", err)
	}
	if !verifiedToken.Valid {
		return nil, fmt.Errorf("assertion signature is invalid")
	}

	return client, nil
}

// IssueAccessToken creates a JWT access token for an authenticated backend
// service client.
func (m *BackendServiceManager) IssueAccessToken(_ context.Context, client *BackendServiceClient, requestedScopes []string) (*BackendServiceToken, error) {
	// Validate requested scopes are a subset of granted scopes
	grantedSet := make(map[string]bool, len(client.Scopes))
	for _, s := range client.Scopes {
		grantedSet[s] = true
	}
	for _, rs := range requestedScopes {
		if !grantedSet[rs] {
			return nil, fmt.Errorf("scope %q is not granted to client %q", rs, client.ClientID)
		}
	}

	lifetime := client.TokenLifetime
	if lifetime == 0 {
		lifetime = 5 * time.Minute
	}

	now := time.Now()
	tokenClaims := map[string]interface{}{
		"iss":   m.issuer,
		"sub":   client.ClientID,
		"scope": strings.Join(requestedScopes, " "),
		"exp":   now.Add(lifetime).Unix(),
		"iat":   now.Unix(),
		"jti":   uuid.New().String(),
	}

	accessToken, err := m.signAccessToken(tokenClaims)
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	return &BackendServiceToken{
		AccessToken: accessToken,
		TokenType:   "bearer",
		ExpiresIn:   int(lifetime.Seconds()),
		Scope:       strings.Join(requestedScopes, " "),
	}, nil
}

// HandleTokenRequest is the top-level handler for a token request using the
// client_credentials grant type with a JWT assertion.
func (m *BackendServiceManager) HandleTokenRequest(ctx context.Context, grantType, clientAssertionType, clientAssertion, scope string) (*BackendServiceToken, error) {
	if grantType != "client_credentials" {
		return nil, &OAuthError{
			Code:        "unsupported_grant_type",
			Description: "grant_type must be 'client_credentials'",
		}
	}
	if clientAssertionType != "urn:ietf:params:oauth:client-assertion-type:jwt-bearer" {
		return nil, &OAuthError{
			Code:        "invalid_request",
			Description: "client_assertion_type must be 'urn:ietf:params:oauth:client-assertion-type:jwt-bearer'",
		}
	}
	if clientAssertion == "" {
		return nil, &OAuthError{
			Code:        "invalid_request",
			Description: "client_assertion is required",
		}
	}

	client, err := m.AuthenticateClient(ctx, clientAssertion)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	var requestedScopes []string
	if scope != "" {
		requestedScopes = strings.Fields(scope)
	} else {
		// Default to all granted scopes
		requestedScopes = client.Scopes
	}

	return m.IssueAccessToken(ctx, client, requestedScopes)
}

// ---------------------------------------------------------------------------
// HTTP Endpoints
// ---------------------------------------------------------------------------

// backendServiceHandler holds the HTTP handlers for backend service endpoints.
type backendServiceHandler struct {
	mgr *BackendServiceManager
}

// RegisterBackendServiceEndpoints registers SMART Backend Services HTTP
// endpoints on the given Echo group.
//
// Endpoints:
//   - POST   /auth/token             — token endpoint (client_credentials grant)
//   - POST   /auth/register-backend  — register a new backend service client
//   - GET    /auth/backend-clients   — list registered backend service clients
//   - DELETE /auth/backend-clients/:id — delete a backend service client
func RegisterBackendServiceEndpoints(g *echo.Group, mgr *BackendServiceManager) {
	h := &backendServiceHandler{mgr: mgr}
	g.POST("/token", h.handleToken)
	g.POST("/register-backend", h.handleRegister)
	g.GET("/backend-clients", h.handleList)
	g.DELETE("/backend-clients/:id", h.handleDelete)
}

// registerBackendRequest is the JSON body for client registration.
type registerBackendRequest struct {
	ClientName string   `json:"client_name"`
	JWKSURL    string   `json:"jwks_url"`
	Scopes     []string `json:"scopes"`
	TenantID   string   `json:"tenant_id"`
}

func (h *backendServiceHandler) handleToken(c echo.Context) error {
	grantType := c.FormValue("grant_type")
	if grantType != "client_credentials" {
		return c.JSON(http.StatusBadRequest, &OAuthError{
			Code:        "unsupported_grant_type",
			Description: "this endpoint handles grant_type=client_credentials only",
		})
	}

	token, err := h.mgr.HandleTokenRequest(
		c.Request().Context(),
		grantType,
		c.FormValue("client_assertion_type"),
		c.FormValue("client_assertion"),
		c.FormValue("scope"),
	)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, &OAuthError{
			Code:        "invalid_client",
			Description: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, token)
}

func (h *backendServiceHandler) handleRegister(c echo.Context) error {
	var req registerBackendRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, &OAuthError{
			Code:        "invalid_request",
			Description: "invalid request body",
		})
	}

	client, err := h.mgr.RegisterClient(c.Request().Context(), req.ClientName, req.JWKSURL, req.Scopes, req.TenantID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &OAuthError{
			Code:        "invalid_request",
			Description: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, client)
}

func (h *backendServiceHandler) handleList(c echo.Context) error {
	tenantID := c.QueryParam("tenant_id")
	clients, err := h.mgr.store.ListClients(c.Request().Context(), tenantID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &OAuthError{
			Code:        "server_error",
			Description: err.Error(),
		})
	}
	if clients == nil {
		clients = []*BackendServiceClient{}
	}
	return c.JSON(http.StatusOK, clients)
}

func (h *backendServiceHandler) handleDelete(c echo.Context) error {
	clientID := c.Param("id")
	if err := h.mgr.store.DeleteClient(c.Request().Context(), clientID); err != nil {
		return c.JSON(http.StatusNotFound, &OAuthError{
			Code:        "not_found",
			Description: err.Error(),
		})
	}
	return c.NoContent(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// JTI Replay Protection
// ---------------------------------------------------------------------------

// checkAndRecordJTI verifies a JTI has not been used before and records it.
func (m *BackendServiceManager) checkAndRecordJTI(jti string, exp time.Time) error {
	m.jtiMu.Lock()
	defer m.jtiMu.Unlock()

	if _, exists := m.jtiCache[jti]; exists {
		return fmt.Errorf("jti %q has already been used (replay detected)", jti)
	}
	m.jtiCache[jti] = exp
	return nil
}

// cleanupJTI removes expired JTI entries from the cache.
func (m *BackendServiceManager) cleanupJTI() {
	m.jtiMu.Lock()
	defer m.jtiMu.Unlock()

	now := time.Now()
	for jti, exp := range m.jtiCache {
		if now.After(exp) {
			delete(m.jtiCache, jti)
		}
	}
}

// ---------------------------------------------------------------------------
// JWT Helpers
// ---------------------------------------------------------------------------

// signAccessToken creates an HMAC-SHA256 signed JWT for the issued access token.
// This matches the signing approach used in SMARTServer.signJWT.
func (m *BackendServiceManager) signAccessToken(claims map[string]interface{}) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshaling JWT header: %w", err)
	}

	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshaling JWT payload: %w", err)
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	signingInput := headerB64 + "." + payloadB64

	mac := hmac.New(sha256.New, m.signingKey)
	mac.Write([]byte(signingInput))
	signature := mac.Sum(nil)

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + signatureB64, nil
}

// findPublicKey locates a public key from the client's registered keys by KID.
func (m *BackendServiceManager) findPublicKey(client *BackendServiceClient, kid string) (crypto.PublicKey, error) {
	if len(client.PublicKeys) == 0 {
		return nil, fmt.Errorf("client %q has no registered public keys", client.ClientID)
	}

	// If kid is provided, find exact match
	if kid != "" {
		for _, k := range client.PublicKeys {
			if k.KID == kid {
				return k.PublicKey, nil
			}
		}
		return nil, fmt.Errorf("no key with kid %q found for client %q", kid, client.ClientID)
	}

	// If no kid, use the first key
	return client.PublicKeys[0].PublicKey, nil
}

// verifyAudience checks the aud claim matches the token endpoint URL.
// aud can be a string or an array of strings per JWT spec.
func (m *BackendServiceManager) verifyAudience(audClaim interface{}) bool {
	switch v := audClaim.(type) {
	case string:
		return v == m.tokenURL
	case []interface{}:
		for _, a := range v {
			if s, ok := a.(string); ok && s == m.tokenURL {
				return true
			}
		}
	}
	return false
}

// parseNumericDate converts a JWT numeric date (float64 or json.Number) to time.Time.
func parseNumericDate(v interface{}) (time.Time, error) {
	switch n := v.(type) {
	case float64:
		return time.Unix(int64(n), 0), nil
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(i, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unexpected exp type: %T", v)
	}
}
