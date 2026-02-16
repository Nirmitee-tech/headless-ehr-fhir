package conformance

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

// =========== NamingSystem Repository ===========

type namingSystemRepoPG struct{ pool *pgxpool.Pool }

func NewNamingSystemRepoPG(pool *pgxpool.Pool) NamingSystemRepository {
	return &namingSystemRepoPG{pool: pool}
}

func (r *namingSystemRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const nsCols = `id, fhir_id, name, status, kind, date, publisher, responsible,
	type_code, type_display, description, usage_note, jurisdiction,
	version_id, created_at, updated_at`

func (r *namingSystemRepoPG) scanNS(row pgx.Row) (*NamingSystem, error) {
	var ns NamingSystem
	err := row.Scan(&ns.ID, &ns.FHIRID, &ns.Name, &ns.Status, &ns.Kind,
		&ns.Date, &ns.Publisher, &ns.Responsible,
		&ns.TypeCode, &ns.TypeDisplay, &ns.Description, &ns.UsageNote, &ns.Jurisdiction,
		&ns.VersionID, &ns.CreatedAt, &ns.UpdatedAt)
	return &ns, err
}

func (r *namingSystemRepoPG) Create(ctx context.Context, ns *NamingSystem) error {
	ns.ID = uuid.New()
	if ns.FHIRID == "" {
		ns.FHIRID = ns.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO naming_system (id, fhir_id, name, status, kind, date, publisher, responsible,
			type_code, type_display, description, usage_note, jurisdiction)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		ns.ID, ns.FHIRID, ns.Name, ns.Status, ns.Kind, ns.Date, ns.Publisher, ns.Responsible,
		ns.TypeCode, ns.TypeDisplay, ns.Description, ns.UsageNote, ns.Jurisdiction)
	return err
}

func (r *namingSystemRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*NamingSystem, error) {
	return r.scanNS(r.conn(ctx).QueryRow(ctx, `SELECT `+nsCols+` FROM naming_system WHERE id = $1`, id))
}

func (r *namingSystemRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*NamingSystem, error) {
	return r.scanNS(r.conn(ctx).QueryRow(ctx, `SELECT `+nsCols+` FROM naming_system WHERE fhir_id = $1`, fhirID))
}

func (r *namingSystemRepoPG) Update(ctx context.Context, ns *NamingSystem) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE naming_system SET name=$2, status=$3, kind=$4, publisher=$5,
			description=$6, usage_note=$7, updated_at=NOW()
		WHERE id = $1`,
		ns.ID, ns.Name, ns.Status, ns.Kind, ns.Publisher,
		ns.Description, ns.UsageNote)
	return err
}

func (r *namingSystemRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM naming_system WHERE id = $1`, id)
	return err
}

func (r *namingSystemRepoPG) List(ctx context.Context, limit, offset int) ([]*NamingSystem, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM naming_system`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+nsCols+` FROM naming_system ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*NamingSystem
	for rows.Next() {
		ns, err := r.scanNS(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ns)
	}
	return items, total, nil
}

var nsSearchParams = map[string]fhir.SearchParamConfig{
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"name":   {Type: fhir.SearchParamString, Column: "name"},
	"kind":   {Type: fhir.SearchParamToken, Column: "kind"},
}

func (r *namingSystemRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*NamingSystem, int, error) {
	qb := fhir.NewSearchQuery("naming_system", nsCols)
	qb.ApplyParams(params, nsSearchParams)
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
	var items []*NamingSystem
	for rows.Next() {
		ns, err := r.scanNS(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ns)
	}
	return items, total, nil
}

func (r *namingSystemRepoPG) AddUniqueID(ctx context.Context, uid *NamingSystemUniqueID) error {
	uid.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO naming_system_unique_id (id, naming_system_id, type, value, preferred,
			comment, period_start, period_end)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		uid.ID, uid.NamingSystemID, uid.Type, uid.Value, uid.Preferred,
		uid.Comment, uid.PeriodStart, uid.PeriodEnd)
	return err
}

