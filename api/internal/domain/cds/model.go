package cds

import (
	"time"

	"github.com/google/uuid"
)

// CDSRule maps to the cds_rule table.
type CDSRule struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	RuleName        string     `db:"rule_name" json:"rule_name"`
	RuleType        string     `db:"rule_type" json:"rule_type"`
	Description     *string    `db:"description" json:"description,omitempty"`
	Severity        *string    `db:"severity" json:"severity,omitempty"`
	Category        *string    `db:"category" json:"category,omitempty"`
	TriggerEvent    *string    `db:"trigger_event" json:"trigger_event,omitempty"`
	ConditionExpr   *string    `db:"condition_expr" json:"condition_expr,omitempty"`
	ActionType      *string    `db:"action_type" json:"action_type,omitempty"`
	ActionDetail    *string    `db:"action_detail" json:"action_detail,omitempty"`
	EvidenceSource  *string    `db:"evidence_source" json:"evidence_source,omitempty"`
	EvidenceURL     *string    `db:"evidence_url" json:"evidence_url,omitempty"`
	Active          bool       `db:"active" json:"active"`
	Version         *string    `db:"version" json:"version,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// CDSAlert maps to the cds_alert table.
type CDSAlert struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	RuleID          uuid.UUID  `db:"rule_id" json:"rule_id"`
	PatientID       uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID     *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	PractitionerID  *uuid.UUID `db:"practitioner_id" json:"practitioner_id,omitempty"`
	Status          string     `db:"status" json:"status"`
	Severity        *string    `db:"severity" json:"severity,omitempty"`
	Summary         string     `db:"summary" json:"summary"`
	Detail          *string    `db:"detail" json:"detail,omitempty"`
	SuggestedAction *string    `db:"suggested_action" json:"suggested_action,omitempty"`
	Source          *string    `db:"source" json:"source,omitempty"`
	ExpiresAt       *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	FiredAt         time.Time  `db:"fired_at" json:"fired_at"`
	ResolvedAt      *time.Time `db:"resolved_at" json:"resolved_at,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// CDSAlertResponse maps to the cds_alert_response table.
type CDSAlertResponse struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	AlertID         uuid.UUID  `db:"alert_id" json:"alert_id"`
	PractitionerID  uuid.UUID  `db:"practitioner_id" json:"practitioner_id"`
	Action          string     `db:"action" json:"action"`
	Reason          *string    `db:"reason" json:"reason,omitempty"`
	Comment         *string    `db:"comment" json:"comment,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
}

// DrugInteraction maps to the drug_interaction table.
type DrugInteraction struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	MedicationAID   *uuid.UUID `db:"medication_a_id" json:"medication_a_id,omitempty"`
	MedicationAName string     `db:"medication_a_name" json:"medication_a_name"`
	MedicationBID   *uuid.UUID `db:"medication_b_id" json:"medication_b_id,omitempty"`
	MedicationBName string     `db:"medication_b_name" json:"medication_b_name"`
	Severity        string     `db:"severity" json:"severity"`
	Description     *string    `db:"description" json:"description,omitempty"`
	ClinicalEffect  *string    `db:"clinical_effect" json:"clinical_effect,omitempty"`
	Management      *string    `db:"management" json:"management,omitempty"`
	EvidenceLevel   *string    `db:"evidence_level" json:"evidence_level,omitempty"`
	Source          *string    `db:"source" json:"source,omitempty"`
	Active          bool       `db:"active" json:"active"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// OrderSet maps to the order_set table.
type OrderSet struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	Name            string     `db:"name" json:"name"`
	Description     *string    `db:"description" json:"description,omitempty"`
	Category        *string    `db:"category" json:"category,omitempty"`
	Status          string     `db:"status" json:"status"`
	AuthorID        *uuid.UUID `db:"author_id" json:"author_id,omitempty"`
	Version         *string    `db:"version" json:"version,omitempty"`
	ApprovalDate    *time.Time `db:"approval_date" json:"approval_date,omitempty"`
	Active          bool       `db:"active" json:"active"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// OrderSetSection maps to the order_set_section table.
type OrderSetSection struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	OrderSetID      uuid.UUID  `db:"order_set_id" json:"order_set_id"`
	Name            string     `db:"name" json:"name"`
	Description     *string    `db:"description" json:"description,omitempty"`
	SortOrder       int        `db:"sort_order" json:"sort_order"`
}

// OrderSetItem maps to the order_set_item table.
type OrderSetItem struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	SectionID       uuid.UUID  `db:"section_id" json:"section_id"`
	ItemType        string     `db:"item_type" json:"item_type"`
	ItemName        string     `db:"item_name" json:"item_name"`
	ItemCode        *string    `db:"item_code" json:"item_code,omitempty"`
	DefaultDose     *string    `db:"default_dose" json:"default_dose,omitempty"`
	DefaultFrequency *string   `db:"default_frequency" json:"default_frequency,omitempty"`
	DefaultDuration *string    `db:"default_duration" json:"default_duration,omitempty"`
	Instructions    *string    `db:"instructions" json:"instructions,omitempty"`
	IsRequired      bool       `db:"is_required" json:"is_required"`
	SortOrder       int        `db:"sort_order" json:"sort_order"`
}

// ClinicalPathway maps to the clinical_pathway table.
type ClinicalPathway struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	Name            string     `db:"name" json:"name"`
	Description     *string    `db:"description" json:"description,omitempty"`
	Condition       *string    `db:"condition" json:"condition,omitempty"`
	Category        *string    `db:"category" json:"category,omitempty"`
	Version         *string    `db:"version" json:"version,omitempty"`
	AuthorID        *uuid.UUID `db:"author_id" json:"author_id,omitempty"`
	Active          bool       `db:"active" json:"active"`
	ExpectedDuration *string   `db:"expected_duration" json:"expected_duration,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// ClinicalPathwayPhase maps to the clinical_pathway_phase table.
