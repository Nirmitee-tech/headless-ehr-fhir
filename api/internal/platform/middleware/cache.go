package middleware

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// CacheConfig
// ---------------------------------------------------------------------------

// CacheConfig holds HTTP cache and ETag configuration.
type CacheConfig struct {
	MaxAge             int        // Cache max-age in seconds (default 300 = 5 min)
	Private            bool       // Set Cache-Control: private (default true for PHI)
	NoStore            bool       // Set Cache-Control: no-store for sensitive endpoints
	VaryHeaders        []string   // Headers to include in Vary (default: ["Accept", "Authorization"])
	ETagEnabled        bool       // Enable ETag generation (default true)
	ConditionalEnabled bool       // Support If-None-Match / If-Modified-Since (default true)
	ExcludePaths       []string   // Paths to skip caching (e.g., "/fhir/$export")
	CacheStore         CacheStore // Optional response cache store
}

// DefaultCacheConfig returns a CacheConfig with sensible defaults for a FHIR/EHR API.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxAge:             300,
		Private:            true,
		NoStore:            false,
		VaryHeaders:        []string{"Accept", "Authorization"},
		ETagEnabled:        true,
		ConditionalEnabled: true,
	}
}

// ---------------------------------------------------------------------------
// CacheStore interface
// ---------------------------------------------------------------------------

// CacheStore defines the interface for a response cache backend.
type CacheStore interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte, ttl time.Duration)
	Delete(key string)
	Clear()
}

// ---------------------------------------------------------------------------
// InMemoryCacheStore
// ---------------------------------------------------------------------------

// cacheEntry holds a cached value and its expiration time.
type cacheEntry struct {
	data      []byte
	expiresAt time.Time
}

// InMemoryCacheStore is a thread-safe in-memory CacheStore with lazy expiration.
type InMemoryCacheStore struct {
	entries map[string]*cacheEntry
	mu      sync.RWMutex
}

// NewInMemoryCacheStore creates a new InMemoryCacheStore.
func NewInMemoryCacheStore() *InMemoryCacheStore {
	return &InMemoryCacheStore{
		entries: make(map[string]*cacheEntry),
	}
}

// Get retrieves a value from the cache. Performs lazy expiration: deletes the
// entry and returns a miss if it has expired.
func (s *InMemoryCacheStore) Get(key string) ([]byte, bool) {
	s.mu.RLock()
	entry, ok := s.entries[key]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		s.mu.Lock()
		delete(s.entries, key)
		s.mu.Unlock()
		return nil, false
	}
	return entry.data, true
}

// Set stores a value in the cache with the given TTL.
func (s *InMemoryCacheStore) Set(key string, value []byte, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = &cacheEntry{
		data:      value,
		expiresAt: time.Now().Add(ttl),
	}
}

// Delete removes a single entry from the cache.
func (s *InMemoryCacheStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
}

// Clear removes all entries from the cache.
func (s *InMemoryCacheStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = make(map[string]*cacheEntry)
}

// StartCleanup runs a background goroutine that periodically removes expired
// entries. It stops when the context is cancelled.
func (s *InMemoryCacheStore) StartCleanup(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.mu.Lock()
				now := time.Now()
				for k, v := range s.entries {
					if now.After(v.expiresAt) {
						delete(s.entries, k)
					}
				}
				s.mu.Unlock()
			}
		}
	}()
}

// ---------------------------------------------------------------------------
// Buffered response writer
// ---------------------------------------------------------------------------

// bufferedResponseWriter captures the response body in a buffer so we can
// inspect it (for ETag computation) before flushing to the real writer.
type bufferedResponseWriter struct {
	writer     http.ResponseWriter
	buf        *bytes.Buffer
	statusCode int
}

func newBufferedResponseWriter(w http.ResponseWriter) *bufferedResponseWriter {
	return &bufferedResponseWriter{
		writer:     w,
		buf:        &bytes.Buffer{},
		statusCode: http.StatusOK,
	}
}

