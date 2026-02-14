package oncology

import (
	"time"

	"github.com/google/uuid"
)

// CancerDiagnosis maps to the cancer_diagnosis table.
type CancerDiagnosis struct {
	ID                   uuid.UUID  `db:"id" json:"id"`
	PatientID            uuid.UUID  `db:"patient_id" json:"patient_id"`
	ConditionID          *uuid.UUID `db:"condition_id" json:"condition_id,omitempty"`
	DiagnosisDate        time.Time  `db:"diagnosis_date" json:"diagnosis_date"`
	CancerType           *string    `db:"cancer_type" json:"cancer_type,omitempty"`
	CancerSite           *string    `db:"cancer_site" json:"cancer_site,omitempty"`
	HistologyCode        *string    `db:"histology_code" json:"histology_code,omitempty"`
	HistologyDisplay     *string    `db:"histology_display" json:"histology_display,omitempty"`
	MorphologyCode       *string    `db:"morphology_code" json:"morphology_code,omitempty"`
	MorphologyDisplay    *string    `db:"morphology_display" json:"morphology_display,omitempty"`
	StagingSystem        *string    `db:"staging_system" json:"staging_system,omitempty"`
	StageGroup           *string    `db:"stage_group" json:"stage_group,omitempty"`
	TStage               *string    `db:"t_stage" json:"t_stage,omitempty"`
	NStage               *string    `db:"n_stage" json:"n_stage,omitempty"`
	MStage               *string    `db:"m_stage" json:"m_stage,omitempty"`
	Grade                *string    `db:"grade" json:"grade,omitempty"`
	Laterality           *string    `db:"laterality" json:"laterality,omitempty"`
	CurrentStatus        string     `db:"current_status" json:"current_status"`
	DiagnosingProviderID *uuid.UUID `db:"diagnosing_provider_id" json:"diagnosing_provider_id,omitempty"`
	ManagingProviderID   *uuid.UUID `db:"managing_provider_id" json:"managing_provider_id,omitempty"`
	ICD10Code            *string    `db:"icd10_code" json:"icd10_code,omitempty"`
	ICD10Display         *string    `db:"icd10_display" json:"icd10_display,omitempty"`
	ICDO3Topography      *string    `db:"icdo3_topography" json:"icdo3_topography,omitempty"`
	ICDO3Morphology      *string    `db:"icdo3_morphology" json:"icdo3_morphology,omitempty"`
	Note                 *string    `db:"note" json:"note,omitempty"`
	CreatedAt            time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updated_at"`
}

