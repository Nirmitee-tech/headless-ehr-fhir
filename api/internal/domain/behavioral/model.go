package behavioral

import (
	"time"

	"github.com/google/uuid"
)

// PsychiatricAssessment maps to the psychiatric_assessment table.
type PsychiatricAssessment struct {
	ID                   uuid.UUID  `db:"id" json:"id"`
	PatientID            uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID          uuid.UUID  `db:"encounter_id" json:"encounter_id"`
	AssessorID           uuid.UUID  `db:"assessor_id" json:"assessor_id"`
	AssessmentDate       time.Time  `db:"assessment_date" json:"assessment_date"`
	ChiefComplaint       *string    `db:"chief_complaint" json:"chief_complaint,omitempty"`
	HistoryPresentIllness *string   `db:"history_present_illness" json:"history_present_illness,omitempty"`
	PsychiatricHistory   *string    `db:"psychiatric_history" json:"psychiatric_history,omitempty"`
	SubstanceUseHistory  *string    `db:"substance_use_history" json:"substance_use_history,omitempty"`
	FamilyPsychHistory   *string    `db:"family_psych_history" json:"family_psych_history,omitempty"`
	MentalStatusExam     *string    `db:"mental_status_exam" json:"mental_status_exam,omitempty"`
	Appearance           *string    `db:"appearance" json:"appearance,omitempty"`
	Behavior             *string    `db:"behavior" json:"behavior,omitempty"`
	Speech               *string    `db:"speech" json:"speech,omitempty"`
	Mood                 *string    `db:"mood" json:"mood,omitempty"`
	Affect               *string    `db:"affect" json:"affect,omitempty"`
	ThoughtProcess       *string    `db:"thought_process" json:"thought_process,omitempty"`
	ThoughtContent       *string    `db:"thought_content" json:"thought_content,omitempty"`
	Perceptions          *string    `db:"perceptions" json:"perceptions,omitempty"`
	Cognition            *string    `db:"cognition" json:"cognition,omitempty"`
	Insight              *string    `db:"insight" json:"insight,omitempty"`
	Judgment             *string    `db:"judgment" json:"judgment,omitempty"`
	RiskAssessment       *string    `db:"risk_assessment" json:"risk_assessment,omitempty"`
	SuicideRiskLevel     *string    `db:"suicide_risk_level" json:"suicide_risk_level,omitempty"`
	HomicideRiskLevel    *string    `db:"homicide_risk_level" json:"homicide_risk_level,omitempty"`
	DiagnosisCode        *string    `db:"diagnosis_code" json:"diagnosis_code,omitempty"`
	DiagnosisDisplay     *string    `db:"diagnosis_display" json:"diagnosis_display,omitempty"`
	DiagnosisSystem      *string    `db:"diagnosis_system" json:"diagnosis_system,omitempty"`
	Formulation          *string    `db:"formulation" json:"formulation,omitempty"`
	TreatmentPlan        *string    `db:"treatment_plan" json:"treatment_plan,omitempty"`
	Disposition          *string    `db:"disposition" json:"disposition,omitempty"`
	Note                 *string    `db:"note" json:"note,omitempty"`
	CreatedAt            time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updated_at"`
}

// SafetyPlan maps to the safety_plan table.
type SafetyPlan struct {
	ID                     uuid.UUID  `db:"id" json:"id"`
	PatientID              uuid.UUID  `db:"patient_id" json:"patient_id"`
	CreatedByID            uuid.UUID  `db:"created_by_id" json:"created_by_id"`
	Status                 string     `db:"status" json:"status"`
	PlanDate               time.Time  `db:"plan_date" json:"plan_date"`
	WarningSigns           *string    `db:"warning_signs" json:"warning_signs,omitempty"`
	CopingStrategies       *string    `db:"coping_strategies" json:"coping_strategies,omitempty"`
	SocialDistractions     *string    `db:"social_distractions" json:"social_distractions,omitempty"`
	PeopleToContact        *string    `db:"people_to_contact" json:"people_to_contact,omitempty"`
	ProfessionalsToContact *string    `db:"professionals_to_contact" json:"professionals_to_contact,omitempty"`
	EmergencyContacts      *string    `db:"emergency_contacts" json:"emergency_contacts,omitempty"`
	MeansRestriction       *string    `db:"means_restriction" json:"means_restriction,omitempty"`
	ReasonsForLiving       *string    `db:"reasons_for_living" json:"reasons_for_living,omitempty"`
	PatientSignature       *bool      `db:"patient_signature" json:"patient_signature,omitempty"`
	ProviderSignature      *bool      `db:"provider_signature" json:"provider_signature,omitempty"`
	ReviewDate             *time.Time `db:"review_date" json:"review_date,omitempty"`
	Note                   *string    `db:"note" json:"note,omitempty"`
	CreatedAt              time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at" json:"updated_at"`
}

