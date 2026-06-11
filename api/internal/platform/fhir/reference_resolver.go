package fhir

// Reference resolution: rewrite internal-UUID references in outgoing FHIR
// resources to the target resource's external FHIR id.
//
// Why this exists
// ---------------
// Resources are stored with UUID foreign keys (e.g. condition.patient_id), but
// each resource is *addressed* and *searched* by its external `fhir_id` (e.g.
// "patient-john-smith"). Domain `ToFHIR()` methods currently emit references
// using the raw UUID ("Patient/<uuid>"), which does not match the id the target
// resource is served under — breaking reference integrity and US Core search
// conformance (Inferno: "reference did not match search parameter").
//
// The correct long-term fix is to resolve UUID -> fhir_id at the data/query
// layer when each model is built. That touches ~60 call sites across ~20
// domains, so as a bounded, reversible, well-tested step we resolve references
// centrally on the way out — mirroring the existing projection/_include
// response-transform middleware pattern in this package. The walker and the
// resolution policy are pure and unit-tested; only the lookup is I/O.

import (
	"context"
	"strings"
)

// resolvableRefTables maps a referenced FHIR resource type to the database
// table that holds its (id, fhir_id) mapping. Only types that (a) are targets
// of references in our resources and (b) carry a fhir_id column are listed.
// Table names are singular, matching the actual schema.
var resolvableRefTables = map[string]string{
	"Patient":            "patient",
	"Practitioner":       "practitioner",
	"PractitionerRole":   "practitioner_role",
	"Organization":       "organization",
	"Location":           "location",
	"Encounter":          "encounter",
	"Condition":          "condition",
	"Observation":        "observation",
	"Procedure":          "procedure",
	"MedicationRequest":  "medication_request",
	"Medication":         "medication",
	"AllergyIntolerance": "allergy_intolerance",
	"DiagnosticReport":   "diagnostic_report",
	"CarePlan":           "care_plan",
	"CareTeam":           "care_team",
	"Immunization":       "immunization",
	"Device":             "device",
	"Goal":               "goal",
	"DocumentReference":  "document_reference",
	"Coverage":           "coverage",
}

// ReferenceLookup resolves internal UUIDs to fhir_ids for a single resource
// type/table. Implementations should batch and may cache. It is the only I/O
// seam in reference resolution, which keeps the walker unit-testable.
//
// table is the database table name (e.g. "patient"); ids are UUID strings.
// The returned map is keyed by UUID; ids with no row are simply omitted.
type ReferenceLookup func(ctx context.Context, table string, ids []string) (map[string]string, error)

// ReferenceResolver rewrites UUID references to fhir_id references across a
// FHIR resource (and any nested/contained resources), using a ReferenceLookup
// for the actual id mapping. A resolver carries a per-instance cache, so it
// should be constructed per request (cheap) to bound cache lifetime to the
// response being built.
type ReferenceResolver struct {
	lookup ReferenceLookup
	cache  map[string]string // key: "Table/uuid" -> fhir_id
}

// NewReferenceResolver builds a resolver backed by the given lookup.
func NewReferenceResolver(lookup ReferenceLookup) *ReferenceResolver {
	return &ReferenceResolver{lookup: lookup, cache: make(map[string]string)}
}

// Resolve rewrites, in place, every "Type/<uuid>" reference found anywhere in
// resource to "Type/<fhir_id>" when the type is a known reference target and a
// mapping exists. References that are not UUID-shaped, target unknown types, or
// have no matching row are left unchanged. Resolve is safe to call on a single
// resource or a Bundle entry's resource map.
func (r *ReferenceResolver) Resolve(ctx context.Context, resource map[string]interface{}) {
	if resource == nil {
		return
	}
	// Pass 1: collect candidate UUIDs grouped by referenced type.
	pendingByType := map[string]map[string]struct{}{} // refType -> set of uuids
	r.walk(resource, func(refType, refID string) {
		if _, ok := resolvableRefTables[refType]; !ok || !isUUID(refID) {
			return
		}
		if _, cached := r.cache[refType+"/"+refID]; cached {
			return
		}
		if pendingByType[refType] == nil {
			pendingByType[refType] = map[string]struct{}{}
		}
		pendingByType[refType][refID] = struct{}{}
	})

	// Pass 2: batch-resolve uncached UUIDs per type and populate the cache.
	for refType, idSet := range pendingByType {
		ids := make([]string, 0, len(idSet))
		for id := range idSet {
			ids = append(ids, id)
		}
		mapped, err := r.lookup(ctx, resolvableRefTables[refType], ids)
		if err != nil {
			// Resolution is best-effort: on lookup failure we leave references
			// as-is rather than corrupt the response. The caller logs.
			continue
		}
		for uuidStr, fhirID := range mapped {
			if fhirID != "" {
				r.cache[refType+"/"+uuidStr] = fhirID
			}
		}
	}

	// Pass 3: rewrite references in place from the cache.
	r.rewrite(resource)
}

// walk visits every reference string ("Type/id") in the resource graph and
// invokes fn(type, id). It descends into nested maps and slices, including
// `contained` resources.
func (r *ReferenceResolver) walk(node interface{}, fn func(refType, refID string)) {
	switch v := node.(type) {
	case map[string]interface{}:
		for key, val := range v {
			if key == "reference" {
				if s, ok := val.(string); ok {
					if t, id, ok := splitReference(s); ok {
						fn(t, id)
					}
				}
				continue
			}
			r.walk(val, fn)
		}
	case []interface{}:
		for _, item := range v {
			r.walk(item, fn)
		}
	}
}

// rewrite replaces reference strings in place using the resolver cache.
func (r *ReferenceResolver) rewrite(node interface{}) {
	switch v := node.(type) {
	case map[string]interface{}:
		for key, val := range v {
			if key == "reference" {
				if s, ok := val.(string); ok {
					if t, id, ok := splitReference(s); ok {
						if fhirID, hit := r.cache[t+"/"+id]; hit {
							v[key] = t + "/" + fhirID
						}
					}
				}
				continue
			}
			r.rewrite(val)
		}
	case []interface{}:
		for _, item := range v {
			r.rewrite(item)
		}
	}
}

// splitReference parses a relative FHIR reference "ResourceType/id" into its
// parts. It returns ok=false for absolute URLs, contained refs ("#x"), urn:
// references, and anything not in Type/id shape — none of which we rewrite.
func splitReference(ref string) (resourceType, id string, ok bool) {
	if ref == "" || strings.HasPrefix(ref, "#") || strings.Contains(ref, "://") || strings.HasPrefix(ref, "urn:") {
		return "", "", false
	}
	slash := strings.IndexByte(ref, '/')
	if slash <= 0 || slash == len(ref)-1 {
		return "", "", false
	}
	// A relative reference has exactly one slash separating type and id.
	if strings.IndexByte(ref[slash+1:], '/') != -1 {
		return "", "", false
	}
	return ref[:slash], ref[slash+1:], true
}