// TreatmentProtocol maps to the treatment_protocol table.
type TreatmentProtocol struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	CancerDiagnosisID     uuid.UUID  `db:"cancer_diagnosis_id" json:"cancer_diagnosis_id"`
	ProtocolName          string     `db:"protocol_name" json:"protocol_name"`
	ProtocolCode          *string    `db:"protocol_code" json:"protocol_code,omitempty"`
	ProtocolType          *string    `db:"protocol_type" json:"protocol_type,omitempty"`
	Intent                *string    `db:"intent" json:"intent,omitempty"`
	NumberOfCycles        *int       `db:"number_of_cycles" json:"number_of_cycles,omitempty"`
	CycleLengthDays       *int       `db:"cycle_length_days" json:"cycle_length_days,omitempty"`
	StartDate             *time.Time `db:"start_date" json:"start_date,omitempty"`
	EndDate               *time.Time `db:"end_date" json:"end_date,omitempty"`
	Status                string     `db:"status" json:"status"`
	PrescribingProviderID *uuid.UUID `db:"prescribing_provider_id" json:"prescribing_provider_id,omitempty"`
	ClinicalTrialID       *string    `db:"clinical_trial_id" json:"clinical_trial_id,omitempty"`
	Note                  *string    `db:"note" json:"note,omitempty"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

// TreatmentProtocolDrug maps to the treatment_protocol_drug table.
type TreatmentProtocolDrug struct {
	ID                    uuid.UUID `db:"id" json:"id"`
	ProtocolID            uuid.UUID `db:"protocol_id" json:"protocol_id"`
	DrugName              string    `db:"drug_name" json:"drug_name"`
	DrugCode              *string   `db:"drug_code" json:"drug_code,omitempty"`
	DrugCodeSystem        *string   `db:"drug_code_system" json:"drug_code_system,omitempty"`
	Route                 *string   `db:"route" json:"route,omitempty"`
	DoseValue             *float64  `db:"dose_value" json:"dose_value,omitempty"`
	DoseUnit              *string   `db:"dose_unit" json:"dose_unit,omitempty"`
	DoseCalculationMethod *string   `db:"dose_calculation_method" json:"dose_calculation_method,omitempty"`
	Frequency             *string   `db:"frequency" json:"frequency,omitempty"`
	AdministrationDay     *string   `db:"administration_day" json:"administration_day,omitempty"`
	InfusionDurationMin   *int      `db:"infusion_duration_min" json:"infusion_duration_min,omitempty"`
	Premedication         *string   `db:"premedication" json:"premedication,omitempty"`
	SequenceOrder         *int      `db:"sequence_order" json:"sequence_order,omitempty"`
	Note                  *string   `db:"note" json:"note,omitempty"`
}

// ChemoCycle maps to the chemotherapy_cycle table.
type ChemoCycle struct {
	ID                   uuid.UUID  `db:"id" json:"id"`
	ProtocolID           uuid.UUID  `db:"protocol_id" json:"protocol_id"`
	CycleNumber          int        `db:"cycle_number" json:"cycle_number"`
	PlannedStartDate     *time.Time `db:"planned_start_date" json:"planned_start_date,omitempty"`
	ActualStartDate      *time.Time `db:"actual_start_date" json:"actual_start_date,omitempty"`
	ActualEndDate        *time.Time `db:"actual_end_date" json:"actual_end_date,omitempty"`
	Status               string     `db:"status" json:"status"`
	DoseReductionPct     *float64   `db:"dose_reduction_pct" json:"dose_reduction_pct,omitempty"`
	DoseReductionReason  *string    `db:"dose_reduction_reason" json:"dose_reduction_reason,omitempty"`
	DelayDays            *int       `db:"delay_days" json:"delay_days,omitempty"`
	DelayReason          *string    `db:"delay_reason" json:"delay_reason,omitempty"`
	BSAM2                *float64   `db:"bsa_m2" json:"bsa_m2,omitempty"`
	WeightKG             *float64   `db:"weight_kg" json:"weight_kg,omitempty"`
	HeightCM             *float64   `db:"height_cm" json:"height_cm,omitempty"`
	CreatinineClearance  *float64   `db:"creatinine_clearance" json:"creatinine_clearance,omitempty"`
	ProviderID           *uuid.UUID `db:"provider_id" json:"provider_id,omitempty"`
	Note                 *string    `db:"note" json:"note,omitempty"`
	CreatedAt            time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updated_at"`
}

// ChemoAdministration maps to the chemotherapy_administration table.
type ChemoAdministration struct {
	ID                      uuid.UUID  `db:"id" json:"id"`
	CycleID                 uuid.UUID  `db:"cycle_id" json:"cycle_id"`
	ProtocolDrugID          *uuid.UUID `db:"protocol_drug_id" json:"protocol_drug_id,omitempty"`
	DrugName                string     `db:"drug_name" json:"drug_name"`
	AdministrationDatetime  time.Time  `db:"administration_datetime" json:"administration_datetime"`
	DoseGiven               *float64   `db:"dose_given" json:"dose_given,omitempty"`
	DoseUnit                *string    `db:"dose_unit" json:"dose_unit,omitempty"`
	Route                   *string    `db:"route" json:"route,omitempty"`
	InfusionDurationMin     *int       `db:"infusion_duration_min" json:"infusion_duration_min,omitempty"`
	InfusionRate            *string    `db:"infusion_rate" json:"infusion_rate,omitempty"`
	Site                    *string    `db:"site" json:"site,omitempty"`
	SequenceNumber          *int       `db:"sequence_number" json:"sequence_number,omitempty"`
	ReactionType            *string    `db:"reaction_type" json:"reaction_type,omitempty"`
	ReactionSeverity        *string    `db:"reaction_severity" json:"reaction_severity,omitempty"`
	ReactionAction          *string    `db:"reaction_action" json:"reaction_action,omitempty"`
	AdministeringNurseID    *uuid.UUID `db:"administering_nurse_id" json:"administering_nurse_id,omitempty"`
	SupervisingProviderID   *uuid.UUID `db:"supervising_provider_id" json:"supervising_provider_id,omitempty"`
	Note                    *string    `db:"note" json:"note,omitempty"`
}

