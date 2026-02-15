package careteam

import (
	"context"

	"github.com/google/uuid"
)

type CareTeamRepository interface {
	Create(ctx context.Context, ct *CareTeam) error
	GetByID(ctx context.Context, id uuid.UUID) (*CareTeam, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*CareTeam, error)
	Update(ctx context.Context, ct *CareTeam) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*CareTeam, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*CareTeam, int, error)
	// Participants
	AddParticipant(ctx context.Context, careTeamID uuid.UUID, p *CareTeamParticipant) error
	RemoveParticipant(ctx context.Context, careTeamID uuid.UUID, participantID uuid.UUID) error
	GetParticipants(ctx context.Context, careTeamID uuid.UUID) ([]*CareTeamParticipant, error)
}
