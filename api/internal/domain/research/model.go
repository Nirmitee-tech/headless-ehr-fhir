package research

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// ResearchStudy maps to the research_study table (FHIR ResearchStudy resource).
type ResearchStudy struct {
	ID                       uuid.UUID  `db:"id" json:"id"`
	FHIRID                   string     `db:"fhir_id" json:"fhir_id"`
	Title                    string     `db:"title" json:"title"`
	ProtocolNumber           string     `db:"protocol_number" json:"protocol_number"`
	Status                   string     `db:"status" json:"status"`
	Phase                    *string    `db:"phase" json:"phase,omitempty"`
	Category                 *string    `db:"category" json:"category,omitempty"`
	Focus                    *string    `db:"focus" json:"focus,omitempty"`
	Description              *string    `db:"description" json:"description,omitempty"`
	SponsorName              *string    `db:"sponsor_name" json:"sponsor_name,omitempty"`
	SponsorContact           *string    `db:"sponsor_contact" json:"sponsor_contact,omitempty"`
	PrincipalInvestigatorID  *uuid.UUID `db:"principal_investigator_id" json:"principal_investigator_id,omitempty"`
	SiteName                 *string    `db:"site_name" json:"site_name,omitempty"`
	SiteContact              *string    `db:"site_contact" json:"site_contact,omitempty"`
	IRBNumber                *string    `db:"irb_number" json:"irb_number,omitempty"`
	IRBApprovalDate          *time.Time `db:"irb_approval_date" json:"irb_approval_date,omitempty"`
	IRBExpirationDate        *time.Time `db:"irb_expiration_date" json:"irb_expiration_date,omitempty"`
	StartDate                *time.Time `db:"start_date" json:"start_date,omitempty"`
	EndDate                  *time.Time `db:"end_date" json:"end_date,omitempty"`
	EnrollmentTarget         *int       `db:"enrollment_target" json:"enrollment_target,omitempty"`
	EnrollmentActual         *int       `db:"enrollment_actual" json:"enrollment_actual,omitempty"`
	PrimaryEndpoint          *string    `db:"primary_endpoint" json:"primary_endpoint,omitempty"`
	SecondaryEndpoints       *string    `db:"secondary_endpoints" json:"secondary_endpoints,omitempty"`
	InclusionCriteria        *string    `db:"inclusion_criteria" json:"inclusion_criteria,omitempty"`
	ExclusionCriteria        *string    `db:"exclusion_criteria" json:"exclusion_criteria,omitempty"`
	Note                     *string    `db:"note" json:"note,omitempty"`
	VersionID                int        `db:"version_id" json:"version_id"`
	CreatedAt                time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt                time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (s *ResearchStudy) GetVersionID() int { return s.VersionID }

// SetVersionID sets the current version.
func (s *ResearchStudy) SetVersionID(v int) { s.VersionID = v }

func (s *ResearchStudy) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ResearchStudy",
		"id":           s.FHIRID,
		"title":        s.Title,
		"status":       mapStudyStatusToFHIR(s.Status),
		"identifier": []fhir.Identifier{{
			Use:    "official",
			System: "urn:ehr:research:protocol",
			Value:  s.ProtocolNumber,
		}},
		"meta": fhir.Meta{
			LastUpdated: s.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/ResearchStudy"},
		},
	}
	if s.Phase != nil {
		result["phase"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System: "http://terminology.hl7.org/CodeSystem/research-study-phase",
				Code:   *s.Phase,
			}},
		}
	}
	if s.Category != nil {
		result["category"] = []fhir.CodeableConcept{{
			Text: *s.Category,
		}}
	}
	if s.Description != nil {
		result["description"] = *s.Description
	}
	if s.SponsorName != nil {
		result["sponsor"] = fhir.Reference{Display: *s.SponsorName}
	}
	if s.PrincipalInvestigatorID != nil {
		result["principalInvestigator"] = fhir.Reference{
			Reference: fhir.FormatReference("Practitioner", s.PrincipalInvestigatorID.String()),
		}
	}
	if s.StartDate != nil || s.EndDate != nil {
		period := fhir.Period{}
		if s.StartDate != nil {
			period.Start = s.StartDate
		}
		if s.EndDate != nil {
			period.End = s.EndDate
		}
		result["period"] = period
	}
	if s.EnrollmentTarget != nil {
		result["enrollment"] = []fhir.Reference{{
			Display: strVal(intToStr(s.EnrollmentTarget)) + " target participants",
		}}
	}
	if s.Note != nil {
		result["note"] = []map[string]string{{"text": *s.Note}}
	}
	return result
}

// ResearchArm maps to the research_arm table.
type ResearchArm struct {
	ID               uuid.UUID `db:"id" json:"id"`
	StudyID          uuid.UUID `db:"study_id" json:"study_id"`
	Name             string    `db:"name" json:"name"`
	ArmType          *string   `db:"arm_type" json:"arm_type,omitempty"`
	Description      *string   `db:"description" json:"description,omitempty"`
	TargetEnrollment *int      `db:"target_enrollment" json:"target_enrollment,omitempty"`
	ActualEnrollment *int      `db:"actual_enrollment" json:"actual_enrollment,omitempty"`
}

