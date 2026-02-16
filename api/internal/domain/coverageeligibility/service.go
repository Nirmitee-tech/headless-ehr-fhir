package coverageeligibility

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	reqRepo  CoverageEligibilityRequestRepository
	respRepo CoverageEligibilityResponseRepository
	vt       *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(reqRepo CoverageEligibilityRequestRepository, respRepo CoverageEligibilityResponseRepository) *Service {
	return &Service{reqRepo: reqRepo, respRepo: respRepo}
}

var validRequestStatuses = map[string]bool{
	"active": true, "cancelled": true, "draft": true, "entered-in-error": true,
}

var validResponseStatuses = map[string]bool{
	"active": true, "cancelled": true, "draft": true, "entered-in-error": true,
}

var validOutcomes = map[string]bool{
	"queued": true, "complete": true, "error": true, "partial": true,
}

// -- CoverageEligibilityRequest methods --

func (s *Service) CreateRequest(ctx context.Context, r *CoverageEligibilityRequest) error {
	if r.PatientID == uuid.Nil {
		return fmt.Errorf("patient is required")
	}
	if r.Status == "" {
		r.Status = "active"
	}
	if !validRequestStatuses[r.Status] {
		return fmt.Errorf("invalid status: %s", r.Status)
	}
	if err := s.reqRepo.Create(ctx, r); err != nil {
		return err
	}
	r.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "CoverageEligibilityRequest", r.FHIRID, r.ToFHIR())
	}
	return nil
}

func (s *Service) GetRequest(ctx context.Context, id uuid.UUID) (*CoverageEligibilityRequest, error) {
	return s.reqRepo.GetByID(ctx, id)
}

func (s *Service) GetRequestByFHIRID(ctx context.Context, fhirID string) (*CoverageEligibilityRequest, error) {
	return s.reqRepo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateRequest(ctx context.Context, r *CoverageEligibilityRequest) error {
	if r.Status != "" && !validRequestStatuses[r.Status] {
		return fmt.Errorf("invalid status: %s", r.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "CoverageEligibilityRequest", r.FHIRID, r.VersionID, r.ToFHIR())
		if err == nil {
			r.VersionID = newVer
		}
	}
	return s.reqRepo.Update(ctx, r)
}

func (s *Service) DeleteRequest(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		r, err := s.reqRepo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "CoverageEligibilityRequest", r.FHIRID, r.VersionID)
		}
	}
	return s.reqRepo.Delete(ctx, id)
}

func (s *Service) SearchRequests(ctx context.Context, params map[string]string, limit, offset int) ([]*CoverageEligibilityRequest, int, error) {
	return s.reqRepo.Search(ctx, params, limit, offset)
}

// -- CoverageEligibilityResponse methods --

func (s *Service) CreateResponse(ctx context.Context, r *CoverageEligibilityResponse) error {
	if r.PatientID == uuid.Nil {
		return fmt.Errorf("patient is required")
	}
	if r.Status == "" {
		r.Status = "active"
	}
	if !validResponseStatuses[r.Status] {
		return fmt.Errorf("invalid status: %s", r.Status)
	}
	if r.Outcome != "" && !validOutcomes[r.Outcome] {
		return fmt.Errorf("invalid outcome: %s", r.Outcome)
	}
	if err := s.respRepo.Create(ctx, r); err != nil {
		return err
	}
	r.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "CoverageEligibilityResponse", r.FHIRID, r.ToFHIR())
	}
	return nil
}

func (s *Service) GetResponse(ctx context.Context, id uuid.UUID) (*CoverageEligibilityResponse, error) {
	return s.respRepo.GetByID(ctx, id)
}

func (s *Service) GetResponseByFHIRID(ctx context.Context, fhirID string) (*CoverageEligibilityResponse, error) {
	return s.respRepo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateResponse(ctx context.Context, r *CoverageEligibilityResponse) error {
	if r.Status != "" && !validResponseStatuses[r.Status] {
		return fmt.Errorf("invalid status: %s", r.Status)
	}
	if r.Outcome != "" && !validOutcomes[r.Outcome] {
		return fmt.Errorf("invalid outcome: %s", r.Outcome)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "CoverageEligibilityResponse", r.FHIRID, r.VersionID, r.ToFHIR())
		if err == nil {
			r.VersionID = newVer
		}
	}
	return s.respRepo.Update(ctx, r)
}

func (s *Service) DeleteResponse(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		r, err := s.respRepo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "CoverageEligibilityResponse", r.FHIRID, r.VersionID)
		}
	}
	return s.respRepo.Delete(ctx, id)
}

func (s *Service) SearchResponses(ctx context.Context, params map[string]string, limit, offset int) ([]*CoverageEligibilityResponse, int, error) {
	return s.respRepo.Search(ctx, params, limit, offset)
}
