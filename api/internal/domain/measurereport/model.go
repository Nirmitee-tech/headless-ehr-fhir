package measurereport

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// MeasureReport maps to the measure_report table (FHIR MeasureReport resource).
type MeasureReport struct {
	ID                   uuid.UUID  `db:"id" json:"id"`
	FHIRID               string     `db:"fhir_id" json:"fhir_id"`
	Status               string     `db:"status" json:"status"`
	Type                 string     `db:"type" json:"type"`
	MeasureURL           *string    `db:"measure_url" json:"measure_url,omitempty"`
	SubjectPatientID     *uuid.UUID `db:"subject_patient_id" json:"subject_patient_id,omitempty"`
	Date                 *time.Time `db:"date" json:"date,omitempty"`
	ReporterOrgID        *uuid.UUID `db:"reporter_org_id" json:"reporter_org_id,omitempty"`
	PeriodStart          time.Time  `db:"period_start" json:"period_start"`
	PeriodEnd            time.Time  `db:"period_end" json:"period_end"`
	ImprovementNotation  *string    `db:"improvement_notation" json:"improvement_notation,omitempty"`
	GroupCode            *string    `db:"group_code" json:"group_code,omitempty"`
	GroupPopulationCode  *string    `db:"group_population_code" json:"group_population_code,omitempty"`
	GroupPopulationCount *int       `db:"group_population_count" json:"group_population_count,omitempty"`
	GroupMeasureScore    *float64   `db:"group_measure_score" json:"group_measure_score,omitempty"`
	VersionID            int        `db:"version_id" json:"version_id"`
	CreatedAt            time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (mr *MeasureReport) GetVersionID() int { return mr.VersionID }

// SetVersionID sets the current version.
func (mr *MeasureReport) SetVersionID(v int) { mr.VersionID = v }

func (mr *MeasureReport) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "MeasureReport",
		"id":           mr.FHIRID,
		"status":       mr.Status,
		"type":         mr.Type,
		"period": fhir.Period{
			Start: &mr.PeriodStart,
			End:   &mr.PeriodEnd,
		},
		"meta": fhir.Meta{LastUpdated: mr.UpdatedAt},
	}
	if mr.MeasureURL != nil {
		result["measure"] = *mr.MeasureURL
	}
	if mr.SubjectPatientID != nil {
		result["subject"] = fhir.Reference{Reference: fhir.FormatReference("Patient", mr.SubjectPatientID.String())}
	}
	if mr.Date != nil {
		result["date"] = mr.Date.Format("2006-01-02T15:04:05Z")
	}
	if mr.ReporterOrgID != nil {
		result["reporter"] = fhir.Reference{Reference: fhir.FormatReference("Organization", mr.ReporterOrgID.String())}
	}
	if mr.ImprovementNotation != nil {
		result["improvementNotation"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *mr.ImprovementNotation}}}
	}
	if mr.GroupCode != nil {
		group := map[string]interface{}{
			"code": fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *mr.GroupCode}}},
		}
		if mr.GroupPopulationCode != nil {
			pop := map[string]interface{}{
				"code": fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *mr.GroupPopulationCode}}},
			}
			if mr.GroupPopulationCount != nil {
				pop["count"] = *mr.GroupPopulationCount
			}
			group["population"] = []map[string]interface{}{pop}
		}
		if mr.GroupMeasureScore != nil {
			group["measureScore"] = map[string]interface{}{"value": *mr.GroupMeasureScore}
		}
		result["group"] = []map[string]interface{}{group}
	}
	return result
}
