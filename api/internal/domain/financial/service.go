package financial

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	accounts               AccountRepository
	insurancePlans         InsurancePlanRepository
	paymentNotices         PaymentNoticeRepository
	paymentReconciliations PaymentReconciliationRepository
	chargeItems            ChargeItemRepository
	chargeItemDefinitions  ChargeItemDefinitionRepository
	contracts              ContractRepository
	enrollmentRequests     EnrollmentRequestRepository
	enrollmentResponses    EnrollmentResponseRepository
	vt                     *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(
	accounts AccountRepository,
	insurancePlans InsurancePlanRepository,
	paymentNotices PaymentNoticeRepository,
	paymentReconciliations PaymentReconciliationRepository,
	chargeItems ChargeItemRepository,
	chargeItemDefinitions ChargeItemDefinitionRepository,
	contracts ContractRepository,
	enrollmentRequests EnrollmentRequestRepository,
	enrollmentResponses EnrollmentResponseRepository,
) *Service {
	return &Service{
		accounts:               accounts,
		insurancePlans:         insurancePlans,
		paymentNotices:         paymentNotices,
		paymentReconciliations: paymentReconciliations,
		chargeItems:            chargeItems,
		chargeItemDefinitions:  chargeItemDefinitions,
		contracts:              contracts,
		enrollmentRequests:     enrollmentRequests,
		enrollmentResponses:    enrollmentResponses,
	}
}

// -- Account --

var validAccountStatuses = map[string]bool{
	"active": true, "inactive": true, "entered-in-error": true, "on-hold": true, "unknown": true,
}

func (s *Service) CreateAccount(ctx context.Context, a *Account) error {
	if a.Status == "" {
		a.Status = "active"
	}
	if !validAccountStatuses[a.Status] {
		return fmt.Errorf("invalid account status: %s", a.Status)
	}
	if err := s.accounts.Create(ctx, a); err != nil {
		return err
	}
	a.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Account", a.FHIRID, a.ToFHIR())
	}
	return nil
}

func (s *Service) GetAccount(ctx context.Context, id uuid.UUID) (*Account, error) {
	return s.accounts.GetByID(ctx, id)
}

func (s *Service) GetAccountByFHIRID(ctx context.Context, fhirID string) (*Account, error) {
	return s.accounts.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateAccount(ctx context.Context, a *Account) error {
	if a.Status != "" && !validAccountStatuses[a.Status] {
		return fmt.Errorf("invalid account status: %s", a.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Account", a.FHIRID, a.VersionID, a.ToFHIR())
		if err == nil {
			a.VersionID = newVer
		}
	}
	return s.accounts.Update(ctx, a)
}

func (s *Service) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		a, err := s.accounts.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Account", a.FHIRID, a.VersionID)
		}
	}
	return s.accounts.Delete(ctx, id)
}

func (s *Service) ListAccounts(ctx context.Context, limit, offset int) ([]*Account, int, error) {
	return s.accounts.List(ctx, limit, offset)
}

func (s *Service) SearchAccounts(ctx context.Context, params map[string]string, limit, offset int) ([]*Account, int, error) {
	return s.accounts.Search(ctx, params, limit, offset)
}

// -- InsurancePlan --

var validInsurancePlanStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

func (s *Service) CreateInsurancePlan(ctx context.Context, ip *InsurancePlan) error {
	if ip.Status == "" {
		ip.Status = "draft"
	}
	if !validInsurancePlanStatuses[ip.Status] {
		return fmt.Errorf("invalid insurance plan status: %s", ip.Status)
	}
	if err := s.insurancePlans.Create(ctx, ip); err != nil {
		return err
	}
	ip.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "InsurancePlan", ip.FHIRID, ip.ToFHIR())
	}
	return nil
}

func (s *Service) GetInsurancePlan(ctx context.Context, id uuid.UUID) (*InsurancePlan, error) {
	return s.insurancePlans.GetByID(ctx, id)
}

