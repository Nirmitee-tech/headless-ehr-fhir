package devicerequest

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// DeviceRequest maps to the device_request table (FHIR DeviceRequest resource).
type DeviceRequest struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	FHIRID           string     `db:"fhir_id" json:"fhir_id"`
	Status           string     `db:"status" json:"status"`
	Intent           string     `db:"intent" json:"intent"`
	Priority         *string    `db:"priority" json:"priority,omitempty"`
	CodeCode         *string    `db:"code_code" json:"code_code,omitempty"`
	CodeDisplay      *string    `db:"code_display" json:"code_display,omitempty"`
	CodeSystem       *string    `db:"code_system" json:"code_system,omitempty"`
	SubjectPatientID uuid.UUID  `db:"subject_patient_id" json:"subject_patient_id"`
	EncounterID      *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	AuthoredOn       *time.Time `db:"authored_on" json:"authored_on,omitempty"`
	RequesterID      *uuid.UUID `db:"requester_id" json:"requester_id,omitempty"`
	PerformerID      *uuid.UUID `db:"performer_id" json:"performer_id,omitempty"`
	ReasonCode       *string    `db:"reason_code" json:"reason_code,omitempty"`
	ReasonDisplay    *string    `db:"reason_display" json:"reason_display,omitempty"`
	Note             *string    `db:"note" json:"note,omitempty"`
	VersionID        int        `db:"version_id" json:"version_id"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

func (d *DeviceRequest) GetVersionID() int  { return d.VersionID }
func (d *DeviceRequest) SetVersionID(v int) { d.VersionID = v }

func (d *DeviceRequest) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "DeviceRequest",
		"id":           d.FHIRID,
		"status":       d.Status,
		"intent":       d.Intent,
		"subject":      fhir.Reference{Reference: fhir.FormatReference("Patient", d.SubjectPatientID.String())},
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", d.VersionID),
			LastUpdated: d.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/DeviceRequest"},
		},
	}
	if d.Priority != nil {
		result["priority"] = *d.Priority
	}
	if d.CodeCode != nil {
		result["codeCodeableConcept"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *d.CodeCode, Display: strVal(d.CodeDisplay), System: strVal(d.CodeSystem)}}}
	}
	if d.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", d.EncounterID.String())}
	}
	if d.AuthoredOn != nil {
		result["authoredOn"] = d.AuthoredOn.Format(time.RFC3339)
	}
	if d.RequesterID != nil {
		result["requester"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", d.RequesterID.String())}
	}
	if d.PerformerID != nil {
		result["performer"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", d.PerformerID.String())}
	}
	if d.ReasonCode != nil {
		result["reasonCode"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *d.ReasonCode, Display: strVal(d.ReasonDisplay)}}}}
	}
	if d.Note != nil {
		result["note"] = []map[string]interface{}{{"text": *d.Note}}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
