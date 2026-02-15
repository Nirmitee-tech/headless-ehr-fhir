package financial

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Account maps to the account table (FHIR Account resource).
type Account struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	FHIRID             string     `db:"fhir_id" json:"fhir_id"`
	Status             string     `db:"status" json:"status"`
	TypeCode           *string    `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay        *string    `db:"type_display" json:"type_display,omitempty"`
	Name               *string    `db:"name" json:"name,omitempty"`
	SubjectPatientID   *uuid.UUID `db:"subject_patient_id" json:"subject_patient_id,omitempty"`
	ServicePeriodStart *time.Time `db:"service_period_start" json:"service_period_start,omitempty"`
	ServicePeriodEnd   *time.Time `db:"service_period_end" json:"service_period_end,omitempty"`
	OwnerOrgID         *uuid.UUID `db:"owner_org_id" json:"owner_org_id,omitempty"`
	Description        *string    `db:"description" json:"description,omitempty"`
	VersionID          int        `db:"version_id" json:"version_id"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

func (a *Account) GetVersionID() int  { return a.VersionID }
func (a *Account) SetVersionID(v int) { a.VersionID = v }

func (a *Account) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Account",
		"id":           a.FHIRID,
		"status":       a.Status,
		"meta":         fhir.Meta{LastUpdated: a.UpdatedAt},
	}
	if a.TypeCode != nil {
		cc := fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *a.TypeCode}},
		}
		if a.TypeDisplay != nil {
			cc.Coding[0].Display = *a.TypeDisplay
		}
		result["type"] = cc
	}
	if a.Name != nil {
		result["name"] = *a.Name
	}
	if a.SubjectPatientID != nil {
		result["subject"] = []fhir.Reference{{Reference: fhir.FormatReference("Patient", a.SubjectPatientID.String())}}
	}
	if a.ServicePeriodStart != nil || a.ServicePeriodEnd != nil {
		result["servicePeriod"] = fhir.Period{Start: a.ServicePeriodStart, End: a.ServicePeriodEnd}
	}
	if a.OwnerOrgID != nil {
		result["owner"] = fhir.Reference{Reference: fhir.FormatReference("Organization", a.OwnerOrgID.String())}
	}
	if a.Description != nil {
		result["description"] = *a.Description
	}
	return result
}

// InsurancePlan maps to the insurance_plan table (FHIR InsurancePlan resource).
type InsurancePlan struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	FHIRID             string     `db:"fhir_id" json:"fhir_id"`
	Status             string     `db:"status" json:"status"`
	TypeCode           *string    `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay        *string    `db:"type_display" json:"type_display,omitempty"`
	Name               *string    `db:"name" json:"name,omitempty"`
	Alias              *string    `db:"alias" json:"alias,omitempty"`
	PeriodStart        *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd          *time.Time `db:"period_end" json:"period_end,omitempty"`
	OwnedByOrgID       *uuid.UUID `db:"owned_by_org_id" json:"owned_by_org_id,omitempty"`
	AdministeredByOrgID *uuid.UUID `db:"administered_by_org_id" json:"administered_by_org_id,omitempty"`
	CoverageArea       *string    `db:"coverage_area" json:"coverage_area,omitempty"`
	NetworkName        *string    `db:"network_name" json:"network_name,omitempty"`
	VersionID          int        `db:"version_id" json:"version_id"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

func (ip *InsurancePlan) GetVersionID() int  { return ip.VersionID }
func (ip *InsurancePlan) SetVersionID(v int) { ip.VersionID = v }

func (ip *InsurancePlan) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "InsurancePlan",
		"id":           ip.FHIRID,
		"status":       ip.Status,
		"meta":         fhir.Meta{LastUpdated: ip.UpdatedAt},
	}
	if ip.TypeCode != nil {
		cc := fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *ip.TypeCode}},
		}
		if ip.TypeDisplay != nil {
			cc.Coding[0].Display = *ip.TypeDisplay
		}
		result["type"] = []fhir.CodeableConcept{cc}
	}
	if ip.Name != nil {
		result["name"] = *ip.Name
	}
	if ip.Alias != nil {
		result["alias"] = []string{*ip.Alias}
	}
	if ip.PeriodStart != nil || ip.PeriodEnd != nil {
		result["period"] = fhir.Period{Start: ip.PeriodStart, End: ip.PeriodEnd}
	}
	if ip.OwnedByOrgID != nil {
		result["ownedBy"] = fhir.Reference{Reference: fhir.FormatReference("Organization", ip.OwnedByOrgID.String())}
	}
	if ip.AdministeredByOrgID != nil {
		result["administeredBy"] = fhir.Reference{Reference: fhir.FormatReference("Organization", ip.AdministeredByOrgID.String())}
	}
	if ip.CoverageArea != nil {
		result["coverageArea"] = []fhir.Reference{{Display: *ip.CoverageArea}}
	}
	if ip.NetworkName != nil {
		result["network"] = []map[string]interface{}{{"name": *ip.NetworkName}}
	}
	return result
}

