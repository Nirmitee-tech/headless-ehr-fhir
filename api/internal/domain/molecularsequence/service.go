package molecularsequence

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo MolecularSequenceRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo MolecularSequenceRepository) *Service {
	return &Service{repo: repo}
}

var validMolecularSequenceTypes = map[string]bool{
	"aa": true, "dna": true, "rna": true,
}

func (s *Service) CreateMolecularSequence(ctx context.Context, m *MolecularSequence) error {
	if m.Type == "" {
		return fmt.Errorf("type is required")
	}
	if !validMolecularSequenceTypes[m.Type] {
		return fmt.Errorf("invalid type: %s", m.Type)
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return err
	}
	m.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "MolecularSequence", m.FHIRID, m.ToFHIR())
	}
	return nil
}

func (s *Service) GetMolecularSequence(ctx context.Context, id uuid.UUID) (*MolecularSequence, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetMolecularSequenceByFHIRID(ctx context.Context, fhirID string) (*MolecularSequence, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateMolecularSequence(ctx context.Context, m *MolecularSequence) error {
	if m.Type != "" && !validMolecularSequenceTypes[m.Type] {
		return fmt.Errorf("invalid type: %s", m.Type)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "MolecularSequence", m.FHIRID, m.VersionID, m.ToFHIR())
		if err == nil {
			m.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, m)
}

func (s *Service) DeleteMolecularSequence(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		m, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "MolecularSequence", m.FHIRID, m.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchMolecularSequences(ctx context.Context, params map[string]string, limit, offset int) ([]*MolecularSequence, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
