package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ============================================================================
// CQL Library & Expression Types
// ============================================================================

// CQLLibrary represents a parsed CQL library with named expressions.
type CQLLibrary struct {
	Name        string
	Version     string
	URL         string
	Status      string // draft|active|retired
	Parameters  []CQLParameter
	Definitions map[string]CQLDefinition // named expressions
	CreatedAt   time.Time
}

// CQLParameter defines an input parameter for a CQL library.
type CQLParameter struct {
	Name         string
	Type         string // Patient, Period, String, Integer, etc.
	DefaultValue interface{}
}

// CQLDefinition is a named CQL expression.
type CQLDefinition struct {
	Name       string
	Expression string // FHIRPath-compatible expression
	Context    string // "Patient" or "Unfiltered"
	Type       string // result type hint
}

// ============================================================================
// CQL Engine
// ============================================================================

// CQLEngine evaluates Clinical Quality Language expressions and measures.
// It builds on the existing FHIRPath engine and adds CQL-specific functions
// such as AgeInYears(), HasConditionCode(), and resource-collection
// evaluation against a patient context with associated clinical data.
type CQLEngine struct {
	fhirpath *FHIRPathEngine
}

// NewCQLEngine creates a new CQL evaluation engine backed by a FHIRPath engine.
func NewCQLEngine() *CQLEngine {
	return &CQLEngine{
		fhirpath: NewFHIRPathEngine(),
	}
}

// EvaluateExpression evaluates a single CQL/FHIRPath expression against
// patient data.  The resources parameter is a map of resource type to list of
// resources (e.g. {"Condition": [...], "Observation": [...]}).
func (e *CQLEngine) EvaluateExpression(
	ctx context.Context,
	expression string,
	patient map[string]interface{},
	resources map[string][]map[string]interface{},
) (interface{}, error) {
	if patient == nil {
		return nil, fmt.Errorf("cql: patient is nil")
	}
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil, fmt.Errorf("cql: empty expression")
	}

	// Handle CQL built-in functions that are not part of FHIRPath.
	if fn, args, ok := parseCQLFunction(expression); ok {
		return e.evalCQLFunction(fn, args, patient, resources)
	}

	// Determine if the expression references a resource type collection.
	resType := expressionResourceType(expression)
	if resType != "" && resType != "Patient" {
		return e.evalResourceExpression(resType, expression, patient, resources)
	}

	// Delegate to the FHIRPath engine against the Patient resource.
	result, err := e.fhirpath.Evaluate(patient, expression)
	if err != nil {
		return nil, fmt.Errorf("cql: %w", err)
	}
	return singletonOrCollection(result), nil
}

// EvaluateLibrary evaluates all definitions in a CQL library for a patient.
func (e *CQLEngine) EvaluateLibrary(
	ctx context.Context,
	library *CQLLibrary,
	patient map[string]interface{},
	resources map[string][]map[string]interface{},
) (map[string]interface{}, error) {
	if library == nil {
		return nil, fmt.Errorf("cql: library is nil")
	}

	results := make(map[string]interface{}, len(library.Definitions))
	for name, def := range library.Definitions {
		val, err := e.EvaluateExpression(ctx, def.Expression, patient, resources)
		if err != nil {
			// Store nil for failed evaluations rather than aborting.
			results[name] = nil
			continue
		}
		results[name] = val
	}
	return results, nil
}

// ---------------------------------------------------------------------------
// CQL built-in function helpers
// ---------------------------------------------------------------------------

// parseCQLFunction detects CQL-specific function calls not handled by FHIRPath
// and extracts the function name and arguments.
func parseCQLFunction(expr string) (string, []string, bool) {
	expr = strings.TrimSpace(expr)

	cqlFunctions := []string{
		"AgeInYears()",
		"AgeInYearsAt(",
		"HasConditionCode(",
		"HasObservationCode(",
		"GetObservationValue(",
		"HasEncounterType(",
		"GetBPComponent(",
		"HasDiagnosticReportCode(",
	}

	for _, fn := range cqlFunctions {
		if strings.HasPrefix(expr, fn) {
			if fn == "AgeInYears()" && expr == "AgeInYears()" {
				return "AgeInYears", nil, true
			}
			// Extract arguments between parentheses.
			start := strings.Index(expr, "(")
			end := strings.LastIndex(expr, ")")
			if start >= 0 && end > start {
				argStr := expr[start+1 : end]
				args := splitCQLArgs(argStr)
				name := expr[:start]
				return name, args, true
			}
		}
	}
	return "", nil, false
}

// splitCQLArgs splits a comma-separated argument string, stripping quotes.
func splitCQLArgs(s string) []string {
	parts := strings.Split(s, ",")
	var args []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, "'\"")
		if p != "" {
			args = append(args, p)
		}
	}
	return args
}

func (e *CQLEngine) evalCQLFunction(
	fn string,
	args []string,
	patient map[string]interface{},
	resources map[string][]map[string]interface{},
) (interface{}, error) {
	switch fn {
	case "AgeInYears":
		return cqlAgeInYears(patient, time.Now())
	case "AgeInYearsAt":
		if len(args) < 1 {
			return nil, fmt.Errorf("cql: AgeInYearsAt requires a date argument")
		}
		t, err := parseFlexDate(args[0])
		if err != nil {
			return nil, fmt.Errorf("cql: AgeInYearsAt: %w", err)
		}
		return cqlAgeInYears(patient, t)
	case "HasConditionCode":
		if len(args) < 1 {
			return nil, fmt.Errorf("cql: HasConditionCode requires a code prefix")
		}
		return hasResourceCode(resources, "Condition", args[0]), nil
	case "HasObservationCode":
		if len(args) < 1 {
			return nil, fmt.Errorf("cql: HasObservationCode requires a code")
		}
		return hasResourceCode(resources, "Observation", args[0]), nil
	case "HasEncounterType":
		if len(args) < 1 {
			return nil, fmt.Errorf("cql: HasEncounterType requires a code")
		}
		return hasEncounterType(resources, args[0]), nil
	case "HasDiagnosticReportCode":
		if len(args) < 1 {
			return nil, fmt.Errorf("cql: HasDiagnosticReportCode requires a code")
		}
		return hasResourceCode(resources, "DiagnosticReport", args[0]), nil
	case "GetObservationValue":
		if len(args) < 1 {
			return nil, fmt.Errorf("cql: GetObservationValue requires a LOINC code")
		}
		return getObservationValue(resources, args[0])
	case "GetBPComponent":
		if len(args) < 1 {
			return nil, fmt.Errorf("cql: GetBPComponent requires a component code")
		}
		return getBPComponent(resources, args[0])
	default:
		return nil, fmt.Errorf("cql: unknown function %q", fn)
	}
}

// cqlAgeInYears calculates the patient's age in complete years at the given date.
func cqlAgeInYears(patient map[string]interface{}, at time.Time) (int, error) {
	bd, ok := patient["birthDate"].(string)
	if !ok || bd == "" {
		return 0, fmt.Errorf("cql: patient has no birthDate")
	}
	birth, err := parseFlexDate(bd)
	if err != nil {
		return 0, fmt.Errorf("cql: invalid birthDate %q: %w", bd, err)
	}
	age := at.Year() - birth.Year()
	if at.YearDay() < birth.YearDay() {
		age--
	}
	return age, nil
}

