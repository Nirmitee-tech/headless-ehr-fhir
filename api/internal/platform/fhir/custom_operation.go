package fhir

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// OperationScope — bitmask for where a custom operation can be invoked
// ---------------------------------------------------------------------------

// OperationScope defines where an operation can be invoked.
type OperationScope int

const (
	// OperationScopeSystem indicates the operation is available at the system level: POST /fhir/$operation
	OperationScopeSystem OperationScope = 1 << iota
	// OperationScopeType indicates the operation is available at the type level: POST /fhir/Patient/$operation
	OperationScopeType
	// OperationScopeInstance indicates the operation is available at the instance level: POST /fhir/Patient/123/$operation
	OperationScopeInstance
)

// maxOperationCodeLength is the maximum allowed length for an operation code.
const maxOperationCodeLength = 255

// ---------------------------------------------------------------------------
// OperationParamDef — describes a parameter for a custom operation
// ---------------------------------------------------------------------------

// OperationParamDef describes a parameter for a custom operation.
type OperationParamDef struct {
	Name          string `json:"name"`
	Use           string `json:"use"`      // in | out
	Min           int    `json:"min"`
	Max           string `json:"max"`      // "1" or "*"
	Type          string `json:"type"`     // FHIR type (string, Reference, Resource, etc.)
	Required      bool   `json:"required"`
	Documentation string `json:"documentation"`
}

// ---------------------------------------------------------------------------
// CustomOperationDef — defines a custom FHIR operation
// ---------------------------------------------------------------------------

// CustomOperationDef defines a custom FHIR operation.
type CustomOperationDef struct {
	Name          string              `json:"name"`          // e.g., "$my-operation" (with $)
	Code          string              `json:"code"`          // e.g., "my-operation" (without $)
	Title         string              `json:"title"`
	Description   string              `json:"description"`
	Scope         OperationScope      `json:"scope"`
	ResourceTypes []string            `json:"resourceTypes"` // Which resource types (empty = all)
	Parameters    []OperationParamDef `json:"parameters"`
	AffectsState  bool                `json:"affectsState"` // If true, only POST allowed
	Idempotent    bool                `json:"idempotent"`
	System        bool                `json:"system"`   // Available at system level
	Type          bool                `json:"type"`     // Available at type level
	Instance      bool                `json:"instance"` // Available at instance level
	CreatedAt     time.Time           `json:"createdAt"`
	UpdatedAt     time.Time           `json:"updatedAt"`
}

// ---------------------------------------------------------------------------
// OperationInvocation — represents a single invocation of a custom operation
// ---------------------------------------------------------------------------

// OperationInvocation represents a single invocation of a custom operation.
type OperationInvocation struct {
	OperationCode string
	Scope         OperationScope
	ResourceType  string                 // empty for system-level
	ResourceID    string                 // empty for type/system level
	Parameters    map[string]interface{} // Input parameters
	RequestBody   map[string]interface{} // Raw request body (Parameters resource)
}

// ---------------------------------------------------------------------------
// OperationHandler — function signature for custom operation handlers
// ---------------------------------------------------------------------------

// OperationHandler is the function signature for custom operation handlers.
type OperationHandler func(ctx *OperationContext) (*OperationResponse, error)

// ---------------------------------------------------------------------------
// OperationContext — provides context to operation handlers
// ---------------------------------------------------------------------------

// OperationContext provides context to operation handlers.
type OperationContext struct {
	Invocation *OperationInvocation
	RequestID  string
	TenantID   string
	UserID     string
}

// ---------------------------------------------------------------------------
// OperationResponse — what operation handlers return
// ---------------------------------------------------------------------------

// OperationResponse is what operation handlers return.
type OperationResponse struct {
	StatusCode  int
	Resource    map[string]interface{} // Response resource (Parameters, Bundle, etc.)
	ContentType string
}

// ---------------------------------------------------------------------------
// CustomOperationRegistry — manages custom operations
// ---------------------------------------------------------------------------

// CustomOperationRegistry manages custom operations with thread-safe
// registration, unregistration, lookup, and invocation.
type CustomOperationRegistry struct {
	mu         sync.RWMutex
	operations map[string]*CustomOperationDef
	handlers   map[string]OperationHandler
}

