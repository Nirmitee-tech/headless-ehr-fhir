package structuredefinition

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// StructureDefinition maps to the structure_definition table (FHIR StructureDefinition resource).
type StructureDefinition struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	FHIRID         string     `db:"fhir_id" json:"fhir_id"`
	Status         string     `db:"status" json:"status"`
	URL            string     `db:"url" json:"url"`
	Name           string     `db:"name" json:"name"`
	Title          *string    `db:"title" json:"title,omitempty"`
	Description    *string    `db:"description" json:"description,omitempty"`
	Publisher      *string    `db:"publisher" json:"publisher,omitempty"`
	Date           *time.Time `db:"date" json:"date,omitempty"`
	Kind           string     `db:"kind" json:"kind"`
	Abstract       bool       `db:"abstract" json:"abstract"`
	Type           string     `db:"type" json:"type"`
	BaseDefinition *string    `db:"base_definition" json:"base_definition,omitempty"`
	Derivation     *string    `db:"derivation" json:"derivation,omitempty"`
	ContextType    *string    `db:"context_type" json:"context_type,omitempty"`
	VersionID      int        `db:"version_id" json:"version_id"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}

func (s *StructureDefinition) GetVersionID() int  { return s.VersionID }
func (s *StructureDefinition) SetVersionID(v int) { s.VersionID = v }

func (s *StructureDefinition) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "StructureDefinition",
		"id":           s.FHIRID,
		"url":          s.URL,
		"name":         s.Name,
		"status":       s.Status,
		"kind":         s.Kind,
		"abstract":     s.Abstract,
		"type":         s.Type,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", s.VersionID),
			LastUpdated: s.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/StructureDefinition"},
		},
	}
	if s.BaseDefinition != nil {
		result["baseDefinition"] = *s.BaseDefinition
	}
	if s.Derivation != nil {
		result["derivation"] = *s.Derivation
	}
	if s.Title != nil {
		result["title"] = *s.Title
	}
	if s.Description != nil {
		result["description"] = *s.Description
	}
	if s.Publisher != nil {
		result["publisher"] = *s.Publisher
	}
	if s.Date != nil {
		result["date"] = s.Date.Format("2006-01-02")
	}
	return result
}
