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
// $populate Operation Types
// ============================================================================

// PopulateRequest contains the parameters for $populate.
type PopulateRequest struct {
	QuestionnaireID string                 // ID of questionnaire to populate
	Questionnaire   map[string]interface{} // Or inline questionnaire
	Subject         string                 // Patient reference
	Context         string                 // Encounter context (optional)
	LaunchContext   map[string]interface{} // SMART launch context (optional)
	Local           bool                   // Only use local data (no external calls)
}

// QuestionnaireItem represents a parsed questionnaire item.
type QuestionnaireItem struct {
	LinkID         string
	Text           string
	Type           string // group, display, boolean, decimal, integer, date, dateTime, time, string, text, url, choice, open-choice, attachment, reference, quantity
	Required       bool
	Repeats        bool
	ReadOnly       bool
	MaxLength      int
	AnswerValueSet string // URI of value set for answers
	AnswerOption   []AnswerOption
	Initial        []InitialValue
	Item           []QuestionnaireItem // Nested items
	EnableWhen     []EnableWhenCondition
	EnableBehavior string // all | any
	Definition     string // ElementDefinition URI
	Code           []QuestionnaireCode
	Extension      []QuestionnaireExtension
}

// AnswerOption is a predefined answer choice.
type AnswerOption struct {
	ValueCoding    map[string]interface{}
	ValueString    string
	ValueInteger   *int
	ValueDate      string
	ValueReference map[string]interface{}
}

// InitialValue is a default value for an item.
type InitialValue struct {
	ValueString    string
	ValueBoolean   *bool
	ValueDecimal   *float64
	ValueInteger   *int
	ValueDate      string
	ValueDateTime  string
	ValueCoding    map[string]interface{}
	ValueQuantity  map[string]interface{}
	ValueReference map[string]interface{}
}

// EnableWhenCondition describes when an item should be enabled.
type EnableWhenCondition struct {
	Question string
	Operator string // exists | = | != | > | < | >= | <=
	Answer   interface{}
}

// QuestionnaireCode is a code associated with a questionnaire item.
type QuestionnaireCode struct {
	System  string
	Code    string
	Display string
}

// QuestionnaireExtension represents a FHIR extension on a questionnaire item.
type QuestionnaireExtension struct {
	URL   string
	Value interface{}
}

// PopulateContext provides data context for population.
type PopulateContext struct {
	Patient      map[string]interface{}
	Encounter    map[string]interface{}
	Practitioner map[string]interface{}
	Observations []map[string]interface{}
	Conditions   []map[string]interface{}
	Medications  []map[string]interface{}
	AllResources map[string][]map[string]interface{} // type -> resources
}

// PopulationSource defines where to get data for population.
type PopulationSource struct {
	ResourceType string
	FHIRPath     string
	ValueField   string
}

// PopulateResult contains the output of $populate.
type PopulateResult struct {
	QuestionnaireResponse map[string]interface{}
	Warnings              []string
	PopulatedCount        int
	TotalItems            int
}

// PopulateDataResolver interface for resolving patient data.
type PopulateDataResolver interface {
	ResolvePatient(ctx interface{}, patientRef string) (map[string]interface{}, error)
	ResolveResources(ctx interface{}, patientRef, resourceType string) ([]map[string]interface{}, error)
	ResolveQuestionnaire(ctx interface{}, questionnaireID string) (map[string]interface{}, error)
}

// ParsedQuestionnaire holds parsed questionnaire data.
type ParsedQuestionnaire struct {
	ID          string
	URL         string
	Title       string
	Status      string
	Items       []QuestionnaireItem
	SubjectType []string
}

// ============================================================================
// Parsing Functions
// ============================================================================

// ParseQuestionnaire parses a FHIR Questionnaire into structured items.
func ParseQuestionnaire(data map[string]interface{}) (*ParsedQuestionnaire, error) {
	if data == nil {
		return nil, fmt.Errorf("Questionnaire data is nil")
	}

	rt, _ := data["resourceType"].(string)
	if rt != "" && rt != "Questionnaire" {
		return nil, fmt.Errorf("expected resourceType Questionnaire, got %s", rt)
	}

	parsed := &ParsedQuestionnaire{}
	parsed.ID, _ = data["id"].(string)
	parsed.URL, _ = data["url"].(string)
	parsed.Title, _ = data["title"].(string)
	parsed.Status, _ = data["status"].(string)

	// Parse subjectType
	if stRaw, ok := data["subjectType"].([]interface{}); ok {
		for _, st := range stRaw {
			if s, ok := st.(string); ok {
				parsed.SubjectType = append(parsed.SubjectType, s)
			}
		}
	}

	// Parse items
	if itemsRaw, ok := data["item"].([]interface{}); ok {
		parsed.Items = parseQuestionnaireItems(itemsRaw)
	}

	return parsed, nil
}