// Header returns the underlying writer's header map so that headers set by
// handlers are visible to both the middleware and the final flush.
func (w *bufferedResponseWriter) Header() http.Header {
	return w.writer.Header()
}

// Write captures bytes into the buffer instead of sending them immediately.
func (w *bufferedResponseWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

// WriteHeader captures the status code without writing it to the underlying writer.
func (w *bufferedResponseWriter) WriteHeader(code int) {
	w.statusCode = code
}

// Flush implements http.Flusher (no-op for buffer).
func (w *bufferedResponseWriter) Flush() {}

// flushTo writes the buffered status and body to the underlying writer.
func (w *bufferedResponseWriter) flushTo() error {
	w.writer.WriteHeader(w.statusCode)
	if w.buf.Len() > 0 {
		_, err := w.writer.Write(w.buf.Bytes())
		return err
	}
	return nil
}

// ---------------------------------------------------------------------------
// ETagMiddleware
// ---------------------------------------------------------------------------

// ETagMiddleware returns Echo middleware that computes and sets ETag,
// Cache-Control, and Vary headers on GET/HEAD responses. When ConditionalEnabled
// is true, it handles If-None-Match for 304 Not Modified responses.
func ETagMiddleware(config CacheConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			// Skip non-GET/HEAD methods.
			if req.Method != http.MethodGet && req.Method != http.MethodHead {
				return next(c)
			}

			// Skip excluded paths.
			if shouldSkip(req.URL.Path, config.ExcludePaths) {
				return next(c)
			}

			// Replace the response writer with a buffered version.
			res := c.Response()
			origWriter := res.Writer
			buf := newBufferedResponseWriter(origWriter)
			res.Writer = buf

			// Execute the next handler, writing into the buffer.
			if err := next(c); err != nil {
				res.Writer = origWriter
				return err
			}

			// Restore original writer.
			res.Writer = origWriter

			// Skip ETag/cache headers for error responses.
			if buf.statusCode >= 400 {
				return buf.flushTo()
			}

			// Build and set Cache-Control.
			cc := buildCacheControl(config)
			res.Header().Set("Cache-Control", cc)

			// Set Vary header.
			if len(config.VaryHeaders) > 0 {
				res.Header().Set("Vary", strings.Join(config.VaryHeaders, ", "))
			}

			// Compute ETag from body.
			if config.ETagEnabled {
				body := buf.buf.Bytes()
				etag := computeETag(body)
				res.Header().Set("ETag", etag)

				// Conditional: If-None-Match.
				if config.ConditionalEnabled {
					ifNoneMatch := req.Header.Get("If-None-Match")
					if ifNoneMatch != "" && etagMatch(ifNoneMatch, etag) {
						// Return 304 Not Modified with no body.
						origWriter.WriteHeader(http.StatusNotModified)
						return nil
					}
				}
			}

			// Flush the buffered response to the client.
			return buf.flushTo()
		}
	}
}

// ---------------------------------------------------------------------------
// ConditionalRequestMiddleware
// ---------------------------------------------------------------------------

// ConditionalRequestMiddleware returns Echo middleware that handles conditional
// HTTP requests: If-Modified-Since (304), If-None-Match (304), and
// If-Match (412 Precondition Failed).
func ConditionalRequestMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			// Buffer the response so we can inspect headers set by the handler.
			origWriter := res.Writer
			buf := newBufferedResponseWriter(origWriter)
			res.Writer = buf

			if err := next(c); err != nil {
				res.Writer = origWriter
				return err
			}

			res.Writer = origWriter

			// If-Modified-Since: return 304 if the resource hasn't changed.
			ifModSince := req.Header.Get("If-Modified-Since")
			if ifModSince != "" {
				lastMod := res.Header().Get("Last-Modified")
				if lastMod != "" {
					ims, errIMS := http.ParseTime(ifModSince)
					lm, errLM := http.ParseTime(lastMod)
					if errIMS == nil && errLM == nil && !lm.After(ims) {
						origWriter.WriteHeader(http.StatusNotModified)
						return nil
					}
				}
			}

			// If-None-Match: return 304 if ETag matches.
			ifNoneMatch := req.Header.Get("If-None-Match")
			if ifNoneMatch != "" {
				etag := res.Header().Get("ETag")
				if etag != "" && etagMatch(ifNoneMatch, etag) {
					origWriter.WriteHeader(http.StatusNotModified)
					return nil
				}
			}

			// If-Match: return 412 if ETag does NOT match.
			ifMatch := req.Header.Get("If-Match")
			if ifMatch != "" {
				etag := res.Header().Get("ETag")
				if etag != "" && !etagMatch(ifMatch, etag) {
					origWriter.WriteHeader(http.StatusPreconditionFailed)
					return nil
				}
			}

			// No conditional matched; flush the buffered response.
			return buf.flushTo()
		}
	}
}

