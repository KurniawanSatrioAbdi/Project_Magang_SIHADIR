package config

import (
	"log"

	"presensi-kominfo/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {

	dsn := "root:@tcp(127.0.0.1:3306)/presensi_kominfo?charset=utf8mb4&parseTime=True&loc=Local"

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Database gagal terkoneksi")
	}

	database.AutoMigrate(&models.Presensi{})

	DB = database
}
