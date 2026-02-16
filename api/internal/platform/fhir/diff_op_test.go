package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestDiffResources_IdenticalMaps(t *testing.T) {
	a := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"active":       true,
	}
	b := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"active":       true,
	}

	diffs := DiffResources(a, b)
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs for identical maps, got %d: %+v", len(diffs), diffs)
	}
}

func TestDiffResources_AddedFields(t *testing.T) {
	old := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
	}
	new := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"active":       true,
		"gender":       "male",
	}

	diffs := DiffResources(old, new)

	added := filterByType(diffs, "added")
	if len(added) != 2 {
		t.Fatalf("expected 2 added entries, got %d: %+v", len(added), added)
	}

	paths := map[string]bool{}
	for _, d := range added {
		paths[d.Path] = true
		if d.OldValue != nil {
			t.Errorf("added entry %q should have nil OldValue, got %v", d.Path, d.OldValue)
		}
		if d.NewValue == nil {
			t.Errorf("added entry %q should have non-nil NewValue", d.Path)
		}
	}
	if !paths["active"] {
		t.Error("expected 'active' in added paths")
	}
	if !paths["gender"] {
		t.Error("expected 'gender' in added paths")
	}
}

func TestDiffResources_RemovedFields(t *testing.T) {
	old := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"active":       true,
		"gender":       "male",
	}
	new := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
	}

	diffs := DiffResources(old, new)

	removed := filterByType(diffs, "removed")
	if len(removed) != 2 {
		t.Fatalf("expected 2 removed entries, got %d: %+v", len(removed), removed)
	}

	paths := map[string]bool{}
	for _, d := range removed {
		paths[d.Path] = true
		if d.NewValue != nil {
			t.Errorf("removed entry %q should have nil NewValue, got %v", d.Path, d.NewValue)
		}
		if d.OldValue == nil {
			t.Errorf("removed entry %q should have non-nil OldValue", d.Path)
		}
	}
	if !paths["active"] {
		t.Error("expected 'active' in removed paths")
	}
	if !paths["gender"] {
		t.Error("expected 'gender' in removed paths")
	}
}

func TestDiffResources_ChangedValues(t *testing.T) {
	old := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"active":       true,
		"gender":       "male",
	}
	new := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"active":       false,
		"gender":       "female",
	}

	diffs := DiffResources(old, new)

	changed := filterByType(diffs, "changed")
	if len(changed) != 2 {
		t.Fatalf("expected 2 changed entries, got %d: %+v", len(changed), changed)
	}

	for _, d := range changed {
		if d.OldValue == nil {
			t.Errorf("changed entry %q should have non-nil OldValue", d.Path)
		}
		if d.NewValue == nil {
			t.Errorf("changed entry %q should have non-nil NewValue", d.Path)
		}
	}
}

func TestDiffResources_NestedChanges(t *testing.T) {
	old := map[string]interface{}{
		"resourceType": "Patient",
		"name": map[string]interface{}{
			"family": "Smith",
			"given":  "John",
		},
	}
	new := map[string]interface{}{
		"resourceType": "Patient",
		"name": map[string]interface{}{
			"family": "Doe",
			"given":  "John",
		},
	}

	diffs := DiffResources(old, new)

	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff for nested change, got %d: %+v", len(diffs), diffs)
	}

	d := diffs[0]
	if d.Path != "name.family" {
		t.Errorf("path = %q, want 'name.family'", d.Path)
	}
	if d.Type != "changed" {
		t.Errorf("type = %q, want 'changed'", d.Type)
	}
	if d.OldValue != "Smith" {
		t.Errorf("oldValue = %v, want 'Smith'", d.OldValue)
	}
	if d.NewValue != "Doe" {
		t.Errorf("newValue = %v, want 'Doe'", d.NewValue)
	}
}

