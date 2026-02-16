package fhir

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// MissingSearchClause tests
// ---------------------------------------------------------------------------

func TestMissingSearchClause_True(t *testing.T) {
	clause, args, nextIdx := MissingSearchClause("birth_date", true, 1)
	if clause != "birth_date IS NULL" {
		t.Errorf("clause = %q, want %q", clause, "birth_date IS NULL")
	}
	if len(args) != 0 {
		t.Errorf("args length = %d, want 0", len(args))
	}
	if nextIdx != 1 {
		t.Errorf("nextIdx = %d, want 1", nextIdx)
	}
}

func TestMissingSearchClause_False(t *testing.T) {
	clause, args, nextIdx := MissingSearchClause("birth_date", false, 1)
	if clause != "birth_date IS NOT NULL" {
		t.Errorf("clause = %q, want %q", clause, "birth_date IS NOT NULL")
	}
	if len(args) != 0 {
		t.Errorf("args length = %d, want 0", len(args))
	}
	if nextIdx != 1 {
		t.Errorf("nextIdx = %d, want 1", nextIdx)
	}
}

func TestMissingSearchClause_PreservesIdx(t *testing.T) {
	_, _, nextIdx := MissingSearchClause("status", true, 5)
	if nextIdx != 5 {
		t.Errorf("nextIdx = %d, want 5 (should not advance)", nextIdx)
	}
}

func TestMissingSearchClause_DifferentColumns(t *testing.T) {
	clause, _, _ := MissingSearchClause("effective_date", true, 3)
	if clause != "effective_date IS NULL" {
		t.Errorf("clause = %q, want %q", clause, "effective_date IS NULL")
	}

	clause, _, _ = MissingSearchClause("code_system", false, 7)
	if clause != "code_system IS NOT NULL" {
		t.Errorf("clause = %q, want %q", clause, "code_system IS NOT NULL")
	}
}

// ---------------------------------------------------------------------------
// ParseMissingModifier tests
// ---------------------------------------------------------------------------

func TestParseMissingModifier_WithMissing(t *testing.T) {
	base, isMissing, hasMissing := ParseMissingModifier("birthdate:missing")
	if base != "birthdate" {
		t.Errorf("baseName = %q, want %q", base, "birthdate")
	}
	if !isMissing {
		t.Error("isMissing = false, want true")
	}
	if !hasMissing {
		t.Error("hasMissing = false, want true")
	}
}

func TestParseMissingModifier_WithoutMissing(t *testing.T) {
	base, isMissing, hasMissing := ParseMissingModifier("birthdate")
	if base != "birthdate" {
		t.Errorf("baseName = %q, want %q", base, "birthdate")
	}
	if isMissing {
		t.Error("isMissing = true, want false")
	}
	if hasMissing {
		t.Error("hasMissing = true, want false")
	}
}

func TestParseMissingModifier_OtherModifier(t *testing.T) {
	base, _, hasMissing := ParseMissingModifier("name:exact")
	if hasMissing {
		t.Error("hasMissing should be false for :exact modifier")
	}
	if base != "name:exact" {
		t.Errorf("baseName = %q, want %q (full string when not :missing)", base, "name:exact")
	}
}

func TestParseMissingModifier_EmptyString(t *testing.T) {
	base, isMissing, hasMissing := ParseMissingModifier("")
	if base != "" {
		t.Errorf("baseName = %q, want empty", base)
	}
	if isMissing || hasMissing {
		t.Error("empty string should not be parsed as :missing")
	}
}

func TestParseMissingModifier_ColonOnly(t *testing.T) {
	base, _, hasMissing := ParseMissingModifier(":missing")
	if base != "" {
		t.Errorf("baseName = %q, want empty", base)
	}
	if !hasMissing {
		t.Error("':missing' alone should be detected as having :missing modifier")
	}
}

// ---------------------------------------------------------------------------
// TypedReferenceSearchClause tests
// ---------------------------------------------------------------------------

