package episodeofcare

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	episodes EpisodeOfCareRepository
	vt       *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(episodes EpisodeOfCareRepository) *Service {
	return &Service{episodes: episodes}
}

var validEpisodeOfCareStatuses = map[string]bool{
	"planned": true, "waitlist": true, "active": true, "onhold": true,
	"finished": true, "cancelled": true, "entered-in-error": true,
}

func (s *Service) CreateEpisodeOfCare(ctx context.Context, e *EpisodeOfCare) error {
	if e.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if e.Status == "" {
		e.Status = "planned"
	}
	if !validEpisodeOfCareStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	if err := s.episodes.Create(ctx, e); err != nil {
		return err
	}
	e.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "EpisodeOfCare", e.FHIRID, e.ToFHIR())
	}
	return nil
}

func (s *Service) GetEpisodeOfCare(ctx context.Context, id uuid.UUID) (*EpisodeOfCare, error) {
	return s.episodes.GetByID(ctx, id)
}

func (s *Service) GetEpisodeOfCareByFHIRID(ctx context.Context, fhirID string) (*EpisodeOfCare, error) {
	return s.episodes.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateEpisodeOfCare(ctx context.Context, e *EpisodeOfCare) error {
	if e.Status != "" && !validEpisodeOfCareStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "EpisodeOfCare", e.FHIRID, e.VersionID, e.ToFHIR())
		if err == nil {
			e.VersionID = newVer
		}
	}
	return s.episodes.Update(ctx, e)
}

func (s *Service) DeleteEpisodeOfCare(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		e, err := s.episodes.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "EpisodeOfCare", e.FHIRID, e.VersionID)
		}
	}
	return s.episodes.Delete(ctx, id)
}

func (s *Service) ListEpisodesByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*EpisodeOfCare, int, error) {
	return s.episodes.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchEpisodesOfCare(ctx context.Context, params map[string]string, limit, offset int) ([]*EpisodeOfCare, int, error) {
	return s.episodes.Search(ctx, params, limit, offset)
}
