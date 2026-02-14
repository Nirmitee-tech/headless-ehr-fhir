package documents

import (
	"encoding/json"
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
// Consent.ToFHIR
// ---------------------------------------------------------------------------

func TestConsentToFHIR_RequiredFields(t *testing.T) {
	patientID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	c := Consent{
		ID:        uuid.New(),
		FHIRID:    "consent-1",
		Status:    "active",
		PatientID: patientID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := c.ToFHIR()

	// Verify resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("missing key 'resourceType'")
	} else if rt != "Consent" {
		t.Errorf("resourceType = %v, want Consent", rt)
	}

	// Verify id
	if id, ok := result["id"]; !ok {
		t.Error("missing key 'id'")
	} else if id != "consent-1" {
		t.Errorf("id = %v, want consent-1", id)
	}

	// Verify status
	if st, ok := result["status"]; !ok {
		t.Error("missing key 'status'")
	} else if st != "active" {
		t.Errorf("status = %v, want active", st)
	}

	// Verify patient reference
	if _, ok := result["patient"]; !ok {
		t.Error("missing key 'patient'")
	}

	// Verify meta
	if _, ok := result["meta"]; !ok {
		t.Error("missing key 'meta'")
	}

	// Verify optional fields are absent when not set
	optionalKeys := []string{"scope", "category", "performer", "organization", "policy", "provision", "dateTime"}
	for _, key := range optionalKeys {
		if _, ok := result[key]; ok {
			t.Errorf("optional key %q should not be present when field is nil", key)
		}
	}
}

func TestConsentToFHIR_WithOptionalFields(t *testing.T) {
	patientID := uuid.New()
	performerID := uuid.New()
	orgID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)
	provStart := now.Add(-24 * time.Hour)
	provEnd := now.Add(24 * time.Hour)

	c := Consent{
		ID:              uuid.New(),
		FHIRID:          "consent-2",
		Status:          "active",
		Scope:           ptrStr("patient-privacy"),
		CategoryCode:    ptrStr("59284-0"),
		CategoryDisplay: ptrStr("Consent Document"),
		PatientID:       patientID,
		PerformerID:     ptrUUID(performerID),
		OrganizationID:  ptrUUID(orgID),
		PolicyURI:       ptrStr("http://example.org/policy"),
		ProvisionType:   ptrStr("permit"),
		ProvisionStart:  ptrTime(provStart),
		ProvisionEnd:    ptrTime(provEnd),
		ProvisionAction: ptrStr("access"),
		DateTime:        ptrTime(now),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result := c.ToFHIR()

	// scope must be present
	if _, ok := result["scope"]; !ok {
		t.Error("missing key 'scope' when Scope is set")
	}

	// category must be present
	if _, ok := result["category"]; !ok {
		t.Error("missing key 'category' when CategoryCode is set")
	}

	// performer must be present
	if _, ok := result["performer"]; !ok {
		t.Error("missing key 'performer' when PerformerID is set")
	}

	// organization must be present
	if _, ok := result["organization"]; !ok {
		t.Error("missing key 'organization' when OrganizationID is set")
	}

	// policy must be present
	if _, ok := result["policy"]; !ok {
		t.Error("missing key 'policy' when PolicyURI is set")
	}

	// provision must be present with nested period and action
	provisionVal, ok := result["provision"]
	if !ok {
		t.Fatal("missing key 'provision' when ProvisionType is set")
	}
	provision, ok := provisionVal.(map[string]interface{})
	if !ok {
		t.Fatal("provision is not a map[string]interface{}")
	}
	if _, ok := provision["type"]; !ok {
		t.Error("provision missing 'type' key")
	}
	if _, ok := provision["period"]; !ok {
		t.Error("provision missing 'period' key when ProvisionStart/End are set")
	}
	if _, ok := provision["action"]; !ok {
		t.Error("provision missing 'action' key when ProvisionAction is set")
	}

	// dateTime must be present
	if dt, ok := result["dateTime"]; !ok {
		t.Error("missing key 'dateTime' when DateTime is set")
	} else {
		dtStr, ok := dt.(string)
		if !ok {
			t.Error("dateTime is not a string")
		} else if _, err := time.Parse(time.RFC3339, dtStr); err != nil {
			t.Errorf("dateTime %q is not valid RFC3339: %v", dtStr, err)
		}
	}
}

// ---------------------------------------------------------------------------
// DocumentReference.ToFHIR
// ---------------------------------------------------------------------------

func TestDocumentReferenceToFHIR_RequiredFields(t *testing.T) {
	patientID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	d := DocumentReference{
		ID:        uuid.New(),
		FHIRID:    "docref-1",
		Status:    "current",
		PatientID: patientID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := d.ToFHIR()

	// Verify resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("missing key 'resourceType'")
	} else if rt != "DocumentReference" {
		t.Errorf("resourceType = %v, want DocumentReference", rt)
	}

	// Verify id
	if id, ok := result["id"]; !ok {
		t.Error("missing key 'id'")
	} else if id != "docref-1" {
		t.Errorf("id = %v, want docref-1", id)
	}

	// Verify status
	if st, ok := result["status"]; !ok {
		t.Error("missing key 'status'")
	} else if st != "current" {
		t.Errorf("status = %v, want current", st)
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
	optionalKeys := []string{"docStatus", "type", "category", "author", "custodian", "context", "date", "description", "securityLabel", "content"}
	for _, key := range optionalKeys {
		if _, ok := result[key]; ok {
			t.Errorf("optional key %q should not be present when field is nil", key)
		}
	}
}

func TestDocumentReferenceToFHIR_WithOptionalFields(t *testing.T) {
	patientID := uuid.New()
	authorID := uuid.New()
	custodianID := uuid.New()
	encounterID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	d := DocumentReference{
		ID:              uuid.New(),
		FHIRID:          "docref-2",
		Status:          "current",
		DocStatus:       ptrStr("final"),
		TypeCode:        ptrStr("34133-9"),
		TypeDisplay:     ptrStr("Summary of episode note"),
		CategoryCode:    ptrStr("clinical-note"),
		CategoryDisplay: ptrStr("Clinical Note"),
		PatientID:       patientID,
		AuthorID:        ptrUUID(authorID),
		CustodianID:     ptrUUID(custodianID),
		EncounterID:     ptrUUID(encounterID),
		Date:            ptrTime(now),
		Description:     ptrStr("Patient summary document"),
		SecurityLabel:   ptrStr("R"),
		ContentType:     ptrStr("application/pdf"),
		ContentURL:      ptrStr("https://example.org/doc.pdf"),
		ContentSize:     ptrInt(1024),
		ContentHash:     ptrStr("abc123hash"),
		ContentTitle:    ptrStr("Summary.pdf"),
		FormatCode:      ptrStr("urn:ihe:pcc:xds-ms:2007"),
		FormatDisplay:   ptrStr("Medical Summary"),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result := d.ToFHIR()

	if result["docStatus"] != "final" {
		t.Errorf("docStatus = %v, want final", result["docStatus"])
	}

	if _, ok := result["type"]; !ok {
		t.Error("missing key 'type' when TypeCode is set")
	}

	if _, ok := result["category"]; !ok {
		t.Error("missing key 'category' when CategoryCode is set")
	}

	if _, ok := result["author"]; !ok {
		t.Error("missing key 'author' when AuthorID is set")
	}

	if _, ok := result["custodian"]; !ok {
		t.Error("missing key 'custodian' when CustodianID is set")
	}

	// context with encounter
	contextVal, ok := result["context"]
	if !ok {
		t.Fatal("missing key 'context' when EncounterID is set")
	}
	contextMap, ok := contextVal.(map[string]interface{})
	if !ok {
		t.Fatal("context is not a map[string]interface{}")
	}
	if _, ok := contextMap["encounter"]; !ok {
		t.Error("context missing 'encounter' key")
	}

	// date as RFC3339 string
	if dt, ok := result["date"]; !ok {
		t.Error("missing key 'date' when Date is set")
	} else {
		dtStr, ok := dt.(string)
		if !ok {
			t.Error("date is not a string")
		} else if _, err := time.Parse(time.RFC3339, dtStr); err != nil {
			t.Errorf("date %q is not valid RFC3339: %v", dtStr, err)
		}
	}

	if result["description"] != "Patient summary document" {
		t.Errorf("description = %v, want 'Patient summary document'", result["description"])
	}

	if _, ok := result["securityLabel"]; !ok {
		t.Error("missing key 'securityLabel' when SecurityLabel is set")
	}

	// content with attachment and format
	contentVal, ok := result["content"]
	if !ok {
		t.Fatal("missing key 'content' when content fields are set")
	}
	contentSlice, ok := contentVal.([]interface{})
	if !ok {
		t.Fatal("content is not a []interface{}")
	}
	if len(contentSlice) != 1 {
		t.Fatalf("content length = %d, want 1", len(contentSlice))
	}
	contentEntry, ok := contentSlice[0].(map[string]interface{})
	if !ok {
		t.Fatal("content[0] is not a map[string]interface{}")
	}

	// Verify attachment fields
	attachmentVal, ok := contentEntry["attachment"]
	if !ok {
		t.Fatal("content[0] missing 'attachment' key")
	}
	attachment, ok := attachmentVal.(map[string]interface{})
	if !ok {
		t.Fatal("attachment is not a map[string]interface{}")
	}
	if attachment["contentType"] != "application/pdf" {
		t.Errorf("attachment contentType = %v, want application/pdf", attachment["contentType"])
	}
	if attachment["url"] != "https://example.org/doc.pdf" {
		t.Errorf("attachment url = %v, want https://example.org/doc.pdf", attachment["url"])
	}
	if attachment["size"] != 1024 {
		t.Errorf("attachment size = %v, want 1024", attachment["size"])
	}
	if attachment["hash"] != "abc123hash" {
		t.Errorf("attachment hash = %v, want abc123hash", attachment["hash"])
	}
	if attachment["title"] != "Summary.pdf" {
		t.Errorf("attachment title = %v, want Summary.pdf", attachment["title"])
	}

	// Verify format
	if _, ok := contentEntry["format"]; !ok {
		t.Error("content[0] missing 'format' key when FormatCode is set")
	}
}

// ---------------------------------------------------------------------------
// Composition.ToFHIR
// ---------------------------------------------------------------------------

func TestCompositionToFHIR_RequiredFields(t *testing.T) {
	patientID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	comp := Composition{
		ID:        uuid.New(),
		FHIRID:    "comp-1",
		Status:    "final",
		PatientID: patientID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := comp.ToFHIR()

	// Verify resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("missing key 'resourceType'")
	} else if rt != "Composition" {
		t.Errorf("resourceType = %v, want Composition", rt)
	}

	// Verify id
	if id, ok := result["id"]; !ok {
		t.Error("missing key 'id'")
	} else if id != "comp-1" {
		t.Errorf("id = %v, want comp-1", id)
	}

	// Verify status
	if st, ok := result["status"]; !ok {
		t.Error("missing key 'status'")
	} else if st != "final" {
		t.Errorf("status = %v, want final", st)
	}

	// Verify subject
	if _, ok := result["subject"]; !ok {
		t.Error("missing key 'subject'")
	}

	// Verify meta
	if _, ok := result["meta"]; !ok {
		t.Error("missing key 'meta'")
	}

	// Verify optional fields are absent
	optionalKeys := []string{"type", "category", "encounter", "date", "author", "title", "confidentiality", "custodian"}
	for _, key := range optionalKeys {
		if _, ok := result[key]; ok {
			t.Errorf("optional key %q should not be present when field is nil", key)
		}
	}
}

func TestCompositionToFHIR_WithOptionalFields(t *testing.T) {
	patientID := uuid.New()
	encounterID := uuid.New()
	authorID := uuid.New()
	custodianID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	comp := Composition{
		ID:              uuid.New(),
		FHIRID:          "comp-2",
		Status:          "final",
		TypeCode:        ptrStr("11488-4"),
		TypeDisplay:     ptrStr("Consult note"),
		CategoryCode:    ptrStr("LP173421-1"),
		CategoryDisplay: ptrStr("Report"),
		PatientID:       patientID,
		EncounterID:     ptrUUID(encounterID),
		Date:            ptrTime(now),
		AuthorID:        ptrUUID(authorID),
		Title:           ptrStr("Consultation Note"),
		Confidentiality: ptrStr("N"),
		CustodianID:     ptrUUID(custodianID),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result := comp.ToFHIR()

	if _, ok := result["type"]; !ok {
		t.Error("missing key 'type' when TypeCode is set")
	}

	if _, ok := result["category"]; !ok {
		t.Error("missing key 'category' when CategoryCode is set")
	}

	if _, ok := result["encounter"]; !ok {
		t.Error("missing key 'encounter' when EncounterID is set")
	}

	// date as RFC3339
	if dt, ok := result["date"]; !ok {
		t.Error("missing key 'date' when Date is set")
	} else {
		dtStr, ok := dt.(string)
		if !ok {
			t.Error("date is not a string")
		} else if _, err := time.Parse(time.RFC3339, dtStr); err != nil {
			t.Errorf("date %q is not valid RFC3339: %v", dtStr, err)
		}
	}

	if _, ok := result["author"]; !ok {
		t.Error("missing key 'author' when AuthorID is set")
	}

	if title, ok := result["title"]; !ok {
		t.Error("missing key 'title' when Title is set")
	} else if title != "Consultation Note" {
		t.Errorf("title = %v, want 'Consultation Note'", title)
	}

	if conf, ok := result["confidentiality"]; !ok {
		t.Error("missing key 'confidentiality' when Confidentiality is set")
	} else if conf != "N" {
		t.Errorf("confidentiality = %v, want N", conf)
	}

	if _, ok := result["custodian"]; !ok {
		t.Error("missing key 'custodian' when CustodianID is set")
	}
}

// ---------------------------------------------------------------------------
// ClinicalNote JSON marshal/unmarshal
// ---------------------------------------------------------------------------

func TestClinicalNote_JSONRoundTrip(t *testing.T) {
	noteID := uuid.New()
	patientID := uuid.New()
	authorID := uuid.New()
	encounterID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	note := ClinicalNote{
		ID:          noteID,
		PatientID:   patientID,
		EncounterID: ptrUUID(encounterID),
		AuthorID:    authorID,
		NoteType:    "progress",
		Status:      "final",
		Title:       ptrStr("Progress Note"),
		Subjective:  ptrStr("Patient reports improvement"),
		Objective:   ptrStr("Vitals stable"),
		Assessment:  ptrStr("Improving"),
		Plan:        ptrStr("Continue current treatment"),
		NoteText:    ptrStr("Full note text here"),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(note)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded ClinicalNote
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.ID != noteID {
		t.Errorf("ID = %v, want %v", decoded.ID, noteID)
	}
	if decoded.PatientID != patientID {
		t.Errorf("PatientID = %v, want %v", decoded.PatientID, patientID)
	}
	if decoded.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", decoded.AuthorID, authorID)
	}
	if decoded.NoteType != "progress" {
		t.Errorf("NoteType = %v, want progress", decoded.NoteType)
	}
	if decoded.Status != "final" {
		t.Errorf("Status = %v, want final", decoded.Status)
	}
	if decoded.Title == nil || *decoded.Title != "Progress Note" {
		t.Errorf("Title = %v, want 'Progress Note'", decoded.Title)
	}
	if decoded.Subjective == nil || *decoded.Subjective != "Patient reports improvement" {
		t.Errorf("Subjective = %v, want 'Patient reports improvement'", decoded.Subjective)
	}
	if decoded.EncounterID == nil || *decoded.EncounterID != encounterID {
		t.Errorf("EncounterID = %v, want %v", decoded.EncounterID, encounterID)
	}
}

func TestClinicalNote_JSONOmitsNilOptionalFields(t *testing.T) {
	note := ClinicalNote{
		ID:        uuid.New(),
		PatientID: uuid.New(),
		AuthorID:  uuid.New(),
		NoteType:  "progress",
		Status:    "draft",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	data, err := json.Marshal(note)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal into map failed: %v", err)
	}

	omittedKeys := []string{"encounter_id", "title", "subjective", "objective", "assessment", "plan", "note_text", "signed_by", "signed_at", "cosigned_by", "cosigned_at", "amended_by", "amended_at", "amended_reason"}
	for _, key := range omittedKeys {
		if _, ok := m[key]; ok {
			t.Errorf("expected key %q to be omitted for nil field, but it was present", key)
		}
	}
}
