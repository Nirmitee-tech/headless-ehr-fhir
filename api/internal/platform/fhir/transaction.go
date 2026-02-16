package fhir

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// BundleEntryRequest represents the request details for an entry in a
// transaction or batch Bundle, including conditional HTTP headers.
type BundleEntryRequest struct {
	Method          string `json:"method"`
	URL             string `json:"url"`
	IfNoneMatch     string `json:"ifNoneMatch,omitempty"`
	IfModifiedSince string `json:"ifModifiedSince,omitempty"`
	IfMatch         string `json:"ifMatch,omitempty"`
	IfNoneExist     string `json:"ifNoneExist,omitempty"`
}

// BundleEntryResponse represents the response details for an entry after
// a transaction or batch Bundle has been processed.
type BundleEntryResponse struct {
	Status       string      `json:"status"`
	Location     string      `json:"location,omitempty"`
	ETag         string      `json:"etag,omitempty"`
	LastModified string      `json:"lastModified,omitempty"`
	Outcome      interface{} `json:"outcome,omitempty"`
}

// TransactionEntry represents a single entry in a transaction or batch Bundle.
type TransactionEntry struct {
	FullURL  string                 `json:"fullUrl,omitempty"`
	Resource map[string]interface{} `json:"resource,omitempty"`
	Request  BundleEntryRequest     `json:"request"`
	Response *BundleEntryResponse   `json:"response,omitempty"`
}

// TransactionBundle is the parsed representation of a FHIR transaction or
// batch Bundle ready for processing.
type TransactionBundle struct {
	ResourceType string             `json:"resourceType"`
	Type         string             `json:"type"`
	Entries      []TransactionEntry `json:"entry,omitempty"`
}

// validHTTPMethods is the set of HTTP methods valid in a Bundle entry request.
var validHTTPMethods = map[string]bool{
	"GET":    true,
	"POST":   true,
	"PUT":    true,
	"DELETE": true,
	"PATCH":  true,
	"HEAD":   true,
}

// methodSortOrder defines the FHIR processing order for transaction entries.
// Per the FHIR spec: DELETE first, then POST, then PUT/PATCH, then GET/HEAD.
var methodSortOrder = map[string]int{
	"DELETE": 0,
	"POST":  1,
	"PUT":   2,
	"PATCH": 3,
	"GET":   4,
	"HEAD":  5,
}

