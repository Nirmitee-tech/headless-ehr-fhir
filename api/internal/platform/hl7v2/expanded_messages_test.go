package hl7v2

import (
	"strings"
	"testing"
)

// =========================================================================
// Test Helpers — FHIR-like resource builders for new message types
// =========================================================================

func testMergePatient() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           "surviving-patient-1",
		"name": []interface{}{
			map[string]interface{}{
				"family": "Johnson",
				"given":  []interface{}{"Robert"},
			},
		},
		"birthDate": "1975-03-20",
		"gender":    "male",
		"identifier": []interface{}{
			map[string]interface{}{
				"value": "MRN99999",
			},
		},
	}
}

func testMergeParams() map[string]interface{} {
	return map[string]interface{}{
		"priorPatientID": "MRN11111",
		"priorAccountID": "ACCT11111",
	}
}

func testMergeAccountParams() map[string]interface{} {
	return map[string]interface{}{
		"priorAccountID": "ACCT22222",
	}
}

func testPharmacyOrder() map[string]interface{} {
	return map[string]interface{}{
		"orderControl": "RE",
		"orderID":      "RX-001",
	}
}

func testPharmacyGive() map[string]interface{} {
	return map[string]interface{}{
		"giveCode":       "313820",
		"giveCodeText":   "Acetaminophen 500mg Oral Tablet",
		"giveCodeSystem": "RXNORM",
		"giveAmount":     "2",
		"giveUnits":      "TAB",
		"giveDosageForm": "TAB",
	}
}

func testDiagnosis() map[string]interface{} {
	return map[string]interface{}{
		"code":        "J06.9",
		"description": "Acute upper respiratory infection",
		"type":        "A",
		"codeSystem":  "I10",
	}
}

func testVisit() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Encounter",
		"id":           "enc-bar-001",
		"class": map[string]interface{}{
			"code": "AMB",
		},
		"status": "finished",
		"location": []interface{}{
			map[string]interface{}{
				"location": map[string]interface{}{
					"display": "Clinic Room 3",
				},
			},
		},
		"participant": []interface{}{
			map[string]interface{}{
				"individual": map[string]interface{}{
					"display": "Dr. Alice Brown",
				},
			},
		},
	}
}

// =========================================================================
// Sample raw messages for parser tests
// =========================================================================

const sampleADTA40 = "MSH|^~\\&|EHR|EHRFac|Destination|DestFac|20240215100000||ADT^A40|MERGE001|P|2.5.1\r" +
	"EVN|A40|20240215100000\r" +
	"PID|1||MRN99999||Johnson^Robert||19750320|M\r" +
	"MRG|MRN11111||ACCT11111"

const sampleADTA41 = "MSH|^~\\&|EHR|EHRFac|Destination|DestFac|20240215100000||ADT^A41|MERGE002|P|2.5.1\r" +
	"EVN|A41|20240215100000\r" +
	"PID|1||MRN99999||Johnson^Robert||19750320|M\r" +
	"MRG|||ACCT22222"

const sampleRGVO15 = "MSH|^~\\&|EHR|EHRFac|Pharmacy|PharmFac|20240215100000||RGV^O15|PHARM001|P|2.5.1\r" +
	"PID|1||MRN99999||Johnson^Robert||19750320|M\r" +
	"ORC|RE|RX-001\r" +
	"RXG|1||313820^Acetaminophen 500mg Oral Tablet^RXNORM|2||TAB|TAB"

const sampleBARP01 = "MSH|^~\\&|EHR|EHRFac|Billing|BillFac|20240215100000||BAR^P01|BILL001|P|2.5.1\r" +
	"EVN|P01|20240215100000\r" +
	"PID|1||MRN99999||Johnson^Robert||19750320|M\r" +
	"PV1|1|O|Clinic Room 3||||Dr. Alice Brown\r" +
	"DG1|1|I10|J06.9^Acute upper respiratory infection^I10||20240215|A"

const sampleBARP05 = "MSH|^~\\&|EHR|EHRFac|Billing|BillFac|20240215100000||BAR^P05|BILL002|P|2.5.1\r" +
	"EVN|P05|20240215100000\r" +
	"PID|1||MRN99999||Johnson^Robert||19750320|M\r" +
	"PV1|1|O|Clinic Room 3||||Dr. Alice Brown\r" +
	"DG1|1|I10|J06.9^Acute upper respiratory infection^I10||20240215|A"

