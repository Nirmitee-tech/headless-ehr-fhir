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
// PlanDefinition Model
// ============================================================================

// PlanDefinition represents a FHIR R4 PlanDefinition resource.
type PlanDefinition struct {
	ID          string       `json:"id"`
	URL         string       `json:"url,omitempty"`
	Version     string       `json:"version,omitempty"`
	Name        string       `json:"name,omitempty"`
	Title       string       `json:"title,omitempty"`
	Status      string       `json:"status"`
	Type        string       `json:"type"`
	SubjectType string       `json:"subjectType,omitempty"`
	Goal        []PlanGoal   `json:"goal,omitempty"`
	Action      []PlanAction `json:"action,omitempty"`
}

// PlanGoal describes a goal within a PlanDefinition.
type PlanGoal struct {
	Description string       `json:"description"`
	Priority    string       `json:"priority,omitempty"`
	Target      []GoalTarget `json:"target,omitempty"`
}

// GoalTarget specifies a measurable target for a goal.
type GoalTarget struct {
	Measure     string `json:"measure,omitempty"`
	DetailValue string `json:"detailValue,omitempty"`
	DueDuration string `json:"dueDuration,omitempty"`
}

// PlanAction is a recursive action tree within a PlanDefinition.
type PlanAction struct {
	ID                  string          `json:"id,omitempty"`
	Title               string          `json:"title,omitempty"`
	Description         string          `json:"description,omitempty"`
	Priority            string          `json:"priority,omitempty"`
	Condition           []PlanCondition `json:"condition,omitempty"`
	Trigger             []PlanTrigger   `json:"trigger,omitempty"`
	Input               []DataRequirement `json:"input,omitempty"`
	Output              []DataRequirement `json:"output,omitempty"`
	RelatedAction       []RelatedAction `json:"relatedAction,omitempty"`
	Type                string          `json:"type,omitempty"`
	DefinitionCanonical string          `json:"definitionCanonical,omitempty"`
	DynamicValue        []DynamicValue  `json:"dynamicValue,omitempty"`
	SelectionBehavior   string          `json:"selectionBehavior,omitempty"`
	GroupingBehavior    string          `json:"groupingBehavior,omitempty"`
	Action              []PlanAction    `json:"action,omitempty"`
}

// PlanCondition is a condition that determines action applicability.
type PlanCondition struct {
	Kind       string `json:"kind"`
	Expression string `json:"expression"`
}

// PlanTrigger describes an event that triggers the action.
type PlanTrigger struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// DataRequirement describes data needed by or produced by an action.
type DataRequirement struct {
	Type    string `json:"type"`
	Profile string `json:"profile,omitempty"`
}

// RelatedAction describes ordering relationships between actions.
type RelatedAction struct {
	ActionID       string `json:"actionId"`
	Relationship   string `json:"relationship"`
	OffsetDuration string `json:"offsetDuration,omitempty"`
}

// DynamicValue specifies a computed field value using a FHIRPath expression.
type DynamicValue struct {
	Path       string `json:"path"`
	Expression string `json:"expression"`
}

// ============================================================================
// ActivityDefinition Model
// ============================================================================

// ActivityDefinition represents a FHIR R4 ActivityDefinition.
type ActivityDefinition struct {
	ID           string         `json:"id"`
	URL          string         `json:"url,omitempty"`
	Name         string         `json:"name,omitempty"`
	Title        string         `json:"title,omitempty"`
	Status       string         `json:"status"`
	Kind         string         `json:"kind"`
	Code         *ActivityCode  `json:"code,omitempty"`
	Product      string         `json:"product,omitempty"`
	Dosage       string         `json:"dosage,omitempty"`
	Timing       string         `json:"timing,omitempty"`
	DynamicValue []DynamicValue `json:"dynamicValue,omitempty"`
}

// ActivityCode is a coding triple for what to order.
type ActivityCode struct {
	System  string `json:"system,omitempty"`
	Code    string `json:"code"`
	Display string `json:"display,omitempty"`
}

// ============================================================================
// Apply Result
// ============================================================================

// ApplyResult is the output of PlanDefinitionEngine.Apply.
type ApplyResult struct {
	CarePlan     map[string]interface{}   `json:"carePlan,omitempty"`
	RequestGroup map[string]interface{}   `json:"requestGroup,omitempty"`
	Resources    []map[string]interface{} `json:"resources,omitempty"`
}

// ============================================================================
// PlanDefinition Engine
// ============================================================================

