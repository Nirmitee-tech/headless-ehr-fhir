package substancesourcematerial

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type SubstanceSourceMaterial struct {
	ID                          uuid.UUID `db:"id" json:"id"`
	FHIRID                      string    `db:"fhir_id" json:"fhir_id"`
	SourceMaterialClassCode     *string   `db:"source_material_class_code" json:"source_material_class_code,omitempty"`
	SourceMaterialClassDisplay  *string   `db:"source_material_class_display" json:"source_material_class_display,omitempty"`
	SourceMaterialTypeCode      *string   `db:"source_material_type_code" json:"source_material_type_code,omitempty"`
	SourceMaterialTypeDisplay   *string   `db:"source_material_type_display" json:"source_material_type_display,omitempty"`
	SourceMaterialStateCode     *string   `db:"source_material_state_code" json:"source_material_state_code,omitempty"`
	SourceMaterialStateDisplay  *string   `db:"source_material_state_display" json:"source_material_state_display,omitempty"`
	OrganismID                  *string   `db:"organism_id" json:"organism_id,omitempty"`
	OrganismName                *string   `db:"organism_name" json:"organism_name,omitempty"`
	CountryOfOriginCode         *string   `db:"country_of_origin_code" json:"country_of_origin_code,omitempty"`
	CountryOfOriginDisplay      *string   `db:"country_of_origin_display" json:"country_of_origin_display,omitempty"`
	GeographicalLocation        *string   `db:"geographical_location" json:"geographical_location,omitempty"`
	VersionID                   int       `db:"version_id" json:"version_id"`
	CreatedAt                   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                   time.Time `db:"updated_at" json:"updated_at"`
}

func (m *SubstanceSourceMaterial) GetVersionID() int  { return m.VersionID }
func (m *SubstanceSourceMaterial) SetVersionID(v int) { m.VersionID = v }

func (m *SubstanceSourceMaterial) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "SubstanceSourceMaterial",
		"id":           m.FHIRID,
		"meta":         fhir.Meta{LastUpdated: m.UpdatedAt},
	}
	if m.SourceMaterialClassCode != nil { result["sourceMaterialClass"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.SourceMaterialClassCode, Display: strVal(m.SourceMaterialClassDisplay)}}} }
	if m.SourceMaterialTypeCode != nil { result["sourceMaterialType"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.SourceMaterialTypeCode, Display: strVal(m.SourceMaterialTypeDisplay)}}} }
	if m.SourceMaterialStateCode != nil { result["sourceMaterialState"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.SourceMaterialStateCode, Display: strVal(m.SourceMaterialStateDisplay)}}} }
	if m.OrganismID != nil || m.OrganismName != nil {
		org := map[string]interface{}{}
		if m.OrganismID != nil { org["id"] = *m.OrganismID }
		if m.OrganismName != nil { org["name"] = *m.OrganismName }
		result["organism"] = org
	}
	if m.CountryOfOriginCode != nil { result["countryOfOrigin"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *m.CountryOfOriginCode, Display: strVal(m.CountryOfOriginDisplay)}}}} }
	if m.GeographicalLocation != nil { result["geographicalLocation"] = []string{*m.GeographicalLocation} }
	return result
}

func strVal(s *string) string { if s == nil { return "" }; return *s }
