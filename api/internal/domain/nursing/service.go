package nursing

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	templates   FlowsheetTemplateRepository
	entries     FlowsheetEntryRepository
	assessments NursingAssessmentRepository
	fallRisk    FallRiskRepository
	skin        SkinAssessmentRepository
	pain        PainAssessmentRepository
	linesDrains LinesDrainsRepository
	restraints  RestraintRepository
	intakeOutput IntakeOutputRepository
	vt           *fhir.VersionTracker
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
	templates FlowsheetTemplateRepository,
	entries FlowsheetEntryRepository,
	assessments NursingAssessmentRepository,
	fallRisk FallRiskRepository,
	skin SkinAssessmentRepository,
	pain PainAssessmentRepository,
	linesDrains LinesDrainsRepository,
	restraints RestraintRepository,
	intakeOutput IntakeOutputRepository,
) *Service {
	return &Service{
		templates:    templates,
		entries:      entries,
		assessments:  assessments,
		fallRisk:     fallRisk,
		skin:         skin,
		pain:         pain,
		linesDrains:  linesDrains,
		restraints:   restraints,
		intakeOutput: intakeOutput,
	}
}

// -- Flowsheet Template --

func (s *Service) CreateTemplate(ctx context.Context, t *FlowsheetTemplate) error {
	if t.Name == "" {
		return fmt.Errorf("name is required")
	}
	return s.templates.Create(ctx, t)
}

func (s *Service) GetTemplate(ctx context.Context, id uuid.UUID) (*FlowsheetTemplate, error) {
	return s.templates.GetByID(ctx, id)
}

func (s *Service) UpdateTemplate(ctx context.Context, t *FlowsheetTemplate) error {
	return s.templates.Update(ctx, t)
}

func (s *Service) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	return s.templates.Delete(ctx, id)
}

func (s *Service) ListTemplates(ctx context.Context, limit, offset int) ([]*FlowsheetTemplate, int, error) {
	return s.templates.List(ctx, limit, offset)
}

func (s *Service) AddTemplateRow(ctx context.Context, r *FlowsheetRow) error {
	if r.TemplateID == uuid.Nil {
		return fmt.Errorf("template_id is required")
	}
	if r.Label == "" {
		return fmt.Errorf("label is required")
	}
	return s.templates.AddRow(ctx, r)
}

func (s *Service) GetTemplateRows(ctx context.Context, templateID uuid.UUID) ([]*FlowsheetRow, error) {
	return s.templates.GetRows(ctx, templateID)
}

// -- Flowsheet Entry --

func (s *Service) CreateEntry(ctx context.Context, e *FlowsheetEntry) error {
	if e.TemplateID == uuid.Nil {
		return fmt.Errorf("template_id is required")
	}
	if e.RowID == uuid.Nil {
		return fmt.Errorf("row_id is required")
	}
	if e.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if e.EncounterID == uuid.Nil {
		return fmt.Errorf("encounter_id is required")
	}
	if e.RecordedByID == uuid.Nil {
		return fmt.Errorf("recorded_by_id is required")
	}
	return s.entries.Create(ctx, e)
}

func (s *Service) GetEntry(ctx context.Context, id uuid.UUID) (*FlowsheetEntry, error) {
	return s.entries.GetByID(ctx, id)
}

func (s *Service) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	return s.entries.Delete(ctx, id)
}

func (s *Service) ListEntriesByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*FlowsheetEntry, int, error) {
	return s.entries.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) ListEntriesByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*FlowsheetEntry, int, error) {
	return s.entries.ListByEncounter(ctx, encounterID, limit, offset)
}

func (s *Service) SearchEntries(ctx context.Context, params map[string]string, limit, offset int) ([]*FlowsheetEntry, int, error) {
	return s.entries.Search(ctx, params, limit, offset)
}

// -- Nursing Assessment --

func (s *Service) CreateAssessment(ctx context.Context, a *NursingAssessment) error {
	if a.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if a.EncounterID == uuid.Nil {
		return fmt.Errorf("encounter_id is required")
	}
	if a.NurseID == uuid.Nil {
		return fmt.Errorf("nurse_id is required")
	}
	if a.AssessmentType == "" {
		return fmt.Errorf("assessment_type is required")
	}
	if a.Status == "" {
		a.Status = "in-progress"
	}
	return s.assessments.Create(ctx, a)
}

func (s *Service) GetAssessment(ctx context.Context, id uuid.UUID) (*NursingAssessment, error) {
	return s.assessments.GetByID(ctx, id)
}

func (s *Service) UpdateAssessment(ctx context.Context, a *NursingAssessment) error {
	return s.assessments.Update(ctx, a)
}

func (s *Service) DeleteAssessment(ctx context.Context, id uuid.UUID) error {
	return s.assessments.Delete(ctx, id)
}

func (s *Service) ListAssessmentsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*NursingAssessment, int, error) {
	return s.assessments.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) ListAssessmentsByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*NursingAssessment, int, error) {
	return s.assessments.ListByEncounter(ctx, encounterID, limit, offset)
}

// -- Fall Risk Assessment --

func (s *Service) CreateFallRisk(ctx context.Context, a *FallRiskAssessment) error {
	if a.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if a.AssessedByID == uuid.Nil {
		return fmt.Errorf("assessed_by_id is required")
	}
	return s.fallRisk.Create(ctx, a)
}

