package fhir

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ============================================================================
// $apply Operation Types
// ============================================================================

// ApplyRequest holds the input parameters for the $apply operation.
type ApplyRequest struct {
	PlanDefinitionID string                 // ID of PlanDefinition to apply
	Subject          string                 // Patient reference (e.g. "Patient/123")
	Encounter        string                 // Optional encounter context
	Practitioner     string                 // Optional practitioner context
	Organization     string                 // Optional organization context
	Parameters       map[string]interface{} // Additional parameters
}

// ApplyActivity represents a single generated activity from $apply.
type ApplyActivity struct {
	ResourceType string                 // Type of resource generated
	Resource     map[string]interface{} // The generated resource
	ActionID     string                 // ID of the action that generated this
}

// ApplyPlanDefinitionAction represents actions within a PlanDefinition for $apply.
type ApplyPlanDefinitionAction struct {
	ID                string
	Title             string
	Description       string
	Priority          string // routine | urgent | asap | stat
	Condition         []ApplyActionCondition
	Timing            map[string]interface{}
	Type              string // create | update | remove | fire-event
	DefinitionURI     string // Referenced ActivityDefinition
	Action            []ApplyPlanDefinitionAction
	RelatedAction     []ApplyRelatedAction
	SelectionBehavior string // any | all | all-or-none | exactly-one | at-most-one | one-or-more
	GroupingBehavior  string // visual-group | logical-group | sentence-group
}

// ApplyActionCondition defines when an action applies.
type ApplyActionCondition struct {
	Kind       string // applicability | start | stop
	Expression string // FHIRPath expression
}

// ApplyRelatedAction describes dependencies between actions.
type ApplyRelatedAction struct {
	ActionID       string
	Relationship   string // before-start | before | before-end | concurrent-with-start | concurrent | concurrent-with-end | after-start | after | after-end
	OffsetDuration map[string]interface{}
}

// ParsedPlanDefinition is the structured result of parsing a PlanDefinition JSON.
type ParsedPlanDefinition struct {
	ID      string
	URL     string
	Version string
	Name    string
	Title   string
	Status  string
	Type    string
	Actions []ApplyPlanDefinitionAction
	Raw     map[string]interface{}
}

// ParsedActivityDefinition is the structured result of parsing an ActivityDefinition JSON.
type ParsedActivityDefinition struct {
	ID            string
	URL           string
	Name          string
	Title         string
	Status        string
	Kind          string
	Code          map[string]interface{}
	Dosage        []interface{}
	Timing        map[string]interface{}
	DynamicValues []ApplyDynamicValue
	Raw           map[string]interface{}
}

// ApplyDynamicValue specifies a computed field value using a FHIRPath expression.
type ApplyDynamicValue struct {
	Path       string
	Expression string
}

// ApplyOperationResult holds the output of the $apply operation.
// This is the public result returned by ApplyPlanDefinition.
// The field names align with the FHIR $apply output spec.
type ApplyOperationResult struct {
	ResourceType string                 // "CarePlan" or "RequestGroup"
	CarePlan     map[string]interface{} // Generated CarePlan (for PlanDefinition)
	RequestGroup map[string]interface{} // Generated RequestGroup
	Activities   []ApplyActivity        // Individual activities
}

// ============================================================================
// Parsing Functions
// ============================================================================

// ParseApplyPlanDefinition parses a PlanDefinition JSON map into a ParsedPlanDefinition.
func ParseApplyPlanDefinition(data map[string]interface{}) (*ParsedPlanDefinition, error) {
	if data == nil {
		return nil, fmt.Errorf("PlanDefinition data is nil")
	}

	rt, _ := data["resourceType"].(string)
	if rt != "" && rt != "PlanDefinition" {
		return nil, fmt.Errorf("expected resourceType PlanDefinition, got %s", rt)
	}

	parsed := &ParsedPlanDefinition{
		Raw: data,
	}

	parsed.ID, _ = data["id"].(string)
	parsed.URL, _ = data["url"].(string)
	parsed.Version, _ = data["version"].(string)
	parsed.Name, _ = data["name"].(string)
	parsed.Title, _ = data["title"].(string)
	parsed.Status, _ = data["status"].(string)
	parsed.Type, _ = data["type"].(string)

	// Parse actions
	if actionsRaw, ok := data["action"]; ok {
		parsed.Actions = parseApplyActions(actionsRaw)
	}

	return parsed, nil
}

