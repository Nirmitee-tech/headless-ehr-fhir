package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== ConfidentialityLevel Tests ===========

func TestConfidentialityLevel_Ordering(t *testing.T) {
	// The hierarchy must be U < L < M < N < R < V.
	codes := []string{
		LabelUnrestricted,
		LabelLow,
		LabelModerate,
		LabelNormal,
		LabelRestricted,
		LabelVeryRestricted,
	}

	for i := 0; i < len(codes)-1; i++ {
		levelA := ConfidentialityLevel(codes[i])
		levelB := ConfidentialityLevel(codes[i+1])
		if levelA >= levelB {
			t.Errorf("expected ConfidentialityLevel(%q)=%d < ConfidentialityLevel(%q)=%d",
				codes[i], levelA, codes[i+1], levelB)
		}
	}
}

func TestConfidentialityLevel_UnknownCode(t *testing.T) {
	level := ConfidentialityLevel("UNKNOWN")
	if level != -1 {
		t.Errorf("expected -1 for unknown code, got %d", level)
	}
}

func TestConfidentialityLevel_AllKnownCodes(t *testing.T) {
	tests := []struct {
		code  string
		level int
	}{
		{LabelUnrestricted, 0},
		{LabelLow, 1},
		{LabelModerate, 2},
		{LabelNormal, 3},
		{LabelRestricted, 4},
		{LabelVeryRestricted, 5},
	}
	for _, tt := range tests {
		got := ConfidentialityLevel(tt.code)
		if got != tt.level {
			t.Errorf("ConfidentialityLevel(%q) = %d, want %d", tt.code, got, tt.level)
		}
	}
}

// =========== CanAccessResource Tests ===========

func TestCanAccessResource_NoSecurityLabels(t *testing.T) {
	sc := &SecurityContext{MaxConfidentiality: LabelNormal}
	meta := map[string]interface{}{}

	if !CanAccessResource(sc, meta) {
		t.Error("should allow access when resource has no security labels")
	}
}

func TestCanAccessResource_NilMeta(t *testing.T) {
	sc := &SecurityContext{MaxConfidentiality: LabelNormal}

	if !CanAccessResource(sc, nil) {
		t.Error("should allow access when meta is nil")
	}
}

func TestCanAccessResource_ConfidentialityAllowed(t *testing.T) {
	sc := &SecurityContext{MaxConfidentiality: LabelRestricted}
	meta := map[string]interface{}{
		"security": []interface{}{
			map[string]interface{}{
				"system": SecurityLabelSystem,
				"code":   LabelNormal,
			},
		},
	}

	if !CanAccessResource(sc, meta) {
		t.Error("user with R clearance should access N resource")
	}
}

func TestCanAccessResource_ConfidentialityExact(t *testing.T) {
	sc := &SecurityContext{MaxConfidentiality: LabelNormal}
	meta := map[string]interface{}{
		"security": []interface{}{
			map[string]interface{}{
				"system": SecurityLabelSystem,
				"code":   LabelNormal,
			},
		},
	}

	if !CanAccessResource(sc, meta) {
		t.Error("user with N clearance should access N resource")
	}
}

func TestCanAccessResource_ConfidentialityDenied(t *testing.T) {
	sc := &SecurityContext{MaxConfidentiality: LabelNormal}
	meta := map[string]interface{}{
		"security": []interface{}{
			map[string]interface{}{
				"system": SecurityLabelSystem,
				"code":   LabelRestricted,
			},
		},
	}

	if CanAccessResource(sc, meta) {
		t.Error("user with N clearance should NOT access R resource")
	}
}

func TestCanAccessResource_ConfidentialityVeryRestrictedDenied(t *testing.T) {
	sc := &SecurityContext{MaxConfidentiality: LabelRestricted}
	meta := map[string]interface{}{
		"security": []interface{}{
			map[string]interface{}{
				"system": SecurityLabelSystem,
				"code":   LabelVeryRestricted,
			},
		},
	}

	if CanAccessResource(sc, meta) {
		t.Error("user with R clearance should NOT access V resource")
	}
}

func TestCanAccessResource_SensitivityAllowed(t *testing.T) {
	sc := &SecurityContext{
		MaxConfidentiality: LabelNormal,
		AllowedLabels:      []string{LabelHIV, LabelPSY},
	}
	meta := map[string]interface{}{
		"security": []interface{}{
			map[string]interface{}{
				"system": ActCodeSystem,
				"code":   LabelHIV,
			},
		},
	}

	if !CanAccessResource(sc, meta) {
		t.Error("user with HIV label should access HIV-labeled resource")
	}
}

