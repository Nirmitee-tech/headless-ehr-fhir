package portal

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// ── Mock Repositories ──

type mockAccountRepo struct {
	data map[uuid.UUID]*PortalAccount
}

func (m *mockAccountRepo) Create(_ context.Context, a *PortalAccount) error {
	a.ID = uuid.New()
	m.data[a.ID] = a
	return nil
}
func (m *mockAccountRepo) GetByID(_ context.Context, id uuid.UUID) (*PortalAccount, error) {
	if a, ok := m.data[id]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockAccountRepo) Update(_ context.Context, a *PortalAccount) error {
	if _, ok := m.data[a.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.data[a.ID] = a
	return nil
}
func (m *mockAccountRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.data, id)
	return nil
}
func (m *mockAccountRepo) List(_ context.Context, limit, offset int) ([]*PortalAccount, int, error) {
	var out []*PortalAccount
	for _, a := range m.data {
		out = append(out, a)
	}
	return out, len(out), nil
}
func (m *mockAccountRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*PortalAccount, int, error) {
	var out []*PortalAccount
	for _, a := range m.data {
		if a.PatientID == patientID {
			out = append(out, a)
		}
	}
	return out, len(out), nil
}

type mockMessageRepo struct {
	data map[uuid.UUID]*PortalMessage
}

