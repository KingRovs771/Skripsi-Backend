package controllers

import (
	"Skripsi-Backend/database"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type TrendRow struct {
	TanggalTes           time.Time `json:"tanggal_tes" gorm:"column:created_at"`
	TotalScorePHQ9       int64     `json:"total_score_phq9" gorm:"column:total_scorephq9"`
	TotalScoreGAD7       int64     `json:"total_score_gad7" gorm:"column:total_scoregad7"`
	KategoriDepresiFinal string    `json:"kategori_depresi_final" gorm:"column:kategori_depresi"`
	KategoriCemasFinal   string    `json:"kategori_cemas_final" gorm:"column:kategori_cemas"`
}

// GetStudentDiagnosisTrend mengembalikan data longitudinal diagnosis siswa
// untuk keperluan visualisasi grafik tren PHQ-9 & GAD-7.
func GetStudentDiagnosisTrend(c *gin.Context) {
	studentUID := c.Param("student_uid")

	userUID, existsUID := c.Get("user_uid")
	userType, existsType := c.Get("user_type")

	if !existsUID || !existsType {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":  "Error",
			"Message": "Tidak terotentikasi",
		})
		return
	}

	// Proteksi UU PDP: Siswa hanya boleh mengakses data miliknya sendiri.
	// Admin, Guru BK, dan Pakar diperbolehkan melihat data siswa mana pun.
	if userType == "student" && userUID.(string) != studentUID {
		c.JSON(http.StatusForbidden, gin.H{
			"Status":  "Error",
			"Message": "Akses ditolak: Anda tidak dapat melihat data tren siswa lain",
		})
		return
	}

	query := `
		SELECT
			ts.created_at,
			ts.total_scorephq9,
			ts.total_scoregad7,
			COALESCE(p_dep.nama_penyakit, hd.final_depresi_penyakit, '') AS kategori_depresi,
			COALESCE(p_cem.nama_penyakit, hd.final_cemas_penyakit,   '') AS kategori_cemas
		FROM test_sessions ts
		JOIN hasil_diagnoses hd   ON ts.test_session_id = hd.session_test_uid
		LEFT JOIN penyakits p_dep ON hd.final_depresi_penyakit = p_dep.kode_penyakit
		LEFT JOIN penyakits p_cem ON hd.final_cemas_penyakit   = p_cem.kode_penyakit
		WHERE ts.user_uid = ? AND ts.status = 'SELESAI'
		ORDER BY ts.created_at ASC
	`

	var trendData []TrendRow
	if err := database.DB.Raw(query, studentUID).Scan(&trendData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal mengambil data tren diagnosis",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status": "Success",
		"Data":   trendData,
	})
}
