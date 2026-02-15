package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

var (
	backendTestSigningKey = []byte("test-backend-signing-key-32bytes!")
	backendTestIssuer     = "https://ehr.example.com"
	backendTestTokenURL   = "https://ehr.example.com/auth/token"
)

// testRSAKeyPair holds an RSA key pair for testing.
type testRSAKeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	KID        string
}

func generateTestKeyPair(t *testing.T) *testRSAKeyPair {
	t.Helper()
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	return &testRSAKeyPair{
		PrivateKey: privKey,
		PublicKey:  &privKey.PublicKey,
		KID:        "test-kid-1",
	}
}

func newTestBackendServiceManager(t *testing.T) *BackendServiceManager {
	t.Helper()
	return NewBackendServiceManager(
		NewInMemoryBackendServiceStore(),
		backendTestSigningKey,
		backendTestIssuer,
		backendTestTokenURL,
	)
}

func registerTestBackendClient(t *testing.T, mgr *BackendServiceManager, kp *testRSAKeyPair, scopes []string) *BackendServiceClient {
	t.Helper()
	client, err := mgr.RegisterClient(context.Background(), "Test Backend App", "https://app.example.com/.well-known/jwks.json", scopes, "tenant-1")
	if err != nil {
		t.Fatalf("failed to register backend client: %v", err)
	}
	// Inject the test public key
	client.PublicKeys = []BackendServiceKey{
		{
			KID:       kp.KID,
			Algorithm: "RS384",
			PublicKey: kp.PublicKey,
		},
	}
	// Update the client in the store
	if err := mgr.store.UpdateClient(context.Background(), client); err != nil {
		t.Fatalf("failed to update client with public key: %v", err)
	}
	return client
}

func createTestAssertion(t *testing.T, kp *testRSAKeyPair, clientID, aud, jti string, exp time.Time) string {
	t.Helper()
	claims := jwt.MapClaims{
		"iss": clientID,
		"sub": clientID,
		"aud": aud,
		"jti": jti,
		"exp": jwt.NewNumericDate(exp),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS384, claims)
	token.Header["kid"] = kp.KID
	signed, err := token.SignedString(kp.PrivateKey)
	if err != nil {
		t.Fatalf("failed to sign test assertion: %v", err)
	}
	return signed
}

// ---------------------------------------------------------------------------
// Registration Tests
// ---------------------------------------------------------------------------

func TestBackendService_RegisterClient(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	client, err := mgr.RegisterClient(context.Background(), "My Service", "https://app.example.com/.well-known/jwks.json", nil, "tenant-1")
	if err != nil {
		t.Fatalf("RegisterClient failed: %v", err)
	}
	if client.ClientID == "" {
		t.Error("expected non-empty ClientID")
	}
	if client.ClientName != "My Service" {
		t.Errorf("expected name 'My Service', got %q", client.ClientName)
	}
	if client.JWKSURL != "https://app.example.com/.well-known/jwks.json" {
		t.Errorf("expected JWKS URL set, got %q", client.JWKSURL)
	}
	if client.Status != "active" {
		t.Errorf("expected status 'active', got %q", client.Status)
	}
	if client.TenantID != "tenant-1" {
		t.Errorf("expected tenant 'tenant-1', got %q", client.TenantID)
	}
	if client.TokenLifetime != 5*time.Minute {
		t.Errorf("expected default token lifetime of 5m, got %v", client.TokenLifetime)
	}
}

func TestBackendService_RegisterClient_WithScopes(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	scopes := []string{"system/*.read", "system/Patient.read"}
	client, err := mgr.RegisterClient(context.Background(), "Scoped Service", "https://app.example.com/jwks", scopes, "tenant-1")
	if err != nil {
		t.Fatalf("RegisterClient failed: %v", err)
	}
	if len(client.Scopes) != 2 {
		t.Fatalf("expected 2 scopes, got %d", len(client.Scopes))
	}
	if client.Scopes[0] != "system/*.read" {
		t.Errorf("expected scope 'system/*.read', got %q", client.Scopes[0])
	}
	if client.Scopes[1] != "system/Patient.read" {
		t.Errorf("expected scope 'system/Patient.read', got %q", client.Scopes[1])
	}
}

