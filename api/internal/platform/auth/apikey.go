package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Sentinel errors
// ---------------------------------------------------------------------------

var (
	// ErrKeyNotFound indicates the requested API key does not exist in the store.
	ErrKeyNotFound = errors.New("api key not found")

	// ErrKeyRevoked indicates the API key has been revoked and can no longer be used.
	ErrKeyRevoked = errors.New("api key revoked")

	// ErrKeyExpired indicates the API key has passed its expiration time.
	ErrKeyExpired = errors.New("api key expired")

	// ErrInvalidKey indicates the provided raw key does not match any stored hash.
	ErrInvalidKey = errors.New("invalid api key")

	// ErrInsufficientScopes indicates the API key does not have the required
	// SMART scopes for the requested operation.
	ErrInsufficientScopes = errors.New("insufficient scopes")
)

// ---------------------------------------------------------------------------
// APIKey struct
// ---------------------------------------------------------------------------

// APIKey represents a managed API key for programmatic access to the EHR platform.
// The actual key material is never stored; only a SHA-256 hash is persisted.
type APIKey struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	KeyHash    string            `json:"-"` // never serialize
	KeyPrefix  string            `json:"key_prefix"`
	TenantID   string            `json:"tenant_id"`
	ClientID   string            `json:"client_id"`
	Scopes     []string          `json:"scopes"`
	RateLimit  int               `json:"rate_limit"`
	Status     string            `json:"status"`
	ExpiresAt  *time.Time        `json:"expires_at,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	RevokedAt  *time.Time        `json:"revoked_at,omitempty"`
	LastUsedAt *time.Time        `json:"last_used_at,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// ---------------------------------------------------------------------------
// APIKeyStore interface
// ---------------------------------------------------------------------------

// APIKeyStore defines the contract for persisting and querying API keys.
// Implementations may be backed by in-memory maps, a relational database, etc.
type APIKeyStore interface {
	// CreateKey persists a new API key.
	CreateKey(ctx context.Context, key *APIKey) error

	// GetByID retrieves an API key by its unique ID.
	GetByID(ctx context.Context, id string) (*APIKey, error)

	// GetByHash retrieves an API key by its SHA-256 hash.
	GetByHash(ctx context.Context, hash string) (*APIKey, error)

	// ListByTenant returns API keys belonging to a tenant with pagination.
	// Returns the matching keys and the total count (before pagination).
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*APIKey, int, error)

	// ListByClient returns all API keys belonging to a specific client.
	ListByClient(ctx context.Context, clientID string) ([]*APIKey, error)

	// UpdateKey persists changes to an existing API key.
	UpdateKey(ctx context.Context, key *APIKey) error

	// DeleteKey permanently removes an API key from the store.
	DeleteKey(ctx context.Context, id string) error
}

// ---------------------------------------------------------------------------
// InMemoryAPIKeyStore
// ---------------------------------------------------------------------------

// InMemoryAPIKeyStore provides a thread-safe in-memory implementation of
// APIKeyStore. It is suitable for development, testing, and single-node
// deployments.
type InMemoryAPIKeyStore struct {
	mu      sync.RWMutex
	byID    map[string]*APIKey
	byHash  map[string]string // hash -> ID
	ordered []string          // insertion-order IDs for stable pagination
}

// NewInMemoryAPIKeyStore creates a new empty in-memory store.
func NewInMemoryAPIKeyStore() *InMemoryAPIKeyStore {
	return &InMemoryAPIKeyStore{
		byID:   make(map[string]*APIKey),
		byHash: make(map[string]string),
	}
}

// CreateKey implements APIKeyStore.
func (s *InMemoryAPIKeyStore) CreateKey(_ context.Context, key *APIKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp := copyKey(key)
	s.byID[cp.ID] = cp
	if cp.KeyHash != "" {
		s.byHash[cp.KeyHash] = cp.ID
	}
	s.ordered = append(s.ordered, cp.ID)
	return nil
}

// GetByID implements APIKeyStore.
func (s *InMemoryAPIKeyStore) GetByID(_ context.Context, id string) (*APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	k, ok := s.byID[id]
	if !ok {
		return nil, ErrKeyNotFound
	}
	return copyKey(k), nil
}

// GetByHash implements APIKeyStore.
func (s *InMemoryAPIKeyStore) GetByHash(_ context.Context, hash string) (*APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.byHash[hash]
	if !ok {
		return nil, ErrKeyNotFound
	}
	k, ok := s.byID[id]
	if !ok {
		return nil, ErrKeyNotFound
	}
	return copyKey(k), nil
}

