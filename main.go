package main

import (
	"log"
	"rofl-bot/config"
	"rofl-bot/controller"
	_ "rofl-bot/docs"
	"rofl-bot/repository"
	"rofl-bot/router"
	"rofl-bot/service"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Rofl Bot API
// @version 1.0
// @description API для tg Mini App "Приколы у Алёны"

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите "Bearer {access_token}" (без кавычек)

func main() {
	log.Println("Starting alena rofls ...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	config.InitDB(cfg)
	_ = config.DB

	authRep := repository.NewAuthRep(config.DB)
	authService := service.NewAuthService(cfg, authRep)
	authCtrl := controller.NewAuthCtrl(authService)

	r := router.SetupRouter(cfg, authCtrl)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":8000")
}