// PaymentNotice maps to the payment_notice table (FHIR PaymentNotice resource).
type PaymentNotice struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	FHIRID            string     `db:"fhir_id" json:"fhir_id"`
	Status            string     `db:"status" json:"status"`
	RequestReference  *string    `db:"request_reference" json:"request_reference,omitempty"`
	ResponseReference *string    `db:"response_reference" json:"response_reference,omitempty"`
	Created           time.Time  `db:"created" json:"created"`
	ProviderID        *uuid.UUID `db:"provider_id" json:"provider_id,omitempty"`
	PaymentReference  *string    `db:"payment_reference" json:"payment_reference,omitempty"`
	PaymentDate       *time.Time `db:"payment_date" json:"payment_date,omitempty"`
	PayeeOrgID        *uuid.UUID `db:"payee_org_id" json:"payee_org_id,omitempty"`
	RecipientOrgID    *uuid.UUID `db:"recipient_org_id" json:"recipient_org_id,omitempty"`
	AmountValue       *float64   `db:"amount_value" json:"amount_value,omitempty"`
	AmountCurrency    *string    `db:"amount_currency" json:"amount_currency,omitempty"`
	PaymentStatusCode *string    `db:"payment_status_code" json:"payment_status_code,omitempty"`
	VersionID         int        `db:"version_id" json:"version_id"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

func (pn *PaymentNotice) GetVersionID() int  { return pn.VersionID }
func (pn *PaymentNotice) SetVersionID(v int) { pn.VersionID = v }

func (pn *PaymentNotice) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "PaymentNotice",
		"id":           pn.FHIRID,
		"status":       pn.Status,
		"created":      pn.Created.Format("2006-01-02T15:04:05Z"),
		"meta":         fhir.Meta{LastUpdated: pn.UpdatedAt},
	}
	if pn.RequestReference != nil {
		result["request"] = fhir.Reference{Reference: *pn.RequestReference}
	}
	if pn.ResponseReference != nil {
		result["response"] = fhir.Reference{Reference: *pn.ResponseReference}
	}
	if pn.ProviderID != nil {
		result["provider"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", pn.ProviderID.String())}
	}
	if pn.PaymentReference != nil {
		result["payment"] = fhir.Reference{Reference: *pn.PaymentReference}
	}
	if pn.PaymentDate != nil {
		result["paymentDate"] = pn.PaymentDate.Format("2006-01-02")
	}
	if pn.PayeeOrgID != nil {
		result["payee"] = fhir.Reference{Reference: fhir.FormatReference("Organization", pn.PayeeOrgID.String())}
	}
	if pn.RecipientOrgID != nil {
		result["recipient"] = fhir.Reference{Reference: fhir.FormatReference("Organization", pn.RecipientOrgID.String())}
	}
	if pn.AmountValue != nil {
		amount := map[string]interface{}{"value": *pn.AmountValue}
		if pn.AmountCurrency != nil {
			amount["currency"] = *pn.AmountCurrency
		}
		result["amount"] = amount
	}
	if pn.PaymentStatusCode != nil {
		result["paymentStatus"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *pn.PaymentStatusCode}},
		}
	}
	return result
}

// PaymentReconciliation maps to the payment_reconciliation table (FHIR PaymentReconciliation resource).
type PaymentReconciliation struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	FHIRID            string     `db:"fhir_id" json:"fhir_id"`
	Status            string     `db:"status" json:"status"`
	PeriodStart       *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd         *time.Time `db:"period_end" json:"period_end,omitempty"`
	Created           time.Time  `db:"created" json:"created"`
	PaymentIssuerOrgID *uuid.UUID `db:"payment_issuer_org_id" json:"payment_issuer_org_id,omitempty"`
	RequestReference  *string    `db:"request_reference" json:"request_reference,omitempty"`
	RequestorID       *uuid.UUID `db:"requestor_id" json:"requestor_id,omitempty"`
	Outcome           *string    `db:"outcome" json:"outcome,omitempty"`
	Disposition       *string    `db:"disposition" json:"disposition,omitempty"`
	PaymentDate       time.Time  `db:"payment_date" json:"payment_date"`
	PaymentAmount     float64    `db:"payment_amount" json:"payment_amount"`
	PaymentCurrency   *string    `db:"payment_currency" json:"payment_currency,omitempty"`
	PaymentIdentifier *string    `db:"payment_identifier" json:"payment_identifier,omitempty"`
	FormCode          *string    `db:"form_code" json:"form_code,omitempty"`
	ProcessNote       *string    `db:"process_note" json:"process_note,omitempty"`
	VersionID         int        `db:"version_id" json:"version_id"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

