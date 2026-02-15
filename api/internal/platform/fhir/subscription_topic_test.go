package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestEngine() *SubscriptionTopicEngine {
	return NewSubscriptionTopicEngine()
}

func makeEncounter(id, status, classCode string) map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Encounter",
		"id":           id,
		"status":       status,
		"class": map[string]interface{}{
			"code": classCode,
		},
	}
}

func makeDiagnosticReport(id, status string) map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "DiagnosticReport",
		"id":           id,
		"status":       status,
	}
}

func resourceJSON(r map[string]interface{}) json.RawMessage {
	data, _ := json.Marshal(r)
	return data
}

// ---------------------------------------------------------------------------
// Test RegisterTopic and Retrieve
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_RegisterAndRetrieve(t *testing.T) {
	engine := newTestEngine()

	topic := &SubscriptionTopic{
		ID:     "test-topic-1",
		URL:    "http://example.org/SubscriptionTopic/test-topic-1",
		Name:   "TestTopic",
		Title:  "Test Topic",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{
				ResourceType: "Patient",
				Interaction:  []string{"create"},
			},
		},
	}

	engine.RegisterTopic(topic)

	got := engine.GetTopic("test-topic-1")
	if got == nil {
		t.Fatal("expected to retrieve registered topic, got nil")
	}
	if got.URL != topic.URL {
		t.Errorf("expected URL %q, got %q", topic.URL, got.URL)
	}
	if got.Name != "TestTopic" {
		t.Errorf("expected Name 'TestTopic', got %q", got.Name)
	}
}

func TestSubscriptionTopic_ListTopics(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "topic-a",
		URL:    "http://example.org/SubscriptionTopic/topic-a",
		Status: "active",
	})
	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "topic-b",
		URL:    "http://example.org/SubscriptionTopic/topic-b",
		Status: "active",
	})

	topics := engine.ListTopics()
	if len(topics) != 2 {
		t.Fatalf("expected 2 topics, got %d", len(topics))
	}
}

// ---------------------------------------------------------------------------
// Test Subscribe validates against topic canFilterBy
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_SubscribeValidatesFilters(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "enc-topic",
		URL:    "http://example.org/SubscriptionTopic/enc-topic",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Encounter", Interaction: []string{"create"}},
		},
		CanFilterBy: []TopicCanFilterBy{
			{Resource: "Encounter", FilterParameter: "status", Modifier: []string{"eq", "in"}},
			{Resource: "Encounter", FilterParameter: "class"},
		},
	})

	sub := &TopicSubscription{
		ID:          "sub-1",
		TopicURL:    "http://example.org/SubscriptionTopic/enc-topic",
		Status:      "requested",
		ChannelType: "rest-hook",
		Endpoint:    "https://example.com/hook",
		Content:     "full-resource",
		FilterBy: []TopicSubscriptionFilter{
			{FilterParameter: "status", Value: "in-progress", Modifier: "eq"},
		},
	}

	err := engine.Subscribe(sub)
	if err != nil {
		t.Fatalf("expected subscribe to succeed, got: %v", err)
	}

	// Verify the subscription is active
	got := engine.GetSubscription("sub-1")
	if got == nil {
		t.Fatal("expected to retrieve subscription, got nil")
	}
	if got.Status != "active" {
		t.Errorf("expected status 'active', got %q", got.Status)
	}
}

func TestSubscriptionTopic_SubscribeRejectsInvalidFilters(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "enc-topic",
		URL:    "http://example.org/SubscriptionTopic/enc-topic",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Encounter", Interaction: []string{"create"}},
		},
		CanFilterBy: []TopicCanFilterBy{
			{Resource: "Encounter", FilterParameter: "status"},
		},
	})

	sub := &TopicSubscription{
		ID:          "sub-bad",
		TopicURL:    "http://example.org/SubscriptionTopic/enc-topic",
		Status:      "requested",
		ChannelType: "rest-hook",
		Endpoint:    "https://example.com/hook",
		Content:     "full-resource",
		FilterBy: []TopicSubscriptionFilter{
			{FilterParameter: "unknown-field", Value: "abc"},
		},
	}

	err := engine.Subscribe(sub)
	if err == nil {
		t.Fatal("expected subscribe to fail for invalid filter, got nil")
	}
	if !strings.Contains(err.Error(), "unknown-field") {
		t.Errorf("expected error to mention 'unknown-field', got: %v", err)
	}
}

func TestSubscriptionTopic_SubscribeRejectsUnknownTopic(t *testing.T) {
	engine := newTestEngine()

	sub := &TopicSubscription{
		ID:          "sub-no-topic",
		TopicURL:    "http://example.org/SubscriptionTopic/nonexistent",
		Status:      "requested",
		ChannelType: "rest-hook",
		Endpoint:    "https://example.com/hook",
		Content:     "full-resource",
	}

	err := engine.Subscribe(sub)
	if err == nil {
		t.Fatal("expected subscribe to fail for unknown topic, got nil")
	}
}

