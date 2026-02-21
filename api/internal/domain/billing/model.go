package billing

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Coverage maps to the coverage table (FHIR Coverage resource).
type Coverage struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	FHIRID           string     `db:"fhir_id" json:"fhir_id"`
	Status           string     `db:"status" json:"status"`
	TypeCode         *string    `db:"type_code" json:"type_code,omitempty"`
	PatientID        uuid.UUID  `db:"patient_id" json:"patient_id"`
	SubscriberID     *string    `db:"subscriber_id" json:"subscriber_id,omitempty"`
	SubscriberName   *string    `db:"subscriber_name" json:"subscriber_name,omitempty"`
	SubscriberDOB    *time.Time `db:"subscriber_dob" json:"subscriber_dob,omitempty"`
	Relationship     *string    `db:"relationship" json:"relationship,omitempty"`
	DependentNumber  *string    `db:"dependent_number" json:"dependent_number,omitempty"`
	PayorOrgID       *uuid.UUID `db:"payor_org_id" json:"payor_org_id,omitempty"`
	PayorName        *string    `db:"payor_name" json:"payor_name,omitempty"`
	PolicyNumber     *string    `db:"policy_number" json:"policy_number,omitempty"`
	GroupNumber      *string    `db:"group_number" json:"group_number,omitempty"`
	GroupName        *string    `db:"group_name" json:"group_name,omitempty"`
	PlanName         *string    `db:"plan_name" json:"plan_name,omitempty"`
	PlanType         *string    `db:"plan_type" json:"plan_type,omitempty"`
	MemberID         *string    `db:"member_id" json:"member_id,omitempty"`
	BINNumber        *string    `db:"bin_number" json:"bin_number,omitempty"`
	PCNNumber        *string    `db:"pcn_number" json:"pcn_number,omitempty"`
	RxGroup          *string    `db:"rx_group" json:"rx_group,omitempty"`
	PlanTypeUS       *string    `db:"plan_type_us" json:"plan_type_us,omitempty"`
	ABPMJAYId        *string    `db:"ab_pmjay_id" json:"ab_pmjay_id,omitempty"`
	ABPMJAYFamilyID  *string    `db:"ab_pmjay_family_id" json:"ab_pmjay_family_id,omitempty"`
	StateSchemeID    *string    `db:"state_scheme_id" json:"state_scheme_id,omitempty"`
	StateSchemeName  *string    `db:"state_scheme_name" json:"state_scheme_name,omitempty"`
	ESISNumber       *string    `db:"esis_number" json:"esis_number,omitempty"`
	CGHSBenefID      *string    `db:"cghs_beneficiary_id" json:"cghs_beneficiary_id,omitempty"`
	ECHSCardNumber   *string    `db:"echs_card_number" json:"echs_card_number,omitempty"`
	PeriodStart      *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd        *time.Time `db:"period_end" json:"period_end,omitempty"`
	Network          *string    `db:"network" json:"network,omitempty"`
	CopayAmount      *float64   `db:"copay_amount" json:"copay_amount,omitempty"`
	CopayPercentage  *float64   `db:"copay_percentage" json:"copay_percentage,omitempty"`
	DeductibleAmount *float64   `db:"deductible_amount" json:"deductible_amount,omitempty"`
	DeductibleMet    *float64   `db:"deductible_met" json:"deductible_met,omitempty"`
	MaxBenefitAmount *float64   `db:"max_benefit_amount" json:"max_benefit_amount,omitempty"`
	OutOfPocketMax   *float64   `db:"out_of_pocket_max" json:"out_of_pocket_max,omitempty"`
	Currency         *string    `db:"currency" json:"currency,omitempty"`
	CoverageOrder    *int       `db:"coverage_order" json:"coverage_order,omitempty"`
	Note             *string    `db:"note" json:"note,omitempty"`
	VersionID        int        `db:"version_id" json:"version_id"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (c *Coverage) GetVersionID() int { return c.VersionID }

// SetVersionID sets the current version.
func (c *Coverage) SetVersionID(v int) { c.VersionID = v }

func (c *Coverage) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Coverage",
		"id":           c.FHIRID,
		"status":       c.Status,
		"beneficiary":  fhir.Reference{Reference: fhir.FormatReference("Patient", c.PatientID.String())},
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", c.VersionID),
			LastUpdated: c.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Coverage"},
		},
	}
	if c.TypeCode != nil {
		result["type"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System: "http://terminology.hl7.org/CodeSystem/v3-ActCode",
				Code:   *c.TypeCode,
			}},
		}
	}
	if c.SubscriberID != nil {
		result["subscriberId"] = *c.SubscriberID
	}
	if c.Relationship != nil {
		result["relationship"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System: "http://terminology.hl7.org/CodeSystem/subscriber-relationship",
				Code:   *c.Relationship,
			}},
		}
	}
	if c.PayorOrgID != nil {
		result["payor"] = []fhir.Reference{{Reference: fhir.FormatReference("Organization", c.PayorOrgID.String())}}
	} else if c.PayorName != nil {
		result["payor"] = []fhir.Reference{{Display: *c.PayorName}}
	}
	if c.PeriodStart != nil || c.PeriodEnd != nil {
		result["period"] = fhir.Period{Start: c.PeriodStart, End: c.PeriodEnd}
	}
	if c.CoverageOrder != nil {
		result["order"] = *c.CoverageOrder
	}
	if c.PolicyNumber != nil {
		result["identifier"] = []fhir.Identifier{{Value: *c.PolicyNumber}}
	}
	return result
}

// Claim maps to the claim table (FHIR Claim resource).
type Claim struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	FHIRID                string     `db:"fhir_id" json:"fhir_id"`
	Status                string     `db:"status" json:"status"`
	TypeCode              *string    `db:"type_code" json:"type_code,omitempty"`
	SubTypeCode           *string    `db:"sub_type_code" json:"sub_type_code,omitempty"`
	UseCode               *string    `db:"use_code" json:"use_code,omitempty"`
	PatientID             uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID           *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	InsurerOrgID          *uuid.UUID `db:"insurer_org_id" json:"insurer_org_id,omitempty"`
	ProviderID            *uuid.UUID `db:"provider_id" json:"provider_id,omitempty"`
	ProviderOrgID         *uuid.UUID `db:"provider_org_id" json:"provider_org_id,omitempty"`
	CoverageID            *uuid.UUID `db:"coverage_id" json:"coverage_id,omitempty"`
	PriorityCode          *string    `db:"priority_code" json:"priority_code,omitempty"`
	PrescriptionID        *uuid.UUID `db:"prescription_id" json:"prescription_id,omitempty"`
	ReferralID            *uuid.UUID `db:"referral_id" json:"referral_id,omitempty"`
	FacilityID            *uuid.UUID `db:"facility_id" json:"facility_id,omitempty"`
	BillablePeriodStart   *time.Time `db:"billable_period_start" json:"billable_period_start,omitempty"`
	BillablePeriodEnd     *time.Time `db:"billable_period_end" json:"billable_period_end,omitempty"`
	CreatedDate           *time.Time `db:"created_date" json:"created_date,omitempty"`
	TotalAmount           *float64   `db:"total_amount" json:"total_amount,omitempty"`
	Currency              *string    `db:"currency" json:"currency,omitempty"`
	PlaceOfService        *string    `db:"place_of_service" json:"place_of_service,omitempty"`
	ABPMJAYClaimID        *string    `db:"ab_pmjay_claim_id" json:"ab_pmjay_claim_id,omitempty"`
	ABPMJAYPackageCode    *string    `db:"ab_pmjay_package_code" json:"ab_pmjay_package_code,omitempty"`
	ROHINIClaimID         *string    `db:"rohini_claim_id" json:"rohini_claim_id,omitempty"`
	RelatedClaimID        *uuid.UUID `db:"related_claim_id" json:"related_claim_id,omitempty"`
	RelatedClaimRelation  *string    `db:"related_claim_relationship" json:"related_claim_relationship,omitempty"`
	VersionID             int        `db:"version_id" json:"version_id"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (cl *Claim) GetVersionID() int { return cl.VersionID }

