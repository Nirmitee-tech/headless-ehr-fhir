package billing

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	coverages      CoverageRepository
	claims         ClaimRepository
	claimResponses ClaimResponseRepository
	eobs           ExplanationOfBenefitRepository
	invoices       InvoiceRepository
	vt             *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(cov CoverageRepository, cl ClaimRepository, cr ClaimResponseRepository, eob ExplanationOfBenefitRepository, inv InvoiceRepository) *Service {
	return &Service{coverages: cov, claims: cl, claimResponses: cr, eobs: eob, invoices: inv}
}

// -- Coverage --

var validCoverageStatuses = map[string]bool{
	"active": true, "cancelled": true, "draft": true, "entered-in-error": true,
}

func (s *Service) CreateCoverage(ctx context.Context, c *Coverage) error {
	if c.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if c.PayorOrgID == nil && c.PayorName == nil {
		return fmt.Errorf("payor information is required (payor_org_id or payor_name)")
	}
	if c.Status == "" {
		c.Status = "active"
	}
	if !validCoverageStatuses[c.Status] {
		return fmt.Errorf("invalid coverage status: %s", c.Status)
	}
	if err := s.coverages.Create(ctx, c); err != nil {
		return err
	}
	c.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Coverage", c.FHIRID, c.ToFHIR())
	}
	return nil
}

func (s *Service) GetCoverage(ctx context.Context, id uuid.UUID) (*Coverage, error) {
	return s.coverages.GetByID(ctx, id)
}

func (s *Service) GetCoverageByFHIRID(ctx context.Context, fhirID string) (*Coverage, error) {
	return s.coverages.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateCoverage(ctx context.Context, c *Coverage) error {
	if c.Status != "" && !validCoverageStatuses[c.Status] {
		return fmt.Errorf("invalid coverage status: %s", c.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Coverage", c.FHIRID, c.VersionID, c.ToFHIR())
		if err == nil {
			c.VersionID = newVer
		}
	}
	return s.coverages.Update(ctx, c)
}

func (s *Service) DeleteCoverage(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		c, err := s.coverages.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Coverage", c.FHIRID, c.VersionID)
		}
	}
	return s.coverages.Delete(ctx, id)
}

func (s *Service) ListCoveragesByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Coverage, int, error) {
	return s.coverages.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchCoverages(ctx context.Context, params map[string]string, limit, offset int) ([]*Coverage, int, error) {
	return s.coverages.Search(ctx, params, limit, offset)
}

// -- Claim --

var validClaimStatuses = map[string]bool{
	"active": true, "cancelled": true, "draft": true, "entered-in-error": true,
}

func (s *Service) CreateClaim(ctx context.Context, c *Claim) error {
	if c.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if c.Status == "" {
		c.Status = "draft"
	}
	if !validClaimStatuses[c.Status] {
		return fmt.Errorf("invalid claim status: %s", c.Status)
	}
	if err := s.claims.Create(ctx, c); err != nil {
		return err
	}
	c.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Claim", c.FHIRID, c.ToFHIR())
	}
	return nil
}

func (s *Service) GetClaim(ctx context.Context, id uuid.UUID) (*Claim, error) {
	return s.claims.GetByID(ctx, id)
}

func (s *Service) GetClaimByFHIRID(ctx context.Context, fhirID string) (*Claim, error) {
	return s.claims.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateClaim(ctx context.Context, c *Claim) error {
	if c.Status != "" && !validClaimStatuses[c.Status] {
		return fmt.Errorf("invalid claim status: %s", c.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Claim", c.FHIRID, c.VersionID, c.ToFHIR())
		if err == nil {
			c.VersionID = newVer
		}
	}
	return s.claims.Update(ctx, c)
}

func (s *Service) DeleteClaim(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		c, err := s.claims.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Claim", c.FHIRID, c.VersionID)
		}
	}
	return s.claims.Delete(ctx, id)
}

