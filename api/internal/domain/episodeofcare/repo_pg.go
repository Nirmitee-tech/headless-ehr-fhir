package episodeofcare

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

type episodeOfCareRepoPG struct{ pool *pgxpool.Pool }

func NewEpisodeOfCareRepoPG(pool *pgxpool.Pool) EpisodeOfCareRepository {
	return &episodeOfCareRepoPG{pool: pool}
}

func (r *episodeOfCareRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const eocCols = `id, fhir_id, status, type_code, type_display,
	diagnosis_condition_id, diagnosis_role, patient_id, managing_org_id,
	period_start, period_end, referral_request_id, care_manager_id,
	version_id, created_at, updated_at`

func (r *episodeOfCareRepoPG) scanRow(row pgx.Row) (*EpisodeOfCare, error) {
	var e EpisodeOfCare
	err := row.Scan(&e.ID, &e.FHIRID, &e.Status, &e.TypeCode, &e.TypeDisplay,
		&e.DiagnosisConditionID, &e.DiagnosisRole, &e.PatientID, &e.ManagingOrgID,
		&e.PeriodStart, &e.PeriodEnd, &e.ReferralRequestID, &e.CareManagerID,
		&e.VersionID, &e.CreatedAt, &e.UpdatedAt)
	return &e, err
}

func (r *episodeOfCareRepoPG) Create(ctx context.Context, e *EpisodeOfCare) error {
	e.ID = uuid.New()
	if e.FHIRID == "" {
		e.FHIRID = e.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO episode_of_care (id, fhir_id, status, type_code, type_display,
			diagnosis_condition_id, diagnosis_role, patient_id, managing_org_id,
			period_start, period_end, referral_request_id, care_manager_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		e.ID, e.FHIRID, e.Status, e.TypeCode, e.TypeDisplay,
		e.DiagnosisConditionID, e.DiagnosisRole, e.PatientID, e.ManagingOrgID,
		e.PeriodStart, e.PeriodEnd, e.ReferralRequestID, e.CareManagerID)
	return err
}

func (r *episodeOfCareRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*EpisodeOfCare, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+eocCols+` FROM episode_of_care WHERE id = $1`, id))
}

func (r *episodeOfCareRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*EpisodeOfCare, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+eocCols+` FROM episode_of_care WHERE fhir_id = $1`, fhirID))
}

func (r *episodeOfCareRepoPG) Update(ctx context.Context, e *EpisodeOfCare) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE episode_of_care SET status=$2, type_code=$3, type_display=$4,
			diagnosis_condition_id=$5, diagnosis_role=$6, managing_org_id=$7,
			period_start=$8, period_end=$9, care_manager_id=$10, updated_at=NOW()
		WHERE id = $1`,
		e.ID, e.Status, e.TypeCode, e.TypeDisplay,
		e.DiagnosisConditionID, e.DiagnosisRole, e.ManagingOrgID,
		e.PeriodStart, e.PeriodEnd, e.CareManagerID)
	return err
}

func (r *episodeOfCareRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM episode_of_care WHERE id = $1`, id)
	return err
}

func (r *episodeOfCareRepoPG) List(ctx context.Context, limit, offset int) ([]*EpisodeOfCare, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM episode_of_care`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+eocCols+` FROM episode_of_care ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*EpisodeOfCare
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

func (r *episodeOfCareRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*EpisodeOfCare, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM episode_of_care WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+eocCols+` FROM episode_of_care WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*EpisodeOfCare
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

func (r *episodeOfCareRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*EpisodeOfCare, int, error) {
	query := `SELECT ` + eocCols + ` FROM episode_of_care WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM episode_of_care WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient"]; ok {
		query += fmt.Sprintf(` AND patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND patient_id = $%d`, idx)
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
	var items []*EpisodeOfCare
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}
