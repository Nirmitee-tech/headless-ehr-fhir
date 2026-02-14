package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestOIDCProvider_Discovery(t *testing.T) {
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(JWKSResponse{Keys: []JWKSKey{}})
	}))
	defer jwksServer.Close()

	discoveryDoc := map[string]interface{}{
		"issuer":                 "https://idp.example.com",
		"authorization_endpoint": "https://idp.example.com/authorize",
		"token_endpoint":         "https://idp.example.com/token",
		"userinfo_endpoint":      "https://idp.example.com/userinfo",
		"jwks_uri":               jwksServer.URL,
		"scopes_supported":       []string{"openid", "profile", "fhirUser"},
		"response_types_supported":                []string{"code"},
		"grant_types_supported":                   []string{"authorization_code", "client_credentials"},
		"subject_types_supported":                 []string{"public"},
		"id_token_signing_alg_values_supported":   []string{"RS256"},
		"token_endpoint_auth_methods_supported":   []string{"client_secret_basic"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(discoveryDoc)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	provider, err := NewOIDCProvider(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if provider.AuthorizationEndpoint != "https://idp.example.com/authorize" {
		t.Errorf("expected authorization_endpoint=https://idp.example.com/authorize, got %s", provider.AuthorizationEndpoint)
	}
	if provider.TokenEndpoint != "https://idp.example.com/token" {
		t.Errorf("expected token_endpoint=https://idp.example.com/token, got %s", provider.TokenEndpoint)
	}
	if provider.UserinfoEndpoint != "https://idp.example.com/userinfo" {
		t.Errorf("expected userinfo_endpoint=https://idp.example.com/userinfo, got %s", provider.UserinfoEndpoint)
	}
	if provider.JWKSURI != jwksServer.URL {
		t.Errorf("expected jwks_uri=%s, got %s", jwksServer.URL, provider.JWKSURI)
	}
	if len(provider.ScopesSupported) != 3 {
		t.Errorf("expected 3 scopes, got %d", len(provider.ScopesSupported))
	}

	// Check SupportsScope helper
	if !provider.SupportsScope("openid") {
		t.Error("expected SupportsScope(openid) to be true")
	}
	if provider.SupportsScope("nonexistent") {
		t.Error("expected SupportsScope(nonexistent) to be false")
	}
}

func TestOIDCProvider_InvalidIssuer(t *testing.T) {
	// Test with a server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	_, err := NewOIDCProvider(server.URL)
	if err == nil {
		t.Fatal("expected error for invalid issuer")
	}

	// Test with an unreachable server
	_, err = NewOIDCProvider("http://127.0.0.1:1")
	if err == nil {
		t.Fatal("expected error for unreachable issuer")
	}
}

func TestOIDCProvider_MissingJWKSURI(t *testing.T) {
	discoveryDoc := map[string]interface{}{
		"issuer":                 "https://idp.example.com",
		"authorization_endpoint": "https://idp.example.com/authorize",
		"token_endpoint":         "https://idp.example.com/token",
		// jwks_uri intentionally omitted
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(discoveryDoc)
	}))
	defer server.Close()

	_, err := NewOIDCProvider(server.URL)
	if err == nil {
		t.Fatal("expected error for missing jwks_uri")
	}
}

func TestOIDCProvider_JWKSKeyFunc(t *testing.T) {
	// Generate a test RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	kid := "test-key-1"
	jwksHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := JWKSResponse{
			Keys: []JWKSKey{
				rsaPublicKeyToJWK(privateKey, kid),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	jwksServer := httptest.NewServer(jwksHandler)
	defer jwksServer.Close()

	discoveryDoc := map[string]interface{}{
		"issuer":                 "https://idp.example.com",
		"authorization_endpoint": "https://idp.example.com/authorize",
		"token_endpoint":         "https://idp.example.com/token",
		"jwks_uri":               jwksServer.URL,
	}

	discoveryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(discoveryDoc)
			return
		}
		http.NotFound(w, r)
	}))
	defer discoveryServer.Close()

	provider, err := NewOIDCProvider(discoveryServer.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	keyFunc := provider.JWKSKeyFunc()
	if keyFunc == nil {
		t.Fatal("JWKSKeyFunc returned nil")
	}
}

