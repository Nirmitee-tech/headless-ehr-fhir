package substancenucleicacid

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

type snaRepoPG struct{ pool *pgxpool.Pool }

func NewSubstanceNucleicAcidRepoPG(pool *pgxpool.Pool) SubstanceNucleicAcidRepository {
	return &snaRepoPG{pool: pool}
}

func (r *snaRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil { return tx }
	if c := db.ConnFromContext(ctx); c != nil { return c }
	return r.pool
}

const snaCols = `id, fhir_id, sequence_type_code, sequence_type_display, number_of_subunits, area_of_hybridisation,
	oligo_nucleotide_type_code, oligo_nucleotide_type_display,
	version_id, created_at, updated_at`

func (r *snaRepoPG) scanRow(row pgx.Row) (*SubstanceNucleicAcid, error) {
	var m SubstanceNucleicAcid
	err := row.Scan(&m.ID, &m.FHIRID, &m.SequenceTypeCode, &m.SequenceTypeDisplay, &m.NumberOfSubunits, &m.AreaOfHybridisation,
		&m.OligoNucleotideTypeCode, &m.OligoNucleotideTypeDisplay,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *snaRepoPG) Create(ctx context.Context, m *SubstanceNucleicAcid) error {
	m.ID = uuid.New()
	if m.FHIRID == "" { m.FHIRID = m.ID.String() }
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO substance_nucleic_acid (id, fhir_id, sequence_type_code, sequence_type_display, number_of_subunits, area_of_hybridisation,
			oligo_nucleotide_type_code, oligo_nucleotide_type_display)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		m.ID, m.FHIRID, m.SequenceTypeCode, m.SequenceTypeDisplay, m.NumberOfSubunits, m.AreaOfHybridisation,
		m.OligoNucleotideTypeCode, m.OligoNucleotideTypeDisplay)
	return err
}

func (r *snaRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*SubstanceNucleicAcid, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+snaCols+` FROM substance_nucleic_acid WHERE id = $1`, id)) }
func (r *snaRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*SubstanceNucleicAcid, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+snaCols+` FROM substance_nucleic_acid WHERE fhir_id = $1`, fhirID)) }

func (r *snaRepoPG) Update(ctx context.Context, m *SubstanceNucleicAcid) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE substance_nucleic_acid SET sequence_type_code=$2, sequence_type_display=$3, number_of_subunits=$4, area_of_hybridisation=$5,
			oligo_nucleotide_type_code=$6, oligo_nucleotide_type_display=$7, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.SequenceTypeCode, m.SequenceTypeDisplay, m.NumberOfSubunits, m.AreaOfHybridisation,
		m.OligoNucleotideTypeCode, m.OligoNucleotideTypeDisplay)
	return err
}

func (r *snaRepoPG) Delete(ctx context.Context, id uuid.UUID) error { _, err := r.conn(ctx).Exec(ctx, `DELETE FROM substance_nucleic_acid WHERE id = $1`, id); return err }

func (r *snaRepoPG) List(ctx context.Context, limit, offset int) ([]*SubstanceNucleicAcid, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM substance_nucleic_acid`).Scan(&total); err != nil { return nil, 0, err }
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+snaCols+` FROM substance_nucleic_acid ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, err }; defer rows.Close()
	var items []*SubstanceNucleicAcid
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}

func (r *snaRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SubstanceNucleicAcid, int, error) {
	query := `SELECT ` + snaCols + ` FROM substance_nucleic_acid WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM substance_nucleic_acid WHERE 1=1`
	var args []interface{}; idx := 1
	_ = params
	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil { return nil, 0, err }
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1); args = append(args, limit, offset)
	rows, err := r.conn(ctx).Query(ctx, query, args...); if err != nil { return nil, 0, err }; defer rows.Close()
	var items []*SubstanceNucleicAcid
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}