func TestCanAccessResource_SensitivityDenied(t *testing.T) {
	sc := &SecurityContext{
		MaxConfidentiality: LabelNormal,
		AllowedLabels:      []string{LabelPSY},
	}
	meta := map[string]interface{}{
		"security": []interface{}{
			map[string]interface{}{
				"system": ActCodeSystem,
				"code":   LabelHIV,
			},
		},
	}

	if CanAccessResource(sc, meta) {
		t.Error("user without HIV label should NOT access HIV-labeled resource")
	}
}

func TestCanAccessResource_SensitivityNoLabelsAllowed(t *testing.T) {
	sc := &SecurityContext{
		MaxConfidentiality: LabelVeryRestricted,
		AllowedLabels:      nil,
	}
	meta := map[string]interface{}{
		"security": []interface{}{
			map[string]interface{}{
				"system": ActCodeSystem,
				"code":   LabelSDV,
			},
		},
	}

	if CanAccessResource(sc, meta) {
		t.Error("user with no allowed sensitivity labels should NOT access SDV resource")
	}
}

func TestCanAccessResource_MultipleLabelsMixed(t *testing.T) {
	sc := &SecurityContext{
		MaxConfidentiality: LabelRestricted,
		AllowedLabels:      []string{LabelHIV},
	}
	meta := map[string]interface{}{
		"security": []interface{}{
			map[string]interface{}{
				"system": SecurityLabelSystem,
				"code":   LabelRestricted,
			},
			map[string]interface{}{
				"system": ActCodeSystem,
				"code":   LabelHIV,
			},
		},
	}

	if !CanAccessResource(sc, meta) {
		t.Error("user with R clearance and HIV label should access R+HIV resource")
	}
}

func TestCanAccessResource_MultipleLabelsDeniedByConfidentiality(t *testing.T) {
	sc := &SecurityContext{
		MaxConfidentiality: LabelNormal,
		AllowedLabels:      []string{LabelHIV},
	}
	meta := map[string]interface{}{
		"security": []interface{}{
			map[string]interface{}{
				"system": SecurityLabelSystem,
				"code":   LabelVeryRestricted,
			},
			map[string]interface{}{
				"system": ActCodeSystem,
				"code":   LabelHIV,
			},
		},
	}

	if CanAccessResource(sc, meta) {
		t.Error("user with N clearance should NOT access V+HIV resource even with HIV label")
	}
}

func TestCanAccessResource_MultipleLabelsDeniedBySensitivity(t *testing.T) {
	sc := &SecurityContext{
		MaxConfidentiality: LabelVeryRestricted,
		AllowedLabels:      []string{LabelHIV},
	}
	meta := map[string]interface{}{
		"security": []interface{}{
			map[string]interface{}{
				"system": SecurityLabelSystem,
				"code":   LabelRestricted,
			},
			map[string]interface{}{
				"system": ActCodeSystem,
				"code":   LabelPSY,
			},
		},
	}

	if CanAccessResource(sc, meta) {
		t.Error("user without PSY label should NOT access R+PSY resource")
	}
}

// =========== BreakGlass Tests ===========

func TestCanAccessResource_BreakGlass(t *testing.T) {
	sc := &SecurityContext{
		MaxConfidentiality: LabelUnrestricted,
		AllowedLabels:      nil,
		BreakGlass:         true,
	}
	meta := map[string]interface{}{
		"security": []interface{}{
			map[string]interface{}{
				"system": SecurityLabelSystem,
				"code":   LabelVeryRestricted,
			},
			map[string]interface{}{
				"system": ActCodeSystem,
				"code":   LabelHIV,
			},
		},
	}

	if !CanAccessResource(sc, meta) {
		t.Error("break-glass should always grant access regardless of labels")
	}
}

func TestCanAccessResource_BreakGlassEmptyMeta(t *testing.T) {
	sc := &SecurityContext{
		BreakGlass: true,
	}

	if !CanAccessResource(sc, nil) {
		t.Error("break-glass should grant access even with nil meta")
	}
}

// =========== SecurityContextFromRequest Tests ===========

