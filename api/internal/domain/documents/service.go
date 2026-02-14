package documents

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	consents   ConsentRepository
	docRefs    DocumentReferenceRepository
	notes      ClinicalNoteRepository
	comps      CompositionRepository
}

func NewService(consents ConsentRepository, docRefs DocumentReferenceRepository, notes ClinicalNoteRepository, comps CompositionRepository) *Service {
	return &Service{consents: consents, docRefs: docRefs, notes: notes, comps: comps}
}

// -- Consent --

var validConsentStatuses = map[string]bool{
	"draft": true, "proposed": true, "active": true,
	"rejected": true, "inactive": true, "entered-in-error": true,
}

func (s *Service) CreateConsent(ctx context.Context, c *Consent) error {
	if c.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if c.Status == "" {
		c.Status = "draft"
	}
	if !validConsentStatuses[c.Status] {
		return fmt.Errorf("invalid status: %s", c.Status)
	}
	return s.consents.Create(ctx, c)
}

func (s *Service) GetConsent(ctx context.Context, id uuid.UUID) (*Consent, error) {
	return s.consents.GetByID(ctx, id)
}

func (s *Service) GetConsentByFHIRID(ctx context.Context, fhirID string) (*Consent, error) {
	return s.consents.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateConsent(ctx context.Context, c *Consent) error {
	if c.Status != "" && !validConsentStatuses[c.Status] {
		return fmt.Errorf("invalid status: %s", c.Status)
	}
	return s.consents.Update(ctx, c)
}

func (s *Service) DeleteConsent(ctx context.Context, id uuid.UUID) error {
	return s.consents.Delete(ctx, id)
}

func (s *Service) ListConsentsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Consent, int, error) {
	return s.consents.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchConsents(ctx context.Context, params map[string]string, limit, offset int) ([]*Consent, int, error) {
	return s.consents.Search(ctx, params, limit, offset)
}

// -- DocumentReference --

var validDocRefStatuses = map[string]bool{
	"current": true, "superseded": true, "entered-in-error": true,
}

func (s *Service) CreateDocumentReference(ctx context.Context, d *DocumentReference) error {
	if d.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if d.Status == "" {
		d.Status = "current"
	}
	if !validDocRefStatuses[d.Status] {
		return fmt.Errorf("invalid status: %s", d.Status)
	}
	return s.docRefs.Create(ctx, d)
}

func (s *Service) GetDocumentReference(ctx context.Context, id uuid.UUID) (*DocumentReference, error) {
	return s.docRefs.GetByID(ctx, id)
}

func (s *Service) GetDocumentReferenceByFHIRID(ctx context.Context, fhirID string) (*DocumentReference, error) {
	return s.docRefs.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateDocumentReference(ctx context.Context, d *DocumentReference) error {
	if d.Status != "" && !validDocRefStatuses[d.Status] {
		return fmt.Errorf("invalid status: %s", d.Status)
	}
	return s.docRefs.Update(ctx, d)
}

func (s *Service) DeleteDocumentReference(ctx context.Context, id uuid.UUID) error {
	return s.docRefs.Delete(ctx, id)
}

func (s *Service) ListDocumentReferencesByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*DocumentReference, int, error) {
	return s.docRefs.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchDocumentReferences(ctx context.Context, params map[string]string, limit, offset int) ([]*DocumentReference, int, error) {
	return s.docRefs.Search(ctx, params, limit, offset)
}

// -- ClinicalNote --

var validNoteStatuses = map[string]bool{
	"draft": true, "final": true, "amended": true, "entered-in-error": true,
}

func (s *Service) CreateClinicalNote(ctx context.Context, n *ClinicalNote) error {
	if n.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if n.AuthorID == uuid.Nil {
		return fmt.Errorf("author_id is required")
	}
	if n.NoteType == "" {
		return fmt.Errorf("note_type is required")
	}
	if n.Status == "" {
		n.Status = "draft"
	}
	if !validNoteStatuses[n.Status] {
		return fmt.Errorf("invalid status: %s", n.Status)
	}
	return s.notes.Create(ctx, n)
}

func (s *Service) GetClinicalNote(ctx context.Context, id uuid.UUID) (*ClinicalNote, error) {
	return s.notes.GetByID(ctx, id)
}

func (s *Service) UpdateClinicalNote(ctx context.Context, n *ClinicalNote) error {
	if n.Status != "" && !validNoteStatuses[n.Status] {
		return fmt.Errorf("invalid status: %s", n.Status)
	}
	return s.notes.Update(ctx, n)
}

func (s *Service) DeleteClinicalNote(ctx context.Context, id uuid.UUID) error {
	return s.notes.Delete(ctx, id)
}

func (s *Service) ListClinicalNotesByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ClinicalNote, int, error) {
	return s.notes.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) ListClinicalNotesByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*ClinicalNote, int, error) {
	return s.notes.ListByEncounter(ctx, encounterID, limit, offset)
}

// -- Composition --

var validCompStatuses = map[string]bool{
	"preliminary": true, "final": true, "amended": true, "entered-in-error": true,
}

func (s *Service) CreateComposition(ctx context.Context, comp *Composition) error {
	if comp.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if comp.Status == "" {
		comp.Status = "preliminary"
	}
	if !validCompStatuses[comp.Status] {
		return fmt.Errorf("invalid status: %s", comp.Status)
	}
	return s.comps.Create(ctx, comp)
}

func (s *Service) GetComposition(ctx context.Context, id uuid.UUID) (*Composition, error) {
	return s.comps.GetByID(ctx, id)
}

func (s *Service) GetCompositionByFHIRID(ctx context.Context, fhirID string) (*Composition, error) {
	return s.comps.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateComposition(ctx context.Context, comp *Composition) error {
	if comp.Status != "" && !validCompStatuses[comp.Status] {
		return fmt.Errorf("invalid status: %s", comp.Status)
	}
	return s.comps.Update(ctx, comp)
}

func (s *Service) DeleteComposition(ctx context.Context, id uuid.UUID) error {
	return s.comps.Delete(ctx, id)
}

func (s *Service) ListCompositionsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Composition, int, error) {
	return s.comps.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) AddCompositionSection(ctx context.Context, sec *CompositionSection) error {
	if sec.CompositionID == uuid.Nil {
		return fmt.Errorf("composition_id is required")
	}
	return s.comps.AddSection(ctx, sec)
}

func (s *Service) GetCompositionSections(ctx context.Context, compositionID uuid.UUID) ([]*CompositionSection, error) {
	return s.comps.GetSections(ctx, compositionID)
}