// SetVersionID sets the current version.
func (cl *Claim) SetVersionID(v int) { cl.VersionID = v }

func (cl *Claim) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Claim",
		"id":           cl.FHIRID,
		"status":       cl.Status,
		"patient":      fhir.Reference{Reference: fhir.FormatReference("Patient", cl.PatientID.String())},
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", cl.VersionID),
			LastUpdated: cl.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Claim"},
		},
	}
	if cl.TypeCode != nil {
		result["type"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System: "http://terminology.hl7.org/CodeSystem/claim-type",
				Code:   *cl.TypeCode,
			}},
		}
	}
	if cl.UseCode != nil {
		result["use"] = *cl.UseCode
	}
	if cl.ProviderID != nil {
		result["provider"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", cl.ProviderID.String())}
	} else if cl.ProviderOrgID != nil {
		result["provider"] = fhir.Reference{Reference: fhir.FormatReference("Organization", cl.ProviderOrgID.String())}
	}
	if cl.InsurerOrgID != nil {
		result["insurer"] = fhir.Reference{Reference: fhir.FormatReference("Organization", cl.InsurerOrgID.String())}
	}
	if cl.CoverageID != nil {
		result["insurance"] = []map[string]interface{}{{
			"sequence": 1,
			"focal":    true,
			"coverage": fhir.Reference{Reference: fhir.FormatReference("Coverage", cl.CoverageID.String())},
		}}
	}
	if cl.PriorityCode != nil {
		result["priority"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *cl.PriorityCode}},
		}
	}
	if cl.BillablePeriodStart != nil || cl.BillablePeriodEnd != nil {
		result["billablePeriod"] = fhir.Period{Start: cl.BillablePeriodStart, End: cl.BillablePeriodEnd}
	}
	if cl.TotalAmount != nil {
		cur := "USD"
		if cl.Currency != nil {
			cur = *cl.Currency
		}
		result["total"] = map[string]interface{}{
			"value":    *cl.TotalAmount,
			"currency": cur,
		}
	}
	if cl.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", cl.EncounterID.String())}
	}
	if cl.FacilityID != nil {
		result["facility"] = fhir.Reference{Reference: fhir.FormatReference("Location", cl.FacilityID.String())}
	}
	return result
}

