package person

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo PersonRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo PersonRepository) *Service {
	return &Service{repo: repo}
}

var validGenders = map[string]bool{
	"male": true, "female": true, "other": true, "unknown": true,
}

func (s *Service) CreatePerson(ctx context.Context, p *Person) error {
	if p.Gender != nil && !validGenders[*p.Gender] {
		return fmt.Errorf("invalid gender: %s", *p.Gender)
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return err
	}
	p.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Person", p.FHIRID, p.ToFHIR())
	}
	return nil
}

func (s *Service) GetPerson(ctx context.Context, id uuid.UUID) (*Person, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetPersonByFHIRID(ctx context.Context, fhirID string) (*Person, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdatePerson(ctx context.Context, p *Person) error {
	if p.Gender != nil && !validGenders[*p.Gender] {
		return fmt.Errorf("invalid gender: %s", *p.Gender)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Person", p.FHIRID, p.VersionID, p.ToFHIR())
		if err == nil {
			p.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, p)
}

func (s *Service) DeletePerson(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		p, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Person", p.FHIRID, p.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchPersons(ctx context.Context, params map[string]string, limit, offset int) ([]*Person, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
