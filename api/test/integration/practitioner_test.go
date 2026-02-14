package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/identity"
	"github.com/google/uuid"
)

func TestPractitionerCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("pract")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *identity.Practitioner
		dob := time.Date(1975, 6, 15, 0, 0, 0, 0, time.UTC)
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			p := &identity.Practitioner{
				Active:               true,
				Prefix:               ptrStr("Dr."),
				FirstName:            "Alice",
				MiddleName:           ptrStr("Marie"),
				LastName:             "Johnson",
				Suffix:               ptrStr("MD"),
				Gender:               ptrStr("female"),
				BirthDate:            &dob,
				NPINumber:            ptrStr("1234567890"),
				DEANumber:            ptrStr("AJ1234567"),
				StateLicenseNum:      ptrStr("IL-12345"),
				StateLicenseState:    ptrStr("IL"),
				Phone:                ptrStr("555-0400"),
				Email:                ptrStr("alice.johnson@hospital.com"),
				AddressLine1:         ptrStr("456 Oak Ave"),
				City:                 ptrStr("Springfield"),
				State:                ptrStr("IL"),
				PostalCode:           ptrStr("62701"),
				Country:              ptrStr("US"),
				QualificationSummary: ptrStr("Board certified cardiologist"),
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			created = p
			return nil
		})
		if err != nil {
			t.Fatalf("Create practitioner: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID after create")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID after create")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		pract := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "GetDoc", "Smith")

		var fetched *identity.Practitioner
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, pract.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.FirstName != "GetDoc" {
			t.Errorf("expected FirstName=GetDoc, got %s", fetched.FirstName)
		}
		if fetched.LastName != "Smith" {
			t.Errorf("expected LastName=Smith, got %s", fetched.LastName)
		}
		if !fetched.Active {
			t.Error("expected Active=true")
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		pract := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "FhirDoc", "Williams")

		var fetched *identity.Practitioner
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, pract.FHIRID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.ID != pract.ID {
			t.Errorf("expected ID=%s, got %s", pract.ID, fetched.ID)
		}
	})

	t.Run("GetByNPI", func(t *testing.T) {
		var pract *identity.Practitioner
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			p := &identity.Practitioner{
				Active:    true,
				FirstName: "NpiDoc",
				LastName:  "Test",
				NPINumber: ptrStr("9876543210"),
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			pract = p
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *identity.Practitioner
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByNPI(ctx, "9876543210")
			return err
		})
		if err != nil {
			t.Fatalf("GetByNPI: %v", err)
		}
		if fetched.ID != pract.ID {
			t.Errorf("expected ID=%s, got %s", pract.ID, fetched.ID)
		}
		if fetched.NPINumber == nil || *fetched.NPINumber != "9876543210" {
			t.Errorf("expected NPINumber=9876543210, got %v", fetched.NPINumber)
		}
	})

	t.Run("Update", func(t *testing.T) {
		pract := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "UpdateDoc", "Before")

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			pract.LastName = "After"
			pract.Email = ptrStr("updated@hospital.com")
			pract.QualificationSummary = ptrStr("Updated qualifications")
			pract.Active = false
			return repo.Update(ctx, pract)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *identity.Practitioner
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, pract.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.LastName != "After" {
			t.Errorf("expected LastName=After, got %s", fetched.LastName)
		}
		if fetched.Email == nil || *fetched.Email != "updated@hospital.com" {
			t.Errorf("expected Email=updated@hospital.com, got %v", fetched.Email)
		}
		if fetched.QualificationSummary == nil || *fetched.QualificationSummary != "Updated qualifications" {
			t.Errorf("expected QualificationSummary=Updated qualifications, got %v", fetched.QualificationSummary)
		}
		if fetched.Active {
			t.Error("expected Active=false after update")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*identity.Practitioner
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 practitioner in list")
		}
		if len(results) != total {
			t.Errorf("expected results count=%d to match total=%d", len(results), total)
		}
	})

	t.Run("Search_ByFamily", func(t *testing.T) {
		createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "SearchDoc1", "Anderson")
		createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "SearchDoc2", "Anderson")

		var results []*identity.Practitioner
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"family": "Anderson"}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by family: %v", err)
		}
		if total < 2 {
			t.Errorf("expected at least 2 results for family=Anderson, got %d", total)
		}
		for _, r := range results {
			if r.LastName != "Anderson" {
				t.Errorf("expected LastName=Anderson, got %s", r.LastName)
			}
		}
	})

	t.Run("Search_ByName", func(t *testing.T) {
		createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "UniqueDocName", "Lastname")

		var results []*identity.Practitioner
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"name": "UniqueDocName"}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by name: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for name=UniqueDocName")
		}
		found := false
		for _, r := range results {
			if r.FirstName == "UniqueDocName" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected to find practitioner with FirstName=UniqueDocName")
		}
	})

	t.Run("Search_ByIdentifier", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			p := &identity.Practitioner{
				Active:    true,
				FirstName: "IdentDoc",
				LastName:  "SearchTest",
				NPINumber: ptrStr("1111111111"),
			}
			return repo.Create(ctx, p)
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var results []*identity.Practitioner
		var total int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"identifier": "1111111111"}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by identifier: %v", err)
		}
		if total != 1 {
			t.Errorf("expected 1 result for identifier=1111111111, got %d", total)
		}
		if len(results) == 1 && (results[0].NPINumber == nil || *results[0].NPINumber != "1111111111") {
			t.Errorf("expected NPI=1111111111, got %v", results[0].NPINumber)
		}
	})

	t.Run("AddRole_GetRoles_RemoveRole", func(t *testing.T) {
		pract := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "RoleDoc", "Test")
		orgID := createTestOrganization(t, ctx, globalDB.Pool, tenantID)

		// Add role
		var roleID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			now := time.Now()
			role := &identity.PractitionerRole{
				PractitionerID:    pract.ID,
				OrganizationID:    &orgID,
				RoleCode:          "physician",
				RoleDisplay:       ptrStr("Attending Physician"),
				PeriodStart:       &now,
				Active:            true,
				TelehealthCapable: true,
				AcceptingPatients: true,
			}
			if err := repo.AddRole(ctx, role); err != nil {
				return err
			}
			roleID = role.ID
			return nil
		})
		if err != nil {
			t.Fatalf("AddRole: %v", err)
		}
		if roleID == uuid.Nil {
			t.Fatal("expected non-nil role ID")
		}

		// Get roles
		var roles []*identity.PractitionerRole
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			var err error
			roles, err = repo.GetRoles(ctx, pract.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetRoles: %v", err)
		}
		if len(roles) != 1 {
			t.Fatalf("expected 1 role, got %d", len(roles))
		}
		if roles[0].RoleCode != "physician" {
			t.Errorf("expected RoleCode=physician, got %s", roles[0].RoleCode)
		}
		if !roles[0].TelehealthCapable {
			t.Error("expected TelehealthCapable=true")
		}
		if !roles[0].AcceptingPatients {
			t.Error("expected AcceptingPatients=true")
		}
		if roles[0].FHIRID == "" {
			t.Error("expected non-empty FHIR ID on role")
		}

		// Remove role
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			return repo.RemoveRole(ctx, roleID)
		})
		if err != nil {
			t.Fatalf("RemoveRole: %v", err)
		}

		// Verify removal
		var rolesAfter []*identity.PractitionerRole
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			var err error
			rolesAfter, err = repo.GetRoles(ctx, pract.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetRoles after remove: %v", err)
		}
		if len(rolesAfter) != 0 {
			t.Errorf("expected 0 roles after remove, got %d", len(rolesAfter))
		}
	})

	t.Run("MultipleRoles", func(t *testing.T) {
		pract := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "MultiRoleDoc", "Test")
		orgID := createTestOrganization(t, ctx, globalDB.Pool, tenantID)

		// Add two roles
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			role1 := &identity.PractitionerRole{
				PractitionerID: pract.ID,
				OrganizationID: &orgID,
				RoleCode:       "physician",
				Active:         true,
			}
			if err := repo.AddRole(ctx, role1); err != nil {
				return err
			}
			role2 := &identity.PractitionerRole{
				PractitionerID: pract.ID,
				OrganizationID: &orgID,
				RoleCode:       "surgeon",
				Active:         true,
			}
			return repo.AddRole(ctx, role2)
		})
		if err != nil {
			t.Fatalf("AddRoles: %v", err)
		}

		var roles []*identity.PractitionerRole
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			var err error
			roles, err = repo.GetRoles(ctx, pract.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetRoles: %v", err)
		}
		if len(roles) != 2 {
			t.Fatalf("expected 2 roles, got %d", len(roles))
		}
	})

	t.Run("Delete", func(t *testing.T) {
		pract := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "DeleteDoc", "Test")

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			return repo.Delete(ctx, pract.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := identity.NewPractitionerRepo(globalDB.Pool)
			_, err := repo.GetByID(ctx, pract.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted practitioner")
		}
	})
}
