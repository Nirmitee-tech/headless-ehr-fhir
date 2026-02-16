package specimendefinition

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

type specimenDefinitionRepoPG struct{ pool *pgxpool.Pool }

func NewSpecimenDefinitionRepoPG(pool *pgxpool.Pool) SpecimenDefinitionRepository {
	return &specimenDefinitionRepoPG{pool: pool}
}

func (r *specimenDefinitionRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const sdCols = `id, fhir_id, type_code, type_display,
	patient_preparation, time_aspect,
	collection_code, collection_display,
	handling_temperature_low, handling_temperature_high, handling_temperature_unit,
	handling_max_duration, handling_instruction,
	version_id, created_at, updated_at`

func (r *specimenDefinitionRepoPG) scanRow(row pgx.Row) (*SpecimenDefinition, error) {
	var s SpecimenDefinition
	err := row.Scan(&s.ID, &s.FHIRID, &s.TypeCode, &s.TypeDisplay,
		&s.PatientPreparation, &s.TimeAspect,
		&s.CollectionCode, &s.CollectionDisplay,
		&s.HandlingTemperatureLow, &s.HandlingTemperatureHigh, &s.HandlingTemperatureUnit,
		&s.HandlingMaxDuration, &s.HandlingInstruction,
		&s.VersionID, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *specimenDefinitionRepoPG) Create(ctx context.Context, s *SpecimenDefinition) error {
	s.ID = uuid.New()
	if s.FHIRID == "" {
		s.FHIRID = s.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO specimen_definition (id, fhir_id, type_code, type_display,
			patient_preparation, time_aspect,
			collection_code, collection_display,
			handling_temperature_low, handling_temperature_high, handling_temperature_unit,
			handling_max_duration, handling_instruction)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		s.ID, s.FHIRID, s.TypeCode, s.TypeDisplay,
		s.PatientPreparation, s.TimeAspect,
		s.CollectionCode, s.CollectionDisplay,
		s.HandlingTemperatureLow, s.HandlingTemperatureHigh, s.HandlingTemperatureUnit,
		s.HandlingMaxDuration, s.HandlingInstruction)
	return err
}

func (r *specimenDefinitionRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*SpecimenDefinition, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+sdCols+` FROM specimen_definition WHERE id = $1`, id))
}

func (r *specimenDefinitionRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*SpecimenDefinition, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+sdCols+` FROM specimen_definition WHERE fhir_id = $1`, fhirID))
}

func (r *specimenDefinitionRepoPG) Update(ctx context.Context, s *SpecimenDefinition) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE specimen_definition SET type_code=$2, type_display=$3,
			patient_preparation=$4, time_aspect=$5,
			collection_code=$6, collection_display=$7,
			handling_temperature_low=$8, handling_temperature_high=$9, handling_temperature_unit=$10,
			handling_max_duration=$11, handling_instruction=$12, updated_at=NOW()
		WHERE id = $1`,
		s.ID, s.TypeCode, s.TypeDisplay,
		s.PatientPreparation, s.TimeAspect,
		s.CollectionCode, s.CollectionDisplay,
		s.HandlingTemperatureLow, s.HandlingTemperatureHigh, s.HandlingTemperatureUnit,
		s.HandlingMaxDuration, s.HandlingInstruction)
	return err
}

func (r *specimenDefinitionRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM specimen_definition WHERE id = $1`, id)
	return err
}

func (r *specimenDefinitionRepoPG) List(ctx context.Context, limit, offset int) ([]*SpecimenDefinition, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM specimen_definition`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+sdCols+` FROM specimen_definition ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*SpecimenDefinition
	for rows.Next() {
		s, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

func (r *specimenDefinitionRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SpecimenDefinition, int, error) {
	query := `SELECT ` + sdCols + ` FROM specimen_definition WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM specimen_definition WHERE 1=1`
	var args []interface{}
	idx := 1

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
	var items []*SpecimenDefinition
	for rows.Next() {
		s, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}