// ResearchEnrollment maps to the research_enrollment table.
type ResearchEnrollment struct {
	ID                  uuid.UUID  `db:"id" json:"id"`
	StudyID             uuid.UUID  `db:"study_id" json:"study_id"`
	ArmID               *uuid.UUID `db:"arm_id" json:"arm_id,omitempty"`
	PatientID           uuid.UUID  `db:"patient_id" json:"patient_id"`
	ConsentID           *uuid.UUID `db:"consent_id" json:"consent_id,omitempty"`
	Status              string     `db:"status" json:"status"`
	EnrolledDate        *time.Time `db:"enrolled_date" json:"enrolled_date,omitempty"`
	ScreeningDate       *time.Time `db:"screening_date" json:"screening_date,omitempty"`
	RandomizationDate   *time.Time `db:"randomization_date" json:"randomization_date,omitempty"`
	CompletionDate      *time.Time `db:"completion_date" json:"completion_date,omitempty"`
	WithdrawalDate      *time.Time `db:"withdrawal_date" json:"withdrawal_date,omitempty"`
	WithdrawalReason    *string    `db:"withdrawal_reason" json:"withdrawal_reason,omitempty"`
	RandomizationNumber *string    `db:"randomization_number" json:"randomization_number,omitempty"`
	SubjectNumber       *string    `db:"subject_number" json:"subject_number,omitempty"`
	EnrolledByID        *uuid.UUID `db:"enrolled_by_id" json:"enrolled_by_id,omitempty"`
	Note                *string    `db:"note" json:"note,omitempty"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}

// ResearchAdverseEvent maps to the research_adverse_event table.
type ResearchAdverseEvent struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	EnrollmentID      uuid.UUID  `db:"enrollment_id" json:"enrollment_id"`
	EventDate         time.Time  `db:"event_date" json:"event_date"`
	ReportedDate      time.Time  `db:"reported_date" json:"reported_date"`
	ReportedByID      *uuid.UUID `db:"reported_by_id" json:"reported_by_id,omitempty"`
	Description       string     `db:"description" json:"description"`
	Severity          *string    `db:"severity" json:"severity,omitempty"`
	Seriousness       *string    `db:"seriousness" json:"seriousness,omitempty"`
	Causality         *string    `db:"causality" json:"causality,omitempty"`
	Expectedness      *string    `db:"expectedness" json:"expectedness,omitempty"`
	Outcome           *string    `db:"outcome" json:"outcome,omitempty"`
	ActionTaken       *string    `db:"action_taken" json:"action_taken,omitempty"`
	ResolutionDate    *time.Time `db:"resolution_date" json:"resolution_date,omitempty"`
	ReportedToIRB     *bool      `db:"reported_to_irb" json:"reported_to_irb,omitempty"`
	IRBReportDate     *time.Time `db:"irb_report_date" json:"irb_report_date,omitempty"`
	ReportedToSponsor *bool      `db:"reported_to_sponsor" json:"reported_to_sponsor,omitempty"`
	SponsorReportDate *time.Time `db:"sponsor_report_date" json:"sponsor_report_date,omitempty"`
	Note              *string    `db:"note" json:"note,omitempty"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

// ResearchProtocolDeviation maps to the research_protocol_deviation table.
type ResearchProtocolDeviation struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	EnrollmentID      uuid.UUID  `db:"enrollment_id" json:"enrollment_id"`
	DeviationDate     time.Time  `db:"deviation_date" json:"deviation_date"`
	ReportedDate      time.Time  `db:"reported_date" json:"reported_date"`
	ReportedByID      *uuid.UUID `db:"reported_by_id" json:"reported_by_id,omitempty"`
	Category          *string    `db:"category" json:"category,omitempty"`
	Description       string     `db:"description" json:"description"`
	Severity          *string    `db:"severity" json:"severity,omitempty"`
	CorrectiveAction  *string    `db:"corrective_action" json:"corrective_action,omitempty"`
	PreventiveAction  *string    `db:"preventive_action" json:"preventive_action,omitempty"`
	ImpactOnSubject   *string    `db:"impact_on_subject" json:"impact_on_subject,omitempty"`
	ImpactOnStudy     *string    `db:"impact_on_study" json:"impact_on_study,omitempty"`
	ReportedToIRB     *bool      `db:"reported_to_irb" json:"reported_to_irb,omitempty"`
	IRBReportDate     *time.Time `db:"irb_report_date" json:"irb_report_date,omitempty"`
	ReportedToSponsor *bool      `db:"reported_to_sponsor" json:"reported_to_sponsor,omitempty"`
	SponsorReportDate *time.Time `db:"sponsor_report_date" json:"sponsor_report_date,omitempty"`
	Note              *string    `db:"note" json:"note,omitempty"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

// mapStudyStatusToFHIR converts internal status to FHIR ResearchStudy status.
func mapStudyStatusToFHIR(status string) string {
	mapping := map[string]string{
		"in-review":              "in-review",
		"approved":               "approved",
		"active-recruiting":      "active",
		"active-not-recruiting":  "active",
		"temporarily-closed":     "temporarily-closed-to-accrual",
		"closed":                 "closed-to-accrual",
		"completed":              "completed",
		"withdrawn":              "withdrawn",
		"suspended":              "administratively-completed",
	}
	if mapped, ok := mapping[status]; ok {
		return mapped
	}
	return status
}

func strVal(s string) string {
	return s
}

func intToStr(i *int) string {
	if i == nil {
		return ""
	}
	return fmt.Sprintf("%d", *i)
}
