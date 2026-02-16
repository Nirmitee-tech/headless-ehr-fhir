package devicedefinition

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

type deviceDefinitionRepoPG struct{ pool *pgxpool.Pool }

func NewDeviceDefinitionRepoPG(pool *pgxpool.Pool) DeviceDefinitionRepository {
	return &deviceDefinitionRepoPG{pool: pool}
}

func (r *deviceDefinitionRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const ddCols = `id, fhir_id, manufacturer_string, model_number,
	device_name, device_name_type, type_code, type_display,
	specialization, safety_code, safety_display,
	owner_id, parent_device_id, description,
	version_id, created_at, updated_at`

func (r *deviceDefinitionRepoPG) scanRow(row pgx.Row) (*DeviceDefinition, error) {
	var d DeviceDefinition
	err := row.Scan(&d.ID, &d.FHIRID, &d.ManufacturerString, &d.ModelNumber,
		&d.DeviceName, &d.DeviceNameType, &d.TypeCode, &d.TypeDisplay,
		&d.Specialization, &d.SafetyCode, &d.SafetyDisplay,
		&d.OwnerID, &d.ParentDeviceID, &d.Description,
		&d.VersionID, &d.CreatedAt, &d.UpdatedAt)
	return &d, err
}

func (r *deviceDefinitionRepoPG) Create(ctx context.Context, d *DeviceDefinition) error {
	d.ID = uuid.New()
	if d.FHIRID == "" {
		d.FHIRID = d.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO device_definition (id, fhir_id, manufacturer_string, model_number,
			device_name, device_name_type, type_code, type_display,
			specialization, safety_code, safety_display,
			owner_id, parent_device_id, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		d.ID, d.FHIRID, d.ManufacturerString, d.ModelNumber,
		d.DeviceName, d.DeviceNameType, d.TypeCode, d.TypeDisplay,
		d.Specialization, d.SafetyCode, d.SafetyDisplay,
		d.OwnerID, d.ParentDeviceID, d.Description)
	return err
}

func (r *deviceDefinitionRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*DeviceDefinition, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+ddCols+` FROM device_definition WHERE id = $1`, id))
}

func (r *deviceDefinitionRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*DeviceDefinition, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+ddCols+` FROM device_definition WHERE fhir_id = $1`, fhirID))
}

func (r *deviceDefinitionRepoPG) Update(ctx context.Context, d *DeviceDefinition) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE device_definition SET manufacturer_string=$2, model_number=$3,
			device_name=$4, device_name_type=$5, type_code=$6, type_display=$7,
			specialization=$8, safety_code=$9, safety_display=$10,
			owner_id=$11, parent_device_id=$12, description=$13, updated_at=NOW()
		WHERE id = $1`,
		d.ID, d.ManufacturerString, d.ModelNumber,
		d.DeviceName, d.DeviceNameType, d.TypeCode, d.TypeDisplay,
		d.Specialization, d.SafetyCode, d.SafetyDisplay,
		d.OwnerID, d.ParentDeviceID, d.Description)
	return err
}

func (r *deviceDefinitionRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM device_definition WHERE id = $1`, id)
	return err
}

func (r *deviceDefinitionRepoPG) List(ctx context.Context, limit, offset int) ([]*DeviceDefinition, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM device_definition`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+ddCols+` FROM device_definition ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*DeviceDefinition
	for rows.Next() {
		d, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}

var ddSearchParams = map[string]fhir.SearchParamConfig{
	"manufacturer": {Type: fhir.SearchParamString, Column: "manufacturer_string"},
	"model-number": {Type: fhir.SearchParamString, Column: "model_number"},
	"type":         {Type: fhir.SearchParamToken, Column: "type_code"},
}

func (r *deviceDefinitionRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*DeviceDefinition, int, error) {
	qb := fhir.NewSearchQuery("device_definition", ddCols)
	qb.ApplyParams(params, ddSearchParams)
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
	var items []*DeviceDefinition
	for rows.Next() {
		d, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}