// NewCustomOperationRegistry creates a new empty CustomOperationRegistry.
func NewCustomOperationRegistry() *CustomOperationRegistry {
	return &CustomOperationRegistry{
		operations: make(map[string]*CustomOperationDef),
		handlers:   make(map[string]OperationHandler),
	}
}

// Register adds a custom operation with its handler to the registry.
// Returns an error if the definition is nil, the handler is nil, the
// definition is invalid, or an operation with the same code is already
// registered.
func (r *CustomOperationRegistry) Register(def *CustomOperationDef, handler OperationHandler) error {
	if def == nil {
		return fmt.Errorf("operation definition must not be nil")
	}
	if handler == nil {
		return fmt.Errorf("operation handler must not be nil")
	}

	issues := ValidateOperationDef(def)
	for _, issue := range issues {
		if issue.Severity == SeverityError || issue.Severity == SeverityFatal {
			return fmt.Errorf("invalid operation definition: %s", issue.Diagnostics)
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.operations[def.Code]; exists {
		return fmt.Errorf("operation %q is already registered", def.Code)
	}

	r.operations[def.Code] = def
	r.handlers[def.Code] = handler
	return nil
}

// Unregister removes a custom operation by code.
func (r *CustomOperationRegistry) Unregister(code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.operations[code]; !exists {
		return fmt.Errorf("operation %q not found", code)
	}

	delete(r.operations, code)
	delete(r.handlers, code)
	return nil
}

// Get returns an operation definition by code.
func (r *CustomOperationRegistry) Get(code string) (*CustomOperationDef, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	def, ok := r.operations[code]
	return def, ok
}

// List returns all registered operations sorted alphabetically by code.
func (r *CustomOperationRegistry) List() []*CustomOperationDef {
	r.mu.RLock()
	defer r.mu.RUnlock()

	codes := make([]string, 0, len(r.operations))
	for code := range r.operations {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	result := make([]*CustomOperationDef, 0, len(codes))
	for _, code := range codes {
		result = append(result, r.operations[code])
	}
	return result
}

// ListForResourceType returns operations available for a specific resource type.
// Operations with an empty ResourceTypes list are considered available for all types.
func (r *CustomOperationRegistry) ListForResourceType(resourceType string) []*CustomOperationDef {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*CustomOperationDef
	for _, def := range r.operations {
		if len(def.ResourceTypes) == 0 {
			result = append(result, def)
			continue
		}
		for _, rt := range def.ResourceTypes {
			if strings.EqualFold(rt, resourceType) {
				result = append(result, def)
				break
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Code < result[j].Code
	})
	return result
}

// ListByScope returns operations matching the given scope bitmask.
func (r *CustomOperationRegistry) ListByScope(scope OperationScope) []*CustomOperationDef {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*CustomOperationDef
	for _, def := range r.operations {
		if def.Scope&scope != 0 {
			result = append(result, def)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Code < result[j].Code
	})
	return result
}

// ---------------------------------------------------------------------------
// ValidateOperationDef — validates a custom operation definition
// ---------------------------------------------------------------------------

// ValidateOperationDef validates a CustomOperationDef and returns any issues found.
func ValidateOperationDef(def *CustomOperationDef) []ValidationIssue {
	var issues []ValidationIssue

	if def == nil {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityFatal,
			Code:        VIssueTypeStructure,
			Diagnostics: "Operation definition is nil",
		})
		return issues
	}

	// Validate Name
	if def.Name == "" {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Location:    "OperationDefinition.name",
			Diagnostics: "Operation name is required",
		})
	} else if !strings.HasPrefix(def.Name, "$") {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeValue,
			Location:    "OperationDefinition.name",
			Diagnostics: "Operation name must start with $ prefix",
		})
	}

	// Validate Code
	if def.Code == "" {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Location:    "OperationDefinition.code",
			Diagnostics: "Operation code is required",
		})
	} else {
		if strings.HasPrefix(def.Code, "$") {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeValue,
				Location:    "OperationDefinition.code",
				Diagnostics: "Operation code must not contain $ prefix (use code without $)",
			})
		}
		if len(def.Code) > maxOperationCodeLength {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeValue,
				Location:    "OperationDefinition.code",
				Diagnostics: fmt.Sprintf("Operation code length must not exceed %d characters", maxOperationCodeLength),
			})
		}
	}

	// Validate Name length
	if len(def.Name) > maxOperationCodeLength+1 { // +1 for the $ prefix
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeValue,
			Location:    "OperationDefinition.name",
			Diagnostics: fmt.Sprintf("Operation name length must not exceed %d characters", maxOperationCodeLength+1),
		})
	}

	// Validate Parameters
	paramNames := make(map[string]bool)
	for i, param := range def.Parameters {
		loc := fmt.Sprintf("OperationDefinition.parameter[%d]", i)

		if param.Name == "" {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeRequired,
				Location:    loc + ".name",
				Diagnostics: "Parameter name is required",
			})
		} else if paramNames[param.Name] {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityWarning,
				Code:        VIssueTypeValue,
				Location:    loc + ".name",
				Diagnostics: fmt.Sprintf("Duplicate parameter name %q", param.Name),
			})
		}
		paramNames[param.Name] = true

		if param.Use != "in" && param.Use != "out" {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeValue,
				Location:    loc + ".use",
				Diagnostics: fmt.Sprintf("Parameter use must be 'in' or 'out', got %q", param.Use),
			})
		}

		if param.Type == "" {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeRequired,
				Location:    loc + ".type",
				Diagnostics: "Parameter type is required",
			})
		}

		if param.Max != "*" && param.Max != "" {
			if _, err := strconv.Atoi(param.Max); err != nil {
				issues = append(issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeValue,
					Location:    loc + ".max",
					Diagnostics: fmt.Sprintf("Parameter max must be a number or '*', got %q", param.Max),
				})
			}
		}
	}

	return issues
}

