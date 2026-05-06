package realtime

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// ══════════════════════════════════════════════════════
//  MESSAGE TYPES
// ══════════════════════════════════════════════════════

type MessageType string

const (
	MsgRefreshWeb MessageType = "refresh_web"
	MsgRefreshAll MessageType = "refresh_all"
)

type Message struct {
	Type    MessageType    `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Target  string         `json:"target,omitempty"` // "web" or "all"
}

// ══════════════════════════════════════════════════════
//  CLIENT
// ══════════════════════════════════════════════════════

type ClientType string

const (
	ClientWeb ClientType = "web"
)

type Client struct {
	Conn *websocket.Conn
	Type ClientType
	Send chan []byte
}

// ══════════════════════════════════════════════════════
//  HUB — Central WebSocket broadcaster
// ══════════════════════════════════════════════════════

type Hub struct {
	clients    map[*Client]bool
	mu         sync.RWMutex
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("🔌 WS client connected: type=%s (total: %d)",
				client.Type, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("🔌 WS client disconnected: type=%s", client.Type)

		case msg := <-h.broadcast:
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("❌ WS marshal error: %v", err)
				continue
			}

			h.mu.RLock()
			for client := range h.clients {
				shouldSend := false

				switch msg.Target {
				case "all":
					shouldSend = true
				case "web":
					shouldSend = client.Type == ClientWeb
				}

				if shouldSend {
					select {
					case client.Send <- data:
					default:
						// Client buffer full, disconnect
						go func(c *Client) {
							h.unregister <- c
						}(client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register adds a client to the hub
func (h *Hub) Register(c *Client) {
	h.register <- c
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

// Broadcast sends a message to matching clients
func (h *Hub) Broadcast(msg Message) {
	h.broadcast <- msg
}

// BroadcastAll sends refresh to all connected clients
func (h *Hub) BroadcastAll() {
	h.Broadcast(Message{
		Type:   MsgRefreshAll,
		Target: "all",
	})
}

// BroadcastWeb sends a refresh signal to web clients only
func (h *Hub) BroadcastWeb() {
	h.Broadcast(Message{
		Type:   MsgRefreshWeb,
		Target: "web",
	})
}

// ══════════════════════════════════════════════════════
//  POCKETBASE HOOKS — triggers broadcasts on data changes
// ══════════════════════════════════════════════════════

var hubInstance *Hub

func RegisterPBHooks(pb *pocketbase.PocketBase) {
	pb.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// content_blocks changes (events + news) → refresh all displays and web
		se.App.OnRecordAfterCreateSuccess("content_blocks").BindFunc(func(e *core.RecordEvent) error {
			if hubInstance != nil {
				hubInstance.BroadcastAll()
			}
			return e.Next()
		})
		se.App.OnRecordAfterUpdateSuccess("content_blocks").BindFunc(func(e *core.RecordEvent) error {
			if hubInstance != nil {
				hubInstance.BroadcastAll()
			}
			return e.Next()
		})
		se.App.OnRecordAfterDeleteSuccess("content_blocks").BindFunc(func(e *core.RecordEvent) error {
			if hubInstance != nil {
				hubInstance.BroadcastAll()
			}
			return e.Next()
		})

		return se.Next()
	})
}

// SetHubInstance stores the hub reference for PB hooks
func SetHubInstance(h *Hub) {
	hubInstance = h
}
