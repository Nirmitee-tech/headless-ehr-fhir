package identity

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	patients      PatientRepository
	practitioners PractitionerRepository
}

func NewService(patients PatientRepository, practitioners PractitionerRepository) *Service {
	return &Service{patients: patients, practitioners: practitioners}
}

// -- Patient --

func (s *Service) CreatePatient(ctx context.Context, p *Patient) error {
	if p.FirstName == "" || p.LastName == "" {
		return fmt.Errorf("first_name and last_name are required")
	}
	if p.MRN == "" {
		return fmt.Errorf("mrn is required")
	}
	p.Active = true
	return s.patients.Create(ctx, p)
}

func (s *Service) GetPatient(ctx context.Context, id uuid.UUID) (*Patient, error) {
	return s.patients.GetByID(ctx, id)
}

func (s *Service) GetPatientByFHIRID(ctx context.Context, fhirID string) (*Patient, error) {
	return s.patients.GetByFHIRID(ctx, fhirID)
}

func (s *Service) GetPatientByMRN(ctx context.Context, mrn string) (*Patient, error) {
	return s.patients.GetByMRN(ctx, mrn)
}

func (s *Service) UpdatePatient(ctx context.Context, p *Patient) error {
	if p.FirstName == "" || p.LastName == "" {
		return fmt.Errorf("first_name and last_name are required")
	}
	return s.patients.Update(ctx, p)
}

func (s *Service) DeletePatient(ctx context.Context, id uuid.UUID) error {
	return s.patients.Delete(ctx, id)
}

func (s *Service) ListPatients(ctx context.Context, limit, offset int) ([]*Patient, int, error) {
	return s.patients.List(ctx, limit, offset)
}

func (s *Service) SearchPatients(ctx context.Context, params map[string]string, limit, offset int) ([]*Patient, int, error) {
	return s.patients.Search(ctx, params, limit, offset)
}

func (s *Service) AddPatientContact(ctx context.Context, c *PatientContact) error {
	if c.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if c.Relationship == "" {
		return fmt.Errorf("relationship is required")
	}
	return s.patients.AddContact(ctx, c)
}

func (s *Service) GetPatientContacts(ctx context.Context, patientID uuid.UUID) ([]*PatientContact, error) {
	return s.patients.GetContacts(ctx, patientID)
}

func (s *Service) RemovePatientContact(ctx context.Context, id uuid.UUID) error {
	return s.patients.RemoveContact(ctx, id)
}

func (s *Service) AddPatientIdentifier(ctx context.Context, ident *PatientIdentifier) error {
	if ident.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if ident.SystemURI == "" || ident.Value == "" {
		return fmt.Errorf("system_uri and value are required")
	}
	return s.patients.AddIdentifier(ctx, ident)
}

func (s *Service) GetPatientIdentifiers(ctx context.Context, patientID uuid.UUID) ([]*PatientIdentifier, error) {
	return s.patients.GetIdentifiers(ctx, patientID)
}

func (s *Service) RemovePatientIdentifier(ctx context.Context, id uuid.UUID) error {
	return s.patients.RemoveIdentifier(ctx, id)
}

// -- Practitioner --

func (s *Service) CreatePractitioner(ctx context.Context, p *Practitioner) error {
	if p.FirstName == "" || p.LastName == "" {
		return fmt.Errorf("first_name and last_name are required")
	}
	p.Active = true
	return s.practitioners.Create(ctx, p)
}

func (s *Service) GetPractitioner(ctx context.Context, id uuid.UUID) (*Practitioner, error) {
	return s.practitioners.GetByID(ctx, id)
}

func (s *Service) GetPractitionerByFHIRID(ctx context.Context, fhirID string) (*Practitioner, error) {
	return s.practitioners.GetByFHIRID(ctx, fhirID)
}

func (s *Service) GetPractitionerByNPI(ctx context.Context, npi string) (*Practitioner, error) {
	return s.practitioners.GetByNPI(ctx, npi)
}

func (s *Service) UpdatePractitioner(ctx context.Context, p *Practitioner) error {
	if p.FirstName == "" || p.LastName == "" {
		return fmt.Errorf("first_name and last_name are required")
	}
	return s.practitioners.Update(ctx, p)
}

func (s *Service) DeletePractitioner(ctx context.Context, id uuid.UUID) error {
	return s.practitioners.Delete(ctx, id)
}

func (s *Service) ListPractitioners(ctx context.Context, limit, offset int) ([]*Practitioner, int, error) {
	return s.practitioners.List(ctx, limit, offset)
}

func (s *Service) SearchPractitioners(ctx context.Context, params map[string]string, limit, offset int) ([]*Practitioner, int, error) {
	return s.practitioners.Search(ctx, params, limit, offset)
}

func (s *Service) AddPractitionerRole(ctx context.Context, role *PractitionerRole) error {
	if role.PractitionerID == uuid.Nil {
		return fmt.Errorf("practitioner_id is required")
	}
	if role.RoleCode == "" {
		return fmt.Errorf("role_code is required")
	}
	role.Active = true
	return s.practitioners.AddRole(ctx, role)
}

func (s *Service) GetPractitionerRoles(ctx context.Context, practitionerID uuid.UUID) ([]*PractitionerRole, error) {
	return s.practitioners.GetRoles(ctx, practitionerID)
}

func (s *Service) RemovePractitionerRole(ctx context.Context, id uuid.UUID) error {
	return s.practitioners.RemoveRole(ctx, id)
}
