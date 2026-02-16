package medproductpharmaceutical

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

type mppRepoPG struct{ pool *pgxpool.Pool }

func NewMedicinalProductPharmaceuticalRepoPG(pool *pgxpool.Pool) MedicinalProductPharmaceuticalRepository {
	return &mppRepoPG{pool: pool}
}

func (r *mppRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil { return tx }
	if c := db.ConnFromContext(ctx); c != nil { return c }
	return r.pool
}

const mppCols = `id, fhir_id, administrable_dose_form_code, administrable_dose_form_display,
	unit_of_presentation_code, unit_of_presentation_display,
	ingredient_reference, device_reference,
	version_id, created_at, updated_at`

func (r *mppRepoPG) scanRow(row pgx.Row) (*MedicinalProductPharmaceutical, error) {
	var m MedicinalProductPharmaceutical
	err := row.Scan(&m.ID, &m.FHIRID, &m.AdministrableDoseFormCode, &m.AdministrableDoseFormDisplay,
		&m.UnitOfPresentationCode, &m.UnitOfPresentationDisplay,
		&m.IngredientReference, &m.DeviceReference,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *mppRepoPG) Create(ctx context.Context, m *MedicinalProductPharmaceutical) error {
	m.ID = uuid.New()
	if m.FHIRID == "" { m.FHIRID = m.ID.String() }
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medicinal_product_pharmaceutical (id, fhir_id, administrable_dose_form_code, administrable_dose_form_display,
			unit_of_presentation_code, unit_of_presentation_display,
			ingredient_reference, device_reference)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		m.ID, m.FHIRID, m.AdministrableDoseFormCode, m.AdministrableDoseFormDisplay,
		m.UnitOfPresentationCode, m.UnitOfPresentationDisplay,
		m.IngredientReference, m.DeviceReference)
	return err
}

func (r *mppRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductPharmaceutical, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mppCols+` FROM medicinal_product_pharmaceutical WHERE id = $1`, id)) }
func (r *mppRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductPharmaceutical, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mppCols+` FROM medicinal_product_pharmaceutical WHERE fhir_id = $1`, fhirID)) }

func (r *mppRepoPG) Update(ctx context.Context, m *MedicinalProductPharmaceutical) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE medicinal_product_pharmaceutical SET administrable_dose_form_code=$2, administrable_dose_form_display=$3,
			unit_of_presentation_code=$4, unit_of_presentation_display=$5,
			ingredient_reference=$6, device_reference=$7, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.AdministrableDoseFormCode, m.AdministrableDoseFormDisplay,
		m.UnitOfPresentationCode, m.UnitOfPresentationDisplay,
		m.IngredientReference, m.DeviceReference)
	return err
}

func (r *mppRepoPG) Delete(ctx context.Context, id uuid.UUID) error { _, err := r.conn(ctx).Exec(ctx, `DELETE FROM medicinal_product_pharmaceutical WHERE id = $1`, id); return err }

func (r *mppRepoPG) List(ctx context.Context, limit, offset int) ([]*MedicinalProductPharmaceutical, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM medicinal_product_pharmaceutical`).Scan(&total); err != nil { return nil, 0, err }
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+mppCols+` FROM medicinal_product_pharmaceutical ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, err }; defer rows.Close()
	var items []*MedicinalProductPharmaceutical
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}

var mppSearchParams = map[string]fhir.SearchParamConfig{
	"route": {Type: fhir.SearchParamToken, Column: "administrable_dose_form_code"},
}

func (r *mppRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductPharmaceutical, int, error) {
	qb := fhir.NewSearchQuery("medicinal_product_pharmaceutical", mppCols)
	qb.ApplyParams(params, mppSearchParams)
	qb.OrderBy("created_at DESC")

	var total int
	if err := r.conn(ctx).QueryRow(ctx, qb.CountSQL(), qb.CountArgs()...).Scan(&total); err != nil { return nil, 0, err }

	rows, err := r.conn(ctx).Query(ctx, qb.DataSQL(limit, offset), qb.DataArgs(limit, offset)...)
	if err != nil { return nil, 0, err }; defer rows.Close()
	var items []*MedicinalProductPharmaceutical
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}
