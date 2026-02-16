package fhir

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Tokenizer tests
// ---------------------------------------------------------------------------

func TestTokenizeFilter_SimpleExpression(t *testing.T) {
	tokens, err := tokenizeFilter(`name eq "Smith"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0].Type != tokenWord || tokens[0].Value != "name" {
		t.Errorf("token[0] = %+v, want word 'name'", tokens[0])
	}
	if tokens[1].Type != tokenWord || tokens[1].Value != "eq" {
		t.Errorf("token[1] = %+v, want word 'eq'", tokens[1])
	}
	if tokens[2].Type != tokenString || tokens[2].Value != "Smith" {
		t.Errorf("token[2] = %+v, want string 'Smith'", tokens[2])
	}
}

func TestTokenizeFilter_QuotedStringWithSpaces(t *testing.T) {
	tokens, err := tokenizeFilter(`name eq "John Smith"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[2].Type != tokenString || tokens[2].Value != "John Smith" {
		t.Errorf("token[2] = %+v, want string 'John Smith'", tokens[2])
	}
}

func TestTokenizeFilter_Parentheses(t *testing.T) {
	tokens, err := tokenizeFilter(`(status eq active)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 5 {
		t.Fatalf("expected 5 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0].Type != tokenLParen {
		t.Errorf("token[0].Type = %d, want tokenLParen", tokens[0].Type)
	}
	if tokens[4].Type != tokenRParen {
		t.Errorf("token[4].Type = %d, want tokenRParen", tokens[4].Type)
	}
}

func TestTokenizeFilter_AndOrNot(t *testing.T) {
	tokens, err := tokenizeFilter(`name eq "Smith" and not status eq active`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// name, eq, "Smith", and, not, status, eq, active
	if len(tokens) != 8 {
		t.Fatalf("expected 8 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[3].Type != tokenAnd {
		t.Errorf("token[3].Type = %d, want tokenAnd", tokens[3].Type)
	}
	if tokens[4].Type != tokenNot {
		t.Errorf("token[4].Type = %d, want tokenNot", tokens[4].Type)
	}
}

func TestTokenizeFilter_OrKeyword(t *testing.T) {
	tokens, err := tokenizeFilter(`status eq active or status eq planned`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// status, eq, active, or, status, eq, planned
	if len(tokens) != 7 {
		t.Fatalf("expected 7 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[3].Type != tokenOr {
		t.Errorf("token[3].Type = %d, want tokenOr", tokens[3].Type)
	}
}

func TestTokenizeFilter_Empty(t *testing.T) {
	tokens, err := tokenizeFilter("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens, got %d", len(tokens))
	}
}

func TestTokenizeFilter_NestedParens(t *testing.T) {
	tokens, err := tokenizeFilter(`((a eq b))`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// (, (, a, eq, b, ), )
	if len(tokens) != 7 {
		t.Fatalf("expected 7 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0].Type != tokenLParen || tokens[1].Type != tokenLParen {
		t.Error("expected two left parens at start")
	}
	if tokens[5].Type != tokenRParen || tokens[6].Type != tokenRParen {
		t.Error("expected two right parens at end")
	}
}

func TestTokenizeFilter_UnclosedQuote(t *testing.T) {
	_, err := tokenizeFilter(`name eq "Smith`)
	if err == nil {
		t.Fatal("expected error for unclosed quote, got nil")
	}
	if !strings.Contains(err.Error(), "unclosed") {
		t.Errorf("error = %q, want it to mention unclosed quote", err.Error())
	}
}

func TestTokenizeFilter_UnquotedValues(t *testing.T) {
	tokens, err := tokenizeFilter(`birthdate ge 2000-01-01`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[2].Type != tokenWord || tokens[2].Value != "2000-01-01" {
		t.Errorf("token[2] = %+v, want word '2000-01-01'", tokens[2])
	}
}

func TestTokenizeFilter_SystemCodeValue(t *testing.T) {
	tokens, err := tokenizeFilter(`code eq http://loinc.org|8867-4`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[2].Value != "http://loinc.org|8867-4" {
		t.Errorf("token[2].Value = %q, want %q", tokens[2].Value, "http://loinc.org|8867-4")
	}
}

// ---------------------------------------------------------------------------
// Parser tests
// ---------------------------------------------------------------------------

func TestParseFilterExpression_SimpleEq(t *testing.T) {
	expr, err := ParseFilterExpression(`name eq "Smith"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expr.Type != FilterExprParam {
		t.Errorf("Type = %d, want FilterExprParam", expr.Type)
	}
	if expr.Param != "name" {
		t.Errorf("Param = %q, want %q", expr.Param, "name")
	}
	if expr.Operator != "eq" {
		t.Errorf("Operator = %q, want %q", expr.Operator, "eq")
	}
	if expr.Value != "Smith" {
		t.Errorf("Value = %q, want %q", expr.Value, "Smith")
	}
}

func TestParseFilterExpression_And(t *testing.T) {
	expr, err := ParseFilterExpression(`name eq "Smith" and birthdate ge 2000-01-01`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expr.Type != FilterExprAnd {
		t.Errorf("Type = %d, want FilterExprAnd", expr.Type)
	}
	if expr.Left == nil || expr.Right == nil {
		t.Fatal("And node should have Left and Right children")
	}
	if expr.Left.Param != "name" {
		t.Errorf("Left.Param = %q, want %q", expr.Left.Param, "name")
	}
	if expr.Right.Param != "birthdate" {
		t.Errorf("Right.Param = %q, want %q", expr.Right.Param, "birthdate")
	}
}

func TestParseFilterExpression_Or(t *testing.T) {
	expr, err := ParseFilterExpression(`status eq active or status eq planned`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expr.Type != FilterExprOr {
		t.Errorf("Type = %d, want FilterExprOr", expr.Type)
	}
	if expr.Left == nil || expr.Right == nil {
		t.Fatal("Or node should have Left and Right children")
	}
	if expr.Left.Value != "active" {
		t.Errorf("Left.Value = %q, want %q", expr.Left.Value, "active")
	}
	if expr.Right.Value != "planned" {
		t.Errorf("Right.Value = %q, want %q", expr.Right.Value, "planned")
	}
}

func TestParseFilterExpression_Not(t *testing.T) {
	expr, err := ParseFilterExpression(`not status eq inactive`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expr.Type != FilterExprNot {
		t.Errorf("Type = %d, want FilterExprNot", expr.Type)
	}
	if expr.Child == nil {
		t.Fatal("Not node should have a Child")
	}
	if expr.Child.Param != "status" {
		t.Errorf("Child.Param = %q, want %q", expr.Child.Param, "status")
	}
}

func TestParseFilterExpression_Parenthesized(t *testing.T) {
	expr, err := ParseFilterExpression(`(status eq active or status eq planned) and date ge 2024-01-01`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expr.Type != FilterExprAnd {
		t.Errorf("Type = %d, want FilterExprAnd", expr.Type)
	}
	if expr.Left.Type != FilterExprOr {
		t.Errorf("Left.Type = %d, want FilterExprOr", expr.Left.Type)
	}
	if expr.Right.Param != "date" {
		t.Errorf("Right.Param = %q, want %q", expr.Right.Param, "date")
	}
}

func TestParseFilterExpression_ComplexNested(t *testing.T) {
	expr, err := ParseFilterExpression(`(name eq "Smith" and gender eq male) or (name eq "Jones" and gender eq female)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expr.Type != FilterExprOr {
		t.Errorf("Type = %d, want FilterExprOr", expr.Type)
	}
	if expr.Left.Type != FilterExprAnd {
		t.Errorf("Left.Type = %d, want FilterExprAnd", expr.Left.Type)
	}
	if expr.Right.Type != FilterExprAnd {
		t.Errorf("Right.Type = %d, want FilterExprAnd", expr.Right.Type)
	}
}

func TestParseFilterExpression_Precedence_AndBeforeOr(t *testing.T) {
	// "a eq 1 or b eq 2 and c eq 3" should parse as "a eq 1 or (b eq 2 and c eq 3)"
	// because AND has higher precedence than OR.
	expr, err := ParseFilterExpression(`a eq 1 or b eq 2 and c eq 3`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expr.Type != FilterExprOr {
		t.Errorf("Type = %d, want FilterExprOr (AND should bind tighter)", expr.Type)
	}
	if expr.Right.Type != FilterExprAnd {
		t.Errorf("Right.Type = %d, want FilterExprAnd", expr.Right.Type)
	}
}

func TestParseFilterExpression_PresentOperator(t *testing.T) {
	expr, err := ParseFilterExpression(`name pr`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expr.Type != FilterExprParam {
		t.Errorf("Type = %d, want FilterExprParam", expr.Type)
	}
	if expr.Param != "name" {
		t.Errorf("Param = %q, want %q", expr.Param, "name")
	}
	if expr.Operator != "pr" {
		t.Errorf("Operator = %q, want %q", expr.Operator, "pr")
	}
	if expr.Value != "" {
		t.Errorf("Value = %q, want empty string for 'pr' operator", expr.Value)
	}
}

func TestParseFilterExpression_TripleAnd(t *testing.T) {
	expr, err := ParseFilterExpression(`a eq 1 and b eq 2 and c eq 3`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be left-associative: ((a eq 1 and b eq 2) and c eq 3)
	if expr.Type != FilterExprAnd {
		t.Errorf("Type = %d, want FilterExprAnd", expr.Type)
	}
	if expr.Left.Type != FilterExprAnd {
		t.Errorf("Left.Type = %d, want FilterExprAnd (left-associative)", expr.Left.Type)
	}
	if expr.Right.Param != "c" {
		t.Errorf("Right.Param = %q, want %q", expr.Right.Param, "c")
	}
}

// ---------------------------------------------------------------------------
// Parser error tests
// ---------------------------------------------------------------------------

func TestParseFilterExpression_UnclosedParen(t *testing.T) {
	_, err := ParseFilterExpression(`(status eq active`)
	if err == nil {
		t.Fatal("expected error for unclosed paren, got nil")
	}
	if !strings.Contains(err.Error(), "expected ')'") {
		t.Errorf("error = %q, want it to mention expected ')'", err.Error())
	}
}

func TestParseFilterExpression_EmptyInput(t *testing.T) {
	_, err := ParseFilterExpression("")
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("error = %q, want it to mention empty", err.Error())
	}
}

func TestParseFilterExpression_MissingOperator(t *testing.T) {
	_, err := ParseFilterExpression(`name`)
	if err == nil {
		t.Fatal("expected error for missing operator, got nil")
	}
}

func TestParseFilterExpression_MissingValue(t *testing.T) {
	_, err := ParseFilterExpression(`name eq`)
	if err == nil {
		t.Fatal("expected error for missing value, got nil")
	}
}

func TestParseFilterExpression_InvalidOp(t *testing.T) {
	_, err := ParseFilterExpression(`name xyz "Smith"`)
	if err == nil {
		t.Fatal("expected error for invalid operator, got nil")
	}
	if !strings.Contains(err.Error(), "unknown") || !strings.Contains(err.Error(), "xyz") {
		t.Errorf("error = %q, want it to mention unknown operator 'xyz'", err.Error())
	}
}

func TestParseFilterExpression_ExtraCloseParen(t *testing.T) {
	_, err := ParseFilterExpression(`status eq active)`)
	if err == nil {
		t.Fatal("expected error for extra close paren, got nil")
	}
}

// ---------------------------------------------------------------------------
// Validation tests
// ---------------------------------------------------------------------------

func TestValidateFilterExpr_Valid(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "name",
		Operator: "eq",
		Value:    "Smith",
	}
	ctx := &FilterContext{
		ResourceType: "Patient",
		ColumnMappings: map[string]FilterColumnMapping{
			"name": {Column: "last_name", ParamType: "string"},
		},
		TableAlias: "r",
	}
	issues := ValidateFilterExpression(expr, ctx)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d: %v", len(issues), issues)
	}
}

func TestValidateFilterExpr_UnknownParam(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "unknown_field",
		Operator: "eq",
		Value:    "test",
	}
	ctx := &FilterContext{
		ResourceType: "Patient",
		ColumnMappings: map[string]FilterColumnMapping{
			"name": {Column: "last_name", ParamType: "string"},
		},
		TableAlias: "r",
	}
	issues := ValidateFilterExpression(expr, ctx)
	if len(issues) == 0 {
		t.Fatal("expected validation issues for unknown param, got none")
	}
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "unknown_field") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected issue mentioning 'unknown_field', got %v", issues)
	}
}

func TestValidateFilterExpr_InvalidOperatorForType(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "birthdate",
		Operator: "co", // contains is not valid for dates
		Value:    "2000",
	}
	ctx := &FilterContext{
		ResourceType: "Patient",
		ColumnMappings: map[string]FilterColumnMapping{
			"birthdate": {Column: "birth_date", ParamType: "date"},
		},
		TableAlias: "r",
	}
	issues := ValidateFilterExpression(expr, ctx)
	if len(issues) == 0 {
		t.Fatal("expected validation issues for invalid operator/type combo, got none")
	}
}

func TestValidateFilterExpr_AndNode(t *testing.T) {
	expr := &FilterExprNode{
		Type: FilterExprAnd,
		Left: &FilterExprNode{
			Type:     FilterExprParam,
			Param:    "name",
			Operator: "eq",
			Value:    "Smith",
		},
		Right: &FilterExprNode{
			Type:     FilterExprParam,
			Param:    "unknown",
			Operator: "eq",
			Value:    "test",
		},
	}
	ctx := &FilterContext{
		ResourceType: "Patient",
		ColumnMappings: map[string]FilterColumnMapping{
			"name": {Column: "last_name", ParamType: "string"},
		},
		TableAlias: "r",
	}
	issues := ValidateFilterExpression(expr, ctx)
	if len(issues) == 0 {
		t.Fatal("expected validation issue for unknown param in right child")
	}
}

// ---------------------------------------------------------------------------
// CompileFilterToSQL tests - String filters
// ---------------------------------------------------------------------------

func TestCompileFilterToSQL_EqString(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "name",
		Operator: "eq",
		Value:    "Smith",
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.last_name = $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.last_name = $1")
	}
	if len(args) != 1 || args[0] != "Smith" {
		t.Errorf("args = %v, want [Smith]", args)
	}
}

