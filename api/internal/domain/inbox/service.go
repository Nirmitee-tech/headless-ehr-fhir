package inbox

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	pools    MessagePoolRepository
	messages InboxMessageRepository
	cosigns  CosignRequestRepository
	lists    PatientListRepository
	handoffs HandoffRepository
}

func NewService(
	pools MessagePoolRepository,
	messages InboxMessageRepository,
	cosigns CosignRequestRepository,
	lists PatientListRepository,
	handoffs HandoffRepository,
) *Service {
	return &Service{
		pools:    pools,
		messages: messages,
		cosigns:  cosigns,
		lists:    lists,
		handoffs: handoffs,
	}
}

// -- Message Pool --

func (s *Service) CreateMessagePool(ctx context.Context, p *MessagePool) error {
	if p.PoolName == "" {
		return fmt.Errorf("pool_name is required")
	}
	if p.PoolType == "" {
		return fmt.Errorf("pool_type is required")
	}
	p.IsActive = true
	return s.pools.Create(ctx, p)
}

func (s *Service) GetMessagePool(ctx context.Context, id uuid.UUID) (*MessagePool, error) {
	return s.pools.GetByID(ctx, id)
}

func (s *Service) UpdateMessagePool(ctx context.Context, p *MessagePool) error {
	return s.pools.Update(ctx, p)
}

func (s *Service) DeleteMessagePool(ctx context.Context, id uuid.UUID) error {
	return s.pools.Delete(ctx, id)
}

func (s *Service) ListMessagePools(ctx context.Context, limit, offset int) ([]*MessagePool, int, error) {
	return s.pools.List(ctx, limit, offset)
}

// -- Inbox Message --

var validMessageStatuses = map[string]bool{
	"unread": true, "read": true, "in-progress": true,
	"done": true, "forwarded": true, "deleted": true,
}

func (s *Service) CreateInboxMessage(ctx context.Context, m *InboxMessage) error {
	if m.MessageType == "" {
		return fmt.Errorf("message_type is required")
	}
	if m.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	if m.Status == "" {
		m.Status = "unread"
	}
	if !validMessageStatuses[m.Status] {
		return fmt.Errorf("invalid status: %s", m.Status)
	}
	if m.Priority == "" {
		m.Priority = "normal"
	}
	return s.messages.Create(ctx, m)
}

func (s *Service) GetInboxMessage(ctx context.Context, id uuid.UUID) (*InboxMessage, error) {
	return s.messages.GetByID(ctx, id)
}

func (s *Service) UpdateInboxMessage(ctx context.Context, m *InboxMessage) error {
	if m.Status != "" && !validMessageStatuses[m.Status] {
		return fmt.Errorf("invalid status: %s", m.Status)
	}
	return s.messages.Update(ctx, m)
}

func (s *Service) DeleteInboxMessage(ctx context.Context, id uuid.UUID) error {
	return s.messages.Delete(ctx, id)
}

func (s *Service) ListInboxMessagesByRecipient(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]*InboxMessage, int, error) {
	return s.messages.ListByRecipient(ctx, recipientID, limit, offset)
}

