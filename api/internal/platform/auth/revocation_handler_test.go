package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// newTestContextWithRole creates an echo.Context with the given HTTP method,
// path, body, and user roles injected into the request context.
func newTestContextWithRole(e *echo.Echo, method, path, body string, roles []string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), UserRolesKey, roles)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestHandleRevokeToken_Success(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	e := echo.New()
	g := e.Group("/api/v1")
	RegisterRevocationRoutes(g, store)

	body := `{"jti":"token-xyz","expires_at":"2099-01-01T00:00:00Z"}`
	c, rec := newTestContextWithRole(e, http.MethodPost, "/api/v1/auth/revoke", body, []string{"admin"})
	c.SetPath("/api/v1/auth/revoke")

	// Execute the handler directly (bypassing the role middleware for unit test)
	handler := handleRevokeToken(store)
	err := handler(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}

	if !store.IsRevoked("token-xyz") {
		t.Error("expected token-xyz to be revoked")
	}
}

func TestHandleRevokeToken_MissingJTI(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	e := echo.New()
	body := `{"expires_at":"2099-01-01T00:00:00Z"}`
	c, _ := newTestContextWithRole(e, http.MethodPost, "/api/v1/auth/revoke", body, []string{"admin"})

	handler := handleRevokeToken(store)
	err := handler(c)

	if err == nil {
		t.Fatal("expected error for missing JTI")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", httpErr.Code)
	}
}

func TestHandleRevokeToken_DefaultExpiry(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	e := echo.New()
	body := `{"jti":"token-default-expiry"}`
	c, rec := newTestContextWithRole(e, http.MethodPost, "/api/v1/auth/revoke", body, []string{"admin"})

	handler := handleRevokeToken(store)
	err := handler(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
	if !store.IsRevoked("token-default-expiry") {
		t.Error("expected token to be revoked with default expiry")
	}
}

func TestHandleRevokeToken_WithUserID(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	e := echo.New()
	body := `{"jti":"token-user","user_id":"user-42","expires_at":"2099-01-01T00:00:00Z"}`
	c, _ := newTestContextWithRole(e, http.MethodPost, "/api/v1/auth/revoke", body, []string{"admin"})

	handler := handleRevokeToken(store)
	err := handler(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !store.IsRevoked("token-user") {
		t.Error("expected token to be revoked with user association")
	}
}

func TestHandleRevokeUser_Success(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	// Pre-populate tokens for user-42
	store.RevokeForUser("jti-a", "user-42", time.Now().Add(1*time.Hour))
	store.RevokeForUser("jti-b", "user-42", time.Now().Add(1*time.Hour))

	e := echo.New()
	body := `{"user_id":"user-42"}`
	c, rec := newTestContextWithRole(e, http.MethodPost, "/api/v1/auth/revoke-user", body, []string{"admin"})

	handler := handleRevokeUser(store)
	err := handler(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp revokeUserResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.RevokedCount != 2 {
		t.Errorf("expected revoked_count=2, got %d", resp.RevokedCount)
	}
}

func TestHandleRevokeUser_MissingUserID(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	e := echo.New()
	body := `{}`
	c, _ := newTestContextWithRole(e, http.MethodPost, "/api/v1/auth/revoke-user", body, []string{"admin"})

	handler := handleRevokeUser(store)
	err := handler(c)

	if err == nil {
		t.Fatal("expected error for missing user_id")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", httpErr.Code)
	}
}

func TestHandleListRevocations(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	store.RevokeForUser("jti-1", "user-1", time.Now().Add(1*time.Hour))
	store.Revoke("jti-2", time.Now().Add(1*time.Hour))

	e := echo.New()
	c, rec := newTestContextWithRole(e, http.MethodGet, "/api/v1/auth/revocations", "", []string{"admin"})

	handler := handleListRevocations(store)
	err := handler(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp revocationListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Count != 2 {
		t.Errorf("expected count=2, got %d", resp.Count)
	}
	if len(resp.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(resp.Entries))
	}
}

func TestRevocationRoutes_NonAdminDenied(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	e := echo.New()
	g := e.Group("/api/v1")
	RegisterRevocationRoutes(g, store)

	// Simulate a request from a non-admin user through the full middleware chain
	endpoints := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/api/v1/auth/revoke", `{"jti":"x","expires_at":"2099-01-01T00:00:00Z"}`},
		{http.MethodPost, "/api/v1/auth/revoke-user", `{"user_id":"u"}`},
		{http.MethodGet, "/api/v1/auth/revocations", ""},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			var req *http.Request
			if ep.body != "" {
				req = httptest.NewRequest(ep.method, ep.path, strings.NewReader(ep.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(ep.method, ep.path, nil)
			}

			// Set a non-admin role
			ctx := context.WithValue(req.Context(), UserRolesKey, []string{"physician"})
			req = req.WithContext(ctx)

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Errorf("expected 403 for non-admin, got %d", rec.Code)
			}
		})
	}
}

func TestRevocationRoutes_AdminAllowed(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	e := echo.New()
	g := e.Group("/api/v1")
	RegisterRevocationRoutes(g, store)

	// Pre-populate a token for revoke-user test
	store.RevokeForUser("jti-pre", "user-1", time.Now().Add(1*time.Hour))

	endpoints := []struct {
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{http.MethodPost, "/api/v1/auth/revoke", `{"jti":"test-jti","expires_at":"2099-01-01T00:00:00Z"}`, http.StatusNoContent},
		{http.MethodPost, "/api/v1/auth/revoke-user", `{"user_id":"user-1"}`, http.StatusOK},
		{http.MethodGet, "/api/v1/auth/revocations", "", http.StatusOK},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			var req *http.Request
			if ep.body != "" {
				req = httptest.NewRequest(ep.method, ep.path, strings.NewReader(ep.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(ep.method, ep.path, nil)
			}

			ctx := context.WithValue(req.Context(), UserRolesKey, []string{"admin"})
			req = req.WithContext(ctx)

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != ep.wantStatus {
				t.Errorf("expected %d, got %d (body: %s)", ep.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
