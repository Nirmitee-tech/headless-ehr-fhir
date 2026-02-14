package provenance

import (
	"context"

	"github.com/google/uuid"
)

// ProvenanceRepository defines CRUD operations for Provenance resources.
type ProvenanceRepository interface {
	Create(ctx context.Context, p *Provenance) error
	GetByID(ctx context.Context, id uuid.UUID) (*Provenance, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Provenance, error)
	Update(ctx context.Context, p *Provenance) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Provenance, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Provenance, int, error)
	// Agents
	AddAgent(ctx context.Context, a *ProvenanceAgent) error
	GetAgents(ctx context.Context, provenanceID uuid.UUID) ([]*ProvenanceAgent, error)
	// Entities
	AddEntity(ctx context.Context, e *ProvenanceEntity) error
	GetEntities(ctx context.Context, provenanceID uuid.UUID) ([]*ProvenanceEntity, error)
}
