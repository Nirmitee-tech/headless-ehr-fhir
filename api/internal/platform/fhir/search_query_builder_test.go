package fhir

import (
	"strings"
	"testing"
)

func TestSearchQueryBasic(t *testing.T) {
	q := NewSearchQuery("encounter", "id, status")
	q.Add("patient_id = $1", "patient-123")
	q.OrderBy("created_at DESC")

	countSQL := q.CountSQL()
	if !strings.Contains(countSQL, "SELECT COUNT(*) FROM encounter WHERE 1=1 AND patient_id = $1") {
		t.Errorf("unexpected count SQL: %s", countSQL)
	}
	if len(q.CountArgs()) != 1 || q.CountArgs()[0] != "patient-123" {
		t.Errorf("unexpected count args: %v", q.CountArgs())
	}

	dataSQL := q.DataSQL(10, 0)
	if !strings.Contains(dataSQL, "ORDER BY created_at DESC") {
		t.Errorf("expected ORDER BY in data SQL: %s", dataSQL)
	}
	if !strings.Contains(dataSQL, "LIMIT $2 OFFSET $3") {
		t.Errorf("expected LIMIT/OFFSET in data SQL: %s", dataSQL)
	}

	dataArgs := q.DataArgs(10, 0)
	if len(dataArgs) != 3 || dataArgs[1] != 10 || dataArgs[2] != 0 {
		t.Errorf("unexpected data args: %v", dataArgs)
	}
}

func TestSearchQueryApplyParams(t *testing.T) {
	configs := map[string]SearchParamConfig{
		"patient":         {Type: SearchParamReference, Column: "patient_id"},
		"status":          {Type: SearchParamToken, Column: "status"},
		"code":            {Type: SearchParamToken, Column: "code_value", SysColumn: "code_system"},
		"date":            {Type: SearchParamDate, Column: "effective_date"},
		"name":            {Type: SearchParamString, Column: "name"},
		"value-quantity":  {Type: SearchParamNumber, Column: "value_quantity"},
	}

	t.Run("reference param strips ResourceType prefix", func(t *testing.T) {
		q := NewSearchQuery("observation", "id")
		q.ApplyParams(map[string]string{"patient": "Patient/abc-123"}, configs)
		if len(q.CountArgs()) != 1 || q.CountArgs()[0] != "abc-123" {
			t.Errorf("reference should strip prefix, got args: %v", q.CountArgs())
		}
	})

	t.Run("token param with system|code", func(t *testing.T) {
		q := NewSearchQuery("observation", "id")
		q.ApplyParams(map[string]string{"code": "http://loinc.org|1234-5"}, configs)
		args := q.CountArgs()
		if len(args) != 2 {
			t.Fatalf("expected 2 args for system|code, got %d: %v", len(args), args)
		}
		if args[0] != "http://loinc.org" || args[1] != "1234-5" {
			t.Errorf("unexpected token args: %v", args)
		}
	})

	t.Run("simple token param", func(t *testing.T) {
		q := NewSearchQuery("observation", "id")
		q.ApplyParams(map[string]string{"status": "final"}, configs)
		sql := q.CountSQL()
		if !strings.Contains(sql, "status = $1") {
			t.Errorf("expected exact match for simple token: %s", sql)
		}
	})

	t.Run("date param with prefix", func(t *testing.T) {
		q := NewSearchQuery("observation", "id")
		q.ApplyParams(map[string]string{"date": "gt2023-01-01"}, configs)
		sql := q.CountSQL()
		if !strings.Contains(sql, "effective_date >") {
			t.Errorf("expected > for gt prefix: %s", sql)
		}
	})

	t.Run("string param default prefix match", func(t *testing.T) {
		q := NewSearchQuery("patient", "id")
		q.ApplyParams(map[string]string{"name": "Smith"}, configs)
		sql := q.CountSQL()
		if !strings.Contains(sql, "ILIKE") {
			t.Errorf("expected ILIKE for string search: %s", sql)
		}
		args := q.CountArgs()
		if len(args) != 1 {
			t.Fatalf("expected 1 arg, got %d", len(args))
		}
		if args[0] != "Smith%" {
			t.Errorf("expected prefix match pattern, got: %v", args[0])
		}
	})

	t.Run("number param with prefix", func(t *testing.T) {
		q := NewSearchQuery("observation", "id")
		q.ApplyParams(map[string]string{"value-quantity": "ge100"}, configs)
		sql := q.CountSQL()
		if !strings.Contains(sql, "value_quantity >=") {
			t.Errorf("expected >= for ge prefix: %s", sql)
		}
	})

	t.Run("multiple params combined", func(t *testing.T) {
		q := NewSearchQuery("observation", "id")
		q.ApplyParams(map[string]string{
			"patient": "p1",
			"status":  "final",
		}, configs)
		sql := q.CountSQL()
		if !strings.Contains(sql, "AND") {
			t.Errorf("expected AND clauses: %s", sql)
		}
		if len(q.CountArgs()) != 2 {
			t.Errorf("expected 2 args, got %d", len(q.CountArgs()))
		}
	})

	t.Run("unknown param ignored", func(t *testing.T) {
		q := NewSearchQuery("observation", "id")
		q.ApplyParams(map[string]string{"unknown-param": "foo"}, configs)
		if len(q.CountArgs()) != 0 {
			t.Errorf("expected 0 args for unknown param, got %d", len(q.CountArgs()))
		}
	})
}

func TestSearchQueryIdx(t *testing.T) {
	q := NewSearchQuery("test", "id")
	if q.Idx() != 1 {
		t.Errorf("initial idx should be 1, got %d", q.Idx())
	}
	q.Add("a = $1", "v1")
	if q.Idx() != 2 {
		t.Errorf("idx should be 2 after one arg, got %d", q.Idx())
	}
	q.Add("b = $2 AND c = $3", "v2", "v3")
	if q.Idx() != 4 {
		t.Errorf("idx should be 4 after three args, got %d", q.Idx())
	}
}
