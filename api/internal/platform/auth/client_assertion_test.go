package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"
)

func makeTestJWKSAndSigner(t *testing.T) (jwksJSON string, sign func(claims map[string]interface{}) string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pub := key.PublicKey
	eBytes := big.NewInt(int64(pub.E)).Bytes()
	set := map[string]interface{}{"keys": []map[string]interface{}{{
		"kty": "RSA", "alg": "RS384", "kid": "test-key-1", "use": "sig",
		"n": base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		"e": base64.RawURLEncoding.EncodeToString(eBytes),
	}}}
	setJSON, _ := json.Marshal(set)

	sign = func(claims map[string]interface{}) string {
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS384","kid":"test-key-1","typ":"JWT"}`))
		payload, _ := json.Marshal(claims)
		body := base64.RawURLEncoding.EncodeToString(payload)
		digest := sha512.Sum384([]byte(header + "." + body))
		sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA384, digest[:])
		if err != nil {
			t.Fatal(err)
		}
		return header + "." + body + "." + base64.RawURLEncoding.EncodeToString(sig)
	}
	return string(setJSON), sign
}

func TestVerifyClientAssertion(t *testing.T) {
	jwksJSON, sign := makeTestJWKSAndSigner(t)
	aud := []string{"http://localhost:8000/auth/token"}
	valid := func(jti string) map[string]interface{} {
		return map[string]interface{}{
			"iss": "client-a", "sub": "client-a",
			"aud": "http://localhost:8000/auth/token",
			"exp": time.Now().Add(4 * time.Minute).Unix(),
			"jti": jti,
		}
	}

	t.Run("valid assertion passes", func(t *testing.T) {
		if err := verifyClientAssertion(sign(valid("jti-1")), "client-a", jwksJSON, aud); err != nil {
			t.Fatalf("expected pass, got: %v", err)
		}
	})

	t.Run("replayed jti rejected", func(t *testing.T) {
		a := sign(valid("jti-replay"))
		if err := verifyClientAssertion(a, "client-a", jwksJSON, aud); err != nil {
			t.Fatalf("first use should pass: %v", err)
		}
		if err := verifyClientAssertion(a, "client-a", jwksJSON, aud); err == nil {
			t.Fatal("replayed assertion must be rejected")
		}
	})

	t.Run("wrong client_id rejected", func(t *testing.T) {
		if err := verifyClientAssertion(sign(valid("jti-2")), "client-b", jwksJSON, aud); err == nil {
			t.Fatal("iss/sub mismatch must be rejected")
		}
	})

	t.Run("wrong audience rejected", func(t *testing.T) {
		c := valid("jti-3")
		c["aud"] = "https://attacker.example.com/token"
		if err := verifyClientAssertion(sign(c), "client-a", jwksJSON, aud); err == nil {
			t.Fatal("wrong aud must be rejected")
		}
	})

	t.Run("expired assertion rejected", func(t *testing.T) {
		c := valid("jti-4")
		c["exp"] = time.Now().Add(-1 * time.Minute).Unix()
		if err := verifyClientAssertion(sign(c), "client-a", jwksJSON, aud); err == nil {
			t.Fatal("expired assertion must be rejected")
		}
	})

	t.Run("tampered payload rejected", func(t *testing.T) {
		a := sign(valid("jti-5"))
		parts := []byte(a)
		// flip a char inside the payload segment
		mid := len(a) / 2
		if parts[mid] == 'A' {
			parts[mid] = 'B'
		} else {
			parts[mid] = 'A'
		}
		if err := verifyClientAssertion(string(parts), "client-a", jwksJSON, aud); err == nil {
			t.Fatal("tampered assertion must be rejected")
		}
	})

	t.Run("private key material never used for verification", func(t *testing.T) {
		var set jwkSet
		_ = json.Unmarshal([]byte(jwksJSON), &set)
		set.Keys[0].D = "c29tZS1wcml2YXRlLWtleQ" // mark as private
		withPriv, _ := json.Marshal(set)
		if err := verifyClientAssertion(sign(valid("jti-6")), "client-a", string(withPriv), aud); err == nil {
			t.Fatal("JWKS containing only private-marked keys must not verify")
		}
	})
}

func TestIsValidV2Operation(t *testing.T) {
	for op, want := range map[string]bool{
		"r": true, "rs": true, "cruds": true, "us": true, "c": true,
		"": false, "sr": false, "x": false, "rr": false, "read": false,
	} {
		if got := isValidV2Operation(op); got != want {
			t.Errorf("isValidV2Operation(%q) = %v, want %v", op, got, want)
		}
	}
}

func ExampleClientAssertionType() {
	fmt.Println(ClientAssertionType)
	// Output: urn:ietf:params:oauth:client-assertion-type:jwt-bearer
}
