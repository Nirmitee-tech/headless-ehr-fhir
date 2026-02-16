package substancereferenceinformation

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type SubstanceReferenceInformation struct {
	ID                            uuid.UUID `db:"id" json:"id"`
	FHIRID                        string    `db:"fhir_id" json:"fhir_id"`
	Comment                       *string   `db:"comment" json:"comment,omitempty"`
	GeneElementTypeCode           *string   `db:"gene_element_type_code" json:"gene_element_type_code,omitempty"`
	GeneElementTypeDisplay        *string   `db:"gene_element_type_display" json:"gene_element_type_display,omitempty"`
	GeneElementSourceReference    *string   `db:"gene_element_source_reference" json:"gene_element_source_reference,omitempty"`
	ClassificationCode            *string   `db:"classification_code" json:"classification_code,omitempty"`
	ClassificationDisplay         *string   `db:"classification_display" json:"classification_display,omitempty"`
	ClassificationDomainCode      *string   `db:"classification_domain_code" json:"classification_domain_code,omitempty"`
	ClassificationDomainDisplay   *string   `db:"classification_domain_display" json:"classification_domain_display,omitempty"`
	TargetTypeCode                *string   `db:"target_type_code" json:"target_type_code,omitempty"`
	TargetTypeDisplay             *string   `db:"target_type_display" json:"target_type_display,omitempty"`
	VersionID                     int       `db:"version_id" json:"version_id"`
	CreatedAt                     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                     time.Time `db:"updated_at" json:"updated_at"`
}

func (m *SubstanceReferenceInformation) GetVersionID() int  { return m.VersionID }
func (m *SubstanceReferenceInformation) SetVersionID(v int) { m.VersionID = v }

func (m *SubstanceReferenceInformation) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "SubstanceReferenceInformation",
		"id":           m.FHIRID,
		"meta":         fhir.Meta{LastUpdated: m.UpdatedAt},
	}
	if m.Comment != nil { result["comment"] = *m.Comment }
	if m.GeneElementTypeCode != nil || m.GeneElementSourceReference != nil {
		ge := map[string]interface{}{}
		if m.GeneElementTypeCode != nil { ge["type"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.GeneElementTypeCode, Display: strVal(m.GeneElementTypeDisplay)}}} }
		if m.GeneElementSourceReference != nil { ge["source"] = []fhir.Reference{{Reference: *m.GeneElementSourceReference}} }
		result["geneElement"] = []interface{}{ge}
	}
	if m.ClassificationCode != nil {
		cls := map[string]interface{}{}
		cls["classification"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.ClassificationCode, Display: strVal(m.ClassificationDisplay)}}}
		if m.ClassificationDomainCode != nil { cls["domain"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.ClassificationDomainCode, Display: strVal(m.ClassificationDomainDisplay)}}} }
		result["classification"] = []interface{}{cls}
	}
	if m.TargetTypeCode != nil {
		tgt := map[string]interface{}{}
		tgt["type"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: *m.TargetTypeCode, Display: strVal(m.TargetTypeDisplay)}}}
		result["target"] = []interface{}{tgt}
	}
	return result
}

func strVal(s *string) string { if s == nil { return "" }; return *s }
