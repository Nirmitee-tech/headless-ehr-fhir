package verificationresult

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

type verificationResultRepoPG struct{ pool *pgxpool.Pool }

func NewVerificationResultRepoPG(pool *pgxpool.Pool) VerificationResultRepository {
	return &verificationResultRepoPG{pool: pool}
}

func (r *verificationResultRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const vrCols = `id, fhir_id, status, target_type, target_reference,
	need_code, need_display, status_date,
	validation_type_code, validation_type_display,
	validation_process_code, validation_process_display,
	frequency_value, frequency_unit, last_performed, next_scheduled,
	failure_action_code, failure_action_display,
	version_id, created_at, updated_at`

func (r *verificationResultRepoPG) scanRow(row pgx.Row) (*VerificationResult, error) {
	var v VerificationResult
	err := row.Scan(&v.ID, &v.FHIRID, &v.Status, &v.TargetType, &v.TargetReference,
		&v.NeedCode, &v.NeedDisplay, &v.StatusDate,
		&v.ValidationTypeCode, &v.ValidationTypeDisplay,
		&v.ValidationProcessCode, &v.ValidationProcessDisplay,
		&v.FrequencyValue, &v.FrequencyUnit, &v.LastPerformed, &v.NextScheduled,
		&v.FailureActionCode, &v.FailureActionDisplay,
		&v.VersionID, &v.CreatedAt, &v.UpdatedAt)
	return &v, err
}

func (r *verificationResultRepoPG) Create(ctx context.Context, v *VerificationResult) error {
	v.ID = uuid.New()
	if v.FHIRID == "" {
		v.FHIRID = v.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO verification_result (id, fhir_id, status, target_type, target_reference,
			need_code, need_display, status_date,
			validation_type_code, validation_type_display,
			validation_process_code, validation_process_display,
			frequency_value, frequency_unit, last_performed, next_scheduled,
			failure_action_code, failure_action_display)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		v.ID, v.FHIRID, v.Status, v.TargetType, v.TargetReference,
		v.NeedCode, v.NeedDisplay, v.StatusDate,
		v.ValidationTypeCode, v.ValidationTypeDisplay,
		v.ValidationProcessCode, v.ValidationProcessDisplay,
		v.FrequencyValue, v.FrequencyUnit, v.LastPerformed, v.NextScheduled,
		v.FailureActionCode, v.FailureActionDisplay)
	return err
}

func (r *verificationResultRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*VerificationResult, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+vrCols+` FROM verification_result WHERE id = $1`, id))
}

func (r *verificationResultRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*VerificationResult, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+vrCols+` FROM verification_result WHERE fhir_id = $1`, fhirID))
}

func (r *verificationResultRepoPG) Update(ctx context.Context, v *VerificationResult) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE verification_result SET status=$2, target_type=$3, target_reference=$4,
			need_code=$5, need_display=$6, status_date=$7,
			validation_type_code=$8, validation_type_display=$9,
			validation_process_code=$10, validation_process_display=$11,
			frequency_value=$12, frequency_unit=$13, last_performed=$14, next_scheduled=$15,
			failure_action_code=$16, failure_action_display=$17, updated_at=NOW()
		WHERE id = $1`,
		v.ID, v.Status, v.TargetType, v.TargetReference,
		v.NeedCode, v.NeedDisplay, v.StatusDate,
		v.ValidationTypeCode, v.ValidationTypeDisplay,
		v.ValidationProcessCode, v.ValidationProcessDisplay,
		v.FrequencyValue, v.FrequencyUnit, v.LastPerformed, v.NextScheduled,
		v.FailureActionCode, v.FailureActionDisplay)
	return err
}

func (r *verificationResultRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM verification_result WHERE id = $1`, id)
	return err
}

func (r *verificationResultRepoPG) List(ctx context.Context, limit, offset int) ([]*VerificationResult, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM verification_result`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+vrCols+` FROM verification_result ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*VerificationResult
	for rows.Next() {
		v, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, v)
	}
	return items, total, nil
}

func (r *verificationResultRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*VerificationResult, int, error) {
	query := `SELECT ` + vrCols + ` FROM verification_result WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM verification_result WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["target"]; ok {
		query += fmt.Sprintf(` AND target_reference = $%d`, idx)
		countQuery += fmt.Sprintf(` AND target_reference = $%d`, idx)
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
	var items []*VerificationResult
	for rows.Next() {
		v, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, v)
	}
	return items, total, nil
}
