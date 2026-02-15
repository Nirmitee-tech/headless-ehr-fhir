package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// =============================================================================
// Patient Link / Merge Model
// =============================================================================

// PatientLink represents a link between two patient records (FHIR Patient.link).
type PatientMergeLink struct {
	ID        string    // unique link identifier
	SourceID  string    // source patient FHIR ID
	TargetID  string    // target patient FHIR ID
	Type      string    // replaced-by|replaces|refer|seealso
	CreatedAt time.Time // when the link was created
	CreatedBy string    // user/system that created the link
}

// MergeRequest represents a Patient/$merge request.
type MergeRequest struct {
	SourcePatient map[string]interface{} // Patient to merge FROM (will be deprecated)
	TargetPatient map[string]interface{} // Patient to merge INTO (surviving record)
	ResultPatient map[string]interface{} // Optional: desired result patient (overrides)
	PreviewOnly   bool                   // If true, don't execute — just return preview
}

// MergeResult represents the outcome of a merge operation.
type MergeResult struct {
	Outcome      map[string]interface{} // OperationOutcome
	Input        MergeInput             // echo of input
	Result       map[string]interface{} // resulting Patient resource
	LinksCreated []PatientMergeLink
}

// MergeInput echoes the source and target from the merge request.
type MergeInput struct {
	Source map[string]interface{}
	Target map[string]interface{}
}

// =============================================================================
// Survivorship Rules
// =============================================================================

// SurvivorshipRule defines how to pick the winning value when merging fields.
type SurvivorshipRule struct {
	Field    string // FHIR path (e.g., "name", "telecom", "address")
	Strategy string // "target-wins" | "source-wins" | "most-recent" | "merge-lists"
}

// DefaultSurvivorshipRules returns standard rules for patient merge.
func DefaultSurvivorshipRules() []SurvivorshipRule {
	return []SurvivorshipRule{
		{Field: "identifier", Strategy: "merge-lists"},
		{Field: "name", Strategy: "merge-lists"},
		{Field: "telecom", Strategy: "merge-lists"},
		{Field: "address", Strategy: "merge-lists"},
		{Field: "gender", Strategy: "target-wins"},
		{Field: "birthDate", Strategy: "target-wins"},
		{Field: "maritalStatus", Strategy: "most-recent"},
		{Field: "communication", Strategy: "merge-lists"},
		{Field: "generalPractitioner", Strategy: "target-wins"},
		{Field: "managingOrganization", Strategy: "target-wins"},
	}
}

// =============================================================================
// Reference Rewriter
// =============================================================================

// ReferenceRewriter updates references from one patient to another across resources.
type ReferenceRewriter struct {
	// resourceStore maps resource type -> list of resources (for in-memory rewriting)
	resourceStore map[string][]map[string]interface{}
	mu            sync.RWMutex
}

// NewReferenceRewriter creates a new ReferenceRewriter.
func NewReferenceRewriter() *ReferenceRewriter {
	return &ReferenceRewriter{
		resourceStore: make(map[string][]map[string]interface{}),
	}
}

// AddResources adds resources that may contain patient references.
func (r *ReferenceRewriter) AddResources(resourceType string, resources []map[string]interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resourceStore[resourceType] = append(r.resourceStore[resourceType], resources...)
}

// RewriteReferences updates all references from sourcePatientID to targetPatientID.
// Returns the count of references updated.
func (r *ReferenceRewriter) RewriteReferences(sourcePatientID, targetPatientID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	sourceRef := "Patient/" + sourcePatientID
	targetRef := "Patient/" + targetPatientID
	count := 0

	for _, resources := range r.resourceStore {
		for _, resource := range resources {
			count += rewriteRefsInValue(resource, sourceRef, targetRef)
		}
	}

	return count
}

