package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/admin"
	"github.com/google/uuid"
)

func TestOrganizationCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("org")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *admin.Organization
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewOrganizationRepo(globalDB.Pool)
			org := &admin.Organization{
				Name:         "General Hospital",
				TypeCode:     "prov",
				Active:       true,
				NPINumber:    ptrStr("1234567890"),
				AddressLine1: ptrStr("123 Main St"),
				City:         ptrStr("Springfield"),
				State:        ptrStr("IL"),
				PostalCode:   ptrStr("62701"),
				Country:      ptrStr("US"),
				Phone:        ptrStr("555-0100"),
				Email:        ptrStr("info@generalhospital.com"),
			}
			if err := repo.Create(ctx, org); err != nil {
				return err
			}
			created = org
			return nil
		})
		if err != nil {
			t.Fatalf("Create organization: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID after create")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID after create")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var orgID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewOrganizationRepo(globalDB.Pool)
			org := &admin.Organization{
				Name:     "GetByID Hospital",
				TypeCode: "prov",
				Active:   true,
			}
			if err := repo.Create(ctx, org); err != nil {
				return err
			}
			orgID = org.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *admin.Organization
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewOrganizationRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, orgID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Name != "GetByID Hospital" {
			t.Errorf("expected Name=GetByID Hospital, got %s", fetched.Name)
		}
		if fetched.TypeCode != "prov" {
			t.Errorf("expected TypeCode=prov, got %s", fetched.TypeCode)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var org *admin.Organization
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewOrganizationRepo(globalDB.Pool)
			o := &admin.Organization{
				Name:     "FHIR Org",
				TypeCode: "dept",
				Active:   true,
			}
			if err := repo.Create(ctx, o); err != nil {
				return err
			}
			org = o
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *admin.Organization
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewOrganizationRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, org.FHIRID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.ID != org.ID {
			t.Errorf("expected ID=%s, got %s", org.ID, fetched.ID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var org *admin.Organization
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewOrganizationRepo(globalDB.Pool)
			o := &admin.Organization{
				Name:     "Update Org",
				TypeCode: "prov",
				Active:   true,
			}
			if err := repo.Create(ctx, o); err != nil {
				return err
			}
			org = o
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewOrganizationRepo(globalDB.Pool)
			org.Name = "Updated Org Name"
			org.Phone = ptrStr("555-9999")
			org.Active = false
			return repo.Update(ctx, org)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *admin.Organization
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewOrganizationRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, org.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Name != "Updated Org Name" {
			t.Errorf("expected Name=Updated Org Name, got %s", fetched.Name)
		}
		if fetched.Phone == nil || *fetched.Phone != "555-9999" {
			t.Errorf("expected Phone=555-9999, got %v", fetched.Phone)
		}
		if fetched.Active {
			t.Error("expected Active=false after update")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*admin.Organization
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewOrganizationRepo(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 organization in list")
		}
		if len(results) != total {
			t.Errorf("expected results count=%d to match total=%d", len(results), total)
		}
	})

	t.Run("Search_ByName", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewOrganizationRepo(globalDB.Pool)
			o := &admin.Organization{
				Name:     "Searchable Clinic",
				TypeCode: "prov",
				Active:   true,
			}
			return repo.Create(ctx, o)
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var results []*admin.Organization
		var total int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewOrganizationRepo(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"name": "Searchable"}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by name: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for name=Searchable")
		}
		for _, r := range results {
			if r.Name != "Searchable Clinic" {
				t.Errorf("expected name containing Searchable, got %s", r.Name)
			}
		}
	})

	t.Run("Search_ByType", func(t *testing.T) {
		var results []*admin.Organization
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewOrganizationRepo(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"type": "prov"}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by type: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for type=prov")
		}
		for _, r := range results {
			if r.TypeCode != "prov" {
				t.Errorf("expected type_code=prov, got %s", r.TypeCode)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var orgID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewOrganizationRepo(globalDB.Pool)
			o := &admin.Organization{
				Name:     "Delete Me Org",
				TypeCode: "prov",
				Active:   true,
			}
			if err := repo.Create(ctx, o); err != nil {
				return err
			}
			orgID = o.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewOrganizationRepo(globalDB.Pool)
			return repo.Delete(ctx, orgID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewOrganizationRepo(globalDB.Pool)
			_, err := repo.GetByID(ctx, orgID)
			return err
		})
		if err == nil {
			t.Fatal("expected error when getting deleted organization, got nil")
		}
	})
}

func TestDepartmentCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("dept")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	// Create prerequisite organization
	orgID := createTestOrganization(t, ctx, globalDB.Pool, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *admin.Department
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewDepartmentRepo(globalDB.Pool)
			dept := &admin.Department{
				OrganizationID: orgID,
				Name:           "Cardiology",
				Code:           ptrStr("CARD"),
				Description:    ptrStr("Cardiology Department"),
				Active:         true,
			}
			if err := repo.Create(ctx, dept); err != nil {
				return err
			}
			created = dept
			return nil
		})
		if err != nil {
			t.Fatalf("Create department: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID after create")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewDepartmentRepo(globalDB.Pool)
			dept := &admin.Department{
				OrganizationID: uuid.New(), // non-existent
				Name:           "Orphan Department",
				Active:         true,
			}
			return repo.Create(ctx, dept)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent organization")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var deptID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewDepartmentRepo(globalDB.Pool)
			dept := &admin.Department{
				OrganizationID: orgID,
				Name:           "Neurology",
				Code:           ptrStr("NEURO"),
				Active:         true,
			}
			if err := repo.Create(ctx, dept); err != nil {
				return err
			}
			deptID = dept.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *admin.Department
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewDepartmentRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, deptID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Name != "Neurology" {
			t.Errorf("expected Name=Neurology, got %s", fetched.Name)
		}
		if fetched.Code == nil || *fetched.Code != "NEURO" {
			t.Errorf("expected Code=NEURO, got %v", fetched.Code)
		}
		if fetched.OrganizationID != orgID {
			t.Errorf("expected OrganizationID=%s, got %s", orgID, fetched.OrganizationID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var dept *admin.Department
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewDepartmentRepo(globalDB.Pool)
			d := &admin.Department{
				OrganizationID: orgID,
				Name:           "Oncology",
				Code:           ptrStr("ONC"),
				Active:         true,
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			dept = d
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewDepartmentRepo(globalDB.Pool)
			dept.Name = "Oncology & Hematology"
			dept.Description = ptrStr("Combined department")
			dept.Active = false
			return repo.Update(ctx, dept)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *admin.Department
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewDepartmentRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, dept.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Name != "Oncology & Hematology" {
			t.Errorf("expected Name=Oncology & Hematology, got %s", fetched.Name)
		}
		if fetched.Description == nil || *fetched.Description != "Combined department" {
			t.Errorf("expected Description=Combined department, got %v", fetched.Description)
		}
		if fetched.Active {
			t.Error("expected Active=false after update")
		}
	})

	t.Run("ListByOrganization", func(t *testing.T) {
		// Create a few more departments for the same org
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewDepartmentRepo(globalDB.Pool)
			d := &admin.Department{
				OrganizationID: orgID,
				Name:           "Radiology",
				Active:         true,
			}
			return repo.Create(ctx, d)
		})
		if err != nil {
			t.Fatalf("Create extra dept: %v", err)
		}

		var results []*admin.Department
		var total int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewDepartmentRepo(globalDB.Pool)
			var err error
			results, total, err = repo.ListByOrganization(ctx, orgID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByOrganization: %v", err)
		}
		if total < 2 {
			t.Errorf("expected at least 2 departments for org, got %d", total)
		}
		for _, r := range results {
			if r.OrganizationID != orgID {
				t.Errorf("expected organization_id=%s, got %s", orgID, r.OrganizationID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var deptID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewDepartmentRepo(globalDB.Pool)
			d := &admin.Department{
				OrganizationID: orgID,
				Name:           "Delete Me Dept",
				Active:         true,
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			deptID = d.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewDepartmentRepo(globalDB.Pool)
			return repo.Delete(ctx, deptID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewDepartmentRepo(globalDB.Pool)
			_, err := repo.GetByID(ctx, deptID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted department")
		}
	})
}

func TestLocationCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("loc")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *admin.Location
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewLocationRepo(globalDB.Pool)
			loc := &admin.Location{
				Status:           "active",
				Name:             "Main Building",
				Description:      ptrStr("Primary hospital building"),
				Mode:             ptrStr("instance"),
				TypeCode:         ptrStr("HOSP"),
				TypeDisplay:      ptrStr("Hospital"),
				PhysicalTypeCode: ptrStr("bu"),
				AddressLine1:     ptrStr("123 Main St"),
				City:             ptrStr("Springfield"),
				State:            ptrStr("IL"),
				PostalCode:       ptrStr("62701"),
				Country:          ptrStr("US"),
				Latitude:         ptrFloat(39.7817),
				Longitude:        ptrFloat(-89.6501),
				Phone:            ptrStr("555-0200"),
				Email:            ptrStr("main@hospital.com"),
			}
			if err := repo.Create(ctx, loc); err != nil {
				return err
			}
			created = loc
			return nil
		})
		if err != nil {
			t.Fatalf("Create location: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID after create")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID after create")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var locID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewLocationRepo(globalDB.Pool)
			loc := &admin.Location{
				Status: "active",
				Name:   "Wing A",
			}
			if err := repo.Create(ctx, loc); err != nil {
				return err
			}
			locID = loc.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *admin.Location
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewLocationRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, locID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Name != "Wing A" {
			t.Errorf("expected Name=Wing A, got %s", fetched.Name)
		}
		if fetched.Status != "active" {
			t.Errorf("expected Status=active, got %s", fetched.Status)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var loc *admin.Location
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewLocationRepo(globalDB.Pool)
			l := &admin.Location{
				Status: "active",
				Name:   "FHIR Location",
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			loc = l
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *admin.Location
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewLocationRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, loc.FHIRID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.ID != loc.ID {
			t.Errorf("expected ID=%s, got %s", loc.ID, fetched.ID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var loc *admin.Location
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewLocationRepo(globalDB.Pool)
			l := &admin.Location{
				Status: "active",
				Name:   "Update Location",
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			loc = l
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewLocationRepo(globalDB.Pool)
			loc.Status = "suspended"
			loc.Name = "Updated Location Name"
			loc.OperationalStatus = ptrStr("C")
			loc.City = ptrStr("Chicago")
			return repo.Update(ctx, loc)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *admin.Location
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewLocationRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, loc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "suspended" {
			t.Errorf("expected Status=suspended, got %s", fetched.Status)
		}
		if fetched.Name != "Updated Location Name" {
			t.Errorf("expected Name=Updated Location Name, got %s", fetched.Name)
		}
		if fetched.OperationalStatus == nil || *fetched.OperationalStatus != "C" {
			t.Errorf("expected OperationalStatus=C, got %v", fetched.OperationalStatus)
		}
		if fetched.City == nil || *fetched.City != "Chicago" {
			t.Errorf("expected City=Chicago, got %v", fetched.City)
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*admin.Location
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewLocationRepo(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 location in list")
		}
		if len(results) != total {
			t.Errorf("expected results count=%d to match total=%d", len(results), total)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var locID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewLocationRepo(globalDB.Pool)
			loc := &admin.Location{
				Status: "active",
				Name:   "Delete Me Location",
			}
			if err := repo.Create(ctx, loc); err != nil {
				return err
			}
			locID = loc.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewLocationRepo(globalDB.Pool)
			return repo.Delete(ctx, locID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewLocationRepo(globalDB.Pool)
			_, err := repo.GetByID(ctx, locID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted location")
		}
	})
}

func TestSystemUserCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("sysuser")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *admin.SystemUser
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			user := &admin.SystemUser{
				Username:    "jdoe",
				UserType:    "clinician",
				Status:      "active",
				DisplayName: ptrStr("John Doe"),
				Email:       ptrStr("jdoe@hospital.com"),
				Phone:       ptrStr("555-0300"),
				MFAEnabled:  false,
				EmployeeID:  ptrStr("EMP-001"),
				HireDate:    ptrTime(time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC)),
				Note:        ptrStr("Test user"),
			}
			if err := repo.Create(ctx, user); err != nil {
				return err
			}
			created = user
			return nil
		})
		if err != nil {
			t.Fatalf("Create system user: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID after create")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var userID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			user := &admin.SystemUser{
				Username: "getbyid_user",
				UserType: "admin",
				Status:   "active",
			}
			if err := repo.Create(ctx, user); err != nil {
				return err
			}
			userID = user.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *admin.SystemUser
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, userID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Username != "getbyid_user" {
			t.Errorf("expected Username=getbyid_user, got %s", fetched.Username)
		}
		if fetched.UserType != "admin" {
			t.Errorf("expected UserType=admin, got %s", fetched.UserType)
		}
	})

	t.Run("GetByUsername", func(t *testing.T) {
		var user *admin.SystemUser
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			u := &admin.SystemUser{
				Username: "unique_username_test",
				UserType: "clinician",
				Status:   "active",
			}
			if err := repo.Create(ctx, u); err != nil {
				return err
			}
			user = u
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *admin.SystemUser
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByUsername(ctx, "unique_username_test")
			return err
		})
		if err != nil {
			t.Fatalf("GetByUsername: %v", err)
		}
		if fetched.ID != user.ID {
			t.Errorf("expected ID=%s, got %s", user.ID, fetched.ID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var user *admin.SystemUser
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			u := &admin.SystemUser{
				Username: "update_user",
				UserType: "clinician",
				Status:   "active",
			}
			if err := repo.Create(ctx, u); err != nil {
				return err
			}
			user = u
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			user.DisplayName = ptrStr("Updated Display Name")
			user.Status = "suspended"
			user.MFAEnabled = true
			user.Note = ptrStr("Account suspended for review")
			return repo.Update(ctx, user)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *admin.SystemUser
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, user.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.DisplayName == nil || *fetched.DisplayName != "Updated Display Name" {
			t.Errorf("expected DisplayName=Updated Display Name, got %v", fetched.DisplayName)
		}
		if fetched.Status != "suspended" {
			t.Errorf("expected Status=suspended, got %s", fetched.Status)
		}
		if !fetched.MFAEnabled {
			t.Error("expected MFAEnabled=true after update")
		}
		if fetched.Note == nil || *fetched.Note != "Account suspended for review" {
			t.Errorf("expected Note='Account suspended for review', got %v", fetched.Note)
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*admin.SystemUser
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 system user in list")
		}
		if len(results) != total {
			t.Errorf("expected results count=%d to match total=%d", len(results), total)
		}
	})

	t.Run("AssignRole_and_GetRoles", func(t *testing.T) {
		var user *admin.SystemUser
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			u := &admin.SystemUser{
				Username: "role_user",
				UserType: "clinician",
				Status:   "active",
			}
			if err := repo.Create(ctx, u); err != nil {
				return err
			}
			user = u
			return nil
		})
		if err != nil {
			t.Fatalf("Create user: %v", err)
		}

		// Assign role
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			assignment := &admin.UserRoleAssignment{
				UserID:    user.ID,
				RoleName:  "physician",
				Active:    true,
				StartDate: time.Now(),
			}
			return repo.AssignRole(ctx, assignment)
		})
		if err != nil {
			t.Fatalf("AssignRole: %v", err)
		}

		// Get roles
		var roles []*admin.UserRoleAssignment
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			var err error
			roles, err = repo.GetRoles(ctx, user.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetRoles: %v", err)
		}
		if len(roles) != 1 {
			t.Fatalf("expected 1 role, got %d", len(roles))
		}
		if roles[0].RoleName != "physician" {
			t.Errorf("expected RoleName=physician, got %s", roles[0].RoleName)
		}
		if !roles[0].Active {
			t.Error("expected role to be active")
		}
	})

	t.Run("RemoveRole", func(t *testing.T) {
		var user *admin.SystemUser
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			u := &admin.SystemUser{
				Username: "remove_role_user",
				UserType: "clinician",
				Status:   "active",
			}
			if err := repo.Create(ctx, u); err != nil {
				return err
			}
			user = u
			return nil
		})
		if err != nil {
			t.Fatalf("Create user: %v", err)
		}

		// Assign role
		var assignmentID uuid.UUID
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			assignment := &admin.UserRoleAssignment{
				UserID:    user.ID,
				RoleName:  "nurse",
				Active:    true,
				StartDate: time.Now(),
			}
			if err := repo.AssignRole(ctx, assignment); err != nil {
				return err
			}
			assignmentID = assignment.ID
			return nil
		})
		if err != nil {
			t.Fatalf("AssignRole: %v", err)
		}

		// Remove role (sets active=false)
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			return repo.RemoveRole(ctx, assignmentID)
		})
		if err != nil {
			t.Fatalf("RemoveRole: %v", err)
		}

		// Verify removal - GetRoles only returns active roles
		var roles []*admin.UserRoleAssignment
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			var err error
			roles, err = repo.GetRoles(ctx, user.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetRoles after remove: %v", err)
		}
		if len(roles) != 0 {
			t.Errorf("expected 0 active roles after remove, got %d", len(roles))
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var userID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			u := &admin.SystemUser{
				Username: "delete_me_user",
				UserType: "admin",
				Status:   "active",
			}
			if err := repo.Create(ctx, u); err != nil {
				return err
			}
			userID = u.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			return repo.Delete(ctx, userID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := admin.NewSystemUserRepo(globalDB.Pool)
			_, err := repo.GetByID(ctx, userID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted system user")
		}
	})
}
