package fhir

import (
	"encoding/json"
	"testing"
)

func TestApplyElements(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"meta":         map[string]interface{}{"versionId": "1"},
		"name":         "John",
		"gender":       "male",
		"birthDate":    "1990-01-01",
	}

	result := ApplyElements(resource, "name,gender")

	// Mandatory elements always included
	if result["resourceType"] != "Patient" {
		t.Error("resourceType should always be included")
	}
	if result["id"] != "123" {
		t.Error("id should always be included")
	}
	if result["meta"] == nil {
		t.Error("meta should always be included")
	}

	// Requested elements
	if result["name"] != "John" {
		t.Error("name should be included")
	}
	if result["gender"] != "male" {
		t.Error("gender should be included")
	}

	// Non-requested elements
	if _, ok := result["birthDate"]; ok {
		t.Error("birthDate should not be included")
	}
}

func TestApplyElements_Empty(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"name":         "John",
	}

	result := ApplyElements(resource, "")
	if len(result) != len(resource) {
		t.Error("empty elements string should return all fields")
	}
}

func TestApplySummary_True(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType":  "Patient",
		"id":            "123",
		"meta":          map[string]interface{}{},
		"name":          []interface{}{"John"},
		"gender":        "male",
		"birthDate":     "1990-01-01",
		"text":          map[string]interface{}{"div": "<div>text</div>"},
		"communication": []interface{}{"en"},
	}

	result := ApplySummary(resource, "true")

	if result["resourceType"] != "Patient" {
		t.Error("resourceType should be in summary")
	}
	if result["name"] == nil {
		t.Error("name should be in Patient summary")
	}
	if result["gender"] == nil {
		t.Error("gender should be in Patient summary")
	}
	// communication is NOT in Patient summary elements
	if _, ok := result["communication"]; ok {
		t.Error("communication should not be in Patient summary")
	}
	// Check SUBSETTED tag
	meta, _ := result["meta"].(map[string]interface{})
	if meta == nil {
		t.Fatal("meta should exist")
	}
	tags, _ := meta["tag"].([]interface{})
	if len(tags) == 0 {
		t.Error("should have SUBSETTED tag")
	}
}

func TestApplySummary_Text(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"meta":         map[string]interface{}{},
		"text":         map[string]interface{}{"div": "<div>...</div>"},
		"name":         "John",
	}

	result := ApplySummary(resource, "text")

	if result["id"] != "123" {
		t.Error("id should be included in text mode")
	}
	if result["text"] == nil {
		t.Error("text should be included in text mode")
	}
	if _, ok := result["name"]; ok {
		t.Error("name should not be included in text mode")
	}
}

func TestApplySummary_Data(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"text":         map[string]interface{}{"div": "<div>...</div>"},
		"name":         "John",
	}

	result := ApplySummary(resource, "data")

	if _, ok := result["text"]; ok {
		t.Error("text should be removed in data mode")
	}
	if result["name"] != "John" {
		t.Error("name should be preserved in data mode")
	}
}

func TestApplySummary_False(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"name":         "John",
	}

	result := ApplySummary(resource, "false")
	if len(result) != len(resource) {
		t.Error("false summary should return full resource")
	}
}

func TestApplyProjection_ElementsPrecedence(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"meta":         map[string]interface{}{},
		"name":         "John",
		"gender":       "male",
	}

	// When both are specified, _elements takes precedence
	result := ApplyProjection(resource, "name", "true")
	if _, ok := result["gender"]; ok {
		t.Error("gender should not be included when _elements=name")
	}
}

func TestApplyProjectionToBundle(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "1",
		"meta":         map[string]interface{}{},
		"name":         "John",
		"gender":       "male",
	}
	raw, _ := json.Marshal(resource)

	bundle := &Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Entry: []BundleEntry{
			{Resource: raw},
		},
	}

	ApplyProjectionToBundle(bundle, "name", "")

	var result map[string]interface{}
	json.Unmarshal(bundle.Entry[0].Resource, &result)

	if result["name"] != "John" {
		t.Error("name should be included")
	}
	if _, ok := result["gender"]; ok {
		t.Error("gender should be filtered out")
	}
}

func TestApplyProjection_BothEmpty(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"meta":         map[string]interface{}{},
		"name":         "John",
		"gender":       "male",
	}

	result := ApplyProjection(resource, "", "")

	// Should return original resource unchanged
	if len(result) != len(resource) {
		t.Errorf("expected %d fields, got %d", len(resource), len(result))
	}
	if result["name"] != "John" {
		t.Error("name should be present")
	}
	if result["gender"] != "male" {
		t.Error("gender should be present")
	}
}

func TestApplyProjection_SummaryOnly(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType":  "Patient",
		"id":            "123",
		"meta":          map[string]interface{}{},
		"name":          []interface{}{"John"},
		"gender":        "male",
		"communication": []interface{}{"en"},
	}

	result := ApplyProjection(resource, "", "true")

	// name and gender are in Patient summary; communication is not
	if result["name"] == nil {
		t.Error("name should be in summary result")
	}
	if result["gender"] == nil {
		t.Error("gender should be in summary result")
	}
	if _, ok := result["communication"]; ok {
		t.Error("communication should not be in summary result")
	}
}

func TestApplySummary_UnknownMode(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"name":         "John",
		"gender":       "male",
	}

	result := ApplySummary(resource, "bogus")

	// Unknown mode should return original resource
	if len(result) != len(resource) {
		t.Errorf("expected %d fields, got %d", len(resource), len(result))
	}
	if result["name"] != "John" {
		t.Error("name should be present for unknown summary mode")
	}
}

