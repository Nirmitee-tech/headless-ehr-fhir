package hl7v2

import (
	"testing"
)

// =========== Sample Messages ===========

const sampleADT = "MSH|^~\\&|SendingApp|SendingFac|ReceivingApp|ReceivingFac|20240115143025||ADT^A01|MSG00001|P|2.5.1\rEVN|A01|20240115143025\rPID|1||MRN12345^^^MRNAuth||Doe^John^A||19800515|M|||123 Main St^^Springfield^IL^62701||555-555-1234\rPV1|1|I|ICU^101^A||||1234^Smith^Robert|||MED||||||||I|VN12345"

const sampleORU = "MSH|^~\\&|LabSystem|LabFac|EHR|EHRFac|20240115150000||ORU^R01|MSG00002|P|2.5.1\rPID|1||MRN12345^^^MRNAuth||Doe^John||19800515|M\rOBR|1|ORD001|LAB001|85025^CBC^LN|||20240115140000\rOBX|1|NM|718-7^Hemoglobin^LN||13.5|g/dL|12.0-17.5|N|||F\rOBX|2|NM|4544-3^Hematocrit^LN||40.1|%|36.0-53.0|N|||F"

const sampleORM = "MSH|^~\\&|OrderApp|OrderFac|LabSystem|LabFac|20240115120000||ORM^O01|MSG00003|P|2.5.1\rPID|1||MRN12345^^^MRNAuth||Doe^John||19800515|M\rORC|NW|ORD001||||||20240115120000\rOBR|1|ORD001||85025^CBC^LN|||20240115120000"

// =========== Parser Tests ===========

func TestParse_ADT_A01(t *testing.T) {
	msg, err := Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.Type != "ADT^A01" {
		t.Errorf("expected Type 'ADT^A01', got %q", msg.Type)
	}
	if msg.ControlID != "MSG00001" {
		t.Errorf("expected ControlID 'MSG00001', got %q", msg.ControlID)
	}
	if msg.Version != "2.5.1" {
		t.Errorf("expected Version '2.5.1', got %q", msg.Version)
	}
	if msg.SendingApp != "SendingApp" {
		t.Errorf("expected SendingApp 'SendingApp', got %q", msg.SendingApp)
	}
	if msg.SendingFac != "SendingFac" {
		t.Errorf("expected SendingFac 'SendingFac', got %q", msg.SendingFac)
	}
	if msg.ReceivingApp != "ReceivingApp" {
		t.Errorf("expected ReceivingApp 'ReceivingApp', got %q", msg.ReceivingApp)
	}
	if msg.ReceivingFac != "ReceivingFac" {
		t.Errorf("expected ReceivingFac 'ReceivingFac', got %q", msg.ReceivingFac)
	}
	if msg.Timestamp.Year() != 2024 || msg.Timestamp.Month() != 1 || msg.Timestamp.Day() != 15 {
		t.Errorf("unexpected timestamp: %v", msg.Timestamp)
	}
}

func TestParse_PID_Segment(t *testing.T) {
	msg, err := Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pid := msg.GetSegment("PID")
	if pid == nil {
		t.Fatal("expected PID segment")
	}

	// PID-3.1 = MRN12345
	patID := msg.PatientID()
	if patID != "MRN12345" {
		t.Errorf("expected PatientID 'MRN12345', got %q", patID)
	}

	// PID-5 = Doe^John^A
	family, given := msg.PatientName()
	if family != "Doe" {
		t.Errorf("expected family 'Doe', got %q", family)
	}
	if given != "John" {
		t.Errorf("expected given 'John', got %q", given)
	}

	// PID-7 = 19800515
	dob := msg.DateOfBirth()
	if dob != "19800515" {
		t.Errorf("expected DOB '19800515', got %q", dob)
	}

	// PID-8 = M
	gender := msg.Gender()
	if gender != "M" {
		t.Errorf("expected Gender 'M', got %q", gender)
	}
}

func TestParse_MultipleSegments(t *testing.T) {
	msg, err := Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have MSH, EVN, PID, PV1
	if len(msg.Segments) != 4 {
		t.Errorf("expected 4 segments, got %d", len(msg.Segments))
	}

	names := []string{"MSH", "EVN", "PID", "PV1"}
	for i, name := range names {
		if msg.Segments[i].Name != name {
			t.Errorf("expected segment %d to be %q, got %q", i, name, msg.Segments[i].Name)
		}
	}
}

