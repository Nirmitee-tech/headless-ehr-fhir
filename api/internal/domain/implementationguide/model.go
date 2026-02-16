package implementationguide

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// ImplementationGuide maps to the implementation_guide table (FHIR ImplementationGuide resource).
type ImplementationGuide struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	FHIRID        string     `db:"fhir_id" json:"fhir_id"`
	Status        string     `db:"status" json:"status"`
	URL           string     `db:"url" json:"url"`
	Name          string     `db:"name" json:"name"`
	Title         *string    `db:"title" json:"title,omitempty"`
	Description   *string    `db:"description" json:"description,omitempty"`
	Publisher     *string    `db:"publisher" json:"publisher,omitempty"`
	Date          *time.Time `db:"date" json:"date,omitempty"`
	PackageID     *string    `db:"package_id" json:"package_id,omitempty"`
	FHIRVersion   *string    `db:"fhir_version" json:"fhir_version,omitempty"`
	License       *string    `db:"license" json:"license,omitempty"`
	DependsOnURI  *string    `db:"depends_on_uri" json:"depends_on_uri,omitempty"`
	GlobalType    *string    `db:"global_type" json:"global_type,omitempty"`
	GlobalProfile *string    `db:"global_profile" json:"global_profile,omitempty"`
	VersionID     int        `db:"version_id" json:"version_id"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

func (ig *ImplementationGuide) GetVersionID() int  { return ig.VersionID }
func (ig *ImplementationGuide) SetVersionID(v int) { ig.VersionID = v }

func (ig *ImplementationGuide) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ImplementationGuide",
		"id":           ig.FHIRID,
		"url":          ig.URL,
		"name":         ig.Name,
		"status":       ig.Status,
		"meta":         fhir.Meta{LastUpdated: ig.UpdatedAt},
	}
	if ig.Title != nil {
		result["title"] = *ig.Title
	}
	if ig.Description != nil {
		result["description"] = *ig.Description
	}
	if ig.Publisher != nil {
		result["publisher"] = *ig.Publisher
	}
	if ig.Date != nil {
		result["date"] = ig.Date.Format("2006-01-02")
	}
	if ig.PackageID != nil {
		result["packageId"] = *ig.PackageID
	}
	if ig.FHIRVersion != nil {
		result["fhirVersion"] = []string{*ig.FHIRVersion}
	}
	if ig.License != nil {
		result["license"] = *ig.License
	}
	if ig.DependsOnURI != nil {
		result["dependsOn"] = []map[string]interface{}{
			{"uri": *ig.DependsOnURI},
		}
	}
	if ig.GlobalType != nil && ig.GlobalProfile != nil {
		result["global"] = []map[string]interface{}{
			{"type": *ig.GlobalType, "profile": *ig.GlobalProfile},
		}
	}
	return result
}