// parseQuestionnaireItems parses an array of questionnaire items.
func parseQuestionnaireItems(itemsRaw []interface{}) []QuestionnaireItem {
	var items []QuestionnaireItem
	for _, raw := range itemsRaw {
		itemMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		item := parseQuestionnaireItem(itemMap)
		items = append(items, item)
	}
	return items
}

// parseQuestionnaireItem parses a single questionnaire item map.
func parseQuestionnaireItem(m map[string]interface{}) QuestionnaireItem {
	item := QuestionnaireItem{}

	item.LinkID, _ = m["linkId"].(string)
	item.Text, _ = m["text"].(string)
	item.Type, _ = m["type"].(string)
	item.Definition, _ = m["definition"].(string)
	item.AnswerValueSet, _ = m["answerValueSet"].(string)
	item.EnableBehavior, _ = m["enableBehavior"].(string)

	if req, ok := m["required"].(bool); ok {
		item.Required = req
	}
	if rep, ok := m["repeats"].(bool); ok {
		item.Repeats = rep
	}
	if ro, ok := m["readOnly"].(bool); ok {
		item.ReadOnly = ro
	}
	if ml, ok := m["maxLength"].(float64); ok {
		item.MaxLength = int(ml)
	}

	// Parse code
	if codeRaw, ok := m["code"].([]interface{}); ok {
		for _, c := range codeRaw {
			codeMap, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			qc := QuestionnaireCode{}
			qc.System, _ = codeMap["system"].(string)
			qc.Code, _ = codeMap["code"].(string)
			qc.Display, _ = codeMap["display"].(string)
			item.Code = append(item.Code, qc)
		}
	}

	// Parse answerOption
	if aoRaw, ok := m["answerOption"].([]interface{}); ok {
		for _, ao := range aoRaw {
			aoMap, ok := ao.(map[string]interface{})
			if !ok {
				continue
			}
			opt := AnswerOption{}
			if vc, ok := aoMap["valueCoding"].(map[string]interface{}); ok {
				opt.ValueCoding = vc
			}
			if vs, ok := aoMap["valueString"].(string); ok {
				opt.ValueString = vs
			}
			if vi, ok := aoMap["valueInteger"].(float64); ok {
				v := int(vi)
				opt.ValueInteger = &v
			}
			if vd, ok := aoMap["valueDate"].(string); ok {
				opt.ValueDate = vd
			}
			if vr, ok := aoMap["valueReference"].(map[string]interface{}); ok {
				opt.ValueReference = vr
			}
			item.AnswerOption = append(item.AnswerOption, opt)
		}
	}

	// Parse initial values
	if initRaw, ok := m["initial"].([]interface{}); ok {
		for _, init := range initRaw {
			initMap, ok := init.(map[string]interface{})
			if !ok {
				continue
			}
			iv := InitialValue{}
			if vs, ok := initMap["valueString"].(string); ok {
				iv.ValueString = vs
			}
			if vb, ok := initMap["valueBoolean"].(bool); ok {
				iv.ValueBoolean = &vb
			}
			if vd, ok := initMap["valueDecimal"].(float64); ok {
				iv.ValueDecimal = &vd
			}
			if vi, ok := initMap["valueInteger"].(float64); ok {
				v := int(vi)
				iv.ValueInteger = &v
			}
			if vd, ok := initMap["valueDate"].(string); ok {
				iv.ValueDate = vd
			}
			if vdt, ok := initMap["valueDateTime"].(string); ok {
				iv.ValueDateTime = vdt
			}
			if vc, ok := initMap["valueCoding"].(map[string]interface{}); ok {
				iv.ValueCoding = vc
			}
			if vq, ok := initMap["valueQuantity"].(map[string]interface{}); ok {
				iv.ValueQuantity = vq
			}
			if vr, ok := initMap["valueReference"].(map[string]interface{}); ok {
				iv.ValueReference = vr
			}
			item.Initial = append(item.Initial, iv)
		}
	}

	// Parse enableWhen
	if ewRaw, ok := m["enableWhen"].([]interface{}); ok {
		for _, ew := range ewRaw {
			ewMap, ok := ew.(map[string]interface{})
			if !ok {
				continue
			}
			cond := EnableWhenCondition{}
			cond.Question, _ = ewMap["question"].(string)
			cond.Operator, _ = ewMap["operator"].(string)
			// Extract the answer from various answer[x] fields
			for _, ansKey := range []string{
				"answerBoolean", "answerDecimal", "answerInteger",
				"answerDate", "answerDateTime", "answerTime",
				"answerString", "answerCoding", "answerQuantity",
				"answerReference",
			} {
				if v, ok := ewMap[ansKey]; ok {
					cond.Answer = v
					break
				}
			}
			item.EnableWhen = append(item.EnableWhen, cond)
		}
	}

	// Parse extensions
	if extRaw, ok := m["extension"].([]interface{}); ok {
		for _, ext := range extRaw {
			extMap, ok := ext.(map[string]interface{})
			if !ok {
				continue
			}
			qe := QuestionnaireExtension{}
			qe.URL, _ = extMap["url"].(string)
			// Extract value from various value[x] fields
			for k, v := range extMap {
				if strings.HasPrefix(k, "value") {
					qe.Value = v
					break
				}
			}
			item.Extension = append(item.Extension, qe)
		}
	}

	// Parse nested items
	if subItems, ok := m["item"].([]interface{}); ok {
		item.Item = parseQuestionnaireItems(subItems)
	}

	return item
}