// ---------------------------------------------------------------------------
// Test Evaluate matches resource type trigger
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_EvaluateMatchesResourceType(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "patient-create",
		URL:    "http://example.org/SubscriptionTopic/patient-create",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Patient", Interaction: []string{"create"}},
		},
	})

	sub := &TopicSubscription{
		ID:          "sub-pat",
		TopicURL:    "http://example.org/SubscriptionTopic/patient-create",
		Status:      "requested",
		ChannelType: "rest-hook",
		Endpoint:    "https://example.com/hook",
		Content:     "full-resource",
	}
	if err := engine.Subscribe(sub); err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	// Matching event
	event := ResourceEvent{
		ResourceType: "Patient",
		ResourceID:   "p-1",
		Action:       "create",
		Resource:     resourceJSON(map[string]interface{}{"resourceType": "Patient", "id": "p-1"}),
	}
	bundles := engine.Evaluate(event)
	if len(bundles) != 1 {
		t.Fatalf("expected 1 notification bundle, got %d", len(bundles))
	}

	// Non-matching resource type
	event2 := ResourceEvent{
		ResourceType: "Observation",
		ResourceID:   "o-1",
		Action:       "create",
		Resource:     resourceJSON(map[string]interface{}{"resourceType": "Observation", "id": "o-1"}),
	}
	bundles2 := engine.Evaluate(event2)
	if len(bundles2) != 0 {
		t.Fatalf("expected 0 notification bundles for non-matching type, got %d", len(bundles2))
	}
}

// ---------------------------------------------------------------------------
// Test Evaluate matches interaction type
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_EvaluateMatchesInteractionCreate(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "create-only",
		URL:    "http://example.org/SubscriptionTopic/create-only",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Patient", Interaction: []string{"create"}},
		},
	})

	sub := &TopicSubscription{
		ID: "sub-co", TopicURL: "http://example.org/SubscriptionTopic/create-only",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	_ = engine.Subscribe(sub)

	// create matches
	b := engine.Evaluate(ResourceEvent{ResourceType: "Patient", ResourceID: "1", Action: "create",
		Resource: resourceJSON(map[string]interface{}{"resourceType": "Patient", "id": "1"})})
	if len(b) != 1 {
		t.Fatalf("expected 1 bundle for create, got %d", len(b))
	}

	// update does not match
	b2 := engine.Evaluate(ResourceEvent{ResourceType: "Patient", ResourceID: "1", Action: "update",
		Resource: resourceJSON(map[string]interface{}{"resourceType": "Patient", "id": "1"})})
	if len(b2) != 0 {
		t.Fatalf("expected 0 bundles for update, got %d", len(b2))
	}
}

func TestSubscriptionTopic_EvaluateMatchesInteractionUpdate(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "update-only",
		URL:    "http://example.org/SubscriptionTopic/update-only",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Patient", Interaction: []string{"update"}},
		},
	})

	sub := &TopicSubscription{
		ID: "sub-uo", TopicURL: "http://example.org/SubscriptionTopic/update-only",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	_ = engine.Subscribe(sub)

	b := engine.Evaluate(ResourceEvent{ResourceType: "Patient", ResourceID: "1", Action: "update",
		Resource: resourceJSON(map[string]interface{}{"resourceType": "Patient", "id": "1"})})
	if len(b) != 1 {
		t.Fatalf("expected 1 bundle for update, got %d", len(b))
	}

	b2 := engine.Evaluate(ResourceEvent{ResourceType: "Patient", ResourceID: "1", Action: "create",
		Resource: resourceJSON(map[string]interface{}{"resourceType": "Patient", "id": "1"})})
	if len(b2) != 0 {
		t.Fatalf("expected 0 bundles for create on update-only topic, got %d", len(b2))
	}
}

func TestSubscriptionTopic_EvaluateMatchesInteractionDelete(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "delete-only",
		URL:    "http://example.org/SubscriptionTopic/delete-only",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Patient", Interaction: []string{"delete"}},
		},
	})

	sub := &TopicSubscription{
		ID: "sub-do", TopicURL: "http://example.org/SubscriptionTopic/delete-only",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	_ = engine.Subscribe(sub)

	b := engine.Evaluate(ResourceEvent{ResourceType: "Patient", ResourceID: "1", Action: "delete",
		Resource: resourceJSON(map[string]interface{}{"resourceType": "Patient", "id": "1"})})
	if len(b) != 1 {
		t.Fatalf("expected 1 bundle for delete, got %d", len(b))
	}

	b2 := engine.Evaluate(ResourceEvent{ResourceType: "Patient", ResourceID: "1", Action: "create",
		Resource: resourceJSON(map[string]interface{}{"resourceType": "Patient", "id": "1"})})
	if len(b2) != 0 {
		t.Fatalf("expected 0 bundles for create on delete-only topic, got %d", len(b2))
	}
}

// ---------------------------------------------------------------------------
// Test Evaluate applies FHIRPath criteria
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_EvaluateAppliesFHIRPathCriteria(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "enc-inprogress",
		URL:    "http://example.org/SubscriptionTopic/enc-inprogress",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{
				ResourceType:      "Encounter",
				Interaction:       []string{"create"},
				FHIRPathCriteria: "status = 'in-progress'",
			},
		},
	})

	sub := &TopicSubscription{
		ID: "sub-fp", TopicURL: "http://example.org/SubscriptionTopic/enc-inprogress",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	_ = engine.Subscribe(sub)

	// Matching: status=in-progress
	enc := makeEncounter("e1", "in-progress", "AMB")
	b := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e1", Action: "create", Resource: resourceJSON(enc)})
	if len(b) != 1 {
		t.Fatalf("expected 1 bundle for matching FHIRPath, got %d", len(b))
	}

	// Non-matching: status=planned
	enc2 := makeEncounter("e2", "planned", "AMB")
	b2 := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e2", Action: "create", Resource: resourceJSON(enc2)})
	if len(b2) != 0 {
		t.Fatalf("expected 0 bundles for non-matching FHIRPath, got %d", len(b2))
	}
}

