package obstetrics

import (
	"time"

	"github.com/google/uuid"
)

// Pregnancy maps to the pregnancy table.
type Pregnancy struct {
	ID                     uuid.UUID  `db:"id" json:"id"`
	PatientID              uuid.UUID  `db:"patient_id" json:"patient_id"`
	Status                 string     `db:"status" json:"status"`
	OnsetDate              *time.Time `db:"onset_date" json:"onset_date,omitempty"`
	EstimatedDueDate       *time.Time `db:"estimated_due_date" json:"estimated_due_date,omitempty"`
	LastMenstrualPeriod    *time.Time `db:"last_menstrual_period" json:"last_menstrual_period,omitempty"`
	ConceptionMethod       *string    `db:"conception_method" json:"conception_method,omitempty"`
	Gravida                *int       `db:"gravida" json:"gravida,omitempty"`
	Para                   *int       `db:"para" json:"para,omitempty"`
	MultipleGestation      *bool      `db:"multiple_gestation" json:"multiple_gestation,omitempty"`
	NumberOfFetuses        *int       `db:"number_of_fetuses" json:"number_of_fetuses,omitempty"`
	RiskLevel              *string    `db:"risk_level" json:"risk_level,omitempty"`
	RiskFactors            *string    `db:"risk_factors" json:"risk_factors,omitempty"`
	BloodType              *string    `db:"blood_type" json:"blood_type,omitempty"`
	RhFactor               *string    `db:"rh_factor" json:"rh_factor,omitempty"`
	PrePregnancyWeight     *float64   `db:"pre_pregnancy_weight" json:"pre_pregnancy_weight,omitempty"`
	PrePregnancyBMI        *float64   `db:"pre_pregnancy_bmi" json:"pre_pregnancy_bmi,omitempty"`
	PrimaryProviderID      *uuid.UUID `db:"primary_provider_id" json:"primary_provider_id,omitempty"`
	ManagingOrganizationID *uuid.UUID `db:"managing_organization_id" json:"managing_organization_id,omitempty"`
	Note                   *string    `db:"note" json:"note,omitempty"`
	OutcomeDate            *time.Time `db:"outcome_date" json:"outcome_date,omitempty"`
	OutcomeSummary         *string    `db:"outcome_summary" json:"outcome_summary,omitempty"`
	CreatedAt              time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at" json:"updated_at"`
}

// PrenatalVisit maps to the prenatal_visit table.
type PrenatalVisit struct {
	ID                     uuid.UUID  `db:"id" json:"id"`
	PregnancyID            uuid.UUID  `db:"pregnancy_id" json:"pregnancy_id"`
	EncounterID            *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	VisitDate              time.Time  `db:"visit_date" json:"visit_date"`
	GestationalAgeWeeks    *int       `db:"gestational_age_weeks" json:"gestational_age_weeks,omitempty"`
	GestationalAgeDays     *int       `db:"gestational_age_days" json:"gestational_age_days,omitempty"`
	Weight                 *float64   `db:"weight" json:"weight,omitempty"`
	BloodPressureSystolic  *int       `db:"blood_pressure_systolic" json:"blood_pressure_systolic,omitempty"`
	BloodPressureDiastolic *int       `db:"blood_pressure_diastolic" json:"blood_pressure_diastolic,omitempty"`
	FundalHeight           *float64   `db:"fundal_height" json:"fundal_height,omitempty"`
	FetalHeartRate         *int       `db:"fetal_heart_rate" json:"fetal_heart_rate,omitempty"`
	FetalPresentation      *string    `db:"fetal_presentation" json:"fetal_presentation,omitempty"`
	FetalMovement          *string    `db:"fetal_movement" json:"fetal_movement,omitempty"`
	UrineProtein           *string    `db:"urine_protein" json:"urine_protein,omitempty"`
	UrineGlucose           *string    `db:"urine_glucose" json:"urine_glucose,omitempty"`
	Edema                  *string    `db:"edema" json:"edema,omitempty"`
	CervicalDilation       *float64   `db:"cervical_dilation" json:"cervical_dilation,omitempty"`
	CervicalEffacement     *int       `db:"cervical_effacement" json:"cervical_effacement,omitempty"`
	GroupBStrepStatus      *string    `db:"group_b_strep_status" json:"group_b_strep_status,omitempty"`
	ProviderID             *uuid.UUID `db:"provider_id" json:"provider_id,omitempty"`
	Note                   *string    `db:"note" json:"note,omitempty"`
	NextVisitDate          *time.Time `db:"next_visit_date" json:"next_visit_date,omitempty"`
	CreatedAt              time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at" json:"updated_at"`
}