// ListByTenant implements APIKeyStore.
func (s *InMemoryAPIKeyStore) ListByTenant(_ context.Context, tenantID string, limit, offset int) ([]*APIKey, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Collect matching keys in insertion order.
	var matching []*APIKey
	for _, id := range s.ordered {
		k, ok := s.byID[id]
		if !ok {
			continue
		}
		if k.TenantID == tenantID {
			matching = append(matching, k)
		}
	}

	total := len(matching)

	// Apply pagination.
	if offset > len(matching) {
		offset = len(matching)
	}
	matching = matching[offset:]
	if limit > 0 && limit < len(matching) {
		matching = matching[:limit]
	}

	result := make([]*APIKey, len(matching))
	for i, k := range matching {
		result[i] = copyKey(k)
	}
	return result, total, nil
}

// ListByClient implements APIKeyStore.
func (s *InMemoryAPIKeyStore) ListByClient(_ context.Context, clientID string) ([]*APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*APIKey
	for _, id := range s.ordered {
		k, ok := s.byID[id]
		if !ok {
			continue
		}
		if k.ClientID == clientID {
			result = append(result, copyKey(k))
		}
	}
	return result, nil
}

// UpdateKey implements APIKeyStore.
func (s *InMemoryAPIKeyStore) UpdateKey(_ context.Context, key *APIKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.byID[key.ID]
	if !ok {
		return ErrKeyNotFound
	}

	// If the hash changed, update the index.
	if existing.KeyHash != key.KeyHash {
		delete(s.byHash, existing.KeyHash)
		if key.KeyHash != "" {
			s.byHash[key.KeyHash] = key.ID
		}
	}

	s.byID[key.ID] = copyKey(key)
	return nil
}

// DeleteKey implements APIKeyStore.
func (s *InMemoryAPIKeyStore) DeleteKey(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.byID[id]
	if !ok {
		return ErrKeyNotFound
	}

	delete(s.byHash, existing.KeyHash)
	delete(s.byID, id)

	// Remove from ordered slice.
	for i, oid := range s.ordered {
		if oid == id {
			s.ordered = append(s.ordered[:i], s.ordered[i+1:]...)
			break
		}
	}
	return nil
}

// copyKey returns a deep copy of an APIKey to prevent mutation through shared pointers.
func copyKey(k *APIKey) *APIKey {
	cp := *k
	if k.Scopes != nil {
		cp.Scopes = make([]string, len(k.Scopes))
		copy(cp.Scopes, k.Scopes)
	}
	if k.Metadata != nil {
		cp.Metadata = make(map[string]string, len(k.Metadata))
		for mk, mv := range k.Metadata {
			cp.Metadata[mk] = mv
		}
	}
	if k.ExpiresAt != nil {
		t := *k.ExpiresAt
		cp.ExpiresAt = &t
	}
	if k.RevokedAt != nil {
		t := *k.RevokedAt
		cp.RevokedAt = &t
	}
	if k.LastUsedAt != nil {
		t := *k.LastUsedAt
		cp.LastUsedAt = &t
	}
	return &cp
}

// ---------------------------------------------------------------------------
// APIKeyManager
// ---------------------------------------------------------------------------

const (
	// apiKeyPrefix is prepended to every generated key for easy identification
	// in logs and configuration files.
	apiKeyPrefix = "ehr_k1_"

	// apiKeyRandomBytes is the number of random bytes used to generate the
	// key material (encoded as hex => 32 hex chars).
	apiKeyRandomBytes = 16
)

// APIKeyManager orchestrates API key lifecycle operations: generation,
// validation, revocation, and rotation.
type APIKeyManager struct {
	store APIKeyStore
}

// NewAPIKeyManager creates a new manager backed by the given store.
func NewAPIKeyManager(store APIKeyStore) *APIKeyManager {
	return &APIKeyManager{store: store}
}

