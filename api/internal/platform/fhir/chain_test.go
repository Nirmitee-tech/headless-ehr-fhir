package fhir

import (
	"context"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// ParseChainedParam tests
// ---------------------------------------------------------------------------

func TestParseChainedParam(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		valid  bool
		source string
		target string
		tParam string
	}{
		{"basic chain", "subject:Patient.name", true, "subject", "Patient", "name"},
		{"without type", "subject.name", true, "subject", "", "name"},
		{"not chained", "name", false, "", "", ""},
		{"empty target param", "subject.", false, "", "", ""},
		{"typed chain", "performer:Practitioner.name", true, "performer", "Practitioner", "name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := ParseChainedParam(tt.input)
			if ok != tt.valid {
				t.Fatalf("ParseChainedParam(%q) valid = %v, want %v", tt.input, ok, tt.valid)
			}
			if !ok {
				return
			}
			if result.SourceParam != tt.source {
				t.Errorf("SourceParam = %q, want %q", result.SourceParam, tt.source)
			}
			if result.TargetType != tt.target {
				t.Errorf("TargetType = %q, want %q", result.TargetType, tt.target)
			}
			if result.TargetParam != tt.tParam {
				t.Errorf("TargetParam = %q, want %q", result.TargetParam, tt.tParam)
			}
		})
	}
}

func TestParseChainedParam_MultipleDotsPicksFirst(t *testing.T) {
	// "subject:Patient.name.family" should parse with targetParam = "name.family"
	result, ok := ParseChainedParam("subject:Patient.name.family")
	if !ok {
		t.Fatal("expected valid chained param")
	}
	if result.SourceParam != "subject" {
		t.Errorf("SourceParam = %q, want 'subject'", result.SourceParam)
	}
	if result.TargetType != "Patient" {
		t.Errorf("TargetType = %q, want 'Patient'", result.TargetType)
	}
	if result.TargetParam != "name.family" {
		t.Errorf("TargetParam = %q, want 'name.family'", result.TargetParam)
	}
}

// ---------------------------------------------------------------------------
// ParseHasParam tests
// ---------------------------------------------------------------------------

func TestParseHasParam(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		valid  bool
		tType  string
		tParam string
		sParam string
	}{
		{"valid _has", "_has:Observation:subject:code", true, "Observation", "subject", "code"},
		{"not _has", "name", false, "", "", ""},
		{"incomplete", "_has:Observation:subject", false, "", "", ""},
		{"complex", "_has:MedicationRequest:patient:status", true, "MedicationRequest", "patient", "status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := ParseHasParam(tt.input)
			if ok != tt.valid {
				t.Fatalf("ParseHasParam(%q) valid = %v, want %v", tt.input, ok, tt.valid)
			}
			if !ok {
				return
			}
			if result.TargetType != tt.tType {
				t.Errorf("TargetType = %q, want %q", result.TargetType, tt.tType)
			}
			if result.TargetParam != tt.tParam {
				t.Errorf("TargetParam = %q, want %q", result.TargetParam, tt.tParam)
			}
			if result.SearchParam != tt.sParam {
				t.Errorf("SearchParam = %q, want %q", result.SearchParam, tt.sParam)
			}
		})
	}
}

func TestParseHasParam_ExactlyHasPrefix(t *testing.T) {
	// "_has:" alone with less than 3 parts should fail
	_, ok := ParseHasParam("_has:")
	if ok {
		t.Error("expected false for _has: with no parts")
	}
}

func TestParseHasParam_OnePart(t *testing.T) {
	_, ok := ParseHasParam("_has:Observation")
	if ok {
		t.Error("expected false for _has: with only 1 part")
	}
}

// ---------------------------------------------------------------------------
// BuildChainedINClause tests
// ---------------------------------------------------------------------------

func TestBuildChainedINClause(t *testing.T) {
	// Empty IDs - should return false condition
	clause, args, idx := BuildChainedINClause("patient_id", nil, 1)
	if clause != "1=0" {
		t.Errorf("empty ids clause = %q, want '1=0'", clause)
	}
	if args != nil {
		t.Errorf("empty ids args should be nil")
	}
	if idx != 1 {
		t.Errorf("empty ids nextIdx = %d, want 1", idx)
	}

	// Multiple IDs
	clause, args, idx = BuildChainedINClause("patient_id", []string{"a", "b", "c"}, 1)
	if clause != "patient_id IN ($1, $2, $3)" {
		t.Errorf("clause = %q", clause)
	}
	if len(args) != 3 {
		t.Errorf("args count = %d, want 3", len(args))
	}
	if idx != 4 {
		t.Errorf("nextIdx = %d, want 4", idx)
	}

	// Single ID with non-1 start index
	clause, _, idx = BuildChainedINClause("id", []string{"x"}, 5)
	if clause != "id IN ($5)" {
		t.Errorf("clause = %q", clause)
	}
	if idx != 6 {
		t.Errorf("nextIdx = %d, want 6", idx)
	}
}

func TestBuildChainedINClause_SingleID(t *testing.T) {
	clause, args, idx := BuildChainedINClause("col", []string{"single"}, 1)
	if clause != "col IN ($1)" {
		t.Errorf("clause = %q, want %q", clause, "col IN ($1)")
	}
	if len(args) != 1 || args[0] != "single" {
		t.Errorf("args = %v, want [single]", args)
	}
	if idx != 2 {
		t.Errorf("nextIdx = %d, want 2", idx)
	}
}