// ---------------------------------------------------------------------------
// Test Evaluate applies subscription filters
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_EvaluateAppliesSubscriptionFilters(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "enc-any",
		URL:    "http://example.org/SubscriptionTopic/enc-any",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Encounter", Interaction: []string{"create"}},
		},
		CanFilterBy: []TopicCanFilterBy{
			{Resource: "Encounter", FilterParameter: "status"},
		},
	})

	// Subscription filtering only status=in-progress encounters
	sub := &TopicSubscription{
		ID: "sub-filter", TopicURL: "http://example.org/SubscriptionTopic/enc-any",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
		FilterBy: []TopicSubscriptionFilter{
			{FilterParameter: "status", Value: "in-progress"},
		},
	}
	_ = engine.Subscribe(sub)

	// Matching
	enc := makeEncounter("e1", "in-progress", "AMB")
	b := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e1", Action: "create", Resource: resourceJSON(enc)})
	if len(b) != 1 {
		t.Fatalf("expected 1 bundle for matching filter, got %d", len(b))
	}

	// Non-matching
	enc2 := makeEncounter("e2", "finished", "AMB")
	b2 := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e2", Action: "create", Resource: resourceJSON(enc2)})
	if len(b2) != 0 {
		t.Fatalf("expected 0 bundles for non-matching filter, got %d", len(b2))
	}
}

// ---------------------------------------------------------------------------
// Test Evaluate returns empty for non-matching events
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_EvaluateReturnsEmptyForNonMatching(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "pat-topic",
		URL:    "http://example.org/SubscriptionTopic/pat-topic",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Patient", Interaction: []string{"create"}},
		},
	})

	sub := &TopicSubscription{
		ID: "sub-empty", TopicURL: "http://example.org/SubscriptionTopic/pat-topic",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	_ = engine.Subscribe(sub)

	// Wrong resource type
	b := engine.Evaluate(ResourceEvent{ResourceType: "Observation", ResourceID: "o1", Action: "create",
		Resource: resourceJSON(map[string]interface{}{"resourceType": "Observation", "id": "o1"})})
	if len(b) != 0 {
		t.Fatalf("expected 0 bundles, got %d", len(b))
	}

	// Wrong action
	b2 := engine.Evaluate(ResourceEvent{ResourceType: "Patient", ResourceID: "p1", Action: "delete",
		Resource: resourceJSON(map[string]interface{}{"resourceType": "Patient", "id": "p1"})})
	if len(b2) != 0 {
		t.Fatalf("expected 0 bundles, got %d", len(b2))
	}
}

// ---------------------------------------------------------------------------
// Test NotificationBundle structure
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_NotificationBundleStructure(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "pat-create",
		URL:    "http://example.org/SubscriptionTopic/pat-create",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Patient", Interaction: []string{"create"}},
		},
	})

	sub := &TopicSubscription{
		ID: "sub-bundle", TopicURL: "http://example.org/SubscriptionTopic/pat-create",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	_ = engine.Subscribe(sub)

	event := ResourceEvent{ResourceType: "Patient", ResourceID: "p1", Action: "create",
		Resource: resourceJSON(map[string]interface{}{"resourceType": "Patient", "id": "p1", "name": []interface{}{map[string]interface{}{"family": "Doe"}}})}
	bundles := engine.Evaluate(event)
	if len(bundles) != 1 {
		t.Fatalf("expected 1 bundle, got %d", len(bundles))
	}

	nb := bundles[0]
	if nb.Type != "event-notification" {
		t.Errorf("expected type 'event-notification', got %q", nb.Type)
	}
	if nb.SubscriptionID != "sub-bundle" {
		t.Errorf("expected subscription ref 'sub-bundle', got %q", nb.SubscriptionID)
	}
	if nb.TopicURL != "http://example.org/SubscriptionTopic/pat-create" {
		t.Errorf("expected topic URL, got %q", nb.TopicURL)
	}
	if nb.FocusResource == nil {
		t.Fatal("expected focus resource to be non-nil")
	}
	if nb.FocusResource["id"] != "p1" {
		t.Errorf("expected focus resource id 'p1', got %v", nb.FocusResource["id"])
	}
}