func (r *namingSystemRepoPG) GetUniqueIDs(ctx context.Context, namingSystemID uuid.UUID) ([]*NamingSystemUniqueID, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, naming_system_id, type, value, preferred, comment, period_start, period_end
		FROM naming_system_unique_id WHERE naming_system_id = $1 ORDER BY id`, namingSystemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*NamingSystemUniqueID
	for rows.Next() {
		var uid NamingSystemUniqueID
		if err := rows.Scan(&uid.ID, &uid.NamingSystemID, &uid.Type, &uid.Value,
			&uid.Preferred, &uid.Comment, &uid.PeriodStart, &uid.PeriodEnd); err != nil {
			return nil, err
		}
		items = append(items, &uid)
	}
	return items, nil
}

// =========== OperationDefinition Repository ===========

type opDefRepoPG struct{ pool *pgxpool.Pool }

func NewOperationDefinitionRepoPG(pool *pgxpool.Pool) OperationDefinitionRepository {
	return &opDefRepoPG{pool: pool}
}

func (r *opDefRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const odCols = `id, fhir_id, url, name, title, status, kind, description, code,
	system, type, instance, input_profile, output_profile, publisher,
	version_id, created_at, updated_at`

func (r *opDefRepoPG) scanOD(row pgx.Row) (*OperationDefinition, error) {
	var od OperationDefinition
	err := row.Scan(&od.ID, &od.FHIRID, &od.URL, &od.Name, &od.Title,
		&od.Status, &od.Kind, &od.Description, &od.Code,
		&od.System, &od.Type, &od.Instance, &od.InputProfile, &od.OutputProfile, &od.Publisher,
		&od.VersionID, &od.CreatedAt, &od.UpdatedAt)
	return &od, err
}

func (r *opDefRepoPG) Create(ctx context.Context, od *OperationDefinition) error {
	od.ID = uuid.New()
	if od.FHIRID == "" {
		od.FHIRID = od.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO operation_definition (id, fhir_id, url, name, title, status, kind,
			description, code, system, type, instance, input_profile, output_profile, publisher)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		od.ID, od.FHIRID, od.URL, od.Name, od.Title, od.Status, od.Kind,
		od.Description, od.Code, od.System, od.Type, od.Instance,
		od.InputProfile, od.OutputProfile, od.Publisher)
	return err
}

func (r *opDefRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*OperationDefinition, error) {
	return r.scanOD(r.conn(ctx).QueryRow(ctx, `SELECT `+odCols+` FROM operation_definition WHERE id = $1`, id))
}

func (r *opDefRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*OperationDefinition, error) {
	return r.scanOD(r.conn(ctx).QueryRow(ctx, `SELECT `+odCols+` FROM operation_definition WHERE fhir_id = $1`, fhirID))
}

func (r *opDefRepoPG) Update(ctx context.Context, od *OperationDefinition) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE operation_definition SET name=$2, status=$3, kind=$4, description=$5,
			code=$6, system=$7, type=$8, instance=$9, publisher=$10, updated_at=NOW()
		WHERE id = $1`,
		od.ID, od.Name, od.Status, od.Kind, od.Description,
		od.Code, od.System, od.Type, od.Instance, od.Publisher)
	return err
}

func (r *opDefRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM operation_definition WHERE id = $1`, id)
	return err
}

func (r *opDefRepoPG) List(ctx context.Context, limit, offset int) ([]*OperationDefinition, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM operation_definition`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+odCols+` FROM operation_definition ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*OperationDefinition
	for rows.Next() {
		od, err := r.scanOD(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, od)
	}
	return items, total, nil
}

var odSearchParams = map[string]fhir.SearchParamConfig{
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"name":   {Type: fhir.SearchParamString, Column: "name"},
	"code":   {Type: fhir.SearchParamToken, Column: "code"},
}

func (r *opDefRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*OperationDefinition, int, error) {
	qb := fhir.NewSearchQuery("operation_definition", odCols)
	qb.ApplyParams(params, odSearchParams)
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
	var items []*OperationDefinition
	for rows.Next() {
		od, err := r.scanOD(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, od)
	}
	return items, total, nil
}

func (r *opDefRepoPG) AddParameter(ctx context.Context, p *OperationDefinitionParameter) error {
	p.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO operation_definition_parameter (id, operation_definition_id, name, use,
			min_val, max_val, documentation, type, search_type)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		p.ID, p.OperationDefinitionID, p.Name, p.Use,
		p.MinVal, p.MaxVal, p.Documentation, p.Type, p.SearchType)
	return err
}

func (r *opDefRepoPG) GetParameters(ctx context.Context, opDefID uuid.UUID) ([]*OperationDefinitionParameter, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, operation_definition_id, name, use, min_val, max_val,
			documentation, type, search_type
		FROM operation_definition_parameter WHERE operation_definition_id = $1 ORDER BY id`, opDefID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*OperationDefinitionParameter
	for rows.Next() {
		var p OperationDefinitionParameter
		if err := rows.Scan(&p.ID, &p.OperationDefinitionID, &p.Name, &p.Use,
			&p.MinVal, &p.MaxVal, &p.Documentation, &p.Type, &p.SearchType); err != nil {
			return nil, err
		}
		items = append(items, &p)
	}
	return items, nil
}

