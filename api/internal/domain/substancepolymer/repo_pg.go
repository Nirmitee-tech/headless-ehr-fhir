package substancepolymer

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

type spRepoPG struct{ pool *pgxpool.Pool }

func NewSubstancePolymerRepoPG(pool *pgxpool.Pool) SubstancePolymerRepository {
	return &spRepoPG{pool: pool}
}

func (r *spRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil { return tx }
	if c := db.ConnFromContext(ctx); c != nil { return c }
	return r.pool
}

const spCols = `id, fhir_id, class_code, class_display, geometry_code, geometry_display,
	copolymer_connectivity_code, copolymer_connectivity_display, modification,
	version_id, created_at, updated_at`

func (r *spRepoPG) scanRow(row pgx.Row) (*SubstancePolymer, error) {
	var m SubstancePolymer
	err := row.Scan(&m.ID, &m.FHIRID, &m.ClassCode, &m.ClassDisplay, &m.GeometryCode, &m.GeometryDisplay,
		&m.CopolymerConnectivityCode, &m.CopolymerConnectivityDisplay, &m.Modification,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *spRepoPG) Create(ctx context.Context, m *SubstancePolymer) error {
	m.ID = uuid.New()
	if m.FHIRID == "" { m.FHIRID = m.ID.String() }
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO substance_polymer (id, fhir_id, class_code, class_display, geometry_code, geometry_display,
			copolymer_connectivity_code, copolymer_connectivity_display, modification)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		m.ID, m.FHIRID, m.ClassCode, m.ClassDisplay, m.GeometryCode, m.GeometryDisplay,
		m.CopolymerConnectivityCode, m.CopolymerConnectivityDisplay, m.Modification)
	return err
}

func (r *spRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*SubstancePolymer, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+spCols+` FROM substance_polymer WHERE id = $1`, id)) }
func (r *spRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*SubstancePolymer, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+spCols+` FROM substance_polymer WHERE fhir_id = $1`, fhirID)) }

func (r *spRepoPG) Update(ctx context.Context, m *SubstancePolymer) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE substance_polymer SET class_code=$2, class_display=$3, geometry_code=$4, geometry_display=$5,
			copolymer_connectivity_code=$6, copolymer_connectivity_display=$7, modification=$8, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.ClassCode, m.ClassDisplay, m.GeometryCode, m.GeometryDisplay,
		m.CopolymerConnectivityCode, m.CopolymerConnectivityDisplay, m.Modification)
	return err
}

func (r *spRepoPG) Delete(ctx context.Context, id uuid.UUID) error { _, err := r.conn(ctx).Exec(ctx, `DELETE FROM substance_polymer WHERE id = $1`, id); return err }

func (r *spRepoPG) List(ctx context.Context, limit, offset int) ([]*SubstancePolymer, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM substance_polymer`).Scan(&total); err != nil { return nil, 0, err }
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+spCols+` FROM substance_polymer ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, err }; defer rows.Close()
	var items []*SubstancePolymer
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}

var spSearchParams = map[string]fhir.SearchParamConfig{
	"class": {Type: fhir.SearchParamToken, Column: "class_code"},
}

func (r *spRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SubstancePolymer, int, error) {
	qb := fhir.NewSearchQuery("substance_polymer", spCols)
	qb.ApplyParams(params, spSearchParams)
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
	var items []*SubstancePolymer
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}