// ---------------------------------------------------------------------------
// Test content levels: empty, id-only, full-resource
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_ContentLevelEmpty(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "ct-topic",
		URL:    "http://example.org/SubscriptionTopic/ct-topic",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Patient", Interaction: []string{"create"}},
		},
	})

	sub := &TopicSubscription{
		ID: "sub-empty-content", TopicURL: "http://example.org/SubscriptionTopic/ct-topic",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "empty",
	}
	_ = engine.Subscribe(sub)

	event := ResourceEvent{ResourceType: "Patient", ResourceID: "p1", Action: "create",
		Resource: resourceJSON(map[string]interface{}{"resourceType": "Patient", "id": "p1", "name": []interface{}{map[string]interface{}{"family": "Doe"}}})}
	bundles := engine.Evaluate(event)
	if len(bundles) != 1 {
		t.Fatalf("expected 1 bundle, got %d", len(bundles))
	}
	if bundles[0].FocusResource != nil {
		t.Error("expected nil focus resource for 'empty' content level")
	}
}

func TestSubscriptionTopic_ContentLevelIDOnly(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "ct-topic",
		URL:    "http://example.org/SubscriptionTopic/ct-topic",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Patient", Interaction: []string{"create"}},
		},
	})

	sub := &TopicSubscription{
		ID: "sub-id-only", TopicURL: "http://example.org/SubscriptionTopic/ct-topic",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "id-only",
	}
	_ = engine.Subscribe(sub)

	event := ResourceEvent{ResourceType: "Patient", ResourceID: "p1", Action: "create",
		Resource: resourceJSON(map[string]interface{}{"resourceType": "Patient", "id": "p1", "name": []interface{}{map[string]interface{}{"family": "Doe"}}})}
	bundles := engine.Evaluate(event)
	if len(bundles) != 1 {
		t.Fatalf("expected 1 bundle, got %d", len(bundles))
	}
	focus := bundles[0].FocusResource
	if focus == nil {
		t.Fatal("expected non-nil focus resource for 'id-only'")
	}
	if focus["id"] != "p1" {
		t.Errorf("expected id 'p1', got %v", focus["id"])
	}
	if focus["resourceType"] != "Patient" {
		t.Errorf("expected resourceType 'Patient', got %v", focus["resourceType"])
	}
	// Should NOT contain name
	if _, ok := focus["name"]; ok {
		t.Error("expected 'id-only' to strip extra fields like 'name'")
	}
}

func TestSubscriptionTopic_ContentLevelFullResource(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "ct-topic",
		URL:    "http://example.org/SubscriptionTopic/ct-topic",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Patient", Interaction: []string{"create"}},
		},
	})

	sub := &TopicSubscription{
		ID: "sub-full", TopicURL: "http://example.org/SubscriptionTopic/ct-topic",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	_ = engine.Subscribe(sub)

	event := ResourceEvent{ResourceType: "Patient", ResourceID: "p1", Action: "create",
		Resource: resourceJSON(map[string]interface{}{"resourceType": "Patient", "id": "p1", "name": []interface{}{map[string]interface{}{"family": "Doe"}}})}
	bundles := engine.Evaluate(event)
	if len(bundles) != 1 {
		t.Fatalf("expected 1 bundle, got %d", len(bundles))
	}
	focus := bundles[0].FocusResource
	if focus == nil {
		t.Fatal("expected non-nil focus resource for 'full-resource'")
	}
	if _, ok := focus["name"]; !ok {
		t.Error("expected 'full-resource' to include all fields like 'name'")
	}
}

