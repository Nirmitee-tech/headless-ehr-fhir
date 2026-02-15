package integration

import (
	"context"
	"testing"

	"github.com/ehr/ehr/internal/domain/conformance"
	"github.com/google/uuid"
)

// ==================== NamingSystem Tests ====================

func TestNamingSystemCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("conf")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *conformance.NamingSystem
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewNamingSystemRepoPG(globalDB.Pool)
			ns := &conformance.NamingSystem{
				Name:        "US-SSN",
				Status:      "active",
				Kind:        "identifier",
				Date:        ptrStr("2024-01-01"),
				Publisher:   ptrStr("HL7 International"),
				Description: ptrStr("United States Social Security Number"),
				TypeCode:    ptrStr("SS"),
				Responsible: ptrStr("HHS"),
			}
			if err := repo.Create(ctx, ns); err != nil {
				return err
			}
			created = ns
			return nil
		})
		if err != nil {
			t.Fatalf("Create naming system: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var nsID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewNamingSystemRepoPG(globalDB.Pool)
			ns := &conformance.NamingSystem{
				Name:   "GetByID-NS",
				Status: "active",
				Kind:   "identifier",
			}
			if err := repo.Create(ctx, ns); err != nil {
				return err
			}
			nsID = ns.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *conformance.NamingSystem
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewNamingSystemRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, nsID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Name != "GetByID-NS" {
			t.Errorf("expected Name=GetByID-NS, got %s", fetched.Name)
		}
		if fetched.Status != "active" {
			t.Errorf("expected Status=active, got %s", fetched.Status)
		}
		if fetched.Kind != "identifier" {
			t.Errorf("expected Kind=identifier, got %s", fetched.Kind)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var ns *conformance.NamingSystem
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewNamingSystemRepoPG(globalDB.Pool)
			n := &conformance.NamingSystem{
				Name:   "FHIR-NS",
				Status: "active",
				Kind:   "codesystem",
			}
			if err := repo.Create(ctx, n); err != nil {
				return err
			}
			ns = n
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *conformance.NamingSystem
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewNamingSystemRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, ns.FHIRID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.ID != ns.ID {
			t.Errorf("expected ID=%s, got %s", ns.ID, fetched.ID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var ns *conformance.NamingSystem
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewNamingSystemRepoPG(globalDB.Pool)
			n := &conformance.NamingSystem{
				Name:   "Update-NS",
				Status: "draft",
				Kind:   "identifier",
			}
			if err := repo.Create(ctx, n); err != nil {
				return err
			}
			ns = n
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewNamingSystemRepoPG(globalDB.Pool)
			ns.Name = "Updated-NS-Name"
			ns.Status = "active"
			ns.Publisher = ptrStr("Updated Publisher")
			ns.Description = ptrStr("Updated description")
			return repo.Update(ctx, ns)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *conformance.NamingSystem
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewNamingSystemRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, ns.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Name != "Updated-NS-Name" {
			t.Errorf("expected Name=Updated-NS-Name, got %s", fetched.Name)
		}
		if fetched.Status != "active" {
			t.Errorf("expected Status=active, got %s", fetched.Status)
		}
		if fetched.Publisher == nil || *fetched.Publisher != "Updated Publisher" {
			t.Errorf("expected Publisher=Updated Publisher, got %v", fetched.Publisher)
		}
	})

	t.Run("Search_ByName", func(t *testing.T) {
		// Create a searchable naming system
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewNamingSystemRepoPG(globalDB.Pool)
			ns := &conformance.NamingSystem{
				Name:   "Searchable-NamingSystem",
				Status: "active",
				Kind:   "identifier",
			}
			return repo.Create(ctx, ns)
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var results []*conformance.NamingSystem
		var total int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewNamingSystemRepoPG(globalDB.Pool)
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
		found := false
		for _, r := range results {
			if r.Name == "Searchable-NamingSystem" {
				found = true
			}
		}
		if !found {
			t.Error("expected to find Searchable-NamingSystem in results")
		}
	})

	t.Run("Search_ByStatus", func(t *testing.T) {
		var results []*conformance.NamingSystem
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewNamingSystemRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"status": "active"}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by status: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for status=active")
		}
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var nsID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewNamingSystemRepoPG(globalDB.Pool)
			ns := &conformance.NamingSystem{
				Name:   "Delete-NS",
				Status: "active",
				Kind:   "identifier",
			}
			if err := repo.Create(ctx, ns); err != nil {
				return err
			}
			nsID = ns.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewNamingSystemRepoPG(globalDB.Pool)
			return repo.Delete(ctx, nsID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewNamingSystemRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, nsID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted naming system")
		}
	})
}

