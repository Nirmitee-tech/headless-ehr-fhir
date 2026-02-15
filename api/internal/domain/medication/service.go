package medication

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	medications     MedicationRepository
	requests        MedicationRequestRepository
	administrations MedicationAdministrationRepository
	dispenses       MedicationDispenseRepository
	statements      MedicationStatementRepository
	vt              *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(
	meds MedicationRepository,
	reqs MedicationRequestRepository,
	admins MedicationAdministrationRepository,
	disps MedicationDispenseRepository,
	stmts MedicationStatementRepository,
) *Service {
	return &Service{
		medications:     meds,
		requests:        reqs,
		administrations: admins,
		dispenses:       disps,
		statements:      stmts,
	}
}

// -- Medication --

var validMedStatuses = map[string]bool{
	"active": true, "inactive": true, "entered-in-error": true,
}

func (s *Service) CreateMedication(ctx context.Context, m *Medication) error {
	if m.CodeValue == "" {
		return fmt.Errorf("code_value is required")
	}
	if m.CodeDisplay == "" {
		return fmt.Errorf("code_display is required")
	}
	if m.Status == "" {
		m.Status = "active"
	}
	if !validMedStatuses[m.Status] {
		return fmt.Errorf("invalid status: %s", m.Status)
	}
	return s.medications.Create(ctx, m)
}

func (s *Service) GetMedication(ctx context.Context, id uuid.UUID) (*Medication, error) {
	return s.medications.GetByID(ctx, id)
}

func (s *Service) GetMedicationByFHIRID(ctx context.Context, fhirID string) (*Medication, error) {
	return s.medications.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateMedication(ctx context.Context, m *Medication) error {
	if m.Status != "" && !validMedStatuses[m.Status] {
		return fmt.Errorf("invalid status: %s", m.Status)
	}
	return s.medications.Update(ctx, m)
}

func (s *Service) DeleteMedication(ctx context.Context, id uuid.UUID) error {
	return s.medications.Delete(ctx, id)
}

func (s *Service) SearchMedications(ctx context.Context, params map[string]string, limit, offset int) ([]*Medication, int, error) {
	return s.medications.Search(ctx, params, limit, offset)
}

func (s *Service) AddIngredient(ctx context.Context, ing *MedicationIngredient) error {
	if ing.MedicationID == uuid.Nil {
		return fmt.Errorf("medication_id is required")
	}
	if ing.ItemDisplay == "" {
		return fmt.Errorf("item_display is required")
	}
	return s.medications.AddIngredient(ctx, ing)
}

func (s *Service) GetIngredients(ctx context.Context, medicationID uuid.UUID) ([]*MedicationIngredient, error) {
	return s.medications.GetIngredients(ctx, medicationID)
}

func (s *Service) RemoveIngredient(ctx context.Context, id uuid.UUID) error {
	return s.medications.RemoveIngredient(ctx, id)
}

// -- MedicationRequest --

var validMedRequestStatuses = map[string]bool{
	"active": true, "on-hold": true, "cancelled": true, "completed": true,
	"entered-in-error": true, "stopped": true, "draft": true, "unknown": true,
}

var validMedRequestIntents = map[string]bool{
	"proposal": true, "plan": true, "order": true, "original-order": true,
	"reflex-order": true, "filler-order": true, "instance-order": true, "option": true,
}

func (s *Service) CreateMedicationRequest(ctx context.Context, mr *MedicationRequest) error {
	if mr.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if mr.MedicationID == uuid.Nil {
		return fmt.Errorf("medication_id is required")
	}
	if mr.RequesterID == uuid.Nil {
		return fmt.Errorf("requester_id is required")
	}
	if mr.Status == "" {
		mr.Status = "draft"
	}
	if !validMedRequestStatuses[mr.Status] {
		return fmt.Errorf("invalid status: %s", mr.Status)
	}
	if mr.Intent == "" {
		mr.Intent = "order"
	}
	if !validMedRequestIntents[mr.Intent] {
		return fmt.Errorf("invalid intent: %s", mr.Intent)
	}
	if err := s.requests.Create(ctx, mr); err != nil {
		return err
	}
	mr.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "MedicationRequest", mr.FHIRID, mr.ToFHIR())
	}
	return nil
}

func (s *Service) GetMedicationRequest(ctx context.Context, id uuid.UUID) (*MedicationRequest, error) {
	return s.requests.GetByID(ctx, id)
}

func (s *Service) GetMedicationRequestByFHIRID(ctx context.Context, fhirID string) (*MedicationRequest, error) {
	return s.requests.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateMedicationRequest(ctx context.Context, mr *MedicationRequest) error {
	if mr.Status != "" && !validMedRequestStatuses[mr.Status] {
		return fmt.Errorf("invalid status: %s", mr.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "MedicationRequest", mr.FHIRID, mr.VersionID, mr.ToFHIR())
		if err == nil {
			mr.VersionID = newVer
		}
	}
	return s.requests.Update(ctx, mr)
}

func (s *Service) DeleteMedicationRequest(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		mr, err := s.requests.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "MedicationRequest", mr.FHIRID, mr.VersionID)
		}
	}
	return s.requests.Delete(ctx, id)
}