// ClaimDiagnosis maps to the claim_diagnosis junction table.
type ClaimDiagnosis struct {
	ID                   uuid.UUID `db:"id" json:"id"`
	ClaimID              uuid.UUID `db:"claim_id" json:"claim_id"`
	Sequence             int       `db:"sequence" json:"sequence"`
	DiagnosisCodeSystem  *string   `db:"diagnosis_code_system" json:"diagnosis_code_system,omitempty"`
	DiagnosisCode        string    `db:"diagnosis_code" json:"diagnosis_code"`
	DiagnosisDisplay     *string   `db:"diagnosis_display" json:"diagnosis_display,omitempty"`
	TypeCode             *string   `db:"type_code" json:"type_code,omitempty"`
	OnAdmission          *bool     `db:"on_admission" json:"on_admission,omitempty"`
	PackageCode          *string   `db:"package_code" json:"package_code,omitempty"`
}

// ClaimProcedure maps to the claim_procedure junction table.
type ClaimProcedure struct {
	ID                   uuid.UUID  `db:"id" json:"id"`
	ClaimID              uuid.UUID  `db:"claim_id" json:"claim_id"`
	Sequence             int        `db:"sequence" json:"sequence"`
	TypeCode             *string    `db:"type_code" json:"type_code,omitempty"`
	Date                 *time.Time `db:"date" json:"date,omitempty"`
	ProcedureCodeSystem  *string    `db:"procedure_code_system" json:"procedure_code_system,omitempty"`
	ProcedureCode        string     `db:"procedure_code" json:"procedure_code"`
	ProcedureDisplay     *string    `db:"procedure_display" json:"procedure_display,omitempty"`
	UDI                  *string    `db:"udi" json:"udi,omitempty"`
}

