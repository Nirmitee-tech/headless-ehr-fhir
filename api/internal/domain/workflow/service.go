package workflow

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	activityDefs      ActivityDefinitionRepository
	requestGroups     RequestGroupRepository
	guidanceResponses GuidanceResponseRepository
	vt                *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(ad ActivityDefinitionRepository, rg RequestGroupRepository, gr GuidanceResponseRepository) *Service {
	return &Service{activityDefs: ad, requestGroups: rg, guidanceResponses: gr}
}

// -- ActivityDefinition --

var validActivityDefinitionStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

func (s *Service) CreateActivityDefinition(ctx context.Context, a *ActivityDefinition) error {
	if a.Status == "" {
		return fmt.Errorf("status is required")
	}
	if !validActivityDefinitionStatuses[a.Status] {
		return fmt.Errorf("invalid status: %s", a.Status)
	}
	if err := s.activityDefs.Create(ctx, a); err != nil {
		return err
	}
	a.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ActivityDefinition", a.FHIRID, a.ToFHIR())
	}
	return nil
}

func (s *Service) GetActivityDefinition(ctx context.Context, id uuid.UUID) (*ActivityDefinition, error) {
	return s.activityDefs.GetByID(ctx, id)
}

func (s *Service) GetActivityDefinitionByFHIRID(ctx context.Context, fhirID string) (*ActivityDefinition, error) {
	return s.activityDefs.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateActivityDefinition(ctx context.Context, a *ActivityDefinition) error {
	if a.Status != "" && !validActivityDefinitionStatuses[a.Status] {
		return fmt.Errorf("invalid status: %s", a.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ActivityDefinition", a.FHIRID, a.VersionID, a.ToFHIR())
		if err == nil {
			a.VersionID = newVer
		}
	}
	return s.activityDefs.Update(ctx, a)
}

func (s *Service) DeleteActivityDefinition(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		a, err := s.activityDefs.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ActivityDefinition", a.FHIRID, a.VersionID)
		}
	}
	return s.activityDefs.Delete(ctx, id)
}

func (s *Service) ListActivityDefinitions(ctx context.Context, limit, offset int) ([]*ActivityDefinition, int, error) {
	return s.activityDefs.List(ctx, limit, offset)
}

func (s *Service) SearchActivityDefinitions(ctx context.Context, params map[string]string, limit, offset int) ([]*ActivityDefinition, int, error) {
	return s.activityDefs.Search(ctx, params, limit, offset)
}

// -- RequestGroup --

var validRequestGroupStatuses = map[string]bool{
	"draft": true, "active": true, "on-hold": true, "revoked": true,
	"completed": true, "entered-in-error": true, "unknown": true,
}

var validRequestGroupIntents = map[string]bool{
	"proposal": true, "plan": true, "directive": true, "order": true,
	"original-order": true, "reflex-order": true, "filler-order": true,
	"instance-order": true, "option": true,
}

func (s *Service) CreateRequestGroup(ctx context.Context, rg *RequestGroup) error {
	if rg.Status == "" {
		rg.Status = "draft"
	}
	if !validRequestGroupStatuses[rg.Status] {
		return fmt.Errorf("invalid status: %s", rg.Status)
	}
	if rg.Intent == "" {
		return fmt.Errorf("intent is required")
	}
	if !validRequestGroupIntents[rg.Intent] {
		return fmt.Errorf("invalid intent: %s", rg.Intent)
	}
	if err := s.requestGroups.Create(ctx, rg); err != nil {
		return err
	}
	rg.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "RequestGroup", rg.FHIRID, rg.ToFHIR())
	}
	return nil
}

func (s *Service) GetRequestGroup(ctx context.Context, id uuid.UUID) (*RequestGroup, error) {
	return s.requestGroups.GetByID(ctx, id)
}

