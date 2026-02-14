package terminology

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ehr/ehr/internal/platform/db"
)

type queryable interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// =========== LOINC Repository ===========

type loincRepoPG struct{ pool *pgxpool.Pool }

func NewLOINCRepoPG(pool *pgxpool.Pool) LOINCRepository { return &loincRepoPG{pool: pool} }

func (r *loincRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

func (r *loincRepoPG) Search(ctx context.Context, query string, limit int) ([]*LOINCCode, error) {
	if limit <= 0 {
		limit = 20
	}
	pattern := "%" + query + "%"
	rows, err := r.conn(ctx).Query(ctx,
		`SELECT code, display, COALESCE(component,''), COALESCE(property,''),
		        COALESCE(time_aspect,''), COALESCE(system_uri,'http://loinc.org'), COALESCE(category,'')
		 FROM reference_loinc
		 WHERE code ILIKE $1 OR display ILIKE $1 OR component ILIKE $1
		 ORDER BY display LIMIT $2`, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("loinc search: %w", err)
	}
	defer rows.Close()
	var results []*LOINCCode
	for rows.Next() {
		var c LOINCCode
		if err := rows.Scan(&c.Code, &c.Display, &c.Component, &c.Property, &c.TimeAspect, &c.SystemURI, &c.Category); err != nil {
			return nil, err
		}
		results = append(results, &c)
	}
	return results, rows.Err()
}

func (r *loincRepoPG) GetByCode(ctx context.Context, code string) (*LOINCCode, error) {
	var c LOINCCode
	err := r.conn(ctx).QueryRow(ctx,
		`SELECT code, display, COALESCE(component,''), COALESCE(property,''),
		        COALESCE(time_aspect,''), COALESCE(system_uri,'http://loinc.org'), COALESCE(category,'')
		 FROM reference_loinc WHERE code = $1`, code).
		Scan(&c.Code, &c.Display, &c.Component, &c.Property, &c.TimeAspect, &c.SystemURI, &c.Category)
	if err != nil {
		return nil, fmt.Errorf("loinc get: %w", err)
	}
	return &c, nil
}

// =========== ICD-10 Repository ===========

type icd10RepoPG struct{ pool *pgxpool.Pool }

func NewICD10RepoPG(pool *pgxpool.Pool) ICD10Repository { return &icd10RepoPG{pool: pool} }

func (r *icd10RepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

func (r *icd10RepoPG) Search(ctx context.Context, query string, limit int) ([]*ICD10Code, error) {
	if limit <= 0 {
		limit = 20
	}
	pattern := "%" + query + "%"
	rows, err := r.conn(ctx).Query(ctx,
		`SELECT code, display, COALESCE(category,''), COALESCE(chapter,''),
		        COALESCE(system_uri,'http://hl7.org/fhir/sid/icd-10-cm')
		 FROM reference_icd10
		 WHERE code ILIKE $1 OR display ILIKE $1 OR category ILIKE $1
		 ORDER BY code LIMIT $2`, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("icd10 search: %w", err)
	}
	defer rows.Close()
	var results []*ICD10Code
	for rows.Next() {
		var c ICD10Code
		if err := rows.Scan(&c.Code, &c.Display, &c.Category, &c.Chapter, &c.SystemURI); err != nil {
			return nil, err
		}
		results = append(results, &c)
	}
	return results, rows.Err()
}

func (r *icd10RepoPG) GetByCode(ctx context.Context, code string) (*ICD10Code, error) {
	var c ICD10Code
	err := r.conn(ctx).QueryRow(ctx,
		`SELECT code, display, COALESCE(category,''), COALESCE(chapter,''),
		        COALESCE(system_uri,'http://hl7.org/fhir/sid/icd-10-cm')
		 FROM reference_icd10 WHERE code = $1`, code).
		Scan(&c.Code, &c.Display, &c.Category, &c.Chapter, &c.SystemURI)
	if err != nil {
		return nil, fmt.Errorf("icd10 get: %w", err)
	}
	return &c, nil
}

// =========== SNOMED Repository ===========

type snomedRepoPG struct{ pool *pgxpool.Pool }

func NewSNOMEDRepoPG(pool *pgxpool.Pool) SNOMEDRepository { return &snomedRepoPG{pool: pool} }

func (r *snomedRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

func (r *snomedRepoPG) Search(ctx context.Context, query string, limit int) ([]*SNOMEDCode, error) {
	if limit <= 0 {
		limit = 20
	}
	pattern := "%" + query + "%"
	rows, err := r.conn(ctx).Query(ctx,
		`SELECT code, display, COALESCE(semantic_tag,''), COALESCE(category,''),
		        COALESCE(system_uri,'http://snomed.info/sct')
		 FROM reference_snomed
		 WHERE code ILIKE $1 OR display ILIKE $1 OR category ILIKE $1
		 ORDER BY display LIMIT $2`, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("snomed search: %w", err)
	}
	defer rows.Close()
	var results []*SNOMEDCode
	for rows.Next() {
		var c SNOMEDCode
		if err := rows.Scan(&c.Code, &c.Display, &c.SemanticTag, &c.Category, &c.SystemURI); err != nil {
			return nil, err
		}
		results = append(results, &c)
	}
	return results, rows.Err()
}

func (r *snomedRepoPG) GetByCode(ctx context.Context, code string) (*SNOMEDCode, error) {
	var c SNOMEDCode
	err := r.conn(ctx).QueryRow(ctx,
		`SELECT code, display, COALESCE(semantic_tag,''), COALESCE(category,''),
		        COALESCE(system_uri,'http://snomed.info/sct')
		 FROM reference_snomed WHERE code = $1`, code).
		Scan(&c.Code, &c.Display, &c.SemanticTag, &c.Category, &c.SystemURI)
	if err != nil {
		return nil, fmt.Errorf("snomed get: %w", err)
	}
	return &c, nil
}

// =========== RxNorm Repository ===========

type rxnormRepoPG struct{ pool *pgxpool.Pool }

func NewRxNormRepoPG(pool *pgxpool.Pool) RxNormRepository { return &rxnormRepoPG{pool: pool} }

func (r *rxnormRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

func (r *rxnormRepoPG) Search(ctx context.Context, query string, limit int) ([]*RxNormCode, error) {
	if limit <= 0 {
		limit = 20
	}
	pattern := "%" + query + "%"
	rows, err := r.conn(ctx).Query(ctx,
		`SELECT rxnorm_code, display, COALESCE(generic_name,''), COALESCE(drug_class,''),
		        COALESCE(route,''), COALESCE(form,''),
		        COALESCE(system_uri,'http://www.nlm.nih.gov/research/umls/rxnorm')
		 FROM reference_medication
		 WHERE rxnorm_code ILIKE $1 OR display ILIKE $1 OR generic_name ILIKE $1 OR drug_class ILIKE $1
		 ORDER BY display LIMIT $2`, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("rxnorm search: %w", err)
	}
	defer rows.Close()
	var results []*RxNormCode
	for rows.Next() {
		var c RxNormCode
		if err := rows.Scan(&c.RxNormCode, &c.Display, &c.GenericName, &c.DrugClass, &c.Route, &c.Form, &c.SystemURI); err != nil {
			return nil, err
		}
		results = append(results, &c)
	}
	return results, rows.Err()
}

func (r *rxnormRepoPG) GetByCode(ctx context.Context, code string) (*RxNormCode, error) {
	var c RxNormCode
	err := r.conn(ctx).QueryRow(ctx,
		`SELECT rxnorm_code, display, COALESCE(generic_name,''), COALESCE(drug_class,''),
		        COALESCE(route,''), COALESCE(form,''),
		        COALESCE(system_uri,'http://www.nlm.nih.gov/research/umls/rxnorm')
		 FROM reference_medication WHERE rxnorm_code = $1`, code).
		Scan(&c.RxNormCode, &c.Display, &c.GenericName, &c.DrugClass, &c.Route, &c.Form, &c.SystemURI)
	if err != nil {
		return nil, fmt.Errorf("rxnorm get: %w", err)
	}
	return &c, nil
}

// =========== CPT Repository ===========

type cptRepoPG struct{ pool *pgxpool.Pool }

func NewCPTRepoPG(pool *pgxpool.Pool) CPTRepository { return &cptRepoPG{pool: pool} }

func (r *cptRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

func (r *cptRepoPG) Search(ctx context.Context, query string, limit int) ([]*CPTCode, error) {
	if limit <= 0 {
		limit = 20
	}
	pattern := "%" + query + "%"
	rows, err := r.conn(ctx).Query(ctx,
		`SELECT code, display, COALESCE(category,''), COALESCE(subcategory,''),
		        COALESCE(system_uri,'http://www.ama-assn.org/go/cpt')
		 FROM reference_cpt
		 WHERE code ILIKE $1 OR display ILIKE $1 OR category ILIKE $1 OR subcategory ILIKE $1
		 ORDER BY code LIMIT $2`, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("cpt search: %w", err)
	}
	defer rows.Close()
	var results []*CPTCode
	for rows.Next() {
		var c CPTCode
		if err := rows.Scan(&c.Code, &c.Display, &c.Category, &c.Subcategory, &c.SystemURI); err != nil {
			return nil, err
		}
		results = append(results, &c)
	}
	return results, rows.Err()
}

func (r *cptRepoPG) GetByCode(ctx context.Context, code string) (*CPTCode, error) {
	var c CPTCode
	err := r.conn(ctx).QueryRow(ctx,
		`SELECT code, display, COALESCE(category,''), COALESCE(subcategory,''),
		        COALESCE(system_uri,'http://www.ama-assn.org/go/cpt')
		 FROM reference_cpt WHERE code = $1`, code).
		Scan(&c.Code, &c.Display, &c.Category, &c.Subcategory, &c.SystemURI)
	if err != nil {
		return nil, fmt.Errorf("cpt get: %w", err)
	}
	return &c, nil
}