// =========================================================================
// ADT A40 — Merge Patient Tests
// =========================================================================

func TestGenerateADT_A40(t *testing.T) {
	data, err := GenerateADT_A40(testMergePatient(), testMergeParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw := string(data)

	if !strings.HasPrefix(raw, "MSH|") {
		t.Error("expected message to start with MSH|")
	}
	if !strings.Contains(raw, "ADT^A40") {
		t.Error("expected ADT^A40 in message")
	}
	if !strings.Contains(raw, "EVN|A40") {
		t.Error("expected EVN|A40 segment")
	}
	if !strings.Contains(raw, "PID|") {
		t.Error("expected PID segment")
	}
	if !strings.Contains(raw, "Johnson^Robert") {
		t.Error("expected surviving patient name Johnson^Robert")
	}
	if !strings.Contains(raw, "MRG|") {
		t.Error("expected MRG segment")
	}
	if !strings.Contains(raw, "MRN11111") {
		t.Error("expected prior patient ID MRN11111 in MRG")
	}
	if !strings.Contains(raw, "ACCT11111") {
		t.Error("expected prior account ID ACCT11111 in MRG")
	}
}

func TestGenerateADT_A40_NilPatient(t *testing.T) {
	_, err := GenerateADT_A40(nil, testMergeParams())
	if err == nil {
		t.Error("expected error for nil patient")
	}
}

func TestGenerateADT_A40_NilMergeParams(t *testing.T) {
	_, err := GenerateADT_A40(testMergePatient(), nil)
	if err == nil {
		t.Error("expected error for nil merge params")
	}
}

func TestParseADT_A40(t *testing.T) {
	msg, err := Parse([]byte(sampleADTA40))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.Type != "ADT^A40" {
		t.Errorf("expected Type 'ADT^A40', got %q", msg.Type)
	}
	if msg.ControlID != "MERGE001" {
		t.Errorf("expected ControlID 'MERGE001', got %q", msg.ControlID)
	}

	// Check PID (surviving patient)
	patID := msg.PatientID()
	if patID != "MRN99999" {
		t.Errorf("expected PatientID 'MRN99999', got %q", patID)
	}

	family, given := msg.PatientName()
	if family != "Johnson" || given != "Robert" {
		t.Errorf("expected Johnson^Robert, got %s^%s", family, given)
	}

	// Check MRG segment
	mrg := msg.GetSegment("MRG")
	if mrg == nil {
		t.Fatal("expected MRG segment")
	}
	if mrg.GetField(1) != "MRN11111" {
		t.Errorf("expected MRG-1 'MRN11111', got %q", mrg.GetField(1))
	}
	if mrg.GetField(3) != "ACCT11111" {
		t.Errorf("expected MRG-3 'ACCT11111', got %q", mrg.GetField(3))
	}
}

func TestGenerateADT_A40_RoundTrip(t *testing.T) {
	data, err := GenerateADT_A40(testMergePatient(), testMergeParams())
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}

	msg, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if msg.Type != "ADT^A40" {
		t.Errorf("expected Type 'ADT^A40', got %q", msg.Type)
	}
	if msg.Version != "2.5.1" {
		t.Errorf("expected Version '2.5.1', got %q", msg.Version)
	}

	patID := msg.PatientID()
	if patID != "MRN99999" {
		t.Errorf("expected PatientID 'MRN99999', got %q", patID)
	}

	mrg := msg.GetSegment("MRG")
	if mrg == nil {
		t.Fatal("expected MRG segment")
	}
	if mrg.GetField(1) != "MRN11111" {
		t.Errorf("expected MRG-1 'MRN11111', got %q", mrg.GetField(1))
	}
	if mrg.GetField(3) != "ACCT11111" {
		t.Errorf("expected MRG-3 'ACCT11111', got %q", mrg.GetField(3))
	}
}

// =========================================================================
// ADT A41 — Merge Account Tests
// =========================================================================

