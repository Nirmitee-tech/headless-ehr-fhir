package examplescenario

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

type exampleScenarioRepoPG struct{ pool *pgxpool.Pool }

func NewExampleScenarioRepoPG(pool *pgxpool.Pool) ExampleScenarioRepository {
	return &exampleScenarioRepoPG{pool: pool}
}

func (r *exampleScenarioRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const esCols = `id, fhir_id, status, url, name, title, description, publisher, date,
	purpose, copyright,
	version_id, created_at, updated_at`

func (r *exampleScenarioRepoPG) scanRow(row pgx.Row) (*ExampleScenario, error) {
	var e ExampleScenario
	err := row.Scan(&e.ID, &e.FHIRID, &e.Status, &e.URL, &e.Name, &e.Title, &e.Description, &e.Publisher, &e.Date,
		&e.Purpose, &e.Copyright,
		&e.VersionID, &e.CreatedAt, &e.UpdatedAt)
	return &e, err
}

func (r *exampleScenarioRepoPG) Create(ctx context.Context, e *ExampleScenario) error {
	e.ID = uuid.New()
	if e.FHIRID == "" {
		e.FHIRID = e.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO example_scenario (id, fhir_id, status, url, name, title, description, publisher, date,
			purpose, copyright)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		e.ID, e.FHIRID, e.Status, e.URL, e.Name, e.Title, e.Description, e.Publisher, e.Date,
		e.Purpose, e.Copyright)
	return err
}

func (r *exampleScenarioRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ExampleScenario, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+esCols+` FROM example_scenario WHERE id = $1`, id))
}

func (r *exampleScenarioRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ExampleScenario, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+esCols+` FROM example_scenario WHERE fhir_id = $1`, fhirID))
}

func (r *exampleScenarioRepoPG) Update(ctx context.Context, e *ExampleScenario) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE example_scenario SET status=$2, url=$3, name=$4, title=$5, description=$6, publisher=$7, date=$8,
			purpose=$9, copyright=$10, updated_at=NOW()
		WHERE id = $1`,
		e.ID, e.Status, e.URL, e.Name, e.Title, e.Description, e.Publisher, e.Date,
		e.Purpose, e.Copyright)
	return err
}

func (r *exampleScenarioRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM example_scenario WHERE id = $1`, id)
	return err
}

func (r *exampleScenarioRepoPG) List(ctx context.Context, limit, offset int) ([]*ExampleScenario, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM example_scenario`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+esCols+` FROM example_scenario ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ExampleScenario
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

var exampleScenarioSearchParams = map[string]fhir.SearchParamConfig{
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"url":    {Type: fhir.SearchParamToken, Column: "url"},
	"name":   {Type: fhir.SearchParamString, Column: "name"},
}

func (r *exampleScenarioRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ExampleScenario, int, error) {
	qb := fhir.NewSearchQuery("example_scenario", esCols)
	qb.ApplyParams(params, exampleScenarioSearchParams)
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
	var items []*ExampleScenario
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}
