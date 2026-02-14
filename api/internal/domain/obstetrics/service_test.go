package obstetrics

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// -- Mock Repositories --

type mockPregnancyRepo struct {
	records map[uuid.UUID]*Pregnancy
}

func newMockPregnancyRepo() *mockPregnancyRepo {
	return &mockPregnancyRepo{records: make(map[uuid.UUID]*Pregnancy)}
}

func (m *mockPregnancyRepo) Create(_ context.Context, p *Pregnancy) error {
	p.ID = uuid.New()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	m.records[p.ID] = p
	return nil
}

func (m *mockPregnancyRepo) GetByID(_ context.Context, id uuid.UUID) (*Pregnancy, error) {
	p, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return p, nil
}

func (m *mockPregnancyRepo) Update(_ context.Context, p *Pregnancy) error {
	m.records[p.ID] = p
	return nil
}

func (m *mockPregnancyRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockPregnancyRepo) List(_ context.Context, limit, offset int) ([]*Pregnancy, int, error) {
	var result []*Pregnancy
	for _, p := range m.records {
		result = append(result, p)
	}
	return result, len(result), nil
}

func (m *mockPregnancyRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*Pregnancy, int, error) {
	var result []*Pregnancy
	for _, p := range m.records {
		if p.PatientID == patientID {
			result = append(result, p)
		}
	}
	return result, len(result), nil
}

type mockPrenatalVisitRepo struct {
	records map[uuid.UUID]*PrenatalVisit
}

func newMockPrenatalVisitRepo() *mockPrenatalVisitRepo {
	return &mockPrenatalVisitRepo{records: make(map[uuid.UUID]*PrenatalVisit)}
}

func (m *mockPrenatalVisitRepo) Create(_ context.Context, v *PrenatalVisit) error {
	v.ID = uuid.New()
	v.CreatedAt = time.Now()
	v.UpdatedAt = time.Now()
	m.records[v.ID] = v
	return nil
}

func (m *mockPrenatalVisitRepo) GetByID(_ context.Context, id uuid.UUID) (*PrenatalVisit, error) {
	v, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return v, nil
}

func (m *mockPrenatalVisitRepo) Update(_ context.Context, v *PrenatalVisit) error {
	m.records[v.ID] = v
	return nil
}

func (m *mockPrenatalVisitRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockPrenatalVisitRepo) ListByPregnancy(_ context.Context, pregnancyID uuid.UUID, limit, offset int) ([]*PrenatalVisit, int, error) {
	var result []*PrenatalVisit
	for _, v := range m.records {
		if v.PregnancyID == pregnancyID {
			result = append(result, v)
		}
	}
	return result, len(result), nil
}

type mockLaborRepo struct {
	records        map[uuid.UUID]*LaborRecord
	cervicalExams  map[uuid.UUID]*LaborCervicalExam
	fetalMonitors  map[uuid.UUID]*FetalMonitoring
}

func newMockLaborRepo() *mockLaborRepo {
	return &mockLaborRepo{
		records:       make(map[uuid.UUID]*LaborRecord),
		cervicalExams: make(map[uuid.UUID]*LaborCervicalExam),
		fetalMonitors: make(map[uuid.UUID]*FetalMonitoring),
	}
}

func (m *mockLaborRepo) Create(_ context.Context, l *LaborRecord) error {
	l.ID = uuid.New()
	l.CreatedAt = time.Now()
	l.UpdatedAt = time.Now()
	m.records[l.ID] = l
	return nil
}

func (m *mockLaborRepo) GetByID(_ context.Context, id uuid.UUID) (*LaborRecord, error) {
	l, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return l, nil
}

func (m *mockLaborRepo) Update(_ context.Context, l *LaborRecord) error {
	m.records[l.ID] = l
	return nil
}

func (m *mockLaborRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockLaborRepo) List(_ context.Context, limit, offset int) ([]*LaborRecord, int, error) {
	var result []*LaborRecord
	for _, l := range m.records {
		result = append(result, l)
	}
	return result, len(result), nil
}

func (m *mockLaborRepo) AddCervicalExam(_ context.Context, e *LaborCervicalExam) error {
	e.ID = uuid.New()
	m.cervicalExams[e.ID] = e
	return nil
}

