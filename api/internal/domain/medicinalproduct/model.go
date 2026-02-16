package medicinalproduct

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type MedicinalProduct struct {
	ID                                   uuid.UUID `db:"id" json:"id"`
	FHIRID                               string    `db:"fhir_id" json:"fhir_id"`
	Status                               *string   `db:"status" json:"status,omitempty"`
	TypeCode                             *string   `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay                          *string   `db:"type_display" json:"type_display,omitempty"`
	DomainCode                           *string   `db:"domain_code" json:"domain_code,omitempty"`
	DomainDisplay                        *string   `db:"domain_display" json:"domain_display,omitempty"`
	Description                          *string   `db:"description" json:"description,omitempty"`
	CombinedPharmaceuticalDoseFormCode   *string   `db:"combined_pharmaceutical_dose_form_code" json:"combined_pharmaceutical_dose_form_code,omitempty"`
	CombinedPharmaceuticalDoseFormDisplay *string   `db:"combined_pharmaceutical_dose_form_display" json:"combined_pharmaceutical_dose_form_display,omitempty"`
	LegalStatusOfSupplyCode              *string   `db:"legal_status_of_supply_code" json:"legal_status_of_supply_code,omitempty"`
	AdditionalMonitoring                 *bool     `db:"additional_monitoring" json:"additional_monitoring,omitempty"`
	VersionID                            int       `db:"version_id" json:"version_id"`
	CreatedAt                            time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                            time.Time `db:"updated_at" json:"updated_at"`
}

func (m *MedicinalProduct) GetVersionID() int  { return m.VersionID }
func (m *MedicinalProduct) SetVersionID(v int) { m.VersionID = v }

func (m *MedicinalProduct) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "MedicinalProduct",
		"id":           m.FHIRID,
		"meta":         fhir.Meta{LastUpdated: m.UpdatedAt},
	}
	if m.TypeCode != nil {
		result["type"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.TypeCode, Display: strVal(m.TypeDisplay)}}}
	}
	if m.DomainCode != nil {
		result["domain"] = fhir.Coding{Code: *m.DomainCode, Display: strVal(m.DomainDisplay)}
	}
	if m.Description != nil {
		result["description"] = *m.Description
	}
	if m.CombinedPharmaceuticalDoseFormCode != nil {
		result["combinedPharmaceuticalDoseForm"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.CombinedPharmaceuticalDoseFormCode, Display: strVal(m.CombinedPharmaceuticalDoseFormDisplay)}}}
	}
	if m.LegalStatusOfSupplyCode != nil {
		result["legalStatusOfSupply"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.LegalStatusOfSupplyCode}}}
	}
	if m.AdditionalMonitoring != nil {
		result["additionalMonitoringIndicator"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: boolToCode(*m.AdditionalMonitoring)}}}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func boolToCode(b bool) string {
	if b {
		return "Y"
	}
	return "N"
}
