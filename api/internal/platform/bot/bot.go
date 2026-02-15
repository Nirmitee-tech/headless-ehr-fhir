// Package bot provides a server-side scripting (Bots) system for the EHR platform.
// Bots are server-side scripts that execute in response to FHIR resource events
// (create, update, delete), cron schedules, webhooks, or manual triggers. The bot
// engine uses a FHIRPath-based DSL for script actions, providing a safe, sandboxed
// execution environment without external runtime dependencies.
package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Domain types
// ---------------------------------------------------------------------------

// Bot represents a server-side script that executes in response to triggers.
type Bot struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	Status        string            `json:"status"`
	Trigger       BotTrigger        `json:"trigger"`
	Code          string            `json:"code"`
	Runtime       string            `json:"runtime"`
	Config        map[string]string `json:"config,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	LastRunAt     *time.Time        `json:"last_run_at,omitempty"`
	LastRunStatus string            `json:"last_run_status,omitempty"`
	RunCount      int               `json:"run_count"`
}

// BotTrigger defines when a bot executes.
type BotTrigger struct {
	Type         string `json:"type"`
	ResourceType string `json:"resource_type,omitempty"`
	Event        string `json:"event,omitempty"`
	CronSchedule string `json:"cron_schedule,omitempty"`
	Criteria     string `json:"criteria,omitempty"`
}

// BotAction represents a single action in a bot script.
type BotAction struct {
	Type       string                 `json:"type"`
	Target     string                 `json:"target,omitempty"`
	Expression string                 `json:"expression,omitempty"`
	Value      interface{}            `json:"value,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
	OnTrue     []BotAction            `json:"on_true,omitempty"`
	OnFalse    []BotAction            `json:"on_false,omitempty"`
}

// BotInput is the input to a bot execution.
type BotInput struct {
	Resource     map[string]interface{} `json:"resource"`
	ResourceType string                 `json:"resource_type"`
	Event        string                 `json:"event"`
	Params       map[string]interface{} `json:"params,omitempty"`
}

// BotOutput is the result of a bot execution.
type BotOutput struct {
	BotID           string                   `json:"bot_id"`
	BotName         string                   `json:"bot_name"`
	Status          string                   `json:"status"`
	Duration        time.Duration            `json:"duration_ms"`
	Logs            []string                 `json:"logs,omitempty"`
	OutputResources []map[string]interface{} `json:"output_resources,omitempty"`
	Error           string                   `json:"error,omitempty"`
	ActionsExecuted int                      `json:"actions_executed"`
}

