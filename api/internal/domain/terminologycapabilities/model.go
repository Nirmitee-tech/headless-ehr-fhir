package terminologycapabilities

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// TerminologyCapabilities maps to the terminology_capabilities table (FHIR TerminologyCapabilities resource).
type TerminologyCapabilities struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	FHIRID          string     `db:"fhir_id" json:"fhir_id"`
	Status          string     `db:"status" json:"status"`
	URL             *string    `db:"url" json:"url,omitempty"`
	Name            *string    `db:"name" json:"name,omitempty"`
	Title           *string    `db:"title" json:"title,omitempty"`
	Description     *string    `db:"description" json:"description,omitempty"`
	Publisher       *string    `db:"publisher" json:"publisher,omitempty"`
	Date            *time.Time `db:"date" json:"date,omitempty"`
	Kind            string     `db:"kind" json:"kind"`
	CodeSearch      *string    `db:"code_search" json:"code_search,omitempty"`
	Translation     bool       `db:"translation" json:"translation"`
	Closure         bool       `db:"closure" json:"closure"`
	SoftwareName    *string    `db:"software_name" json:"software_name,omitempty"`
	SoftwareVersion *string    `db:"software_version" json:"software_version,omitempty"`
	VersionID       int        `db:"version_id" json:"version_id"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

func (tc *TerminologyCapabilities) GetVersionID() int  { return tc.VersionID }
func (tc *TerminologyCapabilities) SetVersionID(v int) { tc.VersionID = v }

func (tc *TerminologyCapabilities) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "TerminologyCapabilities",
		"id":           tc.FHIRID,
		"status":       tc.Status,
		"kind":         tc.Kind,
		"translation":  tc.Translation,
		"closure":      map[string]interface{}{"translation": tc.Closure},
		"meta":         fhir.Meta{LastUpdated: tc.UpdatedAt},
	}
	if tc.URL != nil {
		result["url"] = *tc.URL
	}
	if tc.Name != nil {
		result["name"] = *tc.Name
	}
	if tc.Title != nil {
		result["title"] = *tc.Title
	}
	if tc.Description != nil {
		result["description"] = *tc.Description
	}
	if tc.Publisher != nil {
		result["publisher"] = *tc.Publisher
	}
	if tc.Date != nil {
		result["date"] = tc.Date.Format("2006-01-02")
	}
	if tc.CodeSearch != nil {
		result["codeSearch"] = *tc.CodeSearch
	}
	if tc.SoftwareName != nil {
		software := map[string]interface{}{"name": *tc.SoftwareName}
		if tc.SoftwareVersion != nil {
			software["version"] = *tc.SoftwareVersion
		}
		result["software"] = software
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
