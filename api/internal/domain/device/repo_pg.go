package device

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

type deviceRepoPG struct{ pool *pgxpool.Pool }

func NewDeviceRepoPG(pool *pgxpool.Pool) DeviceRepository {
	return &deviceRepoPG{pool: pool}
}

func (r *deviceRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const deviceCols = `id, fhir_id, status, status_reason, distinct_identifier,
	manufacturer_name, manufacture_date, expiration_date, lot_number,
	serial_number, model_number, device_name, device_name_type,
	type_code, type_display, type_system, version_value,
	patient_id, owner_id, location_id, contact_phone, contact_email,
	url, note, safety_code, safety_display, udi_carrier, udi_entry_type,
	created_at, updated_at`

func (r *deviceRepoPG) scanDevice(row pgx.Row) (*Device, error) {
	var d Device
	err := row.Scan(&d.ID, &d.FHIRID, &d.Status, &d.StatusReason,
		&d.DistinctIdentifier, &d.ManufacturerName, &d.ManufactureDate,
		&d.ExpirationDate, &d.LotNumber, &d.SerialNumber, &d.ModelNumber,
		&d.DeviceName, &d.DeviceNameType, &d.TypeCode, &d.TypeDisplay,
		&d.TypeSystem, &d.VersionValue, &d.PatientID, &d.OwnerID,
		&d.LocationID, &d.ContactPhone, &d.ContactEmail, &d.URL,
		&d.Note, &d.SafetyCode, &d.SafetyDisplay, &d.UDICarrier,
		&d.UDIEntryType, &d.CreatedAt, &d.UpdatedAt)
	return &d, err
}

func (r *deviceRepoPG) Create(ctx context.Context, d *Device) error {
	d.ID = uuid.New()
	if d.FHIRID == "" {
		d.FHIRID = d.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO device (id, fhir_id, status, status_reason, distinct_identifier,
			manufacturer_name, manufacture_date, expiration_date, lot_number,
			serial_number, model_number, device_name, device_name_type,
			type_code, type_display, type_system, version_value,
			patient_id, owner_id, location_id, contact_phone, contact_email,
			url, note, safety_code, safety_display, udi_carrier, udi_entry_type)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28)`,
		d.ID, d.FHIRID, d.Status, d.StatusReason, d.DistinctIdentifier,
		d.ManufacturerName, d.ManufactureDate, d.ExpirationDate, d.LotNumber,
		d.SerialNumber, d.ModelNumber, d.DeviceName, d.DeviceNameType,
		d.TypeCode, d.TypeDisplay, d.TypeSystem, d.VersionValue,
		d.PatientID, d.OwnerID, d.LocationID, d.ContactPhone, d.ContactEmail,
		d.URL, d.Note, d.SafetyCode, d.SafetyDisplay, d.UDICarrier, d.UDIEntryType)
	return err
}

func (r *deviceRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Device, error) {
	return r.scanDevice(r.conn(ctx).QueryRow(ctx, `SELECT `+deviceCols+` FROM device WHERE id = $1`, id))
}

func (r *deviceRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Device, error) {
	return r.scanDevice(r.conn(ctx).QueryRow(ctx, `SELECT `+deviceCols+` FROM device WHERE fhir_id = $1`, fhirID))
}

func (r *deviceRepoPG) Update(ctx context.Context, d *Device) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE device SET status=$2, status_reason=$3, manufacturer_name=$4,
			device_name=$5, serial_number=$6, model_number=$7, note=$8, updated_at=NOW()
		WHERE id = $1`,
		d.ID, d.Status, d.StatusReason, d.ManufacturerName,
		d.DeviceName, d.SerialNumber, d.ModelNumber, d.Note)
	return err
}

func (r *deviceRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM device WHERE id = $1`, id)
	return err
}

func (r *deviceRepoPG) List(ctx context.Context, limit, offset int) ([]*Device, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM device`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+deviceCols+` FROM device ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Device
	for rows.Next() {
		d, err := r.scanDevice(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}

func (r *deviceRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Device, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM device WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+deviceCols+` FROM device WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Device
	for rows.Next() {
		d, err := r.scanDevice(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}

func (r *deviceRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Device, int, error) {
	query := `SELECT ` + deviceCols + ` FROM device WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM device WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient"]; ok {
		query += fmt.Sprintf(` AND patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND patient_id = $%d`, idx)
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
	if p, ok := params["manufacturer"]; ok {
		query += fmt.Sprintf(` AND manufacturer_name ILIKE '%%' || $%d || '%%'`, idx)
		countQuery += fmt.Sprintf(` AND manufacturer_name ILIKE '%%' || $%d || '%%'`, idx)
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
	var items []*Device
	for rows.Next() {
		d, err := r.scanDevice(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}
