package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// =========== ParseAvailabilityRequest Tests ===========

func TestParseAvailabilityRequest_Valid(t *testing.T) {
	params := url.Values{}
	params.Set("start", "2025-06-01T08:00:00Z")
	params.Set("end", "2025-06-01T17:00:00Z")
	params.Set("duration", "30")
	params.Set("slot-type", "routine")
	params.Set("service-type", "general-practice")
	params.Set("specialty", "general-practice")
	params.Set("practitioner", "Practitioner/123")
	params.Set("location", "Location/456")
	params.Set("status", "free,busy-tentative")
	params.Set("_include", "Schedule")

	req, err := ParseAvailabilityRequest(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStart, _ := time.Parse(time.RFC3339, "2025-06-01T08:00:00Z")
	expectedEnd, _ := time.Parse(time.RFC3339, "2025-06-01T17:00:00Z")

	if !req.Start.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, req.Start)
	}
	if !req.End.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, req.End)
	}
	if req.Duration != 30 {
		t.Errorf("expected duration 30, got %d", req.Duration)
	}
	if req.SlotType != "routine" {
		t.Errorf("expected slot-type 'routine', got %q", req.SlotType)
	}
	if req.ServiceType != "general-practice" {
		t.Errorf("expected service-type 'general-practice', got %q", req.ServiceType)
	}
	if req.Specialty != "general-practice" {
		t.Errorf("expected specialty 'general-practice', got %q", req.Specialty)
	}
	if req.Practitioner != "Practitioner/123" {
		t.Errorf("expected practitioner 'Practitioner/123', got %q", req.Practitioner)
	}
	if req.Location != "Location/456" {
		t.Errorf("expected location 'Location/456', got %q", req.Location)
	}
	if len(req.Status) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(req.Status))
	}
	if req.Status[0] != SlotStatusFree {
		t.Errorf("expected status[0] 'free', got %q", req.Status[0])
	}
	if req.Status[1] != SlotStatusBusyTentative {
		t.Errorf("expected status[1] 'busy-tentative', got %q", req.Status[1])
	}
	if !req.IncludeSchedule {
		t.Error("expected IncludeSchedule to be true")
	}
}

func TestParseAvailabilityRequest_MissingDates(t *testing.T) {
	params := url.Values{}
	params.Set("duration", "30")

	req, err := ParseAvailabilityRequest(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Missing dates should yield zero times, not an error (validation catches this).
	if !req.Start.IsZero() {
		t.Errorf("expected zero start time, got %v", req.Start)
	}
	if !req.End.IsZero() {
		t.Errorf("expected zero end time, got %v", req.End)
	}
}

func TestParseAvailabilityRequest_InvalidStartDate(t *testing.T) {
	params := url.Values{}
	params.Set("start", "not-a-date")
	params.Set("end", "2025-06-01T17:00:00Z")

	_, err := ParseAvailabilityRequest(params)
	if err == nil {
		t.Fatal("expected error for invalid start date")
	}
	if !strings.Contains(err.Error(), "start") {
		t.Errorf("expected error to mention 'start', got %q", err.Error())
	}
}

func TestParseAvailabilityRequest_InvalidEndDate(t *testing.T) {
	params := url.Values{}
	params.Set("start", "2025-06-01T08:00:00Z")
	params.Set("end", "not-a-date")

	_, err := ParseAvailabilityRequest(params)
	if err == nil {
		t.Fatal("expected error for invalid end date")
	}
	if !strings.Contains(err.Error(), "end") {
		t.Errorf("expected error to mention 'end', got %q", err.Error())
	}
}

func TestParseAvailabilityRequest_WithAllParams(t *testing.T) {
	params := url.Values{}
	params.Set("start", "2025-06-01T00:00:00Z")
	params.Set("end", "2025-06-07T23:59:59Z")
	params.Set("duration", "60")
	params.Set("slot-type", "urgent")
	params.Set("service-type", "cardiology")
	params.Set("specialty", "cardiology")
	params.Set("practitioner", "Practitioner/dr-smith")
	params.Set("location", "Location/clinic-a")
	params.Set("status", "free")
	params.Set("_include", "Schedule")

	req, err := ParseAvailabilityRequest(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Duration != 60 {
		t.Errorf("expected duration 60, got %d", req.Duration)
	}
	if req.SlotType != "urgent" {
		t.Errorf("expected slot-type 'urgent', got %q", req.SlotType)
	}
	if req.IncludeSchedule != true {
		t.Error("expected IncludeSchedule to be true when _include=Schedule")
	}
}

func TestParseAvailabilityRequest_DefaultStatus(t *testing.T) {
	params := url.Values{}
	params.Set("start", "2025-06-01T08:00:00Z")
	params.Set("end", "2025-06-01T17:00:00Z")

	req, err := ParseAvailabilityRequest(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// When no status specified, should default to empty (the handler adds default).
	if len(req.Status) != 0 {
		t.Errorf("expected empty status when not specified, got %v", req.Status)
	}
}

func TestParseAvailabilityRequest_InvalidDuration(t *testing.T) {
	params := url.Values{}
	params.Set("start", "2025-06-01T08:00:00Z")
	params.Set("end", "2025-06-01T17:00:00Z")
	params.Set("duration", "abc")

	_, err := ParseAvailabilityRequest(params)
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
	if !strings.Contains(err.Error(), "duration") {
		t.Errorf("expected error to mention 'duration', got %q", err.Error())
	}
}

// =========== ValidateAvailabilityRequest Tests ===========

func TestValidateAvailabilityRequest_Valid(t *testing.T) {
	start := time.Date(2025, 6, 1, 8, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 1, 17, 0, 0, 0, time.UTC)
	req := &AvailabilityRequest{
		Start:    start,
		End:      end,
		Duration: 30,
	}

	issues := ValidateAvailabilityRequest(req)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d: %v", len(issues), issues)
	}
}

func TestValidateAvailabilityRequest_MissingStart(t *testing.T) {
	end := time.Date(2025, 6, 1, 17, 0, 0, 0, time.UTC)
	req := &AvailabilityRequest{
		End: end,
	}

	issues := ValidateAvailabilityRequest(req)
	if len(issues) == 0 {
		t.Fatal("expected validation issues for missing start")
	}

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "start") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected issue mentioning 'start'")
	}
}

