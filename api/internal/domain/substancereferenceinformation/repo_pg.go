package substancereferenceinformation

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

type sriRepoPG struct{ pool *pgxpool.Pool }

func NewSubstanceReferenceInformationRepoPG(pool *pgxpool.Pool) SubstanceReferenceInformationRepository {
	return &sriRepoPG{pool: pool}
}

func (r *sriRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil { return tx }
	if c := db.ConnFromContext(ctx); c != nil { return c }
	return r.pool
}

const sriCols = `id, fhir_id, comment, gene_element_type_code, gene_element_type_display,
	gene_element_source_reference, classification_code, classification_display,
	classification_domain_code, classification_domain_display,
	target_type_code, target_type_display,
	version_id, created_at, updated_at`

func (r *sriRepoPG) scanRow(row pgx.Row) (*SubstanceReferenceInformation, error) {
	var m SubstanceReferenceInformation
	err := row.Scan(&m.ID, &m.FHIRID, &m.Comment, &m.GeneElementTypeCode, &m.GeneElementTypeDisplay,
		&m.GeneElementSourceReference, &m.ClassificationCode, &m.ClassificationDisplay,
		&m.ClassificationDomainCode, &m.ClassificationDomainDisplay,
		&m.TargetTypeCode, &m.TargetTypeDisplay,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *sriRepoPG) Create(ctx context.Context, m *SubstanceReferenceInformation) error {
	m.ID = uuid.New()
	if m.FHIRID == "" { m.FHIRID = m.ID.String() }
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO substance_reference_information (id, fhir_id, comment, gene_element_type_code, gene_element_type_display,
			gene_element_source_reference, classification_code, classification_display,
			classification_domain_code, classification_domain_display,
			target_type_code, target_type_display)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		m.ID, m.FHIRID, m.Comment, m.GeneElementTypeCode, m.GeneElementTypeDisplay,
		m.GeneElementSourceReference, m.ClassificationCode, m.ClassificationDisplay,
		m.ClassificationDomainCode, m.ClassificationDomainDisplay,
		m.TargetTypeCode, m.TargetTypeDisplay)
	return err
}

func (r *sriRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*SubstanceReferenceInformation, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+sriCols+` FROM substance_reference_information WHERE id = $1`, id)) }
func (r *sriRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*SubstanceReferenceInformation, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+sriCols+` FROM substance_reference_information WHERE fhir_id = $1`, fhirID)) }

func (r *sriRepoPG) Update(ctx context.Context, m *SubstanceReferenceInformation) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE substance_reference_information SET comment=$2, gene_element_type_code=$3, gene_element_type_display=$4,
			gene_element_source_reference=$5, classification_code=$6, classification_display=$7,
			classification_domain_code=$8, classification_domain_display=$9,
			target_type_code=$10, target_type_display=$11, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.Comment, m.GeneElementTypeCode, m.GeneElementTypeDisplay,
		m.GeneElementSourceReference, m.ClassificationCode, m.ClassificationDisplay,
		m.ClassificationDomainCode, m.ClassificationDomainDisplay,
		m.TargetTypeCode, m.TargetTypeDisplay)
	return err
}

func (r *sriRepoPG) Delete(ctx context.Context, id uuid.UUID) error { _, err := r.conn(ctx).Exec(ctx, `DELETE FROM substance_reference_information WHERE id = $1`, id); return err }

func (r *sriRepoPG) List(ctx context.Context, limit, offset int) ([]*SubstanceReferenceInformation, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM substance_reference_information`).Scan(&total); err != nil { return nil, 0, err }
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+sriCols+` FROM substance_reference_information ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, err }; defer rows.Close()
	var items []*SubstanceReferenceInformation
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}

func (r *sriRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SubstanceReferenceInformation, int, error) {
	query := `SELECT ` + sriCols + ` FROM substance_reference_information WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM substance_reference_information WHERE 1=1`
	var args []interface{}; idx := 1
	_ = params
	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil { return nil, 0, err }
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1); args = append(args, limit, offset)
	rows, err := r.conn(ctx).Query(ctx, query, args...); if err != nil { return nil, 0, err }; defer rows.Close()
	var items []*SubstanceReferenceInformation
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}
