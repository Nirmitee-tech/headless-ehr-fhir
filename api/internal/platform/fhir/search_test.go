package fhir

import (
	"testing"
	"time"
)

func TestParseSearchValue(t *testing.T) {
	tests := []struct {
		input    string
		prefix   SearchPrefix
		value    string
	}{
		{"2023-01-01", PrefixEq, "2023-01-01"},
		{"gt2023-01-01", PrefixGt, "2023-01-01"},
		{"lt2023-12-31", PrefixLt, "2023-12-31"},
		{"ge100", PrefixGe, "100"},
		{"le200", PrefixLe, "200"},
		{"ne50", PrefixNe, "50"},
		{"sa2023-06-01", PrefixSa, "2023-06-01"},
		{"eb2023-06-30", PrefixEb, "2023-06-30"},
		{"ap2023-06-15", PrefixAp, "2023-06-15"},
		{"eq2023-01-01", PrefixEq, "2023-01-01"},
		{"abc", PrefixEq, "abc"},
		{"", PrefixEq, ""},
		{"g", PrefixEq, "g"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseSearchValue(tt.input)
			if result.Prefix != tt.prefix {
				t.Errorf("ParseSearchValue(%q).Prefix = %q, want %q", tt.input, result.Prefix, tt.prefix)
			}
			if result.Value != tt.value {
				t.Errorf("ParseSearchValue(%q).Value = %q, want %q", tt.input, result.Value, tt.value)
			}
		})
	}
}

func TestParseParamModifier(t *testing.T) {
	tests := []struct {
		input    string
		param    string
		modifier SearchModifier
	}{
		{"name:exact", "name", ModifierExact},
		{"name:contains", "name", ModifierContains},
		{"code:not", "code", ModifierNot},
		{"name", "name", ""},
		{"status:above", "status", ModifierAbove},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			param, mod := ParseParamModifier(tt.input)
			if param != tt.param {
				t.Errorf("ParseParamModifier(%q) param = %q, want %q", tt.input, param, tt.param)
			}
			if mod != tt.modifier {
				t.Errorf("ParseParamModifier(%q) modifier = %q, want %q", tt.input, mod, tt.modifier)
			}
		})
	}
}

func TestDateSearchClause(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		wantSQL  string
		wantArgs int
	}{
		{"exact date", "2023-01-15", "(effective_datetime >= $1 AND effective_datetime <= $2)", 2},
		{"gt prefix", "gt2023-01-15", "effective_datetime > $1", 1},
		{"lt prefix", "lt2023-01-15", "effective_datetime < $1", 1},
		{"ge prefix", "ge2023-01-15", "effective_datetime >= $1", 1},
		{"le prefix", "le2023-01-15", "effective_datetime <= $1", 1},
		{"ne prefix", "ne2023-01-15", "effective_datetime != $1", 1},
		{"ap prefix", "ap2023-01-15", "(effective_datetime >= $1 AND effective_datetime <= $2)", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clause, args, _ := DateSearchClause("effective_datetime", tt.value, 1)
			if clause != tt.wantSQL {
				t.Errorf("DateSearchClause(%q) clause = %q, want %q", tt.value, clause, tt.wantSQL)
			}
			if len(args) != tt.wantArgs {
				t.Errorf("DateSearchClause(%q) args count = %d, want %d", tt.value, len(args), tt.wantArgs)
			}
		})
	}
}

func TestNumberSearchClause(t *testing.T) {
	tests := []struct {
		value   string
		wantSQL string
	}{
		{"100", "value = $1"},
		{"gt100", "value > $1"},
		{"lt50", "value < $1"},
		{"ge10", "value >= $1"},
		{"le200", "value <= $1"},
		{"ne0", "value != $1"},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			clause, _, _ := NumberSearchClause("value", tt.value, 1)
			if clause != tt.wantSQL {
				t.Errorf("NumberSearchClause(%q) = %q, want %q", tt.value, clause, tt.wantSQL)
			}
		})
	}
}