func TestValidateAvailabilityRequest_MissingEnd(t *testing.T) {
	start := time.Date(2025, 6, 1, 8, 0, 0, 0, time.UTC)
	req := &AvailabilityRequest{
		Start: start,
	}

	issues := ValidateAvailabilityRequest(req)
	if len(issues) == 0 {
		t.Fatal("expected validation issues for missing end")
	}

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "end") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected issue mentioning 'end'")
	}
}

func TestValidateAvailabilityRequest_StartAfterEnd(t *testing.T) {
	start := time.Date(2025, 6, 2, 8, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 1, 17, 0, 0, 0, time.UTC)
	req := &AvailabilityRequest{
		Start: start,
		End:   end,
	}

	issues := ValidateAvailabilityRequest(req)
	if len(issues) == 0 {
		t.Fatal("expected validation issues for start after end")
	}

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "start") && strings.Contains(issue.Diagnostics, "end") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected issue mentioning start/end ordering")
	}
}

func TestValidateAvailabilityRequest_NegativeDuration(t *testing.T) {
	start := time.Date(2025, 6, 1, 8, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 1, 17, 0, 0, 0, time.UTC)
	req := &AvailabilityRequest{
		Start:    start,
		End:      end,
		Duration: -10,
	}

	issues := ValidateAvailabilityRequest(req)
	if len(issues) == 0 {
		t.Fatal("expected validation issues for negative duration")
	}
}

// =========== ParseTimeOfDay Tests ===========

func TestParseTimeOfDay_Valid(t *testing.T) {
	tests := []struct {
		input   string
		hour    int
		minute  int
	}{
		{"08:00", 8, 0},
		{"13:30", 13, 30},
		{"00:00", 0, 0},
		{"23:59", 23, 59},
	}

	for _, tc := range tests {
		h, m, err := ParseTimeOfDay(tc.input)
		if err != nil {
			t.Errorf("ParseTimeOfDay(%q): unexpected error: %v", tc.input, err)
			continue
		}
		if h != tc.hour || m != tc.minute {
			t.Errorf("ParseTimeOfDay(%q): expected %d:%d, got %d:%d", tc.input, tc.hour, tc.minute, h, m)
		}
	}
}

func TestParseTimeOfDay_Invalid(t *testing.T) {
	invalids := []string{
		"",
		"8:00",
		"abc",
		"25:00",
		"12:60",
		"12:345",
		"-1:00",
		"08:0a",
	}

	for _, input := range invalids {
		_, _, err := ParseTimeOfDay(input)
		if err == nil {
			t.Errorf("ParseTimeOfDay(%q): expected error, got nil", input)
		}
	}
}

func TestParseTimeOfDay_EdgeCases(t *testing.T) {
	// Midnight
	h, m, err := ParseTimeOfDay("00:00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h != 0 || m != 0 {
		t.Errorf("expected 0:0, got %d:%d", h, m)
	}

	// End of day
	h, m, err = ParseTimeOfDay("23:59")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h != 23 || m != 59 {
		t.Errorf("expected 23:59, got %d:%d", h, m)
	}
}

// =========== GenerateTimeSlots Tests ===========

func TestGenerateTimeSlots_Basic(t *testing.T) {
	rule := &AvailabilityRule{
		DaysOfWeek:   []time.Weekday{time.Monday},
		StartTime:    "09:00",
		EndTime:      "12:00",
		SlotDuration: 30,
	}

	// 2025-06-02 is a Monday.
	start := time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 2, 23, 59, 59, 0, time.UTC)

	slots := GenerateTimeSlots(rule, start, end)
	// 09:00-12:00 with 30 min slots = 6 slots
	if len(slots) != 6 {
		t.Errorf("expected 6 slots, got %d", len(slots))
	}

	if len(slots) > 0 {
		first := slots[0]
		expectedStart := time.Date(2025, 6, 2, 9, 0, 0, 0, time.UTC)
		if !first.Start.Equal(expectedStart) {
			t.Errorf("expected first slot start %v, got %v", expectedStart, first.Start)
		}
		if first.Duration != 30 {
			t.Errorf("expected duration 30, got %d", first.Duration)
		}
	}

	if len(slots) > 5 {
		last := slots[5]
		expectedLastStart := time.Date(2025, 6, 2, 11, 30, 0, 0, time.UTC)
		if !last.Start.Equal(expectedLastStart) {
			t.Errorf("expected last slot start %v, got %v", expectedLastStart, last.Start)
		}
		expectedLastEnd := time.Date(2025, 6, 2, 12, 0, 0, 0, time.UTC)
		if !last.End.Equal(expectedLastEnd) {
			t.Errorf("expected last slot end %v, got %v", expectedLastEnd, last.End)
		}
	}
}

func TestGenerateTimeSlots_WithBreaks(t *testing.T) {
	rule := &AvailabilityRule{
		DaysOfWeek:   []time.Weekday{time.Monday},
		StartTime:    "09:00",
		EndTime:      "17:00",
		SlotDuration: 60,
		BreakStart:   "12:00",
		BreakEnd:     "13:00",
	}

	start := time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 2, 23, 59, 59, 0, time.UTC)

	slots := GenerateTimeSlots(rule, start, end)
	// 09-12 = 3 slots, 13-17 = 4 slots = 7 total (skipping 12-13 break)
	if len(slots) != 7 {
		t.Errorf("expected 7 slots (with break), got %d", len(slots))
	}

	// Verify no slot overlaps with the break period.
	breakStart := time.Date(2025, 6, 2, 12, 0, 0, 0, time.UTC)
	breakEnd := time.Date(2025, 6, 2, 13, 0, 0, 0, time.UTC)
	for _, s := range slots {
		if OverlapsTimeRange(s.Start, s.End, breakStart, breakEnd) {
			t.Errorf("slot %v-%v overlaps with break %v-%v", s.Start, s.End, breakStart, breakEnd)
		}
	}
}

