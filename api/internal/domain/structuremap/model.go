package structuremap

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// StructureMap maps to the structure_map table (FHIR StructureMap resource).
type StructureMap struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	FHIRID        string     `db:"fhir_id" json:"fhir_id"`
	Status        string     `db:"status" json:"status"`
	URL           string     `db:"url" json:"url"`
	Name          string     `db:"name" json:"name"`
	Title         *string    `db:"title" json:"title,omitempty"`
	Description   *string    `db:"description" json:"description,omitempty"`
	Publisher     *string    `db:"publisher" json:"publisher,omitempty"`
	Date          *time.Time `db:"date" json:"date,omitempty"`
	StructureURL  *string    `db:"structure_url" json:"structure_url,omitempty"`
	StructureMode *string    `db:"structure_mode" json:"structure_mode,omitempty"`
	ImportURI     *string    `db:"import_uri" json:"import_uri,omitempty"`
	VersionID     int        `db:"version_id" json:"version_id"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

func (sm *StructureMap) GetVersionID() int  { return sm.VersionID }
func (sm *StructureMap) SetVersionID(v int) { sm.VersionID = v }

func (sm *StructureMap) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "StructureMap",
		"id":           sm.FHIRID,
		"status":       sm.Status,
		"url":          sm.URL,
		"name":         sm.Name,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", sm.VersionID),
			LastUpdated: sm.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/StructureMap"},
		},
	}
	if sm.Title != nil {
		result["title"] = *sm.Title
	}
	if sm.Description != nil {
		result["description"] = *sm.Description
	}
	if sm.Publisher != nil {
		result["publisher"] = *sm.Publisher
	}
	if sm.Date != nil {
		result["date"] = sm.Date.Format("2006-01-02")
	}
	if sm.StructureURL != nil {
		structure := map[string]interface{}{
			"url": *sm.StructureURL,
		}
		if sm.StructureMode != nil {
			structure["mode"] = *sm.StructureMode
		}
		result["structure"] = []map[string]interface{}{structure}
	}
	if sm.ImportURI != nil {
		result["import"] = []string{*sm.ImportURI}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