func (s *Service) GetFallRisk(ctx context.Context, id uuid.UUID) (*FallRiskAssessment, error) {
	return s.fallRisk.GetByID(ctx, id)
}

func (s *Service) ListFallRiskByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*FallRiskAssessment, int, error) {
	return s.fallRisk.ListByPatient(ctx, patientID, limit, offset)
}

// -- Skin Assessment --

func (s *Service) CreateSkinAssessment(ctx context.Context, a *SkinAssessment) error {
	if a.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if a.AssessedByID == uuid.Nil {
		return fmt.Errorf("assessed_by_id is required")
	}
	return s.skin.Create(ctx, a)
}

func (s *Service) GetSkinAssessment(ctx context.Context, id uuid.UUID) (*SkinAssessment, error) {
	return s.skin.GetByID(ctx, id)
}

func (s *Service) ListSkinAssessmentsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*SkinAssessment, int, error) {
	return s.skin.ListByPatient(ctx, patientID, limit, offset)
}

// -- Pain Assessment --

func (s *Service) CreatePainAssessment(ctx context.Context, a *PainAssessment) error {
	if a.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if a.AssessedByID == uuid.Nil {
		return fmt.Errorf("assessed_by_id is required")
	}
	return s.pain.Create(ctx, a)
}

func (s *Service) GetPainAssessment(ctx context.Context, id uuid.UUID) (*PainAssessment, error) {
	return s.pain.GetByID(ctx, id)
}

func (s *Service) ListPainAssessmentsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PainAssessment, int, error) {
	return s.pain.ListByPatient(ctx, patientID, limit, offset)
}

// -- Lines/Drains/Airways --

var validLinesDrainsStatuses = map[string]bool{
	"active": true, "removed": true, "replaced": true, "capped": true,
}

func (s *Service) CreateLinesDrains(ctx context.Context, l *LinesDrainsAirways) error {
	if l.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if l.EncounterID == uuid.Nil {
		return fmt.Errorf("encounter_id is required")
	}
	if l.Type == "" {
		return fmt.Errorf("type is required")
	}
	if l.Status == "" {
		l.Status = "active"
	}
	if !validLinesDrainsStatuses[l.Status] {
		return fmt.Errorf("invalid status: %s", l.Status)
	}
	return s.linesDrains.Create(ctx, l)
}

func (s *Service) GetLinesDrains(ctx context.Context, id uuid.UUID) (*LinesDrainsAirways, error) {
	return s.linesDrains.GetByID(ctx, id)
}

func (s *Service) UpdateLinesDrains(ctx context.Context, l *LinesDrainsAirways) error {
	if l.Status != "" && !validLinesDrainsStatuses[l.Status] {
		return fmt.Errorf("invalid status: %s", l.Status)
	}
	return s.linesDrains.Update(ctx, l)
}

func (s *Service) DeleteLinesDrains(ctx context.Context, id uuid.UUID) error {
	return s.linesDrains.Delete(ctx, id)
}

func (s *Service) ListLinesDrainsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*LinesDrainsAirways, int, error) {
	return s.linesDrains.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) ListLinesDrainsByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*LinesDrainsAirways, int, error) {
	return s.linesDrains.ListByEncounter(ctx, encounterID, limit, offset)
}

// -- Restraint --

func (s *Service) CreateRestraint(ctx context.Context, r *RestraintRecord) error {
	if r.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if r.RestraintType == "" {
		return fmt.Errorf("restraint_type is required")
	}
	if r.AppliedByID == uuid.Nil {
		return fmt.Errorf("applied_by_id is required")
	}
	return s.restraints.Create(ctx, r)
}

func (s *Service) GetRestraint(ctx context.Context, id uuid.UUID) (*RestraintRecord, error) {
	return s.restraints.GetByID(ctx, id)
}

func (s *Service) UpdateRestraint(ctx context.Context, r *RestraintRecord) error {
	return s.restraints.Update(ctx, r)
}

func (s *Service) ListRestraintsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*RestraintRecord, int, error) {
	return s.restraints.ListByPatient(ctx, patientID, limit, offset)
}

// -- Intake/Output --

func (s *Service) CreateIntakeOutput(ctx context.Context, r *IntakeOutputRecord) error {
	if r.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if r.EncounterID == uuid.Nil {
		return fmt.Errorf("encounter_id is required")
	}
	if r.Category == "" {
		return fmt.Errorf("category is required")
	}
	if r.RecordedByID == uuid.Nil {
		return fmt.Errorf("recorded_by_id is required")
	}
	return s.intakeOutput.Create(ctx, r)
}

func (s *Service) GetIntakeOutput(ctx context.Context, id uuid.UUID) (*IntakeOutputRecord, error) {
	return s.intakeOutput.GetByID(ctx, id)
}

func (s *Service) DeleteIntakeOutput(ctx context.Context, id uuid.UUID) error {
	return s.intakeOutput.Delete(ctx, id)
}

func (s *Service) ListIntakeOutputByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*IntakeOutputRecord, int, error) {
	return s.intakeOutput.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) ListIntakeOutputByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*IntakeOutputRecord, int, error) {
	return s.intakeOutput.ListByEncounter(ctx, encounterID, limit, offset)
}
