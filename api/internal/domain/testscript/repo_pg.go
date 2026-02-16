package testscript

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

type testScriptRepoPG struct{ pool *pgxpool.Pool }

func NewTestScriptRepoPG(pool *pgxpool.Pool) TestScriptRepository {
	return &testScriptRepoPG{pool: pool}
}

func (r *testScriptRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const tsCols = `id, fhir_id, status, url, name, title, description, publisher, date,
	purpose, copyright, profile_reference, origin_index, destination_index,
	version_id, created_at, updated_at`

func (r *testScriptRepoPG) scanRow(row pgx.Row) (*TestScript, error) {
	var ts TestScript
	err := row.Scan(&ts.ID, &ts.FHIRID, &ts.Status, &ts.URL, &ts.Name, &ts.Title, &ts.Description, &ts.Publisher, &ts.Date,
		&ts.Purpose, &ts.Copyright, &ts.ProfileReference, &ts.OriginIndex, &ts.DestinationIndex,
		&ts.VersionID, &ts.CreatedAt, &ts.UpdatedAt)
	return &ts, err
}

func (r *testScriptRepoPG) Create(ctx context.Context, ts *TestScript) error {
	ts.ID = uuid.New()
	if ts.FHIRID == "" {
		ts.FHIRID = ts.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO test_script (id, fhir_id, status, url, name, title, description, publisher, date,
			purpose, copyright, profile_reference, origin_index, destination_index)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		ts.ID, ts.FHIRID, ts.Status, ts.URL, ts.Name, ts.Title, ts.Description, ts.Publisher, ts.Date,
		ts.Purpose, ts.Copyright, ts.ProfileReference, ts.OriginIndex, ts.DestinationIndex)
	return err
}

func (r *testScriptRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*TestScript, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+tsCols+` FROM test_script WHERE id = $1`, id))
}

func (r *testScriptRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*TestScript, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+tsCols+` FROM test_script WHERE fhir_id = $1`, fhirID))
}

func (r *testScriptRepoPG) Update(ctx context.Context, ts *TestScript) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE test_script SET status=$2, url=$3, name=$4, title=$5, description=$6, publisher=$7, date=$8,
			purpose=$9, copyright=$10, profile_reference=$11, origin_index=$12, destination_index=$13, updated_at=NOW()
		WHERE id = $1`,
		ts.ID, ts.Status, ts.URL, ts.Name, ts.Title, ts.Description, ts.Publisher, ts.Date,
		ts.Purpose, ts.Copyright, ts.ProfileReference, ts.OriginIndex, ts.DestinationIndex)
	return err
}

func (r *testScriptRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM test_script WHERE id = $1`, id)
	return err
}

func (r *testScriptRepoPG) List(ctx context.Context, limit, offset int) ([]*TestScript, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM test_script`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+tsCols+` FROM test_script ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*TestScript
	for rows.Next() {
		ts, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ts)
	}
	return items, total, nil
}

var tsSearchParams = map[string]fhir.SearchParamConfig{
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"url":    {Type: fhir.SearchParamURI, Column: "url"},
	"name":   {Type: fhir.SearchParamString, Column: "name"},
}

func (r *testScriptRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*TestScript, int, error) {
	qb := fhir.NewSearchQuery("test_script", tsCols)
	qb.ApplyParams(params, tsSearchParams)
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
	var items []*TestScript
	for rows.Next() {
		ts, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ts)
	}
	return items, total, nil
}
