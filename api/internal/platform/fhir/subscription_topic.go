package fhir

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// ============================================================================
// Models
// ============================================================================

// SubscriptionTopic represents an R5-style SubscriptionTopic resource,
// backported to R4 via the Subscriptions R5 Backport IG.
type SubscriptionTopic struct {
	ID                string                   `json:"id"`
	URL               string                   `json:"url"`
	Version           string                   `json:"version,omitempty"`
	Name              string                   `json:"name,omitempty"`
	Title             string                   `json:"title,omitempty"`
	Status            string                   `json:"status"` // draft | active | retired
	ResourceTrigger   []TopicResourceTrigger   `json:"resourceTrigger,omitempty"`
	EventTrigger      []TopicEventTrigger      `json:"eventTrigger,omitempty"`
	CanFilterBy       []TopicCanFilterBy       `json:"canFilterBy,omitempty"`
	NotificationShape []TopicNotificationShape `json:"notificationShape,omitempty"`
}

// TopicResourceTrigger defines what resource changes trigger the topic.
type TopicResourceTrigger struct {
	ResourceType     string              `json:"resourceType"`
	Interaction      []string            `json:"interaction,omitempty"`      // create, update, delete
	FHIRPathCriteria string              `json:"fhirPathCriteria,omitempty"` // FHIRPath boolean expression
	QueryCriteria    *TopicQueryCriteria `json:"queryCriteria,omitempty"`
}

// TopicQueryCriteria matches field=value pairs against the current (and optionally previous) resource.
type TopicQueryCriteria struct {
	Previous string `json:"previous,omitempty"` // e.g. "status=planned"
	Current  string `json:"current,omitempty"`  // e.g. "status=in-progress"
}

// TopicEventTrigger describes an event-based trigger.
type TopicEventTrigger struct {
	Description string `json:"description,omitempty"`
	Event       Coding `json:"event"`
}

// TopicCanFilterBy describes a filter parameter a subscription may use.
type TopicCanFilterBy struct {
	Resource        string   `json:"resource,omitempty"`
	FilterParameter string   `json:"filterParameter"`
	Modifier        []string `json:"modifier,omitempty"` // eq, in, not-in, etc.
}

// TopicNotificationShape describes what is included in a notification.
type TopicNotificationShape struct {
	Resource string   `json:"resource"`
	Include  []string `json:"include,omitempty"`
}

// TopicSubscription is an R5-style Subscription that references a topic.
type TopicSubscription struct {
	ID              string                    `json:"id"`
	TopicURL        string                    `json:"topic"`
	Status          string                    `json:"status"` // requested, active, error, off
	ChannelType     string                    `json:"channelType"`
	Endpoint        string                    `json:"endpoint"`
	Header          []string                  `json:"header,omitempty"`
	HeartbeatPeriod int                       `json:"heartbeatPeriod,omitempty"` // seconds
	Content         string                    `json:"content"`                  // empty, id-only, full-resource
	Timeout         int                       `json:"timeout,omitempty"`        // seconds
	MaxCount        int                       `json:"maxCount,omitempty"`
	End             *time.Time                `json:"end,omitempty"`
	FilterBy        []TopicSubscriptionFilter `json:"filterBy,omitempty"`
}

// TopicSubscriptionFilter is a runtime filter applied by a subscription.
type TopicSubscriptionFilter struct {
	FilterParameter string `json:"filterParameter"`
	Value           string `json:"value"`
	Modifier        string `json:"modifier,omitempty"` // eq, in, not-in, etc.
}

// NotificationBundle is the output of evaluating a ResourceEvent against subscriptions.
type NotificationBundle struct {
	Type             string                 `json:"type"` // event-notification, heartbeat, handshake
	SubscriptionID   string                 `json:"subscriptionId"`
	TopicURL         string                 `json:"topicUrl"`
	FocusResource    map[string]interface{} `json:"focusResource,omitempty"`
	IncludedResource []map[string]interface{} `json:"includedResource,omitempty"`
}

// SubscriptionStatus holds runtime status information for a subscription.
type SubscriptionStatus struct {
	SubscriptionID string `json:"subscriptionId"`
	TopicURL       string `json:"topicUrl"`
	Status         string `json:"status"`
	EventCount     int    `json:"eventCount"`
}