// ClaimItem maps to the claim_item table (line items).
type ClaimItem struct {
	ID                       uuid.UUID  `db:"id" json:"id"`
	ClaimID                  uuid.UUID  `db:"claim_id" json:"claim_id"`
	Sequence                 int        `db:"sequence" json:"sequence"`
	ProductOrServiceSystem   *string    `db:"product_or_service_system" json:"product_or_service_system,omitempty"`
	ProductOrServiceCode     string     `db:"product_or_service_code" json:"product_or_service_code"`
	ProductOrServiceDisplay  *string    `db:"product_or_service_display" json:"product_or_service_display,omitempty"`
	ServicedDate             *time.Time `db:"serviced_date" json:"serviced_date,omitempty"`
	ServicedPeriodStart      *time.Time `db:"serviced_period_start" json:"serviced_period_start,omitempty"`
	ServicedPeriodEnd        *time.Time `db:"serviced_period_end" json:"serviced_period_end,omitempty"`
	LocationCode             *string    `db:"location_code" json:"location_code,omitempty"`
	QuantityValue            *float64   `db:"quantity_value" json:"quantity_value,omitempty"`
	QuantityUnit             *string    `db:"quantity_unit" json:"quantity_unit,omitempty"`
	UnitPrice                *float64   `db:"unit_price" json:"unit_price,omitempty"`
	Factor                   *float64   `db:"factor" json:"factor,omitempty"`
	NetAmount                *float64   `db:"net_amount" json:"net_amount,omitempty"`
	Currency                 *string    `db:"currency" json:"currency,omitempty"`
	RevenueCode              *string    `db:"revenue_code" json:"revenue_code,omitempty"`
	RevenueDisplay           *string    `db:"revenue_display" json:"revenue_display,omitempty"`
	BodySiteCode             *string    `db:"body_site_code" json:"body_site_code,omitempty"`
	SubSiteCode              *string    `db:"sub_site_code" json:"sub_site_code,omitempty"`
	EncounterID              *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	Note                     *string    `db:"note" json:"note,omitempty"`
}

// ClaimResponse maps to the claim_response table (FHIR ClaimResponse resource).
type ClaimResponse struct {
	ID                       uuid.UUID  `db:"id" json:"id"`
	FHIRID                   string     `db:"fhir_id" json:"fhir_id"`
	ClaimID                  uuid.UUID  `db:"claim_id" json:"claim_id"`
	Status                   string     `db:"status" json:"status"`
	TypeCode                 *string    `db:"type_code" json:"type_code,omitempty"`
	UseCode                  *string    `db:"use_code" json:"use_code,omitempty"`
	Outcome                  *string    `db:"outcome" json:"outcome,omitempty"`
	Disposition              *string    `db:"disposition" json:"disposition,omitempty"`
	PreAuthRef               *string    `db:"pre_auth_ref" json:"pre_auth_ref,omitempty"`
	PaymentTypeCode          *string    `db:"payment_type_code" json:"payment_type_code,omitempty"`
	PaymentAdjustment        *float64   `db:"payment_adjustment" json:"payment_adjustment,omitempty"`
	PaymentAdjustmentReason  *string    `db:"payment_adjustment_reason" json:"payment_adjustment_reason,omitempty"`
	PaymentAmount            *float64   `db:"payment_amount" json:"payment_amount,omitempty"`
	PaymentDate              *time.Time `db:"payment_date" json:"payment_date,omitempty"`
	PaymentIdentifier        *string    `db:"payment_identifier" json:"payment_identifier,omitempty"`
	TotalAmount              *float64   `db:"total_amount" json:"total_amount,omitempty"`
	ProcessNote              *string    `db:"process_note" json:"process_note,omitempty"`
	CommunicationRequest     *string    `db:"communication_request" json:"communication_request,omitempty"`
	VersionID                int        `db:"version_id" json:"version_id"`
	CreatedAt                time.Time  `db:"created_at" json:"created_at"`
}

// GetVersionID returns the current version.
func (cr *ClaimResponse) GetVersionID() int { return cr.VersionID }

// SetVersionID sets the current version.
func (cr *ClaimResponse) SetVersionID(v int) { cr.VersionID = v }

