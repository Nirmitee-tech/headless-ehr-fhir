package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Data Structures
// ---------------------------------------------------------------------------

// SMARTClient represents a registered SMART on FHIR application.
type SMARTClient struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret,omitempty"`
	RedirectURIs []string `json:"redirect_uris"`
	Scope        string   `json:"scope"`
	Name         string   `json:"client_name"`
	LaunchURL    string   `json:"launch_url,omitempty"`
	IsPublic     bool     `json:"is_public"`
}

// AuthorizationCode is a short-lived code exchanged for tokens.
type AuthorizationCode struct {
	Code                string
	ClientID            string
	RedirectURI         string
	Scope               string
	ExpiresAt           time.Time
	PatientID           string
	EncounterID         string
	UserID              string
	CodeChallenge       string
	CodeChallengeMethod string
}

// SMARTLaunchContext holds EHR launch context data for the SMART authorization
// server. This is distinct from the existing LaunchContext type used by the
// launch context store.
type SMARTLaunchContext struct {
	ID          string
	PatientID   string
	EncounterID string
	UserID      string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

// TokenResponse is the OAuth2 token response with SMART extensions.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Patient      string `json:"patient,omitempty"`
	Encounter    string `json:"encounter,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// RefreshTokenData holds the data associated with a refresh token.
type RefreshTokenData struct {
	Token       string
	ClientID    string
	Scope       string
	PatientID   string
	EncounterID string
	UserID      string
	ExpiresAt   time.Time
}

// AuthorizationRequest represents the OAuth2 authorization request parameters.
type AuthorizationRequest struct {
	ResponseType        string
	ClientID            string
	RedirectURI         string
	Scope               string
	State               string
	Aud                 string
	Launch              string
	CodeChallenge       string
	CodeChallengeMethod string
}

// AuthorizationResponse is the result of a successful authorization.
type AuthorizationResponse struct {
	Code        string
	RedirectURI string
	State       string
}

// TokenRequest represents the OAuth2 token exchange request parameters.
type TokenRequest struct {
	GrantType    string
	Code         string
	RedirectURI  string
	ClientID     string
	ClientSecret string
	CodeVerifier string
	RefreshToken string
}

// TokenClaims represents the claims extracted from an introspected token.
type TokenClaims struct {
	Active    bool   `json:"active"`
	Subject   string `json:"sub,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
	Scope     string `json:"scope,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty"`
	Issuer    string `json:"iss,omitempty"`
	TokenID   string `json:"jti,omitempty"`
	Patient   string `json:"patient,omitempty"`
	Encounter string `json:"encounter,omitempty"`
	FHIRUser  string `json:"fhirUser,omitempty"`
}

// ---------------------------------------------------------------------------
// SMARTServer
// ---------------------------------------------------------------------------

// SMARTServer implements the SMART on FHIR authorization server.
type SMARTServer struct {
	mu             sync.RWMutex
	clients        map[string]*SMARTClient
	authCodes      map[string]*AuthorizationCode
	launchContexts map[string]*SMARTLaunchContext
	refreshTokens  map[string]*RefreshTokenData
	signingKey     []byte
	issuer         string
	codeExpiry     time.Duration
	tokenExpiry    time.Duration
	refreshExpiry  time.Duration
}

// NewSMARTServer creates a new SMART authorization server.
func NewSMARTServer(issuer string, signingKey []byte) *SMARTServer {
	return &SMARTServer{
		clients:        make(map[string]*SMARTClient),
		authCodes:      make(map[string]*AuthorizationCode),
		launchContexts: make(map[string]*SMARTLaunchContext),
		refreshTokens:  make(map[string]*RefreshTokenData),
		signingKey:     signingKey,
		issuer:         issuer,
		codeExpiry:     5 * time.Minute,
		tokenExpiry:    1 * time.Hour,
		refreshExpiry:  24 * time.Hour,
	}
}

// RegisterClient registers a SMART application.
func (s *SMARTServer) RegisterClient(client *SMARTClient) error {
	if client.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[client.ClientID]; exists {
		return fmt.Errorf("client_id %q already registered", client.ClientID)
	}

	s.clients[client.ClientID] = client
	return nil
}

