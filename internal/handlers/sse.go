package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"swadesh-dashboard/internal/models"

	"github.com/labstack/echo/v4"
)

// SSEHub manages Server-Sent Events connections
type SSEHub struct {
	clients    map[chan string]bool
	broadcast  chan models.MLPayload
	register   chan chan string
	unregister chan chan string
	mu         sync.RWMutex
}

// NewSSEHub creates a new SSE hub
func NewSSEHub(broadcast chan models.MLPayload) *SSEHub {
	return &SSEHub{
		clients:    make(map[chan string]bool),
		broadcast:  broadcast,
		register:   make(chan chan string),
		unregister: make(chan chan string),
	}
}

// Run starts the SSE hub event loop
func (h *SSEHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("SSE client connected (total: %d)", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client)
			}
			h.mu.Unlock()
			log.Printf("SSE client disconnected (total: %d)", len(h.clients))

		case payload := <-h.broadcast:
			data, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Failed to marshal payload: %v", err)
				continue
			}

			message := fmt.Sprintf("data: %s\n\n", data)

			h.mu.RLock()
			for client := range h.clients {
				select {
				case client <- message:
				default:
					// Client buffer full, skip
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Handler returns the Echo handler for SSE connections
func (h *SSEHub) Handler() echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Cache-Control", "no-cache")
		c.Response().Header().Set("Connection", "keep-alive")
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")

		// Create client channel
		client := make(chan string, 10)
		h.register <- client

		// Ensure cleanup on disconnect
		defer func() {
			h.unregister <- client
		}()

		// Send initial connection message
		fmt.Fprintf(c.Response(), "data: {\"type\":\"connected\"}\n\n")
		c.Response().Flush()

		// Stream events
		for {
			select {
			case message, ok := <-client:
				if !ok {
					return nil
				}
				fmt.Fprint(c.Response(), message)
				c.Response().Flush()

			case <-c.Request().Context().Done():
				return nil
			}
		}
	}
}
