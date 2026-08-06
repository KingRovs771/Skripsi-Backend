package seeder

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func hashPasswordAdministratorSeeder(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func SeederUsersAdministrator() {
	var count int64

	database.DB.Model(&models.Administrator{}).Count(&count)

	if count > 0 {
		log.Println("Users Administrator Already Exists")
		return
	}
	rawPasswordAdmin := "yolnDA2623*KingRovs771"
	hashedPassword, err := hashPasswordAdministratorSeeder(rawPasswordAdmin)
	if err != nil {
		log.Fatal("Gagal Hashing")
	}

	SeederAdministrator := []models.Administrator{
		{
			AdminUID:    uuid.NewString(),
			RoleUID:     "b9160dfa-48a1-47a3-b3bd-8c85add9a95c",
			NamaLengkap: "RzBudi",
			Phone:       "087787577749",
			Email:       "rizkybudiarto890@gmail.com",
			Alamat:      "Teguhan RT 01 RW 01, Sragen Wetan, Sragen",
			Password:    hashedPassword,
		},
	}

	for _, admin := range SeederAdministrator {
		if err := database.DB.Create(&admin).Error; err != nil {
			log.Fatalf("❌ Gagal membuat Users %s: %v", admin.NamaLengkap, err)
		}
		log.Printf("✅ Users Admin dibuat: %s (%s)\n", admin.NamaLengkap, admin.AdminUID)
	}

	log.Println("🌱 Seeder users berhasil dijalankan!")
}