// ============================================================================
// Validation
// ============================================================================

// ValidatePopulateRequest validates a $populate request.
func ValidatePopulateRequest(req *PopulateRequest) []ValidationIssue {
	var issues []ValidationIssue

	if req == nil {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Diagnostics: "PopulateRequest is nil",
		})
		return issues
	}

	if req.QuestionnaireID == "" && req.Questionnaire == nil {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Diagnostics: "questionnaire ID or inline questionnaire is required",
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
// Population Functions
// ============================================================================

// PopulateQuestionnaire generates a QuestionnaireResponse from a Questionnaire and data context.
func PopulateQuestionnaire(questionnaire *ParsedQuestionnaire, ctx *PopulateContext, req *PopulateRequest) (*PopulateResult, error) {
	if questionnaire == nil {
		return nil, fmt.Errorf("questionnaire is nil")
	}

	if ctx == nil {
		ctx = &PopulateContext{}
	}

	subject := ""
	if req != nil {
		subject = req.Subject
	}

	result := &PopulateResult{
		Warnings: make([]string, 0),
	}

	// Build empty QR shell
	qr := BuildEmptyQuestionnaireResponse(questionnaire, subject)

	// Populate items
	var qrItems []interface{}
	totalItems := 0
	populatedCount := 0

	for i := range questionnaire.Items {
		totalItems += countItems(&questionnaire.Items[i])
		qrItem, populated := PopulateItem(&questionnaire.Items[i], ctx)
		if qrItem != nil {
			qrItems = append(qrItems, qrItem)
		}
		if populated {
			populatedCount += countPopulatedItems(qrItem)
		}
	}

	if len(qrItems) > 0 {
		qr["item"] = qrItems
	}

	result.QuestionnaireResponse = qr
	result.TotalItems = totalItems
	result.PopulatedCount = populatedCount

	return result, nil
}

// countItems counts the total number of items including nested ones.
func countItems(item *QuestionnaireItem) int {
	count := 1
	for i := range item.Item {
		count += countItems(&item.Item[i])
	}
	return count
}

// countPopulatedItems counts populated items in a QR item tree.
func countPopulatedItems(qrItem map[string]interface{}) int {
	if qrItem == nil {
		return 0
	}
	count := 0
	if _, hasAnswer := qrItem["answer"]; hasAnswer {
		count = 1
	}
	if subItems, ok := qrItem["item"].([]interface{}); ok {
		for _, sub := range subItems {
			if subMap, ok := sub.(map[string]interface{}); ok {
				count += countPopulatedItems(subMap)
			}
		}
	}
	return count
}

// PopulateItem populates a single questionnaire item from context.
// Returns the QR item and whether it was populated with data.
func PopulateItem(item *QuestionnaireItem, ctx *PopulateContext) (map[string]interface{}, bool) {
	if item == nil {
		return nil, false
	}

	// Display items are never populated with answers
	if item.Type == "display" {
		return nil, false
	}

	// Handle group items by recursing into children
	if item.Type == "group" {
		return populateGroupItem(item, ctx)
	}

	// Try to extract a value from the context
	value, found := ExtractPopulationValue(item, ctx)

	// If no value found from context, check initial values
	if !found && len(item.Initial) > 0 {
		value = extractInitialValue(item.Initial[0])
		if value != nil {
			found = true
		}
	}

	if found && value != nil {
		return BuildQuestionnaireResponseItem(item, value), true
	}

	return nil, false
}

// populateGroupItem handles population of group items with nested children.
func populateGroupItem(item *QuestionnaireItem, ctx *PopulateContext) (map[string]interface{}, bool) {
	groupResult := map[string]interface{}{
		"linkId": item.LinkID,
	}
	if item.Text != "" {
		groupResult["text"] = item.Text
	}

	var subItems []interface{}
	anyPopulated := false

	for i := range item.Item {
		subResult, populated := PopulateItem(&item.Item[i], ctx)
		if subResult != nil {
			subItems = append(subItems, subResult)
			if populated {
				anyPopulated = true
			}
		}
	}

	if len(subItems) > 0 {
		groupResult["item"] = subItems
	}

	return groupResult, anyPopulated
}

// extractInitialValue extracts a value from an InitialValue.
func extractInitialValue(iv InitialValue) interface{} {
	if iv.ValueString != "" {
		return iv.ValueString
	}
	if iv.ValueBoolean != nil {
		return *iv.ValueBoolean
	}
	if iv.ValueDecimal != nil {
		return *iv.ValueDecimal
	}
	if iv.ValueInteger != nil {
		return *iv.ValueInteger
	}
	if iv.ValueDate != "" {
		return iv.ValueDate
	}
	if iv.ValueDateTime != "" {
		return iv.ValueDateTime
	}
	if iv.ValueCoding != nil {
		return iv.ValueCoding
	}
	if iv.ValueQuantity != nil {
		return iv.ValueQuantity
	}
	if iv.ValueReference != nil {
		return iv.ValueReference
	}
	return nil
}

// BuildQuestionnaireResponseItem builds a QR item with answer from context.
func BuildQuestionnaireResponseItem(item *QuestionnaireItem, answer interface{}) map[string]interface{} {
	qrItem := map[string]interface{}{
		"linkId": item.LinkID,
	}
	if item.Text != "" {
		qrItem["text"] = item.Text
	}

	if answer != nil {
		answerEntry := buildAnswerEntry(item.Type, answer)
		qrItem["answer"] = []interface{}{answerEntry}
	}

	return qrItem
}

// buildAnswerEntry builds an answer entry with the appropriate value[x] field.
func buildAnswerEntry(itemType string, value interface{}) map[string]interface{} {
	answer := make(map[string]interface{})

	switch itemType {
	case "boolean":
		answer["valueBoolean"] = value
	case "decimal":
		answer["valueDecimal"] = value
	case "integer":
		answer["valueInteger"] = value
	case "date":
		answer["valueDate"] = value
	case "dateTime":
		answer["valueDateTime"] = value
	case "time":
		answer["valueTime"] = value
	case "string", "text":
		answer["valueString"] = value
	case "url":
		answer["valueUri"] = value
	case "choice", "open-choice":
		if m, ok := value.(map[string]interface{}); ok {
			answer["valueCoding"] = m
		} else {
			answer["valueString"] = value
		}
	case "reference":
		if m, ok := value.(map[string]interface{}); ok {
			answer["valueReference"] = m
		}
	case "quantity":
		if m, ok := value.(map[string]interface{}); ok {
			answer["valueQuantity"] = m
		}
	case "attachment":
		if m, ok := value.(map[string]interface{}); ok {
			answer["valueAttachment"] = m
		}
	default:
		answer["valueString"] = fmt.Sprintf("%v", value)
	}

	return answer
}

// BuildEmptyQuestionnaireResponse creates an empty QuestionnaireResponse shell.
func BuildEmptyQuestionnaireResponse(questionnaire *ParsedQuestionnaire, subject string) map[string]interface{} {
	qr := map[string]interface{}{
		"resourceType": "QuestionnaireResponse",
		"id":           uuid.New().String(),
		"status":       "in-progress",
		"authored":     time.Now().UTC().Format(time.RFC3339),
	}

	// Set the questionnaire reference
	if questionnaire.URL != "" {
		qr["questionnaire"] = questionnaire.URL
	} else if questionnaire.ID != "" {
		qr["questionnaire"] = "Questionnaire/" + questionnaire.ID
	}

	// Set subject
	if subject != "" {
		qr["subject"] = map[string]interface{}{
			"reference": subject,
		}
	}

	return qr
}

// ============================================================================
// Enable-When Evaluation
// ============================================================================

// EvaluateEnableWhen checks if an item's enable-when conditions are met.
func EvaluateEnableWhen(conditions []EnableWhenCondition, behavior string, answers map[string]interface{}) bool {
	if len(conditions) == 0 {
		return true
	}

	if behavior == "" {
		behavior = "all"
	}

	for _, cond := range conditions {
		result := evaluateSingleCondition(cond, answers)
		if behavior == "any" && result {
			return true
		}
		if behavior == "all" && !result {
			return false
		}
	}

	if behavior == "any" {
		return false
	}
	return true
}

// evaluateSingleCondition evaluates a single enable-when condition.
func evaluateSingleCondition(cond EnableWhenCondition, answers map[string]interface{}) bool {
	actualValue, exists := answers[cond.Question]

	switch cond.Operator {
	case "exists":
		expected, ok := cond.Answer.(bool)
		if !ok {
			return false
		}
		return exists == expected

	case "=":
		if !exists {
			return false
		}
		return fmt.Sprintf("%v", actualValue) == fmt.Sprintf("%v", cond.Answer)

	case "!=":
		if !exists {
			return true
		}
		return fmt.Sprintf("%v", actualValue) != fmt.Sprintf("%v", cond.Answer)

	case ">":
		return compareNumeric(actualValue, cond.Answer) > 0

	case "<":
		return compareNumeric(actualValue, cond.Answer) < 0

	case ">=":
		return compareNumeric(actualValue, cond.Answer) >= 0

	case "<=":
		return compareNumeric(actualValue, cond.Answer) <= 0

	default:
		return false
	}
}

// compareNumeric compares two values numerically. Returns -1, 0, or 1.
func compareNumeric(a, b interface{}) int {
	af := toFloat64(a)
	bf := toFloat64(b)
	if af < bf {
		return -1
	}
	if af > bf {
		return 1
	}
	return 0
}

// toFloat64 converts a value to float64.
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	default:
		return 0
	}
}

