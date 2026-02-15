package fhirlist

import (
	"context"

	"github.com/google/uuid"
)

type FHIRListRepository interface {
	Create(ctx context.Context, l *FHIRList) error
	GetByID(ctx context.Context, id uuid.UUID) (*FHIRList, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*FHIRList, error)
	Update(ctx context.Context, l *FHIRList) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*FHIRList, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*FHIRList, int, error)
	// Entries
	AddEntry(ctx context.Context, entry *FHIRListEntry) error
	GetEntries(ctx context.Context, listID uuid.UUID) ([]*FHIRListEntry, error)
}
