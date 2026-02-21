package substanceprotein

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type SubstanceProtein struct {
	ID                uuid.UUID `db:"id" json:"id"`
	FHIRID            string    `db:"fhir_id" json:"fhir_id"`
	SequenceTypeCode  *string   `db:"sequence_type_code" json:"sequence_type_code,omitempty"`
	SequenceTypeDisplay *string `db:"sequence_type_display" json:"sequence_type_display,omitempty"`
	NumberOfSubunits  *int      `db:"number_of_subunits" json:"number_of_subunits,omitempty"`
	DisulfideLinkage  *string   `db:"disulfide_linkage" json:"disulfide_linkage,omitempty"`
	VersionID         int       `db:"version_id" json:"version_id"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}

func (m *SubstanceProtein) GetVersionID() int  { return m.VersionID }
func (m *SubstanceProtein) SetVersionID(v int) { m.VersionID = v }

func (m *SubstanceProtein) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "SubstanceProtein",
		"id":           m.FHIRID,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", m.VersionID),
			LastUpdated: m.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/SubstanceProtein"},
		},
	}
	if m.SequenceTypeCode != nil { result["sequenceType"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.SequenceTypeCode, Display: strVal(m.SequenceTypeDisplay)}}} }
	if m.NumberOfSubunits != nil { result["numberOfSubunits"] = *m.NumberOfSubunits }
	if m.DisulfideLinkage != nil { result["disulfideLinkage"] = []string{*m.DisulfideLinkage} }
	return result
}

func strVal(s *string) string { if s == nil { return "" }; return *s }
