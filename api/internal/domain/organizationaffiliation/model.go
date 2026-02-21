package organizationaffiliation

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// OrganizationAffiliation maps to the organization_affiliation table (FHIR OrganizationAffiliation resource).
type OrganizationAffiliation struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	FHIRID             string     `db:"fhir_id" json:"fhir_id"`
	Active             bool       `db:"active" json:"active"`
	OrganizationID     *uuid.UUID `db:"organization_id" json:"organization_id,omitempty"`
	ParticipatingOrgID *uuid.UUID `db:"participating_org_id" json:"participating_org_id,omitempty"`
	PeriodStart        *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd          *time.Time `db:"period_end" json:"period_end,omitempty"`
	CodeCode           *string    `db:"code_code" json:"code_code,omitempty"`
	CodeDisplay        *string    `db:"code_display" json:"code_display,omitempty"`
	SpecialtyCode      *string    `db:"specialty_code" json:"specialty_code,omitempty"`
	SpecialtyDisplay   *string    `db:"specialty_display" json:"specialty_display,omitempty"`
	LocationID         *uuid.UUID `db:"location_id" json:"location_id,omitempty"`
	TelecomPhone       *string    `db:"telecom_phone" json:"telecom_phone,omitempty"`
	TelecomEmail       *string    `db:"telecom_email" json:"telecom_email,omitempty"`
	VersionID          int        `db:"version_id" json:"version_id"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

func (o *OrganizationAffiliation) GetVersionID() int  { return o.VersionID }
func (o *OrganizationAffiliation) SetVersionID(v int)  { o.VersionID = v }

func (o *OrganizationAffiliation) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "OrganizationAffiliation",
		"id":           o.FHIRID,
		"active":       o.Active,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", o.VersionID),
			LastUpdated: o.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/OrganizationAffiliation"},
		},
	}
	if o.OrganizationID != nil {
		result["organization"] = fhir.Reference{Reference: fhir.FormatReference("Organization", o.OrganizationID.String())}
	}
	if o.ParticipatingOrgID != nil {
		result["participatingOrganization"] = fhir.Reference{Reference: fhir.FormatReference("Organization", o.ParticipatingOrgID.String())}
	}
	if o.PeriodStart != nil || o.PeriodEnd != nil {
		result["period"] = fhir.Period{Start: o.PeriodStart, End: o.PeriodEnd}
	}
	if o.CodeCode != nil {
		result["code"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *o.CodeCode, Display: strVal(o.CodeDisplay)}}}}
	}
	if o.SpecialtyCode != nil {
		result["specialty"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *o.SpecialtyCode, Display: strVal(o.SpecialtyDisplay)}}}}
	}
	if o.LocationID != nil {
		result["location"] = []fhir.Reference{{Reference: fhir.FormatReference("Location", o.LocationID.String())}}
	}
	var telecom []fhir.ContactPoint
	if o.TelecomPhone != nil {
		telecom = append(telecom, fhir.ContactPoint{System: "phone", Value: *o.TelecomPhone})
	}
	if o.TelecomEmail != nil {
		telecom = append(telecom, fhir.ContactPoint{System: "email", Value: *o.TelecomEmail})
	}
	if len(telecom) > 0 {
		result["telecom"] = telecom
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
