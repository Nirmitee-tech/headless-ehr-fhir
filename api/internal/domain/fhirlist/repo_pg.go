package fhirlist

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

type fhirListRepoPG struct{ pool *pgxpool.Pool }

func NewFHIRListRepoPG(pool *pgxpool.Pool) FHIRListRepository {
	return &fhirListRepoPG{pool: pool}
}

func (r *fhirListRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const listCols = `id, fhir_id, status, mode, title,
	code_code, code_display, subject_patient_id, encounter_id,
	date, source_practitioner_id, ordered_by, note,
	version_id, created_at, updated_at`

func (r *fhirListRepoPG) scanRow(row pgx.Row) (*FHIRList, error) {
	var l FHIRList
	err := row.Scan(&l.ID, &l.FHIRID, &l.Status, &l.Mode, &l.Title,
		&l.CodeCode, &l.CodeDisplay, &l.SubjectPatientID, &l.EncounterID,
		&l.Date, &l.SourcePractitionerID, &l.OrderedBy, &l.Note,
		&l.VersionID, &l.CreatedAt, &l.UpdatedAt)
	return &l, err
}

func (r *fhirListRepoPG) Create(ctx context.Context, l *FHIRList) error {
	l.ID = uuid.New()
	if l.FHIRID == "" {
		l.FHIRID = l.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO fhir_list (id, fhir_id, status, mode, title,
			code_code, code_display, subject_patient_id, encounter_id,
			date, source_practitioner_id, ordered_by, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		l.ID, l.FHIRID, l.Status, l.Mode, l.Title,
		l.CodeCode, l.CodeDisplay, l.SubjectPatientID, l.EncounterID,
		l.Date, l.SourcePractitionerID, l.OrderedBy, l.Note)
	return err
}

func (r *fhirListRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*FHIRList, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+listCols+` FROM fhir_list WHERE id = $1`, id))
}

func (r *fhirListRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*FHIRList, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+listCols+` FROM fhir_list WHERE fhir_id = $1`, fhirID))
}

func (r *fhirListRepoPG) Update(ctx context.Context, l *FHIRList) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE fhir_list SET status=$2, mode=$3, title=$4,
			code_code=$5, code_display=$6, ordered_by=$7, note=$8, updated_at=NOW()
		WHERE id = $1`,
		l.ID, l.Status, l.Mode, l.Title,
		l.CodeCode, l.CodeDisplay, l.OrderedBy, l.Note)
	return err
}

func (r *fhirListRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM fhir_list WHERE id = $1`, id)
	return err
}

func (r *fhirListRepoPG) List(ctx context.Context, limit, offset int) ([]*FHIRList, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM fhir_list`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+listCols+` FROM fhir_list ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*FHIRList
	for rows.Next() {
		l, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, l)
	}
	return items, total, nil
}

func (r *fhirListRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*FHIRList, int, error) {
	query := `SELECT ` + listCols + ` FROM fhir_list WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM fhir_list WHERE 1=1`
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
	var items []*FHIRList
	for rows.Next() {
		l, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, l)
	}
	return items, total, nil
}

func (r *fhirListRepoPG) AddEntry(ctx context.Context, entry *FHIRListEntry) error {
	entry.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO fhir_list_entry (id, list_id, item_reference, item_display,
			date, deleted, flag_code, flag_display)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		entry.ID, entry.ListID, entry.ItemReference, entry.ItemDisplay,
		entry.Date, entry.Deleted, entry.FlagCode, entry.FlagDisplay)
	return err
}

func (r *fhirListRepoPG) GetEntries(ctx context.Context, listID uuid.UUID) ([]*FHIRListEntry, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, list_id, item_reference, item_display,
			date, deleted, flag_code, flag_display
		FROM fhir_list_entry WHERE list_id = $1 ORDER BY date`, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*FHIRListEntry
	for rows.Next() {
		var e FHIRListEntry
		if err := rows.Scan(&e.ID, &e.ListID, &e.ItemReference, &e.ItemDisplay,
			&e.Date, &e.Deleted, &e.FlagCode, &e.FlagDisplay); err != nil {
			return nil, err
		}
		items = append(items, &e)
	}
	return items, nil
}
