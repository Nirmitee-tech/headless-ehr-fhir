package admin

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Organization maps to the organization table.
type Organization struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	FHIRID          string     `db:"fhir_id" json:"fhir_id"`
	Name            string     `db:"name" json:"name"`
	TypeCode        string     `db:"type_code" json:"type_code"`
	Active          bool       `db:"active" json:"active"`
	ParentOrgID     *uuid.UUID `db:"parent_org_id" json:"parent_org_id,omitempty"`
	NPINumber       *string    `db:"npi_number" json:"npi_number,omitempty"`
	TINNumber       *string    `db:"tin_number" json:"tin_number,omitempty"`
	CLIANumber      *string    `db:"clia_number" json:"clia_number,omitempty"`
	RohiniID        *string    `db:"rohini_id" json:"rohini_id,omitempty"`
	ABDMFacilityID  *string    `db:"abdm_facility_id" json:"abdm_facility_id,omitempty"`
	NABHAccred      *string    `db:"nabh_accreditation" json:"nabh_accreditation,omitempty"`
	AddressLine1    *string    `db:"address_line1" json:"address_line1,omitempty"`
	AddressLine2    *string    `db:"address_line2" json:"address_line2,omitempty"`
	City            *string    `db:"city" json:"city,omitempty"`
	District        *string    `db:"district" json:"district,omitempty"`
	State           *string    `db:"state" json:"state,omitempty"`
	PostalCode      *string    `db:"postal_code" json:"postal_code,omitempty"`
	Country         *string    `db:"country" json:"country,omitempty"`
	Phone           *string    `db:"phone" json:"phone,omitempty"`
	Email           *string    `db:"email" json:"email,omitempty"`
	Website         *string    `db:"website" json:"website,omitempty"`
	VersionID       int        `db:"version_id" json:"version_id"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (o *Organization) GetVersionID() int { return o.VersionID }

// SetVersionID sets the current version.
func (o *Organization) SetVersionID(v int) { o.VersionID = v }

func (o *Organization) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Organization",
		"id":           o.FHIRID,
		"active":       o.Active,
		"name":         o.Name,
		"type": []fhir.CodeableConcept{
			{
				Coding: []fhir.Coding{{
					System:  "http://terminology.hl7.org/CodeSystem/organization-type",
					Code:    o.TypeCode,
					Display: o.TypeCode,
				}},
			},
		},
		"meta": fhir.Meta{LastUpdated: o.UpdatedAt},
	}

	var telecoms []fhir.ContactPoint
	if o.Phone != nil {
		telecoms = append(telecoms, fhir.ContactPoint{System: "phone", Value: *o.Phone})
	}
	if o.Email != nil {
		telecoms = append(telecoms, fhir.ContactPoint{System: "email", Value: *o.Email})
	}
	if len(telecoms) > 0 {
		result["telecom"] = telecoms
	}

	if o.AddressLine1 != nil {
		addr := fhir.Address{Use: "work"}
		if o.AddressLine1 != nil {
			addr.Line = append(addr.Line, *o.AddressLine1)
		}
		if o.AddressLine2 != nil {
			addr.Line = append(addr.Line, *o.AddressLine2)
		}
		if o.City != nil {
			addr.City = *o.City
		}
		if o.State != nil {
			addr.State = *o.State
		}
		if o.PostalCode != nil {
			addr.PostalCode = *o.PostalCode
		}
		if o.Country != nil {
			addr.Country = *o.Country
		}
		result["address"] = []fhir.Address{addr}
	}

	if o.ParentOrgID != nil {
		result["partOf"] = fhir.Reference{
			Reference: fhir.FormatReference("Organization", o.ParentOrgID.String()),
		}
	}

	return result
}

// Department maps to the department table.
type Department struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	OrganizationID     uuid.UUID  `db:"organization_id" json:"organization_id"`
	Name               string     `db:"name" json:"name"`
	Code               *string    `db:"code" json:"code,omitempty"`
	Description        *string    `db:"description" json:"description,omitempty"`
	HeadPractitionerID *uuid.UUID `db:"head_practitioner_id" json:"head_practitioner_id,omitempty"`
	Active             bool       `db:"active" json:"active"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
}

