package medication

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Medication maps to the medication table (FHIR Medication resource / drug catalog).
type Medication struct {
	ID                      uuid.UUID  `db:"id" json:"id"`
	FHIRID                  string     `db:"fhir_id" json:"fhir_id"`
	CodeSystem              *string    `db:"code_system" json:"code_system,omitempty"`
	CodeValue               string     `db:"code_value" json:"code_value"`
	CodeDisplay             string     `db:"code_display" json:"code_display"`
	Status                  string     `db:"status" json:"status"`
	FormCode                *string    `db:"form_code" json:"form_code,omitempty"`
	FormDisplay             *string    `db:"form_display" json:"form_display,omitempty"`
	AmountNumerator         *float64   `db:"amount_numerator" json:"amount_numerator,omitempty"`
	AmountNumeratorUnit     *string    `db:"amount_numerator_unit" json:"amount_numerator_unit,omitempty"`
	AmountDenominator       *float64   `db:"amount_denominator" json:"amount_denominator,omitempty"`
	AmountDenominatorUnit   *string    `db:"amount_denominator_unit" json:"amount_denominator_unit,omitempty"`
	Schedule                *string    `db:"schedule" json:"schedule,omitempty"`
	IsBrand                 *bool      `db:"is_brand" json:"is_brand,omitempty"`
	IsOverTheCounter        *bool      `db:"is_over_the_counter" json:"is_over_the_counter,omitempty"`
	ManufacturerID          *uuid.UUID `db:"manufacturer_id" json:"manufacturer_id,omitempty"`
	ManufacturerName        *string    `db:"manufacturer_name" json:"manufacturer_name,omitempty"`
	LotNumber               *string    `db:"lot_number" json:"lot_number,omitempty"`
	ExpirationDate          *time.Time `db:"expiration_date" json:"expiration_date,omitempty"`
	NDCCode                 *string    `db:"ndc_code" json:"ndc_code,omitempty"`
	GTINCode                *string    `db:"gtin_code" json:"gtin_code,omitempty"`
	DPCOScheduled           *bool      `db:"dpco_scheduled" json:"dpco_scheduled,omitempty"`
	CDSCOApproval           *string    `db:"cdsco_approval" json:"cdsco_approval,omitempty"`
	IsNarcotic              *bool      `db:"is_narcotic" json:"is_narcotic,omitempty"`
	IsAntibiotic            *bool      `db:"is_antibiotic" json:"is_antibiotic,omitempty"`
	IsHighAlert             *bool      `db:"is_high_alert" json:"is_high_alert,omitempty"`
	RequiresReconstitution  *bool      `db:"requires_reconstitution" json:"requires_reconstitution,omitempty"`
	Description             *string    `db:"description" json:"description,omitempty"`
	Note                    *string    `db:"note" json:"note,omitempty"`
	CreatedAt               time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time  `db:"updated_at" json:"updated_at"`
}

// MedicationIngredient maps to the medication_ingredient table.
type MedicationIngredient struct {
	ID                      uuid.UUID `db:"id" json:"id"`
	MedicationID            uuid.UUID `db:"medication_id" json:"medication_id"`
	ItemCode                *string   `db:"item_code" json:"item_code,omitempty"`
	ItemDisplay             string    `db:"item_display" json:"item_display"`
	ItemSystem              *string   `db:"item_system" json:"item_system,omitempty"`
	StrengthNumerator       *float64  `db:"strength_numerator" json:"strength_numerator,omitempty"`
	StrengthNumeratorUnit   *string   `db:"strength_numerator_unit" json:"strength_numerator_unit,omitempty"`
	StrengthDenominator     *float64  `db:"strength_denominator" json:"strength_denominator,omitempty"`
	StrengthDenominatorUnit *string   `db:"strength_denominator_unit" json:"strength_denominator_unit,omitempty"`
	IsActive                *bool     `db:"is_active" json:"is_active,omitempty"`
}