// GenerateKey creates a new API key with the given parameters and persists it
// in the store. It returns the APIKey struct and the raw key string. The raw
// key is only available at creation time and must be shown to the caller
// exactly once.
func (m *APIKeyManager) GenerateKey(
	ctx context.Context,
	name, tenantID, clientID string,
	scopes []string,
	rateLimit int,
	expiresAt *time.Time,
) (*APIKey, string, error) {
	rawKey, err := generateRawKey()
	if err != nil {
		return nil, "", fmt.Errorf("generating raw key: %w", err)
	}

	hash := hashKey(rawKey)

	now := time.Now()
	key := &APIKey{
		ID:        uuid.New().String(),
		Name:      name,
		KeyHash:   hash,
		KeyPrefix: rawKey[:8],
		TenantID:  tenantID,
		ClientID:  clientID,
		Scopes:    scopes,
		RateLimit: rateLimit,
		Status:    "active",
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}

	if err := m.store.CreateKey(ctx, key); err != nil {
		return nil, "", fmt.Errorf("storing key: %w", err)
	}

	// Return a copy so the caller cannot mutate the store's copy.
	returned, err := m.store.GetByID(ctx, key.ID)
	if err != nil {
		return nil, "", fmt.Errorf("retrieving created key: %w", err)
	}
	return returned, rawKey, nil
}

// ValidateKey hashes the provided raw key, looks it up in the store, and
// verifies the key is active and not expired. On success it updates
// LastUsedAt and returns the APIKey.
func (m *APIKeyManager) ValidateKey(ctx context.Context, rawKey string) (*APIKey, error) {
	hash := hashKey(rawKey)

	key, err := m.store.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return nil, ErrInvalidKey
		}
		return nil, fmt.Errorf("looking up key: %w", err)
	}

	// Check status.
	if key.Status == "revoked" {
		return nil, ErrKeyRevoked
	}

	// Check expiry.
	if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
		return nil, ErrKeyExpired
	}

	// Update last used timestamp.
	now := time.Now()
	key.LastUsedAt = &now
	if err := m.store.UpdateKey(ctx, key); err != nil {
		// Non-fatal: log but don't fail the request.
		_ = err
	}

	return key, nil
}

// RevokeKey sets the key's status to "revoked" and records the revocation
// timestamp. The operation is idempotent: revoking an already-revoked key
// succeeds silently.
func (m *APIKeyManager) RevokeKey(ctx context.Context, id string) error {
	key, err := m.store.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if key.Status == "revoked" {
		return nil // idempotent
	}

	now := time.Now()
	key.Status = "revoked"
	key.RevokedAt = &now
	return m.store.UpdateKey(ctx, key)
}

// RotateKey revokes the existing key and creates a new one with the same
// configuration (name, tenant, client, scopes, rate limit). Returns the
// new APIKey and the raw key string.
func (m *APIKeyManager) RotateKey(ctx context.Context, id string) (*APIKey, string, error) {
	old, err := m.store.GetByID(ctx, id)
	if err != nil {
		return nil, "", err
	}

	// Revoke old key.
	if err := m.RevokeKey(ctx, id); err != nil {
		return nil, "", fmt.Errorf("revoking old key: %w", err)
	}

	// Create new key with the same configuration.
	return m.GenerateKey(ctx, old.Name, old.TenantID, old.ClientID, old.Scopes, old.RateLimit, old.ExpiresAt)
}

// ListKeys returns API keys for the given tenant with pagination.
func (m *APIKeyManager) ListKeys(ctx context.Context, tenantID string, limit, offset int) ([]*APIKey, int, error) {
	return m.store.ListByTenant(ctx, tenantID, limit, offset)
}

