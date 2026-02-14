package fhir

import (
	"context"
	"strings"
	"testing"
)

func TestParseChainedParam(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		valid    bool
		source   string
		target   string
		tParam   string
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

func TestParseHasParam(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		valid   bool
		tType   string
		tParam  string
		sParam  string
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