func (s *Service) GetInsurancePlanByFHIRID(ctx context.Context, fhirID string) (*InsurancePlan, error) {
	return s.insurancePlans.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateInsurancePlan(ctx context.Context, ip *InsurancePlan) error {
	if ip.Status != "" && !validInsurancePlanStatuses[ip.Status] {
		return fmt.Errorf("invalid insurance plan status: %s", ip.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "InsurancePlan", ip.FHIRID, ip.VersionID, ip.ToFHIR())
		if err == nil {
			ip.VersionID = newVer
		}
	}
	return s.insurancePlans.Update(ctx, ip)
}

func (s *Service) DeleteInsurancePlan(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		ip, err := s.insurancePlans.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "InsurancePlan", ip.FHIRID, ip.VersionID)
		}
	}
	return s.insurancePlans.Delete(ctx, id)
}

func (s *Service) ListInsurancePlans(ctx context.Context, limit, offset int) ([]*InsurancePlan, int, error) {
	return s.insurancePlans.List(ctx, limit, offset)
}

func (s *Service) SearchInsurancePlans(ctx context.Context, params map[string]string, limit, offset int) ([]*InsurancePlan, int, error) {
	return s.insurancePlans.Search(ctx, params, limit, offset)
}

// -- PaymentNotice --

var validPaymentNoticeStatuses = map[string]bool{
	"active": true, "cancelled": true, "draft": true, "entered-in-error": true,
}

func (s *Service) CreatePaymentNotice(ctx context.Context, pn *PaymentNotice) error {
	if pn.Status == "" {
		pn.Status = "active"
	}
	if !validPaymentNoticeStatuses[pn.Status] {
		return fmt.Errorf("invalid payment notice status: %s", pn.Status)
	}
	if err := s.paymentNotices.Create(ctx, pn); err != nil {
		return err
	}
	pn.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "PaymentNotice", pn.FHIRID, pn.ToFHIR())
	}
	return nil
}

func (s *Service) GetPaymentNotice(ctx context.Context, id uuid.UUID) (*PaymentNotice, error) {
	return s.paymentNotices.GetByID(ctx, id)
}

func (s *Service) GetPaymentNoticeByFHIRID(ctx context.Context, fhirID string) (*PaymentNotice, error) {
	return s.paymentNotices.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdatePaymentNotice(ctx context.Context, pn *PaymentNotice) error {
	if pn.Status != "" && !validPaymentNoticeStatuses[pn.Status] {
		return fmt.Errorf("invalid payment notice status: %s", pn.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "PaymentNotice", pn.FHIRID, pn.VersionID, pn.ToFHIR())
		if err == nil {
			pn.VersionID = newVer
		}
	}
	return s.paymentNotices.Update(ctx, pn)
}

func (s *Service) DeletePaymentNotice(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		pn, err := s.paymentNotices.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "PaymentNotice", pn.FHIRID, pn.VersionID)
		}
	}
	return s.paymentNotices.Delete(ctx, id)
}

func (s *Service) ListPaymentNotices(ctx context.Context, limit, offset int) ([]*PaymentNotice, int, error) {
	return s.paymentNotices.List(ctx, limit, offset)
}

func (s *Service) SearchPaymentNotices(ctx context.Context, params map[string]string, limit, offset int) ([]*PaymentNotice, int, error) {
	return s.paymentNotices.Search(ctx, params, limit, offset)
}

// -- PaymentReconciliation --

var validPaymentReconciliationStatuses = map[string]bool{
	"active": true, "cancelled": true, "draft": true, "entered-in-error": true,
}

func (s *Service) CreatePaymentReconciliation(ctx context.Context, pr *PaymentReconciliation) error {
	if pr.Status == "" {
		pr.Status = "active"
	}
	if !validPaymentReconciliationStatuses[pr.Status] {
		return fmt.Errorf("invalid payment reconciliation status: %s", pr.Status)
	}
	if err := s.paymentReconciliations.Create(ctx, pr); err != nil {
		return err
	}
	pr.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "PaymentReconciliation", pr.FHIRID, pr.ToFHIR())
	}
	return nil
}

func (s *Service) GetPaymentReconciliation(ctx context.Context, id uuid.UUID) (*PaymentReconciliation, error) {
	return s.paymentReconciliations.GetByID(ctx, id)
}