// topicEvent records an event that was produced for a subscription.
type topicEvent struct {
	Timestamp    time.Time              `json:"timestamp"`
	ResourceType string                 `json:"resourceType"`
	ResourceID   string                 `json:"resourceId"`
	Action       string                 `json:"action"`
	Resource     map[string]interface{} `json:"resource,omitempty"`
}

// ============================================================================
// Engine
// ============================================================================

var validChannelTypesR5 = map[string]bool{
	"rest-hook":  true,
	"websocket":  true,
	"email":      true,
}

var validContentLevels = map[string]bool{
	"empty":         true,
	"id-only":       true,
	"full-resource": true,
}

// SubscriptionTopicEngine manages R5-style subscription topics and evaluates
// resource events against them to produce notification bundles.
type SubscriptionTopicEngine struct {
	mu            sync.RWMutex
	topics        map[string]*SubscriptionTopic   // keyed by ID
	topicsByURL   map[string]*SubscriptionTopic   // keyed by canonical URL
	subscriptions map[string]*TopicSubscription   // keyed by ID
	eventCounts   map[string]int                  // subscription ID -> count
	eventLog      map[string][]topicEvent         // subscription ID -> recent events
	fhirPath      *FHIRPathEngine
}

// NewSubscriptionTopicEngine creates a new engine.
func NewSubscriptionTopicEngine() *SubscriptionTopicEngine {
	return &SubscriptionTopicEngine{
		topics:        make(map[string]*SubscriptionTopic),
		topicsByURL:   make(map[string]*SubscriptionTopic),
		subscriptions: make(map[string]*TopicSubscription),
		eventCounts:   make(map[string]int),
		eventLog:      make(map[string][]topicEvent),
		fhirPath:      NewFHIRPathEngine(),
	}
}

// RegisterTopic registers a SubscriptionTopic with the engine.
func (e *SubscriptionTopicEngine) RegisterTopic(topic *SubscriptionTopic) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.topics[topic.ID] = topic
	if topic.URL != "" {
		e.topicsByURL[topic.URL] = topic
	}
}

// GetTopic returns a topic by ID, or nil if not found.
func (e *SubscriptionTopicEngine) GetTopic(id string) *SubscriptionTopic {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.topics[id]
}

// ListTopics returns all registered topics.
func (e *SubscriptionTopicEngine) ListTopics() []*SubscriptionTopic {
	e.mu.RLock()
	defer e.mu.RUnlock()
	result := make([]*SubscriptionTopic, 0, len(e.topics))
	for _, t := range e.topics {
		result = append(result, t)
	}
	return result
}

// Subscribe registers a topic-based subscription. It validates the subscription
// against the referenced topic's canFilterBy. On success the subscription
// status is set to "active".
func (e *SubscriptionTopicEngine) Subscribe(sub *TopicSubscription) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Look up the topic
	topic, ok := e.topicsByURL[sub.TopicURL]
	if !ok {
		return fmt.Errorf("unknown subscription topic: %s", sub.TopicURL)
	}

	// Validate channel type
	if !validChannelTypesR5[sub.ChannelType] {
		return fmt.Errorf("unsupported channel type: %s (supported: rest-hook, websocket, email)", sub.ChannelType)
	}

	// Validate content level
	if sub.Content != "" && !validContentLevels[sub.Content] {
		return fmt.Errorf("unsupported content level: %s (supported: empty, id-only, full-resource)", sub.Content)
	}

	// Validate filters against topic's canFilterBy
	if err := validateFilters(sub.FilterBy, topic.CanFilterBy); err != nil {
		return err
	}

	// Activate
	sub.Status = "active"
	e.subscriptions[sub.ID] = sub
	e.eventCounts[sub.ID] = 0
	if e.eventLog[sub.ID] == nil {
		e.eventLog[sub.ID] = []topicEvent{}
	}

	return nil
}

