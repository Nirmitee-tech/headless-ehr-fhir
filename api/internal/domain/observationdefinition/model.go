package observationdefinition

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// ObservationDefinition maps to the observation_definition table (FHIR ObservationDefinition resource).
type ObservationDefinition struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	FHIRID                string     `db:"fhir_id" json:"fhir_id"`
	Status                string     `db:"status" json:"status"`
	CategoryCode          *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay       *string    `db:"category_display" json:"category_display,omitempty"`
	CodeCode              string     `db:"code_code" json:"code_code"`
	CodeSystem            *string    `db:"code_system" json:"code_system,omitempty"`
	CodeDisplay           *string    `db:"code_display" json:"code_display,omitempty"`
	PermittedDataType     *string    `db:"permitted_data_type" json:"permitted_data_type,omitempty"`
	MultipleResultsAllowed bool     `db:"multiple_results_allowed" json:"multiple_results_allowed"`
	MethodCode            *string    `db:"method_code" json:"method_code,omitempty"`
	MethodDisplay         *string    `db:"method_display" json:"method_display,omitempty"`
	PreferredReportName   *string    `db:"preferred_report_name" json:"preferred_report_name,omitempty"`
	UnitCode              *string    `db:"unit_code" json:"unit_code,omitempty"`
	UnitDisplay           *string    `db:"unit_display" json:"unit_display,omitempty"`
	NormalValueLow        *float64   `db:"normal_value_low" json:"normal_value_low,omitempty"`
	NormalValueHigh       *float64   `db:"normal_value_high" json:"normal_value_high,omitempty"`
	VersionID             int        `db:"version_id" json:"version_id"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

func (od *ObservationDefinition) GetVersionID() int  { return od.VersionID }
func (od *ObservationDefinition) SetVersionID(v int)  { od.VersionID = v }

func (od *ObservationDefinition) ToFHIR() map[string]interface{} {
	codeCoding := fhir.Coding{Code: od.CodeCode, Display: strVal(od.CodeDisplay)}
	if od.CodeSystem != nil {
		codeCoding.System = *od.CodeSystem
	}
	result := map[string]interface{}{
		"resourceType": "ObservationDefinition",
		"id":           od.FHIRID,
		"status":       od.Status,
		"code":         fhir.CodeableConcept{Coding: []fhir.Coding{codeCoding}},
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", od.VersionID),
			LastUpdated: od.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/ObservationDefinition"},
		},
	}
	if od.CategoryCode != nil {
		result["category"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *od.CategoryCode, Display: strVal(od.CategoryDisplay)}}}}
	}
	if od.PermittedDataType != nil {
		result["permittedDataType"] = []string{*od.PermittedDataType}
	}
	result["multipleResultsAllowed"] = od.MultipleResultsAllowed
	if od.MethodCode != nil {
		result["method"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *od.MethodCode, Display: strVal(od.MethodDisplay)}}}
	}
	if od.PreferredReportName != nil {
		result["preferredReportName"] = *od.PreferredReportName
	}
	if od.UnitCode != nil || od.UnitDisplay != nil {
		unit := fhir.Coding{Code: strVal(od.UnitCode), Display: strVal(od.UnitDisplay)}
		result["quantitativeDetails"] = map[string]interface{}{
			"unit":          unit,
			"customaryUnit": unit,
		}
	}
	if od.NormalValueLow != nil || od.NormalValueHigh != nil {
		rangeMap := map[string]interface{}{}
		if od.NormalValueLow != nil {
			rangeMap["low"] = map[string]interface{}{"value": *od.NormalValueLow}
		}
		if od.NormalValueHigh != nil {
			rangeMap["high"] = map[string]interface{}{"value": *od.NormalValueHigh}
		}
		result["qualifiedInterval"] = []map[string]interface{}{{"range": rangeMap}}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
