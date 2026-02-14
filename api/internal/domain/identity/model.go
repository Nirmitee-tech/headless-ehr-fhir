package identity

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Patient maps to the patient table.
type Patient struct {
	ID                     uuid.UUID  `db:"id" json:"id"`
	FHIRID                 string     `db:"fhir_id" json:"fhir_id"`
	Active                 bool       `db:"active" json:"active"`
	MRN                    string     `db:"mrn" json:"mrn"`
	Prefix                 *string    `db:"prefix" json:"prefix,omitempty"`
	FirstName              string     `db:"first_name" json:"first_name"`
	MiddleName             *string    `db:"middle_name" json:"middle_name,omitempty"`
	LastName               string     `db:"last_name" json:"last_name"`
	Suffix                 *string    `db:"suffix" json:"suffix,omitempty"`
	MaidenName             *string    `db:"maiden_name" json:"maiden_name,omitempty"`
	BirthDate              *time.Time `db:"birth_date" json:"birth_date,omitempty"`
	Gender                 *string    `db:"gender" json:"gender,omitempty"`
	DeceasedBoolean        bool       `db:"deceased_boolean" json:"deceased_boolean"`
	DeceasedDatetime       *time.Time `db:"deceased_datetime" json:"deceased_datetime,omitempty"`
	MaritalStatus          *string    `db:"marital_status" json:"marital_status,omitempty"`
	MultipleBirth          bool       `db:"multiple_birth" json:"multiple_birth"`
	MultipleBirthInt       *int       `db:"multiple_birth_int" json:"multiple_birth_int,omitempty"`
	PhotoURL               *string    `db:"photo_url" json:"photo_url,omitempty"`
	SSNHash                *string    `db:"ssn_hash" json:"-"`
	AbhaID                 *string    `db:"abha_id" json:"abha_id,omitempty"`
	AbhaAddress            *string    `db:"abha_address" json:"abha_address,omitempty"`
	AadhaarHash            *string    `db:"aadhaar_hash" json:"-"`
	PhoneHome              *string    `db:"phone_home" json:"phone_home,omitempty"`
	PhoneMobile            *string    `db:"phone_mobile" json:"phone_mobile,omitempty"`
	PhoneWork              *string    `db:"phone_work" json:"phone_work,omitempty"`
	Email                  *string    `db:"email" json:"email,omitempty"`
	AddressUse             *string    `db:"address_use" json:"address_use,omitempty"`
	AddressLine1           *string    `db:"address_line1" json:"address_line1,omitempty"`
	AddressLine2           *string    `db:"address_line2" json:"address_line2,omitempty"`
	City                   *string    `db:"city" json:"city,omitempty"`
	District               *string    `db:"district" json:"district,omitempty"`
	State                  *string    `db:"state" json:"state,omitempty"`
	PostalCode             *string    `db:"postal_code" json:"postal_code,omitempty"`
	Country                *string    `db:"country" json:"country,omitempty"`
	PreferredLanguage      *string    `db:"preferred_language" json:"preferred_language,omitempty"`
	InterpreterNeeded      bool       `db:"interpreter_needed" json:"interpreter_needed"`
	PrimaryCareProviderID  *uuid.UUID `db:"primary_care_provider_id" json:"primary_care_provider_id,omitempty"`
	ManagingOrgID          *uuid.UUID `db:"managing_org_id" json:"managing_org_id,omitempty"`
	CreatedAt              time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at" json:"updated_at"`
}

