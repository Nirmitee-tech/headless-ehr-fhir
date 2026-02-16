package devicedefinition

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// DeviceDefinition maps to the device_definition table (FHIR DeviceDefinition resource).
type DeviceDefinition struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	FHIRID             string     `db:"fhir_id" json:"fhir_id"`
	ManufacturerString *string    `db:"manufacturer_string" json:"manufacturer_string,omitempty"`
	ModelNumber        *string    `db:"model_number" json:"model_number,omitempty"`
	DeviceName         *string    `db:"device_name" json:"device_name,omitempty"`
	DeviceNameType     *string    `db:"device_name_type" json:"device_name_type,omitempty"`
	TypeCode           *string    `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay        *string    `db:"type_display" json:"type_display,omitempty"`
	Specialization     *string    `db:"specialization" json:"specialization,omitempty"`
	SafetyCode         *string    `db:"safety_code" json:"safety_code,omitempty"`
	SafetyDisplay      *string    `db:"safety_display" json:"safety_display,omitempty"`
	OwnerID            *uuid.UUID `db:"owner_id" json:"owner_id,omitempty"`
	ParentDeviceID     *uuid.UUID `db:"parent_device_id" json:"parent_device_id,omitempty"`
	Description        *string    `db:"description" json:"description,omitempty"`
	VersionID          int        `db:"version_id" json:"version_id"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

func (d *DeviceDefinition) GetVersionID() int  { return d.VersionID }
func (d *DeviceDefinition) SetVersionID(v int) { d.VersionID = v }

func (d *DeviceDefinition) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "DeviceDefinition",
		"id":           d.FHIRID,
		"meta":         fhir.Meta{LastUpdated: d.UpdatedAt},
	}
	if d.ManufacturerString != nil {
		result["manufacturerString"] = *d.ManufacturerString
	}
	if d.ModelNumber != nil {
		result["modelNumber"] = *d.ModelNumber
	}
	if d.DeviceName != nil {
		entry := map[string]interface{}{"name": *d.DeviceName}
		if d.DeviceNameType != nil {
			entry["type"] = *d.DeviceNameType
		}
		result["deviceName"] = []map[string]interface{}{entry}
	}
	if d.TypeCode != nil {
		result["type"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *d.TypeCode, Display: strVal(d.TypeDisplay)}}}
	}
	if d.Specialization != nil {
		result["specialization"] = *d.Specialization
	}
	if d.SafetyCode != nil {
		result["safety"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *d.SafetyCode, Display: strVal(d.SafetyDisplay)}}}}
	}
	if d.OwnerID != nil {
		result["owner"] = fhir.Reference{Reference: fhir.FormatReference("Organization", d.OwnerID.String())}
	}
	if d.ParentDeviceID != nil {
		result["parentDevice"] = fhir.Reference{Reference: fhir.FormatReference("DeviceDefinition", d.ParentDeviceID.String())}
	}
	if d.Description != nil {
		result["description"] = *d.Description
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
