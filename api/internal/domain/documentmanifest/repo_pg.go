package documentmanifest

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

type documentManifestRepoPG struct{ pool *pgxpool.Pool }

func NewDocumentManifestRepoPG(pool *pgxpool.Pool) DocumentManifestRepository {
	return &documentManifestRepoPG{pool: pool}
}

func (r *documentManifestRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const dmCols = `id, fhir_id, status, type_code, type_display, subject_reference,
	created, author_reference, recipient_reference, source_url, description,
	version_id, created_at, updated_at`

func (r *documentManifestRepoPG) scanRow(row pgx.Row) (*DocumentManifest, error) {
	var d DocumentManifest
	err := row.Scan(&d.ID, &d.FHIRID, &d.Status, &d.TypeCode, &d.TypeDisplay, &d.SubjectReference,
		&d.Created, &d.AuthorReference, &d.RecipientReference, &d.SourceURL, &d.Description,
		&d.VersionID, &d.CreatedAt, &d.UpdatedAt)
	return &d, err
}

func (r *documentManifestRepoPG) Create(ctx context.Context, d *DocumentManifest) error {
	d.ID = uuid.New()
	if d.FHIRID == "" {
		d.FHIRID = d.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO document_manifest (id, fhir_id, status, type_code, type_display, subject_reference,
			created, author_reference, recipient_reference, source_url, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		d.ID, d.FHIRID, d.Status, d.TypeCode, d.TypeDisplay, d.SubjectReference,
		d.Created, d.AuthorReference, d.RecipientReference, d.SourceURL, d.Description)
	return err
}

func (r *documentManifestRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*DocumentManifest, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+dmCols+` FROM document_manifest WHERE id = $1`, id))
}

func (r *documentManifestRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*DocumentManifest, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+dmCols+` FROM document_manifest WHERE fhir_id = $1`, fhirID))
}

func (r *documentManifestRepoPG) Update(ctx context.Context, d *DocumentManifest) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE document_manifest SET status=$2, type_code=$3, type_display=$4, subject_reference=$5,
			created=$6, author_reference=$7, recipient_reference=$8, source_url=$9, description=$10, updated_at=NOW()
		WHERE id = $1`,
		d.ID, d.Status, d.TypeCode, d.TypeDisplay, d.SubjectReference,
		d.Created, d.AuthorReference, d.RecipientReference, d.SourceURL, d.Description)
	return err
}

func (r *documentManifestRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM document_manifest WHERE id = $1`, id)
	return err
}

func (r *documentManifestRepoPG) List(ctx context.Context, limit, offset int) ([]*DocumentManifest, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM document_manifest`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+dmCols+` FROM document_manifest ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*DocumentManifest
	for rows.Next() {
		d, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}

var dmSearchParams = map[string]fhir.SearchParamConfig{
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
	"subject": {Type: fhir.SearchParamReference, Column: "subject_reference"},
	"type":    {Type: fhir.SearchParamToken, Column: "type_code"},
}

func (r *documentManifestRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*DocumentManifest, int, error) {
	qb := fhir.NewSearchQuery("document_manifest", dmCols)
	qb.ApplyParams(params, dmSearchParams)
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
	var items []*DocumentManifest
	for rows.Next() {
		d, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}
