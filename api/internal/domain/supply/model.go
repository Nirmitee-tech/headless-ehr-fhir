package supply

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// SupplyRequest maps to the supply_request table (FHIR SupplyRequest resource).
type SupplyRequest struct {
	ID                  uuid.UUID  `db:"id" json:"id"`
	FHIRID              string     `db:"fhir_id" json:"fhir_id"`
	Status              string     `db:"status" json:"status"`
	CategoryCode        *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay     *string    `db:"category_display" json:"category_display,omitempty"`
	Priority            *string    `db:"priority" json:"priority,omitempty"`
	ItemCode            string     `db:"item_code" json:"item_code"`
	ItemDisplay         *string    `db:"item_display" json:"item_display,omitempty"`
	ItemSystem          *string    `db:"item_system" json:"item_system,omitempty"`
	QuantityValue       float64    `db:"quantity_value" json:"quantity_value"`
	QuantityUnit        *string    `db:"quantity_unit" json:"quantity_unit,omitempty"`
	OccurrenceDate      *time.Time `db:"occurrence_date" json:"occurrence_date,omitempty"`
	AuthoredOn          *time.Time `db:"authored_on" json:"authored_on,omitempty"`
	RequesterID         *uuid.UUID `db:"requester_id" json:"requester_id,omitempty"`
	SupplierOrgID       *uuid.UUID `db:"supplier_org_id" json:"supplier_org_id,omitempty"`
	DeliverToLocationID *uuid.UUID `db:"deliver_to_location_id" json:"deliver_to_location_id,omitempty"`
	ReasonCode          *string    `db:"reason_code" json:"reason_code,omitempty"`
	ReasonDisplay       *string    `db:"reason_display" json:"reason_display,omitempty"`
	VersionID           int        `db:"version_id" json:"version_id"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (s *SupplyRequest) GetVersionID() int { return s.VersionID }

// SetVersionID sets the current version.
func (s *SupplyRequest) SetVersionID(v int) { s.VersionID = v }

func (s *SupplyRequest) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "SupplyRequest",
		"id":           s.FHIRID,
		"status":       s.Status,
		"itemCodeableConcept": fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System:  strVal(s.ItemSystem),
				Code:    s.ItemCode,
				Display: strVal(s.ItemDisplay),
			}},
		},
		"quantity": map[string]interface{}{
			"value": s.QuantityValue,
			"unit":  strVal(s.QuantityUnit),
		},
		"meta": fhir.Meta{LastUpdated: s.UpdatedAt},
	}
	if s.CategoryCode != nil {
		result["category"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				Code:    *s.CategoryCode,
				Display: strVal(s.CategoryDisplay),
			}},
		}
	}
	if s.Priority != nil {
		result["priority"] = *s.Priority
	}
	if s.OccurrenceDate != nil {
		result["occurrenceDateTime"] = s.OccurrenceDate.Format("2006-01-02")
	}
	if s.AuthoredOn != nil {
		result["authoredOn"] = s.AuthoredOn.Format("2006-01-02")
	}
	if s.RequesterID != nil {
		result["requester"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", s.RequesterID.String())}
	}
	if s.SupplierOrgID != nil {
		result["supplier"] = []fhir.Reference{{Reference: fhir.FormatReference("Organization", s.SupplierOrgID.String())}}
	}
	if s.DeliverToLocationID != nil {
		result["deliverTo"] = fhir.Reference{Reference: fhir.FormatReference("Location", s.DeliverToLocationID.String())}
	}
	if s.ReasonCode != nil {
		result["reasonCode"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *s.ReasonCode, Display: strVal(s.ReasonDisplay)}},
		}}
	}
	return result
}

// SupplyDelivery maps to the supply_delivery table (FHIR SupplyDelivery resource).
type SupplyDelivery struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	FHIRID                string     `db:"fhir_id" json:"fhir_id"`
	Status                string     `db:"status" json:"status"`
	BasedOnID             *uuid.UUID `db:"based_on_id" json:"based_on_id,omitempty"`
	PatientID             *uuid.UUID `db:"patient_id" json:"patient_id,omitempty"`
	TypeCode              *string    `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay           *string    `db:"type_display" json:"type_display,omitempty"`
	SuppliedItemCode      *string    `db:"supplied_item_code" json:"supplied_item_code,omitempty"`
	SuppliedItemDisplay   *string    `db:"supplied_item_display" json:"supplied_item_display,omitempty"`
	SuppliedItemQuantity  *float64   `db:"supplied_item_quantity" json:"supplied_item_quantity,omitempty"`
	SuppliedItemUnit      *string    `db:"supplied_item_unit" json:"supplied_item_unit,omitempty"`
	OccurrenceDate        *time.Time `db:"occurrence_date" json:"occurrence_date,omitempty"`
	SupplierID            *uuid.UUID `db:"supplier_id" json:"supplier_id,omitempty"`
	DestinationLocationID *uuid.UUID `db:"destination_location_id" json:"destination_location_id,omitempty"`
	ReceiverID            *uuid.UUID `db:"receiver_id" json:"receiver_id,omitempty"`
	VersionID             int        `db:"version_id" json:"version_id"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (s *SupplyDelivery) GetVersionID() int { return s.VersionID }

// SetVersionID sets the current version.
func (s *SupplyDelivery) SetVersionID(v int) { s.VersionID = v }

func (s *SupplyDelivery) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "SupplyDelivery",
		"id":           s.FHIRID,
		"status":       s.Status,
		"meta":         fhir.Meta{LastUpdated: s.UpdatedAt},
	}
	if s.BasedOnID != nil {
		result["basedOn"] = []fhir.Reference{{Reference: fhir.FormatReference("SupplyRequest", s.BasedOnID.String())}}
	}
	if s.PatientID != nil {
		result["patient"] = fhir.Reference{Reference: fhir.FormatReference("Patient", s.PatientID.String())}
	}
	if s.TypeCode != nil {
		result["type"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				Code:    *s.TypeCode,
				Display: strVal(s.TypeDisplay),
			}},
		}
	}
	if s.SuppliedItemCode != nil || s.SuppliedItemQuantity != nil {
		item := map[string]interface{}{}
		if s.SuppliedItemQuantity != nil {
			qty := map[string]interface{}{"value": *s.SuppliedItemQuantity}
			if s.SuppliedItemUnit != nil {
				qty["unit"] = *s.SuppliedItemUnit
			}
			item["quantity"] = qty
		}
		if s.SuppliedItemCode != nil {
			item["itemCodeableConcept"] = fhir.CodeableConcept{
				Coding: []fhir.Coding{{
					Code:    *s.SuppliedItemCode,
					Display: strVal(s.SuppliedItemDisplay),
				}},
			}
		}
		result["suppliedItem"] = item
	}
	if s.OccurrenceDate != nil {
		result["occurrenceDateTime"] = s.OccurrenceDate.Format("2006-01-02")
	}
	if s.SupplierID != nil {
		result["supplier"] = fhir.Reference{Reference: fhir.FormatReference("Organization", s.SupplierID.String())}
	}
	if s.DestinationLocationID != nil {
		result["destination"] = fhir.Reference{Reference: fhir.FormatReference("Location", s.DestinationLocationID.String())}
	}
	if s.ReceiverID != nil {
		result["receiver"] = []fhir.Reference{{Reference: fhir.FormatReference("Practitioner", s.ReceiverID.String())}}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
