package fhir

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// helper to create a test registry with a Patient fetcher and Observation:subject reference.
func testRegistry() *IncludeRegistry {
	reg := NewIncludeRegistry()
	reg.RegisterFetcher("Patient", func(ctx context.Context, fhirID string) (map[string]interface{}, error) {
		return map[string]interface{}{
			"resourceType": "Patient",
			"id":           fhirID,
			"name": []interface{}{
				map[string]interface{}{
					"family": "Smith",
					"given":  []interface{}{"John"},
				},
			},
		}, nil
	})
	reg.RegisterReference("Observation", "subject", "Patient")
	return reg
}

// searchBundleJSON returns a serialized FHIR searchset bundle containing the given
// resources as matched entries.
func searchBundleJSON(resources ...map[string]interface{}) []byte {
	entries := make([]BundleEntry, len(resources))
	for i, r := range resources {
		raw, _ := json.Marshal(r)
		entries[i] = BundleEntry{
			FullURL:  "Observation/" + r["id"].(string),
			Resource: raw,
			Search:   &BundleSearch{Mode: "match"},
		}
	}
	total := len(resources)
	b := Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        &total,
		Entry:        entries,
	}
	data, _ := json.Marshal(b)
	return data
}

func TestIncludeMiddleware_Passthrough_NoParams(t *testing.T) {
	reg := testRegistry()
	mw := IncludeMiddleware(reg)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	bundleData := searchBundleJSON(map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"subject":      map[string]interface{}{"reference": "Patient/pat-1"},
	})

	handler := mw(func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, bundleData)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Without _include params, the bundle should pass through unchanged with
	// no included entries added.
	var bundle Bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(bundle.Entry) != 1 {
		t.Errorf("entries = %d, want 1 (no include entries added)", len(bundle.Entry))
	}
}

func TestIncludeMiddleware_ResolvesInclude(t *testing.T) {
	reg := testRegistry()
	mw := IncludeMiddleware(reg)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation?_include=Observation:subject", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	bundleData := searchBundleJSON(map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"subject":      map[string]interface{}{"reference": "Patient/pat-1"},
	})

	handler := mw(func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, bundleData)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var bundle Bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Should have 1 match entry + 1 include entry.
	if len(bundle.Entry) != 2 {
		t.Fatalf("entries = %d, want 2 (1 match + 1 include)", len(bundle.Entry))
	}

	// First entry should be the original match.
	if bundle.Entry[0].Search == nil || bundle.Entry[0].Search.Mode != "match" {
		t.Error("first entry should have search.mode='match'")
	}

	// Second entry should be the included Patient.
	if bundle.Entry[1].Search == nil || bundle.Entry[1].Search.Mode != "include" {
		t.Error("second entry should have search.mode='include'")
	}

	var patient map[string]interface{}
	if err := json.Unmarshal(bundle.Entry[1].Resource, &patient); err != nil {
		t.Fatalf("unmarshal included resource: %v", err)
	}
	if patient["resourceType"] != "Patient" {
		t.Errorf("included resourceType = %v, want Patient", patient["resourceType"])
	}
	if patient["id"] != "pat-1" {
		t.Errorf("included id = %v, want pat-1", patient["id"])
	}
}

func TestIncludeMiddleware_DeduplicatesIncludes(t *testing.T) {
	reg := testRegistry()
	mw := IncludeMiddleware(reg)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation?_include=Observation:subject", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Two observations referencing the same patient.
	bundleData := searchBundleJSON(
		map[string]interface{}{
			"resourceType": "Observation",
			"id":           "obs-1",
			"subject":      map[string]interface{}{"reference": "Patient/pat-1"},
		},
		map[string]interface{}{
			"resourceType": "Observation",
			"id":           "obs-2",
			"subject":      map[string]interface{}{"reference": "Patient/pat-1"},
		},
	)

	handler := mw(func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, bundleData)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	var bundle Bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// 2 matches + 1 deduplicated include.
	if len(bundle.Entry) != 3 {
		t.Errorf("entries = %d, want 3 (2 matches + 1 deduplicated include)", len(bundle.Entry))
	}
}