func TestSecurityContextFromRequest_Defaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	sc := SecurityContextFromRequest(req)

	if sc.MaxConfidentiality != LabelNormal {
		t.Errorf("default MaxConfidentiality should be N, got %q", sc.MaxConfidentiality)
	}
	if sc.BreakGlass {
		t.Error("default BreakGlass should be false")
	}
	if len(sc.AllowedLabels) != 0 {
		t.Errorf("default AllowedLabels should be empty, got %v", sc.AllowedLabels)
	}
	if sc.Purpose != "" {
		t.Errorf("default Purpose should be empty, got %q", sc.Purpose)
	}
}

func TestSecurityContextFromRequest_AllHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Security-Max-Confidentiality", "R")
	req.Header.Set("X-Security-Labels", "HIV, PSY, ETH")
	req.Header.Set("X-Break-Glass", "true")
	req.Header.Set("X-Purpose-Of-Use", "TREAT")

	sc := SecurityContextFromRequest(req)

	if sc.MaxConfidentiality != "R" {
		t.Errorf("MaxConfidentiality should be R, got %q", sc.MaxConfidentiality)
	}
	if !sc.BreakGlass {
		t.Error("BreakGlass should be true")
	}
	if sc.Purpose != "TREAT" {
		t.Errorf("Purpose should be TREAT, got %q", sc.Purpose)
	}

	expectedLabels := []string{"HIV", "PSY", "ETH"}
	if len(sc.AllowedLabels) != len(expectedLabels) {
		t.Fatalf("expected %d labels, got %d", len(expectedLabels), len(sc.AllowedLabels))
	}
	for i, label := range sc.AllowedLabels {
		if label != expectedLabels[i] {
			t.Errorf("AllowedLabels[%d] = %q, want %q", i, label, expectedLabels[i])
		}
	}
}

func TestSecurityContextFromRequest_BreakGlassCaseInsensitive(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Break-Glass", "TRUE")
	sc := SecurityContextFromRequest(req)
	if !sc.BreakGlass {
		t.Error("BreakGlass should accept case-insensitive 'TRUE'")
	}
}

// =========== SecurityLabelMiddleware Tests ===========

func securityLabelTestHandler(resource map[string]interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, resource)
	}
}

func TestSecurityLabelMiddleware_AllowsSingleResource(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"meta": map[string]interface{}{
			"security": []interface{}{
				map[string]interface{}{
					"system": SecurityLabelSystem,
					"code":   LabelNormal,
				},
			},
		},
	}

	e := echo.New()
	e.Use(SecurityLabelMiddleware())
	e.GET("/fhir/Patient/:id", securityLabelTestHandler(resource))

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	req.Header.Set("X-Security-Max-Confidentiality", "R")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["id"] != "123" {
		t.Error("resource should be returned unchanged")
	}
}

func TestSecurityLabelMiddleware_DeniedSingleResource(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "secret-456",
		"meta": map[string]interface{}{
			"security": []interface{}{
				map[string]interface{}{
					"system": SecurityLabelSystem,
					"code":   LabelVeryRestricted,
				},
			},
		},
	}

	e := echo.New()
	e.Use(SecurityLabelMiddleware())
	e.GET("/fhir/Patient/:id", securityLabelTestHandler(resource))

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/secret-456", nil)
	req.Header.Set("X-Security-Max-Confidentiality", "N")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", result["resourceType"])
	}
}

