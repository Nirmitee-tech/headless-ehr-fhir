package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/surgery"
	"github.com/google/uuid"
)

func TestORRoomCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("orroom")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *surgery.ORRoom
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewORRoomRepoPG(globalDB.Pool)
			r := &surgery.ORRoom{
				Name:     "OR-1",
				Status:   "available",
				RoomType: ptrStr("general"),
				Equipment: ptrStr("Electrocautery, laparoscopic tower"),
				IsActive: true,
				Note:     ptrStr("Main surgical suite"),
			}
			if err := repo.Create(ctx, r); err != nil {
				return err
			}
			created = r
			return nil
		})
		if err != nil {
			t.Fatalf("Create OR room: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		room := createTestORRoom(t, ctx, tenantID, "OR-GET-1")

		var fetched *surgery.ORRoom
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewORRoomRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, room.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Name != "OR-GET-1" {
			t.Errorf("expected name=OR-GET-1, got %s", fetched.Name)
		}
		if fetched.Status != "available" {
			t.Errorf("expected status=available, got %s", fetched.Status)
		}
	})

	t.Run("Update", func(t *testing.T) {
		room := createTestORRoom(t, ctx, tenantID, "OR-UPD-1")

		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewORRoomRepoPG(globalDB.Pool)
			room.Status = "occupied"
			room.Equipment = ptrStr("Electrocautery, robot")
			room.DecontaminatedAt = &now
			room.Note = ptrStr("Updated equipment list")
			return repo.Update(ctx, room)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *surgery.ORRoom
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewORRoomRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, room.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "occupied" {
			t.Errorf("expected status=occupied, got %s", fetched.Status)
		}
		if fetched.DecontaminatedAt == nil {
			t.Error("expected non-nil DecontaminatedAt")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		room := createTestORRoom(t, ctx, tenantID, "OR-DEL-1")

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewORRoomRepoPG(globalDB.Pool)
			return repo.Delete(ctx, room.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewORRoomRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, room.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted OR room")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*surgery.ORRoom
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewORRoomRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 OR room")
		}
		_ = results
	})

	t.Run("Search", func(t *testing.T) {
		var results []*surgery.ORRoom
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewORRoomRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"status": "available"}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		for _, r := range results {
			if r.Status != "available" {
				t.Errorf("expected status=available, got %s", r.Status)
			}
		}
		_ = total
	})
}

func TestSurgicalCaseCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("surgcase")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "SurgPatient", "Test", "MRN-SURG-001")
	surgeon := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "SurgDoc", "Smith")
	room := createTestORRoom(t, ctx, tenantID, "OR-CASE-1")

	t.Run("Create", func(t *testing.T) {
		var created *surgery.SurgicalCase
		now := time.Now()
		schedStart := now.Add(2 * time.Hour)
		schedEnd := now.Add(5 * time.Hour)
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			sc := &surgery.SurgicalCase{
				PatientID:        patient.ID,
				PrimarySurgeonID: surgeon.ID,
				ORRoomID:         &room.ID,
				Status:           "scheduled",
				CaseClass:        ptrStr("elective"),
				ASAClass:         ptrStr("II"),
				ScheduledDate:    now,
				ScheduledStart:   &schedStart,
				ScheduledEnd:     &schedEnd,
				AnesthesiaType:   ptrStr("general"),
				Laterality:       ptrStr("right"),
				PreOpDiagnosis:   ptrStr("Right inguinal hernia"),
				Note:             ptrStr("Laparoscopic repair"),
			}
			if err := repo.Create(ctx, sc); err != nil {
				return err
			}
			created = sc
			return nil
		})
		if err != nil {
			t.Fatalf("Create surgical case: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		sc := createTestSurgicalCase(t, ctx, tenantID, patient.ID, surgeon.ID, &room.ID)

		var fetched *surgery.SurgicalCase
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, sc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %s", patient.ID, fetched.PatientID)
		}
		if fetched.PrimarySurgeonID != surgeon.ID {
			t.Errorf("expected surgeon_id=%s, got %s", surgeon.ID, fetched.PrimarySurgeonID)
		}
		if fetched.Status != "scheduled" {
			t.Errorf("expected status=scheduled, got %s", fetched.Status)
		}
	})

	t.Run("Update", func(t *testing.T) {
		sc := createTestSurgicalCase(t, ctx, tenantID, patient.ID, surgeon.ID, &room.ID)

		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			sc.Status = "in-progress"
			sc.ActualStart = &now
			sc.PostOpDiagnosis = ptrStr("Right inguinal hernia, direct")
			sc.WoundClass = ptrStr("clean")
			return repo.Update(ctx, sc)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *surgery.SurgicalCase
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, sc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "in-progress" {
			t.Errorf("expected status=in-progress, got %s", fetched.Status)
		}
		if fetched.ActualStart == nil {
			t.Error("expected non-nil ActualStart")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		sc := createTestSurgicalCase(t, ctx, tenantID, patient.ID, surgeon.ID, &room.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			return repo.Delete(ctx, sc.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, sc.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted surgical case")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*surgery.SurgicalCase
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 surgical case")
		}
		_ = results
	})

	t.Run("ListByPatient", func(t *testing.T) {
		createTestSurgicalCase(t, ctx, tenantID, patient.ID, surgeon.ID, &room.ID)

		var results []*surgery.SurgicalCase
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 case for patient")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*surgery.SurgicalCase
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"status": "scheduled"}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		for _, r := range results {
			if r.Status != "scheduled" {
				t.Errorf("expected status=scheduled, got %s", r.Status)
			}
		}
		_ = total
	})

	t.Run("AddProcedure_GetProcedures_RemoveProcedure", func(t *testing.T) {
		sc := createTestSurgicalCase(t, ctx, tenantID, patient.ID, surgeon.ID, &room.ID)

		var procID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			proc := &surgery.SurgicalCaseProcedure{
				SurgicalCaseID:   sc.ID,
				ProcedureCode:    "49650",
				ProcedureDisplay: "Laparoscopic inguinal hernia repair",
				CodeSystem:       ptrStr("http://www.ama-assn.org/go/cpt"),
				CPTCode:          ptrStr("49650"),
				IsPrimary:        true,
				BodySiteCode:     ptrStr("right-inguinal"),
				BodySiteDisplay:  ptrStr("Right inguinal region"),
				Sequence:         1,
			}
			if err := repo.AddProcedure(ctx, proc); err != nil {
				return err
			}
			procID = proc.ID
			return nil
		})
		if err != nil {
			t.Fatalf("AddProcedure: %v", err)
		}

		var procs []*surgery.SurgicalCaseProcedure
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			var err error
			procs, err = repo.GetProcedures(ctx, sc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetProcedures: %v", err)
		}
		if len(procs) != 1 {
			t.Fatalf("expected 1 procedure, got %d", len(procs))
		}
		if procs[0].ProcedureCode != "49650" {
			t.Errorf("expected code=49650, got %s", procs[0].ProcedureCode)
		}

		// Remove
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			return repo.RemoveProcedure(ctx, procID)
		})
		if err != nil {
			t.Fatalf("RemoveProcedure: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			var err error
			procs, err = repo.GetProcedures(ctx, sc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetProcedures after remove: %v", err)
		}
		if len(procs) != 0 {
			t.Errorf("expected 0 procedures after remove, got %d", len(procs))
		}
	})

	t.Run("AddTeamMember_GetTeamMembers_RemoveTeamMember", func(t *testing.T) {
		sc := createTestSurgicalCase(t, ctx, tenantID, patient.ID, surgeon.ID, &room.ID)
		nurse := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "ScrubNurse", "Jones")

		var memberID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			tm := &surgery.SurgicalCaseTeam{
				SurgicalCaseID: sc.ID,
				PractitionerID: nurse.ID,
				Role:           "scrub_nurse",
				RoleDisplay:    ptrStr("Scrub Nurse"),
			}
			if err := repo.AddTeamMember(ctx, tm); err != nil {
				return err
			}
			memberID = tm.ID
			return nil
		})
		if err != nil {
			t.Fatalf("AddTeamMember: %v", err)
		}

		var members []*surgery.SurgicalCaseTeam
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			var err error
			members, err = repo.GetTeamMembers(ctx, sc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetTeamMembers: %v", err)
		}
		if len(members) != 1 {
			t.Fatalf("expected 1 team member, got %d", len(members))
		}
		if members[0].Role != "scrub_nurse" {
			t.Errorf("expected role=scrub_nurse, got %s", members[0].Role)
		}

		// Remove
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			return repo.RemoveTeamMember(ctx, memberID)
		})
		if err != nil {
			t.Fatalf("RemoveTeamMember: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			var err error
			members, err = repo.GetTeamMembers(ctx, sc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetTeamMembers after remove: %v", err)
		}
		if len(members) != 0 {
			t.Errorf("expected 0 team members after remove, got %d", len(members))
		}
	})

	t.Run("AddTimeEvent_GetTimeEvents", func(t *testing.T) {
		sc := createTestSurgicalCase(t, ctx, tenantID, patient.ID, surgeon.ID, &room.ID)

		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			event := &surgery.SurgicalTimeEvent{
				SurgicalCaseID: sc.ID,
				EventType:      "patient_in_room",
				EventTime:      now,
				Note:           ptrStr("Patient arrived in OR"),
			}
			return repo.AddTimeEvent(ctx, event)
		})
		if err != nil {
			t.Fatalf("AddTimeEvent: %v", err)
		}

		// Add second event
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			event := &surgery.SurgicalTimeEvent{
				SurgicalCaseID: sc.ID,
				EventType:      "incision",
				EventTime:      now.Add(15 * time.Minute),
			}
			return repo.AddTimeEvent(ctx, event)
		})
		if err != nil {
			t.Fatalf("AddTimeEvent (second): %v", err)
		}

		var events []*surgery.SurgicalTimeEvent
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			var err error
			events, err = repo.GetTimeEvents(ctx, sc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetTimeEvents: %v", err)
		}
		if len(events) != 2 {
			t.Fatalf("expected 2 time events, got %d", len(events))
		}
		if events[0].EventType != "patient_in_room" {
			t.Errorf("expected first event type=patient_in_room, got %s", events[0].EventType)
		}
	})

	t.Run("AddCount_GetCounts", func(t *testing.T) {
		sc := createTestSurgicalCase(t, ctx, tenantID, patient.ID, surgeon.ID, &room.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			count := &surgery.SurgicalCount{
				SurgicalCaseID: sc.ID,
				CountType:      "sponge",
				ItemName:       "Laparotomy sponge",
				ExpectedCount:  10,
				ActualCount:    10,
				IsCorrect:      true,
				CountTime:      time.Now(),
				Note:           ptrStr("Initial count"),
			}
			return repo.AddCount(ctx, count)
		})
		if err != nil {
			t.Fatalf("AddCount: %v", err)
		}

		var counts []*surgery.SurgicalCount
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			var err error
			counts, err = repo.GetCounts(ctx, sc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetCounts: %v", err)
		}
		if len(counts) != 1 {
			t.Fatalf("expected 1 count, got %d", len(counts))
		}
		if !counts[0].IsCorrect {
			t.Error("expected count to be correct")
		}
	})

	t.Run("AddSupply_GetSupplies", func(t *testing.T) {
		sc := createTestSurgicalCase(t, ctx, tenantID, patient.ID, surgeon.ID, &room.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			supply := &surgery.SurgicalSupplyUsed{
				SurgicalCaseID: sc.ID,
				SupplyName:     "Surgical gloves",
				SupplyCode:     ptrStr("GLV-001"),
				Quantity:       4,
				UnitOfMeasure:  ptrStr("pair"),
				LotNumber:      ptrStr("LOT-2024-001"),
				Note:           ptrStr("Sterile latex-free"),
			}
			return repo.AddSupply(ctx, supply)
		})
		if err != nil {
			t.Fatalf("AddSupply: %v", err)
		}

		var supplies []*surgery.SurgicalSupplyUsed
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
			var err error
			supplies, err = repo.GetSupplies(ctx, sc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetSupplies: %v", err)
		}
		if len(supplies) != 1 {
			t.Fatalf("expected 1 supply, got %d", len(supplies))
		}
		if supplies[0].SupplyName != "Surgical gloves" {
			t.Errorf("expected supply=Surgical gloves, got %s", supplies[0].SupplyName)
		}
	})
}

func TestPreferenceCardCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("prefcard")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	surgeon := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "PrefSurgeon", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *surgery.SurgicalPreferenceCard
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewPreferenceCardRepoPG(globalDB.Pool)
			pc := &surgery.SurgicalPreferenceCard{
				SurgeonID:        surgeon.ID,
				ProcedureCode:    "49650",
				ProcedureDisplay: "Laparoscopic hernia repair",
				GloveSizeL:       ptrStr("7.5"),
				GloveSizeR:       ptrStr("7.5"),
				Gown:             ptrStr("large"),
				SkinPrep:         ptrStr("ChloraPrep"),
				Position:         ptrStr("supine"),
				Instruments:      ptrStr("Standard laparoscopic set"),
				Supplies:         ptrStr("Mesh, tacks"),
				Sutures:          ptrStr("3-0 Vicryl"),
				IsActive:         true,
				Note:             ptrStr("Preferred setup"),
			}
			if err := repo.Create(ctx, pc); err != nil {
				return err
			}
			created = pc
			return nil
		})
		if err != nil {
			t.Fatalf("Create preference card: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		pc := createTestPreferenceCard(t, ctx, tenantID, surgeon.ID)

		var fetched *surgery.SurgicalPreferenceCard
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewPreferenceCardRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, pc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.SurgeonID != surgeon.ID {
			t.Errorf("expected surgeon_id=%s, got %s", surgeon.ID, fetched.SurgeonID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		pc := createTestPreferenceCard(t, ctx, tenantID, surgeon.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewPreferenceCardRepoPG(globalDB.Pool)
			pc.SkinPrep = ptrStr("Betadine")
			pc.Instruments = ptrStr("Updated instrument list")
			pc.Note = ptrStr("Updated preferences")
			return repo.Update(ctx, pc)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *surgery.SurgicalPreferenceCard
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewPreferenceCardRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, pc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.SkinPrep == nil || *fetched.SkinPrep != "Betadine" {
			t.Errorf("expected skin_prep=Betadine, got %v", fetched.SkinPrep)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		pc := createTestPreferenceCard(t, ctx, tenantID, surgeon.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewPreferenceCardRepoPG(globalDB.Pool)
			return repo.Delete(ctx, pc.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewPreferenceCardRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, pc.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted preference card")
		}
	})

	t.Run("ListBySurgeon", func(t *testing.T) {
		createTestPreferenceCard(t, ctx, tenantID, surgeon.ID)

		var results []*surgery.SurgicalPreferenceCard
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewPreferenceCardRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListBySurgeon(ctx, surgeon.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListBySurgeon: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 preference card for surgeon")
		}
		for _, r := range results {
			if r.SurgeonID != surgeon.ID {
				t.Errorf("expected surgeon_id=%s, got %s", surgeon.ID, r.SurgeonID)
			}
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*surgery.SurgicalPreferenceCard
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewPreferenceCardRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"surgeon_id": surgeon.ID.String()}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result")
		}
		_ = results
	})
}

func TestImplantLogCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("implant")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "ImplPatient", "Test", "MRN-IMPL-001")
	surgeon := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "ImplSurgeon", "Smith")
	room := createTestORRoom(t, ctx, tenantID, "OR-IMPL-1")
	sc := createTestSurgicalCase(t, ctx, tenantID, patient.ID, surgeon.ID, &room.ID)

	t.Run("Create", func(t *testing.T) {
		var created *surgery.ImplantLog
		now := time.Now()
		expDate := now.Add(10 * 365 * 24 * time.Hour)
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewImplantLogRepoPG(globalDB.Pool)
			il := &surgery.ImplantLog{
				SurgicalCaseID:  &sc.ID,
				PatientID:       patient.ID,
				ImplantType:     "mesh",
				Manufacturer:    ptrStr("Ethicon"),
				ModelNumber:     ptrStr("PM-2000"),
				SerialNumber:    ptrStr("SN-12345"),
				LotNumber:       ptrStr("LOT-ABC-001"),
				ExpirationDate:  &expDate,
				BodySiteCode:    ptrStr("right-inguinal"),
				BodySiteDisplay: ptrStr("Right inguinal region"),
				ImplantedBy:     &surgeon.ID,
				ImplantDate:     &now,
				Note:            ptrStr("Placed without complications"),
			}
			if err := repo.Create(ctx, il); err != nil {
				return err
			}
			created = il
			return nil
		})
		if err != nil {
			t.Fatalf("Create implant log: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		il := createTestImplantLog(t, ctx, tenantID, patient.ID, &sc.ID)

		var fetched *surgery.ImplantLog
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewImplantLogRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, il.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %s", patient.ID, fetched.PatientID)
		}
		if fetched.ImplantType != "knee_prosthesis" {
			t.Errorf("expected implant_type=knee_prosthesis, got %s", fetched.ImplantType)
		}
	})

	t.Run("Update", func(t *testing.T) {
		il := createTestImplantLog(t, ctx, tenantID, patient.ID, &sc.ID)

		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewImplantLogRepoPG(globalDB.Pool)
			il.ExplantDate = &now
			il.Note = ptrStr("Explanted due to infection")
			return repo.Update(ctx, il)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *surgery.ImplantLog
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewImplantLogRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, il.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.ExplantDate == nil {
			t.Error("expected non-nil ExplantDate")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		il := createTestImplantLog(t, ctx, tenantID, patient.ID, &sc.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewImplantLogRepoPG(globalDB.Pool)
			return repo.Delete(ctx, il.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewImplantLogRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, il.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted implant log")
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		createTestImplantLog(t, ctx, tenantID, patient.ID, &sc.ID)

		var results []*surgery.ImplantLog
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewImplantLogRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 implant for patient")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("ListByCase", func(t *testing.T) {
		var results []*surgery.ImplantLog
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewImplantLogRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByCase(ctx, sc.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByCase: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 implant for case")
		}
		_ = results
	})

	t.Run("Search", func(t *testing.T) {
		var results []*surgery.ImplantLog
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := surgery.NewImplantLogRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"patient_id": patient.ID.String()}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result")
		}
		_ = results
	})
}

