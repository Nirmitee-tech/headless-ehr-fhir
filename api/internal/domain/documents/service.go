package documents

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	consents   ConsentRepository
	docRefs    DocumentReferenceRepository
	notes      ClinicalNoteRepository
	comps      CompositionRepository
	templates  DocumentTemplateRepository
}

func NewService(consents ConsentRepository, docRefs DocumentReferenceRepository, notes ClinicalNoteRepository, comps CompositionRepository, templates ...DocumentTemplateRepository) *Service {
	s := &Service{consents: consents, docRefs: docRefs, notes: notes, comps: comps}
	if len(templates) > 0 {
		s.templates = templates[0]
	}
	return s
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

// -- DocumentTemplate --

var validTemplateStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true,
}

func (s *Service) CreateTemplate(ctx context.Context, t *DocumentTemplate) error {
	if t.Name == "" {
		return fmt.Errorf("name is required")
	}
	if t.Status == "" {
		t.Status = "draft"
	}
	if !validTemplateStatuses[t.Status] {
		return fmt.Errorf("invalid status: %s", t.Status)
	}
	if err := s.templates.Create(ctx, t); err != nil {
		return err
	}
	// Create sections if provided inline
	for i := range t.Sections {
		t.Sections[i].TemplateID = t.ID
		if err := s.templates.AddSection(ctx, &t.Sections[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) GetTemplate(ctx context.Context, id uuid.UUID) (*DocumentTemplate, error) {
	t, err := s.templates.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	sections, err := s.templates.GetSections(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, sec := range sections {
		t.Sections = append(t.Sections, *sec)
	}
	return t, nil
}

func (s *Service) UpdateTemplate(ctx context.Context, t *DocumentTemplate) error {
	if t.Status != "" && !validTemplateStatuses[t.Status] {
		return fmt.Errorf("invalid status: %s", t.Status)
	}
	return s.templates.Update(ctx, t)
}

func (s *Service) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	return s.templates.Delete(ctx, id)
}

func (s *Service) ListTemplates(ctx context.Context, limit, offset int) ([]*DocumentTemplate, int, error) {
	return s.templates.List(ctx, limit, offset)
}

// RenderTemplate renders a template by substituting {{variable}} placeholders with provided values.
func (s *Service) RenderTemplate(ctx context.Context, templateID uuid.UUID, variables map[string]string) (*RenderedDocument, error) {
	t, err := s.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	rendered := &RenderedDocument{
		TemplateID:   t.ID,
		TemplateName: t.Name,
		RenderedAt:   time.Now(),
	}

	for _, sec := range t.Sections {
		content := sec.ContentTemplate
		for key, val := range variables {
			content = strings.ReplaceAll(content, "{{"+key+"}}", val)
		}
		rendered.Sections = append(rendered.Sections, RenderedSection{
			Title:   sec.Title,
			Content: content,
		})
	}

	return rendered, nil
}
