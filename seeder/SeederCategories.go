package seeder

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"log"
)

func SeederCategories() {
	var count int64

	database.DB.Model(&models.Category{}).Count(&count)
	if count > 0 {
		log.Println("Categories Seeder Already Exists")
		return
	}

	CategoriesSeeder := []models.Category{
		{
			NameCategory: "Anxiety",
			Description:  "Anxiety (kecemasan) adalah respons alami tubuh terhadap stres, berupa perasaan takut, khawatir, atau tegang yang wajar saat menghadapi situasi menantang. ",
		},
		{
			NameCategory: "Stress",
			Description:  "Stres adalah reaksi fisik dan emosional alami manusia terhadap tekanan, ancaman, atau perubahan lingkungan yang menuntut penyesuaian diri.",
		},
		{
			NameCategory: "ADHD",
			Description:  "ADHD (Attention Deficit Hyperactivity Disorder) atau Gangguan Pemusatan Perhatian dan Hiperaktivitas (GPPH) adalah gangguan perkembangan saraf/mental yang umum, ditandai pola menetap berupa kesulitan fokus, hiperaktif, dan perilaku impulsif.",
		},
		{
			NameCategory: "Depresi",
			Description:  "Depresi adalah gangguan suasana hati (mood) serius yang ditandai dengan perasaan sedih mendalam, putus asa, dan kehilangan minat terhadap aktivitas yang disukai secara terus-menerus selama minimal 2 minggu. ",
		},
	}

	for _, categories := range CategoriesSeeder {
		if err := database.DB.Create(&categories).Error; err != nil {
			log.Fatalf("❌ Gagal membuat Category %s: %v", categories.NameCategory, err)
		}
		log.Printf("✅ Category dibuat: %s (%s)", categories.NameCategory, categories.CategoryUID)
	}

	log.Println("🌱 Seeder Category berhasil dijalankan!")
}