func TestGenerateTimeSlots_NoMatchDay(t *testing.T) {
	rule := &AvailabilityRule{
		DaysOfWeek:   []time.Weekday{time.Saturday},
		StartTime:    "09:00",
		EndTime:      "17:00",
		SlotDuration: 30,
	}

	// 2025-06-02 is a Monday, rule is for Saturday.
	start := time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 2, 23, 59, 59, 0, time.UTC)

	slots := GenerateTimeSlots(rule, start, end)
	if len(slots) != 0 {
		t.Errorf("expected 0 slots for non-matching day, got %d", len(slots))
	}
}

func TestGenerateTimeSlots_MultiDay(t *testing.T) {
	rule := &AvailabilityRule{
		DaysOfWeek:   []time.Weekday{time.Monday, time.Tuesday},
		StartTime:    "09:00",
		EndTime:      "10:00",
		SlotDuration: 30,
	}

	// Monday to Tuesday.
	start := time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 3, 23, 59, 59, 0, time.UTC)

	slots := GenerateTimeSlots(rule, start, end)
	// 2 slots per day x 2 days = 4 slots
	if len(slots) != 4 {
		t.Errorf("expected 4 slots across 2 days, got %d", len(slots))
	}
}

func TestGenerateTimeSlots_ZeroDuration(t *testing.T) {
	rule := &AvailabilityRule{
		DaysOfWeek:   []time.Weekday{time.Monday},
		StartTime:    "09:00",
		EndTime:      "10:00",
		SlotDuration: 0,
	}

	start := time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 2, 23, 59, 59, 0, time.UTC)

	slots := GenerateTimeSlots(rule, start, end)
	if len(slots) != 0 {
		t.Errorf("expected 0 slots for zero duration, got %d", len(slots))
	}
}

// =========== OverlapsTimeRange Tests ===========

func TestOverlapsTimeRange_Before(t *testing.T) {
	// Range 1 is entirely before range 2.
	s1 := time.Date(2025, 6, 1, 8, 0, 0, 0, time.UTC)
	e1 := time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC)
	s2 := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	e2 := time.Date(2025, 6, 1, 11, 0, 0, 0, time.UTC)

	if OverlapsTimeRange(s1, e1, s2, e2) {
		t.Error("expected no overlap when range1 is before range2")
	}
}

func TestOverlapsTimeRange_After(t *testing.T) {
	// Range 1 is entirely after range 2.
	s1 := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	e1 := time.Date(2025, 6, 1, 13, 0, 0, 0, time.UTC)
	s2 := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	e2 := time.Date(2025, 6, 1, 11, 0, 0, 0, time.UTC)

	if OverlapsTimeRange(s1, e1, s2, e2) {
		t.Error("expected no overlap when range1 is after range2")
	}
}

func TestOverlapsTimeRange_PartialOverlap(t *testing.T) {
	s1 := time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC)
	e1 := time.Date(2025, 6, 1, 11, 0, 0, 0, time.UTC)
	s2 := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	e2 := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	if !OverlapsTimeRange(s1, e1, s2, e2) {
		t.Error("expected overlap for partial overlap")
	}
}

func TestOverlapsTimeRange_Contained(t *testing.T) {
	// Range 2 is fully contained within range 1.
	s1 := time.Date(2025, 6, 1, 8, 0, 0, 0, time.UTC)
	e1 := time.Date(2025, 6, 1, 17, 0, 0, 0, time.UTC)
	s2 := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	e2 := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	if !OverlapsTimeRange(s1, e1, s2, e2) {
		t.Error("expected overlap when range2 is contained in range1")
	}
}

func TestOverlapsTimeRange_ExactMatch(t *testing.T) {
	s := time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC)
	e := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)

	if !OverlapsTimeRange(s, e, s, e) {
		t.Error("expected overlap for exact same range")
	}
}

func TestOverlapsTimeRange_Adjacent(t *testing.T) {
	// Adjacent ranges (end of one == start of another) should NOT overlap.
	s1 := time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC)
	e1 := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	s2 := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	e2 := time.Date(2025, 6, 1, 11, 0, 0, 0, time.UTC)

	if OverlapsTimeRange(s1, e1, s2, e2) {
		t.Error("expected no overlap for adjacent ranges (end == start)")
	}
}

// =========== MergeAvailability Tests ===========

func TestMergeAvailability_NoConflicts(t *testing.T) {
	available := []TimeSlot{
		{Start: time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 9, 30, 0, 0, time.UTC), Duration: 30},
		{Start: time.Date(2025, 6, 1, 9, 30, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC), Duration: 30},
		{Start: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 10, 30, 0, 0, time.UTC), Duration: 30},
	}
	busy := []TimeSlot{}

	result := MergeAvailability(available, busy)
	if len(result) != 3 {
		t.Errorf("expected 3 slots with no conflicts, got %d", len(result))
	}
}

func TestMergeAvailability_PartialOverlap(t *testing.T) {
	available := []TimeSlot{
		{Start: time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 9, 30, 0, 0, time.UTC), Duration: 30},
		{Start: time.Date(2025, 6, 1, 9, 30, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC), Duration: 30},
		{Start: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 10, 30, 0, 0, time.UTC), Duration: 30},
	}
	busy := []TimeSlot{
		{Start: time.Date(2025, 6, 1, 9, 15, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 9, 45, 0, 0, time.UTC), Duration: 30},
	}

	result := MergeAvailability(available, busy)
	// The busy slot overlaps with slots 1 and 2, so only slot 3 remains.
	if len(result) != 1 {
		t.Errorf("expected 1 slot after partial overlap removal, got %d", len(result))
	}
}

func TestMergeAvailability_FullOverlap(t *testing.T) {
	available := []TimeSlot{
		{Start: time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 9, 30, 0, 0, time.UTC), Duration: 30},
	}
	busy := []TimeSlot{
		{Start: time.Date(2025, 6, 1, 8, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC), Duration: 120},
	}

	result := MergeAvailability(available, busy)
	if len(result) != 0 {
		t.Errorf("expected 0 slots after full overlap, got %d", len(result))
	}
}