// MedicationRequest maps to the medication_request table (FHIR MedicationRequest resource).
type MedicationRequest struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	FHIRID                string     `db:"fhir_id" json:"fhir_id"`
	Status                string     `db:"status" json:"status"`
	StatusReasonCode      *string    `db:"status_reason_code" json:"status_reason_code,omitempty"`
	StatusReasonDisplay   *string    `db:"status_reason_display" json:"status_reason_display,omitempty"`
	Intent                string     `db:"intent" json:"intent"`
	CategoryCode          *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay       *string    `db:"category_display" json:"category_display,omitempty"`
	Priority              *string    `db:"priority" json:"priority,omitempty"`
	MedicationID          uuid.UUID  `db:"medication_id" json:"medication_id"`
	PatientID             uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID           *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	RequesterID           uuid.UUID  `db:"requester_id" json:"requester_id"`
	PerformerID           *uuid.UUID `db:"performer_id" json:"performer_id,omitempty"`
	RecorderID            *uuid.UUID `db:"recorder_id" json:"recorder_id,omitempty"`
	ReasonCode            *string    `db:"reason_code" json:"reason_code,omitempty"`
	ReasonDisplay         *string    `db:"reason_display" json:"reason_display,omitempty"`
	ReasonConditionID     *uuid.UUID `db:"reason_condition_id" json:"reason_condition_id,omitempty"`
	DosageText            *string    `db:"dosage_text" json:"dosage_text,omitempty"`
	DosageTimingCode      *string    `db:"dosage_timing_code" json:"dosage_timing_code,omitempty"`
	DosageTimingDisplay   *string    `db:"dosage_timing_display" json:"dosage_timing_display,omitempty"`
	DosageRouteCode       *string    `db:"dosage_route_code" json:"dosage_route_code,omitempty"`
	DosageRouteDisplay    *string    `db:"dosage_route_display" json:"dosage_route_display,omitempty"`
	DosageSiteCode        *string    `db:"dosage_site_code" json:"dosage_site_code,omitempty"`
	DosageSiteDisplay     *string    `db:"dosage_site_display" json:"dosage_site_display,omitempty"`
	DosageMethodCode      *string    `db:"dosage_method_code" json:"dosage_method_code,omitempty"`
	DosageMethodDisplay   *string    `db:"dosage_method_display" json:"dosage_method_display,omitempty"`
	DoseQuantity          *float64   `db:"dose_quantity" json:"dose_quantity,omitempty"`
	DoseUnit              *string    `db:"dose_unit" json:"dose_unit,omitempty"`
	MaxDosePerPeriod      *float64   `db:"max_dose_per_period" json:"max_dose_per_period,omitempty"`
	MaxDosePerPeriodUnit  *string    `db:"max_dose_per_period_unit" json:"max_dose_per_period_unit,omitempty"`
	RateQuantity          *float64   `db:"rate_quantity" json:"rate_quantity,omitempty"`
	RateUnit              *string    `db:"rate_unit" json:"rate_unit,omitempty"`
	AsNeeded              *bool      `db:"as_needed" json:"as_needed,omitempty"`
	AsNeededCode          *string    `db:"as_needed_code" json:"as_needed_code,omitempty"`
	AsNeededDisplay       *string    `db:"as_needed_display" json:"as_needed_display,omitempty"`
	QuantityValue         *float64   `db:"quantity_value" json:"quantity_value,omitempty"`
	QuantityUnit          *string    `db:"quantity_unit" json:"quantity_unit,omitempty"`
	DaysSupply            *int       `db:"days_supply" json:"days_supply,omitempty"`
	RefillsAllowed        *int       `db:"refills_allowed" json:"refills_allowed,omitempty"`
	ValidityStart         *time.Time `db:"validity_start" json:"validity_start,omitempty"`
	ValidityEnd           *time.Time `db:"validity_end" json:"validity_end,omitempty"`
	SubstitutionAllowed   *bool      `db:"substitution_allowed" json:"substitution_allowed,omitempty"`
	SubstitutionReason    *string    `db:"substitution_reason" json:"substitution_reason,omitempty"`
	AuthoredOn            *time.Time `db:"authored_on" json:"authored_on,omitempty"`
	PriorAuthNumber       *string    `db:"prior_auth_number" json:"prior_auth_number,omitempty"`
	ERxReference          *string    `db:"erx_reference" json:"erx_reference,omitempty"`
	ABDMPrescriptionID    *string    `db:"abdm_prescription_id" json:"abdm_prescription_id,omitempty"`
	Note                  *string    `db:"note" json:"note,omitempty"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

func (mr *MedicationRequest) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           mr.FHIRID,
		"status":       mr.Status,
		"intent":       mr.Intent,
		"medicationReference": fhir.Reference{
			Reference: fhir.FormatReference("Medication", mr.MedicationID.String()),
		},
		"subject": fhir.Reference{Reference: fhir.FormatReference("Patient", mr.PatientID.String())},
		"requester": fhir.Reference{Reference: fhir.FormatReference("Practitioner", mr.RequesterID.String())},
		"meta":     fhir.Meta{LastUpdated: mr.UpdatedAt},
	}
	if mr.CategoryCode != nil {
		result["category"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{
				System: "http://terminology.hl7.org/CodeSystem/medicationrequest-category",
				Code:   *mr.CategoryCode,
				Display: strVal(mr.CategoryDisplay),
			}},
		}}
	}
	if mr.Priority != nil {
		result["priority"] = *mr.Priority
	}
	if mr.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", mr.EncounterID.String())}
	}
	if mr.AuthoredOn != nil {
		result["authoredOn"] = mr.AuthoredOn.Format(time.RFC3339)
	}
	if mr.ReasonCode != nil {
		result["reasonCode"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *mr.ReasonCode, Display: strVal(mr.ReasonDisplay)}},
		}}
	}
	// dosageInstruction
	dosage := map[string]interface{}{}
	hasDosage := false
	if mr.DosageText != nil {
		dosage["text"] = *mr.DosageText
		hasDosage = true
	}
	if mr.DosageTimingCode != nil {
		dosage["timing"] = map[string]interface{}{
			"code": fhir.CodeableConcept{
				Coding: []fhir.Coding{{Code: *mr.DosageTimingCode, Display: strVal(mr.DosageTimingDisplay)}},
			},
		}
		hasDosage = true
	}
	if mr.DosageRouteCode != nil {
		dosage["route"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *mr.DosageRouteCode, Display: strVal(mr.DosageRouteDisplay)}},
		}
		hasDosage = true
	}
	if mr.DoseQuantity != nil {
		dosage["doseAndRate"] = []map[string]interface{}{{
			"doseQuantity": map[string]interface{}{
				"value": *mr.DoseQuantity,
				"unit":  strVal(mr.DoseUnit),
			},
		}}
		hasDosage = true
	}
	if mr.AsNeeded != nil {
		dosage["asNeededBoolean"] = *mr.AsNeeded
		hasDosage = true
	}
	if hasDosage {
		result["dosageInstruction"] = []interface{}{dosage}
	}
	// dispenseRequest
	dispReq := map[string]interface{}{}
	hasDispReq := false
	if mr.QuantityValue != nil {
		dispReq["quantity"] = map[string]interface{}{
			"value": *mr.QuantityValue,
			"unit":  strVal(mr.QuantityUnit),
		}
		hasDispReq = true
	}
	if mr.DaysSupply != nil {
		dispReq["expectedSupplyDuration"] = map[string]interface{}{
			"value":  *mr.DaysSupply,
			"unit":   "days",
			"system": "http://unitsofmeasure.org",
			"code":   "d",
		}
		hasDispReq = true
	}
	if mr.RefillsAllowed != nil {
		dispReq["numberOfRepeatsAllowed"] = *mr.RefillsAllowed
		hasDispReq = true
	}
	if mr.ValidityStart != nil || mr.ValidityEnd != nil {
		dispReq["validityPeriod"] = fhir.Period{Start: mr.ValidityStart, End: mr.ValidityEnd}
		hasDispReq = true
	}
	if hasDispReq {
		result["dispenseRequest"] = dispReq
	}
	if mr.SubstitutionAllowed != nil {
		result["substitution"] = map[string]interface{}{
			"allowedBoolean": *mr.SubstitutionAllowed,
		}
	}
	if mr.Note != nil {
		result["note"] = []map[string]string{{"text": *mr.Note}}
	}
	return result
}

// MedicationAdministration maps to the medication_administration table (FHIR MedicationAdministration).
type MedicationAdministration struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	FHIRID                string     `db:"fhir_id" json:"fhir_id"`
	Status                string     `db:"status" json:"status"`
	StatusReasonCode      *string    `db:"status_reason_code" json:"status_reason_code,omitempty"`
	StatusReasonDisplay   *string    `db:"status_reason_display" json:"status_reason_display,omitempty"`
	CategoryCode          *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay       *string    `db:"category_display" json:"category_display,omitempty"`
	MedicationID          uuid.UUID  `db:"medication_id" json:"medication_id"`
	PatientID             uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID           *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	MedicationRequestID   *uuid.UUID `db:"medication_request_id" json:"medication_request_id,omitempty"`
	PerformerID           *uuid.UUID `db:"performer_id" json:"performer_id,omitempty"`
	PerformerRoleCode     *string    `db:"performer_role_code" json:"performer_role_code,omitempty"`
	PerformerRoleDisplay  *string    `db:"performer_role_display" json:"performer_role_display,omitempty"`
	EffectiveDatetime     *time.Time `db:"effective_datetime" json:"effective_datetime,omitempty"`
	EffectiveStart        *time.Time `db:"effective_start" json:"effective_start,omitempty"`
	EffectiveEnd          *time.Time `db:"effective_end" json:"effective_end,omitempty"`
	ReasonCode            *string    `db:"reason_code" json:"reason_code,omitempty"`
	ReasonDisplay         *string    `db:"reason_display" json:"reason_display,omitempty"`
	ReasonConditionID     *uuid.UUID `db:"reason_condition_id" json:"reason_condition_id,omitempty"`
	DosageText            *string    `db:"dosage_text" json:"dosage_text,omitempty"`
	DosageRouteCode       *string    `db:"dosage_route_code" json:"dosage_route_code,omitempty"`
	DosageRouteDisplay    *string    `db:"dosage_route_display" json:"dosage_route_display,omitempty"`
	DosageSiteCode        *string    `db:"dosage_site_code" json:"dosage_site_code,omitempty"`
	DosageSiteDisplay     *string    `db:"dosage_site_display" json:"dosage_site_display,omitempty"`
	DosageMethodCode      *string    `db:"dosage_method_code" json:"dosage_method_code,omitempty"`
	DosageMethodDisplay   *string    `db:"dosage_method_display" json:"dosage_method_display,omitempty"`
	DoseQuantity          *float64   `db:"dose_quantity" json:"dose_quantity,omitempty"`
	DoseUnit              *string    `db:"dose_unit" json:"dose_unit,omitempty"`
	RateQuantity          *float64   `db:"rate_quantity" json:"rate_quantity,omitempty"`
	RateUnit              *string    `db:"rate_unit" json:"rate_unit,omitempty"`
	Note                  *string    `db:"note" json:"note,omitempty"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