func TestTokenSearchClause(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantSQL string
		args    int
	}{
		{"code only", "1234", "code_value = $1", 1},
		{"system|code", "http://loinc.org|1234", "(code_system = $1 AND code_value = $2)", 2},
		{"|code", "|1234", "code_value = $1", 1},
		{"system|", "http://loinc.org|", "code_system = $1", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clause, args, _ := TokenSearchClause("code_system", "code_value", tt.value, 1)
			if clause != tt.wantSQL {
				t.Errorf("TokenSearchClause(%q) = %q, want %q", tt.value, clause, tt.wantSQL)
			}
			if len(args) != tt.args {
				t.Errorf("TokenSearchClause(%q) args = %d, want %d", tt.value, len(args), tt.args)
			}
		})
	}
}

func TestStringSearchClause(t *testing.T) {
	tests := []struct {
		value    string
		modifier SearchModifier
		wantSQL  string
	}{
		{"John", "", "name ILIKE $1"},              // default prefix match
		{"John", ModifierExact, "name = $1"},        // exact
		{"ohn", ModifierContains, "name ILIKE $1"},  // contains
	}

	for _, tt := range tests {
		t.Run(string(tt.modifier), func(t *testing.T) {
			clause, _, _ := StringSearchClause("name", tt.value, tt.modifier, 1)
			if clause != tt.wantSQL {
				t.Errorf("StringSearchClause modifier=%q: got %q, want %q", tt.modifier, clause, tt.wantSQL)
			}
		})
	}
}

func TestReferenceSearchClause(t *testing.T) {
	tests := []struct {
		value   string
		wantArg string
	}{
		{"Patient/123", "123"},
		{"123", "123"},
		{"Organization/abc-def", "abc-def"},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			_, args, _ := ReferenceSearchClause("patient_id", tt.value, 1)
			if args[0].(string) != tt.wantArg {
				t.Errorf("ReferenceSearchClause(%q) arg = %q, want %q", tt.value, args[0], tt.wantArg)
			}
		})
	}
}

func TestParseFlexDate(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"2023-01-15", true},
		{"2023-01-15T10:30:00Z", true},
		{"2023-01-15T10:30:00", true},
		{"2023-01", true},
		{"2023", true},
		{"not-a-date", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := parseFlexDate(tt.input)
			if tt.valid && err != nil {
				t.Errorf("parseFlexDate(%q) returned error: %v", tt.input, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("parseFlexDate(%q) should have returned error", tt.input)
			}
		})
	}
}

func TestDateSearchClauseArgTypes(t *testing.T) {
	clause, args, nextIdx := DateSearchClause("created_at", "gt2023-06-15", 1)
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
	if clause != "created_at > $1" {
		t.Errorf("clause = %q, want %q", clause, "created_at > $1")
	}
	if len(args) != 1 {
		t.Fatalf("args length = %d, want 1", len(args))
	}
	if _, ok := args[0].(time.Time); !ok {
		t.Errorf("arg[0] should be time.Time, got %T", args[0])
	}
}

func TestDateSearchClause_ApproximatePrefix(t *testing.T) {
	clause, args, nextIdx := DateSearchClause("effective_date", "ap2023-06-15", 1)
	wantClause := "(effective_date >= $1 AND effective_date <= $2)"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args for approximate search, got %d", len(args))
	}
	if nextIdx != 3 {
		t.Errorf("nextIdx = %d, want 3", nextIdx)
	}

	low, ok := args[0].(time.Time)
	if !ok {
		t.Fatalf("arg[0] should be time.Time, got %T", args[0])
	}
	high, ok := args[1].(time.Time)
	if !ok {
		t.Fatalf("arg[1] should be time.Time, got %T", args[1])
	}

	// The range should be +/- 1 day from the parsed date
	target, _ := time.Parse("2006-01-02", "2023-06-15")
	expectedLow := target.Add(-24 * time.Hour)
	expectedHigh := target.Add(24 * time.Hour)
	if !low.Equal(expectedLow) {
		t.Errorf("low bound = %v, want %v", low, expectedLow)
	}
	if !high.Equal(expectedHigh) {
		t.Errorf("high bound = %v, want %v", high, expectedHigh)
	}
}

