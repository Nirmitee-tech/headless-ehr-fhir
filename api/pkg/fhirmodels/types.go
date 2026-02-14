package fhirmodels

// Common FHIR value set constants used across the application.

// EncounterStatus values per FHIR R4.
const (
	EncounterStatusPlanned         = "planned"
	EncounterStatusArrived         = "arrived"
	EncounterStatusTriaged         = "triaged"
	EncounterStatusInProgress      = "in-progress"
	EncounterStatusOnLeave         = "onleave"
	EncounterStatusFinished        = "finished"
	EncounterStatusCancelled       = "cancelled"
	EncounterStatusEnteredInError  = "entered-in-error"
)

// EncounterClass codes per FHIR R4 v3-ActCode.
const (
	EncounterClassAmbulatory      = "AMB"
	EncounterClassEmergency       = "EMER"
	EncounterClassInpatient       = "IMP"
	EncounterClassShortStay       = "SS"
	EncounterClassVirtual         = "VR"
	EncounterClassHomeHealth      = "HH"
	EncounterClassObstetric       = "OBSENC"
	EncounterClassAcute           = "ACUTE"
	EncounterClassNonAcute        = "NONAC"
	EncounterClassPreAdmission    = "PRENC"
	EncounterClassField           = "FLD"
)

// ParticipantType codes.
const (
	ParticipantAttender   = "ATND"
	ParticipantAdmitter   = "ADM"
	ParticipantConsultant = "CON"
	ParticipantReferrer   = "REF"
	ParticipantSecondary  = "SPRF"
	ParticipantPrimary    = "PPRF"
	ParticipantDischarger = "DIS"
)

// ObservationCategory codes.
const (
	ObsCategoryVitalSigns     = "vital-signs"
	ObsCategoryLaboratory     = "laboratory"
	ObsCategoryImaging        = "imaging"
	ObsCategorySocialHistory  = "social-history"
	ObsCategorySurvey         = "survey"
	ObsCategoryExam           = "exam"
	ObsCategoryProcedure      = "procedure"
	ObsCategoryActivity       = "activity"
	ObsCategoryTherapy        = "therapy"
)

// ConditionClinicalStatus codes.
const (
	ConditionActive     = "active"
	ConditionRecurrence = "recurrence"
	ConditionRelapse    = "relapse"
	ConditionInactive   = "inactive"
	ConditionRemission  = "remission"
	ConditionResolved   = "resolved"
)

// AdministrativeGender codes.
const (
	GenderMale    = "male"
	GenderFemale  = "female"
	GenderOther   = "other"
	GenderUnknown = "unknown"
)