// PlanDefinitionEngine evaluates PlanDefinitions against a subject and
// produces patient-specific CarePlans, RequestGroups and request resources.
type PlanDefinitionEngine struct {
	fhirpath            *FHIRPathEngine
	mu                  sync.RWMutex
	planDefinitions     map[string]*PlanDefinition
	activityDefinitions map[string]*ActivityDefinition
}

// NewPlanDefinitionEngine creates a new engine backed by a FHIRPath engine.
func NewPlanDefinitionEngine(fhirpath *FHIRPathEngine) *PlanDefinitionEngine {
	return &PlanDefinitionEngine{
		fhirpath:            fhirpath,
		planDefinitions:     make(map[string]*PlanDefinition),
		activityDefinitions: make(map[string]*ActivityDefinition),
	}
}

// RegisterPlanDefinition stores a PlanDefinition for lookup.
func (e *PlanDefinitionEngine) RegisterPlanDefinition(pd *PlanDefinition) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.planDefinitions[pd.ID] = pd
}

// RegisterActivityDefinition stores an ActivityDefinition for lookup.
func (e *PlanDefinitionEngine) RegisterActivityDefinition(ad *ActivityDefinition) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.activityDefinitions[ad.ID] = ad
	e.activityDefinitions["ActivityDefinition/"+ad.ID] = ad
}

// GetPlanDefinition returns a PlanDefinition by ID.
func (e *PlanDefinitionEngine) GetPlanDefinition(id string) *PlanDefinition {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.planDefinitions[id]
}

// GetActivityDefinition returns an ActivityDefinition by ID.
func (e *PlanDefinitionEngine) GetActivityDefinition(id string) *ActivityDefinition {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.activityDefinitions[id]
}

// ListPlanDefinitions returns all stored PlanDefinitions.
func (e *PlanDefinitionEngine) ListPlanDefinitions() []*PlanDefinition {
	e.mu.RLock()
	defer e.mu.RUnlock()
	result := make([]*PlanDefinition, 0, len(e.planDefinitions))
	for _, pd := range e.planDefinitions {
		result = append(result, pd)
	}
	return result
}

// ListActivityDefinitions returns all stored ActivityDefinitions.
func (e *PlanDefinitionEngine) ListActivityDefinitions() []*ActivityDefinition {
	e.mu.RLock()
	defer e.mu.RUnlock()
	seen := make(map[string]bool)
	result := make([]*ActivityDefinition, 0)
	for _, ad := range e.activityDefinitions {
		if !seen[ad.ID] {
			seen[ad.ID] = true
			result = append(result, ad)
		}
	}
	return result
}

// DeletePlanDefinition removes a PlanDefinition by ID.
func (e *PlanDefinitionEngine) DeletePlanDefinition(id string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, ok := e.planDefinitions[id]; ok {
		delete(e.planDefinitions, id)
		return true
	}
	return false
}

