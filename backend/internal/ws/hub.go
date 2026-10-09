package ws

import (
	"sync"

	"github.com/gofiber/websocket/v2"
)

// Hub manages admin websocket clients and broadcasts.
type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]bool
}

var globalHub = NewHub()

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]bool),
	}
}

// GetHub returns global hub instance.
func GetHub() *Hub {
	return globalHub
}

// Register adds a connection.
func (h *Hub) Register(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = true
}

// Unregister removes a connection.
func (h *Hub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, conn)
}

// Broadcast sends message to all clients non-blocking.
func (h *Hub) Broadcast(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.clients {
		// non-blocking per client in goroutine to avoid slow client blocking
		go func(c *websocket.Conn) {
			_ = c.WriteMessage(websocket.TextMessage, msg)
		}(conn)
	}
}

// Count returns number of connected clients.
func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
