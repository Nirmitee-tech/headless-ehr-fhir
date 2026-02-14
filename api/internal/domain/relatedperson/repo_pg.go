package relatedperson

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

type relatedPersonRepoPG struct{ pool *pgxpool.Pool }

func NewRelatedPersonRepoPG(pool *pgxpool.Pool) RelatedPersonRepository {
	return &relatedPersonRepoPG{pool: pool}
}

func (r *relatedPersonRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const rpCols = `id, fhir_id, active, patient_id, relationship_code, relationship_display,
	family_name, given_name, phone, email, gender, birth_date,
	address_line, address_city, address_state, address_postal_code,
	period_start, period_end, created_at, updated_at`

func (r *relatedPersonRepoPG) scanRP(row pgx.Row) (*RelatedPerson, error) {
	var rp RelatedPerson
	err := row.Scan(&rp.ID, &rp.FHIRID, &rp.Active, &rp.PatientID,
		&rp.RelationshipCode, &rp.RelationshipDisplay,
		&rp.FamilyName, &rp.GivenName, &rp.Phone, &rp.Email,
		&rp.Gender, &rp.BirthDate,
		&rp.AddressLine, &rp.AddressCity, &rp.AddressState, &rp.AddressPostalCode,
		&rp.PeriodStart, &rp.PeriodEnd, &rp.CreatedAt, &rp.UpdatedAt)
	return &rp, err
}

func (r *relatedPersonRepoPG) Create(ctx context.Context, rp *RelatedPerson) error {
	rp.ID = uuid.New()
	if rp.FHIRID == "" {
		rp.FHIRID = rp.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO related_person (id, fhir_id, active, patient_id, relationship_code, relationship_display,
			family_name, given_name, phone, email, gender, birth_date,
			address_line, address_city, address_state, address_postal_code,
			period_start, period_end)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		rp.ID, rp.FHIRID, rp.Active, rp.PatientID, rp.RelationshipCode, rp.RelationshipDisplay,
		rp.FamilyName, rp.GivenName, rp.Phone, rp.Email, rp.Gender, rp.BirthDate,
		rp.AddressLine, rp.AddressCity, rp.AddressState, rp.AddressPostalCode,
		rp.PeriodStart, rp.PeriodEnd)
	return err
}

func (r *relatedPersonRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*RelatedPerson, error) {
	return r.scanRP(r.conn(ctx).QueryRow(ctx, `SELECT `+rpCols+` FROM related_person WHERE id = $1`, id))
}

func (r *relatedPersonRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*RelatedPerson, error) {
	return r.scanRP(r.conn(ctx).QueryRow(ctx, `SELECT `+rpCols+` FROM related_person WHERE fhir_id = $1`, fhirID))
}

func (r *relatedPersonRepoPG) Update(ctx context.Context, rp *RelatedPerson) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE related_person SET active=$2, relationship_code=$3, relationship_display=$4,
			family_name=$5, given_name=$6, phone=$7, email=$8, updated_at=NOW()
		WHERE id = $1`,
		rp.ID, rp.Active, rp.RelationshipCode, rp.RelationshipDisplay,
		rp.FamilyName, rp.GivenName, rp.Phone, rp.Email)
	return err
}

func (r *relatedPersonRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM related_person WHERE id = $1`, id)
	return err
}

func (r *relatedPersonRepoPG) List(ctx context.Context, limit, offset int) ([]*RelatedPerson, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM related_person`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+rpCols+` FROM related_person ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*RelatedPerson
	for rows.Next() {
		rp, err := r.scanRP(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, rp)
	}
	return items, total, nil
}

func (r *relatedPersonRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*RelatedPerson, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM related_person WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+rpCols+` FROM related_person WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*RelatedPerson
	for rows.Next() {
		rp, err := r.scanRP(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, rp)
	}
	return items, total, nil
}

func (r *relatedPersonRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*RelatedPerson, int, error) {
	query := `SELECT ` + rpCols + ` FROM related_person WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM related_person WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient"]; ok {
		query += fmt.Sprintf(` AND patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND patient_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["relationship"]; ok {
		query += fmt.Sprintf(` AND relationship_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND relationship_code = $%d`, idx)
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
	var items []*RelatedPerson
	for rows.Next() {
		rp, err := r.scanRP(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, rp)
	}
	return items, total, nil
}

func (r *relatedPersonRepoPG) AddCommunication(ctx context.Context, c *RelatedPersonCommunication) error {
	c.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO related_person_communication (id, related_person_id, language_code, language_display, preferred)
		VALUES ($1,$2,$3,$4,$5)`,
		c.ID, c.RelatedPersonID, c.LanguageCode, c.LanguageDisplay, c.Preferred)
	return err
}

func (r *relatedPersonRepoPG) GetCommunications(ctx context.Context, relatedPersonID uuid.UUID) ([]*RelatedPersonCommunication, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, related_person_id, language_code, language_display, preferred
		FROM related_person_communication WHERE related_person_id = $1`, relatedPersonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*RelatedPersonCommunication
	for rows.Next() {
		var c RelatedPersonCommunication
		if err := rows.Scan(&c.ID, &c.RelatedPersonID, &c.LanguageCode, &c.LanguageDisplay, &c.Preferred); err != nil {
			return nil, err
		}
		items = append(items, &c)
	}
	return items, nil
}
