package basic

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Basic maps to the basic table (FHIR Basic resource).
type Basic struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	FHIRID           string     `db:"fhir_id" json:"fhir_id"`
	CodeCode         string     `db:"code_code" json:"code_code"`
	CodeSystem       *string    `db:"code_system" json:"code_system,omitempty"`
	CodeDisplay      *string    `db:"code_display" json:"code_display,omitempty"`
	SubjectType      *string    `db:"subject_type" json:"subject_type,omitempty"`
	SubjectReference *string    `db:"subject_reference" json:"subject_reference,omitempty"`
	AuthorID         *uuid.UUID `db:"author_id" json:"author_id,omitempty"`
	AuthorDate       *time.Time `db:"author_date" json:"author_date,omitempty"`
	VersionID        int        `db:"version_id" json:"version_id"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

func (b *Basic) GetVersionID() int  { return b.VersionID }
func (b *Basic) SetVersionID(v int) { b.VersionID = v }

func (b *Basic) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Basic",
		"id":           b.FHIRID,
		"meta":         fhir.Meta{LastUpdated: b.UpdatedAt},
		"code": fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				Code:    b.CodeCode,
				System:  strVal(b.CodeSystem),
				Display: strVal(b.CodeDisplay),
			}},
		},
	}
	if b.SubjectType != nil && b.SubjectReference != nil {
		result["subject"] = fhir.Reference{Reference: fhir.FormatReference(*b.SubjectType, *b.SubjectReference)}
	}
	if b.AuthorID != nil {
		result["author"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", b.AuthorID.String())}
	}
	if b.AuthorDate != nil {
		result["created"] = b.AuthorDate.Format("2006-01-02")
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
