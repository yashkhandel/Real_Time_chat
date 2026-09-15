package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Envelope is the wire format for every message sent over a WebSocket connection.
type Envelope struct {
	Type      string      `json:"type"` // "message" | "typing" | "presence" | "notification"
	RoomID    uint        `json:"room_id,omitempty"`
	Data      interface{} `json:"data"`
}

// Client wraps one active WebSocket connection for one user.
type Client struct {
	UserID uint
	Conn   *websocket.Conn
	Send   chan Envelope
	Rooms  map[uint]bool
	mu     sync.Mutex
}

func (c *Client) WriteLoop() {
	for msg := range c.Send {
		c.mu.Lock()
		err := c.Conn.WriteJSON(msg)
		c.mu.Unlock()
		if err != nil {
			return
		}
	}
}

// Hub keeps track of all locally-connected clients and relays messages between
// them. It also publishes to Redis so that other backend instances (in a
// horizontally-scaled deployment) receive the same events — this is what
// lets the chat system scale beyond a single process.
type Hub struct {
	mu         sync.RWMutex
	clients    map[uint]map[*Client]bool // userID -> set of connections (multi-tab support)
	redis      *redis.Client
	channel    string
}

func NewHub(redisAddr string) *Hub {
	h := &Hub{
		clients: make(map[uint]map[*Client]bool),
		redis:   redis.NewClient(&redis.Options{Addr: redisAddr}),
		channel: "chat:broadcast",
	}
	go h.subscribeRedis()
	return h
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[c.UserID] == nil {
		h.clients[c.UserID] = make(map[*Client]bool)
	}
	h.clients[c.UserID][c] = true
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients[c.UserID], c)
	if len(h.clients[c.UserID]) == 0 {
		delete(h.clients, c.UserID)
	}
	close(c.Send)
}

// BroadcastToUsers publishes an envelope to Redis, which this same instance
// (and any sibling instances in a multi-server deployment) is subscribed to.
// Delivery to locally-connected clients happens exclusively through the Redis
// subscription loop below — never call deliverLocal directly here, or messages
// would be delivered twice on single-instance setups (once directly, once via
// the round-trip through Redis).
func (h *Hub) BroadcastToUsers(userIDs []uint, env Envelope) {
	payload, err := json.Marshal(struct {
		UserIDs []uint   `json:"user_ids"`
		Env     Envelope `json:"env"`
	}{UserIDs: userIDs, Env: env})
	if err != nil {
		log.Println("ws: marshal for redis publish failed:", err)
		return
	}
	h.redis.Publish(context.Background(), h.channel, payload)
}

func (h *Hub) deliverLocal(userIDs []uint, env Envelope) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, uid := range userIDs {
		for client := range h.clients[uid] {
			select {
			case client.Send <- env:
			default:
				// slow consumer; drop rather than block the hub
			}
		}
	}
}

func (h *Hub) subscribeRedis() {
	ctx := context.Background()
	sub := h.redis.Subscribe(ctx, h.channel)
	ch := sub.Channel()

	for msg := range ch {
		var payload struct {
			UserIDs []uint   `json:"user_ids"`
			Env     Envelope `json:"env"`
		}
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			continue
		}
		h.deliverLocal(payload.UserIDs, payload.Env)
	}
}

// IsOnline reports whether a user has at least one live connection on this instance.
func (h *Hub) IsOnline(userID uint) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[userID]) > 0
}
