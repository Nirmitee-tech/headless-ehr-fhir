package fhir

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

// System URI constants for code systems used in concept maps.
const (
	systemLOINC  = "http://loinc.org"
	systemICD10  = "http://hl7.org/fhir/sid/icd-10-cm"
	systemSNOMED = "http://snomed.info/sct"
)

// ConceptMapTranslator translates codes between code systems.
type ConceptMapTranslator struct {
	maps    map[string]*ConceptMap // keyed by "sourceURI|targetURI"
	byURL   map[string]*ConceptMap // keyed by ConceptMap URL
	byID    map[string]*ConceptMap // keyed by ConceptMap ID
	allMaps []*ConceptMap
}

// ConceptMap holds mappings between two code systems.
type ConceptMap struct {
	ID        string
	URL       string
	Name      string
	SourceURI string
	TargetURI string
	Mappings  map[string][]TranslationMapping // source code → target mappings
}

// TranslationMapping represents a single code-to-code mapping.
type TranslationMapping struct {
	SourceCode    string
	SourceDisplay string
	TargetCode    string
	TargetDisplay string
	Equivalence   string // "equivalent", "wider", "narrower", "inexact", "unmatched"
}

// TranslateRequest holds the parameters for a $translate call.
type TranslateRequest struct {
	Code          string
	System        string
	TargetSystem  string
	ConceptMapURL string // optional, to select a specific map
}

// TranslateResponse holds the result of a $translate call.
type TranslateResponse struct {
	Result  bool             // Whether a mapping was found
	Message string           // Human-readable message
	Matches []TranslateMatch // Translation results
}

// TranslateMatch represents one translation result.
type TranslateMatch struct {
	Equivalence string
	Code        string
	Display     string
	System      string
}

// NewConceptMapTranslator creates a translator with built-in concept maps.
func NewConceptMapTranslator() *ConceptMapTranslator {
	t := &ConceptMapTranslator{
		maps:  make(map[string]*ConceptMap),
		byURL: make(map[string]*ConceptMap),
		byID:  make(map[string]*ConceptMap),
	}
	t.loadBuiltinMaps()
	return t
}

// loadBuiltinMaps registers all built-in concept maps.
func (t *ConceptMapTranslator) loadBuiltinMaps() {
	// 1. SNOMED → ICD-10
	snomedToICD10 := &ConceptMap{
		ID:        "snomed-to-icd10",
		URL:       "http://ehr.example.org/fhir/ConceptMap/snomed-to-icd10",
		Name:      "SNOMED CT to ICD-10-CM",
		SourceURI: systemSNOMED,
		TargetURI: systemICD10,
		Mappings:  make(map[string][]TranslationMapping),
	}

	snomedICD10Pairs := []struct {
		sCode, sDisplay, tCode, tDisplay string
	}{
		{"73211009", "Diabetes mellitus", "E11.9", "Type 2 diabetes mellitus without complications"},
		{"38341003", "Hypertension", "I10", "Essential (primary) hypertension"},
		{"195967001", "Asthma", "J45.909", "Unspecified asthma, uncomplicated"},
		{"84114007", "Heart failure", "I50.9", "Heart failure, unspecified"},
		{"13645005", "COPD", "J44.1", "Chronic obstructive pulmonary disease with acute exacerbation"},
		{"49436004", "Atrial fibrillation", "I48.91", "Unspecified atrial fibrillation"},
		{"44054006", "Type 2 diabetes", "E11.9", "Type 2 diabetes mellitus without complications"},
		{"46635009", "Type 1 diabetes", "E10.9", "Type 1 diabetes mellitus without complications"},
		{"22298006", "Myocardial infarction", "I21.9", "Acute myocardial infarction, unspecified"},
		{"230690007", "Stroke", "I63.9", "Cerebral infarction, unspecified"},
		{"40055000", "Chronic kidney disease", "N18.9", "Chronic kidney disease, unspecified"},
		{"266257000", "Transient ischemic attack", "G45.9", "Transient cerebral ischemic attack, unspecified"},
		{"68496003", "Polyarthritis", "M13.0", "Polyarthritis, unspecified"},
		{"69896004", "Rheumatoid arthritis", "M06.9", "Rheumatoid arthritis, unspecified"},
		{"396275006", "Osteoarthritis", "M19.90", "Unspecified osteoarthritis, unspecified site"},
	}

	for _, p := range snomedICD10Pairs {
		snomedToICD10.Mappings[p.sCode] = append(snomedToICD10.Mappings[p.sCode], TranslationMapping{
			SourceCode:    p.sCode,
			SourceDisplay: p.sDisplay,
			TargetCode:    p.tCode,
			TargetDisplay: p.tDisplay,
			Equivalence:   "equivalent",
		})
	}

	t.registerMap(snomedToICD10)

	// 2. ICD-10 → SNOMED (reverse of above)
	icd10ToSNOMED := &ConceptMap{
		ID:        "icd10-to-snomed",
		URL:       "http://ehr.example.org/fhir/ConceptMap/icd10-to-snomed",
		Name:      "ICD-10-CM to SNOMED CT",
		SourceURI: systemICD10,
		TargetURI: systemSNOMED,
		Mappings:  make(map[string][]TranslationMapping),
	}

	for _, p := range snomedICD10Pairs {
		icd10ToSNOMED.Mappings[p.tCode] = append(icd10ToSNOMED.Mappings[p.tCode], TranslationMapping{
			SourceCode:    p.tCode,
			SourceDisplay: p.tDisplay,
			TargetCode:    p.sCode,
			TargetDisplay: p.sDisplay,
			Equivalence:   "equivalent",
		})
	}

	t.registerMap(icd10ToSNOMED)

	// 3. LOINC → SNOMED (common lab tests)
	loincToSNOMED := &ConceptMap{
		ID:        "loinc-to-snomed",
		URL:       "http://ehr.example.org/fhir/ConceptMap/loinc-to-snomed",
		Name:      "LOINC to SNOMED CT",
		SourceURI: systemLOINC,
		TargetURI: systemSNOMED,
		Mappings:  make(map[string][]TranslationMapping),
	}

	loincSNOMEDPairs := []struct {
		sCode, sDisplay, tCode, tDisplay string
	}{
		{"2339-0", "Glucose", "33747003", "Glucose measurement"},
		{"2345-7", "Glucose, serum", "33747003", "Glucose measurement"},
		{"718-7", "Hemoglobin", "259695003", "Hemoglobin measurement"},
		{"4548-4", "HbA1c", "43396009", "Hemoglobin A1c measurement"},
		{"2160-0", "Creatinine", "70901006", "Creatinine measurement"},
		{"6690-2", "WBC", "767002", "White blood cell count"},
		{"789-8", "RBC", "14089001", "Red blood cell count"},
		{"777-3", "Platelets", "61928009", "Platelet count"},
		{"2823-3", "Potassium", "59573005", "Potassium measurement"},
		{"2951-2", "Sodium", "104934005", "Sodium measurement"},
	}

	for _, p := range loincSNOMEDPairs {
		loincToSNOMED.Mappings[p.sCode] = append(loincToSNOMED.Mappings[p.sCode], TranslationMapping{
			SourceCode:    p.sCode,
			SourceDisplay: p.sDisplay,
			TargetCode:    p.tCode,
			TargetDisplay: p.tDisplay,
			Equivalence:   "equivalent",
		})
	}

	t.registerMap(loincToSNOMED)
}