// ==================== OperationDefinition Tests ====================

func TestOperationDefinitionCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("conf")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *conformance.OperationDefinition
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewOperationDefinitionRepoPG(globalDB.Pool)
			od := &conformance.OperationDefinition{
				URL:       ptrStr("http://hl7.org/fhir/OperationDefinition/Patient-everything"),
				Name:      "patient-everything",
				Title:     ptrStr("Patient Everything"),
				Status:    "active",
				Kind:      "operation",
				Code:      "everything",
				System:    ptrBool(false),
				Type:      ptrBool(true),
				Instance:  ptrBool(true),
				Publisher: ptrStr("HL7 International"),
			}
			if err := repo.Create(ctx, od); err != nil {
				return err
			}
			created = od
			return nil
		})
		if err != nil {
			t.Fatalf("Create operation definition: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var odID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewOperationDefinitionRepoPG(globalDB.Pool)
			od := &conformance.OperationDefinition{
				Name:   "getbyid-op",
				Status: "active",
				Kind:   "operation",
				Code:   "validate",
			}
			if err := repo.Create(ctx, od); err != nil {
				return err
			}
			odID = od.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *conformance.OperationDefinition
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewOperationDefinitionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, odID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Name != "getbyid-op" {
			t.Errorf("expected Name=getbyid-op, got %s", fetched.Name)
		}
		if fetched.Code != "validate" {
			t.Errorf("expected Code=validate, got %s", fetched.Code)
		}
		if fetched.Kind != "operation" {
			t.Errorf("expected Kind=operation, got %s", fetched.Kind)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var od *conformance.OperationDefinition
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewOperationDefinitionRepoPG(globalDB.Pool)
			o := &conformance.OperationDefinition{
				Name:   "fhirid-op",
				Status: "active",
				Kind:   "query",
				Code:   "lookup",
			}
			if err := repo.Create(ctx, o); err != nil {
				return err
			}
			od = o
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *conformance.OperationDefinition
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewOperationDefinitionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, od.FHIRID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.ID != od.ID {
			t.Errorf("expected ID=%s, got %s", od.ID, fetched.ID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var od *conformance.OperationDefinition
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewOperationDefinitionRepoPG(globalDB.Pool)
			o := &conformance.OperationDefinition{
				Name:   "update-op",
				Status: "draft",
				Kind:   "operation",
				Code:   "expand",
			}
			if err := repo.Create(ctx, o); err != nil {
				return err
			}
			od = o
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewOperationDefinitionRepoPG(globalDB.Pool)
			od.Name = "updated-op-name"
			od.Status = "active"
			od.Code = "expand-updated"
			od.Publisher = ptrStr("Updated Publisher")
			od.System = ptrBool(true)
			return repo.Update(ctx, od)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *conformance.OperationDefinition
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewOperationDefinitionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, od.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Name != "updated-op-name" {
			t.Errorf("expected Name=updated-op-name, got %s", fetched.Name)
		}
		if fetched.Status != "active" {
			t.Errorf("expected Status=active, got %s", fetched.Status)
		}
		if fetched.Code != "expand-updated" {
			t.Errorf("expected Code=expand-updated, got %s", fetched.Code)
		}
		if fetched.Publisher == nil || *fetched.Publisher != "Updated Publisher" {
			t.Errorf("expected Publisher=Updated Publisher, got %v", fetched.Publisher)
		}
		if fetched.System == nil || !*fetched.System {
			t.Errorf("expected System=true, got %v", fetched.System)
		}
	})

	t.Run("Search_ByName", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewOperationDefinitionRepoPG(globalDB.Pool)
			od := &conformance.OperationDefinition{
				Name:   "searchable-operation",
				Status: "active",
				Kind:   "operation",
				Code:   "meta",
			}
			return repo.Create(ctx, od)
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var results []*conformance.OperationDefinition
		var total int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewOperationDefinitionRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"name": "searchable"}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by name: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for name=searchable")
		}
		found := false
		for _, r := range results {
			if r.Name == "searchable-operation" {
				found = true
			}
		}
		if !found {
			t.Error("expected to find searchable-operation in results")
		}
	})

	t.Run("Search_ByCode", func(t *testing.T) {
		var results []*conformance.OperationDefinition
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewOperationDefinitionRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"code": "meta"}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by code: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for code=meta")
		}
		for _, r := range results {
			if r.Code != "meta" {
				t.Errorf("expected code=meta, got %s", r.Code)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var odID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewOperationDefinitionRepoPG(globalDB.Pool)
			od := &conformance.OperationDefinition{
				Name:   "delete-op",
				Status: "active",
				Kind:   "operation",
				Code:   "delete-test",
			}
			if err := repo.Create(ctx, od); err != nil {
				return err
			}
			odID = od.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewOperationDefinitionRepoPG(globalDB.Pool)
			return repo.Delete(ctx, odID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewOperationDefinitionRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, odID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted operation definition")
		}
	})
}