func TestIncludeMiddleware_UnknownIncludeParam(t *testing.T) {
	reg := testRegistry()
	mw := IncludeMiddleware(reg)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation?_include=Unknown:field", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	bundleData := searchBundleJSON(map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"subject":      map[string]interface{}{"reference": "Patient/pat-1"},
	})

	handler := mw(func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, bundleData)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	var bundle Bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Unknown include should be gracefully ignored; only the original entry remains.
	if len(bundle.Entry) != 1 {
		t.Errorf("entries = %d, want 1 (unknown include should be ignored)", len(bundle.Entry))
	}
}

func TestIncludeMiddleware_NonSearchsetPassthrough(t *testing.T) {
	reg := testRegistry()
	mw := IncludeMiddleware(reg)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation?_include=Observation:subject", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Return a transaction-response bundle, not a searchset.
	txBundle := Bundle{
		ResourceType: "Bundle",
		Type:         "transaction-response",
	}
	txData, _ := json.Marshal(txBundle)

	handler := mw(func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, txData)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	var bundle Bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if bundle.Type != "transaction-response" {
		t.Errorf("bundle type = %q, want transaction-response", bundle.Type)
	}
}

func TestIncludeMiddleware_ErrorStatusPassthrough(t *testing.T) {
	reg := testRegistry()
	mw := IncludeMiddleware(reg)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation?_include=Observation:subject", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := mw(func(c echo.Context) error {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db failure"})
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestIncludeMiddleware_NonJSONPassthrough(t *testing.T) {
	reg := testRegistry()
	mw := IncludeMiddleware(reg)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation?_include=Observation:subject", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "not json at all")
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	// Non-JSON body should pass through without error.
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "not json at all" {
		t.Errorf("body = %q, want original text", rec.Body.String())
	}
}

func TestIncludeMiddleware_EmptyBundlePassthrough(t *testing.T) {
	reg := testRegistry()
	mw := IncludeMiddleware(reg)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation?_include=Observation:subject", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Empty searchset bundle.
	total := 0
	emptyBundle := Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        &total,
		Entry:        nil,
	}
	data, _ := json.Marshal(emptyBundle)

	handler := mw(func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, data)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	var bundle Bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(bundle.Entry) != 0 {
		t.Errorf("entries = %d, want 0", len(bundle.Entry))
	}
}

func TestIncludeMiddleware_MultipleIncludeParams(t *testing.T) {
	reg := NewIncludeRegistry()
	reg.RegisterFetcher("Patient", func(ctx context.Context, fhirID string) (map[string]interface{}, error) {
		return map[string]interface{}{
			"resourceType": "Patient",
			"id":           fhirID,
		}, nil
	})
	reg.RegisterFetcher("Encounter", func(ctx context.Context, fhirID string) (map[string]interface{}, error) {
		return map[string]interface{}{
			"resourceType": "Encounter",
			"id":           fhirID,
		}, nil
	})
	reg.RegisterReference("Observation", "subject", "Patient")
	reg.RegisterReference("Observation", "encounter", "Encounter")

	mw := IncludeMiddleware(reg)

	e := echo.New()
	// Two separate _include params.
	req := httptest.NewRequest(http.MethodGet,
		"/fhir/Observation?_include=Observation:subject&_include=Observation:encounter", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	bundleData := searchBundleJSON(map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"subject":      map[string]interface{}{"reference": "Patient/pat-1"},
		"encounter":    map[string]interface{}{"reference": "Encounter/enc-1"},
	})

	handler := mw(func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, bundleData)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	var bundle Bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// 1 match + 2 includes (Patient + Encounter).
	if len(bundle.Entry) != 3 {
		t.Fatalf("entries = %d, want 3", len(bundle.Entry))
	}

	includeTypes := map[string]bool{}
	for _, entry := range bundle.Entry {
		if entry.Search != nil && entry.Search.Mode == "include" {
			var res map[string]interface{}
			json.Unmarshal(entry.Resource, &res)
			includeTypes[res["resourceType"].(string)] = true
		}
	}
	if !includeTypes["Patient"] {
		t.Error("expected Patient in include entries")
	}
	if !includeTypes["Encounter"] {
		t.Error("expected Encounter in include entries")
	}
}

func TestIncludeMiddleware_HandlerReturnsError(t *testing.T) {
	reg := testRegistry()
	mw := IncludeMiddleware(reg)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation?_include=Observation:subject", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := mw(func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	})

	err := handler(c)
	if err == nil {
		t.Fatal("expected error from handler")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("error code = %d, want %d", httpErr.Code, http.StatusUnauthorized)
	}
}