// CreateLaunchContext creates a new EHR launch context.
func (s *SMARTServer) CreateLaunchContext(patientID, encounterID, userID string) (*SMARTLaunchContext, error) {
	id, err := generateRandomHex(32)
	if err != nil {
		return nil, fmt.Errorf("generating launch context ID: %w", err)
	}

	now := time.Now()
	lc := &SMARTLaunchContext{
		ID:          id,
		PatientID:   patientID,
		EncounterID: encounterID,
		UserID:      userID,
		CreatedAt:   now,
		ExpiresAt:   now.Add(s.codeExpiry),
	}

	s.mu.Lock()
	s.launchContexts[id] = lc
	s.mu.Unlock()

	return lc, nil
}

// Authorize handles the authorization request.
func (s *SMARTServer) Authorize(req *AuthorizationRequest) (*AuthorizationResponse, error) {
	if req.ResponseType != "code" {
		return nil, &OAuthError{Code: "unsupported_response_type", Description: "response_type must be 'code'"}
	}

	s.mu.RLock()
	client, ok := s.clients[req.ClientID]
	s.mu.RUnlock()

	if !ok {
		return nil, &OAuthError{Code: "invalid_request", Description: "unknown client_id"}
	}

	if !isValidRedirectURI(client.RedirectURIs, req.RedirectURI) {
		return nil, &OAuthError{Code: "invalid_request", Description: "redirect_uri not registered for this client"}
	}

	// Validate and negotiate scopes
	negotiatedScope, err := negotiateScopes(req.Scope, client.Scope)
	if err != nil {
		return nil, &OAuthError{Code: "invalid_scope", Description: err.Error()}
	}

	// Build the authorization code
	code, err := generateRandomHex(32)
	if err != nil {
		return nil, fmt.Errorf("generating authorization code: %w", err)
	}

	ac := &AuthorizationCode{
		Code:                code,
		ClientID:            req.ClientID,
		RedirectURI:         req.RedirectURI,
		Scope:               negotiatedScope,
		ExpiresAt:           time.Now().Add(s.codeExpiry),
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
	}

	// Handle EHR launch context
	if req.Launch != "" {
		s.mu.Lock()
		lc, lcOK := s.launchContexts[req.Launch]
		if lcOK {
			delete(s.launchContexts, req.Launch)
		}
		s.mu.Unlock()

		if !lcOK || time.Now().After(lc.ExpiresAt) {
			return nil, &OAuthError{Code: "invalid_request", Description: "invalid or expired launch context"}
		}

		ac.PatientID = lc.PatientID
		ac.EncounterID = lc.EncounterID
		ac.UserID = lc.UserID
	}

	s.mu.Lock()
	s.authCodes[code] = ac
	s.mu.Unlock()

	return &AuthorizationResponse{
		Code:        code,
		RedirectURI: req.RedirectURI,
		State:       req.State,
	}, nil
}

