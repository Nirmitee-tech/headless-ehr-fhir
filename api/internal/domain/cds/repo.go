package cds

import (
	"context"

	"github.com/google/uuid"
)

type CDSRuleRepository interface {
	Create(ctx context.Context, r *CDSRule) error
	GetByID(ctx context.Context, id uuid.UUID) (*CDSRule, error)
	Update(ctx context.Context, r *CDSRule) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*CDSRule, int, error)
}

type CDSAlertRepository interface {
	Create(ctx context.Context, a *CDSAlert) error
	GetByID(ctx context.Context, id uuid.UUID) (*CDSAlert, error)
	Update(ctx context.Context, a *CDSAlert) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*CDSAlert, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*CDSAlert, int, error)
	// Responses
	AddResponse(ctx context.Context, resp *CDSAlertResponse) error
	GetResponses(ctx context.Context, alertID uuid.UUID) ([]*CDSAlertResponse, error)
}

type DrugInteractionRepository interface {
	Create(ctx context.Context, d *DrugInteraction) error
	GetByID(ctx context.Context, id uuid.UUID) (*DrugInteraction, error)
	Update(ctx context.Context, d *DrugInteraction) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*DrugInteraction, int, error)
}

type OrderSetRepository interface {
	Create(ctx context.Context, o *OrderSet) error
	GetByID(ctx context.Context, id uuid.UUID) (*OrderSet, error)
	Update(ctx context.Context, o *OrderSet) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*OrderSet, int, error)
	// Sections
	AddSection(ctx context.Context, s *OrderSetSection) error
	GetSections(ctx context.Context, orderSetID uuid.UUID) ([]*OrderSetSection, error)
	// Items
	AddItem(ctx context.Context, item *OrderSetItem) error
	GetItems(ctx context.Context, sectionID uuid.UUID) ([]*OrderSetItem, error)
}

type ClinicalPathwayRepository interface {
	Create(ctx context.Context, p *ClinicalPathway) error
	GetByID(ctx context.Context, id uuid.UUID) (*ClinicalPathway, error)
	Update(ctx context.Context, p *ClinicalPathway) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*ClinicalPathway, int, error)
	// Phases
	AddPhase(ctx context.Context, phase *ClinicalPathwayPhase) error
	GetPhases(ctx context.Context, pathwayID uuid.UUID) ([]*ClinicalPathwayPhase, error)
}

type PatientPathwayEnrollmentRepository interface {
	Create(ctx context.Context, e *PatientPathwayEnrollment) error
	GetByID(ctx context.Context, id uuid.UUID) (*PatientPathwayEnrollment, error)
	Update(ctx context.Context, e *PatientPathwayEnrollment) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*PatientPathwayEnrollment, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PatientPathwayEnrollment, int, error)
}

type FormularyRepository interface {
	Create(ctx context.Context, f *Formulary) error
	GetByID(ctx context.Context, id uuid.UUID) (*Formulary, error)
	Update(ctx context.Context, f *Formulary) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Formulary, int, error)
	// Items
	AddItem(ctx context.Context, item *FormularyItem) error
	GetItems(ctx context.Context, formularyID uuid.UUID) ([]*FormularyItem, error)
}

type MedReconciliationRepository interface {
	Create(ctx context.Context, mr *MedicationReconciliation) error
	GetByID(ctx context.Context, id uuid.UUID) (*MedicationReconciliation, error)
	Update(ctx context.Context, mr *MedicationReconciliation) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*MedicationReconciliation, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationReconciliation, int, error)
	// Items
	AddItem(ctx context.Context, item *MedicationReconciliationItem) error
	GetItems(ctx context.Context, reconciliationID uuid.UUID) ([]*MedicationReconciliationItem, error)
}