// hasResourceCode checks whether any resource of the given type has a code
// whose coding contains a code starting with the given prefix.
func hasResourceCode(resources map[string][]map[string]interface{}, resType, codePrefix string) bool {
	for _, res := range resources[resType] {
		if resourceMatchesCode(res, codePrefix) {
			return true
		}
	}
	return false
}

// resourceMatchesCode checks if a resource's code element contains a coding
// with a code that starts with the given prefix.
func resourceMatchesCode(res map[string]interface{}, codePrefix string) bool {
	codeObj, _ := res["code"].(map[string]interface{})
	if codeObj == nil {
		return false
	}
	codings, _ := codeObj["coding"].([]interface{})
	for _, c := range codings {
		coding, _ := c.(map[string]interface{})
		if coding == nil {
			continue
		}
		code, _ := coding["code"].(string)
		if strings.HasPrefix(code, codePrefix) {
			return true
		}
	}
	return false
}

// hasEncounterType checks if any encounter has a type or class matching the code.
func hasEncounterType(resources map[string][]map[string]interface{}, code string) bool {
	for _, enc := range resources["Encounter"] {
		// Check class
		if classObj, ok := enc["class"].(map[string]interface{}); ok {
			if classCode, _ := classObj["code"].(string); classCode == code {
				return true
			}
		}
		// Check type array
		types, _ := enc["type"].([]interface{})
		for _, t := range types {
			tMap, _ := t.(map[string]interface{})
			if tMap == nil {
				continue
			}
			codings, _ := tMap["coding"].([]interface{})
			for _, c := range codings {
				coding, _ := c.(map[string]interface{})
				if coding == nil {
					continue
				}
				codeVal, _ := coding["code"].(string)
				if codeVal == code {
					return true
				}
			}
		}
	}
	return false
}

// getObservationValue finds the first Observation matching the LOINC code and
// returns its valueQuantity.value.
func getObservationValue(resources map[string][]map[string]interface{}, loincCode string) (interface{}, error) {
	for _, obs := range resources["Observation"] {
		if resourceMatchesCode(obs, loincCode) {
			vq, _ := obs["valueQuantity"].(map[string]interface{})
			if vq != nil {
				return vq["value"], nil
			}
		}
	}
	return nil, nil
}

// getBPComponent finds a BP observation (85354-9) and extracts a component
// value by its LOINC code.
func getBPComponent(resources map[string][]map[string]interface{}, componentCode string) (interface{}, error) {
	for _, obs := range resources["Observation"] {
		if !resourceMatchesCode(obs, "85354-9") {
			continue
		}
		comps, _ := obs["component"].([]interface{})
		for _, comp := range comps {
			compMap, _ := comp.(map[string]interface{})
			if compMap == nil {
				continue
			}
			codeObj, _ := compMap["code"].(map[string]interface{})
			if codeObj == nil {
				continue
			}
			codings, _ := codeObj["coding"].([]interface{})
			for _, c := range codings {
				coding, _ := c.(map[string]interface{})
				if coding == nil {
					continue
				}
				cc, _ := coding["code"].(string)
				if cc == componentCode {
					vq, _ := compMap["valueQuantity"].(map[string]interface{})
					if vq != nil {
						return vq["value"], nil
					}
				}
			}
		}
	}
	return nil, nil
}

// expressionResourceType checks if an expression starts with a FHIR resource
// type name followed by a dot or function call.
func expressionResourceType(expr string) string {
	knownTypes := []string{
		"Condition", "Observation", "Encounter", "Procedure",
		"MedicationRequest", "DiagnosticReport", "Immunization",
		"AllergyIntolerance", "CarePlan", "Patient",
	}
	for _, rt := range knownTypes {
		if strings.HasPrefix(expr, rt+".") || strings.HasPrefix(expr, rt+" ") || expr == rt {
			return rt
		}
	}
	return ""
}

// evalResourceExpression evaluates an expression against a collection of
// resources of a given type.
func (e *CQLEngine) evalResourceExpression(
	resType string,
	expression string,
	patient map[string]interface{},
	resources map[string][]map[string]interface{},
) (interface{}, error) {
	resList := resources[resType]
	collection := make([]interface{}, len(resList))
	for i, r := range resList {
		collection[i] = r
	}

	// Strip the resource type prefix to get the remaining expression.
	// e.g. "Condition.exists()" -> ".exists()" -> "exists()"
	// e.g. "Observation.count()" -> ".count()" -> "count()"
	suffix := strings.TrimPrefix(expression, resType)
	suffix = strings.TrimPrefix(suffix, ".")

	// Handle collection-level functions directly.
	switch {
	case suffix == "exists()":
		return len(collection) > 0, nil
	case suffix == "count()":
		return len(collection), nil
	case suffix == "first()":
		if len(collection) > 0 {
			return collection[0], nil
		}
		return nil, nil
	case suffix == "last()":
		if len(collection) > 0 {
			return collection[len(collection)-1], nil
		}
		return nil, nil
	case suffix == "" || suffix == resType:
		return singletonOrCollection(collection), nil
	default:
		// For more complex expressions, evaluate against each resource.
		// Try to evaluate on each resource individually and collect results.
		var results []interface{}
		for _, res := range resList {
			// Set resourceType so FHIRPath recognizes it properly.
			resMap := res
			if _, ok := resMap["resourceType"]; !ok {
				resMap["resourceType"] = resType
			}
			result, err := e.fhirpath.Evaluate(resMap, resType+"."+suffix)
			if err != nil {
				// Try without the type prefix.
				result, err = e.fhirpath.Evaluate(resMap, suffix)
				if err != nil {
					continue
				}
			}
			results = append(results, result...)
		}
		return singletonOrCollection(results), nil
	}
}

// singletonOrCollection returns the single value if the collection has one
// element, the full collection for multiple, or nil for empty.
func singletonOrCollection(coll []interface{}) interface{} {
	switch len(coll) {
	case 0:
		return nil
	case 1:
		return coll[0]
	default:
		return coll
	}
}

// ============================================================================
// FHIR Measure Resource
// ============================================================================

// Measure represents a FHIR Measure resource for quality reporting.
type Measure struct {
	ID               string
	URL              string
	Name             string
	Title            string
	Status           string // draft|active|retired
	Description      string
	Date             time.Time
	Library          []string // CQL library canonical URLs
	Scoring          string   // proportion|ratio|continuous-variable|cohort
	Type             []string // process|outcome|structure|patient-reported-outcome
	Group            []MeasureGroup
	SupplementalData []MeasureSupplementalData
}

// MeasureGroup represents a group within a measure (a population set).
type MeasureGroup struct {
	ID          string
	Code        string
	Description string
	Population  []MeasurePopulation
	Stratifier  []MeasureStratifier
}

// MeasurePopulation defines a population criteria.
type MeasurePopulation struct {
	Code       string // initial-population|numerator|denominator|...
	Expression string // CQL expression name (references a CQLDefinition)
}

