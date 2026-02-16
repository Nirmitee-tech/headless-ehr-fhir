package scheduling

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Schedule maps to the schedule table (FHIR Schedule resource).
type Schedule struct {
	ID                   uuid.UUID  `db:"id" json:"id"`
	FHIRID               string     `db:"fhir_id" json:"fhir_id"`
	Active               *bool      `db:"active" json:"active,omitempty"`
	PractitionerID       uuid.UUID  `db:"practitioner_id" json:"practitioner_id"`
	LocationID           *uuid.UUID `db:"location_id" json:"location_id,omitempty"`
	ServiceTypeCode      *string    `db:"service_type_code" json:"service_type_code,omitempty"`
	ServiceTypeDisplay   *string    `db:"service_type_display" json:"service_type_display,omitempty"`
	SpecialtyCode        *string    `db:"specialty_code" json:"specialty_code,omitempty"`
	SpecialtyDisplay     *string    `db:"specialty_display" json:"specialty_display,omitempty"`
	PlanningHorizonStart *time.Time `db:"planning_horizon_start" json:"planning_horizon_start,omitempty"`
	PlanningHorizonEnd   *time.Time `db:"planning_horizon_end" json:"planning_horizon_end,omitempty"`
	Comment              *string    `db:"comment" json:"comment,omitempty"`
	VersionID            int        `db:"version_id" json:"version_id"`
	CreatedAt            time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (s *Schedule) GetVersionID() int { return s.VersionID }

// SetVersionID sets the current version.
func (s *Schedule) SetVersionID(v int) { s.VersionID = v }

func (s *Schedule) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Schedule",
		"id":           s.FHIRID,
		"actor": []fhir.Reference{{
			Reference: fhir.FormatReference("Practitioner", s.PractitionerID.String()),
		}},
		"meta": fhir.Meta{LastUpdated: s.UpdatedAt},
	}
	if s.Active != nil {
		result["active"] = *s.Active
	}
	if s.ServiceTypeCode != nil {
		result["serviceType"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *s.ServiceTypeCode, Display: strVal(s.ServiceTypeDisplay)}},
		}}
	}
	if s.SpecialtyCode != nil {
		result["specialty"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *s.SpecialtyCode, Display: strVal(s.SpecialtyDisplay)}},
		}}
	}
	if s.PlanningHorizonStart != nil || s.PlanningHorizonEnd != nil {
		result["planningHorizon"] = fhir.Period{Start: s.PlanningHorizonStart, End: s.PlanningHorizonEnd}
	}
	if s.LocationID != nil {
		result["actor"] = append(result["actor"].([]fhir.Reference), fhir.Reference{
			Reference: fhir.FormatReference("Location", s.LocationID.String()),
		})
	}
	if s.Comment != nil {
		result["comment"] = *s.Comment
	}
	return result
}

// Slot maps to the slot table (FHIR Slot resource).
type Slot struct {
	ID                     uuid.UUID  `db:"id" json:"id"`
	FHIRID                 string     `db:"fhir_id" json:"fhir_id"`
	ScheduleID             uuid.UUID  `db:"schedule_id" json:"schedule_id"`
	Status                 string     `db:"status" json:"status"`
	StartTime              time.Time  `db:"start_time" json:"start_time"`
	EndTime                time.Time  `db:"end_time" json:"end_time"`
	Overbooked             *bool      `db:"overbooked" json:"overbooked,omitempty"`
	Comment                *string    `db:"comment" json:"comment,omitempty"`
	ServiceTypeCode        *string    `db:"service_type_code" json:"service_type_code,omitempty"`
	ServiceTypeDisplay     *string    `db:"service_type_display" json:"service_type_display,omitempty"`
	SpecialtyCode          *string    `db:"specialty_code" json:"specialty_code,omitempty"`
	SpecialtyDisplay       *string    `db:"specialty_display" json:"specialty_display,omitempty"`
	AppointmentTypeCode    *string    `db:"appointment_type_code" json:"appointment_type_code,omitempty"`
	AppointmentTypeDisplay *string    `db:"appointment_type_display" json:"appointment_type_display,omitempty"`
	VersionID              int        `db:"version_id" json:"version_id"`
	CreatedAt              time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (sl *Slot) GetVersionID() int { return sl.VersionID }

// SetVersionID sets the current version.
func (sl *Slot) SetVersionID(v int) { sl.VersionID = v }

func (sl *Slot) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Slot",
		"id":           sl.FHIRID,
		"schedule":     fhir.Reference{Reference: fhir.FormatReference("Schedule", sl.ScheduleID.String())},
		"status":       sl.Status,
		"start":        sl.StartTime.Format(time.RFC3339),
		"end":          sl.EndTime.Format(time.RFC3339),
		"meta":         fhir.Meta{LastUpdated: sl.UpdatedAt},
	}
	if sl.Overbooked != nil {
		result["overbooked"] = *sl.Overbooked
	}
	if sl.ServiceTypeCode != nil {
		result["serviceType"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *sl.ServiceTypeCode, Display: strVal(sl.ServiceTypeDisplay)}},
		}}
	}
	if sl.SpecialtyCode != nil {
		result["specialty"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *sl.SpecialtyCode, Display: strVal(sl.SpecialtyDisplay)}},
		}}
	}
	if sl.AppointmentTypeCode != nil {
		result["appointmentType"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *sl.AppointmentTypeCode, Display: strVal(sl.AppointmentTypeDisplay)}},
		}
	}
	if sl.Comment != nil {
		result["comment"] = *sl.Comment
	}
	return result
}

