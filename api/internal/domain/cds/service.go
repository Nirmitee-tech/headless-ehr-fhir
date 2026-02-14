package cds

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	rules       CDSRuleRepository
	alerts      CDSAlertRepository
	interactions DrugInteractionRepository
	orderSets   OrderSetRepository
	pathways    ClinicalPathwayRepository
	enrollments PatientPathwayEnrollmentRepository
	formularies FormularyRepository
	medReconc   MedReconciliationRepository
}

func NewService(
	rules CDSRuleRepository,
	alerts CDSAlertRepository,
	interactions DrugInteractionRepository,
	orderSets OrderSetRepository,
	pathways ClinicalPathwayRepository,
	enrollments PatientPathwayEnrollmentRepository,
	formularies FormularyRepository,
	medReconc MedReconciliationRepository,
) *Service {
	return &Service{
		rules:       rules,
		alerts:      alerts,
		interactions: interactions,
		orderSets:   orderSets,
		pathways:    pathways,
		enrollments: enrollments,
		formularies: formularies,
		medReconc:   medReconc,
	}
}

// -- CDS Rule --

func (s *Service) CreateCDSRule(ctx context.Context, r *CDSRule) error {
	if r.RuleName == "" {
		return fmt.Errorf("rule_name is required")
	}
	if r.RuleType == "" {
		return fmt.Errorf("rule_type is required")
	}
	return s.rules.Create(ctx, r)
}

func (s *Service) GetCDSRule(ctx context.Context, id uuid.UUID) (*CDSRule, error) {
	return s.rules.GetByID(ctx, id)
}

func (s *Service) UpdateCDSRule(ctx context.Context, r *CDSRule) error {
	return s.rules.Update(ctx, r)
}

func (s *Service) DeleteCDSRule(ctx context.Context, id uuid.UUID) error {
	return s.rules.Delete(ctx, id)
}

func (s *Service) ListCDSRules(ctx context.Context, limit, offset int) ([]*CDSRule, int, error) {
	return s.rules.List(ctx, limit, offset)
}

// -- CDS Alert --

var validAlertStatuses = map[string]bool{
	"fired": true, "accepted": true, "overridden": true,
	"auto-resolved": true, "expired": true, "suppressed": true,
}

func (s *Service) CreateCDSAlert(ctx context.Context, a *CDSAlert) error {
	if a.RuleID == uuid.Nil {
		return fmt.Errorf("rule_id is required")
	}
	if a.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if a.Summary == "" {
		return fmt.Errorf("summary is required")
	}
	if a.Status == "" {
		a.Status = "fired"
	}
	if !validAlertStatuses[a.Status] {
		return fmt.Errorf("invalid status: %s", a.Status)
	}
	return s.alerts.Create(ctx, a)
}

func (s *Service) GetCDSAlert(ctx context.Context, id uuid.UUID) (*CDSAlert, error) {
	return s.alerts.GetByID(ctx, id)
}

func (s *Service) UpdateCDSAlert(ctx context.Context, a *CDSAlert) error {
	if a.Status != "" && !validAlertStatuses[a.Status] {
		return fmt.Errorf("invalid status: %s", a.Status)
	}
	return s.alerts.Update(ctx, a)
}

func (s *Service) DeleteCDSAlert(ctx context.Context, id uuid.UUID) error {
	return s.alerts.Delete(ctx, id)
}

func (s *Service) ListCDSAlerts(ctx context.Context, limit, offset int) ([]*CDSAlert, int, error) {
	return s.alerts.List(ctx, limit, offset)
}

func (s *Service) ListCDSAlertsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*CDSAlert, int, error) {
	return s.alerts.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) AddAlertResponse(ctx context.Context, resp *CDSAlertResponse) error {
	if resp.AlertID == uuid.Nil {
		return fmt.Errorf("alert_id is required")
	}
	if resp.PractitionerID == uuid.Nil {
		return fmt.Errorf("practitioner_id is required")
	}
	if resp.Action == "" {
		return fmt.Errorf("action is required")
	}
	return s.alerts.AddResponse(ctx, resp)
}

func (s *Service) GetAlertResponses(ctx context.Context, alertID uuid.UUID) ([]*CDSAlertResponse, error) {
	return s.alerts.GetResponses(ctx, alertID)
}

// -- Drug Interaction --

func (s *Service) CreateDrugInteraction(ctx context.Context, d *DrugInteraction) error {
	if d.MedicationAName == "" {
		return fmt.Errorf("medication_a_name is required")
	}
	if d.MedicationBName == "" {
		return fmt.Errorf("medication_b_name is required")
	}
	if d.Severity == "" {
		return fmt.Errorf("severity is required")
	}
	return s.interactions.Create(ctx, d)
}

