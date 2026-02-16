package medproductinteraction

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type MedicinalProductInteraction struct {
	ID               uuid.UUID `db:"id" json:"id"`
	FHIRID           string    `db:"fhir_id" json:"fhir_id"`
	SubjectReference *string   `db:"subject_reference" json:"subject_reference,omitempty"`
	Description      *string   `db:"description" json:"description,omitempty"`
	TypeCode         *string   `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay      *string   `db:"type_display" json:"type_display,omitempty"`
	EffectCode       *string   `db:"effect_code" json:"effect_code,omitempty"`
	EffectDisplay    *string   `db:"effect_display" json:"effect_display,omitempty"`
	IncidenceCode    *string   `db:"incidence_code" json:"incidence_code,omitempty"`
	IncidenceDisplay *string   `db:"incidence_display" json:"incidence_display,omitempty"`
	ManagementCode   *string   `db:"management_code" json:"management_code,omitempty"`
	ManagementDisplay *string  `db:"management_display" json:"management_display,omitempty"`
	VersionID        int       `db:"version_id" json:"version_id"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

func (m *MedicinalProductInteraction) GetVersionID() int  { return m.VersionID }
func (m *MedicinalProductInteraction) SetVersionID(v int) { m.VersionID = v }

func (m *MedicinalProductInteraction) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "MedicinalProductInteraction",
		"id":           m.FHIRID,
		"meta":         fhir.Meta{LastUpdated: m.UpdatedAt},
	}
	if m.SubjectReference != nil { result["subject"] = []fhir.Reference{{Reference: *m.SubjectReference}} }
	if m.Description != nil { result["description"] = *m.Description }
	if m.TypeCode != nil { result["type"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.TypeCode, Display: strVal(m.TypeDisplay)}}} }
	if m.EffectCode != nil { result["effect"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.EffectCode, Display: strVal(m.EffectDisplay)}}} }
	if m.IncidenceCode != nil { result["incidence"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.IncidenceCode, Display: strVal(m.IncidenceDisplay)}}} }
	if m.ManagementCode != nil { result["management"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.ManagementCode, Display: strVal(m.ManagementDisplay)}}} }
	return result
}

func strVal(s *string) string { if s == nil { return "" }; return *s }
