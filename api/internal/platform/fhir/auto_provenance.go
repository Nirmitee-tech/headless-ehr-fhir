package fhir

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// ProvenanceRecord — in-memory representation of a created provenance entry
// ---------------------------------------------------------------------------

// ProvenanceRecord holds the data for an auto-created FHIR Provenance resource.
type ProvenanceRecord struct {
	ID              string    `json:"id"`
	TargetReference string    `json:"targetReference"`
	Recorded        time.Time `json:"recorded"`
	AgentWho        string    `json:"agentWho"`
	AgentType       string    `json:"agentType"`
	ActivityCode    string    `json:"activityCode"`
	ActivityDisplay string    `json:"activityDisplay"`
	Reason          string    `json:"reason,omitempty"`
}

// ToFHIR converts a ProvenanceRecord to a FHIR-compliant Provenance resource map.
func (r ProvenanceRecord) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Provenance",
		"id":           r.ID,
		"target": []map[string]interface{}{
			{"reference": r.TargetReference},
		},
		"recorded": r.Recorded.Format(time.RFC3339),
		"agent": []map[string]interface{}{
			{
				"type": map[string]interface{}{
					"coding": []map[string]interface{}{
						{
							"system":  "http://terminology.hl7.org/CodeSystem/provenance-participant-type",
							"code":    r.AgentType,
							"display": r.AgentType,
						},
					},
				},
				"who": map[string]interface{}{
					"reference": r.AgentWho,
				},
			},
		},
		"activity": map[string]interface{}{
			"coding": []map[string]interface{}{
				{
					"system":  "http://terminology.hl7.org/CodeSystem/v3-DataOperation",
					"code":    r.ActivityCode,
					"display": r.ActivityDisplay,
				},
			},
		},
	}

	if r.Reason != "" {
		result["reason"] = []map[string]interface{}{
			{"text": r.Reason},
		}
	}

	return result
}

// ---------------------------------------------------------------------------
// ProvenanceStore — thread-safe in-memory store
// ---------------------------------------------------------------------------

// ProvenanceStore is a thread-safe in-memory store for auto-created provenance
// records.
type ProvenanceStore struct {
	mu      sync.Mutex
	records []ProvenanceRecord
}

// NewProvenanceStore creates a new empty ProvenanceStore.
func NewProvenanceStore() *ProvenanceStore {
	return &ProvenanceStore{
		records: make([]ProvenanceRecord, 0),
	}
}

// Add appends a provenance record to the store in a thread-safe manner.
func (s *ProvenanceStore) Add(r ProvenanceRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, r)
}

// All returns a copy of all stored provenance records.
func (s *ProvenanceStore) All() []ProvenanceRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]ProvenanceRecord, len(s.records))
	copy(result, s.records)
	return result
}

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

// writeMethods is the set of HTTP methods that trigger provenance creation.
var writeMethods = map[string]bool{
	http.MethodPost:   true,
	http.MethodPut:    true,
	http.MethodPatch:  true,
	http.MethodDelete: true,
}

