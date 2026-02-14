package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/ehr/ehr/internal/platform/auth"
)

// AuditEntry represents an audit log entry produced by the middleware.
// It captures who accessed what, when, from where, and the action type.
type AuditEntry struct {
	UserID       string
	UserRoles    []string
	ResourceType string
	PatientID    string
	Action       string // read, create, update, delete, search
	IPAddress    string
	UserAgent    string
	Path         string
	Method       string
	IsBreakGlass bool
	BreakGlassReason string
	Timestamp    time.Time
	RequestID    string
	StatusCode   int
}

// AuditRecorder is the interface that the audit middleware uses to persist
// audit entries. This decouples the middleware from the concrete hipaa.AuditLogger
// so that tests can provide a mock implementation.
type AuditRecorder interface {
	RecordAccess(entry AuditEntry) error
}

// AuditRecorderFunc is a function adapter for AuditRecorder.
type AuditRecorderFunc func(entry AuditEntry) error

func (f AuditRecorderFunc) RecordAccess(entry AuditEntry) error {
	return f(entry)
}

// Audit returns Echo middleware that intercepts requests to /fhir/* and /api/v1/*,
// extracts the authenticated user from JWT claims, determines the FHIR resource type
// from the URL path, and logs PHI access for HIPAA compliance.
//
// If no AuditRecorder is provided, it falls back to structured zerolog logging.
// Break-glass detection: if the X-Break-Glass header is present, the access is
// logged as an emergency override.
func Audit(logger zerolog.Logger, recorders ...AuditRecorder) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			path := req.URL.Path

			// Only audit FHIR and API routes
			if !isAuditablePath(path) {
				return next(c)
			}

			// Execute the handler first so we capture the response status
			err := next(c)

			// Build audit entry
			entry := AuditEntry{
				Timestamp: time.Now().UTC(),
				Path:      path,
				Method:    req.Method,
				IPAddress: c.RealIP(),
				UserAgent: req.UserAgent(),
				StatusCode: c.Response().Status,
			}

			// Extract authenticated user from JWT claims via context
			ctx := req.Context()
			entry.UserID = auth.UserIDFromContext(ctx)
			entry.UserRoles = auth.RolesFromContext(ctx)

			// Request ID from middleware chain
			if rid, ok := c.Get("request_id").(string); ok {
				entry.RequestID = rid
			}

			// Determine action type from HTTP method
			entry.Action = httpMethodToAction(req.Method)

			// Extract FHIR resource type from path
			entry.ResourceType = extractResourceType(path)

			// Extract patient ID from path or query params
			entry.PatientID = extractPatientID(c)

			// Break-glass detection
			if bgReason := req.Header.Get("X-Break-Glass"); bgReason != "" {
				entry.IsBreakGlass = true
				entry.BreakGlassReason = bgReason
			}

			// Record the audit entry
			if len(recorders) > 0 && recorders[0] != nil {
				if recErr := recorders[0].RecordAccess(entry); recErr != nil {
					logger.Error().Err(recErr).
						Str("request_id", entry.RequestID).
						Msg("failed to record audit entry")
				}
			}

			// Always emit a structured log for audit trail
			evt := logger.Info()
			if entry.IsBreakGlass {
				evt = logger.Warn()
			}
			evt.
				Str("type", "hipaa_audit").
				Str("request_id", entry.RequestID).
				Str("user_id", entry.UserID).
				Strs("user_roles", entry.UserRoles).
				Str("resource_type", entry.ResourceType).
				Str("patient_id", entry.PatientID).
				Str("action", entry.Action).
				Str("method", entry.Method).
				Str("path", entry.Path).
				Str("remote_ip", entry.IPAddress).
				Int("status", entry.StatusCode).
				Bool("break_glass", entry.IsBreakGlass).
				Str("break_glass_reason", entry.BreakGlassReason).
				Msg("phi_access")

			return err
		}
	}
}

// isAuditablePath returns true if the path is under /fhir/ or /api/v1/.
func isAuditablePath(path string) bool {
	return strings.HasPrefix(path, "/fhir/") || strings.HasPrefix(path, "/api/v1/")
}

// httpMethodToAction maps HTTP methods to FHIR audit action codes.
func httpMethodToAction(method string) string {
	switch method {
	case http.MethodGet, http.MethodHead:
		return "read"
	case http.MethodPost:
		return "create"
	case http.MethodPut, http.MethodPatch:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return "read"
	}
}

// extractResourceType parses the FHIR resource type from a URL path.
//
// Supported patterns:
//   - /fhir/Patient          -> Patient
//   - /fhir/Patient/123      -> Patient
//   - /api/v1/patients       -> patients
//   - /api/v1/patients/123   -> patients
func extractResourceType(path string) string {
	var segments []string
	if strings.HasPrefix(path, "/fhir/") {
		segments = strings.Split(strings.TrimPrefix(path, "/fhir/"), "/")
	} else if strings.HasPrefix(path, "/api/v1/") {
		segments = strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
	}
	if len(segments) > 0 && segments[0] != "" {
		return segments[0]
	}
	return "unknown"
}

// extractPatientID attempts to find a patient identifier in the request.
// It checks the URL path for /Patient/<id> patterns and query params for patient=<id>.
func extractPatientID(c echo.Context) string {
	path := c.Request().URL.Path

	// Check FHIR path: /fhir/Patient/<uuid>
	if strings.HasPrefix(path, "/fhir/Patient/") {
		segments := strings.Split(strings.TrimPrefix(path, "/fhir/Patient/"), "/")
		if len(segments) > 0 && isUUIDLike(segments[0]) {
			return segments[0]
		}
	}

	// Check API path: /api/v1/patients/<uuid>
	if strings.HasPrefix(path, "/api/v1/patients/") {
		segments := strings.Split(strings.TrimPrefix(path, "/api/v1/patients/"), "/")
		if len(segments) > 0 && isUUIDLike(segments[0]) {
			return segments[0]
		}
	}

	// Check query parameter: ?patient=<id> or ?patient=Patient/<id>
	if patient := c.QueryParam("patient"); patient != "" {
		// Strip "Patient/" prefix if present (FHIR reference format)
		patient = strings.TrimPrefix(patient, "Patient/")
		return patient
	}

	return ""
}

// isUUIDLike checks if a string looks like a UUID (basic length/format check).
func isUUIDLike(s string) bool {
	if len(s) < 1 {
		return false
	}
	_, err := uuid.Parse(s)
	return err == nil
}