// Appointment maps to the appointment table (FHIR Appointment resource).
type Appointment struct {
	ID                     uuid.UUID  `db:"id" json:"id"`
	FHIRID                 string     `db:"fhir_id" json:"fhir_id"`
	Status                 string     `db:"status" json:"status"`
	CancellationReason     *string    `db:"cancellation_reason" json:"cancellation_reason,omitempty"`
	ServiceTypeCode        *string    `db:"service_type_code" json:"service_type_code,omitempty"`
	ServiceTypeDisplay     *string    `db:"service_type_display" json:"service_type_display,omitempty"`
	SpecialtyCode          *string    `db:"specialty_code" json:"specialty_code,omitempty"`
	SpecialtyDisplay       *string    `db:"specialty_display" json:"specialty_display,omitempty"`
	AppointmentTypeCode    *string    `db:"appointment_type_code" json:"appointment_type_code,omitempty"`
	AppointmentTypeDisplay *string    `db:"appointment_type_display" json:"appointment_type_display,omitempty"`
	Priority               *int       `db:"priority" json:"priority,omitempty"`
	Description            *string    `db:"description" json:"description,omitempty"`
	StartTime              *time.Time `db:"start_time" json:"start_time,omitempty"`
	EndTime                *time.Time `db:"end_time" json:"end_time,omitempty"`
	MinutesDuration        *int       `db:"minutes_duration" json:"minutes_duration,omitempty"`
	SlotID                 *uuid.UUID `db:"slot_id" json:"slot_id,omitempty"`
	PatientID              uuid.UUID  `db:"patient_id" json:"patient_id"`
	PractitionerID         *uuid.UUID `db:"practitioner_id" json:"practitioner_id,omitempty"`
	LocationID             *uuid.UUID `db:"location_id" json:"location_id,omitempty"`
	ReasonCode             *string    `db:"reason_code" json:"reason_code,omitempty"`
	ReasonDisplay          *string    `db:"reason_display" json:"reason_display,omitempty"`
	ReasonConditionID      *uuid.UUID `db:"reason_condition_id" json:"reason_condition_id,omitempty"`
	Note                   *string    `db:"note" json:"note,omitempty"`
	PatientInstruction     *string    `db:"patient_instruction" json:"patient_instruction,omitempty"`
	IsTelehealth           *bool      `db:"is_telehealth" json:"is_telehealth,omitempty"`
	TelehealthURL          *string    `db:"telehealth_url" json:"telehealth_url,omitempty"`
	VersionID              int        `db:"version_id" json:"version_id"`
	CreatedAt              time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (a *Appointment) GetVersionID() int { return a.VersionID }

// SetVersionID sets the current version.
func (a *Appointment) SetVersionID(v int) { a.VersionID = v }

func (a *Appointment) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Appointment",
		"id":           a.FHIRID,
		"status":       a.Status,
		"meta":         fhir.Meta{LastUpdated: a.UpdatedAt},
	}
	if a.ServiceTypeCode != nil {
		result["serviceType"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *a.ServiceTypeCode, Display: strVal(a.ServiceTypeDisplay)}},
		}}
	}
	if a.SpecialtyCode != nil {
		result["specialty"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *a.SpecialtyCode, Display: strVal(a.SpecialtyDisplay)}},
		}}
	}
	if a.AppointmentTypeCode != nil {
		result["appointmentType"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *a.AppointmentTypeCode, Display: strVal(a.AppointmentTypeDisplay)}},
		}
	}
	if a.Priority != nil {
		result["priority"] = *a.Priority
	}
	if a.Description != nil {
		result["description"] = *a.Description
	}
	if a.StartTime != nil {
		result["start"] = a.StartTime.Format(time.RFC3339)
	}
	if a.EndTime != nil {
		result["end"] = a.EndTime.Format(time.RFC3339)
	}
	if a.MinutesDuration != nil {
		result["minutesDuration"] = *a.MinutesDuration
	}
	if a.SlotID != nil {
		result["slot"] = []fhir.Reference{{Reference: fhir.FormatReference("Slot", a.SlotID.String())}}
	}
	if a.ReasonCode != nil {
		result["reasonCode"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *a.ReasonCode, Display: strVal(a.ReasonDisplay)}},
		}}
	}
	if a.CancellationReason != nil {
		result["cancelationReason"] = fhir.CodeableConcept{
			Text: *a.CancellationReason,
		}
	}

	// Build participants array
	participants := []map[string]interface{}{}
	participants = append(participants, map[string]interface{}{
		"actor":  fhir.Reference{Reference: fhir.FormatReference("Patient", a.PatientID.String())},
		"status": "accepted",
	})
	if a.PractitionerID != nil {
		participants = append(participants, map[string]interface{}{
			"actor":  fhir.Reference{Reference: fhir.FormatReference("Practitioner", a.PractitionerID.String())},
			"status": "accepted",
		})
	}
	if a.LocationID != nil {
		participants = append(participants, map[string]interface{}{
			"actor":  fhir.Reference{Reference: fhir.FormatReference("Location", a.LocationID.String())},
			"status": "accepted",
		})
	}
	result["participant"] = participants

	if a.Note != nil {
		result["comment"] = *a.Note
	}
	if a.PatientInstruction != nil {
		result["patientInstruction"] = *a.PatientInstruction
	}
	return result
}

