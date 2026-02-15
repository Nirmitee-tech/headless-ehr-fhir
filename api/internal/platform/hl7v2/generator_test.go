package hl7v2

import (
	"strings"
	"testing"
)

// =========== Test Helpers ===========

func testPatient() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           "patient-123",
		"name": []interface{}{
			map[string]interface{}{
				"family": "Doe",
				"given":  []interface{}{"John"},
			},
		},
		"birthDate": "1980-05-15",
		"gender":    "male",
		"identifier": []interface{}{
			map[string]interface{}{
				"value": "MRN12345",
				"type": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{
							"code": "MR",
						},
					},
				},
			},
		},
		"address": []interface{}{
			map[string]interface{}{
				"line":       []interface{}{"123 Main St"},
				"city":       "Springfield",
				"state":      "IL",
				"postalCode": "62701",
			},
		},
		"telecom": []interface{}{
			map[string]interface{}{
				"system": "phone",
				"value":  "555-555-1234",
			},
		},
	}
}

func testEncounter() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Encounter",
		"id":           "enc-001",
		"class": map[string]interface{}{
			"code": "IMP",
		},
		"status": "in-progress",
		"location": []interface{}{
			map[string]interface{}{
				"location": map[string]interface{}{
					"display": "ICU Room 101",
				},
			},
		},
		"participant": []interface{}{
			map[string]interface{}{
				"individual": map[string]interface{}{
					"display": "Dr. Robert Smith",
				},
			},
		},
	}
}

func testServiceRequest() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "ServiceRequest",
		"id":           "sr-001",
		"status":       "active",
		"intent":       "order",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://loinc.org",
					"code":    "85025",
					"display": "CBC",
				},
			},
		},
		"authoredOn": "2024-01-15T12:00:00Z",
	}
}

func testDiagnosticReport() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "DiagnosticReport",
		"id":           "dr-001",
		"status":       "final",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://loinc.org",
					"code":    "85025",
					"display": "CBC",
				},
			},
		},
		"effectiveDateTime": "2024-01-15T14:00:00Z",
	}
}

func testObservations() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"resourceType": "Observation",
			"id":           "obs-001",
			"status":       "final",
			"code": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system":  "http://loinc.org",
						"code":    "718-7",
						"display": "Hemoglobin",
					},
				},
			},
			"valueQuantity": map[string]interface{}{
				"value": 13.5,
				"unit":  "g/dL",
			},
			"referenceRange": []interface{}{
				map[string]interface{}{
					"low": map[string]interface{}{
						"value": 12.0,
						"unit":  "g/dL",
					},
					"high": map[string]interface{}{
						"value": 17.5,
						"unit":  "g/dL",
					},
				},
			},
		},
		{
			"resourceType": "Observation",
			"id":           "obs-002",
			"status":       "final",
			"code": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system":  "http://loinc.org",
						"code":    "4544-3",
						"display": "Hematocrit",
					},
				},
			},
			"valueQuantity": map[string]interface{}{
				"value": 40.1,
				"unit":  "%",
			},
			"referenceRange": []interface{}{
				map[string]interface{}{
					"low": map[string]interface{}{
						"value": 36.0,
						"unit":  "%",
					},
					"high": map[string]interface{}{
						"value": 53.0,
						"unit":  "%",
					},
				},
			},
		},
	}
}

// =========== ADT Tests ===========