// ---------------------------------------------------------------------------
// ParseOperationParameters — parses a FHIR Parameters resource into a map
// ---------------------------------------------------------------------------

// valueKeys is the ordered list of FHIR Parameters value[x] keys to check.
var valueKeys = []string{
	"valueString",
	"valueBoolean",
	"valueInteger",
	"valueDecimal",
	"valueUri",
	"valueUrl",
	"valueCode",
	"valueDate",
	"valueDateTime",
	"valueInstant",
	"valueTime",
	"valueCoding",
	"valueCodeableConcept",
	"valueQuantity",
	"valueRange",
	"valuePeriod",
	"valueReference",
	"valueIdentifier",
	"valueAttachment",
	"resource",
}

// ParseOperationParameters parses a FHIR Parameters resource body into a simple
// parameter name -> value map.
func ParseOperationParameters(body map[string]interface{}) (map[string]interface{}, error) {
	if body == nil {
		return nil, fmt.Errorf("request body is nil")
	}

	rt, _ := body["resourceType"].(string)
	if rt != "Parameters" {
		return nil, fmt.Errorf("expected resourceType 'Parameters', got %q", rt)
	}

	result := make(map[string]interface{})

	paramList, ok := body["parameter"]
	if !ok {
		return result, nil
	}

	params, ok := paramList.([]interface{})
	if !ok {
		return nil, fmt.Errorf("parameter field must be an array")
	}

	for _, item := range params {
		p, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := p["name"].(string)
		if name == "" {
			continue
		}

		// Look for value[x] fields
		var value interface{}
		for _, vk := range valueKeys {
			if v, exists := p[vk]; exists {
				value = v
				break
			}
		}

		if value != nil {
			result[name] = value
		}
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// BuildParametersResource — creates a FHIR Parameters resource from a map
// ---------------------------------------------------------------------------

// BuildParametersResource creates a FHIR Parameters resource from a map of
// parameter names to values.
func BuildParametersResource(params map[string]interface{}) map[string]interface{} {
	paramList := make([]interface{}, 0)

	if params != nil {
		// Sort keys for deterministic output
		keys := make([]string, 0, len(params))
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, name := range keys {
			value := params[name]
			entry := map[string]interface{}{
				"name": name,
			}

			switch v := value.(type) {
			case string:
				entry["valueString"] = v
			case bool:
				entry["valueBoolean"] = v
			case float64:
				// Check if it's an integer
				if v == float64(int64(v)) {
					entry["valueInteger"] = v
				} else {
					entry["valueDecimal"] = v
				}
			case int:
				entry["valueInteger"] = v
			case int64:
				entry["valueInteger"] = v
			case map[string]interface{}:
				// Map values are treated as resource parameters
				if _, hasRT := v["resourceType"]; hasRT {
					entry["resource"] = v
				} else {
					entry["resource"] = v
				}
			default:
				entry["valueString"] = fmt.Sprintf("%v", v)
			}

			paramList = append(paramList, entry)
		}
	}

	return map[string]interface{}{
		"resourceType": "Parameters",
		"parameter":    paramList,
	}
}

// ---------------------------------------------------------------------------
// ValidateOperationInput — validates input parameters against definition
// ---------------------------------------------------------------------------

// ValidateOperationInput validates input parameters against the operation
// definition. It checks required parameters and flags unknown parameters.
func ValidateOperationInput(def *CustomOperationDef, params map[string]interface{}) []ValidationIssue {
	var issues []ValidationIssue

	if params == nil {
		params = make(map[string]interface{})
	}

	// Collect input parameter definitions
	inputParams := make(map[string]*OperationParamDef)
	for i := range def.Parameters {
		p := &def.Parameters[i]
		if p.Use == "in" {
			inputParams[p.Name] = p
		}
	}

	// Check required parameters
	for name, paramDef := range inputParams {
		if (paramDef.Required || paramDef.Min > 0) {
			if _, ok := params[name]; !ok {
				issues = append(issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeRequired,
					Location:    "Parameters.parameter:" + name,
					Diagnostics: fmt.Sprintf("Required input parameter '%s' is missing", name),
				})
			}
		}
	}

	// Check for unknown parameters
	for name := range params {
		if _, known := inputParams[name]; !known {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityWarning,
				Code:        VIssueTypeValue,
				Location:    "Parameters.parameter:" + name,
				Diagnostics: fmt.Sprintf("Unknown input parameter '%s'", name),
			})
		}
	}

	return issues
}