// parseApplyActions parses an action array from JSON.
func parseApplyActions(actionsRaw interface{}) []ApplyPlanDefinitionAction {
	var result []ApplyPlanDefinitionAction

	var actionSlice []interface{}
	switch v := actionsRaw.(type) {
	case []interface{}:
		actionSlice = v
	case []map[string]interface{}:
		for _, m := range v {
			actionSlice = append(actionSlice, m)
		}
	default:
		return result
	}

	for _, item := range actionSlice {
		actionMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		action := parseApplySingleAction(actionMap)
		result = append(result, action)
	}

	return result
}

// parseApplySingleAction parses a single action map.
func parseApplySingleAction(m map[string]interface{}) ApplyPlanDefinitionAction {
	action := ApplyPlanDefinitionAction{}

	action.ID, _ = m["id"].(string)
	action.Title, _ = m["title"].(string)
	action.Description, _ = m["description"].(string)
	action.Priority, _ = m["priority"].(string)
	action.SelectionBehavior, _ = m["selectionBehavior"].(string)
	action.GroupingBehavior, _ = m["groupingBehavior"].(string)
	action.DefinitionURI, _ = m["definitionCanonical"].(string)

	// Parse type from CodeableConcept
	if typeObj, ok := m["type"].(map[string]interface{}); ok {
		if codings, ok := typeObj["coding"].([]interface{}); ok && len(codings) > 0 {
			if coding, ok := codings[0].(map[string]interface{}); ok {
				action.Type, _ = coding["code"].(string)
			}
		}
	}

	// Parse conditions
	if condRaw, ok := m["condition"].([]interface{}); ok {
		for _, c := range condRaw {
			condMap, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			cond := ApplyActionCondition{}
			cond.Kind, _ = condMap["kind"].(string)
			// Expression can be a string or an object with an "expression" field
			if exprStr, ok := condMap["expression"].(string); ok {
				cond.Expression = exprStr
			} else if exprObj, ok := condMap["expression"].(map[string]interface{}); ok {
				cond.Expression, _ = exprObj["expression"].(string)
			}
			action.Condition = append(action.Condition, cond)
		}
	}

	// Parse timing (various timing[x] elements)
	for _, timingKey := range []string{"timingTiming", "timingDateTime", "timingAge", "timingPeriod", "timingDuration", "timingRange"} {
		if timing, ok := m[timingKey]; ok {
			if timingMap, ok := timing.(map[string]interface{}); ok {
				action.Timing = timingMap
			}
		}
	}

	// Parse relatedAction
	if relRaw, ok := m["relatedAction"].([]interface{}); ok {
		for _, r := range relRaw {
			relMap, ok := r.(map[string]interface{})
			if !ok {
				continue
			}
			rel := ApplyRelatedAction{}
			rel.ActionID, _ = relMap["actionId"].(string)
			rel.Relationship, _ = relMap["relationship"].(string)
			if offset, ok := relMap["offsetDuration"].(map[string]interface{}); ok {
				rel.OffsetDuration = offset
			}
			action.RelatedAction = append(action.RelatedAction, rel)
		}
	}

	// Parse nested actions
	if subActions, ok := m["action"]; ok {
		action.Action = parseApplyActions(subActions)
	}

	return action
}

// ParseApplyActivityDefinition parses an ActivityDefinition JSON map.
func ParseApplyActivityDefinition(data map[string]interface{}) (*ParsedActivityDefinition, error) {
	if data == nil {
		return nil, fmt.Errorf("ActivityDefinition data is nil")
	}

	rt, _ := data["resourceType"].(string)
	if rt != "" && rt != "ActivityDefinition" {
		return nil, fmt.Errorf("expected resourceType ActivityDefinition, got %s", rt)
	}

	parsed := &ParsedActivityDefinition{
		Raw: data,
	}

	parsed.ID, _ = data["id"].(string)
	parsed.URL, _ = data["url"].(string)
	parsed.Name, _ = data["name"].(string)
	parsed.Title, _ = data["title"].(string)
	parsed.Status, _ = data["status"].(string)
	parsed.Kind, _ = data["kind"].(string)

	if code, ok := data["code"].(map[string]interface{}); ok {
		parsed.Code = code
	}

	if dosage, ok := data["dosage"].([]interface{}); ok {
		parsed.Dosage = dosage
	}

	if timing, ok := data["timing"].(map[string]interface{}); ok {
		parsed.Timing = timing
	}

	// Parse dynamicValue
	if dvRaw, ok := data["dynamicValue"].([]interface{}); ok {
		for _, dv := range dvRaw {
			dvMap, ok := dv.(map[string]interface{})
			if !ok {
				continue
			}
			adv := ApplyDynamicValue{}
			adv.Path, _ = dvMap["path"].(string)
			if exprStr, ok := dvMap["expression"].(string); ok {
				adv.Expression = exprStr
			} else if exprObj, ok := dvMap["expression"].(map[string]interface{}); ok {
				adv.Expression, _ = exprObj["expression"].(string)
			}
			parsed.DynamicValues = append(parsed.DynamicValues, adv)
		}
	}

	return parsed, nil
}