// BotExecutionLog records a bot execution for audit.
type BotExecutionLog struct {
	ID        string    `json:"id"`
	BotID     string    `json:"bot_id"`
	BotName   string    `json:"bot_name"`
	Input     BotInput  `json:"input"`
	Output    BotOutput `json:"output"`
	Timestamp time.Time `json:"timestamp"`
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

var validBotStatuses = map[string]bool{
	"active":   true,
	"inactive": true,
	"error":    true,
}

var validTriggerTypes = map[string]bool{
	"subscription": true,
	"cron":         true,
	"manual":       true,
	"webhook":      true,
}

func validateBot(b Bot) error {
	if b.ID == "" {
		return fmt.Errorf("bot id is required")
	}
	if b.Name == "" {
		return fmt.Errorf("bot name is required")
	}
	if b.Trigger.Type == "" {
		return fmt.Errorf("bot trigger type is required")
	}
	if !validTriggerTypes[b.Trigger.Type] {
		return fmt.Errorf("invalid trigger type: %s (supported: subscription, cron, manual, webhook)", b.Trigger.Type)
	}
	if b.Status != "" && !validBotStatuses[b.Status] {
		return fmt.Errorf("invalid status: %s (supported: active, inactive, error)", b.Status)
	}
	return nil
}

// ---------------------------------------------------------------------------
// BotEngine
// ---------------------------------------------------------------------------

const (
	defaultMaxLogs          = 1000
	defaultMaxActions       = 100
	defaultExecutionTimeout = 30 * time.Second
	defaultWebhookTimeout   = 10 * time.Second
)

// BotEngine executes bots in response to events.
type BotEngine struct {
	bots             map[string]*Bot
	botOrder         []string // preserve insertion order
	execLogs         []BotExecutionLog
	fhirpath         *fhir.FHIRPathEngine
	mu               sync.RWMutex
	maxLogs          int
	maxActions       int
	executionTimeout time.Duration
	webhookTimeout   time.Duration
}

// NewBotEngine creates a new BotEngine with sensible defaults.
func NewBotEngine() *BotEngine {
	return &BotEngine{
		bots:             make(map[string]*Bot),
		fhirpath:         fhir.NewFHIRPathEngine(),
		maxLogs:          defaultMaxLogs,
		maxActions:       defaultMaxActions,
		executionTimeout: defaultExecutionTimeout,
		webhookTimeout:   defaultWebhookTimeout,
	}
}

// RegisterBot registers or updates a bot.
func (e *BotEngine) RegisterBot(bot Bot) error {
	if err := validateBot(bot); err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	existing, exists := e.bots[bot.ID]
	if exists {
		// Preserve original CreatedAt and run stats
		bot.CreatedAt = existing.CreatedAt
		bot.RunCount = existing.RunCount
		bot.LastRunAt = existing.LastRunAt
		bot.LastRunStatus = existing.LastRunStatus
	} else {
		bot.CreatedAt = now
		e.botOrder = append(e.botOrder, bot.ID)
	}
	bot.UpdatedAt = now
	if bot.Status == "" {
		bot.Status = "active"
	}
	if bot.Runtime == "" {
		bot.Runtime = "fhirpath"
	}

	stored := bot
	e.bots[bot.ID] = &stored
	return nil
}

// GetBot retrieves a bot by ID.
func (e *BotEngine) GetBot(id string) (*Bot, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	bot, ok := e.bots[id]
	if !ok {
		return nil, fmt.Errorf("bot %s not found", id)
	}
	copy := *bot
	return &copy, nil
}

// ListBots returns all bots, optionally filtered by status.
func (e *BotEngine) ListBots(status string) []Bot {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]Bot, 0, len(e.bots))
	for _, id := range e.botOrder {
		bot, ok := e.bots[id]
		if !ok {
			continue
		}
		if status == "" || bot.Status == status {
			result = append(result, *bot)
		}
	}
	return result
}

// DeleteBot removes a bot.
func (e *BotEngine) DeleteBot(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.bots[id]; !ok {
		return fmt.Errorf("bot %s not found", id)
	}
	delete(e.bots, id)
	for i, oid := range e.botOrder {
		if oid == id {
			e.botOrder = append(e.botOrder[:i], e.botOrder[i+1:]...)
			break
		}
	}
	return nil
}

// Execute runs a bot with the given input resource.
func (e *BotEngine) Execute(ctx context.Context, botID string, input BotInput) (*BotOutput, error) {
	bot, err := e.GetBot(botID)
	if err != nil {
		return nil, err
	}

	output := &BotOutput{
		BotID:   bot.ID,
		BotName: bot.Name,
	}

	// Check if bot is active
	if bot.Status != "active" {
		output.Status = "error"
		output.Error = fmt.Sprintf("bot %s is not active (status: %s)", bot.ID, bot.Status)
		e.recordExecution(bot, input, output)
		return output, nil
	}

	// Parse actions from code
	var actions []BotAction
	if err := json.Unmarshal([]byte(bot.Code), &actions); err != nil {
		output.Status = "error"
		output.Error = fmt.Sprintf("failed to parse bot code: %v", err)
		e.recordExecution(bot, input, output)
		return output, nil
	}

	// Check action limit (count total including nested)
	totalActions := countActions(actions)
	if totalActions > e.maxActions {
		output.Status = "error"
		output.Error = fmt.Sprintf("action limit exceeded: %d actions (max %d)", totalActions, e.maxActions)
		e.recordExecution(bot, input, output)
		return output, nil
	}

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(ctx, e.executionTimeout)
	defer cancel()

	start := time.Now()

	// Deep-copy the resource so actions don't mutate the original
	resource := deepCopyMap(input.Resource)

	// Track whether any transform/set-status was applied
	modified := false

	execErr := e.executeActions(execCtx, actions, resource, output, &modified)
	output.Duration = time.Since(start)

	if execErr != nil {
		output.Status = "error"
		output.Error = execErr.Error()
	} else {
		output.Status = "success"
	}

	// If any transform or set-status was applied, add the resource as first output
	if modified {
		// Prepend the modified resource
		output.OutputResources = append([]map[string]interface{}{resource}, output.OutputResources...)
	}

	e.recordExecution(bot, input, output)
	return output, nil
}

