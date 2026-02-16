package familyhistory

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

// =========== FamilyMemberHistory Repository ===========

type familyMemberHistoryRepoPG struct{ pool *pgxpool.Pool }

func NewFamilyMemberHistoryRepoPG(pool *pgxpool.Pool) FamilyMemberHistoryRepository {
	return &familyMemberHistoryRepoPG{pool: pool}
}

func (r *familyMemberHistoryRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const fmhCols = `id, fhir_id, status, patient_id, date, name,
	relationship_code, relationship_display, sex, born_date,
	deceased_boolean, deceased_age, note, created_at, updated_at`

func (r *familyMemberHistoryRepoPG) scanFMH(row pgx.Row) (*FamilyMemberHistory, error) {
	var f FamilyMemberHistory
	err := row.Scan(&f.ID, &f.FHIRID, &f.Status, &f.PatientID, &f.Date, &f.Name,
		&f.RelationshipCode, &f.RelationshipDisplay, &f.Sex, &f.BornDate,
		&f.DeceasedBoolean, &f.DeceasedAge, &f.Note, &f.CreatedAt, &f.UpdatedAt)
	return &f, err
}

func (r *familyMemberHistoryRepoPG) Create(ctx context.Context, f *FamilyMemberHistory) error {
	f.ID = uuid.New()
	if f.FHIRID == "" {
		f.FHIRID = f.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO family_member_history (id, fhir_id, status, patient_id, date, name,
			relationship_code, relationship_display, sex, born_date,
			deceased_boolean, deceased_age, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		f.ID, f.FHIRID, f.Status, f.PatientID, f.Date, f.Name,
		f.RelationshipCode, f.RelationshipDisplay, f.Sex, f.BornDate,
		f.DeceasedBoolean, f.DeceasedAge, f.Note)
	return err
}

func (r *familyMemberHistoryRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*FamilyMemberHistory, error) {
	return r.scanFMH(r.conn(ctx).QueryRow(ctx, `SELECT `+fmhCols+` FROM family_member_history WHERE id = $1`, id))
}

func (r *familyMemberHistoryRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*FamilyMemberHistory, error) {
	return r.scanFMH(r.conn(ctx).QueryRow(ctx, `SELECT `+fmhCols+` FROM family_member_history WHERE fhir_id = $1`, fhirID))
}

func (r *familyMemberHistoryRepoPG) Update(ctx context.Context, f *FamilyMemberHistory) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE family_member_history SET status=$2, name=$3, relationship_code=$4,
			relationship_display=$5, note=$6, updated_at=NOW()
		WHERE id = $1`,
		f.ID, f.Status, f.Name, f.RelationshipCode, f.RelationshipDisplay, f.Note)
	return err
}

func (r *familyMemberHistoryRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM family_member_history WHERE id = $1`, id)
	return err
}

func (r *familyMemberHistoryRepoPG) List(ctx context.Context, limit, offset int) ([]*FamilyMemberHistory, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM family_member_history`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+fmhCols+` FROM family_member_history ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*FamilyMemberHistory
	for rows.Next() {
		f, err := r.scanFMH(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, f)
	}
	return items, total, nil
}

func (r *familyMemberHistoryRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*FamilyMemberHistory, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM family_member_history WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+fmhCols+` FROM family_member_history WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*FamilyMemberHistory
	for rows.Next() {
		f, err := r.scanFMH(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, f)
	}
	return items, total, nil
}

var familyMemberHistorySearchParams = map[string]fhir.SearchParamConfig{
	"patient":      {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":       {Type: fhir.SearchParamToken, Column: "status"},
	"relationship": {Type: fhir.SearchParamToken, Column: "relationship_code"},
}

func (r *familyMemberHistoryRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*FamilyMemberHistory, int, error) {
	qb := fhir.NewSearchQuery("family_member_history", fmhCols)
	qb.ApplyParams(params, familyMemberHistorySearchParams)
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
	var items []*FamilyMemberHistory
	for rows.Next() {
		f, err := r.scanFMH(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, f)
	}
	return items, total, nil
}

func (r *familyMemberHistoryRepoPG) AddCondition(ctx context.Context, c *FamilyMemberCondition) error {
	c.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO family_member_condition (id, family_member_id, code, display,
			outcome_code, outcome_display, contributed_to_death, onset_age)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		c.ID, c.FamilyMemberID, c.Code, c.Display,
		c.OutcomeCode, c.OutcomeDisplay, c.ContributedToDeath, c.OnsetAge)
	return err
}

func (r *familyMemberHistoryRepoPG) GetConditions(ctx context.Context, familyMemberID uuid.UUID) ([]*FamilyMemberCondition, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, family_member_id, code, display, outcome_code, outcome_display,
			contributed_to_death, onset_age
		FROM family_member_condition WHERE family_member_id = $1`, familyMemberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*FamilyMemberCondition
	for rows.Next() {
		var c FamilyMemberCondition
		if err := rows.Scan(&c.ID, &c.FamilyMemberID, &c.Code, &c.Display,
			&c.OutcomeCode, &c.OutcomeDisplay, &c.ContributedToDeath, &c.OnsetAge); err != nil {
			return nil, err
		}
		items = append(items, &c)
	}
	return items, nil
}
