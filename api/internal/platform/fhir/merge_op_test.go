package fhir

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
)

// =============================================================================
// Survivorship Rules Tests (8+)
// =============================================================================

func TestSurvivorshipRules_TargetWinsGender(t *testing.T) {
	mdm := NewMDMService()
	source := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "src-1",
		"gender":       "female",
	}
	target := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "tgt-1",
		"gender":       "male",
	}
	result := mdm.ApplySurvivorshipRules(source, target)
	if result["gender"] != "male" {
		t.Errorf("expected target gender 'male', got %v", result["gender"])
	}
}

func TestSurvivorshipRules_TargetWinsBirthDate(t *testing.T) {
	mdm := NewMDMService()
	source := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "src-1",
		"birthDate":    "1990-01-01",
	}
	target := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "tgt-1",
		"birthDate":    "1990-06-15",
	}
	result := mdm.ApplySurvivorshipRules(source, target)
	if result["birthDate"] != "1990-06-15" {
		t.Errorf("expected target birthDate '1990-06-15', got %v", result["birthDate"])
	}
}

func TestSurvivorshipRules_MergeListsIdentifiers_DedupBySystem(t *testing.T) {
	mdm := NewMDMService()
	source := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "src-1",
		"identifier": []interface{}{
			map[string]interface{}{"system": "http://hospital.org/mrn", "value": "MRN-111"},
			map[string]interface{}{"system": "http://ssn", "value": "999-99-9999"},
		},
	}
	target := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "tgt-1",
		"identifier": []interface{}{
			map[string]interface{}{"system": "http://hospital.org/mrn", "value": "MRN-222"},
		},
	}
	result := mdm.ApplySurvivorshipRules(source, target)
	identifiers, ok := result["identifier"].([]interface{})
	if !ok {
		t.Fatal("expected identifier to be a slice")
	}
	// Should have 2: the target MRN (target wins for dedup by system) + the SSN from source
	if len(identifiers) != 2 {
		t.Errorf("expected 2 identifiers (dedup by system), got %d", len(identifiers))
	}
	// Check SSN is present from source
	found := false
	for _, id := range identifiers {
		m, _ := id.(map[string]interface{})
		if m["system"] == "http://ssn" {
			found = true
		}
	}
	if !found {
		t.Error("expected SSN identifier from source to be kept")
	}
}

func TestSurvivorshipRules_MergeListsTelecom_DedupBySystemValue(t *testing.T) {
	mdm := NewMDMService()
	source := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "src-1",
		"telecom": []interface{}{
			map[string]interface{}{"system": "phone", "value": "555-1111"},
			map[string]interface{}{"system": "email", "value": "john@example.com"},
		},
	}
	target := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "tgt-1",
		"telecom": []interface{}{
			map[string]interface{}{"system": "phone", "value": "555-1111"},
			map[string]interface{}{"system": "phone", "value": "555-2222"},
		},
	}
	result := mdm.ApplySurvivorshipRules(source, target)
	telecoms, ok := result["telecom"].([]interface{})
	if !ok {
		t.Fatal("expected telecom to be a slice")
	}
	// 555-1111 (dedup), 555-2222, john@example.com = 3
	if len(telecoms) != 3 {
		t.Errorf("expected 3 telecoms after dedup, got %d", len(telecoms))
	}
}

func TestSurvivorshipRules_MergeListsNames(t *testing.T) {
	mdm := NewMDMService()
	source := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "src-1",
		"name": []interface{}{
			map[string]interface{}{"use": "nickname", "family": "Smith", "given": []interface{}{"Johnny"}},
		},
	}
	target := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "tgt-1",
		"name": []interface{}{
			map[string]interface{}{"use": "official", "family": "Smith", "given": []interface{}{"John"}},
		},
	}
	result := mdm.ApplySurvivorshipRules(source, target)
	names, ok := result["name"].([]interface{})
	if !ok {
		t.Fatal("expected name to be a slice")
	}
	if len(names) != 2 {
		t.Errorf("expected 2 names merged, got %d", len(names))
	}
}

func TestSurvivorshipRules_MergeListsAddresses(t *testing.T) {
	mdm := NewMDMService()
	source := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "src-1",
		"address": []interface{}{
			map[string]interface{}{"use": "home", "line": []interface{}{"123 Main St"}, "city": "Springfield"},
		},
	}
	target := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "tgt-1",
		"address": []interface{}{
			map[string]interface{}{"use": "work", "line": []interface{}{"456 Oak Ave"}, "city": "Shelbyville"},
		},
	}
	result := mdm.ApplySurvivorshipRules(source, target)
	addrs, ok := result["address"].([]interface{})
	if !ok {
		t.Fatal("expected address to be a slice")
	}
	if len(addrs) != 2 {
		t.Errorf("expected 2 addresses merged, got %d", len(addrs))
	}
}

func TestSurvivorshipRules_SourceHasFieldTargetDoesNot(t *testing.T) {
	mdm := NewMDMService()
	source := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "src-1",
		"gender":       "female",
		"birthDate":    "1985-03-20",
	}
	target := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "tgt-1",
	}
	result := mdm.ApplySurvivorshipRules(source, target)
	// For target-wins fields, if target is absent, source value is kept.
	if result["gender"] != "female" {
		t.Errorf("expected source gender to be kept when target missing, got %v", result["gender"])
	}
	if result["birthDate"] != "1985-03-20" {
		t.Errorf("expected source birthDate to be kept when target missing, got %v", result["birthDate"])
	}
}

