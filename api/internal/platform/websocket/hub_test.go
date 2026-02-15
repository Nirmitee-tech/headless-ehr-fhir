package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Hub tests
// ---------------------------------------------------------------------------

func TestHub_RegisterClient(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:     "client-1",
		Topics: []string{"Patient/123"},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}

	hub.Register(client)

	if hub.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", hub.ClientCount())
	}
	if hub.TopicCount("Patient/123") != 1 {
		t.Fatalf("expected 1 client on Patient/123, got %d", hub.TopicCount("Patient/123"))
	}
}

func TestHub_UnregisterClient(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:     "client-2",
		Topics: []string{"Patient/456"},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}

	hub.Register(client)
	hub.Unregister(client)

	if hub.ClientCount() != 0 {
		t.Fatalf("expected 0 clients, got %d", hub.ClientCount())
	}
	if hub.TopicCount("Patient/456") != 0 {
		t.Fatalf("expected 0 clients on Patient/456, got %d", hub.TopicCount("Patient/456"))
	}
}

func TestHub_BroadcastToTopic(t *testing.T) {
	hub := NewHub()

	subscriber := &Client{
		ID:     "sub-1",
		Topics: []string{"Patient/123"},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}
	nonSubscriber := &Client{
		ID:     "non-sub-1",
		Topics: []string{"Encounter/999"},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}

	hub.Register(subscriber)
	hub.Register(nonSubscriber)

	event := Event{
		Type:         "resource.created",
		Topic:        "Patient/123",
		ResourceType: "Patient",
		ResourceID:   "123",
		Timestamp:    time.Now(),
	}

	hub.Broadcast("Patient/123", event)

	select {
	case msg := <-subscriber.Send:
		var received Event
		if err := json.Unmarshal(msg, &received); err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		if received.Type != "resource.created" {
			t.Fatalf("expected event type resource.created, got %s", received.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("subscriber did not receive event")
	}

	select {
	case <-nonSubscriber.Send:
		t.Fatal("non-subscriber should not have received event")
	default:
		// expected
	}
}

func TestHub_BroadcastAll(t *testing.T) {
	hub := NewHub()

	c1 := &Client{
		ID:     "all-1",
		Topics: []string{"Patient/1"},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}
	c2 := &Client{
		ID:     "all-2",
		Topics: []string{"Encounter/2"},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}

	hub.Register(c1)
	hub.Register(c2)

	event := Event{
		Type:         "system.alert",
		Topic:        "system",
		ResourceType: "System",
		Timestamp:    time.Now(),
	}

	hub.BroadcastAll(event)

	for _, c := range []*Client{c1, c2} {
		select {
		case msg := <-c.Send:
			var received Event
			if err := json.Unmarshal(msg, &received); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			if received.Type != "system.alert" {
				t.Fatalf("expected system.alert, got %s", received.Type)
			}
		case <-time.After(time.Second):
			t.Fatalf("client %s did not receive broadcast", c.ID)
		}
	}
}

func TestHub_ClientCount(t *testing.T) {
	hub := NewHub()

	if hub.ClientCount() != 0 {
		t.Fatalf("expected 0, got %d", hub.ClientCount())
	}

	clients := make([]*Client, 5)
	for i := range clients {
		clients[i] = &Client{
			ID:     "count-" + string(rune('a'+i)),
			Topics: []string{"Topic/x"},
			Send:   make(chan []byte, 256),
			hub:    hub,
		}
		hub.Register(clients[i])
	}

	if hub.ClientCount() != 5 {
		t.Fatalf("expected 5, got %d", hub.ClientCount())
	}

	hub.Unregister(clients[0])
	hub.Unregister(clients[1])

	if hub.ClientCount() != 3 {
		t.Fatalf("expected 3, got %d", hub.ClientCount())
	}
}

func TestHub_TopicCount(t *testing.T) {
	hub := NewHub()

	c1 := &Client{
		ID:     "tc-1",
		Topics: []string{"Patient/1"},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}
	c2 := &Client{
		ID:     "tc-2",
		Topics: []string{"Patient/1"},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}
	c3 := &Client{
		ID:     "tc-3",
		Topics: []string{"Encounter/5"},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}

	hub.Register(c1)
	hub.Register(c2)
	hub.Register(c3)

	if hub.TopicCount("Patient/1") != 2 {
		t.Fatalf("expected 2 on Patient/1, got %d", hub.TopicCount("Patient/1"))
	}
	if hub.TopicCount("Encounter/5") != 1 {
		t.Fatalf("expected 1 on Encounter/5, got %d", hub.TopicCount("Encounter/5"))
	}
	if hub.TopicCount("NonExistent") != 0 {
		t.Fatalf("expected 0 on NonExistent, got %d", hub.TopicCount("NonExistent"))
	}
}

func TestHub_MultipleTopics(t *testing.T) {
	hub := NewHub()

	client := &Client{
		ID:     "multi-1",
		Topics: []string{"Patient/1", "Encounter/2"},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}
	hub.Register(client)

	event := Event{
		Type:         "resource.updated",
		Topic:        "Patient/1",
		ResourceType: "Patient",
		ResourceID:   "1",
		Timestamp:    time.Now(),
	}
	hub.Broadcast("Patient/1", event)

	select {
	case msg := <-client.Send:
		var received Event
		if err := json.Unmarshal(msg, &received); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if received.Topic != "Patient/1" {
			t.Fatalf("expected topic Patient/1, got %s", received.Topic)
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive event on Patient/1")
	}

	// Verify client is registered on both topics
	if hub.TopicCount("Patient/1") != 1 {
		t.Fatalf("expected 1 on Patient/1, got %d", hub.TopicCount("Patient/1"))
	}
	if hub.TopicCount("Encounter/2") != 1 {
		t.Fatalf("expected 1 on Encounter/2, got %d", hub.TopicCount("Encounter/2"))
	}
}

func TestHub_UnregisterClosesChannel(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:     "close-1",
		Topics: []string{"Topic/a"},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}

	hub.Register(client)
	hub.Unregister(client)

	// Reading from a closed channel returns zero value immediately
	_, ok := <-client.Send
	if ok {
		t.Fatal("expected Send channel to be closed after unregister")
	}
}

func TestHub_BroadcastToEmptyTopic(t *testing.T) {
	hub := NewHub()

	event := Event{
		Type:         "resource.deleted",
		Topic:        "NoOneHere",
		ResourceType: "Observation",
		ResourceID:   "999",
		Timestamp:    time.Now(),
	}

	// Should not panic
	hub.Broadcast("NoOneHere", event)
}

func TestHub_ConcurrentRegisterUnregister(t *testing.T) {
	hub := NewHub()
	const n = 100

	var wg sync.WaitGroup
	wg.Add(n * 2)

	clients := make([]*Client, n)
	for i := 0; i < n; i++ {
		clients[i] = &Client{
			ID:     "concurrent-" + string(rune(i)),
			Topics: []string{"Topic/concurrent"},
			Send:   make(chan []byte, 256),
			hub:    hub,
		}
	}

	// Register all concurrently
	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			hub.Register(clients[idx])
		}(i)
	}

	// Unregister all concurrently
	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			hub.Unregister(clients[idx])
		}(i)
	}

	wg.Wait()

	// Final count should be consistent (all registered then unregistered, or some mix)
	count := hub.ClientCount()
	if count < 0 {
		t.Fatalf("client count should not be negative, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// Event tests
// ---------------------------------------------------------------------------

func TestEvent_JSONSerialization(t *testing.T) {
	ts := time.Date(2026, 2, 15, 10, 30, 0, 0, time.UTC)
	event := Event{
		Type:         "resource.created",
		Topic:        "Patient/abc-123",
		ResourceType: "Patient",
		ResourceID:   "abc-123",
		Timestamp:    ts,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	if decoded.Type != event.Type {
		t.Fatalf("Type mismatch: %s vs %s", decoded.Type, event.Type)
	}
	if decoded.Topic != event.Topic {
		t.Fatalf("Topic mismatch: %s vs %s", decoded.Topic, event.Topic)
	}
	if decoded.ResourceType != event.ResourceType {
		t.Fatalf("ResourceType mismatch: %s vs %s", decoded.ResourceType, event.ResourceType)
	}
	if decoded.ResourceID != event.ResourceID {
		t.Fatalf("ResourceID mismatch: %s vs %s", decoded.ResourceID, event.ResourceID)
	}
	if !decoded.Timestamp.Equal(event.Timestamp) {
		t.Fatalf("Timestamp mismatch: %v vs %v", decoded.Timestamp, event.Timestamp)
	}
}

func TestEvent_WithData(t *testing.T) {
	payload := json.RawMessage(`{"name":"John Doe","birthDate":"1990-01-01"}`)
	event := Event{
		Type:         "resource.updated",
		Topic:        "Patient/xyz",
		ResourceType: "Patient",
		ResourceID:   "xyz",
		Timestamp:    time.Now(),
		Data:         payload,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event with data: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal event with data: %v", err)
	}

	if decoded.Data == nil {
		t.Fatal("expected Data to be non-nil")
	}

	var payloadMap map[string]interface{}
	if err := json.Unmarshal(decoded.Data, &payloadMap); err != nil {
		t.Fatalf("failed to unmarshal Data payload: %v", err)
	}
	if payloadMap["name"] != "John Doe" {
		t.Fatalf("expected name John Doe, got %v", payloadMap["name"])
	}
}

// ---------------------------------------------------------------------------
// Publisher tests
// ---------------------------------------------------------------------------

func TestHub_PublishEvent(t *testing.T) {
	hub := NewHub()

	client := &Client{
		ID:     "pub-1",
		Topics: []string{"Observation/100"},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}
	hub.Register(client)

	var publisher EventPublisher = hub

	event := Event{
		Type:         "resource.created",
		Topic:        "Observation/100",
		ResourceType: "Observation",
		ResourceID:   "100",
		Timestamp:    time.Now(),
	}

	if err := publisher.Publish(context.Background(), event); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	select {
	case msg := <-client.Send:
		var received Event
		if err := json.Unmarshal(msg, &received); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if received.ResourceID != "100" {
			t.Fatalf("expected ResourceID 100, got %s", received.ResourceID)
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive published event")
	}
}

func TestHub_PublishBroadcastsToSubscribers(t *testing.T) {
	hub := NewHub()

	c1 := &Client{
		ID:     "multi-pub-1",
		Topics: []string{"Patient/200"},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}
	c2 := &Client{
		ID:     "multi-pub-2",
		Topics: []string{"Patient/200"},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}
	c3 := &Client{
		ID:     "multi-pub-3",
		Topics: []string{"Encounter/300"},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}

	hub.Register(c1)
	hub.Register(c2)
	hub.Register(c3)

	event := Event{
		Type:         "resource.updated",
		Topic:        "Patient/200",
		ResourceType: "Patient",
		ResourceID:   "200",
		Timestamp:    time.Now(),
	}

	if err := hub.Publish(context.Background(), event); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Both subscribers should get the event
	for _, c := range []*Client{c1, c2} {
		select {
		case msg := <-c.Send:
			var received Event
			if err := json.Unmarshal(msg, &received); err != nil {
				t.Fatalf("client %s: failed to unmarshal: %v", c.ID, err)
			}
			if received.ResourceID != "200" {
				t.Fatalf("client %s: expected ResourceID 200, got %s", c.ID, received.ResourceID)
			}
		case <-time.After(time.Second):
			t.Fatalf("client %s did not receive event", c.ID)
		}
	}

	// Non-subscriber should not receive it
	select {
	case <-c3.Send:
		t.Fatal("c3 should not have received event for Patient/200")
	default:
		// expected
	}
}

// ---------------------------------------------------------------------------
// Handler tests
// ---------------------------------------------------------------------------

func TestWebSocketHandler_RegisterRoutes(t *testing.T) {
	hub := NewHub()
	handler := NewWebSocketHandler(hub)

	e := echo.New()
	g := e.Group("")
	handler.RegisterRoutes(g)

	routes := e.Routes()
	found := false
	for _, r := range routes {
		if r.Path == "/ws" && r.Method == http.MethodGet {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected GET /ws route to be registered")
	}
}

func TestWebSocketHandler_SubscribeMessage(t *testing.T) {
	msg := ClientMessage{
		Action: "subscribe",
		Topics: []string{"Patient/123", "Encounter/*"},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ClientMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Action != "subscribe" {
		t.Fatalf("expected action subscribe, got %s", decoded.Action)
	}
	if len(decoded.Topics) != 2 {
		t.Fatalf("expected 2 topics, got %d", len(decoded.Topics))
	}
	if decoded.Topics[0] != "Patient/123" {
		t.Fatalf("expected Patient/123, got %s", decoded.Topics[0])
	}
	if decoded.Topics[1] != "Encounter/*" {
		t.Fatalf("expected Encounter/*, got %s", decoded.Topics[1])
	}
}

func TestWebSocketHandler_UnsubscribeMessage(t *testing.T) {
	msg := ClientMessage{
		Action: "unsubscribe",
		Topics: []string{"Patient/123"},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ClientMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Action != "unsubscribe" {
		t.Fatalf("expected action unsubscribe, got %s", decoded.Action)
	}
	if len(decoded.Topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(decoded.Topics))
	}
}

func TestWebSocketHandler_InvalidMessage(t *testing.T) {
	invalidJSON := `{not valid json`

	var msg ClientMessage
	err := json.Unmarshal([]byte(invalidJSON), &msg)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestWebSocketHandler_HandleConnectRequiresWebSocket(t *testing.T) {
	hub := NewHub()
	handler := NewWebSocketHandler(hub)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandleConnect(c)

	// gorilla/websocket upgrader will reject non-WS requests
	if err == nil && rec.Code == http.StatusSwitchingProtocols {
		t.Fatal("expected upgrade to fail for non-websocket request")
	}
}

func TestHub_SubscribeAddsTopics(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:     "dynamic-sub-1",
		Topics: []string{},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}
	hub.Register(client)

	hub.Subscribe(client, []string{"Patient/new", "Encounter/new"})

	if hub.TopicCount("Patient/new") != 1 {
		t.Fatalf("expected 1 on Patient/new, got %d", hub.TopicCount("Patient/new"))
	}
	if hub.TopicCount("Encounter/new") != 1 {
		t.Fatalf("expected 1 on Encounter/new, got %d", hub.TopicCount("Encounter/new"))
	}
	if len(client.Topics) != 2 {
		t.Fatalf("expected 2 topics on client, got %d", len(client.Topics))
	}
}

func TestHub_UnsubscribeRemovesTopics(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:     "dynamic-unsub-1",
		Topics: []string{"Patient/1", "Encounter/2", "Observation/3"},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}
	hub.Register(client)

	hub.Unsubscribe(client, []string{"Patient/1", "Observation/3"})

	if hub.TopicCount("Patient/1") != 0 {
		t.Fatalf("expected 0 on Patient/1, got %d", hub.TopicCount("Patient/1"))
	}
	if hub.TopicCount("Encounter/2") != 1 {
		t.Fatalf("expected 1 on Encounter/2, got %d", hub.TopicCount("Encounter/2"))
	}
	if hub.TopicCount("Observation/3") != 0 {
		t.Fatalf("expected 0 on Observation/3, got %d", hub.TopicCount("Observation/3"))
	}
	if len(client.Topics) != 1 {
		t.Fatalf("expected 1 topic remaining, got %d", len(client.Topics))
	}
}

func TestClientMessage_ProcessSubscribe(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:     "process-1",
		Topics: []string{},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}
	hub.Register(client)

	raw := `{"action":"subscribe","topics":["Patient/123","Encounter/*"]}`
	var msg ClientMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	hub.ProcessMessage(client, msg)

	if hub.TopicCount("Patient/123") != 1 {
		t.Fatalf("expected 1 subscriber on Patient/123, got %d", hub.TopicCount("Patient/123"))
	}
}

func TestClientMessage_ProcessUnsubscribe(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:     "process-2",
		Topics: []string{"Patient/123", "Encounter/456"},
		Send:   make(chan []byte, 256),
		hub:    hub,
	}
	hub.Register(client)

	raw := `{"action":"unsubscribe","topics":["Patient/123"]}`
	var msg ClientMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	hub.ProcessMessage(client, msg)

	if hub.TopicCount("Patient/123") != 0 {
		t.Fatalf("expected 0 on Patient/123, got %d", hub.TopicCount("Patient/123"))
	}
	if hub.TopicCount("Encounter/456") != 1 {
		t.Fatalf("expected 1 on Encounter/456, got %d", hub.TopicCount("Encounter/456"))
	}
}

func TestWebSocketHandler_FullUpgradeWithDialer(t *testing.T) {
	hub := NewHub()
	handler := NewWebSocketHandler(hub)

	e := echo.New()
	g := e.Group("")
	handler.RegisterRoutes(g)

	server := httptest.NewServer(e)
	defer server.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	dialer := gorillawebsocket.Dialer{}
	conn, resp, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	defer conn.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected 101, got %d", resp.StatusCode)
	}

	// Client should have been registered in the hub
	// Give the goroutine a moment to register
	time.Sleep(50 * time.Millisecond)
	if hub.ClientCount() < 1 {
		t.Fatal("expected at least 1 client registered after connect")
	}

	// Send a subscribe message
	subMsg := ClientMessage{
		Action: "subscribe",
		Topics: []string{"Patient/test-ws"},
	}
	if err := conn.WriteJSON(subMsg); err != nil {
		t.Fatalf("failed to send subscribe: %v", err)
	}

	// Give the server time to process the subscribe
	time.Sleep(50 * time.Millisecond)

	if hub.TopicCount("Patient/test-ws") != 1 {
		t.Fatalf("expected 1 subscriber on Patient/test-ws, got %d", hub.TopicCount("Patient/test-ws"))
	}

	// Now broadcast an event and verify we receive it
	event := Event{
		Type:         "resource.created",
		Topic:        "Patient/test-ws",
		ResourceType: "Patient",
		ResourceID:   "test-ws",
		Timestamp:    time.Now(),
	}
	hub.Broadcast("Patient/test-ws", event)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var received Event
	if err := conn.ReadJSON(&received); err != nil {
		t.Fatalf("failed to read event: %v", err)
	}
	if received.Type != "resource.created" {
		t.Fatalf("expected resource.created, got %s", received.Type)
	}
	if received.ResourceID != "test-ws" {
		t.Fatalf("expected ResourceID test-ws, got %s", received.ResourceID)
	}
}
