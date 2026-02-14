package behavioral

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

func TestPsychiatricAssessment_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	original := &PsychiatricAssessment{
		ID:                    uuid.New(),
		PatientID:             uuid.New(),
		EncounterID:           uuid.New(),
		AssessorID:            uuid.New(),
		AssessmentDate:        now,
		ChiefComplaint:        ptrStr("anxiety and insomnia"),
		HistoryPresentIllness: ptrStr("3-week history of worsening anxiety"),
		PsychiatricHistory:    ptrStr("prior GAD diagnosis"),
		SubstanceUseHistory:   ptrStr("social alcohol use"),
		MentalStatusExam:      ptrStr("alert and oriented"),
		Appearance:            ptrStr("well-groomed"),
		Behavior:              ptrStr("cooperative"),
		Speech:                ptrStr("normal rate and rhythm"),
		Mood:                  ptrStr("anxious"),
		Affect:                ptrStr("constricted"),
		ThoughtProcess:        ptrStr("linear"),
		ThoughtContent:        ptrStr("no SI/HI"),
		Perceptions:           ptrStr("no hallucinations"),
		Cognition:             ptrStr("intact"),
		Insight:               ptrStr("fair"),
		Judgment:              ptrStr("fair"),
		RiskAssessment:        ptrStr("low risk"),
		SuicideRiskLevel:      ptrStr("low"),
		HomicideRiskLevel:     ptrStr("none"),
		DiagnosisCode:         ptrStr("F41.1"),
		DiagnosisDisplay:      ptrStr("Generalized anxiety disorder"),
		DiagnosisSystem:       ptrStr("http://hl7.org/fhir/sid/icd-10-cm"),
		Formulation:           ptrStr("biopsychosocial formulation"),
		TreatmentPlan:         ptrStr("start SSRI, refer to therapy"),
		Disposition:           ptrStr("outpatient follow-up"),
		Note:                  ptrStr("patient amenable to treatment"),
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded PsychiatricAssessment
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.PatientID != original.PatientID {
		t.Errorf("PatientID mismatch")
	}
	if decoded.EncounterID != original.EncounterID {
		t.Errorf("EncounterID mismatch")
	}
	if decoded.AssessorID != original.AssessorID {
		t.Errorf("AssessorID mismatch")
	}
	if *decoded.ChiefComplaint != *original.ChiefComplaint {
		t.Errorf("ChiefComplaint mismatch")
	}
	if *decoded.Mood != *original.Mood {
		t.Errorf("Mood mismatch")
	}
	if *decoded.SuicideRiskLevel != *original.SuicideRiskLevel {
		t.Errorf("SuicideRiskLevel mismatch")
	}
	if *decoded.DiagnosisCode != *original.DiagnosisCode {
		t.Errorf("DiagnosisCode mismatch")
	}
	if *decoded.TreatmentPlan != *original.TreatmentPlan {
		t.Errorf("TreatmentPlan mismatch")
	}
}

