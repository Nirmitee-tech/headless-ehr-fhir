package auth

import (
	"sync"
	"time"
)

// revocationEntry stores metadata about a revoked JWT token.
type revocationEntry struct {
	ExpiresAt time.Time
	UserID    string
}

// TokenRevocationStore manages revoked JWT tokens in memory.
// Revoked token JTIs (JWT ID claims) are stored with automatic
// cleanup of expired entries. Thread-safe for concurrent access.
type TokenRevocationStore struct {
	mu       sync.RWMutex
	entries  map[string]revocationEntry // JTI -> entry
	userJTIs map[string][]string        // userID -> []JTI
	done     chan struct{}
}

// NewTokenRevocationStore creates a new store and starts a background
// goroutine that cleans up expired entries every 5 minutes.
func NewTokenRevocationStore() *TokenRevocationStore {
	s := &TokenRevocationStore{
		entries:  make(map[string]revocationEntry),
		userJTIs: make(map[string][]string),
		done:     make(chan struct{}),
	}
	go s.cleanupLoop()
	return s
}

// Revoke adds a token's JTI to the revocation list. The expiresAt time
// indicates when the token would have naturally expired; the entry is
// automatically cleaned up after that time since there is no need to
// track a revocation once the token is expired anyway.
func (s *TokenRevocationStore) Revoke(jti string, expiresAt time.Time) {
	s.RevokeForUser(jti, "", expiresAt)
}

// RevokeForUser adds a token's JTI to the revocation list and associates
// it with a user ID for bulk revocation support.
func (s *TokenRevocationStore) RevokeForUser(jti, userID string, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[jti] = revocationEntry{
		ExpiresAt: expiresAt,
		UserID:    userID,
	}

	if userID != "" {
		s.userJTIs[userID] = append(s.userJTIs[userID], jti)
	}
}

// IsRevoked checks if a token JTI has been revoked.
func (s *TokenRevocationStore) IsRevoked(jti string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.entries[jti]
	return ok
}

// RevokeAllForUser revokes all known tokens for a specific user ID by
// marking every JTI associated with that user as revoked. Returns the
// number of tokens that were revoked.
func (s *TokenRevocationStore) RevokeAllForUser(userID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	jtis, ok := s.userJTIs[userID]
	if !ok {
		return 0
	}

	count := 0
	for _, jti := range jtis {
		if _, exists := s.entries[jti]; exists {
			count++
		}
	}
	return count
}

// Count returns the number of currently revoked tokens.
func (s *TokenRevocationStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.entries)
}

// Entries returns a snapshot of all current revocation entries.
// Each entry includes the JTI, user ID, and expiry time.
func (s *TokenRevocationStore) Entries() []RevocationInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]RevocationInfo, 0, len(s.entries))
	for jti, entry := range s.entries {
		result = append(result, RevocationInfo{
			JTI:       jti,
			UserID:    entry.UserID,
			ExpiresAt: entry.ExpiresAt,
		})
	}
	return result
}

// RevocationInfo is a public representation of a revocation entry.
type RevocationInfo struct {
	JTI       string    `json:"jti"`
	UserID    string    `json:"user_id,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Close stops the background cleanup goroutine. It is safe to call
// multiple times but only the first call has effect.
func (s *TokenRevocationStore) Close() {
	select {
	case <-s.done:
		// already closed
	default:
		close(s.done)
	}
}

// cleanupLoop periodically removes expired revocation entries.
func (s *TokenRevocationStore) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

// cleanup removes revocation entries whose tokens have expired.
// Once a token is past its natural expiry there is no need to keep
// tracking it in the revocation list.
func (s *TokenRevocationStore) cleanup() {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for jti, entry := range s.entries {
		if now.After(entry.ExpiresAt) {
			userID := entry.UserID
			delete(s.entries, jti)

			// Clean up user -> JTI mapping
			if userID != "" {
				jtis := s.userJTIs[userID]
				for i, id := range jtis {
					if id == jti {
						s.userJTIs[userID] = append(jtis[:i], jtis[i+1:]...)
						break
					}
				}
				if len(s.userJTIs[userID]) == 0 {
					delete(s.userJTIs, userID)
				}
			}
		}
	}
}
