package immunization

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Immunization maps to the immunization table (FHIR Immunization resource).
type Immunization struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	FHIRID             string     `db:"fhir_id" json:"fhir_id"`
	Status             string     `db:"status" json:"status"`
	PatientID          uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID        *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	VaccineCodeSystem  *string    `db:"vaccine_code_system" json:"vaccine_code_system,omitempty"`
	VaccineCode        string     `db:"vaccine_code" json:"vaccine_code"`
	VaccineDisplay     string     `db:"vaccine_display" json:"vaccine_display"`
	OccurrenceDateTime *time.Time `db:"occurrence_datetime" json:"occurrence_datetime,omitempty"`
	OccurrenceString   *string    `db:"occurrence_string" json:"occurrence_string,omitempty"`
	PrimarySource      bool       `db:"primary_source" json:"primary_source"`
	LotNumber          *string    `db:"lot_number" json:"lot_number,omitempty"`
	ExpirationDate     *time.Time `db:"expiration_date" json:"expiration_date,omitempty"`
	SiteCode           *string    `db:"site_code" json:"site_code,omitempty"`
	SiteDisplay        *string    `db:"site_display" json:"site_display,omitempty"`
	RouteCode          *string    `db:"route_code" json:"route_code,omitempty"`
	RouteDisplay       *string    `db:"route_display" json:"route_display,omitempty"`
	DoseQuantity       *float64   `db:"dose_quantity" json:"dose_quantity,omitempty"`
	DoseUnit           *string    `db:"dose_unit" json:"dose_unit,omitempty"`
	PerformerID        *uuid.UUID `db:"performer_id" json:"performer_id,omitempty"`
	Note               *string    `db:"note" json:"note,omitempty"`
	VersionID          int        `db:"version_id" json:"version_id"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (im *Immunization) GetVersionID() int { return im.VersionID }

// SetVersionID sets the current version.
func (im *Immunization) SetVersionID(v int) { im.VersionID = v }

func (im *Immunization) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Immunization",
		"id":           im.FHIRID,
		"status":       im.Status,
		"vaccineCode": fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System:  strVal(im.VaccineCodeSystem, "http://hl7.org/fhir/sid/cvx"),
				Code:    im.VaccineCode,
				Display: im.VaccineDisplay,
			}},
			Text: im.VaccineDisplay,
		},
		"patient":       fhir.Reference{Reference: fhir.FormatReference("Patient", im.PatientID.String())},
		"primarySource": im.PrimarySource,
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", im.VersionID),
			LastUpdated: im.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/us/core/StructureDefinition/us-core-immunization"},
		},
	}
	if im.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", im.EncounterID.String())}
	}
	if im.OccurrenceDateTime != nil {
		result["occurrenceDateTime"] = im.OccurrenceDateTime.Format(time.RFC3339)
	} else if im.OccurrenceString != nil {
		result["occurrenceString"] = *im.OccurrenceString
	}
	if im.SiteCode != nil {
		result["site"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *im.SiteCode, Display: ptrVal(im.SiteDisplay)}},
		}
	}
	if im.RouteCode != nil {
		result["route"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *im.RouteCode, Display: ptrVal(im.RouteDisplay)}},
		}
	}
	if im.DoseQuantity != nil {
		dq := map[string]interface{}{"value": *im.DoseQuantity}
		if im.DoseUnit != nil {
			dq["unit"] = *im.DoseUnit
		}
		result["doseQuantity"] = dq
	}
	if im.PerformerID != nil {
		result["performer"] = []map[string]interface{}{
			{"actor": fhir.Reference{Reference: fhir.FormatReference("Practitioner", im.PerformerID.String())}},
		}
	}
	if im.LotNumber != nil {
		result["lotNumber"] = *im.LotNumber
	}
	if im.ExpirationDate != nil {
		result["expirationDate"] = im.ExpirationDate.Format("2006-01-02")
	}
	if im.Note != nil {
		result["note"] = []map[string]string{{"text": *im.Note}}
	}
	return result
}

// ImmunizationRecommendation maps to the immunization_recommendation table.
type ImmunizationRecommendation struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	FHIRID          string     `db:"fhir_id" json:"fhir_id"`
	PatientID       uuid.UUID  `db:"patient_id" json:"patient_id"`
	Date            time.Time  `db:"date" json:"date"`
	VaccineCode     string     `db:"vaccine_code" json:"vaccine_code"`
	VaccineDisplay  string     `db:"vaccine_display" json:"vaccine_display"`
	ForecastStatus  string     `db:"forecast_status" json:"forecast_status"`
	ForecastDisplay *string    `db:"forecast_display" json:"forecast_display,omitempty"`
	DateCriterion   *time.Time `db:"date_criterion" json:"date_criterion,omitempty"`
	SeriesDoses     *int       `db:"series_doses" json:"series_doses,omitempty"`
	DoseNumber      *int       `db:"dose_number" json:"dose_number,omitempty"`
	Description     *string    `db:"description" json:"description,omitempty"`
	VersionID       int        `db:"version_id" json:"version_id"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (r *ImmunizationRecommendation) GetVersionID() int { return r.VersionID }

// SetVersionID sets the current version.
func (r *ImmunizationRecommendation) SetVersionID(v int) { r.VersionID = v }

func (r *ImmunizationRecommendation) ToFHIR() map[string]interface{} {
	rec := map[string]interface{}{
		"vaccineCode": []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{
				System:  "http://hl7.org/fhir/sid/cvx",
				Code:    r.VaccineCode,
				Display: r.VaccineDisplay,
			}},
		}},
		"forecastStatus": fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: r.ForecastStatus, Display: ptrVal(r.ForecastDisplay)}},
		},
	}
	if r.DateCriterion != nil {
		rec["dateCriterion"] = []map[string]interface{}{
			{"value": r.DateCriterion.Format(time.RFC3339)},
		}
	}
	if r.SeriesDoses != nil {
		rec["seriesDosesPositiveInt"] = *r.SeriesDoses
	}
	if r.DoseNumber != nil {
		rec["doseNumberPositiveInt"] = *r.DoseNumber
	}
	if r.Description != nil {
		rec["description"] = *r.Description
	}

	result := map[string]interface{}{
		"resourceType":   "ImmunizationRecommendation",
		"id":             r.FHIRID,
		"patient":        fhir.Reference{Reference: fhir.FormatReference("Patient", r.PatientID.String())},
		"date":           r.Date.Format(time.RFC3339),
		"recommendation": []map[string]interface{}{rec},
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", r.VersionID),
			LastUpdated: r.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/ImmunizationRecommendation"},
		},
	}
	return result
}

func strVal(s *string, defaultVal string) string {
	if s == nil {
		return defaultVal
	}
	return *s
}

func ptrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
