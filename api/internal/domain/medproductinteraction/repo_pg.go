package medproductinteraction

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

type mpiRepoPG struct{ pool *pgxpool.Pool }

func NewMedicinalProductInteractionRepoPG(pool *pgxpool.Pool) MedicinalProductInteractionRepository {
	return &mpiRepoPG{pool: pool}
}

func (r *mpiRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil { return tx }
	if c := db.ConnFromContext(ctx); c != nil { return c }
	return r.pool
}

const mpiCols = `id, fhir_id, subject_reference, description, type_code, type_display,
	effect_code, effect_display, incidence_code, incidence_display,
	management_code, management_display,
	version_id, created_at, updated_at`

func (r *mpiRepoPG) scanRow(row pgx.Row) (*MedicinalProductInteraction, error) {
	var m MedicinalProductInteraction
	err := row.Scan(&m.ID, &m.FHIRID, &m.SubjectReference, &m.Description, &m.TypeCode, &m.TypeDisplay,
		&m.EffectCode, &m.EffectDisplay, &m.IncidenceCode, &m.IncidenceDisplay,
		&m.ManagementCode, &m.ManagementDisplay,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *mpiRepoPG) Create(ctx context.Context, m *MedicinalProductInteraction) error {
	m.ID = uuid.New()
	if m.FHIRID == "" { m.FHIRID = m.ID.String() }
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medicinal_product_interaction (id, fhir_id, subject_reference, description, type_code, type_display,
			effect_code, effect_display, incidence_code, incidence_display,
			management_code, management_display)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		m.ID, m.FHIRID, m.SubjectReference, m.Description, m.TypeCode, m.TypeDisplay,
		m.EffectCode, m.EffectDisplay, m.IncidenceCode, m.IncidenceDisplay,
		m.ManagementCode, m.ManagementDisplay)
	return err
}

func (r *mpiRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductInteraction, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mpiCols+` FROM medicinal_product_interaction WHERE id = $1`, id)) }
func (r *mpiRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductInteraction, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mpiCols+` FROM medicinal_product_interaction WHERE fhir_id = $1`, fhirID)) }

func (r *mpiRepoPG) Update(ctx context.Context, m *MedicinalProductInteraction) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE medicinal_product_interaction SET subject_reference=$2, description=$3, type_code=$4, type_display=$5,
			effect_code=$6, effect_display=$7, incidence_code=$8, incidence_display=$9,
			management_code=$10, management_display=$11, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.SubjectReference, m.Description, m.TypeCode, m.TypeDisplay,
		m.EffectCode, m.EffectDisplay, m.IncidenceCode, m.IncidenceDisplay,
		m.ManagementCode, m.ManagementDisplay)
	return err
}

func (r *mpiRepoPG) Delete(ctx context.Context, id uuid.UUID) error { _, err := r.conn(ctx).Exec(ctx, `DELETE FROM medicinal_product_interaction WHERE id = $1`, id); return err }

func (r *mpiRepoPG) List(ctx context.Context, limit, offset int) ([]*MedicinalProductInteraction, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM medicinal_product_interaction`).Scan(&total); err != nil { return nil, 0, err }
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+mpiCols+` FROM medicinal_product_interaction ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, err }; defer rows.Close()
	var items []*MedicinalProductInteraction
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}

func (r *mpiRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductInteraction, int, error) {
	query := `SELECT ` + mpiCols + ` FROM medicinal_product_interaction WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM medicinal_product_interaction WHERE 1=1`
	var args []interface{}; idx := 1
	if p, ok := params["subject"]; ok { query += fmt.Sprintf(` AND subject_reference = $%d`, idx); countQuery += fmt.Sprintf(` AND subject_reference = $%d`, idx); args = append(args, p); idx++ }
	if p, ok := params["type"]; ok { query += fmt.Sprintf(` AND type_code = $%d`, idx); countQuery += fmt.Sprintf(` AND type_code = $%d`, idx); args = append(args, p); idx++ }
	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil { return nil, 0, err }
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1); args = append(args, limit, offset)
	rows, err := r.conn(ctx).Query(ctx, query, args...); if err != nil { return nil, 0, err }; defer rows.Close()
	var items []*MedicinalProductInteraction
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}