func (ma *MedicationAdministration) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "MedicationAdministration",
		"id":           ma.FHIRID,
		"status":       ma.Status,
		"medicationReference": fhir.Reference{
			Reference: fhir.FormatReference("Medication", ma.MedicationID.String()),
		},
		"subject": fhir.Reference{Reference: fhir.FormatReference("Patient", ma.PatientID.String())},
		"meta":    fhir.Meta{LastUpdated: ma.UpdatedAt},
	}
	if ma.CategoryCode != nil {
		result["category"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *ma.CategoryCode, Display: strVal(ma.CategoryDisplay)}},
		}
	}
	if ma.EncounterID != nil {
		result["context"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", ma.EncounterID.String())}
	}
	if ma.MedicationRequestID != nil {
		result["request"] = fhir.Reference{Reference: fhir.FormatReference("MedicationRequest", ma.MedicationRequestID.String())}
	}
	if ma.PerformerID != nil {
		performer := map[string]interface{}{
			"actor": fhir.Reference{Reference: fhir.FormatReference("Practitioner", ma.PerformerID.String())},
		}
		if ma.PerformerRoleCode != nil {
			performer["function"] = fhir.CodeableConcept{
				Coding: []fhir.Coding{{Code: *ma.PerformerRoleCode, Display: strVal(ma.PerformerRoleDisplay)}},
			}
		}
		result["performer"] = []interface{}{performer}
	}
	if ma.EffectiveDatetime != nil {
		result["effectiveDateTime"] = ma.EffectiveDatetime.Format(time.RFC3339)
	} else if ma.EffectiveStart != nil {
		result["effectivePeriod"] = fhir.Period{Start: ma.EffectiveStart, End: ma.EffectiveEnd}
	}
	if ma.ReasonCode != nil {
		result["reasonCode"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *ma.ReasonCode, Display: strVal(ma.ReasonDisplay)}},
		}}
	}
	// dosage
	dosage := map[string]interface{}{}
	hasDosage := false
	if ma.DosageText != nil {
		dosage["text"] = *ma.DosageText
		hasDosage = true
	}
	if ma.DosageRouteCode != nil {
		dosage["route"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *ma.DosageRouteCode, Display: strVal(ma.DosageRouteDisplay)}},
		}
		hasDosage = true
	}
	if ma.DosageSiteCode != nil {
		dosage["site"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *ma.DosageSiteCode, Display: strVal(ma.DosageSiteDisplay)}},
		}
		hasDosage = true
	}
	if ma.DoseQuantity != nil {
		dosage["dose"] = map[string]interface{}{
			"value": *ma.DoseQuantity,
			"unit":  strVal(ma.DoseUnit),
		}
		hasDosage = true
	}
	if ma.RateQuantity != nil {
		dosage["rateQuantity"] = map[string]interface{}{
			"value": *ma.RateQuantity,
			"unit":  strVal(ma.RateUnit),
		}
		hasDosage = true
	}
	if hasDosage {
		result["dosage"] = dosage
	}
	if ma.Note != nil {
		result["note"] = []map[string]string{{"text": *ma.Note}}
	}
	return result
}

