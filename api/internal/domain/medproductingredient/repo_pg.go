package medproductingredient

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

func NewMedicinalProductIngredientRepoPG(pool *pgxpool.Pool) MedicinalProductIngredientRepository {
	return &mpiRepoPG{pool: pool}
}

func (r *mpiRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const mpiCols = `id, fhir_id, role_code, role_display, allergenic_indicator,
	substance_code, substance_display, strength_numerator_value, strength_numerator_unit,
	strength_denominator_value, strength_denominator_unit, manufacturer_reference,
	version_id, created_at, updated_at`

func (r *mpiRepoPG) scanRow(row pgx.Row) (*MedicinalProductIngredient, error) {
	var m MedicinalProductIngredient
	err := row.Scan(&m.ID, &m.FHIRID, &m.RoleCode, &m.RoleDisplay, &m.AllergenicIndicator,
		&m.SubstanceCode, &m.SubstanceDisplay, &m.StrengthNumeratorValue, &m.StrengthNumeratorUnit,
		&m.StrengthDenominatorValue, &m.StrengthDenominatorUnit, &m.ManufacturerReference,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *mpiRepoPG) Create(ctx context.Context, m *MedicinalProductIngredient) error {
	m.ID = uuid.New()
	if m.FHIRID == "" {
		m.FHIRID = m.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medicinal_product_ingredient (id, fhir_id, role_code, role_display, allergenic_indicator,
			substance_code, substance_display, strength_numerator_value, strength_numerator_unit,
			strength_denominator_value, strength_denominator_unit, manufacturer_reference)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		m.ID, m.FHIRID, m.RoleCode, m.RoleDisplay, m.AllergenicIndicator,
		m.SubstanceCode, m.SubstanceDisplay, m.StrengthNumeratorValue, m.StrengthNumeratorUnit,
		m.StrengthDenominatorValue, m.StrengthDenominatorUnit, m.ManufacturerReference)
	return err
}

func (r *mpiRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductIngredient, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mpiCols+` FROM medicinal_product_ingredient WHERE id = $1`, id))
}

func (r *mpiRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductIngredient, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mpiCols+` FROM medicinal_product_ingredient WHERE fhir_id = $1`, fhirID))
}

func (r *mpiRepoPG) Update(ctx context.Context, m *MedicinalProductIngredient) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE medicinal_product_ingredient SET role_code=$2, role_display=$3, allergenic_indicator=$4,
			substance_code=$5, substance_display=$6, strength_numerator_value=$7, strength_numerator_unit=$8,
			strength_denominator_value=$9, strength_denominator_unit=$10, manufacturer_reference=$11, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.RoleCode, m.RoleDisplay, m.AllergenicIndicator,
		m.SubstanceCode, m.SubstanceDisplay, m.StrengthNumeratorValue, m.StrengthNumeratorUnit,
		m.StrengthDenominatorValue, m.StrengthDenominatorUnit, m.ManufacturerReference)
	return err
}

func (r *mpiRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM medicinal_product_ingredient WHERE id = $1`, id)
	return err
}

func (r *mpiRepoPG) List(ctx context.Context, limit, offset int) ([]*MedicinalProductIngredient, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM medicinal_product_ingredient`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+mpiCols+` FROM medicinal_product_ingredient ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MedicinalProductIngredient
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}

func (r *mpiRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductIngredient, int, error) {
	query := `SELECT ` + mpiCols + ` FROM medicinal_product_ingredient WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM medicinal_product_ingredient WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["role"]; ok {
		query += fmt.Sprintf(` AND role_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND role_code = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["substance"]; ok {
		query += fmt.Sprintf(` AND substance_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND substance_code = $%d`, idx)
		args = append(args, p)
		idx++
	}

	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)
	rows, err := r.conn(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MedicinalProductIngredient
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}