// ExchangeCode exchanges an authorization code for tokens.
func (s *SMARTServer) ExchangeCode(req *TokenRequest) (*TokenResponse, error) {
	if req.GrantType != "authorization_code" {
		return nil, &OAuthError{Code: "unsupported_grant_type", Description: "grant_type must be 'authorization_code'"}
	}

	s.mu.Lock()
	ac, ok := s.authCodes[req.Code]
	if ok {
		delete(s.authCodes, req.Code) // one-time use
	}
	s.mu.Unlock()

	if !ok {
		return nil, &OAuthError{Code: "invalid_grant", Description: "invalid or already used authorization code"}
	}

	if time.Now().After(ac.ExpiresAt) {
		return nil, &OAuthError{Code: "invalid_grant", Description: "authorization code has expired"}
	}

	if ac.RedirectURI != req.RedirectURI {
		return nil, &OAuthError{Code: "invalid_grant", Description: "redirect_uri does not match"}
	}

	if ac.ClientID != req.ClientID {
		return nil, &OAuthError{Code: "invalid_grant", Description: "client_id does not match"}
	}

	// Look up the client for secret/public validation
	s.mu.RLock()
	client, clientOK := s.clients[req.ClientID]
	s.mu.RUnlock()

	if !clientOK {
		return nil, &OAuthError{Code: "invalid_client", Description: "unknown client"}
	}

	// Validate client authentication
	if client.IsPublic {
		// Public clients require PKCE
		if ac.CodeChallenge == "" {
			return nil, &OAuthError{Code: "invalid_request", Description: "PKCE is required for public clients"}
		}
	} else {
		// Confidential clients must provide a valid client_secret
		if !timingSafeEqual(req.ClientSecret, client.ClientSecret) {
			return nil, &OAuthError{Code: "invalid_client", Description: "invalid client_secret"}
		}
	}

	// Verify PKCE if code_challenge was provided
	if ac.CodeChallenge != "" {
		if req.CodeVerifier == "" {
			return nil, &OAuthError{Code: "invalid_grant", Description: "code_verifier is required"}
		}
		if !verifyPKCE(req.CodeVerifier, ac.CodeChallenge) {
			return nil, &OAuthError{Code: "invalid_grant", Description: "PKCE verification failed"}
		}
	}

	// Generate tokens
	now := time.Now()
	tokenID, _ := generateRandomHex(16)

	claims := map[string]interface{}{
		"iss":   s.issuer,
		"sub":   ac.UserID,
		"aud":   s.issuer + "/fhir",
		"exp":   now.Add(s.tokenExpiry).Unix(),
		"iat":   now.Unix(),
		"jti":   tokenID,
		"scope": ac.Scope,
	}

	if ac.PatientID != "" {
		claims["patient"] = ac.PatientID
	}
	if ac.EncounterID != "" {
		claims["encounter"] = ac.EncounterID
	}
	if ac.UserID != "" {
		claims["fhirUser"] = "Practitioner/" + ac.UserID
	}

	accessToken, err := s.signJWT(claims)
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	resp := &TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.tokenExpiry.Seconds()),
		Scope:       ac.Scope,
		Patient:     ac.PatientID,
		Encounter:   ac.EncounterID,
	}

	// Generate refresh token if offline_access scope is requested
	if containsScope(ac.Scope, "offline_access") {
		refreshToken, rtErr := generateRandomHex(32)
		if rtErr != nil {
			return nil, fmt.Errorf("generating refresh token: %w", rtErr)
		}

		rtData := &RefreshTokenData{
			Token:       refreshToken,
			ClientID:    ac.ClientID,
			Scope:       ac.Scope,
			PatientID:   ac.PatientID,
			EncounterID: ac.EncounterID,
			UserID:      ac.UserID,
			ExpiresAt:   now.Add(s.refreshExpiry),
		}

		s.mu.Lock()
		s.refreshTokens[refreshToken] = rtData
		s.mu.Unlock()

		resp.RefreshToken = refreshToken
	}

	return resp, nil
}

// RefreshAccessToken exchanges a refresh token for a new access token.
func (s *SMARTServer) RefreshAccessToken(refreshToken, clientID string) (*TokenResponse, error) {
	s.mu.RLock()
	rtData, ok := s.refreshTokens[refreshToken]
	s.mu.RUnlock()

	if !ok {
		return nil, &OAuthError{Code: "invalid_grant", Description: "invalid refresh token"}
	}

	if time.Now().After(rtData.ExpiresAt) {
		s.mu.Lock()
		delete(s.refreshTokens, refreshToken)
		s.mu.Unlock()
		return nil, &OAuthError{Code: "invalid_grant", Description: "refresh token has expired"}
	}

	if rtData.ClientID != clientID {
		return nil, &OAuthError{Code: "invalid_grant", Description: "client_id does not match refresh token"}
	}

	now := time.Now()
	tokenID, _ := generateRandomHex(16)

	claims := map[string]interface{}{
		"iss":   s.issuer,
		"sub":   rtData.UserID,
		"aud":   s.issuer + "/fhir",
		"exp":   now.Add(s.tokenExpiry).Unix(),
		"iat":   now.Unix(),
		"jti":   tokenID,
		"scope": rtData.Scope,
	}

	if rtData.PatientID != "" {
		claims["patient"] = rtData.PatientID
	}
	if rtData.EncounterID != "" {
		claims["encounter"] = rtData.EncounterID
	}
	if rtData.UserID != "" {
		claims["fhirUser"] = "Practitioner/" + rtData.UserID
	}

	accessToken, err := s.signJWT(claims)
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.tokenExpiry.Seconds()),
		Scope:        rtData.Scope,
		RefreshToken: refreshToken,
		Patient:      rtData.PatientID,
		Encounter:    rtData.EncounterID,
	}, nil
}