// MeasureStratifier defines stratification criteria.
type MeasureStratifier struct {
	Code       string
	Expression string
}

// MeasureSupplementalData for SDE (supplemental data elements).
type MeasureSupplementalData struct {
	Code       string
	Expression string
}

// ToFHIR converts the Measure to a FHIR JSON map.
func (m *Measure) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Measure",
		"id":           m.ID,
		"url":          m.URL,
		"name":         m.Name,
		"title":        m.Title,
		"status":       m.Status,
		"description":  m.Description,
	}
	if !m.Date.IsZero() {
		result["date"] = m.Date.Format(time.RFC3339)
	}
	if len(m.Library) > 0 {
		result["library"] = m.Library
	}
	if m.Scoring != "" {
		result["scoring"] = map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://terminology.hl7.org/CodeSystem/measure-scoring",
					"code":   m.Scoring,
				},
			},
		}
	}
	if len(m.Type) > 0 {
		var types []interface{}
		for _, t := range m.Type {
			types = append(types, map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system": "http://terminology.hl7.org/CodeSystem/measure-type",
						"code":   t,
					},
				},
			})
		}
		result["type"] = types
	}
	if len(m.Group) > 0 {
		var groups []interface{}
		for _, g := range m.Group {
			gMap := map[string]interface{}{}
			if g.ID != "" {
				gMap["id"] = g.ID
			}
			if g.Code != "" {
				gMap["code"] = map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{"code": g.Code},
					},
				}
			}
			if g.Description != "" {
				gMap["description"] = g.Description
			}
			if len(g.Population) > 0 {
				var pops []interface{}
				for _, p := range g.Population {
					pops = append(pops, map[string]interface{}{
						"code": map[string]interface{}{
							"coding": []interface{}{
								map[string]interface{}{
									"system": "http://terminology.hl7.org/CodeSystem/measure-population",
									"code":   p.Code,
								},
							},
						},
						"criteria": map[string]interface{}{
							"language":   "text/cql-identifier",
							"expression": p.Expression,
						},
					})
				}
				gMap["population"] = pops
			}
			if len(g.Stratifier) > 0 {
				var strats []interface{}
				for _, s := range g.Stratifier {
					strats = append(strats, map[string]interface{}{
						"code": map[string]interface{}{
							"coding": []interface{}{
								map[string]interface{}{"code": s.Code},
							},
						},
						"criteria": map[string]interface{}{
							"language":   "text/cql-identifier",
							"expression": s.Expression,
						},
					})
				}
				gMap["stratifier"] = strats
			}
			groups = append(groups, gMap)
		}
		result["group"] = groups
	}
	return result
}

// isValidScoringType validates a measure scoring type.
func isValidScoringType(s string) bool {
	switch s {
	case "proportion", "ratio", "continuous-variable", "cohort":
		return true
	}
	return false
}

// isValidPopulationCode validates a population code.
func isValidPopulationCode(code string) bool {
	switch code {
	case "initial-population", "numerator", "denominator",
		"denominator-exclusion", "denominator-exception",
		"numerator-exclusion", "measure-population",
		"measure-observation":
		return true
	}
	return false
}

// ============================================================================
// MeasureReport
// ============================================================================

// MeasureReport represents the output of $evaluate-measure.
type MeasureReport struct {
	ID                string
	Status            string  // complete|pending|error
	Type              string  // individual|subject-list|summary|data-collection
	Measure           string  // Measure canonical URL
	Subject           *string // Patient reference (for individual)
	Period            MeasurePeriod
	Group             []MeasureReportGroup
	EvaluatedResource []string // references to resources used
}

// MeasurePeriod represents a time period for measure evaluation.
type MeasurePeriod struct {
	Start time.Time
	End   time.Time
}

// MeasureReportGroup holds population results for a measure group.
type MeasureReportGroup struct {
	Code         string
	Population   []MeasureReportPopulation
	MeasureScore *float64
	Stratifier   []MeasureReportStratifier
}

// MeasureReportPopulation holds population count and optional subject list.
type MeasureReportPopulation struct {
	Code           string
	Count          int
	SubjectResults []string // patient references (for subject-list/population)
}

// MeasureReportStratifier holds stratification results.
type MeasureReportStratifier struct {
	Code    string
	Stratum []MeasureReportStratum
}

// MeasureReportStratum holds a single stratum result.
type MeasureReportStratum struct {
	Value      string
	Population []MeasureReportPopulation
}

// ToFHIR converts the MeasureReport to a FHIR JSON map.
func (r *MeasureReport) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "MeasureReport",
		"id":           r.ID,
		"status":       r.Status,
		"type":         r.Type,
		"measure":      r.Measure,
		"period": map[string]interface{}{
			"start": r.Period.Start.Format("2006-01-02"),
			"end":   r.Period.End.Format("2006-01-02"),
		},
	}
	if r.Subject != nil {
		result["subject"] = map[string]interface{}{
			"reference": *r.Subject,
		}
	}
	if len(r.Group) > 0 {
		var groups []interface{}
		for _, g := range r.Group {
			gMap := map[string]interface{}{}
			if g.Code != "" {
				gMap["code"] = map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{"code": g.Code},
					},
				}
			}
			if len(g.Population) > 0 {
				var pops []interface{}
				for _, p := range g.Population {
					popMap := map[string]interface{}{
						"code": map[string]interface{}{
							"coding": []interface{}{
								map[string]interface{}{
									"system": "http://terminology.hl7.org/CodeSystem/measure-population",
									"code":   p.Code,
								},
							},
						},
						"count": p.Count,
					}
					if len(p.SubjectResults) > 0 {
						var refs []interface{}
						for _, ref := range p.SubjectResults {
							refs = append(refs, map[string]interface{}{"reference": ref})
						}
						popMap["subjectResults"] = refs
					}
					pops = append(pops, popMap)
				}
				gMap["population"] = pops
			}
			if g.MeasureScore != nil {
				gMap["measureScore"] = map[string]interface{}{
					"value": *g.MeasureScore,
				}
			}
			groups = append(groups, gMap)
		}
		result["group"] = groups
	}
	if len(r.EvaluatedResource) > 0 {
		var refs []interface{}
		for _, ref := range r.EvaluatedResource {
			refs = append(refs, map[string]interface{}{"reference": ref})
		}
		result["evaluatedResource"] = refs
	}
	return result
}

// ============================================================================
// PatientBundle
// ============================================================================

// PatientBundle is a patient + their clinical data for measure evaluation.
type PatientBundle struct {
	Patient   map[string]interface{}
	Resources map[string][]map[string]interface{}
}

// ============================================================================
// Built-in Quality Measure URLs
// ============================================================================

const (
	// CMS122URL is the canonical URL for the Diabetes HbA1c measure.
	CMS122URL = "http://hl7.org/fhir/us/cqfmeasures/Measure/CMS122"
	// CMS125URL is the canonical URL for the Breast Cancer Screening measure.
	CMS125URL = "http://hl7.org/fhir/us/cqfmeasures/Measure/CMS125"
	// CMS165URL is the canonical URL for the Controlling High BP measure.
	CMS165URL = "http://hl7.org/fhir/us/cqfmeasures/Measure/CMS165"
)