// rewriteRefsInValue recursively walks a value and replaces patient references.
func rewriteRefsInValue(v interface{}, sourceRef, targetRef string) int {
	count := 0

	switch val := v.(type) {
	case map[string]interface{}:
		// Check for a reference field directly.
		if ref, ok := val["reference"].(string); ok {
			if ref == sourceRef {
				val["reference"] = targetRef
				count++
			}
		}
		// Recurse into all map values.
		for _, child := range val {
			count += rewriteRefsInValue(child, sourceRef, targetRef)
		}
	case []interface{}:
		for _, item := range val {
			count += rewriteRefsInValue(item, sourceRef, targetRef)
		}
	}

	return count
}

// =============================================================================
// MDM (Master Data Management) Service
// =============================================================================

// MDMService manages patient identity linking and golden records.
type MDMService struct {
	links         map[string]*PatientMergeLink // link ID -> link
	linksBySource map[string][]string          // source patient ID -> link IDs
	linksByTarget map[string][]string          // target patient ID -> link IDs
	rules         []SurvivorshipRule
	rewriter      *ReferenceRewriter
	mu            sync.RWMutex
}

// NewMDMService creates a new MDMService with default survivorship rules.
func NewMDMService() *MDMService {
	return &MDMService{
		links:         make(map[string]*PatientMergeLink),
		linksBySource: make(map[string][]string),
		linksByTarget: make(map[string][]string),
		rules:         DefaultSurvivorshipRules(),
		rewriter:      NewReferenceRewriter(),
	}
}

// Merge executes a patient merge operation.
func (s *MDMService) Merge(ctx context.Context, req MergeRequest) (*MergeResult, error) {
	if err := validateMergeRequest(req); err != nil {
		return nil, err
	}

	sourceID, _ := req.SourcePatient["id"].(string)
	targetID, _ := req.TargetPatient["id"].(string)

	// Apply survivorship rules to produce the merged patient.
	merged := s.ApplySurvivorshipRules(req.SourcePatient, req.TargetPatient)

	// Apply result patient overrides if provided.
	if req.ResultPatient != nil {
		for k, v := range req.ResultPatient {
			merged[k] = v
		}
	}

	// Create the link.
	link := &PatientMergeLink{
		ID:        uuid.New().String(),
		SourceID:  sourceID,
		TargetID:  targetID,
		Type:      "replaced-by",
		CreatedAt: time.Now(),
		CreatedBy: "system",
	}

	s.mu.Lock()
	s.links[link.ID] = link
	s.linksBySource[sourceID] = append(s.linksBySource[sourceID], link.ID)
	s.linksByTarget[targetID] = append(s.linksByTarget[targetID], link.ID)
	s.mu.Unlock()

	// Rewrite references across all known resources.
	s.rewriter.RewriteReferences(sourceID, targetID)

	// Build the OperationOutcome.
	outcome := map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue": []interface{}{
			map[string]interface{}{
				"severity":    "information",
				"code":        "informational",
				"diagnostics": fmt.Sprintf("Patient %s merged into %s", sourceID, targetID),
			},
		},
	}

	return &MergeResult{
		Outcome:      outcome,
		Input:        MergeInput{Source: req.SourcePatient, Target: req.TargetPatient},
		Result:       merged,
		LinksCreated: []PatientMergeLink{*link},
	}, nil
}

// Preview returns what would happen without executing.
func (s *MDMService) Preview(ctx context.Context, req MergeRequest) (*MergeResult, error) {
	if err := validateMergeRequest(req); err != nil {
		return nil, err
	}

	sourceID, _ := req.SourcePatient["id"].(string)
	targetID, _ := req.TargetPatient["id"].(string)

	// Apply survivorship rules to produce the merged patient.
	merged := s.ApplySurvivorshipRules(req.SourcePatient, req.TargetPatient)

	// Apply result patient overrides if provided.
	if req.ResultPatient != nil {
		for k, v := range req.ResultPatient {
			merged[k] = v
		}
	}

	outcome := map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue": []interface{}{
			map[string]interface{}{
				"severity":    "information",
				"code":        "informational",
				"diagnostics": fmt.Sprintf("Preview: Patient %s would be merged into %s", sourceID, targetID),
			},
		},
	}

	return &MergeResult{
		Outcome:      outcome,
		Input:        MergeInput{Source: req.SourcePatient, Target: req.TargetPatient},
		Result:       merged,
		LinksCreated: nil,
	}, nil
}

