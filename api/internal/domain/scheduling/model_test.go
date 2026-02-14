package scheduling

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
// Schedule.ToFHIR
// ---------------------------------------------------------------------------

func TestScheduleToFHIR_RequiredFields(t *testing.T) {
	practID := uuid.New()
	now := time.Now()

	s := Schedule{
		ID:             uuid.New(),
		FHIRID:         "sched-100",
		PractitionerID: practID,
		UpdatedAt:      now,
	}

	result := s.ToFHIR()

	// resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("expected resourceType to be present")
	} else if rt != "Schedule" {
		t.Errorf("resourceType = %v, want Schedule", rt)
	}

	// id
	if id, ok := result["id"]; !ok {
		t.Error("expected id to be present")
	} else if id != "sched-100" {
		t.Errorf("id = %v, want sched-100", id)
	}

	// actor
	if _, ok := result["actor"]; !ok {
		t.Error("expected actor to be present")
	}

	// meta
	if _, ok := result["meta"]; !ok {
		t.Error("expected meta to be present")
	}

	// optional fields must be absent
	for _, key := range []string{
		"active", "serviceType", "specialty", "planningHorizon", "comment",
	} {
		if _, ok := result[key]; ok {
			t.Errorf("expected %s to be absent when not set", key)
		}
	}
}

func TestScheduleToFHIR_WithOptionalFields(t *testing.T) {
	practID := uuid.New()
	locID := uuid.New()
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	now := time.Now()

	s := Schedule{
		ID:                   uuid.New(),
		FHIRID:               "sched-200",
		PractitionerID:       practID,
		Active:               ptrBool(true),
		LocationID:           ptrUUID(locID),
		ServiceTypeCode:      ptrStr("general"),
		ServiceTypeDisplay:   ptrStr("General Practice"),
		SpecialtyCode:        ptrStr("cardiology"),
		SpecialtyDisplay:     ptrStr("Cardiology"),
		PlanningHorizonStart: ptrTime(start),
		PlanningHorizonEnd:   ptrTime(end),
		Comment:              ptrStr("Mon-Fri only"),
		UpdatedAt:            now,
	}

	result := s.ToFHIR()

	for _, key := range []string{
		"active", "serviceType", "specialty", "planningHorizon", "comment",
	} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected %s to be present", key)
		}
	}

	// active should be true
	if a, ok := result["active"]; ok && a != true {
		t.Errorf("active = %v, want true", a)
	}

	// comment value
	if c, ok := result["comment"]; ok && c != "Mon-Fri only" {
		t.Errorf("comment = %v, want 'Mon-Fri only'", c)
	}
}

func TestScheduleToFHIR_LocationAppendsToActor(t *testing.T) {
	practID := uuid.New()
	locID := uuid.New()
	now := time.Now()

	s := Schedule{
		ID:             uuid.New(),
		FHIRID:         "sched-300",
		PractitionerID: practID,
		LocationID:     ptrUUID(locID),
		UpdatedAt:      now,
	}

	result := s.ToFHIR()

	// When LocationID is set, actor should have 2 entries (practitioner + location)
	if _, ok := result["actor"]; !ok {
		t.Fatal("expected actor to be present")
	}
}

// ---------------------------------------------------------------------------
// Slot.ToFHIR
// ---------------------------------------------------------------------------

