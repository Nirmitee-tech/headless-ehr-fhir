package communicationrequest

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// CommunicationRequest maps to the communication_request table (FHIR CommunicationRequest resource).
type CommunicationRequest struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	FHIRID          string     `db:"fhir_id" json:"fhir_id"`
	Status          string     `db:"status" json:"status"`
	PatientID       *uuid.UUID `db:"patient_id" json:"patient_id,omitempty"`
	EncounterID     *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	RequesterID     *uuid.UUID `db:"requester_id" json:"requester_id,omitempty"`
	RecipientID     *uuid.UUID `db:"recipient_id" json:"recipient_id,omitempty"`
	SenderID        *uuid.UUID `db:"sender_id" json:"sender_id,omitempty"`
	CategoryCode    *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay *string    `db:"category_display" json:"category_display,omitempty"`
	Priority        *string    `db:"priority" json:"priority,omitempty"`
	MediumCode      *string    `db:"medium_code" json:"medium_code,omitempty"`
	MediumDisplay   *string    `db:"medium_display" json:"medium_display,omitempty"`
	PayloadText     *string    `db:"payload_text" json:"payload_text,omitempty"`
	OccurrenceDate  *time.Time `db:"occurrence_date" json:"occurrence_date,omitempty"`
	AuthoredOn      *time.Time `db:"authored_on" json:"authored_on,omitempty"`
	ReasonCode      *string    `db:"reason_code" json:"reason_code,omitempty"`
	ReasonDisplay   *string    `db:"reason_display" json:"reason_display,omitempty"`
	Note            *string    `db:"note" json:"note,omitempty"`
	VersionID       int        `db:"version_id" json:"version_id"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

func (cr *CommunicationRequest) GetVersionID() int  { return cr.VersionID }
func (cr *CommunicationRequest) SetVersionID(v int)  { cr.VersionID = v }

func (cr *CommunicationRequest) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "CommunicationRequest",
		"id":           cr.FHIRID,
		"status":       cr.Status,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", cr.VersionID),
			LastUpdated: cr.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/CommunicationRequest"},
		},
	}
	if cr.PatientID != nil {
		result["subject"] = fhir.Reference{Reference: fhir.FormatReference("Patient", cr.PatientID.String())}
	}
	if cr.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", cr.EncounterID.String())}
	}
	if cr.RequesterID != nil {
		result["requester"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", cr.RequesterID.String())}
	}
	if cr.RecipientID != nil {
		result["recipient"] = []fhir.Reference{{Reference: fhir.FormatReference("Practitioner", cr.RecipientID.String())}}
	}
	if cr.SenderID != nil {
		result["sender"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", cr.SenderID.String())}
	}
	if cr.CategoryCode != nil {
		result["category"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *cr.CategoryCode, Display: strVal(cr.CategoryDisplay)}}}}
	}
	if cr.Priority != nil {
		result["priority"] = *cr.Priority
	}
	if cr.MediumCode != nil {
		result["medium"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *cr.MediumCode, Display: strVal(cr.MediumDisplay)}}}}
	}
	if cr.PayloadText != nil {
		result["payload"] = []map[string]interface{}{{"contentString": *cr.PayloadText}}
	}
	if cr.OccurrenceDate != nil {
		result["occurrenceDateTime"] = cr.OccurrenceDate.Format("2006-01-02T15:04:05Z")
	}
	if cr.AuthoredOn != nil {
		result["authoredOn"] = cr.AuthoredOn.Format("2006-01-02T15:04:05Z")
	}
	if cr.ReasonCode != nil {
		result["reasonCode"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *cr.ReasonCode, Display: strVal(cr.ReasonDisplay)}}}}
	}
	if cr.Note != nil {
		result["note"] = []map[string]interface{}{{"text": *cr.Note}}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
