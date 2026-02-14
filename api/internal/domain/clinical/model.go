package clinical

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Condition maps to the condition table (FHIR Condition resource).
type Condition struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	FHIRID             string     `db:"fhir_id" json:"fhir_id"`
	PatientID          uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID        *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	RecorderID         *uuid.UUID `db:"recorder_id" json:"recorder_id,omitempty"`
	AsserterID         *uuid.UUID `db:"asserter_id" json:"asserter_id,omitempty"`
	ClinicalStatus     string     `db:"clinical_status" json:"clinical_status"`
	VerificationStatus *string    `db:"verification_status" json:"verification_status,omitempty"`
	CategoryCode       *string    `db:"category_code" json:"category_code,omitempty"`
	SeverityCode       *string    `db:"severity_code" json:"severity_code,omitempty"`
	SeverityDisplay    *string    `db:"severity_display" json:"severity_display,omitempty"`
	CodeSystem         *string    `db:"code_system" json:"code_system,omitempty"`
	CodeValue          string     `db:"code_value" json:"code_value"`
	CodeDisplay        string     `db:"code_display" json:"code_display"`
	AltCodeSystem      *string    `db:"alt_code_system" json:"alt_code_system,omitempty"`
	AltCodeValue       *string    `db:"alt_code_value" json:"alt_code_value,omitempty"`
	AltCodeDisplay     *string    `db:"alt_code_display" json:"alt_code_display,omitempty"`
	BodySiteCode       *string    `db:"body_site_code" json:"body_site_code,omitempty"`
	BodySiteDisplay    *string    `db:"body_site_display" json:"body_site_display,omitempty"`
	OnsetDatetime      *time.Time `db:"onset_datetime" json:"onset_datetime,omitempty"`
	OnsetAge           *int       `db:"onset_age" json:"onset_age,omitempty"`
	OnsetString        *string    `db:"onset_string" json:"onset_string,omitempty"`
	AbatementDatetime  *time.Time `db:"abatement_datetime" json:"abatement_datetime,omitempty"`
	AbatementAge       *int       `db:"abatement_age" json:"abatement_age,omitempty"`
	AbatementString    *string    `db:"abatement_string" json:"abatement_string,omitempty"`
	StageSummaryCode   *string    `db:"stage_summary_code" json:"stage_summary_code,omitempty"`
	StageSummaryDisp   *string    `db:"stage_summary_display" json:"stage_summary_display,omitempty"`
	StageTypeCode      *string    `db:"stage_type_code" json:"stage_type_code,omitempty"`
	EvidenceCode       *string    `db:"evidence_code" json:"evidence_code,omitempty"`
	EvidenceDisplay    *string    `db:"evidence_display" json:"evidence_display,omitempty"`
	RecordedDate       *time.Time `db:"recorded_date" json:"recorded_date,omitempty"`
	Note               *string    `db:"note" json:"note,omitempty"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

func (c *Condition) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Condition",
		"id":           c.FHIRID,
		"clinicalStatus": fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System: "http://terminology.hl7.org/CodeSystem/condition-clinical",
				Code:   c.ClinicalStatus,
			}},
		},
		"code": fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System:  strVal(c.CodeSystem),
				Code:    c.CodeValue,
				Display: c.CodeDisplay,
			}},
			Text: c.CodeDisplay,
		},
		"subject": fhir.Reference{Reference: fhir.FormatReference("Patient", c.PatientID.String())},
		"meta":    fhir.Meta{LastUpdated: c.UpdatedAt},
	}
	if c.VerificationStatus != nil {
		result["verificationStatus"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System: "http://terminology.hl7.org/CodeSystem/condition-ver-status",
				Code:   *c.VerificationStatus,
			}},
		}
	}
	if c.CategoryCode != nil {
		result["category"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *c.CategoryCode}},
		}}
	}
	if c.SeverityCode != nil {
		result["severity"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *c.SeverityCode, Display: strVal(c.SeverityDisplay)}},
		}
	}
	if c.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", c.EncounterID.String())}
	}
	if c.OnsetDatetime != nil {
		result["onsetDateTime"] = c.OnsetDatetime.Format(time.RFC3339)
	}
	if c.AbatementDatetime != nil {
		result["abatementDateTime"] = c.AbatementDatetime.Format(time.RFC3339)
	}
	if c.BodySiteCode != nil {
		result["bodySite"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *c.BodySiteCode, Display: strVal(c.BodySiteDisplay)}},
		}}
	}
	if c.RecordedDate != nil {
		result["recordedDate"] = c.RecordedDate.Format("2006-01-02")
	}
	if c.Note != nil {
		result["note"] = []map[string]string{{"text": *c.Note}}
	}
	return result
}

// Observation maps to the observation table (FHIR Observation resource).
type Observation struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	FHIRID                string     `db:"fhir_id" json:"fhir_id"`
	Status                string     `db:"status" json:"status"`
	CategoryCode          *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay       *string    `db:"category_display" json:"category_display,omitempty"`
	CodeSystem            *string    `db:"code_system" json:"code_system,omitempty"`
	CodeValue             string     `db:"code_value" json:"code_value"`
	CodeDisplay           string     `db:"code_display" json:"code_display"`
	PatientID             uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID           *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	PerformerID           *uuid.UUID `db:"performer_id" json:"performer_id,omitempty"`
	EffectiveDatetime     *time.Time `db:"effective_datetime" json:"effective_datetime,omitempty"`
	Issued                *time.Time `db:"issued" json:"issued,omitempty"`
	ValueQuantity         *float64   `db:"value_quantity" json:"value_quantity,omitempty"`
	ValueUnit             *string    `db:"value_unit" json:"value_unit,omitempty"`
	ValueSystem           *string    `db:"value_system" json:"value_system,omitempty"`
	ValueCode             *string    `db:"value_code" json:"value_code,omitempty"`
	ValueString           *string    `db:"value_string" json:"value_string,omitempty"`
	ValueBoolean          *bool      `db:"value_boolean" json:"value_boolean,omitempty"`
	ValueInteger          *int       `db:"value_integer" json:"value_integer,omitempty"`
	ValueCodeableCode     *string    `db:"value_codeable_code" json:"value_codeable_code,omitempty"`
	ValueCodeableDisplay  *string    `db:"value_codeable_display" json:"value_codeable_display,omitempty"`
	ReferenceRangeLow     *float64   `db:"reference_range_low" json:"reference_range_low,omitempty"`
	ReferenceRangeHigh    *float64   `db:"reference_range_high" json:"reference_range_high,omitempty"`
	ReferenceRangeUnit    *string    `db:"reference_range_unit" json:"reference_range_unit,omitempty"`
	ReferenceRangeText    *string    `db:"reference_range_text" json:"reference_range_text,omitempty"`
	InterpretationCode    *string    `db:"interpretation_code" json:"interpretation_code,omitempty"`
	InterpretationDisplay *string    `db:"interpretation_display" json:"interpretation_display,omitempty"`
	BodySiteCode          *string    `db:"body_site_code" json:"body_site_code,omitempty"`
	BodySiteDisplay       *string    `db:"body_site_display" json:"body_site_display,omitempty"`
	DataAbsentReason      *string    `db:"data_absent_reason" json:"data_absent_reason,omitempty"`
	Note                  *string    `db:"note" json:"note,omitempty"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

