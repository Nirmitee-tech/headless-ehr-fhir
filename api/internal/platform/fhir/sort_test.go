package fhir

import (
	"testing"
)

func TestParseSort(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []SortSpec
	}{
		{"empty", "", nil},
		{"single asc", "date", []SortSpec{{Field: "date", Descending: false}}},
		{"single desc", "-date", []SortSpec{{Field: "date", Descending: true}}},
		{"multiple", "-date,status", []SortSpec{
			{Field: "date", Descending: true},
			{Field: "status", Descending: false},
		}},
		{"with spaces", " -date , status ", []SortSpec{
			{Field: "date", Descending: true},
			{Field: "status", Descending: false},
		}},
		{"three fields", "name,-date,status", []SortSpec{
			{Field: "name", Descending: false},
			{Field: "date", Descending: true},
			{Field: "status", Descending: false},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseSort(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("ParseSort(%q) returned %d specs, want %d", tt.input, len(result), len(tt.expected))
			}
			for i, spec := range result {
				if spec.Field != tt.expected[i].Field {
					t.Errorf("spec[%d].Field = %q, want %q", i, spec.Field, tt.expected[i].Field)
				}
				if spec.Descending != tt.expected[i].Descending {
					t.Errorf("spec[%d].Descending = %v, want %v", i, spec.Descending, tt.expected[i].Descending)
				}
			}
		})
	}
}

func TestBuildOrderClause(t *testing.T) {
	fieldMap := map[string]string{
		"date":   "created_at",
		"status": "status",
		"name":   "family_name",
	}

	tests := []struct {
		name         string
		specs        []SortSpec
		defaultOrder string
		expected     string
	}{
		{"empty specs with default", nil, "created_at DESC", " ORDER BY created_at DESC"},
		{"empty specs no default", nil, "", ""},
		{"single asc", []SortSpec{{Field: "date"}}, "", " ORDER BY created_at ASC"},
		{"single desc", []SortSpec{{Field: "date", Descending: true}}, "", " ORDER BY created_at DESC"},
		{"multiple", []SortSpec{
			{Field: "date", Descending: true},
			{Field: "status"},
		}, "", " ORDER BY created_at DESC, status ASC"},
		{"unknown field falls through", []SortSpec{{Field: "unknown"}}, "created_at DESC", " ORDER BY created_at DESC"},
		{"mixed known and unknown", []SortSpec{
			{Field: "date", Descending: true},
			{Field: "unknown"},
			{Field: "name"},
		}, "", " ORDER BY created_at DESC, family_name ASC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildOrderClause(tt.specs, fieldMap, tt.defaultOrder)
			if result != tt.expected {
				t.Errorf("BuildOrderClause() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildOrderClauseNullsLast(t *testing.T) {
	fieldMap := map[string]string{
		"date": "effective_datetime",
	}

	specs := []SortSpec{{Field: "date", Descending: true}}
	result := BuildOrderClauseNullsLast(specs, fieldMap, "")
	expected := " ORDER BY effective_datetime DESC NULLS LAST"
	if result != expected {
		t.Errorf("BuildOrderClauseNullsLast() = %q, want %q", result, expected)
	}
}

func TestBuildOrderClauseNullsLast_Comprehensive(t *testing.T) {
	fieldMap := map[string]string{
		"date":   "effective_datetime",
		"status": "status",
		"name":   "family_name",
	}

	tests := []struct {
		name         string
		specs        []SortSpec
		defaultOrder string
		expected     string
	}{
		{
			"desc field with NULLS LAST",
			[]SortSpec{{Field: "date", Descending: true}},
			"",
			" ORDER BY effective_datetime DESC NULLS LAST",
		},
		{
			"asc field without NULLS LAST",
			[]SortSpec{{Field: "date", Descending: false}},
			"",
			" ORDER BY effective_datetime ASC",
		},
		{
			"mixed asc and desc",
			[]SortSpec{
				{Field: "date", Descending: true},
				{Field: "status", Descending: false},
				{Field: "name", Descending: true},
			},
			"",
			" ORDER BY effective_datetime DESC NULLS LAST, status ASC, family_name DESC NULLS LAST",
		},
		{
			"no matching fields with default",
			[]SortSpec{{Field: "unknown"}},
			"created_at DESC",
			" ORDER BY created_at DESC",
		},
		{
			"no matching fields without default",
			[]SortSpec{{Field: "unknown"}},
			"",
			"",
		},
		{
			"empty specs with default",
			nil,
			"created_at DESC",
			" ORDER BY created_at DESC",
		},
		{
			"empty specs without default",
			nil,
			"",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildOrderClauseNullsLast(tt.specs, fieldMap, tt.defaultOrder)
			if result != tt.expected {
				t.Errorf("BuildOrderClauseNullsLast() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseSort_EmptyFieldAfterComma(t *testing.T) {
	// "date,,status" should skip the empty field in the middle
	specs := ParseSort("date,,status")
	if len(specs) != 2 {
		t.Fatalf("expected 2 specs, got %d", len(specs))
	}
	if specs[0].Field != "date" {
		t.Errorf("expected first field 'date', got %q", specs[0].Field)
	}
	if specs[0].Descending {
		t.Error("expected first field ASC")
	}
	if specs[1].Field != "status" {
		t.Errorf("expected second field 'status', got %q", specs[1].Field)
	}
	if specs[1].Descending {
		t.Error("expected second field ASC")
	}
}

func TestParseSort_BareDash(t *testing.T) {
	// A bare "-" should produce an empty field which is skipped
	specs := ParseSort("-")
	if len(specs) != 0 {
		t.Errorf("expected 0 specs for bare dash, got %d", len(specs))
	}
}

func TestParseSort_CommaOnly(t *testing.T) {
	specs := ParseSort(",")
	if len(specs) != 0 {
		t.Errorf("expected 0 specs for comma-only input, got %d", len(specs))
	}
}

func TestBuildOrderClause_AllUnknownFieldsNoDefault(t *testing.T) {
	fieldMap := map[string]string{"date": "created_at"}
	specs := []SortSpec{
		{Field: "unknown1"},
		{Field: "unknown2"},
	}
	result := BuildOrderClause(specs, fieldMap, "")
	if result != "" {
		t.Errorf("expected empty string when all fields unknown and no default, got %q", result)
	}
}

func TestBuildOrderClauseNullsLast_ASCField(t *testing.T) {
	fieldMap := map[string]string{"name": "family_name"}
	specs := []SortSpec{{Field: "name", Descending: false}}
	result := BuildOrderClauseNullsLast(specs, fieldMap, "")
	expected := " ORDER BY family_name ASC"
	if result != expected {
		t.Errorf("BuildOrderClauseNullsLast ASC = %q, want %q", result, expected)
	}
}

func TestBuildOrderClause_SingleFieldNonDefaultIdx(t *testing.T) {
	fieldMap := map[string]string{"name": "family_name", "date": "created_at"}
	specs := []SortSpec{{Field: "name", Descending: false}}
	result := BuildOrderClause(specs, fieldMap, "")
	expected := " ORDER BY family_name ASC"
	if result != expected {
		t.Errorf("BuildOrderClause = %q, want %q", result, expected)
	}
}
