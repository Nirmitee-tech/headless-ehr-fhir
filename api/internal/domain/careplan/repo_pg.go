package careplan

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

// =========== CarePlan Repository ===========

type carePlanRepoPG struct{ pool *pgxpool.Pool }

func NewCarePlanRepoPG(pool *pgxpool.Pool) CarePlanRepository {
	return &carePlanRepoPG{pool: pool}
}

func (r *carePlanRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const cpCols = `id, fhir_id, status, intent, category_code, category_display,
	title, description, patient_id, encounter_id, period_start, period_end,
	author_id, note, created_at, updated_at`

func (r *carePlanRepoPG) scanCP(row pgx.Row) (*CarePlan, error) {
	var cp CarePlan
	err := row.Scan(&cp.ID, &cp.FHIRID, &cp.Status, &cp.Intent,
		&cp.CategoryCode, &cp.CategoryDisplay,
		&cp.Title, &cp.Description, &cp.PatientID, &cp.EncounterID,
		&cp.PeriodStart, &cp.PeriodEnd, &cp.AuthorID, &cp.Note,
		&cp.CreatedAt, &cp.UpdatedAt)
	return &cp, err
}

func (r *carePlanRepoPG) Create(ctx context.Context, cp *CarePlan) error {
	cp.ID = uuid.New()
	if cp.FHIRID == "" {
		cp.FHIRID = cp.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO care_plan (id, fhir_id, status, intent, category_code, category_display,
			title, description, patient_id, encounter_id, period_start, period_end,
			author_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		cp.ID, cp.FHIRID, cp.Status, cp.Intent, cp.CategoryCode, cp.CategoryDisplay,
		cp.Title, cp.Description, cp.PatientID, cp.EncounterID,
		cp.PeriodStart, cp.PeriodEnd, cp.AuthorID, cp.Note)
	return err
}

func (r *carePlanRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*CarePlan, error) {
	return r.scanCP(r.conn(ctx).QueryRow(ctx, `SELECT `+cpCols+` FROM care_plan WHERE id = $1`, id))
}

func (r *carePlanRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*CarePlan, error) {
	return r.scanCP(r.conn(ctx).QueryRow(ctx, `SELECT `+cpCols+` FROM care_plan WHERE fhir_id = $1`, fhirID))
}

func (r *carePlanRepoPG) Update(ctx context.Context, cp *CarePlan) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE care_plan SET status=$2, intent=$3, title=$4, description=$5, note=$6, updated_at=NOW()
		WHERE id = $1`,
		cp.ID, cp.Status, cp.Intent, cp.Title, cp.Description, cp.Note)
	return err
}

func (r *carePlanRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM care_plan WHERE id = $1`, id)
	return err
}

func (r *carePlanRepoPG) List(ctx context.Context, limit, offset int) ([]*CarePlan, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM care_plan`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+cpCols+` FROM care_plan ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*CarePlan
	for rows.Next() {
		cp, err := r.scanCP(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, cp)
	}
	return items, total, nil
}