func TestSurvivorshipRules_BothHaveField_StrategyApplied(t *testing.T) {
	mdm := NewMDMService()
	source := map[string]interface{}{
		"resourceType":         "Patient",
		"id":                   "src-1",
		"gender":               "female",
		"managingOrganization": map[string]interface{}{"reference": "Organization/org-src"},
	}
	target := map[string]interface{}{
		"resourceType":         "Patient",
		"id":                   "tgt-1",
		"gender":               "male",
		"managingOrganization": map[string]interface{}{"reference": "Organization/org-tgt"},
	}
	result := mdm.ApplySurvivorshipRules(source, target)
	if result["gender"] != "male" {
		t.Errorf("expected target gender to win, got %v", result["gender"])
	}
	org, ok := result["managingOrganization"].(map[string]interface{})
	if !ok {
		t.Fatal("expected managingOrganization to be map")
	}
	if org["reference"] != "Organization/org-tgt" {
		t.Errorf("expected target managingOrganization, got %v", org["reference"])
	}
}

func TestSurvivorshipRules_MergeListsCommunication(t *testing.T) {
	mdm := NewMDMService()
	source := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "src-1",
		"communication": []interface{}{
			map[string]interface{}{"language": map[string]interface{}{"coding": []interface{}{map[string]interface{}{"code": "es"}}}},
		},
	}
	target := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "tgt-1",
		"communication": []interface{}{
			map[string]interface{}{"language": map[string]interface{}{"coding": []interface{}{map[string]interface{}{"code": "en"}}}},
		},
	}
	result := mdm.ApplySurvivorshipRules(source, target)
	comms, ok := result["communication"].([]interface{})
	if !ok {
		t.Fatal("expected communication to be a slice")
	}
	if len(comms) != 2 {
		t.Errorf("expected 2 communication entries merged, got %d", len(comms))
	}
}

func TestSurvivorshipRules_TargetWinsGeneralPractitioner(t *testing.T) {
	mdm := NewMDMService()
	source := map[string]interface{}{
		"resourceType":        "Patient",
		"id":                  "src-1",
		"generalPractitioner": []interface{}{map[string]interface{}{"reference": "Practitioner/src-doc"}},
	}
	target := map[string]interface{}{
		"resourceType":        "Patient",
		"id":                  "tgt-1",
		"generalPractitioner": []interface{}{map[string]interface{}{"reference": "Practitioner/tgt-doc"}},
	}
	result := mdm.ApplySurvivorshipRules(source, target)
	gp, ok := result["generalPractitioner"].([]interface{})
	if !ok {
		t.Fatal("expected generalPractitioner to be a slice")
	}
	if len(gp) != 1 {
		t.Errorf("expected 1 generalPractitioner (target-wins), got %d", len(gp))
	}
	gpMap, _ := gp[0].(map[string]interface{})
	if gpMap["reference"] != "Practitioner/tgt-doc" {
		t.Errorf("expected target's generalPractitioner, got %v", gpMap["reference"])
	}
}

func TestSurvivorshipRules_MaritalStatusMostRecent(t *testing.T) {
	mdm := NewMDMService()
	source := map[string]interface{}{
		"resourceType":  "Patient",
		"id":            "src-1",
		"maritalStatus": map[string]interface{}{"coding": []interface{}{map[string]interface{}{"code": "S"}}},
	}
	target := map[string]interface{}{
		"resourceType":  "Patient",
		"id":            "tgt-1",
		"maritalStatus": map[string]interface{}{"coding": []interface{}{map[string]interface{}{"code": "M"}}},
	}
	result := mdm.ApplySurvivorshipRules(source, target)
	ms, ok := result["maritalStatus"].(map[string]interface{})
	if !ok {
		t.Fatal("expected maritalStatus to be a map")
	}
	// most-recent: target wins on tie
	coding, _ := ms["coding"].([]interface{})
	if len(coding) == 0 {
		t.Fatal("expected coding in maritalStatus")
	}
	c, _ := coding[0].(map[string]interface{})
	if c["code"] != "M" {
		t.Errorf("expected most-recent (target wins on tie) maritalStatus 'M', got %v", c["code"])
	}
}

// =============================================================================
// Reference Rewriter Tests (8+)
// =============================================================================

func TestRewriter_SubjectReference(t *testing.T) {
	rw := NewReferenceRewriter()
	rw.AddResources("Observation", []map[string]interface{}{
		{
			"resourceType": "Observation",
			"id":           "obs-1",
			"subject":      map[string]interface{}{"reference": "Patient/src-1"},
		},
	})
	count := rw.RewriteReferences("src-1", "tgt-1")
	if count != 1 {
		t.Errorf("expected 1 reference rewritten, got %d", count)
	}
	obs := rw.resourceStore["Observation"][0]
	subj, _ := obs["subject"].(map[string]interface{})
	if subj["reference"] != "Patient/tgt-1" {
		t.Errorf("expected Patient/tgt-1, got %v", subj["reference"])
	}
}

func TestRewriter_PatientReference(t *testing.T) {
	rw := NewReferenceRewriter()
	rw.AddResources("AllergyIntolerance", []map[string]interface{}{
		{
			"resourceType": "AllergyIntolerance",
			"id":           "ai-1",
			"patient":      map[string]interface{}{"reference": "Patient/src-1"},
		},
	})
	count := rw.RewriteReferences("src-1", "tgt-1")
	if count != 1 {
		t.Errorf("expected 1 reference rewritten, got %d", count)
	}
	ai := rw.resourceStore["AllergyIntolerance"][0]
	pat, _ := ai["patient"].(map[string]interface{})
	if pat["reference"] != "Patient/tgt-1" {
		t.Errorf("expected Patient/tgt-1, got %v", pat["reference"])
	}
}

