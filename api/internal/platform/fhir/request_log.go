package fhir

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// FHIRRequestLog holds structured information about a single FHIR HTTP
// request/response cycle. It captures FHIR-specific metadata such as resource
// type, interaction kind, and relevant headers alongside standard HTTP fields.
type FHIRRequestLog struct {
	Timestamp    time.Time `json:"timestamp"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	ResourceType string    `json:"resourceType,omitempty"`
	ResourceID   string    `json:"resourceId,omitempty"`
	Operation    string    `json:"operation,omitempty"`
	Interaction  string    `json:"interaction"`
	StatusCode   int       `json:"statusCode"`
	Duration     int64     `json:"durationMs"`
	ResponseSize int       `json:"responseSize"`
	ClientIP     string    `json:"clientIp"`
	UserAgent    string    `json:"userAgent,omitempty"`
	TenantID     string    `json:"tenantId,omitempty"`
	RequestID    string    `json:"requestId,omitempty"`
	PreferHeader string    `json:"prefer,omitempty"`
	IfMatch      string    `json:"ifMatch,omitempty"`
	IfNoneMatch  string    `json:"ifNoneMatch,omitempty"`
}

// FHIRLogSink is an interface for consuming structured FHIR request log
// entries. Implementations can write to files, send to log aggregation
// services, or buffer entries for asynchronous processing.
type FHIRLogSink interface {
	Log(entry FHIRRequestLog)
}

// ChannelLogSink buffers FHIRRequestLog entries in a channel for asynchronous
// processing. Consumers read entries via the Entries channel.
type ChannelLogSink struct {
	ch chan FHIRRequestLog
}

// NewChannelLogSink creates a ChannelLogSink with the given buffer size.
func NewChannelLogSink(bufSize int) *ChannelLogSink {
	return &ChannelLogSink{
		ch: make(chan FHIRRequestLog, bufSize),
	}
}

// Log sends a log entry to the channel. If the channel buffer is full the
// entry is silently dropped to avoid blocking request processing.
func (s *ChannelLogSink) Log(entry FHIRRequestLog) {
	select {
	case s.ch <- entry:
	default:
		// Drop the entry rather than blocking the request path.
	}
}

// Entries returns the read-only channel of buffered log entries.
func (s *ChannelLogSink) Entries() <-chan FHIRRequestLog {
	return s.ch
}

// FHIRRequestLoggerMiddleware returns Echo middleware that records a
// FHIRRequestLog entry for every request. The entry is sent to the provided
// FHIRLogSink after the downstream handler completes.
func FHIRRequestLoggerMiddleware(sink FHIRLogSink) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			path := req.URL.Path

			// Capture the response body size using a recording writer.
			origWriter := c.Response().Writer
			rec := &logResponseRecorder{
				ResponseWriter: origWriter,
				body:           &bytes.Buffer{},
				statusCode:     http.StatusOK,
			}
			c.Response().Writer = rec

			err := next(c)

			duration := time.Since(start)

			resourceType, resourceID, operation := ExtractResourceInfo(path)
			interaction := ClassifyInteraction(req.Method, path)

			entry := FHIRRequestLog{
				Timestamp:    start,
				Method:       req.Method,
				Path:         path,
				ResourceType: resourceType,
				ResourceID:   resourceID,
				Operation:    operation,
				Interaction:  interaction,
				StatusCode:   rec.statusCode,
				Duration:     duration.Milliseconds(),
				ResponseSize: rec.body.Len(),
				ClientIP:     c.RealIP(),
				UserAgent:    req.UserAgent(),
				TenantID:     req.Header.Get("X-Tenant-ID"),
				RequestID:    req.Header.Get("X-Request-ID"),
				PreferHeader: req.Header.Get("Prefer"),
				IfMatch:      req.Header.Get("If-Match"),
				IfNoneMatch:  req.Header.Get("If-None-Match"),
			}

			sink.Log(entry)

			return err
		}
	}
}

// ClassifyInteraction determines the FHIR interaction type from the HTTP
// method and request path. It returns one of the standard FHIR interaction
// names: read, vread, search-type, create, update, delete, history-instance,
// history-type, or operation.
func ClassifyInteraction(method, path string) string {
	segments := trimmedPathSegments(path)

	// Check for FHIR operations ($validate, $export, etc.).
	for _, seg := range segments {
		if strings.HasPrefix(seg, "$") {
			return "operation"
		}
	}

	// Determine the segment layout relative to the resource type.
	// Typical layouts after trimming the base (e.g. "/fhir"):
	//   [ResourceType]
	//   [ResourceType, id]
	//   [ResourceType, id, _history]
	//   [ResourceType, id, _history, vid]
	//   [ResourceType, _history]
	resSegments := resourceSegments(segments)

	switch method {
	case http.MethodGet:
		return classifyGet(resSegments)
	case http.MethodPost:
		return "create"
	case http.MethodPut:
		return "update"
	case http.MethodDelete:
		return "delete"
	case http.MethodPatch:
		return "update"
	default:
		return "unknown"
	}
}

// classifyGet determines the specific GET interaction from the resource path
// segments following the resource type.
func classifyGet(segments []string) string {
	n := len(segments)
	switch {
	case n == 0:
		// GET / (system-level)
		return "search-type"
	case n == 1:
		// GET /ResourceType
		return "search-type"
	case n == 2:
		// GET /ResourceType/{id}
		if segments[1] == "_history" {
			// GET /ResourceType/_history
			return "history-type"
		}
		return "read"
	case n == 3:
		// GET /ResourceType/{id}/_history
		if segments[2] == "_history" {
			return "history-instance"
		}
		return "read"
	case n >= 4:
		// GET /ResourceType/{id}/_history/{vid}
		if segments[2] == "_history" {
			return "vread"
		}
		return "read"
	}
	return "read"
}

// ExtractResourceInfo parses a FHIR URL path and returns the resource type,
// resource ID, and operation name (if any). The path may include a base prefix
// such as "/fhir".
func ExtractResourceInfo(path string) (resourceType, resourceID, operation string) {
	segments := trimmedPathSegments(path)
	resSegs := resourceSegments(segments)

	for i, seg := range resSegs {
		if strings.HasPrefix(seg, "$") {
			operation = seg
			// Resource type is the segment before the operation, unless the
			// operation is preceded by an ID.
			switch {
			case i >= 2:
				resourceType = resSegs[i-2]
				resourceID = resSegs[i-1]
			case i >= 1:
				resourceType = resSegs[i-1]
			}
			return
		}
	}

	if len(resSegs) >= 1 {
		resourceType = resSegs[0]
	}
	if len(resSegs) >= 2 && resSegs[1] != "_history" {
		resourceID = resSegs[1]
	}
	return
}

// trimmedPathSegments splits the URL path into non-empty segments.
func trimmedPathSegments(path string) []string {
	raw := strings.Split(path, "/")
	segments := make([]string, 0, len(raw))
	for _, s := range raw {
		if s != "" {
			segments = append(segments, s)
		}
	}
	return segments
}

// resourceSegments strips known base-path prefixes (e.g. "fhir") and returns
// the remaining path segments that describe the resource interaction.
func resourceSegments(segments []string) []string {
	if len(segments) == 0 {
		return segments
	}
	// Skip common base-path segments that are not FHIR resource types.
	start := 0
	if len(segments) > 0 && isBasePath(segments[0]) {
		start = 1
	}
	return segments[start:]
}

// isBasePath returns true for well-known FHIR server base path segments.
func isBasePath(s string) bool {
	switch strings.ToLower(s) {
	case "fhir", "r4", "api", "ehr":
		return true
	}
	return false
}

// logResponseRecorder captures the status code and response body size written
// by downstream handlers, while still forwarding all writes to the original
// ResponseWriter.
type logResponseRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	wroteHead  bool
}

func (r *logResponseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.wroteHead = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *logResponseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHead {
		r.statusCode = http.StatusOK
		r.wroteHead = true
	}
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
