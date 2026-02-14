package terminology

import (
	"encoding/json"
	"strings"
	"testing"
)

func ptrStr(s string) *string       { return &s }
func ptrInt(i int) *int             { return &i }
func ptrFloat(f float64) *float64   { return &f }
func ptrBool(b bool) *bool          { return &b }

func TestLOINCCode_JSONRoundTrip(t *testing.T) {
	original := &LOINCCode{
		Code:       "2093-3",
		Display:    "Cholesterol [Mass/volume] in Serum or Plasma",
		Component:  "Cholesterol",
		Property:   "MCnc",
		TimeAspect: "Pt",
		SystemURI:  SystemLOINC,
		Category:   "Chemistry",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded LOINCCode
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Code != original.Code {
		t.Errorf("Code mismatch: got %q, want %q", decoded.Code, original.Code)
	}
	if decoded.Display != original.Display {
		t.Errorf("Display mismatch: got %q, want %q", decoded.Display, original.Display)
	}
	if decoded.Component != original.Component {
		t.Errorf("Component mismatch: got %q, want %q", decoded.Component, original.Component)
	}
	if decoded.Property != original.Property {
		t.Errorf("Property mismatch: got %q, want %q", decoded.Property, original.Property)
	}
	if decoded.TimeAspect != original.TimeAspect {
		t.Errorf("TimeAspect mismatch: got %q, want %q", decoded.TimeAspect, original.TimeAspect)
	}
	if decoded.SystemURI != original.SystemURI {
		t.Errorf("SystemURI mismatch: got %q, want %q", decoded.SystemURI, original.SystemURI)
	}
	if decoded.Category != original.Category {
		t.Errorf("Category mismatch: got %q, want %q", decoded.Category, original.Category)
	}
}

func TestLOINCCode_OmitEmptyFields(t *testing.T) {
	m := &LOINCCode{
		Code:      "2093-3",
		Display:   "Cholesterol",
		SystemURI: SystemLOINC,
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	// Component, Property, TimeAspect, Category have omitempty on json tags
	if strings.Contains(s, `"component"`) {
		t.Error("empty Component should be omitted")
	}
	if strings.Contains(s, `"property"`) {
		t.Error("empty Property should be omitted")
	}
	if strings.Contains(s, `"time_aspect"`) {
		t.Error("empty TimeAspect should be omitted")
	}
	if strings.Contains(s, `"category"`) {
		t.Error("empty Category should be omitted")
	}
}

func TestICD10Code_JSONRoundTrip(t *testing.T) {
	original := &ICD10Code{
		Code:      "E11.9",
		Display:   "Type 2 diabetes mellitus without complications",
		Category:  "Endocrine, nutritional and metabolic diseases",
		Chapter:   "IV",
		SystemURI: SystemICD10,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded ICD10Code
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Code != original.Code {
		t.Errorf("Code mismatch: got %q, want %q", decoded.Code, original.Code)
	}
	if decoded.Display != original.Display {
		t.Errorf("Display mismatch: got %q, want %q", decoded.Display, original.Display)
	}
	if decoded.Category != original.Category {
		t.Errorf("Category mismatch: got %q, want %q", decoded.Category, original.Category)
	}
	if decoded.Chapter != original.Chapter {
		t.Errorf("Chapter mismatch: got %q, want %q", decoded.Chapter, original.Chapter)
	}
	if decoded.SystemURI != original.SystemURI {
		t.Errorf("SystemURI mismatch: got %q, want %q", decoded.SystemURI, original.SystemURI)
	}
}

func TestICD10Code_OmitEmptyFields(t *testing.T) {
	m := &ICD10Code{
		Code:      "E11.9",
		Display:   "Type 2 diabetes mellitus without complications",
		SystemURI: SystemICD10,
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"category"`) {
		t.Error("empty Category should be omitted")
	}
	if strings.Contains(s, `"chapter"`) {
		t.Error("empty Chapter should be omitted")
	}
}

func TestSearchResult_JSONRoundTrip(t *testing.T) {
	original := &SearchResult{
		Code:      "80146002",
		Display:   "Appendectomy",
		SystemURI: SystemSNOMED,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded SearchResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Code != original.Code {
		t.Errorf("Code mismatch: got %q, want %q", decoded.Code, original.Code)
	}
	if decoded.Display != original.Display {
		t.Errorf("Display mismatch: got %q, want %q", decoded.Display, original.Display)
	}
	if decoded.SystemURI != original.SystemURI {
		t.Errorf("SystemURI mismatch: got %q, want %q", decoded.SystemURI, original.SystemURI)
	}
}

func TestSearchResult_JSONFieldNames(t *testing.T) {
	sr := &SearchResult{
		Code:      "12345",
		Display:   "Test Code",
		SystemURI: "http://example.com",
	}

	data, err := json.Marshal(sr)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	// Verify the JSON field names match the struct tags
	if !strings.Contains(s, `"code"`) {
		t.Error("expected 'code' field in JSON output")
	}
	if !strings.Contains(s, `"display"`) {
		t.Error("expected 'display' field in JSON output")
	}
	if !strings.Contains(s, `"system"`) {
		t.Error("expected 'system' field in JSON output")
	}
	// SystemURI maps to "system" not "system_uri"
	if strings.Contains(s, `"system_uri"`) {
		t.Error("SystemURI should map to 'system' not 'system_uri'")
	}
}
