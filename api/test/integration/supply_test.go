package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/supply"
	"github.com/google/uuid"
)

// ---- SupplyRequest Tests ----

func TestSupplyRequestCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("sup")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "SupReqDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *supply.SupplyRequest
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyRequestRepoPG(globalDB.Pool)
			sr := &supply.SupplyRequest{
				Status:        "active",
				CategoryCode:  ptrStr("central"),
				Priority:      ptrStr("routine"),
				ItemCode:      "46181005",
				ItemDisplay:   ptrStr("Surgical gloves"),
				ItemSystem:    ptrStr("http://snomed.info/sct"),
				QuantityValue: 100,
				QuantityUnit:  ptrStr("pairs"),
				AuthoredOn:    &now,
				RequesterID:   &practitioner.ID,
			}
			if err := repo.Create(ctx, sr); err != nil {
				return err
			}
			created = sr
			return nil
		})
		if err != nil {
			t.Fatalf("Create supply request: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		fakeID := uuid.New()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyRequestRepoPG(globalDB.Pool)
			sr := &supply.SupplyRequest{
				Status:        "active",
				ItemCode:      "12345",
				QuantityValue: 10,
				RequesterID:   &fakeID,
			}
			return repo.Create(ctx, sr)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent practitioner")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var srID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyRequestRepoPG(globalDB.Pool)
			sr := &supply.SupplyRequest{
				Status:        "active",
				CategoryCode:  ptrStr("central"),
				ItemCode:      "71388002",
				ItemDisplay:   ptrStr("Bandage"),
				QuantityValue: 50,
				QuantityUnit:  ptrStr("rolls"),
				RequesterID:   &practitioner.ID,
			}
			if err := repo.Create(ctx, sr); err != nil {
				return err
			}
			srID = sr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *supply.SupplyRequest
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyRequestRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, srID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.ItemCode != "71388002" {
			t.Errorf("expected item_code=71388002, got %s", fetched.ItemCode)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if fetched.QuantityValue != 50 {
			t.Errorf("expected quantity_value=50, got %f", fetched.QuantityValue)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyRequestRepoPG(globalDB.Pool)
			sr := &supply.SupplyRequest{
				Status:        "active",
				ItemCode:      "469008",
				ItemDisplay:   ptrStr("Syringe"),
				QuantityValue: 200,
				QuantityUnit:  ptrStr("units"),
			}
			if err := repo.Create(ctx, sr); err != nil {
				return err
			}
			fhirID = sr.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *supply.SupplyRequest
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyRequestRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, fhirID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.FHIRID != fhirID {
			t.Errorf("expected fhir_id=%s, got %s", fhirID, fetched.FHIRID)
		}
		if fetched.ItemCode != "469008" {
			t.Errorf("expected item_code=469008, got %s", fetched.ItemCode)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var sr *supply.SupplyRequest
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyRequestRepoPG(globalDB.Pool)
			sr = &supply.SupplyRequest{
				Status:        "active",
				CategoryCode:  ptrStr("central"),
				Priority:      ptrStr("routine"),
				ItemCode:      "87612001",
				ItemDisplay:   ptrStr("Gauze pad"),
				QuantityValue: 25,
				QuantityUnit:  ptrStr("boxes"),
				AuthoredOn:    &now,
				RequesterID:   &practitioner.ID,
			}
			return repo.Create(ctx, sr)
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyRequestRepoPG(globalDB.Pool)
			sr.Status = "completed"
			sr.Priority = ptrStr("urgent")
			sr.QuantityValue = 50
			return repo.Update(ctx, sr)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *supply.SupplyRequest
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyRequestRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, sr.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected status=completed, got %s", fetched.Status)
		}
		if fetched.Priority == nil || *fetched.Priority != "urgent" {
			t.Errorf("expected priority=urgent, got %v", fetched.Priority)
		}
		if fetched.QuantityValue != 50 {
			t.Errorf("expected quantity_value=50, got %f", fetched.QuantityValue)
		}
	})

	t.Run("Search_ByStatus", func(t *testing.T) {
		// Create a supply request with known status for search
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyRequestRepoPG(globalDB.Pool)
			sr := &supply.SupplyRequest{
				Status:        "active",
				CategoryCode:  ptrStr("central"),
				ItemCode:      "search-status-item",
				QuantityValue: 10,
			}
			return repo.Create(ctx, sr)
		})
		if err != nil {
			t.Fatalf("Create for search: %v", err)
		}

		var results []*supply.SupplyRequest
		var total int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyRequestRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"status": "active",
			}, 100, 0)
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

	t.Run("Search_ByCategory", func(t *testing.T) {
		var results []*supply.SupplyRequest
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyRequestRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"category": "central",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by category: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for category=central")
		}
		for _, r := range results {
			if r.CategoryCode == nil || *r.CategoryCode != "central" {
				t.Errorf("expected category_code=central, got %v", r.CategoryCode)
			}
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*supply.SupplyRequest
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyRequestRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 supply request")
		}
		if len(results) == 0 {
			t.Error("expected non-empty results")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var srID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyRequestRepoPG(globalDB.Pool)
			sr := &supply.SupplyRequest{
				Status:        "active",
				ItemCode:      "delete-test-item",
				QuantityValue: 1,
			}
			if err := repo.Create(ctx, sr); err != nil {
				return err
			}
			srID = sr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyRequestRepoPG(globalDB.Pool)
			return repo.Delete(ctx, srID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyRequestRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, srID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted supply request")
		}
	})
}

