package oncology

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	diagnoses  CancerDiagnosisRepository
	protocols  TreatmentProtocolRepository
	chemo      ChemoCycleRepository
	radiation  RadiationTherapyRepository
	markers    TumorMarkerRepository
	boards     TumorBoardRepository
}

func NewService(
	diagnoses CancerDiagnosisRepository,
	protocols TreatmentProtocolRepository,
	chemo ChemoCycleRepository,
	radiation RadiationTherapyRepository,
	markers TumorMarkerRepository,
	boards TumorBoardRepository,
) *Service {
	return &Service{
		diagnoses: diagnoses,
		protocols: protocols,
		chemo:     chemo,
		radiation: radiation,
		markers:   markers,
		boards:    boards,
	}
}

// -- Cancer Diagnosis --

var validCancerStatuses = map[string]bool{
	"active-treatment": true, "surveillance": true, "remission": true,
	"progression": true, "deceased": true, "lost-to-followup": true,
}

func (s *Service) CreateCancerDiagnosis(ctx context.Context, d *CancerDiagnosis) error {
	if d.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if d.DiagnosisDate.IsZero() {
		return fmt.Errorf("diagnosis_date is required")
	}
	if d.CurrentStatus == "" {
		d.CurrentStatus = "active-treatment"
	}
	if !validCancerStatuses[d.CurrentStatus] {
		return fmt.Errorf("invalid current_status: %s", d.CurrentStatus)
	}
	return s.diagnoses.Create(ctx, d)
}

func (s *Service) GetCancerDiagnosis(ctx context.Context, id uuid.UUID) (*CancerDiagnosis, error) {
	return s.diagnoses.GetByID(ctx, id)
}

func (s *Service) UpdateCancerDiagnosis(ctx context.Context, d *CancerDiagnosis) error {
	if d.CurrentStatus != "" && !validCancerStatuses[d.CurrentStatus] {
		return fmt.Errorf("invalid current_status: %s", d.CurrentStatus)
	}
	return s.diagnoses.Update(ctx, d)
}

func (s *Service) DeleteCancerDiagnosis(ctx context.Context, id uuid.UUID) error {
	return s.diagnoses.Delete(ctx, id)
}

func (s *Service) ListCancerDiagnoses(ctx context.Context, limit, offset int) ([]*CancerDiagnosis, int, error) {
	return s.diagnoses.List(ctx, limit, offset)
}

func (s *Service) ListCancerDiagnosesByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*CancerDiagnosis, int, error) {
	return s.diagnoses.ListByPatient(ctx, patientID, limit, offset)
}

// -- Treatment Protocol --

func (s *Service) CreateTreatmentProtocol(ctx context.Context, p *TreatmentProtocol) error {
	if p.CancerDiagnosisID == uuid.Nil {
		return fmt.Errorf("cancer_diagnosis_id is required")
	}
	if p.ProtocolName == "" {
		return fmt.Errorf("protocol_name is required")
	}
	if p.Status == "" {
		p.Status = "planned"
	}
	return s.protocols.Create(ctx, p)
}

func (s *Service) GetTreatmentProtocol(ctx context.Context, id uuid.UUID) (*TreatmentProtocol, error) {
	return s.protocols.GetByID(ctx, id)
}

func (s *Service) UpdateTreatmentProtocol(ctx context.Context, p *TreatmentProtocol) error {
	return s.protocols.Update(ctx, p)
}

func (s *Service) DeleteTreatmentProtocol(ctx context.Context, id uuid.UUID) error {
	return s.protocols.Delete(ctx, id)
}

func (s *Service) ListTreatmentProtocols(ctx context.Context, limit, offset int) ([]*TreatmentProtocol, int, error) {
	return s.protocols.List(ctx, limit, offset)
}

func (s *Service) AddProtocolDrug(ctx context.Context, d *TreatmentProtocolDrug) error {
	if d.ProtocolID == uuid.Nil {
		return fmt.Errorf("protocol_id is required")
	}
	if d.DrugName == "" {
		return fmt.Errorf("drug_name is required")
	}
	return s.protocols.AddDrug(ctx, d)
}

func (s *Service) GetProtocolDrugs(ctx context.Context, protocolID uuid.UUID) ([]*TreatmentProtocolDrug, error) {
	return s.protocols.GetDrugs(ctx, protocolID)
}

// -- Chemo Cycle --

var validChemoCycleStatuses = map[string]bool{
	"planned": true, "active": true, "completed": true,
	"held": true, "cancelled": true, "modified": true,
}

func (s *Service) CreateChemoCycle(ctx context.Context, c *ChemoCycle) error {
	if c.ProtocolID == uuid.Nil {
		return fmt.Errorf("protocol_id is required")
	}
	if c.CycleNumber <= 0 {
		return fmt.Errorf("cycle_number must be positive")
	}
	if c.Status == "" {
		c.Status = "planned"
	}
	if !validChemoCycleStatuses[c.Status] {
		return fmt.Errorf("invalid chemo cycle status: %s", c.Status)
	}
	return s.chemo.Create(ctx, c)
}

func (s *Service) GetChemoCycle(ctx context.Context, id uuid.UUID) (*ChemoCycle, error) {
	return s.chemo.GetByID(ctx, id)
}

