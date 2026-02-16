package riskevidencesynthesis

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo RiskEvidenceSynthesisRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo RiskEvidenceSynthesisRepository) *Service {
	return &Service{repo: repo}
}

var validStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

func (s *Service) CreateRiskEvidenceSynthesis(ctx context.Context, e *RiskEvidenceSynthesis) error {
	if e.Status == "" {
		e.Status = "draft"
	}
	if !validStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return err
	}
	e.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "RiskEvidenceSynthesis", e.FHIRID, e.ToFHIR())
	}
	return nil
}

func (s *Service) GetRiskEvidenceSynthesis(ctx context.Context, id uuid.UUID) (*RiskEvidenceSynthesis, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetRiskEvidenceSynthesisByFHIRID(ctx context.Context, fhirID string) (*RiskEvidenceSynthesis, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateRiskEvidenceSynthesis(ctx context.Context, e *RiskEvidenceSynthesis) error {
	if e.Status != "" && !validStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "RiskEvidenceSynthesis", e.FHIRID, e.VersionID, e.ToFHIR())
		if err == nil {
			e.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, e)
}

func (s *Service) DeleteRiskEvidenceSynthesis(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		e, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "RiskEvidenceSynthesis", e.FHIRID, e.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchRiskEvidenceSyntheses(ctx context.Context, params map[string]string, limit, offset int) ([]*RiskEvidenceSynthesis, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