// ============================================================================
// Measure Evaluator
// ============================================================================

// MeasureEvaluator evaluates FHIR Measures against patient data.
type MeasureEvaluator struct {
	cql       *CQLEngine
	mu        sync.RWMutex
	libraries map[string]*CQLLibrary // URL -> library
	measures  map[string]*Measure    // URL -> measure
	reports   map[string]*MeasureReport
	// measureByID maps measure ID -> URL for handler lookups.
	measureByID map[string]string
}

// NewMeasureEvaluator creates a MeasureEvaluator with built-in quality measures.
func NewMeasureEvaluator() *MeasureEvaluator {
	e := &MeasureEvaluator{
		cql:         NewCQLEngine(),
		libraries:   make(map[string]*CQLLibrary),
		measures:    make(map[string]*Measure),
		reports:     make(map[string]*MeasureReport),
		measureByID: make(map[string]string),
	}
	e.registerBuiltinMeasures()
	return e
}

// RegisterLibrary registers a CQL library.
func (e *MeasureEvaluator) RegisterLibrary(lib *CQLLibrary) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.libraries[lib.URL] = lib
}

// RegisterMeasure registers a FHIR Measure.
func (e *MeasureEvaluator) RegisterMeasure(m *Measure) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.measures[m.URL] = m
	e.measureByID[m.ID] = m.URL
}

// EvaluateIndividual evaluates a measure for a single patient.
func (e *MeasureEvaluator) EvaluateIndividual(
	ctx context.Context,
	measureURL string,
	patient map[string]interface{},
	resources map[string][]map[string]interface{},
	period MeasurePeriod,
) (*MeasureReport, error) {
	e.mu.RLock()
	measure, ok := e.measures[measureURL]
	e.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("measure not found: %s", measureURL)
	}

	patientID, _ := patient["id"].(string)
	subjectRef := "Patient/" + patientID

	reportGroups := e.evaluateGroups(ctx, measure, patient, resources, period)

	report := &MeasureReport{
		ID:      uuid.New().String(),
		Status:  "complete",
		Type:    "individual",
		Measure: measureURL,
		Subject: &subjectRef,
		Period:  period,
		Group:   reportGroups,
	}

	e.mu.Lock()
	e.reports[report.ID] = report
	e.mu.Unlock()

	return report, nil
}

// EvaluatePopulation evaluates a measure across multiple patients.
func (e *MeasureEvaluator) EvaluatePopulation(
	ctx context.Context,
	measureURL string,
	patients []PatientBundle,
	period MeasurePeriod,
) (*MeasureReport, error) {
	e.mu.RLock()
	measure, ok := e.measures[measureURL]
	e.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("measure not found: %s", measureURL)
	}

	// Initialize aggregate population counts.
	type popAccumulator struct {
		count    int
		subjects []string
	}

	// For each group, we accumulate counts across patients.
	type groupAccumulator struct {
		populations map[string]*popAccumulator
	}

	groupAccs := make([]groupAccumulator, len(measure.Group))
	for i, grp := range measure.Group {
		groupAccs[i] = groupAccumulator{
			populations: make(map[string]*popAccumulator),
		}
		for _, pop := range grp.Population {
			groupAccs[i].populations[pop.Code] = &popAccumulator{}
		}
	}

	// Evaluate each patient.
	for _, pb := range patients {
		patientGroups := e.evaluateGroups(ctx, measure, pb.Patient, pb.Resources, period)
		patientID, _ := pb.Patient["id"].(string)
		patientRef := "Patient/" + patientID

		for gi, grp := range patientGroups {
			if gi >= len(groupAccs) {
				break
			}
			for _, pop := range grp.Population {
				acc := groupAccs[gi].populations[pop.Code]
				if acc == nil {
					acc = &popAccumulator{}
					groupAccs[gi].populations[pop.Code] = acc
				}
				acc.count += pop.Count
				if pop.Count > 0 {
					acc.subjects = append(acc.subjects, patientRef)
				}
			}
		}
	}

	// Build report groups.
	var reportGroups []MeasureReportGroup
	for gi, grp := range measure.Group {
		rg := MeasureReportGroup{Code: grp.Code}
		var denomCount, numCount int

		for _, pop := range grp.Population {
			acc := groupAccs[gi].populations[pop.Code]
			count := 0
			var subjects []string
			if acc != nil {
				count = acc.count
				subjects = acc.subjects
			}
			rg.Population = append(rg.Population, MeasureReportPopulation{
				Code:           pop.Code,
				Count:          count,
				SubjectResults: subjects,
			})
			if pop.Code == "denominator" {
				denomCount = count
			}
			if pop.Code == "numerator" {
				numCount = count
			}
		}

		// Calculate measure score for proportion scoring.
		if measure.Scoring == "proportion" && denomCount > 0 {
			score := float64(numCount) / float64(denomCount)
			rg.MeasureScore = &score
		}

		reportGroups = append(reportGroups, rg)
	}

	report := &MeasureReport{
		ID:      uuid.New().String(),
		Status:  "complete",
		Type:    "summary",
		Measure: measureURL,
		Period:  period,
		Group:   reportGroups,
	}

	e.mu.Lock()
	e.reports[report.ID] = report
	e.mu.Unlock()

	return report, nil
}

// evaluateGroups evaluates all population criteria for a single patient.
func (e *MeasureEvaluator) evaluateGroups(
	ctx context.Context,
	measure *Measure,
	patient map[string]interface{},
	resources map[string][]map[string]interface{},
	period MeasurePeriod,
) []MeasureReportGroup {
	var groups []MeasureReportGroup

	for _, grp := range measure.Group {
		rg := MeasureReportGroup{Code: grp.Code}
		popResults := make(map[string]bool)

		// First pass: evaluate each population expression.
		for _, pop := range grp.Population {
			result := e.evaluatePopulationExpression(ctx, pop.Expression, patient, resources, period)
			popResults[pop.Code] = result
		}

		// Apply measure logic: if not in initial-population, cannot be in
		// other populations.  If in denominator-exclusion, excluded from
		// denominator (and hence numerator).
		inIP := popResults["initial-population"]
		inDenomExcl := popResults["denominator-exclusion"]

		for _, pop := range grp.Population {
			count := 0
			switch pop.Code {
			case "initial-population":
				if inIP {
					count = 1
				}
			case "denominator":
				if inIP && !inDenomExcl {
					count = 1
				}
			case "denominator-exclusion":
				if inIP && inDenomExcl {
					count = 1
				}
			case "numerator":
				if inIP && !inDenomExcl && popResults["numerator"] {
					count = 1
				}
			case "numerator-exclusion":
				if inIP && popResults["numerator-exclusion"] {
					count = 1
				}
			case "denominator-exception":
				if inIP && popResults["denominator-exception"] {
					count = 1
				}
			default:
				if popResults[pop.Code] {
					count = 1
				}
			}
			rg.Population = append(rg.Population, MeasureReportPopulation{
				Code:  pop.Code,
				Count: count,
			})
		}

		groups = append(groups, rg)
	}

	return groups
}

