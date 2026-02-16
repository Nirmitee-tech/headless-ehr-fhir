package organizationaffiliation

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

type organizationAffiliationRepoPG struct{ pool *pgxpool.Pool }

func NewOrganizationAffiliationRepoPG(pool *pgxpool.Pool) OrganizationAffiliationRepository {
	return &organizationAffiliationRepoPG{pool: pool}
}

func (r *organizationAffiliationRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const oaCols = `id, fhir_id, active, organization_id, participating_org_id,
	period_start, period_end, code_code, code_display,
	specialty_code, specialty_display, location_id,
	telecom_phone, telecom_email,
	version_id, created_at, updated_at`

func (r *organizationAffiliationRepoPG) scanRow(row pgx.Row) (*OrganizationAffiliation, error) {
	var o OrganizationAffiliation
	err := row.Scan(&o.ID, &o.FHIRID, &o.Active, &o.OrganizationID, &o.ParticipatingOrgID,
		&o.PeriodStart, &o.PeriodEnd, &o.CodeCode, &o.CodeDisplay,
		&o.SpecialtyCode, &o.SpecialtyDisplay, &o.LocationID,
		&o.TelecomPhone, &o.TelecomEmail,
		&o.VersionID, &o.CreatedAt, &o.UpdatedAt)
	return &o, err
}

func (r *organizationAffiliationRepoPG) Create(ctx context.Context, o *OrganizationAffiliation) error {
	o.ID = uuid.New()
	if o.FHIRID == "" {
		o.FHIRID = o.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO organization_affiliation (id, fhir_id, active, organization_id, participating_org_id,
			period_start, period_end, code_code, code_display,
			specialty_code, specialty_display, location_id,
			telecom_phone, telecom_email)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		o.ID, o.FHIRID, o.Active, o.OrganizationID, o.ParticipatingOrgID,
		o.PeriodStart, o.PeriodEnd, o.CodeCode, o.CodeDisplay,
		o.SpecialtyCode, o.SpecialtyDisplay, o.LocationID,
		o.TelecomPhone, o.TelecomEmail)
	return err
}

func (r *organizationAffiliationRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*OrganizationAffiliation, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+oaCols+` FROM organization_affiliation WHERE id = $1`, id))
}

func (r *organizationAffiliationRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*OrganizationAffiliation, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+oaCols+` FROM organization_affiliation WHERE fhir_id = $1`, fhirID))
}

func (r *organizationAffiliationRepoPG) Update(ctx context.Context, o *OrganizationAffiliation) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE organization_affiliation SET active=$2, organization_id=$3, participating_org_id=$4,
			period_start=$5, period_end=$6, code_code=$7, code_display=$8,
			specialty_code=$9, specialty_display=$10, location_id=$11,
			telecom_phone=$12, telecom_email=$13, updated_at=NOW()
		WHERE id = $1`,
		o.ID, o.Active, o.OrganizationID, o.ParticipatingOrgID,
		o.PeriodStart, o.PeriodEnd, o.CodeCode, o.CodeDisplay,
		o.SpecialtyCode, o.SpecialtyDisplay, o.LocationID,
		o.TelecomPhone, o.TelecomEmail)
	return err
}

func (r *organizationAffiliationRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM organization_affiliation WHERE id = $1`, id)
	return err
}

func (r *organizationAffiliationRepoPG) List(ctx context.Context, limit, offset int) ([]*OrganizationAffiliation, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM organization_affiliation`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+oaCols+` FROM organization_affiliation ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*OrganizationAffiliation
	for rows.Next() {
		o, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, o)
	}
	return items, total, nil
}

func (r *organizationAffiliationRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*OrganizationAffiliation, int, error) {
	query := `SELECT ` + oaCols + ` FROM organization_affiliation WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM organization_affiliation WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["active"]; ok {
		query += fmt.Sprintf(` AND active = $%d`, idx)
		countQuery += fmt.Sprintf(` AND active = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["organization"]; ok {
		query += fmt.Sprintf(` AND organization_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND organization_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["participating-organization"]; ok {
		query += fmt.Sprintf(` AND participating_org_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND participating_org_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["specialty"]; ok {
		query += fmt.Sprintf(` AND specialty_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND specialty_code = $%d`, idx)
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
	var items []*OrganizationAffiliation
	for rows.Next() {
		o, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, o)
	}
	return items, total, nil
}