func TestMergeAvailability_MultipleBusy(t *testing.T) {
	available := []TimeSlot{
		{Start: time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 9, 30, 0, 0, time.UTC), Duration: 30},
		{Start: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 10, 30, 0, 0, time.UTC), Duration: 30},
		{Start: time.Date(2025, 6, 1, 11, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 11, 30, 0, 0, time.UTC), Duration: 30},
	}
	busy := []TimeSlot{
		{Start: time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 9, 30, 0, 0, time.UTC), Duration: 30},
		{Start: time.Date(2025, 6, 1, 11, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 11, 30, 0, 0, time.UTC), Duration: 30},
	}

	result := MergeAvailability(available, busy)
	// Only the 10:00-10:30 slot should remain.
	if len(result) != 1 {
		t.Errorf("expected 1 slot after multiple busy removal, got %d", len(result))
	}
	if len(result) > 0 {
		expectedStart := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
		if !result[0].Start.Equal(expectedStart) {
			t.Errorf("expected remaining slot at 10:00, got %v", result[0].Start)
		}
	}
}

func TestMergeAvailability_EmptyAvailable(t *testing.T) {
	result := MergeAvailability(nil, []TimeSlot{
		{Start: time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 9, 30, 0, 0, time.UTC), Duration: 30},
	})
	if len(result) != 0 {
		t.Errorf("expected 0 slots for empty available, got %d", len(result))
	}
}

// =========== FilterSlotsByDuration Tests ===========

func TestFilterSlotsByDuration_TooShort(t *testing.T) {
	slots := []TimeSlot{
		{Start: time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 9, 15, 0, 0, time.UTC), Duration: 15},
		{Start: time.Date(2025, 6, 1, 9, 15, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 9, 30, 0, 0, time.UTC), Duration: 15},
	}

	result := FilterSlotsByDuration(slots, 30)
	if len(result) != 0 {
		t.Errorf("expected 0 slots for min duration 30, got %d", len(result))
	}
}

func TestFilterSlotsByDuration_ExactMatch(t *testing.T) {
	slots := []TimeSlot{
		{Start: time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 9, 30, 0, 0, time.UTC), Duration: 30},
	}

	result := FilterSlotsByDuration(slots, 30)
	if len(result) != 1 {
		t.Errorf("expected 1 slot for exact duration match, got %d", len(result))
	}
}

func TestFilterSlotsByDuration_Longer(t *testing.T) {
	slots := []TimeSlot{
		{Start: time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC), Duration: 60},
		{Start: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 10, 15, 0, 0, time.UTC), Duration: 15},
	}

	result := FilterSlotsByDuration(slots, 30)
	if len(result) != 1 {
		t.Errorf("expected 1 slot meeting min duration 30, got %d", len(result))
	}
}

func TestFilterSlotsByDuration_ZeroMinDuration(t *testing.T) {
	slots := []TimeSlot{
		{Start: time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 1, 9, 15, 0, 0, time.UTC), Duration: 15},
	}

	result := FilterSlotsByDuration(slots, 0)
	if len(result) != 1 {
		t.Errorf("expected all slots for min duration 0, got %d", len(result))
	}
}

// =========== BuildSlotResource Tests ===========

func TestBuildSlotResource_Free(t *testing.T) {
	avail := &ScheduleAvailability{
		Start:       time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC),
		End:         time.Date(2025, 6, 1, 9, 30, 0, 0, time.UTC),
		SlotType:    "routine",
		Status:      "free",
		Comment:     "Available for walk-ins",
		ServiceType: "general-practice",
		Specialty:   "general-practice",
		Actor:       "Practitioner/dr-jones",
	}

	slot := BuildSlotResource(avail, "Schedule/sched-1")

	if slot["resourceType"] != "Slot" {
		t.Errorf("expected resourceType 'Slot', got %v", slot["resourceType"])
	}
	if slot["status"] != "free" {
		t.Errorf("expected status 'free', got %v", slot["status"])
	}
	if slot["start"] != "2025-06-01T09:00:00Z" {
		t.Errorf("expected start '2025-06-01T09:00:00Z', got %v", slot["start"])
	}
	if slot["end"] != "2025-06-01T09:30:00Z" {
		t.Errorf("expected end '2025-06-01T09:30:00Z', got %v", slot["end"])
	}
	if slot["comment"] != "Available for walk-ins" {
		t.Errorf("expected comment, got %v", slot["comment"])
	}

	schedRef, ok := slot["schedule"].(map[string]interface{})
	if !ok {
		t.Fatal("expected schedule reference")
	}
	if schedRef["reference"] != "Schedule/sched-1" {
		t.Errorf("expected schedule reference 'Schedule/sched-1', got %v", schedRef["reference"])
	}
}

func TestBuildSlotResource_Busy(t *testing.T) {
	avail := &ScheduleAvailability{
		Start:  time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC),
		End:    time.Date(2025, 6, 1, 10, 30, 0, 0, time.UTC),
		Status: "busy",
	}

	slot := BuildSlotResource(avail, "Schedule/sched-2")
	if slot["status"] != "busy" {
		t.Errorf("expected status 'busy', got %v", slot["status"])
	}
}

func TestBuildSlotResource_WithAllFields(t *testing.T) {
	avail := &ScheduleAvailability{
		Start:       time.Date(2025, 6, 1, 14, 0, 0, 0, time.UTC),
		End:         time.Date(2025, 6, 1, 15, 0, 0, 0, time.UTC),
		SlotType:    "urgent",
		Status:      "busy-tentative",
		Comment:     "Pending confirmation",
		ServiceType: "cardiology",
		Specialty:   "cardiology",
		Actor:       "Practitioner/dr-smith",
	}

	slot := BuildSlotResource(avail, "Schedule/sched-3")
	if slot["status"] != "busy-tentative" {
		t.Errorf("expected status 'busy-tentative', got %v", slot["status"])
	}

	// Verify specialty is set.
	specialty, ok := slot["specialty"].([]interface{})
	if !ok || len(specialty) == 0 {
		t.Fatal("expected specialty array")
	}
}

// =========== BuildScheduleResource Tests ===========

