package media

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

type mediaRepoPG struct{ pool *pgxpool.Pool }

func NewMediaRepoPG(pool *pgxpool.Pool) MediaRepository {
	return &mediaRepoPG{pool: pool}
}

func (r *mediaRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const mediaCols = `id, fhir_id, status, type_code, type_display,
	modality_code, modality_display,
	subject_patient_id, encounter_id, created_date, operator_id,
	reason_code, body_site_code, body_site_display, device_name,
	height, width, frames, duration,
	content_type, content_url, content_size, content_title,
	note, version_id, created_at, updated_at`

func (r *mediaRepoPG) scanRow(row pgx.Row) (*Media, error) {
	var m Media
	err := row.Scan(&m.ID, &m.FHIRID, &m.Status, &m.TypeCode, &m.TypeDisplay,
		&m.ModalityCode, &m.ModalityDisplay,
		&m.SubjectPatientID, &m.EncounterID, &m.CreatedDate, &m.OperatorID,
		&m.ReasonCode, &m.BodySiteCode, &m.BodySiteDisplay, &m.DeviceName,
		&m.Height, &m.Width, &m.Frames, &m.Duration,
		&m.ContentType, &m.ContentURL, &m.ContentSize, &m.ContentTitle,
		&m.Note, &m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *mediaRepoPG) Create(ctx context.Context, m *Media) error {
	m.ID = uuid.New()
	if m.FHIRID == "" {
		m.FHIRID = m.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO media (id, fhir_id, status, type_code, type_display,
			modality_code, modality_display,
			subject_patient_id, encounter_id, created_date, operator_id,
			reason_code, body_site_code, body_site_display, device_name,
			height, width, frames, duration,
			content_type, content_url, content_size, content_title, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)`,
		m.ID, m.FHIRID, m.Status, m.TypeCode, m.TypeDisplay,
		m.ModalityCode, m.ModalityDisplay,
		m.SubjectPatientID, m.EncounterID, m.CreatedDate, m.OperatorID,
		m.ReasonCode, m.BodySiteCode, m.BodySiteDisplay, m.DeviceName,
		m.Height, m.Width, m.Frames, m.Duration,
		m.ContentType, m.ContentURL, m.ContentSize, m.ContentTitle, m.Note)
	return err
}

func (r *mediaRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Media, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mediaCols+` FROM media WHERE id = $1`, id))
}

func (r *mediaRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Media, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mediaCols+` FROM media WHERE fhir_id = $1`, fhirID))
}

func (r *mediaRepoPG) Update(ctx context.Context, m *Media) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE media SET status=$2, type_code=$3, type_display=$4,
			modality_code=$5, modality_display=$6,
			subject_patient_id=$7, encounter_id=$8, created_date=$9, operator_id=$10,
			reason_code=$11, body_site_code=$12, body_site_display=$13, device_name=$14,
			height=$15, width=$16, frames=$17, duration=$18,
			content_type=$19, content_url=$20, content_size=$21, content_title=$22,
			note=$23, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.Status, m.TypeCode, m.TypeDisplay,
		m.ModalityCode, m.ModalityDisplay,
		m.SubjectPatientID, m.EncounterID, m.CreatedDate, m.OperatorID,
		m.ReasonCode, m.BodySiteCode, m.BodySiteDisplay, m.DeviceName,
		m.Height, m.Width, m.Frames, m.Duration,
		m.ContentType, m.ContentURL, m.ContentSize, m.ContentTitle, m.Note)
	return err
}

func (r *mediaRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM media WHERE id = $1`, id)
	return err
}

func (r *mediaRepoPG) List(ctx context.Context, limit, offset int) ([]*Media, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM media`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+mediaCols+` FROM media ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Media
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}

func (r *mediaRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Media, int, error) {
	query := `SELECT ` + mediaCols + ` FROM media WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM media WHERE 1=1`
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
	if p, ok := params["type"]; ok {
		query += fmt.Sprintf(` AND type_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND type_code = $%d`, idx)
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
	var items []*Media
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}