// RadiationTherapy maps to the radiation_therapy table.
type RadiationTherapy struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	CancerDiagnosisID     uuid.UUID  `db:"cancer_diagnosis_id" json:"cancer_diagnosis_id"`
	TherapyType           *string    `db:"therapy_type" json:"therapy_type,omitempty"`
	Modality              *string    `db:"modality" json:"modality,omitempty"`
	Technique             *string    `db:"technique" json:"technique,omitempty"`
	TargetSite            *string    `db:"target_site" json:"target_site,omitempty"`
	Laterality            *string    `db:"laterality" json:"laterality,omitempty"`
	TotalDoseCGY          *float64   `db:"total_dose_cgy" json:"total_dose_cgy,omitempty"`
	DosePerFractionCGY    *float64   `db:"dose_per_fraction_cgy" json:"dose_per_fraction_cgy,omitempty"`
	PlannedFractions      *int       `db:"planned_fractions" json:"planned_fractions,omitempty"`
	CompletedFractions    *int       `db:"completed_fractions" json:"completed_fractions,omitempty"`
	StartDate             *time.Time `db:"start_date" json:"start_date,omitempty"`
	EndDate               *time.Time `db:"end_date" json:"end_date,omitempty"`
	Status                string     `db:"status" json:"status"`
	PrescribingProviderID *uuid.UUID `db:"prescribing_provider_id" json:"prescribing_provider_id,omitempty"`
	TreatingFacilityID    *uuid.UUID `db:"treating_facility_id" json:"treating_facility_id,omitempty"`
	EnergyType            *string    `db:"energy_type" json:"energy_type,omitempty"`
	EnergyValue           *string    `db:"energy_value" json:"energy_value,omitempty"`
	TreatmentVolumeCC     *float64   `db:"treatment_volume_cc" json:"treatment_volume_cc,omitempty"`
	Note                  *string    `db:"note" json:"note,omitempty"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

// RadiationSession maps to the radiation_therapy_session table.
type RadiationSession struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	RadiationTherapyID uuid.UUID  `db:"radiation_therapy_id" json:"radiation_therapy_id"`
	SessionNumber      int        `db:"session_number" json:"session_number"`
	SessionDate        time.Time  `db:"session_date" json:"session_date"`
	DoseDeliveredCGY   *float64   `db:"dose_delivered_cgy" json:"dose_delivered_cgy,omitempty"`
	FieldName          *string    `db:"field_name" json:"field_name,omitempty"`
	SetupVerified      *bool      `db:"setup_verified" json:"setup_verified,omitempty"`
	ImagingType        *string    `db:"imaging_type" json:"imaging_type,omitempty"`
	SkinReactionGrade  *int       `db:"skin_reaction_grade" json:"skin_reaction_grade,omitempty"`
	FatigueGrade       *int       `db:"fatigue_grade" json:"fatigue_grade,omitempty"`
	OtherToxicity      *string    `db:"other_toxicity" json:"other_toxicity,omitempty"`
	ToxicityGrade      *int       `db:"toxicity_grade" json:"toxicity_grade,omitempty"`
	MachineID          *string    `db:"machine_id" json:"machine_id,omitempty"`
	TherapistID        *uuid.UUID `db:"therapist_id" json:"therapist_id,omitempty"`
	PhysicistID        *uuid.UUID `db:"physicist_id" json:"physicist_id,omitempty"`
	Note               *string    `db:"note" json:"note,omitempty"`
}

// TumorMarker maps to the tumor_marker table.
type TumorMarker struct {
	ID                  uuid.UUID  `db:"id" json:"id"`
	CancerDiagnosisID   *uuid.UUID `db:"cancer_diagnosis_id" json:"cancer_diagnosis_id,omitempty"`
	PatientID           uuid.UUID  `db:"patient_id" json:"patient_id"`
	MarkerName          string     `db:"marker_name" json:"marker_name"`
	MarkerCode          *string    `db:"marker_code" json:"marker_code,omitempty"`
	MarkerCodeSystem    *string    `db:"marker_code_system" json:"marker_code_system,omitempty"`
	ValueQuantity       *float64   `db:"value_quantity" json:"value_quantity,omitempty"`
	ValueUnit           *string    `db:"value_unit" json:"value_unit,omitempty"`
	ValueString         *string    `db:"value_string" json:"value_string,omitempty"`
	ValueInterpretation *string    `db:"value_interpretation" json:"value_interpretation,omitempty"`
	ReferenceRangeLow   *float64   `db:"reference_range_low" json:"reference_range_low,omitempty"`
	ReferenceRangeHigh  *float64   `db:"reference_range_high" json:"reference_range_high,omitempty"`
	ReferenceRangeText  *string    `db:"reference_range_text" json:"reference_range_text,omitempty"`
	SpecimenType        *string    `db:"specimen_type" json:"specimen_type,omitempty"`
	CollectionDatetime  *time.Time `db:"collection_datetime" json:"collection_datetime,omitempty"`
	ResultDatetime      *time.Time `db:"result_datetime" json:"result_datetime,omitempty"`
	PerformingLab       *string    `db:"performing_lab" json:"performing_lab,omitempty"`
	OrderingProviderID  *uuid.UUID `db:"ordering_provider_id" json:"ordering_provider_id,omitempty"`
	Note                *string    `db:"note" json:"note,omitempty"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}