func TestCompileFilterToSQL_NeString(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "name",
		Operator: "ne",
		Value:    "Smith",
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.last_name != $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.last_name != $1")
	}
	if len(args) != 1 || args[0] != "Smith" {
		t.Errorf("args = %v, want [Smith]", args)
	}
}

func TestCompileFilterToSQL_Contains(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "name",
		Operator: "co",
		Value:    "mit",
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.last_name ILIKE $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.last_name ILIKE $1")
	}
	if len(args) != 1 || args[0] != "%mit%" {
		t.Errorf("args = %v, want [%%mit%%]", args)
	}
}

func TestCompileFilterToSQL_StartsWith(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "name",
		Operator: "sw",
		Value:    "Sm",
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.last_name ILIKE $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.last_name ILIKE $1")
	}
	if len(args) != 1 || args[0] != "Sm%" {
		t.Errorf("args = %v, want [Sm%%]", args)
	}
}

func TestCompileFilterToSQL_EndsWith(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "name",
		Operator: "ew",
		Value:    "son",
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.last_name ILIKE $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.last_name ILIKE $1")
	}
	if len(args) != 1 || args[0] != "%son" {
		t.Errorf("args = %v, want [%%son]", args)
	}
}

// ---------------------------------------------------------------------------
// CompileFilterToSQL tests - Date filters
// ---------------------------------------------------------------------------