// ---------------------------------------------------------------------------
// Test built-in encounter-start topic
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_BuiltInEncounterStart(t *testing.T) {
	engine := newTestEngine()
	engine.RegisterBuiltInTopics()

	sub := &TopicSubscription{
		ID: "sub-enc-start", TopicURL: "http://ehr.example.org/SubscriptionTopic/encounter-start",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	if err := engine.Subscribe(sub); err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	// Matching: Encounter create with status=in-progress
	enc := makeEncounter("e1", "in-progress", "AMB")
	b := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e1", Action: "create", Resource: resourceJSON(enc)})
	if len(b) != 1 {
		t.Fatalf("expected 1 bundle for encounter-start, got %d", len(b))
	}

	// Non-matching: status=planned
	enc2 := makeEncounter("e2", "planned", "AMB")
	b2 := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e2", Action: "create", Resource: resourceJSON(enc2)})
	if len(b2) != 0 {
		t.Fatalf("expected 0 bundles for planned encounter, got %d", len(b2))
	}

	// Non-matching: update (encounter-start is create only)
	enc3 := makeEncounter("e3", "in-progress", "AMB")
	b3 := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e3", Action: "update", Resource: resourceJSON(enc3)})
	if len(b3) != 0 {
		t.Fatalf("expected 0 bundles for update on encounter-start, got %d", len(b3))
	}
}

// ---------------------------------------------------------------------------
// Test built-in encounter-end topic
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_BuiltInEncounterEnd(t *testing.T) {
	engine := newTestEngine()
	engine.RegisterBuiltInTopics()

	sub := &TopicSubscription{
		ID: "sub-enc-end", TopicURL: "http://ehr.example.org/SubscriptionTopic/encounter-end",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	if err := engine.Subscribe(sub); err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	// Matching: Encounter update with status=finished
	enc := makeEncounter("e1", "finished", "AMB")
	b := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e1", Action: "update", Resource: resourceJSON(enc)})
	if len(b) != 1 {
		t.Fatalf("expected 1 bundle for encounter-end, got %d", len(b))
	}

	// Non-matching: status=in-progress
	enc2 := makeEncounter("e2", "in-progress", "AMB")
	b2 := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e2", Action: "update", Resource: resourceJSON(enc2)})
	if len(b2) != 0 {
		t.Fatalf("expected 0 bundles for in-progress encounter on encounter-end, got %d", len(b2))
	}

	// Non-matching: create (encounter-end is update only)
	enc3 := makeEncounter("e3", "finished", "AMB")
	b3 := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e3", Action: "create", Resource: resourceJSON(enc3)})
	if len(b3) != 0 {
		t.Fatalf("expected 0 bundles for create on encounter-end, got %d", len(b3))
	}
}

// ---------------------------------------------------------------------------
// Test built-in new-lab-result topic
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_BuiltInNewLabResult(t *testing.T) {
	engine := newTestEngine()
	engine.RegisterBuiltInTopics()

	sub := &TopicSubscription{
		ID: "sub-lab", TopicURL: "http://ehr.example.org/SubscriptionTopic/new-lab-result",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	if err := engine.Subscribe(sub); err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	// Matching: DiagnosticReport create with status=final
	dr := makeDiagnosticReport("dr1", "final")
	b := engine.Evaluate(ResourceEvent{ResourceType: "DiagnosticReport", ResourceID: "dr1", Action: "create", Resource: resourceJSON(dr)})
	if len(b) != 1 {
		t.Fatalf("expected 1 bundle for new-lab-result, got %d", len(b))
	}

	// Non-matching: status=preliminary
	dr2 := makeDiagnosticReport("dr2", "preliminary")
	b2 := engine.Evaluate(ResourceEvent{ResourceType: "DiagnosticReport", ResourceID: "dr2", Action: "create", Resource: resourceJSON(dr2)})
	if len(b2) != 0 {
		t.Fatalf("expected 0 bundles for preliminary report, got %d", len(b2))
	}
}

// ---------------------------------------------------------------------------
// Test built-in admission-discharge topic
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_BuiltInAdmissionDischarge(t *testing.T) {
	engine := newTestEngine()
	engine.RegisterBuiltInTopics()

	sub := &TopicSubscription{
		ID: "sub-ad", TopicURL: "http://ehr.example.org/SubscriptionTopic/admission-discharge",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	if err := engine.Subscribe(sub); err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	// Matching: Encounter create, class=IMP, status=in-progress
	enc := makeEncounter("e1", "in-progress", "IMP")
	b := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e1", Action: "create", Resource: resourceJSON(enc)})
	if len(b) != 1 {
		t.Fatalf("expected 1 bundle for admission, got %d", len(b))
	}

	// Matching: Encounter update, class=IMP, status=finished
	enc2 := makeEncounter("e2", "finished", "IMP")
	b2 := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e2", Action: "update", Resource: resourceJSON(enc2)})
	if len(b2) != 1 {
		t.Fatalf("expected 1 bundle for discharge, got %d", len(b2))
	}

	// Non-matching: class=AMB (outpatient, not inpatient)
	enc3 := makeEncounter("e3", "in-progress", "AMB")
	b3 := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e3", Action: "create", Resource: resourceJSON(enc3)})
	if len(b3) != 0 {
		t.Fatalf("expected 0 bundles for outpatient encounter, got %d", len(b3))
	}

	// Non-matching: class=IMP but status=planned
	enc4 := makeEncounter("e4", "planned", "IMP")
	b4 := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e4", Action: "create", Resource: resourceJSON(enc4)})
	if len(b4) != 0 {
		t.Fatalf("expected 0 bundles for planned inpatient encounter, got %d", len(b4))
	}
}

// ---------------------------------------------------------------------------
// Test multiple subscriptions on same topic
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_MultipleSubscriptionsSameTopic(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "multi-topic",
		URL:    "http://example.org/SubscriptionTopic/multi-topic",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Patient", Interaction: []string{"create"}},
		},
	})

	for _, id := range []string{"sub-a", "sub-b", "sub-c"} {
		sub := &TopicSubscription{
			ID: id, TopicURL: "http://example.org/SubscriptionTopic/multi-topic",
			Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook-" + id, Content: "full-resource",
		}
		if err := engine.Subscribe(sub); err != nil {
			t.Fatalf("subscribe %s failed: %v", id, err)
		}
	}

	event := ResourceEvent{ResourceType: "Patient", ResourceID: "p1", Action: "create",
		Resource: resourceJSON(map[string]interface{}{"resourceType": "Patient", "id": "p1"})}
	bundles := engine.Evaluate(event)
	if len(bundles) != 3 {
		t.Fatalf("expected 3 bundles for 3 subscriptions, got %d", len(bundles))
	}
}

// ---------------------------------------------------------------------------
// Test subscription expiration (end time)
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_SubscriptionExpiration(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "exp-topic",
		URL:    "http://example.org/SubscriptionTopic/exp-topic",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Patient", Interaction: []string{"create"}},
		},
	})

	pastTime := time.Now().Add(-1 * time.Hour)
	sub := &TopicSubscription{
		ID: "sub-expired", TopicURL: "http://example.org/SubscriptionTopic/exp-topic",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
		End: &pastTime,
	}
	if err := engine.Subscribe(sub); err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	event := ResourceEvent{ResourceType: "Patient", ResourceID: "p1", Action: "create",
		Resource: resourceJSON(map[string]interface{}{"resourceType": "Patient", "id": "p1"})}
	bundles := engine.Evaluate(event)
	if len(bundles) != 0 {
		t.Fatalf("expected 0 bundles for expired subscription, got %d", len(bundles))
	}

	// Verify subscription status changed to off
	got := engine.GetSubscription("sub-expired")
	if got == nil {
		t.Fatal("expected to find subscription")
	}
	if got.Status != "off" {
		t.Errorf("expected expired subscription status 'off', got %q", got.Status)
	}
}

