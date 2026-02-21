package medproductpackaged

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type MedicinalProductPackaged struct {
	ID                              uuid.UUID `db:"id" json:"id"`
	FHIRID                          string    `db:"fhir_id" json:"fhir_id"`
	SubjectReference                *string   `db:"subject_reference" json:"subject_reference,omitempty"`
	Description                     *string   `db:"description" json:"description,omitempty"`
	LegalStatusOfSupplyCode         *string   `db:"legal_status_of_supply_code" json:"legal_status_of_supply_code,omitempty"`
	LegalStatusOfSupplyDisplay      *string   `db:"legal_status_of_supply_display" json:"legal_status_of_supply_display,omitempty"`
	MarketingStatusCode             *string   `db:"marketing_status_code" json:"marketing_status_code,omitempty"`
	MarketingStatusDisplay          *string   `db:"marketing_status_display" json:"marketing_status_display,omitempty"`
	MarketingAuthorizationReference *string   `db:"marketing_authorization_reference" json:"marketing_authorization_reference,omitempty"`
	ManufacturerReference           *string   `db:"manufacturer_reference" json:"manufacturer_reference,omitempty"`
	VersionID                       int       `db:"version_id" json:"version_id"`
	CreatedAt                       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                       time.Time `db:"updated_at" json:"updated_at"`
}

func (m *MedicinalProductPackaged) GetVersionID() int  { return m.VersionID }
func (m *MedicinalProductPackaged) SetVersionID(v int) { m.VersionID = v }

func (m *MedicinalProductPackaged) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "MedicinalProductPackaged",
		"id":           m.FHIRID,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", m.VersionID),
			LastUpdated: m.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/MedicinalProductPackaged"},
		},
	}
	if m.SubjectReference != nil {
		result["subject"] = []fhir.Reference{{Reference: *m.SubjectReference}}
	}
	if m.Description != nil {
		result["description"] = *m.Description
	}
	if m.LegalStatusOfSupplyCode != nil {
		result["legalStatusOfSupply"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.LegalStatusOfSupplyCode, Display: strVal(m.LegalStatusOfSupplyDisplay)}}}
	}
	if m.MarketingStatusCode != nil {
		result["marketingStatus"] = []map[string]interface{}{{"status": fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.MarketingStatusCode, Display: strVal(m.MarketingStatusDisplay)}}}}}
	}
	if m.MarketingAuthorizationReference != nil {
		result["marketingAuthorization"] = fhir.Reference{Reference: *m.MarketingAuthorizationReference}
	}
	if m.ManufacturerReference != nil {
		result["manufacturer"] = []fhir.Reference{{Reference: *m.ManufacturerReference}}
	}
	return result
}

func strVal(s *string) string {
	if s == nil { return "" }
	return *s
}
