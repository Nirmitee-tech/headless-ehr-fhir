package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

// GraphQLResourceResolver resolves FHIR resources for the GraphQL engine.
// It supports fetching by ID and searching with parameters.
type GraphQLResourceResolver interface {
	ResolveByID(ctx context.Context, resourceType, id string) (map[string]interface{}, error)
	ResolveSearch(ctx context.Context, resourceType string, params map[string]string, limit int) ([]map[string]interface{}, error)
}

// GraphQLRequest represents an incoming GraphQL query request.
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents the response from a GraphQL query execution.
type GraphQLResponse struct {
	Data   interface{}    `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLError represents a single error from GraphQL execution.
type GraphQLError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

// parsedQuery holds the result of parsing a simple GraphQL query string.
type parsedQuery struct {
	ResourceType string
	ID           string
	Params       map[string]string
	Fields       []string
	IsList       bool
}

// GraphQLEngine executes simplified FHIR-compatible GraphQL queries.
type GraphQLEngine struct {
	mu        sync.RWMutex
	resolvers map[string]GraphQLResourceResolver
}

// NewGraphQLEngine creates a new GraphQL engine with no resolvers registered.
func NewGraphQLEngine() *GraphQLEngine {
	return &GraphQLEngine{
		resolvers: make(map[string]GraphQLResourceResolver),
	}
}

// RegisterResolver registers a resolver for the given FHIR resource type.
func (e *GraphQLEngine) RegisterResolver(resourceType string, resolver GraphQLResourceResolver) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.resolvers[resourceType] = resolver
}

// Execute parses and executes a GraphQL request, returning a response.
func (e *GraphQLEngine) Execute(ctx context.Context, req GraphQLRequest) *GraphQLResponse {
	query := req.Query

	// Substitute variables into the query.
	if len(req.Variables) > 0 {
		query = substituteVariables(query, req.Variables)
	}

	parsed, err := parseGraphQLQuery(query)
	if err != nil {
		return &GraphQLResponse{
			Errors: []GraphQLError{{Message: err.Error()}},
		}
	}

	e.mu.RLock()
	resolver, ok := e.resolvers[parsed.ResourceType]
	e.mu.RUnlock()

	if !ok {
		return &GraphQLResponse{
			Errors: []GraphQLError{{
				Message: fmt.Sprintf("no resolver registered for resource type %s", parsed.ResourceType),
				Path:    []string{parsed.ResourceType},
			}},
		}
	}

	if parsed.IsList {
		return e.executeList(ctx, resolver, parsed)
	}
	return e.executeSingle(ctx, resolver, parsed)
}

// executeSingle resolves a single resource by ID and applies field selection.
func (e *GraphQLEngine) executeSingle(ctx context.Context, resolver GraphQLResourceResolver, pq *parsedQuery) *GraphQLResponse {
	resource, err := resolver.ResolveByID(ctx, pq.ResourceType, pq.ID)
	if err != nil {
		return &GraphQLResponse{
			Errors: []GraphQLError{{
				Message: fmt.Sprintf("resource %s/%s not found: %v", pq.ResourceType, pq.ID, err),
				Path:    []string{pq.ResourceType},
			}},
		}
	}

	filtered := applyFieldSelection(resource, pq.Fields)
	queryName := pq.ResourceType
	return &GraphQLResponse{
		Data: map[string]interface{}{
			queryName: filtered,
		},
	}
}

// executeList resolves a list of resources via search and applies field selection.
func (e *GraphQLEngine) executeList(ctx context.Context, resolver GraphQLResourceResolver, pq *parsedQuery) *GraphQLResponse {
	limit := 100 // default
	params := make(map[string]string)
	for k, v := range pq.Params {
		if k == "_count" {
			if n, err := strconv.Atoi(v); err == nil {
				limit = n
			}
			continue
		}
		params[k] = v
	}

	results, err := resolver.ResolveSearch(ctx, pq.ResourceType, params, limit)
	if err != nil {
		return &GraphQLResponse{
			Errors: []GraphQLError{{
				Message: fmt.Sprintf("search for %s failed: %v", pq.ResourceType, err),
				Path:    []string{pq.ResourceType + "List"},
			}},
		}
	}

	filtered := make([]interface{}, 0, len(results))
	for _, r := range results {
		filtered = append(filtered, applyFieldSelection(r, pq.Fields))
	}

	queryName := pq.ResourceType + "List"
	return &GraphQLResponse{
		Data: map[string]interface{}{
			queryName: filtered,
		},
	}
}

// applyFieldSelection filters a resource map to include only the requested
// fields. Nested object fields are preserved as-is.
func applyFieldSelection(resource map[string]interface{}, fields []string) map[string]interface{} {
	if len(fields) == 0 {
		return resource
	}
	result := make(map[string]interface{}, len(fields))
	for _, field := range fields {
		if val, ok := resource[field]; ok {
			result[field] = val
		}
	}
	return result
}

// substituteVariables replaces $varName references in the query with variable
// values.
func substituteVariables(query string, variables map[string]interface{}) string {
	for name, val := range variables {
		var strVal string
		switch v := val.(type) {
		case string:
			strVal = `"` + v + `"`
		default:
			strVal = fmt.Sprintf("%v", v)
		}
		query = strings.ReplaceAll(query, "$"+name, strVal)
	}
	return query
}

// graphqlQueryPattern matches simple GraphQL queries:
//
//	{ TypeName(args) { fields } }
var graphqlQueryPattern = regexp.MustCompile(
	`^\s*\{\s*(\w+)\s*(?:\(([^)]*)\))?\s*\{\s*([^}]+)\}\s*\}\s*$`,
)

// parseGraphQLQuery parses a simplified GraphQL query string into a structured
// representation.
func parseGraphQLQuery(query string) (*parsedQuery, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("empty query")
	}

	matches := graphqlQueryPattern.FindStringSubmatch(query)
	if matches == nil {
		return nil, fmt.Errorf("invalid GraphQL query syntax")
	}

	typeName := matches[1]
	argsStr := matches[2]
	fieldsStr := matches[3]

	pq := &parsedQuery{
		Params: make(map[string]string),
	}

	// Check for List suffix.
	if strings.HasSuffix(typeName, "List") {
		pq.ResourceType = strings.TrimSuffix(typeName, "List")
		pq.IsList = true
	} else {
		pq.ResourceType = typeName
	}

	// Parse arguments.
	if argsStr != "" {
		if err := parseArguments(argsStr, pq); err != nil {
			return nil, err
		}
	}

	// Parse fields.
	pq.Fields = parseFields(fieldsStr)

	return pq, nil
}

// argPattern matches key: "value" or key: $var pairs in argument lists.
var argPattern = regexp.MustCompile(`(\w+)\s*:\s*(?:"([^"]*)"|([\w$]+))`)