func (s *Service) ListInboxMessagesByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*InboxMessage, int, error) {
	return s.messages.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchInboxMessages(ctx context.Context, params map[string]string, limit, offset int) ([]*InboxMessage, int, error) {
	return s.messages.Search(ctx, params, limit, offset)
}

func (s *Service) AddPoolMember(ctx context.Context, m *MessagePoolMember) error {
	if m.PoolID == uuid.Nil {
		return fmt.Errorf("pool_id is required")
	}
	if m.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	m.IsActive = true
	return s.messages.AddPoolMember(ctx, m)
}

func (s *Service) GetPoolMembers(ctx context.Context, poolID uuid.UUID) ([]*MessagePoolMember, error) {
	return s.messages.GetPoolMembers(ctx, poolID)
}

func (s *Service) RemovePoolMember(ctx context.Context, id uuid.UUID) error {
	return s.messages.RemovePoolMember(ctx, id)
}

// -- Cosign Request --

var validCosignStatuses = map[string]bool{
	"pending": true, "cosigned": true, "rejected": true, "expired": true,
}

func (s *Service) CreateCosignRequest(ctx context.Context, r *CosignRequest) error {
	if r.DocumentType == "" {
		return fmt.Errorf("document_type is required")
	}
	if r.RequesterID == uuid.Nil {
		return fmt.Errorf("requester_id is required")
	}
	if r.CosignerID == uuid.Nil {
		return fmt.Errorf("cosigner_id is required")
	}
	if r.Status == "" {
		r.Status = "pending"
	}
	if !validCosignStatuses[r.Status] {
		return fmt.Errorf("invalid status: %s", r.Status)
	}
	return s.cosigns.Create(ctx, r)
}

func (s *Service) GetCosignRequest(ctx context.Context, id uuid.UUID) (*CosignRequest, error) {
	return s.cosigns.GetByID(ctx, id)
}

func (s *Service) UpdateCosignRequest(ctx context.Context, r *CosignRequest) error {
	if r.Status != "" && !validCosignStatuses[r.Status] {
		return fmt.Errorf("invalid status: %s", r.Status)
	}
	return s.cosigns.Update(ctx, r)
}

func (s *Service) ListCosignRequestsByCosigner(ctx context.Context, cosignerID uuid.UUID, limit, offset int) ([]*CosignRequest, int, error) {
	return s.cosigns.ListByCosigner(ctx, cosignerID, limit, offset)
}

func (s *Service) ListCosignRequestsByRequester(ctx context.Context, requesterID uuid.UUID, limit, offset int) ([]*CosignRequest, int, error) {
	return s.cosigns.ListByRequester(ctx, requesterID, limit, offset)
}

// -- Patient List --

func (s *Service) CreatePatientList(ctx context.Context, l *PatientList) error {
	if l.ListName == "" {
		return fmt.Errorf("list_name is required")
	}
	if l.ListType == "" {
		return fmt.Errorf("list_type is required")
	}
	if l.OwnerID == uuid.Nil {
		return fmt.Errorf("owner_id is required")
	}
	l.IsActive = true
	return s.lists.Create(ctx, l)
}

func (s *Service) GetPatientList(ctx context.Context, id uuid.UUID) (*PatientList, error) {
	return s.lists.GetByID(ctx, id)
}

func (s *Service) UpdatePatientList(ctx context.Context, l *PatientList) error {
	return s.lists.Update(ctx, l)
}

func (s *Service) DeletePatientList(ctx context.Context, id uuid.UUID) error {
	return s.lists.Delete(ctx, id)
}

func (s *Service) ListPatientListsByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*PatientList, int, error) {
	return s.lists.ListByOwner(ctx, ownerID, limit, offset)
}

func (s *Service) AddPatientListMember(ctx context.Context, m *PatientListMember) error {
	if m.ListID == uuid.Nil {
		return fmt.Errorf("list_id is required")
	}
	if m.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	return s.lists.AddMember(ctx, m)
}

func (s *Service) GetPatientListMembers(ctx context.Context, listID uuid.UUID, limit, offset int) ([]*PatientListMember, int, error) {
	return s.lists.GetMembers(ctx, listID, limit, offset)
}

func (s *Service) RemovePatientListMember(ctx context.Context, id uuid.UUID) error {
	return s.lists.RemoveMember(ctx, id)
}

func (s *Service) UpdatePatientListMember(ctx context.Context, m *PatientListMember) error {
	return s.lists.UpdateMember(ctx, m)
}

// -- Handoff --

func (s *Service) CreateHandoff(ctx context.Context, h *HandoffRecord) error {
	if h.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if h.FromProviderID == uuid.Nil {
		return fmt.Errorf("from_provider_id is required")
	}
	if h.ToProviderID == uuid.Nil {
		return fmt.Errorf("to_provider_id is required")
	}
	if h.HandoffType == "" {
		h.HandoffType = "ipass"
	}
	if h.Status == "" {
		h.Status = "draft"
	}
	return s.handoffs.Create(ctx, h)
}

func (s *Service) GetHandoff(ctx context.Context, id uuid.UUID) (*HandoffRecord, error) {
	return s.handoffs.GetByID(ctx, id)
}

func (s *Service) UpdateHandoff(ctx context.Context, h *HandoffRecord) error {
	return s.handoffs.Update(ctx, h)
}

func (s *Service) ListHandoffsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*HandoffRecord, int, error) {
	return s.handoffs.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) ListHandoffsByProvider(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]*HandoffRecord, int, error) {
	return s.handoffs.ListByProvider(ctx, providerID, limit, offset)
}
