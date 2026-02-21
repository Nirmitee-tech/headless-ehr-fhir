package careplan

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// CarePlan maps to the care_plan table (FHIR CarePlan resource).
type CarePlan struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	FHIRID          string     `db:"fhir_id" json:"fhir_id"`
	Status          string     `db:"status" json:"status"`
	Intent          string     `db:"intent" json:"intent"`
	CategoryCode    *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay *string    `db:"category_display" json:"category_display,omitempty"`
	Title           *string    `db:"title" json:"title,omitempty"`
	Description     *string    `db:"description" json:"description,omitempty"`
	PatientID       uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID     *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	PeriodStart     *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd       *time.Time `db:"period_end" json:"period_end,omitempty"`
	AuthorID        *uuid.UUID `db:"author_id" json:"author_id,omitempty"`
	Note            *string    `db:"note" json:"note,omitempty"`
	VersionID       int        `db:"version_id" json:"version_id"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (cp *CarePlan) GetVersionID() int { return cp.VersionID }

// SetVersionID sets the current version.
func (cp *CarePlan) SetVersionID(v int) { cp.VersionID = v }

func (cp *CarePlan) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "CarePlan",
		"id":           cp.FHIRID,
		"status":       cp.Status,
		"intent":       cp.Intent,
		"subject":      fhir.Reference{Reference: fhir.FormatReference("Patient", cp.PatientID.String())},
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", cp.VersionID),
			LastUpdated: cp.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/us/core/StructureDefinition/us-core-careplan"},
		},
	}
	if cp.CategoryCode != nil {
		result["category"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *cp.CategoryCode, Display: strVal(cp.CategoryDisplay)}},
		}}
	}
	if cp.Title != nil {
		result["title"] = *cp.Title
	}
	if cp.Description != nil {
		result["description"] = *cp.Description
	}
	if cp.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", cp.EncounterID.String())}
	}
	if cp.PeriodStart != nil {
		period := fhir.Period{Start: cp.PeriodStart, End: cp.PeriodEnd}
		result["period"] = period
	}
	if cp.AuthorID != nil {
		result["author"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", cp.AuthorID.String())}
	}
	if cp.Note != nil {
		result["note"] = []map[string]string{{"text": *cp.Note}}
	}
	return result
}

// CarePlanActivity maps to the care_plan_activity table.
type CarePlanActivity struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	CarePlanID     uuid.UUID  `db:"care_plan_id" json:"care_plan_id"`
	DetailCode     *string    `db:"detail_code" json:"detail_code,omitempty"`
	DetailDisplay  *string    `db:"detail_display" json:"detail_display,omitempty"`
	Status         string     `db:"status" json:"status"`
	ScheduledStart *time.Time `db:"scheduled_start" json:"scheduled_start,omitempty"`
	ScheduledEnd   *time.Time `db:"scheduled_end" json:"scheduled_end,omitempty"`
	Description    *string    `db:"description" json:"description,omitempty"`
}

// Goal maps to the goal table (FHIR Goal resource).
type Goal struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	FHIRID             string     `db:"fhir_id" json:"fhir_id"`
	LifecycleStatus    string     `db:"lifecycle_status" json:"lifecycle_status"`
	AchievementStatus  *string    `db:"achievement_status" json:"achievement_status,omitempty"`
	CategoryCode       *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay    *string    `db:"category_display" json:"category_display,omitempty"`
	Description        string     `db:"description" json:"description"`
	PatientID          uuid.UUID  `db:"patient_id" json:"patient_id"`
	TargetMeasure      *string    `db:"target_measure" json:"target_measure,omitempty"`
	TargetDetailString *string    `db:"target_detail_string" json:"target_detail_string,omitempty"`
	TargetDueDate      *time.Time `db:"target_due_date" json:"target_due_date,omitempty"`
	ExpressedByID      *uuid.UUID `db:"expressed_by_id" json:"expressed_by_id,omitempty"`
	Note               *string    `db:"note" json:"note,omitempty"`
	VersionID          int        `db:"version_id" json:"version_id"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (g *Goal) GetVersionID() int { return g.VersionID }

// SetVersionID sets the current version.
func (g *Goal) SetVersionID(v int) { g.VersionID = v }

func (g *Goal) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType":    "Goal",
		"id":              g.FHIRID,
		"lifecycleStatus": g.LifecycleStatus,
		"description": fhir.CodeableConcept{
			Text: g.Description,
		},
		"subject": fhir.Reference{Reference: fhir.FormatReference("Patient", g.PatientID.String())},
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", g.VersionID),
			LastUpdated: g.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/us/core/StructureDefinition/us-core-goal"},
		},
	}
	if g.AchievementStatus != nil {
		result["achievementStatus"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *g.AchievementStatus}},
		}
	}
	if g.CategoryCode != nil {
		result["category"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *g.CategoryCode, Display: strVal(g.CategoryDisplay)}},
		}}
	}
	if g.TargetMeasure != nil || g.TargetDetailString != nil || g.TargetDueDate != nil {
		target := map[string]interface{}{}
		if g.TargetMeasure != nil {
			target["measure"] = fhir.CodeableConcept{
				Coding: []fhir.Coding{{Code: *g.TargetMeasure}},
			}
		}
		if g.TargetDetailString != nil {
			target["detailString"] = *g.TargetDetailString
		}
		if g.TargetDueDate != nil {
			target["dueDate"] = g.TargetDueDate.Format("2006-01-02")
		}
		result["target"] = []map[string]interface{}{target}
	}
	if g.ExpressedByID != nil {
		result["expressedBy"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", g.ExpressedByID.String())}
	}
	if g.Note != nil {
		result["note"] = []map[string]string{{"text": *g.Note}}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
