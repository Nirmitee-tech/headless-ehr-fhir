package compartmentdefinition

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// CompartmentDefinition maps to the compartment_definition table (FHIR CompartmentDefinition resource).
type CompartmentDefinition struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	FHIRID        string     `db:"fhir_id" json:"fhir_id"`
	Status        string     `db:"status" json:"status"`
	URL           string     `db:"url" json:"url"`
	Name          string     `db:"name" json:"name"`
	Description   *string    `db:"description" json:"description,omitempty"`
	Publisher     *string    `db:"publisher" json:"publisher,omitempty"`
	Date          *time.Time `db:"date" json:"date,omitempty"`
	Code          string     `db:"code" json:"code"`
	Search        bool       `db:"search" json:"search"`
	ResourceType  *string    `db:"resource_type" json:"resource_type,omitempty"`
	ResourceParam *string    `db:"resource_param" json:"resource_param,omitempty"`
	VersionID     int        `db:"version_id" json:"version_id"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

func (cd *CompartmentDefinition) GetVersionID() int  { return cd.VersionID }
func (cd *CompartmentDefinition) SetVersionID(v int) { cd.VersionID = v }

func (cd *CompartmentDefinition) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "CompartmentDefinition",
		"id":           cd.FHIRID,
		"status":       cd.Status,
		"url":          cd.URL,
		"name":         cd.Name,
		"code":         cd.Code,
		"search":       cd.Search,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", cd.VersionID),
			LastUpdated: cd.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/CompartmentDefinition"},
		},
	}
	if cd.Description != nil {
		result["description"] = *cd.Description
	}
	if cd.Publisher != nil {
		result["publisher"] = *cd.Publisher
	}
	if cd.Date != nil {
		result["date"] = cd.Date.Format("2006-01-02")
	}
	if cd.ResourceType != nil {
		resource := map[string]interface{}{
			"code": *cd.ResourceType,
		}
		if cd.ResourceParam != nil {
			resource["param"] = []map[string]string{{"name": *cd.ResourceParam}}
		}
		result["resource"] = []map[string]interface{}{resource}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
