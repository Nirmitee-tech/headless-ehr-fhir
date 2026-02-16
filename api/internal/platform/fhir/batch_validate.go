package fhir

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

// BatchValidateRequest holds the list of resources to validate in a single batch.
type BatchValidateRequest struct {
	Resources []json.RawMessage `json:"resources"`
}

// BatchValidateResult holds the aggregated validation results for a batch.
type BatchValidateResult struct {
	TotalCount   int                    `json:"totalCount"`
	ValidCount   int                    `json:"validCount"`
	InvalidCount int                    `json:"invalidCount"`
	Results      []SingleValidateResult `json:"results"`
}

// SingleValidateResult holds the validation outcome for a single resource within a batch.
type SingleValidateResult struct {
	Index        int                      `json:"index"`
	ResourceType string                   `json:"resourceType,omitempty"`
	ResourceID   string                   `json:"resourceId,omitempty"`
	Valid        bool                     `json:"valid"`
	Issues       []map[string]interface{} `json:"issues,omitempty"`
}

// resourcesRequiringStatus lists resource types that must have a status field
// according to the FHIR R4 required-fields registry.
var resourcesRequiringStatus = func() map[string]bool {
	m := make(map[string]bool)
	for rt, fields := range requiredFieldsRegistry {
		for _, f := range fields {
			if f == "status" || f == "lifecycleStatus" {
				m[rt] = true
				break
			}
		}
	}
	return m
}()

// ValidateResourceStructure performs basic structural validation on a parsed
// FHIR resource map. It checks:
//   - resourceType is present and valid (via IsValidResourceType)
//   - id format is valid when present (alphanumeric, hyphens, dots, 1-64 chars)
//   - meta.versionId is a string when present
//   - status field is present for resources that require it
func ValidateResourceStructure(resource map[string]interface{}) []ValidationIssue {
	var issues []ValidationIssue

	// 1. Validate resourceType is present and a non-empty string.
	rtVal, hasRT := resource["resourceType"]
	if !hasRT {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityFatal,
			Code:        VIssueTypeStructure,
			Location:    "resourceType",
			Diagnostics: "resourceType is required",
		})
		return issues
	}

	rt, ok := rtVal.(string)
	if !ok || rt == "" {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityFatal,
			Code:        VIssueTypeStructure,
			Location:    "resourceType",
			Diagnostics: "resourceType must be a non-empty string",
		})
		return issues
	}

	// Validate that resourceType is a known FHIR R4 type.
	if !IsValidResourceType(rt) {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeStructure,
			Location:    "resourceType",
			Diagnostics: fmt.Sprintf("Unknown resource type '%s'", rt),
		})
	}

	// 2. Validate id format if present.
	if idVal, hasID := resource["id"]; hasID {
		idStr, isStr := idVal.(string)
		if !isStr {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeValue,
				Location:    "id",
				Diagnostics: "id must be a string",
			})
		} else if idStr != "" && !fhirIDPattern.MatchString(idStr) {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeValue,
				Location:    "id",
				Diagnostics: fmt.Sprintf("id '%s' does not match FHIR id format (alphanumeric, hyphens, dots, 1-64 chars)", idStr),
			})
		}
	}

	// 3. Validate meta.versionId is a string if present.
	if metaVal, hasMeta := resource["meta"]; hasMeta {
		metaMap, metaOk := metaVal.(map[string]interface{})
		if !metaOk {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeStructure,
				Location:    "meta",
				Diagnostics: "meta must be an object",
			})
		} else if vid, hasVID := metaMap["versionId"]; hasVID {
			if _, vidOk := vid.(string); !vidOk {
				issues = append(issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeValue,
					Location:    "meta.versionId",
					Diagnostics: "meta.versionId must be a string",
				})
			}
		}
	}

	// 4. Validate that status is present for resource types that require it.
	if resourcesRequiringStatus[rt] {
		if _, hasStatus := resource["status"]; !hasStatus {
			// Goal uses lifecycleStatus instead of status.
			if rt == "Goal" {
				if _, hasLC := resource["lifecycleStatus"]; !hasLC {
					issues = append(issues, ValidationIssue{
						Severity:    SeverityError,
						Code:        VIssueTypeRequired,
						Location:    fmt.Sprintf("%s.lifecycleStatus", rt),
						Diagnostics: fmt.Sprintf("Required field 'lifecycleStatus' is missing for %s", rt),
					})
				}
			} else {
				issues = append(issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeRequired,
					Location:    fmt.Sprintf("%s.status", rt),
					Diagnostics: fmt.Sprintf("Required field 'status' is missing for %s", rt),
				})
			}
		}
	}

	return issues
}

// BatchValidateHandler handles POST /fhir/$batch-validate requests.
type BatchValidateHandler struct {
	validator *ResourceValidator
}

// NewBatchValidateHandler creates a new BatchValidateHandler.
func NewBatchValidateHandler(validator *ResourceValidator) *BatchValidateHandler {
	return &BatchValidateHandler{validator: validator}
}

// RegisterRoutes adds the $batch-validate route to the given FHIR route group.
func (h *BatchValidateHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/$batch-validate", h.Handle)
}

