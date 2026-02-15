package task

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

type taskRepoPG struct{ pool *pgxpool.Pool }

func NewTaskRepoPG(pool *pgxpool.Pool) TaskRepository {
	return &taskRepoPG{pool: pool}
}

func (r *taskRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const taskCols = `id, fhir_id, status, status_reason, intent, priority,
	code_value, code_display, code_system, description,
	focus_resource_type, focus_resource_id,
	for_patient_id, encounter_id, authored_on, last_modified,
	requester_id, owner_id, reason_code, reason_display, note,
	restriction_repetitions, restriction_period_start, restriction_period_end,
	input_json, output_json, created_at, updated_at`

func (r *taskRepoPG) scanTask(row pgx.Row) (*Task, error) {
	var t Task
	err := row.Scan(&t.ID, &t.FHIRID, &t.Status, &t.StatusReason, &t.Intent, &t.Priority,
		&t.CodeValue, &t.CodeDisplay, &t.CodeSystem, &t.Description,
		&t.FocusResourceType, &t.FocusResourceID,
		&t.ForPatientID, &t.EncounterID, &t.AuthoredOn, &t.LastModified,
		&t.RequesterID, &t.OwnerID, &t.ReasonCode, &t.ReasonDisplay, &t.Note,
		&t.RestrictionRepetitions, &t.RestrictionPeriodStart, &t.RestrictionPeriodEnd,
		&t.InputJSON, &t.OutputJSON, &t.CreatedAt, &t.UpdatedAt)
	return &t, err
}

func (r *taskRepoPG) Create(ctx context.Context, t *Task) error {
	t.ID = uuid.New()
	if t.FHIRID == "" {
		t.FHIRID = t.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO task (id, fhir_id, status, status_reason, intent, priority,
			code_value, code_display, code_system, description,
			focus_resource_type, focus_resource_id,
			for_patient_id, encounter_id, authored_on, last_modified,
			requester_id, owner_id, reason_code, reason_display, note,
			restriction_repetitions, restriction_period_start, restriction_period_end,
			input_json, output_json)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26)`,
		t.ID, t.FHIRID, t.Status, t.StatusReason, t.Intent, t.Priority,
		t.CodeValue, t.CodeDisplay, t.CodeSystem, t.Description,
		t.FocusResourceType, t.FocusResourceID,
		t.ForPatientID, t.EncounterID, t.AuthoredOn, t.LastModified,
		t.RequesterID, t.OwnerID, t.ReasonCode, t.ReasonDisplay, t.Note,
		t.RestrictionRepetitions, t.RestrictionPeriodStart, t.RestrictionPeriodEnd,
		t.InputJSON, t.OutputJSON)
	return err
}

func (r *taskRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	return r.scanTask(r.conn(ctx).QueryRow(ctx, `SELECT `+taskCols+` FROM task WHERE id = $1`, id))
}

func (r *taskRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Task, error) {
	return r.scanTask(r.conn(ctx).QueryRow(ctx, `SELECT `+taskCols+` FROM task WHERE fhir_id = $1`, fhirID))
}

func (r *taskRepoPG) Update(ctx context.Context, t *Task) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE task SET status=$2, status_reason=$3, intent=$4, priority=$5,
			description=$6, note=$7, owner_id=$8, last_modified=NOW(), updated_at=NOW()
		WHERE id = $1`,
		t.ID, t.Status, t.StatusReason, t.Intent, t.Priority,
		t.Description, t.Note, t.OwnerID)
	return err
}

func (r *taskRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM task WHERE id = $1`, id)
	return err
}

func (r *taskRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Task, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM task WHERE for_patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+taskCols+` FROM task WHERE for_patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Task
	for rows.Next() {
		t, err := r.scanTask(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, nil
}

func (r *taskRepoPG) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*Task, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM task WHERE owner_id = $1`, ownerID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+taskCols+` FROM task WHERE owner_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Task
	for rows.Next() {
		t, err := r.scanTask(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, nil
}

func (r *taskRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Task, int, error) {
	query := `SELECT ` + taskCols + ` FROM task WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM task WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient"]; ok {
		query += fmt.Sprintf(` AND for_patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND for_patient_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["owner"]; ok {
		query += fmt.Sprintf(` AND owner_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND owner_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["intent"]; ok {
		query += fmt.Sprintf(` AND intent = $%d`, idx)
		countQuery += fmt.Sprintf(` AND intent = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["priority"]; ok {
		query += fmt.Sprintf(` AND priority = $%d`, idx)
		countQuery += fmt.Sprintf(` AND priority = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["code"]; ok {
		query += fmt.Sprintf(` AND code_value = $%d`, idx)
		countQuery += fmt.Sprintf(` AND code_value = $%d`, idx)
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
	var items []*Task
	for rows.Next() {
		t, err := r.scanTask(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, nil
}
