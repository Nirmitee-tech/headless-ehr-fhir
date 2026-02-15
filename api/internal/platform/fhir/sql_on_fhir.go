package fhir

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

// ViewDefinition represents a SQL-on-FHIR v2 ViewDefinition resource.
type ViewDefinition struct {
	ID        string         `json:"id"`
	URL       string         `json:"url,omitempty"`
	Name      string         `json:"name"`
	Title     string         `json:"title,omitempty"`
	Status    string         `json:"status,omitempty"`
	Resource  string         `json:"resource"`
	Select    []ViewColumn   `json:"select"`
	Where     []ViewWhere    `json:"where,omitempty"`
	Constants []ViewConstant `json:"constant,omitempty"`
}

// ViewColumn defines a single column in a ViewDefinition output.
type ViewColumn struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Collection  bool   `json:"collection,omitempty"`
	Description string `json:"description,omitempty"`
}

// ViewWhere defines a filter expression. All where clauses are ANDed.
type ViewWhere struct {
	Path string `json:"path"`
}

// ViewConstant defines a named constant available in FHIRPath expressions.
type ViewConstant struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ViewResult holds the tabular output of executing a ViewDefinition.
type ViewResult struct {
	Columns []ViewColumn    `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
}

// ViewDefinitionEngine evaluates ViewDefinition resources against FHIR resources.
type ViewDefinitionEngine struct {
	fhirpath *FHIRPathEngine
}

// NewViewDefinitionEngine creates a new engine backed by the given FHIRPathEngine.
func NewViewDefinitionEngine(fhirpath *FHIRPathEngine) *ViewDefinitionEngine {
	return &ViewDefinitionEngine{fhirpath: fhirpath}
}

// Execute evaluates a ViewDefinition against a set of FHIR resources.
func (e *ViewDefinitionEngine) Execute(_ context.Context, view *ViewDefinition, resources []map[string]interface{}) (*ViewResult, error) {
	result := &ViewResult{
		Columns: make([]ViewColumn, len(view.Select)),
		Rows:    make([][]interface{}, 0),
	}
	copy(result.Columns, view.Select)

	if len(resources) == 0 {
		return result, nil
	}

	for _, resource := range resources {
		rt, _ := resource["resourceType"].(string)
		if rt != "" && rt != view.Resource {
			continue
		}

		include := true
		for _, w := range view.Where {
			match, err := e.fhirpath.EvaluateBool(resource, w.Path)
			if err != nil {
				include = false
				break
			}
			if !match {
				include = false
				break
			}
		}
		if !include {
			continue
		}

		row := make([]interface{}, len(view.Select))
		for i, col := range view.Select {
			val, err := e.fhirpath.Evaluate(resource, col.Path)
			if err != nil {
				row[i] = nil
				continue
			}
			if len(val) == 0 {
				row[i] = nil
			} else if col.Collection {
				coerced := make([]interface{}, len(val))
				for j, v := range val {
					coerced[j] = coerceValue(v, col.Type)
				}
				row[i] = coerced
			} else {
				row[i] = coerceValue(val[0], col.Type)
			}
		}
		result.Rows = append(result.Rows, row)
	}

	return result, nil
}

func coerceValue(val interface{}, typ string) interface{} {
	if val == nil {
		return nil
	}
	switch typ {
	case "string":
		return fmt.Sprintf("%v", val)
	case "integer":
		return coerceToInt(val)
	case "decimal":
		return coerceToFloat(val)
	case "boolean":
		return coerceToBool(val)
	case "dateTime", "date":
		return fmt.Sprintf("%v", val)
	case "base64Binary":
		return fmt.Sprintf("%v", val)
	default:
		return val
	}
}

func coerceToInt(val interface{}) interface{} {
	switch v := val.(type) {
	case float64:
		return int64(math.Round(v))
	case float32:
		return int64(math.Round(float64(v)))
	case int:
		return int64(v)
	case int64:
		return v
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return int64(math.Round(f))
		}
		return nil
	case bool:
		if v {
			return int64(1)
		}
		return int64(0)
	default:
		return nil
	}
}

func coerceToFloat(val interface{}) interface{} {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
		return nil
	case bool:
		if v {
			return float64(1)
		}
		return float64(0)
	default:
		return nil
	}
}

func coerceToBool(val interface{}) interface{} {
	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "1"
	case float64:
		return v != 0
	case int:
		return v != 0
	case int64:
		return v != 0
	default:
		return val != nil
	}
}

// ToCSV renders a ViewResult as a CSV string.
func (e *ViewDefinitionEngine) ToCSV(result *ViewResult) string {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	header := make([]string, len(result.Columns))
	for i, col := range result.Columns {
		header[i] = col.Name
	}
	_ = w.Write(header)

	for _, row := range result.Rows {
		record := make([]string, len(row))
		for i, val := range row {
			if val == nil {
				record[i] = ""
			} else {
				record[i] = formatCSVValue(val)
			}
		}
		_ = w.Write(record)
	}
	w.Flush()
	return buf.String()
}

func formatCSVValue(val interface{}) string {
	switch v := val.(type) {
	case []interface{}:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ToJSON renders a ViewResult as an array of JSON objects.
func (e *ViewDefinitionEngine) ToJSON(result *ViewResult) []map[string]interface{} {
	objects := make([]map[string]interface{}, len(result.Rows))
	for i, row := range result.Rows {
		obj := make(map[string]interface{}, len(result.Columns))
		for j, col := range result.Columns {
			obj[col.Name] = row[j]
		}
		objects[i] = obj
	}
	return objects
}

// ToNDJSON renders a ViewResult as newline-delimited JSON.
func (e *ViewDefinitionEngine) ToNDJSON(result *ViewResult) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	for _, row := range result.Rows {
		obj := make(map[string]interface{}, len(result.Columns))
		for j, col := range result.Columns {
			obj[col.Name] = row[j]
		}
		_ = enc.Encode(obj)
	}
	return buf.String()
}

// GenerateSQL generates a PostgreSQL CREATE VIEW statement.
func (e *ViewDefinitionEngine) GenerateSQL(view *ViewDefinition) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("CREATE OR REPLACE VIEW %s AS\nSELECT\n", sanitizeSQLName(view.Name)))
	for i, col := range view.Select {
		sqlExpr := fhirPathToJSONBExpression(col.Path, col.Type)
		buf.WriteString(fmt.Sprintf("  %s AS %s", sqlExpr, sanitizeSQLName(col.Name)))
		if i < len(view.Select)-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}
	buf.WriteString(fmt.Sprintf("FROM fhir_resource\nWHERE resource->>'resourceType' = '%s'", view.Resource))
	for _, w := range view.Where {
		sqlWhere := fhirPathWhereToSQL(w.Path)
		if sqlWhere != "" {
			buf.WriteString(fmt.Sprintf("\n  AND %s", sqlWhere))
		}
	}
	buf.WriteString(";\n")
	return buf.String()
}

func sanitizeSQLName(name string) string {
	var sb strings.Builder
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			sb.WriteRune(ch)
		}
	}
	s := sb.String()
	if s == "" {
		return "unnamed"
	}
	return s
}

func fhirPathToJSONBExpression(path string, colType string) string {
	if idx := strings.Index(path, "."); idx > 0 {
		first := path[:idx]
		if len(first) > 0 && first[0] >= 'A' && first[0] <= 'Z' {
			path = path[idx+1:]
		}
	}
	if idx := strings.Index(path, ".where("); idx >= 0 {
		before := path[:idx]
		afterWhere := path[idx:]
		closeParen := strings.Index(afterWhere, ")")
		if closeParen >= 0 && closeParen+1 < len(afterWhere) {
			path = before + afterWhere[closeParen+1:]
		} else {
			path = before
		}
	}
	path = strings.ReplaceAll(path, ".first()", "")
	parts := strings.Split(path, ".")
	if len(parts) == 1 {
		cast := sqlTypeCast(colType)
		return fmt.Sprintf("resource->>'%s'%s", parts[0], cast)
	}
	var expr strings.Builder
	expr.WriteString("resource")
	for i, part := range parts {
		if i == len(parts)-1 {
			expr.WriteString(fmt.Sprintf("->>'%s'", part))
		} else {
			expr.WriteString(fmt.Sprintf("->'%s'", part))
		}
	}
	expr.WriteString(sqlTypeCast(colType))
	return expr.String()
}

func sqlTypeCast(colType string) string {
	switch colType {
	case "integer":
		return "::integer"
	case "decimal":
		return "::numeric"
	case "boolean":
		return "::boolean"
	case "date":
		return "::date"
	case "dateTime":
		return "::timestamptz"
	default:
		return ""
	}
}

func fhirPathWhereToSQL(path string) string {
	eqParts := strings.SplitN(path, "=", 2)
	if len(eqParts) != 2 {
		return fmt.Sprintf("/* unsupported where: %s */", path)
	}
	lhs := strings.TrimSpace(eqParts[0])
	rhs := strings.TrimSpace(eqParts[1])
	rhs = strings.Trim(rhs, "'\"")
	lhsParts := strings.Split(lhs, ".")
	if len(lhsParts) == 1 {
		return fmt.Sprintf("resource->>'%s' = '%s'", lhsParts[0], rhs)
	}
	var expr strings.Builder
	expr.WriteString("resource")
	for i, part := range lhsParts {
		if i == len(lhsParts)-1 {
			expr.WriteString(fmt.Sprintf("->>'%s'", part))
		} else {
			expr.WriteString(fmt.Sprintf("->'%s'", part))
		}
	}
	return fmt.Sprintf("%s = '%s'", expr.String(), rhs)
}

// BuiltInViewDefinitions returns the set of standard pre-registered views.
func BuiltInViewDefinitions() []ViewDefinition {
	return []ViewDefinition{
		builtinPatientDemographics(),
		builtinActiveConditions(),
		builtinLabResults(),
		builtinActiveMedications(),
		builtinVitalSigns(),
		builtinEncountersSummary(),
	}
}

func builtinPatientDemographics() ViewDefinition {
	return ViewDefinition{
		ID:       "patient_demographics",
		URL:      "http://ehr.example.org/ViewDefinition/patient_demographics",
		Name:     "patient_demographics",
		Title:    "Patient Demographics",
		Status:   "active",
		Resource: "Patient",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string", Description: "Patient resource ID"},
			{Path: "name.where(use='official').family", Name: "family_name", Type: "string", Description: "Official family name"},
			{Path: "name.where(use='official').given.first()", Name: "given_name", Type: "string", Description: "First given name"},
			{Path: "birthDate", Name: "birth_date", Type: "date", Description: "Date of birth"},
			{Path: "gender", Name: "gender", Type: "string", Description: "Administrative gender"},
			{Path: "identifier.where(system.contains('mrn')).value", Name: "mrn", Type: "string", Description: "Medical record number"},
			{Path: "telecom.where(system='phone').value.first()", Name: "phone", Type: "string", Description: "Phone number"},
			{Path: "telecom.where(system='email').value.first()", Name: "email", Type: "string", Description: "Email address"},
			{Path: "active", Name: "active", Type: "boolean", Description: "Whether active"},
		},
	}
}

func builtinActiveConditions() ViewDefinition {
	return ViewDefinition{
		ID:       "active_conditions",
		URL:      "http://ehr.example.org/ViewDefinition/active_conditions",
		Name:     "active_conditions",
		Title:    "Active Conditions",
		Status:   "active",
		Resource: "Condition",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string"},
			{Path: "subject.reference", Name: "patient_id", Type: "string"},
			{Path: "code.coding.code.first()", Name: "code", Type: "string"},
			{Path: "code.coding.display.first()", Name: "code_display", Type: "string"},
			{Path: "clinicalStatus.coding.code.first()", Name: "clinical_status", Type: "string"},
			{Path: "verificationStatus.coding.code.first()", Name: "verification_status", Type: "string"},
			{Path: "onsetDateTime", Name: "onset_date", Type: "dateTime"},
			{Path: "category.coding.code.first()", Name: "category", Type: "string"},
		},
		Where: []ViewWhere{
			{Path: "clinicalStatus.coding.code = 'active'"},
		},
	}
}

func builtinLabResults() ViewDefinition {
	return ViewDefinition{
		ID:       "lab_results",
		URL:      "http://ehr.example.org/ViewDefinition/lab_results",
		Name:     "lab_results",
		Title:    "Laboratory Results",
		Status:   "active",
		Resource: "Observation",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string"},
			{Path: "subject.reference", Name: "patient_id", Type: "string"},
			{Path: "code.coding.code.first()", Name: "code", Type: "string"},
			{Path: "code.coding.display.first()", Name: "code_display", Type: "string"},
			{Path: "valueQuantity.value", Name: "value_quantity", Type: "decimal"},
			{Path: "valueQuantity.unit", Name: "value_unit", Type: "string"},
			{Path: "effectiveDateTime", Name: "effective_date", Type: "dateTime"},
			{Path: "status", Name: "status", Type: "string"},
			{Path: "category.coding.code.first()", Name: "category_code", Type: "string"},
		},
		Where: []ViewWhere{
			{Path: "category.coding.code = 'laboratory'"},
		},
	}
}

func builtinActiveMedications() ViewDefinition {
	return ViewDefinition{
		ID:       "active_medications",
		URL:      "http://ehr.example.org/ViewDefinition/active_medications",
		Name:     "active_medications",
		Title:    "Active Medications",
		Status:   "active",
		Resource: "MedicationRequest",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string"},
			{Path: "subject.reference", Name: "patient_id", Type: "string"},
			{Path: "medicationCodeableConcept.coding.code.first()", Name: "medication_code", Type: "string"},
			{Path: "medicationCodeableConcept.coding.display.first()", Name: "medication_display", Type: "string"},
			{Path: "status", Name: "status", Type: "string"},
			{Path: "intent", Name: "intent", Type: "string"},
			{Path: "authoredOn", Name: "authored_on", Type: "dateTime"},
			{Path: "dosageInstruction.text.first()", Name: "dosage_text", Type: "string"},
		},
		Where: []ViewWhere{
			{Path: "status = 'active'"},
		},
	}
}

func builtinVitalSigns() ViewDefinition {
	return ViewDefinition{
		ID:       "vital_signs",
		URL:      "http://ehr.example.org/ViewDefinition/vital_signs",
		Name:     "vital_signs",
		Title:    "Vital Signs",
		Status:   "active",
		Resource: "Observation",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string"},
			{Path: "subject.reference", Name: "patient_id", Type: "string"},
			{Path: "code.coding.code.first()", Name: "code", Type: "string"},
			{Path: "code.coding.display.first()", Name: "code_display", Type: "string"},
			{Path: "valueQuantity.value", Name: "value_quantity", Type: "decimal"},
			{Path: "valueQuantity.unit", Name: "value_unit", Type: "string"},
			{Path: "effectiveDateTime", Name: "effective_date", Type: "dateTime"},
			{Path: "component.where(code.coding.code='8480-6').valueQuantity.value.first()", Name: "systolic", Type: "decimal"},
			{Path: "component.where(code.coding.code='8462-4').valueQuantity.value.first()", Name: "diastolic", Type: "decimal"},
		},
		Where: []ViewWhere{
			{Path: "category.coding.code = 'vital-signs'"},
		},
	}
}

func builtinEncountersSummary() ViewDefinition {
	return ViewDefinition{
		ID:       "encounters_summary",
		URL:      "http://ehr.example.org/ViewDefinition/encounters_summary",
		Name:     "encounters_summary",
		Title:    "Encounters Summary",
		Status:   "active",
		Resource: "Encounter",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string"},
			{Path: "subject.reference", Name: "patient_id", Type: "string"},
			{Path: "status", Name: "status", Type: "string"},
			{Path: "class.code", Name: "class_code", Type: "string"},
			{Path: "type.coding.code.first()", Name: "type_code", Type: "string"},
			{Path: "type.coding.display.first()", Name: "type_display", Type: "string"},
			{Path: "period.start", Name: "period_start", Type: "dateTime"},
			{Path: "period.end", Name: "period_end", Type: "dateTime"},
			{Path: "reasonCode.coding.code.first()", Name: "reason_code", Type: "string"},
			{Path: "reasonCode.coding.display.first()", Name: "reason_display", Type: "string"},
		},
	}
}

// ViewDefinitionHandler provides CRUD and execution endpoints.
type ViewDefinitionHandler struct {
	mu     sync.RWMutex
	views  map[string]*ViewDefinition
	engine *ViewDefinitionEngine
}

// NewViewDefinitionHandler creates a new handler.
func NewViewDefinitionHandler(engine *ViewDefinitionEngine) *ViewDefinitionHandler {
	return &ViewDefinitionHandler{
		views:  make(map[string]*ViewDefinition),
		engine: engine,
	}
}

// LoadBuiltIns loads the built-in view definitions into the handler.
func (h *ViewDefinitionHandler) LoadBuiltIns() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, v := range BuiltInViewDefinitions() {
		vc := v
		h.views[vc.ID] = &vc
	}
}

// RegisterRoutes registers all ViewDefinition routes on the given Echo group.
func (h *ViewDefinitionHandler) RegisterRoutes(fhirGroup *echo.Group) {
	fhirGroup.GET("/ViewDefinition", h.List)
	fhirGroup.GET("/ViewDefinition/:id", h.Get)
	fhirGroup.POST("/ViewDefinition", h.Create)
	fhirGroup.PUT("/ViewDefinition/:id", h.Update)
	fhirGroup.DELETE("/ViewDefinition/:id", h.Delete)
	fhirGroup.POST("/ViewDefinition/:id/$execute", h.ExecuteView)
	fhirGroup.GET("/ViewDefinition/:id/$sql", h.SQLView)
}

// List returns all registered view definitions.
func (h *ViewDefinitionHandler) List(c echo.Context) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]ViewDefinition, 0, len(h.views))
	for _, v := range h.views {
		result = append(result, *v)
	}
	return c.JSON(http.StatusOK, result)
}

// Get returns a single view definition by ID.
func (h *ViewDefinitionHandler) Get(c echo.Context) error {
	id := c.Param("id")
	h.mu.RLock()
	defer h.mu.RUnlock()
	view, ok := h.views[id]
	if !ok {
		return c.JSON(http.StatusNotFound, ErrorOutcome("ViewDefinition not found: "+id))
	}
	return c.JSON(http.StatusOK, view)
}

// Create registers a new view definition.
func (h *ViewDefinitionHandler) Create(c echo.Context) error {
	var view ViewDefinition
	if err := c.Bind(&view); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid request body: "+err.Error()))
	}
	if view.Resource == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("resource field is required"))
	}
	if len(view.Select) == 0 {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("at least one column is required"))
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.views[view.ID] = &view
	return c.JSON(http.StatusCreated, view)
}

// Update replaces an existing view definition.
func (h *ViewDefinitionHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var view ViewDefinition
	if err := c.Bind(&view); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid request body: "+err.Error()))
	}
	view.ID = id
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.views[id]; !ok {
		return c.JSON(http.StatusNotFound, ErrorOutcome("ViewDefinition not found: "+id))
	}
	h.views[id] = &view
	return c.JSON(http.StatusOK, view)
}

// Delete removes a view definition.
func (h *ViewDefinitionHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.views[id]; !ok {
		return c.JSON(http.StatusNotFound, ErrorOutcome("ViewDefinition not found: "+id))
	}
	delete(h.views, id)
	return c.NoContent(http.StatusNoContent)
}

// ExecuteView runs a ViewDefinition against provided resources.
func (h *ViewDefinitionHandler) ExecuteView(c echo.Context) error {
	id := c.Param("id")
	format := c.QueryParam("_format")
	if format == "" {
		format = "json"
	}
	countParam := c.QueryParam("_count")
	maxRows := 0
	if countParam != "" {
		if n, err := strconv.Atoi(countParam); err == nil && n > 0 {
			maxRows = n
		}
	}
	h.mu.RLock()
	view, ok := h.views[id]
	h.mu.RUnlock()
	if !ok {
		return c.JSON(http.StatusNotFound, ErrorOutcome("ViewDefinition not found: "+id))
	}
	var resources []map[string]interface{}
	if c.Request().Body != nil {
		if err := json.NewDecoder(c.Request().Body).Decode(&resources); err != nil {
			resources = nil
		}
	}
	result, err := h.engine.Execute(c.Request().Context(), view, resources)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorOutcome("execution failed: "+err.Error()))
	}
	if maxRows > 0 && len(result.Rows) > maxRows {
		result.Rows = result.Rows[:maxRows]
	}
	switch format {
	case "csv":
		csvData := h.engine.ToCSV(result)
		return c.Blob(http.StatusOK, "text/csv", []byte(csvData))
	case "ndjson":
		ndjson := h.engine.ToNDJSON(result)
		return c.Blob(http.StatusOK, "application/x-ndjson", []byte(ndjson))
	default:
		return c.JSON(http.StatusOK, h.engine.ToJSON(result))
	}
}

// SQLView returns the generated PostgreSQL CREATE VIEW statement.
func (h *ViewDefinitionHandler) SQLView(c echo.Context) error {
	id := c.Param("id")
	h.mu.RLock()
	view, ok := h.views[id]
	h.mu.RUnlock()
	if !ok {
		return c.JSON(http.StatusNotFound, ErrorOutcome("ViewDefinition not found: "+id))
	}
	sqlStmt := h.engine.GenerateSQL(view)
	return c.String(http.StatusOK, sqlStmt)
}
