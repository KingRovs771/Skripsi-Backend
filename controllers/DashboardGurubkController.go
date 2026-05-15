package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"Skripsi-Backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GurubkDashboardSummary struct {
	Sekolah struct {
		Nama string `json:"nama_sekolah"`
		NPSN string `json:"npsn"`
	} `json:"sekolah"`
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

func GetGurubkDashboardData(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
		return
	}

	var teacher models.Teachers
	if err := database.DB.Where("n_ip = ?", claims.ID).First(&teacher).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teacher profile"})
		return
	}

	var sekolah models.Sekolah
	database.DB.Where("CAST(npsn AS TEXT) = ?", teacher.NPSN).First(&sekolah)
	
	var data GurubkDashboardSummary
	
	// Default values
	if sekolah.NamaSekolah != "" {
		data.Sekolah.Nama = sekolah.NamaSekolah
	} else {
		data.Sekolah.Nama = "Sekolah Tidak Diketahui"
	}
	data.Sekolah.NPSN = teacher.NPSN
	
	// We should join with students to get accurate counts for this NPSN
	// 1. Hitung Total Siswa
	database.DB.Model(&models.Students{}).Where("npsn = ?", teacher.NPSN).Count(&data.Stats.TotalSiswa)

	// 2. Hitung Total Tes Selesai (Hasil Diagnosis) by students of this NPSN
	database.DB.Table("hasil_diagnoses").
		Joins("JOIN test_sessions ON hasil_diagnoses.session_test_uid = test_sessions.test_session_id").
		Joins("JOIN students ON test_sessions.user_uid = students.students_uid").
		Where("students.npsn = ?", teacher.NPSN).
		Count(&data.Stats.TesSelesai)

	// 3. Hitung yang butuh perhatian (Confidence Score > 75%)
	database.DB.Table("hasil_diagnoses").
		Joins("JOIN test_sessions ON hasil_diagnoses.session_test_uid = test_sessions.test_session_id").
		Joins("JOIN students ON test_sessions.user_uid = students.students_uid").
		Where("students.npsn = ? AND (hasil_diagnoses.nn_depresi_confidence > ? OR hasil_diagnoses.nn_cemas_confidence > ?)", teacher.NPSN, 75, 75).
		Count(&data.Stats.ButuhPerhatian)

	// 4. Data Grafik - Sebaran Penyakit
	var grafikDepresi []struct {
		Kategori string `json:"kategori"`
		Jumlah   int64  `json:"jumlah"`
	}
	database.DB.Table("hasil_diagnoses").
		Select("hasil_diagnoses.final_depresi_penyakit as kategori, count(*) as jumlah").
		Joins("JOIN test_sessions ON hasil_diagnoses.session_test_uid = test_sessions.test_session_id").
		Joins("JOIN students ON test_sessions.user_uid = students.students_uid").
		Where("students.npsn = ?", teacher.NPSN).
		Group("hasil_diagnoses.final_depresi_penyakit").
		Scan(&grafikDepresi)

	for _, item := range grafikDepresi {
		if item.Kategori != "" {
			data.Grafik = append(data.Grafik, item)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Berhasil mendapatkan data dashboard Guru BK",
		"Data":    data,
	})
}