// Apply evaluates a PlanDefinition against a subject and returns an ApplyResult.
func (e *PlanDefinitionEngine) Apply(ctx context.Context, plan *PlanDefinition, subject map[string]interface{}, params map[string]interface{}) (*ApplyResult, error) {
	if subject == nil {
		return nil, fmt.Errorf("subject is required for $apply")
	}
	if plan.Status == "retired" {
		return nil, fmt.Errorf("cannot apply retired PlanDefinition %s", plan.ID)
	}

	result := &ApplyResult{
		Resources: make([]map[string]interface{}, 0),
	}

	subjectRef := ""
	if rt, ok := subject["resourceType"].(string); ok {
		if id, ok := subject["id"].(string); ok {
			subjectRef = rt + "/" + id
		}
	}

	// Process actions -- take a read lock on the activity definitions
	e.mu.RLock()
	rgActions := make([]interface{}, 0)
	e.processActions(ctx, plan.Action, subject, params, result, &rgActions)
	e.mu.RUnlock()

	// Build CarePlan
	cpID := uuid.New().String()
	carePlan := map[string]interface{}{
		"resourceType": "CarePlan",
		"id":           cpID,
		"status":       "active",
		"intent":       "plan",
		"subject":      map[string]interface{}{"reference": subjectRef},
		"created":      time.Now().UTC().Format(time.RFC3339),
	}
	if plan.Title != "" {
		carePlan["title"] = plan.Title
	}
	if plan.Name != "" {
		carePlan["description"] = plan.Name
	}
	if len(plan.Goal) > 0 {
		goals := make([]interface{}, 0, len(plan.Goal))
		for _, g := range plan.Goal {
			goal := map[string]interface{}{
				"description": map[string]interface{}{"text": g.Description},
			}
			if g.Priority != "" {
				goal["priority"] = map[string]interface{}{
					"coding": []interface{}{map[string]interface{}{"code": g.Priority}},
				}
			}
			goals = append(goals, goal)
		}
		carePlan["goal"] = goals
	}
	if len(result.Resources) > 0 {
		activities := make([]interface{}, 0, len(result.Resources))
		for _, r := range result.Resources {
			activities = append(activities, map[string]interface{}{
				"reference": map[string]interface{}{
					"reference": fmt.Sprintf("%s/%s", r["resourceType"], r["id"]),
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
		"subject":      map[string]interface{}{"reference": subjectRef},
		"authoredOn":   time.Now().UTC().Format(time.RFC3339),
	}
	if len(rgActions) > 0 {
		requestGroup["action"] = rgActions
	}
	result.RequestGroup = requestGroup

	return result, nil
}

// processActions recursively evaluates actions (caller must hold e.mu.RLock).
func (e *PlanDefinitionEngine) processActions(
	ctx context.Context,
	actions []PlanAction,
	subject map[string]interface{},
	params map[string]interface{},
	result *ApplyResult,
	rgActions *[]interface{},
) {
	for _, action := range actions {
		if !e.evaluateConditions(action.Condition, subject, params) {
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

		// Related actions
		if len(action.RelatedAction) > 0 {
			related := make([]interface{}, 0, len(action.RelatedAction))
			for _, ra := range action.RelatedAction {
				rel := map[string]interface{}{
					"actionId":     ra.ActionID,
					"relationship": ra.Relationship,
				}
				if ra.OffsetDuration != "" {
					rel["offsetDuration"] = ra.OffsetDuration
				}
				related = append(related, rel)
			}
			rgAction["relatedAction"] = related
		}

		// Process nested actions
		if len(action.Action) > 0 {
			subActions := make([]interface{}, 0)
			e.processActions(ctx, action.Action, subject, params, result, &subActions)
			if len(subActions) > 0 {
				rgAction["action"] = subActions
			}
		}

		// Resolve ActivityDefinition
		if action.DefinitionCanonical != "" {
			resource := e.resolveActivityDefinition(action.DefinitionCanonical, subject, params, action.DynamicValue)
			if resource != nil {
				result.Resources = append(result.Resources, resource)
				rgAction["resource"] = map[string]interface{}{
					"reference": fmt.Sprintf("%s/%s", resource["resourceType"], resource["id"]),
				}
			}
		}

		if action.Type != "" {
			rgAction["type"] = map[string]interface{}{
				"coding": []interface{}{map[string]interface{}{"code": action.Type}},
			}
		}

		*rgActions = append(*rgActions, rgAction)
	}
}

// evaluateConditions checks applicability conditions.
func (e *PlanDefinitionEngine) evaluateConditions(conditions []PlanCondition, subject map[string]interface{}, params map[string]interface{}) bool {
	for _, cond := range conditions {
		if cond.Kind != "applicability" {
			continue
		}
		if cond.Expression == "" {
			continue
		}
		expression := cond.Expression
		if params != nil {
			expression = pdSubstituteParams(expression, params)
		}
		ok, err := e.fhirpath.EvaluateBool(subject, expression)
		if err != nil || !ok {
			return false
		}
	}
	return true
}

// pdSubstituteParams replaces %param_name in expressions with values.
func pdSubstituteParams(expression string, params map[string]interface{}) string {
	for key, value := range params {
		placeholder := "%" + key
		replacement := fmt.Sprintf("%v", value)
		expression = strings.ReplaceAll(expression, placeholder, replacement)
	}
	return expression
}

// resolveActivityDefinition looks up an AD and generates a request resource.
// Caller must hold e.mu.RLock.
func (e *PlanDefinitionEngine) resolveActivityDefinition(
	canonical string,
	subject map[string]interface{},
	params map[string]interface{},
	actionDynamicValues []DynamicValue,
) map[string]interface{} {
	adID := canonical
	if strings.HasPrefix(canonical, "ActivityDefinition/") {
		adID = strings.TrimPrefix(canonical, "ActivityDefinition/")
	}

	ad := e.activityDefinitions[adID]
	if ad == nil {
		ad = e.activityDefinitions[canonical]
	}
	if ad == nil {
		return nil
	}

	subjectRef := ""
	if rt, ok := subject["resourceType"].(string); ok {
		if id, ok := subject["id"].(string); ok {
			subjectRef = rt + "/" + id
		}
	}

	resourceID := uuid.New().String()
	resource := map[string]interface{}{
		"id":     resourceID,
		"status": "draft",
	}

	switch ad.Kind {
	case "MedicationRequest":
		resource["resourceType"] = "MedicationRequest"
		resource["intent"] = "order"
		resource["subject"] = map[string]interface{}{"reference": subjectRef}
		if ad.Code != nil {
			resource["medicationCodeableConcept"] = map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{"system": ad.Code.System, "code": ad.Code.Code, "display": ad.Code.Display},
				},
			}
		}
		if ad.Dosage != "" {
			resource["dosageInstruction"] = []interface{}{map[string]interface{}{"text": ad.Dosage}}
		}
	case "ServiceRequest":
		resource["resourceType"] = "ServiceRequest"
		resource["intent"] = "order"
		resource["subject"] = map[string]interface{}{"reference": subjectRef}
		if ad.Code != nil {
			resource["code"] = map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{"system": ad.Code.System, "code": ad.Code.Code, "display": ad.Code.Display},
				},
			}
		}
	case "Task":
		resource["resourceType"] = "Task"
		resource["intent"] = "order"
		resource["for"] = map[string]interface{}{"reference": subjectRef}
		if ad.Code != nil {
			resource["code"] = map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{"system": ad.Code.System, "code": ad.Code.Code, "display": ad.Code.Display},
				},
			}
		}
	case "CommunicationRequest":
		resource["resourceType"] = "CommunicationRequest"
		resource["subject"] = map[string]interface{}{"reference": subjectRef}
	default:
		resource["resourceType"] = ad.Kind
		resource["subject"] = map[string]interface{}{"reference": subjectRef}
	}

	if ad.Title != "" {
		resource["description"] = ad.Title
	}

	// Apply DynamicValues from ActivityDefinition
	e.applyDynamicValues(resource, ad.DynamicValue, subject, params)
	// Apply DynamicValues from action (overrides)
	e.applyDynamicValues(resource, actionDynamicValues, subject, params)

	return resource
}