// ExecuteByTrigger finds and runs all matching bots for a trigger event.
func (e *BotEngine) ExecuteByTrigger(ctx context.Context, resourceType, event string, resource map[string]interface{}) []BotOutput {
	e.mu.RLock()
	var matching []*Bot
	for _, id := range e.botOrder {
		bot := e.bots[id]
		if bot == nil || bot.Status != "active" {
			continue
		}
		if bot.Trigger.Type != "subscription" {
			continue
		}
		if bot.Trigger.ResourceType != resourceType {
			continue
		}
		if bot.Trigger.Event != "*" && bot.Trigger.Event != event {
			continue
		}
		// Evaluate criteria if present
		if bot.Trigger.Criteria != "" {
			match, err := e.fhirpath.EvaluateBool(resource, bot.Trigger.Criteria)
			if err != nil || !match {
				continue
			}
		}
		copy := *bot
		matching = append(matching, &copy)
	}
	e.mu.RUnlock()

	results := make([]BotOutput, len(matching))
	var wg sync.WaitGroup
	for i, bot := range matching {
		wg.Add(1)
		go func(idx int, b *Bot) {
			defer wg.Done()
			input := BotInput{
				Resource:     resource,
				ResourceType: resourceType,
				Event:        event,
			}
			out, err := e.Execute(ctx, b.ID, input)
			if err != nil {
				results[idx] = BotOutput{
					BotID:   b.ID,
					BotName: b.Name,
					Status:  "error",
					Error:   err.Error(),
				}
			} else {
				results[idx] = *out
			}
		}(i, bot)
	}
	wg.Wait()
	return results
}

// GetExecutionLogs returns execution logs for a specific bot.
func (e *BotEngine) GetExecutionLogs(botID string) []BotExecutionLog {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []BotExecutionLog
	for _, log := range e.execLogs {
		if log.BotID == botID {
			result = append(result, log)
		}
	}
	return result
}

// GetAllExecutionLogs returns all execution logs.
func (e *BotEngine) GetAllExecutionLogs() []BotExecutionLog {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]BotExecutionLog, len(e.execLogs))
	copy(result, e.execLogs)
	return result
}

// recordExecution logs a bot execution and updates bot stats.
func (e *BotEngine) recordExecution(bot *Bot, input BotInput, output *BotOutput) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Update bot stats
	if b, ok := e.bots[bot.ID]; ok {
		now := time.Now()
		b.LastRunAt = &now
		b.LastRunStatus = output.Status
		b.RunCount++
	}

	// Record execution log
	log := BotExecutionLog{
		ID:        uuid.New().String(),
		BotID:     bot.ID,
		BotName:   bot.Name,
		Input:     input,
		Output:    *output,
		Timestamp: time.Now(),
	}

	if len(e.execLogs) >= e.maxLogs {
		// Ring buffer: remove oldest
		e.execLogs = e.execLogs[1:]
	}
	e.execLogs = append(e.execLogs, log)
}

// ---------------------------------------------------------------------------
// Action execution
// ---------------------------------------------------------------------------

// executeActions runs a sequence of bot actions against a resource.
func (e *BotEngine) executeActions(ctx context.Context, actions []BotAction, resource map[string]interface{}, output *BotOutput, modified *bool) error {
	for _, action := range actions {
		select {
		case <-ctx.Done():
			return fmt.Errorf("execution timeout exceeded")
		default:
		}

		if err := e.executeAction(ctx, action, resource, output, modified); err != nil {
			return err
		}
		output.ActionsExecuted++
	}
	return nil
}

func (e *BotEngine) executeAction(ctx context.Context, action BotAction, resource map[string]interface{}, output *BotOutput, modified *bool) error {
	switch action.Type {
	case "log":
		return e.executeLog(action, resource, output)
	case "condition":
		return e.executeCondition(ctx, action, resource, output, modified)
	case "transform":
		return e.executeTransform(action, resource, output, modified)
	case "create":
		return e.executeCreate(action, resource, output)
	case "validate":
		return e.executeValidate(action, resource, output)
	case "webhook":
		return e.executeWebhook(ctx, action, resource, output)
	case "send-notification":
		return e.executeSendNotification(action, resource, output)
	case "set-status":
		return e.executeSetStatus(action, resource, output, modified)
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

func (e *BotEngine) executeLog(action BotAction, resource map[string]interface{}, output *BotOutput) error {
	// If value is set, use it as a literal string
	if action.Value != nil {
		if s, ok := action.Value.(string); ok {
			output.Logs = append(output.Logs, s)
			return nil
		}
	}
	// Otherwise evaluate expression
	if action.Expression != "" {
		result, err := e.fhirpath.EvaluateString(resource, action.Expression)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("[error evaluating %q: %v]", action.Expression, err))
			return nil
		}
		output.Logs = append(output.Logs, result)
		return nil
	}
	output.Logs = append(output.Logs, "")
	return nil
}

