package specimendefinition

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// SpecimenDefinition maps to the specimen_definition table (FHIR SpecimenDefinition resource).
type SpecimenDefinition struct {
	ID                       uuid.UUID `db:"id" json:"id"`
	FHIRID                   string    `db:"fhir_id" json:"fhir_id"`
	TypeCode                 *string   `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay              *string   `db:"type_display" json:"type_display,omitempty"`
	PatientPreparation       *string   `db:"patient_preparation" json:"patient_preparation,omitempty"`
	TimeAspect               *string   `db:"time_aspect" json:"time_aspect,omitempty"`
	CollectionCode           *string   `db:"collection_code" json:"collection_code,omitempty"`
	CollectionDisplay        *string   `db:"collection_display" json:"collection_display,omitempty"`
	HandlingTemperatureLow   *float64  `db:"handling_temperature_low" json:"handling_temperature_low,omitempty"`
	HandlingTemperatureHigh  *float64  `db:"handling_temperature_high" json:"handling_temperature_high,omitempty"`
	HandlingTemperatureUnit  *string   `db:"handling_temperature_unit" json:"handling_temperature_unit,omitempty"`
	HandlingMaxDuration      *string   `db:"handling_max_duration" json:"handling_max_duration,omitempty"`
	HandlingInstruction      *string   `db:"handling_instruction" json:"handling_instruction,omitempty"`
	VersionID                int       `db:"version_id" json:"version_id"`
	CreatedAt                time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                time.Time `db:"updated_at" json:"updated_at"`
}

func (s *SpecimenDefinition) GetVersionID() int  { return s.VersionID }
func (s *SpecimenDefinition) SetVersionID(v int) { s.VersionID = v }

func (s *SpecimenDefinition) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "SpecimenDefinition",
		"id":           s.FHIRID,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", s.VersionID),
			LastUpdated: s.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/SpecimenDefinition"},
		},
	}
	if s.TypeCode != nil {
		result["typeCollected"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *s.TypeCode, Display: strVal(s.TypeDisplay)}}}
	}
	if s.PatientPreparation != nil {
		result["patientPreparation"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *s.PatientPreparation}}}}
	}
	if s.TimeAspect != nil {
		result["timeAspect"] = *s.TimeAspect
	}
	if s.CollectionCode != nil {
		result["collection"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *s.CollectionCode, Display: strVal(s.CollectionDisplay)}}}}
	}
	if s.HandlingTemperatureLow != nil || s.HandlingTemperatureHigh != nil || s.HandlingMaxDuration != nil || s.HandlingInstruction != nil {
		handling := map[string]interface{}{}
		if s.HandlingTemperatureLow != nil || s.HandlingTemperatureHigh != nil {
			tempRange := map[string]interface{}{}
			if s.HandlingTemperatureLow != nil {
				low := map[string]interface{}{"value": *s.HandlingTemperatureLow}
				if s.HandlingTemperatureUnit != nil {
					low["unit"] = *s.HandlingTemperatureUnit
				}
				tempRange["low"] = low
			}
			if s.HandlingTemperatureHigh != nil {
				high := map[string]interface{}{"value": *s.HandlingTemperatureHigh}
				if s.HandlingTemperatureUnit != nil {
					high["unit"] = *s.HandlingTemperatureUnit
				}
				tempRange["high"] = high
			}
			handling["temperatureRange"] = tempRange
		}
		if s.HandlingMaxDuration != nil {
			handling["maxDuration"] = map[string]interface{}{"value": *s.HandlingMaxDuration}
		}
		if s.HandlingInstruction != nil {
			handling["instruction"] = *s.HandlingInstruction
		}
		result["typeTested"] = []map[string]interface{}{
			{"handling": []map[string]interface{}{handling}},
		}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
