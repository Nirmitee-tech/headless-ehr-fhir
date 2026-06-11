package auth

// SMART App Launch STU2 asymmetric ("private_key_jwt") client authentication.
// Confidential clients authenticate to the token endpoint with a signed JWT
// (client_assertion) instead of a client_secret, per RFC 7523 and the SMART
// App Launch "client-confidential-asymmetric" capability.
//
// Supported signing algorithms: RS384 and ES384 (the two required by SMART).

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
)

// ClientAssertionType is the only assertion type accepted at the token endpoint.
const ClientAssertionType = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"

// maxAssertionLifetime bounds how far in the future an assertion's exp may be.
const maxAssertionLifetime = 10 * time.Minute

// jwkKey is a minimal JWK representation covering RSA and EC public keys.
type jwkKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	// RSA
	N string `json:"n"`
	E string `json:"e"`
	// EC
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
	// Private-key component; keys carrying it are never used for verification.
	D string `json:"d"`
}

type jwkSet struct {
	Keys []jwkKey `json:"keys"`
}

// jtiReplayCache tracks seen assertion jti values until their expiry so a
// captured assertion cannot be replayed.
type jtiReplayCache struct {
	mu   sync.Mutex
	seen map[string]time.Time
}

var assertionJTICache = &jtiReplayCache{seen: make(map[string]time.Time)}

// checkAndStore returns false if jti was already seen (replay); otherwise it
// records the jti until exp and returns true. Expired entries are evicted.
func (c *jtiReplayCache) checkAndStore(jti string, exp time.Time) bool {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.seen {
		if now.After(v) {
			delete(c.seen, k)
		}
	}
	if _, dup := c.seen[jti]; dup {
		return false
	}
	c.seen[jti] = exp
	return true
}

// assertionClaims are the registered claims required in a client assertion.
type assertionClaims struct {
	Iss string          `json:"iss"`
	Sub string          `json:"sub"`
	Aud json.RawMessage `json:"aud"` // string or []string per RFC 7519
	Exp int64           `json:"exp"`
	Jti string          `json:"jti"`
}

func (a *assertionClaims) audiences() []string {
	var single string
	if err := json.Unmarshal(a.Aud, &single); err == nil {
		return []string{single}
	}
	var multi []string
	if err := json.Unmarshal(a.Aud, &multi); err == nil {
		return multi
	}
	return nil
}

// peekAssertionIssuer extracts the iss claim WITHOUT verifying the signature.
// Used only to look up the client whose registered JWKS will then verify it.
func peekAssertionIssuer(assertion string) (string, error) {
	parts := strings.Split(assertion, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("malformed assertion: expected 3 segments")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("malformed assertion payload: %w", err)
	}
	var c assertionClaims
	if err := json.Unmarshal(payload, &c); err != nil {
		return "", fmt.Errorf("malformed assertion claims: %w", err)
	}
	if c.Iss == "" {
		return "", fmt.Errorf("assertion missing iss claim")
	}
	return c.Iss, nil
}

// verifyClientAssertion fully validates a client assertion JWT against the
// client's registered JWKS: signature (RS384/ES384), iss == sub == clientID,
// aud ∈ acceptedAud, exp in the future (bounded), and jti not replayed.
func verifyClientAssertion(assertion, clientID, jwksJSON string, acceptedAud []string) error {
	parts := strings.Split(assertion, ".")
	if len(parts) != 3 {
		return fmt.Errorf("malformed assertion: expected 3 segments")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return fmt.Errorf("malformed assertion header: %w", err)
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return fmt.Errorf("malformed assertion header: %w", err)
	}
	if header.Alg != "RS384" && header.Alg != "ES384" {
		return fmt.Errorf("unsupported assertion alg %q: must be RS384 or ES384", header.Alg)
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("malformed assertion payload: %w", err)
	}
	var claims assertionClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return fmt.Errorf("malformed assertion claims: %w", err)
	}

	// Claim validation
	if claims.Iss != clientID || claims.Sub != clientID {
		return fmt.Errorf("assertion iss and sub must equal the client_id")
	}
	audOK := false
	for _, aud := range claims.audiences() {
		for _, accepted := range acceptedAud {
			if aud == accepted {
				audOK = true
			}
		}
	}
	if !audOK {
		return fmt.Errorf("assertion aud does not match the token endpoint")
	}
	if claims.Exp == 0 {
		return fmt.Errorf("assertion missing exp claim")
	}
	exp := time.Unix(claims.Exp, 0)
	now := time.Now()
	if now.After(exp) {
		return fmt.Errorf("assertion has expired")
	}
	if exp.After(now.Add(maxAssertionLifetime)) {
		return fmt.Errorf("assertion exp too far in the future (max %s)", maxAssertionLifetime)
	}
	if claims.Jti == "" {
		return fmt.Errorf("assertion missing jti claim")
	}

	// Signature verification against the registered JWKS
	var set jwkSet
	if err := json.Unmarshal([]byte(jwksJSON), &set); err != nil {
		return fmt.Errorf("client JWKS is invalid: %w", err)
	}
	key, err := selectJWK(set, header.Kid, header.Alg)
	if err != nil {
		return err
	}

	signingInput := parts[0] + "." + parts[1]
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("malformed assertion signature: %w", err)
	}
	digest := sha512.Sum384([]byte(signingInput))

	switch key.Kty {
	case "RSA":
		pub, err := jwkToRSAPublicKey(key)
		if err != nil {
			return err
		}
		if err := rsa.VerifyPKCS1v15(pub, crypto.SHA384, digest[:], sig); err != nil {
			return fmt.Errorf("assertion signature verification failed")
		}
	case "EC":
		pub, err := jwkToECDSAPublicKey(key)
		if err != nil {
			return err
		}
		// JOSE EC signatures are r||s, each padded to the curve byte size (48 for P-384).
		if len(sig) != 96 {
			return fmt.Errorf("invalid ES384 signature length %d", len(sig))
		}
		r := new(big.Int).SetBytes(sig[:48])
		s := new(big.Int).SetBytes(sig[48:])
		if !ecdsa.Verify(pub, digest[:], r, s) {
			return fmt.Errorf("assertion signature verification failed")
		}
	default:
		return fmt.Errorf("unsupported JWK key type %q", key.Kty)
	}

	// Replay protection — only after the signature is proven valid.
	if !assertionJTICache.checkAndStore(claims.Jti, exp) {
		return fmt.Errorf("assertion jti has already been used")
	}
	return nil
}

