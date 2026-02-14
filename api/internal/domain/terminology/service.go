package terminology

import (
	"context"
	"fmt"
)

// Service provides terminology lookup and validation operations.
type Service struct {
	loinc  LOINCRepository
	icd10  ICD10Repository
	snomed SNOMEDRepository
	rxnorm RxNormRepository
	cpt    CPTRepository
}

// NewService creates a new terminology service.
func NewService(loinc LOINCRepository, icd10 ICD10Repository, snomed SNOMEDRepository, rxnorm RxNormRepository, cpt CPTRepository) *Service {
	return &Service{loinc: loinc, icd10: icd10, snomed: snomed, rxnorm: rxnorm, cpt: cpt}
}

// -- LOINC --

// SearchLOINC searches LOINC codes by query text.
func (s *Service) SearchLOINC(ctx context.Context, query string, limit int) ([]*LOINCCode, error) {
	if query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}
	if limit <= 0 {
		limit = 20
	}
	return s.loinc.Search(ctx, query, limit)
}

// LookupLOINC looks up a single LOINC code.
func (s *Service) LookupLOINC(ctx context.Context, code string) (*LOINCCode, error) {
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}
	return s.loinc.GetByCode(ctx, code)
}

// -- ICD-10 --

// SearchICD10 searches ICD-10-CM codes by query text.
func (s *Service) SearchICD10(ctx context.Context, query string, limit int) ([]*ICD10Code, error) {
	if query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}
	if limit <= 0 {
		limit = 20
	}
	return s.icd10.Search(ctx, query, limit)
}

// LookupICD10 looks up a single ICD-10 code.
func (s *Service) LookupICD10(ctx context.Context, code string) (*ICD10Code, error) {
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}
	return s.icd10.GetByCode(ctx, code)
}

// -- SNOMED --

// SearchSNOMED searches SNOMED CT codes by query text.
func (s *Service) SearchSNOMED(ctx context.Context, query string, limit int) ([]*SNOMEDCode, error) {
	if query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}
	if limit <= 0 {
		limit = 20
	}
	return s.snomed.Search(ctx, query, limit)
}

// LookupSNOMED looks up a single SNOMED CT code.
func (s *Service) LookupSNOMED(ctx context.Context, code string) (*SNOMEDCode, error) {
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}
	return s.snomed.GetByCode(ctx, code)
}

// -- RxNorm --

// SearchRxNorm searches RxNorm medication codes by query text.
func (s *Service) SearchRxNorm(ctx context.Context, query string, limit int) ([]*RxNormCode, error) {
	if query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}
	if limit <= 0 {
		limit = 20
	}
	return s.rxnorm.Search(ctx, query, limit)
}

// LookupRxNorm looks up a single RxNorm code.
func (s *Service) LookupRxNorm(ctx context.Context, code string) (*RxNormCode, error) {
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}
	return s.rxnorm.GetByCode(ctx, code)
}

// -- CPT --

// SearchCPT searches CPT codes by query text.
func (s *Service) SearchCPT(ctx context.Context, query string, limit int) ([]*CPTCode, error) {
	if query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}
	if limit <= 0 {
		limit = 20
	}
	return s.cpt.Search(ctx, query, limit)
}

// LookupCPT looks up a single CPT code.
func (s *Service) LookupCPT(ctx context.Context, code string) (*CPTCode, error) {
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}
	return s.cpt.GetByCode(ctx, code)
}

// -- FHIR Operations --

// Lookup implements the FHIR CodeSystem $lookup operation.
// It resolves the system URI to the appropriate code system and returns a Parameters resource.
func (s *Service) Lookup(ctx context.Context, req *LookupRequest) (*LookupResponse, error) {
	if req.System == "" {
		return nil, fmt.Errorf("system is required")
	}
	if req.Code == "" {
		return nil, fmt.Errorf("code is required")
	}

	var display string
	switch req.System {
	case SystemLOINC:
		c, err := s.loinc.GetByCode(ctx, req.Code)
		if err != nil {
			return nil, fmt.Errorf("code not found in LOINC: %s", req.Code)
		}
		display = c.Display
	case SystemICD10:
		c, err := s.icd10.GetByCode(ctx, req.Code)
		if err != nil {
			return nil, fmt.Errorf("code not found in ICD-10: %s", req.Code)
		}
		display = c.Display
	case SystemSNOMED:
		c, err := s.snomed.GetByCode(ctx, req.Code)
		if err != nil {
			return nil, fmt.Errorf("code not found in SNOMED CT: %s", req.Code)
		}
		display = c.Display
	case SystemRxNorm:
		c, err := s.rxnorm.GetByCode(ctx, req.Code)
		if err != nil {
			return nil, fmt.Errorf("code not found in RxNorm: %s", req.Code)
		}
		display = c.Display
	case SystemCPT:
		c, err := s.cpt.GetByCode(ctx, req.Code)
		if err != nil {
			return nil, fmt.Errorf("code not found in CPT: %s", req.Code)
		}
		display = c.Display
	default:
		return nil, fmt.Errorf("unsupported code system: %s", req.System)
	}

	return &LookupResponse{
		ResourceType: "Parameters",
		Parameter: []LookupParameter{
			{Name: "name", ValueString: display},
			{Name: "display", ValueString: display},
		},
	}, nil
}

// ValidateCode implements the FHIR CodeSystem $validate-code operation.
func (s *Service) ValidateCode(ctx context.Context, req *ValidateCodeRequest) (*ValidateCodeResponse, error) {
	if req.System == "" {
		return nil, fmt.Errorf("system is required")
	}
	if req.Code == "" {
		return nil, fmt.Errorf("code is required")
	}

	var display string
	var found bool

	switch req.System {
	case SystemLOINC:
		c, err := s.loinc.GetByCode(ctx, req.Code)
		if err == nil {
			found = true
			display = c.Display
		}
	case SystemICD10:
		c, err := s.icd10.GetByCode(ctx, req.Code)
		if err == nil {
			found = true
			display = c.Display
		}
	case SystemSNOMED:
		c, err := s.snomed.GetByCode(ctx, req.Code)
		if err == nil {
			found = true
			display = c.Display
		}
	case SystemRxNorm:
		c, err := s.rxnorm.GetByCode(ctx, req.Code)
		if err == nil {
			found = true
			display = c.Display
		}
	case SystemCPT:
		c, err := s.cpt.GetByCode(ctx, req.Code)
		if err == nil {
			found = true
			display = c.Display
		}
	default:
		return nil, fmt.Errorf("unsupported code system: %s", req.System)
	}

	result := found
	params := []ValidateCodeParameter{
		{Name: "result", ValueBoolean: &result},
	}
	if found {
		params = append(params, ValidateCodeParameter{Name: "display", ValueString: display})
	} else {
		params = append(params, ValidateCodeParameter{Name: "message", ValueString: fmt.Sprintf("code '%s' not found in system '%s'", req.Code, req.System)})
	}

	return &ValidateCodeResponse{
		ResourceType: "Parameters",
		Parameter:    params,
	}, nil
}
