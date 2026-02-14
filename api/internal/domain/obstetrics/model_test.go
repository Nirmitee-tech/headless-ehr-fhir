package obstetrics

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

func TestPregnancy_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	edd := now.Add(180 * 24 * time.Hour)
	lmp := now.Add(-90 * 24 * time.Hour)
	providerID := uuid.New()
	orgID := uuid.New()

	original := &Pregnancy{
		ID:                     uuid.New(),
		PatientID:              uuid.New(),
		Status:                 "active",
		OnsetDate:              ptrTime(lmp.Add(14 * 24 * time.Hour)),
		EstimatedDueDate:       ptrTime(edd),
		LastMenstrualPeriod:    ptrTime(lmp),
		ConceptionMethod:      ptrStr("natural"),
		Gravida:                ptrInt(2),
		Para:                   ptrInt(1),
		MultipleGestation:      ptrBool(false),
		NumberOfFetuses:        ptrInt(1),
		RiskLevel:              ptrStr("low"),
		RiskFactors:            ptrStr("none identified"),
		BloodType:              ptrStr("O"),
		RhFactor:               ptrStr("positive"),
		PrePregnancyWeight:     ptrFloat(65.0),
		PrePregnancyBMI:        ptrFloat(23.5),
		PrimaryProviderID:      ptrUUID(providerID),
		ManagingOrganizationID: ptrUUID(orgID),
		Note:                   ptrStr("uncomplicated pregnancy"),
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Pregnancy
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.PatientID != original.PatientID {
		t.Errorf("PatientID mismatch")
	}
	if decoded.Status != original.Status {
		t.Errorf("Status mismatch: got %q, want %q", decoded.Status, original.Status)
	}
	if *decoded.Gravida != *original.Gravida {
		t.Errorf("Gravida mismatch")
	}
	if *decoded.Para != *original.Para {
		t.Errorf("Para mismatch")
	}
	if *decoded.BloodType != *original.BloodType {
		t.Errorf("BloodType mismatch")
	}
	if *decoded.PrePregnancyWeight != *original.PrePregnancyWeight {
		t.Errorf("PrePregnancyWeight mismatch")
	}
	if *decoded.PrimaryProviderID != *original.PrimaryProviderID {
		t.Errorf("PrimaryProviderID mismatch")
	}
}

