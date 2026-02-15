package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// mockResolver is a test implementation of ResourceResolver.
type mockResolver struct {
	resources map[string]map[string]interface{}
}

func (m *mockResolver) ResolveReference(ctx context.Context, ref string) (map[string]interface{}, error) {
	r, ok := m.resources[ref]
	if !ok {
		return nil, fmt.Errorf("not found: %s", ref)
	}
	return r, nil
}

// newTestComposition creates a minimal valid Composition for testing.
func newTestComposition() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Composition",
		"id":           "comp-1",
		"status":       "final",
		"type":         map[string]interface{}{"text": "Progress note"},
		"date":         "2024-01-15",
		"author": []interface{}{
			map[string]interface{}{"reference": "Practitioner/pract-1"},
		},
		"title":   "Progress Note",
		"subject": map[string]interface{}{"reference": "Patient/pat-1"},
	}
}

// newTestResolver creates a mock resolver preloaded with common test resources.
func newTestResolver() *mockResolver {
	return &mockResolver{
		resources: map[string]map[string]interface{}{
			"Patient/pat-1": {
				"resourceType": "Patient",
				"id":           "pat-1",
				"name":         []interface{}{map[string]interface{}{"family": "Smith"}},
			},
			"Practitioner/pract-1": {
				"resourceType": "Practitioner",
				"id":           "pract-1",
				"name":         []interface{}{map[string]interface{}{"family": "Jones"}},
			},
			"Practitioner/pract-2": {
				"resourceType": "Practitioner",
				"id":           "pract-2",
				"name":         []interface{}{map[string]interface{}{"family": "Williams"}},
			},
			"Encounter/enc-1": {
				"resourceType": "Encounter",
				"id":           "enc-1",
				"status":       "finished",
			},
			"Organization/org-1": {
				"resourceType": "Organization",
				"id":           "org-1",
				"name":         "Test Hospital",
			},
			"Condition/cond-1": {
				"resourceType": "Condition",
				"id":           "cond-1",
				"subject":      map[string]interface{}{"reference": "Patient/pat-1"},
			},
			"Observation/obs-1": {
				"resourceType": "Observation",
				"id":           "obs-1",
				"status":       "final",
				"code":         map[string]interface{}{"text": "BP"},
			},
			"MedicationRequest/med-1": {
				"resourceType": "MedicationRequest",
				"id":           "med-1",
				"status":       "active",
			},
			"Composition/comp-1": {
				"resourceType": "Composition",
				"id":           "comp-1",
				"status":       "final",
				"type":         map[string]interface{}{"text": "Progress note"},
				"date":         "2024-01-15",
				"author": []interface{}{
					map[string]interface{}{"reference": "Practitioner/pract-1"},
				},
				"title":   "Progress Note",
				"subject": map[string]interface{}{"reference": "Patient/pat-1"},
			},
		},
	}
}

// =========== DocumentGenerator Tests ===========

func TestDocumentGenerator_BasicDocument(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)

	composition := newTestComposition()
	bundle, err := gen.GenerateDocument(context.Background(), composition, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, ok := bundle["entry"].([]interface{})
	if !ok {
		t.Fatal("expected entry array in bundle")
	}

	// Composition + Patient + Practitioner = 3 entries
	if len(entries) != 3 {
		t.Errorf("expected 3 entries (Composition, Patient, Practitioner), got %d", len(entries))
	}
}

func TestDocumentGenerator_BundleType(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)

	composition := newTestComposition()
	bundle, err := gen.GenerateDocument(context.Background(), composition, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bundle["type"] != "document" {
		t.Errorf("expected Bundle.type 'document', got %v", bundle["type"])
	}
}

func TestDocumentGenerator_CompositionFirst(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)

	composition := newTestComposition()
	bundle, err := gen.GenerateDocument(context.Background(), composition, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := bundle["entry"].([]interface{})
	firstEntry := entries[0].(map[string]interface{})

	if firstEntry["fullUrl"] != "Composition/comp-1" {
		t.Errorf("expected first entry fullUrl 'Composition/comp-1', got %v", firstEntry["fullUrl"])
	}

	// Verify the resource inside the first entry is the Composition.
	raw := firstEntry["resource"].(json.RawMessage)
	var res map[string]interface{}
	if err := json.Unmarshal(raw, &res); err != nil {
		t.Fatalf("failed to unmarshal first entry resource: %v", err)
	}
	if res["resourceType"] != "Composition" {
		t.Errorf("expected first entry resource to be Composition, got %v", res["resourceType"])
	}
}

