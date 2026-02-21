package person

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Person maps to the person table (FHIR Person resource).
type Person struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	FHIRID            string     `db:"fhir_id" json:"fhir_id"`
	Active            bool       `db:"active" json:"active"`
	NameFamily        *string    `db:"name_family" json:"name_family,omitempty"`
	NameGiven         *string    `db:"name_given" json:"name_given,omitempty"`
	Gender            *string    `db:"gender" json:"gender,omitempty"`
	BirthDate         *time.Time `db:"birth_date" json:"birth_date,omitempty"`
	AddressLine       *string    `db:"address_line" json:"address_line,omitempty"`
	AddressCity       *string    `db:"address_city" json:"address_city,omitempty"`
	AddressState      *string    `db:"address_state" json:"address_state,omitempty"`
	AddressPostalCode *string    `db:"address_postal_code" json:"address_postal_code,omitempty"`
	TelecomPhone      *string    `db:"telecom_phone" json:"telecom_phone,omitempty"`
	TelecomEmail      *string    `db:"telecom_email" json:"telecom_email,omitempty"`
	ManagingOrgID     *uuid.UUID `db:"managing_org_id" json:"managing_org_id,omitempty"`
	VersionID         int        `db:"version_id" json:"version_id"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

func (p *Person) GetVersionID() int  { return p.VersionID }
func (p *Person) SetVersionID(v int) { p.VersionID = v }

func (p *Person) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Person",
		"id":           p.FHIRID,
		"active":       p.Active,
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", p.VersionID),
			LastUpdated: p.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Person"},
		},
	}
	if p.NameFamily != nil || p.NameGiven != nil {
		name := fhir.HumanName{}
		if p.NameFamily != nil {
			name.Family = *p.NameFamily
		}
		if p.NameGiven != nil {
			name.Given = []string{*p.NameGiven}
		}
		result["name"] = []fhir.HumanName{name}
	}
	if p.Gender != nil {
		result["gender"] = *p.Gender
	}
	if p.BirthDate != nil {
		result["birthDate"] = p.BirthDate.Format("2006-01-02")
	}
	if p.AddressLine != nil || p.AddressCity != nil || p.AddressState != nil || p.AddressPostalCode != nil {
		addr := fhir.Address{}
		if p.AddressLine != nil {
			addr.Line = []string{*p.AddressLine}
		}
		if p.AddressCity != nil {
			addr.City = *p.AddressCity
		}
		if p.AddressState != nil {
			addr.State = *p.AddressState
		}
		if p.AddressPostalCode != nil {
			addr.PostalCode = *p.AddressPostalCode
		}
		result["address"] = []fhir.Address{addr}
	}
	var telecom []fhir.ContactPoint
	if p.TelecomPhone != nil {
		telecom = append(telecom, fhir.ContactPoint{System: "phone", Value: *p.TelecomPhone})
	}
	if p.TelecomEmail != nil {
		telecom = append(telecom, fhir.ContactPoint{System: "email", Value: *p.TelecomEmail})
	}
	if len(telecom) > 0 {
		result["telecom"] = telecom
	}
	if p.ManagingOrgID != nil {
		result["managingOrganization"] = fhir.Reference{Reference: fhir.FormatReference("Organization", p.ManagingOrgID.String())}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