// ============================================================================
// Value Extraction
// ============================================================================

// ExtractPopulationValue extracts a value from context resources based on item definition/code.
func ExtractPopulationValue(item *QuestionnaireItem, ctx *PopulateContext) (interface{}, bool) {
	if ctx == nil {
		return nil, false
	}

	// Try to extract from patient demographics via definition
	if item.Definition != "" && ctx.Patient != nil {
		val, found := extractFromPatientByDefinition(item.Definition, ctx.Patient)
		if found {
			return val, true
		}
	}

	// Try to extract from observations by code
	if len(item.Code) > 0 && len(ctx.Observations) > 0 {
		for _, code := range item.Code {
			matches := MatchResourceByCode(code, ctx.Observations)
			if len(matches) > 0 {
				return extractValueFromObservation(matches[0]), true
			}
		}
	}

	// Try to extract from conditions by code
	if len(item.Code) > 0 && len(ctx.Conditions) > 0 {
		for _, code := range item.Code {
			matches := MatchResourceByCode(code, ctx.Conditions)
			if len(matches) > 0 {
				return extractCodeFromResource(matches[0]), true
			}
		}
	}

	// Try to extract from medications by code
	if len(item.Code) > 0 && len(ctx.Medications) > 0 {
		for _, code := range item.Code {
			matches := matchMedicationByCode(code, ctx.Medications)
			if len(matches) > 0 {
				return extractMedicationCode(matches[0]), true
			}
		}
	}

	return nil, false
}