// ---------------------------------------------------------------------------
// ValidateOperationOutput — validates output parameters against definition
// ---------------------------------------------------------------------------

// ValidateOperationOutput validates output parameters against the operation
// definition. It checks that required output parameters are present.
func ValidateOperationOutput(def *CustomOperationDef, result map[string]interface{}) []ValidationIssue {
	var issues []ValidationIssue

	if result == nil {
		result = make(map[string]interface{})
	}

	for _, paramDef := range def.Parameters {
		if paramDef.Use != "out" {
			continue
		}
		if paramDef.Min > 0 {
			if _, ok := result[paramDef.Name]; !ok {
				issues = append(issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeRequired,
					Location:    "Parameters.parameter:" + paramDef.Name,
					Diagnostics: fmt.Sprintf("Required output parameter '%s' is missing", paramDef.Name),
				})
			}
		}
	}

	return issues
}

// ---------------------------------------------------------------------------
// RouteOperation — determines which operation to invoke based on path
// ---------------------------------------------------------------------------

// RouteOperation determines which custom operation to invoke based on the HTTP
// method and request path. It returns the matched definition, an invocation
// descriptor, and any error.
//
// Supported path patterns:
//   - /fhir/$code                    (system-level)
//   - /fhir/{ResourceType}/$code     (type-level)
//   - /fhir/{ResourceType}/{id}/$code (instance-level)
func (r *CustomOperationRegistry) RouteOperation(method, path string) (*CustomOperationDef, *OperationInvocation, error) {
	if path == "" {
		return nil, nil, fmt.Errorf("empty request path")
	}

	// Find the $ segment
	dollarIdx := strings.LastIndex(path, "/$")
	if dollarIdx < 0 {
		return nil, nil, fmt.Errorf("path does not contain an operation invocation (no $)")
	}

	opCode := path[dollarIdx+2:] // everything after "/$"
	prefix := path[:dollarIdx]    // everything before "/$"

	// Determine scope from path structure
	var scope OperationScope
	var resourceType, resourceID string

	// Strip /fhir prefix if present
	prefix = strings.TrimPrefix(prefix, "/fhir")
	prefix = strings.TrimPrefix(prefix, "/")

	parts := strings.Split(prefix, "/")
	// Filter empty parts
	var segments []string
	for _, p := range parts {
		if p != "" {
			segments = append(segments, p)
		}
	}

	switch len(segments) {
	case 0:
		scope = OperationScopeSystem
	case 1:
		scope = OperationScopeType
		resourceType = segments[0]
	case 2:
		scope = OperationScopeInstance
		resourceType = segments[0]
		resourceID = segments[1]
	default:
		return nil, nil, fmt.Errorf("invalid operation path: too many segments")
	}

	// Look up the operation
	r.mu.RLock()
	def, exists := r.operations[opCode]
	r.mu.RUnlock()

	if !exists {
		return nil, nil, fmt.Errorf("operation $%s not found", opCode)
	}

	// Verify scope is supported
	if def.Scope&scope == 0 {
		return nil, nil, fmt.Errorf("operation $%s does not support %s scope", opCode, scopeName(scope))
	}

	// Verify resource type if specified
	if resourceType != "" && len(def.ResourceTypes) > 0 {
		matched := false
		for _, rt := range def.ResourceTypes {
			if strings.EqualFold(rt, resourceType) {
				matched = true
				break
			}
		}
		if !matched {
			return nil, nil, fmt.Errorf("operation $%s is not available for resource type %s", opCode, resourceType)
		}
	}

	// Verify HTTP method for AffectsState
	if def.AffectsState && method != http.MethodPost {
		return nil, nil, fmt.Errorf("operation $%s affects state and requires POST method", opCode)
	}

	inv := &OperationInvocation{
		OperationCode: opCode,
		Scope:         scope,
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		Parameters:    make(map[string]interface{}),
	}

	return def, inv, nil
}