func TestSlotToFHIR_RequiredFields(t *testing.T) {
	schedID := uuid.New()
	start := time.Date(2025, 6, 15, 9, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 15, 9, 30, 0, 0, time.UTC)
	now := time.Now()

	sl := Slot{
		ID:         uuid.New(),
		FHIRID:     "slot-100",
		ScheduleID: schedID,
		Status:     "free",
		StartTime:  start,
		EndTime:    end,
		UpdatedAt:  now,
	}

	result := sl.ToFHIR()

	// resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("expected resourceType to be present")
	} else if rt != "Slot" {
		t.Errorf("resourceType = %v, want Slot", rt)
	}

	// id
	if id, ok := result["id"]; !ok {
		t.Error("expected id to be present")
	} else if id != "slot-100" {
		t.Errorf("id = %v, want slot-100", id)
	}

	// schedule
	if _, ok := result["schedule"]; !ok {
		t.Error("expected schedule to be present")
	}

	// status
	if s, ok := result["status"]; !ok {
		t.Error("expected status to be present")
	} else if s != "free" {
		t.Errorf("status = %v, want free", s)
	}

	// start
	if st, ok := result["start"]; !ok {
		t.Error("expected start to be present")
	} else if st != start.Format(time.RFC3339) {
		t.Errorf("start = %v, want %v", st, start.Format(time.RFC3339))
	}

	// end
	if e, ok := result["end"]; !ok {
		t.Error("expected end to be present")
	} else if e != end.Format(time.RFC3339) {
		t.Errorf("end = %v, want %v", e, end.Format(time.RFC3339))
	}

	// meta
	if _, ok := result["meta"]; !ok {
		t.Error("expected meta to be present")
	}

	// optional fields must be absent
	for _, key := range []string{
		"overbooked", "serviceType", "specialty", "appointmentType", "comment",
	} {
		if _, ok := result[key]; ok {
			t.Errorf("expected %s to be absent when not set", key)
		}
	}
}

func TestSlotToFHIR_WithOptionalFields(t *testing.T) {
	schedID := uuid.New()
	start := time.Date(2025, 6, 15, 9, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 15, 9, 30, 0, 0, time.UTC)
	now := time.Now()

	sl := Slot{
		ID:                     uuid.New(),
		FHIRID:                 "slot-200",
		ScheduleID:             schedID,
		Status:                 "busy",
		StartTime:              start,
		EndTime:                end,
		Overbooked:             ptrBool(true),
		ServiceTypeCode:        ptrStr("consult"),
		ServiceTypeDisplay:     ptrStr("Consultation"),
		SpecialtyCode:          ptrStr("cardiology"),
		SpecialtyDisplay:       ptrStr("Cardiology"),
		AppointmentTypeCode:    ptrStr("followup"),
		AppointmentTypeDisplay: ptrStr("Follow-up"),
		Comment:                ptrStr("Double booked"),
		UpdatedAt:              now,
	}

	result := sl.ToFHIR()

	for _, key := range []string{
		"overbooked", "serviceType", "specialty", "appointmentType", "comment",
	} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected %s to be present", key)
		}
	}

	// Check overbooked value
	if ob, ok := result["overbooked"]; ok && ob != true {
		t.Errorf("overbooked = %v, want true", ob)
	}

	// Check comment
	if c, ok := result["comment"]; ok && c != "Double booked" {
		t.Errorf("comment = %v, want 'Double booked'", c)
	}
}

// ---------------------------------------------------------------------------
// Appointment.ToFHIR
// ---------------------------------------------------------------------------

func TestAppointmentToFHIR_RequiredFields(t *testing.T) {
	patID := uuid.New()
	now := time.Now()

	a := Appointment{
		ID:        uuid.New(),
		FHIRID:    "appt-100",
		Status:    "booked",
		PatientID: patID,
		UpdatedAt: now,
	}

	result := a.ToFHIR()

	// resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("expected resourceType to be present")
	} else if rt != "Appointment" {
		t.Errorf("resourceType = %v, want Appointment", rt)
	}

	// id
	if id, ok := result["id"]; !ok {
		t.Error("expected id to be present")
	} else if id != "appt-100" {
		t.Errorf("id = %v, want appt-100", id)
	}

	// status
	if s, ok := result["status"]; !ok {
		t.Error("expected status to be present")
	} else if s != "booked" {
		t.Errorf("status = %v, want booked", s)
	}

	// meta
	if _, ok := result["meta"]; !ok {
		t.Error("expected meta to be present")
	}

	// participant must always be present (patient is always added)
	if _, ok := result["participant"]; !ok {
		t.Error("expected participant to be present")
	}

	// optional fields must be absent
	for _, key := range []string{
		"serviceType", "specialty", "appointmentType",
		"priority", "description", "start", "end",
		"minutesDuration", "slot", "reasonCode",
		"cancelationReason", "comment", "patientInstruction",
	} {
		if _, ok := result[key]; ok {
			t.Errorf("expected %s to be absent when not set", key)
		}
	}
}

