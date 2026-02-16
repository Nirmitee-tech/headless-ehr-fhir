package medproductmanufactured

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type MedicinalProductManufactured struct {
	ID                           uuid.UUID `db:"id" json:"id"`
	FHIRID                       string    `db:"fhir_id" json:"fhir_id"`
	ManufacturedDoseFormCode     string    `db:"manufactured_dose_form_code" json:"manufactured_dose_form_code"`
	ManufacturedDoseFormDisplay  *string   `db:"manufactured_dose_form_display" json:"manufactured_dose_form_display,omitempty"`
	UnitOfPresentationCode       *string   `db:"unit_of_presentation_code" json:"unit_of_presentation_code,omitempty"`
	UnitOfPresentationDisplay    *string   `db:"unit_of_presentation_display" json:"unit_of_presentation_display,omitempty"`
	QuantityValue                *float64  `db:"quantity_value" json:"quantity_value,omitempty"`
	QuantityUnit                 *string   `db:"quantity_unit" json:"quantity_unit,omitempty"`
	ManufacturerReference        *string   `db:"manufacturer_reference" json:"manufacturer_reference,omitempty"`
	IngredientReference          *string   `db:"ingredient_reference" json:"ingredient_reference,omitempty"`
	VersionID                    int       `db:"version_id" json:"version_id"`
	CreatedAt                    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                    time.Time `db:"updated_at" json:"updated_at"`
}

func (m *MedicinalProductManufactured) GetVersionID() int  { return m.VersionID }
func (m *MedicinalProductManufactured) SetVersionID(v int) { m.VersionID = v }

func (m *MedicinalProductManufactured) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType":        "MedicinalProductManufactured",
		"id":                  m.FHIRID,
		"manufacturedDoseForm": fhir.CodeableConcept{Coding: []fhir.Coding{{Code: m.ManufacturedDoseFormCode, Display: strVal(m.ManufacturedDoseFormDisplay)}}},
		"meta":                fhir.Meta{LastUpdated: m.UpdatedAt},
	}
	if m.UnitOfPresentationCode != nil {
		result["unitOfPresentation"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.UnitOfPresentationCode, Display: strVal(m.UnitOfPresentationDisplay)}}}
	}
	if m.QuantityValue != nil {
		q := map[string]interface{}{"value": *m.QuantityValue}
		if m.QuantityUnit != nil {
			q["unit"] = *m.QuantityUnit
		}
		result["quantity"] = q
	}
	if m.ManufacturerReference != nil {
		result["manufacturer"] = []fhir.Reference{{Reference: *m.ManufacturerReference}}
	}
	if m.IngredientReference != nil {
		result["ingredient"] = []fhir.Reference{{Reference: *m.IngredientReference}}
	}
	return result
}

func strVal(s *string) string {
	if s == nil { return "" }
	return *s
}