func TestBuildScheduleResource(t *testing.T) {
	sched := BuildScheduleResource("sched-1", "Practitioner/dr-jones", "general-practice", "general-practice")

	if sched["resourceType"] != "Schedule" {
		t.Errorf("expected resourceType 'Schedule', got %v", sched["resourceType"])
	}
	if sched["id"] != "sched-1" {
		t.Errorf("expected id 'sched-1', got %v", sched["id"])
	}

	actors, ok := sched["actor"].([]interface{})
	if !ok || len(actors) == 0 {
		t.Fatal("expected actor array")
	}
	actorRef := actors[0].(map[string]interface{})
	if actorRef["reference"] != "Practitioner/dr-jones" {
		t.Errorf("expected actor reference 'Practitioner/dr-jones', got %v", actorRef["reference"])
	}
}

func TestBuildScheduleResource_EmptyOptionalFields(t *testing.T) {
	sched := BuildScheduleResource("sched-2", "Practitioner/dr-smith", "", "")

	if sched["resourceType"] != "Schedule" {
		t.Errorf("expected resourceType 'Schedule', got %v", sched["resourceType"])
	}
	if sched["id"] != "sched-2" {
		t.Errorf("expected id 'sched-2', got %v", sched["id"])
	}
}

// =========== BuildAvailabilityBundle Tests ===========

func TestBuildAvailabilityBundle_Single(t *testing.T) {
	result := &AvailabilityResult{
		Slots: []map[string]interface{}{
			{"resourceType": "Slot", "id": "slot-1", "status": "free"},
		},
		Total: 1,
	}

	bundle := BuildAvailabilityBundle(result, "http://example.com/fhir")

	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected resourceType 'Bundle', got %v", bundle["resourceType"])
	}
	if bundle["type"] != "searchset" {
		t.Errorf("expected type 'searchset', got %v", bundle["type"])
	}

	totalVal, ok := bundle["total"].(int)
	if !ok {
		t.Fatal("expected total to be int")
	}
	if totalVal != 1 {
		t.Errorf("expected total 1, got %d", totalVal)
	}

	entries, ok := bundle["entry"].([]interface{})
	if !ok || len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %v", bundle["entry"])
	}
}

func TestBuildAvailabilityBundle_Multiple(t *testing.T) {
	result := &AvailabilityResult{
		Slots: []map[string]interface{}{
			{"resourceType": "Slot", "id": "slot-1", "status": "free"},
			{"resourceType": "Slot", "id": "slot-2", "status": "free"},
		},
		Schedules: []map[string]interface{}{
			{"resourceType": "Schedule", "id": "sched-1"},
		},
		Total: 2,
	}

	bundle := BuildAvailabilityBundle(result, "http://example.com/fhir")

	entries, ok := bundle["entry"].([]interface{})
	if !ok {
		t.Fatal("expected entries array")
	}
	// 2 slots + 1 schedule = 3 entries
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestBuildAvailabilityBundle_Empty(t *testing.T) {
	result := &AvailabilityResult{
		Slots: []map[string]interface{}{},
		Total: 0,
	}

	bundle := BuildAvailabilityBundle(result, "http://example.com/fhir")

	totalVal, ok := bundle["total"].(int)
	if !ok {
		t.Fatal("expected total to be int")
	}
	if totalVal != 0 {
		t.Errorf("expected total 0, got %d", totalVal)
	}

	entries, ok := bundle["entry"].([]interface{})
	if !ok || len(entries) != 0 {
		t.Errorf("expected 0 entries, got %v", bundle["entry"])
	}
}

// =========== InMemoryAvailabilityStore Tests ===========

func TestInMemoryAvailabilityStore_AddSlot(t *testing.T) {
	store := NewInMemoryAvailabilityStore()
	slot := map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-1",
		"status":       "free",
		"start":        "2025-06-01T09:00:00Z",
		"end":          "2025-06-01T09:30:00Z",
	}

	store.AddSlot(slot)

	req := &AvailabilityRequest{
		Start: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 6, 1, 23, 59, 59, 0, time.UTC),
	}

	results, err := store.FindSlots(nil, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 slot, got %d", len(results))
	}
}

func TestInMemoryAvailabilityStore_AddSchedule(t *testing.T) {
	store := NewInMemoryAvailabilityStore()
	sched := map[string]interface{}{
		"resourceType": "Schedule",
		"id":           "sched-1",
	}

	store.AddSchedule(sched)

	result, err := store.GetSchedule(nil, "sched-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["id"] != "sched-1" {
		t.Errorf("expected schedule id 'sched-1', got %v", result["id"])
	}
}

func TestInMemoryAvailabilityStore_FindSlots_WithStatusFilter(t *testing.T) {
	store := NewInMemoryAvailabilityStore()
	store.AddSlot(map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-1",
		"status":       "free",
		"start":        "2025-06-01T09:00:00Z",
		"end":          "2025-06-01T09:30:00Z",
	})
	store.AddSlot(map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-2",
		"status":       "busy",
		"start":        "2025-06-01T10:00:00Z",
		"end":          "2025-06-01T10:30:00Z",
	})

	req := &AvailabilityRequest{
		Start:  time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		End:    time.Date(2025, 6, 1, 23, 59, 59, 0, time.UTC),
		Status: []SlotStatus{SlotStatusFree},
	}

	results, err := store.FindSlots(nil, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 free slot, got %d", len(results))
	}
	if len(results) > 0 && results[0]["id"] != "slot-1" {
		t.Errorf("expected slot-1, got %v", results[0]["id"])
	}
}

func TestInMemoryAvailabilityStore_FindSlots_WithPractitionerFilter(t *testing.T) {
	store := NewInMemoryAvailabilityStore()
	store.AddSlot(map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-1",
		"status":       "free",
		"start":        "2025-06-01T09:00:00Z",
		"end":          "2025-06-01T09:30:00Z",
		"schedule":     map[string]interface{}{"reference": "Schedule/sched-dr-jones"},
	})
	store.AddSchedule(map[string]interface{}{
		"resourceType": "Schedule",
		"id":           "sched-dr-jones",
		"actor": []interface{}{
			map[string]interface{}{"reference": "Practitioner/dr-jones"},
		},
	})
	store.AddSlot(map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-2",
		"status":       "free",
		"start":        "2025-06-01T10:00:00Z",
		"end":          "2025-06-01T10:30:00Z",
		"schedule":     map[string]interface{}{"reference": "Schedule/sched-dr-smith"},
	})
	store.AddSchedule(map[string]interface{}{
		"resourceType": "Schedule",
		"id":           "sched-dr-smith",
		"actor": []interface{}{
			map[string]interface{}{"reference": "Practitioner/dr-smith"},
		},
	})

	req := &AvailabilityRequest{
		Start:        time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		End:          time.Date(2025, 6, 1, 23, 59, 59, 0, time.UTC),
		Practitioner: "Practitioner/dr-jones",
	}

	results, err := store.FindSlots(nil, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 slot for dr-jones, got %d", len(results))
	}
}

