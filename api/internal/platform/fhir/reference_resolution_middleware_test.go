package fhir

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// staticResolver builds a newResolver func backed by a fixed mapping, for
// middleware tests that don't touch a database.
func staticResolver() func() *ReferenceResolver {
	data := map[string]map[string]string{
		"patient":      {uuidJohn: fhirJohn},
		"practitioner": {uuidDrS: fhirDrS},
	}
	return func() *ReferenceResolver {
		return NewReferenceResolver(func(_ context.Context, table string, ids []string) (map[string]string, error) {
			out := map[string]string{}
			for _, id := range ids {
				if v, ok := data[table][id]; ok {
					out[id] = v
				}
			}
			return out, nil
		})
	}
}

func runMW(t *testing.T, handler echo.HandlerFunc, target string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	e.Use(ReferenceResolutionMiddleware(staticResolver()))
	e.GET("/fhir/*", handler)
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestRefMiddleware_SingleResource(t *testing.T) {
	rec := runMW(t, func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"resourceType": "Condition",
			"id":           "c1",
			"subject":      map[string]interface{}{"reference": "Patient/" + uuidJohn},
		})
	}, "/fhir/Condition/c1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if got := out["subject"].(map[string]interface{})["reference"]; got != "Patient/"+fhirJohn {
		t.Errorf("reference not resolved: got %q", got)
	}
}

func TestRefMiddleware_Bundle(t *testing.T) {
	rec := runMW(t, func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"resourceType": "Bundle",
			"type":         "searchset",
			"entry": []interface{}{
				map[string]interface{}{"resource": map[string]interface{}{
					"resourceType": "Observation",
					"subject":      map[string]interface{}{"reference": "Patient/" + uuidJohn},
					"performer":    []interface{}{map[string]interface{}{"reference": "Practitioner/" + uuidDrS}},
				}},
			},
		})
	}, "/fhir/Observation?patient=patient-john-smith")

	var out map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	res := out["entry"].([]interface{})[0].(map[string]interface{})["resource"].(map[string]interface{})
	if got := res["subject"].(map[string]interface{})["reference"]; got != "Patient/"+fhirJohn {
		t.Errorf("bundle subject not resolved: got %q", got)
	}
	if got := res["performer"].([]interface{})[0].(map[string]interface{})["reference"]; got != "Practitioner/"+fhirDrS {
		t.Errorf("bundle performer not resolved: got %q", got)
	}
}

func TestRefMiddleware_PassesThroughNonFHIR(t *testing.T) {
	// An OperationOutcome-less error payload (no resourceType) must be untouched
	// and the status code preserved.
	rec := runMW(t, func(c echo.Context) error {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"message": "not found"})
	}, "/fhir/Condition/missing")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status not preserved: got %d", rec.Code)
	}
	var out map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	if out["message"] != "not found" {
		t.Errorf("non-FHIR body altered: %v", out)
	}
}

func TestRefMiddleware_PreservesErrorStatus(t *testing.T) {
	rec := runMW(t, func(c echo.Context) error {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"resourceType": "OperationOutcome",
			"issue":        []interface{}{map[string]interface{}{"severity": "error", "code": "forbidden"}},
		})
	}, "/fhir/Patient/x")
	if rec.Code != http.StatusForbidden {
		t.Errorf("status not preserved: got %d", rec.Code)
	}
}