// LaborRecord maps to the labor_record table.
type LaborRecord struct {
	ID                      uuid.UUID  `db:"id" json:"id"`
	PregnancyID             uuid.UUID  `db:"pregnancy_id" json:"pregnancy_id"`
	EncounterID             *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	AdmissionDatetime       *time.Time `db:"admission_datetime" json:"admission_datetime,omitempty"`
	LaborOnsetDatetime      *time.Time `db:"labor_onset_datetime" json:"labor_onset_datetime,omitempty"`
	LaborOnsetType          *string    `db:"labor_onset_type" json:"labor_onset_type,omitempty"`
	MembraneRuptureDatetime *time.Time `db:"membrane_rupture_datetime" json:"membrane_rupture_datetime,omitempty"`
	MembraneRuptureType     *string    `db:"membrane_rupture_type" json:"membrane_rupture_type,omitempty"`
	AmnioticFluidColor      *string    `db:"amniotic_fluid_color" json:"amniotic_fluid_color,omitempty"`
	AmnioticFluidVolume     *string    `db:"amniotic_fluid_volume" json:"amniotic_fluid_volume,omitempty"`
	InductionMethod         *string    `db:"induction_method" json:"induction_method,omitempty"`
	InductionReason         *string    `db:"induction_reason" json:"induction_reason,omitempty"`
	AugmentationMethod      *string    `db:"augmentation_method" json:"augmentation_method,omitempty"`
	AnesthesiaType          *string    `db:"anesthesia_type" json:"anesthesia_type,omitempty"`
	AnesthesiaStart         *time.Time `db:"anesthesia_start" json:"anesthesia_start,omitempty"`
	Status                  string     `db:"status" json:"status"`
	AttendingProviderID     *uuid.UUID `db:"attending_provider_id" json:"attending_provider_id,omitempty"`
	Note                    *string    `db:"note" json:"note,omitempty"`
	CreatedAt               time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time  `db:"updated_at" json:"updated_at"`
}

// LaborCervicalExam maps to the labor_cervical_exam table.
type LaborCervicalExam struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	LaborRecordID  uuid.UUID  `db:"labor_record_id" json:"labor_record_id"`
	ExamDatetime   time.Time  `db:"exam_datetime" json:"exam_datetime"`
	DilationCM     *float64   `db:"dilation_cm" json:"dilation_cm,omitempty"`
	EffacementPct  *int       `db:"effacement_pct" json:"effacement_pct,omitempty"`
	Station        *string    `db:"station" json:"station,omitempty"`
	FetalPosition  *string    `db:"fetal_position" json:"fetal_position,omitempty"`
	MembraneStatus *string    `db:"membrane_status" json:"membrane_status,omitempty"`
	ExaminerID     *uuid.UUID `db:"examiner_id" json:"examiner_id,omitempty"`
	Note           *string    `db:"note" json:"note,omitempty"`
}