// ---------------------------------------------------------------------------
// ResponseCacheMiddleware
// ---------------------------------------------------------------------------

// ResponseCacheMiddleware returns Echo middleware that caches GET responses
// by URL + Accept header. Requests with an Authorization header skip the cache
// to protect private data.
func ResponseCacheMiddleware(store CacheStore, ttl time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			// Only cache GET requests.
			if req.Method != http.MethodGet {
				return next(c)
			}

			// Skip caching for authorized requests (private data).
			if req.Header.Get("Authorization") != "" {
				c.Response().Header().Set("X-Cache", "SKIP")
				return next(c)
			}

			key := cacheKey(req.Method, req.URL.Path, req.Header.Get("Accept"))

			// Check cache.
			if data, ok := store.Get(key); ok {
				c.Response().Header().Set("X-Cache", "HIT")
				c.Response().Writer.WriteHeader(http.StatusOK)
				_, err := c.Response().Writer.Write(data)
				return err
			}

			// Cache miss: buffer the response.
			res := c.Response()
			origWriter := res.Writer
			buf := newBufferedResponseWriter(origWriter)
			res.Writer = buf

			if err := next(c); err != nil {
				res.Writer = origWriter
				return err
			}

			res.Writer = origWriter

			// Only cache successful responses.
			if buf.statusCode < 400 {
				store.Set(key, buf.buf.Bytes(), ttl)
			}

			res.Header().Set("X-Cache", "MISS")
			return buf.flushTo()
		}
	}
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

// computeETag returns a weak ETag based on the MD5 hash of the body.
func computeETag(body []byte) string {
	hash := md5.Sum(body)
	return fmt.Sprintf(`W/"%x"`, hash)
}

// cacheKey builds a cache key from the HTTP method, path, and Accept header.
func cacheKey(method, path, accept string) string {
	return method + ":" + path + ":" + accept
}

// shouldSkip returns true if the path matches any of the excluded paths.
func shouldSkip(path string, excludes []string) bool {
	for _, ex := range excludes {
		if path == ex {
			return true
		}
	}
	return false
}

// buildCacheControl constructs a Cache-Control header value from the config.
func buildCacheControl(config CacheConfig) string {
	var parts []string
	if config.NoStore {
		parts = append(parts, "no-store")
	}
	if config.Private {
		parts = append(parts, "private")
	} else {
		parts = append(parts, "public")
	}
	parts = append(parts, fmt.Sprintf("max-age=%d", config.MaxAge))
	return strings.Join(parts, ", ")
}

// etagMatch checks if the provided If-None-Match (or If-Match) header value
// matches the given ETag. Supports comma-separated lists and the wildcard "*".
func etagMatch(headerVal, etag string) bool {
	headerVal = strings.TrimSpace(headerVal)
	if headerVal == "*" {
		return true
	}
	for _, candidate := range strings.Split(headerVal, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == etag {
			return true
		}
		// Weak comparison: W/"x" matches W/"x" or "x".
		if stripWeakPrefix(candidate) == stripWeakPrefix(etag) {
			return true
		}
	}
	return false
}

// stripWeakPrefix removes the W/ prefix from a weak ETag.
func stripWeakPrefix(etag string) string {
	if strings.HasPrefix(etag, `W/`) {
		return etag[2:]
	}
	return etag
}