// ============================================================================
// Apply Functions
// ============================================================================

// ApplyPlanDefinition applies a PlanDefinition to generate a CarePlan with activities.
func ApplyPlanDefinition(planDef *ParsedPlanDefinition, req *ApplyRequest) (*ApplyOperationResult, error) {
	if planDef == nil {
		return nil, fmt.Errorf("PlanDefinition is nil")
	}
	if req == nil {
		return nil, fmt.Errorf("ApplyRequest is nil")
	}
	if planDef.Status == "retired" {
		return nil, fmt.Errorf("cannot apply retired PlanDefinition %s", planDef.ID)
	}

	now := time.Now().UTC().Format(time.RFC3339)

	result := &ApplyOperationResult{
		ResourceType: "CarePlan",
		Activities:   make([]ApplyActivity, 0),
	}

	// Process actions into RequestGroup actions
	rgActions := make([]interface{}, 0)
	processApplyActions(planDef.Actions, req, result, &rgActions)

	// Build CarePlan
	cpID := uuid.New().String()
	carePlan := map[string]interface{}{
		"resourceType":       "CarePlan",
		"id":                 cpID,
		"status":             "active",
		"intent":             "plan",
		"subject":            map[string]interface{}{"reference": req.Subject},
		"created":            now,
		"instantiatesCanonical": []interface{}{"PlanDefinition/" + planDef.ID},
	}

	if planDef.Title != "" {
		carePlan["title"] = planDef.Title
	}

	if req.Encounter != "" {
		carePlan["encounter"] = map[string]interface{}{"reference": req.Encounter}
	}

	if req.Practitioner != "" {
		carePlan["author"] = map[string]interface{}{"reference": req.Practitioner}
	}

	// Link activities in the CarePlan
	if len(result.Activities) > 0 {
		activities := make([]interface{}, 0, len(result.Activities))
		for _, act := range result.Activities {
			activities = append(activities, map[string]interface{}{
				"reference": map[string]interface{}{
					"reference": fmt.Sprintf("%s/%s", act.Resource["resourceType"], act.Resource["id"]),
				},
			})
		}
		carePlan["activity"] = activities
	}

	result.CarePlan = carePlan

	// Build RequestGroup
	rgID := uuid.New().String()
	requestGroup := map[string]interface{}{
		"resourceType": "RequestGroup",
		"id":           rgID,
		"status":       "draft",
		"intent":       "proposal",
		"subject":      map[string]interface{}{"reference": req.Subject},
		"authoredOn":   now,
	}

	if len(rgActions) > 0 {
		requestGroup["action"] = rgActions
	}

	result.RequestGroup = requestGroup

	return result, nil
}

