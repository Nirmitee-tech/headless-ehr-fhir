package substancesourcematerial

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

type ssmRepoPG struct{ pool *pgxpool.Pool }

func NewSubstanceSourceMaterialRepoPG(pool *pgxpool.Pool) SubstanceSourceMaterialRepository {
	return &ssmRepoPG{pool: pool}
}

func (r *ssmRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil { return tx }
	if c := db.ConnFromContext(ctx); c != nil { return c }
	return r.pool
}

const ssmCols = `id, fhir_id, source_material_class_code, source_material_class_display,
	source_material_type_code, source_material_type_display,
	source_material_state_code, source_material_state_display,
	organism_id, organism_name, country_of_origin_code, country_of_origin_display,
	geographical_location,
	version_id, created_at, updated_at`

func (r *ssmRepoPG) scanRow(row pgx.Row) (*SubstanceSourceMaterial, error) {
	var m SubstanceSourceMaterial
	err := row.Scan(&m.ID, &m.FHIRID, &m.SourceMaterialClassCode, &m.SourceMaterialClassDisplay,
		&m.SourceMaterialTypeCode, &m.SourceMaterialTypeDisplay,
		&m.SourceMaterialStateCode, &m.SourceMaterialStateDisplay,
		&m.OrganismID, &m.OrganismName, &m.CountryOfOriginCode, &m.CountryOfOriginDisplay,
		&m.GeographicalLocation,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *ssmRepoPG) Create(ctx context.Context, m *SubstanceSourceMaterial) error {
	m.ID = uuid.New()
	if m.FHIRID == "" { m.FHIRID = m.ID.String() }
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO substance_source_material (id, fhir_id, source_material_class_code, source_material_class_display,
			source_material_type_code, source_material_type_display,
			source_material_state_code, source_material_state_display,
			organism_id, organism_name, country_of_origin_code, country_of_origin_display,
			geographical_location)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		m.ID, m.FHIRID, m.SourceMaterialClassCode, m.SourceMaterialClassDisplay,
		m.SourceMaterialTypeCode, m.SourceMaterialTypeDisplay,
		m.SourceMaterialStateCode, m.SourceMaterialStateDisplay,
		m.OrganismID, m.OrganismName, m.CountryOfOriginCode, m.CountryOfOriginDisplay,
		m.GeographicalLocation)
	return err
}

func (r *ssmRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*SubstanceSourceMaterial, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+ssmCols+` FROM substance_source_material WHERE id = $1`, id)) }
func (r *ssmRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*SubstanceSourceMaterial, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+ssmCols+` FROM substance_source_material WHERE fhir_id = $1`, fhirID)) }

func (r *ssmRepoPG) Update(ctx context.Context, m *SubstanceSourceMaterial) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE substance_source_material SET source_material_class_code=$2, source_material_class_display=$3,
			source_material_type_code=$4, source_material_type_display=$5,
			source_material_state_code=$6, source_material_state_display=$7,
			organism_id=$8, organism_name=$9, country_of_origin_code=$10, country_of_origin_display=$11,
			geographical_location=$12, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.SourceMaterialClassCode, m.SourceMaterialClassDisplay,
		m.SourceMaterialTypeCode, m.SourceMaterialTypeDisplay,
		m.SourceMaterialStateCode, m.SourceMaterialStateDisplay,
		m.OrganismID, m.OrganismName, m.CountryOfOriginCode, m.CountryOfOriginDisplay,
		m.GeographicalLocation)
	return err
}

func (r *ssmRepoPG) Delete(ctx context.Context, id uuid.UUID) error { _, err := r.conn(ctx).Exec(ctx, `DELETE FROM substance_source_material WHERE id = $1`, id); return err }

func (r *ssmRepoPG) List(ctx context.Context, limit, offset int) ([]*SubstanceSourceMaterial, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM substance_source_material`).Scan(&total); err != nil { return nil, 0, err }
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+ssmCols+` FROM substance_source_material ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, err }; defer rows.Close()
	var items []*SubstanceSourceMaterial
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}

func (r *ssmRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SubstanceSourceMaterial, int, error) {
	query := `SELECT ` + ssmCols + ` FROM substance_source_material WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM substance_source_material WHERE 1=1`
	var args []interface{}; idx := 1
	_ = params
	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil { return nil, 0, err }
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1); args = append(args, limit, offset)
	rows, err := r.conn(ctx).Query(ctx, query, args...); if err != nil { return nil, 0, err }; defer rows.Close()
	var items []*SubstanceSourceMaterial
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}
