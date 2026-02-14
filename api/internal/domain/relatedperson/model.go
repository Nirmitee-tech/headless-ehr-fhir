package relatedperson

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// RelatedPerson maps to the related_person table (FHIR RelatedPerson resource).
type RelatedPerson struct {
	ID                  uuid.UUID  `db:"id" json:"id"`
	FHIRID              string     `db:"fhir_id" json:"fhir_id"`
	Active              bool       `db:"active" json:"active"`
	PatientID           uuid.UUID  `db:"patient_id" json:"patient_id"`
	RelationshipCode    string     `db:"relationship_code" json:"relationship_code"`
	RelationshipDisplay string     `db:"relationship_display" json:"relationship_display"`
	FamilyName          *string    `db:"family_name" json:"family_name,omitempty"`
	GivenName           *string    `db:"given_name" json:"given_name,omitempty"`
	Phone               *string    `db:"phone" json:"phone,omitempty"`
	Email               *string    `db:"email" json:"email,omitempty"`
	Gender              *string    `db:"gender" json:"gender,omitempty"`
	BirthDate           *time.Time `db:"birth_date" json:"birth_date,omitempty"`
	AddressLine         *string    `db:"address_line" json:"address_line,omitempty"`
	AddressCity         *string    `db:"address_city" json:"address_city,omitempty"`
	AddressState        *string    `db:"address_state" json:"address_state,omitempty"`
	AddressPostalCode   *string    `db:"address_postal_code" json:"address_postal_code,omitempty"`
	PeriodStart         *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd           *time.Time `db:"period_end" json:"period_end,omitempty"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}

func (rp *RelatedPerson) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "RelatedPerson",
		"id":           rp.FHIRID,
		"active":       rp.Active,
		"patient":      fhir.Reference{Reference: fhir.FormatReference("Patient", rp.PatientID.String())},
		"relationship": []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: rp.RelationshipCode, Display: rp.RelationshipDisplay}},
		}},
		"meta": fhir.Meta{LastUpdated: rp.UpdatedAt},
	}
	if rp.FamilyName != nil || rp.GivenName != nil {
		name := map[string]interface{}{}
		if rp.FamilyName != nil {
			name["family"] = *rp.FamilyName
		}
		if rp.GivenName != nil {
			name["given"] = []string{*rp.GivenName}
		}
		result["name"] = []map[string]interface{}{name}
	}
	if rp.Gender != nil {
		result["gender"] = *rp.Gender
	}
	if rp.BirthDate != nil {
		result["birthDate"] = rp.BirthDate.Format("2006-01-02")
	}
	var telecom []map[string]string
	if rp.Phone != nil {
		telecom = append(telecom, map[string]string{"system": "phone", "value": *rp.Phone})
	}
	if rp.Email != nil {
		telecom = append(telecom, map[string]string{"system": "email", "value": *rp.Email})
	}
	if len(telecom) > 0 {
		result["telecom"] = telecom
	}
	return result
}

// RelatedPersonCommunication maps to the related_person_communication table.
type RelatedPersonCommunication struct {
	ID              uuid.UUID `db:"id" json:"id"`
	RelatedPersonID uuid.UUID `db:"related_person_id" json:"related_person_id"`
	LanguageCode    string    `db:"language_code" json:"language_code"`
	LanguageDisplay string    `db:"language_display" json:"language_display"`
	Preferred       bool      `db:"preferred" json:"preferred"`
}
