package visionprescription

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

type visionPrescriptionRepoPG struct{ pool *pgxpool.Pool }

func NewVisionPrescriptionRepoPG(pool *pgxpool.Pool) VisionPrescriptionRepository {
	return &visionPrescriptionRepoPG{pool: pool}
}

func (r *visionPrescriptionRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const vpCols = `id, fhir_id, status, created, patient_id, encounter_id,
	date_written, prescriber_id, version_id, created_at, updated_at`

func (r *visionPrescriptionRepoPG) scanVP(row pgx.Row) (*VisionPrescription, error) {
	var v VisionPrescription
	err := row.Scan(&v.ID, &v.FHIRID, &v.Status, &v.Created,
		&v.PatientID, &v.EncounterID, &v.DateWritten, &v.PrescriberID,
		&v.VersionID, &v.CreatedAt, &v.UpdatedAt)
	return &v, err
}

func (r *visionPrescriptionRepoPG) Create(ctx context.Context, v *VisionPrescription) error {
	v.ID = uuid.New()
	if v.FHIRID == "" {
		v.FHIRID = v.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO vision_prescription (id, fhir_id, status, created, patient_id,
			encounter_id, date_written, prescriber_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		v.ID, v.FHIRID, v.Status, v.Created, v.PatientID,
		v.EncounterID, v.DateWritten, v.PrescriberID)
	return err
}

func (r *visionPrescriptionRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*VisionPrescription, error) {
	return r.scanVP(r.conn(ctx).QueryRow(ctx, `SELECT `+vpCols+` FROM vision_prescription WHERE id = $1`, id))
}

func (r *visionPrescriptionRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*VisionPrescription, error) {
	return r.scanVP(r.conn(ctx).QueryRow(ctx, `SELECT `+vpCols+` FROM vision_prescription WHERE fhir_id = $1`, fhirID))
}

func (r *visionPrescriptionRepoPG) Update(ctx context.Context, v *VisionPrescription) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE vision_prescription SET status=$2, created=$3, encounter_id=$4,
			date_written=$5, prescriber_id=$6, updated_at=NOW()
		WHERE id = $1`,
		v.ID, v.Status, v.Created, v.EncounterID,
		v.DateWritten, v.PrescriberID)
	return err
}

func (r *visionPrescriptionRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM vision_prescription WHERE id = $1`, id)
	return err
}

func (r *visionPrescriptionRepoPG) List(ctx context.Context, limit, offset int) ([]*VisionPrescription, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM vision_prescription`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+vpCols+` FROM vision_prescription ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*VisionPrescription
	for rows.Next() {
		v, err := r.scanVP(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, v)
	}
	return items, total, nil
}

var vpSearchParams = map[string]fhir.SearchParamConfig{
	"patient": {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
}

func (r *visionPrescriptionRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*VisionPrescription, int, error) {
	qb := fhir.NewSearchQuery("vision_prescription", vpCols)
	qb.ApplyParams(params, vpSearchParams)
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
	var items []*VisionPrescription
	for rows.Next() {
		v, err := r.scanVP(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, v)
	}
	return items, total, nil
}

// -- Lens Spec --

const lsCols = `id, prescription_id, product_code, product_display, eye,
	sphere, cylinder, axis, prism_amount, prism_base, add_power,
	power, back_curve, diameter, duration_value, duration_unit,
	color, brand, note`

func (r *visionPrescriptionRepoPG) scanLensSpec(row pgx.Row) (*VisionPrescriptionLensSpec, error) {
	var ls VisionPrescriptionLensSpec
	err := row.Scan(&ls.ID, &ls.PrescriptionID, &ls.ProductCode, &ls.ProductDisplay,
		&ls.Eye, &ls.Sphere, &ls.Cylinder, &ls.Axis, &ls.PrismAmount,
		&ls.PrismBase, &ls.AddPower, &ls.Power, &ls.BackCurve,
		&ls.Diameter, &ls.DurationValue, &ls.DurationUnit,
		&ls.Color, &ls.Brand, &ls.Note)
	return &ls, err
}

func (r *visionPrescriptionRepoPG) AddLensSpec(ctx context.Context, ls *VisionPrescriptionLensSpec) error {
	ls.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO vision_prescription_lens_spec (id, prescription_id, product_code, product_display,
			eye, sphere, cylinder, axis, prism_amount, prism_base, add_power,
			power, back_curve, diameter, duration_value, duration_unit,
			color, brand, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		ls.ID, ls.PrescriptionID, ls.ProductCode, ls.ProductDisplay,
		ls.Eye, ls.Sphere, ls.Cylinder, ls.Axis, ls.PrismAmount,
		ls.PrismBase, ls.AddPower, ls.Power, ls.BackCurve,
		ls.Diameter, ls.DurationValue, ls.DurationUnit,
		ls.Color, ls.Brand, ls.Note)
	return err
}

func (r *visionPrescriptionRepoPG) GetLensSpecs(ctx context.Context, prescriptionID uuid.UUID) ([]*VisionPrescriptionLensSpec, error) {
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+lsCols+` FROM vision_prescription_lens_spec WHERE prescription_id = $1 ORDER BY id`, prescriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*VisionPrescriptionLensSpec
	for rows.Next() {
		ls, err := r.scanLensSpec(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, ls)
	}
	return items, nil
}