func (p *Patient) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Patient",
		"id":           p.FHIRID,
		"active":       p.Active,
		"meta":         fhir.Meta{LastUpdated: p.UpdatedAt},
	}

	// Name
	name := fhir.HumanName{
		Use:    "official",
		Family: p.LastName,
		Given:  []string{p.FirstName},
	}
	if p.MiddleName != nil {
		name.Given = append(name.Given, *p.MiddleName)
	}
	if p.Prefix != nil {
		name.Prefix = []string{*p.Prefix}
	}
	if p.Suffix != nil {
		name.Suffix = []string{*p.Suffix}
	}
	result["name"] = []fhir.HumanName{name}

	// Identifiers
	identifiers := []fhir.Identifier{
		{
			Use:    "usual",
			Type:   &fhir.CodeableConcept{Coding: []fhir.Coding{{System: "http://terminology.hl7.org/CodeSystem/v2-0203", Code: "MR"}}},
			Value:  p.MRN,
		},
	}
	if p.AbhaID != nil {
		identifiers = append(identifiers, fhir.Identifier{
			System: "https://healthid.ndhm.gov.in",
			Value:  *p.AbhaID,
		})
	}
	result["identifier"] = identifiers

	if p.Gender != nil {
		result["gender"] = *p.Gender
	}
	if p.BirthDate != nil {
		result["birthDate"] = p.BirthDate.Format("2006-01-02")
	}
	if p.DeceasedBoolean {
		result["deceasedBoolean"] = true
	}
	if p.DeceasedDatetime != nil {
		result["deceasedDateTime"] = p.DeceasedDatetime
	}

	// Telecom
	var telecoms []fhir.ContactPoint
	if p.PhoneMobile != nil {
		telecoms = append(telecoms, fhir.ContactPoint{System: "phone", Value: *p.PhoneMobile, Use: "mobile"})
	}
	if p.PhoneHome != nil {
		telecoms = append(telecoms, fhir.ContactPoint{System: "phone", Value: *p.PhoneHome, Use: "home"})
	}
	if p.Email != nil {
		telecoms = append(telecoms, fhir.ContactPoint{System: "email", Value: *p.Email})
	}
	if len(telecoms) > 0 {
		result["telecom"] = telecoms
	}

	// Address
	if p.AddressLine1 != nil {
		addr := fhir.Address{}
		if p.AddressUse != nil {
			addr.Use = *p.AddressUse
		}
		addr.Line = []string{*p.AddressLine1}
		if p.AddressLine2 != nil {
			addr.Line = append(addr.Line, *p.AddressLine2)
		}
		if p.City != nil {
			addr.City = *p.City
		}
		if p.District != nil {
			addr.District = *p.District
		}
		if p.State != nil {
			addr.State = *p.State
		}
		if p.PostalCode != nil {
			addr.PostalCode = *p.PostalCode
		}
		if p.Country != nil {
			addr.Country = *p.Country
		}
		result["address"] = []fhir.Address{addr}
	}

	// Managing Organization
	if p.ManagingOrgID != nil {
		result["managingOrganization"] = fhir.Reference{
			Reference: fhir.FormatReference("Organization", p.ManagingOrgID.String()),
		}
	}

	// General Practitioner
	if p.PrimaryCareProviderID != nil {
		result["generalPractitioner"] = []fhir.Reference{
			{Reference: fhir.FormatReference("Practitioner", p.PrimaryCareProviderID.String())},
		}
	}

	if p.PreferredLanguage != nil {
		result["communication"] = []map[string]interface{}{
			{
				"language": fhir.CodeableConcept{
					Coding: []fhir.Coding{{Code: *p.PreferredLanguage}},
				},
				"preferred": true,
			},
		}
	}

	return result
}

// PatientContact maps to the patient_contact table.
type PatientContact struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	PatientID      uuid.UUID  `db:"patient_id" json:"patient_id"`
	Relationship   string     `db:"relationship" json:"relationship"`
	Prefix         *string    `db:"prefix" json:"prefix,omitempty"`
	FirstName      *string    `db:"first_name" json:"first_name,omitempty"`
	LastName       *string    `db:"last_name" json:"last_name,omitempty"`
	Phone          *string    `db:"phone" json:"phone,omitempty"`
	Email          *string    `db:"email" json:"email,omitempty"`
	AddressLine1   *string    `db:"address_line1" json:"address_line1,omitempty"`
	City           *string    `db:"city" json:"city,omitempty"`
	State          *string    `db:"state" json:"state,omitempty"`
	PostalCode     *string    `db:"postal_code" json:"postal_code,omitempty"`
	Country        *string    `db:"country" json:"country,omitempty"`
	Gender         *string    `db:"gender" json:"gender,omitempty"`
	IsPrimaryContact bool    `db:"is_primary_contact" json:"is_primary_contact"`
	PeriodStart    *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd      *time.Time `db:"period_end" json:"period_end,omitempty"`
}

