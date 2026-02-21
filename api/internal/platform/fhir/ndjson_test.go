package fhir

import (
	"bufio"
	"bytes"
	"encoding/json"
	"testing"
)

func TestNDJSONWriter_SingleResource(t *testing.T) {
	var buf bytes.Buffer
	w := NewNDJSONWriter(&buf)

	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"active":       true,
	}

	if err := w.WriteResource(resource); err != nil {
		t.Fatalf("WriteResource failed: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Parse the output
	lines := scanNDJSON(t, buf.Bytes())
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}

	if lines[0]["resourceType"] != "Patient" {
		t.Errorf("expected resourceType 'Patient', got %v", lines[0]["resourceType"])
	}
	if lines[0]["id"] != "p1" {
		t.Errorf("expected id 'p1', got %v", lines[0]["id"])
	}
	if lines[0]["active"] != true {
		t.Errorf("expected active true, got %v", lines[0]["active"])
	}
}

func TestNDJSONWriter_MultipleResources(t *testing.T) {
	var buf bytes.Buffer
	w := NewNDJSONWriter(&buf)

	resources := []map[string]interface{}{
		{"resourceType": "Patient", "id": "p1", "name": "Alice"},
		{"resourceType": "Patient", "id": "p2", "name": "Bob"},
		{"resourceType": "Patient", "id": "p3", "name": "Charlie"},
	}

	for _, r := range resources {
		if err := w.WriteResource(r); err != nil {
			t.Fatalf("WriteResource failed: %v", err)
		}
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	lines := scanNDJSON(t, buf.Bytes())
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	expectedIDs := []string{"p1", "p2", "p3"}
	for i, expected := range expectedIDs {
		if lines[i]["id"] != expected {
			t.Errorf("line %d: expected id %q, got %v", i, expected, lines[i]["id"])
		}
	}
}

func TestNDJSONWriter_Flush(t *testing.T) {
	var buf bytes.Buffer
	w := NewNDJSONWriter(&buf)

	// Before Flush, the buffered writer may not have written to buf yet
	resource := map[string]interface{}{"resourceType": "Observation", "id": "o1"}
	if err := w.WriteResource(resource); err != nil {
		t.Fatalf("WriteResource failed: %v", err)
	}

	// Flush should push data to the underlying writer
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected non-empty buffer after Flush")
	}

	lines := scanNDJSON(t, buf.Bytes())
	if len(lines) != 1 {
		t.Fatalf("expected 1 line after Flush, got %d", len(lines))
	}
	if lines[0]["id"] != "o1" {
		t.Errorf("expected id 'o1', got %v", lines[0]["id"])
	}
}

func TestNDJSONWriter_EmptyFlush(t *testing.T) {
	var buf bytes.Buffer
	w := NewNDJSONWriter(&buf)

	// Flush with nothing written should succeed and produce empty output
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %d bytes", buf.Len())
	}
}

func TestNDJSONWriter_EachLineIsValidJSON(t *testing.T) {
	var buf bytes.Buffer
	w := NewNDJSONWriter(&buf)

	for i := 0; i < 10; i++ {
		r := map[string]interface{}{
			"resourceType": "Condition",
			"id":           i,
		}
		if err := w.WriteResource(r); err != nil {
			t.Fatalf("WriteResource failed: %v", err)
		}
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Every non-empty line must be valid JSON
	scanner := bufio.NewScanner(&buf)
	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if !json.Valid([]byte(line)) {
			t.Errorf("line %d is not valid JSON: %s", lineNum, line)
		}
		lineNum++
	}
	if lineNum != 10 {
		t.Errorf("expected 10 JSON lines, got %d", lineNum)
	}
}

func TestNDJSONWriter_StructResource(t *testing.T) {
	// NDJSONWriter should work with typed structs, not just maps
	type simpleResource struct {
		ResourceType string `json:"resourceType"`
		ID           string `json:"id"`
		Status       string `json:"status"`
	}

	var buf bytes.Buffer
	w := NewNDJSONWriter(&buf)

	r := simpleResource{
		ResourceType: "MedicationRequest",
		ID:           "mr-1",
		Status:       "active",
	}
	if err := w.WriteResource(r); err != nil {
		t.Fatalf("WriteResource failed: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	lines := scanNDJSON(t, buf.Bytes())
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0]["resourceType"] != "MedicationRequest" {
		t.Errorf("expected resourceType 'MedicationRequest', got %v", lines[0]["resourceType"])
	}
	if lines[0]["status"] != "active" {
		t.Errorf("expected status 'active', got %v", lines[0]["status"])
	}
}

func TestNDJSONWriter_MarshalError(t *testing.T) {
	var buf bytes.Buffer
	w := NewNDJSONWriter(&buf)

	// Channels cannot be marshalled to JSON
	badResource := map[string]interface{}{
		"badField": make(chan int),
	}

	err := w.WriteResource(badResource)
	if err == nil {
		t.Error("expected error for un-marshallable resource")
	}
}

// scanNDJSON parses NDJSON bytes into a slice of maps.
func scanNDJSON(t *testing.T, data []byte) []map[string]interface{} {
	t.Helper()
	var results []map[string]interface{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var resource map[string]interface{}
		if err := json.Unmarshal([]byte(line), &resource); err != nil {
			t.Fatalf("invalid NDJSON line: %v\nline: %s", err, line)
		}
		results = append(results, resource)
	}
	return results
}
