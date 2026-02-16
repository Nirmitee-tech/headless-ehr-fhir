package testreport

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

type testReportRepoPG struct{ pool *pgxpool.Pool }

func NewTestReportRepoPG(pool *pgxpool.Pool) TestReportRepository {
	return &testReportRepoPG{pool: pool}
}

func (r *testReportRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const trCols = `id, fhir_id, status, name, test_script_reference, result, score, tester, issued,
	participant_type, participant_uri,
	version_id, created_at, updated_at`

func (r *testReportRepoPG) scanRow(row pgx.Row) (*TestReport, error) {
	var e TestReport
	err := row.Scan(&e.ID, &e.FHIRID, &e.Status, &e.Name, &e.TestScriptReference, &e.Result, &e.Score, &e.Tester, &e.Issued,
		&e.ParticipantType, &e.ParticipantURI,
		&e.VersionID, &e.CreatedAt, &e.UpdatedAt)
	return &e, err
}

func (r *testReportRepoPG) Create(ctx context.Context, e *TestReport) error {
	e.ID = uuid.New()
	if e.FHIRID == "" {
		e.FHIRID = e.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO test_report (id, fhir_id, status, name, test_script_reference, result, score, tester, issued,
			participant_type, participant_uri)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		e.ID, e.FHIRID, e.Status, e.Name, e.TestScriptReference, e.Result, e.Score, e.Tester, e.Issued,
		e.ParticipantType, e.ParticipantURI)
	return err
}

func (r *testReportRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*TestReport, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+trCols+` FROM test_report WHERE id = $1`, id))
}

func (r *testReportRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*TestReport, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+trCols+` FROM test_report WHERE fhir_id = $1`, fhirID))
}

func (r *testReportRepoPG) Update(ctx context.Context, e *TestReport) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE test_report SET status=$2, name=$3, test_script_reference=$4, result=$5, score=$6, tester=$7, issued=$8,
			participant_type=$9, participant_uri=$10, updated_at=NOW()
		WHERE id = $1`,
		e.ID, e.Status, e.Name, e.TestScriptReference, e.Result, e.Score, e.Tester, e.Issued,
		e.ParticipantType, e.ParticipantURI)
	return err
}

func (r *testReportRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM test_report WHERE id = $1`, id)
	return err
}

func (r *testReportRepoPG) List(ctx context.Context, limit, offset int) ([]*TestReport, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM test_report`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+trCols+` FROM test_report ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*TestReport
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

func (r *testReportRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*TestReport, int, error) {
	query := `SELECT ` + trCols + ` FROM test_report WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM test_report WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["result"]; ok {
		query += fmt.Sprintf(` AND result = $%d`, idx)
		countQuery += fmt.Sprintf(` AND result = $%d`, idx)
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
	var items []*TestReport
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}