// IntrospectToken validates and returns claims for an access token.
func (s *SMARTServer) IntrospectToken(token string) (*TokenClaims, error) {
	claims, err := s.parseJWT(token)
	if err != nil {
		return &TokenClaims{Active: false}, nil
	}

	exp, _ := claims["exp"].(float64)
	if time.Now().Unix() > int64(exp) {
		return &TokenClaims{Active: false}, nil
	}

	sub, _ := claims["sub"].(string)
	clientID, _ := claims["client_id"].(string)
	scope, _ := claims["scope"].(string)
	iat, _ := claims["iat"].(float64)
	iss, _ := claims["iss"].(string)
	jti, _ := claims["jti"].(string)
	patient, _ := claims["patient"].(string)
	encounter, _ := claims["encounter"].(string)
	fhirUser, _ := claims["fhirUser"].(string)

	return &TokenClaims{
		Active:    true,
		Subject:   sub,
		ClientID:  clientID,
		Scope:     scope,
		ExpiresAt: int64(exp),
		IssuedAt:  int64(iat),
		Issuer:    iss,
		TokenID:   jti,
		Patient:   patient,
		Encounter: encounter,
		FHIRUser:  fhirUser,
	}, nil
}

// StartCleanup starts a background goroutine to clean expired codes/contexts.
func (s *SMARTServer) StartCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.cleanup()
			}
		}
	}()
}

// cleanup removes expired auth codes, launch contexts, and refresh tokens.
func (s *SMARTServer) cleanup() {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for code, ac := range s.authCodes {
		if now.After(ac.ExpiresAt) {
			delete(s.authCodes, code)
		}
	}

	for id, lc := range s.launchContexts {
		if now.After(lc.ExpiresAt) {
			delete(s.launchContexts, id)
		}
	}

	for token, rt := range s.refreshTokens {
		if now.After(rt.ExpiresAt) {
			delete(s.refreshTokens, token)
		}
	}
}

// ---------------------------------------------------------------------------
// JWT Helpers
// ---------------------------------------------------------------------------

// signJWT creates a JWT signed with HMAC-SHA256.
func (s *SMARTServer) signJWT(claims map[string]interface{}) (string, error) {
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

	mac := hmac.New(sha256.New, s.signingKey)
	mac.Write([]byte(signingInput))
	signature := mac.Sum(nil)

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + signatureB64, nil
}

// parseJWT parses and verifies a JWT signed with HMAC-SHA256.
func (s *SMARTServer) parseJWT(tokenStr string) (map[string]interface{}, error) {
	parts := strings.SplitN(tokenStr, ".", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Verify signature
	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, s.signingKey)
	mac.Write([]byte(signingInput))
	expectedSig := mac.Sum(nil)

	actualSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decoding signature: %w", err)
	}

	if !hmac.Equal(expectedSig, actualSig) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Decode payload
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decoding payload: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("unmarshaling claims: %w", err)
	}

	return claims, nil
}

// ---------------------------------------------------------------------------
// PKCE Helpers
// ---------------------------------------------------------------------------

// verifyPKCE verifies a PKCE code_verifier against a code_challenge using S256.
func verifyPKCE(verifier, challenge string) bool {
	hash := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(hash[:])
	return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}

// ---------------------------------------------------------------------------
// Scope Helpers
// ---------------------------------------------------------------------------

// validSMARTScopes defines the set of recognized non-resource SMART scopes.
var validSMARTScopes = map[string]bool{
	"openid":           true,
	"fhirUser":         true,
	"profile":          true,
	"launch":           true,
	"launch/patient":   true,
	"launch/encounter": true,
	"offline_access":   true,
}

// isValidSMARTScope checks if a scope string is a valid SMART scope.
func isValidSMARTScope(scope string) bool {
	if validSMARTScopes[scope] {
		return true
	}
	// Check if it is a valid resource-level scope
	_, err := ParseSMARTScope(scope)
	return err == nil
}

