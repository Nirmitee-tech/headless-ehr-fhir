package device

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Device maps to the device table (FHIR Device resource).
type Device struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	FHIRID             string     `db:"fhir_id" json:"fhir_id"`
	Status             string     `db:"status" json:"status"`
	StatusReason       *string    `db:"status_reason" json:"status_reason,omitempty"`
	DistinctIdentifier *string    `db:"distinct_identifier" json:"distinct_identifier,omitempty"`
	ManufacturerName   *string    `db:"manufacturer_name" json:"manufacturer_name,omitempty"`
	ManufactureDate    *time.Time `db:"manufacture_date" json:"manufacture_date,omitempty"`
	ExpirationDate     *time.Time `db:"expiration_date" json:"expiration_date,omitempty"`
	LotNumber          *string    `db:"lot_number" json:"lot_number,omitempty"`
	SerialNumber       *string    `db:"serial_number" json:"serial_number,omitempty"`
	ModelNumber        *string    `db:"model_number" json:"model_number,omitempty"`
	DeviceName         string     `db:"device_name" json:"device_name"`
	DeviceNameType     string     `db:"device_name_type" json:"device_name_type,omitempty"`
	TypeCode           *string    `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay        *string    `db:"type_display" json:"type_display,omitempty"`
	TypeSystem         *string    `db:"type_system" json:"type_system,omitempty"`
	VersionValue       *string    `db:"version_value" json:"version_value,omitempty"`
	PatientID          *uuid.UUID `db:"patient_id" json:"patient_id,omitempty"`
	OwnerID            *uuid.UUID `db:"owner_id" json:"owner_id,omitempty"`
	LocationID         *uuid.UUID `db:"location_id" json:"location_id,omitempty"`
	ContactPhone       *string    `db:"contact_phone" json:"contact_phone,omitempty"`
	ContactEmail       *string    `db:"contact_email" json:"contact_email,omitempty"`
	URL                *string    `db:"url" json:"url,omitempty"`
	Note               *string    `db:"note" json:"note,omitempty"`
	SafetyCode         *string    `db:"safety_code" json:"safety_code,omitempty"`
	SafetyDisplay      *string    `db:"safety_display" json:"safety_display,omitempty"`
	UDICarrier         *string    `db:"udi_carrier" json:"udi_carrier,omitempty"`
	UDIEntryType       *string    `db:"udi_entry_type" json:"udi_entry_type,omitempty"`
	VersionID          int        `db:"version_id" json:"version_id"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (d *Device) GetVersionID() int { return d.VersionID }

// SetVersionID sets the current version.
func (d *Device) SetVersionID(v int) { d.VersionID = v }

func (d *Device) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Device",
		"id":           d.FHIRID,
		"status":       d.Status,
		"deviceName": []map[string]string{{
			"name": d.DeviceName,
			"type": d.DeviceNameType,
		}},
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", d.VersionID),
			LastUpdated: d.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Device"},
		},
	}
	if d.StatusReason != nil {
		result["statusReason"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *d.StatusReason}},
		}}
	}
	if d.DistinctIdentifier != nil {
		result["distinctIdentifier"] = *d.DistinctIdentifier
	}
	if d.ManufacturerName != nil {
		result["manufacturer"] = *d.ManufacturerName
	}
	if d.ManufactureDate != nil {
		result["manufactureDate"] = d.ManufactureDate.Format("2006-01-02")
	}
	if d.ExpirationDate != nil {
		result["expirationDate"] = d.ExpirationDate.Format("2006-01-02")
	}
	if d.LotNumber != nil {
		result["lotNumber"] = *d.LotNumber
	}
	if d.SerialNumber != nil {
		result["serialNumber"] = *d.SerialNumber
	}
	if d.ModelNumber != nil {
		result["modelNumber"] = *d.ModelNumber
	}
	if d.TypeCode != nil {
		result["type"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System:  strVal(d.TypeSystem),
				Code:    *d.TypeCode,
				Display: strVal(d.TypeDisplay),
			}},
		}
	}
	if d.VersionValue != nil {
		result["version"] = []map[string]string{{"value": *d.VersionValue}}
	}
	if d.PatientID != nil {
		result["patient"] = fhir.Reference{Reference: fhir.FormatReference("Patient", d.PatientID.String())}
	}
	if d.OwnerID != nil {
		result["owner"] = fhir.Reference{Reference: fhir.FormatReference("Organization", d.OwnerID.String())}
	}
	if d.LocationID != nil {
		result["location"] = fhir.Reference{Reference: fhir.FormatReference("Location", d.LocationID.String())}
	}
	if d.ContactPhone != nil || d.ContactEmail != nil {
		var contacts []fhir.ContactPoint
		if d.ContactPhone != nil {
			contacts = append(contacts, fhir.ContactPoint{System: "phone", Value: *d.ContactPhone})
		}
		if d.ContactEmail != nil {
			contacts = append(contacts, fhir.ContactPoint{System: "email", Value: *d.ContactEmail})
		}
		result["contact"] = contacts
	}
	if d.URL != nil {
		result["url"] = *d.URL
	}
	if d.Note != nil {
		result["note"] = []map[string]string{{"text": *d.Note}}
	}
	if d.SafetyCode != nil {
		result["safety"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *d.SafetyCode, Display: strVal(d.SafetyDisplay)}},
		}}
	}
	if d.UDICarrier != nil {
		udi := map[string]string{"carrierHRF": *d.UDICarrier}
		if d.UDIEntryType != nil {
			udi["entryType"] = *d.UDIEntryType
		}
		result["udiCarrier"] = []map[string]string{udi}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
