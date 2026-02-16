package terminologycapabilities

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

type terminologyCapabilitiesRepoPG struct{ pool *pgxpool.Pool }

func NewTerminologyCapabilitiesRepoPG(pool *pgxpool.Pool) TerminologyCapabilitiesRepository {
	return &terminologyCapabilitiesRepoPG{pool: pool}
}

func (r *terminologyCapabilitiesRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const tcCols = `id, fhir_id, status, url, name, title, description, publisher, date,
	kind, code_search, translation, closure, software_name, software_version,
	version_id, created_at, updated_at`

func (r *terminologyCapabilitiesRepoPG) scanRow(row pgx.Row) (*TerminologyCapabilities, error) {
	var tc TerminologyCapabilities
	err := row.Scan(&tc.ID, &tc.FHIRID, &tc.Status, &tc.URL, &tc.Name, &tc.Title, &tc.Description, &tc.Publisher, &tc.Date,
		&tc.Kind, &tc.CodeSearch, &tc.Translation, &tc.Closure, &tc.SoftwareName, &tc.SoftwareVersion,
		&tc.VersionID, &tc.CreatedAt, &tc.UpdatedAt)
	return &tc, err
}

func (r *terminologyCapabilitiesRepoPG) Create(ctx context.Context, tc *TerminologyCapabilities) error {
	tc.ID = uuid.New()
	if tc.FHIRID == "" {
		tc.FHIRID = tc.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO terminology_capabilities (id, fhir_id, status, url, name, title, description, publisher, date,
			kind, code_search, translation, closure, software_name, software_version)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		tc.ID, tc.FHIRID, tc.Status, tc.URL, tc.Name, tc.Title, tc.Description, tc.Publisher, tc.Date,
		tc.Kind, tc.CodeSearch, tc.Translation, tc.Closure, tc.SoftwareName, tc.SoftwareVersion)
	return err
}

func (r *terminologyCapabilitiesRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*TerminologyCapabilities, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+tcCols+` FROM terminology_capabilities WHERE id = $1`, id))
}

func (r *terminologyCapabilitiesRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*TerminologyCapabilities, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+tcCols+` FROM terminology_capabilities WHERE fhir_id = $1`, fhirID))
}

func (r *terminologyCapabilitiesRepoPG) Update(ctx context.Context, tc *TerminologyCapabilities) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE terminology_capabilities SET status=$2, url=$3, name=$4, title=$5, description=$6, publisher=$7, date=$8,
			kind=$9, code_search=$10, translation=$11, closure=$12, software_name=$13, software_version=$14, updated_at=NOW()
		WHERE id = $1`,
		tc.ID, tc.Status, tc.URL, tc.Name, tc.Title, tc.Description, tc.Publisher, tc.Date,
		tc.Kind, tc.CodeSearch, tc.Translation, tc.Closure, tc.SoftwareName, tc.SoftwareVersion)
	return err
}

func (r *terminologyCapabilitiesRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM terminology_capabilities WHERE id = $1`, id)
	return err
}

func (r *terminologyCapabilitiesRepoPG) List(ctx context.Context, limit, offset int) ([]*TerminologyCapabilities, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM terminology_capabilities`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+tcCols+` FROM terminology_capabilities ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*TerminologyCapabilities
	for rows.Next() {
		tc, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, tc)
	}
	return items, total, nil
}

var tcSearchParams = map[string]fhir.SearchParamConfig{
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"url":    {Type: fhir.SearchParamURI, Column: "url"},
	"name":   {Type: fhir.SearchParamString, Column: "name"},
}

func (r *terminologyCapabilitiesRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*TerminologyCapabilities, int, error) {
	qb := fhir.NewSearchQuery("terminology_capabilities", tcCols)
	qb.ApplyParams(params, tcSearchParams)
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
	var items []*TerminologyCapabilities
	for rows.Next() {
		tc, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, tc)
	}
	return items, total, nil
}