func (m *mockMessageRepo) Create(_ context.Context, msg *PortalMessage) error {
	msg.ID = uuid.New()
	m.data[msg.ID] = msg
	return nil
}
func (m *mockMessageRepo) GetByID(_ context.Context, id uuid.UUID) (*PortalMessage, error) {
	if msg, ok := m.data[id]; ok {
		return msg, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockMessageRepo) Update(_ context.Context, msg *PortalMessage) error {
	if _, ok := m.data[msg.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.data[msg.ID] = msg
	return nil
}
func (m *mockMessageRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.data, id)
	return nil
}
func (m *mockMessageRepo) List(_ context.Context, limit, offset int) ([]*PortalMessage, int, error) {
	var out []*PortalMessage
	for _, msg := range m.data {
		out = append(out, msg)
	}
	return out, len(out), nil
}
func (m *mockMessageRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*PortalMessage, int, error) {
	var out []*PortalMessage
	for _, msg := range m.data {
		if msg.PatientID == patientID {
			out = append(out, msg)
		}
	}
	return out, len(out), nil
}

type mockQuestionnaireRepo struct {
	data  map[uuid.UUID]*Questionnaire
	items map[uuid.UUID][]*QuestionnaireItem
}

func (m *mockQuestionnaireRepo) Create(_ context.Context, q *Questionnaire) error {
	q.ID = uuid.New()
	if q.FHIRID == "" {
		q.FHIRID = "fhir-" + q.ID.String()
	}
	m.data[q.ID] = q
	return nil
}
func (m *mockQuestionnaireRepo) GetByID(_ context.Context, id uuid.UUID) (*Questionnaire, error) {
	if q, ok := m.data[id]; ok {
		return q, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockQuestionnaireRepo) GetByFHIRID(_ context.Context, fhirID string) (*Questionnaire, error) {
	for _, q := range m.data {
		if q.FHIRID == fhirID {
			return q, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockQuestionnaireRepo) Update(_ context.Context, q *Questionnaire) error {
	if _, ok := m.data[q.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.data[q.ID] = q
	return nil
}
func (m *mockQuestionnaireRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.data, id)
	return nil
}
func (m *mockQuestionnaireRepo) List(_ context.Context, limit, offset int) ([]*Questionnaire, int, error) {
	var out []*Questionnaire
	for _, q := range m.data {
		out = append(out, q)
	}
	return out, len(out), nil
}
func (m *mockQuestionnaireRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*Questionnaire, int, error) {
	var out []*Questionnaire
	for _, q := range m.data {
		out = append(out, q)
	}
	return out, len(out), nil
}
func (m *mockQuestionnaireRepo) AddItem(_ context.Context, item *QuestionnaireItem) error {
	item.ID = uuid.New()
	m.items[item.QuestionnaireID] = append(m.items[item.QuestionnaireID], item)
	return nil
}
func (m *mockQuestionnaireRepo) GetItems(_ context.Context, questionnaireID uuid.UUID) ([]*QuestionnaireItem, error) {
	return m.items[questionnaireID], nil
}

type mockQRRepo struct {
	data  map[uuid.UUID]*QuestionnaireResponse
	items map[uuid.UUID][]*QuestionnaireResponseItem
}

func (m *mockQRRepo) Create(_ context.Context, qr *QuestionnaireResponse) error {
	qr.ID = uuid.New()
	if qr.FHIRID == "" {
		qr.FHIRID = "fhir-" + qr.ID.String()
	}
	m.data[qr.ID] = qr
	return nil
}
func (m *mockQRRepo) GetByID(_ context.Context, id uuid.UUID) (*QuestionnaireResponse, error) {
	if qr, ok := m.data[id]; ok {
		return qr, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockQRRepo) GetByFHIRID(_ context.Context, fhirID string) (*QuestionnaireResponse, error) {
	for _, qr := range m.data {
		if qr.FHIRID == fhirID {
			return qr, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockQRRepo) Update(_ context.Context, qr *QuestionnaireResponse) error {
	if _, ok := m.data[qr.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.data[qr.ID] = qr
	return nil
}
func (m *mockQRRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.data, id)
	return nil
}
func (m *mockQRRepo) List(_ context.Context, limit, offset int) ([]*QuestionnaireResponse, int, error) {
	var out []*QuestionnaireResponse
	for _, qr := range m.data {
		out = append(out, qr)
	}
	return out, len(out), nil
}
func (m *mockQRRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*QuestionnaireResponse, int, error) {
	var out []*QuestionnaireResponse
	for _, qr := range m.data {
		if qr.PatientID == patientID {
			out = append(out, qr)
		}
	}
	return out, len(out), nil
}
func (m *mockQRRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*QuestionnaireResponse, int, error) {
	var out []*QuestionnaireResponse
	for _, qr := range m.data {
		out = append(out, qr)
	}
	return out, len(out), nil
}
func (m *mockQRRepo) AddResponseItem(_ context.Context, item *QuestionnaireResponseItem) error {
	item.ID = uuid.New()
	m.items[item.ResponseID] = append(m.items[item.ResponseID], item)
	return nil
}
func (m *mockQRRepo) GetResponseItems(_ context.Context, responseID uuid.UUID) ([]*QuestionnaireResponseItem, error) {
	return m.items[responseID], nil
}

type mockCheckinRepo struct {
	data map[uuid.UUID]*PatientCheckin
}

func (m *mockCheckinRepo) Create(_ context.Context, c *PatientCheckin) error {
	c.ID = uuid.New()
	m.data[c.ID] = c
	return nil
}
func (m *mockCheckinRepo) GetByID(_ context.Context, id uuid.UUID) (*PatientCheckin, error) {
	if c, ok := m.data[id]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockCheckinRepo) Update(_ context.Context, c *PatientCheckin) error {
	if _, ok := m.data[c.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.data[c.ID] = c
	return nil
}
func (m *mockCheckinRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.data, id)
	return nil
}
func (m *mockCheckinRepo) List(_ context.Context, limit, offset int) ([]*PatientCheckin, int, error) {
	var out []*PatientCheckin
	for _, c := range m.data {
		out = append(out, c)
	}
	return out, len(out), nil
}
func (m *mockCheckinRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*PatientCheckin, int, error) {
	var out []*PatientCheckin
	for _, c := range m.data {
		if c.PatientID == patientID {
			out = append(out, c)
		}
	}
	return out, len(out), nil
}

// ── Helper ──

func newTestService() *Service {
	return NewService(
		&mockAccountRepo{data: make(map[uuid.UUID]*PortalAccount)},
		&mockMessageRepo{data: make(map[uuid.UUID]*PortalMessage)},
		&mockQuestionnaireRepo{data: make(map[uuid.UUID]*Questionnaire), items: make(map[uuid.UUID][]*QuestionnaireItem)},
		&mockQRRepo{data: make(map[uuid.UUID]*QuestionnaireResponse), items: make(map[uuid.UUID][]*QuestionnaireResponseItem)},
		&mockCheckinRepo{data: make(map[uuid.UUID]*PatientCheckin)},
	)
}

// ── Portal Account Tests ──

func TestService_CreatePortalAccount(t *testing.T) {
	svc := newTestService()
	a := &PortalAccount{PatientID: uuid.New()}
	if err := svc.CreatePortalAccount(nil, a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if a.Status != "pending-activation" {
		t.Errorf("expected default status 'pending-activation', got %s", a.Status)
	}
}

func TestService_CreatePortalAccount_MissingPatientID(t *testing.T) {
	svc := newTestService()
	a := &PortalAccount{}
	if err := svc.CreatePortalAccount(nil, a); err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestService_CreatePortalAccount_InvalidStatus(t *testing.T) {
	svc := newTestService()
	a := &PortalAccount{PatientID: uuid.New(), Status: "bogus"}
	if err := svc.CreatePortalAccount(nil, a); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_CreatePortalAccount_ValidStatuses(t *testing.T) {
	statuses := []string{"active", "inactive", "locked", "pending-activation", "suspended"}
	for _, status := range statuses {
		svc := newTestService()
		a := &PortalAccount{PatientID: uuid.New(), Status: status}
		if err := svc.CreatePortalAccount(nil, a); err != nil {
			t.Errorf("status %q should be valid, got error: %v", status, err)
		}
	}
}

func TestService_GetPortalAccount(t *testing.T) {
	svc := newTestService()
	a := &PortalAccount{PatientID: uuid.New()}
	svc.CreatePortalAccount(nil, a)
	got, err := svc.GetPortalAccount(nil, a.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != a.ID {
		t.Error("ID mismatch")
	}
}

func TestService_GetPortalAccount_NotFound(t *testing.T) {
	svc := newTestService()
	if _, err := svc.GetPortalAccount(nil, uuid.New()); err == nil {
		t.Error("expected error for not found")
	}
}

func TestService_UpdatePortalAccount(t *testing.T) {
	svc := newTestService()
	a := &PortalAccount{PatientID: uuid.New()}
	svc.CreatePortalAccount(nil, a)
	a.Status = "active"
	if err := svc.UpdatePortalAccount(nil, a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_UpdatePortalAccount_InvalidStatus(t *testing.T) {
	svc := newTestService()
	a := &PortalAccount{PatientID: uuid.New()}
	svc.CreatePortalAccount(nil, a)
	a.Status = "bad"
	if err := svc.UpdatePortalAccount(nil, a); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_DeletePortalAccount(t *testing.T) {
	svc := newTestService()
	a := &PortalAccount{PatientID: uuid.New()}
	svc.CreatePortalAccount(nil, a)
	if err := svc.DeletePortalAccount(nil, a.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_ListPortalAccounts(t *testing.T) {
	svc := newTestService()
	svc.CreatePortalAccount(nil, &PortalAccount{PatientID: uuid.New()})
	svc.CreatePortalAccount(nil, &PortalAccount{PatientID: uuid.New()})
	items, total, err := svc.ListPortalAccounts(nil, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("expected 2, got %d", len(items))
	}
}

func TestService_ListPortalAccountsByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreatePortalAccount(nil, &PortalAccount{PatientID: pid})
	svc.CreatePortalAccount(nil, &PortalAccount{PatientID: uuid.New()})
	items, total, err := svc.ListPortalAccountsByPatient(nil, pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1, got %d", len(items))
	}
}

// ── Portal Message Tests ──

func TestService_CreatePortalMessage(t *testing.T) {
	svc := newTestService()
	m := &PortalMessage{PatientID: uuid.New(), Body: "Hello"}
	if err := svc.CreatePortalMessage(nil, m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Status != "sent" {
		t.Errorf("expected default status 'sent', got %s", m.Status)
	}
}

func TestService_CreatePortalMessage_MissingPatientID(t *testing.T) {
	svc := newTestService()
	m := &PortalMessage{Body: "Hello"}
	if err := svc.CreatePortalMessage(nil, m); err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestService_CreatePortalMessage_MissingBody(t *testing.T) {
	svc := newTestService()
	m := &PortalMessage{PatientID: uuid.New()}
	if err := svc.CreatePortalMessage(nil, m); err == nil {
		t.Error("expected error for missing body")
	}
}

func TestService_CreatePortalMessage_InvalidStatus(t *testing.T) {
	svc := newTestService()
	m := &PortalMessage{PatientID: uuid.New(), Body: "Hi", Status: "bogus"}
	if err := svc.CreatePortalMessage(nil, m); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_GetPortalMessage(t *testing.T) {
	svc := newTestService()
	m := &PortalMessage{PatientID: uuid.New(), Body: "Hello"}
	svc.CreatePortalMessage(nil, m)
	got, err := svc.GetPortalMessage(nil, m.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Body != "Hello" {
		t.Errorf("expected 'Hello', got %s", got.Body)
	}
}

func TestService_UpdatePortalMessage(t *testing.T) {
	svc := newTestService()
	m := &PortalMessage{PatientID: uuid.New(), Body: "Hello"}
	svc.CreatePortalMessage(nil, m)
	m.Status = "read"
	if err := svc.UpdatePortalMessage(nil, m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_UpdatePortalMessage_InvalidStatus(t *testing.T) {
	svc := newTestService()
	m := &PortalMessage{PatientID: uuid.New(), Body: "Hello"}
	svc.CreatePortalMessage(nil, m)
	m.Status = "bad"
	if err := svc.UpdatePortalMessage(nil, m); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_DeletePortalMessage(t *testing.T) {
	svc := newTestService()
	m := &PortalMessage{PatientID: uuid.New(), Body: "Hello"}
	svc.CreatePortalMessage(nil, m)
	if err := svc.DeletePortalMessage(nil, m.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_ListPortalMessages(t *testing.T) {
	svc := newTestService()
	svc.CreatePortalMessage(nil, &PortalMessage{PatientID: uuid.New(), Body: "A"})
	svc.CreatePortalMessage(nil, &PortalMessage{PatientID: uuid.New(), Body: "B"})
	items, total, err := svc.ListPortalMessages(nil, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("expected 2, got %d", len(items))
	}
}

func TestService_ListPortalMessagesByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreatePortalMessage(nil, &PortalMessage{PatientID: pid, Body: "A"})
	svc.CreatePortalMessage(nil, &PortalMessage{PatientID: uuid.New(), Body: "B"})
	items, total, err := svc.ListPortalMessagesByPatient(nil, pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1, got %d", len(items))
	}
}

// ── Questionnaire Tests ──

func TestService_CreateQuestionnaire(t *testing.T) {
	svc := newTestService()
	q := &Questionnaire{Name: "PHQ-9"}
	if err := svc.CreateQuestionnaire(nil, q); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Status != "draft" {
		t.Errorf("expected default status 'draft', got %s", q.Status)
	}
}

func TestService_CreateQuestionnaire_MissingName(t *testing.T) {
	svc := newTestService()
	q := &Questionnaire{}
	if err := svc.CreateQuestionnaire(nil, q); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestService_CreateQuestionnaire_InvalidStatus(t *testing.T) {
	svc := newTestService()
	q := &Questionnaire{Name: "PHQ-9", Status: "bogus"}
	if err := svc.CreateQuestionnaire(nil, q); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_CreateQuestionnaire_ValidStatuses(t *testing.T) {
	statuses := []string{"draft", "active", "retired", "unknown"}
	for _, status := range statuses {
		svc := newTestService()
		q := &Questionnaire{Name: "Q", Status: status}
		if err := svc.CreateQuestionnaire(nil, q); err != nil {
			t.Errorf("status %q should be valid, got error: %v", status, err)
		}
	}
}

func TestService_GetQuestionnaire(t *testing.T) {
	svc := newTestService()
	q := &Questionnaire{Name: "PHQ-9"}
	svc.CreateQuestionnaire(nil, q)
	got, err := svc.GetQuestionnaire(nil, q.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "PHQ-9" {
		t.Errorf("expected 'PHQ-9', got %s", got.Name)
	}
}

func TestService_GetQuestionnaireByFHIRID(t *testing.T) {
	svc := newTestService()
	q := &Questionnaire{Name: "PHQ-9"}
	svc.CreateQuestionnaire(nil, q)
	got, err := svc.GetQuestionnaireByFHIRID(nil, q.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != q.ID {
		t.Error("ID mismatch")
	}
}

func TestService_UpdateQuestionnaire(t *testing.T) {
	svc := newTestService()
	q := &Questionnaire{Name: "PHQ-9"}
	svc.CreateQuestionnaire(nil, q)
	q.Status = "active"
	if err := svc.UpdateQuestionnaire(nil, q); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_UpdateQuestionnaire_InvalidStatus(t *testing.T) {
	svc := newTestService()
	q := &Questionnaire{Name: "PHQ-9"}
	svc.CreateQuestionnaire(nil, q)
	q.Status = "bad"
	if err := svc.UpdateQuestionnaire(nil, q); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_DeleteQuestionnaire(t *testing.T) {
	svc := newTestService()
	q := &Questionnaire{Name: "PHQ-9"}
	svc.CreateQuestionnaire(nil, q)
	if err := svc.DeleteQuestionnaire(nil, q.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_ListQuestionnaires(t *testing.T) {
	svc := newTestService()
	svc.CreateQuestionnaire(nil, &Questionnaire{Name: "Q1"})
	svc.CreateQuestionnaire(nil, &Questionnaire{Name: "Q2"})
	items, total, err := svc.ListQuestionnaires(nil, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("expected 2, got %d", len(items))
	}
}

func TestService_SearchQuestionnaires(t *testing.T) {
	svc := newTestService()
	svc.CreateQuestionnaire(nil, &Questionnaire{Name: "Q1"})
	items, total, err := svc.SearchQuestionnaires(nil, map[string]string{"name": "Q1"}, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(items) < 1 {
		t.Error("expected items")
	}
}

func TestService_AddQuestionnaireItem(t *testing.T) {
	svc := newTestService()
	q := &Questionnaire{Name: "PHQ-9"}
	svc.CreateQuestionnaire(nil, q)
	item := &QuestionnaireItem{QuestionnaireID: q.ID, LinkID: "q1", Text: "How are you?"}
	if err := svc.AddQuestionnaireItem(nil, item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestService_AddQuestionnaireItem_MissingQuestionnaireID(t *testing.T) {
	svc := newTestService()
	item := &QuestionnaireItem{LinkID: "q1", Text: "How are you?"}
	if err := svc.AddQuestionnaireItem(nil, item); err == nil {
		t.Error("expected error for missing questionnaire_id")
	}
}

func TestService_AddQuestionnaireItem_MissingLinkID(t *testing.T) {
	svc := newTestService()
	item := &QuestionnaireItem{QuestionnaireID: uuid.New(), Text: "How are you?"}
	if err := svc.AddQuestionnaireItem(nil, item); err == nil {
		t.Error("expected error for missing link_id")
	}
}

func TestService_AddQuestionnaireItem_MissingText(t *testing.T) {
	svc := newTestService()
	item := &QuestionnaireItem{QuestionnaireID: uuid.New(), LinkID: "q1"}
	if err := svc.AddQuestionnaireItem(nil, item); err == nil {
		t.Error("expected error for missing text")
	}
}

func TestService_GetQuestionnaireItems(t *testing.T) {
	svc := newTestService()
	q := &Questionnaire{Name: "PHQ-9"}
	svc.CreateQuestionnaire(nil, q)
	svc.AddQuestionnaireItem(nil, &QuestionnaireItem{QuestionnaireID: q.ID, LinkID: "q1", Text: "Q1"})
	svc.AddQuestionnaireItem(nil, &QuestionnaireItem{QuestionnaireID: q.ID, LinkID: "q2", Text: "Q2"})
	items, err := svc.GetQuestionnaireItems(nil, q.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2, got %d", len(items))
	}
}

// ── Questionnaire Response Tests ──

func TestService_CreateQuestionnaireResponse(t *testing.T) {
	svc := newTestService()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	if err := svc.CreateQuestionnaireResponse(nil, qr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if qr.Status != "in-progress" {
		t.Errorf("expected default status 'in-progress', got %s", qr.Status)
	}
}

func TestService_CreateQuestionnaireResponse_MissingQuestionnaireID(t *testing.T) {
	svc := newTestService()
	qr := &QuestionnaireResponse{PatientID: uuid.New()}
	if err := svc.CreateQuestionnaireResponse(nil, qr); err == nil {
		t.Error("expected error for missing questionnaire_id")
	}
}

func TestService_CreateQuestionnaireResponse_MissingPatientID(t *testing.T) {
	svc := newTestService()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New()}
	if err := svc.CreateQuestionnaireResponse(nil, qr); err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestService_CreateQuestionnaireResponse_InvalidStatus(t *testing.T) {
	svc := newTestService()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New(), Status: "bogus"}
	if err := svc.CreateQuestionnaireResponse(nil, qr); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_CreateQuestionnaireResponse_ValidStatuses(t *testing.T) {
	statuses := []string{"in-progress", "completed", "amended", "entered-in-error", "stopped"}
	for _, status := range statuses {
		svc := newTestService()
		qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New(), Status: status}
		if err := svc.CreateQuestionnaireResponse(nil, qr); err != nil {
			t.Errorf("status %q should be valid, got error: %v", status, err)
		}
	}
}

func TestService_GetQuestionnaireResponse(t *testing.T) {
	svc := newTestService()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	svc.CreateQuestionnaireResponse(nil, qr)
	got, err := svc.GetQuestionnaireResponse(nil, qr.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != qr.ID {
		t.Error("ID mismatch")
	}
}

func TestService_GetQuestionnaireResponseByFHIRID(t *testing.T) {
	svc := newTestService()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	svc.CreateQuestionnaireResponse(nil, qr)
	got, err := svc.GetQuestionnaireResponseByFHIRID(nil, qr.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != qr.ID {
		t.Error("ID mismatch")
	}
}

func TestService_UpdateQuestionnaireResponse(t *testing.T) {
	svc := newTestService()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	svc.CreateQuestionnaireResponse(nil, qr)
	qr.Status = "completed"
	if err := svc.UpdateQuestionnaireResponse(nil, qr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_UpdateQuestionnaireResponse_InvalidStatus(t *testing.T) {
	svc := newTestService()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	svc.CreateQuestionnaireResponse(nil, qr)
	qr.Status = "bad"
	if err := svc.UpdateQuestionnaireResponse(nil, qr); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_DeleteQuestionnaireResponse(t *testing.T) {
	svc := newTestService()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	svc.CreateQuestionnaireResponse(nil, qr)
	if err := svc.DeleteQuestionnaireResponse(nil, qr.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_ListQuestionnaireResponses(t *testing.T) {
	svc := newTestService()
	svc.CreateQuestionnaireResponse(nil, &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()})
	items, total, err := svc.ListQuestionnaireResponses(nil, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1, got %d", len(items))
	}
}

func TestService_ListQuestionnaireResponsesByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateQuestionnaireResponse(nil, &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: pid})
	svc.CreateQuestionnaireResponse(nil, &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()})
	items, total, err := svc.ListQuestionnaireResponsesByPatient(nil, pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1, got %d", len(items))
	}
}

func TestService_SearchQuestionnaireResponses(t *testing.T) {
	svc := newTestService()
	svc.CreateQuestionnaireResponse(nil, &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()})
	items, total, err := svc.SearchQuestionnaireResponses(nil, map[string]string{"status": "in-progress"}, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(items) < 1 {
		t.Error("expected items")
	}
}

func TestService_AddQuestionnaireResponseItem(t *testing.T) {
	svc := newTestService()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	svc.CreateQuestionnaireResponse(nil, qr)
	item := &QuestionnaireResponseItem{ResponseID: qr.ID, LinkID: "q1"}
	if err := svc.AddQuestionnaireResponseItem(nil, item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestService_AddQuestionnaireResponseItem_MissingResponseID(t *testing.T) {
	svc := newTestService()
	item := &QuestionnaireResponseItem{LinkID: "q1"}
	if err := svc.AddQuestionnaireResponseItem(nil, item); err == nil {
		t.Error("expected error for missing response_id")
	}
}

func TestService_AddQuestionnaireResponseItem_MissingLinkID(t *testing.T) {
	svc := newTestService()
	item := &QuestionnaireResponseItem{ResponseID: uuid.New()}
	if err := svc.AddQuestionnaireResponseItem(nil, item); err == nil {
		t.Error("expected error for missing link_id")
	}
}

func TestService_GetQuestionnaireResponseItems(t *testing.T) {
	svc := newTestService()
	qr := &QuestionnaireResponse{QuestionnaireID: uuid.New(), PatientID: uuid.New()}
	svc.CreateQuestionnaireResponse(nil, qr)
	svc.AddQuestionnaireResponseItem(nil, &QuestionnaireResponseItem{ResponseID: qr.ID, LinkID: "q1"})
	svc.AddQuestionnaireResponseItem(nil, &QuestionnaireResponseItem{ResponseID: qr.ID, LinkID: "q2"})
	items, err := svc.GetQuestionnaireResponseItems(nil, qr.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2, got %d", len(items))
	}
}

// ── Patient Checkin Tests ──

func TestService_CreatePatientCheckin(t *testing.T) {
	svc := newTestService()
	c := &PatientCheckin{PatientID: uuid.New()}
	if err := svc.CreatePatientCheckin(nil, c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Status != "pending" {
		t.Errorf("expected default status 'pending', got %s", c.Status)
	}
}

func TestService_CreatePatientCheckin_MissingPatientID(t *testing.T) {
	svc := newTestService()
	c := &PatientCheckin{}
	if err := svc.CreatePatientCheckin(nil, c); err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestService_GetPatientCheckin(t *testing.T) {
	svc := newTestService()
	c := &PatientCheckin{PatientID: uuid.New()}
	svc.CreatePatientCheckin(nil, c)
	got, err := svc.GetPatientCheckin(nil, c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != c.ID {
		t.Error("ID mismatch")
	}
}

func TestService_UpdatePatientCheckin(t *testing.T) {
	svc := newTestService()
	c := &PatientCheckin{PatientID: uuid.New()}
	svc.CreatePatientCheckin(nil, c)
	c.Status = "completed"
	if err := svc.UpdatePatientCheckin(nil, c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_DeletePatientCheckin(t *testing.T) {
	svc := newTestService()
	c := &PatientCheckin{PatientID: uuid.New()}
	svc.CreatePatientCheckin(nil, c)
	if err := svc.DeletePatientCheckin(nil, c.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_ListPatientCheckins(t *testing.T) {
	svc := newTestService()
	svc.CreatePatientCheckin(nil, &PatientCheckin{PatientID: uuid.New()})
	svc.CreatePatientCheckin(nil, &PatientCheckin{PatientID: uuid.New()})
	items, total, err := svc.ListPatientCheckins(nil, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("expected 2, got %d", len(items))
	}
}

func TestService_ListPatientCheckinsByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreatePatientCheckin(nil, &PatientCheckin{PatientID: pid})
	svc.CreatePatientCheckin(nil, &PatientCheckin{PatientID: uuid.New()})
	items, total, err := svc.ListPatientCheckinsByPatient(nil, pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1, got %d", len(items))
	}
}

// ── Additional NotFound Tests ──

func TestService_GetPortalMessage_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetPortalMessage(nil, uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestService_GetQuestionnaire_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetQuestionnaire(nil, uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestService_GetQuestionnaireByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetQuestionnaireByFHIRID(nil, "nonexistent")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestService_GetQuestionnaireResponse_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetQuestionnaireResponse(nil, uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestService_GetQuestionnaireResponseByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetQuestionnaireResponseByFHIRID(nil, "nonexistent")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestService_GetPatientCheckin_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetPatientCheckin(nil, uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestService_CreatePortalMessage_ValidStatuses(t *testing.T) {
	statuses := []string{"sent", "delivered", "read", "replied", "closed", "archived"}
	for _, status := range statuses {
		svc := newTestService()
		m := &PortalMessage{PatientID: uuid.New(), Body: "Hi", Status: status}
		if err := svc.CreatePortalMessage(nil, m); err != nil {
			t.Errorf("status %q should be valid, got error: %v", status, err)
		}
	}
}

func TestService_UpdatePortalAccount_NotFound(t *testing.T) {
	svc := newTestService()
	a := &PortalAccount{ID: uuid.New(), Status: "active"}
	if err := svc.UpdatePortalAccount(nil, a); err == nil {
		t.Error("expected error for not found")
	}
}

func TestService_UpdatePortalMessage_NotFound(t *testing.T) {
	svc := newTestService()
	m := &PortalMessage{ID: uuid.New(), Status: "read"}
	if err := svc.UpdatePortalMessage(nil, m); err == nil {
		t.Error("expected error for not found")
	}
}

func TestService_UpdateQuestionnaire_NotFound(t *testing.T) {
	svc := newTestService()
	q := &Questionnaire{ID: uuid.New(), Status: "active"}
	if err := svc.UpdateQuestionnaire(nil, q); err == nil {
		t.Error("expected error for not found")
	}
}

func TestService_UpdateQuestionnaireResponse_NotFound(t *testing.T) {
	svc := newTestService()
	qr := &QuestionnaireResponse{ID: uuid.New(), Status: "completed"}
	if err := svc.UpdateQuestionnaireResponse(nil, qr); err == nil {
		t.Error("expected error for not found")
	}
}

func TestService_UpdatePatientCheckin_NotFound(t *testing.T) {
	svc := newTestService()
	c := &PatientCheckin{ID: uuid.New(), Status: "completed"}
	if err := svc.UpdatePatientCheckin(nil, c); err == nil {
		t.Error("expected error for not found")
	}
}