func TestGenerateADT_A41(t *testing.T) {
	data, err := GenerateADT_A41(testMergePatient(), testMergeAccountParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw := string(data)

	if !strings.Contains(raw, "ADT^A41") {
		t.Error("expected ADT^A41 in message")
	}
	if !strings.Contains(raw, "EVN|A41") {
		t.Error("expected EVN|A41 segment")
	}
	if !strings.Contains(raw, "MRG|") {
		t.Error("expected MRG segment")
	}
	if !strings.Contains(raw, "ACCT22222") {
		t.Error("expected prior account ID ACCT22222 in MRG")
	}
}

func TestGenerateADT_A41_NilPatient(t *testing.T) {
	_, err := GenerateADT_A41(nil, testMergeAccountParams())
	if err == nil {
		t.Error("expected error for nil patient")
	}
}

func TestGenerateADT_A41_NilMergeParams(t *testing.T) {
	_, err := GenerateADT_A41(testMergePatient(), nil)
	if err == nil {
		t.Error("expected error for nil merge params")
	}
}

func TestParseADT_A41(t *testing.T) {
	msg, err := Parse([]byte(sampleADTA41))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.Type != "ADT^A41" {
		t.Errorf("expected Type 'ADT^A41', got %q", msg.Type)
	}

	mrg := msg.GetSegment("MRG")
	if mrg == nil {
		t.Fatal("expected MRG segment")
	}
	if mrg.GetField(3) != "ACCT22222" {
		t.Errorf("expected MRG-3 'ACCT22222', got %q", mrg.GetField(3))
	}
}

func TestGenerateADT_A41_RoundTrip(t *testing.T) {
	data, err := GenerateADT_A41(testMergePatient(), testMergeAccountParams())
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}

	msg, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if msg.Type != "ADT^A41" {
		t.Errorf("expected Type 'ADT^A41', got %q", msg.Type)
	}

	mrg := msg.GetSegment("MRG")
	if mrg == nil {
		t.Fatal("expected MRG segment")
	}
	if mrg.GetField(3) != "ACCT22222" {
		t.Errorf("expected MRG-3 'ACCT22222', got %q", mrg.GetField(3))
	}
}

// =========================================================================
// RGV O15 — Pharmacy/Treatment Give Tests
// =========================================================================

func TestGenerateRGV_O15(t *testing.T) {
	data, err := GenerateRGV_O15(testMergePatient(), testPharmacyOrder(), testPharmacyGive())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw := string(data)

	if !strings.HasPrefix(raw, "MSH|") {
		t.Error("expected message to start with MSH|")
	}
	if !strings.Contains(raw, "RGV^O15") {
		t.Error("expected RGV^O15 in message")
	}
	if !strings.Contains(raw, "PID|") {
		t.Error("expected PID segment")
	}
	if !strings.Contains(raw, "ORC|") {
		t.Error("expected ORC segment")
	}
	if !strings.Contains(raw, "RE") {
		t.Error("expected order control RE in ORC")
	}
	if !strings.Contains(raw, "RXG|") {
		t.Error("expected RXG segment")
	}
	if !strings.Contains(raw, "313820") {
		t.Error("expected give code 313820 in RXG")
	}
	if !strings.Contains(raw, "Acetaminophen 500mg Oral Tablet") {
		t.Error("expected give code text in RXG")
	}
}

func TestGenerateRGV_O15_NilPatient(t *testing.T) {
	_, err := GenerateRGV_O15(nil, testPharmacyOrder(), testPharmacyGive())
	if err == nil {
		t.Error("expected error for nil patient")
	}
}

func TestGenerateRGV_O15_NilOrder(t *testing.T) {
	_, err := GenerateRGV_O15(testMergePatient(), nil, testPharmacyGive())
	if err == nil {
		t.Error("expected error for nil order")
	}
}

func TestGenerateRGV_O15_NilGive(t *testing.T) {
	_, err := GenerateRGV_O15(testMergePatient(), testPharmacyOrder(), nil)
	if err == nil {
		t.Error("expected error for nil give")
	}
}