type ClinicalPathwayPhase struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	PathwayID       uuid.UUID  `db:"pathway_id" json:"pathway_id"`
	Name            string     `db:"name" json:"name"`
	Description     *string    `db:"description" json:"description,omitempty"`
	Duration        *string    `db:"duration" json:"duration,omitempty"`
	Goals           *string    `db:"goals" json:"goals,omitempty"`
	Interventions   *string    `db:"interventions" json:"interventions,omitempty"`
	SortOrder       int        `db:"sort_order" json:"sort_order"`
}

// PatientPathwayEnrollment maps to the patient_pathway_enrollment table.
type PatientPathwayEnrollment struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	PathwayID       uuid.UUID  `db:"pathway_id" json:"pathway_id"`
	PatientID       uuid.UUID  `db:"patient_id" json:"patient_id"`
	PractitionerID  *uuid.UUID `db:"practitioner_id" json:"practitioner_id,omitempty"`
	Status          string     `db:"status" json:"status"`
	CurrentPhaseID  *uuid.UUID `db:"current_phase_id" json:"current_phase_id,omitempty"`
	EnrolledAt      time.Time  `db:"enrolled_at" json:"enrolled_at"`
	CompletedAt     *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	Note            *string    `db:"note" json:"note,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// Formulary maps to the formulary table.
type Formulary struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	Name            string     `db:"name" json:"name"`
	Description     *string    `db:"description" json:"description,omitempty"`
	OrganizationID  *uuid.UUID `db:"organization_id" json:"organization_id,omitempty"`
	EffectiveDate   *time.Time `db:"effective_date" json:"effective_date,omitempty"`
	ExpirationDate  *time.Time `db:"expiration_date" json:"expiration_date,omitempty"`
	Version         *string    `db:"version" json:"version,omitempty"`
	Active          bool       `db:"active" json:"active"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// FormularyItem maps to the formulary_item table.
type FormularyItem struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	FormularyID     uuid.UUID  `db:"formulary_id" json:"formulary_id"`
	MedicationID    *uuid.UUID `db:"medication_id" json:"medication_id,omitempty"`
	MedicationName  string     `db:"medication_name" json:"medication_name"`
	TierLevel       *int       `db:"tier_level" json:"tier_level,omitempty"`
	RequiresPriorAuth bool     `db:"requires_prior_auth" json:"requires_prior_auth"`
	StepTherapyReq bool       `db:"step_therapy_req" json:"step_therapy_req"`
	QuantityLimit   *string    `db:"quantity_limit" json:"quantity_limit,omitempty"`
	PreferredStatus *string    `db:"preferred_status" json:"preferred_status,omitempty"`
	Note            *string    `db:"note" json:"note,omitempty"`
}

// MedicationReconciliation maps to the medication_reconciliation table.
type MedicationReconciliation struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	PatientID       uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID     *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	PractitionerID  *uuid.UUID `db:"practitioner_id" json:"practitioner_id,omitempty"`
	Status          string     `db:"status" json:"status"`
	ReconcType      *string    `db:"reconc_type" json:"reconc_type,omitempty"`
	CompletedAt     *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	Note            *string    `db:"note" json:"note,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// MedicationReconciliationItem maps to the medication_reconciliation_item table.
type MedicationReconciliationItem struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	ReconciliationID  uuid.UUID  `db:"reconciliation_id" json:"reconciliation_id"`
	MedicationID      *uuid.UUID `db:"medication_id" json:"medication_id,omitempty"`
	MedicationName    string     `db:"medication_name" json:"medication_name"`
	SourceList        *string    `db:"source_list" json:"source_list,omitempty"`
	Dose              *string    `db:"dose" json:"dose,omitempty"`
	Frequency         *string    `db:"frequency" json:"frequency,omitempty"`
	Route             *string    `db:"route" json:"route,omitempty"`
	Action            *string    `db:"action" json:"action,omitempty"`
	Reason            *string    `db:"reason" json:"reason,omitempty"`
	VerifiedByID      *uuid.UUID `db:"verified_by_id" json:"verified_by_id,omitempty"`
	VerifiedAt        *time.Time `db:"verified_at" json:"verified_at,omitempty"`
}
