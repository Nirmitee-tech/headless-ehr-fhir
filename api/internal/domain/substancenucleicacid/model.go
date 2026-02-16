package substancenucleicacid

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type SubstanceNucleicAcid struct {
	ID                       uuid.UUID `db:"id" json:"id"`
	FHIRID                   string    `db:"fhir_id" json:"fhir_id"`
	SequenceTypeCode         *string   `db:"sequence_type_code" json:"sequence_type_code,omitempty"`
	SequenceTypeDisplay      *string   `db:"sequence_type_display" json:"sequence_type_display,omitempty"`
	NumberOfSubunits         *int      `db:"number_of_subunits" json:"number_of_subunits,omitempty"`
	AreaOfHybridisation      *string   `db:"area_of_hybridisation" json:"area_of_hybridisation,omitempty"`
	OligoNucleotideTypeCode  *string   `db:"oligo_nucleotide_type_code" json:"oligo_nucleotide_type_code,omitempty"`
	OligoNucleotideTypeDisplay *string `db:"oligo_nucleotide_type_display" json:"oligo_nucleotide_type_display,omitempty"`
	VersionID                int       `db:"version_id" json:"version_id"`
	CreatedAt                time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                time.Time `db:"updated_at" json:"updated_at"`
}

func (m *SubstanceNucleicAcid) GetVersionID() int  { return m.VersionID }
func (m *SubstanceNucleicAcid) SetVersionID(v int) { m.VersionID = v }

func (m *SubstanceNucleicAcid) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "SubstanceNucleicAcid",
		"id":           m.FHIRID,
		"meta":         fhir.Meta{LastUpdated: m.UpdatedAt},
	}
	if m.SequenceTypeCode != nil { result["sequenceType"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.SequenceTypeCode, Display: strVal(m.SequenceTypeDisplay)}}} }
	if m.NumberOfSubunits != nil { result["numberOfSubunits"] = *m.NumberOfSubunits }
	if m.AreaOfHybridisation != nil { result["areaOfHybridisation"] = *m.AreaOfHybridisation }
	if m.OligoNucleotideTypeCode != nil { result["oligoNucleotideType"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.OligoNucleotideTypeCode, Display: strVal(m.OligoNucleotideTypeDisplay)}}} }
	return result
}

func strVal(s *string) string { if s == nil { return "" }; return *s }