func TestCompileFilterToSQL_DateGt(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "birthdate",
		Operator: "gt",
		Value:    "2000-01-01",
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.birth_date > $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.birth_date > $1")
	}
	if len(args) != 1 || args[0] != "2000-01-01" {
		t.Errorf("args = %v, want [2000-01-01]", args)
	}
}

func TestCompileFilterToSQL_DateLt(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "birthdate",
		Operator: "lt",
		Value:    "2000-01-01",
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.birth_date < $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.birth_date < $1")
	}
	if len(args) != 1 || args[0] != "2000-01-01" {
		t.Errorf("args = %v, want [2000-01-01]", args)
	}
}

func TestCompileFilterToSQL_DateGe(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "birthdate",
		Operator: "ge",
		Value:    "2000-01-01",
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.birth_date >= $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.birth_date >= $1")
	}
	if len(args) != 1 || args[0] != "2000-01-01" {
		t.Errorf("args = %v, want [2000-01-01]", args)
	}
}

func TestCompileFilterToSQL_DateLe(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "birthdate",
		Operator: "le",
		Value:    "2000-01-01",
	}
	ctx := testFilterContext()
	sql, _, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.birth_date <= $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.birth_date <= $1")
	}
}

func TestCompileFilterToSQL_DateSa(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "birthdate",
		Operator: "sa",
		Value:    "2000-01-01",
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.birth_date > $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.birth_date > $1")
	}
	if len(args) != 1 || args[0] != "2000-01-01" {
		t.Errorf("args = %v, want [2000-01-01]", args)
	}
}

func TestCompileFilterToSQL_DateEb(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "birthdate",
		Operator: "eb",
		Value:    "2000-01-01",
	}
	ctx := testFilterContext()
	sql, _, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.birth_date < $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.birth_date < $1")
	}
}

