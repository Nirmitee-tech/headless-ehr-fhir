package familyhistory

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// FamilyMemberHistory maps to the family_member_history table (FHIR FamilyMemberHistory resource).
type FamilyMemberHistory struct {
	ID                  uuid.UUID  `db:"id" json:"id"`
	FHIRID              string     `db:"fhir_id" json:"fhir_id"`
	Status              string     `db:"status" json:"status"`
	PatientID           uuid.UUID  `db:"patient_id" json:"patient_id"`
	Date                *time.Time `db:"date" json:"date,omitempty"`
	Name                *string    `db:"name" json:"name,omitempty"`
	RelationshipCode    string     `db:"relationship_code" json:"relationship_code"`
	RelationshipDisplay string     `db:"relationship_display" json:"relationship_display"`
	Sex                 *string    `db:"sex" json:"sex,omitempty"`
	BornDate            *time.Time `db:"born_date" json:"born_date,omitempty"`
	DeceasedBoolean     *bool      `db:"deceased_boolean" json:"deceased_boolean,omitempty"`
	DeceasedAge         *int       `db:"deceased_age" json:"deceased_age,omitempty"`
	Note                *string    `db:"note" json:"note,omitempty"`
	VersionID           int        `db:"version_id" json:"version_id"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (f *FamilyMemberHistory) GetVersionID() int { return f.VersionID }

// SetVersionID sets the current version.
func (f *FamilyMemberHistory) SetVersionID(v int) { f.VersionID = v }

func (f *FamilyMemberHistory) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "FamilyMemberHistory",
		"id":           f.FHIRID,
		"status":       f.Status,
		"patient":      fhir.Reference{Reference: fhir.FormatReference("Patient", f.PatientID.String())},
		"relationship": fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: f.RelationshipCode, Display: f.RelationshipDisplay}},
		},
		"meta": fhir.Meta{LastUpdated: f.UpdatedAt},
	}
	if f.Name != nil {
		result["name"] = *f.Name
	}
	if f.Date != nil {
		result["date"] = f.Date.Format(time.RFC3339)
	}
	if f.Sex != nil {
		result["sex"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *f.Sex}},
		}
	}
	if f.DeceasedBoolean != nil {
		result["deceasedBoolean"] = *f.DeceasedBoolean
	}
	if f.DeceasedAge != nil {
		result["deceasedAge"] = map[string]interface{}{
			"value": *f.DeceasedAge,
			"unit":  "a",
		}
	}
	if f.Note != nil {
		result["note"] = []map[string]string{{"text": *f.Note}}
	}
	return result
}

// FamilyMemberCondition maps to the family_member_condition table.
type FamilyMemberCondition struct {
	ID                uuid.UUID `db:"id" json:"id"`
	FamilyMemberID    uuid.UUID `db:"family_member_id" json:"family_member_id"`
	Code              string    `db:"code" json:"code"`
	Display           string    `db:"display" json:"display"`
	OutcomeCode       *string   `db:"outcome_code" json:"outcome_code,omitempty"`
	OutcomeDisplay    *string   `db:"outcome_display" json:"outcome_display,omitempty"`
	ContributedToDeath *bool    `db:"contributed_to_death" json:"contributed_to_death,omitempty"`
	OnsetAge          *int      `db:"onset_age" json:"onset_age,omitempty"`
}
