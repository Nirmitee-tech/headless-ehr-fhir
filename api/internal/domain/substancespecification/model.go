package substancespecification

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// SubstanceSpecification maps to the substance_specification table (FHIR SubstanceSpecification resource).
type SubstanceSpecification struct {
	ID                    uuid.UUID `db:"id" json:"id"`
	FHIRID                string    `db:"fhir_id" json:"fhir_id"`
	Status                *string   `db:"status" json:"status,omitempty"`
	TypeCode              *string   `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay           *string   `db:"type_display" json:"type_display,omitempty"`
	DomainCode            *string   `db:"domain_code" json:"domain_code,omitempty"`
	DomainDisplay         *string   `db:"domain_display" json:"domain_display,omitempty"`
	Description           *string   `db:"description" json:"description,omitempty"`
	SourceReference       *string   `db:"source_reference" json:"source_reference,omitempty"`
	Comment               *string   `db:"comment" json:"comment,omitempty"`
	MolecularWeightAmount *float64  `db:"molecular_weight_amount" json:"molecular_weight_amount,omitempty"`
	MolecularWeightUnit   *string   `db:"molecular_weight_unit" json:"molecular_weight_unit,omitempty"`
	VersionID             int       `db:"version_id" json:"version_id"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time `db:"updated_at" json:"updated_at"`
}

func (s *SubstanceSpecification) GetVersionID() int  { return s.VersionID }
func (s *SubstanceSpecification) SetVersionID(v int) { s.VersionID = v }

func (s *SubstanceSpecification) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "SubstanceSpecification",
		"id":           s.FHIRID,
		"meta":         fhir.Meta{LastUpdated: s.UpdatedAt},
	}
	if s.Status != nil {
		result["status"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *s.Status}}}
	}
	if s.TypeCode != nil {
		result["type"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *s.TypeCode, Display: strVal(s.TypeDisplay)}}}
	}
	if s.DomainCode != nil {
		result["domain"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *s.DomainCode, Display: strVal(s.DomainDisplay)}}}
	}
	if s.Description != nil {
		result["description"] = *s.Description
	}
	if s.SourceReference != nil {
		result["source"] = []fhir.Reference{{Reference: *s.SourceReference}}
	}
	if s.Comment != nil {
		result["comment"] = *s.Comment
	}
	if s.MolecularWeightAmount != nil {
		mw := map[string]interface{}{
			"amount": map[string]interface{}{"value": *s.MolecularWeightAmount},
		}
		if s.MolecularWeightUnit != nil {
			mw["amount"].(map[string]interface{})["unit"] = *s.MolecularWeightUnit
		}
		result["molecularWeight"] = []interface{}{mw}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