func TestParseRGV_O15(t *testing.T) {
	msg, err := Parse([]byte(sampleRGVO15))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.Type != "RGV^O15" {
		t.Errorf("expected Type 'RGV^O15', got %q", msg.Type)
	}
	if msg.ControlID != "PHARM001" {
		t.Errorf("expected ControlID 'PHARM001', got %q", msg.ControlID)
	}

	// Check ORC
	orc := msg.GetSegment("ORC")
	if orc == nil {
		t.Fatal("expected ORC segment")
	}
	if orc.GetField(1) != "RE" {
		t.Errorf("expected ORC-1 'RE', got %q", orc.GetField(1))
	}
	if orc.GetField(2) != "RX-001" {
		t.Errorf("expected ORC-2 'RX-001', got %q", orc.GetField(2))
	}

	// Check RXG
	rxg := msg.GetSegment("RXG")
	if rxg == nil {
		t.Fatal("expected RXG segment")
	}
	// RXG-1 = give sub-ID counter
	if rxg.GetField(1) != "1" {
		t.Errorf("expected RXG-1 '1', got %q", rxg.GetField(1))
	}
	// RXG-3 = give code (coded element)
	if rxg.GetComponent(3, 1) != "313820" {
		t.Errorf("expected RXG-3.1 '313820', got %q", rxg.GetComponent(3, 1))
	}
	if rxg.GetComponent(3, 2) != "Acetaminophen 500mg Oral Tablet" {
		t.Errorf("expected RXG-3.2 'Acetaminophen 500mg Oral Tablet', got %q", rxg.GetComponent(3, 2))
	}
	// RXG-4 = give amount
	if rxg.GetField(4) != "2" {
		t.Errorf("expected RXG-4 '2', got %q", rxg.GetField(4))
	}
	// RXG-6 = give units
	if rxg.GetField(6) != "TAB" {
		t.Errorf("expected RXG-6 'TAB', got %q", rxg.GetField(6))
	}
	// RXG-7 = give dosage form
	if rxg.GetField(7) != "TAB" {
		t.Errorf("expected RXG-7 'TAB', got %q", rxg.GetField(7))
	}
}

func TestGenerateRGV_O15_RoundTrip(t *testing.T) {
	data, err := GenerateRGV_O15(testMergePatient(), testPharmacyOrder(), testPharmacyGive())
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}

	msg, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if msg.Type != "RGV^O15" {
		t.Errorf("expected Type 'RGV^O15', got %q", msg.Type)
	}
	if msg.Version != "2.5.1" {
		t.Errorf("expected Version '2.5.1', got %q", msg.Version)
	}

	// Verify patient survived round-trip
	patID := msg.PatientID()
	if patID != "MRN99999" {
		t.Errorf("expected PatientID 'MRN99999', got %q", patID)
	}

	// Verify RXG survived round-trip
	rxg := msg.GetSegment("RXG")
	if rxg == nil {
		t.Fatal("expected RXG segment")
	}
	if rxg.GetComponent(3, 1) != "313820" {
		t.Errorf("expected RXG-3.1 '313820', got %q", rxg.GetComponent(3, 1))
	}
	if rxg.GetField(4) != "2" {
		t.Errorf("expected RXG-4 '2', got %q", rxg.GetField(4))
	}
	if rxg.GetField(6) != "TAB" {
		t.Errorf("expected RXG-6 'TAB', got %q", rxg.GetField(6))
	}
}

// =========================================================================
// BAR P01 — Add Patient Account Tests
// =========================================================================

func TestGenerateBAR_P01(t *testing.T) {
	data, err := GenerateBAR_P01(testMergePatient(), testVisit(), testDiagnosis())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw := string(data)

	if !strings.HasPrefix(raw, "MSH|") {
		t.Error("expected message to start with MSH|")
	}
	if !strings.Contains(raw, "BAR^P01") {
		t.Error("expected BAR^P01 in message")
	}
	if !strings.Contains(raw, "EVN|P01") {
		t.Error("expected EVN|P01 segment")
	}
	if !strings.Contains(raw, "PID|") {
		t.Error("expected PID segment")
	}
	if !strings.Contains(raw, "PV1|") {
		t.Error("expected PV1 segment")
	}
	if !strings.Contains(raw, "DG1|") {
		t.Error("expected DG1 segment")
	}
	if !strings.Contains(raw, "J06.9") {
		t.Error("expected diagnosis code J06.9 in DG1")
	}
	if !strings.Contains(raw, "Acute upper respiratory infection") {
		t.Error("expected diagnosis description in DG1")
	}
}

func TestGenerateBAR_P01_NilPatient(t *testing.T) {
	_, err := GenerateBAR_P01(nil, testVisit(), testDiagnosis())
	if err == nil {
		t.Error("expected error for nil patient")
	}
}

func TestGenerateBAR_P01_NilDiagnosis(t *testing.T) {
	_, err := GenerateBAR_P01(testMergePatient(), testVisit(), nil)
	if err == nil {
		t.Error("expected error for nil diagnosis")
	}
}

