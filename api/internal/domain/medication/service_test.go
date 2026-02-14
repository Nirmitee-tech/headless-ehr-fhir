package medication

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// -- Mock Repositories --

type mockMedRepo struct {
	meds        map[uuid.UUID]*Medication
	ingredients map[uuid.UUID]*MedicationIngredient
}

func newMockMedRepo() *mockMedRepo {
	return &mockMedRepo{
		meds:        make(map[uuid.UUID]*Medication),
		ingredients: make(map[uuid.UUID]*MedicationIngredient),
	}
}

func (m *mockMedRepo) Create(_ context.Context, med *Medication) error {
	med.ID = uuid.New()
	if med.FHIRID == "" {
		med.FHIRID = med.ID.String()
	}
	med.CreatedAt = time.Now()
	med.UpdatedAt = time.Now()
	m.meds[med.ID] = med
	return nil
}

func (m *mockMedRepo) GetByID(_ context.Context, id uuid.UUID) (*Medication, error) {
	med, ok := m.meds[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return med, nil
}

func (m *mockMedRepo) GetByFHIRID(_ context.Context, fhirID string) (*Medication, error) {
	for _, med := range m.meds {
		if med.FHIRID == fhirID {
			return med, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockMedRepo) Update(_ context.Context, med *Medication) error {
	if _, ok := m.meds[med.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.meds[med.ID] = med
	return nil
}

func (m *mockMedRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.meds, id)
	return nil
}

func (m *mockMedRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*Medication, int, error) {
	var result []*Medication
	for _, med := range m.meds {
		result = append(result, med)
	}
	total := len(result)
	if offset >= len(result) {
		return nil, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (m *mockMedRepo) AddIngredient(_ context.Context, ing *MedicationIngredient) error {
	ing.ID = uuid.New()
	m.ingredients[ing.ID] = ing
	return nil
}

func (m *mockMedRepo) GetIngredients(_ context.Context, medicationID uuid.UUID) ([]*MedicationIngredient, error) {
	var result []*MedicationIngredient
	for _, ing := range m.ingredients {
		if ing.MedicationID == medicationID {
			result = append(result, ing)
		}
	}
	return result, nil
}

func (m *mockMedRepo) RemoveIngredient(_ context.Context, id uuid.UUID) error {
	delete(m.ingredients, id)
	return nil
}

type mockMedRequestRepo struct {
	reqs map[uuid.UUID]*MedicationRequest
}

func newMockMedRequestRepo() *mockMedRequestRepo {
	return &mockMedRequestRepo{reqs: make(map[uuid.UUID]*MedicationRequest)}
}

func (m *mockMedRequestRepo) Create(_ context.Context, mr *MedicationRequest) error {
	mr.ID = uuid.New()
	if mr.FHIRID == "" {
		mr.FHIRID = mr.ID.String()
	}
	mr.CreatedAt = time.Now()
	mr.UpdatedAt = time.Now()
	m.reqs[mr.ID] = mr
	return nil
}

func (m *mockMedRequestRepo) GetByID(_ context.Context, id uuid.UUID) (*MedicationRequest, error) {
	mr, ok := m.reqs[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return mr, nil
}

func (m *mockMedRequestRepo) GetByFHIRID(_ context.Context, fhirID string) (*MedicationRequest, error) {
	for _, mr := range m.reqs {
		if mr.FHIRID == fhirID {
			return mr, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockMedRequestRepo) Update(_ context.Context, mr *MedicationRequest) error {
	m.reqs[mr.ID] = mr
	return nil
}

func (m *mockMedRequestRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.reqs, id)
	return nil
}

func (m *mockMedRequestRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationRequest, int, error) {
	var result []*MedicationRequest
	for _, mr := range m.reqs {
		if mr.PatientID == patientID {
			result = append(result, mr)
		}
	}
	return result, len(result), nil
}

func (m *mockMedRequestRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*MedicationRequest, int, error) {
	var result []*MedicationRequest
	for _, mr := range m.reqs {
		result = append(result, mr)
	}
	return result, len(result), nil
}

type mockMedAdminRepo struct {
	admins map[uuid.UUID]*MedicationAdministration
}

func newMockMedAdminRepo() *mockMedAdminRepo {
	return &mockMedAdminRepo{admins: make(map[uuid.UUID]*MedicationAdministration)}
}

func (m *mockMedAdminRepo) Create(_ context.Context, ma *MedicationAdministration) error {
	ma.ID = uuid.New()
	if ma.FHIRID == "" {
		ma.FHIRID = ma.ID.String()
	}
	ma.CreatedAt = time.Now()
	ma.UpdatedAt = time.Now()
	m.admins[ma.ID] = ma
	return nil
}

func (m *mockMedAdminRepo) GetByID(_ context.Context, id uuid.UUID) (*MedicationAdministration, error) {
	ma, ok := m.admins[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return ma, nil
}

func (m *mockMedAdminRepo) GetByFHIRID(_ context.Context, fhirID string) (*MedicationAdministration, error) {
	for _, ma := range m.admins {
		if ma.FHIRID == fhirID {
			return ma, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockMedAdminRepo) Update(_ context.Context, ma *MedicationAdministration) error {
	m.admins[ma.ID] = ma
	return nil
}

func (m *mockMedAdminRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.admins, id)
	return nil
}

func (m *mockMedAdminRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationAdministration, int, error) {
	var result []*MedicationAdministration
	for _, ma := range m.admins {
		if ma.PatientID == patientID {
			result = append(result, ma)
		}
	}
	return result, len(result), nil
}

func (m *mockMedAdminRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*MedicationAdministration, int, error) {
	var result []*MedicationAdministration
	for _, ma := range m.admins {
		result = append(result, ma)
	}
	return result, len(result), nil
}

type mockMedDispenseRepo struct {
	dispenses map[uuid.UUID]*MedicationDispense
}

func newMockMedDispenseRepo() *mockMedDispenseRepo {
	return &mockMedDispenseRepo{dispenses: make(map[uuid.UUID]*MedicationDispense)}
}

func (m *mockMedDispenseRepo) Create(_ context.Context, md *MedicationDispense) error {
	md.ID = uuid.New()
	if md.FHIRID == "" {
		md.FHIRID = md.ID.String()
	}
	md.CreatedAt = time.Now()
	md.UpdatedAt = time.Now()
	m.dispenses[md.ID] = md
	return nil
}

func (m *mockMedDispenseRepo) GetByID(_ context.Context, id uuid.UUID) (*MedicationDispense, error) {
	md, ok := m.dispenses[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return md, nil
}

func (m *mockMedDispenseRepo) GetByFHIRID(_ context.Context, fhirID string) (*MedicationDispense, error) {
	for _, md := range m.dispenses {
		if md.FHIRID == fhirID {
			return md, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockMedDispenseRepo) Update(_ context.Context, md *MedicationDispense) error {
	m.dispenses[md.ID] = md
	return nil
}

func (m *mockMedDispenseRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.dispenses, id)
	return nil
}

func (m *mockMedDispenseRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationDispense, int, error) {
	var result []*MedicationDispense
	for _, md := range m.dispenses {
		if md.PatientID == patientID {
			result = append(result, md)
		}
	}
	return result, len(result), nil
}

func (m *mockMedDispenseRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*MedicationDispense, int, error) {
	var result []*MedicationDispense
	for _, md := range m.dispenses {
		result = append(result, md)
	}
	return result, len(result), nil
}

type mockMedStatementRepo struct {
	stmts map[uuid.UUID]*MedicationStatement
}

func newMockMedStatementRepo() *mockMedStatementRepo {
	return &mockMedStatementRepo{stmts: make(map[uuid.UUID]*MedicationStatement)}
}

func (m *mockMedStatementRepo) Create(_ context.Context, ms *MedicationStatement) error {
	ms.ID = uuid.New()
	if ms.FHIRID == "" {
		ms.FHIRID = ms.ID.String()
	}
	ms.CreatedAt = time.Now()
	ms.UpdatedAt = time.Now()
	m.stmts[ms.ID] = ms
	return nil
}

func (m *mockMedStatementRepo) GetByID(_ context.Context, id uuid.UUID) (*MedicationStatement, error) {
	ms, ok := m.stmts[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return ms, nil
}

func (m *mockMedStatementRepo) GetByFHIRID(_ context.Context, fhirID string) (*MedicationStatement, error) {
	for _, ms := range m.stmts {
		if ms.FHIRID == fhirID {
			return ms, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockMedStatementRepo) Update(_ context.Context, ms *MedicationStatement) error {
	m.stmts[ms.ID] = ms
	return nil
}

func (m *mockMedStatementRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.stmts, id)
	return nil
}

func (m *mockMedStatementRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationStatement, int, error) {
	var result []*MedicationStatement
	for _, ms := range m.stmts {
		if ms.PatientID == patientID {
			result = append(result, ms)
		}
	}
	return result, len(result), nil
}

func (m *mockMedStatementRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*MedicationStatement, int, error) {
	var result []*MedicationStatement
	for _, ms := range m.stmts {
		result = append(result, ms)
	}
	return result, len(result), nil
}

// -- Tests --

func newTestService() *Service {
	return NewService(newMockMedRepo(), newMockMedRequestRepo(), newMockMedAdminRepo(), newMockMedDispenseRepo(), newMockMedStatementRepo())
}

func TestCreateMedication(t *testing.T) {
	svc := newTestService()
	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	err := svc.CreateMedication(context.Background(), m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if m.Status != "active" {
		t.Errorf("expected default status 'active', got %s", m.Status)
	}
}

func TestCreateMedication_CodeValueRequired(t *testing.T) {
	svc := newTestService()
	m := &Medication{CodeDisplay: "Aspirin"}
	err := svc.CreateMedication(context.Background(), m)
	if err == nil {
		t.Error("expected error for missing code_value")
	}
}

func TestCreateMedication_CodeDisplayRequired(t *testing.T) {
	svc := newTestService()
	m := &Medication{CodeValue: "12345"}
	err := svc.CreateMedication(context.Background(), m)
	if err == nil {
		t.Error("expected error for missing code_display")
	}
}

func TestCreateMedication_InvalidStatus(t *testing.T) {
	svc := newTestService()
	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin", Status: "bogus"}
	err := svc.CreateMedication(context.Background(), m)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestGetMedication(t *testing.T) {
	svc := newTestService()
	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	svc.CreateMedication(context.Background(), m)

	fetched, err := svc.GetMedication(context.Background(), m.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.CodeDisplay != "Aspirin" {
		t.Errorf("expected 'Aspirin', got %s", fetched.CodeDisplay)
	}
}

func TestDeleteMedication(t *testing.T) {
	svc := newTestService()
	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	svc.CreateMedication(context.Background(), m)

	err := svc.DeleteMedication(context.Background(), m.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetMedication(context.Background(), m.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestSearchMedications(t *testing.T) {
	svc := newTestService()
	svc.CreateMedication(context.Background(), &Medication{CodeValue: "1", CodeDisplay: "A"})
	svc.CreateMedication(context.Background(), &Medication{CodeValue: "2", CodeDisplay: "B"})

	items, total, err := svc.SearchMedications(context.Background(), nil, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestAddIngredient(t *testing.T) {
	svc := newTestService()
	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	svc.CreateMedication(context.Background(), m)

	ing := &MedicationIngredient{MedicationID: m.ID, ItemDisplay: "Acetylsalicylic acid"}
	err := svc.AddIngredient(context.Background(), ing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ings, err := svc.GetIngredients(context.Background(), m.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ings) != 1 {
		t.Fatalf("expected 1 ingredient, got %d", len(ings))
	}
}

func TestAddIngredient_MedicationIDRequired(t *testing.T) {
	svc := newTestService()
	ing := &MedicationIngredient{ItemDisplay: "Test"}
	err := svc.AddIngredient(context.Background(), ing)
	if err == nil {
		t.Error("expected error for missing medication_id")
	}
}

func TestAddIngredient_ItemDisplayRequired(t *testing.T) {
	svc := newTestService()
	ing := &MedicationIngredient{MedicationID: uuid.New()}
	err := svc.AddIngredient(context.Background(), ing)
	if err == nil {
		t.Error("expected error for missing item_display")
	}
}

func TestCreateMedicationRequest(t *testing.T) {
	svc := newTestService()
	mr := &MedicationRequest{
		PatientID:    uuid.New(),
		MedicationID: uuid.New(),
		RequesterID:  uuid.New(),
	}
	err := svc.CreateMedicationRequest(context.Background(), mr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mr.Status != "draft" {
		t.Errorf("expected default status 'draft', got %s", mr.Status)
	}
	if mr.Intent != "order" {
		t.Errorf("expected default intent 'order', got %s", mr.Intent)
	}
}

func TestCreateMedicationRequest_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	mr := &MedicationRequest{MedicationID: uuid.New(), RequesterID: uuid.New()}
	err := svc.CreateMedicationRequest(context.Background(), mr)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateMedicationRequest_MedicationIDRequired(t *testing.T) {
	svc := newTestService()
	mr := &MedicationRequest{PatientID: uuid.New(), RequesterID: uuid.New()}
	err := svc.CreateMedicationRequest(context.Background(), mr)
	if err == nil {
		t.Error("expected error for missing medication_id")
	}
}

func TestCreateMedicationRequest_RequesterIDRequired(t *testing.T) {
	svc := newTestService()
	mr := &MedicationRequest{PatientID: uuid.New(), MedicationID: uuid.New()}
	err := svc.CreateMedicationRequest(context.Background(), mr)
	if err == nil {
		t.Error("expected error for missing requester_id")
	}
}

func TestCreateMedicationAdministration(t *testing.T) {
	svc := newTestService()
	ma := &MedicationAdministration{PatientID: uuid.New(), MedicationID: uuid.New()}
	err := svc.CreateMedicationAdministration(context.Background(), ma)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ma.Status != "in-progress" {
		t.Errorf("expected default status 'in-progress', got %s", ma.Status)
	}
}

func TestCreateMedicationAdministration_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	ma := &MedicationAdministration{MedicationID: uuid.New()}
	err := svc.CreateMedicationAdministration(context.Background(), ma)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateMedicationDispense(t *testing.T) {
	svc := newTestService()
	md := &MedicationDispense{PatientID: uuid.New(), MedicationID: uuid.New()}
	err := svc.CreateMedicationDispense(context.Background(), md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if md.Status != "preparation" {
		t.Errorf("expected default status 'preparation', got %s", md.Status)
	}
}

func TestCreateMedicationDispense_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	md := &MedicationDispense{MedicationID: uuid.New()}
	err := svc.CreateMedicationDispense(context.Background(), md)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateMedicationStatement(t *testing.T) {
	svc := newTestService()
	ms := &MedicationStatement{PatientID: uuid.New()}
	err := svc.CreateMedicationStatement(context.Background(), ms)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ms.Status != "active" {
		t.Errorf("expected default status 'active', got %s", ms.Status)
	}
}

func TestCreateMedicationStatement_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	ms := &MedicationStatement{}
	err := svc.CreateMedicationStatement(context.Background(), ms)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestMedicationRequestToFHIR(t *testing.T) {
	mr := &MedicationRequest{
		FHIRID:       "mr-123",
		Status:       "active",
		Intent:       "order",
		PatientID:    uuid.New(),
		MedicationID: uuid.New(),
		RequesterID:  uuid.New(),
		UpdatedAt:    time.Now(),
	}
	fhirRes := mr.ToFHIR()
	if fhirRes["resourceType"] != "MedicationRequest" {
		t.Errorf("expected MedicationRequest, got %v", fhirRes["resourceType"])
	}
	if fhirRes["id"] != "mr-123" {
		t.Errorf("expected mr-123, got %v", fhirRes["id"])
	}
	if fhirRes["status"] != "active" {
		t.Errorf("expected active, got %v", fhirRes["status"])
	}
	if fhirRes["intent"] != "order" {
		t.Errorf("expected order, got %v", fhirRes["intent"])
	}
}

func TestMedicationAdministrationToFHIR(t *testing.T) {
	ma := &MedicationAdministration{
		FHIRID:       "ma-123",
		Status:       "completed",
		PatientID:    uuid.New(),
		MedicationID: uuid.New(),
		UpdatedAt:    time.Now(),
	}
	fhirRes := ma.ToFHIR()
	if fhirRes["resourceType"] != "MedicationAdministration" {
		t.Errorf("expected MedicationAdministration, got %v", fhirRes["resourceType"])
	}
	if fhirRes["id"] != "ma-123" {
		t.Errorf("expected ma-123, got %v", fhirRes["id"])
	}
}

func TestMedicationDispenseToFHIR(t *testing.T) {
	md := &MedicationDispense{
		FHIRID:       "md-123",
		Status:       "completed",
		PatientID:    uuid.New(),
		MedicationID: uuid.New(),
		UpdatedAt:    time.Now(),
	}
	fhirRes := md.ToFHIR()
	if fhirRes["resourceType"] != "MedicationDispense" {
		t.Errorf("expected MedicationDispense, got %v", fhirRes["resourceType"])
	}
	if fhirRes["id"] != "md-123" {
		t.Errorf("expected md-123, got %v", fhirRes["id"])
	}
}