func (e *BotEngine) executeCondition(ctx context.Context, action BotAction, resource map[string]interface{}, output *BotOutput, modified *bool) error {
	if action.Expression == "" {
		return fmt.Errorf("condition action requires expression")
	}

	result, err := e.fhirpath.EvaluateBool(resource, action.Expression)
	if err != nil {
		// Evaluation error defaults to false branch
		result = false
	}

	if result {
		if len(action.OnTrue) > 0 {
			return e.executeActions(ctx, action.OnTrue, resource, output, modified)
		}
	} else {
		if len(action.OnFalse) > 0 {
			return e.executeActions(ctx, action.OnFalse, resource, output, modified)
		}
	}
	return nil
}

func (e *BotEngine) executeTransform(action BotAction, resource map[string]interface{}, output *BotOutput, modified *bool) error {
	if action.Target == "" {
		return fmt.Errorf("transform action requires target path")
	}

	var value interface{}
	if action.Expression != "" {
		// Evaluate expression against resource
		result, err := e.fhirpath.Evaluate(resource, action.Expression)
		if err != nil {
			return fmt.Errorf("transform expression error: %w", err)
		}
		if len(result) > 0 {
			value = result[0]
		}
	} else {
		value = action.Value
	}

	setNestedField(resource, action.Target, value)
	*modified = true
	return nil
}

func (e *BotEngine) executeCreate(action BotAction, resource map[string]interface{}, output *BotOutput) error {
	// Convert Value to a map
	newResource := make(map[string]interface{})

	switch v := action.Value.(type) {
	case map[string]interface{}:
		for k, val := range v {
			newResource[k] = val
		}
	default:
		// Try JSON marshal/unmarshal for other types
		data, err := json.Marshal(action.Value)
		if err != nil {
			return fmt.Errorf("create action: cannot marshal value: %w", err)
		}
		if err := json.Unmarshal(data, &newResource); err != nil {
			return fmt.Errorf("create action: value must be a JSON object: %w", err)
		}
	}

	output.OutputResources = append(output.OutputResources, newResource)
	return nil
}

func (e *BotEngine) executeValidate(action BotAction, resource map[string]interface{}, output *BotOutput) error {
	if action.Expression == "" {
		return fmt.Errorf("validate action requires expression")
	}

	result, err := e.fhirpath.EvaluateBool(resource, action.Expression)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	if !result {
		return fmt.Errorf("validation failed: expression %q evaluated to false", action.Expression)
	}
	return nil
}

func (e *BotEngine) executeWebhook(ctx context.Context, action BotAction, resource map[string]interface{}, output *BotOutput) error {
	urlVal, ok := action.Config["url"]
	if !ok {
		return fmt.Errorf("webhook action requires config.url")
	}
	url, ok := urlVal.(string)
	if !ok || url == "" {
		return fmt.Errorf("webhook action requires a non-empty config.url string")
	}

	payload, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("webhook: failed to marshal resource: %w", err)
	}

	client := &http.Client{Timeout: e.webhookTimeout}

	// Use execution context for cancellation
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("webhook: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook delivery failed: %w", err)
	}
	defer resp.Body.Close()
	io.ReadAll(io.LimitReader(resp.Body, 1024)) // drain

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}

	output.Logs = append(output.Logs, fmt.Sprintf("webhook delivered to %s (status %d)", url, resp.StatusCode))
	return nil
}

func (e *BotEngine) executeSendNotification(action BotAction, resource map[string]interface{}, output *BotOutput) error {
	message := ""
	if s, ok := action.Value.(string); ok {
		message = s
	} else if action.Expression != "" {
		result, err := e.fhirpath.EvaluateString(resource, action.Expression)
		if err == nil {
			message = result
		}
	}

	severity := "info"
	if action.Config != nil {
		if s, ok := action.Config["severity"].(string); ok {
			severity = s
		}
	}

	output.Logs = append(output.Logs, fmt.Sprintf("notification [%s]: %s", severity, message))
	return nil
}