func TestBuildChainedINClause_ArgValues(t *testing.T) {
	_, args, _ := BuildChainedINClause("id", []string{"a", "b"}, 10)
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != "a" {
		t.Errorf("args[0] = %v, want 'a'", args[0])
	}
	if args[1] != "b" {
		t.Errorf("args[1] = %v, want 'b'", args[1])
	}
}

// ---------------------------------------------------------------------------
// ChainResolver constructor tests
// ---------------------------------------------------------------------------

func TestNewChainResolver(t *testing.T) {
	t.Run("with nil registry", func(t *testing.T) {
		cr := NewChainResolver(nil)
		if cr == nil {
			t.Fatal("NewChainResolver(nil) returned nil, want non-nil")
		}
		if cr.registry != nil {
			t.Errorf("registry should be nil")
		}
	})

	t.Run("with valid registry", func(t *testing.T) {
		reg := NewIncludeRegistry()
		cr := NewChainResolver(reg)
		if cr == nil {
			t.Fatal("NewChainResolver returned nil")
		}
		if cr.registry != reg {
			t.Errorf("registry not set correctly")
		}
	})
}

func TestNewChainResolverWithRegistry(t *testing.T) {
	incReg := NewIncludeRegistry()
	chainReg := NewChainRegistry()
	cr := NewChainResolverWithRegistry(incReg, chainReg)
	if cr == nil {
		t.Fatal("NewChainResolverWithRegistry returned nil")
	}
	if cr.registry != incReg {
		t.Error("include registry not set correctly")
	}
	if cr.chainRegistry != chainReg {
		t.Error("chain registry not set correctly")
	}
}

// ---------------------------------------------------------------------------
// ResolveChainedParam / ResolveHasParam (existing stub behavior preserved)
// ---------------------------------------------------------------------------

func TestResolveChainedParam(t *testing.T) {
	ctx := context.Background()

	t.Run("nil registry returns error", func(t *testing.T) {
		cr := NewChainResolver(nil)
		chain := &ChainedParam{
			SourceParam: "subject",
			TargetType:  "Patient",
			TargetParam: "name",
			Value:       "John",
		}
		ids, err := cr.ResolveChainedParam(ctx, chain)
		if err == nil {
			t.Fatal("expected error for nil registry, got nil")
		}
		if !strings.Contains(err.Error(), "no include registry configured") {
			t.Errorf("error = %q, want it to contain 'no include registry configured'", err.Error())
		}
		if ids != nil {
			t.Errorf("ids should be nil, got %v", ids)
		}
	})

	t.Run("non-nil registry returns search execution error", func(t *testing.T) {
		reg := NewIncludeRegistry()
		cr := NewChainResolver(reg)
		chain := &ChainedParam{
			SourceParam: "subject",
			TargetType:  "Patient",
			TargetParam: "name",
			Value:       "John",
		}
		ids, err := cr.ResolveChainedParam(ctx, chain)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "chained parameter resolution requires search execution on Patient") {
			t.Errorf("error = %q, want it to mention target type 'Patient'", err.Error())
		}
		if ids != nil {
			t.Errorf("ids should be nil, got %v", ids)
		}
	})

	t.Run("error message contains correct target type", func(t *testing.T) {
		reg := NewIncludeRegistry()
		cr := NewChainResolver(reg)
		chain := &ChainedParam{
			SourceParam: "performer",
			TargetType:  "Practitioner",
			TargetParam: "name",
			Value:       "Smith",
		}
		_, err := cr.ResolveChainedParam(ctx, chain)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "Practitioner") {
			t.Errorf("error = %q, want it to mention 'Practitioner'", err.Error())
		}
	})
}

func TestResolveHasParam(t *testing.T) {
	ctx := context.Background()

	t.Run("nil registry returns error", func(t *testing.T) {
		cr := NewChainResolver(nil)
		has := &HasParam{
			TargetType:  "Observation",
			TargetParam: "subject",
			SearchParam: "code",
			Value:       "1234",
		}
		ids, err := cr.ResolveHasParam(ctx, has, "Patient")
		if err == nil {
			t.Fatal("expected error for nil registry, got nil")
		}
		if !strings.Contains(err.Error(), "no include registry configured") {
			t.Errorf("error = %q, want it to contain 'no include registry configured'", err.Error())
		}
		if ids != nil {
			t.Errorf("ids should be nil, got %v", ids)
		}
	})

	t.Run("non-nil registry returns search execution error", func(t *testing.T) {
		reg := NewIncludeRegistry()
		cr := NewChainResolver(reg)
		has := &HasParam{
			TargetType:  "Observation",
			TargetParam: "subject",
			SearchParam: "code",
			Value:       "1234",
		}
		ids, err := cr.ResolveHasParam(ctx, has, "Patient")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "_has parameter resolution requires search execution on Observation") {
			t.Errorf("error = %q, want it to mention target type 'Observation'", err.Error())
		}
		if ids != nil {
			t.Errorf("ids should be nil, got %v", ids)
		}
	})

	t.Run("error message contains correct target type", func(t *testing.T) {
		reg := NewIncludeRegistry()
		cr := NewChainResolver(reg)
		has := &HasParam{
			TargetType:  "MedicationRequest",
			TargetParam: "patient",
			SearchParam: "status",
			Value:       "active",
		}
		_, err := cr.ResolveHasParam(ctx, has, "Patient")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "MedicationRequest") {
			t.Errorf("error = %q, want it to mention 'MedicationRequest'", err.Error())
		}
	})
}