func (pr *PaymentReconciliation) GetVersionID() int  { return pr.VersionID }
func (pr *PaymentReconciliation) SetVersionID(v int) { pr.VersionID = v }

func (pr *PaymentReconciliation) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "PaymentReconciliation",
		"id":           pr.FHIRID,
		"status":       pr.Status,
		"created":      pr.Created.Format("2006-01-02T15:04:05Z"),
		"paymentDate":  pr.PaymentDate.Format("2006-01-02"),
		"meta":         fhir.Meta{LastUpdated: pr.UpdatedAt},
	}
	cur := "USD"
	if pr.PaymentCurrency != nil {
		cur = *pr.PaymentCurrency
	}
	result["paymentAmount"] = map[string]interface{}{
		"value":    pr.PaymentAmount,
		"currency": cur,
	}
	if pr.PeriodStart != nil || pr.PeriodEnd != nil {
		result["period"] = fhir.Period{Start: pr.PeriodStart, End: pr.PeriodEnd}
	}
	if pr.PaymentIssuerOrgID != nil {
		result["paymentIssuer"] = fhir.Reference{Reference: fhir.FormatReference("Organization", pr.PaymentIssuerOrgID.String())}
	}
	if pr.RequestReference != nil {
		result["request"] = fhir.Reference{Reference: *pr.RequestReference}
	}
	if pr.RequestorID != nil {
		result["requestor"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", pr.RequestorID.String())}
	}
	if pr.Outcome != nil {
		result["outcome"] = *pr.Outcome
	}
	if pr.Disposition != nil {
		result["disposition"] = *pr.Disposition
	}
	if pr.PaymentIdentifier != nil {
		result["paymentIdentifier"] = fhir.Identifier{Value: *pr.PaymentIdentifier}
	}
	if pr.FormCode != nil {
		result["formCode"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *pr.FormCode}},
		}
	}
	if pr.ProcessNote != nil {
		result["processNote"] = []map[string]string{{"text": *pr.ProcessNote}}
	}
	return result
}