func TestGenerateADT_A01(t *testing.T) {
	data, err := GenerateADT("A01", testPatient(), testEncounter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw := string(data)

	// Should start with MSH
	if !strings.HasPrefix(raw, "MSH|") {
		t.Error("expected message to start with MSH|")
	}

	// Should contain ADT^A01
	if !strings.Contains(raw, "ADT^A01") {
		t.Error("expected ADT^A01 in message")
	}

	// Should contain PID segment with patient name
	if !strings.Contains(raw, "PID|") {
		t.Error("expected PID segment")
	}
	if !strings.Contains(raw, "Doe^John") {
		t.Error("expected patient name Doe^John in PID")
	}

	// Should contain EVN segment
	if !strings.Contains(raw, "EVN|") {
		t.Error("expected EVN segment")
	}

	// Should contain PV1 segment
	if !strings.Contains(raw, "PV1|") {
		t.Error("expected PV1 segment")
	}
}

func TestGenerateADT_A03(t *testing.T) {
	data, err := GenerateADT("A03", testPatient(), testEncounter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw := string(data)
	if !strings.Contains(raw, "ADT^A03") {
		t.Error("expected ADT^A03 in message")
	}
}

func TestGenerateADT_A08(t *testing.T) {
	data, err := GenerateADT("A08", testPatient(), testEncounter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw := string(data)
	if !strings.Contains(raw, "ADT^A08") {
		t.Error("expected ADT^A08 in message")
	}
}

func TestGenerateADT_RoundTrip(t *testing.T) {
	data, err := GenerateADT("A01", testPatient(), testEncounter())
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}

	msg, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if msg.Type != "ADT^A01" {
		t.Errorf("expected Type 'ADT^A01', got %q", msg.Type)
	}
	if msg.Version != "2.5.1" {
		t.Errorf("expected Version '2.5.1', got %q", msg.Version)
	}

	family, given := msg.PatientName()
	if family != "Doe" {
		t.Errorf("expected family 'Doe', got %q", family)
	}
	if given != "John" {
		t.Errorf("expected given 'John', got %q", given)
	}

	dob := msg.DateOfBirth()
	if dob != "19800515" {
		t.Errorf("expected DOB '19800515', got %q", dob)
	}

	gender := msg.Gender()
	if gender != "M" {
		t.Errorf("expected Gender 'M', got %q", gender)
	}
}

// =========== ORM Tests ===========

func TestGenerateORM_NewOrder(t *testing.T) {
	data, err := GenerateORM(testServiceRequest(), testPatient())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw := string(data)
	if !strings.Contains(raw, "ORM^O01") {
		t.Error("expected ORM^O01 in message")
	}
	if !strings.Contains(raw, "ORC|") {
		t.Error("expected ORC segment")
	}
	if !strings.Contains(raw, "NW") {
		t.Error("expected NW (new order) in ORC")
	}
	if !strings.Contains(raw, "OBR|") {
		t.Error("expected OBR segment")
	}
	if !strings.Contains(raw, "85025") {
		t.Error("expected order code 85025 in OBR")
	}
}

func TestGenerateORM_RoundTrip(t *testing.T) {
	data, err := GenerateORM(testServiceRequest(), testPatient())
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}

	msg, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if msg.Type != "ORM^O01" {
		t.Errorf("expected Type 'ORM^O01', got %q", msg.Type)
	}

	orc := msg.GetSegment("ORC")
	if orc == nil {
		t.Fatal("expected ORC segment")
	}
	if orc.GetField(1) != "NW" {
		t.Errorf("expected ORC-1 'NW', got %q", orc.GetField(1))
	}
}

// =========== ORU Tests ===========

func TestGenerateORU_SingleResult(t *testing.T) {
	obs := testObservations()[:1]
	data, err := GenerateORU(testDiagnosticReport(), obs, testPatient())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw := string(data)
	if !strings.Contains(raw, "ORU^R01") {
		t.Error("expected ORU^R01 in message")
	}
	if !strings.Contains(raw, "OBR|") {
		t.Error("expected OBR segment")
	}

	// Should have exactly 1 OBX segment
	count := strings.Count(raw, "OBX|")
	if count != 1 {
		t.Errorf("expected 1 OBX segment, got %d", count)
	}
}

func TestGenerateORU_MultipleResults(t *testing.T) {
	data, err := GenerateORU(testDiagnosticReport(), testObservations(), testPatient())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw := string(data)
	count := strings.Count(raw, "OBX|")
	if count != 2 {
		t.Errorf("expected 2 OBX segments, got %d", count)
	}

	if !strings.Contains(raw, "13.5") {
		t.Error("expected hemoglobin value 13.5")
	}
	if !strings.Contains(raw, "40.1") {
		t.Error("expected hematocrit value 40.1")
	}
}

func TestGenerateORU_RoundTrip(t *testing.T) {
	data, err := GenerateORU(testDiagnosticReport(), testObservations(), testPatient())
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}

	msg, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if msg.Type != "ORU^R01" {
		t.Errorf("expected Type 'ORU^R01', got %q", msg.Type)
	}

	obxSegs := msg.GetSegments("OBX")
	if len(obxSegs) != 2 {
		t.Errorf("expected 2 OBX segments, got %d", len(obxSegs))
	}

	// Check first OBX
	if len(obxSegs) >= 1 {
		val := obxSegs[0].GetField(5)
		if val != "13.5" {
			t.Errorf("expected OBX-5 '13.5', got %q", val)
		}
	}
}

// =========== Error Cases ===========

func TestGenerateADT_NilPatient(t *testing.T) {
	_, err := GenerateADT("A01", nil, testEncounter())
	if err == nil {
		t.Error("expected error for nil patient")
	}
}

// =========== Helper Tests ===========

func TestEscapeHL7(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal text", "normal text"},
		{"pipe|char", "pipe\\F\\char"},
		{"caret^char", "caret\\S\\char"},
		{"tilde~char", "tilde\\R\\char"},
		{"backslash\\char", "backslash\\E\\char"},
		{"amp&char", "amp\\T\\char"},
		{"all|special^chars~here\\and&there", "all\\F\\special\\S\\chars\\R\\here\\E\\and\\T\\there"},
	}

	for _, tt := range tests {
		result := escapeHL7(tt.input)
		if result != tt.expected {
			t.Errorf("escapeHL7(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestBuildPID_MinimalPatient(t *testing.T) {
	patient := map[string]interface{}{
		"name": []interface{}{
			map[string]interface{}{
				"family": "Smith",
				"given":  []interface{}{"Jane"},
			},
		},
	}

	pid := buildPID(patient)
	if !strings.HasPrefix(pid, "PID|") {
		t.Error("expected PID segment prefix")
	}
	if !strings.Contains(pid, "Smith^Jane") {
		t.Error("expected patient name Smith^Jane in PID")
	}
}