// registerMap adds a ConceptMap to all internal indexes.
func (t *ConceptMapTranslator) registerMap(cm *ConceptMap) {
	key := cm.SourceURI + "|" + cm.TargetURI
	t.maps[key] = cm
	t.byURL[cm.URL] = cm
	t.byID[cm.ID] = cm
	t.allMaps = append(t.allMaps, cm)
}

// Translate performs a code translation using the built-in concept maps.
func (t *ConceptMapTranslator) Translate(req *TranslateRequest) (*TranslateResponse, error) {
	var cm *ConceptMap

	if req.ConceptMapURL != "" {
		// Look up by URL.
		var ok bool
		cm, ok = t.byURL[req.ConceptMapURL]
		if !ok {
			return nil, fmt.Errorf("ConceptMap not found for URL: %s", req.ConceptMapURL)
		}
	} else {
		// Look up by source|target system pair.
		if req.System == "" {
			return nil, fmt.Errorf("system parameter is required")
		}
		if req.TargetSystem == "" {
			return nil, fmt.Errorf("targetsystem parameter is required")
		}
		key := req.System + "|" + req.TargetSystem
		var ok bool
		cm, ok = t.maps[key]
		if !ok {
			return nil, fmt.Errorf("no ConceptMap found for source system '%s' and target system '%s'", req.System, req.TargetSystem)
		}
	}

	mappings, found := cm.Mappings[req.Code]
	if !found || len(mappings) == 0 {
		return &TranslateResponse{
			Result:  false,
			Message: "No mapping found for code '" + req.Code + "' in system '" + req.System + "'",
		}, nil
	}

	matches := make([]TranslateMatch, 0, len(mappings))
	for _, m := range mappings {
		matches = append(matches, TranslateMatch{
			Equivalence: m.Equivalence,
			Code:        m.TargetCode,
			Display:     m.TargetDisplay,
			System:      cm.TargetURI,
		})
	}

	return &TranslateResponse{
		Result:  true,
		Message: "Mapping found",
		Matches: matches,
	}, nil
}

// ListConceptMaps returns summary information for all registered concept maps.
func (t *ConceptMapTranslator) ListConceptMaps() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(t.allMaps))
	for _, cm := range t.allMaps {
		result = append(result, map[string]interface{}{
			"resourceType": "ConceptMap",
			"id":           cm.ID,
			"url":          cm.URL,
			"name":         cm.Name,
			"status":       "active",
			"sourceUri":    cm.SourceURI,
			"targetUri":    cm.TargetURI,
		})
	}
	return result
}

// TranslateHandler provides the ConceptMap/$translate HTTP endpoints.
type TranslateHandler struct {
	translator *ConceptMapTranslator
}

