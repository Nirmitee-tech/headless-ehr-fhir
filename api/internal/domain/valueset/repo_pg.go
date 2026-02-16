package valueset

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

type valueSetRepoPG struct{ pool *pgxpool.Pool }

func NewValueSetRepoPG(pool *pgxpool.Pool) ValueSetRepository {
	return &valueSetRepoPG{pool: pool}
}

func (r *valueSetRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const vsCols = `id, fhir_id, status, url, name, title, description, publisher, date,
	immutable, purpose, copyright, compose_include_system, compose_include_version,
	expansion_identifier, expansion_timestamp,
	version_id, created_at, updated_at`

func (r *valueSetRepoPG) scanRow(row pgx.Row) (*ValueSet, error) {
	var vs ValueSet
	err := row.Scan(&vs.ID, &vs.FHIRID, &vs.Status, &vs.URL, &vs.Name, &vs.Title,
		&vs.Description, &vs.Publisher, &vs.Date,
		&vs.Immutable, &vs.Purpose, &vs.Copyright,
		&vs.ComposeIncludeSystem, &vs.ComposeIncludeVersion,
		&vs.ExpansionIdentifier, &vs.ExpansionTimestamp,
		&vs.VersionID, &vs.CreatedAt, &vs.UpdatedAt)
	return &vs, err
}

func (r *valueSetRepoPG) Create(ctx context.Context, vs *ValueSet) error {
	vs.ID = uuid.New()
	if vs.FHIRID == "" {
		vs.FHIRID = vs.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO value_set (id, fhir_id, status, url, name, title, description, publisher, date,
			immutable, purpose, copyright, compose_include_system, compose_include_version,
			expansion_identifier, expansion_timestamp)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		vs.ID, vs.FHIRID, vs.Status, vs.URL, vs.Name, vs.Title,
		vs.Description, vs.Publisher, vs.Date,
		vs.Immutable, vs.Purpose, vs.Copyright,
		vs.ComposeIncludeSystem, vs.ComposeIncludeVersion,
		vs.ExpansionIdentifier, vs.ExpansionTimestamp)
	return err
}

func (r *valueSetRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ValueSet, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+vsCols+` FROM value_set WHERE id = $1`, id))
}

func (r *valueSetRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ValueSet, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+vsCols+` FROM value_set WHERE fhir_id = $1`, fhirID))
}

func (r *valueSetRepoPG) Update(ctx context.Context, vs *ValueSet) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE value_set SET status=$2, url=$3, name=$4, title=$5, description=$6,
			publisher=$7, date=$8, immutable=$9, purpose=$10, copyright=$11,
			compose_include_system=$12, compose_include_version=$13,
			expansion_identifier=$14, expansion_timestamp=$15, updated_at=NOW()
		WHERE id = $1`,
		vs.ID, vs.Status, vs.URL, vs.Name, vs.Title, vs.Description,
		vs.Publisher, vs.Date, vs.Immutable, vs.Purpose, vs.Copyright,
		vs.ComposeIncludeSystem, vs.ComposeIncludeVersion,
		vs.ExpansionIdentifier, vs.ExpansionTimestamp)
	return err
}

func (r *valueSetRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM value_set WHERE id = $1`, id)
	return err
}

func (r *valueSetRepoPG) List(ctx context.Context, limit, offset int) ([]*ValueSet, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM value_set`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+vsCols+` FROM value_set ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ValueSet
	for rows.Next() {
		vs, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, vs)
	}
	return items, total, nil
}

var vsSearchParams = map[string]fhir.SearchParamConfig{
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"url":    {Type: fhir.SearchParamURI, Column: "url"},
	"name":   {Type: fhir.SearchParamString, Column: "name"},
	"title":  {Type: fhir.SearchParamString, Column: "title"},
}

func (r *valueSetRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ValueSet, int, error) {
	qb := fhir.NewSearchQuery("value_set", vsCols)
	qb.ApplyParams(params, vsSearchParams)
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
	var items []*ValueSet
	for rows.Next() {
		vs, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, vs)
	}
	return items, total, nil
}
