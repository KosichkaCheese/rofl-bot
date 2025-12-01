package config

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"log"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(postgres.Open("postgresql://rofl_db_user:mGuaV5D9GACpgDe3ZXtn92i8IJJqnUQm@dpg-d4m4p6juibrs738c2v60-a/rofl_db"), &gorm.Config{})
	// DB, err = gorm.Open(postgres.Open("host=localhost port=5438 dbname=test user=test password=test sslmode=disable TimeZone=UTC"), &gorm.Config{}) //локальная бд
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	log.Println("Database connected")
}