// Handle processes a batch validation request. It accepts either:
//   - A JSON object with a "resources" array (BatchValidateRequest)
//   - A FHIR Bundle whose entries each contain a resource to validate
func (h *BatchValidateHandler) Handle(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, buildBatchValidateError("Failed to read request body"))
	}

	if len(body) == 0 {
		return c.JSON(http.StatusBadRequest, buildBatchValidateError("Request body is empty"))
	}

	// Try to determine if the payload is a FHIR Bundle or a BatchValidateRequest.
	var rawObj map[string]interface{}
	if err := json.Unmarshal(body, &rawObj); err != nil {
		return c.JSON(http.StatusBadRequest, buildBatchValidateError("Invalid JSON: "+err.Error()))
	}

	var rawResources []json.RawMessage

	if rt, _ := rawObj["resourceType"].(string); rt == "Bundle" {
		// Parse as a FHIR Bundle and extract resources from entries.
		resources, extractErr := extractBundleResources(body)
		if extractErr != nil {
			return c.JSON(http.StatusBadRequest, buildBatchValidateError(extractErr.Error()))
		}
		rawResources = resources
	} else {
		// Parse as a BatchValidateRequest.
		var req BatchValidateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return c.JSON(http.StatusBadRequest, buildBatchValidateError("Invalid JSON: "+err.Error()))
		}
		if len(req.Resources) == 0 {
			return c.JSON(http.StatusBadRequest, buildBatchValidateError("No resources provided for validation"))
		}
		rawResources = req.Resources
	}

	result := h.validateResources(rawResources)
	return c.JSON(http.StatusOK, result)
}

// validateResources validates each raw JSON resource and returns a BatchValidateResult.
func (h *BatchValidateHandler) validateResources(rawResources []json.RawMessage) *BatchValidateResult {
	result := &BatchValidateResult{
		TotalCount: len(rawResources),
		Results:    make([]SingleValidateResult, 0, len(rawResources)),
	}

	for i, raw := range rawResources {
		single := h.validateSingle(i, raw)
		if single.Valid {
			result.ValidCount++
		} else {
			result.InvalidCount++
		}
		result.Results = append(result.Results, single)
	}

	return result
}

// validateSingle validates a single raw JSON resource and returns a SingleValidateResult.
func (h *BatchValidateHandler) validateSingle(index int, raw json.RawMessage) SingleValidateResult {
	single := SingleValidateResult{
		Index: index,
		Valid: true,
	}

	var resource map[string]interface{}
	if err := json.Unmarshal(raw, &resource); err != nil {
		single.Valid = false
		single.Issues = append(single.Issues, map[string]interface{}{
			"severity":    string(SeverityFatal),
			"code":        string(VIssueTypeStructure),
			"diagnostics": "Invalid JSON: " + err.Error(),
		})
		return single
	}

	// Extract resource type and id for the result metadata.
	if rt, ok := resource["resourceType"].(string); ok {
		single.ResourceType = rt
	}
	if id, ok := resource["id"].(string); ok {
		single.ResourceID = id
	}

	// Run structural validation.
	structIssues := ValidateResourceStructure(resource)

	// Run the full ResourceValidator for deeper checks.
	fullResult := h.validator.Validate(resource)

	// Merge issues, deduplicating by combining structural issues with full validation.
	// Use the full validator result as the primary source. Structural issues provide
	// a subset of those checks, so we rely on the full result and only add structural
	// issues that cover checks not present in the full validator (status requirement
	// from resourcesRequiringStatus).
	issueSet := make(map[string]bool)
	var mergedIssues []ValidationIssue

	for _, issue := range fullResult.Issues {
		key := fmt.Sprintf("%s|%s|%s", issue.Severity, issue.Code, issue.Diagnostics)
		if !issueSet[key] {
			issueSet[key] = true
			mergedIssues = append(mergedIssues, issue)
		}
	}
	for _, issue := range structIssues {
		key := fmt.Sprintf("%s|%s|%s", issue.Severity, issue.Code, issue.Diagnostics)
		if !issueSet[key] {
			issueSet[key] = true
			mergedIssues = append(mergedIssues, issue)
		}
	}

	if !fullResult.Valid || len(structIssues) > 0 {
		// Check if any issue is an error or fatal to determine validity.
		for _, issue := range mergedIssues {
			if issue.Severity == SeverityError || issue.Severity == SeverityFatal {
				single.Valid = false
				break
			}
		}
	}

	// Convert ValidationIssue structs to map representation for the response.
	if len(mergedIssues) > 0 {
		single.Issues = make([]map[string]interface{}, 0, len(mergedIssues))
		for _, issue := range mergedIssues {
			entry := map[string]interface{}{
				"severity":    string(issue.Severity),
				"code":        string(issue.Code),
				"diagnostics": issue.Diagnostics,
			}
			if issue.Location != "" {
				entry["location"] = issue.Location
			}
			single.Issues = append(single.Issues, entry)
		}
	}

	return single
}

// extractBundleResources parses a FHIR Bundle from raw JSON and extracts the
// resource from each entry.
func extractBundleResources(body []byte) ([]json.RawMessage, error) {
	var bundle struct {
		ResourceType string `json:"resourceType"`
		Entry        []struct {
			Resource json.RawMessage `json:"resource"`
		} `json:"entry"`
	}

	if err := json.Unmarshal(body, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse Bundle: %s", err.Error())
	}

	if len(bundle.Entry) == 0 {
		return nil, fmt.Errorf("Bundle contains no entries")
	}

	resources := make([]json.RawMessage, 0, len(bundle.Entry))
	for i, entry := range bundle.Entry {
		if len(entry.Resource) == 0 {
			return nil, fmt.Errorf("Bundle entry[%d] has no resource", i)
		}
		resources = append(resources, entry.Resource)
	}

	return resources, nil
}

// buildBatchValidateError creates a simple error response for batch validation failures.
func buildBatchValidateError(message string) map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue": []map[string]interface{}{
			{
				"severity":    string(SeverityFatal),
				"code":        string(VIssueTypeStructure),
				"diagnostics": message,
			},
		},
	}
}