func (s *Service) ListClaimsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Claim, int, error) {
	return s.claims.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchClaims(ctx context.Context, params map[string]string, limit, offset int) ([]*Claim, int, error) {
	return s.claims.Search(ctx, params, limit, offset)
}

func (s *Service) AddClaimDiagnosis(ctx context.Context, d *ClaimDiagnosis) error {
	if d.ClaimID == uuid.Nil {
		return fmt.Errorf("claim_id is required")
	}
	if d.DiagnosisCode == "" {
		return fmt.Errorf("diagnosis_code is required")
	}
	return s.claims.AddDiagnosis(ctx, d)
}

func (s *Service) GetClaimDiagnoses(ctx context.Context, claimID uuid.UUID) ([]*ClaimDiagnosis, error) {
	return s.claims.GetDiagnoses(ctx, claimID)
}

func (s *Service) AddClaimProcedure(ctx context.Context, p *ClaimProcedure) error {
	if p.ClaimID == uuid.Nil {
		return fmt.Errorf("claim_id is required")
	}
	if p.ProcedureCode == "" {
		return fmt.Errorf("procedure_code is required")
	}
	return s.claims.AddProcedure(ctx, p)
}

func (s *Service) GetClaimProcedures(ctx context.Context, claimID uuid.UUID) ([]*ClaimProcedure, error) {
	return s.claims.GetProcedures(ctx, claimID)
}

func (s *Service) AddClaimItem(ctx context.Context, item *ClaimItem) error {
	if item.ClaimID == uuid.Nil {
		return fmt.Errorf("claim_id is required")
	}
	if item.ProductOrServiceCode == "" {
		return fmt.Errorf("product_or_service_code is required")
	}
	return s.claims.AddItem(ctx, item)
}

func (s *Service) GetClaimItems(ctx context.Context, claimID uuid.UUID) ([]*ClaimItem, error) {
	return s.claims.GetItems(ctx, claimID)
}

// -- ClaimResponse --

func (s *Service) CreateClaimResponse(ctx context.Context, cr *ClaimResponse) error {
	if cr.ClaimID == uuid.Nil {
		return fmt.Errorf("claim_id is required")
	}
	if cr.Status == "" {
		cr.Status = "active"
	}
	if err := s.claimResponses.Create(ctx, cr); err != nil {
		return err
	}
	cr.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ClaimResponse", cr.FHIRID, cr.ToFHIR())
	}
	return nil
}

func (s *Service) GetClaimResponse(ctx context.Context, id uuid.UUID) (*ClaimResponse, error) {
	return s.claimResponses.GetByID(ctx, id)
}

func (s *Service) GetClaimResponseByFHIRID(ctx context.Context, fhirID string) (*ClaimResponse, error) {
	return s.claimResponses.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateClaimResponse(ctx context.Context, cr *ClaimResponse) error {
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ClaimResponse", cr.FHIRID, cr.VersionID, cr.ToFHIR())
		if err == nil {
			cr.VersionID = newVer
		}
	}
	return s.claimResponses.Update(ctx, cr)
}

func (s *Service) DeleteClaimResponse(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		cr, err := s.claimResponses.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ClaimResponse", cr.FHIRID, cr.VersionID)
		}
	}
	return s.claimResponses.Delete(ctx, id)
}

func (s *Service) ListClaimResponsesByClaim(ctx context.Context, claimID uuid.UUID, limit, offset int) ([]*ClaimResponse, int, error) {
	return s.claimResponses.ListByClaim(ctx, claimID, limit, offset)
}

func (s *Service) SearchClaimResponses(ctx context.Context, params map[string]string, limit, offset int) ([]*ClaimResponse, int, error) {
	return s.claimResponses.Search(ctx, params, limit, offset)
}

// -- ExplanationOfBenefit --

var validEOBStatuses = map[string]bool{
	"active": true, "cancelled": true, "draft": true, "entered-in-error": true,
}