// parseArguments extracts key-value pairs from the GraphQL arguments string.
func parseArguments(argsStr string, pq *parsedQuery) error {
	matches := argPattern.FindAllStringSubmatch(argsStr, -1)
	for _, m := range matches {
		key := m[1]
		value := m[2]
		if value == "" {
			value = m[3]
		}
		if key == "id" {
			pq.ID = value
		}
		pq.Params[key] = value
	}
	return nil
}

// parseFields splits a comma-separated field list into individual field names.
func parseFields(fieldsStr string) []string {
	parts := strings.Split(fieldsStr, ",")
	fields := make([]string, 0, len(parts))
	for _, p := range parts {
		f := strings.TrimSpace(p)
		if f != "" {
			fields = append(fields, f)
		}
	}
	return fields
}

// =========== HTTP Handler ===========

// GraphQLHandler provides HTTP endpoints for the FHIR $graphql operation.
type GraphQLHandler struct {
	engine *GraphQLEngine
}

// NewGraphQLHandler creates a new GraphQL HTTP handler.
func NewGraphQLHandler(engine *GraphQLEngine) *GraphQLHandler {
	return &GraphQLHandler{engine: engine}
}

// RegisterRoutes adds $graphql routes to the given FHIR group.
func (h *GraphQLHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/$graphql", h.HandlePost)
	g.GET("/$graphql", h.HandleGet)
}

