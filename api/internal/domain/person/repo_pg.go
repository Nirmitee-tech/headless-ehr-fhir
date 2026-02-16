package person

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

type personRepoPG struct{ pool *pgxpool.Pool }

func NewPersonRepoPG(pool *pgxpool.Pool) PersonRepository {
	return &personRepoPG{pool: pool}
}

func (r *personRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const personCols = `id, fhir_id, active, name_family, name_given, gender, birth_date,
	address_line, address_city, address_state, address_postal_code,
	telecom_phone, telecom_email, managing_org_id,
	version_id, created_at, updated_at`

func (r *personRepoPG) scanRow(row pgx.Row) (*Person, error) {
	var p Person
	err := row.Scan(&p.ID, &p.FHIRID, &p.Active, &p.NameFamily, &p.NameGiven, &p.Gender, &p.BirthDate,
		&p.AddressLine, &p.AddressCity, &p.AddressState, &p.AddressPostalCode,
		&p.TelecomPhone, &p.TelecomEmail, &p.ManagingOrgID,
		&p.VersionID, &p.CreatedAt, &p.UpdatedAt)
	return &p, err
}

func (r *personRepoPG) Create(ctx context.Context, p *Person) error {
	p.ID = uuid.New()
	if p.FHIRID == "" {
		p.FHIRID = p.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO person (id, fhir_id, active, name_family, name_given, gender, birth_date,
			address_line, address_city, address_state, address_postal_code,
			telecom_phone, telecom_email, managing_org_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		p.ID, p.FHIRID, p.Active, p.NameFamily, p.NameGiven, p.Gender, p.BirthDate,
		p.AddressLine, p.AddressCity, p.AddressState, p.AddressPostalCode,
		p.TelecomPhone, p.TelecomEmail, p.ManagingOrgID)
	return err
}

func (r *personRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Person, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+personCols+` FROM person WHERE id = $1`, id))
}

func (r *personRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Person, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+personCols+` FROM person WHERE fhir_id = $1`, fhirID))
}

func (r *personRepoPG) Update(ctx context.Context, p *Person) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE person SET active=$2, name_family=$3, name_given=$4, gender=$5, birth_date=$6,
			address_line=$7, address_city=$8, address_state=$9, address_postal_code=$10,
			telecom_phone=$11, telecom_email=$12, managing_org_id=$13, updated_at=NOW()
		WHERE id = $1`,
		p.ID, p.Active, p.NameFamily, p.NameGiven, p.Gender, p.BirthDate,
		p.AddressLine, p.AddressCity, p.AddressState, p.AddressPostalCode,
		p.TelecomPhone, p.TelecomEmail, p.ManagingOrgID)
	return err
}

func (r *personRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM person WHERE id = $1`, id)
	return err
}

func (r *personRepoPG) List(ctx context.Context, limit, offset int) ([]*Person, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM person`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+personCols+` FROM person ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Person
	for rows.Next() {
		p, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, nil
}

var personSearchParams = map[string]fhir.SearchParamConfig{
	"name":   {Type: fhir.SearchParamString, Column: "name_family"},
	"gender": {Type: fhir.SearchParamToken, Column: "gender"},
	"active": {Type: fhir.SearchParamToken, Column: "active"},
}

func (r *personRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Person, int, error) {
	qb := fhir.NewSearchQuery("person", personCols)
	qb.ApplyParams(params, personSearchParams)
	qb.OrderBy("created_at DESC")

	var total int
	if err := r.conn(ctx).QueryRow(ctx, qb.CountSQL(), qb.CountArgs()...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx, qb.DataSQL(limit, offset), qb.DataArgs(limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Person
	for rows.Next() {
		p, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, nil
}