func (s *Service) GetDrugInteraction(ctx context.Context, id uuid.UUID) (*DrugInteraction, error) {
	return s.interactions.GetByID(ctx, id)
}

func (s *Service) UpdateDrugInteraction(ctx context.Context, d *DrugInteraction) error {
	return s.interactions.Update(ctx, d)
}

func (s *Service) DeleteDrugInteraction(ctx context.Context, id uuid.UUID) error {
	return s.interactions.Delete(ctx, id)
}

func (s *Service) ListDrugInteractions(ctx context.Context, limit, offset int) ([]*DrugInteraction, int, error) {
	return s.interactions.List(ctx, limit, offset)
}

// -- Order Set --

func (s *Service) CreateOrderSet(ctx context.Context, o *OrderSet) error {
	if o.Name == "" {
		return fmt.Errorf("name is required")
	}
	if o.Status == "" {
		o.Status = "draft"
	}
	return s.orderSets.Create(ctx, o)
}

func (s *Service) GetOrderSet(ctx context.Context, id uuid.UUID) (*OrderSet, error) {
	return s.orderSets.GetByID(ctx, id)
}

func (s *Service) UpdateOrderSet(ctx context.Context, o *OrderSet) error {
	return s.orderSets.Update(ctx, o)
}

func (s *Service) DeleteOrderSet(ctx context.Context, id uuid.UUID) error {
	return s.orderSets.Delete(ctx, id)
}

func (s *Service) ListOrderSets(ctx context.Context, limit, offset int) ([]*OrderSet, int, error) {
	return s.orderSets.List(ctx, limit, offset)
}

func (s *Service) AddOrderSetSection(ctx context.Context, sec *OrderSetSection) error {
	if sec.OrderSetID == uuid.Nil {
		return fmt.Errorf("order_set_id is required")
	}
	if sec.Name == "" {
		return fmt.Errorf("name is required")
	}
	return s.orderSets.AddSection(ctx, sec)
}

func (s *Service) GetOrderSetSections(ctx context.Context, orderSetID uuid.UUID) ([]*OrderSetSection, error) {
	return s.orderSets.GetSections(ctx, orderSetID)
}

func (s *Service) AddOrderSetItem(ctx context.Context, item *OrderSetItem) error {
	if item.SectionID == uuid.Nil {
		return fmt.Errorf("section_id is required")
	}
	if item.ItemName == "" {
		return fmt.Errorf("item_name is required")
	}
	return s.orderSets.AddItem(ctx, item)
}

func (s *Service) GetOrderSetItems(ctx context.Context, sectionID uuid.UUID) ([]*OrderSetItem, error) {
	return s.orderSets.GetItems(ctx, sectionID)
}

// -- Clinical Pathway --

func (s *Service) CreateClinicalPathway(ctx context.Context, p *ClinicalPathway) error {
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	return s.pathways.Create(ctx, p)
}

func (s *Service) GetClinicalPathway(ctx context.Context, id uuid.UUID) (*ClinicalPathway, error) {
	return s.pathways.GetByID(ctx, id)
}

func (s *Service) UpdateClinicalPathway(ctx context.Context, p *ClinicalPathway) error {
	return s.pathways.Update(ctx, p)
}

func (s *Service) DeleteClinicalPathway(ctx context.Context, id uuid.UUID) error {
	return s.pathways.Delete(ctx, id)
}

func (s *Service) ListClinicalPathways(ctx context.Context, limit, offset int) ([]*ClinicalPathway, int, error) {
	return s.pathways.List(ctx, limit, offset)
}

func (s *Service) AddPathwayPhase(ctx context.Context, phase *ClinicalPathwayPhase) error {
	if phase.PathwayID == uuid.Nil {
		return fmt.Errorf("pathway_id is required")
	}
	if phase.Name == "" {
		return fmt.Errorf("name is required")
	}
	return s.pathways.AddPhase(ctx, phase)
}

func (s *Service) GetPathwayPhases(ctx context.Context, pathwayID uuid.UUID) ([]*ClinicalPathwayPhase, error) {
	return s.pathways.GetPhases(ctx, pathwayID)
}

// -- Patient Pathway Enrollment --

var validEnrollmentStatuses = map[string]bool{
	"active": true, "completed": true, "withdrawn": true, "deviated": true,
}

