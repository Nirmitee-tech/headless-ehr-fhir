package linkage

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

type linkageRepoPG struct{ pool *pgxpool.Pool }

func NewLinkageRepoPG(pool *pgxpool.Pool) LinkageRepository {
	return &linkageRepoPG{pool: pool}
}

func (r *linkageRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const lnkCols = `id, fhir_id, active, author_id, source_type, source_reference,
	alternate_type, alternate_reference,
	version_id, created_at, updated_at`

func (r *linkageRepoPG) scanRow(row pgx.Row) (*Linkage, error) {
	var l Linkage
	err := row.Scan(&l.ID, &l.FHIRID, &l.Active, &l.AuthorID, &l.SourceType, &l.SourceReference,
		&l.AlternateType, &l.AlternateReference,
		&l.VersionID, &l.CreatedAt, &l.UpdatedAt)
	return &l, err
}

func (r *linkageRepoPG) Create(ctx context.Context, l *Linkage) error {
	l.ID = uuid.New()
	if l.FHIRID == "" {
		l.FHIRID = l.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO linkage (id, fhir_id, active, author_id, source_type, source_reference,
			alternate_type, alternate_reference)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		l.ID, l.FHIRID, l.Active, l.AuthorID, l.SourceType, l.SourceReference,
		l.AlternateType, l.AlternateReference)
	return err
}

func (r *linkageRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Linkage, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+lnkCols+` FROM linkage WHERE id = $1`, id))
}

func (r *linkageRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Linkage, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+lnkCols+` FROM linkage WHERE fhir_id = $1`, fhirID))
}

func (r *linkageRepoPG) Update(ctx context.Context, l *Linkage) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE linkage SET active=$2, author_id=$3, source_type=$4, source_reference=$5,
			alternate_type=$6, alternate_reference=$7, updated_at=NOW()
		WHERE id = $1`,
		l.ID, l.Active, l.AuthorID, l.SourceType, l.SourceReference,
		l.AlternateType, l.AlternateReference)
	return err
}

func (r *linkageRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM linkage WHERE id = $1`, id)
	return err
}

func (r *linkageRepoPG) List(ctx context.Context, limit, offset int) ([]*Linkage, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM linkage`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+lnkCols+` FROM linkage ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Linkage
	for rows.Next() {
		l, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, l)
	}
	return items, total, nil
}

var linkageSearchParams = map[string]fhir.SearchParamConfig{
	"author": {Type: fhir.SearchParamReference, Column: "author_id"},
	"source": {Type: fhir.SearchParamReference, Column: "source_reference"},
}

func (r *linkageRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Linkage, int, error) {
	qb := fhir.NewSearchQuery("linkage", lnkCols)
	qb.ApplyParams(params, linkageSearchParams)
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
	var items []*Linkage
	for rows.Next() {
		l, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, l)
	}
	return items, total, nil
}