// selectJWK picks the verification key: by kid when present, otherwise the
// first public key matching the assertion's algorithm family.
func selectJWK(set jwkSet, kid, alg string) (*jwkKey, error) {
	wantKty := "RSA"
	if alg == "ES384" {
		wantKty = "EC"
	}
	for i := range set.Keys {
		k := &set.Keys[i]
		if k.D != "" {
			continue // never verify with an entry carrying private material
		}
		if kid != "" && k.Kid == kid && k.Kty == wantKty {
			return k, nil
		}
	}
	if kid == "" {
		for i := range set.Keys {
			k := &set.Keys[i]
			if k.D == "" && k.Kty == wantKty {
				return k, nil
			}
		}
	}
	return nil, fmt.Errorf("no matching public key (kid=%q, alg=%s) in client JWKS", kid, alg)
}

func jwkToRSAPublicKey(k *jwkKey) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("invalid RSA modulus: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("invalid RSA exponent: %w", err)
	}
	e := 0
	for _, b := range eBytes {
		e = e<<8 | int(b)
	}
	if e <= 1 {
		return nil, fmt.Errorf("invalid RSA exponent")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
}

func jwkToECDSAPublicKey(k *jwkKey) (*ecdsa.PublicKey, error) {
	if k.Crv != "P-384" {
		return nil, fmt.Errorf("unsupported EC curve %q: must be P-384", k.Crv)
	}
	xBytes, err := base64.RawURLEncoding.DecodeString(k.X)
	if err != nil {
		return nil, fmt.Errorf("invalid EC x coordinate: %w", err)
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(k.Y)
	if err != nil {
		return nil, fmt.Errorf("invalid EC y coordinate: %w", err)
	}
	return &ecdsa.PublicKey{
		Curve: elliptic.P384(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}

// VerifyClientAssertion authenticates a private_key_jwt client assertion and
// returns the authenticated client_id. The assertion's iss identifies the
// client; its registered JWKS verifies the signature.
func (s *SMARTServer) VerifyClientAssertion(assertionType, assertion string) (string, *OAuthError) {
	if assertionType != ClientAssertionType {
		return "", &OAuthError{Code: "invalid_request", Description: "unsupported client_assertion_type"}
	}
	clientID, err := peekAssertionIssuer(assertion)
	if err != nil {
		return "", &OAuthError{Code: "invalid_client", Description: err.Error()}
	}

	s.mu.RLock()
	client, ok := s.clients[clientID]
	s.mu.RUnlock()
	if !ok {
		return "", &OAuthError{Code: "invalid_client", Description: "unknown client"}
	}
	if client.JWKS == "" {
		return "", &OAuthError{Code: "invalid_client", Description: "client is not registered for asymmetric authentication"}
	}

	acceptedAud := []string{s.issuer + "/auth/token", s.issuer}
	if err := verifyClientAssertion(assertion, clientID, client.JWKS, acceptedAud); err != nil {
		return "", &OAuthError{Code: "invalid_client", Description: err.Error()}
	}
	return clientID, nil
}
