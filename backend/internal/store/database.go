package store

import (
	"log"

	"github.com/jj/novelist/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := DB.AutoMigrate(
		&model.User{},
		&model.Project{},
		&model.Character{},
		&model.WorldSetting{},
		&model.Outline{},
		&model.Chapter{},
		&model.Discussion{},
		&model.Conversation{},
		&model.Setting{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database connected and migrated")
}

func GetDB() *gorm.DB {
	return DB
}