// ---------------------------------------------------------------------------
// Test subscription status tracking (event count)
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_SubscriptionStatusTracking(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "status-topic",
		URL:    "http://example.org/SubscriptionTopic/status-topic",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Patient", Interaction: []string{"create"}},
		},
	})

	sub := &TopicSubscription{
		ID: "sub-status", TopicURL: "http://example.org/SubscriptionTopic/status-topic",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	_ = engine.Subscribe(sub)

	// Send 3 events
	for i := 0; i < 3; i++ {
		event := ResourceEvent{ResourceType: "Patient", ResourceID: "p" + string(rune('1'+i)), Action: "create",
			Resource: resourceJSON(map[string]interface{}{"resourceType": "Patient", "id": "p" + string(rune('1'+i))})}
		engine.Evaluate(event)
	}

	status := engine.GetSubscriptionStatus("sub-status")
	if status == nil {
		t.Fatal("expected subscription status")
	}
	if status.EventCount != 3 {
		t.Errorf("expected event count 3, got %d", status.EventCount)
	}
	if status.Status != "active" {
		t.Errorf("expected status 'active', got %q", status.Status)
	}
}

// ---------------------------------------------------------------------------
// Test handler GET /SubscriptionTopic list
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_HandlerListTopics(t *testing.T) {
	engine := newTestEngine()
	engine.RegisterBuiltInTopics()
	handler := NewTopicHandler(engine)

	e := echo.New()
	handler.RegisterRoutes(e.Group(""))

	req := httptest.NewRequest(http.MethodGet, "/SubscriptionTopic", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected resourceType 'Bundle', got %v", bundle["resourceType"])
	}
	entries, ok := bundle["entry"].([]interface{})
	if !ok {
		t.Fatal("expected entries array")
	}
	if len(entries) < 4 {
		t.Errorf("expected at least 4 built-in topics, got %d", len(entries))
	}
}

func TestSubscriptionTopic_HandlerGetTopic(t *testing.T) {
	engine := newTestEngine()
	engine.RegisterBuiltInTopics()
	handler := NewTopicHandler(engine)

	e := echo.New()
	handler.RegisterRoutes(e.Group(""))

	req := httptest.NewRequest(http.MethodGet, "/SubscriptionTopic/encounter-start", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var topic map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &topic); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if topic["resourceType"] != "SubscriptionTopic" {
		t.Errorf("expected resourceType 'SubscriptionTopic', got %v", topic["resourceType"])
	}
	if topic["id"] != "encounter-start" {
		t.Errorf("expected id 'encounter-start', got %v", topic["id"])
	}
}