// processApplyActions recursively processes actions and builds request group actions.
func processApplyActions(
	actions []ApplyPlanDefinitionAction,
	req *ApplyRequest,
	result *ApplyOperationResult,
	rgActions *[]interface{},
) {
	for _, action := range actions {
		// Check applicability conditions
		if !evaluateApplyConditions(action.Condition) {
			continue
		}

		rgAction := map[string]interface{}{}

		if action.ID != "" {
			rgAction["id"] = action.ID
		}
		if action.Title != "" {
			rgAction["title"] = action.Title
		}
		if action.Description != "" {
			rgAction["description"] = action.Description
		}
		if action.Priority != "" {
			rgAction["priority"] = action.Priority
		}
		if action.SelectionBehavior != "" {
			rgAction["selectionBehavior"] = action.SelectionBehavior
		}
		if action.GroupingBehavior != "" {
			rgAction["groupingBehavior"] = action.GroupingBehavior
		}

		// Type
		if action.Type != "" {
			rgAction["type"] = map[string]interface{}{
				"coding": []interface{}{map[string]interface{}{"code": action.Type}},
			}
		}

		// Timing
		if action.Timing != nil {
			rgAction["timingTiming"] = action.Timing
		}

		// Related actions
		if len(action.RelatedAction) > 0 {
			related := make([]interface{}, 0, len(action.RelatedAction))
			for _, ra := range action.RelatedAction {
				rel := map[string]interface{}{
					"actionId":     ra.ActionID,
					"relationship": ra.Relationship,
				}
				if ra.OffsetDuration != nil {
					rel["offsetDuration"] = ra.OffsetDuration
				}
				related = append(related, rel)
			}
			rgAction["relatedAction"] = related
		}

		// Process nested actions
		if len(action.Action) > 0 {
			subActions := make([]interface{}, 0)
			processApplyActions(action.Action, req, result, &subActions)
			if len(subActions) > 0 {
				rgAction["action"] = subActions
			}
		}

		// Generate activity for DefinitionURI
		if action.DefinitionURI != "" {
			activity := generateApplyActivity(action, req)
			result.Activities = append(result.Activities, activity)
			rgAction["resource"] = map[string]interface{}{
				"reference": fmt.Sprintf("%s/%s", activity.Resource["resourceType"], activity.Resource["id"]),
			}
		}

		*rgActions = append(*rgActions, rgAction)
	}
}

// generateApplyActivity creates a resource from an action's definition reference.
func generateApplyActivity(action ApplyPlanDefinitionAction, req *ApplyRequest) ApplyActivity {
	resourceID := uuid.New().String()

	// Determine resource type from the definition URI
	resourceType := "ServiceRequest"
	if strings.Contains(action.DefinitionURI, "Medication") {
		resourceType = "MedicationRequest"
	} else if strings.Contains(action.DefinitionURI, "Task") {
		resourceType = "Task"
	} else if strings.Contains(action.DefinitionURI, "Communication") {
		resourceType = "CommunicationRequest"
	}

	resource := map[string]interface{}{
		"resourceType": resourceType,
		"id":           resourceID,
		"status":       "draft",
	}

	// Set subject/for based on resource type
	switch resourceType {
	case "Task":
		resource["intent"] = "order"
		resource["for"] = map[string]interface{}{"reference": req.Subject}
	case "CommunicationRequest":
		resource["subject"] = map[string]interface{}{"reference": req.Subject}
	default:
		resource["intent"] = "order"
		resource["subject"] = map[string]interface{}{"reference": req.Subject}
	}

	// Set encounter
	if req.Encounter != "" {
		resource["encounter"] = map[string]interface{}{"reference": req.Encounter}
	}

	// Set practitioner
	if req.Practitioner != "" {
		switch resourceType {
		case "MedicationRequest":
			resource["requester"] = map[string]interface{}{"reference": req.Practitioner}
		case "ServiceRequest":
			resource["requester"] = map[string]interface{}{"reference": req.Practitioner}
		case "Task":
			resource["requester"] = map[string]interface{}{"reference": req.Practitioner}
		}
	}

	// Set organization
	if req.Organization != "" {
		resource["performer"] = map[string]interface{}{"reference": req.Organization}
	}

	if action.Title != "" {
		resource["description"] = action.Title
	}

	return ApplyActivity{
		ResourceType: resourceType,
		Resource:     resource,
		ActionID:     action.ID,
	}
}

// evaluateApplyConditions checks applicability conditions.
// Returns true if all applicability conditions evaluate to true.
// Non-applicability conditions (start, stop) are ignored.
func evaluateApplyConditions(conditions []ApplyActionCondition) bool {
	for _, cond := range conditions {
		if cond.Kind != "applicability" {
			continue
		}
		if cond.Expression == "" {
			continue
		}
		// Simple evaluation: "true" / "false" literals.
		// In production, this would use a FHIRPath engine.
		expr := strings.TrimSpace(strings.ToLower(cond.Expression))
		if expr == "false" {
			return false
		}
	}
	return true
}