// scopeName returns a human-readable name for an OperationScope.
func scopeName(s OperationScope) string {
	switch s {
	case OperationScopeSystem:
		return "system"
	case OperationScopeType:
		return "type"
	case OperationScopeInstance:
		return "instance"
	default:
		return "unknown"
	}
}

// ---------------------------------------------------------------------------
// CustomOperationMiddleware — middleware for automatic operation routing
// ---------------------------------------------------------------------------

// CustomOperationMiddleware returns Echo middleware that intercepts requests
// whose path contains a $ operation invocation and routes them through the
// custom operation registry. Non-operation requests pass through unchanged.
func CustomOperationMiddleware(registry *CustomOperationRegistry) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if !strings.Contains(path, "/$") {
				return next(c)
			}

			handler := CustomOperationHandler(registry)
			return handler(c)
		}
	}
}

// ---------------------------------------------------------------------------
// CustomOperationHandler — HTTP handler for custom operation endpoints
// ---------------------------------------------------------------------------

// CustomOperationHandler returns an Echo handler that dispatches custom
// operation requests. It parses the path, looks up the operation, validates
// parameters, calls the handler, and returns the response.
func CustomOperationHandler(registry *CustomOperationRegistry) echo.HandlerFunc {
	return func(c echo.Context) error {
		method := c.Request().Method
		path := c.Request().URL.Path

		// Route the operation
		def, inv, err := registry.RouteOperation(method, path)
		if err != nil {
			msg := err.Error()

			if strings.Contains(msg, "not found") {
				return c.JSON(http.StatusNotFound, ErrorOutcome(msg))
			}
			if strings.Contains(msg, "POST method") {
				return c.JSON(http.StatusMethodNotAllowed, ErrorOutcome(msg))
			}
			if strings.Contains(msg, "does not support") || strings.Contains(msg, "not available") {
				return c.JSON(http.StatusBadRequest, ErrorOutcome(msg))
			}
			return c.JSON(http.StatusBadRequest, ErrorOutcome(msg))
		}

		// Parse parameters based on method
		var params map[string]interface{}
		var requestBody map[string]interface{}

		if method == http.MethodGet {
			// Parse query parameters
			params = make(map[string]interface{})
			for key, values := range c.QueryParams() {
				if len(values) == 1 {
					params[key] = values[0]
				} else {
					iface := make([]interface{}, len(values))
					for i, v := range values {
						iface[i] = v
					}
					params[key] = iface
				}
			}
		} else {
			// Parse request body (POST)
			body, readErr := io.ReadAll(c.Request().Body)
			if readErr != nil {
				return c.JSON(http.StatusBadRequest, ErrorOutcome("failed to read request body"))
			}

			if len(body) > 0 {
				if jsonErr := json.Unmarshal(body, &requestBody); jsonErr != nil {
					return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid JSON: "+jsonErr.Error()))
				}

				// Check if it's a Parameters resource
				rt, _ := requestBody["resourceType"].(string)
				if rt == "Parameters" {
					params, err = ParseOperationParameters(requestBody)
					if err != nil {
						return c.JSON(http.StatusBadRequest, ErrorOutcome("failed to parse parameters: "+err.Error()))
					}
				} else {
					params = make(map[string]interface{})
					requestBody = requestBody
				}
			} else {
				params = make(map[string]interface{})
			}
		}

		inv.Parameters = params
		inv.RequestBody = requestBody

		// Look up the handler
		registry.mu.RLock()
		handler, exists := registry.handlers[def.Code]
		registry.mu.RUnlock()

		if !exists {
			return c.JSON(http.StatusInternalServerError, ErrorOutcome("handler not found for operation $"+def.Code))
		}

		// Build operation context
		opCtx := &OperationContext{
			Invocation: inv,
			RequestID:  c.Request().Header.Get("X-Request-ID"),
			TenantID:   c.Request().Header.Get("X-Tenant-ID"),
			UserID:     c.Request().Header.Get("X-User-ID"),
		}

		// Execute the handler
		resp, handlerErr := handler(opCtx)
		if handlerErr != nil {
			return c.JSON(http.StatusInternalServerError, ErrorOutcome("operation $"+def.Code+" failed: "+handlerErr.Error()))
		}

		if resp == nil {
			return c.JSON(http.StatusInternalServerError, ErrorOutcome("operation $"+def.Code+" returned nil response"))
		}

		// Set content type
		contentType := resp.ContentType
		if contentType == "" {
			contentType = "application/fhir+json"
		}
		c.Response().Header().Set(echo.HeaderContentType, contentType)

		// Set idempotent cache hint header
		if def.Idempotent {
			c.Response().Header().Set("X-Idempotent", "true")
		}

		statusCode := resp.StatusCode
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		return c.JSON(statusCode, resp.Resource)
	}
}