// validateFilters ensures every subscription filter is permitted by the topic.
func validateFilters(filters []TopicSubscriptionFilter, canFilterBy []TopicCanFilterBy) error {
	allowed := make(map[string]*TopicCanFilterBy, len(canFilterBy))
	for i := range canFilterBy {
		allowed[canFilterBy[i].FilterParameter] = &canFilterBy[i]
	}
	for _, f := range filters {
		def, ok := allowed[f.FilterParameter]
		if !ok {
			return fmt.Errorf("filter parameter %q is not allowed by the topic (allowed: %v)",
				f.FilterParameter, allowedParamNames(canFilterBy))
		}
		// If the topic defines specific modifiers, check the subscription uses one of them
		if f.Modifier != "" && len(def.Modifier) > 0 {
			found := false
			for _, m := range def.Modifier {
				if m == f.Modifier {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("modifier %q not allowed for filter parameter %q (allowed: %v)",
					f.Modifier, f.FilterParameter, def.Modifier)
			}
		}
	}
	return nil
}

func allowedParamNames(canFilterBy []TopicCanFilterBy) []string {
	names := make([]string, len(canFilterBy))
	for i, c := range canFilterBy {
		names[i] = c.FilterParameter
	}
	return names
}

// GetSubscription returns a subscription by ID, or nil if not found.
func (e *SubscriptionTopicEngine) GetSubscription(id string) *TopicSubscription {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.subscriptions[id]
}

// GetSubscriptionStatus returns runtime status for a subscription.
func (e *SubscriptionTopicEngine) GetSubscriptionStatus(id string) *SubscriptionStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	sub, ok := e.subscriptions[id]
	if !ok {
		return nil
	}
	return &SubscriptionStatus{
		SubscriptionID: sub.ID,
		TopicURL:       sub.TopicURL,
		Status:         sub.Status,
		EventCount:     e.eventCounts[sub.ID],
	}
}

// GetSubscriptionEvents returns recent events for a subscription.
func (e *SubscriptionTopicEngine) GetSubscriptionEvents(id string) []topicEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	events := e.eventLog[id]
	if events == nil {
		return []topicEvent{}
	}
	// Return a copy
	result := make([]topicEvent, len(events))
	copy(result, events)
	return result
}

// Evaluate evaluates a ResourceEvent against all active subscriptions and
// returns a slice of NotificationBundles for matching subscriptions.
func (e *SubscriptionTopicEngine) Evaluate(event ResourceEvent) []*NotificationBundle {
	// Parse the resource once
	var resource map[string]interface{}
	if len(event.Resource) > 0 {
		if err := json.Unmarshal(event.Resource, &resource); err != nil {
			return nil
		}
	}

	e.mu.RLock()
	// Snapshot subscriptions and topics to avoid holding the lock during FHIRPath evaluation
	type subInfo struct {
		sub   *TopicSubscription
		topic *SubscriptionTopic
	}
	var candidates []subInfo
	for _, sub := range e.subscriptions {
		if sub.Status != "active" {
			continue
		}
		topic := e.topicsByURL[sub.TopicURL]
		if topic == nil || topic.Status != "active" {
			continue
		}
		candidates = append(candidates, subInfo{sub: sub, topic: topic})
	}
	e.mu.RUnlock()

	var results []*NotificationBundle

	for _, c := range candidates {
		sub := c.sub
		topic := c.topic

		// Check expiration
		if sub.End != nil && time.Now().After(*sub.End) {
			e.mu.Lock()
			sub.Status = "off"
			e.mu.Unlock()
			continue
		}

		if !e.matchesTopic(topic, event, resource) {
			continue
		}

		if !e.matchesSubscriptionFilters(sub.FilterBy, resource) {
			continue
		}

		nb := e.buildNotificationBundle(sub, topic, event, resource)
		results = append(results, nb)

		// Track event count and log
		e.mu.Lock()
		e.eventCounts[sub.ID]++
		e.eventLog[sub.ID] = append(e.eventLog[sub.ID], topicEvent{
			Timestamp:    time.Now(),
			ResourceType: event.ResourceType,
			ResourceID:   event.ResourceID,
			Action:       event.Action,
			Resource:     resource,
		})
		// Keep event log bounded
		if len(e.eventLog[sub.ID]) > 1000 {
			e.eventLog[sub.ID] = e.eventLog[sub.ID][len(e.eventLog[sub.ID])-500:]
		}
		e.mu.Unlock()
	}

	return results
}

