package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PakarDashboardSummary struct {
	Stats struct {
		TotalSiswa     int64 `json:"total_siswa"`
		TesSelesai     int64 `json:"tes_selesai"`
		ButuhPerhatian int64 `json:"butuh_perhatian"`
	} `json:"stats"`

	Grafik []struct {
		Kategori string `json:"kategori"`
		Jumlah   int64  `json:"jumlah"`
	} `json:"grafik"`
}

func GetPakarDashboardData(c *gin.Context) {
	var data PakarDashboardSummary

	// 1. Hitung Total Siswa
	database.DB.Model(&models.Students{}).Count(&data.Stats.TotalSiswa)

	// 2. Hitung Total Tes Selesai (Hasil Diagnosis)
	database.DB.Model(&models.HasilDiagnosis{}).Count(&data.Stats.TesSelesai)

	// 3. Hitung yang butuh perhatian (Confidence Score > 75%)
	// Menggunakan rata-rata confidence depresi dan cemas sebagai indikator atau salah satunya
	database.DB.Model(&models.HasilDiagnosis{}).
		Where("nn_depresi_confidence > ? OR nn_cemas_confidence > ?", 75, 75).
		Count(&data.Stats.ButuhPerhatian)

	// 4. Data Grafik - Sebaran Penyakit (Gabungan Depresi dan Cemas)
	// Kita ambil distribusi dari final_depresi_penyakit
	var grafikDepresi []struct {
		Kategori string `json:"kategori"`
		Jumlah   int64  `json:"jumlah"`
	}
	database.DB.Model(&models.HasilDiagnosis{}).
		Select("final_depresi_penyakit as kategori, count(*) as jumlah").
		Group("final_depresi_penyakit").
		Scan(&grafikDepresi)

	// Masukkan ke grafik utama
	for _, item := range grafikDepresi {
		if item.Kategori != "" {
			data.Grafik = append(data.Grafik, item)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Berhasil mendapatkan data dashboard pakar",
		"Data":    data,
	})
}
