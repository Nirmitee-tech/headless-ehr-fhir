package inbox

import (
	"context"

	"github.com/google/uuid"
)

type MessagePoolRepository interface {
	Create(ctx context.Context, p *MessagePool) error
	GetByID(ctx context.Context, id uuid.UUID) (*MessagePool, error)
	Update(ctx context.Context, p *MessagePool) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*MessagePool, int, error)
}

type InboxMessageRepository interface {
	Create(ctx context.Context, m *InboxMessage) error
	GetByID(ctx context.Context, id uuid.UUID) (*InboxMessage, error)
	Update(ctx context.Context, m *InboxMessage) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByRecipient(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]*InboxMessage, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*InboxMessage, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*InboxMessage, int, error)
	// Pool member management
	AddPoolMember(ctx context.Context, m *MessagePoolMember) error
	GetPoolMembers(ctx context.Context, poolID uuid.UUID) ([]*MessagePoolMember, error)
	RemovePoolMember(ctx context.Context, id uuid.UUID) error
}

type CosignRequestRepository interface {
	Create(ctx context.Context, r *CosignRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*CosignRequest, error)
	Update(ctx context.Context, r *CosignRequest) error
	ListByCosigner(ctx context.Context, cosignerID uuid.UUID, limit, offset int) ([]*CosignRequest, int, error)
	ListByRequester(ctx context.Context, requesterID uuid.UUID, limit, offset int) ([]*CosignRequest, int, error)
}

type PatientListRepository interface {
	Create(ctx context.Context, l *PatientList) error
	GetByID(ctx context.Context, id uuid.UUID) (*PatientList, error)
	Update(ctx context.Context, l *PatientList) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*PatientList, int, error)
	// Member management
	AddMember(ctx context.Context, m *PatientListMember) error
	GetMembers(ctx context.Context, listID uuid.UUID, limit, offset int) ([]*PatientListMember, int, error)
	RemoveMember(ctx context.Context, id uuid.UUID) error
	UpdateMember(ctx context.Context, m *PatientListMember) error
}

type HandoffRepository interface {
	Create(ctx context.Context, h *HandoffRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*HandoffRecord, error)
	Update(ctx context.Context, h *HandoffRecord) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*HandoffRecord, int, error)
	ListByProvider(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]*HandoffRecord, int, error)
}