func TestSecurityLabelMiddleware_FiltersBundleEntries(t *testing.T) {
	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"total":        float64(3),
		"entry": []interface{}{
			map[string]interface{}{
				"resource": map[string]interface{}{
					"resourceType": "Patient",
					"id":           "p1",
					"meta": map[string]interface{}{
						"security": []interface{}{
							map[string]interface{}{
								"system": SecurityLabelSystem,
								"code":   LabelNormal,
							},
						},
					},
				},
			},
			map[string]interface{}{
				"resource": map[string]interface{}{
					"resourceType": "Patient",
					"id":           "p2",
					"meta": map[string]interface{}{
						"security": []interface{}{
							map[string]interface{}{
								"system": SecurityLabelSystem,
								"code":   LabelVeryRestricted,
							},
						},
					},
				},
			},
			map[string]interface{}{
				"resource": map[string]interface{}{
					"resourceType": "Patient",
					"id":           "p3",
					"meta": map[string]interface{}{
						"security": []interface{}{
							map[string]interface{}{
								"system": SecurityLabelSystem,
								"code":   LabelLow,
							},
						},
					},
				},
			},
		},
	}

	e := echo.New()
	e.Use(SecurityLabelMiddleware())
	e.GET("/fhir/Patient", securityLabelTestHandler(bundle))

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("X-Security-Max-Confidentiality", "N")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	entries, ok := result["entry"].([]interface{})
	if !ok {
		t.Fatal("expected entry array in result")
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries after filtering, got %d", len(entries))
	}

	// Verify correct resources remain (p1=N and p3=L, but not p2=V).
	ids := make(map[string]bool)
	for _, entry := range entries {
		e := entry.(map[string]interface{})
		r := e["resource"].(map[string]interface{})
		ids[r["id"].(string)] = true
	}

	if !ids["p1"] {
		t.Error("p1 (Normal) should be in filtered results")
	}
	if ids["p2"] {
		t.Error("p2 (VeryRestricted) should NOT be in filtered results")
	}
	if !ids["p3"] {
		t.Error("p3 (Low) should be in filtered results")
	}

	// Check total was updated.
	total, ok := result["total"].(float64)
	if !ok || total != 2 {
		t.Errorf("expected total=2, got %v", result["total"])
	}
}

func TestSecurityLabelMiddleware_BreakGlassAllowsAll(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "secret-789",
		"meta": map[string]interface{}{
			"security": []interface{}{
				map[string]interface{}{
					"system": SecurityLabelSystem,
					"code":   LabelVeryRestricted,
				},
				map[string]interface{}{
					"system": ActCodeSystem,
					"code":   LabelHIV,
				},
			},
		},
	}

	e := echo.New()
	e.Use(SecurityLabelMiddleware())
	e.GET("/fhir/Patient/:id", securityLabelTestHandler(resource))

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/secret-789", nil)
	req.Header.Set("X-Security-Max-Confidentiality", "U")
	req.Header.Set("X-Break-Glass", "true")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 with break-glass, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["id"] != "secret-789" {
		t.Error("break-glass should return the resource")
	}
}

func TestSecurityLabelMiddleware_BreakGlassBundleNoFiltering(t *testing.T) {
	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"total":        float64(2),
		"entry": []interface{}{
			map[string]interface{}{
				"resource": map[string]interface{}{
					"resourceType": "Patient",
					"id":           "p1",
					"meta": map[string]interface{}{
						"security": []interface{}{
							map[string]interface{}{
								"system": SecurityLabelSystem,
								"code":   LabelVeryRestricted,
							},
						},
					},
				},
			},
			map[string]interface{}{
				"resource": map[string]interface{}{
					"resourceType": "Patient",
					"id":           "p2",
					"meta": map[string]interface{}{
						"security": []interface{}{
							map[string]interface{}{
								"system": ActCodeSystem,
								"code":   LabelHIV,
							},
						},
					},
				},
			},
		},
	}

	e := echo.New()
	e.Use(SecurityLabelMiddleware())
	e.GET("/fhir/Patient", securityLabelTestHandler(bundle))

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("X-Security-Max-Confidentiality", "U")
	req.Header.Set("X-Break-Glass", "true")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	entries := result["entry"].([]interface{})
	if len(entries) != 2 {
		t.Errorf("break-glass should not filter: expected 2 entries, got %d", len(entries))
	}
}

func TestSecurityLabelMiddleware_NoMetaPassesThrough(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "no-meta",
	}

	e := echo.New()
	e.Use(SecurityLabelMiddleware())
	e.GET("/fhir/Patient/:id", securityLabelTestHandler(resource))

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/no-meta", nil)
	req.Header.Set("X-Security-Max-Confidentiality", "L")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestSecurityLabelMiddleware_NonJSONPassesThrough(t *testing.T) {
	e := echo.New()
	e.Use(SecurityLabelMiddleware())
	e.GET("/fhir/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "plain text")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "plain text" {
		t.Errorf("expected original body, got %q", rec.Body.String())
	}
}