func TestDocumentGenerator_SectionEntries(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)

	composition := newTestComposition()
	composition["section"] = []interface{}{
		map[string]interface{}{
			"title": "Problems",
			"entry": []interface{}{
				map[string]interface{}{"reference": "Condition/cond-1"},
			},
		},
	}

	bundle, err := gen.GenerateDocument(context.Background(), composition, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := bundle["entry"].([]interface{})

	// Composition + Patient + Practitioner + Condition = 4
	if len(entries) != 4 {
		t.Errorf("expected 4 entries, got %d", len(entries))
	}

	// Check that Condition is present.
	found := false
	for _, e := range entries {
		entry := e.(map[string]interface{})
		if entry["fullUrl"] == "Condition/cond-1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Condition/cond-1 in bundle entries")
	}
}

func TestDocumentGenerator_NestedSections(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)

	composition := newTestComposition()
	composition["section"] = []interface{}{
		map[string]interface{}{
			"title": "Problems",
			"entry": []interface{}{
				map[string]interface{}{"reference": "Condition/cond-1"},
			},
			"section": []interface{}{
				map[string]interface{}{
					"title": "Observations",
					"entry": []interface{}{
						map[string]interface{}{"reference": "Observation/obs-1"},
					},
				},
			},
		},
	}

	bundle, err := gen.GenerateDocument(context.Background(), composition, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := bundle["entry"].([]interface{})

	// Composition + Patient + Practitioner + Condition + Observation = 5
	if len(entries) != 5 {
		t.Errorf("expected 5 entries, got %d", len(entries))
	}

	foundCondition := false
	foundObservation := false
	for _, e := range entries {
		entry := e.(map[string]interface{})
		if entry["fullUrl"] == "Condition/cond-1" {
			foundCondition = true
		}
		if entry["fullUrl"] == "Observation/obs-1" {
			foundObservation = true
		}
	}
	if !foundCondition {
		t.Error("expected Condition/cond-1 in bundle entries")
	}
	if !foundObservation {
		t.Error("expected Observation/obs-1 in bundle entries from nested section")
	}
}

func TestDocumentGenerator_UnresolvableReference(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)

	composition := newTestComposition()
	composition["section"] = []interface{}{
		map[string]interface{}{
			"entry": []interface{}{
				map[string]interface{}{"reference": "Condition/nonexistent"},
			},
		},
	}

	bundle, err := gen.GenerateDocument(context.Background(), composition, false)
	if err != nil {
		t.Fatalf("expected no error when reference is unresolvable, got: %v", err)
	}

	entries := bundle["entry"].([]interface{})

	// Composition + Patient + Practitioner = 3 (unresolvable is skipped)
	if len(entries) != 3 {
		t.Errorf("expected 3 entries (skipping unresolvable), got %d", len(entries))
	}
}

func TestDocumentGenerator_NilComposition(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)

	_, err := gen.GenerateDocument(context.Background(), nil, false)
	if err == nil {
		t.Fatal("expected error for nil composition")
	}

	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("expected error message to mention nil, got: %v", err)
	}
}

func TestDocumentGenerator_InvalidComposition(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)

	// Missing resourceType.
	composition := map[string]interface{}{
		"id":     "comp-1",
		"status": "final",
	}

	_, err := gen.GenerateDocument(context.Background(), composition, false)
	if err == nil {
		t.Fatal("expected error for invalid composition (missing resourceType)")
	}

	// Wrong resourceType.
	composition2 := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-1",
	}

	_, err = gen.GenerateDocument(context.Background(), composition2, false)
	if err == nil {
		t.Fatal("expected error for non-Composition resourceType")
	}

	// Missing required field (title).
	composition3 := map[string]interface{}{
		"resourceType": "Composition",
		"id":           "comp-1",
		"status":       "final",
		"type":         map[string]interface{}{"text": "note"},
		"date":         "2024-01-15",
		"author":       []interface{}{map[string]interface{}{"reference": "Practitioner/pract-1"}},
		// missing "title"
	}

	_, err = gen.GenerateDocument(context.Background(), composition3, false)
	if err == nil {
		t.Fatal("expected error for Composition missing title")
	}

	if !strings.Contains(err.Error(), "title") {
		t.Errorf("expected error to mention 'title', got: %v", err)
	}
}