// extractFromPatientByDefinition extracts a patient field value based on a FHIR definition URI.
func extractFromPatientByDefinition(definition string, patient map[string]interface{}) (interface{}, bool) {
	// Definition format: http://hl7.org/fhir/StructureDefinition/Patient#Patient.field.subfield
	parts := strings.Split(definition, "#")
	if len(parts) != 2 {
		return nil, false
	}

	path := parts[1] // e.g., "Patient.name.family"
	segments := strings.Split(path, ".")
	if len(segments) < 2 || segments[0] != "Patient" {
		return nil, false
	}

	field := segments[1]

	switch field {
	case "name":
		return extractPatientNameField(patient, segments)
	case "gender":
		if val, ok := patient["gender"].(string); ok {
			return val, true
		}
	case "birthDate":
		if val, ok := patient["birthDate"].(string); ok {
			return val, true
		}
	case "telecom":
		return extractPatientTelecomField(patient, segments)
	case "address":
		return extractPatientAddressField(patient, segments)
	}

	return nil, false
}

// extractPatientNameField extracts name subfields from a patient resource.
func extractPatientNameField(patient map[string]interface{}, segments []string) (interface{}, bool) {
	names, ok := patient["name"].([]interface{})
	if !ok || len(names) == 0 {
		return nil, false
	}
	nameMap, ok := names[0].(map[string]interface{})
	if !ok {
		return nil, false
	}

	if len(segments) < 3 {
		return nameMap, true
	}

	subField := segments[2]
	switch subField {
	case "family":
		if val, ok := nameMap["family"].(string); ok {
			return val, true
		}
	case "given":
		if given, ok := nameMap["given"].([]interface{}); ok && len(given) > 0 {
			return given[0], true
		}
	}
	return nil, false
}

