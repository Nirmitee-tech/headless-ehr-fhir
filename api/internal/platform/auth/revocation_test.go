package auth

import (
	"sync"
	"testing"
	"time"
)

func TestRevoke_and_IsRevoked(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	jti := "token-abc-123"
	store.Revoke(jti, time.Now().Add(1*time.Hour))

	if !store.IsRevoked(jti) {
		t.Errorf("expected JTI %q to be revoked", jti)
	}
}

func TestIsRevoked_NotRevoked(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	if store.IsRevoked("unknown-jti") {
		t.Error("expected unknown JTI to not be revoked")
	}
}

func TestRevokeForUser(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	store.RevokeForUser("jti-1", "user-42", time.Now().Add(1*time.Hour))
	store.RevokeForUser("jti-2", "user-42", time.Now().Add(1*time.Hour))
	store.RevokeForUser("jti-3", "user-99", time.Now().Add(1*time.Hour))

	if !store.IsRevoked("jti-1") {
		t.Error("expected jti-1 to be revoked")
	}
	if !store.IsRevoked("jti-2") {
		t.Error("expected jti-2 to be revoked")
	}
	if !store.IsRevoked("jti-3") {
		t.Error("expected jti-3 to be revoked")
	}
}

func TestRevokeAllForUser(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	store.RevokeForUser("jti-1", "user-42", time.Now().Add(1*time.Hour))
	store.RevokeForUser("jti-2", "user-42", time.Now().Add(1*time.Hour))
	store.RevokeForUser("jti-3", "user-99", time.Now().Add(1*time.Hour))

	count := store.RevokeAllForUser("user-42")
	if count != 2 {
		t.Errorf("expected RevokeAllForUser to return 2, got %d", count)
	}

	// Both user-42 tokens should still be in the store as revoked
	if !store.IsRevoked("jti-1") {
		t.Error("expected jti-1 to remain revoked")
	}
	if !store.IsRevoked("jti-2") {
		t.Error("expected jti-2 to remain revoked")
	}
	// user-99 token should be unaffected
	if !store.IsRevoked("jti-3") {
		t.Error("expected jti-3 to remain revoked")
	}
}

func TestRevokeAllForUser_UnknownUser(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	count := store.RevokeAllForUser("nonexistent-user")
	if count != 0 {
		t.Errorf("expected 0 for unknown user, got %d", count)
	}
}

func TestCleanup_RemovesExpiredEntries(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	// Add one expired and one active entry
	store.RevokeForUser("expired-jti", "user-1", time.Now().Add(-1*time.Second))
	store.RevokeForUser("active-jti", "user-2", time.Now().Add(1*time.Hour))

	if store.Count() != 2 {
		t.Fatalf("expected 2 entries before cleanup, got %d", store.Count())
	}

	// Trigger manual cleanup
	store.cleanup()

	if store.Count() != 1 {
		t.Errorf("expected 1 entry after cleanup, got %d", store.Count())
	}

	if store.IsRevoked("expired-jti") {
		t.Error("expected expired JTI to be cleaned up")
	}
	if !store.IsRevoked("active-jti") {
		t.Error("expected active JTI to remain")
	}
}

func TestCleanup_RemovesUserMapping(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	store.RevokeForUser("expired-jti", "user-1", time.Now().Add(-1*time.Second))
	store.cleanup()

	// After cleanup, the user mapping should also be cleaned up
	store.mu.RLock()
	jtis, exists := store.userJTIs["user-1"]
	store.mu.RUnlock()

	if exists && len(jtis) > 0 {
		t.Errorf("expected user-1 JTI mapping to be cleaned up, found %v", jtis)
	}
}

func TestCount(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	if store.Count() != 0 {
		t.Errorf("expected 0 for empty store, got %d", store.Count())
	}

	store.Revoke("jti-1", time.Now().Add(1*time.Hour))
	store.Revoke("jti-2", time.Now().Add(1*time.Hour))

	if store.Count() != 2 {
		t.Errorf("expected 2, got %d", store.Count())
	}
}

func TestEntries(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	expiry := time.Now().Add(1 * time.Hour)
	store.RevokeForUser("jti-a", "user-1", expiry)
	store.Revoke("jti-b", expiry)

	entries := store.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	found := make(map[string]bool)
	for _, e := range entries {
		found[e.JTI] = true
	}
	if !found["jti-a"] || !found["jti-b"] {
		t.Errorf("expected both jti-a and jti-b in entries, got %v", entries)
	}
}

func TestConcurrentAccess(t *testing.T) {
	store := NewTokenRevocationStore()
	defer store.Close()

	var wg sync.WaitGroup
	const goroutines = 100

	// Half the goroutines revoke tokens, half check revocation
	wg.Add(goroutines * 2)

	for i := 0; i < goroutines; i++ {
		jti := "jti-" + time.Now().Format("150405.000000000") + "-" + string(rune('A'+i%26))
		go func(jti string) {
			defer wg.Done()
			store.Revoke(jti, time.Now().Add(1*time.Hour))
		}(jti)

		go func(jti string) {
			defer wg.Done()
			_ = store.IsRevoked(jti)
		}(jti)
	}

	wg.Wait()

	// No panics or data races = pass
	// Count should be > 0
	if store.Count() == 0 {
		t.Error("expected some entries after concurrent writes")
	}
}

func TestClose_StopsCleanupGoroutine(t *testing.T) {
	store := NewTokenRevocationStore()
	store.Close()

	// Closing again should not panic (idempotent)
	store.Close()

	// Store should still be usable after close (just no background cleanup)
	store.Revoke("jti-after-close", time.Now().Add(1*time.Hour))
	if !store.IsRevoked("jti-after-close") {
		t.Error("expected store to still work after Close")
	}
}