func (cr *ClaimResponse) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ClaimResponse",
		"id":           cr.FHIRID,
		"status":       cr.Status,
		"request":      fhir.Reference{Reference: fhir.FormatReference("Claim", cr.ClaimID.String())},
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", cr.VersionID),
			LastUpdated: cr.CreatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/ClaimResponse"},
		},
	}
	if cr.TypeCode != nil {
		result["type"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System: "http://terminology.hl7.org/CodeSystem/claim-type",
				Code:   *cr.TypeCode,
			}},
		}
	}
	if cr.UseCode != nil {
		result["use"] = *cr.UseCode
	}
	if cr.Outcome != nil {
		result["outcome"] = *cr.Outcome
	}
	if cr.Disposition != nil {
		result["disposition"] = *cr.Disposition
	}
	if cr.PreAuthRef != nil {
		result["preAuthRef"] = []string{*cr.PreAuthRef}
	}
	if cr.PaymentAmount != nil {
		payment := map[string]interface{}{
			"amount": map[string]interface{}{"value": *cr.PaymentAmount},
		}
		if cr.PaymentDate != nil {
			payment["date"] = cr.PaymentDate.Format("2006-01-02")
		}
		if cr.PaymentTypeCode != nil {
			payment["type"] = fhir.CodeableConcept{
				Coding: []fhir.Coding{{Code: *cr.PaymentTypeCode}},
			}
		}
		result["payment"] = payment
	}
	if cr.TotalAmount != nil {
		result["total"] = []map[string]interface{}{
			{
				"category": fhir.CodeableConcept{Coding: []fhir.Coding{{Code: "submitted"}}},
				"amount":   map[string]interface{}{"value": *cr.TotalAmount},
			},
		}
	}
	if cr.ProcessNote != nil {
		result["processNote"] = []map[string]string{{"text": *cr.ProcessNote}}
	}
	return result
}

// ClaimResponseItem maps to the claim_response_item table.
type ClaimResponseItem struct {
	ID                    uuid.UUID `db:"id" json:"id"`
	ClaimResponseID       uuid.UUID `db:"claim_response_id" json:"claim_response_id"`
	ItemSequence          int       `db:"item_sequence" json:"item_sequence"`
	AdjudicationCategory  *string   `db:"adjudication_category" json:"adjudication_category,omitempty"`
	AdjudicationAmount    *float64  `db:"adjudication_amount" json:"adjudication_amount,omitempty"`
	AdjudicationValue     *float64  `db:"adjudication_value" json:"adjudication_value,omitempty"`
	AdjudicationReason    *string   `db:"adjudication_reason" json:"adjudication_reason,omitempty"`
	Note                  *string   `db:"note" json:"note,omitempty"`
}

// ExplanationOfBenefit maps to the explanation_of_benefit table (FHIR ExplanationOfBenefit resource).
type ExplanationOfBenefit struct {
	ID                         uuid.UUID  `db:"id" json:"id"`
	FHIRID                     string     `db:"fhir_id" json:"fhir_id"`
	Status                     string     `db:"status" json:"status"`
	TypeCode                   *string    `db:"type_code" json:"type_code,omitempty"`
	UseCode                    *string    `db:"use_code" json:"use_code,omitempty"`
	PatientID                  uuid.UUID  `db:"patient_id" json:"patient_id"`
	ClaimID                    *uuid.UUID `db:"claim_id" json:"claim_id,omitempty"`
	ClaimResponseID            *uuid.UUID `db:"claim_response_id" json:"claim_response_id,omitempty"`
	CoverageID                 *uuid.UUID `db:"coverage_id" json:"coverage_id,omitempty"`
	InsurerOrgID               *uuid.UUID `db:"insurer_org_id" json:"insurer_org_id,omitempty"`
	ProviderID                 *uuid.UUID `db:"provider_id" json:"provider_id,omitempty"`
	Outcome                    *string    `db:"outcome" json:"outcome,omitempty"`
	Disposition                *string    `db:"disposition" json:"disposition,omitempty"`
	BillablePeriodStart        *time.Time `db:"billable_period_start" json:"billable_period_start,omitempty"`
	BillablePeriodEnd          *time.Time `db:"billable_period_end" json:"billable_period_end,omitempty"`
	TotalSubmitted             *float64   `db:"total_submitted" json:"total_submitted,omitempty"`
	TotalBenefit               *float64   `db:"total_benefit" json:"total_benefit,omitempty"`
	TotalPatientResponsibility *float64   `db:"total_patient_responsibility" json:"total_patient_responsibility,omitempty"`
	TotalPayment               *float64   `db:"total_payment" json:"total_payment,omitempty"`
	PaymentDate                *time.Time `db:"payment_date" json:"payment_date,omitempty"`
	Currency                   *string    `db:"currency" json:"currency,omitempty"`
	VersionID                  int        `db:"version_id" json:"version_id"`
	CreatedAt                  time.Time  `db:"created_at" json:"created_at"`
}

