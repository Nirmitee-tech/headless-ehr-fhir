package fhir

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// HeadResponse captures headers and status from a GET response for HEAD.
type HeadResponse struct {
	StatusCode    int
	Headers       map[string][]string
	ContentLength int64
	ContentType   string
	ETag          string
	LastModified  string
}

// HeadMethodConfig configures HEAD method behavior.
type HeadMethodConfig struct {
	EnableContentLength    bool          // Calculate Content-Length from response body
	CacheHeaders           bool          // Cache HEAD responses for subsequent requests
	CacheTTL               time.Duration // TTL for cached HEAD responses
	AllowedPaths           []string      // Paths where HEAD is allowed (empty = all)
	IncludeResourceHeaders bool          // Include X-FHIR-* resource metadata headers
}

// DefaultHeadMethodConfig returns sensible defaults for HEAD method handling.
// Content-Length calculation and FHIR resource headers are enabled by default.
// Caching is disabled by default and must be opted into.
func DefaultHeadMethodConfig() HeadMethodConfig {
	return HeadMethodConfig{
		EnableContentLength:    true,
		CacheHeaders:           false,
		CacheTTL:               0,
		AllowedPaths:           nil,
		IncludeResourceHeaders: true,
	}
}

// IsHeadRequest checks if the current request is a HEAD request.
func IsHeadRequest(c echo.Context) bool {
	return c.Request().Method == http.MethodHead
}

// HeadResponseWriter wraps the response writer to capture body for HEAD
// processing. It buffers all written data and status codes so the middleware
// can inspect the full response before flushing headers without the body.
type HeadResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
	headers    http.Header
}

// NewHeadResponseWriter creates a HeadResponseWriter that captures writes
// destined for the underlying http.ResponseWriter.
func NewHeadResponseWriter(w http.ResponseWriter) *HeadResponseWriter {
	return &HeadResponseWriter{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
		headers:        w.Header().Clone(),
	}
}

