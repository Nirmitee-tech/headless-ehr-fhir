package research

import (
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

// ---------------------------------------------------------------------------
// ResearchStudy.ToFHIR
// ---------------------------------------------------------------------------

func TestResearchStudyToFHIR_RequiredFields(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	s := ResearchStudy{
		ID:             uuid.New(),
		FHIRID:         "study-1",
		Title:          "Phase III Trial of Drug X",
		ProtocolNumber: "PROTO-2024-001",
		Status:         "active-recruiting",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result := s.ToFHIR()

	// Verify resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("missing key 'resourceType'")
	} else if rt != "ResearchStudy" {
		t.Errorf("resourceType = %v, want ResearchStudy", rt)
	}

	// Verify id
	if id, ok := result["id"]; !ok {
		t.Error("missing key 'id'")
	} else if id != "study-1" {
		t.Errorf("id = %v, want study-1", id)
	}

	// Verify title
	if title, ok := result["title"]; !ok {
		t.Error("missing key 'title'")
	} else if title != "Phase III Trial of Drug X" {
		t.Errorf("title = %v, want 'Phase III Trial of Drug X'", title)
	}

	// Verify status is mapped via mapStudyStatusToFHIR
	if st, ok := result["status"]; !ok {
		t.Error("missing key 'status'")
	} else if st != "active" {
		t.Errorf("status = %v, want 'active' (mapped from 'active-recruiting')", st)
	}

	// Verify identifier with protocol number
	if _, ok := result["identifier"]; !ok {
		t.Error("missing key 'identifier'")
	}

	// Verify meta
	if _, ok := result["meta"]; !ok {
		t.Error("missing key 'meta'")
	}

	// Verify optional fields are absent
	optionalKeys := []string{"phase", "category", "description", "sponsor", "principalInvestigator", "period", "enrollment", "note"}
	for _, key := range optionalKeys {
		if _, ok := result[key]; ok {
			t.Errorf("optional key %q should not be present when field is nil", key)
		}
	}
}

func TestResearchStudyToFHIR_StatusMapping(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		internalStatus string
		fhirStatus     string
	}{
		{"in-review", "in-review"},
		{"approved", "approved"},
		{"active-recruiting", "active"},
		{"active-not-recruiting", "active"},
		{"temporarily-closed", "temporarily-closed-to-accrual"},
		{"closed", "closed-to-accrual"},
		{"completed", "completed"},
		{"withdrawn", "withdrawn"},
		{"suspended", "administratively-completed"},
		{"unknown-status", "unknown-status"}, // unmapped passes through
	}

	for _, tc := range tests {
		t.Run(tc.internalStatus, func(t *testing.T) {
			s := ResearchStudy{
				ID:             uuid.New(),
				FHIRID:         "study-status",
				Title:          "Test",
				ProtocolNumber: "P-001",
				Status:         tc.internalStatus,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			result := s.ToFHIR()
			if result["status"] != tc.fhirStatus {
				t.Errorf("status mapping %q: got %v, want %v", tc.internalStatus, result["status"], tc.fhirStatus)
			}
		})
	}
}

func TestResearchStudyToFHIR_WithOptionalFields(t *testing.T) {
	piID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)
	startDate := now.Add(-90 * 24 * time.Hour)
	endDate := now.Add(365 * 24 * time.Hour)

	s := ResearchStudy{
		ID:                      uuid.New(),
		FHIRID:                  "study-2",
		Title:                   "Phase II Oncology Trial",
		ProtocolNumber:          "PROTO-2024-002",
		Status:                  "active-recruiting",
		Phase:                   ptrStr("phase-2"),
		Category:                ptrStr("interventional"),
		Description:             ptrStr("A randomized controlled trial"),
		SponsorName:             ptrStr("PharmaCo Inc."),
		PrincipalInvestigatorID: ptrUUID(piID),
		StartDate:               ptrTime(startDate),
		EndDate:                 ptrTime(endDate),
		EnrollmentTarget:        ptrInt(200),
		Note:                    ptrStr("Multi-site study"),
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	result := s.ToFHIR()

	// phase
	if _, ok := result["phase"]; !ok {
		t.Error("missing key 'phase' when Phase is set")
	}

	// category
	if _, ok := result["category"]; !ok {
		t.Error("missing key 'category' when Category is set")
	}

	// description
	if desc, ok := result["description"]; !ok {
		t.Error("missing key 'description' when Description is set")
	} else if desc != "A randomized controlled trial" {
		t.Errorf("description = %v, want 'A randomized controlled trial'", desc)
	}

	// sponsor (Reference with Display)
	if _, ok := result["sponsor"]; !ok {
		t.Error("missing key 'sponsor' when SponsorName is set")
	}

	// principalInvestigator
	if _, ok := result["principalInvestigator"]; !ok {
		t.Error("missing key 'principalInvestigator' when PrincipalInvestigatorID is set")
	}

	// period with start and end
	if _, ok := result["period"]; !ok {
		t.Error("missing key 'period' when StartDate/EndDate are set")
	}

	// enrollment
	if _, ok := result["enrollment"]; !ok {
		t.Error("missing key 'enrollment' when EnrollmentTarget is set")
	}

	// note
	if noteVal, ok := result["note"]; !ok {
		t.Error("missing key 'note' when Note is set")
	} else {
		noteSlice, ok := noteVal.([]map[string]string)
		if !ok {
			t.Fatal("note is not []map[string]string")
		}
		if len(noteSlice) != 1 {
			t.Fatalf("note length = %d, want 1", len(noteSlice))
		}
		if noteSlice[0]["text"] != "Multi-site study" {
			t.Errorf("note[0].text = %v, want 'Multi-site study'", noteSlice[0]["text"])
		}
	}
}

func TestResearchStudyToFHIR_PeriodStartOnly(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	startDate := now.Add(-30 * 24 * time.Hour)

	s := ResearchStudy{
		ID:             uuid.New(),
		FHIRID:         "study-period-start",
		Title:          "Period Start Test",
		ProtocolNumber: "P-003",
		Status:         "approved",
		StartDate:      ptrTime(startDate),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result := s.ToFHIR()

	if _, ok := result["period"]; !ok {
		t.Error("missing key 'period' when only StartDate is set")
	}
}

func TestResearchStudyToFHIR_PeriodEndOnly(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	endDate := now.Add(180 * 24 * time.Hour)

	s := ResearchStudy{
		ID:             uuid.New(),
		FHIRID:         "study-period-end",
		Title:          "Period End Test",
		ProtocolNumber: "P-004",
		Status:         "completed",
		EndDate:        ptrTime(endDate),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result := s.ToFHIR()

	if _, ok := result["period"]; !ok {
		t.Error("missing key 'period' when only EndDate is set")
	}
}

// ---------------------------------------------------------------------------
// ResearchArm struct creation
// ---------------------------------------------------------------------------

func TestResearchArm_StructCreation(t *testing.T) {
	studyID := uuid.New()
	armID := uuid.New()

	arm := ResearchArm{
		ID:               armID,
		StudyID:          studyID,
		Name:             "Treatment Arm A",
		ArmType:          ptrStr("experimental"),
		Description:      ptrStr("Receives Drug X at 100mg daily"),
		TargetEnrollment: ptrInt(100),
		ActualEnrollment: ptrInt(42),
	}

	if arm.ID != armID {
		t.Errorf("ID = %v, want %v", arm.ID, armID)
	}
	if arm.StudyID != studyID {
		t.Errorf("StudyID = %v, want %v", arm.StudyID, studyID)
	}
	if arm.Name != "Treatment Arm A" {
		t.Errorf("Name = %v, want 'Treatment Arm A'", arm.Name)
	}
	if arm.ArmType == nil || *arm.ArmType != "experimental" {
		t.Errorf("ArmType = %v, want experimental", arm.ArmType)
	}
	if arm.Description == nil || *arm.Description != "Receives Drug X at 100mg daily" {
		t.Errorf("Description = %v, want 'Receives Drug X at 100mg daily'", arm.Description)
	}
	if arm.TargetEnrollment == nil || *arm.TargetEnrollment != 100 {
		t.Errorf("TargetEnrollment = %v, want 100", arm.TargetEnrollment)
	}
	if arm.ActualEnrollment == nil || *arm.ActualEnrollment != 42 {
		t.Errorf("ActualEnrollment = %v, want 42", arm.ActualEnrollment)
	}
}

func TestResearchArm_NilOptionalFields(t *testing.T) {
	arm := ResearchArm{
		ID:      uuid.New(),
		StudyID: uuid.New(),
		Name:    "Control Arm",
	}

	if arm.ArmType != nil {
		t.Errorf("ArmType should be nil, got %v", arm.ArmType)
	}
	if arm.Description != nil {
		t.Errorf("Description should be nil, got %v", arm.Description)
	}
	if arm.TargetEnrollment != nil {
		t.Errorf("TargetEnrollment should be nil, got %v", arm.TargetEnrollment)
	}
	if arm.ActualEnrollment != nil {
		t.Errorf("ActualEnrollment should be nil, got %v", arm.ActualEnrollment)
	}
}

// ---------------------------------------------------------------------------
// ResearchEnrollment struct creation
// ---------------------------------------------------------------------------

func TestResearchEnrollment_StructCreation(t *testing.T) {
	studyID := uuid.New()
	armID := uuid.New()
	patientID := uuid.New()
	consentID := uuid.New()
	enrolledByID := uuid.New()
	enrollmentID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)
	screenDate := now.Add(-14 * 24 * time.Hour)
	randDate := now.Add(-7 * 24 * time.Hour)

	enrollment := ResearchEnrollment{
		ID:                  enrollmentID,
		StudyID:             studyID,
		ArmID:               ptrUUID(armID),
		PatientID:           patientID,
		ConsentID:           ptrUUID(consentID),
		Status:              "enrolled",
		EnrolledDate:        ptrTime(now),
		ScreeningDate:       ptrTime(screenDate),
		RandomizationDate:   ptrTime(randDate),
		RandomizationNumber: ptrStr("RAND-0042"),
		SubjectNumber:       ptrStr("SUBJ-0042"),
		EnrolledByID:        ptrUUID(enrolledByID),
		Note:                ptrStr("Meets all inclusion criteria"),
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if enrollment.ID != enrollmentID {
		t.Errorf("ID = %v, want %v", enrollment.ID, enrollmentID)
	}
	if enrollment.StudyID != studyID {
		t.Errorf("StudyID = %v, want %v", enrollment.StudyID, studyID)
	}
	if enrollment.PatientID != patientID {
		t.Errorf("PatientID = %v, want %v", enrollment.PatientID, patientID)
	}
	if enrollment.Status != "enrolled" {
		t.Errorf("Status = %v, want enrolled", enrollment.Status)
	}
	if enrollment.ArmID == nil || *enrollment.ArmID != armID {
		t.Errorf("ArmID = %v, want %v", enrollment.ArmID, armID)
	}
	if enrollment.ConsentID == nil || *enrollment.ConsentID != consentID {
		t.Errorf("ConsentID = %v, want %v", enrollment.ConsentID, consentID)
	}
	if enrollment.RandomizationNumber == nil || *enrollment.RandomizationNumber != "RAND-0042" {
		t.Errorf("RandomizationNumber = %v, want RAND-0042", enrollment.RandomizationNumber)
	}
	if enrollment.SubjectNumber == nil || *enrollment.SubjectNumber != "SUBJ-0042" {
		t.Errorf("SubjectNumber = %v, want SUBJ-0042", enrollment.SubjectNumber)
	}
	if enrollment.EnrolledByID == nil || *enrollment.EnrolledByID != enrolledByID {
		t.Errorf("EnrolledByID = %v, want %v", enrollment.EnrolledByID, enrolledByID)
	}
	if enrollment.Note == nil || *enrollment.Note != "Meets all inclusion criteria" {
		t.Errorf("Note = %v, want 'Meets all inclusion criteria'", enrollment.Note)
	}
}

func TestResearchEnrollment_NilOptionalFields(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	enrollment := ResearchEnrollment{
		ID:        uuid.New(),
		StudyID:   uuid.New(),
		PatientID: uuid.New(),
		Status:    "screening",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if enrollment.ArmID != nil {
		t.Errorf("ArmID should be nil, got %v", enrollment.ArmID)
	}
	if enrollment.ConsentID != nil {
		t.Errorf("ConsentID should be nil, got %v", enrollment.ConsentID)
	}
	if enrollment.EnrolledDate != nil {
		t.Errorf("EnrolledDate should be nil, got %v", enrollment.EnrolledDate)
	}
	if enrollment.WithdrawalDate != nil {
		t.Errorf("WithdrawalDate should be nil, got %v", enrollment.WithdrawalDate)
	}
	if enrollment.WithdrawalReason != nil {
		t.Errorf("WithdrawalReason should be nil, got %v", enrollment.WithdrawalReason)
	}
	if enrollment.RandomizationNumber != nil {
		t.Errorf("RandomizationNumber should be nil, got %v", enrollment.RandomizationNumber)
	}
	if enrollment.Note != nil {
		t.Errorf("Note should be nil, got %v", enrollment.Note)
	}
}
