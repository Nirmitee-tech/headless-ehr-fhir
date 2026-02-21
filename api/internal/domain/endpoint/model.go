package endpoint

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Endpoint maps to the endpoint table (FHIR Endpoint resource).
type Endpoint struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	FHIRID                string     `db:"fhir_id" json:"fhir_id"`
	Status                string     `db:"status" json:"status"`
	ConnectionTypeCode    *string    `db:"connection_type_code" json:"connection_type_code,omitempty"`
	ConnectionTypeDisplay *string    `db:"connection_type_display" json:"connection_type_display,omitempty"`
	Name                  *string    `db:"name" json:"name,omitempty"`
	ManagingOrgID         *uuid.UUID `db:"managing_org_id" json:"managing_org_id,omitempty"`
	ContactPhone          *string    `db:"contact_phone" json:"contact_phone,omitempty"`
	ContactEmail          *string    `db:"contact_email" json:"contact_email,omitempty"`
	PeriodStart           *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd             *time.Time `db:"period_end" json:"period_end,omitempty"`
	PayloadTypeCode       *string    `db:"payload_type_code" json:"payload_type_code,omitempty"`
	PayloadTypeDisplay    *string    `db:"payload_type_display" json:"payload_type_display,omitempty"`
	PayloadMimeType       *string    `db:"payload_mime_type" json:"payload_mime_type,omitempty"`
	Address               string     `db:"address" json:"address"`
	Header                *string    `db:"header" json:"header,omitempty"`
	VersionID             int        `db:"version_id" json:"version_id"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

func (e *Endpoint) GetVersionID() int  { return e.VersionID }
func (e *Endpoint) SetVersionID(v int) { e.VersionID = v }

func (e *Endpoint) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Endpoint",
		"id":           e.FHIRID,
		"status":       e.Status,
		"address":      e.Address,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", e.VersionID),
			LastUpdated: e.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Endpoint"},
		},
	}
	if e.ConnectionTypeCode != nil {
		result["connectionType"] = fhir.Coding{Code: *e.ConnectionTypeCode, Display: strVal(e.ConnectionTypeDisplay)}
	}
	if e.Name != nil {
		result["name"] = *e.Name
	}
	if e.ManagingOrgID != nil {
		result["managingOrganization"] = fhir.Reference{Reference: fhir.FormatReference("Organization", e.ManagingOrgID.String())}
	}
	if e.PeriodStart != nil || e.PeriodEnd != nil {
		result["period"] = fhir.Period{Start: e.PeriodStart, End: e.PeriodEnd}
	}
	if e.PayloadTypeCode != nil {
		result["payloadType"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *e.PayloadTypeCode, Display: strVal(e.PayloadTypeDisplay)}}}}
	}
	if e.PayloadMimeType != nil {
		result["payloadMimeType"] = []string{*e.PayloadMimeType}
	}
	if e.Header != nil {
		result["header"] = []string{*e.Header}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
