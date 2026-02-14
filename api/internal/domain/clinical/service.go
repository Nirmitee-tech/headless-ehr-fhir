package clinical

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	conditions   ConditionRepository
	observations ObservationRepository
	allergies    AllergyRepository
	procedures   ProcedureRepository
}

func NewService(cond ConditionRepository, obs ObservationRepository, allergy AllergyRepository, proc ProcedureRepository) *Service {
	return &Service{conditions: cond, observations: obs, allergies: allergy, procedures: proc}
}

// -- Condition --

var validClinicalStatuses = map[string]bool{
	"active": true, "recurrence": true, "relapse": true,
	"inactive": true, "remission": true, "resolved": true,
}

func (s *Service) CreateCondition(ctx context.Context, c *Condition) error {
	if c.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if c.CodeValue == "" {
		return fmt.Errorf("code_value is required")
	}
	if c.CodeDisplay == "" {
		return fmt.Errorf("code_display is required")
	}
	if c.ClinicalStatus == "" {
		c.ClinicalStatus = "active"
	}
	if !validClinicalStatuses[c.ClinicalStatus] {
		return fmt.Errorf("invalid clinical_status: %s", c.ClinicalStatus)
	}
	return s.conditions.Create(ctx, c)
}

func (s *Service) GetCondition(ctx context.Context, id uuid.UUID) (*Condition, error) {
	return s.conditions.GetByID(ctx, id)
}

