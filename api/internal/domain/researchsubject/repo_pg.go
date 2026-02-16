package researchsubject

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

type researchSubjectRepoPG struct{ pool *pgxpool.Pool }

func NewResearchSubjectRepoPG(pool *pgxpool.Pool) ResearchSubjectRepository {
	return &researchSubjectRepoPG{pool: pool}
}

func (r *researchSubjectRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const rsCols = `id, fhir_id, status, study_reference, individual_reference,
	consent_reference, period_start, period_end, assigned_arm, actual_arm,
	version_id, created_at, updated_at`

func (r *researchSubjectRepoPG) scanRow(row pgx.Row) (*ResearchSubject, error) {
	var rs ResearchSubject
	err := row.Scan(&rs.ID, &rs.FHIRID, &rs.Status, &rs.StudyReference, &rs.IndividualReference,
		&rs.ConsentReference, &rs.PeriodStart, &rs.PeriodEnd, &rs.AssignedArm, &rs.ActualArm,
		&rs.VersionID, &rs.CreatedAt, &rs.UpdatedAt)
	return &rs, err
}

func (r *researchSubjectRepoPG) Create(ctx context.Context, rs *ResearchSubject) error {
	rs.ID = uuid.New()
	if rs.FHIRID == "" {
		rs.FHIRID = rs.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO research_subject (id, fhir_id, status, study_reference, individual_reference,
			consent_reference, period_start, period_end, assigned_arm, actual_arm)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		rs.ID, rs.FHIRID, rs.Status, rs.StudyReference, rs.IndividualReference,
		rs.ConsentReference, rs.PeriodStart, rs.PeriodEnd, rs.AssignedArm, rs.ActualArm)
	return err
}

func (r *researchSubjectRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ResearchSubject, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+rsCols+` FROM research_subject WHERE id = $1`, id))
}

func (r *researchSubjectRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ResearchSubject, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+rsCols+` FROM research_subject WHERE fhir_id = $1`, fhirID))
}

func (r *researchSubjectRepoPG) Update(ctx context.Context, rs *ResearchSubject) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE research_subject SET status=$2, study_reference=$3, individual_reference=$4,
			consent_reference=$5, period_start=$6, period_end=$7, assigned_arm=$8, actual_arm=$9, updated_at=NOW()
		WHERE id = $1`,
		rs.ID, rs.Status, rs.StudyReference, rs.IndividualReference,
		rs.ConsentReference, rs.PeriodStart, rs.PeriodEnd, rs.AssignedArm, rs.ActualArm)
	return err
}

func (r *researchSubjectRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM research_subject WHERE id = $1`, id)
	return err
}

func (r *researchSubjectRepoPG) List(ctx context.Context, limit, offset int) ([]*ResearchSubject, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM research_subject`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+rsCols+` FROM research_subject ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ResearchSubject
	for rows.Next() {
		rs, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, rs)
	}
	return items, total, nil
}

var researchSubjectSearchParams = map[string]fhir.SearchParamConfig{
	"status":     {Type: fhir.SearchParamToken, Column: "status"},
	"study":      {Type: fhir.SearchParamReference, Column: "study_reference"},
	"individual": {Type: fhir.SearchParamReference, Column: "individual_reference"},
}

func (r *researchSubjectRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ResearchSubject, int, error) {
	qb := fhir.NewSearchQuery("research_subject", rsCols)
	qb.ApplyParams(params, researchSubjectSearchParams)
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
	var items []*ResearchSubject
	for rows.Next() {
		rs, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, rs)
	}
	return items, total, nil
}
