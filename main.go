package main

import (
	"log"
	"rofl-bot/config"
	"rofl-bot/controller"
	_ "rofl-bot/docs"
	"rofl-bot/router"
	"rofl-bot/service"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	log.Println("Starting alena rofls ...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	config.InitDB(cfg)
	_ = config.DB

	authService := service.NewAuthService(cfg)
	authCtrl := controller.NewAuthCtrl(authService)

	r := router.SetupRouter(authCtrl)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":8000")
}