// negotiateScopes returns the intersection of requested and allowed scopes.
// If a requested scope is not valid, an error is returned.
func negotiateScopes(requested, allowed string) (string, error) {
	requestedScopes := strings.Fields(requested)
	if len(requestedScopes) == 0 {
		return "", fmt.Errorf("no scopes requested")
	}

	// Validate all requested scopes
	for _, s := range requestedScopes {
		if !isValidSMARTScope(s) {
			return "", fmt.Errorf("invalid scope: %s", s)
		}
	}

	allowedScopes := make(map[string]bool)
	for _, s := range strings.Fields(allowed) {
		allowedScopes[s] = true
	}

	// Negotiate: keep only scopes that are in the client's allowed set
	var negotiated []string
	for _, s := range requestedScopes {
		if allowedScopes[s] {
			negotiated = append(negotiated, s)
		}
	}

	if len(negotiated) == 0 {
		return "", fmt.Errorf("no requested scopes are allowed for this client")
	}

	return strings.Join(negotiated, " "), nil
}

// containsScope checks if a space-separated scope string contains a specific scope.
func containsScope(scopeStr, target string) bool {
	for _, s := range strings.Fields(scopeStr) {
		if s == target {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Utility Helpers
// ---------------------------------------------------------------------------

// generateRandomHex generates a cryptographically random hex string of n bytes.
func generateRandomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// isValidRedirectURI checks if a redirect URI is registered for a client.
func isValidRedirectURI(registered []string, uri string) bool {
	for _, r := range registered {
		if r == uri {
			return true
		}
	}
	return false
}

// timingSafeEqual compares two strings in constant time.
func timingSafeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// OAuthError represents an OAuth 2.0 error response.
type OAuthError struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
}

func (e *OAuthError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Description)
}

// ---------------------------------------------------------------------------
// SMARTHandler — HTTP endpoints
// ---------------------------------------------------------------------------

// SMARTHandler provides SMART on FHIR HTTP endpoints.
type SMARTHandler struct {
	server *SMARTServer
}

// NewSMARTHandler creates a new SMART HTTP handler.
func NewSMARTHandler(server *SMARTServer) *SMARTHandler {
	return &SMARTHandler{server: server}
}

// RegisterRoutes registers SMART authorization endpoints on the echo instance.
func (h *SMARTHandler) RegisterRoutes(e *echo.Echo) {
	e.GET("/auth/authorize", h.handleAuthorize)
	e.POST("/auth/token", h.handleToken)
	e.POST("/auth/register", h.handleRegister)
	e.POST("/auth/launch", h.handleLaunch)
	e.POST("/auth/introspect", h.handleIntrospect)
	e.GET("/.well-known/smart-configuration", h.handleSMARTConfiguration)
}

// handleAuthorize handles GET /auth/authorize.
func (h *SMARTHandler) handleAuthorize(c echo.Context) error {
	req := &AuthorizationRequest{
		ResponseType:        c.QueryParam("response_type"),
		ClientID:            c.QueryParam("client_id"),
		RedirectURI:         c.QueryParam("redirect_uri"),
		Scope:               c.QueryParam("scope"),
		State:               c.QueryParam("state"),
		Aud:                 c.QueryParam("aud"),
		Launch:              c.QueryParam("launch"),
		CodeChallenge:       c.QueryParam("code_challenge"),
		CodeChallengeMethod: c.QueryParam("code_challenge_method"),
	}

	// Validate required parameters
	if req.ResponseType == "" || req.ClientID == "" || req.RedirectURI == "" || req.Scope == "" || req.State == "" {
		return h.redirectWithError(c, req.RedirectURI, "invalid_request", "missing required parameters", req.State)
	}

	resp, err := h.server.Authorize(req)
	if err != nil {
		oauthErr, ok := err.(*OAuthError)
		if ok {
			return h.redirectWithError(c, req.RedirectURI, oauthErr.Code, oauthErr.Description, req.State)
		}
		return h.redirectWithError(c, req.RedirectURI, "server_error", "internal server error", req.State)
	}

	// Redirect with authorization code
	redirectURL, parseErr := url.Parse(resp.RedirectURI)
	if parseErr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid redirect URI")
	}

	q := redirectURL.Query()
	q.Set("code", resp.Code)
	q.Set("state", resp.State)
	redirectURL.RawQuery = q.Encode()

	return c.Redirect(http.StatusFound, redirectURL.String())
}

