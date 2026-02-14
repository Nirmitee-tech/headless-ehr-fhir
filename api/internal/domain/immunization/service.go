package immunization

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	immunizations   ImmunizationRepository
	recommendations RecommendationRepository
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
	return s.immunizations.Create(ctx, im)
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
	return s.immunizations.Update(ctx, im)
}

func (s *Service) DeleteImmunization(ctx context.Context, id uuid.UUID) error {
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
	return s.recommendations.Create(ctx, r)
}

func (s *Service) GetRecommendation(ctx context.Context, id uuid.UUID) (*ImmunizationRecommendation, error) {
	return s.recommendations.GetByID(ctx, id)
}

func (s *Service) GetRecommendationByFHIRID(ctx context.Context, fhirID string) (*ImmunizationRecommendation, error) {
	return s.recommendations.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateRecommendation(ctx context.Context, r *ImmunizationRecommendation) error {
	return s.recommendations.Update(ctx, r)
}

func (s *Service) DeleteRecommendation(ctx context.Context, id uuid.UUID) error {
	return s.recommendations.Delete(ctx, id)
}

func (s *Service) ListRecommendationsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ImmunizationRecommendation, int, error) {
	return s.recommendations.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchRecommendations(ctx context.Context, params map[string]string, limit, offset int) ([]*ImmunizationRecommendation, int, error) {
	return s.recommendations.Search(ctx, params, limit, offset)
}
