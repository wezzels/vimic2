// Package api provides WebSocket support
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stsgym/vimic2/internal/pipeline"
)

// WebSocket message types (alias gorilla/websocket constants)
const (
	TextMessage   = websocket.TextMessage
	BinaryMessage = websocket.BinaryMessage
	CloseMessage  = websocket.CloseMessage
	PingMessage   = websocket.PingMessage
	PongMessage   = websocket.PongMessage
)

// DefaultUpgrader is the default WebSocket upgrader
var DefaultUpgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// WebSocketServer handles WebSocket connections
type WebSocketServer struct {
	coordinator *pipeline.Coordinator
	clients     map[*WebSocketClient]bool
	register    chan *WebSocketClient
	unregister  chan *WebSocketClient
	broadcast   chan *WebSocketMessage
	mu          sync.RWMutex
}

// WebSocketClient represents a WebSocket client
type WebSocketClient struct {
	conn      *websocket.Conn
	send      chan []byte
	closeChan chan struct{}
	filters   map[string]bool
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WebSocketEventType represents event types
type WebSocketEventType string

const (
	EventTypePipelineCreate  WebSocketEventType = "pipeline:create"
	EventTypePipelineStart   WebSocketEventType = "pipeline:start"
	EventTypePipelineStop    WebSocketEventType = "pipeline:stop"
	EventTypePipelineDestroy WebSocketEventType = "pipeline:destroy"
	EventTypePipelineUpdate  WebSocketEventType = "pipeline:update"
	EventTypeJobCreate       WebSocketEventType = "job:create"
	EventTypeJobStart        WebSocketEventType = "job:start"
	EventTypeJobComplete     WebSocketEventType = "job:complete"
	EventTypeJobFail         WebSocketEventType = "job:fail"
	EventTypeJobLog          WebSocketEventType = "job:log"
	EventTypeRunnerCreate    WebSocketEventType = "runner:create"
	EventTypeRunnerStart     WebSocketEventType = "runner:start"
	EventTypeRunnerStop      WebSocketEventType = "runner:stop"
	EventTypeRunnerDestroy   WebSocketEventType = "runner:destroy"
	EventTypeLogStream       WebSocketEventType = "log:stream"
)

// NewWebSocketServer creates a new WebSocket server
func NewWebSocketServer(coordinator *pipeline.Coordinator) *WebSocketServer {
	ws := &WebSocketServer{
		coordinator: coordinator,
		clients:     make(map[*WebSocketClient]bool),
		register:    make(chan *WebSocketClient),
		unregister:  make(chan *WebSocketClient),
		broadcast:   make(chan *WebSocketMessage, 1000),
	}

	// Start event processor
	go ws.processEvents()

	// Subscribe to coordinator events
	go ws.subscribeToCoordinator()

	return ws
}

// Run starts the WebSocket server
func (ws *WebSocketServer) Run() {
	for {
		select {
		case client := <-ws.register:
			ws.mu.Lock()
			ws.clients[client] = true
			ws.mu.Unlock()
			fmt.Printf("[WebSocket] Client connected. Total: %d\n", len(ws.clients))

		case client := <-ws.unregister:
			ws.mu.Lock()
			if _, ok := ws.clients[client]; ok {
				delete(ws.clients, client)
				close(client.send)
			}
			ws.mu.Unlock()
			fmt.Printf("[WebSocket] Client disconnected. Total: %d\n", len(ws.clients))

		case message := <-ws.broadcast:
			ws.mu.RLock()
			data, err := json.Marshal(message)
			if err != nil {
				ws.mu.RUnlock()
				continue
			}

			for client := range ws.clients {
				// Check filters
				if !ws.shouldSend(client, message) {
					continue
				}

				select {
				case client.send <- data:
				default:
					// Client buffer full, close connection
					close(client.send)
					delete(ws.clients, client)
				}
			}
			ws.mu.RUnlock()
		}
	}
}

// shouldSend checks if a message should be sent to a client based on filters
func (ws *WebSocketServer) shouldSend(client *WebSocketClient, message *WebSocketMessage) bool {
	// No filters, send all messages
	if len(client.filters) == 0 {
		return true
	}

	// Check if message type is in filters
	if client.filters[string(message.Type)] {
		return true
	}

	// Check payload for pipeline ID
	if payload, ok := message.Payload.(map[string]interface{}); ok {
		if pipelineID, ok := payload["pipeline_id"].(string); ok {
			if client.filters[pipelineID] {
				return true
			}
		}
	}

	return false
}

// subscribeToCoordinator subscribes to coordinator events
func (ws *WebSocketServer) subscribeToCoordinator() {
	// TODO: Implement event subscription when coordinator supports it
	// events := ws.coordinator.Events()
	// for event := range events {
	// 	message := &WebSocketMessage{
	// 		Type:    string(eventToWebSocketType(event)),
	// 		Payload: event,
	// 	}
	// 	ws.broadcast <- message
	// }
}

// processEvents processes WebSocket events
func (ws *WebSocketServer) processEvents() {
	// This could be used for additional event processing
}

// Broadcast broadcasts a message to all clients
func (ws *WebSocketServer) Broadcast(message *WebSocketMessage) {
	ws.broadcast <- message
}

// BroadcastTo broadcasts a message to specific clients
func (ws *WebSocketServer) BroadcastTo(clientIDs []string, message *WebSocketMessage) {
	// TODO: Implement targeted broadcasting
	ws.Broadcast(message)
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := DefaultUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("[WebSocket] Upgrade error: %v\n", err)
		return
	}

	// Create client
	client := &WebSocketClient{
		conn:      conn,
		send:      make(chan []byte, 256),
		closeChan: make(chan struct{}),
		filters:   make(map[string]bool),
	}

	// Register client
	s.ws.register <- client

	// Start write goroutine
	go client.writePump()

	// Start read goroutine
	go client.readPump()
}

// writePump pumps messages from the client to the WebSocket
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case <-c.closeChan:
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump pumps messages from the WebSocket to the client
func (c *WebSocketClient) readPump() {
	defer func() {
		c.closeChan <- struct{}{}
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		// Parse message
		var msg struct {
			Type    string      `json:"type"`
			Payload interface{} `json:"payload"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// Handle message
		switch msg.Type {
		case "subscribe":
			// Subscribe to specific events
			if payload, ok := msg.Payload.(map[string]interface{}); ok {
				if events, ok := payload["events"].([]interface{}); ok {
					for _, event := range events {
						if eventStr, ok := event.(string); ok {
							c.filters[eventStr] = true
						}
					}
				}
				if pipelineID, ok := payload["pipeline_id"].(string); ok {
					c.filters[pipelineID] = true
				}
			}

		case "unsubscribe":
			// Unsubscribe from events
			if payload, ok := msg.Payload.(map[string]interface{}); ok {
				if events, ok := payload["events"].([]interface{}); ok {
					for _, event := range events {
						if eventStr, ok := event.(string); ok {
							delete(c.filters, eventStr)
						}
					}
				}
				if pipelineID, ok := payload["pipeline_id"].(string); ok {
					delete(c.filters, pipelineID)
				}
			}

		case "ping":
			// Respond to ping
			c.send <- []byte(`{"type":"pong"}`)
		}
	}
}