// extractPatientTelecomField extracts telecom subfields from a patient resource.
func extractPatientTelecomField(patient map[string]interface{}, segments []string) (interface{}, bool) {
	telecoms, ok := patient["telecom"].([]interface{})
	if !ok || len(telecoms) == 0 {
		return nil, false
	}
	telecomMap, ok := telecoms[0].(map[string]interface{})
	if !ok {
		return nil, false
	}

	if len(segments) < 3 {
		return telecomMap, true
	}

	subField := segments[2]
	switch subField {
	case "value":
		if val, ok := telecomMap["value"].(string); ok {
			return val, true
		}
	case "system":
		if val, ok := telecomMap["system"].(string); ok {
			return val, true
		}
	}
	return nil, false
}

// extractPatientAddressField extracts address subfields from a patient resource.
func extractPatientAddressField(patient map[string]interface{}, segments []string) (interface{}, bool) {
	addresses, ok := patient["address"].([]interface{})
	if !ok || len(addresses) == 0 {
		return nil, false
	}
	addrMap, ok := addresses[0].(map[string]interface{})
	if !ok {
		return nil, false
	}

	if len(segments) < 3 {
		return addrMap, true
	}

	subField := segments[2]
	switch subField {
	case "line":
		if lines, ok := addrMap["line"].([]interface{}); ok && len(lines) > 0 {
			return lines[0], true
		}
	case "city":
		if val, ok := addrMap["city"].(string); ok {
			return val, true
		}
	case "state":
		if val, ok := addrMap["state"].(string); ok {
			return val, true
		}
	case "postalCode":
		if val, ok := addrMap["postalCode"].(string); ok {
			return val, true
		}
	case "country":
		if val, ok := addrMap["country"].(string); ok {
			return val, true
		}
	}
	return nil, false
}

// extractValueFromObservation extracts the value from an Observation resource.
func extractValueFromObservation(obs map[string]interface{}) interface{} {
	// Check value[x] types
	if vq, ok := obs["valueQuantity"].(map[string]interface{}); ok {
		return vq
	}
	if vc, ok := obs["valueCodeableConcept"].(map[string]interface{}); ok {
		return vc
	}
	if vs, ok := obs["valueString"].(string); ok {
		return vs
	}
	if vb, ok := obs["valueBoolean"].(bool); ok {
		return vb
	}
	if vi, ok := obs["valueInteger"].(float64); ok {
		return vi
	}
	if vdt, ok := obs["valueDateTime"].(string); ok {
		return vdt
	}
	return nil
}

// extractCodeFromResource extracts the code coding from a resource (Condition, etc.).
func extractCodeFromResource(resource map[string]interface{}) interface{} {
	codeObj, ok := resource["code"].(map[string]interface{})
	if !ok {
		return nil
	}
	codings, ok := codeObj["coding"].([]interface{})
	if !ok || len(codings) == 0 {
		return nil
	}
	coding, ok := codings[0].(map[string]interface{})
	if !ok {
		return nil
	}
	return coding
}

// extractMedicationCode extracts the medication code from a MedicationRequest.
func extractMedicationCode(med map[string]interface{}) interface{} {
	if mcc, ok := med["medicationCodeableConcept"].(map[string]interface{}); ok {
		codings, ok := mcc["coding"].([]interface{})
		if ok && len(codings) > 0 {
			if coding, ok := codings[0].(map[string]interface{}); ok {
				return coding
			}
		}
	}
	return nil
}

// ============================================================================
// Code Matching
// ============================================================================

// MatchResourceByCode finds resources matching a questionnaire item's code.
func MatchResourceByCode(code QuestionnaireCode, resources []map[string]interface{}) []map[string]interface{} {
	var matches []map[string]interface{}
	for _, res := range resources {
		if populateResourceMatchesCode(code, res) {
			matches = append(matches, res)
		}
	}
	return matches
}

