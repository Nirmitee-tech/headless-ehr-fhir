package medproductpackaged

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

type mppRepoPG struct{ pool *pgxpool.Pool }

func NewMedicinalProductPackagedRepoPG(pool *pgxpool.Pool) MedicinalProductPackagedRepository {
	return &mppRepoPG{pool: pool}
}

func (r *mppRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil { return tx }
	if c := db.ConnFromContext(ctx); c != nil { return c }
	return r.pool
}

const mppCols = `id, fhir_id, subject_reference, description,
	legal_status_of_supply_code, legal_status_of_supply_display,
	marketing_status_code, marketing_status_display,
	marketing_authorization_reference, manufacturer_reference,
	version_id, created_at, updated_at`

func (r *mppRepoPG) scanRow(row pgx.Row) (*MedicinalProductPackaged, error) {
	var m MedicinalProductPackaged
	err := row.Scan(&m.ID, &m.FHIRID, &m.SubjectReference, &m.Description,
		&m.LegalStatusOfSupplyCode, &m.LegalStatusOfSupplyDisplay,
		&m.MarketingStatusCode, &m.MarketingStatusDisplay,
		&m.MarketingAuthorizationReference, &m.ManufacturerReference,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *mppRepoPG) Create(ctx context.Context, m *MedicinalProductPackaged) error {
	m.ID = uuid.New()
	if m.FHIRID == "" { m.FHIRID = m.ID.String() }
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medicinal_product_packaged (id, fhir_id, subject_reference, description,
			legal_status_of_supply_code, legal_status_of_supply_display,
			marketing_status_code, marketing_status_display,
			marketing_authorization_reference, manufacturer_reference)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		m.ID, m.FHIRID, m.SubjectReference, m.Description,
		m.LegalStatusOfSupplyCode, m.LegalStatusOfSupplyDisplay,
		m.MarketingStatusCode, m.MarketingStatusDisplay,
		m.MarketingAuthorizationReference, m.ManufacturerReference)
	return err
}

func (r *mppRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductPackaged, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mppCols+` FROM medicinal_product_packaged WHERE id = $1`, id))
}

func (r *mppRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductPackaged, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mppCols+` FROM medicinal_product_packaged WHERE fhir_id = $1`, fhirID))
}

func (r *mppRepoPG) Update(ctx context.Context, m *MedicinalProductPackaged) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE medicinal_product_packaged SET subject_reference=$2, description=$3,
			legal_status_of_supply_code=$4, legal_status_of_supply_display=$5,
			marketing_status_code=$6, marketing_status_display=$7,
			marketing_authorization_reference=$8, manufacturer_reference=$9, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.SubjectReference, m.Description,
		m.LegalStatusOfSupplyCode, m.LegalStatusOfSupplyDisplay,
		m.MarketingStatusCode, m.MarketingStatusDisplay,
		m.MarketingAuthorizationReference, m.ManufacturerReference)
	return err
}

func (r *mppRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM medicinal_product_packaged WHERE id = $1`, id)
	return err
}

func (r *mppRepoPG) List(ctx context.Context, limit, offset int) ([]*MedicinalProductPackaged, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM medicinal_product_packaged`).Scan(&total); err != nil { return nil, 0, err }
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+mppCols+` FROM medicinal_product_packaged ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	var items []*MedicinalProductPackaged
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil { return nil, 0, err }
		items = append(items, m)
	}
	return items, total, nil
}

func (r *mppRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductPackaged, int, error) {
	query := `SELECT ` + mppCols + ` FROM medicinal_product_packaged WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM medicinal_product_packaged WHERE 1=1`
	var args []interface{}
	idx := 1
	if p, ok := params["subject"]; ok {
		query += fmt.Sprintf(` AND subject_reference = $%d`, idx)
		countQuery += fmt.Sprintf(` AND subject_reference = $%d`, idx)
		args = append(args, p)
		idx++
	}
	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil { return nil, 0, err }
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)
	rows, err := r.conn(ctx).Query(ctx, query, args...)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	var items []*MedicinalProductPackaged
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil { return nil, 0, err }
		items = append(items, m)
	}
	return items, total, nil
}