// matchesTopic checks whether a resource event matches a topic's triggers.
func (e *SubscriptionTopicEngine) matchesTopic(topic *SubscriptionTopic, event ResourceEvent, resource map[string]interface{}) bool {
	for _, trigger := range topic.ResourceTrigger {
		if trigger.ResourceType != event.ResourceType {
			continue
		}

		// Check interaction
		if len(trigger.Interaction) > 0 {
			matched := false
			for _, interaction := range trigger.Interaction {
				if interaction == event.Action {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// Evaluate FHIRPath criteria
		if trigger.FHIRPathCriteria != "" && resource != nil {
			result, err := e.fhirPath.EvaluateBool(resource, trigger.FHIRPathCriteria)
			if err != nil || !result {
				continue
			}
		}

		// Evaluate query criteria
		if trigger.QueryCriteria != nil {
			if !e.matchesQueryCriteria(trigger.QueryCriteria, resource) {
				continue
			}
		}

		return true
	}
	return false
}

// matchesQueryCriteria checks field=value pairs from query criteria against the resource.
func (e *SubscriptionTopicEngine) matchesQueryCriteria(qc *TopicQueryCriteria, resource map[string]interface{}) bool {
	if qc.Current != "" && resource != nil {
		params := parseQueryParams(qc.Current)
		for key, expected := range params {
			actual := extractFieldValue(resource, key)
			if actual != expected {
				return false
			}
		}
	}
	// previous criteria would require the previous version of the resource,
	// which is not available in the current event model — skip for now
	return true
}

// parseQueryParams splits "key1=val1&key2=val2" into a map.
func parseQueryParams(qs string) map[string]string {
	result := make(map[string]string)
	for _, pair := range strings.Split(qs, "&") {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return result
}

// extractFieldValue reads a dotted field path from a resource map.
func extractFieldValue(resource map[string]interface{}, key string) string {
	parts := strings.Split(key, ".")
	var current interface{} = resource
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		default:
			return ""
		}
	}
	switch v := current.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return ""
	}
}

// matchesSubscriptionFilters applies subscription-level filters to the resource.
func (e *SubscriptionTopicEngine) matchesSubscriptionFilters(filters []TopicSubscriptionFilter, resource map[string]interface{}) bool {
	if len(filters) == 0 || resource == nil {
		return true
	}
	for _, f := range filters {
		actual := extractFieldValue(resource, f.FilterParameter)
		switch f.Modifier {
		case "in":
			// value is comma-separated list
			values := strings.Split(f.Value, ",")
			matched := false
			for _, v := range values {
				if strings.TrimSpace(v) == actual {
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		case "not-in":
			values := strings.Split(f.Value, ",")
			for _, v := range values {
				if strings.TrimSpace(v) == actual {
					return false
				}
			}
		default: // eq or empty modifier = equality
			if actual != f.Value {
				return false
			}
		}
	}
	return true
}

// buildNotificationBundle constructs a NotificationBundle based on the subscription's
// content level (empty, id-only, full-resource).
func (e *SubscriptionTopicEngine) buildNotificationBundle(
	sub *TopicSubscription,
	topic *SubscriptionTopic,
	event ResourceEvent,
	resource map[string]interface{},
) *NotificationBundle {
	nb := &NotificationBundle{
		Type:           "event-notification",
		SubscriptionID: sub.ID,
		TopicURL:       topic.URL,
	}

	switch sub.Content {
	case "empty":
		// No resource content
	case "id-only":
		if resource != nil {
			nb.FocusResource = map[string]interface{}{
				"resourceType": resource["resourceType"],
				"id":           resource["id"],
			}
		}
	case "full-resource":
		if resource != nil {
			// Deep copy
			copied := make(map[string]interface{})
			raw, _ := json.Marshal(resource)
			_ = json.Unmarshal(raw, &copied)
			nb.FocusResource = copied
		}
	default:
		// Default to full-resource if not specified
		if resource != nil {
			copied := make(map[string]interface{})
			raw, _ := json.Marshal(resource)
			_ = json.Unmarshal(raw, &copied)
			nb.FocusResource = copied
		}
	}

	return nb
}

// ============================================================================
// Built-in Topics
// ============================================================================

const builtInTopicBase = "http://ehr.example.org/SubscriptionTopic/"

// RegisterBuiltInTopics registers the four standard built-in topics.
func (e *SubscriptionTopicEngine) RegisterBuiltInTopics() {
	e.RegisterTopic(&SubscriptionTopic{
		ID:     "encounter-start",
		URL:    builtInTopicBase + "encounter-start",
		Name:   "EncounterStart",
		Title:  "Encounter Start",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{
				ResourceType:     "Encounter",
				Interaction:      []string{"create"},
				FHIRPathCriteria: "status = 'in-progress'",
			},
		},
		CanFilterBy: []TopicCanFilterBy{
			{Resource: "Encounter", FilterParameter: "status", Modifier: []string{"eq"}},
			{Resource: "Encounter", FilterParameter: "class.code"},
		},
	})

	e.RegisterTopic(&SubscriptionTopic{
		ID:     "encounter-end",
		URL:    builtInTopicBase + "encounter-end",
		Name:   "EncounterEnd",
		Title:  "Encounter End",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{
				ResourceType:     "Encounter",
				Interaction:      []string{"update"},
				FHIRPathCriteria: "status = 'finished'",
			},
		},
		CanFilterBy: []TopicCanFilterBy{
			{Resource: "Encounter", FilterParameter: "status", Modifier: []string{"eq"}},
		},
	})

	e.RegisterTopic(&SubscriptionTopic{
		ID:     "new-lab-result",
		URL:    builtInTopicBase + "new-lab-result",
		Name:   "NewLabResult",
		Title:  "New Lab Result",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{
				ResourceType:     "DiagnosticReport",
				Interaction:      []string{"create"},
				FHIRPathCriteria: "status = 'final'",
			},
		},
		CanFilterBy: []TopicCanFilterBy{
			{Resource: "DiagnosticReport", FilterParameter: "status", Modifier: []string{"eq"}},
			{Resource: "DiagnosticReport", FilterParameter: "code"},
		},
	})

	e.RegisterTopic(&SubscriptionTopic{
		ID:     "admission-discharge",
		URL:    builtInTopicBase + "admission-discharge",
		Name:   "AdmissionDischarge",
		Title:  "Admission / Discharge",
		Status: "active",
		ResourceTrigger: []TopicResourceTrigger{
			{
				ResourceType:     "Encounter",
				Interaction:      []string{"create", "update"},
				FHIRPathCriteria: "class.code = 'IMP' and (status = 'in-progress' or status = 'finished')",
			},
		},
		CanFilterBy: []TopicCanFilterBy{
			{Resource: "Encounter", FilterParameter: "status", Modifier: []string{"eq", "in"}},
			{Resource: "Encounter", FilterParameter: "class.code"},
		},
	})
}

