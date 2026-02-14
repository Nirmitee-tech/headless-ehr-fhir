package inbox

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

func TestMessagePool_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	orgID := uuid.New()
	deptID := uuid.New()

	original := &MessagePool{
		ID:             uuid.New(),
		PoolName:       "Internal Medicine Pool",
		PoolType:       "department",
		OrganizationID: ptrUUID(orgID),
		DepartmentID:   ptrUUID(deptID),
		Description:    ptrStr("Messages for internal medicine department"),
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded MessagePool
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.PoolName != original.PoolName {
		t.Errorf("PoolName mismatch: got %q, want %q", decoded.PoolName, original.PoolName)
	}
	if decoded.PoolType != original.PoolType {
		t.Errorf("PoolType mismatch: got %q, want %q", decoded.PoolType, original.PoolType)
	}
	if decoded.IsActive != original.IsActive {
		t.Errorf("IsActive mismatch: got %v, want %v", decoded.IsActive, original.IsActive)
	}
	if *decoded.OrganizationID != *original.OrganizationID {
		t.Errorf("OrganizationID mismatch")
	}
	if *decoded.DepartmentID != *original.DepartmentID {
		t.Errorf("DepartmentID mismatch")
	}
	if *decoded.Description != *original.Description {
		t.Errorf("Description mismatch")
	}
}

func TestMessagePool_OptionalFieldsNil(t *testing.T) {
	m := &MessagePool{
		ID:        uuid.New(),
		PoolName:  "Empty Pool",
		PoolType:  "individual",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"organization_id"`) {
		t.Error("nil OrganizationID should be omitted")
	}
	if strings.Contains(s, `"department_id"`) {
		t.Error("nil DepartmentID should be omitted")
	}
	if strings.Contains(s, `"description"`) {
		t.Error("nil Description should be omitted")
	}
}

func TestInboxMessage_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	patientID := uuid.New()
	encounterID := uuid.New()
	senderID := uuid.New()
	recipientID := uuid.New()
	poolID := uuid.New()
	parentID := uuid.New()
	threadID := uuid.New()
	dueDate := now.Add(48 * time.Hour)
	readAt := now.Add(30 * time.Minute)

	original := &InboxMessage{
		ID:          uuid.New(),
		MessageType: "result_note",
		Priority:    "high",
		Subject:     "Lab Results - Critical Value",
		Body:        ptrStr("Potassium level 6.2 mEq/L - requires immediate attention"),
		PatientID:   ptrUUID(patientID),
		EncounterID: ptrUUID(encounterID),
		SenderID:    ptrUUID(senderID),
		RecipientID: ptrUUID(recipientID),
		PoolID:      ptrUUID(poolID),
		Status:      "unread",
		ParentID:    ptrUUID(parentID),
		ThreadID:    ptrUUID(threadID),
		IsUrgent:    true,
		DueDate:     ptrTime(dueDate),
		ReadAt:      ptrTime(readAt),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded InboxMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.MessageType != original.MessageType {
		t.Errorf("MessageType mismatch: got %q, want %q", decoded.MessageType, original.MessageType)
	}
	if decoded.Priority != original.Priority {
		t.Errorf("Priority mismatch: got %q, want %q", decoded.Priority, original.Priority)
	}
	if decoded.Subject != original.Subject {
		t.Errorf("Subject mismatch: got %q, want %q", decoded.Subject, original.Subject)
	}
	if decoded.Status != original.Status {
		t.Errorf("Status mismatch: got %q, want %q", decoded.Status, original.Status)
	}
	if decoded.IsUrgent != original.IsUrgent {
		t.Errorf("IsUrgent mismatch: got %v, want %v", decoded.IsUrgent, original.IsUrgent)
	}
	if *decoded.Body != *original.Body {
		t.Errorf("Body mismatch")
	}
	if *decoded.PatientID != *original.PatientID {
		t.Errorf("PatientID mismatch")
	}
	if *decoded.SenderID != *original.SenderID {
		t.Errorf("SenderID mismatch")
	}
	if *decoded.RecipientID != *original.RecipientID {
		t.Errorf("RecipientID mismatch")
	}
	if *decoded.ThreadID != *original.ThreadID {
		t.Errorf("ThreadID mismatch")
	}
}

func TestInboxMessage_OptionalFieldsNil(t *testing.T) {
	m := &InboxMessage{
		ID:          uuid.New(),
		MessageType: "general",
		Priority:    "normal",
		Subject:     "Test",
		Status:      "unread",
		IsUrgent:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"body"`) {
		t.Error("nil Body should be omitted")
	}
	if strings.Contains(s, `"patient_id"`) {
		t.Error("nil PatientID should be omitted")
	}
	if strings.Contains(s, `"sender_id"`) {
		t.Error("nil SenderID should be omitted")
	}
	if strings.Contains(s, `"recipient_id"`) {
		t.Error("nil RecipientID should be omitted")
	}
	if strings.Contains(s, `"parent_id"`) {
		t.Error("nil ParentID should be omitted")
	}
	if strings.Contains(s, `"due_date"`) {
		t.Error("nil DueDate should be omitted")
	}
	if strings.Contains(s, `"read_at"`) {
		t.Error("nil ReadAt should be omitted")
	}
	if strings.Contains(s, `"completed_at"`) {
		t.Error("nil CompletedAt should be omitted")
	}
}