func (o *Observation) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Observation",
		"id":           o.FHIRID,
		"status":       o.Status,
		"code": fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System:  strVal(o.CodeSystem),
				Code:    o.CodeValue,
				Display: o.CodeDisplay,
			}},
		},
		"subject": fhir.Reference{Reference: fhir.FormatReference("Patient", o.PatientID.String())},
		"meta":    fhir.Meta{LastUpdated: o.UpdatedAt},
	}
	if o.CategoryCode != nil {
		result["category"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{
				System:  "http://terminology.hl7.org/CodeSystem/observation-category",
				Code:    *o.CategoryCode,
				Display: strVal(o.CategoryDisplay),
			}},
		}}
	}
	if o.EffectiveDatetime != nil {
		result["effectiveDateTime"] = o.EffectiveDatetime.Format(time.RFC3339)
	}
	if o.ValueQuantity != nil {
		result["valueQuantity"] = map[string]interface{}{
			"value":  *o.ValueQuantity,
			"unit":   strVal(o.ValueUnit),
			"system": strVal(o.ValueSystem),
			"code":   strVal(o.ValueCode),
		}
	} else if o.ValueString != nil {
		result["valueString"] = *o.ValueString
	} else if o.ValueBoolean != nil {
		result["valueBoolean"] = *o.ValueBoolean
	} else if o.ValueInteger != nil {
		result["valueInteger"] = *o.ValueInteger
	} else if o.ValueCodeableCode != nil {
		result["valueCodeableConcept"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *o.ValueCodeableCode, Display: strVal(o.ValueCodeableDisplay)}},
		}
	}
	if o.ReferenceRangeLow != nil || o.ReferenceRangeHigh != nil {
		rr := map[string]interface{}{}
		if o.ReferenceRangeLow != nil {
			rr["low"] = map[string]interface{}{"value": *o.ReferenceRangeLow, "unit": strVal(o.ReferenceRangeUnit)}
		}
		if o.ReferenceRangeHigh != nil {
			rr["high"] = map[string]interface{}{"value": *o.ReferenceRangeHigh, "unit": strVal(o.ReferenceRangeUnit)}
		}
		result["referenceRange"] = []interface{}{rr}
	}
	if o.InterpretationCode != nil {
		result["interpretation"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *o.InterpretationCode, Display: strVal(o.InterpretationDisplay)}},
		}}
	}
	if o.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", o.EncounterID.String())}
	}
	if o.Note != nil {
		result["note"] = []map[string]string{{"text": *o.Note}}
	}
	return result
}

