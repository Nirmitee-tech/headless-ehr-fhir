package financial

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

// =========== Account Repository ===========

type accountRepoPG struct{ pool *pgxpool.Pool }

func NewAccountRepoPG(pool *pgxpool.Pool) AccountRepository { return &accountRepoPG{pool: pool} }

func (r *accountRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const acctCols = `id, fhir_id, status, type_code, type_display, name,
	subject_patient_id, service_period_start, service_period_end,
	owner_org_id, description, version_id, created_at, updated_at`

func (r *accountRepoPG) scanAccount(row pgx.Row) (*Account, error) {
	var a Account
	err := row.Scan(&a.ID, &a.FHIRID, &a.Status, &a.TypeCode, &a.TypeDisplay, &a.Name,
		&a.SubjectPatientID, &a.ServicePeriodStart, &a.ServicePeriodEnd,
		&a.OwnerOrgID, &a.Description, &a.VersionID, &a.CreatedAt, &a.UpdatedAt)
	return &a, err
}

func (r *accountRepoPG) Create(ctx context.Context, a *Account) error {
	a.ID = uuid.New()
	if a.FHIRID == "" {
		a.FHIRID = a.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO account (id, fhir_id, status, type_code, type_display, name,
			subject_patient_id, service_period_start, service_period_end,
			owner_org_id, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		a.ID, a.FHIRID, a.Status, a.TypeCode, a.TypeDisplay, a.Name,
		a.SubjectPatientID, a.ServicePeriodStart, a.ServicePeriodEnd,
		a.OwnerOrgID, a.Description)
	return err
}

func (r *accountRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	return r.scanAccount(r.conn(ctx).QueryRow(ctx, `SELECT `+acctCols+` FROM account WHERE id = $1`, id))
}

func (r *accountRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Account, error) {
	return r.scanAccount(r.conn(ctx).QueryRow(ctx, `SELECT `+acctCols+` FROM account WHERE fhir_id = $1`, fhirID))
}

func (r *accountRepoPG) Update(ctx context.Context, a *Account) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE account SET status=$2, type_code=$3, type_display=$4, name=$5,
			subject_patient_id=$6, service_period_start=$7, service_period_end=$8,
			owner_org_id=$9, description=$10, updated_at=NOW()
		WHERE id = $1`,
		a.ID, a.Status, a.TypeCode, a.TypeDisplay, a.Name,
		a.SubjectPatientID, a.ServicePeriodStart, a.ServicePeriodEnd,
		a.OwnerOrgID, a.Description)
	return err
}

func (r *accountRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM account WHERE id = $1`, id)
	return err
}