func TestDocumentGenerator_DeduplicateReferences(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)

	composition := newTestComposition()
	// Patient/pat-1 is both subject and a section entry reference.
	composition["section"] = []interface{}{
		map[string]interface{}{
			"entry": []interface{}{
				map[string]interface{}{"reference": "Patient/pat-1"},
			},
		},
	}

	bundle, err := gen.GenerateDocument(context.Background(), composition, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := bundle["entry"].([]interface{})

	// Composition + Patient + Practitioner = 3 (Patient not duplicated)
	if len(entries) != 3 {
		t.Errorf("expected 3 entries (deduplicated), got %d", len(entries))
	}

	// Count Patient entries.
	patientCount := 0
	for _, e := range entries {
		entry := e.(map[string]interface{})
		if entry["fullUrl"] == "Patient/pat-1" {
			patientCount++
		}
	}
	if patientCount != 1 {
		t.Errorf("expected exactly 1 Patient/pat-1 entry, got %d", patientCount)
	}
}

func TestDocumentGenerator_MultipleAuthors(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)

	composition := newTestComposition()
	composition["author"] = []interface{}{
		map[string]interface{}{"reference": "Practitioner/pract-1"},
		map[string]interface{}{"reference": "Practitioner/pract-2"},
	}

	bundle, err := gen.GenerateDocument(context.Background(), composition, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := bundle["entry"].([]interface{})

	// Composition + Patient + Practitioner1 + Practitioner2 = 4
	if len(entries) != 4 {
		t.Errorf("expected 4 entries, got %d", len(entries))
	}

	foundPract1 := false
	foundPract2 := false
	for _, e := range entries {
		entry := e.(map[string]interface{})
		if entry["fullUrl"] == "Practitioner/pract-1" {
			foundPract1 = true
		}
		if entry["fullUrl"] == "Practitioner/pract-2" {
			foundPract2 = true
		}
	}
	if !foundPract1 {
		t.Error("expected Practitioner/pract-1 in bundle entries")
	}
	if !foundPract2 {
		t.Error("expected Practitioner/pract-2 in bundle entries")
	}
}

func TestDocumentGenerator_HasTimestamp(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)

	composition := newTestComposition()
	bundle, err := gen.GenerateDocument(context.Background(), composition, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ts, ok := bundle["timestamp"].(string)
	if !ok || ts == "" {
		t.Error("expected Bundle to have a non-empty timestamp string")
	}
}

func TestDocumentGenerator_HasIdentifier(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)

	composition := newTestComposition()
	bundle, err := gen.GenerateDocument(context.Background(), composition, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	identifier, ok := bundle["identifier"].(map[string]interface{})
	if !ok {
		t.Fatal("expected Bundle to have an identifier object")
	}

	system, _ := identifier["system"].(string)
	if system != "urn:ietf:rfc:3986" {
		t.Errorf("expected identifier.system 'urn:ietf:rfc:3986', got %q", system)
	}

	value, _ := identifier["value"].(string)
	if !strings.HasPrefix(value, "urn:uuid:") {
		t.Errorf("expected identifier.value to start with 'urn:uuid:', got %q", value)
	}

	// Verify the UUID portion is non-empty.
	uuidPart := strings.TrimPrefix(value, "urn:uuid:")
	if len(uuidPart) < 36 {
		t.Errorf("expected a full UUID in identifier.value, got %q", uuidPart)
	}
}

// =========== collectReferences Tests ===========

func TestCollectReferences_AllFields(t *testing.T) {
	composition := map[string]interface{}{
		"resourceType": "Composition",
		"id":           "comp-1",
		"subject":      map[string]interface{}{"reference": "Patient/pat-1"},
		"author": []interface{}{
			map[string]interface{}{"reference": "Practitioner/pract-1"},
		},
		"custodian": map[string]interface{}{"reference": "Organization/org-1"},
		"encounter": map[string]interface{}{"reference": "Encounter/enc-1"},
		"attester": []interface{}{
			map[string]interface{}{
				"party": map[string]interface{}{"reference": "Practitioner/pract-2"},
			},
		},
		"section": []interface{}{
			map[string]interface{}{
				"entry": []interface{}{
					map[string]interface{}{"reference": "Condition/cond-1"},
				},
				"author": []interface{}{
					map[string]interface{}{"reference": "Practitioner/pract-1"},
				},
			},
		},
	}

	refs := collectReferences(composition)

	expected := map[string]bool{
		"Patient/pat-1":        false,
		"Practitioner/pract-1": false,
		"Organization/org-1":   false,
		"Encounter/enc-1":      false,
		"Practitioner/pract-2": false,
		"Condition/cond-1":     false,
	}

	for _, ref := range refs {
		if _, ok := expected[ref]; ok {
			expected[ref] = true
		}
	}

	for ref, found := range expected {
		if !found {
			t.Errorf("expected reference %q not found in collected references: %v", ref, refs)
		}
	}
}