// FetalMonitoring maps to the fetal_monitoring table.
type FetalMonitoring struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	LaborRecordID         uuid.UUID  `db:"labor_record_id" json:"labor_record_id"`
	MonitoringDatetime    time.Time  `db:"monitoring_datetime" json:"monitoring_datetime"`
	MonitoringType        *string    `db:"monitoring_type" json:"monitoring_type,omitempty"`
	FetalHeartRate        *int       `db:"fetal_heart_rate" json:"fetal_heart_rate,omitempty"`
	BaselineRate          *int       `db:"baseline_rate" json:"baseline_rate,omitempty"`
	Variability           *string    `db:"variability" json:"variability,omitempty"`
	Accelerations         *string    `db:"accelerations" json:"accelerations,omitempty"`
	Decelerations         *string    `db:"decelerations" json:"decelerations,omitempty"`
	DecelerationType      *string    `db:"deceleration_type" json:"deceleration_type,omitempty"`
	ContractionFrequency  *string    `db:"contraction_frequency" json:"contraction_frequency,omitempty"`
	ContractionDuration   *string    `db:"contraction_duration" json:"contraction_duration,omitempty"`
	ContractionIntensity  *string    `db:"contraction_intensity" json:"contraction_intensity,omitempty"`
	UterineRestingTone    *string    `db:"uterine_resting_tone" json:"uterine_resting_tone,omitempty"`
	MVUs                  *int       `db:"mvus" json:"mvus,omitempty"`
	Interpretation        *string    `db:"interpretation" json:"interpretation,omitempty"`
	Category              *string    `db:"category" json:"category,omitempty"`
	RecorderID            *uuid.UUID `db:"recorder_id" json:"recorder_id,omitempty"`
	Note                  *string    `db:"note" json:"note,omitempty"`
}

// DeliveryRecord maps to the delivery_record table.
type DeliveryRecord struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	PregnancyID           uuid.UUID  `db:"pregnancy_id" json:"pregnancy_id"`
	LaborRecordID         *uuid.UUID `db:"labor_record_id" json:"labor_record_id,omitempty"`
	PatientID             uuid.UUID  `db:"patient_id" json:"patient_id"`
	DeliveryDatetime      time.Time  `db:"delivery_datetime" json:"delivery_datetime"`
	DeliveryMethod        string     `db:"delivery_method" json:"delivery_method"`
	DeliveryType          *string    `db:"delivery_type" json:"delivery_type,omitempty"`
	DeliveringProviderID  uuid.UUID  `db:"delivering_provider_id" json:"delivering_provider_id"`
	AssistantProviderID   *uuid.UUID `db:"assistant_provider_id" json:"assistant_provider_id,omitempty"`
	DeliveryLocationID    *uuid.UUID `db:"delivery_location_id" json:"delivery_location_id,omitempty"`
	BirthOrder            *int       `db:"birth_order" json:"birth_order,omitempty"`
	PlacentaDelivery      *string    `db:"placenta_delivery" json:"placenta_delivery,omitempty"`
	PlacentaDatetime      *time.Time `db:"placenta_datetime" json:"placenta_datetime,omitempty"`
	PlacentaIntact        *bool      `db:"placenta_intact" json:"placenta_intact,omitempty"`
	CordVessels           *int       `db:"cord_vessels" json:"cord_vessels,omitempty"`
	CordBloodCollected    *bool      `db:"cord_blood_collected" json:"cord_blood_collected,omitempty"`
	Episiotomy            *bool      `db:"episiotomy" json:"episiotomy,omitempty"`
	EpisiotomyType        *string    `db:"episiotomy_type" json:"episiotomy_type,omitempty"`
	LacerationDegree      *string    `db:"laceration_degree" json:"laceration_degree,omitempty"`
	RepairMethod          *string    `db:"repair_method" json:"repair_method,omitempty"`
	BloodLossML           *int       `db:"blood_loss_ml" json:"blood_loss_ml,omitempty"`
	Complications         *string    `db:"complications" json:"complications,omitempty"`
	Note                  *string    `db:"note" json:"note,omitempty"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

// NewbornRecord maps to the newborn_record table.
type NewbornRecord struct {
	ID                  uuid.UUID  `db:"id" json:"id"`
	DeliveryID          uuid.UUID  `db:"delivery_id" json:"delivery_id"`
	PatientID           *uuid.UUID `db:"patient_id" json:"patient_id,omitempty"`
	BirthDatetime       time.Time  `db:"birth_datetime" json:"birth_datetime"`
	Sex                 *string    `db:"sex" json:"sex,omitempty"`
	BirthWeightGrams    *int       `db:"birth_weight_grams" json:"birth_weight_grams,omitempty"`
	BirthLengthCM       *float64   `db:"birth_length_cm" json:"birth_length_cm,omitempty"`
	HeadCircumferenceCM *float64   `db:"head_circumference_cm" json:"head_circumference_cm,omitempty"`
	Apgar1Min           *int       `db:"apgar_1min" json:"apgar_1min,omitempty"`
	Apgar5Min           *int       `db:"apgar_5min" json:"apgar_5min,omitempty"`
	Apgar10Min          *int       `db:"apgar_10min" json:"apgar_10min,omitempty"`
	ResuscitationType   *string    `db:"resuscitation_type" json:"resuscitation_type,omitempty"`
	GestationalAgeWeeks *int       `db:"gestational_age_weeks" json:"gestational_age_weeks,omitempty"`
	GestationalAgeDays  *int       `db:"gestational_age_days" json:"gestational_age_days,omitempty"`
	BirthStatus         *string    `db:"birth_status" json:"birth_status,omitempty"`
	NICUAdmission       *bool      `db:"nicu_admission" json:"nicu_admission,omitempty"`
	NICUReason          *string    `db:"nicu_reason" json:"nicu_reason,omitempty"`
	VitaminKGiven       *bool      `db:"vitamin_k_given" json:"vitamin_k_given,omitempty"`
	EyeProphylaxisGiven *bool      `db:"eye_prophylaxis_given" json:"eye_prophylaxis_given,omitempty"`
	HepatitisBGiven     *bool      `db:"hepatitis_b_given" json:"hepatitis_b_given,omitempty"`
	NewbornScreening    *string    `db:"newborn_screening" json:"newborn_screening,omitempty"`
	FeedingMethod       *string    `db:"feeding_method" json:"feeding_method,omitempty"`
	Note                *string    `db:"note" json:"note,omitempty"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}

