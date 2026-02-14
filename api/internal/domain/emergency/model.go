package emergency

import (
	"time"

	"github.com/google/uuid"
)

// TriageRecord maps to the triage_record table.
type TriageRecord struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	PatientID         uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID       uuid.UUID  `db:"encounter_id" json:"encounter_id"`
	TriageNurseID     uuid.UUID  `db:"triage_nurse_id" json:"triage_nurse_id"`
	ArrivalTime       *time.Time `db:"arrival_time" json:"arrival_time,omitempty"`
	TriageTime        *time.Time `db:"triage_time" json:"triage_time,omitempty"`
	ChiefComplaint    string     `db:"chief_complaint" json:"chief_complaint"`
	AcuityLevel       *int       `db:"acuity_level" json:"acuity_level,omitempty"`
	AcuitySystem      *string    `db:"acuity_system" json:"acuity_system,omitempty"`
	PainScale         *int       `db:"pain_scale" json:"pain_scale,omitempty"`
	ArrivalMode       *string    `db:"arrival_mode" json:"arrival_mode,omitempty"`
	HeartRate         *int       `db:"heart_rate" json:"heart_rate,omitempty"`
	BloodPressureSys  *int       `db:"blood_pressure_sys" json:"blood_pressure_sys,omitempty"`
	BloodPressureDia  *int       `db:"blood_pressure_dia" json:"blood_pressure_dia,omitempty"`
	Temperature       *float64   `db:"temperature" json:"temperature,omitempty"`
	RespiratoryRate   *int       `db:"respiratory_rate" json:"respiratory_rate,omitempty"`
	OxygenSaturation  *int       `db:"oxygen_saturation" json:"oxygen_saturation,omitempty"`
	GlasgowComaScore  *int       `db:"glasgow_coma_score" json:"glasgow_coma_score,omitempty"`
	InjuryDescription *string    `db:"injury_description" json:"injury_description,omitempty"`
	AllergyNote       *string    `db:"allergy_note" json:"allergy_note,omitempty"`
	MedicationNote    *string    `db:"medication_note" json:"medication_note,omitempty"`
	Note              *string    `db:"note" json:"note,omitempty"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

// EDTracking maps to the ed_tracking table.
type EDTracking struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	PatientID        uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID      uuid.UUID  `db:"encounter_id" json:"encounter_id"`
	TriageRecordID   *uuid.UUID `db:"triage_record_id" json:"triage_record_id,omitempty"`
	CurrentStatus    string     `db:"current_status" json:"current_status"`
	BedAssignment    *string    `db:"bed_assignment" json:"bed_assignment,omitempty"`
	AttendingID      *uuid.UUID `db:"attending_id" json:"attending_id,omitempty"`
	NurseID          *uuid.UUID `db:"nurse_id" json:"nurse_id,omitempty"`
	ArrivalTime      *time.Time `db:"arrival_time" json:"arrival_time,omitempty"`
	DischargeTime    *time.Time `db:"discharge_time" json:"discharge_time,omitempty"`
	Disposition      *string    `db:"disposition" json:"disposition,omitempty"`
	DispositionDest  *string    `db:"disposition_dest" json:"disposition_dest,omitempty"`
	LengthOfStayMins *int       `db:"length_of_stay_mins" json:"length_of_stay_mins,omitempty"`
	Note             *string    `db:"note" json:"note,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

// EDStatusHistory maps to the ed_status_history table.
type EDStatusHistory struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	EDTrackingID uuid.UUID  `db:"ed_tracking_id" json:"ed_tracking_id"`
	Status       string     `db:"status" json:"status"`
	ChangedAt    time.Time  `db:"changed_at" json:"changed_at"`
	ChangedBy    *uuid.UUID `db:"changed_by" json:"changed_by,omitempty"`
	Note         *string    `db:"note" json:"note,omitempty"`
}

// TraumaActivation maps to the trauma_activation table.
type TraumaActivation struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	PatientID         uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID       *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	EDTrackingID      *uuid.UUID `db:"ed_tracking_id" json:"ed_tracking_id,omitempty"`
	ActivationLevel   string     `db:"activation_level" json:"activation_level"`
	ActivationTime    time.Time  `db:"activation_time" json:"activation_time"`
	DeactivationTime  *time.Time `db:"deactivation_time" json:"deactivation_time,omitempty"`
	MechanismOfInjury *string    `db:"mechanism_of_injury" json:"mechanism_of_injury,omitempty"`
	ActivatedBy       *uuid.UUID `db:"activated_by" json:"activated_by,omitempty"`
	TeamLeadID        *uuid.UUID `db:"team_lead_id" json:"team_lead_id,omitempty"`
	Outcome           *string    `db:"outcome" json:"outcome,omitempty"`
	Note              *string    `db:"note" json:"note,omitempty"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}
