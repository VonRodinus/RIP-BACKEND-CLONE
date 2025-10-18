package db

import (
	"RIP/internal/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	dsn := "host=localhost user=vonrodinus password=VonRodinus005 dbname=chronus_db port=5432 sslmode=disable"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	DB.AutoMigrate(&models.Artifact{}, &models.TPQRequest{}, &models.TPQRequestItem{}, &models.User{})
}
