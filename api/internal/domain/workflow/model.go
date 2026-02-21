package workflow

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// ActivityDefinition maps to the activity_definition table (FHIR ActivityDefinition resource).
type ActivityDefinition struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	FHIRID             string     `db:"fhir_id" json:"fhir_id"`
	URL                *string    `db:"url" json:"url,omitempty"`
	Status             string     `db:"status" json:"status"`
	Name               *string    `db:"name" json:"name,omitempty"`
	Title              *string    `db:"title" json:"title,omitempty"`
	Description        *string    `db:"description" json:"description,omitempty"`
	Purpose            *string    `db:"purpose" json:"purpose,omitempty"`
	Kind               *string    `db:"kind" json:"kind,omitempty"`
	CodeCode           *string    `db:"code_code" json:"code_code,omitempty"`
	CodeDisplay        *string    `db:"code_display" json:"code_display,omitempty"`
	CodeSystem         *string    `db:"code_system" json:"code_system,omitempty"`
	Intent             *string    `db:"intent" json:"intent,omitempty"`
	Priority           *string    `db:"priority" json:"priority,omitempty"`
	DoNotPerform       bool       `db:"do_not_perform" json:"do_not_perform"`
	TimingDescription  *string    `db:"timing_description" json:"timing_description,omitempty"`
	LocationID         *uuid.UUID `db:"location_id" json:"location_id,omitempty"`
	QuantityValue      *float64   `db:"quantity_value" json:"quantity_value,omitempty"`
	QuantityUnit       *string    `db:"quantity_unit" json:"quantity_unit,omitempty"`
	DosageText         *string    `db:"dosage_text" json:"dosage_text,omitempty"`
	Publisher          *string    `db:"publisher" json:"publisher,omitempty"`
	EffectiveStart     *time.Time `db:"effective_start" json:"effective_start,omitempty"`
	EffectiveEnd       *time.Time `db:"effective_end" json:"effective_end,omitempty"`
	ApprovalDate       *time.Time `db:"approval_date" json:"approval_date,omitempty"`
	LastReviewDate     *time.Time `db:"last_review_date" json:"last_review_date,omitempty"`
	VersionID          int        `db:"version_id" json:"version_id"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (a *ActivityDefinition) GetVersionID() int { return a.VersionID }

// SetVersionID sets the current version.
func (a *ActivityDefinition) SetVersionID(v int) { a.VersionID = v }

func (a *ActivityDefinition) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ActivityDefinition",
		"id":           a.FHIRID,
		"status":       a.Status,
		"doNotPerform": a.DoNotPerform,
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", a.VersionID),
			LastUpdated: a.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/ActivityDefinition"},
		},
	}
	if a.URL != nil {
		result["url"] = *a.URL
	}
	if a.Name != nil {
		result["name"] = *a.Name
	}
	if a.Title != nil {
		result["title"] = *a.Title
	}
	if a.Description != nil {
		result["description"] = *a.Description
	}
	if a.Purpose != nil {
		result["purpose"] = *a.Purpose
	}
	if a.Kind != nil {
		result["kind"] = *a.Kind
	}
	if a.CodeCode != nil {
		result["code"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System:  strVal(a.CodeSystem),
				Code:    *a.CodeCode,
				Display: strVal(a.CodeDisplay),
			}},
		}
	}
	if a.Intent != nil {
		result["intent"] = *a.Intent
	}
	if a.Priority != nil {
		result["priority"] = *a.Priority
	}
	if a.TimingDescription != nil {
		result["timingString"] = *a.TimingDescription
	}
	if a.LocationID != nil {
		result["location"] = fhir.Reference{Reference: fhir.FormatReference("Location", a.LocationID.String())}
	}
	if a.QuantityValue != nil {
		qty := map[string]interface{}{"value": *a.QuantityValue}
		if a.QuantityUnit != nil {
			qty["unit"] = *a.QuantityUnit
		}
		result["quantity"] = qty
	}
	if a.DosageText != nil {
		result["dosage"] = []map[string]string{{"text": *a.DosageText}}
	}
	if a.Publisher != nil {
		result["publisher"] = *a.Publisher
	}
	if a.EffectiveStart != nil || a.EffectiveEnd != nil {
		result["effectivePeriod"] = fhir.Period{Start: a.EffectiveStart, End: a.EffectiveEnd}
	}
	if a.ApprovalDate != nil {
		result["approvalDate"] = a.ApprovalDate.Format("2006-01-02")
	}
	if a.LastReviewDate != nil {
		result["lastReviewDate"] = a.LastReviewDate.Format("2006-01-02")
	}
	return result
}

// RequestGroup maps to the request_group table (FHIR RequestGroup resource).
type RequestGroup struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	FHIRID            string     `db:"fhir_id" json:"fhir_id"`
	Status            string     `db:"status" json:"status"`
	Intent            string     `db:"intent" json:"intent"`
	Priority          *string    `db:"priority" json:"priority,omitempty"`
	CodeCode          *string    `db:"code_code" json:"code_code,omitempty"`
	CodeDisplay       *string    `db:"code_display" json:"code_display,omitempty"`
	SubjectPatientID  *uuid.UUID `db:"subject_patient_id" json:"subject_patient_id,omitempty"`
	EncounterID       *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	AuthoredOn        *time.Time `db:"authored_on" json:"authored_on,omitempty"`
	AuthorID          *uuid.UUID `db:"author_id" json:"author_id,omitempty"`
	ReasonCode        *string    `db:"reason_code" json:"reason_code,omitempty"`
	ReasonDisplay     *string    `db:"reason_display" json:"reason_display,omitempty"`
	Note              *string    `db:"note" json:"note,omitempty"`
	VersionID         int        `db:"version_id" json:"version_id"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (rg *RequestGroup) GetVersionID() int { return rg.VersionID }

