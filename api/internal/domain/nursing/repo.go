package nursing

import (
	"context"

	"github.com/google/uuid"
)

type FlowsheetTemplateRepository interface {
	Create(ctx context.Context, t *FlowsheetTemplate) error
	GetByID(ctx context.Context, id uuid.UUID) (*FlowsheetTemplate, error)
	Update(ctx context.Context, t *FlowsheetTemplate) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*FlowsheetTemplate, int, error)
	// Rows
	AddRow(ctx context.Context, r *FlowsheetRow) error
	GetRows(ctx context.Context, templateID uuid.UUID) ([]*FlowsheetRow, error)
}

type FlowsheetEntryRepository interface {
	Create(ctx context.Context, e *FlowsheetEntry) error
	GetByID(ctx context.Context, id uuid.UUID) (*FlowsheetEntry, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*FlowsheetEntry, int, error)
	ListByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*FlowsheetEntry, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*FlowsheetEntry, int, error)
}

type NursingAssessmentRepository interface {
	Create(ctx context.Context, a *NursingAssessment) error
	GetByID(ctx context.Context, id uuid.UUID) (*NursingAssessment, error)
	Update(ctx context.Context, a *NursingAssessment) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*NursingAssessment, int, error)
	ListByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*NursingAssessment, int, error)
}

type FallRiskRepository interface {
	Create(ctx context.Context, a *FallRiskAssessment) error
	GetByID(ctx context.Context, id uuid.UUID) (*FallRiskAssessment, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*FallRiskAssessment, int, error)
}

type SkinAssessmentRepository interface {
	Create(ctx context.Context, a *SkinAssessment) error
	GetByID(ctx context.Context, id uuid.UUID) (*SkinAssessment, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*SkinAssessment, int, error)
}

type PainAssessmentRepository interface {
	Create(ctx context.Context, a *PainAssessment) error
	GetByID(ctx context.Context, id uuid.UUID) (*PainAssessment, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PainAssessment, int, error)
}

type LinesDrainsRepository interface {
	Create(ctx context.Context, l *LinesDrainsAirways) error
	GetByID(ctx context.Context, id uuid.UUID) (*LinesDrainsAirways, error)
	Update(ctx context.Context, l *LinesDrainsAirways) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*LinesDrainsAirways, int, error)
	ListByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*LinesDrainsAirways, int, error)
}

type RestraintRepository interface {
	Create(ctx context.Context, r *RestraintRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*RestraintRecord, error)
	Update(ctx context.Context, r *RestraintRecord) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*RestraintRecord, int, error)
}

type IntakeOutputRepository interface {
	Create(ctx context.Context, r *IntakeOutputRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*IntakeOutputRecord, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*IntakeOutputRecord, int, error)
	ListByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*IntakeOutputRecord, int, error)
}
