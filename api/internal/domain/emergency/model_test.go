package emergency

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

func TestTriageRecord_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	original := &TriageRecord{
		ID:               uuid.New(),
		PatientID:        uuid.New(),
		EncounterID:      uuid.New(),
		TriageNurseID:    uuid.New(),
		ArrivalTime:      ptrTime(now.Add(-1 * time.Hour)),
		TriageTime:       ptrTime(now),
		ChiefComplaint:   "chest pain",
		AcuityLevel:      ptrInt(2),
		AcuitySystem:     ptrStr("ESI"),
		PainScale:        ptrInt(7),
		ArrivalMode:      ptrStr("ambulance"),
		HeartRate:        ptrInt(110),
		BloodPressureSys: ptrInt(140),
		BloodPressureDia: ptrInt(90),
		Temperature:      ptrFloat(37.5),
		RespiratoryRate:  ptrInt(20),
		OxygenSaturation: ptrInt(95),
		GlasgowComaScore: ptrInt(15),
		InjuryDescription: ptrStr("none"),
		AllergyNote:      ptrStr("NKDA"),
		MedicationNote:   ptrStr("metoprolol 50mg daily"),
		Note:             ptrStr("patient appears distressed"),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded TriageRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: got %v, want %v", decoded.ID, original.ID)
	}
	if decoded.PatientID != original.PatientID {
		t.Errorf("PatientID mismatch")
	}
	if decoded.EncounterID != original.EncounterID {
		t.Errorf("EncounterID mismatch")
	}
	if decoded.TriageNurseID != original.TriageNurseID {
		t.Errorf("TriageNurseID mismatch")
	}
	if decoded.ChiefComplaint != original.ChiefComplaint {
		t.Errorf("ChiefComplaint mismatch: got %q, want %q", decoded.ChiefComplaint, original.ChiefComplaint)
	}
	if *decoded.AcuityLevel != *original.AcuityLevel {
		t.Errorf("AcuityLevel mismatch")
	}
	if *decoded.PainScale != *original.PainScale {
		t.Errorf("PainScale mismatch")
	}
	if *decoded.HeartRate != *original.HeartRate {
		t.Errorf("HeartRate mismatch")
	}
	if *decoded.Temperature != *original.Temperature {
		t.Errorf("Temperature mismatch")
	}
}

func TestTriageRecord_OptionalFieldsNil(t *testing.T) {
	m := &TriageRecord{
		ID:             uuid.New(),
		PatientID:      uuid.New(),
		EncounterID:    uuid.New(),
		TriageNurseID:  uuid.New(),
		ChiefComplaint: "headache",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"acuity_level"`) {
		t.Error("nil AcuityLevel should be omitted")
	}
	if strings.Contains(s, `"pain_scale"`) {
		t.Error("nil PainScale should be omitted")
	}
	if strings.Contains(s, `"heart_rate"`) {
		t.Error("nil HeartRate should be omitted")
	}
	if strings.Contains(s, `"temperature"`) {
		t.Error("nil Temperature should be omitted")
	}
	if strings.Contains(s, `"arrival_mode"`) {
		t.Error("nil ArrivalMode should be omitted")
	}
	if strings.Contains(s, `"oxygen_saturation"`) {
		t.Error("nil OxygenSaturation should be omitted")
	}
	if strings.Contains(s, `"injury_description"`) {
		t.Error("nil InjuryDescription should be omitted")
	}
	if strings.Contains(s, `"note"`) {
		t.Error("nil Note should be omitted")
	}
}

func TestEDTracking_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	triageID := uuid.New()
	attendingID := uuid.New()
	nurseID := uuid.New()

	original := &EDTracking{
		ID:               uuid.New(),
		PatientID:        uuid.New(),
		EncounterID:      uuid.New(),
		TriageRecordID:   ptrUUID(triageID),
		CurrentStatus:    "in_progress",
		BedAssignment:    ptrStr("Bay 3"),
		AttendingID:      ptrUUID(attendingID),
		NurseID:          ptrUUID(nurseID),
		ArrivalTime:      ptrTime(now.Add(-2 * time.Hour)),
		DischargeTime:    ptrTime(now),
		Disposition:      ptrStr("admitted"),
		DispositionDest:  ptrStr("ICU"),
		LengthOfStayMins: ptrInt(120),
		Note:             ptrStr("critical patient"),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded EDTracking
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.CurrentStatus != original.CurrentStatus {
		t.Errorf("CurrentStatus mismatch: got %q, want %q", decoded.CurrentStatus, original.CurrentStatus)
	}
	if *decoded.BedAssignment != *original.BedAssignment {
		t.Errorf("BedAssignment mismatch")
	}
	if *decoded.TriageRecordID != *original.TriageRecordID {
		t.Errorf("TriageRecordID mismatch")
	}
	if *decoded.LengthOfStayMins != *original.LengthOfStayMins {
		t.Errorf("LengthOfStayMins mismatch")
	}
}

func TestTraumaActivation_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	encounterID := uuid.New()
	edTrackingID := uuid.New()
	activatedBy := uuid.New()
	teamLead := uuid.New()

	original := &TraumaActivation{
		ID:                uuid.New(),
		PatientID:         uuid.New(),
		EncounterID:       ptrUUID(encounterID),
		EDTrackingID:      ptrUUID(edTrackingID),
		ActivationLevel:   "level_1",
		ActivationTime:    now,
		DeactivationTime:  ptrTime(now.Add(3 * time.Hour)),
		MechanismOfInjury: ptrStr("MVC high speed"),
		ActivatedBy:       ptrUUID(activatedBy),
		TeamLeadID:        ptrUUID(teamLead),
		Outcome:           ptrStr("admitted to surgery"),
		Note:              ptrStr("multiple injuries"),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded TraumaActivation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.PatientID != original.PatientID {
		t.Errorf("PatientID mismatch")
	}
	if decoded.ActivationLevel != original.ActivationLevel {
		t.Errorf("ActivationLevel mismatch: got %q, want %q", decoded.ActivationLevel, original.ActivationLevel)
	}
	if *decoded.MechanismOfInjury != *original.MechanismOfInjury {
		t.Errorf("MechanismOfInjury mismatch")
	}
	if *decoded.Outcome != *original.Outcome {
		t.Errorf("Outcome mismatch")
	}
	if *decoded.EncounterID != *original.EncounterID {
		t.Errorf("EncounterID mismatch")
	}
}

func TestTraumaActivation_OptionalFieldsNil(t *testing.T) {
	m := &TraumaActivation{
		ID:              uuid.New(),
		PatientID:       uuid.New(),
		ActivationLevel: "level_2",
		ActivationTime:  time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"encounter_id"`) {
		t.Error("nil EncounterID should be omitted")
	}
	if strings.Contains(s, `"mechanism_of_injury"`) {
		t.Error("nil MechanismOfInjury should be omitted")
	}
	if strings.Contains(s, `"outcome"`) {
		t.Error("nil Outcome should be omitted")
	}
	if strings.Contains(s, `"deactivation_time"`) {
		t.Error("nil DeactivationTime should be omitted")
	}
}