func TestParse_EmptyInput(t *testing.T) {
	_, err := Parse([]byte{})
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParse_NoMSH(t *testing.T) {
	_, err := Parse([]byte("PID|1||MRN12345\rPV1|1|I"))
	if err == nil {
		t.Error("expected error for message without MSH")
	}
}

func TestParse_Components(t *testing.T) {
	msg, err := Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pid := msg.GetSegment("PID")
	if pid == nil {
		t.Fatal("expected PID segment")
	}

	// PID-5 = Doe^John^A — should have 3 components
	comp := pid.GetComponent(5, 1)
	if comp != "Doe" {
		t.Errorf("expected PID-5.1 'Doe', got %q", comp)
	}
	comp = pid.GetComponent(5, 2)
	if comp != "John" {
		t.Errorf("expected PID-5.2 'John', got %q", comp)
	}
	comp = pid.GetComponent(5, 3)
	if comp != "A" {
		t.Errorf("expected PID-5.3 'A', got %q", comp)
	}
}

func TestParse_Repetitions(t *testing.T) {
	raw := "MSH|^~\\&|App|Fac|||20240115143025||ADT^A01|CTRL1|P|2.5.1\rPID|1||ID1~ID2~ID3||Doe^John"
	msg, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pid := msg.GetSegment("PID")
	if pid == nil {
		t.Fatal("expected PID segment")
	}

	// PID-3 should have 3 repetitions
	if len(pid.Fields) < 3 {
		t.Fatalf("expected at least 3 fields in PID, got %d", len(pid.Fields))
	}

	field := pid.Fields[2] // PID-3 (0-indexed as field 2 since PID-1 is index 0)
	if len(field.Repeats) != 3 {
		t.Errorf("expected 3 repetitions, got %d", len(field.Repeats))
	}
	if len(field.Repeats) >= 1 && (len(field.Repeats[0]) == 0 || field.Repeats[0][0] != "ID1") {
		t.Errorf("expected first repetition 'ID1', got %v", field.Repeats[0])
	}
	if len(field.Repeats) >= 2 && (len(field.Repeats[1]) == 0 || field.Repeats[1][0] != "ID2") {
		t.Errorf("expected second repetition 'ID2', got %v", field.Repeats[1])
	}
}

func TestParse_ORU_R01(t *testing.T) {
	msg, err := Parse([]byte(sampleORU))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.Type != "ORU^R01" {
		t.Errorf("expected Type 'ORU^R01', got %q", msg.Type)
	}

	obxSegments := msg.GetSegments("OBX")
	if len(obxSegments) != 2 {
		t.Errorf("expected 2 OBX segments, got %d", len(obxSegments))
	}

	// Check first OBX value
	if len(obxSegments) >= 1 {
		val := obxSegments[0].GetField(5)
		if val != "13.5" {
			t.Errorf("expected OBX-5 '13.5', got %q", val)
		}
		unit := obxSegments[0].GetField(6)
		if unit != "g/dL" {
			t.Errorf("expected OBX-6 'g/dL', got %q", unit)
		}
	}
}

func TestParse_ORM_O01(t *testing.T) {
	msg, err := Parse([]byte(sampleORM))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

	obr := msg.GetSegment("OBR")
	if obr == nil {
		t.Fatal("expected OBR segment")
	}
}

func TestParse_WindowsLineEndings(t *testing.T) {
	raw := "MSH|^~\\&|App|Fac|||20240115143025||ADT^A01|CTRL1|P|2.5.1\r\nPID|1||MRN001||Smith^Jane\r\n"
	msg, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.Type != "ADT^A01" {
		t.Errorf("expected Type 'ADT^A01', got %q", msg.Type)
	}

	pid := msg.GetSegment("PID")
	if pid == nil {
		t.Fatal("expected PID segment with \\r\\n line endings")
	}
}

func TestParse_UnixLineEndings(t *testing.T) {
	raw := "MSH|^~\\&|App|Fac|||20240115143025||ADT^A01|CTRL1|P|2.5.1\nPID|1||MRN001||Smith^Jane\n"
	msg, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.Type != "ADT^A01" {
		t.Errorf("expected Type 'ADT^A01', got %q", msg.Type)
	}

	pid := msg.GetSegment("PID")
	if pid == nil {
		t.Fatal("expected PID segment with \\n line endings")
	}
}

func TestMessage_PatientID(t *testing.T) {
	msg, err := Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	patID := msg.PatientID()
	if patID != "MRN12345" {
		t.Errorf("expected PatientID 'MRN12345', got %q", patID)
	}
}

func TestMessage_PatientName(t *testing.T) {
	msg, err := Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	family, given := msg.PatientName()
	if family != "Doe" {
		t.Errorf("expected family 'Doe', got %q", family)
	}
	if given != "John" {
		t.Errorf("expected given 'John', got %q", given)
	}
}

func TestMessage_GetSegments(t *testing.T) {
	msg, err := Parse([]byte(sampleORU))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obxSegs := msg.GetSegments("OBX")
	if len(obxSegs) != 2 {
		t.Errorf("expected 2 OBX segments, got %d", len(obxSegs))
	}

	// Non-existent segment
	zzzSegs := msg.GetSegments("ZZZ")
	if len(zzzSegs) != 0 {
		t.Errorf("expected 0 ZZZ segments, got %d", len(zzzSegs))
	}
}

func TestSegment_GetComponent(t *testing.T) {
	msg, err := Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pid := msg.GetSegment("PID")
	if pid == nil {
		t.Fatal("expected PID segment")
	}

	// PID-3.1 = MRN12345
	comp := pid.GetComponent(3, 1)
	if comp != "MRN12345" {
		t.Errorf("expected PID-3.1 'MRN12345', got %q", comp)
	}

	// PID-3.4 = MRNAuth
	comp = pid.GetComponent(3, 4)
	if comp != "MRNAuth" {
		t.Errorf("expected PID-3.4 'MRNAuth', got %q", comp)
	}

	// Out of range component returns empty string
	comp = pid.GetComponent(3, 99)
	if comp != "" {
		t.Errorf("expected empty string for out-of-range component, got %q", comp)
	}

	// Out of range field returns empty string
	comp = pid.GetComponent(99, 1)
	if comp != "" {
		t.Errorf("expected empty string for out-of-range field, got %q", comp)
	}
}

func TestParse_NilInput(t *testing.T) {
	_, err := Parse(nil)
	if err == nil {
		t.Error("expected error for nil input")
	}
}
