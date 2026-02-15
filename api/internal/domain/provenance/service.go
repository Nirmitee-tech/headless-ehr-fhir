package provenance

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Service provides business logic for the Provenance domain.
type Service struct {
	provenances ProvenanceRepository
	vt          *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

// NewService creates a new Provenance domain service.
func NewService(p ProvenanceRepository) *Service {
	return &Service{provenances: p}
}

func (s *Service) CreateProvenance(ctx context.Context, p *Provenance) error {
	if p.TargetType == "" {
		return fmt.Errorf("target_type is required")
	}
	if p.TargetID == "" {
		return fmt.Errorf("target_id is required")
	}
	if err := s.provenances.Create(ctx, p); err != nil {
		return err
	}
	p.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Provenance", p.FHIRID, p.ToFHIR())
	}
	return nil
}

func (s *Service) GetProvenance(ctx context.Context, id uuid.UUID) (*Provenance, error) {
	return s.provenances.GetByID(ctx, id)
}

func (s *Service) GetProvenanceByFHIRID(ctx context.Context, fhirID string) (*Provenance, error) {
	return s.provenances.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateProvenance(ctx context.Context, p *Provenance) error {
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Provenance", p.FHIRID, p.VersionID, p.ToFHIR())
		if err == nil {
			p.VersionID = newVer
		}
	}
	return s.provenances.Update(ctx, p)
}

func (s *Service) DeleteProvenance(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		p, err := s.provenances.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Provenance", p.FHIRID, p.VersionID)
		}
	}
	return s.provenances.Delete(ctx, id)
}

func (s *Service) SearchProvenances(ctx context.Context, params map[string]string, limit, offset int) ([]*Provenance, int, error) {
	return s.provenances.Search(ctx, params, limit, offset)
}

func (s *Service) AddAgent(ctx context.Context, a *ProvenanceAgent) error {
	if a.ProvenanceID == uuid.Nil {
		return fmt.Errorf("provenance_id is required")
	}
	if a.WhoType == "" || a.WhoID == "" {
		return fmt.Errorf("who_type and who_id are required")
	}
	return s.provenances.AddAgent(ctx, a)
}

func (s *Service) GetAgents(ctx context.Context, provenanceID uuid.UUID) ([]*ProvenanceAgent, error) {
	return s.provenances.GetAgents(ctx, provenanceID)
}

func (s *Service) AddEntity(ctx context.Context, e *ProvenanceEntity) error {
	if e.ProvenanceID == uuid.Nil {
		return fmt.Errorf("provenance_id is required")
	}
	if e.Role == "" {
		return fmt.Errorf("role is required")
	}
	return s.provenances.AddEntity(ctx, e)
}

func (s *Service) GetEntities(ctx context.Context, provenanceID uuid.UUID) ([]*ProvenanceEntity, error) {
	return s.provenances.GetEntities(ctx, provenanceID)
}
