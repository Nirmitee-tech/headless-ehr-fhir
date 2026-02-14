package nursing

import (
	"time"

	"github.com/google/uuid"
)

// FlowsheetTemplate maps to the flowsheet_template table.
type FlowsheetTemplate struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Description *string    `db:"description" json:"description,omitempty"`
	Category    *string    `db:"category" json:"category,omitempty"`
	IsActive    bool       `db:"is_active" json:"is_active"`
	CreatedBy   *uuid.UUID `db:"created_by" json:"created_by,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

// FlowsheetRow maps to the flowsheet_row table.
type FlowsheetRow struct {
	ID             uuid.UUID `db:"id" json:"id"`
	TemplateID     uuid.UUID `db:"template_id" json:"template_id"`
	Label          string    `db:"label" json:"label"`
	DataType       string    `db:"data_type" json:"data_type"`
	Unit           *string   `db:"unit" json:"unit,omitempty"`
	AllowedValues  []string  `db:"allowed_values" json:"allowed_values,omitempty"`
	SortOrder      int       `db:"sort_order" json:"sort_order"`
	IsRequired     bool      `db:"is_required" json:"is_required"`
}

// FlowsheetEntry maps to the flowsheet_entry table.
type FlowsheetEntry struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	TemplateID   uuid.UUID  `db:"template_id" json:"template_id"`
	RowID        uuid.UUID  `db:"row_id" json:"row_id"`
	PatientID    uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID  uuid.UUID  `db:"encounter_id" json:"encounter_id"`
	ValueText    *string    `db:"value_text" json:"value_text,omitempty"`
	ValueNumeric *float64   `db:"value_numeric" json:"value_numeric,omitempty"`
	RecordedAt   time.Time  `db:"recorded_at" json:"recorded_at"`
	RecordedByID uuid.UUID  `db:"recorded_by_id" json:"recorded_by_id"`
	Note         *string    `db:"note" json:"note,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
}

// NursingAssessment maps to the nursing_assessment table.
type NursingAssessment struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	PatientID       uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID     uuid.UUID  `db:"encounter_id" json:"encounter_id"`
	NurseID         uuid.UUID  `db:"nurse_id" json:"nurse_id"`
	AssessmentType  string     `db:"assessment_type" json:"assessment_type"`
	AssessmentData  *string    `db:"assessment_data" json:"assessment_data,omitempty"`
	Status          string     `db:"status" json:"status"`
	CompletedAt     *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	Note            *string    `db:"note" json:"note,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// FallRiskAssessment maps to the fall_risk_assessment table.
type FallRiskAssessment struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	PatientID       uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID     *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	AssessedByID    uuid.UUID  `db:"assessed_by_id" json:"assessed_by_id"`
	ToolUsed        *string    `db:"tool_used" json:"tool_used,omitempty"`
	TotalScore      *int       `db:"total_score" json:"total_score,omitempty"`
	RiskLevel       *string    `db:"risk_level" json:"risk_level,omitempty"`
	HistoryOfFalls  *bool      `db:"history_of_falls" json:"history_of_falls,omitempty"`
	Medications     *bool      `db:"medications" json:"medications,omitempty"`
	GaitBalance     *string    `db:"gait_balance" json:"gait_balance,omitempty"`
	MentalStatus    *string    `db:"mental_status" json:"mental_status,omitempty"`
	Interventions   *string    `db:"interventions" json:"interventions,omitempty"`
	Note            *string    `db:"note" json:"note,omitempty"`
	AssessedAt      time.Time  `db:"assessed_at" json:"assessed_at"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
}