// ObservationComponent maps to observation_component table.
type ObservationComponent struct {
	ID                    uuid.UUID `db:"id" json:"id"`
	ObservationID         uuid.UUID `db:"observation_id" json:"observation_id"`
	CodeSystem            *string   `db:"code_system" json:"code_system,omitempty"`
	CodeValue             string    `db:"code_value" json:"code_value"`
	CodeDisplay           string    `db:"code_display" json:"code_display"`
	ValueQuantity         *float64  `db:"value_quantity" json:"value_quantity,omitempty"`
	ValueUnit             *string   `db:"value_unit" json:"value_unit,omitempty"`
	ValueString           *string   `db:"value_string" json:"value_string,omitempty"`
	ValueCodeableCode     *string   `db:"value_codeable_code" json:"value_codeable_code,omitempty"`
	ValueCodeableDisplay  *string   `db:"value_codeable_display" json:"value_codeable_display,omitempty"`
	InterpretationCode    *string   `db:"interpretation_code" json:"interpretation_code,omitempty"`
	InterpretationDisplay *string   `db:"interpretation_display" json:"interpretation_display,omitempty"`
	ReferenceRangeLow     *float64  `db:"reference_range_low" json:"reference_range_low,omitempty"`
	ReferenceRangeHigh    *float64  `db:"reference_range_high" json:"reference_range_high,omitempty"`
	ReferenceRangeUnit    *string   `db:"reference_range_unit" json:"reference_range_unit,omitempty"`
}