// ChargeItem maps to the charge_item table (FHIR ChargeItem resource).
type ChargeItem struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	FHIRID                string     `db:"fhir_id" json:"fhir_id"`
	Status                string     `db:"status" json:"status"`
	CodeCode              *string    `db:"code_code" json:"code_code,omitempty"`
	CodeDisplay           *string    `db:"code_display" json:"code_display,omitempty"`
	CodeSystem            *string    `db:"code_system" json:"code_system,omitempty"`
	SubjectPatientID      uuid.UUID  `db:"subject_patient_id" json:"subject_patient_id"`
	ContextEncounterID    *uuid.UUID `db:"context_encounter_id" json:"context_encounter_id,omitempty"`
	OccurrenceDate        *time.Time `db:"occurrence_date" json:"occurrence_date,omitempty"`
	PerformerID           *uuid.UUID `db:"performer_id" json:"performer_id,omitempty"`
	PerformingOrgID       *uuid.UUID `db:"performing_org_id" json:"performing_org_id,omitempty"`
	QuantityValue         *float64   `db:"quantity_value" json:"quantity_value,omitempty"`
	FactorOverride        *float64   `db:"factor_override" json:"factor_override,omitempty"`
	PriceOverrideValue    *float64   `db:"price_override_value" json:"price_override_value,omitempty"`
	PriceOverrideCurrency *string    `db:"price_override_currency" json:"price_override_currency,omitempty"`
	OverrideReason        *string    `db:"override_reason" json:"override_reason,omitempty"`
	EntererID             *uuid.UUID `db:"enterer_id" json:"enterer_id,omitempty"`
	EnteredDate           *time.Time `db:"entered_date" json:"entered_date,omitempty"`
	AccountID             *uuid.UUID `db:"account_id" json:"account_id,omitempty"`
	Note                  *string    `db:"note" json:"note,omitempty"`
	VersionID             int        `db:"version_id" json:"version_id"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

func (ci *ChargeItem) GetVersionID() int  { return ci.VersionID }
func (ci *ChargeItem) SetVersionID(v int) { ci.VersionID = v }

func (ci *ChargeItem) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ChargeItem",
		"id":           ci.FHIRID,
		"status":       ci.Status,
		"subject":      fhir.Reference{Reference: fhir.FormatReference("Patient", ci.SubjectPatientID.String())},
		"meta":         fhir.Meta{LastUpdated: ci.UpdatedAt},
	}
	if ci.CodeCode != nil {
		coding := fhir.Coding{Code: *ci.CodeCode}
		if ci.CodeDisplay != nil {
			coding.Display = *ci.CodeDisplay
		}
		if ci.CodeSystem != nil {
			coding.System = *ci.CodeSystem
		}
		result["code"] = fhir.CodeableConcept{Coding: []fhir.Coding{coding}}
	}
	if ci.ContextEncounterID != nil {
		result["context"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", ci.ContextEncounterID.String())}
	}
	if ci.OccurrenceDate != nil {
		result["occurrenceDateTime"] = ci.OccurrenceDate.Format("2006-01-02T15:04:05Z")
	}
	if ci.PerformerID != nil {
		result["performer"] = []map[string]interface{}{
			{"actor": fhir.Reference{Reference: fhir.FormatReference("Practitioner", ci.PerformerID.String())}},
		}
	}
	if ci.PerformingOrgID != nil {
		result["performingOrganization"] = fhir.Reference{Reference: fhir.FormatReference("Organization", ci.PerformingOrgID.String())}
	}
	if ci.QuantityValue != nil {
		result["quantity"] = map[string]interface{}{"value": *ci.QuantityValue}
	}
	if ci.FactorOverride != nil {
		result["factorOverride"] = *ci.FactorOverride
	}
	if ci.PriceOverrideValue != nil {
		po := map[string]interface{}{"value": *ci.PriceOverrideValue}
		if ci.PriceOverrideCurrency != nil {
			po["currency"] = *ci.PriceOverrideCurrency
		}
		result["priceOverride"] = po
	}
	if ci.OverrideReason != nil {
		result["overrideReason"] = *ci.OverrideReason
	}
	if ci.EntererID != nil {
		result["enterer"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", ci.EntererID.String())}
	}
	if ci.EnteredDate != nil {
		result["enteredDate"] = ci.EnteredDate.Format("2006-01-02T15:04:05Z")
	}
	if ci.AccountID != nil {
		result["account"] = []fhir.Reference{{Reference: fhir.FormatReference("Account", ci.AccountID.String())}}
	}
	if ci.Note != nil {
		result["note"] = []map[string]string{{"text": *ci.Note}}
	}
	return result
}

// ChargeItemDefinition maps to the charge_item_definition table (FHIR ChargeItemDefinition resource).
type ChargeItemDefinition struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	FHIRID         string     `db:"fhir_id" json:"fhir_id"`
	URL            *string    `db:"url" json:"url,omitempty"`
	Status         string     `db:"status" json:"status"`
	Title          *string    `db:"title" json:"title,omitempty"`
	Description    *string    `db:"description" json:"description,omitempty"`
	CodeCode       *string    `db:"code_code" json:"code_code,omitempty"`
	CodeDisplay    *string    `db:"code_display" json:"code_display,omitempty"`
	CodeSystem     *string    `db:"code_system" json:"code_system,omitempty"`
	EffectiveStart *time.Time `db:"effective_start" json:"effective_start,omitempty"`
	EffectiveEnd   *time.Time `db:"effective_end" json:"effective_end,omitempty"`
	Publisher      *string    `db:"publisher" json:"publisher,omitempty"`
	ApprovalDate   *time.Time `db:"approval_date" json:"approval_date,omitempty"`
	LastReviewDate *time.Time `db:"last_review_date" json:"last_review_date,omitempty"`
	VersionID      int        `db:"version_id" json:"version_id"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}