// NewTranslateHandler creates a new TranslateHandler.
func NewTranslateHandler(translator *ConceptMapTranslator) *TranslateHandler {
	return &TranslateHandler{translator: translator}
}

// RegisterRoutes adds ConceptMap routes to the given FHIR group.
func (h *TranslateHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/ConceptMap", h.ListConceptMaps)
	g.GET("/ConceptMap/$translate", h.Translate)
	g.POST("/ConceptMap/$translate", h.TranslatePost)
	g.GET("/ConceptMap/:id/$translate", h.TranslateByMap)
}

// ListConceptMaps handles GET /fhir/ConceptMap — returns a Bundle of available maps.
func (h *TranslateHandler) ListConceptMaps(c echo.Context) error {
	maps := h.translator.ListConceptMaps()
	entries := make([]map[string]interface{}, 0, len(maps))
	for _, m := range maps {
		entries = append(entries, map[string]interface{}{
			"resource": m,
		})
	}

	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"total":        len(entries),
		"entry":        entries,
	}
	return c.JSON(http.StatusOK, bundle)
}

// Translate handles GET /fhir/ConceptMap/$translate with query parameters.
func (h *TranslateHandler) Translate(c echo.Context) error {
	code := c.QueryParam("code")
	system := c.QueryParam("system")
	targetSystem := c.QueryParam("targetsystem")
	conceptMapURL := c.QueryParam("url")

	if code == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'code' is required"))
	}
	if system == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'system' is required"))
	}
	if targetSystem == "" && conceptMapURL == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'targetsystem' or 'url' is required"))
	}

	req := &TranslateRequest{
		Code:          code,
		System:        system,
		TargetSystem:  targetSystem,
		ConceptMapURL: conceptMapURL,
	}
	return h.doTranslate(c, req)
}

// TranslatePost handles POST /fhir/ConceptMap/$translate with a Parameters resource body.
func (h *TranslateHandler) TranslatePost(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "structure", "Failed to read request body"))
	}

	var params struct {
		ResourceType string `json:"resourceType"`
		Parameter    []struct {
			Name         string `json:"name"`
			ValueCode    string `json:"valueCode,omitempty"`
			ValueUri     string `json:"valueUri,omitempty"`
			ValueString  string `json:"valueString,omitempty"`
		} `json:"parameter"`
	}

	if err := json.Unmarshal(body, &params); err != nil {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "structure", "Invalid JSON: "+err.Error()))
	}

	req := &TranslateRequest{}
	for _, p := range params.Parameter {
		switch p.Name {
		case "code":
			req.Code = p.ValueCode
		case "system":
			req.System = p.ValueUri
		case "targetsystem":
			req.TargetSystem = p.ValueUri
		case "url":
			req.ConceptMapURL = p.ValueUri
		}
	}

	if req.Code == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'code' is required"))
	}
	if req.System == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'system' is required"))
	}

	return h.doTranslate(c, req)
}

// TranslateByMap handles GET /fhir/ConceptMap/:id/$translate.
func (h *TranslateHandler) TranslateByMap(c echo.Context) error {
	mapID := c.Param("id")
	code := c.QueryParam("code")
	system := c.QueryParam("system")

	if code == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'code' is required"))
	}
	if system == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'system' is required"))
	}

	cm, ok := h.translator.byID[mapID]
	if !ok {
		return c.JSON(http.StatusNotFound, operationOutcome("error", "not-found", "ConceptMap '"+mapID+"' not found"))
	}

	req := &TranslateRequest{
		Code:          code,
		System:        system,
		ConceptMapURL: cm.URL,
	}
	return h.doTranslate(c, req)
}

// doTranslate performs the translation and returns a FHIR Parameters response.
func (h *TranslateHandler) doTranslate(c echo.Context, req *TranslateRequest) error {
	resp, err := h.translator.Translate(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "not-found", err.Error()))
	}

	return c.JSON(http.StatusOK, buildTranslateParametersResponse(resp))
}

// buildTranslateParametersResponse converts a TranslateResponse to a FHIR Parameters resource.
func buildTranslateParametersResponse(resp *TranslateResponse) map[string]interface{} {
	params := []interface{}{
		map[string]interface{}{
			"name":         "result",
			"valueBoolean": resp.Result,
		},
		map[string]interface{}{
			"name":        "message",
			"valueString": resp.Message,
		},
	}

	for _, m := range resp.Matches {
		matchParts := []interface{}{
			map[string]interface{}{
				"name":      "equivalence",
				"valueCode": m.Equivalence,
			},
			map[string]interface{}{
				"name": "concept",
				"valueCoding": map[string]interface{}{
					"system":  m.System,
					"code":    m.Code,
					"display": m.Display,
				},
			},
		}

		params = append(params, map[string]interface{}{
			"name": "match",
			"part": matchParts,
		})
	}

	return map[string]interface{}{
		"resourceType": "Parameters",
		"parameter":    params,
	}
}

// operationOutcome builds a minimal FHIR OperationOutcome.
func operationOutcome(severity, code, diagnostics string) map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue": []map[string]interface{}{
			{
				"severity":    severity,
				"code":        code,
				"diagnostics": diagnostics,
			},
		},
	}
}