// PostpartumRecord maps to the postpartum_record table.
type PostpartumRecord struct {
	ID                     uuid.UUID  `db:"id" json:"id"`
	PregnancyID            uuid.UUID  `db:"pregnancy_id" json:"pregnancy_id"`
	PatientID              uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID            *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	VisitDate              time.Time  `db:"visit_date" json:"visit_date"`
	DaysPostpartum         *int       `db:"days_postpartum" json:"days_postpartum,omitempty"`
	WeeksPostpartum        *int       `db:"weeks_postpartum" json:"weeks_postpartum,omitempty"`
	UterineInvolution      *string    `db:"uterine_involution" json:"uterine_involution,omitempty"`
	LochiaType             *string    `db:"lochia_type" json:"lochia_type,omitempty"`
	LochiaAmount           *string    `db:"lochia_amount" json:"lochia_amount,omitempty"`
	PerineumStatus         *string    `db:"perineum_status" json:"perineum_status,omitempty"`
	IncisionStatus         *string    `db:"incision_status" json:"incision_status,omitempty"`
	BreastStatus           *string    `db:"breast_status" json:"breast_status,omitempty"`
	BreastfeedingStatus    *string    `db:"breastfeeding_status" json:"breastfeeding_status,omitempty"`
	ContraceptionPlan      *string    `db:"contraception_plan" json:"contraception_plan,omitempty"`
	MoodScreeningScore     *int       `db:"mood_screening_score" json:"mood_screening_score,omitempty"`
	MoodScreeningTool      *string    `db:"mood_screening_tool" json:"mood_screening_tool,omitempty"`
	DepressionRisk         *string    `db:"depression_risk" json:"depression_risk,omitempty"`
	BloodPressureSystolic  *int       `db:"blood_pressure_systolic" json:"blood_pressure_systolic,omitempty"`
	BloodPressureDiastolic *int       `db:"blood_pressure_diastolic" json:"blood_pressure_diastolic,omitempty"`
	Weight                 *float64   `db:"weight" json:"weight,omitempty"`
	ProviderID             *uuid.UUID `db:"provider_id" json:"provider_id,omitempty"`
	Note                   *string    `db:"note" json:"note,omitempty"`
	CreatedAt              time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at" json:"updated_at"`
}