func (s *Service) GetPaymentReconciliationByFHIRID(ctx context.Context, fhirID string) (*PaymentReconciliation, error) {
	return s.paymentReconciliations.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdatePaymentReconciliation(ctx context.Context, pr *PaymentReconciliation) error {
	if pr.Status != "" && !validPaymentReconciliationStatuses[pr.Status] {
		return fmt.Errorf("invalid payment reconciliation status: %s", pr.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "PaymentReconciliation", pr.FHIRID, pr.VersionID, pr.ToFHIR())
		if err == nil {
			pr.VersionID = newVer
		}
	}
	return s.paymentReconciliations.Update(ctx, pr)
}

func (s *Service) DeletePaymentReconciliation(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		pr, err := s.paymentReconciliations.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "PaymentReconciliation", pr.FHIRID, pr.VersionID)
		}
	}
	return s.paymentReconciliations.Delete(ctx, id)
}

func (s *Service) ListPaymentReconciliations(ctx context.Context, limit, offset int) ([]*PaymentReconciliation, int, error) {
	return s.paymentReconciliations.List(ctx, limit, offset)
}

func (s *Service) SearchPaymentReconciliations(ctx context.Context, params map[string]string, limit, offset int) ([]*PaymentReconciliation, int, error) {
	return s.paymentReconciliations.Search(ctx, params, limit, offset)
}

// -- ChargeItem --

var validChargeItemStatuses = map[string]bool{
	"planned": true, "billable": true, "not-billable": true, "aborted": true,
	"billed": true, "entered-in-error": true, "unknown": true,
}

func (s *Service) CreateChargeItem(ctx context.Context, ci *ChargeItem) error {
	if ci.SubjectPatientID == uuid.Nil {
		return fmt.Errorf("subject_patient_id is required")
	}
	if ci.Status == "" {
		ci.Status = "planned"
	}
	if !validChargeItemStatuses[ci.Status] {
		return fmt.Errorf("invalid charge item status: %s", ci.Status)
	}
	if err := s.chargeItems.Create(ctx, ci); err != nil {
		return err
	}
	ci.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ChargeItem", ci.FHIRID, ci.ToFHIR())
	}
	return nil
}

func (s *Service) GetChargeItem(ctx context.Context, id uuid.UUID) (*ChargeItem, error) {
	return s.chargeItems.GetByID(ctx, id)
}

func (s *Service) GetChargeItemByFHIRID(ctx context.Context, fhirID string) (*ChargeItem, error) {
	return s.chargeItems.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateChargeItem(ctx context.Context, ci *ChargeItem) error {
	if ci.Status != "" && !validChargeItemStatuses[ci.Status] {
		return fmt.Errorf("invalid charge item status: %s", ci.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ChargeItem", ci.FHIRID, ci.VersionID, ci.ToFHIR())
		if err == nil {
			ci.VersionID = newVer
		}
	}
	return s.chargeItems.Update(ctx, ci)
}

func (s *Service) DeleteChargeItem(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		ci, err := s.chargeItems.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ChargeItem", ci.FHIRID, ci.VersionID)
		}
	}
	return s.chargeItems.Delete(ctx, id)
}

func (s *Service) ListChargeItems(ctx context.Context, limit, offset int) ([]*ChargeItem, int, error) {
	return s.chargeItems.List(ctx, limit, offset)
}

func (s *Service) SearchChargeItems(ctx context.Context, params map[string]string, limit, offset int) ([]*ChargeItem, int, error) {
	return s.chargeItems.Search(ctx, params, limit, offset)
}

// -- ChargeItemDefinition --

var validChargeItemDefinitionStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

func (s *Service) CreateChargeItemDefinition(ctx context.Context, cd *ChargeItemDefinition) error {
	if cd.Status == "" {
		cd.Status = "draft"
	}
	if !validChargeItemDefinitionStatuses[cd.Status] {
		return fmt.Errorf("invalid charge item definition status: %s", cd.Status)
	}
	if err := s.chargeItemDefinitions.Create(ctx, cd); err != nil {
		return err
	}
	cd.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ChargeItemDefinition", cd.FHIRID, cd.ToFHIR())
	}
	return nil
}

