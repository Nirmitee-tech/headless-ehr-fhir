package surgery

import (
	"time"

	"github.com/google/uuid"
)

// ORRoom maps to the or_room table.
type ORRoom struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	Name            string     `db:"name" json:"name"`
	LocationID      *uuid.UUID `db:"location_id" json:"location_id,omitempty"`
	Status          string     `db:"status" json:"status"`
	RoomType        *string    `db:"room_type" json:"room_type,omitempty"`
	Equipment       *string    `db:"equipment" json:"equipment,omitempty"`
	IsActive        bool       `db:"is_active" json:"is_active"`
	DecontaminatedAt *time.Time `db:"decontaminated_at" json:"decontaminated_at,omitempty"`
	Note            *string    `db:"note" json:"note,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// SurgicalCase maps to the surgical_case table. This is the main surgical resource.
type SurgicalCase struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	PatientID         uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID       *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	PrimarySurgeonID  uuid.UUID  `db:"primary_surgeon_id" json:"primary_surgeon_id"`
	AnesthesiologistID *uuid.UUID `db:"anesthesiologist_id" json:"anesthesiologist_id,omitempty"`
	ORRoomID          *uuid.UUID `db:"or_room_id" json:"or_room_id,omitempty"`
	Status            string     `db:"status" json:"status"`
	CaseClass         *string    `db:"case_class" json:"case_class,omitempty"`
	ASAClass          *string    `db:"asa_class" json:"asa_class,omitempty"`
	WoundClass        *string    `db:"wound_class" json:"wound_class,omitempty"`
	ScheduledDate     time.Time  `db:"scheduled_date" json:"scheduled_date"`
	ScheduledStart    *time.Time `db:"scheduled_start" json:"scheduled_start,omitempty"`
	ScheduledEnd      *time.Time `db:"scheduled_end" json:"scheduled_end,omitempty"`
	ActualStart       *time.Time `db:"actual_start" json:"actual_start,omitempty"`
	ActualEnd         *time.Time `db:"actual_end" json:"actual_end,omitempty"`
	AnesthesiaType    *string    `db:"anesthesia_type" json:"anesthesia_type,omitempty"`
	Laterality        *string    `db:"laterality" json:"laterality,omitempty"`
	PreOpDiagnosis    *string    `db:"pre_op_diagnosis" json:"pre_op_diagnosis,omitempty"`
	PostOpDiagnosis   *string    `db:"post_op_diagnosis" json:"post_op_diagnosis,omitempty"`
	CancelReason      *string    `db:"cancel_reason" json:"cancel_reason,omitempty"`
	Note              *string    `db:"note" json:"note,omitempty"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

// SurgicalCaseProcedure maps to the surgical_case_procedure table.
type SurgicalCaseProcedure struct {
	ID              uuid.UUID `db:"id" json:"id"`
	SurgicalCaseID  uuid.UUID `db:"surgical_case_id" json:"surgical_case_id"`
	ProcedureCode   string    `db:"procedure_code" json:"procedure_code"`
	ProcedureDisplay string   `db:"procedure_display" json:"procedure_display"`
	CodeSystem      *string   `db:"code_system" json:"code_system,omitempty"`
	CPTCode         *string   `db:"cpt_code" json:"cpt_code,omitempty"`
	IsPrimary       bool      `db:"is_primary" json:"is_primary"`
	BodySiteCode    *string   `db:"body_site_code" json:"body_site_code,omitempty"`
	BodySiteDisplay *string   `db:"body_site_display" json:"body_site_display,omitempty"`
	Sequence        int       `db:"sequence" json:"sequence"`
}

// SurgicalCaseTeam maps to the surgical_case_team table.
type SurgicalCaseTeam struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	SurgicalCaseID  uuid.UUID  `db:"surgical_case_id" json:"surgical_case_id"`
	PractitionerID  uuid.UUID  `db:"practitioner_id" json:"practitioner_id"`
	Role            string     `db:"role" json:"role"`
	RoleDisplay     *string    `db:"role_display" json:"role_display,omitempty"`
	StartTime       *time.Time `db:"start_time" json:"start_time,omitempty"`
	EndTime         *time.Time `db:"end_time" json:"end_time,omitempty"`
}

// SurgicalTimeEvent maps to the surgical_time_event table.
type SurgicalTimeEvent struct {
	ID             uuid.UUID `db:"id" json:"id"`
	SurgicalCaseID uuid.UUID `db:"surgical_case_id" json:"surgical_case_id"`
	EventType      string    `db:"event_type" json:"event_type"`
	EventTime      time.Time `db:"event_time" json:"event_time"`
	RecordedBy     *uuid.UUID `db:"recorded_by" json:"recorded_by,omitempty"`
	Note           *string   `db:"note" json:"note,omitempty"`
}

