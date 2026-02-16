package researchelementdefinition

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// ResearchElementDefinition maps to the research_element_definition table (FHIR ResearchElementDefinition resource).
type ResearchElementDefinition struct {
	ID                          uuid.UUID  `db:"id" json:"id"`
	FHIRID                      string     `db:"fhir_id" json:"fhir_id"`
	Status                      string     `db:"status" json:"status"`
	URL                         *string    `db:"url" json:"url,omitempty"`
	Name                        *string    `db:"name" json:"name,omitempty"`
	Title                       *string    `db:"title" json:"title,omitempty"`
	Description                 *string    `db:"description" json:"description,omitempty"`
	Publisher                   *string    `db:"publisher" json:"publisher,omitempty"`
	Date                        *time.Time `db:"date" json:"date,omitempty"`
	Type                        string     `db:"type" json:"type"`
	CharacteristicDefinitionRef *string    `db:"characteristic_definition_reference" json:"characteristic_definition_reference,omitempty"`
	VersionID                   int        `db:"version_id" json:"version_id"`
	CreatedAt                   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt                   time.Time  `db:"updated_at" json:"updated_at"`
}

func (e *ResearchElementDefinition) GetVersionID() int  { return e.VersionID }
func (e *ResearchElementDefinition) SetVersionID(v int) { e.VersionID = v }

func (e *ResearchElementDefinition) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ResearchElementDefinition",
		"id":           e.FHIRID,
		"status":       e.Status,
		"type":         e.Type,
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
	if e.CharacteristicDefinitionRef != nil {
		characteristic := map[string]interface{}{
			"definitionReference": fhir.Reference{Reference: *e.CharacteristicDefinitionRef},
		}
		result["characteristic"] = []interface{}{characteristic}
	}
	return result
}