func (s *Service) GetChargeItemDefinition(ctx context.Context, id uuid.UUID) (*ChargeItemDefinition, error) {
	return s.chargeItemDefinitions.GetByID(ctx, id)
}

func (s *Service) GetChargeItemDefinitionByFHIRID(ctx context.Context, fhirID string) (*ChargeItemDefinition, error) {
	return s.chargeItemDefinitions.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateChargeItemDefinition(ctx context.Context, cd *ChargeItemDefinition) error {
	if cd.Status != "" && !validChargeItemDefinitionStatuses[cd.Status] {
		return fmt.Errorf("invalid charge item definition status: %s", cd.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ChargeItemDefinition", cd.FHIRID, cd.VersionID, cd.ToFHIR())
		if err == nil {
			cd.VersionID = newVer
		}
	}
	return s.chargeItemDefinitions.Update(ctx, cd)
}

func (s *Service) DeleteChargeItemDefinition(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		cd, err := s.chargeItemDefinitions.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ChargeItemDefinition", cd.FHIRID, cd.VersionID)
		}
	}
	return s.chargeItemDefinitions.Delete(ctx, id)
}

func (s *Service) ListChargeItemDefinitions(ctx context.Context, limit, offset int) ([]*ChargeItemDefinition, int, error) {
	return s.chargeItemDefinitions.List(ctx, limit, offset)
}

func (s *Service) SearchChargeItemDefinitions(ctx context.Context, params map[string]string, limit, offset int) ([]*ChargeItemDefinition, int, error) {
	return s.chargeItemDefinitions.Search(ctx, params, limit, offset)
}

// -- Contract --

var validContractStatuses = map[string]bool{
	"amended": true, "appended": true, "cancelled": true, "disputed": true,
	"entered-in-error": true, "executable": true, "executed": true,
	"negotiable": true, "offered": true, "policy": true, "rejected": true,
	"renewed": true, "revoked": true, "resolved": true, "terminated": true,
}

func (s *Service) CreateContract(ctx context.Context, ct *Contract) error {
	if ct.Status == "" {
		ct.Status = "offered"
	}
	if !validContractStatuses[ct.Status] {
		return fmt.Errorf("invalid contract status: %s", ct.Status)
	}
	if err := s.contracts.Create(ctx, ct); err != nil {
		return err
	}
	ct.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Contract", ct.FHIRID, ct.ToFHIR())
	}
	return nil
}

func (s *Service) GetContract(ctx context.Context, id uuid.UUID) (*Contract, error) {
	return s.contracts.GetByID(ctx, id)
}

func (s *Service) GetContractByFHIRID(ctx context.Context, fhirID string) (*Contract, error) {
	return s.contracts.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateContract(ctx context.Context, ct *Contract) error {
	if ct.Status != "" && !validContractStatuses[ct.Status] {
		return fmt.Errorf("invalid contract status: %s", ct.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Contract", ct.FHIRID, ct.VersionID, ct.ToFHIR())
		if err == nil {
			ct.VersionID = newVer
		}
	}
	return s.contracts.Update(ctx, ct)
}

func (s *Service) DeleteContract(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		ct, err := s.contracts.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Contract", ct.FHIRID, ct.VersionID)
		}
	}
	return s.contracts.Delete(ctx, id)
}

func (s *Service) ListContracts(ctx context.Context, limit, offset int) ([]*Contract, int, error) {
	return s.contracts.List(ctx, limit, offset)
}

func (s *Service) SearchContracts(ctx context.Context, params map[string]string, limit, offset int) ([]*Contract, int, error) {
	return s.contracts.Search(ctx, params, limit, offset)
}

// -- EnrollmentRequest --

var validEnrollmentRequestStatuses = map[string]bool{
	"active": true, "cancelled": true, "draft": true, "entered-in-error": true,
}