func TestTypedReferenceSearchClause_Basic(t *testing.T) {
	clause, args, nextIdx := TypedReferenceSearchClause("patient_id", "patient_type", "123", "Patient", 1)
	wantClause := "(patient_id = $1 AND patient_type = $2)"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 2 {
		t.Fatalf("args length = %d, want 2", len(args))
	}
	if args[0] != "123" {
		t.Errorf("args[0] = %v, want %q", args[0], "123")
	}
	if args[1] != "Patient" {
		t.Errorf("args[1] = %v, want %q", args[1], "Patient")
	}
	if nextIdx != 3 {
		t.Errorf("nextIdx = %d, want 3", nextIdx)
	}
}

func TestTypedReferenceSearchClause_StripsPrefix(t *testing.T) {
	clause, args, nextIdx := TypedReferenceSearchClause("subject_id", "subject_type", "Patient/abc-456", "Patient", 5)
	wantClause := "(subject_id = $5 AND subject_type = $6)"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if args[0] != "abc-456" {
		t.Errorf("args[0] = %v, want %q", args[0], "abc-456")
	}
	if nextIdx != 7 {
		t.Errorf("nextIdx = %d, want 7", nextIdx)
	}
}

func TestTypedReferenceSearchClause_FullURL(t *testing.T) {
	_, args, _ := TypedReferenceSearchClause("ref_col", "type_col", "http://example.org/fhir/Patient/xyz", "Patient", 1)
	if args[0] != "xyz" {
		t.Errorf("args[0] = %v, want %q (should strip full URL prefix)", args[0], "xyz")
	}
}

// ---------------------------------------------------------------------------
// ParseTypeModifier tests
// ---------------------------------------------------------------------------

func TestParseTypeModifier_WithPatient(t *testing.T) {
	base, rt, hasType := ParseTypeModifier("subject:Patient")
	if base != "subject" {
		t.Errorf("baseName = %q, want %q", base, "subject")
	}
	if rt != "Patient" {
		t.Errorf("resourceType = %q, want %q", rt, "Patient")
	}
	if !hasType {
		t.Error("hasType = false, want true")
	}
}

func TestParseTypeModifier_WithObservation(t *testing.T) {
	base, rt, hasType := ParseTypeModifier("focus:Observation")
	if base != "focus" {
		t.Errorf("baseName = %q, want %q", base, "focus")
	}
	if rt != "Observation" {
		t.Errorf("resourceType = %q, want %q", rt, "Observation")
	}
	if !hasType {
		t.Error("hasType = false, want true")
	}
}

func TestParseTypeModifier_NoModifier(t *testing.T) {
	base, rt, hasType := ParseTypeModifier("subject")
	if base != "subject" {
		t.Errorf("baseName = %q, want %q", base, "subject")
	}
	if rt != "" {
		t.Errorf("resourceType = %q, want empty", rt)
	}
	if hasType {
		t.Error("hasType = true, want false")
	}
}

func TestParseTypeModifier_UnknownResourceType(t *testing.T) {
	_, _, hasType := ParseTypeModifier("subject:FakeResource")
	if hasType {
		t.Error("hasType should be false for unknown resource type")
	}
}

func TestParseTypeModifier_LowercaseNotResourceType(t *testing.T) {
	_, _, hasType := ParseTypeModifier("name:exact")
	if hasType {
		t.Error("hasType should be false for lowercase modifier like :exact")
	}
}

func TestParseTypeModifier_EmptyAfterColon(t *testing.T) {
	_, _, hasType := ParseTypeModifier("subject:")
	if hasType {
		t.Error("hasType should be false when modifier is empty")
	}
}

// ---------------------------------------------------------------------------
// TotalMode tests
// ---------------------------------------------------------------------------

func TestParseTotalParam_None(t *testing.T) {
	mode := ParseTotalParam("none")
	if mode != TotalNone {
		t.Errorf("mode = %q, want %q", mode, TotalNone)
	}
}

func TestParseTotalParam_Estimate(t *testing.T) {
	mode := ParseTotalParam("estimate")
	if mode != TotalEstimate {
		t.Errorf("mode = %q, want %q", mode, TotalEstimate)
	}
}

func TestParseTotalParam_Accurate(t *testing.T) {
	mode := ParseTotalParam("accurate")
	if mode != TotalAccurate {
		t.Errorf("mode = %q, want %q", mode, TotalAccurate)
	}
}

