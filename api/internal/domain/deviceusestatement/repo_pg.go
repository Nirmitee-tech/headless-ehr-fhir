package deviceusestatement

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

type deviceUseStatementRepoPG struct{ pool *pgxpool.Pool }

func NewDeviceUseStatementRepoPG(pool *pgxpool.Pool) DeviceUseStatementRepository {
	return &deviceUseStatementRepoPG{pool: pool}
}

func (r *deviceUseStatementRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const dusCols = `id, fhir_id, status, subject_patient_id, device_id,
	timing_date, timing_period_start, timing_period_end,
	recorded_on, source_id,
	reason_code, reason_display,
	body_site_code, body_site_display, note,
	version_id, created_at, updated_at`

func (r *deviceUseStatementRepoPG) scanRow(row pgx.Row) (*DeviceUseStatement, error) {
	var d DeviceUseStatement
	err := row.Scan(&d.ID, &d.FHIRID, &d.Status, &d.SubjectPatientID, &d.DeviceID,
		&d.TimingDate, &d.TimingPeriodStart, &d.TimingPeriodEnd,
		&d.RecordedOn, &d.SourceID,
		&d.ReasonCode, &d.ReasonDisplay,
		&d.BodySiteCode, &d.BodySiteDisplay, &d.Note,
		&d.VersionID, &d.CreatedAt, &d.UpdatedAt)
	return &d, err
}

func (r *deviceUseStatementRepoPG) Create(ctx context.Context, d *DeviceUseStatement) error {
	d.ID = uuid.New()
	if d.FHIRID == "" {
		d.FHIRID = d.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO device_use_statement (id, fhir_id, status, subject_patient_id, device_id,
			timing_date, timing_period_start, timing_period_end,
			recorded_on, source_id,
			reason_code, reason_display,
			body_site_code, body_site_display, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		d.ID, d.FHIRID, d.Status, d.SubjectPatientID, d.DeviceID,
		d.TimingDate, d.TimingPeriodStart, d.TimingPeriodEnd,
		d.RecordedOn, d.SourceID,
		d.ReasonCode, d.ReasonDisplay,
		d.BodySiteCode, d.BodySiteDisplay, d.Note)
	return err
}

func (r *deviceUseStatementRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*DeviceUseStatement, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+dusCols+` FROM device_use_statement WHERE id = $1`, id))
}

func (r *deviceUseStatementRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*DeviceUseStatement, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+dusCols+` FROM device_use_statement WHERE fhir_id = $1`, fhirID))
}

func (r *deviceUseStatementRepoPG) Update(ctx context.Context, d *DeviceUseStatement) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE device_use_statement SET status=$2, device_id=$3,
			timing_date=$4, timing_period_start=$5, timing_period_end=$6,
			recorded_on=$7, source_id=$8,
			reason_code=$9, reason_display=$10,
			body_site_code=$11, body_site_display=$12, note=$13, updated_at=NOW()
		WHERE id = $1`,
		d.ID, d.Status, d.DeviceID,
		d.TimingDate, d.TimingPeriodStart, d.TimingPeriodEnd,
		d.RecordedOn, d.SourceID,
		d.ReasonCode, d.ReasonDisplay,
		d.BodySiteCode, d.BodySiteDisplay, d.Note)
	return err
}

func (r *deviceUseStatementRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM device_use_statement WHERE id = $1`, id)
	return err
}

func (r *deviceUseStatementRepoPG) List(ctx context.Context, limit, offset int) ([]*DeviceUseStatement, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM device_use_statement`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+dusCols+` FROM device_use_statement ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*DeviceUseStatement
	for rows.Next() {
		d, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}

func (r *deviceUseStatementRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*DeviceUseStatement, int, error) {
	query := `SELECT ` + dusCols + ` FROM device_use_statement WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM device_use_statement WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient"]; ok {
		query += fmt.Sprintf(` AND subject_patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND subject_patient_id = $%d`, idx)
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
	var items []*DeviceUseStatement
	for rows.Next() {
		d, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}
