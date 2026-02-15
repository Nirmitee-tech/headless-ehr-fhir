package supply

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

// ---- SupplyRequest Repo ----

type supplyRequestRepoPG struct{ pool *pgxpool.Pool }

func NewSupplyRequestRepoPG(pool *pgxpool.Pool) SupplyRequestRepository {
	return &supplyRequestRepoPG{pool: pool}
}

func (r *supplyRequestRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const supplyRequestCols = `id, fhir_id, status, category_code, category_display,
	priority, item_code, item_display, item_system,
	quantity_value, quantity_unit, occurrence_date, authored_on,
	requester_id, supplier_org_id, deliver_to_location_id,
	reason_code, reason_display,
	created_at, updated_at`

func (r *supplyRequestRepoPG) scanSupplyRequest(row pgx.Row) (*SupplyRequest, error) {
	var s SupplyRequest
	err := row.Scan(&s.ID, &s.FHIRID, &s.Status, &s.CategoryCode, &s.CategoryDisplay,
		&s.Priority, &s.ItemCode, &s.ItemDisplay, &s.ItemSystem,
		&s.QuantityValue, &s.QuantityUnit, &s.OccurrenceDate, &s.AuthoredOn,
		&s.RequesterID, &s.SupplierOrgID, &s.DeliverToLocationID,
		&s.ReasonCode, &s.ReasonDisplay,
		&s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *supplyRequestRepoPG) Create(ctx context.Context, s *SupplyRequest) error {
	s.ID = uuid.New()
	if s.FHIRID == "" {
		s.FHIRID = s.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO supply_request (id, fhir_id, status, category_code, category_display,
			priority, item_code, item_display, item_system,
			quantity_value, quantity_unit, occurrence_date, authored_on,
			requester_id, supplier_org_id, deliver_to_location_id,
			reason_code, reason_display)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		s.ID, s.FHIRID, s.Status, s.CategoryCode, s.CategoryDisplay,
		s.Priority, s.ItemCode, s.ItemDisplay, s.ItemSystem,
		s.QuantityValue, s.QuantityUnit, s.OccurrenceDate, s.AuthoredOn,
		s.RequesterID, s.SupplierOrgID, s.DeliverToLocationID,
		s.ReasonCode, s.ReasonDisplay)
	return err
}

func (r *supplyRequestRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*SupplyRequest, error) {
	return r.scanSupplyRequest(r.conn(ctx).QueryRow(ctx, `SELECT `+supplyRequestCols+` FROM supply_request WHERE id = $1`, id))
}

func (r *supplyRequestRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*SupplyRequest, error) {
	return r.scanSupplyRequest(r.conn(ctx).QueryRow(ctx, `SELECT `+supplyRequestCols+` FROM supply_request WHERE fhir_id = $1`, fhirID))
}

func (r *supplyRequestRepoPG) Update(ctx context.Context, s *SupplyRequest) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE supply_request SET status=$2, category_code=$3, category_display=$4,
			priority=$5, item_code=$6, item_display=$7, item_system=$8,
			quantity_value=$9, quantity_unit=$10, occurrence_date=$11, authored_on=$12,
			requester_id=$13, supplier_org_id=$14, deliver_to_location_id=$15,
			reason_code=$16, reason_display=$17, updated_at=NOW()
		WHERE id = $1`,
		s.ID, s.Status, s.CategoryCode, s.CategoryDisplay,
		s.Priority, s.ItemCode, s.ItemDisplay, s.ItemSystem,
		s.QuantityValue, s.QuantityUnit, s.OccurrenceDate, s.AuthoredOn,
		s.RequesterID, s.SupplierOrgID, s.DeliverToLocationID,
		s.ReasonCode, s.ReasonDisplay)
	return err
}

func (r *supplyRequestRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM supply_request WHERE id = $1`, id)
	return err
}

