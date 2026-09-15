package auth

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(getSecret())

func getSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "dev-secret-change-me"
}

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func GenerateToken(userID uint) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(48 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken validates a raw JWT string and returns the user ID — used by
// both the HTTP middleware and the WebSocket handshake (which can't send headers
// after the initial upgrade, so the token often arrives as a query param).
func ParseToken(tokenStr string) (uint, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, err
	}
	return claims.UserID, nil
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		var tokenStr string
		if strings.HasPrefix(header, "Bearer ") {
			tokenStr = strings.TrimPrefix(header, "Bearer ")
		} else {
			tokenStr = c.Query("token") // fallback for WS upgrade requests
		}

		userID, err := ParseToken(tokenStr)
		if err != nil || userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing token"})
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

func UserIDFromContext(c *gin.Context) uint {
	v, _ := c.Get("user_id")
	id, _ := v.(uint)
	return id
}

// helper for places that need string parsing (route params)
func ParseUint(s string) uint {
	n, _ := strconv.ParseUint(s, 10, 64)
	return uint(n)
}
