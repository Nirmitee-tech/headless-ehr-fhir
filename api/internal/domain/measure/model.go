package measure

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Measure maps to the measure table (FHIR Measure resource).
type Measure struct {
	ID                   uuid.UUID  `db:"id" json:"id"`
	FHIRID               string     `db:"fhir_id" json:"fhir_id"`
	Status               string     `db:"status" json:"status"`
	URL                  *string    `db:"url" json:"url,omitempty"`
	Name                 *string    `db:"name" json:"name,omitempty"`
	Title                *string    `db:"title" json:"title,omitempty"`
	Description          *string    `db:"description" json:"description,omitempty"`
	Publisher            *string    `db:"publisher" json:"publisher,omitempty"`
	Date                 *time.Time `db:"date" json:"date,omitempty"`
	EffectivePeriodStart *time.Time `db:"effective_period_start" json:"effective_period_start,omitempty"`
	EffectivePeriodEnd   *time.Time `db:"effective_period_end" json:"effective_period_end,omitempty"`
	ScoringCode          *string    `db:"scoring_code" json:"scoring_code,omitempty"`
	ScoringDisplay       *string    `db:"scoring_display" json:"scoring_display,omitempty"`
	SubjectCode          *string    `db:"subject_code" json:"subject_code,omitempty"`
	SubjectDisplay       *string    `db:"subject_display" json:"subject_display,omitempty"`
	ApprovalDate         *time.Time `db:"approval_date" json:"approval_date,omitempty"`
	LastReviewDate       *time.Time `db:"last_review_date" json:"last_review_date,omitempty"`
	VersionID            int        `db:"version_id" json:"version_id"`
	CreatedAt            time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updated_at"`
}

func (m *Measure) GetVersionID() int  { return m.VersionID }
func (m *Measure) SetVersionID(v int) { m.VersionID = v }

func (m *Measure) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Measure",
		"id":           m.FHIRID,
		"status":       m.Status,
		"meta":         fhir.Meta{LastUpdated: m.UpdatedAt},
	}
	if m.URL != nil {
		result["url"] = *m.URL
	}
	if m.Name != nil {
		result["name"] = *m.Name
	}
	if m.Title != nil {
		result["title"] = *m.Title
	}
	if m.Description != nil {
		result["description"] = *m.Description
	}
	if m.Publisher != nil {
		result["publisher"] = *m.Publisher
	}
	if m.Date != nil {
		result["date"] = m.Date.Format("2006-01-02")
	}
	if m.EffectivePeriodStart != nil || m.EffectivePeriodEnd != nil {
		result["effectivePeriod"] = fhir.Period{Start: m.EffectivePeriodStart, End: m.EffectivePeriodEnd}
	}
	if m.ScoringCode != nil {
		result["scoring"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.ScoringCode, Display: strVal(m.ScoringDisplay)}}}
	}
	if m.SubjectCode != nil {
		result["subject"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.SubjectCode, Display: strVal(m.SubjectDisplay)}}}
	}
	if m.ApprovalDate != nil {
		result["approvalDate"] = m.ApprovalDate.Format("2006-01-02")
	}
	if m.LastReviewDate != nil {
		result["lastReviewDate"] = m.LastReviewDate.Format("2006-01-02")
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