// generateRawKey produces a cryptographically random key string with the
// platform prefix: ehr_k1_<32-hex-chars>.
func generateRawKey() (string, error) {
	b := make([]byte, apiKeyRandomBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return apiKeyPrefix + hex.EncodeToString(b), nil
}

// hashKey returns the hex-encoded SHA-256 hash of the raw key string.
func hashKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

// APIKeyMiddlewareOption configures the API key middleware.
type APIKeyMiddlewareOption func(*apiKeyMiddlewareCfg)

type apiKeyMiddlewareCfg struct {
	enforceScopes bool
}

// WithScopeEnforcement enables or disables SMART scope enforcement in the
// API key middleware. When enabled, the middleware extracts the FHIR resource
// type from the request path and verifies that the key's scopes permit the
// requested operation.
func WithScopeEnforcement(enforce bool) APIKeyMiddlewareOption {
	return func(cfg *apiKeyMiddlewareCfg) {
		cfg.enforceScopes = enforce
	}
}

// APIKeyMiddleware returns an Echo middleware that authenticates requests
// using API keys. It checks the X-API-Key header first, then falls back to
// Authorization: Bearer if the token starts with the ehr_k1_ prefix.
//
// If a regular JWT Bearer token is present (i.e., not starting with ehr_k1_),
// the middleware passes the request through to the next handler so that the
// standard OAuth/JWT flow can take over.
func APIKeyMiddleware(manager *APIKeyManager, opts ...APIKeyMiddlewareOption) echo.MiddlewareFunc {
	cfg := &apiKeyMiddlewareCfg{}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rawKey := extractAPIKey(c)

			// No API key found — check if there is a non-API-key Bearer token.
			if rawKey == "" {
				authHeader := c.Request().Header.Get("Authorization")
				if authHeader != "" {
					parts := strings.SplitN(authHeader, " ", 2)
					if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") && !strings.HasPrefix(parts[1], apiKeyPrefix) {
						// Regular JWT — skip API key middleware.
						return next(c)
					}
				}
				// No credentials at all — let downstream middleware handle it.
				return next(c)
			}

			key, err := manager.ValidateKey(c.Request().Context(), rawKey)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidKey):
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid api key")
				case errors.Is(err, ErrKeyRevoked):
					return echo.NewHTTPError(http.StatusUnauthorized, "api key revoked")
				case errors.Is(err, ErrKeyExpired):
					return echo.NewHTTPError(http.StatusUnauthorized, "api key expired")
				default:
					return echo.NewHTTPError(http.StatusInternalServerError, "api key validation error")
				}
			}

			// Scope enforcement (optional).
			if cfg.enforceScopes && len(key.Scopes) > 0 {
				if err := enforceAPIKeyScopes(c, key.Scopes); err != nil {
					return err
				}
			}

			// Populate echo context for downstream handlers.
			c.Set("api_key_id", key.ID)
			c.Set("tenant_id", key.TenantID)
			c.Set("client_id", key.ClientID)
			c.Set("scopes", key.Scopes)

			return next(c)
		}
	}
}

// extractAPIKey returns the raw API key from the request, checking X-API-Key
// header first and then the Authorization: Bearer header.
func extractAPIKey(c echo.Context) string {
	// 1. X-API-Key header.
	if apiKey := c.Request().Header.Get("X-API-Key"); apiKey != "" {
		return apiKey
	}

	// 2. Authorization: Bearer ehr_k1_...
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	token := parts[1]
	if strings.HasPrefix(token, apiKeyPrefix) {
		return token
	}
	return ""
}

// enforceAPIKeyScopes checks that the key's scopes permit the requested
// FHIR operation. It extracts the resource type from the URL path and the
// operation from the HTTP method.
func enforceAPIKeyScopes(c echo.Context, scopes []string) error {
	// Try to determine the FHIR resource type from the path.
	path := c.Request().URL.Path
	resourceType := extractFHIRResourceType(path)
	if resourceType == "" {
		// Non-FHIR endpoint — no scope enforcement.
		return nil
	}

	operation := httpMethodToScopeOperation(c.Request().Method)

	// Parse the key's scopes and check.
	smartScopes := ParseSMARTScopes(scopes)
	if ScopeAllows(smartScopes, resourceType, operation) {
		return nil
	}

	return echo.NewHTTPError(http.StatusForbidden,
		fmt.Sprintf("insufficient scope: requires %s.%s", resourceType, operation))
}

// extractFHIRResourceType extracts a FHIR resource type from URL paths like
// /fhir/Patient/123 or /fhir/Observation.
func extractFHIRResourceType(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "fhir" {
		return parts[1]
	}
	// Also support paths like /Patient, /Patient/123 for non-prefixed routes.
	if len(parts) >= 1 {
		candidate := parts[0]
		if len(candidate) > 0 && candidate[0] >= 'A' && candidate[0] <= 'Z' {
			return candidate
		}
	}
	return ""
}

// httpMethodToScopeOperation maps HTTP methods to SMART scope operations.
func httpMethodToScopeOperation(method string) string {
	switch method {
	case http.MethodGet, http.MethodHead:
		return "read"
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return "write"
	default:
		return "read"
	}
}

// ---------------------------------------------------------------------------
// HTTP Handler
// ---------------------------------------------------------------------------

// APIKeyHandler provides Echo HTTP handlers for API key management endpoints.
type APIKeyHandler struct {
	manager *APIKeyManager
}

