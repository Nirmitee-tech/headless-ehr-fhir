package billing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// -- Mock Repositories --

type mockCoverageRepo struct {
	items map[uuid.UUID]*Coverage
}

func newMockCoverageRepo() *mockCoverageRepo {
	return &mockCoverageRepo{items: make(map[uuid.UUID]*Coverage)}
}

func (m *mockCoverageRepo) Create(_ context.Context, c *Coverage) error {
	c.ID = uuid.New()
	if c.FHIRID == "" {
		c.FHIRID = c.ID.String()
	}
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	m.items[c.ID] = c
	return nil
}

func (m *mockCoverageRepo) GetByID(_ context.Context, id uuid.UUID) (*Coverage, error) {
	c, ok := m.items[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

func (m *mockCoverageRepo) GetByFHIRID(_ context.Context, fhirID string) (*Coverage, error) {
	for _, c := range m.items {
		if c.FHIRID == fhirID {
			return c, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockCoverageRepo) Update(_ context.Context, c *Coverage) error {
	m.items[c.ID] = c
	return nil
}

func (m *mockCoverageRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockCoverageRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*Coverage, int, error) {
	var result []*Coverage
	for _, c := range m.items {
		if c.PatientID == patientID {
			result = append(result, c)
		}
	}
	return result, len(result), nil
}

func (m *mockCoverageRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*Coverage, int, error) {
	var result []*Coverage
	for _, c := range m.items {
		result = append(result, c)
	}
	return result, len(result), nil
}

type mockClaimRepo struct {
	items      map[uuid.UUID]*Claim
	diagnoses  map[uuid.UUID]*ClaimDiagnosis
	procedures map[uuid.UUID]*ClaimProcedure
	claimItems map[uuid.UUID]*ClaimItem
}

func newMockClaimRepo() *mockClaimRepo {
	return &mockClaimRepo{
		items:      make(map[uuid.UUID]*Claim),
		diagnoses:  make(map[uuid.UUID]*ClaimDiagnosis),
		procedures: make(map[uuid.UUID]*ClaimProcedure),
		claimItems: make(map[uuid.UUID]*ClaimItem),
	}
}

func (m *mockClaimRepo) Create(_ context.Context, c *Claim) error {
	c.ID = uuid.New()
	if c.FHIRID == "" {
		c.FHIRID = c.ID.String()
	}
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	m.items[c.ID] = c
	return nil
}

func (m *mockClaimRepo) GetByID(_ context.Context, id uuid.UUID) (*Claim, error) {
	c, ok := m.items[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

func (m *mockClaimRepo) GetByFHIRID(_ context.Context, fhirID string) (*Claim, error) {
	for _, c := range m.items {
		if c.FHIRID == fhirID {
			return c, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockClaimRepo) Update(_ context.Context, c *Claim) error {
	m.items[c.ID] = c
	return nil
}

func (m *mockClaimRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockClaimRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*Claim, int, error) {
	var result []*Claim
	for _, c := range m.items {
		if c.PatientID == patientID {
			result = append(result, c)
		}
	}
	return result, len(result), nil
}

func (m *mockClaimRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*Claim, int, error) {
	var result []*Claim
	for _, c := range m.items {
		result = append(result, c)
	}
	return result, len(result), nil
}

func (m *mockClaimRepo) AddDiagnosis(_ context.Context, d *ClaimDiagnosis) error {
	d.ID = uuid.New()
	m.diagnoses[d.ID] = d
	return nil
}

func (m *mockClaimRepo) GetDiagnoses(_ context.Context, claimID uuid.UUID) ([]*ClaimDiagnosis, error) {
	var result []*ClaimDiagnosis
	for _, d := range m.diagnoses {
		if d.ClaimID == claimID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (m *mockClaimRepo) AddProcedure(_ context.Context, p *ClaimProcedure) error {
	p.ID = uuid.New()
	m.procedures[p.ID] = p
	return nil
}

func (m *mockClaimRepo) GetProcedures(_ context.Context, claimID uuid.UUID) ([]*ClaimProcedure, error) {
	var result []*ClaimProcedure
	for _, p := range m.procedures {
		if p.ClaimID == claimID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockClaimRepo) AddItem(_ context.Context, item *ClaimItem) error {
	item.ID = uuid.New()
	m.claimItems[item.ID] = item
	return nil
}

func (m *mockClaimRepo) GetItems(_ context.Context, claimID uuid.UUID) ([]*ClaimItem, error) {
	var result []*ClaimItem
	for _, item := range m.claimItems {
		if item.ClaimID == claimID {
			result = append(result, item)
		}
	}
	return result, nil
}

type mockClaimResponseRepo struct {
	items map[uuid.UUID]*ClaimResponse
}

func newMockClaimResponseRepo() *mockClaimResponseRepo {
	return &mockClaimResponseRepo{items: make(map[uuid.UUID]*ClaimResponse)}
}

func (m *mockClaimResponseRepo) Create(_ context.Context, cr *ClaimResponse) error {
	cr.ID = uuid.New()
	if cr.FHIRID == "" {
		cr.FHIRID = cr.ID.String()
	}
	cr.CreatedAt = time.Now()
	m.items[cr.ID] = cr
	return nil
}

func (m *mockClaimResponseRepo) GetByID(_ context.Context, id uuid.UUID) (*ClaimResponse, error) {
	cr, ok := m.items[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return cr, nil
}

func (m *mockClaimResponseRepo) GetByFHIRID(_ context.Context, fhirID string) (*ClaimResponse, error) {
	for _, cr := range m.items {
		if cr.FHIRID == fhirID {
			return cr, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockClaimResponseRepo) ListByClaim(_ context.Context, claimID uuid.UUID, limit, offset int) ([]*ClaimResponse, int, error) {
	var result []*ClaimResponse
	for _, cr := range m.items {
		if cr.ClaimID == claimID {
			result = append(result, cr)
		}
	}
	return result, len(result), nil
}

func (m *mockClaimResponseRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*ClaimResponse, int, error) {
	var result []*ClaimResponse
	for _, cr := range m.items {
		result = append(result, cr)
	}
	return result, len(result), nil
}

type mockInvoiceRepo struct {
	items     map[uuid.UUID]*Invoice
	lineItems map[uuid.UUID]*InvoiceLineItem
}

func newMockInvoiceRepo() *mockInvoiceRepo {
	return &mockInvoiceRepo{
		items:     make(map[uuid.UUID]*Invoice),
		lineItems: make(map[uuid.UUID]*InvoiceLineItem),
	}
}

func (m *mockInvoiceRepo) Create(_ context.Context, inv *Invoice) error {
	inv.ID = uuid.New()
	if inv.FHIRID == "" {
		inv.FHIRID = inv.ID.String()
	}
	inv.CreatedAt = time.Now()
	m.items[inv.ID] = inv
	return nil
}

func (m *mockInvoiceRepo) GetByID(_ context.Context, id uuid.UUID) (*Invoice, error) {
	inv, ok := m.items[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return inv, nil
}

func (m *mockInvoiceRepo) GetByFHIRID(_ context.Context, fhirID string) (*Invoice, error) {
	for _, inv := range m.items {
		if inv.FHIRID == fhirID {
			return inv, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockInvoiceRepo) Update(_ context.Context, inv *Invoice) error {
	m.items[inv.ID] = inv
	return nil
}

func (m *mockInvoiceRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockInvoiceRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*Invoice, int, error) {
	var result []*Invoice
	for _, inv := range m.items {
		if inv.PatientID == patientID {
			result = append(result, inv)
		}
	}
	return result, len(result), nil
}

func (m *mockInvoiceRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*Invoice, int, error) {
	var result []*Invoice
	for _, inv := range m.items {
		result = append(result, inv)
	}
	return result, len(result), nil
}

func (m *mockInvoiceRepo) AddLineItem(_ context.Context, li *InvoiceLineItem) error {
	li.ID = uuid.New()
	m.lineItems[li.ID] = li
	return nil
}

func (m *mockInvoiceRepo) GetLineItems(_ context.Context, invoiceID uuid.UUID) ([]*InvoiceLineItem, error) {
	var result []*InvoiceLineItem
	for _, li := range m.lineItems {
		if li.InvoiceID == invoiceID {
			result = append(result, li)
		}
	}
	return result, nil
}

// -- Tests --

func newTestService() *Service {
	return NewService(newMockCoverageRepo(), newMockClaimRepo(), newMockClaimResponseRepo(), newMockInvoiceRepo())
}

// -- Coverage Tests --

func TestCreateCoverage(t *testing.T) {
	svc := newTestService()
	payorName := "Blue Cross"
	c := &Coverage{PatientID: uuid.New(), PayorName: &payorName}
	err := svc.CreateCoverage(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Status != "active" {
		t.Errorf("expected default status 'active', got %s", c.Status)
	}
}

func TestCreateCoverage_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	payorName := "Blue Cross"
	c := &Coverage{PayorName: &payorName}
	err := svc.CreateCoverage(context.Background(), c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateCoverage_PayorRequired(t *testing.T) {
	svc := newTestService()
	c := &Coverage{PatientID: uuid.New()}
	err := svc.CreateCoverage(context.Background(), c)
	if err == nil {
		t.Error("expected error for missing payor information")
	}
}

func TestGetCoverage(t *testing.T) {
	svc := newTestService()
	payorName := "Blue Cross"
	c := &Coverage{PatientID: uuid.New(), PayorName: &payorName}
	svc.CreateCoverage(context.Background(), c)

	fetched, err := svc.GetCoverage(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != c.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestDeleteCoverage(t *testing.T) {
	svc := newTestService()
	payorName := "Blue Cross"
	c := &Coverage{PatientID: uuid.New(), PayorName: &payorName}
	svc.CreateCoverage(context.Background(), c)
	err := svc.DeleteCoverage(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetCoverage(context.Background(), c.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

// -- Claim Tests --

func TestCreateClaim(t *testing.T) {
	svc := newTestService()
	c := &Claim{PatientID: uuid.New()}
	err := svc.CreateClaim(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Status != "draft" {
		t.Errorf("expected default status 'draft', got %s", c.Status)
	}
}

func TestCreateClaim_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	c := &Claim{}
	err := svc.CreateClaim(context.Background(), c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestGetClaim(t *testing.T) {
	svc := newTestService()
	c := &Claim{PatientID: uuid.New()}
	svc.CreateClaim(context.Background(), c)

	fetched, err := svc.GetClaim(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != c.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestDeleteClaim(t *testing.T) {
	svc := newTestService()
	c := &Claim{PatientID: uuid.New()}
	svc.CreateClaim(context.Background(), c)
	err := svc.DeleteClaim(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetClaim(context.Background(), c.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestAddClaimDiagnosis(t *testing.T) {
	svc := newTestService()
	c := &Claim{PatientID: uuid.New()}
	svc.CreateClaim(context.Background(), c)

	d := &ClaimDiagnosis{ClaimID: c.ID, DiagnosisCode: "J06.9"}
	err := svc.AddClaimDiagnosis(context.Background(), d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	diagnoses, err := svc.GetClaimDiagnoses(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diagnoses) != 1 {
		t.Errorf("expected 1 diagnosis, got %d", len(diagnoses))
	}
}

func TestAddClaimDiagnosis_ClaimIDRequired(t *testing.T) {
	svc := newTestService()
	d := &ClaimDiagnosis{DiagnosisCode: "J06.9"}
	err := svc.AddClaimDiagnosis(context.Background(), d)
	if err == nil {
		t.Error("expected error for missing claim_id")
	}
}

func TestAddClaimDiagnosis_DiagnosisCodeRequired(t *testing.T) {
	svc := newTestService()
	d := &ClaimDiagnosis{ClaimID: uuid.New()}
	err := svc.AddClaimDiagnosis(context.Background(), d)
	if err == nil {
		t.Error("expected error for missing diagnosis_code")
	}
}

func TestAddClaimProcedure(t *testing.T) {
	svc := newTestService()
	c := &Claim{PatientID: uuid.New()}
	svc.CreateClaim(context.Background(), c)

	p := &ClaimProcedure{ClaimID: c.ID, ProcedureCode: "99213"}
	err := svc.AddClaimProcedure(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	procedures, err := svc.GetClaimProcedures(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(procedures) != 1 {
		t.Errorf("expected 1 procedure, got %d", len(procedures))
	}
}

func TestAddClaimProcedure_ProcedureCodeRequired(t *testing.T) {
	svc := newTestService()
	p := &ClaimProcedure{ClaimID: uuid.New()}
	err := svc.AddClaimProcedure(context.Background(), p)
	if err == nil {
		t.Error("expected error for missing procedure_code")
	}
}

func TestAddClaimItem(t *testing.T) {
	svc := newTestService()
	c := &Claim{PatientID: uuid.New()}
	svc.CreateClaim(context.Background(), c)

	item := &ClaimItem{ClaimID: c.ID, ProductOrServiceCode: "99213"}
	err := svc.AddClaimItem(context.Background(), item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items, err := svc.GetClaimItems(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestAddClaimItem_ProductOrServiceCodeRequired(t *testing.T) {
	svc := newTestService()
	item := &ClaimItem{ClaimID: uuid.New()}
	err := svc.AddClaimItem(context.Background(), item)
	if err == nil {
		t.Error("expected error for missing product_or_service_code")
	}
}

// -- ClaimResponse Tests --

func TestCreateClaimResponse(t *testing.T) {
	svc := newTestService()
	cr := &ClaimResponse{ClaimID: uuid.New()}
	err := svc.CreateClaimResponse(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cr.Status != "active" {
		t.Errorf("expected default status 'active', got %s", cr.Status)
	}
}

func TestCreateClaimResponse_ClaimIDRequired(t *testing.T) {
	svc := newTestService()
	cr := &ClaimResponse{}
	err := svc.CreateClaimResponse(context.Background(), cr)
	if err == nil {
		t.Error("expected error for missing claim_id")
	}
}

func TestGetClaimResponse(t *testing.T) {
	svc := newTestService()
	cr := &ClaimResponse{ClaimID: uuid.New()}
	svc.CreateClaimResponse(context.Background(), cr)

	fetched, err := svc.GetClaimResponse(context.Background(), cr.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != cr.ID {
		t.Error("unexpected ID mismatch")
	}
}

// -- Invoice Tests --

func TestCreateInvoice(t *testing.T) {
	svc := newTestService()
	inv := &Invoice{PatientID: uuid.New()}
	err := svc.CreateInvoice(context.Background(), inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Status != "draft" {
		t.Errorf("expected default status 'draft', got %s", inv.Status)
	}
}

func TestCreateInvoice_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	inv := &Invoice{}
	err := svc.CreateInvoice(context.Background(), inv)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestGetInvoice(t *testing.T) {
	svc := newTestService()
	inv := &Invoice{PatientID: uuid.New()}
	svc.CreateInvoice(context.Background(), inv)

	fetched, err := svc.GetInvoice(context.Background(), inv.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != inv.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestDeleteInvoice(t *testing.T) {
	svc := newTestService()
	inv := &Invoice{PatientID: uuid.New()}
	svc.CreateInvoice(context.Background(), inv)
	err := svc.DeleteInvoice(context.Background(), inv.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetInvoice(context.Background(), inv.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestAddInvoiceLineItem(t *testing.T) {
	svc := newTestService()
	inv := &Invoice{PatientID: uuid.New()}
	svc.CreateInvoice(context.Background(), inv)

	li := &InvoiceLineItem{InvoiceID: inv.ID, Sequence: 1}
	err := svc.AddInvoiceLineItem(context.Background(), li)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items, err := svc.GetInvoiceLineItems(context.Background(), inv.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 line item, got %d", len(items))
	}
}

func TestAddInvoiceLineItem_InvoiceIDRequired(t *testing.T) {
	svc := newTestService()
	li := &InvoiceLineItem{Sequence: 1}
	err := svc.AddInvoiceLineItem(context.Background(), li)
	if err == nil {
		t.Error("expected error for missing invoice_id")
	}
}

// -- Coverage: GetByFHIRID, Update, List, Search --

func TestGetCoverageByFHIRID(t *testing.T) {
	svc := newTestService()
	payorName := "Blue Cross"
	c := &Coverage{PatientID: uuid.New(), PayorName: &payorName}
	svc.CreateCoverage(context.Background(), c)

	fetched, err := svc.GetCoverageByFHIRID(context.Background(), c.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != c.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestGetCoverageByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetCoverageByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent FHIR ID")
	}
}

func TestUpdateCoverage(t *testing.T) {
	svc := newTestService()
	payorName := "Blue Cross"
	c := &Coverage{PatientID: uuid.New(), PayorName: &payorName}
	svc.CreateCoverage(context.Background(), c)

	c.Status = "cancelled"
	err := svc.UpdateCoverage(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fetched, _ := svc.GetCoverage(context.Background(), c.ID)
	if fetched.Status != "cancelled" {
		t.Errorf("expected status 'cancelled', got %s", fetched.Status)
	}
}

func TestUpdateCoverage_InvalidStatus(t *testing.T) {
	svc := newTestService()
	c := &Coverage{Status: "bogus"}
	err := svc.UpdateCoverage(context.Background(), c)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestListCoveragesByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	payorName := "Blue Cross"
	svc.CreateCoverage(context.Background(), &Coverage{PatientID: patientID, PayorName: &payorName})
	svc.CreateCoverage(context.Background(), &Coverage{PatientID: patientID, PayorName: &payorName})
	svc.CreateCoverage(context.Background(), &Coverage{PatientID: uuid.New(), PayorName: &payorName})

	results, total, err := svc.ListCoveragesByPatient(context.Background(), patientID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestSearchCoverages(t *testing.T) {
	svc := newTestService()
	payorName := "Blue Cross"
	svc.CreateCoverage(context.Background(), &Coverage{PatientID: uuid.New(), PayorName: &payorName})

	results, total, err := svc.SearchCoverages(context.Background(), map[string]string{"status": "active"}, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(results) < 1 {
		t.Error("expected at least 1 result")
	}
}

// -- Claim: GetByFHIRID, Update, List, Search --

func TestGetClaimByFHIRID(t *testing.T) {
	svc := newTestService()
	c := &Claim{PatientID: uuid.New()}
	svc.CreateClaim(context.Background(), c)

	fetched, err := svc.GetClaimByFHIRID(context.Background(), c.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != c.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestGetClaimByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetClaimByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent FHIR ID")
	}
}

func TestUpdateClaim(t *testing.T) {
	svc := newTestService()
	c := &Claim{PatientID: uuid.New()}
	svc.CreateClaim(context.Background(), c)

	c.Status = "active"
	err := svc.UpdateClaim(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateClaim_InvalidStatus(t *testing.T) {
	svc := newTestService()
	c := &Claim{Status: "bogus"}
	err := svc.UpdateClaim(context.Background(), c)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestListClaimsByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreateClaim(context.Background(), &Claim{PatientID: patientID})
	svc.CreateClaim(context.Background(), &Claim{PatientID: patientID})
	svc.CreateClaim(context.Background(), &Claim{PatientID: uuid.New()})

	results, total, err := svc.ListClaimsByPatient(context.Background(), patientID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestSearchClaims(t *testing.T) {
	svc := newTestService()
	svc.CreateClaim(context.Background(), &Claim{PatientID: uuid.New()})

	results, total, err := svc.SearchClaims(context.Background(), map[string]string{"status": "draft"}, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(results) < 1 {
		t.Error("expected at least 1 result")
	}
}

// -- ClaimResponse: GetByFHIRID, ListByClaim, Search --

func TestGetClaimResponseByFHIRID(t *testing.T) {
	svc := newTestService()
	cr := &ClaimResponse{ClaimID: uuid.New()}
	svc.CreateClaimResponse(context.Background(), cr)

	fetched, err := svc.GetClaimResponseByFHIRID(context.Background(), cr.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != cr.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestGetClaimResponseByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetClaimResponseByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent FHIR ID")
	}
}

func TestListClaimResponsesByClaim(t *testing.T) {
	svc := newTestService()
	claimID := uuid.New()
	svc.CreateClaimResponse(context.Background(), &ClaimResponse{ClaimID: claimID})
	svc.CreateClaimResponse(context.Background(), &ClaimResponse{ClaimID: claimID})
	svc.CreateClaimResponse(context.Background(), &ClaimResponse{ClaimID: uuid.New()})

	results, total, err := svc.ListClaimResponsesByClaim(context.Background(), claimID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestSearchClaimResponses(t *testing.T) {
	svc := newTestService()
	svc.CreateClaimResponse(context.Background(), &ClaimResponse{ClaimID: uuid.New()})

	results, total, err := svc.SearchClaimResponses(context.Background(), map[string]string{}, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(results) < 1 {
		t.Error("expected at least 1 result")
	}
}

// -- Invoice: GetByFHIRID, Update, List, Search --

func TestGetInvoiceByFHIRID(t *testing.T) {
	svc := newTestService()
	inv := &Invoice{PatientID: uuid.New()}
	svc.CreateInvoice(context.Background(), inv)

	fetched, err := svc.GetInvoiceByFHIRID(context.Background(), inv.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != inv.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestGetInvoiceByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetInvoiceByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent FHIR ID")
	}
}

func TestUpdateInvoice(t *testing.T) {
	svc := newTestService()
	inv := &Invoice{PatientID: uuid.New()}
	svc.CreateInvoice(context.Background(), inv)

	inv.Status = "issued"
	err := svc.UpdateInvoice(context.Background(), inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateInvoice_InvalidStatus(t *testing.T) {
	svc := newTestService()
	inv := &Invoice{Status: "bogus"}
	err := svc.UpdateInvoice(context.Background(), inv)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestListInvoicesByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreateInvoice(context.Background(), &Invoice{PatientID: patientID})
	svc.CreateInvoice(context.Background(), &Invoice{PatientID: patientID})
	svc.CreateInvoice(context.Background(), &Invoice{PatientID: uuid.New()})

	results, total, err := svc.ListInvoicesByPatient(context.Background(), patientID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestSearchInvoices(t *testing.T) {
	svc := newTestService()
	svc.CreateInvoice(context.Background(), &Invoice{PatientID: uuid.New()})

	results, total, err := svc.SearchInvoices(context.Background(), map[string]string{}, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(results) < 1 {
		t.Error("expected at least 1 result")
	}
}

// -- ToFHIR Tests --

func TestCoverageToFHIR(t *testing.T) {
	payorName := "Blue Cross"
	c := &Coverage{
		FHIRID:    "cov-123",
		Status:    "active",
		PatientID: uuid.New(),
		PayorName: &payorName,
		UpdatedAt: time.Now(),
	}
	fhirRes := c.ToFHIR()
	if fhirRes["resourceType"] != "Coverage" {
		t.Errorf("expected Coverage, got %v", fhirRes["resourceType"])
	}
	if fhirRes["status"] != "active" {
		t.Errorf("expected active, got %v", fhirRes["status"])
	}
}

func TestClaimToFHIR(t *testing.T) {
	c := &Claim{
		FHIRID:    "claim-123",
		Status:    "draft",
		PatientID: uuid.New(),
		UpdatedAt: time.Now(),
	}
	fhirRes := c.ToFHIR()
	if fhirRes["resourceType"] != "Claim" {
		t.Errorf("expected Claim, got %v", fhirRes["resourceType"])
	}
	if fhirRes["status"] != "draft" {
		t.Errorf("expected draft, got %v", fhirRes["status"])
	}
}

func TestClaimResponseToFHIR(t *testing.T) {
	cr := &ClaimResponse{
		FHIRID:    "cr-123",
		ClaimID:   uuid.New(),
		Status:    "active",
		CreatedAt: time.Now(),
	}
	fhirRes := cr.ToFHIR()
	if fhirRes["resourceType"] != "ClaimResponse" {
		t.Errorf("expected ClaimResponse, got %v", fhirRes["resourceType"])
	}
	if fhirRes["status"] != "active" {
		t.Errorf("expected active, got %v", fhirRes["status"])
	}
}

func TestExplanationOfBenefitToFHIR(t *testing.T) {
	eob := &ExplanationOfBenefit{
		FHIRID:    "eob-123",
		Status:    "active",
		PatientID: uuid.New(),
		CreatedAt: time.Now(),
	}
	fhirRes := eob.ToFHIR()
	if fhirRes["resourceType"] != "ExplanationOfBenefit" {
		t.Errorf("expected ExplanationOfBenefit, got %v", fhirRes["resourceType"])
	}
}