// AllergyIntolerance maps to the allergy_intolerance table.
type AllergyIntolerance struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	FHIRID             string     `db:"fhir_id" json:"fhir_id"`
	PatientID          uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID        *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	RecorderID         *uuid.UUID `db:"recorder_id" json:"recorder_id,omitempty"`
	AsserterID         *uuid.UUID `db:"asserter_id" json:"asserter_id,omitempty"`
	ClinicalStatus     *string    `db:"clinical_status" json:"clinical_status,omitempty"`
	VerificationStatus *string    `db:"verification_status" json:"verification_status,omitempty"`
	Type               *string    `db:"type" json:"type,omitempty"`
	Category           []string   `db:"category" json:"category,omitempty"`
	Criticality        *string    `db:"criticality" json:"criticality,omitempty"`
	CodeSystem         *string    `db:"code_system" json:"code_system,omitempty"`
	CodeValue          *string    `db:"code_value" json:"code_value,omitempty"`
	CodeDisplay        *string    `db:"code_display" json:"code_display,omitempty"`
	OnsetDatetime      *time.Time `db:"onset_datetime" json:"onset_datetime,omitempty"`
	OnsetAge           *int       `db:"onset_age" json:"onset_age,omitempty"`
	OnsetString        *string    `db:"onset_string" json:"onset_string,omitempty"`
	RecordedDate       *time.Time `db:"recorded_date" json:"recorded_date,omitempty"`
	LastOccurrence     *time.Time `db:"last_occurrence" json:"last_occurrence,omitempty"`
	Note               *string    `db:"note" json:"note,omitempty"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

func (a *AllergyIntolerance) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "AllergyIntolerance",
		"id":           a.FHIRID,
		"patient":      fhir.Reference{Reference: fhir.FormatReference("Patient", a.PatientID.String())},
		"meta":         fhir.Meta{LastUpdated: a.UpdatedAt},
	}
	if a.ClinicalStatus != nil {
		result["clinicalStatus"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System: "http://terminology.hl7.org/CodeSystem/allergyintolerance-clinical",
				Code:   *a.ClinicalStatus,
			}},
		}
	}
	if a.VerificationStatus != nil {
		result["verificationStatus"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System: "http://terminology.hl7.org/CodeSystem/allergyintolerance-verification",
				Code:   *a.VerificationStatus,
			}},
		}
	}
	if a.Type != nil {
		result["type"] = *a.Type
	}
	if len(a.Category) > 0 {
		result["category"] = a.Category
	}
	if a.Criticality != nil {
		result["criticality"] = *a.Criticality
	}
	if a.CodeValue != nil {
		result["code"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{System: strVal(a.CodeSystem), Code: *a.CodeValue, Display: strVal(a.CodeDisplay)}},
		}
	}
	if a.OnsetDatetime != nil {
		result["onsetDateTime"] = a.OnsetDatetime.Format(time.RFC3339)
	}
	if a.RecordedDate != nil {
		result["recordedDate"] = a.RecordedDate.Format("2006-01-02")
	}
	return result
}

// AllergyReaction maps to allergy_reaction table.
type AllergyReaction struct {
	ID                   uuid.UUID  `db:"id" json:"id"`
	AllergyID            uuid.UUID  `db:"allergy_id" json:"allergy_id"`
	SubstanceCode        *string    `db:"substance_code" json:"substance_code,omitempty"`
	SubstanceDisplay     *string    `db:"substance_display" json:"substance_display,omitempty"`
	ManifestationCode    string     `db:"manifestation_code" json:"manifestation_code"`
	ManifestationDisplay string     `db:"manifestation_display" json:"manifestation_display"`
	Description          *string    `db:"description" json:"description,omitempty"`
	Severity             *string    `db:"severity" json:"severity,omitempty"`
	ExposureRouteCode    *string    `db:"exposure_route_code" json:"exposure_route_code,omitempty"`
	ExposureRouteDisplay *string    `db:"exposure_route_display" json:"exposure_route_display,omitempty"`
	Onset                *time.Time `db:"onset" json:"onset,omitempty"`
	Note                 *string    `db:"note" json:"note,omitempty"`
}

// ProcedureRecord maps to the procedure_record table (FHIR Procedure resource).
type ProcedureRecord struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	FHIRID             string     `db:"fhir_id" json:"fhir_id"`
	Status             string     `db:"status" json:"status"`
	StatusReasonCode   *string    `db:"status_reason_code" json:"status_reason_code,omitempty"`
	PatientID          uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID        *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	RecorderID         *uuid.UUID `db:"recorder_id" json:"recorder_id,omitempty"`
	AsserterID         *uuid.UUID `db:"asserter_id" json:"asserter_id,omitempty"`
	CodeSystem         *string    `db:"code_system" json:"code_system,omitempty"`
	CodeValue          string     `db:"code_value" json:"code_value"`
	CodeDisplay        string     `db:"code_display" json:"code_display"`
	CategoryCode       *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay    *string    `db:"category_display" json:"category_display,omitempty"`
	PerformedDatetime  *time.Time `db:"performed_datetime" json:"performed_datetime,omitempty"`
	PerformedStart     *time.Time `db:"performed_start" json:"performed_start,omitempty"`
	PerformedEnd       *time.Time `db:"performed_end" json:"performed_end,omitempty"`
	PerformedString    *string    `db:"performed_string" json:"performed_string,omitempty"`
	BodySiteCode       *string    `db:"body_site_code" json:"body_site_code,omitempty"`
	BodySiteDisplay    *string    `db:"body_site_display" json:"body_site_display,omitempty"`
	OutcomeCode        *string    `db:"outcome_code" json:"outcome_code,omitempty"`
	OutcomeDisplay     *string    `db:"outcome_display" json:"outcome_display,omitempty"`
	ComplicationCode   *string    `db:"complication_code" json:"complication_code,omitempty"`
	ComplicationDisp   *string    `db:"complication_display" json:"complication_display,omitempty"`
	ReasonCode         *string    `db:"reason_code" json:"reason_code,omitempty"`
	ReasonDisplay      *string    `db:"reason_display" json:"reason_display,omitempty"`
	ReasonConditionID  *uuid.UUID `db:"reason_condition_id" json:"reason_condition_id,omitempty"`
	LocationID         *uuid.UUID `db:"location_id" json:"location_id,omitempty"`
	AnesthesiaType     *string    `db:"anesthesia_type" json:"anesthesia_type,omitempty"`
	CPTCode            *string    `db:"cpt_code" json:"cpt_code,omitempty"`
	HCPCSCode          *string    `db:"hcpcs_code" json:"hcpcs_code,omitempty"`
	Note               *string    `db:"note" json:"note,omitempty"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

