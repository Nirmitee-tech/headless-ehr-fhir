package episodeofcare

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// EpisodeOfCare maps to the episode_of_care table (FHIR EpisodeOfCare resource).
type EpisodeOfCare struct {
	ID                   uuid.UUID  `db:"id" json:"id"`
	FHIRID               string     `db:"fhir_id" json:"fhir_id"`
	Status               string     `db:"status" json:"status"`
	TypeCode             *string    `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay          *string    `db:"type_display" json:"type_display,omitempty"`
	DiagnosisConditionID *uuid.UUID `db:"diagnosis_condition_id" json:"diagnosis_condition_id,omitempty"`
	DiagnosisRole        *string    `db:"diagnosis_role" json:"diagnosis_role,omitempty"`
	PatientID            uuid.UUID  `db:"patient_id" json:"patient_id"`
	ManagingOrgID        *uuid.UUID `db:"managing_org_id" json:"managing_org_id,omitempty"`
	PeriodStart          *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd            *time.Time `db:"period_end" json:"period_end,omitempty"`
	ReferralRequestID    *uuid.UUID `db:"referral_request_id" json:"referral_request_id,omitempty"`
	CareManagerID        *uuid.UUID `db:"care_manager_id" json:"care_manager_id,omitempty"`
	VersionID            int        `db:"version_id" json:"version_id"`
	CreatedAt            time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (e *EpisodeOfCare) GetVersionID() int { return e.VersionID }

// SetVersionID sets the current version.
func (e *EpisodeOfCare) SetVersionID(v int) { e.VersionID = v }

func (e *EpisodeOfCare) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "EpisodeOfCare",
		"id":           e.FHIRID,
		"status":       e.Status,
		"patient":      fhir.Reference{Reference: fhir.FormatReference("Patient", e.PatientID.String())},
		"meta":         fhir.Meta{LastUpdated: e.UpdatedAt},
	}
	if e.TypeCode != nil {
		result["type"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *e.TypeCode, Display: strVal(e.TypeDisplay)}}}}
	}
	if e.DiagnosisConditionID != nil {
		diag := map[string]interface{}{
			"condition": fhir.Reference{Reference: fhir.FormatReference("Condition", e.DiagnosisConditionID.String())},
		}
		if e.DiagnosisRole != nil {
			diag["role"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *e.DiagnosisRole}}}
		}
		result["diagnosis"] = []map[string]interface{}{diag}
	}
	if e.PeriodStart != nil || e.PeriodEnd != nil {
		result["period"] = fhir.Period{Start: e.PeriodStart, End: e.PeriodEnd}
	}
	if e.ManagingOrgID != nil {
		result["managingOrganization"] = fhir.Reference{Reference: fhir.FormatReference("Organization", e.ManagingOrgID.String())}
	}
	if e.CareManagerID != nil {
		result["careManager"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", e.CareManagerID.String())}
	}
	if e.ReferralRequestID != nil {
		result["referralRequest"] = []fhir.Reference{{Reference: fhir.FormatReference("ServiceRequest", e.ReferralRequestID.String())}}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