func (s *Service) CreatePathwayEnrollment(ctx context.Context, e *PatientPathwayEnrollment) error {
	if e.PathwayID == uuid.Nil {
		return fmt.Errorf("pathway_id is required")
	}
	if e.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if e.Status == "" {
		e.Status = "active"
	}
	if !validEnrollmentStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	return s.enrollments.Create(ctx, e)
}

func (s *Service) GetPathwayEnrollment(ctx context.Context, id uuid.UUID) (*PatientPathwayEnrollment, error) {
	return s.enrollments.GetByID(ctx, id)
}

func (s *Service) UpdatePathwayEnrollment(ctx context.Context, e *PatientPathwayEnrollment) error {
	if e.Status != "" && !validEnrollmentStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	return s.enrollments.Update(ctx, e)
}

func (s *Service) DeletePathwayEnrollment(ctx context.Context, id uuid.UUID) error {
	return s.enrollments.Delete(ctx, id)
}

func (s *Service) ListPathwayEnrollments(ctx context.Context, limit, offset int) ([]*PatientPathwayEnrollment, int, error) {
	return s.enrollments.List(ctx, limit, offset)
}

func (s *Service) ListPathwayEnrollmentsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PatientPathwayEnrollment, int, error) {
	return s.enrollments.ListByPatient(ctx, patientID, limit, offset)
}

// -- Formulary --

func (s *Service) CreateFormulary(ctx context.Context, f *Formulary) error {
	if f.Name == "" {
		return fmt.Errorf("name is required")
	}
	return s.formularies.Create(ctx, f)
}

func (s *Service) GetFormulary(ctx context.Context, id uuid.UUID) (*Formulary, error) {
	return s.formularies.GetByID(ctx, id)
}

func (s *Service) UpdateFormulary(ctx context.Context, f *Formulary) error {
	return s.formularies.Update(ctx, f)
}

func (s *Service) DeleteFormulary(ctx context.Context, id uuid.UUID) error {
	return s.formularies.Delete(ctx, id)
}

func (s *Service) ListFormularies(ctx context.Context, limit, offset int) ([]*Formulary, int, error) {
	return s.formularies.List(ctx, limit, offset)
}

func (s *Service) AddFormularyItem(ctx context.Context, item *FormularyItem) error {
	if item.FormularyID == uuid.Nil {
		return fmt.Errorf("formulary_id is required")
	}
	if item.MedicationName == "" {
		return fmt.Errorf("medication_name is required")
	}
	return s.formularies.AddItem(ctx, item)
}

func (s *Service) GetFormularyItems(ctx context.Context, formularyID uuid.UUID) ([]*FormularyItem, error) {
	return s.formularies.GetItems(ctx, formularyID)
}

// -- Medication Reconciliation --

var validMedReconcStatuses = map[string]bool{
	"in-progress": true, "completed": true, "pending-verification": true,
}

func (s *Service) CreateMedReconciliation(ctx context.Context, mr *MedicationReconciliation) error {
	if mr.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if mr.Status == "" {
		mr.Status = "in-progress"
	}
	if !validMedReconcStatuses[mr.Status] {
		return fmt.Errorf("invalid status: %s", mr.Status)
	}
	return s.medReconc.Create(ctx, mr)
}

func (s *Service) GetMedReconciliation(ctx context.Context, id uuid.UUID) (*MedicationReconciliation, error) {
	return s.medReconc.GetByID(ctx, id)
}

func (s *Service) UpdateMedReconciliation(ctx context.Context, mr *MedicationReconciliation) error {
	if mr.Status != "" && !validMedReconcStatuses[mr.Status] {
		return fmt.Errorf("invalid status: %s", mr.Status)
	}
	return s.medReconc.Update(ctx, mr)
}

func (s *Service) DeleteMedReconciliation(ctx context.Context, id uuid.UUID) error {
	return s.medReconc.Delete(ctx, id)
}

func (s *Service) ListMedReconciliations(ctx context.Context, limit, offset int) ([]*MedicationReconciliation, int, error) {
	return s.medReconc.List(ctx, limit, offset)
}

func (s *Service) ListMedReconciliationsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationReconciliation, int, error) {
	return s.medReconc.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) AddMedReconciliationItem(ctx context.Context, item *MedicationReconciliationItem) error {
	if item.ReconciliationID == uuid.Nil {
		return fmt.Errorf("reconciliation_id is required")
	}
	if item.MedicationName == "" {
		return fmt.Errorf("medication_name is required")
	}
	return s.medReconc.AddItem(ctx, item)
}

func (s *Service) GetMedReconciliationItems(ctx context.Context, reconciliationID uuid.UUID) ([]*MedicationReconciliationItem, error) {
	return s.medReconc.GetItems(ctx, reconciliationID)
}
