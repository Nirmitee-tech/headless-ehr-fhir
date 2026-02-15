package relatedperson

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Service provides business logic for the RelatedPerson domain.
type Service struct {
	relatedPersons RelatedPersonRepository
	vt             *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

// NewService creates a new RelatedPerson domain service.
func NewService(rp RelatedPersonRepository) *Service {
	return &Service{relatedPersons: rp}
}

func (s *Service) CreateRelatedPerson(ctx context.Context, rp *RelatedPerson) error {
	if rp.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if rp.RelationshipCode == "" {
		return fmt.Errorf("relationship_code is required")
	}
	if rp.RelationshipDisplay == "" {
		return fmt.Errorf("relationship_display is required")
	}
	if err := s.relatedPersons.Create(ctx, rp); err != nil {
		return err
	}
	rp.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "RelatedPerson", rp.FHIRID, rp.ToFHIR())
	}
	return nil
}

func (s *Service) GetRelatedPerson(ctx context.Context, id uuid.UUID) (*RelatedPerson, error) {
	return s.relatedPersons.GetByID(ctx, id)
}

func (s *Service) GetRelatedPersonByFHIRID(ctx context.Context, fhirID string) (*RelatedPerson, error) {
	return s.relatedPersons.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateRelatedPerson(ctx context.Context, rp *RelatedPerson) error {
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "RelatedPerson", rp.FHIRID, rp.VersionID, rp.ToFHIR())
		if err == nil {
			rp.VersionID = newVer
		}
	}
	return s.relatedPersons.Update(ctx, rp)
}

func (s *Service) DeleteRelatedPerson(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		rp, err := s.relatedPersons.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "RelatedPerson", rp.FHIRID, rp.VersionID)
		}
	}
	return s.relatedPersons.Delete(ctx, id)
}

func (s *Service) ListRelatedPersonsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*RelatedPerson, int, error) {
	return s.relatedPersons.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchRelatedPersons(ctx context.Context, params map[string]string, limit, offset int) ([]*RelatedPerson, int, error) {
	return s.relatedPersons.Search(ctx, params, limit, offset)
}

func (s *Service) AddCommunication(ctx context.Context, c *RelatedPersonCommunication) error {
	if c.RelatedPersonID == uuid.Nil {
		return fmt.Errorf("related_person_id is required")
	}
	if c.LanguageCode == "" {
		return fmt.Errorf("language_code is required")
	}
	return s.relatedPersons.AddCommunication(ctx, c)
}

func (s *Service) GetCommunications(ctx context.Context, relatedPersonID uuid.UUID) ([]*RelatedPersonCommunication, error) {
	return s.relatedPersons.GetCommunications(ctx, relatedPersonID)
}
