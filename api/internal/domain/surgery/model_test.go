package surgery

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

func TestORRoom_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	locationID := uuid.New()
	decontamAt := now.Add(-2 * time.Hour)

	original := &ORRoom{
		ID:               uuid.New(),
		Name:             "OR-1",
		LocationID:       ptrUUID(locationID),
		Status:           "available",
		RoomType:         ptrStr("general"),
		Equipment:        ptrStr("laparoscopic tower, C-arm"),
		IsActive:         true,
		DecontaminatedAt: ptrTime(decontamAt),
		Note:             ptrStr("recently updated equipment"),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded ORRoom
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, original.Name)
	}
	if decoded.Status != original.Status {
		t.Errorf("Status mismatch: got %q, want %q", decoded.Status, original.Status)
	}
	if decoded.IsActive != original.IsActive {
		t.Errorf("IsActive mismatch: got %v, want %v", decoded.IsActive, original.IsActive)
	}
	if *decoded.LocationID != *original.LocationID {
		t.Errorf("LocationID mismatch")
	}
	if *decoded.RoomType != *original.RoomType {
		t.Errorf("RoomType mismatch")
	}
	if *decoded.Equipment != *original.Equipment {
		t.Errorf("Equipment mismatch")
	}
	if *decoded.Note != *original.Note {
		t.Errorf("Note mismatch")
	}
}

func TestORRoom_OptionalFieldsNil(t *testing.T) {
	m := &ORRoom{
		ID:        uuid.New(),
		Name:      "OR-2",
		Status:    "available",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"location_id"`) {
		t.Error("nil LocationID should be omitted")
	}
	if strings.Contains(s, `"room_type"`) {
		t.Error("nil RoomType should be omitted")
	}
	if strings.Contains(s, `"equipment"`) {
		t.Error("nil Equipment should be omitted")
	}
	if strings.Contains(s, `"decontaminated_at"`) {
		t.Error("nil DecontaminatedAt should be omitted")
	}
	if strings.Contains(s, `"note"`) {
		t.Error("nil Note should be omitted")
	}
}

func TestSurgicalCase_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	encounterID := uuid.New()
	anesthesiologistID := uuid.New()
	orRoomID := uuid.New()
	scheduledStart := now.Add(2 * time.Hour)
	scheduledEnd := now.Add(5 * time.Hour)
	actualStart := now.Add(2*time.Hour + 15*time.Minute)
	actualEnd := now.Add(4*time.Hour + 45*time.Minute)

	original := &SurgicalCase{
		ID:                 uuid.New(),
		PatientID:          uuid.New(),
		EncounterID:        ptrUUID(encounterID),
		PrimarySurgeonID:   uuid.New(),
		AnesthesiologistID: ptrUUID(anesthesiologistID),
		ORRoomID:           ptrUUID(orRoomID),
		Status:             "completed",
		CaseClass:          ptrStr("elective"),
		ASAClass:           ptrStr("III"),
		WoundClass:         ptrStr("clean"),
		ScheduledDate:      now,
		ScheduledStart:     ptrTime(scheduledStart),
		ScheduledEnd:       ptrTime(scheduledEnd),
		ActualStart:        ptrTime(actualStart),
		ActualEnd:          ptrTime(actualEnd),
		AnesthesiaType:     ptrStr("general"),
		Laterality:         ptrStr("right"),
		PreOpDiagnosis:     ptrStr("right inguinal hernia"),
		PostOpDiagnosis:    ptrStr("right inguinal hernia, direct"),
		Note:               ptrStr("uncomplicated procedure"),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded SurgicalCase
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.PatientID != original.PatientID {
		t.Errorf("PatientID mismatch")
	}
	if decoded.PrimarySurgeonID != original.PrimarySurgeonID {
		t.Errorf("PrimarySurgeonID mismatch")
	}
	if decoded.Status != original.Status {
		t.Errorf("Status mismatch: got %q, want %q", decoded.Status, original.Status)
	}
	if *decoded.CaseClass != *original.CaseClass {
		t.Errorf("CaseClass mismatch")
	}
	if *decoded.ASAClass != *original.ASAClass {
		t.Errorf("ASAClass mismatch")
	}
	if *decoded.AnesthesiaType != *original.AnesthesiaType {
		t.Errorf("AnesthesiaType mismatch")
	}
	if *decoded.PreOpDiagnosis != *original.PreOpDiagnosis {
		t.Errorf("PreOpDiagnosis mismatch")
	}
	if *decoded.PostOpDiagnosis != *original.PostOpDiagnosis {
		t.Errorf("PostOpDiagnosis mismatch")
	}
	if *decoded.ORRoomID != *original.ORRoomID {
		t.Errorf("ORRoomID mismatch")
	}
}

func TestSurgicalCase_OptionalFieldsNil(t *testing.T) {
	m := &SurgicalCase{
		ID:               uuid.New(),
		PatientID:        uuid.New(),
		PrimarySurgeonID: uuid.New(),
		Status:           "scheduled",
		ScheduledDate:    time.Now(),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"encounter_id"`) {
		t.Error("nil EncounterID should be omitted")
	}
	if strings.Contains(s, `"anesthesiologist_id"`) {
		t.Error("nil AnesthesiologistID should be omitted")
	}
	if strings.Contains(s, `"case_class"`) {
		t.Error("nil CaseClass should be omitted")
	}
	if strings.Contains(s, `"cancel_reason"`) {
		t.Error("nil CancelReason should be omitted")
	}
	if strings.Contains(s, `"actual_start"`) {
		t.Error("nil ActualStart should be omitted")
	}
}