func TestCompileFilterToSQL_DateAp(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "birthdate",
		Operator: "ap",
		Value:    "2000-06-15",
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "BETWEEN") || !strings.Contains(sql, "$1") || !strings.Contains(sql, "$2") {
		t.Errorf("SQL = %q, want BETWEEN clause with two placeholders", sql)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args for approximate, got %d", len(args))
	}
}

func TestCompileFilterToSQL_DateEq(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "birthdate",
		Operator: "eq",
		Value:    "2000-01-01",
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.birth_date = $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.birth_date = $1")
	}
	if len(args) != 1 || args[0] != "2000-01-01" {
		t.Errorf("args = %v, want [2000-01-01]", args)
	}
}

// ---------------------------------------------------------------------------
// CompileFilterToSQL tests - Token filters
// ---------------------------------------------------------------------------

func TestCompileFilterToSQL_TokenEq(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "status",
		Operator: "eq",
		Value:    "active",
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.status = $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.status = $1")
	}
	if len(args) != 1 || args[0] != "active" {
		t.Errorf("args = %v, want [active]", args)
	}
}

func TestCompileFilterToSQL_TokenNe(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "status",
		Operator: "ne",
		Value:    "inactive",
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.status != $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.status != $1")
	}
	if len(args) != 1 || args[0] != "inactive" {
		t.Errorf("args = %v, want [inactive]", args)
	}
}

func TestCompileFilterToSQL_TokenIn(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "status",
		Operator: "in",
		Value:    "active,planned,completed",
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "IN") {
		t.Errorf("SQL = %q, want it to contain IN clause", sql)
	}
	if len(args) != 3 {
		t.Errorf("expected 3 args for IN, got %d", len(args))
	}
}

func TestCompileFilterToSQL_TokenNi(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "status",
		Operator: "ni",
		Value:    "inactive,cancelled",
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "NOT IN") {
		t.Errorf("SQL = %q, want it to contain NOT IN clause", sql)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args for NOT IN, got %d", len(args))
	}
}

func TestCompileFilterToSQL_TokenWithSystem(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "code",
		Operator: "eq",
		Value:    "http://loinc.org|8867-4",
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "code_system") && !strings.Contains(sql, "code_value") {
		t.Errorf("SQL = %q, want it to reference code_system and code_value columns", sql)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args for system|code, got %d", len(args))
	}
}

// ---------------------------------------------------------------------------
// CompileFilterToSQL tests - Number filters
// ---------------------------------------------------------------------------

func TestCompileFilterToSQL_NumberGt(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "value_quantity",
		Operator: "gt",
		Value:    "100",
	}
	ctx := testFilterContextObs()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.value_quantity > $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.value_quantity > $1")
	}
	if len(args) != 1 || args[0] != "100" {
		t.Errorf("args = %v, want [100]", args)
	}
}

func TestCompileFilterToSQL_NumberLt(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "value_quantity",
		Operator: "lt",
		Value:    "50",
	}
	ctx := testFilterContextObs()
	sql, _, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.value_quantity < $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.value_quantity < $1")
	}
}

func TestCompileFilterToSQL_NumberEq(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "value_quantity",
		Operator: "eq",
		Value:    "75",
	}
	ctx := testFilterContextObs()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.value_quantity = $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.value_quantity = $1")
	}
	if len(args) != 1 || args[0] != "75" {
		t.Errorf("args = %v, want [75]", args)
	}
}

func TestCompileFilterToSQL_NumberNe(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "value_quantity",
		Operator: "ne",
		Value:    "0",
	}
	ctx := testFilterContextObs()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.value_quantity != $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.value_quantity != $1")
	}
	if len(args) != 1 || args[0] != "0" {
		t.Errorf("args = %v, want [0]", args)
	}
}

// ---------------------------------------------------------------------------
// CompileFilterToSQL tests - Present operator
// ---------------------------------------------------------------------------

func TestCompileFilterToSQL_Present(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "name",
		Operator: "pr",
		Value:    "",
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "r.last_name IS NOT NULL" {
		t.Errorf("SQL = %q, want %q", sql, "r.last_name IS NOT NULL")
	}
	if len(args) != 0 {
		t.Errorf("expected 0 args for present, got %d", len(args))
	}
}

func TestCompileFilterToSQL_NotPresent(t *testing.T) {
	expr := &FilterExprNode{
		Type: FilterExprNot,
		Child: &FilterExprNode{
			Type:     FilterExprParam,
			Param:    "name",
			Operator: "pr",
			Value:    "",
		},
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "NOT") && !strings.Contains(sql, "IS NULL") {
		t.Errorf("SQL = %q, want it to contain NOT or IS NULL", sql)
	}
	if len(args) != 0 {
		t.Errorf("expected 0 args for not present, got %d", len(args))
	}
}

// ---------------------------------------------------------------------------
// CompileFilterToSQL tests - Boolean combinators
// ---------------------------------------------------------------------------