func TestParseBAR_P01(t *testing.T) {
	msg, err := Parse([]byte(sampleBARP01))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.Type != "BAR^P01" {
		t.Errorf("expected Type 'BAR^P01', got %q", msg.Type)
	}
	if msg.ControlID != "BILL001" {
		t.Errorf("expected ControlID 'BILL001', got %q", msg.ControlID)
	}

	// Check PV1
	pv1 := msg.GetSegment("PV1")
	if pv1 == nil {
		t.Fatal("expected PV1 segment")
	}
	if pv1.GetField(2) != "O" {
		t.Errorf("expected PV1-2 'O' (outpatient), got %q", pv1.GetField(2))
	}

	// Check DG1
	dg1 := msg.GetSegment("DG1")
	if dg1 == nil {
		t.Fatal("expected DG1 segment")
	}
	if dg1.GetField(1) != "1" {
		t.Errorf("expected DG1-1 '1', got %q", dg1.GetField(1))
	}
	if dg1.GetField(2) != "I10" {
		t.Errorf("expected DG1-2 'I10', got %q", dg1.GetField(2))
	}
	// DG1-3 is coded element: code^description^system
	if dg1.GetComponent(3, 1) != "J06.9" {
		t.Errorf("expected DG1-3.1 'J06.9', got %q", dg1.GetComponent(3, 1))
	}
	if dg1.GetComponent(3, 2) != "Acute upper respiratory infection" {
		t.Errorf("expected DG1-3.2 'Acute upper respiratory infection', got %q", dg1.GetComponent(3, 2))
	}
	// DG1-6 = diagnosis type
	if dg1.GetField(6) != "A" {
		t.Errorf("expected DG1-6 'A', got %q", dg1.GetField(6))
	}
}

func TestGenerateBAR_P01_RoundTrip(t *testing.T) {
	data, err := GenerateBAR_P01(testMergePatient(), testVisit(), testDiagnosis())
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}

	msg, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if msg.Type != "BAR^P01" {
		t.Errorf("expected Type 'BAR^P01', got %q", msg.Type)
	}
	if msg.Version != "2.5.1" {
		t.Errorf("expected Version '2.5.1', got %q", msg.Version)
	}

	dg1 := msg.GetSegment("DG1")
	if dg1 == nil {
		t.Fatal("expected DG1 segment")
	}
	if dg1.GetComponent(3, 1) != "J06.9" {
		t.Errorf("expected DG1-3.1 'J06.9', got %q", dg1.GetComponent(3, 1))
	}
	if dg1.GetComponent(3, 2) != "Acute upper respiratory infection" {
		t.Errorf("expected DG1-3.2 description, got %q", dg1.GetComponent(3, 2))
	}
	if dg1.GetField(6) != "A" {
		t.Errorf("expected DG1-6 'A', got %q", dg1.GetField(6))
	}
}

// =========================================================================
// BAR P05 — Update Account Tests
// =========================================================================

func TestGenerateBAR_P05(t *testing.T) {
	data, err := GenerateBAR_P05(testMergePatient(), testVisit(), testDiagnosis())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw := string(data)

	if !strings.Contains(raw, "BAR^P05") {
		t.Error("expected BAR^P05 in message")
	}
	if !strings.Contains(raw, "EVN|P05") {
		t.Error("expected EVN|P05 segment")
	}
	if !strings.Contains(raw, "PID|") {
		t.Error("expected PID segment")
	}
	if !strings.Contains(raw, "PV1|") {
		t.Error("expected PV1 segment")
	}
	if !strings.Contains(raw, "DG1|") {
		t.Error("expected DG1 segment")
	}
}

func TestGenerateBAR_P05_NilPatient(t *testing.T) {
	_, err := GenerateBAR_P05(nil, testVisit(), testDiagnosis())
	if err == nil {
		t.Error("expected error for nil patient")
	}
}

func TestGenerateBAR_P05_NilDiagnosis(t *testing.T) {
	_, err := GenerateBAR_P05(testMergePatient(), testVisit(), nil)
	if err == nil {
		t.Error("expected error for nil diagnosis")
	}
}