// applyDynamicValues evaluates FHIRPath expressions and sets computed fields.
func (e *PlanDefinitionEngine) applyDynamicValues(
	resource map[string]interface{},
	dynamicValues []DynamicValue,
	subject map[string]interface{},
	params map[string]interface{},
) {
	for _, dv := range dynamicValues {
		if dv.Path == "" || dv.Expression == "" {
			continue
		}
		expression := dv.Expression
		if params != nil {
			expression = pdSubstituteParams(expression, params)
		}
		res, err := e.fhirpath.Evaluate(subject, expression)
		if err != nil || len(res) == 0 {
			continue
		}
		resource[dv.Path] = res[0]
	}
}

// ============================================================================
// Built-in PlanDefinitions
// ============================================================================

// RegisterBuiltins registers all built-in clinical protocol PlanDefinitions.
func (e *PlanDefinitionEngine) RegisterBuiltins() {
	e.registerDiabetesManagement()
	e.registerSepsisBundle()
	e.registerCHFDischarge()
	e.registerPreventiveCare()
}

func (e *PlanDefinitionEngine) registerDiabetesManagement() {
	e.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-diabetes-hba1c-lab", Name: "HbA1cLabOrder", Title: "HbA1c Laboratory Test",
		Status: "active", Kind: "ServiceRequest",
		Code: &ActivityCode{System: "http://loinc.org", Code: "4548-4", Display: "HbA1c"},
	})
	e.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-diabetes-metformin", Name: "MetforminOrder", Title: "Metformin 500mg",
		Status: "active", Kind: "MedicationRequest", Product: "Metformin hydrochloride",
		Dosage: "500mg twice daily with meals",
		Code:   &ActivityCode{System: "http://www.nlm.nih.gov/research/umls/rxnorm", Code: "860975", Display: "Metformin 500 MG Oral Tablet"},
	})
	e.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-diabetes-followup", Name: "DiabetesFollowUp", Title: "Follow-up appointment in 3 months",
		Status: "active", Kind: "ServiceRequest",
		Code: &ActivityCode{System: "http://snomed.info/sct", Code: "185389009", Display: "Follow-up visit"},
	})
	e.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-diabetes-goal-task", Name: "DiabetesGoalTask", Title: "Create glycemic control goal",
		Status: "active", Kind: "Task",
		Code: &ActivityCode{System: "http://snomed.info/sct", Code: "698472009", Display: "Glycemic control care plan goal"},
	})
	e.RegisterPlanDefinition(&PlanDefinition{
		ID: "diabetes-management", URL: "http://example.org/fhir/PlanDefinition/diabetes-management",
		Version: "1.0", Name: "DiabetesManagementProtocol", Title: "Diabetes Management Protocol",
		Status: "active", Type: "clinical-protocol", SubjectType: "Patient",
		Goal: []PlanGoal{{
			Description: "Achieve HbA1c < 7%", Priority: "high-priority",
			Target: []GoalTarget{{Measure: "HbA1c", DetailValue: "< 7%", DueDuration: "6 months"}},
		}},
		Action: []PlanAction{
			{ID: "dm-action-hba1c", Title: "Order HbA1c Lab", Description: "Order HbA1c laboratory test",
				Priority: "routine", Type: "create",
				Condition: []PlanCondition{{Kind: "applicability", Expression: "%hba1c_value > 9"}},
				DefinitionCanonical: "ActivityDefinition/ad-diabetes-hba1c-lab"},
			{ID: "dm-action-metformin", Title: "Start Metformin", Description: "Initiate metformin therapy",
				Priority: "urgent", Type: "create",
				Condition: []PlanCondition{{Kind: "applicability", Expression: "%hba1c_value > 9"}},
				DefinitionCanonical: "ActivityDefinition/ad-diabetes-metformin"},
			{ID: "dm-action-followup", Title: "Schedule Follow-up", Description: "Schedule follow-up in 3 months",
				Type: "create",
				Condition: []PlanCondition{{Kind: "applicability", Expression: "%hba1c_value > 9"}},
				RelatedAction: []RelatedAction{{ActionID: "dm-action-metformin", Relationship: "after-start"}},
				DefinitionCanonical: "ActivityDefinition/ad-diabetes-followup"},
			{ID: "dm-action-goal", Title: "Create Care Plan Goal", Description: "Create glycemic control goal",
				Type: "create",
				Condition: []PlanCondition{{Kind: "applicability", Expression: "%hba1c_value > 9"}},
				DefinitionCanonical: "ActivityDefinition/ad-diabetes-goal-task"},
		},
	})
}