// ==================== MessageDefinition Tests ====================

func TestMessageDefinitionCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("conf")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *conformance.MessageDefinition
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageDefinitionRepoPG(globalDB.Pool)
			md := &conformance.MessageDefinition{
				URL:               ptrStr("http://example.org/fhir/MessageDefinition/patient-admit"),
				Name:              ptrStr("PatientAdmitNotification"),
				Title:             ptrStr("Patient Admit Notification"),
				Status:            "active",
				Date:              ptrStr("2024-06-01"),
				Publisher:         ptrStr("Example Health System"),
				EventCodingCode:   "admin-notify",
				EventCodingSystem: ptrStr("http://example.org/fhir/message-events"),
				Category:          ptrStr("notification"),
			}
			if err := repo.Create(ctx, md); err != nil {
				return err
			}
			created = md
			return nil
		})
		if err != nil {
			t.Fatalf("Create message definition: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var mdID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageDefinitionRepoPG(globalDB.Pool)
			md := &conformance.MessageDefinition{
				Status:          "active",
				EventCodingCode: "lab-result",
			}
			if err := repo.Create(ctx, md); err != nil {
				return err
			}
			mdID = md.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *conformance.MessageDefinition
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageDefinitionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, mdID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.EventCodingCode != "lab-result" {
			t.Errorf("expected EventCodingCode=lab-result, got %s", fetched.EventCodingCode)
		}
		if fetched.Status != "active" {
			t.Errorf("expected Status=active, got %s", fetched.Status)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var md *conformance.MessageDefinition
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageDefinitionRepoPG(globalDB.Pool)
			m := &conformance.MessageDefinition{
				Status:          "active",
				EventCodingCode: "fhirid-event",
			}
			if err := repo.Create(ctx, m); err != nil {
				return err
			}
			md = m
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *conformance.MessageDefinition
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageDefinitionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, md.FHIRID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.ID != md.ID {
			t.Errorf("expected ID=%s, got %s", md.ID, fetched.ID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var md *conformance.MessageDefinition
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageDefinitionRepoPG(globalDB.Pool)
			m := &conformance.MessageDefinition{
				Name:            ptrStr("UpdateTest"),
				Status:          "draft",
				EventCodingCode: "update-event",
			}
			if err := repo.Create(ctx, m); err != nil {
				return err
			}
			md = m
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageDefinitionRepoPG(globalDB.Pool)
			md.Name = ptrStr("UpdatedMsgDef")
			md.Status = "active"
			md.EventCodingCode = "updated-event"
			md.Description = ptrStr("Updated description for message def")
			md.Category = ptrStr("consequence")
			return repo.Update(ctx, md)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *conformance.MessageDefinition
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageDefinitionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, md.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Name == nil || *fetched.Name != "UpdatedMsgDef" {
			t.Errorf("expected Name=UpdatedMsgDef, got %v", fetched.Name)
		}
		if fetched.Status != "active" {
			t.Errorf("expected Status=active, got %s", fetched.Status)
		}
		if fetched.EventCodingCode != "updated-event" {
			t.Errorf("expected EventCodingCode=updated-event, got %s", fetched.EventCodingCode)
		}
		if fetched.Description == nil || *fetched.Description != "Updated description for message def" {
			t.Errorf("expected Description=Updated description for message def, got %v", fetched.Description)
		}
		if fetched.Category == nil || *fetched.Category != "consequence" {
			t.Errorf("expected Category=consequence, got %v", fetched.Category)
		}
	})

	t.Run("Search_ByStatus", func(t *testing.T) {
		var results []*conformance.MessageDefinition
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageDefinitionRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"status": "active"}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by status: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for status=active")
		}
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Search_ByEvent", func(t *testing.T) {
		// Create with a unique event code for searching
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageDefinitionRepoPG(globalDB.Pool)
			md := &conformance.MessageDefinition{
				Status:          "active",
				EventCodingCode: "search-event-unique",
			}
			return repo.Create(ctx, md)
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var results []*conformance.MessageDefinition
		var total int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageDefinitionRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"event": "search-event-unique"}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by event: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for event=search-event-unique")
		}
		for _, r := range results {
			if r.EventCodingCode != "search-event-unique" {
				t.Errorf("expected event_coding_code=search-event-unique, got %s", r.EventCodingCode)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var mdID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageDefinitionRepoPG(globalDB.Pool)
			md := &conformance.MessageDefinition{
				Status:          "active",
				EventCodingCode: "delete-event",
			}
			if err := repo.Create(ctx, md); err != nil {
				return err
			}
			mdID = md.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageDefinitionRepoPG(globalDB.Pool)
			return repo.Delete(ctx, mdID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageDefinitionRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, mdID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted message definition")
		}
	})
}