// PatientIdentifier maps to the patient_identifier table.
type PatientIdentifier struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	PatientID   uuid.UUID  `db:"patient_id" json:"patient_id"`
	SystemURI   string     `db:"system_uri" json:"system_uri"`
	Value       string     `db:"value" json:"value"`
	TypeCode    *string    `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay *string    `db:"type_display" json:"type_display,omitempty"`
	Assigner    *string    `db:"assigner" json:"assigner,omitempty"`
	PeriodStart *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd   *time.Time `db:"period_end" json:"period_end,omitempty"`
}

// Practitioner maps to the practitioner table.
type Practitioner struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	FHIRID                string     `db:"fhir_id" json:"fhir_id"`
	Active                bool       `db:"active" json:"active"`
	Prefix                *string    `db:"prefix" json:"prefix,omitempty"`
	FirstName             string     `db:"first_name" json:"first_name"`
	MiddleName            *string    `db:"middle_name" json:"middle_name,omitempty"`
	LastName              string     `db:"last_name" json:"last_name"`
	Suffix                *string    `db:"suffix" json:"suffix,omitempty"`
	Gender                *string    `db:"gender" json:"gender,omitempty"`
	BirthDate             *time.Time `db:"birth_date" json:"birth_date,omitempty"`
	PhotoURL              *string    `db:"photo_url" json:"photo_url,omitempty"`
	NPINumber             *string    `db:"npi_number" json:"npi_number,omitempty"`
	DEANumber             *string    `db:"dea_number" json:"dea_number,omitempty"`
	StateLicenseNum       *string    `db:"state_license_num" json:"state_license_num,omitempty"`
	StateLicenseState     *string    `db:"state_license_state" json:"state_license_state,omitempty"`
	MedicalCouncilReg     *string    `db:"medical_council_reg" json:"medical_council_reg,omitempty"`
	AbhaID                *string    `db:"abha_id" json:"abha_id,omitempty"`
	HPRID                 *string    `db:"hpr_id" json:"hpr_id,omitempty"`
	Phone                 *string    `db:"phone" json:"phone,omitempty"`
	Email                 *string    `db:"email" json:"email,omitempty"`
	AddressLine1          *string    `db:"address_line1" json:"address_line1,omitempty"`
	City                  *string    `db:"city" json:"city,omitempty"`
	State                 *string    `db:"state" json:"state,omitempty"`
	PostalCode            *string    `db:"postal_code" json:"postal_code,omitempty"`
	Country               *string    `db:"country" json:"country,omitempty"`
	QualificationSummary  *string    `db:"qualification_summary" json:"qualification_summary,omitempty"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

func (p *Practitioner) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Practitioner",
		"id":           p.FHIRID,
		"active":       p.Active,
		"meta":         fhir.Meta{LastUpdated: p.UpdatedAt},
	}

	name := fhir.HumanName{
		Use:    "official",
		Family: p.LastName,
		Given:  []string{p.FirstName},
	}
	if p.MiddleName != nil {
		name.Given = append(name.Given, *p.MiddleName)
	}
	if p.Prefix != nil {
		name.Prefix = []string{*p.Prefix}
	}
	if p.Suffix != nil {
		name.Suffix = []string{*p.Suffix}
	}
	result["name"] = []fhir.HumanName{name}

	var identifiers []fhir.Identifier
	if p.NPINumber != nil {
		identifiers = append(identifiers, fhir.Identifier{
			System: "http://hl7.org/fhir/sid/us-npi",
			Value:  *p.NPINumber,
		})
	}
	if p.HPRID != nil {
		identifiers = append(identifiers, fhir.Identifier{
			System: "https://hpr.ndhm.gov.in",
			Value:  *p.HPRID,
		})
	}
	if len(identifiers) > 0 {
		result["identifier"] = identifiers
	}

	if p.Gender != nil {
		result["gender"] = *p.Gender
	}
	if p.BirthDate != nil {
		result["birthDate"] = p.BirthDate.Format("2006-01-02")
	}

	var telecoms []fhir.ContactPoint
	if p.Phone != nil {
		telecoms = append(telecoms, fhir.ContactPoint{System: "phone", Value: *p.Phone})
	}
	if p.Email != nil {
		telecoms = append(telecoms, fhir.ContactPoint{System: "email", Value: *p.Email})
	}
	if len(telecoms) > 0 {
		result["telecom"] = telecoms
	}

	if p.AddressLine1 != nil {
		addr := fhir.Address{Use: "work"}
		addr.Line = []string{*p.AddressLine1}
		if p.City != nil {
			addr.City = *p.City
		}
		if p.State != nil {
			addr.State = *p.State
		}
		if p.PostalCode != nil {
			addr.PostalCode = *p.PostalCode
		}
		if p.Country != nil {
			addr.Country = *p.Country
		}
		result["address"] = []fhir.Address{addr}
	}

	return result
}

// PractitionerRole maps to the practitioner_role table.
type PractitionerRole struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	FHIRID           string     `db:"fhir_id" json:"fhir_id"`
	PractitionerID   uuid.UUID  `db:"practitioner_id" json:"practitioner_id"`
	OrganizationID   *uuid.UUID `db:"organization_id" json:"organization_id,omitempty"`
	DepartmentID     *uuid.UUID `db:"department_id" json:"department_id,omitempty"`
	RoleCode         string     `db:"role_code" json:"role_code"`
	RoleDisplay      *string    `db:"role_display" json:"role_display,omitempty"`
	PeriodStart      *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd        *time.Time `db:"period_end" json:"period_end,omitempty"`
	Active           bool       `db:"active" json:"active"`
	TelehealthCapable bool     `db:"telehealth_capable" json:"telehealth_capable"`
	AcceptingPatients bool      `db:"accepting_patients" json:"accepting_patients"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
}