// ---------------------------------------------------------------------------
// ChainRegistry tests
// ---------------------------------------------------------------------------

func TestChainRegistry_RegisterAndResolve(t *testing.T) {
	r := NewChainRegistry()
	r.Register("Patient", "general-practitioner", "name", ChainedSearchConfig{
		ReferenceColumn: "practitioner_id",
		TargetTable:     "practitioners",
		TargetColumn:    "name",
		TargetType:      SearchParamString,
	})

	config, err := r.Resolve("Patient", "general-practitioner.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.ReferenceColumn != "practitioner_id" {
		t.Errorf("ReferenceColumn = %q, want %q", config.ReferenceColumn, "practitioner_id")
	}
	if config.TargetTable != "practitioners" {
		t.Errorf("TargetTable = %q, want %q", config.TargetTable, "practitioners")
	}
	if config.TargetColumn != "name" {
		t.Errorf("TargetColumn = %q, want %q", config.TargetColumn, "name")
	}
	if config.TargetType != SearchParamString {
		t.Errorf("TargetType = %d, want %d (SearchParamString)", config.TargetType, SearchParamString)
	}
}

func TestChainRegistry_ResolveWithTypeModifier(t *testing.T) {
	r := NewChainRegistry()
	r.Register("Observation", "patient", "name", ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "last_name",
		TargetType:      SearchParamString,
	})

	// chainPath with :Type modifier should still resolve
	config, err := r.Resolve("Observation", "patient:Patient.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TargetColumn != "last_name" {
		t.Errorf("TargetColumn = %q, want %q", config.TargetColumn, "last_name")
	}
}

func TestChainRegistry_ResolveUnknownPath(t *testing.T) {
	r := NewChainRegistry()
	_, err := r.Resolve("Patient", "unknown.name")
	if err == nil {
		t.Fatal("expected error for unknown chain path, got nil")
	}
	if !strings.Contains(err.Error(), "unknown chain path") {
		t.Errorf("error = %q, want it to contain 'unknown chain path'", err.Error())
	}
}

func TestChainRegistry_ResolveMissingDot(t *testing.T) {
	r := NewChainRegistry()
	_, err := r.Resolve("Patient", "nodot")
	if err == nil {
		t.Fatal("expected error for missing dot, got nil")
	}
	if !strings.Contains(err.Error(), "missing dot separator") {
		t.Errorf("error = %q, want it to mention missing dot separator", err.Error())
	}
}

func TestChainRegistry_ResolveUnknownResource(t *testing.T) {
	r := NewChainRegistry()
	r.Register("Patient", "general-practitioner", "name", ChainedSearchConfig{
		ReferenceColumn: "practitioner_id",
		TargetTable:     "practitioners",
		TargetColumn:    "name",
		TargetType:      SearchParamString,
	})

	// Should fail for a different source resource
	_, err := r.Resolve("Observation", "general-practitioner.name")
	if err == nil {
		t.Fatal("expected error for wrong source resource, got nil")
	}
}

func TestChainRegistry_RegisterOverwrite(t *testing.T) {
	r := NewChainRegistry()
	r.Register("Patient", "general-practitioner", "name", ChainedSearchConfig{
		ReferenceColumn: "practitioner_id",
		TargetTable:     "practitioners",
		TargetColumn:    "name",
		TargetType:      SearchParamString,
	})
	// Overwrite
	r.Register("Patient", "general-practitioner", "name", ChainedSearchConfig{
		ReferenceColumn: "practitioner_id",
		TargetTable:     "practitioners",
		TargetColumn:    "full_name",
		TargetType:      SearchParamString,
	})

	config, err := r.Resolve("Patient", "general-practitioner.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TargetColumn != "full_name" {
		t.Errorf("TargetColumn = %q, want %q (overwritten value)", config.TargetColumn, "full_name")
	}
}

func TestChainRegistry_MultipleEntries(t *testing.T) {
	r := NewChainRegistry()
	r.Register("Observation", "patient", "name", ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "last_name",
		TargetType:      SearchParamString,
	})
	r.Register("Observation", "patient", "identifier", ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "identifier_value",
		TargetType:      SearchParamToken,
		TargetSysColumn: "identifier_system",
	})

	cfg1, err := r.Resolve("Observation", "patient.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg1.TargetColumn != "last_name" {
		t.Errorf("name TargetColumn = %q, want 'last_name'", cfg1.TargetColumn)
	}

	cfg2, err := r.Resolve("Observation", "patient.identifier")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg2.TargetColumn != "identifier_value" {
		t.Errorf("identifier TargetColumn = %q, want 'identifier_value'", cfg2.TargetColumn)
	}
	if cfg2.TargetSysColumn != "identifier_system" {
		t.Errorf("identifier TargetSysColumn = %q, want 'identifier_system'", cfg2.TargetSysColumn)
	}
}

// ---------------------------------------------------------------------------
// ChainedSearchClause tests
// ---------------------------------------------------------------------------