func (e *PlanDefinitionEngine) registerSepsisBundle() {
	e.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-sepsis-blood-culture", Name: "BloodCultureOrder", Title: "Blood Cultures x2",
		Status: "active", Kind: "ServiceRequest",
		Code: &ActivityCode{System: "http://loinc.org", Code: "600-7", Display: "Blood culture"},
	})
	e.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-sepsis-antibiotics", Name: "BroadSpectrumAntibiotics", Title: "Broad-spectrum antibiotics",
		Status: "active", Kind: "MedicationRequest", Dosage: "Per protocol",
		Code: &ActivityCode{System: "http://www.nlm.nih.gov/research/umls/rxnorm", Code: "1665060", Display: "Piperacillin/Tazobactam"},
	})
	e.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-sepsis-iv-fluids", Name: "IVFluidBolus", Title: "IV Fluid Bolus 30mL/kg",
		Status: "active", Kind: "MedicationRequest", Dosage: "30mL/kg crystalloid bolus",
		Code: &ActivityCode{System: "http://www.nlm.nih.gov/research/umls/rxnorm", Code: "313002", Display: "Normal Saline 0.9%"},
	})
	e.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-sepsis-lactate", Name: "LactateLab", Title: "Serum Lactate",
		Status: "active", Kind: "ServiceRequest",
		Code: &ActivityCode{System: "http://loinc.org", Code: "2524-7", Display: "Lactate"},
	})
	e.RegisterPlanDefinition(&PlanDefinition{
		ID: "sepsis-bundle-sep1", URL: "http://example.org/fhir/PlanDefinition/sepsis-bundle-sep1",
		Version: "1.0", Name: "SepsisBundleSEP1", Title: "Sepsis Bundle (SEP-1)",
		Status: "active", Type: "order-set", SubjectType: "Patient",
		Goal: []PlanGoal{{Description: "Complete sepsis bundle within 3 hours", Priority: "high-priority"}},
		Action: []PlanAction{
			{ID: "sep-action-cultures", Title: "Order Blood Cultures", Description: "Obtain blood cultures before antibiotics",
				Priority: "stat", Type: "create",
				Condition: []PlanCondition{{Kind: "applicability", Expression: "%temperature > 38.3"}},
				DefinitionCanonical: "ActivityDefinition/ad-sepsis-blood-culture"},
			{ID: "sep-action-abx", Title: "Start Broad-spectrum Antibiotics", Description: "Administer within 1 hour",
				Priority: "stat", Type: "create",
				Condition: []PlanCondition{{Kind: "applicability", Expression: "%temperature > 38.3"}},
				RelatedAction: []RelatedAction{{ActionID: "sep-action-cultures", Relationship: "after-start"}},
				DefinitionCanonical: "ActivityDefinition/ad-sepsis-antibiotics"},
			{ID: "sep-action-fluids", Title: "IV Fluid Bolus", Description: "30mL/kg crystalloid bolus",
				Type: "create",
				Condition: []PlanCondition{{Kind: "applicability", Expression: "%temperature > 38.3"}},
				DefinitionCanonical: "ActivityDefinition/ad-sepsis-iv-fluids"},
			{ID: "sep-action-lactate", Title: "Order Lactate", Description: "Measure serum lactate",
				Priority: "stat", Type: "create",
				Condition: []PlanCondition{{Kind: "applicability", Expression: "%temperature > 38.3"}},
				DefinitionCanonical: "ActivityDefinition/ad-sepsis-lactate"},
		},
	})
}

