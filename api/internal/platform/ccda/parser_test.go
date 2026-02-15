package ccda

import (
	"testing"
)

func TestParser_Parse_BasicCCD(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")
	parser := NewParser()

	data := &PatientData{
		Patient: testPatient(),
	}

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("failed to generate CCD: %v", err)
	}

	parsed, err := parser.Parse(xmlData)
	if err != nil {
		t.Fatalf("failed to parse CCD: %v", err)
	}

	if parsed.Title != "Continuity of Care Document" {
		t.Errorf("expected title 'Continuity of Care Document', got %q", parsed.Title)
	}

	if parsed.Created.IsZero() {
		t.Error("expected Created time to be set")
	}

	if parsed.Patient.Name != "John Doe" {
		t.Errorf("expected patient name 'John Doe', got %q", parsed.Patient.Name)
	}

	if parsed.Patient.Gender != "male" {
		t.Errorf("expected gender 'male', got %q", parsed.Patient.Gender)
	}

	if parsed.Patient.DOB != "1980-01-15" {
		t.Errorf("expected DOB '1980-01-15', got %q", parsed.Patient.DOB)
	}

	if len(parsed.Patient.Identifiers) == 0 {
		t.Error("expected at least one patient identifier")
	} else {
		if parsed.Patient.Identifiers[0].Extension != "patient-123" {
			t.Errorf("expected patient ID 'patient-123', got %q", parsed.Patient.Identifiers[0].Extension)
		}
	}
}

func TestParser_Parse_WithAllergies(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")
	parser := NewParser()

	data := &PatientData{
		Patient:   testPatient(),
		Allergies: testAllergies(),
	}

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("failed to generate CCD: %v", err)
	}

	parsed, err := parser.Parse(xmlData)
	if err != nil {
		t.Fatalf("failed to parse CCD: %v", err)
	}

	var allergiesSection *ParsedSection
	for i := range parsed.Sections {
		if parsed.Sections[i].Type == "allergies" {
			allergiesSection = &parsed.Sections[i]
			break
		}
	}

	if allergiesSection == nil {
		t.Fatal("expected allergies section")
	}

	if len(allergiesSection.Entries) == 0 {
		t.Fatal("expected at least one allergy entry")
	}

	entry := allergiesSection.Entries[0]
	if entry["substance"] != "Penicillin" {
		t.Errorf("expected substance 'Penicillin', got %v", entry["substance"])
	}
	if entry["clinicalStatus"] != "active" {
		t.Errorf("expected clinicalStatus 'active', got %v", entry["clinicalStatus"])
	}
}

func TestParser_Parse_WithMedications(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")
	parser := NewParser()

	data := &PatientData{
		Patient:     testPatient(),
		Medications: testMedications(),
	}

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("failed to generate CCD: %v", err)
	}

	parsed, err := parser.Parse(xmlData)
	if err != nil {
		t.Fatalf("failed to parse CCD: %v", err)
	}

	var medsSection *ParsedSection
	for i := range parsed.Sections {
		if parsed.Sections[i].Type == "medications" {
			medsSection = &parsed.Sections[i]
			break
		}
	}

	if medsSection == nil {
		t.Fatal("expected medications section")
	}

	if len(medsSection.Entries) == 0 {
		t.Fatal("expected at least one medication entry")
	}

	entry := medsSection.Entries[0]
	if entry["status"] != "active" {
		t.Errorf("expected status 'active', got %v", entry["status"])
	}

	med, ok := entry["medication"].(string)
	if !ok {
		t.Fatalf("expected medication string, got %T", entry["medication"])
	}
	if med != "Lisinopril 10 MG" {
		t.Errorf("expected medication 'Lisinopril 10 MG', got %q", med)
	}
}

func TestParser_Parse_WithProblems(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")
	parser := NewParser()

	data := &PatientData{
		Patient:    testPatient(),
		Conditions: testConditions(),
	}

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("failed to generate CCD: %v", err)
	}

	parsed, err := parser.Parse(xmlData)
	if err != nil {
		t.Fatalf("failed to parse CCD: %v", err)
	}

	var problemsSection *ParsedSection
	for i := range parsed.Sections {
		if parsed.Sections[i].Type == "problems" {
			problemsSection = &parsed.Sections[i]
			break
		}
	}

	if problemsSection == nil {
		t.Fatal("expected problems section")
	}

	if len(problemsSection.Entries) == 0 {
		t.Fatal("expected at least one problem entry")
	}

	entry := problemsSection.Entries[0]
	if entry["problem"] != "Essential hypertension" {
		t.Errorf("expected problem 'Essential hypertension', got %v", entry["problem"])
	}
	if entry["clinicalStatus"] != "active" {
		t.Errorf("expected clinicalStatus 'active', got %v", entry["clinicalStatus"])
	}
}

func TestParser_Parse_InvalidXML(t *testing.T) {
	parser := NewParser()

	_, err := parser.Parse([]byte("this is not xml"))
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}

func TestParser_Parse_EmptyInput(t *testing.T) {
	parser := NewParser()

	_, err := parser.Parse([]byte{})
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParser_Parse_RoundTrip(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")
	parser := NewParser()

	data := fullPatientData()

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("failed to generate CCD: %v", err)
	}

	parsed, err := parser.Parse(xmlData)
	if err != nil {
		t.Fatalf("failed to parse CCD: %v", err)
	}

	// Verify patient data survived the round trip
	if parsed.Patient.Name != "John Doe" {
		t.Errorf("expected patient name 'John Doe', got %q", parsed.Patient.Name)
	}

	// Verify all 10 sections were parsed
	expectedTypes := map[string]bool{
		"allergies":      false,
		"medications":    false,
		"problems":       false,
		"procedures":     false,
		"results":        false,
		"vital_signs":    false,
		"immunizations":  false,
		"social_history": false,
		"plan_of_care":   false,
		"encounters":     false,
	}

	for _, s := range parsed.Sections {
		expectedTypes[s.Type] = true
	}

	for stype, found := range expectedTypes {
		if !found {
			t.Errorf("expected section type %q in parsed output", stype)
		}
	}

	// Verify specific entries survived round-trip
	for _, s := range parsed.Sections {
		if s.Type == "allergies" && len(s.Entries) == 0 {
			t.Error("expected allergy entries in round trip")
		}
		if s.Type == "medications" && len(s.Entries) == 0 {
			t.Error("expected medication entries in round trip")
		}
		if s.Type == "problems" && len(s.Entries) == 0 {
			t.Error("expected problem entries in round trip")
		}
		if s.Type == "procedures" && len(s.Entries) == 0 {
			t.Error("expected procedure entries in round trip")
		}
		if s.Type == "results" && len(s.Entries) == 0 {
			t.Error("expected result entries in round trip")
		}
		if s.Type == "vital_signs" && len(s.Entries) == 0 {
			t.Error("expected vital sign entries in round trip")
		}
		if s.Type == "immunizations" && len(s.Entries) == 0 {
			t.Error("expected immunization entries in round trip")
		}
		if s.Type == "encounters" && len(s.Entries) == 0 {
			t.Error("expected encounter entries in round trip")
		}
	}
}
