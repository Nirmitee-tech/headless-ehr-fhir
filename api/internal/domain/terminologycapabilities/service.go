package terminologycapabilities

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo TerminologyCapabilitiesRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo TerminologyCapabilitiesRepository) *Service {
	return &Service{repo: repo}
}

var validStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

func (s *Service) CreateTerminologyCapabilities(ctx context.Context, tc *TerminologyCapabilities) error {
	if tc.Status == "" {
		tc.Status = "draft"
	}
	if !validStatuses[tc.Status] {
		return fmt.Errorf("invalid status: %s", tc.Status)
	}
	if tc.Kind == "" {
		tc.Kind = "instance"
	}
	if tc.CodeSearch != nil && *tc.CodeSearch != "explicit" && *tc.CodeSearch != "all" {
		return fmt.Errorf("invalid codeSearch: %s", *tc.CodeSearch)
	}
	if err := s.repo.Create(ctx, tc); err != nil {
		return err
	}
	tc.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "TerminologyCapabilities", tc.FHIRID, tc.ToFHIR())
	}
	return nil
}

func (s *Service) GetTerminologyCapabilities(ctx context.Context, id uuid.UUID) (*TerminologyCapabilities, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetTerminologyCapabilitiesByFHIRID(ctx context.Context, fhirID string) (*TerminologyCapabilities, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateTerminologyCapabilities(ctx context.Context, tc *TerminologyCapabilities) error {
	if tc.Status != "" && !validStatuses[tc.Status] {
		return fmt.Errorf("invalid status: %s", tc.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "TerminologyCapabilities", tc.FHIRID, tc.VersionID, tc.ToFHIR())
		if err == nil {
			tc.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, tc)
}

func (s *Service) DeleteTerminologyCapabilities(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		tc, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "TerminologyCapabilities", tc.FHIRID, tc.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchTerminologyCapabilities(ctx context.Context, params map[string]string, limit, offset int) ([]*TerminologyCapabilities, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