// AutoProvenanceMiddleware returns Echo middleware that auto-creates FHIR
// Provenance resources for every successful write operation on /fhir/*
// endpoints. Provenance creation happens asynchronously after the response is
// sent so it does not block the client.
//
// Opt-out: set the X-No-Provenance: true request header to skip.
func AutoProvenanceMiddleware(store *ProvenanceStore) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			path := req.URL.Path

			// Only process write methods on FHIR endpoints.
			if !writeMethods[req.Method] || !strings.HasPrefix(path, "/fhir/") {
				return next(c)
			}

			// Opt-out header check.
			if strings.EqualFold(req.Header.Get("X-No-Provenance"), "true") {
				return next(c)
			}

			// Capture metadata before the handler runs.
			method := req.Method
			reason := req.Header.Get("X-Provenance-Reason")
			recorded := time.Now().UTC()

			// Extract authenticated user from Echo context. The auth middleware
			// sets "user_id" via c.Set() or context.WithValue with auth.UserIDKey.
			agentWho := ""
			if uid, ok := c.Get("user_id").(string); ok && uid != "" {
				agentWho = uid
			}

			// Intercept the response body so we can read the resource type/id.
			origWriter := c.Response().Writer
			rec := &autoProvenanceRecorder{
				ResponseWriter: origWriter,
				body:           &bytes.Buffer{},
				statusCode:     http.StatusOK,
			}
			c.Response().Writer = rec

			// Call the next handler.
			if err := next(c); err != nil {
				// Restore original writer and return the error.
				c.Response().Writer = origWriter
				return err
			}

			statusCode := rec.statusCode

			// Flush the captured body to the original writer.
			c.Response().Writer = origWriter
			bodyBytes := rec.body.Bytes()
			if len(bodyBytes) > 0 {
				_, _ = origWriter.Write(bodyBytes)
			}

			// Re-read agent after handler (auth middleware may have set it
			// during handler chain).
			if agentWho == "" {
				if uid, ok := c.Get("user_id").(string); ok && uid != "" {
					agentWho = uid
				}
			}
			if agentWho == "" {
				agentWho = "system/anonymous"
			}

			// Only create provenance for successful responses (2xx).
			if statusCode < 200 || statusCode >= 300 {
				return nil
			}

			// Determine target reference from response body or URL path.
			targetRef := ""

			// Try to parse resource type and id from response body.
			if len(bodyBytes) > 0 {
				var resource map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &resource); err == nil {
					rt, _ := resource["resourceType"].(string)
					id, _ := resource["id"].(string)
					if rt != "" && id != "" {
						targetRef = rt + "/" + id
					}
				}
			}

			// Fallback for DELETE with no body: extract from URL path.
			if targetRef == "" {
				targetRef = extractTargetFromPath(path)
			}

			// If we still don't have a target, skip provenance creation.
			if targetRef == "" {
				return nil
			}

			// Map HTTP method to activity code/display.
			activityCode, activityDisplay := methodToActivity(method)

			// Build the provenance record.
			record := ProvenanceRecord{
				ID:              uuid.New().String(),
				TargetReference: targetRef,
				Recorded:        recorded,
				AgentWho:        agentWho,
				AgentType:       "author",
				ActivityCode:    activityCode,
				ActivityDisplay: activityDisplay,
				Reason:          reason,
			}

			// Create provenance asynchronously so we don't block the response.
			go store.Add(record)

			return nil
		}
	}
}

// autoProvenanceRecorder captures the response body for post-processing while
// also tracking the status code.
type autoProvenanceRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (r *autoProvenanceRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *autoProvenanceRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// methodToActivity maps an HTTP method to a FHIR provenance activity code and
// display value.
func methodToActivity(method string) (code, display string) {
	switch method {
	case http.MethodPost:
		return "create", "Create"
	case http.MethodPut:
		return "update", "Update"
	case http.MethodPatch:
		return "update", "Update"
	case http.MethodDelete:
		return "delete", "Delete"
	default:
		return "unknown", "Unknown"
	}
}

// extractTargetFromPath extracts a FHIR resource reference from a URL path
// like /fhir/Patient/pat-123 -> "Patient/pat-123".
func extractTargetFromPath(path string) string {
	// Strip the /fhir/ prefix.
	trimmed := strings.TrimPrefix(path, "/fhir/")
	if trimmed == "" {
		return ""
	}

	segments := strings.SplitN(trimmed, "/", 3)
	if len(segments) < 2 {
		return ""
	}

	resourceType := segments[0]
	resourceID := segments[1]

	// Ignore sub-paths like _history, _search, etc.
	if strings.HasPrefix(resourceID, "_") || resourceID == "" {
		return ""
	}

	return resourceType + "/" + resourceID
}
