package immunizationevaluation

import (
	"context"
	"fmt"
	"strings"

	"github.com/ehr/ehr/internal/platform/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type queryable interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

type ImmunizationEvaluationRepoPG struct {
	pool *pgxpool.Pool
}

func NewImmunizationEvaluationRepoPG(pool *pgxpool.Pool) *ImmunizationEvaluationRepoPG {
	return &ImmunizationEvaluationRepoPG{pool: pool}
}

func (r *ImmunizationEvaluationRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const ieCols = `id, fhir_id, status, patient_id, date, authority_reference,
	target_disease_code, target_disease_display, immunization_event_reference,
	dose_status_code, dose_status_display, dose_status_reason_code, dose_status_reason_display,
	series, dose_number, series_doses, description,
	version_id, created_at, updated_at`

func scanIE(row pgx.Row) (*ImmunizationEvaluation, error) {
	var ie ImmunizationEvaluation
	err := row.Scan(
		&ie.ID, &ie.FHIRID, &ie.Status, &ie.PatientID, &ie.Date, &ie.AuthorityReference,
		&ie.TargetDiseaseCode, &ie.TargetDiseaseDisplay, &ie.ImmunizationEventRef,
		&ie.DoseStatusCode, &ie.DoseStatusDisplay, &ie.DoseStatusReasonCode, &ie.DoseStatusReasonDisplay,
		&ie.Series, &ie.DoseNumber, &ie.SeriesDoses, &ie.Description,
		&ie.VersionID, &ie.CreatedAt, &ie.UpdatedAt,
	)
	return &ie, err
}

func (r *ImmunizationEvaluationRepoPG) Create(ctx context.Context, ie *ImmunizationEvaluation) error {
	if ie.FHIRID == "" {
		ie.FHIRID = uuid.New().String()
	}
	q := `INSERT INTO immunization_evaluation (fhir_id, status, patient_id, date, authority_reference,
		target_disease_code, target_disease_display, immunization_event_reference,
		dose_status_code, dose_status_display, dose_status_reason_code, dose_status_reason_display,
		series, dose_number, series_doses, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		RETURNING id, version_id, created_at, updated_at`
	return r.conn(ctx).QueryRow(ctx, q,
		ie.FHIRID, ie.Status, ie.PatientID, ie.Date, ie.AuthorityReference,
		ie.TargetDiseaseCode, ie.TargetDiseaseDisplay, ie.ImmunizationEventRef,
		ie.DoseStatusCode, ie.DoseStatusDisplay, ie.DoseStatusReasonCode, ie.DoseStatusReasonDisplay,
		ie.Series, ie.DoseNumber, ie.SeriesDoses, ie.Description,
	).Scan(&ie.ID, &ie.VersionID, &ie.CreatedAt, &ie.UpdatedAt)
}

func (r *ImmunizationEvaluationRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ImmunizationEvaluation, error) {
	q := fmt.Sprintf("SELECT %s FROM immunization_evaluation WHERE id = $1", ieCols)
	return scanIE(r.conn(ctx).QueryRow(ctx, q, id))
}

func (r *ImmunizationEvaluationRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ImmunizationEvaluation, error) {
	q := fmt.Sprintf("SELECT %s FROM immunization_evaluation WHERE fhir_id = $1", ieCols)
	return scanIE(r.conn(ctx).QueryRow(ctx, q, fhirID))
}

func (r *ImmunizationEvaluationRepoPG) Update(ctx context.Context, ie *ImmunizationEvaluation) error {
	q := `UPDATE immunization_evaluation SET status=$1, patient_id=$2, date=$3, authority_reference=$4,
		target_disease_code=$5, target_disease_display=$6, immunization_event_reference=$7,
		dose_status_code=$8, dose_status_display=$9, dose_status_reason_code=$10, dose_status_reason_display=$11,
		series=$12, dose_number=$13, series_doses=$14, description=$15,
		version_id=version_id+1, updated_at=NOW()
		WHERE id=$16 RETURNING version_id, updated_at`
	return r.conn(ctx).QueryRow(ctx, q,
		ie.Status, ie.PatientID, ie.Date, ie.AuthorityReference,
		ie.TargetDiseaseCode, ie.TargetDiseaseDisplay, ie.ImmunizationEventRef,
		ie.DoseStatusCode, ie.DoseStatusDisplay, ie.DoseStatusReasonCode, ie.DoseStatusReasonDisplay,
		ie.Series, ie.DoseNumber, ie.SeriesDoses, ie.Description,
		ie.ID,
	).Scan(&ie.VersionID, &ie.UpdatedAt)
}

func (r *ImmunizationEvaluationRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, "DELETE FROM immunization_evaluation WHERE id = $1", id)
	return err
}

func (r *ImmunizationEvaluationRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ImmunizationEvaluation, int, error) {
	where := []string{}
	args := []interface{}{}
	idx := 1

	if v, ok := params["status"]; ok {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, v)
		idx++
	}
	if v, ok := params["patient"]; ok {
		where = append(where, fmt.Sprintf("patient_id = $%d", idx))
		args = append(args, v)
		idx++
	}
	if v, ok := params["target-disease"]; ok {
		where = append(where, fmt.Sprintf("target_disease_code = $%d", idx))
		args = append(args, v)
		idx++
	}
	if v, ok := params["dose-status"]; ok {
		where = append(where, fmt.Sprintf("dose_status_code = $%d", idx))
		args = append(args, v)
		idx++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	countQ := fmt.Sprintf("SELECT COUNT(*) FROM immunization_evaluation %s", whereClause)
	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := fmt.Sprintf("SELECT %s FROM immunization_evaluation %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		ieCols, whereClause, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*ImmunizationEvaluation
	for rows.Next() {
		ie, err := scanIE(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ie)
	}
	return items, total, nil
}