func (e *PlanDefinitionEngine) registerCHFDischarge() {
	e.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-chf-followup", Name: "CHFFollowUp", Title: "Follow-up within 7 days",
		Status: "active", Kind: "ServiceRequest",
		Code: &ActivityCode{System: "http://snomed.info/sct", Code: "185389009", Display: "Follow-up visit within 7 days"},
	})
	e.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-chf-ace-inhibitor", Name: "ACEInhibitor", Title: "ACE Inhibitor - Lisinopril",
		Status: "active", Kind: "MedicationRequest", Dosage: "10mg daily", Product: "Lisinopril",
		Code: &ActivityCode{System: "http://www.nlm.nih.gov/research/umls/rxnorm", Code: "314076", Display: "Lisinopril 10 MG Oral Tablet"},
	})
	e.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-chf-daily-weight", Name: "DailyWeightMonitoring", Title: "Daily weight monitoring",
		Status: "active", Kind: "Task",
		Code: &ActivityCode{System: "http://loinc.org", Code: "29463-7", Display: "Daily weight monitoring"},
	})
	e.RegisterPlanDefinition(&PlanDefinition{
		ID: "chf-discharge-protocol", URL: "http://example.org/fhir/PlanDefinition/chf-discharge-protocol",
		Version: "1.0", Name: "CHFDischargeProtocol", Title: "CHF Discharge Protocol",
		Status: "active", Type: "clinical-protocol", SubjectType: "Patient",
		Goal: []PlanGoal{{Description: "Prevent heart failure readmission within 30 days", Priority: "high-priority"}},
		Action: []PlanAction{
			{ID: "chf-action-followup", Title: "Schedule Follow-up Within 7 Days",
				Description: "Outpatient follow-up within 7 days of discharge", Type: "create",
				Condition: []PlanCondition{{Kind: "applicability", Expression: "%discharge = true"}},
				DefinitionCanonical: "ActivityDefinition/ad-chf-followup"},
			{ID: "chf-action-ace", Title: "Prescribe ACE Inhibitor",
				Description: "Start or continue ACE inhibitor therapy", Type: "create",
				Condition: []PlanCondition{{Kind: "applicability", Expression: "%discharge = true"}},
				DefinitionCanonical: "ActivityDefinition/ad-chf-ace-inhibitor"},
			{ID: "chf-action-weight", Title: "Daily Weight Monitoring",
				Description: "Patient to monitor weight daily", Type: "create",
				Condition: []PlanCondition{{Kind: "applicability", Expression: "%discharge = true"}},
				DefinitionCanonical: "ActivityDefinition/ad-chf-daily-weight"},
		},
	})
}

