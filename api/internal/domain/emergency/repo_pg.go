package emergency

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

// =========== Triage Repository ===========

type triageRepoPG struct{ pool *pgxpool.Pool }

func NewTriageRepoPG(pool *pgxpool.Pool) TriageRepository { return &triageRepoPG{pool: pool} }

func (r *triageRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const triageCols = `id, patient_id, encounter_id, triage_nurse_id, arrival_time, triage_time,
	chief_complaint, acuity_level, acuity_system, pain_scale, arrival_mode,
	heart_rate, blood_pressure_sys, blood_pressure_dia, temperature, respiratory_rate,
	oxygen_saturation, glasgow_coma_score, injury_description,
	allergy_note, medication_note, note, created_at, updated_at`

func (r *triageRepoPG) scanTriage(row pgx.Row) (*TriageRecord, error) {
	var t TriageRecord
	err := row.Scan(&t.ID, &t.PatientID, &t.EncounterID, &t.TriageNurseID, &t.ArrivalTime, &t.TriageTime,
		&t.ChiefComplaint, &t.AcuityLevel, &t.AcuitySystem, &t.PainScale, &t.ArrivalMode,
		&t.HeartRate, &t.BloodPressureSys, &t.BloodPressureDia, &t.Temperature, &t.RespiratoryRate,
		&t.OxygenSaturation, &t.GlasgowComaScore, &t.InjuryDescription,
		&t.AllergyNote, &t.MedicationNote, &t.Note, &t.CreatedAt, &t.UpdatedAt)
	return &t, err
}

func (r *triageRepoPG) Create(ctx context.Context, t *TriageRecord) error {
	t.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO triage_record (id, patient_id, encounter_id, triage_nurse_id, arrival_time, triage_time,
			chief_complaint, acuity_level, acuity_system, pain_scale, arrival_mode,
			heart_rate, blood_pressure_sys, blood_pressure_dia, temperature, respiratory_rate,
			oxygen_saturation, glasgow_coma_score, injury_description,
			allergy_note, medication_note, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`,
		t.ID, t.PatientID, t.EncounterID, t.TriageNurseID, t.ArrivalTime, t.TriageTime,
		t.ChiefComplaint, t.AcuityLevel, t.AcuitySystem, t.PainScale, t.ArrivalMode,
		t.HeartRate, t.BloodPressureSys, t.BloodPressureDia, t.Temperature, t.RespiratoryRate,
		t.OxygenSaturation, t.GlasgowComaScore, t.InjuryDescription,
		t.AllergyNote, t.MedicationNote, t.Note)
	return err
}

func (r *triageRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*TriageRecord, error) {
	return r.scanTriage(r.conn(ctx).QueryRow(ctx, `SELECT `+triageCols+` FROM triage_record WHERE id = $1`, id))
}

func (r *triageRepoPG) Update(ctx context.Context, t *TriageRecord) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE triage_record SET chief_complaint=$2, acuity_level=$3, acuity_system=$4,
			pain_scale=$5, heart_rate=$6, blood_pressure_sys=$7, blood_pressure_dia=$8,
			temperature=$9, respiratory_rate=$10, oxygen_saturation=$11,
			glasgow_coma_score=$12, injury_description=$13, note=$14, updated_at=NOW()
		WHERE id = $1`,
		t.ID, t.ChiefComplaint, t.AcuityLevel, t.AcuitySystem,
		t.PainScale, t.HeartRate, t.BloodPressureSys, t.BloodPressureDia,
		t.Temperature, t.RespiratoryRate, t.OxygenSaturation,
		t.GlasgowComaScore, t.InjuryDescription, t.Note)
	return err
}

func (r *triageRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM triage_record WHERE id = $1`, id)
	return err
}

func (r *triageRepoPG) List(ctx context.Context, limit, offset int) ([]*TriageRecord, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM triage_record`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+triageCols+` FROM triage_record ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*TriageRecord
	for rows.Next() {
		t, err := r.scanTriage(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, nil
}

func (r *triageRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*TriageRecord, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM triage_record WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+triageCols+` FROM triage_record WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*TriageRecord
	for rows.Next() {
		t, err := r.scanTriage(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, nil
}

func (r *triageRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*TriageRecord, int, error) {
	query := `SELECT ` + triageCols + ` FROM triage_record WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM triage_record WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient_id"]; ok {
		query += fmt.Sprintf(` AND patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND patient_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["encounter_id"]; ok {
		query += fmt.Sprintf(` AND encounter_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND encounter_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["acuity_level"]; ok {
		query += fmt.Sprintf(` AND acuity_level = $%d`, idx)
		countQuery += fmt.Sprintf(` AND acuity_level = $%d`, idx)
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
	var items []*TriageRecord
	for rows.Next() {
		t, err := r.scanTriage(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, nil
}

// =========== ED Tracking Repository ===========

type edTrackingRepoPG struct{ pool *pgxpool.Pool }

func NewEDTrackingRepoPG(pool *pgxpool.Pool) EDTrackingRepository {
	return &edTrackingRepoPG{pool: pool}
}

func (r *edTrackingRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const edTrackCols = `id, patient_id, encounter_id, triage_record_id, current_status,
	bed_assignment, attending_id, nurse_id, arrival_time, discharge_time,
	disposition, disposition_dest, length_of_stay_mins, note, created_at, updated_at`

func (r *edTrackingRepoPG) scanEDTracking(row pgx.Row) (*EDTracking, error) {
	var t EDTracking
	err := row.Scan(&t.ID, &t.PatientID, &t.EncounterID, &t.TriageRecordID, &t.CurrentStatus,
		&t.BedAssignment, &t.AttendingID, &t.NurseID, &t.ArrivalTime, &t.DischargeTime,
		&t.Disposition, &t.DispositionDest, &t.LengthOfStayMins, &t.Note, &t.CreatedAt, &t.UpdatedAt)
	return &t, err
}

func (r *edTrackingRepoPG) Create(ctx context.Context, t *EDTracking) error {
	t.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO ed_tracking (id, patient_id, encounter_id, triage_record_id, current_status,
			bed_assignment, attending_id, nurse_id, arrival_time, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		t.ID, t.PatientID, t.EncounterID, t.TriageRecordID, t.CurrentStatus,
		t.BedAssignment, t.AttendingID, t.NurseID, t.ArrivalTime, t.Note)
	return err
}

func (r *edTrackingRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*EDTracking, error) {
	return r.scanEDTracking(r.conn(ctx).QueryRow(ctx, `SELECT `+edTrackCols+` FROM ed_tracking WHERE id = $1`, id))
}

func (r *edTrackingRepoPG) Update(ctx context.Context, t *EDTracking) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE ed_tracking SET current_status=$2, bed_assignment=$3, attending_id=$4, nurse_id=$5,
			discharge_time=$6, disposition=$7, disposition_dest=$8, length_of_stay_mins=$9,
			note=$10, updated_at=NOW()
		WHERE id = $1`,
		t.ID, t.CurrentStatus, t.BedAssignment, t.AttendingID, t.NurseID,
		t.DischargeTime, t.Disposition, t.DispositionDest, t.LengthOfStayMins, t.Note)
	return err
}

func (r *edTrackingRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM ed_tracking WHERE id = $1`, id)
	return err
}

func (r *edTrackingRepoPG) List(ctx context.Context, limit, offset int) ([]*EDTracking, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM ed_tracking`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+edTrackCols+` FROM ed_tracking ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*EDTracking
	for rows.Next() {
		t, err := r.scanEDTracking(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, nil
}

func (r *edTrackingRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*EDTracking, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM ed_tracking WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+edTrackCols+` FROM ed_tracking WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*EDTracking
	for rows.Next() {
		t, err := r.scanEDTracking(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, nil
}

func (r *edTrackingRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*EDTracking, int, error) {
	query := `SELECT ` + edTrackCols + ` FROM ed_tracking WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM ed_tracking WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient_id"]; ok {
		query += fmt.Sprintf(` AND patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND patient_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["current_status"]; ok {
		query += fmt.Sprintf(` AND current_status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND current_status = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["attending_id"]; ok {
		query += fmt.Sprintf(` AND attending_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND attending_id = $%d`, idx)
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
	var items []*EDTracking
	for rows.Next() {
		t, err := r.scanEDTracking(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, nil
}

// -- Status History --

func (r *edTrackingRepoPG) AddStatusHistory(ctx context.Context, h *EDStatusHistory) error {
	h.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO ed_status_history (id, ed_tracking_id, status, changed_at, changed_by, note)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		h.ID, h.EDTrackingID, h.Status, h.ChangedAt, h.ChangedBy, h.Note)
	return err
}

func (r *edTrackingRepoPG) GetStatusHistory(ctx context.Context, trackingID uuid.UUID) ([]*EDStatusHistory, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, ed_tracking_id, status, changed_at, changed_by, note
		FROM ed_status_history WHERE ed_tracking_id = $1 ORDER BY changed_at`, trackingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*EDStatusHistory
	for rows.Next() {
		var h EDStatusHistory
		if err := rows.Scan(&h.ID, &h.EDTrackingID, &h.Status, &h.ChangedAt, &h.ChangedBy, &h.Note); err != nil {
			return nil, err
		}
		items = append(items, &h)
	}
	return items, nil
}

// =========== Trauma Repository ===========

type traumaRepoPG struct{ pool *pgxpool.Pool }

func NewTraumaRepoPG(pool *pgxpool.Pool) TraumaRepository { return &traumaRepoPG{pool: pool} }

func (r *traumaRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const traumaCols = `id, patient_id, encounter_id, ed_tracking_id, activation_level,
	activation_time, deactivation_time, mechanism_of_injury, activated_by, team_lead_id,
	outcome, note, created_at, updated_at`

func (r *traumaRepoPG) scanTrauma(row pgx.Row) (*TraumaActivation, error) {
	var t TraumaActivation
	err := row.Scan(&t.ID, &t.PatientID, &t.EncounterID, &t.EDTrackingID, &t.ActivationLevel,
		&t.ActivationTime, &t.DeactivationTime, &t.MechanismOfInjury, &t.ActivatedBy, &t.TeamLeadID,
		&t.Outcome, &t.Note, &t.CreatedAt, &t.UpdatedAt)
	return &t, err
}

func (r *traumaRepoPG) Create(ctx context.Context, t *TraumaActivation) error {
	t.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO trauma_activation (id, patient_id, encounter_id, ed_tracking_id, activation_level,
			activation_time, mechanism_of_injury, activated_by, team_lead_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		t.ID, t.PatientID, t.EncounterID, t.EDTrackingID, t.ActivationLevel,
		t.ActivationTime, t.MechanismOfInjury, t.ActivatedBy, t.TeamLeadID, t.Note)
	return err
}

func (r *traumaRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*TraumaActivation, error) {
	return r.scanTrauma(r.conn(ctx).QueryRow(ctx, `SELECT `+traumaCols+` FROM trauma_activation WHERE id = $1`, id))
}

func (r *traumaRepoPG) Update(ctx context.Context, t *TraumaActivation) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE trauma_activation SET activation_level=$2, deactivation_time=$3,
			mechanism_of_injury=$4, team_lead_id=$5, outcome=$6, note=$7, updated_at=NOW()
		WHERE id = $1`,
		t.ID, t.ActivationLevel, t.DeactivationTime,
		t.MechanismOfInjury, t.TeamLeadID, t.Outcome, t.Note)
	return err
}

func (r *traumaRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM trauma_activation WHERE id = $1`, id)
	return err
}

func (r *traumaRepoPG) List(ctx context.Context, limit, offset int) ([]*TraumaActivation, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM trauma_activation`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+traumaCols+` FROM trauma_activation ORDER BY activation_time DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*TraumaActivation
	for rows.Next() {
		t, err := r.scanTrauma(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, nil
}

func (r *traumaRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*TraumaActivation, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM trauma_activation WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+traumaCols+` FROM trauma_activation WHERE patient_id = $1 ORDER BY activation_time DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*TraumaActivation
	for rows.Next() {
		t, err := r.scanTrauma(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, nil
}

func (r *traumaRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*TraumaActivation, int, error) {
	query := `SELECT ` + traumaCols + ` FROM trauma_activation WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM trauma_activation WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient_id"]; ok {
		query += fmt.Sprintf(` AND patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND patient_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["activation_level"]; ok {
		query += fmt.Sprintf(` AND activation_level = $%d`, idx)
		countQuery += fmt.Sprintf(` AND activation_level = $%d`, idx)
		args = append(args, p)
		idx++
	}

	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(` ORDER BY activation_time DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*TraumaActivation
	for rows.Next() {
		t, err := r.scanTrauma(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, nil
}