func TestDateSearchClause_ExactDatetime(t *testing.T) {
	// An exact datetime (not just date) should produce an equality clause
	clause, args, nextIdx := DateSearchClause("effective_date", "2023-06-15T10:30:00Z", 1)
	wantClause := "effective_date = $1"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg for exact datetime, got %d", len(args))
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
	if _, ok := args[0].(time.Time); !ok {
		t.Errorf("arg[0] should be time.Time, got %T", args[0])
	}
}

func TestDateSearchClause_UnparseableDate(t *testing.T) {
	// A value that cannot be parsed by parseFlexDate should fall back to text match
	clause, args, nextIdx := DateSearchClause("effective_date", "not-a-real-date", 1)
	wantClause := "effective_date::text = $1"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg for fallback, got %d", len(args))
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
	if args[0] != "not-a-real-date" {
		t.Errorf("arg[0] = %v, want 'not-a-real-date'", args[0])
	}
}

func TestStringSearchClause_TextModifier(t *testing.T) {
	clause, args, nextIdx := StringSearchClause("description", "headache", ModifierText, 1)
	wantClause := "description ILIKE $1"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
	// The text modifier should produce a contains-style pattern (%value%)
	wantArg := "%headache%"
	if args[0] != wantArg {
		t.Errorf("arg[0] = %v, want %q", args[0], wantArg)
	}
}

func TestNumberSearchClause_SaAndEbPrefixes(t *testing.T) {
	// "sa" (starts after) prefix should behave like "gt"
	clause, args, nextIdx := NumberSearchClause("value", "sa100", 1)
	if clause != "value > $1" {
		t.Errorf("sa clause = %q, want %q", clause, "value > $1")
	}
	if len(args) != 1 || args[0] != "100" {
		t.Errorf("sa args = %v, want [100]", args)
	}
	if nextIdx != 2 {
		t.Errorf("sa nextIdx = %d, want 2", nextIdx)
	}

	// "eb" (ends before) prefix should behave like "lt"
	clause, args, nextIdx = NumberSearchClause("value", "eb50", 1)
	if clause != "value < $1" {
		t.Errorf("eb clause = %q, want %q", clause, "value < $1")
	}
	if len(args) != 1 || args[0] != "50" {
		t.Errorf("eb args = %v, want [50]", args)
	}
	if nextIdx != 2 {
		t.Errorf("eb nextIdx = %d, want 2", nextIdx)
	}
}

func TestNumberSearchClause_ArgIdxAdvancement(t *testing.T) {
	// Verify correct argIdx advancement starting from a non-1 index
	clause, _, nextIdx := NumberSearchClause("amount", "ge500", 5)
	if clause != "amount >= $5" {
		t.Errorf("clause = %q, want %q", clause, "amount >= $5")
	}
	if nextIdx != 6 {
		t.Errorf("nextIdx = %d, want 6", nextIdx)
	}
}