// Location maps to the location table.
type Location struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	FHIRID            string     `db:"fhir_id" json:"fhir_id"`
	Status            string     `db:"status" json:"status"`
	OperationalStatus *string    `db:"operational_status" json:"operational_status,omitempty"`
	Name              string     `db:"name" json:"name"`
	Description       *string    `db:"description" json:"description,omitempty"`
	Mode              *string    `db:"mode" json:"mode,omitempty"`
	TypeCode          *string    `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay       *string    `db:"type_display" json:"type_display,omitempty"`
	PhysicalTypeCode  *string    `db:"physical_type_code" json:"physical_type_code,omitempty"`
	OrganizationID    *uuid.UUID `db:"organization_id" json:"organization_id,omitempty"`
	PartOfLocationID  *uuid.UUID `db:"part_of_location_id" json:"part_of_location_id,omitempty"`
	AddressLine1      *string    `db:"address_line1" json:"address_line1,omitempty"`
	City              *string    `db:"city" json:"city,omitempty"`
	State             *string    `db:"state" json:"state,omitempty"`
	PostalCode        *string    `db:"postal_code" json:"postal_code,omitempty"`
	Country           *string    `db:"country" json:"country,omitempty"`
	Latitude          *float64   `db:"latitude" json:"latitude,omitempty"`
	Longitude         *float64   `db:"longitude" json:"longitude,omitempty"`
	Phone             *string    `db:"phone" json:"phone,omitempty"`
	Email             *string    `db:"email" json:"email,omitempty"`
	VersionID         int        `db:"version_id" json:"version_id"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
}

// GetVersionID returns the current version.
func (l *Location) GetVersionID() int { return l.VersionID }

// SetVersionID sets the current version.
func (l *Location) SetVersionID(v int) { l.VersionID = v }

func (l *Location) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Location",
		"id":           l.FHIRID,
		"status":       l.Status,
		"name":         l.Name,
		"meta":         fhir.Meta{LastUpdated: l.CreatedAt},
	}

	if l.TypeCode != nil {
		result["type"] = []fhir.CodeableConcept{
			{Coding: []fhir.Coding{{Code: *l.TypeCode, Display: strPtrVal(l.TypeDisplay)}}},
		}
	}

	if l.PhysicalTypeCode != nil {
		result["physicalType"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *l.PhysicalTypeCode}},
		}
	}

	if l.OrganizationID != nil {
		result["managingOrganization"] = fhir.Reference{
			Reference: fhir.FormatReference("Organization", l.OrganizationID.String()),
		}
	}

	if l.PartOfLocationID != nil {
		result["partOf"] = fhir.Reference{
			Reference: fhir.FormatReference("Location", l.PartOfLocationID.String()),
		}
	}

	return result
}

// SystemUser maps to the system_user table.
type SystemUser struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	Username              string     `db:"username" json:"username"`
	PractitionerID        *uuid.UUID `db:"practitioner_id" json:"practitioner_id,omitempty"`
	UserType              string     `db:"user_type" json:"user_type"`
	Status                string     `db:"status" json:"status"`
	DisplayName           *string    `db:"display_name" json:"display_name,omitempty"`
	Email                 *string    `db:"email" json:"email,omitempty"`
	Phone                 *string    `db:"phone" json:"phone,omitempty"`
	LastLogin             *time.Time `db:"last_login" json:"last_login,omitempty"`
	FailedLoginCount      int        `db:"failed_login_count" json:"failed_login_count"`
	PasswordLastChanged   *time.Time `db:"password_last_changed" json:"password_last_changed,omitempty"`
	MFAEnabled            bool       `db:"mfa_enabled" json:"mfa_enabled"`
	PrimaryDepartmentID   *uuid.UUID `db:"primary_department_id" json:"primary_department_id,omitempty"`
	EmployeeID            *string    `db:"employee_id" json:"employee_id,omitempty"`
	HireDate              *time.Time `db:"hire_date" json:"hire_date,omitempty"`
	TerminationDate       *time.Time `db:"termination_date" json:"termination_date,omitempty"`
	HIPAATrainingDate     *time.Time `db:"hipaa_training_date" json:"hipaa_training_date,omitempty"`
	LastComplianceTraining *time.Time `db:"last_compliance_training" json:"last_compliance_training,omitempty"`
	Note                  *string    `db:"note" json:"note,omitempty"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

// UserRoleAssignment maps to the user_role_assignment table.
type UserRoleAssignment struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	UserID         uuid.UUID  `db:"user_id" json:"user_id"`
	RoleName       string     `db:"role_name" json:"role_name"`
	OrganizationID *uuid.UUID `db:"organization_id" json:"organization_id,omitempty"`
	DepartmentID   *uuid.UUID `db:"department_id" json:"department_id,omitempty"`
	LocationID     *uuid.UUID `db:"location_id" json:"location_id,omitempty"`
	StartDate      time.Time  `db:"start_date" json:"start_date"`
	EndDate        *time.Time `db:"end_date" json:"end_date,omitempty"`
	Active         bool       `db:"active" json:"active"`
	GrantedByID    *uuid.UUID `db:"granted_by_id" json:"granted_by_id,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
}

func strPtrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
