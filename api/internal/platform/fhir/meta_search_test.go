package fhir

import (
	"testing"
)

// ---------------------------------------------------------------------------
// MetaSearchClause (SQL generation)
// ---------------------------------------------------------------------------

func TestMetaSearchClause_TagSystemAndCode(t *testing.T) {
	clause, args, next := MetaSearchClause("_tag", "http://example.org|important", "resource_json", 1)

	wantClause := "EXISTS (SELECT 1 FROM jsonb_array_elements(resource_json->'meta'->'tag') elem WHERE elem->>'system' = $1 AND elem->>'code' = $2)"
	if clause != wantClause {
		t.Errorf("clause =\n  %s\nwant\n  %s", clause, wantClause)
	}
	if len(args) != 2 {
		t.Fatalf("args len = %d, want 2", len(args))
	}
	if args[0] != "http://example.org" {
		t.Errorf("args[0] = %v, want http://example.org", args[0])
	}
	if args[1] != "important" {
		t.Errorf("args[1] = %v, want important", args[1])
	}
	if next != 3 {
		t.Errorf("next = %d, want 3", next)
	}
}

func TestMetaSearchClause_TagSystemOnly(t *testing.T) {
	clause, args, next := MetaSearchClause("_tag", "http://example.org|", "resource_json", 5)

	wantClause := "EXISTS (SELECT 1 FROM jsonb_array_elements(resource_json->'meta'->'tag') elem WHERE elem->>'system' = $5)"
	if clause != wantClause {
		t.Errorf("clause =\n  %s\nwant\n  %s", clause, wantClause)
	}
	if len(args) != 1 || args[0] != "http://example.org" {
		t.Errorf("args = %v, want [http://example.org]", args)
	}
	if next != 6 {
		t.Errorf("next = %d, want 6", next)
	}
}

func TestMetaSearchClause_TagCodeOnly(t *testing.T) {
	clause, args, next := MetaSearchClause("_tag", "|important", "resource_json", 1)

	wantClause := "EXISTS (SELECT 1 FROM jsonb_array_elements(resource_json->'meta'->'tag') elem WHERE elem->>'code' = $1)"
	if clause != wantClause {
		t.Errorf("clause =\n  %s\nwant\n  %s", clause, wantClause)
	}
	if len(args) != 1 || args[0] != "important" {
		t.Errorf("args = %v, want [important]", args)
	}
	if next != 2 {
		t.Errorf("next = %d, want 2", next)
	}
}

func TestMetaSearchClause_TagCodeNoPipe(t *testing.T) {
	clause, args, next := MetaSearchClause("_tag", "important", "resource_json", 1)

	wantClause := "EXISTS (SELECT 1 FROM jsonb_array_elements(resource_json->'meta'->'tag') elem WHERE elem->>'code' = $1)"
	if clause != wantClause {
		t.Errorf("clause =\n  %s\nwant\n  %s", clause, wantClause)
	}
	if len(args) != 1 || args[0] != "important" {
		t.Errorf("args = %v, want [important]", args)
	}
	if next != 2 {
		t.Errorf("next = %d, want 2", next)
	}
}

func TestMetaSearchClause_Security(t *testing.T) {
	clause, args, next := MetaSearchClause("_security", "http://terminology.hl7.org/CodeSystem/v3-Confidentiality|R", "meta_json", 3)

	wantClause := "EXISTS (SELECT 1 FROM jsonb_array_elements(meta_json->'meta'->'security') elem WHERE elem->>'system' = $3 AND elem->>'code' = $4)"
	if clause != wantClause {
		t.Errorf("clause =\n  %s\nwant\n  %s", clause, wantClause)
	}
	if len(args) != 2 {
		t.Fatalf("args len = %d, want 2", len(args))
	}
	if args[0] != "http://terminology.hl7.org/CodeSystem/v3-Confidentiality" {
		t.Errorf("args[0] = %v", args[0])
	}
	if args[1] != "R" {
		t.Errorf("args[1] = %v, want R", args[1])
	}
	if next != 5 {
		t.Errorf("next = %d, want 5", next)
	}
}