// ============================================================================
// HTTP Handler
// ============================================================================

// TopicHandler provides HTTP handlers for SubscriptionTopic and topic-based Subscriptions.
type TopicHandler struct {
	engine *SubscriptionTopicEngine
}

// NewTopicHandler creates a new TopicHandler.
func NewTopicHandler(engine *SubscriptionTopicEngine) *TopicHandler {
	return &TopicHandler{engine: engine}
}

// RegisterRoutes registers routes on the provided Echo group.
// Expects the group to be the FHIR base (e.g., /fhir).
func (h *TopicHandler) RegisterRoutes(fhirGroup *echo.Group) {
	fhirGroup.GET("/SubscriptionTopic", h.ListTopics)
	fhirGroup.GET("/SubscriptionTopic/:id", h.GetTopicByID)
	fhirGroup.POST("/SubscriptionTopic", h.CreateTopic)
	fhirGroup.POST("/Subscription", h.CreateSubscription)
	fhirGroup.GET("/Subscription/:id/$status", h.GetSubscriptionStatus)
	fhirGroup.GET("/Subscription/:id/$events", h.GetSubscriptionEvents)
	fhirGroup.POST("/Subscription/:id/$events", h.ReplayEvents)
}

// ListTopics handles GET /SubscriptionTopic.
func (h *TopicHandler) ListTopics(c echo.Context) error {
	topics := h.engine.ListTopics()
	entries := make([]interface{}, len(topics))
	for i, t := range topics {
		entries[i] = t.toFHIR()
	}

	total := len(entries)
	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"total":        total,
		"entry":        toEntries(entries),
	}
	return c.JSON(http.StatusOK, bundle)
}