// SurgicalPreferenceCard maps to the surgical_preference_card table.
type SurgicalPreferenceCard struct {
	ID              uuid.UUID `db:"id" json:"id"`
	SurgeonID       uuid.UUID `db:"surgeon_id" json:"surgeon_id"`
	ProcedureCode   string    `db:"procedure_code" json:"procedure_code"`
	ProcedureDisplay string   `db:"procedure_display" json:"procedure_display"`
	GloveSizeL      *string   `db:"glove_size_l" json:"glove_size_l,omitempty"`
	GloveSizeR      *string   `db:"glove_size_r" json:"glove_size_r,omitempty"`
	Gown            *string   `db:"gown" json:"gown,omitempty"`
	SkinPrep        *string   `db:"skin_prep" json:"skin_prep,omitempty"`
	Position        *string   `db:"position" json:"position,omitempty"`
	Instruments     *string   `db:"instruments" json:"instruments,omitempty"`
	Supplies        *string   `db:"supplies" json:"supplies,omitempty"`
	Sutures         *string   `db:"sutures" json:"sutures,omitempty"`
	Dressings       *string   `db:"dressings" json:"dressings,omitempty"`
	SpecialEquipment *string  `db:"special_equipment" json:"special_equipment,omitempty"`
	Note            *string   `db:"note" json:"note,omitempty"`
	IsActive        bool      `db:"is_active" json:"is_active"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

// SurgicalCount maps to the surgical_count table.
type SurgicalCount struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	SurgicalCaseID uuid.UUID  `db:"surgical_case_id" json:"surgical_case_id"`
	CountType      string     `db:"count_type" json:"count_type"`
	ItemName       string     `db:"item_name" json:"item_name"`
	ExpectedCount  int        `db:"expected_count" json:"expected_count"`
	ActualCount    int        `db:"actual_count" json:"actual_count"`
	IsCorrect      bool       `db:"is_correct" json:"is_correct"`
	CountedBy      *uuid.UUID `db:"counted_by" json:"counted_by,omitempty"`
	VerifiedBy     *uuid.UUID `db:"verified_by" json:"verified_by,omitempty"`
	CountTime      time.Time  `db:"count_time" json:"count_time"`
	Note           *string    `db:"note" json:"note,omitempty"`
}

// ImplantLog maps to the implant_log table.
type ImplantLog struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	SurgicalCaseID  *uuid.UUID `db:"surgical_case_id" json:"surgical_case_id,omitempty"`
	PatientID       uuid.UUID  `db:"patient_id" json:"patient_id"`
	DeviceID        *uuid.UUID `db:"device_id" json:"device_id,omitempty"`
	ImplantType     string     `db:"implant_type" json:"implant_type"`
	Manufacturer    *string    `db:"manufacturer" json:"manufacturer,omitempty"`
	ModelNumber     *string    `db:"model_number" json:"model_number,omitempty"`
	SerialNumber    *string    `db:"serial_number" json:"serial_number,omitempty"`
	LotNumber       *string    `db:"lot_number" json:"lot_number,omitempty"`
	ExpirationDate  *time.Time `db:"expiration_date" json:"expiration_date,omitempty"`
	BodySiteCode    *string    `db:"body_site_code" json:"body_site_code,omitempty"`
	BodySiteDisplay *string    `db:"body_site_display" json:"body_site_display,omitempty"`
	ImplantedBy     *uuid.UUID `db:"implanted_by" json:"implanted_by,omitempty"`
	ImplantDate     *time.Time `db:"implant_date" json:"implant_date,omitempty"`
	ExplantDate     *time.Time `db:"explant_date" json:"explant_date,omitempty"`
	Note            *string    `db:"note" json:"note,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// SurgicalSupplyUsed maps to the surgical_supply_used table.
type SurgicalSupplyUsed struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	SurgicalCaseID uuid.UUID  `db:"surgical_case_id" json:"surgical_case_id"`
	SupplyName     string     `db:"supply_name" json:"supply_name"`
	SupplyCode     *string    `db:"supply_code" json:"supply_code,omitempty"`
	Quantity       int        `db:"quantity" json:"quantity"`
	UnitOfMeasure  *string    `db:"unit_of_measure" json:"unit_of_measure,omitempty"`
	LotNumber      *string    `db:"lot_number" json:"lot_number,omitempty"`
	RecordedBy     *uuid.UUID `db:"recorded_by" json:"recorded_by,omitempty"`
	Note           *string    `db:"note" json:"note,omitempty"`
}
