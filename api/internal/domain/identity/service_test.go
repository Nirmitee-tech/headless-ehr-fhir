package identity

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// -- Mock Patient Repository --

type mockPatientRepo struct {
	patients    map[uuid.UUID]*Patient
	contacts    map[uuid.UUID]*PatientContact
	identifiers map[uuid.UUID]*PatientIdentifier
}

func newMockPatientRepo() *mockPatientRepo {
	return &mockPatientRepo{
		patients:    make(map[uuid.UUID]*Patient),
		contacts:    make(map[uuid.UUID]*PatientContact),
		identifiers: make(map[uuid.UUID]*PatientIdentifier),
	}
}

func (m *mockPatientRepo) Create(_ context.Context, p *Patient) error {
	p.ID = uuid.New()
	if p.FHIRID == "" {
		p.FHIRID = p.ID.String()
	}
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	m.patients[p.ID] = p
	return nil
}

func (m *mockPatientRepo) GetByID(_ context.Context, id uuid.UUID) (*Patient, error) {
	p, ok := m.patients[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return p, nil
}

func (m *mockPatientRepo) GetByFHIRID(_ context.Context, fhirID string) (*Patient, error) {
	for _, p := range m.patients {
		if p.FHIRID == fhirID {
			return p, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockPatientRepo) GetByMRN(_ context.Context, mrn string) (*Patient, error) {
	for _, p := range m.patients {
		if p.MRN == mrn {
			return p, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockPatientRepo) Update(_ context.Context, p *Patient) error {
	m.patients[p.ID] = p
	return nil
}

func (m *mockPatientRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.patients, id)
	return nil
}

func (m *mockPatientRepo) List(_ context.Context, limit, offset int) ([]*Patient, int, error) {
	var result []*Patient
	for _, p := range m.patients {
		result = append(result, p)
	}
	return result, len(result), nil
}

func (m *mockPatientRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*Patient, int, error) {
	return m.List(context.Background(), limit, offset)
}

func (m *mockPatientRepo) AddContact(_ context.Context, c *PatientContact) error {
	c.ID = uuid.New()
	m.contacts[c.ID] = c
	return nil
}

func (m *mockPatientRepo) GetContacts(_ context.Context, patientID uuid.UUID) ([]*PatientContact, error) {
	var result []*PatientContact
	for _, c := range m.contacts {
		if c.PatientID == patientID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockPatientRepo) RemoveContact(_ context.Context, id uuid.UUID) error {
	delete(m.contacts, id)
	return nil
}

func (m *mockPatientRepo) AddIdentifier(_ context.Context, ident *PatientIdentifier) error {
	ident.ID = uuid.New()
	m.identifiers[ident.ID] = ident
	return nil
}

func (m *mockPatientRepo) GetIdentifiers(_ context.Context, patientID uuid.UUID) ([]*PatientIdentifier, error) {
	var result []*PatientIdentifier
	for _, i := range m.identifiers {
		if i.PatientID == patientID {
			result = append(result, i)
		}
	}
	return result, nil
}

func (m *mockPatientRepo) RemoveIdentifier(_ context.Context, id uuid.UUID) error {
	delete(m.identifiers, id)
	return nil
}

// -- Mock Practitioner Repository --

type mockPractRepo struct {
	practitioners map[uuid.UUID]*Practitioner
	roles         map[uuid.UUID]*PractitionerRole
}

func newMockPractRepo() *mockPractRepo {
	return &mockPractRepo{
		practitioners: make(map[uuid.UUID]*Practitioner),
		roles:         make(map[uuid.UUID]*PractitionerRole),
	}
}

func (m *mockPractRepo) Create(_ context.Context, p *Practitioner) error {
	p.ID = uuid.New()
	if p.FHIRID == "" {
		p.FHIRID = p.ID.String()
	}
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	m.practitioners[p.ID] = p
	return nil
}

func (m *mockPractRepo) GetByID(_ context.Context, id uuid.UUID) (*Practitioner, error) {
	p, ok := m.practitioners[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return p, nil
}

func (m *mockPractRepo) GetByFHIRID(_ context.Context, fhirID string) (*Practitioner, error) {
	for _, p := range m.practitioners {
		if p.FHIRID == fhirID {
			return p, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockPractRepo) GetByNPI(_ context.Context, npi string) (*Practitioner, error) {
	for _, p := range m.practitioners {
		if p.NPINumber != nil && *p.NPINumber == npi {
			return p, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockPractRepo) Update(_ context.Context, p *Practitioner) error {
	m.practitioners[p.ID] = p
	return nil
}

func (m *mockPractRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.practitioners, id)
	return nil
}

func (m *mockPractRepo) List(_ context.Context, limit, offset int) ([]*Practitioner, int, error) {
	var result []*Practitioner
	for _, p := range m.practitioners {
		result = append(result, p)
	}
	return result, len(result), nil
}

func (m *mockPractRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*Practitioner, int, error) {
	return m.List(context.Background(), limit, offset)
}

func (m *mockPractRepo) AddRole(_ context.Context, role *PractitionerRole) error {
	role.ID = uuid.New()
	if role.FHIRID == "" {
		role.FHIRID = role.ID.String()
	}
	m.roles[role.ID] = role
	return nil
}

func (m *mockPractRepo) GetRoles(_ context.Context, practitionerID uuid.UUID) ([]*PractitionerRole, error) {
	var result []*PractitionerRole
	for _, r := range m.roles {
		if r.PractitionerID == practitionerID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockPractRepo) RemoveRole(_ context.Context, id uuid.UUID) error {
	delete(m.roles, id)
	return nil
}

// -- Tests --

func newTestService() *Service {
	return NewService(newMockPatientRepo(), newMockPractRepo())
}

func TestCreatePatient(t *testing.T) {
	svc := newTestService()

	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN001"}
	err := svc.CreatePatient(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if !p.Active {
		t.Error("expected active to be true")
	}
	if p.FHIRID == "" {
		t.Error("expected FHIR ID to be set")
	}
}

func TestCreatePatient_NameRequired(t *testing.T) {
	svc := newTestService()

	p := &Patient{MRN: "MRN001", LastName: "Doe"}
	err := svc.CreatePatient(context.Background(), p)
	if err == nil {
		t.Error("expected error for missing first_name")
	}

	p2 := &Patient{MRN: "MRN001", FirstName: "John"}
	err = svc.CreatePatient(context.Background(), p2)
	if err == nil {
		t.Error("expected error for missing last_name")
	}
}

func TestCreatePatient_MRNRequired(t *testing.T) {
	svc := newTestService()

	p := &Patient{FirstName: "John", LastName: "Doe"}
	err := svc.CreatePatient(context.Background(), p)
	if err == nil {
		t.Error("expected error for missing MRN")
	}
}

func TestGetPatient(t *testing.T) {
	svc := newTestService()

	p := &Patient{FirstName: "Jane", LastName: "Smith", MRN: "MRN002"}
	svc.CreatePatient(context.Background(), p)

	fetched, err := svc.GetPatient(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.FirstName != "Jane" {
		t.Errorf("expected Jane, got %s", fetched.FirstName)
	}
}

func TestGetPatientByMRN(t *testing.T) {
	svc := newTestService()

	p := &Patient{FirstName: "Jane", LastName: "Smith", MRN: "MRN-UNIQUE"}
	svc.CreatePatient(context.Background(), p)

	fetched, err := svc.GetPatientByMRN(context.Background(), "MRN-UNIQUE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != p.ID {
		t.Errorf("expected same ID, got %s vs %s", fetched.ID, p.ID)
	}
}

func TestDeletePatient(t *testing.T) {
	svc := newTestService()

	p := &Patient{FirstName: "Jane", LastName: "Smith", MRN: "MRN003"}
	svc.CreatePatient(context.Background(), p)

	err := svc.DeletePatient(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.GetPatient(context.Background(), p.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestPatientContacts(t *testing.T) {
	svc := newTestService()

	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN004"}
	svc.CreatePatient(context.Background(), p)

	name := "Jane"
	contact := &PatientContact{
		PatientID:    p.ID,
		Relationship: "emergency",
		FirstName:    &name,
	}
	err := svc.AddPatientContact(context.Background(), contact)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	contacts, err := svc.GetPatientContacts(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(contacts) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(contacts))
	}
	if contacts[0].Relationship != "emergency" {
		t.Errorf("expected emergency, got %s", contacts[0].Relationship)
	}
}

func TestPatientIdentifiers(t *testing.T) {
	svc := newTestService()

	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN005"}
	svc.CreatePatient(context.Background(), p)

	ident := &PatientIdentifier{
		PatientID: p.ID,
		SystemURI: "http://hospital.com/mrn",
		Value:     "MRN005",
	}
	err := svc.AddPatientIdentifier(context.Background(), ident)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	idents, err := svc.GetPatientIdentifiers(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(idents) != 1 {
		t.Fatalf("expected 1 identifier, got %d", len(idents))
	}
}

func TestCreatePractitioner(t *testing.T) {
	svc := newTestService()

	p := &Practitioner{FirstName: "Dr. Sarah", LastName: "Johnson"}
	err := svc.CreatePractitioner(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.Active {
		t.Error("expected active to be true")
	}
}

func TestCreatePractitioner_NameRequired(t *testing.T) {
	svc := newTestService()

	p := &Practitioner{LastName: "Johnson"}
	err := svc.CreatePractitioner(context.Background(), p)
	if err == nil {
		t.Error("expected error for missing first_name")
	}
}

func TestPractitionerRoles(t *testing.T) {
	svc := newTestService()

	p := &Practitioner{FirstName: "Sarah", LastName: "Johnson"}
	svc.CreatePractitioner(context.Background(), p)

	role := &PractitionerRole{
		PractitionerID: p.ID,
		RoleCode:       "doctor",
	}
	err := svc.AddPractitionerRole(context.Background(), role)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !role.Active {
		t.Error("expected role to be active")
	}

	roles, err := svc.GetPractitionerRoles(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roles) != 1 {
		t.Fatalf("expected 1 role, got %d", len(roles))
	}
}

func TestPractitionerRole_Validation(t *testing.T) {
	svc := newTestService()

	// Missing practitioner_id
	role := &PractitionerRole{RoleCode: "doctor"}
	err := svc.AddPractitionerRole(context.Background(), role)
	if err == nil {
		t.Error("expected error for missing practitioner_id")
	}

	// Missing role_code
	role2 := &PractitionerRole{PractitionerID: uuid.New()}
	err = svc.AddPractitionerRole(context.Background(), role2)
	if err == nil {
		t.Error("expected error for missing role_code")
	}
}

func TestGetPatientByFHIRID(t *testing.T) {
	svc := newTestService()
	p := &Patient{FirstName: "Jane", LastName: "Smith", MRN: "MRN-F1"}
	svc.CreatePatient(context.Background(), p)

	fetched, err := svc.GetPatientByFHIRID(context.Background(), p.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != p.ID {
		t.Errorf("expected same ID")
	}
}

func TestGetPatientByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetPatientByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdatePatient(t *testing.T) {
	svc := newTestService()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN-U1"}
	svc.CreatePatient(context.Background(), p)

	p.FirstName = "Jonathan"
	err := svc.UpdatePatient(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdatePatient_NameRequired(t *testing.T) {
	svc := newTestService()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN-U2"}
	svc.CreatePatient(context.Background(), p)

	p.FirstName = ""
	err := svc.UpdatePatient(context.Background(), p)
	if err == nil {
		t.Error("expected error for missing first_name")
	}
}

func TestListPatients(t *testing.T) {
	svc := newTestService()
	svc.CreatePatient(context.Background(), &Patient{FirstName: "A", LastName: "B", MRN: "M1"})
	svc.CreatePatient(context.Background(), &Patient{FirstName: "C", LastName: "D", MRN: "M2"})

	result, total, err := svc.ListPatients(context.Background(), 10, 0)
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

func TestSearchPatients(t *testing.T) {
	svc := newTestService()
	svc.CreatePatient(context.Background(), &Patient{FirstName: "John", LastName: "Doe", MRN: "M-S1"})

	result, total, err := svc.SearchPatients(context.Background(), map[string]string{"name": "Doe"}, 10, 0)
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

func TestRemovePatientContact(t *testing.T) {
	svc := newTestService()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN-RC"}
	svc.CreatePatient(context.Background(), p)

	name := "Jane"
	contact := &PatientContact{PatientID: p.ID, Relationship: "emergency", FirstName: &name}
	svc.AddPatientContact(context.Background(), contact)

	err := svc.RemovePatientContact(context.Background(), contact.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	contacts, _ := svc.GetPatientContacts(context.Background(), p.ID)
	if len(contacts) != 0 {
		t.Errorf("expected 0 contacts after removal, got %d", len(contacts))
	}
}

func TestRemovePatientIdentifier(t *testing.T) {
	svc := newTestService()
	p := &Patient{FirstName: "John", LastName: "Doe", MRN: "MRN-RI"}
	svc.CreatePatient(context.Background(), p)

	ident := &PatientIdentifier{PatientID: p.ID, SystemURI: "http://hospital.com/mrn", Value: "MRN-RI"}
	svc.AddPatientIdentifier(context.Background(), ident)

	err := svc.RemovePatientIdentifier(context.Background(), ident.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	idents, _ := svc.GetPatientIdentifiers(context.Background(), p.ID)
	if len(idents) != 0 {
		t.Errorf("expected 0 identifiers after removal, got %d", len(idents))
	}
}

func TestGetPractitioner(t *testing.T) {
	svc := newTestService()
	p := &Practitioner{FirstName: "Sarah", LastName: "Johnson"}
	svc.CreatePractitioner(context.Background(), p)

	fetched, err := svc.GetPractitioner(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.FirstName != "Sarah" {
		t.Errorf("expected Sarah, got %s", fetched.FirstName)
	}
}

func TestGetPractitioner_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetPractitioner(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestGetPractitionerByFHIRID(t *testing.T) {
	svc := newTestService()
	p := &Practitioner{FirstName: "Sarah", LastName: "Johnson"}
	svc.CreatePractitioner(context.Background(), p)

	fetched, err := svc.GetPractitionerByFHIRID(context.Background(), p.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != p.ID {
		t.Errorf("expected same ID")
	}
}

func TestGetPractitionerByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetPractitionerByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestGetPractitionerByNPI(t *testing.T) {
	svc := newTestService()
	npi := "1234567890"
	p := &Practitioner{FirstName: "Sarah", LastName: "Johnson", NPINumber: &npi}
	svc.CreatePractitioner(context.Background(), p)

	fetched, err := svc.GetPractitionerByNPI(context.Background(), "1234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != p.ID {
		t.Errorf("expected same ID")
	}
}

func TestGetPractitionerByNPI_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetPractitionerByNPI(context.Background(), "0000000000")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdatePractitioner(t *testing.T) {
	svc := newTestService()
	p := &Practitioner{FirstName: "Sarah", LastName: "Johnson"}
	svc.CreatePractitioner(context.Background(), p)

	p.FirstName = "Dr. Sarah"
	err := svc.UpdatePractitioner(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdatePractitioner_NameRequired(t *testing.T) {
	svc := newTestService()
	p := &Practitioner{FirstName: "Sarah", LastName: "Johnson"}
	svc.CreatePractitioner(context.Background(), p)

	p.FirstName = ""
	err := svc.UpdatePractitioner(context.Background(), p)
	if err == nil {
		t.Error("expected error for missing first_name")
	}
}

func TestDeletePractitioner(t *testing.T) {
	svc := newTestService()
	p := &Practitioner{FirstName: "Sarah", LastName: "Johnson"}
	svc.CreatePractitioner(context.Background(), p)
	err := svc.DeletePractitioner(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetPractitioner(context.Background(), p.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestListPractitioners(t *testing.T) {
	svc := newTestService()
	svc.CreatePractitioner(context.Background(), &Practitioner{FirstName: "A", LastName: "B"})
	svc.CreatePractitioner(context.Background(), &Practitioner{FirstName: "C", LastName: "D"})

	result, total, err := svc.ListPractitioners(context.Background(), 10, 0)
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

func TestSearchPractitioners(t *testing.T) {
	svc := newTestService()
	svc.CreatePractitioner(context.Background(), &Practitioner{FirstName: "Sarah", LastName: "Johnson"})

	result, total, err := svc.SearchPractitioners(context.Background(), map[string]string{"name": "Johnson"}, 10, 0)
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

func TestRemovePractitionerRole(t *testing.T) {
	svc := newTestService()
	p := &Practitioner{FirstName: "Sarah", LastName: "Johnson"}
	svc.CreatePractitioner(context.Background(), p)

	role := &PractitionerRole{PractitionerID: p.ID, RoleCode: "doctor"}
	svc.AddPractitionerRole(context.Background(), role)

	err := svc.RemovePractitionerRole(context.Background(), role.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	roles, _ := svc.GetPractitionerRoles(context.Background(), p.ID)
	if len(roles) != 0 {
		t.Errorf("expected 0 roles after removal, got %d", len(roles))
	}
}

func TestPatientToFHIR(t *testing.T) {
	mobile := "+1-555-9999"
	email := "john@example.com"
	addr := "456 Oak Ave"
	city := "Chicago"
	state := "IL"
	gender := "male"
	bd := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)

	p := &Patient{
		FHIRID:      "pat-123",
		Active:      true,
		FirstName:   "John",
		LastName:    "Doe",
		MRN:         "MRN123",
		Gender:      &gender,
		BirthDate:   &bd,
		PhoneMobile: &mobile,
		Email:       &email,
		AddressLine1: &addr,
		City:        &city,
		State:       &state,
		UpdatedAt:   time.Now(),
	}

	fhir := p.ToFHIR()

	if fhir["resourceType"] != "Patient" {
		t.Errorf("expected Patient, got %v", fhir["resourceType"])
	}
	if fhir["id"] != "pat-123" {
		t.Errorf("expected pat-123, got %v", fhir["id"])
	}
	if fhir["active"] != true {
		t.Error("expected active true")
	}
	if fhir["gender"] != "male" {
		t.Errorf("expected male, got %v", fhir["gender"])
	}
	if fhir["birthDate"] != "1990-05-15" {
		t.Errorf("expected 1990-05-15, got %v", fhir["birthDate"])
	}
	if fhir["telecom"] == nil {
		t.Error("expected telecom")
	}
	if fhir["address"] == nil {
		t.Error("expected address")
	}
	if fhir["identifier"] == nil {
		t.Error("expected identifier with MRN")
	}
}

func TestPractitionerToFHIR(t *testing.T) {
	npi := "1234567890"
	p := &Practitioner{
		FHIRID:    "pract-1",
		Active:    true,
		FirstName: "Sarah",
		LastName:  "Johnson",
		NPINumber: &npi,
		UpdatedAt: time.Now(),
	}

	fhir := p.ToFHIR()

	if fhir["resourceType"] != "Practitioner" {
		t.Errorf("expected Practitioner, got %v", fhir["resourceType"])
	}
	if fhir["id"] != "pract-1" {
		t.Errorf("expected pract-1, got %v", fhir["id"])
	}
	if fhir["identifier"] == nil {
		t.Error("expected identifier with NPI")
	}
}
