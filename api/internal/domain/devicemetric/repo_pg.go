package devicemetric

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

type deviceMetricRepoPG struct{ pool *pgxpool.Pool }

func NewDeviceMetricRepoPG(pool *pgxpool.Pool) DeviceMetricRepository {
	return &deviceMetricRepoPG{pool: pool}
}

func (r *deviceMetricRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const dmCols = `id, fhir_id, type_code, type_display,
	source_id, parent_id, unit_code, unit_display,
	operational_status, color, category,
	calibration_type, calibration_state, calibration_time,
	measurement_period_value, measurement_period_unit,
	version_id, created_at, updated_at`

func (r *deviceMetricRepoPG) scanRow(row pgx.Row) (*DeviceMetric, error) {
	var m DeviceMetric
	err := row.Scan(&m.ID, &m.FHIRID, &m.TypeCode, &m.TypeDisplay,
		&m.SourceID, &m.ParentID, &m.UnitCode, &m.UnitDisplay,
		&m.OperationalStatus, &m.Color, &m.Category,
		&m.CalibrationType, &m.CalibrationState, &m.CalibrationTime,
		&m.MeasurementPeriodValue, &m.MeasurementPeriodUnit,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *deviceMetricRepoPG) Create(ctx context.Context, m *DeviceMetric) error {
	m.ID = uuid.New()
	if m.FHIRID == "" {
		m.FHIRID = m.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO device_metric (id, fhir_id, type_code, type_display,
			source_id, parent_id, unit_code, unit_display,
			operational_status, color, category,
			calibration_type, calibration_state, calibration_time,
			measurement_period_value, measurement_period_unit)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		m.ID, m.FHIRID, m.TypeCode, m.TypeDisplay,
		m.SourceID, m.ParentID, m.UnitCode, m.UnitDisplay,
		m.OperationalStatus, m.Color, m.Category,
		m.CalibrationType, m.CalibrationState, m.CalibrationTime,
		m.MeasurementPeriodValue, m.MeasurementPeriodUnit)
	return err
}

func (r *deviceMetricRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*DeviceMetric, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+dmCols+` FROM device_metric WHERE id = $1`, id))
}

func (r *deviceMetricRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*DeviceMetric, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+dmCols+` FROM device_metric WHERE fhir_id = $1`, fhirID))
}

func (r *deviceMetricRepoPG) Update(ctx context.Context, m *DeviceMetric) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE device_metric SET type_code=$2, type_display=$3,
			source_id=$4, parent_id=$5, unit_code=$6, unit_display=$7,
			operational_status=$8, color=$9, category=$10,
			calibration_type=$11, calibration_state=$12, calibration_time=$13,
			measurement_period_value=$14, measurement_period_unit=$15, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.TypeCode, m.TypeDisplay,
		m.SourceID, m.ParentID, m.UnitCode, m.UnitDisplay,
		m.OperationalStatus, m.Color, m.Category,
		m.CalibrationType, m.CalibrationState, m.CalibrationTime,
		m.MeasurementPeriodValue, m.MeasurementPeriodUnit)
	return err
}

func (r *deviceMetricRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM device_metric WHERE id = $1`, id)
	return err
}

func (r *deviceMetricRepoPG) List(ctx context.Context, limit, offset int) ([]*DeviceMetric, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM device_metric`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+dmCols+` FROM device_metric ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*DeviceMetric
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}

func (r *deviceMetricRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*DeviceMetric, int, error) {
	query := `SELECT ` + dmCols + ` FROM device_metric WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM device_metric WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["source"]; ok {
		query += fmt.Sprintf(` AND source_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND source_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["type"]; ok {
		query += fmt.Sprintf(` AND type_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND type_code = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["category"]; ok {
		query += fmt.Sprintf(` AND category = $%d`, idx)
		countQuery += fmt.Sprintf(` AND category = $%d`, idx)
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
	var items []*DeviceMetric
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}