// SkinAssessment maps to the skin_assessment table.
type SkinAssessment struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	PatientID       uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID     *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	AssessedByID    uuid.UUID  `db:"assessed_by_id" json:"assessed_by_id"`
	ToolUsed        *string    `db:"tool_used" json:"tool_used,omitempty"`
	TotalScore      *int       `db:"total_score" json:"total_score,omitempty"`
	RiskLevel       *string    `db:"risk_level" json:"risk_level,omitempty"`
	SkinIntegrity   *string    `db:"skin_integrity" json:"skin_integrity,omitempty"`
	MoistureLevel   *string    `db:"moisture_level" json:"moisture_level,omitempty"`
	Mobility        *string    `db:"mobility" json:"mobility,omitempty"`
	Nutrition       *string    `db:"nutrition" json:"nutrition,omitempty"`
	WoundPresent    *bool      `db:"wound_present" json:"wound_present,omitempty"`
	WoundLocation   *string    `db:"wound_location" json:"wound_location,omitempty"`
	WoundStage      *string    `db:"wound_stage" json:"wound_stage,omitempty"`
	Interventions   *string    `db:"interventions" json:"interventions,omitempty"`
	Note            *string    `db:"note" json:"note,omitempty"`
	AssessedAt      time.Time  `db:"assessed_at" json:"assessed_at"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
}

// PainAssessment maps to the pain_assessment table.
type PainAssessment struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	PatientID       uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID     *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	AssessedByID    uuid.UUID  `db:"assessed_by_id" json:"assessed_by_id"`
	ToolUsed        *string    `db:"tool_used" json:"tool_used,omitempty"`
	PainScore       *int       `db:"pain_score" json:"pain_score,omitempty"`
	PainLocation    *string    `db:"pain_location" json:"pain_location,omitempty"`
	PainCharacter   *string    `db:"pain_character" json:"pain_character,omitempty"`
	PainDuration    *string    `db:"pain_duration" json:"pain_duration,omitempty"`
	PainRadiation   *string    `db:"pain_radiation" json:"pain_radiation,omitempty"`
	Aggravating     *string    `db:"aggravating" json:"aggravating,omitempty"`
	Alleviating     *string    `db:"alleviating" json:"alleviating,omitempty"`
	Interventions   *string    `db:"interventions" json:"interventions,omitempty"`
	ReassessScore   *int       `db:"reassess_score" json:"reassess_score,omitempty"`
	Note            *string    `db:"note" json:"note,omitempty"`
	AssessedAt      time.Time  `db:"assessed_at" json:"assessed_at"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
}

// LinesDrainsAirways maps to the lines_drains_airways table.
type LinesDrainsAirways struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	PatientID       uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID     uuid.UUID  `db:"encounter_id" json:"encounter_id"`
	Type            string     `db:"type" json:"type"`
	Description     *string    `db:"description" json:"description,omitempty"`
	Site            *string    `db:"site" json:"site,omitempty"`
	Size            *string    `db:"size" json:"size,omitempty"`
	InsertedAt      *time.Time `db:"inserted_at" json:"inserted_at,omitempty"`
	InsertedByID    *uuid.UUID `db:"inserted_by_id" json:"inserted_by_id,omitempty"`
	RemovedAt       *time.Time `db:"removed_at" json:"removed_at,omitempty"`
	RemovedByID     *uuid.UUID `db:"removed_by_id" json:"removed_by_id,omitempty"`
	Status          string     `db:"status" json:"status"`
	DeviceID        *uuid.UUID `db:"device_id" json:"device_id,omitempty"`
	Note            *string    `db:"note" json:"note,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// RestraintRecord maps to the restraint_record table.
type RestraintRecord struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	PatientID        uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID      *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	RestraintType    string     `db:"restraint_type" json:"restraint_type"`
	Reason           *string    `db:"reason" json:"reason,omitempty"`
	BodySite         *string    `db:"body_site" json:"body_site,omitempty"`
	AppliedAt        time.Time  `db:"applied_at" json:"applied_at"`
	AppliedByID      uuid.UUID  `db:"applied_by_id" json:"applied_by_id"`
	RemovedAt        *time.Time `db:"removed_at" json:"removed_at,omitempty"`
	RemovedByID      *uuid.UUID `db:"removed_by_id" json:"removed_by_id,omitempty"`
	OrderID          *uuid.UUID `db:"order_id" json:"order_id,omitempty"`
	LastAssessedAt   *time.Time `db:"last_assessed_at" json:"last_assessed_at,omitempty"`
	LastAssessedByID *uuid.UUID `db:"last_assessed_by_id" json:"last_assessed_by_id,omitempty"`
	SkinCondition    *string    `db:"skin_condition" json:"skin_condition,omitempty"`
	Circulation      *string    `db:"circulation" json:"circulation,omitempty"`
	Note             *string    `db:"note" json:"note,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

// IntakeOutputRecord maps to the intake_output_record table.
type IntakeOutputRecord struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	PatientID    uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID  uuid.UUID  `db:"encounter_id" json:"encounter_id"`
	Category     string     `db:"category" json:"category"`
	Type         *string    `db:"type" json:"type,omitempty"`
	Volume       *float64   `db:"volume" json:"volume,omitempty"`
	Unit         *string    `db:"unit" json:"unit,omitempty"`
	Route        *string    `db:"route" json:"route,omitempty"`
	RecordedAt   time.Time  `db:"recorded_at" json:"recorded_at"`
	RecordedByID uuid.UUID  `db:"recorded_by_id" json:"recorded_by_id"`
	Note         *string    `db:"note" json:"note,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
}
