package documentmanifest

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// DocumentManifest maps to the document_manifest table (FHIR DocumentManifest resource).
type DocumentManifest struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	FHIRID             string     `db:"fhir_id" json:"fhir_id"`
	Status             string     `db:"status" json:"status"`
	TypeCode           *string    `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay        *string    `db:"type_display" json:"type_display,omitempty"`
	SubjectReference   *string    `db:"subject_reference" json:"subject_reference,omitempty"`
	Created            *time.Time `db:"created" json:"created,omitempty"`
	AuthorReference    *string    `db:"author_reference" json:"author_reference,omitempty"`
	RecipientReference *string    `db:"recipient_reference" json:"recipient_reference,omitempty"`
	SourceURL          *string    `db:"source_url" json:"source_url,omitempty"`
	Description        *string    `db:"description" json:"description,omitempty"`
	VersionID          int        `db:"version_id" json:"version_id"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

func (d *DocumentManifest) GetVersionID() int  { return d.VersionID }
func (d *DocumentManifest) SetVersionID(v int) { d.VersionID = v }

func (d *DocumentManifest) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "DocumentManifest",
		"id":           d.FHIRID,
		"status":       d.Status,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", d.VersionID),
			LastUpdated: d.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/DocumentManifest"},
		},
	}
	if d.TypeCode != nil {
		result["type"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *d.TypeCode, Display: strVal(d.TypeDisplay)}}}
	}
	if d.SubjectReference != nil {
		result["subject"] = fhir.Reference{Reference: *d.SubjectReference}
	}
	if d.Created != nil {
		result["created"] = d.Created.Format("2006-01-02T15:04:05Z")
	}
	if d.AuthorReference != nil {
		result["author"] = []fhir.Reference{{Reference: *d.AuthorReference}}
	}
	if d.RecipientReference != nil {
		result["recipient"] = []fhir.Reference{{Reference: *d.RecipientReference}}
	}
	if d.SourceURL != nil {
		result["source"] = *d.SourceURL
	}
	if d.Description != nil {
		result["description"] = *d.Description
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