// redirectWithError sends an OAuth2 error redirect.
func (h *SMARTHandler) redirectWithError(c echo.Context, redirectURI, errCode, errDesc, state string) error {
	if redirectURI == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             errCode,
			"error_description": errDesc,
		})
	}

	redirectURL, parseErr := url.Parse(redirectURI)
	if parseErr != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             errCode,
			"error_description": errDesc,
		})
	}

	q := redirectURL.Query()
	q.Set("error", errCode)
	q.Set("error_description", errDesc)
	if state != "" {
		q.Set("state", state)
	}
	redirectURL.RawQuery = q.Encode()

	return c.Redirect(http.StatusFound, redirectURL.String())
}

// handleToken handles POST /auth/token.
func (h *SMARTHandler) handleToken(c echo.Context) error {
	grantType := c.FormValue("grant_type")

	switch grantType {
	case "authorization_code":
		return h.handleTokenAuthorizationCode(c)
	case "refresh_token":
		return h.handleTokenRefresh(c)
	default:
		return c.JSON(http.StatusBadRequest, &OAuthError{
			Code:        "unsupported_grant_type",
			Description: "grant_type must be 'authorization_code' or 'refresh_token'",
		})
	}
}

// handleTokenAuthorizationCode handles the authorization_code grant type.
func (h *SMARTHandler) handleTokenAuthorizationCode(c echo.Context) error {
	clientID, clientSecret := h.extractClientCredentials(c)

	req := &TokenRequest{
		GrantType:    "authorization_code",
		Code:         c.FormValue("code"),
		RedirectURI:  c.FormValue("redirect_uri"),
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CodeVerifier: c.FormValue("code_verifier"),
	}

	resp, err := h.server.ExchangeCode(req)
	if err != nil {
		oauthErr, ok := err.(*OAuthError)
		if ok {
			status := http.StatusBadRequest
			if oauthErr.Code == "invalid_client" {
				status = http.StatusUnauthorized
			}
			return c.JSON(status, oauthErr)
		}
		return c.JSON(http.StatusInternalServerError, &OAuthError{
			Code:        "server_error",
			Description: "internal server error",
		})
	}

	return c.JSON(http.StatusOK, resp)
}

// handleTokenRefresh handles the refresh_token grant type.
func (h *SMARTHandler) handleTokenRefresh(c echo.Context) error {
	clientID, _ := h.extractClientCredentials(c)

	refreshToken := c.FormValue("refresh_token")
	if refreshToken == "" {
		return c.JSON(http.StatusBadRequest, &OAuthError{
			Code:        "invalid_request",
			Description: "refresh_token is required",
		})
	}

	resp, err := h.server.RefreshAccessToken(refreshToken, clientID)
	if err != nil {
		oauthErr, ok := err.(*OAuthError)
		if ok {
			return c.JSON(http.StatusBadRequest, oauthErr)
		}
		return c.JSON(http.StatusInternalServerError, &OAuthError{
			Code:        "server_error",
			Description: "internal server error",
		})
	}

	return c.JSON(http.StatusOK, resp)
}

// extractClientCredentials extracts client_id and client_secret from the
// request, supporting both form body and HTTP Basic authentication.
func (h *SMARTHandler) extractClientCredentials(c echo.Context) (string, string) {
	// Try HTTP Basic auth first
	clientID, clientSecret, ok := c.Request().BasicAuth()
	if ok && clientID != "" {
		return clientID, clientSecret
	}

	// Fall back to form values
	return c.FormValue("client_id"), c.FormValue("client_secret")
}