// ApplyActivityDefinition applies an ActivityDefinition to generate a single resource.
func ApplyActivityDefinition(actDef *ParsedActivityDefinition, req *ApplyRequest) (*ApplyActivity, error) {
	if actDef == nil {
		return nil, fmt.Errorf("ActivityDefinition is nil")
	}
	if req == nil {
		return nil, fmt.Errorf("ApplyRequest is nil")
	}
	if actDef.Status == "retired" {
		return nil, fmt.Errorf("cannot apply retired ActivityDefinition %s", actDef.ID)
	}

	resourceID := uuid.New().String()
	resource := map[string]interface{}{
		"id":     resourceID,
		"status": "draft",
	}

	kind := actDef.Kind
	resource["resourceType"] = kind

	switch kind {
	case "MedicationRequest":
		resource["intent"] = "order"
		resource["subject"] = map[string]interface{}{"reference": req.Subject}
		if actDef.Code != nil {
			resource["medicationCodeableConcept"] = actDef.Code
		}
		if actDef.Dosage != nil {
			resource["dosageInstruction"] = actDef.Dosage
		}
		if req.Practitioner != "" {
			resource["requester"] = map[string]interface{}{"reference": req.Practitioner}
		}
	case "ServiceRequest":
		resource["intent"] = "order"
		resource["subject"] = map[string]interface{}{"reference": req.Subject}
		if actDef.Code != nil {
			resource["code"] = actDef.Code
		}
		if req.Practitioner != "" {
			resource["requester"] = map[string]interface{}{"reference": req.Practitioner}
		}
	case "Task":
		resource["intent"] = "order"
		resource["for"] = map[string]interface{}{"reference": req.Subject}
		if actDef.Code != nil {
			resource["code"] = actDef.Code
		}
		if req.Practitioner != "" {
			resource["requester"] = map[string]interface{}{"reference": req.Practitioner}
		}
	case "CommunicationRequest":
		resource["subject"] = map[string]interface{}{"reference": req.Subject}
	default:
		resource["subject"] = map[string]interface{}{"reference": req.Subject}
	}

	// Set encounter context
	if req.Encounter != "" {
		resource["encounter"] = map[string]interface{}{"reference": req.Encounter}
	}

	// Set organization context
	if req.Organization != "" {
		resource["performer"] = map[string]interface{}{"reference": req.Organization}
	}

	if actDef.Title != "" {
		resource["description"] = actDef.Title
	}

	return &ApplyActivity{
		ResourceType: kind,
		Resource:     resource,
		ActionID:     actDef.ID,
	}, nil
}

// ============================================================================
// Validation
// ============================================================================

// ValidateApplyRequest validates the $apply request parameters.
func ValidateApplyRequest(req *ApplyRequest) []ValidationIssue {
	var issues []ValidationIssue

	if req == nil {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Diagnostics: "ApplyRequest is nil",
		})
		return issues
	}

	if req.PlanDefinitionID == "" {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Diagnostics: "PlanDefinition ID or planDefinitionID is required",
		})
	}

	if req.Subject == "" {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Diagnostics: "subject is required",
		})
	} else if !strings.Contains(req.Subject, "/") {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeValue,
			Diagnostics: "subject must be a FHIR reference (e.g. Patient/123)",
		})
	}

	return issues
}

// ============================================================================
// HTTP Handlers
// ============================================================================

// ApplyOpResourceResolver is the interface for resolving resources during $apply.
// It uses a generic interface{} context parameter to avoid conflicts with the
// existing ResourceResolver defined in document_op.go.
type ApplyOpResourceResolver interface {
	ResolveReference(ctx interface{}, reference string) (map[string]interface{}, error)
}

// ApplyHandler returns an echo.HandlerFunc for POST /fhir/PlanDefinition/:id/$apply.
func ApplyHandler(resolver ApplyOpResourceResolver) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, ErrorOutcome("PlanDefinition ID is required"))
		}

		// Resolve the PlanDefinition
		ref := "PlanDefinition/" + id
		pdData, err := resolver.ResolveReference(c.Request().Context(), ref)
		if err != nil || pdData == nil {
			return c.JSON(http.StatusNotFound, NotFoundOutcome("PlanDefinition", id))
		}

		// Parse the PlanDefinition
		planDef, err := ParseApplyPlanDefinition(pdData)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
		}

		// Read request body
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorOutcome("failed to read request body"))
		}
		if len(body) == 0 {
			return c.JSON(http.StatusBadRequest, ErrorOutcome("request body is empty"))
		}

		var bodyMap map[string]interface{}
		if err := json.Unmarshal(body, &bodyMap); err != nil {
			return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid JSON: "+err.Error()))
		}

		// Parse the apply request from the body
		applyReq := parseApplyRequestFromBody(bodyMap)
		applyReq.PlanDefinitionID = id

		// Validate
		if applyReq.Subject == "" {
			return c.JSON(http.StatusBadRequest, ErrorOutcome("subject parameter is required"))
		}

		// Apply
		result, err := ApplyPlanDefinition(planDef, applyReq)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
		}

		// Build response bundle
		entries := make([]interface{}, 0)
		if result.CarePlan != nil {
			entries = append(entries, result.CarePlan)
		}
		if result.RequestGroup != nil {
			entries = append(entries, result.RequestGroup)
		}
		for _, act := range result.Activities {
			entries = append(entries, act.Resource)
		}

		bundle := NewSearchBundle(entries, len(entries), "/fhir/PlanDefinition/"+id+"/$apply")
		return c.JSON(http.StatusOK, bundle)
	}
}