// MedicationDispense maps to the medication_dispense table (FHIR MedicationDispense).
type MedicationDispense struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	FHIRID                string     `db:"fhir_id" json:"fhir_id"`
	Status                string     `db:"status" json:"status"`
	StatusReasonCode      *string    `db:"status_reason_code" json:"status_reason_code,omitempty"`
	StatusReasonDisplay   *string    `db:"status_reason_display" json:"status_reason_display,omitempty"`
	CategoryCode          *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay       *string    `db:"category_display" json:"category_display,omitempty"`
	MedicationID          uuid.UUID  `db:"medication_id" json:"medication_id"`
	PatientID             uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID           *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	MedicationRequestID   *uuid.UUID `db:"medication_request_id" json:"medication_request_id,omitempty"`
	PerformerID           *uuid.UUID `db:"performer_id" json:"performer_id,omitempty"`
	LocationID            *uuid.UUID `db:"location_id" json:"location_id,omitempty"`
	QuantityValue         *float64   `db:"quantity_value" json:"quantity_value,omitempty"`
	QuantityUnit          *string    `db:"quantity_unit" json:"quantity_unit,omitempty"`
	DaysSupply            *int       `db:"days_supply" json:"days_supply,omitempty"`
	WhenPrepared          *time.Time `db:"when_prepared" json:"when_prepared,omitempty"`
	WhenHandedOver        *time.Time `db:"when_handed_over" json:"when_handed_over,omitempty"`
	DestinationID         *uuid.UUID `db:"destination_id" json:"destination_id,omitempty"`
	ReceiverID            *uuid.UUID `db:"receiver_id" json:"receiver_id,omitempty"`
	WasSubstituted        *bool      `db:"was_substituted" json:"was_substituted,omitempty"`
	SubstitutionTypeCode  *string    `db:"substitution_type_code" json:"substitution_type_code,omitempty"`
	SubstitutionReason    *string    `db:"substitution_reason" json:"substitution_reason,omitempty"`
	Note                  *string    `db:"note" json:"note,omitempty"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

func (md *MedicationDispense) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "MedicationDispense",
		"id":           md.FHIRID,
		"status":       md.Status,
		"medicationReference": fhir.Reference{
			Reference: fhir.FormatReference("Medication", md.MedicationID.String()),
		},
		"subject": fhir.Reference{Reference: fhir.FormatReference("Patient", md.PatientID.String())},
		"meta":    fhir.Meta{LastUpdated: md.UpdatedAt},
	}
	if md.CategoryCode != nil {
		result["category"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *md.CategoryCode, Display: strVal(md.CategoryDisplay)}},
		}
	}
	if md.EncounterID != nil {
		result["context"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", md.EncounterID.String())}
	}
	if md.MedicationRequestID != nil {
		result["authorizingPrescription"] = []fhir.Reference{{
			Reference: fhir.FormatReference("MedicationRequest", md.MedicationRequestID.String()),
		}}
	}
	if md.PerformerID != nil {
		result["performer"] = []map[string]interface{}{{
			"actor": fhir.Reference{Reference: fhir.FormatReference("Practitioner", md.PerformerID.String())},
		}}
	}
	if md.LocationID != nil {
		result["location"] = fhir.Reference{Reference: fhir.FormatReference("Location", md.LocationID.String())}
	}
	if md.QuantityValue != nil {
		result["quantity"] = map[string]interface{}{
			"value": *md.QuantityValue,
			"unit":  strVal(md.QuantityUnit),
		}
	}
	if md.DaysSupply != nil {
		result["daysSupply"] = map[string]interface{}{
			"value":  *md.DaysSupply,
			"unit":   "days",
			"system": "http://unitsofmeasure.org",
			"code":   "d",
		}
	}
	if md.WhenPrepared != nil {
		result["whenPrepared"] = md.WhenPrepared.Format(time.RFC3339)
	}
	if md.WhenHandedOver != nil {
		result["whenHandedOver"] = md.WhenHandedOver.Format(time.RFC3339)
	}
	if md.DestinationID != nil {
		result["destination"] = fhir.Reference{Reference: fhir.FormatReference("Location", md.DestinationID.String())}
	}
	if md.WasSubstituted != nil {
		sub := map[string]interface{}{
			"wasSubstituted": *md.WasSubstituted,
		}
		if md.SubstitutionTypeCode != nil {
			sub["type"] = fhir.CodeableConcept{
				Coding: []fhir.Coding{{Code: *md.SubstitutionTypeCode}},
			}
		}
		if md.SubstitutionReason != nil {
			sub["reason"] = []fhir.CodeableConcept{{
				Text: *md.SubstitutionReason,
			}}
		}
		result["substitution"] = sub
	}
	if md.Note != nil {
		result["note"] = []map[string]string{{"text": *md.Note}}
	}
	return result
}

// MedicationStatement maps to the medication_statement table (FHIR MedicationStatement).
type MedicationStatement struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	FHIRID                string     `db:"fhir_id" json:"fhir_id"`
	Status                string     `db:"status" json:"status"`
	StatusReasonCode      *string    `db:"status_reason_code" json:"status_reason_code,omitempty"`
	StatusReasonDisplay   *string    `db:"status_reason_display" json:"status_reason_display,omitempty"`
	CategoryCode          *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay       *string    `db:"category_display" json:"category_display,omitempty"`
	MedicationCode        *string    `db:"medication_code" json:"medication_code,omitempty"`
	MedicationDisplay     *string    `db:"medication_display" json:"medication_display,omitempty"`
	MedicationID          *uuid.UUID `db:"medication_id" json:"medication_id,omitempty"`
	PatientID             uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID           *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	InformationSourceID   *uuid.UUID `db:"information_source_id" json:"information_source_id,omitempty"`
	EffectiveDatetime     *time.Time `db:"effective_datetime" json:"effective_datetime,omitempty"`
	EffectiveStart        *time.Time `db:"effective_start" json:"effective_start,omitempty"`
	EffectiveEnd          *time.Time `db:"effective_end" json:"effective_end,omitempty"`
	DateAsserted          *time.Time `db:"date_asserted" json:"date_asserted,omitempty"`
	ReasonCode            *string    `db:"reason_code" json:"reason_code,omitempty"`
	ReasonDisplay         *string    `db:"reason_display" json:"reason_display,omitempty"`
	DosageText            *string    `db:"dosage_text" json:"dosage_text,omitempty"`
	DosageRouteCode       *string    `db:"dosage_route_code" json:"dosage_route_code,omitempty"`
	DosageRouteDisplay    *string    `db:"dosage_route_display" json:"dosage_route_display,omitempty"`
	DoseQuantity          *float64   `db:"dose_quantity" json:"dose_quantity,omitempty"`
	DoseUnit              *string    `db:"dose_unit" json:"dose_unit,omitempty"`
	DosageTimingCode      *string    `db:"dosage_timing_code" json:"dosage_timing_code,omitempty"`
	DosageTimingDisplay   *string    `db:"dosage_timing_display" json:"dosage_timing_display,omitempty"`
	Note                  *string    `db:"note" json:"note,omitempty"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
