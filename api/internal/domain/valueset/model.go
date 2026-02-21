package valueset

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// ValueSet maps to the value_set table (FHIR ValueSet resource).
type ValueSet struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	FHIRID                string     `db:"fhir_id" json:"fhir_id"`
	Status                string     `db:"status" json:"status"`
	URL                   *string    `db:"url" json:"url,omitempty"`
	Name                  *string    `db:"name" json:"name,omitempty"`
	Title                 *string    `db:"title" json:"title,omitempty"`
	Description           *string    `db:"description" json:"description,omitempty"`
	Publisher             *string    `db:"publisher" json:"publisher,omitempty"`
	Date                  *time.Time `db:"date" json:"date,omitempty"`
	Immutable             bool       `db:"immutable" json:"immutable"`
	Purpose               *string    `db:"purpose" json:"purpose,omitempty"`
	Copyright             *string    `db:"copyright" json:"copyright,omitempty"`
	ComposeIncludeSystem  *string    `db:"compose_include_system" json:"compose_include_system,omitempty"`
	ComposeIncludeVersion *string    `db:"compose_include_version" json:"compose_include_version,omitempty"`
	ExpansionIdentifier   *string    `db:"expansion_identifier" json:"expansion_identifier,omitempty"`
	ExpansionTimestamp    *time.Time `db:"expansion_timestamp" json:"expansion_timestamp,omitempty"`
	VersionID             int        `db:"version_id" json:"version_id"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

func (vs *ValueSet) GetVersionID() int  { return vs.VersionID }
func (vs *ValueSet) SetVersionID(v int) { vs.VersionID = v }

func (vs *ValueSet) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ValueSet",
		"id":           vs.FHIRID,
		"status":       vs.Status,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", vs.VersionID),
			LastUpdated: vs.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/ValueSet"},
		},
	}
	if vs.URL != nil {
		result["url"] = *vs.URL
	}
	if vs.Name != nil {
		result["name"] = *vs.Name
	}
	if vs.Title != nil {
		result["title"] = *vs.Title
	}
	if vs.Description != nil {
		result["description"] = *vs.Description
	}
	if vs.Publisher != nil {
		result["publisher"] = *vs.Publisher
	}
	if vs.Date != nil {
		result["date"] = vs.Date.Format("2006-01-02")
	}
	if vs.Immutable {
		result["immutable"] = vs.Immutable
	}
	if vs.Purpose != nil {
		result["purpose"] = *vs.Purpose
	}
	if vs.Copyright != nil {
		result["copyright"] = *vs.Copyright
	}
	if vs.ComposeIncludeSystem != nil {
		include := map[string]interface{}{
			"system": *vs.ComposeIncludeSystem,
		}
		if vs.ComposeIncludeVersion != nil {
			include["version"] = *vs.ComposeIncludeVersion
		}
		result["compose"] = map[string]interface{}{
			"include": []map[string]interface{}{include},
		}
	}
	if vs.ExpansionIdentifier != nil || vs.ExpansionTimestamp != nil {
		expansion := map[string]interface{}{}
		if vs.ExpansionIdentifier != nil {
			expansion["identifier"] = *vs.ExpansionIdentifier
		}
		if vs.ExpansionTimestamp != nil {
			expansion["timestamp"] = vs.ExpansionTimestamp.Format(time.RFC3339)
		}
		result["expansion"] = expansion
	}
	return result
}
