package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestManager(t *testing.T) *APIKeyManager {
	t.Helper()
	store := NewInMemoryAPIKeyStore()
	return NewAPIKeyManager(store)
}

// ---------------------------------------------------------------------------
// Key generation
// ---------------------------------------------------------------------------

func TestAPIKeyManager_GenerateKey(t *testing.T) {
	mgr := newTestManager(t)
	key, rawKey, err := mgr.GenerateKey(context.Background(), "Test Key", "tenant-1", "client-1", []string{"patient/*.read"}, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == nil {
		t.Fatal("expected key, got nil")
	}
	if rawKey == "" {
		t.Fatal("expected raw key, got empty string")
	}
	if !strings.HasPrefix(rawKey, "ehr_k1_") {
		t.Errorf("expected raw key to have prefix ehr_k1_, got %s", rawKey)
	}
	if key.ID == "" {
		t.Error("expected key ID to be set")
	}
	if key.Name != "Test Key" {
		t.Errorf("expected name 'Test Key', got %q", key.Name)
	}
	if key.TenantID != "tenant-1" {
		t.Errorf("expected tenant-1, got %s", key.TenantID)
	}
	if key.ClientID != "client-1" {
		t.Errorf("expected client-1, got %s", key.ClientID)
	}
	if key.Status != "active" {
		t.Errorf("expected status active, got %s", key.Status)
	}
	if key.KeyPrefix == "" {
		t.Error("expected key prefix to be set")
	}
	if len(key.KeyPrefix) != 8 {
		t.Errorf("expected key prefix length 8, got %d", len(key.KeyPrefix))
	}
}

func TestAPIKeyManager_GenerateKey_StoresHash(t *testing.T) {
	mgr := newTestManager(t)
	key, rawKey, err := mgr.GenerateKey(context.Background(), "Test Key", "tenant-1", "client-1", nil, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key.KeyHash == "" {
		t.Fatal("expected key hash to be set")
	}
	if key.KeyHash == rawKey {
		t.Error("key hash must not equal raw key (plaintext stored!)")
	}
}

func TestAPIKeyManager_GenerateKey_UniqueKeys(t *testing.T) {
	mgr := newTestManager(t)
	_, raw1, err := mgr.GenerateKey(context.Background(), "Key A", "tenant-1", "client-1", nil, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, raw2, err := mgr.GenerateKey(context.Background(), "Key B", "tenant-1", "client-1", nil, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw1 == raw2 {
		t.Error("two generated keys must be different")
	}
}

func TestAPIKeyManager_GenerateKey_WithExpiry(t *testing.T) {
	mgr := newTestManager(t)
	exp := time.Now().Add(24 * time.Hour)
	key, _, err := mgr.GenerateKey(context.Background(), "Expiring Key", "tenant-1", "client-1", nil, 0, &exp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt to be set")
	}
	if !key.ExpiresAt.Equal(exp) {
		t.Errorf("expected ExpiresAt=%v, got %v", exp, *key.ExpiresAt)
	}
}

func TestAPIKeyManager_GenerateKey_WithScopes(t *testing.T) {
	mgr := newTestManager(t)
	scopes := []string{"patient/*.read", "system/*.write"}
	key, _, err := mgr.GenerateKey(context.Background(), "Scoped Key", "tenant-1", "client-1", scopes, 100, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(key.Scopes) != 2 {
		t.Fatalf("expected 2 scopes, got %d", len(key.Scopes))
	}
	if key.Scopes[0] != "patient/*.read" || key.Scopes[1] != "system/*.write" {
		t.Errorf("unexpected scopes: %v", key.Scopes)
	}
	if key.RateLimit != 100 {
		t.Errorf("expected rate limit 100, got %d", key.RateLimit)
	}
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

func TestAPIKeyManager_ValidateKey(t *testing.T) {
	mgr := newTestManager(t)
	_, rawKey, err := mgr.GenerateKey(context.Background(), "Valid Key", "tenant-1", "client-1", []string{"patient/*.read"}, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	validated, err := mgr.ValidateKey(context.Background(), rawKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if validated == nil {
		t.Fatal("expected validated key, got nil")
	}
	if validated.Name != "Valid Key" {
		t.Errorf("expected name 'Valid Key', got %q", validated.Name)
	}
}

func TestAPIKeyManager_ValidateKey_Invalid(t *testing.T) {
	mgr := newTestManager(t)
	_, err := mgr.ValidateKey(context.Background(), "ehr_k1_invalidkeyvalue1234567890abcdef")
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
	if err != ErrInvalidKey {
		t.Errorf("expected ErrInvalidKey, got %v", err)
	}
}

func TestAPIKeyManager_ValidateKey_Revoked(t *testing.T) {
	mgr := newTestManager(t)
	key, rawKey, err := mgr.GenerateKey(context.Background(), "Revoke Me", "tenant-1", "client-1", nil, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mgr.RevokeKey(context.Background(), key.ID); err != nil {
		t.Fatalf("unexpected error revoking: %v", err)
	}

	_, err = mgr.ValidateKey(context.Background(), rawKey)
	if err == nil {
		t.Fatal("expected error for revoked key")
	}
	if err != ErrKeyRevoked {
		t.Errorf("expected ErrKeyRevoked, got %v", err)
	}
}

func TestAPIKeyManager_ValidateKey_Expired(t *testing.T) {
	mgr := newTestManager(t)
	exp := time.Now().Add(-1 * time.Hour) // already expired
	_, rawKey, err := mgr.GenerateKey(context.Background(), "Expired Key", "tenant-1", "client-1", nil, 0, &exp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = mgr.ValidateKey(context.Background(), rawKey)
	if err == nil {
		t.Fatal("expected error for expired key")
	}
	if err != ErrKeyExpired {
		t.Errorf("expected ErrKeyExpired, got %v", err)
	}
}

func TestAPIKeyManager_ValidateKey_UpdatesLastUsed(t *testing.T) {
	mgr := newTestManager(t)
	key, rawKey, err := mgr.GenerateKey(context.Background(), "Track Usage", "tenant-1", "client-1", nil, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key.LastUsedAt != nil {
		t.Error("expected LastUsedAt to be nil initially")
	}

	before := time.Now()
	_, err = mgr.ValidateKey(context.Background(), rawKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Re-fetch from store to verify LastUsedAt was persisted
	updated, err := mgr.store.GetByID(context.Background(), key.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching key: %v", err)
	}
	if updated.LastUsedAt == nil {
		t.Fatal("expected LastUsedAt to be set after validation")
	}
	if updated.LastUsedAt.Before(before) {
		t.Error("expected LastUsedAt to be after the validation call")
	}
}

// ---------------------------------------------------------------------------
// Revocation
// ---------------------------------------------------------------------------

func TestAPIKeyManager_RevokeKey(t *testing.T) {
	mgr := newTestManager(t)
	key, _, err := mgr.GenerateKey(context.Background(), "Revoke Me", "tenant-1", "client-1", nil, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mgr.RevokeKey(context.Background(), key.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	revoked, err := mgr.store.GetByID(context.Background(), key.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if revoked.Status != "revoked" {
		t.Errorf("expected status revoked, got %s", revoked.Status)
	}
	if revoked.RevokedAt == nil {
		t.Error("expected RevokedAt to be set")
	}
}

func TestAPIKeyManager_RevokeKey_NotFound(t *testing.T) {
	mgr := newTestManager(t)
	err := mgr.RevokeKey(context.Background(), "non-existent-id")
	if err == nil {
		t.Fatal("expected error for non-existent key")
	}
	if err != ErrKeyNotFound {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestAPIKeyManager_RevokeKey_AlreadyRevoked(t *testing.T) {
	mgr := newTestManager(t)
	key, _, err := mgr.GenerateKey(context.Background(), "Revoke Twice", "tenant-1", "client-1", nil, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mgr.RevokeKey(context.Background(), key.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second revocation should be idempotent (no error)
	if err := mgr.RevokeKey(context.Background(), key.ID); err != nil {
		t.Fatalf("expected idempotent revoke, got error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Rotation
// ---------------------------------------------------------------------------

func TestAPIKeyManager_RotateKey(t *testing.T) {
	mgr := newTestManager(t)
	scopes := []string{"patient/*.read", "system/*.write"}
	oldKey, oldRaw, err := mgr.GenerateKey(context.Background(), "Rotate Me", "tenant-1", "client-1", scopes, 50, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	newKey, newRaw, err := mgr.RotateKey(context.Background(), oldKey.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Old key should be revoked
	old, _ := mgr.store.GetByID(context.Background(), oldKey.ID)
	if old.Status != "revoked" {
		t.Errorf("expected old key to be revoked, got %s", old.Status)
	}

	// New key should have same tenant, client, scopes, and rate limit
	if newKey.TenantID != oldKey.TenantID {
		t.Errorf("expected same tenant %s, got %s", oldKey.TenantID, newKey.TenantID)
	}
	if newKey.ClientID != oldKey.ClientID {
		t.Errorf("expected same client %s, got %s", oldKey.ClientID, newKey.ClientID)
	}
	if len(newKey.Scopes) != len(scopes) {
		t.Errorf("expected %d scopes, got %d", len(scopes), len(newKey.Scopes))
	}
	if newKey.RateLimit != 50 {
		t.Errorf("expected rate limit 50, got %d", newKey.RateLimit)
	}
	if newKey.Status != "active" {
		t.Errorf("expected new key active, got %s", newKey.Status)
	}
	if newRaw == oldRaw {
		t.Error("new raw key must differ from old raw key")
	}
	if newKey.ID == oldKey.ID {
		t.Error("new key must have a different ID")
	}
}

func TestAPIKeyManager_RotateKey_NotFound(t *testing.T) {
	mgr := newTestManager(t)
	_, _, err := mgr.RotateKey(context.Background(), "non-existent-id")
	if err == nil {
		t.Fatal("expected error for non-existent key")
	}
	if err != ErrKeyNotFound {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Listing
// ---------------------------------------------------------------------------

func TestAPIKeyManager_ListKeys(t *testing.T) {
	mgr := newTestManager(t)
	for i := 0; i < 3; i++ {
		_, _, err := mgr.GenerateKey(context.Background(), "Key", "tenant-1", "client-1", nil, 0, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	keys, total, err := mgr.ListKeys(context.Background(), "tenant-1", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
}

func TestAPIKeyManager_ListKeys_Pagination(t *testing.T) {
	mgr := newTestManager(t)
	for i := 0; i < 5; i++ {
		_, _, err := mgr.GenerateKey(context.Background(), "Key", "tenant-1", "client-1", nil, 0, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	keys, total, err := mgr.ListKeys(context.Background(), "tenant-1", 2, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys (limit=2), got %d", len(keys))
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}

	keys2, total2, err := mgr.ListKeys(context.Background(), "tenant-1", 2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys2) != 2 {
		t.Errorf("expected 2 keys (limit=2, offset=2), got %d", len(keys2))
	}
	if total2 != 5 {
		t.Errorf("expected total 5, got %d", total2)
	}

	keys3, _, err := mgr.ListKeys(context.Background(), "tenant-1", 2, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys3) != 1 {
		t.Errorf("expected 1 key (limit=2, offset=4), got %d", len(keys3))
	}
}

func TestAPIKeyManager_ListKeys_ExcludesOtherTenants(t *testing.T) {
	mgr := newTestManager(t)
	_, _, _ = mgr.GenerateKey(context.Background(), "T1 Key", "tenant-1", "client-1", nil, 0, nil)
	_, _, _ = mgr.GenerateKey(context.Background(), "T2 Key", "tenant-2", "client-2", nil, 0, nil)
	_, _, _ = mgr.GenerateKey(context.Background(), "T1 Key 2", "tenant-1", "client-1", nil, 0, nil)

	keys, total, err := mgr.ListKeys(context.Background(), "tenant-1", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys for tenant-1, got %d", len(keys))
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	for _, k := range keys {
		if k.TenantID != "tenant-1" {
			t.Errorf("expected tenant-1, got %s", k.TenantID)
		}
	}
}

// ---------------------------------------------------------------------------
// Store
// ---------------------------------------------------------------------------

func TestInMemoryAPIKeyStore_CRUD(t *testing.T) {
	store := NewInMemoryAPIKeyStore()
	ctx := context.Background()

	key := &APIKey{
		ID:        "test-id-1",
		Name:      "Test Key",
		KeyHash:   "somehash",
		TenantID:  "tenant-1",
		ClientID:  "client-1",
		Status:    "active",
		CreatedAt: time.Now(),
	}

	// Create
	if err := store.CreateKey(ctx, key); err != nil {
		t.Fatalf("CreateKey: %v", err)
	}

	// Read by ID
	got, err := store.GetByID(ctx, "test-id-1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "Test Key" {
		t.Errorf("expected name 'Test Key', got %q", got.Name)
	}

	// Read by hash
	gotHash, err := store.GetByHash(ctx, "somehash")
	if err != nil {
		t.Fatalf("GetByHash: %v", err)
	}
	if gotHash.ID != "test-id-1" {
		t.Errorf("expected ID test-id-1, got %s", gotHash.ID)
	}

	// Update
	key.Name = "Updated Key"
	if err := store.UpdateKey(ctx, key); err != nil {
		t.Fatalf("UpdateKey: %v", err)
	}
	updated, _ := store.GetByID(ctx, "test-id-1")
	if updated.Name != "Updated Key" {
		t.Errorf("expected updated name, got %q", updated.Name)
	}

	// List by tenant
	keys, total, err := store.ListByTenant(ctx, "tenant-1", 10, 0)
	if err != nil {
		t.Fatalf("ListByTenant: %v", err)
	}
	if len(keys) != 1 || total != 1 {
		t.Errorf("expected 1 key, got %d (total %d)", len(keys), total)
	}

	// List by client
	clientKeys, err := store.ListByClient(ctx, "client-1")
	if err != nil {
		t.Fatalf("ListByClient: %v", err)
	}
	if len(clientKeys) != 1 {
		t.Errorf("expected 1 key for client, got %d", len(clientKeys))
	}

	// Delete
	if err := store.DeleteKey(ctx, "test-id-1"); err != nil {
		t.Fatalf("DeleteKey: %v", err)
	}
	_, err = store.GetByID(ctx, "test-id-1")
	if err != ErrKeyNotFound {
		t.Errorf("expected ErrKeyNotFound after delete, got %v", err)
	}
}

func TestInMemoryAPIKeyStore_ConcurrentAccess(t *testing.T) {
	store := NewInMemoryAPIKeyStore()
	ctx := context.Background()

	var wg sync.WaitGroup
	errCh := make(chan error, 200)

	// Concurrent writes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := &APIKey{
				ID:        "concurrent-" + time.Now().String() + string(rune('a'+idx%26)),
				Name:      "Concurrent Key",
				KeyHash:   "hash-concurrent-" + time.Now().String() + string(rune('a'+idx%26)),
				TenantID:  "tenant-1",
				ClientID:  "client-1",
				Status:    "active",
				CreatedAt: time.Now(),
			}
			if err := store.CreateKey(ctx, key); err != nil {
				errCh <- err
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, _ = store.ListByTenant(ctx, "tenant-1", 100, 0)
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

func TestAPIKeyMiddleware_ValidKey(t *testing.T) {
	mgr := newTestManager(t)
	_, rawKey, err := mgr.GenerateKey(context.Background(), "MW Key", "tenant-1", "client-1", []string{"patient/*.read"}, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("X-API-Key", rawKey)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "ok")
	}

	mw := APIKeyMiddleware(mgr)
	h := mw(handler)
	err = h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handlerCalled {
		t.Error("handler was not called")
	}
}

func TestAPIKeyMiddleware_BearerKey(t *testing.T) {
	mgr := newTestManager(t)
	_, rawKey, err := mgr.GenerateKey(context.Background(), "Bearer Key", "tenant-1", "client-1", []string{"patient/*.read"}, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Authorization", "Bearer "+rawKey)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "ok")
	}

	mw := APIKeyMiddleware(mgr)
	h := mw(handler)
	err = h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handlerCalled {
		t.Error("handler was not called")
	}
}

func TestAPIKeyMiddleware_InvalidKey(t *testing.T) {
	mgr := newTestManager(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("X-API-Key", "ehr_k1_invalidkeyvalue1234567890abcdef")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := APIKeyMiddleware(mgr)
	h := mw(handler)
	err := h(c)

	if err == nil {
		t.Fatal("expected error for invalid key")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", httpErr.Code)
	}
}

func TestAPIKeyMiddleware_InsufficientScopes(t *testing.T) {
	mgr := newTestManager(t)
	// Key only has read scope
	_, rawKey, err := mgr.GenerateKey(context.Background(), "Read Only", "tenant-1", "client-1", []string{"patient/*.read"}, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e := echo.New()
	// POST = write operation
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", strings.NewReader(`{}`))
	req.Header.Set("X-API-Key", rawKey)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/:resource")
	c.SetParamNames("resource")
	c.SetParamValues("Patient")

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := APIKeyMiddleware(mgr, WithScopeEnforcement(true))
	h := mw(handler)
	err = h(c)

	if err == nil {
		t.Fatal("expected error for insufficient scopes")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", httpErr.Code)
	}
}

func TestAPIKeyMiddleware_SkipsJWT(t *testing.T) {
	mgr := newTestManager(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	// Regular JWT bearer token (not ehr_k1_ prefix)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.fake")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "ok")
	}

	mw := APIKeyMiddleware(mgr)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handlerCalled {
		t.Error("handler was not called; should have skipped to next for JWT")
	}
}

func TestAPIKeyMiddleware_SetsContext(t *testing.T) {
	mgr := newTestManager(t)
	_, rawKey, err := mgr.GenerateKey(context.Background(), "Context Key", "tenant-ctx", "client-ctx", []string{"patient/*.read"}, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("X-API-Key", rawKey)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		apiKeyID, _ := c.Get("api_key_id").(string)
		tenantID, _ := c.Get("tenant_id").(string)
		clientID, _ := c.Get("client_id").(string)
		scopes, _ := c.Get("scopes").([]string)

		if apiKeyID == "" {
			t.Error("expected api_key_id to be set")
		}
		if tenantID != "tenant-ctx" {
			t.Errorf("expected tenant_id=tenant-ctx, got %s", tenantID)
		}
		if clientID != "client-ctx" {
			t.Errorf("expected client_id=client-ctx, got %s", clientID)
		}
		if len(scopes) != 1 || scopes[0] != "patient/*.read" {
			t.Errorf("expected scopes=[patient/*.read], got %v", scopes)
		}
		return c.String(http.StatusOK, "ok")
	}

	mw := APIKeyMiddleware(mgr)
	h := mw(handler)
	err = h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------------

func TestAPIKeyHandler_CreateKey(t *testing.T) {
	mgr := newTestManager(t)
	h := NewAPIKeyHandler(mgr)

	e := echo.New()
	body := `{"name":"New Key","tenant_id":"tenant-1","client_id":"client-1","scopes":["patient/*.read"]}`
	req := httptest.NewRequest(http.MethodPost, "/api-keys", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateKey(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	rawKey, ok := resp["raw_key"].(string)
	if !ok || rawKey == "" {
		t.Error("expected raw_key in response")
	}
	if !strings.HasPrefix(rawKey, "ehr_k1_") {
		t.Errorf("expected raw_key with prefix ehr_k1_, got %s", rawKey)
	}
	keyObj, ok := resp["key"].(map[string]interface{})
	if !ok {
		t.Fatal("expected key object in response")
	}
	if keyObj["key_hash"] != nil {
		t.Error("key_hash must not be exposed in response")
	}
}

func TestAPIKeyHandler_ListKeys(t *testing.T) {
	mgr := newTestManager(t)
	h := NewAPIKeyHandler(mgr)

	// Create some keys
	_, _, _ = mgr.GenerateKey(context.Background(), "Key 1", "tenant-list", "client-1", nil, 0, nil)
	_, _, _ = mgr.GenerateKey(context.Background(), "Key 2", "tenant-list", "client-1", nil, 0, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api-keys?tenant_id=tenant-list", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListKeys(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	keys, ok := resp["keys"].([]interface{})
	if !ok {
		t.Fatal("expected keys array in response")
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}

	// Ensure no key_hash exposed
	for _, k := range keys {
		km := k.(map[string]interface{})
		if km["key_hash"] != nil {
			t.Error("key_hash must not be exposed in list response")
		}
	}
}

func TestAPIKeyHandler_RevokeKey(t *testing.T) {
	mgr := newTestManager(t)
	h := NewAPIKeyHandler(mgr)

	key, _, _ := mgr.GenerateKey(context.Background(), "Revoke Via Handler", "tenant-1", "client-1", nil, 0, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api-keys/"+key.ID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api-keys/:id")
	c.SetParamNames("id")
	c.SetParamValues(key.ID)

	err := h.RevokeKey(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	// Verify it is revoked
	revoked, _ := mgr.store.GetByID(context.Background(), key.ID)
	if revoked.Status != "revoked" {
		t.Errorf("expected revoked, got %s", revoked.Status)
	}
}

func TestAPIKeyHandler_RotateKey(t *testing.T) {
	mgr := newTestManager(t)
	h := NewAPIKeyHandler(mgr)

	key, _, _ := mgr.GenerateKey(context.Background(), "Rotate Via Handler", "tenant-1", "client-1", []string{"patient/*.read"}, 0, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api-keys/"+key.ID+"/rotate", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api-keys/:id/rotate")
	c.SetParamNames("id")
	c.SetParamValues(key.ID)

	err := h.RotateKey(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	newRaw, ok := resp["raw_key"].(string)
	if !ok || newRaw == "" {
		t.Error("expected raw_key in rotation response")
	}
	if !strings.HasPrefix(newRaw, "ehr_k1_") {
		t.Errorf("expected ehr_k1_ prefix, got %s", newRaw)
	}
}
