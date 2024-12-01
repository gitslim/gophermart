package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gitslim/gophermart/internal/web/handlers"
	"github.com/gitslim/gophermart/internal/web/middleware"
)

// NewRouter настраивает маршрутизацию
func NewRouter(handler *handlers.Handler, auth *middleware.AuthMiddleware) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.GzipMiddleware())

	// Пинг для проверки здоровья
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Публичные маршруты
	r.POST("/api/user/register", handler.Register)
	r.POST("/api/user/login", handler.Login)

	// Защищенные маршруты
	authorized := r.Group("/api")
	authorized.Use(auth.AuthRequired)
	{
		// Заказы
		authorized.POST("/user/orders", handler.UploadOrder)
		authorized.GET("/user/orders", handler.GetOrders)

		// Баланс
		authorized.GET("/user/balance", handler.GetBalance)
		authorized.POST("/user/balance/withdraw", handler.Withdraw)
		authorized.GET("/user/withdrawals", handler.GetWithdrawals)
	}

	return r
}