// GetVersionID returns the current version.
func (eob *ExplanationOfBenefit) GetVersionID() int { return eob.VersionID }

// SetVersionID sets the current version.
func (eob *ExplanationOfBenefit) SetVersionID(v int) { eob.VersionID = v }

func (eob *ExplanationOfBenefit) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ExplanationOfBenefit",
		"id":           eob.FHIRID,
		"status":       eob.Status,
		"patient":      fhir.Reference{Reference: fhir.FormatReference("Patient", eob.PatientID.String())},
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", eob.VersionID),
			LastUpdated: eob.CreatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/ExplanationOfBenefit"},
		},
	}
	if eob.TypeCode != nil {
		result["type"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System: "http://terminology.hl7.org/CodeSystem/claim-type",
				Code:   *eob.TypeCode,
			}},
		}
	}
	if eob.UseCode != nil {
		result["use"] = *eob.UseCode
	}
	if eob.Outcome != nil {
		result["outcome"] = *eob.Outcome
	}
	if eob.Disposition != nil {
		result["disposition"] = *eob.Disposition
	}
	if eob.InsurerOrgID != nil {
		result["insurer"] = fhir.Reference{Reference: fhir.FormatReference("Organization", eob.InsurerOrgID.String())}
	}
	if eob.ProviderID != nil {
		result["provider"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", eob.ProviderID.String())}
	}
	if eob.ClaimID != nil {
		result["claim"] = fhir.Reference{Reference: fhir.FormatReference("Claim", eob.ClaimID.String())}
	}
	if eob.ClaimResponseID != nil {
		result["claimResponse"] = fhir.Reference{Reference: fhir.FormatReference("ClaimResponse", eob.ClaimResponseID.String())}
	}
	if eob.CoverageID != nil {
		result["insurance"] = []map[string]interface{}{{
			"focal":    true,
			"coverage": fhir.Reference{Reference: fhir.FormatReference("Coverage", eob.CoverageID.String())},
		}}
	}
	if eob.BillablePeriodStart != nil || eob.BillablePeriodEnd != nil {
		result["billablePeriod"] = fhir.Period{Start: eob.BillablePeriodStart, End: eob.BillablePeriodEnd}
	}
	var totals []map[string]interface{}
	if eob.TotalSubmitted != nil {
		totals = append(totals, map[string]interface{}{
			"category": fhir.CodeableConcept{Coding: []fhir.Coding{{Code: "submitted"}}},
			"amount":   map[string]interface{}{"value": *eob.TotalSubmitted, "currency": strVal(eob.Currency)},
		})
	}
	if eob.TotalBenefit != nil {
		totals = append(totals, map[string]interface{}{
			"category": fhir.CodeableConcept{Coding: []fhir.Coding{{Code: "benefit"}}},
			"amount":   map[string]interface{}{"value": *eob.TotalBenefit, "currency": strVal(eob.Currency)},
		})
	}
	if len(totals) > 0 {
		result["total"] = totals
	}
	if eob.TotalPayment != nil {
		payment := map[string]interface{}{
			"amount": map[string]interface{}{"value": *eob.TotalPayment, "currency": strVal(eob.Currency)},
		}
		if eob.PaymentDate != nil {
			payment["date"] = eob.PaymentDate.Format("2006-01-02")
		}
		result["payment"] = payment
	}
	return result
}