func TestAppointmentToFHIR_WithOptionalFields(t *testing.T) {
	patID := uuid.New()
	practID := uuid.New()
	locID := uuid.New()
	slotID := uuid.New()
	startTime := time.Date(2025, 7, 20, 14, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 7, 20, 14, 30, 0, 0, time.UTC)
	now := time.Now()

	a := Appointment{
		ID:                     uuid.New(),
		FHIRID:                 "appt-200",
		Status:                 "booked",
		PatientID:              patID,
		PractitionerID:         ptrUUID(practID),
		LocationID:             ptrUUID(locID),
		ServiceTypeCode:        ptrStr("consult"),
		ServiceTypeDisplay:     ptrStr("Consultation"),
		SpecialtyCode:          ptrStr("cardiology"),
		SpecialtyDisplay:       ptrStr("Cardiology"),
		AppointmentTypeCode:    ptrStr("routine"),
		AppointmentTypeDisplay: ptrStr("Routine"),
		Priority:               ptrInt(5),
		Description:            ptrStr("Follow-up consultation"),
		StartTime:              ptrTime(startTime),
		EndTime:                ptrTime(endTime),
		MinutesDuration:        ptrInt(30),
		SlotID:                 ptrUUID(slotID),
		ReasonCode:             ptrStr("check-up"),
		ReasonDisplay:          ptrStr("Annual Check-up"),
		CancellationReason:     ptrStr("patient requested"),
		Note:                   ptrStr("Patient prefers afternoon"),
		PatientInstruction:     ptrStr("Bring previous records"),
		UpdatedAt:              now,
	}

	result := a.ToFHIR()

	for _, key := range []string{
		"serviceType", "specialty", "appointmentType",
		"priority", "description", "start", "end",
		"minutesDuration", "slot", "reasonCode",
		"cancelationReason", "comment", "patientInstruction",
	} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected %s to be present", key)
		}
	}

	// participant should contain patient + practitioner + location = 3 entries
	if _, ok := result["participant"]; !ok {
		t.Fatal("expected participant to be present")
	}

	// Check specific values
	if p, ok := result["priority"]; ok && p != 5 {
		t.Errorf("priority = %v, want 5", p)
	}
	if d, ok := result["description"]; ok && d != "Follow-up consultation" {
		t.Errorf("description = %v, want 'Follow-up consultation'", d)
	}
	if md, ok := result["minutesDuration"]; ok && md != 30 {
		t.Errorf("minutesDuration = %v, want 30", md)
	}
	if st, ok := result["start"]; ok {
		if st != startTime.Format(time.RFC3339) {
			t.Errorf("start = %v, want %v", st, startTime.Format(time.RFC3339))
		}
	}
	if et, ok := result["end"]; ok {
		if et != endTime.Format(time.RFC3339) {
			t.Errorf("end = %v, want %v", et, endTime.Format(time.RFC3339))
		}
	}
	if pi, ok := result["patientInstruction"]; ok && pi != "Bring previous records" {
		t.Errorf("patientInstruction = %v, want 'Bring previous records'", pi)
	}
}

func TestAppointmentToFHIR_ParticipantOnlyPatient(t *testing.T) {
	patID := uuid.New()
	now := time.Now()

	a := Appointment{
		ID:        uuid.New(),
		FHIRID:    "appt-300",
		Status:    "proposed",
		PatientID: patID,
		UpdatedAt: now,
	}

	result := a.ToFHIR()

	participants, ok := result["participant"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected participant to be []map[string]interface{}")
	}
	if len(participants) != 1 {
		t.Errorf("participant count = %d, want 1 (patient only)", len(participants))
	}
}

func TestAppointmentToFHIR_ParticipantWithPractitioner(t *testing.T) {
	patID := uuid.New()
	practID := uuid.New()
	now := time.Now()

	a := Appointment{
		ID:             uuid.New(),
		FHIRID:         "appt-400",
		Status:         "booked",
		PatientID:      patID,
		PractitionerID: ptrUUID(practID),
		UpdatedAt:      now,
	}

	result := a.ToFHIR()

	participants, ok := result["participant"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected participant to be []map[string]interface{}")
	}
	if len(participants) != 2 {
		t.Errorf("participant count = %d, want 2 (patient + practitioner)", len(participants))
	}
}
