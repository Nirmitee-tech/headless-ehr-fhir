package immunization

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	immunizations   ImmunizationRepository
	recommendations RecommendationRepository
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

func NewService(imm ImmunizationRepository, rec RecommendationRepository) *Service {
	return &Service{immunizations: imm, recommendations: rec}
}

// -- Immunization --

var validImmStatuses = map[string]bool{
	"completed": true, "entered-in-error": true, "not-done": true,
}

func (s *Service) CreateImmunization(ctx context.Context, im *Immunization) error {
	if im.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if im.VaccineCode == "" {
		return fmt.Errorf("vaccine_code is required")
	}
	if im.VaccineDisplay == "" {
		return fmt.Errorf("vaccine_display is required")
	}
	if im.Status == "" {
		im.Status = "completed"
	}
	if !validImmStatuses[im.Status] {
		return fmt.Errorf("invalid status: %s", im.Status)
	}
	if err := s.immunizations.Create(ctx, im); err != nil {
		return err
	}
	im.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Immunization", im.FHIRID, im.ToFHIR())
	}
	return nil
}

func (s *Service) GetImmunization(ctx context.Context, id uuid.UUID) (*Immunization, error) {
	return s.immunizations.GetByID(ctx, id)
}

func (s *Service) GetImmunizationByFHIRID(ctx context.Context, fhirID string) (*Immunization, error) {
	return s.immunizations.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateImmunization(ctx context.Context, im *Immunization) error {
	if im.Status != "" && !validImmStatuses[im.Status] {
		return fmt.Errorf("invalid status: %s", im.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Immunization", im.FHIRID, im.VersionID, im.ToFHIR())
		if err == nil {
			im.VersionID = newVer
		}
	}
	return s.immunizations.Update(ctx, im)
}

func (s *Service) DeleteImmunization(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		im, err := s.immunizations.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Immunization", im.FHIRID, im.VersionID)
		}
	}
	return s.immunizations.Delete(ctx, id)
}

func (s *Service) ListImmunizationsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Immunization, int, error) {
	return s.immunizations.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchImmunizations(ctx context.Context, params map[string]string, limit, offset int) ([]*Immunization, int, error) {
	return s.immunizations.Search(ctx, params, limit, offset)
}

// -- ImmunizationRecommendation --

func (s *Service) CreateRecommendation(ctx context.Context, r *ImmunizationRecommendation) error {
	if r.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if r.VaccineCode == "" {
		return fmt.Errorf("vaccine_code is required")
	}
	if r.ForecastStatus == "" {
		return fmt.Errorf("forecast_status is required")
	}
	if err := s.recommendations.Create(ctx, r); err != nil {
		return err
	}
	r.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ImmunizationRecommendation", r.FHIRID, r.ToFHIR())
	}
	return nil
}

func (s *Service) GetRecommendation(ctx context.Context, id uuid.UUID) (*ImmunizationRecommendation, error) {
	return s.recommendations.GetByID(ctx, id)
}

func (s *Service) GetRecommendationByFHIRID(ctx context.Context, fhirID string) (*ImmunizationRecommendation, error) {
	return s.recommendations.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateRecommendation(ctx context.Context, r *ImmunizationRecommendation) error {
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ImmunizationRecommendation", r.FHIRID, r.VersionID, r.ToFHIR())
		if err == nil {
			r.VersionID = newVer
		}
	}
	return s.recommendations.Update(ctx, r)
}

func (s *Service) DeleteRecommendation(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		r, err := s.recommendations.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ImmunizationRecommendation", r.FHIRID, r.VersionID)
		}
	}
	return s.recommendations.Delete(ctx, id)
}

func (s *Service) ListRecommendationsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ImmunizationRecommendation, int, error) {
	return s.recommendations.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchRecommendations(ctx context.Context, params map[string]string, limit, offset int) ([]*ImmunizationRecommendation, int, error) {
	return s.recommendations.Search(ctx, params, limit, offset)
}