func (e *BotEngine) executeSetStatus(action BotAction, resource map[string]interface{}, output *BotOutput, modified *bool) error {
	statusVal, ok := action.Value.(string)
	if !ok {
		return fmt.Errorf("set-status action requires a string value")
	}
	resource["status"] = statusVal
	*modified = true
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// countActions recursively counts all actions including nested branches.
func countActions(actions []BotAction) int {
	count := 0
	for _, a := range actions {
		count++
		count += countActions(a.OnTrue)
		count += countActions(a.OnFalse)
	}
	return count
}

// deepCopyMap creates a deep copy of a map[string]interface{}.
func deepCopyMap(m map[string]interface{}) map[string]interface{} {
	data, err := json.Marshal(m)
	if err != nil {
		// Fallback: shallow copy
		result := make(map[string]interface{}, len(m))
		for k, v := range m {
			result[k] = v
		}
		return result
	}
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}

// setNestedField sets a value at a dot-separated path in a map.
func setNestedField(m map[string]interface{}, path string, value interface{}) {
	parts := strings.Split(path, ".")
	current := m
	for i := 0; i < len(parts)-1; i++ {
		next, ok := current[parts[i]]
		if !ok {
			// Create intermediate map
			newMap := make(map[string]interface{})
			current[parts[i]] = newMap
			current = newMap
			continue
		}
		nextMap, ok := next.(map[string]interface{})
		if !ok {
			// Overwrite non-map with map
			newMap := make(map[string]interface{})
			current[parts[i]] = newMap
			current = newMap
			continue
		}
		current = nextMap
	}
	current[parts[len(parts)-1]] = value
}

// ---------------------------------------------------------------------------
// Example bots
// ---------------------------------------------------------------------------

// RegisterExampleBots registers 3 built-in example bots that showcase the system.
func RegisterExampleBots(e *BotEngine) {
	// 1. Lab Critical Alert Bot
	e.RegisterBot(Bot{
		ID:          "example-lab-critical-alert",
		Name:        "Lab Critical Alert",
		Description: "Monitors new observations for critical lab values and creates alert flags",
		Status:      "active",
		Trigger: BotTrigger{
			Type:         "subscription",
			ResourceType: "Observation",
			Event:        "create",
			Criteria:     "status = 'final'",
		},
		Code: `[
			{
				"type": "condition",
				"expression": "valueQuantity.value > 200",
				"on_true": [
					{"type": "log", "expression": "subject.reference"},
					{
						"type": "create",
						"target": "Flag",
						"value": {
							"resourceType": "Flag",
							"status": "active",
							"category": [{"coding": [{"system": "http://terminology.hl7.org/CodeSystem/flag-category", "code": "clinical", "display": "Clinical"}]}],
							"code": {"text": "Critical Lab Value: High Glucose"},
							"subject": {"reference": "Patient/unknown"}
						}
					}
				],
				"on_false": [
					{"type": "log", "value": "Observation within normal range"}
				]
			}
		]`,
		Runtime: "fhirpath",
	})

	// 2. New Patient Welcome Bot
	e.RegisterBot(Bot{
		ID:          "example-new-patient-welcome",
		Name:        "New Patient Welcome",
		Description: "Creates a welcome task when a new patient is registered",
		Status:      "active",
		Trigger: BotTrigger{
			Type:         "subscription",
			ResourceType: "Patient",
			Event:        "create",
		},
		Code: `[
			{"type": "log", "expression": "name[0].family"},
			{
				"type": "create",
				"target": "Task",
				"value": {
					"resourceType": "Task",
					"status": "requested",
					"intent": "proposal",
					"description": "Send welcome packet to new patient",
					"priority": "routine"
				}
			}
		]`,
		Runtime: "fhirpath",
	})

	// 3. Auto-Complete Encounter Bot
	e.RegisterBot(Bot{
		ID:          "example-auto-complete-encounter",
		Name:        "Auto-Complete Encounter",
		Description: "Automatically sets period.end when an encounter is finished",
		Status:      "active",
		Trigger: BotTrigger{
			Type:         "subscription",
			ResourceType: "Encounter",
			Event:        "update",
			Criteria:     "status = 'finished'",
		},
		Code: `[
			{
				"type": "condition",
				"expression": "period.end.exists()",
				"on_true": [
					{"type": "log", "value": "Encounter already has period.end"}
				],
				"on_false": [
					{"type": "transform", "target": "period.end", "value": "` + time.Now().UTC().Format(time.RFC3339) + `"},
					{"type": "log", "expression": "id"}
				]
			}
		]`,
		Runtime: "fhirpath",
	})
}

// ---------------------------------------------------------------------------
// HTTP Handler
// ---------------------------------------------------------------------------

// BotHandler exposes bot management via Echo HTTP routes.
type BotHandler struct {
	engine *BotEngine
}

// NewBotHandler creates a new BotHandler.
func NewBotHandler(engine *BotEngine) *BotHandler {
	return &BotHandler{engine: engine}
}

// RegisterRoutes binds all bot management routes to the given Echo group.
func (h *BotHandler) RegisterRoutes(g *echo.Group) {
	g.GET("", h.ListBots)
	g.POST("", h.CreateBot)
	g.GET("/logs", h.ListAllLogs)
	g.GET("/:id", h.GetBot)
	g.PUT("/:id", h.UpdateBot)
	g.DELETE("/:id", h.DeleteBot)
	g.POST("/:id/execute", h.ExecuteBot)
	g.GET("/:id/logs", h.GetBotLogs)
	g.POST("/:id/activate", h.ActivateBot)
	g.POST("/:id/deactivate", h.DeactivateBot)
}

// CreateBot handles POST /bots.
func (h *BotHandler) CreateBot(c echo.Context) error {
	var bot Bot
	if err := c.Bind(&bot); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := h.engine.RegisterBot(bot); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	registered, _ := h.engine.GetBot(bot.ID)
	return c.JSON(http.StatusCreated, registered)
}

// ListBots handles GET /bots.
func (h *BotHandler) ListBots(c echo.Context) error {
	status := c.QueryParam("status")
	bots := h.engine.ListBots(status)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  bots,
		"total": len(bots),
	})
}

