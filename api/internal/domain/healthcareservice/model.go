package healthcareservice

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// HealthcareService maps to the healthcare_service table (FHIR HealthcareService resource).
type HealthcareService struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	FHIRID                string     `db:"fhir_id" json:"fhir_id"`
	Active                bool       `db:"active" json:"active"`
	ProvidedByOrgID       *uuid.UUID `db:"provided_by_org_id" json:"provided_by_org_id,omitempty"`
	CategoryCode          *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay       *string    `db:"category_display" json:"category_display,omitempty"`
	TypeCode              *string    `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay           *string    `db:"type_display" json:"type_display,omitempty"`
	Name                  string     `db:"name" json:"name"`
	Comment               *string    `db:"comment" json:"comment,omitempty"`
	TelecomPhone          *string    `db:"telecom_phone" json:"telecom_phone,omitempty"`
	TelecomEmail          *string    `db:"telecom_email" json:"telecom_email,omitempty"`
	ServiceProvisionCode  *string    `db:"service_provision_code" json:"service_provision_code,omitempty"`
	ProgramName           *string    `db:"program_name" json:"program_name,omitempty"`
	LocationID            *uuid.UUID `db:"location_id" json:"location_id,omitempty"`
	AppointmentRequired   bool       `db:"appointment_required" json:"appointment_required"`
	AvailableTime         *string    `db:"available_time" json:"available_time,omitempty"`
	NotAvailable          *string    `db:"not_available" json:"not_available,omitempty"`
	AvailabilityExceptions *string   `db:"availability_exceptions" json:"availability_exceptions,omitempty"`
	VersionID             int        `db:"version_id" json:"version_id"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (hs *HealthcareService) GetVersionID() int { return hs.VersionID }

// SetVersionID sets the current version.
func (hs *HealthcareService) SetVersionID(v int) { hs.VersionID = v }

func (hs *HealthcareService) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType":      "HealthcareService",
		"id":                hs.FHIRID,
		"active":            hs.Active,
		"name":              hs.Name,
		"appointmentRequired": hs.AppointmentRequired,
		"meta":              fhir.Meta{
			VersionID:   fmt.Sprintf("%d", hs.VersionID),
			LastUpdated: hs.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/HealthcareService"},
		},
	}
	if hs.ProvidedByOrgID != nil {
		result["providedBy"] = fhir.Reference{Reference: fhir.FormatReference("Organization", hs.ProvidedByOrgID.String())}
	}
	if hs.CategoryCode != nil {
		result["category"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *hs.CategoryCode, Display: strVal(hs.CategoryDisplay)}}}}
	}
	if hs.TypeCode != nil {
		result["type"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *hs.TypeCode, Display: strVal(hs.TypeDisplay)}}}}
	}
	if hs.Comment != nil {
		result["comment"] = *hs.Comment
	}
	if hs.TelecomPhone != nil || hs.TelecomEmail != nil {
		var telecoms []fhir.ContactPoint
		if hs.TelecomPhone != nil {
			telecoms = append(telecoms, fhir.ContactPoint{System: "phone", Value: *hs.TelecomPhone})
		}
		if hs.TelecomEmail != nil {
			telecoms = append(telecoms, fhir.ContactPoint{System: "email", Value: *hs.TelecomEmail})
		}
		result["telecom"] = telecoms
	}
	if hs.ServiceProvisionCode != nil {
		result["serviceProvisionCode"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *hs.ServiceProvisionCode}}}}
	}
	if hs.ProgramName != nil {
		result["program"] = []fhir.CodeableConcept{{Text: *hs.ProgramName}}
	}
	if hs.LocationID != nil {
		result["location"] = []fhir.Reference{{Reference: fhir.FormatReference("Location", hs.LocationID.String())}}
	}
	if hs.AvailabilityExceptions != nil {
		result["availabilityExceptions"] = *hs.AvailabilityExceptions
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
