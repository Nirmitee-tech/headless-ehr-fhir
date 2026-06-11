package fhir

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

// fakeLookup builds a ReferenceLookup from a static table->(uuid->fhir_id) map
// and records how many times it was called, so tests can assert batching/caching.
func fakeLookup(data map[string]map[string]string, calls *int) ReferenceLookup {
	return func(_ context.Context, table string, ids []string) (map[string]string, error) {
		if calls != nil {
			*calls++
		}
		out := map[string]string{}
		for _, id := range ids {
			if fhirID, ok := data[table][id]; ok {
				out[id] = fhirID
			}
		}
		return out, nil
	}
}

const (
	uuidJohn   = "a0000000-0000-4000-8000-000000000100"
	uuidDrS    = "b0000000-0000-4000-8000-000000000200"
	uuidEnc    = "c0000000-0000-4000-8000-000000000300"
	fhirJohn   = "patient-john-smith"
	fhirDrS    = "practitioner-dr-smith"
	fhirEncVal = "encounter-john-office"
)

func sampleData() map[string]map[string]string {
	return map[string]map[string]string{
		"patient":      {uuidJohn: fhirJohn},
		"practitioner": {uuidDrS: fhirDrS},
		"encounter":    {uuidEnc: fhirEncVal},
	}
}

func TestResolve_RewritesUUIDReferences(t *testing.T) {
	res := map[string]interface{}{
		"resourceType": "Condition",
		"id":           "condition-john-diabetes",
		"subject":      map[string]interface{}{"reference": "Patient/" + uuidJohn},
		"recorder":     map[string]interface{}{"reference": "Practitioner/" + uuidDrS},
		"encounter":    map[string]interface{}{"reference": "Encounter/" + uuidEnc},
	}
	NewReferenceResolver(fakeLookup(sampleData(), nil)).Resolve(context.Background(), res)

	checks := map[string]string{
		"subject":   "Patient/" + fhirJohn,
		"recorder":  "Practitioner/" + fhirDrS,
		"encounter": "Encounter/" + fhirEncVal,
	}
	for field, want := range checks {
		got := res[field].(map[string]interface{})["reference"]
		if got != want {
			t.Errorf("%s: got %q, want %q", field, got, want)
		}
	}
}

func TestResolve_LeavesFhirIdReferencesUntouched(t *testing.T) {
	// Idempotency: a reference already in fhir_id form must not change, and must
	// not trigger a lookup (UUID pattern guards it).
	calls := 0
	res := map[string]interface{}{
		"resourceType": "Observation",
		"subject":      map[string]interface{}{"reference": "Patient/" + fhirJohn},
	}
	NewReferenceResolver(fakeLookup(sampleData(), &calls)).Resolve(context.Background(), res)

	if got := res["subject"].(map[string]interface{})["reference"]; got != "Patient/"+fhirJohn {
		t.Errorf("fhir_id ref changed: got %q", got)
	}
	if calls != 0 {
		t.Errorf("expected no lookup for non-UUID reference, got %d calls", calls)
	}
}

func TestResolve_UnknownTypeAndMissingRowLeftAsIs(t *testing.T) {
	missingUUID := "d0000000-0000-4000-8000-000000000999"
	res := map[string]interface{}{
		"resourceType": "Condition",
		// Unknown target type -> not in referenceTargetTable -> untouched.
		"asserter": map[string]interface{}{"reference": "RelatedPerson/" + uuidJohn},
		// Known type but no row in lookup -> untouched (not corrupted).
		"subject": map[string]interface{}{"reference": "Patient/" + missingUUID},
	}
	NewReferenceResolver(fakeLookup(sampleData(), nil)).Resolve(context.Background(), res)

	if got := res["asserter"].(map[string]interface{})["reference"]; got != "RelatedPerson/"+uuidJohn {
		t.Errorf("unknown-type ref changed: got %q", got)
	}
	if got := res["subject"].(map[string]interface{})["reference"]; got != "Patient/"+missingUUID {
		t.Errorf("missing-row ref changed: got %q", got)
	}
}

