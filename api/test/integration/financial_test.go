package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/financial"
	"github.com/google/uuid"
)

// =========== Account Tests ===========

func TestAccountCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("fin")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "AcctPatient", "Test", "MRN-ACCT-001")
	orgID := createTestOrganization(t, ctx, globalDB.Pool, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *financial.Account
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewAccountRepoPG(globalDB.Pool)
			a := &financial.Account{
				Status:           "active",
				TypeCode:         ptrStr("PBILLACCT"),
				Name:             ptrStr("Patient Billing Account"),
				SubjectPatientID: &patient.ID,
				OwnerOrgID:       &orgID,
				Description:      ptrStr("Primary billing account"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			created = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create account: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var acctID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewAccountRepoPG(globalDB.Pool)
			a := &financial.Account{
				Status:           "active",
				TypeCode:         ptrStr("PBILLACCT"),
				Name:             ptrStr("Get Test Account"),
				SubjectPatientID: &patient.ID,
				Description:      ptrStr("For GetByID test"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			acctID = a.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.Account
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewAccountRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, acctID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if fetched.Name == nil || *fetched.Name != "Get Test Account" {
			t.Errorf("expected name=Get Test Account, got %v", fetched.Name)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewAccountRepoPG(globalDB.Pool)
			a := &financial.Account{
				Status: "active",
				Name:   ptrStr("FHIR Lookup Account"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			fhirID = a.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.Account
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewAccountRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, fhirID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.Name == nil || *fetched.Name != "FHIR Lookup Account" {
			t.Errorf("expected name=FHIR Lookup Account, got %v", fetched.Name)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var acct *financial.Account
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewAccountRepoPG(globalDB.Pool)
			a := &financial.Account{
				Status:      "active",
				Name:        ptrStr("Update Test Account"),
				Description: ptrStr("Before update"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			acct = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewAccountRepoPG(globalDB.Pool)
			acct.Status = "inactive"
			acct.Description = ptrStr("After update")
			return repo.Update(ctx, acct)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *financial.Account
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewAccountRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, acct.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "inactive" {
			t.Errorf("expected status=inactive, got %s", fetched.Status)
		}
		if fetched.Description == nil || *fetched.Description != "After update" {
			t.Errorf("expected description=After update, got %v", fetched.Description)
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*financial.Account
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewAccountRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"status": "active",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 account")
		}
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var acctID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewAccountRepoPG(globalDB.Pool)
			a := &financial.Account{
				Status: "active",
				Name:   ptrStr("Delete Test Account"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			acctID = a.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewAccountRepoPG(globalDB.Pool)
			return repo.Delete(ctx, acctID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewAccountRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, acctID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted account")
		}
	})
}

// =========== InsurancePlan Tests ===========

func TestInsurancePlanCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("fin")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	orgID := createTestOrganization(t, ctx, globalDB.Pool, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *financial.InsurancePlan
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewInsurancePlanRepoPG(globalDB.Pool)
			ip := &financial.InsurancePlan{
				Status:              "active",
				TypeCode:            ptrStr("medical"),
				Name:                ptrStr("Gold PPO Plan"),
				OwnedByOrgID:        &orgID,
				AdministeredByOrgID: &orgID,
			}
			if err := repo.Create(ctx, ip); err != nil {
				return err
			}
			created = ip
			return nil
		})
		if err != nil {
			t.Fatalf("Create insurance plan: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var ipID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewInsurancePlanRepoPG(globalDB.Pool)
			ip := &financial.InsurancePlan{
				Status:   "active",
				TypeCode: ptrStr("dental"),
				Name:     ptrStr("Dental Basic Plan"),
			}
			if err := repo.Create(ctx, ip); err != nil {
				return err
			}
			ipID = ip.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.InsurancePlan
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewInsurancePlanRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, ipID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Name == nil || *fetched.Name != "Dental Basic Plan" {
			t.Errorf("expected name=Dental Basic Plan, got %v", fetched.Name)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewInsurancePlanRepoPG(globalDB.Pool)
			ip := &financial.InsurancePlan{
				Status: "active",
				Name:   ptrStr("FHIR Lookup Plan"),
			}
			if err := repo.Create(ctx, ip); err != nil {
				return err
			}
			fhirID = ip.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.InsurancePlan
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewInsurancePlanRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, fhirID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.Name == nil || *fetched.Name != "FHIR Lookup Plan" {
			t.Errorf("expected name=FHIR Lookup Plan, got %v", fetched.Name)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var ip *financial.InsurancePlan
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewInsurancePlanRepoPG(globalDB.Pool)
			p := &financial.InsurancePlan{
				Status: "active",
				Name:   ptrStr("Update Test Plan"),
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			ip = p
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewInsurancePlanRepoPG(globalDB.Pool)
			ip.Status = "retired"
			ip.Name = ptrStr("Updated Plan Name")
			return repo.Update(ctx, ip)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *financial.InsurancePlan
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewInsurancePlanRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, ip.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "retired" {
			t.Errorf("expected status=retired, got %s", fetched.Status)
		}
		if fetched.Name == nil || *fetched.Name != "Updated Plan Name" {
			t.Errorf("expected name=Updated Plan Name, got %v", fetched.Name)
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*financial.InsurancePlan
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewInsurancePlanRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"status": "active",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 insurance plan")
		}
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var ipID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewInsurancePlanRepoPG(globalDB.Pool)
			ip := &financial.InsurancePlan{
				Status: "active",
				Name:   ptrStr("Delete Test Plan"),
			}
			if err := repo.Create(ctx, ip); err != nil {
				return err
			}
			ipID = ip.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewInsurancePlanRepoPG(globalDB.Pool)
			return repo.Delete(ctx, ipID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewInsurancePlanRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, ipID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted insurance plan")
		}
	})
}

// =========== PaymentNotice Tests ===========

func TestPaymentNoticeCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("fin")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *financial.PaymentNotice
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentNoticeRepoPG(globalDB.Pool)
			pn := &financial.PaymentNotice{
				Status:            "active",
				Created:           now,
				PaymentDate:       ptrTime(now),
				AmountValue:       ptrFloat(250.00),
				AmountCurrency:    ptrStr("USD"),
				PaymentStatusCode: ptrStr("paid"),
			}
			if err := repo.Create(ctx, pn); err != nil {
				return err
			}
			created = pn
			return nil
		})
		if err != nil {
			t.Fatalf("Create payment notice: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var pnID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentNoticeRepoPG(globalDB.Pool)
			pn := &financial.PaymentNotice{
				Status:            "active",
				Created:           now,
				AmountValue:       ptrFloat(100.00),
				AmountCurrency:    ptrStr("USD"),
				PaymentStatusCode: ptrStr("cleared"),
			}
			if err := repo.Create(ctx, pn); err != nil {
				return err
			}
			pnID = pn.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.PaymentNotice
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentNoticeRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, pnID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if fetched.AmountValue == nil || *fetched.AmountValue != 100.00 {
			t.Errorf("expected amount=100.00, got %v", fetched.AmountValue)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		now := time.Now()
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentNoticeRepoPG(globalDB.Pool)
			pn := &financial.PaymentNotice{
				Status:  "active",
				Created: now,
			}
			if err := repo.Create(ctx, pn); err != nil {
				return err
			}
			fhirID = pn.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.PaymentNotice
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentNoticeRepoPG(globalDB.Pool)
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
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var pn *financial.PaymentNotice
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentNoticeRepoPG(globalDB.Pool)
			p := &financial.PaymentNotice{
				Status:            "active",
				Created:           now,
				AmountValue:       ptrFloat(200.00),
				AmountCurrency:    ptrStr("USD"),
				PaymentStatusCode: ptrStr("pending"),
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			pn = p
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentNoticeRepoPG(globalDB.Pool)
			pn.Status = "cancelled"
			pn.PaymentStatusCode = ptrStr("declined")
			return repo.Update(ctx, pn)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *financial.PaymentNotice
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentNoticeRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, pn.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "cancelled" {
			t.Errorf("expected status=cancelled, got %s", fetched.Status)
		}
		if fetched.PaymentStatusCode == nil || *fetched.PaymentStatusCode != "declined" {
			t.Errorf("expected payment_status=declined, got %v", fetched.PaymentStatusCode)
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*financial.PaymentNotice
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentNoticeRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"status": "active",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 payment notice")
		}
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var pnID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentNoticeRepoPG(globalDB.Pool)
			pn := &financial.PaymentNotice{
				Status:  "active",
				Created: now,
			}
			if err := repo.Create(ctx, pn); err != nil {
				return err
			}
			pnID = pn.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentNoticeRepoPG(globalDB.Pool)
			return repo.Delete(ctx, pnID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentNoticeRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, pnID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted payment notice")
		}
	})
}

// =========== PaymentReconciliation Tests ===========

func TestPaymentReconciliationCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("fin")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *financial.PaymentReconciliation
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentReconciliationRepoPG(globalDB.Pool)
			pr := &financial.PaymentReconciliation{
				Status:          "active",
				Created:         now,
				PaymentDate:     now,
				PaymentAmount:   1500.00,
				PaymentCurrency: ptrStr("USD"),
				Outcome:         ptrStr("complete"),
				Disposition:     ptrStr("Payment processed successfully"),
			}
			if err := repo.Create(ctx, pr); err != nil {
				return err
			}
			created = pr
			return nil
		})
		if err != nil {
			t.Fatalf("Create payment reconciliation: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var prID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentReconciliationRepoPG(globalDB.Pool)
			pr := &financial.PaymentReconciliation{
				Status:          "active",
				Created:         now,
				PaymentDate:     now,
				PaymentAmount:   750.50,
				PaymentCurrency: ptrStr("USD"),
				Outcome:         ptrStr("complete"),
			}
			if err := repo.Create(ctx, pr); err != nil {
				return err
			}
			prID = pr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.PaymentReconciliation
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentReconciliationRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, prID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.PaymentAmount != 750.50 {
			t.Errorf("expected amount=750.50, got %f", fetched.PaymentAmount)
		}
		if fetched.Outcome == nil || *fetched.Outcome != "complete" {
			t.Errorf("expected outcome=complete, got %v", fetched.Outcome)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		now := time.Now()
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentReconciliationRepoPG(globalDB.Pool)
			pr := &financial.PaymentReconciliation{
				Status:        "active",
				Created:       now,
				PaymentDate:   now,
				PaymentAmount: 500.00,
			}
			if err := repo.Create(ctx, pr); err != nil {
				return err
			}
			fhirID = pr.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.PaymentReconciliation
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentReconciliationRepoPG(globalDB.Pool)
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
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var pr *financial.PaymentReconciliation
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentReconciliationRepoPG(globalDB.Pool)
			p := &financial.PaymentReconciliation{
				Status:        "active",
				Created:       now,
				PaymentDate:   now,
				PaymentAmount: 300.00,
				Outcome:       ptrStr("queued"),
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			pr = p
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentReconciliationRepoPG(globalDB.Pool)
			pr.Outcome = ptrStr("complete")
			pr.Disposition = ptrStr("Reconciliation finalized")
			pr.PaymentAmount = 350.00
			return repo.Update(ctx, pr)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *financial.PaymentReconciliation
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentReconciliationRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, pr.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Outcome == nil || *fetched.Outcome != "complete" {
			t.Errorf("expected outcome=complete, got %v", fetched.Outcome)
		}
		if fetched.Disposition == nil || *fetched.Disposition != "Reconciliation finalized" {
			t.Errorf("expected disposition=Reconciliation finalized, got %v", fetched.Disposition)
		}
		if fetched.PaymentAmount != 350.00 {
			t.Errorf("expected amount=350.00, got %f", fetched.PaymentAmount)
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*financial.PaymentReconciliation
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentReconciliationRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"status": "active",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 payment reconciliation")
		}
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var prID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentReconciliationRepoPG(globalDB.Pool)
			pr := &financial.PaymentReconciliation{
				Status:        "active",
				Created:       now,
				PaymentDate:   now,
				PaymentAmount: 100.00,
			}
			if err := repo.Create(ctx, pr); err != nil {
				return err
			}
			prID = pr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentReconciliationRepoPG(globalDB.Pool)
			return repo.Delete(ctx, prID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewPaymentReconciliationRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, prID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted payment reconciliation")
		}
	})
}

// =========== ChargeItem Tests ===========

func TestChargeItemCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("fin")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "CIPatient", "Test", "MRN-CI-001")

	t.Run("Create", func(t *testing.T) {
		var created *financial.ChargeItem
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemRepoPG(globalDB.Pool)
			ci := &financial.ChargeItem{
				Status:           "billable",
				CodeCode:         ptrStr("99213"),
				CodeDisplay:      ptrStr("Office visit, established patient"),
				SubjectPatientID: patient.ID,
				OccurrenceDate:   ptrTime(now),
				QuantityValue:    ptrFloat(1),
				PriceOverrideValue: ptrFloat(150.00),
			}
			if err := repo.Create(ctx, ci); err != nil {
				return err
			}
			created = ci
			return nil
		})
		if err != nil {
			t.Fatalf("Create charge item: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemRepoPG(globalDB.Pool)
			ci := &financial.ChargeItem{
				Status:           "billable",
				SubjectPatientID: uuid.New(), // non-existent
				CodeCode:         ptrStr("99999"),
			}
			return repo.Create(ctx, ci)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var ciID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemRepoPG(globalDB.Pool)
			ci := &financial.ChargeItem{
				Status:           "billable",
				CodeCode:         ptrStr("99214"),
				CodeDisplay:      ptrStr("Office visit, detailed"),
				SubjectPatientID: patient.ID,
				OccurrenceDate:   ptrTime(now),
			}
			if err := repo.Create(ctx, ci); err != nil {
				return err
			}
			ciID = ci.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.ChargeItem
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, ciID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.CodeCode == nil || *fetched.CodeCode != "99214" {
			t.Errorf("expected code=99214, got %v", fetched.CodeCode)
		}
		if fetched.SubjectPatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %s", patient.ID, fetched.SubjectPatientID)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		now := time.Now()
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemRepoPG(globalDB.Pool)
			ci := &financial.ChargeItem{
				Status:           "billable",
				CodeCode:         ptrStr("99215"),
				SubjectPatientID: patient.ID,
				OccurrenceDate:   ptrTime(now),
			}
			if err := repo.Create(ctx, ci); err != nil {
				return err
			}
			fhirID = ci.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.ChargeItem
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, fhirID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.CodeCode == nil || *fetched.CodeCode != "99215" {
			t.Errorf("expected code=99215, got %v", fetched.CodeCode)
		}
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var ci *financial.ChargeItem
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemRepoPG(globalDB.Pool)
			c := &financial.ChargeItem{
				Status:             "planned",
				CodeCode:           ptrStr("99211"),
				CodeDisplay:        ptrStr("Minimal visit"),
				SubjectPatientID:   patient.ID,
				OccurrenceDate:     ptrTime(now),
				PriceOverrideValue: ptrFloat(50.00),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			ci = c
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemRepoPG(globalDB.Pool)
			ci.Status = "billable"
			ci.PriceOverrideValue = ptrFloat(75.00)
			ci.Note = ptrStr("Updated charge amount")
			return repo.Update(ctx, ci)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *financial.ChargeItem
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, ci.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "billable" {
			t.Errorf("expected status=billable, got %s", fetched.Status)
		}
		if fetched.PriceOverrideValue == nil || *fetched.PriceOverrideValue != 75.00 {
			t.Errorf("expected price=75.00, got %v", fetched.PriceOverrideValue)
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*financial.ChargeItem
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "billable",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 charge item")
		}
		for _, r := range results {
			if r.SubjectPatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.SubjectPatientID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var ciID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemRepoPG(globalDB.Pool)
			ci := &financial.ChargeItem{
				Status:           "billable",
				SubjectPatientID: patient.ID,
				OccurrenceDate:   ptrTime(now),
			}
			if err := repo.Create(ctx, ci); err != nil {
				return err
			}
			ciID = ci.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemRepoPG(globalDB.Pool)
			return repo.Delete(ctx, ciID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, ciID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted charge item")
		}
	})
}

// =========== ChargeItemDefinition Tests ===========

func TestChargeItemDefinitionCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("fin")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *financial.ChargeItemDefinition
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemDefinitionRepoPG(globalDB.Pool)
			cd := &financial.ChargeItemDefinition{
				Status:      "active",
				URL:         ptrStr("http://example.org/fhir/ChargeItemDefinition/office-visit"),
				Title:       ptrStr("Office Visit Charge Definition"),
				CodeCode:    ptrStr("office-visit"),
				CodeDisplay: ptrStr("Office Visit"),
				Publisher:   ptrStr("Example Health System"),
			}
			if err := repo.Create(ctx, cd); err != nil {
				return err
			}
			created = cd
			return nil
		})
		if err != nil {
			t.Fatalf("Create charge item definition: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var cdID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemDefinitionRepoPG(globalDB.Pool)
			cd := &financial.ChargeItemDefinition{
				Status:      "active",
				Title:       ptrStr("Lab Test Charge"),
				CodeCode:    ptrStr("lab-basic"),
				CodeDisplay: ptrStr("Basic Lab Panel"),
			}
			if err := repo.Create(ctx, cd); err != nil {
				return err
			}
			cdID = cd.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.ChargeItemDefinition
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemDefinitionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, cdID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Title == nil || *fetched.Title != "Lab Test Charge" {
			t.Errorf("expected title=Lab Test Charge, got %v", fetched.Title)
		}
		if fetched.CodeCode == nil || *fetched.CodeCode != "lab-basic" {
			t.Errorf("expected code=lab-basic, got %v", fetched.CodeCode)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemDefinitionRepoPG(globalDB.Pool)
			cd := &financial.ChargeItemDefinition{
				Status: "active",
				Title:  ptrStr("FHIR Lookup Charge Def"),
			}
			if err := repo.Create(ctx, cd); err != nil {
				return err
			}
			fhirID = cd.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.ChargeItemDefinition
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemDefinitionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, fhirID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.Title == nil || *fetched.Title != "FHIR Lookup Charge Def" {
			t.Errorf("expected title=FHIR Lookup Charge Def, got %v", fetched.Title)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var cd *financial.ChargeItemDefinition
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemDefinitionRepoPG(globalDB.Pool)
			c := &financial.ChargeItemDefinition{
				Status:    "draft",
				Title:     ptrStr("Draft Charge Def"),
				Publisher: ptrStr("Draft Publisher"),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			cd = c
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemDefinitionRepoPG(globalDB.Pool)
			cd.Status = "active"
			cd.Title = ptrStr("Active Charge Def")
			cd.Publisher = ptrStr("Final Publisher")
			return repo.Update(ctx, cd)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *financial.ChargeItemDefinition
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemDefinitionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, cd.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if fetched.Title == nil || *fetched.Title != "Active Charge Def" {
			t.Errorf("expected title=Active Charge Def, got %v", fetched.Title)
		}
		if fetched.Publisher == nil || *fetched.Publisher != "Final Publisher" {
			t.Errorf("expected publisher=Final Publisher, got %v", fetched.Publisher)
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*financial.ChargeItemDefinition
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemDefinitionRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"status": "active",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 charge item definition")
		}
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var cdID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemDefinitionRepoPG(globalDB.Pool)
			cd := &financial.ChargeItemDefinition{
				Status: "active",
				Title:  ptrStr("Delete Test Charge Def"),
			}
			if err := repo.Create(ctx, cd); err != nil {
				return err
			}
			cdID = cd.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemDefinitionRepoPG(globalDB.Pool)
			return repo.Delete(ctx, cdID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewChargeItemDefinitionRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, cdID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted charge item definition")
		}
	})
}

// =========== Contract Tests ===========

func TestContractCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("fin")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "ContractPatient", "Test", "MRN-CT-001")

	t.Run("Create", func(t *testing.T) {
		var created *financial.Contract
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewContractRepoPG(globalDB.Pool)
			ct := &financial.Contract{
				Status:           "executed",
				TypeCode:         ptrStr("consent"),
				Title:            ptrStr("Patient Consent Agreement"),
				Issued:           ptrTime(now),
				SubjectPatientID: &patient.ID,
			}
			if err := repo.Create(ctx, ct); err != nil {
				return err
			}
			created = ct
			return nil
		})
		if err != nil {
			t.Fatalf("Create contract: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var ctID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewContractRepoPG(globalDB.Pool)
			ct := &financial.Contract{
				Status:           "executed",
				TypeCode:         ptrStr("privacy"),
				Title:            ptrStr("Privacy Policy Agreement"),
				Issued:           ptrTime(now),
				SubjectPatientID: &patient.ID,
			}
			if err := repo.Create(ctx, ct); err != nil {
				return err
			}
			ctID = ct.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.Contract
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewContractRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, ctID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Title == nil || *fetched.Title != "Privacy Policy Agreement" {
			t.Errorf("expected title=Privacy Policy Agreement, got %v", fetched.Title)
		}
		if fetched.TypeCode == nil || *fetched.TypeCode != "privacy" {
			t.Errorf("expected type=privacy, got %v", fetched.TypeCode)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewContractRepoPG(globalDB.Pool)
			ct := &financial.Contract{
				Status: "executed",
				Title:  ptrStr("FHIR Lookup Contract"),
			}
			if err := repo.Create(ctx, ct); err != nil {
				return err
			}
			fhirID = ct.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.Contract
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewContractRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, fhirID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.Title == nil || *fetched.Title != "FHIR Lookup Contract" {
			t.Errorf("expected title=FHIR Lookup Contract, got %v", fetched.Title)
		}
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var ct *financial.Contract
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewContractRepoPG(globalDB.Pool)
			c := &financial.Contract{
				Status: "offered",
				Title:  ptrStr("Pending Contract"),
				Issued: ptrTime(now),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			ct = c
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewContractRepoPG(globalDB.Pool)
			ct.Status = "executed"
			ct.Title = ptrStr("Executed Contract")
			ct.SubjectPatientID = &patient.ID
			return repo.Update(ctx, ct)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *financial.Contract
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewContractRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, ct.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "executed" {
			t.Errorf("expected status=executed, got %s", fetched.Status)
		}
		if fetched.Title == nil || *fetched.Title != "Executed Contract" {
			t.Errorf("expected title=Executed Contract, got %v", fetched.Title)
		}
		if fetched.SubjectPatientID == nil || *fetched.SubjectPatientID != patient.ID {
			t.Errorf("expected subject_patient_id=%s, got %v", patient.ID, fetched.SubjectPatientID)
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*financial.Contract
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewContractRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"status":  "executed",
				"patient": patient.ID.String(),
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 contract")
		}
		for _, r := range results {
			if r.Status != "executed" {
				t.Errorf("expected status=executed, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var ctID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewContractRepoPG(globalDB.Pool)
			ct := &financial.Contract{
				Status: "executed",
				Title:  ptrStr("Delete Test Contract"),
			}
			if err := repo.Create(ctx, ct); err != nil {
				return err
			}
			ctID = ct.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewContractRepoPG(globalDB.Pool)
			return repo.Delete(ctx, ctID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewContractRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, ctID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted contract")
		}
	})
}

// =========== EnrollmentRequest Tests ===========

func TestEnrollmentRequestCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("fin")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "EnrollReqPatient", "Test", "MRN-ER-001")

	t.Run("Create", func(t *testing.T) {
		var created *financial.EnrollmentRequest
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentRequestRepoPG(globalDB.Pool)
			er := &financial.EnrollmentRequest{
				Status:             "active",
				Created:            now,
				CandidatePatientID: &patient.ID,
			}
			if err := repo.Create(ctx, er); err != nil {
				return err
			}
			created = er
			return nil
		})
		if err != nil {
			t.Fatalf("Create enrollment request: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var erID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentRequestRepoPG(globalDB.Pool)
			er := &financial.EnrollmentRequest{
				Status:             "active",
				Created:            now,
				CandidatePatientID: &patient.ID,
			}
			if err := repo.Create(ctx, er); err != nil {
				return err
			}
			erID = er.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.EnrollmentRequest
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentRequestRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, erID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if fetched.CandidatePatientID == nil || *fetched.CandidatePatientID != patient.ID {
			t.Errorf("expected candidate_patient_id=%s, got %v", patient.ID, fetched.CandidatePatientID)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		now := time.Now()
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentRequestRepoPG(globalDB.Pool)
			er := &financial.EnrollmentRequest{
				Status:  "active",
				Created: now,
			}
			if err := repo.Create(ctx, er); err != nil {
				return err
			}
			fhirID = er.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.EnrollmentRequest
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentRequestRepoPG(globalDB.Pool)
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
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var er *financial.EnrollmentRequest
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentRequestRepoPG(globalDB.Pool)
			e := &financial.EnrollmentRequest{
				Status:  "active",
				Created: now,
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			er = e
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentRequestRepoPG(globalDB.Pool)
			er.Status = "cancelled"
			er.CandidatePatientID = &patient.ID
			return repo.Update(ctx, er)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *financial.EnrollmentRequest
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentRequestRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, er.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "cancelled" {
			t.Errorf("expected status=cancelled, got %s", fetched.Status)
		}
		if fetched.CandidatePatientID == nil || *fetched.CandidatePatientID != patient.ID {
			t.Errorf("expected candidate_patient_id=%s, got %v", patient.ID, fetched.CandidatePatientID)
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*financial.EnrollmentRequest
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentRequestRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"status": "active",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 enrollment request")
		}
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var erID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentRequestRepoPG(globalDB.Pool)
			er := &financial.EnrollmentRequest{
				Status:  "active",
				Created: now,
			}
			if err := repo.Create(ctx, er); err != nil {
				return err
			}
			erID = er.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentRequestRepoPG(globalDB.Pool)
			return repo.Delete(ctx, erID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentRequestRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, erID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted enrollment request")
		}
	})
}

// =========== EnrollmentResponse Tests ===========

func TestEnrollmentResponseCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("fin")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *financial.EnrollmentResponse
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentResponseRepoPG(globalDB.Pool)
			er := &financial.EnrollmentResponse{
				Status:      "active",
				Outcome:     ptrStr("complete"),
				Disposition: ptrStr("Enrollment approved"),
				Created:     now,
			}
			if err := repo.Create(ctx, er); err != nil {
				return err
			}
			created = er
			return nil
		})
		if err != nil {
			t.Fatalf("Create enrollment response: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var erID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentResponseRepoPG(globalDB.Pool)
			er := &financial.EnrollmentResponse{
				Status:      "active",
				Outcome:     ptrStr("complete"),
				Disposition: ptrStr("Enrollment confirmed"),
				Created:     now,
			}
			if err := repo.Create(ctx, er); err != nil {
				return err
			}
			erID = er.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.EnrollmentResponse
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentResponseRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, erID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Outcome == nil || *fetched.Outcome != "complete" {
			t.Errorf("expected outcome=complete, got %v", fetched.Outcome)
		}
		if fetched.Disposition == nil || *fetched.Disposition != "Enrollment confirmed" {
			t.Errorf("expected disposition=Enrollment confirmed, got %v", fetched.Disposition)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		now := time.Now()
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentResponseRepoPG(globalDB.Pool)
			er := &financial.EnrollmentResponse{
				Status:  "active",
				Created: now,
			}
			if err := repo.Create(ctx, er); err != nil {
				return err
			}
			fhirID = er.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *financial.EnrollmentResponse
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentResponseRepoPG(globalDB.Pool)
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
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var er *financial.EnrollmentResponse
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentResponseRepoPG(globalDB.Pool)
			e := &financial.EnrollmentResponse{
				Status:  "active",
				Outcome: ptrStr("queued"),
				Created: now,
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			er = e
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentResponseRepoPG(globalDB.Pool)
			er.Outcome = ptrStr("complete")
			er.Disposition = ptrStr("Enrollment finalized")
			return repo.Update(ctx, er)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *financial.EnrollmentResponse
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentResponseRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, er.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Outcome == nil || *fetched.Outcome != "complete" {
			t.Errorf("expected outcome=complete, got %v", fetched.Outcome)
		}
		if fetched.Disposition == nil || *fetched.Disposition != "Enrollment finalized" {
			t.Errorf("expected disposition=Enrollment finalized, got %v", fetched.Disposition)
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*financial.EnrollmentResponse
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentResponseRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"status": "active",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 enrollment response")
		}
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var erID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentResponseRepoPG(globalDB.Pool)
			er := &financial.EnrollmentResponse{
				Status:  "active",
				Created: now,
			}
			if err := repo.Create(ctx, er); err != nil {
				return err
			}
			erID = er.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentResponseRepoPG(globalDB.Pool)
			return repo.Delete(ctx, erID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := financial.NewEnrollmentResponseRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, erID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted enrollment response")
		}
	})
}