// handleRegister handles POST /auth/register (dynamic client registration).
func (h *SMARTHandler) handleRegister(c echo.Context) error {
	var regReq struct {
		ClientName              string   `json:"client_name"`
		RedirectURIs            []string `json:"redirect_uris"`
		Scope                   string   `json:"scope"`
		TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
		LaunchURL               string   `json:"launch_url"`
	}

	if err := c.Bind(&regReq); err != nil {
		return c.JSON(http.StatusBadRequest, &OAuthError{
			Code:        "invalid_request",
			Description: "invalid request body",
		})
	}

	if regReq.ClientName == "" || len(regReq.RedirectURIs) == 0 || regReq.Scope == "" {
		return c.JSON(http.StatusBadRequest, &OAuthError{
			Code:        "invalid_request",
			Description: "client_name, redirect_uris, and scope are required",
		})
	}

	clientID, err := generateRandomHex(16)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &OAuthError{
			Code:        "server_error",
			Description: "failed to generate client_id",
		})
	}

	isPublic := regReq.TokenEndpointAuthMethod == "none"

	client := &SMARTClient{
		ClientID:     clientID,
		RedirectURIs: regReq.RedirectURIs,
		Scope:        regReq.Scope,
		Name:         regReq.ClientName,
		LaunchURL:    regReq.LaunchURL,
		IsPublic:     isPublic,
	}

	if !isPublic {
		secret, genErr := generateRandomHex(32)
		if genErr != nil {
			return c.JSON(http.StatusInternalServerError, &OAuthError{
				Code:        "server_error",
				Description: "failed to generate client_secret",
			})
		}
		client.ClientSecret = secret
	}

	if err := h.server.RegisterClient(client); err != nil {
		return c.JSON(http.StatusInternalServerError, &OAuthError{
			Code:        "server_error",
			Description: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, client)
}

// handleLaunch handles POST /auth/launch (creates EHR launch context).
func (h *SMARTHandler) handleLaunch(c echo.Context) error {
	var req struct {
		PatientID   string `json:"patient_id"`
		EncounterID string `json:"encounter_id"`
		UserID      string `json:"user_id"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, &OAuthError{
			Code:        "invalid_request",
			Description: "invalid request body",
		})
	}

	if req.PatientID == "" {
		return c.JSON(http.StatusBadRequest, &OAuthError{
			Code:        "invalid_request",
			Description: "patient_id is required",
		})
	}

	lc, err := h.server.CreateLaunchContext(req.PatientID, req.EncounterID, req.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &OAuthError{
			Code:        "server_error",
			Description: "failed to create launch context",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"launch": lc.ID,
	})
}

// handleIntrospect handles POST /auth/introspect.
func (h *SMARTHandler) handleIntrospect(c echo.Context) error {
	token := c.FormValue("token")
	if token == "" {
		return c.JSON(http.StatusOK, &TokenClaims{Active: false})
	}

	claims, err := h.server.IntrospectToken(token)
	if err != nil {
		return c.JSON(http.StatusOK, &TokenClaims{Active: false})
	}

	return c.JSON(http.StatusOK, claims)
}

// handleSMARTConfiguration handles GET /.well-known/smart-configuration.
func (h *SMARTHandler) handleSMARTConfiguration(c echo.Context) error {
	cfg := map[string]interface{}{
		"issuer":                 h.server.issuer,
		"authorization_endpoint": h.server.issuer + "/auth/authorize",
		"token_endpoint":         h.server.issuer + "/auth/token",
		"registration_endpoint":  h.server.issuer + "/auth/register",
		"introspection_endpoint": h.server.issuer + "/auth/introspect",
		"scopes_supported": []string{
			"patient/*.read", "patient/*.write",
			"user/*.read", "user/*.write",
			"launch", "launch/patient", "launch/encounter",
			"openid", "fhirUser",
			"offline_access",
		},
		"response_types_supported": []string{"code"},
		"capabilities": []string{
			"launch-ehr",
			"launch-standalone",
			"client-public",
			"client-confidential-symmetric",
			"sso-openid-connect",
			"permission-patient",
			"permission-user",
			"context-ehr-patient",
			"context-ehr-encounter",
			"context-standalone-patient",
		},
		"code_challenge_methods_supported":        []string{"S256"},
		"grant_types_supported":                   []string{"authorization_code", "refresh_token"},
		"token_endpoint_auth_methods_supported":   []string{"client_secret_basic", "client_secret_post", "none"},
	}

	return c.JSON(http.StatusOK, cfg)
}
