package fhir

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// Confidentiality classification labels from the HL7 v3 ConfidentialityClassification
// code system. These form a hierarchy: U < L < M < N < R < V.
const (
	LabelUnrestricted   = "U" // unrestricted
	LabelLow            = "L" // low
	LabelModerate       = "M" // moderate
	LabelNormal         = "N" // normal
	LabelRestricted     = "R" // restricted
	LabelVeryRestricted = "V" // very restricted
)

// Data sensitivity labels from the HL7 v3 ActCode code system.
const (
	LabelHIV = "HIV" // HIV/AIDS
	LabelPSY = "PSY" // psychiatry
	LabelSDV = "SDV" // sexual and domestic violence
	LabelETH = "ETH" // substance abuse
	LabelSTD = "STD" // sexually transmitted disease
)

// Handling instruction labels.
const (
	LabelNoRedisclosure = "NORELINK"
	LabelNoCollection   = "NOCOLLECT"
	LabelNoIntegration  = "NOINTEGRATE"
	LabelBreakGlass     = "BREAK-THE-GLASS"
)

// SecurityLabelSystem is the FHIR code system URI for confidentiality classifications.
const SecurityLabelSystem = "http://terminology.hl7.org/CodeSystem/v3-Confidentiality"

// ActCodeSystem is the FHIR code system URI for act codes including sensitivity labels.
const ActCodeSystem = "http://terminology.hl7.org/CodeSystem/v3-ActCode"

// confidentialityOrder maps confidentiality codes to their numeric level for
// hierarchical comparison. Higher values indicate more restricted access.
var confidentialityOrder = map[string]int{
	LabelUnrestricted:   0,
	LabelLow:            1,
	LabelModerate:       2,
	LabelNormal:         3,
	LabelRestricted:     4,
	LabelVeryRestricted: 5,
}

// ConfidentialityLevel returns a numeric level for the given confidentiality code.
// Higher values mean more restricted. Unknown codes return -1.
func ConfidentialityLevel(code string) int {
	if level, ok := confidentialityOrder[code]; ok {
		return level
	}
	return -1
}

// SecurityContext represents the security context of the current request, capturing
// the caller's authorization level for security label enforcement.
type SecurityContext struct {
	// MaxConfidentiality is the highest confidentiality level the user can access.
	// Must be one of the confidentiality constants (U, L, M, N, R, V).
	MaxConfidentiality string

	// AllowedLabels contains the specific sensitivity labels the user is permitted
	// to access (e.g., HIV, PSY, SDV). If empty, only confidentiality is checked.
	AllowedLabels []string

	// BreakGlass indicates whether break-the-glass override is active. When true,
	// all security label checks are bypassed.
	BreakGlass bool

	// Purpose is the purpose of use for the request (e.g., TREAT, ETREAT, HPAYMT).
	Purpose string
}

// SecurityContextFromRequest extracts a SecurityContext from HTTP request headers.
//
// Headers used:
//
//	X-Security-Max-Confidentiality: single code (default "N")
//	X-Security-Labels: comma-separated list of allowed sensitivity labels
//	X-Break-Glass: "true" to enable break-the-glass override
//	X-Purpose-Of-Use: purpose code
func SecurityContextFromRequest(r *http.Request) *SecurityContext {
	sc := &SecurityContext{
		MaxConfidentiality: LabelNormal,
	}

	if v := r.Header.Get("X-Security-Max-Confidentiality"); v != "" {
		sc.MaxConfidentiality = v
	}

	if v := r.Header.Get("X-Security-Labels"); v != "" {
		labels := strings.Split(v, ",")
		for i := range labels {
			labels[i] = strings.TrimSpace(labels[i])
		}
		sc.AllowedLabels = labels
	}

	if strings.EqualFold(r.Header.Get("X-Break-Glass"), "true") {
		sc.BreakGlass = true
	}

	if v := r.Header.Get("X-Purpose-Of-Use"); v != "" {
		sc.Purpose = v
	}

	return sc
}

// CanAccessResource checks whether the given security context allows access to a
// resource described by the provided generic meta map. The meta map is expected to
// follow the FHIR JSON structure with an optional "security" key containing an
// array of coding objects (each with "system" and "code" string fields).
//
// Access rules:
//  1. Break-glass always grants access.
//  2. Confidentiality labels on the resource must not exceed the caller's
//     MaxConfidentiality level.
//  3. Sensitivity labels (HIV, PSY, etc.) on the resource must appear in the
//     caller's AllowedLabels list.
func CanAccessResource(ctx *SecurityContext, resourceMeta map[string]interface{}) bool {
	if ctx.BreakGlass {
		return true
	}

	securityCodings := extractSecurityCodings(resourceMeta)
	if len(securityCodings) == 0 {
		return true
	}

	maxLevel := ConfidentialityLevel(ctx.MaxConfidentiality)

	allowedSet := make(map[string]bool, len(ctx.AllowedLabels))
	for _, l := range ctx.AllowedLabels {
		allowedSet[l] = true
	}

	for _, coding := range securityCodings {
		code := coding.code

		// Check confidentiality hierarchy.
		if coding.system == SecurityLabelSystem || coding.system == "" {
			resourceLevel := ConfidentialityLevel(code)
			if resourceLevel >= 0 && maxLevel >= 0 && resourceLevel > maxLevel {
				return false
			}
		}

		// Check sensitivity labels.
		if coding.system == ActCodeSystem || coding.system == "" {
			if isSensitivityLabel(code) && !allowedSet[code] {
				return false
			}
		}
	}

	return true
}

