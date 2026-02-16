package substancespecification

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

type substanceSpecRepoPG struct{ pool *pgxpool.Pool }

func NewSubstanceSpecificationRepoPG(pool *pgxpool.Pool) SubstanceSpecificationRepository {
	return &substanceSpecRepoPG{pool: pool}
}

func (r *substanceSpecRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const ssCols = `id, fhir_id, status, type_code, type_display, domain_code, domain_display,
	description, source_reference, comment, molecular_weight_amount, molecular_weight_unit,
	version_id, created_at, updated_at`

func (r *substanceSpecRepoPG) scanRow(row pgx.Row) (*SubstanceSpecification, error) {
	var s SubstanceSpecification
	err := row.Scan(&s.ID, &s.FHIRID, &s.Status, &s.TypeCode, &s.TypeDisplay, &s.DomainCode, &s.DomainDisplay,
		&s.Description, &s.SourceReference, &s.Comment, &s.MolecularWeightAmount, &s.MolecularWeightUnit,
		&s.VersionID, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *substanceSpecRepoPG) Create(ctx context.Context, s *SubstanceSpecification) error {
	s.ID = uuid.New()
	if s.FHIRID == "" {
		s.FHIRID = s.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO substance_specification (id, fhir_id, status, type_code, type_display, domain_code, domain_display,
			description, source_reference, comment, molecular_weight_amount, molecular_weight_unit)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		s.ID, s.FHIRID, s.Status, s.TypeCode, s.TypeDisplay, s.DomainCode, s.DomainDisplay,
		s.Description, s.SourceReference, s.Comment, s.MolecularWeightAmount, s.MolecularWeightUnit)
	return err
}

func (r *substanceSpecRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*SubstanceSpecification, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+ssCols+` FROM substance_specification WHERE id = $1`, id))
}

func (r *substanceSpecRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*SubstanceSpecification, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+ssCols+` FROM substance_specification WHERE fhir_id = $1`, fhirID))
}

func (r *substanceSpecRepoPG) Update(ctx context.Context, s *SubstanceSpecification) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE substance_specification SET status=$2, type_code=$3, type_display=$4, domain_code=$5, domain_display=$6,
			description=$7, source_reference=$8, comment=$9, molecular_weight_amount=$10, molecular_weight_unit=$11, updated_at=NOW()
		WHERE id = $1`,
		s.ID, s.Status, s.TypeCode, s.TypeDisplay, s.DomainCode, s.DomainDisplay,
		s.Description, s.SourceReference, s.Comment, s.MolecularWeightAmount, s.MolecularWeightUnit)
	return err
}

func (r *substanceSpecRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM substance_specification WHERE id = $1`, id)
	return err
}

func (r *substanceSpecRepoPG) List(ctx context.Context, limit, offset int) ([]*SubstanceSpecification, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM substance_specification`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+ssCols+` FROM substance_specification ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*SubstanceSpecification
	for rows.Next() {
		s, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

var ssSearchParams = map[string]fhir.SearchParamConfig{
	"type":   {Type: fhir.SearchParamToken, Column: "type_code"},
	"domain": {Type: fhir.SearchParamToken, Column: "domain_code"},
}

func (r *substanceSpecRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SubstanceSpecification, int, error) {
	qb := fhir.NewSearchQuery("substance_specification", ssCols)
	qb.ApplyParams(params, ssSearchParams)
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
	var items []*SubstanceSpecification
	for rows.Next() {
		s, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}