func TestCompileFilterToSQL_And(t *testing.T) {
	expr := &FilterExprNode{
		Type: FilterExprAnd,
		Left: &FilterExprNode{
			Type:     FilterExprParam,
			Param:    "name",
			Operator: "eq",
			Value:    "Smith",
		},
		Right: &FilterExprNode{
			Type:     FilterExprParam,
			Param:    "gender",
			Operator: "eq",
			Value:    "male",
		},
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "(r.last_name = $1 AND r.gender = $2)" {
		t.Errorf("SQL = %q, want %q", sql, "(r.last_name = $1 AND r.gender = $2)")
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

func TestCompileFilterToSQL_Or(t *testing.T) {
	expr := &FilterExprNode{
		Type: FilterExprOr,
		Left: &FilterExprNode{
			Type:     FilterExprParam,
			Param:    "status",
			Operator: "eq",
			Value:    "active",
		},
		Right: &FilterExprNode{
			Type:     FilterExprParam,
			Param:    "status",
			Operator: "eq",
			Value:    "planned",
		},
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "(r.status = $1 OR r.status = $2)" {
		t.Errorf("SQL = %q, want %q", sql, "(r.status = $1 OR r.status = $2)")
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

func TestCompileFilterToSQL_Not(t *testing.T) {
	expr := &FilterExprNode{
		Type: FilterExprNot,
		Child: &FilterExprNode{
			Type:     FilterExprParam,
			Param:    "status",
			Operator: "eq",
			Value:    "inactive",
		},
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "NOT (r.status = $1)" {
		t.Errorf("SQL = %q, want %q", sql, "NOT (r.status = $1)")
	}
	if len(args) != 1 || args[0] != "inactive" {
		t.Errorf("args = %v, want [inactive]", args)
	}
}

func TestCompileFilterToSQL_ComplexNested(t *testing.T) {
	// (status eq active or status eq planned) and name eq "Smith"
	expr := &FilterExprNode{
		Type: FilterExprAnd,
		Left: &FilterExprNode{
			Type: FilterExprOr,
			Left: &FilterExprNode{
				Type:     FilterExprParam,
				Param:    "status",
				Operator: "eq",
				Value:    "active",
			},
			Right: &FilterExprNode{
				Type:     FilterExprParam,
				Param:    "status",
				Operator: "eq",
				Value:    "planned",
			},
		},
		Right: &FilterExprNode{
			Type:     FilterExprParam,
			Param:    "name",
			Operator: "eq",
			Value:    "Smith",
		},
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "((r.status = $1 OR r.status = $2) AND r.last_name = $3)"
	if sql != expected {
		t.Errorf("SQL = %q, want %q", sql, expected)
	}
	if len(args) != 3 {
		t.Errorf("expected 3 args, got %d", len(args))
	}
}

func TestCompileFilterToSQL_UnknownParam(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "unknown",
		Operator: "eq",
		Value:    "test",
	}
	ctx := testFilterContext()
	_, _, err := CompileFilterToSQL(expr, ctx, 1)
	if err == nil {
		t.Fatal("expected error for unknown param, got nil")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error = %q, want it to mention unknown param", err.Error())
	}
}

// ---------------------------------------------------------------------------
// compileStringFilter tests
// ---------------------------------------------------------------------------

func TestCompileStringFilter_Eq(t *testing.T) {
	sql, args := compileStringFilter("r.name", "eq", "Smith", 1)
	if sql != "r.name = $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.name = $1")
	}
	if len(args) != 1 || args[0] != "Smith" {
		t.Errorf("args = %v, want [Smith]", args)
	}
}

func TestCompileStringFilter_Co(t *testing.T) {
	sql, args := compileStringFilter("r.name", "co", "mit", 1)
	if sql != "r.name ILIKE $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.name ILIKE $1")
	}
	if len(args) != 1 || args[0] != "%mit%" {
		t.Errorf("args = %v, want [%%mit%%]", args)
	}
}

func TestCompileStringFilter_Sw(t *testing.T) {
	sql, args := compileStringFilter("r.name", "sw", "Sm", 1)
	if sql != "r.name ILIKE $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.name ILIKE $1")
	}
	if len(args) != 1 || args[0] != "Sm%" {
		t.Errorf("args = %v, want [Sm%%]", args)
	}
}

func TestCompileStringFilter_Ew(t *testing.T) {
	sql, args := compileStringFilter("r.name", "ew", "son", 1)
	if sql != "r.name ILIKE $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.name ILIKE $1")
	}
	if len(args) != 1 || args[0] != "%son" {
		t.Errorf("args = %v, want [%%son]", args)
	}
}

// ---------------------------------------------------------------------------
// compileTokenFilter tests
// ---------------------------------------------------------------------------

func TestCompileTokenFilter_CodeOnly(t *testing.T) {
	sql, args := compileTokenFilter("r.code_value", "r.code_system", "eq", "8867-4", 1)
	if sql != "r.code_value = $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.code_value = $1")
	}
	if len(args) != 1 || args[0] != "8867-4" {
		t.Errorf("args = %v, want [8867-4]", args)
	}
}

func TestCompileTokenFilter_SystemCode(t *testing.T) {
	sql, args := compileTokenFilter("r.code_value", "r.code_system", "eq", "http://loinc.org|8867-4", 1)
	if !strings.Contains(sql, "r.code_system") {
		t.Errorf("SQL = %q, want it to reference code_system", sql)
	}
	if !strings.Contains(sql, "r.code_value") {
		t.Errorf("SQL = %q, want it to reference code_value", sql)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args for system|code, got %d", len(args))
	}
}

func TestCompileTokenFilter_SystemOnly(t *testing.T) {
	sql, args := compileTokenFilter("r.code_value", "r.code_system", "eq", "http://loinc.org|", 1)
	if !strings.Contains(sql, "r.code_system") {
		t.Errorf("SQL = %q, want it to reference code_system", sql)
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg for system only, got %d", len(args))
	}
}

// ---------------------------------------------------------------------------
// compileDateFilter tests
// ---------------------------------------------------------------------------

func TestCompileDateFilter_Eq(t *testing.T) {
	sql, args := compileDateFilter("r.birth_date", "eq", "2000-01-01", 1)
	if sql != "r.birth_date = $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.birth_date = $1")
	}
	if len(args) != 1 || args[0] != "2000-01-01" {
		t.Errorf("args = %v, want [2000-01-01]", args)
	}
}

func TestCompileDateFilter_Gt(t *testing.T) {
	sql, args := compileDateFilter("r.birth_date", "gt", "2000-01-01", 1)
	if sql != "r.birth_date > $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.birth_date > $1")
	}
	if len(args) != 1 || args[0] != "2000-01-01" {
		t.Errorf("args = %v, want [2000-01-01]", args)
	}
}