func TestBackendService_GetClient(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	client, _ := mgr.RegisterClient(context.Background(), "Test Service", "https://app.example.com/jwks", nil, "tenant-1")

	retrieved, err := mgr.store.GetClient(context.Background(), client.ClientID)
	if err != nil {
		t.Fatalf("GetClient failed: %v", err)
	}
	if retrieved.ClientID != client.ClientID {
		t.Errorf("expected client ID %q, got %q", client.ClientID, retrieved.ClientID)
	}
	if retrieved.ClientName != "Test Service" {
		t.Errorf("expected name 'Test Service', got %q", retrieved.ClientName)
	}
}

func TestBackendService_ListClients(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	mgr.RegisterClient(context.Background(), "Service A", "https://a.example.com/jwks", nil, "tenant-1")
	mgr.RegisterClient(context.Background(), "Service B", "https://b.example.com/jwks", nil, "tenant-1")
	mgr.RegisterClient(context.Background(), "Service C", "https://c.example.com/jwks", nil, "tenant-2")

	clients, err := mgr.store.ListClients(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("ListClients failed: %v", err)
	}
	if len(clients) != 2 {
		t.Fatalf("expected 2 clients for tenant-1, got %d", len(clients))
	}
}

func TestBackendService_DeleteClient(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	client, _ := mgr.RegisterClient(context.Background(), "To Delete", "https://del.example.com/jwks", nil, "tenant-1")

	err := mgr.store.DeleteClient(context.Background(), client.ClientID)
	if err != nil {
		t.Fatalf("DeleteClient failed: %v", err)
	}

	_, err = mgr.store.GetClient(context.Background(), client.ClientID)
	if err == nil {
		t.Error("expected error for deleted client, got nil")
	}
}

// ---------------------------------------------------------------------------
// Authentication Tests (JWT Assertion)
// ---------------------------------------------------------------------------

func TestBackendService_AuthenticateClient_ValidAssertion(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	client := registerTestBackendClient(t, mgr, kp, []string{"system/*.read"})

	assertion := createTestAssertion(t, kp, client.ClientID, backendTestTokenURL, "unique-jti-1", time.Now().Add(2*time.Minute))

	authenticated, err := mgr.AuthenticateClient(context.Background(), assertion)
	if err != nil {
		t.Fatalf("AuthenticateClient failed: %v", err)
	}
	if authenticated.ClientID != client.ClientID {
		t.Errorf("expected client ID %q, got %q", client.ClientID, authenticated.ClientID)
	}
}

func TestBackendService_AuthenticateClient_InvalidSignature(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	client := registerTestBackendClient(t, mgr, kp, []string{"system/*.read"})

	// Create assertion with a DIFFERENT key
	wrongKP := generateTestKeyPair(t)
	wrongKP.KID = kp.KID // same kid but different key
	assertion := createTestAssertion(t, wrongKP, client.ClientID, backendTestTokenURL, "jti-wrong-sig", time.Now().Add(2*time.Minute))

	_, err := mgr.AuthenticateClient(context.Background(), assertion)
	if err == nil {
		t.Error("expected error for invalid signature, got nil")
	}
}

func TestBackendService_AuthenticateClient_ExpiredJWT(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	client := registerTestBackendClient(t, mgr, kp, []string{"system/*.read"})

	assertion := createTestAssertion(t, kp, client.ClientID, backendTestTokenURL, "jti-expired", time.Now().Add(-1*time.Minute))

	_, err := mgr.AuthenticateClient(context.Background(), assertion)
	if err == nil {
		t.Error("expected error for expired JWT, got nil")
	}
}

func TestBackendService_AuthenticateClient_WrongAudience(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	client := registerTestBackendClient(t, mgr, kp, []string{"system/*.read"})

	assertion := createTestAssertion(t, kp, client.ClientID, "https://wrong-server.example.com/token", "jti-wrong-aud", time.Now().Add(2*time.Minute))

	_, err := mgr.AuthenticateClient(context.Background(), assertion)
	if err == nil {
		t.Error("expected error for wrong audience, got nil")
	}
}

