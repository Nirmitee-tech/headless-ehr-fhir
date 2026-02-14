package oncology

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func ptrStr(s string) *string       { return &s }
func ptrInt(i int) *int             { return &i }
func ptrFloat(f float64) *float64   { return &f }
func ptrBool(b bool) *bool          { return &b }
func ptrTime(t time.Time) *time.Time { return &t }
func ptrUUID(u uuid.UUID) *uuid.UUID { return &u }

func TestCancerDiagnosis_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	conditionID := uuid.New()
	diagProvID := uuid.New()
	mgmtProvID := uuid.New()

	original := &CancerDiagnosis{
		ID:                   uuid.New(),
		PatientID:            uuid.New(),
		ConditionID:          ptrUUID(conditionID),
		DiagnosisDate:        now,
		CancerType:           ptrStr("non-small cell lung cancer"),
		CancerSite:           ptrStr("right upper lobe"),
		HistologyCode:        ptrStr("8046/3"),
		HistologyDisplay:     ptrStr("Non-small cell carcinoma"),
		MorphologyCode:       ptrStr("8140/3"),
		MorphologyDisplay:    ptrStr("Adenocarcinoma NOS"),
		StagingSystem:        ptrStr("AJCC 8th edition"),
		StageGroup:           ptrStr("IIIA"),
		TStage:               ptrStr("T2a"),
		NStage:               ptrStr("N2"),
		MStage:               ptrStr("M0"),
		Grade:                ptrStr("G2"),
		Laterality:           ptrStr("right"),
		CurrentStatus:        "active_treatment",
		DiagnosingProviderID: ptrUUID(diagProvID),
		ManagingProviderID:   ptrUUID(mgmtProvID),
		ICD10Code:            ptrStr("C34.11"),
		ICD10Display:         ptrStr("Malignant neoplasm of upper lobe, right bronchus or lung"),
		Note:                 ptrStr("biopsy confirmed adenocarcinoma"),
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded CancerDiagnosis
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.PatientID != original.PatientID {
		t.Errorf("PatientID mismatch")
	}
	if decoded.CurrentStatus != original.CurrentStatus {
		t.Errorf("CurrentStatus mismatch: got %q, want %q", decoded.CurrentStatus, original.CurrentStatus)
	}
	if *decoded.CancerType != *original.CancerType {
		t.Errorf("CancerType mismatch")
	}
	if *decoded.StageGroup != *original.StageGroup {
		t.Errorf("StageGroup mismatch")
	}
	if *decoded.TStage != *original.TStage {
		t.Errorf("TStage mismatch")
	}
	if *decoded.ICD10Code != *original.ICD10Code {
		t.Errorf("ICD10Code mismatch")
	}
	if *decoded.DiagnosingProviderID != *original.DiagnosingProviderID {
		t.Errorf("DiagnosingProviderID mismatch")
	}
}

