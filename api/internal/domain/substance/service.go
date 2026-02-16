package substance

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
)

type Service struct {
	repo SubstanceRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo SubstanceRepository) *Service {
	return &Service{repo: repo}
}

var validSubstanceStatuses = map[string]bool{
	"active": true, "inactive": true, "entered-in-error": true,
}

func (s *Service) CreateSubstance(ctx context.Context, sub *Substance) error {
	if sub.CodeCode == "" {
		return fmt.Errorf("code is required")
	}
	if sub.Status == "" {
		sub.Status = "active"
	}
	if !validSubstanceStatuses[sub.Status] {
		return fmt.Errorf("invalid status: %s", sub.Status)
	}
	if err := s.repo.Create(ctx, sub); err != nil {
		return err
	}
	sub.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Substance", sub.FHIRID, sub.ToFHIR())
	}
	return nil
}

func (s *Service) GetSubstance(ctx context.Context, id string) (*Substance, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetSubstanceByFHIRID(ctx context.Context, fhirID string) (*Substance, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateSubstance(ctx context.Context, sub *Substance) error {
	if sub.Status != "" && !validSubstanceStatuses[sub.Status] {
		return fmt.Errorf("invalid status: %s", sub.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Substance", sub.FHIRID, sub.VersionID, sub.ToFHIR())
		if err == nil {
			sub.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, sub)
}

func (s *Service) DeleteSubstance(ctx context.Context, id string) error {
	if s.vt != nil {
		sub, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Substance", sub.FHIRID, sub.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchSubstances(ctx context.Context, params map[string]string, limit, offset int) ([]*Substance, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