func TestRewriter_NestedReferences(t *testing.T) {
	rw := NewReferenceRewriter()
	rw.AddResources("Encounter", []map[string]interface{}{
		{
			"resourceType": "Encounter",
			"id":           "enc-1",
			"subject":      map[string]interface{}{"reference": "Patient/src-1"},
			"participant": []interface{}{
				map[string]interface{}{
					"individual": map[string]interface{}{"reference": "Patient/src-1"},
				},
			},
		},
	})
	count := rw.RewriteReferences("src-1", "tgt-1")
	if count != 2 {
		t.Errorf("expected 2 references rewritten, got %d", count)
	}
	enc := rw.resourceStore["Encounter"][0]
	subj, _ := enc["subject"].(map[string]interface{})
	if subj["reference"] != "Patient/tgt-1" {
		t.Errorf("subject: expected Patient/tgt-1, got %v", subj["reference"])
	}
	parts, _ := enc["participant"].([]interface{})
	part, _ := parts[0].(map[string]interface{})
	ind, _ := part["individual"].(map[string]interface{})
	if ind["reference"] != "Patient/tgt-1" {
		t.Errorf("individual: expected Patient/tgt-1, got %v", ind["reference"])
	}
}

func TestRewriter_NoReferencesToRewrite(t *testing.T) {
	rw := NewReferenceRewriter()
	rw.AddResources("Observation", []map[string]interface{}{
		{
			"resourceType": "Observation",
			"id":           "obs-1",
			"subject":      map[string]interface{}{"reference": "Patient/other-1"},
		},
	})
	count := rw.RewriteReferences("src-1", "tgt-1")
	if count != 0 {
		t.Errorf("expected 0 references rewritten, got %d", count)
	}
}

func TestRewriter_MultipleResourcesRewritten(t *testing.T) {
	rw := NewReferenceRewriter()
	rw.AddResources("Observation", []map[string]interface{}{
		{"resourceType": "Observation", "id": "obs-1", "subject": map[string]interface{}{"reference": "Patient/src-1"}},
		{"resourceType": "Observation", "id": "obs-2", "subject": map[string]interface{}{"reference": "Patient/src-1"}},
		{"resourceType": "Observation", "id": "obs-3", "subject": map[string]interface{}{"reference": "Patient/other"}},
	})
	rw.AddResources("Condition", []map[string]interface{}{
		{"resourceType": "Condition", "id": "cond-1", "subject": map[string]interface{}{"reference": "Patient/src-1"}},
	})
	count := rw.RewriteReferences("src-1", "tgt-1")
	if count != 3 {
		t.Errorf("expected 3 references rewritten, got %d", count)
	}
}

func TestRewriter_OnlyMatchingReferencesChanged(t *testing.T) {
	rw := NewReferenceRewriter()
	rw.AddResources("Observation", []map[string]interface{}{
		{
			"resourceType": "Observation",
			"id":           "obs-1",
			"subject":      map[string]interface{}{"reference": "Patient/src-1"},
			"performer":    []interface{}{map[string]interface{}{"reference": "Practitioner/doc-1"}},
		},
	})
	count := rw.RewriteReferences("src-1", "tgt-1")
	if count != 1 {
		t.Errorf("expected 1 reference rewritten, got %d", count)
	}
	obs := rw.resourceStore["Observation"][0]
	// Performer should not be changed.
	perfs, _ := obs["performer"].([]interface{})
	perf, _ := perfs[0].(map[string]interface{})
	if perf["reference"] != "Practitioner/doc-1" {
		t.Errorf("expected performer unchanged, got %v", perf["reference"])
	}
}

func TestRewriter_ConcurrentSafety(t *testing.T) {
	rw := NewReferenceRewriter()
	// Add a bunch of resources concurrently.
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			rw.AddResources("Observation", []map[string]interface{}{
				{
					"resourceType": "Observation",
					"id":           "obs-concurrent",
					"subject":      map[string]interface{}{"reference": "Patient/src-1"},
				},
			})
		}(i)
	}
	wg.Wait()
	// Now rewrite concurrently.
	var wg2 sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			rw.RewriteReferences("src-1", "tgt-1")
		}()
	}
	wg2.Wait()
	// No panic is a pass.
}

func TestRewriter_StringPatternMatchAnywhere(t *testing.T) {
	rw := NewReferenceRewriter()
	rw.AddResources("DocumentReference", []map[string]interface{}{
		{
			"resourceType": "DocumentReference",
			"id":           "doc-1",
			"context": map[string]interface{}{
				"related": []interface{}{
					map[string]interface{}{"reference": "Patient/src-1"},
				},
			},
		},
	})
	count := rw.RewriteReferences("src-1", "tgt-1")
	if count != 1 {
		t.Errorf("expected 1 reference rewritten via deep scan, got %d", count)
	}
}

func TestRewriter_EmptyStore(t *testing.T) {
	rw := NewReferenceRewriter()
	count := rw.RewriteReferences("src-1", "tgt-1")
	if count != 0 {
		t.Errorf("expected 0 references on empty store, got %d", count)
	}
}

// =============================================================================
// MDM Service Tests (12+)
// =============================================================================

func TestMDMService_MergeTwoPatients_CreatesLink(t *testing.T) {
	mdm := NewMDMService()
	req := MergeRequest{
		SourcePatient: map[string]interface{}{"resourceType": "Patient", "id": "dup-1", "gender": "male"},
		TargetPatient: map[string]interface{}{"resourceType": "Patient", "id": "main-1", "gender": "male"},
	}
	result, err := mdm.Merge(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LinksCreated) != 1 {
		t.Fatalf("expected 1 link created, got %d", len(result.LinksCreated))
	}
	link := result.LinksCreated[0]
	if link.SourceID != "dup-1" {
		t.Errorf("expected source ID 'dup-1', got %s", link.SourceID)
	}
	if link.TargetID != "main-1" {
		t.Errorf("expected target ID 'main-1', got %s", link.TargetID)
	}
	if link.Type != "replaced-by" {
		t.Errorf("expected link type 'replaced-by', got %s", link.Type)
	}
}

