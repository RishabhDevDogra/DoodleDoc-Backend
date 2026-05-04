package hub

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// message type names sent over the WebSocket.
const (
	MsgDocumentCreated = "DocumentCreated"
	MsgDocumentUpdated = "DocumentUpdated"
	MsgDocumentDeleted = "DocumentDeleted"
	MsgEventAdded      = "EventAdded"
	MsgCommentAdded    = "CommentAdded"
	MsgConnected       = "Connected"
)

// Message is the envelope sent over the WebSocket to every client.
type Message struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

// client represents one connected browser tab.
type client struct {
	conn *websocket.Conn
	send chan []byte
}

// Hub manages all active WebSocket connections and broadcasts to them.
// It broadcasts events to all connected clients.
type Hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}
}

var upgrader = websocket.Upgrader{
	// Allow all origins for now; restrict in production via config.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// New creates a Hub ready for use. No goroutines are started here;
// they launch per connection in ServeWS.
func New() *Hub {
	return &Hub{clients: make(map[*client]struct{})}
}

// ServeWS upgrades an HTTP request to a WebSocket connection and
// registers the client with the hub. Mount this at /hubs/document.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	c := &client{conn: conn, send: make(chan []byte, 256)}

	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()

	// Send a Connected message immediately so the frontend knows it's live.
	h.sendTo(c, Message{Type: MsgConnected, Payload: nil})

	// Writer goroutine: drains the send channel to the WebSocket.
	go func() {
		defer func() {
			h.unregister(c)
			conn.Close()
		}()
		for msg := range c.send {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}()

	// Reader goroutine: keeps the connection alive and handles client pings.
	// We don't expect messages from clients (server pushes only), but we must
	// read to process control frames (ping/pong/close).
	go func() {
		defer func() {
			h.unregister(c)
			conn.Close()
		}()
		conn.SetReadLimit(512)
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

// ── Broadcaster interface methods ─────────────────────────────────────────────

func (h *Hub) BroadcastDocumentCreated(id, title string) {
	h.broadcast(Message{
		Type:    MsgDocumentCreated,
		Payload: map[string]string{"documentId": id, "title": title},
	})
}

func (h *Hub) BroadcastDocumentUpdated(id string) {
	h.broadcast(Message{
		Type:    MsgDocumentUpdated,
		Payload: map[string]string{"documentId": id},
	})
}

func (h *Hub) BroadcastDocumentDeleted(id string) {
	h.broadcast(Message{
		Type:    MsgDocumentDeleted,
		Payload: map[string]string{"documentId": id},
	})
}

func (h *Hub) BroadcastEventAdded(docID, eventType, description string, ts time.Time) {
	h.broadcast(Message{
		Type: MsgEventAdded,
		Payload: map[string]any{
			"documentId":  docID,
			"eventType":   eventType,
			"description": description,
			"timestamp":   ts,
		},
	})
}

func (h *Hub) BroadcastCommentAdded(docID string) {
	h.broadcast(Message{
		Type:    MsgCommentAdded,
		Payload: map[string]string{"documentId": docID},
	})
}

// ── internal helpers ──────────────────────────────────────────────────────────

func (h *Hub) broadcast(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("hub marshal error: %v", err)
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		select {
		case c.send <- data:
		default:
			// Slow client — drop the message rather than blocking broadcasts.
		}
	}
}

func (h *Hub) sendTo(c *client, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case c.send <- data:
	default:
	}
}

func (h *Hub) unregister(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
	}
}
