package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/identity"
	"github.com/google/uuid"
)

func TestPatientCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("patient")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *identity.Patient
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			dob := time.Date(1990, 3, 15, 0, 0, 0, 0, time.UTC)
			p := &identity.Patient{
				Active:    true,
				MRN:       "MRN-CREATE-001",
				FirstName: "John",
				LastName:  "Doe",
				Gender:    ptrStr("male"),
				BirthDate: &dob,
				Email:     ptrStr("john.doe@example.com"),
				City:      ptrStr("Springfield"),
				State:     ptrStr("IL"),
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			created = p
			return nil
		})
		if err != nil {
			t.Fatalf("Create patient: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID after create")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID after create")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "Jane", "Smith", "MRN-GET-001")

		var fetched *identity.Patient
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, patient.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.FirstName != "Jane" {
			t.Errorf("expected FirstName=Jane, got %s", fetched.FirstName)
		}
		if fetched.LastName != "Smith" {
			t.Errorf("expected LastName=Smith, got %s", fetched.LastName)
		}
		if fetched.MRN != "MRN-GET-001" {
			t.Errorf("expected MRN=MRN-GET-001, got %s", fetched.MRN)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "FhirFirst", "FhirLast", "MRN-FHIR-001")

		var fetched *identity.Patient
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, patient.FHIRID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.ID != patient.ID {
			t.Errorf("expected ID=%s, got %s", patient.ID, fetched.ID)
		}
	})

	t.Run("GetByMRN", func(t *testing.T) {
		patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "MrnFirst", "MrnLast", "MRN-BYMRN-001")

		var fetched *identity.Patient
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByMRN(ctx, "MRN-BYMRN-001")
			return err
		})
		if err != nil {
			t.Fatalf("GetByMRN: %v", err)
		}
		if fetched.ID != patient.ID {
			t.Errorf("expected ID=%s, got %s", patient.ID, fetched.ID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "UpdateFirst", "UpdateLast", "MRN-UPD-001")

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			patient.FirstName = "UpdatedFirst"
			patient.Email = ptrStr("updated@example.com")
			patient.City = ptrStr("Chicago")
			return repo.Update(ctx, patient)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		// Verify the update
		var fetched *identity.Patient
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, patient.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.FirstName != "UpdatedFirst" {
			t.Errorf("expected FirstName=UpdatedFirst, got %s", fetched.FirstName)
		}
		if fetched.Email == nil || *fetched.Email != "updated@example.com" {
			t.Errorf("expected Email=updated@example.com, got %v", fetched.Email)
		}
		if fetched.City == nil || *fetched.City != "Chicago" {
			t.Errorf("expected City=Chicago, got %v", fetched.City)
		}
	})

	t.Run("Search_ByName", func(t *testing.T) {
		createTestPatient(t, ctx, globalDB.Pool, tenantID, "SearchAlice", "Johnson", "MRN-SRCH-001")
		createTestPatient(t, ctx, globalDB.Pool, tenantID, "SearchBob", "Johnson", "MRN-SRCH-002")
		createTestPatient(t, ctx, globalDB.Pool, tenantID, "SearchCharlie", "Williams", "MRN-SRCH-003")

		var results []*identity.Patient
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"family": "Johnson"}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by family: %v", err)
		}
		if total < 2 {
			t.Errorf("expected at least 2 results for family=Johnson, got %d", total)
		}
		for _, r := range results {
			if r.LastName != "Johnson" {
				t.Errorf("expected all results to have LastName=Johnson, got %s", r.LastName)
			}
		}
	})

	t.Run("Search_ByGender", func(t *testing.T) {
		var results []*identity.Patient
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"gender": "male"}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by gender: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for gender=male")
		}
		for _, r := range results {
			if r.Gender == nil || *r.Gender != "male" {
				t.Errorf("expected gender=male, got %v", r.Gender)
			}
		}
	})

	t.Run("Search_ByIdentifier", func(t *testing.T) {
		p := createTestPatient(t, ctx, globalDB.Pool, tenantID, "IdentFirst", "IdentLast", "MRN-IDENT-SEARCH")

		var results []*identity.Patient
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"identifier": "MRN-IDENT-SEARCH"}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by identifier: %v", err)
		}
		if total != 1 {
			t.Errorf("expected 1 result for identifier=MRN-IDENT-SEARCH, got %d", total)
		}
		if len(results) == 1 && results[0].ID != p.ID {
			t.Errorf("expected ID=%s, got %s", p.ID, results[0].ID)
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*identity.Patient
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 patient in list")
		}
		if len(results) != total {
			t.Errorf("expected results count=%d to match total=%d", len(results), total)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "DeleteMe", "Please", "MRN-DEL-001")

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			return repo.Delete(ctx, patient.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		// Verify deletion
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			_, err := repo.GetByID(ctx, patient.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error when getting deleted patient, got nil")
		}
	})

	t.Run("UniqueConstraint_MRN", func(t *testing.T) {
		createTestPatient(t, ctx, globalDB.Pool, tenantID, "First1", "Last1", "MRN-UNIQUE-001")

		// Try to create another patient with the same MRN
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			p := &identity.Patient{
				Active:    true,
				MRN:       "MRN-UNIQUE-001",
				FirstName: "First2",
				LastName:  "Last2",
			}
			return repo.Create(ctx, p)
		})
		if err == nil {
			t.Fatal("expected error for duplicate MRN, got nil")
		}
	})

	t.Run("Contacts", func(t *testing.T) {
		patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "ContactPatient", "Test", "MRN-CONTACT-001")

		// Add contact
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			contact := &identity.PatientContact{
				PatientID:    patient.ID,
				Relationship: "emergency",
				FirstName:    ptrStr("Emergency"),
				LastName:     ptrStr("Contact"),
				Phone:        ptrStr("555-1234"),
			}
			return repo.AddContact(ctx, contact)
		})
		if err != nil {
			t.Fatalf("AddContact: %v", err)
		}

		// Get contacts
		var contacts []*identity.PatientContact
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			var err error
			contacts, err = repo.GetContacts(ctx, patient.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetContacts: %v", err)
		}
		if len(contacts) != 1 {
			t.Fatalf("expected 1 contact, got %d", len(contacts))
		}
		if contacts[0].Relationship != "emergency" {
			t.Errorf("expected relationship=emergency, got %s", contacts[0].Relationship)
		}

		// Remove contact
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			return repo.RemoveContact(ctx, contacts[0].ID)
		})
		if err != nil {
			t.Fatalf("RemoveContact: %v", err)
		}

		// Verify removal
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			var err error
			contacts, err = repo.GetContacts(ctx, patient.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetContacts after remove: %v", err)
		}
		if len(contacts) != 0 {
			t.Errorf("expected 0 contacts after remove, got %d", len(contacts))
		}
	})

	t.Run("Identifiers", func(t *testing.T) {
		patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "IdentPatient", "Test", "MRN-IDENTIFIERS-001")

		// Add identifier
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			ident := &identity.PatientIdentifier{
				PatientID: patient.ID,
				SystemURI: "http://example.com/ids",
				Value:     "EXT-12345",
				TypeCode:  ptrStr("MR"),
			}
			return repo.AddIdentifier(ctx, ident)
		})
		if err != nil {
			t.Fatalf("AddIdentifier: %v", err)
		}

		// Get identifiers
		var idents []*identity.PatientIdentifier
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPatientRepo(globalDB.Pool)
			var err error
			idents, err = repo.GetIdentifiers(ctx, patient.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetIdentifiers: %v", err)
		}
		if len(idents) != 1 {
			t.Fatalf("expected 1 identifier, got %d", len(idents))
		}
		if idents[0].Value != "EXT-12345" {
			t.Errorf("expected value=EXT-12345, got %s", idents[0].Value)
		}
	})
}
