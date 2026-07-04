package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"Skripsi-Backend/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type StudentHistorySummary struct {
	NISN        string    `json:"nisn"`
	Nama        string    `json:"nama"`
	TotalTes    int64     `json:"total_tes"`
	TerakhirTes time.Time `json:"terakhir_tes"`
}

type StudentInfo struct {
	NISN  string `json:"nisn"`
	Nama  string `json:"nama"`
	Kelas string `json:"kelas"`
	Email string `json:"email"`
}

type TestResult struct {
	ID                  int       `gorm:"column:id" json:"id"`
	SessionUID          string    `gorm:"column:session_uid" json:"session_uid"`
	NamaTes             string    `gorm:"column:nama_tes" json:"nama_tes"`
	Skor                int64     `gorm:"column:skor" json:"skor"`
	SkorPHQ9            int64     `gorm:"column:skor_phq9" json:"skor_phq9"`
	SkorGAD7            int64     `gorm:"column:skor_gad7" json:"skor_gad7"`
	Kategori            string    `gorm:"column:kategori" json:"kategori"`
	DepresiPenyakit     string    `gorm:"column:depresi_penyakit" json:"depresi_penyakit"`
	CemasPenyakit       string    `gorm:"column:cemas_penyakit" json:"cemas_penyakit"`
	Tanggal             time.Time `gorm:"column:tanggal" json:"tanggal"`
	Rekomendasi         string    `gorm:"column:rekomendasi" json:"rekomendasi"`
	CeritaSiswa         string    `gorm:"column:cerita_siswa" json:"cerita_siswa"`
	IsVisibleToStudent  bool      `gorm:"column:is_visible_to_student" json:"is_visible_to_student"`
	ReviewedByGurubk    bool      `gorm:"column:reviewed_by_gurubk" json:"reviewed_by_gurubk"`
	NNDepresiConfidence float64   `gorm:"column:nn_depresi_confidence" json:"nn_depresi_confidence"`
	NNCemasConfidence   float64   `gorm:"column:nn_cemas_confidence" json:"nn_cemas_confidence"`
}

type ReviewRequest struct {
	IsVisibleToStudent bool   `json:"is_visible_to_student"`
	ReviewedByGurubk   bool   `json:"reviewed_by_gurubk"`
	Rekomendasi        string `json:"rekomendasi"`
}

func GetGurubkHistory(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
		return
	}

	var teacherNPSN string
	if err := database.DB.Table("teachers").Where("n_ip = ?", claims.ID).Select("npsn").Scan(&teacherNPSN).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teacher profile"})
		return
	}

	var results []StudentHistorySummary
	query := `
		SELECT s.nisn, s.nama_lengkap as nama, 
		       COUNT(ts.test_session_id) as total_tes, 
		       MAX(ts.created_at) as terakhir_tes
		FROM students s
		JOIN test_sessions ts ON s.students_uid = ts.user_uid
		WHERE s.npsn = ? AND ts.status = 'SELESAI'
		GROUP BY s.nisn, s.nama_lengkap
	`
	if err := database.DB.Raw(query, teacherNPSN).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Data": results,
	})
}

func GetGurubkHistoryDetail(c *gin.Context) {
	nisn := c.Param("nisn")
	var studentInfo StudentInfo
	if err := database.DB.Table("students").Where("nisn = ?", nisn).Select("nisn, nama_lengkap as nama, kelas, email").Scan(&studentInfo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	var results []TestResult
	query := `
		SELECT hd.result_id as id, ts.test_session_id as session_uid, 
		       'Diagnosis Kesehatan Mental' as nama_tes,
		       (ts.total_scorephq9 + ts.total_scoregad7) as skor,
		       ts.total_scorephq9 as skor_phq9,
		       ts.total_scoregad7 as skor_gad7,
		       CONCAT(p_depresi.nama_penyakit, ' & ', p_cemas.nama_penyakit) as kategori,
		       p_depresi.nama_penyakit as depresi_penyakit,
		       p_cemas.nama_penyakit as cemas_penyakit,
		       ts.created_at as tanggal,
		       hd.rekomendasi,
		       sf.cerita_siswa,
		       hd.is_visible_to_student, hd.reviewed_by_gurubk,
		       hd.nn_depresi_confidence, hd.nn_cemas_confidence
		FROM students s
		JOIN test_sessions ts ON s.students_uid = ts.user_uid
		JOIN hasil_diagnoses hd ON ts.test_session_id = hd.session_test_uid
		LEFT JOIN student_feedbacks sf ON ts.test_session_id = sf.test_session_uid
		LEFT JOIN penyakits p_depresi ON hd.final_depresi_penyakit = p_depresi.kode_penyakit
		LEFT JOIN penyakits p_cemas ON hd.final_cemas_penyakit = p_cemas.kode_penyakit
		WHERE s.nisn = ? AND ts.status = 'SELESAI'
		ORDER BY ts.created_at DESC
	`
	if err := database.DB.Raw(query, nisn).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Data": gin.H{
			"student": studentInfo,
			"results": results,
		},
	})
}

func UpdateHistoryReview(c *gin.Context) {
	resultId := c.Param("id")
	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid"})
		return
	}

	if err := database.DB.Model(&models.HasilDiagnosis{}).Where("result_id = ?", resultId).Updates(map[string]interface{}{
		"is_visible_to_student": req.IsVisibleToStudent,
		"reviewed_by_gurubk":    req.ReviewedByGurubk,
		"rekomendasi":           req.Rekomendasi,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status berhasil diupdate"})
}

func GetStudentHistory(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
		return
	}

	studentUID := claims.ID

	var results []TestResult
	query := `
		SELECT hd.result_id as id, ts.test_session_id as session_uid, 
		       'Diagnosis Kesehatan Mental' as nama_tes,
		       (ts.total_scorephq9 + ts.total_scoregad7) as skor,
		       ts.total_scorephq9 as skor_phq9,
		       ts.total_scoregad7 as skor_gad7,
		       CONCAT(p_depresi.nama_penyakit, ' & ', p_cemas.nama_penyakit) as kategori,
		       p_depresi.nama_penyakit as depresi_penyakit,
		       p_cemas.nama_penyakit as cemas_penyakit,
		       ts.created_at as tanggal,
		       hd.rekomendasi,
		       sf.cerita_siswa,
		       hd.is_visible_to_student, hd.reviewed_by_gurubk,
		       hd.nn_depresi_confidence, hd.nn_cemas_confidence
		FROM test_sessions ts
		JOIN hasil_diagnoses hd ON ts.test_session_id = hd.session_test_uid
		LEFT JOIN student_feedbacks sf ON ts.test_session_id = sf.test_session_uid
		LEFT JOIN penyakits p_depresi ON hd.final_depresi_penyakit = p_depresi.kode_penyakit
		LEFT JOIN penyakits p_cemas ON hd.final_cemas_penyakit = p_cemas.kode_penyakit
		WHERE ts.user_uid = ? AND ts.status = 'SELESAI'
		ORDER BY ts.created_at DESC
	`
	if err := database.DB.Raw(query, studentUID).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Data": results,
	})
}
