package communicationrequest

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

type communicationRequestRepoPG struct{ pool *pgxpool.Pool }

func NewCommunicationRequestRepoPG(pool *pgxpool.Pool) CommunicationRequestRepository {
	return &communicationRequestRepoPG{pool: pool}
}

func (r *communicationRequestRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const crCols = `id, fhir_id, status, patient_id, encounter_id,
	requester_id, recipient_id, sender_id,
	category_code, category_display, priority,
	medium_code, medium_display, payload_text,
	occurrence_date, authored_on,
	reason_code, reason_display, note,
	version_id, created_at, updated_at`

func (r *communicationRequestRepoPG) scanRow(row pgx.Row) (*CommunicationRequest, error) {
	var cr CommunicationRequest
	err := row.Scan(&cr.ID, &cr.FHIRID, &cr.Status, &cr.PatientID, &cr.EncounterID,
		&cr.RequesterID, &cr.RecipientID, &cr.SenderID,
		&cr.CategoryCode, &cr.CategoryDisplay, &cr.Priority,
		&cr.MediumCode, &cr.MediumDisplay, &cr.PayloadText,
		&cr.OccurrenceDate, &cr.AuthoredOn,
		&cr.ReasonCode, &cr.ReasonDisplay, &cr.Note,
		&cr.VersionID, &cr.CreatedAt, &cr.UpdatedAt)
	return &cr, err
}

func (r *communicationRequestRepoPG) Create(ctx context.Context, cr *CommunicationRequest) error {
	cr.ID = uuid.New()
	if cr.FHIRID == "" {
		cr.FHIRID = cr.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO communication_request (id, fhir_id, status, patient_id, encounter_id,
			requester_id, recipient_id, sender_id,
			category_code, category_display, priority,
			medium_code, medium_display, payload_text,
			occurrence_date, authored_on,
			reason_code, reason_display, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		cr.ID, cr.FHIRID, cr.Status, cr.PatientID, cr.EncounterID,
		cr.RequesterID, cr.RecipientID, cr.SenderID,
		cr.CategoryCode, cr.CategoryDisplay, cr.Priority,
		cr.MediumCode, cr.MediumDisplay, cr.PayloadText,
		cr.OccurrenceDate, cr.AuthoredOn,
		cr.ReasonCode, cr.ReasonDisplay, cr.Note)
	return err
}

func (r *communicationRequestRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*CommunicationRequest, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+crCols+` FROM communication_request WHERE id = $1`, id))
}

func (r *communicationRequestRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*CommunicationRequest, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+crCols+` FROM communication_request WHERE fhir_id = $1`, fhirID))
}

func (r *communicationRequestRepoPG) Update(ctx context.Context, cr *CommunicationRequest) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE communication_request SET status=$2, patient_id=$3, encounter_id=$4,
			requester_id=$5, recipient_id=$6, sender_id=$7,
			category_code=$8, category_display=$9, priority=$10,
			medium_code=$11, medium_display=$12, payload_text=$13,
			occurrence_date=$14, authored_on=$15,
			reason_code=$16, reason_display=$17, note=$18, updated_at=NOW()
		WHERE id = $1`,
		cr.ID, cr.Status, cr.PatientID, cr.EncounterID,
		cr.RequesterID, cr.RecipientID, cr.SenderID,
		cr.CategoryCode, cr.CategoryDisplay, cr.Priority,
		cr.MediumCode, cr.MediumDisplay, cr.PayloadText,
		cr.OccurrenceDate, cr.AuthoredOn,
		cr.ReasonCode, cr.ReasonDisplay, cr.Note)
	return err
}

func (r *communicationRequestRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM communication_request WHERE id = $1`, id)
	return err
}

func (r *communicationRequestRepoPG) List(ctx context.Context, limit, offset int) ([]*CommunicationRequest, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM communication_request`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+crCols+` FROM communication_request ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*CommunicationRequest
	for rows.Next() {
		cr, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, cr)
	}
	return items, total, nil
}

var communicationRequestSearchParams = map[string]fhir.SearchParamConfig{
	"patient":  {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":   {Type: fhir.SearchParamToken, Column: "status"},
	"priority": {Type: fhir.SearchParamToken, Column: "priority"},
	"category": {Type: fhir.SearchParamToken, Column: "category_code"},
}

func (r *communicationRequestRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*CommunicationRequest, int, error) {
	qb := fhir.NewSearchQuery("communication_request", crCols)
	qb.ApplyParams(params, communicationRequestSearchParams)
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
	var items []*CommunicationRequest
	for rows.Next() {
		cr, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, cr)
	}
	return items, total, nil
}