func TestPregnancy_OptionalFieldsNil(t *testing.T) {
	m := &Pregnancy{
		ID:        uuid.New(),
		PatientID: uuid.New(),
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"estimated_due_date"`) {
		t.Error("nil EstimatedDueDate should be omitted")
	}
	if strings.Contains(s, `"gravida"`) {
		t.Error("nil Gravida should be omitted")
	}
	if strings.Contains(s, `"risk_level"`) {
		t.Error("nil RiskLevel should be omitted")
	}
	if strings.Contains(s, `"blood_type"`) {
		t.Error("nil BloodType should be omitted")
	}
	if strings.Contains(s, `"outcome_summary"`) {
		t.Error("nil OutcomeSummary should be omitted")
	}
}

func TestLaborRecord_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	encounterID := uuid.New()
	providerID := uuid.New()

	original := &LaborRecord{
		ID:                      uuid.New(),
		PregnancyID:             uuid.New(),
		EncounterID:             ptrUUID(encounterID),
		AdmissionDatetime:       ptrTime(now.Add(-6 * time.Hour)),
		LaborOnsetDatetime:      ptrTime(now.Add(-8 * time.Hour)),
		LaborOnsetType:          ptrStr("spontaneous"),
		MembraneRuptureDatetime: ptrTime(now.Add(-4 * time.Hour)),
		MembraneRuptureType:     ptrStr("spontaneous"),
		AmnioticFluidColor:      ptrStr("clear"),
		AmnioticFluidVolume:     ptrStr("adequate"),
		AnesthesiaType:          ptrStr("epidural"),
		AnesthesiaStart:         ptrTime(now.Add(-3 * time.Hour)),
		Status:                  "active",
		AttendingProviderID:     ptrUUID(providerID),
		Note:                    ptrStr("progressing normally"),
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded LaborRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.PregnancyID != original.PregnancyID {
		t.Errorf("PregnancyID mismatch")
	}
	if decoded.Status != original.Status {
		t.Errorf("Status mismatch: got %q, want %q", decoded.Status, original.Status)
	}
	if *decoded.LaborOnsetType != *original.LaborOnsetType {
		t.Errorf("LaborOnsetType mismatch")
	}
	if *decoded.AmnioticFluidColor != *original.AmnioticFluidColor {
		t.Errorf("AmnioticFluidColor mismatch")
	}
	if *decoded.AnesthesiaType != *original.AnesthesiaType {
		t.Errorf("AnesthesiaType mismatch")
	}
	if *decoded.AttendingProviderID != *original.AttendingProviderID {
		t.Errorf("AttendingProviderID mismatch")
	}
}

func TestDeliveryRecord_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	laborID := uuid.New()
	assistantID := uuid.New()
	locationID := uuid.New()

	original := &DeliveryRecord{
		ID:                   uuid.New(),
		PregnancyID:          uuid.New(),
		LaborRecordID:        ptrUUID(laborID),
		PatientID:            uuid.New(),
		DeliveryDatetime:     now,
		DeliveryMethod:       "vaginal",
		DeliveryType:         ptrStr("spontaneous"),
		DeliveringProviderID: uuid.New(),
		AssistantProviderID:  ptrUUID(assistantID),
		DeliveryLocationID:   ptrUUID(locationID),
		BirthOrder:           ptrInt(1),
		PlacentaDelivery:     ptrStr("spontaneous"),
		PlacentaIntact:       ptrBool(true),
		CordVessels:          ptrInt(3),
		CordBloodCollected:   ptrBool(false),
		Episiotomy:           ptrBool(false),
		LacerationDegree:     ptrStr("second"),
		RepairMethod:         ptrStr("continuous suture"),
		BloodLossML:          ptrInt(350),
		Note:                 ptrStr("uncomplicated delivery"),
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded DeliveryRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.DeliveryMethod != original.DeliveryMethod {
		t.Errorf("DeliveryMethod mismatch: got %q, want %q", decoded.DeliveryMethod, original.DeliveryMethod)
	}
	if decoded.DeliveringProviderID != original.DeliveringProviderID {
		t.Errorf("DeliveringProviderID mismatch")
	}
	if *decoded.BirthOrder != *original.BirthOrder {
		t.Errorf("BirthOrder mismatch")
	}
	if *decoded.BloodLossML != *original.BloodLossML {
		t.Errorf("BloodLossML mismatch")
	}
	if *decoded.PlacentaIntact != *original.PlacentaIntact {
		t.Errorf("PlacentaIntact mismatch")
	}
	if *decoded.CordVessels != *original.CordVessels {
		t.Errorf("CordVessels mismatch")
	}
}

func TestNewbornRecord_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	patientID := uuid.New()

	original := &NewbornRecord{
		ID:                  uuid.New(),
		DeliveryID:          uuid.New(),
		PatientID:           ptrUUID(patientID),
		BirthDatetime:       now,
		Sex:                 ptrStr("female"),
		BirthWeightGrams:    ptrInt(3250),
		BirthLengthCM:       ptrFloat(50.0),
		HeadCircumferenceCM: ptrFloat(34.5),
		Apgar1Min:           ptrInt(8),
		Apgar5Min:           ptrInt(9),
		Apgar10Min:          ptrInt(9),
		GestationalAgeWeeks: ptrInt(39),
		GestationalAgeDays:  ptrInt(3),
		BirthStatus:         ptrStr("live_birth"),
		NICUAdmission:       ptrBool(false),
		VitaminKGiven:       ptrBool(true),
		EyeProphylaxisGiven: ptrBool(true),
		HepatitisBGiven:     ptrBool(true),
		NewbornScreening:    ptrStr("pending"),
		FeedingMethod:       ptrStr("breastfeeding"),
		Note:                ptrStr("healthy newborn"),
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded NewbornRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.DeliveryID != original.DeliveryID {
		t.Errorf("DeliveryID mismatch")
	}
	if *decoded.Sex != *original.Sex {
		t.Errorf("Sex mismatch")
	}
	if *decoded.BirthWeightGrams != *original.BirthWeightGrams {
		t.Errorf("BirthWeightGrams mismatch")
	}
	if *decoded.Apgar1Min != *original.Apgar1Min {
		t.Errorf("Apgar1Min mismatch")
	}
	if *decoded.Apgar5Min != *original.Apgar5Min {
		t.Errorf("Apgar5Min mismatch")
	}
	if *decoded.GestationalAgeWeeks != *original.GestationalAgeWeeks {
		t.Errorf("GestationalAgeWeeks mismatch")
	}
	if *decoded.VitaminKGiven != *original.VitaminKGiven {
		t.Errorf("VitaminKGiven mismatch")
	}
	if *decoded.FeedingMethod != *original.FeedingMethod {
		t.Errorf("FeedingMethod mismatch")
	}
}

func TestNewbornRecord_OptionalFieldsNil(t *testing.T) {
	m := &NewbornRecord{
		ID:            uuid.New(),
		DeliveryID:    uuid.New(),
		BirthDatetime: time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"sex"`) {
		t.Error("nil Sex should be omitted")
	}
	if strings.Contains(s, `"birth_weight_grams"`) {
		t.Error("nil BirthWeightGrams should be omitted")
	}
	if strings.Contains(s, `"apgar_1min"`) {
		t.Error("nil Apgar1Min should be omitted")
	}
	if strings.Contains(s, `"nicu_admission"`) {
		t.Error("nil NICUAdmission should be omitted")
	}
	if strings.Contains(s, `"feeding_method"`) {
		t.Error("nil FeedingMethod should be omitted")
	}
}