func (cd *ChargeItemDefinition) GetVersionID() int  { return cd.VersionID }
func (cd *ChargeItemDefinition) SetVersionID(v int) { cd.VersionID = v }

func (cd *ChargeItemDefinition) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ChargeItemDefinition",
		"id":           cd.FHIRID,
		"status":       cd.Status,
		"meta":         fhir.Meta{LastUpdated: cd.UpdatedAt},
	}
	if cd.URL != nil {
		result["url"] = *cd.URL
	}
	if cd.Title != nil {
		result["title"] = *cd.Title
	}
	if cd.Description != nil {
		result["description"] = *cd.Description
	}
	if cd.CodeCode != nil {
		coding := fhir.Coding{Code: *cd.CodeCode}
		if cd.CodeDisplay != nil {
			coding.Display = *cd.CodeDisplay
		}
		if cd.CodeSystem != nil {
			coding.System = *cd.CodeSystem
		}
		result["code"] = fhir.CodeableConcept{Coding: []fhir.Coding{coding}}
	}
	if cd.EffectiveStart != nil || cd.EffectiveEnd != nil {
		result["effectivePeriod"] = fhir.Period{Start: cd.EffectiveStart, End: cd.EffectiveEnd}
	}
	if cd.Publisher != nil {
		result["publisher"] = *cd.Publisher
	}
	if cd.ApprovalDate != nil {
		result["approvalDate"] = cd.ApprovalDate.Format("2006-01-02")
	}
	if cd.LastReviewDate != nil {
		result["lastReviewDate"] = cd.LastReviewDate.Format("2006-01-02")
	}
	return result
}

// Contract maps to the contract table (FHIR Contract resource).
type Contract struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	FHIRID           string     `db:"fhir_id" json:"fhir_id"`
	Status           string     `db:"status" json:"status"`
	TypeCode         *string    `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay      *string    `db:"type_display" json:"type_display,omitempty"`
	SubTypeCode      *string    `db:"sub_type_code" json:"sub_type_code,omitempty"`
	Title            *string    `db:"title" json:"title,omitempty"`
	Issued           *time.Time `db:"issued" json:"issued,omitempty"`
	AppliesStart     *time.Time `db:"applies_start" json:"applies_start,omitempty"`
	AppliesEnd       *time.Time `db:"applies_end" json:"applies_end,omitempty"`
	SubjectPatientID *uuid.UUID `db:"subject_patient_id" json:"subject_patient_id,omitempty"`
	AuthorityOrgID   *uuid.UUID `db:"authority_org_id" json:"authority_org_id,omitempty"`
	ScopeCode        *string    `db:"scope_code" json:"scope_code,omitempty"`
	ScopeDisplay     *string    `db:"scope_display" json:"scope_display,omitempty"`
	VersionID        int        `db:"version_id" json:"version_id"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

func (ct *Contract) GetVersionID() int  { return ct.VersionID }
func (ct *Contract) SetVersionID(v int) { ct.VersionID = v }

func (ct *Contract) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Contract",
		"id":           ct.FHIRID,
		"status":       ct.Status,
		"meta":         fhir.Meta{LastUpdated: ct.UpdatedAt},
	}
	if ct.TypeCode != nil {
		cc := fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *ct.TypeCode}},
		}
		if ct.TypeDisplay != nil {
			cc.Coding[0].Display = *ct.TypeDisplay
		}
		result["type"] = cc
	}
	if ct.SubTypeCode != nil {
		result["subType"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *ct.SubTypeCode}},
		}}
	}
	if ct.Title != nil {
		result["title"] = *ct.Title
	}
	if ct.Issued != nil {
		result["issued"] = ct.Issued.Format("2006-01-02T15:04:05Z")
	}
	if ct.AppliesStart != nil || ct.AppliesEnd != nil {
		result["applies"] = fhir.Period{Start: ct.AppliesStart, End: ct.AppliesEnd}
	}
	if ct.SubjectPatientID != nil {
		result["subject"] = []fhir.Reference{{Reference: fhir.FormatReference("Patient", ct.SubjectPatientID.String())}}
	}
	if ct.AuthorityOrgID != nil {
		result["authority"] = []fhir.Reference{{Reference: fhir.FormatReference("Organization", ct.AuthorityOrgID.String())}}
	}
	if ct.ScopeCode != nil {
		cc := fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *ct.ScopeCode}},
		}
		if ct.ScopeDisplay != nil {
			cc.Coding[0].Display = *ct.ScopeDisplay
		}
		result["scope"] = cc
	}
	return result
}

