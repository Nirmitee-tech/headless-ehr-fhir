package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/billing"
	"github.com/google/uuid"
)

func TestCoverageCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("cov")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "CovPatient", "Test", "MRN-COV-001")

	t.Run("Create", func(t *testing.T) {
		var created *billing.Coverage
		now := time.Now()
		endDate := now.Add(365 * 24 * time.Hour)
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewCoverageRepoPG(globalDB.Pool)
			c := &billing.Coverage{
				Status:           "active",
				TypeCode:         ptrStr("EHCPOL"),
				PatientID:        patient.ID,
				SubscriberID:     ptrStr("SUB-12345"),
				SubscriberName:   ptrStr("John Test"),
				Relationship:     ptrStr("self"),
				PayorName:        ptrStr("Blue Cross Blue Shield"),
				PolicyNumber:     ptrStr("POL-99887766"),
				GroupNumber:      ptrStr("GRP-5544"),
				GroupName:        ptrStr("Employer Group Plan"),
				PlanName:         ptrStr("Gold PPO"),
				PlanType:         ptrStr("PPO"),
				MemberID:         ptrStr("MEM-001"),
				PeriodStart:      &now,
				PeriodEnd:        &endDate,
				Network:          ptrStr("in-network"),
				CopayAmount:      ptrFloat(25.00),
				DeductibleAmount: ptrFloat(1500.00),
				DeductibleMet:    ptrFloat(500.00),
				MaxBenefitAmount: ptrFloat(1000000.00),
				OutOfPocketMax:   ptrFloat(5000.00),
				Currency:         ptrStr("USD"),
				CoverageOrder:    ptrInt(1),
				Note:             ptrStr("Primary insurance coverage"),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			created = c
			return nil
		})
		if err != nil {
			t.Fatalf("Create coverage: %v", err)
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
			repo := billing.NewCoverageRepoPG(globalDB.Pool)
			c := &billing.Coverage{
				Status:    "active",
				PatientID: uuid.New(), // non-existent
			}
			return repo.Create(ctx, c)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var covID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewCoverageRepoPG(globalDB.Pool)
			c := &billing.Coverage{
				Status:       "active",
				TypeCode:     ptrStr("HIP"),
				PatientID:    patient.ID,
				PolicyNumber: ptrStr("HIP-001"),
				PayorName:    ptrStr("Aetna"),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			covID = c.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *billing.Coverage
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewCoverageRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, covID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.PolicyNumber == nil || *fetched.PolicyNumber != "HIP-001" {
			t.Errorf("expected policy_number=HIP-001, got %v", fetched.PolicyNumber)
		}
		if fetched.PayorName == nil || *fetched.PayorName != "Aetna" {
			t.Errorf("expected payor_name=Aetna, got %v", fetched.PayorName)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewCoverageRepoPG(globalDB.Pool)
			c := &billing.Coverage{
				Status:    "active",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			fhirID = c.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *billing.Coverage
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewCoverageRepoPG(globalDB.Pool)
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
		var cov *billing.Coverage
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewCoverageRepoPG(globalDB.Pool)
			c := &billing.Coverage{
				Status:       "active",
				PatientID:    patient.ID,
				PolicyNumber: ptrStr("UPD-001"),
				PayorName:    ptrStr("UnitedHealthcare"),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			cov = c
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewCoverageRepoPG(globalDB.Pool)
			cov.Status = "cancelled"
			cov.Note = ptrStr("Coverage terminated by employer")
			cov.DeductibleAmount = ptrFloat(2000.00)
			return repo.Update(ctx, cov)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *billing.Coverage
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewCoverageRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, cov.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "cancelled" {
			t.Errorf("expected status=cancelled, got %s", fetched.Status)
		}
		if fetched.Note == nil || *fetched.Note != "Coverage terminated by employer" {
			t.Errorf("expected note set, got %v", fetched.Note)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*billing.Coverage
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewCoverageRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 coverage")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search_ByStatus", func(t *testing.T) {
		var results []*billing.Coverage
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewCoverageRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "active",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by status: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var covID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewCoverageRepoPG(globalDB.Pool)
			c := &billing.Coverage{
				Status:    "active",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			covID = c.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewCoverageRepoPG(globalDB.Pool)
			return repo.Delete(ctx, covID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewCoverageRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, covID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted coverage")
		}
	})
}

func TestClaimCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("claim")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "ClaimPatient", "Test", "MRN-CLAIM-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "ClaimDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *billing.Claim
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			c := &billing.Claim{
				Status:              "active",
				TypeCode:            ptrStr("professional"),
				UseCode:             ptrStr("claim"),
				PatientID:           patient.ID,
				ProviderID:          &practitioner.ID,
				PriorityCode:        ptrStr("normal"),
				BillablePeriodStart: &now,
				TotalAmount:         ptrFloat(250.00),
				Currency:            ptrStr("USD"),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			created = c
			return nil
		})
		if err != nil {
			t.Fatalf("Create claim: %v", err)
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
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			c := &billing.Claim{
				Status:    "active",
				PatientID: uuid.New(), // non-existent
			}
			return repo.Create(ctx, c)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var claimID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			c := &billing.Claim{
				Status:      "active",
				TypeCode:    ptrStr("institutional"),
				UseCode:     ptrStr("claim"),
				PatientID:   patient.ID,
				TotalAmount: ptrFloat(1200.00),
				Currency:    ptrStr("USD"),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			claimID = c.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *billing.Claim
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, claimID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.TypeCode == nil || *fetched.TypeCode != "institutional" {
			t.Errorf("expected type_code=institutional, got %v", fetched.TypeCode)
		}
		if fetched.TotalAmount == nil || *fetched.TotalAmount != 1200.00 {
			t.Errorf("expected total_amount=1200.00, got %v", fetched.TotalAmount)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			c := &billing.Claim{
				Status:    "active",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			fhirID = c.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *billing.Claim
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
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
		var claim *billing.Claim
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			c := &billing.Claim{
				Status:      "draft",
				TypeCode:    ptrStr("professional"),
				UseCode:     ptrStr("claim"),
				PatientID:   patient.ID,
				TotalAmount: ptrFloat(300.00),
				Currency:    ptrStr("USD"),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			claim = c
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			claim.Status = "active"
			claim.TotalAmount = ptrFloat(350.00)
			return repo.Update(ctx, claim)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *billing.Claim
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, claim.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if fetched.TotalAmount == nil || *fetched.TotalAmount != 350.00 {
			t.Errorf("expected total_amount=350.00, got %v", fetched.TotalAmount)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*billing.Claim
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 claim")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search_ByUse", func(t *testing.T) {
		var results []*billing.Claim
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"use":     "claim",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by use: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.UseCode == nil || *r.UseCode != "claim" {
				t.Errorf("expected use_code=claim, got %v", r.UseCode)
			}
		}
	})

	t.Run("Diagnoses", func(t *testing.T) {
		// Create a claim
		var claimID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			c := &billing.Claim{
				Status:    "active",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			claimID = c.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create claim: %v", err)
		}

		// Add diagnosis
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			d := &billing.ClaimDiagnosis{
				ClaimID:             claimID,
				Sequence:            1,
				DiagnosisCodeSystem: ptrStr("http://hl7.org/fhir/sid/icd-10-cm"),
				DiagnosisCode:       "J06.9",
				DiagnosisDisplay:    ptrStr("Acute upper respiratory infection"),
				TypeCode:            ptrStr("principal"),
				OnAdmission:         ptrBool(true),
			}
			return repo.AddDiagnosis(ctx, d)
		})
		if err != nil {
			t.Fatalf("AddDiagnosis: %v", err)
		}

		// Add second diagnosis
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			d := &billing.ClaimDiagnosis{
				ClaimID:          claimID,
				Sequence:         2,
				DiagnosisCode:    "I10",
				DiagnosisDisplay: ptrStr("Essential hypertension"),
				TypeCode:         ptrStr("secondary"),
			}
			return repo.AddDiagnosis(ctx, d)
		})
		if err != nil {
			t.Fatalf("AddDiagnosis 2: %v", err)
		}

		// Get diagnoses
		var diags []*billing.ClaimDiagnosis
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			var err error
			diags, err = repo.GetDiagnoses(ctx, claimID)
			return err
		})
		if err != nil {
			t.Fatalf("GetDiagnoses: %v", err)
		}
		if len(diags) != 2 {
			t.Fatalf("expected 2 diagnoses, got %d", len(diags))
		}
		if diags[0].DiagnosisCode != "J06.9" {
			t.Errorf("expected first diagnosis=J06.9, got %s", diags[0].DiagnosisCode)
		}
		if diags[0].Sequence != 1 {
			t.Errorf("expected sequence=1, got %d", diags[0].Sequence)
		}
		if diags[1].DiagnosisCode != "I10" {
			t.Errorf("expected second diagnosis=I10, got %s", diags[1].DiagnosisCode)
		}
	})

	t.Run("Procedures", func(t *testing.T) {
		var claimID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			c := &billing.Claim{
				Status:    "active",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			claimID = c.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create claim: %v", err)
		}

		now := time.Now()
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			p := &billing.ClaimProcedure{
				ClaimID:             claimID,
				Sequence:            1,
				TypeCode:            ptrStr("primary"),
				Date:                &now,
				ProcedureCodeSystem: ptrStr("http://www.ama-assn.org/go/cpt"),
				ProcedureCode:       "99213",
				ProcedureDisplay:    ptrStr("Office visit, established patient"),
			}
			return repo.AddProcedure(ctx, p)
		})
		if err != nil {
			t.Fatalf("AddProcedure: %v", err)
		}

		var procs []*billing.ClaimProcedure
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			var err error
			procs, err = repo.GetProcedures(ctx, claimID)
			return err
		})
		if err != nil {
			t.Fatalf("GetProcedures: %v", err)
		}
		if len(procs) != 1 {
			t.Fatalf("expected 1 procedure, got %d", len(procs))
		}
		if procs[0].ProcedureCode != "99213" {
			t.Errorf("expected procedure_code=99213, got %s", procs[0].ProcedureCode)
		}
	})

	t.Run("Items", func(t *testing.T) {
		var claimID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			c := &billing.Claim{
				Status:    "active",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			claimID = c.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create claim: %v", err)
		}

		now := time.Now()
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			item := &billing.ClaimItem{
				ClaimID:                 claimID,
				Sequence:                1,
				ProductOrServiceCode:    "99213",
				ProductOrServiceDisplay: ptrStr("Office visit"),
				ServicedDate:            &now,
				QuantityValue:           ptrFloat(1),
				QuantityUnit:            ptrStr("visit"),
				UnitPrice:               ptrFloat(150.00),
				NetAmount:               ptrFloat(150.00),
				Currency:                ptrStr("USD"),
				RevenueCode:             ptrStr("0510"),
				RevenueDisplay:          ptrStr("Clinic"),
				Note:                    ptrStr("Routine office visit"),
			}
			return repo.AddItem(ctx, item)
		})
		if err != nil {
			t.Fatalf("AddItem: %v", err)
		}

		// Add second item
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			item := &billing.ClaimItem{
				ClaimID:              claimID,
				Sequence:             2,
				ProductOrServiceCode: "85025",
				ProductOrServiceDisplay: ptrStr("Complete blood count"),
				ServicedDate:         &now,
				QuantityValue:        ptrFloat(1),
				UnitPrice:            ptrFloat(35.00),
				NetAmount:            ptrFloat(35.00),
				Currency:             ptrStr("USD"),
			}
			return repo.AddItem(ctx, item)
		})
		if err != nil {
			t.Fatalf("AddItem 2: %v", err)
		}

		var items []*billing.ClaimItem
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			var err error
			items, err = repo.GetItems(ctx, claimID)
			return err
		})
		if err != nil {
			t.Fatalf("GetItems: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(items))
		}
		if items[0].ProductOrServiceCode != "99213" {
			t.Errorf("expected first item code=99213, got %s", items[0].ProductOrServiceCode)
		}
		if items[0].NetAmount == nil || *items[0].NetAmount != 150.00 {
			t.Errorf("expected net_amount=150.00, got %v", items[0].NetAmount)
		}
		if items[1].Sequence != 2 {
			t.Errorf("expected sequence=2, got %d", items[1].Sequence)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var claimID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			c := &billing.Claim{
				Status:    "draft",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			claimID = c.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			return repo.Delete(ctx, claimID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, claimID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted claim")
		}
	})
}

func TestClaimResponseCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("clmrsp")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "CRPatient", "Test", "MRN-CR-001")

	// Create prerequisite claim
	var claimID uuid.UUID
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := billing.NewClaimRepoPG(globalDB.Pool)
		c := &billing.Claim{
			Status:      "active",
			TypeCode:    ptrStr("professional"),
			UseCode:     ptrStr("claim"),
			PatientID:   patient.ID,
			TotalAmount: ptrFloat(500.00),
			Currency:    ptrStr("USD"),
		}
		if err := repo.Create(ctx, c); err != nil {
			return err
		}
		claimID = c.ID
		return nil
	})
	if err != nil {
		t.Fatalf("Create prerequisite claim: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		var created *billing.ClaimResponse
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimResponseRepoPG(globalDB.Pool)
			cr := &billing.ClaimResponse{
				ClaimID:       claimID,
				Status:        "active",
				TypeCode:      ptrStr("professional"),
				UseCode:       ptrStr("claim"),
				Outcome:       ptrStr("complete"),
				Disposition:   ptrStr("Claim settled as per contract"),
				PaymentAmount: ptrFloat(425.00),
				PaymentDate:   &now,
				TotalAmount:   ptrFloat(500.00),
				ProcessNote:   ptrStr("Processed per network agreement"),
			}
			if err := repo.Create(ctx, cr); err != nil {
				return err
			}
			created = cr
			return nil
		})
		if err != nil {
			t.Fatalf("Create claim response: %v", err)
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
			repo := billing.NewClaimResponseRepoPG(globalDB.Pool)
			cr := &billing.ClaimResponse{
				ClaimID: uuid.New(), // non-existent
				Status:  "active",
			}
			return repo.Create(ctx, cr)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent claim")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var crID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimResponseRepoPG(globalDB.Pool)
			cr := &billing.ClaimResponse{
				ClaimID:       claimID,
				Status:        "active",
				Outcome:       ptrStr("partial"),
				PaymentAmount: ptrFloat(200.00),
			}
			if err := repo.Create(ctx, cr); err != nil {
				return err
			}
			crID = cr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *billing.ClaimResponse
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimResponseRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, crID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.ClaimID != claimID {
			t.Errorf("expected claim_id=%s, got %s", claimID, fetched.ClaimID)
		}
		if fetched.Outcome == nil || *fetched.Outcome != "partial" {
			t.Errorf("expected outcome=partial, got %v", fetched.Outcome)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimResponseRepoPG(globalDB.Pool)
			cr := &billing.ClaimResponse{
				ClaimID: claimID,
				Status:  "active",
			}
			if err := repo.Create(ctx, cr); err != nil {
				return err
			}
			fhirID = cr.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *billing.ClaimResponse
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimResponseRepoPG(globalDB.Pool)
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

	t.Run("ListByClaim", func(t *testing.T) {
		var results []*billing.ClaimResponse
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimResponseRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByClaim(ctx, claimID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByClaim: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 claim response")
		}
		for _, r := range results {
			if r.ClaimID != claimID {
				t.Errorf("expected claim_id=%s, got %s", claimID, r.ClaimID)
			}
		}
	})

	t.Run("Search_ByOutcome", func(t *testing.T) {
		var results []*billing.ClaimResponse
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewClaimResponseRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"request": claimID.String(),
				"outcome": "complete",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by outcome: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.Outcome == nil || *r.Outcome != "complete" {
				t.Errorf("expected outcome=complete, got %v", r.Outcome)
			}
		}
	})
}

func TestInvoiceCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("inv")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "InvPatient", "Test", "MRN-INV-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "InvDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *billing.Invoice
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
			inv := &billing.Invoice{
				Status:        "issued",
				TypeCode:      ptrStr("self-pay"),
				PatientID:     patient.ID,
				ParticipantID: &practitioner.ID,
				TotalNet:      ptrFloat(200.00),
				TotalGross:    ptrFloat(236.00),
				TotalTax:      ptrFloat(36.00),
				Currency:      ptrStr("USD"),
				PaymentTerms:  ptrStr("Net 30"),
				Note:          ptrStr("Office visit invoice"),
			}
			if err := repo.Create(ctx, inv); err != nil {
				return err
			}
			created = inv
			return nil
		})
		if err != nil {
			t.Fatalf("Create invoice: %v", err)
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
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
			inv := &billing.Invoice{
				Status:    "issued",
				PatientID: uuid.New(), // non-existent
			}
			return repo.Create(ctx, inv)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var invID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
			inv := &billing.Invoice{
				Status:     "issued",
				PatientID:  patient.ID,
				TotalNet:   ptrFloat(100.00),
				TotalGross: ptrFloat(118.00),
				Currency:   ptrStr("USD"),
			}
			if err := repo.Create(ctx, inv); err != nil {
				return err
			}
			invID = inv.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *billing.Invoice
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, invID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.TotalNet == nil || *fetched.TotalNet != 100.00 {
			t.Errorf("expected total_net=100.00, got %v", fetched.TotalNet)
		}
		if fetched.Status != "issued" {
			t.Errorf("expected status=issued, got %s", fetched.Status)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
			inv := &billing.Invoice{
				Status:    "draft",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, inv); err != nil {
				return err
			}
			fhirID = inv.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *billing.Invoice
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
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
		var inv *billing.Invoice
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
			i := &billing.Invoice{
				Status:     "draft",
				PatientID:  patient.ID,
				TotalNet:   ptrFloat(500.00),
				TotalGross: ptrFloat(590.00),
				TotalTax:   ptrFloat(90.00),
			}
			if err := repo.Create(ctx, i); err != nil {
				return err
			}
			inv = i
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
			inv.Status = "balanced"
			inv.TotalNet = ptrFloat(500.00)
			inv.TotalGross = ptrFloat(590.00)
			inv.TotalTax = ptrFloat(90.00)
			inv.PaymentTerms = ptrStr("Paid in full")
			inv.Note = ptrStr("Payment received")
			return repo.Update(ctx, inv)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *billing.Invoice
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, inv.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "balanced" {
			t.Errorf("expected status=balanced, got %s", fetched.Status)
		}
		if fetched.PaymentTerms == nil || *fetched.PaymentTerms != "Paid in full" {
			t.Errorf("expected payment_terms set, got %v", fetched.PaymentTerms)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*billing.Invoice
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 invoice")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search_ByStatus", func(t *testing.T) {
		var results []*billing.Invoice
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "issued",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by status: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.Status != "issued" {
				t.Errorf("expected status=issued, got %s", r.Status)
			}
		}
	})

	t.Run("LineItems", func(t *testing.T) {
		// Create an invoice
		var invID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
			inv := &billing.Invoice{
				Status:    "issued",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, inv); err != nil {
				return err
			}
			invID = inv.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create invoice: %v", err)
		}

		// Add line item
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
			li := &billing.InvoiceLineItem{
				InvoiceID:      invID,
				Sequence:       1,
				Description:    ptrStr("Office visit consultation"),
				ServiceCode:    ptrStr("99213"),
				ServiceDisplay: ptrStr("Office visit, established patient"),
				Quantity:       ptrFloat(1),
				UnitPrice:      ptrFloat(150.00),
				NetAmount:      ptrFloat(150.00),
				TaxAmount:      ptrFloat(0),
				GrossAmount:    ptrFloat(150.00),
				Currency:       ptrStr("USD"),
			}
			return repo.AddLineItem(ctx, li)
		})
		if err != nil {
			t.Fatalf("AddLineItem: %v", err)
		}

		// Add second line item
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
			li := &billing.InvoiceLineItem{
				InvoiceID:      invID,
				Sequence:       2,
				Description:    ptrStr("Lab - Complete blood count"),
				ServiceCode:    ptrStr("85025"),
				ServiceDisplay: ptrStr("CBC with differential"),
				Quantity:       ptrFloat(1),
				UnitPrice:      ptrFloat(35.00),
				NetAmount:      ptrFloat(35.00),
				TaxAmount:      ptrFloat(0),
				GrossAmount:    ptrFloat(35.00),
				Currency:       ptrStr("USD"),
			}
			return repo.AddLineItem(ctx, li)
		})
		if err != nil {
			t.Fatalf("AddLineItem 2: %v", err)
		}

		// Get line items
		var items []*billing.InvoiceLineItem
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
			var err error
			items, err = repo.GetLineItems(ctx, invID)
			return err
		})
		if err != nil {
			t.Fatalf("GetLineItems: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("expected 2 line items, got %d", len(items))
		}
		if items[0].ServiceCode == nil || *items[0].ServiceCode != "99213" {
			t.Errorf("expected first item code=99213, got %v", items[0].ServiceCode)
		}
		if items[0].NetAmount == nil || *items[0].NetAmount != 150.00 {
			t.Errorf("expected net_amount=150.00, got %v", items[0].NetAmount)
		}
		if items[1].Sequence != 2 {
			t.Errorf("expected second item sequence=2, got %d", items[1].Sequence)
		}
		if items[1].ServiceCode == nil || *items[1].ServiceCode != "85025" {
			t.Errorf("expected second item code=85025, got %v", items[1].ServiceCode)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var invID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
			inv := &billing.Invoice{
				Status:    "draft",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, inv); err != nil {
				return err
			}
			invID = inv.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
			return repo.Delete(ctx, invID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewInvoiceRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, invID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted invoice")
		}
	})
}

func TestExplanationOfBenefitCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("eob")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "EOBPatient", "Test", "MRN-EOB-001")

	t.Run("Create", func(t *testing.T) {
		var created *billing.ExplanationOfBenefit
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewExplanationOfBenefitRepoPG(globalDB.Pool)
			eob := &billing.ExplanationOfBenefit{
				Status:         "active",
				TypeCode:       ptrStr("professional"),
				UseCode:        ptrStr("claim"),
				PatientID:      patient.ID,
				Outcome:        ptrStr("complete"),
				TotalSubmitted: ptrFloat(500.00),
				TotalBenefit:   ptrFloat(400.00),
				TotalPayment:   ptrFloat(400.00),
				Currency:       ptrStr("USD"),
			}
			if err := repo.Create(ctx, eob); err != nil {
				return err
			}
			created = eob
			return nil
		})
		if err != nil {
			t.Fatalf("Create EOB: %v", err)
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
			repo := billing.NewExplanationOfBenefitRepoPG(globalDB.Pool)
			eob := &billing.ExplanationOfBenefit{
				Status:    "active",
				PatientID: uuid.New(),
			}
			return repo.Create(ctx, eob)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var eobID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewExplanationOfBenefitRepoPG(globalDB.Pool)
			eob := &billing.ExplanationOfBenefit{
				Status:         "active",
				TypeCode:       ptrStr("institutional"),
				PatientID:      patient.ID,
				Outcome:        ptrStr("partial"),
				TotalSubmitted: ptrFloat(1000.00),
				Currency:       ptrStr("USD"),
			}
			if err := repo.Create(ctx, eob); err != nil {
				return err
			}
			eobID = eob.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *billing.ExplanationOfBenefit
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewExplanationOfBenefitRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, eobID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if *fetched.TypeCode != "institutional" {
			t.Errorf("expected type_code=institutional, got %s", *fetched.TypeCode)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewExplanationOfBenefitRepoPG(globalDB.Pool)
			eob := &billing.ExplanationOfBenefit{
				Status:    "active",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, eob); err != nil {
				return err
			}
			fhirID = eob.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *billing.ExplanationOfBenefit
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewExplanationOfBenefitRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, fhirID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.FHIRID != fhirID {
			t.Errorf("expected FHIR ID=%s, got %s", fhirID, fetched.FHIRID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var eobID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewExplanationOfBenefitRepoPG(globalDB.Pool)
			eob := &billing.ExplanationOfBenefit{
				Status:    "active",
				PatientID: patient.ID,
				Outcome:   ptrStr("partial"),
			}
			if err := repo.Create(ctx, eob); err != nil {
				return err
			}
			eobID = eob.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewExplanationOfBenefitRepoPG(globalDB.Pool)
			eob, err := repo.GetByID(ctx, eobID)
			if err != nil {
				return err
			}
			eob.Status = "cancelled"
			eob.Outcome = ptrStr("complete")
			eob.Disposition = ptrStr("Claim denied")
			return repo.Update(ctx, eob)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *billing.ExplanationOfBenefit
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewExplanationOfBenefitRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, eobID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "cancelled" {
			t.Errorf("expected status=cancelled, got %s", fetched.Status)
		}
		if *fetched.Outcome != "complete" {
			t.Errorf("expected outcome=complete, got %s", *fetched.Outcome)
		}
	})

	t.Run("Search", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewExplanationOfBenefitRepoPG(globalDB.Pool)
			items, total, err := repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
			}, 100, 0)
			if err != nil {
				return err
			}
			if total == 0 {
				t.Error("expected non-zero total")
			}
			for _, eob := range items {
				if eob.PatientID != patient.ID {
					t.Errorf("expected patient_id=%s, got %s", patient.ID, eob.PatientID)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var eobID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewExplanationOfBenefitRepoPG(globalDB.Pool)
			eob := &billing.ExplanationOfBenefit{
				Status:    "active",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, eob); err != nil {
				return err
			}
			eobID = eob.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewExplanationOfBenefitRepoPG(globalDB.Pool)
			return repo.Delete(ctx, eobID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := billing.NewExplanationOfBenefitRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, eobID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted EOB")
		}
	})
}
