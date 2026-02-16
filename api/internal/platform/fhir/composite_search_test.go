package fhir

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// CompositeSearchClause — basic splitting and clause generation
// ---------------------------------------------------------------------------

func TestCompositeSearchClause_TokenAndQuantity(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value", SysColumn: "code_system"},
			{Name: "value", Type: SearchParamQuantity, Column: "value_quantity"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "http://loinc.org|8480-6$gt100", 1)

	// Token component: system|code produces 2 args
	// Quantity component: gt100 produces 1 arg
	if !strings.Contains(clause, "code_system = $1") {
		t.Errorf("clause should contain code_system = $1, got %q", clause)
	}
	if !strings.Contains(clause, "code_value = $2") {
		t.Errorf("clause should contain code_value = $2, got %q", clause)
	}
	if !strings.Contains(clause, "value_quantity > $3") {
		t.Errorf("clause should contain value_quantity > $3, got %q", clause)
	}
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d: %v", len(args), args)
	}
	if args[0] != "http://loinc.org" {
		t.Errorf("args[0] = %v, want %q", args[0], "http://loinc.org")
	}
	if args[1] != "8480-6" {
		t.Errorf("args[1] = %v, want %q", args[1], "8480-6")
	}
	if args[2] != "100" {
		t.Errorf("args[2] = %v, want %q", args[2], "100")
	}
	if nextIdx != 4 {
		t.Errorf("nextIdx = %d, want 4", nextIdx)
	}
}

func TestCompositeSearchClause_TokenAndToken(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value", SysColumn: "code_system"},
			{Name: "value", Type: SearchParamToken, Column: "value_code", SysColumn: "value_system"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "http://loinc.org|1234$http://snomed.info/sct|5678", 1)

	if !strings.Contains(clause, "code_system = $1") {
		t.Errorf("clause should contain code_system = $1, got %q", clause)
	}
	if !strings.Contains(clause, "code_value = $2") {
		t.Errorf("clause should contain code_value = $2, got %q", clause)
	}
	if !strings.Contains(clause, "value_system = $3") {
		t.Errorf("clause should contain value_system = $3, got %q", clause)
	}
	if !strings.Contains(clause, "value_code = $4") {
		t.Errorf("clause should contain value_code = $4, got %q", clause)
	}
	if len(args) != 4 {
		t.Fatalf("expected 4 args, got %d: %v", len(args), args)
	}
	if nextIdx != 5 {
		t.Errorf("nextIdx = %d, want 5", nextIdx)
	}
}

func TestCompositeSearchClause_TokenAndDate(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value", SysColumn: "code_system"},
			{Name: "value", Type: SearchParamDate, Column: "value_date"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "http://loinc.org|1234$ge2023-01-01", 1)

	if !strings.Contains(clause, "code_system = $1") {
		t.Errorf("clause should contain code_system = $1, got %q", clause)
	}
	if !strings.Contains(clause, "code_value = $2") {
		t.Errorf("clause should contain code_value = $2, got %q", clause)
	}
	if !strings.Contains(clause, "value_date >= $3") {
		t.Errorf("clause should contain value_date >= $3, got %q", clause)
	}
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d: %v", len(args), args)
	}
	if nextIdx != 4 {
		t.Errorf("nextIdx = %d, want 4", nextIdx)
	}
}

