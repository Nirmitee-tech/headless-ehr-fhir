package basic

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

type basicRepoPG struct{ pool *pgxpool.Pool }

func NewBasicRepoPG(pool *pgxpool.Pool) BasicRepository {
	return &basicRepoPG{pool: pool}
}

func (r *basicRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const bscCols = `id, fhir_id, code_code, code_system, code_display,
	subject_type, subject_reference, author_id, author_date,
	version_id, created_at, updated_at`

func (r *basicRepoPG) scanRow(row pgx.Row) (*Basic, error) {
	var b Basic
	err := row.Scan(&b.ID, &b.FHIRID, &b.CodeCode, &b.CodeSystem, &b.CodeDisplay,
		&b.SubjectType, &b.SubjectReference, &b.AuthorID, &b.AuthorDate,
		&b.VersionID, &b.CreatedAt, &b.UpdatedAt)
	return &b, err
}

func (r *basicRepoPG) Create(ctx context.Context, b *Basic) error {
	b.ID = uuid.New()
	if b.FHIRID == "" {
		b.FHIRID = b.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO basic (id, fhir_id, code_code, code_system, code_display,
			subject_type, subject_reference, author_id, author_date)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		b.ID, b.FHIRID, b.CodeCode, b.CodeSystem, b.CodeDisplay,
		b.SubjectType, b.SubjectReference, b.AuthorID, b.AuthorDate)
	return err
}

func (r *basicRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Basic, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+bscCols+` FROM basic WHERE id = $1`, id))
}

func (r *basicRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Basic, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+bscCols+` FROM basic WHERE fhir_id = $1`, fhirID))
}

func (r *basicRepoPG) Update(ctx context.Context, b *Basic) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE basic SET code_code=$2, code_system=$3, code_display=$4,
			subject_type=$5, subject_reference=$6, author_id=$7, author_date=$8,
			updated_at=NOW()
		WHERE id = $1`,
		b.ID, b.CodeCode, b.CodeSystem, b.CodeDisplay,
		b.SubjectType, b.SubjectReference, b.AuthorID, b.AuthorDate)
	return err
}

func (r *basicRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM basic WHERE id = $1`, id)
	return err
}

func (r *basicRepoPG) List(ctx context.Context, limit, offset int) ([]*Basic, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM basic`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+bscCols+` FROM basic ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Basic
	for rows.Next() {
		b, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, b)
	}
	return items, total, nil
}

var basicSearchParams = map[string]fhir.SearchParamConfig{
	"code":    {Type: fhir.SearchParamToken, Column: "code_code"},
	"subject": {Type: fhir.SearchParamReference, Column: "subject_reference"},
	"author":  {Type: fhir.SearchParamReference, Column: "author_id"},
}

func (r *basicRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Basic, int, error) {
	qb := fhir.NewSearchQuery("basic", bscCols)
	qb.ApplyParams(params, basicSearchParams)
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
	var items []*Basic
	for rows.Next() {
		b, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, b)
	}
	return items, total, nil
}