func (s *Service) CreateEnrollmentRequest(ctx context.Context, er *EnrollmentRequest) error {
	if er.Status == "" {
		er.Status = "active"
	}
	if !validEnrollmentRequestStatuses[er.Status] {
		return fmt.Errorf("invalid enrollment request status: %s", er.Status)
	}
	if err := s.enrollmentRequests.Create(ctx, er); err != nil {
		return err
	}
	er.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "EnrollmentRequest", er.FHIRID, er.ToFHIR())
	}
	return nil
}

func (s *Service) GetEnrollmentRequest(ctx context.Context, id uuid.UUID) (*EnrollmentRequest, error) {
	return s.enrollmentRequests.GetByID(ctx, id)
}

func (s *Service) GetEnrollmentRequestByFHIRID(ctx context.Context, fhirID string) (*EnrollmentRequest, error) {
	return s.enrollmentRequests.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateEnrollmentRequest(ctx context.Context, er *EnrollmentRequest) error {
	if er.Status != "" && !validEnrollmentRequestStatuses[er.Status] {
		return fmt.Errorf("invalid enrollment request status: %s", er.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "EnrollmentRequest", er.FHIRID, er.VersionID, er.ToFHIR())
		if err == nil {
			er.VersionID = newVer
		}
	}
	return s.enrollmentRequests.Update(ctx, er)
}

func (s *Service) DeleteEnrollmentRequest(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		er, err := s.enrollmentRequests.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "EnrollmentRequest", er.FHIRID, er.VersionID)
		}
	}
	return s.enrollmentRequests.Delete(ctx, id)
}

func (s *Service) ListEnrollmentRequests(ctx context.Context, limit, offset int) ([]*EnrollmentRequest, int, error) {
	return s.enrollmentRequests.List(ctx, limit, offset)
}

func (s *Service) SearchEnrollmentRequests(ctx context.Context, params map[string]string, limit, offset int) ([]*EnrollmentRequest, int, error) {
	return s.enrollmentRequests.Search(ctx, params, limit, offset)
}

// -- EnrollmentResponse --

var validEnrollmentResponseStatuses = map[string]bool{
	"active": true, "cancelled": true, "draft": true, "entered-in-error": true,
}

func (s *Service) CreateEnrollmentResponse(ctx context.Context, er *EnrollmentResponse) error {
	if er.Status == "" {
		er.Status = "active"
	}
	if !validEnrollmentResponseStatuses[er.Status] {
		return fmt.Errorf("invalid enrollment response status: %s", er.Status)
	}
	if err := s.enrollmentResponses.Create(ctx, er); err != nil {
		return err
	}
	er.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "EnrollmentResponse", er.FHIRID, er.ToFHIR())
	}
	return nil
}

func (s *Service) GetEnrollmentResponse(ctx context.Context, id uuid.UUID) (*EnrollmentResponse, error) {
	return s.enrollmentResponses.GetByID(ctx, id)
}

func (s *Service) GetEnrollmentResponseByFHIRID(ctx context.Context, fhirID string) (*EnrollmentResponse, error) {
	return s.enrollmentResponses.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateEnrollmentResponse(ctx context.Context, er *EnrollmentResponse) error {
	if er.Status != "" && !validEnrollmentResponseStatuses[er.Status] {
		return fmt.Errorf("invalid enrollment response status: %s", er.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "EnrollmentResponse", er.FHIRID, er.VersionID, er.ToFHIR())
		if err == nil {
			er.VersionID = newVer
		}
	}
	return s.enrollmentResponses.Update(ctx, er)
}

func (s *Service) DeleteEnrollmentResponse(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		er, err := s.enrollmentResponses.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "EnrollmentResponse", er.FHIRID, er.VersionID)
		}
	}
	return s.enrollmentResponses.Delete(ctx, id)
}

func (s *Service) ListEnrollmentResponses(ctx context.Context, limit, offset int) ([]*EnrollmentResponse, int, error) {
	return s.enrollmentResponses.List(ctx, limit, offset)
}

func (s *Service) SearchEnrollmentResponses(ctx context.Context, params map[string]string, limit, offset int) ([]*EnrollmentResponse, int, error) {
	return s.enrollmentResponses.Search(ctx, params, limit, offset)
}