// GetBot handles GET /bots/:id.
func (h *BotHandler) GetBot(c echo.Context) error {
	id := c.Param("id")
	bot, err := h.engine.GetBot(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "bot not found"})
	}
	return c.JSON(http.StatusOK, bot)
}

// UpdateBot handles PUT /bots/:id.
func (h *BotHandler) UpdateBot(c echo.Context) error {
	id := c.Param("id")
	var bot Bot
	if err := c.Bind(&bot); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	bot.ID = id
	if err := h.engine.RegisterBot(bot); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	updated, _ := h.engine.GetBot(id)
	return c.JSON(http.StatusOK, updated)
}

// DeleteBot handles DELETE /bots/:id.
func (h *BotHandler) DeleteBot(c echo.Context) error {
	id := c.Param("id")
	if err := h.engine.DeleteBot(id); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "bot not found"})
	}
	return c.NoContent(http.StatusNoContent)
}

// ExecuteBot handles POST /bots/:id/execute.
func (h *BotHandler) ExecuteBot(c echo.Context) error {
	id := c.Param("id")

	// Verify bot exists
	_, err := h.engine.GetBot(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "bot not found"})
	}

	var input BotInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	output, err := h.engine.Execute(c.Request().Context(), id, input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, output)
}

// GetBotLogs handles GET /bots/:id/logs.
func (h *BotHandler) GetBotLogs(c echo.Context) error {
	id := c.Param("id")
	logs := h.engine.GetExecutionLogs(id)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  logs,
		"total": len(logs),
	})
}

// ListAllLogs handles GET /bots/logs.
func (h *BotHandler) ListAllLogs(c echo.Context) error {
	logs := h.engine.GetAllExecutionLogs()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  logs,
		"total": len(logs),
	})
}

// ActivateBot handles POST /bots/:id/activate.
func (h *BotHandler) ActivateBot(c echo.Context) error {
	id := c.Param("id")
	bot, err := h.engine.GetBot(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "bot not found"})
	}
	bot.Status = "active"
	h.engine.RegisterBot(*bot)
	return c.JSON(http.StatusOK, map[string]string{"status": "active"})
}

// DeactivateBot handles POST /bots/:id/deactivate.
func (h *BotHandler) DeactivateBot(c echo.Context) error {
	id := c.Param("id")
	bot, err := h.engine.GetBot(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "bot not found"})
	}
	bot.Status = "inactive"
	h.engine.RegisterBot(*bot)
	return c.JSON(http.StatusOK, map[string]string{"status": "inactive"})
}
