package config

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"log"
	"os"
	"rofl-bot/domain"
	"time"
)

var DB *gorm.DB

func InitDB(cfg *Config) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true, // Игнорировать record not found
			Colorful:                  true,
		},
	)

	var err error
	DB, err = gorm.Open(postgres.Open("postgresql://rofl_db_user:mGuaV5D9GACpgDe3ZXtn92i8IJJqnUQm@dpg-d4m4p6juibrs738c2v60-a/rofl_db"), &gorm.Config{Logger: newLogger})
	// DB, err = gorm.Open(postgres.Open(cfg.DB_URL), &gorm.Config{Logger: newLogger}) //локальная бд
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	log.Println("Database connected")
	log.Println("Migrating database...")

	// DB.Migrator().DropTable(&domain.Role{}, &domain.Token{}, &domain.User{}, &domain.Event{}, &domain.EventMember{}, &domain.Product{}, &domain.ProductMember{}, &domain.Invite{})
	DB.AutoMigrate(&domain.User{}, &domain.Token{}, &domain.Event{}, &domain.Role{}, &domain.EventMember{}, &domain.Product{}, &domain.ProductMember{}, &domain.Invite{})

	DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&domain.Role{Id: 1, Name: "admin"})
	DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&domain.Role{Id: 2, Name: "user"})

	log.Println("Database migrated")
}
