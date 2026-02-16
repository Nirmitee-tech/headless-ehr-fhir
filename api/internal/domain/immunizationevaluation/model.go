package immunizationevaluation

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type ImmunizationEvaluation struct {
	ID                       uuid.UUID  `db:"id" json:"id"`
	FHIRID                   string     `db:"fhir_id" json:"fhir_id"`
	Status                   string     `db:"status" json:"status"`
	PatientID                uuid.UUID  `db:"patient_id" json:"patient_id"`
	Date                     *time.Time `db:"date" json:"date,omitempty"`
	AuthorityReference       *string    `db:"authority_reference" json:"authority_reference,omitempty"`
	TargetDiseaseCode        string     `db:"target_disease_code" json:"target_disease_code"`
	TargetDiseaseDisplay     *string    `db:"target_disease_display" json:"target_disease_display,omitempty"`
	ImmunizationEventRef     string     `db:"immunization_event_reference" json:"immunization_event_reference"`
	DoseStatusCode           string     `db:"dose_status_code" json:"dose_status_code"`
	DoseStatusDisplay        *string    `db:"dose_status_display" json:"dose_status_display,omitempty"`
	DoseStatusReasonCode     *string    `db:"dose_status_reason_code" json:"dose_status_reason_code,omitempty"`
	DoseStatusReasonDisplay  *string    `db:"dose_status_reason_display" json:"dose_status_reason_display,omitempty"`
	Series                   *string    `db:"series" json:"series,omitempty"`
	DoseNumber               *string    `db:"dose_number" json:"dose_number,omitempty"`
	SeriesDoses              *string    `db:"series_doses" json:"series_doses,omitempty"`
	Description              *string    `db:"description" json:"description,omitempty"`
	VersionID                int        `db:"version_id" json:"version_id"`
	CreatedAt                time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt                time.Time  `db:"updated_at" json:"updated_at"`
}

func (ie *ImmunizationEvaluation) GetVersionID() int  { return ie.VersionID }
func (ie *ImmunizationEvaluation) SetVersionID(v int) { ie.VersionID = v }

func (ie *ImmunizationEvaluation) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ImmunizationEvaluation",
		"id":           ie.FHIRID,
		"status":       ie.Status,
		"patient":      fhir.Reference{Reference: fhir.FormatReference("Patient", ie.PatientID.String())},
		"targetDisease": fhir.CodeableConcept{Coding: []fhir.Coding{{Code: ie.TargetDiseaseCode, Display: strVal(ie.TargetDiseaseDisplay)}}},
		"immunizationEvent": fhir.Reference{Reference: ie.ImmunizationEventRef},
		"doseStatus":   fhir.CodeableConcept{Coding: []fhir.Coding{{Code: ie.DoseStatusCode, Display: strVal(ie.DoseStatusDisplay)}}},
		"meta":         fhir.Meta{LastUpdated: ie.UpdatedAt},
	}
	if ie.Date != nil {
		result["date"] = ie.Date.Format("2006-01-02T15:04:05Z")
	}
	if ie.AuthorityReference != nil {
		result["authority"] = fhir.Reference{Reference: *ie.AuthorityReference}
	}
	if ie.DoseStatusReasonCode != nil {
		result["doseStatusReason"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *ie.DoseStatusReasonCode, Display: strVal(ie.DoseStatusReasonDisplay)}}}}
	}
	if ie.Series != nil {
		result["series"] = *ie.Series
	}
	if ie.DoseNumber != nil {
		result["doseNumberString"] = *ie.DoseNumber
	}
	if ie.SeriesDoses != nil {
		result["seriesDosesString"] = *ie.SeriesDoses
	}
	if ie.Description != nil {
		result["description"] = *ie.Description
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