// LegalHold maps to the legal_hold table.
type LegalHold struct {
	ID                      uuid.UUID  `db:"id" json:"id"`
	PatientID               uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID             *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	InitiatedByID           uuid.UUID  `db:"initiated_by_id" json:"initiated_by_id"`
	Status                  string     `db:"status" json:"status"`
	HoldType                string     `db:"hold_type" json:"hold_type"`
	AuthorityStatute        *string    `db:"authority_statute" json:"authority_statute,omitempty"`
	StartDatetime           time.Time  `db:"start_datetime" json:"start_datetime"`
	EndDatetime             *time.Time `db:"end_datetime" json:"end_datetime,omitempty"`
	DurationHours           *int       `db:"duration_hours" json:"duration_hours,omitempty"`
	Reason                  string     `db:"reason" json:"reason"`
	CriteriaMet             *string    `db:"criteria_met" json:"criteria_met,omitempty"`
	CertifyingPhysicianID   *uuid.UUID `db:"certifying_physician_id" json:"certifying_physician_id,omitempty"`
	CertificationDatetime   *time.Time `db:"certification_datetime" json:"certification_datetime,omitempty"`
	CourtHearingDate        *time.Time `db:"court_hearing_date" json:"court_hearing_date,omitempty"`
	CourtOrderNumber        *string    `db:"court_order_number" json:"court_order_number,omitempty"`
	LegalCounselNotified    *bool      `db:"legal_counsel_notified" json:"legal_counsel_notified,omitempty"`
	PatientRightsGiven      *bool      `db:"patient_rights_given" json:"patient_rights_given,omitempty"`
	ReleaseReason           *string    `db:"release_reason" json:"release_reason,omitempty"`
	ReleaseAuthorizedByID   *uuid.UUID `db:"release_authorized_by_id" json:"release_authorized_by_id,omitempty"`
	Note                    *string    `db:"note" json:"note,omitempty"`
	CreatedAt               time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time  `db:"updated_at" json:"updated_at"`
}

// SeclusionRestraintEvent maps to the seclusion_restraint_event table.
type SeclusionRestraintEvent struct {
	ID                     uuid.UUID  `db:"id" json:"id"`
	PatientID              uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID            *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	OrderedByID            uuid.UUID  `db:"ordered_by_id" json:"ordered_by_id"`
	EventType              string     `db:"event_type" json:"event_type"`
	RestraintType          *string    `db:"restraint_type" json:"restraint_type,omitempty"`
	StartDatetime          time.Time  `db:"start_datetime" json:"start_datetime"`
	EndDatetime            *time.Time `db:"end_datetime" json:"end_datetime,omitempty"`
	Reason                 string     `db:"reason" json:"reason"`
	BehaviorDescription    *string    `db:"behavior_description" json:"behavior_description,omitempty"`
	AlternativesAttempted  *string    `db:"alternatives_attempted" json:"alternatives_attempted,omitempty"`
	MonitoringFrequencyMin *int       `db:"monitoring_frequency_min" json:"monitoring_frequency_min,omitempty"`
	LastMonitoringCheck    *time.Time `db:"last_monitoring_check" json:"last_monitoring_check,omitempty"`
	PatientConditionDuring *string    `db:"patient_condition_during" json:"patient_condition_during,omitempty"`
	InjuriesNoted          *string    `db:"injuries_noted" json:"injuries_noted,omitempty"`
	NutritionOffered       *bool      `db:"nutrition_offered" json:"nutrition_offered,omitempty"`
	ToiletingOffered       *bool      `db:"toileting_offered" json:"toileting_offered,omitempty"`
	DiscontinuedByID       *uuid.UUID `db:"discontinued_by_id" json:"discontinued_by_id,omitempty"`
	DiscontinuationReason  *string    `db:"discontinuation_reason" json:"discontinuation_reason,omitempty"`
	DebriefCompleted       *bool      `db:"debrief_completed" json:"debrief_completed,omitempty"`
	DebriefNotes           *string    `db:"debrief_notes" json:"debrief_notes,omitempty"`
	Note                   *string    `db:"note" json:"note,omitempty"`
	CreatedAt              time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at" json:"updated_at"`
}

// GroupTherapySession maps to the group_therapy_session table.
type GroupTherapySession struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	SessionName       string     `db:"session_name" json:"session_name"`
	SessionType       *string    `db:"session_type" json:"session_type,omitempty"`
	FacilitatorID     uuid.UUID  `db:"facilitator_id" json:"facilitator_id"`
	CoFacilitatorID   *uuid.UUID `db:"co_facilitator_id" json:"co_facilitator_id,omitempty"`
	Status            string     `db:"status" json:"status"`
	ScheduledDatetime time.Time  `db:"scheduled_datetime" json:"scheduled_datetime"`
	ActualStart       *time.Time `db:"actual_start" json:"actual_start,omitempty"`
	ActualEnd         *time.Time `db:"actual_end" json:"actual_end,omitempty"`
	Location          *string    `db:"location" json:"location,omitempty"`
	MaxParticipants   *int       `db:"max_participants" json:"max_participants,omitempty"`
	Topic             *string    `db:"topic" json:"topic,omitempty"`
	SessionGoals      *string    `db:"session_goals" json:"session_goals,omitempty"`
	SessionNotes      *string    `db:"session_notes" json:"session_notes,omitempty"`
	MaterialsUsed     *string    `db:"materials_used" json:"materials_used,omitempty"`
	Note              *string    `db:"note" json:"note,omitempty"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

// GroupTherapyAttendance maps to the group_therapy_attendance table.
type GroupTherapyAttendance struct {
	ID                 uuid.UUID `db:"id" json:"id"`
	SessionID          uuid.UUID `db:"session_id" json:"session_id"`
	PatientID          uuid.UUID `db:"patient_id" json:"patient_id"`
	AttendanceStatus   string    `db:"attendance_status" json:"attendance_status"`
	ParticipationLevel *string   `db:"participation_level" json:"participation_level,omitempty"`
	BehaviorNotes      *string   `db:"behavior_notes" json:"behavior_notes,omitempty"`
	MoodBefore         *string   `db:"mood_before" json:"mood_before,omitempty"`
	MoodAfter          *string   `db:"mood_after" json:"mood_after,omitempty"`
	Note               *string   `db:"note" json:"note,omitempty"`
}
