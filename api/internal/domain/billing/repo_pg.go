package billing

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

// =========== Coverage Repository ===========

type coverageRepoPG struct{ pool *pgxpool.Pool }

func NewCoverageRepoPG(pool *pgxpool.Pool) CoverageRepository { return &coverageRepoPG{pool: pool} }

func (r *coverageRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const covCols = `id, fhir_id, status, type_code, patient_id,
	subscriber_id, subscriber_name, subscriber_dob, relationship, dependent_number,
	payor_org_id, payor_name, policy_number, group_number, group_name, plan_name, plan_type,
	member_id, bin_number, pcn_number, rx_group, plan_type_us,
	ab_pmjay_id, ab_pmjay_family_id, state_scheme_id, state_scheme_name,
	esis_number, cghs_beneficiary_id, echs_card_number,
	period_start, period_end, network,
	copay_amount, copay_percentage, deductible_amount, deductible_met,
	max_benefit_amount, out_of_pocket_max, currency, coverage_order,
	note, created_at, updated_at`

func (r *coverageRepoPG) scanCoverage(row pgx.Row) (*Coverage, error) {
	var c Coverage
	err := row.Scan(&c.ID, &c.FHIRID, &c.Status, &c.TypeCode, &c.PatientID,
		&c.SubscriberID, &c.SubscriberName, &c.SubscriberDOB, &c.Relationship, &c.DependentNumber,
		&c.PayorOrgID, &c.PayorName, &c.PolicyNumber, &c.GroupNumber, &c.GroupName, &c.PlanName, &c.PlanType,
		&c.MemberID, &c.BINNumber, &c.PCNNumber, &c.RxGroup, &c.PlanTypeUS,
		&c.ABPMJAYId, &c.ABPMJAYFamilyID, &c.StateSchemeID, &c.StateSchemeName,
		&c.ESISNumber, &c.CGHSBenefID, &c.ECHSCardNumber,
		&c.PeriodStart, &c.PeriodEnd, &c.Network,
		&c.CopayAmount, &c.CopayPercentage, &c.DeductibleAmount, &c.DeductibleMet,
		&c.MaxBenefitAmount, &c.OutOfPocketMax, &c.Currency, &c.CoverageOrder,
		&c.Note, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}

func (r *coverageRepoPG) Create(ctx context.Context, c *Coverage) error {
	c.ID = uuid.New()
	if c.FHIRID == "" {
		c.FHIRID = c.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO coverage (id, fhir_id, status, type_code, patient_id,
			subscriber_id, subscriber_name, subscriber_dob, relationship, dependent_number,
			payor_org_id, payor_name, policy_number, group_number, group_name, plan_name, plan_type,
			member_id, bin_number, pcn_number, rx_group, plan_type_us,
			ab_pmjay_id, ab_pmjay_family_id, state_scheme_id, state_scheme_name,
			esis_number, cghs_beneficiary_id, echs_card_number,
			period_start, period_end, network,
			copay_amount, copay_percentage, deductible_amount, deductible_met,
			max_benefit_amount, out_of_pocket_max, currency, coverage_order, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38,$39,$40,$41)`,
		c.ID, c.FHIRID, c.Status, c.TypeCode, c.PatientID,
		c.SubscriberID, c.SubscriberName, c.SubscriberDOB, c.Relationship, c.DependentNumber,
		c.PayorOrgID, c.PayorName, c.PolicyNumber, c.GroupNumber, c.GroupName, c.PlanName, c.PlanType,
		c.MemberID, c.BINNumber, c.PCNNumber, c.RxGroup, c.PlanTypeUS,
		c.ABPMJAYId, c.ABPMJAYFamilyID, c.StateSchemeID, c.StateSchemeName,
		c.ESISNumber, c.CGHSBenefID, c.ECHSCardNumber,
		c.PeriodStart, c.PeriodEnd, c.Network,
		c.CopayAmount, c.CopayPercentage, c.DeductibleAmount, c.DeductibleMet,
		c.MaxBenefitAmount, c.OutOfPocketMax, c.Currency, c.CoverageOrder, c.Note)
	return err
}

func (r *coverageRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Coverage, error) {
	return r.scanCoverage(r.conn(ctx).QueryRow(ctx, `SELECT `+covCols+` FROM coverage WHERE id = $1`, id))
}

func (r *coverageRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Coverage, error) {
	return r.scanCoverage(r.conn(ctx).QueryRow(ctx, `SELECT `+covCols+` FROM coverage WHERE fhir_id = $1`, fhirID))
}

func (r *coverageRepoPG) Update(ctx context.Context, c *Coverage) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE coverage SET status=$2, type_code=$3, payor_org_id=$4, payor_name=$5,
			policy_number=$6, group_number=$7, plan_name=$8,
			period_start=$9, period_end=$10, network=$11,
			copay_amount=$12, deductible_amount=$13, note=$14, updated_at=NOW()
		WHERE id = $1`,
		c.ID, c.Status, c.TypeCode, c.PayorOrgID, c.PayorName,
		c.PolicyNumber, c.GroupNumber, c.PlanName,
		c.PeriodStart, c.PeriodEnd, c.Network,
		c.CopayAmount, c.DeductibleAmount, c.Note)
	return err
}

func (r *coverageRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM coverage WHERE id = $1`, id)
	return err
}

func (r *coverageRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Coverage, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM coverage WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+covCols+` FROM coverage WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Coverage
	for rows.Next() {
		c, err := r.scanCoverage(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, nil
}

var coverageSearchParams = map[string]fhir.SearchParamConfig{
	"patient": {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
	"type":    {Type: fhir.SearchParamToken, Column: "type_code"},
}

func (r *coverageRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Coverage, int, error) {
	qb := fhir.NewSearchQuery("coverage", covCols)
	qb.ApplyParams(params, coverageSearchParams)
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
	var items []*Coverage
	for rows.Next() {
		c, err := r.scanCoverage(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, nil
}

// =========== Claim Repository ===========

type claimRepoPG struct{ pool *pgxpool.Pool }

func NewClaimRepoPG(pool *pgxpool.Pool) ClaimRepository { return &claimRepoPG{pool: pool} }

func (r *claimRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const claimCols = `id, fhir_id, status, type_code, sub_type_code, use_code,
	patient_id, encounter_id, insurer_org_id, provider_id, provider_org_id,
	coverage_id, priority_code, prescription_id, referral_id, facility_id,
	billable_period_start, billable_period_end, created_date,
	total_amount, currency, place_of_service,
	ab_pmjay_claim_id, ab_pmjay_package_code, rohini_claim_id,
	related_claim_id, related_claim_relationship,
	created_at, updated_at`

func (r *claimRepoPG) scanClaim(row pgx.Row) (*Claim, error) {
	var c Claim
	err := row.Scan(&c.ID, &c.FHIRID, &c.Status, &c.TypeCode, &c.SubTypeCode, &c.UseCode,
		&c.PatientID, &c.EncounterID, &c.InsurerOrgID, &c.ProviderID, &c.ProviderOrgID,
		&c.CoverageID, &c.PriorityCode, &c.PrescriptionID, &c.ReferralID, &c.FacilityID,
		&c.BillablePeriodStart, &c.BillablePeriodEnd, &c.CreatedDate,
		&c.TotalAmount, &c.Currency, &c.PlaceOfService,
		&c.ABPMJAYClaimID, &c.ABPMJAYPackageCode, &c.ROHINIClaimID,
		&c.RelatedClaimID, &c.RelatedClaimRelation,
		&c.CreatedAt, &c.UpdatedAt)
	return &c, err
}

func (r *claimRepoPG) Create(ctx context.Context, c *Claim) error {
	c.ID = uuid.New()
	if c.FHIRID == "" {
		c.FHIRID = c.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO claim (id, fhir_id, status, type_code, sub_type_code, use_code,
			patient_id, encounter_id, insurer_org_id, provider_id, provider_org_id,
			coverage_id, priority_code, prescription_id, referral_id, facility_id,
			billable_period_start, billable_period_end,
			total_amount, currency, place_of_service,
			ab_pmjay_claim_id, ab_pmjay_package_code, rohini_claim_id,
			related_claim_id, related_claim_relationship)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26)`,
		c.ID, c.FHIRID, c.Status, c.TypeCode, c.SubTypeCode, c.UseCode,
		c.PatientID, c.EncounterID, c.InsurerOrgID, c.ProviderID, c.ProviderOrgID,
		c.CoverageID, c.PriorityCode, c.PrescriptionID, c.ReferralID, c.FacilityID,
		c.BillablePeriodStart, c.BillablePeriodEnd,
		c.TotalAmount, c.Currency, c.PlaceOfService,
		c.ABPMJAYClaimID, c.ABPMJAYPackageCode, c.ROHINIClaimID,
		c.RelatedClaimID, c.RelatedClaimRelation)
	return err
}

func (r *claimRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Claim, error) {
	return r.scanClaim(r.conn(ctx).QueryRow(ctx, `SELECT `+claimCols+` FROM claim WHERE id = $1`, id))
}

func (r *claimRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Claim, error) {
	return r.scanClaim(r.conn(ctx).QueryRow(ctx, `SELECT `+claimCols+` FROM claim WHERE fhir_id = $1`, fhirID))
}

func (r *claimRepoPG) Update(ctx context.Context, c *Claim) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE claim SET status=$2, type_code=$3, use_code=$4,
			coverage_id=$5, total_amount=$6, currency=$7, updated_at=NOW()
		WHERE id = $1`,
		c.ID, c.Status, c.TypeCode, c.UseCode,
		c.CoverageID, c.TotalAmount, c.Currency)
	return err
}

func (r *claimRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM claim WHERE id = $1`, id)
	return err
}

func (r *claimRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Claim, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM claim WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+claimCols+` FROM claim WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Claim
	for rows.Next() {
		c, err := r.scanClaim(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, nil
}

var claimSearchParams = map[string]fhir.SearchParamConfig{
	"patient": {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
	"use":     {Type: fhir.SearchParamToken, Column: "use_code"},
}

func (r *claimRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Claim, int, error) {
	qb := fhir.NewSearchQuery("claim", claimCols)
	qb.ApplyParams(params, claimSearchParams)
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
	var items []*Claim
	for rows.Next() {
		c, err := r.scanClaim(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, nil
}

func (r *claimRepoPG) AddDiagnosis(ctx context.Context, d *ClaimDiagnosis) error {
	d.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO claim_diagnosis (id, claim_id, sequence, diagnosis_code_system,
			diagnosis_code, diagnosis_display, type_code, on_admission, package_code)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		d.ID, d.ClaimID, d.Sequence, d.DiagnosisCodeSystem,
		d.DiagnosisCode, d.DiagnosisDisplay, d.TypeCode, d.OnAdmission, d.PackageCode)
	return err
}

func (r *claimRepoPG) GetDiagnoses(ctx context.Context, claimID uuid.UUID) ([]*ClaimDiagnosis, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, claim_id, sequence, diagnosis_code_system,
			diagnosis_code, diagnosis_display, type_code, on_admission, package_code
		FROM claim_diagnosis WHERE claim_id = $1 ORDER BY sequence`, claimID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ClaimDiagnosis
	for rows.Next() {
		var d ClaimDiagnosis
		if err := rows.Scan(&d.ID, &d.ClaimID, &d.Sequence, &d.DiagnosisCodeSystem,
			&d.DiagnosisCode, &d.DiagnosisDisplay, &d.TypeCode, &d.OnAdmission, &d.PackageCode); err != nil {
			return nil, err
		}
		items = append(items, &d)
	}
	return items, nil
}

func (r *claimRepoPG) AddProcedure(ctx context.Context, p *ClaimProcedure) error {
	p.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO claim_procedure (id, claim_id, sequence, type_code, date,
			procedure_code_system, procedure_code, procedure_display, udi)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		p.ID, p.ClaimID, p.Sequence, p.TypeCode, p.Date,
		p.ProcedureCodeSystem, p.ProcedureCode, p.ProcedureDisplay, p.UDI)
	return err
}

func (r *claimRepoPG) GetProcedures(ctx context.Context, claimID uuid.UUID) ([]*ClaimProcedure, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, claim_id, sequence, type_code, date,
			procedure_code_system, procedure_code, procedure_display, udi
		FROM claim_procedure WHERE claim_id = $1 ORDER BY sequence`, claimID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ClaimProcedure
	for rows.Next() {
		var p ClaimProcedure
		if err := rows.Scan(&p.ID, &p.ClaimID, &p.Sequence, &p.TypeCode, &p.Date,
			&p.ProcedureCodeSystem, &p.ProcedureCode, &p.ProcedureDisplay, &p.UDI); err != nil {
			return nil, err
		}
		items = append(items, &p)
	}
	return items, nil
}

func (r *claimRepoPG) AddItem(ctx context.Context, item *ClaimItem) error {
	item.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO claim_item (id, claim_id, sequence,
			product_or_service_system, product_or_service_code, product_or_service_display,
			serviced_date, serviced_period_start, serviced_period_end, location_code,
			quantity_value, quantity_unit, unit_price, factor, net_amount, currency,
			revenue_code, revenue_display, body_site_code, sub_site_code, encounter_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`,
		item.ID, item.ClaimID, item.Sequence,
		item.ProductOrServiceSystem, item.ProductOrServiceCode, item.ProductOrServiceDisplay,
		item.ServicedDate, item.ServicedPeriodStart, item.ServicedPeriodEnd, item.LocationCode,
		item.QuantityValue, item.QuantityUnit, item.UnitPrice, item.Factor, item.NetAmount, item.Currency,
		item.RevenueCode, item.RevenueDisplay, item.BodySiteCode, item.SubSiteCode, item.EncounterID, item.Note)
	return err
}

func (r *claimRepoPG) GetItems(ctx context.Context, claimID uuid.UUID) ([]*ClaimItem, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, claim_id, sequence,
			product_or_service_system, product_or_service_code, product_or_service_display,
			serviced_date, serviced_period_start, serviced_period_end, location_code,
			quantity_value, quantity_unit, unit_price, factor, net_amount, currency,
			revenue_code, revenue_display, body_site_code, sub_site_code, encounter_id, note
		FROM claim_item WHERE claim_id = $1 ORDER BY sequence`, claimID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ClaimItem
	for rows.Next() {
		var ci ClaimItem
		if err := rows.Scan(&ci.ID, &ci.ClaimID, &ci.Sequence,
			&ci.ProductOrServiceSystem, &ci.ProductOrServiceCode, &ci.ProductOrServiceDisplay,
			&ci.ServicedDate, &ci.ServicedPeriodStart, &ci.ServicedPeriodEnd, &ci.LocationCode,
			&ci.QuantityValue, &ci.QuantityUnit, &ci.UnitPrice, &ci.Factor, &ci.NetAmount, &ci.Currency,
			&ci.RevenueCode, &ci.RevenueDisplay, &ci.BodySiteCode, &ci.SubSiteCode, &ci.EncounterID, &ci.Note); err != nil {
			return nil, err
		}
		items = append(items, &ci)
	}
	return items, nil
}

// =========== ClaimResponse Repository ===========

type claimResponseRepoPG struct{ pool *pgxpool.Pool }

func NewClaimResponseRepoPG(pool *pgxpool.Pool) ClaimResponseRepository {
	return &claimResponseRepoPG{pool: pool}
}

func (r *claimResponseRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const crCols = `id, fhir_id, claim_id, status, type_code, use_code,
	outcome, disposition, pre_auth_ref,
	payment_type_code, payment_adjustment, payment_adjustment_reason,
	payment_amount, payment_date, payment_identifier,
	total_amount, process_note, communication_request, created_at`

func (r *claimResponseRepoPG) scanCR(row pgx.Row) (*ClaimResponse, error) {
	var cr ClaimResponse
	err := row.Scan(&cr.ID, &cr.FHIRID, &cr.ClaimID, &cr.Status, &cr.TypeCode, &cr.UseCode,
		&cr.Outcome, &cr.Disposition, &cr.PreAuthRef,
		&cr.PaymentTypeCode, &cr.PaymentAdjustment, &cr.PaymentAdjustmentReason,
		&cr.PaymentAmount, &cr.PaymentDate, &cr.PaymentIdentifier,
		&cr.TotalAmount, &cr.ProcessNote, &cr.CommunicationRequest, &cr.CreatedAt)
	return &cr, err
}

func (r *claimResponseRepoPG) Create(ctx context.Context, cr *ClaimResponse) error {
	cr.ID = uuid.New()
	if cr.FHIRID == "" {
		cr.FHIRID = cr.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO claim_response (id, fhir_id, claim_id, status, type_code, use_code,
			outcome, disposition, pre_auth_ref,
			payment_type_code, payment_adjustment, payment_adjustment_reason,
			payment_amount, payment_date, payment_identifier,
			total_amount, process_note, communication_request)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		cr.ID, cr.FHIRID, cr.ClaimID, cr.Status, cr.TypeCode, cr.UseCode,
		cr.Outcome, cr.Disposition, cr.PreAuthRef,
		cr.PaymentTypeCode, cr.PaymentAdjustment, cr.PaymentAdjustmentReason,
		cr.PaymentAmount, cr.PaymentDate, cr.PaymentIdentifier,
		cr.TotalAmount, cr.ProcessNote, cr.CommunicationRequest)
	return err
}

func (r *claimResponseRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ClaimResponse, error) {
	return r.scanCR(r.conn(ctx).QueryRow(ctx, `SELECT `+crCols+` FROM claim_response WHERE id = $1`, id))
}

func (r *claimResponseRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ClaimResponse, error) {
	return r.scanCR(r.conn(ctx).QueryRow(ctx, `SELECT `+crCols+` FROM claim_response WHERE fhir_id = $1`, fhirID))
}

func (r *claimResponseRepoPG) Update(ctx context.Context, cr *ClaimResponse) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE claim_response SET status=$2, outcome=$3, disposition=$4,
			payment_amount=$5, total_amount=$6, process_note=$7
		WHERE id = $1`,
		cr.ID, cr.Status, cr.Outcome, cr.Disposition,
		cr.PaymentAmount, cr.TotalAmount, cr.ProcessNote)
	return err
}

func (r *claimResponseRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM claim_response WHERE id = $1`, id)
	return err
}

func (r *claimResponseRepoPG) ListByClaim(ctx context.Context, claimID uuid.UUID, limit, offset int) ([]*ClaimResponse, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM claim_response WHERE claim_id = $1`, claimID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+crCols+` FROM claim_response WHERE claim_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, claimID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ClaimResponse
	for rows.Next() {
		cr, err := r.scanCR(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, cr)
	}
	return items, total, nil
}

var claimResponseSearchParams = map[string]fhir.SearchParamConfig{
	"request": {Type: fhir.SearchParamReference, Column: "claim_id"},
	"outcome": {Type: fhir.SearchParamToken, Column: "outcome"},
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
}

func (r *claimResponseRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ClaimResponse, int, error) {
	qb := fhir.NewSearchQuery("claim_response", crCols)
	qb.ApplyParams(params, claimResponseSearchParams)
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
	var items []*ClaimResponse
	for rows.Next() {
		cr, err := r.scanCR(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, cr)
	}
	return items, total, nil
}

// =========== ExplanationOfBenefit Repository ===========

type eobRepoPG struct{ pool *pgxpool.Pool }

func NewExplanationOfBenefitRepoPG(pool *pgxpool.Pool) ExplanationOfBenefitRepository {
	return &eobRepoPG{pool: pool}
}

func (r *eobRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const eobCols = `id, fhir_id, status, type_code, use_code,
	patient_id, claim_id, claim_response_id, coverage_id,
	insurer_org_id, provider_id, outcome, disposition,
	billable_period_start, billable_period_end,
	total_submitted, total_benefit, total_patient_responsibility,
	total_payment, payment_date, currency, created_at`

func (r *eobRepoPG) scanEOB(row pgx.Row) (*ExplanationOfBenefit, error) {
	var eob ExplanationOfBenefit
	err := row.Scan(&eob.ID, &eob.FHIRID, &eob.Status, &eob.TypeCode, &eob.UseCode,
		&eob.PatientID, &eob.ClaimID, &eob.ClaimResponseID, &eob.CoverageID,
		&eob.InsurerOrgID, &eob.ProviderID, &eob.Outcome, &eob.Disposition,
		&eob.BillablePeriodStart, &eob.BillablePeriodEnd,
		&eob.TotalSubmitted, &eob.TotalBenefit, &eob.TotalPatientResponsibility,
		&eob.TotalPayment, &eob.PaymentDate, &eob.Currency, &eob.CreatedAt)
	return &eob, err
}

func (r *eobRepoPG) Create(ctx context.Context, eob *ExplanationOfBenefit) error {
	eob.ID = uuid.New()
	if eob.FHIRID == "" {
		eob.FHIRID = eob.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO explanation_of_benefit (id, fhir_id, status, type_code, use_code,
			patient_id, claim_id, claim_response_id, coverage_id,
			insurer_org_id, provider_id, outcome, disposition,
			billable_period_start, billable_period_end,
			total_submitted, total_benefit, total_patient_responsibility,
			total_payment, payment_date, currency)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`,
		eob.ID, eob.FHIRID, eob.Status, eob.TypeCode, eob.UseCode,
		eob.PatientID, eob.ClaimID, eob.ClaimResponseID, eob.CoverageID,
		eob.InsurerOrgID, eob.ProviderID, eob.Outcome, eob.Disposition,
		eob.BillablePeriodStart, eob.BillablePeriodEnd,
		eob.TotalSubmitted, eob.TotalBenefit, eob.TotalPatientResponsibility,
		eob.TotalPayment, eob.PaymentDate, eob.Currency)
	return err
}

func (r *eobRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ExplanationOfBenefit, error) {
	return r.scanEOB(r.conn(ctx).QueryRow(ctx, `SELECT `+eobCols+` FROM explanation_of_benefit WHERE id = $1`, id))
}

func (r *eobRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ExplanationOfBenefit, error) {
	return r.scanEOB(r.conn(ctx).QueryRow(ctx, `SELECT `+eobCols+` FROM explanation_of_benefit WHERE fhir_id = $1`, fhirID))
}

func (r *eobRepoPG) Update(ctx context.Context, eob *ExplanationOfBenefit) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE explanation_of_benefit SET status=$2, type_code=$3, use_code=$4,
			outcome=$5, disposition=$6,
			total_submitted=$7, total_benefit=$8, total_patient_responsibility=$9,
			total_payment=$10, payment_date=$11, currency=$12
		WHERE id = $1`,
		eob.ID, eob.Status, eob.TypeCode, eob.UseCode,
		eob.Outcome, eob.Disposition,
		eob.TotalSubmitted, eob.TotalBenefit, eob.TotalPatientResponsibility,
		eob.TotalPayment, eob.PaymentDate, eob.Currency)
	return err
}

func (r *eobRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM explanation_of_benefit WHERE id = $1`, id)
	return err
}

func (r *eobRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ExplanationOfBenefit, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM explanation_of_benefit WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+eobCols+` FROM explanation_of_benefit WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ExplanationOfBenefit
	for rows.Next() {
		eob, err := r.scanEOB(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, eob)
	}
	return items, total, nil
}

var eobSearchParams = map[string]fhir.SearchParamConfig{
	"patient": {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
}

func (r *eobRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ExplanationOfBenefit, int, error) {
	qb := fhir.NewSearchQuery("explanation_of_benefit", eobCols)
	qb.ApplyParams(params, eobSearchParams)
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
	var items []*ExplanationOfBenefit
	for rows.Next() {
		eob, err := r.scanEOB(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, eob)
	}
	return items, total, nil
}

// =========== Invoice Repository ===========

type invoiceRepoPG struct{ pool *pgxpool.Pool }

func NewInvoiceRepoPG(pool *pgxpool.Pool) InvoiceRepository { return &invoiceRepoPG{pool: pool} }

func (r *invoiceRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const invCols = `id, fhir_id, status, type_code, patient_id, encounter_id,
	issuer_org_id, date, participant_id,
	total_net, total_gross, total_tax, currency, payment_terms,
	gstin, gst_amount, sac_code, note, created_at`

func (r *invoiceRepoPG) scanInvoice(row pgx.Row) (*Invoice, error) {
	var inv Invoice
	err := row.Scan(&inv.ID, &inv.FHIRID, &inv.Status, &inv.TypeCode, &inv.PatientID, &inv.EncounterID,
		&inv.IssuerOrgID, &inv.Date, &inv.ParticipantID,
		&inv.TotalNet, &inv.TotalGross, &inv.TotalTax, &inv.Currency, &inv.PaymentTerms,
		&inv.GSTIN, &inv.GSTAmount, &inv.SACCode, &inv.Note, &inv.CreatedAt)
	return &inv, err
}

func (r *invoiceRepoPG) Create(ctx context.Context, inv *Invoice) error {
	inv.ID = uuid.New()
	if inv.FHIRID == "" {
		inv.FHIRID = inv.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO invoice (id, fhir_id, status, type_code, patient_id, encounter_id,
			issuer_org_id, participant_id,
			total_net, total_gross, total_tax, currency, payment_terms,
			gstin, gst_amount, sac_code, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		inv.ID, inv.FHIRID, inv.Status, inv.TypeCode, inv.PatientID, inv.EncounterID,
		inv.IssuerOrgID, inv.ParticipantID,
		inv.TotalNet, inv.TotalGross, inv.TotalTax, inv.Currency, inv.PaymentTerms,
		inv.GSTIN, inv.GSTAmount, inv.SACCode, inv.Note)
	return err
}

func (r *invoiceRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Invoice, error) {
	return r.scanInvoice(r.conn(ctx).QueryRow(ctx, `SELECT `+invCols+` FROM invoice WHERE id = $1`, id))
}

func (r *invoiceRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Invoice, error) {
	return r.scanInvoice(r.conn(ctx).QueryRow(ctx, `SELECT `+invCols+` FROM invoice WHERE fhir_id = $1`, fhirID))
}

func (r *invoiceRepoPG) Update(ctx context.Context, inv *Invoice) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE invoice SET status=$2, total_net=$3, total_gross=$4, total_tax=$5,
			payment_terms=$6, note=$7
		WHERE id = $1`,
		inv.ID, inv.Status, inv.TotalNet, inv.TotalGross, inv.TotalTax,
		inv.PaymentTerms, inv.Note)
	return err
}

func (r *invoiceRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM invoice WHERE id = $1`, id)
	return err
}

func (r *invoiceRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Invoice, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM invoice WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+invCols+` FROM invoice WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Invoice
	for rows.Next() {
		inv, err := r.scanInvoice(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, inv)
	}
	return items, total, nil
}

var invoiceSearchParams = map[string]fhir.SearchParamConfig{
	"patient": {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
}

func (r *invoiceRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Invoice, int, error) {
	qb := fhir.NewSearchQuery("invoice", invCols)
	qb.ApplyParams(params, invoiceSearchParams)
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
	var items []*Invoice
	for rows.Next() {
		inv, err := r.scanInvoice(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, inv)
	}
	return items, total, nil
}

func (r *invoiceRepoPG) AddLineItem(ctx context.Context, li *InvoiceLineItem) error {
	li.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO invoice_line_item (id, invoice_id, sequence, description,
			service_code, service_display, quantity, unit_price,
			net_amount, tax_amount, gross_amount, currency)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		li.ID, li.InvoiceID, li.Sequence, li.Description,
		li.ServiceCode, li.ServiceDisplay, li.Quantity, li.UnitPrice,
		li.NetAmount, li.TaxAmount, li.GrossAmount, li.Currency)
	return err
}

func (r *invoiceRepoPG) GetLineItems(ctx context.Context, invoiceID uuid.UUID) ([]*InvoiceLineItem, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, invoice_id, sequence, description,
			service_code, service_display, quantity, unit_price,
			net_amount, tax_amount, gross_amount, currency
		FROM invoice_line_item WHERE invoice_id = $1 ORDER BY sequence`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*InvoiceLineItem
	for rows.Next() {
		var li InvoiceLineItem
		if err := rows.Scan(&li.ID, &li.InvoiceID, &li.Sequence, &li.Description,
			&li.ServiceCode, &li.ServiceDisplay, &li.Quantity, &li.UnitPrice,
			&li.NetAmount, &li.TaxAmount, &li.GrossAmount, &li.Currency); err != nil {
			return nil, err
		}
		items = append(items, &li)
	}
	return items, nil
}