func TestAddSubsettedTag_NoMeta(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"name":         "John",
	}

	// Calling ApplySummary with "text" mode will call addSubsettedTag.
	// The resource has no "meta" key, so addSubsettedTag should create one.
	result := ApplySummary(resource, "text")

	meta, ok := result["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("meta should be created when missing")
	}
	tags, ok := meta["tag"].([]interface{})
	if !ok || len(tags) == 0 {
		t.Fatal("should have SUBSETTED tag")
	}
	tag := tags[0].(map[string]interface{})
	if tag["code"] != "SUBSETTED" {
		t.Errorf("expected SUBSETTED tag, got %v", tag["code"])
	}
	if tag["system"] != "http://terminology.hl7.org/CodeSystem/v3-ObservationValue" {
		t.Errorf("unexpected system: %v", tag["system"])
	}
}

func TestApplyProjectionToBundle_UnmarshalError(t *testing.T) {
	bundle := &Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Entry: []BundleEntry{
			{Resource: json.RawMessage(`{invalid json}`)},
		},
	}

	// Should not panic; invalid JSON entry is skipped
	ApplyProjectionToBundle(bundle, "name", "")

	// The entry resource should remain unchanged (unmarshal failed, so continue)
	if string(bundle.Entry[0].Resource) != `{invalid json}` {
		t.Errorf("expected entry to remain unchanged, got %s", string(bundle.Entry[0].Resource))
	}
}

func TestApplyProjectionToBundle_EmptyResource(t *testing.T) {
	bundle := &Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Entry: []BundleEntry{
			{Resource: json.RawMessage{}},
		},
	}

	// Should not panic; empty resource is skipped via len check
	ApplyProjectionToBundle(bundle, "name", "")

	if len(bundle.Entry[0].Resource) != 0 {
		t.Errorf("expected empty resource to remain empty, got %s", string(bundle.Entry[0].Resource))
	}
}

func TestApplyProjectionToBundle_SummaryCount(t *testing.T) {
	raw, _ := json.Marshal(map[string]interface{}{"id": "1"})
	total := 5
	bundle := &Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        &total,
		Entry:        []BundleEntry{{Resource: raw}},
	}

	ApplyProjectionToBundle(bundle, "", "count")

	if bundle.Entry != nil {
		t.Error("entries should be nil for _summary=count")
	}
	if *bundle.Total != 5 {
		t.Error("total should be preserved")
	}
}

func TestApplyProjectionToBundle_NoOp(t *testing.T) {
	raw, _ := json.Marshal(map[string]interface{}{"id": "1", "name": "John"})
	bundle := &Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Entry:        []BundleEntry{{Resource: raw}},
	}

	originalResource := string(bundle.Entry[0].Resource)
	ApplyProjectionToBundle(bundle, "", "")

	if string(bundle.Entry[0].Resource) != originalResource {
		t.Error("expected resource to be unchanged when no projection is specified")
	}
}

func TestApplyElements_SpacesInElementList(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "1",
		"meta":         map[string]interface{}{},
		"name":         "John",
		"gender":       "male",
		"birthDate":    "1990-01-01",
	}

	result := ApplyElements(resource, " name , gender ")
	if result["name"] != "John" {
		t.Error("name should be included (spaces trimmed)")
	}
	if result["gender"] != "male" {
		t.Error("gender should be included (spaces trimmed)")
	}
	if _, ok := result["birthDate"]; ok {
		t.Error("birthDate should not be included")
	}
}

func TestApplySummary_TrueUnknownResourceType(t *testing.T) {
	// A resource type not in SummaryElements should use DefaultSummaryElements
	resource := map[string]interface{}{
		"resourceType": "UnknownResource",
		"id":           "1",
		"meta":         map[string]interface{}{},
		"status":       "active",
		"code":         "1234",
		"customField":  "should be excluded",
	}

	result := ApplySummary(resource, "true")
	if result["status"] != "active" {
		t.Error("status should be in default summary")
	}
	if result["code"] != "1234" {
		t.Error("code should be in default summary")
	}
	if _, ok := result["customField"]; ok {
		t.Error("customField should not be in default summary")
	}
}

func TestApplySummary_EmptyString(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"name":         "John",
	}

	result := ApplySummary(resource, "")
	if len(result) != len(resource) {
		t.Error("empty summary mode should return full resource")
	}
}

func TestApplySummary_TextWithNoText(t *testing.T) {
	// Resource has no "text" field
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"meta":         map[string]interface{}{},
		"name":         "John",
	}

	result := ApplySummary(resource, "text")
	if result["id"] != "123" {
		t.Error("id should be included in text mode")
	}
	if _, ok := result["text"]; ok {
		t.Error("text field should not appear if not in original resource")
	}
	if _, ok := result["name"]; ok {
		t.Error("name should not be in text mode")
	}
}

func TestApplyProjectionToBundle_SummaryTrue(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType":  "Patient",
		"id":            "1",
		"meta":          map[string]interface{}{},
		"name":          []interface{}{"John"},
		"communication": []interface{}{"en"},
	}
	raw, _ := json.Marshal(resource)

	bundle := &Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Entry:        []BundleEntry{{Resource: raw}},
	}

	ApplyProjectionToBundle(bundle, "", "true")

	var result map[string]interface{}
	json.Unmarshal(bundle.Entry[0].Resource, &result)

	if result["name"] == nil {
		t.Error("name should be in Patient summary")
	}
	if _, ok := result["communication"]; ok {
		t.Error("communication should not be in Patient summary")
	}
}