// evaluatePopulationExpression evaluates a named CQL expression that maps to
// a built-in measure population function or a library definition.
func (e *MeasureEvaluator) evaluatePopulationExpression(
	ctx context.Context,
	exprName string,
	patient map[string]interface{},
	resources map[string][]map[string]interface{},
	period MeasurePeriod,
) bool {
	// First try built-in expressions (they take precedence for built-in measures).
	if result, handled := e.tryBuiltinExpression(exprName, patient, resources, period); handled {
		return result
	}

	// Try to find the expression in registered libraries.
	e.mu.RLock()
	for _, lib := range e.libraries {
		if def, ok := lib.Definitions[exprName]; ok {
			e.mu.RUnlock()
			val, err := e.cql.EvaluateExpression(ctx, def.Expression, patient, resources)
			if err != nil {
				return false
			}
			return cqlToBool(val)
		}
	}
	e.mu.RUnlock()

	// Try to evaluate as a raw CQL expression.
	val, err := e.cql.EvaluateExpression(ctx, exprName, patient, resources)
	if err != nil {
		return false
	}
	return cqlToBool(val)
}

// tryBuiltinExpression evaluates built-in named expressions for the
// three quality measures (CMS122, CMS125, CMS165). Returns (result, true)
// if the expression was recognized, or (false, false) if not.
func (e *MeasureEvaluator) tryBuiltinExpression(
	exprName string,
	patient map[string]interface{},
	resources map[string][]map[string]interface{},
	period MeasurePeriod,
) (bool, bool) {
	switch exprName {
	// --- CMS122: Diabetes HbA1c ---
	case "CMS122_InitialPopulation":
		return e.cms122InitialPopulation(patient, resources, period), true
	case "CMS122_Denominator":
		return e.cms122InitialPopulation(patient, resources, period), true
	case "CMS122_DenominatorExclusion":
		return e.cms122DenominatorExclusion(resources), true
	case "CMS122_Numerator":
		return e.cms122Numerator(patient, resources, period), true

	// --- CMS125: Breast Cancer Screening ---
	case "CMS125_InitialPopulation":
		return e.cms125InitialPopulation(patient, period), true
	case "CMS125_Denominator":
		return e.cms125InitialPopulation(patient, period), true
	case "CMS125_Numerator":
		return e.cms125Numerator(resources, period), true

	// --- CMS165: Controlling High BP ---
	case "CMS165_InitialPopulation":
		return e.cms165InitialPopulation(patient, resources, period), true
	case "CMS165_Denominator":
		return e.cms165InitialPopulation(patient, resources, period), true
	case "CMS165_DenominatorExclusion":
		return e.cms165DenominatorExclusion(resources), true
	case "CMS165_Numerator":
		return e.cms165Numerator(resources, period), true

	default:
		return false, false
	}
}

// ---------------------------------------------------------------------------
// CMS122 — Diabetes: Hemoglobin A1c (HbA1c) Poor Control (> 9%)
// ---------------------------------------------------------------------------

func (e *MeasureEvaluator) cms122InitialPopulation(
	patient map[string]interface{},
	resources map[string][]map[string]interface{},
	period MeasurePeriod,
) bool {
	age, err := cqlAgeInYears(patient, period.End)
	if err != nil || age < 18 || age > 75 {
		return false
	}
	return hasResourceCode(resources, "Condition", "E11")
}

func (e *MeasureEvaluator) cms122DenominatorExclusion(
	resources map[string][]map[string]interface{},
) bool {
	return hasEncounterType(resources, "hospice") ||
		hasEncounterType(resources, "385765002")
}

func (e *MeasureEvaluator) cms122Numerator(
	patient map[string]interface{},
	resources map[string][]map[string]interface{},
	period MeasurePeriod,
) bool {
	// Find the most recent HbA1c in period with value <= 9%.
	var latestDate time.Time
	var latestValue float64
	found := false

	for _, obs := range resources["Observation"] {
		if !resourceMatchesCode(obs, "4548-4") {
			continue
		}
		dateStr, _ := obs["effectiveDateTime"].(string)
		if dateStr == "" {
			continue
		}
		obsDate, err := parseFlexDate(dateStr)
		if err != nil {
			continue
		}
		if obsDate.Before(period.Start) || obsDate.After(period.End) {
			continue
		}
		vq, _ := obs["valueQuantity"].(map[string]interface{})
		if vq == nil {
			continue
		}
		val := cqlToFloat64(vq["value"])
		if !found || obsDate.After(latestDate) {
			latestDate = obsDate
			latestValue = val
			found = true
		}
	}

	return found && latestValue <= 9.0
}

// ---------------------------------------------------------------------------
// CMS125 — Breast Cancer Screening
// ---------------------------------------------------------------------------

func (e *MeasureEvaluator) cms125InitialPopulation(
	patient map[string]interface{},
	period MeasurePeriod,
) bool {
	gender, _ := patient["gender"].(string)
	if gender != "female" {
		return false
	}
	age, err := cqlAgeInYears(patient, period.End)
	if err != nil {
		return false
	}
	return age >= 52 && age <= 74
}