// EnrollmentRequest maps to the enrollment_request table (FHIR EnrollmentRequest resource).
type EnrollmentRequest struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	FHIRID             string     `db:"fhir_id" json:"fhir_id"`
	Status             string     `db:"status" json:"status"`
	Created            time.Time  `db:"created" json:"created"`
	InsurerOrgID       *uuid.UUID `db:"insurer_org_id" json:"insurer_org_id,omitempty"`
	ProviderID         *uuid.UUID `db:"provider_id" json:"provider_id,omitempty"`
	CandidatePatientID *uuid.UUID `db:"candidate_patient_id" json:"candidate_patient_id,omitempty"`
	CoverageID         *uuid.UUID `db:"coverage_id" json:"coverage_id,omitempty"`
	VersionID          int        `db:"version_id" json:"version_id"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

func (er *EnrollmentRequest) GetVersionID() int  { return er.VersionID }
func (er *EnrollmentRequest) SetVersionID(v int) { er.VersionID = v }

func (er *EnrollmentRequest) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "EnrollmentRequest",
		"id":           er.FHIRID,
		"status":       er.Status,
		"created":      er.Created.Format("2006-01-02T15:04:05Z"),
		"meta":         fhir.Meta{LastUpdated: er.UpdatedAt},
	}
	if er.InsurerOrgID != nil {
		result["insurer"] = fhir.Reference{Reference: fhir.FormatReference("Organization", er.InsurerOrgID.String())}
	}
	if er.ProviderID != nil {
		result["provider"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", er.ProviderID.String())}
	}
	if er.CandidatePatientID != nil {
		result["candidate"] = fhir.Reference{Reference: fhir.FormatReference("Patient", er.CandidatePatientID.String())}
	}
	if er.CoverageID != nil {
		result["coverage"] = fhir.Reference{Reference: fhir.FormatReference("Coverage", er.CoverageID.String())}
	}
	return result
}

// EnrollmentResponse maps to the enrollment_response table (FHIR EnrollmentResponse resource).
type EnrollmentResponse struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	FHIRID         string     `db:"fhir_id" json:"fhir_id"`
	Status         string     `db:"status" json:"status"`
	RequestID      *uuid.UUID `db:"request_id" json:"request_id,omitempty"`
	Outcome        *string    `db:"outcome" json:"outcome,omitempty"`
	Disposition    *string    `db:"disposition" json:"disposition,omitempty"`
	Created        time.Time  `db:"created" json:"created"`
	OrganizationID *uuid.UUID `db:"organization_id" json:"organization_id,omitempty"`
	VersionID      int        `db:"version_id" json:"version_id"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}

func (er *EnrollmentResponse) GetVersionID() int  { return er.VersionID }
func (er *EnrollmentResponse) SetVersionID(v int) { er.VersionID = v }

func (er *EnrollmentResponse) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "EnrollmentResponse",
		"id":           er.FHIRID,
		"status":       er.Status,
		"created":      er.Created.Format("2006-01-02T15:04:05Z"),
		"meta":         fhir.Meta{LastUpdated: er.UpdatedAt},
	}
	if er.RequestID != nil {
		result["request"] = fhir.Reference{Reference: fhir.FormatReference("EnrollmentRequest", er.RequestID.String())}
	}
	if er.Outcome != nil {
		result["outcome"] = *er.Outcome
	}
	if er.Disposition != nil {
		result["disposition"] = *er.Disposition
	}
	if er.OrganizationID != nil {
		result["organization"] = fhir.Reference{Reference: fhir.FormatReference("Organization", er.OrganizationID.String())}
	}
	return result
}