// =========== Test Helpers ===========

func createTestORRoom(t *testing.T, ctx context.Context, tenantID, name string) *surgery.ORRoom {
	t.Helper()
	var result *surgery.ORRoom
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := surgery.NewORRoomRepoPG(globalDB.Pool)
		r := &surgery.ORRoom{
			Name:     name,
			Status:   "available",
			RoomType: ptrStr("general"),
			IsActive: true,
		}
		if err := repo.Create(ctx, r); err != nil {
			return err
		}
		result = r
		return nil
	})
	if err != nil {
		t.Fatalf("create test OR room: %v", err)
	}
	return result
}

func createTestSurgicalCase(t *testing.T, ctx context.Context, tenantID string, patientID, surgeonID uuid.UUID, roomID *uuid.UUID) *surgery.SurgicalCase {
	t.Helper()
	var result *surgery.SurgicalCase
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := surgery.NewSurgicalCaseRepoPG(globalDB.Pool)
		sc := &surgery.SurgicalCase{
			PatientID:        patientID,
			PrimarySurgeonID: surgeonID,
			ORRoomID:         roomID,
			Status:           "scheduled",
			CaseClass:        ptrStr("elective"),
			ScheduledDate:    time.Now(),
			PreOpDiagnosis:   ptrStr("Test diagnosis"),
		}
		if err := repo.Create(ctx, sc); err != nil {
			return err
		}
		result = sc
		return nil
	})
	if err != nil {
		t.Fatalf("create test surgical case: %v", err)
	}
	return result
}

