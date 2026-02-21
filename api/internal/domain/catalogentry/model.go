package catalogentry

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// CatalogEntry maps to the catalog_entry table (FHIR CatalogEntry resource).
type CatalogEntry struct {
	ID                      uuid.UUID  `db:"id" json:"id"`
	FHIRID                  string     `db:"fhir_id" json:"fhir_id"`
	Type                    *string    `db:"type" json:"type,omitempty"`
	Orderable               bool       `db:"orderable" json:"orderable"`
	ReferencedItemType      string     `db:"referenced_item_type" json:"referenced_item_type"`
	ReferencedItemReference string     `db:"referenced_item_reference" json:"referenced_item_reference"`
	Status                  string     `db:"status" json:"status"`
	EffectivePeriodStart    *time.Time `db:"effective_period_start" json:"effective_period_start,omitempty"`
	EffectivePeriodEnd      *time.Time `db:"effective_period_end" json:"effective_period_end,omitempty"`
	AdditionalIdentifier    *string    `db:"additional_identifier" json:"additional_identifier,omitempty"`
	ClassificationCode      *string    `db:"classification_code" json:"classification_code,omitempty"`
	ClassificationDisplay   *string    `db:"classification_display" json:"classification_display,omitempty"`
	ValidityPeriodStart     *time.Time `db:"validity_period_start" json:"validity_period_start,omitempty"`
	ValidityPeriodEnd       *time.Time `db:"validity_period_end" json:"validity_period_end,omitempty"`
	LastUpdatedTS           *time.Time `db:"last_updated_ts" json:"last_updated_ts,omitempty"`
	VersionID               int        `db:"version_id" json:"version_id"`
	CreatedAt               time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time  `db:"updated_at" json:"updated_at"`
}

func (ce *CatalogEntry) GetVersionID() int  { return ce.VersionID }
func (ce *CatalogEntry) SetVersionID(v int) { ce.VersionID = v }

func (ce *CatalogEntry) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType":   "CatalogEntry",
		"id":             ce.FHIRID,
		"orderable":      ce.Orderable,
		"referencedItem": fhir.Reference{Reference: fhir.FormatReference(ce.ReferencedItemType, ce.ReferencedItemReference)},
		"status":         ce.Status,
		"meta":           fhir.Meta{
			VersionID:   fmt.Sprintf("%d", ce.VersionID),
			LastUpdated: ce.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/CatalogEntry"},
		},
	}
	if ce.Type != nil {
		result["type"] = *ce.Type
	}
	if ce.EffectivePeriodStart != nil || ce.EffectivePeriodEnd != nil {
		result["effectivePeriod"] = fhir.Period{Start: ce.EffectivePeriodStart, End: ce.EffectivePeriodEnd}
	}
	if ce.AdditionalIdentifier != nil {
		result["additionalIdentifier"] = []fhir.Identifier{{Value: *ce.AdditionalIdentifier}}
	}
	if ce.ClassificationCode != nil {
		result["classification"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: *ce.ClassificationCode, Display: strVal(ce.ClassificationDisplay)}}}}
	}
	if ce.ValidityPeriodStart != nil || ce.ValidityPeriodEnd != nil {
		result["validityPeriod"] = fhir.Period{Start: ce.ValidityPeriodStart, End: ce.ValidityPeriodEnd}
	}
	if ce.LastUpdatedTS != nil {
		result["lastUpdated"] = ce.LastUpdatedTS.Format("2006-01-02T15:04:05Z")
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