func (m *mockLaborRepo) GetCervicalExams(_ context.Context, laborRecordID uuid.UUID) ([]*LaborCervicalExam, error) {
	var result []*LaborCervicalExam
	for _, e := range m.cervicalExams {
		if e.LaborRecordID == laborRecordID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *mockLaborRepo) AddFetalMonitoring(_ context.Context, f *FetalMonitoring) error {
	f.ID = uuid.New()
	m.fetalMonitors[f.ID] = f
	return nil
}

func (m *mockLaborRepo) GetFetalMonitoring(_ context.Context, laborRecordID uuid.UUID) ([]*FetalMonitoring, error) {
	var result []*FetalMonitoring
	for _, f := range m.fetalMonitors {
		if f.LaborRecordID == laborRecordID {
			result = append(result, f)
		}
	}
	return result, nil
}

type mockDeliveryRepo struct {
	records map[uuid.UUID]*DeliveryRecord
}

func newMockDeliveryRepo() *mockDeliveryRepo {
	return &mockDeliveryRepo{records: make(map[uuid.UUID]*DeliveryRecord)}
}

func (m *mockDeliveryRepo) Create(_ context.Context, d *DeliveryRecord) error {
	d.ID = uuid.New()
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
	m.records[d.ID] = d
	return nil
}

func (m *mockDeliveryRepo) GetByID(_ context.Context, id uuid.UUID) (*DeliveryRecord, error) {
	d, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return d, nil
}

func (m *mockDeliveryRepo) Update(_ context.Context, d *DeliveryRecord) error {
	m.records[d.ID] = d
	return nil
}

func (m *mockDeliveryRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockDeliveryRepo) ListByPregnancy(_ context.Context, pregnancyID uuid.UUID, limit, offset int) ([]*DeliveryRecord, int, error) {
	var result []*DeliveryRecord
	for _, d := range m.records {
		if d.PregnancyID == pregnancyID {
			result = append(result, d)
		}
	}
	return result, len(result), nil
}

type mockNewbornRepo struct {
	records map[uuid.UUID]*NewbornRecord
}

func newMockNewbornRepo() *mockNewbornRepo {
	return &mockNewbornRepo{records: make(map[uuid.UUID]*NewbornRecord)}
}

func (m *mockNewbornRepo) Create(_ context.Context, n *NewbornRecord) error {
	n.ID = uuid.New()
	n.CreatedAt = time.Now()
	n.UpdatedAt = time.Now()
	m.records[n.ID] = n
	return nil
}

func (m *mockNewbornRepo) GetByID(_ context.Context, id uuid.UUID) (*NewbornRecord, error) {
	n, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return n, nil
}

func (m *mockNewbornRepo) Update(_ context.Context, n *NewbornRecord) error {
	m.records[n.ID] = n
	return nil
}

func (m *mockNewbornRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockNewbornRepo) List(_ context.Context, limit, offset int) ([]*NewbornRecord, int, error) {
	var result []*NewbornRecord
	for _, n := range m.records {
		result = append(result, n)
	}
	return result, len(result), nil
}

type mockPostpartumRepo struct {
	records map[uuid.UUID]*PostpartumRecord
}

func newMockPostpartumRepo() *mockPostpartumRepo {
	return &mockPostpartumRepo{records: make(map[uuid.UUID]*PostpartumRecord)}
}

func (m *mockPostpartumRepo) Create(_ context.Context, p *PostpartumRecord) error {
	p.ID = uuid.New()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	m.records[p.ID] = p
	return nil
}

func (m *mockPostpartumRepo) GetByID(_ context.Context, id uuid.UUID) (*PostpartumRecord, error) {
	p, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return p, nil
}

func (m *mockPostpartumRepo) Update(_ context.Context, p *PostpartumRecord) error {
	m.records[p.ID] = p
	return nil
}

func (m *mockPostpartumRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockPostpartumRepo) List(_ context.Context, limit, offset int) ([]*PostpartumRecord, int, error) {
	var result []*PostpartumRecord
	for _, p := range m.records {
		result = append(result, p)
	}
	return result, len(result), nil
}

// -- Tests --

func newTestService() *Service {
	return NewService(
		newMockPregnancyRepo(),
		newMockPrenatalVisitRepo(),
		newMockLaborRepo(),
		newMockDeliveryRepo(),
		newMockNewbornRepo(),
		newMockPostpartumRepo(),
	)
}

// -- Pregnancy Tests --

func TestCreatePregnancy(t *testing.T) {
	svc := newTestService()
	p := &Pregnancy{PatientID: uuid.New()}
	err := svc.CreatePregnancy(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if p.Status != "active" {
		t.Errorf("expected default status 'active', got %s", p.Status)
	}
}

func TestCreatePregnancy_PatientRequired(t *testing.T) {
	svc := newTestService()
	p := &Pregnancy{}
	err := svc.CreatePregnancy(context.Background(), p)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreatePregnancy_InvalidStatus(t *testing.T) {
	svc := newTestService()
	p := &Pregnancy{PatientID: uuid.New(), Status: "invalid-status"}
	err := svc.CreatePregnancy(context.Background(), p)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestGetPregnancy(t *testing.T) {
	svc := newTestService()
	p := &Pregnancy{PatientID: uuid.New()}
	svc.CreatePregnancy(context.Background(), p)

	fetched, err := svc.GetPregnancy(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.PatientID != p.PatientID {
		t.Error("patient_id mismatch")
	}
}

func TestDeletePregnancy(t *testing.T) {
	svc := newTestService()
	p := &Pregnancy{PatientID: uuid.New()}
	svc.CreatePregnancy(context.Background(), p)
	err := svc.DeletePregnancy(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetPregnancy(context.Background(), p.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestListPregnancies(t *testing.T) {
	svc := newTestService()
	svc.CreatePregnancy(context.Background(), &Pregnancy{PatientID: uuid.New()})
	svc.CreatePregnancy(context.Background(), &Pregnancy{PatientID: uuid.New()})
	items, total, err := svc.ListPregnancies(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2 items, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

// -- Prenatal Visit Tests --

func TestCreatePrenatalVisit(t *testing.T) {
	svc := newTestService()
	v := &PrenatalVisit{PregnancyID: uuid.New()}
	err := svc.CreatePrenatalVisit(context.Background(), v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if v.VisitDate.IsZero() {
		t.Error("expected visit_date to be defaulted")
	}
}

func TestCreatePrenatalVisit_PregnancyRequired(t *testing.T) {
	svc := newTestService()
	v := &PrenatalVisit{}
	err := svc.CreatePrenatalVisit(context.Background(), v)
	if err == nil {
		t.Error("expected error for missing pregnancy_id")
	}
}

func TestGetPrenatalVisit(t *testing.T) {
	svc := newTestService()
	v := &PrenatalVisit{PregnancyID: uuid.New()}
	svc.CreatePrenatalVisit(context.Background(), v)

	fetched, err := svc.GetPrenatalVisit(context.Background(), v.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.PregnancyID != v.PregnancyID {
		t.Error("pregnancy_id mismatch")
	}
}

func TestDeletePrenatalVisit(t *testing.T) {
	svc := newTestService()
	v := &PrenatalVisit{PregnancyID: uuid.New()}
	svc.CreatePrenatalVisit(context.Background(), v)
	err := svc.DeletePrenatalVisit(context.Background(), v.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetPrenatalVisit(context.Background(), v.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

// -- Labor Record Tests --

func TestCreateLaborRecord(t *testing.T) {
	svc := newTestService()
	l := &LaborRecord{PregnancyID: uuid.New()}
	err := svc.CreateLaborRecord(context.Background(), l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if l.Status != "active" {
		t.Errorf("expected default status 'active', got %s", l.Status)
	}
}

func TestCreateLaborRecord_PregnancyRequired(t *testing.T) {
	svc := newTestService()
	l := &LaborRecord{}
	err := svc.CreateLaborRecord(context.Background(), l)
	if err == nil {
		t.Error("expected error for missing pregnancy_id")
	}
}

func TestGetLaborRecord(t *testing.T) {
	svc := newTestService()
	l := &LaborRecord{PregnancyID: uuid.New()}
	svc.CreateLaborRecord(context.Background(), l)

	fetched, err := svc.GetLaborRecord(context.Background(), l.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.PregnancyID != l.PregnancyID {
		t.Error("pregnancy_id mismatch")
	}
}

func TestDeleteLaborRecord(t *testing.T) {
	svc := newTestService()
	l := &LaborRecord{PregnancyID: uuid.New()}
	svc.CreateLaborRecord(context.Background(), l)
	err := svc.DeleteLaborRecord(context.Background(), l.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetLaborRecord(context.Background(), l.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

// -- Cervical Exam Tests --

func TestAddCervicalExam(t *testing.T) {
	svc := newTestService()
	laborID := uuid.New()
	e := &LaborCervicalExam{LaborRecordID: laborID}
	err := svc.AddCervicalExam(context.Background(), e)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.ExamDatetime.IsZero() {
		t.Error("expected exam_datetime to be defaulted")
	}
}

func TestAddCervicalExam_LaborRecordRequired(t *testing.T) {
	svc := newTestService()
	e := &LaborCervicalExam{}
	err := svc.AddCervicalExam(context.Background(), e)
	if err == nil {
		t.Error("expected error for missing labor_record_id")
	}
}

// -- Fetal Monitoring Tests --

func TestAddFetalMonitoring(t *testing.T) {
	svc := newTestService()
	laborID := uuid.New()
	f := &FetalMonitoring{LaborRecordID: laborID}
	err := svc.AddFetalMonitoring(context.Background(), f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.MonitoringDatetime.IsZero() {
		t.Error("expected monitoring_datetime to be defaulted")
	}
}

func TestAddFetalMonitoring_LaborRecordRequired(t *testing.T) {
	svc := newTestService()
	f := &FetalMonitoring{}
	err := svc.AddFetalMonitoring(context.Background(), f)
	if err == nil {
		t.Error("expected error for missing labor_record_id")
	}
}

// -- Delivery Tests --

func TestCreateDelivery(t *testing.T) {
	svc := newTestService()
	d := &DeliveryRecord{
		PregnancyID:          uuid.New(),
		PatientID:            uuid.New(),
		DeliveryDatetime:     time.Now(),
		DeliveryMethod:       "vaginal",
		DeliveringProviderID: uuid.New(),
	}
	err := svc.CreateDelivery(context.Background(), d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestCreateDelivery_PregnancyRequired(t *testing.T) {
	svc := newTestService()
	d := &DeliveryRecord{
		PatientID:            uuid.New(),
		DeliveryDatetime:     time.Now(),
		DeliveryMethod:       "vaginal",
		DeliveringProviderID: uuid.New(),
	}
	err := svc.CreateDelivery(context.Background(), d)
	if err == nil {
		t.Error("expected error for missing pregnancy_id")
	}
}

func TestCreateDelivery_PatientRequired(t *testing.T) {
	svc := newTestService()
	d := &DeliveryRecord{
		PregnancyID:          uuid.New(),
		DeliveryDatetime:     time.Now(),
		DeliveryMethod:       "vaginal",
		DeliveringProviderID: uuid.New(),
	}
	err := svc.CreateDelivery(context.Background(), d)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateDelivery_DatetimeRequired(t *testing.T) {
	svc := newTestService()
	d := &DeliveryRecord{
		PregnancyID:          uuid.New(),
		PatientID:            uuid.New(),
		DeliveryMethod:       "vaginal",
		DeliveringProviderID: uuid.New(),
	}
	err := svc.CreateDelivery(context.Background(), d)
	if err == nil {
		t.Error("expected error for missing delivery_datetime")
	}
}

func TestCreateDelivery_MethodRequired(t *testing.T) {
	svc := newTestService()
	d := &DeliveryRecord{
		PregnancyID:          uuid.New(),
		PatientID:            uuid.New(),
		DeliveryDatetime:     time.Now(),
		DeliveringProviderID: uuid.New(),
	}
	err := svc.CreateDelivery(context.Background(), d)
	if err == nil {
		t.Error("expected error for missing delivery_method")
	}
}

func TestCreateDelivery_ProviderRequired(t *testing.T) {
	svc := newTestService()
	d := &DeliveryRecord{
		PregnancyID:      uuid.New(),
		PatientID:        uuid.New(),
		DeliveryDatetime: time.Now(),
		DeliveryMethod:   "vaginal",
	}
	err := svc.CreateDelivery(context.Background(), d)
	if err == nil {
		t.Error("expected error for missing delivering_provider_id")
	}
}

func TestGetDelivery(t *testing.T) {
	svc := newTestService()
	d := &DeliveryRecord{
		PregnancyID:          uuid.New(),
		PatientID:            uuid.New(),
		DeliveryDatetime:     time.Now(),
		DeliveryMethod:       "vaginal",
		DeliveringProviderID: uuid.New(),
	}
	svc.CreateDelivery(context.Background(), d)

	fetched, err := svc.GetDelivery(context.Background(), d.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.DeliveryMethod != "vaginal" {
		t.Errorf("expected delivery_method 'vaginal', got %s", fetched.DeliveryMethod)
	}
}

// -- Newborn Tests --

func TestCreateNewborn(t *testing.T) {
	svc := newTestService()
	n := &NewbornRecord{DeliveryID: uuid.New(), BirthDatetime: time.Now()}
	err := svc.CreateNewborn(context.Background(), n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestCreateNewborn_DeliveryRequired(t *testing.T) {
	svc := newTestService()
	n := &NewbornRecord{BirthDatetime: time.Now()}
	err := svc.CreateNewborn(context.Background(), n)
	if err == nil {
		t.Error("expected error for missing delivery_id")
	}
}

func TestCreateNewborn_BirthDatetimeRequired(t *testing.T) {
	svc := newTestService()
	n := &NewbornRecord{DeliveryID: uuid.New()}
	err := svc.CreateNewborn(context.Background(), n)
	if err == nil {
		t.Error("expected error for missing birth_datetime")
	}
}

func TestGetNewborn(t *testing.T) {
	svc := newTestService()
	n := &NewbornRecord{DeliveryID: uuid.New(), BirthDatetime: time.Now()}
	svc.CreateNewborn(context.Background(), n)

	fetched, err := svc.GetNewborn(context.Background(), n.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.DeliveryID != n.DeliveryID {
		t.Error("delivery_id mismatch")
	}
}

func TestDeleteNewborn(t *testing.T) {
	svc := newTestService()
	n := &NewbornRecord{DeliveryID: uuid.New(), BirthDatetime: time.Now()}
	svc.CreateNewborn(context.Background(), n)
	err := svc.DeleteNewborn(context.Background(), n.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetNewborn(context.Background(), n.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

// -- Postpartum Tests --

func TestCreatePostpartum(t *testing.T) {
	svc := newTestService()
	p := &PostpartumRecord{PregnancyID: uuid.New(), PatientID: uuid.New()}
	err := svc.CreatePostpartum(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if p.VisitDate.IsZero() {
		t.Error("expected visit_date to be defaulted")
	}
}

func TestCreatePostpartum_PregnancyRequired(t *testing.T) {
	svc := newTestService()
	p := &PostpartumRecord{PatientID: uuid.New()}
	err := svc.CreatePostpartum(context.Background(), p)
	if err == nil {
		t.Error("expected error for missing pregnancy_id")
	}
}

func TestCreatePostpartum_PatientRequired(t *testing.T) {
	svc := newTestService()
	p := &PostpartumRecord{PregnancyID: uuid.New()}
	err := svc.CreatePostpartum(context.Background(), p)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestGetPostpartum(t *testing.T) {
	svc := newTestService()
	p := &PostpartumRecord{PregnancyID: uuid.New(), PatientID: uuid.New()}
	svc.CreatePostpartum(context.Background(), p)

	fetched, err := svc.GetPostpartum(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.PregnancyID != p.PregnancyID {
		t.Error("pregnancy_id mismatch")
	}
}

func TestDeletePostpartum(t *testing.T) {
	svc := newTestService()
	p := &PostpartumRecord{PregnancyID: uuid.New(), PatientID: uuid.New()}
	svc.CreatePostpartum(context.Background(), p)
	err := svc.DeletePostpartum(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetPostpartum(context.Background(), p.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}
