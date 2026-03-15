package seeder

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"log"

	"github.com/google/uuid"
)

func SeederSekolah() {
	var count int64

	database.DB.Model(&models.Sekolah{}).Count(&count)

	if count > 0 {
		log.Println("Sekolah Already exists")
		return
	}

	SekolahSeeder := []models.Sekolah{
		{
			SekolahUID:    uuid.NewString(),
			NPSN:          20354028,
			NamaSekolah:   "SMA Negeri 1 Sragen",
			Jenjang:       "SMA",
			AlamatSekolah: " Jl. Perintis Kemerdekaan No.16, Dusun Kebayanan Sragen Manggis, Sragen Wetan, Kec. Sragen, Kabupaten Sragen, Jawa Tengah 57214",
		},
		{
			SekolahUID:    uuid.NewString(),
			NPSN:          20313028,
			NamaSekolah:   "SMA Negeri 2 Sragen",
			Jenjang:       "SMA",
			AlamatSekolah: "Jl. Anggrek No.34, Kebayan 1, Sragen Kulon, Kec. Sragen, Kabupaten Sragen, Jawa Tengah 57212",
		},
		{
			SekolahUID:    uuid.NewString(),
			NPSN:          20313027,
			NamaSekolah:   "SMA Negeri 3 Sragen",
			Jenjang:       "SMA",
			AlamatSekolah: "Jl. Dr. Sutomo No.2, Kebayan 1, Sragen Kulon, Kec. Sragen, Kabupaten Sragen, Jawa Tengah 5721",
		},
		{
			SekolahUID:    uuid.NewString(),
			NPSN:          20313046,
			NamaSekolah:   "SMK Negeri 1 Sragen",
			Jenjang:       "SMA",
			AlamatSekolah: "Jl. Ronggowarsito, Dusun Kebayanan Sragen Manggis, Sragen Wetan, Kec. Sragen, Kota Surakarta, Jawa Tengah 57214",
		},
		{
			SekolahUID:    uuid.NewString(),
			NPSN:          20312904,
			NamaSekolah:   "SMK Negeri 2 Sragen",
			Jenjang:       "SMA",
			AlamatSekolah: "Jl. Dr. Sutomo No.4, Kebayan 1, Sragen Kulon, Kec. Sragen, Kabupaten Sragen, Jawa Tengah 57212",
		},
		{
			SekolahUID:    uuid.NewString(),
			NPSN:          20312960,
			NamaSekolah:   "SMP Negeri 1 Sragen",
			Jenjang:       "SMP",
			AlamatSekolah: "Jl. Sukowati No.162, Kebayan 3, Sragen Kulon, Kec. Sragen, Kabupaten Sragen, Jawa Tengah 57212",
		},
		{
			SekolahUID:    uuid.NewString(),
			NPSN:          20312941,
			NamaSekolah:   "SMP Negeri 2 Sragen",
			Jenjang:       "SMP",
			AlamatSekolah: "Jl. Sukowati No.257, Karang Duwo, Sragen Tengah, Kec. Sragen, Kabupaten Sragen, Jawa Tengah 57211",
		},
		{
			SekolahUID:    uuid.NewString(),
			NPSN:          20312934,
			NamaSekolah:   "SMP Negeri 3 Sragen",
			Jenjang:       "SMP",
			AlamatSekolah: "Jalan Gatot Subroto.57, RW No.15, Kebayan 3, Sragen Kulon, Kec. Sragen, Kabupaten Sragen, Jawa Tengah 57212",
		},
		{
			SekolahUID:    uuid.NewString(),
			NPSN:          20312933,
			NamaSekolah:   "SMP Negeri 4 Sragen",
			Jenjang:       "SMP",
			AlamatSekolah: "Jl. Patimura No.4/5, Mageru, Sragen Tengah, Kec. Sragen, Kabupaten Sragen, Jawa Tengah 57211",
		},
		{
			SekolahUID:    uuid.NewString(),
			NPSN:          20312932,
			NamaSekolah:   "SMP Negeri 5 Sragen",
			Jenjang:       "SMP",
			AlamatSekolah: "Jl. Mawar No.4, Kebayan 1, Sragen Kulon, Kec. Sragen, Kabupaten Sragen, Jawa Tengah 57212",
		},
	}

	for _, sekolah := range SekolahSeeder {
		if err := database.DB.Create(&sekolah).Error; err != nil {
			log.Fatal("❌ Gagal membuat Sekolah %s: %v", sekolah.NamaSekolah)
		}
		log.Println("✅ Sekolah dibuat: %s (%s)", sekolah.NamaSekolah, sekolah.NPSN)
	}
}
