package fhir

import (
	"strings"
	"testing"
)

// Verify comma-separated token values become an OR (IN clause) for code-only
// params, and an OR of sys|code clauses when a system column exists.
func TestApplyParam_TokenCommaOR_NoSysColumn(t *testing.T) {
	q := NewSearchQuery("medication_request", "*")
	q.ApplyParam(SearchParamConfig{Type: SearchParamToken, Column: "intent"}, "order,proposal,plan")
	sql := q.CountSQL()
	if !strings.Contains(sql, "intent IN ($1, $2, $3)") {
		t.Fatalf("expected IN clause with 3 placeholders, got: %s", sql)
	}
	args := q.CountArgs()
	if len(args) != 3 || args[0] != "order" || args[2] != "plan" {
		t.Fatalf("expected [order proposal plan], got %v", args)
	}
}

func TestApplyParam_TokenSingle_NoSysColumn(t *testing.T) {
	q := NewSearchQuery("care_team", "*")
	q.ApplyParam(SearchParamConfig{Type: SearchParamToken, Column: "status"}, "active")
	if !strings.Contains(q.CountSQL(), "status IN ($1)") {
		t.Fatalf("single value should still produce IN ($1), got: %s", q.CountSQL())
	}
}

func TestApplyParam_TokenCommaOR_WithSysColumn(t *testing.T) {
	q := NewSearchQuery("observation", "*")
	q.ApplyParam(SearchParamConfig{Type: SearchParamToken, Column: "code_value", SysColumn: "code_system"}, "1234,5678")
	sql := q.CountSQL()
	if !strings.Contains(sql, " OR ") || strings.Count(sql, "code_value") < 2 {
		t.Fatalf("expected OR of two sys|code clauses, got: %s", sql)
	}
}
