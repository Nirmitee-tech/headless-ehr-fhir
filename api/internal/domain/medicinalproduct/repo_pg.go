package medicinalproduct

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

type medicinalProductRepoPG struct{ pool *pgxpool.Pool }

func NewMedicinalProductRepoPG(pool *pgxpool.Pool) MedicinalProductRepository {
	return &medicinalProductRepoPG{pool: pool}
}

func (r *medicinalProductRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const mpCols = `id, fhir_id, status, type_code, type_display, domain_code, domain_display,
	description, combined_pharmaceutical_dose_form_code, combined_pharmaceutical_dose_form_display,
	legal_status_of_supply_code, additional_monitoring,
	version_id, created_at, updated_at`

func (r *medicinalProductRepoPG) scanRow(row pgx.Row) (*MedicinalProduct, error) {
	var m MedicinalProduct
	err := row.Scan(&m.ID, &m.FHIRID, &m.Status, &m.TypeCode, &m.TypeDisplay, &m.DomainCode, &m.DomainDisplay,
		&m.Description, &m.CombinedPharmaceuticalDoseFormCode, &m.CombinedPharmaceuticalDoseFormDisplay,
		&m.LegalStatusOfSupplyCode, &m.AdditionalMonitoring,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *medicinalProductRepoPG) Create(ctx context.Context, m *MedicinalProduct) error {
	m.ID = uuid.New()
	if m.FHIRID == "" {
		m.FHIRID = m.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medicinal_product (id, fhir_id, status, type_code, type_display, domain_code, domain_display,
			description, combined_pharmaceutical_dose_form_code, combined_pharmaceutical_dose_form_display,
			legal_status_of_supply_code, additional_monitoring)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		m.ID, m.FHIRID, m.Status, m.TypeCode, m.TypeDisplay, m.DomainCode, m.DomainDisplay,
		m.Description, m.CombinedPharmaceuticalDoseFormCode, m.CombinedPharmaceuticalDoseFormDisplay,
		m.LegalStatusOfSupplyCode, m.AdditionalMonitoring)
	return err
}

func (r *medicinalProductRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProduct, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mpCols+` FROM medicinal_product WHERE id = $1`, id))
}

func (r *medicinalProductRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProduct, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mpCols+` FROM medicinal_product WHERE fhir_id = $1`, fhirID))
}

func (r *medicinalProductRepoPG) Update(ctx context.Context, m *MedicinalProduct) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE medicinal_product SET status=$2, type_code=$3, type_display=$4, domain_code=$5, domain_display=$6,
			description=$7, combined_pharmaceutical_dose_form_code=$8, combined_pharmaceutical_dose_form_display=$9,
			legal_status_of_supply_code=$10, additional_monitoring=$11, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.Status, m.TypeCode, m.TypeDisplay, m.DomainCode, m.DomainDisplay,
		m.Description, m.CombinedPharmaceuticalDoseFormCode, m.CombinedPharmaceuticalDoseFormDisplay,
		m.LegalStatusOfSupplyCode, m.AdditionalMonitoring)
	return err
}

func (r *medicinalProductRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM medicinal_product WHERE id = $1`, id)
	return err
}

func (r *medicinalProductRepoPG) List(ctx context.Context, limit, offset int) ([]*MedicinalProduct, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM medicinal_product`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+mpCols+` FROM medicinal_product ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MedicinalProduct
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}

func (r *medicinalProductRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProduct, int, error) {
	query := `SELECT ` + mpCols + ` FROM medicinal_product WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM medicinal_product WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["type"]; ok {
		query += fmt.Sprintf(` AND type_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND type_code = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["domain"]; ok {
		query += fmt.Sprintf(` AND domain_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND domain_code = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
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
	var items []*MedicinalProduct
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}
