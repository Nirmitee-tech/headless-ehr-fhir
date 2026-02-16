package fhir

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// SearchParameterStore tests
// ---------------------------------------------------------------------------

func TestNewSearchParameterStore(t *testing.T) {
	store := NewSearchParameterStore()
	if store == nil {
		t.Fatal("expected non-nil store")
	}
	if got := len(store.List()); got != 0 {
		t.Fatalf("expected empty store, got %d entries", got)
	}
}

func TestSearchParameterStore_Create(t *testing.T) {
	store := NewSearchParameterStore()

	sp := &SearchParameterResource{
		ResourceType: "SearchParameter",
		ID:           "test-param",
		URL:          "http://example.com/fhir/SearchParameter/test-param",
		Name:         "TestParam",
		Status:       "active",
		Code:         "test",
		Base:         []string{"Patient"},
		Type:         "string",
	}

	if err := store.Create(sp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it was stored.
	got, err := store.Get("test-param")
	if err != nil {
		t.Fatalf("unexpected error on Get: %v", err)
	}
	if got.Name != "TestParam" {
		t.Errorf("expected Name=TestParam, got %q", got.Name)
	}
	if got.ResourceType != "SearchParameter" {
		t.Errorf("expected ResourceType=SearchParameter, got %q", got.ResourceType)
	}
}

func TestSearchParameterStore_Create_DefaultsResourceType(t *testing.T) {
	store := NewSearchParameterStore()

	sp := &SearchParameterResource{
		ID:     "no-rt",
		URL:    "http://example.com/fhir/SearchParameter/no-rt",
		Name:   "NoRT",
		Status: "active",
		Code:   "no-rt",
		Base:   []string{"Patient"},
		Type:   "string",
	}

	if err := store.Create(sp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := store.Get("no-rt")
	if got.ResourceType != "SearchParameter" {
		t.Errorf("expected ResourceType to default to SearchParameter, got %q", got.ResourceType)
	}
}

func TestSearchParameterStore_Create_NoID(t *testing.T) {
	store := NewSearchParameterStore()

	sp := &SearchParameterResource{
		URL:  "http://example.com/fhir/SearchParameter/x",
		Name: "X",
	}

	err := store.Create(sp)
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestSearchParameterStore_Create_Duplicate(t *testing.T) {
	store := NewSearchParameterStore()

	sp := &SearchParameterResource{
		ID:   "dup",
		URL:  "http://example.com/fhir/SearchParameter/dup",
		Name: "Dup",
	}

	_ = store.Create(sp)
	err := store.Create(sp)
	if err == nil {
		t.Fatal("expected error for duplicate ID")
	}
}

func TestSearchParameterStore_Create_IsolatesMutation(t *testing.T) {
	store := NewSearchParameterStore()

	sp := &SearchParameterResource{
		ID:   "iso",
		URL:  "http://example.com/fhir/SearchParameter/iso",
		Name: "Original",
	}
	_ = store.Create(sp)

	// Mutate the original after creation.
	sp.Name = "Mutated"

	got, _ := store.Get("iso")
	if got.Name != "Original" {
		t.Error("store should hold a copy; mutation should not propagate")
	}
}

func TestSearchParameterStore_Get_NotFound(t *testing.T) {
	store := NewSearchParameterStore()

	_, err := store.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestSearchParameterStore_Get_IsolatesMutation(t *testing.T) {
	store := NewSearchParameterStore()

	_ = store.Create(&SearchParameterResource{
		ID:   "g-iso",
		URL:  "http://example.com/fhir/SearchParameter/g-iso",
		Name: "Before",
	})

	got, _ := store.Get("g-iso")
	got.Name = "After"

	got2, _ := store.Get("g-iso")
	if got2.Name != "Before" {
		t.Error("Get should return a copy; mutation should not propagate")
	}
}

func TestSearchParameterStore_Update(t *testing.T) {
	store := NewSearchParameterStore()

	_ = store.Create(&SearchParameterResource{
		ID:     "upd",
		URL:    "http://example.com/fhir/SearchParameter/upd",
		Name:   "Before",
		Status: "draft",
		Code:   "upd",
		Base:   []string{"Patient"},
		Type:   "string",
	})

	updated := &SearchParameterResource{
		URL:    "http://example.com/fhir/SearchParameter/upd",
		Name:   "After",
		Status: "active",
		Code:   "upd",
		Base:   []string{"Patient"},
		Type:   "token",
	}
	if err := store.Update("upd", updated); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := store.Get("upd")
	if got.Name != "After" {
		t.Errorf("expected Name=After, got %q", got.Name)
	}
	if got.Type != "token" {
		t.Errorf("expected Type=token, got %q", got.Type)
	}
	if got.ID != "upd" {
		t.Errorf("expected ID to be set from path, got %q", got.ID)
	}
}

func TestSearchParameterStore_Update_NotFound(t *testing.T) {
	store := NewSearchParameterStore()

	err := store.Update("nope", &SearchParameterResource{Name: "X"})
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestSearchParameterStore_Delete(t *testing.T) {
	store := NewSearchParameterStore()

	_ = store.Create(&SearchParameterResource{
		ID:   "del",
		URL:  "http://example.com/fhir/SearchParameter/del",
		Name: "Del",
	})

	if err := store.Delete("del"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := store.Get("del")
	if err == nil {
		t.Fatal("expected not-found after delete")
	}
}

func TestSearchParameterStore_Delete_NotFound(t *testing.T) {
	store := NewSearchParameterStore()

	err := store.Delete("nope")
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestSearchParameterStore_Search_NoFilters(t *testing.T) {
	store := NewSearchParameterStore()

	_ = store.Create(&SearchParameterResource{ID: "a", URL: "u1", Name: "A"})
	_ = store.Create(&SearchParameterResource{ID: "b", URL: "u2", Name: "B"})
	_ = store.Create(&SearchParameterResource{ID: "c", URL: "u3", Name: "C"})

	results := store.Search(nil)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Verify sorted order.
	if results[0].ID != "a" || results[1].ID != "b" || results[2].ID != "c" {
		t.Error("expected results sorted by ID")
	}
}

func TestSearchParameterStore_Search_ByName(t *testing.T) {
	store := NewSearchParameterStore()

	_ = store.Create(&SearchParameterResource{ID: "1", URL: "u1", Name: "Alpha"})
	_ = store.Create(&SearchParameterResource{ID: "2", URL: "u2", Name: "Beta"})

	results := store.Search(map[string]string{"name": "alpha"}) // case-insensitive
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "Alpha" {
		t.Errorf("expected Alpha, got %q", results[0].Name)
	}
}

func TestSearchParameterStore_Search_ByCode(t *testing.T) {
	store := NewSearchParameterStore()

	_ = store.Create(&SearchParameterResource{ID: "1", URL: "u1", Name: "A", Code: "name"})
	_ = store.Create(&SearchParameterResource{ID: "2", URL: "u2", Name: "B", Code: "code"})
	_ = store.Create(&SearchParameterResource{ID: "3", URL: "u3", Name: "C", Code: "name"})

	results := store.Search(map[string]string{"code": "name"})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestSearchParameterStore_Search_ByStatus(t *testing.T) {
	store := NewSearchParameterStore()

	_ = store.Create(&SearchParameterResource{ID: "1", URL: "u1", Name: "A", Status: "active"})
	_ = store.Create(&SearchParameterResource{ID: "2", URL: "u2", Name: "B", Status: "draft"})
	_ = store.Create(&SearchParameterResource{ID: "3", URL: "u3", Name: "C", Status: "active"})

	results := store.Search(map[string]string{"status": "active"})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestSearchParameterStore_Search_ByType(t *testing.T) {
	store := NewSearchParameterStore()

	_ = store.Create(&SearchParameterResource{ID: "1", URL: "u1", Name: "A", Type: "token"})
	_ = store.Create(&SearchParameterResource{ID: "2", URL: "u2", Name: "B", Type: "string"})

	results := store.Search(map[string]string{"type": "token"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Type != "token" {
		t.Errorf("expected token, got %q", results[0].Type)
	}
}

func TestSearchParameterStore_Search_ByBase(t *testing.T) {
	store := NewSearchParameterStore()

	_ = store.Create(&SearchParameterResource{ID: "1", URL: "u1", Name: "A", Base: []string{"Patient"}})
	_ = store.Create(&SearchParameterResource{ID: "2", URL: "u2", Name: "B", Base: []string{"Observation"}})
	_ = store.Create(&SearchParameterResource{ID: "3", URL: "u3", Name: "C", Base: []string{"Patient", "Practitioner"}})

	results := store.Search(map[string]string{"base": "patient"}) // case-insensitive
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestSearchParameterStore_Search_ByURL(t *testing.T) {
	store := NewSearchParameterStore()

	_ = store.Create(&SearchParameterResource{ID: "1", URL: "http://hl7.org/fhir/SearchParameter/Patient-name", Name: "A"})
	_ = store.Create(&SearchParameterResource{ID: "2", URL: "http://hl7.org/fhir/SearchParameter/Patient-gender", Name: "B"})

	results := store.Search(map[string]string{"url": "http://hl7.org/fhir/SearchParameter/Patient-name"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSearchParameterStore_Search_MultipleFilters(t *testing.T) {
	store := NewSearchParameterStore()

	_ = store.Create(&SearchParameterResource{ID: "1", URL: "u1", Name: "A", Status: "active", Type: "token"})
	_ = store.Create(&SearchParameterResource{ID: "2", URL: "u2", Name: "B", Status: "active", Type: "string"})
	_ = store.Create(&SearchParameterResource{ID: "3", URL: "u3", Name: "C", Status: "draft", Type: "token"})

	results := store.Search(map[string]string{"status": "active", "type": "token"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "1" {
		t.Errorf("expected ID=1, got %q", results[0].ID)
	}
}

// ---------------------------------------------------------------------------
// DefaultSearchParameters tests
// ---------------------------------------------------------------------------

func TestDefaultSearchParameters_Count(t *testing.T) {
	params := DefaultSearchParameters()
	if len(params) < 20 {
		t.Fatalf("expected at least 20 default search parameters, got %d", len(params))
	}
}

func TestDefaultSearchParameters_UniqueIDs(t *testing.T) {
	params := DefaultSearchParameters()
	seen := make(map[string]bool, len(params))
	for _, sp := range params {
		if seen[sp.ID] {
			t.Errorf("duplicate ID: %q", sp.ID)
		}
		seen[sp.ID] = true
	}
}

func TestDefaultSearchParameters_AllHaveResourceType(t *testing.T) {
	for _, sp := range DefaultSearchParameters() {
		if sp.ResourceType != "SearchParameter" {
			t.Errorf("ID=%q: expected ResourceType=SearchParameter, got %q", sp.ID, sp.ResourceType)
		}
	}
}

func TestDefaultSearchParameters_ContainsCrossResourceParams(t *testing.T) {
	params := DefaultSearchParameters()
	expected := map[string]bool{
		"_id": false, "_lastUpdated": false, "_tag": false, "_security": false,
		"_profile": false, "_text": false, "_content": false, "_has": false, "_list": false,
	}
	for _, sp := range params {
		if _, ok := expected[sp.Code]; ok {
			expected[sp.Code] = true
		}
	}
	for code, found := range expected {
		if !found {
			t.Errorf("expected cross-resource param %q not found in defaults", code)
		}
	}
}

func TestDefaultSearchParameters_ContainsResourceSpecific(t *testing.T) {
	params := DefaultSearchParameters()

	// Check for some resource-specific params.
	found := map[string]bool{}
	for _, sp := range params {
		key := sp.Base[0] + "." + sp.Code
		found[key] = true
	}

	checks := []string{
		"Patient.name", "Patient.birthdate", "Patient.gender",
		"Observation.code", "Observation.patient", "Observation.category",
		"Encounter.status", "Encounter.class",
		"Condition.code", "Condition.clinical-status",
		"MedicationRequest.patient", "MedicationRequest.status",
	}
	for _, c := range checks {
		if !found[c] {
			t.Errorf("expected resource-specific param %q not found", c)
		}
	}
}

func TestNewDefaultSearchParameterStore(t *testing.T) {
	store := NewDefaultSearchParameterStore()
	all := store.List()
	if len(all) < 20 {
		t.Fatalf("expected at least 20 entries in default store, got %d", len(all))
	}
}

// ---------------------------------------------------------------------------
// validateSearchParameter tests
// ---------------------------------------------------------------------------

func TestValidateSearchParameter_Valid(t *testing.T) {
	sp := &SearchParameterResource{
		URL:    "http://example.com/fhir/SearchParameter/test",
		Name:   "Test",
		Status: "active",
		Code:   "test",
		Base:   []string{"Patient"},
		Type:   "string",
	}
	if err := validateSearchParameter(sp); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateSearchParameter_MissingURL(t *testing.T) {
	sp := &SearchParameterResource{Name: "T", Status: "active", Code: "t", Base: []string{"P"}, Type: "string"}
	sp.URL = ""
	if err := validateSearchParameter(sp); err == nil {
		t.Error("expected error for missing URL")
	}
}

func TestValidateSearchParameter_MissingName(t *testing.T) {
	sp := &SearchParameterResource{URL: "u", Status: "active", Code: "t", Base: []string{"P"}, Type: "string"}
	if err := validateSearchParameter(sp); err == nil {
		t.Error("expected error for missing Name")
	}
}

func TestValidateSearchParameter_MissingStatus(t *testing.T) {
	sp := &SearchParameterResource{URL: "u", Name: "N", Code: "t", Base: []string{"P"}, Type: "string"}
	if err := validateSearchParameter(sp); err == nil {
		t.Error("expected error for missing Status")
	}
}

func TestValidateSearchParameter_InvalidStatus(t *testing.T) {
	sp := &SearchParameterResource{URL: "u", Name: "N", Status: "bogus", Code: "t", Base: []string{"P"}, Type: "string"}
	if err := validateSearchParameter(sp); err == nil {
		t.Error("expected error for invalid Status")
	}
}

func TestValidateSearchParameter_MissingCode(t *testing.T) {
	sp := &SearchParameterResource{URL: "u", Name: "N", Status: "active", Base: []string{"P"}, Type: "string"}
	if err := validateSearchParameter(sp); err == nil {
		t.Error("expected error for missing Code")
	}
}

func TestValidateSearchParameter_MissingBase(t *testing.T) {
	sp := &SearchParameterResource{URL: "u", Name: "N", Status: "active", Code: "t", Type: "string"}
	if err := validateSearchParameter(sp); err == nil {
		t.Error("expected error for missing Base")
	}
}

func TestValidateSearchParameter_MissingType(t *testing.T) {
	sp := &SearchParameterResource{URL: "u", Name: "N", Status: "active", Code: "t", Base: []string{"P"}}
	if err := validateSearchParameter(sp); err == nil {
		t.Error("expected error for missing Type")
	}
}

func TestValidateSearchParameter_InvalidType(t *testing.T) {
	sp := &SearchParameterResource{URL: "u", Name: "N", Status: "active", Code: "t", Base: []string{"P"}, Type: "invalid"}
	if err := validateSearchParameter(sp); err == nil {
		t.Error("expected error for invalid Type")
	}
}

// ---------------------------------------------------------------------------
// boolPtr test
// ---------------------------------------------------------------------------

func TestBoolPtr(t *testing.T) {
	p := boolPtr(true)
	if p == nil || *p != true {
		t.Error("boolPtr(true) should return pointer to true")
	}
	p2 := boolPtr(false)
	if p2 == nil || *p2 != false {
		t.Error("boolPtr(false) should return pointer to false")
	}
}

// ---------------------------------------------------------------------------
// SearchParameterHandler HTTP tests
// ---------------------------------------------------------------------------

func newSearchParamTestServer() (*echo.Echo, *SearchParameterStore) {
	e := echo.New()
	store := NewSearchParameterStore()
	handler := NewSearchParameterHandler(store)
	g := e.Group("/fhir")
	handler.RegisterRoutes(g)
	return e, store
}

func TestSearchParameterHandler_Create(t *testing.T) {
	e, _ := newSearchParamTestServer()

	sp := SearchParameterResource{
		ID:     "custom-1",
		URL:    "http://example.com/fhir/SearchParameter/custom-1",
		Name:   "Custom1",
		Status: "active",
		Code:   "custom",
		Base:   []string{"Patient"},
		Type:   "string",
	}
	body, _ := json.Marshal(sp)

	req := httptest.NewRequest(http.MethodPost, "/fhir/SearchParameter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var result SearchParameterResource
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.ID != "custom-1" {
		t.Errorf("expected ID=custom-1, got %q", result.ID)
	}
	if result.ResourceType != "SearchParameter" {
		t.Errorf("expected ResourceType=SearchParameter, got %q", result.ResourceType)
	}
}

func TestSearchParameterHandler_Create_InvalidJSON(t *testing.T) {
	e, _ := newSearchParamTestServer()

	req := httptest.NewRequest(http.MethodPost, "/fhir/SearchParameter", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestSearchParameterHandler_Create_MissingRequired(t *testing.T) {
	e, _ := newSearchParamTestServer()

	sp := SearchParameterResource{ID: "x"} // missing required fields
	body, _ := json.Marshal(sp)

	req := httptest.NewRequest(http.MethodPost, "/fhir/SearchParameter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSearchParameterHandler_Create_Duplicate(t *testing.T) {
	e, store := newSearchParamTestServer()

	_ = store.Create(&SearchParameterResource{
		ID:   "dup",
		URL:  "http://example.com/fhir/SearchParameter/dup",
		Name: "Dup",
	})

	sp := SearchParameterResource{
		ID:     "dup",
		URL:    "http://example.com/fhir/SearchParameter/dup",
		Name:   "Dup",
		Status: "active",
		Code:   "dup",
		Base:   []string{"Patient"},
		Type:   "string",
	}
	body, _ := json.Marshal(sp)

	req := httptest.NewRequest(http.MethodPost, "/fhir/SearchParameter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestSearchParameterHandler_Read(t *testing.T) {
	e, store := newSearchParamTestServer()

	_ = store.Create(&SearchParameterResource{
		ID:           "r1",
		ResourceType: "SearchParameter",
		URL:          "http://example.com/fhir/SearchParameter/r1",
		Name:         "R1",
		Status:       "active",
		Code:         "r1",
		Base:         []string{"Patient"},
		Type:         "string",
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/SearchParameter/r1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result SearchParameterResource
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.Name != "R1" {
		t.Errorf("expected Name=R1, got %q", result.Name)
	}
}

func TestSearchParameterHandler_Read_NotFound(t *testing.T) {
	e, _ := newSearchParamTestServer()

	req := httptest.NewRequest(http.MethodGet, "/fhir/SearchParameter/nope", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestSearchParameterHandler_Update(t *testing.T) {
	e, store := newSearchParamTestServer()

	_ = store.Create(&SearchParameterResource{
		ID:     "u1",
		URL:    "http://example.com/fhir/SearchParameter/u1",
		Name:   "Before",
		Status: "draft",
		Code:   "u1",
		Base:   []string{"Patient"},
		Type:   "string",
	})

	updated := SearchParameterResource{
		URL:    "http://example.com/fhir/SearchParameter/u1",
		Name:   "After",
		Status: "active",
		Code:   "u1",
		Base:   []string{"Patient"},
		Type:   "token",
	}
	body, _ := json.Marshal(updated)

	req := httptest.NewRequest(http.MethodPut, "/fhir/SearchParameter/u1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result SearchParameterResource
	_ = json.Unmarshal(rec.Body.Bytes(), &result)
	if result.Name != "After" {
		t.Errorf("expected Name=After, got %q", result.Name)
	}
	if result.ID != "u1" {
		t.Errorf("expected ID=u1 (from path), got %q", result.ID)
	}
}

func TestSearchParameterHandler_Update_NotFound(t *testing.T) {
	e, _ := newSearchParamTestServer()

	sp := SearchParameterResource{
		URL:    "http://example.com/fhir/SearchParameter/nope",
		Name:   "Nope",
		Status: "active",
		Code:   "nope",
		Base:   []string{"Patient"},
		Type:   "string",
	}
	body, _ := json.Marshal(sp)

	req := httptest.NewRequest(http.MethodPut, "/fhir/SearchParameter/nope", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestSearchParameterHandler_Update_InvalidJSON(t *testing.T) {
	e, store := newSearchParamTestServer()

	_ = store.Create(&SearchParameterResource{
		ID:   "uj",
		URL:  "http://example.com/fhir/SearchParameter/uj",
		Name: "UJ",
	})

	req := httptest.NewRequest(http.MethodPut, "/fhir/SearchParameter/uj", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestSearchParameterHandler_Delete(t *testing.T) {
	e, store := newSearchParamTestServer()

	_ = store.Create(&SearchParameterResource{
		ID:   "d1",
		URL:  "http://example.com/fhir/SearchParameter/d1",
		Name: "D1",
	})

	req := httptest.NewRequest(http.MethodDelete, "/fhir/SearchParameter/d1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	// Verify deletion.
	req2 := httptest.NewRequest(http.MethodGet, "/fhir/SearchParameter/d1", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rec2.Code)
	}
}

func TestSearchParameterHandler_Delete_NotFound(t *testing.T) {
	e, _ := newSearchParamTestServer()

	req := httptest.NewRequest(http.MethodDelete, "/fhir/SearchParameter/nope", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestSearchParameterHandler_Search_NoFilters(t *testing.T) {
	e, store := newSearchParamTestServer()

	_ = store.Create(&SearchParameterResource{ID: "s1", URL: "u1", Name: "A", Code: "a", Status: "active", Base: []string{"Patient"}, Type: "string"})
	_ = store.Create(&SearchParameterResource{ID: "s2", URL: "u2", Name: "B", Code: "b", Status: "active", Base: []string{"Observation"}, Type: "token"})

	req := httptest.NewRequest(http.MethodGet, "/fhir/SearchParameter", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &bundle)

	if bundle["resourceType"] != "Bundle" {
		t.Error("expected Bundle resourceType")
	}
	if bundle["type"] != "searchset" {
		t.Error("expected searchset type")
	}

	total, ok := bundle["total"].(float64)
	if !ok || int(total) != 2 {
		t.Errorf("expected total=2, got %v", bundle["total"])
	}
}

func TestSearchParameterHandler_Search_WithFilters(t *testing.T) {
	e, store := newSearchParamTestServer()

	_ = store.Create(&SearchParameterResource{ID: "f1", URL: "u1", Name: "A", Code: "a", Status: "active", Base: []string{"Patient"}, Type: "string"})
	_ = store.Create(&SearchParameterResource{ID: "f2", URL: "u2", Name: "B", Code: "b", Status: "draft", Base: []string{"Observation"}, Type: "token"})
	_ = store.Create(&SearchParameterResource{ID: "f3", URL: "u3", Name: "C", Code: "c", Status: "active", Base: []string{"Patient"}, Type: "token"})

	req := httptest.NewRequest(http.MethodGet, "/fhir/SearchParameter?status=active&base=Patient", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &bundle)

	total, _ := bundle["total"].(float64)
	if int(total) != 2 {
		t.Errorf("expected 2 matching results, got %v", bundle["total"])
	}
}

func TestSearchParameterHandler_Search_ByType(t *testing.T) {
	e, store := newSearchParamTestServer()

	_ = store.Create(&SearchParameterResource{ID: "t1", URL: "u1", Name: "A", Type: "token"})
	_ = store.Create(&SearchParameterResource{ID: "t2", URL: "u2", Name: "B", Type: "string"})
	_ = store.Create(&SearchParameterResource{ID: "t3", URL: "u3", Name: "C", Type: "token"})

	req := httptest.NewRequest(http.MethodGet, "/fhir/SearchParameter?type=token", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var bundle map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &bundle)

	total, _ := bundle["total"].(float64)
	if int(total) != 2 {
		t.Errorf("expected 2 results for type=token, got %v", bundle["total"])
	}
}

// ---------------------------------------------------------------------------
// Integration: full CRUD lifecycle via handler
// ---------------------------------------------------------------------------

func TestSearchParameterHandler_FullLifecycle(t *testing.T) {
	e, _ := newSearchParamTestServer()

	// 1. Create
	sp := SearchParameterResource{
		ID:          "lifecycle-1",
		URL:         "http://example.com/fhir/SearchParameter/lifecycle-1",
		Name:        "Lifecycle",
		Status:      "draft",
		Code:        "lifecycle",
		Base:        []string{"Patient"},
		Type:        "string",
		Description: "Test lifecycle parameter",
		Expression:  "Patient.name.text",
	}
	body, _ := json.Marshal(sp)

	createReq := httptest.NewRequest(http.MethodPost, "/fhir/SearchParameter", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d: %s", createRec.Code, createRec.Body.String())
	}

	// 2. Read
	readReq := httptest.NewRequest(http.MethodGet, "/fhir/SearchParameter/lifecycle-1", nil)
	readRec := httptest.NewRecorder()
	e.ServeHTTP(readRec, readReq)
	if readRec.Code != http.StatusOK {
		t.Fatalf("read: expected 200, got %d", readRec.Code)
	}
	var readResult SearchParameterResource
	_ = json.Unmarshal(readRec.Body.Bytes(), &readResult)
	if readResult.Status != "draft" {
		t.Errorf("read: expected status=draft, got %q", readResult.Status)
	}

	// 3. Update
	sp.Status = "active"
	sp.Description = "Updated description"
	body, _ = json.Marshal(sp)
	updateReq := httptest.NewRequest(http.MethodPut, "/fhir/SearchParameter/lifecycle-1", bytes.NewReader(body))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRec := httptest.NewRecorder()
	e.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d: %s", updateRec.Code, updateRec.Body.String())
	}

	// 4. Search (verify update)
	searchReq := httptest.NewRequest(http.MethodGet, "/fhir/SearchParameter?status=active", nil)
	searchRec := httptest.NewRecorder()
	e.ServeHTTP(searchRec, searchReq)
	var bundle map[string]interface{}
	_ = json.Unmarshal(searchRec.Body.Bytes(), &bundle)
	total, _ := bundle["total"].(float64)
	if int(total) != 1 {
		t.Errorf("search: expected 1 active param, got %v", bundle["total"])
	}

	// 5. Delete
	deleteReq := httptest.NewRequest(http.MethodDelete, "/fhir/SearchParameter/lifecycle-1", nil)
	deleteRec := httptest.NewRecorder()
	e.ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", deleteRec.Code)
	}

	// 6. Confirm deletion
	readReq2 := httptest.NewRequest(http.MethodGet, "/fhir/SearchParameter/lifecycle-1", nil)
	readRec2 := httptest.NewRecorder()
	e.ServeHTTP(readRec2, readReq2)
	if readRec2.Code != http.StatusNotFound {
		t.Fatalf("post-delete read: expected 404, got %d", readRec2.Code)
	}
}

// ---------------------------------------------------------------------------
// Default store integration test
// ---------------------------------------------------------------------------

func TestNewDefaultSearchParameterStore_SearchByBase(t *testing.T) {
	store := NewDefaultSearchParameterStore()

	// Search for Patient-specific params.
	patientParams := store.Search(map[string]string{"base": "Patient"})
	if len(patientParams) < 5 {
		t.Errorf("expected at least 5 Patient search params, got %d", len(patientParams))
	}

	// Search for cross-resource params.
	resourceParams := store.Search(map[string]string{"base": "Resource"})
	if len(resourceParams) < 5 {
		t.Errorf("expected at least 5 cross-resource params, got %d", len(resourceParams))
	}
}

func TestNewDefaultSearchParameterStore_GetById(t *testing.T) {
	store := NewDefaultSearchParameterStore()

	sp, err := store.Get("Resource-id")
	if err != nil {
		t.Fatalf("expected to find Resource-id, got error: %v", err)
	}
	if sp.Code != "_id" {
		t.Errorf("expected code=_id, got %q", sp.Code)
	}

	sp2, err := store.Get("Patient-birthdate")
	if err != nil {
		t.Fatalf("expected to find Patient-birthdate, got error: %v", err)
	}
	if sp2.Code != "birthdate" {
		t.Errorf("expected code=birthdate, got %q", sp2.Code)
	}
}