// ---------------------------------------------------------------------------
// Test handler POST /Subscription with topic
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_HandlerCreateSubscription(t *testing.T) {
	engine := newTestEngine()
	engine.RegisterBuiltInTopics()
	handler := NewTopicHandler(engine)

	e := echo.New()
	handler.RegisterRoutes(e.Group(""))

	body := `{
		"resourceType": "Subscription",
		"topic": "http://ehr.example.org/SubscriptionTopic/encounter-start",
		"status": "requested",
		"channelType": "rest-hook",
		"endpoint": "https://example.com/hook",
		"content": "full-resource"
	}`

	req := httptest.NewRequest(http.MethodPost, "/Subscription", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["status"] != "active" {
		t.Errorf("expected status 'active', got %v", result["status"])
	}
}

// ---------------------------------------------------------------------------
// Test handler GET /Subscription/:id/$status
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_HandlerGetStatus(t *testing.T) {
	engine := newTestEngine()
	engine.RegisterBuiltInTopics()
	handler := NewTopicHandler(engine)

	sub := &TopicSubscription{
		ID: "sub-st", TopicURL: "http://ehr.example.org/SubscriptionTopic/encounter-start",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	_ = engine.Subscribe(sub)

	// Generate an event to bump event count
	enc := makeEncounter("e1", "in-progress", "AMB")
	engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e1", Action: "create", Resource: resourceJSON(enc)})

	e := echo.New()
	handler.RegisterRoutes(e.Group(""))

	req := httptest.NewRequest(http.MethodGet, "/Subscription/sub-st/$status", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["status"] != "active" {
		t.Errorf("expected status 'active', got %v", result["status"])
	}
	// Should have eventCount
	if result["type"] != "query-status" {
		t.Errorf("expected type 'query-status', got %v", result["type"])
	}
}

// ---------------------------------------------------------------------------
// Test handler GET /Subscription/:id/$events
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_HandlerGetEvents(t *testing.T) {
	engine := newTestEngine()
	engine.RegisterBuiltInTopics()
	handler := NewTopicHandler(engine)

	sub := &TopicSubscription{
		ID: "sub-ev", TopicURL: "http://ehr.example.org/SubscriptionTopic/encounter-start",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	_ = engine.Subscribe(sub)

	// Generate events
	enc := makeEncounter("e1", "in-progress", "AMB")
	engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e1", Action: "create", Resource: resourceJSON(enc)})

	e := echo.New()
	handler.RegisterRoutes(e.Group(""))

	req := httptest.NewRequest(http.MethodGet, "/Subscription/sub-ev/$events", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["resourceType"] != "Bundle" {
		t.Errorf("expected resourceType 'Bundle', got %v", result["resourceType"])
	}
	if result["type"] != "history" {
		t.Errorf("expected type 'history', got %v", result["type"])
	}
}

// ---------------------------------------------------------------------------
// Test handler POST /Subscription/:id/$events (replay)
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_HandlerReplayEvents(t *testing.T) {
	engine := newTestEngine()
	engine.RegisterBuiltInTopics()
	handler := NewTopicHandler(engine)

	sub := &TopicSubscription{
		ID: "sub-replay", TopicURL: "http://ehr.example.org/SubscriptionTopic/encounter-start",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	_ = engine.Subscribe(sub)

	// Generate events
	enc := makeEncounter("e1", "in-progress", "AMB")
	engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e1", Action: "create", Resource: resourceJSON(enc)})

	e := echo.New()
	handler.RegisterRoutes(e.Group(""))

	req := httptest.NewRequest(http.MethodPost, "/Subscription/sub-replay/$events", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["resourceType"] != "Bundle" {
		t.Errorf("expected resourceType 'Bundle', got %v", result["resourceType"])
	}
}

// ---------------------------------------------------------------------------
// Test concurrent event evaluation (race safety)
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_ConcurrentEvaluation(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "race-topic",
		URL:    "http://example.org/SubscriptionTopic/race-topic",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Patient", Interaction: []string{"create"}},
		},
	})

	sub := &TopicSubscription{
		ID: "sub-race", TopicURL: "http://example.org/SubscriptionTopic/race-topic",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	_ = engine.Subscribe(sub)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			event := ResourceEvent{
				ResourceType: "Patient",
				ResourceID:   "p-concurrent",
				Action:       "create",
				Resource:     resourceJSON(map[string]interface{}{"resourceType": "Patient", "id": "p-concurrent"}),
			}
			bundles := engine.Evaluate(event)
			if len(bundles) != 1 {
				t.Errorf("goroutine %d: expected 1 bundle, got %d", n, len(bundles))
			}
		}(i)
	}
	wg.Wait()

	// Also test concurrent subscribe + evaluate
	var wg2 sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg2.Add(2)
		go func(n int) {
			defer wg2.Done()
			event := ResourceEvent{
				ResourceType: "Patient",
				ResourceID:   "p-race2",
				Action:       "create",
				Resource:     resourceJSON(map[string]interface{}{"resourceType": "Patient", "id": "p-race2"}),
			}
			engine.Evaluate(event)
		}(i)
		go func(n int) {
			defer wg2.Done()
			_ = engine.Subscribe(&TopicSubscription{
				ID: "sub-race-dyn", TopicURL: "http://example.org/SubscriptionTopic/race-topic",
				Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook-dyn", Content: "full-resource",
			})
		}(i)
	}
	wg2.Wait()
}

// ---------------------------------------------------------------------------
// Test handler POST /SubscriptionTopic (create custom topic)
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_HandlerCreateCustomTopic(t *testing.T) {
	engine := newTestEngine()
	handler := NewTopicHandler(engine)

	e := echo.New()
	handler.RegisterRoutes(e.Group(""))

	body := `{
		"resourceType": "SubscriptionTopic",
		"id": "custom-topic",
		"url": "http://example.org/SubscriptionTopic/custom-topic",
		"status": "active",
		"title": "Custom Topic",
		"resourceTrigger": [
			{
				"resourceType": "Observation",
				"interaction": ["create"]
			}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/SubscriptionTopic", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify it's registered
	got := engine.GetTopic("custom-topic")
	if got == nil {
		t.Fatal("expected custom topic to be registered")
	}
}

// ---------------------------------------------------------------------------
// Test handler GET /SubscriptionTopic/:id not found
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_HandlerGetTopicNotFound(t *testing.T) {
	engine := newTestEngine()
	handler := NewTopicHandler(engine)

	e := echo.New()
	handler.RegisterRoutes(e.Group(""))

	req := httptest.NewRequest(http.MethodGet, "/SubscriptionTopic/nonexistent", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Test event triggers
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_EventTrigger(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "event-topic",
		URL:    "http://example.org/SubscriptionTopic/event-topic",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Encounter", Interaction: []string{"create"}},
		},
		EventTrigger: []TopicEventTrigger{
			{Description: "ADT admission", Event: Coding{System: "http://hl7.org/fhir/event", Code: "admit"}},
		},
	})

	got := engine.GetTopic("event-topic")
	if got == nil {
		t.Fatal("expected topic")
	}
	if len(got.EventTrigger) != 1 {
		t.Errorf("expected 1 event trigger, got %d", len(got.EventTrigger))
	}
	if got.EventTrigger[0].Event.Code != "admit" {
		t.Errorf("expected event code 'admit', got %q", got.EventTrigger[0].Event.Code)
	}
}

// ---------------------------------------------------------------------------
// Test notification shape
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_NotificationShape(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "shape-topic",
		URL:    "http://example.org/SubscriptionTopic/shape-topic",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Patient", Interaction: []string{"create"}},
		},
		NotificationShape: []TopicNotificationShape{
			{Resource: "Patient", Include: []string{"Patient:generalPractitioner"}},
		},
	})

	got := engine.GetTopic("shape-topic")
	if got == nil {
		t.Fatal("expected topic")
	}
	if len(got.NotificationShape) != 1 {
		t.Errorf("expected 1 notification shape, got %d", len(got.NotificationShape))
	}
	if got.NotificationShape[0].Resource != "Patient" {
		t.Errorf("expected shape resource 'Patient', got %q", got.NotificationShape[0].Resource)
	}
}

// ---------------------------------------------------------------------------
// Test filter with 'in' modifier
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_FilterInModifier(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "in-topic",
		URL:    "http://example.org/SubscriptionTopic/in-topic",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Encounter", Interaction: []string{"create", "update"}},
		},
		CanFilterBy: []TopicCanFilterBy{
			{Resource: "Encounter", FilterParameter: "status", Modifier: []string{"eq", "in"}},
		},
	})

	sub := &TopicSubscription{
		ID: "sub-in", TopicURL: "http://example.org/SubscriptionTopic/in-topic",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
		FilterBy: []TopicSubscriptionFilter{
			{FilterParameter: "status", Value: "in-progress,finished", Modifier: "in"},
		},
	}
	_ = engine.Subscribe(sub)

	// Matching: status=in-progress
	enc := makeEncounter("e1", "in-progress", "AMB")
	b := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e1", Action: "create", Resource: resourceJSON(enc)})
	if len(b) != 1 {
		t.Fatalf("expected 1 bundle for in-progress with 'in' modifier, got %d", len(b))
	}

	// Matching: status=finished
	enc2 := makeEncounter("e2", "finished", "AMB")
	b2 := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e2", Action: "update", Resource: resourceJSON(enc2)})
	if len(b2) != 1 {
		t.Fatalf("expected 1 bundle for finished with 'in' modifier, got %d", len(b2))
	}

	// Non-matching: status=planned
	enc3 := makeEncounter("e3", "planned", "AMB")
	b3 := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e3", Action: "create", Resource: resourceJSON(enc3)})
	if len(b3) != 0 {
		t.Fatalf("expected 0 bundles for planned with 'in' modifier, got %d", len(b3))
	}
}

// ---------------------------------------------------------------------------
// Test subscription with websocket and email channel types
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_ChannelTypes(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "ch-topic",
		URL:    "http://example.org/SubscriptionTopic/ch-topic",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{ResourceType: "Patient", Interaction: []string{"create"}},
		},
	})

	for _, chType := range []string{"rest-hook", "websocket", "email"} {
		sub := &TopicSubscription{
			ID: "sub-" + chType, TopicURL: "http://example.org/SubscriptionTopic/ch-topic",
			Status: "requested", ChannelType: chType, Endpoint: "https://example.com/hook", Content: "full-resource",
		}
		if err := engine.Subscribe(sub); err != nil {
			t.Errorf("expected channel type %q to be valid, got error: %v", chType, err)
		}
	}

	// Invalid channel type
	sub := &TopicSubscription{
		ID: "sub-invalid", TopicURL: "http://example.org/SubscriptionTopic/ch-topic",
		Status: "requested", ChannelType: "ftp", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	if err := engine.Subscribe(sub); err == nil {
		t.Error("expected error for invalid channel type 'ftp'")
	}
}

// ---------------------------------------------------------------------------
// Test query criteria matching
// ---------------------------------------------------------------------------

func TestSubscriptionTopic_QueryCriteria(t *testing.T) {
	engine := newTestEngine()

	engine.RegisterTopic(&SubscriptionTopic{
		ID:     "qc-topic",
		URL:    "http://example.org/SubscriptionTopic/qc-topic",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{
				ResourceType: "Encounter",
				Interaction:  []string{"create"},
				QueryCriteria: &TopicQueryCriteria{
					Current: "status=in-progress",
				},
			},
		},
	})

	sub := &TopicSubscription{
		ID: "sub-qc", TopicURL: "http://example.org/SubscriptionTopic/qc-topic",
		Status: "requested", ChannelType: "rest-hook", Endpoint: "https://example.com/hook", Content: "full-resource",
	}
	_ = engine.Subscribe(sub)

	// Matching
	enc := makeEncounter("e1", "in-progress", "AMB")
	b := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e1", Action: "create", Resource: resourceJSON(enc)})
	if len(b) != 1 {
		t.Fatalf("expected 1 bundle for query criteria match, got %d", len(b))
	}

	// Non-matching
	enc2 := makeEncounter("e2", "planned", "AMB")
	b2 := engine.Evaluate(ResourceEvent{ResourceType: "Encounter", ResourceID: "e2", Action: "create", Resource: resourceJSON(enc2)})
	if len(b2) != 0 {
		t.Fatalf("expected 0 bundles for query criteria non-match, got %d", len(b2))
	}
}
