package documents

import (
	"context"
	"fmt"
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

// -- Tests --

func newTestService() *Service {
	return NewService(newMockConsentRepo(), newMockDocRefRepo(), newMockClinicalNoteRepo(), newMockCompositionRepo())
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
