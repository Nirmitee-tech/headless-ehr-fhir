package verificationresult

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo VerificationResultRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo VerificationResultRepository) *Service {
	return &Service{repo: repo}
}

var validVerificationResultStatuses = map[string]bool{
	"attested": true, "validated": true, "in-process": true,
	"req-revalid": true, "val-fail": true, "reval-fail": true,
}

func (s *Service) CreateVerificationResult(ctx context.Context, v *VerificationResult) error {
	if v.Status == "" {
		v.Status = "attested"
	}
	if !validVerificationResultStatuses[v.Status] {
		return fmt.Errorf("invalid status: %s", v.Status)
	}
	if err := s.repo.Create(ctx, v); err != nil {
		return err
	}
	v.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "VerificationResult", v.FHIRID, v.ToFHIR())
	}
	return nil
}

func (s *Service) GetVerificationResult(ctx context.Context, id uuid.UUID) (*VerificationResult, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetVerificationResultByFHIRID(ctx context.Context, fhirID string) (*VerificationResult, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateVerificationResult(ctx context.Context, v *VerificationResult) error {
	if v.Status != "" && !validVerificationResultStatuses[v.Status] {
		return fmt.Errorf("invalid status: %s", v.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "VerificationResult", v.FHIRID, v.VersionID, v.ToFHIR())
		if err == nil {
			v.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, v)
}

func (s *Service) DeleteVerificationResult(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		v, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "VerificationResult", v.FHIRID, v.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchVerificationResults(ctx context.Context, params map[string]string, limit, offset int) ([]*VerificationResult, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