func TestInMemoryAvailabilityStore_GetSchedule_NotFound(t *testing.T) {
	store := NewInMemoryAvailabilityStore()

	_, err := store.GetSchedule(nil, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent schedule")
	}
}

func TestInMemoryAvailabilityStore_CheckConflicts_NoConflict(t *testing.T) {
	store := NewInMemoryAvailabilityStore()
	store.AddSlot(map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-1",
		"status":       "busy",
		"start":        "2025-06-01T09:00:00Z",
		"end":          "2025-06-01T09:30:00Z",
		"schedule":     map[string]interface{}{"reference": "Schedule/sched-1"},
	})
	store.AddSchedule(map[string]interface{}{
		"resourceType": "Schedule",
		"id":           "sched-1",
		"actor":        []interface{}{map[string]interface{}{"reference": "Practitioner/dr-jones"}},
	})

	// Check a non-overlapping time.
	result, err := store.CheckConflicts(nil,
		time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC),
		time.Date(2025, 6, 1, 10, 30, 0, 0, time.UTC),
		"Practitioner/dr-jones",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasConflict {
		t.Error("expected no conflict")
	}
}

func TestInMemoryAvailabilityStore_CheckConflicts_HasConflict(t *testing.T) {
	store := NewInMemoryAvailabilityStore()
	store.AddSlot(map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-1",
		"status":       "busy",
		"start":        "2025-06-01T09:00:00Z",
		"end":          "2025-06-01T09:30:00Z",
		"schedule":     map[string]interface{}{"reference": "Schedule/sched-1"},
	})
	store.AddSchedule(map[string]interface{}{
		"resourceType": "Schedule",
		"id":           "sched-1",
		"actor":        []interface{}{map[string]interface{}{"reference": "Practitioner/dr-jones"}},
	})

	// Check an overlapping time.
	result, err := store.CheckConflicts(nil,
		time.Date(2025, 6, 1, 9, 15, 0, 0, time.UTC),
		time.Date(2025, 6, 1, 9, 45, 0, 0, time.UTC),
		"Practitioner/dr-jones",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasConflict {
		t.Error("expected conflict")
	}
	if len(result.Conflicts) == 0 {
		t.Error("expected at least one conflict entry")
	}
}

func TestInMemoryAvailabilityStore_CheckConflicts_DifferentActor(t *testing.T) {
	store := NewInMemoryAvailabilityStore()
	store.AddSlot(map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-1",
		"status":       "busy",
		"start":        "2025-06-01T09:00:00Z",
		"end":          "2025-06-01T09:30:00Z",
		"schedule":     map[string]interface{}{"reference": "Schedule/sched-1"},
	})
	store.AddSchedule(map[string]interface{}{
		"resourceType": "Schedule",
		"id":           "sched-1",
		"actor":        []interface{}{map[string]interface{}{"reference": "Practitioner/dr-jones"}},
	})

	// Different actor should have no conflict.
	result, err := store.CheckConflicts(nil,
		time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC),
		time.Date(2025, 6, 1, 9, 30, 0, 0, time.UTC),
		"Practitioner/dr-smith",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasConflict {
		t.Error("expected no conflict for different actor")
	}
}

// =========== SlotSearchHandler Tests ===========

func TestSlotSearchHandler_Success(t *testing.T) {
	store := NewInMemoryAvailabilityStore()
	store.AddSlot(map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-1",
		"status":       "free",
		"start":        "2025-06-01T09:00:00Z",
		"end":          "2025-06-01T09:30:00Z",
	})
	store.AddSlot(map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-2",
		"status":       "free",
		"start":        "2025-06-01T10:00:00Z",
		"end":          "2025-06-01T10:30:00Z",
	})

	handler := SlotSearchHandler(store)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet,
		"/fhir/Slot?start=2025-06-01T00:00:00Z&end=2025-06-01T23:59:59Z&status=free",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle, got %v", bundle["resourceType"])
	}
	if bundle["type"] != "searchset" {
		t.Errorf("expected searchset, got %v", bundle["type"])
	}

	totalVal := int(bundle["total"].(float64))
	if totalVal != 2 {
		t.Errorf("expected total 2, got %d", totalVal)
	}
}

func TestSlotSearchHandler_NoResults(t *testing.T) {
	store := NewInMemoryAvailabilityStore()

	handler := SlotSearchHandler(store)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet,
		"/fhir/Slot?start=2025-06-01T00:00:00Z&end=2025-06-01T23:59:59Z&status=free",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	totalVal := int(bundle["total"].(float64))
	if totalVal != 0 {
		t.Errorf("expected total 0, got %d", totalVal)
	}
}

func TestSlotSearchHandler_InvalidParams(t *testing.T) {
	store := NewInMemoryAvailabilityStore()

	handler := SlotSearchHandler(store)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet,
		"/fhir/Slot?start=bad-date&end=2025-06-01T23:59:59Z",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid params, got %d", rec.Code)
	}
}

// =========== ScheduleAvailableHandler Tests ===========

func TestScheduleAvailableHandler_Success(t *testing.T) {
	store := NewInMemoryAvailabilityStore()
	store.AddSchedule(map[string]interface{}{
		"resourceType": "Schedule",
		"id":           "sched-1",
		"actor": []interface{}{
			map[string]interface{}{"reference": "Practitioner/dr-jones"},
		},
	})
	store.AddSlot(map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-1",
		"status":       "free",
		"start":        "2025-06-01T09:00:00Z",
		"end":          "2025-06-01T09:30:00Z",
		"schedule":     map[string]interface{}{"reference": "Schedule/sched-1"},
	})

	handler := ScheduleAvailableHandler(store)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet,
		"/fhir/Schedule/sched-1/$available?start=2025-06-01T00:00:00Z&end=2025-06-01T23:59:59Z",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("sched-1")

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle, got %v", bundle["resourceType"])
	}
}

