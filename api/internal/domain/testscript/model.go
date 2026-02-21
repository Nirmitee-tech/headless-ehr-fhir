package testscript

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// TestScript maps to the test_script table (FHIR TestScript resource).
type TestScript struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	FHIRID           string     `db:"fhir_id" json:"fhir_id"`
	Status           string     `db:"status" json:"status"`
	URL              *string    `db:"url" json:"url,omitempty"`
	Name             string     `db:"name" json:"name"`
	Title            *string    `db:"title" json:"title,omitempty"`
	Description      *string    `db:"description" json:"description,omitempty"`
	Publisher        *string    `db:"publisher" json:"publisher,omitempty"`
	Date             *time.Time `db:"date" json:"date,omitempty"`
	Purpose          *string    `db:"purpose" json:"purpose,omitempty"`
	Copyright        *string    `db:"copyright" json:"copyright,omitempty"`
	ProfileReference *string    `db:"profile_reference" json:"profile_reference,omitempty"`
	OriginIndex      *int       `db:"origin_index" json:"origin_index,omitempty"`
	DestinationIndex *int       `db:"destination_index" json:"destination_index,omitempty"`
	VersionID        int        `db:"version_id" json:"version_id"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

func (ts *TestScript) GetVersionID() int  { return ts.VersionID }
func (ts *TestScript) SetVersionID(v int) { ts.VersionID = v }

func (ts *TestScript) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "TestScript",
		"id":           ts.FHIRID,
		"status":       ts.Status,
		"name":         ts.Name,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", ts.VersionID),
			LastUpdated: ts.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/TestScript"},
		},
	}
	if ts.URL != nil {
		result["url"] = *ts.URL
	}
	if ts.Title != nil {
		result["title"] = *ts.Title
	}
	if ts.Description != nil {
		result["description"] = *ts.Description
	}
	if ts.Publisher != nil {
		result["publisher"] = *ts.Publisher
	}
	if ts.Date != nil {
		result["date"] = ts.Date.Format("2006-01-02")
	}
	if ts.Purpose != nil {
		result["purpose"] = *ts.Purpose
	}
	if ts.Copyright != nil {
		result["copyright"] = *ts.Copyright
	}
	if ts.ProfileReference != nil {
		result["profile"] = []map[string]interface{}{{"reference": *ts.ProfileReference}}
	}
	if ts.OriginIndex != nil {
		result["origin"] = []map[string]interface{}{{"index": *ts.OriginIndex}}
	}
	if ts.DestinationIndex != nil {
		result["destination"] = []map[string]interface{}{{"index": *ts.DestinationIndex}}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
