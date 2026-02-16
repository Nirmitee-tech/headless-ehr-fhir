package media

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Media maps to the media table (FHIR Media resource).
type Media struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	FHIRID           string     `db:"fhir_id" json:"fhir_id"`
	Status           string     `db:"status" json:"status"`
	TypeCode         *string    `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay      *string    `db:"type_display" json:"type_display,omitempty"`
	ModalityCode     *string    `db:"modality_code" json:"modality_code,omitempty"`
	ModalityDisplay  *string    `db:"modality_display" json:"modality_display,omitempty"`
	SubjectPatientID *uuid.UUID `db:"subject_patient_id" json:"subject_patient_id,omitempty"`
	EncounterID      *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	CreatedDate      *time.Time `db:"created_date" json:"created_date,omitempty"`
	OperatorID       *uuid.UUID `db:"operator_id" json:"operator_id,omitempty"`
	ReasonCode       *string    `db:"reason_code" json:"reason_code,omitempty"`
	BodySiteCode     *string    `db:"body_site_code" json:"body_site_code,omitempty"`
	BodySiteDisplay  *string    `db:"body_site_display" json:"body_site_display,omitempty"`
	DeviceName       *string    `db:"device_name" json:"device_name,omitempty"`
	Height           *int       `db:"height" json:"height,omitempty"`
	Width            *int       `db:"width" json:"width,omitempty"`
	Frames           *int       `db:"frames" json:"frames,omitempty"`
	Duration         *float64   `db:"duration" json:"duration,omitempty"`
	ContentType      *string    `db:"content_type" json:"content_type,omitempty"`
	ContentURL       *string    `db:"content_url" json:"content_url,omitempty"`
	ContentSize      *int       `db:"content_size" json:"content_size,omitempty"`
	ContentTitle     *string    `db:"content_title" json:"content_title,omitempty"`
	Note             *string    `db:"note" json:"note,omitempty"`
	VersionID        int        `db:"version_id" json:"version_id"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

func (m *Media) GetVersionID() int  { return m.VersionID }
func (m *Media) SetVersionID(v int) { m.VersionID = v }

func (m *Media) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Media",
		"id":           m.FHIRID,
		"status":       m.Status,
		"meta":         fhir.Meta{LastUpdated: m.UpdatedAt},
	}
	if m.TypeCode != nil {
		result["type"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.TypeCode, Display: strVal(m.TypeDisplay)}}}
	}
	if m.ModalityCode != nil {
		result["modality"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.ModalityCode, Display: strVal(m.ModalityDisplay)}}}
	}
	if m.SubjectPatientID != nil {
		result["subject"] = fhir.Reference{Reference: fhir.FormatReference("Patient", m.SubjectPatientID.String())}
	}
	if m.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", m.EncounterID.String())}
	}
	if m.CreatedDate != nil {
		result["createdDateTime"] = m.CreatedDate.Format(time.RFC3339)
	}
	if m.OperatorID != nil {
		result["operator"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", m.OperatorID.String())}
	}
	if m.ReasonCode != nil {
		result["reasonCode"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *m.ReasonCode}}}}
	}
	if m.BodySiteCode != nil {
		result["bodySite"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.BodySiteCode, Display: strVal(m.BodySiteDisplay)}}}
	}
	if m.DeviceName != nil {
		result["deviceName"] = *m.DeviceName
	}
	if m.Height != nil {
		result["height"] = *m.Height
	}
	if m.Width != nil {
		result["width"] = *m.Width
	}
	if m.Frames != nil {
		result["frames"] = *m.Frames
	}
	if m.Duration != nil {
		result["duration"] = *m.Duration
	}
	content := map[string]interface{}{}
	if m.ContentType != nil {
		content["contentType"] = *m.ContentType
	}
	if m.ContentURL != nil {
		content["url"] = *m.ContentURL
	}
	if m.ContentSize != nil {
		content["size"] = *m.ContentSize
	}
	if m.ContentTitle != nil {
		content["title"] = *m.ContentTitle
	}
	if len(content) > 0 {
		result["content"] = content
	}
	if m.Note != nil {
		result["note"] = []map[string]interface{}{{"text": *m.Note}}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