func TestDiffResources_NestedAdded(t *testing.T) {
	old := map[string]interface{}{
		"resourceType": "Patient",
		"name": map[string]interface{}{
			"family": "Smith",
		},
	}
	new := map[string]interface{}{
		"resourceType": "Patient",
		"name": map[string]interface{}{
			"family": "Smith",
			"given":  "John",
		},
	}

	diffs := DiffResources(old, new)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d: %+v", len(diffs), diffs)
	}
	if diffs[0].Path != "name.given" {
		t.Errorf("path = %q, want 'name.given'", diffs[0].Path)
	}
	if diffs[0].Type != "added" {
		t.Errorf("type = %q, want 'added'", diffs[0].Type)
	}
}

func TestDiffResources_ArrayChanges(t *testing.T) {
	old := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith"},
			map[string]interface{}{"family": "Jones"},
		},
	}
	new := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith"},
			map[string]interface{}{"family": "Doe"},
			map[string]interface{}{"family": "Brown"},
		},
	}

	diffs := DiffResources(old, new)

	// name[0] is the same, name[1].family changed, name[2] is added.
	if len(diffs) != 2 {
		t.Fatalf("expected 2 diffs for array changes, got %d: %+v", len(diffs), diffs)
	}

	// Verify the changed element.
	foundChanged := false
	foundAdded := false
	for _, d := range diffs {
		if d.Path == "name[1].family" && d.Type == "changed" {
			foundChanged = true
			if d.OldValue != "Jones" {
				t.Errorf("oldValue = %v, want 'Jones'", d.OldValue)
			}
			if d.NewValue != "Doe" {
				t.Errorf("newValue = %v, want 'Doe'", d.NewValue)
			}
		}
		if d.Path == "name[2]" && d.Type == "added" {
			foundAdded = true
		}
	}
	if !foundChanged {
		t.Error("expected changed entry at name[1].family")
	}
	if !foundAdded {
		t.Error("expected added entry at name[2]")
	}
}

func TestDiffResources_ArrayShorter(t *testing.T) {
	old := map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
	}
	new := map[string]interface{}{
		"items": []interface{}{"a"},
	}

	diffs := DiffResources(old, new)

	removed := filterByType(diffs, "removed")
	if len(removed) != 2 {
		t.Fatalf("expected 2 removed entries for shorter array, got %d: %+v", len(removed), removed)
	}
}

func TestDiffResources_EmptyMaps(t *testing.T) {
	old := map[string]interface{}{}
	new := map[string]interface{}{}

	diffs := DiffResources(old, new)
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs for two empty maps, got %d", len(diffs))
	}
}

func TestDiffResources_MixedChanges(t *testing.T) {
	old := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"active":       true,
		"birthDate":    "1990-01-01",
	}
	new := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"active":       false,
		"gender":       "male",
	}

	diffs := DiffResources(old, new)

	added := filterByType(diffs, "added")
	removed := filterByType(diffs, "removed")
	changed := filterByType(diffs, "changed")

	if len(added) != 1 {
		t.Errorf("expected 1 added, got %d", len(added))
	}
	if len(removed) != 1 {
		t.Errorf("expected 1 removed, got %d", len(removed))
	}
	if len(changed) != 1 {
		t.Errorf("expected 1 changed, got %d", len(changed))
	}
}

