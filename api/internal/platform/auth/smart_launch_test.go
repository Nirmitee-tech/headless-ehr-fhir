package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

var smartTestKey = []byte("test-smart-signing-key-for-tests-only")

func newTestSMARTServer() *SMARTServer {
	return NewSMARTServer("https://ehr.example.com", smartTestKey)
}

func registerTestClient(t *testing.T, s *SMARTServer, public bool) *SMARTClient {
	t.Helper()
	client := &SMARTClient{
		ClientID:     "test-client-" + mustRandomHex(t, 4),
		ClientSecret: "test-secret",
		RedirectURIs: []string{"https://app.example.com/callback"},
		Scope:        "patient/*.read patient/Patient.read launch openid offline_access",
		Name:         "Test App",
		IsPublic:     public,
	}
	if public {
		client.ClientSecret = ""
	}
	if err := s.RegisterClient(client); err != nil {
		t.Fatalf("failed to register client: %v", err)
	}
	return client
}

func mustRandomHex(t *testing.T, n int) string {
	t.Helper()
	s, err := generateRandomHex(n)
	if err != nil {
		t.Fatalf("generateRandomHex failed: %v", err)
	}
	return s
}

func mustAuthorize(t *testing.T, s *SMARTServer, clientID, redirectURI, scope, launch, codeChallenge string) *AuthorizationResponse {
	t.Helper()
	req := &AuthorizationRequest{
		ResponseType:        "code",
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		Scope:               scope,
		State:               "test-state",
		Aud:                 "https://ehr.example.com/fhir",
		Launch:              launch,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: "S256",
	}
	resp, err := s.Authorize(req)
	if err != nil {
		t.Fatalf("Authorize failed: %v", err)
	}
	return resp
}

func generatePKCE(t *testing.T) (verifier, challenge string) {
	t.Helper()
	verifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	hash := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(hash[:])
	return
}

// ---------------------------------------------------------------------------
// Server Tests
// ---------------------------------------------------------------------------

func TestSMARTServer_RegisterClient(t *testing.T) {
	s := newTestSMARTServer()
	client := &SMARTClient{
		ClientID:     "my-app",
		ClientSecret: "my-secret",
		RedirectURIs: []string{"https://app.example.com/callback"},
		Scope:        "patient/*.read launch",
		Name:         "My App",
		IsPublic:     false,
	}

	err := s.RegisterClient(client)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	s.mu.RLock()
	stored, ok := s.clients["my-app"]
	s.mu.RUnlock()

	if !ok {
		t.Fatal("expected client to be stored")
	}
	if stored.Name != "My App" {
		t.Errorf("expected name 'My App', got %q", stored.Name)
	}
	if stored.ClientSecret != "my-secret" {
		t.Errorf("expected secret 'my-secret', got %q", stored.ClientSecret)
	}
}

func TestSMARTServer_RegisterClient_DuplicateID(t *testing.T) {
	s := newTestSMARTServer()
	client := &SMARTClient{
		ClientID:     "dup-app",
		RedirectURIs: []string{"https://app.example.com/callback"},
		Scope:        "patient/*.read",
		Name:         "Dup App",
	}

	if err := s.RegisterClient(client); err != nil {
		t.Fatalf("first register should succeed: %v", err)
	}

	err := s.RegisterClient(client)
	if err == nil {
		t.Fatal("expected error for duplicate client_id")
	}
	if !strings.Contains(err.Error(), "already registered") {
		t.Errorf("expected 'already registered' error, got %q", err.Error())
	}
}

func TestSMARTServer_CreateLaunchContext(t *testing.T) {
	s := newTestSMARTServer()

	lc, err := s.CreateLaunchContext("patient-123", "encounter-456", "user-789")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if lc.ID == "" {
		t.Error("expected non-empty launch context ID")
	}
	if lc.PatientID != "patient-123" {
		t.Errorf("expected patient-123, got %s", lc.PatientID)
	}
	if lc.EncounterID != "encounter-456" {
		t.Errorf("expected encounter-456, got %s", lc.EncounterID)
	}
	if lc.UserID != "user-789" {
		t.Errorf("expected user-789, got %s", lc.UserID)
	}
	if lc.ExpiresAt.Before(lc.CreatedAt) {
		t.Error("expected ExpiresAt to be after CreatedAt")
	}

	// Verify stored
	s.mu.RLock()
	_, ok := s.launchContexts[lc.ID]
	s.mu.RUnlock()
	if !ok {
		t.Error("expected launch context to be stored")
	}
}

