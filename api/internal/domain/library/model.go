package library

import (
	"encoding/base64"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Library maps to the library table (FHIR Library resource).
type Library struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	FHIRID       string     `db:"fhir_id" json:"fhir_id"`
	Status       string     `db:"status" json:"status"`
	URL          *string    `db:"url" json:"url,omitempty"`
	Name         *string    `db:"name" json:"name,omitempty"`
	Title        *string    `db:"title" json:"title,omitempty"`
	TypeCode     string     `db:"type_code" json:"type_code"`
	TypeDisplay  *string    `db:"type_display" json:"type_display,omitempty"`
	Description  *string    `db:"description" json:"description,omitempty"`
	Publisher    *string    `db:"publisher" json:"publisher,omitempty"`
	Date         *time.Time `db:"date" json:"date,omitempty"`
	ContentType  *string    `db:"content_type" json:"content_type,omitempty"`
	ContentData  []byte     `db:"content_data" json:"content_data,omitempty"`
	VersionID    int        `db:"version_id" json:"version_id"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

func (l *Library) GetVersionID() int  { return l.VersionID }
func (l *Library) SetVersionID(v int) { l.VersionID = v }

func (l *Library) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Library",
		"id":           l.FHIRID,
		"status":       l.Status,
		"type":         fhir.CodeableConcept{Coding: []fhir.Coding{{Code: l.TypeCode, Display: strVal(l.TypeDisplay)}}},
		"meta":         fhir.Meta{LastUpdated: l.UpdatedAt},
	}
	if l.URL != nil {
		result["url"] = *l.URL
	}
	if l.Name != nil {
		result["name"] = *l.Name
	}
	if l.Title != nil {
		result["title"] = *l.Title
	}
	if l.Description != nil {
		result["description"] = *l.Description
	}
	if l.Publisher != nil {
		result["publisher"] = *l.Publisher
	}
	if l.Date != nil {
		result["date"] = l.Date.Format("2006-01-02")
	}
	if l.ContentType != nil || l.ContentData != nil {
		attachment := map[string]interface{}{}
		if l.ContentType != nil {
			attachment["contentType"] = *l.ContentType
		}
		if l.ContentData != nil {
			attachment["data"] = base64.StdEncoding.EncodeToString(l.ContentData)
		}
		result["content"] = []map[string]interface{}{attachment}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