func (r *supplyRequestRepoPG) List(ctx context.Context, limit, offset int) ([]*SupplyRequest, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM supply_request`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+supplyRequestCols+` FROM supply_request ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*SupplyRequest
	for rows.Next() {
		s, err := r.scanSupplyRequest(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

func (r *supplyRequestRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SupplyRequest, int, error) {
	query := `SELECT ` + supplyRequestCols + ` FROM supply_request WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM supply_request WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["category"]; ok {
		query += fmt.Sprintf(` AND category_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND category_code = $%d`, idx)
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
	var items []*SupplyRequest
	for rows.Next() {
		s, err := r.scanSupplyRequest(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

// ---- SupplyDelivery Repo ----

type supplyDeliveryRepoPG struct{ pool *pgxpool.Pool }

func NewSupplyDeliveryRepoPG(pool *pgxpool.Pool) SupplyDeliveryRepository {
	return &supplyDeliveryRepoPG{pool: pool}
}

func (r *supplyDeliveryRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const supplyDeliveryCols = `id, fhir_id, status, based_on_id, patient_id,
	type_code, type_display, supplied_item_code, supplied_item_display,
	supplied_item_quantity, supplied_item_unit, occurrence_date,
	supplier_id, destination_location_id, receiver_id,
	created_at, updated_at`

func (r *supplyDeliveryRepoPG) scanSupplyDelivery(row pgx.Row) (*SupplyDelivery, error) {
	var s SupplyDelivery
	err := row.Scan(&s.ID, &s.FHIRID, &s.Status, &s.BasedOnID, &s.PatientID,
		&s.TypeCode, &s.TypeDisplay, &s.SuppliedItemCode, &s.SuppliedItemDisplay,
		&s.SuppliedItemQuantity, &s.SuppliedItemUnit, &s.OccurrenceDate,
		&s.SupplierID, &s.DestinationLocationID, &s.ReceiverID,
		&s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *supplyDeliveryRepoPG) Create(ctx context.Context, s *SupplyDelivery) error {
	s.ID = uuid.New()
	if s.FHIRID == "" {
		s.FHIRID = s.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO supply_delivery (id, fhir_id, status, based_on_id, patient_id,
			type_code, type_display, supplied_item_code, supplied_item_display,
			supplied_item_quantity, supplied_item_unit, occurrence_date,
			supplier_id, destination_location_id, receiver_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		s.ID, s.FHIRID, s.Status, s.BasedOnID, s.PatientID,
		s.TypeCode, s.TypeDisplay, s.SuppliedItemCode, s.SuppliedItemDisplay,
		s.SuppliedItemQuantity, s.SuppliedItemUnit, s.OccurrenceDate,
		s.SupplierID, s.DestinationLocationID, s.ReceiverID)
	return err
}

func (r *supplyDeliveryRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*SupplyDelivery, error) {
	return r.scanSupplyDelivery(r.conn(ctx).QueryRow(ctx, `SELECT `+supplyDeliveryCols+` FROM supply_delivery WHERE id = $1`, id))
}

func (r *supplyDeliveryRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*SupplyDelivery, error) {
	return r.scanSupplyDelivery(r.conn(ctx).QueryRow(ctx, `SELECT `+supplyDeliveryCols+` FROM supply_delivery WHERE fhir_id = $1`, fhirID))
}

func (r *supplyDeliveryRepoPG) Update(ctx context.Context, s *SupplyDelivery) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE supply_delivery SET status=$2, based_on_id=$3, patient_id=$4,
			type_code=$5, type_display=$6, supplied_item_code=$7, supplied_item_display=$8,
			supplied_item_quantity=$9, supplied_item_unit=$10, occurrence_date=$11,
			supplier_id=$12, destination_location_id=$13, receiver_id=$14, updated_at=NOW()
		WHERE id = $1`,
		s.ID, s.Status, s.BasedOnID, s.PatientID,
		s.TypeCode, s.TypeDisplay, s.SuppliedItemCode, s.SuppliedItemDisplay,
		s.SuppliedItemQuantity, s.SuppliedItemUnit, s.OccurrenceDate,
		s.SupplierID, s.DestinationLocationID, s.ReceiverID)
	return err
}

func (r *supplyDeliveryRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM supply_delivery WHERE id = $1`, id)
	return err
}

func (r *supplyDeliveryRepoPG) List(ctx context.Context, limit, offset int) ([]*SupplyDelivery, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM supply_delivery`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+supplyDeliveryCols+` FROM supply_delivery ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*SupplyDelivery
	for rows.Next() {
		s, err := r.scanSupplyDelivery(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

func (r *supplyDeliveryRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SupplyDelivery, int, error) {
	query := `SELECT ` + supplyDeliveryCols + ` FROM supply_delivery WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM supply_delivery WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["supplier"]; ok {
		query += fmt.Sprintf(` AND supplier_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND supplier_id = $%d`, idx)
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
	var items []*SupplyDelivery
	for rows.Next() {
		s, err := r.scanSupplyDelivery(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}