// =========== MessageDefinition Repository ===========

type msgDefRepoPG struct{ pool *pgxpool.Pool }

func NewMessageDefinitionRepoPG(pool *pgxpool.Pool) MessageDefinitionRepository {
	return &msgDefRepoPG{pool: pool}
}

func (r *msgDefRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const mdCols = `id, fhir_id, url, name, title, status, date, publisher, description, purpose,
	event_coding_code, event_coding_system, event_coding_display,
	category, response_required, version_id, created_at, updated_at`

func (r *msgDefRepoPG) scanMD(row pgx.Row) (*MessageDefinition, error) {
	var md MessageDefinition
	err := row.Scan(&md.ID, &md.FHIRID, &md.URL, &md.Name, &md.Title,
		&md.Status, &md.Date, &md.Publisher, &md.Description, &md.Purpose,
		&md.EventCodingCode, &md.EventCodingSystem, &md.EventCodingDisplay,
		&md.Category, &md.ResponseRequired, &md.VersionID, &md.CreatedAt, &md.UpdatedAt)
	return &md, err
}

func (r *msgDefRepoPG) Create(ctx context.Context, md *MessageDefinition) error {
	md.ID = uuid.New()
	if md.FHIRID == "" {
		md.FHIRID = md.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO message_definition (id, fhir_id, url, name, title, status, date, publisher,
			description, purpose, event_coding_code, event_coding_system, event_coding_display,
			category, response_required)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		md.ID, md.FHIRID, md.URL, md.Name, md.Title, md.Status, md.Date, md.Publisher,
		md.Description, md.Purpose, md.EventCodingCode, md.EventCodingSystem, md.EventCodingDisplay,
		md.Category, md.ResponseRequired)
	return err
}

func (r *msgDefRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MessageDefinition, error) {
	return r.scanMD(r.conn(ctx).QueryRow(ctx, `SELECT `+mdCols+` FROM message_definition WHERE id = $1`, id))
}

func (r *msgDefRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MessageDefinition, error) {
	return r.scanMD(r.conn(ctx).QueryRow(ctx, `SELECT `+mdCols+` FROM message_definition WHERE fhir_id = $1`, fhirID))
}

func (r *msgDefRepoPG) Update(ctx context.Context, md *MessageDefinition) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE message_definition SET name=$2, status=$3, description=$4, purpose=$5,
			event_coding_code=$6, category=$7, response_required=$8, updated_at=NOW()
		WHERE id = $1`,
		md.ID, md.Name, md.Status, md.Description, md.Purpose,
		md.EventCodingCode, md.Category, md.ResponseRequired)
	return err
}

func (r *msgDefRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM message_definition WHERE id = $1`, id)
	return err
}

func (r *msgDefRepoPG) List(ctx context.Context, limit, offset int) ([]*MessageDefinition, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM message_definition`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+mdCols+` FROM message_definition ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MessageDefinition
	for rows.Next() {
		md, err := r.scanMD(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, md)
	}
	return items, total, nil
}

var mdSearchParams = map[string]fhir.SearchParamConfig{
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"event":  {Type: fhir.SearchParamToken, Column: "event_coding_code"},
}

func (r *msgDefRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MessageDefinition, int, error) {
	qb := fhir.NewSearchQuery("message_definition", mdCols)
	qb.ApplyParams(params, mdSearchParams)
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
	var items []*MessageDefinition
	for rows.Next() {
		md, err := r.scanMD(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, md)
	}
	return items, total, nil
}

// =========== MessageHeader Repository ===========

