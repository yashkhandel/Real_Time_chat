package handlers

import (
	"net/http"

	"realtime-chat-backend/auth"
	"realtime-chat-backend/database"
	"realtime-chat-backend/models"

	"github.com/gin-gonic/gin"
)

var avatarPalette = []string{"#3DDAD7", "#F2A93B", "#F2554D", "#8E7CF0", "#5CC8FF"}

type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	var in RegisterInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not process password"})
		return
	}

	user := models.User{
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: hash,
		AvatarColor:  avatarPalette[len(in.Name)%len(avatarPalette)],
	}
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	token, _ := auth.GenerateToken(user.ID)
	c.JSON(http.StatusCreated, gin.H{"token": token, "user": user})
}

func Login(c *gin.Context) {
	var in LoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", in.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if !auth.CheckPassword(user.PasswordHash, in.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, _ := auth.GenerateToken(user.ID)
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}
