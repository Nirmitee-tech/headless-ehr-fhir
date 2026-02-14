package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type contextKey string

const (
	UserIDKey     contextKey = "user_id"
	UserRolesKey  contextKey = "user_roles"
	UserScopesKey contextKey = "user_scopes"
)

type Claims struct {
	jwt.RegisteredClaims
	TenantID   string   `json:"tenant_id"`
	Roles      []string `json:"roles"`
	FHIRScopes []string `json:"fhir_scopes"`
}

type JWTConfig struct {
	Issuer   string
	Audience string
	JWKSURL  string
	// SigningKey is used for development/testing only
	SigningKey []byte
}

// JWKSKey represents a single JSON Web Key from a JWKS endpoint.
type JWKSKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKSResponse represents the response from a JWKS endpoint.
type JWKSResponse struct {
	Keys []JWKSKey `json:"keys"`
}

// JWKSCache caches JWKS keys fetched from a remote endpoint with a configurable TTL.
type JWKSCache struct {
	mu       sync.RWMutex
	keys     map[string]*rsa.PublicKey
	jwksURL  string
	ttl      time.Duration
	fetchedAt time.Time
	client   *http.Client
}

// NewJWKSCache creates a new JWKS cache that fetches keys from the given URL.
func NewJWKSCache(jwksURL string, ttl time.Duration) *JWKSCache {
	return &JWKSCache{
		keys:    make(map[string]*rsa.PublicKey),
		jwksURL: jwksURL,
		ttl:     ttl,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// GetKey returns the RSA public key for the given kid.
// It fetches keys from the JWKS endpoint if the cache is expired or if the kid is not found.
func (c *JWKSCache) GetKey(kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	key, ok := c.keys[kid]
	expired := time.Since(c.fetchedAt) > c.ttl
	c.mu.RUnlock()

	if ok && !expired {
		return key, nil
	}

	// Cache miss or expired: fetch fresh keys
	if err := c.fetch(); err != nil {
		return nil, fmt.Errorf("fetching JWKS: %w", err)
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	key, ok = c.keys[kid]
	if !ok {
		return nil, fmt.Errorf("key with kid %q not found in JWKS", kid)
	}
	return key, nil
}

// fetch retrieves the JWKS from the remote endpoint and updates the cache.
func (c *JWKSCache) fetch() error {
	resp, err := c.client.Get(c.jwksURL)
	if err != nil {
		return fmt.Errorf("GET %s: %w", c.jwksURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks JWKSResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("decoding JWKS response: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" {
			continue
		}
		pubKey, err := parseRSAPublicKey(k)
		if err != nil {
			continue // skip malformed keys
		}
		keys[k.Kid] = pubKey
	}

	c.mu.Lock()
	c.keys = keys
	c.fetchedAt = time.Now()
	c.mu.Unlock()

	return nil
}

// parseRSAPublicKey converts a JWKSKey to an *rsa.PublicKey.
func parseRSAPublicKey(k JWKSKey) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("decoding modulus: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("decoding exponent: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

// defaultJWKSCacheTTL is the default time-to-live for cached JWKS keys.
const defaultJWKSCacheTTL = 5 * time.Minute

// jwksKeyFunc returns a jwt.Keyfunc that fetches public keys from a JWKS endpoint.
// Keys are cached in memory and automatically refreshed on cache miss or TTL expiry.
func jwksKeyFunc(jwksURL string) jwt.Keyfunc {
	cache := NewJWKSCache(jwksURL, defaultJWKSCacheTTL)
	return func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, fmt.Errorf("token has no kid header")
		}
		return cache.GetKey(kid)
	}
}

func JWTMiddleware(cfg JWTConfig) echo.MiddlewareFunc {
	// Resolve JWKS URL: if not explicitly set, try OIDC auto-discovery from issuer.
	resolvedJWKSURL := cfg.JWKSURL
	if resolvedJWKSURL == "" && cfg.Issuer != "" && len(cfg.SigningKey) == 0 {
		provider, err := NewOIDCProvider(cfg.Issuer)
		if err == nil {
			resolvedJWKSURL = provider.JWKSURI
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization format")
			}

			tokenStr := parts[1]
			claims := &Claims{}

			opts := []jwt.ParserOption{
				jwt.WithValidMethods([]string{"RS256", "HS256"}),
			}
			if cfg.Issuer != "" {
				opts = append(opts, jwt.WithIssuer(cfg.Issuer))
			}
			if cfg.Audience != "" {
				opts = append(opts, jwt.WithAudience(cfg.Audience))
			}

			var token *jwt.Token
			var err error

			if len(cfg.SigningKey) > 0 {
				// Dev mode: HMAC signing key
				token, err = jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
					return cfg.SigningKey, nil
				}, opts...)
			} else {
				// Production: JWKS validation
				token, err = jwt.ParseWithClaims(tokenStr, claims, jwksKeyFunc(resolvedJWKSURL), opts...)
			}

			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			// Set values on echo context for tenant middleware
			c.Set("jwt_tenant_id", claims.TenantID)

			// Set values on request context
			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.Subject)
			ctx = context.WithValue(ctx, UserRolesKey, claims.Roles)
			ctx = context.WithValue(ctx, UserScopesKey, claims.FHIRScopes)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// DevAuthMiddleware is a permissive middleware for development that allows
// unauthenticated requests with default values.
func DevAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				// In dev mode, set defaults
				c.Set("jwt_tenant_id", "default")
				ctx := c.Request().Context()
				ctx = context.WithValue(ctx, UserIDKey, "dev-user")
				ctx = context.WithValue(ctx, UserRolesKey, []string{"admin"})
				ctx = context.WithValue(ctx, UserScopesKey, []string{"user/*.*"})
				c.SetRequest(c.Request().WithContext(ctx))
				return next(c)
			}
			// If token is provided, still validate it
			return next(c)
		}
	}
}

func UserIDFromContext(ctx context.Context) string {
	uid, _ := ctx.Value(UserIDKey).(string)
	return uid
}

func RolesFromContext(ctx context.Context) []string {
	roles, _ := ctx.Value(UserRolesKey).([]string)
	return roles
}

func ScopesFromContext(ctx context.Context) []string {
	scopes, _ := ctx.Value(UserScopesKey).([]string)
	return scopes
}
