package researchdefinition

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// ResearchDefinition maps to the research_definition table (FHIR ResearchDefinition resource).
type ResearchDefinition struct {
	ID                  uuid.UUID  `db:"id" json:"id"`
	FHIRID              string     `db:"fhir_id" json:"fhir_id"`
	Status              string     `db:"status" json:"status"`
	URL                 *string    `db:"url" json:"url,omitempty"`
	Name                *string    `db:"name" json:"name,omitempty"`
	Title               *string    `db:"title" json:"title,omitempty"`
	Description         *string    `db:"description" json:"description,omitempty"`
	Publisher           *string    `db:"publisher" json:"publisher,omitempty"`
	Date                *time.Time `db:"date" json:"date,omitempty"`
	PopulationReference *string    `db:"population_reference" json:"population_reference,omitempty"`
	ExposureReference   *string    `db:"exposure_reference" json:"exposure_reference,omitempty"`
	OutcomeReference    *string    `db:"outcome_reference" json:"outcome_reference,omitempty"`
	VersionID           int        `db:"version_id" json:"version_id"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}

func (e *ResearchDefinition) GetVersionID() int  { return e.VersionID }
func (e *ResearchDefinition) SetVersionID(v int) { e.VersionID = v }

func (e *ResearchDefinition) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ResearchDefinition",
		"id":           e.FHIRID,
		"status":       e.Status,
		"meta":         fhir.Meta{LastUpdated: e.UpdatedAt},
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
	if e.ExposureReference != nil {
		result["exposure"] = fhir.Reference{Reference: *e.ExposureReference}
	}
	if e.OutcomeReference != nil {
		result["outcome"] = fhir.Reference{Reference: *e.OutcomeReference}
	}
	return result
}
