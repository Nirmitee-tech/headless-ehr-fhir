package substance

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
)

// Substance maps to the substance table (FHIR Substance resource).
type Substance struct {
	ID              string     `db:"id" json:"id"`
	FHIRID          string     `db:"fhir_id" json:"fhir_id"`
	Status          string     `db:"status" json:"status"`
	CategoryCode    *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay *string    `db:"category_display" json:"category_display,omitempty"`
	CodeCode        string     `db:"code_code" json:"code_code"`
	CodeDisplay     *string    `db:"code_display" json:"code_display,omitempty"`
	CodeSystem      *string    `db:"code_system" json:"code_system,omitempty"`
	Description     *string    `db:"description" json:"description,omitempty"`
	Expiry          *time.Time `db:"expiry" json:"expiry,omitempty"`
	QuantityValue   *float64   `db:"quantity_value" json:"quantity_value,omitempty"`
	QuantityUnit    *string    `db:"quantity_unit" json:"quantity_unit,omitempty"`
	VersionID       int        `db:"version_id" json:"version_id"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

func (s *Substance) GetVersionID() int  { return s.VersionID }
func (s *Substance) SetVersionID(v int) { s.VersionID = v }

func (s *Substance) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Substance",
		"id":           s.FHIRID,
		"status":       s.Status,
		"code":         fhir.CodeableConcept{Coding: []fhir.Coding{{Code: s.CodeCode, Display: strVal(s.CodeDisplay), System: strVal(s.CodeSystem)}}},
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", s.VersionID),
			LastUpdated: s.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Substance"},
		},
	}
	if s.CategoryCode != nil {
		result["category"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *s.CategoryCode, Display: strVal(s.CategoryDisplay)}}}}
	}
	if s.Description != nil {
		result["description"] = *s.Description
	}
	if s.QuantityValue != nil {
		q := map[string]interface{}{"value": *s.QuantityValue}
		if s.QuantityUnit != nil {
			q["unit"] = *s.QuantityUnit
		}
		result["quantity"] = q
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