func (s *Service) CreateExplanationOfBenefit(ctx context.Context, eob *ExplanationOfBenefit) error {
	if eob.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if eob.Status == "" {
		eob.Status = "active"
	}
	if !validEOBStatuses[eob.Status] {
		return fmt.Errorf("invalid explanation_of_benefit status: %s", eob.Status)
	}
	if err := s.eobs.Create(ctx, eob); err != nil {
		return err
	}
	eob.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ExplanationOfBenefit", eob.FHIRID, eob.ToFHIR())
	}
	return nil
}

func (s *Service) GetExplanationOfBenefit(ctx context.Context, id uuid.UUID) (*ExplanationOfBenefit, error) {
	return s.eobs.GetByID(ctx, id)
}

func (s *Service) GetExplanationOfBenefitByFHIRID(ctx context.Context, fhirID string) (*ExplanationOfBenefit, error) {
	return s.eobs.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateExplanationOfBenefit(ctx context.Context, eob *ExplanationOfBenefit) error {
	if eob.Status != "" && !validEOBStatuses[eob.Status] {
		return fmt.Errorf("invalid explanation_of_benefit status: %s", eob.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ExplanationOfBenefit", eob.FHIRID, eob.VersionID, eob.ToFHIR())
		if err == nil {
			eob.VersionID = newVer
		}
	}
	return s.eobs.Update(ctx, eob)
}

func (s *Service) DeleteExplanationOfBenefit(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		eob, err := s.eobs.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ExplanationOfBenefit", eob.FHIRID, eob.VersionID)
		}
	}
	return s.eobs.Delete(ctx, id)
}

func (s *Service) ListExplanationOfBenefitsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ExplanationOfBenefit, int, error) {
	return s.eobs.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchExplanationOfBenefits(ctx context.Context, params map[string]string, limit, offset int) ([]*ExplanationOfBenefit, int, error) {
	return s.eobs.Search(ctx, params, limit, offset)
}

// -- Invoice --

var validInvoiceStatuses = map[string]bool{
	"draft": true, "issued": true, "balanced": true, "cancelled": true, "entered-in-error": true,
}

func (s *Service) CreateInvoice(ctx context.Context, inv *Invoice) error {
	if inv.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if inv.Status == "" {
		inv.Status = "draft"
	}
	if !validInvoiceStatuses[inv.Status] {
		return fmt.Errorf("invalid invoice status: %s", inv.Status)
	}
	return s.invoices.Create(ctx, inv)
}

func (s *Service) GetInvoice(ctx context.Context, id uuid.UUID) (*Invoice, error) {
	return s.invoices.GetByID(ctx, id)
}

func (s *Service) GetInvoiceByFHIRID(ctx context.Context, fhirID string) (*Invoice, error) {
	return s.invoices.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateInvoice(ctx context.Context, inv *Invoice) error {
	if inv.Status != "" && !validInvoiceStatuses[inv.Status] {
		return fmt.Errorf("invalid invoice status: %s", inv.Status)
	}
	return s.invoices.Update(ctx, inv)
}

func (s *Service) DeleteInvoice(ctx context.Context, id uuid.UUID) error {
	return s.invoices.Delete(ctx, id)
}

func (s *Service) ListInvoicesByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Invoice, int, error) {
	return s.invoices.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchInvoices(ctx context.Context, params map[string]string, limit, offset int) ([]*Invoice, int, error) {
	return s.invoices.Search(ctx, params, limit, offset)
}

func (s *Service) AddInvoiceLineItem(ctx context.Context, li *InvoiceLineItem) error {
	if li.InvoiceID == uuid.Nil {
		return fmt.Errorf("invoice_id is required")
	}
	return s.invoices.AddLineItem(ctx, li)
}

func (s *Service) GetInvoiceLineItems(ctx context.Context, invoiceID uuid.UUID) ([]*InvoiceLineItem, error) {
	return s.invoices.GetLineItems(ctx, invoiceID)
}
