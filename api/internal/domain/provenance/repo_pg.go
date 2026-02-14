package provenance

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

type provenanceRepoPG struct{ pool *pgxpool.Pool }

func NewProvenanceRepoPG(pool *pgxpool.Pool) ProvenanceRepository {
	return &provenanceRepoPG{pool: pool}
}

func (r *provenanceRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const provCols = `id, fhir_id, target_type, target_id, recorded,
	activity_code, activity_display, reason_code, reason_display,
	created_at, updated_at`

func (r *provenanceRepoPG) scanProv(row pgx.Row) (*Provenance, error) {
	var p Provenance
	err := row.Scan(&p.ID, &p.FHIRID, &p.TargetType, &p.TargetID, &p.Recorded,
		&p.ActivityCode, &p.ActivityDisplay, &p.ReasonCode, &p.ReasonDisplay,
		&p.CreatedAt, &p.UpdatedAt)
	return &p, err
}

func (r *provenanceRepoPG) Create(ctx context.Context, p *Provenance) error {
	p.ID = uuid.New()
	if p.FHIRID == "" {
		p.FHIRID = p.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO provenance (id, fhir_id, target_type, target_id, recorded,
			activity_code, activity_display, reason_code, reason_display)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		p.ID, p.FHIRID, p.TargetType, p.TargetID, p.Recorded,
		p.ActivityCode, p.ActivityDisplay, p.ReasonCode, p.ReasonDisplay)
	return err
}

func (r *provenanceRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Provenance, error) {
	return r.scanProv(r.conn(ctx).QueryRow(ctx, `SELECT `+provCols+` FROM provenance WHERE id = $1`, id))
}

func (r *provenanceRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Provenance, error) {
	return r.scanProv(r.conn(ctx).QueryRow(ctx, `SELECT `+provCols+` FROM provenance WHERE fhir_id = $1`, fhirID))
}

func (r *provenanceRepoPG) Update(ctx context.Context, p *Provenance) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE provenance SET target_type=$2, target_id=$3, activity_code=$4,
			activity_display=$5, reason_code=$6, reason_display=$7, updated_at=NOW()
		WHERE id = $1`,
		p.ID, p.TargetType, p.TargetID, p.ActivityCode,
		p.ActivityDisplay, p.ReasonCode, p.ReasonDisplay)
	return err
}

func (r *provenanceRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM provenance WHERE id = $1`, id)
	return err
}

func (r *provenanceRepoPG) List(ctx context.Context, limit, offset int) ([]*Provenance, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM provenance`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+provCols+` FROM provenance ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Provenance
	for rows.Next() {
		p, err := r.scanProv(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, nil
}

func (r *provenanceRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Provenance, int, error) {
	query := `SELECT ` + provCols + ` FROM provenance WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM provenance WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["target"]; ok {
		query += fmt.Sprintf(` AND (target_type || '/' || target_id) = $%d`, idx)
		countQuery += fmt.Sprintf(` AND (target_type || '/' || target_id) = $%d`, idx)
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
	var items []*Provenance
	for rows.Next() {
		p, err := r.scanProv(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, nil
}

func (r *provenanceRepoPG) AddAgent(ctx context.Context, a *ProvenanceAgent) error {
	a.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO provenance_agent (id, provenance_id, type_code, type_display,
			who_type, who_id, on_behalf_of_type, on_behalf_of_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		a.ID, a.ProvenanceID, a.TypeCode, a.TypeDisplay,
		a.WhoType, a.WhoID, a.OnBehalfOfType, a.OnBehalfOfID)
	return err
}

func (r *provenanceRepoPG) GetAgents(ctx context.Context, provenanceID uuid.UUID) ([]*ProvenanceAgent, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, provenance_id, type_code, type_display, who_type, who_id,
			on_behalf_of_type, on_behalf_of_id
		FROM provenance_agent WHERE provenance_id = $1`, provenanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ProvenanceAgent
	for rows.Next() {
		var a ProvenanceAgent
		if err := rows.Scan(&a.ID, &a.ProvenanceID, &a.TypeCode, &a.TypeDisplay,
			&a.WhoType, &a.WhoID, &a.OnBehalfOfType, &a.OnBehalfOfID); err != nil {
			return nil, err
		}
		items = append(items, &a)
	}
	return items, nil
}

func (r *provenanceRepoPG) AddEntity(ctx context.Context, e *ProvenanceEntity) error {
	e.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO provenance_entity (id, provenance_id, role, what_type, what_id)
		VALUES ($1,$2,$3,$4,$5)`,
		e.ID, e.ProvenanceID, e.Role, e.WhatType, e.WhatID)
	return err
}

func (r *provenanceRepoPG) GetEntities(ctx context.Context, provenanceID uuid.UUID) ([]*ProvenanceEntity, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, provenance_id, role, what_type, what_id
		FROM provenance_entity WHERE provenance_id = $1`, provenanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ProvenanceEntity
	for rows.Next() {
		var e ProvenanceEntity
		if err := rows.Scan(&e.ID, &e.ProvenanceID, &e.Role, &e.WhatType, &e.WhatID); err != nil {
			return nil, err
		}
		items = append(items, &e)
	}
	return items, nil
}
