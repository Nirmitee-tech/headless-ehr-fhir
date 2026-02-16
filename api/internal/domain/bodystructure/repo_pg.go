package bodystructure

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

type bodyStructureRepoPG struct{ pool *pgxpool.Pool }

func NewBodyStructureRepoPG(pool *pgxpool.Pool) BodyStructureRepository {
	return &bodyStructureRepoPG{pool: pool}
}

func (r *bodyStructureRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const bsCols = `id, fhir_id, active, morphology_code, morphology_display, morphology_system,
	location_code, location_display, location_system,
	location_qualifier_code, location_qualifier_display,
	description, patient_id,
	version_id, created_at, updated_at`

func (r *bodyStructureRepoPG) scanRow(row pgx.Row) (*BodyStructure, error) {
	var b BodyStructure
	err := row.Scan(&b.ID, &b.FHIRID, &b.Active, &b.MorphologyCode, &b.MorphologyDisplay, &b.MorphologySystem,
		&b.LocationCode, &b.LocationDisplay, &b.LocationSystem,
		&b.LocationQualifierCode, &b.LocationQualifierDisplay,
		&b.Description, &b.PatientID,
		&b.VersionID, &b.CreatedAt, &b.UpdatedAt)
	return &b, err
}

func (r *bodyStructureRepoPG) Create(ctx context.Context, b *BodyStructure) error {
	b.ID = uuid.New()
	if b.FHIRID == "" {
		b.FHIRID = b.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO body_structure (id, fhir_id, active, morphology_code, morphology_display, morphology_system,
			location_code, location_display, location_system,
			location_qualifier_code, location_qualifier_display,
			description, patient_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		b.ID, b.FHIRID, b.Active, b.MorphologyCode, b.MorphologyDisplay, b.MorphologySystem,
		b.LocationCode, b.LocationDisplay, b.LocationSystem,
		b.LocationQualifierCode, b.LocationQualifierDisplay,
		b.Description, b.PatientID)
	return err
}

func (r *bodyStructureRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*BodyStructure, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+bsCols+` FROM body_structure WHERE id = $1`, id))
}

func (r *bodyStructureRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*BodyStructure, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+bsCols+` FROM body_structure WHERE fhir_id = $1`, fhirID))
}

func (r *bodyStructureRepoPG) Update(ctx context.Context, b *BodyStructure) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE body_structure SET active=$2, morphology_code=$3, morphology_display=$4, morphology_system=$5,
			location_code=$6, location_display=$7, location_system=$8,
			location_qualifier_code=$9, location_qualifier_display=$10,
			description=$11, updated_at=NOW()
		WHERE id = $1`,
		b.ID, b.Active, b.MorphologyCode, b.MorphologyDisplay, b.MorphologySystem,
		b.LocationCode, b.LocationDisplay, b.LocationSystem,
		b.LocationQualifierCode, b.LocationQualifierDisplay,
		b.Description)
	return err
}

func (r *bodyStructureRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM body_structure WHERE id = $1`, id)
	return err
}

func (r *bodyStructureRepoPG) List(ctx context.Context, limit, offset int) ([]*BodyStructure, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM body_structure`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+bsCols+` FROM body_structure ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*BodyStructure
	for rows.Next() {
		b, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, b)
	}
	return items, total, nil
}

var bodyStructureSearchParams = map[string]fhir.SearchParamConfig{
	"patient": {Type: fhir.SearchParamReference, Column: "patient_id"},
}

func (r *bodyStructureRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*BodyStructure, int, error) {
	qb := fhir.NewSearchQuery("body_structure", bsCols)
	qb.ApplyParams(params, bodyStructureSearchParams)
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
	var items []*BodyStructure
	for rows.Next() {
		b, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, b)
	}
	return items, total, nil
}
