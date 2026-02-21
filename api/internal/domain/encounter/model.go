package encounter

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Encounter maps to the encounter table.
type Encounter struct {
	ID                       uuid.UUID  `db:"id" json:"id"`
	FHIRID                   string     `db:"fhir_id" json:"fhir_id"`
	Status                   string     `db:"status" json:"status"`
	ClassCode                string     `db:"class_code" json:"class_code"`
	ClassDisplay             *string    `db:"class_display" json:"class_display,omitempty"`
	TypeCode                 *string    `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay              *string    `db:"type_display" json:"type_display,omitempty"`
	ServiceTypeCode          *string    `db:"service_type_code" json:"service_type_code,omitempty"`
	ServiceTypeDisplay       *string    `db:"service_type_display" json:"service_type_display,omitempty"`
	PriorityCode             *string    `db:"priority_code" json:"priority_code,omitempty"`
	PatientID                uuid.UUID  `db:"patient_id" json:"patient_id"`
	PrimaryPractitionerID    *uuid.UUID `db:"primary_practitioner_id" json:"primary_practitioner_id,omitempty"`
	ServiceProviderID        *uuid.UUID `db:"service_provider_id" json:"service_provider_id,omitempty"`
	DepartmentID             *uuid.UUID `db:"department_id" json:"department_id,omitempty"`
	PeriodStart              time.Time  `db:"period_start" json:"period_start"`
	PeriodEnd                *time.Time `db:"period_end" json:"period_end,omitempty"`
	LengthMinutes            *int       `db:"length_minutes" json:"length_minutes,omitempty"`
	LocationID               *uuid.UUID `db:"location_id" json:"location_id,omitempty"`
	BedID                    *uuid.UUID `db:"bed_id" json:"bed_id,omitempty"`
	AdmitSourceCode          *string    `db:"admit_source_code" json:"admit_source_code,omitempty"`
	AdmitSourceDisplay       *string    `db:"admit_source_display" json:"admit_source_display,omitempty"`
	DischargeDispositionCode *string    `db:"discharge_disposition_code" json:"discharge_disposition_code,omitempty"`
	DischargeDispositionDisp *string    `db:"discharge_disposition_display" json:"discharge_disposition_display,omitempty"`
	ReAdmission              bool       `db:"re_admission" json:"re_admission"`
	IsTelehealth             bool       `db:"is_telehealth" json:"is_telehealth"`
	TelehealthPlatform       *string    `db:"telehealth_platform" json:"telehealth_platform,omitempty"`
	ReasonText               *string    `db:"reason_text" json:"reason_text,omitempty"`
	DRGCode                  *string    `db:"drg_code" json:"drg_code,omitempty"`
	DRGType                  *string    `db:"drg_type" json:"drg_type,omitempty"`
	VersionID                int        `db:"version_id" json:"version_id"`
	CreatedAt                time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt                time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (e *Encounter) GetVersionID() int { return e.VersionID }

// SetVersionID sets the current version.
func (e *Encounter) SetVersionID(v int) { e.VersionID = v }

func (e *Encounter) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Encounter",
		"id":           e.FHIRID,
		"status":       e.Status,
		"class": fhir.Coding{
			System:  "http://terminology.hl7.org/CodeSystem/v3-ActCode",
			Code:    e.ClassCode,
			Display: strPtrVal(e.ClassDisplay),
		},
		"subject": fhir.Reference{
			Reference: fhir.FormatReference("Patient", e.PatientID.String()),
		},
		"period": fhir.Period{
			Start: &e.PeriodStart,
			End:   e.PeriodEnd,
		},
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", e.VersionID),
			LastUpdated: e.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/us/core/StructureDefinition/us-core-encounter"},
		},
	}

	if e.TypeCode != nil {
		result["type"] = []fhir.CodeableConcept{
			{Coding: []fhir.Coding{{Code: *e.TypeCode, Display: strPtrVal(e.TypeDisplay)}}},
		}
	}

	if e.ServiceTypeCode != nil {
		result["serviceType"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *e.ServiceTypeCode, Display: strPtrVal(e.ServiceTypeDisplay)}},
		}
	}

	if e.PriorityCode != nil {
		result["priority"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System: "http://terminology.hl7.org/CodeSystem/v3-ActPriority",
				Code:   *e.PriorityCode,
			}},
		}
	}

	if e.PrimaryPractitionerID != nil {
		result["participant"] = []map[string]interface{}{
			{
				"type": []fhir.CodeableConcept{
					{Coding: []fhir.Coding{{Code: "ATND", Display: "attender"}}},
				},
				"individual": fhir.Reference{
					Reference: fhir.FormatReference("Practitioner", e.PrimaryPractitionerID.String()),
				},
			},
		}
	}

	if e.ServiceProviderID != nil {
		result["serviceProvider"] = fhir.Reference{
			Reference: fhir.FormatReference("Organization", e.ServiceProviderID.String()),
		}
	}

	if e.LocationID != nil {
		result["location"] = []map[string]interface{}{
			{
				"location": fhir.Reference{
					Reference: fhir.FormatReference("Location", e.LocationID.String()),
				},
				"status": "active",
			},
		}
	}

	if e.ReasonText != nil {
		result["reasonCode"] = []fhir.CodeableConcept{
			{Text: *e.ReasonText},
		}
	}

	// Hospitalization details
	if e.AdmitSourceCode != nil || e.DischargeDispositionCode != nil {
		hosp := map[string]interface{}{}
		if e.AdmitSourceCode != nil {
			hosp["admitSource"] = fhir.CodeableConcept{
				Coding: []fhir.Coding{{Code: *e.AdmitSourceCode, Display: strPtrVal(e.AdmitSourceDisplay)}},
			}
		}
		if e.DischargeDispositionCode != nil {
			hosp["dischargeDisposition"] = fhir.CodeableConcept{
				Coding: []fhir.Coding{{Code: *e.DischargeDispositionCode, Display: strPtrVal(e.DischargeDispositionDisp)}},
			}
		}
		if e.ReAdmission {
			hosp["reAdmission"] = fhir.CodeableConcept{
				Coding: []fhir.Coding{{Code: "R"}},
			}
		}
		result["hospitalization"] = hosp
	}

	return result
}

// EncounterParticipant maps to the encounter_participant table.
type EncounterParticipant struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	EncounterID    uuid.UUID  `db:"encounter_id" json:"encounter_id"`
	PractitionerID uuid.UUID  `db:"practitioner_id" json:"practitioner_id"`
	TypeCode       string     `db:"type_code" json:"type_code"`
	TypeDisplay    *string    `db:"type_display" json:"type_display,omitempty"`
	PeriodStart    *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd      *time.Time `db:"period_end" json:"period_end,omitempty"`
}

// EncounterDiagnosis maps to the encounter_diagnosis table.
type EncounterDiagnosis struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	EncounterID uuid.UUID  `db:"encounter_id" json:"encounter_id"`
	ConditionID *uuid.UUID `db:"condition_id" json:"condition_id,omitempty"`
	UseCode     *string    `db:"use_code" json:"use_code,omitempty"`
	Rank        *int       `db:"rank" json:"rank,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}

// EncounterStatusHistory maps to the encounter_status_history table.
type EncounterStatusHistory struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	EncounterID uuid.UUID  `db:"encounter_id" json:"encounter_id"`
	Status      string     `db:"status" json:"status"`
	PeriodStart time.Time  `db:"period_start" json:"period_start"`
	PeriodEnd   *time.Time `db:"period_end" json:"period_end,omitempty"`
}

func strPtrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
