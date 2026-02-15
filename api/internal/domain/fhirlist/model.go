package fhirlist

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// FHIRList maps to the fhir_list table (FHIR List resource).
// Named FHIRList to avoid conflict with the Go keyword "list".
type FHIRList struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	FHIRID                string     `db:"fhir_id" json:"fhir_id"`
	Status                string     `db:"status" json:"status"`
	Mode                  string     `db:"mode" json:"mode"`
	Title                 *string    `db:"title" json:"title,omitempty"`
	CodeCode              *string    `db:"code_code" json:"code_code,omitempty"`
	CodeDisplay           *string    `db:"code_display" json:"code_display,omitempty"`
	SubjectPatientID      *uuid.UUID `db:"subject_patient_id" json:"subject_patient_id,omitempty"`
	EncounterID           *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	Date                  *time.Time `db:"date" json:"date,omitempty"`
	SourcePractitionerID  *uuid.UUID `db:"source_practitioner_id" json:"source_practitioner_id,omitempty"`
	OrderedBy             *string    `db:"ordered_by" json:"ordered_by,omitempty"`
	Note                  *string    `db:"note" json:"note,omitempty"`
	VersionID             int        `db:"version_id" json:"version_id"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (l *FHIRList) GetVersionID() int { return l.VersionID }

// SetVersionID sets the current version.
func (l *FHIRList) SetVersionID(v int) { l.VersionID = v }

func (l *FHIRList) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "List",
		"id":           l.FHIRID,
		"status":       l.Status,
		"mode":         l.Mode,
		"meta":         fhir.Meta{LastUpdated: l.UpdatedAt},
	}
	if l.Title != nil {
		result["title"] = *l.Title
	}
	if l.CodeCode != nil {
		result["code"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *l.CodeCode, Display: strVal(l.CodeDisplay)}}}
	}
	if l.SubjectPatientID != nil {
		result["subject"] = fhir.Reference{Reference: fhir.FormatReference("Patient", l.SubjectPatientID.String())}
	}
	if l.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", l.EncounterID.String())}
	}
	if l.Date != nil {
		result["date"] = l.Date.Format("2006-01-02T15:04:05Z")
	}
	if l.SourcePractitionerID != nil {
		result["source"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", l.SourcePractitionerID.String())}
	}
	if l.OrderedBy != nil {
		result["orderedBy"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *l.OrderedBy}}}
	}
	if l.Note != nil {
		result["note"] = []map[string]string{{"text": *l.Note}}
	}
	return result
}

// FHIRListEntry maps to the fhir_list_entry table.
type FHIRListEntry struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	ListID        uuid.UUID  `db:"list_id" json:"list_id"`
	ItemReference string     `db:"item_reference" json:"item_reference"`
	ItemDisplay   *string    `db:"item_display" json:"item_display,omitempty"`
	Date          *time.Time `db:"date" json:"date,omitempty"`
	Deleted       bool       `db:"deleted" json:"deleted"`
	FlagCode      *string    `db:"flag_code" json:"flag_code,omitempty"`
	FlagDisplay   *string    `db:"flag_display" json:"flag_display,omitempty"`
}

func (e *FHIRListEntry) ToFHIR() map[string]interface{} {
	entry := map[string]interface{}{
		"item":    fhir.Reference{Reference: e.ItemReference, Display: strVal(e.ItemDisplay)},
		"deleted": e.Deleted,
	}
	if e.Date != nil {
		entry["date"] = e.Date.Format("2006-01-02T15:04:05Z")
	}
	if e.FlagCode != nil {
		entry["flag"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *e.FlagCode, Display: strVal(e.FlagDisplay)}}}
	}
	return entry
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
