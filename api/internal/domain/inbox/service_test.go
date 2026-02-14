package inbox

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// -- Mock Repositories --

type mockMessagePoolRepo struct {
	items map[uuid.UUID]*MessagePool
}

func newMockMessagePoolRepo() *mockMessagePoolRepo {
	return &mockMessagePoolRepo{items: make(map[uuid.UUID]*MessagePool)}
}

func (m *mockMessagePoolRepo) Create(_ context.Context, p *MessagePool) error {
	p.ID = uuid.New()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	m.items[p.ID] = p
	return nil
}

func (m *mockMessagePoolRepo) GetByID(_ context.Context, id uuid.UUID) (*MessagePool, error) {
	p, ok := m.items[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return p, nil
}

func (m *mockMessagePoolRepo) Update(_ context.Context, p *MessagePool) error {
	m.items[p.ID] = p
	return nil
}

func (m *mockMessagePoolRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockMessagePoolRepo) List(_ context.Context, limit, offset int) ([]*MessagePool, int, error) {
	var result []*MessagePool
	for _, p := range m.items {
		result = append(result, p)
	}
	return result, len(result), nil
}

type mockInboxMessageRepo struct {
	items   map[uuid.UUID]*InboxMessage
	members map[uuid.UUID]*MessagePoolMember
}

func newMockInboxMessageRepo() *mockInboxMessageRepo {
	return &mockInboxMessageRepo{
		items:   make(map[uuid.UUID]*InboxMessage),
		members: make(map[uuid.UUID]*MessagePoolMember),
	}
}

func (m *mockInboxMessageRepo) Create(_ context.Context, msg *InboxMessage) error {
	msg.ID = uuid.New()
	msg.CreatedAt = time.Now()
	msg.UpdatedAt = time.Now()
	m.items[msg.ID] = msg
	return nil
}

func (m *mockInboxMessageRepo) GetByID(_ context.Context, id uuid.UUID) (*InboxMessage, error) {
	msg, ok := m.items[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return msg, nil
}

func (m *mockInboxMessageRepo) Update(_ context.Context, msg *InboxMessage) error {
	m.items[msg.ID] = msg
	return nil
}

func (m *mockInboxMessageRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockInboxMessageRepo) ListByRecipient(_ context.Context, recipientID uuid.UUID, limit, offset int) ([]*InboxMessage, int, error) {
	var result []*InboxMessage
	for _, msg := range m.items {
		if msg.RecipientID != nil && *msg.RecipientID == recipientID {
			result = append(result, msg)
		}
	}
	return result, len(result), nil
}

func (m *mockInboxMessageRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*InboxMessage, int, error) {
	var result []*InboxMessage
	for _, msg := range m.items {
		if msg.PatientID != nil && *msg.PatientID == patientID {
			result = append(result, msg)
		}
	}
	return result, len(result), nil
}

func (m *mockInboxMessageRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*InboxMessage, int, error) {
	var result []*InboxMessage
	for _, msg := range m.items {
		result = append(result, msg)
	}
	return result, len(result), nil
}

func (m *mockInboxMessageRepo) AddPoolMember(_ context.Context, member *MessagePoolMember) error {
	member.ID = uuid.New()
	member.JoinedAt = time.Now()
	m.members[member.ID] = member
	return nil
}

func (m *mockInboxMessageRepo) GetPoolMembers(_ context.Context, poolID uuid.UUID) ([]*MessagePoolMember, error) {
	var result []*MessagePoolMember
	for _, member := range m.members {
		if member.PoolID == poolID {
			result = append(result, member)
		}
	}
	return result, nil
}

func (m *mockInboxMessageRepo) RemovePoolMember(_ context.Context, id uuid.UUID) error {
	delete(m.members, id)
	return nil
}

type mockCosignRequestRepo struct {
	items map[uuid.UUID]*CosignRequest
}

func newMockCosignRequestRepo() *mockCosignRequestRepo {
	return &mockCosignRequestRepo{items: make(map[uuid.UUID]*CosignRequest)}
}

func (m *mockCosignRequestRepo) Create(_ context.Context, r *CosignRequest) error {
	r.ID = uuid.New()
	r.RequestedAt = time.Now()
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	m.items[r.ID] = r
	return nil
}

func (m *mockCosignRequestRepo) GetByID(_ context.Context, id uuid.UUID) (*CosignRequest, error) {
	r, ok := m.items[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return r, nil
}

func (m *mockCosignRequestRepo) Update(_ context.Context, r *CosignRequest) error {
	m.items[r.ID] = r
	return nil
}

func (m *mockCosignRequestRepo) ListByCosigner(_ context.Context, cosignerID uuid.UUID, limit, offset int) ([]*CosignRequest, int, error) {
	var result []*CosignRequest
	for _, r := range m.items {
		if r.CosignerID == cosignerID {
			result = append(result, r)
		}
	}
	return result, len(result), nil
}

func (m *mockCosignRequestRepo) ListByRequester(_ context.Context, requesterID uuid.UUID, limit, offset int) ([]*CosignRequest, int, error) {
	var result []*CosignRequest
	for _, r := range m.items {
		if r.RequesterID == requesterID {
			result = append(result, r)
		}
	}
	return result, len(result), nil
}

type mockPatientListRepo struct {
	items   map[uuid.UUID]*PatientList
	members map[uuid.UUID]*PatientListMember
}

func newMockPatientListRepo() *mockPatientListRepo {
	return &mockPatientListRepo{
		items:   make(map[uuid.UUID]*PatientList),
		members: make(map[uuid.UUID]*PatientListMember),
	}
}

func (m *mockPatientListRepo) Create(_ context.Context, l *PatientList) error {
	l.ID = uuid.New()
	l.CreatedAt = time.Now()
	l.UpdatedAt = time.Now()
	m.items[l.ID] = l
	return nil
}

func (m *mockPatientListRepo) GetByID(_ context.Context, id uuid.UUID) (*PatientList, error) {
	l, ok := m.items[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return l, nil
}

func (m *mockPatientListRepo) Update(_ context.Context, l *PatientList) error {
	m.items[l.ID] = l
	return nil
}

func (m *mockPatientListRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockPatientListRepo) ListByOwner(_ context.Context, ownerID uuid.UUID, limit, offset int) ([]*PatientList, int, error) {
	var result []*PatientList
	for _, l := range m.items {
		if l.OwnerID == ownerID {
			result = append(result, l)
		}
	}
	return result, len(result), nil
}

func (m *mockPatientListRepo) AddMember(_ context.Context, member *PatientListMember) error {
	member.ID = uuid.New()
	member.AddedAt = time.Now()
	m.members[member.ID] = member
	return nil
}

func (m *mockPatientListRepo) GetMembers(_ context.Context, listID uuid.UUID, limit, offset int) ([]*PatientListMember, int, error) {
	var result []*PatientListMember
	for _, member := range m.members {
		if member.ListID == listID {
			result = append(result, member)
		}
	}
	return result, len(result), nil
}

func (m *mockPatientListRepo) RemoveMember(_ context.Context, id uuid.UUID) error {
	delete(m.members, id)
	return nil
}

func (m *mockPatientListRepo) UpdateMember(_ context.Context, member *PatientListMember) error {
	m.members[member.ID] = member
	return nil
}

type mockHandoffRepo struct {
	items map[uuid.UUID]*HandoffRecord
}

func newMockHandoffRepo() *mockHandoffRepo {
	return &mockHandoffRepo{items: make(map[uuid.UUID]*HandoffRecord)}
}

func (m *mockHandoffRepo) Create(_ context.Context, h *HandoffRecord) error {
	h.ID = uuid.New()
	h.CreatedAt = time.Now()
	h.UpdatedAt = time.Now()
	m.items[h.ID] = h
	return nil
}

func (m *mockHandoffRepo) GetByID(_ context.Context, id uuid.UUID) (*HandoffRecord, error) {
	h, ok := m.items[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return h, nil
}

func (m *mockHandoffRepo) Update(_ context.Context, h *HandoffRecord) error {
	m.items[h.ID] = h
	return nil
}

func (m *mockHandoffRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*HandoffRecord, int, error) {
	var result []*HandoffRecord
	for _, h := range m.items {
		if h.PatientID == patientID {
			result = append(result, h)
		}
	}
	return result, len(result), nil
}

func (m *mockHandoffRepo) ListByProvider(_ context.Context, providerID uuid.UUID, limit, offset int) ([]*HandoffRecord, int, error) {
	var result []*HandoffRecord
	for _, h := range m.items {
		if h.FromProviderID == providerID || h.ToProviderID == providerID {
			result = append(result, h)
		}
	}
	return result, len(result), nil
}

// -- Tests --

func newTestService() *Service {
	return NewService(
		newMockMessagePoolRepo(),
		newMockInboxMessageRepo(),
		newMockCosignRequestRepo(),
		newMockPatientListRepo(),
		newMockHandoffRepo(),
	)
}

// -- MessagePool Tests --

func TestCreateMessagePool(t *testing.T) {
	svc := newTestService()
	p := &MessagePool{PoolName: "Cardiology Pool", PoolType: "department"}
	err := svc.CreateMessagePool(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.IsActive {
		t.Error("expected is_active to default to true")
	}
}

func TestCreateMessagePool_PoolNameRequired(t *testing.T) {
	svc := newTestService()
	p := &MessagePool{PoolType: "department"}
	err := svc.CreateMessagePool(context.Background(), p)
	if err == nil {
		t.Error("expected error for missing pool_name")
	}
}

func TestCreateMessagePool_PoolTypeRequired(t *testing.T) {
	svc := newTestService()
	p := &MessagePool{PoolName: "Cardiology Pool"}
	err := svc.CreateMessagePool(context.Background(), p)
	if err == nil {
		t.Error("expected error for missing pool_type")
	}
}

func TestGetMessagePool(t *testing.T) {
	svc := newTestService()
	p := &MessagePool{PoolName: "Test Pool", PoolType: "shared"}
	svc.CreateMessagePool(context.Background(), p)

	fetched, err := svc.GetMessagePool(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != p.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestDeleteMessagePool(t *testing.T) {
	svc := newTestService()
	p := &MessagePool{PoolName: "Test Pool", PoolType: "shared"}
	svc.CreateMessagePool(context.Background(), p)
	err := svc.DeleteMessagePool(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetMessagePool(context.Background(), p.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

// -- InboxMessage Tests --

func TestCreateInboxMessage(t *testing.T) {
	svc := newTestService()
	m := &InboxMessage{MessageType: "result", Subject: "Lab Results Ready"}
	err := svc.CreateInboxMessage(context.Background(), m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Status != "unread" {
		t.Errorf("expected default status 'unread', got %s", m.Status)
	}
	if m.Priority != "normal" {
		t.Errorf("expected default priority 'normal', got %s", m.Priority)
	}
}

func TestCreateInboxMessage_MessageTypeRequired(t *testing.T) {
	svc := newTestService()
	m := &InboxMessage{Subject: "Test"}
	err := svc.CreateInboxMessage(context.Background(), m)
	if err == nil {
		t.Error("expected error for missing message_type")
	}
}

func TestCreateInboxMessage_SubjectRequired(t *testing.T) {
	svc := newTestService()
	m := &InboxMessage{MessageType: "result"}
	err := svc.CreateInboxMessage(context.Background(), m)
	if err == nil {
		t.Error("expected error for missing subject")
	}
}

func TestGetInboxMessage(t *testing.T) {
	svc := newTestService()
	m := &InboxMessage{MessageType: "result", Subject: "Lab"}
	svc.CreateInboxMessage(context.Background(), m)

	fetched, err := svc.GetInboxMessage(context.Background(), m.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != m.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestDeleteInboxMessage(t *testing.T) {
	svc := newTestService()
	m := &InboxMessage{MessageType: "result", Subject: "Lab"}
	svc.CreateInboxMessage(context.Background(), m)
	err := svc.DeleteInboxMessage(context.Background(), m.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetInboxMessage(context.Background(), m.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

// -- Pool Member Tests --

func TestAddPoolMember(t *testing.T) {
	svc := newTestService()
	p := &MessagePool{PoolName: "Test Pool", PoolType: "shared"}
	svc.CreateMessagePool(context.Background(), p)

	member := &MessagePoolMember{PoolID: p.ID, UserID: uuid.New()}
	err := svc.AddPoolMember(context.Background(), member)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !member.IsActive {
		t.Error("expected is_active to default to true")
	}

	members, err := svc.GetPoolMembers(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(members) != 1 {
		t.Errorf("expected 1 member, got %d", len(members))
	}
}

func TestAddPoolMember_PoolIDRequired(t *testing.T) {
	svc := newTestService()
	member := &MessagePoolMember{UserID: uuid.New()}
	err := svc.AddPoolMember(context.Background(), member)
	if err == nil {
		t.Error("expected error for missing pool_id")
	}
}

func TestAddPoolMember_UserIDRequired(t *testing.T) {
	svc := newTestService()
	member := &MessagePoolMember{PoolID: uuid.New()}
	err := svc.AddPoolMember(context.Background(), member)
	if err == nil {
		t.Error("expected error for missing user_id")
	}
}

// -- CosignRequest Tests --

func TestCreateCosignRequest(t *testing.T) {
	svc := newTestService()
	r := &CosignRequest{
		DocumentType: "progress_note",
		RequesterID:  uuid.New(),
		CosignerID:   uuid.New(),
	}
	err := svc.CreateCosignRequest(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Status != "pending" {
		t.Errorf("expected default status 'pending', got %s", r.Status)
	}
}

func TestCreateCosignRequest_DocumentTypeRequired(t *testing.T) {
	svc := newTestService()
	r := &CosignRequest{RequesterID: uuid.New(), CosignerID: uuid.New()}
	err := svc.CreateCosignRequest(context.Background(), r)
	if err == nil {
		t.Error("expected error for missing document_type")
	}
}

func TestCreateCosignRequest_RequesterIDRequired(t *testing.T) {
	svc := newTestService()
	r := &CosignRequest{DocumentType: "progress_note", CosignerID: uuid.New()}
	err := svc.CreateCosignRequest(context.Background(), r)
	if err == nil {
		t.Error("expected error for missing requester_id")
	}
}

func TestCreateCosignRequest_CosignerIDRequired(t *testing.T) {
	svc := newTestService()
	r := &CosignRequest{DocumentType: "progress_note", RequesterID: uuid.New()}
	err := svc.CreateCosignRequest(context.Background(), r)
	if err == nil {
		t.Error("expected error for missing cosigner_id")
	}
}

func TestGetCosignRequest(t *testing.T) {
	svc := newTestService()
	r := &CosignRequest{DocumentType: "progress_note", RequesterID: uuid.New(), CosignerID: uuid.New()}
	svc.CreateCosignRequest(context.Background(), r)

	fetched, err := svc.GetCosignRequest(context.Background(), r.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != r.ID {
		t.Error("unexpected ID mismatch")
	}
}

// -- PatientList Tests --

func TestCreatePatientList(t *testing.T) {
	svc := newTestService()
	l := &PatientList{ListName: "My Patients", ListType: "personal", OwnerID: uuid.New()}
	err := svc.CreatePatientList(context.Background(), l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !l.IsActive {
		t.Error("expected is_active to default to true")
	}
}

func TestCreatePatientList_ListNameRequired(t *testing.T) {
	svc := newTestService()
	l := &PatientList{ListType: "personal", OwnerID: uuid.New()}
	err := svc.CreatePatientList(context.Background(), l)
	if err == nil {
		t.Error("expected error for missing list_name")
	}
}

func TestCreatePatientList_ListTypeRequired(t *testing.T) {
	svc := newTestService()
	l := &PatientList{ListName: "My Patients", OwnerID: uuid.New()}
	err := svc.CreatePatientList(context.Background(), l)
	if err == nil {
		t.Error("expected error for missing list_type")
	}
}

func TestCreatePatientList_OwnerIDRequired(t *testing.T) {
	svc := newTestService()
	l := &PatientList{ListName: "My Patients", ListType: "personal"}
	err := svc.CreatePatientList(context.Background(), l)
	if err == nil {
		t.Error("expected error for missing owner_id")
	}
}

func TestGetPatientList(t *testing.T) {
	svc := newTestService()
	l := &PatientList{ListName: "Test List", ListType: "personal", OwnerID: uuid.New()}
	svc.CreatePatientList(context.Background(), l)

	fetched, err := svc.GetPatientList(context.Background(), l.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != l.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestDeletePatientList(t *testing.T) {
	svc := newTestService()
	l := &PatientList{ListName: "Test List", ListType: "personal", OwnerID: uuid.New()}
	svc.CreatePatientList(context.Background(), l)
	err := svc.DeletePatientList(context.Background(), l.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetPatientList(context.Background(), l.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

// -- PatientList Member Tests --

func TestAddPatientListMember(t *testing.T) {
	svc := newTestService()
	l := &PatientList{ListName: "Test", ListType: "personal", OwnerID: uuid.New()}
	svc.CreatePatientList(context.Background(), l)

	member := &PatientListMember{ListID: l.ID, PatientID: uuid.New()}
	err := svc.AddPatientListMember(context.Background(), member)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	members, _, err := svc.GetPatientListMembers(context.Background(), l.ID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(members) != 1 {
		t.Errorf("expected 1 member, got %d", len(members))
	}
}

func TestAddPatientListMember_ListIDRequired(t *testing.T) {
	svc := newTestService()
	member := &PatientListMember{PatientID: uuid.New()}
	err := svc.AddPatientListMember(context.Background(), member)
	if err == nil {
		t.Error("expected error for missing list_id")
	}
}

func TestAddPatientListMember_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	member := &PatientListMember{ListID: uuid.New()}
	err := svc.AddPatientListMember(context.Background(), member)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

// -- Handoff Tests --

func TestCreateHandoff(t *testing.T) {
	svc := newTestService()
	h := &HandoffRecord{
		PatientID:      uuid.New(),
		FromProviderID: uuid.New(),
		ToProviderID:   uuid.New(),
	}
	err := svc.CreateHandoff(context.Background(), h)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.HandoffType != "ipass" {
		t.Errorf("expected default handoff_type 'ipass', got %s", h.HandoffType)
	}
	if h.Status != "draft" {
		t.Errorf("expected default status 'draft', got %s", h.Status)
	}
}

func TestCreateHandoff_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	h := &HandoffRecord{FromProviderID: uuid.New(), ToProviderID: uuid.New()}
	err := svc.CreateHandoff(context.Background(), h)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateHandoff_FromProviderIDRequired(t *testing.T) {
	svc := newTestService()
	h := &HandoffRecord{PatientID: uuid.New(), ToProviderID: uuid.New()}
	err := svc.CreateHandoff(context.Background(), h)
	if err == nil {
		t.Error("expected error for missing from_provider_id")
	}
}

func TestCreateHandoff_ToProviderIDRequired(t *testing.T) {
	svc := newTestService()
	h := &HandoffRecord{PatientID: uuid.New(), FromProviderID: uuid.New()}
	err := svc.CreateHandoff(context.Background(), h)
	if err == nil {
		t.Error("expected error for missing to_provider_id")
	}
}

func TestGetHandoff(t *testing.T) {
	svc := newTestService()
	h := &HandoffRecord{PatientID: uuid.New(), FromProviderID: uuid.New(), ToProviderID: uuid.New()}
	svc.CreateHandoff(context.Background(), h)

	fetched, err := svc.GetHandoff(context.Background(), h.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != h.ID {
		t.Error("unexpected ID mismatch")
	}
}

// -- Additional MessagePool Tests --

func TestUpdateMessagePool(t *testing.T) {
	svc := newTestService()
	p := &MessagePool{PoolName: "Test Pool", PoolType: "shared"}
	svc.CreateMessagePool(context.Background(), p)
	p.PoolName = "Updated Pool"
	err := svc.UpdateMessagePool(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fetched, _ := svc.GetMessagePool(context.Background(), p.ID)
	if fetched.PoolName != "Updated Pool" {
		t.Errorf("expected 'Updated Pool', got %s", fetched.PoolName)
	}
}

func TestListMessagePools(t *testing.T) {
	svc := newTestService()
	svc.CreateMessagePool(context.Background(), &MessagePool{PoolName: "Pool A", PoolType: "department"})
	svc.CreateMessagePool(context.Background(), &MessagePool{PoolName: "Pool B", PoolType: "shared"})
	pools, total, err := svc.ListMessagePools(context.Background(), 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(pools) != 2 {
		t.Errorf("expected 2 pools, got %d", len(pools))
	}
}

// -- Additional InboxMessage Tests --

func TestUpdateInboxMessage(t *testing.T) {
	svc := newTestService()
	m := &InboxMessage{MessageType: "result", Subject: "Lab"}
	svc.CreateInboxMessage(context.Background(), m)
	m.Status = "read"
	err := svc.UpdateInboxMessage(context.Background(), m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateInboxMessage_InvalidStatus(t *testing.T) {
	svc := newTestService()
	m := &InboxMessage{MessageType: "result", Subject: "Lab"}
	svc.CreateInboxMessage(context.Background(), m)
	m.Status = "bogus"
	err := svc.UpdateInboxMessage(context.Background(), m)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestListInboxMessagesByRecipient(t *testing.T) {
	svc := newTestService()
	recipientID := uuid.New()
	m := &InboxMessage{MessageType: "result", Subject: "Lab", RecipientID: &recipientID}
	svc.CreateInboxMessage(context.Background(), m)
	msgs, total, err := svc.ListInboxMessagesByRecipient(context.Background(), recipientID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(msgs) != 1 {
		t.Errorf("expected 1 message, got %d", len(msgs))
	}
}

func TestListInboxMessagesByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	m := &InboxMessage{MessageType: "result", Subject: "Lab", PatientID: &patientID}
	svc.CreateInboxMessage(context.Background(), m)
	msgs, total, err := svc.ListInboxMessagesByPatient(context.Background(), patientID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(msgs) != 1 {
		t.Errorf("expected 1 message, got %d", len(msgs))
	}
}

func TestSearchInboxMessages(t *testing.T) {
	svc := newTestService()
	svc.CreateInboxMessage(context.Background(), &InboxMessage{MessageType: "result", Subject: "Lab"})
	msgs, total, err := svc.SearchInboxMessages(context.Background(), map[string]string{"type": "result"}, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(msgs) < 1 {
		t.Error("expected messages")
	}
}

// -- Additional Pool Member Tests --

func TestRemovePoolMember(t *testing.T) {
	svc := newTestService()
	p := &MessagePool{PoolName: "Test Pool", PoolType: "shared"}
	svc.CreateMessagePool(context.Background(), p)
	member := &MessagePoolMember{PoolID: p.ID, UserID: uuid.New()}
	svc.AddPoolMember(context.Background(), member)
	err := svc.RemovePoolMember(context.Background(), member.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	members, _ := svc.GetPoolMembers(context.Background(), p.ID)
	if len(members) != 0 {
		t.Errorf("expected 0 members after removal, got %d", len(members))
	}
}

// -- Additional CosignRequest Tests --

func TestUpdateCosignRequest(t *testing.T) {
	svc := newTestService()
	r := &CosignRequest{DocumentType: "progress_note", RequesterID: uuid.New(), CosignerID: uuid.New()}
	svc.CreateCosignRequest(context.Background(), r)
	r.Status = "cosigned"
	err := svc.UpdateCosignRequest(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateCosignRequest_InvalidStatus(t *testing.T) {
	svc := newTestService()
	r := &CosignRequest{DocumentType: "progress_note", RequesterID: uuid.New(), CosignerID: uuid.New()}
	svc.CreateCosignRequest(context.Background(), r)
	r.Status = "bogus"
	err := svc.UpdateCosignRequest(context.Background(), r)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestListCosignRequestsByCosigner(t *testing.T) {
	svc := newTestService()
	cosignerID := uuid.New()
	svc.CreateCosignRequest(context.Background(), &CosignRequest{DocumentType: "note", RequesterID: uuid.New(), CosignerID: cosignerID})
	items, total, err := svc.ListCosignRequestsByCosigner(context.Background(), cosignerID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestListCosignRequestsByRequester(t *testing.T) {
	svc := newTestService()
	requesterID := uuid.New()
	svc.CreateCosignRequest(context.Background(), &CosignRequest{DocumentType: "note", RequesterID: requesterID, CosignerID: uuid.New()})
	items, total, err := svc.ListCosignRequestsByRequester(context.Background(), requesterID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

// -- Additional PatientList Tests --

func TestUpdatePatientList(t *testing.T) {
	svc := newTestService()
	l := &PatientList{ListName: "Test List", ListType: "personal", OwnerID: uuid.New()}
	svc.CreatePatientList(context.Background(), l)
	l.ListName = "Updated List"
	err := svc.UpdatePatientList(context.Background(), l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListPatientListsByOwner(t *testing.T) {
	svc := newTestService()
	ownerID := uuid.New()
	svc.CreatePatientList(context.Background(), &PatientList{ListName: "List A", ListType: "personal", OwnerID: ownerID})
	svc.CreatePatientList(context.Background(), &PatientList{ListName: "List B", ListType: "personal", OwnerID: ownerID})
	lists, total, err := svc.ListPatientListsByOwner(context.Background(), ownerID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(lists) != 2 {
		t.Errorf("expected 2 lists, got %d", len(lists))
	}
}

func TestRemovePatientListMember(t *testing.T) {
	svc := newTestService()
	l := &PatientList{ListName: "Test", ListType: "personal", OwnerID: uuid.New()}
	svc.CreatePatientList(context.Background(), l)
	member := &PatientListMember{ListID: l.ID, PatientID: uuid.New()}
	svc.AddPatientListMember(context.Background(), member)
	err := svc.RemovePatientListMember(context.Background(), member.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	members, _, _ := svc.GetPatientListMembers(context.Background(), l.ID, 20, 0)
	if len(members) != 0 {
		t.Errorf("expected 0 members after removal, got %d", len(members))
	}
}

func TestUpdatePatientListMember(t *testing.T) {
	svc := newTestService()
	l := &PatientList{ListName: "Test", ListType: "personal", OwnerID: uuid.New()}
	svc.CreatePatientList(context.Background(), l)
	member := &PatientListMember{ListID: l.ID, PatientID: uuid.New()}
	svc.AddPatientListMember(context.Background(), member)
	err := svc.UpdatePatientListMember(context.Background(), member)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// -- Additional Handoff Tests --

func TestUpdateHandoff(t *testing.T) {
	svc := newTestService()
	h := &HandoffRecord{PatientID: uuid.New(), FromProviderID: uuid.New(), ToProviderID: uuid.New()}
	svc.CreateHandoff(context.Background(), h)
	h.Status = "completed"
	err := svc.UpdateHandoff(context.Background(), h)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListHandoffsByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreateHandoff(context.Background(), &HandoffRecord{PatientID: patientID, FromProviderID: uuid.New(), ToProviderID: uuid.New()})
	items, total, err := svc.ListHandoffsByPatient(context.Background(), patientID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestListHandoffsByProvider(t *testing.T) {
	svc := newTestService()
	providerID := uuid.New()
	svc.CreateHandoff(context.Background(), &HandoffRecord{PatientID: uuid.New(), FromProviderID: providerID, ToProviderID: uuid.New()})
	items, total, err := svc.ListHandoffsByProvider(context.Background(), providerID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}
