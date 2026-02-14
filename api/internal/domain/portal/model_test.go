package portal

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
// Questionnaire.ToFHIR
// ---------------------------------------------------------------------------

func TestQuestionnaireToFHIR_RequiredFields(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	q := Questionnaire{
		ID:        uuid.New(),
		FHIRID:    "quest-1",
		Name:      "patient-intake",
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := q.ToFHIR()

	// Verify resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("missing key 'resourceType'")
	} else if rt != "Questionnaire" {
		t.Errorf("resourceType = %v, want Questionnaire", rt)
	}

	// Verify id
	if id, ok := result["id"]; !ok {
		t.Error("missing key 'id'")
	} else if id != "quest-1" {
		t.Errorf("id = %v, want quest-1", id)
	}

	// Verify name
	if name, ok := result["name"]; !ok {
		t.Error("missing key 'name'")
	} else if name != "patient-intake" {
		t.Errorf("name = %v, want patient-intake", name)
	}

	// Verify status
	if st, ok := result["status"]; !ok {
		t.Error("missing key 'status'")
	} else if st != "active" {
		t.Errorf("status = %v, want active", st)
	}

	// Verify meta
	if _, ok := result["meta"]; !ok {
		t.Error("missing key 'meta'")
	}

	// Verify optional fields are absent when not set
	optionalKeys := []string{"title", "version", "description", "purpose", "subjectType", "date", "publisher", "approvalDate", "lastReviewDate"}
	for _, key := range optionalKeys {
		if _, ok := result[key]; ok {
			t.Errorf("optional key %q should not be present when field is nil", key)
		}
	}
}