func TestCollectReferences_Empty(t *testing.T) {
	composition := map[string]interface{}{
		"resourceType": "Composition",
		"id":           "comp-1",
		"status":       "final",
		"title":        "Empty note",
	}

	refs := collectReferences(composition)

	if len(refs) != 0 {
		t.Errorf("expected 0 references for composition with no refs, got %d: %v", len(refs), refs)
	}
}

func TestCollectReferences_DuplicatesRemoved(t *testing.T) {
	composition := map[string]interface{}{
		"resourceType": "Composition",
		"id":           "comp-1",
		"subject":      map[string]interface{}{"reference": "Patient/pat-1"},
		"author": []interface{}{
			map[string]interface{}{"reference": "Practitioner/pract-1"},
		},
		"section": []interface{}{
			map[string]interface{}{
				"entry": []interface{}{
					map[string]interface{}{"reference": "Patient/pat-1"},
					map[string]interface{}{"reference": "Practitioner/pract-1"},
				},
			},
		},
	}

	refs := collectReferences(composition)

	// Patient/pat-1 and Practitioner/pract-1 each appear twice but should only be collected once.
	counts := make(map[string]int)
	for _, ref := range refs {
		counts[ref]++
	}

	for ref, count := range counts {
		if count > 1 {
			t.Errorf("reference %q appears %d times, expected 1", ref, count)
		}
	}

	if len(refs) != 2 {
		t.Errorf("expected 2 unique references, got %d: %v", len(refs), refs)
	}
}

// =========== DocumentHandler Tests ===========

func TestDocumentHandler_POST_Valid(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)
	handler := NewDocumentHandler(gen)

	e := echo.New()
	body := `{
		"resourceType": "Composition",
		"id": "comp-1",
		"status": "final",
		"type": {"text": "Progress note"},
		"date": "2024-01-15",
		"author": [{"reference": "Practitioner/pract-1"}],
		"title": "Progress Note",
		"subject": {"reference": "Patient/pat-1"}
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/Composition/$document", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GenerateDocumentFromBody(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected resourceType Bundle, got %v", bundle["resourceType"])
	}
	if bundle["type"] != "document" {
		t.Errorf("expected type document, got %v", bundle["type"])
	}

	entries, ok := bundle["entry"].([]interface{})
	if !ok {
		t.Fatal("expected entry array")
	}
	if len(entries) < 1 {
		t.Error("expected at least 1 entry")
	}
}

func TestDocumentHandler_POST_EmptyBody(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)
	handler := NewDocumentHandler(gen)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Composition/$document", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GenerateDocumentFromBody(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestDocumentHandler_POST_InvalidJSON(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)
	handler := NewDocumentHandler(gen)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Composition/$document", strings.NewReader("{not valid json}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GenerateDocumentFromBody(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestDocumentHandler_GET_Valid(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)
	handler := NewDocumentHandler(gen)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Composition/comp-1/$document", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("comp-1")

	err := handler.GenerateDocument(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected resourceType Bundle, got %v", bundle["resourceType"])
	}
	if bundle["type"] != "document" {
		t.Errorf("expected type document, got %v", bundle["type"])
	}

	entries, ok := bundle["entry"].([]interface{})
	if !ok {
		t.Fatal("expected entry array")
	}

	// The composition in the resolver also has subject and author references.
	if len(entries) < 3 {
		t.Errorf("expected at least 3 entries, got %d", len(entries))
	}

	// First entry should be the Composition.
	firstEntry := entries[0].(map[string]interface{})
	if firstEntry["fullUrl"] != "Composition/comp-1" {
		t.Errorf("expected first entry fullUrl 'Composition/comp-1', got %v", firstEntry["fullUrl"])
	}
}

func TestDocumentHandler_GET_NotFound(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)
	handler := NewDocumentHandler(gen)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Composition/nonexistent/$document", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := handler.GenerateDocument(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if outcome["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", outcome["resourceType"])
	}
}

func TestDocumentHandler_RegisterRoutes(t *testing.T) {
	resolver := newTestResolver()
	gen := NewDocumentGenerator(resolver)
	handler := NewDocumentHandler(gen)

	e := echo.New()
	fhirGroup := e.Group("/fhir")
	handler.RegisterRoutes(fhirGroup)

	routes := e.Routes()
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}

	expected := []string{
		"GET:/fhir/Composition/:id/$document",
		"POST:/fhir/Composition/$document",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s (registered: %v)", path, routePaths)
		}
	}
}