func TestMetaSearchClause_Profile(t *testing.T) {
	clause, args, next := MetaSearchClause("_profile", "http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient", "resource_json", 1)

	wantClause := "resource_json->'meta'->'profile' ? $1"
	if clause != wantClause {
		t.Errorf("clause =\n  %s\nwant\n  %s", clause, wantClause)
	}
	if len(args) != 1 || args[0] != "http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient" {
		t.Errorf("args = %v", args)
	}
	if next != 2 {
		t.Errorf("next = %d, want 2", next)
	}
}

func TestMetaSearchClause_UnknownParam(t *testing.T) {
	clause, args, next := MetaSearchClause("_unknown", "val", "resource_json", 1)

	if clause != "TRUE" {
		t.Errorf("clause = %s, want TRUE", clause)
	}
	if len(args) != 0 {
		t.Errorf("args = %v, want empty", args)
	}
	if next != 1 {
		t.Errorf("next = %d, want 1", next)
	}
}

// ---------------------------------------------------------------------------
// IsMetaSearchParam
// ---------------------------------------------------------------------------

func TestIsMetaSearchParam(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"_tag", true},
		{"_security", true},
		{"_profile", true},
		{"_count", false},
		{"_sort", false},
		{"tag", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMetaSearchParam(tt.name); got != tt.want {
				t.Errorf("IsMetaSearchParam(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AddMetaSearchSQL
// ---------------------------------------------------------------------------

func TestAddMetaSearchSQL(t *testing.T) {
	qb := NewSearchQuery("patient", "id, fhir_id")
	params := map[string]string{
		"_tag":     "http://example.org|priority",
		"_profile": "http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient",
	}

	AddMetaSearchSQL(qb, params, "resource_json")

	// The query builder should have accumulated args for both meta params.
	sql := qb.CountSQL()
	if sql == "" {
		t.Fatal("CountSQL() returned empty string")
	}
	args := qb.CountArgs()
	// _tag generates 2 args (system + code), _profile generates 1 arg.
	if len(args) != 3 {
		t.Errorf("CountArgs() len = %d, want 3; args = %v", len(args), args)
	}
}

func TestAddMetaSearchSQL_NoParams(t *testing.T) {
	qb := NewSearchQuery("patient", "id, fhir_id")
	params := map[string]string{
		"status": "active",
	}

	AddMetaSearchSQL(qb, params, "resource_json")

	args := qb.CountArgs()
	if len(args) != 0 {
		t.Errorf("expected no meta args, got %v", args)
	}
}

// ---------------------------------------------------------------------------
// NewMetaSearchFilter
// ---------------------------------------------------------------------------

func TestNewMetaSearchFilter_Nil(t *testing.T) {
	f := NewMetaSearchFilter(map[string]string{"status": "active"})
	if f != nil {
		t.Error("expected nil filter when no meta params present")
	}
}

func TestNewMetaSearchFilter_Tag(t *testing.T) {
	f := NewMetaSearchFilter(map[string]string{
		"_tag": "http://example.org|important",
	})
	if f == nil {
		t.Fatal("expected non-nil filter")
	}
	if len(f.Tags) != 1 {
		t.Fatalf("Tags len = %d, want 1", len(f.Tags))
	}
	if f.Tags[0].System != "http://example.org" || f.Tags[0].Code != "important" {
		t.Errorf("Tags[0] = %+v", f.Tags[0])
	}
}

func TestNewMetaSearchFilter_MultipleTags(t *testing.T) {
	f := NewMetaSearchFilter(map[string]string{
		"_tag": "http://a.org|x,http://b.org|y",
	})
	if f == nil {
		t.Fatal("expected non-nil filter")
	}
	if len(f.Tags) != 2 {
		t.Fatalf("Tags len = %d, want 2", len(f.Tags))
	}
}

func TestNewMetaSearchFilter_Security(t *testing.T) {
	f := NewMetaSearchFilter(map[string]string{
		"_security": "http://terminology.hl7.org/CodeSystem/v3-Confidentiality|R",
	})
	if f == nil {
		t.Fatal("expected non-nil filter")
	}
	if len(f.SecurityLabels) != 1 {
		t.Fatalf("SecurityLabels len = %d, want 1", len(f.SecurityLabels))
	}
	if f.SecurityLabels[0].Code != "R" {
		t.Errorf("SecurityLabels[0].Code = %s, want R", f.SecurityLabels[0].Code)
	}
}

func TestNewMetaSearchFilter_Profile(t *testing.T) {
	f := NewMetaSearchFilter(map[string]string{
		"_profile": "http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient",
	})
	if f == nil {
		t.Fatal("expected non-nil filter")
	}
	if len(f.Profiles) != 1 {
		t.Fatalf("Profiles len = %d, want 1", len(f.Profiles))
	}
	if f.Profiles[0] != "http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient" {
		t.Errorf("Profiles[0] = %s", f.Profiles[0])
	}
}

// ---------------------------------------------------------------------------
// MetaSearchFilter.Match (in-memory filtering)
// ---------------------------------------------------------------------------

func TestMetaSearchFilter_Match_NilFilter(t *testing.T) {
	var f *MetaSearchFilter
	resource := map[string]interface{}{"resourceType": "Patient"}
	if !f.Match(resource) {
		t.Error("nil filter should always match")
	}
}

func TestMetaSearchFilter_Match_TagWithMetaStruct(t *testing.T) {
	f := &MetaSearchFilter{
		Tags: []CodingMatch{{System: "http://example.org", Code: "important"}},
	}

	matching := map[string]interface{}{
		"resourceType": "Patient",
		"meta": Meta{
			Tag: []Coding{{System: "http://example.org", Code: "important"}},
		},
	}
	if !f.Match(matching) {
		t.Error("expected match for resource with matching tag")
	}

	nonMatching := map[string]interface{}{
		"resourceType": "Patient",
		"meta": Meta{
			Tag: []Coding{{System: "http://other.org", Code: "other"}},
		},
	}
	if f.Match(nonMatching) {
		t.Error("expected no match for resource without matching tag")
	}
}

func TestMetaSearchFilter_Match_TagWithMetaPointer(t *testing.T) {
	f := &MetaSearchFilter{
		Tags: []CodingMatch{{System: "http://example.org", Code: "priority"}},
	}

	resource := map[string]interface{}{
		"resourceType": "Condition",
		"meta": &Meta{
			Tag: []Coding{{System: "http://example.org", Code: "priority"}},
		},
	}
	if !f.Match(resource) {
		t.Error("expected match for *Meta with matching tag")
	}
}

func TestMetaSearchFilter_Match_TagWithMapMeta(t *testing.T) {
	f := &MetaSearchFilter{
		Tags: []CodingMatch{{System: "http://example.org", Code: "review"}},
	}

	resource := map[string]interface{}{
		"resourceType": "Observation",
		"meta": map[string]interface{}{
			"tag": []interface{}{
				map[string]interface{}{"system": "http://example.org", "code": "review"},
			},
		},
	}
	if !f.Match(resource) {
		t.Error("expected match for map meta with matching tag")
	}
}

func TestMetaSearchFilter_Match_SecurityLabel(t *testing.T) {
	f := &MetaSearchFilter{
		SecurityLabels: []CodingMatch{{System: "http://terminology.hl7.org/CodeSystem/v3-Confidentiality", Code: "R"}},
	}

	matching := map[string]interface{}{
		"resourceType": "Patient",
		"meta": Meta{
			Security: []Coding{{
				System: "http://terminology.hl7.org/CodeSystem/v3-Confidentiality",
				Code:   "R",
			}},
		},
	}
	if !f.Match(matching) {
		t.Error("expected match for resource with matching security label")
	}

	noSecurity := map[string]interface{}{
		"resourceType": "Patient",
		"meta":         Meta{},
	}
	if f.Match(noSecurity) {
		t.Error("expected no match for resource without security labels")
	}
}

func TestMetaSearchFilter_Match_Profile(t *testing.T) {
	profileURI := "http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient"
	f := &MetaSearchFilter{
		Profiles: []string{profileURI},
	}

	matching := map[string]interface{}{
		"resourceType": "Patient",
		"meta": Meta{
			Profile: []string{profileURI},
		},
	}
	if !f.Match(matching) {
		t.Error("expected match for resource with matching profile")
	}

	noProfile := map[string]interface{}{
		"resourceType": "Patient",
		"meta":         Meta{},
	}
	if f.Match(noProfile) {
		t.Error("expected no match for resource without profiles")
	}
}

func TestMetaSearchFilter_Match_ProfileMapMeta(t *testing.T) {
	profileURI := "http://hl7.org/fhir/us/core/StructureDefinition/us-core-condition"
	f := &MetaSearchFilter{
		Profiles: []string{profileURI},
	}

	resource := map[string]interface{}{
		"resourceType": "Condition",
		"meta": map[string]interface{}{
			"profile": []interface{}{profileURI},
		},
	}
	if !f.Match(resource) {
		t.Error("expected match for map meta with matching profile")
	}
}

func TestMetaSearchFilter_Match_ANDSemantics(t *testing.T) {
	// When both _tag and _profile are specified, the resource must match both.
	f := &MetaSearchFilter{
		Tags:     []CodingMatch{{System: "http://example.org", Code: "important"}},
		Profiles: []string{"http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient"},
	}

	// Matches both
	both := map[string]interface{}{
		"resourceType": "Patient",
		"meta": Meta{
			Tag:     []Coding{{System: "http://example.org", Code: "important"}},
			Profile: []string{"http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient"},
		},
	}
	if !f.Match(both) {
		t.Error("expected match when resource satisfies all criteria")
	}

	// Matches tag but not profile
	tagOnly := map[string]interface{}{
		"resourceType": "Patient",
		"meta": Meta{
			Tag: []Coding{{System: "http://example.org", Code: "important"}},
		},
	}
	if f.Match(tagOnly) {
		t.Error("expected no match when profile is missing")
	}

	// Matches profile but not tag
	profileOnly := map[string]interface{}{
		"resourceType": "Patient",
		"meta": Meta{
			Profile: []string{"http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient"},
		},
	}
	if f.Match(profileOnly) {
		t.Error("expected no match when tag is missing")
	}
}

func TestMetaSearchFilter_Match_ORWithinTags(t *testing.T) {
	// Multiple tags in a comma-separated list have OR semantics:
	// resource must match at least one.
	f := &MetaSearchFilter{
		Tags: []CodingMatch{
			{System: "http://a.org", Code: "x"},
			{System: "http://b.org", Code: "y"},
		},
	}

	matchFirst := map[string]interface{}{
		"resourceType": "Patient",
		"meta": Meta{
			Tag: []Coding{{System: "http://a.org", Code: "x"}},
		},
	}
	if !f.Match(matchFirst) {
		t.Error("expected match when resource has first tag")
	}

	matchSecond := map[string]interface{}{
		"resourceType": "Patient",
		"meta": Meta{
			Tag: []Coding{{System: "http://b.org", Code: "y"}},
		},
	}
	if !f.Match(matchSecond) {
		t.Error("expected match when resource has second tag")
	}

	matchNeither := map[string]interface{}{
		"resourceType": "Patient",
		"meta": Meta{
			Tag: []Coding{{System: "http://c.org", Code: "z"}},
		},
	}
	if f.Match(matchNeither) {
		t.Error("expected no match when resource has neither tag")
	}
}

func TestMetaSearchFilter_Match_NoMeta(t *testing.T) {
	f := &MetaSearchFilter{
		Tags: []CodingMatch{{System: "http://example.org", Code: "x"}},
	}

	resource := map[string]interface{}{
		"resourceType": "Patient",
	}
	if f.Match(resource) {
		t.Error("expected no match for resource without meta")
	}
}

func TestMetaSearchFilter_Match_NilResource(t *testing.T) {
	f := &MetaSearchFilter{
		Tags: []CodingMatch{{Code: "x"}},
	}
	if f.Match(nil) {
		t.Error("expected no match for nil resource")
	}
}

func TestMetaSearchFilter_Match_TagCodeOnlyNoPipe(t *testing.T) {
	f := &MetaSearchFilter{
		Tags: []CodingMatch{{Code: "urgent"}},
	}

	resource := map[string]interface{}{
		"resourceType": "Patient",
		"meta": Meta{
			Tag: []Coding{
				{System: "http://any-system.org", Code: "urgent"},
			},
		},
	}
	if !f.Match(resource) {
		t.Error("expected match when code-only search matches any system")
	}
}

func TestMetaSearchFilter_Match_SystemOnlyPipe(t *testing.T) {
	f := &MetaSearchFilter{
		Tags: []CodingMatch{{System: "http://example.org"}},
	}

	resource := map[string]interface{}{
		"resourceType": "Patient",
		"meta": Meta{
			Tag: []Coding{
				{System: "http://example.org", Code: "anything"},
			},
		},
	}
	if !f.Match(resource) {
		t.Error("expected match when system-only search matches any code")
	}
}

// ---------------------------------------------------------------------------
// FilterResourceList
// ---------------------------------------------------------------------------

func TestFilterResourceList_NilFilter(t *testing.T) {
	resources := []map[string]interface{}{
		{"resourceType": "Patient", "id": "1"},
		{"resourceType": "Patient", "id": "2"},
	}
	result := FilterResourceList(resources, nil)
	if len(result) != 2 {
		t.Errorf("expected 2 resources, got %d", len(result))
	}
}

func TestFilterResourceList_FiltersCorrectly(t *testing.T) {
	f := &MetaSearchFilter{
		Tags: []CodingMatch{{System: "http://example.org", Code: "keep"}},
	}

	resources := []map[string]interface{}{
		{
			"resourceType": "Patient",
			"id":           "1",
			"meta": Meta{
				Tag: []Coding{{System: "http://example.org", Code: "keep"}},
			},
		},
		{
			"resourceType": "Patient",
			"id":           "2",
			"meta":         Meta{},
		},
		{
			"resourceType": "Patient",
			"id":           "3",
			"meta": Meta{
				Tag: []Coding{{System: "http://example.org", Code: "keep"}},
			},
		},
	}

	result := FilterResourceList(resources, f)
	if len(result) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(result))
	}
	if result[0]["id"] != "1" || result[1]["id"] != "3" {
		t.Errorf("unexpected IDs: %v, %v", result[0]["id"], result[1]["id"])
	}
}

func TestFilterResourceList_EmptyInput(t *testing.T) {
	f := &MetaSearchFilter{
		Tags: []CodingMatch{{Code: "x"}},
	}
	result := FilterResourceList(nil, f)
	if len(result) != 0 {
		t.Errorf("expected 0 resources, got %d", len(result))
	}
}

// ---------------------------------------------------------------------------
// parseTokenValue edge cases
// ---------------------------------------------------------------------------

func TestParseTokenValue(t *testing.T) {
	tests := []struct {
		input  string
		system string
		code   string
	}{
		{"http://example.org|abc", "http://example.org", "abc"},
		{"|abc", "", "abc"},
		{"http://example.org|", "http://example.org", ""},
		{"abc", "", "abc"},
		{"|", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cm := parseTokenValue(tt.input)
			if cm.System != tt.system {
				t.Errorf("system = %q, want %q", cm.System, tt.system)
			}
			if cm.Code != tt.code {
				t.Errorf("code = %q, want %q", cm.Code, tt.code)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// codingMatches edge cases
// ---------------------------------------------------------------------------

func TestCodingMatches_EmptyBoth(t *testing.T) {
	// When both system and code are empty, it should not match.
	cm := CodingMatch{}
	c := Coding{System: "http://example.org", Code: "test"}
	if codingMatches(cm, c) {
		t.Error("empty CodingMatch should not match any coding")
	}
}