func TestCompositeSearchClause_TokenAndString(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value", SysColumn: "code_system"},
			{Name: "value", Type: SearchParamString, Column: "value_string"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "http://loinc.org|1234$positive", 1)

	if !strings.Contains(clause, "code_system = $1") {
		t.Errorf("clause should contain code_system = $1, got %q", clause)
	}
	if !strings.Contains(clause, "value_string ILIKE $3") {
		t.Errorf("clause should contain value_string ILIKE $3, got %q", clause)
	}
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d: %v", len(args), args)
	}
	// String search uses default prefix match: value + "%"
	if args[2] != "positive%" {
		t.Errorf("args[2] = %v, want %q", args[2], "positive%")
	}
	if nextIdx != 4 {
		t.Errorf("nextIdx = %d, want 4", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// Single component (degenerate case)
// ---------------------------------------------------------------------------

func TestCompositeSearchClause_SingleComponent(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value", SysColumn: "code_system"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "http://loinc.org|1234", 1)

	// Single component should not wrap in extra parentheses
	wantClause := "(code_system = $1 AND code_value = $2)"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if nextIdx != 3 {
		t.Errorf("nextIdx = %d, want 3", nextIdx)
	}
}

func TestCompositeSearchClause_SingleComponentCodeOnly(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value", SysColumn: "code_system"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "1234", 1)

	// Without pipe, token falls through to code-only match
	if clause != "code_value = $1" {
		t.Errorf("clause = %q, want %q", clause, "code_value = $1")
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// Dollar-sign splitting
// ---------------------------------------------------------------------------

func TestCompositeSearchClause_DollarSplitTwoComponents(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value"},
			{Name: "value", Type: SearchParamNumber, Column: "value_num"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "vital-signs$100", 1)

	if !strings.Contains(clause, "code_value = $1") {
		t.Errorf("clause should contain code_value = $1, got %q", clause)
	}
	if !strings.Contains(clause, "value_num = $2") {
		t.Errorf("clause should contain value_num = $2, got %q", clause)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(args), args)
	}
	if args[0] != "vital-signs" {
		t.Errorf("args[0] = %v, want %q", args[0], "vital-signs")
	}
	if args[1] != "100" {
		t.Errorf("args[1] = %v, want %q", args[1], "100")
	}
	if nextIdx != 3 {
		t.Errorf("nextIdx = %d, want 3", nextIdx)
	}
}

func TestCompositeSearchClause_DollarSplitThreeComponents(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "a", Type: SearchParamToken, Column: "col_a"},
			{Name: "b", Type: SearchParamToken, Column: "col_b"},
			{Name: "c", Type: SearchParamToken, Column: "col_c"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "alpha$beta$gamma", 1)

	if !strings.Contains(clause, "col_a = $1") {
		t.Errorf("clause should contain col_a = $1, got %q", clause)
	}
	if !strings.Contains(clause, "col_b = $2") {
		t.Errorf("clause should contain col_b = $2, got %q", clause)
	}
	if !strings.Contains(clause, "col_c = $3") {
		t.Errorf("clause should contain col_c = $3, got %q", clause)
	}
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(args))
	}
	if nextIdx != 4 {
		t.Errorf("nextIdx = %d, want 4", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// Parameter index tracking
// ---------------------------------------------------------------------------

func TestCompositeSearchClause_StartIdxNonOne(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value", SysColumn: "code_system"},
			{Name: "value", Type: SearchParamNumber, Column: "value_num"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "http://loinc.org|1234$50", 5)

	if !strings.Contains(clause, "code_system = $5") {
		t.Errorf("clause should contain code_system = $5, got %q", clause)
	}
	if !strings.Contains(clause, "code_value = $6") {
		t.Errorf("clause should contain code_value = $6, got %q", clause)
	}
	if !strings.Contains(clause, "value_num = $7") {
		t.Errorf("clause should contain value_num = $7, got %q", clause)
	}
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(args))
	}
	if nextIdx != 8 {
		t.Errorf("nextIdx = %d, want 8", nextIdx)
	}
}

func TestCompositeSearchClause_DateExactArgIdxAdvancement(t *testing.T) {
	// Exact date matching (YYYY-MM-DD) produces 2 args from DateSearchClause
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value"},
			{Name: "value", Type: SearchParamDate, Column: "value_date"},
		},
	}

	_, args, nextIdx := CompositeSearchClause(config, "1234$2023-06-15", 1)

	// Token: 1 arg ($1); Date exact day: 2 args ($2, $3)
	if len(args) != 3 {
		t.Fatalf("expected 3 args for code + exact date, got %d: %v", len(args), args)
	}
	if nextIdx != 4 {
		t.Errorf("nextIdx = %d, want 4", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestCompositeSearchClause_EmptyValue(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value"},
			{Name: "value", Type: SearchParamNumber, Column: "value_num"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "", 1)

	// Empty value splits to [""], first part is "" which is skipped
	if clause != "1=1" {
		t.Errorf("clause = %q, want %q", clause, "1=1")
	}
	if len(args) != 0 {
		t.Errorf("expected 0 args for empty value, got %d: %v", len(args), args)
	}
	if nextIdx != 1 {
		t.Errorf("nextIdx = %d, want 1", nextIdx)
	}
}

func TestCompositeSearchClause_EmptyComponents(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "anything$here", 1)

	if clause != "1=1" {
		t.Errorf("clause = %q, want %q", clause, "1=1")
	}
	if len(args) != 0 {
		t.Errorf("expected 0 args for empty components, got %d", len(args))
	}
	if nextIdx != 1 {
		t.Errorf("nextIdx = %d, want 1", nextIdx)
	}
}

func TestCompositeSearchClause_TooManyParts(t *testing.T) {
	// More $ parts than components: extra parts should be ignored
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "abc$def$ghi", 1)

	// Only first component should be matched
	if clause != "code_value = $1" {
		t.Errorf("clause = %q, want %q", clause, "code_value = $1")
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if args[0] != "abc" {
		t.Errorf("args[0] = %v, want %q", args[0], "abc")
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestCompositeSearchClause_FewerPartsThanComponents(t *testing.T) {
	// Fewer $ parts than components: only provided parts should be matched
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value"},
			{Name: "value", Type: SearchParamNumber, Column: "value_num"},
			{Name: "extra", Type: SearchParamString, Column: "extra_col"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "abc$100", 1)

	if !strings.Contains(clause, "code_value = $1") {
		t.Errorf("clause should contain code_value = $1, got %q", clause)
	}
	if !strings.Contains(clause, "value_num = $2") {
		t.Errorf("clause should contain value_num = $2, got %q", clause)
	}
	// Third component should not appear
	if strings.Contains(clause, "extra_col") {
		t.Errorf("clause should not contain extra_col, got %q", clause)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if nextIdx != 3 {
		t.Errorf("nextIdx = %d, want 3", nextIdx)
	}
}

func TestCompositeSearchClause_EmptySecondPart(t *testing.T) {
	// "code$" splits to ["code", ""] — second part is empty and should be skipped
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value"},
			{Name: "value", Type: SearchParamNumber, Column: "value_num"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "abc$", 1)

	// Only the first component should produce a clause
	if clause != "code_value = $1" {
		t.Errorf("clause = %q, want %q", clause, "code_value = $1")
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestCompositeSearchClause_EmptyFirstPart(t *testing.T) {
	// "$100" splits to ["", "100"] — first part is empty, only second matched
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value"},
			{Name: "value", Type: SearchParamNumber, Column: "value_num"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "$100", 1)

	if clause != "value_num = $1" {
		t.Errorf("clause = %q, want %q", clause, "value_num = $1")
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if args[0] != "100" {
		t.Errorf("args[0] = %v, want %q", args[0], "100")
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestCompositeSearchClause_AllEmptyParts(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value"},
			{Name: "value", Type: SearchParamNumber, Column: "value_num"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "$", 1)

	if clause != "1=1" {
		t.Errorf("clause = %q, want %q", clause, "1=1")
	}
	if len(args) != 0 {
		t.Errorf("expected 0 args, got %d", len(args))
	}
	if nextIdx != 1 {
		t.Errorf("nextIdx = %d, want 1", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// Quantity with unit info (pipe-separated)
// ---------------------------------------------------------------------------

func TestCompositeSearchClause_QuantityWithUnit(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value"},
			{Name: "value", Type: SearchParamQuantity, Column: "value_quantity"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "8480-6$gt5.4|http://unitsofmeasure.org|mmol", 1)

	if !strings.Contains(clause, "code_value = $1") {
		t.Errorf("clause should contain code_value = $1, got %q", clause)
	}
	if !strings.Contains(clause, "value_quantity > $2") {
		t.Errorf("clause should contain value_quantity > $2, got %q", clause)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(args), args)
	}
	if args[1] != "5.4" {
		t.Errorf("args[1] = %v, want %q", args[1], "5.4")
	}
	if nextIdx != 3 {
		t.Errorf("nextIdx = %d, want 3", nextIdx)
	}
}

func TestCompositeSearchClause_QuantityWithoutUnit(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value"},
			{Name: "value", Type: SearchParamQuantity, Column: "value_quantity"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "8480-6$le200", 1)

	if !strings.Contains(clause, "code_value = $1") {
		t.Errorf("clause should contain code_value = $1, got %q", clause)
	}
	if !strings.Contains(clause, "value_quantity <= $2") {
		t.Errorf("clause should contain value_quantity <= $2, got %q", clause)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(args), args)
	}
	if args[1] != "200" {
		t.Errorf("args[1] = %v, want %q", args[1], "200")
	}
	if nextIdx != 3 {
		t.Errorf("nextIdx = %d, want 3", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// URI and Reference component types
// ---------------------------------------------------------------------------

func TestCompositeSearchClause_URIComponent(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "url", Type: SearchParamURI, Column: "profile_url"},
			{Name: "code", Type: SearchParamToken, Column: "code_value"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "http://example.org/profile$active", 1)

	if !strings.Contains(clause, "profile_url = $1") {
		t.Errorf("clause should contain profile_url = $1, got %q", clause)
	}
	if !strings.Contains(clause, "code_value = $2") {
		t.Errorf("clause should contain code_value = $2, got %q", clause)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != "http://example.org/profile" {
		t.Errorf("args[0] = %v, want %q", args[0], "http://example.org/profile")
	}
	if nextIdx != 3 {
		t.Errorf("nextIdx = %d, want 3", nextIdx)
	}
}

func TestCompositeSearchClause_ReferenceComponent(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "subject", Type: SearchParamReference, Column: "patient_id"},
			{Name: "code", Type: SearchParamToken, Column: "code_value"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "Patient/abc-123$final", 1)

	if !strings.Contains(clause, "patient_id = $1") {
		t.Errorf("clause should contain patient_id = $1, got %q", clause)
	}
	if !strings.Contains(clause, "code_value = $2") {
		t.Errorf("clause should contain code_value = $2, got %q", clause)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	// Reference should strip the ResourceType/ prefix
	if args[0] != "abc-123" {
		t.Errorf("args[0] = %v, want %q", args[0], "abc-123")
	}
	if nextIdx != 3 {
		t.Errorf("nextIdx = %d, want 3", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// Token without SysColumn
// ---------------------------------------------------------------------------

func TestCompositeSearchClause_TokenWithoutSysColumn(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "status", Type: SearchParamToken, Column: "status"},
			{Name: "value", Type: SearchParamNumber, Column: "value_num"},
		},
	}

	clause, args, nextIdx := CompositeSearchClause(config, "final$100", 1)

	if !strings.Contains(clause, "status = $1") {
		t.Errorf("clause should contain status = $1, got %q", clause)
	}
	if !strings.Contains(clause, "value_num = $2") {
		t.Errorf("clause should contain value_num = $2, got %q", clause)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != "final" {
		t.Errorf("args[0] = %v, want %q", args[0], "final")
	}
	if nextIdx != 3 {
		t.Errorf("nextIdx = %d, want 3", nextIdx)
	}
}

// ---------------------------------------------------------------------------
// Wrapping in parentheses
// ---------------------------------------------------------------------------

func TestCompositeSearchClause_MultiComponentWrapsInParens(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "a", Type: SearchParamToken, Column: "col_a"},
			{Name: "b", Type: SearchParamToken, Column: "col_b"},
		},
	}

	clause, _, _ := CompositeSearchClause(config, "x$y", 1)

	if !strings.HasPrefix(clause, "(") || !strings.HasSuffix(clause, ")") {
		t.Errorf("multi-component clause should be wrapped in parens, got %q", clause)
	}
}

func TestCompositeSearchClause_SingleClauseNoExtraParens(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value"},
		},
	}

	clause, _, _ := CompositeSearchClause(config, "abc", 1)

	// Single clause should not have extra wrapping parens (the token clause itself might have some)
	if clause != "code_value = $1" {
		t.Errorf("clause = %q, want %q", clause, "code_value = $1")
	}
}

// ---------------------------------------------------------------------------
// quantityComponentClause
// ---------------------------------------------------------------------------

func TestQuantityComponentClause_PlainNumber(t *testing.T) {
	clause, args, nextIdx := quantityComponentClause("val", "100", 1)

	if clause != "val = $1" {
		t.Errorf("clause = %q, want %q", clause, "val = $1")
	}
	if len(args) != 1 || args[0] != "100" {
		t.Errorf("args = %v, want [100]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestQuantityComponentClause_WithPrefix(t *testing.T) {
	clause, args, nextIdx := quantityComponentClause("val", "gt50", 1)

	if clause != "val > $1" {
		t.Errorf("clause = %q, want %q", clause, "val > $1")
	}
	if len(args) != 1 || args[0] != "50" {
		t.Errorf("args = %v, want [50]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestQuantityComponentClause_WithUnitPipe(t *testing.T) {
	clause, args, nextIdx := quantityComponentClause("val", "le5.4|http://unitsofmeasure.org|mmol", 1)

	if clause != "val <= $1" {
		t.Errorf("clause = %q, want %q", clause, "val <= $1")
	}
	if len(args) != 1 || args[0] != "5.4" {
		t.Errorf("args = %v, want [5.4]", args)
	}
	if nextIdx != 2 {
		t.Errorf("nextIdx = %d, want 2", nextIdx)
	}
}

func TestQuantityComponentClause_PipeWithPrefix(t *testing.T) {
	clause, _, _ := quantityComponentClause("val", "ne3.14|http://example.org|unit", 3)

	if clause != "val != $3" {
		t.Errorf("clause = %q, want %q", clause, "val != $3")
	}
}

// ---------------------------------------------------------------------------
// SearchQuery.AddComposite integration
// ---------------------------------------------------------------------------

func TestSearchQuery_AddComposite(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value", SysColumn: "code_system"},
			{Name: "value", Type: SearchParamQuantity, Column: "value_quantity"},
		},
	}

	q := NewSearchQuery("observation", "id, status")
	q.AddComposite(config, "http://loinc.org|8480-6$gt100")

	sql := q.CountSQL()
	if !strings.Contains(sql, "AND") {
		t.Errorf("expected AND in count SQL: %s", sql)
	}
	if !strings.Contains(sql, "code_system") {
		t.Errorf("expected code_system in count SQL: %s", sql)
	}
	if !strings.Contains(sql, "value_quantity") {
		t.Errorf("expected value_quantity in count SQL: %s", sql)
	}

	args := q.CountArgs()
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d: %v", len(args), args)
	}

	if q.Idx() != 4 {
		t.Errorf("idx after AddComposite = %d, want 4", q.Idx())
	}
}

func TestSearchQuery_AddCompositeChainedWithOtherClauses(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value"},
			{Name: "value", Type: SearchParamNumber, Column: "value_num"},
		},
	}

	q := NewSearchQuery("observation", "id")
	q.Add("patient_id = $1", "patient-abc")
	q.AddComposite(config, "1234$gt50")
	q.AddToken("cat_system", "cat_value", "vital-signs")

	sql := q.CountSQL()
	if !strings.Contains(sql, "patient_id = $1") {
		t.Errorf("expected patient_id = $1 in SQL: %s", sql)
	}
	if !strings.Contains(sql, "code_value = $2") {
		t.Errorf("expected code_value = $2 in SQL: %s", sql)
	}
	if !strings.Contains(sql, "value_num > $3") {
		t.Errorf("expected value_num > $3 in SQL: %s", sql)
	}
	if !strings.Contains(sql, "cat_value = $4") {
		t.Errorf("expected cat_value = $4 in SQL: %s", sql)
	}

	args := q.CountArgs()
	if len(args) != 4 {
		t.Fatalf("expected 4 args, got %d: %v", len(args), args)
	}
	if q.Idx() != 5 {
		t.Errorf("idx = %d, want 5", q.Idx())
	}
}

func TestSearchQuery_AddCompositeEmptyValue(t *testing.T) {
	config := CompositeSearchConfig{
		Components: []CompositeComponent{
			{Name: "code", Type: SearchParamToken, Column: "code_value"},
			{Name: "value", Type: SearchParamNumber, Column: "value_num"},
		},
	}

	q := NewSearchQuery("observation", "id")
	q.AddComposite(config, "")

	// Empty value should still add "AND 1=1" to the where clause
	sql := q.CountSQL()
	if !strings.Contains(sql, "1=1 AND 1=1") {
		t.Errorf("expected 1=1 AND 1=1 in SQL for empty composite, got: %s", sql)
	}
	if len(q.CountArgs()) != 0 {
		t.Errorf("expected 0 args, got %d", len(q.CountArgs()))
	}
	if q.Idx() != 1 {
		t.Errorf("idx should remain 1 for empty composite, got %d", q.Idx())
	}
}

// ---------------------------------------------------------------------------
// DefaultCompositeConfigs validation
// ---------------------------------------------------------------------------

func TestDefaultCompositeConfigs_NotEmpty(t *testing.T) {
	configs := DefaultCompositeConfigs()
	if len(configs) == 0 {
		t.Fatal("DefaultCompositeConfigs() returned empty map")
	}
}

func TestDefaultCompositeConfigs_ExpectedKeys(t *testing.T) {
	configs := DefaultCompositeConfigs()

	expected := []string{
		"code-value-quantity",
		"code-value-concept",
		"code-value-date",
		"code-value-string",
		"combo-code-value-quantity",
		"combo-code-value-concept",
	}

	for _, key := range expected {
		if _, ok := configs[key]; !ok {
			t.Errorf("DefaultCompositeConfigs() missing key %q", key)
		}
	}
}

func TestDefaultCompositeConfigs_AllHaveTwoComponents(t *testing.T) {
	configs := DefaultCompositeConfigs()

	for name, config := range configs {
		if len(config.Components) != 2 {
			t.Errorf("config %q has %d components, want 2", name, len(config.Components))
		}
	}
}

func TestDefaultCompositeConfigs_CodeValueQuantity(t *testing.T) {
	configs := DefaultCompositeConfigs()
	cfg, ok := configs["code-value-quantity"]
	if !ok {
		t.Fatal("missing code-value-quantity config")
	}

	if cfg.Components[0].Type != SearchParamToken {
		t.Errorf("first component type = %d, want SearchParamToken (%d)", cfg.Components[0].Type, SearchParamToken)
	}
	if cfg.Components[0].Column != "code_value" {
		t.Errorf("first component column = %q, want %q", cfg.Components[0].Column, "code_value")
	}
	if cfg.Components[0].SysColumn != "code_system" {
		t.Errorf("first component SysColumn = %q, want %q", cfg.Components[0].SysColumn, "code_system")
	}
	if cfg.Components[1].Type != SearchParamQuantity {
		t.Errorf("second component type = %d, want SearchParamQuantity (%d)", cfg.Components[1].Type, SearchParamQuantity)
	}
	if cfg.Components[1].Column != "value_quantity" {
		t.Errorf("second component column = %q, want %q", cfg.Components[1].Column, "value_quantity")
	}
}

func TestDefaultCompositeConfigs_CodeValueConcept(t *testing.T) {
	configs := DefaultCompositeConfigs()
	cfg, ok := configs["code-value-concept"]
	if !ok {
		t.Fatal("missing code-value-concept config")
	}

	if cfg.Components[0].Type != SearchParamToken {
		t.Errorf("first component type = %d, want SearchParamToken", cfg.Components[0].Type)
	}
	if cfg.Components[1].Type != SearchParamToken {
		t.Errorf("second component type = %d, want SearchParamToken", cfg.Components[1].Type)
	}
	if cfg.Components[1].Column != "value_code" {
		t.Errorf("second component column = %q, want %q", cfg.Components[1].Column, "value_code")
	}
	if cfg.Components[1].SysColumn != "value_system" {
		t.Errorf("second component SysColumn = %q, want %q", cfg.Components[1].SysColumn, "value_system")
	}
}

func TestDefaultCompositeConfigs_CodeValueDate(t *testing.T) {
	configs := DefaultCompositeConfigs()
	cfg, ok := configs["code-value-date"]
	if !ok {
		t.Fatal("missing code-value-date config")
	}

	if cfg.Components[1].Type != SearchParamDate {
		t.Errorf("second component type = %d, want SearchParamDate (%d)", cfg.Components[1].Type, SearchParamDate)
	}
	if cfg.Components[1].Column != "value_date" {
		t.Errorf("second component column = %q, want %q", cfg.Components[1].Column, "value_date")
	}
}

func TestDefaultCompositeConfigs_CodeValueString(t *testing.T) {
	configs := DefaultCompositeConfigs()
	cfg, ok := configs["code-value-string"]
	if !ok {
		t.Fatal("missing code-value-string config")
	}

	if cfg.Components[1].Type != SearchParamString {
		t.Errorf("second component type = %d, want SearchParamString (%d)", cfg.Components[1].Type, SearchParamString)
	}
	if cfg.Components[1].Column != "value_string" {
		t.Errorf("second component column = %q, want %q", cfg.Components[1].Column, "value_string")
	}
}

func TestDefaultCompositeConfigs_ComboConfigs(t *testing.T) {
	configs := DefaultCompositeConfigs()

	// combo-code-value-quantity should use combo_ prefixed columns
	cfg := configs["combo-code-value-quantity"]
	if cfg.Components[0].Column != "combo_code_value" {
		t.Errorf("combo quantity first col = %q, want %q", cfg.Components[0].Column, "combo_code_value")
	}
	if cfg.Components[0].SysColumn != "combo_code_system" {
		t.Errorf("combo quantity first sys = %q, want %q", cfg.Components[0].SysColumn, "combo_code_system")
	}
	if cfg.Components[1].Column != "combo_value_quantity" {
		t.Errorf("combo quantity second col = %q, want %q", cfg.Components[1].Column, "combo_value_quantity")
	}

	// combo-code-value-concept should use combo_ prefixed columns
	cfg = configs["combo-code-value-concept"]
	if cfg.Components[0].Column != "combo_code_value" {
		t.Errorf("combo concept first col = %q, want %q", cfg.Components[0].Column, "combo_code_value")
	}
	if cfg.Components[1].Column != "combo_value_code" {
		t.Errorf("combo concept second col = %q, want %q", cfg.Components[1].Column, "combo_value_code")
	}
	if cfg.Components[1].SysColumn != "combo_value_system" {
		t.Errorf("combo concept second sys = %q, want %q", cfg.Components[1].SysColumn, "combo_value_system")
	}
}

// ---------------------------------------------------------------------------
// SearchParamComposite constant
// ---------------------------------------------------------------------------

func TestSearchParamComposite_DistinctFromOtherTypes(t *testing.T) {
	types := []SearchParamType{
		SearchParamToken,
		SearchParamDate,
		SearchParamString,
		SearchParamReference,
		SearchParamNumber,
		SearchParamQuantity,
		SearchParamURI,
	}
	for _, typ := range types {
		if SearchParamComposite == typ {
			t.Errorf("SearchParamComposite (%d) should be distinct from type %d", SearchParamComposite, typ)
		}
	}
}

// ---------------------------------------------------------------------------
// DefaultCompositeConfigs used with CompositeSearchClause end-to-end
// ---------------------------------------------------------------------------

func TestDefaultConfig_CodeValueQuantityEndToEnd(t *testing.T) {
	configs := DefaultCompositeConfigs()
	cfg := configs["code-value-quantity"]

	clause, args, nextIdx := CompositeSearchClause(cfg, "http://loinc.org|8480-6$gt5.4|http://unitsofmeasure.org|mmol", 1)

	if !strings.Contains(clause, "code_system = $1") {
		t.Errorf("clause missing code_system, got %q", clause)
	}
	if !strings.Contains(clause, "code_value = $2") {
		t.Errorf("clause missing code_value, got %q", clause)
	}
	if !strings.Contains(clause, "value_quantity > $3") {
		t.Errorf("clause missing value_quantity >, got %q", clause)
	}
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d: %v", len(args), args)
	}
	if args[0] != "http://loinc.org" {
		t.Errorf("args[0] = %v, want %q", args[0], "http://loinc.org")
	}
	if args[1] != "8480-6" {
		t.Errorf("args[1] = %v, want %q", args[1], "8480-6")
	}
	if args[2] != "5.4" {
		t.Errorf("args[2] = %v, want %q", args[2], "5.4")
	}
	if nextIdx != 4 {
		t.Errorf("nextIdx = %d, want 4", nextIdx)
	}
}

func TestDefaultConfig_CodeValueConceptEndToEnd(t *testing.T) {
	configs := DefaultCompositeConfigs()
	cfg := configs["code-value-concept"]

	clause, args, nextIdx := CompositeSearchClause(cfg, "http://loinc.org|1234$http://snomed.info/sct|5678", 1)

	if !strings.Contains(clause, "code_system = $1") {
		t.Errorf("clause missing code_system, got %q", clause)
	}
	if !strings.Contains(clause, "code_value = $2") {
		t.Errorf("clause missing code_value, got %q", clause)
	}
	if !strings.Contains(clause, "value_system = $3") {
		t.Errorf("clause missing value_system, got %q", clause)
	}
	if !strings.Contains(clause, "value_code = $4") {
		t.Errorf("clause missing value_code, got %q", clause)
	}
	if len(args) != 4 {
		t.Fatalf("expected 4 args, got %d", len(args))
	}
	if nextIdx != 5 {
		t.Errorf("nextIdx = %d, want 5", nextIdx)
	}
}
