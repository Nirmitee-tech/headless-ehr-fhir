package healthcareservice

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

type healthcareServiceRepoPG struct{ pool *pgxpool.Pool }

func NewHealthcareServiceRepoPG(pool *pgxpool.Pool) HealthcareServiceRepository {
	return &healthcareServiceRepoPG{pool: pool}
}

func (r *healthcareServiceRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const hsCols = `id, fhir_id, active, provided_by_org_id,
	category_code, category_display, type_code, type_display,
	name, comment, telecom_phone, telecom_email,
	service_provision_code, program_name, location_id,
	appointment_required, available_time, not_available,
	availability_exceptions, version_id, created_at, updated_at`

func (r *healthcareServiceRepoPG) scanRow(row pgx.Row) (*HealthcareService, error) {
	var hs HealthcareService
	err := row.Scan(&hs.ID, &hs.FHIRID, &hs.Active, &hs.ProvidedByOrgID,
		&hs.CategoryCode, &hs.CategoryDisplay, &hs.TypeCode, &hs.TypeDisplay,
		&hs.Name, &hs.Comment, &hs.TelecomPhone, &hs.TelecomEmail,
		&hs.ServiceProvisionCode, &hs.ProgramName, &hs.LocationID,
		&hs.AppointmentRequired, &hs.AvailableTime, &hs.NotAvailable,
		&hs.AvailabilityExceptions, &hs.VersionID, &hs.CreatedAt, &hs.UpdatedAt)
	return &hs, err
}

func (r *healthcareServiceRepoPG) Create(ctx context.Context, hs *HealthcareService) error {
	hs.ID = uuid.New()
	if hs.FHIRID == "" {
		hs.FHIRID = hs.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO healthcare_service (id, fhir_id, active, provided_by_org_id,
			category_code, category_display, type_code, type_display,
			name, comment, telecom_phone, telecom_email,
			service_provision_code, program_name, location_id,
			appointment_required, available_time, not_available,
			availability_exceptions)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		hs.ID, hs.FHIRID, hs.Active, hs.ProvidedByOrgID,
		hs.CategoryCode, hs.CategoryDisplay, hs.TypeCode, hs.TypeDisplay,
		hs.Name, hs.Comment, hs.TelecomPhone, hs.TelecomEmail,
		hs.ServiceProvisionCode, hs.ProgramName, hs.LocationID,
		hs.AppointmentRequired, hs.AvailableTime, hs.NotAvailable,
		hs.AvailabilityExceptions)
	return err
}

func (r *healthcareServiceRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*HealthcareService, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+hsCols+` FROM healthcare_service WHERE id = $1`, id))
}

func (r *healthcareServiceRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*HealthcareService, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+hsCols+` FROM healthcare_service WHERE fhir_id = $1`, fhirID))
}

func (r *healthcareServiceRepoPG) Update(ctx context.Context, hs *HealthcareService) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE healthcare_service SET active=$2, name=$3, comment=$4,
			telecom_phone=$5, telecom_email=$6, service_provision_code=$7,
			appointment_required=$8, availability_exceptions=$9, updated_at=NOW()
		WHERE id = $1`,
		hs.ID, hs.Active, hs.Name, hs.Comment,
		hs.TelecomPhone, hs.TelecomEmail, hs.ServiceProvisionCode,
		hs.AppointmentRequired, hs.AvailabilityExceptions)
	return err
}

func (r *healthcareServiceRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM healthcare_service WHERE id = $1`, id)
	return err
}

func (r *healthcareServiceRepoPG) List(ctx context.Context, limit, offset int) ([]*HealthcareService, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM healthcare_service`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+hsCols+` FROM healthcare_service ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*HealthcareService
	for rows.Next() {
		hs, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, hs)
	}
	return items, total, nil
}

var healthcareServiceSearchParams = map[string]fhir.SearchParamConfig{
	"active":       {Type: fhir.SearchParamToken, Column: "active"},
	"name":         {Type: fhir.SearchParamString, Column: "name"},
	"organization": {Type: fhir.SearchParamReference, Column: "provided_by_org_id"},
}

func (r *healthcareServiceRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*HealthcareService, int, error) {
	qb := fhir.NewSearchQuery("healthcare_service", hsCols)
	qb.ApplyParams(params, healthcareServiceSearchParams)
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
	var items []*HealthcareService
	for rows.Next() {
		hs, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, hs)
	}
	return items, total, nil
}