func (e *MeasureEvaluator) cms125Numerator(
	resources map[string][]map[string]interface{},
	period MeasurePeriod,
) bool {
	// Mammogram within last 27 months from period end.
	cutoff := period.End.AddDate(-2, -3, 0) // ~27 months back

	for _, dr := range resources["DiagnosticReport"] {
		if !resourceMatchesCode(dr, "24606-6") {
			continue
		}
		dateStr, _ := dr["effectiveDateTime"].(string)
		if dateStr == "" {
			continue
		}
		drDate, err := parseFlexDate(dateStr)
		if err != nil {
			continue
		}
		if !drDate.Before(cutoff) && !drDate.After(period.End) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// CMS165 — Controlling High Blood Pressure
// ---------------------------------------------------------------------------

func (e *MeasureEvaluator) cms165InitialPopulation(
	patient map[string]interface{},
	resources map[string][]map[string]interface{},
	period MeasurePeriod,
) bool {
	age, err := cqlAgeInYears(patient, period.End)
	if err != nil || age < 18 || age > 85 {
		return false
	}
	return hasResourceCode(resources, "Condition", "I10")
}

func (e *MeasureEvaluator) cms165DenominatorExclusion(
	resources map[string][]map[string]interface{},
) bool {
	// ESRD, pregnancy, or hospice.
	if hasResourceCode(resources, "Condition", "N18.6") {
		return true
	}
	if hasResourceCode(resources, "Condition", "O") { // pregnancy ICD-10 chapter
		return true
	}
	return hasEncounterType(resources, "hospice") ||
		hasEncounterType(resources, "385765002")
}

func (e *MeasureEvaluator) cms165Numerator(
	resources map[string][]map[string]interface{},
	period MeasurePeriod,
) bool {
	// Most recent BP in period: systolic < 140 AND diastolic < 90.
	var latestDate time.Time
	var latestSys, latestDia float64
	found := false

	for _, obs := range resources["Observation"] {
		if !resourceMatchesCode(obs, "85354-9") {
			continue
		}
		dateStr, _ := obs["effectiveDateTime"].(string)
		if dateStr == "" {
			continue
		}
		obsDate, err := parseFlexDate(dateStr)
		if err != nil {
			continue
		}
		if obsDate.Before(period.Start) || obsDate.After(period.End) {
			continue
		}
		sys := extractBPComponentValue(obs, "8480-6")
		dia := extractBPComponentValue(obs, "8462-4")
		if !found || obsDate.After(latestDate) {
			latestDate = obsDate
			latestSys = sys
			latestDia = dia
			found = true
		}
	}

	return found && latestSys < 140 && latestDia < 90
}

// extractBPComponentValue extracts the numeric value of a BP component.
func extractBPComponentValue(obs map[string]interface{}, componentCode string) float64 {
	comps, _ := obs["component"].([]interface{})
	for _, comp := range comps {
		compMap, _ := comp.(map[string]interface{})
		if compMap == nil {
			continue
		}
		codeObj, _ := compMap["code"].(map[string]interface{})
		if codeObj == nil {
			continue
		}
		codings, _ := codeObj["coding"].([]interface{})
		for _, c := range codings {
			coding, _ := c.(map[string]interface{})
			if coding == nil {
				continue
			}
			cc, _ := coding["code"].(string)
			if cc == componentCode {
				vq, _ := compMap["valueQuantity"].(map[string]interface{})
				if vq != nil {
					return cqlToFloat64(vq["value"])
				}
			}
		}
	}
	return 0
}

// ---------------------------------------------------------------------------
// Built-in measure registration
// ---------------------------------------------------------------------------

func (e *MeasureEvaluator) registerBuiltinMeasures() {
	// CMS122 — Diabetes HbA1c
	cms122Lib := &CQLLibrary{
		Name:    "DiabetesHbA1cControlLibrary",
		Version: "1.0.0",
		URL:     "http://hl7.org/fhir/us/cqfmeasures/Library/CMS122",
		Status:  "active",
		Definitions: map[string]CQLDefinition{
			"CMS122_InitialPopulation": {
				Name:       "CMS122_InitialPopulation",
				Expression: "CMS122_InitialPopulation",
				Context:    "Patient",
				Type:       "Boolean",
			},
			"CMS122_Denominator": {
				Name:       "CMS122_Denominator",
				Expression: "CMS122_Denominator",
				Context:    "Patient",
				Type:       "Boolean",
			},
			"CMS122_DenominatorExclusion": {
				Name:       "CMS122_DenominatorExclusion",
				Expression: "CMS122_DenominatorExclusion",
				Context:    "Patient",
				Type:       "Boolean",
			},
			"CMS122_Numerator": {
				Name:       "CMS122_Numerator",
				Expression: "CMS122_Numerator",
				Context:    "Patient",
				Type:       "Boolean",
			},
		},
		CreatedAt: time.Now(),
	}
	e.libraries[cms122Lib.URL] = cms122Lib

	cms122 := &Measure{
		ID:          "cms122",
		URL:         CMS122URL,
		Name:        "DiabetesHbA1cControl",
		Title:       "Diabetes: Hemoglobin A1c (HbA1c) Poor Control (> 9%)",
		Status:      "active",
		Description: "Percentage of patients 18-75 years of age with diabetes who had hemoglobin A1c > 9.0% during the measurement period.",
		Scoring:     "proportion",
		Type:        []string{"process"},
		Library:     []string{cms122Lib.URL},
		Date:        time.Now(),
		Group: []MeasureGroup{
			{
				ID:          "cms122-group-1",
				Description: "Diabetes HbA1c control group",
				Population: []MeasurePopulation{
					{Code: "initial-population", Expression: "CMS122_InitialPopulation"},
					{Code: "denominator", Expression: "CMS122_Denominator"},
					{Code: "denominator-exclusion", Expression: "CMS122_DenominatorExclusion"},
					{Code: "numerator", Expression: "CMS122_Numerator"},
				},
			},
		},
	}
	e.measures[cms122.URL] = cms122
	e.measureByID[cms122.ID] = cms122.URL

	// CMS125 — Breast Cancer Screening
	cms125Lib := &CQLLibrary{
		Name:    "BreastCancerScreeningLibrary",
		Version: "1.0.0",
		URL:     "http://hl7.org/fhir/us/cqfmeasures/Library/CMS125",
		Status:  "active",
		Definitions: map[string]CQLDefinition{
			"CMS125_InitialPopulation": {
				Name:       "CMS125_InitialPopulation",
				Expression: "CMS125_InitialPopulation",
				Context:    "Patient",
				Type:       "Boolean",
			},
			"CMS125_Denominator": {
				Name:       "CMS125_Denominator",
				Expression: "CMS125_Denominator",
				Context:    "Patient",
				Type:       "Boolean",
			},
			"CMS125_Numerator": {
				Name:       "CMS125_Numerator",
				Expression: "CMS125_Numerator",
				Context:    "Patient",
				Type:       "Boolean",
			},
		},
		CreatedAt: time.Now(),
	}
	e.libraries[cms125Lib.URL] = cms125Lib

	cms125 := &Measure{
		ID:          "cms125",
		URL:         CMS125URL,
		Name:        "BreastCancerScreening",
		Title:       "Breast Cancer Screening",
		Status:      "active",
		Description: "Percentage of women 52-74 years of age who had a mammogram to screen for breast cancer in the 27 months prior to the end of the measurement period.",
		Scoring:     "proportion",
		Type:        []string{"process"},
		Library:     []string{cms125Lib.URL},
		Date:        time.Now(),
		Group: []MeasureGroup{
			{
				ID:          "cms125-group-1",
				Description: "Breast cancer screening group",
				Population: []MeasurePopulation{
					{Code: "initial-population", Expression: "CMS125_InitialPopulation"},
					{Code: "denominator", Expression: "CMS125_Denominator"},
					{Code: "numerator", Expression: "CMS125_Numerator"},
				},
			},
		},
	}
	e.measures[cms125.URL] = cms125
	e.measureByID[cms125.ID] = cms125.URL

	// CMS165 — Controlling High Blood Pressure
	cms165Lib := &CQLLibrary{
		Name:    "ControllingHighBPLibrary",
		Version: "1.0.0",
		URL:     "http://hl7.org/fhir/us/cqfmeasures/Library/CMS165",
		Status:  "active",
		Definitions: map[string]CQLDefinition{
			"CMS165_InitialPopulation": {
				Name:       "CMS165_InitialPopulation",
				Expression: "CMS165_InitialPopulation",
				Context:    "Patient",
				Type:       "Boolean",
			},
			"CMS165_Denominator": {
				Name:       "CMS165_Denominator",
				Expression: "CMS165_Denominator",
				Context:    "Patient",
				Type:       "Boolean",
			},
			"CMS165_DenominatorExclusion": {
				Name:       "CMS165_DenominatorExclusion",
				Expression: "CMS165_DenominatorExclusion",
				Context:    "Patient",
				Type:       "Boolean",
			},
			"CMS165_Numerator": {
				Name:       "CMS165_Numerator",
				Expression: "CMS165_Numerator",
				Context:    "Patient",
				Type:       "Boolean",
			},
		},
		CreatedAt: time.Now(),
	}
	e.libraries[cms165Lib.URL] = cms165Lib

	cms165 := &Measure{
		ID:          "cms165",
		URL:         CMS165URL,
		Name:        "ControllingHighBloodPressure",
		Title:       "Controlling High Blood Pressure",
		Status:      "active",
		Description: "Percentage of patients 18-85 years of age who had a diagnosis of essential hypertension starting before and continuing into the measurement period, and whose most recent blood pressure was adequately controlled.",
		Scoring:     "proportion",
		Type:        []string{"outcome"},
		Library:     []string{cms165Lib.URL},
		Date:        time.Now(),
		Group: []MeasureGroup{
			{
				ID:          "cms165-group-1",
				Description: "BP control group",
				Population: []MeasurePopulation{
					{Code: "initial-population", Expression: "CMS165_InitialPopulation"},
					{Code: "denominator", Expression: "CMS165_Denominator"},
					{Code: "denominator-exclusion", Expression: "CMS165_DenominatorExclusion"},
					{Code: "numerator", Expression: "CMS165_Numerator"},
				},
			},
		},
	}
	e.measures[cms165.URL] = cms165
	e.measureByID[cms165.ID] = cms165.URL
}

// ---------------------------------------------------------------------------
// Utility helpers
// ---------------------------------------------------------------------------

// cqlToBool converts various types to a boolean.
func cqlToBool(val interface{}) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case float64:
		return v != 0
	case string:
		return v != ""
	default:
		return true
	}
}