// securityCoding is a lightweight internal representation of a coding tuple.
type securityCoding struct {
	system string
	code   string
}

// extractSecurityCodings pulls security codings from a generic meta map.
func extractSecurityCodings(meta map[string]interface{}) []securityCoding {
	if meta == nil {
		return nil
	}

	securityRaw, ok := meta["security"]
	if !ok {
		return nil
	}

	securityList, ok := securityRaw.([]interface{})
	if !ok {
		return nil
	}

	codings := make([]securityCoding, 0, len(securityList))
	for _, item := range securityList {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		sc := securityCoding{}
		if v, ok := m["system"].(string); ok {
			sc.system = v
		}
		if v, ok := m["code"].(string); ok {
			sc.code = v
		}
		if sc.code != "" {
			codings = append(codings, sc)
		}
	}

	return codings
}

// isSensitivityLabel returns true if the code is a recognized data sensitivity label.
func isSensitivityLabel(code string) bool {
	switch code {
	case LabelHIV, LabelPSY, LabelSDV, LabelETH, LabelSTD:
		return true
	}
	return false
}

// SecurityLabelMiddleware returns Echo middleware that enforces FHIR security labels.
// It intercepts responses and:
//   - For single resources: returns 403 Forbidden if the resource's security labels
//     exceed the caller's authorization.
//   - For Bundle search results: filters out entries whose resources the caller is
//     not authorized to see, adjusting the total count accordingly.
//   - Honors the X-Break-Glass header to bypass label checks.
func SecurityLabelMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sc := SecurityContextFromRequest(c.Request())

			// Capture the response body for post-processing.
			rec := &securityLabelRecorder{
				ResponseWriter: c.Response().Writer,
				body:           &bytes.Buffer{},
			}
			c.Response().Writer = rec

			if err := next(c); err != nil {
				// Restore the original writer so Echo's error handler
				// can write directly to the real response.
				c.Response().Writer = rec.ResponseWriter
				return err
			}

			// Restore the original writer and reset Echo's committed state so
			// we can write a new response (possibly with a different status).
			c.Response().Writer = rec.ResponseWriter
			c.Response().Committed = false
			c.Response().Status = 0
			c.Response().Size = 0

			// Parse the response body.
			var resource map[string]interface{}
			if err := json.Unmarshal(rec.body.Bytes(), &resource); err != nil {
				// Not JSON; replay the original status and body.
				writeReplayResponse(c, rec.statusCode, rec.body.Bytes())
				return nil
			}

			resourceType, _ := resource["resourceType"].(string)

			// Skip OperationOutcome responses; they are error responses and should
			// not be subject to security label filtering.
			if resourceType == "OperationOutcome" {
				writeReplayResponse(c, rec.statusCode, rec.body.Bytes())
				return nil
			}

			if resourceType == "Bundle" {
				resource = filterBundleEntries(sc, resource)
			} else {
				meta, _ := resource["meta"].(map[string]interface{})
				if !CanAccessResource(sc, meta) {
					outcome := NewOperationOutcome("error", "forbidden",
						"Access denied: resource security labels exceed authorization")
					return c.JSON(http.StatusForbidden, outcome)
				}
			}

			result, err := json.Marshal(resource)
			if err != nil {
				writeReplayResponse(c, rec.statusCode, rec.body.Bytes())
				return nil
			}

			c.Response().Header().Set("Content-Type", "application/fhir+json")
			writeReplayResponse(c, rec.statusCode, result)
			return nil
		}
	}
}

// filterBundleEntries removes entries from a Bundle whose resources the caller is
// not authorized to view, and updates the total count.
func filterBundleEntries(sc *SecurityContext, bundle map[string]interface{}) map[string]interface{} {
	entriesRaw, ok := bundle["entry"].([]interface{})
	if !ok {
		return bundle
	}

	filtered := make([]interface{}, 0, len(entriesRaw))
	for _, entryRaw := range entriesRaw {
		entry, ok := entryRaw.(map[string]interface{})
		if !ok {
			continue
		}
		res, ok := entry["resource"].(map[string]interface{})
		if !ok {
			filtered = append(filtered, entryRaw)
			continue
		}
		meta, _ := res["meta"].(map[string]interface{})
		if CanAccessResource(sc, meta) {
			filtered = append(filtered, entryRaw)
		}
	}

	bundle["entry"] = filtered
	bundle["total"] = float64(len(filtered))
	return bundle
}

// writeReplayResponse writes a buffered status code and body through Echo's
// response writer, which properly tracks committed state.
func writeReplayResponse(c echo.Context, statusCode int, body []byte) {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	c.Response().WriteHeader(statusCode)
	c.Response().Write(body) //nolint:errcheck
}

// securityLabelRecorder captures the response body and status code for security
// label post-processing. It buffers both the status code and body so the
// middleware can decide whether to forward the original response or replace it.
type securityLabelRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (r *securityLabelRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *securityLabelRecorder) WriteHeader(code int) {
	// Buffer the status code; do NOT forward to the underlying writer yet.
	r.statusCode = code
}