// ==================== MessageHeader Tests ====================

func TestMessageHeaderCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("conf")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	// Create prerequisite organization for sender FK
	orgID := createTestOrganization(t, ctx, globalDB.Pool, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *conformance.MessageHeader
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageHeaderRepoPG(globalDB.Pool)
			mh := &conformance.MessageHeader{
				EventCodingCode:     "admin-notify",
				EventCodingSystem:   ptrStr("http://example.org/fhir/message-events"),
				DestinationName:     ptrStr("Regional Lab System"),
				DestinationEndpoint: ptrStr("http://lab.example.org/fhir"),
				SenderOrgID:         &orgID,
				SourceName:          ptrStr("EHR System"),
				SourceEndpoint:      "http://ehr.example.org/fhir",
			}
			if err := repo.Create(ctx, mh); err != nil {
				return err
			}
			created = mh
			return nil
		})
		if err != nil {
			t.Fatalf("Create message header: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		fakeOrgID := uuid.New()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageHeaderRepoPG(globalDB.Pool)
			mh := &conformance.MessageHeader{
				EventCodingCode: "fk-test",
				SenderOrgID:     &fakeOrgID,
				SourceEndpoint:  "http://test.example.org/fhir",
			}
			return repo.Create(ctx, mh)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent organization")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var mhID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageHeaderRepoPG(globalDB.Pool)
			mh := &conformance.MessageHeader{
				EventCodingCode: "getbyid-event",
				SourceEndpoint:  "http://getbyid.example.org/fhir",
			}
			if err := repo.Create(ctx, mh); err != nil {
				return err
			}
			mhID = mh.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *conformance.MessageHeader
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageHeaderRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, mhID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.EventCodingCode != "getbyid-event" {
			t.Errorf("expected EventCodingCode=getbyid-event, got %s", fetched.EventCodingCode)
		}
		if fetched.SourceEndpoint != "http://getbyid.example.org/fhir" {
			t.Errorf("expected SourceEndpoint=http://getbyid.example.org/fhir, got %s", fetched.SourceEndpoint)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var mh *conformance.MessageHeader
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageHeaderRepoPG(globalDB.Pool)
			m := &conformance.MessageHeader{
				EventCodingCode: "fhirid-event",
				SourceEndpoint:  "http://fhirid.example.org/fhir",
			}
			if err := repo.Create(ctx, m); err != nil {
				return err
			}
			mh = m
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *conformance.MessageHeader
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageHeaderRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, mh.FHIRID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.ID != mh.ID {
			t.Errorf("expected ID=%s, got %s", mh.ID, fetched.ID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var mh *conformance.MessageHeader
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageHeaderRepoPG(globalDB.Pool)
			m := &conformance.MessageHeader{
				EventCodingCode:     "update-event",
				EventCodingSystem:   ptrStr("http://example.org/events"),
				DestinationName:     ptrStr("Original Dest"),
				DestinationEndpoint: ptrStr("http://original.example.org/fhir"),
				SourceEndpoint:      "http://source.example.org/fhir",
			}
			if err := repo.Create(ctx, m); err != nil {
				return err
			}
			mh = m
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageHeaderRepoPG(globalDB.Pool)
			mh.EventCodingCode = "updated-event"
			mh.DestinationName = ptrStr("Updated Dest")
			mh.DestinationEndpoint = ptrStr("http://updated.example.org/fhir")
			mh.SourceEndpoint = "http://updated-source.example.org/fhir"
			mh.ReasonCode = ptrStr("admin-request")
			return repo.Update(ctx, mh)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *conformance.MessageHeader
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageHeaderRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, mh.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.EventCodingCode != "updated-event" {
			t.Errorf("expected EventCodingCode=updated-event, got %s", fetched.EventCodingCode)
		}
		if fetched.DestinationName == nil || *fetched.DestinationName != "Updated Dest" {
			t.Errorf("expected DestinationName=Updated Dest, got %v", fetched.DestinationName)
		}
		if fetched.DestinationEndpoint == nil || *fetched.DestinationEndpoint != "http://updated.example.org/fhir" {
			t.Errorf("expected DestinationEndpoint=http://updated.example.org/fhir, got %v", fetched.DestinationEndpoint)
		}
		if fetched.SourceEndpoint != "http://updated-source.example.org/fhir" {
			t.Errorf("expected SourceEndpoint=http://updated-source.example.org/fhir, got %s", fetched.SourceEndpoint)
		}
		if fetched.ReasonCode == nil || *fetched.ReasonCode != "admin-request" {
			t.Errorf("expected ReasonCode=admin-request, got %v", fetched.ReasonCode)
		}
	})

	t.Run("Search_ByEvent", func(t *testing.T) {
		// Create with a unique event code for searching
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageHeaderRepoPG(globalDB.Pool)
			mh := &conformance.MessageHeader{
				EventCodingCode: "search-header-event",
				SourceEndpoint:  "http://search.example.org/fhir",
			}
			return repo.Create(ctx, mh)
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var results []*conformance.MessageHeader
		var total int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageHeaderRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"event": "search-header-event"}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by event: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for event=search-header-event")
		}
		for _, r := range results {
			if r.EventCodingCode != "search-header-event" {
				t.Errorf("expected event_coding_code=search-header-event, got %s", r.EventCodingCode)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var mhID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageHeaderRepoPG(globalDB.Pool)
			mh := &conformance.MessageHeader{
				EventCodingCode: "delete-event",
				SourceEndpoint:  "http://delete.example.org/fhir",
			}
			if err := repo.Create(ctx, mh); err != nil {
				return err
			}
			mhID = mh.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageHeaderRepoPG(globalDB.Pool)
			return repo.Delete(ctx, mhID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := conformance.NewMessageHeaderRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, mhID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted message header")
		}
	})
}
