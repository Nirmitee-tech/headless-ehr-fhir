package medication

import (
	"context"

	"github.com/google/uuid"
)

type MedicationRepository interface {
	Create(ctx context.Context, m *Medication) error
	GetByID(ctx context.Context, id uuid.UUID) (*Medication, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Medication, error)
	Update(ctx context.Context, m *Medication) error
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Medication, int, error)
	// Ingredients
	AddIngredient(ctx context.Context, ing *MedicationIngredient) error
	GetIngredients(ctx context.Context, medicationID uuid.UUID) ([]*MedicationIngredient, error)
	RemoveIngredient(ctx context.Context, id uuid.UUID) error
}

type MedicationRequestRepository interface {
	Create(ctx context.Context, mr *MedicationRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*MedicationRequest, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MedicationRequest, error)
	Update(ctx context.Context, mr *MedicationRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationRequest, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicationRequest, int, error)
}

type MedicationAdministrationRepository interface {
	Create(ctx context.Context, ma *MedicationAdministration) error
	GetByID(ctx context.Context, id uuid.UUID) (*MedicationAdministration, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MedicationAdministration, error)
	Update(ctx context.Context, ma *MedicationAdministration) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationAdministration, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicationAdministration, int, error)
}

type MedicationDispenseRepository interface {
	Create(ctx context.Context, md *MedicationDispense) error
	GetByID(ctx context.Context, id uuid.UUID) (*MedicationDispense, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MedicationDispense, error)
	Update(ctx context.Context, md *MedicationDispense) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationDispense, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicationDispense, int, error)
}

type MedicationStatementRepository interface {
	Create(ctx context.Context, ms *MedicationStatement) error
	GetByID(ctx context.Context, id uuid.UUID) (*MedicationStatement, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MedicationStatement, error)
	Update(ctx context.Context, ms *MedicationStatement) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationStatement, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicationStatement, int, error)
}
