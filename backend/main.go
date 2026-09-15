package main

import (
	"log"
	"os"
	"time"

	"realtime-chat-backend/auth"
	"realtime-chat-backend/database"
	"realtime-chat-backend/handlers"
	"realtime-chat-backend/ws"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	database.Connect()

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	hub := ws.NewHub(redisAddr)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3001", "http://localhost:5174"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("/api/v1")
	{
		api.POST("/auth/register", handlers.Register)
		api.POST("/auth/login", handlers.Login)

		protected := api.Group("/")
		protected.Use(auth.Middleware())
		{
			protected.GET("/users", handlers.ListUsers)
			protected.DELETE("/users/:id", handlers.DeleteUser)
			protected.POST("/rooms", handlers.CreateRoom)
			protected.GET("/rooms", handlers.ListRooms)
			protected.DELETE("/rooms/:id", handlers.DeleteRoom)
			protected.GET("/rooms/:id/messages", handlers.GetMessages)
		}

		// WebSocket upgrade also goes through auth middleware (token via ?token=)
		wsGroup := api.Group("/")
		wsGroup.Use(auth.Middleware())
		wsGroup.GET("/ws", ws.ServeWS(hub))
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Println("chat server starting on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