func TestCancerDiagnosis_OptionalFieldsNil(t *testing.T) {
	m := &CancerDiagnosis{
		ID:            uuid.New(),
		PatientID:     uuid.New(),
		DiagnosisDate: time.Now(),
		CurrentStatus: "active_treatment",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"cancer_type"`) {
		t.Error("nil CancerType should be omitted")
	}
	if strings.Contains(s, `"stage_group"`) {
		t.Error("nil StageGroup should be omitted")
	}
	if strings.Contains(s, `"t_stage"`) {
		t.Error("nil TStage should be omitted")
	}
	if strings.Contains(s, `"icd10_code"`) {
		t.Error("nil ICD10Code should be omitted")
	}
	if strings.Contains(s, `"note"`) {
		t.Error("nil Note should be omitted")
	}
}

func TestTreatmentProtocol_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	startDate := now.Add(-30 * 24 * time.Hour)
	endDate := now.Add(90 * 24 * time.Hour)
	prescribingID := uuid.New()

	original := &TreatmentProtocol{
		ID:                    uuid.New(),
		CancerDiagnosisID:     uuid.New(),
		ProtocolName:          "FOLFOX",
		ProtocolCode:          ptrStr("FOLFOX-6"),
		ProtocolType:          ptrStr("chemotherapy"),
		Intent:                ptrStr("curative"),
		NumberOfCycles:        ptrInt(12),
		CycleLengthDays:       ptrInt(14),
		StartDate:             ptrTime(startDate),
		EndDate:               ptrTime(endDate),
		Status:                "active",
		PrescribingProviderID: ptrUUID(prescribingID),
		ClinicalTrialID:       ptrStr("NCT12345678"),
		Note:                  ptrStr("standard dosing"),
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded TreatmentProtocol
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.CancerDiagnosisID != original.CancerDiagnosisID {
		t.Errorf("CancerDiagnosisID mismatch")
	}
	if decoded.ProtocolName != original.ProtocolName {
		t.Errorf("ProtocolName mismatch: got %q, want %q", decoded.ProtocolName, original.ProtocolName)
	}
	if decoded.Status != original.Status {
		t.Errorf("Status mismatch")
	}
	if *decoded.NumberOfCycles != *original.NumberOfCycles {
		t.Errorf("NumberOfCycles mismatch")
	}
	if *decoded.CycleLengthDays != *original.CycleLengthDays {
		t.Errorf("CycleLengthDays mismatch")
	}
	if *decoded.Intent != *original.Intent {
		t.Errorf("Intent mismatch")
	}
	if *decoded.ClinicalTrialID != *original.ClinicalTrialID {
		t.Errorf("ClinicalTrialID mismatch")
	}
}

func TestChemoCycle_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	providerID := uuid.New()

	original := &ChemoCycle{
		ID:                  uuid.New(),
		ProtocolID:          uuid.New(),
		CycleNumber:         3,
		PlannedStartDate:    ptrTime(now),
		ActualStartDate:     ptrTime(now.Add(1 * 24 * time.Hour)),
		ActualEndDate:       ptrTime(now.Add(3 * 24 * time.Hour)),
		Status:              "completed",
		DoseReductionPct:    ptrFloat(10.0),
		DoseReductionReason: ptrStr("neutropenia"),
		DelayDays:           ptrInt(1),
		DelayReason:         ptrStr("low ANC"),
		BSAM2:               ptrFloat(1.85),
		WeightKG:            ptrFloat(75.0),
		HeightCM:            ptrFloat(175.0),
		CreatinineClearance: ptrFloat(90.0),
		ProviderID:          ptrUUID(providerID),
		Note:                ptrStr("tolerated well"),
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded ChemoCycle
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.CycleNumber != original.CycleNumber {
		t.Errorf("CycleNumber mismatch: got %d, want %d", decoded.CycleNumber, original.CycleNumber)
	}
	if decoded.Status != original.Status {
		t.Errorf("Status mismatch")
	}
	if *decoded.DoseReductionPct != *original.DoseReductionPct {
		t.Errorf("DoseReductionPct mismatch")
	}
	if *decoded.BSAM2 != *original.BSAM2 {
		t.Errorf("BSAM2 mismatch")
	}
	if *decoded.WeightKG != *original.WeightKG {
		t.Errorf("WeightKG mismatch")
	}
	if *decoded.CreatinineClearance != *original.CreatinineClearance {
		t.Errorf("CreatinineClearance mismatch")
	}
}

func TestChemoCycle_OptionalFieldsNil(t *testing.T) {
	m := &ChemoCycle{
		ID:          uuid.New(),
		ProtocolID:  uuid.New(),
		CycleNumber: 1,
		Status:      "planned",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"dose_reduction_pct"`) {
		t.Error("nil DoseReductionPct should be omitted")
	}
	if strings.Contains(s, `"delay_days"`) {
		t.Error("nil DelayDays should be omitted")
	}
	if strings.Contains(s, `"bsa_m2"`) {
		t.Error("nil BSAM2 should be omitted")
	}
	if strings.Contains(s, `"note"`) {
		t.Error("nil Note should be omitted")
	}
}

func TestTumorMarker_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	cancerDxID := uuid.New()
	orderingProvID := uuid.New()

	original := &TumorMarker{
		ID:                  uuid.New(),
		CancerDiagnosisID:   ptrUUID(cancerDxID),
		PatientID:           uuid.New(),
		MarkerName:          "CEA",
		MarkerCode:          ptrStr("2039-6"),
		MarkerCodeSystem:    ptrStr("http://loinc.org"),
		ValueQuantity:       ptrFloat(5.2),
		ValueUnit:           ptrStr("ng/mL"),
		ValueInterpretation: ptrStr("elevated"),
		ReferenceRangeLow:   ptrFloat(0.0),
		ReferenceRangeHigh:  ptrFloat(3.0),
		ReferenceRangeText:  ptrStr("0.0-3.0 ng/mL"),
		SpecimenType:        ptrStr("serum"),
		CollectionDatetime:  ptrTime(now.Add(-1 * 24 * time.Hour)),
		ResultDatetime:      ptrTime(now),
		PerformingLab:       ptrStr("Central Lab"),
		OrderingProviderID:  ptrUUID(orderingProvID),
		Note:                ptrStr("trending upward"),
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded TumorMarker
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.PatientID != original.PatientID {
		t.Errorf("PatientID mismatch")
	}
	if decoded.MarkerName != original.MarkerName {
		t.Errorf("MarkerName mismatch: got %q, want %q", decoded.MarkerName, original.MarkerName)
	}
	if *decoded.ValueQuantity != *original.ValueQuantity {
		t.Errorf("ValueQuantity mismatch")
	}
	if *decoded.ValueUnit != *original.ValueUnit {
		t.Errorf("ValueUnit mismatch")
	}
	if *decoded.ReferenceRangeHigh != *original.ReferenceRangeHigh {
		t.Errorf("ReferenceRangeHigh mismatch")
	}
	if *decoded.PerformingLab != *original.PerformingLab {
		t.Errorf("PerformingLab mismatch")
	}
}

func TestTumorMarker_OptionalFieldsNil(t *testing.T) {
	m := &TumorMarker{
		ID:         uuid.New(),
		PatientID:  uuid.New(),
		MarkerName: "PSA",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"cancer_diagnosis_id"`) {
		t.Error("nil CancerDiagnosisID should be omitted")
	}
	if strings.Contains(s, `"value_quantity"`) {
		t.Error("nil ValueQuantity should be omitted")
	}
	if strings.Contains(s, `"reference_range_low"`) {
		t.Error("nil ReferenceRangeLow should be omitted")
	}
	if strings.Contains(s, `"performing_lab"`) {
		t.Error("nil PerformingLab should be omitted")
	}
}
