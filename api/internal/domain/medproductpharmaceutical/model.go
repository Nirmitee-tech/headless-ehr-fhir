package medproductpharmaceutical

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type MedicinalProductPharmaceutical struct {
	ID                          uuid.UUID `db:"id" json:"id"`
	FHIRID                      string    `db:"fhir_id" json:"fhir_id"`
	AdministrableDoseFormCode   string    `db:"administrable_dose_form_code" json:"administrable_dose_form_code"`
	AdministrableDoseFormDisplay *string  `db:"administrable_dose_form_display" json:"administrable_dose_form_display,omitempty"`
	UnitOfPresentationCode      *string   `db:"unit_of_presentation_code" json:"unit_of_presentation_code,omitempty"`
	UnitOfPresentationDisplay   *string   `db:"unit_of_presentation_display" json:"unit_of_presentation_display,omitempty"`
	IngredientReference         *string   `db:"ingredient_reference" json:"ingredient_reference,omitempty"`
	DeviceReference             *string   `db:"device_reference" json:"device_reference,omitempty"`
	VersionID                   int       `db:"version_id" json:"version_id"`
	CreatedAt                   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                   time.Time `db:"updated_at" json:"updated_at"`
}

func (m *MedicinalProductPharmaceutical) GetVersionID() int  { return m.VersionID }
func (m *MedicinalProductPharmaceutical) SetVersionID(v int) { m.VersionID = v }

func (m *MedicinalProductPharmaceutical) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "MedicinalProductPharmaceutical",
		"id":           m.FHIRID,
		"meta":         fhir.Meta{LastUpdated: m.UpdatedAt},
		"administrableDoseForm": fhir.CodeableConcept{Coding: []fhir.Coding{{Code: m.AdministrableDoseFormCode, Display: strVal(m.AdministrableDoseFormDisplay)}}},
	}
	if m.UnitOfPresentationCode != nil { result["unitOfPresentation"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.UnitOfPresentationCode, Display: strVal(m.UnitOfPresentationDisplay)}}} }
	if m.IngredientReference != nil { result["ingredient"] = []fhir.Reference{{Reference: *m.IngredientReference}} }
	if m.DeviceReference != nil { result["device"] = fhir.Reference{Reference: *m.DeviceReference} }
	return result
}

func strVal(s *string) string { if s == nil { return "" }; return *s }
