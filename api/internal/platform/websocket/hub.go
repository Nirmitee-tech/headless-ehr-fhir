// Package websocket provides a real-time notification system using WebSockets.
// It implements a hub-and-spoke pattern where clients subscribe to topics
// and receive events broadcast to those topics.
package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// Event represents a real-time notification sent to WebSocket clients.
type Event struct {
	Type         string          `json:"type"`
	Topic        string          `json:"topic"`
	ResourceType string          `json:"resourceType"`
	ResourceID   string          `json:"resourceId,omitempty"`
	Timestamp    time.Time       `json:"timestamp"`
	Data         json.RawMessage `json:"data,omitempty"`
}

// ClientMessage represents an inbound message from a WebSocket client.
type ClientMessage struct {
	Action string   `json:"action"`
	Topics []string `json:"topics"`
}

// EventPublisher defines the interface for publishing events to subscribers.
type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
}

// Conn abstracts a WebSocket connection for testability.
type Conn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

// Client represents a single WebSocket connection.
type Client struct {
	ID     string
	Topics []string
	Send   chan []byte
	hub    *Hub
	conn   Conn
}

// Hub is the central connection manager that tracks clients and their topic
// subscriptions. All operations are thread-safe via sync.RWMutex.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]struct{} // topic -> set of clients
	all     map[*Client]struct{}            // all connected clients
}

// NewHub creates a new Hub ready to manage WebSocket clients.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*Client]struct{}),
		all:     make(map[*Client]struct{}),
	}
}

// Register adds a client to the hub and subscribes it to its initial topics.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.all[client] = struct{}{}

	for _, topic := range client.Topics {
		if h.clients[topic] == nil {
			h.clients[topic] = make(map[*Client]struct{})
		}
		h.clients[topic][client] = struct{}{}
	}
}

// Unregister removes a client from the hub, all topic subscriptions, and
// closes the client's Send channel.
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.all[client]; !ok {
		return
	}

	for _, topic := range client.Topics {
		if subscribers, ok := h.clients[topic]; ok {
			delete(subscribers, client)
			if len(subscribers) == 0 {
				delete(h.clients, topic)
			}
		}
	}

	delete(h.all, client)
	close(client.Send)
}

// Subscribe dynamically adds topics to an already-registered client.
func (h *Hub) Subscribe(client *Client, topics []string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, topic := range topics {
		if h.clients[topic] == nil {
			h.clients[topic] = make(map[*Client]struct{})
		}
		h.clients[topic][client] = struct{}{}
	}
	client.Topics = append(client.Topics, topics...)
}

// Unsubscribe dynamically removes topics from an already-registered client.
func (h *Hub) Unsubscribe(client *Client, topics []string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	removeSet := make(map[string]struct{}, len(topics))
	for _, t := range topics {
		removeSet[t] = struct{}{}
	}

	for _, topic := range topics {
		if subscribers, ok := h.clients[topic]; ok {
			delete(subscribers, client)
			if len(subscribers) == 0 {
				delete(h.clients, topic)
			}
		}
	}

	remaining := make([]string, 0, len(client.Topics))
	for _, t := range client.Topics {
		if _, rm := removeSet[t]; !rm {
			remaining = append(remaining, t)
		}
	}
	client.Topics = remaining
}

// ProcessMessage handles an inbound ClientMessage, dispatching to Subscribe
// or Unsubscribe as appropriate.
func (h *Hub) ProcessMessage(client *Client, msg ClientMessage) {
	switch msg.Action {
	case "subscribe":
		h.Subscribe(client, msg.Topics)
	case "unsubscribe":
		h.Unsubscribe(client, msg.Topics)
	}
}

// Broadcast sends an event to all clients subscribed to the given topic.
func (h *Hub) Broadcast(topic string, event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("websocket: failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	subscribers, ok := h.clients[topic]
	if !ok {
		return
	}

	for client := range subscribers {
		select {
		case client.Send <- data:
		default:
			// Client buffer full; skip to avoid blocking.
		}
	}
}

// BroadcastAll sends an event to every connected client regardless of topic.
func (h *Hub) BroadcastAll(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("websocket: failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.all {
		select {
		case client.Send <- data:
		default:
			// Client buffer full; skip to avoid blocking.
		}
	}
}

// Publish implements the EventPublisher interface by broadcasting the event
// to subscribers of the event's topic.
func (h *Hub) Publish(_ context.Context, event Event) error {
	h.Broadcast(event.Topic, event)
	return nil
}

// ClientCount returns the total number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.all)
}

// TopicCount returns the number of clients subscribed to a specific topic.
func (h *Hub) TopicCount(topic string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[topic])
}

// ---------------------------------------------------------------------------
// WebSocketHandler — Echo HTTP handler for WebSocket connections
// ---------------------------------------------------------------------------

var upgrader = gorillawebsocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins; tighten in production.
	},
}

// WebSocketHandler handles HTTP-to-WebSocket upgrades and message routing.
type WebSocketHandler struct {
	hub *Hub
}

// NewWebSocketHandler creates a new handler bound to the given Hub.
func NewWebSocketHandler(hub *Hub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

// RegisterRoutes registers the WebSocket endpoint on the provided Echo group.
func (wsh *WebSocketHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/ws", wsh.HandleConnect)
}

// HandleConnect upgrades an HTTP connection to WebSocket, registers the
// client with the hub, and starts read/write pumps.
func (wsh *WebSocketHandler) HandleConnect(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	client := &Client{
		ID:     uuid.New().String(),
		Topics: []string{},
		Send:   make(chan []byte, 256),
		hub:    wsh.hub,
		conn:   &gorillaConnAdapter{ws},
	}

	wsh.hub.Register(client)

	go wsh.writePump(client, ws)
	go wsh.readPump(client, ws)

	return nil
}

// readPump reads messages from the WebSocket connection and processes them.
func (wsh *WebSocketHandler) readPump(client *Client, ws *gorillawebsocket.Conn) {
	defer func() {
		wsh.hub.Unregister(client)
		ws.Close()
	}()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}

		var msg ClientMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue // Ignore malformed messages.
		}

		wsh.hub.ProcessMessage(client, msg)
	}
}

// writePump writes messages from the Send channel to the WebSocket connection.
func (wsh *WebSocketHandler) writePump(client *Client, ws *gorillawebsocket.Conn) {
	defer ws.Close()

	for message := range client.Send {
		if err := ws.WriteMessage(gorillawebsocket.TextMessage, message); err != nil {
			break
		}
	}
}

// gorillaConnAdapter wraps a gorilla/websocket.Conn to satisfy the Conn interface.
type gorillaConnAdapter struct {
	conn *gorillawebsocket.Conn
}

func (a *gorillaConnAdapter) ReadMessage() (int, []byte, error) {
	return a.conn.ReadMessage()
}

func (a *gorillaConnAdapter) WriteMessage(messageType int, data []byte) error {
	return a.conn.WriteMessage(messageType, data)
}

func (a *gorillaConnAdapter) Close() error {
	return a.conn.Close()
}
