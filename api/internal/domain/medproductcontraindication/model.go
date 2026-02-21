package medproductcontraindication

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type MedicinalProductContraindication struct {
	ID                            uuid.UUID `db:"id" json:"id"`
	FHIRID                        string    `db:"fhir_id" json:"fhir_id"`
	SubjectReference              *string   `db:"subject_reference" json:"subject_reference,omitempty"`
	DiseaseCode                   *string   `db:"disease_code" json:"disease_code,omitempty"`
	DiseaseDisplay                *string   `db:"disease_display" json:"disease_display,omitempty"`
	DiseaseStatusCode             *string   `db:"disease_status_code" json:"disease_status_code,omitempty"`
	DiseaseStatusDisplay          *string   `db:"disease_status_display" json:"disease_status_display,omitempty"`
	ComorbidityCode               *string   `db:"comorbidity_code" json:"comorbidity_code,omitempty"`
	ComorbidityDisplay            *string   `db:"comorbidity_display" json:"comorbidity_display,omitempty"`
	TherapeuticIndicationRef      *string   `db:"therapeutic_indication_reference" json:"therapeutic_indication_reference,omitempty"`
	PopulationAgeLow              *float64  `db:"population_age_low" json:"population_age_low,omitempty"`
	PopulationAgeHigh             *float64  `db:"population_age_high" json:"population_age_high,omitempty"`
	PopulationGenderCode          *string   `db:"population_gender_code" json:"population_gender_code,omitempty"`
	VersionID                     int       `db:"version_id" json:"version_id"`
	CreatedAt                     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                     time.Time `db:"updated_at" json:"updated_at"`
}

func (m *MedicinalProductContraindication) GetVersionID() int  { return m.VersionID }
func (m *MedicinalProductContraindication) SetVersionID(v int) { m.VersionID = v }

func (m *MedicinalProductContraindication) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "MedicinalProductContraindication",
		"id":           m.FHIRID,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", m.VersionID),
			LastUpdated: m.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/MedicinalProductContraindication"},
		},
	}
	if m.SubjectReference != nil {
		result["subject"] = []fhir.Reference{{Reference: *m.SubjectReference}}
	}
	if m.DiseaseCode != nil {
		result["disease"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.DiseaseCode, Display: strVal(m.DiseaseDisplay)}}}
	}
	if m.DiseaseStatusCode != nil {
		result["diseaseStatus"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.DiseaseStatusCode, Display: strVal(m.DiseaseStatusDisplay)}}}
	}
	if m.ComorbidityCode != nil {
		result["comorbidity"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *m.ComorbidityCode, Display: strVal(m.ComorbidityDisplay)}}}}
	}
	if m.TherapeuticIndicationRef != nil {
		result["therapeuticIndication"] = []fhir.Reference{{Reference: *m.TherapeuticIndicationRef}}
	}
	if m.PopulationAgeLow != nil || m.PopulationGenderCode != nil {
		pop := map[string]interface{}{}
		if m.PopulationAgeLow != nil {
			ageRange := map[string]interface{}{"low": map[string]interface{}{"value": *m.PopulationAgeLow, "unit": "years"}}
			if m.PopulationAgeHigh != nil {
				ageRange["high"] = map[string]interface{}{"value": *m.PopulationAgeHigh, "unit": "years"}
			}
			pop["ageRange"] = ageRange
		}
		if m.PopulationGenderCode != nil {
			pop["gender"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.PopulationGenderCode}}}
		}
		result["population"] = []interface{}{pop}
	}
	return result
}

func strVal(s *string) string {
	if s == nil { return "" }
	return *s
}