// AppointmentParticipant maps to the appointment_participant table.
type AppointmentParticipant struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	AppointmentID uuid.UUID  `db:"appointment_id" json:"appointment_id"`
	ActorType     string     `db:"actor_type" json:"actor_type"`
	ActorID       uuid.UUID  `db:"actor_id" json:"actor_id"`
	RoleCode      *string    `db:"role_code" json:"role_code,omitempty"`
	RoleDisplay   *string    `db:"role_display" json:"role_display,omitempty"`
	Status        string     `db:"status" json:"status"`
	Required      *string    `db:"required" json:"required,omitempty"`
	PeriodStart   *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd     *time.Time `db:"period_end" json:"period_end,omitempty"`
}

// AppointmentResponse maps to the appointment_response table (FHIR AppointmentResponse resource).
type AppointmentResponse struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	FHIRID            string     `db:"fhir_id" json:"fhir_id"`
	AppointmentID     uuid.UUID  `db:"appointment_id" json:"appointment_id"`
	ActorType         string     `db:"actor_type" json:"actor_type"`
	ActorID           uuid.UUID  `db:"actor_id" json:"actor_id"`
	ParticipantStatus string     `db:"participant_status" json:"participant_status"`
	Comment           *string    `db:"comment" json:"comment,omitempty"`
	StartTime         *time.Time `db:"start_time" json:"start_time,omitempty"`
	EndTime           *time.Time `db:"end_time" json:"end_time,omitempty"`
	VersionID         int        `db:"version_id" json:"version_id"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (ar *AppointmentResponse) GetVersionID() int { return ar.VersionID }

// SetVersionID sets the current version.
func (ar *AppointmentResponse) SetVersionID(v int) { ar.VersionID = v }

func (ar *AppointmentResponse) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType":      "AppointmentResponse",
		"id":                ar.FHIRID,
		"appointment":       fhir.Reference{Reference: fhir.FormatReference("Appointment", ar.AppointmentID.String())},
		"actor":             fhir.Reference{Reference: fhir.FormatReference(ar.ActorType, ar.ActorID.String())},
		"participantStatus": ar.ParticipantStatus,
		"meta":              fhir.Meta{LastUpdated: ar.UpdatedAt},
	}
	if ar.Comment != nil {
		result["comment"] = *ar.Comment
	}
	if ar.StartTime != nil {
		result["start"] = ar.StartTime.Format(time.RFC3339)
	}
	if ar.EndTime != nil {
		result["end"] = ar.EndTime.Format(time.RFC3339)
	}
	return result
}

// Waitlist maps to the waitlist table.
type Waitlist struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	PatientID          uuid.UUID  `db:"patient_id" json:"patient_id"`
	PractitionerID     *uuid.UUID `db:"practitioner_id" json:"practitioner_id,omitempty"`
	Department         *string    `db:"department" json:"department,omitempty"`
	ServiceTypeCode    *string    `db:"service_type_code" json:"service_type_code,omitempty"`
	ServiceTypeDisplay *string    `db:"service_type_display" json:"service_type_display,omitempty"`
	Priority           *int       `db:"priority" json:"priority,omitempty"`
	QueueNumber        *int       `db:"queue_number" json:"queue_number,omitempty"`
	Status             string     `db:"status" json:"status"`
	RequestedDate      *time.Time `db:"requested_date" json:"requested_date,omitempty"`
	CheckInTime        *time.Time `db:"check_in_time" json:"check_in_time,omitempty"`
	CalledTime         *time.Time `db:"called_time" json:"called_time,omitempty"`
	CompletedTime      *time.Time `db:"completed_time" json:"completed_time,omitempty"`
	Note               *string    `db:"note" json:"note,omitempty"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
