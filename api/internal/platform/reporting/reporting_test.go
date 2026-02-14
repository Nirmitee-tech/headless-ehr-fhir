package reporting

import (
	"testing"
)

func TestPredefinedMeasures(t *testing.T) {
	if len(PredefinedMeasures) != 4 {
		t.Fatalf("expected 4 predefined measures, got %d", len(PredefinedMeasures))
	}

	expectedIDs := []string{
		"patient-count",
		"encounter-volume-by-type",
		"active-medication-orders",
		"diagnostic-report-summary",
	}

	for i, expectedID := range expectedIDs {
		if PredefinedMeasures[i].ID != expectedID {
			t.Errorf("expected measure[%d].ID = %s, got %s", i, expectedID, PredefinedMeasures[i].ID)
		}
	}
}

func TestPredefinedMeasures_HaveSQL(t *testing.T) {
	for _, m := range PredefinedMeasures {
		if m.SQL == "" {
			t.Errorf("measure %s has empty SQL", m.ID)
		}
		if m.Name == "" {
			t.Errorf("measure %s has empty name", m.ID)
		}
		if m.Description == "" {
			t.Errorf("measure %s has empty description", m.ID)
		}
	}
}

func TestFindMeasure_Exists(t *testing.T) {
	m := FindMeasure("patient-count")
	if m == nil {
		t.Fatal("expected to find patient-count measure")
	}
	if m.Name != "Patient Count" {
		t.Errorf("expected 'Patient Count', got %s", m.Name)
	}
}

func TestFindMeasure_NotFound(t *testing.T) {
	m := FindMeasure("nonexistent")
	if m != nil {
		t.Error("expected nil for nonexistent measure")
	}
}

func TestFindMeasure_AllPredefined(t *testing.T) {
	for _, def := range PredefinedMeasures {
		found := FindMeasure(def.ID)
		if found == nil {
			t.Errorf("expected to find measure %s", def.ID)
		}
		if found != nil && found.ID != def.ID {
			t.Errorf("ID mismatch: expected %s, got %s", def.ID, found.ID)
		}
	}
}

func TestMeasureDefinition_Structure(t *testing.T) {
	m := MeasureDefinition{
		ID:          "test-measure",
		Name:        "Test Measure",
		Description: "A test measure",
		SQL:         "SELECT 1",
		Parameters:  []string{"param1", "param2"},
	}

	if m.ID != "test-measure" {
		t.Errorf("unexpected ID: %s", m.ID)
	}
	if len(m.Parameters) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(m.Parameters))
	}
}

func TestMeasureReport_Structure(t *testing.T) {
	report := MeasureReport{
		MeasureID:   "patient-count",
		MeasureName: "Patient Count",
		Results: []map[string]interface{}{
			{"total": 100, "active_count": 85},
		},
		Parameters: map[string]string{"status": "active"},
	}

	if report.MeasureID != "patient-count" {
		t.Errorf("unexpected MeasureID: %s", report.MeasureID)
	}
	if len(report.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(report.Results))
	}
	if report.Results[0]["total"] != 100 {
		t.Errorf("unexpected total: %v", report.Results[0]["total"])
	}
	if report.Parameters["status"] != "active" {
		t.Errorf("unexpected parameter: %v", report.Parameters["status"])
	}
}

func TestNewHandler(t *testing.T) {
	h := NewHandler(nil)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestPatientCountMeasure_SQL(t *testing.T) {
	m := FindMeasure("patient-count")
	if m == nil {
		t.Fatal("expected patient-count measure")
	}
	if len(m.Parameters) != 0 {
		t.Errorf("expected 0 parameters, got %d", len(m.Parameters))
	}
}

func TestEncounterVolumeMeasure_SQL(t *testing.T) {
	m := FindMeasure("encounter-volume-by-type")
	if m == nil {
		t.Fatal("expected encounter-volume-by-type measure")
	}
	if m.Name != "Encounter Volume by Type" {
		t.Errorf("unexpected name: %s", m.Name)
	}
}

func TestActiveMedicationOrdersMeasure(t *testing.T) {
	m := FindMeasure("active-medication-orders")
	if m == nil {
		t.Fatal("expected active-medication-orders measure")
	}
	if m.Name != "Active Medication Orders" {
		t.Errorf("unexpected name: %s", m.Name)
	}
}

func TestDiagnosticReportSummaryMeasure(t *testing.T) {
	m := FindMeasure("diagnostic-report-summary")
	if m == nil {
		t.Fatal("expected diagnostic-report-summary measure")
	}
	if m.Name != "Diagnostic Report Summary" {
		t.Errorf("unexpected name: %s", m.Name)
	}
}
