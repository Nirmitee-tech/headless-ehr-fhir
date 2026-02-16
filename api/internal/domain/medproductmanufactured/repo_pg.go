package medproductmanufactured

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

type mpmRepoPG struct{ pool *pgxpool.Pool }

func NewMedicinalProductManufacturedRepoPG(pool *pgxpool.Pool) MedicinalProductManufacturedRepository {
	return &mpmRepoPG{pool: pool}
}

func (r *mpmRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil { return tx }
	if c := db.ConnFromContext(ctx); c != nil { return c }
	return r.pool
}

const mpmCols = `id, fhir_id, manufactured_dose_form_code, manufactured_dose_form_display,
	unit_of_presentation_code, unit_of_presentation_display,
	quantity_value, quantity_unit, manufacturer_reference, ingredient_reference,
	version_id, created_at, updated_at`

func (r *mpmRepoPG) scanRow(row pgx.Row) (*MedicinalProductManufactured, error) {
	var m MedicinalProductManufactured
	err := row.Scan(&m.ID, &m.FHIRID, &m.ManufacturedDoseFormCode, &m.ManufacturedDoseFormDisplay,
		&m.UnitOfPresentationCode, &m.UnitOfPresentationDisplay,
		&m.QuantityValue, &m.QuantityUnit, &m.ManufacturerReference, &m.IngredientReference,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *mpmRepoPG) Create(ctx context.Context, m *MedicinalProductManufactured) error {
	m.ID = uuid.New()
	if m.FHIRID == "" { m.FHIRID = m.ID.String() }
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medicinal_product_manufactured (id, fhir_id, manufactured_dose_form_code, manufactured_dose_form_display,
			unit_of_presentation_code, unit_of_presentation_display,
			quantity_value, quantity_unit, manufacturer_reference, ingredient_reference)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		m.ID, m.FHIRID, m.ManufacturedDoseFormCode, m.ManufacturedDoseFormDisplay,
		m.UnitOfPresentationCode, m.UnitOfPresentationDisplay,
		m.QuantityValue, m.QuantityUnit, m.ManufacturerReference, m.IngredientReference)
	return err
}

func (r *mpmRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductManufactured, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mpmCols+` FROM medicinal_product_manufactured WHERE id = $1`, id))
}

func (r *mpmRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductManufactured, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mpmCols+` FROM medicinal_product_manufactured WHERE fhir_id = $1`, fhirID))
}

func (r *mpmRepoPG) Update(ctx context.Context, m *MedicinalProductManufactured) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE medicinal_product_manufactured SET manufactured_dose_form_code=$2, manufactured_dose_form_display=$3,
			unit_of_presentation_code=$4, unit_of_presentation_display=$5,
			quantity_value=$6, quantity_unit=$7, manufacturer_reference=$8, ingredient_reference=$9, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.ManufacturedDoseFormCode, m.ManufacturedDoseFormDisplay,
		m.UnitOfPresentationCode, m.UnitOfPresentationDisplay,
		m.QuantityValue, m.QuantityUnit, m.ManufacturerReference, m.IngredientReference)
	return err
}

func (r *mpmRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM medicinal_product_manufactured WHERE id = $1`, id)
	return err
}

func (r *mpmRepoPG) List(ctx context.Context, limit, offset int) ([]*MedicinalProductManufactured, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM medicinal_product_manufactured`).Scan(&total); err != nil { return nil, 0, err }
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+mpmCols+` FROM medicinal_product_manufactured ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	var items []*MedicinalProductManufactured
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil { return nil, 0, err }
		items = append(items, m)
	}
	return items, total, nil
}

var mpmSearchParams = map[string]fhir.SearchParamConfig{
	"dose-form": {Type: fhir.SearchParamToken, Column: "manufactured_dose_form_code"},
}

func (r *mpmRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductManufactured, int, error) {
	qb := fhir.NewSearchQuery("medicinal_product_manufactured", mpmCols)
	qb.ApplyParams(params, mpmSearchParams)
	qb.OrderBy("created_at DESC")

	var total int
	if err := r.conn(ctx).QueryRow(ctx, qb.CountSQL(), qb.CountArgs()...).Scan(&total); err != nil { return nil, 0, err }

	rows, err := r.conn(ctx).Query(ctx, qb.DataSQL(limit, offset), qb.DataArgs(limit, offset)...)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	var items []*MedicinalProductManufactured
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil { return nil, 0, err }
		items = append(items, m)
	}
	return items, total, nil
}