// Write captures written bytes into the internal buffer instead of sending
// them to the underlying writer.
func (w *HeadResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

// WriteHeader captures the HTTP status code without forwarding it.
func (w *HeadResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

// Header returns the header map that will be sent by WriteHeader.
func (w *HeadResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// HeadMethodMiddleware returns middleware that handles HEAD requests by
// converting them to GET internally, executing the handler, and stripping
// the response body. It preserves all response headers including
// Content-Type, ETag, and Last-Modified.
//
// When config is nil, DefaultHeadMethodConfig is used.
func HeadMethodMiddleware(config *HeadMethodConfig) echo.MiddlewareFunc {
	cfg := DefaultHeadMethodConfig()
	if config != nil {
		cfg = *config
	}

	// Initialise cache if caching is enabled.
	var cache *HeadCache
	if cfg.CacheHeaders && cfg.CacheTTL > 0 {
		cache = NewHeadCache(cfg.CacheTTL)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Only intercept HEAD requests; all other methods pass through.
			if c.Request().Method != http.MethodHead {
				return next(c)
			}

			// Check allowed paths if configured.
			if len(cfg.AllowedPaths) > 0 {
				path := c.Request().URL.Path
				allowed := false
				for _, p := range cfg.AllowedPaths {
					if strings.HasPrefix(path, p) {
						allowed = true
						break
					}
				}
				if !allowed {
					return echo.NewHTTPError(http.StatusMethodNotAllowed, "HEAD method not allowed for this path")
				}
			}

			// Check cache first.
			if cache != nil {
				cacheKey := GenerateCacheKey(
					c.Request().Method,
					c.Request().URL.RequestURI(),
					c.Request().Header,
				)
				if cached, ok := cache.Get(cacheKey); ok {
					return writeCachedHeadResponse(c, cached)
				}
			}

			// Convert HEAD to GET so the handler executes normally.
			c.Request().Method = http.MethodGet

			// Capture the response by wrapping the writer.
			origWriter := c.Response().Writer
			rec := NewHeadResponseWriter(origWriter)
			c.Response().Writer = rec

			// Execute the handler chain.
			if err := next(c); err != nil {
				c.Response().Writer = origWriter
				c.Request().Method = http.MethodHead
				return err
			}

			// Restore the original method.
			c.Request().Method = http.MethodHead

			// Build response headers on the original writer.
			for k, vals := range rec.Header() {
				for _, v := range vals {
					origWriter.Header().Set(k, v)
				}
			}

			// Determine status code; default to 200 if handler didn't set one.
			statusCode := rec.statusCode
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			// Set Content-Length from captured body if enabled.
			if cfg.EnableContentLength {
				origWriter.Header().Set("Content-Length", fmt.Sprintf("%d", rec.body.Len()))
			} else {
				origWriter.Header().Del("Content-Length")
			}

			// Extract and set FHIR resource metadata headers if enabled.
			if cfg.IncludeResourceHeaders && rec.body.Len() > 0 {
				meta := ExtractResourceMetadata(rec.body.Bytes())
				if rt, ok := meta["resourceType"]; ok {
					origWriter.Header().Set("X-FHIR-ResourceType", rt)
				}
				if id, ok := meta["id"]; ok {
					origWriter.Header().Set("X-FHIR-ResourceId", id)
				}
			}

			// Cache the response if caching is enabled.
			if cache != nil {
				cacheKey := GenerateCacheKey(
					c.Request().Method,
					c.Request().URL.RequestURI(),
					c.Request().Header,
				)
				cached := &HeadResponse{
					StatusCode:    statusCode,
					Headers:       cloneHeaders(origWriter.Header()),
					ContentLength: int64(rec.body.Len()),
					ContentType:   origWriter.Header().Get("Content-Type"),
					ETag:          origWriter.Header().Get("ETag"),
					LastModified:  origWriter.Header().Get("Last-Modified"),
				}
				cache.Set(cacheKey, cached)
			}

			// Write status and empty body (HEAD must not include a body).
			origWriter.WriteHeader(statusCode)
			c.Response().Writer = origWriter
			return nil
		}
	}
}

// writeCachedHeadResponse writes a cached HEAD response back to the client.
func writeCachedHeadResponse(c echo.Context, resp *HeadResponse) error {
	w := c.Response().Writer
	for k, vals := range resp.Headers {
		for _, v := range vals {
			w.Header().Set(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	return nil
}

// cloneHeaders creates a deep copy of an http.Header map.
func cloneHeaders(h http.Header) map[string][]string {
	clone := make(map[string][]string, len(h))
	for k, vals := range h {
		vc := make([]string, len(vals))
		copy(vc, vals)
		clone[k] = vc
	}
	return clone
}

// ExtractResourceMetadata extracts FHIR resource metadata headers from a
// JSON response body. It looks for resourceType, id, and meta.versionId /
// meta.lastUpdated fields. Returns an empty map if the body is not valid
// JSON or does not contain the expected fields.
func ExtractResourceMetadata(body []byte) map[string]string {
	result := make(map[string]string)
	if len(body) == 0 {
		return result
	}

	var raw struct {
		ResourceType string `json:"resourceType"`
		ID           string `json:"id"`
		Meta         *struct {
			VersionID   string `json:"versionId"`
			LastUpdated string `json:"lastUpdated"`
		} `json:"meta"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return result
	}

	if raw.ResourceType != "" {
		result["resourceType"] = raw.ResourceType
	}
	if raw.ID != "" {
		result["id"] = raw.ID
	}
	if raw.Meta != nil {
		if raw.Meta.VersionID != "" {
			result["versionId"] = raw.Meta.VersionID
		}
		if raw.Meta.LastUpdated != "" {
			result["lastUpdated"] = raw.Meta.LastUpdated
		}
	}

	return result
}

// BuildHeadHeaders constructs the set of headers for a HEAD response from
// a HeadResponse struct. It includes Content-Length, Content-Type, ETag,
// Last-Modified, and any additional headers from the original response.
func BuildHeadHeaders(resp *HeadResponse) http.Header {
	h := http.Header{}

	// Copy any extra headers first.
	for k, vals := range resp.Headers {
		for _, v := range vals {
			h.Add(k, v)
		}
	}

	// Set standard headers (overwriting any from Headers map).
	h.Set("Content-Length", fmt.Sprintf("%d", resp.ContentLength))
	if resp.ContentType != "" {
		h.Set("Content-Type", resp.ContentType)
	}
	if resp.ETag != "" {
		h.Set("ETag", resp.ETag)
	}
	if resp.LastModified != "" {
		h.Set("Last-Modified", resp.LastModified)
	}

	return h
}

// ValidateHeadResponse validates that a HEAD response matches FHIR
// requirements. It checks for the presence of required headers and valid
// status codes.
func ValidateHeadResponse(statusCode int, headers http.Header) []ValidationIssue {
	var issues []ValidationIssue

	// Status code must be a valid HTTP status.
	if statusCode < 100 || statusCode > 599 {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeValue,
			Diagnostics: fmt.Sprintf("invalid HTTP status code: %d", statusCode),
		})
	}

	// Content-Type should be present for successful responses.
	if headers.Get("Content-Type") == "" {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityWarning,
			Code:        VIssueTypeRequired,
			Diagnostics: "Content-Type header is missing from HEAD response",
		})
	}

	return issues
}

// HeadCacheEntry stores cached HEAD response metadata with expiration info.
type HeadCacheEntry struct {
	Response  *HeadResponse
	CachedAt  time.Time
	ExpiresAt time.Time
}

// HeadCache provides simple in-memory caching for HEAD responses. It is
// safe for concurrent access.
type HeadCache struct {
	mu      sync.RWMutex
	entries map[string]*HeadCacheEntry
	ttl     time.Duration
}

// NewHeadCache creates a new HeadCache with the given TTL for entries.
func NewHeadCache(ttl time.Duration) *HeadCache {
	return &HeadCache{
		entries: make(map[string]*HeadCacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves a cached HEAD response. Returns the response and true if
// a non-expired entry exists, or nil and false otherwise.
func (c *HeadCache) Get(key string) (*HeadResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	return entry.Response, true
}

// Set stores a HEAD response in the cache with the configured TTL.
func (c *HeadCache) Set(key string, resp *HeadResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.entries[key] = &HeadCacheEntry{
		Response:  resp,
		CachedAt:  now,
		ExpiresAt: now.Add(c.ttl),
	}
}

// Invalidate removes a specific entry from the cache.
func (c *HeadCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// Clear removes all entries from the cache.
func (c *HeadCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*HeadCacheEntry)
}

// GenerateCacheKey generates a deterministic cache key for a HEAD request
// based on the HTTP method, request path, and relevant request headers
// (Accept, If-None-Match, If-Modified-Since).
func GenerateCacheKey(method, path string, headers http.Header) string {
	h := sha256.New()
	h.Write([]byte(method))
	h.Write([]byte(":"))
	h.Write([]byte(path))

	// Include relevant headers in sorted order for determinism.
	relevantHeaders := []string{"Accept", "If-None-Match", "If-Modified-Since"}
	sort.Strings(relevantHeaders)
	for _, name := range relevantHeaders {
		val := headers.Get(name)
		if val != "" {
			h.Write([]byte("|"))
			h.Write([]byte(name))
			h.Write([]byte("="))
			h.Write([]byte(val))
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}