// ParseTransactionBundle parses a raw JSON body into a TransactionBundle.
func ParseTransactionBundle(body []byte) (*TransactionBundle, error) {
	// First, parse into a generic structure to extract entries with raw resources.
	var raw struct {
		ResourceType string `json:"resourceType"`
		Type         string `json:"type"`
		Entry        []struct {
			FullURL  string              `json:"fullUrl,omitempty"`
			Resource json.RawMessage     `json:"resource,omitempty"`
			Request  *BundleEntryRequest `json:"request,omitempty"`
		} `json:"entry,omitempty"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if raw.ResourceType != "Bundle" {
		return nil, fmt.Errorf("expected resourceType Bundle, got %q", raw.ResourceType)
	}

	if raw.Type == "" {
		return nil, fmt.Errorf("bundle type is required")
	}

	bundle := &TransactionBundle{
		ResourceType: raw.ResourceType,
		Type:         raw.Type,
		Entries:      make([]TransactionEntry, 0, len(raw.Entry)),
	}

	for i, e := range raw.Entry {
		entry := TransactionEntry{
			FullURL: e.FullURL,
		}

		// Parse the resource into a generic map.
		if len(e.Resource) > 0 {
			var res map[string]interface{}
			if err := json.Unmarshal(e.Resource, &res); err != nil {
				return nil, fmt.Errorf("invalid resource in entry %d: %w", i, err)
			}
			entry.Resource = res
		}

		if e.Request != nil {
			entry.Request = *e.Request
		}

		bundle.Entries = append(bundle.Entries, entry)
	}

	return bundle, nil
}

// ValidateTransactionBundle validates the structure and content of a
// transaction or batch Bundle, returning any issues found.
func ValidateTransactionBundle(bundle *TransactionBundle) []ValidationIssue {
	var issues []ValidationIssue

	if bundle.Type != "transaction" && bundle.Type != "batch" {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeValue,
			Diagnostics: fmt.Sprintf("bundle type must be 'transaction' or 'batch', got %q", bundle.Type),
			Location:    "Bundle.type",
		})
	}

	// Track fullUrls for circular reference detection.
	fullURLSet := make(map[string]bool)

	for i, entry := range bundle.Entries {
		prefix := fmt.Sprintf("Bundle.entry[%d]", i)

		// Each entry must have a request with method and URL.
		if entry.Request.Method == "" {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeRequired,
				Diagnostics: fmt.Sprintf("entry %d: request.method is required", i),
				Location:    prefix + ".request.method",
			})
		} else if !validHTTPMethods[entry.Request.Method] {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeValue,
				Diagnostics: fmt.Sprintf("entry %d: invalid HTTP method %q", i, entry.Request.Method),
				Location:    prefix + ".request.method",
			})
		}

		if entry.Request.URL == "" {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeRequired,
				Diagnostics: fmt.Sprintf("entry %d: request.url is required", i),
				Location:    prefix + ".request.url",
			})
		}

		// Transaction entries must have a fullUrl.
		if bundle.Type == "transaction" && entry.FullURL == "" {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeRequired,
				Diagnostics: fmt.Sprintf("entry %d: fullUrl is required for transaction entries", i),
				Location:    prefix + ".fullUrl",
			})
		}

		// Check for duplicate fullUrls (potential circular references).
		if entry.FullURL != "" {
			if fullURLSet[entry.FullURL] {
				issues = append(issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeBusinessRule,
					Diagnostics: fmt.Sprintf("entry %d: duplicate fullUrl %q detected (circular reference)", i, entry.FullURL),
					Location:    prefix + ".fullUrl",
				})
			}
			fullURLSet[entry.FullURL] = true
		}
	}

	// Check for circular references: resources referencing each other in a cycle.
	issues = append(issues, detectCircularReferences(bundle.Entries)...)

	return issues
}

// detectCircularReferences examines resource references among entries and
// reports any cycles. A cycle exists if entry A references entry B and B
// references A (directly or transitively).
func detectCircularReferences(entries []TransactionEntry) []ValidationIssue {
	// Build an adjacency list: fullUrl -> list of fullUrls it references.
	adj := make(map[string][]string)
	urlSet := make(map[string]bool)
	for _, e := range entries {
		if e.FullURL != "" {
			urlSet[e.FullURL] = true
		}
	}

	for _, e := range entries {
		if e.FullURL == "" || e.Resource == nil {
			continue
		}
		refs := extractReferences(e.Resource)
		for _, ref := range refs {
			if urlSet[ref] && ref != e.FullURL {
				adj[e.FullURL] = append(adj[e.FullURL], ref)
			}
		}
	}

	// Detect cycles using DFS with coloring.
	const (
		white = 0 // unvisited
		gray  = 1 // visiting (on the current path)
		black = 2 // done
	)
	color := make(map[string]int)
	var issues []ValidationIssue

	var dfs func(node string) bool
	dfs = func(node string) bool {
		color[node] = gray
		for _, neighbor := range adj[node] {
			if color[neighbor] == gray {
				issues = append(issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeBusinessRule,
					Diagnostics: fmt.Sprintf("circular reference detected between %s and %s", node, neighbor),
					Location:    "Bundle.entry",
				})
				return true
			}
			if color[neighbor] == white {
				if dfs(neighbor) {
					return true
				}
			}
		}
		color[node] = black
		return false
	}

	for url := range adj {
		if color[url] == white {
			dfs(url)
		}
	}

	return issues
}

// extractReferences recursively extracts all reference strings from a resource map.
func extractReferences(resource map[string]interface{}) []string {
	var refs []string
	var walk func(v interface{})
	walk = func(v interface{}) {
		switch val := v.(type) {
		case map[string]interface{}:
			if ref, ok := val["reference"].(string); ok {
				refs = append(refs, ref)
			}
			for _, child := range val {
				walk(child)
			}
		case []interface{}:
			for _, item := range val {
				walk(item)
			}
		}
	}
	walk(resource)
	return refs
}

// TransactionProcessor handles the execution of transaction and batch Bundles.
type TransactionProcessor struct {
	ResourceHandler func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error)
}

// NewTransactionProcessor creates a new TransactionProcessor with the given
// resource handler function. The handler is invoked for each entry and should
// perform the actual CRUD operation.
func NewTransactionProcessor(handler func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error)) *TransactionProcessor {
	return &TransactionProcessor{
		ResourceHandler: handler,
	}
}

// ProcessTransaction processes a transaction Bundle atomically. If any entry
// fails, the entire transaction is rolled back and an error is returned with
// an OperationOutcome. On success, a transaction-response Bundle is returned.
func (p *TransactionProcessor) ProcessTransaction(bundle *TransactionBundle) (*Bundle, error) {
	sorted := SortTransactionEntries(bundle.Entries)

	// Build the ID mapping for urn:uuid references.
	idMap := make(map[string]string)
	responseEntries := make([]BundleEntry, len(sorted))

	for i, entry := range sorted {
		// Resolve any urn:uuid references in the resource before processing.
		if entry.Resource != nil && len(idMap) > 0 {
			resolveRefsInResource(entry.Resource, idMap)
		}

		// Also resolve urn:uuid references in the request URL.
		url := replaceURNRefs(entry.Request.URL, idMap)

		resp, err := p.ResourceHandler(entry.Request.Method, url, entry.Resource)
		if err != nil {
			oo := ErrorOutcome(fmt.Sprintf("transaction failed at entry %d (%s %s): %s",
				i, entry.Request.Method, entry.Request.URL, err.Error()))
			return nil, fmt.Errorf("transaction failed: %w; outcome: %s", err, oo.Issue[0].Diagnostics)
		}

		// Map the original fullUrl (urn:uuid:...) to the actual assigned ID.
		if entry.FullURL != "" && strings.HasPrefix(entry.FullURL, "urn:uuid:") && resp.Location != "" {
			idMap[entry.FullURL] = resp.Location
		}

		// Convert BundleEntryResponse to a standard BundleEntry for the response Bundle.
		responseEntries[i] = bundleEntryFromResponse(resp)
	}

	// Resolve any remaining urn:uuid references in response entries.
	now := time.Now().UTC()
	return &Bundle{
		ResourceType: "Bundle",
		Type:         "transaction-response",
		Timestamp:    &now,
		Entry:        responseEntries,
	}, nil
}

// ProcessBatch processes a batch Bundle. Each entry is processed independently.
// If an entry fails, the error is captured in that entry's response and
// processing continues with the remaining entries.
func (p *TransactionProcessor) ProcessBatch(bundle *TransactionBundle) *Bundle {
	responseEntries := make([]BundleEntry, len(bundle.Entries))

	for i, entry := range bundle.Entries {
		resp, err := p.ResourceHandler(entry.Request.Method, entry.Request.URL, entry.Resource)
		if err != nil {
			oo := ErrorOutcome(fmt.Sprintf("batch entry %d failed: %s", i, err.Error()))
			responseEntries[i] = BundleEntry{
				Response: &BundleResponse{
					Status:  "400 Bad Request",
					Outcome: oo,
				},
			}
			continue
		}
		responseEntries[i] = bundleEntryFromResponse(resp)
	}

	now := time.Now().UTC()
	return &Bundle{
		ResourceType: "Bundle",
		Type:         "batch-response",
		Timestamp:    &now,
		Entry:        responseEntries,
	}
}

// bundleEntryFromResponse converts a BundleEntryResponse into a standard
// BundleEntry suitable for inclusion in a response Bundle.
func bundleEntryFromResponse(resp *BundleEntryResponse) BundleEntry {
	var lastMod *time.Time
	if resp.LastModified != "" {
		if t, err := time.Parse(time.RFC3339, resp.LastModified); err == nil {
			lastMod = &t
		}
	}

	return BundleEntry{
		FullURL: resp.Location,
		Response: &BundleResponse{
			Status:       resp.Status,
			Location:     resp.Location,
			LastModified: lastMod,
			Outcome:      resp.Outcome,
		},
	}
}

// ResolveInternalReferences replaces urn:uuid references with actual resource
// IDs in all transaction entries. The idMap maps original urn:uuid values to
// their resolved resource locations (e.g., "Patient/123").
func ResolveInternalReferences(entries []TransactionEntry, idMap map[string]string) {
	for i := range entries {
		if entries[i].Resource != nil {
			resolveRefsInResource(entries[i].Resource, idMap)
		}
		// Also resolve references in the request URL.
		entries[i].Request.URL = replaceURNRefs(entries[i].Request.URL, idMap)
	}
}

// resolveRefsInResource walks a resource map and replaces urn:uuid references
// with the mapped actual IDs.
func resolveRefsInResource(resource map[string]interface{}, idMap map[string]string) {
	var walk func(v interface{}) interface{}
	walk = func(v interface{}) interface{} {
		switch val := v.(type) {
		case map[string]interface{}:
			for k, child := range val {
				if k == "reference" {
					if ref, ok := child.(string); ok {
						if mapped, found := idMap[ref]; found {
							val[k] = mapped
						}
					}
				} else {
					val[k] = walk(child)
				}
			}
			return val
		case []interface{}:
			for i, item := range val {
				val[i] = walk(item)
			}
			return val
		case string:
			if mapped, found := idMap[v.(string)]; found {
				return mapped
			}
			return val
		default:
			return val
		}
	}
	walk(resource)
}

// replaceURNRefs replaces urn:uuid references in a string with mapped values.
func replaceURNRefs(s string, idMap map[string]string) string {
	for urn, actual := range idMap {
		s = strings.ReplaceAll(s, urn, actual)
	}
	return s
}

// SortTransactionEntries sorts entries according to the FHIR specification
// processing order: DELETE first, then POST, then PUT/PATCH, then GET/HEAD.
// The sort is stable, preserving the original order of entries with the same
// method type.
func SortTransactionEntries(entries []TransactionEntry) []TransactionEntry {
	sorted := make([]TransactionEntry, len(entries))
	copy(sorted, entries)

	sort.SliceStable(sorted, func(i, j int) bool {
		oi := methodSortOrder[sorted[i].Request.Method]
		oj := methodSortOrder[sorted[j].Request.Method]
		return oi < oj
	})

	return sorted
}

// ParseEntryURL parses a relative FHIR URL from a Bundle entry request.
// It returns the resource type, resource ID (if present), and whether the
// URL represents a search (contains a query string).
//
// Examples:
//
//	"Patient/123"           -> ("Patient", "123", false)
//	"Patient?name=Smith"    -> ("Patient", "", true)
//	"Patient"               -> ("Patient", "", false)
func ParseEntryURL(url string) (resourceType, id string, isSearch bool) {
	// Check for query string (search).
	if idx := strings.Index(url, "?"); idx >= 0 {
		resourceType = url[:idx]
		isSearch = true
		return resourceType, "", true
	}

	// Split on "/" to extract resourceType and optional id.
	parts := strings.SplitN(url, "/", 3)
	resourceType = parts[0]
	if len(parts) >= 2 {
		id = parts[1]
	}
	return resourceType, id, false
}

// TransactionHandler returns an echo.HandlerFunc that processes FHIR
// transaction and batch Bundle submissions via POST /fhir.
// The provided processor is used to execute the individual entries.
func TransactionHandler(processor *TransactionProcessor) echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeStructure,
				fmt.Sprintf("failed to read request body: %s", err.Error()),
			))
		}

		bundle, err := ParseTransactionBundle(body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeStructure,
				fmt.Sprintf("failed to parse Bundle: %s", err.Error()),
			))
		}

		issues := ValidateTransactionBundle(bundle)
		if len(issues) > 0 {
			// Check if any issues are errors.
			for _, issue := range issues {
				if issue.Severity == SeverityError || issue.Severity == SeverityFatal {
					return c.JSON(http.StatusBadRequest, MultiValidationOutcome(issues))
				}
			}
		}

		switch bundle.Type {
		case "transaction":
			result, err := processor.ProcessTransaction(bundle)
			if err != nil {
				return c.JSON(http.StatusBadRequest, NewOperationOutcome(
					IssueSeverityError, IssueTypeProcessing,
					err.Error(),
				))
			}
			return c.JSON(http.StatusOK, result)

		case "batch":
			result := processor.ProcessBatch(bundle)
			return c.JSON(http.StatusOK, result)

		default:
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeValue,
				fmt.Sprintf("unsupported bundle type %q; expected 'transaction' or 'batch'", bundle.Type),
			))
		}
	}
}