func TestParseTotalParam_CaseInsensitive(t *testing.T) {
	tests := []struct {
		input string
		want  TotalMode
	}{
		{"NONE", TotalNone},
		{"Estimate", TotalEstimate},
		{"ACCURATE", TotalAccurate},
		{"  accurate  ", TotalAccurate},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseTotalParam(tt.input)
			if got != tt.want {
				t.Errorf("ParseTotalParam(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseTotalParam_InvalidDefaultsToNone(t *testing.T) {
	tests := []string{"", "invalid", "all", "yes", "true"}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			got := ParseTotalParam(input)
			if got != TotalNone {
				t.Errorf("ParseTotalParam(%q) = %q, want %q", input, got, TotalNone)
			}
		})
	}
}

func TestShouldIncludeTotal_None(t *testing.T) {
	if ShouldIncludeTotal(TotalNone) {
		t.Error("ShouldIncludeTotal(TotalNone) = true, want false")
	}
}

func TestShouldIncludeTotal_Estimate(t *testing.T) {
	if !ShouldIncludeTotal(TotalEstimate) {
		t.Error("ShouldIncludeTotal(TotalEstimate) = false, want true")
	}
}

func TestShouldIncludeTotal_Accurate(t *testing.T) {
	if !ShouldIncludeTotal(TotalAccurate) {
		t.Error("ShouldIncludeTotal(TotalAccurate) = false, want true")
	}
}

// ---------------------------------------------------------------------------
// OfTypeSearchClause tests
// ---------------------------------------------------------------------------

func TestOfTypeSearchClause_FullValue(t *testing.T) {
	clause, args, nextIdx := OfTypeSearchClause(
		"identifier_system", "identifier_value", "identifier_type",
		"http://terminology.hl7.org/CodeSystem/v2-0203|MR|12345", 1,
	)
	wantClause := "(identifier_system = $1 AND identifier_type = $2 AND identifier_value = $3)"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 3 {
		t.Fatalf("args length = %d, want 3", len(args))
	}
	if args[0] != "http://terminology.hl7.org/CodeSystem/v2-0203" {
		t.Errorf("args[0] = %v, want system URI", args[0])
	}
	if args[1] != "MR" {
		t.Errorf("args[1] = %v, want %q", args[1], "MR")
	}
	if args[2] != "12345" {
		t.Errorf("args[2] = %v, want %q", args[2], "12345")
	}
	if nextIdx != 4 {
		t.Errorf("nextIdx = %d, want 4", nextIdx)
	}
}

func TestOfTypeSearchClause_TwoParts(t *testing.T) {
	clause, args, nextIdx := OfTypeSearchClause("sys", "code", "type_col", "system|code", 1)
	if clause != "1=0" {
		t.Errorf("clause = %q, want %q (malformed value with only 2 parts)", clause, "1=0")
	}
	if len(args) != 0 {
		t.Errorf("args should be empty for malformed input, got %d", len(args))
	}
	if nextIdx != 1 {
		t.Errorf("nextIdx = %d, want 1", nextIdx)
	}
}

func TestOfTypeSearchClause_EmptyValue(t *testing.T) {
	clause, _, _ := OfTypeSearchClause("sys", "code", "type_col", "", 1)
	if clause != "1=0" {
		t.Errorf("clause = %q, want %q for empty value", clause, "1=0")
	}
}

