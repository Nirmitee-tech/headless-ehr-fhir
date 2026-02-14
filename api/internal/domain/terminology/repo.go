package terminology

import "context"

// LOINCRepository provides access to LOINC reference codes.
type LOINCRepository interface {
	Search(ctx context.Context, query string, limit int) ([]*LOINCCode, error)
	GetByCode(ctx context.Context, code string) (*LOINCCode, error)
}

// ICD10Repository provides access to ICD-10-CM reference codes.
type ICD10Repository interface {
	Search(ctx context.Context, query string, limit int) ([]*ICD10Code, error)
	GetByCode(ctx context.Context, code string) (*ICD10Code, error)
}

// SNOMEDRepository provides access to SNOMED CT reference codes.
type SNOMEDRepository interface {
	Search(ctx context.Context, query string, limit int) ([]*SNOMEDCode, error)
	GetByCode(ctx context.Context, code string) (*SNOMEDCode, error)
}

// RxNormRepository provides access to medication RxNorm codes.
type RxNormRepository interface {
	Search(ctx context.Context, query string, limit int) ([]*RxNormCode, error)
	GetByCode(ctx context.Context, code string) (*RxNormCode, error)
}

// CPTRepository provides access to CPT procedure codes.
type CPTRepository interface {
	Search(ctx context.Context, query string, limit int) ([]*CPTCode, error)
	GetByCode(ctx context.Context, code string) (*CPTCode, error)
}
