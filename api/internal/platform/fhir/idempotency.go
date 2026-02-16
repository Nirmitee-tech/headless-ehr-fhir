package fhir

import (
	"bytes"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// DefaultIdempotencyTTL is the default time-to-live for cached idempotency
// responses. FHIR servers typically cache for 24 hours to allow safe retries
// of write operations that may fail due to transient network issues.
const DefaultIdempotencyTTL = 24 * time.Hour

// IdempotencyKey represents a cached response for an idempotent request.
// When a client retries a write operation with the same idempotency key,
// the server returns the cached response instead of re-executing the request.
type IdempotencyKey struct {
	Key        string
	Method     string
	Path       string
	StatusCode int
	Headers    http.Header
	Body       []byte
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

// IdempotencyStore defines the persistence interface for idempotency key
// storage. Implementations must be safe for concurrent use.
type IdempotencyStore interface {
	// Get retrieves a cached response by idempotency key. The second return
	// value indicates whether the key was found.
	Get(key string) (*IdempotencyKey, bool)
	// Set stores a response for the given idempotency key.
	Set(key string, entry *IdempotencyKey)
	// Delete removes a cached response by idempotency key.
	Delete(key string)
}

// InMemoryIdempotencyStore is a concurrency-safe, in-memory implementation
// of IdempotencyStore with TTL-based expiration and background cleanup.
type InMemoryIdempotencyStore struct {
	mu      sync.RWMutex
	entries map[string]*IdempotencyKey
	ttl     time.Duration
	nowFunc func() time.Time // for testing; defaults to time.Now
	stop    chan struct{}
}

// NewInMemoryIdempotencyStore creates an InMemoryIdempotencyStore with the
// given TTL for cached entries. A background goroutine runs every hour to
// evict expired entries. If ttl is zero or negative, DefaultIdempotencyTTL
// is used.
func NewInMemoryIdempotencyStore(ttl time.Duration) *InMemoryIdempotencyStore {
	if ttl <= 0 {
		ttl = DefaultIdempotencyTTL
	}
	s := &InMemoryIdempotencyStore{
		entries: make(map[string]*IdempotencyKey),
		ttl:     ttl,
		nowFunc: time.Now,
		stop:    make(chan struct{}),
	}
	go s.cleanupLoop()
	return s
}

// cleanupLoop periodically removes expired entries from the store.
func (s *InMemoryIdempotencyStore) cleanupLoop() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.evictExpired()
		case <-s.stop:
			return
		}
	}
}

// Stop terminates the background cleanup goroutine. It should be called
// when the store is no longer needed.
func (s *InMemoryIdempotencyStore) Stop() {
	close(s.stop)
}

// evictExpired removes all entries whose ExpiresAt is in the past.
func (s *InMemoryIdempotencyStore) evictExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.nowFunc()
	for key, entry := range s.entries {
		if now.After(entry.ExpiresAt) {
			delete(s.entries, key)
		}
	}
}

// Get retrieves a cached response by idempotency key. Returns nil, false if
// the key is not found or has expired.
func (s *InMemoryIdempotencyStore) Get(key string) (*IdempotencyKey, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.entries[key]
	if !ok {
		return nil, false
	}
	if s.nowFunc().After(entry.ExpiresAt) {
		return nil, false
	}
	// Return a copy to prevent callers from mutating the stored entry.
	cp := *entry
	if entry.Headers != nil {
		cp.Headers = entry.Headers.Clone()
	}
	cp.Body = make([]byte, len(entry.Body))
	copy(cp.Body, entry.Body)
	return &cp, true
}

// Set stores a response for the given idempotency key. The entry's ExpiresAt
// field is set based on the store's TTL if it is zero.
func (s *InMemoryIdempotencyStore) Set(key string, entry *IdempotencyKey) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *entry
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = s.nowFunc()
	}
	if cp.ExpiresAt.IsZero() {
		cp.ExpiresAt = cp.CreatedAt.Add(s.ttl)
	}
	if entry.Headers != nil {
		cp.Headers = entry.Headers.Clone()
	}
	cp.Body = make([]byte, len(entry.Body))
	copy(cp.Body, entry.Body)
	s.entries[key] = &cp
}

