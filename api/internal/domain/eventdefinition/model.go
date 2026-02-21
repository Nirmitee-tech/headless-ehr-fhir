package eventdefinition

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// EventDefinition maps to the event_definition table (FHIR EventDefinition resource).
type EventDefinition struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	FHIRID           string     `db:"fhir_id" json:"fhir_id"`
	Status           string     `db:"status" json:"status"`
	URL              *string    `db:"url" json:"url,omitempty"`
	Name             *string    `db:"name" json:"name,omitempty"`
	Title            *string    `db:"title" json:"title,omitempty"`
	Description      *string    `db:"description" json:"description,omitempty"`
	Publisher        *string    `db:"publisher" json:"publisher,omitempty"`
	Date             *time.Time `db:"date" json:"date,omitempty"`
	Purpose          *string    `db:"purpose" json:"purpose,omitempty"`
	TriggerType      string     `db:"trigger_type" json:"trigger_type"`
	TriggerName      *string    `db:"trigger_name" json:"trigger_name,omitempty"`
	TriggerCondition *string    `db:"trigger_condition" json:"trigger_condition,omitempty"`
	VersionID        int        `db:"version_id" json:"version_id"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

func (e *EventDefinition) GetVersionID() int  { return e.VersionID }
func (e *EventDefinition) SetVersionID(v int) { e.VersionID = v }

func (e *EventDefinition) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "EventDefinition",
		"id":           e.FHIRID,
		"status":       e.Status,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", e.VersionID),
			LastUpdated: e.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/EventDefinition"},
		},
	}
	if e.URL != nil {
		result["url"] = *e.URL
	}
	if e.Name != nil {
		result["name"] = *e.Name
	}
	if e.Title != nil {
		result["title"] = *e.Title
	}
	if e.Description != nil {
		result["description"] = *e.Description
	}
	if e.Publisher != nil {
		result["publisher"] = *e.Publisher
	}
	if e.Date != nil {
		result["date"] = e.Date.Format("2006-01-02")
	}
	if e.Purpose != nil {
		result["purpose"] = *e.Purpose
	}
	trigger := map[string]interface{}{
		"type": e.TriggerType,
	}
	if e.TriggerName != nil {
		trigger["name"] = *e.TriggerName
	}
	if e.TriggerCondition != nil {
		trigger["condition"] = *e.TriggerCondition
	}
	result["trigger"] = []map[string]interface{}{trigger}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