// GetTopicByID handles GET /SubscriptionTopic/:id.
func (h *TopicHandler) GetTopicByID(c echo.Context) error {
	id := c.Param("id")
	topic := h.engine.GetTopic(id)
	if topic == nil {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("SubscriptionTopic", id))
	}
	return c.JSON(http.StatusOK, topic.toFHIR())
}

// CreateTopic handles POST /SubscriptionTopic.
func (h *TopicHandler) CreateTopic(c echo.Context) error {
	var topic SubscriptionTopic
	if err := json.NewDecoder(c.Request().Body).Decode(&topic); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid JSON: "+err.Error()))
	}
	if topic.ID == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("SubscriptionTopic.id is required"))
	}
	h.engine.RegisterTopic(&topic)
	c.Response().Header().Set("Location", "/fhir/SubscriptionTopic/"+topic.ID)
	return c.JSON(http.StatusCreated, topic.toFHIR())
}

// createSubscriptionRequest is the JSON body for creating a topic-based subscription.
type createSubscriptionRequest struct {
	ResourceType string                    `json:"resourceType"`
	Topic        string                    `json:"topic"`
	Status       string                    `json:"status"`
	ChannelType  string                    `json:"channelType"`
	Endpoint     string                    `json:"endpoint"`
	Header       []string                  `json:"header,omitempty"`
	Content      string                    `json:"content"`
	FilterBy     []TopicSubscriptionFilter `json:"filterBy,omitempty"`
}

// CreateSubscription handles POST /Subscription (topic-based).
func (h *TopicHandler) CreateSubscription(c echo.Context) error {
	var req createSubscriptionRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid JSON: "+err.Error()))
	}
	if req.Topic == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("Subscription.topic is required"))
	}

	id := fmt.Sprintf("topic-sub-%d", time.Now().UnixNano())
	sub := &TopicSubscription{
		ID:          id,
		TopicURL:    req.Topic,
		Status:      req.Status,
		ChannelType: req.ChannelType,
		Endpoint:    req.Endpoint,
		Header:      req.Header,
		Content:     req.Content,
		FilterBy:    req.FilterBy,
	}
	if sub.Status == "" {
		sub.Status = "requested"
	}

	if err := h.engine.Subscribe(sub); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
	}

	result := map[string]interface{}{
		"resourceType": "Subscription",
		"id":           sub.ID,
		"topic":        sub.TopicURL,
		"status":       sub.Status,
		"channelType":  sub.ChannelType,
		"endpoint":     sub.Endpoint,
		"content":      sub.Content,
	}
	c.Response().Header().Set("Location", "/fhir/Subscription/"+sub.ID)
	return c.JSON(http.StatusCreated, result)
}

// GetSubscriptionStatus handles GET /Subscription/:id/$status.
func (h *TopicHandler) GetSubscriptionStatus(c echo.Context) error {
	id := c.Param("id")
	status := h.engine.GetSubscriptionStatus(id)
	if status == nil {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("Subscription", id))
	}
	result := map[string]interface{}{
		"resourceType":   "SubscriptionStatus",
		"type":           "query-status",
		"subscriptionId": status.SubscriptionID,
		"topicUrl":       status.TopicURL,
		"status":         status.Status,
		"eventCount":     status.EventCount,
	}
	return c.JSON(http.StatusOK, result)
}

// GetSubscriptionEvents handles GET /Subscription/:id/$events.
func (h *TopicHandler) GetSubscriptionEvents(c echo.Context) error {
	id := c.Param("id")
	sub := h.engine.GetSubscription(id)
	if sub == nil {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("Subscription", id))
	}
	events := h.engine.GetSubscriptionEvents(id)
	return c.JSON(http.StatusOK, buildEventBundle(id, events))
}

// ReplayEvents handles POST /Subscription/:id/$events — replays recent events.
func (h *TopicHandler) ReplayEvents(c echo.Context) error {
	id := c.Param("id")
	sub := h.engine.GetSubscription(id)
	if sub == nil {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("Subscription", id))
	}
	events := h.engine.GetSubscriptionEvents(id)
	return c.JSON(http.StatusOK, buildEventBundle(id, events))
}

// ============================================================================
// FHIR serialisation helpers
// ============================================================================

