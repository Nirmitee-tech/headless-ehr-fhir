package encounter

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, enc *Encounter) error
	GetByID(ctx context.Context, id uuid.UUID) (*Encounter, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Encounter, error)
	Update(ctx context.Context, enc *Encounter) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Encounter, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Encounter, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Encounter, int, error)

	// Participants
	AddParticipant(ctx context.Context, p *EncounterParticipant) error
	GetParticipants(ctx context.Context, encounterID uuid.UUID) ([]*EncounterParticipant, error)
	RemoveParticipant(ctx context.Context, id uuid.UUID) error

	// Diagnoses
	AddDiagnosis(ctx context.Context, d *EncounterDiagnosis) error
	GetDiagnoses(ctx context.Context, encounterID uuid.UUID) ([]*EncounterDiagnosis, error)
	RemoveDiagnosis(ctx context.Context, id uuid.UUID) error

	// Status History
	AddStatusHistory(ctx context.Context, sh *EncounterStatusHistory) error
	GetStatusHistory(ctx context.Context, encounterID uuid.UUID) ([]*EncounterStatusHistory, error)
}