// GetLinks returns all links for a patient (as source or target).
func (s *MDMService) GetLinks(patientID string) []PatientMergeLink {
	s.mu.RLock()
	defer s.mu.RUnlock()

	seen := make(map[string]bool)
	var result []PatientMergeLink

	for _, linkID := range s.linksBySource[patientID] {
		if !seen[linkID] {
			seen[linkID] = true
			if link, ok := s.links[linkID]; ok {
				result = append(result, *link)
			}
		}
	}
	for _, linkID := range s.linksByTarget[patientID] {
		if !seen[linkID] {
			seen[linkID] = true
			if link, ok := s.links[linkID]; ok {
				result = append(result, *link)
			}
		}
	}

	return result
}

// GetGoldenRecord returns the ultimate surviving record for a patient chain.
// It follows replaced-by links until it reaches a patient with no outgoing replaced-by link.
func (s *MDMService) GetGoldenRecord(patientID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	current := patientID
	visited := make(map[string]bool)

	for {
		if visited[current] {
			return "", fmt.Errorf("circular link chain detected for patient %s", patientID)
		}
		visited[current] = true

		// Find a replaced-by link where this patient is the source.
		linkIDs := s.linksBySource[current]
		found := false
		for _, linkID := range linkIDs {
			link, ok := s.links[linkID]
			if !ok {
				continue
			}
			if link.Type == "replaced-by" {
				current = link.TargetID
				found = true
				break
			}
		}
		if !found {
			return current, nil
		}
	}
}

// Unlink removes a link (for undo/correction).
func (s *MDMService) Unlink(linkID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	link, ok := s.links[linkID]
	if !ok {
		return fmt.Errorf("link %s not found", linkID)
	}

	// Remove from links map.
	delete(s.links, linkID)

	// Remove from source index.
	s.linksBySource[link.SourceID] = removeFromSlice(s.linksBySource[link.SourceID], linkID)
	if len(s.linksBySource[link.SourceID]) == 0 {
		delete(s.linksBySource, link.SourceID)
	}

	// Remove from target index.
	s.linksByTarget[link.TargetID] = removeFromSlice(s.linksByTarget[link.TargetID], linkID)
	if len(s.linksByTarget[link.TargetID]) == 0 {
		delete(s.linksByTarget, link.TargetID)
	}

	return nil
}

// ApplySurvivorshipRules merges two patient resources using configured rules.
// The result uses the target's ID and resourceType.
func (s *MDMService) ApplySurvivorshipRules(source, target map[string]interface{}) map[string]interface{} {
	// Start with a copy of the target as the base.
	result := make(map[string]interface{})

	// Copy all target fields first.
	for k, v := range target {
		result[k] = v
	}

	// Build a rule map for quick lookup.
	ruleMap := make(map[string]string)
	for _, rule := range s.rules {
		ruleMap[rule.Field] = rule.Strategy
	}

	// Process each rule.
	for _, rule := range s.rules {
		sourceVal, sourceExists := source[rule.Field]
		targetVal, targetExists := target[rule.Field]

		switch rule.Strategy {
		case "target-wins":
			if targetExists && targetVal != nil {
				result[rule.Field] = targetVal
			} else if sourceExists && sourceVal != nil {
				result[rule.Field] = sourceVal
			}

		case "source-wins":
			if sourceExists && sourceVal != nil {
				result[rule.Field] = sourceVal
			} else if targetExists && targetVal != nil {
				result[rule.Field] = targetVal
			}

		case "most-recent":
			// In the absence of timestamps, target wins on tie.
			if targetExists && targetVal != nil {
				result[rule.Field] = targetVal
			} else if sourceExists && sourceVal != nil {
				result[rule.Field] = sourceVal
			}

		case "merge-lists":
			result[rule.Field] = mergeLists(rule.Field, sourceVal, targetVal)
		}
	}

	// Copy any source fields not covered by rules (fallback).
	for k, v := range source {
		if _, covered := ruleMap[k]; covered {
			continue
		}
		// Skip metadata fields.
		if k == "resourceType" || k == "id" || k == "meta" {
			continue
		}
		if _, exists := result[k]; !exists {
			result[k] = v
		}
	}

	return result
}

