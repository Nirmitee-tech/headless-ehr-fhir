package terminology

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// =========== Mock Repositories ===========

type mockLOINCRepo struct {
	store map[string]*LOINCCode
}

func newMockLOINCRepo() *mockLOINCRepo {
	m := &mockLOINCRepo{store: make(map[string]*LOINCCode)}
	m.store["8310-5"] = &LOINCCode{Code: "8310-5", Display: "Body temperature", Component: "Body temperature", SystemURI: SystemLOINC, Category: "vital-signs"}
	m.store["8867-4"] = &LOINCCode{Code: "8867-4", Display: "Heart rate", Component: "Heart rate", SystemURI: SystemLOINC, Category: "vital-signs"}
	m.store["718-7"] = &LOINCCode{Code: "718-7", Display: "Hemoglobin", Component: "Hemoglobin", SystemURI: SystemLOINC, Category: "laboratory"}
	m.store["4548-4"] = &LOINCCode{Code: "4548-4", Display: "Hemoglobin A1c", Component: "HbA1c", SystemURI: SystemLOINC, Category: "laboratory"}
	return m
}

func (m *mockLOINCRepo) Search(_ context.Context, query string, limit int) ([]*LOINCCode, error) {
	var results []*LOINCCode
	q := strings.ToLower(query)
	for _, c := range m.store {
		if strings.Contains(strings.ToLower(c.Code), q) ||
			strings.Contains(strings.ToLower(c.Display), q) ||
			strings.Contains(strings.ToLower(c.Component), q) {
			results = append(results, c)
			if len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}

func (m *mockLOINCRepo) GetByCode(_ context.Context, code string) (*LOINCCode, error) {
	c, ok := m.store[code]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

type mockICD10Repo struct {
	store map[string]*ICD10Code
}

func newMockICD10Repo() *mockICD10Repo {
	m := &mockICD10Repo{store: make(map[string]*ICD10Code)}
	m.store["E11.9"] = &ICD10Code{Code: "E11.9", Display: "Type 2 diabetes mellitus without complications", Category: "Endocrine", SystemURI: SystemICD10}
	m.store["I10"] = &ICD10Code{Code: "I10", Display: "Essential (primary) hypertension", Category: "Circulatory", SystemURI: SystemICD10}
	m.store["J06.9"] = &ICD10Code{Code: "J06.9", Display: "Acute upper respiratory infection, unspecified", Category: "Respiratory", SystemURI: SystemICD10}
	m.store["M54.5"] = &ICD10Code{Code: "M54.5", Display: "Low back pain", Category: "Musculoskeletal", SystemURI: SystemICD10}
	return m
}

func (m *mockICD10Repo) Search(_ context.Context, query string, limit int) ([]*ICD10Code, error) {
	var results []*ICD10Code
	q := strings.ToLower(query)
	for _, c := range m.store {
		if strings.Contains(strings.ToLower(c.Code), q) ||
			strings.Contains(strings.ToLower(c.Display), q) ||
			strings.Contains(strings.ToLower(c.Category), q) {
			results = append(results, c)
			if len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}

func (m *mockICD10Repo) GetByCode(_ context.Context, code string) (*ICD10Code, error) {
	c, ok := m.store[code]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

type mockSNOMEDRepo struct {
	store map[string]*SNOMEDCode
}

func newMockSNOMEDRepo() *mockSNOMEDRepo {
	m := &mockSNOMEDRepo{store: make(map[string]*SNOMEDCode)}
	m.store["80146002"] = &SNOMEDCode{Code: "80146002", Display: "Appendectomy", SemanticTag: "procedure", Category: "surgical", SystemURI: SystemSNOMED}
	m.store["73761001"] = &SNOMEDCode{Code: "73761001", Display: "Colonoscopy", SemanticTag: "procedure", Category: "diagnostic", SystemURI: SystemSNOMED}
	m.store["38341003"] = &SNOMEDCode{Code: "38341003", Display: "Hypertensive disorder", SemanticTag: "finding", Category: "cardiovascular", SystemURI: SystemSNOMED}
	return m
}

func (m *mockSNOMEDRepo) Search(_ context.Context, query string, limit int) ([]*SNOMEDCode, error) {
	var results []*SNOMEDCode
	q := strings.ToLower(query)
	for _, c := range m.store {
		if strings.Contains(strings.ToLower(c.Code), q) ||
			strings.Contains(strings.ToLower(c.Display), q) {
			results = append(results, c)
			if len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}

func (m *mockSNOMEDRepo) GetByCode(_ context.Context, code string) (*SNOMEDCode, error) {
	c, ok := m.store[code]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

type mockRxNormRepo struct {
	store map[string]*RxNormCode
}

func newMockRxNormRepo() *mockRxNormRepo {
	m := &mockRxNormRepo{store: make(map[string]*RxNormCode)}
	m.store["860975"] = &RxNormCode{RxNormCode: "860975", Display: "Metformin 500 mg oral tablet", GenericName: "Metformin", DrugClass: "Biguanide", SystemURI: SystemRxNorm}
	m.store["314076"] = &RxNormCode{RxNormCode: "314076", Display: "Lisinopril 10 mg oral tablet", GenericName: "Lisinopril", DrugClass: "ACE Inhibitor", SystemURI: SystemRxNorm}
	m.store["259255"] = &RxNormCode{RxNormCode: "259255", Display: "Atorvastatin 10 mg oral tablet", GenericName: "Atorvastatin", DrugClass: "HMG-CoA Reductase Inhibitor", SystemURI: SystemRxNorm}
	return m
}

func (m *mockRxNormRepo) Search(_ context.Context, query string, limit int) ([]*RxNormCode, error) {
	var results []*RxNormCode
	q := strings.ToLower(query)
	for _, c := range m.store {
		if strings.Contains(strings.ToLower(c.RxNormCode), q) ||
			strings.Contains(strings.ToLower(c.Display), q) ||
			strings.Contains(strings.ToLower(c.GenericName), q) {
			results = append(results, c)
			if len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}

func (m *mockRxNormRepo) GetByCode(_ context.Context, code string) (*RxNormCode, error) {
	c, ok := m.store[code]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

type mockCPTRepo struct {
	store map[string]*CPTCode
}

func newMockCPTRepo() *mockCPTRepo {
	m := &mockCPTRepo{store: make(map[string]*CPTCode)}
	m.store["99213"] = &CPTCode{Code: "99213", Display: "Office visit, established patient, low complexity", Category: "E&M", SystemURI: SystemCPT}
	m.store["99214"] = &CPTCode{Code: "99214", Display: "Office visit, established patient, moderate complexity", Category: "E&M", SystemURI: SystemCPT}
	m.store["93000"] = &CPTCode{Code: "93000", Display: "Electrocardiogram, routine, 12-lead", Category: "Medicine", SystemURI: SystemCPT}
	return m
}

func (m *mockCPTRepo) Search(_ context.Context, query string, limit int) ([]*CPTCode, error) {
	var results []*CPTCode
	q := strings.ToLower(query)
	for _, c := range m.store {
		if strings.Contains(strings.ToLower(c.Code), q) ||
			strings.Contains(strings.ToLower(c.Display), q) ||
			strings.Contains(strings.ToLower(c.Category), q) {
			results = append(results, c)
			if len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}

func (m *mockCPTRepo) GetByCode(_ context.Context, code string) (*CPTCode, error) {
	c, ok := m.store[code]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

// =========== Helper ===========

func newTestService() *Service {
	return NewService(
		newMockLOINCRepo(),
		newMockICD10Repo(),
		newMockSNOMEDRepo(),
		newMockRxNormRepo(),
		newMockCPTRepo(),
	)
}

// =========== LOINC Tests ===========

func TestSearchLOINC_Success(t *testing.T) {
	svc := newTestService()
	results, err := svc.SearchLOINC(context.Background(), "heart", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for 'heart'")
	}
	for _, r := range results {
		if !strings.Contains(strings.ToLower(r.Display), "heart") && !strings.Contains(strings.ToLower(r.Component), "heart") {
			t.Errorf("result %s does not match 'heart'", r.Code)
		}
	}
}

func TestSearchLOINC_EmptyQuery(t *testing.T) {
	svc := newTestService()
	_, err := svc.SearchLOINC(context.Background(), "", 20)
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestSearchLOINC_ByCode(t *testing.T) {
	svc := newTestService()
	results, err := svc.SearchLOINC(context.Background(), "8310", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for code '8310', got %d", len(results))
	}
}

func TestSearchLOINC_DefaultLimit(t *testing.T) {
	svc := newTestService()
	results, err := svc.SearchLOINC(context.Background(), "hemoglobin", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should use default limit of 20 and return matching results
	if results == nil {
		t.Error("expected non-nil results")
	}
}

func TestLookupLOINC_Success(t *testing.T) {
	svc := newTestService()
	code, err := svc.LookupLOINC(context.Background(), "8310-5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code.Display != "Body temperature" {
		t.Errorf("expected 'Body temperature', got %q", code.Display)
	}
}

func TestLookupLOINC_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.LookupLOINC(context.Background(), "99999-9")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestLookupLOINC_EmptyCode(t *testing.T) {
	svc := newTestService()
	_, err := svc.LookupLOINC(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty code")
	}
}

// =========== ICD-10 Tests ===========

func TestSearchICD10_Success(t *testing.T) {
	svc := newTestService()
	results, err := svc.SearchICD10(context.Background(), "diabetes", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for 'diabetes'")
	}
}

func TestSearchICD10_EmptyQuery(t *testing.T) {
	svc := newTestService()
	_, err := svc.SearchICD10(context.Background(), "", 20)
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestLookupICD10_Success(t *testing.T) {
	svc := newTestService()
	code, err := svc.LookupICD10(context.Background(), "I10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(code.Display, "hypertension") {
		t.Errorf("expected display to contain 'hypertension', got %q", code.Display)
	}
}

func TestLookupICD10_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.LookupICD10(context.Background(), "ZZZ99")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

// =========== SNOMED Tests ===========

func TestSearchSNOMED_Success(t *testing.T) {
	svc := newTestService()
	results, err := svc.SearchSNOMED(context.Background(), "appendectomy", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for 'appendectomy'")
	}
}

func TestSearchSNOMED_EmptyQuery(t *testing.T) {
	svc := newTestService()
	_, err := svc.SearchSNOMED(context.Background(), "", 20)
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestLookupSNOMED_Success(t *testing.T) {
	svc := newTestService()
	code, err := svc.LookupSNOMED(context.Background(), "80146002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code.Display != "Appendectomy" {
		t.Errorf("expected 'Appendectomy', got %q", code.Display)
	}
}

// =========== RxNorm Tests ===========

func TestSearchRxNorm_Success(t *testing.T) {
	svc := newTestService()
	results, err := svc.SearchRxNorm(context.Background(), "metformin", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for 'metformin'")
	}
}

func TestSearchRxNorm_EmptyQuery(t *testing.T) {
	svc := newTestService()
	_, err := svc.SearchRxNorm(context.Background(), "", 20)
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestLookupRxNorm_Success(t *testing.T) {
	svc := newTestService()
	code, err := svc.LookupRxNorm(context.Background(), "860975")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(code.Display, "Metformin") {
		t.Errorf("expected display to contain 'Metformin', got %q", code.Display)
	}
}

// =========== CPT Tests ===========

func TestSearchCPT_Success(t *testing.T) {
	svc := newTestService()
	results, err := svc.SearchCPT(context.Background(), "99213", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for '99213'")
	}
}

func TestSearchCPT_EmptyQuery(t *testing.T) {
	svc := newTestService()
	_, err := svc.SearchCPT(context.Background(), "", 20)
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestLookupCPT_Success(t *testing.T) {
	svc := newTestService()
	code, err := svc.LookupCPT(context.Background(), "99213")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(code.Display, "Office visit") {
		t.Errorf("expected display to contain 'Office visit', got %q", code.Display)
	}
}

func TestLookupCPT_EmptyCode(t *testing.T) {
	svc := newTestService()
	_, err := svc.LookupCPT(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty code")
	}
}

// =========== FHIR $lookup Tests ===========

func TestFHIRLookup_LOINC(t *testing.T) {
	svc := newTestService()
	resp, err := svc.Lookup(context.Background(), &LookupRequest{System: SystemLOINC, Code: "8310-5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ResourceType != "Parameters" {
		t.Errorf("expected resourceType 'Parameters', got %q", resp.ResourceType)
	}
	found := false
	for _, p := range resp.Parameter {
		if p.Name == "display" && p.ValueString == "Body temperature" {
			found = true
		}
	}
	if !found {
		t.Error("expected display parameter with value 'Body temperature'")
	}
}

func TestFHIRLookup_ICD10(t *testing.T) {
	svc := newTestService()
	resp, err := svc.Lookup(context.Background(), &LookupRequest{System: SystemICD10, Code: "I10"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, p := range resp.Parameter {
		if p.Name == "display" && strings.Contains(p.ValueString, "hypertension") {
			found = true
		}
	}
	if !found {
		t.Error("expected display parameter containing 'hypertension'")
	}
}

func TestFHIRLookup_SNOMED(t *testing.T) {
	svc := newTestService()
	resp, err := svc.Lookup(context.Background(), &LookupRequest{System: SystemSNOMED, Code: "80146002"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ResourceType != "Parameters" {
		t.Errorf("expected resourceType 'Parameters', got %q", resp.ResourceType)
	}
}

func TestFHIRLookup_RxNorm(t *testing.T) {
	svc := newTestService()
	resp, err := svc.Lookup(context.Background(), &LookupRequest{System: SystemRxNorm, Code: "860975"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ResourceType != "Parameters" {
		t.Errorf("expected resourceType 'Parameters', got %q", resp.ResourceType)
	}
}

func TestFHIRLookup_CPT(t *testing.T) {
	svc := newTestService()
	resp, err := svc.Lookup(context.Background(), &LookupRequest{System: SystemCPT, Code: "99213"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ResourceType != "Parameters" {
		t.Errorf("expected resourceType 'Parameters', got %q", resp.ResourceType)
	}
}

func TestFHIRLookup_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.Lookup(context.Background(), &LookupRequest{System: SystemLOINC, Code: "99999-9"})
	if err == nil {
		t.Fatal("expected error for not found code")
	}
}

func TestFHIRLookup_MissingSystem(t *testing.T) {
	svc := newTestService()
	_, err := svc.Lookup(context.Background(), &LookupRequest{Code: "8310-5"})
	if err == nil {
		t.Fatal("expected error for missing system")
	}
}

func TestFHIRLookup_MissingCode(t *testing.T) {
	svc := newTestService()
	_, err := svc.Lookup(context.Background(), &LookupRequest{System: SystemLOINC})
	if err == nil {
		t.Fatal("expected error for missing code")
	}
}

func TestFHIRLookup_UnsupportedSystem(t *testing.T) {
	svc := newTestService()
	_, err := svc.Lookup(context.Background(), &LookupRequest{System: "http://unknown.system", Code: "12345"})
	if err == nil {
		t.Fatal("expected error for unsupported system")
	}
}

// =========== FHIR $validate-code Tests ===========

func TestFHIRValidateCode_Valid(t *testing.T) {
	svc := newTestService()
	resp, err := svc.ValidateCode(context.Background(), &ValidateCodeRequest{System: SystemLOINC, Code: "8310-5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ResourceType != "Parameters" {
		t.Errorf("expected resourceType 'Parameters', got %q", resp.ResourceType)
	}
	for _, p := range resp.Parameter {
		if p.Name == "result" && p.ValueBoolean != nil && !*p.ValueBoolean {
			t.Error("expected result to be true for valid code")
		}
	}
}

func TestFHIRValidateCode_Invalid(t *testing.T) {
	svc := newTestService()
	resp, err := svc.ValidateCode(context.Background(), &ValidateCodeRequest{System: SystemLOINC, Code: "99999-9"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range resp.Parameter {
		if p.Name == "result" && p.ValueBoolean != nil && *p.ValueBoolean {
			t.Error("expected result to be false for invalid code")
		}
	}
}

func TestFHIRValidateCode_MissingSystem(t *testing.T) {
	svc := newTestService()
	_, err := svc.ValidateCode(context.Background(), &ValidateCodeRequest{Code: "8310-5"})
	if err == nil {
		t.Fatal("expected error for missing system")
	}
}

func TestFHIRValidateCode_MissingCode(t *testing.T) {
	svc := newTestService()
	_, err := svc.ValidateCode(context.Background(), &ValidateCodeRequest{System: SystemLOINC})
	if err == nil {
		t.Fatal("expected error for missing code")
	}
}

func TestFHIRValidateCode_UnsupportedSystem(t *testing.T) {
	svc := newTestService()
	_, err := svc.ValidateCode(context.Background(), &ValidateCodeRequest{System: "http://unknown.system", Code: "12345"})
	if err == nil {
		t.Fatal("expected error for unsupported system")
	}
}

func TestFHIRValidateCode_ICD10Valid(t *testing.T) {
	svc := newTestService()
	resp, err := svc.ValidateCode(context.Background(), &ValidateCodeRequest{System: SystemICD10, Code: "E11.9"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range resp.Parameter {
		if p.Name == "result" && p.ValueBoolean != nil && !*p.ValueBoolean {
			t.Error("expected result to be true for valid ICD-10 code")
		}
	}
}

func TestFHIRValidateCode_SNOMEDValid(t *testing.T) {
	svc := newTestService()
	resp, err := svc.ValidateCode(context.Background(), &ValidateCodeRequest{System: SystemSNOMED, Code: "80146002"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range resp.Parameter {
		if p.Name == "result" && p.ValueBoolean != nil && !*p.ValueBoolean {
			t.Error("expected result to be true for valid SNOMED code")
		}
	}
}

func TestFHIRValidateCode_RxNormValid(t *testing.T) {
	svc := newTestService()
	resp, err := svc.ValidateCode(context.Background(), &ValidateCodeRequest{System: SystemRxNorm, Code: "860975"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range resp.Parameter {
		if p.Name == "result" && p.ValueBoolean != nil && !*p.ValueBoolean {
			t.Error("expected result to be true for valid RxNorm code")
		}
	}
}

func TestFHIRValidateCode_CPTValid(t *testing.T) {
	svc := newTestService()
	resp, err := svc.ValidateCode(context.Background(), &ValidateCodeRequest{System: SystemCPT, Code: "99213"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range resp.Parameter {
		if p.Name == "result" && p.ValueBoolean != nil && !*p.ValueBoolean {
			t.Error("expected result to be true for valid CPT code")
		}
	}
}
