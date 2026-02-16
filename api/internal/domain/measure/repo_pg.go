package measure

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ehr/ehr/internal/platform/db"
	"github.com/ehr/ehr/internal/platform/fhir"
)

type queryable interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

type measureRepoPG struct{ pool *pgxpool.Pool }

func NewMeasureRepoPG(pool *pgxpool.Pool) MeasureRepository {
	return &measureRepoPG{pool: pool}
}

func (r *measureRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const measureCols = `id, fhir_id, status, url, name, title, description, publisher, date,
	effective_period_start, effective_period_end,
	scoring_code, scoring_display, subject_code, subject_display,
	approval_date, last_review_date,
	version_id, created_at, updated_at`

func (r *measureRepoPG) scanRow(row pgx.Row) (*Measure, error) {
	var m Measure
	err := row.Scan(&m.ID, &m.FHIRID, &m.Status, &m.URL, &m.Name, &m.Title, &m.Description, &m.Publisher, &m.Date,
		&m.EffectivePeriodStart, &m.EffectivePeriodEnd,
		&m.ScoringCode, &m.ScoringDisplay, &m.SubjectCode, &m.SubjectDisplay,
		&m.ApprovalDate, &m.LastReviewDate,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *measureRepoPG) Create(ctx context.Context, m *Measure) error {
	m.ID = uuid.New()
	if m.FHIRID == "" {
		m.FHIRID = m.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO measure (id, fhir_id, status, url, name, title, description, publisher, date,
			effective_period_start, effective_period_end,
			scoring_code, scoring_display, subject_code, subject_display,
			approval_date, last_review_date)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		m.ID, m.FHIRID, m.Status, m.URL, m.Name, m.Title, m.Description, m.Publisher, m.Date,
		m.EffectivePeriodStart, m.EffectivePeriodEnd,
		m.ScoringCode, m.ScoringDisplay, m.SubjectCode, m.SubjectDisplay,
		m.ApprovalDate, m.LastReviewDate)
	return err
}

func (r *measureRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Measure, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+measureCols+` FROM measure WHERE id = $1`, id))
}

func (r *measureRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Measure, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+measureCols+` FROM measure WHERE fhir_id = $1`, fhirID))
}

func (r *measureRepoPG) Update(ctx context.Context, m *Measure) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE measure SET status=$2, url=$3, name=$4, title=$5, description=$6, publisher=$7, date=$8,
			effective_period_start=$9, effective_period_end=$10,
			scoring_code=$11, scoring_display=$12, subject_code=$13, subject_display=$14,
			approval_date=$15, last_review_date=$16, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.Status, m.URL, m.Name, m.Title, m.Description, m.Publisher, m.Date,
		m.EffectivePeriodStart, m.EffectivePeriodEnd,
		m.ScoringCode, m.ScoringDisplay, m.SubjectCode, m.SubjectDisplay,
		m.ApprovalDate, m.LastReviewDate)
	return err
}

func (r *measureRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM measure WHERE id = $1`, id)
	return err
}

func (r *measureRepoPG) List(ctx context.Context, limit, offset int) ([]*Measure, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM measure`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+measureCols+` FROM measure ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Measure
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}

var measureSearchParams = map[string]fhir.SearchParamConfig{
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"name":   {Type: fhir.SearchParamString, Column: "name"},
	"title":  {Type: fhir.SearchParamString, Column: "title"},
	"url":    {Type: fhir.SearchParamToken, Column: "url"},
}

func (r *measureRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Measure, int, error) {
	qb := fhir.NewSearchQuery("measure", measureCols)
	qb.ApplyParams(params, measureSearchParams)
	qb.OrderBy("created_at DESC")

	var total int
	if err := r.conn(ctx).QueryRow(ctx, qb.CountSQL(), qb.CountArgs()...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx, qb.DataSQL(limit, offset), qb.DataArgs(limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Measure
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}
