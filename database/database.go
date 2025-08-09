package database

import (
	"Skripsi-Backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}

	log.Println("Koneksi database berhasil dibuka")
}
func Migrate() {
	log.Println("Menjalankan migrasi database...")

	err := DB.AutoMigrate(&models.Students{})
	if err != nil {
		log.Fatalf("Gagal migrasi database: %v", err)
	}
	log.Println("Migrasi database berhasil")
}
