package medicationknowledge

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo MedicationKnowledgeRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo MedicationKnowledgeRepository) *Service {
	return &Service{repo: repo}
}

var validMedicationKnowledgeStatuses = map[string]bool{
	"active": true, "inactive": true, "entered-in-error": true,
}

func (s *Service) CreateMedicationKnowledge(ctx context.Context, m *MedicationKnowledge) error {
	if m.Status == "" {
		m.Status = "active"
	}
	if !validMedicationKnowledgeStatuses[m.Status] {
		return fmt.Errorf("invalid status: %s", m.Status)
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return err
	}
	m.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "MedicationKnowledge", m.FHIRID, m.ToFHIR())
	}
	return nil
}

func (s *Service) GetMedicationKnowledge(ctx context.Context, id uuid.UUID) (*MedicationKnowledge, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetMedicationKnowledgeByFHIRID(ctx context.Context, fhirID string) (*MedicationKnowledge, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateMedicationKnowledge(ctx context.Context, m *MedicationKnowledge) error {
	if m.Status != "" && !validMedicationKnowledgeStatuses[m.Status] {
		return fmt.Errorf("invalid status: %s", m.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "MedicationKnowledge", m.FHIRID, m.VersionID, m.ToFHIR())
		if err == nil {
			m.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, m)
}

func (s *Service) DeleteMedicationKnowledge(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		m, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "MedicationKnowledge", m.FHIRID, m.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchMedicationKnowledge(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicationKnowledge, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
