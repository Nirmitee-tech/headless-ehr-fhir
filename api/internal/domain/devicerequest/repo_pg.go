package devicerequest

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

type deviceRequestRepoPG struct{ pool *pgxpool.Pool }

func NewDeviceRequestRepoPG(pool *pgxpool.Pool) DeviceRequestRepository {
	return &deviceRequestRepoPG{pool: pool}
}

func (r *deviceRequestRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const drCols = `id, fhir_id, status, intent, priority,
	code_code, code_display, code_system,
	subject_patient_id, encounter_id, authored_on,
	requester_id, performer_id,
	reason_code, reason_display, note,
	version_id, created_at, updated_at`

func (r *deviceRequestRepoPG) scanRow(row pgx.Row) (*DeviceRequest, error) {
	var d DeviceRequest
	err := row.Scan(&d.ID, &d.FHIRID, &d.Status, &d.Intent, &d.Priority,
		&d.CodeCode, &d.CodeDisplay, &d.CodeSystem,
		&d.SubjectPatientID, &d.EncounterID, &d.AuthoredOn,
		&d.RequesterID, &d.PerformerID,
		&d.ReasonCode, &d.ReasonDisplay, &d.Note,
		&d.VersionID, &d.CreatedAt, &d.UpdatedAt)
	return &d, err
}

func (r *deviceRequestRepoPG) Create(ctx context.Context, d *DeviceRequest) error {
	d.ID = uuid.New()
	if d.FHIRID == "" {
		d.FHIRID = d.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO device_request (id, fhir_id, status, intent, priority,
			code_code, code_display, code_system,
			subject_patient_id, encounter_id, authored_on,
			requester_id, performer_id,
			reason_code, reason_display, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		d.ID, d.FHIRID, d.Status, d.Intent, d.Priority,
		d.CodeCode, d.CodeDisplay, d.CodeSystem,
		d.SubjectPatientID, d.EncounterID, d.AuthoredOn,
		d.RequesterID, d.PerformerID,
		d.ReasonCode, d.ReasonDisplay, d.Note)
	return err
}

func (r *deviceRequestRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*DeviceRequest, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+drCols+` FROM device_request WHERE id = $1`, id))
}

func (r *deviceRequestRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*DeviceRequest, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+drCols+` FROM device_request WHERE fhir_id = $1`, fhirID))
}

func (r *deviceRequestRepoPG) Update(ctx context.Context, d *DeviceRequest) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE device_request SET status=$2, intent=$3, priority=$4,
			code_code=$5, code_display=$6, code_system=$7,
			encounter_id=$8, authored_on=$9,
			requester_id=$10, performer_id=$11,
			reason_code=$12, reason_display=$13, note=$14, updated_at=NOW()
		WHERE id = $1`,
		d.ID, d.Status, d.Intent, d.Priority,
		d.CodeCode, d.CodeDisplay, d.CodeSystem,
		d.EncounterID, d.AuthoredOn,
		d.RequesterID, d.PerformerID,
		d.ReasonCode, d.ReasonDisplay, d.Note)
	return err
}

func (r *deviceRequestRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM device_request WHERE id = $1`, id)
	return err
}

func (r *deviceRequestRepoPG) List(ctx context.Context, limit, offset int) ([]*DeviceRequest, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM device_request`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+drCols+` FROM device_request ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*DeviceRequest
	for rows.Next() {
		d, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}

var drSearchParams = map[string]fhir.SearchParamConfig{
	"patient": {Type: fhir.SearchParamReference, Column: "subject_patient_id"},
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
	"intent":  {Type: fhir.SearchParamToken, Column: "intent"},
}

func (r *deviceRequestRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*DeviceRequest, int, error) {
	qb := fhir.NewSearchQuery("device_request", drCols)
	qb.ApplyParams(params, drSearchParams)
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
	var items []*DeviceRequest
	for rows.Next() {
		d, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}