func TestCompileDateFilter_Lt(t *testing.T) {
	sql, args := compileDateFilter("r.birth_date", "lt", "2000-01-01", 1)
	if sql != "r.birth_date < $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.birth_date < $1")
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

func TestCompileDateFilter_Ge(t *testing.T) {
	sql, args := compileDateFilter("r.birth_date", "ge", "2000-01-01", 1)
	if sql != "r.birth_date >= $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.birth_date >= $1")
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

func TestCompileDateFilter_Le(t *testing.T) {
	sql, args := compileDateFilter("r.birth_date", "le", "2000-01-01", 1)
	if sql != "r.birth_date <= $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.birth_date <= $1")
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

func TestCompileDateFilter_Sa(t *testing.T) {
	sql, args := compileDateFilter("r.birth_date", "sa", "2000-01-01", 1)
	if sql != "r.birth_date > $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.birth_date > $1")
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

func TestCompileDateFilter_Eb(t *testing.T) {
	sql, args := compileDateFilter("r.birth_date", "eb", "2000-01-01", 1)
	if sql != "r.birth_date < $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.birth_date < $1")
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

func TestCompileDateFilter_Ap(t *testing.T) {
	sql, args := compileDateFilter("r.birth_date", "ap", "2000-06-15", 1)
	if !strings.Contains(sql, "BETWEEN") {
		t.Errorf("SQL = %q, want BETWEEN clause", sql)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args for approximate, got %d", len(args))
	}
}

// ---------------------------------------------------------------------------
// compileNumberFilter tests
// ---------------------------------------------------------------------------

func TestCompileNumberFilter_Eq(t *testing.T) {
	sql, args := compileNumberFilter("r.value", "eq", "100", 1)
	if sql != "r.value = $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.value = $1")
	}
	if len(args) != 1 || args[0] != "100" {
		t.Errorf("args = %v, want [100]", args)
	}
}

func TestCompileNumberFilter_Ne(t *testing.T) {
	sql, args := compileNumberFilter("r.value", "ne", "0", 1)
	if sql != "r.value != $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.value != $1")
	}
	if len(args) != 1 || args[0] != "0" {
		t.Errorf("args = %v, want [0]", args)
	}
}

func TestCompileNumberFilter_Gt(t *testing.T) {
	sql, args := compileNumberFilter("r.value", "gt", "50", 1)
	if sql != "r.value > $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.value > $1")
	}
	if len(args) != 1 || args[0] != "50" {
		t.Errorf("args = %v, want [50]", args)
	}
}

func TestCompileNumberFilter_Lt(t *testing.T) {
	sql, args := compileNumberFilter("r.value", "lt", "200", 1)
	if sql != "r.value < $1" {
		t.Errorf("SQL = %q, want %q", sql, "r.value < $1")
	}
	if len(args) != 1 || args[0] != "200" {
		t.Errorf("args = %v, want [200]", args)
	}
}

// ---------------------------------------------------------------------------
// compilePresentFilter tests
// ---------------------------------------------------------------------------

func TestCompilePresentFilter_Positive(t *testing.T) {
	sql := compilePresentFilter("r.name", false)
	if sql != "r.name IS NOT NULL" {
		t.Errorf("SQL = %q, want %q", sql, "r.name IS NOT NULL")
	}
}

func TestCompilePresentFilter_Negated(t *testing.T) {
	sql := compilePresentFilter("r.name", true)
	if sql != "r.name IS NULL" {
		t.Errorf("SQL = %q, want %q", sql, "r.name IS NULL")
	}
}

// ---------------------------------------------------------------------------
// DefaultFilterColumnMappings tests
// ---------------------------------------------------------------------------

func TestDefaultFilterColumnMappings_Patient(t *testing.T) {
	mappings := DefaultFilterColumnMappings("Patient")
	if mappings == nil {
		t.Fatal("expected non-nil mappings for Patient")
	}
	expected := []string{"name", "family", "given", "birthdate", "gender", "active"}
	for _, param := range expected {
		if _, ok := mappings[param]; !ok {
			t.Errorf("Patient mappings missing %q", param)
		}
	}
	// Check that name maps to a string type
	if m, ok := mappings["name"]; ok {
		if m.ParamType != "string" {
			t.Errorf("name ParamType = %q, want %q", m.ParamType, "string")
		}
	}
	// Check that birthdate maps to a date type
	if m, ok := mappings["birthdate"]; ok {
		if m.ParamType != "date" {
			t.Errorf("birthdate ParamType = %q, want %q", m.ParamType, "date")
		}
	}
}

func TestDefaultFilterColumnMappings_Observation(t *testing.T) {
	mappings := DefaultFilterColumnMappings("Observation")
	if mappings == nil {
		t.Fatal("expected non-nil mappings for Observation")
	}
	expected := []string{"code", "status", "date", "category", "value-quantity"}
	for _, param := range expected {
		if _, ok := mappings[param]; !ok {
			t.Errorf("Observation mappings missing %q", param)
		}
	}
	// Check token type for code with system column
	if m, ok := mappings["code"]; ok {
		if m.ParamType != "token" {
			t.Errorf("code ParamType = %q, want %q", m.ParamType, "token")
		}
		if m.SysColumn == "" {
			t.Error("code SysColumn should not be empty")
		}
	}
}

func TestDefaultFilterColumnMappings_Unknown(t *testing.T) {
	mappings := DefaultFilterColumnMappings("UnknownResource")
	if mappings == nil {
		t.Fatal("expected non-nil (empty) mappings for unknown resource")
	}
	if len(mappings) != 0 {
		t.Errorf("expected empty mappings for unknown resource, got %d", len(mappings))
	}
}

// ---------------------------------------------------------------------------
// FilterExpressionToString tests
// ---------------------------------------------------------------------------

func TestFilterExpressionToString_SimpleParam(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "name",
		Operator: "eq",
		Value:    "Smith",
	}
	got := FilterExpressionToString(expr)
	if got != `name eq "Smith"` {
		t.Errorf("got %q, want %q", got, `name eq "Smith"`)
	}
}