// mergeLists combines two slices, deduplicating based on the field type.
func mergeLists(field string, sourceVal, targetVal interface{}) interface{} {
	sourceList := toSlice(sourceVal)
	targetList := toSlice(targetVal)

	if len(sourceList) == 0 {
		if len(targetList) == 0 {
			return []interface{}{}
		}
		return targetList
	}
	if len(targetList) == 0 {
		return sourceList
	}

	switch field {
	case "identifier":
		return mergeIdentifiers(sourceList, targetList)
	case "telecom":
		return mergeTelecom(sourceList, targetList)
	default:
		// For name, address, communication: just append all (simple merge).
		return mergeGenericLists(sourceList, targetList)
	}
}

// mergeIdentifiers deduplicates identifiers by system (target wins for same system).
func mergeIdentifiers(sourceList, targetList []interface{}) []interface{} {
	// Index target identifiers by system.
	targetSystems := make(map[string]bool)
	for _, item := range targetList {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		system, _ := m["system"].(string)
		if system != "" {
			targetSystems[system] = true
		}
	}

	// Start with all target identifiers.
	result := make([]interface{}, len(targetList))
	copy(result, targetList)

	// Add source identifiers whose system is not already in target.
	for _, item := range sourceList {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		system, _ := m["system"].(string)
		if system != "" && targetSystems[system] {
			continue // Dedup: target wins.
		}
		result = append(result, item)
	}

	return result
}

// mergeTelecom deduplicates telecom entries by system+value.
func mergeTelecom(sourceList, targetList []interface{}) []interface{} {
	type telecomKey struct {
		system string
		value  string
	}

	seen := make(map[telecomKey]bool)
	result := make([]interface{}, 0, len(targetList)+len(sourceList))

	// Add all target telecoms first.
	for _, item := range targetList {
		m, ok := item.(map[string]interface{})
		if !ok {
			result = append(result, item)
			continue
		}
		sys, _ := m["system"].(string)
		val, _ := m["value"].(string)
		key := telecomKey{system: sys, value: val}
		seen[key] = true
		result = append(result, item)
	}

	// Add source telecoms not already present.
	for _, item := range sourceList {
		m, ok := item.(map[string]interface{})
		if !ok {
			result = append(result, item)
			continue
		}
		sys, _ := m["system"].(string)
		val, _ := m["value"].(string)
		key := telecomKey{system: sys, value: val}
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, item)
	}

	return result
}

// mergeGenericLists appends source items to target items without dedup.
func mergeGenericLists(sourceList, targetList []interface{}) []interface{} {
	result := make([]interface{}, 0, len(targetList)+len(sourceList))
	result = append(result, targetList...)
	result = append(result, sourceList...)
	return result
}

// toSlice converts a value to []interface{}, handling nil and typed slices.
func toSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	if s, ok := v.([]interface{}); ok {
		return s
	}
	return nil
}

// removeFromSlice removes a string from a slice.
func removeFromSlice(s []string, val string) []string {
	result := make([]string, 0, len(s))
	for _, item := range s {
		if item != val {
			result = append(result, item)
		}
	}
	return result
}

// validateMergeRequest validates the merge request inputs.
func validateMergeRequest(req MergeRequest) error {
	if req.SourcePatient == nil {
		return fmt.Errorf("source patient is required")
	}
	if req.TargetPatient == nil {
		return fmt.Errorf("target patient is required")
	}
	sourceID, _ := req.SourcePatient["id"].(string)
	targetID, _ := req.TargetPatient["id"].(string)
	if sourceID != "" && targetID != "" && sourceID == targetID {
		return fmt.Errorf("cannot merge a patient into itself")
	}
	return nil
}

// =============================================================================
// HTTP Handler
// =============================================================================

// MergeHandler provides FHIR Patient/$merge HTTP endpoints.
type MergeHandler struct {
	mdm *MDMService
}

// NewMergeHandler creates a new MergeHandler.
func NewMergeHandler(mdm *MDMService) *MergeHandler {
	return &MergeHandler{mdm: mdm}
}

