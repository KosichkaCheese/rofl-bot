package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"rofl-bot/config"
	"rofl-bot/controller"
	"rofl-bot/domain"
	"rofl-bot/service"
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

func SetupRouter(cfg *config.Config, authCtrl *controller.AuthCtrl) *gin.Engine {
	router := gin.Default()
	router.Use(cors.Default()) //потом поставить middleware
	api := router.Group("/alena-rofl")
	{
		api.GET("/", HealthCheck)
		api.GET("/user_check", service.AuthMiddleware(cfg), UserCheck)
		api.POST("/auth", authCtrl.Auth)
		api.POST("/refresh", authCtrl.Refresh)
		api.DELETE("/clear", authCtrl.ClearTokens)
	}

	return router
}