func TestParseBAR_P05(t *testing.T) {
	msg, err := Parse([]byte(sampleBARP05))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.Type != "BAR^P05" {
		t.Errorf("expected Type 'BAR^P05', got %q", msg.Type)
	}
	if msg.ControlID != "BILL002" {
		t.Errorf("expected ControlID 'BILL002', got %q", msg.ControlID)
	}

	// DG1 should be present
	dg1 := msg.GetSegment("DG1")
	if dg1 == nil {
		t.Fatal("expected DG1 segment")
	}
	if dg1.GetComponent(3, 1) != "J06.9" {
		t.Errorf("expected DG1-3.1 'J06.9', got %q", dg1.GetComponent(3, 1))
	}
}

func TestGenerateBAR_P05_RoundTrip(t *testing.T) {
	data, err := GenerateBAR_P05(testMergePatient(), testVisit(), testDiagnosis())
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}

	msg, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if msg.Type != "BAR^P05" {
		t.Errorf("expected Type 'BAR^P05', got %q", msg.Type)
	}

	// Verify all segments survived
	segNames := []string{"MSH", "EVN", "PID", "PV1", "DG1"}
	for _, name := range segNames {
		if msg.GetSegment(name) == nil {
			t.Errorf("expected segment %s to be present", name)
		}
	}
}

// =========================================================================
// Segment Builder Isolation Tests
// =========================================================================

func TestBuildMRG_MergePatient(t *testing.T) {
	params := testMergeParams()
	seg := buildMRG(params)

	if !strings.HasPrefix(seg, "MRG|") {
		t.Error("expected MRG segment prefix")
	}
	if !strings.Contains(seg, "MRN11111") {
		t.Error("expected prior patient ID MRN11111")
	}
	if !strings.Contains(seg, "ACCT11111") {
		t.Error("expected prior account ID ACCT11111")
	}
}

func TestBuildMRG_AccountOnly(t *testing.T) {
	params := testMergeAccountParams()
	seg := buildMRG(params)

	if !strings.HasPrefix(seg, "MRG|") {
		t.Error("expected MRG segment prefix")
	}
	if !strings.Contains(seg, "ACCT22222") {
		t.Error("expected prior account ID ACCT22222")
	}
}

func TestBuildRXG(t *testing.T) {
	give := testPharmacyGive()
	seg := buildRXG(give)

	if !strings.HasPrefix(seg, "RXG|") {
		t.Error("expected RXG segment prefix")
	}
	if !strings.Contains(seg, "313820") {
		t.Error("expected give code 313820")
	}
	if !strings.Contains(seg, "Acetaminophen 500mg Oral Tablet") {
		t.Error("expected give code text")
	}
	if !strings.Contains(seg, "RXNORM") {
		t.Error("expected give code system RXNORM")
	}
}

func TestBuildDG1(t *testing.T) {
	diag := testDiagnosis()
	seg := buildDG1(1, diag)

	if !strings.HasPrefix(seg, "DG1|") {
		t.Error("expected DG1 segment prefix")
	}
	if !strings.Contains(seg, "J06.9") {
		t.Error("expected diagnosis code J06.9")
	}
	if !strings.Contains(seg, "Acute upper respiratory infection") {
		t.Error("expected diagnosis description")
	}
	if !strings.Contains(seg, "I10") {
		t.Error("expected code system I10")
	}
}

func TestBuildDG1_SetID(t *testing.T) {
	diag := testDiagnosis()
	seg := buildDG1(3, diag)

	if !strings.HasPrefix(seg, "DG1|3|") {
		t.Errorf("expected DG1|3| prefix, got segment starting with: %s", seg[:min(10, len(seg))])
	}
}

func TestBuildORC_Pharmacy(t *testing.T) {
	order := testPharmacyOrder()
	seg := buildPharmacyORC(order)

	if !strings.HasPrefix(seg, "ORC|") {
		t.Error("expected ORC segment prefix")
	}
	if !strings.Contains(seg, "RE") {
		t.Error("expected order control RE")
	}
	if !strings.Contains(seg, "RX-001") {
		t.Error("expected order ID RX-001")
	}
}

// =========================================================================
// Cross-message segment counting
// =========================================================================

