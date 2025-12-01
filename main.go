package main

import (
	"log"
	"rofl-bot/config"
	_ "rofl-bot/docs"
	"rofl-bot/router"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	log.Println("Starting alena rofls ...")

	config.InitDB()
	_ = config.DB

	r := router.SetupRouter()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":8000")
}