func createTestPreferenceCard(t *testing.T, ctx context.Context, tenantID string, surgeonID uuid.UUID) *surgery.SurgicalPreferenceCard {
	t.Helper()
	var result *surgery.SurgicalPreferenceCard
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := surgery.NewPreferenceCardRepoPG(globalDB.Pool)
		pc := &surgery.SurgicalPreferenceCard{
			SurgeonID:        surgeonID,
			ProcedureCode:    "27447",
			ProcedureDisplay: "Total knee arthroplasty",
			GloveSizeL:       ptrStr("8.0"),
			GloveSizeR:       ptrStr("8.0"),
			SkinPrep:         ptrStr("ChloraPrep"),
			Position:         ptrStr("supine"),
			IsActive:         true,
		}
		if err := repo.Create(ctx, pc); err != nil {
			return err
		}
		result = pc
		return nil
	})
	if err != nil {
		t.Fatalf("create test preference card: %v", err)
	}
	return result
}

func createTestImplantLog(t *testing.T, ctx context.Context, tenantID string, patientID uuid.UUID, caseID *uuid.UUID) *surgery.ImplantLog {
	t.Helper()
	var result *surgery.ImplantLog
	now := time.Now()
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := surgery.NewImplantLogRepoPG(globalDB.Pool)
		il := &surgery.ImplantLog{
			SurgicalCaseID: caseID,
			PatientID:      patientID,
			ImplantType:    "knee_prosthesis",
			Manufacturer:   ptrStr("Stryker"),
			ModelNumber:    ptrStr("Triathlon"),
			SerialNumber:   ptrStr("SN-99999"),
			LotNumber:      ptrStr("LOT-XYZ-001"),
			ImplantDate:    &now,
		}
		if err := repo.Create(ctx, il); err != nil {
			return err
		}
		result = il
		return nil
	})
	if err != nil {
		t.Fatalf("create test implant log: %v", err)
	}
	return result
}