func (e *PlanDefinitionEngine) registerPreventiveCare() {
	e.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-prev-colonoscopy", Name: "ColonoscopyScreening", Title: "Colonoscopy screening",
		Status: "active", Kind: "ServiceRequest",
		Code: &ActivityCode{System: "http://snomed.info/sct", Code: "73761001", Display: "Colonoscopy"},
	})
	e.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-prev-mammogram", Name: "MammogramScreening", Title: "Mammogram screening",
		Status: "active", Kind: "ServiceRequest",
		Code: &ActivityCode{System: "http://snomed.info/sct", Code: "71651007", Display: "Mammogram"},
	})
	e.RegisterPlanDefinition(&PlanDefinition{
		ID: "preventive-care-screening", URL: "http://example.org/fhir/PlanDefinition/preventive-care-screening",
		Version: "1.0", Name: "PreventiveCareScreening", Title: "Preventive Care Screening Recommendations",
		Status: "active", Type: "eca-rule", SubjectType: "Patient",
		Goal: []PlanGoal{{Description: "Age-appropriate cancer screening", Priority: "routine"}},
		Action: []PlanAction{
			{ID: "prev-action-colonoscopy", Title: "Colonoscopy Screening",
				Description: "Recommend colonoscopy for adults >= 45", Type: "create",
				Condition: []PlanCondition{{Kind: "applicability", Expression: "%age >= 45"}},
				DefinitionCanonical: "ActivityDefinition/ad-prev-colonoscopy"},
			{ID: "prev-action-mammogram", Title: "Mammogram Screening",
				Description: "Recommend mammogram for females >= 40", Type: "create",
				Condition: []PlanCondition{{Kind: "applicability", Expression: "%age >= 40"}},
				DefinitionCanonical: "ActivityDefinition/ad-prev-mammogram"},
		},
	})
}

// ============================================================================
// PlanDefinition Handler (HTTP)
// ============================================================================

// PlanDefinitionHandler provides FHIR REST endpoints for PlanDefinition,
// ActivityDefinition, and the $apply operation.
type PlanDefinitionHandler struct {
	engine *PlanDefinitionEngine
}

// NewPlanDefinitionHandler creates a new handler.
func NewPlanDefinitionHandler(fhirpath *FHIRPathEngine) *PlanDefinitionHandler {
	engine := NewPlanDefinitionEngine(fhirpath)
	engine.RegisterBuiltins()
	return &PlanDefinitionHandler{engine: engine}
}

// RegisterRoutes registers PlanDefinition and ActivityDefinition routes.
func (h *PlanDefinitionHandler) RegisterRoutes(fhirGroup *echo.Group) {
	fhirGroup.GET("/PlanDefinition", h.ListPlanDefinitions)
	fhirGroup.GET("/PlanDefinition/:id", h.GetPlanDefinition)
	fhirGroup.POST("/PlanDefinition", h.CreatePlanDefinition)
	fhirGroup.PUT("/PlanDefinition/:id", h.UpdatePlanDefinition)
	fhirGroup.DELETE("/PlanDefinition/:id", h.DeletePlanDefinition)
	fhirGroup.POST("/PlanDefinition/:id/$apply", h.ApplyPlanDefinition)
	fhirGroup.GET("/ActivityDefinition", h.ListActivityDefinitions)
	fhirGroup.GET("/ActivityDefinition/:id", h.GetActivityDefinition)
	fhirGroup.POST("/ActivityDefinition", h.CreateActivityDefinition)
}

func (h *PlanDefinitionHandler) ListPlanDefinitions(c echo.Context) error {
	plans := h.engine.ListPlanDefinitions()
	resources := make([]interface{}, 0, len(plans))
	for _, pd := range plans {
		resources = append(resources, pd.toFHIRMap())
	}
	return c.JSON(http.StatusOK, NewSearchBundle(resources, len(resources), "/fhir/PlanDefinition"))
}

func (h *PlanDefinitionHandler) GetPlanDefinition(c echo.Context) error {
	id := c.Param("id")
	pd := h.engine.GetPlanDefinition(id)
	if pd == nil {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("PlanDefinition", id))
	}
	return c.JSON(http.StatusOK, pd.toFHIRMap())
}

func (h *PlanDefinitionHandler) CreatePlanDefinition(c echo.Context) error {
	var pd PlanDefinition
	if err := json.NewDecoder(c.Request().Body).Decode(&pd); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(fmt.Sprintf("invalid body: %v", err)))
	}
	if pd.ID == "" {
		pd.ID = uuid.New().String()
	}
	h.engine.RegisterPlanDefinition(&pd)
	c.Response().Header().Set("Location", "/fhir/PlanDefinition/"+pd.ID)
	return c.JSON(http.StatusCreated, pd.toFHIRMap())
}

