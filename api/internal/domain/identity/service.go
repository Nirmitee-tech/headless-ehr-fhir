package identity

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	patients      PatientRepository
	practitioners PractitionerRepository
	links         PatientLinkRepository
	vt            *fhir.VersionTracker
}

func NewService(patients PatientRepository, practitioners PractitionerRepository, links PatientLinkRepository) *Service {
	return &Service{patients: patients, practitioners: practitioners, links: links}
}

// SetVersionTracker attaches an optional VersionTracker to the service.
// When set, create/update/delete operations record version history.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
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
	if err := s.patients.Create(ctx, p); err != nil {
		return err
	}
	p.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Patient", p.FHIRID, p.ToFHIR())
	}
	return nil
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
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Patient", p.FHIRID, p.VersionID, p.ToFHIR())
		if err == nil {
			p.VersionID = newVer
		}
	}
	return s.patients.Update(ctx, p)
}

func (s *Service) DeletePatient(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		p, err := s.patients.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Patient", p.FHIRID, p.VersionID)
		}
	}
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

// -- Patient Matching / MPI --

// MatchPatient performs deterministic matching on name, DOB, and gender.
func (s *Service) MatchPatient(ctx context.Context, patientID uuid.UUID) ([]*PatientMatchResult, error) {
	source, err := s.patients.GetByID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
	}

	// Search for candidates by name
	params := map[string]string{
		"family": source.LastName,
	}
	candidates, _, err := s.patients.Search(ctx, params, 100, 0)
	if err != nil {
		return nil, err
	}

	var results []*PatientMatchResult
	for _, candidate := range candidates {
		if candidate.ID == source.ID {
			continue
		}
		score := 0.0
		var matchFields []string

		// Name matching
		if strings.EqualFold(candidate.LastName, source.LastName) {
			score += 0.3
			matchFields = append(matchFields, "last_name")
		}
		if strings.EqualFold(candidate.FirstName, source.FirstName) {
			score += 0.3
			matchFields = append(matchFields, "first_name")
		}

		// DOB matching
		if source.BirthDate != nil && candidate.BirthDate != nil {
			if source.BirthDate.Format("2006-01-02") == candidate.BirthDate.Format("2006-01-02") {
				score += 0.25
				matchFields = append(matchFields, "birth_date")
			}
		}

		// Gender matching
		if source.Gender != nil && candidate.Gender != nil {
			if *source.Gender == *candidate.Gender {
				score += 0.15
				matchFields = append(matchFields, "gender")
			}
		}

		if score >= 0.5 {
			results = append(results, &PatientMatchResult{
				Patient:     candidate,
				Score:       score,
				MatchFields: matchFields,
			})
		}
	}
	return results, nil
}

// LinkPatients creates a link between two patients.
func (s *Service) LinkPatients(ctx context.Context, link *PatientLink) error {
	if link.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if link.LinkedPatientID == uuid.Nil {
		return fmt.Errorf("linked_patient_id is required")
	}
	if link.PatientID == link.LinkedPatientID {
		return fmt.Errorf("cannot link a patient to themselves")
	}
	validLinkTypes := map[string]bool{
		"replaced-by": true, "replaces": true, "refer": true, "seealso": true,
	}
	if !validLinkTypes[link.LinkType] {
		return fmt.Errorf("invalid link_type: %s", link.LinkType)
	}
	if link.CreatedAt.IsZero() {
		link.CreatedAt = time.Now()
	}
	return s.links.Create(ctx, link)
}

// GetPatientLinks returns all links for a patient.
func (s *Service) GetPatientLinks(ctx context.Context, patientID uuid.UUID) ([]*PatientLink, error) {
	return s.links.GetByPatientID(ctx, patientID)
}

// UnlinkPatients removes a patient link by ID.
func (s *Service) UnlinkPatients(ctx context.Context, linkID uuid.UUID) error {
	return s.links.Delete(ctx, linkID)
}

// -- Practitioner --

func (s *Service) CreatePractitioner(ctx context.Context, p *Practitioner) error {
	if p.FirstName == "" || p.LastName == "" {
		return fmt.Errorf("first_name and last_name are required")
	}
	p.Active = true
	if err := s.practitioners.Create(ctx, p); err != nil {
		return err
	}
	p.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Practitioner", p.FHIRID, p.ToFHIR())
	}
	return nil
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
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Practitioner", p.FHIRID, p.VersionID, p.ToFHIR())
		if err == nil {
			p.VersionID = newVer
		}
	}
	return s.practitioners.Update(ctx, p)
}

func (s *Service) DeletePractitioner(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		p, err := s.practitioners.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Practitioner", p.FHIRID, p.VersionID)
		}
	}
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
