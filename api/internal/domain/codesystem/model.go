package codesystem

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// CodeSystem maps to the code_system table (FHIR CodeSystem resource).
type CodeSystem struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	FHIRID           string     `db:"fhir_id" json:"fhir_id"`
	Status           string     `db:"status" json:"status"`
	URL              *string    `db:"url" json:"url,omitempty"`
	Name             *string    `db:"name" json:"name,omitempty"`
	Title            *string    `db:"title" json:"title,omitempty"`
	Description      *string    `db:"description" json:"description,omitempty"`
	Publisher        *string    `db:"publisher" json:"publisher,omitempty"`
	Date             *time.Time `db:"date" json:"date,omitempty"`
	Content          string     `db:"content" json:"content"`
	ValueSetURI      *string    `db:"value_set_uri" json:"value_set_uri,omitempty"`
	HierarchyMeaning *string   `db:"hierarchy_meaning" json:"hierarchy_meaning,omitempty"`
	Compositional    bool       `db:"compositional" json:"compositional"`
	VersionNeeded    bool       `db:"version_needed" json:"version_needed"`
	Count            *int       `db:"count" json:"count,omitempty"`
	VersionID        int        `db:"version_id" json:"version_id"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

func (cs *CodeSystem) GetVersionID() int  { return cs.VersionID }
func (cs *CodeSystem) SetVersionID(v int) { cs.VersionID = v }

func (cs *CodeSystem) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "CodeSystem",
		"id":           cs.FHIRID,
		"status":       cs.Status,
		"content":      cs.Content,
		"meta":         fhir.Meta{LastUpdated: cs.UpdatedAt},
	}
	if cs.URL != nil {
		result["url"] = *cs.URL
	}
	if cs.Name != nil {
		result["name"] = *cs.Name
	}
	if cs.Title != nil {
		result["title"] = *cs.Title
	}
	if cs.Description != nil {
		result["description"] = *cs.Description
	}
	if cs.Publisher != nil {
		result["publisher"] = *cs.Publisher
	}
	if cs.Date != nil {
		result["date"] = cs.Date.Format("2006-01-02")
	}
	if cs.ValueSetURI != nil {
		result["valueSet"] = *cs.ValueSetURI
	}
	if cs.HierarchyMeaning != nil {
		result["hierarchyMeaning"] = *cs.HierarchyMeaning
	}
	if cs.Compositional {
		result["compositional"] = cs.Compositional
	}
	if cs.VersionNeeded {
		result["versionNeeded"] = cs.VersionNeeded
	}
	if cs.Count != nil {
		result["count"] = *cs.Count
	}
	return result
}
