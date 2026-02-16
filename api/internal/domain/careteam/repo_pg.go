package careteam

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

type careTeamRepoPG struct{ pool *pgxpool.Pool }

func NewCareTeamRepoPG(pool *pgxpool.Pool) CareTeamRepository {
	return &careTeamRepoPG{pool: pool}
}

func (r *careTeamRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const ctCols = `id, fhir_id, status, name, patient_id, encounter_id,
	category_code, category_display, period_start, period_end,
	managing_organization_id, reason_code, reason_display, note,
	created_at, updated_at`

func (r *careTeamRepoPG) scanCT(row pgx.Row) (*CareTeam, error) {
	var ct CareTeam
	err := row.Scan(&ct.ID, &ct.FHIRID, &ct.Status, &ct.Name,
		&ct.PatientID, &ct.EncounterID,
		&ct.CategoryCode, &ct.CategoryDisplay,
		&ct.PeriodStart, &ct.PeriodEnd,
		&ct.ManagingOrganizationID, &ct.ReasonCode, &ct.ReasonDisplay,
		&ct.Note, &ct.CreatedAt, &ct.UpdatedAt)
	return &ct, err
}

func (r *careTeamRepoPG) Create(ctx context.Context, ct *CareTeam) error {
	ct.ID = uuid.New()
	if ct.FHIRID == "" {
		ct.FHIRID = ct.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO care_team (id, fhir_id, status, name, patient_id, encounter_id,
			category_code, category_display, period_start, period_end,
			managing_organization_id, reason_code, reason_display, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		ct.ID, ct.FHIRID, ct.Status, ct.Name, ct.PatientID, ct.EncounterID,
		ct.CategoryCode, ct.CategoryDisplay, ct.PeriodStart, ct.PeriodEnd,
		ct.ManagingOrganizationID, ct.ReasonCode, ct.ReasonDisplay, ct.Note)
	return err
}

func (r *careTeamRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*CareTeam, error) {
	return r.scanCT(r.conn(ctx).QueryRow(ctx, `SELECT `+ctCols+` FROM care_team WHERE id = $1`, id))
}

func (r *careTeamRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*CareTeam, error) {
	return r.scanCT(r.conn(ctx).QueryRow(ctx, `SELECT `+ctCols+` FROM care_team WHERE fhir_id = $1`, fhirID))
}

func (r *careTeamRepoPG) Update(ctx context.Context, ct *CareTeam) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE care_team SET status=$2, name=$3, note=$4, updated_at=NOW()
		WHERE id = $1`,
		ct.ID, ct.Status, ct.Name, ct.Note)
	return err
}

func (r *careTeamRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM care_team WHERE id = $1`, id)
	return err
}

func (r *careTeamRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*CareTeam, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM care_team WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+ctCols+` FROM care_team WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*CareTeam
	for rows.Next() {
		ct, err := r.scanCT(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ct)
	}
	return items, total, nil
}

var careTeamSearchParams = map[string]fhir.SearchParamConfig{
	"patient":  {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":   {Type: fhir.SearchParamToken, Column: "status"},
	"category": {Type: fhir.SearchParamToken, Column: "category_code"},
}

func (r *careTeamRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*CareTeam, int, error) {
	qb := fhir.NewSearchQuery("care_team", ctCols)
	qb.ApplyParams(params, careTeamSearchParams)
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
	var items []*CareTeam
	for rows.Next() {
		ct, err := r.scanCT(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ct)
	}
	return items, total, nil
}

func (r *careTeamRepoPG) AddParticipant(ctx context.Context, careTeamID uuid.UUID, p *CareTeamParticipant) error {
	p.ID = uuid.New()
	p.CareTeamID = careTeamID
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO care_team_participant (id, care_team_id, member_id, member_type,
			role_code, role_display, period_start, period_end, on_behalf_of_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		p.ID, p.CareTeamID, p.MemberID, p.MemberType,
		p.RoleCode, p.RoleDisplay, p.PeriodStart, p.PeriodEnd, p.OnBehalfOfID)
	return err
}

func (r *careTeamRepoPG) RemoveParticipant(ctx context.Context, careTeamID uuid.UUID, participantID uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM care_team_participant WHERE id = $1 AND care_team_id = $2`, participantID, careTeamID)
	return err
}

func (r *careTeamRepoPG) GetParticipants(ctx context.Context, careTeamID uuid.UUID) ([]*CareTeamParticipant, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, care_team_id, member_id, member_type,
			role_code, role_display, period_start, period_end, on_behalf_of_id
		FROM care_team_participant WHERE care_team_id = $1`, careTeamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*CareTeamParticipant
	for rows.Next() {
		var p CareTeamParticipant
		if err := rows.Scan(&p.ID, &p.CareTeamID, &p.MemberID, &p.MemberType,
			&p.RoleCode, &p.RoleDisplay, &p.PeriodStart, &p.PeriodEnd, &p.OnBehalfOfID); err != nil {
			return nil, err
		}
		items = append(items, &p)
	}
	return items, nil
}
