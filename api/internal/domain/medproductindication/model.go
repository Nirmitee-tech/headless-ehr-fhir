package medproductindication

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type MedicinalProductIndication struct {
	ID                              uuid.UUID `db:"id" json:"id"`
	FHIRID                          string    `db:"fhir_id" json:"fhir_id"`
	SubjectReference                *string   `db:"subject_reference" json:"subject_reference,omitempty"`
	DiseaseSymptomProcedureCode     *string   `db:"disease_symptom_procedure_code" json:"disease_symptom_procedure_code,omitempty"`
	DiseaseSymptomProcedureDisplay  *string   `db:"disease_symptom_procedure_display" json:"disease_symptom_procedure_display,omitempty"`
	DiseaseStatusCode               *string   `db:"disease_status_code" json:"disease_status_code,omitempty"`
	DiseaseStatusDisplay            *string   `db:"disease_status_display" json:"disease_status_display,omitempty"`
	ComorbidityCode                 *string   `db:"comorbidity_code" json:"comorbidity_code,omitempty"`
	ComorbidityDisplay              *string   `db:"comorbidity_display" json:"comorbidity_display,omitempty"`
	IntendedEffectCode              *string   `db:"intended_effect_code" json:"intended_effect_code,omitempty"`
	IntendedEffectDisplay           *string   `db:"intended_effect_display" json:"intended_effect_display,omitempty"`
	DurationValue                   *float64  `db:"duration_value" json:"duration_value,omitempty"`
	DurationUnit                    *string   `db:"duration_unit" json:"duration_unit,omitempty"`
	UndesirableEffectReference      *string   `db:"undesirable_effect_reference" json:"undesirable_effect_reference,omitempty"`
	PopulationAgeLow                *float64  `db:"population_age_low" json:"population_age_low,omitempty"`
	PopulationAgeHigh               *float64  `db:"population_age_high" json:"population_age_high,omitempty"`
	PopulationGenderCode            *string   `db:"population_gender_code" json:"population_gender_code,omitempty"`
	VersionID                       int       `db:"version_id" json:"version_id"`
	CreatedAt                       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                       time.Time `db:"updated_at" json:"updated_at"`
}

func (m *MedicinalProductIndication) GetVersionID() int  { return m.VersionID }
func (m *MedicinalProductIndication) SetVersionID(v int) { m.VersionID = v }

func (m *MedicinalProductIndication) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "MedicinalProductIndication",
		"id":           m.FHIRID,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", m.VersionID),
			LastUpdated: m.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/MedicinalProductIndication"},
		},
	}
	if m.SubjectReference != nil { result["subject"] = []fhir.Reference{{Reference: *m.SubjectReference}} }
	if m.DiseaseSymptomProcedureCode != nil { result["diseaseSymptomProcedure"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.DiseaseSymptomProcedureCode, Display: strVal(m.DiseaseSymptomProcedureDisplay)}}} }
	if m.DiseaseStatusCode != nil { result["diseaseStatus"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.DiseaseStatusCode, Display: strVal(m.DiseaseStatusDisplay)}}} }
	if m.ComorbidityCode != nil { result["comorbidity"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *m.ComorbidityCode, Display: strVal(m.ComorbidityDisplay)}}}} }
	if m.IntendedEffectCode != nil { result["intendedEffect"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.IntendedEffectCode, Display: strVal(m.IntendedEffectDisplay)}}} }
	if m.DurationValue != nil {
		d := map[string]interface{}{"value": *m.DurationValue}
		if m.DurationUnit != nil { d["unit"] = *m.DurationUnit }
		result["duration"] = d
	}
	if m.UndesirableEffectReference != nil { result["undesirableEffect"] = []fhir.Reference{{Reference: *m.UndesirableEffectReference}} }
	if m.PopulationAgeLow != nil || m.PopulationGenderCode != nil {
		pop := map[string]interface{}{}
		if m.PopulationAgeLow != nil { ageRange := map[string]interface{}{"low": map[string]interface{}{"value": *m.PopulationAgeLow, "unit": "years"}}; if m.PopulationAgeHigh != nil { ageRange["high"] = map[string]interface{}{"value": *m.PopulationAgeHigh, "unit": "years"} }; pop["ageRange"] = ageRange }
		if m.PopulationGenderCode != nil { pop["gender"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.PopulationGenderCode}}} }
		result["population"] = []interface{}{pop}
	}
	return result
}

func strVal(s *string) string { if s == nil { return "" }; return *s }
