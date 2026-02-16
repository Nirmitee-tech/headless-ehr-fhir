package conceptmap

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// ConceptMap maps to the concept_map table (FHIR ConceptMap resource).
type ConceptMap struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	FHIRID      string     `db:"fhir_id" json:"fhir_id"`
	Status      string     `db:"status" json:"status"`
	URL         *string    `db:"url" json:"url,omitempty"`
	Name        *string    `db:"name" json:"name,omitempty"`
	Title       *string    `db:"title" json:"title,omitempty"`
	Description *string    `db:"description" json:"description,omitempty"`
	Publisher   *string    `db:"publisher" json:"publisher,omitempty"`
	Date        *time.Time `db:"date" json:"date,omitempty"`
	SourceURI   *string    `db:"source_uri" json:"source_uri,omitempty"`
	TargetURI   *string    `db:"target_uri" json:"target_uri,omitempty"`
	Purpose     *string    `db:"purpose" json:"purpose,omitempty"`
	VersionID   int        `db:"version_id" json:"version_id"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

func (cm *ConceptMap) GetVersionID() int  { return cm.VersionID }
func (cm *ConceptMap) SetVersionID(v int) { cm.VersionID = v }

func (cm *ConceptMap) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ConceptMap",
		"id":           cm.FHIRID,
		"status":       cm.Status,
		"meta":         fhir.Meta{LastUpdated: cm.UpdatedAt},
	}
	if cm.URL != nil {
		result["url"] = *cm.URL
	}
	if cm.Name != nil {
		result["name"] = *cm.Name
	}
	if cm.Title != nil {
		result["title"] = *cm.Title
	}
	if cm.Description != nil {
		result["description"] = *cm.Description
	}
	if cm.Publisher != nil {
		result["publisher"] = *cm.Publisher
	}
	if cm.Date != nil {
		result["date"] = cm.Date.Format("2006-01-02")
	}
	if cm.SourceURI != nil {
		result["sourceUri"] = *cm.SourceURI
	}
	if cm.TargetURI != nil {
		result["targetUri"] = *cm.TargetURI
	}
	if cm.Purpose != nil {
		result["purpose"] = *cm.Purpose
	}
	return result
}