func (t *SubscriptionTopic) toFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "SubscriptionTopic",
		"id":           t.ID,
		"url":          t.URL,
		"status":       t.Status,
	}
	if t.Version != "" {
		result["version"] = t.Version
	}
	if t.Name != "" {
		result["name"] = t.Name
	}
	if t.Title != "" {
		result["title"] = t.Title
	}
	if len(t.ResourceTrigger) > 0 {
		triggers := make([]map[string]interface{}, len(t.ResourceTrigger))
		for i, rt := range t.ResourceTrigger {
			trigger := map[string]interface{}{
				"resource": rt.ResourceType,
			}
			if len(rt.Interaction) > 0 {
				trigger["supportedInteraction"] = rt.Interaction
			}
			if rt.FHIRPathCriteria != "" {
				trigger["fhirPathCriteria"] = rt.FHIRPathCriteria
			}
			if rt.QueryCriteria != nil {
				qc := map[string]interface{}{}
				if rt.QueryCriteria.Previous != "" {
					qc["previous"] = rt.QueryCriteria.Previous
				}
				if rt.QueryCriteria.Current != "" {
					qc["current"] = rt.QueryCriteria.Current
				}
				trigger["queryCriteria"] = qc
			}
			triggers[i] = trigger
		}
		result["resourceTrigger"] = triggers
	}
	if len(t.EventTrigger) > 0 {
		triggers := make([]map[string]interface{}, len(t.EventTrigger))
		for i, et := range t.EventTrigger {
			trigger := map[string]interface{}{}
			if et.Description != "" {
				trigger["description"] = et.Description
			}
			trigger["event"] = map[string]interface{}{
				"system": et.Event.System,
				"code":   et.Event.Code,
			}
			triggers[i] = trigger
		}
		result["eventTrigger"] = triggers
	}
	if len(t.CanFilterBy) > 0 {
		filters := make([]map[string]interface{}, len(t.CanFilterBy))
		for i, cf := range t.CanFilterBy {
			filter := map[string]interface{}{
				"filterParameter": cf.FilterParameter,
			}
			if cf.Resource != "" {
				filter["resource"] = cf.Resource
			}
			if len(cf.Modifier) > 0 {
				filter["modifier"] = cf.Modifier
			}
			filters[i] = filter
		}
		result["canFilterBy"] = filters
	}
	if len(t.NotificationShape) > 0 {
		shapes := make([]map[string]interface{}, len(t.NotificationShape))
		for i, ns := range t.NotificationShape {
			shape := map[string]interface{}{
				"resource": ns.Resource,
			}
			if len(ns.Include) > 0 {
				shape["include"] = ns.Include
			}
			shapes[i] = shape
		}
		result["notificationShape"] = shapes
	}
	return result
}

func toEntries(resources []interface{}) []map[string]interface{} {
	entries := make([]map[string]interface{}, len(resources))
	for i, r := range resources {
		entry := map[string]interface{}{
			"resource": r,
		}
		if m, ok := r.(map[string]interface{}); ok {
			if rt, ok := m["resourceType"].(string); ok {
				if id, ok := m["id"].(string); ok {
					entry["fullUrl"] = fmt.Sprintf("%s/%s", rt, id)
				}
			}
		}
		entries[i] = entry
	}
	return entries
}

func buildEventBundle(subscriptionID string, events []topicEvent) map[string]interface{} {
	entries := make([]map[string]interface{}, 0, len(events))
	for _, ev := range events {
		entry := map[string]interface{}{
			"fullUrl": fmt.Sprintf("%s/%s", ev.ResourceType, ev.ResourceID),
		}
		if ev.Resource != nil {
			entry["resource"] = ev.Resource
		}
		entry["request"] = map[string]interface{}{
			"method": actionToHTTPMethod(ev.Action),
			"url":    fmt.Sprintf("%s/%s", ev.ResourceType, ev.ResourceID),
		}
		entries = append(entries, entry)
	}
	total := len(entries)
	return map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "history",
		"total":        total,
		"entry":        entries,
	}
}

func actionToHTTPMethod(action string) string {
	switch action {
	case "create":
		return "POST"
	case "delete":
		return "DELETE"
	default:
		return "PUT"
	}
}
