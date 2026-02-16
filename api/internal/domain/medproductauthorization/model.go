package medproductauthorization

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type MedicinalProductAuthorization struct {
	ID                        uuid.UUID  `db:"id" json:"id"`
	FHIRID                    string     `db:"fhir_id" json:"fhir_id"`
	Status                    *string    `db:"status" json:"status,omitempty"`
	StatusDate                *time.Time `db:"status_date" json:"status_date,omitempty"`
	SubjectReference          *string    `db:"subject_reference" json:"subject_reference,omitempty"`
	CountryCode               *string    `db:"country_code" json:"country_code,omitempty"`
	CountryDisplay            *string    `db:"country_display" json:"country_display,omitempty"`
	JurisdictionCode          *string    `db:"jurisdiction_code" json:"jurisdiction_code,omitempty"`
	JurisdictionDisplay       *string    `db:"jurisdiction_display" json:"jurisdiction_display,omitempty"`
	ValidityPeriodStart       *time.Time `db:"validity_period_start" json:"validity_period_start,omitempty"`
	ValidityPeriodEnd         *time.Time `db:"validity_period_end" json:"validity_period_end,omitempty"`
	DateOfFirstAuthorization  *time.Time `db:"date_of_first_authorization" json:"date_of_first_authorization,omitempty"`
	InternationalBirthDate    *time.Time `db:"international_birth_date" json:"international_birth_date,omitempty"`
	HolderReference           *string    `db:"holder_reference" json:"holder_reference,omitempty"`
	RegulatorReference        *string    `db:"regulator_reference" json:"regulator_reference,omitempty"`
	VersionID                 int        `db:"version_id" json:"version_id"`
	CreatedAt                 time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt                 time.Time  `db:"updated_at" json:"updated_at"`
}

func (m *MedicinalProductAuthorization) GetVersionID() int  { return m.VersionID }
func (m *MedicinalProductAuthorization) SetVersionID(v int) { m.VersionID = v }

func (m *MedicinalProductAuthorization) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "MedicinalProductAuthorization",
		"id":           m.FHIRID,
		"meta":         fhir.Meta{LastUpdated: m.UpdatedAt},
	}
	if m.Status != nil {
		result["status"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.Status}}}
	}
	if m.StatusDate != nil {
		result["statusDate"] = m.StatusDate.Format("2006-01-02")
	}
	if m.SubjectReference != nil {
		result["subject"] = fhir.Reference{Reference: *m.SubjectReference}
	}
	if m.CountryCode != nil {
		result["country"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *m.CountryCode, Display: strVal(m.CountryDisplay)}}}}
	}
	if m.JurisdictionCode != nil {
		result["jurisdiction"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *m.JurisdictionCode, Display: strVal(m.JurisdictionDisplay)}}}}
	}
	if m.ValidityPeriodStart != nil || m.ValidityPeriodEnd != nil {
		p := fhir.Period{Start: m.ValidityPeriodStart, End: m.ValidityPeriodEnd}
		result["validityPeriod"] = p
	}
	if m.DateOfFirstAuthorization != nil {
		result["dateOfFirstAuthorization"] = m.DateOfFirstAuthorization.Format("2006-01-02")
	}
	if m.InternationalBirthDate != nil {
		result["internationalBirthDate"] = m.InternationalBirthDate.Format("2006-01-02")
	}
	if m.HolderReference != nil {
		result["holder"] = fhir.Reference{Reference: *m.HolderReference}
	}
	if m.RegulatorReference != nil {
		result["regulator"] = fhir.Reference{Reference: *m.RegulatorReference}
	}
	return result
}

func strVal(s *string) string {
	if s == nil { return "" }
	return *s
}