func TestChainedSearchClause_StringParam(t *testing.T) {
	config := ChainedSearchConfig{
		ReferenceColumn: "practitioner_id",
		TargetTable:     "practitioners",
		TargetColumn:    "name",
		TargetType:      SearchParamString,
	}

	clause, args, nextIdx := ChainedSearchClause(config, "Smith", 1)
	expected := "practitioner_id IN (SELECT id FROM practitioners WHERE name ILIKE $1)"
	if clause != expected {
		t.Errorf("clause = %q, want %q", clause, expected)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if args[0] != "Smith%" {
		t.Errorf("args[0] = %v, want 'Smith%%'", args[0])
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestChainedSearchClause_TokenParam(t *testing.T) {
	config := ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "identifier_value",
		TargetType:      SearchParamToken,
		TargetSysColumn: "identifier_system",
	}

	// Simple token (no pipe)
	clause, args, nextIdx := ChainedSearchClause(config, "12345", 1)
	expected := "patient_id IN (SELECT id FROM patients WHERE identifier_value = $1)"
	if clause != expected {
		t.Errorf("clause = %q, want %q", clause, expected)
	}
	if len(args) != 1 || args[0] != "12345" {
		t.Errorf("args = %v, want [12345]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestChainedSearchClause_TokenParamWithSystem(t *testing.T) {
	config := ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "identifier_value",
		TargetType:      SearchParamToken,
		TargetSysColumn: "identifier_system",
	}

	// system|code token
	clause, args, nextIdx := ChainedSearchClause(config, "http://example.com|12345", 1)
	expected := "patient_id IN (SELECT id FROM patients WHERE (identifier_system = $1 AND identifier_value = $2))"
	if clause != expected {
		t.Errorf("clause = %q, want %q", clause, expected)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != "http://example.com" || args[1] != "12345" {
		t.Errorf("args = %v, want [http://example.com 12345]", args)
	}
	if nextIdx != 3 {
		t.Errorf("nextIdx = %d, want 3", nextIdx)
	}
}

func TestChainedSearchClause_TokenParamNoSysColumn(t *testing.T) {
	config := ChainedSearchConfig{
		ReferenceColumn: "encounter_id",
		TargetTable:     "encounters",
		TargetColumn:    "status",
		TargetType:      SearchParamToken,
	}

	clause, args, nextIdx := ChainedSearchClause(config, "finished", 1)
	expected := "encounter_id IN (SELECT id FROM encounters WHERE status = $1)"
	if clause != expected {
		t.Errorf("clause = %q, want %q", clause, expected)
	}
	if len(args) != 1 || args[0] != "finished" {
		t.Errorf("args = %v, want [finished]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestChainedSearchClause_DateParam(t *testing.T) {
	config := ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "birth_date",
		TargetType:      SearchParamDate,
	}

	clause, args, nextIdx := ChainedSearchClause(config, "gt2023-01-01", 1)
	if !strings.Contains(clause, "patient_id IN (SELECT id FROM patients WHERE birth_date >") {
		t.Errorf("clause = %q, expected IN subquery with date > clause", clause)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestChainedSearchClause_ReferenceParam(t *testing.T) {
	config := ChainedSearchConfig{
		ReferenceColumn: "encounter_id",
		TargetTable:     "encounters",
		TargetColumn:    "patient_id",
		TargetType:      SearchParamReference,
	}

	clause, args, nextIdx := ChainedSearchClause(config, "Patient/abc-123", 1)
	// Non-UUID FHIR ID resolves via subquery on the patient table.
	expected := "encounter_id IN (SELECT id FROM encounters WHERE patient_id = (SELECT id FROM patient WHERE fhir_id = $1 LIMIT 1))"
	if clause != expected {
		t.Errorf("clause = %q, want %q", clause, expected)
	}
	if len(args) != 1 || args[0] != "abc-123" {
		t.Errorf("args = %v, want [abc-123]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestChainedSearchClause_NumberParam(t *testing.T) {
	config := ChainedSearchConfig{
		ReferenceColumn: "observation_id",
		TargetTable:     "observations",
		TargetColumn:    "value_quantity",
		TargetType:      SearchParamNumber,
	}

	clause, args, nextIdx := ChainedSearchClause(config, "ge100", 1)
	if !strings.Contains(clause, "observation_id IN (SELECT id FROM observations WHERE value_quantity >=") {
		t.Errorf("clause = %q, expected >= for ge prefix", clause)
	}
	if len(args) != 1 || args[0] != "100" {
		t.Errorf("args = %v, want [100]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestChainedSearchClause_URIParam(t *testing.T) {
	config := ChainedSearchConfig{
		ReferenceColumn: "valueset_id",
		TargetTable:     "valuesets",
		TargetColumn:    "url",
		TargetType:      SearchParamURI,
	}

	clause, args, nextIdx := ChainedSearchClause(config, "http://hl7.org/fhir/ValueSet/languages", 1)
	expected := "valueset_id IN (SELECT id FROM valuesets WHERE url = $1)"
	if clause != expected {
		t.Errorf("clause = %q, want %q", clause, expected)
	}
	if len(args) != 1 || args[0] != "http://hl7.org/fhir/ValueSet/languages" {
		t.Errorf("args = %v", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestChainedSearchClause_StartIdxOffset(t *testing.T) {
	config := ChainedSearchConfig{
		ReferenceColumn: "practitioner_id",
		TargetTable:     "practitioners",
		TargetColumn:    "name",
		TargetType:      SearchParamString,
	}

	clause, _, nextIdx := ChainedSearchClause(config, "Jones", 5)
	if !strings.Contains(clause, "$5") {
		t.Errorf("clause = %q, expected to use $5", clause)
	}
	if nextIdx != 6 {
		t.Errorf("nextIdx = %d, want 6", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// MultiLevelChainedSearchClause tests
// ---------------------------------------------------------------------------

func TestMultiLevelChainedSearchClause_SingleLevel(t *testing.T) {
	configs := []ChainedSearchConfig{
		{
			ReferenceColumn: "patient_id",
			TargetTable:     "patients",
			TargetColumn:    "last_name",
			TargetType:      SearchParamString,
		},
	}

	clause, args, nextIdx, err := MultiLevelChainedSearchClause(configs, "Smith", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "patient_id IN (SELECT id FROM patients WHERE last_name ILIKE $1)"
	if clause != expected {
		t.Errorf("clause = %q, want %q", clause, expected)
	}
	if len(args) != 1 || args[0] != "Smith%" {
		t.Errorf("args = %v, want [Smith%%]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestMultiLevelChainedSearchClause_TwoLevels(t *testing.T) {
	configs := []ChainedSearchConfig{
		{
			// Observation -> Encounter
			ReferenceColumn: "encounter_id",
			TargetTable:     "encounters",
			TargetColumn:    "", // not the leaf; unused for intermediate
			TargetType:      SearchParamToken,
		},
		{
			// Encounter -> Patient (leaf)
			ReferenceColumn: "patient_id",
			TargetTable:     "patients",
			TargetColumn:    "last_name",
			TargetType:      SearchParamString,
		},
	}

	clause, args, nextIdx, err := MultiLevelChainedSearchClause(configs, "Smith", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "encounter_id IN (SELECT id FROM encounters WHERE patient_id IN (SELECT id FROM patients WHERE last_name ILIKE $1))"
	if clause != expected {
		t.Errorf("clause = %q, want %q", clause, expected)
	}
	if len(args) != 1 || args[0] != "Smith%" {
		t.Errorf("args = %v, want [Smith%%]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestMultiLevelChainedSearchClause_ThreeLevels(t *testing.T) {
	configs := []ChainedSearchConfig{
		{
			// DiagnosticReport -> Observation
			ReferenceColumn: "observation_id",
			TargetTable:     "observations",
			TargetColumn:    "",
			TargetType:      SearchParamReference,
		},
		{
			// Observation -> Encounter
			ReferenceColumn: "encounter_id",
			TargetTable:     "encounters",
			TargetColumn:    "",
			TargetType:      SearchParamReference,
		},
		{
			// Encounter -> Patient (leaf)
			ReferenceColumn: "patient_id",
			TargetTable:     "patients",
			TargetColumn:    "last_name",
			TargetType:      SearchParamString,
		},
	}

	clause, args, nextIdx, err := MultiLevelChainedSearchClause(configs, "Doe", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "observation_id IN (SELECT id FROM observations WHERE encounter_id IN (SELECT id FROM encounters WHERE patient_id IN (SELECT id FROM patients WHERE last_name ILIKE $1)))"
	if clause != expected {
		t.Errorf("clause = %q, want %q", clause, expected)
	}
	if len(args) != 1 || args[0] != "Doe%" {
		t.Errorf("args = %v, want [Doe%%]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestMultiLevelChainedSearchClause_ExceedsMaxDepth(t *testing.T) {
	configs := make([]ChainedSearchConfig, MaxChainDepth+1)
	for i := range configs {
		configs[i] = ChainedSearchConfig{
			ReferenceColumn: "ref_id",
			TargetTable:     "table",
			TargetColumn:    "col",
			TargetType:      SearchParamString,
		}
	}

	_, _, _, err := MultiLevelChainedSearchClause(configs, "value", 1)
	if err == nil {
		t.Fatal("expected error for exceeding max chain depth, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds maximum") {
		t.Errorf("error = %q, want it to mention exceeds maximum", err.Error())
	}
}

func TestMultiLevelChainedSearchClause_EmptyConfigs(t *testing.T) {
	_, _, _, err := MultiLevelChainedSearchClause(nil, "value", 1)
	if err == nil {
		t.Fatal("expected error for empty configs, got nil")
	}
	if !strings.Contains(err.Error(), "no chain configs provided") {
		t.Errorf("error = %q, want it to mention no chain configs", err.Error())
	}
}

func TestMultiLevelChainedSearchClause_TwoLevelsTokenLeaf(t *testing.T) {
	configs := []ChainedSearchConfig{
		{
			ReferenceColumn: "patient_id",
			TargetTable:     "patients",
			TargetColumn:    "",
			TargetType:      SearchParamString,
		},
		{
			ReferenceColumn: "organization_id",
			TargetTable:     "organizations",
			TargetColumn:    "status",
			TargetType:      SearchParamToken,
		},
	}

	clause, args, nextIdx, err := MultiLevelChainedSearchClause(configs, "active", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "patient_id IN (SELECT id FROM patients WHERE organization_id IN (SELECT id FROM organizations WHERE status = $1))"
	if clause != expected {
		t.Errorf("clause = %q, want %q", clause, expected)
	}
	if len(args) != 1 || args[0] != "active" {
		t.Errorf("args = %v, want [active]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestMultiLevelChainedSearchClause_StartIdxOffset(t *testing.T) {
	configs := []ChainedSearchConfig{
		{
			ReferenceColumn: "encounter_id",
			TargetTable:     "encounters",
			TargetColumn:    "",
			TargetType:      SearchParamToken,
		},
		{
			ReferenceColumn: "patient_id",
			TargetTable:     "patients",
			TargetColumn:    "last_name",
			TargetType:      SearchParamString,
		},
	}

	clause, _, nextIdx, err := MultiLevelChainedSearchClause(configs, "Jones", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(clause, "$5") {
		t.Errorf("clause = %q, expected to use $5", clause)
	}
	if nextIdx != 6 {
		t.Errorf("nextIdx = %d, want 6", nextIdx)
	}
}

func TestMultiLevelChainedSearchClause_DateLeaf(t *testing.T) {
	configs := []ChainedSearchConfig{
		{
			ReferenceColumn: "patient_id",
			TargetTable:     "patients",
			TargetColumn:    "",
			TargetType:      SearchParamReference,
		},
		{
			ReferenceColumn: "encounter_id",
			TargetTable:     "encounters",
			TargetColumn:    "period_start",
			TargetType:      SearchParamDate,
		},
	}

	clause, args, nextIdx, err := MultiLevelChainedSearchClause(configs, "gt2023-06-01", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(clause, "patient_id IN (SELECT id FROM patients WHERE encounter_id IN (SELECT id FROM encounters WHERE period_start >") {
		t.Errorf("clause = %q, expected nested date subquery", clause)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestMultiLevelChainedSearchClause_TokenWithSystemLeaf(t *testing.T) {
	configs := []ChainedSearchConfig{
		{
			ReferenceColumn: "patient_id",
			TargetTable:     "patients",
			TargetColumn:    "",
			TargetType:      SearchParamReference,
		},
		{
			ReferenceColumn: "org_id",
			TargetTable:     "organizations",
			TargetColumn:    "identifier_value",
			TargetType:      SearchParamToken,
			TargetSysColumn: "identifier_system",
		},
	}

	clause, args, nextIdx, err := MultiLevelChainedSearchClause(configs, "http://npi.org|1234567890", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(clause, "identifier_system = $1") || !strings.Contains(clause, "identifier_value = $2") {
		t.Errorf("clause = %q, expected system|code token search", clause)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != "http://npi.org" || args[1] != "1234567890" {
		t.Errorf("args = %v, want [http://npi.org 1234567890]", args)
	}
	if nextIdx != 3 {
		t.Errorf("nextIdx = %d, want 3", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// ReverseChainClause tests
// ---------------------------------------------------------------------------

func TestReverseChainClause_TokenSearch(t *testing.T) {
	clause, args, nextIdx := ReverseChainClause(
		"patients", "id", "observations", "patient_id", "code",
		SearchParamToken, "1234-5", 1,
	)
	expected := "id IN (SELECT patient_id FROM observations WHERE code = $1)"
	if clause != expected {
		t.Errorf("clause = %q, want %q", clause, expected)
	}
	if len(args) != 1 || args[0] != "1234-5" {
		t.Errorf("args = %v, want [1234-5]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestReverseChainClause_StringSearch(t *testing.T) {
	clause, args, nextIdx := ReverseChainClause(
		"patients", "id", "conditions", "patient_id", "note",
		SearchParamString, "headache", 1,
	)
	expected := "id IN (SELECT patient_id FROM conditions WHERE note ILIKE $1)"
	if clause != expected {
		t.Errorf("clause = %q, want %q", clause, expected)
	}
	if len(args) != 1 || args[0] != "headache%" {
		t.Errorf("args = %v, want [headache%%]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestReverseChainClause_DateSearch(t *testing.T) {
	clause, args, nextIdx := ReverseChainClause(
		"patients", "id", "encounters", "patient_id", "period_start",
		SearchParamDate, "gt2023-01-01", 1,
	)
	if !strings.Contains(clause, "id IN (SELECT patient_id FROM encounters WHERE period_start >") {
		t.Errorf("clause = %q, expected reverse chain with date >", clause)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestReverseChainClause_NumberSearch(t *testing.T) {
	clause, args, nextIdx := ReverseChainClause(
		"patients", "id", "observations", "patient_id", "value_quantity",
		SearchParamNumber, "le100", 1,
	)
	if !strings.Contains(clause, "id IN (SELECT patient_id FROM observations WHERE value_quantity <=") {
		t.Errorf("clause = %q, expected reverse chain with number <=", clause)
	}
	if len(args) != 1 || args[0] != "100" {
		t.Errorf("args = %v, want [100]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestReverseChainClause_DefaultExactMatch(t *testing.T) {
	clause, args, nextIdx := ReverseChainClause(
		"patients", "id", "observations", "patient_id", "url",
		SearchParamURI, "http://example.com/obs", 1,
	)
	expected := "id IN (SELECT patient_id FROM observations WHERE url = $1)"
	if clause != expected {
		t.Errorf("clause = %q, want %q", clause, expected)
	}
	if len(args) != 1 || args[0] != "http://example.com/obs" {
		t.Errorf("args = %v", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestReverseChainClause_StartIdxOffset(t *testing.T) {
	clause, _, nextIdx := ReverseChainClause(
		"patients", "id", "observations", "patient_id", "code",
		SearchParamToken, "xyz", 10,
	)
	if !strings.Contains(clause, "$10") {
		t.Errorf("clause = %q, expected to use $10", clause)
	}
	if nextIdx != 11 {
		t.Errorf("nextIdx = %d, want 11", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// ReverseChainTokenClause tests
// ---------------------------------------------------------------------------

func TestReverseChainTokenClause_SystemAndCode(t *testing.T) {
	clause, args, nextIdx := ReverseChainTokenClause(
		"id", "observations", "patient_id", "code_system", "code",
		"http://loinc.org|8867-4", 1,
	)
	expected := "id IN (SELECT patient_id FROM observations WHERE (code_system = $1 AND code = $2))"
	if clause != expected {
		t.Errorf("clause = %q, want %q", clause, expected)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != "http://loinc.org" || args[1] != "8867-4" {
		t.Errorf("args = %v, want [http://loinc.org 8867-4]", args)
	}
	if nextIdx != 3 {
		t.Errorf("nextIdx = %d, want 3", nextIdx)
	}
}

func TestReverseChainTokenClause_CodeOnly(t *testing.T) {
	clause, args, nextIdx := ReverseChainTokenClause(
		"id", "observations", "patient_id", "code_system", "code",
		"8867-4", 1,
	)
	expected := "id IN (SELECT patient_id FROM observations WHERE code = $1)"
	if clause != expected {
		t.Errorf("clause = %q, want %q", clause, expected)
	}
	if len(args) != 1 || args[0] != "8867-4" {
		t.Errorf("args = %v, want [8867-4]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestReverseChainTokenClause_SystemOnly(t *testing.T) {
	clause, args, nextIdx := ReverseChainTokenClause(
		"id", "observations", "patient_id", "code_system", "code",
		"http://loinc.org|", 1,
	)
	expected := "id IN (SELECT patient_id FROM observations WHERE code_system = $1)"
	if clause != expected {
		t.Errorf("clause = %q, want %q", clause, expected)
	}
	if len(args) != 1 || args[0] != "http://loinc.org" {
		t.Errorf("args = %v, want [http://loinc.org]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestReverseChainTokenClause_PipeCodeOnly(t *testing.T) {
	clause, args, nextIdx := ReverseChainTokenClause(
		"id", "observations", "patient_id", "code_system", "code",
		"|8867-4", 1,
	)
	expected := "id IN (SELECT patient_id FROM observations WHERE code = $1)"
	if clause != expected {
		t.Errorf("clause = %q, want %q", clause, expected)
	}
	if len(args) != 1 || args[0] != "8867-4" {
		t.Errorf("args = %v, want [8867-4]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// DefaultChainRegistry tests
// ---------------------------------------------------------------------------

func TestDefaultChainRegistry_NotNil(t *testing.T) {
	r := DefaultChainRegistry()
	if r == nil {
		t.Fatal("DefaultChainRegistry returned nil")
	}
}

func TestDefaultChainRegistry_PatientPractitionerName(t *testing.T) {
	r := DefaultChainRegistry()
	config, err := r.Resolve("Patient", "general-practitioner.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TargetTable != "practitioners" {
		t.Errorf("TargetTable = %q, want 'practitioners'", config.TargetTable)
	}
	if config.ReferenceColumn != "practitioner_id" {
		t.Errorf("ReferenceColumn = %q, want 'practitioner_id'", config.ReferenceColumn)
	}
	if config.TargetType != SearchParamString {
		t.Errorf("TargetType = %d, want SearchParamString", config.TargetType)
	}
}

func TestDefaultChainRegistry_PatientPractitionerIdentifier(t *testing.T) {
	r := DefaultChainRegistry()
	config, err := r.Resolve("Patient", "general-practitioner.identifier")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TargetType != SearchParamToken {
		t.Errorf("TargetType = %d, want SearchParamToken", config.TargetType)
	}
	if config.TargetSysColumn != "identifier_system" {
		t.Errorf("TargetSysColumn = %q, want 'identifier_system'", config.TargetSysColumn)
	}
}

func TestDefaultChainRegistry_ObservationPatientName(t *testing.T) {
	r := DefaultChainRegistry()
	config, err := r.Resolve("Observation", "patient.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TargetTable != "patients" {
		t.Errorf("TargetTable = %q, want 'patients'", config.TargetTable)
	}
	if config.TargetColumn != "last_name" {
		t.Errorf("TargetColumn = %q, want 'last_name'", config.TargetColumn)
	}
	if config.ReferenceColumn != "patient_id" {
		t.Errorf("ReferenceColumn = %q, want 'patient_id'", config.ReferenceColumn)
	}
}

func TestDefaultChainRegistry_ObservationPatientBirthdate(t *testing.T) {
	r := DefaultChainRegistry()
	config, err := r.Resolve("Observation", "patient.birthdate")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TargetType != SearchParamDate {
		t.Errorf("TargetType = %d, want SearchParamDate", config.TargetType)
	}
	if config.TargetColumn != "birth_date" {
		t.Errorf("TargetColumn = %q, want 'birth_date'", config.TargetColumn)
	}
}

func TestDefaultChainRegistry_ObservationEncounterStatus(t *testing.T) {
	r := DefaultChainRegistry()
	config, err := r.Resolve("Observation", "encounter.status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TargetTable != "encounters" {
		t.Errorf("TargetTable = %q, want 'encounters'", config.TargetTable)
	}
	if config.TargetColumn != "status" {
		t.Errorf("TargetColumn = %q, want 'status'", config.TargetColumn)
	}
}

func TestDefaultChainRegistry_ObservationEncounterDate(t *testing.T) {
	r := DefaultChainRegistry()
	config, err := r.Resolve("Observation", "encounter.date")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TargetType != SearchParamDate {
		t.Errorf("TargetType = %d, want SearchParamDate", config.TargetType)
	}
}

func TestDefaultChainRegistry_EncounterPatientName(t *testing.T) {
	r := DefaultChainRegistry()
	config, err := r.Resolve("Encounter", "patient.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TargetTable != "patients" {
		t.Errorf("TargetTable = %q, want 'patients'", config.TargetTable)
	}
}

func TestDefaultChainRegistry_MedicationRequestPatientName(t *testing.T) {
	r := DefaultChainRegistry()
	config, err := r.Resolve("MedicationRequest", "patient.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TargetTable != "patients" {
		t.Errorf("TargetTable = %q, want 'patients'", config.TargetTable)
	}
	if config.ReferenceColumn != "patient_id" {
		t.Errorf("ReferenceColumn = %q, want 'patient_id'", config.ReferenceColumn)
	}
}

func TestDefaultChainRegistry_ConditionPatientName(t *testing.T) {
	r := DefaultChainRegistry()
	config, err := r.Resolve("Condition", "patient.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TargetTable != "patients" {
		t.Errorf("TargetTable = %q, want 'patients'", config.TargetTable)
	}
}

func TestDefaultChainRegistry_ProcedurePatientName(t *testing.T) {
	r := DefaultChainRegistry()
	config, err := r.Resolve("Procedure", "patient.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TargetTable != "patients" {
		t.Errorf("TargetTable = %q, want 'patients'", config.TargetTable)
	}
}

func TestDefaultChainRegistry_ProcedurePatientIdentifier(t *testing.T) {
	r := DefaultChainRegistry()
	config, err := r.Resolve("Procedure", "patient.identifier")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TargetType != SearchParamToken {
		t.Errorf("TargetType = %d, want SearchParamToken", config.TargetType)
	}
	if config.TargetSysColumn != "identifier_system" {
		t.Errorf("TargetSysColumn = %q, want 'identifier_system'", config.TargetSysColumn)
	}
}

func TestDefaultChainRegistry_UnknownPath(t *testing.T) {
	r := DefaultChainRegistry()
	_, err := r.Resolve("Patient", "unknown.field")
	if err == nil {
		t.Fatal("expected error for unknown chain path in default registry")
	}
}

func TestDefaultChainRegistry_MedicationRequestPatientIdentifier(t *testing.T) {
	r := DefaultChainRegistry()
	config, err := r.Resolve("MedicationRequest", "patient.identifier")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TargetType != SearchParamToken {
		t.Errorf("TargetType = %d, want SearchParamToken", config.TargetType)
	}
}

func TestDefaultChainRegistry_ConditionPatientIdentifier(t *testing.T) {
	r := DefaultChainRegistry()
	config, err := r.Resolve("Condition", "patient.identifier")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TargetType != SearchParamToken {
		t.Errorf("TargetType = %d, want SearchParamToken", config.TargetType)
	}
}

func TestDefaultChainRegistry_ObservationPatientIdentifier(t *testing.T) {
	r := DefaultChainRegistry()
	config, err := r.Resolve("Observation", "patient.identifier")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TargetType != SearchParamToken {
		t.Errorf("TargetType = %d, want SearchParamToken", config.TargetType)
	}
}

func TestDefaultChainRegistry_EncounterPatientIdentifier(t *testing.T) {
	r := DefaultChainRegistry()
	config, err := r.Resolve("Encounter", "patient.identifier")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TargetType != SearchParamToken {
		t.Errorf("TargetType = %d, want SearchParamToken", config.TargetType)
	}
}

// ---------------------------------------------------------------------------
// MaxChainDepth constant test
// ---------------------------------------------------------------------------

func TestMaxChainDepth(t *testing.T) {
	if MaxChainDepth != 3 {
		t.Errorf("MaxChainDepth = %d, want 3 (per FHIR spec)", MaxChainDepth)
	}
}

// ---------------------------------------------------------------------------
// Integration-style: chained search + SearchQuery
// ---------------------------------------------------------------------------

func TestChainedSearchClause_IntegrationWithSearchQuery(t *testing.T) {
	// Simulate: Observation?patient.name=Smith
	// Build a search query and add the chained clause to it.
	q := NewSearchQuery("observations", "id, code, patient_id")

	config := ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "last_name",
		TargetType:      SearchParamString,
	}

	clause, args, nextIdx := ChainedSearchClause(config, "Smith", q.Idx())
	q.Add(clause, args...)

	countSQL := q.CountSQL()
	if !strings.Contains(countSQL, "patient_id IN (SELECT id FROM patients WHERE last_name ILIKE $1)") {
		t.Errorf("count SQL = %q, expected chained subquery", countSQL)
	}
	if len(q.CountArgs()) != 1 {
		t.Errorf("expected 1 count arg, got %d", len(q.CountArgs()))
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestReverseChainClause_IntegrationWithSearchQuery(t *testing.T) {
	// Simulate: Patient?_has:Observation:patient:code=1234-5
	q := NewSearchQuery("patients", "id, last_name")

	clause, args, _ := ReverseChainClause(
		"patients", "id", "observations", "patient_id", "code",
		SearchParamToken, "1234-5", q.Idx(),
	)
	q.Add(clause, args...)

	countSQL := q.CountSQL()
	if !strings.Contains(countSQL, "id IN (SELECT patient_id FROM observations WHERE code = $1)") {
		t.Errorf("count SQL = %q, expected reverse chain subquery", countSQL)
	}
	if len(q.CountArgs()) != 1 || q.CountArgs()[0] != "1234-5" {
		t.Errorf("count args = %v, want [1234-5]", q.CountArgs())
	}
}