func TestBackendService_AuthenticateClient_WrongIssuer(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	registerTestBackendClient(t, mgr, kp, []string{"system/*.read"})

	// Use a different client ID as issuer
	assertion := createTestAssertion(t, kp, "some-unknown-client-id", backendTestTokenURL, "jti-wrong-iss", time.Now().Add(2*time.Minute))

	_, err := mgr.AuthenticateClient(context.Background(), assertion)
	if err == nil {
		t.Error("expected error for wrong issuer (unknown client), got nil")
	}
}

func TestBackendService_AuthenticateClient_MissingJTI(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	client := registerTestBackendClient(t, mgr, kp, []string{"system/*.read"})

	// Create assertion without jti
	claims := jwt.MapClaims{
		"iss": client.ClientID,
		"sub": client.ClientID,
		"aud": backendTestTokenURL,
		"exp": jwt.NewNumericDate(time.Now().Add(2 * time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS384, claims)
	token.Header["kid"] = kp.KID
	assertion, err := token.SignedString(kp.PrivateKey)
	if err != nil {
		t.Fatalf("failed to sign assertion: %v", err)
	}

	_, err = mgr.AuthenticateClient(context.Background(), assertion)
	if err == nil {
		t.Error("expected error for missing jti, got nil")
	}
}

func TestBackendService_AuthenticateClient_ReplayedJTI(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	client := registerTestBackendClient(t, mgr, kp, []string{"system/*.read"})

	assertion1 := createTestAssertion(t, kp, client.ClientID, backendTestTokenURL, "replayed-jti", time.Now().Add(2*time.Minute))

	// First call should succeed
	_, err := mgr.AuthenticateClient(context.Background(), assertion1)
	if err != nil {
		t.Fatalf("first authentication failed: %v", err)
	}

	// Second call with same jti should fail
	assertion2 := createTestAssertion(t, kp, client.ClientID, backendTestTokenURL, "replayed-jti", time.Now().Add(2*time.Minute))
	_, err = mgr.AuthenticateClient(context.Background(), assertion2)
	if err == nil {
		t.Error("expected error for replayed jti, got nil")
	}
}

func TestBackendService_AuthenticateClient_DisabledClient(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	client := registerTestBackendClient(t, mgr, kp, []string{"system/*.read"})

	// Disable the client
	client.Status = "disabled"
	mgr.store.UpdateClient(context.Background(), client)

	assertion := createTestAssertion(t, kp, client.ClientID, backendTestTokenURL, "jti-disabled", time.Now().Add(2*time.Minute))
	_, err := mgr.AuthenticateClient(context.Background(), assertion)
	if err == nil {
		t.Error("expected error for disabled client, got nil")
	}
}

func TestBackendService_AuthenticateClient_UnknownClient(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)

	assertion := createTestAssertion(t, kp, "nonexistent-client-id", backendTestTokenURL, "jti-unknown", time.Now().Add(2*time.Minute))

	_, err := mgr.AuthenticateClient(context.Background(), assertion)
	if err == nil {
		t.Error("expected error for unknown client, got nil")
	}
}

// ---------------------------------------------------------------------------
// Token Issuance Tests
// ---------------------------------------------------------------------------

func TestBackendService_IssueToken(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	client := registerTestBackendClient(t, mgr, kp, []string{"system/*.read", "system/Patient.read"})

	token, err := mgr.IssueAccessToken(context.Background(), client, []string{"system/*.read"})
	if err != nil {
		t.Fatalf("IssueAccessToken failed: %v", err)
	}
	if token.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if token.TokenType != "bearer" {
		t.Errorf("expected token type 'bearer', got %q", token.TokenType)
	}
	if token.Scope != "system/*.read" {
		t.Errorf("expected scope 'system/*.read', got %q", token.Scope)
	}
}

func TestBackendService_IssueToken_ScopeSubset(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	client := registerTestBackendClient(t, mgr, kp, []string{"system/*.read", "system/Patient.read", "system/Observation.read"})

	token, err := mgr.IssueAccessToken(context.Background(), client, []string{"system/Patient.read"})
	if err != nil {
		t.Fatalf("IssueAccessToken failed: %v", err)
	}
	if token.Scope != "system/Patient.read" {
		t.Errorf("expected scope 'system/Patient.read', got %q", token.Scope)
	}
}

func TestBackendService_IssueToken_ExcessiveScope(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	client := registerTestBackendClient(t, mgr, kp, []string{"system/Patient.read"})

	_, err := mgr.IssueAccessToken(context.Background(), client, []string{"system/Patient.read", "system/MedicationRequest.write"})
	if err == nil {
		t.Error("expected error for excessive scope, got nil")
	}
}

func TestBackendService_IssueToken_ExpiresIn(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	client := registerTestBackendClient(t, mgr, kp, []string{"system/*.read"})
	client.TokenLifetime = 10 * time.Minute
	mgr.store.UpdateClient(context.Background(), client)

	token, err := mgr.IssueAccessToken(context.Background(), client, []string{"system/*.read"})
	if err != nil {
		t.Fatalf("IssueAccessToken failed: %v", err)
	}
	if token.ExpiresIn != 600 {
		t.Errorf("expected expires_in 600, got %d", token.ExpiresIn)
	}
}

func TestBackendService_IssueToken_ContainsClientID(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	client := registerTestBackendClient(t, mgr, kp, []string{"system/*.read"})

	tokenResp, err := mgr.IssueAccessToken(context.Background(), client, []string{"system/*.read"})
	if err != nil {
		t.Fatalf("IssueAccessToken failed: %v", err)
	}

	// Parse the JWT to check sub claim
	parts := strings.SplitN(tokenResp.AccessToken, ".", 3)
	if len(parts) != 3 {
		t.Fatal("access token is not a valid JWT")
	}
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("failed to decode token payload: %v", err)
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		t.Fatalf("failed to parse token claims: %v", err)
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub != client.ClientID {
		t.Errorf("expected sub=%q, got %q", client.ClientID, sub)
	}
}

// ---------------------------------------------------------------------------
// Full Flow Tests (HandleTokenRequest)
// ---------------------------------------------------------------------------

func TestBackendService_HandleTokenRequest(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	client := registerTestBackendClient(t, mgr, kp, []string{"system/*.read"})

	assertion := createTestAssertion(t, kp, client.ClientID, backendTestTokenURL, "jti-full-flow", time.Now().Add(2*time.Minute))

	token, err := mgr.HandleTokenRequest(context.Background(), "client_credentials",
		"urn:ietf:params:oauth:client-assertion-type:jwt-bearer", assertion, "system/*.read")
	if err != nil {
		t.Fatalf("HandleTokenRequest failed: %v", err)
	}
	if token.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if token.TokenType != "bearer" {
		t.Errorf("expected token type 'bearer', got %q", token.TokenType)
	}
	if token.Scope != "system/*.read" {
		t.Errorf("expected scope 'system/*.read', got %q", token.Scope)
	}
}

func TestBackendService_HandleTokenRequest_WrongGrantType(t *testing.T) {
	mgr := newTestBackendServiceManager(t)

	_, err := mgr.HandleTokenRequest(context.Background(), "authorization_code",
		"urn:ietf:params:oauth:client-assertion-type:jwt-bearer", "some-assertion", "system/*.read")
	if err == nil {
		t.Error("expected error for wrong grant type, got nil")
	}
}

func TestBackendService_HandleTokenRequest_WrongAssertionType(t *testing.T) {
	mgr := newTestBackendServiceManager(t)

	_, err := mgr.HandleTokenRequest(context.Background(), "client_credentials",
		"urn:ietf:params:oauth:client-assertion-type:saml2-bearer", "some-assertion", "system/*.read")
	if err == nil {
		t.Error("expected error for wrong assertion type, got nil")
	}
}

func TestBackendService_HandleTokenRequest_EmptyAssertion(t *testing.T) {
	mgr := newTestBackendServiceManager(t)

	_, err := mgr.HandleTokenRequest(context.Background(), "client_credentials",
		"urn:ietf:params:oauth:client-assertion-type:jwt-bearer", "", "system/*.read")
	if err == nil {
		t.Error("expected error for empty assertion, got nil")
	}
}

// ---------------------------------------------------------------------------
// Endpoint Tests (HTTP)
// ---------------------------------------------------------------------------

func TestBackendService_RegisterEndpoint(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	e := echo.New()
	g := e.Group("/auth")
	RegisterBackendServiceEndpoints(g, mgr)

	body := `{"client_name":"My Backend","jwks_url":"https://app.example.com/jwks","scopes":["system/*.read"],"tenant_id":"tenant-1"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register-backend", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["client_id"] == nil || resp["client_id"].(string) == "" {
		t.Error("expected client_id in response")
	}
	if resp["client_name"] != "My Backend" {
		t.Errorf("expected client_name 'My Backend', got %v", resp["client_name"])
	}
}

func TestBackendService_TokenEndpoint(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	client := registerTestBackendClient(t, mgr, kp, []string{"system/*.read"})

	e := echo.New()
	g := e.Group("/auth")
	RegisterBackendServiceEndpoints(g, mgr)

	assertion := createTestAssertion(t, kp, client.ClientID, backendTestTokenURL, "jti-endpoint-test", time.Now().Add(2*time.Minute))

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	form.Set("client_assertion", assertion)
	form.Set("scope", "system/*.read")

	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var tokenResp BackendServiceToken
	if err := json.Unmarshal(rec.Body.Bytes(), &tokenResp); err != nil {
		t.Fatalf("failed to parse token response: %v", err)
	}
	if tokenResp.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if tokenResp.TokenType != "bearer" {
		t.Errorf("expected token type 'bearer', got %q", tokenResp.TokenType)
	}
}

func TestBackendService_ListEndpoint(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	mgr.RegisterClient(context.Background(), "Service A", "https://a.example.com/jwks", []string{"system/*.read"}, "tenant-1")
	mgr.RegisterClient(context.Background(), "Service B", "https://b.example.com/jwks", nil, "tenant-1")

	e := echo.New()
	g := e.Group("/auth")
	RegisterBackendServiceEndpoints(g, mgr)

	req := httptest.NewRequest(http.MethodGet, "/auth/backend-clients?tenant_id=tenant-1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var clients []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &clients); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(clients) != 2 {
		t.Errorf("expected 2 clients, got %d", len(clients))
	}
}

func TestBackendService_ConcurrentAuth(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	client := registerTestBackendClient(t, mgr, kp, []string{"system/*.read"})

	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)
	errs := make([]error, n)

	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			jti := fmt.Sprintf("concurrent-jti-%d", idx)
			assertion := createTestAssertion(t, kp, client.ClientID, backendTestTokenURL, jti, time.Now().Add(2*time.Minute))
			_, err := mgr.HandleTokenRequest(context.Background(), "client_credentials",
				"urn:ietf:params:oauth:client-assertion-type:jwt-bearer", assertion, "system/*.read")
			errs[idx] = err
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("concurrent request %d failed: %v", i, err)
		}
	}
}

// ---------------------------------------------------------------------------
// Additional edge-case test: exp too far in future
// ---------------------------------------------------------------------------

func TestBackendService_AuthenticateClient_ExpTooFar(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	client := registerTestBackendClient(t, mgr, kp, []string{"system/*.read"})

	// Exp 10 minutes in the future — beyond the 5-minute max
	assertion := createTestAssertion(t, kp, client.ClientID, backendTestTokenURL, "jti-too-far", time.Now().Add(10*time.Minute))
	_, err := mgr.AuthenticateClient(context.Background(), assertion)
	if err == nil {
		t.Error("expected error for exp too far in future, got nil")
	}
}

// TestBackendService_AuthenticateClient_SubMismatch tests that sub != iss is rejected.
func TestBackendService_AuthenticateClient_SubMismatch(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	kp := generateTestKeyPair(t)
	client := registerTestBackendClient(t, mgr, kp, []string{"system/*.read"})

	// Create assertion where sub differs from iss
	claims := jwt.MapClaims{
		"iss": client.ClientID,
		"sub": "different-subject",
		"aud": backendTestTokenURL,
		"jti": "jti-sub-mismatch",
		"exp": jwt.NewNumericDate(time.Now().Add(2 * time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS384, claims)
	token.Header["kid"] = kp.KID
	assertion, err := token.SignedString(kp.PrivateKey)
	if err != nil {
		t.Fatalf("failed to sign assertion: %v", err)
	}

	_, err = mgr.AuthenticateClient(context.Background(), assertion)
	if err == nil {
		t.Error("expected error for sub != iss, got nil")
	}
}

// TestBackendService_DeleteEndpoint tests DELETE /auth/backend-clients/:id
func TestBackendService_DeleteEndpoint(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	client, _ := mgr.RegisterClient(context.Background(), "To Delete", "https://del.example.com/jwks", nil, "tenant-1")

	e := echo.New()
	g := e.Group("/auth")
	RegisterBackendServiceEndpoints(g, mgr)

	req := httptest.NewRequest(http.MethodDelete, "/auth/backend-clients/"+client.ClientID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify client is deleted
	_, err := mgr.store.GetClient(context.Background(), client.ClientID)
	if err == nil {
		t.Error("expected error for deleted client, got nil")
	}
}

// ---------------------------------------------------------------------------
// JWKS URL validation helper tests
// ---------------------------------------------------------------------------

func TestBackendService_RegisterClient_EmptyName(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	_, err := mgr.RegisterClient(context.Background(), "", "https://app.example.com/jwks", nil, "tenant-1")
	if err == nil {
		t.Error("expected error for empty client name, got nil")
	}
}

func TestBackendService_RegisterClient_EmptyJWKSURL(t *testing.T) {
	mgr := newTestBackendServiceManager(t)
	_, err := mgr.RegisterClient(context.Background(), "Test", "", nil, "tenant-1")
	if err == nil {
		t.Error("expected error for empty JWKS URL, got nil")
	}
}

// ---------------------------------------------------------------------------
// Ensure JWKS public key JSON marshaling works
// ---------------------------------------------------------------------------

func TestBackendServiceKey_RSAPublicKeyRoundTrip(t *testing.T) {
	kp := generateTestKeyPair(t)
	key := BackendServiceKey{
		KID:       kp.KID,
		Algorithm: "RS384",
		PublicKey: kp.PublicKey,
	}

	// Verify key can be used for signature verification
	claims := jwt.MapClaims{"test": "value"}
	token := jwt.NewWithClaims(jwt.SigningMethodRS384, claims)
	token.Header["kid"] = kp.KID
	signed, err := token.SignedString(kp.PrivateKey)
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}

	parsed, err := jwt.Parse(signed, func(t *jwt.Token) (interface{}, error) {
		return key.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("failed to verify: %v", err)
	}
	if !parsed.Valid {
		t.Error("expected valid token")
	}
}

// Test for JTI cleanup
func TestBackendService_JTICleanup(t *testing.T) {
	mgr := newTestBackendServiceManager(t)

	// Manually add an expired JTI entry
	mgr.jtiMu.Lock()
	mgr.jtiCache["old-jti"] = time.Now().Add(-1 * time.Minute)
	mgr.jtiCache["future-jti"] = time.Now().Add(5 * time.Minute)
	mgr.jtiMu.Unlock()

	mgr.cleanupJTI()

	mgr.jtiMu.RLock()
	defer mgr.jtiMu.RUnlock()
	if _, ok := mgr.jtiCache["old-jti"]; ok {
		t.Error("expected old-jti to be cleaned up")
	}
	if _, ok := mgr.jtiCache["future-jti"]; !ok {
		t.Error("expected future-jti to still be present")
	}
}

// Unused helper to satisfy go vet if big is imported for JWKS tests.
var _ = new(big.Int)