func (s *Service) ListMedicationRequestsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationRequest, int, error) {
	return s.requests.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchMedicationRequests(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicationRequest, int, error) {
	return s.requests.Search(ctx, params, limit, offset)
}

// -- MedicationAdministration --

var validMedAdminStatuses = map[string]bool{
	"in-progress": true, "not-done": true, "on-hold": true, "completed": true,
	"entered-in-error": true, "stopped": true, "unknown": true,
}

func (s *Service) CreateMedicationAdministration(ctx context.Context, ma *MedicationAdministration) error {
	if ma.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if ma.MedicationID == uuid.Nil {
		return fmt.Errorf("medication_id is required")
	}
	if ma.Status == "" {
		ma.Status = "in-progress"
	}
	if !validMedAdminStatuses[ma.Status] {
		return fmt.Errorf("invalid status: %s", ma.Status)
	}
	if err := s.administrations.Create(ctx, ma); err != nil {
		return err
	}
	ma.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "MedicationAdministration", ma.FHIRID, ma.ToFHIR())
	}
	return nil
}

func (s *Service) GetMedicationAdministration(ctx context.Context, id uuid.UUID) (*MedicationAdministration, error) {
	return s.administrations.GetByID(ctx, id)
}

func (s *Service) GetMedicationAdministrationByFHIRID(ctx context.Context, fhirID string) (*MedicationAdministration, error) {
	return s.administrations.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateMedicationAdministration(ctx context.Context, ma *MedicationAdministration) error {
	if ma.Status != "" && !validMedAdminStatuses[ma.Status] {
		return fmt.Errorf("invalid status: %s", ma.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "MedicationAdministration", ma.FHIRID, ma.VersionID, ma.ToFHIR())
		if err == nil {
			ma.VersionID = newVer
		}
	}
	return s.administrations.Update(ctx, ma)
}

func (s *Service) DeleteMedicationAdministration(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		ma, err := s.administrations.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "MedicationAdministration", ma.FHIRID, ma.VersionID)
		}
	}
	return s.administrations.Delete(ctx, id)
}

func (s *Service) ListMedicationAdministrationsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationAdministration, int, error) {
	return s.administrations.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchMedicationAdministrations(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicationAdministration, int, error) {
	return s.administrations.Search(ctx, params, limit, offset)
}

// -- MedicationDispense --

var validMedDispenseStatuses = map[string]bool{
	"preparation": true, "in-progress": true, "cancelled": true, "on-hold": true,
	"completed": true, "entered-in-error": true, "stopped": true, "declined": true, "unknown": true,
}

func (s *Service) CreateMedicationDispense(ctx context.Context, md *MedicationDispense) error {
	if md.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if md.MedicationID == uuid.Nil {
		return fmt.Errorf("medication_id is required")
	}
	if md.Status == "" {
		md.Status = "preparation"
	}
	if !validMedDispenseStatuses[md.Status] {
		return fmt.Errorf("invalid status: %s", md.Status)
	}
	if err := s.dispenses.Create(ctx, md); err != nil {
		return err
	}
	md.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "MedicationDispense", md.FHIRID, md.ToFHIR())
	}
	return nil
}

func (s *Service) GetMedicationDispense(ctx context.Context, id uuid.UUID) (*MedicationDispense, error) {
	return s.dispenses.GetByID(ctx, id)
}

func (s *Service) GetMedicationDispenseByFHIRID(ctx context.Context, fhirID string) (*MedicationDispense, error) {
	return s.dispenses.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateMedicationDispense(ctx context.Context, md *MedicationDispense) error {
	if md.Status != "" && !validMedDispenseStatuses[md.Status] {
		return fmt.Errorf("invalid status: %s", md.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "MedicationDispense", md.FHIRID, md.VersionID, md.ToFHIR())
		if err == nil {
			md.VersionID = newVer
		}
	}
	return s.dispenses.Update(ctx, md)
}

func (s *Service) DeleteMedicationDispense(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		md, err := s.dispenses.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "MedicationDispense", md.FHIRID, md.VersionID)
		}
	}
	return s.dispenses.Delete(ctx, id)
}

func (s *Service) ListMedicationDispensesByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationDispense, int, error) {
	return s.dispenses.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchMedicationDispenses(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicationDispense, int, error) {
	return s.dispenses.Search(ctx, params, limit, offset)
}

// -- MedicationStatement --

var validMedStatementStatuses = map[string]bool{
	"active": true, "completed": true, "entered-in-error": true, "intended": true,
	"stopped": true, "on-hold": true, "unknown": true, "not-taken": true,
}

func (s *Service) CreateMedicationStatement(ctx context.Context, ms *MedicationStatement) error {
	if ms.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if ms.Status == "" {
		ms.Status = "active"
	}
	if !validMedStatementStatuses[ms.Status] {
		return fmt.Errorf("invalid status: %s", ms.Status)
	}
	return s.statements.Create(ctx, ms)
}

func (s *Service) GetMedicationStatement(ctx context.Context, id uuid.UUID) (*MedicationStatement, error) {
	return s.statements.GetByID(ctx, id)
}

func (s *Service) GetMedicationStatementByFHIRID(ctx context.Context, fhirID string) (*MedicationStatement, error) {
	return s.statements.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateMedicationStatement(ctx context.Context, ms *MedicationStatement) error {
	if ms.Status != "" && !validMedStatementStatuses[ms.Status] {
		return fmt.Errorf("invalid status: %s", ms.Status)
	}
	return s.statements.Update(ctx, ms)
}

func (s *Service) DeleteMedicationStatement(ctx context.Context, id uuid.UUID) error {
	return s.statements.Delete(ctx, id)
}

func (s *Service) ListMedicationStatementsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationStatement, int, error) {
	return s.statements.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchMedicationStatements(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicationStatement, int, error) {
	return s.statements.Search(ctx, params, limit, offset)
}
