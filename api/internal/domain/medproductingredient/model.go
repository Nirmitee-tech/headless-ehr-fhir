package medproductingredient

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type MedicinalProductIngredient struct {
	ID                       uuid.UUID `db:"id" json:"id"`
	FHIRID                   string    `db:"fhir_id" json:"fhir_id"`
	RoleCode                 string    `db:"role_code" json:"role_code"`
	RoleDisplay              *string   `db:"role_display" json:"role_display,omitempty"`
	AllergenicIndicator      *bool     `db:"allergenic_indicator" json:"allergenic_indicator,omitempty"`
	SubstanceCode            *string   `db:"substance_code" json:"substance_code,omitempty"`
	SubstanceDisplay         *string   `db:"substance_display" json:"substance_display,omitempty"`
	StrengthNumeratorValue   *float64  `db:"strength_numerator_value" json:"strength_numerator_value,omitempty"`
	StrengthNumeratorUnit    *string   `db:"strength_numerator_unit" json:"strength_numerator_unit,omitempty"`
	StrengthDenominatorValue *float64  `db:"strength_denominator_value" json:"strength_denominator_value,omitempty"`
	StrengthDenominatorUnit  *string   `db:"strength_denominator_unit" json:"strength_denominator_unit,omitempty"`
	ManufacturerReference    *string   `db:"manufacturer_reference" json:"manufacturer_reference,omitempty"`
	VersionID                int       `db:"version_id" json:"version_id"`
	CreatedAt                time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                time.Time `db:"updated_at" json:"updated_at"`
}

func (m *MedicinalProductIngredient) GetVersionID() int  { return m.VersionID }
func (m *MedicinalProductIngredient) SetVersionID(v int) { m.VersionID = v }

func (m *MedicinalProductIngredient) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "MedicinalProductIngredient",
		"id":           m.FHIRID,
		"role":         fhir.CodeableConcept{Coding: []fhir.Coding{{Code: m.RoleCode, Display: strVal(m.RoleDisplay)}}},
		"meta":         fhir.Meta{LastUpdated: m.UpdatedAt},
	}
	if m.AllergenicIndicator != nil {
		result["allergenicIndicator"] = *m.AllergenicIndicator
	}
	if m.SubstanceCode != nil {
		substance := map[string]interface{}{
			"code": fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.SubstanceCode, Display: strVal(m.SubstanceDisplay)}}},
		}
		if m.StrengthNumeratorValue != nil {
			strength := map[string]interface{}{
				"presentation": map[string]interface{}{
					"numerator": map[string]interface{}{"value": *m.StrengthNumeratorValue, "unit": strVal(m.StrengthNumeratorUnit)},
				},
			}
			if m.StrengthDenominatorValue != nil {
				strength["presentation"].(map[string]interface{})["denominator"] = map[string]interface{}{
					"value": *m.StrengthDenominatorValue, "unit": strVal(m.StrengthDenominatorUnit),
				}
			}
			substance["strength"] = []interface{}{strength}
		}
		result["specifiedSubstance"] = []interface{}{substance}
	}
	if m.ManufacturerReference != nil {
		result["manufacturer"] = []fhir.Reference{{Reference: *m.ManufacturerReference}}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
