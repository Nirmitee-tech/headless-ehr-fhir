package scheduling

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ehr/ehr/internal/platform/db"
)

type queryable interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

// =========== Schedule Repository ===========

type scheduleRepoPG struct{ pool *pgxpool.Pool }

func NewScheduleRepoPG(pool *pgxpool.Pool) ScheduleRepository { return &scheduleRepoPG{pool: pool} }

func (r *scheduleRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const schedCols = `id, fhir_id, active, practitioner_id, location_id,
	service_type_code, service_type_display, specialty_code, specialty_display,
	planning_horizon_start, planning_horizon_end, comment, created_at, updated_at`

func (r *scheduleRepoPG) scanSchedule(row pgx.Row) (*Schedule, error) {
	var s Schedule
	err := row.Scan(&s.ID, &s.FHIRID, &s.Active, &s.PractitionerID, &s.LocationID,
		&s.ServiceTypeCode, &s.ServiceTypeDisplay, &s.SpecialtyCode, &s.SpecialtyDisplay,
		&s.PlanningHorizonStart, &s.PlanningHorizonEnd, &s.Comment, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *scheduleRepoPG) Create(ctx context.Context, s *Schedule) error {
	s.ID = uuid.New()
	if s.FHIRID == "" {
		s.FHIRID = s.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO schedule (id, fhir_id, active, practitioner_id, location_id,
			service_type_code, service_type_display, specialty_code, specialty_display,
			planning_horizon_start, planning_horizon_end, comment)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		s.ID, s.FHIRID, s.Active, s.PractitionerID, s.LocationID,
		s.ServiceTypeCode, s.ServiceTypeDisplay, s.SpecialtyCode, s.SpecialtyDisplay,
		s.PlanningHorizonStart, s.PlanningHorizonEnd, s.Comment)
	return err
}

func (r *scheduleRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Schedule, error) {
	return r.scanSchedule(r.conn(ctx).QueryRow(ctx, `SELECT `+schedCols+` FROM schedule WHERE id = $1`, id))
}

func (r *scheduleRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Schedule, error) {
	return r.scanSchedule(r.conn(ctx).QueryRow(ctx, `SELECT `+schedCols+` FROM schedule WHERE fhir_id = $1`, fhirID))
}

func (r *scheduleRepoPG) Update(ctx context.Context, s *Schedule) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE schedule SET active=$2, location_id=$3, service_type_code=$4, service_type_display=$5,
			specialty_code=$6, specialty_display=$7, planning_horizon_start=$8, planning_horizon_end=$9,
			comment=$10, updated_at=NOW()
		WHERE id = $1`,
		s.ID, s.Active, s.LocationID, s.ServiceTypeCode, s.ServiceTypeDisplay,
		s.SpecialtyCode, s.SpecialtyDisplay, s.PlanningHorizonStart, s.PlanningHorizonEnd, s.Comment)
	return err
}

func (r *scheduleRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM schedule WHERE id = $1`, id)
	return err
}

