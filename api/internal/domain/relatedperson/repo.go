package relatedperson

import (
	"context"

	"github.com/google/uuid"
)

// RelatedPersonRepository defines CRUD operations for RelatedPerson resources.
type RelatedPersonRepository interface {
	Create(ctx context.Context, rp *RelatedPerson) error
	GetByID(ctx context.Context, id uuid.UUID) (*RelatedPerson, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*RelatedPerson, error)
	Update(ctx context.Context, rp *RelatedPerson) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*RelatedPerson, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*RelatedPerson, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*RelatedPerson, int, error)
	// Communications
	AddCommunication(ctx context.Context, c *RelatedPersonCommunication) error
	GetCommunications(ctx context.Context, relatedPersonID uuid.UUID) ([]*RelatedPersonCommunication, error)
}
