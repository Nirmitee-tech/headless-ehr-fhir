package devicemetric

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// DeviceMetric maps to the device_metric table (FHIR DeviceMetric resource).
type DeviceMetric struct {
	ID                     uuid.UUID  `db:"id" json:"id"`
	FHIRID                 string     `db:"fhir_id" json:"fhir_id"`
	TypeCode               string     `db:"type_code" json:"type_code"`
	TypeDisplay            *string    `db:"type_display" json:"type_display,omitempty"`
	SourceID               *uuid.UUID `db:"source_id" json:"source_id,omitempty"`
	ParentID               *uuid.UUID `db:"parent_id" json:"parent_id,omitempty"`
	UnitCode               *string    `db:"unit_code" json:"unit_code,omitempty"`
	UnitDisplay            *string    `db:"unit_display" json:"unit_display,omitempty"`
	OperationalStatus      *string    `db:"operational_status" json:"operational_status,omitempty"`
	Color                  *string    `db:"color" json:"color,omitempty"`
	Category               string     `db:"category" json:"category"`
	CalibrationType        *string    `db:"calibration_type" json:"calibration_type,omitempty"`
	CalibrationState       *string    `db:"calibration_state" json:"calibration_state,omitempty"`
	CalibrationTime        *time.Time `db:"calibration_time" json:"calibration_time,omitempty"`
	MeasurementPeriodValue *float64   `db:"measurement_period_value" json:"measurement_period_value,omitempty"`
	MeasurementPeriodUnit  *string    `db:"measurement_period_unit" json:"measurement_period_unit,omitempty"`
	VersionID              int        `db:"version_id" json:"version_id"`
	CreatedAt              time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at" json:"updated_at"`
}

func (m *DeviceMetric) GetVersionID() int  { return m.VersionID }
func (m *DeviceMetric) SetVersionID(v int) { m.VersionID = v }

func (m *DeviceMetric) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "DeviceMetric",
		"id":           m.FHIRID,
		"type":         fhir.CodeableConcept{Coding: []fhir.Coding{{Code: m.TypeCode, Display: strVal(m.TypeDisplay)}}},
		"category":     m.Category,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", m.VersionID),
			LastUpdated: m.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/DeviceMetric"},
		},
	}
	if m.SourceID != nil {
		result["source"] = fhir.Reference{Reference: fhir.FormatReference("Device", m.SourceID.String())}
	}
	if m.ParentID != nil {
		result["parent"] = fhir.Reference{Reference: fhir.FormatReference("Device", m.ParentID.String())}
	}
	if m.UnitCode != nil {
		result["unit"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.UnitCode, Display: strVal(m.UnitDisplay)}}}
	}
	if m.OperationalStatus != nil {
		result["operationalStatus"] = *m.OperationalStatus
	}
	if m.Color != nil {
		result["color"] = *m.Color
	}
	if m.CalibrationType != nil || m.CalibrationState != nil || m.CalibrationTime != nil {
		cal := map[string]interface{}{}
		if m.CalibrationType != nil {
			cal["type"] = *m.CalibrationType
		}
		if m.CalibrationState != nil {
			cal["state"] = *m.CalibrationState
		}
		if m.CalibrationTime != nil {
			cal["time"] = m.CalibrationTime.Format("2006-01-02T15:04:05Z")
		}
		result["calibration"] = []map[string]interface{}{cal}
	}
	if m.MeasurementPeriodValue != nil {
		timing := map[string]interface{}{}
		repeat := map[string]interface{}{}
		repeat["period"] = *m.MeasurementPeriodValue
		if m.MeasurementPeriodUnit != nil {
			repeat["periodUnit"] = *m.MeasurementPeriodUnit
		}
		timing["repeat"] = repeat
		result["measurementPeriod"] = timing
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
