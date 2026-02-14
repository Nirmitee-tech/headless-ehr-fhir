package oncology

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// -- Mock Repositories --

type mockCancerDiagnosisRepo struct {
	records map[uuid.UUID]*CancerDiagnosis
}

func newMockCancerDiagnosisRepo() *mockCancerDiagnosisRepo {
	return &mockCancerDiagnosisRepo{records: make(map[uuid.UUID]*CancerDiagnosis)}
}

func (m *mockCancerDiagnosisRepo) Create(_ context.Context, d *CancerDiagnosis) error {
	d.ID = uuid.New()
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
	m.records[d.ID] = d
	return nil
}

func (m *mockCancerDiagnosisRepo) GetByID(_ context.Context, id uuid.UUID) (*CancerDiagnosis, error) {
	d, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return d, nil
}

func (m *mockCancerDiagnosisRepo) Update(_ context.Context, d *CancerDiagnosis) error {
	m.records[d.ID] = d
	return nil
}

func (m *mockCancerDiagnosisRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockCancerDiagnosisRepo) List(_ context.Context, limit, offset int) ([]*CancerDiagnosis, int, error) {
	var result []*CancerDiagnosis
	for _, d := range m.records {
		result = append(result, d)
	}
	return result, len(result), nil
}

func (m *mockCancerDiagnosisRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*CancerDiagnosis, int, error) {
	var result []*CancerDiagnosis
	for _, d := range m.records {
		if d.PatientID == patientID {
			result = append(result, d)
		}
	}
	return result, len(result), nil
}

type mockTreatmentProtocolRepo struct {
	records map[uuid.UUID]*TreatmentProtocol
	drugs   map[uuid.UUID]*TreatmentProtocolDrug
}

func newMockTreatmentProtocolRepo() *mockTreatmentProtocolRepo {
	return &mockTreatmentProtocolRepo{
		records: make(map[uuid.UUID]*TreatmentProtocol),
		drugs:   make(map[uuid.UUID]*TreatmentProtocolDrug),
	}
}

func (m *mockTreatmentProtocolRepo) Create(_ context.Context, p *TreatmentProtocol) error {
	p.ID = uuid.New()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	m.records[p.ID] = p
	return nil
}

func (m *mockTreatmentProtocolRepo) GetByID(_ context.Context, id uuid.UUID) (*TreatmentProtocol, error) {
	p, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return p, nil
}

func (m *mockTreatmentProtocolRepo) Update(_ context.Context, p *TreatmentProtocol) error {
	m.records[p.ID] = p
	return nil
}

func (m *mockTreatmentProtocolRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockTreatmentProtocolRepo) List(_ context.Context, limit, offset int) ([]*TreatmentProtocol, int, error) {
	var result []*TreatmentProtocol
	for _, p := range m.records {
		result = append(result, p)
	}
	return result, len(result), nil
}

func (m *mockTreatmentProtocolRepo) AddDrug(_ context.Context, d *TreatmentProtocolDrug) error {
	d.ID = uuid.New()
	m.drugs[d.ID] = d
	return nil
}

func (m *mockTreatmentProtocolRepo) GetDrugs(_ context.Context, protocolID uuid.UUID) ([]*TreatmentProtocolDrug, error) {
	var result []*TreatmentProtocolDrug
	for _, d := range m.drugs {
		if d.ProtocolID == protocolID {
			result = append(result, d)
		}
	}
	return result, nil
}

type mockChemoCycleRepo struct {
	records         map[uuid.UUID]*ChemoCycle
	administrations map[uuid.UUID]*ChemoAdministration
}

func newMockChemoCycleRepo() *mockChemoCycleRepo {
	return &mockChemoCycleRepo{
		records:         make(map[uuid.UUID]*ChemoCycle),
		administrations: make(map[uuid.UUID]*ChemoAdministration),
	}
}

func (m *mockChemoCycleRepo) Create(_ context.Context, c *ChemoCycle) error {
	c.ID = uuid.New()
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	m.records[c.ID] = c
	return nil
}