// TumorBoardReview maps to the tumor_board_review table.
type TumorBoardReview struct {
	ID                       uuid.UUID  `db:"id" json:"id"`
	CancerDiagnosisID        uuid.UUID  `db:"cancer_diagnosis_id" json:"cancer_diagnosis_id"`
	PatientID                uuid.UUID  `db:"patient_id" json:"patient_id"`
	ReviewDate               time.Time  `db:"review_date" json:"review_date"`
	ReviewType               *string    `db:"review_type" json:"review_type,omitempty"`
	PresentingProviderID     *uuid.UUID `db:"presenting_provider_id" json:"presenting_provider_id,omitempty"`
	Attendees                *string    `db:"attendees" json:"attendees,omitempty"`
	ClinicalSummary          *string    `db:"clinical_summary" json:"clinical_summary,omitempty"`
	PathologySummary         *string    `db:"pathology_summary" json:"pathology_summary,omitempty"`
	ImagingSummary           *string    `db:"imaging_summary" json:"imaging_summary,omitempty"`
	Discussion               *string    `db:"discussion" json:"discussion,omitempty"`
	Recommendations          *string    `db:"recommendations" json:"recommendations,omitempty"`
	TreatmentDecision        *string    `db:"treatment_decision" json:"treatment_decision,omitempty"`
	ClinicalTrialDiscussed   *bool      `db:"clinical_trial_discussed" json:"clinical_trial_discussed,omitempty"`
	ClinicalTrialID          *string    `db:"clinical_trial_id" json:"clinical_trial_id,omitempty"`
	NextReviewDate           *time.Time `db:"next_review_date" json:"next_review_date,omitempty"`
	Note                     *string    `db:"note" json:"note,omitempty"`
	CreatedAt                time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt                time.Time  `db:"updated_at" json:"updated_at"`
}
