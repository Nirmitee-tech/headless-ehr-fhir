package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/scheduling"
	"github.com/google/uuid"
)

func TestScheduleCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("sched")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "SchedDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *scheduling.Schedule
		now := time.Now()
		endDate := now.Add(30 * 24 * time.Hour)
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewScheduleRepoPG(globalDB.Pool)
			s := &scheduling.Schedule{
				Active:               ptrBool(true),
				PractitionerID:       practitioner.ID,
				ServiceTypeCode:      ptrStr("394802001"),
				ServiceTypeDisplay:   ptrStr("General medicine"),
				SpecialtyCode:        ptrStr("394814009"),
				SpecialtyDisplay:     ptrStr("General practice"),
				PlanningHorizonStart: &now,
				PlanningHorizonEnd:   &endDate,
				Comment:              ptrStr("Morning clinic schedule"),
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			created = s
			return nil
		})
		if err != nil {
			t.Fatalf("Create schedule: %v", err)
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
			repo := scheduling.NewScheduleRepoPG(globalDB.Pool)
			s := &scheduling.Schedule{
				Active:         ptrBool(true),
				PractitionerID: uuid.New(), // non-existent
			}
			return repo.Create(ctx, s)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent practitioner")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var schedID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewScheduleRepoPG(globalDB.Pool)
			s := &scheduling.Schedule{
				Active:          ptrBool(true),
				PractitionerID:  practitioner.ID,
				ServiceTypeCode: ptrStr("394579002"),
				ServiceTypeDisplay: ptrStr("Cardiology"),
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			schedID = s.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *scheduling.Schedule
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewScheduleRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, schedID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.PractitionerID != practitioner.ID {
			t.Errorf("expected practitioner_id=%s, got %s", practitioner.ID, fetched.PractitionerID)
		}
		if fetched.ServiceTypeCode == nil || *fetched.ServiceTypeCode != "394579002" {
			t.Errorf("expected service_type_code=394579002, got %v", fetched.ServiceTypeCode)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewScheduleRepoPG(globalDB.Pool)
			s := &scheduling.Schedule{
				Active:         ptrBool(true),
				PractitionerID: practitioner.ID,
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			fhirID = s.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *scheduling.Schedule
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewScheduleRepoPG(globalDB.Pool)
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
		var sched *scheduling.Schedule
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewScheduleRepoPG(globalDB.Pool)
			s := &scheduling.Schedule{
				Active:         ptrBool(true),
				PractitionerID: practitioner.ID,
				Comment:        ptrStr("Original comment"),
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			sched = s
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewScheduleRepoPG(globalDB.Pool)
			sched.Active = ptrBool(false)
			sched.Comment = ptrStr("Schedule deactivated for vacation")
			sched.SpecialtyCode = ptrStr("394586005")
			sched.SpecialtyDisplay = ptrStr("Gynecology")
			return repo.Update(ctx, sched)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *scheduling.Schedule
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewScheduleRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, sched.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Active == nil || *fetched.Active != false {
			t.Errorf("expected active=false, got %v", fetched.Active)
		}
		if fetched.Comment == nil || *fetched.Comment != "Schedule deactivated for vacation" {
			t.Errorf("expected updated comment, got %v", fetched.Comment)
		}
	})

	t.Run("ListByPractitioner", func(t *testing.T) {
		var results []*scheduling.Schedule
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewScheduleRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPractitioner(ctx, practitioner.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPractitioner: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 schedule")
		}
		for _, r := range results {
			if r.PractitionerID != practitioner.ID {
				t.Errorf("expected practitioner_id=%s, got %s", practitioner.ID, r.PractitionerID)
			}
		}
	})

	t.Run("Search_ByActive", func(t *testing.T) {
		var results []*scheduling.Schedule
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewScheduleRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"practitioner": practitioner.ID.String(),
				"active":       "true",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by active: %v", err)
		}
		_ = total
		_ = results
	})

	t.Run("Delete", func(t *testing.T) {
		var schedID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewScheduleRepoPG(globalDB.Pool)
			s := &scheduling.Schedule{
				Active:         ptrBool(true),
				PractitionerID: practitioner.ID,
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			schedID = s.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewScheduleRepoPG(globalDB.Pool)
			return repo.Delete(ctx, schedID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewScheduleRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, schedID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted schedule")
		}
	})
}

func TestSlotCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("slot")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "SlotDoc", "Smith")

	// Create a schedule first (prerequisite for slots)
	var scheduleID uuid.UUID
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := scheduling.NewScheduleRepoPG(globalDB.Pool)
		s := &scheduling.Schedule{
			Active:         ptrBool(true),
			PractitionerID: practitioner.ID,
		}
		if err := repo.Create(ctx, s); err != nil {
			return err
		}
		scheduleID = s.ID
		return nil
	})
	if err != nil {
		t.Fatalf("Create prerequisite schedule: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		var created *scheduling.Slot
		startTime := time.Now().Add(24 * time.Hour).Truncate(time.Second)
		endTime := startTime.Add(30 * time.Minute)
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewSlotRepoPG(globalDB.Pool)
			sl := &scheduling.Slot{
				ScheduleID:         scheduleID,
				Status:             "free",
				StartTime:          startTime,
				EndTime:            endTime,
				ServiceTypeCode:    ptrStr("394802001"),
				ServiceTypeDisplay: ptrStr("General medicine"),
				Comment:            ptrStr("Morning slot"),
			}
			if err := repo.Create(ctx, sl); err != nil {
				return err
			}
			created = sl
			return nil
		})
		if err != nil {
			t.Fatalf("Create slot: %v", err)
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
			repo := scheduling.NewSlotRepoPG(globalDB.Pool)
			sl := &scheduling.Slot{
				ScheduleID: uuid.New(), // non-existent
				Status:     "free",
				StartTime:  time.Now(),
				EndTime:    time.Now().Add(30 * time.Minute),
			}
			return repo.Create(ctx, sl)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent schedule")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var slotID uuid.UUID
		startTime := time.Now().Add(25 * time.Hour).Truncate(time.Second)
		endTime := startTime.Add(30 * time.Minute)
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewSlotRepoPG(globalDB.Pool)
			sl := &scheduling.Slot{
				ScheduleID: scheduleID,
				Status:     "free",
				StartTime:  startTime,
				EndTime:    endTime,
			}
			if err := repo.Create(ctx, sl); err != nil {
				return err
			}
			slotID = sl.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *scheduling.Slot
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewSlotRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, slotID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.ScheduleID != scheduleID {
			t.Errorf("expected schedule_id=%s, got %s", scheduleID, fetched.ScheduleID)
		}
		if fetched.Status != "free" {
			t.Errorf("expected status=free, got %s", fetched.Status)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewSlotRepoPG(globalDB.Pool)
			sl := &scheduling.Slot{
				ScheduleID: scheduleID,
				Status:     "free",
				StartTime:  time.Now().Add(26 * time.Hour),
				EndTime:    time.Now().Add(26*time.Hour + 30*time.Minute),
			}
			if err := repo.Create(ctx, sl); err != nil {
				return err
			}
			fhirID = sl.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *scheduling.Slot
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewSlotRepoPG(globalDB.Pool)
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
		var slot *scheduling.Slot
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewSlotRepoPG(globalDB.Pool)
			sl := &scheduling.Slot{
				ScheduleID: scheduleID,
				Status:     "free",
				StartTime:  time.Now().Add(27 * time.Hour),
				EndTime:    time.Now().Add(27*time.Hour + 30*time.Minute),
			}
			if err := repo.Create(ctx, sl); err != nil {
				return err
			}
			slot = sl
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewSlotRepoPG(globalDB.Pool)
			slot.Status = "busy"
			slot.Overbooked = ptrBool(false)
			slot.Comment = ptrStr("Booked for patient visit")
			return repo.Update(ctx, slot)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *scheduling.Slot
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewSlotRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, slot.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "busy" {
			t.Errorf("expected status=busy, got %s", fetched.Status)
		}
		if fetched.Overbooked == nil || *fetched.Overbooked != false {
			t.Errorf("expected overbooked=false, got %v", fetched.Overbooked)
		}
	})

	t.Run("ListBySchedule", func(t *testing.T) {
		var results []*scheduling.Slot
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewSlotRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListBySchedule(ctx, scheduleID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListBySchedule: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 slot")
		}
		for _, r := range results {
			if r.ScheduleID != scheduleID {
				t.Errorf("expected schedule_id=%s, got %s", scheduleID, r.ScheduleID)
			}
		}
	})

	t.Run("SearchAvailable", func(t *testing.T) {
		var results []*scheduling.Slot
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewSlotRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.SearchAvailable(ctx, map[string]string{
				"schedule": scheduleID.String(),
				"status":   "free",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("SearchAvailable: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.Status != "free" {
				t.Errorf("expected status=free, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var slotID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewSlotRepoPG(globalDB.Pool)
			sl := &scheduling.Slot{
				ScheduleID: scheduleID,
				Status:     "free",
				StartTime:  time.Now().Add(48 * time.Hour),
				EndTime:    time.Now().Add(48*time.Hour + 30*time.Minute),
			}
			if err := repo.Create(ctx, sl); err != nil {
				return err
			}
			slotID = sl.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewSlotRepoPG(globalDB.Pool)
			return repo.Delete(ctx, slotID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewSlotRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, slotID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted slot")
		}
	})
}

func TestAppointmentCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("appt")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "ApptPatient", "Test", "MRN-APPT-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "ApptDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *scheduling.Appointment
		startTime := time.Now().Add(24 * time.Hour)
		endTime := startTime.Add(30 * time.Minute)
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			a := &scheduling.Appointment{
				Status:             "booked",
				ServiceTypeCode:    ptrStr("394802001"),
				ServiceTypeDisplay: ptrStr("General medicine"),
				Description:        ptrStr("Annual physical examination"),
				StartTime:          &startTime,
				EndTime:            &endTime,
				MinutesDuration:    ptrInt(30),
				PatientID:          patient.ID,
				PractitionerID:     &practitioner.ID,
				ReasonCode:         ptrStr("Z00.00"),
				ReasonDisplay:      ptrStr("General adult medical examination"),
				Note:               ptrStr("Patient prefers morning appointments"),
				PatientInstruction: ptrStr("Please arrive 15 minutes early"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			created = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create appointment: %v", err)
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
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			a := &scheduling.Appointment{
				Status:    "booked",
				PatientID: uuid.New(), // non-existent
			}
			return repo.Create(ctx, a)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID_and_Update", func(t *testing.T) {
		startTime := time.Now().Add(25 * time.Hour)
		endTime := startTime.Add(30 * time.Minute)
		var apptID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			a := &scheduling.Appointment{
				Status:          "booked",
				StartTime:       &startTime,
				EndTime:         &endTime,
				MinutesDuration: ptrInt(30),
				PatientID:       patient.ID,
				PractitionerID:  &practitioner.ID,
				Description:     ptrStr("Follow-up visit"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			apptID = a.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		// Get
		var fetched *scheduling.Appointment
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, apptID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "booked" {
			t.Errorf("expected status=booked, got %s", fetched.Status)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %s", patient.ID, fetched.PatientID)
		}

		// Update to cancelled
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			fetched.Status = "cancelled"
			fetched.CancellationReason = ptrStr("Patient requested cancellation")
			fetched.Note = ptrStr("Rescheduling needed")
			return repo.Update(ctx, fetched)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		// Verify
		var updated *scheduling.Appointment
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			var err error
			updated, err = repo.GetByID(ctx, apptID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if updated.Status != "cancelled" {
			t.Errorf("expected status=cancelled, got %s", updated.Status)
		}
		if updated.CancellationReason == nil || *updated.CancellationReason != "Patient requested cancellation" {
			t.Errorf("expected cancellation reason set, got %v", updated.CancellationReason)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			a := &scheduling.Appointment{
				Status:    "proposed",
				PatientID: patient.ID,
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

		var fetched *scheduling.Appointment
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
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

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*scheduling.Appointment
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 appointment")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("ListByPractitioner", func(t *testing.T) {
		var results []*scheduling.Appointment
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPractitioner(ctx, practitioner.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPractitioner: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.PractitionerID == nil || *r.PractitionerID != practitioner.ID {
				t.Errorf("expected practitioner_id=%s, got %v", practitioner.ID, r.PractitionerID)
			}
		}
	})

	t.Run("Search_ByStatus", func(t *testing.T) {
		var results []*scheduling.Appointment
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "booked",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by status: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.Status != "booked" {
				t.Errorf("expected status=booked, got %s", r.Status)
			}
		}
	})

	t.Run("Participants", func(t *testing.T) {
		var apptID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			a := &scheduling.Appointment{
				Status:    "booked",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			apptID = a.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create appointment: %v", err)
		}

		consultant := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "ConsultAppt", "Jones")

		// Add participant
		var participantID uuid.UUID
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			p := &scheduling.AppointmentParticipant{
				AppointmentID: apptID,
				ActorType:     "Practitioner",
				ActorID:       consultant.ID,
				RoleCode:      ptrStr("CON"),
				RoleDisplay:   ptrStr("consultant"),
				Status:        "accepted",
				Required:      ptrStr("required"),
			}
			if err := repo.AddParticipant(ctx, p); err != nil {
				return err
			}
			participantID = p.ID
			return nil
		})
		if err != nil {
			t.Fatalf("AddParticipant: %v", err)
		}

		// Get participants
		var parts []*scheduling.AppointmentParticipant
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			var err error
			parts, err = repo.GetParticipants(ctx, apptID)
			return err
		})
		if err != nil {
			t.Fatalf("GetParticipants: %v", err)
		}
		if len(parts) != 1 {
			t.Fatalf("expected 1 participant, got %d", len(parts))
		}
		if parts[0].ActorType != "Practitioner" {
			t.Errorf("expected actor_type=Practitioner, got %s", parts[0].ActorType)
		}
		if parts[0].ActorID != consultant.ID {
			t.Errorf("expected actor_id=%s, got %s", consultant.ID, parts[0].ActorID)
		}
		if parts[0].Status != "accepted" {
			t.Errorf("expected status=accepted, got %s", parts[0].Status)
		}

		// Remove participant
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			return repo.RemoveParticipant(ctx, participantID)
		})
		if err != nil {
			t.Fatalf("RemoveParticipant: %v", err)
		}

		// Verify removal
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			var err error
			parts, err = repo.GetParticipants(ctx, apptID)
			return err
		})
		if err != nil {
			t.Fatalf("GetParticipants after remove: %v", err)
		}
		if len(parts) != 0 {
			t.Errorf("expected 0 participants after remove, got %d", len(parts))
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var apptID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			a := &scheduling.Appointment{
				Status:    "proposed",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			apptID = a.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			return repo.Delete(ctx, apptID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewAppointmentRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, apptID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted appointment")
		}
	})
}

func TestWaitlistCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("wlist")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "WaitPatient", "Test", "MRN-WAIT-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "WaitDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *scheduling.Waitlist
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewWaitlistRepoPG(globalDB.Pool)
			w := &scheduling.Waitlist{
				PatientID:          patient.ID,
				PractitionerID:     &practitioner.ID,
				Department:         ptrStr("Cardiology"),
				ServiceTypeCode:    ptrStr("394579002"),
				ServiceTypeDisplay: ptrStr("Cardiology"),
				Priority:           ptrInt(2),
				QueueNumber:        ptrInt(5),
				Status:             "waiting",
				RequestedDate:      &now,
				CheckInTime:        &now,
				Note:               ptrStr("Patient waiting for echocardiogram"),
			}
			if err := repo.Create(ctx, w); err != nil {
				return err
			}
			created = w
			return nil
		})
		if err != nil {
			t.Fatalf("Create waitlist entry: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewWaitlistRepoPG(globalDB.Pool)
			w := &scheduling.Waitlist{
				PatientID: uuid.New(), // non-existent
				Status:    "waiting",
			}
			return repo.Create(ctx, w)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var wID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewWaitlistRepoPG(globalDB.Pool)
			w := &scheduling.Waitlist{
				PatientID:  patient.ID,
				Department: ptrStr("Orthopedics"),
				Status:     "waiting",
				Priority:   ptrInt(1),
			}
			if err := repo.Create(ctx, w); err != nil {
				return err
			}
			wID = w.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *scheduling.Waitlist
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewWaitlistRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, wID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %s", patient.ID, fetched.PatientID)
		}
		if fetched.Department == nil || *fetched.Department != "Orthopedics" {
			t.Errorf("expected department=Orthopedics, got %v", fetched.Department)
		}
		if fetched.Status != "waiting" {
			t.Errorf("expected status=waiting, got %s", fetched.Status)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var w *scheduling.Waitlist
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewWaitlistRepoPG(globalDB.Pool)
			entry := &scheduling.Waitlist{
				PatientID:  patient.ID,
				Department: ptrStr("Radiology"),
				Status:     "waiting",
				Priority:   ptrInt(3),
				QueueNumber: ptrInt(10),
			}
			if err := repo.Create(ctx, entry); err != nil {
				return err
			}
			w = entry
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		calledTime := now.Add(15 * time.Minute)
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewWaitlistRepoPG(globalDB.Pool)
			w.Status = "called"
			w.CalledTime = &calledTime
			w.Note = ptrStr("Patient called to room 3")
			return repo.Update(ctx, w)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *scheduling.Waitlist
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewWaitlistRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, w.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "called" {
			t.Errorf("expected status=called, got %s", fetched.Status)
		}
		if fetched.CalledTime == nil {
			t.Error("expected non-nil CalledTime")
		}
		if fetched.Note == nil || *fetched.Note != "Patient called to room 3" {
			t.Errorf("expected note set, got %v", fetched.Note)
		}
	})

	t.Run("ListByDepartment", func(t *testing.T) {
		// Create another waitlist entry for the same department
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewWaitlistRepoPG(globalDB.Pool)
			w := &scheduling.Waitlist{
				PatientID:  patient.ID,
				Department: ptrStr("Cardiology"),
				Status:     "waiting",
				QueueNumber: ptrInt(6),
			}
			return repo.Create(ctx, w)
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var results []*scheduling.Waitlist
		var total int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewWaitlistRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByDepartment(ctx, "Cardiology", 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByDepartment: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 waitlist entry for Cardiology")
		}
		for _, r := range results {
			if r.Department == nil || *r.Department != "Cardiology" {
				t.Errorf("expected department=Cardiology, got %v", r.Department)
			}
		}
	})

	t.Run("ListByPractitioner", func(t *testing.T) {
		var results []*scheduling.Waitlist
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewWaitlistRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPractitioner(ctx, practitioner.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPractitioner: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.PractitionerID == nil || *r.PractitionerID != practitioner.ID {
				t.Errorf("expected practitioner_id=%s, got %v", practitioner.ID, r.PractitionerID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var wID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewWaitlistRepoPG(globalDB.Pool)
			w := &scheduling.Waitlist{
				PatientID: patient.ID,
				Status:    "waiting",
			}
			if err := repo.Create(ctx, w); err != nil {
				return err
			}
			wID = w.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewWaitlistRepoPG(globalDB.Pool)
			return repo.Delete(ctx, wID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := scheduling.NewWaitlistRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, wID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted waitlist entry")
		}
	})
}
