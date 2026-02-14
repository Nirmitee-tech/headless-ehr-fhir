package provenance

import (
	"testing"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

func ptrStr(s string) *string        { return &s }
func ptrTime(t time.Time) *time.Time { return &t }

// ---------------------------------------------------------------------------
// Provenance.ToFHIR
// ---------------------------------------------------------------------------

func TestProvenance_ToFHIR_RequiredFields(t *testing.T) {
	now := time.Now()
	recorded := time.Date(2024, 6, 1, 10, 30, 0, 0, time.UTC)

	p := &Provenance{
		ID:         uuid.New(),
		FHIRID:     "prov-001",
		TargetType: "Patient",
		TargetID:   "abc-123",
		Recorded:   recorded,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	result := p.ToFHIR()

	if result["resourceType"] != "Provenance" {
		t.Errorf("resourceType = %v, want Provenance", result["resourceType"])
	}
	if result["id"] != "prov-001" {
		t.Errorf("id = %v, want prov-001", result["id"])
	}

	targets, ok := result["target"].([]fhir.Reference)
	if !ok || len(targets) == 0 {
		t.Fatal("target missing or wrong type")
	}
	if targets[0].Reference != "Patient/abc-123" {
		t.Errorf("target[0].Reference = %v, want Patient/abc-123", targets[0].Reference)
	}

	if result["recorded"] != recorded.Format(time.RFC3339) {
		t.Errorf("recorded = %v, want %v", result["recorded"], recorded.Format(time.RFC3339))
	}

	meta, ok := result["meta"].(fhir.Meta)
	if !ok {
		t.Fatal("meta is not fhir.Meta")
	}
	if meta.LastUpdated != now {
		t.Errorf("meta.LastUpdated = %v, want %v", meta.LastUpdated, now)
	}
}

func TestProvenance_ToFHIR_OptionalFields(t *testing.T) {
	now := time.Now()
	recorded := time.Date(2024, 6, 1, 10, 30, 0, 0, time.UTC)

	p := &Provenance{
		ID:              uuid.New(),
		FHIRID:          "prov-opt",
		TargetType:      "Observation",
		TargetID:        "obs-456",
		Recorded:        recorded,
		ActivityCode:    ptrStr("CREATE"),
		ActivityDisplay: ptrStr("create"),
		ReasonCode:      ptrStr("TREAT"),
		ReasonDisplay:   ptrStr("Treatment"),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result := p.ToFHIR()

	// activity
	activity, ok := result["activity"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("activity missing or wrong type")
	}
	if len(activity.Coding) == 0 || activity.Coding[0].Code != "CREATE" {
		t.Errorf("activity.Coding[0].Code = %v, want CREATE", activity.Coding[0].Code)
	}
	if activity.Coding[0].Display != "create" {
		t.Errorf("activity.Coding[0].Display = %v, want create", activity.Coding[0].Display)
	}

	// reason
	reasons, ok := result["reason"].([]fhir.CodeableConcept)
	if !ok || len(reasons) == 0 {
		t.Fatal("reason missing or wrong type")
	}
	if len(reasons[0].Coding) == 0 || reasons[0].Coding[0].Code != "TREAT" {
		t.Errorf("reason[0].Coding[0].Code = %v, want TREAT", reasons[0].Coding[0].Code)
	}
	if reasons[0].Coding[0].Display != "Treatment" {
		t.Errorf("reason[0].Coding[0].Display = %v, want Treatment", reasons[0].Coding[0].Display)
	}
}

func TestProvenance_ToFHIR_OptionalFieldsNil(t *testing.T) {
	now := time.Now()
	recorded := time.Date(2024, 6, 1, 10, 30, 0, 0, time.UTC)

	p := &Provenance{
		ID:         uuid.New(),
		FHIRID:     "prov-nil",
		TargetType: "Patient",
		TargetID:   "pat-789",
		Recorded:   recorded,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	result := p.ToFHIR()

	absentKeys := []string{"activity", "reason"}
	for _, key := range absentKeys {
		if _, ok := result[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}

// ---------------------------------------------------------------------------
// strVal helper
// ---------------------------------------------------------------------------

func TestStrVal_Nil(t *testing.T) {
	if v := strVal(nil); v != "" {
		t.Errorf("strVal(nil) = %q, want empty", v)
	}
}

func TestStrVal_NonNil(t *testing.T) {
	s := "hello"
	if v := strVal(&s); v != "hello" {
		t.Errorf("strVal(&hello) = %q, want hello", v)
	}
}