// cqlToFloat64 converts numeric interface{} values to float64.
func cqlToFloat64(val interface{}) float64 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case json.Number:
		f, _ := v.Float64()
		return f
	default:
		return 0
	}
}

// ============================================================================
// HTTP Handler — MeasureHandler
// ============================================================================

// MeasureHandler provides REST endpoints for FHIR Measure and MeasureReport
// resources, including the $evaluate-measure operation.
type MeasureHandler struct {
	evaluator *MeasureEvaluator
}

// NewMeasureHandler creates a new MeasureHandler.
func NewMeasureHandler(evaluator *MeasureEvaluator) *MeasureHandler {
	return &MeasureHandler{evaluator: evaluator}
}

// RegisterRoutes registers the measure routes on the FHIR group.
func (h *MeasureHandler) RegisterRoutes(fhirGroup *echo.Group) {
	fhirGroup.GET("/Measure", h.ListMeasures)
	fhirGroup.GET("/Measure/:id", h.GetMeasure)
	fhirGroup.POST("/Measure", h.CreateMeasure)
	fhirGroup.POST("/Measure/:id/$evaluate-measure", h.EvaluateMeasure)
	fhirGroup.GET("/MeasureReport", h.ListReports)
	fhirGroup.GET("/MeasureReport/:id", h.GetReport)
	fhirGroup.GET("/Library", h.ListLibraries)
	fhirGroup.POST("/Library", h.CreateLibrary)
}

// ListMeasures returns all registered measures as a FHIR Bundle.
func (h *MeasureHandler) ListMeasures(c echo.Context) error {
	h.evaluator.mu.RLock()
	defer h.evaluator.mu.RUnlock()

	entries := make([]interface{}, 0, len(h.evaluator.measures))
	for _, m := range h.evaluator.measures {
		entries = append(entries, map[string]interface{}{
			"resource": m.ToFHIR(),
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

// GetMeasure returns a single measure by ID.
func (h *MeasureHandler) GetMeasure(c echo.Context) error {
	id := c.Param("id")

	h.evaluator.mu.RLock()
	url, ok := h.evaluator.measureByID[id]
	if !ok {
		h.evaluator.mu.RUnlock()
		return c.JSON(http.StatusNotFound, ErrorOutcome("Measure/"+id+" not found"))
	}
	measure := h.evaluator.measures[url]
	h.evaluator.mu.RUnlock()

	return c.JSON(http.StatusOK, measure.ToFHIR())
}

// CreateMeasure registers a custom Measure.
func (h *MeasureHandler) CreateMeasure(c echo.Context) error {
	var body map[string]interface{}
	if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid request body"))
	}

	rt, _ := body["resourceType"].(string)
	if rt != "Measure" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("resourceType must be Measure"))
	}

	m := parseMeasureFromFHIR(body)
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	if m.URL == "" {
		m.URL = "http://example.org/fhir/Measure/" + m.ID
	}

	h.evaluator.RegisterMeasure(m)

	return c.JSON(http.StatusCreated, m.ToFHIR())
}

// EvaluateMeasure handles POST /fhir/Measure/:id/$evaluate-measure.
func (h *MeasureHandler) EvaluateMeasure(c echo.Context) error {
	id := c.Param("id")

	h.evaluator.mu.RLock()
	measureURL, ok := h.evaluator.measureByID[id]
	h.evaluator.mu.RUnlock()
	if !ok {
		return c.JSON(http.StatusNotFound, ErrorOutcome("Measure/"+id+" not found"))
	}

	// Parse period parameters.
	periodStart := c.QueryParam("periodStart")
	periodEnd := c.QueryParam("periodEnd")
	if periodStart == "" || periodEnd == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("periodStart and periodEnd query parameters are required"))
	}

	start, err := parseFlexDate(periodStart)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid periodStart: "+err.Error()))
	}
	end, err := parseFlexDate(periodEnd)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid periodEnd: "+err.Error()))
	}
	// Set end to end-of-day.
	end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	period := MeasurePeriod{Start: start, End: end}
	reportType := c.QueryParam("reportType")
	if reportType == "" {
		reportType = "individual"
	}

	// Parse the request body as a FHIR Bundle.
	var body map[string]interface{}
	if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid request body"))
	}

	bundles := parseBundleToPatientBundles(body)

	ctx := c.Request().Context()

	if reportType == "summary" || reportType == "subject-list" || len(bundles) > 1 {
		report, err := h.evaluator.EvaluatePopulation(ctx, measureURL, bundles, period)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorOutcome(err.Error()))
		}
		if reportType == "subject-list" {
			report.Type = "subject-list"
		}
		return c.JSON(http.StatusOK, report.ToFHIR())
	}

	// Individual evaluation.
	if len(bundles) == 0 {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("no patient data in request body"))
	}
	pb := bundles[0]
	report, err := h.evaluator.EvaluateIndividual(ctx, measureURL, pb.Patient, pb.Resources, period)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, report.ToFHIR())
}