// ActivityDefinitionApplyHandler returns an echo.HandlerFunc for POST /fhir/ActivityDefinition/:id/$apply.
func ActivityDefinitionApplyHandler(resolver ApplyOpResourceResolver) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, ErrorOutcome("ActivityDefinition ID is required"))
		}

		// Resolve the ActivityDefinition
		ref := "ActivityDefinition/" + id
		adData, err := resolver.ResolveReference(c.Request().Context(), ref)
		if err != nil || adData == nil {
			return c.JSON(http.StatusNotFound, NotFoundOutcome("ActivityDefinition", id))
		}

		// Parse the ActivityDefinition
		actDef, err := ParseApplyActivityDefinition(adData)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
		}

		// Read request body
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorOutcome("failed to read request body"))
		}
		if len(body) == 0 {
			return c.JSON(http.StatusBadRequest, ErrorOutcome("request body is empty"))
		}

		var bodyMap map[string]interface{}
		if err := json.Unmarshal(body, &bodyMap); err != nil {
			return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid JSON: "+err.Error()))
		}

		// Parse the apply request from the body
		applyReq := parseApplyRequestFromBody(bodyMap)

		// Validate
		if applyReq.Subject == "" {
			return c.JSON(http.StatusBadRequest, ErrorOutcome("subject parameter is required"))
		}

		// Apply
		activity, err := ApplyActivityDefinition(actDef, applyReq)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
		}

		return c.JSON(http.StatusOK, activity.Resource)
	}
}

// parseApplyRequestFromBody extracts ApplyRequest fields from a request body.
// It supports both simple JSON format and FHIR Parameters resource format.
func parseApplyRequestFromBody(body map[string]interface{}) *ApplyRequest {
	req := &ApplyRequest{
		Parameters: make(map[string]interface{}),
	}

	// Check if this is a FHIR Parameters resource
	if rt, ok := body["resourceType"].(string); ok && rt == "Parameters" {
		if params, ok := body["parameter"].([]interface{}); ok {
			for _, p := range params {
				param, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				name, _ := param["name"].(string)
				switch name {
				case "subject":
					if vs, ok := param["valueString"].(string); ok {
						req.Subject = vs
					} else if ref, ok := param["valueReference"].(map[string]interface{}); ok {
						req.Subject, _ = ref["reference"].(string)
					}
				case "encounter":
					if vs, ok := param["valueString"].(string); ok {
						req.Encounter = vs
					} else if ref, ok := param["valueReference"].(map[string]interface{}); ok {
						req.Encounter, _ = ref["reference"].(string)
					}
				case "practitioner":
					if vs, ok := param["valueString"].(string); ok {
						req.Practitioner = vs
					} else if ref, ok := param["valueReference"].(map[string]interface{}); ok {
						req.Practitioner, _ = ref["reference"].(string)
					}
				case "organization":
					if vs, ok := param["valueString"].(string); ok {
						req.Organization = vs
					} else if ref, ok := param["valueReference"].(map[string]interface{}); ok {
						req.Organization, _ = ref["reference"].(string)
					}
				}
			}
		}
		return req
	}

	// Simple JSON format
	if subject, ok := body["subject"].(string); ok {
		req.Subject = subject
	}
	if encounter, ok := body["encounter"].(string); ok {
		req.Encounter = encounter
	}
	if practitioner, ok := body["practitioner"].(string); ok {
		req.Practitioner = practitioner
	}
	if organization, ok := body["organization"].(string); ok {
		req.Organization = organization
	}
	if params, ok := body["parameters"].(map[string]interface{}); ok {
		req.Parameters = params
	}

	return req
}