func TestResolve_NestedAndContainedAndArrays(t *testing.T) {
	res := map[string]interface{}{
		"resourceType": "CarePlan",
		"subject":      map[string]interface{}{"reference": "Patient/" + uuidJohn},
		// array of references
		"careTeam": []interface{}{
			map[string]interface{}{"reference": "Practitioner/" + uuidDrS},
		},
		// nested object containing a reference
		"activity": []interface{}{
			map[string]interface{}{
				"detail": map[string]interface{}{
					"performer": []interface{}{
						map[string]interface{}{"reference": "Practitioner/" + uuidDrS},
					},
				},
			},
		},
		// contained resource with its own references
		"contained": []interface{}{
			map[string]interface{}{
				"resourceType": "Goal",
				"subject":      map[string]interface{}{"reference": "Patient/" + uuidJohn},
			},
		},
	}
	NewReferenceResolver(fakeLookup(sampleData(), nil)).Resolve(context.Background(), res)

	if got := res["subject"].(map[string]interface{})["reference"]; got != "Patient/"+fhirJohn {
		t.Errorf("top subject: got %q", got)
	}
	if got := res["careTeam"].([]interface{})[0].(map[string]interface{})["reference"]; got != "Practitioner/"+fhirDrS {
		t.Errorf("careTeam array: got %q", got)
	}
	perf := res["activity"].([]interface{})[0].(map[string]interface{})["detail"].(map[string]interface{})["performer"].([]interface{})[0].(map[string]interface{})["reference"]
	if perf != "Practitioner/"+fhirDrS {
		t.Errorf("nested performer: got %q", perf)
	}
	contained := res["contained"].([]interface{})[0].(map[string]interface{})["subject"].(map[string]interface{})["reference"]
	if contained != "Patient/"+fhirJohn {
		t.Errorf("contained subject: got %q", contained)
	}
}

func TestResolve_BatchesAndCaches(t *testing.T) {
	// Two patient references + one practitioner across the resource should yield
	// exactly two lookups (one per table), not four — proving batching + dedupe.
	calls := 0
	res := map[string]interface{}{
		"resourceType": "Encounter",
		"subject":      map[string]interface{}{"reference": "Patient/" + uuidJohn},
		"participant": []interface{}{
			map[string]interface{}{"individual": map[string]interface{}{"reference": "Practitioner/" + uuidDrS}},
		},
		"basedOn": []interface{}{
			map[string]interface{}{"reference": "Patient/" + uuidJohn}, // duplicate patient
		},
	}
	NewReferenceResolver(fakeLookup(sampleData(), &calls)).Resolve(context.Background(), res)
	if calls != 2 {
		t.Errorf("expected 2 batched lookups (patient, practitioner), got %d", calls)
	}
}

func TestSplitReference(t *testing.T) {
	cases := []struct {
		in              string
		wantType, wantID string
		wantOK          bool
	}{
		{"Patient/abc", "Patient", "abc", true},
		{"Patient/" + uuidJohn, "Patient", uuidJohn, true},
		{"#contained", "", "", false},
		{"urn:uuid:1234", "", "", false},
		{"http://ex.com/fhir/Patient/1", "", "", false},
		{"Patient/", "", "", false},
		{"/abc", "", "", false},
		{"Patient/Group/1", "", "", false}, // too many segments
		{"", "", "", false},
	}
	for _, c := range cases {
		gotType, gotID, gotOK := splitReference(c.in)
		if gotType != c.wantType || gotID != c.wantID || gotOK != c.wantOK {
			t.Errorf("splitReference(%q) = (%q,%q,%v), want (%q,%q,%v)",
				c.in, gotType, gotID, gotOK, c.wantType, c.wantID, c.wantOK)
		}
	}
}

func TestResolve_RealisticBundleEntryJSON(t *testing.T) {
	// Exercise the path the middleware uses: unmarshal a resource from JSON,
	// resolve, and confirm the serialized output carries fhir_id references.
	raw := `{
	  "resourceType":"DiagnosticReport",
	  "id":"dr-1",
	  "subject":{"reference":"Patient/` + uuidJohn + `"},
	  "performer":[{"reference":"Practitioner/` + uuidDrS + `"}],
	  "encounter":{"reference":"Encounter/` + uuidEnc + `"}
	}`
	var res map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &res); err != nil {
		t.Fatal(err)
	}
	NewReferenceResolver(fakeLookup(sampleData(), nil)).Resolve(context.Background(), res)

	want := map[string]interface{}{
		"resourceType": "DiagnosticReport",
		"id":           "dr-1",
		"subject":      map[string]interface{}{"reference": "Patient/" + fhirJohn},
		"performer":    []interface{}{map[string]interface{}{"reference": "Practitioner/" + fhirDrS}},
		"encounter":    map[string]interface{}{"reference": "Encounter/" + fhirEncVal},
	}
	if !reflect.DeepEqual(res, want) {
		out, _ := json.MarshalIndent(res, "", "  ")
		t.Errorf("resolved resource mismatch:\n%s", out)
	}
}