func TestScheduleAvailableHandler_NotFound(t *testing.T) {
	store := NewInMemoryAvailabilityStore()

	handler := ScheduleAvailableHandler(store)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet,
		"/fhir/Schedule/nonexistent/$available?start=2025-06-01T00:00:00Z&end=2025-06-01T23:59:59Z",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestScheduleAvailableHandler_NoAvailability(t *testing.T) {
	store := NewInMemoryAvailabilityStore()
	store.AddSchedule(map[string]interface{}{
		"resourceType": "Schedule",
		"id":           "sched-1",
	})

	handler := ScheduleAvailableHandler(store)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet,
		"/fhir/Schedule/sched-1/$available?start=2025-06-01T00:00:00Z&end=2025-06-01T23:59:59Z",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("sched-1")

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	totalVal := int(bundle["total"].(float64))
	if totalVal != 0 {
		t.Errorf("expected total 0, got %d", totalVal)
	}
}

// =========== FindAvailabilityHandler Tests ===========

func TestFindAvailabilityHandler_Success(t *testing.T) {
	store := NewInMemoryAvailabilityStore()
	store.AddSlot(map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-1",
		"status":       "free",
		"start":        "2025-06-01T09:00:00Z",
		"end":          "2025-06-01T09:30:00Z",
	})
	store.AddSlot(map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-2",
		"status":       "free",
		"start":        "2025-06-01T10:00:00Z",
		"end":          "2025-06-01T10:30:00Z",
	})

	handler := FindAvailabilityHandler(store)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet,
		"/fhir/Slot/$find?start=2025-06-01T00:00:00Z&end=2025-06-01T23:59:59Z",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle, got %v", bundle["resourceType"])
	}
}

func TestFindAvailabilityHandler_WithFilters(t *testing.T) {
	store := NewInMemoryAvailabilityStore()
	store.AddSlot(map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-1",
		"status":       "free",
		"start":        "2025-06-01T09:00:00Z",
		"end":          "2025-06-01T09:30:00Z",
		"serviceType": []interface{}{
			map[string]interface{}{"coding": []interface{}{map[string]interface{}{"code": "general-practice"}}},
		},
	})
	store.AddSlot(map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-2",
		"status":       "free",
		"start":        "2025-06-01T10:00:00Z",
		"end":          "2025-06-01T10:30:00Z",
		"serviceType": []interface{}{
			map[string]interface{}{"coding": []interface{}{map[string]interface{}{"code": "cardiology"}}},
		},
	})

	handler := FindAvailabilityHandler(store)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet,
		"/fhir/Slot/$find?start=2025-06-01T00:00:00Z&end=2025-06-01T23:59:59Z&service-type=general-practice",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	totalVal := int(bundle["total"].(float64))
	if totalVal != 1 {
		t.Errorf("expected total 1 with service-type filter, got %d", totalVal)
	}
}

func TestFindAvailabilityHandler_MissingRequiredParams(t *testing.T) {
	store := NewInMemoryAvailabilityStore()
	handler := FindAvailabilityHandler(store)
	e := echo.New()
	// Missing start and end.
	req := httptest.NewRequest(http.MethodGet, "/fhir/Slot/$find", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing params, got %d", rec.Code)
	}
}

// =========== CheckConflictHandler Tests ===========

func TestCheckConflictHandler_NoConflict(t *testing.T) {
	store := NewInMemoryAvailabilityStore()

	handler := CheckConflictHandler(store)
	e := echo.New()
	body := `{"start":"2025-06-01T09:00:00Z","end":"2025-06-01T09:30:00Z","actor":"Practitioner/dr-jones"}`
	req := httptest.NewRequest(http.MethodPost,
		"/fhir/Slot/$check-conflict",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result["resourceType"] != "Parameters" {
		t.Errorf("expected Parameters, got %v", result["resourceType"])
	}
}

func TestCheckConflictHandler_HasConflict(t *testing.T) {
	store := NewInMemoryAvailabilityStore()
	store.AddSlot(map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-1",
		"status":       "busy",
		"start":        "2025-06-01T09:00:00Z",
		"end":          "2025-06-01T09:30:00Z",
		"schedule":     map[string]interface{}{"reference": "Schedule/sched-1"},
	})
	store.AddSchedule(map[string]interface{}{
		"resourceType": "Schedule",
		"id":           "sched-1",
		"actor":        []interface{}{map[string]interface{}{"reference": "Practitioner/dr-jones"}},
	})

	handler := CheckConflictHandler(store)
	e := echo.New()
	body := `{"start":"2025-06-01T09:15:00Z","end":"2025-06-01T09:45:00Z","actor":"Practitioner/dr-jones"}`
	req := httptest.NewRequest(http.MethodPost,
		"/fhir/Slot/$check-conflict",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	params := result["parameter"].([]interface{})
	// Find the hasConflict parameter.
	var hasConflict bool
	for _, p := range params {
		pm := p.(map[string]interface{})
		if pm["name"] == "hasConflict" {
			hasConflict = pm["valueBoolean"].(bool)
		}
	}
	if !hasConflict {
		t.Error("expected hasConflict to be true")
	}
}

func TestCheckConflictHandler_MissingParams(t *testing.T) {
	store := NewInMemoryAvailabilityStore()

	handler := CheckConflictHandler(store)
	e := echo.New()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost,
		"/fhir/Slot/$check-conflict",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing params, got %d", rec.Code)
	}
}

