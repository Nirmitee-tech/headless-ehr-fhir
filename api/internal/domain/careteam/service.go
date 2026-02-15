package careteam

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	careTeams CareTeamRepository
	vt        *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(ct CareTeamRepository) *Service {
	return &Service{careTeams: ct}
}

var validCTStatuses = map[string]bool{
	"proposed": true, "active": true, "suspended": true,
	"inactive": true, "entered-in-error": true,
}

func (s *Service) CreateCareTeam(ctx context.Context, ct *CareTeam) error {
	if ct.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if ct.Status == "" {
		return fmt.Errorf("status is required")
	}
	if !validCTStatuses[ct.Status] {
		return fmt.Errorf("invalid status: %s", ct.Status)
	}
	if err := s.careTeams.Create(ctx, ct); err != nil {
		return err
	}
	ct.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "CareTeam", ct.FHIRID, ct.ToFHIR())
	}
	return nil
}

func (s *Service) GetCareTeam(ctx context.Context, id uuid.UUID) (*CareTeam, error) {
	return s.careTeams.GetByID(ctx, id)
}

func (s *Service) GetCareTeamByFHIRID(ctx context.Context, fhirID string) (*CareTeam, error) {
	return s.careTeams.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateCareTeam(ctx context.Context, ct *CareTeam) error {
	if ct.Status != "" && !validCTStatuses[ct.Status] {
		return fmt.Errorf("invalid status: %s", ct.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "CareTeam", ct.FHIRID, ct.VersionID, ct.ToFHIR())
		if err == nil {
			ct.VersionID = newVer
		}
	}
	return s.careTeams.Update(ctx, ct)
}

func (s *Service) DeleteCareTeam(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		ct, err := s.careTeams.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "CareTeam", ct.FHIRID, ct.VersionID)
		}
	}
	return s.careTeams.Delete(ctx, id)
}

func (s *Service) ListCareTeamsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*CareTeam, int, error) {
	return s.careTeams.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchCareTeams(ctx context.Context, params map[string]string, limit, offset int) ([]*CareTeam, int, error) {
	return s.careTeams.Search(ctx, params, limit, offset)
}

func (s *Service) AddParticipant(ctx context.Context, careTeamID uuid.UUID, p *CareTeamParticipant) error {
	if p.MemberID == uuid.Nil {
		return fmt.Errorf("member_id is required")
	}
	return s.careTeams.AddParticipant(ctx, careTeamID, p)
}

func (s *Service) RemoveParticipant(ctx context.Context, careTeamID uuid.UUID, participantID uuid.UUID) error {
	return s.careTeams.RemoveParticipant(ctx, careTeamID, participantID)
}

func (s *Service) GetParticipants(ctx context.Context, careTeamID uuid.UUID) ([]*CareTeamParticipant, error) {
	return s.careTeams.GetParticipants(ctx, careTeamID)
}
