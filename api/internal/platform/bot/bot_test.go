package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ===========================================================================
// Helpers
// ===========================================================================

func newTestEngine() *BotEngine {
	return NewBotEngine()
}

func mustRegisterBot(t *testing.T, e *BotEngine, bot Bot) {
	t.Helper()
	if err := e.RegisterBot(bot); err != nil {
		t.Fatalf("RegisterBot failed: %v", err)
	}
}

func sampleBot(id, name string) Bot {
	return Bot{
		ID:     id,
		Name:   name,
		Status: "active",
		Trigger: BotTrigger{
			Type:         "subscription",
			ResourceType: "Patient",
			Event:        "create",
		},
		Code:    `[{"type":"log","expression":"id"}]`,
		Runtime: "fhirpath",
		Config:  map[string]string{"key": "value"},
	}
}

func samplePatientResource() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-123",
		"status":       "active",
		"name": []interface{}{
			map[string]interface{}{
				"family": "Smith",
				"given":  []interface{}{"John"},
			},
		},
	}
}

func sampleObservationResource() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-456",
		"status":       "final",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://loinc.org",
					"code":    "2345-7",
					"display": "Glucose",
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/pt-123",
		},
		"valueQuantity": map[string]interface{}{
			"value": float64(250),
			"unit":  "mg/dL",
		},
	}
}

func sampleEncounterResource() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Encounter",
		"id":           "enc-789",
		"status":       "finished",
		"class": map[string]interface{}{
			"code": "AMB",
		},
		"period": map[string]interface{}{
			"start": "2024-01-15T09:00:00Z",
		},
	}
}

// echoContext creates a minimal echo.Context for handler tests.
func echoContext(method, path string, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// ===========================================================================
// Bot CRUD Tests
// ===========================================================================

func TestRegisterBot(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Test Bot")
	err := e.RegisterBot(bot)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	got, err := e.GetBot("bot-1")
	if err != nil {
		t.Fatalf("GetBot failed: %v", err)
	}
	if got.Name != "Test Bot" {
		t.Errorf("expected name 'Test Bot', got %q", got.Name)
	}
	if got.Status != "active" {
		t.Errorf("expected status 'active', got %q", got.Status)
	}
	if got.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if got.RunCount != 0 {
		t.Errorf("expected RunCount 0, got %d", got.RunCount)
	}
}

func TestRegisterBot_MissingName(t *testing.T) {
	e := newTestEngine()
	bot := Bot{
		ID:     "bot-1",
		Status: "active",
		Trigger: BotTrigger{
			Type: "manual",
		},
		Code:    `[]`,
		Runtime: "fhirpath",
	}
	err := e.RegisterBot(bot)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error about name, got: %v", err)
	}
}

func TestRegisterBot_MissingID(t *testing.T) {
	e := newTestEngine()
	bot := Bot{
		Name:   "Test Bot",
		Status: "active",
		Trigger: BotTrigger{
			Type: "manual",
		},
		Code:    `[]`,
		Runtime: "fhirpath",
	}
	err := e.RegisterBot(bot)
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("expected error about id, got: %v", err)
	}
}

func TestRegisterBot_MissingTriggerType(t *testing.T) {
	e := newTestEngine()
	bot := Bot{
		ID:     "bot-1",
		Name:   "Test Bot",
		Status: "active",
		Trigger: BotTrigger{
			ResourceType: "Patient",
		},
		Code:    `[]`,
		Runtime: "fhirpath",
	}
	err := e.RegisterBot(bot)
	if err == nil {
		t.Fatal("expected error for missing trigger type")
	}
}

