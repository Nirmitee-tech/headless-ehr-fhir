package careteam

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// CareTeam maps to the care_team table (FHIR CareTeam resource).
type CareTeam struct {
	ID                     uuid.UUID            `db:"id" json:"id"`
	FHIRID                 string               `db:"fhir_id" json:"fhir_id"`
	Status                 string               `db:"status" json:"status"`
	Name                   *string              `db:"name" json:"name,omitempty"`
	PatientID              uuid.UUID            `db:"patient_id" json:"patient_id"`
	EncounterID            *uuid.UUID           `db:"encounter_id" json:"encounter_id,omitempty"`
	CategoryCode           *string              `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay        *string              `db:"category_display" json:"category_display,omitempty"`
	PeriodStart            *time.Time           `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd              *time.Time           `db:"period_end" json:"period_end,omitempty"`
	ManagingOrganizationID *uuid.UUID           `db:"managing_organization_id" json:"managing_organization_id,omitempty"`
	ReasonCode             *string              `db:"reason_code" json:"reason_code,omitempty"`
	ReasonDisplay          *string              `db:"reason_display" json:"reason_display,omitempty"`
	Note                   *string              `db:"note" json:"note,omitempty"`
	Participants           []CareTeamParticipant `json:"participants,omitempty"`
	VersionID              int                  `db:"version_id" json:"version_id"`
	CreatedAt              time.Time            `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time            `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (ct *CareTeam) GetVersionID() int { return ct.VersionID }

// SetVersionID sets the current version.
func (ct *CareTeam) SetVersionID(v int) { ct.VersionID = v }

func (ct *CareTeam) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "CareTeam",
		"id":           ct.FHIRID,
		"status":       ct.Status,
		"subject":      fhir.Reference{Reference: fhir.FormatReference("Patient", ct.PatientID.String())},
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", ct.VersionID),
			LastUpdated: ct.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/us/core/StructureDefinition/us-core-careteam"},
		},
	}
	if ct.Name != nil {
		result["name"] = *ct.Name
	}
	if ct.CategoryCode != nil {
		result["category"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *ct.CategoryCode, Display: strVal(ct.CategoryDisplay)}},
		}}
	}
	if ct.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", ct.EncounterID.String())}
	}
	if ct.PeriodStart != nil {
		period := fhir.Period{Start: ct.PeriodStart, End: ct.PeriodEnd}
		result["period"] = period
	}
	if ct.ManagingOrganizationID != nil {
		result["managingOrganization"] = []fhir.Reference{{
			Reference: fhir.FormatReference("Organization", ct.ManagingOrganizationID.String()),
		}}
	}
	if ct.ReasonCode != nil {
		result["reasonCode"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *ct.ReasonCode, Display: strVal(ct.ReasonDisplay)}},
		}}
	}
	if ct.Note != nil {
		result["note"] = []map[string]string{{"text": *ct.Note}}
	}
	if len(ct.Participants) > 0 {
		participants := make([]map[string]interface{}, len(ct.Participants))
		for i, p := range ct.Participants {
			participants[i] = p.ToFHIR()
		}
		result["participant"] = participants
	}
	return result
}

// CareTeamParticipant maps to the care_team_participant table.
type CareTeamParticipant struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	CareTeamID   uuid.UUID  `db:"care_team_id" json:"care_team_id"`
	MemberID     uuid.UUID  `db:"member_id" json:"member_id"`
	MemberType   string     `db:"member_type" json:"member_type"`
	RoleCode     *string    `db:"role_code" json:"role_code,omitempty"`
	RoleDisplay  *string    `db:"role_display" json:"role_display,omitempty"`
	PeriodStart  *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd    *time.Time `db:"period_end" json:"period_end,omitempty"`
	OnBehalfOfID *uuid.UUID `db:"on_behalf_of_id" json:"on_behalf_of_id,omitempty"`
}

func (p *CareTeamParticipant) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"member": fhir.Reference{Reference: fhir.FormatReference(p.MemberType, p.MemberID.String())},
	}
	if p.RoleCode != nil {
		result["role"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *p.RoleCode, Display: strVal(p.RoleDisplay)}},
		}}
	}
	if p.PeriodStart != nil {
		result["period"] = fhir.Period{Start: p.PeriodStart, End: p.PeriodEnd}
	}
	if p.OnBehalfOfID != nil {
		result["onBehalfOf"] = fhir.Reference{Reference: fhir.FormatReference("Organization", p.OnBehalfOfID.String())}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
