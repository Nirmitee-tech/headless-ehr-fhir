package evidencevariable

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo EvidenceVariableRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo EvidenceVariableRepository) *Service {
	return &Service{repo: repo}
}

var validStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

var validTypes = map[string]bool{
	"dichotomous": true, "continuous": true, "descriptive": true,
}

func (s *Service) CreateEvidenceVariable(ctx context.Context, e *EvidenceVariable) error {
	if e.Status == "" {
		e.Status = "draft"
	}
	if !validStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	if e.Type == "" {
		e.Type = "dichotomous"
	}
	if !validTypes[e.Type] {
		return fmt.Errorf("invalid type: %s", e.Type)
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return err
	}
	e.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "EvidenceVariable", e.FHIRID, e.ToFHIR())
	}
	return nil
}

func (s *Service) GetEvidenceVariable(ctx context.Context, id uuid.UUID) (*EvidenceVariable, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetEvidenceVariableByFHIRID(ctx context.Context, fhirID string) (*EvidenceVariable, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateEvidenceVariable(ctx context.Context, e *EvidenceVariable) error {
	if e.Status != "" && !validStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "EvidenceVariable", e.FHIRID, e.VersionID, e.ToFHIR())
		if err == nil {
			e.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, e)
}

func (s *Service) DeleteEvidenceVariable(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		e, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "EvidenceVariable", e.FHIRID, e.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchEvidenceVariables(ctx context.Context, params map[string]string, limit, offset int) ([]*EvidenceVariable, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