func (s *Service) UpdateChemoCycle(ctx context.Context, c *ChemoCycle) error {
	if c.Status != "" && !validChemoCycleStatuses[c.Status] {
		return fmt.Errorf("invalid chemo cycle status: %s", c.Status)
	}
	return s.chemo.Update(ctx, c)
}

func (s *Service) DeleteChemoCycle(ctx context.Context, id uuid.UUID) error {
	return s.chemo.Delete(ctx, id)
}

func (s *Service) ListChemoCycles(ctx context.Context, limit, offset int) ([]*ChemoCycle, int, error) {
	return s.chemo.List(ctx, limit, offset)
}

func (s *Service) AddChemoAdministration(ctx context.Context, a *ChemoAdministration) error {
	if a.CycleID == uuid.Nil {
		return fmt.Errorf("cycle_id is required")
	}
	if a.DrugName == "" {
		return fmt.Errorf("drug_name is required")
	}
	if a.AdministrationDatetime.IsZero() {
		a.AdministrationDatetime = time.Now()
	}
	return s.chemo.AddAdministration(ctx, a)
}

func (s *Service) GetChemoAdministrations(ctx context.Context, cycleID uuid.UUID) ([]*ChemoAdministration, error) {
	return s.chemo.GetAdministrations(ctx, cycleID)
}

// -- Radiation Therapy --

var validRadiationStatuses = map[string]bool{
	"planned": true, "in-progress": true, "completed": true, "cancelled": true,
}

func (s *Service) CreateRadiationTherapy(ctx context.Context, r *RadiationTherapy) error {
	if r.CancerDiagnosisID == uuid.Nil {
		return fmt.Errorf("cancer_diagnosis_id is required")
	}
	if r.Status == "" {
		r.Status = "planned"
	}
	if !validRadiationStatuses[r.Status] {
		return fmt.Errorf("invalid radiation status: %s", r.Status)
	}
	return s.radiation.Create(ctx, r)
}

func (s *Service) GetRadiationTherapy(ctx context.Context, id uuid.UUID) (*RadiationTherapy, error) {
	return s.radiation.GetByID(ctx, id)
}

func (s *Service) UpdateRadiationTherapy(ctx context.Context, r *RadiationTherapy) error {
	if r.Status != "" && !validRadiationStatuses[r.Status] {
		return fmt.Errorf("invalid radiation status: %s", r.Status)
	}
	return s.radiation.Update(ctx, r)
}

func (s *Service) DeleteRadiationTherapy(ctx context.Context, id uuid.UUID) error {
	return s.radiation.Delete(ctx, id)
}

func (s *Service) ListRadiationTherapies(ctx context.Context, limit, offset int) ([]*RadiationTherapy, int, error) {
	return s.radiation.List(ctx, limit, offset)
}

func (s *Service) AddRadiationSession(ctx context.Context, sess *RadiationSession) error {
	if sess.RadiationTherapyID == uuid.Nil {
		return fmt.Errorf("radiation_therapy_id is required")
	}
	if sess.SessionNumber <= 0 {
		return fmt.Errorf("session_number must be positive")
	}
	if sess.SessionDate.IsZero() {
		sess.SessionDate = time.Now()
	}
	return s.radiation.AddSession(ctx, sess)
}

func (s *Service) GetRadiationSessions(ctx context.Context, radiationID uuid.UUID) ([]*RadiationSession, error) {
	return s.radiation.GetSessions(ctx, radiationID)
}

// -- Tumor Marker --

func (s *Service) CreateTumorMarker(ctx context.Context, m *TumorMarker) error {
	if m.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if m.MarkerName == "" {
		return fmt.Errorf("marker_name is required")
	}
	return s.markers.Create(ctx, m)
}

func (s *Service) GetTumorMarker(ctx context.Context, id uuid.UUID) (*TumorMarker, error) {
	return s.markers.GetByID(ctx, id)
}

func (s *Service) UpdateTumorMarker(ctx context.Context, m *TumorMarker) error {
	return s.markers.Update(ctx, m)
}

func (s *Service) DeleteTumorMarker(ctx context.Context, id uuid.UUID) error {
	return s.markers.Delete(ctx, id)
}

func (s *Service) ListTumorMarkers(ctx context.Context, limit, offset int) ([]*TumorMarker, int, error) {
	return s.markers.List(ctx, limit, offset)
}

// -- Tumor Board Review --

func (s *Service) CreateTumorBoardReview(ctx context.Context, r *TumorBoardReview) error {
	if r.CancerDiagnosisID == uuid.Nil {
		return fmt.Errorf("cancer_diagnosis_id is required")
	}
	if r.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if r.ReviewDate.IsZero() {
		r.ReviewDate = time.Now()
	}
	return s.boards.Create(ctx, r)
}

func (s *Service) GetTumorBoardReview(ctx context.Context, id uuid.UUID) (*TumorBoardReview, error) {
	return s.boards.GetByID(ctx, id)
}

func (s *Service) UpdateTumorBoardReview(ctx context.Context, r *TumorBoardReview) error {
	return s.boards.Update(ctx, r)
}

func (s *Service) DeleteTumorBoardReview(ctx context.Context, id uuid.UUID) error {
	return s.boards.Delete(ctx, id)
}

func (s *Service) ListTumorBoardReviews(ctx context.Context, limit, offset int) ([]*TumorBoardReview, int, error) {
	return s.boards.List(ctx, limit, offset)
}
