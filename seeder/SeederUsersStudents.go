package seeder

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"log"

	"github.com/google/uuid"
)

func SeederUsersStudents() {
	var count int64

	database.DB.Model(&models.Students{}).Count(&count)

	if count > 0 {
		log.Println("User already exists")
		return
	}
	StudentSeeder := []models.Students{
		{
			StudentsUID: uuid.NewString(),
			RoleUID:     "1659c8be-961d-4912-9b80-b04741cc82ca",
			NISN:        "008995733",
			NamaLengkap: "Students 1",
			NPSN:        "20354028",
			Kelas:       "XII A",
			NoHp:        "083665893938",
			Alamat:      "Jl. Perintis Kemerdekaan No. 16",
			Email:       "student1@sman1sragen.sch.id",
			Password:    "akusayangkamu123",
		},
		{
			StudentsUID: uuid.NewString(),
			RoleUID:     "1659c8be-961d-4912-9b80-b04741cc82ca",
			NISN:        "0048567283",
			NamaLengkap: "Students 2",
			NPSN:        "20312960",
			Kelas:       "XII B",
			NoHp:        "089345778847",
			Alamat:      "Jl. Perintis Kemerdekaan No. 16",
			Email:       "student2@smpn1sragen.sch.id",
			Password:    "akusayangkamu123",
		},

		{
			StudentsUID: uuid.NewString(),
			RoleUID:     "1659c8be-961d-4912-9b80-b04741cc82ca",
			NISN:        "0087483945",
			NamaLengkap: "Students 3",
			NPSN:        "20312904",
			Kelas:       "XII TKJ 2",
			NoHp:        "083665893938",
			Alamat:      "Jl. Perintis Kemerdekaan No. 16",
			Email:       "student3@smkn2sragen.sch.id",
			Password:    "akusayangkamu123",
		},
	}
	for _, student := range StudentSeeder {
		if err := database.DB.Create(&student).Error; err != nil {
			log.Fatal("❌ Gagal membuat user %s: %v", student.NamaLengkap)
		}
		log.Println("✅ User dibuat: %s (%s)", student.NamaLengkap, student.StudentsUID)
	}
	log.Println("🌱 Seeder users berhasil dijalankan!")
}
