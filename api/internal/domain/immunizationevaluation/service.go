package immunizationevaluation

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo ImmunizationEvaluationRepository
	vt   *fhir.VersionTracker
}

func NewService(repo ImmunizationEvaluationRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

var validIEStatuses = map[string]bool{
	"completed": true, "entered-in-error": true,
}

func (s *Service) CreateImmunizationEvaluation(ctx context.Context, ie *ImmunizationEvaluation) error {
	if ie.TargetDiseaseCode == "" {
		return fmt.Errorf("target disease code is required")
	}
	if ie.ImmunizationEventRef == "" {
		return fmt.Errorf("immunization event reference is required")
	}
	if ie.DoseStatusCode == "" {
		return fmt.Errorf("dose status code is required")
	}
	if ie.Status == "" {
		ie.Status = "completed"
	}
	if !validIEStatuses[ie.Status] {
		return fmt.Errorf("invalid status: %s", ie.Status)
	}
	if err := s.repo.Create(ctx, ie); err != nil {
		return err
	}
	ie.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ImmunizationEvaluation", ie.FHIRID, ie.ToFHIR())
	}
	return nil
}

func (s *Service) GetImmunizationEvaluation(ctx context.Context, id uuid.UUID) (*ImmunizationEvaluation, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetImmunizationEvaluationByFHIRID(ctx context.Context, fhirID string) (*ImmunizationEvaluation, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateImmunizationEvaluation(ctx context.Context, ie *ImmunizationEvaluation) error {
	if ie.Status != "" && !validIEStatuses[ie.Status] {
		return fmt.Errorf("invalid status: %s", ie.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ImmunizationEvaluation", ie.FHIRID, ie.VersionID, ie.ToFHIR())
		if err == nil {
			ie.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, ie)
}

func (s *Service) DeleteImmunizationEvaluation(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		ie, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ImmunizationEvaluation", ie.FHIRID, ie.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchImmunizationEvaluations(ctx context.Context, params map[string]string, limit, offset int) ([]*ImmunizationEvaluation, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