// populateResourceMatchesCode checks if a resource's code matches the given QuestionnaireCode.
func populateResourceMatchesCode(code QuestionnaireCode, resource map[string]interface{}) bool {
	codeObj, ok := resource["code"].(map[string]interface{})
	if !ok {
		return false
	}
	codings, ok := codeObj["coding"].([]interface{})
	if !ok {
		return false
	}
	for _, c := range codings {
		coding, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		sys, _ := coding["system"].(string)
		cd, _ := coding["code"].(string)
		if sys == code.System && cd == code.Code {
			return true
		}
	}
	return false
}

// matchMedicationByCode matches MedicationRequest resources by medicationCodeableConcept.
func matchMedicationByCode(code QuestionnaireCode, meds []map[string]interface{}) []map[string]interface{} {
	var matches []map[string]interface{}
	for _, med := range meds {
		mcc, ok := med["medicationCodeableConcept"].(map[string]interface{})
		if !ok {
			continue
		}
		codings, ok := mcc["coding"].([]interface{})
		if !ok {
			continue
		}
		for _, c := range codings {
			coding, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			sys, _ := coding["system"].(string)
			cd, _ := coding["code"].(string)
			if sys == code.System && cd == code.Code {
				matches = append(matches, med)
				break
			}
		}
	}
	return matches
}

// ============================================================================
// Value Conversion
// ============================================================================

// ConvertToAnswerValue converts a FHIR value to appropriate QR answer format.
func ConvertToAnswerValue(value interface{}, itemType string) interface{} {
	switch itemType {
	case "integer":
		if f, ok := value.(float64); ok {
			return int(f)
		}
		return value
	case "decimal":
		return value
	case "boolean":
		return value
	case "date", "dateTime", "time", "string", "text", "url":
		return value
	case "choice", "open-choice":
		return value
	case "quantity":
		return value
	case "reference":
		return value
	default:
		return value
	}
}

// ============================================================================
// Context Building
// ============================================================================

// BuildPopulateContext creates a PopulateContext from resolved data.
func BuildPopulateContext(patient map[string]interface{}, resources map[string][]map[string]interface{}) *PopulateContext {
	ctx := &PopulateContext{
		Patient:      patient,
		AllResources: resources,
	}

	if resources == nil {
		ctx.AllResources = make(map[string][]map[string]interface{})
		return ctx
	}

	if obs, ok := resources["Observation"]; ok {
		ctx.Observations = obs
	}
	if conds, ok := resources["Condition"]; ok {
		ctx.Conditions = conds
	}
	if meds, ok := resources["MedicationRequest"]; ok {
		ctx.Medications = meds
	}

	return ctx
}

// ============================================================================
// In-Memory Resolver (for testing)
// ============================================================================

// InMemoryPopulateResolver is a test implementation of PopulateDataResolver.
type InMemoryPopulateResolver struct {
	patients       map[string]map[string]interface{}
	resources      map[string]map[string][]map[string]interface{} // patient -> type -> resources
	questionnaires map[string]map[string]interface{}
}

// NewInMemoryPopulateResolver creates a new InMemoryPopulateResolver.
func NewInMemoryPopulateResolver() *InMemoryPopulateResolver {
	return &InMemoryPopulateResolver{
		patients:       make(map[string]map[string]interface{}),
		resources:      make(map[string]map[string][]map[string]interface{}),
		questionnaires: make(map[string]map[string]interface{}),
	}
}

// AddPatient adds a patient to the in-memory store.
func (r *InMemoryPopulateResolver) AddPatient(id string, patient map[string]interface{}) {
	r.patients[id] = patient
}

// AddResource adds a resource to the in-memory store for a given patient.
func (r *InMemoryPopulateResolver) AddResource(patientID, resourceType string, resource map[string]interface{}) {
	if r.resources[patientID] == nil {
		r.resources[patientID] = make(map[string][]map[string]interface{})
	}
	r.resources[patientID][resourceType] = append(r.resources[patientID][resourceType], resource)
}

// AddQuestionnaire adds a questionnaire to the in-memory store.
func (r *InMemoryPopulateResolver) AddQuestionnaire(id string, questionnaire map[string]interface{}) {
	r.questionnaires[id] = questionnaire
}

