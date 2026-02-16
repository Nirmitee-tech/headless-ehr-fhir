package fhir

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// ConditionalReadMiddleware creates Echo middleware that implements FHIR
// conditional read semantics (HTTP 304 Not Modified). It intercepts GET
// responses for single resource reads and checks:
//
//   - If-None-Match: if the client-supplied ETag matches the response ETag,
//     a 304 Not Modified is returned with no body.
//   - If-Modified-Since: if the resource has not been modified since the
//     client-supplied timestamp, a 304 Not Modified is returned with no body.
//
// The middleware is a no-op for non-GET requests, search bundle responses,
// and responses with non-200 status codes.
func ConditionalReadMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Only apply to GET requests.
			if c.Request().Method != http.MethodGet {
				return next(c)
			}

			// Only apply when the client sends conditional headers.
			ifNoneMatch := c.Request().Header.Get("If-None-Match")
			ifModifiedSince := c.Request().Header.Get("If-Modified-Since")
			if ifNoneMatch == "" && ifModifiedSince == "" {
				return next(c)
			}

			// Capture the response body by wrapping the response writer.
			origWriter := c.Response().Writer
			rec := &conditionalReadRecorder{
				ResponseWriter: origWriter,
				body:           &bytes.Buffer{},
				statusCode:     http.StatusOK,
			}
			c.Response().Writer = rec

			// Execute the handler chain.
			if err := next(c); err != nil {
				c.Response().Writer = origWriter
				return err
			}

			// Only consider 200 OK responses for conditional read.
			// Non-200 responses (errors, 201 Created, etc.) pass through unchanged.
			if rec.statusCode != http.StatusOK {
				return flushConditionalReadRecorder(origWriter, rec)
			}

			// Check if the response looks like a search bundle by inspecting
			// the Content-Type or body. Search bundles return collections and
			// should not be subject to conditional read at the middleware level.
			// We use a simple heuristic: if the response contains "searchset"
			// near the beginning, skip it.
			if isSearchBundle(rec.body.Bytes()) {
				return flushConditionalReadRecorder(origWriter, rec)
			}

			// Check If-None-Match against the response ETag.
			if ifNoneMatch != "" {
				responseETag := rec.Header().Get("ETag")
				if responseETag != "" && etagsMatch(ifNoneMatch, responseETag) {
					return writeNotModified(origWriter, rec)
				}
			}

			// Check If-Modified-Since against the response Last-Modified.
			if ifModifiedSince != "" {
				lastModified := rec.Header().Get("Last-Modified")
				if lastModified != "" && !modifiedSince(lastModified, ifModifiedSince) {
					return writeNotModified(origWriter, rec)
				}
			}

			// No conditional match; flush the original response.
			return flushConditionalReadRecorder(origWriter, rec)
		}
	}
}

// conditionalReadRecorder captures the response body and status code written
// by downstream handlers so the middleware can inspect ETag/Last-Modified
// headers before deciding whether to return 304.
type conditionalReadRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	wroteHead  bool
}

func (r *conditionalReadRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.wroteHead = true
	// Do NOT forward to the underlying writer yet; we buffer everything.
}

func (r *conditionalReadRecorder) Write(b []byte) (int, error) {
	if !r.wroteHead {
		r.statusCode = http.StatusOK
		r.wroteHead = true
	}
	return r.body.Write(b)
}

// flushConditionalReadRecorder writes the buffered response directly to the
// original http.ResponseWriter.
func flushConditionalReadRecorder(w http.ResponseWriter, rec *conditionalReadRecorder) error {
	for k, vals := range rec.Header() {
		for _, v := range vals {
			w.Header().Set(k, v)
		}
	}
	w.WriteHeader(rec.statusCode)
	_, err := w.Write(rec.body.Bytes())
	return err
}

// writeNotModified sends a 304 Not Modified response, preserving relevant
// headers (ETag, Last-Modified) but omitting the body.
func writeNotModified(w http.ResponseWriter, rec *conditionalReadRecorder) error {
	// Copy ETag and Last-Modified headers.
	if etag := rec.Header().Get("ETag"); etag != "" {
		w.Header().Set("ETag", etag)
	}
	if lm := rec.Header().Get("Last-Modified"); lm != "" {
		w.Header().Set("Last-Modified", lm)
	}
	w.WriteHeader(http.StatusNotModified)
	return nil
}

// etagsMatch compares a client If-None-Match value with a response ETag.
// Both weak and strong comparison are treated as matching per RFC 7232
// section 3.2 (weak comparison for GET conditional requests).
func etagsMatch(ifNoneMatch, responseETag string) bool {
	// Handle wildcard.
	if strings.TrimSpace(ifNoneMatch) == "*" {
		return true
	}

	// Normalize both values: strip W/ prefix and quotes for comparison.
	clientVersion, clientErr := ParseETag(ifNoneMatch)
	serverVersion, serverErr := ParseETag(responseETag)
	if clientErr != nil || serverErr != nil {
		return false
	}
	return clientVersion == serverVersion
}

// modifiedSince returns true if the resource's Last-Modified time is after
// the client's If-Modified-Since time. Returns true (modified) on parse
// errors to ensure the full response is returned when timestamps are invalid.
func modifiedSince(lastModified, ifModifiedSince string) bool {
	// Try common HTTP and FHIR timestamp formats.
	formats := []string{
		http.TimeFormat,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
	}

	var lmTime time.Time
	var parsed bool
	for _, f := range formats {
		t, err := time.Parse(f, lastModified)
		if err == nil {
			lmTime = t
			parsed = true
			break
		}
	}
	if !parsed {
		return true // Can't parse; assume modified.
	}

	var imsTime time.Time
	parsed = false
	for _, f := range formats {
		t, err := time.Parse(f, ifModifiedSince)
		if err == nil {
			imsTime = t
			parsed = true
			break
		}
	}
	if !parsed {
		return true // Can't parse; assume modified.
	}

	return lmTime.After(imsTime)
}

// isSearchBundle performs a fast check to determine if a response body is a
// FHIR search bundle. It looks for "searchset" in the first 512 bytes of the
// response to avoid parsing the entire JSON body.
func isSearchBundle(body []byte) bool {
	limit := 512
	if len(body) < limit {
		limit = len(body)
	}
	return bytes.Contains(body[:limit], []byte(`"searchset"`))
}