func (r *carePlanRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*CarePlan, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM care_plan WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+cpCols+` FROM care_plan WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*CarePlan
	for rows.Next() {
		cp, err := r.scanCP(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, cp)
	}
	return items, total, nil
}

func (r *carePlanRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*CarePlan, int, error) {
	query := `SELECT ` + cpCols + ` FROM care_plan WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM care_plan WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient"]; ok {
		query += fmt.Sprintf(` AND patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND patient_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["category"]; ok {
		query += fmt.Sprintf(` AND category_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND category_code = $%d`, idx)
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
	var items []*CarePlan
	for rows.Next() {
		cp, err := r.scanCP(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, cp)
	}
	return items, total, nil
}

func (r *carePlanRepoPG) AddActivity(ctx context.Context, a *CarePlanActivity) error {
	a.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO care_plan_activity (id, care_plan_id, detail_code, detail_display, status,
			scheduled_start, scheduled_end, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		a.ID, a.CarePlanID, a.DetailCode, a.DetailDisplay, a.Status,
		a.ScheduledStart, a.ScheduledEnd, a.Description)
	return err
}

func (r *carePlanRepoPG) GetActivities(ctx context.Context, carePlanID uuid.UUID) ([]*CarePlanActivity, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, care_plan_id, detail_code, detail_display, status,
			scheduled_start, scheduled_end, description
		FROM care_plan_activity WHERE care_plan_id = $1`, carePlanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*CarePlanActivity
	for rows.Next() {
		var a CarePlanActivity
		if err := rows.Scan(&a.ID, &a.CarePlanID, &a.DetailCode, &a.DetailDisplay,
			&a.Status, &a.ScheduledStart, &a.ScheduledEnd, &a.Description); err != nil {
			return nil, err
		}
		items = append(items, &a)
	}
	return items, nil
}

// =========== Goal Repository ===========

type goalRepoPG struct{ pool *pgxpool.Pool }

func NewGoalRepoPG(pool *pgxpool.Pool) GoalRepository {
	return &goalRepoPG{pool: pool}
}

func (r *goalRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const goalCols = `id, fhir_id, lifecycle_status, achievement_status,
	category_code, category_display, description, patient_id,
	target_measure, target_detail_string, target_due_date,
	expressed_by_id, note, created_at, updated_at`

func (r *goalRepoPG) scanGoal(row pgx.Row) (*Goal, error) {
	var g Goal
	err := row.Scan(&g.ID, &g.FHIRID, &g.LifecycleStatus, &g.AchievementStatus,
		&g.CategoryCode, &g.CategoryDisplay, &g.Description, &g.PatientID,
		&g.TargetMeasure, &g.TargetDetailString, &g.TargetDueDate,
		&g.ExpressedByID, &g.Note, &g.CreatedAt, &g.UpdatedAt)
	return &g, err
}

func (r *goalRepoPG) Create(ctx context.Context, g *Goal) error {
	g.ID = uuid.New()
	if g.FHIRID == "" {
		g.FHIRID = g.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO goal (id, fhir_id, lifecycle_status, achievement_status,
			category_code, category_display, description, patient_id,
			target_measure, target_detail_string, target_due_date,
			expressed_by_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		g.ID, g.FHIRID, g.LifecycleStatus, g.AchievementStatus,
		g.CategoryCode, g.CategoryDisplay, g.Description, g.PatientID,
		g.TargetMeasure, g.TargetDetailString, g.TargetDueDate,
		g.ExpressedByID, g.Note)
	return err
}

func (r *goalRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Goal, error) {
	return r.scanGoal(r.conn(ctx).QueryRow(ctx, `SELECT `+goalCols+` FROM goal WHERE id = $1`, id))
}

func (r *goalRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Goal, error) {
	return r.scanGoal(r.conn(ctx).QueryRow(ctx, `SELECT `+goalCols+` FROM goal WHERE fhir_id = $1`, fhirID))
}

func (r *goalRepoPG) Update(ctx context.Context, g *Goal) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE goal SET lifecycle_status=$2, achievement_status=$3, description=$4,
			target_detail_string=$5, note=$6, updated_at=NOW()
		WHERE id = $1`,
		g.ID, g.LifecycleStatus, g.AchievementStatus, g.Description,
		g.TargetDetailString, g.Note)
	return err
}

func (r *goalRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM goal WHERE id = $1`, id)
	return err
}

func (r *goalRepoPG) List(ctx context.Context, limit, offset int) ([]*Goal, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM goal`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+goalCols+` FROM goal ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Goal
	for rows.Next() {
		g, err := r.scanGoal(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, g)
	}
	return items, total, nil
}

func (r *goalRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Goal, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM goal WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+goalCols+` FROM goal WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Goal
	for rows.Next() {
		g, err := r.scanGoal(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, g)
	}
	return items, total, nil
}

func (r *goalRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Goal, int, error) {
	query := `SELECT ` + goalCols + ` FROM goal WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM goal WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient"]; ok {
		query += fmt.Sprintf(` AND patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND patient_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["lifecycle-status"]; ok {
		query += fmt.Sprintf(` AND lifecycle_status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND lifecycle_status = $%d`, idx)
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
	var items []*Goal
	for rows.Next() {
		g, err := r.scanGoal(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, g)
	}
	return items, total, nil
}
