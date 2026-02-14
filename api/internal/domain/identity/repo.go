package identity

import (
	"context"

	"github.com/google/uuid"
)

type PatientRepository interface {
	Create(ctx context.Context, p *Patient) error
	GetByID(ctx context.Context, id uuid.UUID) (*Patient, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Patient, error)
	GetByMRN(ctx context.Context, mrn string) (*Patient, error)
	Update(ctx context.Context, p *Patient) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Patient, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Patient, int, error)

	// Contacts
	AddContact(ctx context.Context, c *PatientContact) error
	GetContacts(ctx context.Context, patientID uuid.UUID) ([]*PatientContact, error)
	RemoveContact(ctx context.Context, id uuid.UUID) error

	// Identifiers
	AddIdentifier(ctx context.Context, ident *PatientIdentifier) error
	GetIdentifiers(ctx context.Context, patientID uuid.UUID) ([]*PatientIdentifier, error)
	RemoveIdentifier(ctx context.Context, id uuid.UUID) error
}

type PractitionerRepository interface {
	Create(ctx context.Context, p *Practitioner) error
	GetByID(ctx context.Context, id uuid.UUID) (*Practitioner, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Practitioner, error)
	GetByNPI(ctx context.Context, npi string) (*Practitioner, error)
	Update(ctx context.Context, p *Practitioner) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Practitioner, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Practitioner, int, error)

	// Roles
	AddRole(ctx context.Context, role *PractitionerRole) error
	GetRoles(ctx context.Context, practitionerID uuid.UUID) ([]*PractitionerRole, error)
	RemoveRole(ctx context.Context, id uuid.UUID) error
}