type msgHeaderRepoPG struct{ pool *pgxpool.Pool }

func NewMessageHeaderRepoPG(pool *pgxpool.Pool) MessageHeaderRepository {
	return &msgHeaderRepoPG{pool: pool}
}

func (r *msgHeaderRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const mhCols = `id, fhir_id, event_coding_code, event_coding_system, event_coding_display,
	destination_name, destination_endpoint, sender_org_id,
	source_name, source_endpoint, source_software, source_version,
	reason_code, reason_display, response_identifier, response_code,
	focus_reference, definition_url, version_id, created_at, updated_at`

func (r *msgHeaderRepoPG) scanMH(row pgx.Row) (*MessageHeader, error) {
	var mh MessageHeader
	err := row.Scan(&mh.ID, &mh.FHIRID, &mh.EventCodingCode, &mh.EventCodingSystem, &mh.EventCodingDisplay,
		&mh.DestinationName, &mh.DestinationEndpoint, &mh.SenderOrgID,
		&mh.SourceName, &mh.SourceEndpoint, &mh.SourceSoftware, &mh.SourceVersion,
		&mh.ReasonCode, &mh.ReasonDisplay, &mh.ResponseIdentifier, &mh.ResponseCode,
		&mh.FocusReference, &mh.DefinitionURL, &mh.VersionID, &mh.CreatedAt, &mh.UpdatedAt)
	return &mh, err
}

func (r *msgHeaderRepoPG) Create(ctx context.Context, mh *MessageHeader) error {
	mh.ID = uuid.New()
	if mh.FHIRID == "" {
		mh.FHIRID = mh.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO message_header (id, fhir_id, event_coding_code, event_coding_system, event_coding_display,
			destination_name, destination_endpoint, sender_org_id,
			source_name, source_endpoint, source_software, source_version,
			reason_code, reason_display, response_identifier, response_code,
			focus_reference, definition_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		mh.ID, mh.FHIRID, mh.EventCodingCode, mh.EventCodingSystem, mh.EventCodingDisplay,
		mh.DestinationName, mh.DestinationEndpoint, mh.SenderOrgID,
		mh.SourceName, mh.SourceEndpoint, mh.SourceSoftware, mh.SourceVersion,
		mh.ReasonCode, mh.ReasonDisplay, mh.ResponseIdentifier, mh.ResponseCode,
		mh.FocusReference, mh.DefinitionURL)
	return err
}

func (r *msgHeaderRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MessageHeader, error) {
	return r.scanMH(r.conn(ctx).QueryRow(ctx, `SELECT `+mhCols+` FROM message_header WHERE id = $1`, id))
}

func (r *msgHeaderRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MessageHeader, error) {
	return r.scanMH(r.conn(ctx).QueryRow(ctx, `SELECT `+mhCols+` FROM message_header WHERE fhir_id = $1`, fhirID))
}

func (r *msgHeaderRepoPG) Update(ctx context.Context, mh *MessageHeader) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE message_header SET event_coding_code=$2, event_coding_system=$3,
			destination_name=$4, destination_endpoint=$5, source_endpoint=$6,
			reason_code=$7, response_code=$8, updated_at=NOW()
		WHERE id = $1`,
		mh.ID, mh.EventCodingCode, mh.EventCodingSystem,
		mh.DestinationName, mh.DestinationEndpoint, mh.SourceEndpoint,
		mh.ReasonCode, mh.ResponseCode)
	return err
}

func (r *msgHeaderRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM message_header WHERE id = $1`, id)
	return err
}

func (r *msgHeaderRepoPG) List(ctx context.Context, limit, offset int) ([]*MessageHeader, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM message_header`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+mhCols+` FROM message_header ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MessageHeader
	for rows.Next() {
		mh, err := r.scanMH(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, mh)
	}
	return items, total, nil
}

var mhSearchParams = map[string]fhir.SearchParamConfig{
	"event": {Type: fhir.SearchParamToken, Column: "event_coding_code"},
}

func (r *msgHeaderRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MessageHeader, int, error) {
	qb := fhir.NewSearchQuery("message_header", mhCols)
	qb.ApplyParams(params, mhSearchParams)
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
	var items []*MessageHeader
	for rows.Next() {
		mh, err := r.scanMH(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, mh)
	}
	return items, total, nil
}