// ---- SupplyDelivery Tests ----

func TestSupplyDeliveryCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("sup")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "SupDelPatient", "Test", "MRN-SUPDEL-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "SupDelDoc", "Jones")

	t.Run("Create", func(t *testing.T) {
		var created *supply.SupplyDelivery
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			sd := &supply.SupplyDelivery{
				Status:              "completed",
				PatientID:           &patient.ID,
				TypeCode:            ptrStr("device"),
				TypeDisplay:         ptrStr("Device"),
				SuppliedItemCode:    ptrStr("46181005"),
				SuppliedItemDisplay: ptrStr("Surgical gloves"),
				SuppliedItemQuantity: ptrFloat(100),
				SuppliedItemUnit:    ptrStr("pairs"),
				OccurrenceDate:      &now,
				SupplierID:          &practitioner.ID,
			}
			if err := repo.Create(ctx, sd); err != nil {
				return err
			}
			created = sd
			return nil
		})
		if err != nil {
			t.Fatalf("Create supply delivery: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("Create_FK_Violation_Patient", func(t *testing.T) {
		fakePatient := uuid.New()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			sd := &supply.SupplyDelivery{
				Status:    "completed",
				PatientID: &fakePatient,
			}
			return repo.Create(ctx, sd)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("Create_FK_Violation_Supplier", func(t *testing.T) {
		fakeSupplier := uuid.New()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			sd := &supply.SupplyDelivery{
				Status:     "completed",
				PatientID:  &patient.ID,
				SupplierID: &fakeSupplier,
			}
			return repo.Create(ctx, sd)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent supplier")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var sdID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			sd := &supply.SupplyDelivery{
				Status:              "completed",
				PatientID:           &patient.ID,
				TypeCode:            ptrStr("medication"),
				SuppliedItemCode:    ptrStr("71388002"),
				SuppliedItemDisplay: ptrStr("Bandage"),
				SuppliedItemQuantity: ptrFloat(50),
				SuppliedItemUnit:    ptrStr("rolls"),
				OccurrenceDate:      &now,
				SupplierID:          &practitioner.ID,
			}
			if err := repo.Create(ctx, sd); err != nil {
				return err
			}
			sdID = sd.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *supply.SupplyDelivery
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, sdID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected status=completed, got %s", fetched.Status)
		}
		if fetched.SuppliedItemCode == nil || *fetched.SuppliedItemCode != "71388002" {
			t.Errorf("expected supplied_item_code=71388002, got %v", fetched.SuppliedItemCode)
		}
		if fetched.SuppliedItemQuantity == nil || *fetched.SuppliedItemQuantity != 50 {
			t.Errorf("expected supplied_item_quantity=50, got %v", fetched.SuppliedItemQuantity)
		}
		if fetched.PatientID == nil || *fetched.PatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %v", patient.ID, fetched.PatientID)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			sd := &supply.SupplyDelivery{
				Status:           "in-progress",
				PatientID:        &patient.ID,
				SuppliedItemCode: ptrStr("469008"),
				SuppliedItemDisplay: ptrStr("Syringe"),
				SuppliedItemQuantity: ptrFloat(200),
			}
			if err := repo.Create(ctx, sd); err != nil {
				return err
			}
			fhirID = sd.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *supply.SupplyDelivery
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, fhirID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.FHIRID != fhirID {
			t.Errorf("expected fhir_id=%s, got %s", fhirID, fetched.FHIRID)
		}
		if fetched.SuppliedItemCode == nil || *fetched.SuppliedItemCode != "469008" {
			t.Errorf("expected supplied_item_code=469008, got %v", fetched.SuppliedItemCode)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var sd *supply.SupplyDelivery
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			sd = &supply.SupplyDelivery{
				Status:              "in-progress",
				PatientID:           &patient.ID,
				TypeCode:            ptrStr("device"),
				SuppliedItemCode:    ptrStr("87612001"),
				SuppliedItemDisplay: ptrStr("Gauze pad"),
				SuppliedItemQuantity: ptrFloat(25),
				SuppliedItemUnit:    ptrStr("boxes"),
				OccurrenceDate:      &now,
				SupplierID:          &practitioner.ID,
			}
			return repo.Create(ctx, sd)
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			sd.Status = "completed"
			sd.SuppliedItemQuantity = ptrFloat(30)
			sd.TypeCode = ptrStr("medication")
			return repo.Update(ctx, sd)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *supply.SupplyDelivery
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, sd.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected status=completed, got %s", fetched.Status)
		}
		if fetched.SuppliedItemQuantity == nil || *fetched.SuppliedItemQuantity != 30 {
			t.Errorf("expected supplied_item_quantity=30, got %v", fetched.SuppliedItemQuantity)
		}
		if fetched.TypeCode == nil || *fetched.TypeCode != "medication" {
			t.Errorf("expected type_code=medication, got %v", fetched.TypeCode)
		}
	})

	t.Run("Search_ByStatus", func(t *testing.T) {
		// Create a delivery with known status
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			sd := &supply.SupplyDelivery{
				Status:           "completed",
				PatientID:        &patient.ID,
				SuppliedItemCode: ptrStr("search-status-del"),
			}
			return repo.Create(ctx, sd)
		})
		if err != nil {
			t.Fatalf("Create for search: %v", err)
		}

		var results []*supply.SupplyDelivery
		var total int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"status": "completed",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by status: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for status=completed")
		}
		for _, r := range results {
			if r.Status != "completed" {
				t.Errorf("expected status=completed, got %s", r.Status)
			}
		}
	})

	t.Run("Search_BySupplier", func(t *testing.T) {
		var results []*supply.SupplyDelivery
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"supplier": practitioner.ID.String(),
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by supplier: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for supplier")
		}
		for _, r := range results {
			if r.SupplierID == nil || *r.SupplierID != practitioner.ID {
				t.Errorf("expected supplier_id=%s, got %v", practitioner.ID, r.SupplierID)
			}
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*supply.SupplyDelivery
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 supply delivery")
		}
		if len(results) == 0 {
			t.Error("expected non-empty results")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var sdID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			sd := &supply.SupplyDelivery{
				Status:    "completed",
				PatientID: &patient.ID,
			}
			if err := repo.Create(ctx, sd); err != nil {
				return err
			}
			sdID = sd.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			return repo.Delete(ctx, sdID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, sdID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted supply delivery")
		}
	})

	t.Run("Create_WithBasedOn", func(t *testing.T) {
		// Create a SupplyRequest first, then link a SupplyDelivery to it
		var srID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyRequestRepoPG(globalDB.Pool)
			sr := &supply.SupplyRequest{
				Status:        "active",
				ItemCode:      "based-on-item",
				QuantityValue: 10,
			}
			if err := repo.Create(ctx, sr); err != nil {
				return err
			}
			srID = sr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create supply request: %v", err)
		}

		var created *supply.SupplyDelivery
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			sd := &supply.SupplyDelivery{
				Status:              "completed",
				BasedOnID:           &srID,
				PatientID:           &patient.ID,
				SuppliedItemCode:    ptrStr("based-on-item"),
				SuppliedItemQuantity: ptrFloat(10),
			}
			if err := repo.Create(ctx, sd); err != nil {
				return err
			}
			created = sd
			return nil
		})
		if err != nil {
			t.Fatalf("Create supply delivery with based_on: %v", err)
		}

		// Verify the based_on link
		var fetched *supply.SupplyDelivery
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := supply.NewSupplyDeliveryRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, created.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.BasedOnID == nil || *fetched.BasedOnID != srID {
			t.Errorf("expected based_on_id=%s, got %v", srID, fetched.BasedOnID)
		}
	})
}
