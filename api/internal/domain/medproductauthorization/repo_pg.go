package medproductauthorization

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

type mpaRepoPG struct{ pool *pgxpool.Pool }

func NewMedicinalProductAuthorizationRepoPG(pool *pgxpool.Pool) MedicinalProductAuthorizationRepository {
	return &mpaRepoPG{pool: pool}
}

func (r *mpaRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil { return tx }
	if c := db.ConnFromContext(ctx); c != nil { return c }
	return r.pool
}

const mpaCols = `id, fhir_id, status, status_date, subject_reference,
	country_code, country_display, jurisdiction_code, jurisdiction_display,
	validity_period_start, validity_period_end,
	date_of_first_authorization, international_birth_date,
	holder_reference, regulator_reference,
	version_id, created_at, updated_at`

func (r *mpaRepoPG) scanRow(row pgx.Row) (*MedicinalProductAuthorization, error) {
	var m MedicinalProductAuthorization
	err := row.Scan(&m.ID, &m.FHIRID, &m.Status, &m.StatusDate, &m.SubjectReference,
		&m.CountryCode, &m.CountryDisplay, &m.JurisdictionCode, &m.JurisdictionDisplay,
		&m.ValidityPeriodStart, &m.ValidityPeriodEnd,
		&m.DateOfFirstAuthorization, &m.InternationalBirthDate,
		&m.HolderReference, &m.RegulatorReference,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *mpaRepoPG) Create(ctx context.Context, m *MedicinalProductAuthorization) error {
	m.ID = uuid.New()
	if m.FHIRID == "" { m.FHIRID = m.ID.String() }
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medicinal_product_authorization (id, fhir_id, status, status_date, subject_reference,
			country_code, country_display, jurisdiction_code, jurisdiction_display,
			validity_period_start, validity_period_end,
			date_of_first_authorization, international_birth_date,
			holder_reference, regulator_reference)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		m.ID, m.FHIRID, m.Status, m.StatusDate, m.SubjectReference,
		m.CountryCode, m.CountryDisplay, m.JurisdictionCode, m.JurisdictionDisplay,
		m.ValidityPeriodStart, m.ValidityPeriodEnd,
		m.DateOfFirstAuthorization, m.InternationalBirthDate,
		m.HolderReference, m.RegulatorReference)
	return err
}

func (r *mpaRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductAuthorization, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mpaCols+` FROM medicinal_product_authorization WHERE id = $1`, id))
}

func (r *mpaRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductAuthorization, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mpaCols+` FROM medicinal_product_authorization WHERE fhir_id = $1`, fhirID))
}

func (r *mpaRepoPG) Update(ctx context.Context, m *MedicinalProductAuthorization) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE medicinal_product_authorization SET status=$2, status_date=$3, subject_reference=$4,
			country_code=$5, country_display=$6, jurisdiction_code=$7, jurisdiction_display=$8,
			validity_period_start=$9, validity_period_end=$10,
			date_of_first_authorization=$11, international_birth_date=$12,
			holder_reference=$13, regulator_reference=$14, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.Status, m.StatusDate, m.SubjectReference,
		m.CountryCode, m.CountryDisplay, m.JurisdictionCode, m.JurisdictionDisplay,
		m.ValidityPeriodStart, m.ValidityPeriodEnd,
		m.DateOfFirstAuthorization, m.InternationalBirthDate,
		m.HolderReference, m.RegulatorReference)
	return err
}

func (r *mpaRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM medicinal_product_authorization WHERE id = $1`, id)
	return err
}

func (r *mpaRepoPG) List(ctx context.Context, limit, offset int) ([]*MedicinalProductAuthorization, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM medicinal_product_authorization`).Scan(&total); err != nil { return nil, 0, err }
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+mpaCols+` FROM medicinal_product_authorization ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	var items []*MedicinalProductAuthorization
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil { return nil, 0, err }
		items = append(items, m)
	}
	return items, total, nil
}

func (r *mpaRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductAuthorization, int, error) {
	query := `SELECT ` + mpaCols + ` FROM medicinal_product_authorization WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM medicinal_product_authorization WHERE 1=1`
	var args []interface{}
	idx := 1
	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["subject"]; ok {
		query += fmt.Sprintf(` AND subject_reference = $%d`, idx)
		countQuery += fmt.Sprintf(` AND subject_reference = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["country"]; ok {
		query += fmt.Sprintf(` AND country_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND country_code = $%d`, idx)
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
	var items []*MedicinalProductAuthorization
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil { return nil, 0, err }
		items = append(items, m)
	}
	return items, total, nil
}