func TestJWKSCache_Fetch(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	kid := "fetch-test-key"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := JWKSResponse{
			Keys: []JWKSKey{
				rsaPublicKeyToJWK(privateKey, kid),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cache := NewJWKSCache(server.URL, 5*time.Minute)

	key, err := cache.GetKey(kid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == nil {
		t.Fatal("expected non-nil key")
	}

	// Verify the key matches the original
	if key.N.Cmp(privateKey.PublicKey.N) != 0 {
		t.Error("fetched key modulus does not match original")
	}
	if key.E != privateKey.PublicKey.E {
		t.Error("fetched key exponent does not match original")
	}

	// Test cache hit: second call should not need the server
	key2, err := cache.GetKey(kid)
	if err != nil {
		t.Fatalf("unexpected error on cache hit: %v", err)
	}
	if key2.N.Cmp(key.N) != 0 {
		t.Error("cached key should match first fetched key")
	}
}

func TestJWKSCache_KeyRotation(t *testing.T) {
	privateKey1, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key 1: %v", err)
	}
	privateKey2, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key 2: %v", err)
	}

	kid1 := "rotation-key-1"
	kid2 := "rotation-key-2"
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var keys []JWKSKey
		if callCount == 1 {
			// First fetch: only key 1
			keys = []JWKSKey{rsaPublicKeyToJWK(privateKey1, kid1)}
		} else {
			// Subsequent fetches: both keys (simulating key rotation)
			keys = []JWKSKey{
				rsaPublicKeyToJWK(privateKey1, kid1),
				rsaPublicKeyToJWK(privateKey2, kid2),
			}
		}
		resp := JWKSResponse{Keys: keys}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Use a very short TTL so the cache expires between calls
	cache := NewJWKSCache(server.URL, 1*time.Millisecond)

	// Fetch key 1 - should work
	key1, err := cache.GetKey(kid1)
	if err != nil {
		t.Fatalf("unexpected error fetching key1: %v", err)
	}
	if key1 == nil {
		t.Fatal("expected non-nil key1")
	}

	// Wait for cache to expire
	time.Sleep(5 * time.Millisecond)

	// Fetch key 2 - should trigger re-fetch and find the new key
	key2, err := cache.GetKey(kid2)
	if err != nil {
		t.Fatalf("unexpected error fetching key2 after rotation: %v", err)
	}
	if key2 == nil {
		t.Fatal("expected non-nil key2")
	}

	if key2.N.Cmp(privateKey2.PublicKey.N) != 0 {
		t.Error("rotated key modulus does not match")
	}

	if callCount < 2 {
		t.Errorf("expected at least 2 JWKS fetches for key rotation, got %d", callCount)
	}
}

func TestJWKSCache_TTL(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	kid := "ttl-test-key"
	fetchCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchCount++
		resp := JWKSResponse{
			Keys: []JWKSKey{rsaPublicKeyToJWK(privateKey, kid)},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Use a long TTL
	cache := NewJWKSCache(server.URL, 10*time.Minute)

	// First fetch
	_, err = cache.GetKey(kid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetchCount != 1 {
		t.Errorf("expected 1 fetch, got %d", fetchCount)
	}

	// Second fetch should use cache (no additional HTTP call)
	_, err = cache.GetKey(kid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetchCount != 1 {
		t.Errorf("expected 1 fetch (cached), got %d", fetchCount)
	}

	// Now use a very short TTL to test expiry
	cacheShort := NewJWKSCache(server.URL, 1*time.Millisecond)

	_, err = cacheShort.GetKey(kid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	firstCount := fetchCount

	// Wait for cache to expire
	time.Sleep(5 * time.Millisecond)

	_, err = cacheShort.GetKey(kid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fetchCount <= firstCount {
		t.Error("expected additional fetch after TTL expiry")
	}
}

func TestJWKSCache_KeyNotFound(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := JWKSResponse{
			Keys: []JWKSKey{rsaPublicKeyToJWK(privateKey, "existing-key")},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cache := NewJWKSCache(server.URL, 5*time.Minute)

	_, err = cache.GetKey("nonexistent-key")
	if err == nil {
		t.Fatal("expected error for nonexistent key")
	}
}

func TestJWKSCache_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cache := NewJWKSCache(server.URL, 5*time.Minute)

	_, err := cache.GetKey("any-key")
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}

// rsaPublicKeyToJWK converts an RSA private key to a JWKSKey for testing.
func rsaPublicKeyToJWK(privateKey *rsa.PrivateKey, kid string) JWKSKey {
	pub := &privateKey.PublicKey
	return JWKSKey{
		Kty: "RSA",
		Kid: kid,
		Use: "sig",
		Alg: "RS256",
		N:   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}
}

func TestParseRSAPublicKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	jwk := rsaPublicKeyToJWK(privateKey, "parse-test")
	pubKey, err := parseRSAPublicKey(jwk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pubKey.N.Cmp(privateKey.PublicKey.N) != 0 {
		t.Error("parsed key modulus does not match original")
	}
	if pubKey.E != privateKey.PublicKey.E {
		t.Error("parsed key exponent does not match original")
	}
}

func TestParseRSAPublicKey_InvalidModulus(t *testing.T) {
	jwk := JWKSKey{
		Kty: "RSA",
		Kid: "bad-key",
		N:   "!!!invalid-base64!!!",
		E:   "AQAB",
	}
	_, err := parseRSAPublicKey(jwk)
	if err == nil {
		t.Fatal("expected error for invalid modulus")
	}
}

func TestParseRSAPublicKey_InvalidExponent(t *testing.T) {
	jwk := JWKSKey{
		Kty: "RSA",
		Kid: "bad-key",
		N:   base64.RawURLEncoding.EncodeToString(big.NewInt(12345).Bytes()),
		E:   "!!!invalid-base64!!!",
	}
	_, err := parseRSAPublicKey(jwk)
	if err == nil {
		t.Fatal("expected error for invalid exponent")
	}
}

func TestJwksKeyFunc_NoKidHeader(t *testing.T) {
	// Create a mock JWKS server (won't be called since we fail before fetch)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(JWKSResponse{Keys: []JWKSKey{}})
	}))
	defer server.Close()

	keyFunc := jwksKeyFunc(server.URL)

	// Token with no kid header
	token := &jwt.Token{
		Header: map[string]interface{}{},
	}

	_, err := keyFunc(token)
	if err == nil {
		t.Fatal("expected error for token without kid")
	}
	if fmt.Sprintf("%v", err) != "token has no kid header" {
		t.Errorf("unexpected error message: %v", err)
	}
}