// ---------------------------------------------------------------------------
// ToOperationDefinition — converts to FHIR OperationDefinition resource
// ---------------------------------------------------------------------------

// ToOperationDefinition converts a CustomOperationDef to a FHIR
// OperationDefinition resource map suitable for inclusion in a
// CapabilityStatement or direct serialization.
func (def *CustomOperationDef) ToOperationDefinition() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "OperationDefinition",
		"id":           def.Code,
		"name":         def.Name,
		"title":        def.Title,
		"status":       "active",
		"kind":         "operation",
		"code":         def.Code,
		"system":       def.System,
		"type":         def.Type,
		"instance":     def.Instance,
		"affectsState": def.AffectsState,
		"idempotent":   def.Idempotent,
	}

	if def.Description != "" {
		result["description"] = def.Description
	}

	if len(def.ResourceTypes) > 0 {
		result["resource"] = def.ResourceTypes
	}

	if len(def.Parameters) > 0 {
		params := make([]map[string]interface{}, 0, len(def.Parameters))
		for _, p := range def.Parameters {
			param := map[string]interface{}{
				"name": p.Name,
				"use":  p.Use,
				"min":  p.Min,
				"max":  p.Max,
				"type": p.Type,
			}
			if p.Documentation != "" {
				param["documentation"] = p.Documentation
			}
			params = append(params, param)
		}
		result["parameter"] = params
	}

	return result
}
