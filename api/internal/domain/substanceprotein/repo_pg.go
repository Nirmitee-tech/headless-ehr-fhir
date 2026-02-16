package substanceprotein

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ehr/ehr/internal/platform/db"
)

type queryable interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

type spRepoPG struct{ pool *pgxpool.Pool }

func NewSubstanceProteinRepoPG(pool *pgxpool.Pool) SubstanceProteinRepository {
	return &spRepoPG{pool: pool}
}

func (r *spRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil { return tx }
	if c := db.ConnFromContext(ctx); c != nil { return c }
	return r.pool
}

const spCols = `id, fhir_id, sequence_type_code, sequence_type_display, number_of_subunits, disulfide_linkage,
	version_id, created_at, updated_at`

func (r *spRepoPG) scanRow(row pgx.Row) (*SubstanceProtein, error) {
	var m SubstanceProtein
	err := row.Scan(&m.ID, &m.FHIRID, &m.SequenceTypeCode, &m.SequenceTypeDisplay, &m.NumberOfSubunits, &m.DisulfideLinkage,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *spRepoPG) Create(ctx context.Context, m *SubstanceProtein) error {
	m.ID = uuid.New()
	if m.FHIRID == "" { m.FHIRID = m.ID.String() }
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO substance_protein (id, fhir_id, sequence_type_code, sequence_type_display, number_of_subunits, disulfide_linkage)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		m.ID, m.FHIRID, m.SequenceTypeCode, m.SequenceTypeDisplay, m.NumberOfSubunits, m.DisulfideLinkage)
	return err
}

func (r *spRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*SubstanceProtein, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+spCols+` FROM substance_protein WHERE id = $1`, id)) }
func (r *spRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*SubstanceProtein, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+spCols+` FROM substance_protein WHERE fhir_id = $1`, fhirID)) }

func (r *spRepoPG) Update(ctx context.Context, m *SubstanceProtein) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE substance_protein SET sequence_type_code=$2, sequence_type_display=$3, number_of_subunits=$4, disulfide_linkage=$5, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.SequenceTypeCode, m.SequenceTypeDisplay, m.NumberOfSubunits, m.DisulfideLinkage)
	return err
}

func (r *spRepoPG) Delete(ctx context.Context, id uuid.UUID) error { _, err := r.conn(ctx).Exec(ctx, `DELETE FROM substance_protein WHERE id = $1`, id); return err }

func (r *spRepoPG) List(ctx context.Context, limit, offset int) ([]*SubstanceProtein, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM substance_protein`).Scan(&total); err != nil { return nil, 0, err }
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+spCols+` FROM substance_protein ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, err }; defer rows.Close()
	var items []*SubstanceProtein
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}

func (r *spRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SubstanceProtein, int, error) {
	query := `SELECT ` + spCols + ` FROM substance_protein WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM substance_protein WHERE 1=1`
	var args []interface{}; idx := 1
	_ = params
	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil { return nil, 0, err }
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1); args = append(args, limit, offset)
	rows, err := r.conn(ctx).Query(ctx, query, args...); if err != nil { return nil, 0, err }; defer rows.Close()
	var items []*SubstanceProtein
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}
