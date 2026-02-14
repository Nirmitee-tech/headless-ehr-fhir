package documents

import (
	"context"

	"github.com/google/uuid"
)

type ConsentRepository interface {
	Create(ctx context.Context, c *Consent) error
	GetByID(ctx context.Context, id uuid.UUID) (*Consent, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Consent, error)
	Update(ctx context.Context, c *Consent) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Consent, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Consent, int, error)
}

type DocumentReferenceRepository interface {
	Create(ctx context.Context, d *DocumentReference) error
	GetByID(ctx context.Context, id uuid.UUID) (*DocumentReference, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*DocumentReference, error)
	Update(ctx context.Context, d *DocumentReference) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*DocumentReference, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*DocumentReference, int, error)
}

type ClinicalNoteRepository interface {
	Create(ctx context.Context, n *ClinicalNote) error
	GetByID(ctx context.Context, id uuid.UUID) (*ClinicalNote, error)
	Update(ctx context.Context, n *ClinicalNote) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ClinicalNote, int, error)
	ListByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*ClinicalNote, int, error)
}

type CompositionRepository interface {
	Create(ctx context.Context, c *Composition) error
	GetByID(ctx context.Context, id uuid.UUID) (*Composition, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Composition, error)
	Update(ctx context.Context, c *Composition) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Composition, int, error)
	// Sections
	AddSection(ctx context.Context, s *CompositionSection) error
	GetSections(ctx context.Context, compositionID uuid.UUID) ([]*CompositionSection, error)
}
