package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"rofl-bot/service"
)

// @Summary Проверка жизнеспособности сервера
// @Description Проверяет, жив ли сервер
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /alena-rofl/ [get]
func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "я родился",
	})
}

// @Summary Тест авторизации
// @Description Доступ будет только при успешной авторизации
// @Tags auth
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /alena-rofl/user_check [get]
func UserCheck(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized or can't find user_id in context."})
		return
	}
	c.JSON(200, gin.H{
		"message": "Привет, " + userID.(string),
	})
}

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(cors.Default()) //потом поставить middleware
	api := router.Group("/alena-rofl")
	{
		api.GET("/", HealthCheck)
		api.GET("/user_check", service.AuthMiddleware(), UserCheck)
		api.POST("/auth", service.Auth)
		api.POST("/refresh", service.Refresh)
	}

	return router
}