func (s *Service) GetConditionByFHIRID(ctx context.Context, fhirID string) (*Condition, error) {
	return s.conditions.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateCondition(ctx context.Context, c *Condition) error {
	if c.ClinicalStatus != "" && !validClinicalStatuses[c.ClinicalStatus] {
		return fmt.Errorf("invalid clinical_status: %s", c.ClinicalStatus)
	}
	return s.conditions.Update(ctx, c)
}

func (s *Service) DeleteCondition(ctx context.Context, id uuid.UUID) error {
	return s.conditions.Delete(ctx, id)
}

func (s *Service) ListConditionsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Condition, int, error) {
	return s.conditions.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchConditions(ctx context.Context, params map[string]string, limit, offset int) ([]*Condition, int, error) {
	return s.conditions.Search(ctx, params, limit, offset)
}

// -- Observation --

var validObsStatuses = map[string]bool{
	"registered": true, "preliminary": true, "final": true, "amended": true,
	"corrected": true, "cancelled": true, "entered-in-error": true, "unknown": true,
}

func (s *Service) CreateObservation(ctx context.Context, o *Observation) error {
	if o.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if o.CodeValue == "" {
		return fmt.Errorf("code_value is required")
	}
	if o.Status == "" {
		o.Status = "final"
	}
	if !validObsStatuses[o.Status] {
		return fmt.Errorf("invalid status: %s", o.Status)
	}
	return s.observations.Create(ctx, o)
}

func (s *Service) GetObservation(ctx context.Context, id uuid.UUID) (*Observation, error) {
	return s.observations.GetByID(ctx, id)
}

func (s *Service) GetObservationByFHIRID(ctx context.Context, fhirID string) (*Observation, error) {
	return s.observations.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateObservation(ctx context.Context, o *Observation) error {
	return s.observations.Update(ctx, o)
}

func (s *Service) DeleteObservation(ctx context.Context, id uuid.UUID) error {
	return s.observations.Delete(ctx, id)
}

func (s *Service) ListObservationsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Observation, int, error) {
	return s.observations.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchObservations(ctx context.Context, params map[string]string, limit, offset int) ([]*Observation, int, error) {
	return s.observations.Search(ctx, params, limit, offset)
}

func (s *Service) AddObservationComponent(ctx context.Context, c *ObservationComponent) error {
	if c.ObservationID == uuid.Nil {
		return fmt.Errorf("observation_id is required")
	}
	if c.CodeValue == "" {
		return fmt.Errorf("code_value is required")
	}
	return s.observations.AddComponent(ctx, c)
}

func (s *Service) GetObservationComponents(ctx context.Context, observationID uuid.UUID) ([]*ObservationComponent, error) {
	return s.observations.GetComponents(ctx, observationID)
}

// -- AllergyIntolerance --

func (s *Service) CreateAllergy(ctx context.Context, a *AllergyIntolerance) error {
	if a.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if a.ClinicalStatus == nil {
		status := "active"
		a.ClinicalStatus = &status
	}
	return s.allergies.Create(ctx, a)
}

func (s *Service) GetAllergy(ctx context.Context, id uuid.UUID) (*AllergyIntolerance, error) {
	return s.allergies.GetByID(ctx, id)
}

func (s *Service) GetAllergyByFHIRID(ctx context.Context, fhirID string) (*AllergyIntolerance, error) {
	return s.allergies.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateAllergy(ctx context.Context, a *AllergyIntolerance) error {
	return s.allergies.Update(ctx, a)
}

func (s *Service) DeleteAllergy(ctx context.Context, id uuid.UUID) error {
	return s.allergies.Delete(ctx, id)
}

func (s *Service) ListAllergiesByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*AllergyIntolerance, int, error) {
	return s.allergies.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchAllergies(ctx context.Context, params map[string]string, limit, offset int) ([]*AllergyIntolerance, int, error) {
	return s.allergies.Search(ctx, params, limit, offset)
}

func (s *Service) AddAllergyReaction(ctx context.Context, r *AllergyReaction) error {
	if r.AllergyID == uuid.Nil {
		return fmt.Errorf("allergy_id is required")
	}
	if r.ManifestationCode == "" {
		return fmt.Errorf("manifestation_code is required")
	}
	return s.allergies.AddReaction(ctx, r)
}

func (s *Service) GetAllergyReactions(ctx context.Context, allergyID uuid.UUID) ([]*AllergyReaction, error) {
	return s.allergies.GetReactions(ctx, allergyID)
}

func (s *Service) RemoveAllergyReaction(ctx context.Context, id uuid.UUID) error {
	return s.allergies.RemoveReaction(ctx, id)
}

// -- Procedure --

var validProcStatuses = map[string]bool{
	"preparation": true, "in-progress": true, "not-done": true, "on-hold": true,
	"stopped": true, "completed": true, "entered-in-error": true, "unknown": true,
}

func (s *Service) CreateProcedure(ctx context.Context, p *ProcedureRecord) error {
	if p.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if p.CodeValue == "" {
		return fmt.Errorf("code_value is required")
	}
	if p.Status == "" {
		p.Status = "completed"
	}
	if !validProcStatuses[p.Status] {
		return fmt.Errorf("invalid status: %s", p.Status)
	}
	return s.procedures.Create(ctx, p)
}

func (s *Service) GetProcedure(ctx context.Context, id uuid.UUID) (*ProcedureRecord, error) {
	return s.procedures.GetByID(ctx, id)
}

func (s *Service) GetProcedureByFHIRID(ctx context.Context, fhirID string) (*ProcedureRecord, error) {
	return s.procedures.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateProcedure(ctx context.Context, p *ProcedureRecord) error {
	return s.procedures.Update(ctx, p)
}

func (s *Service) DeleteProcedure(ctx context.Context, id uuid.UUID) error {
	return s.procedures.Delete(ctx, id)
}

func (s *Service) ListProceduresByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ProcedureRecord, int, error) {
	return s.procedures.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchProcedures(ctx context.Context, params map[string]string, limit, offset int) ([]*ProcedureRecord, int, error) {
	return s.procedures.Search(ctx, params, limit, offset)
}

func (s *Service) AddProcedurePerformer(ctx context.Context, pf *ProcedurePerformer) error {
	if pf.ProcedureID == uuid.Nil {
		return fmt.Errorf("procedure_id is required")
	}
	if pf.PractitionerID == uuid.Nil {
		return fmt.Errorf("practitioner_id is required")
	}
	return s.procedures.AddPerformer(ctx, pf)
}

func (s *Service) GetProcedurePerformers(ctx context.Context, procedureID uuid.UUID) ([]*ProcedurePerformer, error) {
	return s.procedures.GetPerformers(ctx, procedureID)
}

func (s *Service) RemoveProcedurePerformer(ctx context.Context, id uuid.UUID) error {
	return s.procedures.RemovePerformer(ctx, id)
}