// ResolvePatient resolves a patient by reference.
func (r *InMemoryPopulateResolver) ResolvePatient(_ interface{}, patientRef string) (map[string]interface{}, error) {
	// Extract ID from reference
	id := patientRef
	if strings.Contains(patientRef, "/") {
		parts := strings.SplitN(patientRef, "/", 2)
		id = parts[1]
	}

	patient, ok := r.patients[id]
	if !ok {
		return nil, fmt.Errorf("patient %s not found", patientRef)
	}
	return patient, nil
}

// ResolveResources resolves resources for a patient.
func (r *InMemoryPopulateResolver) ResolveResources(_ interface{}, patientRef, resourceType string) ([]map[string]interface{}, error) {
	// Extract ID from reference
	id := patientRef
	if strings.Contains(patientRef, "/") {
		parts := strings.SplitN(patientRef, "/", 2)
		id = parts[1]
	}

	if typeMap, ok := r.resources[id]; ok {
		if resources, ok := typeMap[resourceType]; ok {
			return resources, nil
		}
	}
	return []map[string]interface{}{}, nil
}

// ResolveQuestionnaire resolves a questionnaire by ID.
func (r *InMemoryPopulateResolver) ResolveQuestionnaire(_ interface{}, questionnaireID string) (map[string]interface{}, error) {
	q, ok := r.questionnaires[questionnaireID]
	if !ok {
		return nil, fmt.Errorf("questionnaire %s not found", questionnaireID)
	}
	return q, nil
}

// ============================================================================
// HTTP Handler
// ============================================================================

// PopulateHandler returns an echo.HandlerFunc for POST /fhir/Questionnaire/{id}/$populate.
func PopulateHandler(resolver PopulateDataResolver) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, ErrorOutcome("Questionnaire ID is required"))
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

		// Parse the populate request from the body
		popReq := parsePopulateRequestFromBody(bodyMap)
		popReq.QuestionnaireID = id

		// Validate subject
		if popReq.Subject == "" {
			return c.JSON(http.StatusBadRequest, ErrorOutcome("subject parameter is required"))
		}

		// Resolve the Questionnaire
		qData, err := resolver.ResolveQuestionnaire(c.Request().Context(), id)
		if err != nil || qData == nil {
			return c.JSON(http.StatusNotFound, NotFoundOutcome("Questionnaire", id))
		}

		// Parse the questionnaire
		questionnaire, err := ParseQuestionnaire(qData)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
		}

		// Resolve patient data
		patient, err := resolver.ResolvePatient(c.Request().Context(), popReq.Subject)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorOutcome("failed to resolve patient: "+err.Error()))
		}

		// Resolve resources for context
		allResources := make(map[string][]map[string]interface{})
		for _, resType := range []string{"Observation", "Condition", "MedicationRequest", "AllergyIntolerance", "Procedure"} {
			resources, err := resolver.ResolveResources(c.Request().Context(), popReq.Subject, resType)
			if err == nil && len(resources) > 0 {
				allResources[resType] = resources
			}
		}

		// Build populate context
		ctx := BuildPopulateContext(patient, allResources)

		// Populate the questionnaire
		result, err := PopulateQuestionnaire(questionnaire, ctx, popReq)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorOutcome(err.Error()))
		}

		return c.JSON(http.StatusOK, result.QuestionnaireResponse)
	}
}

// parsePopulateRequestFromBody extracts PopulateRequest fields from a request body.
// It supports both simple JSON format and FHIR Parameters resource format.
func parsePopulateRequestFromBody(body map[string]interface{}) *PopulateRequest {
	req := &PopulateRequest{}

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
				case "context":
					if vs, ok := param["valueString"].(string); ok {
						req.Context = vs
					} else if ref, ok := param["valueReference"].(map[string]interface{}); ok {
						req.Context, _ = ref["reference"].(string)
					}
				case "local":
					if vb, ok := param["valueBoolean"].(bool); ok {
						req.Local = vb
					}
				case "questionnaire":
					if vr, ok := param["resource"].(map[string]interface{}); ok {
						req.Questionnaire = vr
					}
				case "launchContext":
					if vr, ok := param["resource"].(map[string]interface{}); ok {
						req.LaunchContext = vr
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
	if context, ok := body["context"].(string); ok {
		req.Context = context
	}
	if local, ok := body["local"].(bool); ok {
		req.Local = local
	}
	if q, ok := body["questionnaire"].(map[string]interface{}); ok {
		req.Questionnaire = q
	}
	if lc, ok := body["launchContext"].(map[string]interface{}); ok {
		req.LaunchContext = lc
	}

	return req
}