func TestMDMService_MergeResult_CorrectSurvivingPatient(t *testing.T) {
	mdm := NewMDMService()
	req := MergeRequest{
		SourcePatient: map[string]interface{}{
			"resourceType": "Patient", "id": "dup-1", "gender": "female",
			"birthDate": "1990-01-01",
		},
		TargetPatient: map[string]interface{}{
			"resourceType": "Patient", "id": "main-1", "gender": "male",
		},
	}
	result, err := mdm.Merge(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The result patient should have the target's ID.
	if result.Result["id"] != "main-1" {
		t.Errorf("expected result patient id 'main-1', got %v", result.Result["id"])
	}
	// Target's gender should win.
	if result.Result["gender"] != "male" {
		t.Errorf("expected gender 'male', got %v", result.Result["gender"])
	}
	// Source's birthDate should be kept since target is missing.
	if result.Result["birthDate"] != "1990-01-01" {
		t.Errorf("expected birthDate '1990-01-01', got %v", result.Result["birthDate"])
	}
}

func TestMDMService_PreviewMode_DoesNotCreateLink(t *testing.T) {
	mdm := NewMDMService()
	req := MergeRequest{
		SourcePatient: map[string]interface{}{"resourceType": "Patient", "id": "dup-1"},
		TargetPatient: map[string]interface{}{"resourceType": "Patient", "id": "main-1"},
		PreviewOnly:   true,
	}
	result, err := mdm.Preview(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Result == nil {
		t.Error("expected a result patient in preview")
	}
	// No link should be persisted.
	links := mdm.GetLinks("dup-1")
	if len(links) != 0 {
		t.Errorf("expected no links in preview mode, got %d", len(links))
	}
}

func TestMDMService_GetLinksForPatient(t *testing.T) {
	mdm := NewMDMService()
	// Merge to create a link.
	req := MergeRequest{
		SourcePatient: map[string]interface{}{"resourceType": "Patient", "id": "dup-1"},
		TargetPatient: map[string]interface{}{"resourceType": "Patient", "id": "main-1"},
	}
	_, _ = mdm.Merge(context.Background(), req)

	// Both source and target should return the link.
	srcLinks := mdm.GetLinks("dup-1")
	if len(srcLinks) != 1 {
		t.Errorf("expected 1 link for source, got %d", len(srcLinks))
	}
	tgtLinks := mdm.GetLinks("main-1")
	if len(tgtLinks) != 1 {
		t.Errorf("expected 1 link for target, got %d", len(tgtLinks))
	}
}

func TestMDMService_GetGoldenRecord_DirectLink(t *testing.T) {
	mdm := NewMDMService()
	req := MergeRequest{
		SourcePatient: map[string]interface{}{"resourceType": "Patient", "id": "dup-1"},
		TargetPatient: map[string]interface{}{"resourceType": "Patient", "id": "main-1"},
	}
	_, _ = mdm.Merge(context.Background(), req)

	golden, err := mdm.GetGoldenRecord("dup-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if golden != "main-1" {
		t.Errorf("expected golden record 'main-1', got %s", golden)
	}
}

func TestMDMService_GetGoldenRecord_Chain(t *testing.T) {
	mdm := NewMDMService()
	// A -> B
	_, _ = mdm.Merge(context.Background(), MergeRequest{
		SourcePatient: map[string]interface{}{"resourceType": "Patient", "id": "A"},
		TargetPatient: map[string]interface{}{"resourceType": "Patient", "id": "B"},
	})
	// B -> C
	_, _ = mdm.Merge(context.Background(), MergeRequest{
		SourcePatient: map[string]interface{}{"resourceType": "Patient", "id": "B"},
		TargetPatient: map[string]interface{}{"resourceType": "Patient", "id": "C"},
	})

	golden, err := mdm.GetGoldenRecord("A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if golden != "C" {
		t.Errorf("expected golden record 'C' through chain A->B->C, got %s", golden)
	}
}

func TestMDMService_Unlink(t *testing.T) {
	mdm := NewMDMService()
	req := MergeRequest{
		SourcePatient: map[string]interface{}{"resourceType": "Patient", "id": "dup-1"},
		TargetPatient: map[string]interface{}{"resourceType": "Patient", "id": "main-1"},
	}
	result, _ := mdm.Merge(context.Background(), req)
	linkID := result.LinksCreated[0].ID

	err := mdm.Unlink(linkID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	links := mdm.GetLinks("dup-1")
	if len(links) != 0 {
		t.Errorf("expected 0 links after unlink, got %d", len(links))
	}
}

func TestMDMService_Unlink_UnknownID(t *testing.T) {
	mdm := NewMDMService()
	err := mdm.Unlink("nonexistent-link-id")
	if err == nil {
		t.Error("expected error for unknown link ID")
	}
}

func TestMDMService_Merge_NilSource(t *testing.T) {
	mdm := NewMDMService()
	_, err := mdm.Merge(context.Background(), MergeRequest{
		SourcePatient: nil,
		TargetPatient: map[string]interface{}{"resourceType": "Patient", "id": "main-1"},
	})
	if err == nil {
		t.Error("expected error for nil source")
	}
}

func TestMDMService_Merge_NilTarget(t *testing.T) {
	mdm := NewMDMService()
	_, err := mdm.Merge(context.Background(), MergeRequest{
		SourcePatient: map[string]interface{}{"resourceType": "Patient", "id": "dup-1"},
		TargetPatient: nil,
	})
	if err == nil {
		t.Error("expected error for nil target")
	}
}

func TestMDMService_Merge_SamePatient(t *testing.T) {
	mdm := NewMDMService()
	_, err := mdm.Merge(context.Background(), MergeRequest{
		SourcePatient: map[string]interface{}{"resourceType": "Patient", "id": "same-1"},
		TargetPatient: map[string]interface{}{"resourceType": "Patient", "id": "same-1"},
	})
	if err == nil {
		t.Error("expected error for merging same patient")
	}
}

func TestMDMService_ApplySurvivorshipRules_Complete(t *testing.T) {
	mdm := NewMDMService()
	source := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "src-1",
		"gender":       "female",
		"birthDate":    "1990-01-01",
		"identifier": []interface{}{
			map[string]interface{}{"system": "http://hospital.org/mrn", "value": "MRN-SRC"},
			map[string]interface{}{"system": "http://ssn", "value": "111-22-3333"},
		},
		"name": []interface{}{
			map[string]interface{}{"use": "maiden", "family": "Johnson"},
		},
		"telecom": []interface{}{
			map[string]interface{}{"system": "phone", "value": "555-1111"},
		},
		"address": []interface{}{
			map[string]interface{}{"use": "old", "city": "OldTown"},
		},
		"maritalStatus": map[string]interface{}{"coding": []interface{}{map[string]interface{}{"code": "S"}}},
		"communication": []interface{}{
			map[string]interface{}{"language": map[string]interface{}{"coding": []interface{}{map[string]interface{}{"code": "es"}}}},
		},
		"generalPractitioner":  []interface{}{map[string]interface{}{"reference": "Practitioner/src-doc"}},
		"managingOrganization": map[string]interface{}{"reference": "Organization/src-org"},
	}
	target := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "tgt-1",
		"gender":       "male",
		"birthDate":    "1990-06-15",
		"identifier": []interface{}{
			map[string]interface{}{"system": "http://hospital.org/mrn", "value": "MRN-TGT"},
		},
		"name": []interface{}{
			map[string]interface{}{"use": "official", "family": "Smith"},
		},
		"telecom": []interface{}{
			map[string]interface{}{"system": "email", "value": "smith@example.com"},
		},
		"address": []interface{}{
			map[string]interface{}{"use": "home", "city": "NewTown"},
		},
		"maritalStatus": map[string]interface{}{"coding": []interface{}{map[string]interface{}{"code": "M"}}},
		"communication": []interface{}{
			map[string]interface{}{"language": map[string]interface{}{"coding": []interface{}{map[string]interface{}{"code": "en"}}}},
		},
		"generalPractitioner":  []interface{}{map[string]interface{}{"reference": "Practitioner/tgt-doc"}},
		"managingOrganization": map[string]interface{}{"reference": "Organization/tgt-org"},
	}

	result := mdm.ApplySurvivorshipRules(source, target)

	// target-wins
	if result["gender"] != "male" {
		t.Errorf("gender: expected 'male', got %v", result["gender"])
	}
	if result["birthDate"] != "1990-06-15" {
		t.Errorf("birthDate: expected '1990-06-15', got %v", result["birthDate"])
	}

	// merge-lists
	identifiers, _ := result["identifier"].([]interface{})
	if len(identifiers) != 2 {
		t.Errorf("identifiers: expected 2, got %d", len(identifiers))
	}
	names, _ := result["name"].([]interface{})
	if len(names) != 2 {
		t.Errorf("names: expected 2, got %d", len(names))
	}
	telecoms, _ := result["telecom"].([]interface{})
	if len(telecoms) != 2 {
		t.Errorf("telecoms: expected 2, got %d", len(telecoms))
	}
	addrs, _ := result["address"].([]interface{})
	if len(addrs) != 2 {
		t.Errorf("addresses: expected 2, got %d", len(addrs))
	}
	comms, _ := result["communication"].([]interface{})
	if len(comms) != 2 {
		t.Errorf("communication: expected 2, got %d", len(comms))
	}

	// target-wins for single fields
	org, _ := result["managingOrganization"].(map[string]interface{})
	if org["reference"] != "Organization/tgt-org" {
		t.Errorf("managingOrganization: expected target, got %v", org["reference"])
	}
	gp, _ := result["generalPractitioner"].([]interface{})
	gpMap, _ := gp[0].(map[string]interface{})
	if gpMap["reference"] != "Practitioner/tgt-doc" {
		t.Errorf("generalPractitioner: expected target, got %v", gpMap["reference"])
	}

	// result ID should be target
	if result["id"] != "tgt-1" {
		t.Errorf("result ID should be target 'tgt-1', got %v", result["id"])
	}
}

func TestMDMService_GetGoldenRecord_NoLink(t *testing.T) {
	mdm := NewMDMService()
	golden, err := mdm.GetGoldenRecord("standalone-patient")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A patient with no links is its own golden record.
	if golden != "standalone-patient" {
		t.Errorf("expected patient to be its own golden record, got %s", golden)
	}
}

// =============================================================================
// Handler Tests (10+)
// =============================================================================

func newMergeTestContext(method, path, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/fhir+json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestMergeHandler_PostMerge_Valid(t *testing.T) {
	mdm := NewMDMService()
	handler := NewMergeHandler(mdm)

	params := map[string]interface{}{
		"resourceType": "Parameters",
		"parameter": []interface{}{
			map[string]interface{}{
				"name":     "source-patient",
				"resource": map[string]interface{}{"resourceType": "Patient", "id": "dup-1", "gender": "male"},
			},
			map[string]interface{}{
				"name":     "target-patient",
				"resource": map[string]interface{}{"resourceType": "Patient", "id": "main-1", "gender": "female"},
			},
		},
	}
	body, _ := json.Marshal(params)
	c, rec := newMergeTestContext(http.MethodPost, "/fhir/Patient/$merge", string(body))

	err := handler.HandleMerge(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["resourceType"] != "Parameters" {
		t.Errorf("expected resourceType Parameters, got %v", resp["resourceType"])
	}
}

func TestMergeHandler_PostMerge_PreviewMode(t *testing.T) {
	mdm := NewMDMService()
	handler := NewMergeHandler(mdm)

	params := map[string]interface{}{
		"resourceType": "Parameters",
		"parameter": []interface{}{
			map[string]interface{}{
				"name":     "source-patient",
				"resource": map[string]interface{}{"resourceType": "Patient", "id": "dup-1"},
			},
			map[string]interface{}{
				"name":     "target-patient",
				"resource": map[string]interface{}{"resourceType": "Patient", "id": "main-1"},
			},
			map[string]interface{}{
				"name":         "preview",
				"valueBoolean": true,
			},
		},
	}
	body, _ := json.Marshal(params)
	c, rec := newMergeTestContext(http.MethodPost, "/fhir/Patient/$merge", string(body))

	err := handler.HandleMerge(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	// No links should be created.
	links := mdm.GetLinks("dup-1")
	if len(links) != 0 {
		t.Errorf("expected no links in preview mode, got %d", len(links))
	}
}

func TestMergeHandler_PostMerge_MissingSource(t *testing.T) {
	mdm := NewMDMService()
	handler := NewMergeHandler(mdm)

	params := map[string]interface{}{
		"resourceType": "Parameters",
		"parameter": []interface{}{
			map[string]interface{}{
				"name":     "target-patient",
				"resource": map[string]interface{}{"resourceType": "Patient", "id": "main-1"},
			},
		},
	}
	body, _ := json.Marshal(params)
	c, rec := newMergeTestContext(http.MethodPost, "/fhir/Patient/$merge", string(body))

	err := handler.HandleMerge(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMergeHandler_PostMerge_MissingTarget(t *testing.T) {
	mdm := NewMDMService()
	handler := NewMergeHandler(mdm)

	params := map[string]interface{}{
		"resourceType": "Parameters",
		"parameter": []interface{}{
			map[string]interface{}{
				"name":     "source-patient",
				"resource": map[string]interface{}{"resourceType": "Patient", "id": "dup-1"},
			},
		},
	}
	body, _ := json.Marshal(params)
	c, rec := newMergeTestContext(http.MethodPost, "/fhir/Patient/$merge", string(body))

	err := handler.HandleMerge(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMergeHandler_GetPatientLinks(t *testing.T) {
	mdm := NewMDMService()
	handler := NewMergeHandler(mdm)

	// Create a link first.
	_, _ = mdm.Merge(context.Background(), MergeRequest{
		SourcePatient: map[string]interface{}{"resourceType": "Patient", "id": "dup-1"},
		TargetPatient: map[string]interface{}{"resourceType": "Patient", "id": "main-1"},
	})

	c, rec := newMergeTestContext(http.MethodGet, "/fhir/Patient/dup-1/$links", "")
	c.SetParamNames("id")
	c.SetParamValues("dup-1")

	err := handler.HandleGetLinks(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	entries, _ := resp["entry"].([]interface{})
	if len(entries) != 1 {
		t.Errorf("expected 1 link entry, got %d", len(entries))
	}
}

func TestMergeHandler_GetPatientLinks_Empty(t *testing.T) {
	mdm := NewMDMService()
	handler := NewMergeHandler(mdm)

	c, rec := newMergeTestContext(http.MethodGet, "/fhir/Patient/unknown/$links", "")
	c.SetParamNames("id")
	c.SetParamValues("unknown")

	err := handler.HandleGetLinks(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	total, _ := resp["total"].(float64)
	if total != 0 {
		t.Errorf("expected total 0, got %v", total)
	}
}

func TestMergeHandler_GetGoldenRecord(t *testing.T) {
	mdm := NewMDMService()
	handler := NewMergeHandler(mdm)

	_, _ = mdm.Merge(context.Background(), MergeRequest{
		SourcePatient: map[string]interface{}{"resourceType": "Patient", "id": "dup-1"},
		TargetPatient: map[string]interface{}{"resourceType": "Patient", "id": "main-1"},
	})

	c, rec := newMergeTestContext(http.MethodGet, "/fhir/Patient/dup-1/$golden-record", "")
	c.SetParamNames("id")
	c.SetParamValues("dup-1")

	err := handler.HandleGoldenRecord(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	params, _ := resp["parameter"].([]interface{})
	if len(params) == 0 {
		t.Fatal("expected parameters in response")
	}
	param, _ := params[0].(map[string]interface{})
	if param["valueString"] != "main-1" {
		t.Errorf("expected golden record 'main-1', got %v", param["valueString"])
	}
}

func TestMergeHandler_DeleteLink(t *testing.T) {
	mdm := NewMDMService()
	handler := NewMergeHandler(mdm)

	result, _ := mdm.Merge(context.Background(), MergeRequest{
		SourcePatient: map[string]interface{}{"resourceType": "Patient", "id": "dup-1"},
		TargetPatient: map[string]interface{}{"resourceType": "Patient", "id": "main-1"},
	})
	linkID := result.LinksCreated[0].ID

	c, rec := newMergeTestContext(http.MethodDelete, "/fhir/Patient/$link/"+linkID, "")
	c.SetParamNames("id")
	c.SetParamValues(linkID)

	err := handler.HandleDeleteLink(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestMergeHandler_DeleteLink_NotFound(t *testing.T) {
	mdm := NewMDMService()
	handler := NewMergeHandler(mdm)

	c, rec := newMergeTestContext(http.MethodDelete, "/fhir/Patient/$link/nonexistent", "")
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := handler.HandleDeleteLink(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestMergeHandler_ListAllLinks(t *testing.T) {
	mdm := NewMDMService()
	handler := NewMergeHandler(mdm)

	// Create two merges.
	_, _ = mdm.Merge(context.Background(), MergeRequest{
		SourcePatient: map[string]interface{}{"resourceType": "Patient", "id": "dup-1"},
		TargetPatient: map[string]interface{}{"resourceType": "Patient", "id": "main-1"},
	})
	_, _ = mdm.Merge(context.Background(), MergeRequest{
		SourcePatient: map[string]interface{}{"resourceType": "Patient", "id": "dup-2"},
		TargetPatient: map[string]interface{}{"resourceType": "Patient", "id": "main-2"},
	})

	c, rec := newMergeTestContext(http.MethodGet, "/fhir/Patient/$links", "")

	err := handler.HandleListAllLinks(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	entries, _ := resp["entry"].([]interface{})
	if len(entries) != 2 {
		t.Errorf("expected 2 link entries, got %d", len(entries))
	}
}

// =============================================================================
// Integration Tests (5+)
// =============================================================================

func TestMergeIntegration_FullMerge(t *testing.T) {
	mdm := NewMDMService()
	mdm.rewriter.AddResources("Observation", []map[string]interface{}{
		{"resourceType": "Observation", "id": "obs-1", "subject": map[string]interface{}{"reference": "Patient/dup-1"}},
		{"resourceType": "Observation", "id": "obs-2", "subject": map[string]interface{}{"reference": "Patient/dup-1"}},
	})

	req := MergeRequest{
		SourcePatient: map[string]interface{}{
			"resourceType": "Patient", "id": "dup-1", "gender": "female",
			"identifier": []interface{}{
				map[string]interface{}{"system": "http://ssn", "value": "999-99-9999"},
			},
		},
		TargetPatient: map[string]interface{}{
			"resourceType": "Patient", "id": "main-1", "gender": "male",
			"identifier": []interface{}{
				map[string]interface{}{"system": "http://hospital.org/mrn", "value": "MRN-001"},
			},
		},
	}

	result, err := mdm.Merge(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1. Survivorship applied.
	if result.Result["gender"] != "male" {
		t.Errorf("expected gender 'male', got %v", result.Result["gender"])
	}

	// 2. Link created.
	if len(result.LinksCreated) != 1 {
		t.Errorf("expected 1 link, got %d", len(result.LinksCreated))
	}

	// 3. References rewritten.
	obs := mdm.rewriter.resourceStore["Observation"]
	for _, o := range obs {
		subj, _ := o["subject"].(map[string]interface{})
		if subj["reference"] != "Patient/main-1" {
			t.Errorf("expected rewritten reference Patient/main-1, got %v", subj["reference"])
		}
	}
}

func TestMergeIntegration_MergeThenLookupGoldenRecord(t *testing.T) {
	mdm := NewMDMService()

	_, _ = mdm.Merge(context.Background(), MergeRequest{
		SourcePatient: map[string]interface{}{"resourceType": "Patient", "id": "dup-1"},
		TargetPatient: map[string]interface{}{"resourceType": "Patient", "id": "main-1"},
	})

	golden, err := mdm.GetGoldenRecord("dup-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if golden != "main-1" {
		t.Errorf("expected golden record 'main-1', got %s", golden)
	}
}

func TestMergeIntegration_MergeChain(t *testing.T) {
	mdm := NewMDMService()

	// A into B.
	_, _ = mdm.Merge(context.Background(), MergeRequest{
		SourcePatient: map[string]interface{}{"resourceType": "Patient", "id": "A"},
		TargetPatient: map[string]interface{}{"resourceType": "Patient", "id": "B"},
	})

	// C into B.
	_, _ = mdm.Merge(context.Background(), MergeRequest{
		SourcePatient: map[string]interface{}{"resourceType": "Patient", "id": "C"},
		TargetPatient: map[string]interface{}{"resourceType": "Patient", "id": "B"},
	})

	// A's golden record should be B.
	golden, _ := mdm.GetGoldenRecord("A")
	if golden != "B" {
		t.Errorf("expected golden record 'B' for A, got %s", golden)
	}

	// C's golden record should be B.
	golden, _ = mdm.GetGoldenRecord("C")
	if golden != "B" {
		t.Errorf("expected golden record 'B' for C, got %s", golden)
	}
}

func TestMergeIntegration_PreservesIdentifiersFromBoth(t *testing.T) {
	mdm := NewMDMService()

	req := MergeRequest{
		SourcePatient: map[string]interface{}{
			"resourceType": "Patient", "id": "dup-1",
			"identifier": []interface{}{
				map[string]interface{}{"system": "http://ssn", "value": "111-22-3333"},
				map[string]interface{}{"system": "http://dl", "value": "DL-123"},
			},
		},
		TargetPatient: map[string]interface{}{
			"resourceType": "Patient", "id": "main-1",
			"identifier": []interface{}{
				map[string]interface{}{"system": "http://hospital.org/mrn", "value": "MRN-001"},
			},
		},
	}

	result, err := mdm.Merge(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	identifiers, ok := result.Result["identifier"].([]interface{})
	if !ok {
		t.Fatal("expected identifiers in result")
	}
	// All 3 unique systems should be present.
	if len(identifiers) != 3 {
		t.Errorf("expected 3 identifiers from both patients, got %d", len(identifiers))
	}
}

func TestMergeIntegration_FHIROutputFormat(t *testing.T) {
	mdm := NewMDMService()
	handler := NewMergeHandler(mdm)

	params := map[string]interface{}{
		"resourceType": "Parameters",
		"parameter": []interface{}{
			map[string]interface{}{
				"name":     "source-patient",
				"resource": map[string]interface{}{"resourceType": "Patient", "id": "dup-1", "gender": "male"},
			},
			map[string]interface{}{
				"name":     "target-patient",
				"resource": map[string]interface{}{"resourceType": "Patient", "id": "main-1", "gender": "female"},
			},
		},
	}
	body, _ := json.Marshal(params)
	c, rec := newMergeTestContext(http.MethodPost, "/fhir/Patient/$merge", string(body))

	err := handler.HandleMerge(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	// Validate FHIR Parameters structure.
	if resp["resourceType"] != "Parameters" {
		t.Errorf("expected resourceType 'Parameters', got %v", resp["resourceType"])
	}

	paramList, ok := resp["parameter"].([]interface{})
	if !ok {
		t.Fatal("expected 'parameter' array")
	}

	// Should have outcome, result, and input parameters.
	paramNames := make(map[string]bool)
	for _, p := range paramList {
		pm, _ := p.(map[string]interface{})
		name, _ := pm["name"].(string)
		paramNames[name] = true
	}

	if !paramNames["outcome"] {
		t.Error("expected 'outcome' parameter")
	}
	if !paramNames["result"] {
		t.Error("expected 'result' parameter")
	}
	if !paramNames["input"] {
		t.Error("expected 'input' parameter")
	}
}

func TestMergeHandler_PostMerge_InvalidJSON(t *testing.T) {
	mdm := NewMDMService()
	handler := NewMergeHandler(mdm)

	c, rec := newMergeTestContext(http.MethodPost, "/fhir/Patient/$merge", "not-json")

	err := handler.HandleMerge(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMergeHandler_PostMerge_EmptyBody(t *testing.T) {
	mdm := NewMDMService()
	handler := NewMergeHandler(mdm)

	c, rec := newMergeTestContext(http.MethodPost, "/fhir/Patient/$merge", "")

	err := handler.HandleMerge(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMergeHandler_PostMerge_WrongResourceType(t *testing.T) {
	mdm := NewMDMService()
	handler := NewMergeHandler(mdm)

	params := map[string]interface{}{
		"resourceType": "Bundle",
		"parameter":    []interface{}{},
	}
	body, _ := json.Marshal(params)
	c, rec := newMergeTestContext(http.MethodPost, "/fhir/Patient/$merge", string(body))

	err := handler.HandleMerge(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMergeHandler_RegisterRoutes(t *testing.T) {
	mdm := NewMDMService()
	handler := NewMergeHandler(mdm)

	e := echo.New()
	g := e.Group("/fhir")
	handler.RegisterRoutes(g)

	routes := e.Routes()
	expectedPaths := map[string]bool{
		"/fhir/Patient/$merge":             false,
		"/fhir/Patient/:id/$links":         false,
		"/fhir/Patient/:id/$golden-record": false,
		"/fhir/Patient/$link/:id":          false,
		"/fhir/Patient/$links":             false,
	}
	for _, r := range routes {
		if _, ok := expectedPaths[r.Path]; ok {
			expectedPaths[r.Path] = true
		}
	}
	for path, found := range expectedPaths {
		if !found {
			t.Errorf("expected route %s to be registered", path)
		}
	}
}

func TestMergeIntegration_MergeWithResultPatientOverride(t *testing.T) {
	mdm := NewMDMService()
	req := MergeRequest{
		SourcePatient: map[string]interface{}{
			"resourceType": "Patient", "id": "dup-1", "gender": "male",
		},
		TargetPatient: map[string]interface{}{
			"resourceType": "Patient", "id": "main-1", "gender": "female",
		},
		ResultPatient: map[string]interface{}{
			"gender": "other",
		},
	}

	result, err := mdm.Merge(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The result patient override should take effect.
	if result.Result["gender"] != "other" {
		t.Errorf("expected gender 'other' from result override, got %v", result.Result["gender"])
	}
}

func TestDefaultSurvivorshipRules(t *testing.T) {
	rules := DefaultSurvivorshipRules()
	if len(rules) == 0 {
		t.Fatal("expected non-empty default survivorship rules")
	}

	ruleMap := make(map[string]string)
	for _, r := range rules {
		ruleMap[r.Field] = r.Strategy
	}

	expected := map[string]string{
		"identifier":           "merge-lists",
		"name":                 "merge-lists",
		"telecom":              "merge-lists",
		"address":              "merge-lists",
		"gender":               "target-wins",
		"birthDate":            "target-wins",
		"maritalStatus":        "most-recent",
		"communication":        "merge-lists",
		"generalPractitioner":  "target-wins",
		"managingOrganization": "target-wins",
	}

	for field, strategy := range expected {
		if ruleMap[field] != strategy {
			t.Errorf("rule for %s: expected %s, got %s", field, strategy, ruleMap[field])
		}
	}
}

func TestMDMService_Merge_InputEchoed(t *testing.T) {
	mdm := NewMDMService()
	source := map[string]interface{}{"resourceType": "Patient", "id": "dup-1"}
	target := map[string]interface{}{"resourceType": "Patient", "id": "main-1"}
	result, err := mdm.Merge(context.Background(), MergeRequest{
		SourcePatient: source,
		TargetPatient: target,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Input.Source["id"] != "dup-1" {
		t.Errorf("expected input source id 'dup-1', got %v", result.Input.Source["id"])
	}
	if result.Input.Target["id"] != "main-1" {
		t.Errorf("expected input target id 'main-1', got %v", result.Input.Target["id"])
	}
}