func TestTokenSearchClause_EmptyPipeValue(t *testing.T) {
	// "|" with empty system and empty code should fall through to code-only match
	clause, args, nextIdx := TokenSearchClause("system", "code", "|", 1)
	if clause != "code = $1" {
		t.Errorf("clause = %q, want %q", clause, "code = $1")
	}
	if len(args) != 1 || args[0] != "|" {
		// When both system and code are empty, no pipe branches match,
		// so it falls through to the no-pipe case
		t.Errorf("args = %v, expected fallthrough to no-pipe behavior", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestStringSearchClause_DefaultPrefixMatch(t *testing.T) {
	// Default modifier should produce prefix ILIKE with trailing %
	clause, args, nextIdx := StringSearchClause("name", "Joh", "", 3)
	if clause != "name ILIKE $3" {
		t.Errorf("clause = %q, want %q", clause, "name ILIKE $3")
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if args[0] != "Joh%" {
		t.Errorf("arg = %v, want %q", args[0], "Joh%")
	}
	if nextIdx != 4 {
		t.Errorf("nextIdx = %d, want 4", nextIdx)
	}
}

func TestStringSearchClause_ContainsPattern(t *testing.T) {
	_, args, _ := StringSearchClause("name", "ohn", ModifierContains, 1)
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if args[0] != "%ohn%" {
		t.Errorf("contains arg = %v, want %q", args[0], "%ohn%")
	}
}

func TestDateSearchClause_SaPrefix(t *testing.T) {
	clause, args, nextIdx := DateSearchClause("date_col", "sa2023-06-15", 1)
	if clause != "date_col > $1" {
		t.Errorf("clause = %q, want %q", clause, "date_col > $1")
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if _, ok := args[0].(time.Time); !ok {
		t.Errorf("arg[0] should be time.Time, got %T", args[0])
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestDateSearchClause_EbPrefix(t *testing.T) {
	clause, args, nextIdx := DateSearchClause("date_col", "eb2023-12-31", 1)
	if clause != "date_col < $1" {
		t.Errorf("clause = %q, want %q", clause, "date_col < $1")
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if _, ok := args[0].(time.Time); !ok {
		t.Errorf("arg[0] should be time.Time, got %T", args[0])
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestDateSearchClause_YearOnlyFormat(t *testing.T) {
	// Year-only "2023" should parse and since len != 10, produce equality
	clause, args, nextIdx := DateSearchClause("date_col", "2023", 1)
	if clause != "date_col = $1" {
		t.Errorf("clause = %q, want %q", clause, "date_col = $1")
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if _, ok := args[0].(time.Time); !ok {
		t.Errorf("arg[0] should be time.Time, got %T", args[0])
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestDateSearchClause_MonthOnlyFormat(t *testing.T) {
	clause, args, nextIdx := DateSearchClause("date_col", "2023-06", 1)
	if clause != "date_col = $1" {
		t.Errorf("clause = %q, want %q", clause, "date_col = $1")
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if _, ok := args[0].(time.Time); !ok {
		t.Errorf("arg[0] should be time.Time, got %T", args[0])
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestReferenceSearchClause_SQLFormat(t *testing.T) {
	clause, args, nextIdx := ReferenceSearchClause("subject_id", "Patient/abc-123", 3)
	if clause != "subject_id = $3" {
		t.Errorf("clause = %q, want %q", clause, "subject_id = $3")
	}
	if len(args) != 1 || args[0] != "abc-123" {
		t.Errorf("args = %v, want [abc-123]", args)
	}
	if nextIdx != 4 {
		t.Errorf("nextIdx = %d, want 4", nextIdx)
	}
}

func TestReferenceSearchClause_NestedSlashes(t *testing.T) {
	// Reference with multiple slashes, e.g., "http://example.org/fhir/Patient/123"
	clause, args, nextIdx := ReferenceSearchClause("ref_col", "http://example.org/fhir/Patient/123", 1)
	if clause != "ref_col = $1" {
		t.Errorf("clause = %q, want %q", clause, "ref_col = $1")
	}
	// LastIndex should extract just "123"
	if len(args) != 1 || args[0] != "123" {
		t.Errorf("args = %v, want [123]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestParseParamModifier_MultipleColons(t *testing.T) {
	// "name:exact:extra" should split as param="name", modifier="exact:extra"
	param, mod := ParseParamModifier("name:exact:extra")
	if param != "name" {
		t.Errorf("param = %q, want %q", param, "name")
	}
	if mod != "exact:extra" {
		t.Errorf("modifier = %q, want %q", mod, "exact:extra")
	}
}

func TestParseSearchValue_UpperCasePrefix(t *testing.T) {
	// Prefixes are case-insensitive: "GT2023" should be parsed as PrefixGt
	result := ParseSearchValue("GT2023-01-01")
	if result.Prefix != PrefixGt {
		t.Errorf("prefix = %q, want %q", result.Prefix, PrefixGt)
	}
	if result.Value != "2023-01-01" {
		t.Errorf("value = %q, want %q", result.Value, "2023-01-01")
	}
}
