package molecularsequence

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

type molecularSequenceRepoPG struct{ pool *pgxpool.Pool }

func NewMolecularSequenceRepoPG(pool *pgxpool.Pool) MolecularSequenceRepository {
	return &molecularSequenceRepoPG{pool: pool}
}

func (r *molecularSequenceRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const msCols = `id, fhir_id, type, patient_id, specimen_id, device_id, performer_id,
	coordinate_system, observed_seq, reference_seq_id, reference_seq_strand,
	window_start, window_end,
	version_id, created_at, updated_at`

func (r *molecularSequenceRepoPG) scanRow(row pgx.Row) (*MolecularSequence, error) {
	var m MolecularSequence
	err := row.Scan(&m.ID, &m.FHIRID, &m.Type, &m.PatientID, &m.SpecimenID, &m.DeviceID, &m.PerformerID,
		&m.CoordinateSystem, &m.ObservedSeq, &m.ReferenceSeqID, &m.ReferenceSeqStrand,
		&m.WindowStart, &m.WindowEnd,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *molecularSequenceRepoPG) Create(ctx context.Context, m *MolecularSequence) error {
	m.ID = uuid.New()
	if m.FHIRID == "" {
		m.FHIRID = m.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO molecular_sequence (id, fhir_id, type, patient_id, specimen_id, device_id, performer_id,
			coordinate_system, observed_seq, reference_seq_id, reference_seq_strand,
			window_start, window_end)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		m.ID, m.FHIRID, m.Type, m.PatientID, m.SpecimenID, m.DeviceID, m.PerformerID,
		m.CoordinateSystem, m.ObservedSeq, m.ReferenceSeqID, m.ReferenceSeqStrand,
		m.WindowStart, m.WindowEnd)
	return err
}

func (r *molecularSequenceRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MolecularSequence, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+msCols+` FROM molecular_sequence WHERE id = $1`, id))
}

func (r *molecularSequenceRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MolecularSequence, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+msCols+` FROM molecular_sequence WHERE fhir_id = $1`, fhirID))
}

func (r *molecularSequenceRepoPG) Update(ctx context.Context, m *MolecularSequence) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE molecular_sequence SET type=$2, patient_id=$3, specimen_id=$4, device_id=$5, performer_id=$6,
			coordinate_system=$7, observed_seq=$8, reference_seq_id=$9, reference_seq_strand=$10,
			window_start=$11, window_end=$12, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.Type, m.PatientID, m.SpecimenID, m.DeviceID, m.PerformerID,
		m.CoordinateSystem, m.ObservedSeq, m.ReferenceSeqID, m.ReferenceSeqStrand,
		m.WindowStart, m.WindowEnd)
	return err
}

func (r *molecularSequenceRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM molecular_sequence WHERE id = $1`, id)
	return err
}

func (r *molecularSequenceRepoPG) List(ctx context.Context, limit, offset int) ([]*MolecularSequence, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM molecular_sequence`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+msCols+` FROM molecular_sequence ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MolecularSequence
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}

func (r *molecularSequenceRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MolecularSequence, int, error) {
	query := `SELECT ` + msCols + ` FROM molecular_sequence WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM molecular_sequence WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["type"]; ok {
		query += fmt.Sprintf(` AND type = $%d`, idx)
		countQuery += fmt.Sprintf(` AND type = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["patient"]; ok {
		query += fmt.Sprintf(` AND patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND patient_id = $%d`, idx)
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
	var items []*MolecularSequence
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}
