package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var DB *gorm.DB

func ConnectDB() *gorm.DB {
	db, err := gorm.Open("postgres", "host=127.0.0.1 port=5432 user=user dbname=user password=user sslmode=disable")
	if err != nil {
		panic("error to connect database")
	}
	DB = db
	DB.AutoMigrate(&Track{})

	return db
}
