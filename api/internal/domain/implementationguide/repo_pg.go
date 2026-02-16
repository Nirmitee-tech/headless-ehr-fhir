package implementationguide

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

type implementationGuideRepoPG struct{ pool *pgxpool.Pool }

func NewImplementationGuideRepoPG(pool *pgxpool.Pool) ImplementationGuideRepository {
	return &implementationGuideRepoPG{pool: pool}
}

func (r *implementationGuideRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const igCols = `id, fhir_id, status, url, name, title, description, publisher, date,
	package_id, fhir_version, license, depends_on_uri, global_type, global_profile,
	version_id, created_at, updated_at`

func (r *implementationGuideRepoPG) scanRow(row pgx.Row) (*ImplementationGuide, error) {
	var ig ImplementationGuide
	err := row.Scan(&ig.ID, &ig.FHIRID, &ig.Status, &ig.URL, &ig.Name, &ig.Title, &ig.Description, &ig.Publisher, &ig.Date,
		&ig.PackageID, &ig.FHIRVersion, &ig.License, &ig.DependsOnURI, &ig.GlobalType, &ig.GlobalProfile,
		&ig.VersionID, &ig.CreatedAt, &ig.UpdatedAt)
	return &ig, err
}

func (r *implementationGuideRepoPG) Create(ctx context.Context, ig *ImplementationGuide) error {
	ig.ID = uuid.New()
	if ig.FHIRID == "" {
		ig.FHIRID = ig.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO implementation_guide (id, fhir_id, status, url, name, title, description, publisher, date,
			package_id, fhir_version, license, depends_on_uri, global_type, global_profile)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		ig.ID, ig.FHIRID, ig.Status, ig.URL, ig.Name, ig.Title, ig.Description, ig.Publisher, ig.Date,
		ig.PackageID, ig.FHIRVersion, ig.License, ig.DependsOnURI, ig.GlobalType, ig.GlobalProfile)
	return err
}

func (r *implementationGuideRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ImplementationGuide, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+igCols+` FROM implementation_guide WHERE id = $1`, id))
}

func (r *implementationGuideRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ImplementationGuide, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+igCols+` FROM implementation_guide WHERE fhir_id = $1`, fhirID))
}

func (r *implementationGuideRepoPG) Update(ctx context.Context, ig *ImplementationGuide) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE implementation_guide SET status=$2, url=$3, name=$4, title=$5, description=$6, publisher=$7, date=$8,
			package_id=$9, fhir_version=$10, license=$11, depends_on_uri=$12, global_type=$13, global_profile=$14, updated_at=NOW()
		WHERE id = $1`,
		ig.ID, ig.Status, ig.URL, ig.Name, ig.Title, ig.Description, ig.Publisher, ig.Date,
		ig.PackageID, ig.FHIRVersion, ig.License, ig.DependsOnURI, ig.GlobalType, ig.GlobalProfile)
	return err
}

func (r *implementationGuideRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM implementation_guide WHERE id = $1`, id)
	return err
}

func (r *implementationGuideRepoPG) List(ctx context.Context, limit, offset int) ([]*ImplementationGuide, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM implementation_guide`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+igCols+` FROM implementation_guide ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ImplementationGuide
	for rows.Next() {
		ig, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ig)
	}
	return items, total, nil
}

var implementationGuideSearchParams = map[string]fhir.SearchParamConfig{
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"url":    {Type: fhir.SearchParamToken, Column: "url"},
	"name":   {Type: fhir.SearchParamString, Column: "name"},
}

func (r *implementationGuideRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ImplementationGuide, int, error) {
	qb := fhir.NewSearchQuery("implementation_guide", igCols)
	qb.ApplyParams(params, implementationGuideSearchParams)
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
	var items []*ImplementationGuide
	for rows.Next() {
		ig, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ig)
	}
	return items, total, nil
}