func TestSecurityLabelMiddleware_OperationOutcomePassesThrough(t *testing.T) {
	outcome := map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue": []interface{}{
			map[string]interface{}{
				"severity":    "error",
				"code":        "not-found",
				"diagnostics": "Patient/missing not found",
			},
		},
	}

	e := echo.New()
	e.Use(SecurityLabelMiddleware())
	e.GET("/fhir/Patient/:id", func(c echo.Context) error {
		return c.JSON(http.StatusNotFound, outcome)
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/missing", nil)
	req.Header.Set("X-Security-Max-Confidentiality", "U")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["resourceType"] != "OperationOutcome" {
		t.Error("OperationOutcome should pass through without security filtering")
	}
}

func TestSecurityLabelMiddleware_HandlerError(t *testing.T) {
	e := echo.New()
	e.Use(SecurityLabelMiddleware())
	e.GET("/fhir/error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/error", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestSecurityLabelMiddleware_SensitivityDeniedSingleResource(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-hiv-1",
		"meta": map[string]interface{}{
			"security": []interface{}{
				map[string]interface{}{
					"system": ActCodeSystem,
					"code":   LabelHIV,
				},
			},
		},
	}

	e := echo.New()
	e.Use(SecurityLabelMiddleware())
	e.GET("/fhir/Observation/:id", securityLabelTestHandler(resource))

	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/obs-hiv-1", nil)
	req.Header.Set("X-Security-Max-Confidentiality", "V")
	// No X-Security-Labels header, so no sensitivity labels allowed.
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for HIV resource without HIV label, got %d", rec.Code)
	}
}

func TestSecurityLabelMiddleware_SensitivityAllowedSingleResource(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-hiv-2",
		"meta": map[string]interface{}{
			"security": []interface{}{
				map[string]interface{}{
					"system": ActCodeSystem,
					"code":   LabelHIV,
				},
			},
		},
	}

	e := echo.New()
	e.Use(SecurityLabelMiddleware())
	e.GET("/fhir/Observation/:id", securityLabelTestHandler(resource))

	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/obs-hiv-2", nil)
	req.Header.Set("X-Security-Max-Confidentiality", "V")
	req.Header.Set("X-Security-Labels", "HIV,PSY")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 with HIV label allowed, got %d", rec.Code)
	}
}

// =========== isSensitivityLabel Tests ===========

func TestIsSensitivityLabel(t *testing.T) {
	sensitive := []string{LabelHIV, LabelPSY, LabelSDV, LabelETH, LabelSTD}
	for _, code := range sensitive {
		if !isSensitivityLabel(code) {
			t.Errorf("isSensitivityLabel(%q) should be true", code)
		}
	}

	notSensitive := []string{LabelNormal, LabelRestricted, "UNKNOWN", "NORELINK", ""}
	for _, code := range notSensitive {
		if isSensitivityLabel(code) {
			t.Errorf("isSensitivityLabel(%q) should be false", code)
		}
	}
}

// =========== extractSecurityCodings Tests ===========

func TestExtractSecurityCodings_EmptyMeta(t *testing.T) {
	codings := extractSecurityCodings(map[string]interface{}{})
	if len(codings) != 0 {
		t.Errorf("expected 0 codings, got %d", len(codings))
	}
}

func TestExtractSecurityCodings_ValidCodings(t *testing.T) {
	meta := map[string]interface{}{
		"security": []interface{}{
			map[string]interface{}{
				"system": SecurityLabelSystem,
				"code":   "R",
			},
			map[string]interface{}{
				"system": ActCodeSystem,
				"code":   "HIV",
			},
		},
	}
	codings := extractSecurityCodings(meta)
	if len(codings) != 2 {
		t.Fatalf("expected 2 codings, got %d", len(codings))
	}
	if codings[0].system != SecurityLabelSystem || codings[0].code != "R" {
		t.Errorf("unexpected first coding: %+v", codings[0])
	}
	if codings[1].system != ActCodeSystem || codings[1].code != "HIV" {
		t.Errorf("unexpected second coding: %+v", codings[1])
	}
}

func TestExtractSecurityCodings_SkipsEmptyCode(t *testing.T) {
	meta := map[string]interface{}{
		"security": []interface{}{
			map[string]interface{}{
				"system": SecurityLabelSystem,
				"code":   "",
			},
			map[string]interface{}{
				"system": ActCodeSystem,
				"code":   "HIV",
			},
		},
	}
	codings := extractSecurityCodings(meta)
	if len(codings) != 1 {
		t.Fatalf("expected 1 coding (empty code skipped), got %d", len(codings))
	}
}