// RegisterRoutes adds merge routes to the given FHIR group.
func (h *MergeHandler) RegisterRoutes(fhirGroup *echo.Group) {
	fhirGroup.POST("/Patient/$merge", h.HandleMerge)
	fhirGroup.GET("/Patient/:id/$links", h.HandleGetLinks)
	fhirGroup.GET("/Patient/:id/$golden-record", h.HandleGoldenRecord)
	fhirGroup.DELETE("/Patient/$link/:id", h.HandleDeleteLink)
	fhirGroup.GET("/Patient/$links", h.HandleListAllLinks)
}

// HandleMerge handles POST /fhir/Patient/$merge.
func (h *MergeHandler) HandleMerge(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewOperationOutcome(
			IssueSeverityError, IssueTypeInvalid, "Failed to read request body",
		))
	}
	if len(body) == 0 {
		return c.JSON(http.StatusBadRequest, NewOperationOutcome(
			IssueSeverityError, IssueTypeInvalid, "Request body is empty",
		))
	}

	var params map[string]interface{}
	if err := json.Unmarshal(body, &params); err != nil {
		return c.JSON(http.StatusBadRequest, NewOperationOutcome(
			IssueSeverityError, IssueTypeStructure, "Invalid JSON: "+err.Error(),
		))
	}

	rt, _ := params["resourceType"].(string)
	if rt != "Parameters" {
		return c.JSON(http.StatusBadRequest, NewOperationOutcome(
			IssueSeverityError, IssueTypeStructure, "Expected resourceType 'Parameters'",
		))
	}

	paramList, ok := params["parameter"].([]interface{})
	if !ok {
		return c.JSON(http.StatusBadRequest, NewOperationOutcome(
			IssueSeverityError, IssueTypeStructure, "Expected 'parameter' array in Parameters resource",
		))
	}

	var sourcePatient, targetPatient, resultPatient map[string]interface{}
	preview := false

	for _, pRaw := range paramList {
		p, ok := pRaw.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := p["name"].(string)
		switch name {
		case "source-patient":
			if res, ok := p["resource"].(map[string]interface{}); ok {
				sourcePatient = res
			}
		case "target-patient":
			if res, ok := p["resource"].(map[string]interface{}); ok {
				targetPatient = res
			}
		case "result-patient":
			if res, ok := p["resource"].(map[string]interface{}); ok {
				resultPatient = res
			}
		case "preview":
			if v, ok := p["valueBoolean"].(bool); ok {
				preview = v
			}
		}
	}

	if sourcePatient == nil {
		return c.JSON(http.StatusBadRequest, NewOperationOutcome(
			IssueSeverityError, IssueTypeRequired, "Missing 'source-patient' parameter",
		))
	}
	if targetPatient == nil {
		return c.JSON(http.StatusBadRequest, NewOperationOutcome(
			IssueSeverityError, IssueTypeRequired, "Missing 'target-patient' parameter",
		))
	}

	req := MergeRequest{
		SourcePatient: sourcePatient,
		TargetPatient: targetPatient,
		ResultPatient: resultPatient,
		PreviewOnly:   preview,
	}

	var result *MergeResult
	if preview {
		result, err = h.mdm.Preview(c.Request().Context(), req)
	} else {
		result, err = h.mdm.Merge(c.Request().Context(), req)
	}
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "itself") {
			status = http.StatusBadRequest
		}
		return c.JSON(status, NewOperationOutcome(
			IssueSeverityError, IssueTypeProcessing, err.Error(),
		))
	}

	// Build FHIR Parameters response.
	response := buildMergeResponse(result)
	return c.JSON(http.StatusOK, response)
}