func (r *scheduleRepoPG) ListByPractitioner(ctx context.Context, practitionerID uuid.UUID, limit, offset int) ([]*Schedule, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM schedule WHERE practitioner_id = $1`, practitionerID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+schedCols+` FROM schedule WHERE practitioner_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, practitionerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Schedule
	for rows.Next() {
		s, err := r.scanSchedule(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

func (r *scheduleRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Schedule, int, error) {
	query := `SELECT ` + schedCols + ` FROM schedule WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM schedule WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["practitioner"]; ok {
		query += fmt.Sprintf(` AND practitioner_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND practitioner_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["active"]; ok {
		query += fmt.Sprintf(` AND active = $%d`, idx)
		countQuery += fmt.Sprintf(` AND active = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["service-type"]; ok {
		query += fmt.Sprintf(` AND service_type_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND service_type_code = $%d`, idx)
		args = append(args, p)
		idx++
	}

	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Schedule
	for rows.Next() {
		s, err := r.scanSchedule(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

// =========== Slot Repository ===========

type slotRepoPG struct{ pool *pgxpool.Pool }

func NewSlotRepoPG(pool *pgxpool.Pool) SlotRepository { return &slotRepoPG{pool: pool} }

func (r *slotRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const slotCols = `id, fhir_id, schedule_id, status, start_time, end_time, overbooked, comment,
	service_type_code, service_type_display, specialty_code, specialty_display,
	appointment_type_code, appointment_type_display, created_at, updated_at`

func (r *slotRepoPG) scanSlot(row pgx.Row) (*Slot, error) {
	var sl Slot
	err := row.Scan(&sl.ID, &sl.FHIRID, &sl.ScheduleID, &sl.Status, &sl.StartTime, &sl.EndTime,
		&sl.Overbooked, &sl.Comment,
		&sl.ServiceTypeCode, &sl.ServiceTypeDisplay, &sl.SpecialtyCode, &sl.SpecialtyDisplay,
		&sl.AppointmentTypeCode, &sl.AppointmentTypeDisplay, &sl.CreatedAt, &sl.UpdatedAt)
	return &sl, err
}

func (r *slotRepoPG) Create(ctx context.Context, sl *Slot) error {
	sl.ID = uuid.New()
	if sl.FHIRID == "" {
		sl.FHIRID = sl.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO slot (id, fhir_id, schedule_id, status, start_time, end_time, overbooked, comment,
			service_type_code, service_type_display, specialty_code, specialty_display,
			appointment_type_code, appointment_type_display)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		sl.ID, sl.FHIRID, sl.ScheduleID, sl.Status, sl.StartTime, sl.EndTime,
		sl.Overbooked, sl.Comment,
		sl.ServiceTypeCode, sl.ServiceTypeDisplay, sl.SpecialtyCode, sl.SpecialtyDisplay,
		sl.AppointmentTypeCode, sl.AppointmentTypeDisplay)
	return err
}

func (r *slotRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Slot, error) {
	return r.scanSlot(r.conn(ctx).QueryRow(ctx, `SELECT `+slotCols+` FROM slot WHERE id = $1`, id))
}

func (r *slotRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Slot, error) {
	return r.scanSlot(r.conn(ctx).QueryRow(ctx, `SELECT `+slotCols+` FROM slot WHERE fhir_id = $1`, fhirID))
}

func (r *slotRepoPG) Update(ctx context.Context, sl *Slot) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE slot SET status=$2, overbooked=$3, comment=$4, updated_at=NOW()
		WHERE id = $1`,
		sl.ID, sl.Status, sl.Overbooked, sl.Comment)
	return err
}

func (r *slotRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM slot WHERE id = $1`, id)
	return err
}

func (r *slotRepoPG) ListBySchedule(ctx context.Context, scheduleID uuid.UUID, limit, offset int) ([]*Slot, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM slot WHERE schedule_id = $1`, scheduleID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+slotCols+` FROM slot WHERE schedule_id = $1 ORDER BY start_time ASC LIMIT $2 OFFSET $3`, scheduleID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Slot
	for rows.Next() {
		sl, err := r.scanSlot(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, sl)
	}
	return items, total, nil
}

func (r *slotRepoPG) SearchAvailable(ctx context.Context, params map[string]string, limit, offset int) ([]*Slot, int, error) {
	query := `SELECT ` + slotCols + ` FROM slot WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM slot WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["schedule"]; ok {
		query += fmt.Sprintf(` AND schedule_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND schedule_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["start"]; ok {
		query += fmt.Sprintf(` AND start_time >= $%d`, idx)
		countQuery += fmt.Sprintf(` AND start_time >= $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["service-type"]; ok {
		query += fmt.Sprintf(` AND service_type_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND service_type_code = $%d`, idx)
		args = append(args, p)
		idx++
	}

	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(` ORDER BY start_time ASC LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Slot
	for rows.Next() {
		sl, err := r.scanSlot(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, sl)
	}
	return items, total, nil
}

// =========== Appointment Repository ===========

type appointmentRepoPG struct{ pool *pgxpool.Pool }

func NewAppointmentRepoPG(pool *pgxpool.Pool) AppointmentRepository {
	return &appointmentRepoPG{pool: pool}
}

func (r *appointmentRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const apptCols = `id, fhir_id, status, cancellation_reason,
	service_type_code, service_type_display, specialty_code, specialty_display,
	appointment_type_code, appointment_type_display, priority, description,
	start_time, end_time, minutes_duration, slot_id,
	patient_id, practitioner_id, location_id,
	reason_code, reason_display, reason_condition_id,
	note, patient_instruction, is_telehealth, telehealth_url,
	created_at, updated_at`

func (r *appointmentRepoPG) scanAppt(row pgx.Row) (*Appointment, error) {
	var a Appointment
	err := row.Scan(&a.ID, &a.FHIRID, &a.Status, &a.CancellationReason,
		&a.ServiceTypeCode, &a.ServiceTypeDisplay, &a.SpecialtyCode, &a.SpecialtyDisplay,
		&a.AppointmentTypeCode, &a.AppointmentTypeDisplay, &a.Priority, &a.Description,
		&a.StartTime, &a.EndTime, &a.MinutesDuration, &a.SlotID,
		&a.PatientID, &a.PractitionerID, &a.LocationID,
		&a.ReasonCode, &a.ReasonDisplay, &a.ReasonConditionID,
		&a.Note, &a.PatientInstruction, &a.IsTelehealth, &a.TelehealthURL,
		&a.CreatedAt, &a.UpdatedAt)
	return &a, err
}

func (r *appointmentRepoPG) Create(ctx context.Context, a *Appointment) error {
	a.ID = uuid.New()
	if a.FHIRID == "" {
		a.FHIRID = a.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO appointment (id, fhir_id, status, cancellation_reason,
			service_type_code, service_type_display, specialty_code, specialty_display,
			appointment_type_code, appointment_type_display, priority, description,
			start_time, end_time, minutes_duration, slot_id,
			patient_id, practitioner_id, location_id,
			reason_code, reason_display, reason_condition_id,
			note, patient_instruction, is_telehealth, telehealth_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26)`,
		a.ID, a.FHIRID, a.Status, a.CancellationReason,
		a.ServiceTypeCode, a.ServiceTypeDisplay, a.SpecialtyCode, a.SpecialtyDisplay,
		a.AppointmentTypeCode, a.AppointmentTypeDisplay, a.Priority, a.Description,
		a.StartTime, a.EndTime, a.MinutesDuration, a.SlotID,
		a.PatientID, a.PractitionerID, a.LocationID,
		a.ReasonCode, a.ReasonDisplay, a.ReasonConditionID,
		a.Note, a.PatientInstruction, a.IsTelehealth, a.TelehealthURL)
	return err
}

func (r *appointmentRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Appointment, error) {
	return r.scanAppt(r.conn(ctx).QueryRow(ctx, `SELECT `+apptCols+` FROM appointment WHERE id = $1`, id))
}

func (r *appointmentRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Appointment, error) {
	return r.scanAppt(r.conn(ctx).QueryRow(ctx, `SELECT `+apptCols+` FROM appointment WHERE fhir_id = $1`, fhirID))
}

func (r *appointmentRepoPG) Update(ctx context.Context, a *Appointment) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE appointment SET status=$2, cancellation_reason=$3, start_time=$4, end_time=$5,
			minutes_duration=$6, practitioner_id=$7, location_id=$8,
			reason_code=$9, reason_display=$10, note=$11, patient_instruction=$12,
			is_telehealth=$13, telehealth_url=$14, updated_at=NOW()
		WHERE id = $1`,
		a.ID, a.Status, a.CancellationReason, a.StartTime, a.EndTime,
		a.MinutesDuration, a.PractitionerID, a.LocationID,
		a.ReasonCode, a.ReasonDisplay, a.Note, a.PatientInstruction,
		a.IsTelehealth, a.TelehealthURL)
	return err
}

func (r *appointmentRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM appointment WHERE id = $1`, id)
	return err
}

func (r *appointmentRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Appointment, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM appointment WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+apptCols+` FROM appointment WHERE patient_id = $1 ORDER BY start_time DESC NULLS LAST LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Appointment
	for rows.Next() {
		a, err := r.scanAppt(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

func (r *appointmentRepoPG) ListByPractitioner(ctx context.Context, practitionerID uuid.UUID, limit, offset int) ([]*Appointment, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM appointment WHERE practitioner_id = $1`, practitionerID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+apptCols+` FROM appointment WHERE practitioner_id = $1 ORDER BY start_time DESC NULLS LAST LIMIT $2 OFFSET $3`, practitionerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Appointment
	for rows.Next() {
		a, err := r.scanAppt(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

func (r *appointmentRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Appointment, int, error) {
	query := `SELECT ` + apptCols + ` FROM appointment WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM appointment WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient"]; ok {
		query += fmt.Sprintf(` AND patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND patient_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["practitioner"]; ok {
		query += fmt.Sprintf(` AND practitioner_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND practitioner_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["date"]; ok {
		query += fmt.Sprintf(` AND start_time::date = $%d`, idx)
		countQuery += fmt.Sprintf(` AND start_time::date = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["service-type"]; ok {
		query += fmt.Sprintf(` AND service_type_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND service_type_code = $%d`, idx)
		args = append(args, p)
		idx++
	}

	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(` ORDER BY start_time DESC NULLS LAST LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Appointment
	for rows.Next() {
		a, err := r.scanAppt(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

func (r *appointmentRepoPG) AddParticipant(ctx context.Context, p *AppointmentParticipant) error {
	p.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO appointment_participant (id, appointment_id, actor_type, actor_id,
			role_code, role_display, status, required, period_start, period_end)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		p.ID, p.AppointmentID, p.ActorType, p.ActorID,
		p.RoleCode, p.RoleDisplay, p.Status, p.Required, p.PeriodStart, p.PeriodEnd)
	return err
}

func (r *appointmentRepoPG) GetParticipants(ctx context.Context, appointmentID uuid.UUID) ([]*AppointmentParticipant, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, appointment_id, actor_type, actor_id, role_code, role_display,
			status, required, period_start, period_end
		FROM appointment_participant WHERE appointment_id = $1`, appointmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*AppointmentParticipant
	for rows.Next() {
		var p AppointmentParticipant
		if err := rows.Scan(&p.ID, &p.AppointmentID, &p.ActorType, &p.ActorID,
			&p.RoleCode, &p.RoleDisplay, &p.Status, &p.Required, &p.PeriodStart, &p.PeriodEnd); err != nil {
			return nil, err
		}
		items = append(items, &p)
	}
	return items, nil
}

func (r *appointmentRepoPG) RemoveParticipant(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM appointment_participant WHERE id = $1`, id)
	return err
}

// =========== Waitlist Repository ===========

type waitlistRepoPG struct{ pool *pgxpool.Pool }

func NewWaitlistRepoPG(pool *pgxpool.Pool) WaitlistRepository { return &waitlistRepoPG{pool: pool} }

func (r *waitlistRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const waitlistCols = `id, patient_id, practitioner_id, department,
	service_type_code, service_type_display, priority, queue_number, status,
	requested_date, check_in_time, called_time, completed_time, note,
	created_at, updated_at`

func (r *waitlistRepoPG) scanWaitlist(row pgx.Row) (*Waitlist, error) {
	var w Waitlist
	err := row.Scan(&w.ID, &w.PatientID, &w.PractitionerID, &w.Department,
		&w.ServiceTypeCode, &w.ServiceTypeDisplay, &w.Priority, &w.QueueNumber, &w.Status,
		&w.RequestedDate, &w.CheckInTime, &w.CalledTime, &w.CompletedTime, &w.Note,
		&w.CreatedAt, &w.UpdatedAt)
	return &w, err
}

func (r *waitlistRepoPG) Create(ctx context.Context, w *Waitlist) error {
	w.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO waitlist (id, patient_id, practitioner_id, department,
			service_type_code, service_type_display, priority, queue_number, status,
			requested_date, check_in_time, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		w.ID, w.PatientID, w.PractitionerID, w.Department,
		w.ServiceTypeCode, w.ServiceTypeDisplay, w.Priority, w.QueueNumber, w.Status,
		w.RequestedDate, w.CheckInTime, w.Note)
	return err
}

func (r *waitlistRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Waitlist, error) {
	return r.scanWaitlist(r.conn(ctx).QueryRow(ctx, `SELECT `+waitlistCols+` FROM waitlist WHERE id = $1`, id))
}

func (r *waitlistRepoPG) Update(ctx context.Context, w *Waitlist) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE waitlist SET status=$2, priority=$3, queue_number=$4,
			check_in_time=$5, called_time=$6, completed_time=$7, note=$8, updated_at=NOW()
		WHERE id = $1`,
		w.ID, w.Status, w.Priority, w.QueueNumber,
		w.CheckInTime, w.CalledTime, w.CompletedTime, w.Note)
	return err
}

func (r *waitlistRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM waitlist WHERE id = $1`, id)
	return err
}

func (r *waitlistRepoPG) ListByDepartment(ctx context.Context, department string, limit, offset int) ([]*Waitlist, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM waitlist WHERE department = $1`, department).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+waitlistCols+` FROM waitlist WHERE department = $1 ORDER BY queue_number ASC NULLS LAST LIMIT $2 OFFSET $3`, department, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Waitlist
	for rows.Next() {
		w, err := r.scanWaitlist(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, w)
	}
	return items, total, nil
}

func (r *waitlistRepoPG) ListByPractitioner(ctx context.Context, practitionerID uuid.UUID, limit, offset int) ([]*Waitlist, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM waitlist WHERE practitioner_id = $1`, practitionerID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+waitlistCols+` FROM waitlist WHERE practitioner_id = $1 ORDER BY queue_number ASC NULLS LAST LIMIT $2 OFFSET $3`, practitionerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Waitlist
	for rows.Next() {
		w, err := r.scanWaitlist(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, w)
	}
	return items, total, nil
}