// NewAPIKeyHandler creates a new handler backed by the given manager.
func NewAPIKeyHandler(manager *APIKeyManager) *APIKeyHandler {
	return &APIKeyHandler{manager: manager}
}

// RegisterRoutes registers the API key management routes on the given Echo group.
func (h *APIKeyHandler) RegisterRoutes(g *echo.Group) {
	g.POST("", h.CreateKey)
	g.GET("", h.ListKeys)
	g.GET("/:id", h.GetKey)
	g.DELETE("/:id", h.RevokeKey)
	g.POST("/:id/rotate", h.RotateKey)
}

// createKeyRequest is the JSON request body for creating a new API key.
type createKeyRequest struct {
	Name      string            `json:"name"`
	TenantID  string            `json:"tenant_id"`
	ClientID  string            `json:"client_id"`
	Scopes    []string          `json:"scopes"`
	RateLimit int               `json:"rate_limit"`
	ExpiresAt *time.Time        `json:"expires_at,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// apiKeyResponse is the sanitized JSON representation of an APIKey that never
// exposes the KeyHash.
type apiKeyResponse struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	KeyPrefix  string            `json:"key_prefix"`
	TenantID   string            `json:"tenant_id"`
	ClientID   string            `json:"client_id"`
	Scopes     []string          `json:"scopes"`
	RateLimit  int               `json:"rate_limit"`
	Status     string            `json:"status"`
	ExpiresAt  *time.Time        `json:"expires_at,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	RevokedAt  *time.Time        `json:"revoked_at,omitempty"`
	LastUsedAt *time.Time        `json:"last_used_at,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

func toAPIKeyResponse(k *APIKey) *apiKeyResponse {
	return &apiKeyResponse{
		ID:         k.ID,
		Name:       k.Name,
		KeyPrefix:  k.KeyPrefix,
		TenantID:   k.TenantID,
		ClientID:   k.ClientID,
		Scopes:     k.Scopes,
		RateLimit:  k.RateLimit,
		Status:     k.Status,
		ExpiresAt:  k.ExpiresAt,
		CreatedAt:  k.CreatedAt,
		RevokedAt:  k.RevokedAt,
		LastUsedAt: k.LastUsedAt,
		Metadata:   k.Metadata,
	}
}

// CreateKey handles POST /api-keys. It creates a new API key and returns the
// raw key string exactly once in the response.
func (h *APIKeyHandler) CreateKey(c echo.Context) error {
	var req createKeyRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}

	key, rawKey, err := h.manager.GenerateKey(
		c.Request().Context(),
		req.Name, req.TenantID, req.ClientID,
		req.Scopes, req.RateLimit, req.ExpiresAt,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create api key")
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"key":     toAPIKeyResponse(key),
		"raw_key": rawKey,
		"warning": "Store this key securely. It will not be shown again.",
	})
}

// ListKeys handles GET /api-keys. Returns keys for the specified tenant,
// never exposing the key hash.
func (h *APIKeyHandler) ListKeys(c echo.Context) error {
	tenantID := c.QueryParam("tenant_id")
	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	limit := 50
	offset := 0
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}
	if offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
			offset = v
		}
	}

	keys, total, err := h.manager.ListKeys(c.Request().Context(), tenantID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list api keys")
	}

	responses := make([]*apiKeyResponse, len(keys))
	for i, k := range keys {
		responses[i] = toAPIKeyResponse(k)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"keys":   responses,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetKey handles GET /api-keys/:id.
func (h *APIKeyHandler) GetKey(c echo.Context) error {
	id := c.Param("id")

	key, err := h.manager.store.GetByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "api key not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve api key")
	}

	return c.JSON(http.StatusOK, toAPIKeyResponse(key))
}

// RevokeKey handles DELETE /api-keys/:id.
func (h *APIKeyHandler) RevokeKey(c echo.Context) error {
	id := c.Param("id")

	if err := h.manager.RevokeKey(c.Request().Context(), id); err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "api key not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to revoke api key")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "revoked",
		"message": "api key has been revoked",
	})
}

// RotateKey handles POST /api-keys/:id/rotate.
func (h *APIKeyHandler) RotateKey(c echo.Context) error {
	id := c.Param("id")

	newKey, rawKey, err := h.manager.RotateKey(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "api key not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to rotate api key")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"key":     toAPIKeyResponse(newKey),
		"raw_key": rawKey,
		"warning": "Store this key securely. It will not be shown again.",
	})
}
