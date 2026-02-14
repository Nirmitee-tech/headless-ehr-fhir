package billing

import (
	"context"

	"github.com/google/uuid"
)

type CoverageRepository interface {
	Create(ctx context.Context, c *Coverage) error
	GetByID(ctx context.Context, id uuid.UUID) (*Coverage, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Coverage, error)
	Update(ctx context.Context, c *Coverage) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Coverage, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Coverage, int, error)
}

type ClaimRepository interface {
	Create(ctx context.Context, c *Claim) error
	GetByID(ctx context.Context, id uuid.UUID) (*Claim, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Claim, error)
	Update(ctx context.Context, c *Claim) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Claim, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Claim, int, error)
	// Diagnoses
	AddDiagnosis(ctx context.Context, d *ClaimDiagnosis) error
	GetDiagnoses(ctx context.Context, claimID uuid.UUID) ([]*ClaimDiagnosis, error)
	// Procedures
	AddProcedure(ctx context.Context, p *ClaimProcedure) error
	GetProcedures(ctx context.Context, claimID uuid.UUID) ([]*ClaimProcedure, error)
	// Items
	AddItem(ctx context.Context, item *ClaimItem) error
	GetItems(ctx context.Context, claimID uuid.UUID) ([]*ClaimItem, error)
}

type ClaimResponseRepository interface {
	Create(ctx context.Context, cr *ClaimResponse) error
	GetByID(ctx context.Context, id uuid.UUID) (*ClaimResponse, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ClaimResponse, error)
	ListByClaim(ctx context.Context, claimID uuid.UUID, limit, offset int) ([]*ClaimResponse, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ClaimResponse, int, error)
}

type InvoiceRepository interface {
	Create(ctx context.Context, inv *Invoice) error
	GetByID(ctx context.Context, id uuid.UUID) (*Invoice, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Invoice, error)
	Update(ctx context.Context, inv *Invoice) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Invoice, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Invoice, int, error)
	// Line Items
	AddLineItem(ctx context.Context, li *InvoiceLineItem) error
	GetLineItems(ctx context.Context, invoiceID uuid.UUID) ([]*InvoiceLineItem, error)
}
