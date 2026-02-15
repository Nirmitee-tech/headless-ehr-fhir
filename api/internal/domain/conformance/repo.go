package conformance

import (
	"context"

	"github.com/google/uuid"
)

type NamingSystemRepository interface {
	Create(ctx context.Context, ns *NamingSystem) error
	GetByID(ctx context.Context, id uuid.UUID) (*NamingSystem, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*NamingSystem, error)
	Update(ctx context.Context, ns *NamingSystem) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*NamingSystem, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*NamingSystem, int, error)
	// UniqueIDs
	AddUniqueID(ctx context.Context, uid *NamingSystemUniqueID) error
	GetUniqueIDs(ctx context.Context, namingSystemID uuid.UUID) ([]*NamingSystemUniqueID, error)
}

type OperationDefinitionRepository interface {
	Create(ctx context.Context, od *OperationDefinition) error
	GetByID(ctx context.Context, id uuid.UUID) (*OperationDefinition, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*OperationDefinition, error)
	Update(ctx context.Context, od *OperationDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*OperationDefinition, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*OperationDefinition, int, error)
	// Parameters
	AddParameter(ctx context.Context, p *OperationDefinitionParameter) error
	GetParameters(ctx context.Context, opDefID uuid.UUID) ([]*OperationDefinitionParameter, error)
}

type MessageDefinitionRepository interface {
	Create(ctx context.Context, md *MessageDefinition) error
	GetByID(ctx context.Context, id uuid.UUID) (*MessageDefinition, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MessageDefinition, error)
	Update(ctx context.Context, md *MessageDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*MessageDefinition, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MessageDefinition, int, error)
}

type MessageHeaderRepository interface {
	Create(ctx context.Context, mh *MessageHeader) error
	GetByID(ctx context.Context, id uuid.UUID) (*MessageHeader, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MessageHeader, error)
	Update(ctx context.Context, mh *MessageHeader) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*MessageHeader, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MessageHeader, int, error)
}