func TestDiffToParameters_Format(t *testing.T) {
	diffs := []DiffEntry{
		{
			Path:     "active",
			Type:     "changed",
			OldValue: true,
			NewValue: false,
		},
		{
			Path:     "gender",
			Type:     "added",
			NewValue: "male",
		},
		{
			Path:     "birthDate",
			Type:     "removed",
			OldValue: "1990-01-01",
		},
	}

	result := DiffToParameters(diffs)

	if result["resourceType"] != "Parameters" {
		t.Errorf("resourceType = %v, want 'Parameters'", result["resourceType"])
	}

	params, ok := result["parameter"].([]interface{})
	if !ok {
		t.Fatalf("parameter is not []interface{}")
	}
	if len(params) != 3 {
		t.Fatalf("expected 3 parameters, got %d", len(params))
	}

	// Validate the first parameter (changed).
	first, ok := params[0].(map[string]interface{})
	if !ok {
		t.Fatal("first parameter is not map")
	}
	if first["name"] != "diff" {
		t.Errorf("first parameter name = %v, want 'diff'", first["name"])
	}

	parts, ok := first["part"].([]interface{})
	if !ok {
		t.Fatal("parts is not []interface{}")
	}
	// Changed entry should have path, type, oldValue, newValue = 4 parts.
	if len(parts) != 4 {
		t.Errorf("expected 4 parts for changed entry, got %d", len(parts))
	}

	// Validate the added parameter has no oldValue (3 parts).
	second, _ := params[1].(map[string]interface{})
	addedParts, _ := second["part"].([]interface{})
	if len(addedParts) != 3 {
		t.Errorf("expected 3 parts for added entry, got %d", len(addedParts))
	}

	// Validate the removed parameter has no newValue (3 parts).
	third, _ := params[2].(map[string]interface{})
	removedParts, _ := third["part"].([]interface{})
	if len(removedParts) != 3 {
		t.Errorf("expected 3 parts for removed entry, got %d", len(removedParts))
	}
}

func TestDiffToParameters_Empty(t *testing.T) {
	result := DiffToParameters(nil)

	if result["resourceType"] != "Parameters" {
		t.Errorf("resourceType = %v, want 'Parameters'", result["resourceType"])
	}
	params, ok := result["parameter"].([]interface{})
	if !ok {
		t.Fatal("parameter is not []interface{}")
	}
	if len(params) != 0 {
		t.Errorf("expected 0 parameters for empty diff, got %d", len(params))
	}
}

func TestDiffToParameters_JSON(t *testing.T) {
	diffs := []DiffEntry{
		{Path: "status", Type: "changed", OldValue: "active", NewValue: "inactive"},
	}

	result := DiffToParameters(diffs)
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal Parameters: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal Parameters: %v", err)
	}
	if parsed["resourceType"] != "Parameters" {
		t.Errorf("resourceType after round-trip = %v", parsed["resourceType"])
	}
}

func TestDiffHandler_MissingFromParam(t *testing.T) {
	repo := NewHistoryRepository()
	handler := DiffHandler(repo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123/$diff", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType", "id")
	c.SetParamValues("Patient", "123")

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var outcome OperationOutcome
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(outcome.Issue) == 0 {
		t.Fatal("expected at least one issue")
	}
	if outcome.Issue[0].Code != IssueTypeRequired {
		t.Errorf("issue code = %q, want %q", outcome.Issue[0].Code, IssueTypeRequired)
	}
}

