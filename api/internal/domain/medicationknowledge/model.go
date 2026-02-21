package medicationknowledge

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// MedicationKnowledge maps to the medication_knowledge table (FHIR MedicationKnowledge resource).
type MedicationKnowledge struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	FHIRID          string     `db:"fhir_id" json:"fhir_id"`
	Status          string     `db:"status" json:"status"`
	CodeCode        *string    `db:"code_code" json:"code_code,omitempty"`
	CodeSystem      *string    `db:"code_system" json:"code_system,omitempty"`
	CodeDisplay     *string    `db:"code_display" json:"code_display,omitempty"`
	ManufacturerID  *uuid.UUID `db:"manufacturer_id" json:"manufacturer_id,omitempty"`
	DoseFormCode    *string    `db:"dose_form_code" json:"dose_form_code,omitempty"`
	DoseFormDisplay *string    `db:"dose_form_display" json:"dose_form_display,omitempty"`
	AmountValue     *float64   `db:"amount_value" json:"amount_value,omitempty"`
	AmountUnit      *string    `db:"amount_unit" json:"amount_unit,omitempty"`
	Synonym         *string    `db:"synonym" json:"synonym,omitempty"`
	Description     *string    `db:"description" json:"description,omitempty"`
	VersionID       int        `db:"version_id" json:"version_id"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

func (m *MedicationKnowledge) GetVersionID() int  { return m.VersionID }
func (m *MedicationKnowledge) SetVersionID(v int)  { m.VersionID = v }

func (m *MedicationKnowledge) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "MedicationKnowledge",
		"id":           m.FHIRID,
		"status":       m.Status,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", m.VersionID),
			LastUpdated: m.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/MedicationKnowledge"},
		},
	}
	if m.CodeCode != nil {
		result["code"] = fhir.CodeableConcept{Coding: []fhir.Coding{{
			System:  strVal(m.CodeSystem),
			Code:    *m.CodeCode,
			Display: strVal(m.CodeDisplay),
		}}}
	}
	if m.ManufacturerID != nil {
		result["manufacturer"] = fhir.Reference{Reference: fhir.FormatReference("Organization", m.ManufacturerID.String())}
	}
	if m.DoseFormCode != nil {
		result["doseForm"] = fhir.CodeableConcept{Coding: []fhir.Coding{{
			Code:    *m.DoseFormCode,
			Display: strVal(m.DoseFormDisplay),
		}}}
	}
	if m.AmountValue != nil {
		q := map[string]interface{}{"value": *m.AmountValue}
		if m.AmountUnit != nil {
			q["unit"] = *m.AmountUnit
		}
		result["amount"] = q
	}
	if m.Synonym != nil {
		result["synonym"] = []string{*m.Synonym}
	}
	if m.Description != nil {
		result["description"] = *m.Description
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