// SetVersionID sets the current version.
func (rg *RequestGroup) SetVersionID(v int) { rg.VersionID = v }

func (rg *RequestGroup) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "RequestGroup",
		"id":           rg.FHIRID,
		"status":       rg.Status,
		"intent":       rg.Intent,
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", rg.VersionID),
			LastUpdated: rg.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/RequestGroup"},
		},
	}
	if rg.Priority != nil {
		result["priority"] = *rg.Priority
	}
	if rg.CodeCode != nil {
		result["code"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				Code:    *rg.CodeCode,
				Display: strVal(rg.CodeDisplay),
			}},
		}
	}
	if rg.SubjectPatientID != nil {
		result["subject"] = fhir.Reference{Reference: fhir.FormatReference("Patient", rg.SubjectPatientID.String())}
	}
	if rg.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", rg.EncounterID.String())}
	}
	if rg.AuthoredOn != nil {
		result["authoredOn"] = rg.AuthoredOn.Format("2006-01-02T15:04:05Z")
	}
	if rg.AuthorID != nil {
		result["author"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", rg.AuthorID.String())}
	}
	if rg.ReasonCode != nil {
		result["reasonCode"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *rg.ReasonCode, Display: strVal(rg.ReasonDisplay)}},
		}}
	}
	if rg.Note != nil {
		result["note"] = []map[string]string{{"text": *rg.Note}}
	}
	return result
}

// RequestGroupAction maps to the request_group_action table (child of RequestGroup).
type RequestGroupAction struct {
	ID                    uuid.UUID `db:"id" json:"id"`
	RequestGroupID        uuid.UUID `db:"request_group_id" json:"request_group_id"`
	Prefix                *string   `db:"prefix" json:"prefix,omitempty"`
	Title                 *string   `db:"title" json:"title,omitempty"`
	Description           *string   `db:"description" json:"description,omitempty"`
	Priority              *string   `db:"priority" json:"priority,omitempty"`
	ResourceReference     *string   `db:"resource_reference" json:"resource_reference,omitempty"`
	SelectionBehavior     *string   `db:"selection_behavior" json:"selection_behavior,omitempty"`
	RequiredBehavior      *string   `db:"required_behavior" json:"required_behavior,omitempty"`
	PrecheckBehavior      *string   `db:"precheck_behavior" json:"precheck_behavior,omitempty"`
	CardinalityBehavior   *string   `db:"cardinality_behavior" json:"cardinality_behavior,omitempty"`
}

// GuidanceResponse maps to the guidance_response table (FHIR GuidanceResponse resource).
type GuidanceResponse struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	FHIRID            string     `db:"fhir_id" json:"fhir_id"`
	RequestIdentifier *string    `db:"request_identifier" json:"request_identifier,omitempty"`
	ModuleURI         string     `db:"module_uri" json:"module_uri"`
	Status            string     `db:"status" json:"status"`
	SubjectPatientID  *uuid.UUID `db:"subject_patient_id" json:"subject_patient_id,omitempty"`
	EncounterID       *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	OccurrenceDate    *time.Time `db:"occurrence_date" json:"occurrence_date,omitempty"`
	PerformerID       *uuid.UUID `db:"performer_id" json:"performer_id,omitempty"`
	ReasonCode        *string    `db:"reason_code" json:"reason_code,omitempty"`
	ReasonDisplay     *string    `db:"reason_display" json:"reason_display,omitempty"`
	Note              *string    `db:"note" json:"note,omitempty"`
	ResultReference   *string    `db:"result_reference" json:"result_reference,omitempty"`
	VersionID         int        `db:"version_id" json:"version_id"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (gr *GuidanceResponse) GetVersionID() int { return gr.VersionID }

// SetVersionID sets the current version.
func (gr *GuidanceResponse) SetVersionID(v int) { gr.VersionID = v }

func (gr *GuidanceResponse) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "GuidanceResponse",
		"id":           gr.FHIRID,
		"moduleUri":    gr.ModuleURI,
		"status":       gr.Status,
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", gr.VersionID),
			LastUpdated: gr.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/GuidanceResponse"},
		},
	}
	if gr.RequestIdentifier != nil {
		result["requestIdentifier"] = fhir.Identifier{Value: *gr.RequestIdentifier}
	}
	if gr.SubjectPatientID != nil {
		result["subject"] = fhir.Reference{Reference: fhir.FormatReference("Patient", gr.SubjectPatientID.String())}
	}
	if gr.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", gr.EncounterID.String())}
	}
	if gr.OccurrenceDate != nil {
		result["occurrenceDateTime"] = gr.OccurrenceDate.Format("2006-01-02T15:04:05Z")
	}
	if gr.PerformerID != nil {
		result["performer"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", gr.PerformerID.String())}
	}
	if gr.ReasonCode != nil {
		result["reasonCode"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *gr.ReasonCode, Display: strVal(gr.ReasonDisplay)}},
		}}
	}
	if gr.Note != nil {
		result["note"] = []map[string]string{{"text": *gr.Note}}
	}
	if gr.ResultReference != nil {
		result["result"] = fhir.Reference{Reference: *gr.ResultReference}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
