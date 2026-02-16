package biologicallyderivedproduct

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// BiologicallyDerivedProduct maps to the biologically_derived_product table (FHIR BiologicallyDerivedProduct resource).
type BiologicallyDerivedProduct struct {
	ID                      uuid.UUID  `db:"id" json:"id"`
	FHIRID                  string     `db:"fhir_id" json:"fhir_id"`
	ProductCategory         *string    `db:"product_category" json:"product_category,omitempty"`
	ProductCodeCode         *string    `db:"product_code_code" json:"product_code_code,omitempty"`
	ProductCodeDisplay      *string    `db:"product_code_display" json:"product_code_display,omitempty"`
	Status                  *string    `db:"status" json:"status,omitempty"`
	RequestID               *uuid.UUID `db:"request_id" json:"request_id,omitempty"`
	Quantity                *int       `db:"quantity" json:"quantity,omitempty"`
	ParentID                *uuid.UUID `db:"parent_id" json:"parent_id,omitempty"`
	CollectionSourceType    *string    `db:"collection_source_type" json:"collection_source_type,omitempty"`
	CollectionSourceRef     *string    `db:"collection_source_reference" json:"collection_source_reference,omitempty"`
	CollectionCollectedDate *time.Time `db:"collection_collected_date" json:"collection_collected_date,omitempty"`
	ProcessingDescription   *string    `db:"processing_description" json:"processing_description,omitempty"`
	StorageTemperatureCode  *string    `db:"storage_temperature_code" json:"storage_temperature_code,omitempty"`
	StorageDuration         *string    `db:"storage_duration" json:"storage_duration,omitempty"`
	VersionID               int        `db:"version_id" json:"version_id"`
	CreatedAt               time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time  `db:"updated_at" json:"updated_at"`
}

func (b *BiologicallyDerivedProduct) GetVersionID() int  { return b.VersionID }
func (b *BiologicallyDerivedProduct) SetVersionID(v int) { b.VersionID = v }

func (b *BiologicallyDerivedProduct) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "BiologicallyDerivedProduct",
		"id":           b.FHIRID,
		"meta":         fhir.Meta{LastUpdated: b.UpdatedAt},
	}
	if b.ProductCategory != nil {
		result["productCategory"] = *b.ProductCategory
	}
	if b.ProductCodeCode != nil {
		result["productCode"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *b.ProductCodeCode, Display: strVal(b.ProductCodeDisplay)}},
		}
	}
	if b.Status != nil {
		result["status"] = *b.Status
	}
	if b.RequestID != nil {
		result["request"] = []fhir.Reference{{Reference: fhir.FormatReference("ServiceRequest", b.RequestID.String())}}
	}
	if b.Quantity != nil {
		result["quantity"] = *b.Quantity
	}
	if b.ParentID != nil {
		result["parent"] = []fhir.Reference{{Reference: fhir.FormatReference("BiologicallyDerivedProduct", b.ParentID.String())}}
	}
	if b.CollectionSourceType != nil || b.CollectionCollectedDate != nil {
		collection := map[string]interface{}{}
		if b.CollectionSourceType != nil {
			collection["source"] = fhir.Reference{Reference: fhir.FormatReference(*b.CollectionSourceType, strVal(b.CollectionSourceRef))}
		}
		if b.CollectionCollectedDate != nil {
			collection["collectedDateTime"] = b.CollectionCollectedDate.Format("2006-01-02T15:04:05Z")
		}
		result["collection"] = collection
	}
	if b.ProcessingDescription != nil {
		result["processing"] = []map[string]interface{}{
			{"description": *b.ProcessingDescription},
		}
	}
	if b.StorageTemperatureCode != nil || b.StorageDuration != nil {
		storage := map[string]interface{}{}
		if b.StorageTemperatureCode != nil {
			storage["temperature"] = *b.StorageTemperatureCode
		}
		if b.StorageDuration != nil {
			storage["duration"] = *b.StorageDuration
		}
		result["storage"] = []map[string]interface{}{storage}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
