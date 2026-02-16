package researchsubject

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo ResearchSubjectRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo ResearchSubjectRepository) *Service {
	return &Service{repo: repo}
}

var validResearchSubjectStatuses = map[string]bool{
	"candidate": true, "eligible": true, "follow-up": true, "ineligible": true,
	"not-registered": true, "off-study": true, "on-study": true,
	"on-study-intervention": true, "on-study-observation": true,
	"pending-on-study": true, "potential-candidate": true, "screening": true, "withdrawn": true,
}

func (s *Service) CreateResearchSubject(ctx context.Context, r *ResearchSubject) error {
	if r.Status == "" {
		r.Status = "candidate"
	}
	if !validResearchSubjectStatuses[r.Status] {
		return fmt.Errorf("invalid status: %s", r.Status)
	}
	if err := s.repo.Create(ctx, r); err != nil {
		return err
	}
	r.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ResearchSubject", r.FHIRID, r.ToFHIR())
	}
	return nil
}

func (s *Service) GetResearchSubject(ctx context.Context, id uuid.UUID) (*ResearchSubject, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetResearchSubjectByFHIRID(ctx context.Context, fhirID string) (*ResearchSubject, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateResearchSubject(ctx context.Context, r *ResearchSubject) error {
	if r.Status != "" && !validResearchSubjectStatuses[r.Status] {
		return fmt.Errorf("invalid status: %s", r.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ResearchSubject", r.FHIRID, r.VersionID, r.ToFHIR())
		if err == nil {
			r.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, r)
}

func (s *Service) DeleteResearchSubject(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		r, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ResearchSubject", r.FHIRID, r.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchResearchSubjects(ctx context.Context, params map[string]string, limit, offset int) ([]*ResearchSubject, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