func (s *Service) GetRequestGroupByFHIRID(ctx context.Context, fhirID string) (*RequestGroup, error) {
	return s.requestGroups.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateRequestGroup(ctx context.Context, rg *RequestGroup) error {
	if rg.Status != "" && !validRequestGroupStatuses[rg.Status] {
		return fmt.Errorf("invalid status: %s", rg.Status)
	}
	if rg.Intent != "" && !validRequestGroupIntents[rg.Intent] {
		return fmt.Errorf("invalid intent: %s", rg.Intent)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "RequestGroup", rg.FHIRID, rg.VersionID, rg.ToFHIR())
		if err == nil {
			rg.VersionID = newVer
		}
	}
	return s.requestGroups.Update(ctx, rg)
}

func (s *Service) DeleteRequestGroup(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		rg, err := s.requestGroups.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "RequestGroup", rg.FHIRID, rg.VersionID)
		}
	}
	return s.requestGroups.Delete(ctx, id)
}

func (s *Service) ListRequestGroups(ctx context.Context, limit, offset int) ([]*RequestGroup, int, error) {
	return s.requestGroups.List(ctx, limit, offset)
}

func (s *Service) SearchRequestGroups(ctx context.Context, params map[string]string, limit, offset int) ([]*RequestGroup, int, error) {
	return s.requestGroups.Search(ctx, params, limit, offset)
}

func (s *Service) AddRequestGroupAction(ctx context.Context, a *RequestGroupAction) error {
	if a.RequestGroupID == uuid.Nil {
		return fmt.Errorf("request_group_id is required")
	}
	return s.requestGroups.AddAction(ctx, a)
}

func (s *Service) GetRequestGroupActions(ctx context.Context, requestGroupID uuid.UUID) ([]*RequestGroupAction, error) {
	return s.requestGroups.GetActions(ctx, requestGroupID)
}

// -- GuidanceResponse --

var validGuidanceResponseStatuses = map[string]bool{
	"success": true, "data-requested": true, "data-required": true,
	"in-progress": true, "failure": true, "entered-in-error": true,
}

func (s *Service) CreateGuidanceResponse(ctx context.Context, gr *GuidanceResponse) error {
	if gr.ModuleURI == "" {
		return fmt.Errorf("module_uri is required")
	}
	if gr.Status == "" {
		return fmt.Errorf("status is required")
	}
	if !validGuidanceResponseStatuses[gr.Status] {
		return fmt.Errorf("invalid status: %s", gr.Status)
	}
	if err := s.guidanceResponses.Create(ctx, gr); err != nil {
		return err
	}
	gr.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "GuidanceResponse", gr.FHIRID, gr.ToFHIR())
	}
	return nil
}

func (s *Service) GetGuidanceResponse(ctx context.Context, id uuid.UUID) (*GuidanceResponse, error) {
	return s.guidanceResponses.GetByID(ctx, id)
}

func (s *Service) GetGuidanceResponseByFHIRID(ctx context.Context, fhirID string) (*GuidanceResponse, error) {
	return s.guidanceResponses.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateGuidanceResponse(ctx context.Context, gr *GuidanceResponse) error {
	if gr.Status != "" && !validGuidanceResponseStatuses[gr.Status] {
		return fmt.Errorf("invalid status: %s", gr.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "GuidanceResponse", gr.FHIRID, gr.VersionID, gr.ToFHIR())
		if err == nil {
			gr.VersionID = newVer
		}
	}
	return s.guidanceResponses.Update(ctx, gr)
}

func (s *Service) DeleteGuidanceResponse(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		gr, err := s.guidanceResponses.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "GuidanceResponse", gr.FHIRID, gr.VersionID)
		}
	}
	return s.guidanceResponses.Delete(ctx, id)
}

func (s *Service) ListGuidanceResponses(ctx context.Context, limit, offset int) ([]*GuidanceResponse, int, error) {
	return s.guidanceResponses.List(ctx, limit, offset)
}

func (s *Service) SearchGuidanceResponses(ctx context.Context, params map[string]string, limit, offset int) ([]*GuidanceResponse, int, error) {
	return s.guidanceResponses.Search(ctx, params, limit, offset)
}