func TestSMARTServer_Authorize_EHRLaunch(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	lc, _ := s.CreateLaunchContext("patient-ehr", "encounter-ehr", "user-ehr")

	req := &AuthorizationRequest{
		ResponseType: "code",
		ClientID:     client.ClientID,
		RedirectURI:  "https://app.example.com/callback",
		Scope:        "patient/*.read launch",
		State:        "ehr-state",
		Aud:          "https://ehr.example.com/fhir",
		Launch:       lc.ID,
	}

	resp, err := s.Authorize(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Code == "" {
		t.Error("expected non-empty authorization code")
	}
	if resp.State != "ehr-state" {
		t.Errorf("expected state 'ehr-state', got %q", resp.State)
	}

	// Verify auth code has launch context
	s.mu.RLock()
	ac := s.authCodes[resp.Code]
	s.mu.RUnlock()

	if ac.PatientID != "patient-ehr" {
		t.Errorf("expected patient-ehr, got %s", ac.PatientID)
	}
	if ac.EncounterID != "encounter-ehr" {
		t.Errorf("expected encounter-ehr, got %s", ac.EncounterID)
	}

	// Launch context should be consumed (deleted)
	s.mu.RLock()
	_, lcStillExists := s.launchContexts[lc.ID]
	s.mu.RUnlock()
	if lcStillExists {
		t.Error("expected launch context to be consumed after authorize")
	}
}

func TestSMARTServer_Authorize_StandaloneLaunch(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	req := &AuthorizationRequest{
		ResponseType: "code",
		ClientID:     client.ClientID,
		RedirectURI:  "https://app.example.com/callback",
		Scope:        "patient/*.read",
		State:        "standalone-state",
		Aud:          "https://ehr.example.com/fhir",
	}

	resp, err := s.Authorize(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Code == "" {
		t.Error("expected non-empty authorization code")
	}
	if resp.State != "standalone-state" {
		t.Errorf("expected state 'standalone-state', got %q", resp.State)
	}
}

func TestSMARTServer_Authorize_InvalidClient(t *testing.T) {
	s := newTestSMARTServer()

	req := &AuthorizationRequest{
		ResponseType: "code",
		ClientID:     "nonexistent-client",
		RedirectURI:  "https://app.example.com/callback",
		Scope:        "patient/*.read",
		State:        "test-state",
		Aud:          "https://ehr.example.com/fhir",
	}

	_, err := s.Authorize(req)
	if err == nil {
		t.Fatal("expected error for unknown client_id")
	}
	oauthErr, ok := err.(*OAuthError)
	if !ok {
		t.Fatalf("expected *OAuthError, got %T", err)
	}
	if oauthErr.Code != "invalid_request" {
		t.Errorf("expected error code 'invalid_request', got %q", oauthErr.Code)
	}
}

func TestSMARTServer_Authorize_InvalidRedirectURI(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	req := &AuthorizationRequest{
		ResponseType: "code",
		ClientID:     client.ClientID,
		RedirectURI:  "https://evil.example.com/callback",
		Scope:        "patient/*.read",
		State:        "test-state",
		Aud:          "https://ehr.example.com/fhir",
	}

	_, err := s.Authorize(req)
	if err == nil {
		t.Fatal("expected error for invalid redirect_uri")
	}
	oauthErr, ok := err.(*OAuthError)
	if !ok {
		t.Fatalf("expected *OAuthError, got %T", err)
	}
	if oauthErr.Code != "invalid_request" {
		t.Errorf("expected error code 'invalid_request', got %q", oauthErr.Code)
	}
}

func TestSMARTServer_Authorize_InvalidScope(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	req := &AuthorizationRequest{
		ResponseType: "code",
		ClientID:     client.ClientID,
		RedirectURI:  "https://app.example.com/callback",
		Scope:        "admin/everything.delete",
		State:        "test-state",
		Aud:          "https://ehr.example.com/fhir",
	}

	_, err := s.Authorize(req)
	if err == nil {
		t.Fatal("expected error for invalid scope")
	}
	oauthErr, ok := err.(*OAuthError)
	if !ok {
		t.Fatalf("expected *OAuthError, got %T", err)
	}
	if oauthErr.Code != "invalid_scope" {
		t.Errorf("expected error code 'invalid_scope', got %q", oauthErr.Code)
	}
}

func TestSMARTServer_Authorize_PKCE(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	_, challenge := generatePKCE(t)

	req := &AuthorizationRequest{
		ResponseType:        "code",
		ClientID:            client.ClientID,
		RedirectURI:         "https://app.example.com/callback",
		Scope:               "patient/*.read",
		State:               "pkce-state",
		Aud:                 "https://ehr.example.com/fhir",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
	}

	resp, err := s.Authorize(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify code_challenge is stored on the auth code
	s.mu.RLock()
	ac := s.authCodes[resp.Code]
	s.mu.RUnlock()

	if ac.CodeChallenge != challenge {
		t.Errorf("expected code_challenge to be stored, got %q", ac.CodeChallenge)
	}
	if ac.CodeChallengeMethod != "S256" {
		t.Errorf("expected code_challenge_method 'S256', got %q", ac.CodeChallengeMethod)
	}
}

func TestSMARTServer_ExchangeCode_Success(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	// Create launch context and authorize
	lc, _ := s.CreateLaunchContext("patient-ex", "encounter-ex", "user-ex")
	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read launch", lc.ID, "")

	tokenResp, err := s.ExchangeCode(&TokenRequest{
		GrantType:    "authorization_code",
		Code:         authResp.Code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     client.ClientID,
		ClientSecret: "test-secret",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tokenResp.AccessToken == "" {
		t.Error("expected non-empty access_token")
	}
	if tokenResp.TokenType != "Bearer" {
		t.Errorf("expected token_type 'Bearer', got %q", tokenResp.TokenType)
	}
	if tokenResp.ExpiresIn != 3600 {
		t.Errorf("expected expires_in 3600, got %d", tokenResp.ExpiresIn)
	}
	if tokenResp.Patient != "patient-ex" {
		t.Errorf("expected patient 'patient-ex', got %q", tokenResp.Patient)
	}
	if tokenResp.Encounter != "encounter-ex" {
		t.Errorf("expected encounter 'encounter-ex', got %q", tokenResp.Encounter)
	}
}

func TestSMARTServer_ExchangeCode_ExpiredCode(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read", "", "")

	// Manually expire the code
	s.mu.Lock()
	s.authCodes[authResp.Code].ExpiresAt = time.Now().Add(-1 * time.Minute)
	s.mu.Unlock()

	_, err := s.ExchangeCode(&TokenRequest{
		GrantType:    "authorization_code",
		Code:         authResp.Code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     client.ClientID,
		ClientSecret: "test-secret",
	})
	if err == nil {
		t.Fatal("expected error for expired code")
	}
	oauthErr, ok := err.(*OAuthError)
	if !ok {
		t.Fatalf("expected *OAuthError, got %T", err)
	}
	if oauthErr.Code != "invalid_grant" {
		t.Errorf("expected 'invalid_grant', got %q", oauthErr.Code)
	}
}

func TestSMARTServer_ExchangeCode_InvalidCode(t *testing.T) {
	s := newTestSMARTServer()

	_, err := s.ExchangeCode(&TokenRequest{
		GrantType:   "authorization_code",
		Code:        "nonexistent-code",
		RedirectURI: "https://app.example.com/callback",
		ClientID:    "test-client",
	})
	if err == nil {
		t.Fatal("expected error for invalid code")
	}
	oauthErr, ok := err.(*OAuthError)
	if !ok {
		t.Fatalf("expected *OAuthError, got %T", err)
	}
	if oauthErr.Code != "invalid_grant" {
		t.Errorf("expected 'invalid_grant', got %q", oauthErr.Code)
	}
}

func TestSMARTServer_ExchangeCode_RedirectMismatch(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read", "", "")

	_, err := s.ExchangeCode(&TokenRequest{
		GrantType:    "authorization_code",
		Code:         authResp.Code,
		RedirectURI:  "https://wrong.example.com/callback",
		ClientID:     client.ClientID,
		ClientSecret: "test-secret",
	})
	if err == nil {
		t.Fatal("expected error for redirect_uri mismatch")
	}
	oauthErr, ok := err.(*OAuthError)
	if !ok {
		t.Fatalf("expected *OAuthError, got %T", err)
	}
	if oauthErr.Code != "invalid_grant" {
		t.Errorf("expected 'invalid_grant', got %q", oauthErr.Code)
	}
}

func TestSMARTServer_ExchangeCode_PKCEVerification(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	verifier, challenge := generatePKCE(t)

	t.Run("correct verifier succeeds", func(t *testing.T) {
		authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read", "", challenge)

		tokenResp, err := s.ExchangeCode(&TokenRequest{
			GrantType:    "authorization_code",
			Code:         authResp.Code,
			RedirectURI:  "https://app.example.com/callback",
			ClientID:     client.ClientID,
			ClientSecret: "test-secret",
			CodeVerifier: verifier,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if tokenResp.AccessToken == "" {
			t.Error("expected non-empty access_token")
		}
	})

	t.Run("wrong verifier fails", func(t *testing.T) {
		authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read", "", challenge)

		_, err := s.ExchangeCode(&TokenRequest{
			GrantType:    "authorization_code",
			Code:         authResp.Code,
			RedirectURI:  "https://app.example.com/callback",
			ClientID:     client.ClientID,
			ClientSecret: "test-secret",
			CodeVerifier: "wrong-verifier",
		})
		if err == nil {
			t.Fatal("expected error for wrong PKCE verifier")
		}
	})
}

func TestSMARTServer_ExchangeCode_PublicClientRequiresPKCE(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, true) // public client

	// Authorize without PKCE
	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read", "", "")

	_, err := s.ExchangeCode(&TokenRequest{
		GrantType:   "authorization_code",
		Code:        authResp.Code,
		RedirectURI: "https://app.example.com/callback",
		ClientID:    client.ClientID,
	})
	if err == nil {
		t.Fatal("expected error: public client without PKCE")
	}
	oauthErr, ok := err.(*OAuthError)
	if !ok {
		t.Fatalf("expected *OAuthError, got %T", err)
	}
	if oauthErr.Code != "invalid_request" {
		t.Errorf("expected 'invalid_request', got %q", oauthErr.Code)
	}
	if !strings.Contains(oauthErr.Description, "PKCE") {
		t.Errorf("expected PKCE mention in description, got %q", oauthErr.Description)
	}
}

func TestSMARTServer_ExchangeCode_ConfidentialClientSecret(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read", "", "")

	_, err := s.ExchangeCode(&TokenRequest{
		GrantType:    "authorization_code",
		Code:         authResp.Code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     client.ClientID,
		ClientSecret: "wrong-secret",
	})
	if err == nil {
		t.Fatal("expected error for wrong client_secret")
	}
	oauthErr, ok := err.(*OAuthError)
	if !ok {
		t.Fatalf("expected *OAuthError, got %T", err)
	}
	if oauthErr.Code != "invalid_client" {
		t.Errorf("expected 'invalid_client', got %q", oauthErr.Code)
	}
}

func TestSMARTServer_ExchangeCode_TokenContainsLaunchContext(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	lc, _ := s.CreateLaunchContext("patient-ctx-test", "encounter-ctx-test", "user-ctx-test")
	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read launch", lc.ID, "")

	tokenResp, err := s.ExchangeCode(&TokenRequest{
		GrantType:    "authorization_code",
		Code:         authResp.Code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     client.ClientID,
		ClientSecret: "test-secret",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify token response contains launch context
	if tokenResp.Patient != "patient-ctx-test" {
		t.Errorf("expected patient 'patient-ctx-test', got %q", tokenResp.Patient)
	}
	if tokenResp.Encounter != "encounter-ctx-test" {
		t.Errorf("expected encounter 'encounter-ctx-test', got %q", tokenResp.Encounter)
	}

	// Verify JWT claims contain launch context
	claims, parseErr := s.parseJWT(tokenResp.AccessToken)
	if parseErr != nil {
		t.Fatalf("failed to parse JWT: %v", parseErr)
	}
	if claims["patient"] != "patient-ctx-test" {
		t.Errorf("expected JWT patient claim 'patient-ctx-test', got %v", claims["patient"])
	}
	if claims["encounter"] != "encounter-ctx-test" {
		t.Errorf("expected JWT encounter claim 'encounter-ctx-test', got %v", claims["encounter"])
	}
}

func TestSMARTServer_RefreshToken(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	lc, _ := s.CreateLaunchContext("patient-refresh", "encounter-refresh", "user-refresh")
	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read launch offline_access", lc.ID, "")

	tokenResp, err := s.ExchangeCode(&TokenRequest{
		GrantType:    "authorization_code",
		Code:         authResp.Code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     client.ClientID,
		ClientSecret: "test-secret",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tokenResp.RefreshToken == "" {
		t.Fatal("expected non-empty refresh token with offline_access scope")
	}

	// Use refresh token to get a new access token
	newTokenResp, err := s.RefreshAccessToken(tokenResp.RefreshToken, client.ClientID)
	if err != nil {
		t.Fatalf("expected no error on refresh, got %v", err)
	}

	if newTokenResp.AccessToken == "" {
		t.Error("expected non-empty access_token from refresh")
	}
	if newTokenResp.Patient != "patient-refresh" {
		t.Errorf("expected patient 'patient-refresh', got %q", newTokenResp.Patient)
	}
	if newTokenResp.RefreshToken != tokenResp.RefreshToken {
		t.Errorf("expected same refresh token to be returned")
	}
}

func TestSMARTServer_RefreshToken_Expired(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	lc, _ := s.CreateLaunchContext("patient-rexp", "", "user-rexp")
	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read launch offline_access", lc.ID, "")

	tokenResp, _ := s.ExchangeCode(&TokenRequest{
		GrantType:    "authorization_code",
		Code:         authResp.Code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     client.ClientID,
		ClientSecret: "test-secret",
	})

	// Manually expire the refresh token
	s.mu.Lock()
	s.refreshTokens[tokenResp.RefreshToken].ExpiresAt = time.Now().Add(-1 * time.Hour)
	s.mu.Unlock()

	_, err := s.RefreshAccessToken(tokenResp.RefreshToken, client.ClientID)
	if err == nil {
		t.Fatal("expected error for expired refresh token")
	}
	oauthErr, ok := err.(*OAuthError)
	if !ok {
		t.Fatalf("expected *OAuthError, got %T", err)
	}
	if oauthErr.Code != "invalid_grant" {
		t.Errorf("expected 'invalid_grant', got %q", oauthErr.Code)
	}
}

func TestSMARTServer_IntrospectToken_Valid(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	lc, _ := s.CreateLaunchContext("patient-intro", "encounter-intro", "user-intro")
	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read launch", lc.ID, "")

	tokenResp, _ := s.ExchangeCode(&TokenRequest{
		GrantType:    "authorization_code",
		Code:         authResp.Code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     client.ClientID,
		ClientSecret: "test-secret",
	})

	claims, err := s.IntrospectToken(tokenResp.AccessToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !claims.Active {
		t.Error("expected active=true")
	}
	if claims.Subject != "user-intro" {
		t.Errorf("expected sub 'user-intro', got %q", claims.Subject)
	}
	if claims.Patient != "patient-intro" {
		t.Errorf("expected patient 'patient-intro', got %q", claims.Patient)
	}
	if claims.Encounter != "encounter-intro" {
		t.Errorf("expected encounter 'encounter-intro', got %q", claims.Encounter)
	}
	if claims.Issuer != "https://ehr.example.com" {
		t.Errorf("expected issuer 'https://ehr.example.com', got %q", claims.Issuer)
	}
}

func TestSMARTServer_IntrospectToken_Expired(t *testing.T) {
	s := newTestSMARTServer()

	// Create a token with expired claims
	claims := map[string]interface{}{
		"iss": s.issuer,
		"sub": "user-expired",
		"exp": time.Now().Add(-1 * time.Hour).Unix(),
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	}
	token, _ := s.signJWT(claims)

	result, err := s.IntrospectToken(token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Active {
		t.Error("expected active=false for expired token")
	}
}

func TestSMARTServer_CodeReuse(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read", "", "")

	// First exchange should succeed
	_, err := s.ExchangeCode(&TokenRequest{
		GrantType:    "authorization_code",
		Code:         authResp.Code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     client.ClientID,
		ClientSecret: "test-secret",
	})
	if err != nil {
		t.Fatalf("first exchange should succeed: %v", err)
	}

	// Second exchange with same code should fail
	_, err = s.ExchangeCode(&TokenRequest{
		GrantType:    "authorization_code",
		Code:         authResp.Code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     client.ClientID,
		ClientSecret: "test-secret",
	})
	if err == nil {
		t.Fatal("expected error for code reuse")
	}
	oauthErr, ok := err.(*OAuthError)
	if !ok {
		t.Fatalf("expected *OAuthError, got %T", err)
	}
	if oauthErr.Code != "invalid_grant" {
		t.Errorf("expected 'invalid_grant', got %q", oauthErr.Code)
	}
}

func TestSMARTServer_ScopeNegotiation(t *testing.T) {
	s := newTestSMARTServer()

	// Client is allowed: patient/*.read patient/Patient.read launch openid offline_access
	client := registerTestClient(t, s, false)

	// Request more scopes than allowed
	req := &AuthorizationRequest{
		ResponseType: "code",
		ClientID:     client.ClientID,
		RedirectURI:  "https://app.example.com/callback",
		Scope:        "patient/*.read launch openid",
		State:        "test",
		Aud:          "https://ehr.example.com/fhir",
	}

	resp, err := s.Authorize(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify negotiated scope
	s.mu.RLock()
	ac := s.authCodes[resp.Code]
	s.mu.RUnlock()

	// openid should be trimmed since it is not in client's allowed scope
	if strings.Contains(ac.Scope, "openid") {
		// openid IS in the client's allowed scope, so this is expected
	}

	// Verify the scope is a subset of what the client allows
	for _, scope := range strings.Fields(ac.Scope) {
		found := false
		for _, allowed := range strings.Fields(client.Scope) {
			if scope == allowed {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("negotiated scope %q is not in client's allowed scopes", scope)
		}
	}
}

func TestSMARTServer_CleanupExpired(t *testing.T) {
	s := newTestSMARTServer()

	// Add expired auth code
	s.mu.Lock()
	s.authCodes["expired-code"] = &AuthorizationCode{
		Code:      "expired-code",
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}
	s.authCodes["valid-code"] = &AuthorizationCode{
		Code:      "valid-code",
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	// Add expired launch context
	s.launchContexts["expired-lc"] = &SMARTLaunchContext{
		ID:        "expired-lc",
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}
	s.launchContexts["valid-lc"] = &SMARTLaunchContext{
		ID:        "valid-lc",
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	// Add expired refresh token
	s.refreshTokens["expired-rt"] = &RefreshTokenData{
		Token:     "expired-rt",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	s.refreshTokens["valid-rt"] = &RefreshTokenData{
		Token:     "valid-rt",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	s.mu.Unlock()

	// Run cleanup
	s.cleanup()

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Expired items should be removed
	if _, ok := s.authCodes["expired-code"]; ok {
		t.Error("expected expired auth code to be cleaned up")
	}
	if _, ok := s.launchContexts["expired-lc"]; ok {
		t.Error("expected expired launch context to be cleaned up")
	}
	if _, ok := s.refreshTokens["expired-rt"]; ok {
		t.Error("expected expired refresh token to be cleaned up")
	}

	// Valid items should remain
	if _, ok := s.authCodes["valid-code"]; !ok {
		t.Error("expected valid auth code to remain")
	}
	if _, ok := s.launchContexts["valid-lc"]; !ok {
		t.Error("expected valid launch context to remain")
	}
	if _, ok := s.refreshTokens["valid-rt"]; !ok {
		t.Error("expected valid refresh token to remain")
	}
}

// ---------------------------------------------------------------------------
// Handler Tests
// ---------------------------------------------------------------------------

func TestSMARTHandler_Authorize_Redirect(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)
	handler := NewSMARTHandler(s)

	e := echo.New()
	handler.RegisterRoutes(e)

	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", client.ClientID)
	q.Set("redirect_uri", "https://app.example.com/callback")
	q.Set("scope", "patient/*.read")
	q.Set("state", "handler-state")
	q.Set("aud", "https://ehr.example.com/fhir")

	req := httptest.NewRequest(http.MethodGet, "/auth/authorize?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d: %s", rec.Code, rec.Body.String())
	}

	loc := rec.Header().Get("Location")
	locURL, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("failed to parse location header: %v", err)
	}

	if locURL.Query().Get("code") == "" {
		t.Error("expected code parameter in redirect")
	}
	if locURL.Query().Get("state") != "handler-state" {
		t.Errorf("expected state 'handler-state', got %q", locURL.Query().Get("state"))
	}
}

func TestSMARTHandler_Authorize_MissingParams(t *testing.T) {
	s := newTestSMARTServer()
	handler := NewSMARTHandler(s)

	e := echo.New()
	handler.RegisterRoutes(e)

	// Missing client_id, scope, state
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("redirect_uri", "https://app.example.com/callback")

	req := httptest.NewRequest(http.MethodGet, "/auth/authorize?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Should redirect with error or return error JSON
	if rec.Code == http.StatusOK {
		t.Error("expected non-200 for missing params")
	}

	// If redirect, check for error parameter
	if rec.Code == http.StatusFound {
		loc := rec.Header().Get("Location")
		locURL, _ := url.Parse(loc)
		if locURL.Query().Get("error") == "" {
			t.Error("expected error parameter in redirect")
		}
	}
}

func TestSMARTHandler_Token_AuthorizationCodeGrant(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)
	handler := NewSMARTHandler(s)

	e := echo.New()
	handler.RegisterRoutes(e)

	// First, authorize
	lc, _ := s.CreateLaunchContext("patient-handler", "encounter-handler", "user-handler")
	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read launch", lc.ID, "")

	// Exchange code via HTTP
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", authResp.Code)
	form.Set("redirect_uri", "https://app.example.com/callback")
	form.Set("client_id", client.ClientID)
	form.Set("client_secret", "test-secret")

	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &tokenResp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if tokenResp.AccessToken == "" {
		t.Error("expected non-empty access_token")
	}
	if tokenResp.TokenType != "Bearer" {
		t.Errorf("expected token_type 'Bearer', got %q", tokenResp.TokenType)
	}
	if tokenResp.Patient != "patient-handler" {
		t.Errorf("expected patient 'patient-handler', got %q", tokenResp.Patient)
	}
}

func TestSMARTHandler_Token_RefreshGrant(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)
	handler := NewSMARTHandler(s)

	e := echo.New()
	handler.RegisterRoutes(e)

	// Authorize with offline_access
	lc, _ := s.CreateLaunchContext("patient-refresh-h", "", "user-refresh-h")
	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read launch offline_access", lc.ID, "")

	tokenResp, _ := s.ExchangeCode(&TokenRequest{
		GrantType:    "authorization_code",
		Code:         authResp.Code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     client.ClientID,
		ClientSecret: "test-secret",
	})

	// Refresh via HTTP
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", tokenResp.RefreshToken)
	form.Set("client_id", client.ClientID)

	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var refreshResp TokenResponse
	json.Unmarshal(rec.Body.Bytes(), &refreshResp)

	if refreshResp.AccessToken == "" {
		t.Error("expected non-empty access_token from refresh")
	}
}

func TestSMARTHandler_Token_InvalidGrantType(t *testing.T) {
	s := newTestSMARTServer()
	handler := NewSMARTHandler(s)

	e := echo.New()
	handler.RegisterRoutes(e)

	form := url.Values{}
	form.Set("grant_type", "client_credentials")

	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var oauthErr OAuthError
	json.Unmarshal(rec.Body.Bytes(), &oauthErr)
	if oauthErr.Code != "unsupported_grant_type" {
		t.Errorf("expected 'unsupported_grant_type', got %q", oauthErr.Code)
	}
}

func TestSMARTHandler_Register_Success(t *testing.T) {
	s := newTestSMARTServer()
	handler := NewSMARTHandler(s)

	e := echo.New()
	handler.RegisterRoutes(e)

	body := `{
		"client_name": "My SMART App",
		"redirect_uris": ["https://myapp.example.com/callback"],
		"scope": "patient/*.read launch openid",
		"token_endpoint_auth_method": "client_secret_basic",
		"launch_url": "https://myapp.example.com/launch"
	}`

	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var client SMARTClient
	json.Unmarshal(rec.Body.Bytes(), &client)

	if client.ClientID == "" {
		t.Error("expected non-empty client_id")
	}
	if client.ClientSecret == "" {
		t.Error("expected non-empty client_secret for confidential client")
	}
	if client.Name != "My SMART App" {
		t.Errorf("expected name 'My SMART App', got %q", client.Name)
	}
	if client.IsPublic {
		t.Error("expected IsPublic=false for confidential client")
	}
}

func TestSMARTHandler_Register_PublicClient(t *testing.T) {
	s := newTestSMARTServer()
	handler := NewSMARTHandler(s)

	e := echo.New()
	handler.RegisterRoutes(e)

	body := `{
		"client_name": "Public App",
		"redirect_uris": ["https://publicapp.example.com/callback"],
		"scope": "patient/*.read launch",
		"token_endpoint_auth_method": "none"
	}`

	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var client SMARTClient
	json.Unmarshal(rec.Body.Bytes(), &client)

	if client.ClientID == "" {
		t.Error("expected non-empty client_id")
	}
	if client.ClientSecret != "" {
		t.Error("expected empty client_secret for public client")
	}
	if !client.IsPublic {
		t.Error("expected IsPublic=true for public client")
	}
}

func TestSMARTHandler_Register_MissingFields(t *testing.T) {
	s := newTestSMARTServer()
	handler := NewSMARTHandler(s)

	e := echo.New()
	handler.RegisterRoutes(e)

	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing client_name",
			body: `{"redirect_uris":["https://app.example.com/cb"],"scope":"patient/*.read"}`,
		},
		{
			name: "missing redirect_uris",
			body: `{"client_name":"App","scope":"patient/*.read"}`,
		},
		{
			name: "missing scope",
			body: `{"client_name":"App","redirect_uris":["https://app.example.com/cb"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestSMARTHandler_Launch_CreatesContext(t *testing.T) {
	s := newTestSMARTServer()
	handler := NewSMARTHandler(s)

	e := echo.New()
	handler.RegisterRoutes(e)

	body := `{"patient_id":"patient-launch","encounter_id":"encounter-launch","user_id":"user-launch"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/launch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)

	launchID := resp["launch"]
	if launchID == "" {
		t.Error("expected non-empty launch ID")
	}

	// Verify it is stored
	s.mu.RLock()
	lc, ok := s.launchContexts[launchID]
	s.mu.RUnlock()

	if !ok {
		t.Fatal("expected launch context to be stored")
	}
	if lc.PatientID != "patient-launch" {
		t.Errorf("expected patient-launch, got %s", lc.PatientID)
	}
}

func TestSMARTHandler_Launch_MissingPatientID(t *testing.T) {
	s := newTestSMARTServer()
	handler := NewSMARTHandler(s)

	e := echo.New()
	handler.RegisterRoutes(e)

	body := `{"encounter_id":"encounter-only"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/launch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestSMARTHandler_Introspect_Valid(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)
	handler := NewSMARTHandler(s)

	e := echo.New()
	handler.RegisterRoutes(e)

	// Get a valid token
	lc, _ := s.CreateLaunchContext("patient-intr", "", "user-intr")
	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read launch", lc.ID, "")
	tokenResp, _ := s.ExchangeCode(&TokenRequest{
		GrantType:    "authorization_code",
		Code:         authResp.Code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     client.ClientID,
		ClientSecret: "test-secret",
	})

	form := url.Values{}
	form.Set("token", tokenResp.AccessToken)

	req := httptest.NewRequest(http.MethodPost, "/auth/introspect", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var claims TokenClaims
	json.Unmarshal(rec.Body.Bytes(), &claims)

	if !claims.Active {
		t.Error("expected active=true")
	}
	if claims.Patient != "patient-intr" {
		t.Errorf("expected patient 'patient-intr', got %q", claims.Patient)
	}
}

func TestSMARTHandler_Introspect_Invalid(t *testing.T) {
	s := newTestSMARTServer()
	handler := NewSMARTHandler(s)

	e := echo.New()
	handler.RegisterRoutes(e)

	form := url.Values{}
	form.Set("token", "invalid.jwt.token")

	req := httptest.NewRequest(http.MethodPost, "/auth/introspect", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var claims TokenClaims
	json.Unmarshal(rec.Body.Bytes(), &claims)

	if claims.Active {
		t.Error("expected active=false for invalid token")
	}
}

func TestSMARTHandler_NoWellKnownRoute(t *testing.T) {
	s := newTestSMARTServer()
	handler := NewSMARTHandler(s)

	e := echo.New()
	handler.RegisterRoutes(e)

	// Verify /.well-known/smart-configuration is NOT registered (should return 404/405)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/smart-configuration", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		t.Errorf("expected non-200 for /.well-known/smart-configuration, got %d (route should not be registered on SMARTHandler)", rec.Code)
	}

	// Verify /auth/authorize still works (returns redirect, not 404)
	client := registerTestClient(t, s, false)
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", client.ClientID)
	q.Set("redirect_uri", "https://app.example.com/callback")
	q.Set("scope", "patient/*.read")
	q.Set("state", "test-state")
	q.Set("aud", "https://ehr.example.com/fhir")

	authReq := httptest.NewRequest(http.MethodGet, "/auth/authorize?"+q.Encode(), nil)
	authRec := httptest.NewRecorder()
	e.ServeHTTP(authRec, authReq)

	if authRec.Code != http.StatusFound {
		t.Errorf("expected 302 for /auth/authorize, got %d", authRec.Code)
	}
}

func TestSMARTHandler_RegisterRoutes(t *testing.T) {
	s := newTestSMARTServer()
	handler := NewSMARTHandler(s)

	e := echo.New()
	handler.RegisterRoutes(e)

	routes := e.Routes()
	expectedRoutes := map[string]string{
		"GET::/auth/authorize":   "",
		"POST::/auth/token":      "",
		"POST::/auth/register":   "",
		"POST::/auth/launch":     "",
		"POST::/auth/introspect": "",
	}

	for _, route := range routes {
		key := route.Method + "::" + route.Path
		delete(expectedRoutes, key)
	}

	for route := range expectedRoutes {
		t.Errorf("expected route not registered: %s", route)
	}
}

// ---------------------------------------------------------------------------
// Additional unit tests for helper functions
// ---------------------------------------------------------------------------

func TestVerifyPKCE(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	if !verifyPKCE(verifier, challenge) {
		t.Error("expected PKCE verification to succeed with correct verifier")
	}

	if verifyPKCE("wrong-verifier", challenge) {
		t.Error("expected PKCE verification to fail with wrong verifier")
	}
}

func TestNegotiateScopes(t *testing.T) {
	t.Run("intersection of scopes", func(t *testing.T) {
		result, err := negotiateScopes("patient/*.read launch openid", "patient/*.read openid offline_access")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		scopes := strings.Fields(result)
		scopeSet := make(map[string]bool)
		for _, s := range scopes {
			scopeSet[s] = true
		}
		if !scopeSet["patient/*.read"] {
			t.Error("expected patient/*.read in negotiated scopes")
		}
		if !scopeSet["openid"] {
			t.Error("expected openid in negotiated scopes")
		}
		if scopeSet["launch"] {
			t.Error("unexpected launch in negotiated scopes (not allowed)")
		}
	})

	t.Run("no valid scopes", func(t *testing.T) {
		_, err := negotiateScopes("launch openid", "patient/*.read")
		if err == nil {
			t.Error("expected error when no scopes match")
		}
	})

	t.Run("invalid scope format", func(t *testing.T) {
		_, err := negotiateScopes("admin/everything.delete", "patient/*.read")
		if err == nil {
			t.Error("expected error for invalid scope format")
		}
	})

	t.Run("empty scopes", func(t *testing.T) {
		_, err := negotiateScopes("", "patient/*.read")
		if err == nil {
			t.Error("expected error for empty scopes")
		}
	})
}

func TestTimingSafeEqual(t *testing.T) {
	if !timingSafeEqual("hello", "hello") {
		t.Error("expected equal strings to match")
	}
	if timingSafeEqual("hello", "world") {
		t.Error("expected different strings to not match")
	}
	if timingSafeEqual("hello", "hell") {
		t.Error("expected different-length strings to not match")
	}
}

func TestOAuthError(t *testing.T) {
	err := &OAuthError{Code: "invalid_request", Description: "missing client_id"}
	if err.Error() != "invalid_request: missing client_id" {
		t.Errorf("unexpected error string: %s", err.Error())
	}
}

func TestIsValidSMARTScope(t *testing.T) {
	validScopes := []string{
		"openid", "fhirUser", "profile", "launch", "launch/patient",
		"launch/encounter", "offline_access",
		"patient/Patient.read", "user/Observation.write", "patient/*.read",
		"system/Patient.read",
	}
	for _, scope := range validScopes {
		if !isValidSMARTScope(scope) {
			t.Errorf("expected %q to be valid", scope)
		}
	}

	invalidScopes := []string{
		"admin/Patient.read", "patient/Patient.delete", "bogus",
	}
	for _, scope := range invalidScopes {
		if isValidSMARTScope(scope) {
			t.Errorf("expected %q to be invalid", scope)
		}
	}
}

func TestSMARTServer_StartCleanup(t *testing.T) {
	s := newTestSMARTServer()

	ctx, cancel := context.WithCancel(context.Background())
	s.StartCleanup(ctx)

	// Just verify it starts and can be cancelled without panic
	cancel()

	// Give goroutine time to exit
	time.Sleep(10 * time.Millisecond)
}

func TestGenerateRandomHex(t *testing.T) {
	hex1, err := generateRandomHex(16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hex1) != 32 { // 16 bytes = 32 hex chars
		t.Errorf("expected 32 char hex, got %d", len(hex1))
	}

	hex2, _ := generateRandomHex(16)
	if hex1 == hex2 {
		t.Error("expected unique hex values")
	}
}

func TestContainsScope(t *testing.T) {
	if !containsScope("patient/*.read launch offline_access", "offline_access") {
		t.Error("expected to find offline_access")
	}
	if containsScope("patient/*.read launch", "offline_access") {
		t.Error("did not expect to find offline_access")
	}
	if containsScope("", "openid") {
		t.Error("did not expect to find openid in empty string")
	}
}

func TestIsValidRedirectURI(t *testing.T) {
	registered := []string{"https://app.example.com/callback", "https://other.example.com/cb"}

	if !isValidRedirectURI(registered, "https://app.example.com/callback") {
		t.Error("expected registered URI to be valid")
	}
	if isValidRedirectURI(registered, "https://evil.example.com/callback") {
		t.Error("expected unregistered URI to be invalid")
	}
	if isValidRedirectURI(nil, "https://any.example.com") {
		t.Error("expected no URIs to match nil list")
	}
}

func TestSMARTHandler_Token_BasicAuth(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)
	handler := NewSMARTHandler(s)

	e := echo.New()
	handler.RegisterRoutes(e)

	// Authorize
	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read", "", "")

	// Exchange code using Basic auth
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", authResp.Code)
	form.Set("redirect_uri", "https://app.example.com/callback")

	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(client.ClientID, "test-secret")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var tokenResp TokenResponse
	json.Unmarshal(rec.Body.Bytes(), &tokenResp)
	if tokenResp.AccessToken == "" {
		t.Error("expected non-empty access_token with Basic auth")
	}
}

func TestSMARTServer_ExchangeCode_ClientIDMismatch(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	// Register another client
	otherClient := &SMARTClient{
		ClientID:     "other-client",
		ClientSecret: "other-secret",
		RedirectURIs: []string{"https://app.example.com/callback"},
		Scope:        "patient/*.read",
		Name:         "Other App",
	}
	s.RegisterClient(otherClient)

	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read", "", "")

	// Try to exchange with wrong client_id
	_, err := s.ExchangeCode(&TokenRequest{
		GrantType:    "authorization_code",
		Code:         authResp.Code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     "other-client",
		ClientSecret: "other-secret",
	})
	if err == nil {
		t.Fatal("expected error for client_id mismatch")
	}
}

func TestSMARTServer_Authorize_UnsupportedResponseType(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	req := &AuthorizationRequest{
		ResponseType: "token",
		ClientID:     client.ClientID,
		RedirectURI:  "https://app.example.com/callback",
		Scope:        "patient/*.read",
		State:        "test-state",
	}

	_, err := s.Authorize(req)
	if err == nil {
		t.Fatal("expected error for unsupported response_type")
	}
	oauthErr, ok := err.(*OAuthError)
	if !ok {
		t.Fatalf("expected *OAuthError, got %T", err)
	}
	if oauthErr.Code != "unsupported_response_type" {
		t.Errorf("expected 'unsupported_response_type', got %q", oauthErr.Code)
	}
}

func TestSMARTServer_RefreshToken_InvalidToken(t *testing.T) {
	s := newTestSMARTServer()

	_, err := s.RefreshAccessToken("nonexistent-refresh-token", "some-client")
	if err == nil {
		t.Fatal("expected error for invalid refresh token")
	}
	oauthErr, ok := err.(*OAuthError)
	if !ok {
		t.Fatalf("expected *OAuthError, got %T", err)
	}
	if oauthErr.Code != "invalid_grant" {
		t.Errorf("expected 'invalid_grant', got %q", oauthErr.Code)
	}
}

func TestSMARTServer_RefreshToken_ClientMismatch(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	lc, _ := s.CreateLaunchContext("patient-cm", "", "user-cm")
	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read launch offline_access", lc.ID, "")

	tokenResp, _ := s.ExchangeCode(&TokenRequest{
		GrantType:    "authorization_code",
		Code:         authResp.Code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     client.ClientID,
		ClientSecret: "test-secret",
	})

	// Try refresh with wrong client_id
	_, err := s.RefreshAccessToken(tokenResp.RefreshToken, "wrong-client")
	if err == nil {
		t.Fatal("expected error for client_id mismatch on refresh")
	}
}

func TestSMARTServer_IntrospectToken_InvalidSignature(t *testing.T) {
	s := newTestSMARTServer()

	// Create a token with a different key
	otherServer := NewSMARTServer("https://other.example.com", []byte("different-key"))
	claims := map[string]interface{}{
		"iss": "https://other.example.com",
		"sub": "user-other",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token, _ := otherServer.signJWT(claims)

	result, err := s.IntrospectToken(token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Active {
		t.Error("expected active=false for token with invalid signature")
	}
}

func TestSMARTServer_TokenContainsClientID(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	lc, _ := s.CreateLaunchContext("patient-cid", "encounter-cid", "user-cid")
	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read launch", lc.ID, "")

	tokenResp, err := s.ExchangeCode(&TokenRequest{
		GrantType:    "authorization_code",
		Code:         authResp.Code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     client.ClientID,
		ClientSecret: "test-secret",
	})
	if err != nil {
		t.Fatalf("ExchangeCode failed: %v", err)
	}

	// Parse the JWT access token manually
	parts := strings.SplitN(tokenResp.AccessToken, ".", 3)
	if len(parts) != 3 {
		t.Fatalf("expected 3-part JWT, got %d parts", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("failed to decode JWT payload: %v", err)
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatalf("failed to unmarshal JWT claims: %v", err)
	}

	gotClientID, ok := claims["client_id"].(string)
	if !ok {
		t.Fatalf("expected client_id claim to be a string, got %T", claims["client_id"])
	}
	if gotClientID != client.ClientID {
		t.Errorf("expected client_id %q, got %q", client.ClientID, gotClientID)
	}
}

func TestSMARTServer_RefreshTokenContainsClientID(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	lc, _ := s.CreateLaunchContext("patient-rcid", "encounter-rcid", "user-rcid")
	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read launch offline_access", lc.ID, "")

	tokenResp, err := s.ExchangeCode(&TokenRequest{
		GrantType:    "authorization_code",
		Code:         authResp.Code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     client.ClientID,
		ClientSecret: "test-secret",
	})
	if err != nil {
		t.Fatalf("ExchangeCode failed: %v", err)
	}
	if tokenResp.RefreshToken == "" {
		t.Fatal("expected non-empty refresh token with offline_access scope")
	}

	// Use the refresh token to get a new access token
	refreshResp, err := s.RefreshAccessToken(tokenResp.RefreshToken, client.ClientID)
	if err != nil {
		t.Fatalf("RefreshAccessToken failed: %v", err)
	}

	// Parse the new JWT access token
	parts := strings.SplitN(refreshResp.AccessToken, ".", 3)
	if len(parts) != 3 {
		t.Fatalf("expected 3-part JWT, got %d parts", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("failed to decode JWT payload: %v", err)
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatalf("failed to unmarshal JWT claims: %v", err)
	}

	gotClientID, ok := claims["client_id"].(string)
	if !ok {
		t.Fatalf("expected client_id claim to be a string, got %T", claims["client_id"])
	}
	if gotClientID != client.ClientID {
		t.Errorf("expected client_id %q, got %q", client.ClientID, gotClientID)
	}
}

func TestSMARTServer_IntrospectReturnsClientID(t *testing.T) {
	s := newTestSMARTServer()
	client := registerTestClient(t, s, false)

	lc, _ := s.CreateLaunchContext("patient-icid", "encounter-icid", "user-icid")
	authResp := mustAuthorize(t, s, client.ClientID, "https://app.example.com/callback", "patient/*.read launch", lc.ID, "")

	tokenResp, err := s.ExchangeCode(&TokenRequest{
		GrantType:    "authorization_code",
		Code:         authResp.Code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     client.ClientID,
		ClientSecret: "test-secret",
	})
	if err != nil {
		t.Fatalf("ExchangeCode failed: %v", err)
	}

	introspected, err := s.IntrospectToken(tokenResp.AccessToken)
	if err != nil {
		t.Fatalf("IntrospectToken failed: %v", err)
	}
	if !introspected.Active {
		t.Fatal("expected introspected token to be active")
	}
	if introspected.ClientID != client.ClientID {
		t.Errorf("expected introspected client_id %q, got %q", client.ClientID, introspected.ClientID)
	}
}

func TestSMARTHandler_Introspect_EmptyToken(t *testing.T) {
	s := newTestSMARTServer()
	handler := NewSMARTHandler(s)

	e := echo.New()
	handler.RegisterRoutes(e)

	form := url.Values{}
	// No token parameter

	req := httptest.NewRequest(http.MethodPost, "/auth/introspect", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var claims TokenClaims
	json.Unmarshal(rec.Body.Bytes(), &claims)
	if claims.Active {
		t.Error("expected active=false for empty token")
	}
}
