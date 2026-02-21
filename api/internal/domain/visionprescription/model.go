package visionprescription

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// VisionPrescription maps to the vision_prescription table (FHIR VisionPrescription resource).
type VisionPrescription struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	FHIRID       string     `db:"fhir_id" json:"fhir_id"`
	Status       string     `db:"status" json:"status"`
	Created      *time.Time `db:"created" json:"created,omitempty"`
	PatientID    uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID  *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	DateWritten  *time.Time `db:"date_written" json:"date_written,omitempty"`
	PrescriberID *uuid.UUID `db:"prescriber_id" json:"prescriber_id,omitempty"`
	VersionID    int        `db:"version_id" json:"version_id"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (v *VisionPrescription) GetVersionID() int { return v.VersionID }

// SetVersionID sets the current version.
func (v *VisionPrescription) SetVersionID(ver int) { v.VersionID = ver }

// ToFHIR converts the VisionPrescription to a FHIR-compliant map.
func (v *VisionPrescription) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "VisionPrescription",
		"id":           v.FHIRID,
		"status":       v.Status,
		"patient":      fhir.Reference{Reference: fhir.FormatReference("Patient", v.PatientID.String())},
		"meta":         fhir.Meta{
			LastUpdated: v.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/VisionPrescription"},
		},
	}
	if v.Created != nil {
		result["created"] = v.Created.Format("2006-01-02T15:04:05Z")
	}
	if v.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", v.EncounterID.String())}
	}
	if v.DateWritten != nil {
		result["dateWritten"] = v.DateWritten.Format("2006-01-02")
	}
	if v.PrescriberID != nil {
		result["prescriber"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", v.PrescriberID.String())}
	}
	return result
}

// VisionPrescriptionLensSpec maps to the vision_prescription_lens_spec table.
type VisionPrescriptionLensSpec struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	PrescriptionID uuid.UUID  `db:"prescription_id" json:"prescription_id"`
	ProductCode    string     `db:"product_code" json:"product_code"`
	ProductDisplay *string    `db:"product_display" json:"product_display,omitempty"`
	Eye            string     `db:"eye" json:"eye"`
	Sphere         *float64   `db:"sphere" json:"sphere,omitempty"`
	Cylinder       *float64   `db:"cylinder" json:"cylinder,omitempty"`
	Axis           *int       `db:"axis" json:"axis,omitempty"`
	PrismAmount    *float64   `db:"prism_amount" json:"prism_amount,omitempty"`
	PrismBase      *string    `db:"prism_base" json:"prism_base,omitempty"`
	AddPower       *float64   `db:"add_power" json:"add_power,omitempty"`
	Power          *float64   `db:"power" json:"power,omitempty"`
	BackCurve      *float64   `db:"back_curve" json:"back_curve,omitempty"`
	Diameter       *float64   `db:"diameter" json:"diameter,omitempty"`
	DurationValue  *float64   `db:"duration_value" json:"duration_value,omitempty"`
	DurationUnit   *string    `db:"duration_unit" json:"duration_unit,omitempty"`
	Color          *string    `db:"color" json:"color,omitempty"`
	Brand          *string    `db:"brand" json:"brand,omitempty"`
	Note           *string    `db:"note" json:"note,omitempty"`
}

// Validate checks required fields on VisionPrescriptionLensSpec.
func (ls *VisionPrescriptionLensSpec) Validate() error {
	if ls.ProductCode == "" {
		return fmt.Errorf("product_code is required")
	}
	if ls.Eye != "right" && ls.Eye != "left" {
		return fmt.Errorf("eye must be 'right' or 'left'")
	}
	return nil
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