// HandleGetLinks handles GET /fhir/Patient/:id/$links.
func (h *MergeHandler) HandleGetLinks(c echo.Context) error {
	patientID := c.Param("id")
	links := h.mdm.GetLinks(patientID)

	entries := make([]interface{}, 0, len(links))
	for _, link := range links {
		entries = append(entries, map[string]interface{}{
			"resource": map[string]interface{}{
				"resourceType": "Basic",
				"id":           link.ID,
				"code": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{
							"system": "http://terminology.hl7.org/CodeSystem/v3-ActCode",
							"code":   "PATMERGE",
						},
					},
				},
				"subject": map[string]interface{}{
					"reference": "Patient/" + link.SourceID,
				},
				"extension": []interface{}{
					map[string]interface{}{
						"url":            "http://ehr.org/fhir/StructureDefinition/merge-target",
						"valueReference": map[string]interface{}{"reference": "Patient/" + link.TargetID},
					},
					map[string]interface{}{
						"url":       "http://ehr.org/fhir/StructureDefinition/link-type",
						"valueCode": link.Type,
					},
				},
				"created": link.CreatedAt.Format(time.RFC3339),
			},
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

// HandleGoldenRecord handles GET /fhir/Patient/:id/$golden-record.
func (h *MergeHandler) HandleGoldenRecord(c echo.Context) error {
	patientID := c.Param("id")
	golden, err := h.mdm.GetGoldenRecord(patientID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
			IssueSeverityError, IssueTypeProcessing, err.Error(),
		))
	}

	response := map[string]interface{}{
		"resourceType": "Parameters",
		"parameter": []interface{}{
			map[string]interface{}{
				"name":        "golden-record-id",
				"valueString": golden,
			},
			map[string]interface{}{
				"name": "golden-record-reference",
				"valueReference": map[string]interface{}{
					"reference": "Patient/" + golden,
				},
			},
		},
	}
	return c.JSON(http.StatusOK, response)
}

// HandleDeleteLink handles DELETE /fhir/Patient/$link/:id.
func (h *MergeHandler) HandleDeleteLink(c echo.Context) error {
	linkID := c.Param("id")
	err := h.mdm.Unlink(linkID)
	if err != nil {
		return c.JSON(http.StatusNotFound, NewOperationOutcome(
			IssueSeverityError, IssueTypeNotFound, err.Error(),
		))
	}

	outcome := map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue": []interface{}{
			map[string]interface{}{
				"severity":    "information",
				"code":        "informational",
				"diagnostics": fmt.Sprintf("Link %s removed", linkID),
			},
		},
	}
	return c.JSON(http.StatusOK, outcome)
}

// HandleListAllLinks handles GET /fhir/Patient/$links.
func (h *MergeHandler) HandleListAllLinks(c echo.Context) error {
	h.mdm.mu.RLock()
	allLinks := make([]PatientMergeLink, 0, len(h.mdm.links))
	for _, link := range h.mdm.links {
		allLinks = append(allLinks, *link)
	}
	h.mdm.mu.RUnlock()

	entries := make([]interface{}, 0, len(allLinks))
	for _, link := range allLinks {
		entries = append(entries, map[string]interface{}{
			"resource": map[string]interface{}{
				"resourceType": "Basic",
				"id":           link.ID,
				"subject": map[string]interface{}{
					"reference": "Patient/" + link.SourceID,
				},
				"extension": []interface{}{
					map[string]interface{}{
						"url":            "http://ehr.org/fhir/StructureDefinition/merge-target",
						"valueReference": map[string]interface{}{"reference": "Patient/" + link.TargetID},
					},
					map[string]interface{}{
						"url":       "http://ehr.org/fhir/StructureDefinition/link-type",
						"valueCode": link.Type,
					},
				},
				"created": link.CreatedAt.Format(time.RFC3339),
			},
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

// buildMergeResponse builds the FHIR Parameters response for a merge operation.
func buildMergeResponse(result *MergeResult) map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Parameters",
		"parameter": []interface{}{
			map[string]interface{}{
				"name":     "outcome",
				"resource": result.Outcome,
			},
			map[string]interface{}{
				"name":     "result",
				"resource": result.Result,
			},
			map[string]interface{}{
				"name": "input",
				"part": []interface{}{
					map[string]interface{}{
						"name":     "source",
						"resource": result.Input.Source,
					},
					map[string]interface{}{
						"name":     "target",
						"resource": result.Input.Target,
					},
				},
			},
		},
	}
}
