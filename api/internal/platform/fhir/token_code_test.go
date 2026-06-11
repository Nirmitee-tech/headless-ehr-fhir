package fhir

import "testing"

func TestTokenCode(t *testing.T) {
	cases := map[string]string{
		"http://loinc.org|1234": "1234",
		"|1234":                 "1234",
		"1234":                  "1234",
		"http://loinc.org|":     "",
		"":                      "",
		"laboratory":            "laboratory",
		"http://terminology.hl7.org/CodeSystem/observation-category|laboratory": "laboratory",
	}
	for in, want := range cases {
		if got := tokenCode(in); got != want {
			t.Errorf("tokenCode(%q) = %q, want %q", in, got, want)
		}
	}
}