func TestDiffHandler_InvalidFromParam(t *testing.T) {
	repo := NewHistoryRepository()
	handler := DiffHandler(repo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123/$diff?from=abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType", "id")
	c.SetParamValues("Patient", "123")

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDiffHandler_InvalidToParam(t *testing.T) {
	repo := NewHistoryRepository()
	handler := DiffHandler(repo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123/$diff?from=1&to=xyz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType", "id")
	c.SetParamValues("Patient", "123")

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDiffHandler_NegativeFromParam(t *testing.T) {
	repo := NewHistoryRepository()
	handler := DiffHandler(repo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123/$diff?from=-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType", "id")
	c.SetParamValues("Patient", "123")

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDiffHandler_ZeroFromParam(t *testing.T) {
	repo := NewHistoryRepository()
	handler := DiffHandler(repo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123/$diff?from=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType", "id")
	c.SetParamValues("Patient", "123")

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDiffHandler_SameVersion(t *testing.T) {
	// When from == to, we expect an empty Parameters response without hitting the DB.
	// Since there is no DB, this tests that the short-circuit path works.
	repo := NewHistoryRepository()
	handler := DiffHandler(repo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123/$diff?from=1&to=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType", "id")
	c.SetParamValues("Patient", "123")

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["resourceType"] != "Parameters" {
		t.Errorf("resourceType = %v, want 'Parameters'", result["resourceType"])
	}

	params, ok := result["parameter"].([]interface{})
	if !ok {
		t.Fatal("parameter is not []interface{}")
	}
	if len(params) != 0 {
		t.Errorf("expected 0 parameters for same version diff, got %d", len(params))
	}
}

func TestDiffHandler_NoDBReturnsNotFound(t *testing.T) {
	// With no DB in context, fetching versions should fail with 404.
	repo := NewHistoryRepository()
	handler := DiffHandler(repo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123/$diff?from=1&to=2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType", "id")
	c.SetParamValues("Patient", "123")

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestDiffHandler_NoDBDefaultToReturnsNotFound(t *testing.T) {
	// When 'to' is omitted the handler tries to find the latest version, which
	// fails without a DB, resulting in a 404.
	repo := NewHistoryRepository()
	handler := DiffHandler(repo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123/$diff?from=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType", "id")
	c.SetParamValues("Patient", "123")

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestDiffResources_NumericValues(t *testing.T) {
	old := map[string]interface{}{
		"count": float64(10),
	}
	new := map[string]interface{}{
		"count": float64(20),
	}

	diffs := DiffResources(old, new)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Type != "changed" {
		t.Errorf("type = %q, want 'changed'", diffs[0].Type)
	}
	if diffs[0].OldValue != float64(10) {
		t.Errorf("oldValue = %v, want 10", diffs[0].OldValue)
	}
	if diffs[0].NewValue != float64(20) {
		t.Errorf("newValue = %v, want 20", diffs[0].NewValue)
	}
}

func TestDiffResources_TypeMismatch(t *testing.T) {
	// When a scalar value becomes a map, it should be reported as changed.
	old := map[string]interface{}{
		"value": "simple",
	}
	new := map[string]interface{}{
		"value": map[string]interface{}{"nested": "complex"},
	}

	diffs := DiffResources(old, new)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff for type mismatch, got %d: %+v", len(diffs), diffs)
	}
	if diffs[0].Type != "changed" {
		t.Errorf("type = %q, want 'changed'", diffs[0].Type)
	}
}

func TestDiffResources_DeeplyNested(t *testing.T) {
	old := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": "old_value",
			},
		},
	}
	new := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": "new_value",
			},
		},
	}

	diffs := DiffResources(old, new)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Path != "level1.level2.level3" {
		t.Errorf("path = %q, want 'level1.level2.level3'", diffs[0].Path)
	}
}

func TestDiffToParameters_PartStructure(t *testing.T) {
	diffs := []DiffEntry{
		{Path: "name.family", Type: "changed", OldValue: "Smith", NewValue: "Doe"},
	}

	result := DiffToParameters(diffs)
	params := result["parameter"].([]interface{})
	param := params[0].(map[string]interface{})
	parts := param["part"].([]interface{})

	// Verify each part has the expected name and value.
	expectedParts := map[string]string{
		"path":     "name.family",
		"type":     "changed",
		"oldValue": "Smith",
		"newValue": "Doe",
	}

	for _, p := range parts {
		part := p.(map[string]interface{})
		name := part["name"].(string)
		value := part["valueString"].(string)
		if expected, ok := expectedParts[name]; ok {
			if value != expected {
				t.Errorf("part %q = %q, want %q", name, value, expected)
			}
		} else {
			t.Errorf("unexpected part name %q", name)
		}
	}
}

// filterByType returns only diffs with the given type.
func filterByType(diffs []DiffEntry, typ string) []DiffEntry {
	var result []DiffEntry
	for _, d := range diffs {
		if d.Type == typ {
			result = append(result, d)
		}
	}
	return result
}
