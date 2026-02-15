package documents

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
)

// -- Mock Repositories --

type mockConsentRepo struct {
	items map[uuid.UUID]*Consent
}

func newMockConsentRepo() *mockConsentRepo {
	return &mockConsentRepo{items: make(map[uuid.UUID]*Consent)}
}

func (m *mockConsentRepo) Create(_ context.Context, c *Consent) error {
	c.ID = uuid.New()
	if c.FHIRID == "" {
		c.FHIRID = c.ID.String()
	}
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	m.items[c.ID] = c
	return nil
}

func (m *mockConsentRepo) GetByID(_ context.Context, id uuid.UUID) (*Consent, error) {
	c, ok := m.items[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

func (m *mockConsentRepo) GetByFHIRID(_ context.Context, fhirID string) (*Consent, error) {
	for _, c := range m.items {
		if c.FHIRID == fhirID {
			return c, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockConsentRepo) Update(_ context.Context, c *Consent) error {
	m.items[c.ID] = c
	return nil
}

func (m *mockConsentRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockConsentRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*Consent, int, error) {
	var result []*Consent
	for _, c := range m.items {
		if c.PatientID == patientID {
			result = append(result, c)
		}
	}
	return result, len(result), nil
}

func (m *mockConsentRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*Consent, int, error) {
	var result []*Consent
	for _, c := range m.items {
		result = append(result, c)
	}
	return result, len(result), nil
}

type mockDocRefRepo struct {
	items map[uuid.UUID]*DocumentReference
}

func newMockDocRefRepo() *mockDocRefRepo {
	return &mockDocRefRepo{items: make(map[uuid.UUID]*DocumentReference)}
}

func (m *mockDocRefRepo) Create(_ context.Context, d *DocumentReference) error {
	d.ID = uuid.New()
	if d.FHIRID == "" {
		d.FHIRID = d.ID.String()
	}
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
	m.items[d.ID] = d
	return nil
}

func (m *mockDocRefRepo) GetByID(_ context.Context, id uuid.UUID) (*DocumentReference, error) {
	d, ok := m.items[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return d, nil
}

func (m *mockDocRefRepo) GetByFHIRID(_ context.Context, fhirID string) (*DocumentReference, error) {
	for _, d := range m.items {
		if d.FHIRID == fhirID {
			return d, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockDocRefRepo) Update(_ context.Context, d *DocumentReference) error {
	m.items[d.ID] = d
	return nil
}

func (m *mockDocRefRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockDocRefRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*DocumentReference, int, error) {
	var result []*DocumentReference
	for _, d := range m.items {
		if d.PatientID == patientID {
			result = append(result, d)
		}
	}
	return result, len(result), nil
}

func (m *mockDocRefRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*DocumentReference, int, error) {
	var result []*DocumentReference
	for _, d := range m.items {
		result = append(result, d)
	}
	return result, len(result), nil
}

type mockClinicalNoteRepo struct {
	items map[uuid.UUID]*ClinicalNote
}

func newMockClinicalNoteRepo() *mockClinicalNoteRepo {
	return &mockClinicalNoteRepo{items: make(map[uuid.UUID]*ClinicalNote)}
}

func (m *mockClinicalNoteRepo) Create(_ context.Context, n *ClinicalNote) error {
	n.ID = uuid.New()
	n.CreatedAt = time.Now()
	n.UpdatedAt = time.Now()
	m.items[n.ID] = n
	return nil
}

func (m *mockClinicalNoteRepo) GetByID(_ context.Context, id uuid.UUID) (*ClinicalNote, error) {
	n, ok := m.items[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return n, nil
}

func (m *mockClinicalNoteRepo) Update(_ context.Context, n *ClinicalNote) error {
	m.items[n.ID] = n
	return nil
}

func (m *mockClinicalNoteRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockClinicalNoteRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*ClinicalNote, int, error) {
	var result []*ClinicalNote
	for _, n := range m.items {
		if n.PatientID == patientID {
			result = append(result, n)
		}
	}
	return result, len(result), nil
}

func (m *mockClinicalNoteRepo) ListByEncounter(_ context.Context, encounterID uuid.UUID, limit, offset int) ([]*ClinicalNote, int, error) {
	var result []*ClinicalNote
	for _, n := range m.items {
		if n.EncounterID != nil && *n.EncounterID == encounterID {
			result = append(result, n)
		}
	}
	return result, len(result), nil
}

type mockCompositionRepo struct {
	items    map[uuid.UUID]*Composition
	sections map[uuid.UUID]*CompositionSection
}

func newMockCompositionRepo() *mockCompositionRepo {
	return &mockCompositionRepo{
		items:    make(map[uuid.UUID]*Composition),
		sections: make(map[uuid.UUID]*CompositionSection),
	}
}

func (m *mockCompositionRepo) Create(_ context.Context, c *Composition) error {
	c.ID = uuid.New()
	if c.FHIRID == "" {
		c.FHIRID = c.ID.String()
	}
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	m.items[c.ID] = c
	return nil
}

func (m *mockCompositionRepo) GetByID(_ context.Context, id uuid.UUID) (*Composition, error) {
	c, ok := m.items[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

func (m *mockCompositionRepo) GetByFHIRID(_ context.Context, fhirID string) (*Composition, error) {
	for _, c := range m.items {
		if c.FHIRID == fhirID {
			return c, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockCompositionRepo) Update(_ context.Context, c *Composition) error {
	m.items[c.ID] = c
	return nil
}

func (m *mockCompositionRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockCompositionRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*Composition, int, error) {
	var result []*Composition
	for _, c := range m.items {
		if c.PatientID == patientID {
			result = append(result, c)
		}
	}
	return result, len(result), nil
}

func (m *mockCompositionRepo) AddSection(_ context.Context, s *CompositionSection) error {
	s.ID = uuid.New()
	m.sections[s.ID] = s
	return nil
}

func (m *mockCompositionRepo) GetSections(_ context.Context, compositionID uuid.UUID) ([]*CompositionSection, error) {
	var result []*CompositionSection
	for _, s := range m.sections {
		if s.CompositionID == compositionID {
			result = append(result, s)
		}
	}
	return result, nil
}

// -- Mock DocumentTemplate Repository --

type mockDocTemplateRepo struct {
	items    map[uuid.UUID]*DocumentTemplate
	sections map[uuid.UUID]*TemplateSection
}

func newMockDocTemplateRepo() *mockDocTemplateRepo {
	return &mockDocTemplateRepo{
		items:    make(map[uuid.UUID]*DocumentTemplate),
		sections: make(map[uuid.UUID]*TemplateSection),
	}
}

func (m *mockDocTemplateRepo) Create(_ context.Context, t *DocumentTemplate) error {
	t.ID = uuid.New()
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	// Store a copy without Sections to avoid double-counting when GetTemplate
	// appends sections from GetSections.
	stored := *t
	stored.Sections = nil
	m.items[t.ID] = &stored
	return nil
}

func (m *mockDocTemplateRepo) GetByID(_ context.Context, id uuid.UUID) (*DocumentTemplate, error) {
	t, ok := m.items[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return t, nil
}

func (m *mockDocTemplateRepo) Update(_ context.Context, t *DocumentTemplate) error {
	m.items[t.ID] = t
	return nil
}

func (m *mockDocTemplateRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockDocTemplateRepo) List(_ context.Context, limit, offset int) ([]*DocumentTemplate, int, error) {
	var result []*DocumentTemplate
	for _, t := range m.items {
		result = append(result, t)
	}
	return result, len(result), nil
}

func (m *mockDocTemplateRepo) AddSection(_ context.Context, s *TemplateSection) error {
	s.ID = uuid.New()
	m.sections[s.ID] = s
	return nil
}

func (m *mockDocTemplateRepo) GetSections(_ context.Context, templateID uuid.UUID) ([]*TemplateSection, error) {
	var result []*TemplateSection
	for _, s := range m.sections {
		if s.TemplateID == templateID {
			result = append(result, s)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].SortOrder < result[j].SortOrder
	})
	return result, nil
}

// -- Tests --

func newTestService() *Service {
	return NewService(newMockConsentRepo(), newMockDocRefRepo(), newMockClinicalNoteRepo(), newMockCompositionRepo(), newMockDocTemplateRepo())
}

// -- Consent Tests --

func TestCreateConsent(t *testing.T) {
	svc := newTestService()
	c := &Consent{PatientID: uuid.New()}
	err := svc.CreateConsent(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Status != "draft" {
		t.Errorf("expected default status 'draft', got %s", c.Status)
	}
}

func TestCreateConsent_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	c := &Consent{}
	err := svc.CreateConsent(context.Background(), c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestGetConsent(t *testing.T) {
	svc := newTestService()
	c := &Consent{PatientID: uuid.New()}
	svc.CreateConsent(context.Background(), c)

	fetched, err := svc.GetConsent(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != c.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestDeleteConsent(t *testing.T) {
	svc := newTestService()
	c := &Consent{PatientID: uuid.New()}
	svc.CreateConsent(context.Background(), c)
	err := svc.DeleteConsent(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetConsent(context.Background(), c.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

// -- DocumentReference Tests --

func TestCreateDocumentReference(t *testing.T) {
	svc := newTestService()
	d := &DocumentReference{PatientID: uuid.New()}
	err := svc.CreateDocumentReference(context.Background(), d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Status != "current" {
		t.Errorf("expected default status 'current', got %s", d.Status)
	}
}

func TestCreateDocumentReference_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	d := &DocumentReference{}
	err := svc.CreateDocumentReference(context.Background(), d)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestGetDocumentReference(t *testing.T) {
	svc := newTestService()
	d := &DocumentReference{PatientID: uuid.New()}
	svc.CreateDocumentReference(context.Background(), d)

	fetched, err := svc.GetDocumentReference(context.Background(), d.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != d.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestDeleteDocumentReference(t *testing.T) {
	svc := newTestService()
	d := &DocumentReference{PatientID: uuid.New()}
	svc.CreateDocumentReference(context.Background(), d)
	err := svc.DeleteDocumentReference(context.Background(), d.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetDocumentReference(context.Background(), d.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

// -- ClinicalNote Tests --

func TestCreateClinicalNote(t *testing.T) {
	svc := newTestService()
	n := &ClinicalNote{PatientID: uuid.New(), AuthorID: uuid.New(), NoteType: "progress"}
	err := svc.CreateClinicalNote(context.Background(), n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Status != "draft" {
		t.Errorf("expected default status 'draft', got %s", n.Status)
	}
}

func TestCreateClinicalNote_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	n := &ClinicalNote{AuthorID: uuid.New(), NoteType: "progress"}
	err := svc.CreateClinicalNote(context.Background(), n)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateClinicalNote_AuthorIDRequired(t *testing.T) {
	svc := newTestService()
	n := &ClinicalNote{PatientID: uuid.New(), NoteType: "progress"}
	err := svc.CreateClinicalNote(context.Background(), n)
	if err == nil {
		t.Error("expected error for missing author_id")
	}
}

func TestCreateClinicalNote_NoteTypeRequired(t *testing.T) {
	svc := newTestService()
	n := &ClinicalNote{PatientID: uuid.New(), AuthorID: uuid.New()}
	err := svc.CreateClinicalNote(context.Background(), n)
	if err == nil {
		t.Error("expected error for missing note_type")
	}
}

func TestGetClinicalNote(t *testing.T) {
	svc := newTestService()
	n := &ClinicalNote{PatientID: uuid.New(), AuthorID: uuid.New(), NoteType: "progress"}
	svc.CreateClinicalNote(context.Background(), n)

	fetched, err := svc.GetClinicalNote(context.Background(), n.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != n.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestDeleteClinicalNote(t *testing.T) {
	svc := newTestService()
	n := &ClinicalNote{PatientID: uuid.New(), AuthorID: uuid.New(), NoteType: "progress"}
	svc.CreateClinicalNote(context.Background(), n)
	err := svc.DeleteClinicalNote(context.Background(), n.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetClinicalNote(context.Background(), n.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

// -- Composition Tests --

func TestCreateComposition(t *testing.T) {
	svc := newTestService()
	c := &Composition{PatientID: uuid.New()}
	err := svc.CreateComposition(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Status != "preliminary" {
		t.Errorf("expected default status 'preliminary', got %s", c.Status)
	}
}

func TestCreateComposition_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	c := &Composition{}
	err := svc.CreateComposition(context.Background(), c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestGetComposition(t *testing.T) {
	svc := newTestService()
	c := &Composition{PatientID: uuid.New()}
	svc.CreateComposition(context.Background(), c)

	fetched, err := svc.GetComposition(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != c.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestDeleteComposition(t *testing.T) {
	svc := newTestService()
	c := &Composition{PatientID: uuid.New()}
	svc.CreateComposition(context.Background(), c)
	err := svc.DeleteComposition(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetComposition(context.Background(), c.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestAddCompositionSection(t *testing.T) {
	svc := newTestService()
	c := &Composition{PatientID: uuid.New()}
	svc.CreateComposition(context.Background(), c)

	title := "History of Present Illness"
	sec := &CompositionSection{CompositionID: c.ID, Title: &title}
	err := svc.AddCompositionSection(context.Background(), sec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sections, err := svc.GetCompositionSections(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sections) != 1 {
		t.Errorf("expected 1 section, got %d", len(sections))
	}
}

func TestAddCompositionSection_CompositionIDRequired(t *testing.T) {
	svc := newTestService()
	sec := &CompositionSection{}
	err := svc.AddCompositionSection(context.Background(), sec)
	if err == nil {
		t.Error("expected error for missing composition_id")
	}
}

// -- Additional Consent Tests --

func TestGetConsentByFHIRID(t *testing.T) {
	svc := newTestService()
	c := &Consent{PatientID: uuid.New()}
	svc.CreateConsent(context.Background(), c)

	fetched, err := svc.GetConsentByFHIRID(context.Background(), c.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != c.ID {
		t.Errorf("expected same ID")
	}
}

func TestGetConsentByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetConsentByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateConsent(t *testing.T) {
	svc := newTestService()
	c := &Consent{PatientID: uuid.New()}
	svc.CreateConsent(context.Background(), c)

	c.Status = "active"
	err := svc.UpdateConsent(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateConsent_InvalidStatus(t *testing.T) {
	svc := newTestService()
	c := &Consent{PatientID: uuid.New()}
	svc.CreateConsent(context.Background(), c)

	c.Status = "bogus"
	err := svc.UpdateConsent(context.Background(), c)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestListConsentsByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreateConsent(context.Background(), &Consent{PatientID: patientID})
	svc.CreateConsent(context.Background(), &Consent{PatientID: uuid.New()})

	result, total, err := svc.ListConsentsByPatient(context.Background(), patientID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
}

func TestSearchConsents(t *testing.T) {
	svc := newTestService()
	svc.CreateConsent(context.Background(), &Consent{PatientID: uuid.New()})

	result, total, err := svc.SearchConsents(context.Background(), map[string]string{}, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(result) < 1 {
		t.Error("expected results")
	}
}

// -- Additional DocumentReference Tests --

func TestGetDocumentReferenceByFHIRID(t *testing.T) {
	svc := newTestService()
	d := &DocumentReference{PatientID: uuid.New()}
	svc.CreateDocumentReference(context.Background(), d)

	fetched, err := svc.GetDocumentReferenceByFHIRID(context.Background(), d.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != d.ID {
		t.Errorf("expected same ID")
	}
}

func TestGetDocumentReferenceByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetDocumentReferenceByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateDocumentReference(t *testing.T) {
	svc := newTestService()
	d := &DocumentReference{PatientID: uuid.New()}
	svc.CreateDocumentReference(context.Background(), d)

	d.Status = "superseded"
	err := svc.UpdateDocumentReference(context.Background(), d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateDocumentReference_InvalidStatus(t *testing.T) {
	svc := newTestService()
	d := &DocumentReference{PatientID: uuid.New()}
	svc.CreateDocumentReference(context.Background(), d)

	d.Status = "bogus"
	err := svc.UpdateDocumentReference(context.Background(), d)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestListDocumentReferencesByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreateDocumentReference(context.Background(), &DocumentReference{PatientID: patientID})
	svc.CreateDocumentReference(context.Background(), &DocumentReference{PatientID: uuid.New()})

	result, total, err := svc.ListDocumentReferencesByPatient(context.Background(), patientID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
}

func TestSearchDocumentReferences(t *testing.T) {
	svc := newTestService()
	svc.CreateDocumentReference(context.Background(), &DocumentReference{PatientID: uuid.New()})

	result, total, err := svc.SearchDocumentReferences(context.Background(), map[string]string{}, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(result) < 1 {
		t.Error("expected results")
	}
}

// -- Additional ClinicalNote Tests --

func TestUpdateClinicalNote(t *testing.T) {
	svc := newTestService()
	n := &ClinicalNote{PatientID: uuid.New(), AuthorID: uuid.New(), NoteType: "progress"}
	svc.CreateClinicalNote(context.Background(), n)

	n.Status = "final"
	err := svc.UpdateClinicalNote(context.Background(), n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateClinicalNote_InvalidStatus(t *testing.T) {
	svc := newTestService()
	n := &ClinicalNote{PatientID: uuid.New(), AuthorID: uuid.New(), NoteType: "progress"}
	svc.CreateClinicalNote(context.Background(), n)

	n.Status = "bogus"
	err := svc.UpdateClinicalNote(context.Background(), n)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestListClinicalNotesByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreateClinicalNote(context.Background(), &ClinicalNote{PatientID: patientID, AuthorID: uuid.New(), NoteType: "progress"})
	svc.CreateClinicalNote(context.Background(), &ClinicalNote{PatientID: uuid.New(), AuthorID: uuid.New(), NoteType: "progress"})

	result, total, err := svc.ListClinicalNotesByPatient(context.Background(), patientID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
}

func TestListClinicalNotesByEncounter(t *testing.T) {
	svc := newTestService()
	encounterID := uuid.New()
	otherEncounterID := uuid.New()
	svc.CreateClinicalNote(context.Background(), &ClinicalNote{PatientID: uuid.New(), AuthorID: uuid.New(), NoteType: "progress", EncounterID: &encounterID})
	svc.CreateClinicalNote(context.Background(), &ClinicalNote{PatientID: uuid.New(), AuthorID: uuid.New(), NoteType: "progress", EncounterID: &otherEncounterID})

	result, total, err := svc.ListClinicalNotesByEncounter(context.Background(), encounterID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
}

// -- Additional Composition Tests --

func TestGetCompositionByFHIRID(t *testing.T) {
	svc := newTestService()
	c := &Composition{PatientID: uuid.New()}
	svc.CreateComposition(context.Background(), c)

	fetched, err := svc.GetCompositionByFHIRID(context.Background(), c.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != c.ID {
		t.Errorf("expected same ID")
	}
}

func TestGetCompositionByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetCompositionByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateComposition(t *testing.T) {
	svc := newTestService()
	c := &Composition{PatientID: uuid.New()}
	svc.CreateComposition(context.Background(), c)

	c.Status = "final"
	err := svc.UpdateComposition(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateComposition_InvalidStatus(t *testing.T) {
	svc := newTestService()
	c := &Composition{PatientID: uuid.New()}
	svc.CreateComposition(context.Background(), c)

	c.Status = "bogus"
	err := svc.UpdateComposition(context.Background(), c)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestListCompositionsByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreateComposition(context.Background(), &Composition{PatientID: patientID})
	svc.CreateComposition(context.Background(), &Composition{PatientID: uuid.New()})

	result, total, err := svc.ListCompositionsByPatient(context.Background(), patientID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
}

// -- ToFHIR Tests --

func TestConsentToFHIR(t *testing.T) {
	c := &Consent{
		FHIRID:    "consent-123",
		Status:    "active",
		PatientID: uuid.New(),
		UpdatedAt: time.Now(),
	}
	fhirRes := c.ToFHIR()
	if fhirRes["resourceType"] != "Consent" {
		t.Errorf("expected Consent, got %v", fhirRes["resourceType"])
	}
	if fhirRes["status"] != "active" {
		t.Errorf("expected active, got %v", fhirRes["status"])
	}
}

func TestDocumentReferenceToFHIR(t *testing.T) {
	d := &DocumentReference{
		FHIRID:    "docref-123",
		Status:    "current",
		PatientID: uuid.New(),
		UpdatedAt: time.Now(),
	}
	fhirRes := d.ToFHIR()
	if fhirRes["resourceType"] != "DocumentReference" {
		t.Errorf("expected DocumentReference, got %v", fhirRes["resourceType"])
	}
	if fhirRes["status"] != "current" {
		t.Errorf("expected current, got %v", fhirRes["status"])
	}
}

func TestCompositionToFHIR(t *testing.T) {
	c := &Composition{
		FHIRID:    "comp-123",
		Status:    "final",
		PatientID: uuid.New(),
		UpdatedAt: time.Now(),
	}
	fhirRes := c.ToFHIR()
	if fhirRes["resourceType"] != "Composition" {
		t.Errorf("expected Composition, got %v", fhirRes["resourceType"])
	}
	if fhirRes["status"] != "final" {
		t.Errorf("expected final, got %v", fhirRes["status"])
	}
}

// -- DocumentTemplate Tests --

func TestCreateTemplate(t *testing.T) {
	svc := newTestService()
	tmpl := &DocumentTemplate{Name: "Discharge Summary"}
	err := svc.CreateTemplate(context.Background(), tmpl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if tmpl.Status != "draft" {
		t.Errorf("expected default status 'draft', got %s", tmpl.Status)
	}
}

func TestCreateTemplate_NameRequired(t *testing.T) {
	svc := newTestService()
	tmpl := &DocumentTemplate{}
	err := svc.CreateTemplate(context.Background(), tmpl)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestCreateTemplate_InvalidStatus(t *testing.T) {
	svc := newTestService()
	tmpl := &DocumentTemplate{Name: "Test", Status: "bogus"}
	err := svc.CreateTemplate(context.Background(), tmpl)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestCreateTemplate_WithSections(t *testing.T) {
	svc := newTestService()
	tmpl := &DocumentTemplate{
		Name: "Discharge Summary",
		Sections: []TemplateSection{
			{Title: "Chief Complaint", ContentTemplate: "Patient presented with {{complaint}}.", SortOrder: 1},
			{Title: "Assessment", ContentTemplate: "Assessment: {{assessment}}", SortOrder: 2},
		},
	}
	err := svc.CreateTemplate(context.Background(), tmpl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestGetTemplate(t *testing.T) {
	svc := newTestService()
	tmpl := &DocumentTemplate{Name: "History and Physical"}
	svc.CreateTemplate(context.Background(), tmpl)

	fetched, err := svc.GetTemplate(context.Background(), tmpl.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.Name != "History and Physical" {
		t.Errorf("expected 'History and Physical', got %s", fetched.Name)
	}
}

func TestGetTemplate_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetTemplate(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateTemplate(t *testing.T) {
	svc := newTestService()
	tmpl := &DocumentTemplate{Name: "Progress Note"}
	svc.CreateTemplate(context.Background(), tmpl)

	tmpl.Status = "active"
	err := svc.UpdateTemplate(context.Background(), tmpl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateTemplate_InvalidStatus(t *testing.T) {
	svc := newTestService()
	tmpl := &DocumentTemplate{Name: "Progress Note"}
	svc.CreateTemplate(context.Background(), tmpl)

	tmpl.Status = "bogus"
	err := svc.UpdateTemplate(context.Background(), tmpl)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestDeleteTemplate(t *testing.T) {
	svc := newTestService()
	tmpl := &DocumentTemplate{Name: "Old Template"}
	svc.CreateTemplate(context.Background(), tmpl)

	err := svc.DeleteTemplate(context.Background(), tmpl.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.GetTemplate(context.Background(), tmpl.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestListTemplates(t *testing.T) {
	svc := newTestService()
	svc.CreateTemplate(context.Background(), &DocumentTemplate{Name: "Template A"})
	svc.CreateTemplate(context.Background(), &DocumentTemplate{Name: "Template B"})

	result, total, err := svc.ListTemplates(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}
}

func TestRenderTemplate(t *testing.T) {
	svc := newTestService()

	// Create template with sections containing placeholders
	tmpl := &DocumentTemplate{
		Name: "Discharge Summary",
		Sections: []TemplateSection{
			{Title: "Patient Info", ContentTemplate: "Patient: {{patient_name}}, DOB: {{dob}}", SortOrder: 1},
			{Title: "Diagnosis", ContentTemplate: "Primary Diagnosis: {{diagnosis}}", SortOrder: 2},
		},
	}
	err := svc.CreateTemplate(context.Background(), tmpl)
	if err != nil {
		t.Fatalf("unexpected error creating template: %v", err)
	}

	variables := map[string]string{
		"patient_name": "John Doe",
		"dob":          "1990-05-15",
		"diagnosis":    "Pneumonia",
	}

	rendered, err := svc.RenderTemplate(context.Background(), tmpl.ID, variables)
	if err != nil {
		t.Fatalf("unexpected error rendering: %v", err)
	}

	if rendered.TemplateID != tmpl.ID {
		t.Error("expected template ID to match")
	}
	if rendered.TemplateName != "Discharge Summary" {
		t.Errorf("expected 'Discharge Summary', got %s", rendered.TemplateName)
	}
	if len(rendered.Sections) != 2 {
		t.Fatalf("expected 2 rendered sections, got %d", len(rendered.Sections))
	}
	if rendered.Sections[0].Content != "Patient: John Doe, DOB: 1990-05-15" {
		t.Errorf("unexpected rendered content: %s", rendered.Sections[0].Content)
	}
	if rendered.Sections[1].Content != "Primary Diagnosis: Pneumonia" {
		t.Errorf("unexpected rendered content: %s", rendered.Sections[1].Content)
	}
	if rendered.RenderedAt.IsZero() {
		t.Error("expected RenderedAt to be set")
	}
}

func TestRenderTemplate_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.RenderTemplate(context.Background(), uuid.New(), map[string]string{})
	if err == nil {
		t.Error("expected error for non-existent template")
	}
}

func TestRenderTemplate_NoVariables(t *testing.T) {
	svc := newTestService()
	tmpl := &DocumentTemplate{
		Name: "Simple Template",
		Sections: []TemplateSection{
			{Title: "Body", ContentTemplate: "This has {{placeholder}} in it.", SortOrder: 1},
		},
	}
	svc.CreateTemplate(context.Background(), tmpl)

	rendered, err := svc.RenderTemplate(context.Background(), tmpl.ID, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Without variable substitution, placeholder stays
	if rendered.Sections[0].Content != "This has {{placeholder}} in it." {
		t.Errorf("unexpected content: %s", rendered.Sections[0].Content)
	}
}

func TestCreateTemplate_ValidStatuses(t *testing.T) {
	validStatuses := []string{"draft", "active", "retired"}
	for _, status := range validStatuses {
		svc := newTestService()
		tmpl := &DocumentTemplate{Name: "Test " + status, Status: status}
		err := svc.CreateTemplate(context.Background(), tmpl)
		if err != nil {
			t.Errorf("expected no error for valid status %s, got: %v", status, err)
		}
	}
}
