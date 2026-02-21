package riskevidencesynthesis

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// RiskEvidenceSynthesis maps to the risk_evidence_synthesis table (FHIR RiskEvidenceSynthesis resource).
type RiskEvidenceSynthesis struct {
	ID                       uuid.UUID  `db:"id" json:"id"`
	FHIRID                   string     `db:"fhir_id" json:"fhir_id"`
	Status                   string     `db:"status" json:"status"`
	URL                      *string    `db:"url" json:"url,omitempty"`
	Name                     *string    `db:"name" json:"name,omitempty"`
	Title                    *string    `db:"title" json:"title,omitempty"`
	Description              *string    `db:"description" json:"description,omitempty"`
	Publisher                *string    `db:"publisher" json:"publisher,omitempty"`
	Date                     *time.Time `db:"date" json:"date,omitempty"`
	PopulationReference      *string    `db:"population_reference" json:"population_reference,omitempty"`
	OutcomeReference         *string    `db:"outcome_reference" json:"outcome_reference,omitempty"`
	SampleSizeDescription    *string    `db:"sample_size_description" json:"sample_size_description,omitempty"`
	RiskEstimateDescription  *string    `db:"risk_estimate_description" json:"risk_estimate_description,omitempty"`
	RiskEstimateValue        *float64   `db:"risk_estimate_value" json:"risk_estimate_value,omitempty"`
	VersionID                int        `db:"version_id" json:"version_id"`
	CreatedAt                time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt                time.Time  `db:"updated_at" json:"updated_at"`
}

func (e *RiskEvidenceSynthesis) GetVersionID() int  { return e.VersionID }
func (e *RiskEvidenceSynthesis) SetVersionID(v int) { e.VersionID = v }

func (e *RiskEvidenceSynthesis) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "RiskEvidenceSynthesis",
		"id":           e.FHIRID,
		"status":       e.Status,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", e.VersionID),
			LastUpdated: e.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/RiskEvidenceSynthesis"},
		},
	}
	if e.URL != nil {
		result["url"] = *e.URL
	}
	if e.Name != nil {
		result["name"] = *e.Name
	}
	if e.Title != nil {
		result["title"] = *e.Title
	}
	if e.Description != nil {
		result["description"] = *e.Description
	}
	if e.Publisher != nil {
		result["publisher"] = *e.Publisher
	}
	if e.Date != nil {
		result["date"] = e.Date.Format("2006-01-02")
	}
	if e.PopulationReference != nil {
		result["population"] = fhir.Reference{Reference: *e.PopulationReference}
	}
	if e.OutcomeReference != nil {
		result["outcome"] = fhir.Reference{Reference: *e.OutcomeReference}
	}
	if e.SampleSizeDescription != nil {
		result["sampleSize"] = map[string]interface{}{
			"description": *e.SampleSizeDescription,
		}
	}
	if e.RiskEstimateDescription != nil || e.RiskEstimateValue != nil {
		riskEstimate := map[string]interface{}{}
		if e.RiskEstimateDescription != nil {
			riskEstimate["description"] = *e.RiskEstimateDescription
		}
		if e.RiskEstimateValue != nil {
			riskEstimate["value"] = *e.RiskEstimateValue
		}
		result["riskEstimate"] = riskEstimate
	}
	return result
}