func (r *accountRepoPG) List(ctx context.Context, limit, offset int) ([]*Account, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM account`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+acctCols+` FROM account ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Account
	for rows.Next() {
		a, err := r.scanAccount(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

var accountSearchParams = map[string]fhir.SearchParamConfig{
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
	"name":    {Type: fhir.SearchParamString, Column: "name"},
	"patient": {Type: fhir.SearchParamReference, Column: "subject_patient_id"},
}

func (r *accountRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Account, int, error) {
	qb := fhir.NewSearchQuery("account", acctCols)
	qb.ApplyParams(params, accountSearchParams)
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
	var items []*Account
	for rows.Next() {
		a, err := r.scanAccount(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

// =========== InsurancePlan Repository ===========

type insurancePlanRepoPG struct{ pool *pgxpool.Pool }

func NewInsurancePlanRepoPG(pool *pgxpool.Pool) InsurancePlanRepository {
	return &insurancePlanRepoPG{pool: pool}
}

func (r *insurancePlanRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const ipCols = `id, fhir_id, status, type_code, type_display, name, alias,
	period_start, period_end, owned_by_org_id, administered_by_org_id,
	coverage_area, network_name, version_id, created_at, updated_at`

func (r *insurancePlanRepoPG) scanInsurancePlan(row pgx.Row) (*InsurancePlan, error) {
	var ip InsurancePlan
	err := row.Scan(&ip.ID, &ip.FHIRID, &ip.Status, &ip.TypeCode, &ip.TypeDisplay, &ip.Name, &ip.Alias,
		&ip.PeriodStart, &ip.PeriodEnd, &ip.OwnedByOrgID, &ip.AdministeredByOrgID,
		&ip.CoverageArea, &ip.NetworkName, &ip.VersionID, &ip.CreatedAt, &ip.UpdatedAt)
	return &ip, err
}

func (r *insurancePlanRepoPG) Create(ctx context.Context, ip *InsurancePlan) error {
	ip.ID = uuid.New()
	if ip.FHIRID == "" {
		ip.FHIRID = ip.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO insurance_plan (id, fhir_id, status, type_code, type_display, name, alias,
			period_start, period_end, owned_by_org_id, administered_by_org_id,
			coverage_area, network_name)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		ip.ID, ip.FHIRID, ip.Status, ip.TypeCode, ip.TypeDisplay, ip.Name, ip.Alias,
		ip.PeriodStart, ip.PeriodEnd, ip.OwnedByOrgID, ip.AdministeredByOrgID,
		ip.CoverageArea, ip.NetworkName)
	return err
}

func (r *insurancePlanRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*InsurancePlan, error) {
	return r.scanInsurancePlan(r.conn(ctx).QueryRow(ctx, `SELECT `+ipCols+` FROM insurance_plan WHERE id = $1`, id))
}

func (r *insurancePlanRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*InsurancePlan, error) {
	return r.scanInsurancePlan(r.conn(ctx).QueryRow(ctx, `SELECT `+ipCols+` FROM insurance_plan WHERE fhir_id = $1`, fhirID))
}

func (r *insurancePlanRepoPG) Update(ctx context.Context, ip *InsurancePlan) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE insurance_plan SET status=$2, type_code=$3, type_display=$4, name=$5, alias=$6,
			period_start=$7, period_end=$8, owned_by_org_id=$9, administered_by_org_id=$10,
			coverage_area=$11, network_name=$12, updated_at=NOW()
		WHERE id = $1`,
		ip.ID, ip.Status, ip.TypeCode, ip.TypeDisplay, ip.Name, ip.Alias,
		ip.PeriodStart, ip.PeriodEnd, ip.OwnedByOrgID, ip.AdministeredByOrgID,
		ip.CoverageArea, ip.NetworkName)
	return err
}

func (r *insurancePlanRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM insurance_plan WHERE id = $1`, id)
	return err
}

func (r *insurancePlanRepoPG) List(ctx context.Context, limit, offset int) ([]*InsurancePlan, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM insurance_plan`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+ipCols+` FROM insurance_plan ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*InsurancePlan
	for rows.Next() {
		ip, err := r.scanInsurancePlan(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ip)
	}
	return items, total, nil
}

var insurancePlanSearchParams = map[string]fhir.SearchParamConfig{
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"name":   {Type: fhir.SearchParamString, Column: "name"},
	"type":   {Type: fhir.SearchParamToken, Column: "type_code"},
}

func (r *insurancePlanRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*InsurancePlan, int, error) {
	qb := fhir.NewSearchQuery("insurance_plan", ipCols)
	qb.ApplyParams(params, insurancePlanSearchParams)
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
	var items []*InsurancePlan
	for rows.Next() {
		ip, err := r.scanInsurancePlan(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ip)
	}
	return items, total, nil
}

// =========== PaymentNotice Repository ===========

type paymentNoticeRepoPG struct{ pool *pgxpool.Pool }

func NewPaymentNoticeRepoPG(pool *pgxpool.Pool) PaymentNoticeRepository {
	return &paymentNoticeRepoPG{pool: pool}
}

func (r *paymentNoticeRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const pnCols = `id, fhir_id, status, request_reference, response_reference, created,
	provider_id, payment_reference, payment_date, payee_org_id, recipient_org_id,
	amount_value, amount_currency, payment_status_code, version_id, created_at, updated_at`

func (r *paymentNoticeRepoPG) scanPaymentNotice(row pgx.Row) (*PaymentNotice, error) {
	var pn PaymentNotice
	err := row.Scan(&pn.ID, &pn.FHIRID, &pn.Status, &pn.RequestReference, &pn.ResponseReference, &pn.Created,
		&pn.ProviderID, &pn.PaymentReference, &pn.PaymentDate, &pn.PayeeOrgID, &pn.RecipientOrgID,
		&pn.AmountValue, &pn.AmountCurrency, &pn.PaymentStatusCode, &pn.VersionID, &pn.CreatedAt, &pn.UpdatedAt)
	return &pn, err
}

func (r *paymentNoticeRepoPG) Create(ctx context.Context, pn *PaymentNotice) error {
	pn.ID = uuid.New()
	if pn.FHIRID == "" {
		pn.FHIRID = pn.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO payment_notice (id, fhir_id, status, request_reference, response_reference, created,
			provider_id, payment_reference, payment_date, payee_org_id, recipient_org_id,
			amount_value, amount_currency, payment_status_code)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		pn.ID, pn.FHIRID, pn.Status, pn.RequestReference, pn.ResponseReference, pn.Created,
		pn.ProviderID, pn.PaymentReference, pn.PaymentDate, pn.PayeeOrgID, pn.RecipientOrgID,
		pn.AmountValue, pn.AmountCurrency, pn.PaymentStatusCode)
	return err
}

func (r *paymentNoticeRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*PaymentNotice, error) {
	return r.scanPaymentNotice(r.conn(ctx).QueryRow(ctx, `SELECT `+pnCols+` FROM payment_notice WHERE id = $1`, id))
}

func (r *paymentNoticeRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*PaymentNotice, error) {
	return r.scanPaymentNotice(r.conn(ctx).QueryRow(ctx, `SELECT `+pnCols+` FROM payment_notice WHERE fhir_id = $1`, fhirID))
}

func (r *paymentNoticeRepoPG) Update(ctx context.Context, pn *PaymentNotice) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE payment_notice SET status=$2, request_reference=$3, response_reference=$4,
			provider_id=$5, payment_reference=$6, payment_date=$7,
			payee_org_id=$8, recipient_org_id=$9,
			amount_value=$10, amount_currency=$11, payment_status_code=$12, updated_at=NOW()
		WHERE id = $1`,
		pn.ID, pn.Status, pn.RequestReference, pn.ResponseReference,
		pn.ProviderID, pn.PaymentReference, pn.PaymentDate,
		pn.PayeeOrgID, pn.RecipientOrgID,
		pn.AmountValue, pn.AmountCurrency, pn.PaymentStatusCode)
	return err
}

func (r *paymentNoticeRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM payment_notice WHERE id = $1`, id)
	return err
}

func (r *paymentNoticeRepoPG) List(ctx context.Context, limit, offset int) ([]*PaymentNotice, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM payment_notice`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+pnCols+` FROM payment_notice ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*PaymentNotice
	for rows.Next() {
		pn, err := r.scanPaymentNotice(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, pn)
	}
	return items, total, nil
}

var paymentNoticeSearchParams = map[string]fhir.SearchParamConfig{
	"status":   {Type: fhir.SearchParamToken, Column: "status"},
	"provider": {Type: fhir.SearchParamReference, Column: "provider_id"},
}

func (r *paymentNoticeRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*PaymentNotice, int, error) {
	qb := fhir.NewSearchQuery("payment_notice", pnCols)
	qb.ApplyParams(params, paymentNoticeSearchParams)
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
	var items []*PaymentNotice
	for rows.Next() {
		pn, err := r.scanPaymentNotice(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, pn)
	}
	return items, total, nil
}

// =========== PaymentReconciliation Repository ===========

type paymentReconciliationRepoPG struct{ pool *pgxpool.Pool }

func NewPaymentReconciliationRepoPG(pool *pgxpool.Pool) PaymentReconciliationRepository {
	return &paymentReconciliationRepoPG{pool: pool}
}

func (r *paymentReconciliationRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const prCols = `id, fhir_id, status, period_start, period_end, created,
	payment_issuer_org_id, request_reference, requestor_id,
	outcome, disposition, payment_date, payment_amount, payment_currency,
	payment_identifier, form_code, process_note, version_id, created_at, updated_at`

func (r *paymentReconciliationRepoPG) scanPaymentReconciliation(row pgx.Row) (*PaymentReconciliation, error) {
	var pr PaymentReconciliation
	err := row.Scan(&pr.ID, &pr.FHIRID, &pr.Status, &pr.PeriodStart, &pr.PeriodEnd, &pr.Created,
		&pr.PaymentIssuerOrgID, &pr.RequestReference, &pr.RequestorID,
		&pr.Outcome, &pr.Disposition, &pr.PaymentDate, &pr.PaymentAmount, &pr.PaymentCurrency,
		&pr.PaymentIdentifier, &pr.FormCode, &pr.ProcessNote, &pr.VersionID, &pr.CreatedAt, &pr.UpdatedAt)
	return &pr, err
}

func (r *paymentReconciliationRepoPG) Create(ctx context.Context, pr *PaymentReconciliation) error {
	pr.ID = uuid.New()
	if pr.FHIRID == "" {
		pr.FHIRID = pr.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO payment_reconciliation (id, fhir_id, status, period_start, period_end, created,
			payment_issuer_org_id, request_reference, requestor_id,
			outcome, disposition, payment_date, payment_amount, payment_currency,
			payment_identifier, form_code, process_note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		pr.ID, pr.FHIRID, pr.Status, pr.PeriodStart, pr.PeriodEnd, pr.Created,
		pr.PaymentIssuerOrgID, pr.RequestReference, pr.RequestorID,
		pr.Outcome, pr.Disposition, pr.PaymentDate, pr.PaymentAmount, pr.PaymentCurrency,
		pr.PaymentIdentifier, pr.FormCode, pr.ProcessNote)
	return err
}

func (r *paymentReconciliationRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*PaymentReconciliation, error) {
	return r.scanPaymentReconciliation(r.conn(ctx).QueryRow(ctx, `SELECT `+prCols+` FROM payment_reconciliation WHERE id = $1`, id))
}

func (r *paymentReconciliationRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*PaymentReconciliation, error) {
	return r.scanPaymentReconciliation(r.conn(ctx).QueryRow(ctx, `SELECT `+prCols+` FROM payment_reconciliation WHERE fhir_id = $1`, fhirID))
}

func (r *paymentReconciliationRepoPG) Update(ctx context.Context, pr *PaymentReconciliation) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE payment_reconciliation SET status=$2, period_start=$3, period_end=$4,
			payment_issuer_org_id=$5, request_reference=$6, requestor_id=$7,
			outcome=$8, disposition=$9, payment_date=$10, payment_amount=$11, payment_currency=$12,
			payment_identifier=$13, form_code=$14, process_note=$15, updated_at=NOW()
		WHERE id = $1`,
		pr.ID, pr.Status, pr.PeriodStart, pr.PeriodEnd,
		pr.PaymentIssuerOrgID, pr.RequestReference, pr.RequestorID,
		pr.Outcome, pr.Disposition, pr.PaymentDate, pr.PaymentAmount, pr.PaymentCurrency,
		pr.PaymentIdentifier, pr.FormCode, pr.ProcessNote)
	return err
}

func (r *paymentReconciliationRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM payment_reconciliation WHERE id = $1`, id)
	return err
}

func (r *paymentReconciliationRepoPG) List(ctx context.Context, limit, offset int) ([]*PaymentReconciliation, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM payment_reconciliation`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+prCols+` FROM payment_reconciliation ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*PaymentReconciliation
	for rows.Next() {
		pr, err := r.scanPaymentReconciliation(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, pr)
	}
	return items, total, nil
}

var paymentReconciliationSearchParams = map[string]fhir.SearchParamConfig{
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
	"outcome": {Type: fhir.SearchParamToken, Column: "outcome"},
}

func (r *paymentReconciliationRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*PaymentReconciliation, int, error) {
	qb := fhir.NewSearchQuery("payment_reconciliation", prCols)
	qb.ApplyParams(params, paymentReconciliationSearchParams)
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
	var items []*PaymentReconciliation
	for rows.Next() {
		pr, err := r.scanPaymentReconciliation(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, pr)
	}
	return items, total, nil
}

// =========== ChargeItem Repository ===========

type chargeItemRepoPG struct{ pool *pgxpool.Pool }

func NewChargeItemRepoPG(pool *pgxpool.Pool) ChargeItemRepository {
	return &chargeItemRepoPG{pool: pool}
}

func (r *chargeItemRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const ciCols = `id, fhir_id, status, code_code, code_display, code_system,
	subject_patient_id, context_encounter_id, occurrence_date,
	performer_id, performing_org_id, quantity_value, factor_override,
	price_override_value, price_override_currency, override_reason,
	enterer_id, entered_date, account_id, note, version_id, created_at, updated_at`

func (r *chargeItemRepoPG) scanChargeItem(row pgx.Row) (*ChargeItem, error) {
	var ci ChargeItem
	err := row.Scan(&ci.ID, &ci.FHIRID, &ci.Status, &ci.CodeCode, &ci.CodeDisplay, &ci.CodeSystem,
		&ci.SubjectPatientID, &ci.ContextEncounterID, &ci.OccurrenceDate,
		&ci.PerformerID, &ci.PerformingOrgID, &ci.QuantityValue, &ci.FactorOverride,
		&ci.PriceOverrideValue, &ci.PriceOverrideCurrency, &ci.OverrideReason,
		&ci.EntererID, &ci.EnteredDate, &ci.AccountID, &ci.Note, &ci.VersionID, &ci.CreatedAt, &ci.UpdatedAt)
	return &ci, err
}

func (r *chargeItemRepoPG) Create(ctx context.Context, ci *ChargeItem) error {
	ci.ID = uuid.New()
	if ci.FHIRID == "" {
		ci.FHIRID = ci.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO charge_item (id, fhir_id, status, code_code, code_display, code_system,
			subject_patient_id, context_encounter_id, occurrence_date,
			performer_id, performing_org_id, quantity_value, factor_override,
			price_override_value, price_override_currency, override_reason,
			enterer_id, entered_date, account_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`,
		ci.ID, ci.FHIRID, ci.Status, ci.CodeCode, ci.CodeDisplay, ci.CodeSystem,
		ci.SubjectPatientID, ci.ContextEncounterID, ci.OccurrenceDate,
		ci.PerformerID, ci.PerformingOrgID, ci.QuantityValue, ci.FactorOverride,
		ci.PriceOverrideValue, ci.PriceOverrideCurrency, ci.OverrideReason,
		ci.EntererID, ci.EnteredDate, ci.AccountID, ci.Note)
	return err
}

func (r *chargeItemRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ChargeItem, error) {
	return r.scanChargeItem(r.conn(ctx).QueryRow(ctx, `SELECT `+ciCols+` FROM charge_item WHERE id = $1`, id))
}

func (r *chargeItemRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ChargeItem, error) {
	return r.scanChargeItem(r.conn(ctx).QueryRow(ctx, `SELECT `+ciCols+` FROM charge_item WHERE fhir_id = $1`, fhirID))
}

func (r *chargeItemRepoPG) Update(ctx context.Context, ci *ChargeItem) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE charge_item SET status=$2, code_code=$3, code_display=$4, code_system=$5,
			context_encounter_id=$6, occurrence_date=$7,
			performer_id=$8, performing_org_id=$9, quantity_value=$10, factor_override=$11,
			price_override_value=$12, price_override_currency=$13, override_reason=$14,
			enterer_id=$15, entered_date=$16, account_id=$17, note=$18, updated_at=NOW()
		WHERE id = $1`,
		ci.ID, ci.Status, ci.CodeCode, ci.CodeDisplay, ci.CodeSystem,
		ci.ContextEncounterID, ci.OccurrenceDate,
		ci.PerformerID, ci.PerformingOrgID, ci.QuantityValue, ci.FactorOverride,
		ci.PriceOverrideValue, ci.PriceOverrideCurrency, ci.OverrideReason,
		ci.EntererID, ci.EnteredDate, ci.AccountID, ci.Note)
	return err
}

func (r *chargeItemRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM charge_item WHERE id = $1`, id)
	return err
}

func (r *chargeItemRepoPG) List(ctx context.Context, limit, offset int) ([]*ChargeItem, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM charge_item`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+ciCols+` FROM charge_item ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ChargeItem
	for rows.Next() {
		ci, err := r.scanChargeItem(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ci)
	}
	return items, total, nil
}

var chargeItemSearchParams = map[string]fhir.SearchParamConfig{
	"patient": {Type: fhir.SearchParamReference, Column: "subject_patient_id"},
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
	"code":    {Type: fhir.SearchParamToken, Column: "code_code"},
}

func (r *chargeItemRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ChargeItem, int, error) {
	qb := fhir.NewSearchQuery("charge_item", ciCols)
	qb.ApplyParams(params, chargeItemSearchParams)
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
	var items []*ChargeItem
	for rows.Next() {
		ci, err := r.scanChargeItem(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ci)
	}
	return items, total, nil
}

// =========== ChargeItemDefinition Repository ===========

type chargeItemDefinitionRepoPG struct{ pool *pgxpool.Pool }

func NewChargeItemDefinitionRepoPG(pool *pgxpool.Pool) ChargeItemDefinitionRepository {
	return &chargeItemDefinitionRepoPG{pool: pool}
}

func (r *chargeItemDefinitionRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const cdCols = `id, fhir_id, url, status, title, description,
	code_code, code_display, code_system,
	effective_start, effective_end, publisher,
	approval_date, last_review_date, version_id, created_at, updated_at`

func (r *chargeItemDefinitionRepoPG) scanChargeItemDefinition(row pgx.Row) (*ChargeItemDefinition, error) {
	var cd ChargeItemDefinition
	err := row.Scan(&cd.ID, &cd.FHIRID, &cd.URL, &cd.Status, &cd.Title, &cd.Description,
		&cd.CodeCode, &cd.CodeDisplay, &cd.CodeSystem,
		&cd.EffectiveStart, &cd.EffectiveEnd, &cd.Publisher,
		&cd.ApprovalDate, &cd.LastReviewDate, &cd.VersionID, &cd.CreatedAt, &cd.UpdatedAt)
	return &cd, err
}

func (r *chargeItemDefinitionRepoPG) Create(ctx context.Context, cd *ChargeItemDefinition) error {
	cd.ID = uuid.New()
	if cd.FHIRID == "" {
		cd.FHIRID = cd.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO charge_item_definition (id, fhir_id, url, status, title, description,
			code_code, code_display, code_system,
			effective_start, effective_end, publisher,
			approval_date, last_review_date)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		cd.ID, cd.FHIRID, cd.URL, cd.Status, cd.Title, cd.Description,
		cd.CodeCode, cd.CodeDisplay, cd.CodeSystem,
		cd.EffectiveStart, cd.EffectiveEnd, cd.Publisher,
		cd.ApprovalDate, cd.LastReviewDate)
	return err
}

func (r *chargeItemDefinitionRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ChargeItemDefinition, error) {
	return r.scanChargeItemDefinition(r.conn(ctx).QueryRow(ctx, `SELECT `+cdCols+` FROM charge_item_definition WHERE id = $1`, id))
}

func (r *chargeItemDefinitionRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ChargeItemDefinition, error) {
	return r.scanChargeItemDefinition(r.conn(ctx).QueryRow(ctx, `SELECT `+cdCols+` FROM charge_item_definition WHERE fhir_id = $1`, fhirID))
}

func (r *chargeItemDefinitionRepoPG) Update(ctx context.Context, cd *ChargeItemDefinition) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE charge_item_definition SET url=$2, status=$3, title=$4, description=$5,
			code_code=$6, code_display=$7, code_system=$8,
			effective_start=$9, effective_end=$10, publisher=$11,
			approval_date=$12, last_review_date=$13, updated_at=NOW()
		WHERE id = $1`,
		cd.ID, cd.URL, cd.Status, cd.Title, cd.Description,
		cd.CodeCode, cd.CodeDisplay, cd.CodeSystem,
		cd.EffectiveStart, cd.EffectiveEnd, cd.Publisher,
		cd.ApprovalDate, cd.LastReviewDate)
	return err
}

func (r *chargeItemDefinitionRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM charge_item_definition WHERE id = $1`, id)
	return err
}

func (r *chargeItemDefinitionRepoPG) List(ctx context.Context, limit, offset int) ([]*ChargeItemDefinition, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM charge_item_definition`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+cdCols+` FROM charge_item_definition ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ChargeItemDefinition
	for rows.Next() {
		cd, err := r.scanChargeItemDefinition(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, cd)
	}
	return items, total, nil
}

var chargeItemDefinitionSearchParams = map[string]fhir.SearchParamConfig{
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"title":  {Type: fhir.SearchParamString, Column: "title"},
	"code":   {Type: fhir.SearchParamToken, Column: "code_code"},
}

func (r *chargeItemDefinitionRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ChargeItemDefinition, int, error) {
	qb := fhir.NewSearchQuery("charge_item_definition", cdCols)
	qb.ApplyParams(params, chargeItemDefinitionSearchParams)
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
	var items []*ChargeItemDefinition
	for rows.Next() {
		cd, err := r.scanChargeItemDefinition(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, cd)
	}
	return items, total, nil
}

// =========== Contract Repository ===========

type contractRepoPG struct{ pool *pgxpool.Pool }

func NewContractRepoPG(pool *pgxpool.Pool) ContractRepository { return &contractRepoPG{pool: pool} }

func (r *contractRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const ctCols = `id, fhir_id, status, type_code, type_display, sub_type_code,
	title, issued, applies_start, applies_end,
	subject_patient_id, authority_org_id, scope_code, scope_display,
	version_id, created_at, updated_at`

func (r *contractRepoPG) scanContract(row pgx.Row) (*Contract, error) {
	var ct Contract
	err := row.Scan(&ct.ID, &ct.FHIRID, &ct.Status, &ct.TypeCode, &ct.TypeDisplay, &ct.SubTypeCode,
		&ct.Title, &ct.Issued, &ct.AppliesStart, &ct.AppliesEnd,
		&ct.SubjectPatientID, &ct.AuthorityOrgID, &ct.ScopeCode, &ct.ScopeDisplay,
		&ct.VersionID, &ct.CreatedAt, &ct.UpdatedAt)
	return &ct, err
}

func (r *contractRepoPG) Create(ctx context.Context, ct *Contract) error {
	ct.ID = uuid.New()
	if ct.FHIRID == "" {
		ct.FHIRID = ct.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO contract (id, fhir_id, status, type_code, type_display, sub_type_code,
			title, issued, applies_start, applies_end,
			subject_patient_id, authority_org_id, scope_code, scope_display)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		ct.ID, ct.FHIRID, ct.Status, ct.TypeCode, ct.TypeDisplay, ct.SubTypeCode,
		ct.Title, ct.Issued, ct.AppliesStart, ct.AppliesEnd,
		ct.SubjectPatientID, ct.AuthorityOrgID, ct.ScopeCode, ct.ScopeDisplay)
	return err
}

func (r *contractRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Contract, error) {
	return r.scanContract(r.conn(ctx).QueryRow(ctx, `SELECT `+ctCols+` FROM contract WHERE id = $1`, id))
}

func (r *contractRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Contract, error) {
	return r.scanContract(r.conn(ctx).QueryRow(ctx, `SELECT `+ctCols+` FROM contract WHERE fhir_id = $1`, fhirID))
}

func (r *contractRepoPG) Update(ctx context.Context, ct *Contract) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE contract SET status=$2, type_code=$3, type_display=$4, sub_type_code=$5,
			title=$6, issued=$7, applies_start=$8, applies_end=$9,
			subject_patient_id=$10, authority_org_id=$11, scope_code=$12, scope_display=$13, updated_at=NOW()
		WHERE id = $1`,
		ct.ID, ct.Status, ct.TypeCode, ct.TypeDisplay, ct.SubTypeCode,
		ct.Title, ct.Issued, ct.AppliesStart, ct.AppliesEnd,
		ct.SubjectPatientID, ct.AuthorityOrgID, ct.ScopeCode, ct.ScopeDisplay)
	return err
}

func (r *contractRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM contract WHERE id = $1`, id)
	return err
}

func (r *contractRepoPG) List(ctx context.Context, limit, offset int) ([]*Contract, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM contract`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+ctCols+` FROM contract ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Contract
	for rows.Next() {
		ct, err := r.scanContract(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ct)
	}
	return items, total, nil
}

var contractSearchParams = map[string]fhir.SearchParamConfig{
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
	"patient": {Type: fhir.SearchParamReference, Column: "subject_patient_id"},
}

func (r *contractRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Contract, int, error) {
	qb := fhir.NewSearchQuery("contract", ctCols)
	qb.ApplyParams(params, contractSearchParams)
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
	var items []*Contract
	for rows.Next() {
		ct, err := r.scanContract(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ct)
	}
	return items, total, nil
}

// =========== EnrollmentRequest Repository ===========

type enrollmentRequestRepoPG struct{ pool *pgxpool.Pool }

func NewEnrollmentRequestRepoPG(pool *pgxpool.Pool) EnrollmentRequestRepository {
	return &enrollmentRequestRepoPG{pool: pool}
}

func (r *enrollmentRequestRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const erCols = `id, fhir_id, status, created, insurer_org_id, provider_id,
	candidate_patient_id, coverage_id, version_id, created_at, updated_at`

func (r *enrollmentRequestRepoPG) scanEnrollmentRequest(row pgx.Row) (*EnrollmentRequest, error) {
	var er EnrollmentRequest
	err := row.Scan(&er.ID, &er.FHIRID, &er.Status, &er.Created, &er.InsurerOrgID, &er.ProviderID,
		&er.CandidatePatientID, &er.CoverageID, &er.VersionID, &er.CreatedAt, &er.UpdatedAt)
	return &er, err
}

func (r *enrollmentRequestRepoPG) Create(ctx context.Context, er *EnrollmentRequest) error {
	er.ID = uuid.New()
	if er.FHIRID == "" {
		er.FHIRID = er.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO enrollment_request (id, fhir_id, status, created, insurer_org_id, provider_id,
			candidate_patient_id, coverage_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		er.ID, er.FHIRID, er.Status, er.Created, er.InsurerOrgID, er.ProviderID,
		er.CandidatePatientID, er.CoverageID)
	return err
}

func (r *enrollmentRequestRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*EnrollmentRequest, error) {
	return r.scanEnrollmentRequest(r.conn(ctx).QueryRow(ctx, `SELECT `+erCols+` FROM enrollment_request WHERE id = $1`, id))
}

func (r *enrollmentRequestRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*EnrollmentRequest, error) {
	return r.scanEnrollmentRequest(r.conn(ctx).QueryRow(ctx, `SELECT `+erCols+` FROM enrollment_request WHERE fhir_id = $1`, fhirID))
}

func (r *enrollmentRequestRepoPG) Update(ctx context.Context, er *EnrollmentRequest) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE enrollment_request SET status=$2, insurer_org_id=$3, provider_id=$4,
			candidate_patient_id=$5, coverage_id=$6, updated_at=NOW()
		WHERE id = $1`,
		er.ID, er.Status, er.InsurerOrgID, er.ProviderID,
		er.CandidatePatientID, er.CoverageID)
	return err
}

func (r *enrollmentRequestRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM enrollment_request WHERE id = $1`, id)
	return err
}

func (r *enrollmentRequestRepoPG) List(ctx context.Context, limit, offset int) ([]*EnrollmentRequest, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM enrollment_request`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+erCols+` FROM enrollment_request ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*EnrollmentRequest
	for rows.Next() {
		er, err := r.scanEnrollmentRequest(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, er)
	}
	return items, total, nil
}

var enrollmentRequestSearchParams = map[string]fhir.SearchParamConfig{
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
	"patient": {Type: fhir.SearchParamReference, Column: "candidate_patient_id"},
}

func (r *enrollmentRequestRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*EnrollmentRequest, int, error) {
	qb := fhir.NewSearchQuery("enrollment_request", erCols)
	qb.ApplyParams(params, enrollmentRequestSearchParams)
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
	var items []*EnrollmentRequest
	for rows.Next() {
		er, err := r.scanEnrollmentRequest(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, er)
	}
	return items, total, nil
}

// =========== EnrollmentResponse Repository ===========

type enrollmentResponseRepoPG struct{ pool *pgxpool.Pool }

func NewEnrollmentResponseRepoPG(pool *pgxpool.Pool) EnrollmentResponseRepository {
	return &enrollmentResponseRepoPG{pool: pool}
}

func (r *enrollmentResponseRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const erspCols = `id, fhir_id, status, request_id, outcome, disposition,
	created, organization_id, version_id, created_at, updated_at`

func (r *enrollmentResponseRepoPG) scanEnrollmentResponse(row pgx.Row) (*EnrollmentResponse, error) {
	var er EnrollmentResponse
	err := row.Scan(&er.ID, &er.FHIRID, &er.Status, &er.RequestID, &er.Outcome, &er.Disposition,
		&er.Created, &er.OrganizationID, &er.VersionID, &er.CreatedAt, &er.UpdatedAt)
	return &er, err
}

func (r *enrollmentResponseRepoPG) Create(ctx context.Context, er *EnrollmentResponse) error {
	er.ID = uuid.New()
	if er.FHIRID == "" {
		er.FHIRID = er.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO enrollment_response (id, fhir_id, status, request_id, outcome, disposition,
			created, organization_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		er.ID, er.FHIRID, er.Status, er.RequestID, er.Outcome, er.Disposition,
		er.Created, er.OrganizationID)
	return err
}

func (r *enrollmentResponseRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*EnrollmentResponse, error) {
	return r.scanEnrollmentResponse(r.conn(ctx).QueryRow(ctx, `SELECT `+erspCols+` FROM enrollment_response WHERE id = $1`, id))
}

func (r *enrollmentResponseRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*EnrollmentResponse, error) {
	return r.scanEnrollmentResponse(r.conn(ctx).QueryRow(ctx, `SELECT `+erspCols+` FROM enrollment_response WHERE fhir_id = $1`, fhirID))
}

func (r *enrollmentResponseRepoPG) Update(ctx context.Context, er *EnrollmentResponse) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE enrollment_response SET status=$2, request_id=$3, outcome=$4, disposition=$5,
			organization_id=$6, updated_at=NOW()
		WHERE id = $1`,
		er.ID, er.Status, er.RequestID, er.Outcome, er.Disposition,
		er.OrganizationID)
	return err
}

func (r *enrollmentResponseRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM enrollment_response WHERE id = $1`, id)
	return err
}

func (r *enrollmentResponseRepoPG) List(ctx context.Context, limit, offset int) ([]*EnrollmentResponse, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM enrollment_response`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+erspCols+` FROM enrollment_response ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*EnrollmentResponse
	for rows.Next() {
		er, err := r.scanEnrollmentResponse(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, er)
	}
	return items, total, nil
}

var enrollmentResponseSearchParams = map[string]fhir.SearchParamConfig{
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
	"request": {Type: fhir.SearchParamReference, Column: "request_id"},
}

func (r *enrollmentResponseRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*EnrollmentResponse, int, error) {
	qb := fhir.NewSearchQuery("enrollment_response", erspCols)
	qb.ApplyParams(params, enrollmentResponseSearchParams)
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
	var items []*EnrollmentResponse
	for rows.Next() {
		er, err := r.scanEnrollmentResponse(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, er)
	}
	return items, total, nil
}
