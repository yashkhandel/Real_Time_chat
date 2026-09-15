package models

import "time"

// User is an authenticated chat participant
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:100;not null" json:"name"`
	Email        string    `gorm:"size:150;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	AvatarColor  string    `gorm:"size:10;default:'#3DDAD7'" json:"avatar_color"`
	CreatedAt    time.Time `json:"created_at"`
}

// Room represents a chat channel (direct message or group)
type Room struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:150;not null" json:"name"`
	IsGroup   bool      `gorm:"default:false" json:"is_group"`
	CreatedBy uint      `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	Members   []RoomMember `gorm:"foreignKey:RoomID" json:"-"`
}

// RoomMember links a user to a room they belong to
type RoomMember struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	RoomID   uint      `gorm:"index" json:"room_id"`
	UserID   uint      `gorm:"index" json:"user_id"`
	User     User      `gorm:"foreignKey:UserID" json:"user"`
	JoinedAt time.Time `json:"joined_at"`
}

// Message is a persisted chat message
type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RoomID    uint      `gorm:"index" json:"room_id"`
	SenderID  uint      `json:"sender_id"`
	Sender    User      `gorm:"foreignKey:SenderID" json:"sender"`
	Body      string    `gorm:"type:text" json:"body"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

// Notification is a per-user async event (mention, room invite, etc.)
type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Type      string    `gorm:"size:50" json:"type"`
	Payload   string    `gorm:"type:text" json:"payload"`
	Read      bool      `gorm:"default:false" json:"read"`
	CreatedAt time.Time `json:"created_at"`
}
