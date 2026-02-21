package medproductundesirableeffect

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type MedicinalProductUndesirableEffect struct {
	ID                                uuid.UUID `db:"id" json:"id"`
	FHIRID                            string    `db:"fhir_id" json:"fhir_id"`
	SubjectReference                  *string   `db:"subject_reference" json:"subject_reference,omitempty"`
	SymptomConditionEffectCode        *string   `db:"symptom_condition_effect_code" json:"symptom_condition_effect_code,omitempty"`
	SymptomConditionEffectDisplay     *string   `db:"symptom_condition_effect_display" json:"symptom_condition_effect_display,omitempty"`
	ClassificationCode                *string   `db:"classification_code" json:"classification_code,omitempty"`
	ClassificationDisplay             *string   `db:"classification_display" json:"classification_display,omitempty"`
	FrequencyOfOccurrenceCode         *string   `db:"frequency_of_occurrence_code" json:"frequency_of_occurrence_code,omitempty"`
	FrequencyOfOccurrenceDisplay      *string   `db:"frequency_of_occurrence_display" json:"frequency_of_occurrence_display,omitempty"`
	PopulationAgeLow                  *float64  `db:"population_age_low" json:"population_age_low,omitempty"`
	PopulationAgeHigh                 *float64  `db:"population_age_high" json:"population_age_high,omitempty"`
	PopulationGenderCode              *string   `db:"population_gender_code" json:"population_gender_code,omitempty"`
	VersionID                         int       `db:"version_id" json:"version_id"`
	CreatedAt                         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                         time.Time `db:"updated_at" json:"updated_at"`
}

func (m *MedicinalProductUndesirableEffect) GetVersionID() int  { return m.VersionID }
func (m *MedicinalProductUndesirableEffect) SetVersionID(v int) { m.VersionID = v }

func (m *MedicinalProductUndesirableEffect) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "MedicinalProductUndesirableEffect",
		"id":           m.FHIRID,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", m.VersionID),
			LastUpdated: m.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/MedicinalProductUndesirableEffect"},
		},
	}
	if m.SubjectReference != nil { result["subject"] = []fhir.Reference{{Reference: *m.SubjectReference}} }
	if m.SymptomConditionEffectCode != nil { result["symptomConditionEffect"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.SymptomConditionEffectCode, Display: strVal(m.SymptomConditionEffectDisplay)}}} }
	if m.ClassificationCode != nil { result["classification"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.ClassificationCode, Display: strVal(m.ClassificationDisplay)}}} }
	if m.FrequencyOfOccurrenceCode != nil { result["frequencyOfOccurrence"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.FrequencyOfOccurrenceCode, Display: strVal(m.FrequencyOfOccurrenceDisplay)}}} }
	if m.PopulationAgeLow != nil || m.PopulationGenderCode != nil {
		pop := map[string]interface{}{}
		if m.PopulationAgeLow != nil { ageRange := map[string]interface{}{"low": map[string]interface{}{"value": *m.PopulationAgeLow, "unit": "years"}}; if m.PopulationAgeHigh != nil { ageRange["high"] = map[string]interface{}{"value": *m.PopulationAgeHigh, "unit": "years"} }; pop["ageRange"] = ageRange }
		if m.PopulationGenderCode != nil { pop["gender"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.PopulationGenderCode}}} }
		result["population"] = []interface{}{pop}
	}
	return result
}

func strVal(s *string) string { if s == nil { return "" }; return *s }