func TestFilterExpressionToString_Present(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "name",
		Operator: "pr",
	}
	got := FilterExpressionToString(expr)
	if got != "name pr" {
		t.Errorf("got %q, want %q", got, "name pr")
	}
}

func TestFilterExpressionToString_And(t *testing.T) {
	expr := &FilterExprNode{
		Type: FilterExprAnd,
		Left: &FilterExprNode{
			Type:     FilterExprParam,
			Param:    "name",
			Operator: "eq",
			Value:    "Smith",
		},
		Right: &FilterExprNode{
			Type:     FilterExprParam,
			Param:    "gender",
			Operator: "eq",
			Value:    "male",
		},
	}
	got := FilterExpressionToString(expr)
	expected := `name eq "Smith" and gender eq "male"`
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestFilterExpressionToString_Or(t *testing.T) {
	expr := &FilterExprNode{
		Type: FilterExprOr,
		Left: &FilterExprNode{
			Type:     FilterExprParam,
			Param:    "status",
			Operator: "eq",
			Value:    "active",
		},
		Right: &FilterExprNode{
			Type:     FilterExprParam,
			Param:    "status",
			Operator: "eq",
			Value:    "planned",
		},
	}
	got := FilterExpressionToString(expr)
	expected := `status eq "active" or status eq "planned"`
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestFilterExpressionToString_Not(t *testing.T) {
	expr := &FilterExprNode{
		Type: FilterExprNot,
		Child: &FilterExprNode{
			Type:     FilterExprParam,
			Param:    "status",
			Operator: "eq",
			Value:    "inactive",
		},
	}
	got := FilterExpressionToString(expr)
	expected := `not status eq "inactive"`
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestFilterExpressionToString_RoundTrip(t *testing.T) {
	input := `name eq "Smith" and birthdate ge "2000-01-01"`
	expr, err := ParseFilterExpression(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := FilterExpressionToString(expr)
	// Parse it back
	expr2, err := ParseFilterExpression(got)
	if err != nil {
		t.Fatalf("failed to re-parse: %v", err)
	}
	got2 := FilterExpressionToString(expr2)
	if got != got2 {
		t.Errorf("round-trip failed: first=%q, second=%q", got, got2)
	}
}

// ---------------------------------------------------------------------------
// SimplifyFilterExpression tests
// ---------------------------------------------------------------------------

func TestSimplifyFilterExpression_DoubleNegation(t *testing.T) {
	// not(not(x)) -> x
	inner := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "status",
		Operator: "eq",
		Value:    "active",
	}
	expr := &FilterExprNode{
		Type: FilterExprNot,
		Child: &FilterExprNode{
			Type:  FilterExprNot,
			Child: inner,
		},
	}
	result := SimplifyFilterExpression(expr)
	if result.Type != FilterExprParam {
		t.Errorf("Type = %d, want FilterExprParam (double negation removed)", result.Type)
	}
	if result.Param != "status" {
		t.Errorf("Param = %q, want %q", result.Param, "status")
	}
}

func TestSimplifyFilterExpression_AlreadySimple(t *testing.T) {
	expr := &FilterExprNode{
		Type:     FilterExprParam,
		Param:    "name",
		Operator: "eq",
		Value:    "Smith",
	}
	result := SimplifyFilterExpression(expr)
	if result.Type != FilterExprParam || result.Param != "name" {
		t.Errorf("simplification changed a simple expression: %+v", result)
	}
}

func TestSimplifyFilterExpression_NestedDoubleNegation(t *testing.T) {
	// a and not(not(b)) -> a and b
	expr := &FilterExprNode{
		Type: FilterExprAnd,
		Left: &FilterExprNode{
			Type:     FilterExprParam,
			Param:    "a",
			Operator: "eq",
			Value:    "1",
		},
		Right: &FilterExprNode{
			Type: FilterExprNot,
			Child: &FilterExprNode{
				Type: FilterExprNot,
				Child: &FilterExprNode{
					Type:     FilterExprParam,
					Param:    "b",
					Operator: "eq",
					Value:    "2",
				},
			},
		},
	}
	result := SimplifyFilterExpression(expr)
	if result.Type != FilterExprAnd {
		t.Fatalf("Type = %d, want FilterExprAnd", result.Type)
	}
	if result.Right.Type != FilterExprParam {
		t.Errorf("Right.Type = %d, want FilterExprParam (double negation removed)", result.Right.Type)
	}
	if result.Right.Param != "b" {
		t.Errorf("Right.Param = %q, want %q", result.Right.Param, "b")
	}
}

// ---------------------------------------------------------------------------
// ApplyFilterParam tests
// ---------------------------------------------------------------------------

func TestApplyFilterParam_Basic(t *testing.T) {
	q := NewSearchQuery("patients", "*")
	ctx := testFilterContext()
	err := ApplyFilterParam(q, `name eq "Smith"`, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := q.CountSQL()
	if !strings.Contains(sql, "last_name") {
		t.Errorf("SQL = %q, want it to contain last_name", sql)
	}
}

func TestApplyFilterParam_Complex(t *testing.T) {
	q := NewSearchQuery("patients", "*")
	ctx := testFilterContext()
	err := ApplyFilterParam(q, `name eq "Smith" and gender eq "male"`, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := q.CountSQL()
	if !strings.Contains(sql, "last_name") {
		t.Errorf("SQL = %q, want it to contain last_name", sql)
	}
	if !strings.Contains(sql, "gender") {
		t.Errorf("SQL = %q, want it to contain gender", sql)
	}
	if !strings.Contains(sql, "AND") {
		t.Errorf("SQL = %q, want it to contain AND", sql)
	}
}

func TestApplyFilterParam_Invalid(t *testing.T) {
	q := NewSearchQuery("patients", "*")
	ctx := testFilterContext()
	err := ApplyFilterParam(q, "", ctx)
	if err == nil {
		t.Fatal("expected error for empty filter, got nil")
	}
}

func TestApplyFilterParam_UnknownField(t *testing.T) {
	q := NewSearchQuery("patients", "*")
	ctx := testFilterContext()
	err := ApplyFilterParam(q, `unknown_field eq "test"`, ctx)
	if err == nil {
		t.Fatal("expected error for unknown field, got nil")
	}
}

func TestApplyFilterParam_OrExpression(t *testing.T) {
	q := NewSearchQuery("patients", "*")
	ctx := testFilterContext()
	err := ApplyFilterParam(q, `status eq active or status eq planned`, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := q.CountSQL()
	if !strings.Contains(sql, "OR") {
		t.Errorf("SQL = %q, want it to contain OR", sql)
	}
}

// ---------------------------------------------------------------------------
// Edge case tests
// ---------------------------------------------------------------------------

func TestCompileFilterToSQL_DeeplyNested(t *testing.T) {
	// ((a and b) or (c and d)) and e
	expr := &FilterExprNode{
		Type: FilterExprAnd,
		Left: &FilterExprNode{
			Type: FilterExprOr,
			Left: &FilterExprNode{
				Type: FilterExprAnd,
				Left: &FilterExprNode{
					Type: FilterExprParam, Param: "name", Operator: "eq", Value: "A",
				},
				Right: &FilterExprNode{
					Type: FilterExprParam, Param: "name", Operator: "eq", Value: "B",
				},
			},
			Right: &FilterExprNode{
				Type: FilterExprAnd,
				Left: &FilterExprNode{
					Type: FilterExprParam, Param: "name", Operator: "eq", Value: "C",
				},
				Right: &FilterExprNode{
					Type: FilterExprParam, Param: "name", Operator: "eq", Value: "D",
				},
			},
		},
		Right: &FilterExprNode{
			Type: FilterExprParam, Param: "name", Operator: "eq", Value: "E",
		},
	}
	ctx := testFilterContext()
	sql, args, err := CompileFilterToSQL(expr, ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(args) != 5 {
		t.Errorf("expected 5 args, got %d", len(args))
	}
	if !strings.Contains(sql, "$5") {
		t.Errorf("SQL = %q, want it to reference $5", sql)
	}
}

func TestParseFilterExpression_AllOperators(t *testing.T) {
	operators := []string{"eq", "ne", "gt", "lt", "ge", "le", "co", "sw", "ew", "sa", "eb", "ap", "in", "ni", "ss", "sb"}
	for _, op := range operators {
		t.Run(op, func(t *testing.T) {
			input := "name " + op + " value"
			if op == "pr" {
				input = "name pr"
			}
			expr, err := ParseFilterExpression(input)
			if err != nil {
				t.Fatalf("failed to parse with operator %q: %v", op, err)
			}
			if expr.Operator != op {
				t.Errorf("Operator = %q, want %q", expr.Operator, op)
			}
		})
	}
}

func TestParseFilterExpression_SpecialCharsInValue(t *testing.T) {
	expr, err := ParseFilterExpression(`code eq "http://loinc.org|8867-4"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expr.Value != "http://loinc.org|8867-4" {
		t.Errorf("Value = %q, want %q", expr.Value, "http://loinc.org|8867-4")
	}
}

func TestParseFilterExpression_VeryLongExpression(t *testing.T) {
	// Build a long expression: a eq 1 and b eq 2 and ... (10 terms)
	parts := make([]string, 10)
	for i := 0; i < 10; i++ {
		parts[i] = "name eq value"
	}
	input := strings.Join(parts, " and ")
	expr, err := ParseFilterExpression(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The expression should be a left-associative tree of ANDs
	if expr.Type != FilterExprAnd {
		t.Errorf("root Type = %d, want FilterExprAnd", expr.Type)
	}
}

func TestTokenizeFilter_ConsecutiveSpaces(t *testing.T) {
	tokens, err := tokenizeFilter(`name   eq   "Smith"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0].Value != "name" || tokens[1].Value != "eq" || tokens[2].Value != "Smith" {
		t.Errorf("unexpected tokens: %v", tokens)
	}
}

func TestFilterExpressionToString_Nil(t *testing.T) {
	got := FilterExpressionToString(nil)
	if got != "" {
		t.Errorf("got %q, want empty string for nil expression", got)
	}
}

func TestSimplifyFilterExpression_Nil(t *testing.T) {
	result := SimplifyFilterExpression(nil)
	if result != nil {
		t.Errorf("expected nil result for nil input, got %+v", result)
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func testFilterContext() *FilterContext {
	return &FilterContext{
		ResourceType: "Patient",
		ColumnMappings: map[string]FilterColumnMapping{
			"name":      {Column: "last_name", ParamType: "string"},
			"family":    {Column: "last_name", ParamType: "string"},
			"given":     {Column: "first_name", ParamType: "string"},
			"birthdate": {Column: "birth_date", ParamType: "date"},
			"gender":    {Column: "gender", ParamType: "token"},
			"status":    {Column: "status", ParamType: "token"},
			"active":    {Column: "active", ParamType: "token"},
			"code":      {Column: "code_value", SysColumn: "code_system", ParamType: "token"},
		},
		TableAlias: "r",
	}
}

func testFilterContextObs() *FilterContext {
	return &FilterContext{
		ResourceType: "Observation",
		ColumnMappings: map[string]FilterColumnMapping{
			"code":           {Column: "code_value", SysColumn: "code_system", ParamType: "token"},
			"status":         {Column: "status", ParamType: "token"},
			"date":           {Column: "effective_date", ParamType: "date"},
			"category":       {Column: "category", ParamType: "token"},
			"value-quantity":  {Column: "value_quantity", ParamType: "number"},
			"value_quantity": {Column: "value_quantity", ParamType: "number"},
		},
		TableAlias: "r",
	}
}