// ListReports returns all generated MeasureReports as a FHIR Bundle.
func (h *MeasureHandler) ListReports(c echo.Context) error {
	h.evaluator.mu.RLock()
	defer h.evaluator.mu.RUnlock()

	entries := make([]interface{}, 0, len(h.evaluator.reports))
	for _, r := range h.evaluator.reports {
		entries = append(entries, map[string]interface{}{
			"resource": r.ToFHIR(),
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

// GetReport returns a single MeasureReport by ID.
func (h *MeasureHandler) GetReport(c echo.Context) error {
	id := c.Param("id")

	h.evaluator.mu.RLock()
	report, ok := h.evaluator.reports[id]
	h.evaluator.mu.RUnlock()
	if !ok {
		return c.JSON(http.StatusNotFound, ErrorOutcome("MeasureReport/"+id+" not found"))
	}

	return c.JSON(http.StatusOK, report.ToFHIR())
}

// ListLibraries returns all registered CQL libraries as a FHIR Bundle.
func (h *MeasureHandler) ListLibraries(c echo.Context) error {
	h.evaluator.mu.RLock()
	defer h.evaluator.mu.RUnlock()

	entries := make([]interface{}, 0, len(h.evaluator.libraries))
	for _, lib := range h.evaluator.libraries {
		entries = append(entries, map[string]interface{}{
			"resource": cqlLibraryToFHIR(lib),
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

// CreateLibrary registers a CQL library from a FHIR Library resource.
func (h *MeasureHandler) CreateLibrary(c echo.Context) error {
	var body map[string]interface{}
	if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid request body"))
	}

	rt, _ := body["resourceType"].(string)
	if rt != "Library" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("resourceType must be Library"))
	}

	lib := parseLibraryFromFHIR(body)
	h.evaluator.RegisterLibrary(lib)

	return c.JSON(http.StatusCreated, cqlLibraryToFHIR(lib))
}

// ---------------------------------------------------------------------------
// FHIR parse/serialize helpers
// ---------------------------------------------------------------------------

// cqlLibraryToFHIR converts a CQLLibrary to a FHIR Library JSON map.
func cqlLibraryToFHIR(lib *CQLLibrary) map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Library",
		"url":          lib.URL,
		"name":         lib.Name,
		"version":      lib.Version,
		"status":       lib.Status,
	}
	if !lib.CreatedAt.IsZero() {
		result["date"] = lib.CreatedAt.Format(time.RFC3339)
	}
	return result
}

// parseLibraryFromFHIR parses a FHIR Library JSON map into a CQLLibrary.
func parseLibraryFromFHIR(body map[string]interface{}) *CQLLibrary {
	lib := &CQLLibrary{
		URL:         cqlGetString(body, "url"),
		Name:        cqlGetString(body, "name"),
		Version:     cqlGetString(body, "version"),
		Status:      cqlGetString(body, "status"),
		Definitions: make(map[string]CQLDefinition),
		CreatedAt:   time.Now(),
	}

	// Parse CQL content from content array.
	contents, _ := body["content"].([]interface{})
	for _, c := range contents {
		contentMap, _ := c.(map[string]interface{})
		if contentMap == nil {
			continue
		}
		ct, _ := contentMap["contentType"].(string)
		if ct == "text/cql" {
			data, _ := contentMap["data"].(string)
			parseCQLContent(lib, data)
		}
	}

	return lib
}

// parseCQLContent does a basic parse of CQL text to extract define statements.
func parseCQLContent(lib *CQLLibrary, content string) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "define ") {
			// Format: define <name>: <expression>
			rest := strings.TrimPrefix(line, "define ")
			parts := strings.SplitN(rest, ":", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				expr := strings.TrimSpace(parts[1])
				lib.Definitions[name] = CQLDefinition{
					Name:       name,
					Expression: expr,
					Context:    "Patient",
				}
			}
		}
	}
}

// parseMeasureFromFHIR parses a FHIR Measure JSON map into a Measure.
func parseMeasureFromFHIR(body map[string]interface{}) *Measure {
	m := &Measure{
		ID:          cqlGetString(body, "id"),
		URL:         cqlGetString(body, "url"),
		Name:        cqlGetString(body, "name"),
		Title:       cqlGetString(body, "title"),
		Status:      cqlGetString(body, "status"),
		Description: cqlGetString(body, "description"),
		Date:        time.Now(),
	}

	// Parse scoring.
	if scoring, ok := body["scoring"].(map[string]interface{}); ok {
		codings, _ := scoring["coding"].([]interface{})
		if len(codings) > 0 {
			if c, ok := codings[0].(map[string]interface{}); ok {
				m.Scoring, _ = c["code"].(string)
			}
		}
	}

	// Parse groups.
	groups, _ := body["group"].([]interface{})
	for _, g := range groups {
		gMap, _ := g.(map[string]interface{})
		if gMap == nil {
			continue
		}
		mg := MeasureGroup{
			ID: cqlGetString(gMap, "id"),
		}
		pops, _ := gMap["population"].([]interface{})
		for _, p := range pops {
			pMap, _ := p.(map[string]interface{})
			if pMap == nil {
				continue
			}
			mp := MeasurePopulation{}
			if codeObj, ok := pMap["code"].(map[string]interface{}); ok {
				codings, _ := codeObj["coding"].([]interface{})
				if len(codings) > 0 {
					if c, ok := codings[0].(map[string]interface{}); ok {
						mp.Code, _ = c["code"].(string)
					}
				}
			}
			if criteria, ok := pMap["criteria"].(map[string]interface{}); ok {
				mp.Expression, _ = criteria["expression"].(string)
			}
			mg.Population = append(mg.Population, mp)
		}
		m.Group = append(m.Group, mg)
	}

	return m
}

// parseBundleToPatientBundles parses a FHIR Bundle into PatientBundles,
// grouping resources by patient reference.
func parseBundleToPatientBundles(body map[string]interface{}) []PatientBundle {
	entries, _ := body["entry"].([]interface{})
	if len(entries) == 0 {
		return nil
	}

	// Collect all patients and non-patient resources.
	var patients []map[string]interface{}
	var otherResources []map[string]interface{}

	for _, entry := range entries {
		eMap, _ := entry.(map[string]interface{})
		if eMap == nil {
			continue
		}
		res, _ := eMap["resource"].(map[string]interface{})
		if res == nil {
			continue
		}
		rt, _ := res["resourceType"].(string)
		if rt == "Patient" {
			patients = append(patients, res)
		} else {
			otherResources = append(otherResources, res)
		}
	}

	if len(patients) == 0 {
		return nil
	}

	// If only one patient, all resources belong to them.
	if len(patients) == 1 {
		resources := make(map[string][]map[string]interface{})
		for _, res := range otherResources {
			rt, _ := res["resourceType"].(string)
			resources[rt] = append(resources[rt], res)
		}
		return []PatientBundle{{Patient: patients[0], Resources: resources}}
	}

	// Multiple patients: group resources by subject reference.
	bundles := make([]PatientBundle, len(patients))
	patientIndex := make(map[string]int) // Patient/<id> -> index
	for i, p := range patients {
		pid, _ := p["id"].(string)
		patientIndex["Patient/"+pid] = i
		bundles[i] = PatientBundle{
			Patient:   p,
			Resources: make(map[string][]map[string]interface{}),
		}
	}

	for _, res := range otherResources {
		rt, _ := res["resourceType"].(string)
		// Try to find subject reference.
		subj, _ := res["subject"].(map[string]interface{})
		ref, _ := subj["reference"].(string)
		if idx, ok := patientIndex[ref]; ok {
			bundles[idx].Resources[rt] = append(bundles[idx].Resources[rt], res)
		} else {
			// No subject reference; assign to first patient.
			bundles[0].Resources[rt] = append(bundles[0].Resources[rt], res)
		}
	}

	return bundles
}

// cqlGetString safely extracts a string value from a map.
func cqlGetString(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}