func TestRegisterBot_InvalidStatus(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Test Bot")
	bot.Status = "unknown"
	err := e.RegisterBot(bot)
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestGetBot(t *testing.T) {
	e := newTestEngine()
	mustRegisterBot(t, e, sampleBot("bot-1", "Test Bot"))
	got, err := e.GetBot("bot-1")
	if err != nil {
		t.Fatalf("GetBot failed: %v", err)
	}
	if got.ID != "bot-1" {
		t.Errorf("expected ID 'bot-1', got %q", got.ID)
	}
}

func TestGetBot_NotFound(t *testing.T) {
	e := newTestEngine()
	_, err := e.GetBot("nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestListBots(t *testing.T) {
	e := newTestEngine()
	mustRegisterBot(t, e, sampleBot("bot-1", "Bot 1"))
	mustRegisterBot(t, e, sampleBot("bot-2", "Bot 2"))

	inactive := sampleBot("bot-3", "Bot 3")
	inactive.Status = "inactive"
	mustRegisterBot(t, e, inactive)

	all := e.ListBots("")
	if len(all) != 3 {
		t.Errorf("expected 3 bots, got %d", len(all))
	}
}

func TestListBots_FilterByStatus(t *testing.T) {
	e := newTestEngine()
	mustRegisterBot(t, e, sampleBot("bot-1", "Bot 1"))

	inactive := sampleBot("bot-2", "Bot 2")
	inactive.Status = "inactive"
	mustRegisterBot(t, e, inactive)

	active := e.ListBots("active")
	if len(active) != 1 {
		t.Errorf("expected 1 active bot, got %d", len(active))
	}
	if active[0].ID != "bot-1" {
		t.Errorf("expected bot-1, got %q", active[0].ID)
	}

	inactiveList := e.ListBots("inactive")
	if len(inactiveList) != 1 {
		t.Errorf("expected 1 inactive bot, got %d", len(inactiveList))
	}
}

func TestUpdateBot(t *testing.T) {
	e := newTestEngine()
	mustRegisterBot(t, e, sampleBot("bot-1", "Original"))

	updated := sampleBot("bot-1", "Updated")
	updated.Description = "Updated description"
	err := e.RegisterBot(updated)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	got, _ := e.GetBot("bot-1")
	if got.Name != "Updated" {
		t.Errorf("expected name 'Updated', got %q", got.Name)
	}
	if got.Description != "Updated description" {
		t.Errorf("expected updated description, got %q", got.Description)
	}
}

func TestDeleteBot(t *testing.T) {
	e := newTestEngine()
	mustRegisterBot(t, e, sampleBot("bot-1", "Test Bot"))

	err := e.DeleteBot("bot-1")
	if err != nil {
		t.Fatalf("DeleteBot failed: %v", err)
	}

	_, err = e.GetBot("bot-1")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestDeleteBot_NotFound(t *testing.T) {
	e := newTestEngine()
	err := e.DeleteBot("nonexistent")
	if err == nil {
		t.Fatal("expected error for deleting nonexistent bot")
	}
}

// ===========================================================================
// Action Execution Tests
// ===========================================================================

func TestAction_LogWithExpression(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Logger")
	bot.Code = `[{"type":"log","expression":"id"}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Status != "success" {
		t.Errorf("expected success, got %q (error: %s)", out.Status, out.Error)
	}
	if len(out.Logs) == 0 {
		t.Fatal("expected logs to be non-empty")
	}
	if !strings.Contains(out.Logs[0], "pt-123") {
		t.Errorf("expected log to contain 'pt-123', got %q", out.Logs[0])
	}
}

func TestAction_LogWithLiteralString(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Logger")
	bot.Code = `[{"type":"log","value":"Hello World"}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(out.Logs) == 0 {
		t.Fatal("expected logs")
	}
	if out.Logs[0] != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", out.Logs[0])
	}
}

func TestAction_ConditionTrueBranch(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Condition")
	bot.Code = `[{
		"type":"condition",
		"expression":"status = 'active'",
		"on_true":[{"type":"log","value":"is active"}],
		"on_false":[{"type":"log","value":"not active"}]
	}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(out.Logs) == 0 || out.Logs[0] != "is active" {
		t.Errorf("expected 'is active' log, got %v", out.Logs)
	}
}

func TestAction_ConditionFalseBranch(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Condition")
	bot.Code = `[{
		"type":"condition",
		"expression":"status = 'inactive'",
		"on_true":[{"type":"log","value":"is inactive"}],
		"on_false":[{"type":"log","value":"not inactive"}]
	}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(out.Logs) == 0 || out.Logs[0] != "not inactive" {
		t.Errorf("expected 'not inactive' log, got %v", out.Logs)
	}
}

func TestAction_TransformSetsField(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Transform")
	bot.Code = `[{"type":"transform","target":"status","value":"completed"}]`
	mustRegisterBot(t, e, bot)

	resource := samplePatientResource()
	input := BotInput{
		Resource:     resource,
		ResourceType: "Patient",
		Event:        "update",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Status != "success" {
		t.Errorf("expected success, got %q (error: %s)", out.Status, out.Error)
	}
	if len(out.OutputResources) == 0 {
		t.Fatal("expected output resources")
	}
	if out.OutputResources[0]["status"] != "completed" {
		t.Errorf("expected status 'completed', got %v", out.OutputResources[0]["status"])
	}
}

func TestAction_CreateAddsToOutput(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Creator")
	bot.Code = `[{"type":"create","target":"Task","value":{"resourceType":"Task","status":"requested","description":"Follow up"}}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(out.OutputResources) == 0 {
		t.Fatal("expected output resources from create action")
	}
	found := false
	for _, r := range out.OutputResources {
		if rt, ok := r["resourceType"]; ok && rt == "Task" {
			found = true
			if r["status"] != "requested" {
				t.Errorf("expected Task status 'requested', got %v", r["status"])
			}
		}
	}
	if !found {
		t.Error("expected Task resource in output")
	}
}

func TestAction_ValidatePasses(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Validator")
	bot.Code = `[{"type":"validate","expression":"resourceType.exists()"}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Status != "success" {
		t.Errorf("expected success, got %q (error: %s)", out.Status, out.Error)
	}
}

func TestAction_ValidateFails(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Validator")
	bot.Code = `[{"type":"validate","expression":"nonexistent.exists()"}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Status != "error" {
		t.Errorf("expected error status for failed validation, got %q", out.Status)
	}
}

func TestAction_SetStatus(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "StatusSetter")
	bot.Code = `[{"type":"set-status","value":"completed"}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "update",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(out.OutputResources) == 0 {
		t.Fatal("expected output resources")
	}
	if out.OutputResources[0]["status"] != "completed" {
		t.Errorf("expected status 'completed', got %v", out.OutputResources[0]["status"])
	}
}

func TestAction_Webhook(t *testing.T) {
	var receivedBody []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	e := newTestEngine()
	bot := sampleBot("bot-1", "Webhook")
	bot.Code = fmt.Sprintf(`[{"type":"webhook","config":{"url":%q}}]`, ts.URL)
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Status != "success" {
		t.Errorf("expected success, got %q (error: %s)", out.Status, out.Error)
	}
	if len(receivedBody) == 0 {
		t.Error("expected webhook to receive body")
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(receivedBody, &payload); err != nil {
		t.Fatalf("expected valid JSON payload, got: %s", receivedBody)
	}
	if payload["id"] != "pt-123" {
		t.Errorf("expected resource id 'pt-123' in payload, got: %v", payload["id"])
	}
}

func TestAction_WebhookTimeout(t *testing.T) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	ts.Start()
	defer ts.Close()

	e := newTestEngine()
	// Override webhook timeout to make test fast
	e.webhookTimeout = 50 * time.Millisecond
	bot := sampleBot("bot-1", "WebhookTimeout")
	bot.Code = fmt.Sprintf(`[{"type":"webhook","config":{"url":%q}}]`, ts.URL)
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Status != "error" {
		t.Errorf("expected error status due to timeout, got %q", out.Status)
	}
}

func TestAction_MultipleSequentialActions(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Multi")
	bot.Code = `[
		{"type":"log","value":"step 1"},
		{"type":"log","value":"step 2"},
		{"type":"log","value":"step 3"}
	]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(out.Logs) != 3 {
		t.Errorf("expected 3 logs, got %d: %v", len(out.Logs), out.Logs)
	}
	if out.ActionsExecuted != 3 {
		t.Errorf("expected 3 actions executed, got %d", out.ActionsExecuted)
	}
}

func TestAction_NestedConditionActions(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Nested")
	bot.Code = `[{
		"type":"condition",
		"expression":"status = 'active'",
		"on_true":[{
			"type":"condition",
			"expression":"id = 'pt-123'",
			"on_true":[{"type":"log","value":"matched both"}],
			"on_false":[{"type":"log","value":"matched first only"}]
		}],
		"on_false":[{"type":"log","value":"matched neither"}]
	}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(out.Logs) == 0 || out.Logs[0] != "matched both" {
		t.Errorf("expected 'matched both', got %v", out.Logs)
	}
}

func TestAction_MaxActionLimit(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "TooMany")

	actions := make([]BotAction, 101)
	for i := 0; i < 101; i++ {
		actions[i] = BotAction{Type: "log", Value: fmt.Sprintf("step %d", i)}
	}
	code, _ := json.Marshal(actions)
	bot.Code = string(code)
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Status != "error" {
		t.Errorf("expected error status for action limit, got %q", out.Status)
	}
	if !strings.Contains(out.Error, "action limit") {
		t.Errorf("expected action limit error, got: %s", out.Error)
	}
}

func TestAction_ExecutionTimeout(t *testing.T) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	ts.Start()
	defer ts.Close()

	e := newTestEngine()
	e.executionTimeout = 50 * time.Millisecond
	e.webhookTimeout = 5 * time.Second
	bot := sampleBot("bot-1", "Timeout")
	bot.Code = fmt.Sprintf(`[{"type":"webhook","config":{"url":%q}}]`, ts.URL)
	mustRegisterBot(t, e, bot)

	ctx := context.Background()
	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(ctx, "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Status != "error" {
		t.Errorf("expected error status for timeout, got %q", out.Status)
	}
}

func TestAction_EmptyActionsList(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Empty")
	bot.Code = `[]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Status != "success" {
		t.Errorf("expected success, got %q", out.Status)
	}
	if out.ActionsExecuted != 0 {
		t.Errorf("expected 0 actions executed, got %d", out.ActionsExecuted)
	}
}

func TestAction_SendNotification(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Notifier")
	bot.Code = `[{"type":"send-notification","value":"Patient needs follow-up","config":{"severity":"high"}}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Status != "success" {
		t.Errorf("expected success, got %q (error: %s)", out.Status, out.Error)
	}
	if len(out.Logs) == 0 {
		t.Fatal("expected notification log")
	}
	if !strings.Contains(out.Logs[0], "notification") || !strings.Contains(out.Logs[0], "Patient needs follow-up") {
		t.Errorf("expected notification log, got: %s", out.Logs[0])
	}
}

func TestAction_TransformWithExpression(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "ExprTransform")
	bot.Code = `[{"type":"transform","target":"displayName","expression":"name[0].family"}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "update",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(out.OutputResources) == 0 {
		t.Fatal("expected output resources")
	}
	if out.OutputResources[0]["displayName"] != "Smith" {
		t.Errorf("expected displayName 'Smith', got %v", out.OutputResources[0]["displayName"])
	}
}

// ===========================================================================
// Trigger Matching Tests
// ===========================================================================

func TestTrigger_MatchByResourceTypeAndEvent(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Matcher")
	bot.Trigger.ResourceType = "Patient"
	bot.Trigger.Event = "create"
	mustRegisterBot(t, e, bot)

	results := e.ExecuteByTrigger(context.Background(), "Patient", "create", samplePatientResource())
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].BotID != "bot-1" {
		t.Errorf("expected bot-1, got %q", results[0].BotID)
	}
}

func TestTrigger_MatchByResourceTypeWildcardEvent(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Wildcard")
	bot.Trigger.ResourceType = "Patient"
	bot.Trigger.Event = "*"
	mustRegisterBot(t, e, bot)

	for _, event := range []string{"create", "update", "delete"} {
		results := e.ExecuteByTrigger(context.Background(), "Patient", event, samplePatientResource())
		if len(results) != 1 {
			t.Errorf("expected 1 result for event %q, got %d", event, len(results))
		}
	}
}

func TestTrigger_NoMatchDifferentResourceType(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "PatientOnly")
	bot.Trigger.ResourceType = "Patient"
	bot.Trigger.Event = "create"
	mustRegisterBot(t, e, bot)

	results := e.ExecuteByTrigger(context.Background(), "Observation", "create", sampleObservationResource())
	if len(results) != 0 {
		t.Errorf("expected 0 results for different resource type, got %d", len(results))
	}
}

func TestTrigger_CriteriaFiltering(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "CriteriaBot")
	bot.Trigger.ResourceType = "Observation"
	bot.Trigger.Event = "create"
	bot.Trigger.Criteria = "status = 'final'"
	mustRegisterBot(t, e, bot)

	results := e.ExecuteByTrigger(context.Background(), "Observation", "create", sampleObservationResource())
	if len(results) != 1 {
		t.Errorf("expected 1 result for matching criteria, got %d", len(results))
	}

	obs := sampleObservationResource()
	obs["status"] = "preliminary"
	results = e.ExecuteByTrigger(context.Background(), "Observation", "create", obs)
	if len(results) != 0 {
		t.Errorf("expected 0 results for non-matching criteria, got %d", len(results))
	}
}

func TestTrigger_InactiveBotNotTriggered(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Inactive")
	bot.Status = "inactive"
	mustRegisterBot(t, e, bot)

	results := e.ExecuteByTrigger(context.Background(), "Patient", "create", samplePatientResource())
	if len(results) != 0 {
		t.Errorf("expected 0 results for inactive bot, got %d", len(results))
	}
}

func TestTrigger_MultipleBotsTriggered(t *testing.T) {
	e := newTestEngine()
	bot1 := sampleBot("bot-1", "Bot 1")
	bot1.Trigger.ResourceType = "Patient"
	bot1.Trigger.Event = "create"
	mustRegisterBot(t, e, bot1)

	bot2 := sampleBot("bot-2", "Bot 2")
	bot2.Trigger.ResourceType = "Patient"
	bot2.Trigger.Event = "create"
	mustRegisterBot(t, e, bot2)

	results := e.ExecuteByTrigger(context.Background(), "Patient", "create", samplePatientResource())
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestTrigger_NoMatchDifferentEvent(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "CreateOnly")
	bot.Trigger.ResourceType = "Patient"
	bot.Trigger.Event = "create"
	mustRegisterBot(t, e, bot)

	results := e.ExecuteByTrigger(context.Background(), "Patient", "delete", samplePatientResource())
	if len(results) != 0 {
		t.Errorf("expected 0 results for different event, got %d", len(results))
	}
}

// ===========================================================================
// Example Bot Tests
// ===========================================================================

func TestExampleBot_LabCriticalAlert(t *testing.T) {
	e := newTestEngine()
	RegisterExampleBots(e)

	obs := sampleObservationResource()
	results := e.ExecuteByTrigger(context.Background(), "Observation", "create", obs)

	var labResult *BotOutput
	for i, r := range results {
		if r.BotName == "Lab Critical Alert" {
			labResult = &results[i]
			break
		}
	}
	if labResult == nil {
		t.Fatal("expected Lab Critical Alert bot to trigger")
	}
	if labResult.Status != "success" {
		t.Errorf("expected success, got %q (error: %s)", labResult.Status, labResult.Error)
	}
	foundFlag := false
	for _, r := range labResult.OutputResources {
		if rt, ok := r["resourceType"]; ok && rt == "Flag" {
			foundFlag = true
		}
	}
	if !foundFlag {
		t.Error("expected Flag resource in output for critical alert")
	}
}

func TestExampleBot_NewPatientWelcome(t *testing.T) {
	e := newTestEngine()
	RegisterExampleBots(e)

	results := e.ExecuteByTrigger(context.Background(), "Patient", "create", samplePatientResource())

	var welcomeResult *BotOutput
	for i, r := range results {
		if r.BotName == "New Patient Welcome" {
			welcomeResult = &results[i]
			break
		}
	}
	if welcomeResult == nil {
		t.Fatal("expected New Patient Welcome bot to trigger")
	}
	if welcomeResult.Status != "success" {
		t.Errorf("expected success, got %q (error: %s)", welcomeResult.Status, welcomeResult.Error)
	}
	foundTask := false
	for _, r := range welcomeResult.OutputResources {
		if rt, ok := r["resourceType"]; ok && rt == "Task" {
			foundTask = true
		}
	}
	if !foundTask {
		t.Error("expected Task resource in output for welcome")
	}
}

func TestExampleBot_AutoCompleteEncounter(t *testing.T) {
	e := newTestEngine()
	RegisterExampleBots(e)

	enc := sampleEncounterResource()
	results := e.ExecuteByTrigger(context.Background(), "Encounter", "update", enc)

	var encounterResult *BotOutput
	for i, r := range results {
		if r.BotName == "Auto-Complete Encounter" {
			encounterResult = &results[i]
			break
		}
	}
	if encounterResult == nil {
		t.Fatal("expected Auto-Complete Encounter bot to trigger")
	}
	if encounterResult.Status != "success" {
		t.Errorf("expected success, got %q (error: %s)", encounterResult.Status, encounterResult.Error)
	}
	if len(encounterResult.OutputResources) == 0 {
		t.Fatal("expected output resources")
	}
	outResource := encounterResult.OutputResources[0]
	period, ok := outResource["period"].(map[string]interface{})
	if !ok {
		t.Fatal("expected period in output resource")
	}
	if _, hasEnd := period["end"]; !hasEnd {
		t.Error("expected period.end to be set")
	}
}

// ===========================================================================
// Handler Tests
// ===========================================================================

func TestHandler_CreateBot(t *testing.T) {
	e := newTestEngine()
	h := NewBotHandler(e)

	body := `{
		"id":"bot-1",
		"name":"API Bot",
		"status":"active",
		"trigger":{"type":"manual"},
		"code":"[{\"type\":\"log\",\"value\":\"hello\"}]",
		"runtime":"fhirpath"
	}`
	c, rec := echoContext(http.MethodPost, "/api/v1/bots", body)
	err := h.CreateBot(c)
	if err != nil {
		t.Fatalf("CreateBot handler error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var result Bot
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result.ID != "bot-1" {
		t.Errorf("expected ID 'bot-1', got %q", result.ID)
	}
}

func TestHandler_ListBots(t *testing.T) {
	e := newTestEngine()
	mustRegisterBot(t, e, sampleBot("bot-1", "Bot 1"))
	mustRegisterBot(t, e, sampleBot("bot-2", "Bot 2"))
	h := NewBotHandler(e)

	c, rec := echoContext(http.MethodGet, "/api/v1/bots", "")
	err := h.ListBots(c)
	if err != nil {
		t.Fatalf("ListBots handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result struct {
		Data  []Bot `json:"data"`
		Total int   `json:"total"`
	}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result.Total != 2 {
		t.Errorf("expected 2 total, got %d", result.Total)
	}
}

func TestHandler_GetBot(t *testing.T) {
	e := newTestEngine()
	mustRegisterBot(t, e, sampleBot("bot-1", "Get Me"))
	h := NewBotHandler(e)

	c, rec := echoContext(http.MethodGet, "/api/v1/bots/bot-1", "")
	c.SetParamNames("id")
	c.SetParamValues("bot-1")
	err := h.GetBot(c)
	if err != nil {
		t.Fatalf("GetBot handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var result Bot
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result.Name != "Get Me" {
		t.Errorf("expected name 'Get Me', got %q", result.Name)
	}
}

func TestHandler_UpdateBot(t *testing.T) {
	e := newTestEngine()
	mustRegisterBot(t, e, sampleBot("bot-1", "Original"))
	h := NewBotHandler(e)

	body := `{
		"id":"bot-1",
		"name":"Updated",
		"status":"active",
		"trigger":{"type":"manual"},
		"code":"[]",
		"runtime":"fhirpath"
	}`
	c, rec := echoContext(http.MethodPut, "/api/v1/bots/bot-1", body)
	c.SetParamNames("id")
	c.SetParamValues("bot-1")
	err := h.UpdateBot(c)
	if err != nil {
		t.Fatalf("UpdateBot handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteBot(t *testing.T) {
	e := newTestEngine()
	mustRegisterBot(t, e, sampleBot("bot-1", "Delete Me"))
	h := NewBotHandler(e)

	c, rec := echoContext(http.MethodDelete, "/api/v1/bots/bot-1", "")
	c.SetParamNames("id")
	c.SetParamValues("bot-1")
	err := h.DeleteBot(c)
	if err != nil {
		t.Fatalf("DeleteBot handler error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_ExecuteBot(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Execute Me")
	bot.Code = `[{"type":"log","value":"executed!"}]`
	bot.Trigger.Type = "manual"
	mustRegisterBot(t, e, bot)
	h := NewBotHandler(e)

	body := `{"resource":{"resourceType":"Patient","id":"pt-1"},"resource_type":"Patient","event":"manual"}`
	c, rec := echoContext(http.MethodPost, "/api/v1/bots/bot-1/execute", body)
	c.SetParamNames("id")
	c.SetParamValues("bot-1")
	err := h.ExecuteBot(c)
	if err != nil {
		t.Fatalf("ExecuteBot handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result BotOutput
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result.Status != "success" {
		t.Errorf("expected success, got %q", result.Status)
	}
}

func TestHandler_GetExecutionLogs(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Logger Bot")
	bot.Code = `[{"type":"log","value":"logged"}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	e.Execute(context.Background(), "bot-1", input)

	h := NewBotHandler(e)
	c, rec := echoContext(http.MethodGet, "/api/v1/bots/bot-1/logs", "")
	c.SetParamNames("id")
	c.SetParamValues("bot-1")
	err := h.GetBotLogs(c)
	if err != nil {
		t.Fatalf("GetBotLogs handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result struct {
		Data []BotExecutionLog `json:"data"`
	}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if len(result.Data) == 0 {
		t.Error("expected at least one execution log")
	}
}

func TestHandler_ListAllLogs(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Bot 1")
	bot.Code = `[{"type":"log","value":"test"}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	e.Execute(context.Background(), "bot-1", input)

	h := NewBotHandler(e)
	c, rec := echoContext(http.MethodGet, "/api/v1/bots/logs", "")
	err := h.ListAllLogs(c)
	if err != nil {
		t.Fatalf("ListAllLogs handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ActivateBot(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Inactive Bot")
	bot.Status = "inactive"
	mustRegisterBot(t, e, bot)
	h := NewBotHandler(e)

	c, rec := echoContext(http.MethodPost, "/api/v1/bots/bot-1/activate", "")
	c.SetParamNames("id")
	c.SetParamValues("bot-1")
	err := h.ActivateBot(c)
	if err != nil {
		t.Fatalf("ActivateBot handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	got, _ := e.GetBot("bot-1")
	if got.Status != "active" {
		t.Errorf("expected status 'active', got %q", got.Status)
	}
}

func TestHandler_DeactivateBot(t *testing.T) {
	e := newTestEngine()
	mustRegisterBot(t, e, sampleBot("bot-1", "Active Bot"))
	h := NewBotHandler(e)

	c, rec := echoContext(http.MethodPost, "/api/v1/bots/bot-1/deactivate", "")
	c.SetParamNames("id")
	c.SetParamValues("bot-1")
	err := h.DeactivateBot(c)
	if err != nil {
		t.Fatalf("DeactivateBot handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	got, _ := e.GetBot("bot-1")
	if got.Status != "inactive" {
		t.Errorf("expected status 'inactive', got %q", got.Status)
	}
}

func TestHandler_InvalidBot_MissingTrigger(t *testing.T) {
	e := newTestEngine()
	h := NewBotHandler(e)

	body := `{"id":"bot-1","name":"Bad Bot","status":"active","code":"[]","runtime":"fhirpath"}`
	c, rec := echoContext(http.MethodPost, "/api/v1/bots", body)
	err := h.CreateBot(c)
	if err != nil {
		t.Fatalf("CreateBot handler error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_ExecuteNonexistentBot(t *testing.T) {
	e := newTestEngine()
	h := NewBotHandler(e)

	body := `{"resource":{"resourceType":"Patient","id":"pt-1"},"resource_type":"Patient","event":"manual"}`
	c, rec := echoContext(http.MethodPost, "/api/v1/bots/nonexistent/execute", body)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	err := h.ExecuteBot(c)
	if err != nil {
		t.Fatalf("ExecuteBot handler error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_ExecuteInactiveBot(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Inactive")
	bot.Status = "inactive"
	mustRegisterBot(t, e, bot)
	h := NewBotHandler(e)

	body := `{"resource":{"resourceType":"Patient","id":"pt-1"},"resource_type":"Patient","event":"manual"}`
	c, rec := echoContext(http.MethodPost, "/api/v1/bots/bot-1/execute", body)
	c.SetParamNames("id")
	c.SetParamValues("bot-1")
	err := h.ExecuteBot(c)
	if err != nil {
		t.Fatalf("ExecuteBot handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var result BotOutput
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result.Status != "error" {
		t.Errorf("expected error status for inactive bot, got %q", result.Status)
	}
}

// ===========================================================================
// Safety Tests
// ===========================================================================

func TestSafety_TimeoutEnforced(t *testing.T) {
	e := newTestEngine()
	e.executionTimeout = 50 * time.Millisecond

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	ts.Start()
	defer ts.Close()

	bot := sampleBot("bot-1", "Slow Bot")
	bot.Code = fmt.Sprintf(`[{"type":"webhook","config":{"url":%q}}]`, ts.URL)
	mustRegisterBot(t, e, bot)

	start := time.Now()
	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, _ := e.Execute(context.Background(), "bot-1", input)
	elapsed := time.Since(start)

	if elapsed > 5*time.Second {
		t.Errorf("execution took too long: %v (timeout not enforced)", elapsed)
	}
	if out.Status != "error" {
		t.Errorf("expected error status, got %q", out.Status)
	}
}

func TestSafety_ActionLimitEnforced(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Many Actions")

	actions := make([]BotAction, 101)
	for i := range actions {
		actions[i] = BotAction{Type: "log", Value: "x"}
	}
	code, _ := json.Marshal(actions)
	bot.Code = string(code)
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, _ := e.Execute(context.Background(), "bot-1", input)
	if out.Status != "error" {
		t.Errorf("expected error status, got %q", out.Status)
	}
	if !strings.Contains(out.Error, "action limit") {
		t.Errorf("expected action limit error, got: %s", out.Error)
	}
}

func TestSafety_ConcurrentExecution(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Concurrent Bot")
	bot.Code = `[{"type":"log","value":"concurrent"}]`
	mustRegisterBot(t, e, bot)

	var wg sync.WaitGroup
	results := make([]BotOutput, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			input := BotInput{
				Resource:     samplePatientResource(),
				ResourceType: "Patient",
				Event:        "create",
			}
			out, err := e.Execute(context.Background(), "bot-1", input)
			if err != nil {
				t.Errorf("concurrent execute %d failed: %v", idx, err)
				return
			}
			results[idx] = *out
		}(i)
	}
	wg.Wait()

	for i, r := range results {
		if r.Status != "success" {
			t.Errorf("concurrent result %d: expected success, got %q", i, r.Status)
		}
	}
}

func TestSafety_ErrorInOneActionDoesNotCrashOthers(t *testing.T) {
	e := newTestEngine()

	bot1 := sampleBot("bot-1", "Failing Bot")
	bot1.Trigger.ResourceType = "Patient"
	bot1.Trigger.Event = "create"
	bot1.Code = `[{"type":"validate","expression":"nonexistent.exists()"}]`
	mustRegisterBot(t, e, bot1)

	bot2 := sampleBot("bot-2", "Succeeding Bot")
	bot2.Trigger.ResourceType = "Patient"
	bot2.Trigger.Event = "create"
	bot2.Code = `[{"type":"log","value":"I survived"}]`
	mustRegisterBot(t, e, bot2)

	results := e.ExecuteByTrigger(context.Background(), "Patient", "create", samplePatientResource())
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	successCount := 0
	for _, r := range results {
		if r.Status == "success" {
			successCount++
		}
	}
	if successCount == 0 {
		t.Error("expected at least one success when error in another bot")
	}
}

// ===========================================================================
// Execution Log Tests
// ===========================================================================

func TestExecutionLogsRecorded(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Logged Bot")
	bot.Code = `[{"type":"log","value":"test"}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	e.Execute(context.Background(), "bot-1", input)
	e.Execute(context.Background(), "bot-1", input)

	logs := e.GetExecutionLogs("bot-1")
	if len(logs) != 2 {
		t.Errorf("expected 2 execution logs, got %d", len(logs))
	}

	allLogs := e.GetAllExecutionLogs()
	if len(allLogs) != 2 {
		t.Errorf("expected 2 total logs, got %d", len(allLogs))
	}
}

func TestExecutionLogRingBuffer(t *testing.T) {
	e := newTestEngine()
	e.maxLogs = 5

	bot := sampleBot("bot-1", "Ring Buffer Bot")
	bot.Code = `[{"type":"log","value":"test"}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	for i := 0; i < 10; i++ {
		e.Execute(context.Background(), "bot-1", input)
	}

	logs := e.GetAllExecutionLogs()
	if len(logs) > 5 {
		t.Errorf("expected max 5 logs (ring buffer), got %d", len(logs))
	}
}

func TestRunCountIncremented(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Counter Bot")
	bot.Code = `[{"type":"log","value":"test"}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	e.Execute(context.Background(), "bot-1", input)
	e.Execute(context.Background(), "bot-1", input)
	e.Execute(context.Background(), "bot-1", input)

	got, _ := e.GetBot("bot-1")
	if got.RunCount != 3 {
		t.Errorf("expected RunCount 3, got %d", got.RunCount)
	}
	if got.LastRunAt == nil {
		t.Error("expected LastRunAt to be set")
	}
	if got.LastRunStatus != "success" {
		t.Errorf("expected LastRunStatus 'success', got %q", got.LastRunStatus)
	}
}

func TestBotOutputDuration(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Duration Bot")
	bot.Code = `[{"type":"log","value":"test"}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Duration <= 0 {
		t.Error("expected positive duration")
	}
}

// ===========================================================================
// RegisterRoutes Test
// ===========================================================================

func TestHandler_RegisterRoutes(t *testing.T) {
	e := newTestEngine()
	h := NewBotHandler(e)

	ec := echo.New()
	api := ec.Group("/api/v1/bots")
	h.RegisterRoutes(api)

	routes := ec.Routes()
	expectedPaths := map[string]bool{
		"/api/v1/bots":                 false,
		"/api/v1/bots/:id":            false,
		"/api/v1/bots/:id/execute":    false,
		"/api/v1/bots/:id/logs":       false,
		"/api/v1/bots/logs":           false,
		"/api/v1/bots/:id/activate":   false,
		"/api/v1/bots/:id/deactivate": false,
	}
	for _, r := range routes {
		if _, ok := expectedPaths[r.Path]; ok {
			expectedPaths[r.Path] = true
		}
	}
	for path, found := range expectedPaths {
		if !found {
			t.Errorf("expected route %q to be registered", path)
		}
	}
}

// ===========================================================================
// Edge Case Tests
// ===========================================================================

func TestExecute_NonexistentBot(t *testing.T) {
	e := newTestEngine()
	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	_, err := e.Execute(context.Background(), "nonexistent", input)
	if err == nil {
		t.Fatal("expected error for nonexistent bot")
	}
}

func TestExecute_InvalidCode(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Bad Code")
	bot.Code = `not valid json`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Status != "error" {
		t.Errorf("expected error for invalid code, got %q", out.Status)
	}
}

func TestAction_UnknownActionType(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Unknown Action")
	bot.Code = `[{"type":"unknown_action","value":"test"}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Status != "error" {
		t.Errorf("expected error for unknown action type, got %q", out.Status)
	}
}

func TestAction_WebhookMissingURL(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "No URL Webhook")
	bot.Code = `[{"type":"webhook","config":{}}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Status != "error" {
		t.Errorf("expected error for missing webhook URL, got %q", out.Status)
	}
}

func TestAction_CreateWithMapValue(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Map Creator")
	bot.Code = `[{"type":"create","target":"Flag","value":{"resourceType":"Flag","status":"active","code":{"text":"Test Flag"}}}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     samplePatientResource(),
		ResourceType: "Patient",
		Event:        "create",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(out.OutputResources) == 0 {
		t.Fatal("expected output resources")
	}
	flag := out.OutputResources[0]
	if flag["resourceType"] != "Flag" {
		t.Errorf("expected resourceType 'Flag', got %v", flag["resourceType"])
	}
}

func TestAction_TransformNestedPath(t *testing.T) {
	e := newTestEngine()
	bot := sampleBot("bot-1", "Nested Transform")
	bot.Code = `[{"type":"transform","target":"period.end","value":"2024-01-15T17:00:00Z"}]`
	mustRegisterBot(t, e, bot)

	input := BotInput{
		Resource:     sampleEncounterResource(),
		ResourceType: "Encounter",
		Event:        "update",
	}
	out, err := e.Execute(context.Background(), "bot-1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(out.OutputResources) == 0 {
		t.Fatal("expected output resources")
	}
	period, ok := out.OutputResources[0]["period"].(map[string]interface{})
	if !ok {
		t.Fatal("expected period map in output")
	}
	if period["end"] != "2024-01-15T17:00:00Z" {
		t.Errorf("expected period.end, got %v", period["end"])
	}
}

func TestListBots_EmptyEngine(t *testing.T) {
	e := newTestEngine()
	bots := e.ListBots("")
	if len(bots) != 0 {
		t.Errorf("expected 0 bots, got %d", len(bots))
	}
}

func TestHandler_GetBot_NotFound(t *testing.T) {
	e := newTestEngine()
	h := NewBotHandler(e)

	c, rec := echoContext(http.MethodGet, "/api/v1/bots/nonexistent", "")
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	err := h.GetBot(c)
	if err != nil {
		t.Fatalf("GetBot handler error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_DeleteBot_NotFound(t *testing.T) {
	e := newTestEngine()
	h := NewBotHandler(e)

	c, rec := echoContext(http.MethodDelete, "/api/v1/bots/nonexistent", "")
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	err := h.DeleteBot(c)
	if err != nil {
		t.Fatalf("DeleteBot handler error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_ListBotsFilterStatus(t *testing.T) {
	e := newTestEngine()
	mustRegisterBot(t, e, sampleBot("bot-1", "Active Bot"))
	inactive := sampleBot("bot-2", "Inactive Bot")
	inactive.Status = "inactive"
	mustRegisterBot(t, e, inactive)

	h := NewBotHandler(e)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bots?status=active", nil)
	rec := httptest.NewRecorder()
	ec := echo.New()
	c := ec.NewContext(req, rec)

	err := h.ListBots(c)
	if err != nil {
		t.Fatalf("ListBots handler error: %v", err)
	}
	var result struct {
		Data  []Bot `json:"data"`
		Total int   `json:"total"`
	}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result.Total != 1 {
		t.Errorf("expected 1 active bot, got %d", result.Total)
	}
}
