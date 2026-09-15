package ws

import (
	"log"
	"net/http"
	"time"

	"realtime-chat-backend/auth"
	"realtime-chat-backend/database"
	"realtime-chat-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // dev-friendly; tighten in production
}

type incomingMessage struct {
	Type   string `json:"type"` // "message" | "typing"
	RoomID uint   `json:"room_id"`
	Body   string `json:"body,omitempty"`
}

// ServeWS upgrades the HTTP connection and starts the client's read/write loops.
func ServeWS(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := auth.UserIDFromContext(c)

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("ws upgrade failed:", err)
			return
		}

		client := &Client{
			UserID: userID,
			Conn:   conn,
			Send:   make(chan Envelope, 32),
			Rooms:  make(map[uint]bool),
		}
		hub.Register(client)
		go client.WriteLoop()

		announcePresence(hub, userID, true)
		defer func() {
			hub.Unregister(client)
			announcePresence(hub, userID, false)
			conn.Close()
		}()

		conn.SetReadLimit(4096)
		for {
			var msg incomingMessage
			if err := conn.ReadJSON(&msg); err != nil {
				break // client disconnected or sent malformed data
			}

			switch msg.Type {
			case "message":
				handleChatMessage(hub, userID, msg)
			case "typing":
				handleTyping(hub, userID, msg.RoomID)
			}
		}
	}
}

func handleChatMessage(hub *Hub, senderID uint, msg incomingMessage) {
	if msg.Body == "" {
		return
	}

	record := models.Message{RoomID: msg.RoomID, SenderID: senderID, Body: msg.Body}
	if err := database.DB.Create(&record).Error; err != nil {
		log.Println("failed to persist message:", err)
		return
	}
	database.DB.Preload("Sender").First(&record, record.ID)

	memberIDs := roomMemberIDs(msg.RoomID)
	hub.BroadcastToUsers(memberIDs, Envelope{
		Type:   "message",
		RoomID: msg.RoomID,
		Data:   record,
	})
}

func handleTyping(hub *Hub, userID uint, roomID uint) {
	memberIDs := roomMemberIDs(roomID)
	hub.BroadcastToUsers(memberIDs, Envelope{
		Type:   "typing",
		RoomID: roomID,
		Data:   map[string]interface{}{"user_id": userID, "at": time.Now()},
	})
}

func announcePresence(hub *Hub, userID uint, online bool) {
	// Broadcast presence to everyone the user shares a room with.
	var roomIDs []uint
	database.DB.Model(&models.RoomMember{}).Where("user_id = ?", userID).Pluck("room_id", &roomIDs)

	seen := map[uint]bool{}
	var recipients []uint
	for _, rid := range roomIDs {
		for _, uid := range roomMemberIDs(rid) {
			if !seen[uid] {
				seen[uid] = true
				recipients = append(recipients, uid)
			}
		}
	}

	hub.BroadcastToUsers(recipients, Envelope{
		Type: "presence",
		Data: map[string]interface{}{"user_id": userID, "online": online},
	})
}

func roomMemberIDs(roomID uint) []uint {
	var ids []uint
	database.DB.Model(&models.RoomMember{}).Where("room_id = ?", roomID).Pluck("user_id", &ids)
	return ids
}