func TestPsychiatricAssessment_OptionalFieldsNil(t *testing.T) {
	m := &PsychiatricAssessment{
		ID:             uuid.New(),
		PatientID:      uuid.New(),
		EncounterID:    uuid.New(),
		AssessorID:     uuid.New(),
		AssessmentDate: time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"chief_complaint"`) {
		t.Error("nil ChiefComplaint should be omitted")
	}
	if strings.Contains(s, `"mood"`) {
		t.Error("nil Mood should be omitted")
	}
	if strings.Contains(s, `"suicide_risk_level"`) {
		t.Error("nil SuicideRiskLevel should be omitted")
	}
	if strings.Contains(s, `"diagnosis_code"`) {
		t.Error("nil DiagnosisCode should be omitted")
	}
	if strings.Contains(s, `"treatment_plan"`) {
		t.Error("nil TreatmentPlan should be omitted")
	}
}

func TestSafetyPlan_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	reviewDate := now.Add(30 * 24 * time.Hour)

	original := &SafetyPlan{
		ID:                     uuid.New(),
		PatientID:              uuid.New(),
		CreatedByID:            uuid.New(),
		Status:                 "active",
		PlanDate:               now,
		WarningSigns:           ptrStr("increased isolation, poor sleep"),
		CopingStrategies:       ptrStr("deep breathing, journaling"),
		SocialDistractions:     ptrStr("call a friend, go for a walk"),
		PeopleToContact:        ptrStr("spouse, best friend"),
		ProfessionalsToContact: ptrStr("therapist Dr. Smith"),
		EmergencyContacts:      ptrStr("988 Suicide & Crisis Lifeline"),
		MeansRestriction:       ptrStr("removed firearms from home"),
		ReasonsForLiving:       ptrStr("children, future goals"),
		PatientSignature:       ptrBool(true),
		ProviderSignature:      ptrBool(true),
		ReviewDate:             ptrTime(reviewDate),
		Note:                   ptrStr("patient engaged in process"),
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded SafetyPlan
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.Status != original.Status {
		t.Errorf("Status mismatch: got %q, want %q", decoded.Status, original.Status)
	}
	if *decoded.WarningSigns != *original.WarningSigns {
		t.Errorf("WarningSigns mismatch")
	}
	if *decoded.CopingStrategies != *original.CopingStrategies {
		t.Errorf("CopingStrategies mismatch")
	}
	if *decoded.PatientSignature != *original.PatientSignature {
		t.Errorf("PatientSignature mismatch")
	}
	if *decoded.ReasonsForLiving != *original.ReasonsForLiving {
		t.Errorf("ReasonsForLiving mismatch")
	}
}

func TestSafetyPlan_OptionalFieldsNil(t *testing.T) {
	m := &SafetyPlan{
		ID:          uuid.New(),
		PatientID:   uuid.New(),
		CreatedByID: uuid.New(),
		Status:      "active",
		PlanDate:    time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"warning_signs"`) {
		t.Error("nil WarningSigns should be omitted")
	}
	if strings.Contains(s, `"patient_signature"`) {
		t.Error("nil PatientSignature should be omitted")
	}
	if strings.Contains(s, `"review_date"`) {
		t.Error("nil ReviewDate should be omitted")
	}
}

func TestLegalHold_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	encounterID := uuid.New()
	certPhysID := uuid.New()

	original := &LegalHold{
		ID:                    uuid.New(),
		PatientID:             uuid.New(),
		EncounterID:           ptrUUID(encounterID),
		InitiatedByID:         uuid.New(),
		Status:                "active",
		HoldType:              "involuntary_72hr",
		AuthorityStatute:      ptrStr("WIC 5150"),
		StartDatetime:         now,
		EndDatetime:           ptrTime(now.Add(72 * time.Hour)),
		DurationHours:         ptrInt(72),
		Reason:                "danger to self",
		CriteriaMet:           ptrStr("suicidal ideation with plan"),
		CertifyingPhysicianID: ptrUUID(certPhysID),
		CertificationDatetime: ptrTime(now.Add(1 * time.Hour)),
		CourtHearingDate:      ptrTime(now.Add(48 * time.Hour)),
		CourtOrderNumber:      ptrStr("CO-2024-12345"),
		LegalCounselNotified:  ptrBool(true),
		PatientRightsGiven:    ptrBool(true),
		Note:                  ptrStr("patient informed of rights"),
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded LegalHold
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.Status != original.Status {
		t.Errorf("Status mismatch: got %q, want %q", decoded.Status, original.Status)
	}
	if decoded.HoldType != original.HoldType {
		t.Errorf("HoldType mismatch: got %q, want %q", decoded.HoldType, original.HoldType)
	}
	if decoded.Reason != original.Reason {
		t.Errorf("Reason mismatch")
	}
	if *decoded.AuthorityStatute != *original.AuthorityStatute {
		t.Errorf("AuthorityStatute mismatch")
	}
	if *decoded.DurationHours != *original.DurationHours {
		t.Errorf("DurationHours mismatch")
	}
	if *decoded.LegalCounselNotified != *original.LegalCounselNotified {
		t.Errorf("LegalCounselNotified mismatch")
	}
	if *decoded.CertifyingPhysicianID != *original.CertifyingPhysicianID {
		t.Errorf("CertifyingPhysicianID mismatch")
	}
}

func TestLegalHold_OptionalFieldsNil(t *testing.T) {
	m := &LegalHold{
		ID:            uuid.New(),
		PatientID:     uuid.New(),
		InitiatedByID: uuid.New(),
		Status:        "active",
		HoldType:      "involuntary_72hr",
		StartDatetime: time.Now(),
		Reason:        "danger to self",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"encounter_id"`) {
		t.Error("nil EncounterID should be omitted")
	}
	if strings.Contains(s, `"end_datetime"`) {
		t.Error("nil EndDatetime should be omitted")
	}
	if strings.Contains(s, `"court_order_number"`) {
		t.Error("nil CourtOrderNumber should be omitted")
	}
	if strings.Contains(s, `"release_reason"`) {
		t.Error("nil ReleaseReason should be omitted")
	}
}