func (h *PlanDefinitionHandler) UpdatePlanDefinition(c echo.Context) error {
	id := c.Param("id")
	if h.engine.GetPlanDefinition(id) == nil {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("PlanDefinition", id))
	}
	var pd PlanDefinition
	if err := json.NewDecoder(c.Request().Body).Decode(&pd); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(fmt.Sprintf("invalid body: %v", err)))
	}
	pd.ID = id
	h.engine.RegisterPlanDefinition(&pd)
	return c.JSON(http.StatusOK, pd.toFHIRMap())
}

func (h *PlanDefinitionHandler) DeletePlanDefinition(c echo.Context) error {
	id := c.Param("id")
	if !h.engine.DeletePlanDefinition(id) {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("PlanDefinition", id))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *PlanDefinitionHandler) ApplyPlanDefinition(c echo.Context) error {
	id := c.Param("id")
	pd := h.engine.GetPlanDefinition(id)
	if pd == nil {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("PlanDefinition", id))
	}

	var body map[string]interface{}
	if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(fmt.Sprintf("invalid body: %v", err)))
	}

	var subject map[string]interface{}
	var params map[string]interface{}

	if paramList, ok := body["parameter"].([]interface{}); ok {
		for _, p := range paramList {
			param, ok := p.(map[string]interface{})
			if !ok {
				continue
			}
			name, _ := param["name"].(string)
			switch name {
			case "subject":
				if res, ok := param["resource"].(map[string]interface{}); ok {
					subject = res
				}
			case "parameters":
				if res, ok := param["resource"].(map[string]interface{}); ok {
					params = res
				}
			}
		}
	}

	if subject == nil {
		if rt, ok := body["resourceType"].(string); ok && rt == "Patient" {
			subject = body
		}
	}

	if subject == nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("subject parameter is required"))
	}

	result, err := h.engine.Apply(c.Request().Context(), pd, subject, params)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
	}

	entries := make([]interface{}, 0)
	if result.CarePlan != nil {
		entries = append(entries, result.CarePlan)
	}
	if result.RequestGroup != nil {
		entries = append(entries, result.RequestGroup)
	}
	for _, r := range result.Resources {
		entries = append(entries, r)
	}
	return c.JSON(http.StatusOK, NewSearchBundle(entries, len(entries), "/fhir/PlanDefinition/"+id+"/$apply"))
}

func (h *PlanDefinitionHandler) ListActivityDefinitions(c echo.Context) error {
	defs := h.engine.ListActivityDefinitions()
	resources := make([]interface{}, 0, len(defs))
	for _, ad := range defs {
		resources = append(resources, ad.toFHIRMap())
	}
	return c.JSON(http.StatusOK, NewSearchBundle(resources, len(resources), "/fhir/ActivityDefinition"))
}

func (h *PlanDefinitionHandler) GetActivityDefinition(c echo.Context) error {
	id := c.Param("id")
	ad := h.engine.GetActivityDefinition(id)
	if ad == nil {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("ActivityDefinition", id))
	}
	return c.JSON(http.StatusOK, ad.toFHIRMap())
}

func (h *PlanDefinitionHandler) CreateActivityDefinition(c echo.Context) error {
	var ad ActivityDefinition
	if err := json.NewDecoder(c.Request().Body).Decode(&ad); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(fmt.Sprintf("invalid body: %v", err)))
	}
	if ad.ID == "" {
		ad.ID = uuid.New().String()
	}
	h.engine.RegisterActivityDefinition(&ad)
	c.Response().Header().Set("Location", "/fhir/ActivityDefinition/"+ad.ID)
	return c.JSON(http.StatusCreated, ad.toFHIRMap())
}

// ============================================================================
// FHIR Serialization helpers
// ============================================================================

func (pd *PlanDefinition) toFHIRMap() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "PlanDefinition",
		"id":           pd.ID,
		"status":       pd.Status,
	}
	if pd.URL != "" {
		result["url"] = pd.URL
	}
	if pd.Version != "" {
		result["version"] = pd.Version
	}
	if pd.Name != "" {
		result["name"] = pd.Name
	}
	if pd.Title != "" {
		result["title"] = pd.Title
	}
	if pd.Type != "" {
		result["type"] = pd.Type
	}
	return result
}

func (ad *ActivityDefinition) toFHIRMap() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ActivityDefinition",
		"id":           ad.ID,
		"status":       ad.Status,
		"kind":         ad.Kind,
	}
	if ad.URL != "" {
		result["url"] = ad.URL
	}
	if ad.Name != "" {
		result["name"] = ad.Name
	}
	if ad.Title != "" {
		result["title"] = ad.Title
	}
	return result
}
