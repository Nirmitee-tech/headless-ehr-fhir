package biologicallyderivedproduct

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

type bdpRepoPG struct{ pool *pgxpool.Pool }

func NewBiologicallyDerivedProductRepoPG(pool *pgxpool.Pool) BiologicallyDerivedProductRepository {
	return &bdpRepoPG{pool: pool}
}

func (r *bdpRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const bdpCols = `id, fhir_id, product_category, product_code_code, product_code_display,
	status, request_id, quantity, parent_id,
	collection_source_type, collection_source_reference, collection_collected_date,
	processing_description, storage_temperature_code, storage_duration,
	version_id, created_at, updated_at`

func (r *bdpRepoPG) scanRow(row pgx.Row) (*BiologicallyDerivedProduct, error) {
	var b BiologicallyDerivedProduct
	err := row.Scan(&b.ID, &b.FHIRID, &b.ProductCategory, &b.ProductCodeCode, &b.ProductCodeDisplay,
		&b.Status, &b.RequestID, &b.Quantity, &b.ParentID,
		&b.CollectionSourceType, &b.CollectionSourceRef, &b.CollectionCollectedDate,
		&b.ProcessingDescription, &b.StorageTemperatureCode, &b.StorageDuration,
		&b.VersionID, &b.CreatedAt, &b.UpdatedAt)
	return &b, err
}

func (r *bdpRepoPG) Create(ctx context.Context, b *BiologicallyDerivedProduct) error {
	b.ID = uuid.New()
	if b.FHIRID == "" {
		b.FHIRID = b.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO biologically_derived_product (id, fhir_id, product_category, product_code_code, product_code_display,
			status, request_id, quantity, parent_id,
			collection_source_type, collection_source_reference, collection_collected_date,
			processing_description, storage_temperature_code, storage_duration)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		b.ID, b.FHIRID, b.ProductCategory, b.ProductCodeCode, b.ProductCodeDisplay,
		b.Status, b.RequestID, b.Quantity, b.ParentID,
		b.CollectionSourceType, b.CollectionSourceRef, b.CollectionCollectedDate,
		b.ProcessingDescription, b.StorageTemperatureCode, b.StorageDuration)
	return err
}

func (r *bdpRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*BiologicallyDerivedProduct, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+bdpCols+` FROM biologically_derived_product WHERE id = $1`, id))
}

func (r *bdpRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*BiologicallyDerivedProduct, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+bdpCols+` FROM biologically_derived_product WHERE fhir_id = $1`, fhirID))
}

func (r *bdpRepoPG) Update(ctx context.Context, b *BiologicallyDerivedProduct) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE biologically_derived_product SET product_category=$2, product_code_code=$3, product_code_display=$4,
			status=$5, request_id=$6, quantity=$7, parent_id=$8,
			collection_source_type=$9, collection_source_reference=$10, collection_collected_date=$11,
			processing_description=$12, storage_temperature_code=$13, storage_duration=$14, updated_at=NOW()
		WHERE id = $1`,
		b.ID, b.ProductCategory, b.ProductCodeCode, b.ProductCodeDisplay,
		b.Status, b.RequestID, b.Quantity, b.ParentID,
		b.CollectionSourceType, b.CollectionSourceRef, b.CollectionCollectedDate,
		b.ProcessingDescription, b.StorageTemperatureCode, b.StorageDuration)
	return err
}

func (r *bdpRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM biologically_derived_product WHERE id = $1`, id)
	return err
}

func (r *bdpRepoPG) List(ctx context.Context, limit, offset int) ([]*BiologicallyDerivedProduct, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM biologically_derived_product`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+bdpCols+` FROM biologically_derived_product ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*BiologicallyDerivedProduct
	for rows.Next() {
		b, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, b)
	}
	return items, total, nil
}

var bdpSearchParams = map[string]fhir.SearchParamConfig{
	"product-category": {Type: fhir.SearchParamToken, Column: "product_category"},
	"status":           {Type: fhir.SearchParamToken, Column: "status"},
}

func (r *bdpRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*BiologicallyDerivedProduct, int, error) {
	qb := fhir.NewSearchQuery("biologically_derived_product", bdpCols)
	qb.ApplyParams(params, bdpSearchParams)
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
	var items []*BiologicallyDerivedProduct
	for rows.Next() {
		b, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, b)
	}
	return items, total, nil
}