// HandlePost handles POST /fhir/$graphql with a JSON GraphQLRequest body.
func (h *GraphQLHandler) HandlePost(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "structure", "Failed to read request body"))
	}

	if len(body) == 0 {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Request body is required"))
	}

	var req GraphQLRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "structure", "Invalid JSON: "+err.Error()))
	}

	if strings.TrimSpace(req.Query) == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Query is required"))
	}

	resp := h.engine.Execute(c.Request().Context(), req)
	return c.JSON(http.StatusOK, resp)
}

// HandleGet handles GET /fhir/$graphql with the query in a query parameter.
func (h *GraphQLHandler) HandleGet(c echo.Context) error {
	query := c.QueryParam("query")
	if strings.TrimSpace(query) == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Query parameter 'query' is required"))
	}

	req := GraphQLRequest{Query: query}
	resp := h.engine.Execute(c.Request().Context(), req)
	return c.JSON(http.StatusOK, resp)
}

// =========== In-Memory Test Resolver ===========

// InMemoryResourceResolver is a simple in-memory implementation of
// GraphQLResourceResolver for testing purposes.
type InMemoryResourceResolver struct {
	mu        sync.RWMutex
	resources map[string]map[string]map[string]interface{} // [type][id]resource
}

// NewInMemoryResourceResolver creates a new empty in-memory resolver.
func NewInMemoryResourceResolver() *InMemoryResourceResolver {
	return &InMemoryResourceResolver{
		resources: make(map[string]map[string]map[string]interface{}),
	}
}

// AddResource stores a resource in the in-memory store. The resource must
// contain an "id" field.
func (r *InMemoryResourceResolver) AddResource(resourceType string, resource map[string]interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.resources[resourceType] == nil {
		r.resources[resourceType] = make(map[string]map[string]interface{})
	}

	id, _ := resource["id"].(string)
	if id != "" {
		r.resources[resourceType][id] = resource
	}
}

// ResolveByID retrieves a resource by type and ID.
func (r *InMemoryResourceResolver) ResolveByID(_ context.Context, resourceType, id string) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	byType, ok := r.resources[resourceType]
	if !ok {
		return nil, fmt.Errorf("resource %s/%s not found", resourceType, id)
	}

	resource, ok := byType[id]
	if !ok {
		return nil, fmt.Errorf("resource %s/%s not found", resourceType, id)
	}

	return resource, nil
}

// ResolveSearch returns resources matching the given search parameters. It
// performs simple string matching on top-level fields and nested name fields.
func (r *InMemoryResourceResolver) ResolveSearch(_ context.Context, resourceType string, params map[string]string, limit int) ([]map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	byType, ok := r.resources[resourceType]
	if !ok {
		return nil, nil
	}

	var results []map[string]interface{}
	for _, resource := range byType {
		if matchesParams(resource, params) {
			results = append(results, resource)
			if len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}

// matchesParams checks whether a resource matches all the given search
// parameters using simple string matching.
func matchesParams(resource map[string]interface{}, params map[string]string) bool {
	for key, val := range params {
		if key == "name" {
			if !matchesName(resource, val) {
				return false
			}
			continue
		}
		fieldVal, ok := resource[key]
		if !ok {
			return false
		}
		if fmt.Sprintf("%v", fieldVal) != val {
			return false
		}
	}
	return true
}

// matchesName checks if a resource's name field contains the given search
// string (matching against family or given names).
func matchesName(resource map[string]interface{}, search string) bool {
	names, ok := resource["name"].([]interface{})
	if !ok {
		return false
	}
	for _, n := range names {
		nameObj, ok := n.(map[string]interface{})
		if !ok {
			continue
		}
		if family, ok := nameObj["family"].(string); ok {
			if strings.Contains(strings.ToLower(family), strings.ToLower(search)) {
				return true
			}
		}
		if given, ok := nameObj["given"].([]interface{}); ok {
			for _, g := range given {
				if gs, ok := g.(string); ok {
					if strings.Contains(strings.ToLower(gs), strings.ToLower(search)) {
						return true
					}
				}
			}
		}
	}
	return false
}
