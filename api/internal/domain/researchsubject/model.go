package researchsubject

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// ResearchSubject maps to the research_subject table (FHIR ResearchSubject resource).
type ResearchSubject struct {
	ID                  uuid.UUID  `db:"id" json:"id"`
	FHIRID              string     `db:"fhir_id" json:"fhir_id"`
	Status              string     `db:"status" json:"status"`
	StudyReference      *string    `db:"study_reference" json:"study_reference,omitempty"`
	IndividualReference *string    `db:"individual_reference" json:"individual_reference,omitempty"`
	ConsentReference    *string    `db:"consent_reference" json:"consent_reference,omitempty"`
	PeriodStart         *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd           *time.Time `db:"period_end" json:"period_end,omitempty"`
	AssignedArm         *string    `db:"assigned_arm" json:"assigned_arm,omitempty"`
	ActualArm           *string    `db:"actual_arm" json:"actual_arm,omitempty"`
	VersionID           int        `db:"version_id" json:"version_id"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}

func (r *ResearchSubject) GetVersionID() int  { return r.VersionID }
func (r *ResearchSubject) SetVersionID(v int) { r.VersionID = v }

func (r *ResearchSubject) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ResearchSubject",
		"id":           r.FHIRID,
		"status":       r.Status,
		"meta":         fhir.Meta{LastUpdated: r.UpdatedAt},
	}
	if r.StudyReference != nil {
		result["study"] = fhir.Reference{Reference: *r.StudyReference}
	}
	if r.IndividualReference != nil {
		result["individual"] = fhir.Reference{Reference: *r.IndividualReference}
	}
	if r.ConsentReference != nil {
		result["consent"] = fhir.Reference{Reference: *r.ConsentReference}
	}
	if r.PeriodStart != nil || r.PeriodEnd != nil {
		result["period"] = fhir.Period{Start: r.PeriodStart, End: r.PeriodEnd}
	}
	if r.AssignedArm != nil {
		result["assignedArm"] = *r.AssignedArm
	}
	if r.ActualArm != nil {
		result["actualArm"] = *r.ActualArm
	}
	return result
}
