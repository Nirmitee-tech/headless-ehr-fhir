package bodystructure

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// BodyStructure maps to the body_structure table (FHIR BodyStructure resource).
type BodyStructure struct {
	ID                       uuid.UUID  `db:"id" json:"id"`
	FHIRID                   string     `db:"fhir_id" json:"fhir_id"`
	Active                   bool       `db:"active" json:"active"`
	MorphologyCode           *string    `db:"morphology_code" json:"morphology_code,omitempty"`
	MorphologyDisplay        *string    `db:"morphology_display" json:"morphology_display,omitempty"`
	MorphologySystem         *string    `db:"morphology_system" json:"morphology_system,omitempty"`
	LocationCode             *string    `db:"location_code" json:"location_code,omitempty"`
	LocationDisplay          *string    `db:"location_display" json:"location_display,omitempty"`
	LocationSystem           *string    `db:"location_system" json:"location_system,omitempty"`
	LocationQualifierCode    *string    `db:"location_qualifier_code" json:"location_qualifier_code,omitempty"`
	LocationQualifierDisplay *string    `db:"location_qualifier_display" json:"location_qualifier_display,omitempty"`
	Description              *string    `db:"description" json:"description,omitempty"`
	PatientID                uuid.UUID  `db:"patient_id" json:"patient_id"`
	VersionID                int        `db:"version_id" json:"version_id"`
	CreatedAt                time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt                time.Time  `db:"updated_at" json:"updated_at"`
}

func (b *BodyStructure) GetVersionID() int  { return b.VersionID }
func (b *BodyStructure) SetVersionID(v int) { b.VersionID = v }

func (b *BodyStructure) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "BodyStructure",
		"id":           b.FHIRID,
		"active":       b.Active,
		"patient":      fhir.Reference{Reference: fhir.FormatReference("Patient", b.PatientID.String())},
		"meta":         fhir.Meta{LastUpdated: b.UpdatedAt},
	}
	if b.MorphologyCode != nil {
		result["morphology"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *b.MorphologyCode, Display: strVal(b.MorphologyDisplay), System: strVal(b.MorphologySystem)}}}
	}
	if b.LocationCode != nil {
		result["location"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *b.LocationCode, Display: strVal(b.LocationDisplay), System: strVal(b.LocationSystem)}}}
	}
	if b.LocationQualifierCode != nil {
		result["locationQualifier"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *b.LocationQualifierCode, Display: strVal(b.LocationQualifierDisplay)}}}}
	}
	if b.Description != nil {
		result["description"] = *b.Description
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