// Delete removes a cached response by idempotency key.
func (s *InMemoryIdempotencyStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
}

// IdempotencyMiddleware returns an Echo middleware that implements idempotency
// key support for FHIR write operations. It reads the Idempotency-Key header
// (standard) or X-Idempotency-Key header (legacy) from incoming POST, PUT, and
// PATCH requests.
//
// Behaviour:
//   - GET and DELETE requests are passed through without inspection.
//   - If no idempotency key header is present, the request is passed through.
//   - If a key is present and a cached response exists:
//   - If the cached method+path do not match, 422 Unprocessable Entity is
//     returned to prevent key reuse across different operations.
//   - Otherwise the cached response (status, headers, body) is replayed and
//     the X-Idempotency-Replayed header is set to "true".
//   - If a key is present but no cache entry exists, the request is executed,
//     the response is captured and cached, and then returned normally.
func IdempotencyMiddleware(store IdempotencyStore) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			method := c.Request().Method
			// Only apply to write operations.
			if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch {
				return next(c)
			}

			// Read idempotency key from standard or legacy header.
			idempKey := c.Request().Header.Get("Idempotency-Key")
			if idempKey == "" {
				idempKey = c.Request().Header.Get("X-Idempotency-Key")
			}
			if idempKey == "" {
				return next(c)
			}

			path := c.Request().URL.Path

			// Check for a cached response.
			if cached, ok := store.Get(idempKey); ok {
				// Verify that the cached entry matches the current method and path.
				if cached.Method != method || cached.Path != path {
					return c.JSON(http.StatusUnprocessableEntity, NewOperationOutcome(
						IssueSeverityError,
						IssueTypeProcessing,
						"Idempotency key was already used for a different operation",
					))
				}
				// Replay the cached response.
				resp := c.Response()
				for k, vals := range cached.Headers {
					for _, v := range vals {
						resp.Header().Set(k, v)
					}
				}
				resp.Header().Set("X-Idempotency-Replayed", "true")
				resp.WriteHeader(cached.StatusCode)
				_, err := resp.Write(cached.Body)
				return err
			}

			// No cached response: execute the handler and capture the response.
			origWriter := c.Response().Writer
			rec := &idempotencyRecorder{
				ResponseWriter: origWriter,
				body:           &bytes.Buffer{},
				statusCode:     http.StatusOK,
				headers:        make(http.Header),
			}
			c.Response().Writer = rec

			if err := next(c); err != nil {
				c.Response().Writer = origWriter
				return err
			}

			// Restore the original writer.
			c.Response().Writer = origWriter

			// Collect headers that were set during handler execution.
			capturedHeaders := make(http.Header)
			for k, vals := range rec.Header() {
				capturedHeaders[k] = vals
			}

			// Cache the response.
			entry := &IdempotencyKey{
				Key:        idempKey,
				Method:     method,
				Path:       path,
				StatusCode: rec.statusCode,
				Headers:    capturedHeaders,
				Body:       rec.body.Bytes(),
			}
			store.Set(idempKey, entry)

			// Write the captured response to the real client.
			for k, vals := range capturedHeaders {
				for _, v := range vals {
					origWriter.Header().Set(k, v)
				}
			}
			origWriter.WriteHeader(rec.statusCode)
			_, err := origWriter.Write(rec.body.Bytes())
			return err
		}
	}
}

// idempotencyRecorder captures an HTTP response for idempotency caching.
// It implements http.ResponseWriter and buffers the status code, headers,
// and body written by the downstream handler.
type idempotencyRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	headers    http.Header
	wroteHead  bool
}

func (r *idempotencyRecorder) Header() http.Header {
	return r.headers
}

func (r *idempotencyRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.wroteHead = true
}

func (r *idempotencyRecorder) Write(b []byte) (int, error) {
	if !r.wroteHead {
		r.statusCode = http.StatusOK
		r.wroteHead = true
	}
	return r.body.Write(b)
}
