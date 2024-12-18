package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gitslim/gophermart/internal/conf"
	"github.com/gitslim/gophermart/internal/logging"
	"github.com/golang-jwt/jwt/v5"
)

const (
	authCookie = "auth_token"
	userIDKey  = "userID"
)

// AuthMiddleware предоставляет middleware для аутентификации
type AuthMiddleware struct {
	secretKey []byte
	log       logging.Logger
}

// NewAuthMiddleware создает новый экземпляр AuthMiddleware
func NewAuthMiddleware(config *conf.Config, log logging.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		secretKey: []byte(config.SecretKey),
		log:       log,
	}
}

// AuthRequired проверяет JWT токен в куки
func (m *AuthMiddleware) AuthRequired(c *gin.Context) {
	cookie, err := c.Cookie(authCookie)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}

	token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
		return m.secretKey, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}

	c.Set(userIDKey, int64(userID))
	c.Next()
}

// GenerateToken создает новый JWT токен
func (m *AuthMiddleware) GenerateToken(userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(m.secretKey)
}

// SetAuthCookie устанавливает JWT токен в куки
func (m *AuthMiddleware) SetAuthCookie(c *gin.Context, token string) {
	c.SetCookie(
		authCookie,
		token,
		int(24*time.Hour.Seconds()), // максимальное время жизни - 24 часа
		"/",                         // путь
		"",                          // домен
		false,                       // secure
		true,                        // httpOnly
	)
}
