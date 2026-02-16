package communicationrequest

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo CommunicationRequestRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo CommunicationRequestRepository) *Service {
	return &Service{repo: repo}
}

var validCommunicationRequestStatuses = map[string]bool{
	"draft": true, "active": true, "on-hold": true, "revoked": true,
	"completed": true, "entered-in-error": true, "unknown": true,
}

var validCommunicationRequestPriorities = map[string]bool{
	"routine": true, "urgent": true, "asap": true, "stat": true,
}

func (s *Service) CreateCommunicationRequest(ctx context.Context, cr *CommunicationRequest) error {
	if cr.Status == "" {
		cr.Status = "draft"
	}
	if !validCommunicationRequestStatuses[cr.Status] {
		return fmt.Errorf("invalid status: %s", cr.Status)
	}
	if cr.Priority != nil && !validCommunicationRequestPriorities[*cr.Priority] {
		return fmt.Errorf("invalid priority: %s", *cr.Priority)
	}
	if err := s.repo.Create(ctx, cr); err != nil {
		return err
	}
	cr.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "CommunicationRequest", cr.FHIRID, cr.ToFHIR())
	}
	return nil
}

func (s *Service) GetCommunicationRequest(ctx context.Context, id uuid.UUID) (*CommunicationRequest, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetCommunicationRequestByFHIRID(ctx context.Context, fhirID string) (*CommunicationRequest, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateCommunicationRequest(ctx context.Context, cr *CommunicationRequest) error {
	if cr.Status != "" && !validCommunicationRequestStatuses[cr.Status] {
		return fmt.Errorf("invalid status: %s", cr.Status)
	}
	if cr.Priority != nil && !validCommunicationRequestPriorities[*cr.Priority] {
		return fmt.Errorf("invalid priority: %s", *cr.Priority)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "CommunicationRequest", cr.FHIRID, cr.VersionID, cr.ToFHIR())
		if err == nil {
			cr.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, cr)
}

func (s *Service) DeleteCommunicationRequest(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		cr, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "CommunicationRequest", cr.FHIRID, cr.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchCommunicationRequests(ctx context.Context, params map[string]string, limit, offset int) ([]*CommunicationRequest, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
