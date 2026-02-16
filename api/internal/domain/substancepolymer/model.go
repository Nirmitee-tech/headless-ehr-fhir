package substancepolymer

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type SubstancePolymer struct {
	ID                          uuid.UUID `db:"id" json:"id"`
	FHIRID                      string    `db:"fhir_id" json:"fhir_id"`
	ClassCode                   *string   `db:"class_code" json:"class_code,omitempty"`
	ClassDisplay                *string   `db:"class_display" json:"class_display,omitempty"`
	GeometryCode                *string   `db:"geometry_code" json:"geometry_code,omitempty"`
	GeometryDisplay             *string   `db:"geometry_display" json:"geometry_display,omitempty"`
	CopolymerConnectivityCode   *string   `db:"copolymer_connectivity_code" json:"copolymer_connectivity_code,omitempty"`
	CopolymerConnectivityDisplay *string  `db:"copolymer_connectivity_display" json:"copolymer_connectivity_display,omitempty"`
	Modification                *string   `db:"modification" json:"modification,omitempty"`
	VersionID                   int       `db:"version_id" json:"version_id"`
	CreatedAt                   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                   time.Time `db:"updated_at" json:"updated_at"`
}

func (m *SubstancePolymer) GetVersionID() int  { return m.VersionID }
func (m *SubstancePolymer) SetVersionID(v int) { m.VersionID = v }

func (m *SubstancePolymer) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "SubstancePolymer",
		"id":           m.FHIRID,
		"meta":         fhir.Meta{LastUpdated: m.UpdatedAt},
	}
	if m.ClassCode != nil { result["class"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.ClassCode, Display: strVal(m.ClassDisplay)}}} }
	if m.GeometryCode != nil { result["geometry"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.GeometryCode, Display: strVal(m.GeometryDisplay)}}} }
	if m.CopolymerConnectivityCode != nil { result["copolymerConnectivity"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *m.CopolymerConnectivityCode, Display: strVal(m.CopolymerConnectivityDisplay)}}}} }
	if m.Modification != nil { result["modification"] = []string{*m.Modification} }
	return result
}

func strVal(s *string) string { if s == nil { return "" }; return *s }