func (m *mockChemoCycleRepo) GetByID(_ context.Context, id uuid.UUID) (*ChemoCycle, error) {
	c, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

func (m *mockChemoCycleRepo) Update(_ context.Context, c *ChemoCycle) error {
	m.records[c.ID] = c
	return nil
}

func (m *mockChemoCycleRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockChemoCycleRepo) List(_ context.Context, limit, offset int) ([]*ChemoCycle, int, error) {
	var result []*ChemoCycle
	for _, c := range m.records {
		result = append(result, c)
	}
	return result, len(result), nil
}

func (m *mockChemoCycleRepo) AddAdministration(_ context.Context, a *ChemoAdministration) error {
	a.ID = uuid.New()
	m.administrations[a.ID] = a
	return nil
}

func (m *mockChemoCycleRepo) GetAdministrations(_ context.Context, cycleID uuid.UUID) ([]*ChemoAdministration, error) {
	var result []*ChemoAdministration
	for _, a := range m.administrations {
		if a.CycleID == cycleID {
			result = append(result, a)
		}
	}
	return result, nil
}

type mockRadiationTherapyRepo struct {
	records  map[uuid.UUID]*RadiationTherapy
	sessions map[uuid.UUID]*RadiationSession
}

func newMockRadiationTherapyRepo() *mockRadiationTherapyRepo {
	return &mockRadiationTherapyRepo{
		records:  make(map[uuid.UUID]*RadiationTherapy),
		sessions: make(map[uuid.UUID]*RadiationSession),
	}
}

func (m *mockRadiationTherapyRepo) Create(_ context.Context, r *RadiationTherapy) error {
	r.ID = uuid.New()
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	m.records[r.ID] = r
	return nil
}

func (m *mockRadiationTherapyRepo) GetByID(_ context.Context, id uuid.UUID) (*RadiationTherapy, error) {
	r, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return r, nil
}

func (m *mockRadiationTherapyRepo) Update(_ context.Context, r *RadiationTherapy) error {
	m.records[r.ID] = r
	return nil
}

func (m *mockRadiationTherapyRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockRadiationTherapyRepo) List(_ context.Context, limit, offset int) ([]*RadiationTherapy, int, error) {
	var result []*RadiationTherapy
	for _, r := range m.records {
		result = append(result, r)
	}
	return result, len(result), nil
}

func (m *mockRadiationTherapyRepo) AddSession(_ context.Context, s *RadiationSession) error {
	s.ID = uuid.New()
	m.sessions[s.ID] = s
	return nil
}

func (m *mockRadiationTherapyRepo) GetSessions(_ context.Context, radiationID uuid.UUID) ([]*RadiationSession, error) {
	var result []*RadiationSession
	for _, s := range m.sessions {
		if s.RadiationTherapyID == radiationID {
			result = append(result, s)
		}
	}
	return result, nil
}

type mockTumorMarkerRepo struct {
	records map[uuid.UUID]*TumorMarker
}

func newMockTumorMarkerRepo() *mockTumorMarkerRepo {
	return &mockTumorMarkerRepo{records: make(map[uuid.UUID]*TumorMarker)}
}

func (m *mockTumorMarkerRepo) Create(_ context.Context, mk *TumorMarker) error {
	mk.ID = uuid.New()
	mk.CreatedAt = time.Now()
	mk.UpdatedAt = time.Now()
	m.records[mk.ID] = mk
	return nil
}

func (m *mockTumorMarkerRepo) GetByID(_ context.Context, id uuid.UUID) (*TumorMarker, error) {
	mk, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return mk, nil
}

func (m *mockTumorMarkerRepo) Update(_ context.Context, mk *TumorMarker) error {
	m.records[mk.ID] = mk
	return nil
}

func (m *mockTumorMarkerRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockTumorMarkerRepo) List(_ context.Context, limit, offset int) ([]*TumorMarker, int, error) {
	var result []*TumorMarker
	for _, mk := range m.records {
		result = append(result, mk)
	}
	return result, len(result), nil
}

type mockTumorBoardRepo struct {
	records map[uuid.UUID]*TumorBoardReview
}

func newMockTumorBoardRepo() *mockTumorBoardRepo {
	return &mockTumorBoardRepo{records: make(map[uuid.UUID]*TumorBoardReview)}
}

func (m *mockTumorBoardRepo) Create(_ context.Context, r *TumorBoardReview) error {
	r.ID = uuid.New()
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	m.records[r.ID] = r
	return nil
}

func (m *mockTumorBoardRepo) GetByID(_ context.Context, id uuid.UUID) (*TumorBoardReview, error) {
	r, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return r, nil
}

func (m *mockTumorBoardRepo) Update(_ context.Context, r *TumorBoardReview) error {
	m.records[r.ID] = r
	return nil
}

func (m *mockTumorBoardRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockTumorBoardRepo) List(_ context.Context, limit, offset int) ([]*TumorBoardReview, int, error) {
	var result []*TumorBoardReview
	for _, r := range m.records {
		result = append(result, r)
	}
	return result, len(result), nil
}

// -- Tests --

func newTestService() *Service {
	return NewService(
		newMockCancerDiagnosisRepo(),
		newMockTreatmentProtocolRepo(),
		newMockChemoCycleRepo(),
		newMockRadiationTherapyRepo(),
		newMockTumorMarkerRepo(),
		newMockTumorBoardRepo(),
	)
}

// -- Cancer Diagnosis Tests --

func TestCreateCancerDiagnosis(t *testing.T) {
	svc := newTestService()
	d := &CancerDiagnosis{PatientID: uuid.New(), DiagnosisDate: time.Now()}
	err := svc.CreateCancerDiagnosis(context.Background(), d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if d.CurrentStatus != "active-treatment" {
		t.Errorf("expected default status 'active-treatment', got %s", d.CurrentStatus)
	}
}

func TestCreateCancerDiagnosis_PatientRequired(t *testing.T) {
	svc := newTestService()
	d := &CancerDiagnosis{DiagnosisDate: time.Now()}
	err := svc.CreateCancerDiagnosis(context.Background(), d)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateCancerDiagnosis_DateRequired(t *testing.T) {
	svc := newTestService()
	d := &CancerDiagnosis{PatientID: uuid.New()}
	err := svc.CreateCancerDiagnosis(context.Background(), d)
	if err == nil {
		t.Error("expected error for missing diagnosis_date")
	}
}

func TestCreateCancerDiagnosis_InvalidStatus(t *testing.T) {
	svc := newTestService()
	d := &CancerDiagnosis{PatientID: uuid.New(), DiagnosisDate: time.Now(), CurrentStatus: "invalid"}
	err := svc.CreateCancerDiagnosis(context.Background(), d)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestGetCancerDiagnosis(t *testing.T) {
	svc := newTestService()
	d := &CancerDiagnosis{PatientID: uuid.New(), DiagnosisDate: time.Now()}
	svc.CreateCancerDiagnosis(context.Background(), d)

	fetched, err := svc.GetCancerDiagnosis(context.Background(), d.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.PatientID != d.PatientID {
		t.Error("patient_id mismatch")
	}
}

func TestDeleteCancerDiagnosis(t *testing.T) {
	svc := newTestService()
	d := &CancerDiagnosis{PatientID: uuid.New(), DiagnosisDate: time.Now()}
	svc.CreateCancerDiagnosis(context.Background(), d)
	err := svc.DeleteCancerDiagnosis(context.Background(), d.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetCancerDiagnosis(context.Background(), d.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

// -- Treatment Protocol Tests --

func TestCreateTreatmentProtocol(t *testing.T) {
	svc := newTestService()
	p := &TreatmentProtocol{CancerDiagnosisID: uuid.New(), ProtocolName: "FOLFOX"}
	err := svc.CreateTreatmentProtocol(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Status != "planned" {
		t.Errorf("expected default status 'planned', got %s", p.Status)
	}
}

func TestCreateTreatmentProtocol_DiagnosisRequired(t *testing.T) {
	svc := newTestService()
	p := &TreatmentProtocol{ProtocolName: "FOLFOX"}
	err := svc.CreateTreatmentProtocol(context.Background(), p)
	if err == nil {
		t.Error("expected error for missing cancer_diagnosis_id")
	}
}

func TestCreateTreatmentProtocol_NameRequired(t *testing.T) {
	svc := newTestService()
	p := &TreatmentProtocol{CancerDiagnosisID: uuid.New()}
	err := svc.CreateTreatmentProtocol(context.Background(), p)
	if err == nil {
		t.Error("expected error for missing protocol_name")
	}
}

func TestAddProtocolDrug(t *testing.T) {
	svc := newTestService()
	d := &TreatmentProtocolDrug{ProtocolID: uuid.New(), DrugName: "Oxaliplatin"}
	err := svc.AddProtocolDrug(context.Background(), d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestAddProtocolDrug_ProtocolRequired(t *testing.T) {
	svc := newTestService()
	d := &TreatmentProtocolDrug{DrugName: "Oxaliplatin"}
	err := svc.AddProtocolDrug(context.Background(), d)
	if err == nil {
		t.Error("expected error for missing protocol_id")
	}
}

func TestAddProtocolDrug_DrugNameRequired(t *testing.T) {
	svc := newTestService()
	d := &TreatmentProtocolDrug{ProtocolID: uuid.New()}
	err := svc.AddProtocolDrug(context.Background(), d)
	if err == nil {
		t.Error("expected error for missing drug_name")
	}
}

// -- Chemo Cycle Tests --

func TestCreateChemoCycle(t *testing.T) {
	svc := newTestService()
	c := &ChemoCycle{ProtocolID: uuid.New(), CycleNumber: 1}
	err := svc.CreateChemoCycle(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Status != "planned" {
		t.Errorf("expected default status 'planned', got %s", c.Status)
	}
}

func TestCreateChemoCycle_ProtocolRequired(t *testing.T) {
	svc := newTestService()
	c := &ChemoCycle{CycleNumber: 1}
	err := svc.CreateChemoCycle(context.Background(), c)
	if err == nil {
		t.Error("expected error for missing protocol_id")
	}
}

func TestCreateChemoCycle_CycleNumberRequired(t *testing.T) {
	svc := newTestService()
	c := &ChemoCycle{ProtocolID: uuid.New(), CycleNumber: 0}
	err := svc.CreateChemoCycle(context.Background(), c)
	if err == nil {
		t.Error("expected error for invalid cycle_number")
	}
}

func TestCreateChemoCycle_InvalidStatus(t *testing.T) {
	svc := newTestService()
	c := &ChemoCycle{ProtocolID: uuid.New(), CycleNumber: 1, Status: "bogus"}
	err := svc.CreateChemoCycle(context.Background(), c)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestAddChemoAdministration(t *testing.T) {
	svc := newTestService()
	a := &ChemoAdministration{CycleID: uuid.New(), DrugName: "Cisplatin"}
	err := svc.AddChemoAdministration(context.Background(), a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.AdministrationDatetime.IsZero() {
		t.Error("expected administration_datetime to be defaulted")
	}
}

func TestAddChemoAdministration_CycleRequired(t *testing.T) {
	svc := newTestService()
	a := &ChemoAdministration{DrugName: "Cisplatin"}
	err := svc.AddChemoAdministration(context.Background(), a)
	if err == nil {
		t.Error("expected error for missing cycle_id")
	}
}

func TestAddChemoAdministration_DrugNameRequired(t *testing.T) {
	svc := newTestService()
	a := &ChemoAdministration{CycleID: uuid.New()}
	err := svc.AddChemoAdministration(context.Background(), a)
	if err == nil {
		t.Error("expected error for missing drug_name")
	}
}

// -- Radiation Therapy Tests --

func TestCreateRadiationTherapy(t *testing.T) {
	svc := newTestService()
	r := &RadiationTherapy{CancerDiagnosisID: uuid.New()}
	err := svc.CreateRadiationTherapy(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Status != "planned" {
		t.Errorf("expected default status 'planned', got %s", r.Status)
	}
}

func TestCreateRadiationTherapy_DiagnosisRequired(t *testing.T) {
	svc := newTestService()
	r := &RadiationTherapy{}
	err := svc.CreateRadiationTherapy(context.Background(), r)
	if err == nil {
		t.Error("expected error for missing cancer_diagnosis_id")
	}
}

func TestCreateRadiationTherapy_InvalidStatus(t *testing.T) {
	svc := newTestService()
	r := &RadiationTherapy{CancerDiagnosisID: uuid.New(), Status: "bogus"}
	err := svc.CreateRadiationTherapy(context.Background(), r)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestAddRadiationSession(t *testing.T) {
	svc := newTestService()
	s := &RadiationSession{RadiationTherapyID: uuid.New(), SessionNumber: 1}
	err := svc.AddRadiationSession(context.Background(), s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.SessionDate.IsZero() {
		t.Error("expected session_date to be defaulted")
	}
}

func TestAddRadiationSession_RadiationRequired(t *testing.T) {
	svc := newTestService()
	s := &RadiationSession{SessionNumber: 1}
	err := svc.AddRadiationSession(context.Background(), s)
	if err == nil {
		t.Error("expected error for missing radiation_therapy_id")
	}
}

func TestAddRadiationSession_SessionNumberRequired(t *testing.T) {
	svc := newTestService()
	s := &RadiationSession{RadiationTherapyID: uuid.New(), SessionNumber: 0}
	err := svc.AddRadiationSession(context.Background(), s)
	if err == nil {
		t.Error("expected error for invalid session_number")
	}
}

// -- Tumor Marker Tests --

func TestCreateTumorMarker(t *testing.T) {
	svc := newTestService()
	m := &TumorMarker{PatientID: uuid.New(), MarkerName: "PSA"}
	err := svc.CreateTumorMarker(context.Background(), m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestCreateTumorMarker_PatientRequired(t *testing.T) {
	svc := newTestService()
	m := &TumorMarker{MarkerName: "PSA"}
	err := svc.CreateTumorMarker(context.Background(), m)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateTumorMarker_NameRequired(t *testing.T) {
	svc := newTestService()
	m := &TumorMarker{PatientID: uuid.New()}
	err := svc.CreateTumorMarker(context.Background(), m)
	if err == nil {
		t.Error("expected error for missing marker_name")
	}
}

func TestDeleteTumorMarker(t *testing.T) {
	svc := newTestService()
	m := &TumorMarker{PatientID: uuid.New(), MarkerName: "PSA"}
	svc.CreateTumorMarker(context.Background(), m)
	err := svc.DeleteTumorMarker(context.Background(), m.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetTumorMarker(context.Background(), m.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

// -- Tumor Board Review Tests --

func TestCreateTumorBoardReview(t *testing.T) {
	svc := newTestService()
	r := &TumorBoardReview{CancerDiagnosisID: uuid.New(), PatientID: uuid.New()}
	err := svc.CreateTumorBoardReview(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ReviewDate.IsZero() {
		t.Error("expected review_date to be defaulted")
	}
}

func TestCreateTumorBoardReview_DiagnosisRequired(t *testing.T) {
	svc := newTestService()
	r := &TumorBoardReview{PatientID: uuid.New()}
	err := svc.CreateTumorBoardReview(context.Background(), r)
	if err == nil {
		t.Error("expected error for missing cancer_diagnosis_id")
	}
}

func TestCreateTumorBoardReview_PatientRequired(t *testing.T) {
	svc := newTestService()
	r := &TumorBoardReview{CancerDiagnosisID: uuid.New()}
	err := svc.CreateTumorBoardReview(context.Background(), r)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestGetTumorBoardReview(t *testing.T) {
	svc := newTestService()
	r := &TumorBoardReview{CancerDiagnosisID: uuid.New(), PatientID: uuid.New()}
	svc.CreateTumorBoardReview(context.Background(), r)

	fetched, err := svc.GetTumorBoardReview(context.Background(), r.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.PatientID != r.PatientID {
		t.Error("patient_id mismatch")
	}
}

func TestDeleteTumorBoardReview(t *testing.T) {
	svc := newTestService()
	r := &TumorBoardReview{CancerDiagnosisID: uuid.New(), PatientID: uuid.New()}
	svc.CreateTumorBoardReview(context.Background(), r)
	err := svc.DeleteTumorBoardReview(context.Background(), r.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetTumorBoardReview(context.Background(), r.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}