func TestQuestionnaireToFHIR_WithOptionalFields(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	approvalDate := now.Add(-30 * 24 * time.Hour)
	reviewDate := now.Add(-7 * 24 * time.Hour)

	q := Questionnaire{
		ID:             uuid.New(),
		FHIRID:         "quest-2",
		Name:           "phq-9",
		Title:          ptrStr("PHQ-9 Depression Screening"),
		Status:         "active",
		Version:        ptrStr("2.0"),
		Description:    ptrStr("Patient Health Questionnaire"),
		Purpose:        ptrStr("Depression screening"),
		SubjectType:    ptrStr("Patient"),
		Date:           ptrTime(now),
		Publisher:      ptrStr("Health Organization"),
		ApprovalDate:   ptrTime(approvalDate),
		LastReviewDate: ptrTime(reviewDate),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result := q.ToFHIR()

	if title, ok := result["title"]; !ok {
		t.Error("missing key 'title' when Title is set")
	} else if title != "PHQ-9 Depression Screening" {
		t.Errorf("title = %v, want 'PHQ-9 Depression Screening'", title)
	}

	if ver, ok := result["version"]; !ok {
		t.Error("missing key 'version' when Version is set")
	} else if ver != "2.0" {
		t.Errorf("version = %v, want 2.0", ver)
	}

	if desc, ok := result["description"]; !ok {
		t.Error("missing key 'description' when Description is set")
	} else if desc != "Patient Health Questionnaire" {
		t.Errorf("description = %v, want 'Patient Health Questionnaire'", desc)
	}

	if purpose, ok := result["purpose"]; !ok {
		t.Error("missing key 'purpose' when Purpose is set")
	} else if purpose != "Depression screening" {
		t.Errorf("purpose = %v, want 'Depression screening'", purpose)
	}

	// subjectType should be a slice
	if st, ok := result["subjectType"]; !ok {
		t.Error("missing key 'subjectType' when SubjectType is set")
	} else {
		stSlice, ok := st.([]string)
		if !ok {
			t.Error("subjectType is not a []string")
		} else if len(stSlice) != 1 || stSlice[0] != "Patient" {
			t.Errorf("subjectType = %v, want [Patient]", stSlice)
		}
	}

	// date formatted as "2006-01-02"
	if dt, ok := result["date"]; !ok {
		t.Error("missing key 'date' when Date is set")
	} else {
		dtStr, ok := dt.(string)
		if !ok {
			t.Error("date is not a string")
		} else if _, err := time.Parse("2006-01-02", dtStr); err != nil {
			t.Errorf("date %q is not valid date format: %v", dtStr, err)
		}
	}

	if pub, ok := result["publisher"]; !ok {
		t.Error("missing key 'publisher' when Publisher is set")
	} else if pub != "Health Organization" {
		t.Errorf("publisher = %v, want 'Health Organization'", pub)
	}

	// approvalDate formatted as "2006-01-02"
	if ad, ok := result["approvalDate"]; !ok {
		t.Error("missing key 'approvalDate' when ApprovalDate is set")
	} else {
		adStr, ok := ad.(string)
		if !ok {
			t.Error("approvalDate is not a string")
		} else if _, err := time.Parse("2006-01-02", adStr); err != nil {
			t.Errorf("approvalDate %q is not valid date format: %v", adStr, err)
		}
	}

	// lastReviewDate formatted as "2006-01-02"
	if lrd, ok := result["lastReviewDate"]; !ok {
		t.Error("missing key 'lastReviewDate' when LastReviewDate is set")
	} else {
		lrdStr, ok := lrd.(string)
		if !ok {
			t.Error("lastReviewDate is not a string")
		} else if _, err := time.Parse("2006-01-02", lrdStr); err != nil {
			t.Errorf("lastReviewDate %q is not valid date format: %v", lrdStr, err)
		}
	}
}

// ---------------------------------------------------------------------------
// QuestionnaireResponse.ToFHIR
// ---------------------------------------------------------------------------

func TestQuestionnaireResponseToFHIR_RequiredFields(t *testing.T) {
	patientID := uuid.New()
	questionnaireID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	qr := QuestionnaireResponse{
		ID:              uuid.New(),
		FHIRID:          "qr-1",
		QuestionnaireID: questionnaireID,
		PatientID:       patientID,
		Status:          "completed",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result := qr.ToFHIR()

	// Verify resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("missing key 'resourceType'")
	} else if rt != "QuestionnaireResponse" {
		t.Errorf("resourceType = %v, want QuestionnaireResponse", rt)
	}

	// Verify id
	if id, ok := result["id"]; !ok {
		t.Error("missing key 'id'")
	} else if id != "qr-1" {
		t.Errorf("id = %v, want qr-1", id)
	}

	// Verify status
	if st, ok := result["status"]; !ok {
		t.Error("missing key 'status'")
	} else if st != "completed" {
		t.Errorf("status = %v, want completed", st)
	}

	// Verify questionnaire (should be the UUID string)
	if q, ok := result["questionnaire"]; !ok {
		t.Error("missing key 'questionnaire'")
	} else if q != questionnaireID.String() {
		t.Errorf("questionnaire = %v, want %v", q, questionnaireID.String())
	}

	// Verify subject (patient reference)
	if _, ok := result["subject"]; !ok {
		t.Error("missing key 'subject'")
	}

	// Verify meta
	if _, ok := result["meta"]; !ok {
		t.Error("missing key 'meta'")
	}

	// Verify optional fields are absent
	optionalKeys := []string{"encounter", "author", "authored"}
	for _, key := range optionalKeys {
		if _, ok := result[key]; ok {
			t.Errorf("optional key %q should not be present when field is nil", key)
		}
	}
}

func TestQuestionnaireResponseToFHIR_WithOptionalFields(t *testing.T) {
	patientID := uuid.New()
	questionnaireID := uuid.New()
	encounterID := uuid.New()
	authorID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	qr := QuestionnaireResponse{
		ID:              uuid.New(),
		FHIRID:          "qr-2",
		QuestionnaireID: questionnaireID,
		PatientID:       patientID,
		EncounterID:     ptrUUID(encounterID),
		AuthorID:        ptrUUID(authorID),
		Status:          "completed",
		Authored:        ptrTime(now),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result := qr.ToFHIR()

	if _, ok := result["encounter"]; !ok {
		t.Error("missing key 'encounter' when EncounterID is set")
	}

	if _, ok := result["author"]; !ok {
		t.Error("missing key 'author' when AuthorID is set")
	}

	// authored as RFC3339
	if a, ok := result["authored"]; !ok {
		t.Error("missing key 'authored' when Authored is set")
	} else {
		aStr, ok := a.(string)
		if !ok {
			t.Error("authored is not a string")
		} else if _, err := time.Parse(time.RFC3339, aStr); err != nil {
			t.Errorf("authored %q is not valid RFC3339: %v", aStr, err)
		}
	}
}

// ---------------------------------------------------------------------------
// PortalAccount struct creation
// ---------------------------------------------------------------------------

func TestPortalAccount_StructCreation(t *testing.T) {
	patientID := uuid.New()
	accountID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)
	lastLogin := now.Add(-1 * time.Hour)

	account := PortalAccount{
		ID:                accountID,
		PatientID:         patientID,
		Username:          "jdoe",
		Email:             "jdoe@example.com",
		Phone:             ptrStr("555-0100"),
		Status:            "active",
		EmailVerified:     true,
		LastLoginAt:       ptrTime(lastLogin),
		FailedLoginCount:  0,
		MFAEnabled:        true,
		PreferredLanguage: ptrStr("en"),
		Note:              ptrStr("VIP patient"),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if account.ID != accountID {
		t.Errorf("ID = %v, want %v", account.ID, accountID)
	}
	if account.Username != "jdoe" {
		t.Errorf("Username = %v, want jdoe", account.Username)
	}
	if account.Email != "jdoe@example.com" {
		t.Errorf("Email = %v, want jdoe@example.com", account.Email)
	}
	if account.Status != "active" {
		t.Errorf("Status = %v, want active", account.Status)
	}
	if !account.EmailVerified {
		t.Error("EmailVerified = false, want true")
	}
	if !account.MFAEnabled {
		t.Error("MFAEnabled = false, want true")
	}
	if account.Phone == nil || *account.Phone != "555-0100" {
		t.Errorf("Phone = %v, want 555-0100", account.Phone)
	}
	if account.PreferredLanguage == nil || *account.PreferredLanguage != "en" {
		t.Errorf("PreferredLanguage = %v, want en", account.PreferredLanguage)
	}
}

// ---------------------------------------------------------------------------
// PatientCheckin struct creation
// ---------------------------------------------------------------------------

func TestPatientCheckin_StructCreation(t *testing.T) {
	patientID := uuid.New()
	appointmentID := uuid.New()
	checkinID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	checkin := PatientCheckin{
		ID:                checkinID,
		PatientID:         patientID,
		AppointmentID:     ptrUUID(appointmentID),
		Status:            "checked-in",
		CheckinMethod:     ptrStr("kiosk"),
		CheckinTime:       ptrTime(now),
		InsuranceVerified: ptrBool(true),
		CoPayCollected:    ptrBool(true),
		CoPayAmount:       ptrFloat(25.00),
		Note:              ptrStr("Arrived on time"),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if checkin.ID != checkinID {
		t.Errorf("ID = %v, want %v", checkin.ID, checkinID)
	}
	if checkin.PatientID != patientID {
		t.Errorf("PatientID = %v, want %v", checkin.PatientID, patientID)
	}
	if checkin.Status != "checked-in" {
		t.Errorf("Status = %v, want checked-in", checkin.Status)
	}
	if checkin.AppointmentID == nil || *checkin.AppointmentID != appointmentID {
		t.Errorf("AppointmentID = %v, want %v", checkin.AppointmentID, appointmentID)
	}
	if checkin.CheckinMethod == nil || *checkin.CheckinMethod != "kiosk" {
		t.Errorf("CheckinMethod = %v, want kiosk", checkin.CheckinMethod)
	}
	if checkin.InsuranceVerified == nil || !*checkin.InsuranceVerified {
		t.Error("InsuranceVerified should be true")
	}
	if checkin.CoPayCollected == nil || !*checkin.CoPayCollected {
		t.Error("CoPayCollected should be true")
	}
	if checkin.CoPayAmount == nil || *checkin.CoPayAmount != 25.00 {
		t.Errorf("CoPayAmount = %v, want 25.00", checkin.CoPayAmount)
	}
	if checkin.Note == nil || *checkin.Note != "Arrived on time" {
		t.Errorf("Note = %v, want 'Arrived on time'", checkin.Note)
	}
}
