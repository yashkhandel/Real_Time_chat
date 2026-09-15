package handlers

import (
	"net/http"
	"strconv"

	"realtime-chat-backend/auth"
	"realtime-chat-backend/database"
	"realtime-chat-backend/models"

	"github.com/gin-gonic/gin"
)

type CreateRoomInput struct {
	Name      string `json:"name" binding:"required"`
	IsGroup   bool   `json:"is_group"`
	MemberIDs []uint `json:"member_ids" binding:"required"`
}

// CreateRoom creates a room and adds the creator + given members to it.
func CreateRoom(c *gin.Context) {
	userID := auth.UserIDFromContext(c)

	var in CreateRoomInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room := models.Room{Name: in.Name, IsGroup: in.IsGroup, CreatedBy: userID}
	if err := database.DB.Create(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create room"})
		return
	}

	memberSet := map[uint]bool{userID: true}
	for _, id := range in.MemberIDs {
		memberSet[id] = true
	}
	for id := range memberSet {
		database.DB.Create(&models.RoomMember{RoomID: room.ID, UserID: id})
	}

	c.JSON(http.StatusCreated, room)
}

// ListRooms returns all rooms the authenticated user belongs to.
func ListRooms(c *gin.Context) {
	userID := auth.UserIDFromContext(c)

	var roomIDs []uint
	database.DB.Model(&models.RoomMember{}).Where("user_id = ?", userID).Pluck("room_id", &roomIDs)

	var rooms []models.Room
	database.DB.Where("id IN ?", roomIDs).Find(&rooms)

	c.JSON(http.StatusOK, rooms)
}

// GetMessages returns paginated message history for a room (most recent first).
func GetMessages(c *gin.Context) {
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	var messages []models.Message
	database.DB.
		Preload("Sender").
		Where("room_id = ?", roomID).
		Order("created_at desc").
		Limit(limit).
		Find(&messages)

	// reverse to chronological order for the client
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	c.JSON(http.StatusOK, messages)
}

// ListUsers returns all users, for starting new conversations.
func ListUsers(c *gin.Context) {
	var users []models.User
	database.DB.Find(&users)
	c.JSON(http.StatusOK, users)
}

// DeleteRoom removes a room along with its membership rows and message
// history. Only a member of the room may delete it.
func DeleteRoom(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	var membership models.RoomMember
	if err := database.DB.Where("room_id = ? AND user_id = ?", roomID, userID).
		First(&membership).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this room"})
		return
	}

	database.DB.Where("room_id = ?", roomID).Delete(&models.Message{})
	database.DB.Where("room_id = ?", roomID).Delete(&models.RoomMember{})
	if err := database.DB.Delete(&models.Room{}, roomID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete room"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// DeleteUser removes a user account entirely — used to clean up duplicate
// test accounts created while trying things out. Intentionally permissive
// (any signed-in user can remove any account) since this is a demo/test
// project, not a multi-tenant production app.
func DeleteUser(c *gin.Context) {
	targetID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	database.DB.Where("user_id = ?", targetID).Delete(&models.RoomMember{})
	database.DB.Where("sender_id = ?", targetID).Delete(&models.Message{})
	if err := database.DB.Delete(&models.User{}, targetID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
