package deviceusestatement

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// DeviceUseStatement maps to the device_use_statement table (FHIR DeviceUseStatement resource).
type DeviceUseStatement struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	FHIRID            string     `db:"fhir_id" json:"fhir_id"`
	Status            string     `db:"status" json:"status"`
	SubjectPatientID  uuid.UUID  `db:"subject_patient_id" json:"subject_patient_id"`
	DeviceID          *uuid.UUID `db:"device_id" json:"device_id,omitempty"`
	TimingDate        *time.Time `db:"timing_date" json:"timing_date,omitempty"`
	TimingPeriodStart *time.Time `db:"timing_period_start" json:"timing_period_start,omitempty"`
	TimingPeriodEnd   *time.Time `db:"timing_period_end" json:"timing_period_end,omitempty"`
	RecordedOn        *time.Time `db:"recorded_on" json:"recorded_on,omitempty"`
	SourceID          *uuid.UUID `db:"source_id" json:"source_id,omitempty"`
	ReasonCode        *string    `db:"reason_code" json:"reason_code,omitempty"`
	ReasonDisplay     *string    `db:"reason_display" json:"reason_display,omitempty"`
	BodySiteCode      *string    `db:"body_site_code" json:"body_site_code,omitempty"`
	BodySiteDisplay   *string    `db:"body_site_display" json:"body_site_display,omitempty"`
	Note              *string    `db:"note" json:"note,omitempty"`
	VersionID         int        `db:"version_id" json:"version_id"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

func (d *DeviceUseStatement) GetVersionID() int  { return d.VersionID }
func (d *DeviceUseStatement) SetVersionID(v int) { d.VersionID = v }

func (d *DeviceUseStatement) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "DeviceUseStatement",
		"id":           d.FHIRID,
		"status":       d.Status,
		"subject":      fhir.Reference{Reference: fhir.FormatReference("Patient", d.SubjectPatientID.String())},
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", d.VersionID),
			LastUpdated: d.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/DeviceUseStatement"},
		},
	}
	if d.DeviceID != nil {
		result["device"] = fhir.Reference{Reference: fhir.FormatReference("Device", d.DeviceID.String())}
	}
	if d.TimingDate != nil {
		result["timingDateTime"] = d.TimingDate.Format(time.RFC3339)
	}
	if d.TimingPeriodStart != nil || d.TimingPeriodEnd != nil {
		result["timingPeriod"] = fhir.Period{Start: d.TimingPeriodStart, End: d.TimingPeriodEnd}
	}
	if d.RecordedOn != nil {
		result["recordedOn"] = d.RecordedOn.Format(time.RFC3339)
	}
	if d.SourceID != nil {
		result["source"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", d.SourceID.String())}
	}
	if d.ReasonCode != nil {
		result["reasonCode"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *d.ReasonCode, Display: strVal(d.ReasonDisplay)}}}}
	}
	if d.BodySiteCode != nil {
		result["bodySite"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *d.BodySiteCode, Display: strVal(d.BodySiteDisplay)}}}
	}
	if d.Note != nil {
		result["note"] = []map[string]interface{}{{"text": *d.Note}}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
