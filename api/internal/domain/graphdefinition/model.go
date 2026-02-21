package graphdefinition

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// GraphDefinition maps to the graph_definition table (FHIR GraphDefinition resource).
type GraphDefinition struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	FHIRID      string     `db:"fhir_id" json:"fhir_id"`
	Status      string     `db:"status" json:"status"`
	URL         *string    `db:"url" json:"url,omitempty"`
	Name        string     `db:"name" json:"name"`
	Description *string    `db:"description" json:"description,omitempty"`
	Publisher   *string    `db:"publisher" json:"publisher,omitempty"`
	Date        *time.Time `db:"date" json:"date,omitempty"`
	StartType   string     `db:"start_type" json:"start_type"`
	Profile     *string    `db:"profile" json:"profile,omitempty"`
	VersionID   int        `db:"version_id" json:"version_id"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

func (g *GraphDefinition) GetVersionID() int  { return g.VersionID }
func (g *GraphDefinition) SetVersionID(v int) { g.VersionID = v }

func (g *GraphDefinition) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "GraphDefinition",
		"id":           g.FHIRID,
		"name":         g.Name,
		"status":       g.Status,
		"start":        g.StartType,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", g.VersionID),
			LastUpdated: g.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/GraphDefinition"},
		},
	}
	if g.URL != nil {
		result["url"] = *g.URL
	}
	if g.Description != nil {
		result["description"] = *g.Description
	}
	if g.Publisher != nil {
		result["publisher"] = *g.Publisher
	}
	if g.Date != nil {
		result["date"] = g.Date.Format("2006-01-02")
	}
	if g.Profile != nil {
		result["profile"] = *g.Profile
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
