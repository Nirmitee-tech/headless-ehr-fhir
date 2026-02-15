package diagnostics

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// ServiceRequest maps to the service_request table (FHIR ServiceRequest resource).
type ServiceRequest struct {
	ID                  uuid.UUID  `db:"id" json:"id"`
	FHIRID              string     `db:"fhir_id" json:"fhir_id"`
	PatientID           uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID         *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	RequesterID         uuid.UUID  `db:"requester_id" json:"requester_id"`
	PerformerID         *uuid.UUID `db:"performer_id" json:"performer_id,omitempty"`
	Status              string     `db:"status" json:"status"`
	Intent              string     `db:"intent" json:"intent"`
	Priority            *string    `db:"priority" json:"priority,omitempty"`
	CategoryCode        *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay     *string    `db:"category_display" json:"category_display,omitempty"`
	CodeSystem          *string    `db:"code_system" json:"code_system,omitempty"`
	CodeValue           string     `db:"code_value" json:"code_value"`
	CodeDisplay         string     `db:"code_display" json:"code_display"`
	OrderDetailCode     *string    `db:"order_detail_code" json:"order_detail_code,omitempty"`
	OrderDetailDisplay  *string    `db:"order_detail_display" json:"order_detail_display,omitempty"`
	QuantityValue       *float64   `db:"quantity_value" json:"quantity_value,omitempty"`
	QuantityUnit        *string    `db:"quantity_unit" json:"quantity_unit,omitempty"`
	OccurrenceDatetime  *time.Time `db:"occurrence_datetime" json:"occurrence_datetime,omitempty"`
	OccurrenceStart     *time.Time `db:"occurrence_start" json:"occurrence_start,omitempty"`
	OccurrenceEnd       *time.Time `db:"occurrence_end" json:"occurrence_end,omitempty"`
	AuthoredOn          *time.Time `db:"authored_on" json:"authored_on,omitempty"`
	ReasonCode          *string    `db:"reason_code" json:"reason_code,omitempty"`
	ReasonDisplay       *string    `db:"reason_display" json:"reason_display,omitempty"`
	ReasonConditionID   *uuid.UUID `db:"reason_condition_id" json:"reason_condition_id,omitempty"`
	SpecimenRequirement *string    `db:"specimen_requirement" json:"specimen_requirement,omitempty"`
	BodySiteCode        *string    `db:"body_site_code" json:"body_site_code,omitempty"`
	BodySiteDisplay     *string    `db:"body_site_display" json:"body_site_display,omitempty"`
	Note                *string    `db:"note" json:"note,omitempty"`
	PatientInstruction  *string    `db:"patient_instruction" json:"patient_instruction,omitempty"`
	VersionID           int        `db:"version_id" json:"version_id"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (sr *ServiceRequest) GetVersionID() int { return sr.VersionID }

// SetVersionID sets the current version.
func (sr *ServiceRequest) SetVersionID(v int) { sr.VersionID = v }

func (sr *ServiceRequest) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ServiceRequest",
		"id":           sr.FHIRID,
		"status":       sr.Status,
		"intent":       sr.Intent,
		"code": fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System:  strVal(sr.CodeSystem),
				Code:    sr.CodeValue,
				Display: sr.CodeDisplay,
			}},
			Text: sr.CodeDisplay,
		},
		"subject":   fhir.Reference{Reference: fhir.FormatReference("Patient", sr.PatientID.String())},
		"requester": fhir.Reference{Reference: fhir.FormatReference("Practitioner", sr.RequesterID.String())},
		"meta":      fhir.Meta{LastUpdated: sr.UpdatedAt},
	}
	if sr.Priority != nil {
		result["priority"] = *sr.Priority
	}
	if sr.CategoryCode != nil {
		result["category"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *sr.CategoryCode, Display: strVal(sr.CategoryDisplay)}},
		}}
	}
	if sr.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", sr.EncounterID.String())}
	}
	if sr.PerformerID != nil {
		result["performer"] = []fhir.Reference{{Reference: fhir.FormatReference("Practitioner", sr.PerformerID.String())}}
	}
	if sr.OccurrenceDatetime != nil {
		result["occurrenceDateTime"] = sr.OccurrenceDatetime.Format(time.RFC3339)
	} else if sr.OccurrenceStart != nil {
		period := fhir.Period{Start: sr.OccurrenceStart, End: sr.OccurrenceEnd}
		result["occurrencePeriod"] = period
	}
	if sr.AuthoredOn != nil {
		result["authoredOn"] = sr.AuthoredOn.Format(time.RFC3339)
	}
	if sr.ReasonCode != nil {
		result["reasonCode"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *sr.ReasonCode, Display: strVal(sr.ReasonDisplay)}},
		}}
	}
	if sr.BodySiteCode != nil {
		result["bodySite"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *sr.BodySiteCode, Display: strVal(sr.BodySiteDisplay)}},
		}}
	}
	if sr.Note != nil {
		result["note"] = []map[string]string{{"text": *sr.Note}}
	}
	if sr.PatientInstruction != nil {
		result["patientInstruction"] = *sr.PatientInstruction
	}
	return result
}

// Specimen maps to the specimen table (FHIR Specimen resource).
type Specimen struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	FHIRID             string     `db:"fhir_id" json:"fhir_id"`
	PatientID          uuid.UUID  `db:"patient_id" json:"patient_id"`
	AccessionID        *string    `db:"accession_id" json:"accession_id,omitempty"`
	Status             string     `db:"status" json:"status"`
	TypeCode           *string    `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay        *string    `db:"type_display" json:"type_display,omitempty"`
	ReceivedTime       *time.Time `db:"received_time" json:"received_time,omitempty"`
	CollectionCollector *uuid.UUID `db:"collection_collector" json:"collection_collector,omitempty"`
	CollectionDatetime *time.Time `db:"collection_datetime" json:"collection_datetime,omitempty"`
	CollectionQuantity *float64   `db:"collection_quantity" json:"collection_quantity,omitempty"`
	CollectionUnit     *string    `db:"collection_unit" json:"collection_unit,omitempty"`
	CollectionMethod   *string    `db:"collection_method" json:"collection_method,omitempty"`
	CollectionBodySite *string    `db:"collection_body_site" json:"collection_body_site,omitempty"`
	ProcessingDesc     *string    `db:"processing_description" json:"processing_description,omitempty"`
	ProcessingProcedure *string   `db:"processing_procedure" json:"processing_procedure,omitempty"`
	ProcessingDatetime *time.Time `db:"processing_datetime" json:"processing_datetime,omitempty"`
	ContainerDesc      *string    `db:"container_description" json:"container_description,omitempty"`
	ContainerType      *string    `db:"container_type" json:"container_type,omitempty"`
	ConditionCode      *string    `db:"condition_code" json:"condition_code,omitempty"`
	ConditionDisplay   *string    `db:"condition_display" json:"condition_display,omitempty"`
	Note               *string    `db:"note" json:"note,omitempty"`
	VersionID          int        `db:"version_id" json:"version_id"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (sp *Specimen) GetVersionID() int { return sp.VersionID }

// SetVersionID sets the current version.
func (sp *Specimen) SetVersionID(v int) { sp.VersionID = v }

func (sp *Specimen) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Specimen",
		"id":           sp.FHIRID,
		"status":       sp.Status,
		"subject":      fhir.Reference{Reference: fhir.FormatReference("Patient", sp.PatientID.String())},
		"meta":         fhir.Meta{LastUpdated: sp.UpdatedAt},
	}
	if sp.AccessionID != nil {
		result["accessionIdentifier"] = fhir.Identifier{Value: *sp.AccessionID}
	}
	if sp.TypeCode != nil {
		result["type"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *sp.TypeCode, Display: strVal(sp.TypeDisplay)}},
		}
	}
	if sp.ReceivedTime != nil {
		result["receivedTime"] = sp.ReceivedTime.Format(time.RFC3339)
	}
	collection := map[string]interface{}{}
	if sp.CollectionCollector != nil {
		collection["collector"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", sp.CollectionCollector.String())}
	}
	if sp.CollectionDatetime != nil {
		collection["collectedDateTime"] = sp.CollectionDatetime.Format(time.RFC3339)
	}
	if sp.CollectionQuantity != nil {
		collection["quantity"] = map[string]interface{}{
			"value": *sp.CollectionQuantity,
			"unit":  strVal(sp.CollectionUnit),
		}
	}
	if sp.CollectionMethod != nil {
		collection["method"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *sp.CollectionMethod}},
		}
	}
	if sp.CollectionBodySite != nil {
		collection["bodySite"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *sp.CollectionBodySite}},
		}
	}
	if len(collection) > 0 {
		result["collection"] = collection
	}
	if sp.ConditionCode != nil {
		result["condition"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *sp.ConditionCode, Display: strVal(sp.ConditionDisplay)}},
		}}
	}
	if sp.Note != nil {
		result["note"] = []map[string]string{{"text": *sp.Note}}
	}
	return result
}

// DiagnosticReport maps to the diagnostic_report table (FHIR DiagnosticReport resource).
type DiagnosticReport struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	FHIRID            string     `db:"fhir_id" json:"fhir_id"`
	PatientID         uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID       *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	PerformerID       *uuid.UUID `db:"performer_id" json:"performer_id,omitempty"`
	Status            string     `db:"status" json:"status"`
	CategoryCode      *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay   *string    `db:"category_display" json:"category_display,omitempty"`
	CodeSystem        *string    `db:"code_system" json:"code_system,omitempty"`
	CodeValue         string     `db:"code_value" json:"code_value"`
	CodeDisplay       string     `db:"code_display" json:"code_display"`
	EffectiveDatetime *time.Time `db:"effective_datetime" json:"effective_datetime,omitempty"`
	EffectiveStart    *time.Time `db:"effective_start" json:"effective_start,omitempty"`
	EffectiveEnd      *time.Time `db:"effective_end" json:"effective_end,omitempty"`
	Issued            *time.Time `db:"issued" json:"issued,omitempty"`
	SpecimenID        *uuid.UUID `db:"specimen_id" json:"specimen_id,omitempty"`
	Conclusion        *string    `db:"conclusion" json:"conclusion,omitempty"`
	ConclusionCode    *string    `db:"conclusion_code" json:"conclusion_code,omitempty"`
	ConclusionDisplay *string    `db:"conclusion_display" json:"conclusion_display,omitempty"`
	PresentedFormURL  *string    `db:"presented_form_url" json:"presented_form_url,omitempty"`
	PresentedFormType *string    `db:"presented_form_type" json:"presented_form_type,omitempty"`
	Note              *string    `db:"note" json:"note,omitempty"`
	VersionID         int        `db:"version_id" json:"version_id"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (dr *DiagnosticReport) GetVersionID() int { return dr.VersionID }

// SetVersionID sets the current version.
func (dr *DiagnosticReport) SetVersionID(v int) { dr.VersionID = v }

func (dr *DiagnosticReport) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "DiagnosticReport",
		"id":           dr.FHIRID,
		"status":       dr.Status,
		"code": fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System:  strVal(dr.CodeSystem),
				Code:    dr.CodeValue,
				Display: dr.CodeDisplay,
			}},
			Text: dr.CodeDisplay,
		},
		"subject": fhir.Reference{Reference: fhir.FormatReference("Patient", dr.PatientID.String())},
		"meta":    fhir.Meta{LastUpdated: dr.UpdatedAt},
	}
	if dr.CategoryCode != nil {
		result["category"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *dr.CategoryCode, Display: strVal(dr.CategoryDisplay)}},
		}}
	}
	if dr.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", dr.EncounterID.String())}
	}
	if dr.PerformerID != nil {
		result["performer"] = []fhir.Reference{{Reference: fhir.FormatReference("Practitioner", dr.PerformerID.String())}}
	}
	if dr.EffectiveDatetime != nil {
		result["effectiveDateTime"] = dr.EffectiveDatetime.Format(time.RFC3339)
	} else if dr.EffectiveStart != nil {
		result["effectivePeriod"] = fhir.Period{Start: dr.EffectiveStart, End: dr.EffectiveEnd}
	}
	if dr.Issued != nil {
		result["issued"] = dr.Issued.Format(time.RFC3339)
	}
	if dr.SpecimenID != nil {
		result["specimen"] = []fhir.Reference{{Reference: fhir.FormatReference("Specimen", dr.SpecimenID.String())}}
	}
	if dr.Conclusion != nil {
		result["conclusion"] = *dr.Conclusion
	}
	if dr.ConclusionCode != nil {
		result["conclusionCode"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *dr.ConclusionCode, Display: strVal(dr.ConclusionDisplay)}},
		}}
	}
	if dr.PresentedFormURL != nil {
		result["presentedForm"] = []map[string]string{{
			"url":         *dr.PresentedFormURL,
			"contentType": strVal(dr.PresentedFormType),
		}}
	}
	return result
}