func TestParseADT_A40_SegmentCount(t *testing.T) {
	msg, err := Parse([]byte(sampleADTA40))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have MSH, EVN, PID, MRG = 4 segments
	if len(msg.Segments) != 4 {
		t.Errorf("expected 4 segments, got %d", len(msg.Segments))
	}

	expected := []string{"MSH", "EVN", "PID", "MRG"}
	for i, name := range expected {
		if i < len(msg.Segments) && msg.Segments[i].Name != name {
			t.Errorf("expected segment %d to be %q, got %q", i, name, msg.Segments[i].Name)
		}
	}
}

func TestParseRGV_O15_SegmentCount(t *testing.T) {
	msg, err := Parse([]byte(sampleRGVO15))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have MSH, PID, ORC, RXG = 4 segments
	if len(msg.Segments) != 4 {
		t.Errorf("expected 4 segments, got %d", len(msg.Segments))
	}

	expected := []string{"MSH", "PID", "ORC", "RXG"}
	for i, name := range expected {
		if i < len(msg.Segments) && msg.Segments[i].Name != name {
			t.Errorf("expected segment %d to be %q, got %q", i, name, msg.Segments[i].Name)
		}
	}
}

func TestParseBAR_P01_SegmentCount(t *testing.T) {
	msg, err := Parse([]byte(sampleBARP01))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have MSH, EVN, PID, PV1, DG1 = 5 segments
	if len(msg.Segments) != 5 {
		t.Errorf("expected 5 segments, got %d", len(msg.Segments))
	}

	expected := []string{"MSH", "EVN", "PID", "PV1", "DG1"}
	for i, name := range expected {
		if i < len(msg.Segments) && msg.Segments[i].Name != name {
			t.Errorf("expected segment %d to be %q, got %q", i, name, msg.Segments[i].Name)
		}
	}
}

// =========================================================================
// RXG Component Extraction Tests
// =========================================================================

func TestRXG_ComponentExtraction(t *testing.T) {
	msg, err := Parse([]byte(sampleRGVO15))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rxg := msg.GetSegment("RXG")
	if rxg == nil {
		t.Fatal("expected RXG segment")
	}

	// RXG-3 should be 313820^Acetaminophen 500mg Oral Tablet^RXNORM
	code := rxg.GetComponent(3, 1)
	if code != "313820" {
		t.Errorf("expected RXG-3.1 '313820', got %q", code)
	}

	text := rxg.GetComponent(3, 2)
	if text != "Acetaminophen 500mg Oral Tablet" {
		t.Errorf("expected RXG-3.2 description, got %q", text)
	}

	system := rxg.GetComponent(3, 3)
	if system != "RXNORM" {
		t.Errorf("expected RXG-3.3 'RXNORM', got %q", system)
	}
}

// =========================================================================
// DG1 Component Extraction Tests
// =========================================================================

func TestDG1_ComponentExtraction(t *testing.T) {
	msg, err := Parse([]byte(sampleBARP01))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dg1 := msg.GetSegment("DG1")
	if dg1 == nil {
		t.Fatal("expected DG1 segment")
	}

	// DG1-3 = J06.9^Acute upper respiratory infection^I10
	code := dg1.GetComponent(3, 1)
	if code != "J06.9" {
		t.Errorf("expected DG1-3.1 'J06.9', got %q", code)
	}

	desc := dg1.GetComponent(3, 2)
	if desc != "Acute upper respiratory infection" {
		t.Errorf("expected DG1-3.2 description, got %q", desc)
	}

	system := dg1.GetComponent(3, 3)
	if system != "I10" {
		t.Errorf("expected DG1-3.3 'I10', got %q", system)
	}
}

// =========================================================================
// MRG Field Extraction Tests
// =========================================================================

func TestMRG_FieldExtraction(t *testing.T) {
	msg, err := Parse([]byte(sampleADTA40))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mrg := msg.GetSegment("MRG")
	if mrg == nil {
		t.Fatal("expected MRG segment")
	}

	// MRG-1 = prior patient ID
	if mrg.GetField(1) != "MRN11111" {
		t.Errorf("expected MRG-1 'MRN11111', got %q", mrg.GetField(1))
	}

	// MRG-3 = prior account number
	if mrg.GetField(3) != "ACCT11111" {
		t.Errorf("expected MRG-3 'ACCT11111', got %q", mrg.GetField(3))
	}

	// MRG-2 should be empty (not used in A40)
	if mrg.GetField(2) != "" {
		t.Errorf("expected MRG-2 empty, got %q", mrg.GetField(2))
	}
}