func (p *ProcedureRecord) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Procedure",
		"id":           p.FHIRID,
		"status":       p.Status,
		"code": fhir.CodeableConcept{
			Coding: []fhir.Coding{{System: strVal(p.CodeSystem), Code: p.CodeValue, Display: p.CodeDisplay}},
		},
		"subject": fhir.Reference{Reference: fhir.FormatReference("Patient", p.PatientID.String())},
		"meta":    fhir.Meta{LastUpdated: p.UpdatedAt},
	}
	if p.CategoryCode != nil {
		result["category"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *p.CategoryCode, Display: strVal(p.CategoryDisplay)}},
		}
	}
	if p.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", p.EncounterID.String())}
	}
	if p.PerformedDatetime != nil {
		result["performedDateTime"] = p.PerformedDatetime.Format(time.RFC3339)
	} else if p.PerformedStart != nil {
		period := fhir.Period{Start: p.PerformedStart, End: p.PerformedEnd}
		result["performedPeriod"] = period
	}
	if p.BodySiteCode != nil {
		result["bodySite"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *p.BodySiteCode, Display: strVal(p.BodySiteDisplay)}},
		}}
	}
	if p.OutcomeCode != nil {
		result["outcome"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *p.OutcomeCode, Display: strVal(p.OutcomeDisplay)}},
		}
	}
	if p.ReasonCode != nil {
		result["reasonCode"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *p.ReasonCode, Display: strVal(p.ReasonDisplay)}},
		}}
	}
	if p.LocationID != nil {
		result["location"] = fhir.Reference{Reference: fhir.FormatReference("Location", p.LocationID.String())}
	}
	if p.Note != nil {
		result["note"] = []map[string]string{{"text": *p.Note}}
	}
	return result
}

// ProcedurePerformer maps to procedure_performer table.
type ProcedurePerformer struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	ProcedureID    uuid.UUID  `db:"procedure_id" json:"procedure_id"`
	PractitionerID uuid.UUID  `db:"practitioner_id" json:"practitioner_id"`
	RoleCode       *string    `db:"role_code" json:"role_code,omitempty"`
	RoleDisplay    *string    `db:"role_display" json:"role_display,omitempty"`
	OrganizationID *uuid.UUID `db:"organization_id" json:"organization_id,omitempty"`
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
