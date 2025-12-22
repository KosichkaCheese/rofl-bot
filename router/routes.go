package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"rofl-bot/config"
	"rofl-bot/controller"
	"rofl-bot/domain"
	"rofl-bot/service"

	"time"
)

// @Summary Проверка жизнеспособности сервера
// @Description Проверяет, жив ли сервер
// @Produce json
// @Success 200 {object} domain.SuccessResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/ [get]
func HealthCheck(c *gin.Context) {
	c.JSON(200, domain.SuccessResponse{
		Message: "я родился",
	})
}

// @Summary Тест авторизации
// @Description Доступ будет только при успешной авторизации
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /alena-rofl/user_check [get]
func UserCheck(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(401, domain.ErrorResponse{Error: "Unauthorized or can't find user_id in context."})
		return
	}
	c.JSON(200, domain.SuccessResponse{
		Message: "Привет, " + username.(string),
	})
}

func SetupRouter(cfg *config.Config, authCtrl *controller.AuthCtrl, userCtrl *controller.UserCtrl, eventCtrl *controller.EventCtrl, productCtrl *controller.ProductCtrl) *gin.Engine {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"https://tg-app-event-planner.vercel.app"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
		MaxAge:       24 * time.Hour,
	}))

	router.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(204)
	})

	api := router.Group("/alena-rofl")
	{
		api.GET("/", HealthCheck)
		api.GET("/user_check", service.AuthMiddleware(cfg), UserCheck)
		api.POST("/auth", authCtrl.Auth)
		api.POST("/refresh", authCtrl.Refresh)
		api.DELETE("/clear", authCtrl.ClearTokens)

		api.GET("/user/:id", service.AuthMiddleware(cfg), userCtrl.GetUser)
		api.PUT("/user", service.AuthMiddleware(cfg), userCtrl.UpdateUser)
		api.DELETE("/user/:id", service.AuthMiddleware(cfg), userCtrl.DeleteUser)
		api.GET("/user/whoami", service.AuthMiddleware(cfg), userCtrl.SelfID)

		api.GET("/event/:id", service.AuthMiddleware(cfg), eventCtrl.GetEvent)
		api.GET("/events", service.AuthMiddleware(cfg), eventCtrl.GetUserEvents)
		api.GET("/event/:id/admin", service.AuthMiddleware(cfg), eventCtrl.GetEventAdmin)
		api.POST("/event", service.AuthMiddleware(cfg), eventCtrl.CreateEvent)
		api.PUT("/event", service.AuthMiddleware(cfg), eventCtrl.UpdateEvent)
		api.DELETE("/event/:id", service.AuthMiddleware(cfg), eventCtrl.DeleteEvent)
		api.POST("/invite/:id", service.AuthMiddleware(cfg), eventCtrl.JoinEvent)
		api.PUT("/event/:id/leave", service.AuthMiddleware(cfg), eventCtrl.LeaveEvent)
		api.GET("/event/:id/members", service.AuthMiddleware(cfg), eventCtrl.GetEventMembers)
		api.POST("/event/:id/invite", service.AuthMiddleware(cfg), eventCtrl.CreateInvite)

		api.GET("/event/:id/product/:product_id", service.AuthMiddleware(cfg), productCtrl.GetProduct)
		api.GET("/event/:id/products", service.AuthMiddleware(cfg), productCtrl.GetProductsByEvent)
		api.POST("/event/:id/product", service.AuthMiddleware(cfg), productCtrl.CreateProduct)
		api.DELETE("/event/:id/product/:product_id", service.AuthMiddleware(cfg), productCtrl.DeleteProduct)
		api.PUT("/event/:id/product/:product_id", service.AuthMiddleware(cfg), productCtrl.UpdateProduct)
		api.GET("/event/:id/bill", service.AuthMiddleware(cfg), productCtrl.GetUserBill)
	}

	return router
}
