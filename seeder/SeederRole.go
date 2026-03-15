package seeder

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"log"

	"github.com/google/uuid"
)

func SeederRole() {
	var count int64

	database.DB.Model(&models.Role{}).Count(&count)

	if count > 0 {
		log.Println("Role already exists")
		return
	}

	RoleSeeder := []models.Role{
		{
			RoleUID:     uuid.NewString(),
			RoleName:    "Administrator",
			Description: "Role Administrator adalah sebuah Role yang memegang semua kendali Aplikasi Sistem Pakar",
		},
		{
			RoleUID:     uuid.NewString(),
			RoleName:    "Pakar",
			Description: "Role Pakar merupakan Role yang bertugas mengisi basis pengetahuan yang ada di aplikasi  ",
		},
		{
			RoleUID:     uuid.NewString(),
			RoleName:    "Teacher",
			Description: "Role Teacher adalah Role yang bertugas untuk Pengawas Hasil Diagnosis dari Peserta Didik",
		},
		{
			RoleUID:     uuid.NewString(),
			RoleName:    "Student",
			Description: "Role Student adalah sebuah Role yang menjadi objek penelitian diagnosis penyakit mental pada aplikasi ini",
		},
	}

	for _, role := range RoleSeeder {
		if err := database.DB.Create(&role).Error; err != nil {
			log.Fatalf("❌ Gagal membuat user %s: %v", role.RoleName, err)
		}
		log.Printf("✅ User dibuat: %s (%s)", role.RoleName, role.RoleUID)
	}

	log.Println("🌱 Seeder users berhasil dijalankan!")
}
