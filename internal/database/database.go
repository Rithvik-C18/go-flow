package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDatabase(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		return nil, err
	}

	log.Println("Database connection established")

	err = db.AutoMigrate(
		&User{},
		&Workflow{},
		&Node{},
		&Edge{},
	)

	if err != nil {
		return nil, err
	}
	log.Println("Database migrated successfully")

	return db, nil
}