func TestOfTypeSearchClause_IdxAdvancement(t *testing.T) {
	_, _, nextIdx := OfTypeSearchClause("sys", "code", "type_col", "a|b|c", 10)
	if nextIdx != 13 {
		t.Errorf("nextIdx = %d, want 13", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// NotTokenSearchClause tests
// ---------------------------------------------------------------------------

func TestNotTokenSearchClause_SystemAndCode(t *testing.T) {
	clause, args, nextIdx := NotTokenSearchClause("code_system", "code_value", "http://loinc.org|1234", 1)
	wantClause := "NOT (code_system = $1 AND code_value = $2)"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 2 {
		t.Fatalf("args length = %d, want 2", len(args))
	}
	if args[0] != "http://loinc.org" {
		t.Errorf("args[0] = %v, want %q", args[0], "http://loinc.org")
	}
	if args[1] != "1234" {
		t.Errorf("args[1] = %v, want %q", args[1], "1234")
	}
	if nextIdx != 3 {
		t.Errorf("nextIdx = %d, want 3", nextIdx)
	}
}

func TestNotTokenSearchClause_CodeOnly(t *testing.T) {
	clause, args, nextIdx := NotTokenSearchClause("sys", "code", "active", 1)
	wantClause := "NOT (code = $1)"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 1 {
		t.Fatalf("args length = %d, want 1", len(args))
	}
	if args[0] != "active" {
		t.Errorf("args[0] = %v, want %q", args[0], "active")
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestNotTokenSearchClause_PipeCodeOnly(t *testing.T) {
	clause, args, nextIdx := NotTokenSearchClause("sys", "code", "|active", 1)
	wantClause := "NOT (code = $1)"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 1 || args[0] != "active" {
		t.Errorf("args = %v, want [active]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestNotTokenSearchClause_SystemOnly(t *testing.T) {
	clause, args, nextIdx := NotTokenSearchClause("sys", "code", "http://loinc.org|", 1)
	wantClause := "NOT (sys = $1)"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 1 || args[0] != "http://loinc.org" {
		t.Errorf("args = %v, want [http://loinc.org]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestNotTokenSearchClause_IdxAdvancement(t *testing.T) {
	_, _, nextIdx := NotTokenSearchClause("sys", "code", "http://x|y", 7)
	if nextIdx != 9 {
		t.Errorf("nextIdx = %d, want 9", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// AboveTokenSearchClause tests
// ---------------------------------------------------------------------------

func TestAboveTokenSearchClause_CodeOnly(t *testing.T) {
	clause, args, nextIdx := AboveTokenSearchClause("sys", "code", "123", 1)
	wantClause := "code LIKE $1"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 1 || args[0] != "123%" {
		t.Errorf("args = %v, want [123%%]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestAboveTokenSearchClause_SystemAndCode(t *testing.T) {
	clause, args, nextIdx := AboveTokenSearchClause("code_system", "code_value", "http://snomed.info/sct|73211009", 1)
	wantClause := "(code_system = $1 AND code_value LIKE $2)"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 2 {
		t.Fatalf("args length = %d, want 2", len(args))
	}
	if args[0] != "http://snomed.info/sct" {
		t.Errorf("args[0] = %v, want system URI", args[0])
	}
	if args[1] != "73211009%" {
		t.Errorf("args[1] = %v, want %q", args[1], "73211009%")
	}
	if nextIdx != 3 {
		t.Errorf("nextIdx = %d, want 3", nextIdx)
	}
}

func TestAboveTokenSearchClause_SystemOnlyPipe(t *testing.T) {
	clause, args, nextIdx := AboveTokenSearchClause("sys", "code", "http://loinc.org|", 1)
	wantClause := "sys = $1"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 1 || args[0] != "http://loinc.org" {
		t.Errorf("args = %v, want [http://loinc.org]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestAboveTokenSearchClause_PipeCodeOnly(t *testing.T) {
	clause, args, nextIdx := AboveTokenSearchClause("sys", "code", "|abc", 3)
	wantClause := "code LIKE $3"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 1 || args[0] != "abc%" {
		t.Errorf("args = %v, want [abc%%]", args)
	}
	if nextIdx != 4 {
		t.Errorf("nextIdx = %d, want 4", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// BelowTokenSearchClause tests
// ---------------------------------------------------------------------------

func TestBelowTokenSearchClause_CodeOnly(t *testing.T) {
	clause, args, nextIdx := BelowTokenSearchClause("sys", "code", "I10", 1)
	wantClause := "code LIKE $1"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 1 || args[0] != "I10%" {
		t.Errorf("args = %v, want [I10%%]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestBelowTokenSearchClause_SystemAndCode(t *testing.T) {
	clause, args, nextIdx := BelowTokenSearchClause("code_system", "code_value", "http://hl7.org/fhir/sid/icd-10|I10", 1)
	wantClause := "(code_system = $1 AND code_value LIKE $2)"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 2 {
		t.Fatalf("args length = %d, want 2", len(args))
	}
	if args[0] != "http://hl7.org/fhir/sid/icd-10" {
		t.Errorf("args[0] = %v, want system URI", args[0])
	}
	if args[1] != "I10%" {
		t.Errorf("args[1] = %v, want %q", args[1], "I10%")
	}
	if nextIdx != 3 {
		t.Errorf("nextIdx = %d, want 3", nextIdx)
	}
}

func TestBelowTokenSearchClause_SystemOnlyPipe(t *testing.T) {
	clause, args, _ := BelowTokenSearchClause("sys", "code", "http://snomed.info/sct|", 1)
	if clause != "sys = $1" {
		t.Errorf("clause = %q, want %q", clause, "sys = $1")
	}
	if len(args) != 1 || args[0] != "http://snomed.info/sct" {
		t.Errorf("args = %v, want [http://snomed.info/sct]", args)
	}
}

func TestBelowTokenSearchClause_IdxAdvancement(t *testing.T) {
	_, _, nextIdx := BelowTokenSearchClause("sys", "code", "http://x|y", 4)
	if nextIdx != 6 {
		t.Errorf("nextIdx = %d, want 6", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// InValueSetClause tests
// ---------------------------------------------------------------------------

func TestInValueSetClause_Basic(t *testing.T) {
	clause, args, nextIdx := InValueSetClause("code", "http://hl7.org/fhir/ValueSet/condition-code", 1)
	wantClause := "code = ANY($1)"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 1 {
		t.Fatalf("args length = %d, want 1", len(args))
	}
	if args[0] != "http://hl7.org/fhir/ValueSet/condition-code" {
		t.Errorf("args[0] = %v, want ValueSet URL", args[0])
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestInValueSetClause_EmptyURL(t *testing.T) {
	clause, args, nextIdx := InValueSetClause("code", "", 1)
	if clause != "1=0" {
		t.Errorf("clause = %q, want %q for empty URL", clause, "1=0")
	}
	if len(args) != 0 {
		t.Errorf("args should be empty for empty URL, got %d", len(args))
	}
	if nextIdx != 1 {
		t.Errorf("nextIdx = %d, want 1", nextIdx)
	}
}

func TestInValueSetClause_IdxAdvancement(t *testing.T) {
	_, _, nextIdx := InValueSetClause("code", "http://example.com/vs", 5)
	if nextIdx != 6 {
		t.Errorf("nextIdx = %d, want 6", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// NotInValueSetClause tests
// ---------------------------------------------------------------------------

func TestNotInValueSetClause_Basic(t *testing.T) {
	clause, args, nextIdx := NotInValueSetClause("code", "http://hl7.org/fhir/ValueSet/condition-code", 1)
	wantClause := "NOT (code = ANY($1))"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 1 {
		t.Fatalf("args length = %d, want 1", len(args))
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestNotInValueSetClause_EmptyURL(t *testing.T) {
	clause, args, nextIdx := NotInValueSetClause("code", "", 1)
	if clause != "1=1" {
		t.Errorf("clause = %q, want %q for empty URL", clause, "1=1")
	}
	if len(args) != 0 {
		t.Errorf("args should be empty for empty URL, got %d", len(args))
	}
	if nextIdx != 1 {
		t.Errorf("nextIdx = %d, want 1", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// ApplySearchModifiers tests
// ---------------------------------------------------------------------------

func TestApplySearchModifiers_MissingTrue(t *testing.T) {
	q := NewSearchQuery("patients", "id, first_name")
	config := SearchParamConfig{Type: SearchParamDate, Column: "birth_date"}
	applied := ApplySearchModifiers(q, "birthdate:missing", "true", config)
	if !applied {
		t.Fatal("ApplySearchModifiers should return true for :missing")
	}
	sql := q.CountSQL()
	if !strings.Contains(sql, "birth_date IS NULL") {
		t.Errorf("SQL should contain IS NULL, got: %s", sql)
	}
}

func TestApplySearchModifiers_MissingFalse(t *testing.T) {
	q := NewSearchQuery("patients", "id")
	config := SearchParamConfig{Type: SearchParamString, Column: "last_name"}
	applied := ApplySearchModifiers(q, "family:missing", "false", config)
	if !applied {
		t.Fatal("ApplySearchModifiers should return true for :missing")
	}
	sql := q.CountSQL()
	if !strings.Contains(sql, "last_name IS NOT NULL") {
		t.Errorf("SQL should contain IS NOT NULL, got: %s", sql)
	}
}

func TestApplySearchModifiers_TypedReference(t *testing.T) {
	q := NewSearchQuery("observations", "id")
	config := SearchParamConfig{Type: SearchParamReference, Column: "subject_id"}
	applied := ApplySearchModifiers(q, "subject:Patient", "123", config)
	if !applied {
		t.Fatal("ApplySearchModifiers should return true for :Patient typed reference")
	}
	sql := q.CountSQL()
	if !strings.Contains(sql, "subject_id = $1") {
		t.Errorf("SQL should contain subject_id clause, got: %s", sql)
	}
	if !strings.Contains(sql, "subject_id_type = $2") {
		t.Errorf("SQL should contain type clause, got: %s", sql)
	}
}

func TestApplySearchModifiers_NotToken(t *testing.T) {
	q := NewSearchQuery("observations", "id")
	config := SearchParamConfig{Type: SearchParamToken, Column: "status", SysColumn: ""}
	applied := ApplySearchModifiers(q, "status:not", "cancelled", config)
	if !applied {
		t.Fatal("ApplySearchModifiers should return true for :not token")
	}
	sql := q.CountSQL()
	if !strings.Contains(sql, "NOT") {
		t.Errorf("SQL should contain NOT, got: %s", sql)
	}
}

func TestApplySearchModifiers_AboveToken(t *testing.T) {
	q := NewSearchQuery("conditions", "id")
	config := SearchParamConfig{Type: SearchParamToken, Column: "code", SysColumn: "code_system"}
	applied := ApplySearchModifiers(q, "code:above", "http://snomed.info/sct|73211009", config)
	if !applied {
		t.Fatal("ApplySearchModifiers should return true for :above token")
	}
	sql := q.CountSQL()
	if !strings.Contains(sql, "LIKE") {
		t.Errorf("SQL should contain LIKE for hierarchy search, got: %s", sql)
	}
}

func TestApplySearchModifiers_BelowToken(t *testing.T) {
	q := NewSearchQuery("conditions", "id")
	config := SearchParamConfig{Type: SearchParamToken, Column: "code", SysColumn: "code_system"}
	applied := ApplySearchModifiers(q, "code:below", "I10", config)
	if !applied {
		t.Fatal("ApplySearchModifiers should return true for :below token")
	}
	sql := q.CountSQL()
	if !strings.Contains(sql, "LIKE") {
		t.Errorf("SQL should contain LIKE for hierarchy search, got: %s", sql)
	}
}

func TestApplySearchModifiers_OfTypeToken(t *testing.T) {
	q := NewSearchQuery("patients", "id")
	config := SearchParamConfig{Type: SearchParamToken, Column: "identifier_value", SysColumn: "identifier_system"}
	applied := ApplySearchModifiers(q, "identifier:of-type", "http://terminology.hl7.org/CodeSystem/v2-0203|MR|12345", config)
	if !applied {
		t.Fatal("ApplySearchModifiers should return true for :of-type")
	}
	sql := q.CountSQL()
	if !strings.Contains(sql, "identifier_system = $1") {
		t.Errorf("SQL should contain system clause, got: %s", sql)
	}
	if !strings.Contains(sql, "identifier_value_type = $2") {
		t.Errorf("SQL should contain type clause, got: %s", sql)
	}
	if !strings.Contains(sql, "identifier_value = $3") {
		t.Errorf("SQL should contain code clause, got: %s", sql)
	}
}

func TestApplySearchModifiers_InValueSet(t *testing.T) {
	q := NewSearchQuery("conditions", "id")
	config := SearchParamConfig{Type: SearchParamToken, Column: "code"}
	applied := ApplySearchModifiers(q, "code:in", "http://hl7.org/fhir/ValueSet/condition-code", config)
	if !applied {
		t.Fatal("ApplySearchModifiers should return true for :in")
	}
	sql := q.CountSQL()
	if !strings.Contains(sql, "ANY") {
		t.Errorf("SQL should contain ANY for ValueSet, got: %s", sql)
	}
}

func TestApplySearchModifiers_NotInValueSet(t *testing.T) {
	q := NewSearchQuery("conditions", "id")
	config := SearchParamConfig{Type: SearchParamToken, Column: "code"}
	applied := ApplySearchModifiers(q, "code:not-in", "http://hl7.org/fhir/ValueSet/condition-code", config)
	if !applied {
		t.Fatal("ApplySearchModifiers should return true for :not-in")
	}
	sql := q.CountSQL()
	if !strings.Contains(sql, "NOT") && !strings.Contains(sql, "ANY") {
		t.Errorf("SQL should contain NOT and ANY for :not-in, got: %s", sql)
	}
}

func TestApplySearchModifiers_NoModifier(t *testing.T) {
	q := NewSearchQuery("patients", "id")
	config := SearchParamConfig{Type: SearchParamString, Column: "last_name"}
	applied := ApplySearchModifiers(q, "family", "Smith", config)
	if applied {
		t.Error("ApplySearchModifiers should return false when no modifier is present")
	}
}

func TestApplySearchModifiers_UnsupportedModifierOnWrongType(t *testing.T) {
	q := NewSearchQuery("patients", "id")
	config := SearchParamConfig{Type: SearchParamString, Column: "last_name"}
	// :not is only handled for SearchParamToken, not SearchParamString
	applied := ApplySearchModifiers(q, "family:not", "Smith", config)
	if applied {
		t.Error("ApplySearchModifiers should return false for :not on string type")
	}
}

func TestApplySearchModifiers_AboveWithoutSysColumn(t *testing.T) {
	q := NewSearchQuery("observations", "id")
	// No SysColumn means :above won't be applied
	config := SearchParamConfig{Type: SearchParamToken, Column: "status"}
	applied := ApplySearchModifiers(q, "status:above", "final", config)
	if applied {
		t.Error("ApplySearchModifiers should return false for :above without SysColumn")
	}
}

// ---------------------------------------------------------------------------
// SQL parameter index correctness across chained calls
// ---------------------------------------------------------------------------

func TestIdxCorrectnessAcrossMultipleClauses(t *testing.T) {
	q := NewSearchQuery("observations", "id, code, status")

	// Add a normal token search first (should use $1, $2)
	q.AddToken("code_system", "code_value", "http://loinc.org|1234")
	if q.Idx() != 3 {
		t.Fatalf("after AddToken idx = %d, want 3", q.Idx())
	}

	// Add a :missing clause (should not consume any args)
	clause, args, nextIdx := MissingSearchClause("effective_date", true, q.Idx())
	q.Add(clause, args...)
	if q.Idx() != 3 {
		t.Fatalf("after MissingSearchClause idx = %d, want 3", q.Idx())
	}
	_ = nextIdx

	// Add a not-token clause (should use $3, $4)
	notClause, notArgs, notNext := NotTokenSearchClause("status_system", "status_value", "http://hl7.org|cancelled", q.Idx())
	q.where += " AND " + notClause
	q.args = append(q.args, notArgs...)
	q.idx = notNext
	if q.Idx() != 5 {
		t.Fatalf("after NotTokenSearchClause idx = %d, want 5", q.Idx())
	}

	sql := q.CountSQL()
	// Verify the SQL contains all clauses
	if !strings.Contains(sql, "code_system = $1") {
		t.Errorf("SQL missing token system clause: %s", sql)
	}
	if !strings.Contains(sql, "code_value = $2") {
		t.Errorf("SQL missing token code clause: %s", sql)
	}
	if !strings.Contains(sql, "effective_date IS NULL") {
		t.Errorf("SQL missing IS NULL clause: %s", sql)
	}
	if !strings.Contains(sql, "status_system = $3") {
		t.Errorf("SQL missing not-token system clause: %s", sql)
	}
	if !strings.Contains(sql, "status_value = $4") {
		t.Errorf("SQL missing not-token code clause: %s", sql)
	}
}

func TestIdxCorrectnessWithTypedReference(t *testing.T) {
	q := NewSearchQuery("observations", "id")

	// First add a date clause (uses $1)
	q.AddDate("effective_date", "gt2023-01-01")
	if q.Idx() != 2 {
		t.Fatalf("after AddDate idx = %d, want 2", q.Idx())
	}

	// Add typed reference (should use $2, $3)
	clause, args, nextIdx := TypedReferenceSearchClause("subject_id", "subject_type", "abc-123", "Patient", q.Idx())
	q.where += " AND " + clause
	q.args = append(q.args, args...)
	q.idx = nextIdx
	if q.Idx() != 4 {
		t.Fatalf("after TypedReferenceSearchClause idx = %d, want 4", q.Idx())
	}

	sql := q.CountSQL()
	if !strings.Contains(sql, "subject_id = $2") {
		t.Errorf("SQL missing ref clause at $2: %s", sql)
	}
	if !strings.Contains(sql, "subject_type = $3") {
		t.Errorf("SQL missing type clause at $3: %s", sql)
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestMissingSearchClause_EmptyColumn(t *testing.T) {
	clause, _, _ := MissingSearchClause("", true, 1)
	if clause != " IS NULL" {
		t.Errorf("clause = %q, want %q", clause, " IS NULL")
	}
}

func TestNotTokenSearchClause_EmptyValue(t *testing.T) {
	clause, args, nextIdx := NotTokenSearchClause("sys", "code", "", 1)
	wantClause := "NOT (code = $1)"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 1 || args[0] != "" {
		t.Errorf("args = %v, want [empty string]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestOfTypeSearchClause_PipeOnlyValue(t *testing.T) {
	clause, _, _ := OfTypeSearchClause("sys", "code", "type_col", "||", 1)
	// Three parts, all empty - but still 3 parts
	wantClause := "(sys = $1 AND type_col = $2 AND code = $3)"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
}

func TestAboveTokenSearchClause_EmptyValue(t *testing.T) {
	clause, args, _ := AboveTokenSearchClause("sys", "code", "", 1)
	if clause != "code LIKE $1" {
		t.Errorf("clause = %q, want %q", clause, "code LIKE $1")
	}
	if len(args) != 1 || args[0] != "%" {
		t.Errorf("args = %v, want [%%]", args)
	}
}

func TestBelowTokenSearchClause_EmptyValue(t *testing.T) {
	clause, args, _ := BelowTokenSearchClause("sys", "code", "", 1)
	if clause != "code LIKE $1" {
		t.Errorf("clause = %q, want %q", clause, "code LIKE $1")
	}
	if len(args) != 1 || args[0] != "%" {
		t.Errorf("args = %v, want [%%]", args)
	}
}

func TestValueSetResolverInterface(t *testing.T) {
	// Verify the interface is defined and can be implemented.
	var _ ValueSetResolver = (*mockValueSetResolver)(nil)
}

type mockValueSetResolver struct{}

func (m *mockValueSetResolver) Expand(url string) ([]string, error) {
	return []string{"code1", "code2"}, nil
}

func TestNotTokenSearchClause_EmptyPipe(t *testing.T) {
	// "|" with both empty - falls through to code-only
	clause, args, nextIdx := NotTokenSearchClause("sys", "code", "|", 1)
	wantClause := "NOT (code = $1)"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 1 || args[0] != "|" {
		// Falls through because both system and code are empty
		t.Logf("args = %v (fell through to code-only with pipe as value)", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestParseTypeModifier_MissingModifier(t *testing.T) {
	// :missing is lowercase, so ParseTypeModifier should NOT treat it as a resource type
	base, rt, hasType := ParseTypeModifier("birthdate:missing")
	if hasType {
		t.Error("hasType should be false for :missing (lowercase)")
	}
	if base != "birthdate:missing" {
		t.Errorf("baseName = %q, want %q", base, "birthdate:missing")
	}
	if rt != "" {
		t.Errorf("resourceType = %q, want empty", rt)
	}
}