func TestCheckConflictHandler_InvalidJSON(t *testing.T) {
	store := NewInMemoryAvailabilityStore()

	handler := CheckConflictHandler(store)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost,
		"/fhir/Slot/$check-conflict",
		strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

// =========== DefaultAvailabilityRules Tests ===========

func TestDefaultAvailabilityRules(t *testing.T) {
	rules := DefaultAvailabilityRules()
	if len(rules) == 0 {
		t.Fatal("expected at least one default rule")
	}

	for i, rule := range rules {
		if len(rule.DaysOfWeek) == 0 {
			t.Errorf("rule[%d]: expected at least one day of week", i)
		}
		if rule.StartTime == "" {
			t.Errorf("rule[%d]: expected non-empty start time", i)
		}
		if rule.EndTime == "" {
			t.Errorf("rule[%d]: expected non-empty end time", i)
		}
		if rule.SlotDuration <= 0 {
			t.Errorf("rule[%d]: expected positive slot duration", i)
		}

		// Validate time format.
		_, _, err := ParseTimeOfDay(rule.StartTime)
		if err != nil {
			t.Errorf("rule[%d]: invalid start time %q: %v", i, rule.StartTime, err)
		}
		_, _, err = ParseTimeOfDay(rule.EndTime)
		if err != nil {
			t.Errorf("rule[%d]: invalid end time %q: %v", i, rule.EndTime, err)
		}
	}
}

// =========== Edge Case Tests ===========

func TestBuildSlotResource_ZeroLengthSlot(t *testing.T) {
	now := time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC)
	avail := &ScheduleAvailability{
		Start:  now,
		End:    now,
		Status: "free",
	}

	slot := BuildSlotResource(avail, "Schedule/s1")
	if slot["start"] != slot["end"] {
		t.Errorf("expected start == end for zero-length slot")
	}
}

func TestGenerateTimeSlots_LongRange(t *testing.T) {
	rule := &AvailabilityRule{
		DaysOfWeek:   []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		StartTime:    "09:00",
		EndTime:      "17:00",
		SlotDuration: 30,
	}

	// One week range.
	start := time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 8, 23, 59, 59, 0, time.UTC)

	slots := GenerateTimeSlots(rule, start, end)
	// 16 slots/day * 5 weekdays = 80 slots
	if len(slots) != 80 {
		t.Errorf("expected 80 slots for a week, got %d", len(slots))
	}
}

func TestOverlapsTimeRange_ZeroLength(t *testing.T) {
	// Zero-length range (point in time).
	point := time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC)
	s2 := time.Date(2025, 6, 1, 8, 0, 0, 0, time.UTC)
	e2 := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)

	// A zero-length range at 9:00 within 8:00-10:00 should not be considered an overlap
	// since it has no actual duration.
	result := OverlapsTimeRange(point, point, s2, e2)
	if result {
		t.Error("zero-length range should not overlap")
	}
}

func TestSlotStatusConstants(t *testing.T) {
	// Verify all FHIR Slot status constants are correct.
	if SlotStatusFree != "free" {
		t.Errorf("expected 'free', got %q", SlotStatusFree)
	}
	if SlotStatusBusy != "busy" {
		t.Errorf("expected 'busy', got %q", SlotStatusBusy)
	}
	if SlotStatusBusyUnavailable != "busy-unavailable" {
		t.Errorf("expected 'busy-unavailable', got %q", SlotStatusBusyUnavailable)
	}
	if SlotStatusBusyTentative != "busy-tentative" {
		t.Errorf("expected 'busy-tentative', got %q", SlotStatusBusyTentative)
	}
	if SlotStatusEnteredInError != "entered-in-error" {
		t.Errorf("expected 'entered-in-error', got %q", SlotStatusEnteredInError)
	}
}

func TestInMemoryAvailabilityStore_ConcurrentAccess(t *testing.T) {
	store := NewInMemoryAvailabilityStore()
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			store.AddSlot(map[string]interface{}{
				"resourceType": "Slot",
				"id":           "slot-" + time.Now().Format("150405.000000000"),
				"status":       "free",
				"start":        "2025-06-01T09:00:00Z",
				"end":          "2025-06-01T09:30:00Z",
			})
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	req := &AvailabilityRequest{
		Start: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 6, 1, 23, 59, 59, 0, time.UTC),
	}

	results, err := store.FindSlots(nil, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 10 {
		t.Errorf("expected 10 slots after concurrent adds, got %d", len(results))
	}
}

func TestBuildAvailabilityBundle_WithScheduleIncludes(t *testing.T) {
	result := &AvailabilityResult{
		Slots: []map[string]interface{}{
			{"resourceType": "Slot", "id": "slot-1", "status": "free"},
		},
		Schedules: []map[string]interface{}{
			{"resourceType": "Schedule", "id": "sched-1"},
		},
		Total: 1,
	}

	bundle := BuildAvailabilityBundle(result, "http://example.com/fhir")

	entries := bundle["entry"].([]interface{})
	if len(entries) != 2 {
		t.Errorf("expected 2 entries (1 slot + 1 schedule), got %d", len(entries))
	}

	// Verify the schedule entry has search mode "include".
	schedEntry := entries[1].(map[string]interface{})
	search := schedEntry["search"].(map[string]interface{})
	if search["mode"] != "include" {
		t.Errorf("expected schedule search mode 'include', got %v", search["mode"])
	}
}

func TestScheduleAvailableHandler_MissingScheduleID(t *testing.T) {
	store := NewInMemoryAvailabilityStore()
	handler := ScheduleAvailableHandler(store)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet,
		"/fhir/Schedule//$available?start=2025-06-01T00:00:00Z&end=2025-06-01T23:59:59Z",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("")

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty schedule ID, got %d", rec.Code)
	}
}

func TestInMemoryAvailabilityStore_FindSlots_DateRange(t *testing.T) {
	store := NewInMemoryAvailabilityStore()
	store.AddSlot(map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-1",
		"status":       "free",
		"start":        "2025-06-01T09:00:00Z",
		"end":          "2025-06-01T09:30:00Z",
	})
	store.AddSlot(map[string]interface{}{
		"resourceType": "Slot",
		"id":           "slot-2",
		"status":       "free",
		"start":        "2025-06-02T09:00:00Z",
		"end":          "2025-06-02T09:30:00Z",
	})

	req := &AvailabilityRequest{
		Start: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 6, 1, 23, 59, 59, 0, time.UTC),
	}

	results, err := store.FindSlots(nil, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 slot in date range, got %d", len(results))
	}
}

func TestParseAvailabilityRequest_DateOnly(t *testing.T) {
	params := url.Values{}
	params.Set("start", "2025-06-01")
	params.Set("end", "2025-06-02")

	req, err := ParseAvailabilityRequest(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStart := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	if !req.Start.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, req.Start)
	}
}
