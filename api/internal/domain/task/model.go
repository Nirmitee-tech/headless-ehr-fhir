package task

import (
	"encoding/json"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Task maps to the task table (FHIR Task resource).
type Task struct {
	ID                     uuid.UUID        `db:"id" json:"id"`
	FHIRID                 string           `db:"fhir_id" json:"fhir_id"`
	Status                 string           `db:"status" json:"status"`
	StatusReason           *string          `db:"status_reason" json:"status_reason,omitempty"`
	Intent                 string           `db:"intent" json:"intent"`
	Priority               *string          `db:"priority" json:"priority,omitempty"`
	CodeValue              *string          `db:"code_value" json:"code_value,omitempty"`
	CodeDisplay            *string          `db:"code_display" json:"code_display,omitempty"`
	CodeSystem             *string          `db:"code_system" json:"code_system,omitempty"`
	Description            *string          `db:"description" json:"description,omitempty"`
	FocusResourceType      *string          `db:"focus_resource_type" json:"focus_resource_type,omitempty"`
	FocusResourceID        *string          `db:"focus_resource_id" json:"focus_resource_id,omitempty"`
	ForPatientID           uuid.UUID        `db:"for_patient_id" json:"for_patient_id"`
	EncounterID            *uuid.UUID       `db:"encounter_id" json:"encounter_id,omitempty"`
	AuthoredOn             *time.Time       `db:"authored_on" json:"authored_on,omitempty"`
	LastModified           *time.Time       `db:"last_modified" json:"last_modified,omitempty"`
	RequesterID            *uuid.UUID       `db:"requester_id" json:"requester_id,omitempty"`
	OwnerID                *uuid.UUID       `db:"owner_id" json:"owner_id,omitempty"`
	ReasonCode             *string          `db:"reason_code" json:"reason_code,omitempty"`
	ReasonDisplay          *string          `db:"reason_display" json:"reason_display,omitempty"`
	Note                   *string          `db:"note" json:"note,omitempty"`
	RestrictionRepetitions *int             `db:"restriction_repetitions" json:"restriction_repetitions,omitempty"`
	RestrictionPeriodStart *time.Time       `db:"restriction_period_start" json:"restriction_period_start,omitempty"`
	RestrictionPeriodEnd   *time.Time       `db:"restriction_period_end" json:"restriction_period_end,omitempty"`
	InputJSON              *json.RawMessage `db:"input_json" json:"input_json,omitempty"`
	OutputJSON             *json.RawMessage `db:"output_json" json:"output_json,omitempty"`
	VersionID              int              `db:"version_id" json:"version_id"`
	CreatedAt              time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time        `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (t *Task) GetVersionID() int { return t.VersionID }

// SetVersionID sets the current version.
func (t *Task) SetVersionID(v int) { t.VersionID = v }

func (t *Task) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Task",
		"id":           t.FHIRID,
		"status":       t.Status,
		"intent":       t.Intent,
		"for":          fhir.Reference{Reference: fhir.FormatReference("Patient", t.ForPatientID.String())},
		"meta":         fhir.Meta{LastUpdated: t.UpdatedAt},
	}

	if t.StatusReason != nil {
		result["statusReason"] = fhir.CodeableConcept{Text: *t.StatusReason}
	}
	if t.Priority != nil {
		result["priority"] = *t.Priority
	}
	if t.CodeValue != nil {
		result["code"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System:  strVal(t.CodeSystem),
				Code:    *t.CodeValue,
				Display: strVal(t.CodeDisplay),
			}},
		}
	}
	if t.Description != nil {
		result["description"] = *t.Description
	}
	if t.FocusResourceType != nil && t.FocusResourceID != nil {
		result["focus"] = fhir.Reference{
			Reference: fhir.FormatReference(*t.FocusResourceType, *t.FocusResourceID),
		}
	}
	if t.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", t.EncounterID.String())}
	}
	if t.AuthoredOn != nil {
		result["authoredOn"] = t.AuthoredOn.Format(time.RFC3339)
	}
	if t.LastModified != nil {
		result["lastModified"] = t.LastModified.Format(time.RFC3339)
	}
	if t.RequesterID != nil {
		result["requester"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", t.RequesterID.String())}
	}
	if t.OwnerID != nil {
		result["owner"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", t.OwnerID.String())}
	}
	if t.ReasonCode != nil {
		result["reasonCode"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *t.ReasonCode, Display: strVal(t.ReasonDisplay)}},
		}
	}
	if t.Note != nil {
		result["note"] = []map[string]string{{"text": *t.Note}}
	}
	if t.RestrictionRepetitions != nil || t.RestrictionPeriodStart != nil {
		restriction := map[string]interface{}{}
		if t.RestrictionRepetitions != nil {
			restriction["repetitions"] = *t.RestrictionRepetitions
		}
		if t.RestrictionPeriodStart != nil {
			restriction["period"] = fhir.Period{Start: t.RestrictionPeriodStart, End: t.RestrictionPeriodEnd}
		}
		result["restriction"] = restriction
	}
	if t.InputJSON != nil {
		var inputs []interface{}
		if err := json.Unmarshal(*t.InputJSON, &inputs); err == nil {
			result["input"] = inputs
		}
	}
	if t.OutputJSON != nil {
		var outputs []interface{}
		if err := json.Unmarshal(*t.OutputJSON, &outputs); err == nil {
			result["output"] = outputs
		}
	}

	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