// DiagnosticReportResult maps to the diagnostic_report_result junction table.
type DiagnosticReportResult struct {
	DiagnosticReportID uuid.UUID `db:"diagnostic_report_id" json:"diagnostic_report_id"`
	ObservationID      uuid.UUID `db:"observation_id" json:"observation_id"`
}

// ImagingStudy maps to the imaging_study table (FHIR ImagingStudy resource).
type ImagingStudy struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	FHIRID           string     `db:"fhir_id" json:"fhir_id"`
	PatientID        uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID      *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	ReferrerID       *uuid.UUID `db:"referrer_id" json:"referrer_id,omitempty"`
	Status           string     `db:"status" json:"status"`
	ModalityCode     *string    `db:"modality_code" json:"modality_code,omitempty"`
	ModalityDisplay  *string    `db:"modality_display" json:"modality_display,omitempty"`
	StudyUID         *string    `db:"study_uid" json:"study_uid,omitempty"`
	NumberOfSeries   *int       `db:"number_of_series" json:"number_of_series,omitempty"`
	NumberOfInstances *int      `db:"number_of_instances" json:"number_of_instances,omitempty"`
	Description      *string    `db:"description" json:"description,omitempty"`
	Started          *time.Time `db:"started" json:"started,omitempty"`
	Endpoint         *string    `db:"endpoint" json:"endpoint,omitempty"`
	ReasonCode       *string    `db:"reason_code" json:"reason_code,omitempty"`
	ReasonDisplay    *string    `db:"reason_display" json:"reason_display,omitempty"`
	Note             *string    `db:"note" json:"note,omitempty"`
	VersionID        int        `db:"version_id" json:"version_id"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (is *ImagingStudy) GetVersionID() int { return is.VersionID }

// SetVersionID sets the current version.
func (is *ImagingStudy) SetVersionID(v int) { is.VersionID = v }

func (is *ImagingStudy) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ImagingStudy",
		"id":           is.FHIRID,
		"status":       is.Status,
		"subject":      fhir.Reference{Reference: fhir.FormatReference("Patient", is.PatientID.String())},
		"meta":         fhir.Meta{LastUpdated: is.UpdatedAt},
	}
	if is.ModalityCode != nil {
		result["modality"] = []fhir.Coding{{Code: *is.ModalityCode, Display: strVal(is.ModalityDisplay)}}
	}
	if is.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", is.EncounterID.String())}
	}
	if is.ReferrerID != nil {
		result["referrer"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", is.ReferrerID.String())}
	}
	if is.StudyUID != nil {
		result["identifier"] = []fhir.Identifier{{
			System: "urn:dicom:uid",
			Value:  *is.StudyUID,
		}}
	}
	if is.NumberOfSeries != nil {
		result["numberOfSeries"] = *is.NumberOfSeries
	}
	if is.NumberOfInstances != nil {
		result["numberOfInstances"] = *is.NumberOfInstances
	}
	if is.Description != nil {
		result["description"] = *is.Description
	}
	if is.Started != nil {
		result["started"] = is.Started.Format(time.RFC3339)
	}
	if is.Endpoint != nil {
		result["endpoint"] = []fhir.Reference{{Reference: *is.Endpoint}}
	}
	if is.ReasonCode != nil {
		result["reasonCode"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *is.ReasonCode, Display: strVal(is.ReasonDisplay)}},
		}}
	}
	if is.Note != nil {
		result["note"] = []map[string]string{{"text": *is.Note}}
	}
	return result
}

// OrderStatusHistory records a status transition for service requests or medication requests.
type OrderStatusHistory struct {
	ID           uuid.UUID `db:"id" json:"id"`
	ResourceType string    `db:"resource_type" json:"resource_type"` // ServiceRequest, MedicationRequest
	ResourceID   uuid.UUID `db:"resource_id" json:"resource_id"`
	FromStatus   string    `db:"from_status" json:"from_status"`
	ToStatus     string    `db:"to_status" json:"to_status"`
	ChangedBy    string    `db:"changed_by" json:"changed_by"`
	ChangedAt    time.Time `db:"changed_at" json:"changed_at"`
	Reason       *string   `db:"reason" json:"reason,omitempty"`
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