// Invoice maps to the invoice table (FHIR Invoice resource).
type Invoice struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	FHIRID         string     `db:"fhir_id" json:"fhir_id"`
	Status         string     `db:"status" json:"status"`
	TypeCode       *string    `db:"type_code" json:"type_code,omitempty"`
	PatientID      uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID    *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	IssuerOrgID    *uuid.UUID `db:"issuer_org_id" json:"issuer_org_id,omitempty"`
	Date           *time.Time `db:"date" json:"date,omitempty"`
	ParticipantID  *uuid.UUID `db:"participant_id" json:"participant_id,omitempty"`
	TotalNet       *float64   `db:"total_net" json:"total_net,omitempty"`
	TotalGross     *float64   `db:"total_gross" json:"total_gross,omitempty"`
	TotalTax       *float64   `db:"total_tax" json:"total_tax,omitempty"`
	Currency       *string    `db:"currency" json:"currency,omitempty"`
	PaymentTerms   *string    `db:"payment_terms" json:"payment_terms,omitempty"`
	GSTIN          *string    `db:"gstin" json:"gstin,omitempty"`
	GSTAmount      *float64   `db:"gst_amount" json:"gst_amount,omitempty"`
	SACCode        *string    `db:"sac_code" json:"sac_code,omitempty"`
	Note           *string    `db:"note" json:"note,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
}

// InvoiceLineItem maps to the invoice_line_item table.
type InvoiceLineItem struct {
	ID             uuid.UUID `db:"id" json:"id"`
	InvoiceID      uuid.UUID `db:"invoice_id" json:"invoice_id"`
	Sequence       int       `db:"sequence" json:"sequence"`
	Description    *string   `db:"description" json:"description,omitempty"`
	ServiceCode    *string   `db:"service_code" json:"service_code,omitempty"`
	ServiceDisplay *string   `db:"service_display" json:"service_display,omitempty"`
	Quantity       *float64  `db:"quantity" json:"quantity,omitempty"`
	UnitPrice      *float64  `db:"unit_price" json:"unit_price,omitempty"`
	NetAmount      *float64  `db:"net_amount" json:"net_amount,omitempty"`
	TaxAmount      *float64  `db:"tax_amount" json:"tax_amount,omitempty"`
	GrossAmount    *float64  `db:"gross_amount" json:"gross_amount,omitempty"`
	Currency       *string   `db:"currency" json:"currency,omitempty"`
}

func (inv *Invoice) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Invoice",
		"id":           inv.FHIRID,
		"status":       inv.Status,
		"subject":      fhir.Reference{Reference: fhir.FormatReference("Patient", inv.PatientID.String())},
		"meta": fhir.Meta{
			LastUpdated: inv.CreatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Invoice"},
		},
	}
	if inv.TypeCode != nil {
		result["type"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *inv.TypeCode}},
		}
	}
	if inv.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", inv.EncounterID.String())}
	}
	if inv.IssuerOrgID != nil {
		result["issuer"] = fhir.Reference{Reference: fhir.FormatReference("Organization", inv.IssuerOrgID.String())}
	}
	if inv.Date != nil {
		result["date"] = inv.Date.Format("2006-01-02")
	}
	if inv.ParticipantID != nil {
		result["participant"] = []map[string]interface{}{{
			"actor": fhir.Reference{Reference: fhir.FormatReference("Practitioner", inv.ParticipantID.String())},
		}}
	}
	if inv.TotalNet != nil {
		cur := "USD"
		if inv.Currency != nil {
			cur = *inv.Currency
		}
		result["totalNet"] = map[string]interface{}{"value": *inv.TotalNet, "currency": cur}
	}
	if inv.TotalGross != nil {
		cur := "USD"
		if inv.Currency != nil {
			cur = *inv.Currency
		}
		result["totalGross"] = map[string]interface{}{"value": *inv.TotalGross, "currency": cur}
	}
	if inv.PaymentTerms != nil {
		result["paymentTerms"] = *inv.PaymentTerms
	}
	if inv.Note != nil {
		result["note"] = []map[string]string{{"text": *inv.Note}}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
